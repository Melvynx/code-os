import type { SkillsStatus } from "@/api/status"

export const SKILLS_STATE_LABELS = {
  synced: "Synced",
  dirty: "Pending local changes",
  ahead: "Ahead of origin",
  behind: "Behind origin",
  diverged: "Diverged from origin",
  locked: "Sync in progress",
  conflict: "Conflicts",
  invalid: "Checkout mismatch",
  missing: "Checkout missing",
  unconfigured: "Not configured",
} as const

export type SkillsState = keyof typeof SKILLS_STATE_LABELS

export function skillsStateLabel(state: SkillsStatus["state"]) {
  return SKILLS_STATE_LABELS[state]
}

export function skillsStateTone(state: SkillsStatus["state"]) {
  if (state === "synced") return "success" as const
  if (state === "dirty" || state === "ahead" || state === "locked" || state === "unconfigured") return "warning" as const
  return "error" as const
}

export function skillsSearchValues(status: SkillsStatus) {
  return [status.state, status.message, status.directory, status.repository, status.branch, status.origin, ...status.issues]
}

export function canRunSkillsSync(status: SkillsStatus) {
  return status.configured && status.state !== "conflict" && status.state !== "invalid" && status.state !== "missing"
}
