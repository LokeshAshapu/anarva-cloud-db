package drift

import (
	"context"
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
)

type DriftType string

const (
	DriftResourceMissing        DriftType = "RESOURCE_MISSING"
	DriftStatusMismatch         DriftType = "STATUS_MISMATCH"
	DriftConfigurationMismatch DriftType = "CONFIGURATION_MISMATCH"
)

type DriftRecord struct {
	ID               string    `json:"id"`
	AnarvaResourceID string    `json:"anarvaResourceId"`
	Provider         string    `json:"provider"`
	DesiredState     string    `json:"desiredState"`
	ObservedState    string    `json:"observedState"`
	DriftType        DriftType `json:"driftType"`
	Severity         string    `json:"severity"`
	DetectedAt       time.Time `json:"detectedAt"`
	Status           string    `json:"status"`
}

type DriftEngine struct {
	mappingRepo mapping.MappingRepository
}

func NewDriftEngine(mappingRepo mapping.MappingRepository) *DriftEngine {
	return &DriftEngine{mappingRepo: mappingRepo}
}

func (e *DriftEngine) DetectDrift(ctx context.Context, anarvaResourceID, desiredState, observedState string) (*DriftRecord, error) {
	m, err := e.mappingRepo.GetMapping(anarvaResourceID)
	if err != nil {
		return &DriftRecord{
			ID:               fmt.Sprintf("drift-%d", time.Now().UnixNano()),
			AnarvaResourceID: anarvaResourceID,
			Provider:         "UNKNOWN",
			DesiredState:     desiredState,
			ObservedState:    "MISSING",
			DriftType:        DriftResourceMissing,
			Severity:         "HIGH",
			DetectedAt:       time.Now(),
			Status:           "ACTIVE",
		}, nil
	}

	if desiredState != observedState {
		return &DriftRecord{
			ID:               fmt.Sprintf("drift-%d", time.Now().UnixNano()),
			AnarvaResourceID: anarvaResourceID,
			Provider:         m.Provider,
			DesiredState:     desiredState,
			ObservedState:    observedState,
			DriftType:        DriftStatusMismatch,
			Severity:         "MEDIUM",
			DetectedAt:       time.Now(),
			Status:           "ACTIVE",
		}, nil
	}

	return nil, nil
}

func (e *DriftEngine) RepairDrift(ctx context.Context, driftID string) error {
	return nil
}

func (e *DriftEngine) InspectNetworkSecurityDrift(ctx context.Context, sgID string, isPublicIngress bool) (*DriftRecord, error) {
	if isPublicIngress {
		return &DriftRecord{
			ID:               fmt.Sprintf("drift-sec-%d", time.Now().UnixNano()/1e6),
			AnarvaResourceID: sgID,
			Provider:         "AWS",
			DesiredState:     "PRIVATE_INGRESS",
			ObservedState:    "PUBLIC_INGRESS_0.0.0.0/0",
			DriftType:        "SECURITY_DRIFT",
			Severity:         "HIGH",
			DetectedAt:       time.Now(),
			Status:           "ACTIVE",
		}, nil
	}
	return nil, nil
}
