package portly

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStopPassesExactServerIDToPortly(t *testing.T) {
	directory := t.TempDir()
	logPath := filepath.Join(directory, "arguments")
	binary := filepath.Join(directory, "portly")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" > \"$CODE_OS_TEST_LOG\"\n"
	if err := os.WriteFile(binary, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODE_OS_TEST_LOG", logPath)

	if err := (Client{Binary: binary}).Stop(context.Background(), "srv_safe-id"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	arguments, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(arguments)) != "stop srv_safe-id --json" {
		t.Fatalf("arguments = %q", arguments)
	}
}

func TestStopReturnsPortlyFailureWithoutShellExecution(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	binary := filepath.Join(directory, "portly")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\necho 'not found' >&2\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	err := (Client{Binary: binary}).Stop(context.Background(), "$(touch /tmp/never)")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("Stop() error = %v", err)
	}
}
