import { describe, expect, it } from 'vitest'
import {
  GROK_CC_SWITCH_MODEL,
  OPENAI_CC_SWITCH_CODEX_MODEL,
  buildCcSwitchImportDeeplink
} from '@/utils/ccswitchImport'
import type { GroupPlatform } from '@/types'

function paramsFromDeeplink(deeplink: string): URLSearchParams {
  const query = deeplink.split('?')[1] || ''
  return new URLSearchParams(query)
}

function decodeConfig(params: URLSearchParams): { config: string } {
  const bytes = Uint8Array.from(atob(params.get('config') || ''), (char) => char.charCodeAt(0))
  return JSON.parse(new TextDecoder().decode(bytes))
}

describe('ccswitchImport utils', () => {
  it('defaults OpenAI CC Switch imports to the current Codex model', () => {
    expect(OPENAI_CC_SWITCH_CODEX_MODEL).toBe('gpt-5.6-sol')
  })

  it('defaults Grok Build imports to the current Grok model', () => {
    expect(GROK_CC_SWITCH_MODEL).toBe('grok-4.5')
  })

  const baseInput = {
    baseUrl: 'https://api.example.com',
    providerName: 'Sub2API',
    apiKey: 'sk-test',
    usageScript: 'return true',
    model: 'gpt-5.6-sol'
  }

  it('encodes an explicit Codex app and main model', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'openai',
        app: 'codex'
      })
    )

    expect(params.get('resource')).toBe('provider')
    expect(params.get('app')).toBe('codex')
    expect(params.get('enabled')).toBe('true')
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.get('model')).toBe(OPENAI_CC_SWITCH_CODEX_MODEL)
    expect(atob(params.get('usageScript') || '')).toBe(baseInput.usageScript)
  })

  it('uses the stock CC Switch OpenAI name gate for remote compaction', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        app: 'codex',
        providerName: 'Aki',
        remoteCompaction: true
      })
    )

    expect(params.get('name')).toBe('OpenAI')
    const embedded = decodeConfig(params).config
    expect(embedded).toContain('name = "OpenAI"')
    expect(embedded).not.toContain('model_context_window')
    expect(embedded).not.toContain('model_auto_compact_token_limit')
  })

  it('preserves the provider name when Codex remote compaction is disabled', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        app: 'codex',
        providerName: 'Aki',
        remoteCompaction: false
      })
    )

    expect(params.get('name')).toBe('Aki')
    expect(decodeConfig(params).config).toContain('name = "Aki"')
  })

  it.each([
    { platform: 'anthropic' as GroupPlatform, app: 'claude' as const },
    { platform: 'gemini' as GroupPlatform, app: 'gemini' as const }
  ])('preserves the explicit $app target for $platform imports', ({ platform, app }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform,
        app
      })
    )

    expect(params.get('app')).toBe(app)
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.get('model')).toBe(baseInput.model)
  })

  it.each([
    'https://api.example.com',
    'https://api.example.com/',
    'https://api.example.com/v1',
    'https://api.example.com/v1/'
  ])('imports Grok Build with one /v1 suffix for base URL %s', (baseUrl) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        baseUrl,
        platform: 'grok',
        app: 'grokbuild',
        model: GROK_CC_SWITCH_MODEL
      })
    )

    expect(params.get('app')).toBe('grokbuild')
    expect(params.get('endpoint')).toBe('https://api.example.com/v1')
    expect(params.get('model')).toBe(GROK_CC_SWITCH_MODEL)
  })

  it('keeps Antigravity imports on the dedicated endpoint', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'antigravity',
        app: 'gemini'
      })
    )

    expect(params.get('app')).toBe('gemini')
    expect(params.get('endpoint')).toBe(`${baseInput.baseUrl}/antigravity`)
    expect(params.get('model')).toBe(baseInput.model)
  })

  it('adds Claude family models only for Claude imports', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'anthropic',
        app: 'claude',
        haikuModel: 'claude-haiku',
        sonnetModel: 'claude-sonnet',
        opusModel: 'claude-opus'
      })
    )

    expect(params.get('haikuModel')).toBe('claude-haiku')
    expect(params.get('sonnetModel')).toBe('claude-sonnet')
    expect(params.get('opusModel')).toBe('claude-opus')
  })

  it('does not leak Claude family fields into other apps', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        app: 'codex',
        haikuModel: 'ignored'
      })
    )

    expect(params.has('haikuModel')).toBe(false)
  })

  it('ignores remote compaction for non-Codex imports', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        app: 'gemini',
        providerName: 'Gemini Relay',
        remoteCompaction: true
      })
    )

    expect(params.get('name')).toBe('Gemini Relay')
  })
})
