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
	worktree := filepath.Join(directory, "feature-worktree")
	runGit(t, repository, "worktree", "add", "-b", "feature/worktree", worktree)
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktree, "README.md"), []byte("worktree change\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	projects, warnings := NewScanner().Scan(context.Background(), []string{directory})
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v", warnings)
	}
	if len(projects) != 1 {
		t.Fatalf("projects = %d, want 1", len(projects))
	}
	if projects[0].Path != repository {
		t.Fatalf("project path = %q, want primary worktree %q", projects[0].Path, repository)
	}
	if projects[0].Git.Modified != 1 || projects[0].Git.Untracked != 1 {
		t.Fatalf("git state = %+v, want one modified and one untracked", projects[0].Git)
	}
	if len(projects[0].Worktrees) != 2 {
		t.Fatalf("worktrees = %+v, want primary and linked worktree", projects[0].Worktrees)
	}
	if !projects[0].Worktrees[0].Main || projects[0].Worktrees[1].Git.Branch != "feature/worktree" || projects[0].Worktrees[1].Git.Modified != 1 {
		t.Fatalf("worktrees = %+v, want main marker and modified feature branch", projects[0].Worktrees)
	}
	if len(projects[0].Subprojects) != 1 || projects[0].Subprojects[0].Name != "web" {
		t.Fatalf("subprojects = %+v, want web", projects[0].Subprojects)
	}
}

func TestParseWorktreeListPreservesDetachedAndUnavailableEntries(t *testing.T) {
	t.Parallel()
	records := parseWorktreeList("worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /tmp/detached\nHEAD def456\ndetached\nlocked reason\n\nworktree /tmp/old\nHEAD 000000\nbranch refs/heads/old\nprunable missing\n")

	if len(records) != 3 {
		t.Fatalf("records = %+v, want 3", records)
	}
	if !records[0].Main || records[0].Branch != "main" {
		t.Fatalf("main record = %+v", records[0])
	}
	if records[1].Branch != "" || !records[1].Locked {
		t.Fatalf("detached record = %+v", records[1])
	}
	if !records[2].Prunable || records[2].Branch != "old" {
		t.Fatalf("prunable record = %+v", records[2])
	}
}

func TestParseGitStatusNormalizesDetachedHead(t *testing.T) {
	t.Parallel()
	state := parseGitStatus("# branch.oid abc123\n# branch.head (detached)\n? proof.png\n")

	if state.Branch != "" || state.Untracked != 1 {
		t.Fatalf("state = %+v, want detached branch and one untracked file", state)
	}
}

func runGit(t *testing.T, directory string, arguments ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", directory}, arguments...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
}
