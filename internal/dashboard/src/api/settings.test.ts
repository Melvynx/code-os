import { afterEach, describe, expect, it, vi } from "vitest"

import { fetchTrustedIPStatus, revokeCurrentTrustedIP, trustCurrentIP } from "./settings"

describe("trusted IP settings API", () => {
  afterEach(() => vi.unstubAllGlobals())

  it("reads the current exact-IP trust state", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({
      currentIP: "203.0.113.7",
      trusted: true,
      configured: true,
      count: 1,
    }), { status: 200, headers: { "Content-Type": "application/json" } }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(fetchTrustedIPStatus()).resolves.toEqual({
      currentIP: "203.0.113.7",
      trusted: true,
      configured: true,
      count: 1,
    })
    expect(fetchMock).toHaveBeenCalledWith("/api/trusted-ip", expect.objectContaining({ credentials: "same-origin" }))
  })

  it("revokes the current IP with an authenticated same-origin DELETE", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({
      currentIP: "203.0.113.7",
      trusted: false,
      configured: true,
      count: 0,
    }), { status: 200, headers: { "Content-Type": "application/json" } }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(revokeCurrentTrustedIP()).resolves.toEqual(expect.objectContaining({ trusted: false }))
    expect(fetchMock).toHaveBeenCalledWith("/api/trusted-ip", expect.objectContaining({ method: "DELETE", credentials: "same-origin" }))
  })

  it("trusts the current IP with an authenticated same-origin POST", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({
      currentIP: "203.0.113.7",
      trusted: true,
      configured: true,
      count: 1,
    }), { status: 200, headers: { "Content-Type": "application/json" } }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(trustCurrentIP()).resolves.toEqual(expect.objectContaining({ trusted: true }))
    expect(fetchMock).toHaveBeenCalledWith("/api/trusted-ip", expect.objectContaining({ method: "POST", credentials: "same-origin" }))
  })
})
