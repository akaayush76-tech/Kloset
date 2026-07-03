package handlers

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/config"
	"github.com/kloset/backend/internal/models"
	"github.com/kloset/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CheckAvatarHandler returns whether the authenticated user has an avatar set.
// GET /api/avatar/check
func CheckAvatarHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	objID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "User not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching user", err)
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Avatar status retrieved", gin.H{
		"hasAvatar": user.Avatar != "",
		"avatarUrl": user.Avatar,
	})
}

// GenerateAvatarRequest is the request body for POST /api/avatar/generate.
type GenerateAvatarRequest struct {
	Photo             string  `json:"photo" binding:"required"`
	Height            float64 `json:"height"`
	Weight            float64 `json:"weight"`
	Gender            string  `json:"gender"`
	BodyType          string  `json:"bodyType"`
	Chest             float64 `json:"chest"`
	Waist             float64 `json:"waist"`
	Hip               float64 `json:"hip"`
	SleeveLength      float64 `json:"sleeveLength"`
	Inseam            float64 `json:"inseam"`
	CustomDescription string  `json:"customDescription"`
}

// GenerateAvatarHandler generates a full-body avatar from a reference photo using Gemini,
// uploads the result to Cloudinary, and persists the CDN URL to the user's profile.
// POST /api/avatar/generate
func GenerateAvatarHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req GenerateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if !strings.HasPrefix(req.Photo, "data:") {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "photo must be a base64 data URI", nil)
		return
	}

	// Call Gemini — allow up to 90 s for image generation
	genCtx, genCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer genCancel()

	profile := &utils.AvatarGenerationProfile{
		Height:            req.Height,
		Weight:            req.Weight,
		Gender:            req.Gender,
		BodyType:          req.BodyType,
		Chest:             req.Chest,
		Waist:             req.Waist,
		Hip:               req.Hip,
		SleeveLength:      req.SleeveLength,
		Inseam:            req.Inseam,
		CustomDescription: req.CustomDescription,
	}

	pngBytes, err := utils.GenerateAvatarFromPhoto(genCtx, req.Photo, profile)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Avatar generation failed", err)
		return
	}

	// Upload PNG to Cloudinary
	uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer uploadCancel()

	uploadResult, err := utils.UploadReaderToCloudinary(uploadCtx, bytes.NewReader(pngBytes), "kloset/avatars")
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to upload avatar image", err)
		return
	}

	// Persist CDN URL to user document
	objID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("users")

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	res, err := collection.UpdateOne(
		dbCtx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"avatar": uploadResult.URL, "updatedAt": time.Now()}},
	)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to save avatar", err)
		return
	}
	if res.MatchedCount == 0 {
		utils.HTTPErrorHandler(c, http.StatusNotFound, "User not found", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Avatar generated", gin.H{
		"avatarUrl": uploadResult.URL,
	})
}

// SaveAvatarRequest represents the save avatar request body.
type SaveAvatarRequest struct {
	AvatarURL string `json:"avatarUrl" binding:"required"`
}

// SaveAvatarHandler persists an avatar URL to the authenticated user's profile.
// POST /api/avatar/save
func SaveAvatarHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req SaveAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Upload to Cloudinary if a base64 data URI was provided instead of a URL.
	avatarURL := req.AvatarURL
	if strings.HasPrefix(avatarURL, "data:") {
		uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer uploadCancel()

		result, err := utils.UploadDataURIToCloudinary(uploadCtx, avatarURL, "kloset/avatars")
		if err != nil {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error uploading avatar image", err)
			return
		}
		avatarURL = result.URL
	}

	objID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"avatar":    avatarURL,
			"updatedAt": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error saving avatar", err)
		return
	}

	if result.MatchedCount == 0 {
		utils.HTTPErrorHandler(c, http.StatusNotFound, "User not found", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Avatar saved successfully", nil)
}
