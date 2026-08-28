package projects

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestScannerDiscoversRepositoryAndChanges(t *testing.T) {
	directory := t.TempDir()
	repository := filepath.Join(directory, "sample")
	if err := os.MkdirAll(filepath.Join(repository, "packages", "web"), 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, repository, "init", "-b", "main")
	runGit(t, repository, "config", "user.email", "test@example.com")
	runGit(t, repository, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("initial\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "packages", "web", "package.json"), []byte(`{"name":"web"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repository, "add", ".")
	runGit(t, repository, "commit", "-m", "initial")
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	projects, warnings := NewScanner().Scan(context.Background(), []string{directory})
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v", warnings)
	}
	if len(projects) != 1 {
		t.Fatalf("projects = %d, want 1", len(projects))
	}
	if projects[0].Git.Modified != 1 || projects[0].Git.Untracked != 1 {
		t.Fatalf("git state = %+v, want one modified and one untracked", projects[0].Git)
	}
	if len(projects[0].Subprojects) != 1 || projects[0].Subprojects[0].Name != "web" {
		t.Fatalf("subprojects = %+v, want web", projects[0].Subprojects)
	}
}

func runGit(t *testing.T, directory string, arguments ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", directory}, arguments...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
}
