import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('daily check-in locale contract', () => {
  it('keeps the complete runtime key set mirrored in Chinese and English', () => {
    const expectedKeys = [
      'available',
      'availableTitle',
      'loading',
      'claiming',
      'claimed',
      'claimedTitle',
      'error',
      'errorTitle',
      'success',
      'profileRefreshWarning',
    ]

    expect(Object.keys(zh.checkIn).sort()).toEqual(expectedKeys.slice().sort())
    expect(Object.keys(en.checkIn).sort()).toEqual(expectedKeys.slice().sort())
  })
})
