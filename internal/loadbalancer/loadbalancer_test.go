package loadbalancer_test

import (
	"context"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/dns"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/edge"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/health"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/reconciliation"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/routing"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/service"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/tls"
)

func TestLoadBalancer_LifecycleAndApplicationDeployment(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewLoadBalancerRepository()
	prov := provider.NewLocalDockerLoadBalancerProvider()
	routingSvc := routing.NewRoutingService()
	healthSvc := health.NewHealthService()
	tlsSvc := tls.NewTLSService()
	dnsSvc := dns.NewDNSIntegrationService()
	ssrfSvc := edge.NewSSRFValidationService()

	svc := service.NewLoadBalancerService(repo, prov, routingSvc, healthSvc, tlsSvc, dnsSvc, ssrfSvc, nil)

	// 1. Create Load Balancer
	lb, err := svc.CreateLoadBalancer(ctx, "org-test", "proj-test", "prod-alb", domain.LBTypeApplication, domain.LBSchemePublic, "vpc-01", []string{"sub-01"})
	if err != nil {
		t.Fatalf("failed to create load balancer: %v", err)
	}
	if lb.Status != domain.LBStatusActive {
		t.Errorf("expected status ACTIVE, got %s", lb.Status)
	}

	// 2. Deploy Application Workflow
	app, err := svc.DeployApplication(ctx, "org-test", "proj-test", "payment-api", "anarva/payment:v1", "vpc-01", "api.anarva.cloud", 2, 8080)
	if err != nil {
		t.Fatalf("failed to deploy application: %v", err)
	}
	if app.Status != domain.AppRunning {
		t.Errorf("expected application status RUNNING, got %s", app.Status)
	}

	// 3. Delete Load Balancer
	if err := svc.DeleteLoadBalancer(ctx, lb.ID); err != nil {
		t.Errorf("failed to delete load balancer: %v", err)
	}
}

func TestLoadBalancer_RoutingHostAndPathMatching(t *testing.T) {
	routingSvc := routing.NewRoutingService()

	rules := []domain.RoutingRule{
		{ID: "rule-1", ListenerID: "list-1", Priority: 10, Host: "api.example.com", Path: "/v1/*", Action: domain.ActionForward},
		{ID: "rule-2", ListenerID: "list-1", Priority: 20, Host: "app.example.com", Path: "/dashboard", Action: domain.ActionForward},
	}

	// Match Host & Path
	matched, err := routingSvc.MatchRoute("api.example.com", "/v1/users", rules)
	if err != nil || matched.ID != "rule-1" {
		t.Errorf("expected rule-1 match, got: %v, err: %v", matched, err)
	}

	// Priority Conflict Validation
	err = routingSvc.ValidateRule(&domain.RoutingRule{ID: "rule-3", ListenerID: "list-1", Priority: 10}, rules)
	if err == nil {
		t.Errorf("expected priority conflict error, got nil")
	}
}

func TestLoadBalancer_SSRFProtection(t *testing.T) {
	ssrfSvc := edge.NewSSRFValidationService()

	// Metadata Endpoint (Blocked)
	err := ssrfSvc.ValidateOrigin(&domain.Origin{HostnameReference: "169.254.169.254"})
	if err == nil || !testingContains(err.Error(), "SSRF BLOCKED") {
		t.Errorf("expected SSRF BLOCKED for 169.254.169.254, got: %v", err)
	}

	// Loopback Endpoint (Blocked)
	err = ssrfSvc.ValidateOrigin(&domain.Origin{HostnameReference: "127.0.0.1"})
	if err == nil || !testingContains(err.Error(), "SSRF BLOCKED") {
		t.Errorf("expected SSRF BLOCKED for 127.0.0.1, got: %v", err)
	}

	// Valid Workload Origin (Allowed)
	err = ssrfSvc.ValidateOrigin(&domain.Origin{HostnameReference: "10.0.1.50"})
	if err != nil {
		t.Errorf("expected valid origin to pass, got: %v", err)
	}
}

func TestLoadBalancer_ReconciliationDrift(t *testing.T) {
	ctx := context.Background()
	prov := provider.NewLocalDockerLoadBalancerProvider()
	recSvc := reconciliation.NewReconciliationService(prov)

	desired := &domain.LoadBalancer{
		ID:     "lb-non-existent",
		Scheme: domain.LBSchemePublic,
	}

	res, err := recSvc.Reconcile(ctx, desired)
	if err != nil {
		t.Fatalf("reconciliation error: %v", err)
	}
	if !res.DriftDetected {
		t.Errorf("expected drift detected for non-existent load balancer")
	}
}

func testingContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && stringSearch(s, substr)))
}

func stringSearch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
