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
	for _, path := range repositoryPaths {
		project, err := scanner.inspect(ctx, path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("inspect %s: %v", path, err))
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

func (scanner Scanner) inspect(ctx context.Context, path string) (model.Project, error) {
	git, err := scanner.gitState(ctx, path)
	if err != nil {
		return model.Project{}, err
	}
	return model.Project{
		ID:          stableID(path),
		Name:        filepath.Base(path),
		Path:        path,
		Git:         git,
		Subprojects: discoverSubprojects(path),
	}, nil
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
			state.Branch = strings.TrimPrefix(line, "# branch.head ")
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
