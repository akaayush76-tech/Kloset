package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/config"
	"github.com/kloset/backend/internal/handlers"
	"github.com/kloset/backend/internal/middleware"
	"github.com/kloset/backend/internal/utils"
)

func main() {
	// Load environment
	loadEnv()

	// Initialize services
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize MongoDB
	if err := config.InitMongoDB(ctx); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer config.CloseDB(ctx)

	// Initialize Cloudinary
	if err := config.InitCloudinary(ctx); err != nil {
		log.Printf("Cloudinary initialization: %v (optional)", err)
	}

	// Initialize rate limiter
	middleware.InitRateLimiter()

	// Create Gin router
	router := gin.Default()

	// Global middlewares
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.SecurityHeadersMiddleware)
	router.Use(middleware.ResponseTimeMiddleware)
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.RateLimitMiddleware)

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(c, 200, "Service is healthy", gin.H{"status": "ok"})
	})

	router.GET("/api", func(c *gin.Context) {
		utils.SuccessResponse(c, 200, "Kloset Backend API v1", gin.H{
			"version": "1.0.0",
			"service": "Kloset Backend",
		})
	})

	// Authentication routes
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", handlers.RegisterHandler)
		authGroup.POST("/login", handlers.LoginHandler)
		authGroup.GET("/me", middleware.AuthMiddleware, handlers.GetProfileHandler)
		authGroup.PUT("/profile", middleware.AuthMiddleware, handlers.UpdateProfileHandler)
		authGroup.PUT("/change-password", middleware.AuthMiddleware, handlers.ChangePasswordHandler)
		authGroup.POST("/logout", middleware.AuthMiddleware, handlers.LogoutHandler)
	}

	// Products routes
	productsGroup := router.Group("/api/products")
	{
		productsGroup.GET("", handlers.GetProductsHandler)
		productsGroup.GET("/categories", handlers.GetCategoriesHandler)
		productsGroup.GET("/featured", handlers.GetFeaturedHandler)
		productsGroup.GET("/:id", handlers.GetProductHandler)
		productsGroup.GET("/:id/related", handlers.GetRelatedHandler)
		productsGroup.POST("", middleware.AuthMiddleware, handlers.CreateProductHandler)
		productsGroup.PUT("/:id", middleware.AuthMiddleware, handlers.UpdateProductHandler)
		productsGroup.DELETE("/:id", middleware.AuthMiddleware, handlers.DeleteProductHandler)
	}

	// Cart routes
	cartGroup := router.Group("/api/cart")
	cartGroup.Use(middleware.AuthMiddleware)
	{
		cartGroup.GET("", handlers.GetCartHandler)
		cartGroup.GET("/count", handlers.GetCartCountHandler)
		cartGroup.POST("", handlers.AddToCartHandler)
		cartGroup.PUT("/:itemId", handlers.UpdateCartItemHandler)
		cartGroup.DELETE("/:itemId", handlers.RemoveFromCartHandler)
		cartGroup.DELETE("", handlers.ClearCartHandler)
	}

	// Orders routes
	ordersGroup := router.Group("/api/orders")
	ordersGroup.Use(middleware.AuthMiddleware)
	{
		ordersGroup.GET("", handlers.GetOrdersHandler)
		ordersGroup.GET("/stats", handlers.GetOrderStatsHandler)
		ordersGroup.POST("", handlers.CreateOrderHandler)
		ordersGroup.GET("/:id", handlers.GetOrderHandler)
		ordersGroup.PUT("/:id/status", handlers.UpdateOrderStatusHandler)
		ordersGroup.PUT("/:id/cancel", handlers.CancelOrderHandler)
	}

	// Reviews routes
	reviewsGroup := router.Group("/api/reviews")
	{
		reviewsGroup.GET("/product/:productId", handlers.GetProductReviewsHandler)
		reviewsGroup.POST("", middleware.AuthMiddleware, handlers.CreateReviewHandler)
		reviewsGroup.GET("/my", middleware.AuthMiddleware, handlers.GetMyReviewsHandler)
		reviewsGroup.GET("/stats", middleware.AuthMiddleware, handlers.GetReviewStatsHandler)
		reviewsGroup.PUT("/:id", middleware.AuthMiddleware, handlers.UpdateReviewHandler)
		reviewsGroup.DELETE("/:id", middleware.AuthMiddleware, handlers.DeleteReviewHandler)
		reviewsGroup.POST("/:id/helpful", handlers.MarkHelpfulHandler)
	}

	// Wardrobe routes
	wardrobeGroup := router.Group("/api/wardrobe")
	wardrobeGroup.Use(middleware.AuthMiddleware)
	{
		wardrobeGroup.GET("", handlers.GetWardrobeHandler)
		wardrobeGroup.GET("/stats", handlers.GetWardrobeStatsHandler)
		wardrobeGroup.GET("/category/:category", handlers.GetWardrobeByCategoryHandler)
		wardrobeGroup.POST("", handlers.CreateWardrobeItemHandler)
		wardrobeGroup.GET("/:id", handlers.GetWardrobeItemHandler)
		wardrobeGroup.PUT("/:id", handlers.UpdateWardrobeItemHandler)
		wardrobeGroup.DELETE("/:id", handlers.DeleteWardrobeItemHandler)
	}

	// Avatar routes
	avatarGroup := router.Group("/api/avatar")
	avatarGroup.Use(middleware.AuthMiddleware)
	{
		avatarGroup.GET("/check", handlers.CheckAvatarHandler)
		avatarGroup.POST("/save", handlers.SaveAvatarHandler)
	}

	// Recommendation routes
	recGroup := router.Group("/api/recommendations")
	recGroup.Use(middleware.AuthMiddleware)
	{
		recGroup.POST("/outfits", handlers.RecommendOutfitsHandler)
	}

	// Upload routes
	uploadGroup := router.Group("/api/upload")
	{
		uploadGroup.POST("/image", middleware.AuthMiddleware, handlers.UploadImageHandler)
		uploadGroup.POST("/images", middleware.AuthMiddleware, handlers.UploadImagesHandler)
		uploadGroup.POST("/wardrobe", middleware.AuthMiddleware, handlers.UploadWardrobeImageHandler)
		uploadGroup.POST("/product", middleware.AuthMiddleware, handlers.UploadProductImageHandler)
		uploadGroup.DELETE("/image", handlers.DeleteImageHandler)
		uploadGroup.GET("/optimize", handlers.OptimizeImageHandler)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger := utils.GetLogger()
	logger.Info("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Server failed: %v", err)
	}
}

// loadEnv loads environment variables from .env file
func loadEnv() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	envFile := ".env." + env
	if _, err := os.Stat(envFile); err == nil {
		log.Printf("Loading environment from %s", envFile)
		// Simple env loading - for production use github.com/joho/godotenv
		// For now, env vars should be set by system/Docker
	}
}
