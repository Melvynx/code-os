import { Link, useRouterState } from "@tanstack/react-router"
import { RefreshCwIcon, SearchIcon } from "lucide-react"
import type { CSSProperties, ReactNode } from "react"

import { useRefreshSnapshot } from "@/api/queries"
import { AppSidebar } from "@/components/app-sidebar"
import { Button } from "@/components/ui/button"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group"
import { Separator } from "@/components/ui/separator"
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { Spinner } from "@/components/ui/spinner"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { useDashboardSearch } from "@/contexts/search-context"

const PAGE_TITLES: Record<string, string> = {
  "/": "Overview",
  "/projects": "Projects",
  "/applications": "Applications",
  "/git": "Git changes",
  "/screenshots": "Screenshots",
  "/skills-sync": "Skills synchronization",
  "/status": "Status",
  "/settings": "Settings",
}

export function AppShell({ children }: Readonly<{ children: ReactNode }>) {
  const pathname = useRouterState({ select: (state) => state.location.pathname })
  const { query, setQuery } = useDashboardSearch()
  const refreshSnapshot = useRefreshSnapshot()
  const title = PAGE_TITLES[pathname] ?? "Command Center"

  return (
    <SidebarProvider style={{ "--sidebar-width": "19rem" } as CSSProperties}>
      <a href="#main-content" className="fixed left-3 top-3 -translate-y-20 bg-primary px-3 py-2 font-mono text-xs text-primary-foreground focus:translate-y-0">
        Skip to content
      </a>
      <AppSidebar />
      <SidebarInset id="main-content" tabIndex={-1}>
        <header className="flex h-16 shrink-0 items-center gap-2 px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 data-vertical:h-4 data-vertical:self-auto" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink asChild>
                  <Link to="/">Development environment</Link>
                </BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator className="hidden md:block" />
              <BreadcrumbItem>
                <BreadcrumbPage>{title}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          <div className="ml-auto flex items-center gap-2">
            <InputGroup className="w-40 sm:w-64 lg:w-80">
              <InputGroupAddon>
                <SearchIcon />
              </InputGroupAddon>
              <InputGroupInput
                type="search"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Search this view"
                aria-label="Search the current dashboard view"
              />
            </InputGroup>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button size="icon" variant="outline" onClick={() => refreshSnapshot.mutate()} disabled={refreshSnapshot.isPending} aria-label="Refresh environment">
                  {refreshSnapshot.isPending ? <Spinner /> : <RefreshCwIcon />}
                </Button>
              </TooltipTrigger>
              <TooltipContent>Refresh environment</TooltipContent>
            </Tooltip>
            <span className="sr-only" role="status" aria-live="polite">
              {refreshSnapshot.isPending ? "Refreshing environment" : refreshSnapshot.isSuccess ? "Environment refreshed" : refreshSnapshot.isError ? "Refresh failed" : ""}
            </span>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">{children}</div>
      </SidebarInset>
    </SidebarProvider>
  )
}
