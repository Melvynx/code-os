# Code OS Product Context

## Product

Code OS is a self-hosted command center that turns a Mac or Linux VPS into a consistent AI-assisted development environment. It discovers projects and subprojects, surfaces Git work in progress, observes and controls Portly-managed applications, groups agent CPU and memory usage, indexes visual evidence, and synchronizes shared agent skills through Git.

## Users

The primary users are developers and community members who want a productive remote or local coding machine without rebuilding the same process, tunnel, and agent configuration by hand. They need the current state at a glance and safe defaults around credentials and repository state.

## Personality

Austere, technical, trustworthy, and fast. The product should feel like an operational console, not a generic SaaS admin template.

## Design principles

1. Make environment state glanceable.
2. Show evidence and exact source state instead of decorative metrics.
3. Keep security boundaries explicit, especially around tunnels and media URLs.
4. Treat Portly and Git as sources of truth; avoid duplicate process state.
5. Make keyboard navigation, visible focus, reduced motion, and WCAG 2.2 AA contrast standard.
6. Make destructive controls explicit, confirmed, and limited to the exact process shown.

## Current capabilities

- The public root opens the command center; product documentation remains available at `/docs`.
- Portly applications and Codex, Cursor, Claude, OpenCode, Aider, and Gemini process trees are visible with live CPU and resident memory.
- Operators can stop an exact Portly application or send `SIGTERM` to an exact agent process after confirmation. Code OS protects its own process tree.
- A signed session protects the dashboard, APIs, private artifacts, and application gateways. A signed-in user can explicitly trust or revoke the current exact IP from Settings.
- Linux installs can run at boot without an interactive login. A configured private skills repository synchronizes on a two-minute systemd timer.

## Visual direction

Use the `black-grid` system: black canvas, #111 surfaces, #333 separators, white ink, #888 secondary text, #0070f3 actions, #50e3c2 success, and #e00 errors. Use Geist and Geist Mono, restrained radii, and no gradients, blur, glow, or decorative shadows.

Avoid glassmorphism, oversized marketing typography, floating cards, excessive pills, and ornamental charts.
