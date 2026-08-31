import { fireEvent, render, screen } from "@testing-library/react"
import { describe, expect, it, vi } from "vitest"

import { ResourceUsage } from "./resource-usage"

describe("ResourceUsage", () => {
  it("graphs grouped apps and agents and exposes guarded kill controls", () => {
    render(
      <ResourceUsage
        applications={[{
          id: "srv_demo",
          projectId: "project-demo",
          projectName: "Demo",
          name: "web",
          command: "pnpm dev",
          state: "running",
          port: 3000,
          pid: 120,
          healthy: true,
          cpuPercent: 12.5,
          memoryBytes: 256_000_000,
          residentMemoryBytes: 240_000_000,
          restartCount: 0,
        }]}
        agents={[{
          id: "42:9001",
          name: "Codex agent",
          command: "codex",
          pid: 42,
          cpuPercent: 28,
          memoryBytes: 800_000_000,
          processCount: 4,
        }]}
        onStopApplication={vi.fn()}
        onTerminateAgent={vi.fn()}
      />,
    )

    expect(screen.getByText("Codex agent")).toBeInTheDocument()
    expect(screen.getByText("Demo / web")).toBeInTheDocument()
    expect(screen.getByRole("img", { name: "CPU: 28.0%" })).toBeInTheDocument()
    const killButtons = screen.getAllByRole("button", { name: "Kill" })
    expect(killButtons).toHaveLength(2)

    fireEvent.click(killButtons[1]!)
    expect(screen.getByText("Kill Demo / web?")).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Confirm kill" })).toBeInTheDocument()
  })
})
