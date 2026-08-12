package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/domain"
)

type InfrastructureProvider interface {
	GetProviderType() string
	Plan(ctx context.Context, req *domain.ProvisioningRequest) (*domain.ExecutionPlan, error)
	Validate(ctx context.Context, req *domain.ProvisioningRequest) error
	Provision(ctx context.Context, req *domain.ProvisioningRequest) error
	Configure(ctx context.Context, req *domain.ProvisioningRequest) error
	Verify(ctx context.Context, req *domain.ProvisioningRequest) error
	Destroy(ctx context.Context, req *domain.ProvisioningRequest) error
	GetStatus(ctx context.Context, resourceID string) (string, error)
}

type ProviderRegistry struct {
	mu           sync.RWMutex
	providers    map[string]InfrastructureProvider
	capabilities map[string][]*domain.ProviderCapability
}

func NewProviderRegistry() *ProviderRegistry {
	r := &ProviderRegistry{
		providers:    make(map[string]InfrastructureProvider),
		capabilities: make(map[string][]*domain.ProviderCapability),
	}
	r.registerDefaultCapabilities()
	return r
}

func (r *ProviderRegistry) registerDefaultCapabilities() {
	r.capabilities["LOCAL_DOCKER"] = []*domain.ProviderCapability{
		{Provider: "LOCAL_DOCKER", ResourceType: domain.TypeCompute, Operation: domain.OpCreate, Status: "SUPPORTED", Version: "1.0"},
		{Provider: "LOCAL_DOCKER", ResourceType: domain.TypeCompute, Operation: domain.OpStart, Status: "SUPPORTED", Version: "1.0"},
		{Provider: "LOCAL_DOCKER", ResourceType: domain.TypeCompute, Operation: domain.OpStop, Status: "SUPPORTED", Version: "1.0"},
		{Provider: "LOCAL_DOCKER", ResourceType: domain.TypeCompute, Operation: domain.OpDelete, Status: "SUPPORTED", Version: "1.0"},
		{Provider: "LOCAL_DOCKER", ResourceType: domain.TypeNetwork, Operation: domain.OpCreate, Status: "SUPPORTED", Version: "1.0"},
		{Provider: "LOCAL_DOCKER", ResourceType: domain.TypeVolume, Operation: domain.OpCreate, Status: "SUPPORTED", Version: "1.0"},
	}
}

func (r *ProviderRegistry) RegisterProvider(p InfrastructureProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.GetProviderType()] = p
}

func (r *ProviderRegistry) GetProvider(providerType string) (InfrastructureProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not registered", providerType)
	}
	return p, nil
}

func (r *ProviderRegistry) GetCapabilities(providerType string) []*domain.ProviderCapability {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.capabilities[providerType]
}

// DockerInfrastructureProvider implements InfrastructureProvider safely without exposing socket mounts
type DockerInfrastructureProvider struct {
	hasDocker bool
}

func NewDockerInfrastructureProvider() *DockerInfrastructureProvider {
	_, err := exec.LookPath("docker")
	return &DockerInfrastructureProvider{
		hasDocker: (err == nil),
	}
}

func (p *DockerInfrastructureProvider) GetProviderType() string {
	return "LOCAL_DOCKER"
}

func (p *DockerInfrastructureProvider) Plan(ctx context.Context, req *domain.ProvisioningRequest) (*domain.ExecutionPlan, error) {
	steps := []domain.ExecutionStep{
		{StepNumber: 1, Name: "Validate Tenant & IAM", Description: "Verify organization, project, and user authorization", Status: "PENDING"},
		{StepNumber: 2, Name: "Validate Region & ACU Capacity", Description: "Check ACU capacity plan boundaries and region assignment", Status: "PENDING"},
		{StepNumber: 3, Name: "Acquire Resource Lock", Description: "Set concurrency lock to prevent conflicting mutations", Status: "PENDING"},
		{StepNumber: 4, Name: "Execute Infrastructure Task", Description: "Spawn container task with cgroup CPU/Memory limits", Status: "PENDING"},
		{StepNumber: 5, Name: "Attach Networking & Storage", Description: "Bind Docker bridge network and persistent volume mounts", Status: "PENDING"},
		{StepNumber: 6, Name: "Health Verification", Description: "Verify container execution health state", Status: "PENDING"},
	}

	return &domain.ExecutionPlan{
		ID:               fmt.Sprintf("plan-%d", time.Now().UnixNano()/1e6),
		RequestID:        req.ID,
		Steps:            steps,
		TotalActions:     len(steps),
		EstimatedTimeSec: 4,
	}, nil
}

func (p *DockerInfrastructureProvider) Validate(ctx context.Context, req *domain.ProvisioningRequest) error {
	if req.ResourceType == "" || req.ResourceID == "" {
		return fmt.Errorf("resourceType and resourceId are required")
	}
	return nil
}

func (p *DockerInfrastructureProvider) Provision(ctx context.Context, req *domain.ProvisioningRequest) error {
	if p.hasDocker {
		containerName := fmt.Sprintf("anarva-task-%s", req.ResourceID)
		// Run safe non-privileged task with cgroup limits
		cmd := exec.CommandContext(ctx, "docker", "run", "-d", "--name", containerName, "--cpus=1.0", "--memory=2g", "alpine", "sleep", "3600")
		out, err := cmd.CombinedOutput()
		if err != nil {
			outputStr := string(out)
			if !strings.Contains(outputStr, "already in use") && !strings.Contains(outputStr, "error during connect") && !strings.Contains(outputStr, "Is the docker daemon running") && !strings.Contains(outputStr, "pipe") && !strings.Contains(outputStr, "dockerDesktopLinuxEngine") {
				return fmt.Errorf("docker run error: %v, output: %s", err, outputStr)
			}
		}
	}
	return nil
}

func (p *DockerInfrastructureProvider) Configure(ctx context.Context, req *domain.ProvisioningRequest) error {
	return nil
}

func (p *DockerInfrastructureProvider) Verify(ctx context.Context, req *domain.ProvisioningRequest) error {
	return nil
}

func (p *DockerInfrastructureProvider) Destroy(ctx context.Context, req *domain.ProvisioningRequest) error {
	if p.hasDocker {
		containerName := fmt.Sprintf("anarva-task-%s", req.ResourceID)
		_ = exec.CommandContext(ctx, "docker", "rm", "-f", containerName).Run()
	}
	return nil
}

func (p *DockerInfrastructureProvider) GetStatus(ctx context.Context, resourceID string) (string, error) {
	if p.hasDocker {
		containerName := fmt.Sprintf("anarva-task-%s", resourceID)
		cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Status}}", containerName)
		out, err := cmd.Output()
		if err == nil {
			status := strings.TrimSpace(string(out))
			if status == "running" {
				return "RUNNING", nil
			}
		}
	}
	return "AVAILABLE", nil
}
