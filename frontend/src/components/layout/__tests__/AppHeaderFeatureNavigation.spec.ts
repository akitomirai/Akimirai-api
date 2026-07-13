import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../AppHeader.vue'),
  'utf8',
)

describe('AppHeader feature navigation', () => {
  it('places the Model Plaza action before GPTImage and gates it with the catalog flag', () => {
    const modelPlazaButton = source.indexOf('@click="handleModelPlazaClick"')
    const gptImageButton = source.indexOf('@click="handleGptImageClick"')

    expect(modelPlazaButton).toBeGreaterThan(-1)
    expect(gptImageButton).toBeGreaterThan(modelPlazaButton)
    expect(source).toContain('FeatureFlags.modelPlaza')
    expect(source).toContain('v-if="user && modelPlazaEnabled"')
  })

  it('uses the shared return-navigation owner for both feature actions', () => {
    expect(source.match(/createFeatureReturnNavigation\(/g)).toHaveLength(2)
    expect(source).toContain("storageKey: 'sub2api:model-plaza-return-path'")
    expect(source).toContain("storageKey: 'sub2api:gpt-image-return-path'")
    expect(source).not.toContain('getStoredGptImageReturnPath')
  })
})
