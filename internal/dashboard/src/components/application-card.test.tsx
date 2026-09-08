import { render, screen } from "@testing-library/react"
import { describe, expect, it, vi } from "vitest"

import { ApplicationCard } from "./application-card"

describe("ApplicationCard", () => {
  it("renders a scannable product card with domain and public host", () => {
    render(
      <ApplicationCard
        application={{
          id: "srv_start",
          projectId: "prj_lumail",
          projectName: "lumail.io",
          name: "start",
          command: "pnpm dev",
          state: "running",
          port: 3013,
          pid: 99,
          healthy: true,
          url: "http://localhost:3013",
          publicUrl: "https://port3013.mlvcdn.com/",
          domain: "lumail.io",
          cpuPercent: 0.8,
          memoryBytes: 3_100_000_000,
          residentMemoryBytes: 3_100_000_000,
          restartCount: 0,
        }}
        onStop={vi.fn()}
      />,
    )

    expect(screen.getByRole("heading", { name: "Lumail" })).toBeInTheDocument()
    expect(screen.getByText("lumail.io")).toBeInTheDocument()
    expect(screen.getByText(/Start · :3013 · port3013.mlvcdn.com/)).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Open" })).toHaveAttribute("href", "https://port3013.mlvcdn.com/")
    expect(document.querySelector("img")).toHaveAttribute("src", "/api/applications/srv_start/icon")
  })
})
