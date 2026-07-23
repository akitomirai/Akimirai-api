import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get } = vi.hoisted(() => ({ get: vi.fn() }))

vi.mock('@/api/client', () => ({
  apiClient: { get }
}))

import { dailyCheckInsAPI } from '@/api/admin/dailyCheckIns'

describe('admin daily check-ins API', () => {
  beforeEach(() => {
    get.mockReset()
  })

  it('serializes the server-owned current-day request without a client date', async () => {
    const response = { items: [], total: 0, page: 1, page_size: 20, pages: 1 }
    get.mockResolvedValue({ data: response })

    await expect(dailyCheckInsAPI.list({ page: 1, page_size: 20 })).resolves.toEqual(response)
    expect(get).toHaveBeenCalledWith('/admin/daily-check-ins', {
      params: { page: 1, page_size: 20 }
    })
  })

  it('forwards explicit history and exact service-date filters', async () => {
    get.mockResolvedValue({ data: { items: [], total: 0, page: 1, page_size: 20, pages: 1 } })

    await dailyCheckInsAPI.list({ page: 2, page_size: 50, q: 'alice', all: true })
    expect(get).toHaveBeenLastCalledWith('/admin/daily-check-ins', {
      params: { page: 2, page_size: 50, q: 'alice', all: true }
    })

    await dailyCheckInsAPI.list({ page: 1, page_size: 20, service_date: '2026-07-20' })
    expect(get).toHaveBeenLastCalledWith('/admin/daily-check-ins', {
      params: { page: 1, page_size: 20, service_date: '2026-07-20' }
    })
  })
})
