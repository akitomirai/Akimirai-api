import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfilePrivacyFilterCard from '@/components/user/profile/ProfilePrivacyFilterCard.vue'
import type { User } from '@/types'

const {
  updateProfileMock,
  showSuccessMock,
  showErrorMock,
  authStoreState,
} = vi.hoisted(() => ({
  updateProfileMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn(),
  authStoreState: {
    user: null as User | null,
  },
}))

vi.mock('@/api', () => ({
  userAPI: {
    updateProfile: updateProfileMock,
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreState,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock,
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (error: unknown) => (error as Error).message || 'request failed',
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

describe('ProfilePrivacyFilterCard', () => {
  beforeEach(() => {
    updateProfileMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
    authStoreState.user = null
  })

  it('submits normalized config when a filter type is toggled', async () => {
    updateProfileMock.mockResolvedValue({
      privacy_filter_config: {
        enabled: true,
        types: ['email', 'phone'],
      },
    })

    const wrapper = mount(ProfilePrivacyFilterCard, {
      props: {
        config: {
          enabled: true,
          types: ['email'],
        },
      },
    })

    await wrapper.get('[data-testid="privacy-filter-type-phone-input"]').setValue(true)
    await flushPromises()

    expect(updateProfileMock).toHaveBeenCalledWith({
      privacy_filter_config: {
        enabled: true,
        types: ['email', 'phone'],
      },
    })
    expect(showSuccessMock).toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
  })

  it('updates the enabled flag with the existing selected types', async () => {
    updateProfileMock.mockResolvedValue({
      privacy_filter_config: {
        enabled: false,
        types: ['email'],
      },
    })

    const wrapper = mount(ProfilePrivacyFilterCard, {
      props: {
        config: {
          enabled: true,
          types: ['email'],
        },
      },
    })

    await wrapper.get('[data-testid="privacy-filter-enabled-toggle"]').trigger('click')
    await flushPromises()

    expect(updateProfileMock).toHaveBeenCalledWith({
      privacy_filter_config: {
        enabled: false,
        types: ['email'],
      },
    })
    expect(showSuccessMock).toHaveBeenCalledWith('profile.privacyFilter.disabledSuccess')
  })

  it('keeps the submitted enabled state when the response omits privacy_filter_config', async () => {
    updateProfileMock.mockResolvedValue({
      id: 5,
      username: 'alice',
      email: 'alice@example.com',
      role: 'user',
      balance: 0,
      concurrency: 1,
      status: 'active',
      allowed_groups: null,
      balance_notify_enabled: true,
      balance_notify_threshold: null,
      balance_notify_extra_emails: [],
      created_at: '2026-07-05T00:00:00Z',
      updated_at: '2026-07-05T00:00:00Z',
    } satisfies User)

    const wrapper = mount(ProfilePrivacyFilterCard, {
      props: {
        config: {
          enabled: false,
          types: ['email'],
        },
      },
    })

    await wrapper.get('[data-testid="privacy-filter-enabled-toggle"]').trigger('click')
    await flushPromises()

    expect(wrapper.get('[data-testid="privacy-filter-enabled-toggle"]').attributes('aria-checked')).toBe('true')
    expect(authStoreState.user?.privacy_filter_config).toEqual({
      enabled: true,
      types: ['email'],
    })
    expect(showSuccessMock).toHaveBeenCalledWith('profile.privacyFilter.enabledSuccess')
  })
})
