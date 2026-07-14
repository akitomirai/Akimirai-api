import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import TokenCountModeToggle from '../TokenCountModeToggle.vue'
import {
  resetTokenCountModeForTests,
  setTokenCountMode,
} from '@/composables/useTokenCountMode'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => ({
      'usage.countMode.label': 'Token count format',
      'usage.countMode.modern': 'w/unit',
      'usage.countMode.legacy': 'k/unit',
    })[key] ?? key,
  }),
}))

describe('TokenCountModeToggle', () => {
  afterEach(() => {
    localStorage.clear()
    resetTokenCountModeForTests()
  })

  it('renders an accessible segmented control and updates the shared mode', async () => {
    setTokenCountMode('modern')
    const wrapper = mount(TokenCountModeToggle)
    const buttons = wrapper.findAll('button')

    expect(buttons).toHaveLength(2)
    expect(buttons.map((button) => button.text())).toEqual(['w/unit', 'k/unit'])
    expect(buttons[0].attributes('aria-pressed')).toBe('true')
    expect(buttons[1].attributes('aria-pressed')).toBe('false')

    await buttons[1].trigger('click')

    expect(buttons[0].attributes('aria-pressed')).toBe('false')
    expect(buttons[1].attributes('aria-pressed')).toBe('true')
  })
})
