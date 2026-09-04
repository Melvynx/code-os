package diagnostics

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/melvynx/code-os/internal/config"
	"github.com/melvynx/code-os/internal/skills"
)

var png1x1 = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
	0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

func TestInspectReportsDecodableScreenshotPipeline(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	imagePath := filepath.Join(root, "skills-verify", "code-os", "F01.png")
	if err := os.MkdirAll(filepath.Dir(imagePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(imagePath, png1x1, 0o644); err != nil {
		t.Fatal(err)
	}
	bypass := filepath.Join(root, "bypass")
	if err := os.WriteFile(bypass, []byte("key\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	report := Inspect(config.Config{
		Address:         "127.0.0.1:7890",
		ProjectsRoots:   []string{root},
		ScreenshotsRoot: root,
		FilesRoot:       root,
		PortlyBinary:    "git",
		Cloudflare:      config.Cloudflare{DashboardHost: "code-os.example", RequireAccess: true},
		Auth:            config.Auth{BypassKeyFile: bypass},
	}, Options{Skills: skills.Inspector{RunSystemd: missingSystemd}})

	if report.Images.IndexedCount != 1 || report.Images.DecodableCount != 1 || report.Images.UndecodableCount != 0 {
		t.Fatalf("images = %+v", report.Images)
	}
	if len(report.Images.Recent) != 1 || report.Images.Recent[0].Width != 1 || !report.Images.Recent[0].Decodable {
		t.Fatalf("recent = %+v", report.Images.Recent)
	}
	if statusOf(report, "decodable-images") != Pass || statusOf(report, "media-bypass") != Pass {
		t.Fatalf("checks = %+v", report.Checks)
	}
}

func TestInspectFlagsEmptyAndUndecodableImages(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "empty.png"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "broken.png"), []byte("not-a-png"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := Inspect(config.Config{
		Address:         "127.0.0.1:7890",
		ProjectsRoots:   []string{root},
		ScreenshotsRoot: root,
		FilesRoot:       root,
		PortlyBinary:    "git",
	}, Options{Skills: skills.Inspector{RunSystemd: missingSystemd}})
	if report.Images.EmptyCount != 1 || report.Images.UndecodableCount != 1 {
		t.Fatalf("images = %+v", report.Images)
	}
	if statusOf(report, "decodable-images") != Fail {
		t.Fatalf("decodable check = %s", statusOf(report, "decodable-images"))
	}
}

func TestInspectRequiresPrivateFilesRoot(t *testing.T) {
	t.Parallel()
	report := Inspect(config.Config{
		Address:         "127.0.0.1:7890",
		ProjectsRoots:   []string{t.TempDir()},
		ScreenshotsRoot: filepath.Join(t.TempDir(), "missing-screenshots"),
		FilesRoot:       filepath.Join(t.TempDir(), "missing-files"),
		PortlyBinary:    "git",
	}, Options{Skills: skills.Inspector{RunSystemd: missingSystemd}})
	if statusOf(report, "files-root") != Fail || statusOf(report, "screenshots-root") != Warn {
		t.Fatalf("root checks = files:%s screenshots:%s", statusOf(report, "files-root"), statusOf(report, "screenshots-root"))
	}
	if report.Healthy {
		t.Fatal("report should be unhealthy when filesRoot is missing")
	}
}

func statusOf(report Report, id string) Severity {
	for _, item := range report.Checks {
		if item.ID == id {
			return item.Status
		}
	}
	return ""
}

func missingSystemd(args ...string) (string, error) {
	return "", os.ErrNotExist
}
