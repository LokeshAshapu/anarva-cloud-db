package provider

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/domain"
)

var validIdentifierRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// SanitizeIdentifier validates that a database or role identifier contains only alphanumeric characters and underscores.
func SanitizeIdentifier(ident string) (string, error) {
	clean := strings.TrimSpace(ident)
	if clean == "" {
		return "", fmt.Errorf("identifier cannot be empty")
	}
	if !validIdentifierRegex.MatchString(clean) {
		return "", fmt.Errorf("invalid identifier '%s': contains unsafe characters", ident)
	}
	return clean, nil
}

// GenerateStrongPassword generates a 24-character cryptographically random password using crypto/rand.
func GenerateStrongPassword() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure random password: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

type PostgresDataPlaneProvider interface {
	CreateDatabase(ctx context.Context, inst *domain.PostgresInstance) (*domain.PostgresInstance, string, error)
	DeleteDatabase(ctx context.Context, inst *domain.PostgresInstance) error
	GetInstanceHealth(ctx context.Context, inst *domain.PostgresInstance) (bool, error)

	StartDatabase(ctx context.Context, inst *domain.PostgresInstance) error
	StopDatabase(ctx context.Context, inst *domain.PostgresInstance) error
	RestartDatabase(ctx context.Context, inst *domain.PostgresInstance) error
}
type RealPostgresDataPlaneProvider struct {
	adminDSN string
}

func NewRealPostgresDataPlaneProvider(adminDSN string) *RealPostgresDataPlaneProvider {
	return &RealPostgresDataPlaneProvider{adminDSN: adminDSN}
}

func (p *RealPostgresDataPlaneProvider) CreateDatabase(ctx context.Context, inst *domain.PostgresInstance) (*domain.PostgresInstance, string, error) {
	if p.adminDSN == "" {
		return nil, "", fmt.Errorf("missing CUSTOMER_DATABASE_ADMIN_URL for real PostgreSQL provisioning")
	}

	// 1. Sanitize Database and Role Identifiers
	rawSuffix := strings.ReplaceAll(inst.ID, "-", "_")
	rawSuffix = strings.ReplaceAll(rawSuffix, ":", "_")

	dbName, err := SanitizeIdentifier(fmt.Sprintf("db_%s", rawSuffix))
	if err != nil {
		return nil, "", fmt.Errorf("failed to sanitize database name: %w", err)
	}

	roleName, err := SanitizeIdentifier(fmt.Sprintf("usr_%s", rawSuffix))
	if err != nil {
		return nil, "", fmt.Errorf("failed to sanitize role name: %w", err)
	}

	// 2. Generate Cryptographically Strong Password
	password, err := GenerateStrongPassword()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate password: %w", err)
	}

	// 3. Connect to Data-Plane PostgreSQL Server via Admin DSN
	adminDB, err := sql.Open("pgx", p.adminDSN)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open data-plane admin connection: %w", err)
	}
	defer adminDB.Close()

	if err := adminDB.PingContext(ctx); err != nil {
		return nil, "", fmt.Errorf("failed to ping data-plane PostgreSQL admin server: %w", err)
	}

	// 4. Provision Role
	createRoleQuery := fmt.Sprintf(
		"CREATE ROLE %s WITH LOGIN PASSWORD '%s';",
		roleName,
		password,
	)

	if _, err := adminDB.ExecContext(ctx, createRoleQuery); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return nil, "", fmt.Errorf(
				"failed to create postgres role '%s': %w",
				roleName,
				err,
			)
		}

		alterRoleQuery := fmt.Sprintf(
			"ALTER ROLE %s WITH LOGIN PASSWORD '%s';",
			roleName,
			password,
		)

		if _, err := adminDB.ExecContext(ctx, alterRoleQuery); err != nil {
			return nil, "", fmt.Errorf(
				"failed to reset password for existing postgres role '%s': %w",
				roleName,
				err,
			)
		}
	}

	// 5. Provision Database with Rollback Cleanup on Failure
	createDBQuery := fmt.Sprintf("CREATE DATABASE %s WITH OWNER %s;", dbName, roleName)
	if _, err := adminDB.ExecContext(ctx, createDBQuery); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			// Rollback cleanup: attempt to drop role
			_, _ = adminDB.ExecContext(ctx, fmt.Sprintf("DROP ROLE IF EXISTS %s;", roleName))
			return nil, "", fmt.Errorf("failed to create postgres database '%s': %w", dbName, err)
		}
	}

	// 6. Configure Role Permissions
	revokeQuery := fmt.Sprintf("REVOKE ALL ON DATABASE %s FROM PUBLIC;", dbName)
	_, _ = adminDB.ExecContext(ctx, revokeQuery)

	grantQuery := fmt.Sprintf("GRANT ALL ON DATABASE %s TO %s;", dbName, roleName)
	_, _ = adminDB.ExecContext(ctx, grantQuery)

	// 7. Update Instance Record
	inst.ProviderResourceId = dbName
	inst.Host = extractHostFromDSN(p.adminDSN)
	inst.Port = extractPortFromDSN(p.adminDSN)
	inst.Status = domain.StatusAvailable
	inst.RealityLabel = "REAL_POSTGRES (DATA_PLANE_PROVISIONED)"

	return inst, password, nil
}

func (p *RealPostgresDataPlaneProvider) DeleteDatabase(ctx context.Context, inst *domain.PostgresInstance) error {
	if p.adminDSN == "" {
		return nil
	}

	rawSuffix := strings.ReplaceAll(inst.ID, "-", "_")
	rawSuffix = strings.ReplaceAll(rawSuffix, ":", "_")

	dbName, err := SanitizeIdentifier(fmt.Sprintf("db_%s", rawSuffix))
	if err != nil {
		return err
	}

	roleName, err := SanitizeIdentifier(fmt.Sprintf("usr_%s", rawSuffix))
	if err != nil {
		return err
	}

	adminDB, err := sql.Open("pgx", p.adminDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to data-plane admin server for deletion: %w", err)
	}
	defer adminDB.Close()

	// Revoke active connections and drop DB
	dropDBQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", dbName)
	if _, err := adminDB.ExecContext(ctx, dropDBQuery); err != nil {
		// Fallback without WITH (FORCE) if older PG version
		_, _ = adminDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbName))
	}

	dropRoleQuery := fmt.Sprintf("DROP ROLE IF EXISTS %s;", roleName)
	_, _ = adminDB.ExecContext(ctx, dropRoleQuery)

	return nil
}
func (p *RealPostgresDataPlaneProvider) CreateUser(
	ctx context.Context,
	instanceID, username string,
	role domain.UserRole,
	password string,
) (*domain.PostgresUser, error) {
	if p.adminDSN == "" {
		return nil, fmt.Errorf("missing CUSTOMER_DATABASE_ADMIN_URL")
	}

	username, err := SanitizeIdentifier(username)
	if err != nil {
		return nil, err
	}

	if password == "" {
		password, err = GenerateStrongPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to generate secure password: %w", err)
		}
	}

	adminDB, err := sql.Open("pgx", p.adminDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open data-plane admin connection: %w", err)
	}
	defer adminDB.Close()

	if err := adminDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping data-plane PostgreSQL: %w", err)
	}

	escapedPassword := strings.ReplaceAll(password, "'", "''")

	query := fmt.Sprintf(
		"CREATE ROLE %s WITH LOGIN PASSWORD '%s';",
		username,
		escapedPassword,
	)

	if _, err := adminDB.ExecContext(ctx, query); err != nil {
		return nil, fmt.Errorf("failed to create postgres user '%s': %w", username, err)
	}

	return &domain.PostgresUser{
		ID:                  fmt.Sprintf("usr-%s-%s", instanceID, username),
		InstanceID:          instanceID,
		Username:            username,
		Role:                role,
		Status:              "ACTIVE",
		CredentialReference: fmt.Sprintf("secret-ref-%s", instanceID),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}, nil
}
func (p *RealPostgresDataPlaneProvider) DeleteUser(
	ctx context.Context,
	instanceID, username string,
) error {
	if p.adminDSN == "" {
		return fmt.Errorf("missing CUSTOMER_DATABASE_ADMIN_URL")
	}

	username, err := SanitizeIdentifier(username)
	if err != nil {
		return err
	}

	adminDB, err := sql.Open("pgx", p.adminDSN)
	if err != nil {
		return fmt.Errorf("failed to open data-plane admin connection: %w", err)
	}
	defer adminDB.Close()

	if err := adminDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping data-plane PostgreSQL: %w", err)
	}

	query := fmt.Sprintf(
		"DROP ROLE IF EXISTS %s;",
		username,
	)

	if _, err := adminDB.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to delete postgres user '%s': %w", username, err)
	}

	return nil
}
func (p *RealPostgresDataPlaneProvider) RotateCredentials(
	ctx context.Context,
	instanceID, username, newPassword string,
) error {
	if p.adminDSN == "" {
		return fmt.Errorf("missing CUSTOMER_DATABASE_ADMIN_URL")
	}

	username, err := SanitizeIdentifier(username)
	if err != nil {
		return err
	}

	if newPassword == "" {
		newPassword, err = GenerateStrongPassword()
		if err != nil {
			return fmt.Errorf("failed to generate secure password: %w", err)
		}
	}

	adminDB, err := sql.Open("pgx", p.adminDSN)
	if err != nil {
		return fmt.Errorf("failed to open data-plane admin connection: %w", err)
	}
	defer adminDB.Close()

	if err := adminDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping data-plane PostgreSQL: %w", err)
	}

	escapedPassword := strings.ReplaceAll(newPassword, "'", "''")

	query := fmt.Sprintf(
		"ALTER ROLE %s WITH PASSWORD '%s';",
		username,
		escapedPassword,
	)

	if _, err := adminDB.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to rotate credentials for postgres user '%s': %w", username, err)
	}

	return nil
}
func (p *RealPostgresDataPlaneProvider) GetInstanceHealth(ctx context.Context, inst *domain.PostgresInstance) (bool, error) {
	if p.adminDSN == "" {
		return true, nil
	}

	adminDB, err := sql.Open("pgx", p.adminDSN)
	if err != nil {
		return false, err
	}
	defer adminDB.Close()

	if err := adminDB.PingContext(ctx); err != nil {
		return false, err
	}
	return true, nil
}
func (p *RealPostgresDataPlaneProvider) StartDatabase(ctx context.Context, inst *domain.PostgresInstance) error {
	if p.adminDSN == "" {
		return fmt.Errorf("missing CUSTOMER_DATABASE_ADMIN_URL")
	}

	healthy, err := p.GetInstanceHealth(ctx, inst)
	if err != nil {
		return fmt.Errorf("data-plane PostgreSQL is unavailable: %w", err)
	}
	if !healthy {
		return fmt.Errorf("data-plane PostgreSQL is unavailable")
	}

	return nil
}

func (p *RealPostgresDataPlaneProvider) StopDatabase(ctx context.Context, inst *domain.PostgresInstance) error {
	if p.adminDSN == "" {
		return fmt.Errorf("missing CUSTOMER_DATABASE_ADMIN_URL")
	}

	// The current real data plane is a shared PostgreSQL server.
	// A single customer database cannot be physically stopped independently.
	// Lifecycle state is therefore enforced at the control plane.
	return nil
}

func (p *RealPostgresDataPlaneProvider) RestartDatabase(ctx context.Context, inst *domain.PostgresInstance) error {
	if p.adminDSN == "" {
		return fmt.Errorf("missing CUSTOMER_DATABASE_ADMIN_URL")
	}

	healthy, err := p.GetInstanceHealth(ctx, inst)
	if err != nil {
		return fmt.Errorf("data-plane PostgreSQL is unavailable after restart: %w", err)
	}
	if !healthy {
		return fmt.Errorf("data-plane PostgreSQL is unavailable after restart")
	}

	return nil
}

type SimulatedDataPlaneProvider struct {
	mu        sync.RWMutex
	instances map[string]*domain.PostgresInstance
}

func NewSimulatedDataPlaneProvider() *SimulatedDataPlaneProvider {
	return &SimulatedDataPlaneProvider{
		instances: make(map[string]*domain.PostgresInstance),
	}
}

func (p *SimulatedDataPlaneProvider) CreateDatabase(ctx context.Context, inst *domain.PostgresInstance) (*domain.PostgresInstance, string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	rawSuffix := strings.ReplaceAll(inst.ID, "-", "_")
	dbName, _ := SanitizeIdentifier(fmt.Sprintf("db_%s", rawSuffix))
	pass, _ := GenerateStrongPassword()

	inst.ProviderResourceId = dbName
	inst.Host = "localhost"
	inst.Port = 5432
	inst.Status = domain.StatusAvailable
	inst.RealityLabel = "LOCAL_POSTGRES (SIMULATED_PROVISIONING)"

	p.instances[inst.ID] = inst
	return inst, pass, nil
}

func (p *SimulatedDataPlaneProvider) GetInstance(ctx context.Context, instanceID string) (*domain.PostgresInstance, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	inst, ok := p.instances[instanceID]
	if !ok {
		return nil, fmt.Errorf("database instance '%s' not found", instanceID)
	}
	return inst, nil
}

func (p *SimulatedDataPlaneProvider) DeleteDatabase(ctx context.Context, inst *domain.PostgresInstance) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.instances, inst.ID)
	return nil
}

func (p *SimulatedDataPlaneProvider) GetInstanceHealth(ctx context.Context, inst *domain.PostgresInstance) (bool, error) {
	return true, nil
}

func extractHostFromDSN(dsn string) string {
	if idx := strings.Index(dsn, "@"); idx != -1 {
		rest := dsn[idx+1:]
		if slashIdx := strings.Index(rest, "/"); slashIdx != -1 {
			hostPort := rest[:slashIdx]
			if colonIdx := strings.Index(hostPort, ":"); colonIdx != -1 {
				return hostPort[:colonIdx]
			}
			return hostPort
		}
	}
	return "localhost"
}
func extractPortFromDSN(dsn string) int {
	if idx := strings.Index(dsn, "@"); idx != -1 {
		rest := dsn[idx+1:]
		if slashIdx := strings.Index(rest, "/"); slashIdx != -1 {
			hostPort := rest[:slashIdx]
			if colonIdx := strings.LastIndex(hostPort, ":"); colonIdx != -1 {
				if port, err := strconv.Atoi(hostPort[colonIdx+1:]); err == nil && port > 0 {
					return port
				}
			}
		}
	}
	return 5432
}
func (p *SimulatedDataPlaneProvider) RestartDatabase(ctx context.Context, inst *domain.PostgresInstance) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	existing, ok := p.instances[inst.ID]
	if !ok {
		return fmt.Errorf("database instance '%s' not found", inst.ID)
	}

	existing.Status = domain.StatusAvailable
	return nil
}
func (p *SimulatedDataPlaneProvider) StartDatabase(ctx context.Context, inst *domain.PostgresInstance) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	existing, ok := p.instances[inst.ID]
	if !ok {
		return fmt.Errorf("database instance '%s' not found", inst.ID)
	}

	existing.Status = domain.StatusAvailable
	return nil
}
func (p *SimulatedDataPlaneProvider) StopDatabase(ctx context.Context, inst *domain.PostgresInstance) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	existing, ok := p.instances[inst.ID]
	if !ok {
		return fmt.Errorf("database instance '%s' not found", inst.ID)
	}

	existing.Status = domain.StatusStopped
	return nil
}
