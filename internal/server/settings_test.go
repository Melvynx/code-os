package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/melvynx/code-os/internal/config"
)

func TestSettingsResponseNeverContainsCloudflareTokenContents(t *testing.T) {
	t.Parallel()
	server := HTTPServer{Config: config.Config{Cloudflare: config.Cloudflare{TokenFile: "/root/.config/code-os/cloudflare-token"}}}
	response := httptest.NewRecorder()
	server.settings(response, httptest.NewRequest(http.MethodGet, "/api/settings", nil))
	if strings.Contains(response.Body.String(), "cloudflareToken\"") {
		t.Fatalf("settings response exposes the write-only token field: %s", response.Body.String())
	}
}

func TestSettingsUpdateRejectsCrossOriginRequest(t *testing.T) {
	t.Parallel()
	server := HTTPServer{ConfigPath: "/tmp/unused-code-os-config"}
	request := httptest.NewRequest(http.MethodPut, "https://code-os.example/api/settings", strings.NewReader(`{}`))
	request.Host = "code-os.example"
	request.Header.Set("Origin", "https://attacker.example")
	request.Header.Set("X-Forwarded-Proto", "https")
	response := httptest.NewRecorder()
	server.updateSettings(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestGitHubRepositoryValidationRejectsEmbeddedCredentials(t *testing.T) {
	t.Parallel()
	for _, repository := range []string{
		"https://token@github.com/owner/skills.git",
		"https://example.com/owner/skills.git",
		"git@evil.example:owner/skills.git",
	} {
		if validGitHubRepository(repository) {
			t.Fatalf("validGitHubRepository(%q) = true", repository)
		}
	}
}
