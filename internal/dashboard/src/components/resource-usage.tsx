import { BotIcon, BoxesIcon } from "lucide-react"

import type { AgentProcess, Application } from "@/api/schema"
import { ProcessControlDialog } from "@/components/process-control-dialog"
import { Card } from "@/components/ui/card"
import { formatBytes } from "@/lib/format"

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
    <Card className="overflow-hidden rounded-none">
      <div className="flex flex-col gap-4 border-b border-[#333] px-4 py-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="font-mono text-[10px] uppercase tracking-[0.16em] text-[#888]">Live resources</p>
          <h2 className="mt-1 text-base font-medium text-white">CPU and memory by process</h2>
          <p className="mt-1 text-sm text-[#888]">Agent trees are grouped; app data comes directly from Portly.</p>
        </div>
        <div className="flex gap-5 font-mono text-xs tabular-nums">
          <span><b className="text-[#0070f3]">{totalCPU.toFixed(1)}%</b> <span className="text-[#888]">CPU</span></span>
          <span><b className="text-[#50e3c2]">{formatBytes(totalMemory)}</b> <span className="text-[#888]">RAM</span></span>
        </div>
      </div>
      {rows.length ? (
        <div className="divide-y divide-[#222]">
          {rows.map((row) => {
            const isAgent = row.kind === "agent"
            const pending = isAgent ? pendingAgentId === row.id : pendingApplicationId === row.id
            return (
              <div key={`${row.kind}-${row.id}`} className="grid gap-4 px-4 py-4 lg:grid-cols-[minmax(220px,1.2fr)_minmax(150px,1fr)_minmax(180px,1.2fr)_auto] lg:items-center">
                <div className="flex min-w-0 items-start gap-3">
                  <span className={`mt-0.5 flex size-8 shrink-0 items-center justify-center border ${isAgent ? "border-[#0070f3]/60 text-[#0070f3]" : "border-[#50e3c2]/60 text-[#50e3c2]"}`}>
                    {isAgent ? <BotIcon className="size-4" aria-hidden="true" /> : <BoxesIcon className="size-4" aria-hidden="true" />}
                  </span>
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-white">{row.name}</p>
                    <p className="mt-1 font-mono text-[11px] text-[#888]">{row.detail}</p>
                  </div>
                </div>
                <ResourceBar label="CPU" value={`${row.cpuPercent.toFixed(1)}%`} width={(row.cpuPercent / maximumCPU) * 100} color={isAgent ? "#0070f3" : "#50e3c2"} />
                <ResourceBar label="Memory" value={formatBytes(row.memoryBytes)} width={(row.memoryBytes / maximumMemory) * 100} color={isAgent ? "#0070f3" : "#50e3c2"} />
                <ProcessControlDialog
                  action={row.kind}
                  name={row.name}
                  pending={pending}
                  onConfirm={() => isAgent ? onTerminateAgent(row.id) : onStopApplication(row.id)}
                />
              </div>
            )
          })}
        </div>
      ) : (
        <div className="px-4 py-10 text-center font-mono text-xs text-[#888]">No running apps or development agents detected.</div>
      )}
    </Card>
  )
}

function ResourceBar({ label, value, width, color }: Readonly<{ label: string; value: string; width: number; color: string }>) {
  return (
    <div>
      <div className="mb-2 flex items-center justify-between gap-3 font-mono text-[11px]">
        <span className="uppercase text-[#888]">{label}</span>
        <span className="tabular-nums text-white">{value}</span>
      </div>
      <div className="h-2 overflow-hidden bg-[#1a1a1a]" role="img" aria-label={`${label}: ${value}`}>
        <div className="h-full min-w-px transition-[width] duration-300" style={{ width: `${Math.min(100, Math.max(0, width))}%`, backgroundColor: color }} />
      </div>
    </div>
  )
}
