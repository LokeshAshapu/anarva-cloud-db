package service

import (
	"context"
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/dns"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/edge"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/health"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/routing"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/tls"

	activityStream "github.com/anarva-cloud/anarva-cloud-db/internal/activity"
)

type LoadBalancerService struct {
	repo         *repository.LoadBalancerRepository
	provider     provider.LoadBalancerProvider
	routingSvc   *routing.RoutingService
	healthSvc    *health.HealthService
	tlsSvc       *tls.TLSService
	dnsSvc       *dns.DNSIntegrationService
	ssrfSvc      *edge.SSRFValidationService
	actStream    *activityStream.Stream
}

func NewLoadBalancerService(
	repo *repository.LoadBalancerRepository,
	prov provider.LoadBalancerProvider,
	routingSvc *routing.RoutingService,
	healthSvc *health.HealthService,
	tlsSvc *tls.TLSService,
	dnsSvc *dns.DNSIntegrationService,
	ssrfSvc *edge.SSRFValidationService,
	actStream *activityStream.Stream,
) *LoadBalancerService {
	return &LoadBalancerService{
		repo:       repo,
		provider:   prov,
		routingSvc: routingSvc,
		healthSvc:  healthSvc,
		tlsSvc:     tlsSvc,
		dnsSvc:     dnsSvc,
		ssrfSvc:    ssrfSvc,
		actStream:  actStream,
	}
}

func (s *LoadBalancerService) CreateLoadBalancer(ctx context.Context, orgID, projectID, name string, lbType domain.LBType, scheme domain.LBScheme, networkID string, subnetIDs []string) (*domain.LoadBalancer, error) {
	lbID := fmt.Sprintf("lb-%d", time.Now().UnixNano())
	lb := &domain.LoadBalancer{
		ID:             lbID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Name:           name,
		Type:           lbType,
		Scheme:         scheme,
		NetworkID:      networkID,
		SubnetIDs:      subnetIDs,
		Status:         domain.LBStatusCreating,
		RealityLabel:   "LOCAL_LOAD_BALANCER (LIMITED_CAPABILITIES)",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	created, err := s.provider.CreateLoadBalancer(ctx, lb)
	if err != nil {
		return nil, err
	}

	_ = s.repo.SaveLoadBalancer(created)

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: orgID,
			ProjectID:      projectID,
			ResourceID:     name,
			ActorID:        "system@anarva.cloud",
			Action:         activityStream.EventAction("LOAD_BALANCER_CREATED"),
			Timestamp:      time.Now(),
		})
	}

	return created, nil
}

func (s *LoadBalancerService) GetLoadBalancer(ctx context.Context, id string) (*domain.LoadBalancer, error) {
	return s.repo.GetLoadBalancer(id)
}

func (s *LoadBalancerService) ListLoadBalancers(ctx context.Context, orgID, projectID string) ([]*domain.LoadBalancer, error) {
	return s.repo.ListLoadBalancers(orgID, projectID)
}

func (s *LoadBalancerService) DeleteLoadBalancer(ctx context.Context, id string) error {
	lb, err := s.repo.GetLoadBalancer(id)
	if err != nil {
		return err
	}

	if err := s.provider.DeleteLoadBalancer(ctx, id); err != nil {
		return err
	}

	lb.Status = domain.LBStatusDeleted
	_ = s.repo.SaveLoadBalancer(lb)

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: lb.OrganizationID,
			ProjectID:      lb.ProjectID,
			ResourceID:     lb.Name,
			ActorID:        "system@anarva.cloud",
			Action:         activityStream.EventAction("LOAD_BALANCER_DELETED"),
			Timestamp:      time.Now(),
		})
	}
	return nil
}

func (s *LoadBalancerService) DeployApplication(ctx context.Context, orgID, projectID, name, containerImage, networkID, domainName string, acuCount, port int) (*domain.Application, error) {
	appID := fmt.Sprintf("app-%d", time.Now().UnixNano())

	// Create LB for app
	lb, err := s.CreateLoadBalancer(ctx, orgID, projectID, fmt.Sprintf("lb-%s", name), domain.LBTypeApplication, domain.LBSchemePublic, networkID, []string{"sub-101"})
	if err != nil {
		return nil, err
	}

	app := &domain.Application{
		ID:                  appID,
		OrganizationID:      orgID,
		ProjectID:           projectID,
		Name:                name,
		Status:              domain.AppRunning,
		NetworkID:           networkID,
		DeploymentReference: fmt.Sprintf("deploy-ref-%s", appID),
		LoadBalancerID:      lb.ID,
		DomainReference:     domainName,
		ContainerImage:      containerImage,
		ACUCount:            acuCount,
		Health:              "HEALTHY",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	_ = s.repo.SaveApplication(app)

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: orgID,
			ProjectID:      projectID,
			ResourceID:     name,
			ActorID:        "system@anarva.cloud",
			Action:         activityStream.EventAction("APPLICATION_DEPLOYED"),
			Timestamp:      time.Now(),
		})
	}

	return app, nil
}

func (s *LoadBalancerService) ListApplications(ctx context.Context, orgID, projectID string) ([]*domain.Application, error) {
	return s.repo.ListApplications(orgID, projectID)
}

func (s *LoadBalancerService) RequestCertificate(orgID, projectID, domainName string) (*domain.Certificate, error) {
	cert := s.tlsSvc.RequestCertificate(orgID, projectID, domainName)
	_ = s.repo.SaveCertificate(cert)
	return cert, nil
}

func (s *LoadBalancerService) CreateDomain(orgID, projectID, domainName string) (*domain.Domain, error) {
	dom := &domain.Domain{
		ID:                 fmt.Sprintf("dom-%d", time.Now().UnixNano()),
		OrganizationID:     orgID,
		ProjectID:          projectID,
		Name:               domainName,
		VerificationStatus: domain.DomainVerified,
		DNSZoneID:          "zone-default-01",
		CertificateStatus:  domain.CertActive,
		VerificationTXT:    fmt.Sprintf("anarva-domain-verify=%s", domainName),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	_ = s.repo.SaveDomain(dom)
	return dom, nil
}
