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
	ID                  string        `json:"id" gorm:"primaryKey"`
	OrganizationID      string        `json:"organizationId" gorm:"index:idx_vpc_tenant"`
	ProjectID           string        `json:"projectId" gorm:"index:idx_vpc_tenant"`
	Name                string        `json:"name"`
	Provider            string        `json:"provider"`
	ProviderResourceID  string        `json:"providerResourceId,omitempty"`
	RegionID            string        `json:"regionId" gorm:"index:idx_vpc_region"`
	CIDR                string        `json:"cidr"`
	Status              NetworkStatus `json:"status" gorm:"index:idx_vpc_status"`
	DNSEnabled          bool          `json:"dnsEnabled"`
	DefaultRouteTableID string        `json:"defaultRouteTableId,omitempty"`
	RealityLabel        string        `json:"realityLabel"`
	CreatedAt           time.Time     `json:"createdAt"`
	UpdatedAt           time.Time     `json:"updatedAt"`
}

type Subnet struct {
	ID                 string     `json:"id" gorm:"primaryKey"`
	NetworkID          string     `json:"networkId" gorm:"index:idx_subnet_vpc"`
	OrganizationID     string     `json:"organizationId" gorm:"index:idx_subnet_tenant"`
	ProjectID          string     `json:"projectId" gorm:"index:idx_subnet_tenant"`
	Name               string     `json:"name"`
	CIDR               string     `json:"cidr"`
	AvailabilityZone   string     `json:"availabilityZone"`
	Type               SubnetType `json:"type"`
	Status             string     `json:"status"`
	RouteTableID       string     `json:"routeTableId,omitempty"`
	GatewayIP          string     `json:"gatewayIp,omitempty"`
	ProviderResourceID string     `json:"providerResourceId,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type Route struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	RouteTableID string     `json:"routeTableId" gorm:"index:idx_route_table"`
	Destination  string     `json:"destination"`
	Target       string     `json:"target"`
	TargetType   TargetType `json:"targetType"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type RouteTable struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	NetworkID          string    `json:"networkId" gorm:"index:idx_rt_vpc"`
	OrganizationID     string    `json:"organizationId" gorm:"index:idx_rt_tenant"`
	ProjectID          string    `json:"projectId" gorm:"index:idx_rt_tenant"`
	Name               string    `json:"name"`
	Status             string    `json:"status"`
	ProviderResourceID string    `json:"providerResourceId,omitempty"`
	Routes             []Route   `json:"routes" gorm:"-"`
	RoutesJSON         string    `json:"-" gorm:"type:text"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type SecurityRule struct {
	ID              string        `json:"id" gorm:"primaryKey"`
	SecurityGroupID string        `json:"securityGroupId" gorm:"index:idx_rule_sg"`
	Direction       RuleDirection `json:"direction"`
	Protocol        string        `json:"protocol"` // TCP, UDP, ICMP, ALL
	FromPort        int           `json:"fromPort"`
	ToPort          int           `json:"toPort"`
	CIDR            string        `json:"cidr,omitempty"`
	Source          string        `json:"source"`
	Destination     string        `json:"destination"`
	Action          RuleAction    `json:"action"`
	Priority        int           `json:"priority"`
	Description     string        `json:"description"`
}

type SecurityGroup struct {
	ID                 string         `json:"id" gorm:"primaryKey"`
	NetworkID          string         `json:"networkId" gorm:"index:idx_sg_vpc"`
	OrganizationID     string         `json:"organizationId" gorm:"index:idx_sg_tenant"`
	ProjectID          string         `json:"projectId" gorm:"index:idx_sg_tenant"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	Status             string         `json:"status"`
	ProviderResourceID string         `json:"providerResourceId,omitempty"`
	Rules              []SecurityRule `json:"rules" gorm:"-"`
	RulesJSON          string         `json:"-" gorm:"type:text"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

type NetworkInterface struct {
	ID                  string    `json:"id" gorm:"primaryKey"`
	ResourceID          string    `json:"resourceId" gorm:"index:idx_nic_res"`
	ResourceType        string    `json:"resourceType"` // COMPUTE, POSTGRES, LOAD_BALANCER
	NetworkID           string    `json:"networkId" gorm:"index:idx_nic_vpc"`
	SubnetID            string    `json:"subnetId" gorm:"index:idx_nic_subnet"`
	OrganizationID      string    `json:"organizationId" gorm:"index:idx_nic_tenant"`
	ProjectID           string    `json:"projectId" gorm:"index:idx_nic_tenant"`
	PrivateIP           string    `json:"privateIp"`
	PublicIPReference   string    `json:"publicIpReference,omitempty"`
	SecurityGroups      []string  `json:"securityGroups" gorm:"-"`
	SecurityGroupsJSON  string    `json:"-" gorm:"type:text"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type InternetGateway struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	NetworkID          string    `json:"networkId" gorm:"index:idx_igw_vpc"`
	Status             string    `json:"status"` // AVAILABLE, NOT_CONFIGURED, FAILED
	ProviderResourceId string    `json:"providerResourceId"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type NatGateway struct {
	ID                     string    `json:"id" gorm:"primaryKey"`
	NetworkID              string    `json:"networkId" gorm:"index:idx_nat_vpc"`
	SubnetID               string    `json:"subnetId" gorm:"index:idx_nat_subnet"`
	Status                 string    `json:"status"` // AVAILABLE, NOT_CONFIGURED, SIMULATED
	PublicAddressReference string    `json:"publicAddressReference"`
	ProviderResourceId     string    `json:"providerResourceId"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type DNSZone struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	OrganizationID string    `json:"organizationId" gorm:"index:idx_dns_tenant"`
	ProjectID      string    `json:"projectId" gorm:"index:idx_dns_tenant"`
	Name           string    `json:"name"`
	Type           string    `json:"type"` // PUBLIC, PRIVATE
	Provider       string    `json:"provider"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type DNSRecord struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	ZoneID        string    `json:"zoneId" gorm:"index:idx_rec_zone"`
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
	ID                string    `json:"id" gorm:"primaryKey"`
	SourceNetworkID   string    `json:"sourceNetworkId" gorm:"index:idx_peer_src"`
	TargetNetworkID   string    `json:"targetNetworkId" gorm:"index:idx_peer_tgt"`
	Status            string    `json:"status"` // PENDING, ACTIVE, REJECTED, FAILED
	ProviderReference string    `json:"providerReference"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type IPAllocation struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	NetworkID      string    `json:"networkId" gorm:"index:idx_ip_vpc"`
	SubnetID       string    `json:"subnetId" gorm:"index:idx_ip_subnet"`
	OrganizationID string    `json:"organizationId" gorm:"index:idx_ip_tenant"`
	ProjectID      string    `json:"projectId" gorm:"index:idx_ip_tenant"`
	IP             string    `json:"ip"`
	Version        IPVersion `json:"version"`
	Resource       string    `json:"resource"`
	Status         IPStatus  `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ConnectivityTest struct {
	ID          string    `json:"id" gorm:"primaryKey"`
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
	ID           string         `json:"id" gorm:"primaryKey"`
	NetworkID    string         `json:"networkId" gorm:"index:idx_pol_vpc"`
	Name         string         `json:"name"`
	IngressRules []SecurityRule `json:"ingressRules" gorm:"-"`
	EgressRules  []SecurityRule `json:"egressRules" gorm:"-"`
	PortLimits   []int          `json:"portLimits" gorm:"-"`
	AllowedCIDRs []string       `json:"allowedCidrs" gorm:"-"`
	CreatedAt    time.Time      `json:"createdAt"`
}

type NetworkMetrics struct {
	NetworkID      string    `json:"networkId" gorm:"primaryKey"`
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

func ValidateCIDRContainment(parentCIDR, childCIDR string) error {
	_, parentNet, err := net.ParseCIDR(parentCIDR)
	if err != nil {
		return fmt.Errorf("invalid VPC CIDR '%s': %w", parentCIDR, err)
	}

	_, childNet, err := net.ParseCIDR(childCIDR)
	if err != nil {
		return fmt.Errorf("invalid Subnet CIDR '%s': %w", childCIDR, err)
	}

	if !parentNet.Contains(childNet.IP) {
		return fmt.Errorf("SUBNET_CIDR_OUT_OF_BOUNDS: Subnet CIDR '%s' does not fit within VPC CIDR '%s'", childCIDR, parentCIDR)
	}

	lastIP := make(net.IP, len(childNet.IP))
	for i := range childNet.IP {
		lastIP[i] = childNet.IP[i] | ^childNet.Mask[i]
	}
	if !parentNet.Contains(lastIP) {
		return fmt.Errorf("SUBNET_CIDR_OUT_OF_BOUNDS: Subnet CIDR '%s' exceeds VPC CIDR '%s' boundaries", childCIDR, parentCIDR)
	}

	return nil
}

func GenerateNetworkARNV(regionID, projectID, networkName string) string {
	return arnv.GenerateARNV("NETWORK", regionID, projectID, fmt.Sprintf("vpc/%s", networkName))
}
