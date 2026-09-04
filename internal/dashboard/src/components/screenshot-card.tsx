import type { Screenshot } from "@/api/schema"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { formatRelativeTime } from "@/lib/format"

export function screenshotMediaUrl(screenshot: Pick<Screenshot, "id">) {
  return `/media/${encodeURIComponent(screenshot.id)}`
}

export function ScreenshotCard({ screenshot, compact = false, showContext = true }: Readonly<{ screenshot: Screenshot; compact?: boolean; showContext?: boolean }>) {
  const source = screenshotMediaUrl(screenshot)
  const context = screenshot.project ?? screenshot.group ?? "Screenshot"

  return (
    <Dialog>
      <DialogTrigger asChild>
        <button type="button" className="group block w-full overflow-hidden rounded-xl border bg-card text-left transition-colors hover:border-foreground">
          <span className={compact ? "block aspect-video overflow-hidden" : "block aspect-[16/10] overflow-hidden"}>
            <img src={source} alt={screenshot.name} loading="lazy" decoding="async" className="size-full object-cover transition-opacity group-hover:opacity-80" />
          </span>
          {!compact ? (
            <span className="block border-t p-4">
              <strong className="block truncate text-sm font-medium">{screenshot.name}</strong>
              <span className="mt-1 block text-xs text-muted-foreground">{showContext ? `${context} · ` : ""}{formatRelativeTime(screenshot.createdAt)}</span>
            </span>
          ) : null}
        </button>
      </DialogTrigger>
      <DialogContent className="max-w-5xl sm:max-w-5xl">
        <DialogHeader>
          <DialogTitle>{screenshot.name}</DialogTitle>
          <DialogDescription>{context} · {formatRelativeTime(screenshot.createdAt)}</DialogDescription>
        </DialogHeader>
        <img src={source} alt={screenshot.name} className="max-h-[76vh] w-full object-contain" />
      </DialogContent>
    </Dialog>
  )
}
