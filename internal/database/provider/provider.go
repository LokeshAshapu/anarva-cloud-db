package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/arnv"
)

type EngineType string

const (
	EnginePostgreSQL EngineType = "PostgreSQL"
	EngineMySQL      EngineType = "MySQL"
)

type DatabaseCluster struct {
	ID                  string     `json:"id"`
	ResourceID          string     `json:"resourceId"`
	ProjectID           string     `json:"projectId"`
	OrganizationID      string     `json:"organizationId"`
	Name                string     `json:"name"`
	Engine              EngineType `json:"engine"`
	EngineVersion       string     `json:"engineVersion"`
	RegionID            string     `json:"regionId"`
	Environment         string     `json:"environment"`
	Status              string     `json:"status"` // CREATING, AVAILABLE, UPDATING, STOPPING, STOPPED, FAILED
	ComputeUnits        float64    `json:"computeUnits"`
	StorageGB font        int        `json:"storageGb"`
	MaxStorageGB        int        `json:"maxStorageGb"`
	AutoScalingEnabled  bool       `json:"autoScalingEnabled"`
	BackupEnabled       bool       `json:"backupEnabled"`
	BackupRetentionDays int        `json:"backupRetentionDays"`
	PITREnabled         bool       `json:"pitrEnabled"`
	HighAvailability    bool       `json:"highAvailability"`
	Host                string     `json:"host"`
	Port                int        `json:"port"`
	DBName              string     `json:"dbname"`
	Username            string     `json:"username"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type DatabaseProvider interface {
	CreateDatabase(ctx context.Context, cluster *DatabaseCluster) (*DatabaseCluster, error)
	GetDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error)
	ListDatabases(ctx context.Context, orgID, projectID string) ([]*DatabaseCluster, error)
	UpdateDatabase(ctx context.Context, id, orgID string, updater func(*DatabaseCluster)) (*DatabaseCluster, error)
	DeleteDatabase(ctx context.Context, id, orgID string) error
	StartDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error)
	StopDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error)
	RestartDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error)
}

type ControlPlaneDatabaseProvider struct {
	clusters map[string]*DatabaseCluster
}

func NewControlPlaneDatabaseProvider() *ControlPlaneDatabaseProvider {
	p := &ControlPlaneDatabaseProvider{
		clusters: make(map[string]*DatabaseCluster),
	}
	p.seedDefaults()
	return p
}

func (p *ControlPlaneDatabaseProvider) seedDefaults() {
	now := time.Now()
	p.clusters["res-db-prod-1"] = &DatabaseCluster{
		ID:                  "res-db-prod-1",
		ResourceID:          arnv.GenerateARNV("DATABASE", "ap-hyderabad-1", "proj-default", "production-db"),
		ProjectID:           "proj-default",
		OrganizationID:      "org-default",
		Name:                "production-db",
		Engine:              EnginePostgreSQL,
		EngineVersion:       "17.2",
		RegionID:            "ap-hyderabad-1",
		Environment:         "Production",
		Status:              "AVAILABLE",
		ComputeUnits:        2.0,
		StorageGB:           48,
		MaxStorageGB:        256,
		AutoScalingEnabled:  true,
		BackupEnabled:       true,
		BackupRetentionDays: 7,
		PITREnabled:         true,
		HighAvailability:    true,
		Host:                "db-prod-1.anarva.cloud",
		Port:                5432,
		DBName:              "production_db",
		Username:            "anarva_admin",
		CreatedAt:           now.Add(-48 * time.Hour),
		UpdatedAt:           now,
	}

	p.clusters["res-db-analytics-1"] = &DatabaseCluster{
		ID:                  "res-db-analytics-1",
		ResourceID:          arnv.GenerateARNV("DATABASE", "ap-mumbai-1", "proj-default", "analytics-db"),
		ProjectID:           "proj-default",
		OrganizationID:      "org-default",
		Name:                "analytics-db",
		Engine:              EnginePostgreSQL,
		EngineVersion:       "16.4",
		RegionID:            "ap-mumbai-1",
		Environment:         "Production",
		Status:              "AVAILABLE",
		ComputeUnits:        4.0,
		StorageGB:           120,
		MaxStorageGB:        512,
		AutoScalingEnabled:  true,
		BackupEnabled:       true,
		BackupRetentionDays: 14,
		PITREnabled:         true,
		HighAvailability:    false,
		Host:                "db-analytics-1.anarva.cloud",
		Port:                5432,
		DBName:              "analytics_db",
		Username:            "anarva_analytics",
		CreatedAt:           now.Add(-24 * time.Hour),
		UpdatedAt:           now,
	}
}

func (p *ControlPlaneDatabaseProvider) CreateDatabase(ctx context.Context, cluster *DatabaseCluster) (*DatabaseCluster, error) {
	if cluster.ID == "" {
		cluster.ID = fmt.Sprintf("db-%d", time.Now().UnixNano())
	}
	if cluster.ResourceID == "" {
		cluster.ResourceID = arnv.GenerateARNV("DATABASE", cluster.RegionID, cluster.ProjectID, cluster.Name)
	}
	if cluster.Status == "" {
		cluster.Status = "AVAILABLE"
	}
	if cluster.Host == "" {
		cluster.Host = fmt.Sprintf("%s.anarva.cloud", cluster.Name)
	}
	if cluster.Port == 0 {
		if cluster.Engine == EngineMySQL {
			cluster.Port = 3306
		} else {
			cluster.Port = 5432
		}
	}
	now := time.Now()
	cluster.CreatedAt = now
	cluster.UpdatedAt = now

	p.clusters[cluster.ID] = cluster
	return cluster, nil
}

func (p *ControlPlaneDatabaseProvider) GetDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error) {
	c, ok := p.clusters[id]
	if !ok || c.Status == "DELETED" {
		return nil, fmt.Errorf("database cluster not found")
	}
	if orgID != "" && c.OrganizationID != orgID {
		return nil, fmt.Errorf("authorization violation: cross-tenant access denied")
	}
	return c, nil
}

func (p *ControlPlaneDatabaseProvider) ListDatabases(ctx context.Context, orgID, projectID string) ([]*DatabaseCluster, error) {
	var result []*DatabaseCluster
	for _, c := range p.clusters {
		if c.Status == "DELETED" {
			continue
		}
		if orgID != "" && c.OrganizationID != orgID {
			continue
		}
		if projectID != "" && c.ProjectID != projectID {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

func (p *ControlPlaneDatabaseProvider) UpdateDatabase(ctx context.Context, id, orgID string, updater func(*DatabaseCluster)) (*DatabaseCluster, error) {
	c, err := p.GetDatabase(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	updater(c)
	c.UpdatedAt = time.Now()
	return c, nil
}

func (p *ControlPlaneDatabaseProvider) DeleteDatabase(ctx context.Context, id, orgID string) error {
	c, err := p.GetDatabase(ctx, id, orgID)
	if err != nil {
		return err
	}
	c.Status = "DELETED"
	c.UpdatedAt = time.Now()
	return nil
}

func (p *ControlPlaneDatabaseProvider) StartDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error) {
	return p.UpdateDatabase(ctx, id, orgID, func(c *DatabaseCluster) {
		c.Status = "AVAILABLE"
	})
}

func (p *ControlPlaneDatabaseProvider) StopDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error) {
	return p.UpdateDatabase(ctx, id, orgID, func(c *DatabaseCluster) {
		c.Status = "STOPPED"
	})
}

func (p *ControlPlaneDatabaseProvider) RestartDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error) {
	return p.UpdateDatabase(ctx, id, orgID, func(c *DatabaseCluster) {
		c.Status = "AVAILABLE"
	})
}
