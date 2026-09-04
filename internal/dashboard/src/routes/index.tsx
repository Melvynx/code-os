import { Link, createFileRoute } from "@tanstack/react-router"
import { AlertTriangleIcon, AppWindowIcon, CameraIcon, FolderGit2Icon, GitCompareArrowsIcon } from "lucide-react"

import { useSnapshot } from "@/api/queries"
import { ApplicationTable } from "@/components/application-table"
import { MetricCard } from "@/components/metric-card"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { ScreenshotCard } from "@/components/screenshot-card"
import { SectionHeading } from "@/components/section-heading"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Item, ItemActions, ItemContent, ItemDescription, ItemGroup, ItemTitle } from "@/components/ui/item"
import { useDashboardSearch } from "@/contexts/search-context"
import { getGitChangeCount, matchesQuery } from "@/lib/format"
import { getAllWorktrees, getDirtyWorktrees } from "@/lib/worktrees"

export const Route = createFileRoute("/")({ component: OverviewPage })

function OverviewPage() {
  const snapshotQuery = useSnapshot()
  const { query } = useDashboardSearch()

  if (snapshotQuery.isPending) return <PageLoading />
  if (snapshotQuery.isError) return <PageError message={snapshotQuery.error.message} retry={() => void snapshotQuery.refetch()} />

  const { projects, applications, screenshots, warnings } = snapshotQuery.data
  const worktrees = getAllWorktrees(projects)
  const dirtyWorktrees = getDirtyWorktrees(projects)
  const visibleApps = applications.filter((application) => matchesQuery(query, [application.name, application.projectName, application.command, application.port]))
  const visibleChanges = dirtyWorktrees.filter(({ project, worktree }) => matchesQuery(query, [project.name, worktree.path, worktree.git.branch]))
  const visibleScreenshots = screenshots.filter((screenshot) => matchesQuery(query, [screenshot.name, screenshot.project, screenshot.group]))
  const metrics = [
    { label: "Projects", value: projects.length, hint: `${worktrees.length} worktrees`, icon: FolderGit2Icon },
    { label: "Modified", value: dirtyWorktrees.length, hint: "worktrees", icon: GitCompareArrowsIcon },
    { label: "Running", value: applications.filter((application) => application.state === "running").length, hint: "Portly apps", icon: AppWindowIcon },
    { label: "Screenshots", value: screenshots.length, hint: "indexed", icon: CameraIcon },
  ]

  return (
    <div className="flex flex-col gap-10">
      <section aria-label="Environment summary" className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {metrics.map((metric) => <MetricCard key={metric.label} {...metric} />)}
      </section>

      {warnings.length ? (
        <Alert variant="destructive">
          <AlertTriangleIcon />
          <AlertTitle>Diagnostics</AlertTitle>
          <AlertDescription>
            <ul className="flex flex-col gap-1 font-mono text-xs">
              {warnings.map((warning) => <li key={warning}>{warning}</li>)}
            </ul>
          </AlertDescription>
        </Alert>
      ) : null}

      <section aria-labelledby="running-apps-title">
        <SectionHeading id="running-apps-title" eyebrow="Portly" title="Running applications" action={<Button asChild variant="link"><Link to="/applications">View all</Link></Button>} />
        <Card className={visibleApps.length ? "mt-5 gap-0 py-0" : "mt-5"}>
          {visibleApps.length ? <ApplicationTable applications={visibleApps.slice(0, 5)} compact /> : <EmptyState title="No managed applications" description="No running application matches this search." />}
        </Card>
      </section>

      <div className="grid gap-10 xl:grid-cols-2">
        <section aria-labelledby="git-overview-title">
          <SectionHeading id="git-overview-title" eyebrow="Work in progress" title="Git changes" action={<Button asChild variant="link"><Link to="/git">View all</Link></Button>} />
          {visibleChanges.length ? (
            <ItemGroup className="mt-5">
              {visibleChanges.slice(0, 5).map(({ project, worktree }) => (
                <Item key={worktree.id} variant="outline">
                  <ItemContent>
                    <ItemTitle>{project.name}</ItemTitle>
                    <ItemDescription className="font-mono">{worktree.git.branch || "Detached"} · {worktree.path}</ItemDescription>
                  </ItemContent>
                  <ItemActions>
                    <Badge variant={worktree.git.conflicts ? "conflict" : "modified"}>{getGitChangeCount(worktree.git)} changes</Badge>
                  </ItemActions>
                </Item>
              ))}
            </ItemGroup>
          ) : (
            <div className="mt-5"><EmptyState title="Every worktree is clean" description="No Git changes match this search." /></div>
          )}
        </section>

        <section aria-labelledby="screenshots-overview-title">
          <SectionHeading id="screenshots-overview-title" eyebrow="Visual evidence" title="Recent screenshots" action={<Button asChild variant="link"><Link to="/screenshots">Open gallery</Link></Button>} />
          {visibleScreenshots.length ? (
            <div className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-3">
              {visibleScreenshots.slice(0, 6).map((screenshot) => <ScreenshotCard key={screenshot.id} screenshot={screenshot} compact />)}
            </div>
          ) : (
            <div className="mt-5"><EmptyState title="No screenshots indexed" description="No visual artifact matches this search." /></div>
          )}
        </section>
      </div>
    </div>
  )
}
