import { afterEach, describe, expect, it, vi } from 'vitest'
import { streamResponses } from '../chat'

afterEach(() => vi.unstubAllGlobals())

describe('模型问答 Responses 事件', () => {
  const payload = { model: 'gpt-6-sol', input: 'hello' }

  it('JSON 响应也通知完成事件，保留摘要与输出元数据', async () => {
    const response = { output_text: 'hello', output: [{ type: 'reasoning', summary: [{ text: 'summary' }] }] }
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify(response), {
      headers: { 'content-type': 'application/json' },
    })))
    const onOpen = vi.fn()
    const onDelta = vi.fn()
    const onEvent = vi.fn()
    await streamResponses('https://example.com/v1', 'test-key', payload, { onOpen, onDelta, onEvent })
    expect(onOpen).toHaveBeenCalledOnce()
    expect(onDelta).toHaveBeenCalledWith('hello')
    expect(onEvent).toHaveBeenCalledWith({ type: 'response.completed', response })
  })

  it.each(['response.failed', 'response.incomplete'])('事件 %s 没有错误正文时也不能当作成功', async (type) => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(`data: ${JSON.stringify({ type })}\n\n`, {
      headers: { 'content-type': 'text/event-stream' },
    })))
    await expect(streamResponses('https://example.com/v1', 'test-key', payload, { onDelta: vi.fn() }))
      .rejects.toThrow(type === 'response.failed' ? 'Response failed' : 'Response incomplete')
  })

  it('真实搜索事件传给界面，不生成虚构阶段', async () => {
    const event = { type: 'response.web_search_call.searching', item_id: 'search-1' }
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(`data: ${JSON.stringify(event)}\n\n`, {
      headers: { 'content-type': 'text/event-stream' },
    })))
    const onEvent = vi.fn()
    await streamResponses('https://example.com/v1', 'test-key', payload, { onDelta: vi.fn(), onEvent })
    expect(onEvent).toHaveBeenCalledOnce()
    expect(onEvent).toHaveBeenCalledWith(event)
  })
})
