package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	authHttp "github.com/anarva-cloud/anarva-cloud-db/internal/auth/delivery/http"
	"github.com/anarva-cloud/anarva-cloud-db/internal/auth/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/auth/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/auth/usecase"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/database"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/logger"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
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

	log.Info("Starting Anarva Auth Service...")

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// Auto-Migrate Auth domain tables
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Session{},
		&domain.APIKey{},
		&domain.VerificationToken{},
		&domain.AuditLog{},
	); err != nil {
		log.Fatal(fmt.Sprintf("Failed to auto-migrate database schema: %v", err))
	}

	// Repositories
	userRepo := repository.NewUserRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)
	apiKeyRepo := repository.NewAPIKeyRepository(db.DB)
	tokenRepo := repository.NewVerificationTokenRepository(db.DB)
	auditRepo := repository.NewAuditLogRepository(db.DB)

	// JWT Manager
	jwtManager := security.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)

	// UseCase
	authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, apiKeyRepo, tokenRepo, auditRepo, jwtManager, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)

	// HTTP Delivery
	authHandler := authHttp.NewAuthHandler(authUseCase)
	mux := http.NewServeMux()
	authHandler.RegisterRoutes(mux)

	// Prometheus Metrics endpoint
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Info(fmt.Sprintf("Auth Service HTTP listening on port %d", cfg.Server.Port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Sprintf("HTTP server failed: %v", err))
	}
}
