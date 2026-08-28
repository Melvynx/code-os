import { render, screen } from "@testing-library/react"
import { describe, expect, it } from "vitest"

import type { Project, Worktree } from "@/api/schema"
import { GitCard } from "@/components/git-card"

const worktree: Worktree = {
  id: "code-os-main",
  path: "/root/projects/code-os",
  main: true,
  locked: false,
  prunable: false,
  git: {
    branch: "main",
    ahead: 1,
    behind: 0,
    modified: 8,
    added: 0,
    deleted: 3,
    untracked: 11,
    conflicts: 0,
  },
}

const project: Project = {
  id: "code-os",
  name: "code-os",
  path: "/root/projects/code-os",
  git: worktree.git,
  worktrees: [worktree],
  subprojects: [],
}

describe("GitCard", () => {
  it("uses semantic colors only for active Git change values", () => {
    render(<GitCard project={project} worktree={worktree} />)

    expect(screen.getByText("Modified")).toHaveClass("text-[var(--git-modified)]")
    expect(screen.getByText("Deleted")).toHaveClass("text-[var(--git-deleted)]")
    expect(screen.getByText("Untracked")).toHaveClass("text-[var(--git-untracked)]")
    expect(screen.getByText("Added")).not.toHaveClass("text-[var(--git-added)]")
  })

  it("keeps labels alongside every colored value", () => {
    render(<GitCard project={project} worktree={worktree} />)

    for (const label of ["Modified", "Added", "Deleted", "Untracked", "Conflicts"]) {
      expect(screen.getByText(label)).toBeVisible()
    }
  })
})
