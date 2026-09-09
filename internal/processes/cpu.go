package processes

import "time"

const clockTicksPerSecond = 100

type CPUMark struct {
	Ticks uint64
	At    time.Time
}

func cpuPercent(ticks uint64, startTime uint64, _ CPUMark, _ time.Time, uptime float64) float64 {
	lifetime := uptime - float64(startTime)/clockTicksPerSecond
	if lifetime <= 0 {
		return 0
	}
	return float64(ticks) / clockTicksPerSecond / lifetime * 100
}
