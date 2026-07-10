import { describe, expect, it } from 'vitest'
import en from '@/i18n/locales/en/dashboard'
import zh from '@/i18n/locales/zh/dashboard'
import { DEFAULT_PRIVACY_FILTER_TYPES } from '@/utils/privacyFilter'

describe('profile privacy locale keys', () => {
  it.each([en, zh])('defines avatar consent and privacy filter labels', (locale) => {
    expect(locale.profile.avatar.qqCheckAction).toBeTruthy()
    expect(locale.profile.avatar.qqConsentHint).toBeTruthy()
    expect(locale.profile.privacyFilter.title).toBeTruthy()
    expect(locale.profile.privacyFilter.description).toBeTruthy()
    expect(locale.profile.privacyFilter.noRetention).toBeTruthy()

    for (const type of DEFAULT_PRIVACY_FILTER_TYPES) {
      expect(locale.profile.privacyFilter.types[type]).toBeTruthy()
      expect(locale.profile.privacyFilter.hints[type]).toBeTruthy()
    }
  })
})
