package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	delivery "github.com/anarva-cloud/anarva-cloud-db/internal/backup/delivery/http"
	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/usecase"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/database"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/logger"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/storage"
        "github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	backupProvider "github.com/anarva-cloud/anarva-cloud-db/internal/backup/provider"
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

	log.Info("Starting Anarva Backup & Restore Service...")

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to connect to metadata database: %v", err))
	}

	// Auto-Migrate current Backup metadata schema.
if err := db.AutoMigrate(&domain.BackupRecord{}); err != nil {
        log.Fatal(fmt.Sprintf("Failed to auto-migrate backup database schema: %v", err))
}

// Storage Provider Driver.
sProvider, err := storage.NewStorageProvider(cfg.Storage, cfg.Environment)
if err != nil {
        if cfg.Environment == "production" {
                log.Fatal(fmt.Sprintf("Failed to initialize storage provider: %v", err))
        }

        log.Info(fmt.Sprintf(
                "Development Notice: storage provider initialization returned: %v. Falling back to local storage.",
                err,
        ))

        sProvider = storage.NewLocalStorageProvider(cfg.Storage.LocalPath)
}

// Repository + UseCase + Provider.
repo := repository.NewBackupRepository(db.DB)

backupUseCase := usecase.NewBackupUseCase(
        repo,
        sProvider,
        cfg.Storage.S3Bucket,
)

	bakProv := backupProvider.NewControlPlaneBackupProviderWithUseCase(
        	backupUseCase,
        	sProvider,
	)

	// Activity stream.
	actStream := activity.NewStream()

	// HTTP Delivery.
	backupHandler := delivery.NewBackupHandlerWithUseCase(
        	bakProv,
        	backupUseCase,
        	actStream,
	)
	mux := http.NewServeMux()
	backupHandler.RegisterRoutes(mux)

	// Prometheus Metrics endpoint
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port+3), // Run on 8083
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Info(fmt.Sprintf("Backup Service HTTP listening on port %d", cfg.Server.Port+3))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Sprintf("HTTP server failed: %v", err))
	}
}
