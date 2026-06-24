<template>
  <div :class="props.embedded ? 'space-y-4' : 'card'">
    <div
      v-if="!props.embedded"
      class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
    >
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ t('profile.avatar.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('profile.avatar.description') }}
      </p>
    </div>

    <div :class="props.embedded ? 'space-y-3' : 'flex flex-col gap-5 px-6 py-6 sm:flex-row sm:items-start'">
      <div
        :class="props.embedded
          ? 'flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 text-xl font-bold text-white shadow-lg shadow-primary-500/20'
          : 'flex h-24 w-24 shrink-0 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 text-3xl font-bold text-white shadow-lg shadow-primary-500/20'"
      >
        <img
          v-if="avatarPreviewUrl"
          data-testid="profile-avatar-preview"
          :src="avatarPreviewUrl"
          :alt="displayName"
          class="h-full w-full object-cover"
        >
        <span v-else>{{ avatarInitial }}</span>
      </div>

      <div :class="props.embedded ? 'space-y-3' : 'min-w-0 flex-1 space-y-4'">
        <div class="space-y-1">
          <p v-if="props.embedded" class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('profile.avatar.title') }}
          </p>
          <p v-else class="text-sm font-medium text-gray-900 dark:text-white">
            {{ displayName }}
          </p>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('profile.avatar.uploadHint') }}
          </p>
        </div>

        <div
          v-if="showQQAvatarCard"
          class="rounded-lg border border-primary-100 bg-primary-50/60 p-3 dark:border-primary-900/60 dark:bg-primary-950/20"
          data-testid="profile-qq-avatar-card"
        >
          <div v-if="qqAvatarLoading" class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('profile.avatar.qqLoading') }}
          </div>
          <div v-else-if="qqAvatarAvailable" class="flex flex-wrap items-center gap-3">
            <div class="flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-white text-sm font-semibold text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300">
              <img
                data-testid="profile-qq-avatar-preview"
                :src="qqAvatarUrl"
                :alt="t('profile.avatar.qqPreviewAlt')"
                class="h-full w-full object-cover"
                @error="handleQQAvatarImageError"
              >
            </div>
            <div class="min-w-[10rem] flex-1">
              <p class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('profile.avatar.qqTitle') }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ t('profile.avatar.qqHint', { qq: qqAvatarSuggestion?.qq || '' }) }}
              </p>
              <p v-if="qqAvatarError" class="mt-1 text-xs text-red-600 dark:text-red-400">
                {{ qqAvatarError }}
              </p>
            </div>
            <button
              data-testid="profile-qq-avatar-adopt"
              type="button"
              class="btn btn-primary btn-sm"
              :disabled="avatarSaving || qqAvatarCurrent || qqAvatarImageFailed"
              @click="handleQQAvatarAdopt"
            >
              {{ qqAvatarCurrent ? t('profile.avatar.qqCurrent') : t('profile.avatar.qqUseAction') }}
            </button>
          </div>
          <div v-else-if="qqAvatarError" class="text-sm text-red-600 dark:text-red-400">
            {{ qqAvatarError }}
          </div>
          <div v-else-if="qqAvatarUnavailable" class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('profile.avatar.qqUnavailable') }}
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-3">
          <label class="btn btn-secondary btn-sm cursor-pointer">
            <input
              data-testid="profile-avatar-file-input"
              type="file"
              accept="image/*"
              class="hidden"
              @change="handleAvatarFileChange"
            >
            {{ t('profile.avatar.uploadAction') }}
          </label>

          <button
            data-testid="profile-avatar-save"
            type="button"
            class="btn btn-primary btn-sm"
            :disabled="avatarSaving || !avatarDraft"
            @click="handleAvatarSave"
          >
            {{ t('common.save') }}
          </button>

          <button
            data-testid="profile-avatar-delete"
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="avatarSaving"
            @click="handleAvatarDelete"
          >
            {{ t('common.delete') }}
          </button>

          <button
            data-testid="profile-qq-avatar-check"
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="avatarSaving || qqAvatarLoading || !props.user"
            @click="handleQQAvatarCheck"
          >
            {{ t('profile.avatar.qqCheckAction') }}
          </button>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          {{ t('profile.avatar.qqConsentHint') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { userAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import type { QQAvatarSuggestion } from '@/api/user'
import type { User } from '@/types'
import { extractApiErrorMessage } from '@/utils/apiError'
import { resolveUserDisplayName } from '@/utils/userDisplay'

const props = withDefaults(defineProps<{
  user: User | null
  embedded?: boolean
}>(), {
  embedded: false,
})

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const targetAvatarUploadBytes = 20 * 1024
const avatarScaleSteps = [1, 0.92, 0.84, 0.76, 0.68, 0.6, 0.52, 0.44, 0.36]
const avatarQualitySteps = [0.92, 0.84, 0.76, 0.68, 0.6, 0.52, 0.44, 0.36]
const avatarDraft = ref('')
const avatarSaving = ref(false)
const qqAvatarSuggestion = ref<QQAvatarSuggestion | null>(null)
const qqAvatarLoading = ref(false)
const qqAvatarError = ref('')
const qqAvatarImageFailed = ref(false)
const qqAvatarChecked = ref(false)

const displayName = computed(() => resolveUserDisplayName(props.user, t('profile.user')))
const avatarInitial = computed(() => displayName.value.charAt(0).toUpperCase() || 'U')
const currentAvatarUrl = computed(() => {
  if (authStore.user && props.user && authStore.user.id === props.user.id) {
    return authStore.user.avatar_url?.trim() || props.user.avatar_url?.trim() || ''
  }
  return props.user?.avatar_url?.trim() || ''
})
const avatarPreviewUrl = computed(() => avatarDraft.value.trim() || currentAvatarUrl.value)
const qqAvatarUrl = computed(() => qqAvatarSuggestion.value?.avatar_url?.trim() || '')
const qqAvatarAvailable = computed(() => Boolean(qqAvatarSuggestion.value?.available && qqAvatarUrl.value))
const qqAvatarUnavailable = computed(() => qqAvatarChecked.value && !qqAvatarLoading.value && !qqAvatarError.value && !qqAvatarAvailable.value)
const showQQAvatarCard = computed(() => qqAvatarLoading.value || Boolean(qqAvatarError.value) || qqAvatarAvailable.value || qqAvatarUnavailable.value)
const qqAvatarCurrent = computed(() => currentAvatarUrl.value !== '' && currentAvatarUrl.value === qqAvatarUrl.value)

watch(
  () => props.user?.avatar_url,
  () => {
    avatarDraft.value = ''
  }
)

watch(
  () => props.user?.email,
  () => {
    resetQQAvatarSuggestion()
  }
)

function normalizeUploadedAvatar(value: string): string | null {
  const normalized = value.trim()
  if (!normalized) {
    return null
  }

  if (!/^data:image\/[a-zA-Z0-9.+-]+;base64,/i.test(normalized)) {
    appStore.showError(t('profile.avatar.uploadRequired'))
    return null
  }

  return normalized
}

function resetQQAvatarSuggestion() {
  qqAvatarSuggestion.value = null
  qqAvatarError.value = ''
  qqAvatarImageFailed.value = false
  qqAvatarChecked.value = false
}

async function loadQQAvatarSuggestion() {
  if (!props.user) {
    resetQQAvatarSuggestion()
    return
  }

  qqAvatarChecked.value = true
  qqAvatarLoading.value = true
  qqAvatarError.value = ''
  qqAvatarImageFailed.value = false
  try {
    const suggestion = await userAPI.getQQAvatarSuggestion()
    qqAvatarSuggestion.value = suggestion.available ? suggestion : null
  } catch (error: unknown) {
    qqAvatarSuggestion.value = null
    qqAvatarError.value = extractApiErrorMessage(error, t('profile.avatar.qqLoadFailed'))
  } finally {
    qqAvatarLoading.value = false
  }
}

async function saveAvatarURL(avatarURL: string) {
  avatarSaving.value = true
  try {
    const updated = await userAPI.updateProfile({ avatar_url: avatarURL })
    authStore.user = updated
    avatarDraft.value = avatarURL.startsWith('data:image/') ? updated.avatar_url?.trim() || '' : ''
    await authStore.refreshUser?.().catch(() => undefined)
    appStore.showSuccess(t('profile.avatar.saveSuccess'))
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    avatarSaving.value = false
  }
}

function readFileAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(typeof reader.result === 'string' ? reader.result : '')
    reader.onerror = () => reject(reader.error ?? new Error('avatar_read_failed'))
    reader.readAsDataURL(file)
  })
}

function loadImage(dataURL: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error(t('profile.avatar.readFailed')))
    image.src = dataURL
  })
}

function canvasToBlob(canvas: HTMLCanvasElement, type: string, quality: number): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (!blob) {
        reject(new Error(t('profile.avatar.compressFailed')))
        return
      }
      resolve(blob)
    }, type, quality)
  })
}

async function compressAvatarFile(file: File): Promise<File> {
  const sourceDataURL = await readFileAsDataURL(file)
  const image = await loadImage(sourceDataURL)
  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    throw new Error(t('profile.avatar.compressFailed'))
  }

  for (const scale of avatarScaleSteps) {
    const width = Math.max(1, Math.round(image.naturalWidth * scale))
    const height = Math.max(1, Math.round(image.naturalHeight * scale))
    canvas.width = width
    canvas.height = height
    ctx.clearRect(0, 0, width, height)
    ctx.drawImage(image, 0, 0, width, height)

    for (const quality of avatarQualitySteps) {
      const blob = await canvasToBlob(canvas, 'image/webp', quality)
      if (blob.size <= targetAvatarUploadBytes) {
        const fileName = file.name.replace(/\.[^.]+$/, '') || 'avatar'
        return new File([blob], `${fileName}.webp`, { type: 'image/webp' })
      }
    }
  }

  throw new Error(t('profile.avatar.compressTooLarge'))
}

async function prepareAvatarUpload(file: File): Promise<File> {
  if (!file.type.startsWith('image/')) {
    throw new Error(t('profile.avatar.invalidType'))
  }
  if (file.type === 'image/gif') {
    if (file.size > targetAvatarUploadBytes) {
      throw new Error(t('profile.avatar.gifTooLarge'))
    }
    return file
  }
  if (file.size <= targetAvatarUploadBytes) {
    return file
  }
  return compressAvatarFile(file)
}

function handleQQAvatarImageError() {
  qqAvatarImageFailed.value = true
  qqAvatarError.value = t('profile.avatar.qqImageFailed')
}

async function handleQQAvatarCheck() {
  await loadQQAvatarSuggestion()
}

async function handleQQAvatarAdopt() {
  if (avatarSaving.value || !qqAvatarAvailable.value || qqAvatarImageFailed.value) {
    return
  }
  await saveAvatarURL(qqAvatarUrl.value)
}

async function handleAvatarFileChange(event: Event) {
  const input = event.target as HTMLInputElement | null
  const file = input?.files?.[0]
  if (input) {
    input.value = ''
  }
  if (!file) {
    return
  }

  try {
    const preparedFile = await prepareAvatarUpload(file)
    const dataURL = await readFileAsDataURL(preparedFile)
    const normalized = normalizeUploadedAvatar(dataURL)
    if (!normalized) {
      return
    }
    avatarDraft.value = normalized
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  }
}

async function handleAvatarSave() {
  const normalized = normalizeUploadedAvatar(avatarDraft.value)
  if (!normalized) {
    return
  }

  await saveAvatarURL(normalized)
}

async function handleAvatarDelete() {
  if (avatarSaving.value) {
    return
  }
  if (!avatarDraft.value.trim() && !currentAvatarUrl.value) {
    appStore.showError(t('profile.avatar.emptyDeleteHint'))
    return
  }

  avatarSaving.value = true
  try {
    const updated = await userAPI.updateProfile({ avatar_url: '' })
    authStore.user = updated
    avatarDraft.value = ''
    appStore.showSuccess(t('profile.avatar.deleteSuccess'))
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    avatarSaving.value = false
  }
}
</script>
