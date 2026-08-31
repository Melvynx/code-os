import { z } from "zod"

import { ApiError } from "./client"

const optionalString = z.string().optional().default("")

const settingsSchema = z.object({
  environmentName: z.string(),
  projectsRoots: z.array(z.string()),
  screenshotsRoot: z.string(),
  filesRoot: z.string(),
  dataDir: z.string(),
  portlyBinary: z.string(),
  publicPortHost: z.string(),
  cloudflare: z.object({
    dashboardHost: z.string(),
    tunnelMode: z.string(),
    tunnelId: optionalString,
    accountId: optionalString,
    zoneId: optionalString,
    tokenFile: optionalString,
    requireAccess: z.boolean(),
  }),
  auth: z.object({
    username: optionalString,
    passwordFile: optionalString,
    bypassKeyFile: optionalString,
    sessionKeyFile: optionalString,
    trustedIPsFile: optionalString,
  }),
  skills: z.object({ repository: optionalString, directory: optionalString, branch: optionalString }),
  cloudflareTokenConfigured: z.boolean(),
  restartRequired: z.boolean().optional(),
})

const trustedIPStatusSchema = z.object({
  currentIP: z.string(),
  trusted: z.boolean(),
  configured: z.boolean(),
  count: z.number().int().nonnegative(),
})

export type Settings = z.infer<typeof settingsSchema>
export type TrustedIPStatus = z.infer<typeof trustedIPStatusSchema>

export type SettingsUpdate = Omit<Settings, "cloudflareTokenConfigured" | "restartRequired"> & {
  cloudflareToken?: string
}

async function readJSON<T>(response: Response, schema: z.ZodType<T>): Promise<T> {
  if (response.status === 401) {
    const next = `${window.location.pathname}${window.location.search}${window.location.hash}`
    window.location.assign(`/login?next=${encodeURIComponent(next)}`)
    throw new ApiError("Code OS session expired", response.status)
  }
  const payload: unknown = await response.json()
  if (!response.ok) {
    const message = typeof payload === "object" && payload !== null && "error" in payload && typeof payload.error === "string"
      ? payload.error
      : `Settings API returned ${response.status}`
    throw new ApiError(message, response.status)
  }
  return schema.parse(payload)
}

export async function fetchSettings() {
  const response = await fetch("/api/settings", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  return readJSON(response, settingsSchema)
}

export async function saveSettings(update: SettingsUpdate) {
  const response = await fetch("/api/settings", {
    method: "PUT",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify(update),
  })
  return readJSON(response, settingsSchema)
}

export async function fetchTrustedIPStatus() {
  const response = await fetch("/api/trusted-ip", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  return readJSON(response, trustedIPStatusSchema)
}

export async function revokeCurrentTrustedIP() {
  const response = await fetch("/api/trusted-ip", {
    method: "DELETE",
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  return readJSON(response, trustedIPStatusSchema)
}
