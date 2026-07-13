import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import UserRequestLogTable from '../UserRequestLogTable.vue'
import type { UserRequestLog } from '@/types'

const messages: Record<string, string> = {
  'usage.logs.columns.time': 'Time / Type',
  'usage.logs.columns.key': 'API Key / Group',
  'usage.logs.columns.model': 'Model / Reasoning',
  'usage.logs.columns.latency': 'Latency',
  'usage.logs.columns.details': 'Details',
  'usage.logs.kinds.consumption': 'Consumption',
  'usage.logs.kinds.error': 'Error',
  'usage.logs.noGroup': 'No group',
  'usage.logs.openDetail': 'View details',
  'usage.logs.errorFallback': 'Request failed',
  'usage.logs.empty': 'No logs',
  'usage.errors.keyDeleted': 'Deleted',
  'usage.latencyFirstToken': 'First',
  'usage.latencyDuration': 'Total',
  'usage.tokens': 'Tokens',
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => messages[key] ?? key }),
  }
})

const consumption: UserRequestLog = {
  id: 1,
  kind: 'consumption',
  created_at: '2026-07-13T02:13:32Z',
  request_id: 'request-consumption',
  api_key_id: 1,
  api_key_name: 'Baka1',
  api_key_deleted: false,
  group_id: 1,
  group_name: 'Pro pool',
  rate_multiplier: 1,
  model: 'gpt-5.6-sol',
  reasoning_effort: 'max',
  first_token_ms: 3000,
  duration_ms: 4000,
  total_tokens: 17_248,
  actual_cost: 0.011508,
  status_code: null,
  error_code: null,
  error_message: null,
}

const error: UserRequestLog = {
  ...consumption,
  id: 2,
  kind: 'error',
  request_id: 'request-error',
  total_tokens: null,
  actual_cost: null,
  status_code: 429,
  error_code: 'RATE_LIMITED',
  error_message: 'Please retry later',
}

const DataTableStub = {
  name: 'DataTable',
  props: ['columns', 'data', 'loading'],
  emits: ['sort', 'rowClick'],
  template: `
    <table data-test="data-table">
      <thead><tr><th v-for="column in columns" :key="column.key">{{ column.label }}</th></tr></thead>
      <tbody>
        <tr v-for="row in data" :key="row.request_id">
          <td v-for="column in columns" :key="column.key">
            <slot :name="'cell-' + column.key" :row="row" :value="row[column.key]" />
          </td>
        </tr>
      </tbody>
    </table>
  `,
}

const UserErrorDetailModalStub = {
  name: 'UserErrorDetailModal',
  props: ['show', 'errorId'],
  emits: ['update:show'],
  template: '<div data-test="error-detail" :data-show="String(show)" :data-id="errorId" />',
}

const UserUsageDetailModalStub = {
  name: 'UserUsageDetailModal',
  props: ['show', 'log'],
  emits: ['update:show'],
  template: '<div data-test="usage-detail" :data-show="String(show)" :data-id="log?.id" :data-key="log?.api_key_name" :data-group="log?.group_name" />',
}

const mountTable = () => mount(UserRequestLogTable, {
  props: { rows: [consumption, error], loading: false },
  global: {
    stubs: {
      DataTable: DataTableStub,
      Icon: true,
      UserErrorDetailModal: UserErrorDetailModalStub,
      UserUsageDetailModal: UserUsageDetailModalStub,
    },
  },
})

describe('UserRequestLogTable', () => {
  it('renders exactly the five requested compound columns', () => {
    const wrapper = mountTable()

    expect(wrapper.findAll('th').map((header) => header.text())).toEqual([
      'Time / Type',
      'API Key / Group',
      'Model / Reasoning',
      'Latency',
      'Details',
    ])
    expect(wrapper.text()).toContain('Consumption')
    expect(wrapper.text()).toContain('Error')
    expect(wrapper.text()).toContain('Baka1')
    expect(wrapper.text()).toContain('Pro pool')
    expect(wrapper.text()).toContain('1.00x')
    expect(wrapper.text()).toContain('gpt-5.6-sol')
    expect(wrapper.text()).toContain('Max')
    expect(wrapper.text()).toContain('4.00s')
    expect(wrapper.text()).toContain('RATE_LIMITED')
  })

  it('opens the matching detail dialog for consumption and error rows', async () => {
    const wrapper = mountTable()
    const detailButtons = wrapper.findAll('[data-test="request-log-details"]')

    await detailButtons.find((button) => button.attributes('data-kind') === 'consumption')!.trigger('click')
    expect(wrapper.get('[data-test="usage-detail"]').attributes()).toMatchObject({
      'data-show': 'true',
      'data-id': '1',
      'data-key': 'Baka1',
      'data-group': 'Pro pool',
    })

    await detailButtons.find((button) => button.attributes('data-kind') === 'error')!.trigger('click')
    expect(wrapper.get('[data-test="error-detail"]').attributes()).toMatchObject({
      'data-show': 'true',
      'data-id': '2',
    })
  })

  it('forwards supported server sort fields', async () => {
    const wrapper = mountTable()
    wrapper.findComponent({ name: 'DataTable' }).vm.$emit('sort', 'duration_ms', 'asc')
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('sort')).toEqual([['duration_ms', 'asc']])
  })
})
