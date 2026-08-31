package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/melvynx/code-os/internal/config"
)

type settingsResponse struct {
	EnvironmentName           string            `json:"environmentName"`
	ProjectsRoots             []string          `json:"projectsRoots"`
	ScreenshotsRoot           string            `json:"screenshotsRoot"`
	FilesRoot                 string            `json:"filesRoot"`
	DataDir                   string            `json:"dataDir"`
	PortlyBinary              string            `json:"portlyBinary"`
	PublicPortHost            string            `json:"publicPortHost"`
	Cloudflare                config.Cloudflare `json:"cloudflare"`
	Auth                      config.Auth       `json:"auth"`
	Skills                    config.Skills     `json:"skills"`
	CloudflareTokenConfigured bool              `json:"cloudflareTokenConfigured"`
	RestartRequired           bool              `json:"restartRequired,omitempty"`
}

type settingsRequest struct {
	EnvironmentName string            `json:"environmentName"`
	ProjectsRoots   []string          `json:"projectsRoots"`
	ScreenshotsRoot string            `json:"screenshotsRoot"`
	FilesRoot       string            `json:"filesRoot"`
	DataDir         string            `json:"dataDir"`
	PortlyBinary    string            `json:"portlyBinary"`
	PublicPortHost  string            `json:"publicPortHost"`
	Cloudflare      config.Cloudflare `json:"cloudflare"`
	Auth            config.Auth       `json:"auth"`
	Skills          config.Skills     `json:"skills"`
	CloudflareToken string            `json:"cloudflareToken,omitempty"`
}

func (server HTTPServer) settings(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, server.settingsView(false))
}

func (server HTTPServer) updateSettings(response http.ResponseWriter, request *http.Request) {
	if !isSameOrigin(request) {
		writeJSON(response, http.StatusForbidden, map[string]string{"error": "same-origin request required"})
		return
	}
	if server.ConfigPath == "" {
		writeJSON(response, http.StatusNotImplemented, map[string]string{"error": "settings persistence is not configured"})
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, 64<<10)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var input settingsRequest
	if err := decoder.Decode(&input); err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid settings payload"})
		return
	}
	next := server.Config
	next.EnvironmentName = strings.TrimSpace(input.EnvironmentName)
	next.ProjectsRoots = cleanPaths(input.ProjectsRoots)
	next.ScreenshotsRoot = cleanOptionalPath(input.ScreenshotsRoot)
	next.FilesRoot = cleanOptionalPath(input.FilesRoot)
	next.DataDir = cleanOptionalPath(input.DataDir)
	next.PortlyBinary = strings.TrimSpace(input.PortlyBinary)
	next.PublicPortHost = strings.ToLower(strings.TrimSpace(input.PublicPortHost))
	next.Cloudflare = cleanCloudflare(input.Cloudflare)
	next.Auth = cleanAuth(input.Auth)
	next.Skills = cleanSkills(input.Skills)
	if err := validateSettings(next); err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := config.Save(server.ConfigPath, next); err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "could not persist settings"})
		return
	}
	if token := strings.TrimSpace(input.CloudflareToken); token != "" {
		if err := writeConfigSecret(next.Cloudflare.TokenFile, token); err != nil {
			writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "settings saved but Cloudflare token could not be updated"})
			return
		}
	}
	view := settingsView(next, true)
	view.CloudflareTokenConfigured = privateRegularFile(next.Cloudflare.TokenFile)
	writeJSON(response, http.StatusOK, view)
}

func (server HTTPServer) settingsView(restartRequired bool) settingsResponse {
	return settingsView(server.Config, restartRequired)
}

func settingsView(configuration config.Config, restartRequired bool) settingsResponse {
	return settingsResponse{
		EnvironmentName: configuration.EnvironmentName, ProjectsRoots: configuration.ProjectsRoots,
		ScreenshotsRoot: configuration.ScreenshotsRoot, FilesRoot: configuration.FilesRoot,
		DataDir: configuration.DataDir, PortlyBinary: configuration.PortlyBinary,
		PublicPortHost: configuration.PublicPortHost, Cloudflare: configuration.Cloudflare,
		Auth: configuration.Auth, Skills: configuration.Skills,
		CloudflareTokenConfigured: privateRegularFile(configuration.Cloudflare.TokenFile),
		RestartRequired:           restartRequired,
	}
}

func validateSettings(cfg config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	paths := append([]string{}, cfg.ProjectsRoots...)
	paths = append(paths, cfg.ScreenshotsRoot, cfg.FilesRoot, cfg.DataDir, cfg.Skills.Directory,
		cfg.Auth.PasswordFile, cfg.Auth.BypassKeyFile, cfg.Auth.SessionKeyFile, cfg.Auth.TrustedIPsFile, cfg.Cloudflare.TokenFile)
	for _, path := range paths {
		if path == "" {
			continue
		}
		if !filepath.IsAbs(path) || filepath.Clean(path) == string(filepath.Separator) {
			return fmt.Errorf("paths must be absolute and cannot be the filesystem root")
		}
	}
	if cfg.Skills.Repository != "" && !validGitHubRepository(cfg.Skills.Repository) {
		return errors.New("skills repository must be a credential-free GitHub HTTPS or SSH URL")
	}
	configurationRoot, err := os.UserConfigDir()
	if err != nil {
		return errors.New("user configuration directory is unavailable")
	}
	for _, secretPath := range []string{cfg.Cloudflare.TokenFile, cfg.Auth.PasswordFile, cfg.Auth.BypassKeyFile, cfg.Auth.SessionKeyFile, cfg.Auth.TrustedIPsFile} {
		if secretPath != "" && !pathWithin(configurationRoot, secretPath) {
			return errors.New("credential files must stay under the user configuration directory")
		}
	}
	return nil
}

func cleanCloudflare(value config.Cloudflare) config.Cloudflare {
	value.DashboardHost = strings.ToLower(strings.TrimSpace(value.DashboardHost))
	value.TunnelMode = strings.TrimSpace(value.TunnelMode)
	value.TunnelID = strings.TrimSpace(value.TunnelID)
	value.AccountID = strings.TrimSpace(value.AccountID)
	value.ZoneID = strings.TrimSpace(value.ZoneID)
	value.TokenFile = cleanOptionalPath(value.TokenFile)
	return value
}

func cleanAuth(value config.Auth) config.Auth {
	value.Username = strings.TrimSpace(value.Username)
	value.PasswordFile = cleanOptionalPath(value.PasswordFile)
	value.BypassKeyFile = cleanOptionalPath(value.BypassKeyFile)
	value.SessionKeyFile = cleanOptionalPath(value.SessionKeyFile)
	value.TrustedIPsFile = cleanOptionalPath(value.TrustedIPsFile)
	return value
}

func cleanSkills(value config.Skills) config.Skills {
	value.Repository = strings.TrimSpace(value.Repository)
	value.Directory = cleanOptionalPath(value.Directory)
	value.Branch = strings.TrimSpace(value.Branch)
	return value
}

func cleanOptionalPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return filepath.Clean(value)
}

func cleanPaths(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			cleaned = append(cleaned, filepath.Clean(value))
		}
	}
	return cleaned
}

func validGitHubRepository(value string) bool {
	if strings.HasPrefix(value, "git@github.com:") {
		path := strings.TrimSuffix(strings.TrimPrefix(value, "git@github.com:"), ".git")
		return validRepositoryPath(path)
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() != "github.com" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	return validRepositoryPath(strings.TrimSuffix(strings.TrimPrefix(parsed.Path, "/"), ".git"))
}

func validRepositoryPath(path string) bool {
	parts := strings.Split(path, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != "" && parts[0] != "." && parts[1] != "."
}

func isSameOrigin(request *http.Request) bool {
	origin := strings.TrimSpace(request.Header.Get("Origin"))
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	return err == nil && strings.EqualFold(parsed.Host, request.Host) && parsed.Scheme == externalProtocol(request)
}

func writeConfigSecret(path, secret string) error {
	if path == "" {
		return errors.New("secret path is not configured")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".secret-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.WriteString(secret + "\n"); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func privateRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode().Perm() == 0o600
}

func pathWithin(root, candidate string) bool {
	root, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(root, candidate)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
