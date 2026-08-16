package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
)

type SecurityStatusHandler struct {
	secService *security.SecurityService
}

func NewSecurityStatusHandler(secSvc *security.SecurityService) *SecurityStatusHandler {
	return &SecurityStatusHandler{secService: secSvc}
}

func (h *SecurityStatusHandler) GetSecurityStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := getRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	if r.Method != http.MethodGet {
		sendSecurityError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", reqID)
		return
	}

	// Verify Admin Role from server-side context if available
	role, ok := r.Context().Value("role").(string)
	if ok && role != "ADMIN" && role != "OWNER" && role != "AUDITOR" {
		sendSecurityError(w, http.StatusForbidden, "PERMISSION_DENIED", "Access denied: Administrative privileges required to access Security Status API", reqID)
		return
	}

	if h.secService == nil {
		h.secService = security.NewSecurityService(nil, security.NewSecurityEventService())
	}

	resp := h.secService.EvaluateSecurityStatus(r.Context(), reqID)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *SecurityStatusHandler) GetSecurityEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := getRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	if r.Method != http.MethodGet {
		sendSecurityError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", reqID)
		return
	}

	orgID := r.URL.Query().Get("organizationId")
	if orgID == "" {
		orgID = "org-default"
	}

	var events []*security.SecurityEvent
	if h.secService != nil && h.secService.GetEventService() != nil {
		events = h.secService.GetEventService().ListEvents(orgID)
	} else {
		events = []*security.SecurityEvent{}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      events,
		"requestId": reqID,
	})
}

func sendSecurityError(w http.ResponseWriter, statusCode int, code, msg, reqID string) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":      code,
			"message":   msg,
			"requestId": reqID,
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
	return "req-sec-" + time.Now().Format("20060102150405")
}
