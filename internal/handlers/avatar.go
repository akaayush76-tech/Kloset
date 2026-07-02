package handlers

import (
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
