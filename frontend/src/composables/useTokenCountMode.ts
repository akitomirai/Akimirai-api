import { readonly, ref } from 'vue'

export type TokenCountMode = 'modern' | 'legacy'

export const TOKEN_COUNT_MODE_STORAGE_KEY = 'token-count-mode'

const mode = ref<TokenCountMode>('modern')
let initialized = false
let storageListenerAttached = false

const normalizeMode = (value: unknown): TokenCountMode =>
  value === 'legacy' ? 'legacy' : 'modern'

const readStoredMode = (): TokenCountMode => {
  if (typeof window === 'undefined') return 'modern'
  try {
    return normalizeMode(window.localStorage.getItem(TOKEN_COUNT_MODE_STORAGE_KEY))
  } catch {
    return 'modern'
  }
}

const handleStorage = (event: StorageEvent) => {
  if (event.key !== TOKEN_COUNT_MODE_STORAGE_KEY) return
  mode.value = normalizeMode(event.newValue)
}

const initialize = () => {
  if (!initialized) {
    mode.value = readStoredMode()
    initialized = true
  }
  if (typeof window !== 'undefined' && !storageListenerAttached) {
    window.addEventListener('storage', handleStorage)
    storageListenerAttached = true
  }
}

export const getTokenCountMode = (): TokenCountMode => {
  initialize()
  return mode.value
}

export const setTokenCountMode = (nextMode: TokenCountMode): void => {
  initialize()
  mode.value = nextMode
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(TOKEN_COUNT_MODE_STORAGE_KEY, nextMode)
  } catch {
    // Browser privacy modes can reject storage; the in-memory preference still works.
  }
}

export const useTokenCountMode = () => {
  initialize()
  return {
    mode: readonly(mode),
    setMode: setTokenCountMode,
  }
}

export const resetTokenCountModeForTests = (): void => {
  mode.value = 'modern'
  initialized = false
  if (typeof window !== 'undefined' && storageListenerAttached) {
    window.removeEventListener('storage', handleStorage)
    storageListenerAttached = false
  }
}
