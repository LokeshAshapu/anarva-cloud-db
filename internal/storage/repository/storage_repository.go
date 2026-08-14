package repository

import (
	"fmt"
	"sync"

	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
)

type StorageRepository struct {
	mu       sync.RWMutex
	accounts map[string]*domain.StorageAccount
	buckets  map[string]*domain.Bucket
	objects  map[string]*domain.Object
	keys     map[string]*domain.StorageAccessKey
}

func NewStorageRepository() *StorageRepository {
	return &StorageRepository{
		accounts: make(map[string]*domain.StorageAccount),
		buckets:  make(map[string]*domain.Bucket),
		objects:  make(map[string]*domain.Object),
		keys:     make(map[string]*domain.StorageAccessKey),
	}
}

func (r *StorageRepository) SaveAccount(acc *domain.StorageAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accounts[acc.ID] = acc
	return nil
}

func (r *StorageRepository) SaveBucket(b *domain.Bucket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buckets[b.ID] = b
	return nil
}

func (r *StorageRepository) GetBucket(id string) (*domain.Bucket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if b, ok := r.buckets[id]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("bucket '%s' not found", id)
}

func (r *StorageRepository) ListBuckets(orgID, projectID string) ([]*domain.Bucket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.Bucket
	for _, b := range r.buckets {
		if b.Status != "DELETED" {
			res = append(res, b)
		}
	}
	return res, nil
}

func (r *StorageRepository) SaveObject(obj *domain.Object) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.objects[obj.ID] = obj
	return nil
}

func (r *StorageRepository) ListObjects(bucketID string) ([]*domain.Object, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.Object
	for _, obj := range r.objects {
		if obj.BucketID == bucketID && obj.Status != "DELETED" {
			res = append(res, obj)
		}
	}
	return res, nil
}
