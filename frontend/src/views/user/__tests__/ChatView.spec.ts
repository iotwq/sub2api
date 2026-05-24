import { computed, defineComponent, ref } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import mammoth from 'mammoth'
import { getDocument } from 'pdfjs-dist/build/pdf.mjs'
import ChatView from '@/views/user/ChatView.vue'

type MockAttachment = {
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

type MockMessage = {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  attachments?: MockAttachment[]
  createdAt: string
}

type MockSession = {
  id: string
  title: string
  model: string
  deepThinkingEnabled?: boolean
  apiKey?: string
  baseUrl?: string
  apiKeyHint: string
  messages: MockMessage[]
  createdAt: string
  updatedAt: string
}

const routeRef = ref({ fullPath: '/chat' })
const sessionsState = ref<MockSession[]>([])
const activeSessionIdState = ref<string | null>(null)
const draftConfigState = ref<{
  apiKey: string
  baseUrl: string
  model?: string
  sessionTitle?: string
  deepThinkingEnabled?: boolean
} | null>(null)

let idCounter = 0

const appStoreMock = {
  fetchPublicSettings: vi.fn().mockResolvedValue(null),
  cachedPublicSettings: {
    api_base_url: 'http://localhost:8080/v1',
  },
  showSuccess: vi.fn(),
  showError: vi.fn(),
}

const storageState = new Map<string, string>()
const storageMock = {
  getItem: vi.fn((key: string) => storageState.get(key) ?? null),
  setItem: vi.fn((key: string, value: string) => {
    storageState.set(key, String(value))
  }),
  removeItem: vi.fn((key: string) => {
    storageState.delete(key)
  }),
  clear: vi.fn(() => {
    storageState.clear()
  }),
}

const translations: Record<string, string> = {
  'common.delete': '删除',
  'chat.title': '聊天',
  'chat.description': '使用自己的 API Key 直接开始聊天。',
  'chat.sidebarDescription': '会话只保存在当前浏览器本地。',
  'chat.defaultSessionTitle': '新对话',
  'chat.baseUrl': 'API 地址',
  'chat.baseUrlPlaceholder': 'https://your-domain.com/v1',
  'chat.apiKey': 'API Key',
  'chat.apiKeyPlaceholder': '粘贴 key',
  'chat.model': '模型',
  'chat.modelPlaceholder': '选择模型',
  'chat.deepThinking': '深度思考',
  'chat.deepThinkingHint': '默认使用 medium，开启后切换为 xhigh。',
  'chat.deepThinkingStateOn': 'ON',
  'chat.deepThinkingStateOff': 'OFF',
  'chat.localOnlyTitle': '仅保存在本地',
  'chat.localOnlyDescription': '聊天记录只保存在浏览器。',
  'chat.newChat': '新建对话',
  'chat.noSessions': '还没有聊天会话。',
  'chat.noModel': '未选择模型',
  'chat.emptyTitle': '当前没有激活的会话',
  'chat.emptyDescription': '新建一个对话，加载模型后就可以开始聊天。',
  'chat.startTitle': '开始本地聊天',
  'chat.startDescription': '配置好 API 地址、Key 和模型后即可开始聊天。',
  'chat.currentModel': '当前模型：{model}',
  'chat.clearMessages': '清空消息',
  'chat.clearAll': '清空全部',
  'chat.you': '你',
  'chat.assistant': '助手',
  'chat.streaming': '助手正在回复...',
  'chat.inputPlaceholder': '发送消息。Enter 发送，Shift+Enter 换行。',
  'chat.dropzoneTitle': '松开即可添加附件',
  'chat.dropzoneDescription': '支持拖拽图片、PDF、DOCX 和文本文件到这里。',
  'chat.uploadAttachment': '上传附件',
  'chat.uploadImage': '上传图片',
  'chat.uploadFile': '上传文件',
  'chat.removeAttachment': '移除附件',
  'chat.processingAttachments': '正在处理附件...',
  'chat.processingAttachmentWithName': '正在处理附件：{name}',
  'chat.supportedFileHint': '支持图片、PDF、DOCX 和常见文本文件。',
  'chat.attachmentsSummary': '已附加 {count} 个附件：{name}',
  'chat.pdfPageCount': '{count} 页',
  'chat.wordAttachmentLabel': 'Word 文档',
  'chat.imageAttachmentLabel': '图片',
  'chat.textAttachmentLabel': '文本附件',
  'chat.stop': '停止',
  'chat.send': '发送',
  'chat.sending': '发送中...',
  'chat.copyMessage': '复制回复',
  'chat.copyCode': '复制代码',
  'chat.copied': '已复制',
  'chat.messageCopied': '回复已复制到剪贴板',
  'chat.codeCopied': '代码已复制到剪贴板',
  'chat.loadingModels': '加载中...',
  'chat.refreshModels': '同步模型',
  'chat.errors.loadModelsFailed': '加载模型失败。',
  'chat.errors.sendFailed': '发送消息失败。',
  'chat.errors.emptyAssistantReply': '没有收到有效回复。',
  'chat.errors.unsupportedFileType': '当前只支持图片和文本类文件。',
  'chat.errors.imageTooLarge': '单张图片不能超过 10 MB。',
  'chat.errors.fileTooLarge': '单个文本文件不能超过 10 MB。',
  'chat.errors.pdfTooLarge': '单个 PDF 不能超过 10 MB。',
  'chat.errors.docxTooLarge': '单个 DOCX 不能超过 10 MB。',
  'chat.errors.tooManyAttachments': '单条消息最多上传 {count} 个附件。',
  'chat.errors.attachmentReadFailed': '读取附件失败。',
  'chat.errors.pdfNoExtractableText': '这个 PDF 没有提取到可发送的文本，可能是扫描件或图片版 PDF。',
  'chat.errors.docxNoExtractableText': '这个 Word 文档没有提取到可发送的文本内容。',
  'chat.errors.pdfReadFailed': 'PDF 处理失败，请确认文件没有损坏，或换一个更小的 PDF 再试。',
  'chat.errors.docxReadFailed': 'Word 文档处理失败，请确认文件没有损坏，并且格式为 .docx。',
  'chat.errors.pdfReadFailedWithReason': 'PDF 处理失败：{reason}',
  'chat.errors.docxReadFailedWithReason': 'Word 文档处理失败：{reason}',
  'chat.errors.unknownReason': '未知原因',
}

function t(key: string, params?: Record<string, unknown>) {
  const template = translations[key] ?? key
  return template.replace(/\{(\w+)\}/g, (_match, name: string) => String(params?.[name] ?? ''))
}

function resetLocalChatMock() {
  sessionsState.value = []
  activeSessionIdState.value = null
  draftConfigState.value = null
  idCounter = 0
}

function nextId(prefix: string) {
  idCounter += 1
  return `${prefix}_${idCounter}`
}

function nowIso() {
  return new Date('2026-05-08T00:00:00.000Z').toISOString()
}

function createSessionRecord(
  input?: Partial<Pick<MockSession, 'title' | 'model' | 'deepThinkingEnabled' | 'apiKey' | 'baseUrl' | 'apiKeyHint' | 'messages'>>
): MockSession {
  return {
    id: nextId('chat'),
    title: input?.title?.trim() || '新对话',
    model: input?.model?.trim() || '',
    deepThinkingEnabled: input?.deepThinkingEnabled ?? false,
    apiKey: input?.apiKey?.trim() || '',
    baseUrl: input?.baseUrl?.trim() || '',
    apiKeyHint: input?.apiKeyHint?.trim() || '',
    messages: input?.messages?.map((message) => ({
      ...message,
      attachments: message.attachments?.map((attachment) => ({ ...attachment })) || [],
    })) || [],
    createdAt: nowIso(),
    updatedAt: nowIso(),
  }
}

function seedActiveSession(
  input?: Partial<Pick<MockSession, 'title' | 'model' | 'deepThinkingEnabled' | 'apiKey' | 'baseUrl' | 'apiKeyHint' | 'messages'>>
) {
  const session = createSessionRecord(input)
  sessionsState.value = [session, ...sessionsState.value]
  activeSessionIdState.value = session.id
  return session
}

vi.mock('vue-router', () => ({
  useRouter: () => ({
    currentRoute: routeRef,
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t,
    }),
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreMock,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStoreMock,
}))

vi.mock('@/i18n', () => ({
  i18n: {
    global: {
      t,
    },
  },
}))

vi.mock('@/composables/useLocalChat', () => ({
  useLocalChat: () => ({
    sessions: computed(() => sessionsState.value),
    activeSessionId: activeSessionIdState,
    activeSession: computed(
      () => sessionsState.value.find((session) => session.id === activeSessionIdState.value) ?? null
    ),
    createSession: (
      input?: Partial<Pick<MockSession, 'title' | 'model' | 'deepThinkingEnabled' | 'apiKey' | 'baseUrl' | 'apiKeyHint'>>
    ) => {
      return seedActiveSession(input)
    },
    setActiveSession: (id: string | null) => {
      activeSessionIdState.value = id
    },
    updateSession: (id: string, patch: Partial<Omit<MockSession, 'id' | 'createdAt'>>) => {
      sessionsState.value = sessionsState.value.map((session) =>
        session.id === id
          ? {
              ...session,
              ...patch,
              updatedAt: nowIso(),
            }
          : session
      )
    },
    deleteSession: (id: string) => {
      sessionsState.value = sessionsState.value.filter((session) => session.id !== id)
      if (activeSessionIdState.value === id) {
        activeSessionIdState.value = sessionsState.value[0]?.id ?? null
      }
    },
    appendMessage: (
      sessionId: string,
      role: MockMessage['role'],
      content: string,
      attachments: MockAttachment[] = []
    ) => {
      const session = sessionsState.value.find((item) => item.id === sessionId)
      if (!session) return null
      const message: MockMessage = {
        id: nextId(role),
        role,
        content,
        attachments: attachments.map((attachment) => ({ ...attachment })),
        createdAt: nowIso(),
      }
      session.messages.push(message)
      session.updatedAt = nowIso()
      return message
    },
    updateMessage: (sessionId: string, messageId: string, content: string) => {
      const session = sessionsState.value.find((item) => item.id === sessionId)
      const message = session?.messages.find((item) => item.id === messageId)
      if (!message || !session) return
      message.content = content
      session.updatedAt = nowIso()
    },
    clearAllSessions: () => {
      sessionsState.value = []
      activeSessionIdState.value = null
    },
    saveDraftConfig: (config: typeof draftConfigState.value extends infer T ? Exclude<T, null> : never) => {
      draftConfigState.value = config
    },
    consumeDraftConfig: () => {
      const draft = draftConfigState.value
      draftConfigState.value = null
      return draft
    },
  }),
}))

vi.mock('mammoth', () => ({
  default: {
    extractRawText: vi.fn(),
  },
}))

vi.mock('pdfjs-dist/build/pdf.mjs', () => ({
  getDocument: vi.fn(),
}))

vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: defineComponent({
    template: '<div><slot /></div>',
  }),
}))

vi.mock('@/components/icons/Icon.vue', () => ({
  default: defineComponent({
    props: {
      name: { type: String, required: false, default: '' },
    },
    template: '<span :data-icon="name"></span>',
  }),
}))

vi.mock('@/components/common/Select.vue', () => ({
  default: defineComponent({
    props: {
      modelValue: { type: [String, Number, Boolean, null], required: false, default: null },
      options: { type: Array, required: false, default: () => [] },
    },
    emits: ['update:modelValue'],
    template: `
      <select
        :value="modelValue ?? ''"
        @change="$emit('update:modelValue', $event.target.value)"
      >
        <option
          v-for="option in options"
          :key="option.value"
          :value="option.value"
        >
          {{ option.label }}
        </option>
      </select>
    `,
  }),
}))

function createPdfMock(pageCount: number) {
  return {
    promise: Promise.resolve({
      numPages: pageCount,
      getPage: vi.fn(async (pageNumber: number) => ({
        getTextContent: vi.fn(async () => ({
          items: [{ str: `Page ${pageNumber} content` }],
        })),
      })),
    }),
  }
}

function createStreamResponse(lines: string[]) {
  const encoder = new TextEncoder()
  const chunks = lines.map((line) => encoder.encode(line))
  let index = 0

  return {
    ok: true,
    body: {
      getReader: () => ({
        read: vi.fn().mockImplementation(async () => {
          if (index < chunks.length) {
            return { done: false, value: chunks[index++] }
          }
          return { done: true, value: undefined }
        }),
      }),
    },
  } as Response
}

async function mountChatView() {
  const wrapper = mount(ChatView, {
  })

  await flushPromises()
  return wrapper
}

async function startSessionFromUi(wrapper: ReturnType<typeof mount>) {
  const newChatButtons = wrapper.findAll('button').filter((button) => button.text().includes('新建对话'))
  if (!newChatButtons.length) {
    throw new Error('Expected a new chat button')
  }

  await newChatButtons[0].trigger('click')
  await flushPromises()
}

async function attachFiles(
  wrapper: ReturnType<typeof mount>,
  files: File[],
  inputIndex: number
) {
  const input = wrapper.findAll('input[type="file"]')[inputIndex]
  Object.defineProperty(input.element, 'files', {
    configurable: true,
    value: files,
  })
  await input.trigger('change')
  await flushPromises()
}

function createMockFile(name: string, content: string, type: string): File {
  const file = new File([content], name, { type }) as File & {
    text: () => Promise<string>
    arrayBuffer: () => Promise<ArrayBuffer>
  }

  file.text = vi.fn(async () => content)
  file.arrayBuffer = vi.fn(async () => new TextEncoder().encode(content).buffer)

  return file
}

describe('ChatView 附件体验', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'localStorage', {
      configurable: true,
      value: storageMock,
    })
    Object.defineProperty(window, 'isSecureContext', {
      configurable: true,
      value: true,
    })
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
    })
    resetLocalChatMock()
    storageMock.clear()
    vi.clearAllMocks()
    routeRef.value = { fullPath: '/chat' }
    appStoreMock.fetchPublicSettings.mockReset()
    appStoreMock.fetchPublicSettings.mockResolvedValue(null)
    appStoreMock.showSuccess.mockReset()
    appStoreMock.showError.mockReset()
    vi.mocked(getDocument).mockReset()
    vi.mocked(mammoth.extractRawText).mockReset()
  })

  it('支持拖拽上传，并在 drop 后隐藏拖拽提示', async () => {
    seedActiveSession()
    const wrapper = await mountChatView()
    const form = wrapper.find('form')
    const file = createMockFile('dropped.md', '# hello', 'text/markdown')

    await form.trigger('dragenter', {
      dataTransfer: {
        types: ['Files'],
      },
    })
    expect(wrapper.text()).toContain('松开即可添加附件')

    await form.trigger('drop', {
      dataTransfer: {
        types: ['Files'],
        files: [file],
      },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('dropped.md')
    expect(wrapper.text()).toContain('文本附件')
    expect(wrapper.text()).not.toContain('松开即可添加附件')
  })

  it('PDF 附件会显示页数提示', async () => {
    vi.mocked(getDocument).mockReturnValue(createPdfMock(3) as never)

    seedActiveSession()
    const wrapper = await mountChatView()
    const pdfFile = createMockFile('report.pdf', 'fake pdf', 'application/pdf')

    await attachFiles(wrapper, [pdfFile], 0)

    expect(wrapper.text()).toContain('report.pdf')
    expect(wrapper.text()).toContain('3 页')
  })

  it('DOCX 附件会显示 Word 文档标签', async () => {
    vi.mocked(mammoth.extractRawText).mockResolvedValue({
      value: 'Word body',
      messages: [],
    } as never)

    seedActiveSession()
    const wrapper = await mountChatView()
    const docxFile = createMockFile(
      'meeting.docx',
      'fake docx',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    )

    await attachFiles(wrapper, [docxFile], 0)

    expect(wrapper.text()).toContain('meeting.docx')
    expect(wrapper.text()).toContain('Word 文档')
  })

  it('只保留一个上传入口', async () => {
    const wrapper = await mountChatView()

    expect(wrapper.findAll('input[type="file"]')).toHaveLength(1)
    expect(wrapper.text()).toContain('上传附件')
  })

  it('会收敛重复提示，只保留必要信息', async () => {
    const wrapper = await mountChatView()
    const text = wrapper.text()

    expect(text).not.toContain('仅保存在本地')
    expect(text).not.toContain('这里仅支持 OpenAI-compatible 聊天体验')
    expect(text).not.toContain('图片与文件前端限制')
    expect(text).not.toContain('当前模型：')
  })

  it('会渲染代码块复制按钮，并允许一键复制回复', async () => {
    seedActiveSession()
    const wrapper = await mountChatView()
    const session = sessionsState.value[0]

    if (!session) {
      throw new Error('Expected an active chat session')
    }

    session.messages.push({
      id: nextId('assistant'),
      role: 'assistant',
      content: '```ts\nconst answer = 42\n```',
      createdAt: nowIso(),
    })

    await flushPromises()

    const codeCopyButton = wrapper.find('.chat-code-copy')
    expect(codeCopyButton.text()).toBe('复制代码')

    await codeCopyButton.trigger('click')
    await flushPromises()

    expect(codeCopyButton.text()).toBe('已复制')
    expect(appStoreMock.showSuccess).toHaveBeenCalledWith('代码已复制到剪贴板')
  })

  it('默认发送 medium，开启深度思考后发送 xhigh', async () => {
    const requestBodies: Array<Record<string, any>> = []
    global.fetch = vi.fn().mockImplementation(async (_input: RequestInfo | URL, init?: RequestInit) => {
      requestBodies.push(JSON.parse(String(init?.body ?? '{}')))
      return createStreamResponse([
        'data: {"choices":[{"delta":{"content":"已收到"}}]}\n\n',
        'data: [DONE]\n\n',
      ])
    }) as any

    seedActiveSession()
    const wrapper = await mountChatView()
    const apiKeyInput = wrapper.find('input[type="password"]')
    const composer = wrapper.find('textarea')

    await apiKeyInput.setValue('sk-test-12345678')
    await composer.setValue('first question')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(requestBodies[0]?.reasoning?.effort).toBe('medium')

    await wrapper.get('[data-testid="chat-deep-thinking-toggle"]').trigger('click')
    await composer.setValue('second question')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(requestBodies[1]?.reasoning?.effort).toBe('xhigh')
  })

  it('附件批量处理时会保留成功项，并为 PDF 提取失败给出明确提示', async () => {
    const failingPdfPromise = Promise.resolve().then(() => {
      throw new Error('Invalid PDF structure')
    })
    failingPdfPromise.catch(() => undefined)

    vi.mocked(getDocument).mockReturnValue({
      promise: failingPdfPromise,
    } as never)

    seedActiveSession()
    const wrapper = await mountChatView()
    const textFile = createMockFile('notes.txt', 'plain text', 'text/plain')
    const brokenPdf = createMockFile('broken.pdf', 'broken pdf', 'application/pdf')

    await attachFiles(wrapper, [textFile, brokenPdf], 0)

    expect(wrapper.text()).toContain('notes.txt')
    expect(wrapper.text()).toContain('broken.pdf')
    expect(wrapper.text()).toContain('PDF 处理失败：Invalid PDF structure')
  })

  it('进入聊天页时不会自动新建对话，并会复用已有会话', async () => {
    const emptyWrapper = await mountChatView()

    expect(sessionsState.value).toHaveLength(0)
    expect(activeSessionIdState.value).toBeNull()
    expect(emptyWrapper.text()).toContain('开始本地聊天')

    emptyWrapper.unmount()
    resetLocalChatMock()

    const existingSession = createSessionRecord({
      title: '保留的对话',
      model: 'gpt-5.4',
    })
    sessionsState.value = [existingSession]

    const wrapper = await mountChatView()

    expect(sessionsState.value).toHaveLength(1)
    expect(activeSessionIdState.value).toBe(existingSession.id)
    expect(wrapper.text()).toContain('保留的对话')
  })

  it('存在草稿配置时只预填设置，不会自动新建对话', async () => {
    draftConfigState.value = {
      apiKey: 'sk-draft-12345678',
      baseUrl: 'https://example.com/v1',
      model: 'gpt-5.4-mini',
      sessionTitle: '来自草稿',
      deepThinkingEnabled: true,
    }

    const wrapper = await mountChatView()

    expect(sessionsState.value).toHaveLength(0)
    expect(activeSessionIdState.value).toBeNull()
    expect((wrapper.find('input[type="password"]').element as HTMLInputElement).value).toBe('sk-draft-12345678')
    expect(wrapper.text()).toContain('开始本地聊天')
  })

  it('流式回复期间主按钮切换为停止，并且可以中断回复', async () => {
    seedActiveSession()

    global.fetch = vi.fn().mockImplementation(async (_input: RequestInfo | URL, init?: RequestInit) => {
      const signal = init?.signal as AbortSignal | undefined
      const abortError = Object.assign(new Error('Aborted'), { name: 'AbortError' })

      return {
        ok: true,
        body: {
          getReader: () => ({
            read: vi.fn().mockImplementation(() => new Promise((resolve, reject) => {
              if (signal?.aborted) {
                reject(abortError)
                return
              }

              signal?.addEventListener('abort', () => reject(abortError), { once: true })
            })),
          }),
        },
      } as Response
    }) as any

    const wrapper = await mountChatView()
    const apiKeyInput = wrapper.find('input[type="password"]')
    const composer = wrapper.find('textarea')

    await apiKeyInput.setValue('sk-test-12345678')
    await composer.setValue('请持续回复')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const primaryButton = wrapper.find('button[type="submit"]')
    expect(primaryButton.text()).toContain('停止')
    expect(composer.attributes('disabled')).toBeUndefined()

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('button[type="submit"]').text()).toContain('发送')
  })

  it('显式点击新建对话时才会创建会话', async () => {
    const wrapper = await mountChatView()

    expect(sessionsState.value).toHaveLength(0)

    await startSessionFromUi(wrapper)

    expect(sessionsState.value).toHaveLength(1)
    expect(activeSessionIdState.value).toBeTruthy()
  })
})
