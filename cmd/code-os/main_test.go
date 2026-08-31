package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsurePasswordFileSecuresExistingSecret(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(path, []byte("existing-secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ensurePasswordFile(path); err != nil {
		t.Fatalf("ensurePasswordFile() error = %v", err)
	}

	if !isPrivateRegularFile(path) {
		t.Fatal("secret file mode is not 0600")
	}
	value, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(value) != "existing-secret\n" {
		t.Fatal("existing secret was replaced")
	}
}

func TestSystemdUnitStartsCodeOSAtBoot(t *testing.T) {
	t.Parallel()
	unit := renderUnit("/usr/local/bin/code-os", "/home/dev/.config/code-os/config.json", "/home/dev/.local/share/code-os", "")
	for _, expected := range []string{"Restart=on-failure", "WantedBy=default.target", "After=network-online.target"} {
		if !strings.Contains(unit, expected) {
			t.Errorf("systemd unit missing %q", expected)
		}
	}
}

func TestEnsurePasswordFileCreatesPrivateSecret(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "nested", "secret")

	if err := ensurePasswordFile(path); err != nil {
		t.Fatalf("ensurePasswordFile() error = %v", err)
	}
	if !isPrivateRegularFile(path) {
		t.Fatal("generated secret file mode is not 0600")
	}
}

func TestEnsureEmptyPrivateFileCreatesSecureStorage(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "nested", "trusted-ips")

	if err := ensureEmptyPrivateFile(path); err != nil {
		t.Fatalf("ensureEmptyPrivateFile() error = %v", err)
	}
	if !isPrivateRegularFile(path) {
		t.Fatal("trusted IP storage mode is not 0600")
	}
}

func TestEnsureEmptyPrivateFileRejectsSymlink(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	target := filepath.Join(directory, "target")
	path := filepath.Join(directory, "trusted-ips")
	if err := os.WriteFile(target, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}

	if err := ensureEmptyPrivateFile(path); err == nil {
		t.Fatal("ensureEmptyPrivateFile() accepted a symlink")
	}
}
