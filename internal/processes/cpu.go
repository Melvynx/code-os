package processes

import "time"

const clockTicksPerSecond = 100

type CPUMark struct {
	Ticks uint64
	At    time.Time
}

func cpuPercent(ticks uint64, startTime uint64, previous CPUMark, now time.Time, uptime float64) float64 {
	if !previous.At.IsZero() && now.After(previous.At) && ticks >= previous.Ticks {
		elapsed := now.Sub(previous.At).Seconds()
		if elapsed >= 0.2 {
			return float64(ticks-previous.Ticks) / clockTicksPerSecond / elapsed * 100
		}
	}
	lifetime := uptime - float64(startTime)/clockTicksPerSecond
	if lifetime > 0 {
		return float64(ticks) / clockTicksPerSecond / lifetime * 100
	}
	return 0
}
