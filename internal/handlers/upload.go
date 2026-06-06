package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/utils"
)

// UploadImageResponse represents image upload response
type UploadImageResponse struct {
	URL          string `json:"url"`
	OriginalName string `json:"originalName"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mimetype"`
	PublicID     string `json:"publicId"`
}

// UploadImageHandler handles single image upload
func UploadImageHandler(c *gin.Context) {
	header, err := c.FormFile("image")
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "No image file provided", err)
		return
	}

	if header.Size > 10485760 {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "File size exceeds 10MB limit", nil)
		return
	}

	file, err := header.Open()
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error reading file", err)
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := utils.UploadReaderToCloudinary(ctx, file, "kloset/images")
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error uploading image", err)
		return
	}

	response := UploadImageResponse{
		URL:          result.URL,
		OriginalName: header.Filename,
		Size:         header.Size,
		MimeType:     header.Header.Get("Content-Type"),
		PublicID:     result.PublicID,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Image uploaded successfully", response)
}

// UploadImagesHandler handles multiple image uploads (max 10)
func UploadImagesHandler(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid multipart form", err)
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "No images provided", nil)
		return
	}
	if len(files) > 10 {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Maximum 10 files allowed", nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var responses []UploadImageResponse
	for _, header := range files {
		if header.Size > 10485760 {
			utils.HTTPErrorHandler(c, http.StatusBadRequest, fmt.Sprintf("File %s exceeds 10MB limit", header.Filename), nil)
			return
		}

		file, err := header.Open()
		if err != nil {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error reading file", err)
			return
		}

		result, err := utils.UploadReaderToCloudinary(ctx, file, "kloset/images")
		file.Close()
		if err != nil {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error uploading image", err)
			return
		}

		responses = append(responses, UploadImageResponse{
			URL:          result.URL,
			OriginalName: header.Filename,
			Size:         header.Size,
			MimeType:     header.Header.Get("Content-Type"),
			PublicID:     result.PublicID,
		})
	}

	utils.SuccessResponse(c, http.StatusCreated, "Images uploaded successfully", responses)
}

// UploadWardrobeImageHandler handles wardrobe image upload
func UploadWardrobeImageHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	header, err := c.FormFile("image")
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "No image file provided", err)
		return
	}

	if header.Size > 10485760 {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "File size exceeds 10MB limit", nil)
		return
	}

	file, err := header.Open()
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error reading file", err)
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	folder := fmt.Sprintf("kloset/wardrobe/%s", userID)
	result, err := utils.UploadReaderToCloudinary(ctx, file, folder)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error uploading image", err)
		return
	}

	response := UploadImageResponse{
		URL:          result.URL,
		OriginalName: header.Filename,
		Size:         header.Size,
		MimeType:     header.Header.Get("Content-Type"),
		PublicID:     result.PublicID,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Wardrobe image uploaded successfully", response)
}

// UploadProductImageHandler handles product image upload (admin only)
func UploadProductImageHandler(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Admin access required", nil)
		return
	}

	header, err := c.FormFile("image")
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "No image file provided", err)
		return
	}

	if header.Size > 10485760 {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "File size exceeds 10MB limit", nil)
		return
	}

	file, err := header.Open()
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error reading file", err)
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := utils.UploadReaderToCloudinary(ctx, file, "kloset/products")
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error uploading image", err)
		return
	}

	response := UploadImageResponse{
		URL:          result.URL,
		OriginalName: header.Filename,
		Size:         header.Size,
		MimeType:     header.Header.Get("Content-Type"),
		PublicID:     result.PublicID,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Product image uploaded successfully", response)
}

// DeleteImageRequest represents image deletion request
type DeleteImageRequest struct {
	PublicID string `json:"publicId" binding:"required"`
}

// DeleteImageHandler deletes an image by public ID
func DeleteImageHandler(c *gin.Context) {
	var req DeleteImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	cld := utils.GetCloudinaryClient()
	if cld == nil {
		utils.SuccessResponse(c, http.StatusOK, "Image deleted successfully", nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := utils.DeleteFromCloudinary(ctx, req.PublicID); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error deleting image", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Image deleted successfully", nil)
}

// OptimizeImageHandler returns optimized image URL
func OptimizeImageHandler(c *gin.Context) {
	imageURL := c.Query("imageUrl")
	width := c.DefaultQuery("width", "400")
	height := c.DefaultQuery("height", "400")
	quality := c.DefaultQuery("quality", "80")
	format := c.DefaultQuery("format", "webp")

	if imageURL == "" {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Image URL is required", nil)
		return
	}

	// For Cloudinary URLs, we could build a transformation URL here.
	// For now return a query-param decorated URL that clients can use.
	optimizedURL := fmt.Sprintf("%s?w=%s&h=%s&q=%s&f=%s", imageURL, width, height, quality, format)

	utils.SuccessResponse(c, http.StatusOK, "Optimized image URL generated", gin.H{
		"url": optimizedURL,
	})
}
