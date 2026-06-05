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

// CreateOrderRequest represents order creation request
type CreateOrderRequest struct {
	ShippingAddress models.ShippingAddress `json:"shippingAddress" binding:"required"`
	PaymentMethod   string                 `json:"paymentMethod" binding:"required"`
}

// CreateOrderHandler creates a new order from cart
func CreateOrderHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate shipping address
	if !utils.NameValidator(req.ShippingAddress.Name) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Address name must be 2-50 characters", nil)
		return
	}
	if !utils.AddressValidator(req.ShippingAddress.Address) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Address must be 10-200 characters", nil)
		return
	}
	if !utils.PincodeValidator(req.ShippingAddress.Pincode) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid pincode (must be 6 digits)", nil)
		return
	}
	if !utils.PhoneValidator(req.ShippingAddress.Phone) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid phone number", nil)
		return
	}

	// Validate payment method
	if !utils.PaymentMethodValidator(req.PaymentMethod) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid payment method", nil)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	usersCollection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user and cart
	var user models.User
	err = usersCollection.FindOne(ctx, bson.M{"_id": userObjID}).Decode(&user)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusNotFound, "User not found", err)
		return
	}

	if len(user.Cart) == 0 {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Cart is empty", nil)
		return
	}

	// Create order items from cart
	var orderItems []models.OrderItem
	totalAmount := 0.0

	for _, cartItem := range user.Cart {
		orderItems = append(orderItems, models.OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Size:      cartItem.Size,
			Color:     cartItem.Color,
			Price:     cartItem.Price,
		})
		totalAmount += cartItem.Price * float64(cartItem.Quantity)
	}

	// Create order
	order := models.Order{
		ID:              primitive.NewObjectID(),
		UserID:          userObjID,
		Items:           orderItems,
		TotalAmount:     totalAmount,
		ShippingAddress: req.ShippingAddress,
		PaymentMethod:   req.PaymentMethod,
		PaymentStatus:   "pending",
		OrderStatus:     "pending",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Insert order
	ordersCollection := db.Collection("orders")
	result, err := ordersCollection.InsertOne(ctx, order)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error creating order", err)
		return
	}

	// Clear user cart and add order to user
	update := bson.M{
		"$set": bson.M{
			"cart":      []models.CartItem{},
			"updatedAt": time.Now(),
		},
		"$push": bson.M{
			"orders": result.InsertedID,
		},
	}

	_, err = usersCollection.UpdateOne(ctx, bson.M{"_id": userObjID}, update)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error updating user", err)
		return
	}

	order.ID = result.InsertedID.(primitive.ObjectID)
	utils.SuccessResponse(c, http.StatusCreated, "Order created successfully", order)
}

// GetOrdersHandler retrieves user's orders with pagination
func GetOrdersHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	status := c.Query("status")

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
	collection := db.Collection("orders")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter
	filter := bson.M{"userId": userObjID}
	if status != "" {
		filter["orderStatus"] = status
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error counting orders", err)
		return
	}

	// Find orders
	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching orders", err)
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err = cursor.All(ctx, &orders); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding orders", err)
		return
	}

	if orders == nil {
		orders = []models.Order{}
	}

	pagination := utils.PaginationMeta(page, limit, total)
	utils.SuccessResponseWithPagination(c, http.StatusOK, "Orders retrieved successfully", orders, pagination)
}

// GetOrderHandler retrieves a single order
func GetOrderHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	orderID := c.Param("id")
	orderObjID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid order ID", err)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("orders")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var order models.Order
	err = collection.FindOne(ctx, bson.M{"_id": orderObjID, "userId": userObjID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Order not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching order", err)
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Order retrieved successfully", order)
}

// UpdateOrderStatusRequest represents status update request
type UpdateOrderStatusRequest struct {
	OrderStatus    string `json:"orderStatus" binding:"required"`
	TrackingNumber string `json:"trackingNumber"`
	PaymentStatus  string `json:"paymentStatus"`
}

// UpdateOrderStatusHandler updates order status
func UpdateOrderStatusHandler(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Admin access required", nil)
		return
	}

	orderID := c.Param("id")
	orderObjID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid order ID", err)
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if !utils.OrderStatusValidator(req.OrderStatus) {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid order status", nil)
		return
	}

	db := config.GetDB()
	collection := db.Collection("orders")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"orderStatus": req.OrderStatus,
			"updatedAt":   time.Now(),
		},
	}

	if req.TrackingNumber != "" {
		update["$set"].(bson.M)["trackingNumber"] = req.TrackingNumber
	}

	if req.PaymentStatus != "" {
		update["$set"].(bson.M)["paymentStatus"] = req.PaymentStatus
	}

	if req.OrderStatus == "delivered" {
		update["$set"].(bson.M)["deliveredAt"] = time.Now()
	}

	result := collection.FindOneAndUpdate(ctx, bson.M{"_id": orderObjID}, update)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Order not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error updating order", result.Err())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Order status updated successfully", nil)
}

// CancelOrderHandler cancels a pending order
func CancelOrderHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	orderID := c.Param("id")
	orderObjID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid order ID", err)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	db := config.GetDB()
	collection := db.Collection("orders")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get order
	var order models.Order
	err = collection.FindOne(ctx, bson.M{"_id": orderObjID, "userId": userObjID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.HTTPErrorHandler(c, http.StatusNotFound, "Order not found", nil)
		} else {
			utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching order", err)
		}
		return
	}

	// Check if order can be cancelled (only pending orders)
	if order.OrderStatus != "pending" && order.OrderStatus != "confirmed" {
		utils.HTTPErrorHandler(c, http.StatusBadRequest, "Order cannot be cancelled at this stage", nil)
		return
	}

	// Update order status
	update := bson.M{
		"$set": bson.M{
			"orderStatus":   "cancelled",
			"paymentStatus": "refunded",
			"updatedAt":     time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": orderObjID}, update)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error cancelling order", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Order cancelled successfully", nil)
}

// GetOrderStatsHandler retrieves order statistics
func GetOrderStatsHandler(c *gin.Context) {
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
	collection := db.Collection("orders")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all orders for user
	cursor, err := collection.Find(ctx, bson.M{"userId": userObjID})
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error fetching orders", err)
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err = cursor.All(ctx, &orders); err != nil {
		utils.HTTPErrorHandler(c, http.StatusInternalServerError, "Error decoding orders", err)
		return
	}

	// Calculate stats
	stats := gin.H{
		"totalOrders": len(orders),
		"totalSpent":  0.0,
		"statusBreakdown": gin.H{
			"pending":   0,
			"confirmed": 0,
			"shipped":   0,
			"delivered": 0,
			"cancelled": 0,
		},
	}

	totalSpent := 0.0
	for _, order := range orders {
		totalSpent += order.TotalAmount
		statusCount := stats["statusBreakdown"].(gin.H)[order.OrderStatus]
		stats["statusBreakdown"].(gin.H)[order.OrderStatus] = statusCount.(int) + 1
	}

	stats["totalSpent"] = totalSpent

	utils.SuccessResponse(c, http.StatusOK, "Order statistics retrieved successfully", stats)
}
