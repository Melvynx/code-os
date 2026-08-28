import { Link, createFileRoute } from '@tanstack/react-router'
import { ArrowRight, Check, GitBranch, Laptop, Server, ShieldCheck } from 'lucide-react'
import { SkillsDocsShell } from '../components/docs-shell'

export const Route = createFileRoute('/skills-sync')({
  head: () => ({
    meta: [
      { title: 'Skills Sync — StackEnv' },
      { name: 'description', content: 'Synchronize your agent skills between a VPS and your computer through a private Git repository.' },
    ],
    links: [{ rel: 'canonical', href: 'https://stackend.codelynx.dev/skills-sync' }],
  }),
  component: SkillsSyncPage,
})

function SkillsSyncPage() {
  return (
    <SkillsDocsShell active="overview">
      <header className="documentation-header">
        <div><p className="docs-breadcrumb">Documentation / Skills sync</p><h1>Keep every agent on the same skill library.</h1></div>
        <a className="copy-page" href="https://github.com/Melvynx/stackenv/blob/main/website/src/routes/skills-sync.tsx" target="_blank" rel="noreferrer">View source</a>
        <p>Use a private Git repository to synchronize <code>~/.agents</code> between a development VPS and your computer. Both machines run the same small, auditable worker.</p>
      </header>

      <section className="docs-overview-section">
        <h2>Choose where to start</h2>
        <p>Configure the VPS first when it already contains your canonical skills. Then connect your local machine without overwriting local-only work.</p>
        <div className="guide-card-grid">
          <Link to="/skills-sync/vps" className="guide-card">
            <Server aria-hidden="true" /><span><b>Set up the VPS</b><small>Publish the current ~/.agents library and start the systemd timer.</small></span><ArrowRight />
          </Link>
          <Link to="/skills-sync/local" className="guide-card">
            <Laptop aria-hidden="true" /><span><b>Set up your computer</b><small>Clone the shared library and schedule synchronization on macOS.</small></span><ArrowRight />
          </Link>
        </div>
      </section>

      <section className="docs-overview-section">
        <h2>How synchronization works</h2>
        <div className="sync-diagram" aria-label="VPS to private Git repository to computer">
          <div><Server /><span><b>VPS</b><small>~/.agents</small></span></div>
          <span className="sync-connector">push / pull</span>
          <div><GitBranch /><span><b>Private Git repository</b><small>main branch</small></span></div>
          <span className="sync-connector">push / pull</span>
          <div><Laptop /><span><b>Your computer</b><small>~/.agents</small></span></div>
        </div>
        <ul className="docs-check-list">
          <li><Check />Local changes are committed before pulling remote updates.</li>
          <li><Check />Pulls use rebase so the shared history stays linear and reviewable.</li>
          <li><Check />A directory lock prevents two workers from running at once.</li>
          <li><Check />Conflicts stop the worker without discarding either side.</li>
        </ul>
      </section>

      <section className="docs-callout">
        <ShieldCheck />
        <div><h2>Only the skill library is synchronized.</h2><p>Project repositories, StackEnv configuration, Cloudflare tokens, dashboard credentials, screenshots, and media bypass keys stay machine-local.</p></div>
      </section>
    </SkillsDocsShell>
  )
}
