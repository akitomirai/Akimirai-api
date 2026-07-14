import { afterEach, describe, expect, it } from 'vitest'

import {
  TOKEN_COUNT_MODE_STORAGE_KEY,
  getTokenCountMode,
  resetTokenCountModeForTests,
  setTokenCountMode,
  useTokenCountMode,
} from '@/composables/useTokenCountMode'

describe('useTokenCountMode', () => {
  afterEach(() => {
    localStorage.clear()
    resetTokenCountModeForTests()
  })

  it('defaults invalid or missing storage to modern', () => {
    expect(getTokenCountMode()).toBe('modern')

    resetTokenCountModeForTests()
    localStorage.setItem(TOKEN_COUNT_MODE_STORAGE_KEY, 'unsupported')
    expect(getTokenCountMode()).toBe('modern')
  })

  it('persists changes and exposes one reactive mode', () => {
    const first = useTokenCountMode()
    const second = useTokenCountMode()

    setTokenCountMode('legacy')

    expect(first.mode.value).toBe('legacy')
    expect(second.mode.value).toBe('legacy')
    expect(localStorage.getItem(TOKEN_COUNT_MODE_STORAGE_KEY)).toBe('legacy')
  })

  it('restores a persisted mode after module state is reset', () => {
    localStorage.setItem(TOKEN_COUNT_MODE_STORAGE_KEY, 'legacy')
    resetTokenCountModeForTests()

    expect(getTokenCountMode()).toBe('legacy')
  })
})
