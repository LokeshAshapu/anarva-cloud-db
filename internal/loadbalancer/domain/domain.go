package domain

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/arnv"
)

type LBType string

const (
	LBTypeApplication LBType = "APPLICATION"
	LBTypeNetwork     LBType = "NETWORK"
)

type LBScheme string

const (
	LBSchemeInternal LBScheme = "INTERNAL"
	LBSchemePublic   LBScheme = "PUBLIC"
)

type LBStatus string

const (
	LBStatusCreating LBStatus = "CREATING"
	LBStatusActive   LBStatus = "ACTIVE"
	LBStatusUpdating LBStatus = "UPDATING"
	LBStatusDegraded LBStatus = "DEGRADED"
	LBStatusFailed   LBStatus = "FAILED"
	LBStatusDeleting LBStatus = "DELETING"
	LBStatusDeleted  LBStatus = "DELETED"
	LBStatusUnknown  LBStatus = "UNKNOWN"
)

type ListenerProtocol string

const (
	ProtocolHTTP  ListenerProtocol = "HTTP"
	ProtocolHTTPS ListenerProtocol = "HTTPS"
	ProtocolTCP   ListenerProtocol = "TCP"
	ProtocolTLS   ListenerProtocol = "TLS"
)

type PoolAlgorithm string

const (
	AlgoRoundRobin       PoolAlgorithm = "ROUND_ROBIN"
	AlgoLeastConnections PoolAlgorithm = "LEAST_CONNECTIONS"
	AlgoIPHash           PoolAlgorithm = "IP_HASH"
	AlgoWeighted         PoolAlgorithm = "WEIGHTED"
)

type TargetResourceType string

const (
	TargetContainer         TargetResourceType = "CONTAINER"
	TargetKubernetesService TargetResourceType = "KUBERNETES_SERVICE"
	TargetCompute           TargetResourceType = "COMPUTE"
)

type TargetStatus string

const (
	TargetHealthy   TargetStatus = "HEALTHY"
	TargetUnhealthy TargetStatus = "UNHEALTHY"
	TargetInitial   TargetStatus = "INITIAL"
	TargetDraining  TargetStatus = "DRAINING"
	TargetUnknown   TargetStatus = "UNKNOWN"
)

type RoutingAction string

const (
	ActionForward       RoutingAction = "FORWARD"
	ActionRedirect      RoutingAction = "REDIRECT"
	ActionFixedResponse RoutingAction = "FIXED_RESPONSE"
)

type CertStatus string

const (
	CertPending  CertStatus = "PENDING"
	CertActive   CertStatus = "ACTIVE"
	CertExpiring CertStatus = "EXPIRING"
	CertExpired  CertStatus = "EXPIRED"
	CertFailed   CertStatus = "FAILED"
	CertRevoked  CertStatus = "REVOKED"
	CertUnknown  CertStatus = "UNKNOWN"
)

type DomainStatus string

const (
	DomainPending  DomainStatus = "PENDING"
	DomainVerified DomainStatus = "VERIFIED"
	DomainFailed   DomainStatus = "FAILED"
)

type AppStatus string

const (
	AppCreating  AppStatus = "CREATING"
	AppDeploying AppStatus = "DEPLOYING"
	AppRunning   AppStatus = "RUNNING"
	AppDegraded  AppStatus = "DEGRADED"
	AppFailed    AppStatus = "FAILED"
	AppStopped   AppStatus = "STOPPED"
	AppDeleting  AppStatus = "DELETING"
	AppDeleted   AppStatus = "DELETED"
)

type WAFMode string

const (
	WAFModeDetection WAFMode = "DETECTION"
	WAFModeBlocking  WAFMode = "BLOCKING"
)

type CachePolicy string

const (
	CacheNoCache  CachePolicy = "NO_CACHE"
	CacheStatic   CachePolicy = "STATIC"
	CacheStandard CachePolicy = "STANDARD"
	CacheCustom   CachePolicy = "CUSTOM"
)

type OriginType string

const (
	OriginLoadBalancer OriginType = "LOAD_BALANCER"
	OriginContainer    OriginType = "CONTAINER"
	OriginK8sService   OriginType = "KUBERNETES_SERVICE"
	OriginObjectStorage OriginType = "OBJECT_STORAGE"
)

type LoadBalancer struct {
	ID                string    `json:"id"`
	OrganizationID    string    `json:"organizationId"`
	ProjectID         string    `json:"projectId"`
	Name              string    `json:"name"`
	Provider          string    `json:"provider"`
	Type              LBType    `json:"type"`
	Scheme            LBScheme  `json:"scheme"`
	NetworkID         string    `json:"networkId"`
	SubnetIDs         []string  `json:"subnetIds"`
	Status            LBStatus  `json:"status"`
	IPReference       string    `json:"ipReference"`
	HostnameReference string    `json:"hostnameReference"`
	RealityLabel      string    `json:"realityLabel"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type Listener struct {
	ID                   string           `json:"id"`
	LoadBalancerID       string           `json:"loadBalancerId"`
	Protocol             ListenerProtocol `json:"protocol"`
	Port                 int              `json:"port"`
	TLSEnabled           bool             `json:"tlsEnabled"`
	CertificateReference string           `json:"certificateReference,omitempty"`
	DefaultBackendPoolID string           `json:"defaultBackendPoolId,omitempty"`
	Status               string           `json:"status"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

type BackendPool struct {
	ID             string        `json:"id"`
	LoadBalancerID string        `json:"loadBalancerId"`
	Name           string        `json:"name"`
	Protocol       string        `json:"protocol"`
	Port           int           `json:"port"`
	Algorithm      PoolAlgorithm `json:"algorithm"`
	HealthCheckID  string        `json:"healthCheckId"`
	Status         string        `json:"status"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type BackendTarget struct {
	ID               string             `json:"id"`
	BackendPoolID    string             `json:"backendPoolId"`
	ResourceID       string             `json:"resourceId"`
	ResourceType     TargetResourceType `json:"resourceType"`
	AddressReference string             `json:"addressReference"`
	Port             int                `json:"port"`
	Weight           int                `json:"weight"`
	Status           TargetStatus       `json:"status"`
	CreatedAt        time.Time          `json:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt"`
}

type LoadBalancerHealthCheck struct {
	ID                 string    `json:"id"`
	BackendPoolID      string    `json:"backendPoolId"`
	Protocol           string    `json:"protocol"`
	Path               string    `json:"path"`
	Port               int       `json:"port"`
	IntervalSeconds    int       `json:"intervalSeconds"`
	TimeoutSeconds     int       `json:"timeoutSeconds"`
	HealthyThreshold   int       `json:"healthyThreshold"`
	UnhealthyThreshold int       `json:"unhealthyThreshold"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type RoutingRule struct {
	ID            string        `json:"id"`
	ListenerID    string        `json:"listenerId"`
	Priority      int           `json:"priority"`
	Host          string        `json:"host"`
	Path          string        `json:"path"`
	BackendPoolID string        `json:"backendPoolId"`
	Action        RoutingAction `json:"action"`
	Status        string        `json:"status"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

type Certificate struct {
	ID                   string     `json:"id"`
	OrganizationID       string     `json:"organizationId"`
	ProjectID            string     `json:"projectId"`
	Domain               string     `json:"domain"`
	Provider             string     `json:"provider"`
	Status               CertStatus `json:"status"`
	CertificateReference string     `json:"certificateReference"`
	Issuer               string     `json:"issuer"`
	ExpiresAt            time.Time  `json:"expiresAt"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type Domain struct {
	ID                 string       `json:"id"`
	OrganizationID     string       `json:"organizationId"`
	ProjectID          string       `json:"projectId"`
	Name               string       `json:"name"`
	VerificationStatus DomainStatus `json:"verificationStatus"`
	DNSZoneID          string       `json:"dnsZoneId"`
	CertificateStatus  CertStatus   `json:"certificateStatus"`
	VerificationTXT    string       `json:"verificationTxt"`
	CreatedAt          time.Time    `json:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
}

type Application struct {
	ID                  string    `json:"id"`
	OrganizationID      string    `json:"organizationId"`
	ProjectID           string    `json:"projectId"`
	Name                string    `json:"name"`
	Status              AppStatus `json:"status"`
	NetworkID           string    `json:"networkId"`
	DeploymentReference string    `json:"deploymentReference"`
	LoadBalancerID      string    `json:"loadBalancerId"`
	DomainReference     string    `json:"domainReference"`
	ContainerImage      string    `json:"containerImage"`
	ACUCount            int       `json:"acuCount"`
	Health              string    `json:"health"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type FailoverPolicy struct {
	ID              string `json:"id"`
	LoadBalancerID  string `json:"loadBalancerId"`
	PrimaryPool     string `json:"primaryPool"`
	SecondaryPool   string `json:"secondaryPool"`
	HealthThreshold int    `json:"healthThreshold"`
	Status          string `json:"status"`
}

type RateLimitPolicy struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Scope     string    `json:"scope"`  // IP, API_KEY, USER, PROJECT, ROUTE
	Requests  int       `json:"requests"`
	WindowSec int       `json:"windowSec"`
	Action    string    `json:"action"` // BLOCK, THROTTLE
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type WAFPolicy struct {
	ID             string    `json:"id"`
	LoadBalancerID string    `json:"loadBalancerId"`
	Rules          []string  `json:"rules"` // SQLi, XSS, PathTraversal
	Mode           WAFMode   `json:"mode"`  // DETECTION, BLOCKING
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type EdgeDistribution struct {
	ID             string      `json:"id"`
	OrganizationID string      `json:"organizationId"`
	ProjectID      string      `json:"projectId"`
	Domain         string      `json:"domain"`
	Origin         string      `json:"origin"`
	CachePolicy    CachePolicy `json:"cachePolicy"`
	Status         string      `json:"status"`
	Provider       string      `json:"provider"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

type CacheRule struct {
	ID             string   `json:"id"`
	DistributionID string   `json:"distributionId"`
	Path           string   `json:"path"`
	TTL            int      `json:"ttl"`
	CacheKey       string   `json:"cacheKey"`
	Methods        []string `json:"methods"`
	Status         string   `json:"status"`
}

type Origin struct {
	ID                string     `json:"id"`
	DistributionID    string     `json:"distributionId"`
	HostnameReference string     `json:"hostnameReference"`
	Port              int        `json:"port"`
	Protocol          string     `json:"protocol"`
	Path              string     `json:"path"`
	Type              OriginType `json:"type"`
	Status            string     `json:"status"`
}

func GenerateLBARNV(regionID, projectID, lbName string) string {
	return arnv.GenerateARNV("LOAD_BALANCER", regionID, projectID, fmt.Sprintf("lb/%s", lbName))
}
