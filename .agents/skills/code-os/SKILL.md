---
name: code-os
description: Operate and verify Code OS dashboards, projects, Portly apps, Git state, screenshots, login sessions, and media bypass URLs on local or VPS development environments.
---

# Code OS

Use the local Code OS CLI and configuration as the source of truth.

## Inspect

1. Run `code-os doctor`.
2. Read `~/.config/code-os/config.json` for the loopback address, dashboard hostname, and credential file paths.
3. Run `code-os status` for a summary or `code-os scan` for the current snapshot.
4. Inspect the user service with `systemctl --user status code-os.service` on Linux.
5. Open `/app/skills-sync` for `~/.agents` origin match, timer state, and a same-origin **Sync now**.
6. Open `/app/status` for image-pipeline health: roots, bypass key mode, decode counts, and authenticated media render.

Keep Git inspection read-only. Use Portly for persistent development processes.

## Authenticate the dashboard

Open `https://<dashboardHost>/login`. The form uses standard `username` and `current-password` autocomplete fields, so use the configured username and the password stored at `auth.passwordFile`.

Never print, commit, or paste the password file contents.

## Publish and embed private visual evidence

Use a media-only bypass URL when Cursor, Codex, or another image client cannot keep the dashboard session:

```text
https://<dashboardHost>/media/<screenshot-id>?bp=<url-encoded-bypass-key>
```

For a new verification artifact, write it beneath the configured `filesRoot` using a feature directory and a stable evidence filename, then use:

```text
https://<dashboardHost>/files/<feature>/<evidence>.png?bp=<url-encoded-bypass-key>
```

1. Obtain the screenshot ID from an authenticated `/api/snapshot`.
2. Read the key path from `auth.bypassKeyFile`; consume its contents without printing them.
3. URL-encode the key and append it only to `GET` or `HEAD` image requests under `/media/` or `/files/`.
4. Verify HTTP 200, the expected image media type, and a non-zero decoded image size.
5. Verify the same `bp` value still receives HTTP 401 on `/api/health`.

Treat the complete media URL as a bearer secret. Share it only in the private task where the image is needed. Keep it out of source control, issue trackers, analytics, shell tracing, and broad logs. Code OS sends `Cache-Control: private, no-store` and `Referrer-Policy: no-referrer` for bypassed media.

Rotate a leaked key by replacing `auth.bypassKeyFile` with a new random value, setting mode `0600`, and restarting `code-os.service`.

Never start a separate static server or publish verification files anonymously. Code OS is the only supported file and screenshot surface.
