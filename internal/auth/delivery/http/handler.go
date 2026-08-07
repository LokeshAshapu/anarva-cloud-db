package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/auth/usecase"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

type signUpReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type createAPIKeyReq struct {
	Name       string `json:"name"`
	ExpiryDays int    `json:"expiry_days"`
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/signup", h.SignUp)
	mux.HandleFunc("POST /api/v1/auth/verify-email", h.VerifyEmail)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.RefreshToken)
	mux.HandleFunc("POST /api/v1/auth/api-keys", h.CreateAPIKey)
	mux.HandleFunc("GET /api/v1/auth/api-keys", h.ListAPIKeys)
	mux.HandleFunc("DELETE /api/v1/auth/api-keys/{id}", h.RevokeAPIKey)
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req signUpReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid request payload"))
		return
	}

	user, token, err := h.authUseCase.SignUp(r.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		respondError(w, err)
		return
	}

	metrics.RecordHTTPRequest(http.StatusCreated, r.Method, r.URL.Path, 0)
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"user_id":           user.ID,
		"email":             user.Email,
		"verification_code": token,
		"message":           "Registration successful. Please verify your email.",
	})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "token query parameter required"))
		return
	}

	if err := h.authUseCase.VerifyEmail(r.Context(), token); err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Email successfully verified"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid request payload"))
		return
	}

	userAgent := r.UserAgent()
	ipAddress := r.RemoteAddr

	accessTok, refreshTok, expiry, user, err := h.authUseCase.Login(r.Context(), req.Email, req.Password, userAgent, ipAddress)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessTok,
		"refresh_token": refreshTok,
		"expires_in":    int64(expiry.Seconds()),
		"user": map[string]interface{}{
			"id":        user.ID,
			"email":     user.Email,
			"full_name": user.FullName,
			"role":      user.Role,
			"status":    user.Status,
		},
	})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid request payload"))
		return
	}

	accessTok, refreshTok, expiry, err := h.authUseCase.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessTok,
		"refresh_token": refreshTok,
		"expires_in":    int64(expiry.Seconds()),
	})
}

func (h *AuthHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		respondError(w, appErrors.New(appErrors.CodeUnauthorized, "unauthorized"))
		return
	}

	var req createAPIKeyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid request payload"))
		return
	}

	rawKey, apiKey, err := h.authUseCase.CreateAPIKey(r.Context(), userID, req.Name, req.ExpiryDays)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         apiKey.ID,
		"name":       apiKey.Name,
		"raw_key":    rawKey,
		"created_at": apiKey.CreatedAt,
	})
}

func (h *AuthHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		respondError(w, appErrors.New(appErrors.CodeUnauthorized, "unauthorized"))
		return
	}

	keys, err := h.authUseCase.ListAPIKeys(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"api_keys": keys})
}

func (h *AuthHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		respondError(w, appErrors.New(appErrors.CodeUnauthorized, "unauthorized"))
		return
	}

	keyID := r.PathValue("id")
	if keyID == "" {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "API key ID required"))
		return
	}

	if err := h.authUseCase.RevokeAPIKey(r.Context(), userID, keyID); err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
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

func getUserIDFromContext(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}
