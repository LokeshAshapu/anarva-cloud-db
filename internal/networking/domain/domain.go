package domain

import (
	"fmt"
	"net"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/arnv"
)

type NetworkStatus string

const (
	StatusCreating  NetworkStatus = "CREATING"
	StatusAvailable NetworkStatus = "AVAILABLE"
	StatusUpdating  NetworkStatus = "UPDATING"
	StatusDeleting  NetworkStatus = "DELETING"
	StatusDeleted   NetworkStatus = "DELETED"
	StatusFailed    NetworkStatus = "FAILED"
	StatusDegraded  NetworkStatus = "DEGRADED"
	StatusUnknown   NetworkStatus = "UNKNOWN"
)

type SubnetType string

const (
	SubnetPublic   SubnetType = "PUBLIC"
	SubnetPrivate  SubnetType = "PRIVATE"
	SubnetIsolated SubnetType = "ISOLATED"
)

type TargetType string

const (
	TargetLocal           TargetType = "LOCAL"
	TargetInternetGateway TargetType = "INTERNET_GATEWAY"
	TargetNatGateway      TargetType = "NAT_GATEWAY"
	TargetNetworkInterface TargetType = "NETWORK_INTERFACE"
	TargetPeering         TargetType = "PEERING"
)

type RuleDirection string

const (
	DirectionIngress RuleDirection = "INGRESS"
	DirectionEgress  RuleDirection = "EGRESS"
)

type RuleAction string

const (
	ActionAllow RuleAction = "ALLOW"
	ActionDeny  RuleAction = "DENY"
)

type IPVersion string

const (
	IPv4 IPVersion = "IPv4"
	IPv6 IPVersion = "IPv6"
)

type IPStatus string

const (
	IPStatusAvailable IPStatus = "AVAILABLE"
	IPStatusAllocated IPStatus = "ALLOCATED"
	IPStatusReserved  IPStatus = "RESERVED"
	IPStatusReleased  IPStatus = "RELEASED"
)

type VirtualNetwork struct {
	ID                  string        `json:"id"`
	OrganizationID      string        `json:"organizationId"`
	ProjectID           string        `json:"projectId"`
	Name                string        `json:"name"`
	Provider            string        `json:"provider"`
	RegionID            string        `json:"regionId"`
	CIDR                string        `json:"cidr"`
	Status              NetworkStatus `json:"status"`
	DNSEnabled          bool          `json:"dnsEnabled"`
	DefaultRouteTableID string        `json:"defaultRouteTableId"`
	RealityLabel        string        `json:"realityLabel"`
	CreatedAt           time.Time     `json:"createdAt"`
	UpdatedAt           time.Time     `json:"updatedAt"`
}

type Subnet struct {
	ID               string     `json:"id"`
	NetworkID        string     `json:"networkId"`
	Name             string     `json:"name"`
	CIDR             string     `json:"cidr"`
	AvailabilityZone string     `json:"availabilityZone"`
	Type             SubnetType `json:"type"`
	Status           string     `json:"status"`
	RouteTableID     string     `json:"routeTableId"`
	GatewayIP        string     `json:"gatewayIp"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type Route struct {
	ID           string     `json:"id"`
	RouteTableID string     `json:"routeTableId"`
	Destination  string     `json:"destination"`
	Target       string     `json:"target"`
	TargetType   TargetType `json:"targetType"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type RouteTable struct {
	ID        string    `json:"id"`
	NetworkID string    `json:"networkId"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Routes    []Route   `json:"routes"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SecurityRule struct {
	ID              string        `json:"id"`
	SecurityGroupID string        `json:"securityGroupId"`
	Direction       RuleDirection `json:"direction"`
	Protocol        string        `json:"protocol"` // TCP, UDP, ICMP, ALL
	FromPort        int           `json:"fromPort"`
	ToPort          int           `json:"toPort"`
	Source          string        `json:"source"`
	Destination     string        `json:"destination"`
	Action          RuleAction    `json:"action"`
	Priority        int           `json:"priority"`
	Description     string        `json:"description"`
}

type SecurityGroup struct {
	ID          string         `json:"id"`
	NetworkID   string         `json:"networkId"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Rules       []SecurityRule `json:"rules"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type NetworkInterface struct {
	ID                  string   `json:"id"`
	ResourceID          string   `json:"resourceId"`
	ResourceType        string   `json:"resourceType"` // COMPUTE, POSTGRES, LOAD_BALANCER
	NetworkID           string   `json:"networkId"`
	SubnetID            string   `json:"subnetId"`
	PrivateIP           string   `json:"privateIp"`
	PublicIPReference   string   `json:"publicIpReference,omitempty"`
	SecurityGroups      []string `json:"securityGroups"`
	Status              string   `json:"status"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type InternetGateway struct {
	ID                 string    `json:"id"`
	NetworkID          string    `json:"networkId"`
	Status             string    `json:"status"` // AVAILABLE, NOT_CONFIGURED, FAILED
	ProviderResourceId string    `json:"providerResourceId"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type NatGateway struct {
	ID                     string    `json:"id"`
	NetworkID              string    `json:"networkId"`
	SubnetID               string    `json:"subnetId"`
	Status                 string    `json:"status"` // AVAILABLE, NOT_CONFIGURED, SIMULATED
	PublicAddressReference string    `json:"publicAddressReference"`
	ProviderResourceId     string    `json:"providerResourceId"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type DNSZone struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	Name           string    `json:"name"`
	Type           string    `json:"type"` // PUBLIC, PRIVATE
	Provider       string    `json:"provider"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type DNSRecord struct {
	ID            string    `json:"id"`
	ZoneID        string    `json:"zoneId"`
	Name          string    `json:"name"`
	Type          string    `json:"type"` // A, AAAA, CNAME, TXT, MX, NS
	Value         string    `json:"value"`
	TTL           int       `json:"ttl"`
	RoutingPolicy string    `json:"routingPolicy,omitempty"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type NetworkPeering struct {
	ID                string    `json:"id"`
	SourceNetworkID   string    `json:"sourceNetworkId"`
	TargetNetworkID   string    `json:"targetNetworkId"`
	Status            string    `json:"status"` // PENDING, ACTIVE, REJECTED, FAILED
	ProviderReference string    `json:"providerReference"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type IPAllocation struct {
	ID        string    `json:"id"`
	NetworkID string    `json:"networkId"`
	SubnetID  string    `json:"subnetId"`
	IP        string    `json:"ip"`
	Version   IPVersion `json:"version"`
	Resource  string    `json:"resource"`
	Status    IPStatus  `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type ConnectivityTest struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Protocol    string    `json:"protocol"`
	Port        int       `json:"port"`
	Reachable   bool      `json:"reachable"`
	LatencyMs   float64   `json:"latencyMs"`
	Error       string    `json:"error,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

type NetworkPolicy struct {
	ID             string         `json:"id"`
	NetworkID      string         `json:"networkId"`
	Name           string         `json:"name"`
	IngressRules   []SecurityRule `json:"ingressRules"`
	EgressRules    []SecurityRule `json:"egressRules"`
	PortLimits     []int          `json:"portLimits"`
	AllowedCIDRs   []string       `json:"allowedCidrs"`
	CreatedAt      time.Time      `json:"createdAt"`
}

type NetworkMetrics struct {
	NetworkID      string    `json:"networkId"`
	BytesIn        int64     `json:"bytesIn"`
	BytesOut       int64     `json:"bytesOut"`
	PacketsIn      int64     `json:"packetsIn"`
	PacketsOut     int64     `json:"packetsOut"`
	Connections    int       `json:"connections"`
	LatencyMs      float64   `json:"latencyMs"`
	DroppedPackets int64     `json:"droppedPackets"`
	Quality        string    `json:"quality"` // ACTUAL, ESTIMATED, UNKNOWN
	Timestamp      time.Time `json:"timestamp"`
}

func ValidateCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR format '%s': %w", cidr, err)
	}
	return nil
}

func GenerateNetworkARNV(regionID, projectID, networkName string) string {
	return arnv.GenerateARNV("NETWORK", regionID, projectID, fmt.Sprintf("vpc/%s", networkName))
}
