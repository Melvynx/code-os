import { describe, expect, it } from "vitest"

import type { Application } from "@/api/schema"

import {
  applicationDisplayName,
  applicationDomain,
  applicationHostLabel,
  applicationMatchesQuery,
  applicationTitle,
  featuredApplications,
  isFeaturedApplication,
  isPrimaryApplication,
  isSupportingApplication,
  sortApplications,
} from "./applications"

function app(overrides: Partial<Application>): Application {
  return {
    id: "srv",
    projectId: "prj",
    projectName: "demo",
    name: "web",
    command: "pnpm dev",
    state: "running",
    port: 3000,
    cpuPercent: 0,
    memoryBytes: 0,
    residentMemoryBytes: 0,
    restartCount: 0,
    ...overrides,
  }
}

describe("application identity", () => {
  it("turns project names into scannable titles and domains", () => {
    expect(applicationTitle({ projectName: "lumail.io" })).toBe("Lumail")
    expect(applicationTitle({ projectName: "ai-builder-club" })).toBe("AI Builder Club")
    expect(applicationDisplayName({ projectName: "lumail.io", name: "start" })).toBe("Lumail Start")
    expect(applicationDomain({ domain: "kliq.sh", projectName: "kliq" })).toBe("kliq.sh")
    expect(applicationHostLabel({ port: 3013, publicUrl: "https://port3013.mlvcdn.com/", url: "http://localhost:3013" })).toBe("port3013.mlvcdn.com")
  })

  it("treats web and start servers with ports as primary apps", () => {
    expect(isPrimaryApplication(app({ name: "start", projectName: "lumail.io", port: 3013 }))).toBe(true)
    expect(isFeaturedApplication(app({ name: "web", port: 3100 }))).toBe(true)
    expect(isSupportingApplication(app({ name: "convex", port: 0 }))).toBe(true)
    expect(isFeaturedApplication(app({ name: "sync", port: 7742 }))).toBe(false)
    expect(isFeaturedApplication(app({ name: "static", port: 4570 }))).toBe(false)
  })

  it("sorts running linked apps ahead of workers and stopped inventory", () => {
    const ordered = sortApplications([
      app({ id: "stopped", name: "dev", state: "stopped", port: 3003, projectName: "lumail-k9ko" }),
      app({ id: "convex", name: "convex", port: 0, projectName: "ai-builder-club", residentMemoryBytes: 600 }),
      app({ id: "sync", name: "sync", port: 7742, projectName: "dokploy-github-status", residentMemoryBytes: 20 }),
      app({ id: "web", name: "web", port: 3100, projectName: "ai-builder-club", residentMemoryBytes: 100 }),
      app({ id: "start", name: "start", port: 3013, projectName: "lumail.io", residentMemoryBytes: 300 }),
    ])

    expect(ordered.map((item) => item.id)).toEqual(["start", "web", "sync", "convex", "stopped"])
    expect(featuredApplications(ordered).map((item) => item.id)).toEqual(["start", "web"])
  })

  it("matches visible titles, domains, and multi-word queries as you type", () => {
    const lumail = app({ projectName: "lumail.io", name: "start", domain: "lumail.io", publicUrl: "https://port3013.mlvcdn.com/", port: 3013 })
    const club = app({ projectName: "ai-builder-club", name: "web", domain: "ai-builder.club" })
    expect(applicationMatchesQuery(lumail, "Lumail")).toBe(true)
    expect(applicationMatchesQuery(lumail, "lumail start")).toBe(true)
    expect(applicationMatchesQuery(lumail, "port3013")).toBe(true)
    expect(applicationMatchesQuery(club, "ai builder")).toBe(true)
    expect(applicationMatchesQuery(lumail, "kliq")).toBe(false)
  })
})
