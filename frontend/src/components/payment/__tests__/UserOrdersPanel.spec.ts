import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import UserOrdersPanel from '../UserOrdersPanel.vue'
import type { PaymentOrder } from '@/types/payment'

const getMyOrders = vi.hoisted(() => vi.fn())
const getRefundEligibleProviders = vi.hoisted(() => vi.fn())
const cancelOrder = vi.hoisted(() => vi.fn())
const requestRefund = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getMyOrders,
    getRefundEligibleProviders,
    cancelOrder,
    requestRefund,
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showSuccess, showError }),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

const order = (overrides: Partial<PaymentOrder> = {}): PaymentOrder => ({
  id: 1,
  user_id: 2,
  amount: 10,
  pay_amount: 10,
  fee_rate: 0,
  payment_type: 'alipay',
  out_trade_no: 'order-1',
  status: 'PENDING',
  order_type: 'balance',
  created_at: '2026-07-10T00:00:00Z',
  expires_at: '2026-07-10T00:15:00Z',
  refund_amount: 0,
  ...overrides,
})

const mountPanel = () => mount(UserOrdersPanel, {
  global: {
    stubs: {
      Select: true,
      Pagination: true,
      Icon: true,
      OrderTable: {
        props: ['orders'],
        template: '<div><div v-for="row in orders" :key="row.id"><slot name="actions" :row="row" /></div></div>',
      },
      BaseDialog: {
        props: ['show'],
        template: '<div v-if="show"><slot /><slot name="footer" /></div>',
      },
    },
  },
})

describe('UserOrdersPanel', () => {
  beforeEach(() => {
    getMyOrders.mockReset().mockResolvedValue({ data: { items: [order()], total: 1 } })
    getRefundEligibleProviders.mockReset().mockResolvedValue({ data: { provider_instance_ids: [] } })
    cancelOrder.mockReset().mockResolvedValue(undefined)
    requestRefund.mockReset().mockResolvedValue(undefined)
    showSuccess.mockReset()
    showError.mockReset()
  })

  it('loads and cancels a pending order through the shared panel', async () => {
    const wrapper = mountPanel()
    await flushPromises()

    expect(getMyOrders).toHaveBeenCalledWith({ page: 1, page_size: 20, status: undefined })
    await wrapper.get('[data-testid="cancel-order"]').trigger('click')
    await wrapper.get('[data-testid="confirm-cancel"]').trigger('click')
    await flushPromises()

    expect(cancelOrder).toHaveBeenCalledWith(1)
    expect(getMyOrders).toHaveBeenCalledTimes(2)
  })

  it('shows refund only for an eligible completed order', async () => {
    getMyOrders.mockResolvedValue({
      data: { items: [order({ status: 'COMPLETED', provider_instance_id: 'provider-1' })], total: 1 },
    })
    getRefundEligibleProviders.mockResolvedValue({ data: { provider_instance_ids: ['provider-1'] } })
    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.get('[data-testid="refund-order"]').trigger('click')
    await wrapper.get('textarea').setValue('Duplicate payment')
    await wrapper.get('[data-testid="confirm-refund"]').trigger('click')
    await flushPromises()

    expect(requestRefund).toHaveBeenCalledWith(1, { reason: 'Duplicate payment' })
  })
})
