<template>
  <AppLayout>
    <div class="space-y-4">
      <!-- Actions -->
      <div class="flex items-center justify-end gap-2">
        <button @click="refreshLaunchState" :disabled="launchStateLoading" class="btn btn-secondary" :title="t('common.refresh')">
          <Icon name="refresh" size="md" :class="launchStateLoading ? 'animate-spin' : ''" />
        </button>
        <button @click="openPlanEdit(null)" class="btn btn-primary">{{ t('payment.admin.createPlan') }}</button>
      </div>

      <!-- Store Launch Checklist -->
      <section class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.launchChecklistTitle') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.launchChecklistDesc') }}</p>
          </div>
          <span
            class="mt-2 inline-flex w-fit items-center gap-1 rounded-full px-2.5 py-1 text-xs font-medium sm:mt-0"
            :class="launchReady ? 'bg-green-50 text-green-700 dark:bg-green-900/25 dark:text-green-300' : 'bg-amber-50 text-amber-700 dark:bg-amber-900/25 dark:text-amber-300'"
          >
            <Icon :name="launchReady ? 'checkCircle' : 'exclamationTriangle'" size="sm" />
            {{ launchReady ? t('payment.admin.launchReady') : t('payment.admin.launchNeedsSetup') }}
          </span>
        </div>

        <div class="mt-4 grid gap-3 md:grid-cols-3">
          <div
            v-for="item in launchChecklistItems"
            :key="item.key"
            class="rounded-lg border p-4"
            :class="item.done ? 'border-green-100 bg-green-50/60 dark:border-green-900/40 dark:bg-green-900/10' : 'border-amber-100 bg-amber-50/60 dark:border-amber-900/40 dark:bg-amber-900/10'"
          >
            <div class="flex items-start gap-3">
              <span
                class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg"
                :class="item.done ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'"
              >
                <Icon :name="item.done ? 'checkCircle' : 'exclamationTriangle'" size="md" />
              </span>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ item.title }}</h3>
                  <span
                    class="rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="item.done ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'"
                  >
                    {{ item.done ? t('payment.admin.launchDone') : t('payment.admin.launchPending') }}
                  </span>
                </div>
                <p class="mt-2 text-sm leading-5 text-gray-600 dark:text-gray-300">{{ item.description }}</p>
                <button
                  type="button"
                  class="btn btn-secondary btn-sm mt-3"
                  :disabled="item.disabled"
                  @click="item.action()"
                >
                  {{ item.disabled ? t('payment.admin.launchBlocked') : item.actionLabel }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Plans Table -->
      <DataTable :columns="planColumns" :data="plans" :loading="plansLoading">
        <template #empty>
          <div v-if="!hasSubscriptionGroups" class="mx-auto flex max-w-xl flex-col items-center text-center">
            <span class="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-amber-50 text-amber-600 dark:bg-amber-900/25 dark:text-amber-300">
              <Icon name="exclamationTriangle" size="lg" />
            </span>
            <p class="text-lg font-semibold text-gray-900 dark:text-gray-100">
              {{ t('payment.admin.noSubscriptionGroupsTitle') }}
            </p>
            <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">
              {{ t('payment.admin.noSubscriptionGroupsDesc') }}
            </p>
          </div>
          <div v-else class="flex flex-col items-center">
            <Icon
              name="inbox"
              size="xl"
              class="mb-4 h-12 w-12 text-gray-400 dark:text-dark-500"
            />
            <p class="text-lg font-medium text-gray-900 dark:text-gray-100">
              {{ t('empty.noData') }}
            </p>
          </div>
        </template>
        <template #cell-name="{ value, row }">
          <span class="text-sm font-medium" :class="getPlanNameClass(row.group_id)">{{ value }}</span>
        </template>
        <template #cell-group_id="{ value }">
          <span v-if="isGroupMissing(value)" class="text-sm">
            <span class="text-gray-400">#{{ value }}</span>
            <span class="ml-1 badge badge-danger">{{ t('payment.admin.groupMissing') }}</span>
          </span>
          <GroupBadge
            v-else-if="getGroup(value)"
            :name="getGroup(value)!.name"
            :platform="getGroup(value)!.platform"
            :rate-multiplier="getGroup(value)!.rate_multiplier"
          />
          <span v-else class="text-sm text-gray-400">-</span>
        </template>
        <template #cell-price="{ value, row }">
          <div class="text-sm">
            <span class="font-medium text-gray-900 dark:text-white">${{ (value ?? 0).toFixed(2) }}</span>
            <span v-if="row.original_price" class="ml-1 text-xs text-gray-400 line-through">${{ row.original_price.toFixed(2) }}</span>
          </div>
        </template>
        <template #cell-validity_days="{ value, row }">
          <span class="text-sm">{{ value }} {{ t('payment.admin.' + (row.validity_unit || 'days')) }}</span>
        </template>
        <template #cell-for_sale="{ value, row }">
          <button
            type="button"
            :class="[
              'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
              value ? 'bg-primary-500' : 'bg-gray-300 dark:bg-dark-600'
            ]"
            @click="toggleForSale(row)"
          >
            <span :class="[
              'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
              value ? 'translate-x-4' : 'translate-x-0'
            ]" />
          </button>
        </template>
        <template #cell-actions="{ row }">
          <div class="flex items-center gap-2">
            <button @click="openPlanEdit(row)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400">
              <Icon name="edit" size="sm" />
              <span class="text-xs">{{ t('common.edit') }}</span>
            </button>
            <button @click="confirmDeletePlan(row)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400">
              <Icon name="trash" size="sm" />
              <span class="text-xs">{{ t('common.delete') }}</span>
            </button>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Plan Edit Dialog -->
    <PlanEditDialog :show="showPlanDialog" :plan="editingPlan" :groups="groups" @close="showPlanDialog = false" @saved="loadPlans" />

    <ConfirmDialog :show="showDeletePlanDialog" :title="t('payment.admin.deletePlan')" :message="t('payment.admin.deletePlanConfirm')" :confirm-text="t('common.delete')" danger @confirm="handleDeletePlan" @cancel="showDeletePlanDialog = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import adminAPI from '@/api/admin'
import type { ProviderInstance, SubscriptionPlan } from '@/types/payment'
import type { AdminGroup } from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import PlanEditDialog from './PlanEditDialog.vue'
import { platformTextClass } from '@/utils/platformColors'

const { t } = useI18n()
const appStore = useAppStore()
const router = useRouter()

// ==================== Groups ====================

const groups = ref<AdminGroup[]>([])

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch { /* ignore */ }
}

function getGroup(id: number): AdminGroup | undefined {
  return groups.value.find(g => g.id === id)
}

function isGroupMissing(id: number): boolean {
  return id > 0 && !groups.value.find(g => g.id === id)
}

function getPlanNameClass(groupId: number): string {
  const group = getGroup(groupId)
  return group ? platformTextClass(group.platform) : 'text-gray-900 dark:text-white'
}

const hasSubscriptionGroups = computed(() =>
  groups.value.some(g => g.subscription_type === 'subscription' && g.status === 'active'),
)

// ==================== Providers ====================

const providersLoading = ref(false)
const providers = ref<ProviderInstance[]>([])

async function loadProviders() {
  providersLoading.value = true
  try {
    const res = await adminPaymentAPI.getProviders()
    providers.value = res.data || []
  } catch { /* ignore */ }
  finally { providersLoading.value = false }
}

const hasEnabledProviders = computed(() => providers.value.some(provider => provider.enabled))

// ==================== Plans ====================

const plansLoading = ref(false)
const plans = ref<SubscriptionPlan[]>([])
const showPlanDialog = ref(false)
const showDeletePlanDialog = ref(false)
const editingPlan = ref<SubscriptionPlan | null>(null)
const deletingPlanId = ref<number | null>(null)

const planColumns = computed((): Column[] => [
  { key: 'id', label: 'ID' },
  { key: 'name', label: t('payment.admin.planName') },
  { key: 'group_id', label: t('payment.admin.group') },
  { key: 'price', label: t('payment.admin.price') },
  { key: 'validity_days', label: t('payment.admin.validityDays') },
  { key: 'for_sale', label: t('payment.admin.forSale') },
  { key: 'sort_order', label: t('payment.admin.sortOrder') },
  { key: 'actions', label: t('common.actions') },
])

const hasOnSalePlans = computed(() => plans.value.some(plan => plan.for_sale))
const launchReady = computed(() => hasSubscriptionGroups.value && hasOnSalePlans.value && hasEnabledProviders.value)
const launchStateLoading = computed(() => plansLoading.value || providersLoading.value)

const launchChecklistItems = computed(() => [
  {
    key: 'subscription-group',
    done: hasSubscriptionGroups.value,
    title: t('payment.admin.launchSubscriptionGroupTitle'),
    description: t('payment.admin.launchSubscriptionGroupDesc'),
    actionLabel: t('payment.admin.launchSubscriptionGroupAction'),
    disabled: false,
    action: () => router.push('/admin/groups'),
  },
  {
    key: 'subscription-plan',
    done: hasOnSalePlans.value,
    title: t('payment.admin.launchSubscriptionPlanTitle'),
    description: t('payment.admin.launchSubscriptionPlanDesc'),
    actionLabel: t('payment.admin.launchSubscriptionPlanAction'),
    disabled: !hasSubscriptionGroups.value,
    action: () => openPlanEdit(null),
  },
  {
    key: 'payment-provider',
    done: hasEnabledProviders.value,
    title: t('payment.admin.launchPaymentProviderTitle'),
    description: t('payment.admin.launchPaymentProviderDesc'),
    actionLabel: t('payment.admin.launchPaymentProviderAction'),
    disabled: false,
    action: () => router.push('/admin/settings'),
  },
])

async function loadPlans() {
  plansLoading.value = true
  try {
    const res = await adminPaymentAPI.getPlans()
    // Backend returns features as newline-separated string; parse to array
    plans.value = (res.data || []).map((p: Omit<SubscriptionPlan, 'features'> & { features: string | string[] }) => ({
      ...p,
      features: typeof p.features === 'string'
        ? p.features.split('\n').map((f: string) => f.trim()).filter(Boolean)
        : (p.features || []),
    }))
  }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { plansLoading.value = false }
}

function refreshLaunchState() {
  loadGroups()
  loadPlans()
  loadProviders()
}

function openPlanEdit(plan: SubscriptionPlan | null) {
  editingPlan.value = plan
  showPlanDialog.value = true
}


/** Quick toggle for_sale from the list */
async function toggleForSale(plan: SubscriptionPlan) {
  try {
    await adminPaymentAPI.updatePlan(plan.id, { for_sale: !plan.for_sale })
    plan.for_sale = !plan.for_sale
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

function confirmDeletePlan(plan: SubscriptionPlan) { deletingPlanId.value = plan.id; showDeletePlanDialog.value = true }
async function handleDeletePlan() {
  if (!deletingPlanId.value) return
  try { await adminPaymentAPI.deletePlan(deletingPlanId.value); appStore.showSuccess(t('common.deleted')); showDeletePlanDialog.value = false; loadPlans() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

// ==================== Lifecycle ====================

onMounted(() => {
  refreshLaunchState()
})
</script>
