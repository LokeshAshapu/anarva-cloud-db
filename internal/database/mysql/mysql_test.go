package mysql_test

import (
	"context"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/reconciliation"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/service"
)

func TestMySQL_InstanceLifecycleAndHealth(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMySQLRepository()
	prov := provider.NewLocalDockerMySQLProvider()
	svc := service.NewMySQLService(repo, prov, nil)

	// 1. Create MySQL Instance
	inst, err := svc.CreateInstance(ctx, "org-test", "proj-test", "primary-mysql", "8.0", "ap-hyderabad-1", "vpc-01", 2, 20)
	if err != nil {
		t.Fatalf("failed to create MySQL instance: %v", err)
	}
	if inst.Status != domain.StatusAvailable {
		t.Errorf("expected status AVAILABLE, got %s", inst.Status)
	}
	if inst.Port != 3306 {
		t.Errorf("expected MySQL default port 3306, got %d", inst.Port)
	}

	// 2. Health & Connection Info
	hlth, err := svc.GetHealth(ctx, inst.ID)
	if err != nil || hlth.Status != "HEALTHY" {
		t.Errorf("expected HEALTHY status, got %v, err: %v", hlth, err)
	}

	conn, err := svc.GetConnectionInfo(ctx, inst.ID)
	if err != nil || conn["port"] != 3306 {
		t.Errorf("expected port 3306 in connection info, got: %v", conn)
	}

	// 3. Delete Instance
	if err := svc.DeleteInstance(ctx, inst.ID); err != nil {
		t.Errorf("failed to delete instance: %v", err)
	}
}

func TestMySQL_SQLServiceQueryExecution(t *testing.T) {
	ctx := context.Background()
	sqlSvc := service.NewSQLService()

	// Safe Query
	res, err := sqlSvc.ExecuteQuery(ctx, "SELECT * FROM users;")
	if err != nil || res.RowCount == 0 {
		t.Errorf("expected successful query execution, got err: %v", err)
	}

	// Dangerous Query (Blocked)
	_, err = sqlSvc.ExecuteQuery(ctx, "DROP DATABASE production_db;")
	if err == nil {
		t.Errorf("expected SECURITY RISK error for DROP DATABASE, got nil")
	}
}

func TestMySQL_ReconciliationDrift(t *testing.T) {
	ctx := context.Background()
	prov := provider.NewLocalDockerMySQLProvider()
	recSvc := reconciliation.NewReconciliationService(prov)

	desired := &domain.MySQLInstance{
		ID:      "mysql-ghost",
		Version: "8.0",
	}

	res, err := recSvc.Reconcile(ctx, desired)
	if err != nil {
		t.Fatalf("reconciliation error: %v", err)
	}

	if !res.DriftDetected {
		t.Errorf("expected drift detected for missing instance")
	}
}
