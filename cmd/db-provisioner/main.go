package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	delivery "github.com/anarva-cloud/anarva-cloud-db/internal/database/delivery/http"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/driver"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/usecase"
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

	log.Info("Starting Anarva Database Provisioner & Managed Engine Service...")

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to connect to metadata database: %v", err))
	}

	// Auto-Migrate Database Provisioner metadata tables
	if err := db.AutoMigrate(
		&domain.DatabaseInstance{},
	); err != nil {
		log.Fatal(fmt.Sprintf("Failed to auto-migrate database instance metadata tables: %v", err))
	}

	// Repositories & Provisioner Driver
	repo := repository.NewInstanceRepository(db.DB)
	provDriver := driver.NewMockProvisionerDriver()

	// UseCase
	dbUseCase := usecase.NewDatabaseUseCase(repo, provDriver, cfg.JWT.Secret)

	// HTTP Delivery
	dbHandler := delivery.NewDatabaseHandler(dbUseCase)
	mux := http.NewServeMux()
	dbHandler.RegisterRoutes(mux)

	// Prometheus Metrics endpoint
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port+2), // Run on 8082
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Info(fmt.Sprintf("Database Provisioner HTTP listening on port %d", cfg.Server.Port+2))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Sprintf("HTTP server failed: %v", err))
	}
}
