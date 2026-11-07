import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useLocalChat } from '../useLocalChat'

const values = new Map<string, string>()
const storage = {
  getItem: (key: string) => values.get(key) || null,
  setItem: vi.fn((key: string, value: string) => values.set(key, value)),
  removeItem: (key: string) => values.delete(key),
}

beforeEach(() => {
  vi.useFakeTimers()
  Object.defineProperty(window, 'localStorage', { configurable: true, value: storage })
  Object.defineProperty(window, 'sessionStorage', { configurable: true, value: storage })
  const chat = useLocalChat()
  chat.clearAllSessions()
  chat.flushPendingChatUpdates()
  values.clear()
  storage.setItem.mockClear()
})

afterEach(() => {
  useLocalChat().flushPendingChatUpdates()
  vi.useRealTimers()
})

describe('模型问答本地流式保存', () => {
  it('连续片段不逐次写整组会话，每秒合并一次且内容完整', async () => {
    const chat = useLocalChat()
    const session = chat.createSession({ model: 'gpt-6-sol' })
    const message = chat.appendMessage(session.id, 'assistant', '')!
    storage.setItem.mockClear()
    for (let index = 1; index <= 100; index += 1) chat.updateMessage(session.id, message.id, 'A'.repeat(index))
    expect(storage.setItem).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1000)
    expect(storage.setItem).toHaveBeenCalledTimes(2)
    expect(JSON.parse(values.get('local_chat_sessions_v1')!)[0].messages[0].content).toBe('A'.repeat(100))
  })

  it('最终刷新立即保存最新正文和状态并取消延迟写入', async () => {
    const chat = useLocalChat()
    const session = chat.createSession()
    const message = chat.appendMessage(session.id, 'assistant', '')!
    storage.setItem.mockClear()
    chat.updateMessage(session.id, message.id, '最终内容', { status: 'stopped', elapsedMs: 50 })
    chat.flushPendingChatUpdates()
    expect(storage.setItem).toHaveBeenCalledTimes(2)
    expect(JSON.parse(values.get('local_chat_sessions_v1')!)[0].messages[0]).toMatchObject({
      content: '最终内容', response: { status: 'stopped', elapsedMs: 50 },
    })
    await vi.advanceTimersByTimeAsync(1000)
    expect(storage.setItem).toHaveBeenCalledTimes(2)
  })

  it('持久化会话时不保存 API Key 和附件正文，草稿只保留在当前会话存储', () => {
    const chat = useLocalChat()
    const session = chat.createSession({ apiKey: 'sk-secret', baseUrl: 'https://gateway.test/v1' })
    chat.appendMessage(session.id, 'user', 'private', [{
      id: 'image', kind: 'image', name: 'secret.png', mimeType: 'image/png', size: 10,
      dataUrl: 'data:image/png;base64,secret',
    }, {
      id: 'text', kind: 'text', name: 'secret.txt', mimeType: 'text/plain', size: 10,
      textContent: 'private content',
    }])
    const persisted = JSON.parse(values.get('local_chat_sessions_v1')!)
    expect(persisted[0].apiKey).toBe('')
    expect(persisted[0].messages[0].attachments[0].dataUrl).toBeUndefined()
    expect(persisted[0].messages[0].attachments[1].textContent).toBeUndefined()

    chat.saveDraftConfig({ apiKey: 'sk-secret', baseUrl: 'https://gateway.test/v1' })
    expect(values.get('local_chat_draft_config_v1')).toBeUndefined()
    expect(chat.consumeDraftConfig()?.apiKey).toBe('sk-secret')
    expect(values.get('local_chat_draft_config_v1')).toBeUndefined()
  })
})
