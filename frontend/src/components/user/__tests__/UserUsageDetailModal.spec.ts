import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserUsageDetailModal from '../UserUsageDetailModal.vue'
import type { UserRequestLog } from '@/types'

const { getById } = vi.hoisted(() => ({ getById: vi.fn() }))

vi.mock('@/api', () => ({
  usageAPI: { getById },
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const log: UserRequestLog = {
  id: 9,
  kind: 'consumption',
  created_at: '2026-07-13T02:13:32Z',
  request_id: 'request-row-identity',
  api_key_id: 3,
  api_key_name: 'Baka1',
  api_key_deleted: false,
  group_id: 4,
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

const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: '<div v-if="show" data-test="dialog"><slot /></div>',
}

describe('UserUsageDetailModal', () => {
  it('uses the selected unified-log row for identity fields', async () => {
    getById.mockResolvedValue({
      id: 9,
      request_id: 'raw-record-without-relations',
      model: 'raw-model',
      reasoning_effort: null,
      input_tokens: 17_196,
      output_tokens: 52,
      cache_read_tokens: 0,
      cache_creation_tokens: 0,
      rate_multiplier: 1,
      actual_cost: 0.011508,
      total_cost: 0.05754,
      first_token_ms: 9999,
      duration_ms: 9999,
      created_at: '2026-01-01T00:00:00Z',
      api_key: undefined,
      group: undefined,
    })

    const wrapper = mount(UserUsageDetailModal, {
      props: { show: false, log: null },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          LoadingSpinner: true,
        },
      },
    })

    await wrapper.setProps({ show: true, log })
    await flushPromises()

    expect(getById).toHaveBeenCalledWith(9)
    expect(wrapper.text()).toContain('request-row-identity')
    expect(wrapper.text()).toContain('Baka1')
    expect(wrapper.text()).toContain('Pro pool')
    expect(wrapper.text()).toContain('gpt-5.6-sol')
    expect(wrapper.text()).toContain('Max')
    expect(wrapper.text()).not.toContain('raw-record-without-relations')
    expect(wrapper.text()).not.toContain('raw-model')
  })
})
