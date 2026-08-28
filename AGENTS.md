# StackEnv

## Product contract

- StackEnv is a cross-platform command center for local and remote development environments.
- The daemon must listen on loopback by default. Public access is provided through an authenticated tunnel.
- Git inspection is read-only in the first milestone. Never discard, stage, commit, or push user changes.
- Portly remains the source of truth for persistent development processes.
- Screenshots are indexed in place and must not be made public implicitly.
- Cloudflare account, zone, tunnel, hostname, and credential paths are configuration, never constants.

## Development servers

- Always use Portly (`portly ...`) to start, stop, restart, inspect, or keep local development servers running.
- Start with `portly status` and reuse a healthy managed server.
- Never launch a persistent development server directly or through shell backgrounding.

## Checks

- Format with `gofmt -w`.
- Run `go test ./...`, `go vet ./...`, and `go build ./cmd/stackenv` before handoff.
