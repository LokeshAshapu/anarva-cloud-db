package http

import (
	"encoding/json"
	"net/http"

	"github.com/anarva-cloud/anarva-cloud-db/internal/project/usecase"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
)

type ProjectHandler struct {
	useCase usecase.ProjectUseCase
}

func NewProjectHandler(useCase usecase.ProjectUseCase) *ProjectHandler {
	return &ProjectHandler{useCase: useCase}
}

type createOrgReq struct {
	OwnerID string `json:"owner_id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
}

type createProjectReq struct {
	OrgID  string `json:"org_id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Region string `json:"region"`
}

type inviteReq struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (h *ProjectHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/organizations", h.CreateOrganization)
	mux.HandleFunc("GET /api/v1/organizations/{id}", h.GetOrganization)
	mux.HandleFunc("POST /api/v1/projects", h.CreateProject)
	mux.HandleFunc("GET /api/v1/projects/{id}", h.GetProject)
	mux.HandleFunc("GET /api/v1/organizations/{org_id}/projects", h.ListProjects)
	mux.HandleFunc("DELETE /api/v1/projects/{id}", h.DeleteProject)
	mux.HandleFunc("POST /api/v1/organizations/{org_id}/invitations", h.InviteMember)
	mux.HandleFunc("POST /api/v1/invitations/accept", h.AcceptInvitation)
	mux.HandleFunc("GET /api/v1/organizations/{org_id}/members", h.ListMembers)
}

func (h *ProjectHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req createOrgReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid request body"))
		return
	}

	org, err := h.useCase.CreateOrganization(r.Context(), req.OwnerID, req.Name, req.Slug)
	if err != nil {
		respondError(w, err)
		return
	}

	metrics.RecordHTTPRequest(http.StatusCreated, r.Method, r.URL.Path, 0)
	respondJSON(w, http.StatusCreated, org)
}

func (h *ProjectHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	org, err := h.useCase.GetOrganization(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, org)
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid request body"))
		return
	}

	project, err := h.useCase.CreateProject(r.Context(), req.OrgID, req.Name, req.Slug, req.Region)
	if err != nil {
		respondError(w, err)
		return
	}

	metrics.RecordHTTPRequest(http.StatusCreated, r.Method, r.URL.Path, 0)
	respondJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := h.useCase.GetProject(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("org_id")
	projects, err := h.useCase.ListProjects(r.Context(), orgID)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"projects": projects})
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.Header.Get("X-User-ID")

	if err := h.useCase.DeleteProject(r.Context(), userID, id); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *ProjectHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("org_id")
	invitedBy := r.Header.Get("X-User-ID")

	var req inviteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid request body"))
		return
	}

	token, err := h.useCase.InviteMember(r.Context(), orgID, req.Email, req.Role, invitedBy)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"invitation_token": token,
		"message":          "Invitation created successfully",
	})
}

func (h *ProjectHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	userID := r.Header.Get("X-User-ID")

	if err := h.useCase.AcceptInvitation(r.Context(), token, userID); err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *ProjectHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("org_id")
	members, err := h.useCase.ListMembers(r.Context(), orgID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"members": members})
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*appErrors.AppError); ok {
		respondJSON(w, appErr.HTTPStatusCode(), appErr)
		return
	}
	respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
