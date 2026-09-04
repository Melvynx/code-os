import type { GitState, Project } from "@/api/schema"

const BYTE_UNITS = ["B", "KB", "MB", "GB", "TB"] as const
const SECOND = 1_000
const MINUTE = 60 * SECOND
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

export function formatBytes(value: number) {
  if (value <= 0) return "—"

  const unitIndex = Math.min(
    Math.floor(Math.log(value) / Math.log(1_024)),
    BYTE_UNITS.length - 1,
  )
  const amount = value / 1_024 ** unitIndex
  const precision = unitIndex > 1 ? 1 : 0

  return `${amount.toFixed(precision)} ${BYTE_UNITS[unitIndex]}`
}

export function formatRelativeTime(value: string, currentTime = Date.now()) {
  const elapsed = currentTime - new Date(value).getTime()
  const duration = Math.abs(elapsed)
  const unit = duration < MINUTE
    ? `${Math.floor(duration / SECOND)}s`
    : duration < HOUR
      ? `${Math.floor(duration / MINUTE)}m`
      : duration < DAY
        ? `${Math.floor(duration / HOUR)}h`
        : `${Math.floor(duration / DAY)}d`
  return elapsed >= 0 ? `${unit} ago` : `in ${unit}`
}

export function getChangeCount(project: Project) {
  return getGitChangeCount(project.git)
}

export function getGitChangeCount(git: GitState) {
  const { added, conflicts, deleted, modified, untracked } = git
  return added + conflicts + deleted + modified + untracked
}

export function matchesQuery(query: string, values: Array<string | number | undefined>) {
  if (!query) return true
  const normalizedQuery = query.trim().toLocaleLowerCase()
  return values.some((value) =>
    String(value ?? "")
      .toLocaleLowerCase()
      .includes(normalizedQuery),
  )
}
