package provider

import (
	"fmt"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
)

// NewStorageProvider creates an ObjectStorageProvider based on configuration and environment.
// In production mode (ANARVA_ENV=production or APP_ENV=production), selecting LocalStorageProvider or
// providing invalid S3 configuration fails closed immediately.
func NewStorageProvider(cfg config.StorageConfig, environment string) (ObjectStorageProvider, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	env := strings.ToLower(strings.TrimSpace(environment))

	if env == "production" {
		if driver == "local" || driver == "" {
			return nil, fmt.Errorf("CONFIG_VALIDATION_FAILURE: Production mode (ANARVA_ENV=production) forbids LocalStorageProvider (STORAGE_PROVIDER=local). STORAGE_PROVIDER=s3 with valid S3/R2 credentials is required.")
		}
		if driver != "s3" {
			return nil, fmt.Errorf("CONFIG_VALIDATION_FAILURE: Unsupported storage provider '%s' in production mode. Must be 's3'.", driver)
		}
	}

	switch driver {
	case "s3":
		if cfg.S3Bucket == "" || cfg.S3AccessKey == "" || cfg.S3SecretKey == "" {
			return nil, fmt.Errorf("STORAGE_CONFIG_MISSING: STORAGE_PROVIDER=s3 requires STORAGE_S3_BUCKET, STORAGE_S3_ACCESS_KEY, and STORAGE_S3_SECRET_KEY")
		}
		return NewS3StorageProvider(cfg)

	case "local", "":
		if env == "production" {
			return nil, fmt.Errorf("CONFIG_VALIDATION_FAILURE: LocalStorageProvider cannot be initialized in production environment")
		}
		return NewLocalStorageProvider(cfg.LocalPath), nil

	default:
		return nil, fmt.Errorf("STORAGE_PROVIDER_UNSUPPORTED: Unknown storage provider driver '%s'", driver)
	}
}
