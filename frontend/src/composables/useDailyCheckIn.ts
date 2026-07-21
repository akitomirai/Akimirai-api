import {
  computed,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
  type ComputedRef,
  type Ref,
} from 'vue'

import { checkInAPI, type DailyCheckInResponse } from '@/api/checkin'
import { i18n } from '@/i18n'
import { useAppStore, useAuthStore } from '@/stores'

export type DailyCheckInPhase = 'loading' | 'available' | 'claiming' | 'claimed' | 'error'

interface DailyCheckInAPI {
  getStatus(): Promise<DailyCheckInResponse>
  claim(): Promise<DailyCheckInResponse>
}

interface DailyCheckInAuthStore {
  user: { id: number } | null
  refreshUser(): Promise<unknown>
}

interface DailyCheckInAppStore {
  showSuccess(message: string, duration?: number): unknown
  showWarning(message: string, duration?: number): unknown
}

export interface UseDailyCheckInOptions {
  api?: DailyCheckInAPI
  authStore?: DailyCheckInAuthStore
  appStore?: DailyCheckInAppStore
  document?: Document
}

export interface DailyCheckInState {
  phase: Ref<DailyCheckInPhase>
  status: Ref<DailyCheckInResponse | null>
  lastError: Ref<string | null>
  canClaim: ComputedRef<boolean>
  refreshStatus(options?: { force?: boolean }): Promise<DailyCheckInResponse | null>
  handleAction(): Promise<void>
  start(): void
  stop(): void
}

const INVALID_RESET_RETRY_MS = 30_000
const RESET_TIMER_GRACE_MS = 250
const MAX_TIMEOUT_MS = 2_147_000_000

function errorMessage(error: unknown): string {
  if (error instanceof Error && error.message) return error.message
  if (error && typeof error === 'object' && 'message' in error) {
    const message = String((error as { message?: unknown }).message || '')
    if (message) return message
  }
  return 'Daily check-in request failed'
}

function rewardLabel(amount: number): string {
  return Number.isInteger(amount) ? String(amount) : amount.toFixed(2)
}

export function useDailyCheckIn(options: UseDailyCheckInOptions = {}): DailyCheckInState {
  const api = options.api ?? checkInAPI
  const authStore = options.authStore ?? useAuthStore()
  const appStore = options.appStore ?? useAppStore()
  const documentRef = options.document ?? (typeof document === 'undefined' ? undefined : document)

  const phase = ref<DailyCheckInPhase>('loading')
  const status = ref<DailyCheckInResponse | null>(null)
  const lastError = ref<string | null>(null)
  const canClaim = computed(() => phase.value === 'available')

  let statusRequest: Promise<DailyCheckInResponse> | null = null
  let requestSequence = 0
  let resetTimer: ReturnType<typeof setTimeout> | undefined
  let started = false

  const currentUserID = () => authStore.user?.id ?? 0

  function clearResetTimer() {
    if (resetTimer !== undefined) {
      clearTimeout(resetTimer)
      resetTimer = undefined
    }
  }

  function scheduleReset(nextResetAt: string) {
    clearResetTimer()
    const resetAt = Date.parse(nextResetAt)
    let delay = resetAt - Date.now() + RESET_TIMER_GRACE_MS
    if (!Number.isFinite(delay) || delay <= 0) {
      delay = INVALID_RESET_RETRY_MS
    }
    delay = Math.min(delay, MAX_TIMEOUT_MS)
    resetTimer = setTimeout(() => {
      resetTimer = undefined
      if (documentRef?.hidden) return
      void refreshStatus({ force: true })
    }, delay)
  }

  function applyStatus(nextStatus: DailyCheckInResponse) {
    status.value = nextStatus
    lastError.value = null
    phase.value = nextStatus.checked_in ? 'claimed' : 'available'
    scheduleReset(nextStatus.next_reset_at)
  }

  async function requestStatus(
    requestOptions: { force?: boolean; showLoading?: boolean } = {},
  ): Promise<DailyCheckInResponse> {
    if (statusRequest && !requestOptions.force) return statusRequest

    const userID = currentUserID()
    if (!userID) throw new Error('Not authenticated')
    const sequence = ++requestSequence
    if (requestOptions.showLoading !== false && phase.value !== 'claimed') {
      phase.value = 'loading'
    }

    const request = api.getStatus()
      .then((nextStatus) => {
        if (sequence === requestSequence && currentUserID() === userID) {
          applyStatus(nextStatus)
        }
        return nextStatus
      })
      .catch((error: unknown) => {
        if (sequence === requestSequence && currentUserID() === userID) {
          lastError.value = errorMessage(error)
          phase.value = 'error'
          clearResetTimer()
        }
        throw error
      })
      .finally(() => {
        if (statusRequest === request) statusRequest = null
      })
    statusRequest = request
    return request
  }

  async function refreshStatus(
    requestOptions: { force?: boolean } = {},
  ): Promise<DailyCheckInResponse | null> {
    if (!currentUserID()) return null
    try {
      return await requestStatus({ force: requestOptions.force })
    } catch {
      return null
    }
  }

  async function refreshProfileBestEffort() {
    try {
      await authStore.refreshUser()
    } catch {
      appStore.showWarning(String(i18n.global.t('checkIn.profileRefreshWarning')))
    }
  }

  async function announceClaim(nextStatus: DailyCheckInResponse) {
    appStore.showSuccess(String(i18n.global.t('checkIn.success', {
      amount: rewardLabel(nextStatus.reward_amount),
    })))
    await refreshProfileBestEffort()
  }

  async function claim() {
    if (phase.value !== 'available') return
    const userID = currentUserID()
    if (!userID) return

    ++requestSequence
    clearResetTimer()
    phase.value = 'claiming'
    lastError.value = null
    try {
      const claimedStatus = await api.claim()
      if (currentUserID() !== userID) return
      applyStatus(claimedStatus)
      if (claimedStatus.checked_in) {
        await announceClaim(claimedStatus)
      }
      return
    } catch (error: unknown) {
      lastError.value = errorMessage(error)
    }

    try {
      const reconciled = await requestStatus({ force: true, showLoading: false })
      if (currentUserID() === userID && reconciled.checked_in) {
        await announceClaim(reconciled)
      }
    } catch {
      // requestStatus owns the visible error state. A new POST is blocked until GET succeeds.
    }
  }

  async function handleAction() {
    switch (phase.value) {
      case 'available':
        await claim()
        return
      case 'error':
        await refreshStatus({ force: true })
        return
      case 'loading':
      case 'claiming':
      case 'claimed':
        return
    }
  }

  function handleVisibilityChange() {
    if (documentRef?.hidden || phase.value === 'claiming' || !currentUserID()) return
    void refreshStatus({ force: true })
  }

  function start() {
    if (started) return
    started = true
    documentRef?.addEventListener('visibilitychange', handleVisibilityChange)
    if (currentUserID()) void refreshStatus({ force: true })
  }

  function stop() {
    if (!started) return
    started = false
    clearResetTimer()
    documentRef?.removeEventListener('visibilitychange', handleVisibilityChange)
    ++requestSequence
    statusRequest = null
  }

  watch(currentUserID, (nextUserID, previousUserID) => {
    if (nextUserID === previousUserID) return
    ++requestSequence
    statusRequest = null
    clearResetTimer()
    status.value = null
    lastError.value = null
    phase.value = 'loading'
    if (started && nextUserID) void refreshStatus({ force: true })
  })

  onMounted(start)
  onBeforeUnmount(stop)

  return {
    phase,
    status,
    lastError,
    canClaim,
    refreshStatus,
    handleAction,
    start,
    stop,
  }
}
