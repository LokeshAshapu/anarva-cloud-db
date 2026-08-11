package usecase

import (
	"context"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
)

func TestComputeUseCase_ACUValidation(t *testing.T) {
	prov := provider.NewLocalDockerComputeProvider()
	uc := NewComputeUseCase(nil, nil, prov)

	ctx := context.Background()

	// Invalid ACU test
	_, err := uc.CreateInstance(ctx, &domain.ComputeInstance{
		Name: "bad-acu-inst",
		ACU:  3.5, // 3.5 is not a valid standard ACU tier
	})
	if err == nil {
		t.Fatalf("expected error for invalid ACU 3.5, got nil")
	}

	// Valid ACU test
	inst, err := uc.CreateInstance(ctx, &domain.ComputeInstance{
		Name:     "valid-acu-inst",
		ACU:      2.0,
		RegionID: "us-east-1",
	})
	if err != nil {
		t.Fatalf("expected successful creation for 2.0 ACU, got: %v", err)
	}

	if inst.VCPU != 2.0 {
		t.Errorf("expected vCPU 2.0, got %.1f", inst.VCPU)
	}
	if inst.MemoryMB != 4096 {
		t.Errorf("expected MemoryMB 4096, got %d", inst.MemoryMB)
	}
}

func TestComputeUseCase_LifecycleAndExecute(t *testing.T) {
	prov := provider.NewLocalDockerComputeProvider()
	uc := NewComputeUseCase(nil, nil, prov)
	ctx := context.Background()

	inst, err := uc.CreateInstance(ctx, &domain.ComputeInstance{
		Name:     "test-worker",
		ACU:      1.0,
		RegionID: "us-east-1",
	})
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	// Test Execute Command
	res, err := uc.ExecuteCommand(ctx, inst.ID, &domain.CommandExecutionRequest{
		Command: "uname -a",
	})
	if err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}

	// Test Stop Instance
	if err := uc.StopInstance(ctx, inst.ID); err != nil {
		t.Fatalf("failed to stop instance: %v", err)
	}

	fetched, _ := uc.GetInstance(ctx, inst.ID)
	if fetched.Status != domain.StatusStopped {
		t.Errorf("expected status STOPPED, got %s", fetched.Status)
	}

	// Test Command Execution on Stopped Instance (Must Fail)
	_, err = uc.ExecuteCommand(ctx, inst.ID, &domain.CommandExecutionRequest{
		Command: "ls",
	})
	if err == nil {
		t.Fatalf("expected error when executing command on stopped instance, got nil")
	}

	// Test Start Instance
	if err := uc.StartInstance(ctx, inst.ID); err != nil {
		t.Fatalf("failed to start instance: %v", err)
	}
}
