<template>
  <section class="card flex h-full min-h-[360px] flex-col overflow-hidden">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('announcements.title') }}</h2>
        <span
          v-if="unreadCount > 0"
          class="inline-flex items-center rounded-full bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
        >
          {{ t('announcements.newCount', { count: unreadCount }) }}
        </span>
      </div>
    </div>

    <div v-if="loading && dashboardAnnouncements.length === 0" class="flex flex-1 items-center justify-center p-6 text-sm text-gray-500 dark:text-dark-400">
      {{ t('common.loading') }}
    </div>

    <div v-else-if="dashboardAnnouncements.length > 0" class="flex-1 px-6 py-5">
      <div class="relative space-y-5">
        <div class="absolute bottom-2 left-[5px] top-2 w-px bg-gray-200 dark:bg-dark-700"></div>

        <article
          v-for="item in dashboardAnnouncements"
          :key="item.id"
          class="relative grid grid-cols-[12px_minmax(0,1fr)] gap-4"
        >
          <span
            class="relative z-10 mt-1 h-3 w-3 rounded-full border-2 border-white dark:border-dark-900"
            :class="item.read_at ? 'bg-gray-300 dark:bg-dark-600' : 'bg-primary-500 shadow-[0_0_0_4px_rgba(20,184,166,0.12)]'"
          ></span>
          <div class="min-w-0 pb-1">
            <div class="flex min-w-0 items-center gap-2">
              <h3 class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ item.title }}</h3>
              <span
                v-if="!item.read_at"
                class="h-1.5 w-1.5 flex-shrink-0 rounded-full bg-primary-500"
                :aria-label="t('announcements.unread')"
              ></span>
            </div>
            <p
              v-if="announcementPreview(item.content)"
              class="mt-1 line-clamp-2 text-xs leading-5 text-gray-500 dark:text-dark-400"
            >
              {{ announcementPreview(item.content) }}
            </p>
            <time class="mt-2 block text-xs text-gray-400 dark:text-dark-500">
              {{ formatRelativeTime(item.created_at) }}
            </time>
          </div>
        </article>
      </div>
    </div>

    <div v-else class="flex flex-1 flex-col items-center justify-center p-6 text-center">
      <div class="mb-3 flex h-14 w-14 items-center justify-center rounded-2xl bg-gray-100 text-gray-400 dark:bg-dark-800 dark:text-dark-500">
        <Icon name="inbox" size="lg" />
      </div>
      <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('announcements.empty') }}</p>
      <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('announcements.emptyDescription') }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeTime } from '@/utils/format'

const { t } = useI18n()
const announcementStore = useAnnouncementStore()
const { announcements, loading } = storeToRefs(announcementStore)

const unreadCount = computed(() => announcementStore.unreadCount)
const dashboardAnnouncements = computed(() => announcements.value.slice(0, 4))

function announcementPreview(content: string): string {
  return content
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/!\[[^\]]*]\([^)]+\)/g, ' ')
    .replace(/\[([^\]]+)]\([^)]+\)/g, '$1')
    .replace(/<[^>]+>/g, ' ')
    .replace(/[#>*_~`-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

onMounted(() => {
  void announcementStore.fetchAnnouncements()
})
</script>
