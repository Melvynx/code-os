import { Link } from '@tanstack/react-router'
import { Menu, X } from 'lucide-react'
import { useState } from 'react'

export const repositoryUrl = 'https://github.com/Melvynx/stackenv'

function Brand() {
  return (
    <Link to="/" className="brand" aria-label="StackEnv home">
      <span className="brand-mark">S</span>
      <span>StackEnv</span>
    </Link>
  )
}

export function SiteHeader() {
  const [open, setOpen] = useState(false)

  return (
    <header className="site-header">
      <div className="header-inner">
        <Brand />
        <nav className={open ? 'main-nav is-open' : 'main-nav'} aria-label="Primary navigation">
          <a href="/#command-center" onClick={() => setOpen(false)}>Command center</a>
          <a href="/#capabilities" onClick={() => setOpen(false)}>Capabilities</a>
          <a href="/#security" onClick={() => setOpen(false)}>Security</a>
          <Link to="/skills-sync" onClick={() => setOpen(false)}>Skills sync</Link>
          <Link to="/docs" onClick={() => setOpen(false)}>Docs</Link>
        </nav>
        <div className="header-actions">
          <a className="icon-link" href={repositoryUrl} target="_blank" rel="noreferrer" aria-label="GitHub repository">
            <svg width="17" height="17" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 .7C5.7.7.7 5.8.7 12.1c0 5 3.2 9.3 7.7 10.8.6.1.8-.3.8-.6v-2.2c-3.1.7-3.8-1.3-3.8-1.3-.5-1.3-1.2-1.6-1.2-1.6-1-.7.1-.7.1-.7 1.1.1 1.7 1.2 1.7 1.2 1 1.7 2.6 1.2 3.2.9.1-.7.4-1.2.7-1.5-2.5-.3-5.1-1.3-5.1-5.6 0-1.2.4-2.3 1.2-3.1-.1-.3-.5-1.5.1-3.1 0 0 1-.3 3.2 1.2a11 11 0 0 1 5.8 0c2.2-1.5 3.2-1.2 3.2-1.2.6 1.6.2 2.8.1 3.1.7.8 1.2 1.8 1.2 3.1 0 4.4-2.7 5.3-5.2 5.6.4.3.8 1 .8 2v3c0 .4.2.7.8.6a11.5 11.5 0 0 0 7.7-10.8C23.3 5.8 18.3.7 12 .7Z" /></svg>
          </a>
          <Link className="button button-primary button-small" to="/docs">Get started</Link>
          <button className="menu-button" type="button" onClick={() => setOpen((value) => !value)} aria-label="Toggle navigation" aria-expanded={open}>
            {open ? <X size={18} /> : <Menu size={18} />}
          </button>
        </div>
      </div>
    </header>
  )
}

export function SiteFooter() {
  return (
    <footer className="site-footer">
      <div className="footer-inner">
        <div>
          <Brand />
          <p>One command center for every development environment.</p>
        </div>
        <div className="footer-links">
          <Link to="/docs">Documentation</Link>
          <Link to="/skills-sync">Skills Sync</Link>
          <a href={`${repositoryUrl}/releases`} target="_blank" rel="noreferrer">Releases</a>
          <a href={repositoryUrl} target="_blank" rel="noreferrer">Source</a>
          <a href={`${repositoryUrl}/blob/main/LICENSE`} target="_blank" rel="noreferrer">MIT License</a>
        </div>
      </div>
    </footer>
  )
}
