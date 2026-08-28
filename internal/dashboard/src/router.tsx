import { createRouter } from "@tanstack/react-router"

import { PageError } from "@/components/page-state"
import { routeTree } from "@/routeTree.gen"

export const router = createRouter({
  routeTree,
  basepath: "/app",
  defaultPreload: "intent",
  defaultPreloadStaleTime: 0,
  defaultErrorComponent: ({ error, reset }) => (
    <PageError message={error.message || "An unexpected dashboard error occurred."} retry={reset} />
  ),
})

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}
