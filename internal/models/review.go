package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Review represents a review document
type Review struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"userId" json:"userId"`
	ProductID        primitive.ObjectID `bson:"productId" json:"productId"`
	Rating           int                `bson:"rating" json:"rating"`
	Title            string             `bson:"title" json:"title"`
	Comment          string             `bson:"comment" json:"comment"`
	VerifiedPurchase bool               `bson:"verifiedPurchase" json:"verifiedPurchase"`
	Helpful          int                `bson:"helpful" json:"helpful"`
	Images           []string           `bson:"images" json:"images"`
	CreatedAt        time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// ReviewStats represents review statistics
type ReviewStats struct {
	AverageRating float64     `json:"averageRating"`
	TotalReviews  int         `json:"totalReviews"`
	Distribution  map[int]int `json:"distribution"` // Rating count distribution
}
