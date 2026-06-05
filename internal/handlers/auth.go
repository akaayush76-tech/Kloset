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

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	Gender      string `json:"gender"`
	DateOfBirth string `json:"dateOfBirth"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Token string `json:"token"`
}

// RegisterHandler handles user registration
func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate inputs
	if !utils.NameValidator(req.Name) {
		utils.SuccessResponse(c, http.StatusBadRequest, "Name must be 2-50 characters", nil)
		return
	}

	if !utils.EmailValidator(req.Email) {
		utils.SuccessResponse(c, http.StatusBadRequest, "Invalid email format", nil)
		return
	}

	if !utils.PasswordValidator(req.Password) {
		utils.SuccessResponse(c, http.StatusBadRequest, "Password must be at least 6 characters", nil)
		return
	}

	if !utils.PhoneValidator(req.Phone) {
		utils.SuccessResponse(c, http.StatusBadRequest, "Invalid phone number. Must be 10-digit Indian number starting with 6-9", nil)
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	// Parse date of birth
	var dob time.Time
	if req.DateOfBirth != "" {
		parsedDob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			dob = time.Time{}
		} else {
			dob = parsedDob
		}
	}

	// Create user document
	user := models.User{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Email:       req.Email,
		Password:    hashedPassword,
		Phone:       req.Phone,
		Gender:      req.Gender,
		DateOfBirth: dob,
		Preferences: models.UserPreferences{
			Newsletter:      false,
			Notifications:   true,
			PreferredColors: []string{},
			PreferredSizes:  []string{},
		},
		Wishlist:      []primitive.ObjectID{},
		Cart:          []models.CartItem{},
		Orders:        []primitive.ObjectID{},
		Reviews:       []primitive.ObjectID{},
		WardrobeItems: []primitive.ObjectID{},
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Insert into database
	db := config.GetDB()
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			utils.HTTPErrorHandler(c, http.StatusConflict, "Email already registered", err)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error creating user", err)
		}
		return
	}

	// Generate token
	token, err := utils.GenerateToken(result.InsertedID.(primitive.ObjectID).Hex(), user.Email)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error generating token", err)
		return
	}

	// Response
	response := AuthResponse{
		ID:    result.InsertedID.(primitive.ObjectID).Hex(),
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Token: token,
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", response)
}

// LoginHandler handles user login
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate inputs
	if !utils.EmailValidator(req.Email) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid email format", nil)
		return
	}

	if !utils.PasswordValidator(req.Password) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid credentials", nil)
		return
	}

	// Find user by email
	db := config.GetDB()
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Invalid credentials", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error finding user", err)
		}
		return
	}

	// Verify password
	if !utils.ComparePassword(req.Password, user.Password) {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error generating token", err)
		return
	}

	// Response
	response := AuthResponse{
		ID:    user.ID.Hex(),
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Token: token,
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}
