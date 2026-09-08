import { createFileRoute } from "@tanstack/react-router"

import { useSnapshot, useStopApplication, useTerminateAgent } from "@/api/queries"
import { ApplicationCard } from "@/components/application-card"
import { ApplicationTable } from "@/components/application-table"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { ResourceChart } from "@/components/resource-chart"
import { ResourceUsage } from "@/components/resource-usage"
import { SectionHeading } from "@/components/section-heading"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Card } from "@/components/ui/card"
import { useDashboardSearch } from "@/contexts/search-context"
import { applicationMatchesQuery, featuredApplications } from "@/lib/applications"
import { matchesQuery } from "@/lib/format"

export const Route = createFileRoute("/applications")({ component: ApplicationsPage })

function ApplicationsPage() {
  const snapshotQuery = useSnapshot()
  const stopApplication = useStopApplication()
  const terminateAgent = useTerminateAgent()
  const { query } = useDashboardSearch()

  if (snapshotQuery.isPending) return <PageLoading />
  if (snapshotQuery.isError) return <PageError message={snapshotQuery.error.message} retry={() => void snapshotQuery.refetch()} />

  const applications = snapshotQuery.data.applications.filter((application) => applicationMatchesQuery(application, query))
  const agents = snapshotQuery.data.agents.filter((agent) =>
    matchesQuery(query, [agent.name, agent.command, agent.pid]),
  )
  const featured = featuredApplications(applications)

  return (
    <div className="flex flex-col gap-8">
      <SectionHeading title="Applications" description="Persistent development processes supervised by Portly." />
      {featured.length ? (
        <section aria-labelledby="running-apps-title" className="flex flex-col gap-4">
          <SectionHeading id="running-apps-title" title="Running apps" description="Web servers and linked ports, with the live favicon and public hostname." />
          <div className="grid gap-3 xl:grid-cols-2">
            {featured.map((application) => (
              <ApplicationCard
                key={application.id}
                application={application}
                peers={featured}
                pending={stopApplication.isPending && stopApplication.variables === application.id}
                onStop={(id) => stopApplication.mutateAsync(id)}
              />
            ))}
          </div>
        </section>
      ) : null}
      <ResourceChart query={query} />
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
        <section aria-labelledby="all-apps-title" className="flex flex-col gap-4">
          <SectionHeading id="all-apps-title" title="All Portly applications" description="Complete inventory, including supporting processes and stopped servers." />
          <Card className="gap-0 py-0">
            <ApplicationTable applications={applications} pendingApplicationId={stopApplication.isPending ? stopApplication.variables : undefined} onStop={(id) => stopApplication.mutateAsync(id)} />
          </Card>
        </section>
      ) : <EmptyState title="No matching applications" description="Start an application through Portly or try another search." />}
    </div>
  )
}
