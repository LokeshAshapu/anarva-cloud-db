package domain

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/arnv"
)

type RegionStatus string

const (
	RegionStatusActive        RegionStatus = "ACTIVE"
	RegionStatusDegraded      RegionStatus = "DEGRADED"
	RegionStatusMaintenance   RegionStatus = "MAINTENANCE"
	RegionStatusUnavailable   RegionStatus = "UNAVAILABLE"
	RegionStatusNotConfigured RegionStatus = "NOT_CONFIGURED"
	RegionStatusUnknown       RegionStatus = "UNKNOWN"
)

type ZoneStatus string

const (
	ZoneStatusActive      ZoneStatus = "ACTIVE"
	ZoneStatusDegraded    ZoneStatus = "DEGRADED"
	ZoneStatusMaintenance ZoneStatus = "MAINTENANCE"
	ZoneStatusUnavailable ZoneStatus = "UNAVAILABLE"
	ZoneStatusUnknown     ZoneStatus = "UNKNOWN"
)

type PlacementStrategy string

const (
	StrategySingleZone  PlacementStrategy = "SINGLE_ZONE"
	StrategyMultiZone   PlacementStrategy = "MULTI_ZONE"
	StrategyRegional    PlacementStrategy = "REGIONAL"
	StrategyMultiRegion PlacementStrategy = "MULTI_REGION"
	StrategyActivePassive PlacementStrategy = "ACTIVE_PASSIVE"
	StrategyActiveActive PlacementStrategy = "ACTIVE_ACTIVE"
)

type HAStrategy string

const (
	HASingle        HAStrategy = "SINGLE"
	HAMultiZone     HAStrategy = "MULTI_ZONE"
	HAMultiRegion   HAStrategy = "MULTI_REGION"
	HAActivePassive HAStrategy = "ACTIVE_PASSIVE"
	HAActiveActive  HAStrategy = "ACTIVE_ACTIVE"
)

type FailoverMode string

const (
	FailoverManual    FailoverMode = "MANUAL"
	FailoverAutomatic FailoverMode = "AUTOMATIC"
)

type IncidentSeverity string

const (
	SeverityInfo     IncidentSeverity = "INFO"
	SeverityWarning  IncidentSeverity = "WARNING"
	SeverityError    IncidentSeverity = "ERROR"
	SeverityCritical IncidentSeverity = "CRITICAL"
)

type RoutingType string

const (
	RoutingFailover    RoutingType = "FAILOVER"
	RoutingWeighted    RoutingType = "WEIGHTED"
	RoutingLatency     RoutingType = "LATENCY"
	RoutingGeolocation RoutingType = "GEOLOCATION"
)

type Region struct {
	ID                 string       `json:"id"`
	Name               string       `json:"name"`
	Code               string       `json:"code"`
	Provider           string       `json:"provider"`
	Status             RegionStatus `json:"status"`
	LatitudeReference  float64      `json:"latitudeReference"`
	LongitudeReference float64      `json:"longitudeReference"`
	CountryCode        string       `json:"countryCode"`
	CapacityStatus     string       `json:"capacityStatus"`
	RealityLabel       string       `json:"realityLabel"`
	CreatedAt          time.Time    `json:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
}

type AvailabilityZone struct {
	ID                    string     `json:"id"`
	RegionID              string     `json:"regionId"`
	Name                  string     `json:"name"`
	ProviderZoneReference string     `json:"providerZoneReference"`
	Status                ZoneStatus `json:"status"`
	CapacityStatus        string     `json:"capacityStatus"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

type RegionCapacity struct {
	RegionID        string    `json:"regionId"`
	TotalACU        int       `json:"totalAcu"`
	AvailableACU    int       `json:"availableAcu"`
	AllocatedACU    int       `json:"allocatedAcu"`
	CPUCapacity     int       `json:"cpuCapacity"`
	MemoryCapacity  int64     `json:"memoryCapacity"`
	StorageCapacity int64     `json:"storageCapacity"`
	Status          string    `json:"status"`
	Timestamp       time.Time `json:"timestamp"`
}

type PlacementPolicy struct {
	ID                   string            `json:"id"`
	ProjectID            string            `json:"projectId"`
	RegionID             string            `json:"regionId"`
	ZoneStrategy         PlacementStrategy `json:"zoneStrategy"`
	AvailabilityStrategy HAStrategy        `json:"availabilityStrategy"`
	FailureStrategy      string            `json:"failureStrategy"`
	Status               string            `json:"status"`
	CreatedAt            time.Time         `json:"createdAt"`
}

type DataResidencyPolicy struct {
	AllowedRegions []string `json:"allowedRegions"`
	BlockedRegions []string `json:"blockedRegions"`
	Enforcement    string   `json:"enforcement"` // STRICT, AUDIT
	Status         string   `json:"status"`
}

type RegionHealth struct {
	RegionID       string    `json:"regionId"`
	Status         string    `json:"status"`
	LatencyMs      float64   `json:"latencyMs"`
	ResourceHealth string    `json:"resourceHealth"`
	NetworkHealth  string    `json:"networkHealth"`
	DatabaseHealth string    `json:"databaseHealth"`
	Quality        string    `json:"quality"` // ACTUAL, ESTIMATED, UNKNOWN
	Timestamp      time.Time `json:"timestamp"`
}

type GlobalHealth struct {
	Status         string    `json:"status"` // HEALTHY, DEGRADED, PARTIAL_OUTAGE, MAJOR_OUTAGE
	TotalRegions   int       `json:"totalRegions"`
	ActiveRegions  int       `json:"activeRegions"`
	DegradedCount  int       `json:"degradedCount"`
	OutageCount    int       `json:"outageCount"`
	Timestamp      time.Time `json:"timestamp"`
}

type HighAvailabilityPolicy struct {
	ID               string       `json:"id"`
	ResourceID       string       `json:"resourceId"`
	ResourceType     string       `json:"resourceType"`
	Strategy         HAStrategy   `json:"strategy"`
	PrimaryRegion    string       `json:"primaryRegion"`
	PrimaryZone      string       `json:"primaryZone"`
	SecondaryRegions []string     `json:"secondaryRegions"`
	SecondaryZones   []string     `json:"secondaryZones"`
	FailoverMode     FailoverMode `json:"failoverMode"`
	Status           string       `json:"status"`
	CreatedAt        time.Time    `json:"createdAt"`
}

type DatabaseReplication struct {
	ID               string    `json:"id"`
	SourceResourceID string    `json:"sourceResourceId"`
	TargetResourceID string    `json:"targetResourceId"`
	SourceRegion     string    `json:"sourceRegion"`
	TargetRegion     string    `json:"targetRegion"`
	Status           string    `json:"status"` // STREAMING, LAGGING, STOPPED, FAILED
	ReplicationLagMs float64   `json:"replicationLagMs"`
	ReplicationType  string    `json:"replicationType"`
	LastSyncedAt     time.Time `json:"lastSyncedAt"`
}

type FailoverPolicy struct {
	ID              string       `json:"id"`
	ResourceID      string       `json:"resourceId"`
	Primary         string       `json:"primary"`
	Secondary       string       `json:"secondary"`
	HealthThreshold int          `json:"healthThreshold"`
	CooldownSec     int          `json:"cooldownSec"`
	Mode            FailoverMode `json:"mode"`
	GenerationLock  int64        `json:"generationLock"`
	Status          string       `json:"status"`
	CreatedAt       time.Time    `json:"createdAt"`
}

type DisasterRecoveryPolicy struct {
	TargetRPOMinutes int       `json:"targetRpoMinutes"`
	TargetRTOMinutes int       `json:"targetRtoMinutes"`
	BackupPolicy     string    `json:"backupPolicy"`
	ReplicationPolicy string   `json:"replicationPolicy"`
	RegionPolicy     string    `json:"regionPolicy"`
	Status           string    `json:"status"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type RecoveryPlan struct {
	ID             string    `json:"id"`
	ResourceID     string    `json:"resourceId"`
	FailoverRegion string    `json:"failoverRegion"`
	Steps          []string  `json:"steps"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type RegionEvacuationPlan struct {
	ID              string    `json:"id"`
	EvacuatedRegion string    `json:"evacuatedRegion"`
	TargetRegion    string    `json:"targetRegion"`
	Status          string    `json:"status"`
	StepsCompleted  int       `json:"stepsCompleted"`
	TotalSteps      int       `json:"totalSteps"`
	CreatedAt       time.Time `json:"createdAt"`
}

type GlobalRoutingPolicy struct {
	Domain       string      `json:"domain"`
	RoutingType  RoutingType `json:"routingType"`
	Regions      []string    `json:"regions"`
	Weights      []int       `json:"weights"`
	HealthChecks bool        `json:"healthChecks"`
	Status       string      `json:"status"`
}

type DependencyEdge struct {
	ID          string `json:"id"`
	Source      string `json:"source"`
	Target      string `json:"target"`
	Relationship string `json:"relationship"` // dependsOn, requiredBy, replicatesTo, routesTo
}

type InfrastructureIncident struct {
	ID         string           `json:"id"`
	Severity   IncidentSeverity `json:"severity"`
	RegionID   string           `json:"regionId"`
	ZoneID     string           `json:"zoneId,omitempty"`
	ResourceID string           `json:"resourceId,omitempty"`
	Type       string           `json:"type"` // REGION_OUTAGE, ZONE_OUTAGE, DATABASE_FAILURE, NETWORK_FAILURE
	Status     string           `json:"status"`
	Summary    string           `json:"summary"`
	StartedAt  time.Time        `json:"startedAt"`
	ResolvedAt *time.Time       `json:"resolvedAt,omitempty"`
}

func GenerateRegionARNV(provider, regionCode string) string {
	return arnv.GenerateARNV("REGION", regionCode, "global", fmt.Sprintf("region/%s/%s", provider, regionCode))
}
