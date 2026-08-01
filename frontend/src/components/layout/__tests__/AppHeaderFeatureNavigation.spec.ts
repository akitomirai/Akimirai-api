import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../AppHeader.vue'),
  'utf8',
)

describe('AppHeader feature navigation', () => {
  it('projects one check-in owner before Model Plaza and into the mobile menu', () => {
    const desktopCheckIn = source.indexOf('data-testid="daily-check-in-desktop"')
    const modelPlazaButton = source.indexOf('@click="handleModelPlazaClick"')
    const mobileBalance = source.indexOf('<!-- Balance (mobile only) -->')
    const mobileCheckIn = source.indexOf('data-testid="daily-check-in-mobile"')

    expect(desktopCheckIn).toBeGreaterThan(-1)
    expect(desktopCheckIn).toBeLessThan(modelPlazaButton)
    expect(mobileCheckIn).toBeGreaterThan(mobileBalance)
    expect(source.match(/useDailyCheckIn\(/g)).toHaveLength(1)
    expect(source).toContain('hidden xl:inline-flex')
    expect(source).toContain('xl:hidden')
    expect(source.match(/handleCheckInAction/g)?.length).toBeGreaterThanOrEqual(3)
  })

  it('places the authenticated Model Plaza action before GPTImage and gates it with the user catalog flag', () => {
    const modelPlazaButton = source.indexOf('@click="handleModelPlazaClick"')
    const gptImageButton = source.indexOf('@click="handleGptImageClick"')

    expect(modelPlazaButton).toBeGreaterThan(-1)
    expect(gptImageButton).toBeGreaterThan(modelPlazaButton)
    expect(source).toContain('FeatureFlags.availableChannels')
    expect(source).toContain('v-if="user && modelPlazaEnabled"')
  })

  it('uses the shared return-navigation owner for both feature actions', () => {
    expect(source.match(/createFeatureReturnNavigation\(/g)).toHaveLength(2)
    expect(source).toContain("storageKey: 'sub2api:model-plaza-return-path'")
    expect(source).toContain("entryPath: '/available-channels'")
    expect(source).toContain("storageKey: 'sub2api:gpt-image-return-path'")
    expect(source).not.toContain('getStoredGptImageReturnPath')
  })
})
