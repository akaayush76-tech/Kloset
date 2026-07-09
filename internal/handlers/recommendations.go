package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/cache"
	"github.com/kloset/backend/internal/config"
	"github.com/kloset/backend/internal/engine"
	"github.com/kloset/backend/internal/models"
	"github.com/kloset/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type recommendRequest struct {
	TriggerItemID    string `json:"trigger_item_id"`   // optional — empty runs daily-closet mode
	TriggerItemType  string `json:"trigger_item_type"` // required when trigger_item_id is set
	ContextFilter    string `json:"context_filter"`
	Limit            int    `json:"limit"`
	IncludeShopItems bool   `json:"include_shop_items"`
}

// RecommendOutfitsHandler handles POST /api/recommendations/outfits.
func RecommendOutfitsHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req recommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// A trigger item is optional: without one the engine runs in
	// daily-closet mode over the user's Kloset + wishlist.
	if req.TriggerItemID != "" && !utils.TriggerItemTypeValidator(req.TriggerItemType) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "trigger_item_type must be 'catalog' or 'closet'", nil)
		return
	}

	if req.ContextFilter == "" {
		req.ContextFilter = "all"
	}
	if !utils.ContextFilterValidator(req.ContextFilter) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid context_filter value", nil)
		return
	}

	uid := userID.(string)
	cacheKey := fmt.Sprintf("rec:%s:%s:%s:%d:%t", uid, req.TriggerItemID, req.ContextFilter, req.Limit, req.IncludeShopItems)

	if cached, ok := cache.Get(cacheKey); ok {
		utils.SuccessResponse(c, http.StatusOK, "Outfit recommendations retrieved", cached)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := engine.GenerateOutfits(ctx, config.GetDB(), engine.RecommendRequest{
		TriggerItemID:    req.TriggerItemID,
		TriggerItemType:  req.TriggerItemType,
		ContextFilter:    req.ContextFilter,
		Limit:            req.Limit,
		IncludeShopItems: req.IncludeShopItems,
		UserID:           uid,
	})
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to generate recommendations", err)
		return
	}

	cache.Set(cacheKey, result, 5*time.Minute)
	utils.SuccessResponse(c, http.StatusOK, "Outfit recommendations retrieved", result)
}

type wearRequest struct {
	ItemIDs []string `json:"item_ids" binding:"required,min=1"`
}

// LogWornOutfitHandler handles POST /api/recommendations/wear — logs that the
// user wore a combination so future recommendations get the wear-history bonus.
func LogWornOutfitHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user id", err)
		return
	}

	var req wearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	objIDs := make([]primitive.ObjectID, 0, len(req.ItemIDs))
	for _, id := range req.ItemIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			utils.HTTPErrorHandler(c, http.StatusBadRequest, "item_ids must be closet item ids", nil)
			return
		}
		objIDs = append(objIDs, objID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := config.GetDB()

	cursor, err := db.Collection("wardrobeItems").Find(ctx, bson.M{
		"_id":    bson.M{"$in": objIDs},
		"userId": userObjID,
	})
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to load items", err)
		return
	}
	var items []models.WardrobeItem
	if err := cursor.All(ctx, &items); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to load items", err)
		return
	}
	if len(items) == 0 {
		utils.HTTPErrorHandler(c, http.StatusNotFound, "No matching closet items found", nil)
		return
	}

	key := engine.ComboKeyForItems(items)
	if key == "|" {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Outfit must include a top, bottom or full-body item", nil)
		return
	}

	record := models.WearRecord{
		UserID:   userObjID,
		ComboKey: key,
		ItemIDs:  req.ItemIDs,
		WornAt:   time.Now(),
	}
	if _, err := db.Collection("wearHistory").InsertOne(ctx, record); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Failed to log worn outfit", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Outfit logged as worn", record)
}
