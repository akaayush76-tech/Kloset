package engine

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Modifier deltas per spec §4.2.
const (
	bonusCompleteness  = 5  // outfit covers 3+ slots
	bonusAvatarFit     = 8  // every try-on-able item has a "fitted" try-on result
	bonusWearHistory   = 4  // user has worn this top+bottom combination before
	bonusContextFilter = 3  // outfit matches the active context filter
	penaltyMissingItem = -5 // outfit has a shop-to-complete slot
	penaltyLowConf     = -3 // any item auto-tagged with confidence < 0.7
)

const lowConfidenceThreshold = 0.7

// modifierContext carries the per-request user signals needed by applyModifiers.
// Both lookups are fetched once per request, not per combination.
type modifierContext struct {
	contextFilter    string
	fittedProductIDs map[string]bool // catalog product IDs with a "fitted" try-on result
	wornComboKeys    map[string]bool // top+bottom combo keys the user has worn
}

// fetchFittedProductIDs loads the product IDs the user has tried on their
// avatar with a "fitted" (best fit) result.
func fetchFittedProductIDs(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) map[string]bool {
	ids := map[string]bool{}
	cursor, err := db.Collection("tryons").Find(ctx, bson.M{"userId": userID, "fit": "fitted"})
	if err != nil {
		return ids
	}
	defer cursor.Close(ctx)
	var records []struct {
		ProductID string `bson:"productId"`
	}
	if err := cursor.All(ctx, &records); err != nil {
		return ids
	}
	for _, r := range records {
		ids[r.ProductID] = true
	}
	return ids
}

// fetchWornComboKeys loads the top+bottom combo keys the user has logged as worn.
func fetchWornComboKeys(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) map[string]bool {
	keys := map[string]bool{}
	cursor, err := db.Collection("wearHistory").Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return keys
	}
	defer cursor.Close(ctx)
	var records []struct {
		ComboKey string `bson:"comboKey"`
	}
	if err := cursor.All(ctx, &records); err != nil {
		return keys
	}
	for _, r := range records {
		keys[r.ComboKey] = true
	}
	return keys
}

// hasMissingRequiredSlot reports whether the combination needs a
// shop-to-complete item to form a full outfit (spec's missing item penalty).
func hasMissingRequiredSlot(c combo) bool {
	slots := map[string]bool{}
	for _, item := range c.items {
		slots[outfitSlot(item.Category)] = true
	}
	if slots["full_body"] {
		return false
	}
	return !slots["top"] || !slots["bottom"]
}

// applyModifiers applies the spec §4.2 modifiers to a base score and returns
// the adjusted score clamped to 0–100.
func applyModifiers(base int, c combo, mc modifierContext) int {
	score := base

	// Completeness bonus: items in 3+ slots.
	slots := map[string]bool{}
	for _, item := range c.items {
		slots[outfitSlot(item.Category)] = true
	}
	if len(slots) >= 3 {
		score += bonusCompleteness
	}

	// Avatar fit boost: every item that can have a try-on record (catalog /
	// wishlist items) got a "fitted" result — and there is at least one such item.
	tryOnAble := 0
	allFitted := true
	for _, item := range c.items {
		if item.CatalogProductID == "" {
			continue // owned closet items have no try-on records
		}
		tryOnAble++
		if !mc.fittedProductIDs[item.CatalogProductID] {
			allFitted = false
		}
	}
	if tryOnAble > 0 && allFitted {
		score += bonusAvatarFit
	}

	// Wear history bonus: user has worn this top+bottom pair before.
	if mc.wornComboKeys[comboKey(c)] {
		score += bonusWearHistory
	}

	// Context filter match: every item's occasion agrees with the active filter.
	if mc.contextFilter != "" && mc.contextFilter != "all" {
		matches := true
		for _, item := range c.items {
			o := item.Identifiers.Occasion
			if o != "" && o != "all" && o != mc.contextFilter {
				matches = false
				break
			}
		}
		if matches {
			score += bonusContextFilter
		}
	}

	// Missing item penalty: a required slot must be shopped to complete the look.
	if hasMissingRequiredSlot(c) {
		score += penaltyMissingItem
	}

	// Low-confidence tag penalty: any auto-tagged item below the confidence bar.
	for _, item := range c.items {
		if conf := item.Identifiers.TagConfidence; conf > 0 && conf < lowConfidenceThreshold {
			score += penaltyLowConf
			break
		}
	}

	return clampScore(score)
}
