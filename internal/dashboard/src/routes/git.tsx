import { createFileRoute } from "@tanstack/react-router"

import { useSnapshot } from "@/api/queries"
import { GitCard } from "@/components/git-card"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { SectionHeading } from "@/components/section-heading"
import { useDashboardSearch } from "@/contexts/search-context"
import { matchesQuery } from "@/lib/format"
import { getDirtyWorktrees } from "@/lib/worktrees"

export const Route = createFileRoute("/git")({ component: GitPage })

function GitPage() {
  const snapshotQuery = useSnapshot()
  const { query } = useDashboardSearch()

  if (snapshotQuery.isPending) return <PageLoading />
  if (snapshotQuery.isError) return <PageError message={snapshotQuery.error.message} retry={() => void snapshotQuery.refetch()} />

  const worktrees = getDirtyWorktrees(snapshotQuery.data.projects).filter(({ project, worktree }) =>
    matchesQuery(query, [project.name, worktree.path, worktree.git.branch, worktree.git.upstream]),
  )

  return (
    <div className="flex flex-col gap-6">
      <SectionHeading title="Git changes" description="A read-only view of work in progress across every Git worktree." />
      {worktrees.length ? <section aria-label="Modified worktrees" className="grid gap-4">{worktrees.map(({ project, worktree }) => <GitCard key={worktree.id} project={project} worktree={worktree} />)}</section> : <EmptyState title="No Git changes" description="Every matching worktree is clean." />}
    </div>
  )
}
