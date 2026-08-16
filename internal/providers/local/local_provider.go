package local

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/providers"
)

// LocalReferenceProvider serves as the canonical reference implementation of the universal ANARVA provider contract.
type LocalReferenceProvider struct {
	mu           sync.RWMutex
	instances    map[string]*providers.ComputeInstanceDetails
	databases    map[string]*providers.DatabaseDetails
	buckets      map[string]*providers.StorageBucketDetails
	vpcs         map[string]*providers.VPCDetails
	isHealthy    bool
}

func NewLocalReferenceProvider() *LocalReferenceProvider {
	return &LocalReferenceProvider{
		instances: make(map[string]*providers.ComputeInstanceDetails),
		databases: make(map[string]*providers.DatabaseDetails),
		buckets:   make(map[string]*providers.StorageBucketDetails),
		vpcs:      make(map[string]*providers.VPCDetails),
		isHealthy: true,
	}
}

func (p *LocalReferenceProvider) ID() string {
	return "local"
}

func (p *LocalReferenceProvider) Name() string {
	return "Local Reference Provider"
}

func (p *LocalReferenceProvider) CheckHealth(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.isHealthy {
		return providers.ErrProviderUnavailable
	}
	return nil
}

// Compute Lifecycle
func (p *LocalReferenceProvider) LaunchInstance(ctx context.Context, opts providers.ComputeInstanceOpts) (*providers.ComputeInstanceDetails, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if opts.ID == "" {
		return nil, fmt.Errorf("%w: missing instance ID", providers.ErrProviderExecutionFailed)
	}
	if _, exists := p.instances[opts.ID]; exists {
		return nil, providers.ErrResourceAlreadyExists
	}

	details := &providers.ComputeInstanceDetails{
		ID:         opts.ID,
		Name:       opts.Name,
		Status:     "RUNNING",
		CurrentACU: opts.MinACU,
		PublicIP:   "127.0.0.1",
		PrivateIP:  "10.0.0.101",
		CreatedAt:  time.Now(),
	}
	p.instances[opts.ID] = details
	return details, nil
}

func (p *LocalReferenceProvider) StartInstance(ctx context.Context, instanceID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	inst, ok := p.instances[instanceID]
	if !ok {
		return providers.ErrResourceNotFound
	}
	inst.Status = "RUNNING"
	return nil
}

func (p *LocalReferenceProvider) StopInstance(ctx context.Context, instanceID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	inst, ok := p.instances[instanceID]
	if !ok {
		return providers.ErrResourceNotFound
	}
	inst.Status = "STOPPED"
	return nil
}

func (p *LocalReferenceProvider) TerminateInstance(ctx context.Context, instanceID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.instances[instanceID]; !ok {
		return providers.ErrResourceNotFound
	}
	delete(p.instances, instanceID)
	return nil
}

func (p *LocalReferenceProvider) ScaleComputeACU(ctx context.Context, instanceID string, targetACU float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	inst, ok := p.instances[instanceID]
	if !ok {
		return providers.ErrResourceNotFound
	}
	inst.CurrentACU = targetACU
	return nil
}

func (p *LocalReferenceProvider) GetInstanceMetrics(ctx context.Context, instanceID string) (map[string]float64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if _, ok := p.instances[instanceID]; !ok {
		return nil, providers.ErrResourceNotFound
	}
	return map[string]float64{
		"cpu_utilization":    15.5,
		"memory_utilization": 42.0,
	}, nil
}

// Database Lifecycle
func (p *LocalReferenceProvider) ProvisionDatabase(ctx context.Context, opts providers.DatabaseProvisionOpts) (*providers.DatabaseDetails, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.databases[opts.ID]; exists {
		return nil, providers.ErrResourceAlreadyExists
	}

	details := &providers.DatabaseDetails{
		ID:        opts.ID,
		Name:      opts.Name,
		Engine:    string(opts.Engine),
		Status:    "HEALTHY",
		Endpoint:  "localhost",
		Port:      5432,
		CreatedAt: time.Now(),
	}
	p.databases[opts.ID] = details
	return details, nil
}

func (p *LocalReferenceProvider) BackupDatabase(ctx context.Context, databaseID string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if _, ok := p.databases[databaseID]; !ok {
		return "", providers.ErrResourceNotFound
	}
	return fmt.Sprintf("backup-%s-%d", databaseID, time.Now().Unix()), nil
}

func (p *LocalReferenceProvider) RestoreDatabase(ctx context.Context, backupID string, targetDBID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.databases[targetDBID] = &providers.DatabaseDetails{
		ID:        targetDBID,
		Name:      "restored-db",
		Engine:    "postgres",
		Status:    "HEALTHY",
		Endpoint:  "localhost",
		Port:      5432,
		CreatedAt: time.Now(),
	}
	return nil
}

func (p *LocalReferenceProvider) DeleteDatabase(ctx context.Context, databaseID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.databases[databaseID]; !ok {
		return providers.ErrResourceNotFound
	}
	delete(p.databases, databaseID)
	return nil
}

// Storage Lifecycle
func (p *LocalReferenceProvider) CreateBucket(ctx context.Context, bucketName string, region string) (*providers.StorageBucketDetails, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.buckets[bucketName]; exists {
		return nil, providers.ErrResourceAlreadyExists
	}
	b := &providers.StorageBucketDetails{
		Name:      bucketName,
		Region:    region,
		Arn:       fmt.Sprintf("arnv:anarva:s3:::%s", bucketName),
		CreatedAt: time.Now(),
	}
	p.buckets[bucketName] = b
	return b, nil
}

func (p *LocalReferenceProvider) DeleteBucket(ctx context.Context, bucketName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.buckets[bucketName]; !ok {
		return providers.ErrResourceNotFound
	}
	delete(p.buckets, bucketName)
	return nil
}

func (p *LocalReferenceProvider) UploadObject(ctx context.Context, bucketName, objectKey string, data []byte) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if _, ok := p.buckets[bucketName]; !ok {
		return providers.ErrResourceNotFound
	}
	return nil
}

func (p *LocalReferenceProvider) DownloadObject(ctx context.Context, bucketName, objectKey string) ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if _, ok := p.buckets[bucketName]; !ok {
		return nil, providers.ErrResourceNotFound
	}
	return []byte("local_data_buffer"), nil
}

func (p *LocalReferenceProvider) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if _, ok := p.buckets[bucketName]; !ok {
		return providers.ErrResourceNotFound
	}
	return nil
}

// Network Lifecycle
func (p *LocalReferenceProvider) CreateVPC(ctx context.Context, opts providers.VPCOpts) (*providers.VPCDetails, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.vpcs[opts.ID]; exists {
		return nil, providers.ErrResourceAlreadyExists
	}
	vpc := &providers.VPCDetails{
		ID:        opts.ID,
		Name:      opts.Name,
		CIDRBlock: opts.CIDRBlock,
		Region:    opts.Region,
		Status:    "AVAILABLE",
		CreatedAt: time.Now(),
	}
	p.vpcs[opts.ID] = vpc
	return vpc, nil
}

func (p *LocalReferenceProvider) DeleteVPC(ctx context.Context, vpcID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.vpcs[vpcID]; !ok {
		return providers.ErrResourceNotFound
	}
	delete(p.vpcs, vpcID)
	return nil
}

func (p *LocalReferenceProvider) CreateSubnet(ctx context.Context, vpcID, cidr, zone string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if _, ok := p.vpcs[vpcID]; !ok {
		return "", providers.ErrResourceNotFound
	}
	return fmt.Sprintf("sub-%s-1", vpcID), nil
}

func (p *LocalReferenceProvider) CreateSecurityGroup(ctx context.Context, vpcID, groupName string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if _, ok := p.vpcs[vpcID]; !ok {
		return "", providers.ErrResourceNotFound
	}
	return fmt.Sprintf("sg-%s-1", vpcID), nil
}
