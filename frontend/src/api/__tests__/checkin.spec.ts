import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: { get, post },
}))

import { checkInAPI } from '@/api/checkin'

const payload = {
  checked_in: true,
  already_checked_in: false,
  service_date: '2026-07-21',
  reward_amount: 2,
  balance_before: 10,
  balance_after: 12,
  checked_in_at: '2026-07-21T03:04:05+08:00',
  next_reset_at: '2026-07-22T02:00:00+08:00',
}

describe('check-in API', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
  })

  it('loads the server-owned service-day status', async () => {
    get.mockResolvedValue({ data: payload })

    await expect(checkInAPI.getStatus()).resolves.toEqual(payload)
    expect(get).toHaveBeenCalledWith('/user/check-in')
  })

  it('claims without a client-computed date or request body', async () => {
    post.mockResolvedValue({ data: payload })

    await expect(checkInAPI.claim()).resolves.toEqual(payload)
    expect(post).toHaveBeenCalledWith('/user/check-in')
  })
})
