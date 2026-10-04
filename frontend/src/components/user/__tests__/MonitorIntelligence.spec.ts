import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import MonitorIntelligenceTimeline from '../monitor/MonitorIntelligenceTimeline.vue'
import MonitorCard from '../monitor/MonitorCard.vue'
import MonitorRunResultDialog from '@/components/admin/monitor/MonitorRunResultDialog.vue'
import MonitorIntelligenceBadge from '@/components/common/MonitorIntelligenceBadge.vue'
import type { MonitorTimelinePoint, UserMonitorView } from '@/api/channelMonitor'

vi.mock('@/utils/featureFlags', () => ({ isChannelMonitorQuotaVisible: () => false }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const { default: messages } = await import('@/i18n/locales/zh/dashboard')
  return { ...actual, useI18n: () => ({
    te: () => true,
    t: (key: string, params: Record<string, unknown> = {}) => {
      const value = key.split('.').reduce<unknown>((node, part) =>
        node && typeof node === 'object' ? (node as Record<string, unknown>)[part] : undefined, messages)
      return typeof value === 'string' ? value.replace(/\{(\w+)\}/g, (_, name: string) => String(params[name] ?? '')) : key
    },
  }) }
})
const global = {}
const buckets: MonitorTimelinePoint[] = [
  { status: 'operational', intelligence: { status: 'passed', answer: '21' }, checked_at: '2026-09-13T10:00:00Z', latency_ms: 1000, ping_latency_ms: 10 },
  { status: 'operational', intelligence: { status: 'failed', reason: 'answer_mismatch', answer: '29' }, checked_at: '2026-09-13T09:55:00Z', latency_ms: 1000, ping_latency_ms: 10 },
  { status: 'error', intelligence: { status: 'inconclusive', reason: 'request_failed' }, checked_at: '2026-09-13T09:50:00Z', latency_ms: null, ping_latency_ms: null },
]

describe('Monitor intelligence display', () => {
  it('renders 60 chronological bars with independent colors and hover explanations', () => {
    const wrapper = mount(MonitorIntelligenceTimeline, { props: { buckets }, global })
    const bars = wrapper.findAll('[data-status]')
    expect(bars).toHaveLength(60)
    expect(bars[0].attributes('data-status')).toBe('untested')
    expect(bars[57].classes()).toContain('bg-gray-300')
    expect(bars[57].attributes('title')).toContain('请求失败，无法判定')
    expect(bars[58].classes()).toContain('bg-red-500')
    expect(bars[58].attributes('title')).toContain('智力不合格')
    expect(bars[58].attributes('title')).toContain('糖果题答案：29')
    expect(bars[58].attributes('title')).not.toContain('最终答案不正确')
    expect(bars[59].classes()).toContain('bg-emerald-500')
    expect(bars[59].attributes('title')).toContain('智力合格')
    expect(bars[59].attributes('title')).toContain('糖果题答案：21')
    expect(bars[59].attributes('aria-label')).toBe(bars[59].attributes('title'))
    expect(bars[59].attributes('title')).toContain(new Date(buckets[0].checked_at).toLocaleString())
  })

  it.each(['passed', 'failed'] as const)('shows unrecorded instead of guessing the answer for legacy %s results', (status) => {
    const wrapper = mount(MonitorIntelligenceBadge, { props: {
      result: { status, reason: status === 'failed' ? 'answer_mismatch' : undefined }, checkedAt: buckets[0].checked_at,
    }, global })
    expect(wrapper.attributes('title')).toContain('糖果题答案：未记录')
    expect(wrapper.attributes('title')).not.toContain('最终答案不正确')
    expect(wrapper.attributes('title')).not.toContain('答案：21')
    expect(wrapper.attributes('title')).toContain(new Date(buckets[0].checked_at).toLocaleString())
  })

  it('preserves incomplete and untested explanations without inventing answers', () => {
    const incomplete = mount(MonitorIntelligenceBadge, { props: { result: buckets[2].intelligence }, global })
    expect(incomplete.attributes('title')).toBe('智力检测未完成 · 请求失败，无法判定')
    const untested = mount(MonitorIntelligenceBadge, { global })
    expect(untested.attributes('title')).toBe('未检测')
  })

  it('shows history only for enabled monitors while preserving the health timeline', async () => {
    const item: UserMonitorView = {
      id: 1, name: 'test', provider: 'openai', group_name: 'group', primary_model: 'main',
      primary_status: 'operational', primary_latency_ms: 1000, primary_ping_latency_ms: 10,
      availability_7d: 100, extra_models: [{ model: 'extra', status: 'operational', latency_ms: 1000, intelligence: { status: 'failed' } }],
      timeline: buckets, intelligence_enabled: false,
    }
    const wrapper = mount(MonitorCard, {
      props: { item, window: '7d', availabilityValue: 100, countdownSeconds: 30 },
      global: { ...global, stubs: { MonitorMetricPair: true, MonitorAvailabilityRow: true, MonitorTimeline: true, ProviderIcon: true } },
    })
    expect(wrapper.find('[data-testid="intelligence-timeline"]').exists()).toBe(false)
    await wrapper.setProps({ item: { ...item, intelligence_enabled: true } })
    expect(wrapper.find('[data-testid="intelligence-timeline"]').exists()).toBe(true)
    expect(wrapper.find('monitor-timeline-stub').exists()).toBe(true)
    expect(wrapper.text()).toContain('智力合格')
    expect(wrapper.text()).toContain('智力不合格')
    expect(wrapper.text()).toContain('extra')
  })

  it('admin run results display per-model intelligence separately from health', () => {
    const wrapper = mount(MonitorRunResultDialog, {
      props: { show: true, results: buckets.map((b, i) => ({ ...b, model: `model-${i}`, message: '' })) },
      global: { ...global, stubs: { BaseDialog: { template: '<div><slot /></div>' }, MonitorQuotaView: true } },
    })
    expect(wrapper.text()).toContain('智力合格')
    expect(wrapper.text()).toContain('智力不合格')
    expect(wrapper.text()).toContain('智力检测未完成')
    const titles = wrapper.findAll('[title]').map((node) => node.attributes('title')).join('\n')
    expect(titles).toContain('智力合格 · 糖果题答案：21')
    expect(titles).toContain('智力不合格 · 糖果题答案：29')
    expect(titles).not.toContain('最终答案不正确')
  })
})
