package http

import (
	"encoding/json"
	"fmt"
	"net/http"
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
}

func (h *FeedbackHandler) HandleFeedback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var req domain.FeedbackSubmission
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

		res, err := h.feedbackUC.SubmitFeedback(r.Context(), req)
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
			"message":   fmt.Sprintf("Feedback successfully submitted and dispatched to %s", usecase.TargetRecipientEmail),
			"requestId": reqID,
		})
		return
	}

	if r.Method == http.MethodGet {
		list := h.feedbackUC.ListSubmissions()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data":      list,
			"target":    usecase.TargetRecipientEmail,
			"requestId": "req_fb_list",
		})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}
