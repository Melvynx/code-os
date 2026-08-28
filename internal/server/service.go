package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/melvynx/stackenv/internal/model"
	"github.com/melvynx/stackenv/internal/portly"
	"github.com/melvynx/stackenv/internal/projects"
	"github.com/melvynx/stackenv/internal/screenshots"
	"github.com/melvynx/stackenv/internal/store"
)

type Service struct {
	Projects     projects.Scanner
	ProjectRoots []string
	Portly       portly.Client
	Screenshots  screenshots.Indexer
	Store        *store.Store

	mutex    sync.RWMutex
	snapshot model.Snapshot
	media    map[string]string
}

func (service *Service) Refresh(ctx context.Context) model.Snapshot {
	snapshot := model.Snapshot{GeneratedAt: time.Now().UTC()}
	projectsResult, projectWarnings := service.Projects.Scan(ctx, service.ProjectRoots)
	snapshot.Projects = projectsResult
	snapshot.Warnings = append(snapshot.Warnings, projectWarnings...)

	applications, err := service.Portly.Applications(ctx)
	if err != nil {
		snapshot.Warnings = append(snapshot.Warnings, err.Error())
	} else {
		snapshot.Apps = applications
	}

	images, err := service.Screenshots.Scan()
	if err != nil {
		snapshot.Warnings = append(snapshot.Warnings, err.Error())
	} else {
		snapshot.Screenshots = images
	}

	media := make(map[string]string, len(snapshot.Screenshots))
	for _, image := range snapshot.Screenshots {
		media[image.ID] = image.Path
	}
	service.mutex.Lock()
	service.snapshot = snapshot
	service.media = media
	service.mutex.Unlock()

	if service.Store != nil {
		if err := service.Store.Save(ctx, snapshot); err != nil {
			service.mutex.Lock()
			service.snapshot.Warnings = append(service.snapshot.Warnings, fmt.Sprintf("persist snapshot: %v", err))
			snapshot = service.snapshot
			service.mutex.Unlock()
		}
	}
	return snapshot
}

func (service *Service) Snapshot() model.Snapshot {
	service.mutex.RLock()
	defer service.mutex.RUnlock()
	return service.snapshot
}

func (service *Service) MediaPath(id string) (string, bool) {
	service.mutex.RLock()
	defer service.mutex.RUnlock()
	path, exists := service.media[id]
	return path, exists
}

func (service *Service) RunRefreshLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refreshContext, cancel := context.WithTimeout(ctx, interval)
			service.Refresh(refreshContext)
			cancel()
		}
	}
}
