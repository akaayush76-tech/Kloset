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

// GetProductsHandler retrieves products with filtering, sorting, and pagination
func GetProductsHandler(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	category := c.Query("category")
	brand := c.Query("brand")
	search := c.Query("search")
	sortBy := c.DefaultQuery("sortBy", "createdAt")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	// Validate and cap limit
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}

	// Build filter
	filter := bson.M{"isActive": true}

	if category != "" {
		filter["category"] = category
	}
	if brand != "" {
		filter["brand"] = brand
	}
	if search != "" {
		filter["$text"] = bson.M{"$search": search}
	}

	// Build sort
	sortValue := 1
	if sortOrder == "desc" {
		sortValue = -1
	}
	sort := bson.M{sortBy: sortValue}

	// Get total count
	db := config.GetDB()
	collection := db.Collection("products")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error counting products", err)
		return
	}

	// Find products
	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(sort)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching products", err)
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding products", err)
		return
	}

	if products == nil {
		products = []models.Product{}
	}

	pagination := utils.PaginationMeta(page, limit, total)
	utils.SuccessResponseWithPagination(c, http.StatusOK, "Products retrieved successfully", products, pagination)
}

// GetProductHandler retrieves a single product by ID
func GetProductHandler(c *gin.Context) {
	productID := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid product ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("products")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var product models.Product
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Product not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching product", err)
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product retrieved successfully", product)
}

// GetCategoriesHandler retrieves all product categories
func GetCategoriesHandler(c *gin.Context) {
	db := config.GetDB()
	collection := db.Collection("products")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get distinct categories
	categories, err := collection.Distinct(ctx, "category", bson.M{})
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching categories", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Categories retrieved successfully", gin.H{
		"categories": categories,
	})
}

// GetFeaturedHandler retrieves featured products
func GetFeaturedHandler(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	if limit > 100 {
		limit = 100
	}

	db := config.GetDB()
	collection := db.Collection("products")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetLimit(limit).
		SetSort(bson.M{"rating": -1})

	cursor, err := collection.Find(ctx, bson.M{"isActive": true}, opts)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching featured products", err)
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding products", err)
		return
	}

	if products == nil {
		products = []models.Product{}
	}

	utils.SuccessResponse(c, http.StatusOK, "Featured products retrieved successfully", products)
}

// GetRelatedHandler retrieves related products (max 4)
func GetRelatedHandler(c *gin.Context) {
	productID := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid product ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("products")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the product first
	var product models.Product
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Product not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching product", err)
		}
		return
	}

	// Find related products with same category (max 4)
	opts := options.Find().
		SetLimit(4).
		SetSort(bson.M{"rating": -1})

	cursor, err := collection.Find(ctx, bson.M{
		"category": product.Category,
		"_id":      bson.M{"$ne": objID},
		"isActive": true,
	}, opts)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching related products", err)
		return
	}
	defer cursor.Close(ctx)

	var relatedProducts []models.Product
	if err = cursor.All(ctx, &relatedProducts); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding products", err)
		return
	}

	if relatedProducts == nil {
		relatedProducts = []models.Product{}
	}

	utils.SuccessResponse(c, http.StatusOK, "Related products retrieved successfully", relatedProducts)
}

// CreateProductRequest represents product creation request
type CreateProductRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description" binding:"required"`
	Price         float64  `json:"price" binding:"required"`
	OriginalPrice float64  `json:"originalPrice"`
	Category      string   `json:"category" binding:"required"`
	Subcategory   string   `json:"subcategory" binding:"required"`
	Brand         string   `json:"brand" binding:"required"`
	Images        []string `json:"images"`
	Colors        []string `json:"colors"`
	Sizes         []string `json:"sizes"`
	Features      []string `json:"features"`
	DeliveryDays  int      `json:"deliveryDays"`
}

// CreateProductHandler creates a new product (admin only)
func CreateProductHandler(c *gin.Context) {
	// Check admin privilege (would check from database in production)
	_, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Admin access required", nil)
		return
	}

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if !utils.DescriptionValidator(req.Description) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Description must be max 1000 characters", nil)
		return
	}

	product := models.Product{
		ID:            primitive.NewObjectID().Hex(),
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Discount:      0, // Calculate if needed
		Category:      req.Category,
		Subcategory:   req.Subcategory,
		Brand:         req.Brand,
		Images:        req.Images,
		Colors:        req.Colors,
		Sizes:         req.Sizes,
		Features:      req.Features,
		Rating:        0,
		ReviewCount:   0,
		DeliveryDays:  req.DeliveryDays,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	db := config.GetDB()
	collection := db.Collection("products")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, product)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error creating product", err)
		return
	}

	product.ID = result.InsertedID.(string)
	utils.SuccessResponse(c, http.StatusCreated, "Product created successfully", product)
}
