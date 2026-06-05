package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/utils"
)

// UploadImageRequest represents image upload response
type UploadImageResponse struct {
	URL          string `json:"url"`
	OriginalName string `json:"originalName"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mimetype"`
	PublicID     string `json:"publicId"`
}

// UploadImageHandler handles single image upload
func UploadImageHandler(c *gin.Context) {
	// Parse multipart form
	header, err := c.FormFile("image")
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "No image file provided", err)
		return
	}

	// Check file size (10MB limit)
	maxFileSize := int64(10485760) // 10MB
	if header.Size > maxFileSize {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "File size exceeds 10MB limit", nil)
		return
	}

	// In production, upload to Cloudinary
	// For now, return mock response
	response := UploadImageResponse{
		URL:          fmt.Sprintf("https://via.placeholder.com/400?text=%s", header.Filename),
		OriginalName: header.Filename,
		Size:         header.Size,
		MimeType:     header.Header.Get("Content-Type"),
		PublicID:     "mock-public-id",
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

	var responses []UploadImageResponse
	maxFileSize := int64(10485760) // 10MB

	for _, file := range files {
		if file.Size > maxFileSize {
			utils.HTTPErrorHandler(c, http.StatusBadRequest, fmt.Sprintf("File %s exceeds 10MB limit", file.Filename), nil)
			return
		}

		response := UploadImageResponse{
			URL:          fmt.Sprintf("https://via.placeholder.com/400?text=%s", file.Filename),
			OriginalName: file.Filename,
			Size:         file.Size,
			MimeType:     file.Header.Get("Content-Type"),
			PublicID:     fmt.Sprintf("mock-%s", file.Filename),
		}
		responses = append(responses, response)
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

	maxFileSize := int64(10485760) // 10MB
	if header.Size > maxFileSize {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "File size exceeds 10MB limit", nil)
		return
	}

	// Mock response - in production, upload to Cloudinary with wardrobe folder
	response := UploadImageResponse{
		URL:          fmt.Sprintf("https://via.placeholder.com/400?text=%s", header.Filename),
		OriginalName: header.Filename,
		Size:         header.Size,
		MimeType:     header.Header.Get("Content-Type"),
		PublicID:     fmt.Sprintf("wardrobe/%s-%s", userID, header.Filename),
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

	maxFileSize := int64(10485760) // 10MB
	if header.Size > maxFileSize {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "File size exceeds 10MB limit", nil)
		return
	}

	// Mock response - in production, upload to Cloudinary with product folder
	response := UploadImageResponse{
		URL:          fmt.Sprintf("https://via.placeholder.com/400?text=%s", header.Filename),
		OriginalName: header.Filename,
		Size:         header.Size,
		MimeType:     header.Header.Get("Content-Type"),
		PublicID:     fmt.Sprintf("products/%s", header.Filename),
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

	// In production, delete from Cloudinary
	// For now, return mock success response

	utils.SuccessResponse(c, http.StatusOK, "Image deleted successfully", nil)
}

// OptimizeImageRequest represents image optimization request
type OptimizeImageRequest struct {
	ImageURL string `json:"imageUrl" binding:"required"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Quality  int    `json:"quality"`
	Format   string `json:"format"`
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

	// In production, use Cloudinary transformation URL
	// Mock response
	optimizedURL := fmt.Sprintf("%s?w=%s&h=%s&q=%s&f=%s", imageURL, width, height, quality, format)

	utils.SuccessResponse(c, http.StatusOK, "Optimized image URL generated", gin.H{
		"url": optimizedURL,
	})
}
