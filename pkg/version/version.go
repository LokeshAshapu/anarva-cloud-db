package version

import (
	"runtime"
	"time"
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
	gitCommit = "23f0d1045198"
	buildTime = time.Now().Format(time.RFC3339)
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
