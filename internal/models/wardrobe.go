package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WardrobeItem represents a wardrobe item document
type WardrobeItem struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"userId" json:"userId"`
	Name         string             `bson:"name" json:"name"`
	Category     string             `bson:"category" json:"category"` // upper, lower, shoes
	Subtype      string             `bson:"subtype" json:"subtype"`   // t-shirt, jeans, etc
	Color        string             `bson:"color" json:"color"`
	Image        string             `bson:"image" json:"image"`
	Brand        string             `bson:"brand" json:"brand"`
	Size         string             `bson:"size" json:"size"`
	PurchaseDate time.Time          `bson:"purchaseDate" json:"purchaseDate"`
	Price        float64            `bson:"price" json:"price"`
	Tags         []string           `bson:"tags" json:"tags"`
	Condition    string             `bson:"condition" json:"condition"` // new, like-new, good, fair
	Notes        string             `bson:"notes" json:"notes"`
	IsActive     bool               `bson:"isActive" json:"isActive"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// WardrobeStats represents wardrobe statistics
type WardrobeStats struct {
	TotalItems  int            `json:"totalItems"`
	ByCategory  map[string]int `json:"byCategory"`
	ByCondition map[string]int `json:"byCondition"`
	AverageAge  float64        `json:"averageAge"` // in days
}
