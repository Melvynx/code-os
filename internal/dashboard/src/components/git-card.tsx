import { GitBranchIcon } from "lucide-react"

import type { Project, Worktree } from "@/api/schema"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { getGitChangeCount } from "@/lib/format"
import { cn } from "@/lib/utils"

const GIT_CHANGE_METRICS = [
  { key: "modified", label: "Modified", activeClassName: "text-[var(--git-modified)]" },
  { key: "added", label: "Added", activeClassName: "text-[var(--git-added)]" },
  { key: "deleted", label: "Deleted", activeClassName: "text-[var(--git-deleted)]" },
  { key: "untracked", label: "Untracked", activeClassName: "text-[var(--git-untracked)]" },
  { key: "conflicts", label: "Conflicts", activeClassName: "text-[var(--git-conflict)]" },
] as const

export function GitCard({ project, worktree }: Readonly<{ project: Project; worktree: Worktree }>) {
  const hasConflicts = worktree.git.conflicts > 0
  const changeCount = getGitChangeCount(worktree.git)

  return (
    <Card className={cn("min-w-0 transition-colors hover:border-white", hasConflicts && "border-[var(--git-conflict)]")}>
      <CardHeader>
        <div className="min-w-0">
          <GitBranchIcon aria-hidden="true" className="mb-4 size-4 text-[var(--git-info)]" />
          <CardTitle>{project.name}</CardTitle>
          <CardDescription>{worktree.path}</CardDescription>
        </div>
        <Badge variant={hasConflicts ? "conflict" : "modified"}>{changeCount} changes</Badge>
      </CardHeader>
      <CardContent>
        <dl className="grid grid-cols-2 border border-[#333] sm:grid-cols-5">
          {GIT_CHANGE_METRICS.map(({ key, label, activeClassName }, index) => (
            <div key={key} className={index ? "border-l border-[#333] p-3" : "p-3"}>
              <dt className={cn("font-mono text-[10px] uppercase tracking-wider text-[#888]", worktree.git[key] > 0 && activeClassName)}>{label}</dt>
              <dd className={cn("mt-2 font-mono text-lg tabular-nums text-[#888]", worktree.git[key] > 0 && activeClassName)}>{worktree.git[key]}</dd>
            </div>
          ))}
        </dl>
        <div className="mt-4 flex flex-wrap gap-2">
          <Badge variant="untracked">{worktree.git.branch || "Detached"}</Badge>
          <Badge variant="neutral">{worktree.main ? "Main worktree" : "Linked worktree"}</Badge>
          {worktree.locked ? <Badge variant="warning">Locked</Badge> : null}
          {worktree.prunable ? <Badge variant="conflict">Prunable</Badge> : null}
          {worktree.git.ahead ? <Badge variant="info">↑ {worktree.git.ahead} ahead</Badge> : null}
          {worktree.git.behind ? <Badge variant="neutral">↓ {worktree.git.behind} behind</Badge> : null}
        </div>
      </CardContent>
    </Card>
  )
}
