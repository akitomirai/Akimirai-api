<template>
  <AppLayout>
    <div class="image-management-page">
      <header class="page-header">
        <div>
          <p class="text-sm font-medium text-primary-600 dark:text-primary-400">Images</p>
          <h1 class="mt-1 text-2xl font-semibold text-gray-950 dark:text-white">我的作品</h1>
        </div>
        <div class="header-actions">
          <router-link class="soft-button" to="/images">
            <Icon name="sparkles" size="sm" />
            <span>画图</span>
          </router-link>
          <button class="soft-button" type="button" :disabled="loadingWorks" @click="loadWorks">
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': loadingWorks }" />
            <span>刷新</span>
          </button>
        </div>
      </header>

      <section class="toolbar" aria-label="作品筛选">
        <div class="search-control">
          <Icon name="search" size="md" class="text-gray-400 dark:text-dark-300" />
          <input
            v-model="searchQuery"
            type="search"
            placeholder="搜索提示词、模型、尺寸"
            aria-label="搜索图片"
          />
        </div>
      </section>

      <section class="content-panel">
        <div class="panel-toolbar">
          <div class="panel-count">
            <Icon name="grid" size="sm" />
            <span>共 {{ worksTotal }} 个作品</span>
          </div>
          <div class="pager">
            <button type="button" :disabled="currentPage <= 1 || loadingWorks" @click="goToPage(currentPage - 1)">
              <Icon name="chevronLeft" size="sm" />
            </button>
            <span>{{ currentPage }} / {{ currentPages }}</span>
            <button type="button" :disabled="currentPage >= currentPages || loadingWorks" @click="goToPage(currentPage + 1)">
              <Icon name="chevronRight" size="sm" />
            </button>
          </div>
        </div>

        <div v-if="loadingWorks" class="loading-state">
          <Icon name="refresh" size="lg" class="animate-spin text-primary-500" />
          <span>正在加载图片</span>
        </div>

        <div v-else class="panel-body">
          <div v-if="filteredWorks.length > 0" class="masonry-grid">
            <article v-for="work in filteredWorks" :key="work.id" class="image-card">
              <button type="button" class="image-preview" @click="openDetail(toWorkDetail(work))">
                <img v-if="imageSource(work)" :src="imageSource(work)" :alt="work.prompt" />
                <span v-else><Icon name="grid" size="xl" /></span>
              </button>
              <div class="card-body">
                <p class="prompt-text">{{ work.revised_prompt || work.prompt || '无提示词' }}</p>
                <div class="meta-line">
                  <span>{{ work.model || '-' }}</span>
                  <span>{{ work.size || '-' }}</span>
                  <span>{{ work.output_format || '-' }}</span>
                  <span>{{ formatDateTime(work.created_at) || '-' }}</span>
                </div>
                <div class="card-actions">
                  <button type="button" title="查看" @click="openDetail(toWorkDetail(work))">
                    <Icon name="eye" size="sm" />
                  </button>
                  <button type="button" title="复制提示词" @click="copyPrompt(work.revised_prompt || work.prompt)">
                    <Icon name="copy" size="sm" />
                  </button>
                  <button type="button" title="用此图二创" :disabled="!imageSource(work)" @click="useAsReference(toWorkDetail(work))">
                    <Icon name="edit" size="sm" />
                  </button>
                  <button type="button" title="下载" :disabled="!imageSource(work)" @click="downloadDetail(toWorkDetail(work))">
                    <Icon name="download" size="sm" />
                  </button>
                  <button
                    type="button"
                    class="danger-icon"
                    title="删除作品"
                    :disabled="isActing(`work-delete-${work.id}`)"
                    @click="deleteWork(work)"
                  >
                    <Icon name="trash" size="sm" />
                  </button>
                </div>
              </div>
            </article>
          </div>
          <EmptyState v-else :message="works.length > 0 ? '没有匹配的作品' : '暂无作品'" />
        </div>
      </section>

      <div v-if="selectedDetail" class="modal-backdrop" @click.self="selectedDetail = null">
        <section class="detail-modal" role="dialog" aria-modal="true" aria-label="图片详情">
          <header class="detail-header">
            <div>
              <p class="text-sm font-medium text-primary-600 dark:text-primary-400">我的作品</p>
              <h2 class="mt-1 text-lg font-semibold text-gray-950 dark:text-white">图片详情</h2>
            </div>
            <button type="button" class="icon-button" title="关闭" @click="selectedDetail = null">
              <Icon name="x" size="sm" />
            </button>
          </header>
          <div class="detail-body">
            <div class="detail-image">
              <img v-if="selectedDetail.src" :src="selectedDetail.src" :alt="selectedDetail.prompt" />
              <Icon v-else name="grid" size="xl" />
            </div>
            <div class="detail-info">
              <p class="detail-prompt">{{ selectedDetail.prompt || '无提示词' }}</p>
              <dl class="detail-meta">
                <div><dt>模型</dt><dd>{{ selectedDetail.model || '-' }}</dd></div>
                <div><dt>尺寸</dt><dd>{{ selectedDetail.size || '-' }}</dd></div>
                <div><dt>时间</dt><dd>{{ formatDateTime(selectedDetail.createdAt) || '-' }}</dd></div>
              </dl>
              <div class="detail-actions">
                <button class="btn btn-secondary" type="button" @click="copyPrompt(selectedDetail.prompt)"><Icon name="copy" size="sm" />复制提示词</button>
                <button class="btn btn-secondary" type="button" :disabled="!selectedDetail.src" @click="openImage(selectedDetail.src)"><Icon name="externalLink" size="sm" />打开</button>
                <button class="btn btn-secondary" type="button" :disabled="!selectedDetail.src" @click="useAsReference(selectedDetail)"><Icon name="edit" size="sm" />二创</button>
                <button class="btn btn-primary" type="button" :disabled="!selectedDetail.src" @click="downloadDetail(selectedDetail)"><Icon name="download" size="sm" />下载</button>
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'

interface ImageDetail {
  id: string
  src: string
  prompt: string
  model?: string
  size?: string
  createdAt: string
  fileName: string
}

interface ImageWork {
  id: string
  prompt: string
  revised_prompt?: string
  model?: string
  size?: string
  quality?: string
  output_format?: string
  image_url?: string
  b64_json?: string
  mime_type?: string
  created_at: string
  source_turn_id?: string
  source_image_id?: string
}

interface StoredImage {
  id: string
  status: string
  src?: string
  b64_json?: string
  url?: string
  mime_type?: string
}

interface StoredTurn {
  id: string
  prompt?: string
  model?: string
  size?: string
  resolution?: string
  outputFormat?: string
  createdAt?: string
  images?: StoredImage[]
}

interface StoredConversation {
  turns?: StoredTurn[]
}

const PAGE_SIZE = 24
const LOCAL_HISTORY_KEY = 'image_generation_conversations:v2'
const REDRAW_HANDOFF_KEY = 'image_generation_redraw:v1'

const appStore = useAppStore()
const router = useRouter()

const searchQuery = ref('')
const works = ref<ImageWork[]>([])
const allWorks = ref<ImageWork[]>([])
const worksPage = ref(1)
const worksTotal = ref(0)
const worksPages = ref(1)
const loadingWorks = ref(false)
const actionKey = ref('')
const selectedDetail = ref<ImageDetail | null>(null)

const currentPage = computed(() => worksPage.value)
const currentPages = computed(() => worksPages.value)

const filteredWorks = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return works.value
  return works.value.filter((work) => {
    return [
      work.prompt,
      work.revised_prompt,
      work.model,
      work.size,
      work.quality,
      work.output_format,
    ].some((value) => String(value || '').toLowerCase().includes(q))
  })
})

const EmptyState = defineComponent({
  name: 'ImageManagementEmptyState',
  props: { message: { type: String, required: true } },
  setup(props) {
    return () => h('div', { class: 'empty-state' }, [
      h('div', { class: 'empty-icon' }, [h(Icon, { name: 'grid', size: 'xl' })]),
      h('p', props.message),
    ])
  },
})

async function loadWorks() {
  loadingWorks.value = true
  try {
    allWorks.value = readLocalWorks()
    worksTotal.value = allWorks.value.length
    worksPages.value = Math.max(1, Math.ceil(worksTotal.value / PAGE_SIZE))
    worksPage.value = Math.min(Math.max(1, worksPage.value), worksPages.value)
    const start = (worksPage.value - 1) * PAGE_SIZE
    works.value = allWorks.value.slice(start, start + PAGE_SIZE)
  } catch {
    allWorks.value = []
    works.value = []
    worksTotal.value = 0
    worksPages.value = 1
    worksPage.value = 1
    appStore.showError('读取本地作品失败')
  } finally {
    loadingWorks.value = false
  }
}

async function goToPage(page: number) {
  worksPage.value = Math.min(Math.max(1, page), worksPages.value)
  await loadWorks()
}

function imageSource(item: Pick<ImageWork, 'image_url' | 'b64_json' | 'mime_type'>): string {
  if (item.image_url) return item.image_url
  if (item.b64_json) return `data:${item.mime_type || 'image/png'};base64,${item.b64_json}`
  return ''
}

function toWorkDetail(work: ImageWork): ImageDetail {
  return {
    id: work.id,
    src: imageSource(work),
    prompt: work.revised_prompt || work.prompt || '',
    model: work.model,
    size: work.size,
    createdAt: work.created_at,
    fileName: safeFileName(work.prompt || `work-${work.id}`),
  }
}

function openDetail(detail: ImageDetail) { selectedDetail.value = detail }

function openImage(src: string) {
  if (!src) return
  window.open(src, '_blank', 'noopener,noreferrer')
}

async function copyPrompt(prompt: string) {
  try {
    await navigator.clipboard.writeText(prompt || '')
    appStore.showSuccess('已复制提示词')
  } catch { appStore.showError('复制失败') }
}

function useAsReference(detail: ImageDetail) {
  if (!detail.src) { appStore.showError('当前图片没有可用地址'); return }
  sessionStorage.setItem(REDRAW_HANDOFF_KEY, JSON.stringify({ src: detail.src, prompt: detail.prompt, name: `${detail.fileName}.png` }))
  void router.push('/images')
}

async function downloadDetail(detail: ImageDetail) {
  if (!detail.src) { appStore.showError('当前图片没有可用地址'); return }
  try {
    const blob = await sourceToBlob(detail.src)
    const url = URL.createObjectURL(blob)
    triggerDownload(url, `${detail.fileName}.${extensionFromBlob(blob)}`)
    URL.revokeObjectURL(url)
  } catch { triggerDownload(detail.src, `${detail.fileName}.png`) }
}

async function deleteWork(work: ImageWork) {
  if (!window.confirm('确认删除这个作品吗？')) return
  await withAction(`work-delete-${work.id}`, async () => {
    removeWorkFromLocalHistory(work)
    allWorks.value = allWorks.value.filter((item) => item.id !== work.id)
    works.value = works.value.filter((item) => item.id !== work.id)
    worksTotal.value = Math.max(0, worksTotal.value - 1)
    worksPages.value = Math.max(1, Math.ceil(worksTotal.value / PAGE_SIZE))
    appStore.showSuccess('作品已删除')
  })
}

async function withAction(key: string, fn: () => Promise<void>) {
  if (actionKey.value) return
  actionKey.value = key
  try { await fn() }
  catch { appStore.showError('操作失败') }
  finally { actionKey.value = '' }
}

function isActing(key: string): boolean { return actionKey.value === key }

function safeFileName(value: string): string {
  return value.trim().replace(/[\\/:*?"<>|]+/g, '-').slice(0, 48) || 'image'
}

async function sourceToBlob(src: string): Promise<Blob> {
  const response = await fetch(src)
  if (!response.ok) throw new Error('download failed')
  return response.blob()
}

function extensionFromBlob(blob: Blob): string {
  if (blob.type === 'image/jpeg') return 'jpg'
  if (blob.type === 'image/webp') return 'webp'
  return 'png'
}

function triggerDownload(url: string, filename: string) {
  const anchor = document.createElement('a')
  anchor.href = url; anchor.download = filename; anchor.rel = 'noopener noreferrer'
  document.body.appendChild(anchor); anchor.click(); document.body.removeChild(anchor)
}

function readLocalWorks(): ImageWork[] {
  const raw = localStorage.getItem(LOCAL_HISTORY_KEY)
  if (!raw) return []
  const conversations = JSON.parse(raw) as StoredConversation[]
  return conversations.flatMap((conversation) => {
    return (conversation.turns || []).flatMap((turn) => {
      return (turn.images || [])
        .filter((image) => image.status === 'success' && (image.src || image.url || image.b64_json))
        .map((image) => {
          const id = `${turn.id}:${image.id}`
          return {
            id,
            prompt: turn.prompt || '',
            model: turn.model,
            size: turn.size,
            quality: turn.resolution,
            output_format: turn.outputFormat,
            image_url: image.url || image.src,
            b64_json: image.b64_json,
            mime_type: image.mime_type,
            created_at: turn.createdAt || new Date().toISOString(),
            source_turn_id: turn.id,
            source_image_id: image.id,
          }
        })
    })
  }).sort((a, b) => b.created_at.localeCompare(a.created_at))
}

function removeWorkFromLocalHistory(work: ImageWork) {
  const raw = localStorage.getItem(LOCAL_HISTORY_KEY)
  if (!raw || !work.source_turn_id || !work.source_image_id) return
  const conversations = JSON.parse(raw) as StoredConversation[]
  for (const conversation of conversations) {
    for (const turn of conversation.turns || []) {
      if (turn.id !== work.source_turn_id || !turn.images) continue
      turn.images = turn.images.filter((image) => image.id !== work.source_image_id)
    }
  }
  localStorage.setItem(LOCAL_HISTORY_KEY, JSON.stringify(conversations))
}

onMounted(() => { void loadWorks() })
</script>

<style scoped>
.image-management-page {
  min-height: calc(100vh - 7rem);
  display: grid;
  grid-template-rows: auto auto minmax(34rem, 1fr);
  gap: 1rem;
}

.page-header, .header-actions, .panel-toolbar, .panel-count, .card-actions, .detail-actions, .detail-header, .pager {
  display: flex;
  align-items: center;
}

.page-header { justify-content: space-between; gap: 1rem; }
.header-actions, .card-actions, .detail-actions { gap: 0.625rem; }

.soft-button, .search-control {
  min-height: 2.5rem;
  border-radius: 0.5rem;
  border: 1px solid rgb(229 231 235);
  background: white;
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}

.dark .soft-button, .dark .search-control {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39);
}

.soft-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0 0.875rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(55 65 81);
  white-space: nowrap;
}

.soft-button:hover { border-color: rgb(59 130 246 / 0.5); color: rgb(37 99 235); }
.soft-button:disabled { cursor: not-allowed; opacity: 0.5; }
.dark .soft-button { color: rgb(209 213 219); }

.toolbar { display: grid; grid-template-columns: minmax(0, 1fr); gap: 0.75rem; align-items: center; }

.search-control {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0 0.875rem;
}

.search-control input {
  min-width: 0; flex: 1; border: 0; background: transparent;
  color: rgb(17 24 39); outline: none;
}

.dark .search-control input { color: white; }

.content-panel {
  min-height: 34rem;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 0.5rem;
  border: 1px solid rgb(229 231 235);
  background: white;
  box-shadow: 0 1px 3px rgb(15 23 42 / 0.05);
}

.dark .content-panel { border-color: rgb(55 65 81); background: rgb(17 24 39 / 0.72); }

.panel-toolbar {
  min-height: 4rem; justify-content: space-between; gap: 1rem;
  border-bottom: 1px solid rgb(243 244 246); padding: 0 1rem;
}

.dark .panel-toolbar { border-color: rgb(55 65 81); }
.panel-count { gap: 0.5rem; font-size: 0.875rem; color: rgb(75 85 99); }
.dark .panel-count { color: rgb(209 213 219); }
.pager { gap: 0.5rem; font-size: 0.8125rem; color: rgb(107 114 128); }

.pager button, .icon-button, .card-actions button {
  display: grid; width: 2rem; height: 2rem; place-items: center;
  border-radius: 0.5rem; color: rgb(107 114 128);
}

.pager button:hover, .icon-button:hover, .card-actions button:hover { background: rgb(243 244 246); color: rgb(31 41 55); }
.pager button:disabled, .card-actions button:disabled { cursor: not-allowed; opacity: 0.4; }
.dark .pager button:hover, .dark .icon-button:hover, .dark .card-actions button:hover { background: rgb(55 65 81 / 0.8); color: white; }

.panel-body { flex: 1; overflow: auto; padding: 1rem; }

.masonry-grid { column-count: 4; column-gap: 1rem; }

.image-card {
  break-inside: avoid; overflow: hidden; margin-bottom: 1rem;
  border-radius: 0.5rem; border: 1px solid rgb(229 231 235); background: white;
}

.dark .image-card { border-color: rgb(55 65 81); background: rgb(31 41 55 / 0.42); }

.image-preview {
  display: grid; width: 100%; min-height: 12rem; place-items: center;
  background: rgb(243 244 246); color: rgb(156 163 175);
}

.image-preview img { width: 100%; height: auto; display: block; }
.dark .image-preview { background: rgb(31 41 55); }

.card-body { display: grid; gap: 0.75rem; padding: 0.875rem; }

.prompt-text {
  display: -webkit-box; overflow: hidden; -webkit-box-orient: vertical;
  -webkit-line-clamp: 3; font-size: 0.875rem; line-height: 1.45rem; color: rgb(31 41 55);
}

.dark .prompt-text { color: rgb(229 231 235); }

.meta-line { display: flex; flex-wrap: wrap; gap: 0.375rem; }
.meta-line span {
  min-width: 0; border-radius: 0.375rem; background: rgb(243 244 246);
  padding: 0.1875rem 0.45rem; font-size: 0.75rem; color: rgb(107 114 128);
}
.dark .meta-line span { background: rgb(55 65 81 / 0.8); color: rgb(209 213 219); }

.card-actions { flex-wrap: wrap; justify-content: flex-end; }
.card-actions .danger-icon { color: rgb(220 38 38); }

.loading-state, .empty-state {
  flex: 1; min-height: 30rem; display: grid; place-items: center;
  align-content: center; gap: 0.75rem; color: rgb(107 114 128);
}

.empty-icon {
  display: grid; width: 4rem; height: 4rem; place-items: center;
  border-radius: 0.75rem; border: 1px solid rgb(229 231 235);
  background: rgb(249 250 251); color: rgb(156 163 175);
}

.dark .loading-state, .dark .empty-state { color: rgb(156 163 175); }
.dark .empty-icon { border-color: rgb(55 65 81); background: rgb(31 41 55 / 0.55); }

.modal-backdrop {
  position: fixed; inset: 0; z-index: 50; display: grid;
  place-items: center; background: rgb(15 23 42 / 0.58); padding: 1rem;
}

.detail-modal {
  width: min(64rem, 100%); max-height: min(48rem, calc(100vh - 2rem));
  overflow: hidden; border-radius: 0.5rem; border: 1px solid rgb(229 231 235);
  background: white; box-shadow: 0 24px 80px rgb(15 23 42 / 0.32);
}

.dark .detail-modal { border-color: rgb(55 65 81); background: rgb(17 24 39); }
.detail-header { justify-content: space-between; gap: 1rem; border-bottom: 1px solid rgb(243 244 246); padding: 1rem; }
.dark .detail-header { border-color: rgb(55 65 81); }

.detail-body { display: grid; grid-template-columns: minmax(0, 1.1fr) minmax(18rem, 0.9fr); gap: 1rem; overflow: auto; padding: 1rem; }
.detail-image { display: grid; min-height: 24rem; place-items: center; overflow: hidden; border-radius: 0.5rem; background: rgb(243 244 246); color: rgb(156 163 175); }
.detail-image img { width: 100%; height: 100%; object-fit: contain; }
.dark .detail-image { background: rgb(31 41 55); }
.detail-info { display: grid; align-content: start; gap: 1rem; }
.detail-prompt { white-space: pre-wrap; font-size: 0.925rem; line-height: 1.55rem; color: rgb(31 41 55); }
.dark .detail-prompt { color: rgb(229 231 235); }

.detail-meta { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.75rem; }
.detail-meta div { min-width: 0; border-radius: 0.5rem; background: rgb(249 250 251); padding: 0.75rem; }
.dark .detail-meta div { background: rgb(31 41 55 / 0.55); }
.detail-meta dt { font-size: 0.75rem; font-weight: 700; color: rgb(107 114 128); }
.detail-meta dd { margin-top: 0.25rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 0.875rem; color: rgb(31 41 55); }
.dark .detail-meta dd { color: rgb(229 231 235); }
.detail-actions { flex-wrap: wrap; justify-content: flex-end; }

@media (max-width: 1200px) { .masonry-grid { column-count: 3; } }
@media (max-width: 900px) { .toolbar { grid-template-columns: 1fr; } .masonry-grid { column-count: 2; } .detail-body { grid-template-columns: 1fr; } }
@media (max-width: 640px) {
  .image-management-page { min-height: auto; grid-template-rows: auto auto minmax(28rem, 1fr); }
  .page-header, .panel-toolbar { align-items: stretch; flex-direction: column; }
  .header-actions, .pager { justify-content: flex-end; }
  .masonry-grid { column-count: 1; }
  .detail-modal { max-height: calc(100vh - 1rem); }
  .detail-image { min-height: 18rem; }
  .detail-meta { grid-template-columns: 1fr; }
}
</style>
