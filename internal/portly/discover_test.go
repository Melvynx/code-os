package portly

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListGitRootsFindsImmediateRepositories(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	repo := filepath.Join(root, "demo-app")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "not-a-repo"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := ListGitRoots(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != repo {
		t.Fatalf("ListGitRoots() = %v, want [%s]", got, repo)
	}
}

func TestFormatDiscoveryMentionsEnvAndNoAutostart(t *testing.T) {
	t.Parallel()
	text := FormatDiscovery([]DiscoveredProject{{
		Name: "demo-app", Path: "/root/projects/demo-app", Added: true, Detail: "registered (no servers started)",
	}})
	for _, expected := range []string{"demo-app", "no servers started", "0600", "Never print values"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("FormatDiscovery() = %q, want %q", text, expected)
		}
	}
}
