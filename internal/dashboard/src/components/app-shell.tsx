import { Link, useRouterState } from "@tanstack/react-router"
import { AppWindowIcon, CameraIcon, FolderGit2Icon, GitCompareArrowsIcon, HomeIcon, MenuIcon, RefreshCwIcon, SearchIcon, SettingsIcon } from "lucide-react"
import type { ReactNode } from "react"

import { useSnapshot, useRefreshSnapshot } from "@/api/queries"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Sheet, SheetClose, SheetContent, SheetDescription, SheetTitle, SheetTrigger } from "@/components/ui/sheet"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { useDashboardSearch } from "@/contexts/search-context"
import { formatRelativeTime } from "@/lib/format"
import { cn } from "@/lib/utils"

const NAVIGATION = [
  { to: "/", label: "Overview", icon: HomeIcon },
  { to: "/projects", label: "Projects", icon: FolderGit2Icon },
  { to: "/applications", label: "Applications", icon: AppWindowIcon },
  { to: "/git", label: "Git changes", icon: GitCompareArrowsIcon },
  { to: "/screenshots", label: "Screenshots", icon: CameraIcon },
  { to: "/settings", label: "Settings", icon: SettingsIcon },
] as const

const PAGE_TITLES: Record<string, string> = {
  "/": "Overview",
  "/projects": "Projects",
  "/applications": "Applications",
  "/git": "Git changes",
  "/screenshots": "Screenshots",
  "/settings": "Settings",
}

function Brand() {
  return (
    <Link to="/" className="flex items-center gap-3" aria-label="Code OS overview">
      <span aria-hidden="true" className="grid size-8 place-items-center border border-white bg-white font-mono text-xs font-semibold text-black">S</span>
      <span><strong className="block text-sm font-medium text-white">Code OS</strong><span className="block font-mono text-[10px] uppercase tracking-wider text-[#888]">Command Center</span></span>
    </Link>
  )
}

function Navigation({ mobile = false }: Readonly<{ mobile?: boolean }>) {
  return (
    <nav aria-label="Dashboard sections" className="space-y-0.5">
      {NAVIGATION.map(({ to, label, icon: Icon }) => {
        const link = (
          <Link
            to={to}
            activeOptions={{ exact: to === "/" }}
            className="flex h-10 items-center gap-3 rounded-sm px-3 font-mono text-xs uppercase tracking-wide text-[#888] transition-colors hover:bg-[#111] hover:text-white focus-visible:bg-[#111]"
            activeProps={{ className: "bg-[#111] text-white" }}
          >
            <Icon aria-hidden="true" className="size-4" /> {label}
          </Link>
        )
        return mobile ? <SheetClose key={to} asChild>{link}</SheetClose> : <div key={to}>{link}</div>
      })}
    </nav>
  )
}

function EnvironmentStatus() {
  const snapshot = useSnapshot()
  const hostname = window.location.hostname === "127.0.0.1" ? "Code OS local" : window.location.hostname
  const statusLabel = snapshot.isError ? "Connection error" : snapshot.data ? `Updated ${formatRelativeTime(snapshot.data.generatedAt)}` : "Connecting"

  return (
    <div className="flex items-center gap-3 border-t border-[#333] p-4">
      <span aria-hidden="true" className={cn("size-2 bg-[#50e3c2]", snapshot.isError && "bg-[#e00]", snapshot.isFetching && "animate-pulse")} />
      <div className="min-w-0"><strong className="block truncate text-xs font-medium text-white">{hostname}</strong><span className="block truncate text-[11px] text-[#888]">{statusLabel}</span></div>
    </div>
  )
}

function Sidebar() {
  return (
    <aside className="fixed inset-y-0 left-0 z-30 hidden w-56 border-r border-[#333] bg-black lg:flex lg:flex-col">
      <div className="flex h-20 shrink-0 items-center border-b border-[#333] px-5"><Brand /></div>
      <div className="flex-1 p-3"><Navigation /></div>
      <EnvironmentStatus />
    </aside>
  )
}

function MobileNavigation() {
  return (
    <Sheet>
      <SheetTrigger asChild><Button size="icon" variant="outline" aria-label="Open navigation"><MenuIcon /></Button></SheetTrigger>
      <SheetContent>
        <SheetTitle>Code OS navigation</SheetTitle>
        <SheetDescription>Navigate command center sections</SheetDescription>
        <div className="border-b border-[#333] p-5"><Brand /></div>
        <div className="flex-1 p-3"><Navigation mobile /></div>
        <EnvironmentStatus />
      </SheetContent>
    </Sheet>
  )
}

function Header() {
  const pathname = useRouterState({ select: (state) => state.location.pathname })
  const { query, setQuery } = useDashboardSearch()
  const refreshSnapshot = useRefreshSnapshot()
  const title = PAGE_TITLES[pathname] ?? "Command Center"

  return (
    <header className="sticky top-0 z-20 border-b border-[#333] bg-black px-4 py-4 sm:px-6 lg:h-20 lg:px-8 lg:py-0">
      <div className="flex flex-col gap-4 lg:h-full lg:flex-row lg:items-center lg:justify-between">
        <div className="flex items-center gap-3">
          <div className="lg:hidden"><MobileNavigation /></div>
          <div><p className="font-mono text-[10px] uppercase tracking-wider text-[#888]">Development environment</p><h1 className="mt-1 text-2xl font-medium tracking-tight text-white">{title}</h1></div>
        </div>
        <div className="flex items-center gap-2">
          <label className="relative min-w-0 flex-1 sm:w-80 sm:flex-none">
            <span className="sr-only">Search the current dashboard view</span>
            <SearchIcon aria-hidden="true" className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-[#888]" />
            <Input value={query} onChange={(event) => setQuery(event.target.value)} className="pl-9" type="search" placeholder="Search this view" />
          </label>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button size="icon" variant="outline" onClick={() => refreshSnapshot.mutate()} disabled={refreshSnapshot.isPending} aria-label="Refresh environment">
                <RefreshCwIcon className={cn(refreshSnapshot.isPending && "animate-spin")} />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Refresh environment</TooltipContent>
          </Tooltip>
          <span className="sr-only" role="status" aria-live="polite">{refreshSnapshot.isPending ? "Refreshing environment" : refreshSnapshot.isSuccess ? "Environment refreshed" : refreshSnapshot.isError ? "Refresh failed" : ""}</span>
        </div>
      </div>
    </header>
  )
}

export function AppShell({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <div className="min-h-screen bg-black text-white">
      <a href="#main-content" className="fixed left-3 top-3 z-[100] -translate-y-20 bg-white px-3 py-2 font-mono text-xs text-black focus:translate-y-0">Skip to content</a>
      <Sidebar />
      <div className="lg:pl-56">
        <Header />
        <main id="main-content" tabIndex={-1} className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">{children}</main>
      </div>
    </div>
  )
}
