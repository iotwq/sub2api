export interface CommunityChatSearchTextSegment {
  text: string
  matched: boolean
}

export interface CommunityChatSearchScrollMetrics {
  containerScrollTop: number
  containerTop: number
  containerHeight: number
  targetTop: number
  targetHeight: number
}

export function splitCommunityChatSearchText(text: string, query: string): CommunityChatSearchTextSegment[] {
  if (!text) return []
  const normalizedQuery = query.trim()
  if (!normalizedQuery) return [{ text, matched: false }]

  const escapedQuery = normalizedQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const matcher = new RegExp(escapedQuery, 'giu')
  const segments: CommunityChatSearchTextSegment[] = []
  let cursor = 0

  for (const match of text.matchAll(matcher)) {
    const index = match.index ?? 0
    if (index > cursor) segments.push({ text: text.slice(cursor, index), matched: false })
    segments.push({ text: match[0], matched: true })
    cursor = index + match[0].length
  }

  if (cursor === 0) return [{ text, matched: false }]
  if (cursor < text.length) segments.push({ text: text.slice(cursor), matched: false })
  return segments
}

export function hasCommunityChatSearchMatch(text: string, query: string): boolean {
  return splitCommunityChatSearchText(text, query).some((segment) => segment.matched)
}

export function centeredCommunityChatSearchScrollTop(metrics: CommunityChatSearchScrollMetrics): number {
  const targetTopInsideContainer = metrics.containerScrollTop + metrics.targetTop - metrics.containerTop
  const centeredTop = targetTopInsideContainer + metrics.targetHeight / 2 - metrics.containerHeight / 2
  return Math.max(0, centeredTop)
}
