import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AuditLogView from '../AuditLogView.vue'

const { listAuditLogs, showError } = vi.hoisted(() => ({
  listAuditLogs: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    audit: {
      list: listAuditLogs,
      get: vi.fn(),
      clear: vi.fn()
    }
  }
}))

vi.mock('@/api', () => ({
  totpAPI: {
    getStatus: vi.fn(),
    verify: vi.fn()
  }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const DataTableStub = {
  props: ['columns', 'data'],
  template: `
    <div>
      <div data-test="action-column-class">
        {{ columns.find((column) => column.key === 'action')?.class || '' }}
      </div>
      <div v-for="row in data" :key="row.id" data-test="action-cell">
        <slot name="cell-action" :row="row" :value="row.action" />
      </div>
    </div>
  `
}

describe('admin AuditLogView', () => {
  beforeEach(() => {
    listAuditLogs.mockReset()
    showError.mockReset()
    listAuditLogs.mockResolvedValue({
      items: [
        {
          id: 1,
          created_at: '2026-07-22T05:38:01Z',
          actor_user_id: 1,
          actor_email: 'admin@example.com',
          actor_role: 'admin',
          auth_method: 'jwt',
          credential_masked: '',
          action: 'admin.settings.email_template_preview.create.with_an_unbroken_suffix',
          method: 'POST',
          path: '/api/v1/admin/settings/email-template-preview/with/an/unbroken/path/suffix',
          request_id: 'request-1',
          client_ip: '127.0.0.1',
          user_agent: 'vitest',
          status_code: 200,
          latency_ms: 1
        }
      ],
      total: 1,
      page: 1,
      page_size: 20
    })
  })

  it('wraps complete action and path values inside a bounded action column', async () => {
    const wrapper = mount(AuditLogView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          Select: true,
          BaseDialog: true,
          ConfirmDialog: true,
          Icon: true
        }
      }
    })

    await flushPromises()

    const actionColumnClass = wrapper.get('[data-test="action-column-class"]').text()
    expect(actionColumnClass).toContain('min-w-[')
    expect(actionColumnClass).toContain('max-w-[')
    expect(actionColumnClass).toContain('whitespace-normal')

    const actionCell = wrapper.get('[data-test="action-cell"]')
    expect(actionCell.text()).toContain('admin.settings.email_template_preview.create.with_an_unbroken_suffix')
    expect(actionCell.text()).toContain('/api/v1/admin/settings/email-template-preview/with/an/unbroken/path/suffix')
    expect(actionCell.findAll('.truncate')).toHaveLength(0)
    expect(actionCell.findAll('.whitespace-normal').length).toBeGreaterThan(0)
    expect(actionCell.findAll('.break-all').length).toBeGreaterThanOrEqual(2)
  })
})
