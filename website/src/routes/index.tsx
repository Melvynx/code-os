import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowRight, Camera, Check, Cloud, Copy, FolderGit2, GitBranch, Play, ShieldCheck, Terminal } from 'lucide-react'
import { useState } from 'react'
import { CodeOSWindow } from '../components/code-os-window'
import { repositoryUrl } from '../components/site-chrome'

export const Route = createFileRoute('/')({
  head: () => ({ links: [{ rel: 'canonical', href: 'https://code-os.mlvcdn.com' }] }),
  component: HomePage,
})

const installCommand = 'curl -L https://github.com/Melvynx/code-os/releases/latest/download/code-os-linux-amd64 -o ~/.local/bin/code-os && chmod +x ~/.local/bin/code-os'

const capabilities = [
  { icon: FolderGit2, title: 'Projects and worktrees', copy: 'Discover nested repositories, linked worktrees, branches, and dirty state from every configured root.' },
  { icon: Play, title: 'Running applications', copy: 'Read Portly process state without launching duplicates. See commands, ports, health, and memory in context.' },
  { icon: Camera, title: 'Visual evidence', copy: 'Group screenshots by feature and expose narrowly scoped media URLs that agents can actually render.' },
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
        <div className="container hero-grid">
          <div className="hero-content">
            <p className="hero-release"><span /> Code OS 0.1.4 for Linux and macOS</p>
            <h1>The development environment your agents can understand.</h1>
            <p className="hero-copy">Turn a laptop or VPS into one legible workspace for projects, Git worktrees, running processes, changes, and screenshots.</p>
            <div className="hero-actions">
              <Link className="button button-primary" to="/docs">Get started <ArrowRight size={16} /></Link>
              <a className="button button-secondary" href={repositoryUrl} target="_blank" rel="noreferrer">View on GitHub</a>
            </div>
          </div>
          <div className="hero-aside" aria-label="Code OS principles">
            <p>Built for the tools already on your machine.</p>
            <div><span>Git</span><span>Portly</span><span>Cloudflare</span><span>Codex</span><span>Cursor</span></div>
          </div>
        </div>
        <div className="container hero-product" id="command-center"><CodeOSWindow /></div>
      </section>

      <section className="product-strip" aria-label="Code OS facts">
        <div className="container"><span>All Git worktrees</span><span>Loopback by default</span><span>Portly-aware</span><span>Open source · MIT</span></div>
      </section>

      <section className="section feature-intro" id="capabilities">
        <div className="container">
          <div className="section-heading split-heading">
            <h2>Stop reconstructing machine state in every conversation.</h2>
            <p>Code OS reads the environment you already have. It does not replace Git, your process manager, or your tunnel.</p>
          </div>
          <div className="capability-list">
            {capabilities.map(({ icon: Icon, title, copy }) => (
              <article key={title}><Icon aria-hidden="true" /><h3>{title}</h3><p>{copy}</p></article>
            ))}
          </div>
        </div>
      </section>

      <section className="section feature-row">
        <div className="container feature-row-grid">
          <div className="feature-copy">
            <p className="feature-label">Source control</p>
            <h2>Every checkout. One exact Git picture.</h2>
            <p>The main repository is only part of the story. Code OS indexes linked worktrees and reports branch, ahead/behind state, and file-level change counts for each one.</p>
            <Link to="/docs" hash="worktrees">See how discovery works <ArrowRight size={15} /></Link>
          </div>
          <div className="git-preview" aria-label="Git worktree status preview">
            <div className="preview-title"><GitBranch /><span>Git changes</span><small>3 repositories</small></div>
            <GitPreviewRow project="lumail.io" branch="main" changes="119" tone="orange" />
            <GitPreviewRow project="code-os" branch="main" changes="22" tone="blue" />
            <GitPreviewRow project="ai-builder-club" branch="feature/editor" changes="5" tone="green" />
          </div>
        </div>
      </section>

      <section className="section feature-row feature-row-reverse">
        <div className="container feature-row-grid">
          <div className="process-preview" aria-label="Running applications preview">
            <div className="preview-title"><Play /><span>Applications</span><small>Portly state</small></div>
            <ProcessRow name="lumail.io / dev" port="3002" memory="8.4 GB" />
            <ProcessRow name="code-os / website" port="3004" memory="412 MB" />
            <ProcessRow name="ai-builder-club / web" port="3100" memory="841 MB" />
          </div>
          <div className="feature-copy">
            <p className="feature-label">Runtime</p>
            <h2>Know what is running before starting anything else.</h2>
            <p>Portly stays responsible for process supervision. Code OS makes its live state visible beside the repository that owns each application.</p>
            <a href="https://portly.melvynx.dev" target="_blank" rel="noreferrer">Learn about Portly <ArrowRight size={15} /></a>
          </div>
        </div>
      </section>

      <section className="section security-section" id="security">
        <div className="container security-layout">
          <div className="feature-copy">
            <p className="feature-label">Security boundary</p>
            <h2>Remote visibility without a public application port.</h2>
            <p>Code OS binds to loopback. Cloudflare Tunnel carries HTTPS. Origin sessions protect the dashboard and every app port; a scoped key can render one private image.</p>
            <Link to="/docs" hash="security">Read the security model <ArrowRight size={15} /></Link>
          </div>
          <div className="security-list">
            <SecurityRow icon={ShieldCheck} title="Dashboard session" detail="HTTP-only authenticated session" state="required" />
            <SecurityRow icon={Camera} title="Media bypass" detail="GET and HEAD images under /media and /files" state="scoped" />
            <SecurityRow icon={Cloud} title="Cloudflare Tunnel" detail="TLS at your own hostname" state="private" />
          </div>
        </div>
      </section>

      <section className="section final-cta">
        <div className="container final-cta-grid">
          <div><p className="feature-label">Install Code OS</p><h2>Make the machine legible.</h2><p>Install the binary, run the guided setup, and open your command center.</p></div>
          <div>
            <div className="install-line"><Terminal size={17} /><code>curl -L …/code-os-linux-amd64 -o ~/.local/bin/code-os</code><button type="button" onClick={copyInstall} aria-label="Copy install command">{copied ? <Check size={16} /> : <Copy size={16} />}</button></div>
            <div className="final-actions"><Link className="button button-primary" to="/docs">Open installation guide <ArrowRight size={16} /></Link><Link className="text-link" to="/skills-sync">Set up skills sync</Link></div>
          </div>
        </div>
      </section>
    </main>
  )
}

function GitPreviewRow({ project, branch, changes, tone }: Readonly<{ project: string; branch: string; changes: string; tone: string }>) {
  return <div className="git-preview-row"><span className={`preview-dot ${tone}`} /><span><b>{project}</b><small>{branch}</small></span><strong>{changes} changes</strong></div>
}

function ProcessRow({ name, port, memory }: Readonly<{ name: string; port: string; memory: string }>) {
  return <div className="process-preview-row"><span><b>{name}</b><small>pnpm dev</small></span><code>{port}</code><em>healthy</em><strong>{memory}</strong></div>
}

function SecurityRow({ icon: Icon, title, detail, state }: Readonly<{ icon: typeof ShieldCheck; title: string; detail: string; state: string }>) {
  return <div><Icon /><span><b>{title}</b><small>{detail}</small></span><em>{state}</em></div>
}
