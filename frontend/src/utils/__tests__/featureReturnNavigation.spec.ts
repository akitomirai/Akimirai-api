import { describe, expect, it } from 'vitest'
import { createFeatureReturnNavigation, isSafeFeatureReturnPath } from '../featureReturnNavigation'

function createStorage(): Storage {
  const values = new Map<string, string>()
  return {
    get length() { return values.size },
    clear: () => values.clear(),
    getItem: (key) => values.get(key) ?? null,
    key: (index) => Array.from(values.keys())[index] ?? null,
    removeItem: (key) => { values.delete(key) },
    setItem: (key, value) => { values.set(key, value) },
  }
}

describe('createFeatureReturnNavigation', () => {
  it('preserves the full local return path and toggles back from the feature', () => {
    const storage = createStorage()
    const navigation = createFeatureReturnNavigation({
      storageKey: 'model-plaza-return',
      isFeaturePath: (path) => path === '/model-plaza',
      entryPath: '/model-plaza',
      fallbackPath: () => '/dashboard',
      storage,
    })

    expect(navigation.getToggleTarget('/usage?period=24h#requests')).toBe('/model-plaza')
    expect(storage.getItem('model-plaza-return')).toBe('/usage?period=24h#requests')
    expect(navigation.getToggleTarget('/model-plaza?model=gpt-5')).toBe('/usage?period=24h#requests')
    expect(storage.getItem('model-plaza-return')).toBeNull()
  })

  it('rejects cross-origin, protocol-relative, and feature-internal return paths', () => {
    const storage = createStorage()
    storage.setItem('model-plaza-return', '//example.com/escape')
    const navigation = createFeatureReturnNavigation({
      storageKey: 'model-plaza-return',
      isFeaturePath: (path) => path.startsWith('/model-plaza'),
      entryPath: '/model-plaza',
      fallbackPath: () => '/dashboard',
      storage,
    })

    expect(navigation.getToggleTarget('/model-plaza')).toBe('/dashboard')
    expect(isSafeFeatureReturnPath('https://example.com', () => false)).toBe(false)
    expect(isSafeFeatureReturnPath('/model-plaza?model=x', (path) => path === '/model-plaza')).toBe(false)
  })

  it('keeps toggling when browser storage throws', () => {
    const storage = {
      getItem: () => { throw new Error('blocked') },
      setItem: () => { throw new Error('quota') },
      removeItem: () => { throw new Error('blocked') },
    }
    const navigation = createFeatureReturnNavigation({
      storageKey: 'model-plaza-return',
      isFeaturePath: (path) => path === '/model-plaza',
      entryPath: '/model-plaza',
      fallbackPath: () => '/dashboard',
      storage,
    })

    expect(navigation.getToggleTarget('/usage?period=24h')).toBe('/model-plaza')
    expect(navigation.getToggleTarget('/model-plaza')).toBe('/dashboard')
    expect(() => navigation.clearReturnPath()).not.toThrow()
  })
})
