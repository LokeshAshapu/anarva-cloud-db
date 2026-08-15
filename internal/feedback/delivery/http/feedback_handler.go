package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/feedback/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/feedback/usecase"
)

type FeedbackHandler struct {
	feedbackUC *usecase.FeedbackUseCase
}

func NewFeedbackHandler(feedbackUC *usecase.FeedbackUseCase) *FeedbackHandler {
	return &FeedbackHandler{feedbackUC: feedbackUC}
}

func (h *FeedbackHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/feedback", h.HandleFeedback)
	mux.HandleFunc("/api/v1/feedback/", h.HandleFeedbackItem)
	mux.HandleFunc("/api/v1/feedback/analytics", h.HandleFeedbackAnalytics)
}

func (h *FeedbackHandler) HandleFeedback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var req domain.AnarvaFeedback
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code":    "INVALID_JSON",
					"message": "Failed to parse feedback payload",
				},
			})
			return
		}

		reqID := fmt.Sprintf("req_fb_%d", time.Now().UnixNano()/1e6)
		req.RequestID = reqID

		res, err := h.feedbackUC.CreateFeedback(r.Context(), req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code":    "SUBMISSION_FAILED",
					"message": err.Error(),
				},
				"requestId": reqID,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data":      res,
			"message":   fmt.Sprintf("Thank you for your feedback. Feedback ID: %s", res.FeedbackID),
			"requestId": reqID,
		})
		return
	}

	if r.Method == http.MethodGet {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		minRating, _ := strconv.Atoi(r.URL.Query().Get("minRating"))
		status := domain.FeedbackStatus(r.URL.Query().Get("status"))
		projID := r.URL.Query().Get("projectId")

		query := domain.FeedbackQuery{
			OrganizationID: "org-default",
			ProjectID:      projID,
			Status:         status,
			MinRating:      minRating,
			Page:           page,
			PageSize:       pageSize,
		}

		res, err := h.feedbackUC.ListFeedback(r.Context(), query)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{"code": "SERVER_ERROR", "message": err.Error()},
			})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data":      res,
			"requestId": "req_fb_list",
		})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (h *FeedbackHandler) HandleFeedbackItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, `{"error":{"code":"INVALID_PATH","message":"Invalid path"}}`, http.StatusBadRequest)
		return
	}

	feedbackID := parts[3]
	if feedbackID == "analytics" {
		h.HandleFeedbackAnalytics(w, r)
		return
	}

	if r.Method == http.MethodGet {
		fb, err := h.feedbackUC.GetFeedback(r.Context(), "org-default", feedbackID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{"code": "NOT_FOUND", "message": err.Error()},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data":      fb,
			"requestId": "req_fb_get",
		})
		return
	}

	if r.Method == http.MethodPatch && len(parts) >= 5 && parts[4] == "status" {
		var req struct {
			Status domain.FeedbackStatus `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{"code": "INVALID_JSON", "message": "Invalid status payload"},
			})
			return
		}

		updated, err := h.feedbackUC.UpdateFeedbackStatus(r.Context(), "org-default", feedbackID, req.Status)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{"code": "UPDATE_FAILED", "message": err.Error()},
			})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data":      updated,
			"message":   fmt.Sprintf("Feedback status updated to %s", updated.Status),
			"requestId": "req_fb_status_update",
		})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (h *FeedbackHandler) HandleFeedbackAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":{"code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}}`, http.StatusMethodNotAllowed)
		return
	}

	analytics, err := h.feedbackUC.GetFeedbackAnalytics(r.Context(), "org-default")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"code": "SERVER_ERROR", "message": err.Error()},
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      analytics,
		"requestId": "req_fb_analytics",
	})
}
