import { describe, expect, it } from "vitest"

import type { Project } from "@/api/schema"
import { getAllWorktrees, getDirtyWorktrees, getProjectChangeCount, getProjectWorktrees } from "./worktrees"

const cleanGit = {
  branch: "main",
  ahead: 0,
  behind: 0,
  modified: 0,
  added: 0,
  deleted: 0,
  untracked: 0,
  conflicts: 0,
}

function project(worktrees: Project["worktrees"]): Project {
  return {
    id: "lumail",
    name: "lumail.io",
    path: "/projects/lumail.io",
    git: cleanGit,
    worktrees,
    subprojects: [],
  }
}

describe("project worktrees", () => {
  it("falls back to the repository root for older snapshots", () => {
    expect(getProjectWorktrees(project([]))).toEqual([expect.objectContaining({ path: "/projects/lumail.io", main: true })])
  })

  it("flattens every worktree and keeps dirty worktrees distinct", () => {
    const value = project([
      { id: "main", path: "/projects/lumail.io", main: true, locked: false, prunable: false, git: cleanGit },
      { id: "feature", path: "/worktrees/feature", main: false, locked: false, prunable: false, git: { ...cleanGit, branch: "feature", modified: 2 } },
    ])

    expect(getAllWorktrees([value])).toHaveLength(2)
    expect(getDirtyWorktrees([value]).map(({ worktree }) => worktree.id)).toEqual(["feature"])
    expect(getProjectChangeCount(value)).toBe(2)
  })
})
