import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import SubscriptionsView from '../SubscriptionsView.vue'

const routerPush = vi.hoisted(() => vi.fn())
const getMySubscriptions = vi.hoisted(() => vi.fn())
const getCheckoutInfo = vi.hoisted(() => vi.fn())
const refreshUser = vi.hoisted(() => vi.fn())
const purchaseWithBalance = vi.hoisted(() => vi.fn())
const fetchActiveSubscriptions = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
const showWarning = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const authState = vi.hoisted(() => ({
  user: {
    id: 1,
    username: 'demo',
    balance: 100,
  },
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
      t: (key: string, params?: Record<string, unknown>) =>
        params ? `${key} ${JSON.stringify(params)}` : key,
    }),
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess,
    showWarning,
    showError,
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    get user() {
      return authState.user
    },
    refreshUser,
  }),
}))

vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({
    purchaseWithBalance,
    fetchActiveSubscriptions,
  }),
}))

vi.mock('@/api/subscriptions', () => ({
  default: {
    getMySubscriptions,
  },
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getCheckoutInfo,
  },
}))

function planFixture(price = 25) {
  return {
    id: 7,
    group_id: 3,
    group_platform: 'openai',
    group_name: 'OpenAI',
    name: 'Starter',
    description: '',
    price,
    original_price: 0,
    validity_days: 30,
    validity_unit: 'day',
    features: [],
    product_name: '',
    for_sale: true,
    sort_order: 1,
    rate_multiplier: 1,
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    supported_model_scopes: [],
  }
}

function mountView() {
  return mount(SubscriptionsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: true,
        SubscriptionPlanCard: {
          props: ['plan', 'balance', 'balanceActionLoading'],
          emits: ['balance-subscribe', 'recharge', 'select'],
          template: `
            <div>
              <span data-test="plan-name">{{ plan.name }}</span>
              <span data-test="balance">{{ balance }}</span>
              <button data-test="balance-buy" @click="$emit('balance-subscribe', plan)">balance</button>
              <button data-test="recharge" @click="$emit('recharge', plan)">recharge</button>
              <button data-test="external" @click="$emit('select', plan)">external</button>
            </div>
          `,
        },
      },
    },
  })
}

describe('SubscriptionsView balance subscription marketplace', () => {
  beforeEach(() => {
    authState.user.balance = 100
    routerPush.mockReset()
    getMySubscriptions.mockReset().mockResolvedValue([])
    getCheckoutInfo.mockReset().mockResolvedValue({
      data: {
        plans: [planFixture()],
      },
    })
    refreshUser.mockReset().mockResolvedValue(undefined)
    purchaseWithBalance.mockReset().mockResolvedValue({
      price: 25,
      balance_after: 75,
      balance_before: 100,
      shortfall: 0,
      subscription: {},
    })
    fetchActiveSubscriptions.mockReset().mockResolvedValue([])
    showSuccess.mockReset()
    showWarning.mockReset()
    showError.mockReset()
  })

  it('shows current balance and purchasable plans', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('$100.00')
    expect(wrapper.text()).toContain('Starter')
    expect(wrapper.find('[data-test="balance"]').text()).toBe('100')
  })

  it('uses balance to subscribe and refreshes user and subscriptions', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-test="balance-buy"]').trigger('click')
    await flushPromises()

    expect(purchaseWithBalance).toHaveBeenCalledWith(7)
    expect(refreshUser).toHaveBeenCalled()
    expect(getMySubscriptions).toHaveBeenCalledTimes(2)
    expect(fetchActiveSubscriptions).toHaveBeenCalled()
    expect(showSuccess).toHaveBeenCalledWith(expect.stringContaining('payment.balanceSubscribeSuccess'))
  })

  it('shows shortfall and routes to recharge when balance is insufficient', async () => {
    authState.user.balance = 10
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-test="balance-buy"]').trigger('click')
    await flushPromises()

    expect(purchaseWithBalance).not.toHaveBeenCalled()
    expect(showWarning).toHaveBeenCalledWith(expect.stringContaining('$15.00'))
    expect(routerPush).toHaveBeenCalledWith({ path: '/purchase', query: { tab: 'recharge' } })
  })
})
