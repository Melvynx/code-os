import { createFileRoute } from "@tanstack/react-router"
import { PlusIcon } from "lucide-react"
import { useId, useState, type FormEvent, type ReactNode } from "react"
import { toast } from "sonner"

import { useRevokeTrustedIP, useSaveSettings, useSettings, useTrustedIPStatus, useTrustCurrentIP } from "@/api/settings-queries"
import type { Settings } from "@/api/settings"
import { PageError, PageLoading } from "@/components/page-state"
import { SectionHeading } from "@/components/section-heading"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Field, FieldDescription, FieldGroup, FieldLabel, FieldLegend, FieldSet } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import { Spinner } from "@/components/ui/spinner"
import { Textarea } from "@/components/ui/textarea"

export const Route = createFileRoute("/settings")({ component: SettingsPage })

function SettingsPage() {
  const settings = useSettings()
  if (settings.isPending) return <PageLoading label="Loading Code OS settings" />
  if (settings.isError) return <PageError message={settings.error.message} retry={() => void settings.refetch()} />
  return <SettingsForm key={JSON.stringify(settings.data)} initialSettings={settings.data} />
}

function SettingsForm({ initialSettings }: Readonly<{ initialSettings: Settings }>) {
  const [settings, setSettings] = useState(initialSettings)
  const [token, setToken] = useState("")
  const [status, setStatus] = useState("")
  const save = useSaveSettings()
  const set = <K extends keyof Settings>(key: K, value: Settings[K]) => setSettings({ ...settings, [key]: value })
  const ids = {
    environmentName: useId(),
    projectsRoots: useId(),
    screenshotsRoot: useId(),
    filesRoot: useId(),
    dataDir: useId(),
    portlyBinary: useId(),
    dashboardHost: useId(),
    publicPortHost: useId(),
    tunnelId: useId(),
    accountId: useId(),
    zoneId: useId(),
    cloudflareToken: useId(),
    tokenFile: useId(),
    skillsRepository: useId(),
    skillsDirectory: useId(),
    skillsBranch: useId(),
    username: useId(),
    passwordFile: useId(),
    bypassKeyFile: useId(),
    sessionKeyFile: useId(),
    trustedIPsFile: useId(),
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setStatus("Saving…")
    const { cloudflareTokenConfigured: _configured, restartRequired: _restart, ...editable } = settings
    try {
      const saved = await save.mutateAsync({ ...editable, cloudflareToken: token || undefined })
      setSettings(saved)
      setToken("")
      const message = "Saved securely. Restart Code OS to activate the new configuration."
      setStatus(message)
      toast.success(message)
    } catch (error) {
      const message = error instanceof Error ? error.message : "Could not save settings"
      setStatus(message)
      toast.error(message)
    }
  }

  return (
    <form className="flex flex-col gap-8" onSubmit={submit}>
      <SectionHeading title="Environment settings" description="Configure Code OS without exposing stored secrets. Changes are validated and require a service restart." />
      <SettingsGroup title="Environment">
        <Field>
          <FieldLabel htmlFor={ids.environmentName}>Environment name</FieldLabel>
          <Input id={ids.environmentName} value={settings.environmentName} onChange={(event) => set("environmentName", event.target.value)} />
        </Field>
        <Field className="md:col-span-2">
          <FieldLabel htmlFor={ids.projectsRoots}>Project roots (one per line)</FieldLabel>
          <Textarea id={ids.projectsRoots} className="min-h-28 font-mono" value={settings.projectsRoots.join("\n")} onChange={(event) => set("projectsRoots", event.target.value.split("\n").filter(Boolean))} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.screenshotsRoot}>Screenshots root</FieldLabel>
          <Input id={ids.screenshotsRoot} value={settings.screenshotsRoot} onChange={(event) => set("screenshotsRoot", event.target.value)} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.filesRoot}>Private files root</FieldLabel>
          <Input id={ids.filesRoot} value={settings.filesRoot} onChange={(event) => set("filesRoot", event.target.value)} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.dataDir}>Data directory</FieldLabel>
          <Input id={ids.dataDir} value={settings.dataDir} onChange={(event) => set("dataDir", event.target.value)} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.portlyBinary}>Portly binary</FieldLabel>
          <Input id={ids.portlyBinary} value={settings.portlyBinary} onChange={(event) => set("portlyBinary", event.target.value)} />
        </Field>
      </SettingsGroup>

      <SettingsGroup title="Cloudflare gateway">
        <Field>
          <FieldLabel htmlFor={ids.dashboardHost}>Code OS hostname</FieldLabel>
          <Input id={ids.dashboardHost} value={settings.cloudflare.dashboardHost} onChange={(event) => setSettings({ ...settings, cloudflare: { ...settings.cloudflare, dashboardHost: event.target.value } })} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.publicPortHost}>Application host template</FieldLabel>
          <Input id={ids.publicPortHost} value={settings.publicPortHost} onChange={(event) => set("publicPortHost", event.target.value)} placeholder="port{port}.example.com" />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.tunnelId}>Tunnel ID</FieldLabel>
          <Input id={ids.tunnelId} value={settings.cloudflare.tunnelId ?? ""} onChange={(event) => setSettings({ ...settings, cloudflare: { ...settings.cloudflare, tunnelId: event.target.value } })} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.accountId}>Account ID</FieldLabel>
          <Input id={ids.accountId} value={settings.cloudflare.accountId ?? ""} onChange={(event) => setSettings({ ...settings, cloudflare: { ...settings.cloudflare, accountId: event.target.value } })} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.zoneId}>Zone ID</FieldLabel>
          <Input id={ids.zoneId} value={settings.cloudflare.zoneId ?? ""} onChange={(event) => setSettings({ ...settings, cloudflare: { ...settings.cloudflare, zoneId: event.target.value } })} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.cloudflareToken}>Cloudflare token ({settings.cloudflareTokenConfigured ? "configured" : "missing"})</FieldLabel>
          <Input id={ids.cloudflareToken} type="password" autoComplete="off" value={token} onChange={(event) => setToken(event.target.value)} placeholder="Leave blank to keep the current token" />
        </Field>
        <Field className="md:col-span-2">
          <FieldLabel htmlFor={ids.tokenFile}>Token file</FieldLabel>
          <Input id={ids.tokenFile} value={settings.cloudflare.tokenFile ?? ""} onChange={(event) => setSettings({ ...settings, cloudflare: { ...settings.cloudflare, tokenFile: event.target.value } })} />
        </Field>
      </SettingsGroup>

      <SettingsGroup title="Skills synchronization">
        <Field className="md:col-span-2">
          <FieldLabel htmlFor={ids.skillsRepository}>GitHub repository</FieldLabel>
          <Input id={ids.skillsRepository} value={settings.skills.repository ?? ""} onChange={(event) => setSettings({ ...settings, skills: { ...settings.skills, repository: event.target.value } })} placeholder="git@github.com:owner/agents-config.git" />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.skillsDirectory}>Local checkout</FieldLabel>
          <Input id={ids.skillsDirectory} value={settings.skills.directory ?? ""} onChange={(event) => setSettings({ ...settings, skills: { ...settings.skills, directory: event.target.value } })} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.skillsBranch}>Branch</FieldLabel>
          <Input id={ids.skillsBranch} value={settings.skills.branch ?? ""} onChange={(event) => setSettings({ ...settings, skills: { ...settings.skills, branch: event.target.value } })} />
        </Field>
      </SettingsGroup>

      <SettingsGroup title="Authentication">
        <Field>
          <FieldLabel htmlFor={ids.username}>Username</FieldLabel>
          <Input id={ids.username} autoComplete="username" value={settings.auth.username ?? ""} onChange={(event) => setSettings({ ...settings, auth: { ...settings.auth, username: event.target.value } })} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.passwordFile}>Password file</FieldLabel>
          <Input id={ids.passwordFile} value={settings.auth.passwordFile ?? ""} onChange={(event) => setSettings({ ...settings, auth: { ...settings.auth, passwordFile: event.target.value } })} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.bypassKeyFile}>Media bypass key file</FieldLabel>
          <Input id={ids.bypassKeyFile} value={settings.auth.bypassKeyFile ?? ""} onChange={(event) => setSettings({ ...settings, auth: { ...settings.auth, bypassKeyFile: event.target.value } })} />
        </Field>
        <Field>
          <FieldLabel htmlFor={ids.sessionKeyFile}>Session signing key file</FieldLabel>
          <Input id={ids.sessionKeyFile} value={settings.auth.sessionKeyFile ?? ""} onChange={(event) => setSettings({ ...settings, auth: { ...settings.auth, sessionKeyFile: event.target.value } })} />
        </Field>
        <Field className="md:col-span-2">
          <FieldLabel htmlFor={ids.trustedIPsFile}>Trusted IP file</FieldLabel>
          <Input id={ids.trustedIPsFile} value={settings.auth.trustedIPsFile ?? ""} onChange={(event) => setSettings({ ...settings, auth: { ...settings.auth, trustedIPsFile: event.target.value } })} />
        </Field>
        <Field className="md:col-span-2">
          <FieldLabel>Current IP trust</FieldLabel>
          <TrustedIPControl />
        </Field>
      </SettingsGroup>

      <Separator />
      <div className="flex items-center justify-between">
        <p aria-live="polite" className="font-mono text-xs text-muted-foreground">{status}</p>
        <Button type="submit" disabled={save.isPending}>
          {save.isPending ? <Spinner data-icon="inline-start" /> : null}
          {save.isPending ? "Saving" : "Save settings"}
        </Button>
      </div>
    </form>
  )
}

function TrustedIPControl() {
  const status = useTrustedIPStatus()
  const trust = useTrustCurrentIP()
  const revoke = useRevokeTrustedIP()
  if (status.isPending) return <FieldDescription>Reading current IP…</FieldDescription>
  if (status.isError) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Unable to read the current IP</AlertTitle>
        <AlertDescription>{status.error.message}</AlertDescription>
      </Alert>
    )
  }
  if (!status.data.configured) {
    return (
      <Alert>
        <AlertTitle>Trusted IPs are disabled</AlertTitle>
        <AlertDescription>Configure a trusted IP file and restart Code OS to enable this feature.</AlertDescription>
      </Alert>
    )
  }
  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col gap-4 rounded-xl border p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="font-mono text-sm">{status.data.currentIP}</p>
          <p className="mt-1 text-sm text-muted-foreground">{status.data.trusted ? "This IP skips sign-in." : "This IP still requires sign-in."} {status.data.count} trusted {status.data.count === 1 ? "address" : "addresses"}.</p>
        </div>
        {status.data.trusted ? (
          <Button type="button" variant="destructive" disabled={revoke.isPending} onClick={() => revoke.mutate()}>
            {revoke.isPending ? <Spinner data-icon="inline-start" /> : null}
            {revoke.isPending ? "Revoking" : "Stop trusting this IP"}
          </Button>
        ) : (
          <Button type="button" disabled={trust.isPending} onClick={() => trust.mutate()}>
            {trust.isPending ? <Spinner data-icon="inline-start" /> : <PlusIcon data-icon="inline-start" />}
            {trust.isPending ? "Trusting…" : "Trust this IP"}
          </Button>
        )}
      </div>
      {trust.isError || revoke.isError ? (
        <Alert variant="destructive">
          <AlertTitle>Trusted IP update failed</AlertTitle>
          <AlertDescription>{(trust.error ?? revoke.error)?.message}</AlertDescription>
        </Alert>
      ) : null}
    </div>
  )
}

function SettingsGroup({ title, children }: Readonly<{ title: string; children: ReactNode }>) {
  return (
    <FieldSet className="rounded-xl border p-5">
      <FieldLegend>{title}</FieldLegend>
      <FieldGroup className="grid gap-5 md:grid-cols-2">{children}</FieldGroup>
    </FieldSet>
  )
}
