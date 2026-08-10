package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/region"
	"github.com/anarva-cloud/anarva-cloud-db/internal/resource"
)

type ResourceHandler struct {
	registry *resource.Registry
	stream   *activity.Stream
}

func NewResourceHandler(registry *resource.Registry, stream *activity.Stream) *ResourceHandler {
	return &ResourceHandler{
		registry: registry,
		stream:   stream,
	}
}

func (h *ResourceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/resources", h.ListResources)
	mux.HandleFunc("POST /api/v1/resources", h.CreateResource)
	mux.HandleFunc("GET /api/v1/resources/", h.HandleResourceByID)
	mux.HandleFunc("PATCH /api/v1/resources/", h.HandleResourceByID)
	mux.HandleFunc("DELETE /api/v1/resources/", h.HandleResourceByID)
	mux.HandleFunc("GET /api/v1/regions", h.ListRegions)
	mux.HandleFunc("GET /api/v1/activities", h.ListActivities)
}

func (h *ResourceHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organizationId")
	projectID := r.URL.Query().Get("projectId")
	regionID := r.URL.Query().Get("regionId")
	resType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
	query := r.URL.Query().Get("query")

	resources := h.registry.List(orgID, projectID, regionID, resType, status, query)
	respondJSON(w, http.StatusOK, resources)
}

func (h *ResourceHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
	var res resource.CloudResource
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, `{"error":"invalid json body"}`, http.StatusBadRequest)
		return
	}

	if res.Name == "" || res.Type == "" {
		http.Error(w, `{"error":"name and type are required"}`, http.StatusBadRequest)
		return
	}

	if res.OrganizationID == "" {
		res.OrganizationID = "org-default"
	}
	if res.ProjectID == "" {
		res.ProjectID = "proj-default"
	}
	if res.RegionID == "" {
		res.RegionID = "ap-hyderabad-1"
	}
	if res.Status == "" {
		res.Status = resource.StatusAvailable
	}
	if res.Environment == "" {
		res.Environment = "Production"
	}

	if err := h.registry.Create(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.stream.Record(&activity.ActivityEvent{
		OrganizationID: res.OrganizationID,
		ProjectID:      res.ProjectID,
		ResourceID:     res.ResourceID,
		ActorID:        "lokeshashapu@gmail.com",
		Action:         activity.ActionResourceCreated,
		Metadata:       map[string]string{"name": res.Name, "type": string(res.Type)},
	})

	respondJSON(w, http.StatusCreated, res)
}

func (h *ResourceHandler) HandleResourceByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/resources/")
	if id == "" {
		http.Error(w, `{"error":"missing resource id"}`, http.StatusBadRequest)
		return
	}

	orgID := r.URL.Query().Get("organizationId")

	switch r.Method {
	case http.MethodGet:
		res, err := h.registry.GetByID(id, orgID)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusNotFound)
			return
		}
		respondJSON(w, http.StatusOK, res)

	case http.MethodPatch:
		var updateReq struct {
			Name   string                  `json:"name"`
			Status resource.ResourceStatus `json:"status"`
			Tags   []resource.Tag          `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
			http.Error(w, `{"error":"invalid payload"}`, http.StatusBadRequest)
			return
		}

		updated, err := h.registry.Update(id, orgID, func(res *resource.CloudResource) {
			if updateReq.Name != "" {
				res.Name = updateReq.Name
			}
			if updateReq.Status != "" {
				res.Status = updateReq.Status
			}
			if updateReq.Tags != nil {
				res.Tags = updateReq.Tags
			}
		})
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
			return
		}

		h.stream.Record(&activity.ActivityEvent{
			OrganizationID: updated.OrganizationID,
			ProjectID:      updated.ProjectID,
			ResourceID:     updated.ResourceID,
			ActorID:        "lokeshashapu@gmail.com",
			Action:         activity.ActionResourceUpdated,
			Metadata:       map[string]string{"name": updated.Name, "status": string(updated.Status)},
		})

		respondJSON(w, http.StatusOK, updated)

	case http.MethodDelete:
		if err := h.registry.SafeDelete(id, orgID); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
			return
		}

		h.stream.Record(&activity.ActivityEvent{
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			ResourceID:     id,
			ActorID:        "lokeshashapu@gmail.com",
			Action:         activity.ActionResourceDeleted,
			Metadata:       map[string]string{"id": id},
		})

		respondJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
	}
}

func (h *ResourceHandler) ListRegions(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, region.ListRegions())
}

func (h *ResourceHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organizationId")
	respondJSON(w, http.StatusOK, h.stream.List(orgID))
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
