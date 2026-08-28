# StackEnv Design System

## Foundations

- Canvas: `#000000`
- Surface: `#111111`
- Border: `#333333`
- Primary text: `#ffffff`
- Muted text: `#888888`
- Action/link: `#0070f3`
- Success: `#50e3c2`
- Error: `#ee0000`
- Sans: Geist
- Monospace: Geist Mono

Use a compact operational density. Prefer dividers and alignment over elevation. Radii stay between 0 and 8px. Do not use gradients, blur, glow, or decorative shadows.

## Components

Build interactive controls from shadcn/ui and Radix primitives. Preserve their keyboard behavior and accessible names while styling them through the tokens above. Buttons and fields have a 44px minimum touch target and a visible blue focus ring.

Use tables or aligned rows for processes and Git state, cards only for meaningful grouping, badges only for statuses, and responsive horizontal scrolling when dense operational data cannot collapse safely.

## Motion

Keep motion functional and short. Honor `prefers-reduced-motion`; the product must remain understandable with motion disabled.
