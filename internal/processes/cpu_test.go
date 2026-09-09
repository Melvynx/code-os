package processes

import (
	"testing"
	"time"
)

func TestCPUPercentUsesIntervalDeltaThenLifetime(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 9, 2, 0, 0, 0, time.UTC)
	previous := CPUMark{Ticks: 50, At: now.Add(-time.Second)}
	if got := cpuPercent(150, 1_000, previous, now, 20); got != 100 {
		t.Fatalf("delta cpu = %v, want 100", got)
	}
	if got := cpuPercent(100, 1_000, CPUMark{}, now, 20); got != 10 {
		t.Fatalf("lifetime cpu = %v, want 10", got)
	}
}
