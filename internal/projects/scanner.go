package projects

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/melvynx/stackenv/internal/model"
)

const maxDiscoveryDepth = 4

var skippedDirectories = map[string]bool{
	".cache": true, ".next": true, ".turbo": true, "build": true,
	"dist": true, "node_modules": true, "target": true, "vendor": true,
}

type Scanner struct {
	GitBinary string
}

func NewScanner() Scanner {
	return Scanner{GitBinary: "git"}
}

func (scanner Scanner) Scan(ctx context.Context, roots []string) ([]model.Project, []string) {
	var repositoryPaths []string
	var warnings []string
	seen := make(map[string]bool)
	for _, root := range roots {
		paths, err := discoverRepositories(root)
		if err != nil {
			warnings = append(warnings, err.Error())
			continue
		}
		for _, path := range paths {
			if !seen[path] {
				seen[path] = true
				repositoryPaths = append(repositoryPaths, path)
			}
		}
	}
	sort.Strings(repositoryPaths)

	projects := make([]model.Project, 0, len(repositoryPaths))
	seenRepositories := make(map[string]bool)
	for _, candidatePath := range repositoryPaths {
		records, err := scanner.worktreeRecords(ctx, candidatePath)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("inspect worktrees %s: %v", candidatePath, err))
			records = []worktreeRecord{{Path: candidatePath, Main: true}}
		}
		repositoryPath := candidatePath
		if len(records) > 0 && records[0].Path != "" {
			repositoryPath = records[0].Path
		}
		if seenRepositories[repositoryPath] {
			continue
		}
		seenRepositories[repositoryPath] = true

		project, projectWarnings, err := scanner.inspect(ctx, repositoryPath, records)
		warnings = append(warnings, projectWarnings...)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("inspect %s: %v", repositoryPath, err))
			continue
		}
		projects = append(projects, project)
	}
	return projects, warnings
}

func discoverRepositories(root string) ([]string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve projects root %s: %w", root, err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("read projects root %s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("projects root is not a directory: %s", root)
	}

	var repositories []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		depth := 0
		if relative != "." {
			depth = len(strings.Split(relative, string(filepath.Separator)))
		}
		if depth > maxDiscoveryDepth || (path != root && skippedDirectories[entry.Name()]) {
			return filepath.SkipDir
		}
		gitPath := filepath.Join(path, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			repositories = append(repositories, path)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover repositories in %s: %w", root, err)
	}
	return repositories, nil
}

func (scanner Scanner) inspect(ctx context.Context, path string, records []worktreeRecord) (model.Project, []string, error) {
	git, err := scanner.gitState(ctx, path)
	if err != nil {
		return model.Project{}, nil, err
	}
	worktrees, warnings := scanner.inspectWorktrees(ctx, records)
	return model.Project{
		ID:          stableID(path),
		Name:        filepath.Base(path),
		Path:        path,
		Git:         git,
		Worktrees:   worktrees,
		Subprojects: discoverSubprojects(path),
	}, warnings, nil
}

type worktreeRecord struct {
	Path     string
	Head     string
	Branch   string
	Main     bool
	Locked   bool
	Prunable bool
}

func (scanner Scanner) worktreeRecords(ctx context.Context, path string) ([]worktreeRecord, error) {
	command := exec.CommandContext(ctx, scanner.GitBinary, "-C", path, "worktree", "list", "--porcelain")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}
	return parseWorktreeList(string(output)), nil
}

func parseWorktreeList(output string) []worktreeRecord {
	var records []worktreeRecord
	var current *worktreeRecord
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			if current != nil {
				records = append(records, *current)
			}
			current = &worktreeRecord{Path: strings.TrimPrefix(line, "worktree "), Main: len(records) == 0}
		case current == nil:
			continue
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		case line == "locked" || strings.HasPrefix(line, "locked "):
			current.Locked = true
		case line == "prunable" || strings.HasPrefix(line, "prunable "):
			current.Prunable = true
		}
	}
	if current != nil {
		records = append(records, *current)
	}
	return records
}

func (scanner Scanner) inspectWorktrees(ctx context.Context, records []worktreeRecord) ([]model.Worktree, []string) {
	worktrees := make([]model.Worktree, 0, len(records))
	var warnings []string
	for _, record := range records {
		git, err := scanner.gitState(ctx, record.Path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("inspect worktree %s: %v", record.Path, err))
			git.Branch = record.Branch
		}
		worktrees = append(worktrees, model.Worktree{
			ID: stableID(record.Path), Path: record.Path, Head: record.Head, Main: record.Main,
			Locked: record.Locked, Prunable: record.Prunable, Git: git,
		})
	}
	return worktrees, warnings
}

func (scanner Scanner) gitState(ctx context.Context, path string) (model.GitState, error) {
	command := exec.CommandContext(ctx, scanner.GitBinary, "-C", path, "status", "--porcelain=v2", "--branch", "--untracked-files=normal")
	output, err := command.Output()
	if err != nil {
		return model.GitState{}, fmt.Errorf("git status: %w", err)
	}
	return parseGitStatus(string(output)), nil
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
				state.Ahead = signedNumber(fields[2])
				state.Behind = -signedNumber(fields[3])
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

func signedNumber(value string) int {
	number, _ := strconv.Atoi(value)
	return number
}

func discoverSubprojects(root string) []model.Subproject {
	candidates := make(map[string]model.Subproject)
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || !entry.IsDir() {
			return walkErr
		}
		if path != root && skippedDirectories[entry.Name()] {
			return filepath.SkipDir
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		if relative != "." && len(strings.Split(relative, string(filepath.Separator))) > 3 {
			return filepath.SkipDir
		}
		kind := projectKind(path)
		if kind == "" || path == root {
			return nil
		}
		candidates[path] = model.Subproject{Name: filepath.Base(path), Path: path, Kind: kind}
		return nil
	})
	result := make([]model.Subproject, 0, len(candidates))
	for _, candidate := range candidates {
		result = append(result, candidate)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

func projectKind(path string) string {
	markers := []struct {
		file string
		kind string
	}{
		{"package.json", "node"}, {"go.mod", "go"}, {"Cargo.toml", "rust"},
		{"pyproject.toml", "python"}, {"docker-compose.yml", "compose"}, {"compose.yml", "compose"},
	}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(path, marker.file)); err == nil {
			if marker.file == "package.json" && !isPackage(filepath.Join(path, marker.file)) {
				continue
			}
			return marker.kind
		}
	}
	return ""
}

func isPackage(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var payload map[string]json.RawMessage
	return json.Unmarshal(data, &payload) == nil
}

func stableID(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:8])
}

func IsMissingGit(err error) bool {
	return errors.Is(err, exec.ErrNotFound)
}
