import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import QuickStartExamples from '../QuickStartExamples.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

const mountComponent = (props = {}) => mount(QuickStartExamples, {
  props: {
    baseUrl: 'https://example.com/v1',
    model: 'gpt-test',
    ...props
  },
  global: {
    stubs: {
      RouterLink: {
        props: ['to'],
        template: '<a><slot /></a>'
      },
      Icon: {
        template: '<span />'
      }
    }
  }
})

describe('QuickStartExamples', () => {
  it('uses placeholder key on regular pages and keeps Base URL at a single /v1', () => {
    const wrapper = mountComponent()
    const text = wrapper.text()
    const code = wrapper.findAll('pre code').map((node) => node.text()).join('\n\n')

    expect(text).toContain('<YOUR_API_KEY>')
    expect(code).toContain('https://example.com/v1/chat/completions')
    expect(code).toContain('gpt-test')
    expect(code).toContain('model_provider = "custom"')
    expect(code).toContain('wire_api = "responses"')
    expect(code).not.toContain('/v1/v1')
    expect(text).toContain('MODEL_DISABLED')
    expect(text).toContain('NO_AVAILABLE_CHANNEL')
    expect(text).toContain('UPSTREAM_RATE_LIMITED')
    expect(text).toContain('UPSTREAM_5XX')
    expect(text).toContain('REQUEST_FORMAT_INVALID')
  })

  it('can temporarily render the just-created one-time key in examples', () => {
    const wrapper = mountComponent({ apiKey: 'sk-test-placeholder' })
    const code = wrapper.findAll('pre code').map((node) => node.text()).join('\n\n')

    expect(code).toContain('sk-test-placeholder')
    expect(code).not.toContain('<YOUR_API_KEY>')
  })

  it('adds /v1 when the configured Base URL omits it', () => {
    const wrapper = mountComponent({ baseUrl: 'https://example.com/' })
    const code = wrapper.findAll('pre code').map((node) => node.text()).join('\n\n')

    expect(code).toContain('https://example.com/v1/chat/completions')
    expect(code).not.toContain('https://example.com//v1')
  })
})
