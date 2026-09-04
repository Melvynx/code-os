import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export function FactCard({ label, value, hint }: Readonly<{ label: string; value: string; hint?: string }>) {
  return (
    <Card size="sm">
      <CardHeader>
        <CardDescription className="font-mono text-[10px] uppercase tracking-wider">{label}</CardDescription>
        <CardTitle className="break-all text-sm">{value}</CardTitle>
        {hint ? <CardDescription className="font-mono text-[11px]">{hint}</CardDescription> : null}
      </CardHeader>
    </Card>
  )
}
