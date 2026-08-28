import { createFileRoute } from "@tanstack/react-router"

import { useSnapshot } from "@/api/queries"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { ScreenshotCard } from "@/components/screenshot-card"
import { SectionHeading } from "@/components/section-heading"
import { useDashboardSearch } from "@/contexts/search-context"
import { matchesQuery } from "@/lib/format"

export const Route = createFileRoute("/screenshots")({ component: ScreenshotsPage })

function ScreenshotsPage() {
  const snapshotQuery = useSnapshot()
  const { query } = useDashboardSearch()

  if (snapshotQuery.isPending) return <PageLoading />
  if (snapshotQuery.isError) return <PageError message={snapshotQuery.error.message} retry={() => void snapshotQuery.refetch()} />

  const screenshots = snapshotQuery.data.screenshots.filter((screenshot) =>
    matchesQuery(query, [screenshot.name, screenshot.project, screenshot.group]),
  )

  return (
    <div className="space-y-6">
      <SectionHeading title="Screenshots" description="Private visual artifacts indexed in place and served through the authenticated dashboard." />
      {screenshots.length ? <section aria-label="Screenshot gallery" className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">{screenshots.map((screenshot) => <ScreenshotCard key={screenshot.id} screenshot={screenshot} />)}</section> : <EmptyState title="No matching screenshots" description="Capture a visual artifact or try another search." />}
    </div>
  )
}
