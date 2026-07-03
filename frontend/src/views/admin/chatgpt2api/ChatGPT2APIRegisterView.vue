<template>
  <AppLayout>
    <template #header-title>
      <div class="flex items-center gap-2">
        <Icon name="play" size="lg" class="text-primary-600 dark:text-primary-400" />
        <span>ע���</span>
      </div>
    </template>

    <div class="w-full space-y-8">
      <section class="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.34em] text-stone-500 dark:text-dark-400">REGISTER</p>
          <h1 class="mt-2 text-3xl font-bold tracking-tight text-gray-950 dark:text-white">ChatGPTע���</h1>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <button class="register-action-button" :disabled="loading || saving" type="button" @click="loadRegister">
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            ���¼���
          </button>
          <button class="register-action-button" :disabled="saving || form.enabled" type="button" @click="saveRegister">
            <Icon name="check" size="sm" />
            ����
          </button>
          <button
            class="register-action-button register-action-button-primary"
            :disabled="saving"
            type="button"
            @click="form.enabled ? stop() : start()"
          >
            <Icon :name="form.enabled ? 'x' : 'play'" size="sm" />
            {{ form.enabled ? 'ֹͣ' : '���' }}
          </button>
          <button class="register-action-button register-action-button-danger" :disabled="saving || form.enabled" type="button" @click="reset">
            <Icon name="sync" size="sm" />
            ����
          </button>
        </div>
      </section>

      <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-6">
        <div v-for="item in statCards" :key="item.key" class="metric-card" :class="item.activeClass">
          <div class="flex items-start justify-between gap-3">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ item.label }}</p>
            <Icon :name="item.icon" size="sm" :class="item.iconClass" />
          </div>
          <p class="mt-5 font-mono text-3xl font-semibold tracking-tight" :class="item.valueClass">{{ item.value }}</p>
          <p v-if="item.detail" class="mt-1 text-xs font-medium text-gray-400 dark:text-dark-500">{{ item.detail }}</p>
        </div>
      </section>

      <section class="rounded-2xl border border-gray-200 bg-white/90 px-5 py-4 shadow-sm dark:border-dark-700 dark:bg-dark-900/90">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div class="flex flex-wrap items-center gap-3">
            <span class="status-pill" :class="form.enabled ? 'status-pill-running' : 'status-pill-stopped'">
              <span class="h-2.5 w-2.5 rounded-full" :class="form.enabled ? 'bg-emerald-500' : 'bg-gray-400'"></span>
              {{ form.enabled ? '������' : '��ֹͣ' }}
            </span>
            <span class="rounded-lg bg-stone-100 px-3 py-1.5 text-xs font-medium text-stone-600 dark:bg-dark-800 dark:text-dark-300">
              ģʽ {{ modeLabel(form.mode) }}
            </span>
            <span v-if="jobId" class="rounded-lg bg-blue-50 px-3 py-1.5 font-mono text-xs font-medium text-blue-600 dark:bg-blue-900/20 dark:text-blue-300">
              {{ jobId }}
            </span>
          </div>
          <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-dark-400">
            <Icon name="clock" size="sm" />
            <span>������� {{ lastUpdatedText }}</span>
          </div>
        </div>
      </section>

      <section class="grid items-start gap-5 xl:grid-cols-[minmax(0,1.35fr)_minmax(380px,0.65fr)]">
        <div class="register-panel">
          <div class="flex flex-col gap-3 border-b border-gray-100 pb-4 dark:border-dark-800 lg:flex-row lg:items-start lg:justify-between">
            <div class="flex items-start gap-3">
              <div class="panel-icon">
                <Icon name="userPlus" size="md" />
              </div>
              <div>
                <h2 class="text-lg font-semibold tracking-tight text-gray-950 dark:text-white">ע������</h2>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">���������ģ����������� provider��</p>
              </div>
            </div>
            <button class="register-action-button register-action-button-compact" :disabled="saving || form.enabled" type="button" @click="saveRegister">
              <Icon name="check" size="sm" />
              ��������
            </button>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-3">
            <label class="block">
              <span class="field-label">ע��ģʽ</span>
              <select v-model="form.mode" class="register-input" :disabled="form.enabled">
                <option value="total">ע������</option>
                <option value="quota">�ų�ʣ����</option>
                <option value="available">�����˺�����</option>
              </select>
            </label>
            <label class="block">
              <span class="field-label">ע������</span>
              <input v-model.number="form.total" type="number" min="1" class="register-input" :disabled="form.enabled || form.mode !== 'total'" />
            </label>
            <label class="block">
              <span class="field-label">�߳���</span>
              <input v-model.number="form.threads" type="number" min="1" class="register-input" :disabled="form.enabled" />
            </label>
            <label class="block">
              <span class="field-label">Ŀ��ʣ����</span>
              <input v-model.number="form.target_quota" type="number" min="1" class="register-input" :disabled="form.enabled || form.mode !== 'quota'" />
            </label>
            <label class="block">
              <span class="field-label">Ŀ������˺�</span>
              <input v-model.number="form.target_available" type="number" min="1" class="register-input" :disabled="form.enabled || form.mode !== 'available'" />
            </label>
            <label class="block">
              <span class="field-label">�������룩</span>
              <input v-model.number="form.check_interval" type="number" min="1" class="register-input" :disabled="form.enabled || form.mode === 'total'" />
            </label>
            <label class="block md:col-span-3">
              <span class="field-label">ע�����</span>
              <input v-model.trim="form.proxy" class="register-input font-mono text-sm" placeholder="http://user:pass@host:port" :disabled="form.enabled" />
            </label>
          </div>

          <div class="mt-6 border-t border-gray-100 pt-5 dark:border-dark-800">
            <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <h3 class="text-base font-semibold text-gray-950 dark:text-white">��������</h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">Provider ������˳���ֻ������������������·� JSON��</p>
              </div>
              <span class="w-fit rounded-lg bg-stone-100 px-3 py-1.5 text-xs font-semibold text-stone-600 dark:bg-dark-800 dark:text-dark-300">
                {{ mailProviders.length }} providers
              </span>
            </div>

            <div class="mt-4 grid gap-4 md:grid-cols-3">
              <label class="block">
                <span class="field-label">����ʱ</span>
                <input v-model.number="mailDraft.request_timeout" type="number" min="1" class="register-input" :disabled="form.enabled" @change="syncMailJsonFromDraft" />
              </label>
              <label class="block">
                <span class="field-label">�ȴ���֤�볬ʱ</span>
                <input v-model.number="mailDraft.wait_timeout" type="number" min="1" class="register-input" :disabled="form.enabled" @change="syncMailJsonFromDraft" />
              </label>
              <label class="block">
                <span class="field-label">��ѯ���</span>
                <input v-model.number="mailDraft.wait_interval" type="number" min="1" class="register-input" :disabled="form.enabled" @change="syncMailJsonFromDraft" />
              </label>
            </div>

            <label class="mt-4 flex items-start gap-3 rounded-2xl border border-gray-200 bg-white px-4 py-3 text-sm text-gray-700 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200">
              <input v-model="mailDraft.api_use_register_proxy" type="checkbox" class="mt-1 rounded border-gray-300" :disabled="form.enabled" @change="syncMailJsonFromDraft" />
              <span>
                <span class="block font-medium text-gray-900 dark:text-white">��������̨ API ʹ��ע�����</span>
                <span class="mt-1 block text-xs leading-5 text-gray-500 dark:text-dark-400">�رպ�����ƽ̨ API ֱ����OpenAI/Auth0 ע��������ʹ��ע������</span>
              </span>
            </label>

            <div class="mt-4 space-y-3">
              <div v-if="mailProviders.length === 0" class="rounded-2xl border border-dashed border-gray-200 px-4 py-5 text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
                �������� provider
              </div>
              <div v-for="provider in mailProviders" :key="provider.index" class="provider-row">
                <div class="flex min-w-0 items-center gap-3">
                  <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 font-mono text-xs font-semibold text-blue-700 dark:bg-blue-900/30 dark:text-blue-200">
                    {{ provider.index + 1 }}
                  </span>
                  <div class="min-w-0">
                    <div class="flex flex-wrap items-center gap-2">
                      <span class="chip chip-muted">{{ provider.type }}</span>
                      <span class="chip" :class="provider.enabled ? 'chip-success' : 'chip-muted'">{{ provider.enabled ? '����' : 'ͣ��' }}</span>
                    </div>
                    <p class="mt-1 truncate text-xs text-gray-500 dark:text-dark-400" :title="provider.summary">{{ provider.summary }}</p>
                  </div>
                </div>
              </div>
            </div>

            <textarea
              v-model="mailJson"
              rows="12"
              class="register-input mt-4 min-h-72 font-mono text-xs leading-5"
              spellcheck="false"
              :disabled="form.enabled"
              @blur="syncMailDraftFromJson"
            ></textarea>
          </div>
        </div>

        <div class="register-panel">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="text-lg font-semibold tracking-tight text-gray-950 dark:text-white">���н��</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">��ʾע������ĵ�ǰ״̬����������</p>
            </div>
            <span class="status-badge" :class="form.enabled ? 'status-badge-running' : 'status-badge-stopped'">
              {{ form.enabled ? '������' : '��ֹͣ' }}
            </span>
          </div>

          <div class="mt-5 grid grid-cols-2 gap-2 sm:grid-cols-4 xl:grid-cols-2 2xl:grid-cols-4">
            <div v-for="item in runStats" :key="item.label" class="mini-stat">
              <p class="truncate text-xs text-gray-400 dark:text-dark-500">{{ item.label }}</p>
              <p class="mt-1 truncate font-mono text-base font-semibold text-gray-900 dark:text-white" :title="String(item.value)">{{ item.value }}</p>
            </div>
          </div>

          <div class="mt-4 grid grid-cols-3 gap-2">
            <button class="register-control-button register-control-button-primary" :disabled="saving" type="button" @click="form.enabled ? stop() : start()">
              <Icon :name="form.enabled ? 'x' : 'play'" size="sm" />
              {{ form.enabled ? 'ֹͣ' : '���' }}
            </button>
            <button class="register-control-button" :disabled="saving || form.enabled" type="button" @click="reset">
              <Icon name="sync" size="sm" />
              ����
            </button>
            <button class="register-control-button" :disabled="saving || form.enabled" type="button" @click="saveRegister">
              <Icon name="check" size="sm" />
              ����
            </button>
          </div>

          <div class="mt-4 flex items-center gap-2 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-xs leading-5 text-amber-800 dark:border-amber-900/40 dark:bg-amber-900/20 dark:text-amber-200">
            <Icon name="exclamationTriangle" size="sm" class="shrink-0" />
            ���֮ǰע���ȱ������á�
          </div>

          <div class="mt-5 flex flex-col gap-3 border-t border-gray-100 pt-5 dark:border-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h3 class="text-sm font-semibold text-gray-950 dark:text-white">ʵʱ��־</h3>
                <p class="mt-1 text-xs text-amber-700 dark:text-amber-300">���� HTTP 400 �ȴ���ͨ����Ҫ�����µ��������䡣</p>
              </div>
              <span class="rounded-lg bg-stone-100 px-2.5 py-1 text-xs font-semibold text-stone-600 dark:bg-dark-800 dark:text-dark-300">
                {{ terminalLogs.length }}
              </span>
            </div>
            <div class="terminal-box">
              <div v-if="terminalLogs.length === 0" class="text-gray-400">������־</div>
              <div v-for="log in terminalLogs" :key="log.id" class="terminal-line" :class="logLineClass(log.level)">
                <span class="text-gray-400">{{ formatLogTime(log.time) }}</span>
                <span class="pl-2">{{ log.summary }}</span>
              </div>
            </div>
          </div>

          <div class="mt-5 border-t border-gray-100 pt-5 dark:border-dark-800">
            <div class="mb-3 flex items-center justify-between gap-2">
              <h3 class="text-sm font-semibold text-gray-950 dark:text-white">Outlook ��</h3>
              <button class="register-action-button register-action-button-compact" :disabled="saving || form.enabled" type="button" @click="resetOutlook">
                <Icon name="trash" size="sm" />
                ����ȫ��״̬
              </button>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div v-for="item in outlookStats" :key="item.label" class="mini-stat">
                <p class="truncate text-xs text-gray-400 dark:text-dark-500">{{ item.label }}</p>
                <p class="mt-1 font-mono text-base font-semibold text-gray-900 dark:text-white">{{ item.value }}</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="register-panel">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-gray-950 dark:text-white">�����־</h2>
          <button class="register-action-button register-action-button-compact" type="button" @click="loadLogs">
            <Icon name="refresh" size="sm" />
          </button>
        </div>
        <div class="mt-4 overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 text-sm dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-850">
              <tr>
                <th class="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">ʱ��</th>
                <th class="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">�ȼ�</th>
                <th class="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">����</th>
                <th class="px-4 py-3 text-left font-medium text-gray-600 dark:text-gray-300">����</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
              <tr v-if="logs.length === 0">
                <td colspan="4" class="px-4 py-10 text-center text-gray-500">������־</td>
              </tr>
              <tr v-for="log in logs" :key="log.id" class="align-top hover:bg-gray-50/80 dark:hover:bg-dark-850/70">
                <td class="whitespace-nowrap px-4 py-3 text-gray-600 dark:text-gray-300">{{ log.time }}</td>
                <td class="px-4 py-3"><span class="badge" :class="levelClass(log.level)">{{ log.level }}</span></td>
                <td class="px-4 py-3 text-gray-900 dark:text-gray-100">{{ log.summary }}</td>
                <td class="max-w-lg px-4 py-3">
                  <code class="block max-h-28 overflow-auto whitespace-pre-wrap break-all rounded-lg bg-gray-100 px-2 py-1 text-xs text-gray-700 dark:bg-dark-800 dark:text-gray-200">{{ formatDetail(log.detail) }}</code>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { ChatGPT2APILog, ChatGPT2APIRegisterConfig } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'

type RegisterStatIcon = 'checkCircle' | 'xCircle' | 'clock' | 'cpu' | 'user' | 'chart'

interface RegisterStatCard {
  key: string
  label: string
  value: string | number
  detail?: string
  icon: RegisterStatIcon
  iconClass: string
  valueClass: string
  activeClass?: string
}

interface MailDraft {
  request_timeout: number
  wait_timeout: number
  wait_interval: number
  api_use_register_proxy: boolean
}

interface ProviderSummary {
  index: number
  type: string
  enabled: boolean
  summary: string
  raw: Record<string, unknown>
}

const appStore = useAppStore()
const loading = ref(false)
const saving = ref(false)
const mailJson = ref('{}')
const logs = ref<ChatGPT2APILog[]>([])
const form = reactive<ChatGPT2APIRegisterConfig>({
  enabled: false,
  mail: { providers: [] },
  proxy: '',
  total: 1,
  threads: 1,
  mode: 'total',
  target_quota: 100,
  target_available: 10,
  check_interval: 5,
  stats: {},
})
const mailDraft = reactive<MailDraft>({
  request_timeout: 120,
  wait_timeout: 300,
  wait_interval: 5,
  api_use_register_proxy: true,
})

const statCards = computed<RegisterStatCard[]>(() => [
  {
    key: 'success',
    label: '�ɹ�',
    value: statNumber('success'),
    detail: `${statNumber('success_rate')}% �ɹ���`,
    icon: 'checkCircle',
    iconClass: 'text-emerald-600',
    valueClass: 'text-emerald-600',
    activeClass: form.enabled ? 'ring-1 ring-emerald-100 dark:ring-emerald-900/30' : '',
  },
  {
    key: 'fail',
    label: 'ʧ��',
    value: statNumber('fail'),
    icon: 'xCircle',
    iconClass: 'text-rose-500',
    valueClass: 'text-rose-500',
  },
  {
    key: 'done',
    label: '���',
    value: statNumber('done'),
    icon: 'checkCircle',
    iconClass: 'text-blue-500',
    valueClass: 'text-gray-950 dark:text-white',
  },
  {
    key: 'running',
    label: '�����߳�',
    value: `${statNumber('running')}/${statNumber('threads') || form.threads || 0}`,
    icon: 'cpu',
    iconClass: 'text-orange-500',
    valueClass: 'text-gray-950 dark:text-white',
  },
  {
    key: 'available',
    label: '��ǰ����',
    value: statNumber('current_available'),
    icon: 'user',
    iconClass: 'text-stone-500',
    valueClass: 'text-gray-950 dark:text-white',
  },
  {
    key: 'quota',
    label: '��ǰ���',
    value: formatCompact(statNumber('current_quota')),
    detail: `${formatDuration(statNumber('elapsed_seconds'))} ������`,
    icon: 'chart',
    iconClass: 'text-blue-500',
    valueClass: 'text-blue-500',
  },
])

const runStats = computed(() => [
  { label: '�ɹ� / �ɹ���', value: `${statNumber('success')} / ${statNumber('success_rate')}%` },
  { label: 'ʧ��', value: statNumber('fail') },
  { label: '���', value: statNumber('done') },
  { label: '���� / �߳�', value: `${statNumber('running')} / ${statNumber('threads') || form.threads || 0}` },
  { label: '����ʱ��', value: formatDuration(statNumber('elapsed_seconds')) },
  { label: 'ƽ��ע�ᵥ��', value: formatDuration(statNumber('avg_seconds')) },
  { label: '��ǰ���', value: statNumber('current_quota') },
  { label: '�����˺�', value: statNumber('current_available') },
])

const mailProviders = computed<ProviderSummary[]>(() => {
  const mail = parseMailJson()
  const providers = Array.isArray(mail?.providers) ? mail.providers : []
  return providers.map((provider, index) => {
    const raw = asRecord(provider)
    const type = String(raw.type || 'provider')
    return {
      index,
      type,
      enabled: raw.enable !== false,
      summary: providerSummary(raw),
      raw,
    }
  })
})

const outlookStats = computed(() => {
  const provider = mailProviders.value.find((item) => item.type === 'outlook_token')?.raw
  const stats = asRecord(provider?.mailboxes_stats)
  return [
    { label: 'δʹ��', value: numberFrom(stats.unused, 0) },
    { label: 'ռ����', value: numberFrom(stats.in_use, 0) },
    { label: '����', value: numberFrom(stats.used, 0) },
    { label: 'ʧ��', value: numberFrom(stats.failed, 0) },
  ]
})

const terminalLogs = computed(() => logs.value.slice(0, 16).reverse())
const jobId = computed(() => {
  const value = form.stats?.job_id
  return typeof value === 'string' && value ? value : ''
})
const lastUpdatedText = computed(() => {
  const updated = form.stats?.updated_at || form.stats?.finished_at || form.stats?.started_at
  return typeof updated === 'string' && updated ? updated : '-'
})

function applyRegister(data: ChatGPT2APIRegisterConfig) {
  Object.assign(form, data)
  const mail = asRecord(data.mail)
  applyMailDraft(mail)
  mailJson.value = JSON.stringify(Object.keys(mail).length > 0 ? mail : { providers: [] }, null, 2)
}

function buildPayload(): ChatGPT2APIRegisterConfig | null {
  let mail: Record<string, unknown>
  try {
    mail = asRecord(JSON.parse(mailJson.value || '{}'))
  } catch {
    appStore.showError('�������ò�����Ч JSON')
    return null
  }
  mail.request_timeout = normalizePositive(mailDraft.request_timeout, 120)
  mail.wait_timeout = normalizePositive(mailDraft.wait_timeout, 300)
  mail.wait_interval = normalizePositive(mailDraft.wait_interval, 5)
  mail.api_use_register_proxy = mailDraft.api_use_register_proxy
  return { ...form, mail }
}

async function loadRegister() {
  loading.value = true
  try {
    applyRegister(await adminAPI.chatgpt2api.getRegister())
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, '����ע���ʧ��'))
  } finally {
    loading.value = false
  }
}

async function saveRegister() {
  const payload = buildPayload()
  if (!payload) return
  saving.value = true
  try {
    applyRegister(await adminAPI.chatgpt2api.updateRegister(payload))
    appStore.showSuccess('����ɹ�')
    await loadLogs()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, '����ʧ��'))
  } finally {
    saving.value = false
  }
}

async function start() {
  saving.value = true
  try {
    applyRegister(await adminAPI.chatgpt2api.startRegister())
    appStore.showSuccess('�����')
    await loadLogs()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, '���ʧ��'))
  } finally {
    saving.value = false
  }
}

async function stop() {
  saving.value = true
  try {
    applyRegister(await adminAPI.chatgpt2api.stopRegister())
    appStore.showSuccess('��ֹͣ')
    await loadLogs()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, 'ֹͣʧ��'))
  } finally {
    saving.value = false
  }
}

async function reset() {
  saving.value = true
  try {
    applyRegister(await adminAPI.chatgpt2api.resetRegister())
    appStore.showSuccess('������')
    await loadLogs()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, '����ʧ��'))
  } finally {
    saving.value = false
  }
}

async function resetOutlook() {
  saving.value = true
  try {
    applyRegister(await adminAPI.chatgpt2api.resetOutlookPool())
    appStore.showSuccess('Outlook ��������')
    await loadLogs()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, '���� Outlook ��ʧ��'))
  } finally {
    saving.value = false
  }
}

async function loadLogs() {
  try {
    logs.value = await adminAPI.chatgpt2api.listLogs({ type: 'register', limit: 20 })
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, '������־ʧ��'))
  }
}

function applyMailDraft(mail: Record<string, unknown>) {
  mailDraft.request_timeout = normalizePositive(mail.request_timeout, 120)
  mailDraft.wait_timeout = normalizePositive(mail.wait_timeout, 300)
  mailDraft.wait_interval = normalizePositive(mail.wait_interval, 5)
  mailDraft.api_use_register_proxy = mail.api_use_register_proxy !== false
}

function syncMailJsonFromDraft() {
  const mail = parseMailJson()
  if (!mail) return
  mail.request_timeout = normalizePositive(mailDraft.request_timeout, 120)
  mail.wait_timeout = normalizePositive(mailDraft.wait_timeout, 300)
  mail.wait_interval = normalizePositive(mailDraft.wait_interval, 5)
  mail.api_use_register_proxy = mailDraft.api_use_register_proxy
  form.mail = mail
  mailJson.value = JSON.stringify(mail, null, 2)
}

function syncMailDraftFromJson() {
  const mail = parseMailJson()
  if (!mail) {
    appStore.showError('�������ò�����Ч JSON')
    return
  }
  form.mail = mail
  applyMailDraft(mail)
  mailJson.value = JSON.stringify(mail, null, 2)
}

function parseMailJson(): Record<string, unknown> | null {
  try {
    return asRecord(JSON.parse(mailJson.value || '{}'))
  } catch {
    return null
  }
}

function asRecord(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return {}
  return value as Record<string, unknown>
}

function providerSummary(provider: Record<string, unknown>): string {
  const parts: string[] = []
  const domains = Array.isArray(provider.domain) ? provider.domain.map(String).filter(Boolean) : []
  const saved = numberFrom(provider.mailboxes_count, 0)
  const apiBase = typeof provider.api_base === 'string' ? provider.api_base : ''
  const defaultDomain = typeof provider.default_domain === 'string' ? provider.default_domain : ''
  if (domains.length > 0) parts.push(`Domain ${domains.slice(0, 3).join(', ')}`)
  if (defaultDomain) parts.push(`Default ${defaultDomain}`)
  if (apiBase) parts.push(apiBase)
  if (saved > 0) parts.push(`�ѱ��� ${saved} ������`)
  return parts.join(' �� ') || 'δ���ö������'
}

function statNumber(key: string): number {
  return numberFrom(form.stats?.[key], 0)
}

function numberFrom(value: unknown, fallback: number): number {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim()) {
    const parsed = Number(value)
    if (Number.isFinite(parsed)) return parsed
  }
  return fallback
}

function normalizePositive(value: unknown, fallback: number): number {
  const parsed = numberFrom(value, fallback)
  return parsed > 0 ? parsed : fallback
}

function formatDetail(detail?: Record<string, unknown>): string {
  return detail ? JSON.stringify(detail, null, 2) : '-'
}

function formatLogTime(value: string): string {
  if (!value) return '--:--:--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value.slice(11, 19) || value
  return date.toLocaleTimeString('zh-CN', { hour12: false })
}

function formatDuration(seconds: number): string {
  const total = Math.max(0, Math.round(seconds))
  if (total < 60) return `${total}s`
  const minutes = Math.floor(total / 60)
  const restSeconds = total % 60
  if (minutes < 60) return `${minutes}m ${restSeconds}s`
  const hours = Math.floor(minutes / 60)
  const restMinutes = minutes % 60
  return `${hours}h ${restMinutes}m`
}

function formatCompact(value: number): string {
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return String(value)
}

function modeLabel(mode: string): string {
  return ({ total: 'ע������', quota: '�ų�ʣ����', available: '�����˺�����' } as Record<string, string>)[mode] || mode
}

function levelClass(level: string): string {
  if (level === 'error') return 'badge-danger'
  if (level === 'warn') return 'badge-warning'
  if (level === 'debug') return 'badge-gray'
  return 'badge-success'
}

function logLineClass(level: string): string {
  if (level === 'error') return 'text-rose-600 dark:text-rose-300'
  if (level === 'warn') return 'text-amber-700 dark:text-amber-300'
  if (level === 'debug') return 'text-gray-500 dark:text-dark-400'
  return 'text-emerald-700 dark:text-emerald-300'
}

onMounted(async () => {
  await loadRegister()
  await loadLogs()
})
</script>

<style scoped>
.metric-card {
  @apply min-h-[128px] rounded-2xl border border-gray-200 bg-white/90 p-5 shadow-sm transition dark:border-dark-700 dark:bg-dark-900/90;
}

.register-action-button {
  @apply inline-flex h-12 items-center justify-center gap-2 rounded-2xl border border-gray-200 bg-white px-4 text-sm font-medium text-gray-700 shadow-sm transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-200 dark:hover:bg-dark-800;
}

.register-action-button-primary {
  @apply border-gray-950 bg-gray-950 text-white hover:bg-gray-800 dark:border-white dark:bg-white dark:text-gray-950 dark:hover:bg-gray-100;
}

.register-action-button-danger {
  @apply border-rose-500 bg-rose-500 text-white hover:bg-rose-600 dark:border-rose-500 dark:bg-rose-500 dark:text-white;
}

.register-action-button-compact {
  @apply h-10 rounded-xl px-3 text-xs;
}

.register-panel {
  @apply rounded-2xl border border-gray-200 bg-white/90 p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900/90;
}

.panel-icon {
  @apply flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-stone-100 text-stone-600 dark:bg-dark-800 dark:text-dark-300;
}

.field-label {
  @apply mb-2 block text-sm font-medium text-gray-700 dark:text-dark-200;
}

.register-input {
  @apply w-full rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-gray-900 outline-none transition placeholder:text-gray-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/10 disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-400 dark:border-dark-700 dark:bg-dark-800 dark:text-white dark:placeholder:text-dark-500 dark:disabled:bg-dark-850 dark:disabled:text-dark-500;
}

.status-pill {
  @apply inline-flex h-9 items-center gap-2 rounded-xl px-3 text-sm font-medium;
}

.status-pill-running,
.status-badge-running {
  @apply bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300;
}

.status-pill-stopped,
.status-badge-stopped {
  @apply bg-gray-100 text-gray-600 dark:bg-dark-800 dark:text-dark-300;
}

.status-badge {
  @apply inline-flex rounded-lg px-2.5 py-1 text-xs font-semibold;
}

.mini-stat {
  @apply min-w-0 rounded-xl border border-gray-200 bg-white/80 px-3 py-2.5 dark:border-dark-700 dark:bg-dark-800/70;
}

.register-control-button {
  @apply inline-flex h-11 min-w-0 items-center justify-center gap-2 rounded-xl border border-gray-200 bg-white px-3 text-sm font-medium text-gray-700 shadow-sm transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200 dark:hover:bg-dark-700;
}

.register-control-button-primary {
  @apply border-gray-950 bg-gray-950 text-white hover:bg-gray-800 dark:border-white dark:bg-white dark:text-gray-950 dark:hover:bg-gray-100;
}

.terminal-box {
  @apply h-80 overflow-y-auto rounded-2xl border border-gray-200 bg-white/80 p-4 font-mono text-xs leading-6 dark:border-dark-700 dark:bg-dark-950/60;
}

.terminal-line {
  @apply break-words;
}

.provider-row {
  @apply rounded-2xl border border-gray-200 bg-white px-4 py-3 dark:border-dark-700 dark:bg-dark-800;
}

.chip {
  @apply inline-flex h-7 items-center rounded-lg border px-2.5 text-xs font-medium;
}

.chip-muted {
  @apply border-gray-200 bg-gray-50 text-gray-600 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300;
}

.chip-success {
  @apply border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/40 dark:bg-emerald-900/20 dark:text-emerald-300;
}
</style>
