export interface DisplayUserLike {
  username?: string | null
  email?: string | null
}

export function extractQQNumberFromEmail(email?: string | null): string {
  const normalized = email?.trim().toLowerCase() ?? ''
  const match = normalized.match(/^(\d{5,12})@qq\.com$/)
  return match?.[1] ?? ''
}

export function buildEmailDisplayName(email?: string | null, fallback = ''): string {
  const normalized = email?.trim()
  if (!normalized) return fallback

  const qqNumber = extractQQNumberFromEmail(normalized)
  if (qqNumber) return `QQ ${qqNumber}`

  const [localPart] = normalized.split('@')
  return localPart?.trim() || fallback
}

export function resolveUserDisplayName(user?: DisplayUserLike | null, fallback = ''): string {
  const username = user?.username?.trim()
  if (username) return username
  return buildEmailDisplayName(user?.email, fallback)
}
