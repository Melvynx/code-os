import { Outlet, createRootRoute } from "@tanstack/react-router"

import { AppShell } from "@/components/app-shell"
import { EmptyState } from "@/components/page-state"
import { SearchProvider } from "@/contexts/search-context"

export const Route = createRootRoute({
  component: RootLayout,
  notFoundComponent: () => (
    <EmptyState title="Page not found" description="This command center section does not exist." />
  ),
})

function RootLayout() {
  return (
    <SearchProvider>
      <AppShell><Outlet /></AppShell>
    </SearchProvider>
  )
}
