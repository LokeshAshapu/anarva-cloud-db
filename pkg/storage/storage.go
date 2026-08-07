package storage

import (
	"context"
	"io"
)

type StorageProvider interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
