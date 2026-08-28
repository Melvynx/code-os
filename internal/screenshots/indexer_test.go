package screenshots

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanIndexesImagesAndIgnoresOtherFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	directory := filepath.Join(root, "skills-verify", "lumail")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "home.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "notes.txt"), []byte("text"), 0o644); err != nil {
		t.Fatal(err)
	}
	images, err := (Indexer{Root: root}).Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("images = %d, want 1", len(images))
	}
	if images[0].Project != "lumail" || images[0].Group != "skills-verify" {
		t.Fatalf("image metadata = %+v", images[0])
	}
}
