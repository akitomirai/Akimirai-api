import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const copyToClipboard = vi.fn().mockResolvedValue(true)

const messages: Record<string, string> = {
  'keys.endpoints.title': 'API 端点',
  'keys.endpoints.default': '默认',
  'keys.endpoints.copied': '已复制',
  'keys.endpoints.copiedHint': '已复制到剪贴板',
  'keys.endpoints.clickToCopy': '点击可复制此端点',
  'keys.endpoints.testLatency': '测试延迟',
  'keys.endpoints.testingLatency': '正在测试延迟',
  'keys.endpoints.latencyFailed': '延迟测试失败',
  'keys.endpoints.externalSpeedTest': '外部测速',
}

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => messages[key] ?? key,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

import EndpointPopover from '../EndpointPopover.vue'
import Icon from '@/components/icons/Icon.vue'

describe('EndpointPopover', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
    copyToClipboard.mockReset().mockResolvedValue(true)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('将说明提示渲染到 URL 上方而不是旧的 title 图标上', () => {
    const wrapper = mount(EndpointPopover, {
      props: {
        apiBaseUrl: 'https://default.example.com/v1',
        customEndpoints: [
          {
            name: '备用线路',
            endpoint: 'https://backup.example.com/v1',
            description: '自定义说明',
          },
        ],
      },
    })

    expect(wrapper.text()).toContain('自定义说明')
    expect(wrapper.text()).toContain('点击可复制此端点')
    expect(wrapper.find('[role="button"]').attributes('title')).toBeUndefined()
    expect(wrapper.find('[title="自定义说明"]').exists()).toBe(false)
  })

  it('点击 URL 后会复制并切换为已复制提示', async () => {
    const wrapper = mount(EndpointPopover, {
      props: {
        apiBaseUrl: 'https://default.example.com/v1',
        customEndpoints: [],
      },
    })

    await wrapper.find('[role="button"]').trigger('click')
    await flushPromises()

    expect(copyToClipboard).toHaveBeenCalledWith('https://default.example.com/v1', '已复制')
    expect(wrapper.text()).toContain('已复制到剪贴板')
    expect(wrapper.find('button[aria-label="已复制到剪贴板"]').exists()).toBe(true)
  })

  it('点击闪电按钮后测试并显示当前端点延迟', async () => {
    vi.spyOn(Date, 'now').mockReturnValue(1234)

    const wrapper = mount(EndpointPopover, {
      props: {
        apiBaseUrl: 'https://default.example.com/v1',
        customEndpoints: [],
      },
    })

    const performanceNow = vi.spyOn(performance, 'now').mockReturnValue(100)
    const fetchMock = vi.fn().mockImplementation(async () => {
      performanceNow.mockReturnValue(221)
      return {}
    })
    vi.stubGlobal('fetch', fetchMock)

    await wrapper.find('button[aria-label="测试延迟"]').trigger('click')
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledWith(
      'https://default.example.com/v1?_latency_test=1234',
      expect.objectContaining({ cache: 'no-store', mode: 'no-cors' }),
    )
    expect(wrapper.text()).toContain('121ms')
  })

  it('测试进行中时不会重复发起请求，并能显示失败状态', async () => {
    let rejectRequest: ((error: Error) => void) | undefined
    const fetchMock = vi.fn().mockImplementation(() => new Promise((_resolve, reject) => {
      rejectRequest = reject
    }))
    vi.stubGlobal('fetch', fetchMock)

    const wrapper = mount(EndpointPopover, {
      props: {
        apiBaseUrl: 'https://default.example.com/v1',
        customEndpoints: [],
      },
    })

    const latencyButton = wrapper.find('button[aria-label="测试延迟"]')
    await latencyButton.trigger('click')
    await latencyButton.trigger('click')
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(wrapper.find('button[aria-label="正在测试延迟"]').attributes('disabled')).toBeDefined()

    rejectRequest?.(new Error('network failed'))
    await flushPromises()

    expect(wrapper.text()).toContain('延迟测试失败')
  })

  it('使用仪表盘按钮打开原有外部测速链接', () => {
    const wrapper = mount(EndpointPopover, {
      props: {
        apiBaseUrl: 'https://default.example.com/v1',
        customEndpoints: [],
      },
    })

    const externalSpeedLink = wrapper.find('a[aria-label="外部测速"]')
    expect(externalSpeedLink.attributes('href')).toBe(
      'https://www.tcptest.cn/http/https%3A%2F%2Fdefault.example.com%2Fv1',
    )
    expect(externalSpeedLink.attributes('target')).toBe('_blank')
    expect(externalSpeedLink.findComponent(Icon).props('name')).toBe('gauge')
    expect(wrapper.findAllComponents(Icon).map(component => component.props('name'))).toEqual(['gauge', 'bolt'])
  })
})
