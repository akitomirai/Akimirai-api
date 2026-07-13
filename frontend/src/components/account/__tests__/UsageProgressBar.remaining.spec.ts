import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UsageProgressBar from '../UsageProgressBar.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

describe('UsageProgressBar remaining capacity', () => {
  it('shows full remaining capacity as a full green bar', () => {
    const wrapper = mount(UsageProgressBar, {
      props: { label: 'Req', utilization: 100, remainingCapacity: true, color: 'indigo' }
    })

    expect(wrapper.text()).toContain('100%')
    expect(wrapper.get('.h-1\\.5 > div').attributes('style')).toContain('width: 100%')
    expect(wrapper.get('.h-1\\.5 > div').classes()).toContain('bg-green-500')
  })

  it('shows depleted remaining capacity as an empty red bar', () => {
    const wrapper = mount(UsageProgressBar, {
      props: { label: 'Req', utilization: 0, remainingCapacity: true, color: 'indigo' }
    })

    expect(wrapper.text()).toContain('0%')
    expect(wrapper.get('.h-1\\.5 > div').attributes('style')).toContain('width: 0%')
    expect(wrapper.get('.h-1\\.5 > div').classes()).toContain('bg-red-500')
  })

  it('keeps over-limit usage red while displaying the original percentage', () => {
    const wrapper = mount(UsageProgressBar, {
      props: { label: '5h', utilization: 120, color: 'indigo' }
    })

    expect(wrapper.text()).toContain('120%')
    expect(wrapper.get('.h-1\\.5 > div').attributes('style')).toContain('width: 100%')
    expect(wrapper.get('.h-1\\.5 > div').classes()).toContain('bg-red-500')
  })
})
