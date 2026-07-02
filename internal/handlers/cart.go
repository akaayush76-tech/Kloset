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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetCartHandler retrieves user's cart
func GetCartHandler(c *gin.Context) {
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
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching cart", err)
		}
		return
	}

	if user.Cart == nil {
		user.Cart = []models.CartItem{}
	}

	utils.SuccessResponse(c, http.StatusOK, "Cart retrieved successfully", gin.H{
		"items": user.Cart,
		"total": calculateCartTotal(user.Cart),
	})
}

// AddToCartRequest represents add to cart request
type AddToCartRequest struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
	Size      string `json:"size" binding:"required"`
	Color     string `json:"color"`
}

// AddToCartHandler adds item to user's cart
func AddToCartHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate product ID
	productID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid product ID", err)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Get product details
	db := config.GetDB()
	productsCollection := db.Collection("products")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var product models.Product
	err = productsCollection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Product not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching product", err)
		}
		return
	}

	// Create cart item
	cartItem := models.CartItem{
		ID:        primitive.NewObjectID(),
		ProductID: productID,
		Quantity:  req.Quantity,
		Size:      req.Size,
		Color:     req.Color,
		Price:     product.Price,
		AddedAt:   time.Now(),
	}

	// Add to cart
	usersCollection := db.Collection("users")
	update := bson.M{
		"$push": bson.M{"cart": cartItem},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	_, err = usersCollection.UpdateOne(ctx, bson.M{"_id": userObjID}, update)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error adding to cart", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Item added to cart", cartItem)
}

// UpdateCartItemRequest represents update cart item request
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required"`
}

// UpdateCartItemHandler updates quantity of cart item
func UpdateCartItemHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	itemID := c.Param("itemId")
	itemObjID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	var req UpdateCartItemRequest
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
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Update cart item quantity
	update := bson.M{
		"$set": bson.M{
			"cart.$[elem].quantity": req.Quantity,
			"updatedAt":             time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"elem._id": itemObjID}},
	})

	// Note: This is simplified - in production use arrayFilters properly
	result := collection.FindOneAndUpdate(ctx, bson.M{"_id": userObjID}, update, opts)
	if result.Err() != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error updating cart", result.Err())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cart item updated", nil)
}

// RemoveFromCartHandler removes item from cart
func RemoveFromCartHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	itemID := c.Param("itemId")
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
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Remove from cart
	update := bson.M{
		"$pull": bson.M{"cart": bson.M{"_id": itemObjID}},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": userObjID}, update)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error removing from cart", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Item removed from cart", nil)
}

// ClearCartHandler clears entire cart
func ClearCartHandler(c *gin.Context) {
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
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Clear cart
	update := bson.M{
		"$set": bson.M{
			"cart":      []models.CartItem{},
			"updatedAt": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": userObjID}, update)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error clearing cart", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cart cleared successfully", nil)
}

// GetCartCountHandler gets total items count in cart
func GetCartCountHandler(c *gin.Context) {
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
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching cart", err)
		}
		return
	}

	count := len(user.Cart)
	utils.SuccessResponse(c, http.StatusOK, "Cart count retrieved", gin.H{
		"count": count,
	})
}

// Helper function to calculate cart total
func calculateCartTotal(items []models.CartItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}
