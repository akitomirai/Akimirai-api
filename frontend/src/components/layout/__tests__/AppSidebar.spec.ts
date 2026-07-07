import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar navigation visibility', () => {
  it('keeps GPTImage out of the sidebar while preserving store and order entries', () => {
    expect(componentSource).not.toContain("path: '/images'")
    expect(componentSource).not.toContain("path: '/image-management'")

    const purchaseItem = componentSource.match(/\{ path: '\/purchase'[\s\S]*?\}/)?.[0] ?? ''
    const ordersItem = componentSource.match(/\{ path: '\/orders'[\s\S]*?\}/)?.[0] ?? ''

    expect(purchaseItem).toContain("label: t('nav.buySubscription')")
    expect(ordersItem).toContain("label: t('nav.myOrders')")
    expect(purchaseItem).not.toContain('featureFlag')
    expect(ordersItem).not.toContain('featureFlag')
  })
})
