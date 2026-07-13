import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'

import DateRangePicker from '../DateRangePicker.vue'

const messages: Record<string, string> = {
  'dates.today': 'Today',
  'dates.yesterday': 'Yesterday',
  'dates.last24Hours': 'Last 24 Hours',
  'dates.last48Hours': 'Last 48 Hours',
  'dates.last7Days': 'Last 7 Days',
  'dates.last14Days': 'Last 14 Days',
  'dates.last30Days': 'Last 30 Days',
  'dates.thisMonth': 'This Month',
  'dates.lastMonth': 'Last Month',
  'dates.startDate': 'Start Date',
  'dates.endDate': 'End Date',
  'dates.apply': 'Apply',
  'dates.selectDateRange': 'Select date range'
}

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => messages[key] ?? key,
    locale: ref('en')
  })
}))

const formatLocalDate = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

describe('DateRangePicker', () => {
  it('shows the dashboard presets in the requested order', async () => {
    const today = formatLocalDate(new Date())
    const wrapper = mount(DateRangePicker, {
      props: {
        startDate: today,
        endDate: today,
        presetValues: ['yesterday', 'today', 'last24Hours', 'last48Hours', '7days', '14days', '30days']
      },
      global: { stubs: { Icon: true } }
    })

    await wrapper.find('.date-picker-trigger').trigger('click')

    expect(wrapper.findAll('.date-picker-preset').map((node) => node.text())).toEqual([
      'Yesterday',
      'Today',
      'Last 24 Hours',
      'Last 48 Hours',
      'Last 7 Days',
      'Last 14 Days',
      'Last 30 Days'
    ])
  })

  it('uses last 24 hours as the default recognized preset', () => {
    const now = new Date()
    const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000)

    const wrapper = mount(DateRangePicker, {
      props: {
        startDate: formatLocalDate(yesterday),
        endDate: formatLocalDate(now)
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('Last 24 Hours')
  })

  it('emits range updates with last24Hours preset when applied', async () => {
    const now = new Date()
    const today = formatLocalDate(now)

    const wrapper = mount(DateRangePicker, {
      props: {
        startDate: today,
        endDate: today
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    await wrapper.find('.date-picker-trigger').trigger('click')
    const presetButton = wrapper.findAll('.date-picker-preset').find((node) =>
      node.text().includes('Last 24 Hours')
    )
    expect(presetButton).toBeDefined()

    await presetButton!.trigger('click')
    await wrapper.find('.date-picker-apply').trigger('click')

    const nowAfterClick = new Date()
    const yesterdayAfterClick = new Date(nowAfterClick.getTime() - 24 * 60 * 60 * 1000)
    const expectedStart = formatLocalDate(yesterdayAfterClick)
    const expectedEnd = formatLocalDate(nowAfterClick)

    expect(wrapper.emitted('update:startDate')?.[0]).toEqual([expectedStart])
    expect(wrapper.emitted('update:endDate')?.[0]).toEqual([expectedEnd])
    expect(wrapper.emitted('change')?.[0]).toEqual([
      {
        startDate: expectedStart,
        endDate: expectedEnd,
        preset: 'last24Hours'
      }
    ])
  })

  it('emits the last48Hours preset with a two-day calendar envelope', async () => {
    const now = new Date()
    const today = formatLocalDate(now)
    const wrapper = mount(DateRangePicker, {
      props: { startDate: today, endDate: today },
      global: { stubs: { Icon: true } }
    })

    await wrapper.find('.date-picker-trigger').trigger('click')
    const presetButton = wrapper.findAll('.date-picker-preset').find((node) =>
      node.text().includes('Last 48 Hours')
    )
    expect(presetButton).toBeDefined()
    await presetButton!.trigger('click')
    await wrapper.find('.date-picker-apply').trigger('click')

    const start = new Date()
    start.setTime(start.getTime() - 48 * 60 * 60 * 1000)
    expect(wrapper.emitted('change')?.[0]).toEqual([{
      startDate: formatLocalDate(start),
      endDate: formatLocalDate(new Date()),
      preset: 'last48Hours'
    }])
  })

  it.each([
    ['Last 7 Days', 7, '7days'],
    ['Last 14 Days', 14, '14days'],
    ['Last 30 Days', 30, '30days'],
  ] as const)('keeps the %s envelope anchored at the full rolling start', async (label, days, preset) => {
    const today = formatLocalDate(new Date())
    const wrapper = mount(DateRangePicker, {
      props: { startDate: today, endDate: today },
      global: { stubs: { Icon: true } }
    })

    await wrapper.find('.date-picker-trigger').trigger('click')
    const presetButton = wrapper.findAll('.date-picker-preset').find((node) => node.text() === label)
    expect(presetButton).toBeDefined()
    await presetButton!.trigger('click')
    await wrapper.find('.date-picker-apply').trigger('click')

    const expectedStart = new Date()
    expectedStart.setDate(expectedStart.getDate() - days)
    expect(wrapper.emitted('change')?.[0]).toEqual([{
      startDate: formatLocalDate(expectedStart),
      endDate: formatLocalDate(new Date()),
      preset,
    }])
  })
})
