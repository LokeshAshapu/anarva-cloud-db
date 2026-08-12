package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	devDomain "github.com/anarva-cloud/anarva-cloud-db/internal/developer/domain"
	devUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/developer/usecase"
	whUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/webhook/usecase"
)

type DeveloperHandler struct {
	devUC  *devUsecase.DeveloperUseCase
	whUC   *whUsecase.WebhookUseCase
	stream *activity.Stream
}

func NewDeveloperHandler(devUC *devUsecase.DeveloperUseCase, whUC *whUsecase.WebhookUseCase, stream *activity.Stream) *DeveloperHandler {
	return &DeveloperHandler{
		devUC:  devUC,
		whUC:   whUC,
		stream: stream,
	}
}

func (h *DeveloperHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/developer/keys", h.handleAPIKeys)
	mux.HandleFunc("/api/v1/developer/keys/", h.handleAPIKeySubroutes)
	mux.HandleFunc("/api/v1/developer/service-accounts", h.handleServiceAccounts)
	mux.HandleFunc("/api/v1/developer/webhooks", h.handleWebhooks)
	mux.HandleFunc("/api/v1/developer/playground/execute", h.handlePlayground)
	mux.HandleFunc("/api/v1/developer/usage", h.handleUsage)
	mux.HandleFunc("/api/v1/developer/openapi", h.handleOpenAPI)
}

func (h *DeveloperHandler) setStandardHeaders(w http.ResponseWriter, r *http.Request) string {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = fmt.Sprintf("req_%d", time.Now().UnixNano()/1e6)
	}
	w.Header().Set("X-Request-ID", reqID)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("RateLimit-Limit", "100")
	w.Header().Set("RateLimit-Remaining", "99")
	w.Header().Set("RateLimit-Reset", "60")
	return reqID
}

func (h *DeveloperHandler) respondSuccess(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	reqID := h.setStandardHeaders(w, r)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": data,
		"meta": map[string]string{
			"requestId": reqID,
		},
	})
}

func (h *DeveloperHandler) respondError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
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

func (h *DeveloperHandler) handleAPIKeys(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		projectID := r.URL.Query().Get("projectId")
		keys := h.devUC.ListAPIKeys(r.Context(), projectID)
		h.respondSuccess(w, r, http.StatusOK, keys)

	case http.MethodPost:
		var payload struct {
			Name        string   `json:"name"`
			ProjectID   string   `json:"projectId"`
			Permissions []string `json:"permissions"`
			IsLive      bool     `json:"isLive"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.respondError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
			return
		}
		if payload.ProjectID == "" {
			payload.ProjectID = "proj-default"
		}

		key, secret, err := h.devUC.CreateAPIKey(r.Context(), payload.Name, "org-default", payload.ProjectID, "lokeshashapu@gmail.com", payload.Permissions, payload.IsLive)
		if err != nil {
			h.respondError(w, r, http.StatusBadRequest, "CREATE_KEY_FAILED", err.Error())
			return
		}

		if h.stream != nil {
			h.stream.Record(&activity.ActivityEvent{
				OrganizationID: "org-default",
				ProjectID:      payload.ProjectID,
				ActorID:        "lokeshashapu@gmail.com",
				Action:         activity.ActionAPIKeyCreated,
				Metadata:       map[string]string{"keyName": key.Name, "keyPrefix": key.KeyPrefix},
			})
		}

		h.respondSuccess(w, r, http.StatusCreated, map[string]interface{}{
			"apiKey":    key,
			"secretKey": secret, // ONLY SHOWN ONCE DURING CREATION
			"warning":   "Store this secret key securely. It will never be displayed again.",
		})

	default:
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

func (h *DeveloperHandler) handleAPIKeySubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/developer/keys/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "revoke" {
		h.respondError(w, r, http.StatusBadRequest, "INVALID_ROUTE", "Invalid API key route")
		return
	}

	keyID := parts[0]
	if r.Method != http.MethodPost {
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	if err := h.devUC.RevokeAPIKey(r.Context(), keyID); err != nil {
		h.respondError(w, r, http.StatusNotFound, "KEY_NOT_FOUND", err.Error())
		return
	}

	h.respondSuccess(w, r, http.StatusOK, map[string]string{"id": keyID, "status": "REVOKED"})
}

func (h *DeveloperHandler) handleServiceAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		projectID := r.URL.Query().Get("projectId")
		list := h.devUC.ListServiceAccounts(r.Context(), projectID)
		h.respondSuccess(w, r, http.StatusOK, list)

	case http.MethodPost:
		var payload struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Role        string `json:"role"`
			ProjectID   string `json:"projectId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.respondError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
			return
		}
		if payload.ProjectID == "" {
			payload.ProjectID = "proj-default"
		}
		if payload.Role == "" {
			payload.Role = "DEVELOPER"
		}

		sa, err := h.devUC.CreateServiceAccount(r.Context(), payload.Name, payload.Description, payload.Role, "org-default", payload.ProjectID, "lokeshashapu@gmail.com")
		if err != nil {
			h.respondError(w, r, http.StatusBadRequest, "CREATE_SA_FAILED", err.Error())
			return
		}

		h.respondSuccess(w, r, http.StatusCreated, sa)

	default:
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

func (h *DeveloperHandler) handleWebhooks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		projectID := r.URL.Query().Get("projectId")
		endpoints := h.whUC.ListEndpoints(r.Context(), projectID)
		h.respondSuccess(w, r, http.StatusOK, endpoints)

	case http.MethodPost:
		var payload struct {
			URL         string   `json:"url"`
			Description string   `json:"description"`
			ProjectID   string   `json:"projectId"`
			Events      []string `json:"events"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.respondError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
			return
		}
		if payload.ProjectID == "" {
			payload.ProjectID = "proj-default"
		}

		ep, rawSecret, err := h.whUC.CreateEndpoint(r.Context(), payload.URL, payload.Description, "org-default", payload.ProjectID, payload.Events)
		if err != nil {
			h.respondError(w, r, http.StatusBadRequest, "WEBHOOK_CREATION_FAILED", err.Error())
			return
		}

		h.respondSuccess(w, r, http.StatusCreated, map[string]interface{}{
			"endpoint":      ep,
			"signingSecret": rawSecret,
			"warning":       "Store this signing secret securely to verify HMAC-SHA256 headers (X-Anarva-Signature).",
		})

	default:
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

func (h *DeveloperHandler) handlePlayground(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var payload struct {
		Endpoint string                 `json:"endpoint"`
		Method   string                 `json:"method"`
		Headers  map[string]string      `json:"headers"`
		Body     map[string]interface{} `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	// Security: Prevent SSRF by restricting playground to internal /api/v1/ endpoints
	if !strings.HasPrefix(payload.Endpoint, "/api/v1/") {
		h.respondError(w, r, http.StatusForbidden, "FORBIDDEN_ENDPOINT", "Playground requests are restricted exclusively to Anarva Cloud /api/v1/ API endpoints")
		return
	}

	// Execute simulated API response
	h.devUC.RecordUsage(&devDomain.APIUsageRecord{
		ID:             fmt.Sprintf("use-%d", time.Now().UnixNano()/1e6),
		Endpoint:       payload.Endpoint,
		Method:         payload.Method,
		StatusCode:     200,
		ResponseTimeMs: 14.2,
		RequestID:      r.Header.Get("X-Request-ID"),
	})

	h.respondSuccess(w, r, http.StatusOK, map[string]interface{}{
		"status": 200,
		"statusText": "200 OK",
		"payload": map[string]interface{}{
			"data": map[string]interface{}{
				"status": "AVAILABLE",
				"realityLabel": "LOCAL DEVELOPMENT PROVIDER",
				"message": fmt.Sprintf("Playground executed '%s %s' successfully", payload.Method, payload.Endpoint),
			},
		},
	})
}

func (h *DeveloperHandler) handleUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	usage := h.devUC.ListUsage()
	h.respondSuccess(w, r, http.StatusOK, usage)
}

func (h *DeveloperHandler) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	openAPISpec := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]string{
			"title":       "Anarva Cloud Developer API",
			"version":     "1.0.0",
			"description": "Public Developer Infrastructure Control Plane REST API for Anarva Cloud.",
		},
		"paths": map[string]interface{}{
			"/api/v1/compute": map[string]interface{}{
				"get": map[string]string{"summary": "List Anarva Compute Instances"},
				"post": map[string]string{"summary": "Deploy Anarva Compute Instance"},
			},
			"/api/v1/databases": map[string]interface{}{
				"get": map[string]string{"summary": "List Managed PostgreSQL Databases"},
			},
			"/api/v1/storage": map[string]interface{}{
				"get": map[string]string{"summary": "List Object Storage Buckets"},
			},
			"/api/v1/networks": map[string]interface{}{
				"get": map[string]string{"summary": "List VPC Networks & Subnets"},
			},
			"/api/v1/provisioning/plan": map[string]interface{}{
				"post": map[string]string{"summary": "Generate Infrastructure Provisioning Plan Preview"},
			},
		},
	}
	h.respondSuccess(w, r, http.StatusOK, openAPISpec)
}
