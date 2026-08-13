package postgres_test

import (
	"context"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
)

func TestPostgresInstanceLifecycleAndHealth(t *testing.T) {
	ctx := context.Background()
	p := provider.NewLocalDockerPostgresProvider()
	s := service.NewPostgresService(p)

	// 1. Supported Versions
	versions, err := p.SupportedVersions(ctx)
	if err != nil || len(versions) == 0 {
		t.Fatalf("expected supported versions, got err: %v", err)
	}

	// 2. Create Instance
	inst, err := s.CreateInstance(ctx, "org-test", "proj-test", "prod-postgres-db", "17", "ap-hyderabad-1", "vpc-net-1", 2.0, 2048, 50, false)
	if err != nil {
		t.Fatalf("failed to create postgres instance: %v", err)
	}
	if inst.Status != domain.StatusAvailable {
		t.Errorf("expected StatusAvailable, got %s", inst.Status)
	}

	// 3. Health Probe
	health, err := s.GetHealth(ctx, inst.ID)
	if err != nil || !health.ConnectionAvailable {
		t.Errorf("failed to get health or connection not available: %v", err)
	}

	// 4. Test Connection
	connResult, err := s.TestConnection(ctx, inst.ID)
	if err != nil || connResult["reachable"] != true {
		t.Errorf("test connection failed: %v", err)
	}

	// 5. Scale Instance
	scaled, err := s.ScaleInstance(ctx, inst.ID, 4.0, 4096, 100)
	if err != nil || scaled.CPU != 4.0 {
		t.Errorf("failed to scale postgres instance: %v", err)
	}

	// 6. Delete Instance
	if err := s.DeleteInstance(ctx, inst.ID); err != nil {
		t.Errorf("failed to delete postgres instance: %v", err)
	}
}

func TestSQLService_QueryExecution(t *testing.T) {
	ctx := context.Background()
	sqlSvc := service.NewSQLService()

	// Safe SELECT
	res, err := sqlSvc.ExecuteQuery(ctx, "inst-101", "SELECT * FROM users LIMIT 5;")
	if err != nil || res.RowCount == 0 {
		t.Fatalf("expected rows from safe query, got err: %v", err)
	}

	// Dangerous Statement
	_, err = sqlSvc.ExecuteQuery(ctx, "inst-101", "DROP DATABASE production_db;")
	if err == nil {
		t.Errorf("expected error for dangerous SQL query, got nil")
	}
}
