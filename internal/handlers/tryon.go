package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/utils"
)

// TryOnRequest is the request body for POST /api/tryon.
type TryOnRequest struct {
	// AvatarImage is either a CDN URL or a base64 data URI of the user's avatar.
	AvatarImage string `json:"avatarImage" binding:"required"`
	// ProductImage is the URL of the product photo to overlay onto the avatar.
	ProductImage string `json:"productImage" binding:"required"`
	// Prompt is the fully-formed try-on instruction built by the frontend.
	Prompt string `json:"prompt" binding:"required"`
}

// TryOnHandler generates a try-on image by overlaying clothing described in
// Prompt onto AvatarImage using Gemini, then returns the result as a base64
// data URI so the frontend can display it immediately (no Cloudinary upload —
// try-on results are ephemeral session data).
// POST /api/tryon
func TryOnHandler(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req TryOnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Allow up to 90 s for Gemini image generation
	genCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	pngBytes, err := utils.EditImageForTryOn(genCtx, req.AvatarImage, req.ProductImage, req.Prompt)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Try-on generation failed", err)
		return
	}

	// Upload to Cloudinary so the result survives beyond the response and can
	// be shared / referenced later without embedding a large base64 blob.
	uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer uploadCancel()

	uploadResult, err := utils.UploadReaderToCloudinary(uploadCtx, bytes.NewReader(pngBytes), "kloset/tryon")
	if err != nil {
		// Cloudinary unavailable (e.g. no credentials in dev) — fall back to
		// returning the image inline as a data URI so dev flow still works.
		b64 := base64.StdEncoding.EncodeToString(pngBytes)
		utils.SuccessResponse(c, http.StatusOK, "Try-on generated", gin.H{
			"resultUrl": "data:image/png;base64," + b64,
		})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Try-on generated", gin.H{
		"resultUrl": uploadResult.URL,
	})
}

// DeleteTryOnRequest is the request body for DELETE /api/tryon.
type DeleteTryOnRequest struct {
	// URLs are the Cloudinary result URLs to delete (data URIs are silently skipped).
	URLs []string `json:"urls" binding:"required,min=1"`
}

// DeleteTryOnHandler deletes one or more try-on result images from Cloudinary.
// DELETE /api/tryon
func DeleteTryOnHandler(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req DeleteTryOnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var failed []string
	for _, url := range req.URLs {
		publicID, ok := cloudinaryPublicID(url)
		if !ok {
			continue // data URI or non-Cloudinary URL — nothing to delete
		}
		if err := utils.DeleteFromCloudinary(ctx, publicID); err != nil {
			failed = append(failed, publicID)
		}
	}

	if len(failed) > 0 {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError,
			"Some images could not be deleted from Cloudinary", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Try-on images deleted", nil)
}

// cloudinaryPublicID extracts the public ID from a Cloudinary URL.
// e.g. https://res.cloudinary.com/demo/image/upload/v123/kloset/tryon/abc.png → kloset/tryon/abc
func cloudinaryPublicID(rawURL string) (string, bool) {
	const marker = "/upload/"
	idx := strings.Index(rawURL, marker)
	if idx == -1 {
		return "", false
	}
	path := rawURL[idx+len(marker):]

	// Strip optional version segment "v1234567890/"
	if len(path) > 1 && path[0] == 'v' {
		if slash := strings.Index(path, "/"); slash != -1 {
			if _, err := strconv.Atoi(path[1:slash]); err == nil {
				path = path[slash+1:]
			}
		}
	}

	// Strip file extension
	if dot := strings.LastIndex(path, "."); dot != -1 {
		path = path[:dot]
	}

	return path, path != ""
}
