import { Link, useRouterState } from "@tanstack/react-router"
import {
  ActivityIcon,
  AppWindowIcon,
  CameraIcon,
  FolderGit2Icon,
  FolderSyncIcon,
  GalleryVerticalEndIcon,
  GitCompareArrowsIcon,
  HomeIcon,
  SettingsIcon,
} from "lucide-react"
import type { ComponentProps } from "react"

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

const navMain = [
  {
    title: "Workspace",
    items: [
      { title: "Overview", url: "/", icon: HomeIcon },
      { title: "Projects", url: "/projects", icon: FolderGit2Icon },
      { title: "Applications", url: "/applications", icon: AppWindowIcon },
    ],
  },
  {
    title: "Observe",
    items: [
      { title: "Git changes", url: "/git", icon: GitCompareArrowsIcon },
      { title: "Screenshots", url: "/screenshots", icon: CameraIcon },
    ],
  },
  {
    title: "System",
    items: [
      { title: "Skills sync", url: "/skills-sync", icon: FolderSyncIcon },
      { title: "Status", url: "/status", icon: ActivityIcon },
      { title: "Settings", url: "/settings", icon: SettingsIcon },
    ],
  },
] as const

function isActivePath(pathname: string, url: string) {
  return url === "/" ? pathname === "/" : pathname === url
}

export function AppSidebar({ ...props }: ComponentProps<typeof Sidebar>) {
  const pathname = useRouterState({ select: (state) => state.location.pathname })

  return (
    <Sidebar variant="floating" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <Link to="/" aria-label="Code OS overview">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                  <GalleryVerticalEndIcon />
                </div>
                <div className="flex flex-col gap-0.5 leading-none">
                  <span className="font-medium">Code OS</span>
                  <span className="text-muted-foreground">Command Center</span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        {navMain.map((group) => (
          <SidebarGroup key={group.title}>
            <SidebarGroupLabel>{group.title}</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {group.items.map((item) => {
                  const isActive = isActivePath(pathname, item.url)
                  return (
                    <SidebarMenuItem key={item.url}>
                      <SidebarMenuButton asChild isActive={isActive} tooltip={item.title}>
                        <Link to={item.url} activeOptions={{ exact: item.url === "/" }} aria-current={isActive ? "page" : undefined}>
                          <item.icon />
                          <span>{item.title}</span>
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>
    </Sidebar>
  )
}
