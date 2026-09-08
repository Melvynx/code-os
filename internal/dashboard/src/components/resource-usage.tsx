import type { CSSProperties } from "react"
import { BotIcon } from "lucide-react"

import type { AgentProcess, Application } from "@/api/schema"
import { AppIcon } from "@/components/app-icon"
import { EmptyState } from "@/components/page-state"
import { ProcessControlDialog } from "@/components/process-control-dialog"
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Separator } from "@/components/ui/separator"
import { applicationDisplayName, applicationDomain, applicationMemory, applicationRole, applicationTitle, isFeaturedApplication } from "@/lib/applications"
import { formatBytes } from "@/lib/format"
import { applicationColorKey, identityColorBorder, identityColorSurface, uniqueIdentityColor } from "@/lib/identity-color"

type ResourceUsageProps = Readonly<{
  applications: Application[]
  agents: AgentProcess[]
  pendingApplicationId?: string
  pendingAgentId?: string
  onStopApplication: (id: string) => Promise<unknown>
  onTerminateAgent: (id: string) => Promise<unknown>
}>

type ResourceRow =
  | {
      kind: "application"
      id: string
      name: string
      detail: string
      cpuPercent: number
      memoryBytes: number
      application: Application
    }
  | {
      kind: "agent"
      id: string
      name: string
      detail: string
      cpuPercent: number
      memoryBytes: number
    }

export function ResourceUsage({
  applications,
  agents,
  pendingApplicationId,
  pendingAgentId,
  onStopApplication,
  onTerminateAgent,
}: ResourceUsageProps) {
  const rows = resourceRows(applications, agents)
  const colorKeys = rows.map((row) => row.kind === "application" ? applicationColorKey(row.application) : row.name)
  const maximumMemory = Math.max(...rows.map((row) => row.memoryBytes), 1)
  const maximumCPU = Math.max(100, ...rows.map((row) => row.cpuPercent))
  const totalMemory = rows.reduce((total, row) => total + row.memoryBytes, 0)
  const totalCPU = rows.reduce((total, row) => total + row.cpuPercent, 0)

  return (
    <Card>
      <CardHeader className="border-b">
        <CardDescription className="font-mono text-[10px] uppercase tracking-wider">Live resources</CardDescription>
        <CardTitle>CPU and memory by process</CardTitle>
        <CardDescription>Running apps first, then grouped agent trees. App data comes from Portly.</CardDescription>
        <CardAction className="flex gap-5 font-mono text-xs tabular-nums">
          <span><b className="text-info">{totalCPU.toFixed(1)}%</b> <span className="text-muted-foreground">CPU</span></span>
          <span><b className="text-success">{formatBytes(totalMemory)}</b> <span className="text-muted-foreground">RAM</span></span>
        </CardAction>
      </CardHeader>
      {rows.length ? (
        <CardContent className="flex flex-col gap-0 px-0">
          {rows.map((row, index) => {
            const isAgent = row.kind === "agent"
            const pending = isAgent ? pendingAgentId === row.id : pendingApplicationId === row.id
            const accent = uniqueIdentityColor(row.kind === "application" ? applicationColorKey(row.application) : row.name, colorKeys)
            return (
              <div key={`${row.kind}-${row.id}`}>
                {index ? <Separator /> : null}
                <div className="grid gap-4 px-4 py-4 lg:grid-cols-[minmax(240px,1.3fr)_minmax(150px,1fr)_minmax(180px,1.2fr)_auto] lg:items-center">
                  <div className="flex min-w-0 items-start gap-3">
                    {row.kind === "application" ? (
                      <AppIcon application={row.application} accent={accent} size="sm" />
                    ) : (
                      <span
                        className="mt-0.5 flex size-8 shrink-0 items-center justify-center border"
                        style={{
                          color: accent,
                          borderColor: identityColorBorder(accent),
                          background: identityColorSurface(accent),
                        }}
                      >
                        <BotIcon aria-hidden="true" />
                      </span>
                    )}
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium">{row.name}</p>
                      <p className="mt-1 font-mono text-[11px] text-muted-foreground">{row.detail}</p>
                    </div>
                  </div>
                  <ResourceBar
                    label="CPU"
                    value={`${row.cpuPercent.toFixed(1)}%`}
                    width={(row.cpuPercent / maximumCPU) * 100}
                    color={accent}
                  />
                  <ResourceBar
                    label="Memory"
                    value={formatBytes(row.memoryBytes)}
                    width={(row.memoryBytes / maximumMemory) * 100}
                    color={accent}
                  />
                  <ProcessControlDialog
                    action={row.kind}
                    name={row.kind === "application" ? applicationDisplayName(row.application) : row.name}
                    pending={pending}
                    onConfirm={() => isAgent ? onTerminateAgent(row.id) : onStopApplication(row.id)}
                  />
                </div>
              </div>
            )
          })}
        </CardContent>
      ) : (
        <CardContent>
          <EmptyState title="No running processes" description="No running apps or development agents detected." />
        </CardContent>
      )}
    </Card>
  )
}

function resourceRows(applications: Application[], agents: AgentProcess[]): ResourceRow[] {
  const running = applications.filter((application) => application.state === "running" && application.pid)
  const appRows: ResourceRow[] = running.map((application) => {
    const domain = applicationDomain(application)
    return {
      kind: "application",
      id: application.id,
      name: applicationTitle(application),
      detail: [applicationRole(application), domain, `PID ${application.pid}`].filter(Boolean).join(" · "),
      cpuPercent: application.cpuPercent,
      memoryBytes: applicationMemory(application),
      application,
    }
  })
  const agentRows: ResourceRow[] = agents.map((agent) => ({
    kind: "agent",
    id: agent.id,
    name: agent.name,
    detail: `PID ${agent.pid} · ${agent.processCount} process${agent.processCount === 1 ? "" : "es"}`,
    cpuPercent: agent.cpuPercent,
    memoryBytes: agent.memoryBytes,
  }))
  const byMemory = (left: ResourceRow, right: ResourceRow) => right.memoryBytes - left.memoryBytes
  return [
    ...appRows.filter((row) => row.kind === "application" && isFeaturedApplication(row.application)).sort(byMemory),
    ...appRows.filter((row) => row.kind === "application" && !isFeaturedApplication(row.application)).sort(byMemory),
    ...agentRows.sort(byMemory),
  ]
}

function ResourceBar({ label, value, width, color }: Readonly<{ label: string; value: string; width: number; color: string }>) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between gap-3 font-mono text-[11px]">
        <span className="uppercase text-muted-foreground">{label}</span>
        <span className="tabular-nums" style={{ color }}>{value}</span>
      </div>
      <Progress
        value={Math.min(100, Math.max(0, width))}
        aria-label={`${label}: ${value}`}
        className="[&_[data-slot=progress-indicator]]:bg-[var(--bar)]"
        style={{ "--bar": color } as CSSProperties}
      />
    </div>
  )
}
