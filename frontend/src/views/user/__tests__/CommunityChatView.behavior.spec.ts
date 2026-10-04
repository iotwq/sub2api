import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import CommunityChatView from '@/views/user/CommunityChatView.vue'

const mocks = vi.hoisted(() => ({
  auth: { user: { id: 1 }, isAdmin: false },
  app: { showError: vi.fn() },
  api: Object.fromEntries(['listMessages', 'listMessagesBefore', 'listMessagesAfter', 'searchMessages', 'sendTextMessage', 'uploadFileMessage', 'establishAttachmentSession', 'getDirectUnread', 'listDirectConversations', 'listDirectMessages', 'listDirectMessagesBefore', 'listDirectMessagesAfter', 'sendDirectTextMessage', 'uploadDirectFileMessage', 'markDirectRead', 'markCommunityChatMessagesSeen', 'searchDirectUsers'].map(k => [k, vi.fn()])),
  realtime: null as any,
}))
vi.mock('@/api', () => ({ communityChatAPI: mocks.api }))
vi.mock('@/stores', () => ({ useAppStore: () => mocks.app, useAuthStore: () => mocks.auth }))
vi.mock('vue-i18n', async () => {
  const { ref } = await import('vue')
  return { useI18n: () => ({ t: (s: string) => s, locale: ref('zh-CN') }) }
})
vi.mock('@/composables/useCommunityChatRealtime', () => ({ useCommunityChatRealtime: (options: any) => {
  mocks.realtime = options
  return { connected: { value: true }, connecting: { value: false }, start() {}, stop() {} }
} }))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<div><slot /></div>' } }))
vi.mock('@/components/common/BaseDialog.vue', () => ({ default: { props: ['show'], template: '<div v-if="show"><slot /></div>' } }))
vi.mock('@/components/community/CommunityEmojiPicker.vue', () => ({ default: { template: '<span />' } }))
vi.mock('@/components/icons/Icon.vue', () => ({ default: { template: '<span />' } }))

const msg = (id: number, user = 2, conversation = 2) => ({ id, user_id: user, conversation_user_id: conversation, username: 'user', avatar_url: '', message_type: 'text', content: `message ${id}`, image_url: '', image_mime_type: '', image_size_bytes: 0, reply_to_username: '', reply_to_content: '', reply_to_message_type: '', reply_to_image_url: '', sent_at: '2026-09-15T00:00:00Z', created_at: '2026-09-15T00:00:00Z', updated_at: '2026-09-15T00:00:00Z' })
const page = (items: any[], total = items.length) => ({ items, total, page: 1, page_size: 80, pages: 1 })
const conversation = (id: number) => ({ user_id: id, username: `user ${id}`, avatar_url: '', unread_count: 0, updated_at: '2026-09-15T00:00:00Z' })
function deferred() { let resolve!: (v: any) => void; let reject!: (reason: any) => void; const promise = new Promise((r, j) => { resolve = r; reject = j }); return { promise, resolve, reject } }
let wrapper: any
let view: any
async function settle() { for (let i=0;i<5;i++) { await flushPromises(); await new Promise(r => setTimeout(r, 1)) } }
async function open() { wrapper = mount(CommunityChatView); view = wrapper.vm.$.setupState; await settle() }
beforeEach(() => {
  vi.clearAllMocks()
  mocks.auth.isAdmin = false
  for (const fn of Object.values(mocks.api)) fn.mockReset()
  mocks.api.establishAttachmentSession.mockResolvedValue(undefined)
  mocks.api.getDirectUnread.mockResolvedValue({ total: 0, conversations: [] })
  mocks.api.listMessages.mockResolvedValue(page([msg(1)]))
  mocks.api.listMessagesBefore.mockResolvedValue([])
  mocks.api.listMessagesAfter.mockResolvedValue([])
  mocks.api.listDirectMessages.mockResolvedValue(page([msg(1)]))
  mocks.api.listDirectMessagesAfter.mockResolvedValue([])
  mocks.api.listDirectMessagesBefore.mockResolvedValue([])
  mocks.api.listDirectConversations.mockResolvedValue(page([conversation(2)]))
  mocks.api.markDirectRead.mockResolvedValue(1)
  mocks.api.sendTextMessage.mockResolvedValue(msg(99, 1))
  mocks.api.sendDirectTextMessage.mockResolvedValue(msg(99, 1))
  vi.stubGlobal('requestAnimationFrame', (cb: any) => setTimeout(() => cb(performance.now()), 0))
  Object.defineProperty(HTMLElement.prototype, 'scrollTo', { configurable: true, value: function(options: any) { this.scrollTop = options.top } })
})
afterEach(() => { const element = wrapper?.element; wrapper?.unmount(); element?.remove(); wrapper = null; vi.useRealTimers(); vi.unstubAllGlobals(); vi.restoreAllMocks() })

describe('CommunityChatView interaction regressions', () => {
  it('waits for reconnect catch-up before acknowledging a newer live private message', async () => {
    await open()
    await view.openDirectDialog()
    mocks.api.markDirectRead.mockClear()
    const catchUp = deferred()
    mocks.api.listDirectMessagesAfter.mockReturnValueOnce(catchUp.promise)
    mocks.realtime.onConnected()
    view.handleSocketEvent({ type: 'direct_message_created', message: msg(3) })
    await settle()
    expect(mocks.api.markDirectRead).not.toHaveBeenCalled()
    catchUp.resolve([msg(2)])
    await settle()
    expect(view.directMessages.map((m: any) => m.id)).toEqual([1, 2, 3])
    expect(mocks.api.markDirectRead).toHaveBeenCalledWith(3, undefined)
  })

  it.each([false, true])('shows upload progress and keeps failed upload drafts (private=%s)', async (privateChat) => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    await open()
    if (privateChat) await view.openDirectDialog()
    const file = new File(['upload'], 'file.pdf', { type: 'application/pdf' })
    const upload = deferred()
    const api = privateChat ? mocks.api.uploadDirectFileMessage : mocks.api.uploadFileMessage
    api.mockReturnValueOnce(upload.promise)
    if (privateChat) { view.directDraft = '文件说明'; view.setSelectedDirectFile(file) }
    else { view.draft = '文件说明'; view.setSelectedFile(file) }
    const sending = privateChat ? view.handleDirectSend() : view.handleSend()
    await settle()
    expect(wrapper.find('progress').exists()).toBe(true)
    const reportProgress = api.mock.calls[0][3]
    reportProgress({ loaded: 50, total: 100 })
    await settle()
    expect(wrapper.find('progress').attributes('value')).toBe('50')
    reportProgress({ loaded: 100, total: 100 })
    await settle()
    expect(wrapper.text()).toContain('communityChat.uploadProcessing')
    expect(privateChat ? view.sendingDirect : view.sending).toBe(true)
    upload.reject(new Error('network unavailable'))
    await sending
    await settle()
    expect(wrapper.find('progress').exists()).toBe(false)
    expect(privateChat ? view.directDraft : view.draft).toBe('文件说明')
    expect(privateChat ? view.selectedDirectFile : view.selectedFile).toBe(file)
  })

  it('loads text history without waiting for the attachment session', async () => {
    const session = deferred()
    mocks.api.establishAttachmentSession.mockReturnValueOnce(session.promise)
    await open()
    expect(wrapper.text()).toContain('message 1')
    session.resolve(undefined)
  })

  it('recovers an initially failed attachment session and retries only failed media', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    mocks.api.establishAttachmentSession.mockRejectedValueOnce(new Error('offline'))
    mocks.api.listMessages.mockResolvedValue(page([1, 2].map(id => ({ ...msg(id), message_type: 'image', image_url: `/uploads/${id}.png`, image_mime_type: 'image/png' }))))
    await open()
    document.body.appendChild(wrapper.element)
    const images = wrapper.findAll('.community-message-image')
    const failedReload = vi.spyOn(images[0].element, 'src', 'set')
    const healthyReload = vi.spyOn(images[1].element, 'src', 'set')
    await images[0].trigger('error')
    const timeout = vi.spyOn(window, 'setTimeout')
    mocks.realtime.onConnected()
    await settle()
    expect(timeout).toHaveBeenCalled()
    expect(mocks.api.establishAttachmentSession).toHaveBeenCalledTimes(2)
    expect(failedReload).toHaveBeenCalledTimes(1)
    expect(healthyReload).not.toHaveBeenCalled()
  })

  it('retries attachment initialization after 30 seconds and stops retries on unmount', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    vi.useFakeTimers()
    mocks.api.establishAttachmentSession.mockRejectedValueOnce(new Error('offline'))
    wrapper = mount(CommunityChatView)
    view = wrapper.vm.$.setupState
    await vi.advanceTimersByTimeAsync(30000)
    expect(mocks.api.establishAttachmentSession).toHaveBeenCalledTimes(2)
    wrapper.unmount()
    wrapper = null
    await vi.advanceTimersByTimeAsync(5 * 60 * 1000)
    expect(mocks.api.establishAttachmentSession).toHaveBeenCalledTimes(2)
  })

  it('coalesces attachment recovery on reconnect and returning to the page', async () => {
    await open()
    mocks.api.establishAttachmentSession.mockClear()
    const session = deferred()
    mocks.api.establishAttachmentSession.mockReturnValueOnce(session.promise)
    mocks.realtime.onConnected()
    document.dispatchEvent(new Event('visibilitychange'))
    await settle()
    expect(mocks.api.establishAttachmentSession).toHaveBeenCalledTimes(1)
    session.resolve(undefined)
    await settle()
  })

  it.each(['direct_message_created', 'direct_message_deleted'])('refreshes admin conversations quietly without losing loaded pages (%s)', async (type) => {
    mocks.auth.isAdmin = true
    const first = Array.from({ length: 30 }, (_, i) => conversation(i + 2))
    const second = Array.from({ length: 30 }, (_, i) => conversation(i + 32))
    mocks.api.listDirectConversations.mockResolvedValueOnce(page(first, 90)).mockResolvedValueOnce(page(second, 90))
    await open()
    await view.openDirectDialog()
    await view.loadMoreDirectConversations()
    const refresh = deferred()
    mocks.api.listDirectConversations.mockImplementation((pageNumber: number) => pageNumber === 1 ? refresh.promise : Promise.resolve(page(second, 90)))
    view.handleSocketEvent({ type, id: 90, message: msg(90, 3, 3) })
    await settle()
    expect(view.loadingDirectConversations).toBe(false)
    expect(view.directConversations).toHaveLength(60)
    refresh.resolve(page([{ ...first[0], username: 'updated' }, ...first.slice(1)], 90))
    await settle()
    expect(view.directConversations).toHaveLength(60)
    expect(view.directConversationPage).toBe(2)
    expect(view.directConversations[0].username).toBe('updated')
  })

  it('coalesces incoming conversation refreshes and refetches after a pending refresh', async () => {
    mocks.auth.isAdmin = true
    await open()
    await view.openDirectDialog()
    mocks.api.listDirectConversations.mockClear()
    const old = deferred()
    mocks.api.listDirectConversations.mockReturnValueOnce(old.promise).mockResolvedValue(page([{ ...conversation(2), username: 'newest' }]))
    for (let i = 0; i < 5; i++) view.handleSocketEvent({ type: 'direct_message_created', message: msg(90 + i, 3, 3) })
    await settle()
    expect(mocks.api.listDirectConversations).toHaveBeenCalledTimes(1)
    old.resolve(page([conversation(2)]))
    await settle()
    expect(mocks.api.listDirectConversations).toHaveBeenCalledTimes(2)
    expect(view.directConversations[0].username).toBe('newest')
  })

  it.each([false, true])('typing with 800 messages does not reformat history (private=%s)', async (privateChat) => {
    const history = Array.from({ length: 800 }, (_, i) => msg(i + 1))
    mocks.api.listMessages.mockResolvedValue(page(history))
    mocks.api.listDirectMessages.mockResolvedValue(page(history))
    await open()
    if (privateChat) await view.openDirectDialog()
    const dateReads = vi.spyOn(Date.prototype, 'getTime')
    const formatters = vi.spyOn(Intl, 'DateTimeFormat')
    await wrapper.findAll('textarea')[privateChat ? 1 : 0].setValue('新输入')
    expect(formatters).not.toHaveBeenCalled()
    expect(dateReads).not.toHaveBeenCalled()
  })

  it('updates memoized history for live edits, search, deletion state and locale', async () => {
    mocks.auth.isAdmin = true
    await open()
    view.upsertMessage({ ...msg(1), content: 'updated text' })
    await settle()
    expect(wrapper.find('#community-message-1').text()).toContain('updated text')
    view.searchContextMessages = [...view.messages]
    view.searchResults = [...view.messages]
    view.currentSearchResultIndex = 0
    view.activeSearchQuery = 'updated'
    await settle()
    expect(wrapper.find('#community-message-1 mark').text()).toBe('updated')
    view.deletingMessageId = 1
    await settle()
    expect(wrapper.find('#community-message-1 .community-delete-button').attributes('disabled')).toBeDefined()
    const formatters = vi.spyOn(Intl, 'DateTimeFormat')
    view.locale = 'en-US'
    await settle()
    expect(formatters).toHaveBeenCalledTimes(1)
  })

  it.each([false, true])('does not send IME confirmation or Shift+Enter (private=%s)', async (privateChat) => {
    await open()
    if (privateChat) await view.openDirectDialog()
    const input = wrapper.findAll('textarea')[privateChat ? 1 : 0]
    await input.setValue('正在输入中文')
    const send = privateChat ? mocks.api.sendDirectTextMessage : mocks.api.sendTextMessage
    await input.trigger('keydown', { key: 'Enter', isComposing: true })
    await input.trigger('keydown', { key: 'Enter', keyCode: 229 })
    await input.trigger('keydown', { key: 'Enter', shiftKey: true })
    expect(send).not.toHaveBeenCalled()
    await input.trigger('keydown', { key: 'Enter' })
    expect(send).toHaveBeenCalledTimes(1)
  })

  it('keeps separate admin drafts when switching recipients, including new conversations', async () => {
    mocks.auth.isAdmin = true
    await open()
    await view.openDirectDialog()
    view.directDraft = '给 A 的回复'
    await view.selectDirectUser(conversation(3))
    expect(view.directDraft).toBe('')
    await view.handleDirectSend()
    expect(mocks.api.sendDirectTextMessage).not.toHaveBeenCalled()
    view.directDraft = '给 B 的回复'
    await view.selectDirectUser(conversation(2))
    expect(view.directDraft).toBe('给 A 的回复')
    await view.selectDirectUser(conversation(3))
    expect(view.directDraft).toBe('给 B 的回复')
  })

  it('an in-flight send only clears the submitted recipient draft', async () => {
    mocks.auth.isAdmin = true
    await open()
    await view.openDirectDialog()
    view.directDraft = '相同文字'
    const send = deferred()
    mocks.api.sendDirectTextMessage.mockReturnValueOnce(send.promise)
    const sending = view.handleDirectSend()
    await view.selectDirectUser(conversation(3))
    view.directDraft = '相同文字'
    send.resolve(msg(8, 1, 2))
    await sending
    expect(view.directDraft).toBe('相同文字')
    await view.selectDirectUser(conversation(2))
    expect(view.directDraft).toBe('')
  })

  it('acknowledges only the latest rendered private message and waits for initial history', async () => {
    await open()
    const history = deferred()
    mocks.api.listDirectMessages.mockReturnValueOnce(history.promise)
    const opening = view.openDirectDialog()
    await settle()
    view.handleSocketEvent({ type: 'direct_message_created', message: msg(2) })
    await settle()
    expect(mocks.api.markDirectRead).not.toHaveBeenCalled()
    history.resolve(page([msg(1)]))
    await opening
    expect(mocks.api.markDirectRead).toHaveBeenCalledWith(2, undefined)
  })

  it('merges a delayed private history response with live messages from the same open session', async () => {
    await open()
    const history = deferred()
    mocks.api.listDirectMessages.mockReturnValueOnce(history.promise)
    const opening = view.openDirectDialog()
    await settle()
    view.handleSocketEvent({ type: 'direct_message_created', message: msg(2) })
    await settle()
    history.resolve(page([msg(1)]))
    await opening
    expect(view.directMessages.map((m: any) => m.id)).toEqual([1, 2])
  })

  it('reopening replaces stale local history while retaining new live events', async () => {
    await open()
    await view.openDirectDialog()
    view.closeDirectDialog()
    mocks.api.listDirectMessages.mockResolvedValue(page([msg(3)]))
    await view.openDirectDialog()
    expect(view.directMessages.map((m: any) => m.id)).toEqual([3])
  })

  it('ignores a previous private conversation response after switching users', async () => {
    mocks.auth.isAdmin = true
    await open()
    const history = deferred()
    mocks.api.listDirectMessages.mockReturnValueOnce(history.promise)
    const opening = view.openDirectDialog()
    await settle()
    mocks.api.listDirectMessages.mockResolvedValue(page([msg(9, 3, 3)]))
    await view.selectDirectUser(conversation(3))
    history.resolve(page([msg(1)]))
    await opening
    expect(view.directMessages.map((m: any) => m.id)).toEqual([9])
  })

  it('starts group catch-up before unread lookup allows a live message to advance the cursor', async () => {
    await open()
    const unread = deferred()
    mocks.api.getDirectUnread.mockReturnValueOnce(unread.promise)
    mocks.api.listMessagesAfter.mockResolvedValueOnce([msg(2), msg(3)])
    const recovering = view.recoverAfterRealtimeConnect()
    view.handleSocketEvent({ type: 'message_created', message: msg(3) })
    unread.resolve({ total: 0, conversations: [] })
    await recovering
    expect(mocks.api.listMessagesAfter).toHaveBeenCalledWith(1, 100)
    expect(view.messages.map((m: any) => m.id)).toEqual([1, 2, 3])
  })

  it.each(['group', 'direct'])('reloads an empty %s conversation after reconnect', async scope => {
    mocks.api.listMessages.mockResolvedValue(page([]))
    mocks.api.listDirectMessages.mockResolvedValue(page([]))
    await open()
    if (scope === 'direct') await view.openDirectDialog()
    mocks.api.listMessages.mockResolvedValue(page([msg(2)]))
    mocks.api.listDirectMessages.mockResolvedValue(page([msg(2)]))
    await view.recoverAfterRealtimeConnect()
    expect((scope === 'direct' ? view.directMessages : view.messages).map((m: any) => m.id)).toEqual([2])
  })

  it('starts direct catch-up independently of a pending group recovery', async () => {
    await open()
    await view.openDirectDialog()
    const group = deferred()
    mocks.api.listMessagesAfter.mockReturnValueOnce(group.promise)
    mocks.api.listDirectMessagesAfter.mockResolvedValueOnce([msg(2)])
    const recovering = view.recoverAfterRealtimeConnect()
    await settle()
    expect(mocks.api.listDirectMessagesAfter).toHaveBeenCalledWith(1, undefined, 100)
    group.resolve([])
    await recovering
    expect(view.directMessages.map((m: any) => m.id)).toEqual([1, 2])
  })
})
