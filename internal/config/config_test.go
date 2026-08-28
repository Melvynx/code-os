package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "nested", "config.json")
	cfg := Config{
		Version: 1, EnvironmentName: "test-host", EnvironmentType: "remote",
		Address: "127.0.0.1:7890", ProjectsRoots: []string{directory},
		ScreenshotsRoot: directory, FilesRoot: directory, DataDir: directory, PortlyBinary: "portly",
		Cloudflare: Cloudflare{RequireAccess: true},
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Cloudflare.DashboardHost != cfg.Cloudflare.DashboardHost {
		t.Fatalf("dashboard host = %q, want %q", loaded.Cloudflare.DashboardHost, cfg.Cloudflare.DashboardHost)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("config permissions = %o, want 600", got)
	}
}

func TestValidateRejectsUnsafeShape(t *testing.T) {
	t.Parallel()
	cfg := Config{Version: 1, EnvironmentName: "host", EnvironmentType: "remote", Address: "127.0.0.1:7890", DataDir: "/tmp", PortlyBinary: "portly"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing projects root error")
	}
}

func TestValidateRejectsBypassWithoutDashboardAuth(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Version: 1, EnvironmentName: "host", EnvironmentType: "remote",
		Address: "127.0.0.1:7890", ProjectsRoots: []string{"/tmp/projects"},
		DataDir: "/tmp/data", PortlyBinary: "portly",
		Auth: Auth{BypassKeyFile: "/tmp/media-bypass-key"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want bypass auth dependency error")
	}
}
