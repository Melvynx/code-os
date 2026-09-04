import { afterEach, describe, expect, it, vi } from "vitest"

import { fetchSkillsStatus, fetchStatusReport, runSkillsSync } from "./status"

const skillsStatus = {
  state: "synced",
  synced: true,
  configured: true,
  message: "Skills library is synchronized with origin.",
  repository: "https://github.com/Melvynx/agents-config.git",
  directory: "/root/.agents",
  branch: "main",
  currentBranch: "main",
  origin: "https://github.com/Melvynx/agents-config.git",
  originMatches: true,
  git: { branch: "main", ahead: 0, behind: 0, modified: 0, added: 0, deleted: 0, untracked: 0, conflicts: 0 },
  lastCommitHash: "abc123",
  lastCommitMessage: "sync(test)",
  lastCommitAt: "2026-09-01T00:00:00Z",
  lastSyncAt: "2026-09-01T00:00:00Z",
  timerEnabled: true,
  timerActive: true,
  lastTimerAt: "2026-09-01T01:08:22Z",
  nextTimerAt: "2026-09-01T01:10:22Z",
  lastTimerResult: "success",
  lockHeld: false,
  issues: [],
}

describe("status APIs", () => {
  afterEach(() => vi.unstubAllGlobals())

  it("reads skills synchronization over the same-origin API", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(skillsStatus), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    }))
    vi.stubGlobal("fetch", fetchMock)
    await expect(fetchSkillsStatus()).resolves.toEqual(skillsStatus)
    expect(fetchMock).toHaveBeenCalledWith("/api/skills-sync", expect.objectContaining({ credentials: "same-origin" }))
  })

  it("runs skills sync with a same-origin POST", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({
      result: { message: "Code OS skills sync: repository is up to date" },
      status: skillsStatus,
    }), { status: 200, headers: { "Content-Type": "application/json" } }))
    vi.stubGlobal("fetch", fetchMock)
    await expect(runSkillsSync()).resolves.toEqual({
      result: { message: "Code OS skills sync: repository is up to date" },
      status: skillsStatus,
    })
    expect(fetchMock).toHaveBeenCalledWith("/api/skills-sync", expect.objectContaining({ method: "POST", credentials: "same-origin" }))
  })

  it("reads the environment status report", async () => {
    const payload = {
      generatedAt: "2026-09-01T00:00:00Z",
      healthy: true,
      failed: 0,
      warnings: 0,
      checks: [{ id: "loopback", label: "Loopback dashboard", detail: "127.0.0.1:7890", status: "pass", group: "environment" }],
      images: {
        screenshotsRoot: "/files",
        filesRoot: "/files",
        sharedRoots: true,
        screenshotsRootExists: true,
        filesRootExists: true,
        indexedCount: 0,
        filesImageCount: 0,
        decodableCount: 0,
        emptyCount: 0,
        undecodableCount: 0,
        bypassConfigured: true,
        bypassPrivate: true,
        recent: [],
      },
      skills: skillsStatus,
    }
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(payload), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    }))
    vi.stubGlobal("fetch", fetchMock)
    await expect(fetchStatusReport()).resolves.toEqual(payload)
    expect(fetchMock).toHaveBeenCalledWith("/api/status", expect.objectContaining({ credentials: "same-origin" }))
  })
})
