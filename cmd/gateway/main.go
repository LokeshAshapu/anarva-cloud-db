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
	backupHttp "github.com/anarva-cloud/anarva-cloud-db/internal/backup/delivery/http"
	backupProvider "github.com/anarva-cloud/anarva-cloud-db/internal/backup/provider"
	iamHttp "github.com/anarva-cloud/anarva-cloud-db/internal/iam/delivery/http"
	iamService "github.com/anarva-cloud/anarva-cloud-db/internal/iam/service"
	observabilityHttp "github.com/anarva-cloud/anarva-cloud-db/internal/observability/delivery/http"
	observabilityService "github.com/anarva-cloud/anarva-cloud-db/internal/observability/service"
	storageProvider "github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/resource"
	resourceHttp "github.com/anarva-cloud/anarva-cloud-db/internal/resource/delivery/http"

	gwMiddleware "github.com/anarva-cloud/anarva-cloud-db/internal/gateway/middleware"
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
			&backupDomain.BackupSnapshot{},
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

	// Phase 4 Centralized Resource Registry & Activity Stream
	resRegistry := resource.NewRegistry()
	actStream := activity.NewStream()
	authSvc := iamService.NewAuthorizationService()
	obsSvc := observabilityService.NewObservabilityService()
	bakProv := backupProvider.NewControlPlaneBackupProvider(storageProvider.NewLocalStorageProvider())

	// ALWAYS Register All Delivery Handlers into Gateway Mux
	authHttp.NewAuthHandler(aUC).RegisterRoutes(mux)
	projectHttp.NewProjectHandler(pUC).RegisterRoutes(mux)
	databaseHttp.NewDatabaseHandler(dUC).RegisterRoutes(mux)
	backupHttp.NewBackupHandler(bUC).RegisterRoutes(mux)
	backupHttp.NewBackupHandler(bakProv, actStream).RegisterRoutes(mux)
	resourceHttp.NewResourceHandler(resRegistry, actStream).RegisterRoutes(mux)
	iamHttp.NewIAMHandler(authSvc, actStream).RegisterRoutes(mux)
	observabilityHttp.NewObservabilityHandler(obsSvc).RegisterRoutes(mux)

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

	// Wrap middleware chain: CORS -> RateLimit -> Auth -> Mux
	handler := gwMiddleware.CORSMiddleware(rateLimiter.Limit(authMiddleware.Authenticate(mux)))

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
