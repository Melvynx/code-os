package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/melvynx/code-os/internal/config"
)

func TestInstallAtWritesLinuxSkillAndGlobalRules(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	skillsDir := filepath.Join(home, ".agents")
	if err := os.MkdirAll(filepath.Join(skillsDir, "skills", "portly"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "skills", "portly", "SKILL.md"), []byte("Every command launches Portly.app\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "AGENTS.md"), []byte("# Existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := InstallAt(home, skillsDir)
	if err != nil {
		t.Fatal(err)
	}
	if result.SkillPath != filepath.Join(skillsDir, "skills", "portly") {
		t.Fatalf("skill path = %q", result.SkillPath)
	}

	skill, err := os.ReadFile(filepath.Join(result.SkillPath, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(skill), skillMarker) || strings.Contains(string(skill), "Every command launches Portly.app") {
		t.Fatalf("installed skill is not the Code OS Linux skill:\n%s", skill)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "portly", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".cursor", "skills", "portly", "SKILL.md")); err != nil {
		t.Fatal(err)
	}

	agentsFile, err := os.ReadFile(filepath.Join(skillsDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(agentsFile), "# Existing") || !strings.Contains(string(agentsFile), ruleFingerprint) {
		t.Fatalf("AGENTS.md = %s", agentsFile)
	}
	cursorRule, err := os.ReadFile(filepath.Join(home, ".cursor", "rules", cursorRuleName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cursorRule), "alwaysApply: true") || !strings.Contains(string(cursorRule), ruleStartMarker) {
		t.Fatalf("cursor rule = %s", cursorRule)
	}

	again, err := InstallAt(home, skillsDir)
	if err != nil {
		t.Fatal(err)
	}
	if again.SkillPath != result.SkillPath {
		t.Fatalf("second install path = %q", again.SkillPath)
	}
	cfg := config.Config{Skills: config.Skills{Directory: skillsDir}, PublicPortHost: "port{port}.example.com"}
	status := Inspect(cfg, home)
	if !status.SkillInstalled || !status.RulesInstalled || !status.PortsConfigured || !status.ScreenshotsConfigured {
		t.Fatalf("status after install = %+v", status)
	}
	webview, err := os.ReadFile(filepath.Join(skillsDir, "skills", "remote-webview", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(webview), webviewMarker) || !strings.Contains(string(webview), "publicPortHost") {
		t.Fatalf("remote-webview skill is not domain-agnostic:\n%s", webview)
	}
	if _, err := os.Stat(filepath.Join(skillsDir, "skills", "remote-webview", "scripts", "remote_webview.py")); err != nil {
		t.Fatal(err)
	}
	machineRule, err := os.ReadFile(filepath.Join(home, ".cursor", "rules", cursorMachineRuleName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(machineRule), machineFingerprint) || !strings.Contains(string(machineRule), "alwaysApply: true") {
		t.Fatalf("machine rule = %s", machineRule)
	}
}

func TestInspectDetectsMacOSSkill(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	skill := filepath.Join(home, ".agents", "skills", "portly")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("Every command launches Portly.app automatically\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status := Inspect(config.Config{Skills: config.Skills{Directory: filepath.Join(home, ".agents")}}, home)
	if status.SkillInstalled {
		t.Fatal("macOS skill should not count as installed")
	}
	if !strings.Contains(status.Detail, "Portly.app") {
		t.Fatalf("detail = %q", status.Detail)
	}
}

func TestInstallManagedRuleReplacesExistingBlock(t *testing.T) {
	t.Parallel()
	original := "# Top\n\n" + ruleStartMarker + "\nold rule\n" + ruleEndMarker + "\n\n# Bottom\n"
	updated := InstallManagedRule(original)
	if strings.Count(updated, ruleStartMarker) != 1 || !strings.Contains(updated, ruleFingerprint) {
		t.Fatalf("updated = %s", updated)
	}
	if !strings.Contains(updated, "# Top") || !strings.Contains(updated, "# Bottom") {
		t.Fatalf("preserved content lost: %s", updated)
	}
}
