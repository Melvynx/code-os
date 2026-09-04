import { z } from "zod"

import { ApiError } from "./client"

const pointSchema = z.object({
  at: z.string(),
  cpuPercent: z.number(),
  memoryBytes: z.number().nonnegative(),
})

const resourceHistorySchema = z.object({
  generatedAt: z.string().optional().default(""),
  sampleCount: z.number().int().nonnegative(),
  retentionHours: z.number().int().nonnegative().optional().default(6),
  series: z.array(z.object({
    id: z.string(),
    kind: z.enum(["application", "agent"]),
    name: z.string(),
    points: z.array(pointSchema),
  })).nullish().transform((items) => items ?? []),
})

export type ResourceHistory = z.infer<typeof resourceHistorySchema>
export type ResourceSeries = ResourceHistory["series"][number]

export async function fetchResourceHistory() {
  const response = await fetch("/api/resources", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  if (response.status === 401) {
    const next = `${window.location.pathname}${window.location.search}${window.location.hash}`
    window.location.assign(`/login?next=${encodeURIComponent(next)}`)
    throw new ApiError("Code OS session expired", response.status)
  }
  if (!response.ok) {
    throw new ApiError(`Resources API returned ${response.status}`, response.status)
  }
  return resourceHistorySchema.parse(await response.json())
}
