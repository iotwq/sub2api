import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises } from '@vue/test-utils'
import type { CommunityChatEvent } from '@/api/communityChat'
import { useCommunityChatNotificationsStore } from '../communityChatNotifications'

const mocks = vi.hoisted(() => ({
  listMessages: vi.fn(), getDirectUnread: vi.fn(), start: vi.fn(), stop: vi.fn(),
  subscriber: null as null | { onConnected: () => void; onEvent: (event: CommunityChatEvent) => void },
}))
vi.mock('@/api', () => ({ communityChatAPI: {
  listMessages: mocks.listMessages,
  getDirectUnread: mocks.getDirectUnread,
  getCommunityChatLastSeenMessageId: () => 0,
} }))
vi.mock('@/composables/useCommunityChatRealtime', () => ({
  useCommunityChatRealtime: (subscriber: NonNullable<typeof mocks.subscriber>) => {
    mocks.subscriber = subscriber
    return { start: mocks.start, stop: mocks.stop }
  },
}))

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((done) => { resolve = done })
  return { promise, resolve }
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  mocks.listMessages.mockResolvedValue({ items: [] })
  mocks.getDirectUnread.mockResolvedValue({ total: 0 })
})
afterEach(() => { useCommunityChatNotificationsStore().stopSession() })

describe('login-session community notifications', () => {
  it('retains one subscription and unread state when the same session is reused', async () => {
    const store = useCommunityChatNotificationsStore()
    mocks.getDirectUnread.mockResolvedValue({ total: 2 })
    store.startSession(1)
    await flushPromises()
    store.startSession(1)
    expect(store.directUnread).toBe(true)
    expect(mocks.start).toHaveBeenCalledTimes(1)
    expect(mocks.getDirectUnread).toHaveBeenCalledTimes(1)
    expect(mocks.stop).not.toHaveBeenCalled()
  })

  it('coalesces the initial read and socket-open refresh, but refreshes after reconnect', async () => {
    const group = deferred<{ items: [] }>()
    mocks.listMessages.mockReturnValueOnce(group.promise)
    const store = useCommunityChatNotificationsStore()
    store.startSession(1)
    mocks.subscriber!.onConnected()
    expect(mocks.listMessages).toHaveBeenCalledTimes(1)
    expect(mocks.getDirectUnread).toHaveBeenCalledTimes(1)
    group.resolve({ items: [] })
    await flushPromises()
    mocks.subscriber!.onConnected()
    await flushPromises()
    expect(mocks.listMessages).toHaveBeenCalledTimes(2)
    expect(mocks.getDirectUnread).toHaveBeenCalledTimes(2)
  })

  it('ignores responses from a previous user after the account changes', async () => {
    const oldDirect = deferred<{ total: number }>()
    const oldGroup = deferred<{ items: { id: number; user_id: number }[] }>()
    mocks.getDirectUnread.mockReturnValueOnce(oldDirect.promise)
    mocks.listMessages.mockReturnValueOnce(oldGroup.promise)
    const store = useCommunityChatNotificationsStore()
    store.startSession(1)
    store.startSession(2)
    await flushPromises()
    oldDirect.resolve({ total: 5 })
    oldGroup.resolve({ items: [{ id: 9, user_id: 3 }] })
    await flushPromises()
    expect(store.directUnread).toBe(false)
    expect(store.groupUnread).toBe(false)
    expect(mocks.stop).toHaveBeenCalledTimes(1)
    expect(mocks.start).toHaveBeenCalledTimes(2)
  })

  it('clears state on logout and ignores in-flight results', async () => {
    const pending = deferred<{ total: number }>()
    mocks.getDirectUnread.mockReturnValueOnce(pending.promise)
    const store = useCommunityChatNotificationsStore()
    store.startSession(1)
    mocks.subscriber!.onEvent({ type: 'message_created', message: { id: 10, user_id: 2 } as never })
    expect(store.groupUnread).toBe(true)
    store.stopSession()
    pending.resolve({ total: 3 })
    await flushPromises()
    expect(store.groupUnread).toBe(false)
    expect(store.directUnread).toBe(false)
    expect(mocks.stop).toHaveBeenCalledTimes(1)
  })

  it('preserves a new group notification received while the initial snapshot is pending', async () => {
    const group = deferred<{ items: [] }>()
    mocks.listMessages.mockReturnValueOnce(group.promise)
    const store = useCommunityChatNotificationsStore()
    store.startSession(1)
    mocks.subscriber!.onEvent({ type: 'message_created', message: { id: 10, user_id: 2 } as never })
    group.resolve({ items: [] })
    await flushPromises()
    expect(store.groupUnread).toBe(true)
    window.dispatchEvent(new CustomEvent('community-chat:seen', { detail: { userId: 1, scope: 'group' } }))
    expect(store.groupUnread).toBe(false)
  })
})
