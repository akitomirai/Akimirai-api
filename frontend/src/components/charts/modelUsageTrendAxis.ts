import type { ModelUsageTrendGranularity } from '@/api/usage'

const LEGACY_BUCKET_PATTERN = /^(\d{4})-(\d{2})-(\d{2})(?:[ T](\d{2}):(\d{2}))?$/
const WALL_CLOCK_PREFIX_PATTERN = /^(\d{4})-(\d{2})-(\d{2})(?:[ T](\d{2}):(\d{2}))?/
const RFC3339_BUCKET_PATTERN = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(?::\d{2}(?:\.\d+)?)?(?:Z|[+-]\d{2}:\d{2})$/i
const HOUR_MS = 60 * 60 * 1000
const MAX_DENSE_INTERVALS = 1000

interface BucketParts {
  year: string
  month: string
  day: string
  hour: string
  minute: string
  timezoneName: string
}

const parseWallClockParts = (bucket: string): BucketParts | null => {
  const match = WALL_CLOCK_PREFIX_PATTERN.exec(bucket)
  if (!match) return null

  return {
    year: match[1],
    month: match[2],
    day: match[3],
    hour: match[4] ?? '00',
    minute: match[5] ?? '00',
    timezoneName: '',
  }
}

const toUtcTimestamp = (bucket: BucketParts): number => Date.UTC(
  Number(bucket.year),
  Number(bucket.month) - 1,
  Number(bucket.day),
  Number(bucket.hour),
  Number(bucket.minute),
)

const isRFC3339Bucket = (bucket: string): boolean => RFC3339_BUCKET_PATTERN.test(bucket)

const parseBucketTimestamp = (bucket: string): number | null => {
  if (isRFC3339Bucket(bucket)) {
    const timestamp = Date.parse(bucket)
    return Number.isFinite(timestamp) ? timestamp : null
  }

  const legacy = LEGACY_BUCKET_PATTERN.exec(bucket)
  if (!legacy) return null
  return Date.UTC(
    Number(legacy[1]),
    Number(legacy[2]) - 1,
    Number(legacy[3]),
    Number(legacy[4] ?? '0'),
    Number(legacy[5] ?? '0'),
  )
}

const formatLegacyBucket = (
  timestamp: number,
  granularity: ModelUsageTrendGranularity,
): string => {
  const date = new Date(timestamp)
  const year = date.getUTCFullYear()
  const month = String(date.getUTCMonth() + 1).padStart(2, '0')
  const day = String(date.getUTCDate()).padStart(2, '0')
  if (granularity === 'day') return `${year}-${month}-${day}`
  const hour = String(date.getUTCHours()).padStart(2, '0')
  const minute = String(date.getUTCMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hour}:${minute}`
}

const formatRFC3339Bucket = (timestamp: number): string =>
  new Date(timestamp).toISOString().replace('.000Z', 'Z')

const granularityHours: Record<ModelUsageTrendGranularity, number> = {
  hour: 1,
  '2h': 2,
  '4h': 4,
  '8h': 8,
  day: 24,
}

const sortBuckets = (buckets: string[]): string[] => Array.from(new Set(buckets)).sort((a, b) => {
  const aTimestamp = parseBucketTimestamp(a)
  const bTimestamp = parseBucketTimestamp(b)
  if (aTimestamp !== null && bTimestamp !== null) return aTimestamp - bTimestamp
  return a.localeCompare(b)
})

export const completeModelUsageBuckets = (
  buckets: string[],
  granularity: ModelUsageTrendGranularity,
  rangeStart?: string | null,
  rangeEnd?: string | null,
): string[] => {
  const orderedBuckets = sortBuckets(buckets)
  const usesInstantBuckets = orderedBuckets.some(isRFC3339Bucket)

  if (rangeStart && rangeEnd) {
    const step = granularityHours[granularity] * HOUR_MS
    const legacyStart = usesInstantBuckets ? null : parseWallClockParts(rangeStart)
    const legacyEnd = usesInstantBuckets ? null : parseWallClockParts(rangeEnd)
    const startTimestamp = usesInstantBuckets
      ? parseBucketTimestamp(rangeStart)
      : legacyStart ? toUtcTimestamp(legacyStart) : null
    const endTimestamp = usesInstantBuckets
      ? parseBucketTimestamp(rangeEnd)
      : legacyEnd ? toUtcTimestamp(legacyEnd) : null

    if (startTimestamp !== null && endTimestamp !== null) {
      const bucketCount = Math.ceil((endTimestamp - startTimestamp) / step)
      if (bucketCount > 0 && bucketCount <= MAX_DENSE_INTERVALS) {
        const originalsByTimestamp = new Map<number, string>()
        for (const bucket of orderedBuckets) {
          const timestamp = parseBucketTimestamp(bucket)
          if (timestamp !== null) originalsByTimestamp.set(timestamp, bucket)
        }
        return Array.from({ length: bucketCount }, (_, index) => {
          const timestamp = startTimestamp + index * step
          return originalsByTimestamp.get(timestamp)
            ?? (usesInstantBuckets
              ? formatRFC3339Bucket(timestamp)
              : formatLegacyBucket(timestamp, granularity))
        })
      }
    }
  }

  if (granularity !== 'hour' || orderedBuckets.length < 2) return orderedBuckets

  const firstTimestamp = parseBucketTimestamp(orderedBuckets[0])
  const lastTimestamp = parseBucketTimestamp(orderedBuckets[orderedBuckets.length - 1])
  if (firstTimestamp === null || lastTimestamp === null) return orderedBuckets

  const intervalCount = (lastTimestamp - firstTimestamp) / HOUR_MS
  if (!Number.isInteger(intervalCount) || intervalCount < 1 || intervalCount > 48) {
    return orderedBuckets
  }

  const originalsByTimestamp = new Map<number, string>()
  for (const bucket of orderedBuckets) {
    const timestamp = parseBucketTimestamp(bucket)
    if (timestamp !== null) originalsByTimestamp.set(timestamp, bucket)
  }
  return Array.from({ length: intervalCount + 1 }, (_, index) => {
    const timestamp = firstTimestamp + index * HOUR_MS
    return originalsByTimestamp.get(timestamp)
      ?? (usesInstantBuckets
        ? formatRFC3339Bucket(timestamp)
        : formatLegacyBucket(timestamp, granularity))
  })
}

const resolveTimeZone = (timeZone?: string): string | undefined => {
  if (timeZone) return timeZone
  if (typeof Intl === 'undefined') return undefined
  return Intl.DateTimeFormat().resolvedOptions().timeZone
}

const formatInstantParts = (bucket: string, timeZone?: string): BucketParts | null => {
  const timestamp = parseBucketTimestamp(bucket)
  if (timestamp === null) return null

  try {
    const formatter = new Intl.DateTimeFormat('en-US', {
      timeZone: resolveTimeZone(timeZone),
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hourCycle: 'h23',
      timeZoneName: 'shortOffset',
    })
    const parts = formatter.formatToParts(new Date(timestamp))
    const read = (type: Intl.DateTimeFormatPartTypes): string =>
      parts.find((part) => part.type === type)?.value ?? ''
    return {
      year: read('year'),
      month: read('month'),
      day: read('day'),
      hour: read('hour'),
      minute: read('minute'),
      timezoneName: read('timeZoneName'),
    }
  } catch {
    return null
  }
}

const parseDisplayBucket = (bucket: string, timeZone?: string): BucketParts | null =>
  isRFC3339Bucket(bucket) ? formatInstantParts(bucket, timeZone) : parseWallClockParts(bucket)

export const createModelUsageTickFormatter = (
  buckets: string[],
  granularity: ModelUsageTrendGranularity,
  timeZone?: string,
): ((index: number) => string) => {
  const parsedBuckets = buckets.map((bucket) => parseDisplayBucket(bucket, timeZone))
  const validBuckets = parsedBuckets.filter((bucket): bucket is BucketParts => bucket !== null)
  const isSameDay = validBuckets.length > 0
    && validBuckets.every((bucket) => `${bucket.year}-${bucket.month}-${bucket.day}` === `${validBuckets[0].year}-${validBuckets[0].month}-${validBuckets[0].day}`)
  const isSameYear = validBuckets.length > 0
    && validBuckets.every((bucket) => bucket.year === validBuckets[0].year)
  const wallClockCounts = new Map<string, number>()
  for (const bucket of validBuckets) {
    const wallClock = `${bucket.year}-${bucket.month}-${bucket.day} ${bucket.hour}:${bucket.minute}`
    wallClockCounts.set(wallClock, (wallClockCounts.get(wallClock) ?? 0) + 1)
  }

  return (index: number): string => {
    const original = buckets[index] ?? ''
    const bucket = parsedBuckets[index]
    if (!bucket) return original

    if (granularity !== 'day') {
      const time = `${bucket.hour}:${bucket.minute}`
      const wallClock = `${bucket.year}-${bucket.month}-${bucket.day} ${time}`
      const timezoneSuffix = (wallClockCounts.get(wallClock) ?? 0) > 1 && bucket.timezoneName
        ? ` ${bucket.timezoneName}`
        : ''
      return `${isSameDay ? time : `${bucket.month}-${bucket.day} ${time}`}${timezoneSuffix}`
    }

    return isSameYear
      ? `${bucket.month}-${bucket.day}`
      : `${bucket.year}-${bucket.month}-${bucket.day}`
  }
}
