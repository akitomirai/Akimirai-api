import { apiClient } from './client'

export interface DailyCheckInResponse {
  checked_in: boolean
  already_checked_in: boolean
  service_date: string
  reward_amount: number
  balance_before: number
  balance_after: number
  checked_in_at: string | null
  next_reset_at: string
}

export const checkInAPI = {
  async getStatus(): Promise<DailyCheckInResponse> {
    const { data } = await apiClient.get<DailyCheckInResponse>('/user/check-in')
    return data
  },

  async claim(): Promise<DailyCheckInResponse> {
    const { data } = await apiClient.post<DailyCheckInResponse>('/user/check-in')
    return data
  },
}
