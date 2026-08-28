import { snapshotSchema, type Snapshot } from "./schema"

export class ApiError extends Error {
  readonly statusCode: number

  constructor(message: string, statusCode: number) {
    super(message)
    this.name = "ApiError"
    this.statusCode = statusCode
  }
}

async function readSnapshot(response: Response): Promise<Snapshot> {
  if (response.status === 401) {
    const next = `${window.location.pathname}${window.location.search}${window.location.hash}`
    window.location.assign(`/login?next=${encodeURIComponent(next)}`)
    throw new ApiError("StackEnv session expired", response.status)
  }

  if (!response.ok) {
    throw new ApiError(`StackEnv API returned ${response.status}`, response.status)
  }

  const payload: unknown = await response.json()
  return snapshotSchema.parse(payload)
}

export async function fetchSnapshot() {
  const response = await fetch("/api/snapshot", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  return readSnapshot(response)
}

export async function refreshSnapshot() {
  const response = await fetch("/api/refresh", {
    method: "POST",
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  return readSnapshot(response)
}
