import { createFileRoute } from "@tanstack/react-router"

import { useSnapshot, useStopApplication, useTerminateAgent } from "@/api/queries"
import { ApplicationTable } from "@/components/application-table"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { SectionHeading } from "@/components/section-heading"
import { ResourceUsage } from "@/components/resource-usage"
import { Card } from "@/components/ui/card"
import { useDashboardSearch } from "@/contexts/search-context"
import { matchesQuery } from "@/lib/format"

export const Route = createFileRoute("/applications")({ component: ApplicationsPage })

function ApplicationsPage() {
  const snapshotQuery = useSnapshot()
  const stopApplication = useStopApplication()
  const terminateAgent = useTerminateAgent()
  const { query } = useDashboardSearch()

  if (snapshotQuery.isPending) return <PageLoading />
  if (snapshotQuery.isError) return <PageError message={snapshotQuery.error.message} retry={() => void snapshotQuery.refetch()} />

  const applications = snapshotQuery.data.applications.filter((application) =>
    matchesQuery(query, [application.name, application.projectName, application.command, application.directory, application.port, application.state]),
  )
  const agents = snapshotQuery.data.agents.filter((agent) =>
    matchesQuery(query, [agent.name, agent.command, agent.pid]),
  )

  return (
    <div className="space-y-6">
      <SectionHeading title="Applications" description="Persistent development processes supervised by Portly." />
      <ResourceUsage
        applications={applications}
        agents={agents}
        pendingApplicationId={stopApplication.isPending ? stopApplication.variables : undefined}
        pendingAgentId={terminateAgent.isPending ? terminateAgent.variables : undefined}
        onStopApplication={(id) => stopApplication.mutateAsync(id)}
        onTerminateAgent={(id) => terminateAgent.mutateAsync(id)}
      />
      {stopApplication.isError || terminateAgent.isError ? (
        <div role="alert" className="border border-[#e00] px-4 py-3 font-mono text-xs text-[#e00]">
          {(stopApplication.error ?? terminateAgent.error)?.message}
        </div>
      ) : null}
      {applications.length ? (
        <Card className="overflow-hidden rounded-none">
          <ApplicationTable applications={applications} pendingApplicationId={stopApplication.isPending ? stopApplication.variables : undefined} onStop={(id) => stopApplication.mutateAsync(id)} />
        </Card>
      ) : <EmptyState title="No matching applications" description="Start an application through Portly or try another search." />}
    </div>
  )
}
