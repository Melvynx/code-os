import type { StatusCheck } from "@/api/status"
import { Badge } from "@/components/ui/badge"
import { Item, ItemActions, ItemContent, ItemDescription, ItemGroup, ItemMedia, ItemTitle } from "@/components/ui/item"
import { cn } from "@/lib/utils"

const TONE = {
  pass: { badge: "success" as const, dot: "bg-success", label: "Pass" },
  warn: { badge: "warning" as const, dot: "bg-warning", label: "Warning" },
  fail: { badge: "error" as const, dot: "bg-destructive", label: "Fail" },
}

export function CheckList({ checks }: Readonly<{ checks: StatusCheck[] }>) {
  return (
    <ItemGroup>
      {checks.map((check) => {
        const tone = TONE[check.status]
        return (
          <Item key={`${check.group}-${check.id}-${check.label}`} variant="outline">
            <ItemMedia>
              <span aria-hidden="true" className={cn("size-2", tone.dot)} />
            </ItemMedia>
            <ItemContent>
              <ItemTitle>{check.label}</ItemTitle>
              <ItemDescription className="break-all font-mono">{check.detail}</ItemDescription>
            </ItemContent>
            <ItemActions>
              <Badge variant={tone.badge}>{tone.label}</Badge>
            </ItemActions>
          </Item>
        )
      })}
    </ItemGroup>
  )
}
