export interface UsageTokenMetrics {
  input_tokens?: number | null
  output_tokens?: number | null
  cache_creation_tokens?: number | null
  cache_read_tokens?: number | null
}

const toFiniteNumber = (value: number | null | undefined): number =>
  typeof value === 'number' && Number.isFinite(value) ? value : 0

export const hasKnownTokenMetrics = (metrics: UsageTokenMetrics): boolean =>
  [
    metrics.input_tokens,
    metrics.output_tokens,
    metrics.cache_creation_tokens,
    metrics.cache_read_tokens,
  ].some((value) => value != null)

export const totalUsageTokens = (metrics: UsageTokenMetrics): number =>
  toFiniteNumber(metrics.input_tokens)
  + toFiniteNumber(metrics.output_tokens)
  + toFiniteNumber(metrics.cache_creation_tokens)
  + toFiniteNumber(metrics.cache_read_tokens)

export const cacheHitRatio = (metrics: UsageTokenMetrics): number | null => {
  const input = toFiniteNumber(metrics.input_tokens)
  const cacheRead = toFiniteNumber(metrics.cache_read_tokens)
  const cacheCreation = toFiniteNumber(metrics.cache_creation_tokens)
  const total = input + cacheRead + cacheCreation

  if (total <= 0) return null
  return Number(((cacheRead / total) * 100).toFixed(1))
}

export const cacheHitProgressWidth = (metrics: UsageTokenMetrics): string => {
  const ratio = cacheHitRatio(metrics) ?? 0
  return `${Math.min(100, Math.max(0, ratio))}%`
}

export const formatUsageDuration = (ms: number | null | undefined): string => {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(2)}s`

  const totalSeconds = Math.round(ms / 1000)
  if (totalSeconds < 3600) {
    return `${Math.floor(totalSeconds / 60)}m ${totalSeconds % 60}s`
  }

  return `${Math.floor(totalSeconds / 3600)}h ${Math.floor((totalSeconds % 3600) / 60)}m`
}
