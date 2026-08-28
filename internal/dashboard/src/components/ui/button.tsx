import { Slot } from "radix-ui"
import { cva, type VariantProps } from "class-variance-authority"
import type * as React from "react"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex shrink-0 items-center justify-center gap-2 whitespace-nowrap rounded-md border font-mono text-xs font-medium uppercase tracking-wide transition-colors outline-none disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        default: "border-white bg-white text-black hover:bg-[#888]",
        outline: "border-[#333] bg-black text-white hover:border-white hover:bg-[#111]",
        ghost: "border-transparent bg-transparent text-[#888] hover:bg-[#111] hover:text-white",
        link: "border-transparent bg-transparent p-0 text-[#0070f3] hover:underline",
        destructive: "border-[#e00] bg-black text-[#e00] hover:bg-[#111]",
      },
      size: {
        default: "h-9 px-4",
        sm: "h-8 px-3",
        icon: "size-9 p-0",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
)

function Button({
  className,
  variant = "default",
  size = "default",
  asChild = false,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & { asChild?: boolean }) {
  const Component = asChild ? Slot.Root : "button"

  return (
    <Component
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    />
  )
}

export { Button, buttonVariants }
