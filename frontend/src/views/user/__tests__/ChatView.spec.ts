import { computed, defineComponent, ref } from 'vue'
import { mount, flushPromises, enableAutoUnmount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import mammoth from 'mammoth'
import { getDocument } from 'pdfjs-dist/build/pdf.mjs'
import ChatView from '@/views/user/ChatView.vue'
import { saveAs } from 'file-saver'
import { marked } from 'marked'
import type { ChatReasoningEffort, LocalChatResponse } from '@/composables/useLocalChat'

enableAutoUnmount(afterEach)
afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})

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
  response?: LocalChatResponse
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
  reasoningEffort?: ChatReasoningEffort
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
  reasoningEffort?: ChatReasoningEffort
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
  'chat.connectionSettings': '连接设置',
  'chat.showHistory': '会话列表',
  'chat.hideHistory': '收起会话',
  'chat.focusMode': '专注模式',
  'chat.exitFocus': '退出专注',
  'chat.defaultSessionTitle': '新对话',
  'chat.baseUrl': 'API 地址',
  'chat.baseUrlPlaceholder': 'https://your-domain.com/v1',
  'chat.apiKey': 'API Key',
  'chat.apiKeyPlaceholder': '粘贴 key',
  'chat.model': '模型',
  'chat.modelPlaceholder': '选择模型',
  'chat.reasoningEffort': '推理强度',
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
  'chat.scrollToLatest': '回到最新消息',
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
  'chat.editMessage': '编辑问题',
  'chat.regenerate': '重新生成',
  'chat.retry': '重试',
  'chat.branch': '分支',
  'chat.copyCode': '复制代码',
  'chat.codePreview': '展开代码',
  'chat.downloadCode': '下载代码',
  'chat.sources': '参考来源',
  'chat.copied': '已复制',
  'chat.messageCopied': '回复已复制到剪贴板',
  'chat.codeCopied': '代码已复制到剪贴板',
  'chat.loadingModels': '加载中...',
  'chat.refreshModels': '同步模型',
  'chat.errors.loadModelsFailed': '加载模型失败。',
  'chat.errors.apiKeyRequired': '请先输入 API Key。',
  'chat.errors.baseUrlRequired': '请先输入 API 地址。',
  'chat.errors.noModels': '当前 API 地址和密钥未返回可用模型。',
  'chat.errors.sendFailed': '发送消息失败。',
  'chat.errors.emptyAssistantReply': '没有收到有效回复。',
  'chat.errors.networkFailed': '连接已中断，未收到完整回复。请检查网络、反向代理超时或稍后重试。',
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
  input?: Partial<Pick<MockSession, 'title' | 'model' | 'reasoningEffort' | 'apiKey' | 'baseUrl' | 'apiKeyHint' | 'messages'>>
): MockSession {
  return {
    id: nextId('chat'),
    title: input?.title?.trim() || '新对话',
    model: input?.model?.trim() || '',
    reasoningEffort: input?.reasoningEffort,
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
  input?: Partial<Pick<MockSession, 'title' | 'model' | 'reasoningEffort' | 'apiKey' | 'baseUrl' | 'apiKeyHint' | 'messages'>>
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
      locale: ref('zh'),
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
      input?: Partial<Pick<MockSession, 'title' | 'model' | 'reasoningEffort' | 'apiKey' | 'baseUrl' | 'apiKeyHint'>>
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
    updateMessage: (sessionId: string, messageId: string, content: string, response?: Partial<LocalChatResponse>) => {
      const session = sessionsState.value.find((item) => item.id === sessionId)
      const message = session?.messages.find((item) => item.id === messageId)
      if (!message || !session) return
      message.content = content
      if (response) message.response = { status: 'waiting', ...message.response, ...response }
      session.updatedAt = nowIso()
    },
    deleteMessage: (sessionId: string, messageId: string) => {
      const session = sessionsState.value.find((item) => item.id === sessionId)
      if (!session) return
      session.messages = session.messages.filter((message) => message.id !== messageId)
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
    flushPendingChatUpdates: vi.fn(),
  }),
}))

vi.mock('mammoth', () => ({
  default: {
    extractRawText: vi.fn(),
  },
}))

vi.mock('file-saver', () => ({ saveAs: vi.fn() }))

vi.mock('pdfjs-dist/build/pdf.mjs', () => ({
  GlobalWorkerOptions: {
    workerSrc: '',
  },
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
      disabled: { type: Boolean, required: false, default: false },
    },
    emits: ['update:modelValue'],
    template: `
      <select
        :value="modelValue ?? ''"
        :disabled="disabled"
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
    headers: new Headers({ 'content-type': 'text/event-stream' }),
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

function createChatStreamResponse() {
  return createStreamResponse([
    'data: {"choices":[{"delta":{"content":"已收到"},"finish_reason":null}]}\n\n',
    'data: [DONE]\n\n',
  ])
}

async function mountChatView() {
  const wrapper = mount(ChatView, {
    attachTo: document.body,
  })

  await flushPromises()
  return wrapper
}

const MODEL_IDS = [
  'gpt-6-astra', 'gpt-6-sol', 'gpt-5.6-sol', 'claude-opus-5-5',
  'claude-sonnet-5-5', 'grok-4.7', 'gotk-4.6', 'gpt-5.5',
]

function createModelsResponse(ids = MODEL_IDS) {
  return { ok: true, json: async () => ({ data: ids.map((id) => ({ id })) }) } as Response
}

async function syncModels(wrapper: ReturnType<typeof mount>) {
  const button = wrapper.findAll('button').find((button) => button.text() === '同步模型')
  if (!button) throw new Error('Expected a model sync button')
  await button.trigger('click')
  await flushPromises()
}

async function configureAndSyncModels(wrapper: ReturnType<typeof mount>) {
  await wrapper.find('input[type="password"]').setValue('sk-test-12345678')
  await syncModels(wrapper)
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
    global.fetch = vi.fn().mockImplementation(async (input: RequestInfo | URL) => {
      if (String(input).endsWith('/models')) return createModelsResponse()
      if (String(input).endsWith('/chat/completions')) return createChatStreamResponse()
      return createStreamResponse([
        'data: {"type":"response.output_text.delta","delta":"已收到"}\n\n',
        'data: {"type":"response.completed"}\n\n',
      ])
    })
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

  it('配置可收起，历史可折叠，专注模式可用 Escape 退出', async () => {
    seedActiveSession({ apiKey: 'sk-saved-key' })
    const wrapper = await mountChatView()
    expect(wrapper.get('[data-testid="chat-connection-settings"]').isVisible()).toBe(false)
    await wrapper.findAll('button').find((button) => button.text() === '连接设置')!.trigger('click')
    expect(wrapper.get('[data-testid="chat-connection-settings"]').isVisible()).toBe(true)
    await wrapper.findAll('button').find((button) => button.text() === '收起会话')!.trigger('click')
    expect(wrapper.get('[data-testid="chat-history"]').isVisible()).toBe(false)
    await wrapper.findAll('button').find((button) => button.text() === '专注模式')!.trigger('click')
    expect(wrapper.get('.chat-page').classes()).toContain('chat-page-focus')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await flushPromises()
    expect(wrapper.get('.chat-page').classes()).not.toContain('chat-page-focus')
  })

  it('会收敛重复提示，只保留必要信息', async () => {
    const wrapper = await mountChatView()
    const text = wrapper.text()

    expect(text).not.toContain('仅保存在本地')
    expect(text).not.toContain('这里仅支持 OpenAI-compatible 聊天体验')
    expect(text).not.toContain('图片与文件前端限制')
    expect(text).not.toContain('当前模型：')
  })

  it('输入框随内容增高、缩回，超出上限滚动并在窗口调整后重新计算', async () => {
    const wrapper = await mountChatView()
    const composer = wrapper.get('textarea')
    const textarea = composer.element as HTMLTextAreaElement
    let contentHeight = 40
    Object.defineProperty(textarea, 'scrollHeight', { get: () => contentHeight })

    await composer.setValue('短问题')
    expect(textarea.style.height).toBe('40px')
    contentHeight = 112
    await composer.setValue('第一行\n第二行\n第三行\n第四行')
    expect(textarea.style.height).toBe('112px')
    expect(textarea.style.overflowY).toBe('hidden')
    contentHeight = 600
    await composer.setValue('多行长问题')
    expect(textarea.style.height).toBe(`${Math.min(200, window.innerHeight * 0.25)}px`)
    expect(textarea.style.overflowY).toBe('auto')
    contentHeight = 88
    window.dispatchEvent(new Event('resize'))
    expect(textarea.style.height).toBe('88px')
    contentHeight = 40
    await composer.setValue('')
    expect(textarea.style.height).toBe('40px')
    expect(textarea.style.overflowY).toBe('hidden')
  })

  it('使用 API 地址和密钥同步全部模型，不过滤 GPT、Claude 或 Grok 模型名', async () => {
    const wrapper = await mountChatView()
    expect(wrapper.get('[data-testid="chat-model-select"]').findAll('option')).toHaveLength(0)
    await wrapper.find('input[type="text"]').setValue('https://gateway.example/')
    await configureAndSyncModels(wrapper)
    const select = wrapper.get('[data-testid="chat-model-select"]')
    const optionValues = Array.from((select.element as HTMLSelectElement).options).map((option) => option.value)

    expect(optionValues).toEqual(MODEL_IDS)
    expect((select.element as HTMLSelectElement).value).toBe('gpt-6-astra')
    expect(global.fetch).toHaveBeenCalledWith('https://gateway.example/v1/models', {
      method: 'GET',
      headers: { Authorization: 'Bearer sk-test-12345678', Accept: 'application/json' },
    })
    expect(wrapper.text()).not.toContain('深度思考')
    expect(wrapper.text()).not.toContain('联网搜索')
  })

  it('中文输入法仍在组字时按 Enter 不会误发送', async () => {
    seedActiveSession()
    const wrapper = await mountChatView()
    const composer = wrapper.find('textarea')

    await configureAndSyncModels(wrapper)
    vi.mocked(global.fetch).mockClear()
    await composer.setValue('还在输入')
    await composer.trigger('keydown', { key: 'Enter', isComposing: true })
    await flushPromises()

    expect(global.fetch).not.toHaveBeenCalled()
    expect(sessionsState.value[0]?.messages).toHaveLength(0)
    expect((composer.element as HTMLTextAreaElement).value).toBe('还在输入')
  })

  it('同步要求地址和密钥齐全，未同步模型时不能发送', async () => {
    seedActiveSession()
    const wrapper = await mountChatView()
    await wrapper.find('textarea').setValue('hello')
    await syncModels(wrapper)
    expect(wrapper.text()).toContain('请先输入 API Key')
    await wrapper.find('input[type="text"]').setValue('')
    await syncModels(wrapper)
    expect(wrapper.text()).toContain('请先输入 API 地址')
    await wrapper.find('form').trigger('submit.prevent')
    expect(global.fetch).not.toHaveBeenCalled()
    expect(sessionsState.value[0]?.messages).toHaveLength(0)
  })

  it('重新同步保留可用的选择，模型被移除后选择新列表首项', async () => {
    const wrapper = await mountChatView()
    await configureAndSyncModels(wrapper)
    const select = wrapper.get('[data-testid="chat-model-select"]')
    await select.setValue('claude-sonnet-5-5')
    await syncModels(wrapper)
    expect((select.element as HTMLSelectElement).value).toBe('claude-sonnet-5-5')
    vi.mocked(global.fetch).mockResolvedValueOnce(createModelsResponse(['grok-4.7']))
    await syncModels(wrapper)
    expect((select.element as HTMLSelectElement).value).toBe('grok-4.7')
  })

  it.each(['empty', 'failure'])('同步 %s 时不恢复固定模型，禁用发送并提示原因', async (result) => {
    seedActiveSession()
    const wrapper = await mountChatView()
    await configureAndSyncModels(wrapper)
    if (result === 'empty') {
      vi.mocked(global.fetch).mockResolvedValueOnce(createModelsResponse([]))
    } else {
      vi.mocked(global.fetch).mockRejectedValueOnce(new Error('Failed to load models (401)'))
    }
    await syncModels(wrapper)
    await wrapper.find('textarea').setValue('hello')
    expect(wrapper.get('[data-testid="chat-model-select"]').findAll('option')).toHaveLength(0)
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain(result === 'empty' ? '未返回可用模型' : 'Failed to load models (401)')
    vi.mocked(global.fetch).mockClear()
    await wrapper.find('form').trigger('submit.prevent')
    expect(global.fetch).not.toHaveBeenCalled()
  })

  it.each(['text', 'password'])('更换 %s 配置后清空旧模型列表，需重新同步', async (inputType) => {
    seedActiveSession()
    const wrapper = await mountChatView()
    await configureAndSyncModels(wrapper)
    await wrapper.find('textarea').setValue('hello')
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeUndefined()
    await wrapper.find(`input[type="${inputType}"]`).setValue(
      inputType === 'text' ? 'https://new.example/v1' : 'sk-new-key'
    )
    expect(wrapper.get('[data-testid="chat-model-select"]').findAll('option')).toHaveLength(0)
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeDefined()
    await syncModels(wrapper)
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })

  it.each(['success', 'failure'])('更换密钥后忽略旧同步请求的迟到 %s', async (result) => {
    const wrapper = await mountChatView()
    let finish!: (response: Response) => void
    let fail!: (error: Error) => void
    vi.mocked(global.fetch).mockImplementationOnce(() => new Promise<Response>((resolve, reject) => {
      finish = resolve
      fail = reject
    }))
    await wrapper.find('input[type="password"]').setValue('sk-old-key')
    await syncModels(wrapper)
    await wrapper.find('input[type="password"]').setValue('sk-new-key')
    if (result === 'success') finish(createModelsResponse(['old-model']))
    else fail(new Error('old key failed'))
    await flushPromises()
    expect(wrapper.get('[data-testid="chat-model-select"]').findAll('option')).toHaveLength(0)
    expect(wrapper.text()).not.toContain('old key failed')
    await syncModels(wrapper)
    expect(wrapper.get('[data-testid="chat-model-select"]').findAll('option')).toHaveLength(MODEL_IDS.length)
    expect(vi.mocked(global.fetch).mock.calls.at(-1)?.[1]?.headers)
      .toMatchObject({ Authorization: 'Bearer sk-new-key' })
  })

  it('切换会话恢复各自模型和推理强度，旧会话默认 medium', async () => {
    const legacy = seedActiveSession({ title: '旧会话', model: 'gpt-5.5', apiKey: 'sk-saved-key' })
    const recent = seedActiveSession({ title: '新会话', model: 'grok-4.7', apiKey: 'sk-saved-key', reasoningEffort: 'low' })
    const wrapper = await mountChatView()
    const effortSelect = wrapper.get('[data-testid="chat-reasoning-effort-select"]')
    expect((effortSelect.element as HTMLSelectElement).value).toBe('low')
    await effortSelect.setValue('max')
    await wrapper.findAll('[role="button"]').find((button) => button.text().includes('旧会话'))!.trigger('click')
    await flushPromises()
    expect(activeSessionIdState.value).toBe(legacy.id)
    expect((effortSelect.element as HTMLSelectElement).value).toBe('medium')
    expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('gpt-5.5')
    await wrapper.findAll('[role="button"]').find((button) => button.text().includes('新会话'))!.trigger('click')
    await flushPromises()
    expect(activeSessionIdState.value).toBe(recent.id)
    expect((effortSelect.element as HTMLSelectElement).value).toBe('max')
    expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('grok-4.7')
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

    const codeCopyButton = wrapper.find('[data-copy-type="code"]')
    expect(codeCopyButton.text()).toBe('复制代码')

    await codeCopyButton.trigger('click')
    await flushPromises()

    expect(codeCopyButton.text()).toBe('已复制')
    expect(appStoreMock.showSuccess).toHaveBeenCalledWith('代码已复制到剪贴板')
  })

  it('代码高亮、展开和下载保持原始文本，不执行模型提供的 HTML', async () => {
    seedActiveSession({ messages: [{
      id: 'code', role: 'assistant', content: '```js\nconst value = "<script>alert(1)</script>"\n```', createdAt: nowIso(),
    }] })
    const wrapper = await mountChatView()
    expect(wrapper.find('.hljs-keyword').exists()).toBe(true)
    expect(wrapper.find('script').exists()).toBe(false)
    await wrapper.get('[data-code-action="preview"]').trigger('click')
    await flushPromises()
    expect(document.querySelector('[data-testid="chat-code-preview"] code')?.textContent).toContain('const value = "<script>alert(1)</script>"')
    expect(document.querySelector('[data-testid="chat-code-preview"] script')).toBeNull()
    await wrapper.get('[data-code-action="download"]').trigger('click')
    expect(saveAs).toHaveBeenCalledWith(expect.any(Blob), 'snippet.js')
    const blob = vi.mocked(saveAs).mock.calls[0][0] as Blob
    const text = await new Promise<string>((resolve) => {
      const reader = new FileReader()
      reader.onload = () => resolve(String(reader.result))
      reader.readAsText(blob)
    })
    expect(text).toContain('const value = "<script>alert(1)</script>"')
  })

  it('展示真实搜索来源并去重，拒绝脚本链接', async () => {
    seedActiveSession()
    const wrapper = await mountChatView()
    await configureAndSyncModels(wrapper)
    const citation = { type: 'url_citation', title: '来源', url: 'https://source.example/article' }
    vi.mocked(global.fetch).mockResolvedValueOnce(new Response(JSON.stringify({
      output_text: '查询结果',
      output: [{ type: 'message', content: [{ type: 'output_text', text: '查询结果', annotations: [
        citation, citation, { type: 'url_citation', title: 'unsafe', url: 'javascript:alert(1)' },
      ] }] }],
    }), { headers: { 'content-type': 'application/json' } }))
    await wrapper.find('textarea').setValue('查资料')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    const links = wrapper.get('[data-testid="chat-sources"]').findAll('a')
    expect(links).toHaveLength(1)
    expect(links[0].attributes('href')).toBe(citation.url)
    expect(links[0].attributes('rel')).toBe('noopener noreferrer')
    expect(sessionsState.value[0]?.messages.at(-1)?.response?.sources).toEqual([{ url: citation.url, title: citation.title }])
  })

  it('默认 medium，五档推理强度均按选择发送并保存到会话和草稿', async () => {
    seedActiveSession()
    const wrapper = await mountChatView()
    const composer = wrapper.find('textarea')
    const effortSelect = wrapper.get('[data-testid="chat-reasoning-effort-select"]')

    await configureAndSyncModels(wrapper)
    expect(effortSelect.findAll('option').map((option) => option.attributes('value')))
      .toEqual(['low', 'medium', 'high', 'xhigh', 'max'])
    expect((effortSelect.element as HTMLSelectElement).value).toBe('medium')
    await composer.setValue('first question')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(JSON.parse(String(vi.mocked(global.fetch).mock.calls.at(-1)?.[1]?.body)).reasoning.effort).toBe('medium')
    for (const effort of ['low', 'medium', 'high', 'xhigh', 'max']) {
      await effortSelect.setValue(effort)
      await composer.setValue(`question ${effort}`)
      await wrapper.find('form').trigger('submit.prevent')
      await flushPromises()
      const body = JSON.parse(String(vi.mocked(global.fetch).mock.calls.at(-1)?.[1]?.body))
      expect(body.reasoning.effort).toBe(effort)
      expect(sessionsState.value[0]?.reasoningEffort).toBe(effort)
      expect(draftConfigState.value?.reasoningEffort).toBe(effort)
    }
  })

  it.each([
    ['gpt-6-astra', 'responses'],
    ['claude-opus-5-5', 'chat'],
    ['grok-4.7', 'chat'],
    ['gotk-4.6', 'chat'],
  ])(
    '模型 %s 选择兼容的协议入口', async (model, protocol) => {
    const session = seedActiveSession()
    Object.assign(session, { deepThinkingEnabled: true, webSearchEnabled: false })
    const wrapper = await mountChatView()
    const composer = wrapper.find('textarea')

    await configureAndSyncModels(wrapper)
    await wrapper.get('[data-testid="chat-model-select"]').setValue(model)
    await composer.setValue('今天有什么新闻')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const [url, init] = vi.mocked(global.fetch).mock.calls.at(-1)!
    const body = JSON.parse(String(init?.body))
    expect(url).toBe(`http://localhost:8080/v1/${protocol === 'responses' ? 'responses' : 'chat/completions'}`)
    expect(init?.headers).toMatchObject({ Authorization: 'Bearer sk-test-12345678' })
    expect(body.model).toBe(model)
    if (protocol === 'responses') {
      expect(body.tools).toEqual([{ type: 'web_search' }])
      expect(body.reasoning).toEqual({ effort: 'medium' })
      expect(body.input[0]).toEqual({ role: 'user', content: '今天有什么新闻' })
    } else {
      expect(body.tools).toBeUndefined()
      expect(body.reasoning).toBeUndefined()
      expect(body.messages[0]).toEqual({ role: 'user', content: '今天有什么新闻' })
    }
    expect(body.stream).toBe(true)
    if (protocol === 'responses') expect(body.store).toBe(false)
    expect(sessionsState.value[0]?.messages.at(-1)?.content).toBe('已收到')
  })

  it('能处理连接结束前没有空行终止的最后一个 SSE 事件', async () => {
    seedActiveSession()
    const wrapper = await mountChatView()
    const composer = wrapper.find('textarea')

    await configureAndSyncModels(wrapper)
    vi.mocked(global.fetch).mockResolvedValue(createStreamResponse([
      'data: {"type":"response.output_text.delta","delta":"最后一段"}',
    ]))
    await composer.setValue('test partial event')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(sessionsState.value[0]?.messages.at(-1)?.content).toBe('最后一段')
    expect(wrapper.text()).not.toContain('没有收到有效回复')
    expect(sessionsState.value[0]?.messages.at(-1)?.response?.status).toBe('failed')
  })

  it('状态及思考摘要来自实际流式事件，完成状态随消息保存', async () => {
    seedActiveSession()
    const wrapper = await mountChatView()
    await configureAndSyncModels(wrapper)
    let stream!: ReadableStreamDefaultController<Uint8Array>
    vi.mocked(global.fetch).mockResolvedValueOnce(new Response(new ReadableStream<Uint8Array>({
      start(controller) { stream = controller },
    }), { headers: { 'content-type': 'text/event-stream' } }))
    await wrapper.find('textarea').setValue('查资料')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    const sendEvent = async (event: Record<string, unknown>) => {
      stream.enqueue(new TextEncoder().encode(`data: ${JSON.stringify(event)}\n\n`))
      await flushPromises()
    }
    const response = () => sessionsState.value[0]?.messages.at(-1)?.response
    expect(response()?.status).toBe('waiting')
    expect(response()?.reasoningSummary).toBeUndefined()
    await sendEvent({ type: 'response.output_item.added', item: { type: 'reasoning' } })
    expect(response()?.status).toBe('thinking')
    await sendEvent({ type: 'response.reasoning_summary_text.delta', delta: '核对资料来源。' })
    expect(response()?.reasoningSummary).toBe('核对资料来源。')
    await sendEvent({ type: 'response.web_search_call.searching' })
    expect(response()?.status).toBe('searching')
    await sendEvent({ type: 'response.output_text.delta', delta: '结果' })
    expect(response()?.status).toBe('outputting')
    await sendEvent({ type: 'response.completed' })
    stream.close()
    await flushPromises()
    expect(response()?.status).toBe('completed')
    expect(response()?.elapsedMs).toBeGreaterThanOrEqual(0)
    expect(wrapper.text()).toContain('核对资料来源。')
  })

  it('默认 Responses 请求保留历史图片、文本附件和助手回复', async () => {
    seedActiveSession({
      apiKey: 'sk-saved-key',
      model: 'gpt-6-astra',
      messages: [
        {
          id: nextId('user'), role: 'user', content: '分析附件', createdAt: nowIso(),
          attachments: [
            { id: 'image', kind: 'image', name: 'sample.png', mimeType: 'image/png', size: 10, dataUrl: 'data:image/png;base64,aGVsbG8=' },
            { id: 'text', kind: 'text', name: 'notes.txt', mimeType: 'text/plain', size: 10, textContent: 'notes body' },
          ],
        },
        { id: nextId('assistant'), role: 'assistant', content: '历史回复', createdAt: nowIso() },
      ],
    })
    const wrapper = await mountChatView()
    await attachFiles(wrapper, [createMockFile('new.txt', 'new body', 'text/plain')], 0)
    await wrapper.find('textarea').setValue('继续')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    const body = JSON.parse(String(vi.mocked(global.fetch).mock.calls.at(-1)?.[1]?.body))
    expect(body.input).toEqual([
      {
        role: 'user', content: [
          { type: 'input_text', text: '分析附件' },
          { type: 'input_image', image_url: 'data:image/png;base64,aGVsbG8=' },
          { type: 'input_text', text: '[Attached file: notes.txt]\nnotes body\n[End of file: notes.txt]' },
        ],
      },
      { role: 'assistant', content: '历史回复' },
      {
        role: 'user', content: [
          { type: 'input_text', text: '继续' },
          { type: 'input_text', text: '[Attached file: new.txt]\nnew body\n[End of file: new.txt]' },
        ],
      },
    ])
  })

  it('网络层 fetch 失败时显示连接中断提示，而不是空回复提示', async () => {
    seedActiveSession()
    const wrapper = await mountChatView()
    const composer = wrapper.find('textarea')

    await configureAndSyncModels(wrapper)
    vi.mocked(global.fetch).mockRejectedValue(new TypeError('Failed to fetch'))
    await composer.setValue('hello')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const assistantMessage = sessionsState.value[0]?.messages.at(-1)
    expect(assistantMessage?.content).toContain('连接已中断')
    expect(assistantMessage?.content).not.toBe('没有收到有效回复。')
    expect(wrapper.text()).toContain('连接已中断')
    expect((composer.element as HTMLTextAreaElement).value).toBe('hello')
  })

  it.each(['edit', 'regenerate'])('%s 创建正确历史分支，保留原对话和后续消息', async (action) => {
    const original = seedActiveSession({
      apiKey: 'sk-saved-key', model: 'gpt-6-sol', title: '原会话',
      messages: [
        { id: 'u1', role: 'user', content: '原问题', createdAt: nowIso() },
        { id: 'a1', role: 'assistant', content: '原回答', createdAt: nowIso() },
        { id: 'u2', role: 'user', content: '后续问题', createdAt: nowIso() },
        { id: 'a2', role: 'assistant', content: '后续回答', createdAt: nowIso() },
      ],
    })
    const wrapper = await mountChatView()
    const articles = wrapper.findAll('article')
    if (action === 'edit') {
      await articles[0].findAll('button').find((button) => button.text() === '编辑问题')!.trigger('click')
      expect((wrapper.find('textarea').element as HTMLTextAreaElement).value).toBe('原问题')
      await wrapper.find('textarea').setValue('修改的问题')
      await wrapper.find('form').trigger('submit.prevent')
    } else {
      await articles[1].findAll('button').find((button) => button.text() === '重新生成')!.trigger('click')
    }
    await flushPromises()
    const source = sessionsState.value.find((session) => session.id === original.id)!
    expect(source.messages.map((message) => message.content)).toEqual(['原问题', '原回答', '后续问题', '后续回答'])
    expect(activeSessionIdState.value).not.toBe(original.id)
    const body = JSON.parse(String(vi.mocked(global.fetch).mock.calls.at(-1)?.[1]?.body))
    expect(body.input).toEqual([{ role: 'user', content: action === 'edit' ? '修改的问题' : '原问题' }])
    expect(sessionsState.value[0]?.messages.at(-1)?.response?.status).toBe('completed')
  })

  it('失败后手动重试在新分支发送一次，不带错误占位正文', async () => {
    const original = seedActiveSession()
    const wrapper = await mountChatView()
    await configureAndSyncModels(wrapper)
    vi.mocked(global.fetch).mockRejectedValueOnce(new TypeError('Failed to fetch'))
    await wrapper.find('textarea').setValue('重试的问题')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    expect(sessionsState.value[0]?.messages.at(-1)?.response?.status).toBe('failed')
    vi.mocked(global.fetch).mockClear()
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    expect(global.fetch).toHaveBeenCalledOnce()
    expect(activeSessionIdState.value).not.toBe(original.id)
    const body = JSON.parse(String(vi.mocked(global.fetch).mock.calls[0][1]?.body))
    expect(body.input).toEqual([{ role: 'user', content: '重试的问题' }])
    expect(sessionsState.value.find((session) => session.id === original.id)?.messages).toHaveLength(2)
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
      model: 'gpt-5.5',
      apiKey: 'sk-saved-key',
    })
    sessionsState.value = [existingSession]

    const wrapper = await mountChatView()

    expect(sessionsState.value).toHaveLength(1)
    expect(activeSessionIdState.value).toBe(existingSession.id)
    expect(wrapper.text()).toContain('保留的对话')
    expect(sessionsState.value[0]?.model).toBe('gpt-5.5')
    expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('gpt-5.5')
  })

  it('存在草稿配置时只预填设置，不会自动新建对话', async () => {
    draftConfigState.value = {
      apiKey: 'sk-draft-12345678',
      baseUrl: 'https://example.com/v1',
      model: 'claude-sonnet-5-5',
      sessionTitle: '来自草稿',
      reasoningEffort: 'xhigh',
    }

    const wrapper = await mountChatView()

    expect(sessionsState.value).toHaveLength(0)
    expect(activeSessionIdState.value).toBeNull()
    expect((wrapper.find('input[type="password"]').element as HTMLInputElement).value).toBe('sk-draft-12345678')
    expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('claude-sonnet-5-5')
    expect((wrapper.get('[data-testid="chat-reasoning-effort-select"]').element as HTMLSelectElement).value).toBe('xhigh')
    expect(wrapper.text()).toContain('开始本地聊天')

    await startSessionFromUi(wrapper)
    expect(sessionsState.value[0]).toMatchObject({
      model: 'claude-sonnet-5-5', reasoningEffort: 'xhigh',
      apiKey: 'sk-draft-12345678', baseUrl: 'https://example.com/v1',
    })
  })

  it('流式回复期间主按钮切换为停止，并且可以中断回复', async () => {
    seedActiveSession()
    const wrapper = await mountChatView()
    await configureAndSyncModels(wrapper)

    global.fetch = vi.fn().mockImplementation(async (_input: RequestInfo | URL, init?: RequestInit) => {
      const signal = init?.signal as AbortSignal | undefined
      const abortError = Object.assign(new Error('Aborted'), { name: 'AbortError' })

      return {
        ok: true,
        headers: new Headers({ 'content-type': 'text/event-stream' }),
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

    const apiKeyInput = wrapper.find('input[type="password"]')
    const composer = wrapper.find('textarea')

    await apiKeyInput.setValue('sk-test-12345678')
    await composer.setValue('请持续回复')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const primaryButton = wrapper.find('button[type="submit"]')
    expect(primaryButton.text()).toContain('停止')
    expect(composer.attributes('disabled')).toBeUndefined()
    expect(apiKeyInput.attributes('disabled')).toBeDefined()
    expect(wrapper.find('input[type="text"]').attributes('disabled')).toBeDefined()
    expect(wrapper.find('select').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="chat-reasoning-effort-select"]').attributes('disabled')).toBeDefined()
    expect(wrapper.findAll('button').find((button) => button.text().includes('新建对话'))?.attributes('disabled')).toBeDefined()

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('button[type="submit"]').text()).toContain('发送')
    expect(sessionsState.value[0]?.messages.map((message) => message.role)).toEqual(['user'])
  })

  it('密集片段合并显示，耗时刷新不重复解析已有 Markdown', async () => {
    seedActiveSession({ apiKey: 'sk-saved-key' })
    const wrapper = await mountChatView()
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout', 'setInterval', 'clearInterval', 'Date'] })
    let stream!: ReadableStreamDefaultController<Uint8Array>
    vi.mocked(global.fetch).mockResolvedValueOnce(new Response(new ReadableStream<Uint8Array>({
      start(controller) { stream = controller },
    }), { headers: { 'content-type': 'text/event-stream' } }))
    const parse = vi.spyOn(marked, 'parse')
    await wrapper.find('textarea').setValue('长回复')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    stream.enqueue(new TextEncoder().encode('data: {"type":"response.output_text.delta","delta":"A"}\n\n'.repeat(100)))
    await flushPromises()
    expect(sessionsState.value[0]?.messages.at(-1)?.content).toBe('')
    await vi.advanceTimersByTimeAsync(50)
    expect(sessionsState.value[0]?.messages.at(-1)?.content).toBe('A'.repeat(100))
    const parsesAfterText = parse.mock.calls.length
    await vi.advanceTimersByTimeAsync(1000)
    expect(parse.mock.calls.length).toBe(parsesAfterText)
    stream.enqueue(new TextEncoder().encode('data: {"type":"response.completed"}\n\n'))
    stream.close()
    await flushPromises()
    expect(sessionsState.value[0]?.messages.at(-1)?.content).toBe('A'.repeat(100))
    expect(sessionsState.value[0]?.messages.at(-1)?.response?.status).toBe('completed')
  })

  it('未到合并显示时间就停止，仍保留全部已接收文字', async () => {
    seedActiveSession({ apiKey: 'sk-saved-key' })
    const wrapper = await mountChatView()
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout', 'setInterval', 'clearInterval'] })
    vi.mocked(global.fetch).mockImplementationOnce(async (_input, init) => {
      let first = true
      return {
        ok: true, headers: new Headers({ 'content-type': 'text/event-stream' }),
        body: { getReader: () => ({ read: () => {
          if (first) {
            first = false
            return Promise.resolve({ done: false, value: new TextEncoder().encode('data: {"type":"response.output_text.delta","delta":"部分回复"}\n\n') })
          }
          return new Promise((_resolve, reject) => init?.signal?.addEventListener('abort', () => reject(Object.assign(new Error('Aborted'), { name: 'AbortError' })), { once: true }))
        } }) },
      } as unknown as Response
    })
    await wrapper.find('textarea').setValue('开始')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    expect(sessionsState.value[0]?.messages.at(-1)?.content).toBe('')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    expect(sessionsState.value[0]?.messages.at(-1)?.content).toBe('部分回复')
    expect(sessionsState.value[0]?.messages.at(-1)?.response?.status).toBe('stopped')
  })

  it.each(['pagehide', 'unmount'])('页面 %s 时刷新缓冲正文并保存停止状态', async (leaveAction) => {
    seedActiveSession({ apiKey: 'sk-saved-key' })
    const wrapper = await mountChatView()
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout', 'setInterval', 'clearInterval'] })
    let requestSignal: AbortSignal | null | undefined
    vi.mocked(global.fetch).mockImplementationOnce(async (_input, init) => {
      requestSignal = init?.signal
      return new Response(new ReadableStream<Uint8Array>({
        start(controller) {
          controller.enqueue(new TextEncoder().encode('data: {"type":"response.output_text.delta","delta":"离开前已收到的回复"}\n\n'))
          requestSignal?.addEventListener('abort', () => controller.error(new DOMException('Aborted', 'AbortError')), { once: true })
        },
      }), { headers: { 'content-type': 'text/event-stream' } })
    })
    await wrapper.find('textarea').setValue('开始')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    expect(sessionsState.value[0]?.messages.at(-1)?.content).toBe('')
    if (leaveAction === 'pagehide') window.dispatchEvent(new Event('pagehide'))
    else wrapper.unmount()
    expect(requestSignal?.aborted).toBe(true)
    expect(sessionsState.value[0]?.messages.at(-1)).toMatchObject({
      content: '离开前已收到的回复', response: { status: 'stopped' },
    })
    await flushPromises()
    await vi.advanceTimersByTimeAsync(1000)
    expect(sessionsState.value[0]?.messages.at(-1)).toMatchObject({
      content: '离开前已收到的回复', response: { status: 'stopped' },
    })
  })

  it('用户向上阅读时不强制滚动，并提供回到最新消息按钮', async () => {
    seedActiveSession({
      messages: [{
        id: nextId('assistant'),
        role: 'assistant',
        content: '已有回复',
        createdAt: nowIso(),
      }],
    })
    const wrapper = await mountChatView()
    const messages = wrapper.get('[data-testid="chat-messages"]')
    const element = messages.element as HTMLElement

    Object.defineProperties(element, {
      scrollHeight: { configurable: true, value: 1000 },
      clientHeight: { configurable: true, value: 300 },
      scrollTop: { configurable: true, writable: true, value: 0 },
    })
    await messages.trigger('scroll')

    expect(wrapper.get('[data-testid="chat-scroll-to-latest"]').text()).toContain('回到最新消息')

    sessionsState.value[0]?.messages.push({
      id: nextId('assistant'),
      role: 'assistant',
      content: '新增回复',
      createdAt: nowIso(),
    })
    await flushPromises()
    expect(element.scrollTop).toBe(0)

    await wrapper.get('[data-testid="chat-scroll-to-latest"]').trigger('click')
    await flushPromises()
    expect(element.scrollTop).toBe(1000)
  })

  it('显式点击新建对话会创建会话', async () => {
    const wrapper = await mountChatView()

    expect(sessionsState.value).toHaveLength(0)

    await startSessionFromUi(wrapper)

    expect(sessionsState.value).toHaveLength(1)
    expect(activeSessionIdState.value).toBeTruthy()
  })

  it('任务入口只预填草稿，首次发送自动创建会话并保留预先添加的附件', async () => {
    const wrapper = await mountChatView()
    await wrapper.get('[data-testid="chat-starter-debug"]').trigger('click')
    expect(global.fetch).not.toHaveBeenCalled()
    expect(sessionsState.value).toHaveLength(0)
    expect((wrapper.find('textarea').element as HTMLTextAreaElement).value).toBe('chat.starters.debug.prompt')
    await attachFiles(wrapper, [createMockFile('error.log', 'error body', 'text/plain')], 0)
    expect(wrapper.text()).toContain('error.log')
    await configureAndSyncModels(wrapper)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()
    expect(sessionsState.value).toHaveLength(1)
    expect(sessionsState.value[0]?.messages[0]?.attachments?.[0]?.textContent).toBe('error body')
    expect(sessionsState.value[0]?.messages.at(-1)?.response?.status).toBe('completed')
  })

  it('从其他密钥带来的连接草稿不被原会话覆盖', async () => {
    const original = seedActiveSession({ apiKey: 'sk-original', baseUrl: 'https://old.example/v1', model: 'gpt-6-astra' })
    draftConfigState.value = { apiKey: 'sk-new', baseUrl: 'https://new.example/v1' }
    const wrapper = await mountChatView()
    expect(activeSessionIdState.value).toBeNull()
    expect((wrapper.find('input[type="password"]').element as HTMLInputElement).value).toBe('sk-new')
    expect(vi.mocked(global.fetch).mock.calls[0]).toEqual([
      'https://new.example/v1/models', expect.objectContaining({ headers: expect.objectContaining({ Authorization: 'Bearer sk-new' }) }),
    ])
    expect(sessionsState.value.find((session) => session.id === original.id)?.apiKey).toBe('sk-original')
  })
})
