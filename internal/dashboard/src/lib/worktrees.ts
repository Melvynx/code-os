import type { Project, Worktree } from "@/api/schema"
import { getGitChangeCount } from "@/lib/format"

export interface ProjectWorktree {
  project: Project
  worktree: Worktree
}

export function getProjectWorktrees(project: Project): Worktree[] {
  if (project.worktrees.length) return project.worktrees

  return [{
    id: `${project.id}-main`,
    path: project.path,
    main: true,
    locked: false,
    prunable: false,
    git: project.git,
  }]
}

export function getAllWorktrees(projects: Project[]): ProjectWorktree[] {
  return projects.flatMap((project) =>
    getProjectWorktrees(project).map((worktree) => ({ project, worktree })),
  )
}

export function getDirtyWorktrees(projects: Project[]) {
  return getAllWorktrees(projects).filter(({ worktree }) => getGitChangeCount(worktree.git) > 0)
}

export function getProjectChangeCount(project: Project) {
  return getProjectWorktrees(project).reduce((total, worktree) => total + getGitChangeCount(worktree.git), 0)
}
