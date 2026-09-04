import { createFileRoute } from "@tanstack/react-router"

import { useSnapshot, useStopApplication, useTerminateAgent } from "@/api/queries"
import { ApplicationTable } from "@/components/application-table"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { ResourceChart } from "@/components/resource-chart"
import { ResourceUsage } from "@/components/resource-usage"
import { SectionHeading } from "@/components/section-heading"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
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
    <div className="flex flex-col gap-6">
      <SectionHeading title="Applications" description="Persistent development processes supervised by Portly." />
      <ResourceChart />
      <ResourceUsage
        applications={applications}
        agents={agents}
        pendingApplicationId={stopApplication.isPending ? stopApplication.variables : undefined}
        pendingAgentId={terminateAgent.isPending ? terminateAgent.variables : undefined}
        onStopApplication={(id) => stopApplication.mutateAsync(id)}
        onTerminateAgent={(id) => terminateAgent.mutateAsync(id)}
      />
      {stopApplication.isError || terminateAgent.isError ? (
        <Alert variant="destructive">
          <AlertTitle>Process control failed</AlertTitle>
          <AlertDescription>{(stopApplication.error ?? terminateAgent.error)?.message}</AlertDescription>
        </Alert>
      ) : null}
      {applications.length ? (
        <Card className="gap-0 py-0">
          <ApplicationTable applications={applications} pendingApplicationId={stopApplication.isPending ? stopApplication.variables : undefined} onStop={(id) => stopApplication.mutateAsync(id)} />
        </Card>
      ) : <EmptyState title="No matching applications" description="Start an application through Portly or try another search." />}
    </div>
  )
}
