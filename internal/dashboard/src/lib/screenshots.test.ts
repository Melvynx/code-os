import { describe, expect, it } from "vitest"

import type { Screenshot } from "@/api/schema"
import { groupScreenshotsByFeature, screenshotFeature } from "./screenshots"

function screenshot(overrides: Partial<Screenshot>): Screenshot {
  return {
    id: "screenshot-id",
    name: "proof.png",
    url: "/media/screenshot-id",
    size: 1_024,
    createdAt: "2026-08-28T12:00:00Z",
    ...overrides,
  }
}

describe("screenshot features", () => {
  it("prefers the verification run over the top-level storage group", () => {
    expect(screenshotFeature(screenshot({ project: "billing-overage", group: "skills-verify" }))).toBe("billing-overage")
  })

  it("groups screenshots while preserving their recency order", () => {
    const screenshots = [
      screenshot({ id: "a", name: "F02.png", project: "billing-overage" }),
      screenshot({ id: "b", name: "F01.png", project: "billing-overage" }),
      screenshot({ id: "c", name: "home.png", project: "one-click-upload" }),
    ]

    expect(groupScreenshotsByFeature(screenshots)).toEqual([
      { name: "billing-overage", screenshots: screenshots.slice(0, 2) },
      { name: "one-click-upload", screenshots: screenshots.slice(2) },
    ])
  })
})
