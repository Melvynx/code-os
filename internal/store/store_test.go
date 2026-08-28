package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/melvynx/stackenv/internal/model"
)

func TestSnapshotRoundTrip(t *testing.T) {
	t.Parallel()
	store, err := Open(filepath.Join(t.TempDir(), "stackenv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
	want := model.Snapshot{GeneratedAt: time.Now().UTC(), Projects: []model.Project{{ID: "one", Name: "One"}}}
	if err := store.Save(context.Background(), want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got.Projects) != 1 || got.Projects[0].ID != "one" {
		t.Fatalf("Load() = %+v", got)
	}
}
