import { defineComponent, h, reactive, type ComponentPublicInstance } from 'vue'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { DailyCheckInResponse } from '@/api/checkin'
import { useDailyCheckIn, type DailyCheckInState } from '@/composables/useDailyCheckIn'

function response(overrides: Partial<DailyCheckInResponse> = {}): DailyCheckInResponse {
  return {
    checked_in: false,
    already_checked_in: false,
    service_date: '2026-07-21',
    reward_amount: 0,
    balance_before: 10,
    balance_after: 10,
    checked_in_at: null,
    next_reset_at: '2026-07-22T02:00:00+08:00',
    ...overrides,
  }
}

function mountComposable(options: Parameters<typeof useDailyCheckIn>[0]) {
  let state!: DailyCheckInState
  const wrapper = mount(defineComponent({
    setup() {
      state = useDailyCheckIn(options)
      return () => h('div')
    },
  }))
  return { wrapper: wrapper as VueWrapper<ComponentPublicInstance>, state }
}

describe('useDailyCheckIn', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-07-21T03:00:00+08:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('loads status once and deduplicates concurrent GET requests', async () => {
    let resolveStatus!: (value: DailyCheckInResponse) => void
    const pending = new Promise<DailyCheckInResponse>((resolve) => { resolveStatus = resolve })
    const api = { getStatus: vi.fn(() => pending), claim: vi.fn() }
    const authStore = reactive({ user: { id: 7 }, refreshUser: vi.fn() })
    const appStore = { showSuccess: vi.fn(), showWarning: vi.fn() }
    const { wrapper, state } = mountComposable({ api, authStore, appStore })

    const extraRefresh = state.refreshStatus()
    expect(api.getStatus).toHaveBeenCalledTimes(1)
    resolveStatus(response())
    await extraRefresh
    await flushPromises()

    expect(state.phase.value).toBe('available')
    wrapper.unmount()
  })

  it('keeps the claimed state when the balance refresh fails', async () => {
    const claimed = response({
      checked_in: true,
      reward_amount: 2,
      balance_after: 12,
      checked_in_at: '2026-07-21T03:04:05+08:00',
    })
    const api = {
      getStatus: vi.fn().mockResolvedValue(response()),
      claim: vi.fn().mockResolvedValue(claimed),
    }
    const authStore = reactive({
      user: { id: 7 },
      refreshUser: vi.fn().mockRejectedValue(new Error('profile unavailable')),
    })
    const appStore = { showSuccess: vi.fn(), showWarning: vi.fn() }
    const { wrapper, state } = mountComposable({ api, authStore, appStore })
    await flushPromises()

    await state.handleAction()

    expect(state.phase.value).toBe('claimed')
    expect(state.status.value?.reward_amount).toBe(2)
    expect(appStore.showWarning).toHaveBeenCalledTimes(1)
    expect(api.getStatus).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })

  it('reconciles an ambiguous POST with GET before allowing another mutation', async () => {
    const reconciled = response({
      checked_in: true,
      already_checked_in: true,
      reward_amount: 3,
      balance_after: 13,
      checked_in_at: '2026-07-21T03:04:05+08:00',
    })
    const api = {
      getStatus: vi.fn()
        .mockResolvedValueOnce(response())
        .mockResolvedValueOnce(reconciled),
      claim: vi.fn().mockRejectedValue(new Error('connection reset')),
    }
    const authStore = reactive({ user: { id: 7 }, refreshUser: vi.fn().mockResolvedValue({ id: 7 }) })
    const appStore = { showSuccess: vi.fn(), showWarning: vi.fn() }
    const { wrapper, state } = mountComposable({ api, authStore, appStore })
    await flushPromises()

    await state.handleAction()

    expect(api.claim).toHaveBeenCalledTimes(1)
    expect(api.getStatus).toHaveBeenCalledTimes(2)
    expect(state.phase.value).toBe('claimed')
    expect(state.status.value?.reward_amount).toBe(3)
    wrapper.unmount()
  })

  it('uses an error-state click for GET reconciliation instead of a second POST', async () => {
    const api = {
      getStatus: vi.fn()
        .mockResolvedValueOnce(response())
        .mockRejectedValueOnce(new Error('status unavailable'))
        .mockResolvedValueOnce(response()),
      claim: vi.fn().mockRejectedValue(new Error('connection reset')),
    }
    const authStore = reactive({ user: { id: 7 }, refreshUser: vi.fn() })
    const appStore = { showSuccess: vi.fn(), showWarning: vi.fn() }
    const { wrapper, state } = mountComposable({ api, authStore, appStore })
    await flushPromises()

    await state.handleAction()
    expect(state.phase.value).toBe('error')
    await state.handleAction()

    expect(api.claim).toHaveBeenCalledTimes(1)
    expect(api.getStatus).toHaveBeenCalledTimes(3)
    expect(state.phase.value).toBe('available')
    wrapper.unmount()
  })

  it('revalidates at next_reset_at and stops the timer on unmount', async () => {
    const api = {
      getStatus: vi.fn().mockResolvedValue(response({
        next_reset_at: '2026-07-21T03:00:02+08:00',
      })),
      claim: vi.fn(),
    }
    const authStore = reactive({ user: { id: 7 }, refreshUser: vi.fn() })
    const appStore = { showSuccess: vi.fn(), showWarning: vi.fn() }
    const { wrapper } = mountComposable({ api, authStore, appStore })
    await flushPromises()

    await vi.advanceTimersByTimeAsync(2250)
    await flushPromises()
    expect(api.getStatus).toHaveBeenCalledTimes(2)

    wrapper.unmount()
    await vi.advanceTimersByTimeAsync(60_000)
    expect(api.getStatus).toHaveBeenCalledTimes(2)
  })
})
