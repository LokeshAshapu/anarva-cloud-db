package domain

import (
	"fmt"
	"sync/atomic"
	"time"
)

type FeedbackStatus string

const (
	StatusNew        FeedbackStatus = "NEW"
	StatusReviewing  FeedbackStatus = "REVIEWING"
	StatusPlanned    FeedbackStatus = "PLANNED"
	StatusInProgress FeedbackStatus = "IN_PROGRESS"
	StatusResolved   FeedbackStatus = "RESOLVED"
	StatusClosed     FeedbackStatus = "CLOSED"
)

type FeedbackCategory string

const (
	CategoryGeneral        FeedbackCategory = "GENERAL"
	CategoryBugReport      FeedbackCategory = "BUG_REPORT"
	CategoryFeatureRequest FeedbackCategory = "FEATURE_REQUEST"
	CategoryPerformance    FeedbackCategory = "PERFORMANCE"
	CategoryUsability      FeedbackCategory = "USABILITY"
)

type AnarvaFeedback struct {
	FeedbackID     string           `json:"feedbackId"`
	ID             string           `json:"id"`
	UserID         string           `json:"userId"`
	UserEmail      string           `json:"userEmail"`
	UserName       string           `json:"userName,omitempty"`
	OrganizationID string           `json:"organizationId"`
	ProjectID      string           `json:"projectId,omitempty"`
	Rating         int              `json:"rating"` // 1 to 5
	Category       FeedbackCategory `json:"category"`
	Subject        string           `json:"subject"`
	Message        string           `json:"message"`
	Status         FeedbackStatus   `json:"status"`
	TargetEmail    string           `json:"targetEmail"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
	RequestID      string           `json:"requestId"`
}

type FeedbackAnalytics struct {
	TotalFeedback      int                    `json:"totalFeedback"`
	AverageRating      float64                `json:"averageRating"`
	RatingDistribution map[int]int            `json:"ratingDistribution"`
	StatusCounts       map[FeedbackStatus]int `json:"statusCounts"`
}

type FeedbackQuery struct {
	OrganizationID string         `json:"organizationId"`
	ProjectID      string         `json:"projectId"`
	Status         FeedbackStatus `json:"status"`
	MinRating      int            `json:"minRating"`
	Page           int            `json:"page"`
	PageSize       int            `json:"pageSize"`
}

type FeedbackPaginatedResult struct {
	Items      []*AnarvaFeedback `json:"items"`
	TotalCount int               `json:"totalCount"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int               `json:"totalPages"`
}

var fbCounter uint64

func FormatFeedbackID() string {
	seq := atomic.AddUint64(&fbCounter, 1)
	return fmt.Sprintf("fb-%d-%d", time.Now().UnixNano()/1e6, seq)
}
