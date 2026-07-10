import type { GroupPlatform } from '@/types'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.6-sol'

export type CcSwitchApp = 'claude' | 'codex' | 'gemini'

export interface CcSwitchImportFormData {
  app: CcSwitchApp
  name: string
  model: string
  remoteCompaction?: boolean
  haikuModel?: string
  sonnetModel?: string
  opusModel?: string
}

export interface CcSwitchImportConfig {
  app: CcSwitchApp
  endpoint: string
}

export interface CcSwitchImportDeeplinkInput {
  baseUrl: string
  platform?: GroupPlatform | null
  app: CcSwitchApp
  providerName: string
  apiKey: string
  usageScript: string
  model: string
  remoteCompaction?: boolean
  haikuModel?: string
  sonnetModel?: string
  opusModel?: string
}

export function resolveCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  app: CcSwitchApp,
  baseUrl: string
): CcSwitchImportConfig {
  return {
    app,
    endpoint: platform === 'antigravity' ? `${baseUrl}/antigravity` : baseUrl
  }
}

export function buildCcSwitchImportDeeplink(input: CcSwitchImportDeeplinkInput): string {
  const config = resolveCcSwitchImportConfig(input.platform, input.app, input.baseUrl)
  const providerName = input.app === 'codex' && input.remoteCompaction
    ? 'OpenAI'
    : input.providerName
  const entries: [string, string][] = [
    ['resource', 'provider'],
    ['app', config.app],
    ['enabled', 'true'],
    ['model', input.model.trim()],
    ['name', providerName],
    ['homepage', input.baseUrl],
    ['endpoint', config.endpoint],
    ['apiKey', input.apiKey],
    ['configFormat', 'json'],
    ['usageEnabled', 'true'],
    ['usageScript', btoa(input.usageScript)],
    ['usageAutoInterval', '30']
  ]

  if (input.app === 'claude') {
    const optionalModels = [
      ['haikuModel', input.haikuModel],
      ['sonnetModel', input.sonnetModel],
      ['opusModel', input.opusModel]
    ] as const
    for (const [key, value] of optionalModels) {
      if (value?.trim()) entries.push([key, value.trim()])
    }
  }

  return `ccswitch://v1/import?${new URLSearchParams(entries).toString()}`
}
