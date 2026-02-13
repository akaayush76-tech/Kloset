package config

import (
	"context"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
)

var CloudinaryClient *cloudinary.Cloudinary

// InitCloudinary initializes Cloudinary SDK
func InitCloudinary(ctx context.Context) error {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		// Return nil if credentials not set - can be optional in development
		return nil
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return err
	}

	CloudinaryClient = cld
	return nil
}

// GetCloudinary returns the Cloudinary client
func GetCloudinary() *cloudinary.Cloudinary {
	return CloudinaryClient
}
