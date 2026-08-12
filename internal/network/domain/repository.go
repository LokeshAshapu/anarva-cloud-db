package domain

import "context"

type NetworkRepository interface {
	Create(ctx context.Context, net *Network) error
	GetByID(ctx context.Context, id string) (*Network, error)
	ListByProjectID(ctx context.Context, projectID string) ([]*Network, error)
	Update(ctx context.Context, net *Network) error
	Delete(ctx context.Context, id string) error
}

type SubnetRepository interface {
	CreateSubnet(ctx context.Context, sub *Subnet) error
	GetSubnetByID(ctx context.Context, id string) (*Subnet, error)
	ListSubnetsByNetworkID(ctx context.Context, networkID string) ([]*Subnet, error)
	DeleteSubnet(ctx context.Context, id string) error
}

type IPAMRepository interface {
	AllocateIP(ctx context.Context, ip *IPAddress) error
	GetIPByAddress(ctx context.Context, address string) (*IPAddress, error)
	ListIPsBySubnetID(ctx context.Context, subnetID string) ([]*IPAddress, error)
	ReleaseIP(ctx context.Context, id string) error
}

type SecurityGroupRepository interface {
	CreateSecurityGroup(ctx context.Context, sg *SecurityGroup) error
	GetSecurityGroupByID(ctx context.Context, id string) (*SecurityGroup, error)
	ListSecurityGroupsByNetworkID(ctx context.Context, networkID string) ([]*SecurityGroup, error)
	UpdateSecurityGroup(ctx context.Context, sg *SecurityGroup) error
	DeleteSecurityGroup(ctx context.Context, id string) error
}

type DNSRepository interface {
	CreateZone(ctx context.Context, z *DNSZone) error
	GetZoneByID(ctx context.Context, id string) (*DNSZone, error)
	ListZonesByProjectID(ctx context.Context, projectID string) ([]*DNSZone, error)
	CreateRecord(ctx context.Context, r *DNSRecord) error
	ListRecordsByZoneID(ctx context.Context, zoneID string) ([]*DNSRecord, error)
	DeleteRecord(ctx context.Context, id string) error
}

type LoadBalancerRepository interface {
	CreateLoadBalancer(ctx context.Context, lb *LoadBalancer) error
	GetLoadBalancerByID(ctx context.Context, id string) (*LoadBalancer, error)
	ListLoadBalancersByNetworkID(ctx context.Context, networkID string) ([]*LoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, id string) error
}
