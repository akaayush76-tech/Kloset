package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/config"
	"github.com/kloset/backend/internal/models"
	"github.com/kloset/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetProfileHandler retrieves authenticated user profile
func GetProfileHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Convert string ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Find user
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

	utils.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", user)
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Avatar string `json:"avatar"`
	Gender string `json:"gender"`
}

// UpdateProfileHandler updates user profile
func UpdateProfileHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate inputs
	if req.Name != "" && !utils.NameValidator(req.Name) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Name must be 2-50 characters", nil)
		return
	}

	if req.Phone != "" && !utils.PhoneValidator(req.Phone) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid phone number", nil)
		return
	}

	// Convert string ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Prepare update
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if req.Name != "" {
		update["$set"].(bson.M)["name"] = req.Name
	}
	if req.Phone != "" {
		update["$set"].(bson.M)["phone"] = req.Phone
	}
	if req.Avatar != "" {
		update["$set"].(bson.M)["avatar"] = req.Avatar
	}
	if req.Gender != "" {
		update["$set"].(bson.M)["gender"] = req.Gender
	}

	// Update user
	db := config.GetDB()
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := collection.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "User not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error updating profile", result.Err())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile updated successfully", nil)
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// ChangePasswordHandler changes user password
func ChangePasswordHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate new password
	if !utils.PasswordValidator(req.NewPassword) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "New password must be at least 6 characters", nil)
		return
	}

	// Convert string ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Find user
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

	// Verify old password
	if !utils.ComparePassword(req.OldPassword, user.Password) {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Invalid current password", nil)
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	// Update password
	update := bson.M{
		"$set": bson.M{
			"password":  hashedPassword,
			"updatedAt": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error updating password", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)
}

// LogoutHandler handles user logout
func LogoutHandler(c *gin.Context) {
	// In JWT-based authentication, logout is typically handled on client side
	// by removing the token. Server-side logout can be implemented with token blacklist.
	// For now, returning success response.
	utils.SuccessResponse(c, http.StatusOK, "Logged out successfully", nil)
}
