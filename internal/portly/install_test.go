package portly

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExistingPersistentIgnoresCleanupPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ephemeral := filepath.Join(root, ".cleanup-code-os-20260908044922", "portly")
	if err := os.MkdirAll(filepath.Dir(ephemeral), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ephemeral, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := existingPersistent(ephemeral); got == ephemeral {
		t.Fatalf("existingPersistent() returned the ephemeral path %q", got)
	}
}
