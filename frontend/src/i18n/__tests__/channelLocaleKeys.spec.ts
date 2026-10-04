import { describe, expect, it } from 'vitest'

import enAdminChannels from '../locales/en/admin/channels'
import zhAdminChannels from '../locales/zh/admin/channels'

describe.each([
  ['en', enAdminChannels],
  ['zh', zhAdminChannels]
])('channel locale keys (%s)', (_locale, messages) => {
  it('defines the monitor timeout message', () => {
    expect(messages.channelMonitor.runTimeout).toBeTruthy()
  })

  it('defines the per-request video billing hint', () => {
    expect(messages.channels.form.perRequestVideoHint).toBeTruthy()
  })
})
