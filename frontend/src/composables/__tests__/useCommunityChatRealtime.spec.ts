import { afterEach, describe, expect, it, vi } from 'vitest'

import { useCommunityChatRealtime } from '../useCommunityChatRealtime'

class FakeWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3
  static instances: FakeWebSocket[] = []

  readyState = FakeWebSocket.CONNECTING
  onopen: (() => void) | null = null
  onmessage: ((event: MessageEvent<string>) => void) | null = null
  onerror: (() => void) | null = null
  onclose: (() => void) | null = null
  close = vi.fn(() => {
    this.readyState = FakeWebSocket.CLOSED
    this.onclose?.()
  })

  constructor(public url: string, public protocols: string[]) {
    FakeWebSocket.instances.push(this)
  }

  open(): void {
    this.readyState = FakeWebSocket.OPEN
    this.onopen?.()
  }

  receive(data: string): void {
    this.onmessage?.({ data } as MessageEvent<string>)
  }
}

afterEach(() => {
  localStorage.clear()
  FakeWebSocket.instances = []
  vi.unstubAllGlobals()
})

describe('community chat shared realtime connection', () => {
  it('shares one WebSocket and closes it after the last subscriber stops', () => {
    vi.stubGlobal('WebSocket', FakeWebSocket)
    localStorage.setItem('auth_token', 'header.payload.signature')
    const firstEvent = vi.fn()
    const secondEvent = vi.fn()
    const first = useCommunityChatRealtime({ onEvent: firstEvent })
    const second = useCommunityChatRealtime({ onEvent: secondEvent })

    first.start()
    second.start()

    expect(FakeWebSocket.instances).toHaveLength(1)
    const socket = FakeWebSocket.instances[0]
    socket.open()
    socket.receive(JSON.stringify({ type: 'message_deleted', id: 7 }))
    expect(firstEvent).toHaveBeenCalledWith({ type: 'message_deleted', id: 7 })
    expect(secondEvent).toHaveBeenCalledWith({ type: 'message_deleted', id: 7 })

    first.stop()
    expect(socket.close).not.toHaveBeenCalled()
    second.stop()
    expect(socket.close).toHaveBeenCalledTimes(1)
  })
})
