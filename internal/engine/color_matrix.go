package engine

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// Tier base scores per spec §3.5. Overridable via the JSON config file.
const (
	defaultScoreHigh    = 1.0
	defaultScoreMed     = 0.75
	defaultScoreLow     = 0.35
	defaultScoreBlocked = 0.10
	defaultScoreUnknown = 0.40
)

// colorMatrixConfig is the on-disk shape of config/color_matrix.json.
// The matrix is stored as config (not code) so the product/design team can
// tune pairs without a deploy — see the spec's engineering note in §3.5.
type colorMatrixConfig struct {
	Tiers map[string]float64 `json:"tiers"` // high, med, low, blocked, unknown
	Pairs map[string][][2]string `json:"pairs"` // tier name → list of color pairs
}

var (
	colorMatrixMu sync.RWMutex
	colorScores   map[string]map[string]float64
	unknownScore  = defaultScoreUnknown
)

// defaultColorConfig is the embedded fallback used when no config file exists.
var defaultColorConfig = colorMatrixConfig{
	Tiers: map[string]float64{
		"high":    defaultScoreHigh,
		"med":     defaultScoreMed,
		"low":     defaultScoreLow,
		"blocked": defaultScoreBlocked,
		"unknown": defaultScoreUnknown,
	},
	Pairs: map[string][][2]string{
		"high": {
			{"white", "navy"}, {"white", "black"}, {"white", "beige"}, {"white", "grey"},
			{"white", "brown"}, {"white", "blue"}, {"black", "grey"}, {"black", "beige"},
			{"black", "navy"}, {"black", "red"}, {"navy", "beige"}, {"navy", "grey"},
			{"beige", "brown"}, {"brown", "navy"},
		},
		"med": {
			{"blue", "grey"}, {"blue", "white"}, {"blue", "beige"}, {"red", "navy"},
			{"red", "grey"}, {"green", "beige"}, {"green", "brown"}, {"pink", "grey"},
			{"pink", "white"}, {"pink", "navy"}, {"purple", "grey"}, {"purple", "navy"},
			{"orange", "navy"}, {"orange", "brown"}, {"yellow", "navy"}, {"yellow", "grey"},
			{"multicolor", "white"}, {"multicolor", "black"}, {"multicolor", "navy"},
			// Monochrome — valid, not exciting
			{"white", "white"}, {"black", "black"}, {"navy", "navy"}, {"grey", "grey"},
		},
		"low": {
			{"red", "orange"}, {"red", "pink"}, {"red", "red"}, {"orange", "pink"},
			{"yellow", "orange"}, {"neon", "neon"}, {"multicolor", "multicolor"},
		},
		"blocked": {
			{"brown", "black"}, {"red", "green"},
		},
	},
}

func init() {
	applyColorConfig(defaultColorConfig)
}

// applyColorConfig rebuilds the symmetric score lookup from a config.
func applyColorConfig(cfg colorMatrixConfig) {
	scores := make(map[string]map[string]float64)
	set := func(a, b string, s float64) {
		if scores[a] == nil {
			scores[a] = make(map[string]float64)
		}
		scores[a][b] = s
	}
	tierScore := func(name string, fallback float64) float64 {
		if v, ok := cfg.Tiers[name]; ok {
			return v
		}
		return fallback
	}
	for tier, pairs := range cfg.Pairs {
		s := tierScore(tier, defaultScoreUnknown)
		for _, p := range pairs {
			set(p[0], p[1], s)
			set(p[1], p[0], s)
		}
	}

	colorMatrixMu.Lock()
	defer colorMatrixMu.Unlock()
	colorScores = scores
	unknownScore = tierScore("unknown", defaultScoreUnknown)
}

// LoadColorMatrix loads the color matrix from a JSON config file, replacing
// the embedded defaults. A missing file is not an error — the defaults stay.
func LoadColorMatrix(path string) {
	if path == "" {
		path = "config/color_matrix.json"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Printf("color matrix: using embedded defaults (%s not readable: %v)", path, err)
		return
	}
	var cfg colorMatrixConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		log.Printf("color matrix: invalid config %s (%v) — keeping embedded defaults", path, err)
		return
	}
	applyColorConfig(cfg)
	log.Printf("color matrix: loaded %s", path)
}

// ColorHarmonyScore returns a 0.0–1.0 compatibility score for two items
// based on their color_primary and color_tone identifiers.
func ColorHarmonyScore(primaryA, toneA, primaryB, toneB string) float64 {
	colorMatrixMu.RLock()
	base := unknownScore
	if inner, ok := colorScores[primaryA]; ok {
		if s, ok := inner[primaryB]; ok {
			base = s
		}
	}
	colorMatrixMu.RUnlock()

	// Tone modifier per spec §3.5: matching tones +0.1; clashing tones −0.1
	// (both neon, or both pastel but different color family).
	if toneA != "" && toneB != "" {
		switch {
		case toneA == "neon" && toneB == "neon":
			base -= 0.1
		case toneA == "pastel" && toneB == "pastel" && primaryA != primaryB:
			base -= 0.1
		case toneA == toneB:
			base += 0.1
		}
	}

	if base > 1.0 {
		base = 1.0
	}
	if base < 0.0 {
		base = 0.0
	}
	return base
}
