package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserPreferences represents user preferences
type UserPreferences struct {
	Newsletter      bool     `bson:"newsletter" json:"newsletter"`
	Notifications   bool     `bson:"notifications" json:"notifications"`
	PreferredColors []string `bson:"preferredColors" json:"preferredColors"`
	PreferredSizes  []string `bson:"preferredSizes" json:"preferredSizes"`
}

// User represents a user document
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Email       string             `bson:"email" json:"email"`
	Password    string             `bson:"password" json:"-"`
	Phone       string             `bson:"phone" json:"phone"`
	Gender      string             `bson:"gender" json:"gender"`
	DateOfBirth time.Time          `bson:"dateOfBirth" json:"dateOfBirth"`
	Avatar      string             `bson:"avatar" json:"avatar"`

	// Nested arrays for relationships
	Preferences   UserPreferences      `bson:"preferences" json:"preferences"`
	Wishlist      []primitive.ObjectID `bson:"wishlist" json:"wishlist"`
	Cart          []CartItem           `bson:"cart" json:"cart"`
	Orders        []primitive.ObjectID `bson:"orders" json:"orders"`
	Reviews       []primitive.ObjectID `bson:"reviews" json:"reviews"`
	WardrobeItems []primitive.ObjectID `bson:"wardrobeItems" json:"wardrobeItems"`

	IsActive  bool      `bson:"isActive" json:"isActive"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// CartItem represents an item in user's cart
type CartItem struct {
	ProductID primitive.ObjectID `bson:"productId" json:"productId"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Size      string             `bson:"size" json:"size"`
	Color     string             `bson:"color" json:"color,omitempty"`
	Price     float64            `bson:"price" json:"price"`
	AddedAt   time.Time          `bson:"addedAt" json:"addedAt"`
}
