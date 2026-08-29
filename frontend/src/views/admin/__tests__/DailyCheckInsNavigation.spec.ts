import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

import en from '@/i18n/locales/en'
import zh from '@/i18n/locales/zh'

const currentDir = dirname(fileURLToPath(import.meta.url))
const routerSource = readFileSync(resolve(currentDir, '../../../router/index.ts'), 'utf8')
const sidebarSource = readFileSync(resolve(currentDir, '../../../components/layout/AppSidebar.vue'), 'utf8')

describe('admin daily check-in navigation contract', () => {
  it('registers an admin-only route and a sidebar entry', () => {
    expect(routerSource).toMatch(/path: '\/admin\/daily-check-ins',[\s\S]*?requiresAdmin: true/)
    expect(routerSource).toContain("import('@/views/admin/DailyCheckInsView.vue')")
    expect(sidebarSource).toContain("path: '/admin/daily-check-ins'")
    expect(sidebarSource).toContain("t('nav.dailyCheckIns')")
  })

  it('keeps the administrator locale contract symmetric', () => {
    const expectedKeys = [
      'title', 'description', 'loadFailed', 'empty', 'currentDay', 'history',
      'keyword', 'keywordPlaceholder', 'serviceDate', 'allHistory', 'user',
      'reward', 'balanceChange', 'checkedInAt', 'viewUserHistory'
    ]
    const zhMessages = (zh.admin as Record<string, Record<string, string>>).dailyCheckIns
    const enMessages = (en.admin as Record<string, Record<string, string>>).dailyCheckIns
    expect(Object.keys(zhMessages).sort()).toEqual(expectedKeys.slice().sort())
    expect(Object.keys(enMessages).sort()).toEqual(expectedKeys.slice().sort())
  })
})
