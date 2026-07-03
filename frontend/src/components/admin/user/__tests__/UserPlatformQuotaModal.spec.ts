import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

const apiMocks = vi.hoisted(() => ({
  getPlatformQuotas: vi.fn(),
  updatePlatformQuotas: vi.fn(),
  resetPlatformQuotaWindow: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getPlatformQuotas: apiMocks.getPlatformQuotas,
      updatePlatformQuotas: apiMocks.updatePlatformQuotas,
      resetPlatformQuotaWindow: apiMocks.resetPlatformQuotaWindow,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string>) => {
        if (params) {
          return key.replace(/\{(\w+)\}/g, (_, k) => params[k] ?? '')
        }
        return key
      },
    }),
  }
})

vi.mock('@/components/common/BaseDialog.vue', () => ({
  default: {
    name: 'BaseDialog',
    props: ['show', 'title', 'width'],
    template: '<div v-if="show"><slot /><slot name="footer" /></div>',
  },
}))

import UserPlatformQuotaModal from '../UserPlatformQuotaModal.vue'
import type { UserSubscription } from '@/types'

function makeUser(overrides: { subscriptions?: UserSubscription[] } = {}) {
  return { id: 99, email: 'u@example.com', ...overrides } as any
}

/** 鎸傝浇骞惰Е鍙?show锛歠alse 鈫?true锛岀‘淇?watch 琚縺娲?*/
async function mountAndOpen(extraProps: Record<string, unknown> = {}) {
  const w = mount(UserPlatformQuotaModal, {
    props: { show: false, user: makeUser(), ...extraProps },
  })
  await w.setProps({ show: true })
  await flushPromises()
  return w
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  apiMocks.getPlatformQuotas.mockResolvedValue({ platform_quotas: [] })
  apiMocks.updatePlatformQuotas.mockResolvedValue({ platform_quotas: [] })
  apiMocks.resetPlatformQuotaWindow.mockResolvedValue({ platform_quotas: [] })
})

describe('UserPlatformQuotaModal', () => {
  it('鎸傝浇骞?show=true 鏃惰皟鐢?getPlatformQuotas', async () => {
    await mountAndOpen()
    expect(apiMocks.getPlatformQuotas).toHaveBeenCalledWith(99)
  })

  it('绌烘暟鎹覆鏌?5 涓?platform 琛?, async () => {
    const w = await mountAndOpen()
    const html = w.html()
    expect(html).toContain('anthropic')
    expect(html).toContain('openai')
    expect(html).toContain('gemini')
    expect(html).toContain('antigravity')
    expect(html).toContain('grok')
  })

  it('宸叉湁鏁版嵁姝ｇ‘濉厖 limit input', async () => {
    apiMocks.getPlatformQuotas.mockResolvedValueOnce({
      platform_quotas: [
        { platform: 'anthropic', daily_limit_usd: 10, weekly_limit_usd: null, monthly_limit_usd: null,
          daily_usage_usd: 3.2, weekly_usage_usd: 0, monthly_usage_usd: 0 },
      ],
    })
    const w = await mountAndOpen()
    const inputs = w.findAll('input[type=number]')
    // 4 platforms 脳 3 windows = 12 inputs
    expect(inputs.length).toBe(15)
    // 绗竴涓?input 鏄?anthropic.daily = 10
    expect((inputs[0].element as HTMLInputElement).value).toBe('10')
  })

  it('淇濆瓨鎻愪氦瀹屾暣 5 platform payload', async () => {
    apiMocks.getPlatformQuotas.mockResolvedValueOnce({
      platform_quotas: [
        { platform: 'openai', daily_limit_usd: null, weekly_limit_usd: 20, monthly_limit_usd: null,
          daily_usage_usd: 0, weekly_usage_usd: 0, monthly_usage_usd: 0 },
      ],
    })
    const w = await mountAndOpen()
    // 鎵惧埌銆屼繚瀛樸€嶆寜閽紙鍖呭惈涓枃銆屼繚瀛樸€嶅瓧鏍风殑鎸夐挳锛?    const buttons = w.findAll('button')
    const saveBtn = buttons.find((b) => b.text() === 'admin.users.platformQuota.save')
    expect(saveBtn).toBeTruthy()
    await saveBtn!.trigger('click')
    await flushPromises()
    expect(apiMocks.updatePlatformQuotas).toHaveBeenCalledTimes(1)
    const [uid, payload] = apiMocks.updatePlatformQuotas.mock.calls[0]
    expect(uid).toBe(99)
    expect(payload).toHaveLength(5)
    const openai = payload.find((p: any) => p.platform === 'openai')
    expect(openai.weekly_limit_usd).toBe(20)
  })

  it('鍏ㄩ儴娓呯┖鎶婃墍鏈?limit 缃?null锛堢‘璁ら€氳繃锛?, async () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
    apiMocks.getPlatformQuotas.mockResolvedValueOnce({
      platform_quotas: [
        { platform: 'anthropic', daily_limit_usd: 10, weekly_limit_usd: 50, monthly_limit_usd: 100,
          daily_usage_usd: 0, weekly_usage_usd: 0, monthly_usage_usd: 0 },
      ],
    })
    const w = await mountAndOpen()
    const buttons = w.findAll('button')
    const clearBtn = buttons.find((b) => b.text() === 'admin.users.platformQuota.clearAll')
    expect(clearBtn).toBeTruthy()
    await clearBtn!.trigger('click')
    await flushPromises()
    expect(confirmSpy).toHaveBeenCalledTimes(1)
    const inputs = w.findAll('input[type=number]')
    for (const inp of inputs) {
      expect((inp.element as HTMLInputElement).value).toBe('')
    }
    confirmSpy.mockRestore()
  })

  it('鍏ㄩ儴娓呯┖ confirm 鍙栨秷鍒欎繚鎸佸師鍊?, async () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
    apiMocks.getPlatformQuotas.mockResolvedValueOnce({
      platform_quotas: [
        { platform: 'anthropic', daily_limit_usd: 10, weekly_limit_usd: 50, monthly_limit_usd: 100,
          daily_usage_usd: 0, weekly_usage_usd: 0, monthly_usage_usd: 0 },
      ],
    })
    const w = await mountAndOpen()
    const clearBtn = w.findAll('button').find((b) => b.text() === 'admin.users.platformQuota.clearAll')
    await clearBtn!.trigger('click')
    await flushPromises()
    expect(confirmSpy).toHaveBeenCalledTimes(1)
    // anthropic daily 搴斾繚鎸?10锛堟湭琚竻绌猴級
    const inputs = w.findAll('input[type=number]')
    const dailyVal = (inputs[0].element as HTMLInputElement).value
    expect(dailyVal).toBe('10')
    confirmSpy.mockRestore()
  })

  it('閲嶇疆鎸夐挳 confirm 鍙栨秷鍒欎笉璋冪敤 API', async () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
    const w = await mountAndOpen()
    const resetBtns = w.findAll('button').filter((b) => b.text() === '鈫?)
    expect(resetBtns.length).toBeGreaterThan(0)
    await resetBtns[0].trigger('click')
    await flushPromises()
    expect(apiMocks.resetPlatformQuotaWindow).not.toHaveBeenCalled()
    confirmSpy.mockRestore()
  })

  it('閲嶇疆鎸夐挳 confirm 纭鍒欒皟鐢?API', async () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
    const w = await mountAndOpen()
    const resetBtns = w.findAll('button').filter((b) => b.text() === '鈫?)
    await resetBtns[0].trigger('click') // 绗竴涓槸 anthropic.daily
    await flushPromises()
    expect(apiMocks.resetPlatformQuotaWindow).toHaveBeenCalledWith(99, 'anthropic', 'daily')
    confirmSpy.mockRestore()
  })

  describe('subscription warning banner', () => {
    it('displays subscription warning when user has active subscription', async () => {
      const w = mount(UserPlatformQuotaModal, {
        props: {
          show: true,
          user: makeUser({
            subscriptions: [
              {
                id: 1, user_id: 99, group_id: 1, status: 'active',
                starts_at: '2026-01-01T00:00:00Z', expires_at: null,
                daily_usage_usd: 0, weekly_usage_usd: 0, monthly_usage_usd: 0,
                daily_window_start: null, weekly_window_start: null, monthly_window_start: null,
                created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z',
              } as UserSubscription,
            ],
          }),
        },
      })
      await flushPromises()
      expect(w.html()).toContain('admin.users.platformQuota.subscriptionWarning')
    })

    it('hides subscription warning when user has only expired subscriptions', async () => {
      const w = mount(UserPlatformQuotaModal, {
        props: {
          show: true,
          user: makeUser({
            subscriptions: [
              {
                id: 2, user_id: 99, group_id: 1, status: 'expired',
                starts_at: '2025-01-01T00:00:00Z', expires_at: '2025-12-31T00:00:00Z',
                daily_usage_usd: 0, weekly_usage_usd: 0, monthly_usage_usd: 0,
                daily_window_start: null, weekly_window_start: null, monthly_window_start: null,
                created_at: '2025-01-01T00:00:00Z', updated_at: '2025-12-31T00:00:00Z',
              } as UserSubscription,
            ],
          }),
        },
      })
      await flushPromises()
      expect(w.html()).not.toContain('admin.users.platformQuota.subscriptionWarning')
    })

    it('hides subscription warning when subscriptions is empty array', async () => {
      const w = mount(UserPlatformQuotaModal, {
        props: {
          show: true,
          user: makeUser({ subscriptions: [] }),
        },
      })
      await flushPromises()
      expect(w.html()).not.toContain('admin.users.platformQuota.subscriptionWarning')
    })
  })
})
