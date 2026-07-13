import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import CcSwitchImportDialog from '../CcSwitchImportDialog.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: '<section v-if="show"><h2>{{ title }}</h2><slot /><slot name="footer" /></section>'
}

const mountDialog = () => mount(CcSwitchImportDialog, {
  props: {
    show: true,
    platform: 'openai',
    providerName: 'Sub2API',
    modelOptions: ['gpt-5.6-sol', 'claude-sonnet-4']
  },
  global: { stubs: { BaseDialog: BaseDialogStub } }
})

describe('CcSwitchImportDialog', () => {
  it('defaults an OpenAI group to Codex and confirms only after submit', async () => {
    const wrapper = mountDialog()

    expect((wrapper.get('#cc-switch-name').element as HTMLInputElement).value).toBe('Sub2API')
    expect((wrapper.get('#cc-switch-main-model').element as HTMLInputElement).value).toBe('gpt-5.6-sol')
    expect((wrapper.get('[data-test="cc-switch-remote-compaction"]').element as HTMLInputElement).checked).toBe(true)
    expect(wrapper.text()).toContain('keys.ccsImport.remoteCompactionHint')
    expect(wrapper.text()).not.toContain('keys.ccsImport.remoteCompactionNameHint')
    expect(wrapper.emitted('confirm')).toBeUndefined()

    await wrapper.get('#cc-switch-import-form').trigger('submit')

    expect(wrapper.emitted('confirm')?.[0]?.[0]).toEqual({
      app: 'codex',
      name: 'Sub2API',
      model: 'gpt-5.6-sol',
      remoteCompaction: true,
      haikuModel: undefined,
      sonnetModel: undefined,
      opusModel: undefined
    })
  })

  it('allows Codex remote compaction to be disabled before import', async () => {
    const wrapper = mountDialog()

    await wrapper.get('[data-test="cc-switch-remote-compaction"]').setValue(false)
    await wrapper.get('#cc-switch-import-form').trigger('submit')

    expect(wrapper.emitted('confirm')?.[0]?.[0]).toMatchObject({
      app: 'codex',
      remoteCompaction: false
    })
  })

  it('shows and emits Claude family model fields for Claude', async () => {
    const wrapper = mountDialog()
    const claudeButton = wrapper.findAll('button').find((button) => button.text() === 'Claude')
    await claudeButton!.trigger('click')
    await wrapper.get('#cc-switch-haikuModel').setValue('claude-haiku')
    await wrapper.get('#cc-switch-main-model').setValue('claude-sonnet-4')
    await wrapper.get('#cc-switch-import-form').trigger('submit')

    expect(wrapper.emitted('confirm')?.[0]?.[0]).toMatchObject({
      app: 'claude',
      model: 'claude-sonnet-4',
      remoteCompaction: false,
      haikuModel: 'claude-haiku'
    })
  })

  it('cancel closes without confirming', async () => {
    const wrapper = mountDialog()
    const cancelButton = wrapper.findAll('button').find((button) => button.text() === 'common.cancel')
    await cancelButton!.trigger('click')

    expect(wrapper.emitted('close')).toHaveLength(1)
    expect(wrapper.emitted('confirm')).toBeUndefined()
  })
})
