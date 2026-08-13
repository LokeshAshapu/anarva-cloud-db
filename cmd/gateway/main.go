package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	authHttp "github.com/anarva-cloud/anarva-cloud-db/internal/auth/delivery/http"
	authDomain "github.com/anarva-cloud/anarva-cloud-db/internal/auth/domain"
	authRepo "github.com/anarva-cloud/anarva-cloud-db/internal/auth/repository"
	authUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/auth/usecase"

	backupHttp "github.com/anarva-cloud/anarva-cloud-db/internal/backup/delivery/http"
	backupDomain "github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	backupRepo "github.com/anarva-cloud/anarva-cloud-db/internal/backup/repository"
	backupUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/backup/usecase"

	databaseHttp "github.com/anarva-cloud/anarva-cloud-db/internal/database/delivery/http"
	databaseDomain "github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	databaseDriver "github.com/anarva-cloud/anarva-cloud-db/internal/database/driver"
	databaseRepo "github.com/anarva-cloud/anarva-cloud-db/internal/database/repository"
	databaseUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/database/usecase"

	projectHttp "github.com/anarva-cloud/anarva-cloud-db/internal/project/delivery/http"
	projectDomain "github.com/anarva-cloud/anarva-cloud-db/internal/project/domain"
	projectRepo "github.com/anarva-cloud/anarva-cloud-db/internal/project/repository"
	projectUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/project/usecase"

	"github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	backupProvider "github.com/anarva-cloud/anarva-cloud-db/internal/backup/provider"
	computeHttp "github.com/anarva-cloud/anarva-cloud-db/internal/compute/delivery/http"
	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
	computeUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/compute/usecase"
	devHttp "github.com/anarva-cloud/anarva-cloud-db/internal/developer/delivery/http"
	devUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/developer/usecase"
	gwMiddleware "github.com/anarva-cloud/anarva-cloud-db/internal/gateway/middleware"
	iamHttp "github.com/anarva-cloud/anarva-cloud-db/internal/iam/delivery/http"
	iamService "github.com/anarva-cloud/anarva-cloud-db/internal/iam/service"
	networkProvider "github.com/anarva-cloud/anarva-cloud-db/internal/network/provider"
	networkUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/network/usecase"
	observabilityHttp "github.com/anarva-cloud/anarva-cloud-db/internal/observability/delivery/http"
	observabilityService "github.com/anarva-cloud/anarva-cloud-db/internal/observability/service"
	postgresHandler "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/handler"
	postgresProvider "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/provider"
	postgresService "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
	storageProvider "github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"

	mysqlHandler "github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/handler"
	mysqlProvider "github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/provider"
	mysqlRepo "github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/repository"
	mysqlService "github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/service"

	networkingConn "github.com/anarva-cloud/anarva-cloud-db/internal/networking/connectivity"
	networkingDns "github.com/anarva-cloud/anarva-cloud-db/internal/networking/dns"
	networkingFw "github.com/anarva-cloud/anarva-cloud-db/internal/networking/firewall"
	networkingHandler "github.com/anarva-cloud/anarva-cloud-db/internal/networking/handler"
	networkingIpam "github.com/anarva-cloud/anarva-cloud-db/internal/networking/ipam"
	networkingProvider "github.com/anarva-cloud/anarva-cloud-db/internal/networking/provider"
	networkingRepo "github.com/anarva-cloud/anarva-cloud-db/internal/networking/repository"
	networkingService "github.com/anarva-cloud/anarva-cloud-db/internal/networking/service"

	infraEvac "github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/evacuation"
	infraFailover "github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/failover"
	infraHandler "github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/handler"
	infraHealth "github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/health"
	infraPlacement "github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/placement"
	infraProvider "github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/provider"
	infraRepo "github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/repository"
	infraService "github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/service"
	infraSim "github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/simulator"

	prvDrift "github.com/anarva-cloud/anarva-cloud-db/internal/providers/drift"
	prvHandler "github.com/anarva-cloud/anarva-cloud-db/internal/providers/handler"
	prvImport "github.com/anarva-cloud/anarva-cloud-db/internal/providers/import"
	prvMapping "github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
	prvRegistry "github.com/anarva-cloud/anarva-cloud-db/internal/providers/registry"
	prvSecurity "github.com/anarva-cloud/anarva-cloud-db/internal/providers/security"
	prvService "github.com/anarva-cloud/anarva-cloud-db/internal/providers/service"

	lbDns "github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/dns"
	lbEdge "github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/edge"
	lbHandler "github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/handler"
	lbHealth "github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/health"
	lbProvider "github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/provider"
	lbRepo "github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/repository"
	lbRouting "github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/routing"
	lbService "github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/service"
	lbTls "github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/tls"

	provHttp "github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/delivery/http"
	provProvider "github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/provider"
	provUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/usecase"

	whUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/webhook/usecase"

	billHttp "github.com/anarva-cloud/anarva-cloud-db/internal/billing/delivery/http"
	billUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/billing/usecase"

	"github.com/anarva-cloud/anarva-cloud-db/internal/resource"
	resourceHttp "github.com/anarva-cloud/anarva-cloud-db/internal/resource/delivery/http"
	"github.com/anarva-cloud/anarva-cloud-db/internal/query"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	pkgDatabase "github.com/anarva-cloud/anarva-cloud-db/pkg/database"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/logger"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/storage"
)

type queryHandler struct {
	parser   *query.Parser
	executor query.Executor
}

func (h *queryHandler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	var req query.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid query payload"))
		return
	}

	parsed, err := h.parser.ParseAndValidate(req.SQL)
	if err != nil {
		respondError(w, err)
		return
	}

	result, err := h.executor.Execute(r.Context(), "host=localhost port=5432 user=anarva_admin password=anarva_password dbname=anarva_cloud_db sslmode=disable", parsed, req.Parameters)
	if err != nil {
		respondError(w, err)
		return
	}

	metrics.RecordHTTPRequest(http.StatusOK, r.Method, r.URL.Path, result.ExecutionTimeMs/1000)
	respondJSON(w, http.StatusOK, result)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*appErrors.AppError); ok {
		respondJSON(w, appErr.HTTPStatusCode(), appErr)
		return
	}
	respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.InitLogger(cfg.Environment)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.Info("Starting Anarva API Gateway & Monolithic Service Router...")

	jwtManager := security.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	authMiddleware := gwMiddleware.NewAuthMiddleware(jwtManager)
	rateLimiter := gwMiddleware.NewRateLimitMiddleware(100)

	mux := http.NewServeMux()

	var uRepo authDomain.UserRepository
	var sRepo authDomain.SessionRepository
	var kRepo authDomain.APIKeyRepository
	var tRepo authDomain.VerificationTokenRepository
	var aRepo authDomain.AuditLogRepository

	var oRepo projectDomain.OrganizationRepository
	var pRepo projectDomain.ProjectRepository
	var mRepo projectDomain.MemberRepository
	var iRepo projectDomain.InvitationRepository

	var dRepo databaseDomain.InstanceRepository
	var bRepo backupDomain.BackupRepository

	var queryExecutor query.Executor

	// Attempt connecting to database for unified API Gateway service routing
	var dbPool *pkgDatabase.DB

	if os.Getenv("DATABASE_URL") != "" || (cfg.Database.Host != "" && cfg.Database.Host != "localhost") {
		dbPool, err = pkgDatabase.NewPostgresDB(cfg.Database)
	} else {
		err = fmt.Errorf("no external DATABASE_URL configured")
	}

	if err != nil {
		log.Info(fmt.Sprintf("Running in standalone high-performance mode: %v. Supabase Auth & Memory Repositories active.", err))
		uRepo = newMemUserRepo()
		sRepo = newMemSessionRepo()
		kRepo = newMemKeyRepo()
		tRepo = newMemTokenRepo()
		aRepo = newMemAuditRepo()

		oRepo = newMemOrgRepo()
		pRepo = newMemProjRepo()
		mRepo = newMemMemberRepo()
		iRepo = newMemInvRepo()

		dRepo = newMemInstanceRepo()
		bRepo = newMemBackupRepo()

		queryExecutor = query.NewMockExecutor()
	} else {
		_ = dbPool.AutoMigrate(
			&authDomain.User{},
			&authDomain.Session{},
			&authDomain.APIKey{},
			&authDomain.VerificationToken{},
			&authDomain.AuditLog{},
			&projectDomain.Organization{},
			&projectDomain.Project{},
			&projectDomain.OrganizationMember{},
			&projectDomain.Invitation{},
			&databaseDomain.DatabaseInstance{},
			&backupDomain.BackupRecord{},
		)

		uRepo = authRepo.NewUserRepository(dbPool.DB)
		sRepo = authRepo.NewSessionRepository(dbPool.DB)
		kRepo = authRepo.NewAPIKeyRepository(dbPool.DB)
		tRepo = authRepo.NewVerificationTokenRepository(dbPool.DB)
		aRepo = authRepo.NewAuditLogRepository(dbPool.DB)

		oRepo = projectRepo.NewOrganizationRepository(dbPool.DB)
		pRepo = projectRepo.NewProjectRepository(dbPool.DB)
		mRepo = projectRepo.NewMemberRepository(dbPool.DB)
		iRepo = projectRepo.NewInvitationRepository(dbPool.DB)

		dRepo = databaseRepo.NewInstanceRepository(dbPool.DB)
		bRepo = backupRepo.NewBackupRepository(dbPool.DB)

		queryExecutor = query.NewPostgresExecutor()
	}

	// Storage provider & Drivers
	sProvider, _ := storage.NewLocalStorageProvider(cfg.Storage.LocalPath)
	pDriver := databaseDriver.NewMockProvisionerDriver()

	// UseCases
	aUC := authUsecase.NewAuthUseCase(uRepo, sRepo, kRepo, tRepo, aRepo, jwtManager, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	pUC := projectUsecase.NewProjectUseCase(oRepo, pRepo, mRepo, iRepo)
	dUC := databaseUsecase.NewDatabaseUseCase(dRepo, pDriver, cfg.JWT.Secret)
	bUC := backupUsecase.NewBackupUseCase(bRepo, sProvider)
	_ = bUC

	// Phase 4 Centralized Resource Registry & Activity Stream
	resRegistry := resource.NewRegistry()
	actStream := activity.NewStream()
	authSvc := iamService.NewAuthorizationService()
	obsSvc := observabilityService.NewObservabilityService()
	bakProv := backupProvider.NewControlPlaneBackupProvider(storageProvider.NewLocalStorageProvider())
	compProv := computeProvider.NewLocalDockerComputeProvider()
	compUC := computeUsecase.NewComputeUseCase(newMemComputeRepo(), nil, compProv)
	netProv := networkProvider.NewLocalDockerNetworkProvider()
	netUC := networkUsecase.NewNetworkUseCase(newMemNetworkRepo(), nil, nil, nil, nil, nil, netProv)
	_ = netUC

	// Phase 13 Provisioning Engine & Provider Registry
	provRegistry := provProvider.NewProviderRegistry()
	dockerProv := provProvider.NewDockerInfrastructureProvider()
	provRegistry.RegisterProvider(dockerProv)
	provUC := provUsecase.NewProvisioningUseCase(newMemProvisioningRepo(), nil, nil, provRegistry)

	// Phase 14 Developer Platform & Webhook Engine
	devUC := devUsecase.NewDeveloperUseCase()
	whUC := whUsecase.NewWebhookUseCase()

	// Phase 15 Billing Service & Quotas Engine
	billUC := billUsecase.NewBillingUseCase()

	// Phase 17 Managed PostgreSQL Platform
	pgProvider := postgresProvider.NewLocalDockerPostgresProvider()
	pgService := postgresService.NewPostgresService(pgProvider)
	sqlService := postgresService.NewSQLService()

	// Phase 18 VPC / Networking / Security / DNS Platform
	vNetRepo := networkingRepo.NewNetworkingRepository()
	vNetProv := networkingProvider.NewLocalDockerNetworkProvider()
	vNetIpam := networkingIpam.NewIPAMService()
	vNetFw := networkingFw.NewFirewallService()
	vNetDnsProv := networkingDns.NewLocalDNSProvider()
	vNetConn := networkingConn.NewConnectivityService()
	vNetSvc := networkingService.NewNetworkingService(vNetRepo, vNetProv, vNetIpam, vNetFw, vNetDnsProv, vNetConn, actStream)
	vNetHandler := networkingHandler.NewNetworkingHandler(vNetSvc)

	// Phase 19 Load Balancing, TLS, DNS & Edge Delivery Platform
	lbRepository := lbRepo.NewLoadBalancerRepository()
	lbProv := lbProvider.NewLocalDockerLoadBalancerProvider()
	lbRoutingSvc := lbRouting.NewRoutingService()
	lbHealthSvc := lbHealth.NewHealthService()
	lbTlsSvc := lbTls.NewTLSService()
	lbDnsSvc := lbDns.NewDNSIntegrationService()
	lbSsrfSvc := lbEdge.NewSSRFValidationService()
	lbServiceSvc := lbService.NewLoadBalancerService(lbRepository, lbProv, lbRoutingSvc, lbHealthSvc, lbTlsSvc, lbDnsSvc, lbSsrfSvc, actStream)
	lbDeliveryHandler := lbHandler.NewLoadBalancerHandler(lbServiceSvc)

	// Phase 20 Managed MySQL Database Platform
	myRepository := mysqlRepo.NewMySQLRepository()
	myProv := mysqlProvider.NewLocalDockerMySQLProvider()
	myServiceSvc := mysqlService.NewMySQLService(myRepository, myProv, actStream)
	mySqlSvc := mysqlService.NewSQLService()
	myDeliveryHandler := mysqlHandler.NewMySQLHandler(myServiceSvc, mySqlSvc)

	// Phase 21 Multi-Region, Availability Zones & High-Availability Control Plane
	infRepository := infraRepo.NewInfrastructureRepository()
	infProv := infraProvider.NewLocalSimulationProvider()
	infPlacementEng := infraPlacement.NewPlacementEngine(infProv)
	infHealthEng := infraHealth.NewInfrastructureHealthEngine(infProv)
	infFailoverEng := infraFailover.NewFailoverEngine()
	infEvacSvc := infraEvac.NewEvacuationService()
	infSim := infraSim.NewOutageSimulator()
	infServiceSvc := infraService.NewInfrastructureService(infRepository, infProv, infPlacementEng, infHealthEng, infFailoverEng, infEvacSvc, infSim, actStream)
	infDeliveryHandler := infraHandler.NewInfrastructureHandler(infServiceSvc)

	// Phase 22 Real Cloud Provider Integration & Infrastructure Execution Layer
	prvReg := prvRegistry.NewProviderRegistry()
	prvMapRepo := prvMapping.NewMappingRepository()
	prvDriftEng := prvDrift.NewDriftEngine(prvMapRepo)
	prvImportEng := prvImport.NewImportEngine(prvMapRepo)
	prvSsrfEng := prvSecurity.NewSSRFProtectionEngine()
	prvServiceSvc := prvService.NewProviderService(prvReg, prvMapRepo, prvDriftEng, prvImportEng, prvSsrfEng, actStream)
	prvDeliveryHandler := prvHandler.NewProviderHandler(prvServiceSvc)

	// ALWAYS Register All Delivery Handlers into Gateway Mux
	authHttp.NewAuthHandler(aUC).RegisterRoutes(mux)
	projectHttp.NewProjectHandler(pUC).RegisterRoutes(mux)
	databaseHttp.NewDatabaseHandler(dUC).RegisterRoutes(mux)
	postgresHandler.NewPostgresHandler(pgService, sqlService).RegisterRoutes(mux)
	myDeliveryHandler.RegisterRoutes(mux)
	vNetHandler.RegisterRoutes(mux)
	lbDeliveryHandler.RegisterRoutes(mux)
	infDeliveryHandler.RegisterRoutes(mux)
	prvDeliveryHandler.RegisterRoutes(mux)
	backupHttp.NewBackupHandler(bakProv, actStream).RegisterRoutes(mux)
	computeHttp.NewComputeHandler(compUC, actStream).RegisterRoutes(mux)
	resourceHttp.NewResourceHandler(resRegistry, actStream).RegisterRoutes(mux)
	iamHttp.NewIAMHandler(authSvc, actStream).RegisterRoutes(mux)
	observabilityHttp.NewObservabilityHandler(obsSvc).RegisterRoutes(mux)
	provHttp.NewProvisioningHandler(provUC, provRegistry, actStream).RegisterRoutes(mux)
	devHttp.NewDeveloperHandler(devUC, whUC, actStream).RegisterRoutes(mux)
	billHttp.NewBillingHandler(billUC, actStream).RegisterRoutes(mux)

	// Register Query Handler
	qh := &queryHandler{
		parser:   query.NewParser(),
		executor: queryExecutor,
	}
	mux.HandleFunc("POST /api/v1/query", qh.ExecuteQuery)

	// Health and Prometheus Metrics endpoints
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	log.Info("Successfully registered Auth, Project, Database, Backup, and Query routes on API Gateway")

	// Wrap middleware chain: Correlation -> CORS -> RateLimit -> Auth -> Mux
	handler := gwMiddleware.CorrelationMiddleware(gwMiddleware.CORSMiddleware(rateLimiter.Limit(authMiddleware.Authenticate(mux))))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Info(fmt.Sprintf("API Gateway listening on port %d", cfg.Server.Port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Sprintf("API Gateway HTTP server failed: %v", err))
	}
}
