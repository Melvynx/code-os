import { GitBranchIcon } from "lucide-react"

import type { Project, Worktree } from "@/api/schema"
import { Badge } from "@/components/ui/badge"
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { getGitChangeCount } from "@/lib/format"
import { cn } from "@/lib/utils"

const GIT_CHANGE_METRICS = [
  { key: "modified", label: "Modified", activeClassName: "text-git-modified" },
  { key: "added", label: "Added", activeClassName: "text-git-added" },
  { key: "deleted", label: "Deleted", activeClassName: "text-git-deleted" },
  { key: "untracked", label: "Untracked", activeClassName: "text-git-untracked" },
  { key: "conflicts", label: "Conflicts", activeClassName: "text-git-conflict" },
] as const

export function GitCard({ project, worktree }: Readonly<{ project: Project; worktree: Worktree }>) {
  const hasConflicts = worktree.git.conflicts > 0
  const changeCount = getGitChangeCount(worktree.git)

  return (
    <Card className={cn("min-w-0 transition-colors hover:ring-foreground/20", hasConflicts && "ring-git-conflict")}>
      <CardHeader>
        <GitBranchIcon aria-hidden="true" className="mb-2 text-git-info" />
        <CardTitle>{project.name}</CardTitle>
        <CardDescription>{worktree.path}</CardDescription>
        <CardAction>
          <Badge variant={hasConflicts ? "conflict" : "modified"}>{changeCount} changes</Badge>
        </CardAction>
      </CardHeader>
      <CardContent>
        <dl className="grid grid-cols-2 divide-x border sm:grid-cols-5">
          {GIT_CHANGE_METRICS.map(({ key, label, activeClassName }) => (
            <div key={key} className="p-3">
              <dt className={cn("font-mono text-[10px] uppercase tracking-wider text-muted-foreground", worktree.git[key] > 0 && activeClassName)}>{label}</dt>
              <dd className={cn("mt-2 font-mono text-lg tabular-nums text-muted-foreground", worktree.git[key] > 0 && activeClassName)}>{worktree.git[key]}</dd>
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
