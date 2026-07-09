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

// EditImageForTryOn sends the avatar image, the product image, and an instruction
// prompt to Gemini and returns the resulting PNG bytes.
// avatarImage and productImage may each be a CDN/HTTP URL or a base64 data URI.
func EditImageForTryOn(ctx context.Context, avatarImage, productImage, prompt, productName, productColor, productCategory string) ([]byte, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set")
	}

	avatarMime, avatarB64, err := resolveImageToBase64(ctx, avatarImage)
	if err != nil {
		return nil, fmt.Errorf("failed to load avatar image: %w", err)
	}

	productMime, productB64, err := resolveImageToBase64(ctx, productImage)
	if err != nil {
		return nil, fmt.Errorf("failed to load product image: %w", err)
	}

	// Build a specific label for the product image so Gemini knows exactly
	// what it is looking at before processing the image.
	itemType := productCategory
	if itemType == "" {
		itemType = "clothing item"
	}
	var extractHint string
	switch productCategory {
	case "shoes":
		extractHint = "Focus on the shoe only: its silhouette (high-top/low-top), sole, upper material, lace style, logo, and any branding. Ignore any foot or person in the photo."
	case "lower":
		extractHint = "Focus on the garment only: its cut, waistband, pockets, hem, fabric, and color. Ignore any person wearing it."
	default:
		extractHint = "Focus on the garment only: its color, pattern, logo, texture, collar, sleeves, and hem. Ignore any person wearing it."
	}
	productLabel := fmt.Sprintf(
		"IMAGE 2 — %s to try on (visual reference takes priority over any text description): \"%s\"",
		itemType, productName,
	)
	if productColor != "" {
		productLabel += fmt.Sprintf(", %s", productColor)
	}
	productLabel += ". " + extractHint + " Reproduce EXACTLY what you see in this image."

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: "IMAGE 1 — The person to dress (do not change their face, body, or background):"},
					{InlineData: &geminiInlineData{MimeType: avatarMime, Data: avatarB64}},
					{Text: productLabel},
					{InlineData: &geminiInlineData{MimeType: productMime, Data: productB64}},
					{Text: prompt},
				},
			},
		},
		GenerationConfig: map[string]interface{}{
			"responseModalities": []string{"IMAGE", "TEXT"},
			// Lower temperature = more faithful to input images, less hallucination
			"temperature": 0.2,
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

	// Collect any text parts for diagnostics, look for an image part
	var textParts []string
	for _, candidate := range gemResp.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != "" {
				imgBytes, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
				if err != nil {
					return nil, fmt.Errorf("failed to decode result image: %w", err)
				}
				return imgBytes, nil
			}
			if part.Text != "" {
				textParts = append(textParts, part.Text)
			}
		}
	}

	// No image found — include Gemini's text response in the error so it's visible in logs
	if len(textParts) > 0 {
		return nil, fmt.Errorf("gemini returned no image; text response: %s", strings.Join(textParts, " | "))
	}
	return nil, fmt.Errorf("gemini returned no image and no text (empty response); raw: %s", string(body))
}

// resolveImageToBase64 accepts either an HTTP(S) URL or a base64 data URI and
// returns the MIME type and raw base64-encoded image bytes.
func resolveImageToBase64(ctx context.Context, image string) (mimeType string, b64Data string, err error) {
	if strings.HasPrefix(image, "data:") {
		// data:<mimeType>;base64,<data>
		withoutPrefix := strings.TrimPrefix(image, "data:")
		halves := strings.SplitN(withoutPrefix, ";base64,", 2)
		if len(halves) != 2 {
			return "", "", fmt.Errorf("invalid data URI format")
		}
		return halves[0], halves[1], nil
	}

	// Treat as an HTTP URL — download it
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, image, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("image download returned HTTP %d", resp.StatusCode)
	}

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read image bytes: %w", err)
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	// Strip parameters like "; charset=utf-8"
	if idx := strings.Index(ct, ";"); idx != -1 {
		ct = strings.TrimSpace(ct[:idx])
	}

	return ct, base64.StdEncoding.EncodeToString(rawBytes), nil
}
