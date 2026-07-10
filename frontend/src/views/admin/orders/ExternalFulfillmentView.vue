<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="grid gap-4 xl:grid-cols-[1.15fr_1fr]">
        <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('payment.external.title') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('payment.external.subtitle') }}</p>
            </div>
            <div class="flex items-center gap-2">
              <button type="button" class="btn btn-secondary" :disabled="skusLoading" @click="loadSKUs">
                <Icon name="refresh" size="md" :class="skusLoading ? 'animate-spin' : ''" />
              </button>
              <button type="button" class="btn btn-primary" @click="openSKUDialog()">
                <Icon name="plus" size="sm" />
                {{ t('payment.external.createSku') }}
              </button>
            </div>
          </div>

          <div class="mb-4 grid gap-3 md:grid-cols-4">
            <Select v-model="skuFilters.platform" :options="platformOptions" @change="loadSKUs" />
            <Select v-model="skuFilters.enabled" :options="enabledOptions" clearable @change="loadSKUs" />
            <input v-model.trim="skuKeyword" type="text" class="input" :placeholder="t('payment.external.skuSearch')" @input="debounceLoadSKUs" />
            <button type="button" class="btn btn-secondary" @click="resetSKUFilters">{{ t('common.reset') }}</button>
          </div>

          <DataTable :columns="skuColumns" :data="skus" :loading="skusLoading">
            <template #cell-platform="{ value }">
              <span class="badge badge-primary">{{ value }}</span>
            </template>
            <template #cell-enabled="{ value }">
              <span class="badge" :class="value ? 'badge-success' : 'badge-danger'">
                {{ value ? t('common.enabled') : t('common.disabled') }}
              </span>
            </template>
            <template #cell-manual_url="{ value }">
              <a v-if="value" :href="String(value)" target="_blank" rel="noreferrer" class="text-sm text-primary-600 hover:underline dark:text-primary-400">
                {{ value }}
              </a>
              <span v-else class="text-sm text-gray-400">-</span>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex items-center gap-2">
                <button type="button" class="btn btn-ghost btn-sm" @click="openSKUDialog(row)">
                  <Icon name="edit" size="sm" />
                  {{ t('common.edit') }}
                </button>
                <button type="button" class="btn btn-ghost btn-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20" @click="confirmDeleteSKU(row)">
                  <Icon name="trash" size="sm" />
                  {{ t('common.delete') }}
                </button>
              </div>
            </template>
          </DataTable>

          <Pagination
            v-if="skuPagination.total > 0"
            :page="skuPagination.page"
            :total="skuPagination.total"
            :page-size="skuPagination.page_size"
            @update:page="handleSKUPageChange"
            @update:pageSize="handleSKUPageSizeChange"
          />
        </div>

        <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('payment.external.fulfillmentTitle') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('payment.external.fulfillmentSubtitle') }}</p>
            </div>
            <div class="flex items-center gap-2">
              <button type="button" class="btn btn-secondary" :disabled="fulfillmentsLoading" @click="loadFulfillments">
                <Icon name="refresh" size="md" :class="fulfillmentsLoading ? 'animate-spin' : ''" />
              </button>
              <button type="button" class="btn btn-primary" @click="openFulfillmentDialog()">
                <Icon name="plus" size="sm" />
                {{ t('payment.external.createFulfillment') }}
              </button>
            </div>
          </div>

          <div class="mb-4 grid gap-3 md:grid-cols-5">
            <Select v-model="fulfillmentFilters.platform" :options="platformOptions" @change="loadFulfillments" />
            <Select v-model="fulfillmentFilters.status" :options="fulfillmentStatusOptions" clearable @change="loadFulfillments" />
            <Select v-model="fulfillmentFilters.notify_status" :options="notifyStatusOptions" clearable @change="loadFulfillments" />
            <input v-model.trim="fulfillmentFilters.sku_code" type="text" class="input" :placeholder="t('payment.external.skuCode')" @input="debounceLoadFulfillments" />
            <input v-model.trim="fulfillmentKeyword" type="text" class="input" :placeholder="t('payment.external.fulfillmentSearch')" @input="debounceLoadFulfillments" />
          </div>

          <DataTable :columns="fulfillmentColumns" :data="fulfillments" :loading="fulfillmentsLoading">
            <template #cell-status="{ value }">
              <span class="badge" :class="statusBadgeClass(String(value))">{{ t(`payment.external.status.${String(value)}`, String(value)) }}</span>
            </template>
            <template #cell-notify_status="{ value }">
              <span class="badge" :class="notifyBadgeClass(String(value))">{{ t(`payment.external.notify.${String(value)}`, String(value)) }}</span>
            </template>
            <template #cell-redeem_code="{ value }">
              <button
                v-if="value"
                type="button"
                class="inline-flex items-center gap-1 text-sm text-primary-600 hover:underline dark:text-primary-400"
                @click="copyText(String(value))"
              >
                <Icon name="copy" size="sm" />
                {{ value }}
              </button>
              <span v-else class="text-sm text-gray-400">-</span>
            </template>
            <template #cell-delivery_message="{ value }">
              <button
                v-if="value"
                type="button"
                class="text-left text-sm text-gray-700 hover:text-primary-600 dark:text-gray-300 dark:hover:text-primary-300"
                @click="openMessageDialog(String(value))"
              >
                <span class="line-clamp-2">{{ value }}</span>
              </button>
              <span v-else class="text-sm text-gray-400">-</span>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex items-center gap-2">
                <button type="button" class="btn btn-ghost btn-sm" @click="openFulfillmentDetail(row)">
                  <Icon name="eye" size="sm" />
                  {{ t('common.view') }}
                </button>
                <button
                  v-if="row.notify_status === 'failed'"
                  type="button"
                  class="btn btn-ghost btn-sm text-blue-600 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-900/20"
                  @click="retryNotify(row)"
                >
                  <Icon name="refresh" size="sm" />
                  {{ t('payment.external.retryNotify') }}
                </button>
              </div>
            </template>
          </DataTable>

          <Pagination
            v-if="fulfillmentPagination.total > 0"
            :page="fulfillmentPagination.page"
            :total="fulfillmentPagination.total"
            :page-size="fulfillmentPagination.page_size"
            @update:page="handleFulfillmentPageChange"
            @update:pageSize="handleFulfillmentPageSizeChange"
          />
        </div>
      </section>
    </div>

    <BaseDialog :show="showSKUDialog" :title="skuForm.id ? t('payment.external.editSku') : t('payment.external.createSku')" width="wide" @close="closeSKUDialog">
      <form class="space-y-4" @submit.prevent="saveSKU">
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('payment.external.platform') }}</label>
            <input v-model.trim="skuForm.platform" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.skuCode') }}</label>
            <input v-model.trim="skuForm.sku_code" class="input" required />
          </div>
          <div class="md:col-span-2">
            <label class="input-label">{{ t('payment.external.skuName') }}</label>
            <input v-model.trim="skuForm.name" class="input" required />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.redeemType') }}</label>
            <Select v-model="skuForm.redeem_type" :options="redeemTypeOptions" />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.redeemValue') }}</label>
            <input v-model.number="skuForm.redeem_value" type="number" step="0.01" class="input" required />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.amount') }}</label>
            <input v-model.number="skuForm.amount" type="number" step="0.01" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.currency') }}</label>
            <input v-model.trim="skuForm.currency" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.groupId') }}</label>
            <input v-model.number="skuForm.group_id" type="number" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.validityDays') }}</label>
            <input v-model.number="skuForm.validity_days" type="number" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.manualUrl') }}</label>
            <input v-model.trim="skuForm.manual_url" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.expiresInDays') }}</label>
            <input v-model.number="skuForm.expires_in_days" type="number" class="input" />
          </div>
          <div class="md:col-span-2">
            <label class="input-label">{{ t('payment.external.deliveryTemplate') }}</label>
            <textarea v-model.trim="skuForm.delivery_template" rows="4" class="input"></textarea>
          </div>
        </div>
        <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
          <input v-model="skuForm.enabled" type="checkbox" class="h-4 w-4" />
          {{ t('payment.external.enabled') }}
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeSKUDialog">{{ t('common.cancel') }}</button>
          <button type="button" class="btn btn-primary" :disabled="skuSaving" @click="saveSKU">
            {{ skuSaving ? t('common.processing') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="showFulfillmentDialog" :title="t('payment.external.createFulfillment')" width="wide" @close="closeFulfillmentDialog">
      <form class="space-y-4" @submit.prevent="saveFulfillment">
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('payment.external.platform') }}</label>
            <input v-model.trim="fulfillmentForm.platform" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.platformOrderId') }}</label>
            <input v-model.trim="fulfillmentForm.platform_order_id" class="input" required />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.buyerRef') }}</label>
            <input v-model.trim="fulfillmentForm.buyer_ref" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.skuCode') }}</label>
            <input v-model.trim="fulfillmentForm.sku_code" class="input" required />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.amount') }}</label>
            <input v-model.number="fulfillmentForm.amount" type="number" step="0.01" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('payment.external.currency') }}</label>
            <input v-model.trim="fulfillmentForm.currency" class="input" />
          </div>
          <div class="md:col-span-2">
            <label class="input-label">{{ t('payment.external.manualUrl') }}</label>
            <input v-model.trim="fulfillmentForm.manual_url" class="input" />
          </div>
          <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
            <input v-model="fulfillmentForm.notify_feishu" type="checkbox" class="h-4 w-4" />
            {{ t('payment.external.notifyFeishu') }}
          </label>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeFulfillmentDialog">{{ t('common.cancel') }}</button>
          <button type="button" class="btn btn-primary" :disabled="fulfillmentSaving" @click="saveFulfillment">
            {{ fulfillmentSaving ? t('common.processing') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="showMessageDialog" :title="t('payment.external.deliveryMessage')" width="wide" @close="showMessageDialog = false">
      <pre class="whitespace-pre-wrap break-words rounded-lg bg-gray-50 p-4 text-sm text-gray-700 dark:bg-dark-700 dark:text-gray-200">{{ messageDialogText }}</pre>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="showMessageDialog = false">{{ t('common.close') }}</button>
          <button type="button" class="btn btn-primary" @click="copyText(messageDialogText)">{{ t('common.copy') }}</button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="showDeleteConfirm"
      :title="t('payment.external.deleteSku')"
      :message="t('payment.external.deleteSkuConfirm', { name: deletingSKU?.name || '' })"
      :confirm-text="t('common.delete')"
      danger
      @confirm="deleteSKU"
      @cancel="showDeleteConfirm = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import adminPaymentAPI from '@/api/admin/payment'
import { extractI18nErrorMessage, extractApiErrorMessage } from '@/utils/apiError'
import type {
  ExternalFulfillmentSKU,
  ExternalOrderFulfillment
} from '@/types/payment'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const skus = ref<ExternalFulfillmentSKU[]>([])
const fulfillments = ref<ExternalOrderFulfillment[]>([])

const skusLoading = ref(false)
const fulfillmentsLoading = ref(false)
const skuSaving = ref(false)
const fulfillmentSaving = ref(false)

const skuPagination = reactive({ page: 1, page_size: 20, total: 0 })
const fulfillmentPagination = reactive({ page: 1, page_size: 20, total: 0 })

const skuKeyword = ref('')
const fulfillmentKeyword = ref('')
const skuFilters = reactive<{ platform: string; enabled: string | number | boolean | null }>({
  platform: 'xianyu',
  enabled: null
})
const fulfillmentFilters = reactive({
  platform: 'xianyu',
  status: '',
  notify_status: '',
  sku_code: ''
})

const platformOptions = [
  { value: '', label: t('payment.external.allPlatforms') },
  { value: 'xianyu', label: t('payment.external.platformXianyu') }
]
const enabledOptions = [
  { value: true, label: t('common.enabled') },
  { value: false, label: t('common.disabled') }
]
const redeemTypeOptions = [
  { value: 'balance', label: t('payment.external.redeemTypeBalance') },
  { value: 'concurrency', label: t('payment.external.redeemTypeConcurrency') },
  { value: 'subscription', label: t('payment.external.redeemTypeSubscription') },
  { value: 'invitation', label: t('payment.external.redeemTypeInvitation') }
]
const fulfillmentStatusOptions = [
  { value: '', label: t('payment.external.allStatuses') },
  { value: 'pending', label: t('payment.external.status.pending') },
  { value: 'fulfilled', label: t('payment.external.status.fulfilled') },
  { value: 'notify_failed', label: t('payment.external.status.notify_failed') },
  { value: 'failed', label: t('payment.external.status.failed') }
]
const notifyStatusOptions = [
  { value: '', label: t('payment.external.allNotifyStatuses') },
  { value: 'skipped', label: t('payment.external.notify.skipped') },
  { value: 'sent', label: t('payment.external.notify.sent') },
  { value: 'failed', label: t('payment.external.notify.failed') }
]

const skuColumns = computed<Column[]>(() => [
  { key: 'platform', label: t('payment.external.platform') },
  { key: 'sku_code', label: t('payment.external.skuCode') },
  { key: 'name', label: t('payment.external.skuName') },
  { key: 'redeem_type', label: t('payment.external.redeemType') },
  { key: 'redeem_value', label: t('payment.external.redeemValue') },
  { key: 'enabled', label: t('payment.external.enabled') },
  { key: 'manual_url', label: t('payment.external.manualUrl') },
  { key: 'actions', label: t('common.actions') }
])

const fulfillmentColumns = computed<Column[]>(() => [
  { key: 'platform_order_id', label: t('payment.external.platformOrderId') },
  { key: 'sku_code', label: t('payment.external.skuCode') },
  { key: 'redeem_code', label: t('payment.external.redeemCode') },
  { key: 'status', label: t('payment.external.statusLabel') },
  { key: 'notify_status', label: t('payment.external.notifyStatus') },
  { key: 'buyer_ref', label: t('payment.external.buyerRef') },
  { key: 'delivery_message', label: t('payment.external.deliveryMessage') },
  { key: 'actions', label: t('common.actions') }
])

const showSKUDialog = ref(false)
const showFulfillmentDialog = ref(false)
const showMessageDialog = ref(false)
const showDeleteConfirm = ref(false)
const messageDialogText = ref('')
const deletingSKU = ref<ExternalFulfillmentSKU | null>(null)

const skuForm = reactive({
  id: 0,
  platform: 'xianyu',
  sku_code: '',
  name: '',
  amount: 0,
  currency: 'CNY',
  redeem_type: 'balance',
  redeem_value: 0,
  group_id: null as number | null,
  validity_days: 0,
  expires_in_days: null as number | null,
  manual_url: '',
  delivery_template: '',
  enabled: true
})

const fulfillmentForm = reactive({
  platform: 'xianyu',
  platform_order_id: '',
  buyer_ref: '',
  sku_code: '',
  amount: 0,
  currency: 'CNY',
  manual_url: '',
  notify_feishu: true
})

const emptySKUForm = () => ({
  id: 0,
  platform: 'xianyu',
  sku_code: '',
  name: '',
  amount: 0,
  currency: 'CNY',
  redeem_type: 'balance',
  redeem_value: 0,
  group_id: null as number | null,
  validity_days: 0,
  expires_in_days: null as number | null,
  manual_url: '',
  delivery_template: '',
  enabled: true
})

const emptyFulfillmentForm = () => ({
  platform: 'xianyu',
  platform_order_id: '',
  buyer_ref: '',
  sku_code: '',
  amount: 0,
  currency: 'CNY',
  manual_url: '',
  notify_feishu: true
})

function resetSKUForm() {
  Object.assign(skuForm, emptySKUForm())
}

function resetFulfillmentForm() {
  Object.assign(fulfillmentForm, emptyFulfillmentForm())
}

async function loadSKUs() {
  skusLoading.value = true
  try {
    const res = await adminPaymentAPI.listExternalFulfillmentSKUs({
      page: skuPagination.page,
      page_size: skuPagination.page_size,
      platform: skuFilters.platform || undefined,
      enabled: skuFilters.enabled === null ? undefined : Boolean(skuFilters.enabled),
      keyword: skuKeyword.value || undefined
    })
    skus.value = res.data.items || []
    skuPagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    skusLoading.value = false
  }
}

async function loadFulfillments() {
  fulfillmentsLoading.value = true
  try {
    const res = await adminPaymentAPI.listExternalFulfillments({
      page: fulfillmentPagination.page,
      page_size: fulfillmentPagination.page_size,
      platform: fulfillmentFilters.platform || undefined,
      status: fulfillmentFilters.status || undefined,
      notify_status: fulfillmentFilters.notify_status || undefined,
      sku_code: fulfillmentFilters.sku_code || undefined,
      keyword: fulfillmentKeyword.value || undefined
    })
    fulfillments.value = res.data.items || []
    fulfillmentPagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    fulfillmentsLoading.value = false
  }
}

let skuTimer: ReturnType<typeof setTimeout> | null = null
let fulfillmentTimer: ReturnType<typeof setTimeout> | null = null

function debounceLoadSKUs() {
  if (skuTimer) clearTimeout(skuTimer)
  skuTimer = setTimeout(() => loadSKUs(), 300)
}

function debounceLoadFulfillments() {
  if (fulfillmentTimer) clearTimeout(fulfillmentTimer)
  fulfillmentTimer = setTimeout(() => loadFulfillments(), 300)
}

function resetSKUFilters() {
  skuFilters.platform = 'xianyu'
  skuFilters.enabled = null
  skuKeyword.value = ''
  skuPagination.page = 1
  loadSKUs()
}

function handleSKUPageChange(page: number) {
  skuPagination.page = page
  loadSKUs()
}

function handleSKUPageSizeChange(size: number) {
  skuPagination.page_size = size
  skuPagination.page = 1
  loadSKUs()
}

function handleFulfillmentPageChange(page: number) {
  fulfillmentPagination.page = page
  loadFulfillments()
}

function handleFulfillmentPageSizeChange(size: number) {
  fulfillmentPagination.page_size = size
  fulfillmentPagination.page = 1
  loadFulfillments()
}

function openSKUDialog(item?: ExternalFulfillmentSKU | null) {
  if (item) {
    Object.assign(skuForm, {
      id: item.id,
      platform: item.platform || 'xianyu',
      sku_code: item.sku_code || '',
      name: item.name || '',
      amount: item.amount || 0,
      currency: item.currency || 'CNY',
      redeem_type: item.redeem_type || 'balance',
      redeem_value: item.redeem_value || 0,
      group_id: item.group_id ?? null,
      validity_days: item.validity_days || 0,
      expires_in_days: item.expires_in_days ?? null,
      manual_url: item.manual_url || '',
      delivery_template: item.delivery_template || '',
      enabled: !!item.enabled
    })
  } else {
    resetSKUForm()
  }
  showSKUDialog.value = true
}

function closeSKUDialog() {
  showSKUDialog.value = false
}

async function saveSKU() {
  skuSaving.value = true
  try {
    await adminPaymentAPI.upsertExternalFulfillmentSKU({
      platform: skuForm.platform || 'xianyu',
      sku_code: skuForm.sku_code,
      name: skuForm.name,
      amount: skuForm.amount,
      currency: skuForm.currency,
      redeem_type: skuForm.redeem_type,
      redeem_value: skuForm.redeem_value,
      group_id: skuForm.group_id ?? undefined,
      validity_days: skuForm.validity_days,
      expires_in_days: skuForm.expires_in_days ?? undefined,
      manual_url: skuForm.manual_url || undefined,
      delivery_template: skuForm.delivery_template || undefined,
      enabled: skuForm.enabled
    })
    appStore.showSuccess(t('payment.external.skuSaved'))
    showSKUDialog.value = false
    await loadSKUs()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    skuSaving.value = false
  }
}

function confirmDeleteSKU(item: ExternalFulfillmentSKU) {
  deletingSKU.value = item
  showDeleteConfirm.value = true
}

async function deleteSKU() {
  if (!deletingSKU.value) return
  try {
    await adminPaymentAPI.deleteExternalFulfillmentSKU(deletingSKU.value.id)
    appStore.showSuccess(t('payment.external.skuDeleted'))
    showDeleteConfirm.value = false
    deletingSKU.value = null
    await loadSKUs()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

function openFulfillmentDialog(item?: ExternalOrderFulfillment | null) {
  if (item) {
    Object.assign(fulfillmentForm, {
      platform: item.platform || 'xianyu',
      platform_order_id: item.platform_order_id || '',
      buyer_ref: item.buyer_ref || '',
      sku_code: item.sku_code || '',
      amount: item.amount || 0,
      currency: item.currency || 'CNY',
      manual_url: item.manual_url || '',
      notify_feishu: true
    })
  } else {
    resetFulfillmentForm()
  }
  showFulfillmentDialog.value = true
}

function closeFulfillmentDialog() {
  showFulfillmentDialog.value = false
}

async function saveFulfillment() {
  fulfillmentSaving.value = true
  try {
    const res = await adminPaymentAPI.createExternalFulfillment({
      platform: fulfillmentForm.platform || 'xianyu',
      platform_order_id: fulfillmentForm.platform_order_id,
      buyer_ref: fulfillmentForm.buyer_ref || undefined,
      sku_code: fulfillmentForm.sku_code,
      amount: fulfillmentForm.amount || undefined,
      currency: fulfillmentForm.currency || undefined,
      manual_url: fulfillmentForm.manual_url || undefined,
      notify_feishu: fulfillmentForm.notify_feishu
    })
    appStore.showSuccess(res.data.replay ? t('payment.external.fulfillmentReplayed') : t('payment.external.fulfillmentCreated'))
    showFulfillmentDialog.value = false
    await loadFulfillments()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    fulfillmentSaving.value = false
  }
}

function openFulfillmentDetail(item: ExternalOrderFulfillment) {
  messageDialogText.value = [
    `${t('payment.external.platformOrderId')}: ${item.platform_order_id}`,
    `${t('payment.external.skuCode')}: ${item.sku_code}`,
    `${t('payment.external.redeemCode')}: ${item.redeem_code || '-'}`,
    `${t('payment.external.statusLabel')}: ${item.status}`,
    `${t('payment.external.notifyStatus')}: ${item.notify_status}`,
    `${t('payment.external.buyerRef')}: ${item.buyer_ref || '-'}`,
    `${t('payment.external.operator')}: ${item.operator || '-'}`,
    `${t('payment.external.failReason')}: ${item.fail_reason || '-'}`,
    '',
    item.delivery_message || ''
  ].join('\n')
  showMessageDialog.value = true
}

function openMessageDialog(text: string) {
  messageDialogText.value = text
  showMessageDialog.value = true
}

async function retryNotify(item: ExternalOrderFulfillment) {
  try {
    await adminPaymentAPI.retryExternalFulfillmentNotify(item.id)
    appStore.showSuccess(t('payment.external.notifyRetried'))
    await loadFulfillments()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    appStore.showSuccess(t('common.copied'))
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.copyFailed')))
  }
}

function statusBadgeClass(status: string): string {
  if (status === 'fulfilled') return 'badge-success'
  if (status === 'notify_failed' || status === 'failed') return 'badge-danger'
  return 'badge-warning'
}

function notifyBadgeClass(status: string): string {
  if (status === 'sent') return 'badge-success'
  if (status === 'failed') return 'badge-danger'
  return 'badge-gray'
}

onMounted(() => {
  loadSKUs()
  loadFulfillments()
})
</script>
