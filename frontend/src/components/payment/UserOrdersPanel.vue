<template>
  <div class="space-y-4">
    <div class="card p-4">
      <div class="flex flex-wrap items-center gap-3">
        <Select v-model="currentFilter" :options="statusFilters" class="w-full sm:w-40" @change="applyFilter" />
        <div class="flex flex-1 items-center justify-end gap-2">
          <button
            type="button"
            class="btn btn-secondary"
            :disabled="loading"
            :title="t('common.refresh')"
            @click="fetchOrders"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button v-if="showStoreLink" type="button" class="btn btn-primary" @click="router.push('/purchase')">
            {{ t('payment.result.backToRecharge') }}
          </button>
        </div>
      </div>
    </div>

    <OrderTable :orders="orders" :loading="loading">
      <template #actions="{ row }">
        <div class="flex items-center gap-2">
          <button
            v-if="row.status === 'PENDING'"
            type="button"
            data-testid="cancel-order"
            class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-yellow-600 hover:bg-yellow-50 dark:text-yellow-400 dark:hover:bg-yellow-900/20"
            @click="handleCancel(row.id)"
          >
            <Icon name="x" size="sm" />
            <span>{{ t('payment.orders.cancel') }}</span>
          </button>
          <button
            v-if="canRequestRefund(row)"
            type="button"
            data-testid="refund-order"
            class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-purple-600 hover:bg-purple-50 dark:text-purple-400 dark:hover:bg-purple-900/20"
            @click="openRefundDialog(row)"
          >
            <Icon name="dollar" size="sm" />
            <span>{{ t('payment.orders.requestRefund') }}</span>
          </button>
        </div>
      </template>
    </OrderTable>

    <Pagination
      v-if="pagination.total > 0"
      :page="pagination.page"
      :total="pagination.total"
      :page-size="pagination.page_size"
      @update:page="handlePageChange"
      @update:pageSize="handlePageSizeChange"
    />
  </div>

  <BaseDialog :show="cancelTargetId !== null" :title="t('payment.orders.cancel')" width="narrow" @close="cancelTargetId = null">
    <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('payment.confirmCancel') }}</p>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="cancelTargetId = null">{{ t('common.cancel') }}</button>
        <button type="button" data-testid="confirm-cancel" class="btn btn-danger" :disabled="actionLoading" @click="confirmCancel">
          {{ actionLoading ? t('common.processing') : t('payment.orders.cancel') }}
        </button>
      </div>
    </template>
  </BaseDialog>

  <BaseDialog :show="refundTarget !== null" :title="t('payment.orders.requestRefund')" @close="refundTarget = null">
    <div v-if="refundTarget" class="space-y-4">
      <div class="rounded-lg bg-gray-50 p-4 dark:bg-dark-800">
        <div class="flex justify-between gap-4 text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</span>
          <span class="font-mono text-gray-900 dark:text-white">#{{ refundTarget.id }}</span>
        </div>
        <div class="mt-2 flex justify-between gap-4 text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.amount') }}</span>
          <span class="text-gray-900 dark:text-white">${{ refundTarget.amount.toFixed(2) }}</span>
        </div>
      </div>
      <div>
        <label class="input-label">{{ t('payment.refundReason') }}</label>
        <textarea v-model="refundReason" rows="3" class="input mt-1 w-full" :placeholder="t('payment.refundReasonPlaceholder')" />
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="refundTarget = null">{{ t('common.cancel') }}</button>
        <button
          type="button"
          data-testid="confirm-refund"
          class="btn btn-primary"
          :disabled="actionLoading || !refundReason.trim()"
          @click="confirmRefund"
        >
          {{ actionLoading ? t('common.processing') : t('payment.orders.requestRefund') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import type { PaymentOrder } from '@/types/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderTable from '@/components/payment/OrderTable.vue'

withDefaults(defineProps<{
  showStoreLink?: boolean
}>(), {
  showStoreLink: false,
})

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(false)
const actionLoading = ref(false)
const orders = ref<PaymentOrder[]>([])
const refundEligibleProviders = ref<Set<string>>(new Set())
const currentFilter = ref('')
const cancelTargetId = ref<number | null>(null)
const refundTarget = ref<PaymentOrder | null>(null)
const refundReason = ref('')
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const statusFilters = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
])

async function fetchOrders() {
  loading.value = true
  try {
    const response = await paymentAPI.getMyOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      status: currentFilter.value || undefined,
    })
    orders.value = response.data.items || []
    pagination.total = response.data.total || 0
  } catch (error: unknown) {
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function applyFilter() {
  pagination.page = 1
  void fetchOrders()
}

function handlePageChange(page: number) {
  pagination.page = page
  void fetchOrders()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  void fetchOrders()
}

function handleCancel(orderID: number) {
  cancelTargetId.value = orderID
}

async function confirmCancel() {
  if (cancelTargetId.value === null) return
  actionLoading.value = true
  try {
    await paymentAPI.cancelOrder(cancelTargetId.value)
    appStore.showSuccess(t('common.success'))
    cancelTargetId.value = null
    await fetchOrders()
  } catch (error: unknown) {
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openRefundDialog(order: PaymentOrder) {
  refundTarget.value = order
  refundReason.value = ''
}

async function confirmRefund() {
  if (!refundTarget.value || !refundReason.value.trim()) return
  actionLoading.value = true
  try {
    await paymentAPI.requestRefund(refundTarget.value.id, { reason: refundReason.value.trim() })
    appStore.showSuccess(t('common.success'))
    refundTarget.value = null
    refundReason.value = ''
    await fetchOrders()
  } catch (error: unknown) {
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function canRequestRefund(order: PaymentOrder): boolean {
  return order.status === 'COMPLETED'
    && Boolean(order.provider_instance_id)
    && refundEligibleProviders.value.has(order.provider_instance_id || '')
}

async function loadRefundEligibility() {
  try {
    const response = await paymentAPI.getRefundEligibleProviders()
    refundEligibleProviders.value = new Set(response.data.provider_instance_ids || [])
  } catch {
    refundEligibleProviders.value = new Set()
  }
}

onMounted(() => {
  void fetchOrders()
  void loadRefundEligibility()
})

defineExpose({ fetchOrders })
</script>
