import { Link, createFileRoute } from '@tanstack/react-router'
import { ArrowRight, GitBranch, Laptop, Server, ShieldCheck } from 'lucide-react'

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
    <main className="sync-hub-page">
      <section className="sync-hero">
        <div className="container">
          <span className="section-label">SKILLS SYNC · TWO-SIDED SETUP</span>
          <h1>One skill library.<br /><span>Every machine.</span></h1>
          <p>Keep <code>~/.agents</code> synchronized between your development VPS and your computer through a private Git repository.</p>
          <div className="sync-flow" aria-label="Skills synchronization architecture">
            <div><Server /><span><b>VPS</b><small>~/.agents</small></span></div>
            <i /><div className="sync-repository"><GitBranch /><span><b>Private Git</b><small>source of truth</small></span></div>
            <i /><div><Laptop /><span><b>Your computer</b><small>~/.agents</small></span></div>
          </div>
        </div>
      </section>

      <section className="section sync-choose">
        <div className="container">
          <div className="section-heading centered">
            <span className="section-label">CHOOSE A SIDE</span>
            <h2>Configure both machines.</h2>
            <p>Start with the VPS if it already contains your canonical skills, then connect your computer.</p>
          </div>
          <div className="sync-side-grid">
            <article>
              <span className="sync-side-index">01 · REMOTE</span><Server />
              <h2>VPS setup</h2>
              <p>Back up the current library, create the private repository, install the sync worker, and activate its systemd timer.</p>
              <Link to="/skills-sync/vps">Configure the VPS <ArrowRight /></Link>
            </article>
            <article>
              <span className="sync-side-index">02 · LOCAL</span><Laptop />
              <h2>Computer setup</h2>
              <p>Preserve local skills, clone the shared library, install the same worker, and schedule it with macOS launchd.</p>
              <Link to="/skills-sync/local">Configure your computer <ArrowRight /></Link>
            </article>
          </div>
        </div>
      </section>

      <section className="section sync-boundary">
        <div className="container">
          <ShieldCheck />
          <div><span className="section-label">CURRENT WORKFLOW</span><h2>Git transport today.<br />Native StackEnv sync next.</h2></div>
          <p>The guides install a small, auditable Git worker that runs every two minutes. StackEnv does not yet expose native <code>skills sync</code> commands, so the documentation never pretends that it does.</p>
        </div>
      </section>
    </main>
  )
}
