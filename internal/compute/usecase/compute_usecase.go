package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type ComputeUseCase struct {
	repo        domain.ComputeRepository
	volRepo     domain.VolumeRepository
	provider    provider.ComputeProvider
	mappingRepo mapping.MappingRepository
}

func NewComputeUseCase(repo domain.ComputeRepository, volRepo domain.VolumeRepository, prov provider.ComputeProvider) *ComputeUseCase {
	return &ComputeUseCase{
		repo:     repo,
		volRepo:  volRepo,
		provider: prov,
	}
}

func (uc *ComputeUseCase) SetMappingRepository(mapRepo mapping.MappingRepository) {
	uc.mappingRepo = mapRepo
}

func (uc *ComputeUseCase) ListPlans() []*domain.ComputePlan {
	return []*domain.ComputePlan{
		{ID: "plan-0.5", Name: "ANARVA-0.5", ACU: 0.5, VCPU: 0.5, MemoryMB: 1024, StorageLimitGB: 10, NetworkLimitMbps: 100, Description: "Micro worker workload for lightweight background tasks.", Status: "ACTIVE"},
		{ID: "plan-1.0", Name: "ANARVA-1", ACU: 1.0, VCPU: 1.0, MemoryMB: 2048, StorageLimitGB: 20, NetworkLimitMbps: 250, Description: "Standard single vCPU node for API gateways and web services.", Status: "ACTIVE"},
		{ID: "plan-2.0", Name: "ANARVA-2", ACU: 2.0, VCPU: 2.0, MemoryMB: 4096, StorageLimitGB: 40, NetworkLimitMbps: 500, Description: "Balanced dual-core node for production microservices.", Status: "ACTIVE"},
		{ID: "plan-4.0", Name: "ANARVA-4", ACU: 4.0, VCPU: 4.0, MemoryMB: 8192, StorageLimitGB: 80, NetworkLimitMbps: 1000, Description: "Performance node for medium databases and compute jobs.", Status: "ACTIVE"},
		{ID: "plan-8.0", Name: "ANARVA-8", ACU: 8.0, VCPU: 8.0, MemoryMB: 16384, StorageLimitGB: 160, NetworkLimitMbps: 2500, Description: "High-throughput cluster instance.", Status: "ACTIVE"},
		{ID: "plan-16.0", Name: "ANARVA-16", ACU: 16.0, VCPU: 16.0, MemoryMB: 32768, StorageLimitGB: 320, NetworkLimitMbps: 5000, Description: "Enterprise intensive compute server.", Status: "ACTIVE"},
		{ID: "plan-32.0", Name: "ANARVA-32", ACU: 32.0, VCPU: 32.0, MemoryMB: 65536, StorageLimitGB: 640, NetworkLimitMbps: 10000, Description: "Large scale analytical worker node.", Status: "ACTIVE"},
		{ID: "plan-64.0", Name: "ANARVA-64", ACU: 64.0, VCPU: 64.0, MemoryMB: 131072, StorageLimitGB: 1280, NetworkLimitMbps: 20000, Description: "Hyperscale distributed worker.", Status: "ACTIVE"},
		{ID: "plan-128.0", Name: "ANARVA-128", ACU: 128.0, VCPU: 128.0, MemoryMB: 262144, StorageLimitGB: 2560, NetworkLimitMbps: 40000, Description: "Maximum ACU capacity instance.", Status: "ACTIVE"},
	}
}

func (uc *ComputeUseCase) ListImages() []*domain.ComputeImage {
	return []*domain.ComputeImage{
		{ID: "img-ubuntu-24", Name: "Ubuntu 24.04 LTS (Noble Numbat)", Slug: "ubuntu-24.04", Version: "24.04", Type: "OS_IMAGE", Provider: "LOCAL_DOCKER", Architecture: "x86_64", Status: "ACTIVE", Description: "Official Ubuntu LTS Linux container/OS image."},
		{ID: "img-debian-12", Name: "Debian 12 (Bookworm)", Slug: "debian-12", Version: "12.0", Type: "OS_IMAGE", Provider: "LOCAL_DOCKER", Architecture: "x86_64", Status: "ACTIVE", Description: "Stable Debian Linux container/OS image."},
		{ID: "img-alpine-320", Name: "Alpine Linux 3.20", Slug: "alpine-3.20", Version: "3.20", Type: "OS_IMAGE", Provider: "LOCAL_DOCKER", Architecture: "x86_64", Status: "ACTIVE", Description: "Minimal lightweight security-oriented Linux image."},
		{ID: "img-container", Name: "Custom Container Image (Docker Hub)", Slug: "custom-container", Version: "latest", Type: "DOCKER", Provider: "LOCAL_DOCKER", Architecture: "x86_64", Status: "ACTIVE", Description: "Run any arbitrary public or registry container image."},
	}
}

func (uc *ComputeUseCase) CreateInstance(ctx context.Context, inst *domain.ComputeInstance) (*domain.ComputeInstance, error) {
	if !domain.IsValidACU(inst.ACU) {
		return nil, appErrors.New(appErrors.CodeInvalidInput, fmt.Sprintf("invalid ACU capacity %.1f. Must be one of 0.5, 1, 2, 4, 8, 16, 32, 64, 128", inst.ACU))
	}

	if inst.Name == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "instance name is required")
	}

	if inst.ID == "" {
		inst.ID = fmt.Sprintf("acu-inst-%d", time.Now().UnixNano()%900000+100000)
	}

	inst.Slug = inst.Name
	inst.VCPU = inst.ACU
	inst.MemoryMB = int(inst.ACU * 2048)
	inst.StorageGB = int(inst.ACU * 10)
	inst.ResourceID = domain.GenerateComputeARNV(inst.RegionID, inst.ProjectID, inst.Name)
	inst.Status = domain.StatusProvisioning
	inst.Health = domain.HealthHealthy
	inst.DeletedAt = nil
	inst.CreatedAt = time.Now()
	inst.UpdatedAt = time.Now()

	// Provision via provider
	created, err := uc.provider.CreateInstance(ctx, inst)
	if err != nil {
		return nil, err
	}

	if uc.repo != nil {
		if err := uc.repo.Create(ctx, created); err != nil {
			return nil, err
		}
	}

	if uc.mappingRepo != nil {
		_ = uc.mappingRepo.SaveMapping(&mapping.ProviderResourceMapping{
			AnarvaResourceID:     created.ID,
			OrganizationID:       created.OrganizationID,
			ProjectID:            created.ProjectID,
			Provider:             string(created.Provider),
			ProviderResourceID:   created.ProviderInstanceID,
			ProviderResourceType: "COMPUTE_INSTANCE",
			Region:               created.RegionID,
			Zone:                 created.ZoneID,
			Status:               string(created.Status),
			Managed:              true,
		})
	}

	return created, nil
}

func (uc *ComputeUseCase) GetInstance(ctx context.Context, id string) (*domain.ComputeInstance, error) {
	if inst, err := uc.ensureProviderRehydrated(ctx, id); err == nil && inst != nil {
		return inst, nil
	}
	return uc.provider.GetInstance(ctx, id)
}

func (uc *ComputeUseCase) GetInstanceForTenant(ctx context.Context, orgID, projID, id string) (*domain.ComputeInstance, error) {
	var inst *domain.ComputeInstance
	var err error

	if uc.repo != nil {
		inst, err = uc.repo.GetTenantScopedByID(ctx, orgID, projID, id)
	} else {
		inst, err = uc.GetInstance(ctx, id)
		if err == nil {
			if orgID != "" && inst.OrganizationID != "" && inst.OrganizationID != orgID {
				return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Organization '%s' is prohibited from accessing compute instance '%s'", orgID, id)
			}
			if projID != "" && inst.ProjectID != "" && inst.ProjectID != projID {
				return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Project '%s' is prohibited from accessing compute instance '%s'", projID, id)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	if rehydratable, ok := uc.provider.(provider.RehydratableProvider); ok && inst != nil {
		if rehydrErr := rehydratable.RehydrateInstance(ctx, inst); rehydrErr != nil {
			return nil, rehydrErr
		}
	}

	return inst, nil
}

func (uc *ComputeUseCase) ensureProviderRehydrated(ctx context.Context, id string) (*domain.ComputeInstance, error) {
	if uc.repo == nil {
		return nil, fmt.Errorf("repository is nil")
	}
	inst, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rehydratable, ok := uc.provider.(provider.RehydratableProvider); ok && inst != nil {
		if rehydrErr := rehydratable.RehydrateInstance(ctx, inst); rehydrErr != nil {
			return nil, rehydrErr
		}
	}
	return inst, nil
}

func (uc *ComputeUseCase) ListInstances(ctx context.Context, projectID string) ([]*domain.ComputeInstance, error) {
	if uc.repo != nil {
		if list, err := uc.repo.ListByProjectID(ctx, projectID); err == nil && len(list) > 0 {
			for _, inst := range list {
				if rehydratable, ok := uc.provider.(provider.RehydratableProvider); ok && inst != nil {
					_ = rehydratable.RehydrateInstance(ctx, inst)
				}
			}
			return list, nil
		}
	}
	return uc.provider.ListInstances(ctx, projectID)
}

func (uc *ComputeUseCase) ListInstancesForTenant(ctx context.Context, orgID, projectID string) ([]*domain.ComputeInstance, error) {
	var list []*domain.ComputeInstance
	var err error
	if uc.repo != nil {
		list, err = uc.repo.ListByProjectID(ctx, projectID)
	} else {
		list, err = uc.provider.ListInstances(ctx, projectID)
	}
	if err != nil {
		return nil, err
	}

	filtered := make([]*domain.ComputeInstance, 0)
	for _, inst := range list {
		if (orgID == "" || inst.OrganizationID == "" || inst.OrganizationID == orgID) &&
			(projectID == "" || inst.ProjectID == "" || inst.ProjectID == projectID) {
			if rehydratable, ok := uc.provider.(provider.RehydratableProvider); ok && inst != nil {
				_ = rehydratable.RehydrateInstance(ctx, inst)
			}
			filtered = append(filtered, inst)
		}
	}
	return filtered, nil
}

func (uc *ComputeUseCase) StartInstance(ctx context.Context, id string) error {
	_, _ = uc.ensureProviderRehydrated(ctx, id)
	err := uc.provider.StartInstance(ctx, id)
	if err == nil && uc.repo != nil {
		if inst, getErr := uc.repo.GetByID(ctx, id); getErr == nil && inst != nil {
			inst.Status = domain.StatusRunning
			inst.Health = domain.HealthHealthy
			_ = uc.repo.Update(ctx, inst)
		}
	}
	return err
}

func (uc *ComputeUseCase) StopInstance(ctx context.Context, id string) error {
	_, _ = uc.ensureProviderRehydrated(ctx, id)
	err := uc.provider.StopInstance(ctx, id)
	if err == nil && uc.repo != nil {
		if inst, getErr := uc.repo.GetByID(ctx, id); getErr == nil && inst != nil {
			inst.Status = domain.StatusStopped
			inst.Health = domain.HealthUnavailable
			_ = uc.repo.Update(ctx, inst)
		}
	}
	return err
}

func (uc *ComputeUseCase) RestartInstance(ctx context.Context, id string) error {
	_, _ = uc.ensureProviderRehydrated(ctx, id)
	err := uc.provider.RestartInstance(ctx, id)
	if err == nil && uc.repo != nil {
		if inst, getErr := uc.repo.GetByID(ctx, id); getErr == nil && inst != nil {
			inst.Status = domain.StatusRunning
			inst.Health = domain.HealthHealthy
			_ = uc.repo.Update(ctx, inst)
		}
	}
	return err
}

func (uc *ComputeUseCase) DeleteInstance(ctx context.Context, id string) error {
	_, _ = uc.ensureProviderRehydrated(ctx, id)
	if uc.mappingRepo != nil {
		_ = uc.mappingRepo.DeleteMapping(ctx, id)
	}
	if uc.repo != nil {
		_ = uc.repo.Delete(ctx, id)
	}
	return uc.provider.DeleteInstance(ctx, id)
}

func (uc *ComputeUseCase) ExecuteCommand(ctx context.Context, id string, req *domain.CommandExecutionRequest) (*domain.CommandExecutionResult, error) {
	if req.Command == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "command payload cannot be empty")
	}
	_, _ = uc.ensureProviderRehydrated(ctx, id)
	return uc.provider.ExecuteCommand(ctx, id, req)
}

func (uc *ComputeUseCase) GetInstanceMetrics(ctx context.Context, id string) (map[string]interface{}, error) {
	_, _ = uc.ensureProviderRehydrated(ctx, id)
	return uc.provider.GetInstanceMetrics(ctx, id)
}
