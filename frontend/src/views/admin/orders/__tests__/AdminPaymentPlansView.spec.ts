import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AdminGroup } from '@/types'
import type { ProviderInstance, SubscriptionPlan } from '@/types/payment'
import AdminPaymentPlansView from '../AdminPaymentPlansView.vue'

const mocks = vi.hoisted(() => ({
  deletePlan: vi.fn(),
  getConfig: vi.fn(),
  getGroups: vi.fn(),
  getPlans: vi.fn(),
  getProviders: vi.fn(),
  routerPush: vi.fn(),
  updatePlan: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  default: {
    groups: {
      getAll: mocks.getGroups,
    },
  },
  adminAPI: {
    groups: {
      getAll: mocks.getGroups,
    },
  },
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    deletePlan: mocks.deletePlan,
    getConfig: mocks.getConfig,
    getPlans: mocks.getPlans,
    getProviders: mocks.getProviders,
    updatePlan: mocks.updatePlan,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: mocks.routerPush,
    }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

function groupFixture(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 1,
    name: 'Subscription Group',
    description: '',
    platform: 'openai',
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'subscription',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    allow_image_generation: false,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    require_oauth_only: false,
    require_privacy_set: false,
    created_at: '2026-06-23T00:00:00Z',
    updated_at: '2026-06-23T00:00:00Z',
    model_routing: null,
    model_routing_enabled: false,
    mcp_xml_inject: false,
    sort_order: 0,
    ...overrides,
  } as AdminGroup
}

function planFixture(overrides: Partial<SubscriptionPlan> = {}): SubscriptionPlan {
  return {
    id: 10,
    group_id: 1,
    name: 'Launch Plan',
    description: '',
    price: 19.9,
    original_price: 0,
    validity_days: 30,
    validity_unit: 'day',
    features: [],
    for_sale: true,
    sort_order: 1,
    ...overrides,
  }
}

function providerFixture(overrides: Partial<ProviderInstance> = {}): ProviderInstance {
  return {
    id: 3,
    provider_key: 'stripe',
    name: 'Stripe',
    config: {},
    supported_types: ['stripe'],
    enabled: true,
    payment_mode: '',
    refund_enabled: false,
    allow_user_refund: false,
    limits: '',
    sort_order: 0,
    ...overrides,
  }
}

const DataTableStub = {
  props: ['columns', 'data', 'loading'],
  template: `
    <div>
      <div data-test="plan-count">{{ data.length }}</div>
      <slot v-if="!data.length" name="empty" />
      <div v-for="row in data" :key="row.id">
        <slot name="cell-name" :value="row.name" :row="row" />
        <slot name="cell-price" :value="row.price" :row="row" />
      </div>
    </div>
  `,
}

function mountView() {
  return mount(AdminPaymentPlansView, {
    global: {
      plugins: [createPinia()],
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        ConfirmDialog: true,
        DataTable: DataTableStub,
        GroupBadge: true,
        Icon: true,
        PlanEditDialog: true,
      },
    },
  })
}

describe('AdminPaymentPlansView', () => {
  beforeEach(() => {
    mocks.deletePlan.mockReset()
    mocks.getConfig.mockReset().mockResolvedValue({ data: {} })
    mocks.getGroups.mockReset().mockResolvedValue([])
    mocks.getPlans.mockReset().mockResolvedValue({ data: [] })
    mocks.getProviders.mockReset().mockResolvedValue({ data: [] })
    mocks.routerPush.mockReset()
    mocks.updatePlan.mockReset()
  })

  it('shows missing launch prerequisites when store configuration is empty', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('payment.admin.launchChecklistTitle')
    expect(wrapper.text()).toContain('payment.admin.launchNeedsSetup')
    expect(wrapper.text()).toContain('payment.admin.launchSubscriptionGroupTitle')
    expect(wrapper.text()).toContain('payment.admin.launchSubscriptionPlanTitle')
    expect(wrapper.text()).toContain('payment.admin.launchPaymentProviderTitle')
    expect(wrapper.text()).toContain('payment.admin.launchBlocked')
    expect(wrapper.text()).toContain('payment.admin.noSubscriptionGroupsTitle')
  })

  it('marks the store launch checklist ready when group, plan, and provider exist', async () => {
    mocks.getGroups.mockResolvedValue([groupFixture()])
    mocks.getPlans.mockResolvedValue({ data: [planFixture()] })
    mocks.getProviders.mockResolvedValue({ data: [providerFixture()] })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('payment.admin.launchReady')
    expect(wrapper.text()).toContain('Launch Plan')
    expect(wrapper.find('[data-test="plan-count"]').text()).toBe('1')
    expect(wrapper.text()).not.toContain('payment.admin.launchBlocked')
  })

  it('routes admins to prerequisite configuration pages', async () => {
    const wrapper = mountView()
    await flushPromises()

    const buttons = wrapper.findAll('button')
    const groupButton = buttons.find(
      (button) => button.text() === 'payment.admin.launchSubscriptionGroupAction',
    )
    const providerButton = buttons.find(
      (button) => button.text() === 'payment.admin.launchPaymentProviderAction',
    )

    await groupButton?.trigger('click')
    await providerButton?.trigger('click')

    expect(mocks.routerPush).toHaveBeenCalledWith('/admin/groups')
    expect(mocks.routerPush).toHaveBeenCalledWith('/admin/settings')
  })

  it('uses the configured currency symbol and keeps legacy prices in USD', async () => {
    mocks.getPlans.mockResolvedValue({
      data: [
        planFixture({
          id: 1,
          name: 'CNY plan',
          price: 499,
          original_price: 599,
          currency: 'CNY',
        }),
        planFixture({
          id: 2,
          name: 'Legacy plan',
          price: 10,
          currency: '',
        }),
      ],
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('¥499.00CNY')
    expect(wrapper.text()).toContain('¥599.00')
    expect(wrapper.text()).toContain('$10.00')
  })
})
