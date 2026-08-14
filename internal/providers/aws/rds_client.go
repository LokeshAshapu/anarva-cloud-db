package aws

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RDSInstanceInfo struct {
	DBInstanceIdentifier      string
	Engine                    string
	EngineVersion             string
	DBInstanceClass           string
	AllocatedStorageGB        int
	StorageEncrypted          bool
	PubliclyAccessible        bool
	MultiAZ                   bool
	AvailabilityZone          string
	SecondaryAvailabilityZone string
	Status                    string // creating, available, modifying, rebooting, failover, deleting, failed
	EndpointAddress           string
	EndpointPort              int
	DBName                    string
	MasterUsername            string
	SubnetGroupName           string
	SecurityGroupIDs          []string
	Tags                      map[string]string
	CreatedAt                 time.Time
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
	MultiAZ              bool
	Tags                 map[string]string
}

type RDSSnapshotInfo struct {
	SnapshotIdentifier   string
	DBInstanceIdentifier string
	Engine               string
	EngineVersion        string
	AllocatedStorageGB   int
	Status               string // creating, available, deleting, failed
	StorageEncrypted     bool
	KMSKeyID             string
	SnapshotType         string // manual, automated
	SnapshotCreateTime   time.Time
}

type RecoveryWindowInfo struct {
	DBInstanceIdentifier   string
	EarliestRestorableTime time.Time
	LatestRestorableTime   time.Time
}

// RDSClient abstracts AWS RDS operations for real AWS SDK v2 and mock testing
type RDSClient interface {
	VerifyConnectivity(ctx context.Context) error
	CreateDBInstance(ctx context.Context, params RDSCreateParams) (*RDSInstanceInfo, error)
	DescribeDBInstances(ctx context.Context, identifier string) (*RDSInstanceInfo, error)
	DeleteDBInstance(ctx context.Context, identifier string, skipFinalSnapshot bool) error
	VerifySubnetGroup(ctx context.Context, subnetGroupName string) (bool, error)
	CreateDBSnapshot(ctx context.Context, dbIdentifier, snapshotIdentifier string) (*RDSSnapshotInfo, error)
	DeleteDBSnapshot(ctx context.Context, snapshotIdentifier string) error
	DescribeDBSnapshots(ctx context.Context, dbIdentifier string) ([]*RDSSnapshotInfo, error)
	RestoreDBInstanceToPointInTime(ctx context.Context, sourceIdentifier, targetIdentifier string, restoreTime time.Time) (*RDSInstanceInfo, error)
	GetRecoveryWindow(ctx context.Context, dbIdentifier string) (*RecoveryWindowInfo, error)
	RebootDBInstance(ctx context.Context, identifier string, forceFailover bool) (*RDSInstanceInfo, error)
	ModifyDBInstanceMultiAZ(ctx context.Context, identifier string, multiAZ bool) (*RDSInstanceInfo, error)
}

// MockRDSClient provides in-memory simulated RDS responses for unit testing & development
type MockRDSClient struct {
	mu          sync.RWMutex
	instances   map[string]*RDSInstanceInfo
	snapshots   map[string]*RDSSnapshotInfo
	isConnected bool
}

func NewMockRDSClient(isConnected bool) *MockRDSClient {
	m := &MockRDSClient{
		instances:   make(map[string]*RDSInstanceInfo),
		snapshots:   make(map[string]*RDSSnapshotInfo),
		isConnected: isConnected,
	}

	// Seed primary test RDS instance
	now := time.Now()
	m.instances["anarva-rds-prod-01"] = &RDSInstanceInfo{
		DBInstanceIdentifier:      "anarva-rds-prod-01",
		Engine:                    "postgres",
		EngineVersion:             "16.2",
		DBInstanceClass:           "db.t3.micro",
		AllocatedStorageGB:        20,
		StorageEncrypted:          true,
		PubliclyAccessible:        false,
		MultiAZ:                   true,
		AvailabilityZone:          "ap-south-1a",
		SecondaryAvailabilityZone: "ap-south-1b",
		Status:                    "available",
		EndpointAddress:           "anarva-rds-prod-01.c123456789.ap-south-1.rds.amazonaws.com",
		EndpointPort:              5432,
		DBName:                    "anarva_prod",
		MasterUsername:            "anarva_admin",
		SubnetGroupName:           "default-db-subnet-group",
		SecurityGroupIDs:          []string{"sg-0a1b2c3d4e5f6g7h8"},
		CreatedAt:                 now.Add(-720 * time.Hour),
	}

	return m
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
	return true, nil
}

func (m *MockRDSClient) CreateDBInstance(ctx context.Context, params RDSCreateParams) (*RDSInstanceInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API CreateDBInstance unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	secAZ := ""
	if params.MultiAZ {
		secAZ = "ap-south-1b"
	}

	inst := &RDSInstanceInfo{
		DBInstanceIdentifier:      params.DBInstanceIdentifier,
		Engine:                    params.Engine,
		EngineVersion:             params.EngineVersion,
		DBInstanceClass:           params.DBInstanceClass,
		AllocatedStorageGB:        params.AllocatedStorageGB,
		StorageEncrypted:          params.StorageEncrypted,
		PubliclyAccessible:        params.PubliclyAccessible,
		MultiAZ:                   params.MultiAZ,
		AvailabilityZone:          "ap-south-1a",
		SecondaryAvailabilityZone: secAZ,
		Status:                    "available",
		EndpointAddress:           fmt.Sprintf("%s.c123456789.ap-south-1.rds.amazonaws.com", params.DBInstanceIdentifier),
		EndpointPort:              5432,
		DBName:                    params.DBName,
		MasterUsername:            params.MasterUsername,
		SubnetGroupName:           params.SubnetGroupName,
		SecurityGroupIDs:          params.SecurityGroupIDs,
		Tags:                      params.Tags,
		CreatedAt:                 time.Now(),
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

	inst, exists := m.instances[identifier]
	if !exists {
		return nil, fmt.Errorf("DBInstanceNotFound: %s not found", identifier)
	}
	return inst, nil
}

func (m *MockRDSClient) DeleteDBInstance(ctx context.Context, identifier string, skipFinalSnapshot bool) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: AWS API DeleteDBInstance unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.instances[identifier]; !exists {
		return fmt.Errorf("DBInstanceNotFound: %s not found", identifier)
	}
	delete(m.instances, identifier)
	return nil
}

func (m *MockRDSClient) CreateDBSnapshot(ctx context.Context, dbIdentifier, snapshotIdentifier string) (*RDSSnapshotInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API CreateDBSnapshot unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, exists := m.instances[dbIdentifier]
	if !exists {
		return nil, fmt.Errorf("DBInstanceNotFound: %s", dbIdentifier)
	}

	snap := &RDSSnapshotInfo{
		SnapshotIdentifier:   snapshotIdentifier,
		DBInstanceIdentifier: dbIdentifier,
		Engine:               inst.Engine,
		EngineVersion:        inst.EngineVersion,
		AllocatedStorageGB:   inst.AllocatedStorageGB,
		Status:               "available",
		StorageEncrypted:     inst.StorageEncrypted,
		KMSKeyID:             "arn:aws:kms:ap-south-1:123456789012:key/default-rds-key",
		SnapshotType:         "manual",
		SnapshotCreateTime:   time.Now(),
	}

	m.snapshots[snapshotIdentifier] = snap
	return snap, nil
}

func (m *MockRDSClient) DeleteDBSnapshot(ctx context.Context, snapshotIdentifier string) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: AWS API DeleteDBSnapshot unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.snapshots[snapshotIdentifier]; !exists {
		return fmt.Errorf("DBSnapshotNotFound: %s", snapshotIdentifier)
	}
	delete(m.snapshots, snapshotIdentifier)
	return nil
}

func (m *MockRDSClient) DescribeDBSnapshots(ctx context.Context, dbIdentifier string) ([]*RDSSnapshotInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API DescribeDBSnapshots unauthorized")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*RDSSnapshotInfo
	for _, snap := range m.snapshots {
		if dbIdentifier == "" || snap.DBInstanceIdentifier == dbIdentifier {
			result = append(result, snap)
		}
	}
	return result, nil
}

func (m *MockRDSClient) RestoreDBInstanceToPointInTime(ctx context.Context, sourceIdentifier, targetIdentifier string, restoreTime time.Time) (*RDSInstanceInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API RestoreDBInstanceToPointInTime unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	source, exists := m.instances[sourceIdentifier]
	if !exists {
		return nil, fmt.Errorf("DBInstanceNotFound: %s", sourceIdentifier)
	}

	restored := &RDSInstanceInfo{
		DBInstanceIdentifier:      targetIdentifier,
		Engine:                    source.Engine,
		EngineVersion:             source.EngineVersion,
		DBInstanceClass:           source.DBInstanceClass,
		AllocatedStorageGB:        source.AllocatedStorageGB,
		StorageEncrypted:          source.StorageEncrypted,
		PubliclyAccessible:        false, // Always false for network security
		MultiAZ:                   source.MultiAZ,
		AvailabilityZone:          "ap-south-1a",
		SecondaryAvailabilityZone: "ap-south-1b",
		Status:                    "available",
		EndpointAddress:           fmt.Sprintf("%s.c123456789.ap-south-1.rds.amazonaws.com", targetIdentifier),
		EndpointPort:              5432,
		DBName:                    source.DBName,
		MasterUsername:            source.MasterUsername,
		SubnetGroupName:           source.SubnetGroupName,
		SecurityGroupIDs:          source.SecurityGroupIDs,
		CreatedAt:                 time.Now(),
	}

	m.instances[targetIdentifier] = restored
	return restored, nil
}

func (m *MockRDSClient) GetRecoveryWindow(ctx context.Context, dbIdentifier string) (*RecoveryWindowInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API GetRecoveryWindow unauthorized")
	}
	now := time.Now()
	return &RecoveryWindowInfo{
		DBInstanceIdentifier:   dbIdentifier,
		EarliestRestorableTime: now.Add(-7 * 24 * time.Hour),
		LatestRestorableTime:   now.Add(-5 * time.Minute),
	}, nil
}

func (m *MockRDSClient) RebootDBInstance(ctx context.Context, identifier string, forceFailover bool) (*RDSInstanceInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API RebootDBInstance unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, exists := m.instances[identifier]
	if !exists {
		return nil, fmt.Errorf("DBInstanceNotFound: %s", identifier)
	}

	if forceFailover {
		if !inst.MultiAZ {
			return nil, fmt.Errorf("INVALID_DB_INSTANCE_STATE: Cannot force failover on single-AZ instance %s", identifier)
		}
		// Swap primary and secondary AZ
		prevPrimary := inst.AvailabilityZone
		inst.AvailabilityZone = inst.SecondaryAvailabilityZone
		inst.SecondaryAvailabilityZone = prevPrimary
	}

	inst.Status = "available"
	return inst, nil
}

func (m *MockRDSClient) ModifyDBInstanceMultiAZ(ctx context.Context, identifier string, multiAZ bool) (*RDSInstanceInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API ModifyDBInstanceMultiAZ unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, exists := m.instances[identifier]
	if !exists {
		return nil, fmt.Errorf("DBInstanceNotFound: %s", identifier)
	}

	inst.MultiAZ = multiAZ
	if multiAZ {
		inst.SecondaryAvailabilityZone = "ap-south-1b"
	} else {
		inst.SecondaryAvailabilityZone = ""
	}
	inst.Status = "available"
	return inst, nil
}
