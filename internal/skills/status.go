package skills

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/melvynx/code-os/internal/config"
	"github.com/melvynx/code-os/internal/model"
)

type State string

const (
	StateUnconfigured State = "unconfigured"
	StateMissing      State = "missing"
	StateInvalid      State = "invalid"
	StateConflict     State = "conflict"
	StateDiverged     State = "diverged"
	StateBehind       State = "behind"
	StateAhead        State = "ahead"
	StateDirty        State = "dirty"
	StateLocked       State = "locked"
	StateSynced       State = "synced"
)

type Status struct {
	State             State          `json:"state"`
	Synced            bool           `json:"synced"`
	Configured        bool           `json:"configured"`
	Message           string         `json:"message"`
	Repository        string         `json:"repository,omitempty"`
	Directory         string         `json:"directory,omitempty"`
	Branch            string         `json:"branch,omitempty"`
	CurrentBranch     string         `json:"currentBranch,omitempty"`
	Origin            string         `json:"origin,omitempty"`
	OriginMatches     bool           `json:"originMatches"`
	Git               model.GitState `json:"git"`
	LastCommitHash    string         `json:"lastCommitHash,omitempty"`
	LastCommitMessage string         `json:"lastCommitMessage,omitempty"`
	LastCommitAt      *time.Time     `json:"lastCommitAt,omitempty"`
	LastSyncAt        *time.Time     `json:"lastSyncAt,omitempty"`
	TimerEnabled      bool           `json:"timerEnabled"`
	TimerActive       bool           `json:"timerActive"`
	LastTimerAt       *time.Time     `json:"lastTimerAt,omitempty"`
	NextTimerAt       *time.Time     `json:"nextTimerAt,omitempty"`
	LastTimerResult   string         `json:"lastTimerResult,omitempty"`
	LockHeld          bool           `json:"lockHeld"`
	Issues            []string       `json:"issues"`
}

type Inspector struct {
	GitBinary  string
	RunSystemd func(args ...string) (string, error)
}

func Inspect(cfg config.Skills) Status {
	return Inspector{}.Inspect(cfg)
}

func (inspector Inspector) Inspect(cfg config.Skills) Status {
	status := Status{
		Repository: cfg.Repository,
		Directory:  cfg.Directory,
		Branch:     cfg.Branch,
	}
	if cfg.Repository == "" || cfg.Directory == "" || cfg.Branch == "" {
		status.State = StateUnconfigured
		status.Message = "Skills synchronization is not configured."
		status.Issues = []string{"Set the private repository, local checkout, and branch in Settings."}
		return inspector.withTimer(status)
	}
	status.Configured = true

	info, err := os.Stat(cfg.Directory)
	if err != nil || !info.IsDir() {
		status.State = StateMissing
		status.Message = "The skills checkout does not exist yet."
		status.Issues = []string{fmt.Sprintf("Expected a Git checkout at %s.", cfg.Directory)}
		return inspector.withTimer(status)
	}

	gitDirectory, err := inspector.gitOutput(cfg.Directory, "rev-parse", "--absolute-git-dir")
	if err != nil {
		status.State = StateInvalid
		status.Message = "The skills directory is not a Git repository."
		status.Issues = []string{strings.TrimSpace(err.Error())}
		return inspector.withTimer(status)
	}
	status.LockHeld = directoryExists(filepath.Join(gitDirectory, "code-os-sync.lock"))
	status.Git = inspector.gitState(cfg.Directory)
	status.CurrentBranch = status.Git.Branch
	status.Origin, _ = inspector.gitOutput(cfg.Directory, "remote", "get-url", "origin")
	status.OriginMatches = status.Origin == cfg.Repository
	status.LastCommitHash, status.LastCommitMessage, status.LastCommitAt = inspector.lastCommit(cfg.Directory)
	status.LastSyncAt = inspector.lastSyncCommit(cfg.Directory)
	if status.LastSyncAt == nil {
		status.LastSyncAt = status.LastCommitAt
	}

	if !status.OriginMatches {
		status.Issues = append(status.Issues, fmt.Sprintf("origin is %q, expected %q", empty(status.Origin, "missing"), cfg.Repository))
	}
	if status.CurrentBranch != cfg.Branch {
		status.Issues = append(status.Issues, fmt.Sprintf("checkout is on %q, expected %q", empty(status.CurrentBranch, "detached HEAD"), cfg.Branch))
	}
	if status.Git.Conflicts > 0 || markerExists(gitDirectory, "rebase-merge", "rebase-apply", "MERGE_HEAD") {
		status.Issues = append(status.Issues, "resolve the existing Git operation before synchronizing")
	}
	if status.LockHeld {
		status.Issues = append(status.Issues, "another skills sync is already running")
	}

	status.State = deriveState(status)
	status.Synced = status.State == StateSynced
	status.Message = stateMessage(status)
	return inspector.withTimer(status)
}

func deriveState(status Status) State {
	if status.Git.Conflicts > 0 || containsIssue(status.Issues, "resolve the existing Git operation") {
		return StateConflict
	}
	if !status.OriginMatches || status.CurrentBranch != status.Branch {
		return StateInvalid
	}
	if status.LockHeld {
		return StateLocked
	}
	if status.Git.Ahead > 0 && status.Git.Behind > 0 {
		return StateDiverged
	}
	if status.Git.Behind > 0 {
		return StateBehind
	}
	if status.Git.Ahead > 0 {
		return StateAhead
	}
	if status.Git.ChangeCount() > 0 {
		return StateDirty
	}
	return StateSynced
}

func stateMessage(status Status) string {
	switch status.State {
	case StateConflict:
		return "Skills checkout has unresolved Git conflicts."
	case StateInvalid:
		return "Skills checkout does not match the configured repository."
	case StateLocked:
		return "A skills synchronization is already running."
	case StateDiverged:
		return "Skills checkout has diverged from origin."
	case StateBehind:
		return "Skills checkout is behind origin."
	case StateAhead:
		return "Local skills commits are waiting to be pushed."
	case StateDirty:
		return "Local skill changes are waiting for the next sync."
	default:
		if status.LastSyncAt != nil {
			return "Skills library is synchronized with origin."
		}
		return "Skills library matches origin."
	}
}

func (inspector Inspector) withTimer(status Status) Status {
	if runtime.GOOS != "linux" || !status.Configured {
		return status
	}
	run := inspector.RunSystemd
	if run == nil {
		run = runSystemd
	}
	enabled, _ := run("is-enabled", "code-os-skills-sync.timer")
	status.TimerEnabled = strings.TrimSpace(enabled) == "enabled"
	timer, err := run("show", "code-os-skills-sync.timer", "--property=ActiveState,LastTriggerUSec,NextElapseUSecReal")
	if err == nil {
		properties := parseSystemdProperties(timer)
		status.TimerActive = properties["ActiveState"] == "active"
		status.LastTimerAt = parseSystemdTimestamp(properties["LastTriggerUSec"])
		status.NextTimerAt = parseSystemdTimestamp(properties["NextElapseUSecReal"])
	}
	service, err := run("show", "code-os-skills-sync.service", "--property=Result,ExecMainStatus")
	if err == nil {
		properties := parseSystemdProperties(service)
		if result := properties["Result"]; result != "" && result != "n/a" {
			status.LastTimerResult = result
			if code := properties["ExecMainStatus"]; code != "" && code != "0" && code != "n/a" {
				status.LastTimerResult = result + " (" + code + ")"
			}
		}
	}
	if status.Configured && !status.TimerEnabled {
		status.Issues = append(status.Issues, "skills-sync timer is not enabled")
	}
	if status.LastTimerResult != "" && status.LastTimerResult != "success" {
		status.Issues = append(status.Issues, "last timer run: "+status.LastTimerResult)
	}
	return status
}

func (inspector Inspector) gitState(directory string) model.GitState {
	output, err := inspector.gitOutput(directory, "status", "--porcelain=v2", "--branch", "--untracked-files=normal")
	if err != nil {
		return model.GitState{}
	}
	return parseGitStatus(output)
}

func (inspector Inspector) lastCommit(directory string) (string, string, *time.Time) {
	output, err := inspector.gitOutput(directory, "log", "-1", "--format=%H%x09%cI%x09%s")
	if err != nil || output == "" {
		return "", "", nil
	}
	return parseCommitLine(output)
}

func (inspector Inspector) lastSyncCommit(directory string) *time.Time {
	output, err := inspector.gitOutput(directory, "log", "-1", "--grep=^sync(", "--format=%H%x09%cI%x09%s")
	if err != nil || output == "" {
		return nil
	}
	_, _, committedAt := parseCommitLine(output)
	return committedAt
}

func (inspector Inspector) gitOutput(directory string, arguments ...string) (string, error) {
	binary := inspector.GitBinary
	if binary == "" {
		binary = "git"
	}
	command := exec.Command(binary, append([]string{"-C", directory}, arguments...)...)
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(arguments, " "), err)
	}
	return strings.TrimSpace(string(output)), nil
}

func parseCommitLine(output string) (string, string, *time.Time) {
	parts := strings.SplitN(output, "\t", 3)
	if len(parts) < 3 {
		return "", "", nil
	}
	committedAt, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return parts[0], parts[2], nil
	}
	return parts[0], parts[2], &committedAt
}

func parseGitStatus(output string) model.GitState {
	var state model.GitState
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			branch := strings.TrimPrefix(line, "# branch.head ")
			if branch != "(detached)" && branch != "(unknown)" {
				state.Branch = branch
			}
		case strings.HasPrefix(line, "# branch.upstream "):
			state.Upstream = strings.TrimPrefix(line, "# branch.upstream ")
		case strings.HasPrefix(line, "# branch.ab "):
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				state.Ahead = atoi(fields[2])
				state.Behind = -atoi(fields[3])
			}
		case strings.HasPrefix(line, "? "):
			state.Untracked++
		case strings.HasPrefix(line, "u "):
			state.Conflicts++
		case strings.HasPrefix(line, "1 "), strings.HasPrefix(line, "2 "):
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			status := fields[1]
			if strings.Contains(status, "A") {
				state.Added++
			}
			if strings.Contains(status, "D") {
				state.Deleted++
			}
			if strings.ContainsAny(status, "MRT") {
				state.Modified++
			}
		}
	}
	return state
}

func parseSystemdProperties(output string) map[string]string {
	properties := map[string]string{}
	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			properties[key] = value
		}
	}
	return properties
}

func parseSystemdTimestamp(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" || value == "n/a" || value == "0" {
		return nil
	}
	for _, layout := range []string{"Mon 2006-01-02 15:04:05 MST", time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed
		}
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return &parsed
		}
	}
	return nil
}

func runSystemd(args ...string) (string, error) {
	command := exec.Command("systemctl", append([]string{"--user"}, args...)...)
	output, err := command.Output()
	return strings.TrimSpace(string(output)), err
}

func markerExists(gitDirectory string, names ...string) bool {
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(gitDirectory, name)); err == nil {
			return true
		}
	}
	return false
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func containsIssue(issues []string, fragment string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, fragment) {
			return true
		}
	}
	return false
}

func empty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func atoi(value string) int {
	number, _ := strconv.Atoi(value)
	return number
}
