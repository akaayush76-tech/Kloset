package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TryOnRecord represents a persisted try-on result in MongoDB.
type TryOnRecord struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"userId" json:"userId"`
	ProductID       string             `bson:"productId" json:"productId"`
	ProductName     string             `bson:"productName" json:"productName"`
	ProductImage    string             `bson:"productImage" json:"productImage"`
	ProductCategory string             `bson:"productCategory" json:"productCategory"`
	AvatarUrl       string             `bson:"avatarUrl" json:"avatarUrl"`
	ResultUrl       string             `bson:"resultUrl" json:"resultUrl"`
	Size            string             `bson:"size,omitempty" json:"size,omitempty"`
	Color           string             `bson:"color,omitempty" json:"color,omitempty"`
	Fit             string             `bson:"fit" json:"fit"`
	RecommendedSize string             `bson:"recommendedSize,omitempty" json:"recommendedSize,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
}
