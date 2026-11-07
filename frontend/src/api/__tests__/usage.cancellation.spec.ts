import { beforeEach, describe, expect, it, vi } from 'vitest'
import { getStats, getDashboardModels, getDashboardSnapshotV2 } from '../usage'

const get = vi.hoisted(() => vi.fn())
vi.mock('../client', () => ({ apiClient: { get } }))
beforeEach(() => { get.mockReset().mockResolvedValue({ data: {} }) })

describe('usage request cancellation', () => {
  it('passes signals to all statistic requests', async () => {
    const signal = new AbortController().signal
    const params = { model: 'demo' }
    await getStats(params, undefined, { signal })
    await getDashboardModels(params, { signal })
    await getDashboardSnapshotV2(params, { signal })
    expect(get.mock.calls.map(call => call[1])).toEqual([
      { params, signal }, { params, signal }, { params, signal },
    ])
  })

  it('preserves the legacy period and API key signature', async () => {
    await getStats('week', 42)
    expect(get).toHaveBeenCalledWith('/usage/stats', { params: { period: 'week', api_key_id: 42 } })
  })
})
