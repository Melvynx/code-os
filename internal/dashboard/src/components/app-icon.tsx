import { useState } from "react"

import type { Application } from "@/api/schema"
import { applicationIconUrl, applicationTitle } from "@/lib/applications"
import { applicationColorKey, identityColor, identityColorBorder, identityColorSurface } from "@/lib/identity-color"
import { cn } from "@/lib/utils"

type AppIconProps = Readonly<{
  application: Application
  accent?: string
  size?: "sm" | "md" | "lg"
}>

const SIZES = {
  sm: "size-8 p-1",
  md: "size-11 p-1",
  lg: "size-14 p-1.5",
} as const

export function AppIcon({ application, accent, size = "md" }: AppIconProps) {
  const [failed, setFailed] = useState(false)
  const title = applicationTitle(application)
  const initial = title.charAt(0) || "?"
  const color = accent ?? identityColor(applicationColorKey(application))

  return (
    <span
      className={cn(
        "relative flex shrink-0 items-center justify-center overflow-hidden border text-white",
        SIZES[size],
      )}
      style={{ background: identityColorSurface(color), borderColor: identityColorBorder(color) }}
      aria-hidden="true"
    >
      {failed ? (
        <span className="text-sm font-semibold" style={{ color }}>{initial}</span>
      ) : (
        <img
          src={applicationIconUrl(application)}
          alt=""
          className="size-full object-contain"
          onError={() => setFailed(true)}
        />
      )}
    </span>
  )
}
