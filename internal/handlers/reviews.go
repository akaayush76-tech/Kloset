package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/config"
	"github.com/kloset/backend/internal/models"
	"github.com/kloset/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetProductReviewsHandler retrieves reviews for a product with pagination
func GetProductReviewsHandler(c *gin.Context) {
	productID := c.Param("productId")
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	ratingFilter, _ := strconv.Atoi(c.DefaultQuery("rating", "0"))

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}

	productObjID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid product ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("reviews")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter
	filter := bson.M{"productId": productObjID}
	if ratingFilter > 0 && ratingFilter <= 5 {
		filter["rating"] = ratingFilter
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error counting reviews", err)
		return
	}

	// Find reviews
	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching reviews", err)
		return
	}
	defer cursor.Close(ctx)

	var reviews []models.Review
	if err = cursor.All(ctx, &reviews); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding reviews", err)
		return
	}

	if reviews == nil {
		reviews = []models.Review{}
	}

	// Calculate rating distribution
	distribution := make(map[string]int)
	avgRating := 0.0
	for i := 1; i <= 5; i++ {
		distribution[strconv.Itoa(i)] = 0
	}

	for _, review := range reviews {
		distribution[strconv.Itoa(review.Rating)]++
		avgRating += float64(review.Rating)
	}

	if len(reviews) > 0 {
		avgRating /= float64(len(reviews))
	}

	response := gin.H{
		"reviews":       reviews,
		"averageRating": avgRating,
		"distribution":  distribution,
	}

	pagination := utils.PaginationMeta(page, limit, total)
	utils.SuccessResponseWithPagination(c, http.StatusOK, "Reviews retrieved successfully", response, pagination)
}

// CreateReviewRequest represents review creation request
type CreateReviewRequest struct {
	ProductID string `json:"productId" binding:"required"`
	Rating    int    `json:"rating" binding:"required"`
	Title     string `json:"title"`
	Comment   string `json:"comment" binding:"required"`
}

// CreateReviewHandler creates a new review
func CreateReviewHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate rating
	if !utils.RatingValidator(req.Rating) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Rating must be between 1 and 5", nil)
		return
	}

	// Validate comment
	if !utils.CommentValidator(req.Comment) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Comment must be 10-500 characters", nil)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	productObjID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid product ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("reviews")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if review already exists
	count, err := collection.CountDocuments(ctx, bson.M{
		"userId":    userObjID,
		"productId": productObjID,
	})
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error checking existing review", err)
		return
	}

	if count > 0 {
		utils.HTTPErrorHandler(c, http.StatusConflict, "You have already reviewed this product", nil)
		return
	}

	// Create review
	review := models.Review{
		ID:               primitive.NewObjectID(),
		UserID:           userObjID,
		ProductID:        productObjID,
		Rating:           req.Rating,
		Title:            req.Title,
		Comment:          req.Comment,
		VerifiedPurchase: false, // TODO: Check if user purchased this product
		Helpful:          0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	result, err := collection.InsertOne(ctx, review)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			utils.HTTPErrorHandler(c, http.StatusConflict, "You have already reviewed this product", err)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error creating review", err)
		}
		return
	}

	review.ID = result.InsertedID.(primitive.ObjectID)
	utils.SuccessResponse(c, http.StatusCreated, "Review created successfully", review)
}

// UpdateReviewRequest represents review update request
type UpdateReviewRequest struct {
	Rating  int    `json:"rating"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
}

// UpdateReviewHandler updates user's own review
func UpdateReviewHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	reviewID := c.Param("id")
	reviewObjID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid review ID", err)
		return
	}

	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Validate inputs if provided
	if req.Rating > 0 && !utils.RatingValidator(req.Rating) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Rating must be between 1 and 5", nil)
		return
	}

	if req.Comment != "" && !utils.CommentValidator(req.Comment) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Comment must be 10-500 characters", nil)
		return
	}

	db := config.GetDB()
	collection := db.Collection("reviews")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build update
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if req.Rating > 0 {
		update["$set"].(bson.M)["rating"] = req.Rating
	}
	if req.Title != "" {
		update["$set"].(bson.M)["title"] = req.Title
	}
	if req.Comment != "" {
		update["$set"].(bson.M)["comment"] = req.Comment
	}

	result := collection.FindOneAndUpdate(ctx, bson.M{
		"_id":    reviewObjID,
		"userId": userObjID,
	}, update)

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Review not found or not owned by user", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error updating review", result.Err())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Review updated successfully", nil)
}

// DeleteReviewHandler deletes user's own review
func DeleteReviewHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	reviewID := c.Param("id")
	reviewObjID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid review ID", err)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("reviews")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{
		"_id":    reviewObjID,
		"userId": userObjID,
	})

	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error deleting review", err)
		return
	}

	if result.DeletedCount == 0 {
		utils.HTTPErrorHandler(c, http.StatusNotFound, "Review not found or not owned by user", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Review deleted successfully", nil)
}

// GetMyReviewsHandler retrieves authenticated user's reviews
func GetMyReviewsHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("reviews")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get total count
	total, err := collection.CountDocuments(ctx, bson.M{"userId": userObjID})
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error counting reviews", err)
		return
	}

	// Find reviews
	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := collection.Find(ctx, bson.M{"userId": userObjID}, opts)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching reviews", err)
		return
	}
	defer cursor.Close(ctx)

	var reviews []models.Review
	if err = cursor.All(ctx, &reviews); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding reviews", err)
		return
	}

	if reviews == nil {
		reviews = []models.Review{}
	}

	pagination := utils.PaginationMeta(page, limit, total)
	utils.SuccessResponseWithPagination(c, http.StatusOK, "Your reviews retrieved successfully", reviews, pagination)
}

// MarkHelpfulHandler increments helpful count for a review
func MarkHelpfulHandler(c *gin.Context) {
	reviewID := c.Param("id")
	reviewObjID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid review ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("reviews")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$inc": bson.M{"helpful": 1},
	}

	result := collection.FindOneAndUpdate(ctx, bson.M{"_id": reviewObjID}, update)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Review not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error updating review", result.Err())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Review marked as helpful", nil)
}

// GetReviewStatsHandler retrieves review statistics for user
func GetReviewStatsHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("reviews")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"userId": userObjID})
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching reviews", err)
		return
	}
	defer cursor.Close(ctx)

	var reviews []models.Review
	if err = cursor.All(ctx, &reviews); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding reviews", err)
		return
	}

	stats := gin.H{
		"totalReviews":  len(reviews),
		"averageRating": 0.0,
		"distribution":  map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0},
		"totalHelpful":  0,
	}

	avgRating := 0.0
	totalHelpful := 0

	for _, review := range reviews {
		avgRating += float64(review.Rating)
		totalHelpful += review.Helpful
		ratingStr := strconv.Itoa(review.Rating)
		dist := stats["distribution"].(map[string]int)
		dist[ratingStr]++
	}

	if len(reviews) > 0 {
		avgRating /= float64(len(reviews))
	}

	stats["averageRating"] = avgRating
	stats["totalHelpful"] = totalHelpful

	utils.SuccessResponse(c, http.StatusOK, "Review statistics retrieved successfully", stats)
}
