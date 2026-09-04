import { AlertTriangleIcon, InboxIcon } from "lucide-react"

import { Alert, AlertAction, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty"
import { Skeleton } from "@/components/ui/skeleton"

export function PageLoading({ label = "Loading environment" }: Readonly<{ label?: string }>) {
  return (
    <div aria-label={label} aria-busy="true" className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      {Array.from({ length: 8 }, (_, index) => (
        <Skeleton key={index} className="h-36" />
      ))}
    </div>
  )
}

export function PageError({ message, retry }: Readonly<{ message: string; retry: () => void }>) {
  return (
    <Alert variant="destructive">
      <AlertTriangleIcon />
      <AlertTitle>Unable to load the environment</AlertTitle>
      <AlertDescription>{message}</AlertDescription>
      <AlertAction>
        <Button variant="outline" size="sm" onClick={retry}>Try again</Button>
      </AlertAction>
    </Alert>
  )
}

export function EmptyState({ title, description }: Readonly<{ title: string; description: string }>) {
  return (
    <Empty className="border">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <InboxIcon />
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        <EmptyDescription>{description}</EmptyDescription>
      </EmptyHeader>
    </Empty>
  )
}
