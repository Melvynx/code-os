import { describe, expect, it } from "vitest"

import { formatBytes, formatRelativeTime, getChangeCount, matchesQuery } from "./format"

describe("dashboard formatters", () => {
  it("formats bytes with an appropriate unit", () => {
    expect(formatBytes(0)).toBe("—")
    expect(formatBytes(1_024)).toBe("1 KB")
    expect(formatBytes(1_610_612_736)).toBe("1.5 GB")
  })

  it("formats elapsed time", () => {
    const currentTime = Date.parse("2026-08-28T12:00:00Z")
    expect(formatRelativeTime("2026-08-28T11:58:00Z", currentTime)).toBe("2m ago")
  })

  it("counts every visible Git change", () => {
    expect(
      getChangeCount({
        id: "project",
        name: "Code OS",
        path: "/root/projects/code-os",
        worktrees: [],
        subprojects: [],
        git: {
          branch: "main",
          ahead: 0,
          behind: 0,
          modified: 2,
          added: 1,
          deleted: 1,
          untracked: 3,
          conflicts: 0,
        },
      }),
    ).toBe(7)
  })

  it("matches values without regard to case", () => {
    expect(matchesQuery("code", ["Code OS", 7890])).toBe(true)
    expect(matchesQuery("portly", ["Code OS", 7890])).toBe(false)
  })
})
