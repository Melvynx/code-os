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
  ['worktrees', 'Worktrees'], ['screenshots', 'Screenshots'], ['portly', 'Portly'], ['cloudflare', 'Cloudflare'],
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
            <p>Code OS is a read-only inventory and command center for local or remote development machines. It discovers projects and Git worktrees, reads Portly process state, indexes screenshots, and serves the result through an authenticated dashboard.</p>
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
            <p>The guided flow asks where projects and screenshots live, how Portly is configured, and where the dashboard stores its credentials.</p>
            <CodeBlock>{`code-os setup
code-os doctor
code-os dashboard`}</CodeBlock>
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
              <div><code>code-os skills-sync</code><span>Synchronize the configured private skills repository.</span></div>
              <div><code>code-os version</code><span>Print the installed release.</span></div>
            </div>
          </DocsSection>

          <DocsSection id="worktrees" title="Projects and worktrees">
            <p>Each configured root is scanned for repositories and nested applications. Code OS calls <code>git worktree list --porcelain</code> for every repository, so linked checkouts are first-class projects rather than invisible copies.</p>
            <p>The dashboard shows their branch, path, ahead/behind state, and modified, added, deleted, untracked, or conflicted file counts.</p>
          </DocsSection>

          <DocsSection id="screenshots" title="Screenshots and evidence">
            <p>Visual artifacts are grouped by feature directory and ordered by recency. The authenticated gallery provides a human view; an optional media bypass URL lets Cursor or Codex render a specific image directly.</p>
            <CodeBlock>{`https://your-code-os.example/media/SCREENSHOT_ID?bp=YOUR_MEDIA_KEY`}</CodeBlock>
            <CodeBlock>{`https://your-code-os.example/files/FEATURE/EVIDENCE.png?bp=YOUR_MEDIA_KEY`}</CodeBlock>
            <div className="notice"><ShieldCheck /><p><b>Narrow permission</b> The bypass applies only to <code>GET</code> and <code>HEAD</code> image requests under <code>/media</code> and <code>/files</code>. It never authenticates the dashboard, settings, API, or application gateway.</p></div>
          </DocsSection>

          <DocsSection id="portly" title="Portly integration">
            <p>Portly remains the process supervisor and prevents duplicate development servers. Code OS reads its state to present application names, ports, commands, health, and memory alongside the repository that owns each process.</p>
          </DocsSection>

          <DocsSection id="cloudflare" title="Cloudflare Tunnel">
            <p>Keep Code OS bound to <code>127.0.0.1</code>, then route the main hostname and each <code>portNNNN</code> hostname to Code OS through Cloudflare Tunnel. The gateway requires the Code OS session before proxying a healthy Portly application.</p>
            <CodeBlock>{`code-os dashboard
# Tunnel origin example
http://127.0.0.1:7890`}</CodeBlock>
          </DocsSection>

          <DocsSection id="security" title="Security model">
            <ul>
              <li>The dashboard listens on loopback by default.</li>
              <li>Credentials are read from permission-restricted files, not committed config.</li>
              <li>Successful sign-in creates an HTTP-only session cookie.</li>
              <li>A stable 256-bit signing key keeps sessions valid across safe restarts.</li>
              <li>The JSON API always requires a valid session.</li>
              <li>The media bypass is separately generated and scoped to image reads.</li>
              <li>Cloudflare provides the public TLS boundary.</li>
            </ul>
          </DocsSection>

          <DocsSection id="configuration" title="Configuration">
            <p>The exact file is generated by <code>code-os setup</code> and can be updated from <code>/app/settings</code>. Token values are write-only and never returned by the API.</p>
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
    "sessionKeyFile": "/root/.config/code-os/session-key"
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
