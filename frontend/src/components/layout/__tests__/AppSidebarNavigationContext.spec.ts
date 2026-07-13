import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  route: {
    path: '/admin/settings',
    fullPath: '/admin/settings',
    name: 'AdminSettings',
    params: {} as Record<string, string>,
    meta: { requiresAdmin: true } as Record<string, unknown>,
  },
  routerPush: vi.fn(),
  appStore: {
    sidebarCollapsed: false,
    sidebarScrollTop: 0,
    mobileOpen: false,
    backendModeEnabled: false,
    siteName: 'Test',
    siteLogo: '',
    siteVersion: 'test',
    publicSettingsLoaded: true,
    cachedPublicSettings: { custom_menu_items: [] },
    toggleSidebar: vi.fn(),
    setMobileOpen: vi.fn(),
  },
  authStore: {
    isAdmin: true,
    isSimpleMode: true,
  },
  onboardingStore: {
    isCurrentStep: vi.fn(() => false),
    nextStep: vi.fn(),
  },
  adminSettingsStore: {
    opsMonitoringEnabled: true,
    paymentEnabled: true,
    customMenuItems: [] as Array<{
      id: string
      label: string
      visibility: 'admin' | 'user'
      sort_order: number
      icon_svg?: string
    }>,
    fetch: vi.fn(),
  },
}))

vi.mock('vue-router', () => ({
  useRoute: () => mocks.route,
  useRouter: () => ({ push: mocks.routerPush }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => mocks.appStore,
  useAuthStore: () => mocks.authStore,
  useOnboardingStore: () => mocks.onboardingStore,
  useAdminSettingsStore: () => mocks.adminSettingsStore,
}))

vi.mock('@/utils/featureFlags', () => ({
  FeatureFlags: {
    affiliate: {},
    modelPlaza: {},
    channelMonitor: {},
    riskControl: {},
  },
  makeSidebarFlag: () => () => true,
}))

vi.mock('@/composables/useBatchImageAccess', () => ({
  useBatchImageAccess: () => ({
    canUseBatchImage: { value: true },
    refreshBatchImageAccess: vi.fn(),
  }),
}))

import AppSidebar from '../AppSidebar.vue'

const RouterLinkStub = {
  props: ['to'],
  template: '<a :data-to="to"><slot /></a>',
}

function mountSidebar() {
  return mount(AppSidebar, {
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        VersionBadge: true,
      },
    },
  })
}

function setRoute(path: string, name: string, requiresAdmin: boolean) {
  mocks.route.path = path
  mocks.route.fullPath = path
  mocks.route.name = name
  mocks.route.meta = { requiresAdmin }
}

describe('AppSidebar admin and personal navigation', () => {
  beforeEach(() => {
    mocks.authStore.isAdmin = true
    mocks.authStore.isSimpleMode = false
    mocks.appStore.sidebarCollapsed = false
    mocks.appStore.sidebarScrollTop = 0
    mocks.routerPush.mockReset()
    mocks.adminSettingsStore.fetch.mockReset()
    mocks.adminSettingsStore.customMenuItems.length = 0
    mocks.route.params = {}
  })

  it('shows the full personal menu below the admin menu', () => {
    setRoute('/admin/settings', 'AdminSettings', true)
    const wrapper = mountSidebar()
    const sections = wrapper.findAll('nav.sidebar-nav > .sidebar-section')

    expect(sections).toHaveLength(2)
    expect(sections[0].find('[data-to="/admin/dashboard"]').exists()).toBe(true)
    expect(sections[0].find('[data-to="/admin/settings"]').exists()).toBe(true)
    expect(sections[0].find('[data-to="/keys"]').exists()).toBe(false)
    expect(sections[1].get('.sidebar-section-title').text()).toContain('nav.myAccount')
    expect(sections[1].find('[data-to="/dashboard"]').exists()).toBe(true)
    expect(sections[1].find('[data-to="/keys"]').exists()).toBe(true)
    expect(sections[1].find('[data-to="/usage"]').exists()).toBe(true)
    expect(sections[1].find('[data-to="/available-channels"]').exists()).toBe(false)
    expect(sections[1].find('[data-to="/model-plaza"]').exists()).toBe(false)
    expect(sections[1].find('[data-to="/purchase"]').exists()).toBe(true)
    expect(sections[1].find('[data-to="/redeem"]').exists()).toBe(true)
    expect(sections[1].find('[data-to="/affiliate"]').exists()).toBe(true)
    expect(sections[1].find('[data-to="/profile"]').exists()).toBe(true)
    expect(wrapper.get('.sidebar-logo').attributes('data-to')).toBe('/admin/dashboard')
    expect(wrapper.find('[data-testid="sidebar-area-switch"]').exists()).toBe(false)
  })

  it('keeps the original compact admin behavior in simple mode', () => {
    mocks.authStore.isSimpleMode = true
    setRoute('/admin/settings', 'AdminSettings', true)
    const wrapper = mountSidebar()
    const sections = wrapper.findAll('nav.sidebar-nav > .sidebar-section')

    expect(sections).toHaveLength(1)
    expect(sections[0].find('[data-to="/admin/dashboard"]').exists()).toBe(true)
    expect(sections[0].find('[data-to="/dashboard"]').exists()).toBe(true)
    expect(sections[0].find('[data-to="/keys"]').exists()).toBe(true)
    expect(sections[0].find('[data-to="/admin/settings"]').exists()).toBe(true)
  })

  it('shows only the user menu for a regular user', () => {
    mocks.authStore.isAdmin = false
    setRoute('/dashboard', 'Dashboard', false)
    const wrapper = mountSidebar()
    const sections = wrapper.findAll('nav.sidebar-nav > .sidebar-section')

    expect(sections).toHaveLength(1)
    expect(sections[0].find('[data-to="/dashboard"]').exists()).toBe(true)
    expect(sections[0].find('[data-to="/keys"]').exists()).toBe(true)
    expect(sections[0].find('[data-to="/profile"]').exists()).toBe(true)
    expect(sections[0].find('[data-to="/admin/dashboard"]').exists()).toBe(false)
    expect(wrapper.get('.sidebar-logo').attributes('data-to')).toBe('/dashboard')
  })

  it('keeps admin-only custom pages in the admin menu', () => {
    mocks.adminSettingsStore.customMenuItems.push({
      id: 'admin-doc',
      label: 'Admin Doc',
      visibility: 'admin',
      sort_order: 0,
    })
    setRoute('/custom/admin-doc', 'CustomPage', false)
    mocks.route.params = { id: 'admin-doc' }

    const wrapper = mountSidebar()
    const sections = wrapper.findAll('nav.sidebar-nav > .sidebar-section')

    expect(sections[0].find('[data-to="/admin/dashboard"]').exists()).toBe(true)
    expect(sections[0].find('[data-to="/custom/admin-doc"]').exists()).toBe(true)
    expect(sections[1].find('[data-to="/keys"]').exists()).toBe(true)
    expect(wrapper.get('.sidebar-logo').attributes('data-to')).toBe('/admin/dashboard')
  })
})
