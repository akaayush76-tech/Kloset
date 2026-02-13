package config

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

// InitMongoDB initializes MongoDB connection pool
func InitMongoDB(ctx context.Context) error {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	// Create MongoDB client options
	opts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(5 * time.Minute)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return err
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	MongoClient = client
	return nil
}

// GetDB returns the MongoDB database instance
func GetDB() *mongo.Database {
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "kloset_dev"
	}
	return MongoClient.Database(dbName)
}

// CloseDB closes the MongoDB connection
func CloseDB(ctx context.Context) error {
	if MongoClient != nil {
		return MongoClient.Disconnect(ctx)
	}
	return nil
}
