---
name: remote-webview
description: Expose healthy Portly servers through the Code OS gateway at the configured port{port}.<zone> host. Use to start, share, verify, open, close, or sync a private development port. Read the hostname template from Code OS config — never hardcode a CDN domain.
---

<!-- code-os:remote-webview-skill -->

# Remote webview

Expose each healthy persistent Portly server as `https://<publicPortHost>/<path>`.

Read `publicPortHost` from `~/.config/code-os/config.json`. Replace `{port}` with the bound port. Example: `port{port}.example.com` + `3013` → `https://port3013.example.com/`. The zone is whatever the operator configured — not a product constant.

A systemd timer reconciles Portly, Cloudflare DNS, Code OS tunnel ingress, and the dedicated host firewall every 15 seconds.

## Share a server

1. `portly status --json`. Use only `projects[].servers[]`; ignore `temporaryServers`.
2. If the persistent server is stopped, `portly start <server-name> --json`.
3. Wait until `running` and `healthy` is `true`.
4. Sync immediately:

```bash
python3 ~/.agents/skills/remote-webview/scripts/remote_webview.py sync
```

5. Verify the exact public URL in HTTP and a real browser. DNS or tunnel config alone is not proof. Return the HTTPS URL as a Markdown link.

A public `404` usually means the Portly server is stopped or unhealthy. Check Portly first.

## Commands

```bash
python3 ~/.agents/skills/remote-webview/scripts/remote_webview.py sync
python3 ~/.agents/skills/remote-webview/scripts/remote_webview.py status
python3 ~/.agents/skills/remote-webview/scripts/remote_webview.py open 3000
python3 ~/.agents/skills/remote-webview/scripts/remote_webview.py close 3000
```

`close` suppresses a port until `open` removes the suppression.

## Guardrails

- Publish only persistent `running` and healthy Portly servers.
- Keep infrastructure ports and ports below 1024 private.
- Create individual `portNNNN.<zone>` CNAMEs from `publicPortHost`. Never add a wildcard. Never invent a CDN host.
- Preserve every non-managed tunnel ingress and keep `http_status:404` last.
- Route every managed hostname to Code OS on `127.0.0.1:7890`.
- Never print the Cloudflare token.
