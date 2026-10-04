export interface OpenAIOAuthProtection {
  enabled: boolean
  max_concurrency: number
  wait_seconds: number
  strict_rpm_enabled: boolean
  rpm: number
  burst: number
  adaptive_enabled: boolean
  adaptive_mode: 'automatic' | 'observe'
  min_concurrency: number
  failure_threshold: number
  failure_window_seconds: number
  recovery_seconds: number
  tls_profile: 'standard' | 'nodejs24' | 'nodejs22'
  integrity_mode: 'off' | 'observe' | 'enforce'
}

export function readOpenAIOAuthProtection(value?: unknown): OpenAIOAuthProtection {
  const defaults: OpenAIOAuthProtection = {
    enabled: false, max_concurrency: 4, wait_seconds: 30,
    strict_rpm_enabled: true, rpm: 60, burst: 5,
    adaptive_enabled: true, adaptive_mode: 'automatic', min_concurrency: 1,
    failure_threshold: 3, failure_window_seconds: 60, recovery_seconds: 60,
    tls_profile: 'standard', integrity_mode: 'observe'
  }
  return value && typeof value === 'object' && !Array.isArray(value)
    ? { ...defaults, ...value }
    : defaults
}
