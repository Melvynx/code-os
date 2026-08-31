package server

import (
	"path/filepath"

	"github.com/melvynx/code-os/internal/config"
	"github.com/melvynx/code-os/internal/portly"
	"github.com/melvynx/code-os/internal/processes"
	"github.com/melvynx/code-os/internal/projects"
	"github.com/melvynx/code-os/internal/screenshots"
	"github.com/melvynx/code-os/internal/store"
)

func NewService(cfg config.Config, database *store.Store) *Service {
	return &Service{
		Projects:     projects.NewScanner(),
		ProjectRoots: cfg.ProjectsRoots,
		Portly: portly.Client{
			Binary:         cfg.PortlyBinary,
			PublicPortHost: cfg.PublicPortHost,
		},
		Processes:   processes.Scanner{},
		Screenshots: screenshots.Indexer{Root: cfg.ScreenshotsRoot},
		Store:       database,
		media:       make(map[string]string),
	}
}

func DatabasePath(cfg config.Config) string {
	return filepath.Join(cfg.DataDir, "code-os.db")
}
