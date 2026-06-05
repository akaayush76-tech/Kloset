package models

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateUserIndexes creates all indexes for User collection
func CreateUserIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")
	indexModel := mongo.IndexModel{
		Keys:    map[string]int{"email": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	// Index on createdAt descending for sorting
	createdAtIndex := mongo.IndexModel{
		Keys: map[string]int{"createdAt": -1},
	}
	_, err = collection.Indexes().CreateOne(ctx, createdAtIndex)
	return err
}

// CreateProductIndexes creates all indexes for Product collection
func CreateProductIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("products")

	// Text search indexes
	textIndex := mongo.IndexModel{
		Keys: map[string]interface{}{"name": "text", "description": "text", "brand": "text"},
	}
	_, err := collection.Indexes().CreateOne(ctx, textIndex)
	if err != nil {
		return err
	}

	// Compound index: category + subcategory
	categoryIndex := mongo.IndexModel{
		Keys: map[string]int{"category": 1, "subcategory": 1},
	}
	_, err = collection.Indexes().CreateOne(ctx, categoryIndex)
	if err != nil {
		return err
	}

	// Price index
	priceIndex := mongo.IndexModel{
		Keys: map[string]int{"price": 1},
	}
	_, err = collection.Indexes().CreateOne(ctx, priceIndex)
	if err != nil {
		return err
	}

	// Rating index
	ratingIndex := mongo.IndexModel{
		Keys: map[string]int{"rating": -1},
	}
	_, err = collection.Indexes().CreateOne(ctx, ratingIndex)
	if err != nil {
		return err
	}

	// CreatedAt index
	createdAtIndex := mongo.IndexModel{
		Keys: map[string]int{"createdAt": -1},
	}
	_, err = collection.Indexes().CreateOne(ctx, createdAtIndex)
	if err != nil {
		return err
	}

	// Brand index
	brandIndex := mongo.IndexModel{
		Keys: map[string]int{"brand": 1},
	}
	_, err = collection.Indexes().CreateOne(ctx, brandIndex)
	return err
}

// CreateOrderIndexes creates all indexes for Order collection
func CreateOrderIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("orders")

	// Compound index: user + createdAt descending
	userCreatedIndex := mongo.IndexModel{
		Keys: map[string]int{"userId": 1, "createdAt": -1},
	}
	_, err := collection.Indexes().CreateOne(ctx, userCreatedIndex)
	if err != nil {
		return err
	}

	// OrderStatus index
	statusIndex := mongo.IndexModel{
		Keys: map[string]int{"orderStatus": 1},
	}
	_, err = collection.Indexes().CreateOne(ctx, statusIndex)
	if err != nil {
		return err
	}

	// PaymentStatus index
	paymentIndex := mongo.IndexModel{
		Keys: map[string]int{"paymentStatus": 1},
	}
	_, err = collection.Indexes().CreateOne(ctx, paymentIndex)
	if err != nil {
		return err
	}

	// TrackingNumber index
	trackingIndex := mongo.IndexModel{
		Keys: map[string]int{"trackingNumber": 1},
	}
	_, err = collection.Indexes().CreateOne(ctx, trackingIndex)
	return err
}

// CreateReviewIndexes creates all indexes for Review collection
func CreateReviewIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("reviews")

	// Unique constraint: user + product
	uniqueIndex := mongo.IndexModel{
		Keys:    map[string]int{"userId": 1, "productId": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, uniqueIndex)
	if err != nil {
		return err
	}

	// Compound index: product + createdAt
	productCreatedIndex := mongo.IndexModel{
		Keys: map[string]int{"productId": 1, "createdAt": -1},
	}
	_, err = collection.Indexes().CreateOne(ctx, productCreatedIndex)
	if err != nil {
		return err
	}

	// Rating index
	ratingIndex := mongo.IndexModel{
		Keys: map[string]int{"rating": -1},
	}
	_, err = collection.Indexes().CreateOne(ctx, ratingIndex)
	if err != nil {
		return err
	}

	// Helpful index
	helpfulIndex := mongo.IndexModel{
		Keys: map[string]int{"helpful": -1},
	}
	_, err = collection.Indexes().CreateOne(ctx, helpfulIndex)
	return err
}

// CreateWardrobeItemIndexes creates all indexes for WardrobeItem collection
func CreateWardrobeItemIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("wardrobeItems")

	// Compound index: user + category
	userCategoryIndex := mongo.IndexModel{
		Keys: map[string]int{"userId": 1, "category": 1},
	}
	_, err := collection.Indexes().CreateOne(ctx, userCategoryIndex)
	if err != nil {
		return err
	}

	// Compound index: user + subtype
	userSubtypeIndex := mongo.IndexModel{
		Keys: map[string]int{"userId": 1, "subtype": 1},
	}
	_, err = collection.Indexes().CreateOne(ctx, userSubtypeIndex)
	if err != nil {
		return err
	}

	// Compound index: user + createdAt
	userCreatedIndex := mongo.IndexModel{
		Keys: map[string]int{"userId": 1, "createdAt": -1},
	}
	_, err = collection.Indexes().CreateOne(ctx, userCreatedIndex)
	if err != nil {
		return err
	}

	// Compound index for recommendation engine: user + isActive + occasion
	recIndex := mongo.IndexModel{
		Keys: map[string]interface{}{"userId": 1, "isActive": 1, "identifiers.occasion": 1},
	}
	_, err = collection.Indexes().CreateOne(ctx, recIndex)
	return err
}
