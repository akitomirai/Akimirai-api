import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UserDashboardCommercialPath from '../UserDashboardCommercialPath.vue'
import type { UserDashboardStats } from '@/api/usage'
import type { ApiKey, UserErrorRequest } from '@/types'

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      locale: 'en',
      t: (key: string) => key
    }
  }),
  useI18n: () => ({
    t: (key: string, paramsOrFallback?: Record<string, unknown> | string) => {
      if (typeof paramsOrFallback === 'string') return paramsOrFallback
      const params = paramsOrFallback
      if (params?.time) return `${key} ${params.time}`
      return key
    }
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

const stats: UserDashboardStats = {
  total_api_keys: 2,
  active_api_keys: 1,
  total_requests: 10,
  total_input_tokens: 100,
  total_output_tokens: 200,
  total_cache_creation_tokens: 0,
  total_cache_read_tokens: 0,
  total_tokens: 300,
  total_cost: 1.2,
  total_actual_cost: 1.1,
  today_requests: 3,
  today_input_tokens: 40,
  today_output_tokens: 60,
  today_cache_creation_tokens: 0,
  today_cache_read_tokens: 0,
  today_tokens: 100,
  today_cost: 0.42,
  today_actual_cost: 0.4,
  average_duration_ms: 250,
  rpm: 1,
  tpm: 100,
}

const apiKeys = [
  {
    id: 1,
    name: 'prod-key',
    status: 'active',
    last_used_at: '2026-06-29T08:00:00Z',
  } as ApiKey
]

const recentErrors: UserErrorRequest[] = [
  {
    id: 7,
    created_at: '2026-06-29T08:30:00Z',
    model: 'gpt-test',
    inbound_endpoint: '/v1/chat/completions',
    status_code: 502,
    request_id: 'req-safe-1',
    category: 'upstream',
    platform: 'openai',
    message: 'do-not-render token cookie private_key sk-test-placeholder',
    error_code: 'UPSTREAM_5XX',
    explanation: 'Upstream unavailable',
    suggestion: 'Retry later',
    retryable: true,
    charged: false,
    http_status: 502,
    key_name: 'prod-key',
    key_deleted: false,
  }
]

const mountComponent = (props = {}) => mount(UserDashboardCommercialPath, {
  props: {
    stats,
    balance: 12.34,
    apiKeys,
    apiKeysLoading: false,
    apiKeysError: false,
    recentErrors,
    errorsLoading: false,
    errorsError: false,
    errorViewEnabled: true,
    baseUrl: 'https://example.com/v1',
    recommendedModel: 'gpt-test',
    availableModelCount: 3,
    modelsLoading: false,
    modelsError: false,
    paymentEnabled: true,
    ...props
  },
  global: {
    stubs: {
      RouterLink: {
        props: ['to'],
        template: '<a><slot /></a>'
      },
      Icon: {
        template: '<span />'
      }
    }
  }
})

describe('UserDashboardCommercialPath', () => {
  it('renders the commercial dashboard cards and main user entries', () => {
    const wrapper = mountComponent()
    const text = wrapper.text()

    expect(text).toContain('dashboard.commercial.balanceTitle')
    expect(text).toContain('$12.34')
    expect(text).toContain('dashboard.commercial.todayUsageTitle')
    expect(text).toContain('$0.40')
    expect(text).toContain('dashboard.commercial.keyStatusTitle')
    expect(text).toContain('2')
    expect(text).toContain('dashboard.commercial.quickStartTitle')
    expect(text).toContain('https://example.com/v1')
    expect(text).toContain('gpt-test')
    expect(text).toContain('3 available models')
    expect(text).toContain('dashboard.commercial.paymentTitle')
    expect(text).toContain('dashboard.commercial.recharge')
  })

  it('shows safe recent error fields without rendering raw sensitive message content', () => {
    const wrapper = mountComponent()
    const text = wrapper.text()

    expect(text).toContain('UPSTREAM_5XX')
    expect(text).toContain('Upstream unavailable')
    expect(text).toContain('Retry later')
    expect(text).toContain('req-safe-1')
    expect(text).not.toContain('do-not-render')
    expect(text).not.toContain('sk-test-placeholder')
    expect(text).not.toContain('private_key')
    expect(text).not.toContain('cookie')
  })

  it('links model availability errors back to available models', () => {
    const wrapper = mountComponent({
      recentErrors: [
        {
          ...recentErrors[0],
          error_code: 'NO_AVAILABLE_CHANNEL',
          explanation: 'No channel',
          suggestion: 'Switch model',
        }
      ]
    })
    const text = wrapper.text()

    expect(text).toContain('This failure may be caused by a disabled model or no available channel')
    expect(text).toContain('View available models')
  })

  it('uses empty states instead of fabricated values when data is missing', () => {
    const wrapper = mountComponent({
      stats: null,
      balance: null,
      apiKeys: [],
      recentErrors: [],
      paymentEnabled: false
    })
    const text = wrapper.text()

    expect(text).toContain('dashboard.commercial.noBalance')
    expect(text).toContain('dashboard.commercial.noTodayStats')
    expect(text).toContain('dashboard.commercial.noRecentErrors')
    expect(text).toContain('dashboard.commercial.paymentDisabled')
  })
})
