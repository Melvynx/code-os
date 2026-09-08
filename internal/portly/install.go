package portly

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/melvynx/code-os/internal/installbin"
)

const sourceRepository = "https://github.com/Melvynx/portly.git"

func InstallLinux(home, requested string) (string, error) {
	if path := existingPersistent(requested); path != "" {
		return path, nil
	}
	dest, err := installbin.PersistentPath(home, "portly")
	if err != nil {
		return "", fmt.Errorf("resolve Portly install path: %w", err)
	}
	if installbin.IsEphemeral(dest) {
		return "", fmt.Errorf("refusing ephemeral Portly path %s", dest)
	}
	if _, err := os.Stat(dest); err == nil && !installbin.IsEphemeral(dest) {
		return dest, nil
	}
	if err := buildFromSource(dest); err != nil {
		return "", err
	}
	return dest, nil
}

func EnableForever(binary string) error {
	if strings.TrimSpace(binary) == "" {
		return nil
	}
	if output, err := exec.Command(binary, "forever", "enable").CombinedOutput(); err != nil {
		return fmt.Errorf("portly forever enable: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func DaemonHealthy(binary string) bool {
	if strings.TrimSpace(binary) == "" {
		return false
	}
	return exec.Command(binary, "status", "--json").Run() == nil
}

func existingPersistent(requested string) string {
	candidates := []string{requested, "portly", "/usr/local/bin/portly"}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".local", "bin", "portly"))
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		path := candidate
		if !strings.ContainsRune(candidate, filepath.Separator) {
			resolved, err := exec.LookPath(candidate)
			if err != nil {
				continue
			}
			path = resolved
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() || installbin.IsEphemeral(path) {
			continue
		}
		return path
	}
	return ""
}

func buildFromSource(dest string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("portly is not installed and go is unavailable; install the headless CLI to %s (https://github.com/Melvynx/portly/tree/main/cli)", dest)
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("portly is not installed and git is unavailable")
	}
	staging, err := os.MkdirTemp("", "code-os-portly-src-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	if output, err := exec.Command("git", "clone", "--depth", "1", sourceRepository, staging).CombinedOutput(); err != nil {
		return fmt.Errorf("clone Portly: %s", strings.TrimSpace(string(output)))
	}
	command := exec.Command("go", "build", "-o", dest, ".")
	command.Dir = filepath.Join(staging, "cli")
	command.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+runtime.GOOS, "GOARCH="+runtime.GOARCH)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("build Portly: %s", strings.TrimSpace(string(output)))
	}
	return os.Chmod(dest, 0o755)
}
