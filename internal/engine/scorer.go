package engine

import (
	"math"

	"github.com/kloset/backend/internal/models"
)

// Signal weights per spec §3.4 / §4.1.
const (
	weightColor    = 0.40
	weightFit      = 0.30
	weightOccasion = 0.20
	weightSeason   = 0.10
)

// ScoreBreakdown holds per-signal scores (0.0–1.0).
type ScoreBreakdown struct {
	ColorHarmony  float64 `json:"color_harmony"`
	FitCompat     float64 `json:"fit_compat"`
	OccasionMatch float64 `json:"occasion_match"`
	SeasonMatch   float64 `json:"season_match"`
}

// fitCompatScore returns 0.0–1.0 for two fit values per spec §3.4:
// matching fits score highest, intentional contrast slightly lower.
func fitCompatScore(a, b string) float64 {
	if a == "" || b == "" {
		return 0.5
	}
	if a == b {
		if a == "oversized" {
			return 0.3 // oversized + oversized loses all structure
		}
		return 1.0
	}
	if (a == "slim" && b == "tailored") || (a == "tailored" && b == "slim") {
		return 0.9
	}
	// Intentional contrast: one slim/fitted piece against one relaxed/oversized piece.
	slim := map[string]bool{"slim": true, "fitted": true, "tailored": true}
	loose := map[string]bool{"relaxed": true, "oversized": true, "boxy": true}
	if (slim[a] && loose[b]) || (loose[a] && slim[b]) {
		return 0.75
	}
	if a == "regular" || b == "regular" {
		return 0.75 // regular pairs acceptably with everything
	}
	return 0.3 // incompatible proportions
}

// occasionScore measures occasion-tag overlap across the items per spec §3.4:
// full overlap = 1.0, partial overlap = 0.6, no overlap = 0.1.
// Items with an empty or "all" occasion are treated as matching everything.
func occasionScore(items []models.WardrobeItem) float64 {
	occasions := []string{}
	for _, item := range items {
		if o := item.Identifiers.Occasion; o != "" && o != "all" {
			occasions = append(occasions, o)
		}
	}
	if len(occasions) < 2 {
		return 1.0 // zero or one constrained item — nothing to clash with
	}
	first, allMatch, anyMatch := occasions[0], true, false
	for _, o := range occasions[1:] {
		if o == first {
			anyMatch = true
		} else {
			allMatch = false
		}
	}
	switch {
	case allMatch:
		return 1.0
	case anyMatch:
		return 0.6
	default:
		return 0.1
	}
}

// seasonScore measures season-tag overlap per spec §3.4: all-season items
// score 1.0 with any partner; a mismatch between fixed seasons scores 0.4.
func seasonScore(items []models.WardrobeItem) float64 {
	seasons := []string{}
	for _, item := range items {
		if s := item.Identifiers.Season; s != "" && s != "all" {
			seasons = append(seasons, s)
		}
	}
	if len(seasons) < 2 {
		return 1.0
	}
	for _, s := range seasons[1:] {
		if s != seasons[0] {
			return 0.4
		}
	}
	return 1.0
}

// averageColorHarmony computes the mean ColorHarmonyScore across all item pairs.
func averageColorHarmony(items []models.WardrobeItem) float64 {
	if len(items) < 2 {
		return 1.0
	}
	total := 0.0
	pairs := 0
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			a := items[i].Identifiers
			b := items[j].Identifiers
			total += ColorHarmonyScore(a.ColorPrimary, a.ColorTone, b.ColorPrimary, b.ColorTone)
			pairs++
		}
	}
	return total / float64(pairs)
}

// averageFitCompat computes the mean fit compatibility across all item pairs.
func averageFitCompat(items []models.WardrobeItem) float64 {
	if len(items) < 2 {
		return 1.0
	}
	total := 0.0
	pairs := 0
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			total += fitCompatScore(items[i].Identifiers.Fit, items[j].Identifiers.Fit)
			pairs++
		}
	}
	return total / float64(pairs)
}

// ScoreOutfit computes the base 0–100 score and per-signal breakdown for a
// combination. Modifiers (§4.2) are applied separately — see modifiers.go.
func ScoreOutfit(items []models.WardrobeItem) (int, ScoreBreakdown) {
	bd := ScoreBreakdown{
		ColorHarmony:  averageColorHarmony(items),
		FitCompat:     averageFitCompat(items),
		OccasionMatch: occasionScore(items),
		SeasonMatch:   seasonScore(items),
	}
	raw := bd.ColorHarmony*weightColor +
		bd.FitCompat*weightFit +
		bd.OccasionMatch*weightOccasion +
		bd.SeasonMatch*weightSeason
	return clampScore(int(math.Round(raw * 100))), bd
}

func clampScore(score int) int {
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}

// ScoreLabel returns the human-readable rank label for a score.
func ScoreLabel(score int) string {
	switch {
	case score >= 90:
		return "Best match"
	case score >= 75:
		return "Great match"
	case score >= 60:
		return "Good match"
	default:
		return "Fair match"
	}
}
