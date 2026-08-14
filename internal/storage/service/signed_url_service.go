package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
)

type SignedURLService struct {
	prov      provider.ObjectStorageProvider
	secretKey []byte
}

func NewSignedURLService(prov provider.ObjectStorageProvider) *SignedURLService {
	return &SignedURLService{
		prov:      prov,
		secretKey: []byte("anarva-storage-secret-key-2026"),
	}
}

func (s *SignedURLService) GenerateSignedURL(ctx context.Context, bucketID, key, method string, expiresSec int) (*domain.PresignedURL, error) {
	if expiresSec <= 0 {
		expiresSec = 3600
	}

	exp := time.Now().Add(time.Duration(expiresSec) * time.Second)
	payload := fmt.Sprintf("%s:%s:%s:%d", method, bucketID, key, exp.Unix())

	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(payload))
	sig := hex.EncodeToString(h.Sum(nil))

	rawURL := fmt.Sprintf("http://localhost:8080/api/v1/storage/buckets/%s/object/%s?signature=%s&expires=%d", bucketID, key, sig, exp.Unix())

	return &domain.PresignedURL{
		URL:       rawURL,
		Method:    method,
		Bucket:    bucketID,
		Key:       key,
		ExpiresAt: exp,
	}, nil
}
