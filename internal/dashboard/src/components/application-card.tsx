import { ExternalLinkIcon } from "lucide-react"

import type { Application } from "@/api/schema"
import { AppIcon } from "@/components/app-icon"
import { ProcessControlDialog } from "@/components/process-control-dialog"
import { ApplicationStatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  applicationDisplayName,
  applicationDomain,
  applicationHostLabel,
  applicationHref,
  applicationMemory,
  applicationRole,
  applicationTitle,
} from "@/lib/applications"
import { formatBytes } from "@/lib/format"
import { applicationAccent, identityColorBorder, identityColorSurface } from "@/lib/identity-color"

type ApplicationCardProps = Readonly<{
  application: Application
  peers?: Application[]
  pending?: boolean
  onStop?: (id: string) => Promise<unknown>
}>

export function ApplicationCard({ application, peers, pending = false, onStop }: ApplicationCardProps) {
  const href = applicationHref(application)
  const domain = applicationDomain(application)
  const host = applicationHostLabel(application)
  const memory = formatBytes(applicationMemory(application))
  const accent = applicationAccent(application, peers ?? [application])

  return (
    <article
      className="flex min-h-44 flex-col justify-between gap-6 border p-5"
      style={{ background: identityColorSurface(accent), borderColor: identityColorBorder(accent) }}
    >
      <div className="flex items-start gap-4">
        <AppIcon application={application} accent={accent} size="lg" />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <h3 className="text-2xl font-medium tracking-tight text-pretty">{applicationTitle(application)}</h3>
            <ApplicationStatusBadge application={application} />
          </div>
          <p className="mt-1 text-base text-foreground/80">{domain || host || application.projectName}</p>
          <p className="mt-2 font-mono text-xs text-muted-foreground">
            {applicationRole(application)}
            {application.port ? ` · :${application.port}` : ""}
            {host && domain && host !== domain ? ` · ${host}` : ""}
          </p>
        </div>
      </div>
      <div className="flex flex-wrap items-end justify-between gap-4">
        <dl className="grid grid-cols-2 gap-x-6 gap-y-1 font-mono text-xs tabular-nums">
          <div>
            <dt className="text-muted-foreground">CPU</dt>
            <dd className="mt-1 text-sm" style={{ color: accent }}>{application.cpuPercent.toFixed(1)}%</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Memory</dt>
            <dd className="mt-1 text-sm" style={{ color: accent }}>{memory}</dd>
          </div>
        </dl>
        <div className="flex flex-wrap gap-2">
          {href ? (
            <Button asChild>
              <a href={href} target="_blank" rel="noreferrer">
                Open
                <ExternalLinkIcon data-icon="inline-end" />
              </a>
            </Button>
          ) : null}
          {onStop && application.state === "running" ? (
            <ProcessControlDialog action="application" name={applicationDisplayName(application)} pending={pending} onConfirm={() => onStop(application.id)} />
          ) : null}
        </div>
      </div>
    </article>
  )
}
