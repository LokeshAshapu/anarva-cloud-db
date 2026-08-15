package ipam

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
)

type IPAMService struct {
	mu          sync.RWMutex
	allocations map[string]*domain.IPAllocation
}

func NewIPAMService() *IPAMService {
	return &IPAMService{
		allocations: make(map[string]*domain.IPAllocation),
	}
}

func (s *IPAMService) ValidateCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR block '%s': %w", cidr, err)
	}
	return nil
}

func (s *IPAMService) CheckCIDROverlap(existingCIDRs []string, newCIDR string) error {
	_, newNet, err := net.ParseCIDR(newCIDR)
	if err != nil {
		return err
	}

	for _, existing := range existingCIDRs {
		_, existNet, err := net.ParseCIDR(existing)
		if err != nil {
			continue
		}

		if existNet.Contains(newNet.IP) || newNet.Contains(existNet.IP) {
			return fmt.Errorf("CIDR overlap detected: new CIDR '%s' overlaps with existing CIDR '%s'", newCIDR, existing)
		}
	}
	return nil
}

func (s *IPAMService) Allocate(networkID, subnetID, resource string, ipVersion domain.IPVersion) (*domain.IPAllocation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ipAddr := fmt.Sprintf("10.0.%d.%d", len(s.allocations)/254+1, len(s.allocations)%254+10)

	alloc := &domain.IPAllocation{
		ID:        fmt.Sprintf("ip-%d", time.Now().UnixNano()),
		NetworkID: networkID,
		SubnetID:  subnetID,
		IP:        ipAddr,
		Version:   ipVersion,
		Resource:  resource,
		Status:    domain.IPStatusAllocated,
		CreatedAt: time.Now(),
	}

	s.allocations[alloc.ID] = alloc
	return alloc, nil
}

func (s *IPAMService) Release(allocationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if alloc, ok := s.allocations[allocationID]; ok {
		alloc.Status = domain.IPStatusReleased
	}
	return nil
}

func (s *IPAMService) Reserve(networkID, subnetID, ip string) (*domain.IPAllocation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	alloc := &domain.IPAllocation{
		ID:        fmt.Sprintf("ip-res-%d", time.Now().UnixNano()),
		NetworkID: networkID,
		SubnetID:  subnetID,
		IP:        ip,
		Version:   domain.IPv4,
		Resource:  "SYSTEM_RESERVED",
		Status:    domain.IPStatusReserved,
		CreatedAt: time.Now(),
	}
	s.allocations[alloc.ID] = alloc
	return alloc, nil
}

func (s *IPAMService) List(subnetID string) []*domain.IPAllocation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var res []*domain.IPAllocation
	for _, a := range s.allocations {
		if subnetID == "" || a.SubnetID == subnetID {
			res = append(res, a)
		}
	}
	return res
}

func (s *IPAMService) CheckAvailability(subnetID, ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, a := range s.allocations {
		if a.SubnetID == subnetID && a.IP == ip && a.Status == domain.IPStatusAllocated {
			return false
		}
	}
	return true
}

func (s *IPAMService) GetIPAMAllocationSummary(vpcCIDR string, subnets []string) (totalIPs int, allocatedIPs int, availableIPs int) {
	_, parentNet, err := net.ParseCIDR(vpcCIDR)
	if err != nil {
		return 0, 0, 0
	}
	ones, bits := parentNet.Mask.Size()
	totalIPs = 1 << (bits - ones)

	for _, subCIDR := range subnets {
		_, subNet, subErr := net.ParseCIDR(subCIDR)
		if subErr == nil {
			subOnes, subBits := subNet.Mask.Size()
			allocatedIPs += 1 << (subBits - subOnes)
		}
	}

	availableIPs = totalIPs - allocatedIPs
	if availableIPs < 0 {
		availableIPs = 0
	}
	return totalIPs, allocatedIPs, availableIPs
}
