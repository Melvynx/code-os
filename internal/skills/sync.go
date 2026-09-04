package skills

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/melvynx/code-os/internal/config"
)

type Result struct {
	AlreadyRunning bool   `json:"alreadyRunning,omitempty"`
	Cloned         bool   `json:"cloned,omitempty"`
	Message        string `json:"message"`
}

func Sync(cfg config.Skills) (Result, error) {
	if cfg.Repository == "" || cfg.Directory == "" || cfg.Branch == "" {
		return Result{}, errors.New("skills repository, directory, and branch must be configured")
	}
	if _, err := os.Stat(cfg.Directory); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(cfg.Directory), 0o700); err != nil {
			return Result{}, fmt.Errorf("create skills parent directory: %w", err)
		}
		if err := runGit("", "clone", "--branch", cfg.Branch, "--single-branch", cfg.Repository, cfg.Directory); err != nil {
			return Result{}, err
		}
		return Result{Cloned: true, Message: "Code OS skills sync: cloned repository"}, nil
	}
	if _, err := gitOutput(cfg.Directory, "rev-parse", "--is-inside-work-tree"); err != nil {
		return Result{}, fmt.Errorf("skills directory is not a Git repository: %w", err)
	}
	remote, err := gitOutput(cfg.Directory, "remote", "get-url", "origin")
	if err != nil {
		return Result{}, errors.New("skills repository has no origin remote")
	}
	if remote != cfg.Repository {
		return Result{}, fmt.Errorf("skills origin %q does not match configured repository", remote)
	}
	branch, err := gitOutput(cfg.Directory, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil || branch != cfg.Branch {
		return Result{}, fmt.Errorf("skills checkout must be on branch %q", cfg.Branch)
	}
	gitDirectory, err := gitOutput(cfg.Directory, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return Result{}, err
	}
	lockDirectory := filepath.Join(gitDirectory, "code-os-sync.lock")
	if err := os.Mkdir(lockDirectory, 0o700); err != nil {
		if os.IsExist(err) {
			return Result{AlreadyRunning: true, Message: "Code OS skills sync: another sync is already running"}, nil
		}
		return Result{}, fmt.Errorf("create skills sync lock: %w", err)
	}
	defer os.Remove(lockDirectory)
	for _, marker := range []string{"rebase-merge", "rebase-apply", "MERGE_HEAD"} {
		if _, err := os.Stat(filepath.Join(gitDirectory, marker)); err == nil {
			return Result{}, errors.New("resolve the existing Git operation before synchronizing skills")
		}
	}
	if err := runGit(cfg.Directory, "diff", "--quiet", "--diff-filter=U"); err != nil {
		return Result{}, errors.New("resolve skills repository conflicts before synchronizing")
	}
	if err := runGit(cfg.Directory, "add", "-A"); err != nil {
		return Result{}, err
	}
	if err := runGit(cfg.Directory, "diff", "--cached", "--quiet"); err != nil {
		hostname, _ := os.Hostname()
		message := fmt.Sprintf("sync(%s): %s", hostname, time.Now().UTC().Format(time.RFC3339))
		if err := runGit(cfg.Directory, "commit", "-m", message); err != nil {
			return Result{}, err
		}
	}
	if err := runGit(cfg.Directory, "pull", "--rebase", "--autostash", "origin", cfg.Branch); err != nil {
		return Result{}, err
	}
	if err := runGit(cfg.Directory, "push", "origin", cfg.Branch); err != nil {
		return Result{}, err
	}
	return Result{Message: "Code OS skills sync: repository is up to date"}, nil
}

func gitOutput(directory string, arguments ...string) (string, error) {
	args := arguments
	if directory != "" {
		args = append([]string{"-C", directory}, arguments...)
	}
	output, err := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(output)), err
}

func runGit(directory string, arguments ...string) error {
	args := arguments
	if directory != "" {
		args = append([]string{"-C", directory}, arguments...)
	}
	command := exec.Command("git", args...)
	output, err := command.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			return fmt.Errorf("git %s failed: %w", strings.Join(arguments, " "), err)
		}
		return fmt.Errorf("git %s failed: %s", strings.Join(arguments, " "), detail)
	}
	return nil
}
