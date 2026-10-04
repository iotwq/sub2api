import type { OpenAIModel } from '@/types'

export interface OpenAIModelsResponse {
  object?: string
  data: OpenAIModel[]
}

export interface ChatCompletionMessage {
  role: 'system' | 'user' | 'assistant'
  content: ChatCompletionMessageContent
}

export interface ChatCompletionImageURL {
  url: string
  detail?: 'auto' | 'low' | 'high'
}

export interface ChatCompletionTextPart {
  type: 'text'
  text: string
}

export interface ChatCompletionReasoning {
  effort: 'none' | 'low' | 'medium' | 'high' | 'xhigh' | 'max'
}

export interface ChatCompletionImagePart {
  type: 'image_url'
  image_url: ChatCompletionImageURL
}

export type ChatCompletionContentPart = ChatCompletionTextPart | ChatCompletionImagePart
export type ChatCompletionMessageContent = string | ChatCompletionContentPart[]

export interface ChatCompletionRequest {
  model: string
  messages: ChatCompletionMessage[]
  temperature?: number
  stream?: boolean
  reasoning?: ChatCompletionReasoning
}

export type ResponsesMessageRole = 'system' | 'user' | 'assistant' | 'developer'

export interface ResponsesTextPart {
  type: 'input_text' | 'output_text'
  text: string
}

export interface ResponsesImagePart {
  type: 'input_image'
  image_url: string
}

export type ResponsesContentPart = ResponsesTextPart | ResponsesImagePart
export type ResponsesMessageContent = string | ResponsesContentPart[]

export interface ResponsesInputItem {
  role: ResponsesMessageRole
  content: ResponsesMessageContent
}

export interface ResponsesTool {
  type: 'web_search' | string
}

export interface ResponsesRequest {
  model: string
  input: ResponsesInputItem[] | string
  temperature?: number
  stream?: boolean
  tools?: ResponsesTool[]
  reasoning?: ChatCompletionReasoning
  store?: boolean
}

export interface ChatCompletionChunkChoiceDelta {
  role?: 'assistant'
  content?: string
}

export interface ChatCompletionChunkChoice {
  index: number
  delta?: ChatCompletionChunkChoiceDelta
  finish_reason?: string | null
}

export interface ChatCompletionChunk {
  id?: string
  object?: string
  created?: number
  model?: string
  choices?: ChatCompletionChunkChoice[]
}

export interface ResponsesOutputContentPart {
  type?: string
  text?: string
}

export interface ResponsesOutputItem {
  type?: string
  content?: ResponsesOutputContentPart[]
}

export interface ResponsesResponse {
  output_text?: string
  output?: ResponsesOutputItem[]
  error?: { message?: string }
}

export interface ResponsesStreamEvent {
  type?: string
  delta?: string
  text?: string
  response?: ResponsesResponse
  error?: { message?: string }
}

export async function fetchOpenAIModels(baseUrl: string, apiKey: string): Promise<OpenAIModel[]> {
  const normalizedBase = normalizeBaseUrl(baseUrl)
  const response = await fetch(`${normalizedBase}/models`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${apiKey}`,
      Accept: 'application/json',
    },
  })

  if (!response.ok) {
    throw new Error(`Failed to load models (${response.status})`)
  }

  const data = (await response.json()) as OpenAIModelsResponse
  return Array.isArray(data.data) ? data.data : []
}

export async function streamChatCompletions(
  baseUrl: string,
  apiKey: string,
  payload: ChatCompletionRequest,
  options: {
    signal?: AbortSignal
    onDelta: (delta: string) => void
  }
): Promise<void> {
  const normalizedBase = normalizeBaseUrl(baseUrl)
  const response = await fetch(`${normalizedBase}/chat/completions`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${apiKey}`,
      Accept: 'text/event-stream',
    },
    body: JSON.stringify({ ...payload, stream: true }),
    signal: options.signal,
  })

  if (!response.ok) {
    const message = await extractResponseError(response)
    throw new Error(message)
  }

  if (!response.body) {
    throw new Error('Streaming response body is not available')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  const processRawData = (raw: string) => {
    if (!raw || raw === '[DONE]') {
      return
    }

    try {
      const parsed = JSON.parse(raw) as ChatCompletionChunk | { error?: { message?: string } }
      if ('error' in parsed && parsed.error?.message) {
        throw new Error(parsed.error.message)
      }

      if (!('choices' in parsed)) {
        return
      }

      const delta = parsed.choices?.[0]?.delta?.content
      if (delta) {
        options.onDelta(delta)
      }
    } catch (error) {
      if (error instanceof SyntaxError) {
        return
      }
      throw error
    }
  }

  while (true) {
    const { done, value } = await reader.read()
    if (done) {
      break
    }

    buffer += decoder.decode(value, { stream: true })
    buffer = consumeSSEBuffer(buffer, processRawData)
  }

  buffer += decoder.decode()
  flushSSEBuffer(buffer, processRawData)
}

export async function streamResponses(
  baseUrl: string,
  apiKey: string,
  payload: ResponsesRequest,
  options: {
    signal?: AbortSignal
    onDelta: (delta: string) => void
  }
): Promise<void> {
  const normalizedBase = normalizeBaseUrl(baseUrl)
  const response = await fetch(`${normalizedBase}/responses`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${apiKey}`,
      Accept: 'text/event-stream',
    },
    body: JSON.stringify({ ...payload, stream: true }),
    signal: options.signal,
  })

  if (!response.ok) {
    const message = await extractResponseError(response)
    throw new Error(message)
  }

  if (response.headers.get('content-type')?.includes('application/json')) {
    const data = (await response.json()) as ResponsesResponse
    if (data.error?.message) {
      throw new Error(data.error.message)
    }

    const text = extractResponsesText(data)
    if (text) {
      options.onDelta(text)
    }
    return
  }

  if (!response.body) {
    throw new Error('Streaming response body is not available')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let receivedDelta = false

  const processRawData = (raw: string) => {
    if (!raw || raw === '[DONE]') {
      return
    }

    try {
      const parsed = JSON.parse(raw) as ResponsesStreamEvent
      if (parsed.error?.message) {
        throw new Error(parsed.error.message)
      }

      switch (parsed.type) {
        case 'response.output_text.delta':
        case 'response.refusal.delta':
          if (parsed.delta) {
            receivedDelta = true
            options.onDelta(parsed.delta)
          }
          break
        case 'response.failed':
          if (parsed.response?.error?.message) {
            throw new Error(parsed.response.error.message)
          }
          break
        case 'response.completed':
        case 'response.done':
          if (!receivedDelta && parsed.response) {
            const text = extractResponsesText(parsed.response)
            if (text) {
              receivedDelta = true
              options.onDelta(text)
            }
          }
          break
        default:
          break
      }
    } catch (error) {
      if (error instanceof SyntaxError) {
        return
      }
      throw error
    }
  }

  while (true) {
    const { done, value } = await reader.read()
    if (done) {
      break
    }

    buffer += decoder.decode(value, { stream: true })
    buffer = consumeSSEBuffer(buffer, processRawData)
  }

  buffer += decoder.decode()
  flushSSEBuffer(buffer, processRawData)
}

export function normalizeBaseUrl(input: string): string {
  const trimmed = input.trim().replace(/\/+$/, '')
  if (!trimmed) {
    return `${window.location.origin}/v1`
  }

  if (trimmed.endsWith('/v1')) {
    return trimmed
  }

  return `${trimmed}/v1`
}

function extractResponsesText(response: ResponsesResponse): string {
  if (response.output_text) {
    return response.output_text
  }

  return (response.output || [])
    .flatMap((item) => item.content || [])
    .filter((part) => part.type === 'output_text' && part.text)
    .map((part) => part.text)
    .join('')
}

async function extractResponseError(response: Response): Promise<string> {
  try {
    const raw = await response.text()
    const sseError = extractSSEErrorMessage(raw)
    if (sseError) {
      return sseError
    }

    const data = JSON.parse(raw) as {
      error?: { message?: string }
      message?: string
    }
    return data.error?.message || data.message || `Request failed (${response.status})`
  } catch {
    return `Request failed (${response.status})`
  }
}

function consumeSSEBuffer(buffer: string, onData: (raw: string) => void): string {
  const events = buffer.split(/\r?\n\r?\n/)
  const pending = events.pop() ?? ''
  for (const event of events) {
    processSSEEvent(event, onData)
  }
  return pending
}

function flushSSEBuffer(buffer: string, onData: (raw: string) => void): void {
  if (buffer.trim()) {
    processSSEEvent(buffer, onData)
  }
}

function processSSEEvent(event: string, onData: (raw: string) => void): void {
  const raw = event
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line.startsWith('data:'))
    .map((line) => line.slice(5).trim())
    .join('\n')
    .trim()

  if (raw) {
    onData(raw)
  }
}

function extractSSEErrorMessage(raw: string): string {
  let message = ''
  consumeSSEBuffer(`${raw}\n\n`, (data) => {
    if (message || !data || data === '[DONE]') {
      return
    }
    try {
      const parsed = JSON.parse(data) as ResponsesStreamEvent | { error?: { message?: string } }
      message = parsed.error?.message || ''
      if (!message && 'response' in parsed) {
        message = parsed.response?.error?.message || ''
      }
    } catch {
      // ignore malformed SSE error payloads
    }
  })
  return message
}

export const chatAPI = {
  fetchOpenAIModels,
  streamChatCompletions,
  streamResponses,
  normalizeBaseUrl,
}

export default chatAPI
