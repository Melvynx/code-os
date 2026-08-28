import { createFileRoute } from "@tanstack/react-router"
import { useEffect, useState, type FormEvent, type ReactNode } from "react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { SectionHeading } from "@/components/section-heading"

type Settings = {
  environmentName: string
  projectsRoots: string[]
  screenshotsRoot: string
  filesRoot: string
  dataDir: string
  portlyBinary: string
  publicPortHost: string
  cloudflare: {
    dashboardHost: string
    tunnelMode: string
    tunnelId: string
    accountId: string
    zoneId: string
    tokenFile: string
    requireAccess: boolean
  }
  auth: { username: string; passwordFile: string; bypassKeyFile: string; sessionKeyFile: string }
  skills: { repository: string; directory: string; branch: string }
  cloudflareTokenConfigured: boolean
  restartRequired?: boolean
}

export const Route = createFileRoute("/settings")({ component: SettingsPage })

function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null)
  const [token, setToken] = useState("")
  const [status, setStatus] = useState("Loading configuration…")
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    void fetch("/api/settings", { credentials: "same-origin" })
      .then(async (response) => {
        if (!response.ok) throw new Error(`Settings API returned ${response.status}`)
        return response.json() as Promise<Settings>
      })
      .then((value) => { setSettings(value); setStatus("") })
      .catch((error: Error) => setStatus(error.message))
  }, [])

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!settings) return
    setSaving(true)
    setStatus("Saving…")
    try {
	  const { cloudflareTokenConfigured: _configured, restartRequired: _restart, ...editable } = settings
      const response = await fetch("/api/settings", {
        method: "PUT",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify({ ...editable, cloudflareToken: token || undefined }),
      })
      const payload = await response.json() as Settings & { error?: string }
      if (!response.ok) throw new Error(payload.error || `Settings API returned ${response.status}`)
      setSettings(payload)
      setToken("")
      setStatus("Saved securely. Restart Code OS to activate the new configuration.")
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Could not save settings")
    } finally {
      setSaving(false)
    }
  }

  if (!settings) return <div className="border border-[#333] p-6 font-mono text-sm text-[#888]">{status}</div>

  const set = <K extends keyof Settings>(key: K, value: Settings[K]) => setSettings({ ...settings, [key]: value })

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
      </SettingsGroup>

      <div className="flex items-center justify-between border-t border-[#333] pt-6">
        <p aria-live="polite" className="font-mono text-xs text-[#888]">{status}</p>
        <Button type="submit" disabled={saving}>{saving ? "Saving" : "Save settings"}</Button>
      </div>
    </form>
  )
}

function SettingsGroup({ title, children }: Readonly<{ title: string; children: ReactNode }>) {
  return <fieldset className="border border-[#333] p-5"><legend className="px-2 font-mono text-xs uppercase tracking-wider text-[#58a6ff]">{title}</legend><div className="grid gap-5 md:grid-cols-2">{children}</div></fieldset>
}

function Field({ label, children, wide = false }: Readonly<{ label: string; children: ReactNode; wide?: boolean }>) {
  return <label className={wide ? "space-y-2 md:col-span-2" : "space-y-2"}><span className="block font-mono text-xs text-[#8b949e]">{label}</span>{children}</label>
}

const textareaClass = "min-h-28 w-full rounded-md border border-[#333] bg-black px-3 py-2 font-mono text-sm text-white outline-none transition-colors placeholder:text-[#888] hover:border-white"
