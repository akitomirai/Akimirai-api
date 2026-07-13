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

describe('AppSidebar scroll position persistence', () => {
  it('binds a template ref to the sidebar nav element', () => {
    expect(componentSource).toContain('ref="sidebarNavRef"')
    expect(componentSource).toContain('sidebar-nav')
  })

  it('declares sidebarNavRef in script setup', () => {
    expect(componentSource).toContain("const sidebarNavRef = ref<HTMLElement | null>(null)")
  })

  it('saves scroll position on beforeUnmount', () => {
    expect(componentSource).toContain('onBeforeUnmount')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('sidebarNavRef.value.scrollTop')
  })

  it('restores scroll position on mount', () => {
    expect(componentSource).toContain('onMounted')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('nextTick')
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
  it('keeps header-only features out of the sidebar and consolidates orders into the store', () => {
    expect(componentSource).not.toContain("path: '/images'")
    expect(componentSource).not.toContain("path: '/image-management'")
    expect(componentSource).not.toContain("path: '/model-plaza'")
    expect(componentSource).not.toContain("path: '/available-channels'")

    const purchaseItem = componentSource.match(/\{ path: '\/purchase'[\s\S]*?\}/)?.[0] ?? ''
    const ordersItem = componentSource.match(/\{ path: '\/orders'[\s\S]*?\}/)?.[0] ?? ''

    expect(purchaseItem).toContain("label: t('nav.buySubscription')")
    expect(ordersItem).toBe('')
    expect(purchaseItem).not.toContain('featureFlag')
  })

  it('removes channel status from personal navigation while keeping the admin monitor', () => {
    expect(componentSource).not.toContain("{ path: '/monitor'")
    expect(componentSource).toContain("{ path: '/admin/channels/monitor'")
  })

  it('separates model pricing from channel configuration', () => {
    expect(componentSource).toContain("path: '/admin/channels/pricing'")
    expect(componentSource).toContain("label: t('nav.modelPricing')")
    expect(componentSource).toContain("path: '/admin/channels/config'")
    expect(componentSource).toContain("label: t('nav.channelConfiguration')")
  })

  it('removes the user subscription entry while keeping subscription management', () => {
    expect(componentSource).not.toContain("{ path: '/subscriptions'")
    expect(componentSource).toContain("{ path: '/admin/subscriptions'")
  })
})
