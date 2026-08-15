package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/reliability/usecase"
)

type ReliabilityHandler struct {
	reliabilityUC *usecase.ReliabilityUseCase
}

func NewReliabilityHandler(reliabilityUC *usecase.ReliabilityUseCase) *ReliabilityHandler {
	return &ReliabilityHandler{reliabilityUC: reliabilityUC}
}

func (h *ReliabilityHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/operations/", h.HandleGetOperation)
	mux.HandleFunc("/api/v1/audit", h.HandleListAuditLogs)
}

func (h *ReliabilityHandler) HandleGetOperation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":{"code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}}`, http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] == "" {
		http.Error(w, `{"error":{"code":"INVALID_INPUT","message":"Operation ID required"}}`, http.StatusBadRequest)
		return
	}
	opID := parts[4]

	op, err := h.reliabilityUC.GetOperation("org-default", opID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":      "RESOURCE_NOT_FOUND",
				"message":   err.Error(),
				"requestId": "req_op_404",
			},
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      op,
		"requestId": op.RequestID,
	})
}

func (h *ReliabilityHandler) HandleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":{"code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}}`, http.StatusMethodNotAllowed)
		return
	}

	projID := r.URL.Query().Get("projectId")
	events := h.reliabilityUC.ListAuditEvents("org-default", projID)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      events,
		"requestId": "req_audit_list",
	})
}
