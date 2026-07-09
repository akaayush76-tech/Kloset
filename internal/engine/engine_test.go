package engine

import (
	"testing"

	"github.com/kloset/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func item(category string, ids models.ItemIdentifiers) models.WardrobeItem {
	return models.WardrobeItem{ID: primitive.NewObjectID(), Category: category, Identifiers: ids}
}

func TestPreFilterPatternClash(t *testing.T) {
	floralTop := item("upper", models.ItemIdentifiers{Pattern: "floral"})
	checkedBottom := item("lower", models.ItemIdentifiers{Pattern: "checks"})
	solidBottom := item("lower", models.ItemIdentifiers{Pattern: "solid"})

	if PassesPreFilter(floralTop, checkedBottom) {
		t.Error("BLOCK-05: two bold patterns must be blocked")
	}
	if !PassesPreFilter(floralTop, solidBottom) {
		t.Error("solid + pattern must be allowed")
	}
}

func TestPreFilterEthnicMixing(t *testing.T) {
	kurta := item("upper", models.ItemIdentifiers{Style: "ethnic"})
	jeans := item("lower", models.ItemIdentifiers{Style: "streetwear"})
	fusionJeans := item("lower", models.ItemIdentifiers{Style: "fusion"})

	if PassesPreFilter(kurta, jeans) {
		t.Error("BLOCK-03: ethnic + western must be blocked")
	}
	if !PassesPreFilter(kurta, fusionJeans) {
		t.Error("ethnic + fusion must be allowed")
	}
}

func TestScoreWeightsPerSpec(t *testing.T) {
	if weightColor != 0.40 || weightFit != 0.30 || weightOccasion != 0.20 || weightSeason != 0.10 {
		t.Errorf("weights must be 40/30/20/10 per spec, got %v/%v/%v/%v",
			weightColor, weightFit, weightOccasion, weightSeason)
	}
}

func TestColorMatrixBlockedPair(t *testing.T) {
	s := ColorHarmonyScore("brown", "", "black", "")
	if s > 0.11 {
		t.Errorf("brown+black is a blocked pair, want <=0.10, got %v", s)
	}
	s = ColorHarmonyScore("white", "", "navy", "")
	if s != 1.0 {
		t.Errorf("white+navy is high harmony, want 1.0, got %v", s)
	}
}

func TestOccasionOverlap(t *testing.T) {
	full := []models.WardrobeItem{
		item("upper", models.ItemIdentifiers{Occasion: "casual"}),
		item("lower", models.ItemIdentifiers{Occasion: "casual"}),
	}
	if got := occasionScore(full); got != 1.0 {
		t.Errorf("full overlap want 1.0, got %v", got)
	}
	none := []models.WardrobeItem{
		item("upper", models.ItemIdentifiers{Occasion: "date_night"}),
		item("lower", models.ItemIdentifiers{Occasion: "weekend"}),
	}
	if got := occasionScore(none); got != 0.1 {
		t.Errorf("no overlap want 0.1, got %v", got)
	}
}

func TestModifiers(t *testing.T) {
	top := item("upper", models.ItemIdentifiers{Occasion: "casual"})
	bottom := item("lower", models.ItemIdentifiers{Occasion: "casual"})
	shoes := item("shoes", models.ItemIdentifiers{})

	c := combo{items: []models.WardrobeItem{top, bottom, shoes}}
	mc := modifierContext{contextFilter: "casual"}

	// completeness +5, context match +3
	if got := applyModifiers(50, c, mc); got != 58 {
		t.Errorf("want 58 (50 +5 completeness +3 context), got %d", got)
	}

	// wear history +4 on top of that
	mc.wornComboKeys = map[string]bool{comboKey(c): true}
	if got := applyModifiers(50, c, mc); got != 62 {
		t.Errorf("want 62 with wear-history bonus, got %d", got)
	}

	// missing bottom → -5, no completeness
	incomplete := combo{items: []models.WardrobeItem{top}}
	if got := applyModifiers(50, incomplete, modifierContext{}); got != 45 {
		t.Errorf("want 45 with missing-item penalty, got %d", got)
	}

	// low-confidence tag → -3
	lowConf := item("lower", models.ItemIdentifiers{TagConfidence: 0.5})
	c2 := combo{items: []models.WardrobeItem{top, lowConf}}
	if got := applyModifiers(50, c2, modifierContext{}); got != 47 {
		t.Errorf("want 47 with low-confidence penalty, got %d", got)
	}

	// avatar fit boost: all catalog items fitted → +8
	wishTop := item("upper", models.ItemIdentifiers{})
	wishTop.CatalogProductID = "prod-1"
	c3 := combo{items: []models.WardrobeItem{wishTop, bottom}}
	mc3 := modifierContext{fittedProductIDs: map[string]bool{"prod-1": true}}
	if got := applyModifiers(50, c3, mc3); got != 58 {
		t.Errorf("want 58 with avatar-fit boost, got %d", got)
	}
}
