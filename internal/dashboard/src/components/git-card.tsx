import { GitBranchIcon } from "lucide-react"

import type { Project } from "@/api/schema"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { getChangeCount } from "@/lib/format"
import { cn } from "@/lib/utils"

const GIT_CHANGE_METRICS = [
  { key: "modified", label: "Modified", activeClassName: "text-[var(--git-modified)]" },
  { key: "added", label: "Added", activeClassName: "text-[var(--git-added)]" },
  { key: "deleted", label: "Deleted", activeClassName: "text-[var(--git-deleted)]" },
  { key: "untracked", label: "Untracked", activeClassName: "text-[var(--git-untracked)]" },
  { key: "conflicts", label: "Conflicts", activeClassName: "text-[var(--git-conflict)]" },
] as const

export function GitCard({ project }: Readonly<{ project: Project }>) {
  const hasConflicts = project.git.conflicts > 0

  return (
    <Card className={cn("transition-colors hover:border-white", hasConflicts && "border-[var(--git-conflict)]")}>
      <CardHeader>
        <div className="min-w-0">
          <GitBranchIcon aria-hidden="true" className="mb-4 size-4 text-[var(--git-info)]" />
          <CardTitle>{project.name}</CardTitle>
          <CardDescription>{project.path}</CardDescription>
        </div>
        <Badge variant={hasConflicts ? "conflict" : "modified"}>{getChangeCount(project)} changes</Badge>
      </CardHeader>
      <CardContent>
        <dl className="grid grid-cols-2 border border-[#333] sm:grid-cols-5">
          {GIT_CHANGE_METRICS.map(({ key, label, activeClassName }, index) => (
            <div key={key} className={index ? "border-l border-[#333] p-3" : "p-3"}>
              <dt className={cn("font-mono text-[10px] uppercase tracking-wider text-[#888]", project.git[key] > 0 && activeClassName)}>{label}</dt>
              <dd className={cn("mt-2 font-mono text-lg tabular-nums text-[#888]", project.git[key] > 0 && activeClassName)}>{project.git[key]}</dd>
            </div>
          ))}
        </dl>
        <div className="mt-4 flex flex-wrap gap-2">
          <Badge variant="untracked">{project.git.branch || "Detached"}</Badge>
          {project.git.ahead ? <Badge variant="info">↑ {project.git.ahead} ahead</Badge> : null}
          {project.git.behind ? <Badge variant="neutral">↓ {project.git.behind} behind</Badge> : null}
        </div>
      </CardContent>
    </Card>
  )
}
