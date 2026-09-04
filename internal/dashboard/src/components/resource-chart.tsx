import { useMemo, useState, type CSSProperties } from "react"
import { CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts"

import { useResourceHistory } from "@/api/queries"
import type { ResourceSeries } from "@/api/resources"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { EmptyState } from "@/components/page-state"
import { formatBytes } from "@/lib/format"
import { formatChartTime, formatHistoryWindow, formatMemoryAxis, groupResourceSeries, resourceColor, toChartRows } from "@/lib/resources"

export function ResourceChart() {
  const history = useResourceHistory()
  const [hidden, setHidden] = useState<Record<string, boolean>>({})

  const series = useMemo(() => groupResourceSeries(history.data?.series ?? []), [history.data?.series])
  const visibleIds = series.filter((item) => !hidden[item.id]).map((item) => item.id)
  const visible = series.filter((item) => !hidden[item.id])
  const config = useMemo(() => {
    const next: ChartConfig = {}
    for (const item of series) next[item.id] = { label: item.name, color: resourceColor(item, series) }
    return next
  }, [series])
  const groupedHistory = history.data ? { ...history.data, series } : undefined
  const cpuRows = groupedHistory ? toChartRows(groupedHistory, "cpuPercent") : []
  const memoryRows = groupedHistory ? toChartRows(groupedHistory, "memoryBytes") : []
  const sampleCount = history.data?.sampleCount ?? 0

  return (
    <Card className="overflow-visible">
      <CardHeader className="border-b">
        <CardDescription className="font-mono text-[10px] uppercase tracking-wider">History</CardDescription>
        <CardTitle>CPU and memory over time</CardTitle>
        <CardDescription>{history.data ? formatHistoryWindow(history.data) : "Daemon samples every refresh and keeps the last 6 hours."}</CardDescription>
      </CardHeader>
      {series.length ? (
        <CardContent className="border-b py-3">
          <ToggleGroup
            type="multiple"
            variant="outline"
            size="sm"
            value={visibleIds}
            onValueChange={(ids) => {
              setHidden(Object.fromEntries(series.map((item) => [item.id, !ids.includes(item.id)])))
            }}
            className="flex w-full flex-wrap justify-start"
            spacing={2}
          >
            {series.map((item) => (
              <ToggleGroupItem key={item.id} value={item.id} aria-label={item.count > 1 ? `${item.name} (${item.count})` : item.name}>
                <span aria-hidden="true" className="size-1.5" style={{ backgroundColor: resourceColor(item, series) }} />
                {item.name}
                {item.count > 1 ? <Badge variant="neutral">{item.count}</Badge> : null}
              </ToggleGroupItem>
            ))}
          </ToggleGroup>
        </CardContent>
      ) : null}
      <CardContent>
        {history.isError ? (
          <Alert variant="destructive">
            <AlertTitle>Unable to load resource history</AlertTitle>
            <AlertDescription>{history.error.message}</AlertDescription>
          </Alert>
        ) : !series.length || sampleCount === 0 ? (
          <EmptyState title="Collecting samples" description="The next daemon refresh will start the graph." />
        ) : (
          <div className="grid gap-6 xl:grid-cols-2">
            <MetricChart title="CPU" rows={cpuRows} series={visible} config={config} formatter={(value) => `${value.toFixed(1)}%`} tickFormatter={(value) => `${value}%`} />
            <MetricChart title="Memory" rows={memoryRows} series={visible} config={config} formatter={formatBytes} tickFormatter={formatMemoryAxis} />
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function MetricChart({
  title,
  rows,
  series,
  config,
  formatter,
  tickFormatter,
}: Readonly<{
  title: string
  rows: Array<Record<string, number | string>>
  series: ResourceSeries[]
  config: ChartConfig
  formatter: (value: number) => string
  tickFormatter: (value: number) => string
}>) {
  return (
    <div>
      <p className="mb-3 font-mono text-[10px] uppercase tracking-wider text-muted-foreground">{title}</p>
      <ChartContainer config={config} className="aspect-auto h-56 w-full">
        <LineChart accessibilityLayer data={rows} margin={{ top: 12, right: 12, left: 8, bottom: 0 }}>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="at" tickLine={false} axisLine={false} minTickGap={28} tickFormatter={formatChartTime} />
          <YAxis tickLine={false} axisLine={false} width={52} tickFormatter={tickFormatter} />
          <ChartTooltip
            cursor={{ stroke: "var(--border)", strokeWidth: 1 }}
            isAnimationActive={false}
            itemSorter={(item) => -Number(item.value)}
            content={
              <ChartTooltipContent
                className="min-w-56"
                labelFormatter={(value) => formatChartTime(String(value ?? ""))}
                formatter={(value, name, item) => (
                  <>
                    <span
                      className="size-2.5 shrink-0 rounded-[2px] border-(--color-border) bg-(--color-bg)"
                      style={{ "--color-bg": item.color, "--color-border": item.color } as CSSProperties}
                    />
                    <span className="text-muted-foreground">{config[String(item.dataKey ?? name)]?.label ?? String(name)}</span>
                    <span className="ml-auto font-mono font-medium tabular-nums">{formatter(Number(value))}</span>
                  </>
                )}
              />
            }
          />
          {series.map((item) => (
            <Line
              key={item.id}
              type="monotone"
              dataKey={item.id}
              name={item.name}
              stroke={config[item.id]?.color}
              strokeWidth={item.kind === "agent" ? 2 : 1.75}
              dot={false}
              activeDot={{ r: 4 }}
              isAnimationActive={false}
            />
          ))}
        </LineChart>
      </ChartContainer>
    </div>
  )
}
