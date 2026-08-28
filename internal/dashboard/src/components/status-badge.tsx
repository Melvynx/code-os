import type { Application } from "@/api/schema"
import { Badge } from "@/components/ui/badge"

export function ApplicationStatusBadge({ application }: Readonly<{ application: Application }>) {
  if (application.state !== "running") {
    return <Badge variant="neutral">{application.state}</Badge>
  }
  if (application.healthy === false) {
    return <Badge variant="error">Unhealthy</Badge>
  }
  if (application.healthy === true) {
    return <Badge variant="success">Healthy</Badge>
  }
  return <Badge variant="neutral">Running</Badge>
}
