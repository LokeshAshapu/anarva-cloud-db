package worker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
)

type WorkerServer struct {
	config  *Config
	store   MetadataStore
	runtime ContainerRuntime
	server  *http.Server
	mu      sync.RWMutex
}

func NewWorkerServer(cfg *Config, store MetadataStore, runtime ContainerRuntime) (*WorkerServer, error) {
	if cfg == nil {
		var err error
		cfg, err = LoadWorkerConfigFromEnv()
		if err != nil {
			return nil, err
		}
	}

	if store == nil {
		var err error
		store, err = NewFileMetadataStore(cfg.DBPath)
		if err != nil {
			return nil, err
		}
	}

	if runtime == nil {
		runtime = NewDockerContainerRuntime()
	}

	ws := &WorkerServer{
		config:  cfg,
		store:   store,
		runtime: runtime,
	}

	// Reconcile and recover state on startup
	_ = ws.ReconcileStartupState(context.Background())

	return ws, nil
}

func (s *WorkerServer) ReconcileStartupState(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records, err := s.store.List()
	if err != nil {
		return err
	}

	for _, m := range records {
		if m.ContainerID != "" && s.runtime.HasRuntime() {
			status, err := s.runtime.InspectContainerStatus(ctx, m.ContainerID)
			if err == nil {
				if status == "running" {
					m.Status = "RUNNING"
					m.Health = "HEALTHY"
				} else {
					m.Status = "STOPPED"
					m.Health = "UNAVAILABLE"
				}
			} else {
				m.Status = "FAILED"
				m.Health = "UNAVAILABLE"
			}
			_ = s.store.Save(m)
		}
	}

	return nil
}

func (s *WorkerServer) authenticateRequest(r *http.Request) error {
	// Reject browser JWTs or direct unauthenticated access
	if s.config.RequireClientAuth && r.TLS != nil {
		if len(r.TLS.PeerCertificates) == 0 {
			return errors.New("mTLS client certificate is required")
		}
	}

	if s.config.WorkerToken != "" {
		tokenHeader := strings.TrimSpace(r.Header.Get("X-Anarva-Worker-Token"))
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		bearerToken := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenHeader != s.config.WorkerToken && bearerToken != s.config.WorkerToken {
			return errors.New("invalid or missing service authentication token")
		}
	}

	return nil
}

func (s *WorkerServer) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/workloads", s.handleWorkloads)
	mux.HandleFunc("/v1/workloads/", s.handleWorkloadSubroutes)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.authenticateRequest(r); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			})
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func (s *WorkerServer) handleWorkloads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))

	// Check idempotency first
	if idempotencyKey != "" {
		if existing, err := s.store.GetByIdempotencyKey(idempotencyKey); err == nil && existing != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkloadResponse{
				WorkloadID: existing.WorkloadID,
				Status:     existing.Status,
				Health:     existing.Health,
				PrivateIP:  existing.PrivateIP,
				PublicIP:   existing.PublicIP,
				CreatedAt:  existing.CreatedAt,
			})
			return
		}
	}

	var req computeProvider.CreateWorkloadRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1024*1024)).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid JSON workload creation payload",
		})
		return
	}

	if req.WorkloadID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "workloadId is required",
		})
		return
	}

	// Check duplicate workload ID
	if existing, err := s.store.Get(req.WorkloadID); err == nil && existing != nil {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
			Code:    "CONFLICT",
			Message: fmt.Sprintf("Workload '%s' already exists", req.WorkloadID),
		})
		return
	}

	meta := &WorkloadMetadata{
		WorkloadID:     req.WorkloadID,
		ResourceID:     idempotencyKey,
		Name:           req.Name,
		Image:          req.Image,
		Command:        req.Command,
		VCPU:           req.VCPU,
		MemoryMB:       req.MemoryMB,
		OrgID:          req.OrgID,
		ProjectID:      req.ProjectID,
		Region:         req.Region,
		Status:         "RUNNING",
		Health:         "HEALTHY",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if s.runtime.HasRuntime() {
		containerID, err := s.runtime.CreateContainer(r.Context(), meta, req.EnvVars)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
				Code:    "RUNTIME_ERROR",
				Message: fmt.Sprintf("Container creation failed: %v", err),
			})
			return
		}
		meta.ContainerID = containerID
	} else {
		meta.ContainerID = fmt.Sprintf("sim-worker-%s", req.WorkloadID)
	}

	meta.PrivateIP = "10.200.1.5"
	meta.PublicIP = "20.200.1.5"

	_ = s.store.Save(meta)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(computeProvider.WorkloadResponse{
		WorkloadID: meta.WorkloadID,
		Status:     meta.Status,
		Health:     meta.Health,
		PrivateIP:  meta.PrivateIP,
		PublicIP:   meta.PublicIP,
		CreatedAt:  meta.CreatedAt,
	})
}

func (s *WorkerServer) handleWorkloadSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/workloads/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	workloadID := parts[0]
	meta, err := s.store.Get(workloadID)
	if err != nil || meta == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
			Code:    "NOT_FOUND",
			Message: fmt.Sprintf("Workload '%s' not found", workloadID),
		})
		return
	}

	subaction := ""
	if len(parts) > 1 {
		subaction = parts[1]
	}

	switch subaction {
	case "":
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkloadResponse{
				WorkloadID: meta.WorkloadID,
				Status:     meta.Status,
				Health:     meta.Health,
				PrivateIP:  meta.PrivateIP,
				PublicIP:   meta.PublicIP,
				CreatedAt:  meta.CreatedAt,
			})

		case http.MethodDelete:
			if meta.ContainerID != "" && s.runtime.HasRuntime() {
				_ = s.runtime.DeleteContainer(r.Context(), meta.ContainerID)
			}
			_ = s.store.Delete(workloadID)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	case "start":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if meta.ContainerID != "" && s.runtime.HasRuntime() {
			_ = s.runtime.StartContainer(r.Context(), meta.ContainerID)
		}
		meta.Status = "RUNNING"
		meta.Health = "HEALTHY"
		_ = s.store.Save(meta)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(computeProvider.WorkloadResponse{
			WorkloadID: meta.WorkloadID,
			Status:     meta.Status,
			Health:     meta.Health,
		})

	case "stop":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if meta.ContainerID != "" && s.runtime.HasRuntime() {
			_ = s.runtime.StopContainer(r.Context(), meta.ContainerID)
		}
		meta.Status = "STOPPED"
		meta.Health = "UNAVAILABLE"
		_ = s.store.Save(meta)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(computeProvider.WorkloadResponse{
			WorkloadID: meta.WorkloadID,
			Status:     meta.Status,
			Health:     meta.Health,
		})

	case "restart":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if meta.ContainerID != "" && s.runtime.HasRuntime() {
			_ = s.runtime.RestartContainer(r.Context(), meta.ContainerID)
		}
		meta.Status = "RUNNING"
		meta.Health = "HEALTHY"
		_ = s.store.Save(meta)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(computeProvider.WorkloadResponse{
			WorkloadID: meta.WorkloadID,
			Status:     meta.Status,
			Health:     meta.Health,
		})

	case "execute":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req computeProvider.ExecuteCommandWorkerRequest
		if err := json.NewDecoder(io.LimitReader(r.Body, 1024*1024)).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
				Code:    "BAD_REQUEST",
				Message: "Invalid JSON execution payload",
			})
			return
		}

		exitCode := 0
		stdout := "EXECUTED: " + req.Command
		stderr := ""

		if meta.ContainerID != "" && s.runtime.HasRuntime() {
			var execErr error
			exitCode, stdout, stderr, execErr = s.runtime.ExecuteContainerCommand(r.Context(), meta.ContainerID, req.Command, req.Timeout)
			if execErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
					Code:    "EXEC_ERROR",
					Message: execErr.Error(),
				})
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(computeProvider.ExecuteCommandWorkerResponse{
			ExitCode: exitCode,
			Stdout:   stdout,
			Stderr:   stderr,
			Executed: time.Now(),
		})

	case "metrics":
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if meta.ContainerID == "" {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
				Code:    "METRICS_UNAVAILABLE",
				Message: "Workload has no container runtime identifier",
			})
			return
		}

		if !s.runtime.HasRuntime() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
				Code:    "RUNTIME_UNAVAILABLE",
				Message: "Container runtime is unavailable",
			})
			return
		}

		cpu, mem, rx, tx, err := s.runtime.GetContainerMetrics(r.Context(), meta.ContainerID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkerErrorResponse{
				Code:    "METRICS_ERROR",
				Message: err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(computeProvider.WorkerMetricsResponse{
			WorkloadID: meta.WorkloadID,
			CPUUsage:   cpu,
			MemoryMB:   mem,
			NetworkRx:  rx,
			NetworkTx:  tx,
		})

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *WorkerServer) Start(ctx context.Context) error {
	handler := s.Handler()

	s.server = &http.Server{
		Addr:    s.config.ListenAddr,
		Handler: handler,
	}

	if s.config.ServerCertPEM != "" && s.config.ServerKeyPEM != "" {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		if s.config.RequireClientAuth && s.config.CACertPEM != "" {
			caPool := x509.NewCertPool()
			caPool.AppendCertsFromPEM([]byte(s.config.CACertPEM))
			tlsConfig.ClientCAs = caPool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		cert, err := tls.X509KeyPair([]byte(s.config.ServerCertPEM), []byte(s.config.ServerKeyPEM))
		if err != nil {
			return fmt.Errorf("failed to parse server certificate/key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		s.server.TLSConfig = tlsConfig

		ln, err := net.Listen("tcp", s.config.ListenAddr)
		if err != nil {
			return err
		}
		return s.server.ServeTLS(ln, "", "")
	}

	ln, err := net.Listen("tcp", s.config.ListenAddr)
	if err != nil {
		return err
	}
	return s.server.Serve(ln)
}

func (s *WorkerServer) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
