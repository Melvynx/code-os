import { FolderGit2Icon } from "lucide-react"

import type { Project } from "@/api/schema"
import { Badge } from "@/components/ui/badge"
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { WorktreeList } from "@/components/worktree-list"
import { getProjectChangeCount, getProjectWorktrees } from "@/lib/worktrees"

export function ProjectCard({ project }: Readonly<{ project: Project }>) {
  const changeCount = getProjectChangeCount(project)
  const worktreeCount = getProjectWorktrees(project).length

  return (
    <Card className="min-w-0 transition-colors hover:ring-foreground/20">
      <CardHeader>
        <FolderGit2Icon aria-hidden="true" className="mb-2 text-muted-foreground" />
        <CardTitle>{project.name}</CardTitle>
        <CardDescription>{project.path}</CardDescription>
        <CardAction>
          <Badge variant={changeCount ? "warning" : "success"}>{changeCount ? `${changeCount} changes` : "Clean"}</Badge>
        </CardAction>
      </CardHeader>
      <CardContent>
        <dl className="grid grid-cols-2 divide-x border">
          <div className="p-3">
            <dt className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">Worktrees</dt>
            <dd className="mt-2 font-mono text-sm tabular-nums">{worktreeCount}</dd>
          </div>
          <div className="p-3">
            <dt className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">Subprojects</dt>
            <dd className="mt-2 font-mono text-sm tabular-nums">{project.subprojects.length}</dd>
          </div>
        </dl>
        <WorktreeList project={project} />
        <div className="mt-4 flex flex-wrap gap-2">
          {project.subprojects.length ? project.subprojects.slice(0, 8).map((subproject) => (
            <Badge key={`${subproject.path}:${subproject.name}`} variant="neutral">{subproject.name} · {subproject.kind}</Badge>
          )) : <Badge variant="neutral">Repository root</Badge>}
        </div>
      </CardContent>
    </Card>
  )
}
