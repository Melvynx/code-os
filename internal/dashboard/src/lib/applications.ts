import type { Application } from "@/api/schema"

const PRIMARY_ROLES = new Set(["app", "dashboard", "dev", "frontend", "next", "start", "start-all", "ui", "vite", "web"])
const SUPPORTING_ROLES = new Set([
  "convex", "hatchet-tunnel", "hatchet-worker", "jobs", "minio", "postgres", "proxy",
  "qstash", "redis", "static", "sync", "tunnel", "worker",
])
const ROLE_LABELS: Record<string, string> = {
  "start-all": "Start",
  convex: "Convex",
  dashboard: "Dashboard",
  dev: "Dev",
  start: "Start",
  static: "Static",
  sync: "Sync",
  web: "Web",
}
const ACRONYMS = new Set(["ai", "os", "ui"])

export function applicationIconUrl(application: Pick<Application, "id">) {
  return `/api/applications/${encodeURIComponent(application.id)}/icon`
}

export function applicationTitle(application: Pick<Application, "projectName">) {
  const source = application.projectName.includes(".")
    ? application.projectName.slice(0, application.projectName.indexOf("."))
    : application.projectName
  return source
    .split(/[-_]+/)
    .filter(Boolean)
    .map((part) => {
      const lower = part.toLocaleLowerCase()
      if (ACRONYMS.has(lower)) return part.toUpperCase()
      return part.charAt(0).toUpperCase() + part.slice(1)
    })
    .join(" ")
}

export function applicationRole(application: Pick<Application, "name">) {
  const key = application.name.trim().toLocaleLowerCase()
  return ROLE_LABELS[key] ?? application.name.replace(/[-_]+/g, " ").replace(/\b\w/g, (char) => char.toUpperCase())
}

export function applicationDomain(application: Pick<Application, "domain" | "projectName" | "publicUrl" | "url">) {
  if (application.domain) return application.domain
  const publicHost = hostnameOf(application.publicUrl)
  if (publicHost && !publicHost.startsWith("port") && publicHost !== "localhost") return publicHost
  return hostnameOf(application.url) === "localhost" ? "" : hostnameOf(application.url)
}

export function applicationHref(application: Pick<Application, "publicUrl" | "url">) {
  return safeHttpUrl(application.publicUrl) ?? safeHttpUrl(application.url)
}

export function applicationHostLabel(application: Pick<Application, "port" | "publicUrl" | "url">) {
  const href = applicationHref(application)
  if (href) return new URL(href).host
  return application.port ? `localhost:${application.port}` : ""
}

export function isSupportingApplication(application: Pick<Application, "name">) {
  const role = application.name.trim().toLocaleLowerCase()
  return SUPPORTING_ROLES.has(role) || role.includes("worker") || role.includes("tunnel")
}

export function isPrimaryApplication(application: Pick<Application, "name" | "port" | "publicUrl" | "url">) {
  if (isSupportingApplication(application)) return false
  const role = application.name.trim().toLocaleLowerCase()
  const hasAddress = application.port > 0 || Boolean(application.publicUrl || application.url)
  if (!hasAddress) return false
  return PRIMARY_ROLES.has(role) || application.port > 0
}

export function isFeaturedApplication(application: Application) {
  return application.state === "running" && application.port > 0 && isPrimaryApplication(application)
}

export function sortApplications(applications: Application[]) {
  return [...applications].sort((left, right) => {
    const rank = applicationRank(left) - applicationRank(right)
    if (rank !== 0) return rank
    const memory = applicationMemory(right) - applicationMemory(left)
    if (memory !== 0) return memory
    return `${left.projectName}/${left.name}`.localeCompare(`${right.projectName}/${right.name}`)
  })
}

export function featuredApplications(applications: Application[]) {
  return sortApplications(applications.filter(isFeaturedApplication))
}

export function applicationMemory(application: Pick<Application, "memoryBytes" | "residentMemoryBytes">) {
  return application.residentMemoryBytes || application.memoryBytes
}

export function applicationDisplayName(application: Pick<Application, "name" | "projectName">) {
  return `${applicationTitle(application)} ${applicationRole(application)}`.trim()
}

export function applicationSearchValues(application: Application) {
  return [
    applicationTitle(application),
    applicationRole(application),
    applicationDisplayName(application),
    applicationDomain(application),
    applicationHostLabel(application),
    application.projectName,
    application.name,
    application.command,
    application.directory,
    application.domain,
    application.publicUrl,
    application.url,
    application.port,
    application.state,
  ]
}

export function applicationMatchesQuery(application: Application, query: string) {
  const tokens = query.trim().toLocaleLowerCase().split(/\s+/).filter(Boolean)
  if (!tokens.length) return true
  const haystack = applicationSearchValues(application).map((value) => String(value ?? "").toLocaleLowerCase())
  return tokens.every((token) => haystack.some((value) => value.includes(token)))
}

export function safeHttpUrl(value: string | undefined) {
  if (!value) return null
  try {
    const url = new URL(value, typeof window === "undefined" ? "http://localhost" : window.location.origin)
    return url.protocol === "http:" || url.protocol === "https:" ? url.href : null
  } catch {
    return null
  }
}

function applicationRank(application: Application) {
  const running = application.state === "running"
  if (running && isFeaturedApplication(application)) return 0
  if (running && application.port > 0) return 1
  if (running) return 2
  if (application.port > 0) return 3
  return 4
}

function hostnameOf(value: string | undefined) {
  const href = safeHttpUrl(value)
  return href ? new URL(href).hostname : ""
}
