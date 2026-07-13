import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

import UserDashboardFilters from '../UserDashboardFilters.vue'

const mountFilters = () => mount(UserDashboardFilters, {
  props: {
    startDate: '2026-07-01',
    endDate: '2026-07-12',
    granularity: 'day',
    loading: false,
  },
  global: {
    stubs: {
      DateRangePicker: {
        props: ['presetValues'],
        emits: ['update:startDate', 'update:endDate', 'change'],
        template: `
          <button
            data-testid="date-range"
            :data-presets="presetValues.join(',')"
            @click="$emit('update:startDate', '2026-07-02'); $emit('update:endDate', '2026-07-10'); $emit('change', { startDate: '2026-07-02', endDate: '2026-07-10', preset: '7days' })"
          />
        `,
      },
      Select: {
        props: ['modelValue', 'options'],
        emits: ['update:modelValue'],
        template: '<button data-testid="granularity" :data-options="options.map(option => option.value).join(\',\')" @click="$emit(\'update:modelValue\', \'hour\')" />',
      },
    },
  },
})

describe('UserDashboardFilters', () => {
  it('forwards date range and granularity changes through its typed contract', async () => {
    const wrapper = mountFilters()

    await wrapper.get('[data-testid="date-range"]').trigger('click')
    expect(wrapper.emitted('update:startDate')).toEqual([['2026-07-02']])
    expect(wrapper.emitted('update:endDate')).toEqual([['2026-07-10']])
    expect(wrapper.emitted('dateRangeChange')).toEqual([[
      { startDate: '2026-07-02', endDate: '2026-07-10', preset: '7days' },
    ]])
    expect(wrapper.get('[data-testid="date-range"]').attributes('data-presets')).toBe(
      'yesterday,today,last24Hours,last48Hours,7days,14days,30days',
    )
    expect(wrapper.get('[data-testid="granularity"]').attributes('data-options')).toBe(
      'hour,2h,4h,8h,day',
    )

    await wrapper.get('[data-testid="granularity"]').trigger('click')
    expect(wrapper.emitted('update:granularity')).toEqual([['hour']])
    expect(wrapper.emitted('granularityChange')).toHaveLength(1)
  })

  it('emits refresh from the existing command button', async () => {
    const wrapper = mountFilters()
    const refreshButton = wrapper.findAll('button').find(button => button.text() === 'common.refresh')

    expect(refreshButton).toBeDefined()
    await refreshButton!.trigger('click')
    expect(wrapper.emitted('refresh')).toHaveLength(1)
  })
})
