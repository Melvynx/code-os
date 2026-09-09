package processes

import (
	"testing"
	"time"
)

func TestCPUPercentUsesProcessTreeLifetime(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 9, 2, 0, 0, 0, time.UTC)
	// 100 ticks = 1s CPU over 10s of life (startTime 1000 ticks, uptime 20s)
	if got := cpuPercent(100, 1_000, CPUMark{}, now, 20); got != 10 {
		t.Fatalf("lifetime cpu = %v, want 10", got)
	}
	if got := cpuPercent(13635, 0, CPUMark{}, now, 598); got < 20 || got > 25 {
		t.Fatalf("lumail-scale tree cpu = %v, want ~22.8", got)
	}
}
