import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import UserDashboardAnnouncements from '../UserDashboardAnnouncements.vue'
import { useAnnouncementStore } from '@/stores/announcements'
import type { UserAnnouncement } from '@/types'

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      locale: 'zh',
      t: (key: string, params?: Record<string, unknown>) => (
        params?.count ? `${key}:${params.count}` : key
      )
    }
  }),
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => (
      params?.count ? `${key}:${params.count}` : key
    )
  })
}))

vi.mock('@/utils/format', () => ({
  formatRelativeTime: (value: string) => `time:${value}`
}))

function makeAnnouncement(overrides: Partial<UserAnnouncement> = {}): UserAnnouncement {
  return {
    id: 1,
    title: '国模上线',
    content: '**已经支持** 当前国内部分主流国模',
    notify_mode: 'silent',
    created_at: '2026-07-05T10:00:00Z',
    updated_at: '2026-07-05T10:00:00Z',
    ...overrides
  }
}

function mountComponent() {
  return mount(UserDashboardAnnouncements, {
    global: {
      stubs: {
        Icon: { template: '<span />' }
      }
    }
  })
}

describe('UserDashboardAnnouncements', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders the announcement timeline from the announcement store', () => {
    const store = useAnnouncementStore()
    store.announcements = [
      makeAnnouncement(),
      makeAnnouncement({ id: 2, title: 'QQ群', content: '地址：example.com', read_at: '2026-07-05T11:00:00Z' })
    ]
    vi.spyOn(store, 'fetchAnnouncements').mockResolvedValue(undefined)

    const wrapper = mountComponent()
    const text = wrapper.text()

    expect(text).toContain('announcements.title')
    expect(text).toContain('announcements.newCount:1')
    expect(text).toContain('国模上线')
    expect(text).toContain('已经支持 当前国内部分主流国模')
    expect(text).toContain('QQ群')
    expect(text).toContain('time:2026-07-05T10:00:00Z')
    expect(text).not.toContain('快捷操作')
    expect(store.fetchAnnouncements).toHaveBeenCalledTimes(1)
  })

  it('renders the empty state when there are no announcements', () => {
    const store = useAnnouncementStore()
    vi.spyOn(store, 'fetchAnnouncements').mockResolvedValue(undefined)

    const wrapper = mountComponent()
    const text = wrapper.text()

    expect(text).toContain('announcements.empty')
    expect(text).toContain('announcements.emptyDescription')
  })
})
