import { z } from "zod"

export const gitStateSchema = z.object({
  branch: z.string(),
  upstream: z.string().optional(),
  ahead: z.number().int().nonnegative(),
  behind: z.number().int().nonnegative(),
  modified: z.number().int().nonnegative(),
  added: z.number().int().nonnegative(),
  deleted: z.number().int().nonnegative(),
  untracked: z.number().int().nonnegative(),
  conflicts: z.number().int().nonnegative(),
})

const subprojectSchema = z.object({
  name: z.string(),
  path: z.string(),
  kind: z.string(),
})

const worktreeSchema = z.object({
  id: z.string(),
  path: z.string(),
  head: z.string().optional(),
  main: z.boolean(),
  locked: z.boolean(),
  prunable: z.boolean(),
  git: gitStateSchema,
})

const projectSchema = z.object({
  id: z.string(),
  name: z.string(),
  path: z.string(),
  git: gitStateSchema,
  worktrees: z.array(worktreeSchema).nullish().transform((items) => items ?? []),
  subprojects: z.array(subprojectSchema).nullish().transform((items) => items ?? []),
})

const applicationSchema = z.object({
  id: z.string(),
  projectId: z.string(),
  projectName: z.string(),
  name: z.string(),
  command: z.string(),
  directory: z.string().optional(),
  state: z.string(),
  port: z.number().int().nonnegative(),
  pid: z.number().int().optional(),
  healthy: z.boolean().optional(),
  url: z.string().optional(),
  publicUrl: z.string().optional(),
  cpuPercent: z.number(),
  memoryBytes: z.number().nonnegative(),
  residentMemoryBytes: z.number().nonnegative(),
  restartCount: z.number().int().nonnegative(),
})

const agentProcessSchema = z.object({
  id: z.string(),
  name: z.string(),
  command: z.string(),
  pid: z.number().int().positive(),
  cpuPercent: z.number().nonnegative(),
  memoryBytes: z.number().nonnegative(),
  processCount: z.number().int().positive(),
})

const screenshotSchema = z.object({
  id: z.string(),
  name: z.string(),
  url: z.string(),
  project: z.string().optional(),
  group: z.string().optional(),
  size: z.number().nonnegative(),
  createdAt: z.string(),
})

export const snapshotSchema = z.object({
  generatedAt: z.string(),
  projects: z.array(projectSchema).nullish().transform((items) => items ?? []),
  applications: z.array(applicationSchema).nullish().transform((items) => items ?? []),
  agents: z.array(agentProcessSchema).nullish().transform((items) => items ?? []),
  screenshots: z.array(screenshotSchema).nullish().transform((items) => items ?? []),
  warnings: z.array(z.string()).nullish().transform((items) => items ?? []),
})

export type Snapshot = z.infer<typeof snapshotSchema>
export type Project = z.infer<typeof projectSchema>
export type GitState = z.infer<typeof gitStateSchema>
export type Worktree = z.infer<typeof worktreeSchema>
export type Application = z.infer<typeof applicationSchema>
export type AgentProcess = z.infer<typeof agentProcessSchema>
export type Screenshot = z.infer<typeof screenshotSchema>
