package model

import "time"

type GitState struct {
	Branch    string `json:"branch"`
	Upstream  string `json:"upstream,omitempty"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
	Modified  int    `json:"modified"`
	Added     int    `json:"added"`
	Deleted   int    `json:"deleted"`
	Untracked int    `json:"untracked"`
	Conflicts int    `json:"conflicts"`
}

func (g GitState) ChangeCount() int {
	return g.Modified + g.Added + g.Deleted + g.Untracked + g.Conflicts
}

type Subproject struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Kind string `json:"kind"`
}

type Worktree struct {
	ID       string   `json:"id"`
	Path     string   `json:"path"`
	Head     string   `json:"head,omitempty"`
	Main     bool     `json:"main"`
	Locked   bool     `json:"locked"`
	Prunable bool     `json:"prunable"`
	Git      GitState `json:"git"`
}

type Project struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Path        string       `json:"path"`
	Git         GitState     `json:"git"`
	Worktrees   []Worktree   `json:"worktrees"`
	Subprojects []Subproject `json:"subprojects"`
}

type Application struct {
	ID                  string  `json:"id"`
	ProjectID           string  `json:"projectId"`
	ProjectName         string  `json:"projectName"`
	Name                string  `json:"name"`
	Command             string  `json:"command"`
	Directory           string  `json:"directory,omitempty"`
	State               string  `json:"state"`
	Port                int     `json:"port"`
	PID                 int     `json:"pid,omitempty"`
	Healthy             *bool   `json:"healthy,omitempty"`
	URL                 string  `json:"url,omitempty"`
	PublicURL           string  `json:"publicUrl,omitempty"`
	CPUPercent          float64 `json:"cpuPercent"`
	MemoryBytes         int64   `json:"memoryBytes"`
	ResidentMemoryBytes int64   `json:"residentMemoryBytes"`
	RestartCount        int     `json:"restartCount"`
}

type AgentProcess struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Command      string  `json:"command"`
	PID          int     `json:"pid"`
	CPUPercent   float64 `json:"cpuPercent"`
	MemoryBytes  int64   `json:"memoryBytes"`
	ProcessCount int     `json:"processCount"`
}

type Screenshot struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"-"`
	URL       string    `json:"url"`
	Project   string    `json:"project,omitempty"`
	Group     string    `json:"group,omitempty"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

type Snapshot struct {
	GeneratedAt time.Time      `json:"generatedAt"`
	Projects    []Project      `json:"projects"`
	Apps        []Application  `json:"applications"`
	Agents      []AgentProcess `json:"agents"`
	Screenshots []Screenshot   `json:"screenshots"`
	Warnings    []string       `json:"warnings"`
}

type Overview struct {
	GeneratedAt       time.Time `json:"generatedAt"`
	ProjectCount      int       `json:"projectCount"`
	ModifiedProjects  int       `json:"modifiedProjects"`
	WorktreeCount     int       `json:"worktreeCount"`
	ModifiedWorktrees int       `json:"modifiedWorktrees"`
	RunningApps       int       `json:"runningApps"`
	UnhealthyApps     int       `json:"unhealthyApps"`
	ScreenshotCount   int       `json:"screenshotCount"`
	WarningCount      int       `json:"warningCount"`
}

func Summarize(snapshot Snapshot) Overview {
	overview := Overview{
		GeneratedAt:     snapshot.GeneratedAt,
		ProjectCount:    len(snapshot.Projects),
		ScreenshotCount: len(snapshot.Screenshots),
		WarningCount:    len(snapshot.Warnings),
	}
	for _, project := range snapshot.Projects {
		worktrees := project.Worktrees
		if len(worktrees) == 0 {
			worktrees = []Worktree{{Git: project.Git, Main: true}}
		}
		overview.WorktreeCount += len(worktrees)
		projectModified := false
		for _, worktree := range worktrees {
			if worktree.Git.ChangeCount() > 0 {
				overview.ModifiedWorktrees++
				projectModified = true
			}
		}
		if projectModified {
			overview.ModifiedProjects++
		}
	}
	for _, app := range snapshot.Apps {
		if app.State == "running" {
			overview.RunningApps++
		}
		if app.Healthy != nil && !*app.Healthy {
			overview.UnhealthyApps++
		}
	}
	return overview
}
