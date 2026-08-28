import { HeadContent, Outlet, Scripts, createRootRoute } from '@tanstack/react-router'
import type { ReactNode } from 'react'
import { SiteFooter, SiteHeader } from '../components/site-chrome'
import stylesheet from '../styles.css?url'

const siteUrl = 'https://stackend.codelynx.dev'

export const Route = createRootRoute({
  head: () => ({
    meta: [
      { charSet: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { title: 'StackEnv — Make any machine agent-ready' },
      {
        name: 'description',
        content: 'One CLI and command center for projects, Git worktrees, Portly apps, screenshots, and secure remote development.',
      },
      { property: 'og:title', content: 'StackEnv — Make any machine agent-ready' },
      { property: 'og:description', content: 'Your local or VPS development environment, visible from one secure command center.' },
      { property: 'og:url', content: siteUrl },
      { property: 'og:type', content: 'website' },
      { name: 'twitter:card', content: 'summary' },
      { name: 'theme-color', content: '#090b10' },
    ],
    links: [
      { rel: 'stylesheet', href: stylesheet },
      { rel: 'icon', href: '/favicon.svg', type: 'image/svg+xml' },
      { rel: 'manifest', href: '/manifest.webmanifest' },
    ],
  }),
  component: RootComponent,
})

function RootComponent() {
  return (
    <RootDocument>
      <SiteHeader />
      <Outlet />
      <SiteFooter />
    </RootDocument>
  )
}

function RootDocument({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en">
      <head><HeadContent /></head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  )
}
