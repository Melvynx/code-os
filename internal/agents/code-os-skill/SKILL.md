---
name: code-os
description: Operate an already-running Code OS dashboard, Portly apps, Git, screenshots, and media ?bp= URLs. Read hosts from ~/.config/code-os/config.json. Do not use for VPS provisioning, SSH keys, Hostinger, or Cloudflare token creation.
---

<!-- code-os:operate-skill -->

# Code OS

Use the local Code OS CLI and configuration as the source of truth. Hosts, the port template, and the files root come from `~/.config/code-os/config.json`. Never hardcode a dashboard or CDN domain.

Provisioning a new VPS is `$code-vps-setup` only. Never load this skill for SSH keys, Hostinger, or Cloudflare tokens.

## Inspect

1. Run `code-os doctor`.
2. Read `~/.config/code-os/config.json` for the loopback address, `cloudflare.dashboardHost`, `publicPortHost`, `filesRoot`, and credential file paths.
3. Run `code-os status` or `code-os scan`.
4. Inspect `systemctl --user status code-os.service` on Linux.
5. Open `/app/skills-sync` and `/app/status`.

Keep Git inspection read-only. Use Portly for persistent development processes.

## Authenticate

Open `https://<dashboardHost>/login`. Use `auth.username` and the password at `auth.passwordFile`. Never print the password file.

## Screenshots

Store durable verification images under `filesRoot/skills-verify/<run-id>/`.

- Codex: `![F01 - description](/absolute/filesRoot/skills-verify/<run-id>/F01.png)`
- Cursor: `![F01 - description](https://<dashboardHost>/files/skills-verify/<run-id>/F01.png?bp=<url-encoded-key>)`

`?bp=` is media-only (`GET`/`HEAD` under `/media/` and `/files/`). Same key on `/api/health` must be `401`. Treat the complete URL as a bearer secret.

## Public app ports

Replace `{port}` in `publicPortHost`. Example: `port{port}.example.com` → `https://port3013.example.com/`. Use `$remote-webview` to publish. Never invent a host.
