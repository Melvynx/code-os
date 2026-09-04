import type { LucideIcon } from "lucide-react"

import { Card, CardAction, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

type MetricCardProps = {
  icon: LucideIcon
  label: string
  value: number
  hint: string
}

export function MetricCard({ icon: Icon, label, value, hint }: MetricCardProps) {
  return (
    <Card className="transition-colors hover:ring-foreground/20">
      <CardHeader>
        <CardDescription className="font-mono text-[10px] uppercase tracking-wider">{label}</CardDescription>
        <CardTitle className="font-mono text-3xl tabular-nums">{value}</CardTitle>
        <CardDescription>{hint}</CardDescription>
        <CardAction>
          <Icon aria-hidden="true" className="text-muted-foreground" />
        </CardAction>
      </CardHeader>
    </Card>
  )
}
