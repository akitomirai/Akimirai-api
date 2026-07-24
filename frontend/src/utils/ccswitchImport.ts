import type { GroupPlatform } from '@/types'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.6-sol'
export const GROK_CC_SWITCH_MODEL = 'grok-4.5'

export type CcSwitchApp = 'claude' | 'codex' | 'gemini' | 'grokbuild'

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

function withV1Endpoint(baseUrl: string): string {
  const normalizedBaseUrl = baseUrl.replace(/\/+$/, '')
  return normalizedBaseUrl.endsWith('/v1') ? normalizedBaseUrl : `${normalizedBaseUrl}/v1`
}

export function resolveCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  app: CcSwitchApp,
  baseUrl: string
): CcSwitchImportConfig {
  if (app === 'grokbuild') {
    return {
      app,
      endpoint: withV1Endpoint(baseUrl)
    }
  }

  return {
    app,
    endpoint: platform === 'antigravity' ? `${baseUrl}/antigravity` : baseUrl
  }
}

export function buildCcSwitchImportDeeplink(input: CcSwitchImportDeeplinkInput): string {
  const config = resolveCcSwitchImportConfig(input.platform, input.app, input.baseUrl)
  const providerName = input.app === 'codex' && input.remoteCompaction
    ? 'OpenAI'
    : input.providerName.trim()
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

  if (input.app === 'codex') {
    const configToml = `model_provider = "custom"
model = ${JSON.stringify(input.model.trim())}
model_reasoning_effort = "high"
disable_response_storage = true

[model_providers.custom]
name = ${JSON.stringify(providerName)}
base_url = ${JSON.stringify(config.endpoint)}
wire_api = "responses"
requires_openai_auth = true
`
    entries.push(['config', encodeBase64Utf8(JSON.stringify({
      auth: { OPENAI_API_KEY: input.apiKey },
      config: configToml
    }))])
  }

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

function encodeBase64Utf8(value: string): string {
  const bytes = new TextEncoder().encode(value)
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return btoa(binary)
}
