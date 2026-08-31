package portly

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/melvynx/code-os/internal/model"
)

type Client struct {
	Binary         string
	PublicPortHost string
}

type statusPayload struct {
	Projects []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Root    string `json:"root"`
		Servers []struct {
			ID                  string  `json:"id"`
			Name                string  `json:"name"`
			ProjectID           string  `json:"projectID"`
			ProjectName         string  `json:"projectName"`
			Command             string  `json:"command"`
			Directory           string  `json:"directory"`
			State               string  `json:"state"`
			Port                int     `json:"port"`
			PID                 int     `json:"pid"`
			Healthy             *bool   `json:"healthy"`
			URL                 string  `json:"url"`
			CPUPercent          float64 `json:"cpuPercent"`
			MemoryBytes         int64   `json:"memoryBytes"`
			ResidentMemoryBytes int64   `json:"residentMemoryBytes"`
			RestartCount        int     `json:"restartCount"`
		} `json:"servers"`
	} `json:"projects"`
}

func (client Client) Applications(ctx context.Context) ([]model.Application, error) {
	command := exec.CommandContext(ctx, client.Binary, "status", "--json")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("portly status: %w", err)
	}
	var status statusPayload
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, fmt.Errorf("decode portly status: %w", err)
	}
	var applications []model.Application
	for _, project := range status.Projects {
		for _, server := range project.Servers {
			projectID := server.ProjectID
			if projectID == "" {
				projectID = project.ID
			}
			projectName := server.ProjectName
			if projectName == "" {
				projectName = project.Name
			}
			publicURL := ""
			if server.Port > 0 && client.PublicPortHost != "" {
				publicURL = "https://" + strings.ReplaceAll(client.PublicPortHost, "{port}", fmt.Sprintf("%d", server.Port))
			}
			applications = append(applications, model.Application{
				ID: server.ID, ProjectID: projectID, ProjectName: projectName,
				Name: server.Name, Command: server.Command, Directory: server.Directory,
				State: server.State, Port: server.Port, PID: server.PID, Healthy: server.Healthy,
				URL: server.URL, PublicURL: publicURL, CPUPercent: server.CPUPercent,
				MemoryBytes: server.MemoryBytes, ResidentMemoryBytes: server.ResidentMemoryBytes,
				RestartCount: server.RestartCount,
			})
		}
	}
	return applications, nil
}

func (client Client) Stop(ctx context.Context, id string) error {
	command := exec.CommandContext(ctx, client.Binary, "stop", id, "--json")
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("portly stop: %s", message)
	}
	return nil
}
