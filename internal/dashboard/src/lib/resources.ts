import type { ResourceHistory, ResourceSeries } from "@/api/resources"

export const RESOURCE_COLORS = {
  application: ["#50e3c2", "#3dd6b0", "#8af0d8", "#1f8f78"],
  agent: ["#0070f3", "#4d9aff", "#58a6ff", "#79b8ff"],
} as const

export function resourceSeriesKey(series: Pick<ResourceSeries, "kind" | "name">) {
  return `${series.kind}:${series.name}`
}

export function groupResourceSeries(series: ResourceSeries[]) {
  const groups = new Map<string, { kind: ResourceSeries["kind"]; name: string; count: number; byAt: Map<string, { cpuPercent: number; memoryBytes: number }> }>()
  const order: string[] = []

  for (const item of series) {
    const id = resourceSeriesKey(item)
    let group = groups.get(id)
    if (!group) {
      group = { kind: item.kind, name: item.name, count: 0, byAt: new Map() }
      groups.set(id, group)
      order.push(id)
    }
    group.count += 1
    for (const point of item.points) {
      const existing = group.byAt.get(point.at) ?? { cpuPercent: 0, memoryBytes: 0 }
      existing.cpuPercent += point.cpuPercent
      existing.memoryBytes += point.memoryBytes
      group.byAt.set(point.at, existing)
    }
  }

  return order.map((id) => {
    const group = groups.get(id)!
    return {
      id,
      kind: group.kind,
      name: group.name,
      count: group.count,
      points: [...group.byAt.entries()]
        .sort(([left], [right]) => left.localeCompare(right))
        .map(([at, point]) => ({ at, cpuPercent: point.cpuPercent, memoryBytes: point.memoryBytes })),
    }
  })
}

export function resourceColor(series: Pick<ResourceSeries, "id" | "kind">, seriesList: Array<Pick<ResourceSeries, "id" | "kind">>) {
  const palette = RESOURCE_COLORS[series.kind]
  const peers = seriesList.filter((item) => item.kind === series.kind)
  return palette[Math.max(0, peers.findIndex((item) => item.id === series.id)) % palette.length] ?? "#888"
}

export function toChartRows(history: ResourceHistory, metric: "cpuPercent" | "memoryBytes") {
  const rows = new Map<string, Record<string, number | string>>()
  for (const series of history.series) {
    for (const point of series.points) {
      const row = rows.get(point.at) ?? { at: point.at }
      row[series.id] = metric === "memoryBytes" ? point.memoryBytes : Number(point.cpuPercent.toFixed(2))
      rows.set(point.at, row)
    }
  }
  return [...rows.values()].sort((left, right) => String(left.at).localeCompare(String(right.at)))
}

export function formatChartTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })
}

export function formatHistoryWindow(history: Pick<ResourceHistory, "sampleCount" | "retentionHours" | "series">) {
  const hours = history.retentionHours || 6
  const samples = history.sampleCount
  const times = history.series.flatMap((item) => item.points.map((point) => Date.parse(point.at))).filter((value) => Number.isFinite(value))
  if (times.length < 2) {
    return `Cached for ${hours}h. ${samples} sample${samples === 1 ? "" : "s"} so far.`
  }
  const span = Math.max(...times) - Math.min(...times)
  const window = span >= 3_600_000
    ? `${(span / 3_600_000).toFixed(1)}h window`
    : span >= 60_000
      ? `${Math.max(1, Math.round(span / 60_000))}m window`
      : `${Math.max(1, Math.round(span / 1_000))}s window`
  return `Last ${hours} hours cached · ${samples} samples · ${window}`
}

export function formatMemoryAxis(value: number) {
  if (value >= 1_073_741_824) return `${(value / 1_073_741_824).toFixed(1)} GB`
  if (value >= 1_048_576) return `${Math.round(value / 1_048_576)} MB`
  if (value <= 0) return "0"
  return `${Math.round(value / 1_024)} KB`
}
