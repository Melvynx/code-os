import { Link, createFileRoute } from "@tanstack/react-router"
import { toast } from "sonner"

import { useRunSkillsSync, useSkillsStatus } from "@/api/status-queries"
import { FactCard } from "@/components/fact-card"
import { EmptyState, PageError, PageLoading } from "@/components/page-state"
import { SectionHeading } from "@/components/section-heading"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Item, ItemContent, ItemGroup, ItemTitle } from "@/components/ui/item"
import { Spinner } from "@/components/ui/spinner"
import { useDashboardSearch } from "@/contexts/search-context"
import { formatRelativeTime, getGitChangeCount } from "@/lib/format"
import { canRunSkillsSync, skillsSearchValues, skillsStateLabel, skillsStateTone } from "@/lib/skills-sync"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/skills-sync")({ component: SkillsSyncPage })

function SkillsSyncPage() {
  const query = useSkillsStatus()
  const sync = useRunSkillsSync()
  const search = useDashboardSearch()
  if (query.isPending) return <PageLoading label="Loading skills synchronization" />
  if (query.isError) return <PageError message={query.error.message} retry={() => void query.refetch()} />
  if (search.query.trim() && !skillsSearchValues(query.data).join(" ").toLocaleLowerCase().includes(search.query.trim().toLocaleLowerCase())) {
    return <EmptyState title="No matching skills sync fields" description="Try another search or clear the filter." />
  }

  const status = query.data
  const tone = skillsStateTone(status.state)
  const changes = getGitChangeCount(status.git)

  return (
    <div className="flex flex-col gap-8">
      <Card>
        <CardHeader className="gap-5 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <CardDescription className="font-mono text-[10px] uppercase tracking-wider">Skills library</CardDescription>
            <div className="mt-2 flex flex-wrap items-center gap-3">
              <span aria-hidden="true" className={cn("size-2.5", tone === "success" ? "bg-success" : tone === "warning" ? "bg-foreground" : "bg-destructive")} />
              <CardTitle className="text-2xl">{skillsStateLabel(status.state)}</CardTitle>
              <Badge variant={tone}>{status.synced ? "Origin match" : "Needs attention"}</Badge>
            </div>
            <CardDescription className="mt-3 max-w-2xl text-sm leading-relaxed">{status.message}</CardDescription>
            <p className="mt-2 font-mono text-xs text-muted-foreground">
              {status.lastSyncAt ? `Last sync ${formatRelativeTime(status.lastSyncAt)}` : "No sync commit yet"}
              {status.nextTimerAt ? ` · next timer ${formatRelativeTime(status.nextTimerAt)}` : ""}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              disabled={!canRunSkillsSync(status) || sync.isPending}
              onClick={() => {
                sync.mutate(undefined, {
                  onSuccess: (data) => toast.success(data.result.message),
                  onError: (error) => toast.error(error.message),
                })
              }}
            >
              {sync.isPending ? <Spinner data-icon="inline-start" /> : null}
              {sync.isPending ? "Synchronizing" : "Sync now"}
            </Button>
            <Button asChild variant="outline"><Link to="/settings">Open settings</Link></Button>
          </div>
        </CardHeader>
        {sync.isError ? (
          <Alert variant="destructive" className="mx-(--card-spacing) mb-(--card-spacing)">
            <AlertTitle>Sync failed</AlertTitle>
            <AlertDescription>{sync.error.message}</AlertDescription>
          </Alert>
        ) : null}
        {sync.isSuccess ? (
          <Alert className="mx-(--card-spacing) mb-(--card-spacing)">
            <AlertTitle>Sync completed</AlertTitle>
            <AlertDescription>{sync.data.result.message}</AlertDescription>
          </Alert>
        ) : null}
      </Card>

      <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <FactCard label="Checkout" value={status.directory || "Not set"} />
        <FactCard label="Repository" value={status.repository || "Not set"} />
        <FactCard label="Branch" value={status.currentBranch || status.branch || "Not set"} hint={status.originMatches ? "origin matches" : "origin mismatch"} />
        <FactCard label="Timer" value={status.timerEnabled ? (status.timerActive ? "Enabled" : "Installed") : status.configured ? "Not enabled" : "Not installed"} hint={status.lastTimerResult || "systemd user timer"} />
      </section>

      <section>
        <SectionHeading title="Git drift" description="Read-only view of the configured ~/.agents checkout. Sync commits local changes, rebases, and pushes." />
        <dl className="mt-5 grid grid-cols-2 divide-x border sm:grid-cols-5">
          {([
            ["Ahead", status.git.ahead, "text-git-info"],
            ["Behind", status.git.behind, "text-git-modified"],
            ["Changed", changes, "text-git-modified"],
            ["Conflicts", status.git.conflicts, "text-git-conflict"],
            ["Untracked", status.git.untracked, "text-git-untracked"],
          ] as const).map(([label, value, active]) => (
            <div key={label} className="p-3">
              <dt className={cn("font-mono text-[10px] uppercase tracking-wider text-muted-foreground", value > 0 && active)}>{label}</dt>
              <dd className={cn("mt-2 font-mono text-lg tabular-nums text-muted-foreground", value > 0 && active)}>{value}</dd>
            </div>
          ))}
        </dl>
        {status.lastCommitMessage ? (
          <p className="mt-3 font-mono text-xs text-muted-foreground">
            {status.lastCommitHash.slice(0, 7)} {status.lastCommitMessage}
            {status.lastCommitAt ? ` · ${formatRelativeTime(status.lastCommitAt)}` : ""}
          </p>
        ) : null}
      </section>

      {status.issues.length ? (
        <section>
          <SectionHeading title="What is blocking sync" />
          <ItemGroup className="mt-5">
            {status.issues.map((issue) => (
              <Item key={issue} variant="outline">
                <ItemContent>
                  <ItemTitle>{issue}</ItemTitle>
                </ItemContent>
              </Item>
            ))}
          </ItemGroup>
        </section>
      ) : null}
    </div>
  )
}
