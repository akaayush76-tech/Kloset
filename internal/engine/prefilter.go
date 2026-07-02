package engine

import "github.com/kloset/backend/internal/models"

// outfitSlot maps a category to its logical outfit slot.
func outfitSlot(category string) string {
	switch category {
	case "upper":
		return "top"
	case "lower":
		return "bottom"
	case "full_body":
		return "full_body"
	case "outerwear":
		return "outerwear"
	case "shoes":
		return "footwear"
	case "accessory":
		return "accessory"
	default:
		return "unknown"
	}
}

// formalityDistance returns true when two formality levels are incompatibly far apart.
func formalityDistance(a, b string) bool {
	order := map[string]int{"casual": 0, "smart_casual": 1, "formal": 2}
	da, oka := order[a]
	db, okb := order[b]
	if !oka || !okb {
		return false
	}
	diff := da - db
	if diff < 0 {
		diff = -diff
	}
	return diff >= 2 // casual + formal = blocked; smart_casual bridges both
}

// seasonClash returns true when two non-"all" seasons are polar opposites.
func seasonClash(a, b string) bool {
	if a == "all" || b == "all" {
		return false
	}
	opposite := map[string]string{
		"summer": "winter",
		"winter": "summer",
	}
	return opposite[a] == b
}

// PassesPreFilter returns true if candidate can appear in an outfit with the trigger.
// Any false means the combination is a hard block and must be discarded.
func PassesPreFilter(trigger, candidate models.WardrobeItem) bool {
	triggerSlot := outfitSlot(trigger.Category)
	candidateSlot := outfitSlot(candidate.Category)

	// Same slot — can't combine two bottoms, two tops, etc.
	if triggerSlot == candidateSlot {
		return false
	}

	// Full-body item + top or bottom = invalid (full-body already covers both)
	if triggerSlot == "full_body" && (candidateSlot == "top" || candidateSlot == "bottom") {
		return false
	}
	if candidateSlot == "full_body" && (triggerSlot == "top" || triggerSlot == "bottom") {
		return false
	}

	// Season hard mismatch (summer ↔ winter)
	if seasonClash(trigger.Identifiers.Season, candidate.Identifiers.Season) {
		return false
	}

	// Formality extreme clash (casual ↔ formal)
	if formalityDistance(trigger.Identifiers.Formality, candidate.Identifiers.Formality) {
		return false
	}

	return true
}
