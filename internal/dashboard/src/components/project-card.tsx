import { FolderGit2Icon } from "lucide-react"

import type { Project } from "@/api/schema"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { WorktreeList } from "@/components/worktree-list"
import { getProjectChangeCount, getProjectWorktrees } from "@/lib/worktrees"

export function ProjectCard({ project }: Readonly<{ project: Project }>) {
  const changeCount = getProjectChangeCount(project)
  const worktreeCount = getProjectWorktrees(project).length

  return (
    <Card className="min-w-0 transition-colors hover:border-white">
      <CardHeader>
        <div className="min-w-0">
          <FolderGit2Icon aria-hidden="true" className="mb-4 size-4 text-[#888]" />
          <CardTitle>{project.name}</CardTitle>
          <CardDescription>{project.path}</CardDescription>
        </div>
        <Badge variant={changeCount ? "warning" : "success"}>{changeCount ? `${changeCount} changes` : "Clean"}</Badge>
      </CardHeader>
      <CardContent>
        <dl className="grid grid-cols-2 gap-px border border-[#333] bg-[#333]">
          <div className="bg-black p-3">
            <dt className="font-mono text-[10px] uppercase tracking-wider text-[#888]">Worktrees</dt>
            <dd className="mt-2 font-mono text-sm tabular-nums text-white">{worktreeCount}</dd>
          </div>
          <div className="bg-black p-3">
            <dt className="font-mono text-[10px] uppercase tracking-wider text-[#888]">Subprojects</dt>
            <dd className="mt-2 font-mono text-sm tabular-nums text-white">{project.subprojects.length}</dd>
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
