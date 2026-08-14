package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/service"
)

type StorageHandler struct {
	svc *service.StorageService
}

func NewStorageHandler(svc *service.StorageService) *StorageHandler {
	return &StorageHandler{svc: svc}
}

func (h *StorageHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/storage/buckets", h.handleBuckets)
	mux.HandleFunc("POST /api/v1/storage/buckets", h.handleBuckets)
	mux.HandleFunc("GET /api/v1/storage/buckets/", h.handleBucketSubroutes)
	mux.HandleFunc("DELETE /api/v1/storage/buckets/", h.handleBucketSubroutes)
	mux.HandleFunc("POST /api/v1/storage/buckets/", h.handleBucketSubroutes)

	// S3 Compatibility Layer
	mux.HandleFunc("GET /s3/", h.handleS3Compatibility)
	mux.HandleFunc("PUT /s3/", h.handleS3Compatibility)
}

func (h *StorageHandler) handleBuckets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		orgID := r.URL.Query().Get("organizationId")
		projectID := r.URL.Query().Get("projectId")
		buckets, err := h.svc.ListBuckets(r.Context(), orgID, projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": buckets})

	case http.MethodPost:
		var req struct {
			OrganizationID string `json:"organizationId"`
			ProjectID      string `json:"projectId"`
			Name           string `json:"name"`
			Region         string `json:"region"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			req.Name = "production-bucket"
		}
		if req.Region == "" {
			req.Region = "ap-hyderabad-1"
		}

		b, err := h.svc.CreateBucket(r.Context(), req.OrganizationID, req.ProjectID, req.Name, req.Region)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": b})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *StorageHandler) handleBucketSubroutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/storage/buckets/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Invalid bucket ID", http.StatusBadRequest)
		return
	}

	bucketID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "":
		if r.Method == http.MethodDelete {
			if err := h.svc.DeleteBucket(r.Context(), bucketID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "DELETED", "id": bucketID})
		}

	case "objects":
		if r.Method == http.MethodGet {
			prefix := r.URL.Query().Get("prefix")
			objs, err := h.svc.ListObjects(r.Context(), bucketID, prefix)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": objs})
		}

	case "signed-url":
		if r.Method == http.MethodPost {
			var req struct {
				Key        string `json:"key"`
				Method     string `json:"method"`
				ExpiresSec int    `json:"expiresSec"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.Method == "" {
				req.Method = "GET"
			}

			pURL, err := h.svc.GenerateSignedURL(r.Context(), bucketID, req.Key, req.Method, req.ExpiresSec)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": pURL})
		}
	}
}

func (h *StorageHandler) handleS3Compatibility(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("X-Anarva-S3-Compatibility", "ANARVA_COMPATIBLE")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><ListBucketResult><Name>s3-compat</Name></ListBucketResult>`))
}
