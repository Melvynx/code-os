# Code OS

Code OS turns a Mac or Linux host into one legible development environment. It discovers repositories, every Git worktree and subproject; reads Portly applications; groups screenshots by feature; synchronizes a private skills repository; and exposes the result through one secure service.

Project Git inspection is read-only. Portly remains the process supervisor. Code OS is the only supported transport for the dashboard, private files, screenshots, and tunneled development applications.

## Quick start

```bash
go build -o bin/code-os ./cmd/code-os
./bin/code-os setup \
  --projects-root /root/projects \
  --screenshots-root ~/.local/share/code-os/screenshots \
  --files-root ~/.local/share/code-os/files \
  --dashboard-host code-os.mlvcdn.com \
  --dashboard-username code-os \
  --dashboard-password-file ~/.config/code-os/dashboard-password \
  --dashboard-bypass-key-file ~/.config/code-os/media-bypass-key \
  --dashboard-session-key-file ~/.config/code-os/session-key \
  --dashboard-trusted-ips-file ~/.config/code-os/trusted-ips \
  --public-port-host 'port{port}.mlvcdn.com' \
  --skills-repository git@github.com:YOUR_ACCOUNT/agents-config.git \
  --skills-directory ~/.agents
./bin/code-os service install
```

Open `http://127.0.0.1:7890` for the public landing page and `/app/` for the authenticated command center. Configuration defaults to `~/.config/code-os/config.json`; runtime data defaults to `~/.local/share/code-os`.

On Linux, `service install` enables systemd user lingering, installs and starts Code OS immediately, and enables it for every reboot. Running the command again updates the unit and restarts the daemon on the current binary.

## Commands

```text
code-os setup       Write the per-machine configuration
code-os dashboard   Run the loopback command center and gateway
code-os scan        Print the current environment snapshot as JSON
code-os status      Show a compact environment summary
code-os doctor      Check configuration, secrets, Git, Portly, and tunnel inputs
code-os cloudflare  Print the managed tunnel ingress configuration
code-os service     Install the dashboard and skills-sync user services
code-os skills-sync Synchronize the configured private skills repository
code-os version     Print the build version
```

## Build and test

```bash
pnpm --dir website install --frozen-lockfile
pnpm --dir website build
pnpm --dir internal/dashboard install --frozen-lockfile
pnpm --dir internal/dashboard typecheck
pnpm --dir internal/dashboard test
pnpm --dir internal/dashboard build
go test ./...
go vet ./...
go build -o bin/code-os ./cmd/code-os
```

The React/Vite/TanStack Router dashboard uses shadcn/ui primitives and is embedded under `/app/`. The public TanStack Start landing and documentation are embedded at `/`. Both ship inside the Go binary.

## Security model

- The daemon binds to `127.0.0.1`; Cloudflare Tunnel is the TLS transport.
- The landing/docs are public. `/app`, `/api`, `/media`, `/files`, and every `portNNNN` gateway require Code OS origin authentication.
- Login uses a password-manager-compatible form, rate limiting, a stable 256-bit signing key, and `HttpOnly` secure cookies.
- After a valid password sign-in, the user can trust the detected exact public IP. Trusted IPs are stored locally in a mode-`0600` file, apply to the dashboard and protected ports, and can be revoked from Settings.
- There is no anonymous artifact host. Screenshots and verification files stay private and non-cacheable.
- A separate bypass key authorizes only `GET` and `HEAD` image reads under `/media/` and `/files/`. It never grants dashboard, settings, API, or application-gateway access.
- Cloudflare token values are write-only, stored in a `0600` file, and never returned by the settings API.
- Settings only accept absolute non-root paths and credential-free GitHub repository URLs.
- The gateway proxies only running, healthy ports reported by Portly. Local agents continue to use loopback directly.
- Process controls require an authenticated same-origin `POST`. Application IDs must match a currently running Portly server; agent IDs include both PID and kernel start time to prevent PID-reuse mistakes. Code OS refuses to terminate the process tree that hosts itself.

The Applications page graphs current CPU and resident memory for running Portly apps and detected Codex, Cursor, Claude, OpenCode, Aider, and Gemini agents. Agent child processes are grouped under their owning agent. “Kill” uses Portly for applications and `SIGTERM` for agents; both actions require confirmation.

Private image examples:

```text
https://<dashboard-host>/media/<screenshot-id>?bp=<url-encoded-key>
https://<dashboard-host>/files/<feature>/<evidence>.png?bp=<url-encoded-key>
```

Treat the complete URL as a bearer secret. Keep it out of source control, issue trackers, analytics, shell tracing, and broad logs; rotate the bypass key if it leaks.

## Skills synchronization

Set the GitHub repository, local checkout, and branch in `/app/settings`. Run `code-os skills-sync`, or install the two-minute timer with `code-os service install`. Only the skills repository is mutated; project repositories remain read-only. Keep credentials, tokens, machine config, and screenshots outside the skills repository.

## Architecture

```text
Git/worktrees ─┐
Portly JSON ───┼──▶ Code OS (127.0.0.1:7890) ──▶ public landing/docs
Screenshots ───┤                  ├──────────────▶ authenticated app/API/files
Private files ─┘                  └──────────────▶ authenticated port gateway
```
