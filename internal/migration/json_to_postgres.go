package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"

	authDomain "github.com/anarva-cloud/anarva-cloud-db/internal/auth/domain"
	databaseDomain "github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	projDomain "github.com/anarva-cloud/anarva-cloud-db/internal/project/domain"
	pkgLogger "github.com/anarva-cloud/anarva-cloud-db/pkg/logger"
)

type MigrationReport struct {
	UsersMigrated         int `json:"users_migrated"`
	OrganizationsMigrated int `json:"organizations_migrated"`
	ProjectsMigrated      int `json:"projects_migrated"`
	DatabasesMigrated     int `json:"databases_migrated"`
}

func MigrateJSONToPostgres(ctx context.Context, db *gorm.DB, dataDir string) (*MigrationReport, error) {
	if db == nil {
		return nil, fmt.Errorf("nil database connection")
	}
	if dataDir == "" {
		dataDir = "./data"
	}

	report := &MigrationReport{}

	// 1. Migrate Users
	userFile := filepath.Join(dataDir, "anarva_cp_users.json")
	if data, err := os.ReadFile(userFile); err == nil {
		var users map[string]*authDomain.User
		if err := json.Unmarshal(data, &users); err == nil {
			for _, u := range users {
				if u.ID == "" || u.Email == "" {
					continue
				}
				var count int64
				db.WithContext(ctx).Model(&authDomain.User{}).Where("id = ? OR email = ?", u.ID, u.Email).Count(&count)
				if count == 0 {
					if u.CreatedAt.IsZero() {
						u.CreatedAt = time.Now()
					}
					if u.UpdatedAt.IsZero() {
						u.UpdatedAt = time.Now()
					}
					if err := db.WithContext(ctx).Create(u).Error; err == nil {
						report.UsersMigrated++
					}
				}
			}
		}
	}

	// 2. Migrate Organizations
	orgFile := filepath.Join(dataDir, "anarva_cp_orgs.json")
	if data, err := os.ReadFile(orgFile); err == nil {
		var orgs map[string]*projDomain.Organization
		if err := json.Unmarshal(data, &orgs); err == nil {
			for _, o := range orgs {
				if o.ID == "" {
					continue
				}
				var count int64
				db.WithContext(ctx).Model(&projDomain.Organization{}).Where("id = ?", o.ID).Count(&count)
				if count == 0 {
					if o.CreatedAt.IsZero() {
						o.CreatedAt = time.Now()
					}
					if err := db.WithContext(ctx).Create(o).Error; err == nil {
						report.OrganizationsMigrated++
					}
				}
			}
		}
	}

	// 3. Migrate Projects
	projFile := filepath.Join(dataDir, "anarva_cp_projects.json")
	if data, err := os.ReadFile(projFile); err == nil {
		var projs map[string]*projDomain.Project
		if err := json.Unmarshal(data, &projs); err == nil {
			for _, p := range projs {
				if p.ID == "" {
					continue
				}
				var count int64
				db.WithContext(ctx).Model(&projDomain.Project{}).Where("id = ?", p.ID).Count(&count)
				if count == 0 {
					if p.CreatedAt.IsZero() {
						p.CreatedAt = time.Now()
					}
					if err := db.WithContext(ctx).Create(p).Error; err == nil {
						report.ProjectsMigrated++
					}
				}
			}
		}
	}

	// 4. Migrate Databases
	instFile := filepath.Join(dataDir, "anarva_cp_instances.json")
	if data, err := os.ReadFile(instFile); err == nil {
		var insts map[string]*databaseDomain.DatabaseInstance
		if err := json.Unmarshal(data, &insts); err == nil {
			for _, inst := range insts {
				if inst.ID == "" {
					continue
				}
				var count int64
				db.WithContext(ctx).Model(&databaseDomain.DatabaseInstance{}).Where("id = ?", inst.ID).Count(&count)
				if count == 0 {
					if inst.CreatedAt.IsZero() {
						inst.CreatedAt = time.Now()
					}
					if err := db.WithContext(ctx).Create(inst).Error; err == nil {
						report.DatabasesMigrated++
					}
				}
			}
		}
	}

	pkgLogger.Info(fmt.Sprintf("[MIGRATION] JSON to PostgreSQL Migration Complete: %d users, %d orgs, %d projects, %d databases imported into PostgreSQL.",
		report.UsersMigrated, report.OrganizationsMigrated, report.ProjectsMigrated, report.DatabasesMigrated))

	return report, nil
}
