package domain

import (
	"fmt"
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
	StatusUnknown   NetworkStatus = "UNKNOWN"
)

type SubnetType string

const (
	SubnetPublic   SubnetType = "PUBLIC"
	SubnetPrivate  SubnetType = "PRIVATE"
	SubnetInternal SubnetType = "INTERNAL"
)

type IPVersion string

const (
	IPv4 IPVersion = "IPv4"
	IPv6 IPVersion = "IPv6"
)

type IPType string

const (
	IPTypePrivate  IPType = "PRIVATE"
	IPTypePublic   IPType = "PUBLIC"
	IPTypeReserved IPType = "RESERVED"
)

type IPStatus string

const (
	IPStatusAvailable IPStatus = "AVAILABLE"
	IPStatusAllocated IPStatus = "ALLOCATED"
	IPStatusReserved  IPStatus = "RESERVED"
	IPStatusReleased  IPStatus = "RELEASED"
	IPStatusUnknown   IPStatus = "UNKNOWN"
)

type NextHopType string

const (
	NextHopLocal           NextHopType = "LOCAL"
	NextHopInternetGateway NextHopType = "INTERNET_GATEWAY"
	NextHopNatGateway      NextHopType = "NAT_GATEWAY"
	NextHopInstance        NextHopType = "INSTANCE"
	NextHopLoadBalancer    NextHopType = "LOAD_BALANCER"
	NextHopVPN             NextHopType = "VPN"
	NextHopPeering         NextHopType = "PEERING"
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

type LoadBalancerType string

const (
	LBTypeApplication LoadBalancerType = "APPLICATION"
	LBTypeNetwork     LoadBalancerType = "NETWORK"
)

type Network struct {
	ID                string        `json:"id"`
	ResourceID        string        `json:"resourceId"`
	OrganizationID    string        `json:"organizationId"`
	ProjectID         string        `json:"projectId"`
	Name              string        `json:"name"`
	Slug              string        `json:"slug"`
	Description       string        `json:"description"`
	RegionID          string        `json:"regionId"`
	CIDR              string        `json:"cidr"`
	IPv4CIDR          string        `json:"ipv4Cidr"`
	IPv6CIDR          string        `json:"ipv6Cidr,omitempty"`
	Status            NetworkStatus `json:"status"`
	Provider          string        `json:"provider"`
	ProviderNetworkID string        `json:"providerNetworkId,omitempty"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
	DeletedAt         *time.Time    `json:"deletedAt,omitempty"`
}

type Subnet struct {
	ID               string     `json:"id"`
	NetworkID        string     `json:"networkId"`
	OrganizationID   string     `json:"organizationId"`
	ProjectID        string     `json:"projectId"`
	Name             string     `json:"name"`
	CIDR             string     `json:"cidr"`
	Type             SubnetType `json:"type"`
	RegionID         string     `json:"regionId"`
	ZoneID           string     `json:"zoneId"`
	GatewayIP        string     `json:"gatewayIp"`
	Status           string     `json:"status"`
	ProviderSubnetID string     `json:"providerSubnetId,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"deletedAt,omitempty"`
}

type IPAddress struct {
	ID                string    `json:"id"`
	NetworkID         string    `json:"networkId"`
	SubnetID          string    `json:"subnetId"`
	ResourceID        string    `json:"resourceId,omitempty"`
	Address           string    `json:"address"`
	Version           IPVersion `json:"version"`
	Type              IPType    `json:"type"`
	Status            IPStatus  `json:"status"`
	ProviderReference string    `json:"providerReference,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type Route struct {
	ID              string      `json:"id"`
	RouteTableID    string      `json:"routeTableId"`
	DestinationCIDR string      `json:"destinationCidr"`
	NextHopType     NextHopType `json:"nextHopType"`
	NextHopID       string      `json:"nextHopId,omitempty"`
	Priority        int         `json:"priority"`
	Status          string      `json:"status"`
	CreatedAt       time.Time   `json:"createdAt"`
}

type RouteTable struct {
	ID                   string    `json:"id"`
	NetworkID            string    `json:"networkId"`
	OrganizationID       string    `json:"organizationId"`
	ProjectID            string    `json:"projectId"`
	Name                 string    `json:"name"`
	Routes               []Route   `json:"routes"`
	Status               string    `json:"status"`
	ProviderRouteTableID string    `json:"providerRouteTableId,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type InternetGateway struct {
	ID                string    `json:"id"`
	NetworkID         string    `json:"networkId"`
	Name              string    `json:"name"`
	Status            string    `json:"status"`
	Provider          string    `json:"provider"`
	ProviderReference string    `json:"providerReference,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type NatGateway struct {
	ID                string    `json:"id"`
	NetworkID         string    `json:"networkId"`
	SubnetID          string    `json:"subnetId"`
	PublicIPID        string    `json:"publicIpId"`
	Status            string    `json:"status"`
	ProviderReference string    `json:"providerReference,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type SecurityGroupRule struct {
	ID                    string        `json:"id"`
	SecurityGroupID       string        `json:"securityGroupId"`
	Direction             RuleDirection `json:"direction"`
	Protocol              string        `json:"protocol"` // TCP, UDP, ICMP, ALL
	FromPort              int           `json:"fromPort"`
	ToPort                int           `json:"toPort"`
	SourceCIDR            string        `json:"sourceCidr,omitempty"`
	DestinationCIDR       string        `json:"destinationCidr,omitempty"`
	SourceSecurityGroupID string        `json:"sourceSecurityGroupId,omitempty"`
	Action                RuleAction    `json:"action"`
	Priority              int           `json:"priority"`
	CreatedAt             time.Time     `json:"createdAt"`
}

type SecurityGroup struct {
	ID             string              `json:"id"`
	NetworkID      string              `json:"networkId"`
	OrganizationID string              `json:"organizationId"`
	ProjectID      string              `json:"projectId"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Rules          []SecurityGroupRule `json:"rules"`
	Status         string              `json:"status"`
	CreatedAt      time.Time           `json:"createdAt"`
	UpdatedAt      time.Time           `json:"updatedAt"`
}

type DNSZone struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	Name           string    `json:"name"`
	Type           string    `json:"type"` // PUBLIC, PRIVATE
	Status         string    `json:"status"`
	Provider       string    `json:"provider"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type DNSRecord struct {
	ID        string    `json:"id"`
	ZoneID    string    `json:"zoneId"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // A, AAAA, CNAME, TXT, MX, NS, SRV
	Value     string    `json:"value"`
	TTL       int       `json:"ttl"`
	Priority  int       `json:"priority,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Target struct {
	ID         string `json:"id"`
	TargetID   string `json:"targetId"` // Instance ID or Container IP
	Port       int    `json:"port"`
	Health     string `json:"health"` // HEALTHY, UNHEALTHY, UNKNOWN
	Registered time.Time `json:"registeredAt"`
}

type TargetGroup struct {
	ID                  string   `json:"id"`
	LoadBalancerID      string   `json:"loadBalancerId"`
	Name                string   `json:"name"`
	Protocol            string   `json:"protocol"` // HTTP, HTTPS, TCP
	Port                int      `json:"port"`
	HealthCheckProtocol string   `json:"healthCheckProtocol"`
	HealthCheckPort     int      `json:"healthCheckPort"`
	HealthCheckPath     string   `json:"healthCheckPath"`
	Targets             []Target `json:"targets"`
	Status              string   `json:"status"`
}

type Listener struct {
	ID             string `json:"id"`
	LoadBalancerID string `json:"loadBalancerId"`
	Protocol       string `json:"protocol"` // HTTP, HTTPS, TCP
	Port           int    `json:"port"`
	TargetPort     int    `json:"targetPort"`
	CertificateID  string `json:"certificateId,omitempty"`
	Status         string `json:"status"`
}

type LoadBalancer struct {
	ID                string           `json:"id"`
	OrganizationID    string           `json:"organizationId"`
	ProjectID         string           `json:"projectId"`
	NetworkID         string           `json:"networkId"`
	SubnetID          string           `json:"subnetId"`
	Name              string           `json:"name"`
	Type              LoadBalancerType `json:"type"`
	Protocol          string           `json:"protocol"`
	Status            string           `json:"status"`
	Provider          string           `json:"provider"`
	ProviderReference string           `json:"providerReference,omitempty"`
	DNSName           string           `json:"dnsName"`
	Listeners         []Listener       `json:"listeners"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
}

func GenerateNetworkARNV(regionID, projectID, networkName string) string {
	return arnv.GenerateARNV("NETWORK", regionID, projectID, fmt.Sprintf("vpc/%s", networkName))
}
