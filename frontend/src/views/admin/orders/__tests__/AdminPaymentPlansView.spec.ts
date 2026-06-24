import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminPaymentPlansView from '../AdminPaymentPlansView.vue'
import type { AdminGroup } from '@/types'
import type { ProviderInstance, SubscriptionPlan } from '@/types/payment'

const getAllGroups = vi.hoisted(() => vi.fn())
const getPlans = vi.hoisted(() => vi.fn())
const getProviders = vi.hoisted(() => vi.fn())
const updatePlan = vi.hoisted(() => vi.fn())
const deletePlan = vi.hoisted(() => vi.fn())
const routerPush = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin', () => ({
  default: {
    groups: {
      getAll: getAllGroups,
    },
  },
  adminAPI: {
    groups: {
      getAll: getAllGroups,
    },
  },
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getPlans,
    getProviders,
    updatePlan,
    deletePlan,
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
      push: routerPush,
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
  }
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
    validity_unit: 'days',
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

function mountView() {
  return mount(AdminPaymentPlansView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        ConfirmDialog: true,
        GroupBadge: true,
        Icon: true,
        PlanEditDialog: true,
        DataTable: {
          props: ['columns', 'data', 'loading'],
          template: `
            <div>
              <div data-test="plan-count">{{ data.length }}</div>
              <slot v-if="!data.length" name="empty" />
              <div v-for="row in data" :key="row.id">
                <slot name="cell-name" :value="row.name" :row="row" />
              </div>
            </div>
          `,
        },
      },
    },
  })
}

describe('AdminPaymentPlansView launch checklist', () => {
  beforeEach(() => {
    getAllGroups.mockReset().mockResolvedValue([])
    getPlans.mockReset().mockResolvedValue({ data: [] })
    getProviders.mockReset().mockResolvedValue({ data: [] })
    updatePlan.mockReset()
    deletePlan.mockReset()
    routerPush.mockReset()
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
    getAllGroups.mockResolvedValue([groupFixture()])
    getPlans.mockResolvedValue({ data: [planFixture()] })
    getProviders.mockResolvedValue({ data: [providerFixture()] })

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
    const groupButton = buttons.find(button => button.text() === 'payment.admin.launchSubscriptionGroupAction')
    const providerButton = buttons.find(button => button.text() === 'payment.admin.launchPaymentProviderAction')

    await groupButton?.trigger('click')
    await providerButton?.trigger('click')

    expect(routerPush).toHaveBeenCalledWith('/admin/groups')
    expect(routerPush).toHaveBeenCalledWith('/admin/settings')
  })
})
