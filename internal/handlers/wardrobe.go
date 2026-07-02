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

// GetWardrobeHandler retrieves user's wardrobe items with filtering and pagination
func GetWardrobeHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	category := c.Query("category")
	subtype := c.Query("subtype")

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
	collection := db.Collection("wardrobeItems")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter
	filter := bson.M{"userId": userObjID, "isActive": true}
	if category != "" {
		filter["category"] = category
	}
	if subtype != "" {
		filter["subtype"] = subtype
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error counting items", err)
		return
	}

	// Find items
	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching wardrobe", err)
		return
	}
	defer cursor.Close(ctx)

	var items []models.WardrobeItem
	if err = cursor.All(ctx, &items); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding items", err)
		return
	}

	if items == nil {
		items = []models.WardrobeItem{}
	}

	pagination := utils.PaginationMeta(page, limit, total)
	utils.SuccessResponseWithPagination(c, http.StatusOK, "Wardrobe items retrieved successfully", items, pagination)
}

// GetWardrobeItemHandler retrieves a single wardrobe item
func GetWardrobeItemHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	itemID := c.Param("id")
	itemObjID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("wardrobeItems")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var item models.WardrobeItem
	err = collection.FindOne(ctx, bson.M{"_id": itemObjID, "userId": userObjID}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Wardrobe item not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching item", err)
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Wardrobe item retrieved successfully", item)
}

// CreateWardrobeItemRequest represents wardrobe item creation request
type CreateWardrobeItemRequest struct {
	Name         string                `json:"name" binding:"required"`
	Category     string                `json:"category" binding:"required"`
	Subtype      string                `json:"subtype" binding:"required"`
	Color        string                `json:"color" binding:"required"`
	Image        string                `json:"image" binding:"required"`
	Brand        string                `json:"brand"`
	Size         string                `json:"size"`
	PurchaseDate string                `json:"purchaseDate"`
	Price        float64               `json:"price"`
	Tags         []string              `json:"tags"`
	Condition    string                `json:"condition"`
	Notes        string                `json:"notes"`
	Identifiers  models.ItemIdentifiers `json:"identifiers"`
}

// CreateWardrobeItemHandler creates a new wardrobe item
func CreateWardrobeItemHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req CreateWardrobeItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate category
	if !utils.CategoryValidator(req.Category) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid category (upper, lower, shoes)", nil)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse purchase date
	var purchaseDate time.Time
	if req.PurchaseDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.PurchaseDate)
		if err != nil {
			purchaseDate = time.Now()
		} else {
			purchaseDate = parsedDate
		}
	} else {
		purchaseDate = time.Now()
	}

	item := models.WardrobeItem{
		ID:           primitive.NewObjectID(),
		UserID:       userObjID,
		Name:         req.Name,
		Category:     req.Category,
		Subtype:      req.Subtype,
		Color:        req.Color,
		Image:        req.Image,
		Brand:        req.Brand,
		Size:         req.Size,
		PurchaseDate: purchaseDate,
		Price:        req.Price,
		Tags:         req.Tags,
		Condition:    req.Condition,
		Notes:        req.Notes,
		Identifiers:  req.Identifiers,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	db := config.GetDB()
	collection := db.Collection("wardrobeItems")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, item)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error creating wardrobe item", err)
		return
	}

	item.ID = result.InsertedID.(primitive.ObjectID)
	utils.SuccessResponse(c, http.StatusCreated, "Wardrobe item created successfully", item)
}

// UpdateWardrobeItemRequest represents wardrobe item update request
type UpdateWardrobeItemRequest struct {
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Subtype     string                 `json:"subtype"`
	Color       string                 `json:"color"`
	Image       string                 `json:"image"`
	Brand       string                 `json:"brand"`
	Size        string                 `json:"size"`
	Price       float64                `json:"price"`
	Condition   string                 `json:"condition"`
	Notes       string                 `json:"notes"`
	Identifiers *models.ItemIdentifiers `json:"identifiers"`
}

// UpdateWardrobeItemHandler updates a wardrobe item
func UpdateWardrobeItemHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	itemID := c.Param("id")
	itemObjID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	var req UpdateWardrobeItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("wardrobeItems")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build update
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if req.Name != "" {
		update["$set"].(bson.M)["name"] = req.Name
	}
	if req.Category != "" {
		update["$set"].(bson.M)["category"] = req.Category
	}
	if req.Subtype != "" {
		update["$set"].(bson.M)["subtype"] = req.Subtype
	}
	if req.Color != "" {
		update["$set"].(bson.M)["color"] = req.Color
	}
	if req.Image != "" {
		update["$set"].(bson.M)["image"] = req.Image
	}
	if req.Brand != "" {
		update["$set"].(bson.M)["brand"] = req.Brand
	}
	if req.Size != "" {
		update["$set"].(bson.M)["size"] = req.Size
	}
	if req.Price > 0 {
		update["$set"].(bson.M)["price"] = req.Price
	}
	if req.Condition != "" {
		update["$set"].(bson.M)["condition"] = req.Condition
	}
	if req.Notes != "" {
		update["$set"].(bson.M)["notes"] = req.Notes
	}
	if req.Identifiers != nil {
		update["$set"].(bson.M)["identifiers"] = req.Identifiers
	}

	result := collection.FindOneAndUpdate(ctx, bson.M{
		"_id":    itemObjID,
		"userId": userObjID,
	}, update)

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Wardrobe item not found or not owned by user", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error updating item", result.Err())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Wardrobe item updated successfully", nil)
}

// DeleteWardrobeItemHandler deletes a wardrobe item
func DeleteWardrobeItemHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	itemID := c.Param("id")
	itemObjID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("wardrobeItems")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{
		"_id":    itemObjID,
		"userId": userObjID,
	})

	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error deleting item", err)
		return
	}

	if result.DeletedCount == 0 {
		utils.HTTPErrorHandler(c, http.StatusNotFound, "Wardrobe item not found or not owned by user", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Wardrobe item deleted successfully", nil)
}

// GetWardrobeByCategoryHandler retrieves wardrobe items filtered by category
func GetWardrobeByCategoryHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	category := c.Param("category")

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("wardrobeItems")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{
		"userId":   userObjID,
		"category": category,
		"isActive": true,
	})
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching items", err)
		return
	}
	defer cursor.Close(ctx)

	var items []models.WardrobeItem
	if err = cursor.All(ctx, &items); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding items", err)
		return
	}

	if items == nil {
		items = []models.WardrobeItem{}
	}

	utils.SuccessResponse(c, http.StatusOK, "Wardrobe items retrieved successfully", items)
}

// GetWardrobeStatsHandler retrieves wardrobe statistics
func GetWardrobeStatsHandler(c *gin.Context) {
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
	collection := db.Collection("wardrobeItems")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"userId": userObjID, "isActive": true})
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching items", err)
		return
	}
	defer cursor.Close(ctx)

	var items []models.WardrobeItem
	if err = cursor.All(ctx, &items); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding items", err)
		return
	}

	// Calculate stats
	byCategory := make(map[string]int)
	byCondition := make(map[string]int)
	totalAge := 0.0

	for _, item := range items {
		byCategory[item.Category]++
		byCondition[item.Condition]++

		// Calculate age in days
		age := time.Since(item.PurchaseDate).Hours() / 24
		totalAge += age
	}

	avgAge := 0.0
	if len(items) > 0 {
		avgAge = totalAge / float64(len(items))
	}

	stats := gin.H{
		"totalItems":  len(items),
		"byCategory":  byCategory,
		"byCondition": byCondition,
		"averageAge":  avgAge,
	}

	utils.SuccessResponse(c, http.StatusOK, "Wardrobe statistics retrieved successfully", stats)
}
