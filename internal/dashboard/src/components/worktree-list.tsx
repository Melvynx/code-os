import { GitBranchIcon } from "lucide-react"

import type { Project } from "@/api/schema"
import { Badge } from "@/components/ui/badge"
import { Item, ItemActions, ItemContent, ItemDescription, ItemGroup, ItemMedia, ItemTitle } from "@/components/ui/item"
import { getGitChangeCount } from "@/lib/format"
import { getProjectWorktrees } from "@/lib/worktrees"

export function WorktreeList({ project }: Readonly<{ project: Project }>) {
  const worktrees = getProjectWorktrees(project)

  return (
    <ItemGroup className="mt-4">
      {worktrees.map((worktree) => {
        const changeCount = getGitChangeCount(worktree.git)
        return (
          <Item key={worktree.id} variant="outline">
            <ItemMedia>
              <GitBranchIcon aria-hidden="true" className="text-git-info" />
            </ItemMedia>
            <ItemContent>
              <ItemTitle className="font-mono">
                {worktree.git.branch || "Detached"}
                {worktree.main ? <Badge variant="neutral">Main</Badge> : null}
                {worktree.locked ? <Badge variant="warning">Locked</Badge> : null}
                {worktree.prunable ? <Badge variant="conflict">Prunable</Badge> : null}
              </ItemTitle>
              <ItemDescription className="font-mono">{worktree.path}</ItemDescription>
            </ItemContent>
            <ItemActions>
              <Badge variant={changeCount ? "modified" : "success"}>{changeCount ? `${changeCount} changes` : "Clean"}</Badge>
            </ItemActions>
          </Item>
        )
      })}
    </ItemGroup>
  )
}
