import { createFileRoute } from '@tanstack/react-router'
import { Link } from '@tanstack/react-router'
import { ArrowRight, ArrowUpRight, Info, ShieldCheck } from 'lucide-react'
import type { ReactNode } from 'react'
import { repositoryUrl } from '../components/site-chrome'

export const Route = createFileRoute('/docs')({
  head: () => ({
    meta: [
      { title: 'Documentation — Code OS' },
      { name: 'description', content: 'Install, configure, and operate Code OS on a local machine or VPS.' },
    ],
    links: [{ rel: 'canonical', href: 'https://code-os.mlvcdn.com/docs' }],
  }),
  component: DocsPage,
})

const nav = [
  ['overview', 'Overview'], ['install', 'Installation'], ['setup', 'Setup'], ['commands', 'Commands'],
  ['command-center', 'Command center'], ['worktrees', 'Worktrees'], ['runtime', 'Processes'],
  ['screenshots', 'Screenshots'], ['cloudflare', 'Cloudflare'], ['startup', 'Startup'],
  ['security', 'Security'], ['configuration', 'Configuration'],
]

function CodeBlock({ children }: Readonly<{ children: string }>) {
  return <pre><code>{children}</code></pre>
}

function DocsSection({ id, title, children }: Readonly<{ id: string; title: string; children: ReactNode }>) {
  return <section className="docs-section" id={id}><h2>{title}</h2>{children}</section>
}

function DocsPage() {
  return (
    <main className="docs-page">
      <div className="docs-hero container">
        <span className="section-label">DOCUMENTATION · v0.1.4</span>
        <h1>Build a legible<br />development environment.</h1>
        <p>Install Code OS, connect the tools you already use, and expose one secure command center to humans and coding agents.</p>
      </div>
      <div className="container docs-layout">
        <aside className="docs-nav">
          <span>ON THIS PAGE</span>
          {nav.map(([id, label]) => <a href={`#${id}`} key={id}>{label}</a>)}
          <a className="docs-source" href={repositoryUrl} target="_blank" rel="noreferrer">GitHub <ArrowUpRight size={13} /></a>
        </aside>
        <article className="docs-content">
          <DocsSection id="overview" title="Overview">
            <p>Code OS is an operational command center for local or remote development machines. It discovers projects and Git worktrees, observes Portly applications and coding agents, indexes screenshots, and serves the result through one authenticated dashboard.</p>
            <div className="notice"><Info /><p><b>Safe boundary</b> Code OS observes project repositories without mutating them. Only the separately configured private skills repository is synchronized by <code>code-os skills-sync</code>.</p></div>
            <Link className="docs-feature-link" to="/skills-sync"><span>NEW GUIDE</span><b>Synchronize skills between your VPS and computer</b><ArrowRight /></Link>
          </DocsSection>

          <DocsSection id="install" title="Installation">
            <h3>Linux amd64</h3>
            <CodeBlock>{`mkdir -p ~/.local/bin
curl -L https://github.com/Melvynx/code-os/releases/latest/download/code-os-linux-amd64 -o ~/.local/bin/code-os
chmod +x ~/.local/bin/code-os
code-os version`}</CodeBlock>
            <h3>Build from source</h3>
            <CodeBlock>{`git clone https://github.com/Melvynx/code-os.git
cd code-os
corepack pnpm --dir internal/dashboard install --frozen-lockfile
corepack pnpm --dir internal/dashboard build
go build -o code-os ./cmd/code-os`}</CodeBlock>
          </DocsSection>

          <DocsSection id="setup" title="Setup">
            <p>The guided flow asks where projects and screenshots live, how Portly is configured, where the dashboard stores its credentials, and which Cloudflare and skills-sync settings belong to this machine.</p>
            <CodeBlock>{`code-os setup
code-os doctor
code-os service install`}</CodeBlock>
            <p>On Linux, <code>service install</code> starts Code OS immediately, enables it after every reboot, and installs the skills-sync timer when a repository is configured. Use <code>code-os dashboard</code> only for an interactive foreground run.</p>
            <p>Generate dedicated secrets instead of reusing a personal password:</p>
            <CodeBlock>{`openssl rand -base64 32 > dashboard-password
openssl rand -hex 32 > media-bypass-key
openssl rand -hex 32 > session-key
chmod 600 dashboard-password media-bypass-key session-key`}</CodeBlock>
          </DocsSection>

          <DocsSection id="commands" title="Commands">
            <div className="command-list">
              <div><code>code-os setup</code><span>Create or update the configuration.</span></div>
              <div><code>code-os scan</code><span>Refresh the environment snapshot.</span></div>
              <div><code>code-os status</code><span>Print the current state and health.</span></div>
              <div><code>code-os doctor</code><span>Check paths, credentials, and integrations.</span></div>
              <div><code>code-os dashboard</code><span>Run the local command center.</span></div>
              <div><code>code-os cloudflare</code><span>Inspect the configured Tunnel integration.</span></div>
              <div><code>code-os service install</code><span>Install, start, and enable the Linux user services.</span></div>
              <div><code>code-os skills-sync</code><span>Synchronize the configured private skills repository.</span></div>
              <div><code>code-os version</code><span>Print the installed release.</span></div>
            </div>
          </DocsSection>

          <DocsSection id="command-center" title="Command center">
            <p>Opening the Code OS hostname routes directly to <code>/app/</code>. A signed-in or explicitly trusted connection sees the live dashboard; every other connection receives the HTML sign-in form. Public product documentation remains available at <code>/docs</code>.</p>
            <p>The dashboard brings projects, every Git worktree, application state, Git changes, grouped screenshot evidence, and machine settings into one operational view. Settings can change project roots, Cloudflare fields, skills repository details, and credential-file locations. Secret values are write-only.</p>
          </DocsSection>

          <DocsSection id="worktrees" title="Projects and worktrees">
            <p>Each configured root is scanned for repositories and nested applications. Code OS calls <code>git worktree list --porcelain</code> for every repository, so linked checkouts are first-class projects rather than invisible copies.</p>
            <p>The dashboard shows their branch, path, ahead/behind state, and modified, added, deleted, untracked, or conflicted file counts.</p>
          </DocsSection>

          <DocsSection id="runtime" title="Applications and coding agents">
            <p>Portly remains the source of truth for persistent applications and duplicate prevention. The Applications page keeps the complete Portly inventory visible, including stopped and failed entries, and adds a live resource view for currently running processes.</p>
            <p>CPU and resident memory are shown for each running Portly app and for grouped Codex, Cursor, Claude, OpenCode, Aider, and Gemini process trees. Wrapped agents are detected by safe executable signatures without returning their complete command arguments to the browser.</p>
            <h3>Stopping a process</h3>
            <ul>
              <li>Application controls call Portly with the exact currently running application ID.</li>
              <li>Agent controls send <code>SIGTERM</code> to an identifier containing both PID and kernel start time, which prevents PID-reuse mistakes.</li>
              <li>Every action requires a confirmation and an authenticated same-origin <code>POST</code>.</li>
              <li>Code OS refuses to terminate the process tree hosting the command center itself.</li>
            </ul>
          </DocsSection>

          <DocsSection id="screenshots" title="Screenshots and evidence">
            <p>Visual artifacts are grouped by feature directory and ordered by recency. The authenticated gallery provides a human view; an optional media bypass URL lets Cursor or Codex render a specific image directly.</p>
            <CodeBlock>{`https://your-code-os.example/media/SCREENSHOT_ID?bp=YOUR_MEDIA_KEY`}</CodeBlock>
            <CodeBlock>{`https://your-code-os.example/files/FEATURE/EVIDENCE.png?bp=YOUR_MEDIA_KEY`}</CodeBlock>
            <div className="notice"><ShieldCheck /><p><b>Narrow permission</b> The bypass applies only to <code>GET</code> and <code>HEAD</code> image requests under <code>/media</code> and <code>/files</code>. It never authenticates the dashboard, settings, API, or application gateway.</p></div>
          </DocsSection>

          <DocsSection id="cloudflare" title="Cloudflare Tunnel">
            <p>Keep Code OS bound to <code>127.0.0.1</code>, then route the main hostname and each <code>portNNNN</code> hostname to Code OS through Cloudflare Tunnel. The gateway requires the Code OS session before proxying a healthy Portly application.</p>
            <CodeBlock>{`code-os dashboard
# Tunnel origin example
http://127.0.0.1:7890`}</CodeBlock>
          </DocsSection>

          <DocsSection id="startup" title="Start automatically on Linux">
            <p><code>code-os service install</code> writes a hardened systemd user service for the current binary, enables user lingering, reloads systemd, enables the unit, and restarts it immediately. The daemon therefore returns after a VPS reboot without requiring an SSH login.</p>
            <CodeBlock>{`code-os service install
code-os doctor
systemctl --user status code-os.service --no-pager
loginctl show-user "$USER" -p Linger`}</CodeBlock>
            <p>Running the install command again updates the unit to the current binary and safely restarts the service. If skills synchronization is configured, the same command installs and enables its two-minute timer.</p>
          </DocsSection>

          <DocsSection id="security" title="Security model">
            <ul>
              <li>The dashboard listens on loopback by default.</li>
              <li>Credentials are read from permission-restricted files, not committed config.</li>
              <li>Successful sign-in creates an HTTP-only session cookie.</li>
              <li>A stable 256-bit signing key keeps sessions valid across safe restarts.</li>
              <li>After signing in, Settings shows the detected exact IP and a <b>Trust this IP</b> button. The same control can revoke it immediately.</li>
              <li>The JSON API requires a valid session or an IP that you explicitly trusted after signing in.</li>
              <li>The media bypass is separately generated and scoped to image reads.</li>
              <li>Cloudflare provides the public TLS boundary.</li>
            </ul>
          </DocsSection>

          <DocsSection id="configuration" title="Configuration">
            <p>The exact file is generated by <code>code-os setup</code> and can be updated from <code>/app/settings</code>. Token values are write-only and never returned by the API. Configuration changes require a service restart; trusting or revoking the current IP applies immediately.</p>
            <CodeBlock>{`{
  "version": 1,
  "environmentName": "dev-vps",
  "environmentType": "remote",
  "address": "127.0.0.1:7890",
  "projectsRoots": ["/root/projects"],
  "screenshotsRoot": "/root/screenshots",
  "filesRoot": "/root/.local/share/code-os/files",
  "dataDir": "/root/.local/share/code-os",
  "portlyBinary": "portly",
  "auth": {
    "username": "code-os",
    "passwordFile": "/root/.config/code-os/dashboard-password",
    "bypassKeyFile": "/root/.config/code-os/media-bypass-key",
    "sessionKeyFile": "/root/.config/code-os/session-key",
    "trustedIPsFile": "/root/.config/code-os/trusted-ips"
  },
  "skills": {
    "repository": "git@github.com:YOUR_ACCOUNT/agents-config.git",
    "directory": "/root/.agents",
    "branch": "main"
  }
}`}</CodeBlock>
            <p className="docs-end">Need implementation details? Read the <a href={`${repositoryUrl}/blob/main/README.md`} target="_blank" rel="noreferrer">project README <ArrowUpRight size={13} /></a>.</p>
          </DocsSection>
        </article>
      </div>
    </main>
  )
}
