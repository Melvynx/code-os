import { Link } from '@tanstack/react-router'
import { ArrowRight, CheckCircle2, GitBranch, ShieldCheck } from 'lucide-react'
import type { ReactNode } from 'react'
import { SkillsDocsShell } from './docs-shell'

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
  const active = side === 'VPS' ? 'vps' : 'local'

  return (
    <SkillsDocsShell active={active}>
      <header className="documentation-header guide-header">
        <p className="docs-breadcrumb">Documentation / Skills sync / {side}</p>
        <h1>{title}</h1>
        <p>{description}</p>
      </header>
      <div className="sync-guide-layout">
        <nav className="sync-step-nav" aria-label="On this page">
          <p>On this page</p>
          {steps.map((step, index) => <a href={`#step-${index + 1}`} key={step.title}>{step.title}</a>)}
          <a href="#conflicts">Conflict recovery</a>
        </nav>
        <article className="sync-guide-content">
          {steps.map((step, index) => (
            <section id={`step-${index + 1}`} className="sync-guide-step" key={step.title}>
              <p className="step-count">Step {index + 1} of {steps.length}</p>
              <h2>{step.title}</h2>
              {step.children}
            </section>
          ))}
          <section className="sync-recovery" id="conflicts">
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
    </SkillsDocsShell>
  )
}
