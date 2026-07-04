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
	"github.com/kloset/backend/internal/config"
	"github.com/kloset/backend/internal/models"
	"github.com/kloset/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---- Request / Response types ----

type TryOnRequest struct {
	AvatarImage     string `json:"avatarImage" binding:"required"`
	ProductImage    string `json:"productImage" binding:"required"`
	Prompt          string `json:"prompt" binding:"required"`
	ProductID       string `json:"productId"`
	ProductName     string `json:"productName"`
	ProductColor    string `json:"productColor"`
	ProductCategory string `json:"productCategory"`
	Size            string `json:"size"`
	RecommendedSize string `json:"recommendedSize"`
	Fit             string `json:"fit"`
}

type DeleteTryOnRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}

// ---- Handlers ----

// GetTryOnsHandler returns all try-on records for the authenticated user,
// sorted newest-first.
// GET /api/tryon
func GetTryOnsHandler(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDRaw.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := config.GetDB().Collection("tryons")
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := col.Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to fetch try-ons", err)
		return
	}
	defer cursor.Close(ctx)

	var records []models.TryOnRecord
	if err := cursor.All(ctx, &records); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to decode try-ons", err)
		return
	}
	if records == nil {
		records = []models.TryOnRecord{}
	}

	utils.SuccessResponse(c, http.StatusOK, "Try-ons fetched", records)
}

// TryOnHandler generates a try-on image, uploads it to Cloudinary, persists
// the record to MongoDB, and returns the saved record.
// POST /api/tryon
func TryOnHandler(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDRaw.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var req TryOnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Generate try-on image via Gemini
	genCtx, genCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer genCancel()

	pngBytes, err := utils.EditImageForTryOn(
		genCtx,
		req.AvatarImage, req.ProductImage, req.Prompt,
		req.ProductName, req.ProductColor, req.ProductCategory,
	)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Try-on generation failed", err)
		return
	}

	// Upload result to Cloudinary
	uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer uploadCancel()

	var resultUrl string
	uploadResult, err := utils.UploadReaderToCloudinary(uploadCtx, bytes.NewReader(pngBytes), "kloset/tryon")
	if err != nil {
		// Cloudinary unavailable — fall back to base64 data URI
		b64 := base64.StdEncoding.EncodeToString(pngBytes)
		resultUrl = "data:image/png;base64," + b64
	} else {
		resultUrl = uploadResult.URL
	}

	// Persist to MongoDB
	record := models.TryOnRecord{
		UserID:          userID,
		ProductID:       req.ProductID,
		ProductName:     req.ProductName,
		ProductImage:    req.ProductImage,
		ProductCategory: req.ProductCategory,
		AvatarUrl:       req.AvatarImage,
		ResultUrl:       resultUrl,
		Size:            req.Size,
		Color:           req.ProductColor,
		Fit:             req.Fit,
		RecommendedSize: req.RecommendedSize,
		CreatedAt:       time.Now(),
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	col := config.GetDB().Collection("tryons")
	insertResult, err := col.InsertOne(dbCtx, record)
	if err != nil {
		// DB save failed — still return the result URL so the user isn't blocked
		utils.SuccessResponse(c, http.StatusOK, "Try-on generated (not persisted)", gin.H{
			"resultUrl": resultUrl,
			"id":        "",
		})
		return
	}

	record.ID = insertResult.InsertedID.(primitive.ObjectID)
	utils.SuccessResponse(c, http.StatusOK, "Try-on generated", record)
}

// DeleteTryOnsHandler deletes specific try-on records by their MongoDB IDs,
// cleaning up Cloudinary and MongoDB.
// DELETE /api/tryon
func DeleteTryOnsHandler(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDRaw.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var req DeleteTryOnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Convert string IDs to ObjectIDs
	var objectIDs []primitive.ObjectID
	for _, id := range req.IDs {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, oid)
	}
	if len(objectIDs) == 0 {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "No valid IDs provided", nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	col := config.GetDB().Collection("tryons")

	// Fetch records to get Cloudinary URLs before deleting
	cursor, err := col.Find(ctx, bson.M{
		"_id":    bson.M{"$in": objectIDs},
		"userId": userID,
	})
	if err == nil {
		var records []models.TryOnRecord
		if cursor.All(ctx, &records) == nil {
			for _, r := range records {
				if publicID, ok := cloudinaryPublicID(r.ResultUrl); ok {
					utils.DeleteFromCloudinary(ctx, publicID) //nolint:errcheck
				}
			}
		}
		cursor.Close(ctx)
	}

	// Delete from MongoDB
	col.DeleteMany(ctx, bson.M{ //nolint:errcheck
		"_id":    bson.M{"$in": objectIDs},
		"userId": userID,
	})

	utils.SuccessResponse(c, http.StatusOK, "Try-ons deleted", nil)
}

// DeleteAllTryOnsHandler deletes every try-on for the authenticated user.
// DELETE /api/tryon/all
func DeleteAllTryOnsHandler(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDRaw.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	col := config.GetDB().Collection("tryons")

	// Fetch all to clean up Cloudinary
	cursor, err := col.Find(ctx, bson.M{"userId": userID})
	if err == nil {
		var records []models.TryOnRecord
		if cursor.All(ctx, &records) == nil {
			for _, r := range records {
				if publicID, ok := cloudinaryPublicID(r.ResultUrl); ok {
					utils.DeleteFromCloudinary(ctx, publicID) //nolint:errcheck
				}
			}
		}
		cursor.Close(ctx)
	}

	col.DeleteMany(ctx, bson.M{"userId": userID}) //nolint:errcheck

	utils.SuccessResponse(c, http.StatusOK, "All try-ons deleted", nil)
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
