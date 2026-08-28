import { Link } from '@tanstack/react-router'
import { BookOpen, GitBranch, Laptop, Search, Server, ShieldCheck } from 'lucide-react'
import type { ReactNode } from 'react'
import { useState } from 'react'

type DocsSection = 'overview' | 'vps' | 'local'

const navigation = [
  { id: 'overview' as const, label: 'Overview', to: '/skills-sync' as const, icon: BookOpen },
  { id: 'vps' as const, label: 'VPS setup', to: '/skills-sync/vps' as const, icon: Server },
  { id: 'local' as const, label: 'Computer setup', to: '/skills-sync/local' as const, icon: Laptop },
]

const reference = [
  { label: 'Security model', href: '/docs#security', icon: ShieldCheck },
  { label: 'Configuration', href: '/docs#configuration', icon: GitBranch },
]

export function SkillsDocsShell({ active, children }: Readonly<{ active: DocsSection; children: ReactNode }>) {
  const [query, setQuery] = useState('')
  const normalizedQuery = query.trim().toLowerCase()
  const visibleNavigation = navigation.filter((item) => item.label.toLowerCase().includes(normalizedQuery))
  const visibleReference = reference.filter((item) => item.label.toLowerCase().includes(normalizedQuery))

  return (
    <main className="documentation-page">
      <div className="container documentation-layout">
        <aside className="documentation-sidebar">
          <label className="docs-search">
            <Search aria-hidden="true" />
            <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Filter documentation" aria-label="Filter documentation" />
          </label>
          <nav aria-label="Skills sync documentation">
            <p>Skills sync</p>
            {visibleNavigation.map(({ id, label, to, icon: Icon }) => (
              <Link key={id} to={to} className={id === active ? 'active' : undefined}><Icon aria-hidden="true" />{label}</Link>
            ))}
            <p>Reference</p>
            {visibleReference.map(({ label, href, icon: Icon }) => <a key={label} href={href}><Icon aria-hidden="true" />{label}</a>)}
            {visibleNavigation.length === 0 && visibleReference.length === 0 ? <span className="docs-search-empty">No matching page</span> : null}
          </nav>
        </aside>
        <div className="documentation-content">{children}</div>
      </div>
    </main>
  )
}
