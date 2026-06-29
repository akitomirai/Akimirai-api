export const API_KEY_PLACEHOLDER = '<YOUR_API_KEY>'
export const MODEL_PLACEHOLDER = '<MODEL_NAME>'
export const BASE_URL_PLACEHOLDER = '<BASE_URL>'

export interface QuickStartExampleOptions {
  baseUrl?: string | null
  apiKey?: string | null
  model?: string | null
}

export interface QuickStartExamples {
  baseUrl: string
  apiKey: string
  model: string
  curl: string
  openaiSdk: string
  codex: string
}

const trimTrailingSlash = (value: string): string => value.replace(/\/+$/, '')

export function normalizeOpenAIBaseUrl(rawBaseUrl?: string | null, fallback?: string): string {
  const raw = `${rawBaseUrl ?? ''}`.trim()
  const base = trimTrailingSlash(raw || fallback || BASE_URL_PLACEHOLDER)
  if (!base || base === BASE_URL_PLACEHOLDER) return BASE_URL_PLACEHOLDER
  return /\/v1$/i.test(base) ? base : `${base}/v1`
}

export function getBrowserOriginFallback(): string {
  if (typeof window === 'undefined' || !window.location?.origin) {
    return BASE_URL_PLACEHOLDER
  }
  return window.location.origin
}

export function buildQuickStartExamples(options: QuickStartExampleOptions): QuickStartExamples {
  const baseUrl = normalizeOpenAIBaseUrl(options.baseUrl, getBrowserOriginFallback())
  const apiKey = `${options.apiKey ?? ''}`.trim() || API_KEY_PLACEHOLDER
  const model = `${options.model ?? ''}`.trim() || MODEL_PLACEHOLDER

  return {
    baseUrl,
    apiKey,
    model,
    curl: [
      `curl ${baseUrl}/chat/completions \\`,
      `  -H "Authorization: Bearer ${apiKey}" \\`,
      '  -H "Content-Type: application/json" \\',
      '  -d \'{',
      `    "model": "${model}",`,
      '    "messages": [',
      '      { "role": "user", "content": "<USER_MESSAGE>" }',
      '    ]',
      "  }'"
    ].join('\n'),
    openaiSdk: [
      "import OpenAI from 'openai'",
      '',
      'const client = new OpenAI({',
      `  apiKey: '${apiKey}',`,
      `  baseURL: '${baseUrl}'`,
      '})',
      '',
      'const response = await client.chat.completions.create({',
      `  model: '${model}',`,
      "  messages: [{ role: 'user', content: '<USER_MESSAGE>' }]",
      '})',
      '',
      'console.log(response.choices[0]?.message?.content)'
    ].join('\n'),
    codex: [
      'model_provider = "custom"',
      `model = "${model}"`,
      '',
      '[model_providers.custom]',
      'name = "Akimirai"',
      `base_url = "${baseUrl}"`,
      'wire_api = "responses"',
      'env_key = "OPENAI_API_KEY"',
      '',
      '# Set OPENAI_API_KEY to your API key before starting Codex.'
    ].join('\n')
  }
}
