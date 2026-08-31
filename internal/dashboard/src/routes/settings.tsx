import { createFileRoute } from "@tanstack/react-router"
import { useState, type FormEvent, type ReactNode } from "react"

import { useRevokeTrustedIP, useSaveSettings, useSettings, useTrustedIPStatus } from "@/api/settings-queries"
import type { Settings } from "@/api/settings"
import { PageError, PageLoading } from "@/components/page-state"
import { SectionHeading } from "@/components/section-heading"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

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

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setStatus("Saving…")
    const { cloudflareTokenConfigured: _configured, restartRequired: _restart, ...editable } = settings
    try {
      const saved = await save.mutateAsync({ ...editable, cloudflareToken: token || undefined })
      setSettings(saved)
      setToken("")
      setStatus("Saved securely. Restart Code OS to activate the new configuration.")
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Could not save settings")
    }
  }

  return (
    <form className="space-y-8" onSubmit={submit}>
      <SectionHeading title="Environment settings" description="Configure Code OS without exposing stored secrets. Changes are validated and require a service restart." />
      <SettingsGroup title="Environment">
        <Field label="Environment name"><Input value={settings.environmentName} onChange={(event) => set("environmentName", event.target.value)} /></Field>
        <Field label="Project roots (one per line)" wide><textarea className={textareaClass} value={settings.projectsRoots.join("\n")} onChange={(event) => set("projectsRoots", event.target.value.split("\n").filter(Boolean))} /></Field>
        <Field label="Screenshots root"><Input value={settings.screenshotsRoot} onChange={(event) => set("screenshotsRoot", event.target.value)} /></Field>
        <Field label="Private files root"><Input value={settings.filesRoot} onChange={(event) => set("filesRoot", event.target.value)} /></Field>
        <Field label="Data directory"><Input value={settings.dataDir} onChange={(event) => set("dataDir", event.target.value)} /></Field>
        <Field label="Portly binary"><Input value={settings.portlyBinary} onChange={(event) => set("portlyBinary", event.target.value)} /></Field>
      </SettingsGroup>

      <SettingsGroup title="Cloudflare gateway">
        <Field label="Code OS hostname"><Input value={settings.cloudflare.dashboardHost} onChange={(event) => setSettings({ ...settings, cloudflare: { ...settings.cloudflare, dashboardHost: event.target.value } })} /></Field>
        <Field label="Application host template"><Input value={settings.publicPortHost} onChange={(event) => set("publicPortHost", event.target.value)} placeholder="port{port}.example.com" /></Field>
        <Field label="Tunnel ID"><Input value={settings.cloudflare.tunnelId ?? ""} onChange={(event) => setSettings({ ...settings, cloudflare: { ...settings.cloudflare, tunnelId: event.target.value } })} /></Field>
        <Field label="Account ID"><Input value={settings.cloudflare.accountId ?? ""} onChange={(event) => setSettings({ ...settings, cloudflare: { ...settings.cloudflare, accountId: event.target.value } })} /></Field>
        <Field label="Zone ID"><Input value={settings.cloudflare.zoneId ?? ""} onChange={(event) => setSettings({ ...settings, cloudflare: { ...settings.cloudflare, zoneId: event.target.value } })} /></Field>
        <Field label={`Cloudflare token (${settings.cloudflareTokenConfigured ? "configured" : "missing"})`}><Input type="password" autoComplete="off" value={token} onChange={(event) => setToken(event.target.value)} placeholder="Leave blank to keep the current token" /></Field>
        <Field label="Token file"><Input value={settings.cloudflare.tokenFile ?? ""} onChange={(event) => setSettings({ ...settings, cloudflare: { ...settings.cloudflare, tokenFile: event.target.value } })} /></Field>
      </SettingsGroup>

      <SettingsGroup title="Skills synchronization">
        <Field label="GitHub repository" wide><Input value={settings.skills.repository ?? ""} onChange={(event) => setSettings({ ...settings, skills: { ...settings.skills, repository: event.target.value } })} placeholder="git@github.com:owner/agents-config.git" /></Field>
        <Field label="Local checkout"><Input value={settings.skills.directory ?? ""} onChange={(event) => setSettings({ ...settings, skills: { ...settings.skills, directory: event.target.value } })} /></Field>
        <Field label="Branch"><Input value={settings.skills.branch ?? ""} onChange={(event) => setSettings({ ...settings, skills: { ...settings.skills, branch: event.target.value } })} /></Field>
      </SettingsGroup>

      <SettingsGroup title="Authentication">
        <Field label="Username"><Input autoComplete="username" value={settings.auth.username ?? ""} onChange={(event) => setSettings({ ...settings, auth: { ...settings.auth, username: event.target.value } })} /></Field>
        <Field label="Password file"><Input value={settings.auth.passwordFile ?? ""} onChange={(event) => setSettings({ ...settings, auth: { ...settings.auth, passwordFile: event.target.value } })} /></Field>
        <Field label="Media bypass key file"><Input value={settings.auth.bypassKeyFile ?? ""} onChange={(event) => setSettings({ ...settings, auth: { ...settings.auth, bypassKeyFile: event.target.value } })} /></Field>
        <Field label="Session signing key file"><Input value={settings.auth.sessionKeyFile ?? ""} onChange={(event) => setSettings({ ...settings, auth: { ...settings.auth, sessionKeyFile: event.target.value } })} /></Field>
        <Field label="Trusted IP file" wide><Input value={settings.auth.trustedIPsFile ?? ""} onChange={(event) => setSettings({ ...settings, auth: { ...settings.auth, trustedIPsFile: event.target.value } })} /></Field>
        <Field label="Current IP trust" wide><TrustedIPControl /></Field>
      </SettingsGroup>

      <div className="flex items-center justify-between border-t border-[#333] pt-6">
        <p aria-live="polite" className="font-mono text-xs text-[#888]">{status}</p>
        <Button type="submit" disabled={save.isPending}>{save.isPending ? "Saving" : "Save settings"}</Button>
      </div>
    </form>
  )
}

function TrustedIPControl() {
  const status = useTrustedIPStatus()
  const revoke = useRevokeTrustedIP()
  if (status.isPending) return <div className="border border-[#333] p-4 font-mono text-xs text-[#888]">Reading current IP…</div>
  if (status.isError) return <div role="alert" className="border border-[#e00] p-4 text-sm text-white">{status.error.message}</div>
  if (!status.data.configured) return <div className="border border-[#333] p-4 text-sm text-[#888]">Configure a trusted IP file and restart Code OS to enable this feature.</div>
  return (
    <div className="space-y-3">
      <div className="flex flex-col gap-4 border border-[#333] p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="font-mono text-sm text-white">{status.data.currentIP}</p>
          <p className="mt-1 text-sm text-[#888]">{status.data.trusted ? "This IP skips sign-in." : "This IP still requires sign-in."} {status.data.count} trusted {status.data.count === 1 ? "address" : "addresses"}.</p>
        </div>
        {status.data.trusted ? <Button type="button" variant="destructive" disabled={revoke.isPending} onClick={() => revoke.mutate()}>{revoke.isPending ? "Revoking" : "Stop trusting this IP"}</Button> : null}
      </div>
      {revoke.isError ? <p role="alert" className="text-sm text-[#ff7b72]">{revoke.error.message}</p> : null}
    </div>
  )
}

function SettingsGroup({ title, children }: Readonly<{ title: string; children: ReactNode }>) {
  return <fieldset className="border border-[#333] p-5"><legend className="px-2 font-mono text-xs uppercase tracking-wider text-[#58a6ff]">{title}</legend><div className="grid gap-5 md:grid-cols-2">{children}</div></fieldset>
}

function Field({ label, children, wide = false }: Readonly<{ label: string; children: ReactNode; wide?: boolean }>) {
  return <label className={wide ? "space-y-2 md:col-span-2" : "space-y-2"}><span className="block font-mono text-xs text-[#8b949e]">{label}</span>{children}</label>
}

const textareaClass = "min-h-28 w-full rounded-md border border-[#333] bg-black px-3 py-2 font-mono text-sm text-white outline-none transition-colors placeholder:text-[#888] hover:border-white"
