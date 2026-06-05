package engine

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/kloset/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const maxOptionalCandidates = 5

// RecommendRequest mirrors the API request body.
type RecommendRequest struct {
	TriggerItemID    string `json:"trigger_item_id"`
	TriggerItemType  string `json:"trigger_item_type"` // "catalog" | "closet"
	ContextFilter    string `json:"context_filter"`    // "all" | "casual" | "smart_casual" | "date_night" | "weekend"
	Limit            int    `json:"limit"`
	IncludeShopItems bool   `json:"include_shop_items"`
	UserID           string `json:"-"` // injected by handler
}

// OutfitItem is a single item within a returned outfit.
type OutfitItem struct {
	ItemID      string                 `json:"item_id"`
	Name        string                 `json:"name"`
	Brand       string                 `json:"brand"`
	Category    string                 `json:"category"`
	ImageURL    string                 `json:"image_url"`
	IsTrigger   bool                   `json:"is_trigger"`
	Owned       bool                   `json:"owned"`
	Price       float64                `json:"price"`
	Identifiers models.ItemIdentifiers `json:"identifiers"`
}

// Outfit is a ranked outfit combination.
type Outfit struct {
	OutfitID       string         `json:"outfit_id"`
	Rank           int            `json:"rank"`
	Score          int            `json:"score"`
	RankLabel      string         `json:"rank_label"`
	Items          []OutfitItem   `json:"items"`
	ScoreBreakdown ScoreBreakdown `json:"score_breakdown"`
	WhyText        string         `json:"why_text"`
	RuleTags       []string       `json:"rule_tags"`
	MissingItems   []OutfitItem   `json:"missing_items"`
}

// RecommendMeta holds diagnostic metadata for the response.
type RecommendMeta struct {
	TriggerItemID            string `json:"trigger_item_id"`
	ClosetItemsConsidered    int    `json:"closet_items_considered"`
	CombinationsEvaluated    int    `json:"combinations_evaluated"`
	CombinationsAfterFilters int    `json:"combinations_after_filters"`
	Returned                 int    `json:"returned"`
	LatencyMS                int64  `json:"latency_ms"`
}

// RecommendResult is the full response payload.
type RecommendResult struct {
	Outfits []Outfit      `json:"outfits"`
	Meta    RecommendMeta `json:"meta"`
}

// combo holds a set of wardrobe items forming one candidate outfit.
type combo struct {
	items []models.WardrobeItem
	score int
	bd    ScoreBreakdown
}

// comboKey returns a deduplication key based on the top+bottom pair.
func comboKey(c combo) string {
	top, bottom := "", ""
	for _, item := range c.items {
		switch item.Category {
		case "upper":
			top = item.ID.Hex()
		case "lower":
			bottom = item.ID.Hex()
		case "full_body":
			top = item.ID.Hex()
			bottom = "full_body"
		}
	}
	return top + "|" + bottom
}

// fetchTriggerAsWardrobeItem loads either a catalog product or closet item and normalises
// it to WardrobeItem so the engine can treat both uniformly.
func fetchTriggerAsWardrobeItem(ctx context.Context, db *mongo.Database, req RecommendRequest) (models.WardrobeItem, error) {
	if req.TriggerItemType == "catalog" {
		col := db.Collection("products")
		var p models.Product
		if err := col.FindOne(ctx, bson.M{"_id": req.TriggerItemID}).Decode(&p); err != nil {
			return models.WardrobeItem{}, fmt.Errorf("catalog item not found: %w", err)
		}
		image := ""
		if len(p.Images) > 0 {
			image = p.Images[0]
		}
		return models.WardrobeItem{
			ID:          primitive.NewObjectID(), // synthetic — used only for slot exclusion
			Name:        p.Name,
			Category:    p.Category,
			Brand:       p.Brand,
			Price:       p.Price,
			Image:       image,
			Identifiers: p.Identifiers,
		}, nil
	}

	// closet item
	itemObjID, err := primitive.ObjectIDFromHex(req.TriggerItemID)
	if err != nil {
		return models.WardrobeItem{}, fmt.Errorf("invalid trigger item id")
	}
	userObjID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return models.WardrobeItem{}, fmt.Errorf("invalid user id")
	}
	col := db.Collection("wardrobeItems")
	var item models.WardrobeItem
	if err := col.FindOne(ctx, bson.M{"_id": itemObjID, "userId": userObjID}).Decode(&item); err != nil {
		return models.WardrobeItem{}, fmt.Errorf("closet item not found: %w", err)
	}
	return item, nil
}

// fetchClosetItems loads all active wardrobe items for the user.
func fetchClosetItems(ctx context.Context, db *mongo.Database, userID string) ([]models.WardrobeItem, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	col := db.Collection("wardrobeItems")
	cursor, err := col.Find(ctx, bson.M{"userId": userObjID, "isActive": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var items []models.WardrobeItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// groupBySlot partitions closet items by outfit slot, excluding the trigger item.
func groupBySlot(items []models.WardrobeItem, triggerID primitive.ObjectID) map[string][]models.WardrobeItem {
	groups := map[string][]models.WardrobeItem{
		"top": {}, "bottom": {}, "full_body": {},
		"outerwear": {}, "footwear": {}, "accessory": {},
	}
	for _, item := range items {
		if item.ID == triggerID {
			continue
		}
		slot := outfitSlot(item.Category)
		if _, ok := groups[slot]; ok {
			groups[slot] = append(groups[slot], item)
		}
	}
	return groups
}

// takeN returns the first n items from the slice (or all items if fewer than n).
func takeN(items []models.WardrobeItem, n int) []models.WardrobeItem {
	if len(items) <= n {
		return items
	}
	return items[:n]
}

// generateCombinations builds candidate outfit combos from the trigger and closet groups.
// Returns the combos and the total number of combinations evaluated.
func generateCombinations(trigger models.WardrobeItem, groups map[string][]models.WardrobeItem, contextFilter string) ([]combo, int) {
	triggerSlot := outfitSlot(trigger.Category)
	evaluated := 0

	// Required complement slot(s) based on the trigger's slot.
	var requiredComplements []models.WardrobeItem
	switch triggerSlot {
	case "top":
		for _, item := range groups["bottom"] {
			if PassesPreFilter(trigger, item) {
				requiredComplements = append(requiredComplements, item)
			}
		}
	case "bottom":
		for _, item := range groups["top"] {
			if PassesPreFilter(trigger, item) {
				requiredComplements = append(requiredComplements, item)
			}
		}
	// full_body, outerwear, footwear, accessory — no required complement
	}

	// Optional slot candidates (capped to avoid combinatorial explosion).
	optionalGroups := map[string][]models.WardrobeItem{}
	for _, slot := range []string{"outerwear", "footwear", "accessory"} {
		var filtered []models.WardrobeItem
		for _, item := range groups[slot] {
			if PassesPreFilter(trigger, item) {
				filtered = append(filtered, item)
			}
		}
		optionalGroups[slot] = takeN(filtered, maxOptionalCandidates)
	}

	// Build base sets: trigger alone (full_body / standalone) or trigger + one required complement.
	var baseSets [][]models.WardrobeItem
	if triggerSlot == "full_body" || triggerSlot == "outerwear" || triggerSlot == "footwear" || triggerSlot == "accessory" {
		baseSets = append(baseSets, []models.WardrobeItem{trigger})
	} else {
		for _, rc := range requiredComplements {
			baseSets = append(baseSets, []models.WardrobeItem{trigger, rc})
		}
		// If no complement found in closet, still produce a trigger-only base so
		// shop-to-complete can fill the gap.
		if len(requiredComplements) == 0 {
			baseSets = append(baseSets, []models.WardrobeItem{trigger})
		}
	}

	var combos []combo

	score := func(set []models.WardrobeItem) combo {
		evaluated++
		s, bd := ScoreOutfit(set, contextFilter)
		return combo{items: set, score: s, bd: bd}
	}

	for _, base := range baseSets {
		combos = append(combos, score(base))

		for _, ow := range optionalGroups["outerwear"] {
			withOW := append(append([]models.WardrobeItem{}, base...), ow)
			combos = append(combos, score(withOW))

			for _, fw := range optionalGroups["footwear"] {
				withOWFW := append(append([]models.WardrobeItem{}, withOW...), fw)
				combos = append(combos, score(withOWFW))
			}
		}

		for _, fw := range optionalGroups["footwear"] {
			withFW := append(append([]models.WardrobeItem{}, base...), fw)
			combos = append(combos, score(withFW))
		}

		for _, acc := range optionalGroups["accessory"] {
			withAcc := append(append([]models.WardrobeItem{}, base...), acc)
			combos = append(combos, score(withAcc))
		}
	}

	return combos, evaluated
}

// deduplicate keeps only the highest-scoring combo per top+bottom pair.
func deduplicate(combos []combo) []combo {
	best := map[string]combo{}
	for _, c := range combos {
		key := comboKey(c)
		if existing, ok := best[key]; !ok || c.score > existing.score {
			best[key] = c
		}
	}
	result := make([]combo, 0, len(best))
	for _, c := range best {
		result = append(result, c)
	}
	return result
}

// whyText generates a template-based explanation for the combination.
func whyText(items []models.WardrobeItem, bd ScoreBreakdown) string {
	if bd.ColorHarmony >= 0.9 {
		colors := []string{}
		for _, item := range items {
			if item.Identifiers.ColorPrimary != "" {
				colors = append(colors, item.Identifiers.ColorPrimary)
			}
		}
		if len(colors) >= 2 {
			return fmt.Sprintf("%s + %s is a high-contrast combination that always works.", colors[0], colors[1])
		}
	}
	if bd.OccasionMatch == 1.0 {
		for _, item := range items {
			if item.Identifiers.Occasion != "" && item.Identifiers.Occasion != "all" {
				return fmt.Sprintf("Every piece is aligned for %s — a cohesive, ready-to-wear look.", item.Identifiers.Occasion)
			}
		}
	}
	if bd.FitCompat >= 0.9 {
		return "The fit balance between these pieces creates a clean, well-proportioned silhouette."
	}
	return "These pieces complement each other across color, fit, and occasion."
}

// ruleTags generates short descriptor tags for the combination.
func ruleTags(items []models.WardrobeItem, contextFilter string) []string {
	tags := []string{}
	colors, fits := []string{}, []string{}
	seenColor, seenFit := map[string]bool{}, map[string]bool{}
	for _, item := range items {
		if c := item.Identifiers.ColorPrimary; c != "" && !seenColor[c] {
			colors = append(colors, c)
			seenColor[c] = true
		}
		if f := item.Identifiers.Fit; f != "" && !seenFit[f] {
			fits = append(fits, f)
			seenFit[f] = true
		}
	}
	if len(colors) >= 2 {
		tags = append(tags, fmt.Sprintf("Color: %s + %s", colors[0], colors[1]))
	}
	if len(fits) >= 2 {
		tags = append(tags, fmt.Sprintf("Fit: %s + %s", fits[0], fits[1]))
	}
	if contextFilter != "" && contextFilter != "all" {
		tags = append(tags, fmt.Sprintf("Occasion: %s", contextFilter))
	}
	return tags
}

// findShopItem queries the catalog for the best-matching product to fill a missing slot.
func findShopItem(ctx context.Context, db *mongo.Database, slot, contextFilter string, trigger models.WardrobeItem) *OutfitItem {
	categoryMap := map[string]string{
		"footwear": "shoes", "outerwear": "outerwear", "accessory": "accessory",
		"top": "upper", "bottom": "lower",
	}
	cat, ok := categoryMap[slot]
	if !ok {
		return nil
	}

	filter := bson.M{"category": cat, "isActive": true}
	if contextFilter != "" && contextFilter != "all" {
		filter["identifiers.occasion"] = bson.M{"$in": []string{contextFilter, "all"}}
	}
	if trigger.Identifiers.ColorPrimary != "" {
		filter["identifiers.colorPrimary"] = trigger.Identifiers.ColorPrimary
	}

	col := db.Collection("products")
	opts := options.FindOne().SetSort(bson.M{"rating": -1})
	var p models.Product
	if err := col.FindOne(ctx, filter, opts).Decode(&p); err != nil {
		// Relax color filter and retry
		delete(filter, "identifiers.colorPrimary")
		if err2 := col.FindOne(ctx, filter, opts).Decode(&p); err2 != nil {
			return nil
		}
	}

	image := ""
	if len(p.Images) > 0 {
		image = p.Images[0]
	}
	return &OutfitItem{
		ItemID:      p.ID,
		Name:        p.Name,
		Brand:       p.Brand,
		Category:    p.Category,
		ImageURL:    image,
		Owned:       false,
		Price:       p.Price,
		Identifiers: p.Identifiers,
	}
}

// GenerateOutfits is the top-level entry point for the recommendation engine.
func GenerateOutfits(ctx context.Context, db *mongo.Database, req RecommendRequest) (RecommendResult, error) {
	start := time.Now()

	if req.Limit <= 0 || req.Limit > 10 {
		req.Limit = 8
	}
	if req.ContextFilter == "" {
		req.ContextFilter = "all"
	}

	trigger, err := fetchTriggerAsWardrobeItem(ctx, db, req)
	if err != nil {
		return RecommendResult{}, err
	}

	closetItems, err := fetchClosetItems(ctx, db, req.UserID)
	if err != nil {
		return RecommendResult{}, err
	}

	groups := groupBySlot(closetItems, trigger.ID)
	combos, evaluated := generateCombinations(trigger, groups, req.ContextFilter)

	afterFilter := len(combos)
	combos = deduplicate(combos)

	sort.Slice(combos, func(i, j int) bool {
		return combos[i].score > combos[j].score
	})
	if len(combos) > req.Limit {
		combos = combos[:req.Limit]
	}

	triggerSlot := outfitSlot(trigger.Category)

	outfits := make([]Outfit, 0, len(combos))
	for rank, c := range combos {
		presentSlots := map[string]bool{}
		outfitItems := make([]OutfitItem, 0, len(c.items))

		for _, item := range c.items {
			outfitItems = append(outfitItems, OutfitItem{
				ItemID:      item.ID.Hex(),
				Name:        item.Name,
				Brand:       item.Brand,
				Category:    item.Category,
				ImageURL:    item.Image,
				IsTrigger:   item.ID == trigger.ID,
				Owned:       true,
				Price:       item.Price,
				Identifiers: item.Identifiers,
			})
			presentSlots[outfitSlot(item.Category)] = true
		}

		var missingItems []OutfitItem
		if req.IncludeShopItems {
			if triggerSlot != "full_body" {
				if !presentSlots["top"] {
					if si := findShopItem(ctx, db, "top", req.ContextFilter, trigger); si != nil {
						missingItems = append(missingItems, *si)
					}
				}
				if !presentSlots["bottom"] {
					if si := findShopItem(ctx, db, "bottom", req.ContextFilter, trigger); si != nil {
						missingItems = append(missingItems, *si)
					}
				}
			}
			if !presentSlots["footwear"] {
				if si := findShopItem(ctx, db, "footwear", req.ContextFilter, trigger); si != nil {
					missingItems = append(missingItems, *si)
				}
			}
		}

		outfits = append(outfits, Outfit{
			OutfitID:       primitive.NewObjectID().Hex(),
			Rank:           rank + 1,
			Score:          c.score,
			RankLabel:      ScoreLabel(c.score),
			Items:          outfitItems,
			ScoreBreakdown: c.bd,
			WhyText:        whyText(c.items, c.bd),
			RuleTags:       ruleTags(c.items, req.ContextFilter),
			MissingItems:   missingItems,
		})
	}

	return RecommendResult{
		Outfits: outfits,
		Meta: RecommendMeta{
			TriggerItemID:            req.TriggerItemID,
			ClosetItemsConsidered:    len(closetItems),
			CombinationsEvaluated:    evaluated,
			CombinationsAfterFilters: afterFilter,
			Returned:                 len(outfits),
			LatencyMS:                time.Since(start).Milliseconds(),
		},
	}, nil
}
