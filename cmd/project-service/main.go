package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	projectHttp "github.com/anarva-cloud/anarva-cloud-db/internal/project/delivery/http"
	"github.com/anarva-cloud/anarva-cloud-db/internal/project/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/project/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/project/usecase"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/database"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/logger"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
)

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

	log.Info("Starting Anarva Project & Tenant Management Service...")

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// Auto-Migrate Project domain tables
	if err := db.AutoMigrate(
		&domain.Organization{},
		&domain.Project{},
		&domain.OrganizationMember{},
		&domain.Invitation{},
	); err != nil {
		log.Fatal(fmt.Sprintf("Failed to auto-migrate project database schema: %v", err))
	}

	// Repositories
	orgRepo := repository.NewOrganizationRepository(db.DB)
	projectRepo := repository.NewProjectRepository(db.DB)
	memberRepo := repository.NewMemberRepository(db.DB)
	invRepo := repository.NewInvitationRepository(db.DB)

	// UseCase
	projectUseCase := usecase.NewProjectUseCase(orgRepo, projectRepo, memberRepo, invRepo)

	// HTTP Delivery
	projectHandler := projectHttp.NewProjectHandler(projectUseCase)
	mux := http.NewServeMux()
	projectHandler.RegisterRoutes(mux)

	// Prometheus Metrics endpoint
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port+1), // Run on 8081 or config
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Info(fmt.Sprintf("Project Service HTTP listening on port %d", cfg.Server.Port+1))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Sprintf("HTTP server failed: %v", err))
	}
}
