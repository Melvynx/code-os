---
name: code-vps-setup
description: code-vps-setup provisions a new Code OS VPS. Use $code-vps-setup for Hostinger, SSH keys, Cloudflare token URL, tunnel, first install. Never load the code-os skill to provision.
---

# $code-vps-setup

This is the only provisioning skill. Invoke it as `$code-vps-setup`. Read `INSTALLATION.md` first for the skill map. Do not load `code-os`, do not SSH-find other SKILL.md files, do not invent a second Cloudflare token flow.

Provision a Linux VPS so Code OS listens on loopback, Cloudflare Tunnel is TLS only, and origin auth is the custom login page plus a media bypass key.

Do not install Cloudflare Access, Zero Trust login, HTTP Basic Auth, or `WWW-Authenticate`. `cloudflare.requireAccess` in config is a leftover doctor flag. `code-os cloudflare` must show the dashboard host, `publicPortHost`, and `"wildcard": false`.

## Auth model (this is the product)

| Surface | How it authenticates |
| --- | --- |
| `https://<dashboardHost>/login` | Custom HTML form. Username + password from `auth.passwordFile`. Session cookie, 12h, HttpOnly. |
| `https://portNNNN.<zone>/_code-os/login` | Same password, separate gateway cookie. |
| Settings → Trust this IP | After a valid password sign-in, skip the form from that exact IP. |
| `?bp=<key>` | Media only. `GET`/`HEAD` under `/media/` and `/files/`. `401` on `/api/` and `/app/`. |

`/` redirects to `/app/`. `/docs` stays public. Treat a complete `?bp=` URL as a bearer secret.

## Two Cloudflare secrets (do not mix them)

1. **API token** — dashboard-created, stored at `cloudflare.tokenFile` (usually `~/.config/tokens/cloudflare-cli.txt`). Manages DNS + remotely-managed tunnel ingress.
2. **Tunnel run token** — `eyJ…` from tunnel create. Stored at `/etc/cloudflared/code-os.token`. Only `cloudflared` reads it.

Give the user this exact API-token URL. Do not say “go create a token” without it:

```text
https://dash.cloudflare.com/profile/api-tokens?permissionGroupKeys=%5B%7B%22key%22%3A%22account_settings%22%2C%22type%22%3A%22read%22%7D%2C%7B%22key%22%3A%22argotunnel%22%2C%22type%22%3A%22edit%22%7D%2C%7B%22key%22%3A%22dns%22%2C%22type%22%3A%22edit%22%7D%2C%7B%22key%22%3A%22zone%22%2C%22type%22%3A%22read%22%7D%5D&accountId=*&zoneId=all&name=Code%20OS
```

That prefills four rows. The dashboard slug for Cloudflare Tunnel is `argotunnel`. `cloudflare_tunnel` is silently dropped — treat a 3-row form as a failed URL, not a complete token.

Verify the form before the user clicks Create Token:

1. Account → Account Settings → Read
2. Account → Cloudflare Tunnel → Edit
3. Zone → DNS → Edit
4. Zone → Zone → Read

If row 2 is missing, add it with **+ Add more**. Scope one account and the target zone if the form allows. The user pastes the token once; never print it again.

## Inputs to collect before acting

- VPS IPv4
- SSH user (Hostinger images are usually `root`, not the local macOS username)
- Dashboard hostname, e.g. `code-os.example.com`
- Public port template, e.g. `port{port}.example.com`
- Zone that already uses Cloudflare nameservers

## 1. SSH key and Host alias (local machine)

Completion: `ssh <alias>` logs in with the key and does not ask for a password.

1. Create a dedicated ed25519 key. Do not reuse `id_ed25519` unless the user asks.

```bash
ssh-keygen -t ed25519 -f ~/.ssh/<alias> -C "<alias>" -N ""
```

2. Print the public key immediately and copy it to the clipboard. The user pastes it into the provider panel (Hostinger → SSH keys) or into `~<user>/.ssh/authorized_keys` on the server.

```bash
cat ~/.ssh/<alias>.pub
# macOS: pbcopy < ~/.ssh/<alias>.pub
```

3. Write `~/.ssh/config`. Set `User` explicitly. Hostinger will keep asking for a password if `User` falls back to the local account (`melvynx`, etc.).

```sshconfig
Host <alias>
  HostName <ipv4>
  User root
  IdentityFile ~/.ssh/<alias>
  IdentitiesOnly yes
```

4. Do not reboot the VPS after the key is added. `sshd` rereads `authorized_keys` on the next connection.
5. Connect with `ssh <alias>`. If it still asks for a password: wrong `User`, or the public key is on a different account than `User`.

## 2. Binaries on the VPS

Completion: `code-os version` and `cloudflared version` work. Do **not** install Portly by hand — `setup` does that.

```bash
install -d -m 700 \
  /root/projects \
  /root/.config/code-os \
  /root/.config/tokens \
  /root/.local/bin \
  /root/.local/share/code-os/screenshots \
  /root/.local/share/code-os/files
```

Install `code-os` from the latest Linux amd64 release (or build from this repo). Put it on `PATH` at `~/.local/bin/code-os` or `/usr/local/bin/code-os`. Never leave it under a `.cleanup-*` path.

Install `cloudflared` from Cloudflare's Linux package or GitHub release to `/usr/local/bin/cloudflared`. Setup fails if this binary is missing.

Keep the Code OS daemon on `127.0.0.1:7890`. Never bind it to `0.0.0.0`. Do not install Portly.app.

## 3. Cloudflare API token only

Completion: API token file is `0600` and `GET /accounts/<id>/cfd_tunnel` succeeds. **Do not create the tunnel, ingress, DNS, or cloudflared unit by hand.** `code-os setup` provisions them. Creating the tunnel twice leaves you with an id and no run token.

1. Hand the user the token URL from the top of this skill. Wait for the token value.
2. Store it without echoing:

```bash
install -d -m 700 ~/.config/tokens
# write the token, chmod 600, no cat, no tee to the terminal
```

3. Resolve `accountId` and `zoneId` from the API. Pass them as setup flags. They are configuration, never constants in source.

Do not create an Access application, an Access policy, or a Zero Trust login method.

## 4. `code-os setup`

Completion: `~/.config/code-os/config.json` exists, password / bypass / session files are `0600`, Cloudflare provision printed a tunnel id, `code-os doctor` does not fail required checks.

Use `--non-interactive`. **Omit `--cloudflare-tunnel-id`** on a new machine. Only pass it when `/etc/cloudflared/code-os.token` already exists.

```bash
code-os setup --non-interactive \
  --name <environment> \
  --type remote \
  --projects-root /root/projects \
  --screenshots-root ~/.local/share/code-os/screenshots \
  --files-root ~/.local/share/code-os/files \
  --dashboard-host <dashboardHost> \
  --public-port-host 'port{port}.<zone>' \
  --dashboard-username code-os \
  --dashboard-password-file ~/.config/code-os/dashboard-password \
  --dashboard-bypass-key-file ~/.config/code-os/media-bypass-key \
  --dashboard-session-key-file ~/.config/code-os/session-key \
  --dashboard-trusted-ips-file ~/.config/code-os/trusted-ips \
  --cloudflare-account-id <accountId> \
  --cloudflare-zone-id <zoneId> \
  --cloudflare-token-file ~/.config/tokens/cloudflare-cli.txt \
  --skills-repository <private-https-or-ssh-url> \
  --skills-directory ~/.agents
```

If Cloudflare flags are missing, setup now warns and leaves the machine private. If they are present and provision fails, setup exits non-zero — fix that before `service install`.

`setup` then:

- creates `screenshotsRoot` / `filesRoot` / `dataDir` at mode `0700` (keep screenshots and files as **separate** directories)
- generates missing dashboard secrets at `0600`
- installs the headless Portly CLI to `/usr/local/bin/portly` or `~/.local/bin/portly`, then `portly forever enable`
- copies `code-os` itself to a persistent bin if the current binary is ephemeral
- derives `publicPortHost` as `port{port}.<zone>` when the flag is omitted. **Never a wildcard.**
- creates the remotely-managed tunnel, dashboard CNAME, ingress (dashboard host + `http_status:404`), and `cloudflared-code-os.service`
- registers discovered Git repos as Portly projects **without starting servers**
- installs Portly / remote-webview / verify / code-os agent skills

Then:

```bash
code-os service install
code-os doctor
loginctl show-user "$USER" -p Linger
```

`service install` points the user unit at the persistent binary, writes `code-os-port-sync.timer` for individual `portNNNN.<zone>` CNAMEs, and installs the skills-sync timer. Tell the user the **paths** of the password and bypass files (not values) and that they sign in on the **custom** `/login` page.

## 4b. Bootstrap projects, envs, and Node

Do **not** hardcode product names. After `setup`, clone the user's repos under `projectsRoot`, then register each persistent app with Portly (`add-project` + `add-server`). Read the dev command and port from the repo; do not invent them. Do not start every Vite app at once.

Env files stay off Git. Copy them from a private store the user names (1Password, another machine, `scp`). Never print values. Typical destinations:

```text
<repo>/.env
<repo>/.env.local
<repo>/.env.prod
```

Mode `0600`. Confirm the files exist and are non-empty; do not `cat` them.

Heavy TypeScript checks on a small VPS:

```bash
NODE_OPTIONS=--max-old-space-size=6144 pnpm typecheck
```

If `pnpm install` reports `Ignored build scripts`, run `pnpm approve-builds` (or the repo's documented allow-list) before treating the install as complete. `esbuild`, `sharp`, and `@parcel/watcher` usually must run.

## 5. Units are up

Completion: all three are true. Do not write the cloudflared unit by hand unless setup could not (need root).

```bash
systemctl is-active cloudflared-code-os.service
systemctl --user is-active code-os.service
loginctl show-user "$USER" -p Linger
code-os cloudflare   # dashboard host + publicPortHost + wildcard: false
```

## 6. Verify the real auth path

Completion: all four checks pass.

1. `https://<dashboardHost>/login` is the Code OS HTML form (`Sign in to Code OS`, fields `username` + `current-password`). It is not a Cloudflare Access page. On the VPS itself, prove `http://127.0.0.1:7890/login` first. A `530` from the box to its own public hostname is often Cloudflare hairpin — confirm from an external client.
2. Sign in with `auth.username` and the contents of `auth.passwordFile`. Never paste those values into git, issues, or chat logs.
3. `/app/` loads. `/docs` is public without a session.
4. Media bypass works and stays narrow:

```text
https://<dashboardHost>/media/<id>?bp=<url-encoded-key>
https://<dashboardHost>/files/<path>?bp=<url-encoded-key>
```

`GET` image → 200, `Cache-Control: private, no-store`. Same `bp` on `/api/health` → 401.

Optional: Settings → Trust this IP for the current exact public IP.

## Guardrails

- Loopback only for Code OS and Portly app ports. Public ingress is the tunnel.
- No wildcard DNS. Individual `portNNNN` CNAMEs only, for healthy persistent Portly servers.
- Route every managed public hostname to Code OS `127.0.0.1:7890`. Code OS checks Portly health, requires the gateway session, then proxies with `Host: localhost:<port>`.
- Project Git stays read-only. Only the configured skills repository is mutated.
- Never dump token, password, bypass, or session files.

## Handoff

Return: SSH alias, dashboard URL, login username, **paths** to the password and bypass files (not values), tunnel id, and `code-os doctor` summary.

Day-to-day operation after this machine is up: `code-os` skill.
Public port publishing: `remote-webview`.
