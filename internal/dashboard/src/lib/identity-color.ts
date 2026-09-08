const IDENTITY_COLORS = [
  "#ff5a36",
  "#3b82f6",
  "#a78bfa",
  "#f59e0b",
  "#2dd4bf",
  "#f472b6",
  "#38bdf8",
  "#eab308",
  "#fb7185",
  "#818cf8",
  "#4ade80",
  "#fb923c",
  "#c084fc",
  "#22c55e",
] as const

export function identityColor(key: string) {
  let hash = 2166136261
  for (const char of key) {
    hash ^= char.charCodeAt(0)
    hash = Math.imul(hash, 16777619)
  }
  return IDENTITY_COLORS[Math.abs(hash) % IDENTITY_COLORS.length] ?? IDENTITY_COLORS[0]
}

export function uniqueIdentityColor(key: string, peers: string[]) {
  const assigned = new Map<string, string>()
  const ordered = [...new Set(peers)].sort((left, right) => left.localeCompare(right))
  for (const peer of ordered) {
    if (assigned.has(peer)) continue
    const taken = new Set(assigned.values())
    const start = IDENTITY_COLORS.indexOf(identityColor(peer) as typeof IDENTITY_COLORS[number])
    let color = IDENTITY_COLORS[(start + IDENTITY_COLORS.length) % IDENTITY_COLORS.length] ?? IDENTITY_COLORS[0]
    for (let offset = 0; offset < IDENTITY_COLORS.length; offset++) {
      const candidate = IDENTITY_COLORS[(start + offset + IDENTITY_COLORS.length) % IDENTITY_COLORS.length] ?? IDENTITY_COLORS[0]
      if (!taken.has(candidate)) {
        color = candidate
        break
      }
    }
    assigned.set(peer, color)
  }
  return assigned.get(key) ?? identityColor(key)
}

export function identityColorSurface(color: string) {
  return `color-mix(in oklch, ${color} 22%, #161616)`
}

export function identityColorBorder(color: string) {
  return `color-mix(in oklch, ${color} 42%, #2a2a2a)`
}

export function applicationColorKey(application: { projectName: string; name: string }) {
  return `${application.projectName} / ${application.name}`
}

export function applicationAccent(
  application: { projectName: string; name: string },
  peers: Array<{ projectName: string; name: string }> = [application],
) {
  return uniqueIdentityColor(applicationColorKey(application), peers.map(applicationColorKey))
}
