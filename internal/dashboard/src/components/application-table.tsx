import { ExternalLinkIcon } from "lucide-react"

import type { Application } from "@/api/schema"
import { ApplicationStatusBadge } from "@/components/status-badge"
import { ProcessControlDialog } from "@/components/process-control-dialog"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { formatBytes } from "@/lib/format"

function safeHttpUrl(value: string | undefined) {
  if (!value) return null
  try {
    const url = new URL(value, window.location.origin)
    return url.protocol === "http:" || url.protocol === "https:" ? url.href : null
  } catch {
    return null
  }
}

type ApplicationTableProps = Readonly<{
  applications: Application[]
  compact?: boolean
  pendingApplicationId?: string
  onStop?: (id: string) => Promise<unknown>
}>

export function ApplicationTable({ applications, compact = false, pendingApplicationId, onStop }: ApplicationTableProps) {
  return (
    <Table>
      <caption className="sr-only">Applications supervised by Portly</caption>
      <TableHeader>
        <TableRow>
          <TableHead>Application</TableHead>
          <TableHead>Port</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Memory</TableHead>
          {!compact ? <TableHead><span className="sr-only">Actions</span></TableHead> : null}
        </TableRow>
      </TableHeader>
      <TableBody>
        {applications.map((application) => {
          const publicUrl = safeHttpUrl(application.publicUrl ?? application.url)
          return (
            <TableRow key={application.id}>
              <TableCell>
                <p className="font-medium text-white">{application.projectName} / {application.name}</p>
                <p className="mt-1 max-w-sm truncate font-mono text-xs text-[#888]">{application.command}</p>
              </TableCell>
              <TableCell className="font-mono tabular-nums text-white">{application.port || "—"}</TableCell>
              <TableCell><ApplicationStatusBadge application={application} /></TableCell>
              <TableCell className="font-mono tabular-nums text-[#888]">
                {formatBytes(application.residentMemoryBytes || application.memoryBytes)}
              </TableCell>
              {!compact ? (
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    {publicUrl ? (
                      <Button asChild size="sm" variant="outline">
                        <a href={publicUrl} target="_blank" rel="noreferrer">
                          Open <ExternalLinkIcon aria-hidden="true" />
                        </a>
                      </Button>
                    ) : null}
                    {onStop && application.state === "running" ? (
                      <ProcessControlDialog action="application" name={`${application.projectName} / ${application.name}`} pending={pendingApplicationId === application.id} onConfirm={() => onStop(application.id)} />
                    ) : null}
                    {!publicUrl && (!onStop || application.state !== "running") ? <span className="font-mono text-xs text-[#888]">—</span> : null}
                  </div>
                </TableCell>
              ) : null}
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
