import { describe, expect, it } from "vitest"

import type { ResourceHistory } from "@/api/resources"

import { formatHistoryWindow, formatMemoryAxis, groupResourceSeries, resourceColor, toChartRows } from "./resources"

const history: ResourceHistory = {
  generatedAt: "2026-09-01T01:21:00Z",
  sampleCount: 2,
  retentionHours: 6,
  series: [
    {
      id: "cursor",
      kind: "agent",
      name: "Cursor agent",
      points: [
        { at: "2026-09-01T01:20:00Z", cpuPercent: 0.4, memoryBytes: 209_000_000 },
        { at: "2026-09-01T01:21:00Z", cpuPercent: 1.2, memoryBytes: 220_000_000 },
      ],
    },
    {
      id: "srv_web",
      kind: "application",
      name: "demo / web",
      points: [
        { at: "2026-09-01T01:20:00Z", cpuPercent: 8, memoryBytes: 2_100_000_000 },
        { at: "2026-09-01T01:21:00Z", cpuPercent: 6, memoryBytes: 2_000_000_000 },
      ],
    },
  ],
}

describe("resource charts", () => {
  it("builds aligned CPU rows per timestamp", () => {
    expect(toChartRows(history, "cpuPercent")).toEqual([
      { at: "2026-09-01T01:20:00Z", cursor: 0.4, srv_web: 8 },
      { at: "2026-09-01T01:21:00Z", cursor: 1.2, srv_web: 6 },
    ])
  })

  it("keeps agent and app colors distinct", () => {
    expect(resourceColor(history.series[0]!, history.series)).toBe("#0070f3")
    expect(resourceColor(history.series[1]!, history.series)).toBe("#50e3c2")
  })

  it("describes a cached multi-hour window", () => {
    expect(formatHistoryWindow({ ...history, retentionHours: 6 })).toBe("Last 6 hours cached · 2 samples · 1m window")
  })

  it("formats memory ticks compactly", () => {
    expect(formatMemoryAxis(2_147_483_648)).toBe("2.0 GB")
    expect(formatMemoryAxis(209_715_200)).toBe("200 MB")
  })

  it("groups same-name series and sums overlapping samples", () => {
    const grouped = groupResourceSeries([
      {
        id: "42:1",
        kind: "agent",
        name: "Cursor agent",
        points: [
          { at: "2026-09-01T01:20:00Z", cpuPercent: 1.5, memoryBytes: 100 },
          { at: "2026-09-01T01:21:00Z", cpuPercent: 2, memoryBytes: 110 },
        ],
      },
      {
        id: "43:1",
        kind: "agent",
        name: "Cursor agent",
        points: [{ at: "2026-09-01T01:20:00Z", cpuPercent: 2.5, memoryBytes: 200 }],
      },
      {
        id: "srv_web",
        kind: "application",
        name: "demo / web",
        points: [{ at: "2026-09-01T01:20:00Z", cpuPercent: 8, memoryBytes: 50 }],
      },
    ])

    expect(grouped).toHaveLength(2)
    expect(grouped[0]).toMatchObject({
      id: "agent:Cursor agent",
      name: "Cursor agent",
      count: 2,
    })
    expect(grouped[0]!.points).toEqual([
      { at: "2026-09-01T01:20:00Z", cpuPercent: 4, memoryBytes: 300 },
      { at: "2026-09-01T01:21:00Z", cpuPercent: 2, memoryBytes: 110 },
    ])
    expect(grouped[1]).toMatchObject({ id: "application:demo / web", count: 1 })
  })
})
