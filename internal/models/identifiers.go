package models

// ItemIdentifiers holds the 8-dimension identifier set for any clothing item.
type ItemIdentifiers struct {
	ColorPrimary string `bson:"colorPrimary" json:"colorPrimary"` // white, black, navy, beige, grey, brown, red, green, blue, yellow, pink, purple, orange, multicolor
	ColorTone    string `bson:"colorTone"    json:"colorTone"`    // neutral, pastel, neon, earth, bold
	Fit          string `bson:"fit"          json:"fit"`          // slim, regular, oversized, relaxed
	Occasion     string `bson:"occasion"     json:"occasion"`     // casual, smart_casual, date_night, weekend
	Season       string `bson:"season"       json:"season"`       // spring, summer, fall, winter, all
	Formality    string `bson:"formality"    json:"formality"`    // casual, smart_casual, formal
	Style        string `bson:"style"        json:"style"`        // streetwear, classic, bohemian, minimalist
	Pattern      string `bson:"pattern"      json:"pattern"`      // solid, stripes, checks, floral, graphic
}
