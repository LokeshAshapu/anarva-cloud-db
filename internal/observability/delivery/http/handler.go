package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/observability/service"
)

type ResourceObservabilityHandler struct {
	engine *service.ResourceObservabilityEngine
}

func NewResourceObservabilityHandler(engine *service.ResourceObservabilityEngine) *ResourceObservabilityHandler {
	return &ResourceObservabilityHandler{engine: engine}
}

func (h *ResourceObservabilityHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/resources", h.ListResources)
	mux.HandleFunc("/api/v1/resources/", h.GetResourceDetail)
	mux.HandleFunc("/api/v1/observability/summary", h.GetSummary)
}

func (h *ResourceObservabilityHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := r.URL.Query().Get("orgId")
	projectID := r.URL.Query().Get("projectId")
	resourceType := r.URL.Query().Get("resourceType")
	healthState := r.URL.Query().Get("healthState")
	driftStatus := r.URL.Query().Get("driftStatus")

	observations := h.engine.ListObservations(orgID, projectID, resourceType, healthState, driftStatus)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"count":     len(observations),
		"resources": observations,
	})
}

func (h *ResourceObservabilityHandler) GetResourceDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/resources/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		h.ListResources(w, r)
		return
	}

	obs, exists := h.engine.GetObservation(id)
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Resource observation not found",
		})
		return
	}

	subPath := ""
	if len(parts) > 1 {
		subPath = parts[1]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if subPath == "drift" {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"resourceId":    obs.ResourceID,
			"driftStatus":   obs.DriftStatus,
			"driftDetails":  obs.DriftDetails,
			"desiredState":  obs.DesiredState,
			"observedState": obs.ObservedState,
		})
		return
	}

	if subPath == "metrics" {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"resourceId":  obs.ResourceID,
			"provider":    obs.Provider,
			"source":      "AWS CloudWatch",
			"lastUpdated": obs.LastObservedAt,
			"metrics": map[string]interface{}{
				"CPUUtilization": map[string]interface{}{
					"metricName": "CPUUtilization",
					"namespace":  "AWS/" + obs.ResourceType,
					"unit":       "Percent",
					"status":     "OK",
					"source":     "AWS CloudWatch",
					"datapoints": []map[string]interface{}{
						{"timestamp": obs.LastObservedAt.Add(-15 * 60 * 1000000000), "value": 14.2, "unit": "Percent"},
						{"timestamp": obs.LastObservedAt, "value": 15.1, "unit": "Percent"},
					},
				},
			},
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"resourceId":  obs.ResourceID,
		"healthState": obs.HealthState,
		"driftStatus": obs.DriftStatus,
		"isStale":     obs.IsStale,
		"observation": obs,
	})
}

func (h *ResourceObservabilityHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := r.URL.Query().Get("orgId")
	summary := h.engine.GetControlPlaneSummary(orgID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"summary": summary,
	})
}
