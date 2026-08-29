import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import DailyCheckInsView from '../DailyCheckInsView.vue'

const { listDailyCheckIns, getUserById, showError } = vi.hoisted(() => ({
  listDailyCheckIns: vi.fn(),
  getUserById: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    dailyCheckIns: {
      list: listDailyCheckIns
    },
    users: {
      getById: getUserById
    }
  }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

const row = {
  id: 41,
  user_id: 7,
  email: 'alice@example.com',
  username: 'Alice',
  service_date: '2026-07-21',
  reward_amount: 2,
  balance_before: 10,
  balance_after: 12,
  checked_in_at: '2026-07-21T03:04:05+08:00',
  created_at: '2026-07-21T03:04:05+08:00'
}

const DataTableStub = {
  props: ['columns', 'data', 'loading'],
  template: `
    <div>
      <div data-test="columns">{{ columns.map(column => column.key).join(',') }}</div>
      <div v-for="item in data" :key="item.id" data-test="check-in-row">
        <slot name="cell-user" :row="item" :value="item.user_id" />
        <slot name="cell-reward_amount" :row="item" :value="item.reward_amount" />
        <slot name="cell-balance" :row="item" />
      </div>
      <slot v-if="!loading && data.length === 0" name="empty" />
    </div>
  `
}

const PaginationStub = {
  emits: ['update:page', 'update:pageSize'],
  template: `
    <div>
      <button data-test="next-page" @click="$emit('update:page', 2)">next</button>
      <button data-test="page-size" @click="$emit('update:pageSize', 50)">size</button>
    </div>
  `
}

const UserBalanceHistoryModalStub = {
  props: ['show', 'user', 'hideActions'],
  emits: ['close'],
  template: '<div data-test="balance-history-modal" :data-show="show" :data-user-id="user?.id || null" />'
}

function mountView() {
  return mount(DailyCheckInsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
        },
        DataTable: DataTableStub,
        Pagination: PaginationStub,
        UserBalanceHistoryModal: UserBalanceHistoryModalStub,
        Icon: true
      }
    }
  })
}

describe('admin DailyCheckInsView', () => {
  beforeEach(() => {
    listDailyCheckIns.mockReset()
    getUserById.mockReset()
    showError.mockReset()
    listDailyCheckIns.mockResolvedValue({
      items: [row], total: 1, page: 1, page_size: 20, pages: 1
    })
    getUserById.mockResolvedValue({ id: 7, email: row.email, balance: 10, created_at: row.created_at })
  })

  it('uses the backend current day, then supports all history and exact dates', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(listDailyCheckIns).toHaveBeenCalledWith({ page: 1, page_size: 20 })
    expect(wrapper.get('[data-test="columns"]').text().split(',')).toEqual([
      'user', 'service_date', 'reward_amount', 'balance', 'checked_in_at'
    ])
    expect(wrapper.get('[data-test="check-in-row"]').text()).toContain('alice@example.com')
    expect(wrapper.get('[data-test="check-in-row"]').text()).toContain('10')
    expect(wrapper.get('[data-test="check-in-row"]').text()).toContain('12')

    await wrapper.get('[data-test="date-mode-history"]').trigger('click')
    await flushPromises()
    expect(listDailyCheckIns).toHaveBeenLastCalledWith({ page: 1, page_size: 20, all: true })

    await wrapper.get('[data-test="service-date"]').setValue('2026-07-20')
    await flushPromises()
    expect(listDailyCheckIns).toHaveBeenLastCalledWith({
      page: 1, page_size: 20, service_date: '2026-07-20'
    })

    await wrapper.get('[data-test="next-page"]').trigger('click')
    await flushPromises()
    expect(listDailyCheckIns).toHaveBeenLastCalledWith({
      page: 2, page_size: 20, service_date: '2026-07-20'
    })
  })

  it('keeps rendered rows when a refresh fails and surfaces the error', async () => {
    const wrapper = mountView()
    await flushPromises()
    listDailyCheckIns.mockRejectedValueOnce(new Error('network down'))

    await wrapper.get('[data-test="search"]').trigger('click')
    await flushPromises()

    expect(wrapper.get('[data-test="check-in-row"]').text()).toContain('alice@example.com')
    expect(showError).toHaveBeenCalledWith('network down')
  })

  it('opens the read-only balance history for the clicked user', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="daily-check-in-user"]').trigger('click')
    await flushPromises()

    expect(getUserById).toHaveBeenCalledWith(7, true)
    expect(wrapper.get('[data-test="balance-history-modal"]').attributes('data-show')).toBe('true')
    expect(wrapper.get('[data-test="balance-history-modal"]').attributes('data-user-id')).toBe('7')
  })

  it('surfaces a user lookup failure without changing the table', async () => {
    const wrapper = mountView()
    await flushPromises()
    getUserById.mockRejectedValueOnce(new Error('user unavailable'))

    await wrapper.get('[data-test="daily-check-in-user"]').trigger('click')
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('admin.usage.failedToLoadUser')
    expect(wrapper.get('[data-test="check-in-row"]').text()).toContain('alice@example.com')
  })
})
