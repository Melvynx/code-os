import { ExternalLinkIcon } from "lucide-react"

import type { Application } from "@/api/schema"
import { AppIcon } from "@/components/app-icon"
import { ProcessControlDialog } from "@/components/process-control-dialog"
import { ApplicationStatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import {
  applicationDisplayName,
  applicationDomain,
  applicationHostLabel,
  applicationHref,
  applicationMemory,
  applicationRole,
  applicationTitle,
  sortApplications,
} from "@/lib/applications"
import { formatBytes } from "@/lib/format"
import { applicationAccent } from "@/lib/identity-color"

type ApplicationTableProps = Readonly<{
  applications: Application[]
  compact?: boolean
  pendingApplicationId?: string
  onStop?: (id: string) => Promise<unknown>
}>

export function ApplicationTable({ applications, compact = false, pendingApplicationId, onStop }: ApplicationTableProps) {
  const rows = sortApplications(applications)

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
        {rows.map((application) => {
          const href = applicationHref(application)
          const domain = applicationDomain(application)
          const host = applicationHostLabel(application)
          return (
            <TableRow key={application.id}>
              <TableCell className="px-4 py-3 align-top whitespace-normal">
                <div className="flex min-w-0 items-start gap-3">
                  <AppIcon application={application} accent={applicationAccent(application, rows)} size="sm" />
                  <div className="min-w-0">
                    <p className="font-medium">{applicationTitle(application)}</p>
                    <p className="mt-0.5 text-sm text-foreground/75">
                      {domain || host || application.projectName}
                      <span className="text-muted-foreground"> · {applicationRole(application)}</span>
                    </p>
                    {!compact ? <p className="mt-1 max-w-xl truncate font-mono text-[11px] text-muted-foreground">{application.command}</p> : null}
                  </div>
                </div>
              </TableCell>
              <TableCell className="px-4 py-3 align-top font-mono tabular-nums">{application.port || "—"}</TableCell>
              <TableCell className="px-4 py-3 align-top"><ApplicationStatusBadge application={application} /></TableCell>
              <TableCell className="px-4 py-3 align-top font-mono tabular-nums text-muted-foreground">
                {formatBytes(applicationMemory(application))}
              </TableCell>
              {!compact ? (
                <TableCell className="px-4 py-3 align-top text-right">
                  <div className="flex justify-end gap-2">
                    {href ? (
                      <Button asChild size="sm" variant="outline">
                        <a href={href} target="_blank" rel="noreferrer">
                          Open
                          <ExternalLinkIcon data-icon="inline-end" />
                        </a>
                      </Button>
                    ) : null}
                    {onStop && application.state === "running" ? (
                      <ProcessControlDialog action="application" name={applicationDisplayName(application)} pending={pendingApplicationId === application.id} onConfirm={() => onStop(application.id)} />
                    ) : null}
                    {!href && (!onStop || application.state !== "running") ? <span className="font-mono text-xs text-muted-foreground">—</span> : null}
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
