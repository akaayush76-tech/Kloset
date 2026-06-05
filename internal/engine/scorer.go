package engine

import (
	"math"
	"time"

	"github.com/kloset/backend/internal/models"
)

const (
	weightColor    = 0.35
	weightFit      = 0.25
	weightOccasion = 0.25
	weightSeason   = 0.15
)

// ScoreBreakdown holds per-signal scores (0.0–1.0).
type ScoreBreakdown struct {
	ColorHarmony  float64 `json:"color_harmony"`
	FitCompat     float64 `json:"fit_compat"`
	OccasionMatch float64 `json:"occasion_match"`
	SeasonMatch   float64 `json:"season_match"`
}

// currentSeason returns the meteorological season for the current month.
func currentSeason() string {
	month := time.Now().Month()
	switch {
	case month >= 3 && month <= 5:
		return "spring"
	case month >= 6 && month <= 8:
		return "summer"
	case month >= 9 && month <= 11:
		return "fall"
	default:
		return "winter"
	}
}

// fitCompatScore returns 0.0–1.0 for two fit values.
func fitCompatScore(a, b string) float64 {
	if a == "" || b == "" {
		return 0.5
	}
	if a == b {
		return 0.85 // matching fit — good but not always the most interesting
	}
	// Complementary fits
	complementary := map[string][]string{
		"slim":     {"regular", "relaxed"},
		"regular":  {"slim", "relaxed", "oversized"},
		"relaxed":  {"slim", "regular"},
		"oversized": {"slim"},
	}
	for _, comp := range complementary[a] {
		if comp == b {
			return 1.0
		}
	}
	// Oversized + oversized = LOW
	if a == "oversized" && b == "oversized" {
		return 0.3
	}
	return 0.4
}

// occasionScore returns the fraction of items in the combination that match contextFilter.
func occasionScore(items []models.WardrobeItem, contextFilter string) float64 {
	if contextFilter == "all" || contextFilter == "" {
		return 1.0
	}
	if len(items) == 0 {
		return 0.0
	}
	matched := 0
	for _, item := range items {
		if item.Identifiers.Occasion == contextFilter || item.Identifiers.Occasion == "all" || item.Identifiers.Occasion == "" {
			matched++
		}
	}
	return float64(matched) / float64(len(items))
}

// seasonScore returns the fraction of items whose season matches the current season.
func seasonScore(items []models.WardrobeItem) float64 {
	if len(items) == 0 {
		return 0.0
	}
	season := currentSeason()
	matched := 0
	for _, item := range items {
		s := item.Identifiers.Season
		if s == season || s == "all" || s == "" {
			matched++
		}
	}
	return float64(matched) / float64(len(items))
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

// ScoreOutfit computes the final 0–100 score and per-signal breakdown for a combination.
func ScoreOutfit(items []models.WardrobeItem, contextFilter string) (int, ScoreBreakdown) {
	bd := ScoreBreakdown{
		ColorHarmony:  averageColorHarmony(items),
		FitCompat:     averageFitCompat(items),
		OccasionMatch: occasionScore(items, contextFilter),
		SeasonMatch:   seasonScore(items),
	}
	raw := bd.ColorHarmony*weightColor +
		bd.FitCompat*weightFit +
		bd.OccasionMatch*weightOccasion +
		bd.SeasonMatch*weightSeason
	score := int(math.Round(raw * 100))
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score, bd
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
