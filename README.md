# StackEnv

StackEnv turns a Mac or Linux host into an observable development environment. It discovers repositories and subprojects, reads Git work in progress, shows Portly-managed applications, and indexes screenshots in a loopback-only command center.

The first milestone is intentionally read-only for repositories. Portly remains the process supervisor, and Cloudflare Tunnel is the supported public transport for the dashboard.

## Quick start

```bash
go build -o bin/stackenv ./cmd/stackenv
./bin/stackenv setup \
  --projects-root /root/projects \
  --screenshots-root /var/www/code-vps \
  --dashboard-host stackenv.melvynx.dev \
  --dashboard-username stackenv \
  --dashboard-password-file ~/.config/stackenv/dashboard-password
./bin/stackenv dashboard
```

Open `http://127.0.0.1:7890`. Use `stackenv doctor` to inspect dependencies and the generated Cloudflare instructions.

## Commands

```text
stackenv setup       Write the per-machine configuration
stackenv dashboard   Run the loopback command center
stackenv scan        Print the current environment snapshot as JSON
stackenv status      Show a compact environment summary
stackenv doctor      Check configuration, Git, Portly, storage, and tunnel inputs
stackenv version     Print the build version
```

Configuration defaults to `~/.config/stackenv/config.json`. Runtime data defaults to `~/.local/share/stackenv/stackenv.db`.

## Security model

- The daemon binds to `127.0.0.1` unless explicitly configured otherwise.
- The dashboard hostname is distinct from public artifact hosting.
- Screenshots stay private to the dashboard unless another tool explicitly publishes them.
- Cloudflare credentials are referenced by path and are never copied into StackEnv configuration.
- Cloudflare Access is required by the generated deployment checklist before public use.
- HTTP basic authentication can protect the origin when Cloudflare Access administration is unavailable.

## Current boundary

This milestone does not yet push or pull the shared skills repository and does not mutate Cloudflare or Git state. Those operations will be added as separate, auditable fixed points after the command center foundation is verified.

## Architecture

```text
Git repositories ─┐
Portly JSON ──────┼──▶ stackenvd ──▶ loopback dashboard
Screenshots ──────┘       │
                          └──▶ SQLite snapshot index
```

Use `stackenv cloudflare` to print the managed ingress rule. Shared-tunnel output intentionally omits a fallback so unrelated ingress rules cannot be overwritten. A dedicated tunnel configuration includes its own final `http_status:404` fallback.
