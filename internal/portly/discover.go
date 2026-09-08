package portly

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DiscoveredProject struct {
	Name   string
	Path   string
	Known  bool
	Added  bool
	Detail string
}

func ListGitRoots(root string) ([]string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var roots []string
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		path := filepath.Join(root, entry.Name())
		if info, err := os.Stat(filepath.Join(path, ".git")); err == nil && (info.IsDir() || info.Mode().IsRegular()) {
			roots = append(roots, path)
		}
	}
	return roots, nil
}

func RegisterMissingProjects(binary, root string) ([]DiscoveredProject, error) {
	paths, err := ListGitRoots(root)
	if err != nil {
		return nil, err
	}
	known := existingProjectRoots(binary)
	var discovered []DiscoveredProject
	for _, path := range paths {
		item := DiscoveredProject{Name: filepath.Base(path), Path: path}
		if known[filepath.Clean(path)] {
			item.Known = true
			item.Detail = "already registered"
			discovered = append(discovered, item)
			continue
		}
		if strings.TrimSpace(binary) == "" {
			item.Detail = "portly add-project --name " + item.Name + " --path " + path
			discovered = append(discovered, item)
			continue
		}
		if output, err := exec.Command(binary, "add-project", "--name", item.Name, "--path", path, "--json").CombinedOutput(); err != nil {
			item.Detail = strings.TrimSpace(string(output))
			if item.Detail == "" {
				item.Detail = err.Error()
			}
		} else {
			item.Added = true
			item.Detail = "registered (no servers started)"
		}
		discovered = append(discovered, item)
	}
	return discovered, nil
}

func existingProjectRoots(binary string) map[string]bool {
	known := map[string]bool{}
	if strings.TrimSpace(binary) == "" {
		return known
	}
	output, err := exec.Command(binary, "status", "--json").Output()
	if err != nil {
		return known
	}
	var status statusPayload
	if err := json.Unmarshal(output, &status); err != nil {
		return known
	}
	for _, project := range status.Projects {
		if project.Root != "" {
			known[filepath.Clean(project.Root)] = true
		}
	}
	return known
}

func FormatDiscovery(items []DiscoveredProject) string {
	if len(items) == 0 {
		return "No Git repositories under the projects root yet. Clone apps, then register each with portly add-project / add-server (do not start every Vite app)."
	}
	var builder strings.Builder
	builder.WriteString("Portly projects (registered without starting servers; add-server from each repo's real command/port):\n")
	for _, item := range items {
		fmt.Fprintf(&builder, "  %s  %s  %s\n", item.Name, item.Path, item.Detail)
	}
	builder.WriteString("Copy env files from a private store into each repo (mode 0600). Never print values.\n")
	return builder.String()
}
