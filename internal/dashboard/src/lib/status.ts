import type { ImageSample, StatusCheck, StatusReport } from "@/api/status"

export const CHECK_GROUP_LABELS = {
  images: "Images",
  skills: "Skills",
  auth: "Authentication",
  environment: "Environment",
} as const

export type CheckGroup = keyof typeof CHECK_GROUP_LABELS

export function groupStatusChecks(checks: StatusCheck[]) {
  const groups = new Map<string, StatusCheck[]>()
  for (const check of checks) {
    const existing = groups.get(check.group)
    if (existing) existing.push(check)
    else groups.set(check.group, [check])
  }
  return (Object.keys(CHECK_GROUP_LABELS) as CheckGroup[])
    .filter((group) => groups.has(group))
    .map((group) => ({ id: group, label: CHECK_GROUP_LABELS[group], checks: groups.get(group) ?? [] }))
}

export function imagePipelineHealthy(images: StatusReport["images"]) {
  return images.filesRootExists && images.bypassPrivate && images.emptyCount === 0 && images.undecodableCount === 0
}

export function imageSampleLabel(sample: ImageSample) {
  if (sample.issue) return sample.issue
  if (sample.width && sample.height) return `${sample.width}×${sample.height}`
  if (sample.decodable) return "Decodable"
  return "Not verified"
}

export function statusSearchValues(report: StatusReport) {
  return [
    report.healthy ? "healthy" : "unhealthy",
    report.images.screenshotsRoot,
    report.images.filesRoot,
    ...report.checks.flatMap((check) => [check.label, check.detail, check.status]),
    ...report.images.recent.flatMap((sample) => [sample.name, sample.path, sample.issue]),
  ]
}
