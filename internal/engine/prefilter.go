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

// ethnicClash implements BLOCK-03: ethnic items cannot mix with Western items
// unless both are tagged fusion.
func ethnicClash(a, b models.ItemIdentifiers) bool {
	isEthnic := func(id models.ItemIdentifiers) bool { return id.Style == "ethnic" }
	isFusion := func(id models.ItemIdentifiers) bool { return id.Style == "fusion" }
	if isEthnic(a) && !isEthnic(b) {
		return !isFusion(b)
	}
	if isEthnic(b) && !isEthnic(a) {
		return !isFusion(a)
	}
	return false
}

// boldPattern reports whether a pattern counts as bold for BLOCK-05.
// Solid (or untagged) items are never bold.
func boldPattern(pattern string) bool {
	return pattern != "" && pattern != "solid"
}

// seasonClash returns true when two non-"all" seasons are polar opposites.
func seasonClash(a, b string) bool {
	if a == "" || b == "" || a == "all" || b == "all" {
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

	// Formality extreme clash (casual ↔ formal) — BLOCK-01
	if formalityDistance(trigger.Identifiers.Formality, candidate.Identifiers.Formality) {
		return false
	}

	// Ethnic ↔ Western mixing — BLOCK-03
	if ethnicClash(trigger.Identifiers, candidate.Identifiers) {
		return false
	}

	// Two bold patterns in one outfit — BLOCK-05 (solid + pattern is fine)
	if boldPattern(trigger.Identifiers.Pattern) && boldPattern(candidate.Identifiers.Pattern) {
		return false
	}

	return true
}
