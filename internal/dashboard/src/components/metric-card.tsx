import type { LucideIcon } from "lucide-react"

type MetricCardProps = {
  icon: LucideIcon
  label: string
  value: number
  hint: string
}

export function MetricCard({ icon: Icon, label, value, hint }: MetricCardProps) {
  return (
    <article className="rounded-lg border border-[#333] bg-black p-5 transition-colors hover:border-white">
      <div className="flex items-center justify-between">
        <p className="font-mono text-[10px] uppercase tracking-wider text-[#888]">{label}</p>
        <Icon aria-hidden="true" className="size-4 text-[#888]" />
      </div>
      <p className="mt-5 font-mono text-3xl tabular-nums text-white">{value}</p>
      <p className="mt-1 text-xs text-[#888]">{hint}</p>
    </article>
  )
}
