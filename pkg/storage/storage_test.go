package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorageProvider(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "anarva_storage_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	provider, err := NewLocalStorageProvider(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	key := "backups/2026/test_snapshot.dump"
	content := []byte("ANARVA_DATABASE_BACKUP_DUMP_DATA_SAMPLE")

	// Test Upload
	filePath, err := provider.Upload(ctx, key, bytes.NewReader(content), int64(len(content)))
	require.NoError(t, err)
	assert.NotEmpty(t, filePath)

	// Test Exists
	exists, err := provider.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test Download
	reader, err := provider.Download(ctx, key)
	require.NoError(t, err)
	downloadedBytes, err := io.ReadAll(reader)
	_ = reader.Close()
	require.NoError(t, err)
	assert.Equal(t, content, downloadedBytes)

	// Test Delete
	err = provider.Delete(ctx, key)
	require.NoError(t, err)

	existsAfter, _ := provider.Exists(ctx, key)
	assert.False(t, existsAfter)
}
