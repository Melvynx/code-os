import { Link } from '@tanstack/react-router'
import { ArrowLeft, ArrowRight, CheckCircle2, GitBranch, ShieldCheck } from 'lucide-react'
import type { ReactNode } from 'react'

export function SyncCode({ children }: Readonly<{ children: string }>) {
  return <pre><code>{children}</code></pre>
}

export function SyncNotice({ children, security = false }: Readonly<{ children: ReactNode; security?: boolean }>) {
  const Icon = security ? ShieldCheck : CheckCircle2
  return <div className="sync-notice"><Icon /><div>{children}</div></div>
}

export function SyncGuideLayout({
  side,
  title,
  description,
  steps,
  otherSide,
  otherUrl,
}: Readonly<{
  side: string
  title: string
  description: string
  steps: Array<{ title: string; children: ReactNode }>
  otherSide: string
  otherUrl: '/skills-sync/vps' | '/skills-sync/local'
}>) {
  return (
    <main className="sync-guide-page">
      <div className="container sync-guide-header">
        <Link to="/skills-sync" className="back-link"><ArrowLeft /> Skills Sync</Link>
        <span className="section-label">SETUP GUIDE · {side}</span>
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      <div className="container sync-guide-layout">
        <aside className="sync-step-nav">
          <span>IN THIS GUIDE</span>
          {steps.map((step, index) => <a href={`#step-${index + 1}`} key={step.title}><b>0{index + 1}</b>{step.title}</a>)}
        </aside>
        <article className="sync-guide-content">
          {steps.map((step, index) => (
            <section id={`step-${index + 1}`} className="sync-guide-step" key={step.title}>
              <div className="sync-step-number">0{index + 1}</div>
              <h2>{step.title}</h2>
              {step.children}
            </section>
          ))}
          <section className="sync-recovery">
            <span className="section-label">SAFE RECOVERY</span>
            <h2>If Git reports a conflict</h2>
            <p>The worker stops and will not stage anything else while a rebase, merge, or unresolved file exists. Inspect the files and choose one path:</p>
            <SyncCode>{`git -C ~/.agents status

# Keep the rebase and resolve each reported file
git -C ~/.agents add PATH_TO_RESOLVED_FILE
git -C ~/.agents rebase --continue
~/.local/bin/stackenv-skills-sync

# Or return to the exact state before the pull
git -C ~/.agents rebase --abort`}</SyncCode>
          </section>
          <div className="sync-next-guide">
            <GitBranch />
            <div><span>NEXT SIDE</span><b>Configure {otherSide}</b></div>
            <Link to={otherUrl}>Open guide <ArrowRight /></Link>
          </div>
        </article>
      </div>
    </main>
  )
}
