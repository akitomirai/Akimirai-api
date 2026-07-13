export interface FeatureReturnNavigationOptions {
  storageKey: string
  isFeaturePath: (path: string) => boolean
  entryPath: string
  fallbackPath: () => string
  storage?: Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>
}

export interface FeatureReturnNavigation {
  isActive: (path: string) => boolean
  getToggleTarget: (currentFullPath: string) => string
  clearReturnPath: () => void
}

export function createFeatureReturnNavigation(
  options: FeatureReturnNavigationOptions,
): FeatureReturnNavigation {
  const storage = options.storage ?? getSessionStorage()

  function isActive(path: string): boolean {
    return options.isFeaturePath(pathnameOf(path))
  }

  function clearReturnPath(): void {
    safeRemove(storage, options.storageKey)
  }

  function getToggleTarget(currentFullPath: string): string {
    if (!isActive(currentFullPath)) {
      if (isSafeReturnPath(currentFullPath, options.isFeaturePath)) {
        safeSet(storage, options.storageKey, currentFullPath)
      } else {
        clearReturnPath()
      }
      return options.entryPath
    }

    const stored = safeGet(storage, options.storageKey)
    clearReturnPath()
    return isSafeReturnPath(stored, options.isFeaturePath) ? stored : options.fallbackPath()
  }

  return { isActive, getToggleTarget, clearReturnPath }
}

function safeGet(
  storage: FeatureReturnNavigationOptions['storage'] | null,
  key: string,
): string | null {
  try {
    return storage?.getItem(key) ?? null
  } catch {
    return null
  }
}

function safeSet(
  storage: FeatureReturnNavigationOptions['storage'] | null,
  key: string,
  value: string,
): void {
  try {
    storage?.setItem(key, value)
  } catch {
    // Navigation must still work when storage is unavailable or quota-limited.
  }
}

function safeRemove(
  storage: FeatureReturnNavigationOptions['storage'] | null,
  key: string,
): void {
  try {
    storage?.removeItem(key)
  } catch {
    // Best-effort cleanup only.
  }
}

export function isSafeFeatureReturnPath(
  value: string | null | undefined,
  isFeaturePath: (path: string) => boolean,
): value is string {
  return isSafeReturnPath(value, isFeaturePath)
}

function isSafeReturnPath(
  value: string | null | undefined,
  isFeaturePath: (path: string) => boolean,
): value is string {
  if (!value || !value.startsWith('/') || value.startsWith('//')) return false
  const path = pathnameOf(value)
  return path.startsWith('/') && !isFeaturePath(path)
}

function pathnameOf(value: string): string {
  const queryIndex = value.search(/[?#]/)
  return queryIndex === -1 ? value : value.slice(0, queryIndex)
}

function getSessionStorage(): Storage | null {
  if (typeof window === 'undefined') return null
  try {
    return window.sessionStorage
  } catch {
    return null
  }
}
