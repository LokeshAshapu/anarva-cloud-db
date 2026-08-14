package iam

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/iam/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/iam/service"
)

func TestIAM_RoleEvaluation(t *testing.T) {
	authSvc := service.NewAuthorizationService()

	// Owner has full permissions
	ok, err := authSvc.Authorize("lokeshashapu@gmail.com", "org-default", "proj-default", domain.PermOrgDelete)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = authSvc.Authorize("lokeshashapu@gmail.com", "org-default", "proj-default", domain.PermComputeDelete)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = authSvc.Authorize("lokeshashapu@gmail.com", "org-default", "proj-default", domain.PermBillingManage)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestIAM_LastOwnerProtection(t *testing.T) {
	authSvc := service.NewAuthorizationService()

	// Attempting to remove the sole owner must trigger LAST_OWNER_PROTECTION
	err := authSvc.RemoveMember("org-default", "mem-101")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LAST_OWNER_PROTECTION")
}

func TestIAM_ProjectDeletionProtection(t *testing.T) {
	authSvc := service.NewAuthorizationService()

	// Deleting project containing active resources must trigger PROJECT_NOT_EMPTY
	err := authSvc.DeleteProject("org-default", "proj-default")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PROJECT_NOT_EMPTY")
}

func TestIAM_SingleUseInvitations(t *testing.T) {
	authSvc := service.NewAuthorizationService()

	inv, token, err := authSvc.CreateInvitation("org-default", "dev@anarva.cloud", domain.RoleDeveloper)
	require.NoError(t, err)
	assert.NotNil(t, inv)
	assert.NotEmpty(t, token)
	assert.Equal(t, "PENDING", inv.Status)
	assert.Equal(t, domain.RoleDeveloper, inv.Role)
}
