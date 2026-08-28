package main

import (
	"os"
	"path/filepath"
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
