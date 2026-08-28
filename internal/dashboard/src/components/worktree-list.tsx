import { GitBranchIcon } from "lucide-react"

import type { Project } from "@/api/schema"
import { Badge } from "@/components/ui/badge"
import { getGitChangeCount } from "@/lib/format"
import { getProjectWorktrees } from "@/lib/worktrees"

export function WorktreeList({ project }: Readonly<{ project: Project }>) {
  const worktrees = getProjectWorktrees(project)

  return (
    <div className="mt-4 border border-[#333]">
      <div className="flex items-center justify-between border-b border-[#333] px-3 py-2">
        <span className="font-mono text-[10px] uppercase tracking-wider text-[#888]">Worktrees</span>
        <span className="font-mono text-[10px] tabular-nums text-[#888]">{worktrees.length}</span>
      </div>
      <ul className="divide-y divide-[#333]">
        {worktrees.map((worktree) => {
          const changeCount = getGitChangeCount(worktree.git)
          return (
            <li key={worktree.id} className="flex items-start justify-between gap-4 p-3">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <GitBranchIcon aria-hidden="true" className="size-3.5 text-[var(--git-info)]" />
                  <span className="font-mono text-xs text-white">{worktree.git.branch || "Detached"}</span>
                  {worktree.main ? <Badge variant="neutral">Main</Badge> : null}
                  {worktree.locked ? <Badge variant="warning">Locked</Badge> : null}
                  {worktree.prunable ? <Badge variant="conflict">Prunable</Badge> : null}
                </div>
                <p className="mt-1 truncate font-mono text-[11px] text-[#888]">{worktree.path}</p>
              </div>
              <Badge variant={changeCount ? "modified" : "success"}>{changeCount ? `${changeCount} changes` : "Clean"}</Badge>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
