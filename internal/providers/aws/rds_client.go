package aws

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type RDSInstanceInfo struct {
	DBInstanceIdentifier string
	Engine               string
	EngineVersion        string
	DBInstanceClass      string
	AllocatedStorageGB   int
	StorageEncrypted     bool
	PubliclyAccessible   bool
	Status               string // creating, available, modifying, deleting, failed
	EndpointAddress      string
	EndpointPort         int
	DBName               string
	MasterUsername       string
	SubnetGroupName      string
	SecurityGroupIDs     []string
	Tags                 map[string]string
	CreatedAt            time.Time
}

type RDSCreateParams struct {
	DBInstanceIdentifier string
	Engine               string
	EngineVersion        string
	DBInstanceClass      string
	AllocatedStorageGB   int
	MasterUsername       string
	MasterUserPassword   string
	DBName               string
	SubnetGroupName      string
	SecurityGroupIDs     []string
	PubliclyAccessible   bool
	StorageEncrypted     bool
	Tags                 map[string]string
}

// RDSClient abstracts AWS RDS operations for real AWS SDK v2 and mock testing
type RDSClient interface {
	VerifyConnectivity(ctx context.Context) error
	CreateDBInstance(ctx context.Context, params RDSCreateParams) (*RDSInstanceInfo, error)
	DescribeDBInstances(ctx context.Context, identifier string) (*RDSInstanceInfo, error)
	DeleteDBInstance(ctx context.Context, identifier string, skipFinalSnapshot bool) error
	VerifySubnetGroup(ctx context.Context, subnetGroupName string) (bool, error)
}

// MockRDSClient provides in-memory simulated RDS responses for unit testing & development
type MockRDSClient struct {
	mu          sync.RWMutex
	instances   map[string]*RDSInstanceInfo
	isConnected bool
}

func NewMockRDSClient(isConnected bool) *MockRDSClient {
	return &MockRDSClient{
		instances:   make(map[string]*RDSInstanceInfo),
		isConnected: isConnected,
	}
}

func (m *MockRDSClient) VerifyConnectivity(ctx context.Context) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: Invalid AWS credentials for RDS operation")
	}
	return nil
}

func (m *MockRDSClient) VerifySubnetGroup(ctx context.Context, subnetGroupName string) (bool, error) {
	if !m.isConnected {
		return false, fmt.Errorf("AUTH_FAILED: AWS API call unauthorized")
	}
	if subnetGroupName == "" || strings.HasPrefix(subnetGroupName, "dbsubnet-") || subnetGroupName == "default-db-subnet-group" {
		return true, nil
	}
	return true, nil
}

func (m *MockRDSClient) CreateDBInstance(ctx context.Context, params RDSCreateParams) (*RDSInstanceInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API CreateDBInstance unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	inst := &RDSInstanceInfo{
		DBInstanceIdentifier: params.DBInstanceIdentifier,
		Engine:               params.Engine,
		EngineVersion:        params.EngineVersion,
		DBInstanceClass:      params.DBInstanceClass,
		AllocatedStorageGB:   params.AllocatedStorageGB,
		StorageEncrypted:     params.StorageEncrypted,
		PubliclyAccessible:   params.PubliclyAccessible,
		Status:               "available",
		EndpointAddress:      fmt.Sprintf("%s.c9ak15v99k.us-east-1.rds.amazonaws.com", params.DBInstanceIdentifier),
		EndpointPort:         5432,
		DBName:               params.DBName,
		MasterUsername:       params.MasterUsername,
		SubnetGroupName:      params.SubnetGroupName,
		SecurityGroupIDs:     params.SecurityGroupIDs,
		Tags:                 params.Tags,
		CreatedAt:            time.Now(),
	}
	m.instances[params.DBInstanceIdentifier] = inst
	return inst, nil
}

func (m *MockRDSClient) DescribeDBInstances(ctx context.Context, identifier string) (*RDSInstanceInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API DescribeDBInstances unauthorized")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if inst, ok := m.instances[identifier]; ok {
		return inst, nil
	}
	return nil, fmt.Errorf("DBInstanceNotFound: DBInstance %s not found", identifier)
}

func (m *MockRDSClient) DeleteDBInstance(ctx context.Context, identifier string, skipFinalSnapshot bool) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: AWS API DeleteDBInstance unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if inst, ok := m.instances[identifier]; ok {
		inst.Status = "deleting"
		delete(m.instances, identifier)
		return nil
	}
	return fmt.Errorf("DBInstanceNotFound: DBInstance %s not found", identifier)
}
