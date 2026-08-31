import { afterEach, describe, expect, it, vi } from "vitest"

import { stopApplication, terminateAgent } from "./client"

const emptySnapshot = {
  generatedAt: "2026-08-31T12:00:00Z",
  projects: [],
  applications: [],
  agents: [],
  screenshots: [],
  warnings: [],
}

describe("process controls API", () => {
  afterEach(() => vi.unstubAllGlobals())

  it("stops an exact Portly server through the same-origin API", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(emptySnapshot), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(stopApplication("srv_safe/id")).resolves.toEqual(emptySnapshot)
    expect(fetchMock).toHaveBeenCalledWith("/api/applications/srv_safe%2Fid/stop", expect.objectContaining({
      method: "POST",
      credentials: "same-origin",
    }))
  })

  it("terminates an agent using its PID reuse-safe identifier", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 202 }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(terminateAgent("42:9001")).resolves.toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith("/api/agents/42%3A9001/terminate", expect.objectContaining({
      method: "POST",
      credentials: "same-origin",
    }))
  })
})
