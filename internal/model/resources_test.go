package model

import (
	"testing"
	"time"
)

func TestSampleResourcesKeepsRunningAppsAndAgents(t *testing.T) {
	t.Parallel()
	pid := 88
	sample := SampleResources(Snapshot{
		GeneratedAt: time.Date(2026, 9, 1, 1, 20, 0, 0, time.UTC),
		Apps: []Application{
			{ID: "srv_web", ProjectName: "demo", Name: "web", State: "running", PID: pid, CPUPercent: 12, ResidentMemoryBytes: 200},
			{ID: "srv_old", ProjectName: "demo", Name: "old", State: "stopped", CPUPercent: 9, MemoryBytes: 100},
		},
		Agents: []AgentProcess{{ID: "42:1", Name: "Cursor agent", CPUPercent: 0.4, MemoryBytes: 209}},
	})
	if len(sample.Rows) != 2 || sample.Rows[0].Name != "demo / web" || sample.Rows[1].Name != "Cursor agent" {
		t.Fatalf("rows = %+v", sample.Rows)
	}
}

func TestAppendResourceSampleTrimsByRetentionAndBuildsSeries(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	var history []ResourceSample
	for index := 0; index < 5; index++ {
		history = AppendResourceSample(history, ResourceSample{
			At: start.Add(time.Duration(index) * time.Hour),
			Rows: []ResourceRow{{
				ID: "cursor", Kind: "agent", Name: "Cursor agent",
				CPUPercent: float64(index), MemoryBytes: int64(100 + index),
			}},
		}, 2*time.Hour)
	}
	if len(history) != 3 {
		t.Fatalf("history = %d, want 3", len(history))
	}
	if !history[0].At.Equal(start.Add(2 * time.Hour)) {
		t.Fatalf("oldest = %s, want T+2h", history[0].At)
	}
	built := BuildResourceHistory(history)
	if built.SampleCount != 3 || built.RetentionHours != 6 || len(built.Series) != 1 || len(built.Series[0].Points) != 3 {
		t.Fatalf("history = %+v", built)
	}
	if built.Series[0].Points[0].CPUPercent != 2 || built.Series[0].Name != "Cursor agent" {
		t.Fatalf("series = %+v", built.Series[0])
	}
}

func TestBuildResourceHistoryGroupsSameName(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 9, 1, 1, 20, 0, 0, time.UTC)
	built := BuildResourceHistory([]ResourceSample{{
		At: at,
		Rows: []ResourceRow{
			{ID: "42:1", Kind: "agent", Name: "Cursor agent", CPUPercent: 1.5, MemoryBytes: 100},
			{ID: "43:1", Kind: "agent", Name: "Cursor agent", CPUPercent: 2.5, MemoryBytes: 200},
			{ID: "srv_web", Kind: "application", Name: "demo / web", CPUPercent: 8, MemoryBytes: 50},
			{ID: "44:1", Kind: "agent", Name: "Codex agent", CPUPercent: 4, MemoryBytes: 80},
		},
	}, {
		At: at.Add(time.Minute),
		Rows: []ResourceRow{
			{ID: "99:1", Kind: "agent", Name: "Cursor agent", CPUPercent: 3, MemoryBytes: 150},
			{ID: "srv_web", Kind: "application", Name: "demo / web", CPUPercent: 7, MemoryBytes: 40},
		},
	}})
	if len(built.Series) != 3 {
		t.Fatalf("series count = %d, want 3: %+v", len(built.Series), built.Series)
	}
	cursor := built.Series[0]
	if cursor.ID != "agent:Cursor agent" || cursor.Name != "Cursor agent" || len(cursor.Points) != 2 {
		t.Fatalf("cursor = %+v", cursor)
	}
	if cursor.Points[0].CPUPercent != 4 || cursor.Points[0].MemoryBytes != 300 {
		t.Fatalf("cursor first point = %+v", cursor.Points[0])
	}
	if cursor.Points[1].CPUPercent != 3 || cursor.Points[1].MemoryBytes != 150 {
		t.Fatalf("cursor second point = %+v", cursor.Points[1])
	}
}
