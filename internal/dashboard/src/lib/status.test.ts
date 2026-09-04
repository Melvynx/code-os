import { describe, expect, it } from "vitest"

import type { StatusReport } from "@/api/status"

import { groupStatusChecks, imagePipelineHealthy, imageSampleLabel } from "./status"

const report: StatusReport = {
  generatedAt: "2026-09-01T00:00:00Z",
  healthy: true,
  failed: 0,
  warnings: 0,
  checks: [
    { id: "loopback", label: "Loopback dashboard", detail: "127.0.0.1:7890", status: "pass", group: "environment" },
    { id: "media-bypass", label: "Media bypass key", detail: "/root/.config/code-os/media-bypass-key", status: "pass", group: "images" },
    { id: "skills-sync", label: "Skills synchronization", detail: "synced", status: "pass", group: "skills" },
  ],
  images: {
    screenshotsRoot: "/root/.local/share/code-os/files",
    filesRoot: "/root/.local/share/code-os/files",
    sharedRoots: true,
    screenshotsRootExists: true,
    filesRootExists: true,
    indexedCount: 1,
    filesImageCount: 0,
    decodableCount: 1,
    emptyCount: 0,
    undecodableCount: 0,
    bypassConfigured: true,
    bypassPrivate: true,
    recent: [{
      name: "F01.png",
      path: "/root/.local/share/code-os/files/F01.png",
      url: "/media/abc",
      kind: "screenshot",
      size: 120,
      width: 12,
      height: 8,
      decodable: true,
      createdAt: "2026-09-01T00:00:00Z",
      issue: "",
    }],
  },
  skills: {
    state: "synced",
    synced: true,
    configured: true,
    message: "ok",
    repository: "",
    directory: "",
    branch: "",
    currentBranch: "",
    origin: "",
    originMatches: true,
    git: { branch: "main", ahead: 0, behind: 0, modified: 0, added: 0, deleted: 0, untracked: 0, conflicts: 0 },
    lastCommitHash: "",
    lastCommitMessage: "",
    lastCommitAt: null,
    lastSyncAt: null,
    timerEnabled: false,
    timerActive: false,
    lastTimerAt: null,
    nextTimerAt: null,
    lastTimerResult: "",
    lockHeld: false,
    issues: [],
  },
}

describe("status helpers", () => {
  it("groups checks in the dashboard order", () => {
    expect(groupStatusChecks(report.checks).map((group) => group.id)).toEqual(["images", "skills", "environment"])
  })

  it("treats decodable private images as a healthy pipeline", () => {
    expect(imagePipelineHealthy(report.images)).toBe(true)
    expect(imagePipelineHealthy({ ...report.images, undecodableCount: 1 })).toBe(false)
    const recent = report.images.recent[0]
    expect(recent).toBeDefined()
    expect(imageSampleLabel(recent!)).toBe("12×8")
  })
})
