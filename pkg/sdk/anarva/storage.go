package anarva

import (
	"context"
	"encoding/json"
	"fmt"
)

type StorageBucket struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	ProjectID      string `json:"projectId"`
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	Region         string `json:"region"`
	StorageClass   string `json:"storageClass"`
	Versioning     bool   `json:"versioning"`
	PublicAccess   string `json:"publicAccess"`
	EncryptionMode string `json:"encryptionMode"`
	Status         string `json:"status"`
	RealityLabel   string `json:"realityLabel"`
}

type StorageObject struct {
	ID          string `json:"id"`
	BucketID    string `json:"bucketId"`
	Key         string `json:"key"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	Category    string `json:"category"`
	ETag        string `json:"etag"`
}

type PresignedURL struct {
	URL       string `json:"url"`
	Method    string `json:"method"`
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	ExpiresAt string `json:"expiresAt"`
}

type StorageService struct {
	client *Client
}

func (s *StorageService) ListBuckets(ctx context.Context, orgID, projectID string) ([]*StorageBucket, error) {
	path := fmt.Sprintf("/api/v1/storage/buckets?organizationId=%s&projectId=%s", orgID, projectID)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data []*StorageBucket `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *StorageService) CreateBucket(ctx context.Context, orgID, projectID, name, region string) (*StorageBucket, error) {
	body := map[string]string{
		"organizationId": orgID,
		"projectId":      projectID,
		"name":           name,
		"region":         region,
	}
	res, err := s.client.doRequest(ctx, "POST", "/api/v1/storage/buckets", body, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *StorageBucket `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *StorageService) GenerateSignedURL(ctx context.Context, bucketID, key, method string, expiresSec int) (*PresignedURL, error) {
	path := fmt.Sprintf("/api/v1/storage/buckets/%s/signed-url", bucketID)
	body := map[string]interface{}{
		"key":        key,
		"method":     method,
		"expiresSec": expiresSec,
	}
	res, err := s.client.doRequest(ctx, "POST", path, body, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *PresignedURL `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
