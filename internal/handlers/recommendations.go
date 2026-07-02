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
	"github.com/kloset/backend/internal/utils"
)

type recommendRequest struct {
	TriggerItemID    string `json:"trigger_item_id"  binding:"required"`
	TriggerItemType  string `json:"trigger_item_type" binding:"required"`
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

	if !utils.TriggerItemTypeValidator(req.TriggerItemType) {
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

	if req.Limit < 1 || req.Limit > 10 {
		req.Limit = 8
	}

	uid := userID.(string)
	cacheKey := fmt.Sprintf("rec:%s:%s:%s", uid, req.TriggerItemID, req.ContextFilter)

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
