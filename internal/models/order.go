package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Order represents an order document
type Order struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"userId" json:"userId"`
	Items           []OrderItem        `bson:"items" json:"items"`
	TotalAmount     float64            `bson:"totalAmount" json:"totalAmount"`
	Discount        float64            `bson:"discount" json:"discount"`
	ShippingAddress ShippingAddress    `bson:"shippingAddress" json:"shippingAddress"`
	PaymentMethod   string             `bson:"paymentMethod" json:"paymentMethod"`
	PaymentStatus   string             `bson:"paymentStatus" json:"paymentStatus"`
	OrderStatus     string             `bson:"orderStatus" json:"orderStatus"`
	TrackingNumber  string             `bson:"trackingNumber" json:"trackingNumber"`
	Notes           string             `bson:"notes" json:"notes"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
	DeliveredAt     *time.Time         `bson:"deliveredAt" json:"deliveredAt"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID   primitive.ObjectID `bson:"productId" json:"productId"`
	ProductName string             `bson:"productName" json:"productName"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	Size        string             `bson:"size" json:"size"`
	Color       string             `bson:"color" json:"color"`
	Price       float64            `bson:"price" json:"price"`
}

// ShippingAddress represents shipping address
type ShippingAddress struct {
	Name      string `bson:"name" json:"name"`
	Phone     string `bson:"phone" json:"phone"`
	Address   string `bson:"address" json:"address"`
	City      string `bson:"city" json:"city"`
	State     string `bson:"state" json:"state"`
	Pincode   string `bson:"pincode" json:"pincode"`
	Country   string `bson:"country" json:"country"`
	Landmark  string `bson:"landmark" json:"landmark"`
	IsDefault bool   `bson:"isDefault" json:"isDefault"`
}
