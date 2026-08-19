package security

import (
	"context"
	"fmt"
)

type ContextKey string

const (
	UserIDKey    ContextKey = "user_id"
	RoleKey      ContextKey = "role"
	OrgIDKey     ContextKey = "org_id"
	ProjectIDKey ContextKey = "project_id"
)

type TenantContext struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
	Role           string `json:"role"`
}

func GetTenantContext(ctx context.Context) *TenantContext {
	tc := &TenantContext{
		UserID:         "usr-dev",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		Role:           "ADMIN",
	}

	if uid, ok := ctx.Value(UserIDKey).(string); ok && uid != "" {
		tc.UserID = uid
	}
	if role, ok := ctx.Value(RoleKey).(string); ok && role != "" {
		tc.Role = role
	}
	if org, ok := ctx.Value(OrgIDKey).(string); ok && org != "" {
		tc.OrganizationID = org
	}
	if proj, ok := ctx.Value(ProjectIDKey).(string); ok && proj != "" {
		tc.ProjectID = proj
	}

	return tc
}

func (tc *TenantContext) EnforceOwnership(resourceOrgID, resourceProjectID string) error {
	if resourceOrgID != "" && tc.OrganizationID != "" && resourceOrgID != tc.OrganizationID {
		return fmt.Errorf("TENANT_ISOLATION_VIOLATION: Organization access denied")
	}
	if resourceProjectID != "" && tc.ProjectID != "" && resourceProjectID != tc.ProjectID {
		return fmt.Errorf("TENANT_ISOLATION_VIOLATION: Project access denied")
	}
	return nil
}
