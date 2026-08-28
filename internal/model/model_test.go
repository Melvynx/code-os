package model

import "testing"

func TestSummarizeCountsEveryWorktreeAndModifiedProject(t *testing.T) {
	t.Parallel()
	snapshot := Snapshot{Projects: []Project{
		{
			Name: "one",
			Worktrees: []Worktree{
				{Main: true, Git: GitState{Modified: 1}},
				{Git: GitState{Untracked: 2}},
				{Git: GitState{}},
			},
		},
		{Name: "legacy", Git: GitState{Deleted: 1}},
	}}

	overview := Summarize(snapshot)
	if overview.ProjectCount != 2 || overview.ModifiedProjects != 2 {
		t.Fatalf("project counts = %d/%d, want 2/2", overview.ProjectCount, overview.ModifiedProjects)
	}
	if overview.WorktreeCount != 4 || overview.ModifiedWorktrees != 3 {
		t.Fatalf("worktree counts = %d/%d, want 4/3", overview.WorktreeCount, overview.ModifiedWorktrees)
	}
}
