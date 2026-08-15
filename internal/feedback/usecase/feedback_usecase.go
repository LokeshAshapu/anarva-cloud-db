package usecase

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/feedback/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

const TargetRecipientEmail = "23w61a0506@gmail.com"

type FeedbackUseCase struct {
	mu           sync.RWMutex
	submissions  map[string]*domain.FeedbackSubmission
	targetEmail  string
}

func NewFeedbackUseCase() *FeedbackUseCase {
	return &FeedbackUseCase{
		submissions: make(map[string]*domain.FeedbackSubmission),
		targetEmail: TargetRecipientEmail,
	}
}

func (uc *FeedbackUseCase) SubmitFeedback(ctx context.Context, input domain.FeedbackSubmission) (*domain.FeedbackSubmission, error) {
	if input.Message == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "Feedback message cannot be empty")
	}
	if input.UserEmail == "" {
		input.UserEmail = "anonymous-user@anarva.io"
	}
	if input.Category == "" {
		input.Category = domain.CategoryGeneral
	}
	if input.Rating < 1 || input.Rating > 5 {
		input.Rating = 5
	}
	if input.Subject == "" {
		input.Subject = fmt.Sprintf("Console Feedback from %s", input.UserEmail)
	}

	now := time.Now()
	submission := &domain.FeedbackSubmission{
		ID:          domain.FormatFeedbackID(),
		UserEmail:   input.UserEmail,
		UserName:    input.UserName,
		Category:    input.Category,
		Rating:      input.Rating,
		Subject:     input.Subject,
		Message:     input.Message,
		TargetEmail: uc.targetEmail,
		CreatedAt:   now,
		RequestID:   input.RequestID,
		Status:      "DISPATCHED",
	}

	uc.mu.Lock()
	uc.submissions[submission.ID] = submission
	uc.mu.Unlock()

	// Dispatch Email to target recipient (23w61a0506@gmail.com)
	go uc.dispatchEmailNotification(submission)

	return submission, nil
}

func (uc *FeedbackUseCase) dispatchEmailNotification(fb *domain.FeedbackSubmission) {
	subject := fmt.Sprintf("[Anarva Cloud Feedback] %s: %s", fb.Category, fb.Subject)
	body := fmt.Sprintf(`To: %s
Subject: %s
Content-Type: text/plain; charset=UTF-8

==================================================
ANARVA CLOUD CONSOLE — USER FEEDBACK REPORT
==================================================

Target Recipient: %s
Submitted At:     %s
Request ID:       %s

SUBMITTER DETAILS:
--------------------------------------------------
Email:            %s
Name:             %s

FEEDBACK CONTENT:
--------------------------------------------------
Category:         %s
Rating:           %d / 5 Stars
Subject:          %s

MESSAGE:
--------------------------------------------------
%s

==================================================
Sent automatically by Anarva Cloud Console Notification Engine.
`, fb.TargetEmail, subject, fb.TargetEmail, fb.CreatedAt.Format(time.RFC1123), fb.RequestID, fb.UserEmail, fb.UserName, fb.Category, fb.Rating, fb.Subject, fb.Message)

	log.Printf("[FEEDBACK_EMAIL_DISPATCH] Dispatching feedback email to %s for Submission ID %s (Submitter: %s)", fb.TargetEmail, fb.ID, fb.UserEmail)
	log.Printf("[FEEDBACK_EMAIL_BODY]\n%s", body)

	// SMTP Dispatch if SMTP credentials configured in environment
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

func (uc *FeedbackUseCase) ListSubmissions() []*domain.FeedbackSubmission {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	var list []*domain.FeedbackSubmission
	for _, sub := range uc.submissions {
		list = append(list, sub)
	}
	return list
}
