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
	TypeBackup        ResourceType = "BACKUP"
	TypeReplica       ResourceType = "REPLICA"
)

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CloudResource struct {
	ID             string         `json:"id"`
	ResourceID     string         `json:"resourceId"`
	Name           string         `json:"name"`
	Type           ResourceType   `json:"type"`
	Status         ResourceStatus `json:"status"`
	OrganizationID string         `json:"organizationId"`
	ProjectID      string         `json:"projectId"`
	Environment    string         `json:"environment"`
	RegionID       string         `json:"regionId"`
	OwnerID        string         `json:"ownerId"`
	Tags           []Tag          `json:"tags"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// Registry manages in-memory resource CRUD with tenant and project isolation safeguards
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
			ID:             "res-db-prod-1",
			ResourceID:     arnv.GenerateARNV("DATABASE", "ap-hyderabad-1", "proj-default", "production-db"),
			Name:           "production-db",
			Type:           TypeDatabase,
			Status:         StatusAvailable,
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Environment:    "Production",
			RegionID:       "ap-hyderabad-1",
			OwnerID:        "usr-default",
			Tags:           []Tag{{Key: "Environment", Value: "Production"}, {Key: "Team", Value: "Engineering"}},
			CreatedAt:      now.Add(-48 * time.Hour),
			UpdatedAt:      now,
		},
		{
			ID:             "res-db-analytics-1",
			ResourceID:     arnv.GenerateARNV("DATABASE", "ap-mumbai-1", "proj-default", "analytics-db"),
			Name:           "analytics-db",
			Type:           TypeDatabase,
			Status:         StatusAvailable,
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Environment:    "Production",
			RegionID:       "ap-mumbai-1",
			OwnerID:        "usr-default",
			Tags:           []Tag{{Key: "Environment", Value: "Production"}, {Key: "Application", Value: "Analytics"}},
			CreatedAt:      now.Add(-24 * time.Hour),
			UpdatedAt:      now,
		},
		{
			ID:             "res-s3-assets-1",
			ResourceID:     arnv.GenerateARNV("STORAGE_BUCKET", "ap-hyderabad-1", "proj-default", "anarva-media-assets"),
			Name:           "anarva-media-assets",
			Type:           TypeStorageBucket,
			Status:         StatusAvailable,
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Environment:    "Production",
			RegionID:       "ap-hyderabad-1",
			OwnerID:        "usr-default",
			Tags:           []Tag{{Key: "Access", Value: "Public-Read"}},
			CreatedAt:      now.Add(-12 * time.Hour),
			UpdatedAt:      now,
		},
		{
			ID:             "res-ace-worker-1",
			ResourceID:     arnv.GenerateARNV("COMPUTE", "ap-hyderabad-1", "proj-default", "ace-worker-node-01"),
			Name:           "ace-worker-node-01",
			Type:           TypeCompute,
			Status:         StatusAvailable,
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Environment:    "Production",
			RegionID:       "ap-hyderabad-1",
			OwnerID:        "usr-default",
			Tags:           []Tag{{Key: "Role", Value: "Compute"}},
			CreatedAt:      now.Add(-6 * time.Hour),
			UpdatedAt:      now,
		},
		{
			ID:             "res-vpc-prod-1",
			ResourceID:     arnv.GenerateARNV("NETWORK", "ap-hyderabad-1", "proj-default", "anarva-primary-vpc"),
			Name:           "anarva-primary-vpc",
			Type:           TypeNetwork,
			Status:         StatusAvailable,
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Environment:    "Production",
			RegionID:       "ap-hyderabad-1",
			OwnerID:        "usr-default",
			Tags:           []Tag{{Key: "Subnet", Value: "10.0.0.0/16"}},
			CreatedAt:      now.Add(-72 * time.Hour),
			UpdatedAt:      now,
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
	// Tenant Isolation Guard
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

func (r *Registry) SafeDelete(id, orgID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, ok := r.resources[id]
	if !ok {
		return fmt.Errorf("resource not found")
	}
	if orgID != "" && res.OrganizationID != orgID {
		return fmt.Errorf("access denied: cross-tenant authorization violation")
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
