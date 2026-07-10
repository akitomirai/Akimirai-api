export type PrivacyFilterType =
  | 'ip_address'
  | 'email'
  | 'phone'
  | 'id_card'
  | 'bank_card'
  | 'api_key'
  | 'token'
  | 'private_key'
  | 'random_string'

export interface PrivacyFilterConfig {
  enabled: boolean
  types: PrivacyFilterType[]
}

export const DEFAULT_PRIVACY_FILTER_TYPES: PrivacyFilterType[] = [
  'ip_address',
  'email',
  'phone',
  'id_card',
  'bank_card',
  'api_key',
  'token',
  'private_key',
  'random_string',
]

export function normalizePrivacyFilterConfig(
  value?: Partial<PrivacyFilterConfig> | null,
): PrivacyFilterConfig {
  const selected = Array.isArray(value?.types)
    ? value.types.filter((item): item is PrivacyFilterType =>
        DEFAULT_PRIVACY_FILTER_TYPES.includes(item as PrivacyFilterType),
      )
    : [...DEFAULT_PRIVACY_FILTER_TYPES]

  const unique = [...new Set(selected)]
  return {
    enabled: value?.enabled === true,
    types: DEFAULT_PRIVACY_FILTER_TYPES.filter((item) => unique.includes(item)),
  }
}
