package domain

import (
	"fmt"
	"time"
)

type FeedbackCategory string

const (
	CategoryGeneral        FeedbackCategory = "GENERAL"
	CategoryBugReport      FeedbackCategory = "BUG_REPORT"
	CategoryFeatureRequest FeedbackCategory = "FEATURE_REQUEST"
	CategoryPerformance    FeedbackCategory = "PERFORMANCE"
	CategoryUsability      FeedbackCategory = "USABILITY"
)

type FeedbackSubmission struct {
	ID          string           `json:"id"`
	UserEmail   string           `json:"userEmail"`
	UserName    string           `json:"userName,omitempty"`
	Category    FeedbackCategory `json:"category"`
	Rating      int              `json:"rating"` // 1 to 5
	Subject     string           `json:"subject"`
	Message     string           `json:"message"`
	TargetEmail string           `json:"targetEmail"`
	CreatedAt   time.Time        `json:"createdAt"`
	RequestID   string           `json:"requestId"`
	Status      string           `json:"status"` // DISPATCHED, DELIVERED
}

func FormatFeedbackID() string {
	return fmt.Sprintf("fb-%d", time.Now().UnixNano()/1e6)
}
