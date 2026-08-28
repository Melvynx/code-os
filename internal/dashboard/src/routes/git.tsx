import { createFileRoute } from "@tanstack/react-router"

import { useSnapshot } from "@/api/queries"
import { GitCard } from "@/components/git-card"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { SectionHeading } from "@/components/section-heading"
import { useDashboardSearch } from "@/contexts/search-context"
import { getChangeCount, matchesQuery } from "@/lib/format"

export const Route = createFileRoute("/git")({ component: GitPage })

function GitPage() {
  const snapshotQuery = useSnapshot()
  const { query } = useDashboardSearch()

  if (snapshotQuery.isPending) return <PageLoading />
  if (snapshotQuery.isError) return <PageError message={snapshotQuery.error.message} retry={() => void snapshotQuery.refetch()} />

  const projects = snapshotQuery.data.projects.filter((project) =>
    getChangeCount(project) > 0 && matchesQuery(query, [project.name, project.path, project.git.branch, project.git.upstream]),
  )

  return (
    <div className="space-y-6">
      <SectionHeading title="Git changes" description="A read-only view of work in progress across every repository." />
      {projects.length ? <section aria-label="Modified repositories" className="grid gap-4">{projects.map((project) => <GitCard key={project.id} project={project} />)}</section> : <EmptyState title="No Git changes" description="Every matching repository is clean." />}
    </div>
  )
}
