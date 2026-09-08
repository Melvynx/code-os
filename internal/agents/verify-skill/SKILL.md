---
name: verify
description: Verify a change through current evidence from its real surface. Use when the user invokes $verify or asks to prove, check, screenshot, or demonstrate something. On a Code OS host, store screenshots under filesRoot/skills-verify and embed via the configured dashboard host — never a hardcoded CDN.
---

<!-- code-os:verify-skill -->

# Verify

Prove the explicit `$verify` claim — or the latest user request — with current evidence from the real surface. Do not settle for “it probably works.”

1. Define observable pass criteria.
2. Exercise the real user-facing surface. Static checks support runtime proof; they do not replace it.
3. Capture and deliver current evidence.

## Screenshots on a Code OS host

When `~/.config/code-os/config.json` exists, load `$code-os` and read `cloudflare.dashboardHost`, `filesRoot`, and `auth.bypassKeyFile` from that file. Those values are the only hosts and paths to use.

1. Write the image under `filesRoot/skills-verify/<run-id>/`.
2. Codex app/Desktop: `![F01 - description](/absolute/filesRoot/skills-verify/<run-id>/F01.png)`. Do not use a `?bp=` URL in Codex Markdown.
3. Cursor or a renderer without workspace-file access: URL-encode the bypass key without printing it and embed `![F01 - description](https://<dashboardHost>/files/skills-verify/<run-id>/F01.png?bp=<encoded-key>)`.
4. Local-file gate: stored file exists, readable, non-zero bytes, non-zero decoded dimensions. Keep it in place.
5. HTTPS gate: HTTP 200, expected image type, `Cache-Control: private, no-store`, non-zero decoded size. The same `bp` must get HTTP 401 on `/api/health`.

The leading `!` is mandatory. A broken-image placeholder means verification cannot pass. Treat the complete bypass URL as a bearer secret. Never invent `mlvcdn.com` or another CDN; never start a separate static server.

## Non-visual evidence

Preserve the command or request, output, errors, status, and authoritative read-back.

4. Mark each criterion `PASS` or `NOT PROVEN`.

Finish with `PASS` only when every criterion has current evidence. Otherwise `BLOCKED — NOT PROVEN` with the exact blocker.
