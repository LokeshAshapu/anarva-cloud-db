package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authDomain "github.com/anarva-cloud/anarva-cloud-db/internal/auth/domain"
)

func TestUserDurablePersistence_SurvivesProcessRestart(t *testing.T) {
	ctx := context.Background()

	// 1. Create a unique user marker
	ts := time.Now().UnixNano()
	testEmail := fmt.Sprintf("qa_persistence_%d@anarva.io", ts)
	testUserID := fmt.Sprintf("usr-qa-pers-%d", ts)

	// 2. Instantiate repository 1 (Simulating initial server process)
	repo1 := newMemUserRepo()
	newUser := &authDomain.User{
		ID:        testUserID,
		Email:     testEmail,
		FullName:  "QA Persistence User",
		Role:      authDomain.RoleAdmin,
		Status:    authDomain.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo1.Create(ctx, newUser)
	require.NoError(t, err)

	// Verify User 1 exists in repo1
	u1, err1 := repo1.GetByEmail(ctx, testEmail)
	require.NoError(t, err1)
	assert.Equal(t, testUserID, u1.ID)

	// 3. Simulate process restart by instantiating repo 2 (loads from disk state ./data/anarva_cp_users.json)
	repo2 := newMemUserRepo()
	u2, err2 := repo2.GetByEmail(ctx, testEmail)
	require.NoError(t, err2)
	require.NotNil(t, u2)

	assert.Equal(t, testUserID, u2.ID)
	assert.Equal(t, testEmail, u2.Email)
	assert.Equal(t, "QA Persistence User", u2.FullName)
}
