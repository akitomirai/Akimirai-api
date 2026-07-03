import { apiClient } from '../client'

export interface ChatGPT2APIRegisterConfig {
  enabled: boolean
  mail: Record<string, unknown>
  proxy?: string
  total: number
  threads: number
  mode: 'total' | 'quota' | 'available' | string
  target_quota?: number
  target_available?: number
  check_interval?: number
  stats?: Record<string, unknown>
}

export interface ChatGPT2APILog {
  id: number | string
  type?: string
  level: string
  time: string
  summary: string
  message?: string
  detail?: Record<string, unknown>
  created_at?: string
  timestamp?: string
}

export interface ChatGPT2APILogQuery {
  type?: string
  limit?: number
}

async function getRegister(): Promise<ChatGPT2APIRegisterConfig> {
  const { data } = await apiClient.get<ChatGPT2APIRegisterConfig>('/admin/chatgpt2api/register')
  return data
}

async function updateRegister(payload: ChatGPT2APIRegisterConfig): Promise<ChatGPT2APIRegisterConfig> {
  const { data } = await apiClient.put<ChatGPT2APIRegisterConfig>('/admin/chatgpt2api/register', payload)
  return data
}

async function startRegister(): Promise<ChatGPT2APIRegisterConfig> {
  const { data } = await apiClient.post<ChatGPT2APIRegisterConfig>('/admin/chatgpt2api/register/start')
  return data
}

async function stopRegister(): Promise<ChatGPT2APIRegisterConfig> {
  const { data } = await apiClient.post<ChatGPT2APIRegisterConfig>('/admin/chatgpt2api/register/stop')
  return data
}

async function resetRegister(): Promise<ChatGPT2APIRegisterConfig> {
  const { data } = await apiClient.post<ChatGPT2APIRegisterConfig>('/admin/chatgpt2api/register/reset')
  return data
}

async function resetOutlookPool(): Promise<ChatGPT2APIRegisterConfig> {
  const { data } = await apiClient.post<ChatGPT2APIRegisterConfig>('/admin/chatgpt2api/outlook/reset')
  return data
}

async function listLogs(params: ChatGPT2APILogQuery = {}): Promise<ChatGPT2APILog[]> {
  const { data } = await apiClient.get<ChatGPT2APILog[]>('/admin/chatgpt2api/logs', { params })
  return data
}

export default {
  getRegister,
  updateRegister,
  startRegister,
  stopRegister,
  resetRegister,
  resetOutlookPool,
  listLogs,
}
