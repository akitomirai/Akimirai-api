import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ModelMultiSelect from '../ModelMultiSelect.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

const IconStub = {
  props: ['name'],
  template: '<span>{{ name }}</span>'
}

describe('ModelMultiSelect', () => {
  it('searches and emits a multi-selection', async () => {
    const wrapper = mount(ModelMultiSelect, {
      props: {
        modelValue: ['gpt-5.6-sol'],
        options: ['gpt-5.6-sol', 'claude-sonnet-4', 'gpt-5.5'],
        placeholder: 'Select models'
      },
      global: { stubs: { Icon: IconStub } }
    })

    await wrapper.get('[data-test="model-multi-select-trigger"]').trigger('click')
    await wrapper.get('[data-test="model-multi-select-search"]').setValue('claude')

    const option = wrapper.findAll('[role="option"]').find((item) => item.text().includes('claude-sonnet-4'))
    expect(option).toBeDefined()
    await option!.trigger('click')

    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toEqual([
      'gpt-5.6-sol',
      'claude-sonnet-4'
    ])
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
