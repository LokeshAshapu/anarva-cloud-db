package aws

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type EC2InstanceInfo struct {
	InstanceID       string
	ImageID          string
	InstanceType     string
	State            string // pending, running, stopping, stopped, shutting-down, terminated
	PrivateIP        string
	PublicIP         string
	SubnetID         string
	VpcID            string
	AvailabilityZone string
	Tags             map[string]string
	LaunchTime       time.Time
}

type EC2RunParams struct {
	ImageID      string
	InstanceType string
	SubnetID     string
	SecurityGroup string
	KeyName      string
	MinCount     int32
	MaxCount     int32
	Tags         map[string]string
}

// EC2Client abstracts AWS EC2 operations for real AWS SDK v2 and mock testing
type EC2Client interface {
	VerifyConnectivity(ctx context.Context) error
	DescribeAvailabilityZones(ctx context.Context) ([]string, error)
	DescribeImages(ctx context.Context, imageID string) (bool, error)
	RunInstances(ctx context.Context, params EC2RunParams) (*EC2InstanceInfo, error)
	DescribeInstances(ctx context.Context, instanceID string) (*EC2InstanceInfo, error)
	StartInstances(ctx context.Context, instanceID string) error
	StopInstances(ctx context.Context, instanceID string) error
	TerminateInstances(ctx context.Context, instanceID string) error
}

// MockEC2Client provides in-memory simulated EC2 responses for unit testing & development
type MockEC2Client struct {
	mu          sync.RWMutex
	instances   map[string]*EC2InstanceInfo
	validAMIs   map[string]bool
	isConnected bool
}

func NewMockEC2Client(isConnected bool) *MockEC2Client {
	m := &MockEC2Client{
		instances:   make(map[string]*EC2InstanceInfo),
		validAMIs:   make(map[string]bool),
		isConnected: isConnected,
	}
	// Pre-populate standard Linux AMIs
	m.validAMIs["ami-0c55b159cbfafe1f0"] = true // Amazon Linux 2
	m.validAMIs["ami-0e86e20dae9224db8"] = true // Ubuntu 22.04 LTS
	m.validAMIs["ami-0123456789abcdef0"] = true // Test AMI
	return m
}

func (m *MockEC2Client) VerifyConnectivity(ctx context.Context) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: Invalid AWS credentials or unauthenticated AWS API call")
	}
	return nil
}

func (m *MockEC2Client) DescribeAvailabilityZones(ctx context.Context) ([]string, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: Cannot describe availability zones without active credentials")
	}
	return []string{"us-east-1a", "us-east-1b", "us-east-1c"}, nil
}

func (m *MockEC2Client) DescribeImages(ctx context.Context, imageID string) (bool, error) {
	if !m.isConnected {
		return false, fmt.Errorf("AUTH_FAILED: AWS API call unauthorized")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if valid, ok := m.validAMIs[imageID]; ok && valid {
		return true, nil
	}
	if strings.HasPrefix(imageID, "ami-") {
		return true, nil
	}
	return false, nil
}

func (m *MockEC2Client) RunInstances(ctx context.Context, params EC2RunParams) (*EC2InstanceInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API RunInstances unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	instID := fmt.Sprintf("i-%017d", time.Now().UnixNano()%100000000000000000)
	inst := &EC2InstanceInfo{
		InstanceID:       instID,
		ImageID:          params.ImageID,
		InstanceType:     params.InstanceType,
		State:            "running",
		PrivateIP:        "172.31.16.42",
		PublicIP:         "54.210.12.89",
		SubnetID:         params.SubnetID,
		VpcID:            "vpc-0a1b2c3d4e5f6g7h8",
		AvailabilityZone: "us-east-1a",
		Tags:             params.Tags,
		LaunchTime:       time.Now(),
	}
	m.instances[instID] = inst
	return inst, nil
}

func (m *MockEC2Client) DescribeInstances(ctx context.Context, instanceID string) (*EC2InstanceInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API DescribeInstances unauthorized")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if inst, ok := m.instances[instanceID]; ok {
		return inst, nil
	}
	return nil, fmt.Errorf("InvalidInstanceID.NotFound: The instance ID '%s' does not exist", instanceID)
}

func (m *MockEC2Client) StartInstances(ctx context.Context, instanceID string) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: AWS API StartInstances unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if inst, ok := m.instances[instanceID]; ok {
		inst.State = "running"
		return nil
	}
	return fmt.Errorf("InvalidInstanceID.NotFound: %s", instanceID)
}

func (m *MockEC2Client) StopInstances(ctx context.Context, instanceID string) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: AWS API StopInstances unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if inst, ok := m.instances[instanceID]; ok {
		inst.State = "stopped"
		return nil
	}
	return fmt.Errorf("InvalidInstanceID.NotFound: %s", instanceID)
}

func (m *MockEC2Client) TerminateInstances(ctx context.Context, instanceID string) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: AWS API TerminateInstances unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if inst, ok := m.instances[instanceID]; ok {
		inst.State = "terminated"
		return nil
	}
	return fmt.Errorf("InvalidInstanceID.NotFound: %s", instanceID)
}
