package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/config"
	"github.com/kloset/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UpdateProductRequest represents product update request
type UpdateProductRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	OriginalPrice float64  `json:"originalPrice"`
	Category      string   `json:"category"`
	Subcategory   string   `json:"subcategory"`
	Brand         string   `json:"brand"`
	Images        []string `json:"images"`
	Colors        []string `json:"colors"`
	Sizes         []string `json:"sizes"`
	Features      []string `json:"features"`
	IsActive      *bool    `json:"isActive"`
}

// UpdateProductHandler updates an existing product (admin only)
func UpdateProductHandler(c *gin.Context) {
	// Check admin privilege
	_, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Admin access required", nil)
		return
	}

	productID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid product ID", err)
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate description if provided
	if req.Description != "" && !utils.DescriptionValidator(req.Description) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Description must be max 1000 characters", nil)
		return
	}

	// Build update document
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if req.Name != "" {
		update["$set"].(bson.M)["name"] = req.Name
	}
	if req.Description != "" {
		update["$set"].(bson.M)["description"] = req.Description
	}
	if req.Price > 0 {
		update["$set"].(bson.M)["price"] = req.Price
	}
	if req.OriginalPrice > 0 {
		update["$set"].(bson.M)["originalPrice"] = req.OriginalPrice
	}
	if req.Category != "" {
		update["$set"].(bson.M)["category"] = req.Category
	}
	if req.Subcategory != "" {
		update["$set"].(bson.M)["subcategory"] = req.Subcategory
	}
	if req.Brand != "" {
		update["$set"].(bson.M)["brand"] = req.Brand
	}
	if req.Images != nil {
		update["$set"].(bson.M)["images"] = req.Images
	}
	if req.Colors != nil {
		update["$set"].(bson.M)["colors"] = req.Colors
	}
	if req.Sizes != nil {
		update["$set"].(bson.M)["sizes"] = req.Sizes
	}
	if req.Features != nil {
		update["$set"].(bson.M)["features"] = req.Features
	}
	if req.IsActive != nil {
		update["$set"].(bson.M)["isActive"] = *req.IsActive
	}

	db := config.GetDB()
	collection := db.Collection("products")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := collection.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Product not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error updating product", result.Err())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product updated successfully", nil)
}

// DeleteProductHandler deletes a product (admin only)
func DeleteProductHandler(c *gin.Context) {
	// Check admin privilege
	_, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Admin access required", nil)
		return
	}

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

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error deleting product", err)
		return
	}

	if result.DeletedCount == 0 {
		utils.HTTPErrorHandler(c, http.StatusNotFound, "Product not found", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product deleted successfully", nil)
}
