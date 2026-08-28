import { AlertTriangleIcon, InboxIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"

export function PageLoading({ label = "Loading environment" }: Readonly<{ label?: string }>) {
  return (
    <div aria-label={label} aria-busy="true" className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      {Array.from({ length: 8 }, (_, index) => (
        <Skeleton key={index} className="h-36 border border-[#333]" />
      ))}
    </div>
  )
}

export function PageError({ message, retry }: Readonly<{ message: string; retry: () => void }>) {
  return (
    <div role="alert" className="border border-[#e00] bg-black p-6">
      <AlertTriangleIcon aria-hidden="true" className="size-5 text-[#e00]" />
      <h2 className="mt-4 text-base font-medium text-white">Unable to load the environment</h2>
      <p className="mt-2 max-w-xl text-sm text-[#888]">{message}</p>
      <Button className="mt-5" variant="outline" onClick={retry}>Try again</Button>
    </div>
  )
}

export function EmptyState({ title, description }: Readonly<{ title: string; description: string }>) {
  return (
    <div className="border border-dashed border-[#333] px-6 py-16 text-center">
      <InboxIcon aria-hidden="true" className="mx-auto size-5 text-[#888]" />
      <h2 className="mt-4 text-sm font-medium text-white">{title}</h2>
      <p className="mx-auto mt-2 max-w-md text-sm text-[#888]">{description}</p>
    </div>
  )
}
