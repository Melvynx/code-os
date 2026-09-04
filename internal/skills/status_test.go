package skills

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/melvynx/code-os/internal/config"
)

func TestInspectReportsUnconfiguredSkills(t *testing.T) {
	t.Parallel()
	status := Inspect(config.Skills{})
	if status.State != StateUnconfigured || status.Synced {
		t.Fatalf("status = %+v, want unconfigured", status)
	}
}

func TestInspectReportsSyncedCheckout(t *testing.T) {
	t.Parallel()
	directory, repository := newSkillsRepo(t)
	status := Inspector{RunSystemd: missingSystemd}.Inspect(config.Skills{
		Repository: repository,
		Directory:  directory,
		Branch:     "main",
	})
	if status.State != StateSynced || !status.Synced || !status.OriginMatches {
		t.Fatalf("status = %+v, want synced", status)
	}
	if status.LastCommitMessage != "sync(test): 2026-09-01T00:00:00Z" {
		t.Fatalf("last commit = %q", status.LastCommitMessage)
	}
}

func TestInspectReportsDirtyCheckout(t *testing.T) {
	t.Parallel()
	directory, repository := newSkillsRepo(t)
	if err := os.WriteFile(filepath.Join(directory, "skills", "new.md"), []byte("draft\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status := Inspector{RunSystemd: missingSystemd}.Inspect(config.Skills{
		Repository: repository,
		Directory:  directory,
		Branch:     "main",
	})
	if status.State != StateDirty || status.Synced || status.Git.Untracked != 1 {
		t.Fatalf("status = %+v, want dirty", status)
	}
}

func TestInspectReportsOriginMismatch(t *testing.T) {
	t.Parallel()
	directory, _ := newSkillsRepo(t)
	status := Inspector{RunSystemd: missingSystemd}.Inspect(config.Skills{
		Repository: "https://github.com/other/skills.git",
		Directory:  directory,
		Branch:     "main",
	})
	if status.State != StateInvalid || status.OriginMatches {
		t.Fatalf("status = %+v, want invalid origin", status)
	}
}

func TestInspectSurfacesTimerState(t *testing.T) {
	t.Parallel()
	directory, repository := newSkillsRepo(t)
	status := Inspector{RunSystemd: func(args ...string) (string, error) {
		switch args[0] {
		case "is-enabled":
			return "enabled", nil
		case "show":
			if args[1] == "code-os-skills-sync.timer" {
				return "ActiveState=active\nLastTriggerUSec=Tue 2026-09-01 01:08:22 UTC\nNextElapseUSecReal=Tue 2026-09-01 01:10:22 UTC\n", nil
			}
			return "Result=success\nExecMainStatus=0\n", nil
		default:
			return "", os.ErrNotExist
		}
	}}.Inspect(config.Skills{Repository: repository, Directory: directory, Branch: "main"})
	if !status.TimerEnabled || !status.TimerActive || status.LastTimerResult != "success" {
		t.Fatalf("timer status = %+v", status)
	}
	if status.LastTimerAt == nil || !status.LastTimerAt.Equal(time.Date(2026, 9, 1, 1, 8, 22, 0, time.UTC)) {
		t.Fatalf("last timer = %v", status.LastTimerAt)
	}
}

func TestSyncPushesLocalSkillChanges(t *testing.T) {
	t.Parallel()
	directory, repository := newSkillsRepo(t)
	if err := os.WriteFile(filepath.Join(directory, "skills", "new.md"), []byte("draft\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Sync(config.Skills{Repository: repository, Directory: directory, Branch: "main"})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if result.AlreadyRunning || result.Message != "Code OS skills sync: repository is up to date" {
		t.Fatalf("result = %+v", result)
	}
	status := Inspector{RunSystemd: missingSystemd}.Inspect(config.Skills{Repository: repository, Directory: directory, Branch: "main"})
	if !status.Synced || status.Git.ChangeCount() != 0 {
		t.Fatalf("status after sync = %+v", status)
	}
}

func newSkillsRepo(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	origin := filepath.Join(root, "origin.git")
	runGitCommand(t, root, "init", "--bare", origin)
	directory := filepath.Join(root, "agents")
	runGitCommand(t, root, "clone", origin, directory)
	runGitCommand(t, directory, "checkout", "-b", "main")
	runGitCommand(t, directory, "config", "user.email", "test@example.com")
	runGitCommand(t, directory, "config", "user.name", "Test")
	if err := os.MkdirAll(filepath.Join(directory, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "skills", "code-os.md"), []byte("skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitCommand(t, directory, "add", ".")
	runGitCommand(t, directory, "commit", "-m", "sync(test): 2026-09-01T00:00:00Z")
	runGitCommand(t, directory, "push", "-u", "origin", "main")
	return directory, origin
}

func runGitCommand(t *testing.T, directory string, arguments ...string) {
	t.Helper()
	args := arguments
	if directory != "" {
		args = append([]string{"-C", directory}, arguments...)
	}
	command := exec.Command("git", args...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
}

func missingSystemd(args ...string) (string, error) {
	return "", os.ErrNotExist
}
