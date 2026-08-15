package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	APIURL         string `yaml:"api_url"`
	OrganizationID string `yaml:"organization_id"`
	ProjectID      string `yaml:"project_id"`
	Environment    string `yaml:"environment"` // LIVE or TEST
	APIKey         string `yaml:"api_key,omitempty"`
}

type Config struct {
	ActiveProfile string             `yaml:"active_profile"`
	Profiles      map[string]*Profile `yaml:"profiles"`
}

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".anarva")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func GetConfigFile() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func LoadConfig() (*Config, error) {
	file, err := GetConfigFile()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), nil
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]*Profile)
	}
	if cfg.ActiveProfile == "" {
		cfg.ActiveProfile = "default"
	}
	if _, ok := cfg.Profiles["default"]; !ok {
		cfg.Profiles["default"] = DefaultProfile()
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	file, err := GetConfigFile()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0600)
}

func DefaultProfile() *Profile {
	apiURL := "https://anarva-cloud-db-api.onrender.com"
	if url := os.Getenv("ANARVA_API_URL"); url != "" {
		apiURL = url
	}
	return &Profile{
		APIURL:         apiURL,
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		Environment:    "LIVE",
	}
}

func DefaultConfig() *Config {
	return &Config{
		ActiveProfile: "default",
		Profiles: map[string]*Profile{
			"default": DefaultProfile(),
		},
	}
}

func (c *Config) GetCurrentProfile() *Profile {
	p, ok := c.Profiles[c.ActiveProfile]
	if !ok {
		p = DefaultProfile()
		c.Profiles[c.ActiveProfile] = p
	}
	if envURL := os.Getenv("ANARVA_API_URL"); envURL != "" {
		p.APIURL = envURL
	}
	if envKey := os.Getenv("ANARVA_API_KEY"); envKey != "" {
		p.APIKey = envKey
	}
	return p
}
