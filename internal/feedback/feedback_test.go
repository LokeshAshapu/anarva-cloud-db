package feedback

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/feedback/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/feedback/usecase"
)

func TestFeedback_SubmitAndDispatchToTargetEmail(t *testing.T) {
	uc := usecase.NewFeedbackUseCase()

	input := domain.FeedbackSubmission{
		UserEmail: "developer@anarva.io",
		UserName:  "Lead Developer",
		Category:  domain.CategoryFeatureRequest,
		Rating:    5,
		Subject:   "Awesome Platform!",
		Message:   "Please add auto-scaling bounds for Database ACU capacity.",
		RequestID: "req_test_fb_01",
	}

	result, err := uc.SubmitFeedback(context.Background(), input)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "23w61a0506@gmail.com", result.TargetEmail)
	assert.Equal(t, "developer@anarva.io", result.UserEmail)
	assert.Equal(t, domain.CategoryFeatureRequest, result.Category)
	assert.Equal(t, 5, result.Rating)

	list := uc.ListSubmissions()
	assert.Len(t, list, 1)
	assert.Equal(t, "23w61a0506@gmail.com", list[0].TargetEmail)
}
