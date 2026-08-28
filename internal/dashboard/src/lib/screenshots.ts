import type { Screenshot } from "@/api/schema"

export interface ScreenshotFeatureGroup {
  name: string
  screenshots: Screenshot[]
}

export function screenshotFeature(screenshot: Screenshot) {
  return screenshot.project?.trim() || screenshot.group?.trim() || "Other"
}

export function groupScreenshotsByFeature(screenshots: Screenshot[]) {
  const groups = new Map<string, Screenshot[]>()

  for (const screenshot of screenshots) {
    const feature = screenshotFeature(screenshot)
    const group = groups.get(feature)

    if (group) {
      group.push(screenshot)
    } else {
      groups.set(feature, [screenshot])
    }
  }

  return Array.from(groups, ([name, featureScreenshots]) => ({
    name,
    screenshots: featureScreenshots,
  }))
}
