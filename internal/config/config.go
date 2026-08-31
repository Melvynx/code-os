package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const DefaultAddress = "127.0.0.1:7890"

type Cloudflare struct {
	DashboardHost string `json:"dashboardHost"`
	TunnelMode    string `json:"tunnelMode"`
	TunnelID      string `json:"tunnelId,omitempty"`
	AccountID     string `json:"accountId,omitempty"`
	ZoneID        string `json:"zoneId,omitempty"`
	TokenFile     string `json:"tokenFile,omitempty"`
	RequireAccess bool   `json:"requireAccess"`
}

type Auth struct {
	Username       string `json:"username,omitempty"`
	PasswordFile   string `json:"passwordFile,omitempty"`
	BypassKeyFile  string `json:"bypassKeyFile,omitempty"`
	SessionKeyFile string `json:"sessionKeyFile,omitempty"`
	TrustedIPsFile string `json:"trustedIPsFile,omitempty"`
}

type Skills struct {
	Repository string `json:"repository,omitempty"`
	Directory  string `json:"directory,omitempty"`
	Branch     string `json:"branch,omitempty"`
}

type Config struct {
	Version         int        `json:"version"`
	EnvironmentName string     `json:"environmentName"`
	EnvironmentType string     `json:"environmentType"`
	Address         string     `json:"address"`
	ProjectsRoots   []string   `json:"projectsRoots"`
	ScreenshotsRoot string     `json:"screenshotsRoot"`
	FilesRoot       string     `json:"filesRoot"`
	DataDir         string     `json:"dataDir"`
	PortlyBinary    string     `json:"portlyBinary"`
	PublicPortHost  string     `json:"publicPortHost,omitempty"`
	Cloudflare      Cloudflare `json:"cloudflare"`
	Auth            Auth       `json:"auth,omitempty"`
	Skills          Skills     `json:"skills,omitempty"`
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config directory: %w", err)
	}
	return filepath.Join(dir, "code-os", "config.json"), nil
}

func Defaults() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("resolve home directory: %w", err)
	}
	dataDir := filepath.Join(home, ".local", "share", "code-os")
	environmentType := "local"
	if runtime.GOOS == "linux" {
		environmentType = "remote"
	}
	hostname, _ := os.Hostname()
	return Config{
		Version:         1,
		EnvironmentName: hostname,
		EnvironmentType: environmentType,
		Address:         DefaultAddress,
		ProjectsRoots:   []string{filepath.Join(home, "projects")},
		ScreenshotsRoot: filepath.Join(dataDir, "screenshots"),
		FilesRoot:       filepath.Join(dataDir, "files"),
		DataDir:         dataDir,
		PortlyBinary:    "portly",
		Cloudflare: Cloudflare{
			TunnelMode:    "shared",
			RequireAccess: true,
		},
	}, nil
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), ".config-*.json")
	if err != nil {
		return fmt.Errorf("create temporary config: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("secure temporary config: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary config: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary config: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("install config: %w", err)
	}
	return nil
}

func (cfg Config) Validate() error {
	var problems []string
	if cfg.Version != 1 {
		problems = append(problems, "version must be 1")
	}
	if strings.TrimSpace(cfg.EnvironmentName) == "" {
		problems = append(problems, "environmentName is required")
	}
	if cfg.EnvironmentType != "local" && cfg.EnvironmentType != "remote" {
		problems = append(problems, "environmentType must be local or remote")
	}
	if cfg.Address == "" {
		problems = append(problems, "address is required")
	}
	if len(cfg.ProjectsRoots) == 0 {
		problems = append(problems, "at least one projects root is required")
	}
	if cfg.DataDir == "" {
		problems = append(problems, "dataDir is required")
	}
	if cfg.FilesRoot == "" {
		problems = append(problems, "filesRoot is required")
	}
	if cfg.PortlyBinary == "" {
		problems = append(problems, "portlyBinary is required")
	}
	if cfg.Cloudflare.DashboardHost != "" && !strings.Contains(cfg.Cloudflare.DashboardHost, ".") {
		problems = append(problems, "dashboardHost must be a fully-qualified hostname")
	}
	if cfg.Cloudflare.DashboardHost != "" && cfg.Auth.PasswordFile == "" {
		problems = append(problems, "dashboardHost requires origin authentication")
	}
	if cfg.Cloudflare.TunnelMode != "" && cfg.Cloudflare.TunnelMode != "shared" && cfg.Cloudflare.TunnelMode != "dedicated" {
		problems = append(problems, "tunnelMode must be shared or dedicated")
	}
	if (cfg.Auth.Username == "") != (cfg.Auth.PasswordFile == "") {
		problems = append(problems, "auth username and passwordFile must be configured together")
	}
	if cfg.Auth.BypassKeyFile != "" && cfg.Auth.PasswordFile == "" {
		problems = append(problems, "auth bypassKeyFile requires username and passwordFile")
	}
	if cfg.Auth.PasswordFile != "" && cfg.Auth.SessionKeyFile == "" {
		problems = append(problems, "auth sessionKeyFile is required when authentication is configured")
	}
	if cfg.Auth.TrustedIPsFile != "" && cfg.Auth.PasswordFile == "" {
		problems = append(problems, "auth trustedIPsFile requires username and passwordFile")
	}
	if cfg.PublicPortHost != "" && !strings.Contains(cfg.PublicPortHost, "{port}") {
		problems = append(problems, "publicPortHost must contain {port}")
	}
	if cfg.PublicPortHost != "" && cfg.Auth.PasswordFile == "" {
		problems = append(problems, "publicPortHost requires dashboard authentication")
	}
	if cfg.Skills.Repository != "" && cfg.Skills.Directory == "" {
		problems = append(problems, "skills directory is required when repository is configured")
	}
	if cfg.Skills.Branch == "" && cfg.Skills.Repository != "" {
		problems = append(problems, "skills branch is required when repository is configured")
	}
	if len(problems) > 0 {
		return fmt.Errorf("invalid config: %w", errors.New(strings.Join(problems, "; ")))
	}
	return nil
}
