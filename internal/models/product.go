package models

import "time"

// Product represents a product document
type Product struct {
	ID            string   `bson:"_id,omitempty" json:"id"`
	Name          string   `bson:"name" json:"name"`
	Description   string   `bson:"description" json:"description"`
	Price         float64  `bson:"price" json:"price"`
	OriginalPrice float64  `bson:"originalPrice" json:"originalPrice"`
	Discount      int      `bson:"discount" json:"discount"`
	Category      string   `bson:"category" json:"category"`
	Subcategory   string   `bson:"subcategory" json:"subcategory"`
	Brand         string   `bson:"brand" json:"brand"`
	Images        []string `bson:"images" json:"images"`
	Colors        []string `bson:"colors" json:"colors"`
	Sizes         []string `bson:"sizes" json:"sizes"`
	Rating        float64  `bson:"rating" json:"rating"`
	ReviewCount   int      `bson:"reviewCount" json:"reviewCount"`
	Features      []string `bson:"features" json:"features"`
	FitTags       []string `bson:"fitTags" json:"fitTags"`
	Combinations  []string `bson:"combinations" json:"combinations"`
	Outfits       []string `bson:"outfits" json:"outfits"`

	// Delivery information
	DeliveryDays      int     `bson:"deliveryDays" json:"deliveryDays"`
	DeliveryCharge    float64 `bson:"deliveryCharge" json:"deliveryCharge"`
	FreeDeliveryAbove float64 `bson:"freeDeliveryAbove" json:"freeDeliveryAbove"`

	IsActive  bool      `bson:"isActive" json:"isActive"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}
