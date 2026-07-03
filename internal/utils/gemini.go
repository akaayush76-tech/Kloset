package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const geminiAPIBase = "https://generativelanguage.googleapis.com/v1beta/models"
const avatarModel = "gemini-2.5-flash-image"

// ---- request structs ----

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig map[string]interface{} `json:"generationConfig"`
}

// ---- response structs ----

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text       string `json:"text"`
				InlineData *struct {
					MimeType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// AvatarGenerationProfile holds optional user measurements used to build a richer prompt.
type AvatarGenerationProfile struct {
	Height            float64
	Weight            float64
	Gender            string
	BodyType          string
	Chest             float64
	Waist             float64
	Hip               float64
	SleeveLength      float64
	Inseam            float64
	CustomDescription string
}

func buildAvatarPrompt(p *AvatarGenerationProfile) string {
	if p == nil || p.Height == 0 {
		return "Generate a photorealistic, full-body fashion avatar of this person on a clean white studio background with soft, even lighting. Neutral standing pose, full body visible from head to feet, no cropping. Fashion photography quality — crisp details. Preserve the person's face, skin tone, hair, and body proportions exactly."
	}

	bmi := p.Weight / ((p.Height / 100) * (p.Height / 100))

	bmiCategory := "healthy weight"
	switch {
	case bmi < 18.5:
		bmiCategory = "lean / underweight"
	case bmi >= 30:
		bmiCategory = "heavy-set"
	case bmi >= 25:
		bmiCategory = "slightly overweight"
	}

	chestFit := "regular"
	if p.Chest > 0 && p.Chest < 85 {
		chestFit = "fitted"
	} else if p.Chest >= 100 {
		chestFit = "relaxed"
	}
	waistFit := "regular"
	if p.Waist > 0 && p.Waist < 70 {
		waistFit = "slim"
	} else if p.Waist >= 90 {
		waistFit = "loose"
	}
	hipFit := "regular"
	if p.Hip > 0 && p.Hip < 90 {
		hipFit = "narrow"
	} else if p.Hip >= 105 {
		hipFit = "wide"
	}

	gender := p.Gender
	if gender == "" {
		gender = "unspecified"
	}
	bodyType := p.BodyType
	if bodyType == "" {
		bodyType = "balanced"
	}

	prompt := fmt.Sprintf(
		"Generate a photorealistic, full-body fashion avatar of this person on a clean white studio background with soft, even lighting.\n\n"+
			"Subject:\n"+
			"- Gender: %s\n"+
			"- Body type: %s (%s)\n"+
			"- Height: %.0f cm | Weight: %.0f kg | BMI: %.1f\n\n"+
			"Body Measurements:\n"+
			"- Chest / Bust: %.0f cm → %s upper-body silhouette\n"+
			"- Natural waist: %.0f cm → %s mid-section\n"+
			"- Hip circumference: %.0f cm → %s lower body\n"+
			"- Sleeve length: %.0f cm\n"+
			"- Inseam: %.0f cm\n\n"+
			"Ensure body proportions precisely reflect all measurements. Dress the avatar in a stylish casual outfit.\n\n"+
			"Pose and Rendering:\n"+
			"- Standing neutral / slight 3/4 pose facing forward\n"+
			"- Full body visible from head to feet (no cropping)\n"+
			"- Fashion photography quality — crisp details\n"+
			"- Match the facial features, skin tone, and hair from the uploaded reference photo",
		gender, bodyType, bmiCategory,
		p.Height, p.Weight, bmi,
		p.Chest, chestFit,
		p.Waist, waistFit,
		p.Hip, hipFit,
		p.SleeveLength, p.Inseam,
	)

	if p.CustomDescription != "" {
		prompt += fmt.Sprintf("\n\nAdditional notes:\n%s", p.CustomDescription)
	}

	return prompt
}

// GenerateAvatarFromPhoto sends a reference photo to Gemini and returns raw PNG bytes.
// photoDataURI must be a base64 data URI ("data:<mimeType>;base64,<data>").
// profile is optional — if non-nil, measurements are woven into the generation prompt.
func GenerateAvatarFromPhoto(ctx context.Context, photoDataURI string, profile *AvatarGenerationProfile) ([]byte, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set")
	}

	if !strings.HasPrefix(photoDataURI, "data:") {
		return nil, fmt.Errorf("photo must be a data URI (data:<mimeType>;base64,<data>)")
	}

	// Parse "data:<mimeType>;base64,<data>"
	withoutPrefix := strings.TrimPrefix(photoDataURI, "data:")
	halves := strings.SplitN(withoutPrefix, ";base64,", 2)
	if len(halves) != 2 {
		return nil, fmt.Errorf("invalid data URI format")
	}
	mimeType, b64Data := halves[0], halves[1]

	prompt := buildAvatarPrompt(profile)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{InlineData: &geminiInlineData{MimeType: mimeType, Data: b64Data}},
					{Text: prompt},
					{Text: "Use this uploaded photo as the reference for the avatar's face, skin tone, and hair."},
				},
			},
		},
		GenerationConfig: map[string]interface{}{
			"responseModalities": []string{"IMAGE"},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gemini request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiAPIBase, avatarModel, apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var gemResp geminiResponse
	if err := json.Unmarshal(body, &gemResp); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if gemResp.Error != nil {
		return nil, fmt.Errorf("gemini error %d: %s", gemResp.Error.Code, gemResp.Error.Message)
	}

	for _, candidate := range gemResp.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != "" {
				imgBytes, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
				if err != nil {
					return nil, fmt.Errorf("failed to decode image bytes: %w", err)
				}
				return imgBytes, nil
			}
		}
	}

	return nil, fmt.Errorf("no image in gemini response")
}
