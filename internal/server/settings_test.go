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

func TestIsSameOriginAcceptsBrowserFormPosts(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		host    string
		headers map[string]string
		want    bool
	}{
		{
			name: "explicit origin behind https proxy",
			host: "code-os.example",
			headers: map[string]string{
				"Origin":            "https://code-os.example",
				"X-Forwarded-Proto": "https",
			},
			want: true,
		},
		{
			name: "referer fallback without origin behind cloudflare",
			host: "code-os.example",
			headers: map[string]string{
				"Referer":           "https://code-os.example/trust-ip?next=%2Fapp%2F",
				"X-Forwarded-Proto": "https",
			},
			want: true,
		},
		{
			name: "attacker referer without origin",
			host: "code-os.example",
			headers: map[string]string{
				"Referer":           "https://attacker.example/trust-ip",
				"X-Forwarded-Proto": "https",
			},
			want: false,
		},
		{
			name: "navigational form post without origin",
			host: "code-os-demo1.mlvcdn.com",
			headers: map[string]string{
				"Sec-Fetch-Site":    "same-origin",
				"Sec-Fetch-Mode":    "navigate",
				"X-Forwarded-Proto": "https",
			},
			want: true,
		},
		{
			name: "https origin when proxy omits forwarded proto",
			host: "code-os-demo1.mlvcdn.com",
			headers: map[string]string{
				"Origin": "https://code-os-demo1.mlvcdn.com",
			},
			want: true,
		},
		{
			name: "https origin with default port on host",
			host: "code-os.example:443",
			headers: map[string]string{
				"Origin":            "https://code-os.example",
				"X-Forwarded-Proto": "https",
			},
			want: true,
		},
		{
			name: "cf-visitor without x-forwarded-proto",
			host: "code-os.example",
			headers: map[string]string{
				"Origin":     "https://code-os.example",
				"CF-Visitor": `{"scheme":"https"}`,
			},
			want: true,
		},
		{
			name: "attacker origin",
			host: "code-os.example",
			headers: map[string]string{
				"Origin":            "https://attacker.example",
				"X-Forwarded-Proto": "https",
			},
			want: false,
		},
		{
			name: "same-site sibling host",
			host: "code-os.example",
			headers: map[string]string{
				"Origin":            "https://port5173.example",
				"Sec-Fetch-Site":    "same-site",
				"X-Forwarded-Proto": "https",
			},
			want: false,
		},
		{
			name: "cross-site form post",
			host: "code-os.example",
			headers: map[string]string{
				"Origin":         "https://attacker.example",
				"Sec-Fetch-Site": "cross-site",
				"Sec-Fetch-Mode": "navigate",
			},
			want: false,
		},
		{
			name:    "bare post with no origin signals",
			host:    "code-os.example",
			headers: map[string]string{},
			want:    false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(http.MethodPost, "https://"+strings.TrimSuffix(test.host, ":443")+"/auth/trust-ip", nil)
			request.Host = test.host
			for key, value := range test.headers {
				request.Header.Set(key, value)
			}
			if got := isSameOrigin(request); got != test.want {
				t.Fatalf("isSameOrigin() = %t, want %t", got, test.want)
			}
		})
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
