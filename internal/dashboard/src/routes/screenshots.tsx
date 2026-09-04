import { createFileRoute } from "@tanstack/react-router"

import { useSnapshot } from "@/api/queries"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { ScreenshotCard } from "@/components/screenshot-card"
import { SectionHeading } from "@/components/section-heading"
import { Separator } from "@/components/ui/separator"
import { useDashboardSearch } from "@/contexts/search-context"
import { matchesQuery } from "@/lib/format"
import { groupScreenshotsByFeature } from "@/lib/screenshots"

export const Route = createFileRoute("/screenshots")({ component: ScreenshotsPage })

function ScreenshotsPage() {
  const snapshotQuery = useSnapshot()
  const { query } = useDashboardSearch()

  if (snapshotQuery.isPending) return <PageLoading />
  if (snapshotQuery.isError) return <PageError message={snapshotQuery.error.message} retry={() => void snapshotQuery.refetch()} />

  const screenshots = snapshotQuery.data.screenshots.filter((screenshot) =>
    matchesQuery(query, [screenshot.name, screenshot.project, screenshot.group]),
  )
  const featureGroups = groupScreenshotsByFeature(screenshots)

  return (
    <div className="flex flex-col gap-6">
      <SectionHeading title="Screenshots" description="Private visual artifacts indexed in place and served through the authenticated dashboard." />
      {featureGroups.length ? (
        <div className="flex flex-col gap-10">
          {featureGroups.map((feature, index) => {
            const headingId = `screenshot-feature-${index}`
            return (
              <section key={feature.name} aria-labelledby={headingId} className="flex flex-col gap-4">
                <div className="flex items-end justify-between gap-4">
                  <div className="min-w-0">
                    <p className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">Feature</p>
                    <h2 id={headingId} className="mt-1 truncate text-lg font-medium">{feature.name}</h2>
                  </div>
                  <span className="shrink-0 font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
                    {feature.screenshots.length} {feature.screenshots.length === 1 ? "screenshot" : "screenshots"}
                  </span>
                </div>
                <Separator />
                <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                  {feature.screenshots.map((screenshot) => <ScreenshotCard key={screenshot.id} screenshot={screenshot} showContext={false} />)}
                </div>
              </section>
            )
          })}
        </div>
      ) : <EmptyState title="No matching screenshots" description="Capture a visual artifact or try another search." />}
    </div>
  )
}
