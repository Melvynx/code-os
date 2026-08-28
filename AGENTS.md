# Code OS

## Product contract

- Code OS is a cross-platform command center for local and remote development environments.
- The daemon must listen on loopback by default. Public access is provided through an authenticated tunnel.
- Git inspection is read-only in the first milestone. Never discard, stage, commit, or push user changes.
- Portly remains the source of truth for persistent development processes.
- Screenshots are indexed in place and must not be made public implicitly.
- Dashboard authentication uses the HTML login/session flow. Do not restore HTTP Basic Auth or emit `WWW-Authenticate`.
- The bypass key authorizes only `GET` and `HEAD` under `/media/`; `/api/` and the dashboard must reject it. Treat complete `?bp=` URLs as bearer secrets.
- Cloudflare account, zone, tunnel, hostname, and credential paths are configuration, never constants.

## Development servers

- Always use Portly (`portly ...`) to start, stop, restart, inspect, or keep local development servers running.
- Start with `portly status` and reuse a healthy managed server.
- Never launch a persistent development server directly or through shell backgrounding.

## Checks

- Build the embedded dashboard first with `pnpm --dir internal/dashboard build`.
- Run `pnpm --dir internal/dashboard typecheck` and `pnpm --dir internal/dashboard test`.
- Format with `gofmt -w`.
- Run `go test ./...`, `go vet ./...`, and `go build ./cmd/code-os` before handoff.
