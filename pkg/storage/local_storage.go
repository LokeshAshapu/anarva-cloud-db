package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type localStorageProvider struct {
	basePath string
}

func NewLocalStorageProvider(basePath string) (StorageProvider, error) {
	if basePath == "" {
		basePath = "./data/storage"
	}
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}
	return &localStorageProvider{basePath: basePath}, nil
}

func (s *localStorageProvider) Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	fullPath := filepath.Join(s.basePath, key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeInternal, "failed to create storage subdirectory")
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeInternal, "failed to create storage file")
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeInternal, "failed to write storage data")
	}

	return fullPath, nil
}

func (s *localStorageProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, key)
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, appErrors.New(appErrors.CodeNotFound, "storage object not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeInternal, "failed to open storage file")
	}
	return file, nil
}

func (s *localStorageProvider) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.basePath, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return appErrors.Wrap(err, appErrors.CodeInternal, "failed to delete storage object")
	}
	return nil
}

func (s *localStorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := filepath.Join(s.basePath, key)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
