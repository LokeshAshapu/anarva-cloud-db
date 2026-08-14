package domain

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/arnv"
)

type StorageAccountStatus string

const (
	AccountCreating StorageAccountStatus = "CREATING"
	AccountActive   StorageAccountStatus = "ACTIVE"
	AccountDegraded StorageAccountStatus = "DEGRADED"
	AccountDeleting StorageAccountStatus = "DELETING"
	AccountDeleted  StorageAccountStatus = "DELETED"
	AccountFailed   StorageAccountStatus = "FAILED"
	AccountUnknown  StorageAccountStatus = "UNKNOWN"
)

type StorageClass string

const (
	StorageStandard         StorageClass = "STANDARD"
	StorageInfrequentAccess StorageClass = "INFREQUENT_ACCESS"
	StorageArchive          StorageClass = "ARCHIVE"
	StorageDeepArchive      StorageClass = "DEEP_ARCHIVE"
)

type BucketAccess string

const (
	AccessPrivate       BucketAccess = "PRIVATE"
	AccessAuthenticated BucketAccess = "AUTHENTICATED"
	AccessPublic        BucketAccess = "PUBLIC"
)

type EncryptionMode string

const (
	EncryptionNone            EncryptionMode = "NONE"
	EncryptionProviderManaged EncryptionMode = "PROVIDER_MANAGED"
	EncryptionCustomKey       EncryptionMode = "CUSTOM_KEY"
)

type ObjectCategory string

const (
	CategoryImages    ObjectCategory = "IMAGES"
	CategoryVideos    ObjectCategory = "VIDEOS"
	CategoryAudio     ObjectCategory = "AUDIO"
	CategoryDocuments ObjectCategory = "DOCUMENTS"
	CategoryLinks     ObjectCategory = "LINKS"
	CategoryOther     ObjectCategory = "OTHER"
)

type StorageAccount struct {
	ID                  string               `json:"id"`
	OrganizationID      string               `json:"organizationId"`
	ProjectID           string               `json:"projectId"`
	Name                string               `json:"name"`
	Provider            string               `json:"provider"`
	Region              string               `json:"region"`
	Status              StorageAccountStatus `json:"status"`
	DefaultStorageClass StorageClass         `json:"defaultStorageClass"`
	EncryptionMode      EncryptionMode       `json:"encryptionMode"`
	RealityLabel        string               `json:"realityLabel"`
	CreatedAt           time.Time            `json:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt"`
}

type Bucket struct {
	ID               string         `json:"id"`
	StorageAccountID string         `json:"storageAccountId"`
	OrganizationID   string         `json:"organizationId"`
	ProjectID        string         `json:"projectId"`
	Name             string         `json:"name"`
	Provider         string         `json:"provider"`
	ProviderBucketID string         `json:"providerBucketId"`
	Region           string         `json:"region"`
	StorageClass     StorageClass   `json:"storageClass"`
	Versioning       bool           `json:"versioning"`
	PublicAccess     BucketAccess   `json:"publicAccess"`
	EncryptionMode   EncryptionMode `json:"encryptionMode"`
	ObjectLock       bool           `json:"objectLock"`
	Status           string         `json:"status"`
	RealityLabel     string         `json:"realityLabel"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
}

type Object struct {
	ID               string            `json:"id"`
	BucketID         string            `json:"bucketId"`
	Key              string            `json:"key"`
	Size             int64             `json:"size"`
	ContentType      string            `json:"contentType"`
	Category         ObjectCategory    `json:"category"`
	ETag             string            `json:"etag"`
	Checksum         string            `json:"checksum"`
	StorageClass     StorageClass      `json:"storageClass"`
	VersionID        string            `json:"versionId"`
	ProviderObjectID string            `json:"providerObjectId"`
	EncryptionMode   EncryptionMode    `json:"encryptionMode"`
	Metadata         map[string]string `json:"metadata"`
	Status           string            `json:"status"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
}

type ObjectVersion struct {
	ID               string    `json:"id"`
	ObjectID         string    `json:"objectId"`
	VersionID        string    `json:"versionId"`
	Size             int64     `json:"size"`
	ETag             string    `json:"etag"`
	Checksum         string    `json:"checksum"`
	StorageReference string    `json:"storageReference"`
	IsLatest         bool      `json:"isLatest"`
	CreatedAt        time.Time `json:"createdAt"`
	DeletedAt        *time.Time `json:"deletedAt,omitempty"`
}

type MultipartUpload struct {
	ID        string    `json:"id"`
	UploadID  string    `json:"uploadId"`
	BucketID  string    `json:"bucketId"`
	Key       string    `json:"key"`
	Status    string    `json:"status"` // IN_PROGRESS, COMPLETED, ABORTED
	CreatedAt time.Time `json:"createdAt"`
}

type MultipartPart struct {
	UploadID   string `json:"uploadId"`
	PartNumber int    `json:"partNumber"`
	Size       int64  `json:"size"`
	ETag       string `json:"etag"`
	Checksum   string `json:"checksum"`
}

type PresignedURL struct {
	URL       string    `json:"url"`
	Method    string    `json:"method"` // GET, PUT
	Bucket    string    `json:"bucket"`
	Key       string    `json:"key"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type StorageAccessKey struct {
	ID              string    `json:"id"`
	OrganizationID  string    `json:"organizationId"`
	ProjectID       string    `json:"projectId"`
	Name            string    `json:"name"`
	AccessKeyID     string    `json:"accessKeyId"`
	SecretReference string    `json:"secretReference"`
	Scope           string    `json:"scope"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	LastUsedAt      time.Time `json:"lastUsedAt"`
}

func GenerateStorageBucketARNV(regionID, projectID, bucketName string) string {
	return arnv.GenerateARNV("STORAGE", regionID, projectID, fmt.Sprintf("bucket/%s", bucketName))
}
