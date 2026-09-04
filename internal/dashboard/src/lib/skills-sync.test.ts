import { describe, expect, it } from "vitest"

import type { SkillsStatus } from "@/api/status"

import { canRunSkillsSync, skillsStateLabel, skillsStateTone } from "./skills-sync"

const base: SkillsStatus = {
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
  lastCommitHash: "abc",
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

describe("skills sync helpers", () => {
  it("labels synced and conflict states", () => {
    expect(skillsStateLabel("synced")).toBe("Synced")
    expect(skillsStateTone("synced")).toBe("success")
    expect(skillsStateTone("conflict")).toBe("error")
    expect(skillsStateTone("dirty")).toBe("warning")
  })

  it("blocks sync when the checkout is unusable", () => {
    expect(canRunSkillsSync(base)).toBe(true)
    expect(canRunSkillsSync({ ...base, state: "conflict", synced: false })).toBe(false)
    expect(canRunSkillsSync({ ...base, configured: false, state: "unconfigured", synced: false })).toBe(false)
  })
})
