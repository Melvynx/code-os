package installbin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func IsEphemeral(path string) bool {
	cleaned := filepath.ToSlash(filepath.Clean(path))
	return strings.Contains(cleaned, "/.cleanup-") ||
		strings.Contains(cleaned, "/tmp/") ||
		strings.HasPrefix(cleaned, "/tmp") ||
		strings.Contains(cleaned, "/.tmp/")
}

func PersistentBinDir(home string) (string, error) {
	candidates := []string{"/usr/local/bin"}
	if home != "" {
		candidates = append(candidates, filepath.Join(home, ".local", "bin"))
	}
	var last error
	for _, dir := range candidates {
		if err := ensureWritableDir(dir); err != nil {
			last = err
			continue
		}
		return dir, nil
	}
	if last == nil {
		last = fmt.Errorf("no persistent bin directory")
	}
	return "", last
}

func PersistentPath(home, name string) (string, error) {
	dir, err := PersistentBinDir(home)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

func CopyExecutable(source, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create bin directory: %w", err)
	}
	in, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open %s: %w", source, err)
	}
	defer in.Close()
	temporary, err := os.CreateTemp(filepath.Dir(dest), "."+filepath.Base(dest)+"-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := io.Copy(temporary, in); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(0o755); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, dest); err != nil {
		return fmt.Errorf("install %s: %w", dest, err)
	}
	return nil
}

func EnsurePrivateDir(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("directory is required")
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	if err := os.Chmod(path, 0o700); err != nil {
		return fmt.Errorf("secure %s: %w", path, err)
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	return nil
}

func ensureWritableDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	probe := filepath.Join(dir, ".code-os-write-probe")
	if err := os.WriteFile(probe, []byte("ok\n"), 0o600); err != nil {
		return err
	}
	return os.Remove(probe)
}
