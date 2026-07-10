import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'
import enAdminSettings from '../../../i18n/locales/en/admin/settings'
import zhAdminSettings from '../../../i18n/locales/zh/admin/settings'

const source = readFileSync(resolve(process.cwd(), 'src/views/admin/SettingsView.vue'), 'utf8')
const apiSource = readFileSync(resolve(process.cwd(), 'src/api/admin/settings.ts'), 'utf8')

describe('SettingsView OpenAI Fast/Flex user scope', () => {
  it('keeps user IDs in the API contract and settings form', () => {
    expect(apiSource).toContain('user_ids?: number[]')
    expect(source).toContain('rule.user_ids || []')
    expect(source).toContain('addOpenAIFastPolicyUserID')
    expect(source).toContain('removeOpenAIFastPolicyUserID')
  })

  it('drops incomplete IDs and deduplicates valid IDs before save', () => {
    expect(source).toContain('Number.isSafeInteger(userID) && userID > 0')
    expect(source).toContain('[...new Set(userIDs)]')
  })

  it('defines user-scope labels under the OpenAI Fast/Flex namespace', () => {
    expect(enAdminSettings.settings.openaiFastPolicy.userIds).toBe('Specific user IDs')
    expect(zhAdminSettings.settings.openaiFastPolicy.userIds).toBe('指定用户 ID')
  })
})
