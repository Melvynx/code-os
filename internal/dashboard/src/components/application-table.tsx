import { ExternalLinkIcon } from "lucide-react"

import type { Application } from "@/api/schema"
import { ProcessControlDialog } from "@/components/process-control-dialog"
import { ApplicationStatusBadge } from "@/components/status-badge"
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
          <TableHead className="px-4">Application</TableHead>
          <TableHead className="px-4">Port</TableHead>
          <TableHead className="px-4">Status</TableHead>
          <TableHead className="px-4">Memory</TableHead>
          {!compact ? <TableHead className="px-4"><span className="sr-only">Actions</span></TableHead> : null}
        </TableRow>
      </TableHeader>
      <TableBody>
        {applications.map((application) => {
          const publicUrl = safeHttpUrl(application.publicUrl ?? application.url)
          return (
            <TableRow key={application.id}>
              <TableCell className="px-4 py-3 align-top whitespace-normal">
                <div className="flex min-w-0 flex-col gap-1">
                  <p className="font-medium">{application.projectName} / {application.name}</p>
                  <p className="max-w-xl truncate font-mono text-xs text-muted-foreground">{application.command}</p>
                </div>
              </TableCell>
              <TableCell className="px-4 py-3 align-top font-mono tabular-nums">{application.port || "—"}</TableCell>
              <TableCell className="px-4 py-3 align-top"><ApplicationStatusBadge application={application} /></TableCell>
              <TableCell className="px-4 py-3 align-top font-mono tabular-nums text-muted-foreground">
                {formatBytes(application.residentMemoryBytes || application.memoryBytes)}
              </TableCell>
              {!compact ? (
                <TableCell className="px-4 py-3 align-top text-right">
                  <div className="flex justify-end gap-2">
                    {publicUrl ? (
                      <Button asChild size="sm" variant="outline">
                        <a href={publicUrl} target="_blank" rel="noreferrer">
                          Open
                          <ExternalLinkIcon data-icon="inline-end" />
                        </a>
                      </Button>
                    ) : null}
                    {onStop && application.state === "running" ? (
                      <ProcessControlDialog action="application" name={`${application.projectName} / ${application.name}`} pending={pendingApplicationId === application.id} onConfirm={() => onStop(application.id)} />
                    ) : null}
                    {!publicUrl && (!onStop || application.state !== "running") ? <span className="font-mono text-xs text-muted-foreground">—</span> : null}
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
