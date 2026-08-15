package feedback

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/feedback/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/feedback/usecase"
)

func TestFeedback_CreateValidation(t *testing.T) {
	uc := usecase.NewFeedbackUseCase(nil)
	ctx := context.Background()

	// Empty Message Failure
	_, err := uc.CreateFeedback(ctx, domain.AnarvaFeedback{
		Rating:  5,
		Message: "   ",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	// Invalid Rating Failure (Rating 6)
	_, err = uc.CreateFeedback(ctx, domain.AnarvaFeedback{
		Rating:  6,
		Message: "Great job!",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "between 1 and 5 stars")

	// Valid Creation
	valid, err := uc.CreateFeedback(ctx, domain.AnarvaFeedback{
		Rating:  5,
		Subject: "Awesome UX",
		Message: "Cloud console experience is state of the art.",
	})
	require.NoError(t, err)
	assert.NotNil(t, valid)
	assert.Equal(t, domain.StatusNew, valid.Status)
	assert.Equal(t, "23w61a0506@gmail.com", valid.TargetEmail)
}

func TestFeedback_TenantIsolation(t *testing.T) {
	uc := usecase.NewFeedbackUseCase(nil)
	ctx := context.Background()

	fb, err := uc.CreateFeedback(ctx, domain.AnarvaFeedback{
		OrganizationID: "org-alpha",
		ProjectID:      "proj-alpha",
		Rating:         4,
		Message:        "Org Alpha Feedback",
	})
	require.NoError(t, err)

	// Org Beta attempt to read Org Alpha feedback MUST fail
	_, err = uc.GetFeedback(ctx, "org-beta", fb.FeedbackID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TENANT_ISOLATION_VIOLATION")

	// Org Alpha attempt to read succeeds
	item, err := uc.GetFeedback(ctx, "org-alpha", fb.FeedbackID)
	require.NoError(t, err)
	assert.Equal(t, "org-alpha", item.OrganizationID)
}

func TestFeedback_StatusLifecycle(t *testing.T) {
	uc := usecase.NewFeedbackUseCase(nil)
	ctx := context.Background()

	fb, err := uc.CreateFeedback(ctx, domain.AnarvaFeedback{
		OrganizationID: "org-default",
		Rating:         5,
		Message:        "Test status transitions",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.StatusNew, fb.Status)

	// Update to REVIEWING
	updated, err := uc.UpdateFeedbackStatus(ctx, "org-default", fb.FeedbackID, domain.StatusReviewing)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusReviewing, updated.Status)

	// Update to RESOLVED
	updated2, err := uc.UpdateFeedbackStatus(ctx, "org-default", fb.FeedbackID, domain.StatusResolved)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusResolved, updated2.Status)
}

func TestFeedback_PaginationAndAnalytics(t *testing.T) {
	uc := usecase.NewFeedbackUseCase(nil)
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		_, err := uc.CreateFeedback(ctx, domain.AnarvaFeedback{
			OrganizationID: "org-analytics",
			Rating:         i,
			Message:        "Batch feedback item",
		})
		require.NoError(t, err)
	}

	// Analytics Check
	analytics, err := uc.GetFeedbackAnalytics(ctx, "org-analytics")
	require.NoError(t, err)
	assert.Equal(t, 5, analytics.TotalFeedback)
	assert.Equal(t, 3.0, analytics.AverageRating)
	assert.Equal(t, 1, analytics.RatingDistribution[5])

	// Pagination Check
	res, err := uc.ListFeedback(ctx, domain.FeedbackQuery{
		OrganizationID: "org-analytics",
		Page:           1,
		PageSize:       3,
	})
	require.NoError(t, err)
	assert.Len(t, res.Items, 3)
	assert.Equal(t, 5, res.TotalCount)
	assert.Equal(t, 2, res.TotalPages)
}

func TestFeedback_AuditIntegration(t *testing.T) {
	var recordedAction string
	var recordedResID string

	auditFn := func(orgID, projID, actorType, actorID, action, resType, resID, opID, reqID string, metadata map[string]string) {
		recordedAction = action
		recordedResID = resID
	}

	uc := usecase.NewFeedbackUseCase(auditFn)
	ctx := context.Background()

	fb, err := uc.CreateFeedback(ctx, domain.AnarvaFeedback{
		OrganizationID: "org-default",
		Rating:         5,
		Message:        "Testing audit integration",
	})
	require.NoError(t, err)
	assert.Equal(t, "FEEDBACK_SUBMITTED", recordedAction)
	assert.Equal(t, fb.FeedbackID, recordedResID)

	_, err = uc.UpdateFeedbackStatus(ctx, "org-default", fb.FeedbackID, domain.StatusPlanned)
	require.NoError(t, err)
	assert.Equal(t, "FEEDBACK_STATUS_UPDATED", recordedAction)
}
