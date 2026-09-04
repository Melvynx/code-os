import { BotIcon, BoxesIcon } from "lucide-react"

import type { AgentProcess, Application } from "@/api/schema"
import { EmptyState } from "@/components/page-state"
import { ProcessControlDialog } from "@/components/process-control-dialog"
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Separator } from "@/components/ui/separator"
import { formatBytes } from "@/lib/format"
import { cn } from "@/lib/utils"

type ResourceUsageProps = Readonly<{
  applications: Application[]
  agents: AgentProcess[]
  pendingApplicationId?: string
  pendingAgentId?: string
  onStopApplication: (id: string) => Promise<unknown>
  onTerminateAgent: (id: string) => Promise<unknown>
}>

type ResourceRow = {
  id: string
  kind: "application" | "agent"
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
  const rows: ResourceRow[] = [
    ...applications
      .filter((application) => application.state === "running" && application.pid)
      .map((application) => ({
        id: application.id,
        kind: "application" as const,
        name: `${application.projectName} / ${application.name}`,
        detail: `PID ${application.pid} · Portly app`,
        cpuPercent: application.cpuPercent,
        memoryBytes: application.residentMemoryBytes || application.memoryBytes,
      })),
    ...agents.map((agent) => ({
      id: agent.id,
      kind: "agent" as const,
      name: agent.name,
      detail: `PID ${agent.pid} · ${agent.processCount} process${agent.processCount === 1 ? "" : "es"}`,
      cpuPercent: agent.cpuPercent,
      memoryBytes: agent.memoryBytes,
    })),
  ].sort((left, right) => right.memoryBytes - left.memoryBytes)

  const maximumMemory = Math.max(...rows.map((row) => row.memoryBytes), 1)
  const maximumCPU = Math.max(100, ...rows.map((row) => row.cpuPercent))
  const totalMemory = rows.reduce((total, row) => total + row.memoryBytes, 0)
  const totalCPU = rows.reduce((total, row) => total + row.cpuPercent, 0)

  return (
    <Card>
      <CardHeader className="border-b">
        <CardDescription className="font-mono text-[10px] uppercase tracking-wider">Live resources</CardDescription>
        <CardTitle>CPU and memory by process</CardTitle>
        <CardDescription>Agent trees are grouped; app data comes directly from Portly.</CardDescription>
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
            return (
              <div key={`${row.kind}-${row.id}`}>
                {index ? <Separator /> : null}
                <div className="grid gap-4 px-4 py-4 lg:grid-cols-[minmax(220px,1.2fr)_minmax(150px,1fr)_minmax(180px,1.2fr)_auto] lg:items-center">
                  <div className="flex min-w-0 items-start gap-3">
                    <span className={cn("mt-0.5 flex size-8 shrink-0 items-center justify-center border", isAgent ? "border-info/60 text-info" : "border-success/60 text-success")}>
                      {isAgent ? <BotIcon aria-hidden="true" /> : <BoxesIcon aria-hidden="true" />}
                    </span>
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium">{row.name}</p>
                      <p className="mt-1 font-mono text-[11px] text-muted-foreground">{row.detail}</p>
                    </div>
                  </div>
                  <ResourceBar label="CPU" value={`${row.cpuPercent.toFixed(1)}%`} width={(row.cpuPercent / maximumCPU) * 100} agent={isAgent} />
                  <ResourceBar label="Memory" value={formatBytes(row.memoryBytes)} width={(row.memoryBytes / maximumMemory) * 100} agent={isAgent} />
                  <ProcessControlDialog
                    action={row.kind}
                    name={row.name}
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

function ResourceBar({ label, value, width, agent }: Readonly<{ label: string; value: string; width: number; agent: boolean }>) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between gap-3 font-mono text-[11px]">
        <span className="uppercase text-muted-foreground">{label}</span>
        <span className="tabular-nums">{value}</span>
      </div>
      <Progress
        value={Math.min(100, Math.max(0, width))}
        aria-label={`${label}: ${value}`}
        className={cn(agent ? "[&_[data-slot=progress-indicator]]:bg-info" : "[&_[data-slot=progress-indicator]]:bg-success")}
      />
    </div>
  )
}
