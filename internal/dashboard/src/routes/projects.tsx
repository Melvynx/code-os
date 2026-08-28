import { createFileRoute } from "@tanstack/react-router"

import { useSnapshot } from "@/api/queries"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { ProjectCard } from "@/components/project-card"
import { SectionHeading } from "@/components/section-heading"
import { useDashboardSearch } from "@/contexts/search-context"
import { matchesQuery } from "@/lib/format"

export const Route = createFileRoute("/projects")({ component: ProjectsPage })

function ProjectsPage() {
  const snapshotQuery = useSnapshot()
  const { query } = useDashboardSearch()

  if (snapshotQuery.isPending) return <PageLoading />
  if (snapshotQuery.isError) return <PageError message={snapshotQuery.error.message} retry={() => void snapshotQuery.refetch()} />

  const projects = snapshotQuery.data.projects.filter((project) =>
    matchesQuery(query, [project.name, project.path, project.git.branch, ...project.worktrees.flatMap((worktree) => [worktree.path, worktree.git.branch]), ...project.subprojects.flatMap((subproject) => [subproject.name, subproject.kind])]),
  )

  return (
    <div className="space-y-6">
      <SectionHeading title="Projects, worktrees, and subprojects" description="Repositories discovered across configured roots, including every linked Git worktree." />
      {projects.length ? <section aria-label="Discovered projects" className="grid gap-4 xl:grid-cols-2">{projects.map((project) => <ProjectCard key={project.id} project={project} />)}</section> : <EmptyState title="No matching projects" description="Try another search or refresh the environment." />}
    </div>
  )
}
