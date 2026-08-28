package server

import (
	"path/filepath"

	"github.com/melvynx/stackenv/internal/config"
	"github.com/melvynx/stackenv/internal/portly"
	"github.com/melvynx/stackenv/internal/projects"
	"github.com/melvynx/stackenv/internal/screenshots"
	"github.com/melvynx/stackenv/internal/store"
)

func NewService(cfg config.Config, database *store.Store) *Service {
	return &Service{
		Projects:     projects.NewScanner(),
		ProjectRoots: cfg.ProjectsRoots,
		Portly: portly.Client{
			Binary:         cfg.PortlyBinary,
			PublicPortHost: cfg.PublicPortHost,
		},
		Screenshots: screenshots.Indexer{Root: cfg.ScreenshotsRoot},
		Store:       database,
		media:       make(map[string]string),
	}
}

func DatabasePath(cfg config.Config) string {
	return filepath.Join(cfg.DataDir, "stackenv.db")
}
