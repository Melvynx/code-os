package installbin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsEphemeral(t *testing.T) {
	t.Parallel()
	if !IsEphemeral("/root/.cleanup-code-os-20260908044922/bin/code-os") {
		t.Fatal("cleanup path should be ephemeral")
	}
	if !IsEphemeral("/tmp/code-os") {
		t.Fatal("tmp path should be ephemeral")
	}
	if IsEphemeral("/usr/local/bin/portly") || IsEphemeral("/root/.local/bin/code-os") {
		t.Fatal("persistent paths marked ephemeral")
	}
}

func TestCopyExecutableAndPrivateDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	source := filepath.Join(root, "src")
	if err := os.WriteFile(source, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(root, "bin", "tool")
	if err := CopyExecutable(source, dest); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dest)
	if err != nil || info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("dest = %v mode=%v", err, info)
	}
	dir := filepath.Join(root, "files")
	if err := EnsurePrivateDir(dir); err != nil {
		t.Fatal(err)
	}
	info, err = os.Stat(dir)
	if err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("dir mode = %o", info.Mode().Perm())
	}
}
