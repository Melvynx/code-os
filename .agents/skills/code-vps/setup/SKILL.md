---
name: code-vps-setup
description: Walk the Code OS VPS setup: SSH key, origin HTML login, password file, media ?bp= bypass, and the direct Cloudflare token URL. Use when installing Code OS on a new VPS.
---

# Code OS VPS setup

Provision a Linux VPS so Code OS listens on loopback, Cloudflare Tunnel is TLS only, and origin auth is the custom login page plus a media bypass key.

Do not install Cloudflare Access, Zero Trust login, HTTP Basic Auth, or `WWW-Authenticate`. `cloudflare.requireAccess` in config is a leftover doctor flag. Ignore `code-os cloudflare` when it says “Protect … with Cloudflare Access”.

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

## 2. Bootstrap the VPS

Completion: `code-os version` and `cloudflared version` work; `portly status` talks to the local daemon.

```bash
install -d -m 700 \
  /root/projects \
  /root/.config/code-os \
  /root/.config/tokens \
  /root/.local/bin \
  /root/.local/share/code-os/files
```

Install `code-os` from the latest Linux amd64 release (or build from this repo). Put the binary on `PATH`, typically `~/.local/bin/code-os`.

Install `cloudflared` from Cloudflare's Linux package or GitHub release to `/usr/local/bin/cloudflared`.

Install the **headless** Portly CLI (`/usr/local/bin/portly` or `~/.local/bin/portly`). Enable the systemd user `portly.service` on `127.0.0.1:7737`. Do not install Portly.app.

Keep the Code OS daemon on `127.0.0.1:7890`. Never bind it to `0.0.0.0`.

## 3. Cloudflare API token, then tunnel

Completion: API token file is `0600` and `cloudflare-cli auth test` (or an equivalent `GET /accounts/<id>/cfd_tunnel`) succeeds. Tunnel status becomes healthy after step 5.

1. Hand the user the token URL from the top of this skill. Wait for the token value.
2. Store it without echoing:

```bash
install -d -m 700 ~/.config/tokens
# write the token, chmod 600, no cat, no tee to the terminal
```

3. Resolve `accountId` and `zoneId` from the API. Write them into setup flags; they are configuration, never constants in source.
4. Create a remotely-managed tunnel (`config_src=cloudflare`):

```bash
curl -sS "https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/cfd_tunnel" \
  -H "Authorization: Bearer $(tr -d '\n' < ~/.config/tokens/cloudflare-cli.txt)" \
  -H "Content-Type: application/json" \
  --data '{"name":"code-os","config_src":"cloudflare"}'
```

Copy `result.id` (tunnel id) and `result.token` (run token). Put the run token in `/etc/cloudflared/code-os.token` mode `0600`.

5. PUT ingress: dashboard hostname → `http://127.0.0.1:7890`, then catch-all `http_status:404`. Preserve unrelated ingress rules if the tunnel is shared.
6. Create proxied CNAMEs:

- `<dashboardHost>` → `<tunnelId>.cfargotunnel.com`
- later, each `portNNNN` host via the port-sync worker — never a wildcard

Do not create an Access application, an Access policy, or a Zero Trust login method.

## 4. `code-os setup`

Completion: `~/.config/code-os/config.json` exists, password / bypass / session files are `0600`, `code-os doctor` does not fail required checks.

```bash
code-os setup \
  --name <environment> \
  --type remote \
  --projects-root /root/projects \
  --screenshots-root ~/.local/share/code-os/files \
  --files-root ~/.local/share/code-os/files \
  --dashboard-host <dashboardHost> \
  --public-port-host 'port{port}.<zone>' \
  --dashboard-username code-os \
  --dashboard-password-file ~/.config/code-os/dashboard-password \
  --dashboard-bypass-key-file ~/.config/code-os/media-bypass-key \
  --dashboard-session-key-file ~/.config/code-os/session-key \
  --dashboard-trusted-ips-file ~/.config/code-os/trusted-ips \
  --cloudflare-tunnel-id <tunnelId> \
  --cloudflare-account-id <accountId> \
  --cloudflare-zone-id <zoneId> \
  --cloudflare-token-file ~/.config/tokens/cloudflare-cli.txt \
  --skills-repository <private-https-or-ssh-url> \
  --skills-directory ~/.agents
```

`setup` generates missing secret files. Tell the user the password file path and that they sign in on the **custom** `/login` page — not Cloudflare, not HTTP Basic.

## 5. systemd units

Completion: `cloudflared-code-os.service` is active; `systemctl --user is-active code-os.service` is active; linger is `yes`.

Tunnel connector (system unit, not user):

```ini
# /etc/systemd/system/cloudflared-code-os.service
[Unit]
Description=Cloudflare Tunnel for Code OS
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/cloudflared tunnel --no-autoupdate run --token-file /etc/cloudflared/code-os.token
Restart=always
RestartSec=5
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadOnlyPaths=/etc/cloudflared/code-os.token

[Install]
WantedBy=multi-user.target
```

Then:

```bash
systemctl enable --now cloudflared-code-os.service
code-os service install
code-os doctor
loginctl show-user "$USER" -p Linger
```

`service install` writes the user unit, enables linger, starts Code OS now, and installs the skills-sync timer when a repository is configured.

## 6. Verify the real auth path

Completion: all four checks pass.

1. `https://<dashboardHost>/login` is the Code OS HTML form (`Sign in to Code OS`, fields `username` + `current-password`). It is not a Cloudflare Access page.
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

Day-to-day operation: `.agents/skills/code-os/SKILL.md`.
Public port publishing: `remote-webview`.
