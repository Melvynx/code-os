package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/melvynx/code-os/internal/portly"
	"github.com/melvynx/code-os/internal/projects"
	"github.com/melvynx/code-os/internal/screenshots"
)

func TestRefreshUsesCurrentPortlyStateWhenProjectScanTimesOut(t *testing.T) {
	directory := t.TempDir()
	project := filepath.Join(directory, "project")
	if err := os.MkdirAll(filepath.Join(project, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	gitStarted := filepath.Join(directory, "git-started")
	gitBinary := filepath.Join(directory, "slow-git")
	if err := os.WriteFile(gitBinary, []byte(fmt.Sprintf("#!/bin/sh\n: > %q\nexec sleep 30\n", gitStarted)), 0o700); err != nil {
		t.Fatal(err)
	}
	statusFile := filepath.Join(directory, "portly-status.json")
	portlyBinary := filepath.Join(directory, "portly")
	if err := os.WriteFile(portlyBinary, []byte(fmt.Sprintf("#!/bin/sh\nexec cat %q\n", statusFile)), 0o700); err != nil {
		t.Fatal(err)
	}
	service := &Service{
		Projects:     projects.Scanner{GitBinary: gitBinary},
		ProjectRoots: []string{project},
		Portly:       portly.Client{Binary: portlyBinary},
		Screenshots:  screenshots.Indexer{Root: directory},
	}

	writePortlyStatus := func(state string, healthy bool) {
		t.Helper()
		status := fmt.Sprintf(`{"projects":[{"servers":[{"id":"app","port":3000,"state":%q,"healthy":%t}]}]}`, state, healthy)
		if err := os.WriteFile(statusFile, []byte(status), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	refresh := func() {
		t.Helper()
		if err := os.Remove(gitStarted); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan struct{})
		go func() {
			service.Refresh(ctx)
			close(done)
		}()
		deadline := time.NewTimer(10 * time.Second)
		defer deadline.Stop()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				t.Fatal("refresh completed before Git scan started")
			case <-ticker.C:
				if _, err := os.Stat(gitStarted); err == nil {
					cancel()
					<-done
					return
				}
			case <-deadline.C:
				cancel()
				<-done
				t.Fatal("Git scan never started")
			}
		}
	}

	writePortlyStatus("running", true)
	refresh()
	if !service.IsHealthyApplicationPort(3000) {
		t.Fatal("healthy Portly app disappeared after slow Git scan")
	}

	writePortlyStatus("stopped", false)
	refresh()
	if service.IsHealthyApplicationPort(3000) {
		t.Fatal("stopped Portly app is still exposed by gateway")
	}

	writePortlyStatus("running", true)
	refresh()
	if !service.IsHealthyApplicationPort(3000) {
		t.Fatal("running Portly app did not recover")
	}
	if err := os.Remove(statusFile); err != nil {
		t.Fatal(err)
	}
	refresh()
	if service.IsHealthyApplicationPort(3000) {
		t.Fatal("Portly failure left a stale app exposed by gateway")
	}
}
