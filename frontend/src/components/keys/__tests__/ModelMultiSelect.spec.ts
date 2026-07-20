import { describe, expect, it, vi } from 'vitest'
import { DOMWrapper, mount } from '@vue/test-utils'
import ModelMultiSelect from '../ModelMultiSelect.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

const IconStub = {
  props: ['name'],
  template: '<span>{{ name }}</span>'
}

describe('ModelMultiSelect', () => {
  it('renders the dropdown in a fixed body portal without expanding the form', async () => {
    const wrapper = mount(ModelMultiSelect, {
      attachTo: document.body,
      props: {
        modelValue: [],
        options: ['gpt-5.5', 'gpt-5.6-sol'],
        placeholder: 'Select models'
      },
      global: { stubs: { Icon: IconStub } }
    })

    await wrapper.get('[data-test="model-multi-select-trigger"]').trigger('click')

    expect(wrapper.find('[data-test="model-multi-select-dropdown"]').exists()).toBe(false)
    const dropdown = document.body.querySelector<HTMLElement>('[data-test="model-multi-select-dropdown"]')
    expect(dropdown).not.toBeNull()
    expect(dropdown?.style.position).toBe('fixed')

    wrapper.unmount()
  })

  it('searches and emits a multi-selection', async () => {
    const wrapper = mount(ModelMultiSelect, {
      attachTo: document.body,
      props: {
        modelValue: ['gpt-5.6-sol'],
        options: ['gpt-5.6-sol', 'claude-sonnet-4', 'gpt-5.5'],
        placeholder: 'Select models'
      },
      global: { stubs: { Icon: IconStub } }
    })

    await wrapper.get('[data-test="model-multi-select-trigger"]').trigger('click')
    const search = document.body.querySelector<HTMLInputElement>('[data-test="model-multi-select-search"]')
    expect(search).not.toBeNull()
    await new DOMWrapper(search!).setValue('claude')

    const option = Array.from(document.body.querySelectorAll<HTMLElement>('[role="option"]'))
      .find((item) => item.textContent?.includes('claude-sonnet-4'))
    expect(option).toBeDefined()
    await new DOMWrapper(option!).trigger('click')

    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toEqual([
      'gpt-5.6-sol',
      'claude-sonnet-4'
    ])

    wrapper.unmount()
  })

  it('disables selection when the catalog failed', () => {
    const wrapper = mount(ModelMultiSelect, {
      props: {
        modelValue: [],
        options: [],
        error: 'Catalog failed'
      },
      global: { stubs: { Icon: IconStub } }
    })

    expect(wrapper.get('[data-test="model-multi-select-trigger"]').attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('Catalog failed')
  })
})
