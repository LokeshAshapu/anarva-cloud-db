package worker

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrRuntimeUnavailable = errors.New("container runtime engine (Docker) is unavailable")
	ErrUnsafeCommand      = errors.New("unsafe host command execution rejected")
	safeProjectIDRegex    = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

type ContainerRuntime interface {
	HasRuntime() bool
	CreateContainer(ctx context.Context, m *WorkloadMetadata, envVars map[string]string) (string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string) error
	RestartContainer(ctx context.Context, containerID string) error
	DeleteContainer(ctx context.Context, containerID string) error
	ExecuteContainerCommand(ctx context.Context, containerID string, command string, timeoutSec int) (int, string, string, error)
	InspectContainerStatus(ctx context.Context, containerID string) (string, error)
	GetContainerMetrics(ctx context.Context, containerID string) (float64, int, int64, int64, error)
	DiscoverANARVAContainers(ctx context.Context) ([]string, error)
}

type DockerContainerRuntime struct {
	hasDocker bool
}

func NewDockerContainerRuntime() *DockerContainerRuntime {
	path, err := exec.LookPath("docker")
	if err != nil {
		return &DockerContainerRuntime{hasDocker: false}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "info")
	return &DockerContainerRuntime{
		hasDocker: (cmd.Run() == nil),
	}
}

func (r *DockerContainerRuntime) HasRuntime() bool {
	return r.hasDocker
}

func sanitizeProjectID(projectID string) string {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" || !safeProjectIDRegex.MatchString(projectID) {
		return "default"
	}
	return projectID
}

func (r *DockerContainerRuntime) ensureProjectNetwork(ctx context.Context, projectID string) (string, error) {
	safeProj := sanitizeProjectID(projectID)
	netName := fmt.Sprintf("anarva-project-%s-net", safeProj)

	// Check if network exists
	inspectCmd := exec.CommandContext(ctx, "docker", "network", "inspect", netName)
	if err := inspectCmd.Run(); err == nil {
		return netName, nil
	}

	// Create bridge network
	createCmd := exec.CommandContext(ctx, "docker", "network", "create", "--driver", "bridge", netName)
	if err := createCmd.Run(); err != nil {
		return "bridge", nil // fallback to standard bridge if custom network fails
	}
	return netName, nil
}

func (r *DockerContainerRuntime) CreateContainer(ctx context.Context, m *WorkloadMetadata, envVars map[string]string) (string, error) {
	if !r.hasDocker {
		return "", ErrRuntimeUnavailable
	}

	containerName := fmt.Sprintf("anarva-worker-acu-%s", m.WorkloadID)
	netName, _ := r.ensureProjectNetwork(ctx, m.ProjectID)

	args := []string{
		"run", "-d",
		"--name", containerName,
		"--network", netName,
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges:true",
		"--pids-limit", "512",
		"--label", fmt.Sprintf("com.anarva.workload_id=%s", m.WorkloadID),
		"--label", fmt.Sprintf("com.anarva.resource_id=%s", m.ResourceID),
		"--label", fmt.Sprintf("com.anarva.project_id=%s", m.ProjectID),
	}

	if m.VCPU > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%.1f", m.VCPU))
	}
	if m.MemoryMB > 0 {
		args = append(args, "--memory", fmt.Sprintf("%dm", m.MemoryMB))
	}

	// Inject EnvVars in memory
	for k, v := range envVars {
		if k != "" {
			args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
		}
	}

	args = append(args, m.Image)

	if len(m.Command) > 0 {
		args = append(args, m.Command...)
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create container via docker run: %s (%w)", strings.TrimSpace(string(out)), err)
	}

	containerID := strings.TrimSpace(string(out))
	return containerID, nil
}

func (r *DockerContainerRuntime) StartContainer(ctx context.Context, containerID string) error {
	if !r.hasDocker {
		return ErrRuntimeUnavailable
	}
	cmd := exec.CommandContext(ctx, "docker", "start", containerID)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start container: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (r *DockerContainerRuntime) StopContainer(ctx context.Context, containerID string) error {
	if !r.hasDocker {
		return ErrRuntimeUnavailable
	}
	cmd := exec.CommandContext(ctx, "docker", "stop", "-t", "10", containerID)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop container: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (r *DockerContainerRuntime) RestartContainer(ctx context.Context, containerID string) error {
	if !r.hasDocker {
		return ErrRuntimeUnavailable
	}
	cmd := exec.CommandContext(ctx, "docker", "restart", "-t", "10", containerID)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restart container: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (r *DockerContainerRuntime) DeleteContainer(ctx context.Context, containerID string) error {
	if !r.hasDocker {
		return ErrRuntimeUnavailable
	}
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerID)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete container: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (r *DockerContainerRuntime) ExecuteContainerCommand(ctx context.Context, containerID string, command string, timeoutSec int) (int, string, string, error) {
	if !r.hasDocker {
		return 1, "", "", ErrRuntimeUnavailable
	}

	if timeoutSec <= 0 || timeoutSec > 30 {
		timeoutSec = 15
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "docker", "exec", containerID, "sh", "-c", command)
	outBytes, err := cmd.CombinedOutput()

	// Bound response output to 1MB
	if len(outBytes) > 1024*1024 {
		outBytes = outBytes[:1024*1024]
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return exitCode, string(outBytes), "", nil
}

func (r *DockerContainerRuntime) InspectContainerStatus(ctx context.Context, containerID string) (string, error) {
	if !r.hasDocker {
		return "unavailable", ErrRuntimeUnavailable
	}
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Status}}", containerID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "unavailable", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *DockerContainerRuntime) GetContainerMetrics(ctx context.Context, containerID string) (float64, int, int64, int64, error) {
	if !r.hasDocker {
		return 0, 0, 0, 0, ErrRuntimeUnavailable
	}

	cmd := exec.CommandContext(
		ctx,
		"docker",
		"stats",
		"--no-stream",
		"--format",
		"{{.CPUPerc}},{{.MemUsage}},{{.NetIO}}",
		containerID,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("docker stats failed: %s: %w", strings.TrimSpace(string(out)), err)
	}

	line := strings.TrimSpace(string(out))
	parts := strings.Split(line, ",")
	if len(parts) != 3 {
		return 0, 0, 0, 0, fmt.Errorf("unexpected docker stats output: %q", line)
	}

	cpuText := strings.TrimSpace(strings.TrimSuffix(parts[0], "%"))
	cpu, err := strconv.ParseFloat(cpuText, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid docker CPU metric %q: %w", parts[0], err)
	}

	memoryParts := strings.Fields(parts[1])
	if len(memoryParts) < 2 {
		return 0, 0, 0, 0, fmt.Errorf("invalid docker memory metric %q", parts[1])
	}

	memoryMB, err := parseDockerBytes(memoryParts[0])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid docker memory metric %q: %w", memoryParts[0], err)
	}
	memoryMB /= 1024 * 1024

	netParts := strings.Split(parts[2], "/")
	if len(netParts) != 2 {
		return 0, 0, 0, 0, fmt.Errorf("invalid docker network metric %q", parts[2])
	}

	rx, err := parseDockerBytes(strings.TrimSpace(netParts[0]))
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid docker RX metric %q: %w", netParts[0], err)
	}

	tx, err := parseDockerBytes(strings.TrimSpace(netParts[1]))
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid docker TX metric %q: %w", netParts[1], err)
	}

	return cpu, int(memoryMB), rx, tx, nil
}

func parseDockerBytes(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New("empty byte value")
	}

	units := map[string]float64{
		"B":   1,
		"KB":  1024,
		"MB":  1024 * 1024,
		"GB":  1024 * 1024 * 1024,
		"KiB": 1024,
		"MiB": 1024 * 1024,
		"GiB": 1024 * 1024 * 1024,
	}

	for unit, multiplier := range units {
		if strings.HasSuffix(value, unit) {
			number := strings.TrimSpace(strings.TrimSuffix(value, unit))
			v, err := strconv.ParseFloat(number, 64)
			if err != nil {
				return 0, err
			}
			return int64(v * multiplier), nil
		}
	}

	return 0, fmt.Errorf("unsupported byte unit in %q", value)
}

func (r *DockerContainerRuntime) DiscoverANARVAContainers(ctx context.Context) ([]string, error) {
	if !r.hasDocker {
		return nil, nil
	}
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", "label=com.anarva.workload_id", "--format", "{{.ID}}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var ids []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			ids = append(ids, strings.TrimSpace(l))
		}
	}
	return ids, nil
}
