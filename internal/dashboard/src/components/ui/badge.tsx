import { Slot } from "radix-ui"
import { cva, type VariantProps } from "class-variance-authority"
import type * as React from "react"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex w-fit shrink-0 items-center gap-1 rounded-md border px-2 py-1 font-mono text-[10px] font-medium uppercase tracking-wider",
  {
    variants: {
      variant: {
        neutral: "border-[#333] bg-[#111] text-[#888]",
        success: "border-[#50e3c2]/50 bg-black text-[#50e3c2]",
        warning: "border-[#333] bg-[#111] text-white",
        error: "border-[#e00]/60 bg-black text-[#e00]",
        link: "border-[#0070f3]/60 bg-black text-[#0070f3]",
      },
    },
    defaultVariants: { variant: "neutral" },
  },
)

function Badge({
  className,
  variant,
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
  const Component = asChild ? Slot.Root : "span"
  return (
    <Component
      data-slot="badge"
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  )
}

export { Badge, badgeVariants }
