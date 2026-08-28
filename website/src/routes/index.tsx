import { createFileRoute } from '@tanstack/react-router'
import { ArrowRight, Camera, Check, Cloud, Code2, Copy, FolderGit2, GitBranch, Play, ScanSearch, ShieldCheck, Terminal } from 'lucide-react'
import { useState } from 'react'
import { StackEnvWindow } from '../components/stackenv-window'
import { repositoryUrl } from '../components/site-chrome'

export const Route = createFileRoute('/')({
  head: () => ({ links: [{ rel: 'canonical', href: 'https://stackend.codelynx.dev' }] }),
  component: HomePage,
})

const installCommand = 'curl -L https://github.com/Melvynx/stackenv/releases/latest/download/stackenv-linux-amd64 -o ~/.local/bin/stackenv && chmod +x ~/.local/bin/stackenv'

const capabilities = [
  { icon: FolderGit2, title: 'Projects and subprojects', copy: 'Discover repositories and nested applications from one configured root.' },
  { icon: GitBranch, title: 'Every worktree', copy: 'Index the main checkout and every linked Git worktree, including branch and dirty state.' },
  { icon: Play, title: 'Portly applications', copy: 'See the processes Portly supervises, their ports, health, memory, and commands.' },
  { icon: Camera, title: 'Visual evidence', copy: 'Group screenshots by feature and serve them through protected, agent-readable URLs.' },
  { icon: Cloud, title: 'Cloudflare transport', copy: 'Expose the command center through your own HTTPS hostname without opening the app port.' },
  { icon: Code2, title: 'Machine-readable state', copy: 'Inspect the same environment snapshot as JSON from the CLI or authenticated API.' },
]

function HomePage() {
  const [copied, setCopied] = useState(false)

  async function copyInstall() {
    try {
      await navigator.clipboard.writeText(installCommand)
    } catch {
      const textarea = document.createElement('textarea')
      textarea.value = installCommand
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.append(textarea)
      textarea.select()
      document.execCommand('copy')
      textarea.remove()
    }
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1600)
  }

  return (
    <main>
      <section className="hero">
        <div className="hero-glow" />
        <div className="container hero-content">
          <div className="eyebrow"><span className="status-dot" /> v0.1.4 · Linux and macOS</div>
          <h1>Make any machine<br /><span>agent-ready.</span></h1>
          <p className="hero-copy">StackEnv turns a laptop or VPS into a clear, secure development environment—with every project, worktree, process, Git change, and screenshot in one place.</p>
          <div className="hero-actions">
            <a className="button button-primary" href="/docs">Read the docs <ArrowRight size={16} /></a>
            <a className="button button-secondary" href={repositoryUrl} target="_blank" rel="noreferrer">View source</a>
          </div>
          <div className="install-line">
            <Terminal size={17} />
            <code>curl -L …/stackenv-linux-amd64 -o ~/.local/bin/stackenv</code>
            <button type="button" onClick={copyInstall} aria-label="Copy install command">{copied ? <Check size={16} /> : <Copy size={16} />}</button>
          </div>
        </div>
      </section>

      <section className="facts" aria-label="StackEnv facts">
        <div className="container facts-grid">
          <div><strong>All worktrees</strong><span>Not only the main checkout</span></div>
          <div><strong>127.0.0.1</strong><span>Private by default</span></div>
          <div><strong>Portly-aware</strong><span>One process supervisor</span></div>
          <div><strong>MIT</strong><span>Open source</span></div>
        </div>
      </section>

      <section className="section command-section" id="command-center">
        <div className="container">
          <div className="section-heading centered">
            <span className="section-label">COMMAND CENTER</span>
            <h2>Your development environment,<br />finally observable.</h2>
            <p>StackEnv reads the state already present on your machine and makes it useful to both humans and coding agents.</p>
          </div>
          <StackEnvWindow />
          <div className="architecture-line">
            <span><Terminal /> stackenv CLI</span><i /><span><ScanSearch /> read-only scanner</span><i /><span><LayoutGlyph /> command center</span>
          </div>
        </div>
      </section>

      <section className="section" id="capabilities">
        <div className="container">
          <div className="section-heading split-heading">
            <div><span className="section-label">CAPABILITIES</span><h2>Less context hunting.<br />More building.</h2></div>
            <p>A practical layer over Git, Portly, local files, and Cloudflare—not another replacement for them.</p>
          </div>
          <div className="capability-grid">
            {capabilities.map(({ icon: Icon, title, copy }, index) => (
              <article className="capability-card" key={title}>
                <span className="card-index">0{index + 1}</span><Icon />
                <h3>{title}</h3><p>{copy}</p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="section" id="security">
        <div className="container security-panel">
          <div>
            <span className="section-label">SECURE BY CONSTRUCTION</span>
            <h2>Remote visibility.<br />Local trust boundary.</h2>
            <p>The dashboard binds to loopback. Cloudflare Tunnel carries HTTPS. Authentication protects the UI and API, while a scoped media bypass lets agents render screenshots without unlocking the command center.</p>
            <a href="/docs#security">Read the security model <ArrowRight size={15} /></a>
          </div>
          <div className="security-stack">
            <div><ShieldCheck /><span><b>Dashboard auth</b><small>Session-backed sign-in</small></span><em>REQUIRED</em></div>
            <div><Camera /><span><b>Media bypass</b><small>GET/HEAD screenshots only</small></span><em>SCOPED</em></div>
            <div><Cloud /><span><b>Cloudflare Tunnel</b><small>No public application port</small></span><em>TLS</em></div>
          </div>
        </div>
      </section>

      <section className="section install-cta">
        <div className="container">
          <span className="section-label">START HERE</span>
          <h2>One binary. One setup flow.<br />Your whole environment.</h2>
          <div className="setup-steps"><code><b>01</b> stackenv setup</code><code><b>02</b> stackenv doctor</code><code><b>03</b> stackenv dashboard</code></div>
          <a className="button button-primary" href="/docs">Install StackEnv <ArrowRight size={16} /></a>
        </div>
      </section>
    </main>
  )
}

function LayoutGlyph() {
  return <span className="layout-glyph" aria-hidden="true"><i /><i /><i /></span>
}
