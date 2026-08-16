package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/reliability/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/reliability/usecase"
)

type ReliabilityHandler struct {
	reliabilityUC *usecase.ReliabilityUseCase
}

func NewReliabilityHandler(reliabilityUC *usecase.ReliabilityUseCase) *ReliabilityHandler {
	return &ReliabilityHandler{reliabilityUC: reliabilityUC}
}

func (h *ReliabilityHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/operations", h.HandleOperations)
	mux.HandleFunc("/api/v1/operations/", h.HandleOperationSubroutes)
	mux.HandleFunc("/api/v1/audit", h.HandleListAuditLogs)
}

func (h *ReliabilityHandler) HandleOperations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := getRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", reqID, "")
		return
	}

	orgID := r.URL.Query().Get("organizationId")
	if orgID == "" {
		orgID = "org-default"
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	opType := r.URL.Query().Get("type")
	if opType == "" {
		opType = r.URL.Query().Get("operationType")
	}

	var createdAfter, createdBefore time.Time
	if ca := r.URL.Query().Get("createdAfter"); ca != "" {
		createdAfter, _ = time.Parse(time.RFC3339, ca)
	}
	if cb := r.URL.Query().Get("createdBefore"); cb != "" {
		createdBefore, _ = time.Parse(time.RFC3339, cb)
	}

	filters := repository.OperationQueryFilters{
		Status:        r.URL.Query().Get("status"),
		ResourceID:    r.URL.Query().Get("resourceId"),
		OperationType: opType,
		ProjectID:     r.URL.Query().Get("projectId"),
		CreatedAfter:  createdAfter,
		CreatedBefore: createdBefore,
		Page:          page,
		PageSize:      pageSize,
	}

	ops, total, err := h.reliabilityUC.ListOperations(r.Context(), orgID, filters)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to list control-plane operations", reqID, "")
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": ops,
		"meta": map[string]interface{}{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
			"count":    len(ops),
		},
		"requestId": reqID,
	})
}

func (h *ReliabilityHandler) HandleOperationSubroutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := getRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", reqID, "")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/operations/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		sendError(w, http.StatusBadRequest, "INVALID_INPUT", "Operation ID required", reqID, "")
		return
	}

	// GET /api/v1/operations/summary
	if parts[0] == "summary" {
		orgID := r.URL.Query().Get("organizationId")
		if orgID == "" {
			orgID = "org-default"
		}
		summary := h.reliabilityUC.GetOperationsSummary(r.Context(), orgID)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data":      summary,
			"requestId": reqID,
		})
		return
	}

	opID := parts[0]
	orgID := r.URL.Query().Get("organizationId")
	if orgID == "" {
		orgID = "org-default"
	}

	op, err := h.reliabilityUC.GetOperation(orgID, opID)
	if err != nil {
		sendError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Operation not found", reqID, opID)
		return
	}

	// GET /api/v1/operations/{id}/timeline
	if len(parts) > 1 && parts[1] == "timeline" {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data":      op.Timeline,
			"operationId": op.ID,
			"requestId": reqID,
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      op,
		"requestId": reqID,
	})
}

func (h *ReliabilityHandler) HandleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := getRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", reqID, "")
		return
	}

	orgID := r.URL.Query().Get("organizationId")
	if orgID == "" {
		orgID = "org-default"
	}
	projID := r.URL.Query().Get("projectId")
	events := h.reliabilityUC.ListAuditEvents(orgID, projID)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      events,
		"requestId": reqID,
	})
}

func sendError(w http.ResponseWriter, statusCode int, code, msg, reqID, opID string) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":        code,
			"message":     msg,
			"requestId":   reqID,
			"operationId": opID,
		},
	})
}

func getRequestID(r *http.Request) string {
	if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
		return reqID
	}
	if ctxReqID, ok := r.Context().Value("requestID").(string); ok && ctxReqID != "" {
		return ctxReqID
	}
	return "req-" + time.Now().Format("20060102150405")
}
