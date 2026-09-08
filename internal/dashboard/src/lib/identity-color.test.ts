import { describe, expect, it } from "vitest"

import { identityColor, uniqueIdentityColor } from "./identity-color"

describe("identity colors", () => {
  it("stays stable for the same key", () => {
    expect(identityColor("lumail.io / start")).toBe(identityColor("lumail.io / start"))
  })

  it("splits hash collisions in a featured set", () => {
    const peers = ["ai-builder-club / web", "codeline / start", "kliq / start-all", "lumail.io / start"]
    const colors = peers.map((key) => uniqueIdentityColor(key, peers))
    expect(new Set(colors).size).toBe(peers.length)
  })

  it("gives each series a different color in a crowded list", () => {
    const peers = [
      "lumail.io / start",
      "ai-builder-club / web",
      "ai-builder-club / convex",
      "dokploy-github-status / sync",
      "steve-artifacts / static",
      "kliq / start-all",
      "codeline / start",
      "Codex agent",
      "Cursor agent",
      "Codex proxy",
    ]
    const colors = peers.map((key) => uniqueIdentityColor(key, peers))
    expect(new Set(colors).size).toBe(peers.length)
  })
})
