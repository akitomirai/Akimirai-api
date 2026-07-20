import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const routerSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../index.ts'),
  'utf8',
)

describe('Model Plaza and model pricing routes', () => {
  it('uses Model Plaza as the canonical user route and redirects the legacy path', () => {
    expect(routerSource).toContain("path: '/model-plaza'")
    expect(routerSource).toContain("import('@/views/user/ModelPlazaView.vue')")
    expect(routerSource).toMatch(/path: '\/available-channels',[\s\S]*?redirect: '\/model-plaza'/)
    expect(routerSource).not.toContain("import('@/views/user/AvailableChannelsView.vue')")
  })

  it('separates the model pricing projection from channel configuration', () => {
    expect(routerSource).toMatch(/path: '\/admin\/channels\/pricing',[\s\S]*?ModelPricingView\.vue/)
    expect(routerSource).toMatch(/path: '\/admin\/channels\/config',[\s\S]*?ChannelsView\.vue/)
    expect(routerSource).toMatch(/path: '\/admin\/channels',[\s\S]*?redirect: '\/admin\/channels\/pricing'/)
  })
})
