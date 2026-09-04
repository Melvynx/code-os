import { Link, createFileRoute } from "@tanstack/react-router"

import { useStatusReport } from "@/api/status-queries"
import { CheckList } from "@/components/check-list"
import { FactCard } from "@/components/fact-card"
import { ImageProbe } from "@/components/image-probe"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { SectionHeading } from "@/components/section-heading"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useDashboardSearch } from "@/contexts/search-context"
import { formatRelativeTime } from "@/lib/format"
import { groupStatusChecks, imagePipelineHealthy, statusSearchValues } from "@/lib/status"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/status")({ component: StatusPage })

function StatusPage() {
  const query = useStatusReport()
  const { query: search } = useDashboardSearch()

  if (query.isPending) return <PageLoading label="Loading environment status" />
  if (query.isError) return <PageError message={query.error.message} retry={() => void query.refetch()} />

  const report = query.data
  if (search.trim() && !statusSearchValues(report).join(" ").toLocaleLowerCase().includes(search.trim().toLocaleLowerCase())) {
    return <EmptyState title="No matching status fields" description="Try another search or clear the filter." />
  }

  const imagesHealthy = imagePipelineHealthy(report.images)
  const groups = groupStatusChecks(report.checks)

  return (
    <div className="flex flex-col gap-8">
      <Card>
        <CardHeader className="gap-5 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <CardDescription className="font-mono text-[10px] uppercase tracking-wider">Environment</CardDescription>
            <div className="mt-2 flex flex-wrap items-center gap-3">
              <span aria-hidden="true" className={cn("size-2.5", report.healthy ? "bg-success" : "bg-destructive")} />
              <CardTitle className="text-2xl">{report.healthy ? "Healthy" : "Needs attention"}</CardTitle>
              <Badge variant={report.healthy ? "success" : "error"}>{report.failed} failed · {report.warnings} warnings</Badge>
            </div>
            <CardDescription className="mt-3 max-w-2xl text-sm leading-relaxed">
              {imagesHealthy
                ? "Screenshots and private files decode, the media bypass key is private, and Code OS can serve visual evidence."
                : "The image pipeline is incomplete. Check the files root, bypass key, and whether recent artifacts decode."}
            </CardDescription>
            <p className="mt-2 font-mono text-xs text-muted-foreground">Checked {formatRelativeTime(report.generatedAt)}</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button asChild variant="outline"><Link to="/screenshots">Open gallery</Link></Button>
            <Button asChild variant="outline"><Link to="/skills-sync">Skills sync</Link></Button>
          </div>
        </CardHeader>
      </Card>

      <section>
        <SectionHeading title="Image pipeline" description="Indexed screenshots and private files are inspected in place. The dashboard then tries to render each recent artifact through the authenticated media route." />
        <div className="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <FactCard label="Screenshots root" value={report.images.screenshotsRootExists ? report.images.screenshotsRoot : "Missing"} hint={`${report.images.indexedCount} indexed`} />
          <FactCard label="Private files root" value={report.images.filesRootExists ? report.images.filesRoot : "Missing"} hint={report.images.sharedRoots ? "Shared with screenshots" : `${report.images.filesImageCount} files`} />
          <FactCard label="Media bypass" value={report.images.bypassPrivate ? "Private key present" : report.images.bypassConfigured ? "Key is not 0600" : "Not configured"} />
          <FactCard label="Decode" value={`${report.images.decodableCount} valid`} hint={`${report.images.emptyCount} empty · ${report.images.undecodableCount} broken`} />
        </div>
        {report.images.recent.length ? (
          <div className="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            {report.images.recent.map((sample) => <ImageProbe key={sample.path} sample={sample} />)}
          </div>
        ) : (
          <div className="mt-5"><EmptyState title="No images indexed" description="Write a screenshot or evidence file under the configured roots to verify delivery." /></div>
        )}
      </section>

      {groups.map((group) => (
        <section key={group.id}>
          <SectionHeading title={group.label} />
          <div className="mt-5"><CheckList checks={group.checks} /></div>
        </section>
      ))}
    </div>
  )
}
