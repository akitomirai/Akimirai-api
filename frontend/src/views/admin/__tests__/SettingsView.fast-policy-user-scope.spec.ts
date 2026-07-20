import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'
import enAdminSettings from '../../../i18n/locales/en/admin/settings'
import zhAdminSettings from '../../../i18n/locales/zh/admin/settings'

const source = readFileSync(resolve(process.cwd(), 'src/views/admin/SettingsView.vue'), 'utf8')
const selectorSource = readFileSync(
  resolve(process.cwd(), 'src/views/admin/settings/OpenAIFastPolicyUserSelector.vue'),
  'utf8',
)
const apiSource = readFileSync(resolve(process.cwd(), 'src/api/admin/settings.ts'), 'utf8')

describe('SettingsView OpenAI Fast/Flex user scope', () => {
  it('keeps user IDs in the API contract while delegating selection to the email selector', () => {
    expect(apiSource).toContain('user_ids?: number[]')
    expect(source).toContain('rule.user_ids || []')
    expect(source).toContain('<OpenAIFastPolicyUserSelector')
    expect(source).toContain('@update:model-value="rule.user_ids = $event"')
    expect(source).not.toContain('addOpenAIFastPolicyUserID')
    expect(source).not.toContain('removeOpenAIFastPolicyUserID')
    expect(selectorSource).toContain('adminAPI.usage.searchUsers(query)')
    expect(selectorSource).toContain('adminAPI.users.getById(id, true)')
  })

  it('drops incomplete IDs and deduplicates valid IDs before save', () => {
    expect(source).toContain('Number.isSafeInteger(userID) && userID > 0')
    expect(source).toContain('[...new Set(userIDs)]')
  })

  it('defines user-scope labels under the OpenAI Fast/Flex namespace', () => {
    expect(enAdminSettings.settings.openaiFastPolicy.userIds).toBe('Specific users')
    expect(zhAdminSettings.settings.openaiFastPolicy.userIds).toBe('指定用户')
    expect(enAdminSettings.settings.openaiFastPolicy.userSearchPlaceholder).toBe(
      'Search by user email',
    )
    expect(zhAdminSettings.settings.openaiFastPolicy.userSearchPlaceholder).toBe(
      '输入用户邮箱搜索',
    )
  })
})
