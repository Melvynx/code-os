import { afterEach, describe, expect, it, vi } from "vitest"

import { fetchResourceHistory } from "./resources"

describe("resources API", () => {
  afterEach(() => vi.unstubAllGlobals())

  it("reads the resource history over the same-origin API", async () => {
    const payload = {
      generatedAt: "2026-09-01T01:21:00Z",
      sampleCount: 1,
      retentionHours: 6,
      series: [{
        id: "cursor",
        kind: "agent",
        name: "Cursor agent",
        points: [{ at: "2026-09-01T01:21:00Z", cpuPercent: 0.4, memoryBytes: 209_000_000 }],
      }],
    }
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(payload), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    }))
    vi.stubGlobal("fetch", fetchMock)
    await expect(fetchResourceHistory()).resolves.toEqual(payload)
    expect(fetchMock).toHaveBeenCalledWith("/api/resources", expect.objectContaining({ credentials: "same-origin" }))
  })
})
