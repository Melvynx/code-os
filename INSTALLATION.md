# Installation

Start here. This is the only map an agent needs to provision or operate Code OS. Do not invent a second Cloudflare, SSH, or Portly flow.

## Which skill to load

| Job | Invoke | File | When |
| --- | --- | --- | --- |
| **New VPS** | `$code-vps-setup` | `.agents/skills/code-vps-setup/SKILL.md` | SSH key, Hostinger, Cloudflare API token, first `code-os setup`. **Only** provisioning skill. |
| **Operate a running machine** | `$code-os` | installed to `~/.agents/skills/code-os` after setup | doctor, status, login, `?bp=` media. Never use it to provision. |
| **Dev servers** | `$portly` | installed to `~/.agents/skills/portly` | start / stop / add-server. Headless CLI, never Portly.app. |
| **Public app URL** | `$remote-webview` | installed to `~/.agents/skills/remote-webview` | `port{port}.<zone>` CNAMEs. Read `publicPortHost` from config. |
| **Screenshots** | `$verify` | installed to `~/.agents/skills/verify` | write under `filesRoot/skills-verify/`. |

`$code-vps-setup` lives in this repository. `setup` does **not** copy it onto the VPS (a running machine must not re-provision itself). After setup, the four operate skills above are installed into `~/.agents`, `~/.claude/skills`, and `~/.cursor/skills`, plus managed rules in `AGENTS.md` and `~/.cursor/rules/{portly,code-os}.mdc`.

## Order (do not skip)

1. Read this file, then load **only** `$code-vps-setup`.
2. Collect: VPS IPv4, SSH user (Hostinger is usually `root`), dashboard hostname, zone already on Cloudflare nameservers.
3. SSH key + `Host` alias. `User` must be explicit.
4. On the VPS: install `code-os` from `releases/latest` (asset `code-os-linux-amd64`), and `cloudflared` at `/usr/local/bin/cloudflared`. Portly is installed by `setup`.

```bash
curl -fsSL https://github.com/Melvynx/code-os/releases/latest/download/code-os-linux-amd64 -o ~/.local/bin/code-os
chmod +x ~/.local/bin/code-os
code-os version   # 0.1.5 or newer
```
5. Cloudflare **API token** (the prefilled URL in `$code-vps-setup`). Store at `~/.config/tokens/cloudflare-cli.txt` mode `0600`. Resolve `accountId` and `zoneId`. Never print the token.
6. `code-os setup --non-interactive` with the flags in `$code-vps-setup`. **Omit `--cloudflare-tunnel-id`** unless `/etc/cloudflared/code-os.token` already exists. Setup creates the tunnel, dashboard CNAME, ingress (dashboard host + `404` only), Portly, screenshot/files dirs at `0700`, and the agent skills.
7. `code-os service install && code-os doctor`.
8. Prove `/login` is the Code OS HTML form. Switch to `$code-os` for day-to-day work.

## What `setup` does vs what you still do

`setup` does:

- screenshot / files / data dirs at `0700`
- dashboard password, bypass key, session key at `0600` if the files are missing
- Portly CLI to `/usr/local/bin/portly` or `~/.local/bin/portly` (never `.cleanup-*`), then `portly forever enable`
- persistent `code-os` binary if the current one is ephemeral
- `publicPortHost` = `port{port}.<zone>` from the dashboard host when the flag is omitted. **Never `*`.**
- Cloudflare tunnel when token + account + zone are set (failure is fatal)
- `add-project` for Git repos already under `projectsRoot` — **no `add-server`, no start**
- Portly / remote-webview / verify / code-os skills

You still do:

- install `cloudflared` before setup
- clone the user's repos (do not hardcode product names)
- `portly add-server` from each repo's real command / port / health check
- copy env files from a private store, mode `0600`, never print values
- `NODE_OPTIONS=--max-old-space-size=6144` for heavy typechecks
- `pnpm approve-builds` when install ignores `esbuild` / `sharp` / `@parcel/watcher`

## Guardrails

- Loopback only (`127.0.0.1:7890`). No `0.0.0.0`.
- Custom `/login`, not Cloudflare Access, not HTTP Basic, no `WWW-Authenticate`.
- `?bp=` is media-only. Same key must `401` on `/api/health`.
- Individual `portNNNN.<zone>` CNAMEs. No wildcard DNS.
- Hosts are config, never constants. Read `~/.config/code-os/config.json`.
- Project Git is read-only. Only the configured skills repo is mutated.

## Done when

`code-os doctor` has no required failures (a dirty or conflicted `~/.agents` checkout fails `skills-sync` until you resolve it — do not force-push), `portly status` talks to the daemon, `http://127.0.0.1:7890/login` is the Code OS form, and `/docs` is public. A `530` from the VPS to its own public hostname is often Cloudflare hairpin — confirm `/login` from an external client. After building from this repo, run `code-os service install` so the running binary matches.
