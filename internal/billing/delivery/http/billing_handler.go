package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	billUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/billing/usecase"
)

type BillingHandler struct {
	bUC    *billUsecase.BillingUseCase
	stream *activity.Stream
}

func NewBillingHandler(bUC *billUsecase.BillingUseCase, stream *activity.Stream) *BillingHandler {
	return &BillingHandler{
		bUC:    bUC,
		stream: stream,
	}
}

func (h *BillingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/billing/account", h.handleAccount)
	mux.HandleFunc("/api/v1/billing/usage", h.handleUsage)
	mux.HandleFunc("/api/v1/billing/quotas", h.handleQuotas)
	mux.HandleFunc("/api/v1/billing/pricing", h.handlePricing)
	mux.HandleFunc("/api/v1/billing/invoices", h.handleInvoices)
	mux.HandleFunc("/api/v1/billing/cost-estimates", h.handleCostEstimates)
}

func (h *BillingHandler) setStandardHeaders(w http.ResponseWriter, r *http.Request) string {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = fmt.Sprintf("req_%d", time.Now().UnixNano()/1e6)
	}
	w.Header().Set("X-Request-ID", reqID)
	w.Header().Set("Content-Type", "application/json")
	return reqID
}

func (h *BillingHandler) respondSuccess(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	reqID := h.setStandardHeaders(w, r)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": data,
		"meta": map[string]string{
			"requestId": reqID,
		},
	})
}

func (h *BillingHandler) respondError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	reqID := h.setStandardHeaders(w, r)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":      code,
			"message":   message,
			"requestId": reqID,
		},
	})
}

func (h *BillingHandler) handleAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	account, profile := h.bUC.GetBillingAccount(r.Context(), "org-default")
	h.respondSuccess(w, r, http.StatusOK, map[string]interface{}{
		"account": account,
		"profile": profile,
	})
}

func (h *BillingHandler) handleUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	projectID := r.URL.Query().Get("projectId")
	usage := h.bUC.ListUsage(r.Context(), projectID)
	h.respondSuccess(w, r, http.StatusOK, usage)
}

func (h *BillingHandler) handleQuotas(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		projectID := r.URL.Query().Get("projectId")
		quotas := h.bUC.ListQuotas(r.Context(), projectID)
		h.respondSuccess(w, r, http.StatusOK, quotas)

	case http.MethodPost:
		var payload struct {
			OrganizationID  string  `json:"organizationId"`
			ProjectID       string  `json:"projectId"`
			ResourceType    string  `json:"resourceType"`
			Metric          string  `json:"metric"`
			RequestedAmount float64 `json:"requestedAmount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.respondError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
			return
		}

		if payload.OrganizationID == "" {
			payload.OrganizationID = "org-default"
		}
		if payload.ProjectID == "" {
			payload.ProjectID = "proj-default"
		}

		quota, err := h.bUC.ReserveQuota(r.Context(), payload.OrganizationID, payload.ProjectID, payload.ResourceType, payload.Metric, payload.RequestedAmount)
		if err != nil {
			h.respondError(w, r, http.StatusForbidden, "QUOTA_EXCEEDED", err.Error())
			return
		}

		h.respondSuccess(w, r, http.StatusOK, quota)

	default:
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

func (h *BillingHandler) handlePricing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	plans, components := h.bUC.ListPricingPlans(r.Context())
	h.respondSuccess(w, r, http.StatusOK, map[string]interface{}{
		"plans":      plans,
		"components": components,
	})
}

func (h *BillingHandler) handleInvoices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	invoices := h.bUC.ListInvoices(r.Context(), "org-default")
	h.respondSuccess(w, r, http.StatusOK, invoices)
}

func (h *BillingHandler) handleCostEstimates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var payload struct {
		ResourceType  string  `json:"resourceType"`
		Provider      string  `json:"provider"`
		ACU           float64 `json:"acu"`
		ExpectedHours float64 `json:"expectedHours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	est, err := h.bUC.CalculateCostEstimate(r.Context(), payload.ResourceType, payload.Provider, payload.ACU, payload.ExpectedHours)
	if err != nil {
		h.respondError(w, r, http.StatusBadRequest, "ESTIMATE_FAILED", err.Error())
		return
	}

	h.respondSuccess(w, r, http.StatusOK, est)
}
