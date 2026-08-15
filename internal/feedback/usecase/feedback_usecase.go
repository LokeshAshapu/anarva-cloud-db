package usecase

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/feedback/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

const TargetRecipientEmail = "23w61a0506@gmail.com"

type FeedbackUseCase struct {
	mu          sync.RWMutex
	feedbacks   map[string]*domain.AnarvaFeedback
	targetEmail string
	auditLogger func(orgID, projID, actorType, actorID, action, resType, resID, opID, reqID string, metadata map[string]string)
}

func NewFeedbackUseCase(auditLogger func(orgID, projID, actorType, actorID, action, resType, resID, opID, reqID string, metadata map[string]string)) *FeedbackUseCase {
	uc := &FeedbackUseCase{
		feedbacks:   make(map[string]*domain.AnarvaFeedback),
		targetEmail: TargetRecipientEmail,
		auditLogger: auditLogger,
	}
	uc.seedDefaultFeedback()
	return uc
}

func (uc *FeedbackUseCase) seedDefaultFeedback() {
	now := time.Now()
	fbs := []*domain.AnarvaFeedback{
		{
			FeedbackID:     "fb-101",
			ID:             "fb-101",
			UserID:         "usr-operator-01",
			UserEmail:      "lokeshashapu@gmail.com",
			UserName:       "Cloud Operator",
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Rating:         5,
			Category:       domain.CategoryGeneral,
			Subject:        "Database provisioning speed is excellent",
			Message:        "The RDS PostgreSQL cluster provisioning and failover orchestration experience is extremely smooth.",
			Status:         domain.StatusNew,
			TargetEmail:    TargetRecipientEmail,
			CreatedAt:      now.Add(-2 * time.Hour),
			UpdatedAt:      now.Add(-2 * time.Hour),
			RequestID:      "req_seed_fb_101",
		},
		{
			FeedbackID:     "fb-102",
			ID:             "fb-102",
			UserID:         "usr-dev-02",
			UserEmail:      "developer@anarva.io",
			UserName:       "Senior Engineer",
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Rating:         4,
			Category:       domain.CategoryFeatureRequest,
			Subject:        "Add auto-scaling bounds for Database ACU capacity",
			Message:        "Would love to see automated ACU scaling policies based on CloudWatch CPU metrics.",
			Status:         domain.StatusReviewing,
			TargetEmail:    TargetRecipientEmail,
			CreatedAt:      now.Add(-24 * time.Hour),
			UpdatedAt:      now.Add(-12 * time.Hour),
			RequestID:      "req_seed_fb_102",
		},
	}

	for _, fb := range fbs {
		uc.feedbacks[fb.FeedbackID] = fb
	}
}

func (uc *FeedbackUseCase) CreateFeedback(ctx context.Context, input domain.AnarvaFeedback) (*domain.AnarvaFeedback, error) {
	// Validation
	if strings.TrimSpace(input.Message) == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "Feedback message cannot be empty")
	}
	if len(input.Message) > 5000 {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "Feedback message exceeds maximum allowed length of 5000 characters")
	}
	if len(input.Subject) > 250 {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "Feedback subject exceeds maximum allowed length of 250 characters")
	}
	if input.Rating < 1 || input.Rating > 5 {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "Rating must be between 1 and 5 stars")
	}

	// Automatic Server-Side Context Resolution
	if input.OrganizationID == "" {
		input.OrganizationID = "org-default"
	}
	if input.ProjectID == "" {
		input.ProjectID = "proj-default"
	}
	if input.UserID == "" {
		input.UserID = "usr-current-session"
	}
	if input.UserEmail == "" {
		input.UserEmail = "user@anarva.io"
	}
	if input.Category == "" {
		input.Category = domain.CategoryGeneral
	}
	if input.Subject == "" {
		input.Subject = fmt.Sprintf("Feedback from %s", input.UserEmail)
	}

	now := time.Now()
	fbID := domain.FormatFeedbackID()

	fb := &domain.AnarvaFeedback{
		FeedbackID:     fbID,
		ID:             fbID,
		UserID:         input.UserID,
		UserEmail:      input.UserEmail,
		UserName:       input.UserName,
		OrganizationID: input.OrganizationID,
		ProjectID:      input.ProjectID,
		Rating:         input.Rating,
		Category:       input.Category,
		Subject:        input.Subject,
		Message:        input.Message,
		Status:         domain.StatusNew,
		TargetEmail:    uc.targetEmail,
		CreatedAt:      now,
		UpdatedAt:      now,
		RequestID:      input.RequestID,
	}

	uc.mu.Lock()
	uc.feedbacks[fb.FeedbackID] = fb
	uc.mu.Unlock()

	// Audit Integration
	if uc.auditLogger != nil {
		uc.auditLogger(
			fb.OrganizationID,
			fb.ProjectID,
			"USER",
			fb.UserID,
			"FEEDBACK_SUBMITTED",
			"FEEDBACK",
			fb.FeedbackID,
			"",
			fb.RequestID,
			map[string]string{
				"rating":   fmt.Sprintf("%d", fb.Rating),
				"category": string(fb.Category),
				"status":   string(fb.Status),
			},
		)
	}

	// Email Dispatch to 23w61a0506@gmail.com
	go uc.dispatchEmailNotification(fb)

	return fb, nil
}

func (uc *FeedbackUseCase) GetFeedback(ctx context.Context, orgID, feedbackID string) (*domain.AnarvaFeedback, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	fb, exists := uc.feedbacks[feedbackID]
	if !exists {
		return nil, appErrors.New(appErrors.CodeNotFound, "Feedback record not found")
	}

	// Tenant Isolation Verification
	if orgID != "" && fb.OrganizationID != orgID {
		return nil, appErrors.New(appErrors.CodeForbidden, "TENANT_ISOLATION_VIOLATION: Unauthorized access to organization feedback")
	}

	return fb, nil
}

func (uc *FeedbackUseCase) ListFeedback(ctx context.Context, query domain.FeedbackQuery) (*domain.FeedbackPaginatedResult, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	var filtered []*domain.AnarvaFeedback
	for _, fb := range uc.feedbacks {
		// Tenant Isolation Enforcement
		if query.OrganizationID != "" && fb.OrganizationID != query.OrganizationID {
			continue
		}
		if query.ProjectID != "" && fb.ProjectID != query.ProjectID {
			continue
		}
		if query.Status != "" && fb.Status != query.Status {
			continue
		}
		if query.MinRating > 0 && fb.Rating < query.MinRating {
			continue
		}
		filtered = append(filtered, fb)
	}

	totalCount := len(filtered)
	totalPages := (totalCount + query.PageSize - 1) / query.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	startIndex := (query.Page - 1) * query.PageSize
	endIndex := startIndex + query.PageSize

	if startIndex >= totalCount {
		return &domain.FeedbackPaginatedResult{
			Items:      []*domain.AnarvaFeedback{},
			TotalCount: totalCount,
			Page:       query.Page,
			PageSize:   query.PageSize,
			TotalPages: totalPages,
		}, nil
	}

	if endIndex > totalCount {
		endIndex = totalCount
	}

	return &domain.FeedbackPaginatedResult{
		Items:      filtered[startIndex:endIndex],
		TotalCount: totalCount,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (uc *FeedbackUseCase) UpdateFeedbackStatus(ctx context.Context, orgID, feedbackID string, newStatus domain.FeedbackStatus) (*domain.AnarvaFeedback, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	fb, exists := uc.feedbacks[feedbackID]
	if !exists {
		return nil, appErrors.New(appErrors.CodeNotFound, "Feedback record not found")
	}

	// Tenant Isolation Verification
	if orgID != "" && fb.OrganizationID != orgID {
		return nil, appErrors.New(appErrors.CodeForbidden, "TENANT_ISOLATION_VIOLATION: Unauthorized access to organization feedback")
	}

	// Validate Status Transition
	validStatuses := map[domain.FeedbackStatus]bool{
		domain.StatusNew:        true,
		domain.StatusReviewing:  true,
		domain.StatusPlanned:    true,
		domain.StatusInProgress: true,
		domain.StatusResolved:   true,
		domain.StatusClosed:     true,
	}

	if !validStatuses[newStatus] {
		return nil, appErrors.New(appErrors.CodeInvalidInput, fmt.Sprintf("Invalid feedback status '%s'", newStatus))
	}

	oldStatus := fb.Status
	fb.Status = newStatus
	fb.UpdatedAt = time.Now()

	// Audit Integration
	if uc.auditLogger != nil {
		uc.auditLogger(
			fb.OrganizationID,
			fb.ProjectID,
			"USER",
			"usr-admin",
			"FEEDBACK_STATUS_UPDATED",
			"FEEDBACK",
			fb.FeedbackID,
			"",
			fb.RequestID,
			map[string]string{
				"oldStatus": string(oldStatus),
				"newStatus": string(newStatus),
			},
		)
	}

	return fb, nil
}

func (uc *FeedbackUseCase) GetFeedbackAnalytics(ctx context.Context, orgID string) (*domain.FeedbackAnalytics, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	dist := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	statuses := map[domain.FeedbackStatus]int{
		domain.StatusNew:        0,
		domain.StatusReviewing:  0,
		domain.StatusPlanned:    0,
		domain.StatusInProgress: 0,
		domain.StatusResolved:   0,
		domain.StatusClosed:     0,
	}

	total := 0
	sumRating := 0

	for _, fb := range uc.feedbacks {
		if orgID != "" && fb.OrganizationID != orgID {
			continue
		}
		total++
		sumRating += fb.Rating
		dist[fb.Rating]++
		statuses[fb.Status]++
	}

	avg := 0.0
	if total > 0 {
		avg = float64(sumRating) / float64(total)
	}

	return &domain.FeedbackAnalytics{
		TotalFeedback:      total,
		AverageRating:      avg,
		RatingDistribution: dist,
		StatusCounts:       statuses,
	}, nil
}

func (uc *FeedbackUseCase) dispatchEmailNotification(fb *domain.AnarvaFeedback) {
	subject := fmt.Sprintf("[Anarva Cloud Feedback] %s: %s", fb.Category, fb.Subject)
	body := fmt.Sprintf(`To: %s
Subject: %s
Content-Type: text/plain; charset=UTF-8

==================================================
ANARVA CLOUD CONSOLE — USER FEEDBACK REPORT
==================================================

Feedback ID:      %s
Target Recipient: %s
Submitted At:     %s
Request ID:       %s

SUBMITTER DETAILS:
--------------------------------------------------
User ID:          %s
Email:            %s
Organization:     %s
Project:          %s

FEEDBACK CONTENT:
--------------------------------------------------
Category:         %s
Rating:           %d / 5 Stars
Status:           %s
Subject:          %s

MESSAGE:
--------------------------------------------------
%s

==================================================
Sent automatically by Anarva Cloud Console Notification Engine.
`, fb.TargetEmail, subject, fb.FeedbackID, fb.TargetEmail, fb.CreatedAt.Format(time.RFC1123), fb.RequestID, fb.UserID, fb.UserEmail, fb.OrganizationID, fb.ProjectID, fb.Category, fb.Rating, fb.Status, fb.Subject, fb.Message)

	log.Printf("[FEEDBACK_EMAIL_DISPATCH] Dispatching feedback email to %s for Submission ID %s (Submitter: %s)", fb.TargetEmail, fb.FeedbackID, fb.UserEmail)
	log.Printf("[FEEDBACK_EMAIL_BODY]\n%s", body)

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	if smtpHost != "" && smtpUser != "" {
		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		to := []string{fb.TargetEmail}
		err := smtp.SendMail(fmt.Sprintf("%s:%s", smtpHost, smtpPort), auth, smtpUser, to, []byte(body))
		if err != nil {
			log.Printf("[FEEDBACK_EMAIL_ERROR] Failed to send SMTP email to %s: %v", fb.TargetEmail, err)
		} else {
			log.Printf("[FEEDBACK_EMAIL_SUCCESS] Successfully delivered SMTP email to %s", fb.TargetEmail)
		}
	}
}
