package resource

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/arnv"
)

type ResourceStatus string

const (
	StatusCreating    ResourceStatus = "CREATING"
	StatusAvailable   ResourceStatus = "AVAILABLE"
	StatusUpdating    ResourceStatus = "UPDATING"
	StatusDeleting    ResourceStatus = "DELETING"
	StatusDeleted     ResourceStatus = "DELETED"
	StatusFailed      ResourceStatus = "FAILED"
	StatusStopped     ResourceStatus = "STOPPED"
	StatusComingSoon  ResourceStatus = "COMING_SOON"
	StatusMaintenance ResourceStatus = "MAINTENANCE"
)

type ResourceType string

const (
	TypeDatabase      ResourceType = "DATABASE"
	TypeStorageBucket ResourceType = "STORAGE_BUCKET"
	TypeCompute       ResourceType = "COMPUTE"
	TypeNetwork       ResourceType = "NETWORK"
	TypeSubnet        ResourceType = "SUBNET"
	TypeVolume        ResourceType = "VOLUME"
	TypeLoadBalancer  ResourceType = "LOAD_BALANCER"
	TypeBackup        ResourceType = "BACKUP"
	TypeSnapshot      ResourceType = "SNAPSHOT"
	TypeDNSZone       ResourceType = "DNS_ZONE"
)

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CloudResource struct {
	ID                 string         `json:"id"`
	ResourceID         string         `json:"resourceId"`
	Name               string         `json:"name"`
	Type               ResourceType   `json:"type"`
	Status             ResourceStatus `json:"status"`
	OrganizationID     string         `json:"organizationId"`
	ProjectID          string         `json:"projectId"`
	Environment        string         `json:"environment"`
	RegionID           string         `json:"regionId"`
	OwnerID            string         `json:"ownerId"`
	Provider           string         `json:"provider"`
	ProviderResourceID string         `json:"providerResourceId,omitempty"`
	ParentResourceID   string         `json:"parentResourceId,omitempty"`
	Dependencies       []string       `json:"dependencies,omitempty"`
	Tags               []Tag          `json:"tags"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

type Registry struct {
	mu        sync.RWMutex
	resources map[string]*CloudResource
}

func NewRegistry() *Registry {
	r := &Registry{
		resources: make(map[string]*CloudResource),
	}
	r.seedDefaultResources()
	return r
}

func (r *Registry) seedDefaultResources() {
	now := time.Now()
	defaults := []*CloudResource{
		{
			ID:                 "res-db-prod-1",
			ResourceID:         arnv.GenerateARNV("DATABASE", "ap-hyderabad-1", "proj-default", "production-db"),
			Name:               "production-db",
			Type:               TypeDatabase,
			Status:             StatusAvailable,
			OrganizationID:     "org-default",
			ProjectID:          "proj-default",
			Environment:        "Production",
			RegionID:           "ap-hyderabad-1",
			OwnerID:            "usr-default",
			Provider:           "LOCAL_DOCKER",
			ProviderResourceID: "container-db-prod-1",
			ParentResourceID:   "res-vpc-prod-1",
			Tags:               []Tag{{Key: "Environment", Value: "Production"}},
			CreatedAt:          now.Add(-48 * time.Hour),
			UpdatedAt:          now,
		},
		{
			ID:                 "res-s3-assets-1",
			ResourceID:         arnv.GenerateARNV("STORAGE_BUCKET", "ap-hyderabad-1", "proj-default", "anarva-media-assets"),
			Name:               "anarva-media-assets",
			Type:               TypeStorageBucket,
			Status:             StatusAvailable,
			OrganizationID:     "org-default",
			ProjectID:          "proj-default",
			Environment:        "Production",
			RegionID:           "ap-hyderabad-1",
			OwnerID:            "usr-default",
			Provider:           "LOCAL_STORAGE",
			ProviderResourceID: "bucket-media-assets",
			Tags:               []Tag{{Key: "Access", Value: "Public-Read"}},
			CreatedAt:          now.Add(-12 * time.Hour),
			UpdatedAt:          now,
		},
		{
			ID:                 "res-ace-worker-1",
			ResourceID:         arnv.GenerateARNV("COMPUTE", "us-east-1", "proj-default", "ace-worker-node-01"),
			Name:               "ace-worker-node-01",
			Type:               TypeCompute,
			Status:             StatusAvailable,
			OrganizationID:     "org-default",
			ProjectID:          "proj-default",
			Environment:        "Production",
			RegionID:           "us-east-1",
			OwnerID:            "usr-default",
			Provider:           "LOCAL_DOCKER",
			ProviderResourceID: "container-ace-worker-01",
			ParentResourceID:   "res-vpc-prod-1",
			Tags:               []Tag{{Key: "Role", Value: "Compute"}},
			CreatedAt:          now.Add(-6 * time.Hour),
			UpdatedAt:          now,
		},
		{
			ID:                 "res-vpc-prod-1",
			ResourceID:         arnv.GenerateARNV("NETWORK", "us-east-1", "proj-default", "primary-production-vpc"),
			Name:               "Primary Production VPC",
			Type:               TypeNetwork,
			Status:             StatusAvailable,
			OrganizationID:     "org-default",
			ProjectID:          "proj-default",
			Environment:        "Production",
			RegionID:           "us-east-1",
			OwnerID:            "usr-default",
			Provider:           "LOCAL_DOCKER",
			ProviderResourceID: "docker-net-vpc-prod-1",
			Dependencies:       []string{"res-db-prod-1", "res-ace-worker-1"},
			Tags:               []Tag{{Key: "Subnet", Value: "10.0.0.0/16"}},
			CreatedAt:          now.Add(-72 * time.Hour),
			UpdatedAt:          now,
		},
	}

	for _, item := range defaults {
		r.resources[item.ID] = item
	}
}

func (r *Registry) Create(res *CloudResource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if res.ID == "" {
		res.ID = fmt.Sprintf("res-%d", time.Now().UnixNano())
	}
	if res.ResourceID == "" {
		res.ResourceID = arnv.GenerateARNV(string(res.Type), res.RegionID, res.ProjectID, res.Name)
	}
	now := time.Now()
	res.CreatedAt = now
	res.UpdatedAt = now

	r.resources[res.ID] = res
	return nil
}

func (r *Registry) GetByID(id, orgID string) (*CloudResource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res, ok := r.resources[id]
	if !ok || res.Status == StatusDeleted {
		return nil, fmt.Errorf("resource not found")
	}
	if orgID != "" && res.OrganizationID != orgID {
		return nil, fmt.Errorf("access denied: cross-tenant authorization violation")
	}
	return res, nil
}

func (r *Registry) Update(id, orgID string, updater func(*CloudResource)) (*CloudResource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, ok := r.resources[id]
	if !ok || res.Status == StatusDeleted {
		return nil, fmt.Errorf("resource not found")
	}
	if orgID != "" && res.OrganizationID != orgID {
		return nil, fmt.Errorf("access denied: cross-tenant authorization violation")
	}

	updater(res)
	res.UpdatedAt = time.Now()
	return res, nil
}

func (r *Registry) CheckDependenciesBeforeDelete(id, orgID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res, ok := r.resources[id]
	if !ok || res.Status == StatusDeleted {
		return nil, fmt.Errorf("resource not found")
	}
	if orgID != "" && res.OrganizationID != orgID {
		return nil, fmt.Errorf("access denied: cross-tenant authorization violation")
	}

	var dependentIDs []string
	for _, child := range r.resources {
		if child.Status != StatusDeleted && child.ParentResourceID == id {
			dependentIDs = append(dependentIDs, child.Name)
		}
	}
	return dependentIDs, nil
}

func (r *Registry) SafeDelete(id, orgID string) error {
	deps, err := r.CheckDependenciesBeforeDelete(id, orgID)
	if err != nil {
		return err
	}
	if len(deps) > 0 {
		return fmt.Errorf("cannot delete resource '%s': %d dependent resources exist (%s)", id, len(deps), strings.Join(deps, ", "))
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	res, ok := r.resources[id]
	if !ok {
		return fmt.Errorf("resource not found")
	}
	res.Status = StatusDeleted
	res.UpdatedAt = time.Now()
	return nil
}

func (r *Registry) List(orgID, projectID, regionID, resType, statusQuery, query string) []*CloudResource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*CloudResource
	for _, res := range r.resources {
		if res.Status == StatusDeleted {
			continue
		}
		if orgID != "" && res.OrganizationID != orgID {
			continue
		}
		if projectID != "" && res.ProjectID != projectID {
			continue
		}
		if regionID != "" && res.RegionID != regionID {
			continue
		}
		if resType != "" && string(res.Type) != strings.ToUpper(resType) {
			continue
		}
		if statusQuery != "" && string(res.Status) != strings.ToUpper(statusQuery) {
			continue
		}
		if query != "" {
			qLower := strings.ToLower(query)
			if !strings.Contains(strings.ToLower(res.Name), qLower) &&
				!strings.Contains(strings.ToLower(res.ResourceID), qLower) &&
				!strings.Contains(strings.ToLower(res.ID), qLower) {
				continue
			}
		}
		result = append(result, res)
	}
	return result
}
