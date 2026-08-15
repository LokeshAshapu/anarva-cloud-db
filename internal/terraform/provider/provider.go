package provider

import (
	"context"
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/terraform/client"
)

type DatabaseResourceState struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	ProjectID   string  `json:"projectId"`
	Engine      string  `json:"engine"`
	StorageGB   int     `json:"storageGb"`
	ACUUnits    float64 `json:"acuUnits"`
	MultiAZ     bool    `json:"multiAz"`
	Status      string  `json:"status"`
	PrimaryAZ   string  `json:"primaryAz"`
	SecondaryAZ string  `json:"secondaryAz"`
}

type ComputeResourceState struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	ProjectID string  `json:"projectId"`
	ACUUnits  float64 `json:"acuUnits"`
	RegionID  string  `json:"regionId"`
	Status    string  `json:"status"`
}

type StorageBucketResourceState struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ProjectID  string `json:"projectId"`
	RegionID   string `json:"regionId"`
	Encryption string `json:"encryption"`
}

type VPCResourceState struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ProjectID string `json:"projectId"`
	CIDR      string `json:"cidr"`
	RegionID  string `json:"regionId"`
	Status    string `json:"status"`
}

type Provider struct {
	client *client.Client
}

func NewProvider(cfg client.Config) *Provider {
	return &Provider{
		client: client.NewClient(cfg),
	}
}

// Database Resource CRUD Operations
func (p *Provider) CreateDatabase(ctx context.Context, plan *DatabaseResourceState) (*DatabaseResourceState, error) {
	if plan.Name == "" {
		return nil, fmt.Errorf("name is required for anarva_database")
	}
	if plan.ProjectID == "" {
		plan.ProjectID = p.client.Config.ProjectID
	}
	if plan.Engine == "" {
		plan.Engine = "POSTGRESQL"
	}

	payload := map[string]interface{}{
		"name":      plan.Name,
		"projectId": plan.ProjectID,
		"engine":    plan.Engine,
		"storageGb": plan.StorageGB,
		"acuUnits":  plan.ACUUnits,
		"multiAz":   plan.MultiAZ,
	}

	res, status, err := p.client.DoRequest(ctx, "POST", "/api/v1/databases", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create database [%d]: %w", status, err)
	}

	state := &DatabaseResourceState{
		ID:          fmt.Sprintf("res-rds-%s", plan.Name),
		Name:        plan.Name,
		ProjectID:   plan.ProjectID,
		Engine:      plan.Engine,
		StorageGB:   plan.StorageGB,
		ACUUnits:    plan.ACUUnits,
		MultiAZ:     plan.MultiAZ,
		Status:      "AVAILABLE",
		PrimaryAZ:   "ap-south-1a",
		SecondaryAZ: "ap-south-1b",
	}

	if dataMap, ok := res.Data.(map[string]interface{}); ok {
		if id, exists := dataMap["id"].(string); exists {
			state.ID = id
		}
		if st, exists := dataMap["status"].(string); exists {
			state.Status = st
		}
	}

	return state, nil
}

func (p *Provider) ReadDatabase(ctx context.Context, id string) (*DatabaseResourceState, error) {
	res, status, err := p.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/databases/%s", id), nil)
	if status == 404 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	state := &DatabaseResourceState{
		ID:          id,
		Name:        "anarva-rds-prod-01",
		Engine:      "POSTGRESQL",
		StorageGB:   20,
		ACUUnits:    2.0,
		MultiAZ:     true,
		Status:      "AVAILABLE",
		PrimaryAZ:   "ap-south-1a",
		SecondaryAZ: "ap-south-1b",
	}

	if dataMap, ok := res.Data.(map[string]interface{}); ok {
		if name, exists := dataMap["name"].(string); exists {
			state.Name = name
		}
		if st, exists := dataMap["status"].(string); exists {
			state.Status = st
		}
		if multiAZ, exists := dataMap["multiAz"].(bool); exists {
			state.MultiAZ = multiAZ
		}
	}
	return state, nil
}

func (p *Provider) DeleteDatabase(ctx context.Context, id string) error {
	_, status, err := p.client.DoRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/databases/%s", id), nil)
	if status == 404 {
		return nil
	}
	return err
}

// Compute Resource CRUD Operations
func (p *Provider) CreateCompute(ctx context.Context, plan *ComputeResourceState) (*ComputeResourceState, error) {
	if plan.Name == "" {
		return nil, fmt.Errorf("name is required for anarva_compute")
	}
	if plan.ProjectID == "" {
		plan.ProjectID = p.client.Config.ProjectID
	}

	payload := map[string]interface{}{
		"name":      plan.Name,
		"projectId": plan.ProjectID,
		"acuUnits":  plan.ACUUnits,
		"regionId":  plan.RegionID,
	}

	res, status, err := p.client.DoRequest(ctx, "POST", "/api/v1/compute/instances", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute [%d]: %w", status, err)
	}

	state := &ComputeResourceState{
		ID:        fmt.Sprintf("res-ec2-%s", plan.Name),
		Name:      plan.Name,
		ProjectID: plan.ProjectID,
		ACUUnits:  plan.ACUUnits,
		RegionID:  plan.RegionID,
		Status:    "RUNNING",
	}

	if dataMap, ok := res.Data.(map[string]interface{}); ok {
		if id, exists := dataMap["id"].(string); exists {
			state.ID = id
		}
	}

	return state, nil
}

func (p *Provider) ReadCompute(ctx context.Context, id string) (*ComputeResourceState, error) {
	res, status, err := p.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/compute/instances/%s", id), nil)
	if status == 404 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	state := &ComputeResourceState{
		ID:       id,
		Name:     "anarva-app-worker-01",
		ACUUnits: 1.0,
		RegionID: "ap-south-1",
		Status:   "RUNNING",
	}

	if dataMap, ok := res.Data.(map[string]interface{}); ok {
		if name, exists := dataMap["name"].(string); exists {
			state.Name = name
		}
	}
	return state, nil
}

func (p *Provider) DeleteCompute(ctx context.Context, id string) error {
	_, status, err := p.client.DoRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/compute/instances/%s", id), nil)
	if status == 404 {
		return nil
	}
	return err
}

// Storage Bucket Resource CRUD Operations
func (p *Provider) CreateStorageBucket(ctx context.Context, plan *StorageBucketResourceState) (*StorageBucketResourceState, error) {
	if plan.Name == "" {
		return nil, fmt.Errorf("name is required for anarva_storage_bucket")
	}
	if plan.ProjectID == "" {
		plan.ProjectID = p.client.Config.ProjectID
	}

	payload := map[string]interface{}{
		"name":      plan.Name,
		"projectId": plan.ProjectID,
		"regionId":  plan.RegionID,
	}

	res, status, err := p.client.DoRequest(ctx, "POST", "/api/v1/storage/buckets", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage bucket [%d]: %w", status, err)
	}

	state := &StorageBucketResourceState{
		ID:         fmt.Sprintf("res-s3-%s", plan.Name),
		Name:       plan.Name,
		ProjectID:  plan.ProjectID,
		RegionID:   plan.RegionID,
		Encryption: "SSE-S3",
	}

	if dataMap, ok := res.Data.(map[string]interface{}); ok {
		if id, exists := dataMap["id"].(string); exists {
			state.ID = id
		}
	}

	return state, nil
}

func (p *Provider) ReadStorageBucket(ctx context.Context, id string) (*StorageBucketResourceState, error) {
	res, status, err := p.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/storage/buckets/%s", id), nil)
	if status == 404 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	state := &StorageBucketResourceState{
		ID:         id,
		Name:       "anarva-production-media-assets",
		RegionID:   "ap-south-1",
		Encryption: "SSE-S3",
	}

	if dataMap, ok := res.Data.(map[string]interface{}); ok {
		if name, exists := dataMap["name"].(string); exists {
			state.Name = name
		}
	}
	return state, nil
}

func (p *Provider) DeleteStorageBucket(ctx context.Context, id string) error {
	_, status, err := p.client.DoRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/storage/buckets/%s", id), nil)
	if status == 404 {
		return nil
	}
	return err
}

// VPC Resource CRUD Operations
func (p *Provider) CreateVPC(ctx context.Context, plan *VPCResourceState) (*VPCResourceState, error) {
	if plan.Name == "" || plan.CIDR == "" {
		return nil, fmt.Errorf("name and cidr are required for anarva_vpc")
	}
	if plan.ProjectID == "" {
		plan.ProjectID = p.client.Config.ProjectID
	}
	if plan.RegionID == "" {
		plan.RegionID = "us-east-1"
	}

	payload := map[string]interface{}{
		"name":      plan.Name,
		"projectId": plan.ProjectID,
		"cidr":      plan.CIDR,
		"regionId":  plan.RegionID,
	}

	res, status, err := p.client.DoRequest(ctx, "POST", "/api/v1/networks", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create anarva_vpc [%d]: %w", status, err)
	}

	state := &VPCResourceState{
		ID:        fmt.Sprintf("vpc-%s", plan.Name),
		Name:      plan.Name,
		ProjectID: plan.ProjectID,
		CIDR:      plan.CIDR,
		RegionID:  plan.RegionID,
		Status:    "AVAILABLE",
	}

	if dataMap, ok := res.Data.(map[string]interface{}); ok {
		if id, exists := dataMap["id"].(string); exists {
			state.ID = id
		}
	}

	return state, nil
}

func (p *Provider) ReadVPC(ctx context.Context, id string) (*VPCResourceState, error) {
	res, status, err := p.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/networks/%s", id), nil)
	if status == 404 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	state := &VPCResourceState{
		ID:     id,
		Status: "AVAILABLE",
	}

	if dataMap, ok := res.Data.(map[string]interface{}); ok {
		if name, exists := dataMap["name"].(string); exists {
			state.Name = name
		}
		if cidr, exists := dataMap["cidr"].(string); exists {
			state.CIDR = cidr
		}
	}
	return state, nil
}

func (p *Provider) DeleteVPC(ctx context.Context, id string) error {
	_, status, err := p.client.DoRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/networks/%s", id), nil)
	if status == 404 {
		return nil
	}
	return err
}
