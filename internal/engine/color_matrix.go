package engine

// harmonyTier represents color pair compatibility.
type harmonyTier int

const (
	harmonyLow  harmonyTier = 0
	harmonyMed  harmonyTier = 1
	harmonyHigh harmonyTier = 2
)

// colorMatrix maps color_primary pairs → harmony tier.
// The matrix is symmetric; both directions are stored at init time.
var colorMatrix map[string]map[string]harmonyTier

func init() {
	pairs := []struct {
		a, b string
		tier harmonyTier
	}{
		// High harmony pairs
		{"white", "navy", harmonyHigh},
		{"white", "black", harmonyHigh},
		{"white", "beige", harmonyHigh},
		{"white", "grey", harmonyHigh},
		{"white", "brown", harmonyHigh},
		{"white", "blue", harmonyHigh},
		{"black", "grey", harmonyHigh},
		{"black", "white", harmonyHigh},
		{"black", "beige", harmonyHigh},
		{"black", "navy", harmonyHigh},
		{"black", "red", harmonyHigh},
		{"navy", "beige", harmonyHigh},
		{"navy", "white", harmonyHigh},
		{"navy", "grey", harmonyHigh},
		{"beige", "brown", harmonyHigh},
		{"beige", "white", harmonyHigh},
		{"beige", "navy", harmonyHigh},
		{"grey", "navy", harmonyHigh},
		{"grey", "black", harmonyHigh},
		{"grey", "white", harmonyHigh},
		{"brown", "beige", harmonyHigh},
		{"brown", "white", harmonyHigh},
		{"brown", "navy", harmonyHigh},
		// Med harmony pairs
		{"blue", "grey", harmonyMed},
		{"blue", "white", harmonyMed},
		{"blue", "beige", harmonyMed},
		{"red", "navy", harmonyMed},
		{"red", "grey", harmonyMed},
		{"green", "beige", harmonyMed},
		{"green", "brown", harmonyMed},
		{"pink", "grey", harmonyMed},
		{"pink", "white", harmonyMed},
		{"pink", "navy", harmonyMed},
		{"purple", "grey", harmonyMed},
		{"purple", "navy", harmonyMed},
		{"orange", "navy", harmonyMed},
		{"orange", "brown", harmonyMed},
		{"yellow", "navy", harmonyMed},
		{"yellow", "grey", harmonyMed},
		{"multicolor", "white", harmonyMed},
		{"multicolor", "black", harmonyMed},
		{"multicolor", "navy", harmonyMed},
		// Same-color (monochrome — valid, not exciting)
		{"white", "white", harmonyMed},
		{"black", "black", harmonyMed},
		{"navy", "navy", harmonyMed},
		{"grey", "grey", harmonyMed},
		// Low harmony pairs
		{"red", "orange", harmonyLow},
		{"red", "pink", harmonyLow},
		{"red", "red", harmonyLow},
		{"orange", "pink", harmonyLow},
		{"yellow", "orange", harmonyLow},
		{"neon", "neon", harmonyLow},
		{"multicolor", "multicolor", harmonyLow},
	}

	colorMatrix = make(map[string]map[string]harmonyTier)
	for _, p := range pairs {
		if colorMatrix[p.a] == nil {
			colorMatrix[p.a] = make(map[string]harmonyTier)
		}
		if colorMatrix[p.b] == nil {
			colorMatrix[p.b] = make(map[string]harmonyTier)
		}
		colorMatrix[p.a][p.b] = p.tier
		colorMatrix[p.b][p.a] = p.tier
	}
}

// ColorHarmonyScore returns a 0.0–1.0 compatibility score for two items
// based on their color_primary and color_tone identifiers.
func ColorHarmonyScore(primaryA, toneA, primaryB, toneB string) float64 {
	var base float64
	if inner, ok := colorMatrix[primaryA]; ok {
		if tier, ok := inner[primaryB]; ok {
			switch tier {
			case harmonyHigh:
				base = 1.0
			case harmonyMed:
				base = 0.6
			case harmonyLow:
				base = 0.2
			}
		} else {
			base = 0.4 // unknown pair → below-average default
		}
	} else {
		base = 0.4
	}

	// Tone modifier
	if toneA != "" && toneB != "" {
		if toneA == toneB {
			base += 0.1 // matching tones
		} else if (toneA == "neon" && toneB == "neon") ||
			(toneA == "pastel" && toneB == "pastel") {
			base -= 0.1 // both neon or both pastel but already handled by same-tone above; defensive
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
