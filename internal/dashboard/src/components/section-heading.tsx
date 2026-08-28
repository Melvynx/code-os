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
        {eyebrow ? <p className="font-mono text-[10px] uppercase tracking-wider text-[#888]">{eyebrow}</p> : null}
        <h2 id={id} className="mt-1 text-xl font-medium tracking-tight text-white">{title}</h2>
        {description ? <p className="mt-1 max-w-2xl text-sm leading-relaxed text-[#888]">{description}</p> : null}
      </div>
      {action}
    </div>
  )
}
