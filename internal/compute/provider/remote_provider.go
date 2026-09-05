package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
)

var (
	ErrRemoteWorkerUnavailable  = errors.New("remote compute worker service is unavailable")
	ErrRemoteWorkerTimeout      = errors.New("remote compute worker request timed out")
	ErrRemoteWorkerUnauthorized = errors.New("remote compute worker authentication failed")
	ErrRemoteWorkerConflict     = errors.New("remote compute workload conflict or duplicate creation")
	ErrRemoteWorkerNotFound     = errors.New("remote compute workload not found on worker node")
	ErrInsecureEndpoint         = errors.New("insecure http endpoint is prohibited for remote compute worker in production mode; https is required")
	ErrRedirectProhibited       = errors.New("http redirects are prohibited for remote compute worker endpoints")
	ErrInvalidTLSConfig         = errors.New("invalid tls/mtls configuration for remote compute worker")
)

type TLSConfigOptions struct {
	CACertPEM     string // Raw PEM string or file path
	ClientCertPEM string // Raw PEM string or file path
	ClientKeyPEM  string // Raw PEM string or file path
	ServerName    string // Hostname for TLS server verification
	InsecureHTTP  bool   // Allow http:// in local dev mode only
}

func loadPEMMaterial(certOrPath string) ([]byte, error) {
	certOrPath = strings.TrimSpace(certOrPath)
	if certOrPath == "" {
		return nil, errors.New("empty certificate/key input")
	}
	if strings.Contains(certOrPath, "-----BEGIN") {
		return []byte(certOrPath), nil
	}
	data, err := os.ReadFile(certOrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file '%s': %w", certOrPath, err)
	}
	return data, nil
}

func BuildTLSConfig(opts TLSConfigOptions) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false, // NEVER disable verification
	}

	if opts.ServerName != "" {
		tlsConfig.ServerName = opts.ServerName
	}

	if opts.CACertPEM != "" {
		caBytes, err := loadPEMMaterial(opts.CACertPEM)
		if err != nil {
			return nil, fmt.Errorf("invalid CA certificate: %w", err)
		}
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caBytes) {
			return nil, errors.New("failed to parse CA certificate PEM")
		}
		tlsConfig.RootCAs = caPool
	}

	if opts.ClientCertPEM != "" || opts.ClientKeyPEM != "" {
		if opts.ClientCertPEM == "" || opts.ClientKeyPEM == "" {
			return nil, errors.New("both client certificate and client private key are required for mTLS")
		}
		certBytes, err := loadPEMMaterial(opts.ClientCertPEM)
		if err != nil {
			return nil, fmt.Errorf("invalid client certificate: %w", err)
		}
		keyBytes, err := loadPEMMaterial(opts.ClientKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("invalid client key: %w", err)
		}

		clientCert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid client certificate/key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	return tlsConfig, nil
}

type RemoteComputeProvider struct {
	mu         sync.RWMutex
	endpoint   string
	token      string
	httpClient *http.Client
	instances  map[string]*domain.ComputeInstance
}

func NewRemoteComputeProviderWithTLS(endpoint, token string, tlsOpts *TLSConfigOptions, client *http.Client) (*RemoteComputeProvider, error) {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return nil, errors.New("remote compute worker endpoint URL is required")
	}

	insecureHTTP := false
	if tlsOpts != nil {
		insecureHTTP = tlsOpts.InsecureHTTP
	} else if os.Getenv("ANARVA_ENV") != "production" {
		insecureHTTP = true
	}

	if !insecureHTTP && strings.HasPrefix(strings.ToLower(endpoint), "http://") {
		return nil, ErrInsecureEndpoint
	}

	if client == nil {
		transport := &http.Transport{}
		if tlsOpts != nil && (tlsOpts.CACertPEM != "" || tlsOpts.ClientCertPEM != "" || tlsOpts.ServerName != "") {
			tlsCfg, err := BuildTLSConfig(*tlsOpts)
			if err != nil {
				return nil, err
			}
			transport.TLSClientConfig = tlsCfg
		}
		client = &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
		}
	} else if tlsOpts != nil && (tlsOpts.CACertPEM != "" || tlsOpts.ClientCertPEM != "" || tlsOpts.ServerName != "") {
		tlsCfg, err := BuildTLSConfig(*tlsOpts)
		if err != nil {
			return nil, err
		}
		if tr, ok := client.Transport.(*http.Transport); ok && tr != nil {
			tr.TLSClientConfig = tlsCfg
		} else {
			client.Transport = &http.Transport{TLSClientConfig: tlsCfg}
		}
	}

	// Always enforce redirect rejection to prevent credential leakage
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return ErrRedirectProhibited
	}

	return &RemoteComputeProvider{
		endpoint:   endpoint,
		token:      strings.TrimSpace(token),
		httpClient: client,
		instances:  make(map[string]*domain.ComputeInstance),
	}, nil
}

func NewRemoteComputeProvider(endpoint, token string, client *http.Client) (*RemoteComputeProvider, error) {
	return NewRemoteComputeProviderWithTLS(endpoint, token, nil, client)
}

func (p *RemoteComputeProvider) GetProviderType() domain.ProviderType {
	return domain.ProviderType("REMOTE_WORKER")
}

// Worker API Contract Request/Response Types
type CreateWorkloadRequest struct {
	WorkloadID string            `json:"workloadId"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	Command    []string          `json:"command,omitempty"`
	VCPU       float64           `json:"vcpu"`
	MemoryMB   int               `json:"memoryMb"`
	EnvVars    map[string]string `json:"envVars,omitempty"`
	Region     string            `json:"region"`
	OrgID      string            `json:"organizationId,omitempty"`
	ProjectID  string            `json:"projectId,omitempty"`
}

type WorkloadResponse struct {
	WorkloadID string    `json:"workloadId"`
	Status     string    `json:"status"` // RUNNING, STOPPED, DELETED, FAILED
	Health     string    `json:"health"` // HEALTHY, DEGRADED, UNAVAILABLE
	PrivateIP  string    `json:"privateIp,omitempty"`
	PublicIP   string    `json:"publicIp,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ExecuteCommandWorkerRequest struct {
	Command string `json:"command"`
	Timeout int    `json:"timeoutSeconds,omitempty"`
}

type ExecuteCommandWorkerResponse struct {
	ExitCode int       `json:"exitCode"`
	Stdout   string    `json:"stdout"`
	Stderr   string    `json:"stderr"`
	Executed time.Time `json:"executedAt"`
}

type WorkerMetricsResponse struct {
	WorkloadID string                 `json:"workloadId"`
	CPUUsage   float64                `json:"cpuUsagePercent"`
	MemoryMB   int                    `json:"memoryUsageMb"`
	NetworkRx  int64                  `json:"networkRxBytes"`
	NetworkTx  int64                  `json:"networkTxBytes"`
	RawMetrics map[string]interface{} `json:"rawMetrics,omitempty"`
}

type WorkerErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func (p *RemoteComputeProvider) doRequest(ctx context.Context, method, path string, bodyPayload interface{}, headers map[string]string) ([]byte, int, error) {
	fullURL := p.endpoint + path

	var reqBody io.Reader
	if bodyPayload != nil {
		jsonBytes, err := json.Marshal(bodyPayload)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal worker request payload: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create worker HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if p.token != "" {
		req.Header.Set("X-Anarva-Worker-Token", p.token)
		req.Header.Set("Authorization", "Bearer "+p.token)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, ErrRedirectProhibited) || strings.Contains(err.Error(), "prohibited") {
			return nil, 0, ErrRedirectProhibited
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) {
			return nil, 0, ErrRemoteWorkerTimeout
		}
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			if strings.Contains(urlErr.Error(), "certificate") || strings.Contains(urlErr.Error(), "tls") || strings.Contains(urlErr.Error(), "x509") {
				return nil, 0, fmt.Errorf("remote worker tls error: %w", ErrInvalidTLSConfig)
			}
			return nil, 0, ErrRemoteWorkerUnavailable
		}
		return nil, 0, fmt.Errorf("remote worker connection error: %w", err)
	}
	defer resp.Body.Close()

	// Limit response reader to 1MB for safety
	bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if readErr != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read worker response body: %w", readErr)
	}

	// Sanitize HTTP Status Codes
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		return bodyBytes, resp.StatusCode, nil

	case http.StatusUnauthorized, http.StatusForbidden:
		return bodyBytes, resp.StatusCode, ErrRemoteWorkerUnauthorized

	case http.StatusNotFound:
		return bodyBytes, resp.StatusCode, ErrRemoteWorkerNotFound

	case http.StatusConflict:
		return bodyBytes, resp.StatusCode, ErrRemoteWorkerConflict

	default:
		var errResp WorkerErrorResponse
		_ = json.Unmarshal(bodyBytes, &errResp)
		msg := errResp.Message
		if msg == "" {
			msg = errResp.Error
		}
		if msg == "" {
			msg = fmt.Sprintf("remote worker returned HTTP %d", resp.StatusCode)
		}
		return bodyBytes, resp.StatusCode, fmt.Errorf("remote compute provider error: %s", msg)
	}
}

func (p *RemoteComputeProvider) CreateInstance(ctx context.Context, inst *domain.ComputeInstance) (*domain.ComputeInstance, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst.Provider = domain.ProviderType("REMOTE_WORKER")
	inst.DeletedAt = nil

	req := CreateWorkloadRequest{
		WorkloadID: inst.ID,
		Name:       inst.Name,
		Image:      inst.DockerImage,
		Command:    inst.Command,
		VCPU:       inst.VCPU,
		MemoryMB:   inst.MemoryMB,
		EnvVars:    inst.EnvVars,
		Region:     inst.RegionID,
		OrgID:      inst.OrganizationID,
		ProjectID:  inst.ProjectID,
	}

	headers := map[string]string{
		"Idempotency-Key": inst.ResourceID,
	}

	respBytes, _, err := p.doRequest(ctx, http.MethodPost, "/v1/workloads", req, headers)
	if err != nil {
		return nil, err
	}

	var wResp WorkloadResponse
	if err := json.Unmarshal(respBytes, &wResp); err != nil {
		return nil, fmt.Errorf("failed to parse remote worker workload response: %w", err)
	}

	inst.ProviderInstanceID = wResp.WorkloadID
	if wResp.Status != "" {
		inst.Status = domain.InstanceStatus(wResp.Status)
	} else {
		inst.Status = domain.StatusRunning
	}
	if wResp.Health != "" {
		inst.Health = domain.InstanceHealth(wResp.Health)
	} else {
		inst.Health = domain.HealthHealthy
	}

	if wResp.PrivateIP != "" {
		inst.PrivateIP = wResp.PrivateIP
	}
	if wResp.PublicIP != "" {
		inst.PublicIP = wResp.PublicIP
	}

	p.instances[inst.ID] = inst
	return inst, nil
}

func (p *RemoteComputeProvider) GetInstance(ctx context.Context, id string) (*domain.ComputeInstance, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	inst, ok := p.instances[id]
	if !ok {
		return nil, ErrRemoteWorkerNotFound
	}
	return inst, nil
}

func (p *RemoteComputeProvider) ListInstances(ctx context.Context, projectID string) ([]*domain.ComputeInstance, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var list []*domain.ComputeInstance
	for _, inst := range p.instances {
		if inst.ProjectID == projectID && inst.DeletedAt == nil {
			list = append(list, inst)
		}
	}
	return list, nil
}

func (p *RemoteComputeProvider) StartInstance(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, ok := p.instances[id]
	if !ok {
		return ErrRemoteWorkerNotFound
	}

	workloadID := inst.ProviderInstanceID
	if workloadID == "" {
		workloadID = inst.ID
	}

	_, _, err := p.doRequest(ctx, http.MethodPost, "/v1/workloads/"+workloadID+"/start", nil, nil)
	if err != nil {
		return err
	}

	inst.Status = domain.StatusRunning
	inst.Health = domain.HealthHealthy
	inst.UpdatedAt = time.Now()
	return nil
}

func (p *RemoteComputeProvider) StopInstance(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, ok := p.instances[id]
	if !ok {
		return ErrRemoteWorkerNotFound
	}

	workloadID := inst.ProviderInstanceID
	if workloadID == "" {
		workloadID = inst.ID
	}

	_, _, err := p.doRequest(ctx, http.MethodPost, "/v1/workloads/"+workloadID+"/stop", nil, nil)
	if err != nil {
		return err
	}

	inst.Status = domain.StatusStopped
	inst.Health = domain.HealthUnavailable
	inst.UpdatedAt = time.Now()
	return nil
}

func (p *RemoteComputeProvider) RestartInstance(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, ok := p.instances[id]
	if !ok {
		return ErrRemoteWorkerNotFound
	}

	workloadID := inst.ProviderInstanceID
	if workloadID == "" {
		workloadID = inst.ID
	}

	_, _, err := p.doRequest(ctx, http.MethodPost, "/v1/workloads/"+workloadID+"/restart", nil, nil)
	if err != nil {
		return err
	}

	inst.Status = domain.StatusRunning
	inst.Health = domain.HealthHealthy
	inst.UpdatedAt = time.Now()
	return nil
}

func (p *RemoteComputeProvider) DeleteInstance(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, ok := p.instances[id]
	if !ok {
		return ErrRemoteWorkerNotFound
	}

	workloadID := inst.ProviderInstanceID
	if workloadID == "" {
		workloadID = inst.ID
	}

	_, _, err := p.doRequest(ctx, http.MethodDelete, "/v1/workloads/"+workloadID, nil, nil)
	if err != nil && !errors.Is(err, ErrRemoteWorkerNotFound) {
		return err
	}

	now := time.Now()
	inst.Status = domain.StatusDeleted
	inst.DeletedAt = &now
	return nil
}

func (p *RemoteComputeProvider) ResizeInstance(ctx context.Context, id string, newACU float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, ok := p.instances[id]
	if !ok {
		return ErrRemoteWorkerNotFound
	}

	inst.ACU = newACU
	inst.VCPU = newACU
	inst.MemoryMB = int(newACU * 2048)
	inst.StorageGB = int(newACU * 10)
	inst.UpdatedAt = time.Now()
	return nil
}

func (p *RemoteComputeProvider) RebuildInstance(ctx context.Context, id string, imageID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, ok := p.instances[id]
	if !ok {
		return ErrRemoteWorkerNotFound
	}

	inst.ImageID = imageID
	inst.UpdatedAt = time.Now()
	return nil
}

func (p *RemoteComputeProvider) GetInstanceHealth(ctx context.Context, id string) (domain.InstanceHealth, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if inst, ok := p.instances[id]; ok {
		return inst.Health, nil
	}
	return domain.HealthUnknown, nil
}

func (p *RemoteComputeProvider) GetInstanceMetrics(ctx context.Context, id string) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	inst, ok := p.instances[id]
	if !ok {
		return nil, ErrRemoteWorkerNotFound
	}

	workloadID := inst.ProviderInstanceID
	if workloadID == "" {
		workloadID = inst.ID
	}

	respBytes, _, err := p.doRequest(ctx, http.MethodGet, "/v1/workloads/"+workloadID+"/metrics", nil, nil)
	if err != nil {
		return nil, err
	}

	var mResp WorkerMetricsResponse
	if err := json.Unmarshal(respBytes, &mResp); err != nil {
		return nil, fmt.Errorf("failed to parse remote worker metrics response: %w", err)
	}

	return map[string]interface{}{
		"instanceId":        inst.ID,
		"workloadId":        mResp.WorkloadID,
		"status":            inst.Status,
		"cpuUsagePercent":   mResp.CPUUsage,
		"memoryUsageMb":     mResp.MemoryMB,
		"networkRxBytes":    mResp.NetworkRx,
		"networkTxBytes":    mResp.NetworkTx,
		"provider":          p.GetProviderType(),
		"telemetryState":    "HONEST_CONTROL_PLANE_STATE",
		"providerConnected": true,
	}, nil
}

func (p *RemoteComputeProvider) ExecuteCommand(ctx context.Context, id string, req *domain.CommandExecutionRequest) (*domain.CommandExecutionResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	inst, ok := p.instances[id]
	if !ok {
		return nil, ErrRemoteWorkerNotFound
	}

	if inst.Status != domain.StatusRunning {
		return nil, fmt.Errorf("cannot execute command on non-running instance (current state: %s)", inst.Status)
	}

	workloadID := inst.ProviderInstanceID
	if workloadID == "" {
		workloadID = inst.ID
	}

	wReq := ExecuteCommandWorkerRequest{
		Command: req.Command,
		Timeout: req.Timeout,
	}

	respBytes, _, err := p.doRequest(ctx, http.MethodPost, "/v1/workloads/"+workloadID+"/execute", wReq, nil)
	if err != nil {
		return nil, err
	}

	var eResp ExecuteCommandWorkerResponse
	if err := json.Unmarshal(respBytes, &eResp); err != nil {
		return nil, fmt.Errorf("failed to parse remote worker execution response: %w", err)
	}

	return &domain.CommandExecutionResult{
		ExitCode: eResp.ExitCode,
		Stdout:   eResp.Stdout,
		Stderr:   eResp.Stderr,
		Executed: eResp.Executed,
	}, nil
}

func (p *RemoteComputeProvider) RehydrateInstance(ctx context.Context, inst *domain.ComputeInstance) error {
	if inst == nil || inst.ID == "" {
		return fmt.Errorf("invalid instance for re-hydration")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	workloadID := inst.ProviderInstanceID
	if workloadID == "" {
		workloadID = inst.ID
	}

	respBytes, _, err := p.doRequest(ctx, http.MethodGet, "/v1/workloads/"+workloadID, nil, nil)
	if err != nil {
		inst.Status = domain.StatusStopped
		inst.Health = domain.HealthUnavailable
		p.instances[inst.ID] = inst
		return nil
	}

	var wResp WorkloadResponse
	if err := json.Unmarshal(respBytes, &wResp); err == nil {
		if wResp.Status != "" {
			inst.Status = domain.InstanceStatus(wResp.Status)
		}
		if wResp.Health != "" {
			inst.Health = domain.InstanceHealth(wResp.Health)
		}
	}

	p.instances[inst.ID] = inst
	return nil
}
