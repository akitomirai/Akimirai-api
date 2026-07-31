import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const routerSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../index.ts'),
  'utf8',
)

describe('Model Plaza and model pricing routes', () => {
  it('keeps the public plaza separate from authenticated available channels', () => {
    expect(routerSource.match(/path: '\/model-plaza'/g)).toHaveLength(1)
    expect(routerSource).toMatch(
      /path: '\/model-plaza',[\s\S]*?import\('@\/views\/ModelPlazaView\.vue'\)[\s\S]*?requiresAuth: false/,
    )
    expect(routerSource).toMatch(
      /path: '\/available-channels',[\s\S]*?import\('@\/views\/user\/ModelPlazaView\.vue'\)[\s\S]*?requiresAuth: true/,
    )
    expect(routerSource).not.toMatch(/path: '\/available-channels',[\s\S]*?redirect: '\/model-plaza'/)
  })

  it('separates the model pricing projection from channel configuration', () => {
    expect(routerSource).toMatch(/path: '\/admin\/channels\/pricing',[\s\S]*?ModelPricingView\.vue/)
    expect(routerSource).toMatch(/path: '\/admin\/channels\/config',[\s\S]*?ChannelsView\.vue/)
    expect(routerSource).toMatch(/path: '\/admin\/channels',[\s\S]*?redirect: '\/admin\/channels\/pricing'/)
  })
})
