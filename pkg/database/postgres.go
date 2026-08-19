package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
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

type DatabaseIdentity struct {
	Database        string `json:"database"`
	User            string `json:"user"`
	ServerReachable bool   `json:"server_reachable"`
	ServerVersion   string `json:"server_version,omitempty"`
}

type ForensicDiagnosticsResult struct {
	Configured           bool   `json:"configured"`
	Scheme               string `json:"scheme"`
	Hostname             string `json:"hostname"`
	Port                 int    `json:"port"`
	Database             string `json:"database"`
	SSLMode              string `json:"sslmode"`
	DNSResolution        string `json:"dns_resolution"`
	TCPConnection        string `json:"tcp_connection"`
	PostgresPing         string `json:"postgres_ping"`
	ConnectionErrorClass string `json:"connection_error_class"`
	DatabaseDriver       string `json:"database_driver"`
	ErrorMessage         string `json:"error_message,omitempty"`
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

func PerformForensicDiagnostics(cfg config.DatabaseConfig) ForensicDiagnosticsResult {
	res := ForensicDiagnosticsResult{
		DatabaseDriver: "postgres",
		DNSResolution:  "NOT_TESTED",
		TCPConnection:  "NOT_TESTED",
		PostgresPing:   "NOT_TESTED",
	}

	dsn := cfg.DSN()
	if dsn == "" {
		res.ConnectionErrorClass = "DATABASE_CONFIG_MISSING"
		res.ErrorMessage = "No DATABASE_URL or database configuration supplied"
		return res
	}
	res.Configured = true

	var host, dbName, sslMode string
	port := 5432
	scheme := "postgres"

	if u, err := url.Parse(dsn); err == nil && u.Scheme != "" {
		scheme = u.Scheme
		host = u.Hostname()
		if pStr := u.Port(); pStr != "" {
			if pVal, pErr := strconv.Atoi(pStr); pErr == nil {
				port = pVal
			}
		}
		dbName = strings.TrimPrefix(u.Path, "/")
		sslMode = u.Query().Get("sslmode")
	} else {
		host = cfg.Host
		port = cfg.Port
		if port == 0 {
			port = 5432
		}
		dbName = cfg.DBName
		sslMode = cfg.SSLMode
	}

	if sslMode == "" {
		sslMode = "prefer"
	}

	res.Scheme = scheme
	res.Hostname = host
	res.Port = port
	res.Database = dbName
	res.SSLMode = sslMode

	if host == "" {
		res.ConnectionErrorClass = "DATABASE_CONFIG_MISSING"
		res.ErrorMessage = "Hostname is unconfigured"
		return res
	}

	// 1. DNS Resolution Test
	addrs, dnsErr := net.LookupHost(host)
	if dnsErr != nil {
		res.DNSResolution = "FAIL"
		res.ConnectionErrorClass = "DATABASE_DNS_FAILURE"
		res.ErrorMessage = fmt.Sprintf("DNS resolution failed for hostname '%s': %v", host, dnsErr)
		return res
	}
	res.DNSResolution = "PASS"

	// 2. TCP Dial Test
	dialAddr := fmt.Sprintf("%s:%d", addrs[0], port)
	conn, tcpErr := net.DialTimeout("tcp", dialAddr, 3*time.Second)
	if tcpErr != nil {
		res.TCPConnection = "FAIL"
		if strings.Contains(tcpErr.Error(), "refused") {
			res.ConnectionErrorClass = "DATABASE_CONNECTION_REFUSED"
		} else {
			res.ConnectionErrorClass = "DATABASE_CONNECTION_TIMEOUT"
		}
		res.ErrorMessage = fmt.Sprintf("TCP dial to %s failed: %v", dialAddr, tcpErr)
		return res
	}
	_ = conn.Close()
	res.TCPConnection = "PASS"

	// 3. PostgreSQL GORM Connection Ping Test
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		res.PostgresPing = "FAIL"
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "password") || strings.Contains(errStr, "auth") || strings.Contains(errStr, "denied") {
			res.ConnectionErrorClass = "DATABASE_AUTH_FAILED"
		} else if strings.Contains(errStr, "tls") || strings.Contains(errStr, "ssl") || strings.Contains(errStr, "certificate") {
			res.ConnectionErrorClass = "DATABASE_TLS_FAILURE"
		} else if strings.Contains(errStr, "does not exist") {
			res.ConnectionErrorClass = "DATABASE_NOT_FOUND"
		} else {
			res.ConnectionErrorClass = "DATABASE_UNAVAILABLE"
		}
		res.ErrorMessage = err.Error()
		return res
	}

	sqlDB, err := db.DB()
	if err != nil {
		res.PostgresPing = "FAIL"
		res.ConnectionErrorClass = "DATABASE_UNAVAILABLE"
		res.ErrorMessage = err.Error()
		return res
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		res.PostgresPing = "FAIL"
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "password") || strings.Contains(errStr, "auth") || strings.Contains(errStr, "denied") {
			res.ConnectionErrorClass = "DATABASE_AUTH_FAILED"
		} else if strings.Contains(errStr, "tls") || strings.Contains(errStr, "ssl") || strings.Contains(errStr, "certificate") {
			res.ConnectionErrorClass = "DATABASE_TLS_FAILURE"
		} else if strings.Contains(errStr, "does not exist") {
			res.ConnectionErrorClass = "DATABASE_NOT_FOUND"
		} else {
			res.ConnectionErrorClass = "DATABASE_UNAVAILABLE"
		}
		res.ErrorMessage = err.Error()
		return res
	}

	res.PostgresPing = "PASS"
	res.ConnectionErrorClass = "DATABASE_CONNECTED"
	return res
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

func (db *DB) GetDatabaseIdentity(ctx context.Context) (*DatabaseIdentity, error) {
	var currentDB, currentUser, version string
	row := db.DB.WithContext(ctx).Raw("SELECT current_database(), current_user, version()").Row()
	if err := row.Scan(&currentDB, &currentUser, &version); err != nil {
		return nil, err
	}
	return &DatabaseIdentity{
		Database:        currentDB,
		User:            currentUser,
		ServerReachable: true,
		ServerVersion:   version,
	}, nil
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
