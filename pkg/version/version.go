package version

import (
	"runtime"
)

const (
	// ANARVA_VERSION defines the centralized platform version string across backend, console, SDK, CLI, and Terraform provider.
	ANARVA_VERSION = "0.1.0"
)

type VersionInfo struct {
	Version     string `json:"version"`
	GitCommit   string `json:"gitCommit"`
	BuildTime   string `json:"buildTime"`
	GoVersion   string `json:"goVersion"`
	Environment string `json:"environment"`
	Platform    string `json:"platform"`
}

var (
	gitCommit = "phase-60-cd8ca2a"
	buildTime = "2026-08-19T20:45:00Z"
)

func GetVersionInfo(environment string) VersionInfo {
	if environment == "" {
		environment = "production"
	}
	return VersionInfo{
		Version:     ANARVA_VERSION,
		GitCommit:   gitCommit,
		BuildTime:   buildTime,
		GoVersion:   runtime.Version(),
		Environment: environment,
		Platform:    runtime.GOOS + "/" + runtime.GOARCH,
	}
}
