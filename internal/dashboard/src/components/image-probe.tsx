import { useEffect, useState } from "react"

import type { ImageSample } from "@/api/status"
import { Badge } from "@/components/ui/badge"
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { formatBytes, formatRelativeTime } from "@/lib/format"
import { imageSampleLabel } from "@/lib/status"

type ProbeState = { status: "checking" } | { status: "pass"; width: number; height: number } | { status: "fail" }

export function ImageProbe({ sample }: Readonly<{ sample: ImageSample }>) {
  const [probe, setProbe] = useState<ProbeState>({ status: "checking" })

  useEffect(() => {
    if (!sample.url) {
      setProbe({ status: sample.decodable ? "pass" : "fail", width: sample.width, height: sample.height })
      return
    }
    const image = new Image()
    image.onload = () => {
      setProbe(image.naturalWidth > 0 && image.naturalHeight > 0
        ? { status: "pass", width: image.naturalWidth, height: image.naturalHeight }
        : { status: "fail" })
    }
    image.onerror = () => setProbe({ status: "fail" })
    image.src = sample.url
    return () => {
      image.onload = null
      image.onerror = null
    }
  }, [sample.decodable, sample.height, sample.url, sample.width])

  const served = probe.status === "pass"
  const label = probe.status === "checking" ? "Checking render" : served ? `${probe.width}×${probe.height} rendered` : "Did not render"

  return (
    <Card className="overflow-hidden py-0">
      <div className="aspect-video bg-muted">
        {sample.url ? (
          <img src={sample.url} alt={sample.name} className="size-full object-contain" />
        ) : (
          <div className="grid size-full place-items-center font-mono text-xs text-muted-foreground">No media URL</div>
        )}
      </div>
      <CardHeader>
        <CardTitle className="truncate">{sample.name}</CardTitle>
        <CardDescription className="font-mono text-[11px]">{formatBytes(sample.size)} · {formatRelativeTime(sample.createdAt)}</CardDescription>
        <CardAction>
          <Badge variant={served ? "success" : probe.status === "checking" ? "neutral" : "error"}>
            {probe.status === "checking" ? "Checking" : served ? "Rendered" : "Failed"}
          </Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="pb-4">
        <p className="font-mono text-[11px] text-muted-foreground">{label} · decode {imageSampleLabel(sample)}</p>
      </CardContent>
    </Card>
  )
}
