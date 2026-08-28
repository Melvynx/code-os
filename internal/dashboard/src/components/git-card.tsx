import { GitBranchIcon } from "lucide-react"

import type { Project } from "@/api/schema"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { getChangeCount } from "@/lib/format"

const CHANGE_LABELS = [
  ["modified", "Modified"],
  ["added", "Added"],
  ["deleted", "Deleted"],
  ["untracked", "Untracked"],
  ["conflicts", "Conflicts"],
] as const

export function GitCard({ project }: Readonly<{ project: Project }>) {
  return (
    <Card className="transition-colors hover:border-white">
      <CardHeader>
        <div className="min-w-0">
          <GitBranchIcon aria-hidden="true" className="mb-4 size-4 text-[#888]" />
          <CardTitle>{project.name}</CardTitle>
          <CardDescription>{project.path}</CardDescription>
        </div>
        <Badge variant={project.git.conflicts ? "error" : "warning"}>{getChangeCount(project)} changes</Badge>
      </CardHeader>
      <CardContent>
        <dl className="grid grid-cols-2 border border-[#333] sm:grid-cols-5">
          {CHANGE_LABELS.map(([key, label], index) => (
            <div key={key} className={index ? "border-l border-[#333] p-3" : "p-3"}>
              <dt className="font-mono text-[10px] uppercase tracking-wider text-[#888]">{label}</dt>
              <dd className="mt-2 font-mono text-lg tabular-nums text-white">{project.git[key]}</dd>
            </div>
          ))}
        </dl>
        <div className="mt-4 flex flex-wrap gap-2">
          <Badge variant="neutral">{project.git.branch || "Detached"}</Badge>
          {project.git.ahead ? <Badge variant="link">↑ {project.git.ahead} ahead</Badge> : null}
          {project.git.behind ? <Badge variant="neutral">↓ {project.git.behind} behind</Badge> : null}
        </div>
      </CardContent>
    </Card>
  )
}
