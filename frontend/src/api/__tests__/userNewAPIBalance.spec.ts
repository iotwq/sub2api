import { beforeEach, describe, expect, it, vi } from 'vitest'

const { getMock } = vi.hoisted(() => ({ getMock: vi.fn() }))

vi.mock('@/api/client', () => ({
  apiClient: { get: getMock },
  buildGatewayUrl: (path: string) => `https://api.example.com${path}`,
}))

import { generateNewAPIBalanceAccessToken } from '@/api/user'

describe('generateNewAPIBalanceAccessToken', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('calls the root NewAPI token endpoint and returns its one-time token', async () => {
    getMock.mockResolvedValue({
      data: { success: true, message: '', data: 'sub_bal_generated-token' },
    })

    await expect(generateNewAPIBalanceAccessToken()).resolves.toBe('sub_bal_generated-token')
    expect(getMock).toHaveBeenCalledWith('https://api.example.com/api/user/token')
  })
})
