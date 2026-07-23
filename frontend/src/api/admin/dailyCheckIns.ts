import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export interface DailyCheckInAdminRecord {
  id: number
  user_id: number
  email: string
  username: string
  service_date: string
  reward_amount: number
  balance_before: number
  balance_after: number
  checked_in_at: string
  created_at: string
}

export interface DailyCheckInAdminQuery {
  page?: number
  page_size?: number
  q?: string
  service_date?: string
  all?: boolean
}

export type DailyCheckInAdminListResponse = PaginatedResponse<DailyCheckInAdminRecord>

export async function list(
  params: DailyCheckInAdminQuery = {}
): Promise<DailyCheckInAdminListResponse> {
  const { data } = await apiClient.get<DailyCheckInAdminListResponse>('/admin/daily-check-ins', { params })
  return data
}

export const dailyCheckInsAPI = { list }

export default dailyCheckInsAPI
