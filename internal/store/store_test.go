package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/melvynx/code-os/internal/model"
)

func TestSnapshotRoundTrip(t *testing.T) {
	t.Parallel()
	store, err := Open(filepath.Join(t.TempDir(), "code-os.db"))
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

func TestResourceHistoryRoundTrip(t *testing.T) {
	t.Parallel()
	store, err := Open(filepath.Join(t.TempDir(), "code-os.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
	empty, err := store.LoadResources(context.Background())
	if err != nil || empty != nil {
		t.Fatalf("LoadResources() empty = %+v err %v", empty, err)
	}
	want := []model.ResourceSample{{
		At: time.Date(2026, 9, 1, 2, 0, 0, 0, time.UTC),
		Rows: []model.ResourceRow{{
			ID: "cursor", Kind: "agent", Name: "Cursor agent", CPUPercent: 0.4, MemoryBytes: 209,
		}},
	}}
	if err := store.SaveResources(context.Background(), want); err != nil {
		t.Fatalf("SaveResources() error = %v", err)
	}
	got, err := store.LoadResources(context.Background())
	if err != nil {
		t.Fatalf("LoadResources() error = %v", err)
	}
	if len(got) != 1 || got[0].Rows[0].Name != "Cursor agent" {
		t.Fatalf("LoadResources() = %+v", got)
	}
}
