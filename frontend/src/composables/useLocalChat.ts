import { computed, ref } from 'vue'

export type ChatReasoningEffort = 'low' | 'medium' | 'high' | 'xhigh' | 'max'

export type ChatResponseStatus = 'connecting' | 'waiting' | 'thinking' | 'searching' | 'outputting' | 'completed' | 'stopped' | 'failed'

export interface LocalChatResponse {
  status: ChatResponseStatus
  elapsedMs?: number
  reasoningSummary?: string
  sources?: Array<{ url: string; title: string }>
}

export interface LocalChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  attachments?: LocalChatAttachment[]
  createdAt: string
  response?: LocalChatResponse
}

export interface LocalChatAttachment {
  id: string
  kind: 'image' | 'text'
  sourceType?: 'image' | 'text' | 'pdf' | 'docx'
  name: string
  mimeType: string
  size: number
  pageCount?: number
  dataUrl?: string
  textContent?: string
}

export interface LocalChatSession {
  id: string
  title: string
  model: string
  reasoningEffort?: ChatReasoningEffort
  apiKey?: string
  baseUrl?: string
  apiKeyHint: string
  messages: LocalChatMessage[]
  createdAt: string
  updatedAt: string
}

export interface LocalChatDraftConfig {
  apiKey: string
  baseUrl: string
  model?: string
  sessionTitle?: string
  reasoningEffort?: ChatReasoningEffort
}

const CHAT_SESSIONS_KEY = 'local_chat_sessions_v1'
const CHAT_ACTIVE_SESSION_KEY = 'local_chat_active_session_v1'
const CHAT_DRAFT_CONFIG_KEY = 'local_chat_draft_config_v1'

function safeRead<T>(key: string, fallback: T): T {
  try {
    const raw = window.localStorage.getItem(key)
    if (!raw) return fallback
    return JSON.parse(raw) as T
  } catch {
    return fallback
  }
}

function safeWrite(key: string, value: unknown): void {
  try {
    window.localStorage.setItem(key, JSON.stringify(value))
  } catch {
    // ignore storage errors
  }
}

function safeRemove(key: string): void {
  try {
    window.localStorage.removeItem(key)
  } catch {
    // Ignore storage errors.
  }
}

function sanitizeSessionForPersistence(session: LocalChatSession): LocalChatSession {
  return {
    ...session,
    // API keys are intentionally kept only in the active page memory.
    apiKey: '',
    messages: session.messages.map((message) => ({
      ...message,
      // Attachments can contain private documents or large data URLs.
      attachments: message.attachments?.map((attachment) => {
        const persistedAttachment = { ...attachment }
        delete persistedAttachment.dataUrl
        delete persistedAttachment.textContent
        return persistedAttachment
      }),
    })),
  }
}

const sessionsState = ref<LocalChatSession[]>([])
const activeSessionIdState = ref<string | null>(null)
let inMemoryDraftConfig: LocalChatDraftConfig | null = null
let hydrated = false
let pendingPersistence: ReturnType<typeof setTimeout> | undefined

function flushPendingChatUpdates() {
  clearTimeout(pendingPersistence)
  pendingPersistence = undefined
  persistSessions()
}

function scheduleMessagePersistence() {
  if (pendingPersistence !== undefined) return
  pendingPersistence = setTimeout(flushPendingChatUpdates, 1000)
}

function hydrate() {
  if (hydrated) return
  sessionsState.value = safeRead<LocalChatSession[]>(CHAT_SESSIONS_KEY, []).map(sanitizeSessionForPersistence)
  activeSessionIdState.value = safeRead<string | null>(CHAT_ACTIVE_SESSION_KEY, null)
  // Remove drafts written by versions that stored credentials in localStorage.
  safeRemove(CHAT_DRAFT_CONFIG_KEY)
  hydrated = true
  // Remove secrets written by older versions as soon as the chat is opened.
  persistSessions()
}

function persistSessions() {
  const persistedSessions = sessionsState.value.map(sanitizeSessionForPersistence)
  safeWrite(CHAT_SESSIONS_KEY, persistedSessions)
  safeWrite(CHAT_ACTIVE_SESSION_KEY, activeSessionIdState.value)
}

function createId(prefix: string): string {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
}

export function useLocalChat() {
  hydrate()

  const sessions = computed(() =>
    [...sessionsState.value].sort((a, b) => (a.updatedAt < b.updatedAt ? 1 : -1))
  )

  const activeSession = computed(() => {
    if (!activeSessionIdState.value) return null
    return sessionsState.value.find((session) => session.id === activeSessionIdState.value) ?? null
  })

  function createSession(
    input?: Partial<Pick<LocalChatSession, 'title' | 'model' | 'reasoningEffort' | 'apiKey' | 'baseUrl' | 'apiKeyHint'>>
  ) {
    const now = new Date().toISOString()
    const session: LocalChatSession = {
      id: createId('chat'),
      title: input?.title?.trim() || 'New Chat',
      model: input?.model?.trim() || '',
      reasoningEffort: input?.reasoningEffort ?? 'medium',
      apiKey: input?.apiKey?.trim() || '',
      baseUrl: input?.baseUrl?.trim() || '',
      apiKeyHint: input?.apiKeyHint?.trim() || '',
      messages: [],
      createdAt: now,
      updatedAt: now,
    }
    sessionsState.value.unshift(session)
    activeSessionIdState.value = session.id
    persistSessions()
    return session
  }

  function setActiveSession(id: string | null) {
    activeSessionIdState.value = id
    persistSessions()
  }

  function updateSession(id: string, patch: Partial<Omit<LocalChatSession, 'id' | 'createdAt'>>) {
    const index = sessionsState.value.findIndex((session) => session.id === id)
    if (index === -1) return
    sessionsState.value[index] = {
      ...sessionsState.value[index],
      ...patch,
      updatedAt: new Date().toISOString(),
    }
    persistSessions()
  }

  function deleteSession(id: string) {
    const next = sessionsState.value.filter((session) => session.id !== id)
    sessionsState.value = next
    if (activeSessionIdState.value === id) {
      activeSessionIdState.value = next[0]?.id ?? null
    }
    persistSessions()
  }

  function appendMessage(
    sessionId: string,
    role: LocalChatMessage['role'],
    content: string,
    attachments: LocalChatAttachment[] = []
  ): LocalChatMessage | null {
    const session = sessionsState.value.find((item) => item.id === sessionId)
    if (!session) return null
    const message: LocalChatMessage = {
      id: createId(role),
      role,
      content,
      attachments: attachments.map((attachment) => ({ ...attachment })),
      createdAt: new Date().toISOString(),
    }
    session.messages.push(message)
    session.updatedAt = new Date().toISOString()
    if ((session.title === 'New Chat' || !session.title.trim()) && role === 'user') {
      session.title = summarizeTitle(content)
    }
    persistSessions()
    return message
  }

  function updateMessage(sessionId: string, messageId: string, content: string, response?: Partial<LocalChatResponse>) {
    const session = sessionsState.value.find((item) => item.id === sessionId)
    if (!session) return
    const message = session.messages.find((item) => item.id === messageId)
    if (!message) return
    message.content = content
    if (response) message.response = { status: 'waiting', ...message.response, ...response }
    session.updatedAt = new Date().toISOString()
    scheduleMessagePersistence()
  }

  function deleteMessage(sessionId: string, messageId: string) {
    const session = sessionsState.value.find((item) => item.id === sessionId)
    if (!session) return
    session.messages = session.messages.filter((message) => message.id !== messageId)
    session.updatedAt = new Date().toISOString()
    persistSessions()
  }

  function clearAllSessions() {
    sessionsState.value = []
    activeSessionIdState.value = null
    persistSessions()
  }

  function saveDraftConfig(config: LocalChatDraftConfig) {
    // The API key is needed only while navigating inside this page session.
    // Keep the handoff in memory so it is never persisted in browser storage.
    inMemoryDraftConfig = { ...config }
  }

  function consumeDraftConfig(): LocalChatDraftConfig | null {
    const draft = inMemoryDraftConfig
    inMemoryDraftConfig = null
    return draft
  }

  return {
    sessions,
    activeSessionId: activeSessionIdState,
    activeSession,
    createSession,
    setActiveSession,
    updateSession,
    deleteSession,
    appendMessage,
    updateMessage,
    deleteMessage,
    clearAllSessions,
    saveDraftConfig,
    consumeDraftConfig,
    flushPendingChatUpdates,
  }
}

function summarizeTitle(input: string): string {
  const normalized = input.replace(/\s+/g, ' ').trim()
  if (!normalized) return 'New Chat'
  return normalized.length > 36 ? `${normalized.slice(0, 36)}...` : normalized
}
