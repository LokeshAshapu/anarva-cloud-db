# Anarva Cloud Platform — Universal Provider Contract Specification

## Overview
This document specifies the provider-neutral Go interfaces defined in `pkg/providers`.

---

## Provider Domain Contracts

### 1. Compute Provider (`ComputeProvider`)
- `LaunchInstance(ctx context.Context, opts ComputeInstanceOpts) (*ComputeInstanceDetails, error)`
- `StartInstance(ctx context.Context, instanceID string) error`
- `StopInstance(ctx context.Context, instanceID string) error`
- `TerminateInstance(ctx context.Context, instanceID string) error`
- `ScaleComputeACU(ctx context.Context, instanceID string, targetACU float64) error`
- `GetInstanceMetrics(ctx context.Context, instanceID string) (map[string]float64, error)`

### 2. Database Provider (`DatabaseProvider`)
- `ProvisionInstance(ctx context.Context, opts DatabaseProvisionOpts) (*dbDomain.DatabaseInstance, error)`
- `StartInstance(ctx context.Context, instanceID string) error`
- `StopInstance(ctx context.Context, instanceID string) error`
- `TerminateInstance(ctx context.Context, instanceID string) error`
- `ScaleACU(ctx context.Context, instanceID string, minACU, maxACU float64) error`
- `CreateReadReplica(ctx context.Context, primaryInstanceID, region string) (*dbDomain.DatabaseInstance, error)`
- `GetInstanceHealth(ctx context.Context, instanceID string) (string, error)`

### 3. Storage Provider (`StorageProvider`)
- `CreateBucket(ctx context.Context, bucketName string, policy BucketPolicy) error`
- `DeleteBucket(ctx context.Context, bucketName string) error`
- `ListBuckets(ctx context.Context, orgID string) ([]string, error)`
- `PutObject(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) (*ObjectMetadata, error)`
- `GetObject(ctx context.Context, bucketName, key string) (io.ReadCloser, *ObjectMetadata, error)`
- `DeleteObject(ctx context.Context, bucketName, key string) error`
- `ListObjects(ctx context.Context, bucketName, prefix string) ([]*ObjectMetadata, error)`
- `GenerateSignedURL(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error)`

### 4. Network Provider (`NetworkProvider`)
- `CreateVPC(ctx context.Context, spec VPCSpec) (*VPCSpec, error)`
- `DeleteVPC(ctx context.Context, vpcID string) error`
- `ConfigureSecurityGroup(ctx context.Context, vpcID string, rules []SecurityRule) error`
- `GetVPCHealth(ctx context.Context, vpcID string) (string, error)`
