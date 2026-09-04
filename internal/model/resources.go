package model

import "time"

const ResourceHistoryRetention = 6 * time.Hour

type ResourceRow struct {
	ID          string  `json:"id"`
	Kind        string  `json:"kind"`
	Name        string  `json:"name"`
	CPUPercent  float64 `json:"cpuPercent"`
	MemoryBytes int64   `json:"memoryBytes"`
}

type ResourceSample struct {
	At   time.Time     `json:"at"`
	Rows []ResourceRow `json:"rows"`
}

type ResourceSeries struct {
	ID     string          `json:"id"`
	Kind   string          `json:"kind"`
	Name   string          `json:"name"`
	Points []ResourcePoint `json:"points"`
}

type ResourcePoint struct {
	At          time.Time `json:"at"`
	CPUPercent  float64   `json:"cpuPercent"`
	MemoryBytes int64     `json:"memoryBytes"`
}

type ResourceHistory struct {
	GeneratedAt    time.Time        `json:"generatedAt"`
	SampleCount    int              `json:"sampleCount"`
	RetentionHours int              `json:"retentionHours"`
	Series         []ResourceSeries `json:"series"`
}

func SampleResources(snapshot Snapshot) ResourceSample {
	sample := ResourceSample{At: snapshot.GeneratedAt}
	for _, application := range snapshot.Apps {
		if application.State != "running" || application.PID == 0 {
			continue
		}
		memory := application.ResidentMemoryBytes
		if memory == 0 {
			memory = application.MemoryBytes
		}
		sample.Rows = append(sample.Rows, ResourceRow{
			ID: application.ID, Kind: "application",
			Name:       application.ProjectName + " / " + application.Name,
			CPUPercent: application.CPUPercent, MemoryBytes: memory,
		})
	}
	for _, agent := range snapshot.Agents {
		sample.Rows = append(sample.Rows, ResourceRow{
			ID: agent.ID, Kind: "agent", Name: agent.Name,
			CPUPercent: agent.CPUPercent, MemoryBytes: agent.MemoryBytes,
		})
	}
	return sample
}

func AppendResourceSample(history []ResourceSample, sample ResourceSample, retention time.Duration) []ResourceSample {
	if retention <= 0 {
		retention = ResourceHistoryRetention
	}
	if len(history) > 0 && !history[len(history)-1].At.Before(sample.At) && equalResourceRows(history[len(history)-1].Rows, sample.Rows) {
		history[len(history)-1] = sample
		return trimResourceHistory(history, sample.At, retention)
	}
	return trimResourceHistory(append(history, sample), sample.At, retention)
}

func trimResourceHistory(history []ResourceSample, now time.Time, retention time.Duration) []ResourceSample {
	cutoff := now.Add(-retention)
	keep := 0
	for keep < len(history) && history[keep].At.Before(cutoff) {
		keep++
	}
	if keep == 0 {
		return history
	}
	return append([]ResourceSample(nil), history[keep:]...)
}

func resourceSeriesKey(kind, name string) string {
	return kind + ":" + name
}

func groupResourceRows(rows []ResourceRow) []ResourceRow {
	index := map[string]int{}
	grouped := make([]ResourceRow, 0, len(rows))
	for _, row := range rows {
		key := resourceSeriesKey(row.Kind, row.Name)
		position, exists := index[key]
		if !exists {
			index[key] = len(grouped)
			grouped = append(grouped, ResourceRow{
				ID: key, Kind: row.Kind, Name: row.Name,
				CPUPercent: row.CPUPercent, MemoryBytes: row.MemoryBytes,
			})
			continue
		}
		grouped[position].CPUPercent += row.CPUPercent
		grouped[position].MemoryBytes += row.MemoryBytes
	}
	return grouped
}

func BuildResourceHistory(history []ResourceSample) ResourceHistory {
	index := map[string]int{}
	result := ResourceHistory{SampleCount: len(history), RetentionHours: int(ResourceHistoryRetention.Hours())}
	if len(history) > 0 {
		result.GeneratedAt = history[len(history)-1].At
	}
	for _, sample := range history {
		for _, row := range groupResourceRows(sample.Rows) {
			position, exists := index[row.ID]
			if !exists {
				position = len(result.Series)
				index[row.ID] = position
				result.Series = append(result.Series, ResourceSeries{ID: row.ID, Kind: row.Kind, Name: row.Name})
			}
			result.Series[position].Name = row.Name
			result.Series[position].Kind = row.Kind
			result.Series[position].Points = append(result.Series[position].Points, ResourcePoint{
				At: sample.At, CPUPercent: row.CPUPercent, MemoryBytes: row.MemoryBytes,
			})
		}
	}
	return result
}

func equalResourceRows(left, right []ResourceRow) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
