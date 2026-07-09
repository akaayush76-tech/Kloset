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

func wishlistContext(c *gin.Context) (context.Context, context.CancelFunc, primitive.ObjectID, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return nil, nil, primitive.NilObjectID, false
	}
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user id", err)
		return nil, nil, primitive.NilObjectID, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return ctx, cancel, userObjID, true
}

// GetWishlistHandler handles GET /api/wishlist — returns the wishlisted products.
func GetWishlistHandler(c *gin.Context) {
	ctx, cancel, userObjID, ok := wishlistContext(c)
	if !ok {
		return
	}
	defer cancel()
	db := config.GetDB()

	var user models.User
	if err := db.Collection("users").FindOne(ctx, bson.M{"_id": userObjID}).Decode(&user); err != nil {
		utils.HTTPErrorHandler(c, http.StatusNotFound, "User not found", err)
		return
	}

	products := []models.Product{}
	if len(user.Wishlist) > 0 {
		cursor, err := db.Collection("products").Find(ctx, bson.M{"_id": bson.M{"$in": user.Wishlist}})
		if err != nil {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to load wishlist", err)
			return
		}
		if err := cursor.All(ctx, &products); err != nil {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to load wishlist", err)
			return
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Wishlist retrieved", products)
}

// AddToWishlistHandler handles POST /api/wishlist/:productId.
func AddToWishlistHandler(c *gin.Context) {
	ctx, cancel, userObjID, ok := wishlistContext(c)
	if !ok {
		return
	}
	defer cancel()
	db := config.GetDB()

	productID := c.Param("productId")
	var product models.Product
	if err := db.Collection("products").FindOne(ctx, bson.M{"_id": productID, "isActive": true}).Decode(&product); err != nil {
		if err == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Product not found", nil)
			return
		}
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to load product", err)
		return
	}

	_, err := db.Collection("users").UpdateOne(ctx,
		bson.M{"_id": userObjID},
		bson.M{
			"$addToSet": bson.M{"wishlist": productID},
			"$set":      bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to update wishlist", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Added to wishlist", gin.H{"productId": productID})
}

// RemoveFromWishlistHandler handles DELETE /api/wishlist/:productId.
func RemoveFromWishlistHandler(c *gin.Context) {
	ctx, cancel, userObjID, ok := wishlistContext(c)
	if !ok {
		return
	}
	defer cancel()
	db := config.GetDB()

	productID := c.Param("productId")
	_, err := db.Collection("users").UpdateOne(ctx,
		bson.M{"_id": userObjID},
		bson.M{
			"$pull": bson.M{"wishlist": productID},
			"$set":  bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to update wishlist", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Removed from wishlist", gin.H{"productId": productID})
}
