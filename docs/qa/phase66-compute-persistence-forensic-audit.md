# ANARVA Cloud Phase 66 — Compute Control-Plane Persistence Forensic Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 21, 2026  
**Auditor**: Principal Cloud Architect, Distributed Systems Engineer & SRE  
**Baseline Commit**: `1df95bb` (Phase 65 Complete)  
**Audit Mode**: READ-ONLY FORENSIC AUDIT (No production code modified)  

---

## 1. Executive Summary & Current Architecture

In the current codebase (baseline `1df95bb`), compute instance control-plane metadata (`ComputeInstance` and `Volume`) is managed by `ComputeUseCase` (`internal/compute/usecase/compute_usecase.go`) and initialized in `cmd/gateway/main.go:403` using `newMemComputeRepo()` defined in `cmd/gateway/mock_repos.go:800-904`.

### Current Problem
Because compute instances are stored exclusively in an in-memory Go map (`memComputeRepo.instances map[string]*computeDomain.ComputeInstance`), **all compute instance metadata, VM specifications, IP allocations, statuses, and provider instance IDs disappear whenever the gateway process restarts or the Render container redeploys**.

```
CURRENT ARCHITECTURE (Phase 65):
API Request -> ComputeHandler -> ComputeUseCase -> memComputeRepo (Go RAM Map)
                                                        |
                                            [ PROCESS RESTART ]
                                                        v
                                              ALL COMPUTE STATE LOST!

TARGET ARCHITECTURE (Phase 66):
API Request -> ComputeHandler -> ComputeUseCase -> PostgresComputeRepository -> PostgreSQL DB (`compute_instances` & `compute_volumes`)
                                                        |
                                            [ PROCESS RESTART ]
                                                        v
                                              COMPUTE STATE PRESERVED!
```

---

## 2. Current In-Memory Persistence Locations

| Component / File | Current Implementation | Storage Type | Data Stored | Impact of Gateway Restart |
|:---|:---|:---|:---|:---|
| `cmd/gateway/main.go:403` | `compUC := computeUsecase.NewComputeUseCase(newMemComputeRepo(), nil, compProv)` | Go RAM Map | Compute instance repository reference | Memory reference lost |
| `cmd/gateway/mock_repos.go:800-904` | `memComputeRepo struct { instances map[string]*computeDomain.ComputeInstance }` | Process RAM Map | ComputeInstance structs | **TOTAL DATA LOSS** |
| `internal/compute/domain/repository.go` | Interface `ComputeRepository` & `VolumeRepository` | Unimplemented in GORM | Compute instance & volume operations | N/A |

---

## 3. Repository Interface Analysis

The existing domain repository interfaces are defined in `internal/compute/domain/repository.go`:

```go
type ComputeRepository interface {
	Create(ctx context.Context, inst *ComputeInstance) error
	GetByID(ctx context.Context, id string) (*ComputeInstance, error)
	ListByProjectID(ctx context.Context, projectID string) ([]*ComputeInstance, error)
	Update(ctx context.Context, inst *ComputeInstance) error
	Delete(ctx context.Context, id string) error
}

type VolumeRepository interface {
	CreateVolume(ctx context.Context, vol *Volume) error
	GetVolumeByID(ctx context.Context, id string) (*Volume, error)
	ListVolumesByProjectID(ctx context.Context, projectID string) ([]*Volume, error)
	UpdateVolume(ctx context.Context, vol *Volume) error
	DeleteVolume(ctx context.Context, id string) error
}
```

### Analysis of Interface Requirements
- `Create`: Inserts a new `ComputeInstance` record into PostgreSQL.
- `GetByID`: Retrieves a `ComputeInstance` record by primary key `id` (where `deleted_at IS NULL`).
- `ListByProjectID`: Retrieves all active compute instances belonging to a specific `projectID`.
- `Update`: Updates an existing `ComputeInstance` record.
- `Delete`: Soft-deletes a `ComputeInstance` record by setting `deleted_at` timestamp.

---

## 4. Compute Domain Model Analysis

The domain struct `ComputeInstance` is defined in `internal/compute/domain/compute.go`:

```go
type ComputeInstance struct {
	ID                 string                 `json:"id" gorm:"primaryKey;column:id;type:varchar(255)"`
	ResourceID         string                 `json:"resourceId" gorm:"column:resource_id;type:varchar(255);index"`
	OrganizationID     string                 `json:"organizationId" gorm:"column:organization_id;type:varchar(255);index"`
	ProjectID          string                 `json:"projectId" gorm:"column:project_id;type:varchar(255);index"`
	Name               string                 `json:"name" gorm:"column:name;type:varchar(255)"`
	Slug               string                 `json:"slug" gorm:"column:slug;type:varchar(255)"`
	RegionID           string                 `json:"regionId" gorm:"column:region_id;type:varchar(100);index"`
	ZoneID             string                 `json:"zoneId" gorm:"column:zone_id;type:varchar(100)"`
	Status             InstanceStatus         `json:"status" gorm:"column:status;type:varchar(50);index"`
	Health             InstanceHealth         `json:"health" gorm:"column:health;type:varchar(50)"`
	PlanID             string                 `json:"planId" gorm:"column:plan_id;type:varchar(100)"`
	ACU                float64                `json:"acu" gorm:"column:acu;type:numeric(10,2)"`
	VCPU               float64                `json:"vcpu" gorm:"column:vcpu;type:numeric(10,2)"`
	MemoryMB           int                    `json:"memoryMb" gorm:"column:memory_mb"`
	StorageGB          int                    `json:"storageGb" gorm:"column:storage_gb"`
	ImageID            string                 `json:"imageId" gorm:"column:image_id;type:varchar(100)"`
	DockerImage        string                 `json:"dockerImage,omitempty" gorm:"column:docker_image;type:varchar(255)"`
	NetworkID          string                 `json:"networkId" gorm:"column:network_id;type:varchar(255)"`
	SubnetID           string                 `json:"subnetId" gorm:"column:subnet_id;type:varchar(255)"`
	PrivateIP          string                 `json:"privateIp,omitempty" gorm:"column:private_ip;type:varchar(100)"`
	PublicIP           string                 `json:"publicIp,omitempty" gorm:"column:public_ip;type:varchar(100)"`
	Provider           ProviderType           `json:"provider" gorm:"column:provider;type:varchar(100)"`
	ProviderInstanceID string                 `json:"providerInstanceId,omitempty" gorm:"column:provider_instance_id;type:varchar(255);index"`
	Security           InstanceSecurityPolicy `json:"security" gorm:"-"`
	SecurityJSON       string                 `json:"-" gorm:"column:security_json;type:text"`
	EnvVars            map[string]string      `json:"envVars,omitempty" gorm:"-"`
	EnvVarsJSON        string                 `json:"-" gorm:"column:env_vars_json;type:text"`
	CreatedAt          time.Time              `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt          time.Time              `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt          *time.Time             `json:"deletedAt,omitempty" gorm:"column:deleted_at;index"`
}
```

### Complex Data Fields Serialization Handling
- **`Security` (`InstanceSecurityPolicy`)**: Contains `SSHKeyIDs`, `ServiceAccountID`, `SecurityGroupIDs`, `IAMRole`, `SecretRefs`. Will be stored as JSON text string in `security_json` column and transparently serialized/deserialized during GORM hooks (`BeforeSave`, `AfterFind`).
- **`EnvVars` (`map[string]string`)**: Environment key-value overrides for the compute instance. Will be stored as JSON text string in `env_vars_json` column and transparently serialized/deserialized during GORM hooks.

---

## 5. Proposed PostgreSQL Database Schema

```sql
CREATE TABLE compute_instances (
    id VARCHAR(255) PRIMARY KEY,
    resource_id VARCHAR(255) NOT NULL,
    organization_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    region_id VARCHAR(100) NOT NULL,
    zone_id VARCHAR(100),
    status VARCHAR(50) NOT NULL,
    health VARCHAR(50) NOT NULL,
    plan_id VARCHAR(100) NOT NULL,
    acu NUMERIC(10,2) NOT NULL,
    vcpu NUMERIC(10,2) NOT NULL,
    memory_mb INTEGER NOT NULL,
    storage_gb INTEGER NOT NULL,
    image_id VARCHAR(100) NOT NULL,
    docker_image VARCHAR(255),
    network_id VARCHAR(255),
    subnet_id VARCHAR(255),
    private_ip VARCHAR(100),
    public_ip VARCHAR(100),
    provider VARCHAR(100) NOT NULL,
    provider_instance_id VARCHAR(255),
    security_json TEXT,
    env_vars_json TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_compute_instances_org_id ON compute_instances(organization_id);
CREATE INDEX idx_compute_instances_project_id ON compute_instances(project_id);
CREATE INDEX idx_compute_instances_provider_instance_id ON compute_instances(provider_instance_id);
CREATE INDEX idx_compute_instances_deleted_at ON compute_instances(deleted_at);

CREATE TABLE compute_volumes (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    instance_id VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    size_gb INTEGER NOT NULL,
    region_id VARCHAR(100) NOT NULL,
    zone_id VARCHAR(100),
    type VARCHAR(50) NOT NULL,
    provider_volume_id VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_compute_volumes_org_id ON compute_volumes(organization_id);
CREATE INDEX idx_compute_volumes_project_id ON compute_volumes(project_id);
CREATE INDEX idx_compute_volumes_instance_id ON compute_volumes(instance_id);
```

---

## 6. Proposed Repository Architecture

A new package file `internal/compute/repository/postgres_repository.go` will be created:

```go
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
)

type PostgresComputeRepository struct {
	db *gorm.DB
}

func NewPostgresComputeRepository(db *gorm.DB) domain.ComputeRepository {
	return &PostgresComputeRepository{db: db}
}

type PostgresVolumeRepository struct {
	db *gorm.DB
}

func NewPostgresVolumeRepository(db *gorm.DB) domain.VolumeRepository {
	return &PostgresVolumeRepository{db: db}
}
```

---

## 7. Tenant Isolation & Security Requirements

- **Tenant Scoping**: All queries for `ComputeInstance` and `Volume` MUST enforce `organization_id` and `project_id` scoping.
- **Cross-Tenant Access Rejection**: `GetByID` checks `OrganizationID` and returns `TENANT_ISOLATION_VIOLATION` error if an unauthorized tenant attempts to query an instance belonging to another organization.
- **Secret Redaction**: Environment variables stored in `env_vars_json` must not expose raw system credentials or JWT secrets in API response payloads.

---

## 8. Provider Mapping Relationship (Phase 65 Integration)

When a `ComputeInstance` is created, `ComputeUseCase` records:
1. `ComputeInstance` in PostgreSQL (`compute_instances` table).
2. `ProviderResourceMapping` in PostgreSQL (`provider_resource_mappings` table) linking `inst.ID` -> `inst.ProviderInstanceID`.

---

## 9. Gateway Production Wiring & Fallback Plan

In `cmd/gateway/main.go`:
```go
// GORM AutoMigrate registration:
&computeDomain.ComputeInstance{},
&computeDomain.Volume{},

// Service instantiation wiring:
var compRepo computeDomain.ComputeRepository
var volRepo computeDomain.VolumeRepository

if dbPool != nil {
	compRepo = computeRepo.NewPostgresComputeRepository(dbPool.DB)
	volRepo = computeRepo.NewPostgresVolumeRepository(dbPool.DB)
} else if appEnv == "production" {
	log.Fatal("FATAL: Production environment requires PostgreSQL for compute control-plane metadata")
} else {
	compRepo = newMemComputeRepo()
}
```

---

## 10. Definition of Done & Test Plan

1. **Unit & Repository Tests** (`internal/compute/compute_phase66_test.go`):
   - CRUD operations for `PostgresComputeRepository` and `PostgresVolumeRepository`.
   - Tenant isolation & cross-tenant access rejection tests.
   - JSON serialization/deserialization for `Security` policies and `EnvVars`.
   - **Step 9 Restart Persistence Boundary Verification**:
     - Create compute instance in `repo1`, clear reference `repo1 = nil`, instantiate `repo2` connected to DB, retrieve instance, assert fields match 100%.
2. **Regression & Builds**:
   - `go test -v ./internal/compute/...` -> **PASS**
   - `go test -v ./cmd/gateway/...` -> **PASS**
   - `go build -o bin/gateway.exe ./cmd/gateway` -> **PASS**
   - `go build -o bin/anarva.exe ./cmd/anarva` -> **PASS**
   - `cd web && npm run build` -> **PASS**
