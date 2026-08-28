import { createFileRoute } from "@tanstack/react-router"

import { useSnapshot } from "@/api/queries"
import { ApplicationTable } from "@/components/application-table"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { SectionHeading } from "@/components/section-heading"
import { Card } from "@/components/ui/card"
import { useDashboardSearch } from "@/contexts/search-context"
import { matchesQuery } from "@/lib/format"

export const Route = createFileRoute("/applications")({ component: ApplicationsPage })

function ApplicationsPage() {
  const snapshotQuery = useSnapshot()
  const { query } = useDashboardSearch()

  if (snapshotQuery.isPending) return <PageLoading />
  if (snapshotQuery.isError) return <PageError message={snapshotQuery.error.message} retry={() => void snapshotQuery.refetch()} />

  const applications = snapshotQuery.data.applications.filter((application) =>
    matchesQuery(query, [application.name, application.projectName, application.command, application.directory, application.port, application.state]),
  )

  return (
    <div className="space-y-6">
      <SectionHeading title="Applications" description="Persistent development processes supervised by Portly." />
      {applications.length ? <Card className="overflow-hidden rounded-none"><ApplicationTable applications={applications} /></Card> : <EmptyState title="No matching applications" description="Start an application through Portly or try another search." />}
    </div>
  )
}
