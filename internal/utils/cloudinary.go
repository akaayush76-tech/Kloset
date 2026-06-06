package utils

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/kloset/backend/internal/config"
)

// CloudinaryUploadResult holds the result of a Cloudinary upload.
type CloudinaryUploadResult struct {
	URL      string
	PublicID string
}

// UploadReaderToCloudinary uploads an io.Reader (e.g. multipart file) to Cloudinary.
func UploadReaderToCloudinary(ctx context.Context, r io.Reader, folder string) (*CloudinaryUploadResult, error) {
	return uploadAsset(ctx, r, folder)
}

// UploadDataURIToCloudinary uploads a base64 data URI string to Cloudinary.
// dataURI must start with "data:image/...".
func UploadDataURIToCloudinary(ctx context.Context, dataURI, folder string) (*CloudinaryUploadResult, error) {
	if !strings.HasPrefix(dataURI, "data:") {
		return nil, fmt.Errorf("not a data URI")
	}
	return uploadAsset(ctx, dataURI, folder)
}

func uploadAsset(ctx context.Context, asset interface{}, folder string) (*CloudinaryUploadResult, error) {
	cld := config.GetCloudinary()
	if cld == nil {
		return nil, fmt.Errorf("cloudinary not initialised — set CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET")
	}

	uploadCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := cld.Upload.Upload(uploadCtx, asset, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return &CloudinaryUploadResult{
		URL:      resp.SecureURL,
		PublicID: resp.PublicID,
	}, nil
}

// GetCloudinaryClient returns the underlying Cloudinary client (may be nil if unconfigured).
func GetCloudinaryClient() *cloudinary.Cloudinary {
	return config.GetCloudinary()
}

// DeleteFromCloudinary deletes an asset by its public ID.
func DeleteFromCloudinary(ctx context.Context, publicID string) error {
	cld := config.GetCloudinary()
	if cld == nil {
		return fmt.Errorf("cloudinary not initialised")
	}

	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := cld.Upload.Destroy(deleteCtx, uploader.DestroyParams{PublicID: publicID})
	if err != nil {
		return fmt.Errorf("cloudinary delete failed: %w", err)
	}
	return nil
}
