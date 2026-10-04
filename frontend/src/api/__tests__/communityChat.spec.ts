import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  buildCommunityChatWebSocketProtocols,
  buildCommunityChatWebSocketUrl,
  deleteDirectMessage,
  establishAttachmentSession,
  getDirectUnread,
  getCommunityChatLastSeenMessageId,
  listDirectMessagesAfter,
  listMessagesBefore,
  markDirectRead,
  markCommunityChatMessagesSeen,
  searchMessages,
  searchDirectUsers,
  uploadDirectFileMessage,
  uploadFileMessage,
} from '../communityChat'
import { apiClient } from '../client'

afterEach(() => {
  vi.restoreAllMocks()
  localStorage.clear()
})

describe('community chat notification cursors', () => {
  it('keeps separate monotonic cursors per user and message scope', () => {
    expect(getCommunityChatLastSeenMessageId(42, 'group')).toBe(0)

    markCommunityChatMessagesSeen(42, 'group', 12)
    markCommunityChatMessagesSeen(42, 'group', 8)
    markCommunityChatMessagesSeen(42, 'direct', 5)

    expect(getCommunityChatLastSeenMessageId(42, 'group')).toBe(12)
    expect(getCommunityChatLastSeenMessageId(42, 'direct')).toBe(5)
    expect(getCommunityChatLastSeenMessageId(7, 'group')).toBe(0)
  })

  it('notifies same-page components when a message scope is marked as seen', () => {
    const listener = vi.fn()
    window.addEventListener('community-chat:seen', listener)

    markCommunityChatMessagesSeen(42, 'direct', 9)

    expect(listener).toHaveBeenCalledTimes(1)
    const event = listener.mock.calls[0][0] as CustomEvent
    expect(event.detail).toEqual({ userId: 42, scope: 'direct', messageId: 9 })
    window.removeEventListener('community-chat:seen', listener)
  })
})

describe('community chat WebSocket authentication', () => {
  it('keeps the JWT out of the WebSocket URL', () => {
    const url = new URL(buildCommunityChatWebSocketUrl())

    expect(url.pathname).toBe('/api/v1/community-chat/ws')
    expect(url.searchParams.has('token')).toBe(false)
  })

  it('passes the JWT through the prefixed WebSocket subprotocol', () => {
    expect(buildCommunityChatWebSocketProtocols(' header.payload.signature ')).toEqual([
      'sub2api-chat',
      'jwt.header.payload.signature'
    ])
  })
})

describe('community chat attachment authentication', () => {
  it.each([false, true])('uses an upload-only timeout and reports progress (private=%s)', async (privateChat) => {
    const post = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: { id: 9 } })
    const file = new File(['upload'], 'file.pdf', { type: 'application/pdf' })
    const progress = vi.fn()
    if (privateChat) await uploadDirectFileMessage(file, '', 42, progress)
    else await uploadFileMessage(file, '', undefined, progress)
    const config = post.mock.calls[0][2]!
    expect(config.timeout).toBe(10 * 60 * 1000)
    expect(apiClient.defaults.timeout).toBe(30000)
    const event = { loaded: 512, total: 1024, bytes: 512, lengthComputable: true }
    config.onUploadProgress!(event)
    expect(progress).toHaveBeenCalledWith(event)
  })

  it('establishes an authenticated attachment session', async () => {
    const post = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: undefined })

    await establishAttachmentSession()

    expect(post).toHaveBeenCalledWith('/community-chat/attachments/session')
  })
})

describe('community chat direct messages', () => {
  it('uploads a private attachment with its target user', async () => {
    const post = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: { id: 9 } })
    const file = new File(['private'], 'private.pdf', { type: 'application/pdf' })

    await uploadDirectFileMessage(file, '说明', 42)

    const [path, body] = post.mock.calls[0]
    expect(path).toBe('/community-chat/direct/files')
    expect(body).toBeInstanceOf(FormData)
    expect((body as FormData).get('file')).toBe(file)
    expect((body as FormData).get('content')).toBe('说明')
    expect((body as FormData).get('user_id')).toBe('42')
  })

  it('lets admins search users before a conversation exists', async () => {
    const get = vi.spyOn(apiClient, 'get').mockResolvedValue({ data: { items: [], total: 0, page: 1, page_size: 20 } })

    await searchDirectUsers('alice', 1, 20)

    expect(get).toHaveBeenCalledWith('/community-chat/direct/users', {
      params: { search: 'alice', page: 1, page_size: 20 },
    })
  })
})

describe('community chat reliability endpoints', () => {
  it('searches all community chat history with a trimmed keyword', async () => {
    const get = vi.spyOn(apiClient, 'get').mockResolvedValue({ data: { items: [], total: 0, page: 2, page_size: 50 } })

    await searchMessages('  模型价格  ', 2, 50)

    expect(get).toHaveBeenCalledWith('/community-chat/messages', {
      params: { search: '模型价格', page: 2, page_size: 50 },
    })
  })

  it('uses cursor parameters for group and direct history', async () => {
    const get = vi.spyOn(apiClient, 'get').mockResolvedValue({ data: [] })

    await listMessagesBefore(100, 40)
    await listDirectMessagesAfter(200, 42, 60)

    expect(get).toHaveBeenNthCalledWith(1, '/community-chat/messages', {
      params: { before_id: 100, limit: 40 },
    })
    expect(get).toHaveBeenNthCalledWith(2, '/community-chat/direct/messages', {
      params: { after_id: 200, limit: 60, user_id: 42 },
    })
  })

  it('reads server unread state and exposes direct read/delete actions', async () => {
    const get = vi.spyOn(apiClient, 'get').mockResolvedValue({ data: { total: 2, conversations: [] } })
    const post = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: { last_message_id: 19 } })
    const remove = vi.spyOn(apiClient, 'delete').mockResolvedValue({ data: { id: 19 } })

    await expect(getDirectUnread()).resolves.toEqual({ total: 2, conversations: [] })
    await expect(markDirectRead(19, 42)).resolves.toBe(19)
    await deleteDirectMessage(19)

    expect(get).toHaveBeenCalledWith('/community-chat/direct/unread')
    expect(post).toHaveBeenCalledWith('/community-chat/direct/read', { user_id: 42, last_message_id: 19 })
    expect(remove).toHaveBeenCalledWith('/community-chat/direct/messages/19')
  })
})
