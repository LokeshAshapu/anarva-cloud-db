package provider

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
)

// S3APIClient defines the interface for S3 SDK calls to enable unit testing and mocking.
type S3APIClient interface {
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error)
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

type S3Presigner interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type S3StorageProvider struct {
	client        S3APIClient
	presignClient S3Presigner
	bucketName    string
	region        string
	endpoint      string
}

func NewS3StorageProvider(cfg config.StorageConfig) (*S3StorageProvider, error) {
	if cfg.S3AccessKey == "" || cfg.S3SecretKey == "" {
		return nil, fmt.Errorf("STORAGE_CONFIG_MISSING: STORAGE_S3_ACCESS_KEY and STORAGE_S3_SECRET_KEY are required")
	}

	region := cfg.S3Region
	if region == "" {
		region = "us-east-1"
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.S3Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.S3Endpoint,
				SigningRegion:     region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithRegion(region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")),
		awsConfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("STORAGE_CONNECTION_FAILED: failed to load S3 SDK config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.S3Endpoint != "" {
			o.UsePathStyle = true
		}
	})

	presignClient := s3.NewPresignClient(client)

	return &S3StorageProvider{
		client:        client,
		presignClient: presignClient,
		bucketName:    cfg.S3Bucket,
		region:        region,
		endpoint:      cfg.S3Endpoint,
	}, nil
}

func NewS3StorageProviderWithClient(client S3APIClient, presignClient S3Presigner, bucketName, region, endpoint string) *S3StorageProvider {
	return &S3StorageProvider{
		client:        client,
		presignClient: presignClient,
		bucketName:    bucketName,
		region:        region,
		endpoint:      endpoint,
	}
}

func (p *S3StorageProvider) GetProviderType() string {
	return "S3"
}

func (p *S3StorageProvider) getBucketName(bucketID string) string {
	if p.bucketName != "" {
		return p.bucketName
	}
	return bucketID
}

func (p *S3StorageProvider) CreateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error) {
	if bucket == nil || bucket.Name == "" {
		return nil, fmt.Errorf("STORAGE_SECURITY_RISK: Bucket specification cannot be nil or empty")
	}

	targetBucket := bucket.Name
	input := &s3.CreateBucketInput{
		Bucket: aws.String(targetBucket),
	}

	if p.region != "us-east-1" && p.region != "auto" {
		input.CreateBucketConfiguration = &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(p.region),
		}
	}

	_, err := p.client.CreateBucket(ctx, input)
	if err != nil {
		if !strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") && !strings.Contains(err.Error(), "BucketAlreadyExists") {
			return nil, fmt.Errorf("STORAGE_UPLOAD_FAILED: failed to create S3 bucket '%s': %w", targetBucket, err)
		}
	}

	bucket.Provider = "S3"
	bucket.Status = "ACTIVE"
	bucket.RealityLabel = "S3_COMPATIBLE_STORAGE (REAL_CLOUD)"
	bucket.CreatedAt = time.Now()
	bucket.UpdatedAt = time.Now()

	return bucket, nil
}

func (p *S3StorageProvider) DeleteBucket(ctx context.Context, bucketID string) error {
	targetBucket := p.getBucketName(bucketID)
	_, err := p.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(targetBucket),
	})
	if err != nil {
		return fmt.Errorf("STORAGE_DELETE_FAILED: failed to delete S3 bucket '%s': %w", targetBucket, err)
	}
	return nil
}

func (p *S3StorageProvider) ListBuckets(ctx context.Context, orgID, projectID string) ([]*domain.Bucket, error) {
	out, err := p.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("STORAGE_CONNECTION_FAILED: failed to list S3 buckets: %w", err)
	}

	var buckets []*domain.Bucket
	for _, b := range out.Buckets {
		name := aws.ToString(b.Name)
		buckets = append(buckets, &domain.Bucket{
			ID:           name,
			OrganizationID: orgID,
			ProjectID:    projectID,
			Name:         name,
			Provider:     "S3",
			Status:       "ACTIVE",
			RealityLabel: "S3_COMPATIBLE_STORAGE (REAL_CLOUD)",
			CreatedAt:    aws.ToTime(b.CreationDate),
			UpdatedAt:    aws.ToTime(b.CreationDate),
		})
	}

	return buckets, nil
}

func (p *S3StorageProvider) GetBucket(ctx context.Context, bucketID string) (*domain.Bucket, error) {
	targetBucket := p.getBucketName(bucketID)
	_, err := p.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(targetBucket),
	})
	if err != nil {
		return nil, fmt.Errorf("STORAGE_BUCKET_NOT_FOUND: S3 bucket '%s' not found or inaccessible: %w", targetBucket, err)
	}

	return &domain.Bucket{
		ID:           bucketID,
		Name:         targetBucket,
		Provider:     "S3",
		Status:       "ACTIVE",
		RealityLabel: "S3_COMPATIBLE_STORAGE (REAL_CLOUD)",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (p *S3StorageProvider) PutObject(ctx context.Context, bucketID, key string, data io.Reader, size int64, contentType string) (*domain.Object, error) {
	if err := ValidateObjectKey(key); err != nil {
		return nil, err
	}

	targetBucket := p.getBucketName(bucketID)
	input := &s3.PutObjectInput{
		Bucket: aws.String(targetBucket),
		Key:    aws.String(key),
		Body:   data,
	}

	if size > 0 {
		input.ContentLength = aws.Int64(size)
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	out, err := p.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("STORAGE_UPLOAD_FAILED: failed to upload object '%s' to S3 bucket '%s': %w", key, targetBucket, err)
	}

	eTag := ""
	if out.ETag != nil {
		eTag = aws.ToString(out.ETag)
	}

	cat := domain.CategoryOther
	if contentType != "" {
		if strings.HasPrefix(contentType, "image") {
			cat = domain.CategoryImages
		} else if strings.HasPrefix(contentType, "video") {
			cat = domain.CategoryVideos
		}
	}

	return &domain.Object{
		ID:           fmt.Sprintf("s3-obj-%d", time.Now().UnixNano()),
		BucketID:     bucketID,
		Key:          key,
		Size:         size,
		ContentType:  contentType,
		Category:     cat,
		ETag:         eTag,
		StorageClass: domain.StorageStandard,
		VersionID:    "v1",
		Status:       "ACTIVE",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (p *S3StorageProvider) GetObject(ctx context.Context, bucketID, key string) (io.ReadCloser, *domain.Object, error) {
	if err := ValidateObjectKey(key); err != nil {
		return nil, nil, err
	}

	targetBucket := p.getBucketName(bucketID)
	out, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(targetBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("STORAGE_OBJECT_NOT_FOUND: failed to retrieve object '%s' from S3 bucket '%s': %w", key, targetBucket, err)
	}

	size := aws.ToInt64(out.ContentLength)
	contentType := aws.ToString(out.ContentType)
	eTag := aws.ToString(out.ETag)

	obj := &domain.Object{
		ID:           fmt.Sprintf("s3-obj-%s", key),
		BucketID:     bucketID,
		Key:          key,
		Size:         size,
		ContentType:  contentType,
		ETag:         eTag,
		StorageClass: domain.StorageStandard,
		Status:       "ACTIVE",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return out.Body, obj, nil
}

func (p *S3StorageProvider) DeleteObject(ctx context.Context, bucketID, key string) error {
	if err := ValidateObjectKey(key); err != nil {
		return err
	}

	targetBucket := p.getBucketName(bucketID)
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(targetBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("STORAGE_DELETE_FAILED: failed to delete object '%s' from S3 bucket '%s': %w", key, targetBucket, err)
	}
	return nil
}

func (p *S3StorageProvider) ListObjects(ctx context.Context, bucketID, prefix string) ([]*domain.Object, error) {
	targetBucket := p.getBucketName(bucketID)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(targetBucket),
	}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	out, err := p.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("STORAGE_CONNECTION_FAILED: failed to list objects in S3 bucket '%s': %w", targetBucket, err)
	}

	var objects []*domain.Object
	for _, item := range out.Contents {
		key := aws.ToString(item.Key)
		objects = append(objects, &domain.Object{
			ID:           fmt.Sprintf("s3-obj-%s", key),
			BucketID:     bucketID,
			Key:          key,
			Size:         aws.ToInt64(item.Size),
			ETag:         aws.ToString(item.ETag),
			StorageClass: domain.StorageStandard,
			Status:       "ACTIVE",
			CreatedAt:    aws.ToTime(item.LastModified),
			UpdatedAt:    aws.ToTime(item.LastModified),
		})
	}

	return objects, nil
}

func (p *S3StorageProvider) GenerateSignedURL(ctx context.Context, bucketID, key, method string, expiresSec int) (*domain.PresignedURL, error) {
	if err := ValidateObjectKey(key); err != nil {
		return nil, err
	}

	if expiresSec <= 0 {
		expiresSec = 900
	}
	duration := time.Duration(expiresSec) * time.Second
	targetBucket := p.getBucketName(bucketID)

	var req *v4.PresignedHTTPRequest
	var err error

	if strings.ToUpper(method) == "PUT" {
		req, err = p.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(targetBucket),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(duration))
	} else {
		req, err = p.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(targetBucket),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(duration))
	}

	if err != nil {
		return nil, fmt.Errorf("STORAGE_CONNECTION_FAILED: failed to generate presigned S3 URL: %w", err)
	}

	return &domain.PresignedURL{
		URL:       req.URL,
		Method:    strings.ToUpper(method),
		ExpiresAt: time.Now().Add(duration),
	}, nil
}
