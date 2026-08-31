package server

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/melvynx/code-os/internal/model"
	"github.com/melvynx/code-os/internal/portly"
	"github.com/melvynx/code-os/internal/processes"
	"github.com/melvynx/code-os/internal/projects"
	"github.com/melvynx/code-os/internal/screenshots"
	"github.com/melvynx/code-os/internal/store"
)

type Service struct {
	Projects     projects.Scanner
	ProjectRoots []string
	Portly       portly.Client
	Processes    processes.Scanner
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
	agents, err := service.Processes.Scan()
	if err != nil {
		snapshot.Warnings = append(snapshot.Warnings, err.Error())
	} else {
		snapshot.Agents = agents
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

func (service *Service) StopApplication(ctx context.Context, id string) error {
	service.mutex.RLock()
	allowed := false
	for _, application := range service.snapshot.Apps {
		if application.ID == id && application.State == "running" {
			allowed = true
			break
		}
	}
	service.mutex.RUnlock()
	if !allowed {
		return errors.New("running application not found")
	}
	return service.Portly.Stop(ctx, id)
}

func (service *Service) TerminateAgent(id string) error {
	service.mutex.RLock()
	allowed := false
	for _, agent := range service.snapshot.Agents {
		if agent.ID == id {
			allowed = true
			break
		}
	}
	service.mutex.RUnlock()
	if !allowed {
		return errors.New("agent process not found")
	}
	if err := service.Processes.Terminate(id); err != nil {
		return err
	}
	service.mutex.Lock()
	agents := service.snapshot.Agents[:0]
	for _, agent := range service.snapshot.Agents {
		if agent.ID != id {
			agents = append(agents, agent)
		}
	}
	service.snapshot.Agents = agents
	service.mutex.Unlock()
	return nil
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

func (service *Service) IsHealthyApplicationPort(port int) bool {
	service.mutex.RLock()
	defer service.mutex.RUnlock()
	for _, application := range service.snapshot.Apps {
		if application.Port == port && application.State == "running" && application.Healthy != nil && *application.Healthy {
			return true
		}
	}
	return false
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
