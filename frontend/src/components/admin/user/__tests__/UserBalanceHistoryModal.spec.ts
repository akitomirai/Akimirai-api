import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import UserBalanceHistoryModal from '../UserBalanceHistoryModal.vue'

const { getUserBalanceHistory } = vi.hoisted(() => ({
  getUserBalanceHistory: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getUserBalanceHistory
    }
  }
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => ({
        'admin.users.balanceAddedDailyCheckIn': '签到',
      }[key] ?? key)
    })
  }
})

const user = {
  id: 7,
  email: 'alice@example.com',
  username: 'Alice',
  balance: 10.25,
  created_at: '2026-07-21T03:04:05Z'
}

const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: '<div v-if="show"><slot /></div>'
}

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  template: `
    <div>
      <button data-test="daily-check-in-filter" @click="$emit('update:modelValue', 'daily_check_in'); $emit('change')">
        {{ options.map(option => option.label).join('|') }}
      </button>
    </div>
  `
}

const IconStub = {
  props: ['name'],
  template: '<span data-test="history-icon" :data-icon="name" />'
}

describe('UserBalanceHistoryModal daily check-in history', () => {
  it('offers the check-in type and renders decimal reward records', async () => {
    getUserBalanceHistory.mockReset()
    getUserBalanceHistory.mockResolvedValue({
      items: [{
        id: 41,
        code: 'CHECKIN-41',
        type: 'daily_check_in',
        value: 0.25,
        status: 'used',
        used_at: '2026-07-21T03:04:05Z',
        created_at: '2026-07-21T03:04:05Z'
      }],
      total: 1,
      total_recharged: 10
    })

    const wrapper = mount(UserBalanceHistoryModal, {
      props: { show: false, user },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          Icon: IconStub
        }
      }
    })

    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(getUserBalanceHistory).toHaveBeenCalledWith(7, 1, 15, undefined)
    expect(wrapper.get('[data-test="daily-check-in-filter"]').text()).toContain('admin.users.typeDailyCheckIn')
    expect(wrapper.text()).toContain('签到')
    expect(wrapper.text()).toContain('+$0.25')
    expect(wrapper.findAll('[data-test="history-icon"]').at(-1)?.attributes('data-icon')).toBe('dollar')

    await wrapper.get('[data-test="daily-check-in-filter"]').trigger('click')
    await flushPromises()
    expect(getUserBalanceHistory).toHaveBeenLastCalledWith(7, 1, 15, 'daily_check_in')
  })
})
