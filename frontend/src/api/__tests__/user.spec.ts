import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const { get } = vi.hoisted(() => ({
  get: vi.fn()
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get
  }
}))

describe('user api oauth binding urls', () => {
  beforeEach(() => {
    get.mockReset()
    vi.resetModules()
    vi.stubEnv('VITE_API_BASE_URL', 'https://api.example.com/api/v1')
  })

  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('builds third-party bind urls against the bind start endpoint', async () => {
    const { buildOAuthBindingStartURL } = await import('@/api/user')

    expect(buildOAuthBindingStartURL('linuxdo', { redirectTo: '/settings/profile' })).toBe(
      'https://api.example.com/api/v1/auth/oauth/linuxdo/bind/start?redirect=%2Fsettings%2Fprofile&intent=bind_current_user'
    )
    expect(
      buildOAuthBindingStartURL('wechat', {
        redirectTo: '/settings/profile',
        wechatOAuthSettings: {
          wechat_oauth_open_enabled: true,
          wechat_oauth_mp_enabled: false,
          wechat_oauth_mobile_enabled: false
        }
      })
    ).toBe(
      'https://api.example.com/api/v1/auth/oauth/wechat/bind/start?redirect=%2Fsettings%2Fprofile&intent=bind_current_user&mode=open'
    )
  })

  it('loads QQ avatar suggestions from the user avatar endpoint', async () => {
    get.mockResolvedValue({
      data: {
        available: true,
        qq: '123456789',
        avatar_url: 'https://q.qlogo.cn/headimg_dl?dst_uin=123456789&spec=640&img_type=jpg'
      }
    })
    const { getQQAvatarSuggestion } = await import('@/api/user')

    await expect(getQQAvatarSuggestion()).resolves.toEqual({
      available: true,
      qq: '123456789',
      avatar_url: 'https://q.qlogo.cn/headimg_dl?dst_uin=123456789&spec=640&img_type=jpg'
    })
    expect(get).toHaveBeenCalledWith('/user/avatar/qq-suggestion')
  })
})
