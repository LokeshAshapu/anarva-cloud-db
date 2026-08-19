package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	appLogger "github.com/anarva-cloud/anarva-cloud-db/pkg/logger"
)

type DB struct {
	*gorm.DB
}

type SafeDBDiagnostics struct {
	Configured         bool   `json:"configured"`
	Scheme             string `json:"scheme"`
	HostConfigured     bool   `json:"host_configured"`
	PortConfigured     bool   `json:"port_configured"`
	DatabaseConfigured bool   `json:"database_configured"`
	UsernameConfigured bool   `json:"username_configured"`
	PasswordConfigured bool   `json:"password_configured"`
	SSLModeConfigured  bool   `json:"sslmode_configured"`
}

func GetSafeDatabaseDiagnostics(cfg config.DatabaseConfig) SafeDBDiagnostics {
	dsn := cfg.DSN()
	diag := SafeDBDiagnostics{}
	if dsn == "" {
		return diag
	}
	diag.Configured = true

	if u, err := url.Parse(dsn); err == nil && u.Scheme != "" {
		diag.Scheme = u.Scheme
		diag.HostConfigured = (u.Hostname() != "")
		diag.PortConfigured = (u.Port() != "")
		diag.DatabaseConfigured = (u.Path != "" && u.Path != "/")
		if u.User != nil {
			diag.UsernameConfigured = (u.User.Username() != "")
			_, set := u.User.Password()
			diag.PasswordConfigured = set
		}
		diag.SSLModeConfigured = u.Query().Has("sslmode")
		return diag
	}

	diag.Scheme = "postgres"
	diag.HostConfigured = (cfg.Host != "")
	diag.PortConfigured = (cfg.Port > 0)
	diag.DatabaseConfigured = (cfg.DBName != "")
	diag.UsernameConfigured = (cfg.User != "")
	diag.PasswordConfigured = (cfg.Password != "")
	diag.SSLModeConfigured = (cfg.SSLMode != "")
	return diag
}

// NewPostgresDB initializes a production-grade PostgreSQL GORM connection pool.
func NewPostgresDB(cfg config.DatabaseConfig) (*DB, error) {
	dsn := cfg.DSN()
	diag := GetSafeDatabaseDiagnostics(cfg)
	appLogger.Info(fmt.Sprintf("PostgreSQL Connection Diagnostic: configured=%t scheme=%s host_configured=%t port_configured=%t db_configured=%t user_configured=%t pass_configured=%t sslmode_configured=%t",
		diag.Configured, diag.Scheme, diag.HostConfigured, diag.PortConfigured, diag.DatabaseConfigured, diag.UsernameConfigured, diag.PasswordConfigured, diag.SSLModeConfigured))

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}

	appLogger.Info("Successfully established PostgreSQL connection pool")

	return &DB{DB: db}, nil
}

// HealthCheck verifies if the database ping succeeds.
func (db *DB) HealthCheck(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get raw database instance: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database healthcheck ping failed: %w", err)
	}

	return nil
}

// Close closes the underlying SQL database connection pool.
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
