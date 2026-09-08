package agents

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/melvynx/code-os/internal/config"
)

const (
	skillMarker           = "<!-- code-os:portly-skill -->"
	webviewMarker         = "<!-- code-os:remote-webview-skill -->"
	verifyMarker          = "<!-- code-os:verify-skill -->"
	operateMarker         = "<!-- code-os:operate-skill -->"
	ruleStartMarker       = "<!-- portly:managed-rule:start -->"
	ruleEndMarker         = "<!-- portly:managed-rule:end -->"
	ruleFingerprint       = "exits with code `124`"
	machineStartMarker    = "<!-- code-os:machine-rule:start -->"
	machineEndMarker      = "<!-- code-os:machine-rule:end -->"
	machineFingerprint    = "publicPortHost"
	cursorRuleName        = "portly.mdc"
	cursorMachineRuleName = "code-os.mdc"
)

//go:embed portly-skill remote-webview-skill verify-skill code-os-skill
var bundledSkills embed.FS

var bundledSkillNames = []string{"portly", "remote-webview", "verify", "code-os"}

const managedRule = ruleStartMarker + `
## Development servers

- Always use Portly (` + "`portly ...`" + `) to start, stop, restart, inspect, or keep local development servers running.
- Start with ` + "`portly status`" + `. Use ` + "`portly status --details`" + ` only for the full inventory and metrics, and ` + "`--json`" + ` only for machine-readable fields. Reuse a healthy managed server; if an in-scope server is running outside Portly, register it and use ` + "`portly take-over <project/server> --json`" + `.
- For long-lived or reusable work, create a project and server.
- For builds, tests, code generation, previews, demos, and other bounded one-off work, run ` + "`job_id=\"$(portly temp '<command>' --path <folder> --timeout 30m)\"`" + `, then ` + "`portly wait \"$job_id\"`" + `. ` + "`temp`" + ` returns immediately with an ID; ` + "`wait`" + ` prints captured logs and exits with the command's real code. A timeout kills the whole process group and exits with code ` + "`124`" + `.
- Never launch persistent development servers directly, in the background, or through another supervisor.
` + ruleEndMarker

const machineRule = machineStartMarker + `
## Code OS machine

Read ` + "`~/.config/code-os/config.json`" + `. Hosts are configuration, never constants.

- Development servers: Portly only. Never ` + "`pnpm dev`" + `, Vite, or another supervisor.
- Public app URLs: replace ` + "`{port}`" + ` in ` + "`publicPortHost`" + `. Example: ` + "`port{port}.example.com`" + ` → ` + "`https://port3013.example.com/`" + `. Never invent a CDN host.
- Screenshots: write under ` + "`filesRoot/skills-verify/<run-id>/`" + `. Codex embeds the absolute path. Cursor embeds ` + "`https://<dashboardHost>/files/...?bp=`" + `. Same key must 401 on ` + "`/api/health`" + `.
` + machineEndMarker

type Result struct {
	SkillPath string   `json:"skillPath"`
	Skills    []string `json:"skills"`
	RuleFiles []string `json:"ruleFiles"`
	Message   string   `json:"message"`
}

type Status struct {
	SkillInstalled        bool   `json:"skillInstalled"`
	RulesInstalled        bool   `json:"rulesInstalled"`
	PortsConfigured       bool   `json:"portsConfigured"`
	ScreenshotsConfigured bool   `json:"screenshotsConfigured"`
	SkillPath             string `json:"skillPath"`
	Detail                string `json:"detail"`
}

func Install(cfg config.Config) (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, fmt.Errorf("resolve home directory: %w", err)
	}
	return InstallAt(home, skillsDirectory(cfg, home))
}

func InstallAt(home, skillsDir string) (Result, error) {
	if home == "" {
		return Result{}, fmt.Errorf("home directory is required")
	}
	if skillsDir == "" {
		skillsDir = filepath.Join(home, ".agents")
	}
	installed := make([]string, 0, len(bundledSkillNames))
	for _, name := range bundledSkillNames {
		destinations := uniqueSkillDestinations(home, skillsDir, name)
		for _, dest := range destinations {
			if err := installBundledSkill(name, dest); err != nil {
				return Result{}, err
			}
		}
		installed = append(installed, filepath.Join(skillsDir, "skills", name))
	}

	ruleFiles := []string{
		filepath.Join(skillsDir, "AGENTS.md"),
		filepath.Join(home, ".claude", "CLAUDE.md"),
		filepath.Join(home, ".cursor", "rules", cursorRuleName),
		filepath.Join(home, ".cursor", "rules", cursorMachineRuleName),
	}
	written := make([]string, 0, len(ruleFiles))
	for _, path := range ruleFiles {
		if err := installRuleFile(path); err != nil {
			return Result{}, err
		}
		written = append(written, path)
	}
	return Result{
		SkillPath: filepath.Join(skillsDir, "skills", "portly"),
		Skills:    installed,
		RuleFiles: written,
		Message:   "Installed Portly, public-port, and screenshot agent skills",
	}, nil
}

func Inspect(cfg config.Config, home string) Status {
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return Status{Detail: "could not resolve home directory"}
		}
	}
	skillsDir := skillsDirectory(cfg, home)
	skillRoot := filepath.Join(skillsDir, "skills")
	status := Status{SkillPath: filepath.Join(skillRoot, "portly", "SKILL.md")}
	status.SkillInstalled = skillHasMarker(status.SkillPath)
	status.RulesInstalled = hasCurrentRule(filepath.Join(skillsDir, "AGENTS.md")) &&
		hasCurrentRule(filepath.Join(home, ".cursor", "rules", cursorRuleName))
	status.PortsConfigured = fileContains(filepath.Join(skillRoot, "remote-webview", "SKILL.md"), webviewMarker) &&
		!fileContains(filepath.Join(skillRoot, "remote-webview", "SKILL.md"), "portNNNN.mlvcdn.com") &&
		strings.Contains(cfg.PublicPortHost, "{port}")
	status.ScreenshotsConfigured = fileContains(filepath.Join(skillRoot, "verify", "SKILL.md"), verifyMarker) &&
		fileContains(filepath.Join(skillRoot, "code-os", "SKILL.md"), operateMarker)
	switch {
	case status.SkillInstalled && status.RulesInstalled && status.PortsConfigured && status.ScreenshotsConfigured:
		status.Detail = status.SkillPath
	case !status.SkillInstalled && fileContains(status.SkillPath, "Portly.app"):
		status.Detail = "macOS Portly.app skill is installed; Code OS Linux skill is missing"
	case !status.SkillInstalled:
		status.Detail = "Portly skill is not installed in the agent config"
	case !status.RulesInstalled:
		status.Detail = "Portly global agent rules are missing"
	case !status.PortsConfigured:
		status.Detail = "public port skill or port{port}.<zone> template is missing"
	default:
		status.Detail = "screenshot verify skill is missing"
	}
	return status
}

func InstallManagedRule(contents string) string {
	return replaceMarkedBlock(contents, ruleStartMarker, ruleEndMarker, managedRule)
}

func skillsDirectory(cfg config.Config, home string) string {
	if strings.TrimSpace(cfg.Skills.Directory) != "" {
		return cfg.Skills.Directory
	}
	return filepath.Join(home, ".agents")
}

func uniqueSkillDestinations(home, skillsDir, name string) []string {
	primary := filepath.Join(skillsDir, "skills", name)
	seen := map[string]struct{}{resolvedPath(primary): {}}
	destinations := []string{primary}
	for _, extra := range []string{
		filepath.Join(home, ".claude", "skills", name),
		filepath.Join(home, ".cursor", "skills", name),
	} {
		key := resolvedPath(extra)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		destinations = append(destinations, extra)
	}
	return destinations
}

func resolvedPath(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}

func installBundledSkill(name, dest string) error {
	root := name + "-skill"
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create skill parent %s: %w", dest, err)
	}
	staging := dest + ".install"
	if err := os.RemoveAll(staging); err != nil {
		return fmt.Errorf("clear skill staging %s: %w", staging, err)
	}
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return fmt.Errorf("create skill staging %s: %w", staging, err)
	}
	err := fs.WalkDir(bundledSkills, root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		target := filepath.Join(staging, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := bundledSkills.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(relative, ".py") {
			mode = 0o755
		}
		return os.WriteFile(target, data, mode)
	})
	if err != nil {
		os.RemoveAll(staging)
		return fmt.Errorf("copy bundled %s skill: %w", name, err)
	}
	if err := os.RemoveAll(dest); err != nil {
		os.RemoveAll(staging)
		return fmt.Errorf("replace %s skill at %s: %w", name, dest, err)
	}
	if err := os.Rename(staging, dest); err != nil {
		os.RemoveAll(staging)
		return fmt.Errorf("install %s skill at %s: %w", name, dest, err)
	}
	return nil
}

func installRuleFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create rule directory %s: %w", filepath.Dir(path), err)
	}
	contents := ""
	if data, err := os.ReadFile(path); err == nil {
		contents = string(data)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}
	base := filepath.Base(path)
	if (base == cursorRuleName || base == cursorMachineRuleName) && !strings.Contains(contents, "alwaysApply:") {
		contents = "---\nalwaysApply: true\n---\n\n" + strings.TrimPrefix(contents, "\n")
	}
	updated := contents
	if base == cursorMachineRuleName {
		updated = replaceMarkedBlock(updated, machineStartMarker, machineEndMarker, machineRule)
	} else {
		updated = InstallManagedRule(updated)
	}
	if updated == contents {
		return nil
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}

func replaceMarkedBlock(contents, startMarker, endMarker, block string) string {
	if start := strings.Index(contents, startMarker); start >= 0 {
		if end := strings.Index(contents[start:], endMarker); end >= 0 {
			end += start + len(endMarker)
			return contents[:start] + block + contents[end:]
		}
	}
	updated := contents
	if updated != "" && !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	if updated != "" {
		updated += "\n"
	}
	updated += block
	if !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	return updated
}

func skillHasMarker(path string) bool {
	return fileContains(path, skillMarker)
}

func hasCurrentRule(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	contents := string(data)
	return strings.Contains(contents, ruleStartMarker) && strings.Contains(contents, ruleFingerprint)
}

func fileContains(path, needle string) bool {
	data, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(data), needle)
}
