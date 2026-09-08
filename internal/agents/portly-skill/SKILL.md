---
name: portly
description: Supervise Vite, Next, and pnpm/dev servers with the Portly CLI. Use when starting, stopping, restarting, or previewing a local app. On Linux this is the headless daemon — never Portly.app, never pnpm/vite in the background.
---

<!-- code-os:portly-skill -->

# Portly

Use the `portly` CLI. It is the supervisor for every local development process.

On Linux (Code OS VPS), this is the **headless** binary — typically `/usr/local/bin/portly` or `~/.local/bin/portly`. The daemon is systemd user `portly.service` on `127.0.0.1:7737`. Do not install or launch Portly.app. `open` has no window.

Never start Vite, Next, `pnpm dev`, `npm run dev`, or a preview with the shell, `nohup`, `&`, or another supervisor.

## Inspect first

Run `portly status` for the compact active/problem view. Use `--details` only for the full inventory, `--json` only for machine-readable fields. Prefer an exact `project/server` when names collide.

Reuse a healthy managed server. If the in-scope process is already listening outside Portly, register it and run `portly take-over <project/server> --json`.

## Choose project or temporary

- **Long-lived or reusable:** `portly add-project` then `portly add-server --start`.
- **One-off** (build, test, preview, screenshot proof): `job_id="$(portly temp '<command>' --path <folder> --timeout 30m)"` then `portly wait "$job_id"`. Timeout kills the process group and exits `124`.

Do not create a permanent project just to host a throwaway preview.

## Keep the project rules durable

When setting up Portly for a repository, add this rule to the root `AGENTS.md` if an equivalent is missing. Preserve existing instructions.

```md
## Development servers

- Always use Portly (`portly ...`) to start, stop, restart, inspect, or keep local development servers running.
- Start with `portly status`. Use `portly status --details` only for the full inventory and metrics, and `--json` only for machine-readable fields. Reuse a healthy managed server; if an in-scope server is running outside Portly, register it and use `portly take-over <project/server> --json`.
- For long-lived or reusable work, create a project and server.
- For builds, tests, code generation, previews, demos, and other bounded one-off work, run `job_id="$(portly temp '<command>' --path <folder> --timeout 30m)"`, then `portly wait "$job_id"`. `temp` returns immediately with an ID; `wait` prints captured logs and exits with the command's real code. A timeout kills the whole process group and exits with code `124`.
- Never launch persistent development servers directly, in the background, or through another supervisor.
```

Code OS installs the same rule, between `portly:managed-rule` markers, in `~/.agents/AGENTS.md` and Cursor `~/.cursor/rules/portly.mdc`.

## Add and verify a server

1. Confirm the project path and dev command.
2. Ensure the repository `AGENTS.md` has the rule above.
3. `portly add-project --name <name> --path <absolute-path> --json` if absent.
4. `portly port <port> --json`.
5. `portly add-server --project <project> --name <name> --command '<command>' --port <port> --start --json`.
6. Poll `portly status --json` until `running` and `healthy`.
7. Hit the meaningful URL. Read `portly logs <project/server> --tail 100 --json`.

Portly injects `PORT`, `PORTLY=1`, and `PORTLY_SERVER`. Health requires a listener on the configured port.

## Operate

`start`, `stop`, `restart` take a server or `--project`. `portly stop --all --json` stops everything. `portly take-over <project/server> --json` adopts an external listener. `portly forever enable` is the systemd user unit plus linger — not a LaunchAgent.

Config: `~/.config/portly/config.json`. Logs: `~/.config/portly/logs/`.
