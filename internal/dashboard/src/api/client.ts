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
    throw new ApiError("Code OS session expired", response.status)
  }

  if (!response.ok) {
    throw new ApiError(`Code OS API returned ${response.status}`, response.status)
  }

  const payload: unknown = await response.json()
  return snapshotSchema.parse(payload)
}

async function readError(response: Response, fallback: string) {
  if (response.status === 401) {
    const next = `${window.location.pathname}${window.location.search}${window.location.hash}`
    window.location.assign(`/login?next=${encodeURIComponent(next)}`)
    throw new ApiError("Code OS session expired", response.status)
  }
  let message = fallback
  try {
    const payload: unknown = await response.json()
    if (typeof payload === "object" && payload !== null && "error" in payload && typeof payload.error === "string") {
      message = payload.error
    }
  } catch {
    // Keep the transport-level fallback when the server did not return JSON.
  }
  throw new ApiError(message, response.status)
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

export async function stopApplication(id: string) {
  const response = await fetch(`/api/applications/${encodeURIComponent(id)}/stop`, {
    method: "POST",
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  if (!response.ok) return readError(response, `Could not stop application (${response.status})`)
  return readSnapshot(response)
}

export async function terminateAgent(id: string) {
  const response = await fetch(`/api/agents/${encodeURIComponent(id)}/terminate`, {
    method: "POST",
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  })
  if (!response.ok) return readError(response, `Could not terminate agent (${response.status})`)
}
