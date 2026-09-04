package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/melvynx/code-os/internal/config"
	"github.com/melvynx/code-os/internal/diagnostics"
)

func TestStatusAPIReportsImagePipeline(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "proof.png"), png1x1, 0o644); err != nil {
		t.Fatal(err)
	}
	server := HTTPServer{
		Config: config.Config{
			EnvironmentName: "test",
			Address:         "127.0.0.1:7890",
			ProjectsRoots:   []string{root},
			ScreenshotsRoot: root,
			FilesRoot:       root,
			PortlyBinary:    "git",
			Cloudflare:      config.Cloudflare{RequireAccess: true},
		},
		Service: &Service{},
	}
	response := httptest.NewRecorder()
	server.status(response, httptest.NewRequest(http.MethodGet, "/api/status", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	var report diagnostics.Report
	if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Images.IndexedCount != 1 || report.Images.DecodableCount != 1 {
		t.Fatalf("images = %+v", report.Images)
	}
}

func TestSkillsSyncMutationRequiresSameOrigin(t *testing.T) {
	t.Parallel()
	server := HTTPServer{Config: config.Config{EnvironmentName: "test"}}
	request := httptest.NewRequest(http.MethodPost, "https://code-os.example/api/skills-sync", nil)
	request.Host = "code-os.example"
	request.Header.Set("Origin", "https://attacker.example")
	request.Header.Set("X-Forwarded-Proto", "https")
	response := httptest.NewRecorder()
	server.runSkillsSync(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestSkillsSyncStatusReportsUnconfiguredLibrary(t *testing.T) {
	t.Parallel()
	server := HTTPServer{Config: config.Config{EnvironmentName: "test"}}
	response := httptest.NewRecorder()
	server.skillsSyncStatus(response, httptest.NewRequest(http.MethodGet, "/api/skills-sync", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"unconfigured"`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

var png1x1 = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
	0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}
