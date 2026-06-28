import { describe, expect, it } from 'vitest'

import {
  getPaymentVisibleMethodSourceOptions,
  normalizePaymentVisibleMethodSource,
} from '@/api/admin/settings'

describe('admin settings payment visible method helpers', () => {
  it('normalizes aliases into canonical source keys per visible method', () => {
    expect(normalizePaymentVisibleMethodSource('alipay', 'official')).toBe('official_alipay')
    expect(normalizePaymentVisibleMethodSource('alipay', 'alipay_direct')).toBe('official_alipay')
    expect(normalizePaymentVisibleMethodSource('alipay', 'easypay')).toBe('easypay_alipay')
    expect(normalizePaymentVisibleMethodSource('alipay', 'personal_qrcode')).toBe('personal_qrcode_alipay')
    expect(normalizePaymentVisibleMethodSource('alipay', 'personal')).toBe('personal_qrcode_alipay')

    expect(normalizePaymentVisibleMethodSource('wxpay', 'official')).toBe('official_wxpay')
    expect(normalizePaymentVisibleMethodSource('wxpay', 'wechat')).toBe('official_wxpay')
    expect(normalizePaymentVisibleMethodSource('wxpay', 'easypay')).toBe('easypay_wxpay')
    expect(normalizePaymentVisibleMethodSource('wxpay', 'personal_qr')).toBe('personal_qrcode_wxpay')
  })

  it('rejects unknown or cross-method source values', () => {
    expect(normalizePaymentVisibleMethodSource('alipay', 'official_wxpay')).toBe('')
    expect(normalizePaymentVisibleMethodSource('wxpay', 'official_alipay')).toBe('')
    expect(normalizePaymentVisibleMethodSource('alipay', 'personal_qrcode_wxpay')).toBe('')
    expect(normalizePaymentVisibleMethodSource('wxpay', 'personal_qrcode_alipay')).toBe('')
    expect(normalizePaymentVisibleMethodSource('alipay', 'unknown')).toBe('')
    expect(normalizePaymentVisibleMethodSource('wxpay', null)).toBe('')
  })

  it('exposes method-scoped source options instead of arbitrary strings', () => {
    expect(getPaymentVisibleMethodSourceOptions('alipay').map(({ value, labelEn }) => ({ value, labelEn }))).toEqual([
      { value: '', labelEn: 'Not configured' },
      { value: 'official_alipay', labelEn: 'Official Alipay' },
      { value: 'easypay_alipay', labelEn: 'EasyPay Alipay' },
      { value: 'personal_qrcode_alipay', labelEn: 'Personal QR Alipay' },
    ])

    expect(getPaymentVisibleMethodSourceOptions('wxpay').map(({ value, labelEn }) => ({ value, labelEn }))).toEqual([
      { value: '', labelEn: 'Not configured' },
      { value: 'official_wxpay', labelEn: 'Official WeChat Pay' },
      { value: 'easypay_wxpay', labelEn: 'EasyPay WeChat Pay' },
      { value: 'personal_qrcode_wxpay', labelEn: 'Personal QR WeChat Pay' },
    ])
  })
})
