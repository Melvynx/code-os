import type { ReactNode } from "react"

type SectionHeadingProps = {
  id?: string
  eyebrow?: string
  title: string
  description?: string
  action?: ReactNode
}

export function SectionHeading({ id, eyebrow, title, description, action }: SectionHeadingProps) {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        {eyebrow ? <p className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">{eyebrow}</p> : null}
        <h2 id={id} className="mt-1 text-xl font-medium tracking-tight">{title}</h2>
        {description ? <p className="mt-1 max-w-2xl text-sm leading-relaxed text-muted-foreground">{description}</p> : null}
      </div>
      {action}
    </div>
  )
}
