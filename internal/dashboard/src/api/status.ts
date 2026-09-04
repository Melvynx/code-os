import { z } from "zod"

import { ApiError } from "./client"
import { gitStateSchema } from "./schema"

const optionalTimestamp = z.string().nullish()

const skillsStatusSchema = z.object({
  state: z.enum(["unconfigured", "missing", "invalid", "conflict", "diverged", "behind", "ahead", "dirty", "locked", "synced"]),
  synced: z.boolean(),
  configured: z.boolean(),
  message: z.string(),
  repository: z.string().optional().default(""),
  directory: z.string().optional().default(""),
  branch: z.string().optional().default(""),
  currentBranch: z.string().optional().default(""),
  origin: z.string().optional().default(""),
  originMatches: z.boolean(),
  git: gitStateSchema,
  lastCommitHash: z.string().optional().default(""),
  lastCommitMessage: z.string().optional().default(""),
  lastCommitAt: optionalTimestamp,
  lastSyncAt: optionalTimestamp,
  timerEnabled: z.boolean(),
  timerActive: z.boolean(),
  lastTimerAt: optionalTimestamp,
  nextTimerAt: optionalTimestamp,
  lastTimerResult: z.string().optional().default(""),
  lockHeld: z.boolean(),
  issues: z.array(z.string()).nullish().transform((items) => items ?? []),
})

const checkSchema = z.object({
  id: z.string(),
  label: z.string(),
  detail: z.string(),
  status: z.enum(["pass", "warn", "fail"]),
  group: z.string(),
})

const imageSampleSchema = z.object({
  name: z.string(),
  path: z.string(),
  url: z.string().optional().default(""),
  kind: z.string(),
  size: z.number().nonnegative(),
  width: z.number().int().nonnegative().optional().default(0),
  height: z.number().int().nonnegative().optional().default(0),
  decodable: z.boolean(),
  createdAt: z.string(),
  issue: z.string().optional().default(""),
})

const statusReportSchema = z.object({
  generatedAt: z.string(),
  healthy: z.boolean(),
  failed: z.number().int().nonnegative(),
  warnings: z.number().int().nonnegative(),
  checks: z.array(checkSchema),
  images: z.object({
    screenshotsRoot: z.string(),
    filesRoot: z.string(),
    sharedRoots: z.boolean(),
    screenshotsRootExists: z.boolean(),
    filesRootExists: z.boolean(),
    indexedCount: z.number().int().nonnegative(),
    filesImageCount: z.number().int().nonnegative(),
    decodableCount: z.number().int().nonnegative(),
    emptyCount: z.number().int().nonnegative(),
    undecodableCount: z.number().int().nonnegative(),
    bypassConfigured: z.boolean(),
    bypassPrivate: z.boolean(),
    recent: z.array(imageSampleSchema).nullish().transform((items) => items ?? []),
  }),
  skills: skillsStatusSchema,
})

const skillsSyncResponseSchema = z.object({
  result: z.object({
    alreadyRunning: z.boolean().optional(),
    cloned: z.boolean().optional(),
    message: z.string(),
  }),
  status: skillsStatusSchema,
})

export type SkillsStatus = z.infer<typeof skillsStatusSchema>
export type StatusReport = z.infer<typeof statusReportSchema>
export type StatusCheck = z.infer<typeof checkSchema>
export type ImageSample = z.infer<typeof imageSampleSchema>

async function readJSON<T>(response: Response, schema: z.ZodType<T>, fallback: string): Promise<T> {
  if (response.status === 401) {
    const next = `${window.location.pathname}${window.location.search}${window.location.hash}`
    window.location.assign(`/login?next=${encodeURIComponent(next)}`)
    throw new ApiError("Code OS session expired", response.status)
  }
  const payload: unknown = await response.json()
  if (!response.ok) {
    const message = typeof payload === "object" && payload !== null && "error" in payload && typeof payload.error === "string"
      ? payload.error
      : fallback
    throw new ApiError(message, response.status)
  }
  return schema.parse(payload)
}

export async function fetchSkillsStatus() {
  const response = await fetch("/api/skills-sync", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  return readJSON(response, skillsStatusSchema, `Skills sync API returned ${response.status}`)
}

export async function runSkillsSync() {
  const response = await fetch("/api/skills-sync", {
    method: "POST",
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  return readJSON(response, skillsSyncResponseSchema, `Could not synchronize skills (${response.status})`)
}

export async function fetchStatusReport() {
  const response = await fetch("/api/status", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  return readJSON(response, statusReportSchema, `Status API returned ${response.status}`)
}
