package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/melvynx/stackenv/internal/model"
	_ "modernc.org/sqlite"
)

type Store struct {
	database *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	store := &Store{database: database}
	if err := store.migrate(); err != nil {
		database.Close()
		return nil, err
	}
	return store, nil
}

func (store *Store) Close() error {
	return store.database.Close()
}

func (store *Store) migrate() error {
	_, err := store.database.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS snapshots (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			generated_at TEXT NOT NULL,
			payload BLOB NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("migrate sqlite database: %w", err)
	}
	return nil
}

func (store *Store) Save(ctx context.Context, snapshot model.Snapshot) error {
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	_, err = store.database.ExecContext(ctx, `
		INSERT INTO snapshots (id, generated_at, payload) VALUES (1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET generated_at = excluded.generated_at, payload = excluded.payload
	`, snapshot.GeneratedAt.UTC().Format(time.RFC3339Nano), payload)
	if err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}
	return nil
}

func (store *Store) Load(ctx context.Context) (model.Snapshot, error) {
	var payload []byte
	err := store.database.QueryRowContext(ctx, "SELECT payload FROM snapshots WHERE id = 1").Scan(&payload)
	if err != nil {
		return model.Snapshot{}, fmt.Errorf("load snapshot: %w", err)
	}
	var snapshot model.Snapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return model.Snapshot{}, fmt.Errorf("decode snapshot: %w", err)
	}
	return snapshot, nil
}
