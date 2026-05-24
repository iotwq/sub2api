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
  effort: 'none' | 'low' | 'medium' | 'high' | 'xhigh'
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
    onErrorEvent?: (message: string) => void
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

  while (true) {
    const { done, value } = await reader.read()
    if (done) {
      break
    }

    buffer += decoder.decode(value, { stream: true })
    const events = buffer.split('\n\n')
    buffer = events.pop() ?? ''

    for (const event of events) {
      const lines = event
        .split('\n')
        .map((line) => line.trim())
        .filter(Boolean)

      for (const line of lines) {
        if (!line.startsWith('data:')) {
          continue
        }

        const raw = line.slice(5).trim()
        if (!raw || raw === '[DONE]') {
          continue
        }

        try {
          const parsed = JSON.parse(raw) as ChatCompletionChunk | { error?: { message?: string } }
          if ('error' in parsed && parsed.error?.message) {
            options.onErrorEvent?.(parsed.error.message)
            continue
          }

          if (!('choices' in parsed)) {
            continue
          }

          const delta = parsed.choices?.[0]?.delta?.content
          if (delta) {
            options.onDelta(delta)
          }
        } catch {
          // Ignore malformed SSE chunks and keep streaming.
        }
      }
    }
  }
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

async function extractResponseError(response: Response): Promise<string> {
  try {
    const data = (await response.json()) as {
      error?: { message?: string }
      message?: string
    }
    return data.error?.message || data.message || `Request failed (${response.status})`
  } catch {
    return `Request failed (${response.status})`
  }
}

export const chatAPI = {
  fetchOpenAIModels,
  streamChatCompletions,
  normalizeBaseUrl,
}

export default chatAPI
