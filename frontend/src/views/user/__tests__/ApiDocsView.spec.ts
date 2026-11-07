import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ApiDocsView from '@/views/user/ApiDocsView.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn() }),
}))

const AppLayoutStub = defineComponent({
  template: '<main><slot /></main>',
})

describe('ApiDocsView', () => {
  it('shows one model family at a time, picked by category and model tabs', async () => {
    const wrapper = mount(ApiDocsView, {
      attachTo: document.body,
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          Icon: true,
        },
      },
    })

    const visibleTitles = () => wrapper.findAll('.api-docs-section')
      .filter((section) => section.isVisible())
      .map((section) => section.find('h2').text())
    const tabTexts = (selector: string) => wrapper.findAll(selector).map((tab) => tab.text())
    const pressedTab = (selector: string) => wrapper.findAll(selector)
      .find((tab) => tab.attributes('aria-pressed') === 'true')?.text()

    expect(tabTexts('.api-docs-category-tab')).toEqual(['图片生成 13', '视频生成 13', '音乐生成 3'])
    expect(tabTexts('.api-docs-model-tab')).toEqual(['gpt-image2', 'Seedream', 'Nano Banana', 'Grok'])
    expect(visibleTitles()).toEqual(['OpenAI gpt-image2'])

    await wrapper.findAll('.api-docs-model-tab')[3].trigger('click')
    expect(visibleTitles()).toEqual(['Grok 图片模型'])

    await wrapper.findAll('.api-docs-category-tab')[1].trigger('click')
    expect(pressedTab('.api-docs-category-tab')).toBe('视频生成 13')
    expect(tabTexts('.api-docs-model-tab')).toEqual(['SD2.0', 'Wan3.0', 'MiniMax-H3', 'Grok'])
    expect(pressedTab('.api-docs-model-tab')).toBe('SD2.0')
    expect(visibleTitles()).toEqual(['视频模型SD2.0'])

    await wrapper.findAll('.api-docs-model-tab')[3].trigger('click')
    expect(visibleTitles()).toEqual(['Grok 视频模型'])

    await wrapper.findAll('.api-docs-category-tab')[2].trigger('click')
    expect(tabTexts('.api-docs-model-tab')).toEqual(['Suno'])
    expect(visibleTitles()).toEqual(['音乐模型'])

    wrapper.unmount()
  })

  it('groups the public models into nine user-facing families', () => {
    const wrapper = mount(ApiDocsView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          Icon: true,
        },
      },
    })

    const sections = wrapper.findAll('.api-docs-section')
    expect(sections).toHaveLength(9)
    expect(sections.map((section) => section.find('h2').text())).toEqual([
      'OpenAI gpt-image2',
      'OpenAI 兼容 Seedream',
      '音乐模型',
      'Gemini Nano Banana',
      '视频模型SD2.0',
      '视频模型Wan3.0',
      '视频模型MiniMax-H3',
      'Grok 图片模型',
      'Grok 视频模型',
    ])

    const text = wrapper.text()
    for (const expected of [
      'firefly-video-v2',
      'firefly-video-v2-fast',
      'MiniMax-H3',
      'gemini-3-pro-image-preview',
      'gemini-3.1-flash-image-preview',
      'gemini-nano-banana-2.1',
      'grok-imagine-image-quality',
      'grok-imagine-image-2.0',
      'grok-imagine-video-1.5',
      'grok-imagine-video',
      'gpt-image-2',
      'gpt-image-2.5-sunburst',
      'gpt-image-2.5-flare',
      'gpt-image-2.6-sunburst',
      'gpt-image-2.6-flare',
      'seedream-5.0-pro',
      'seedream-5.0-pro-x',
      'seedream-5.0-lite-x',
      'suno-v6-mini',
      'POST /v1/chat/completions',
      'POST /v1beta/models/gemini-3-pro-image-preview:generateContent',
      'POST /v1beta/models/gemini-3.1-flash-image-preview:generateContent',
      'POST /v1beta/models/gemini-nano-banana-2.1:generateContent',
      'POST /v1/videos/generations',
    ]) {
      expect(text).toContain(expected)
    }
    expect(text).not.toMatch(/sora/i)
    expect(text).not.toContain('即梦')
    expect(text).not.toContain('video-ds-2.0')
    expect(text).not.toContain('video-ds-2.0-fast')
    expect(text).not.toContain('as-sd2.0-fast')
    expect(text).not.toMatch(/midjourney/i)

    const openAISection = sections[0]
    const openAIModelRows = openAISection.find('table').findAll('tbody tr')
    expect(openAIModelRows).toHaveLength(5)
    expect(openAIModelRows.map((row) => row.find('td code').text())).toEqual([
      'gpt-image-2', 'gpt-image-2.5-sunburst', 'gpt-image-2.5-flare', 'gpt-image-2.6-sunburst', 'gpt-image-2.6-flare',
    ])
    for (const row of openAIModelRows) {
      expect(row.findAll('td')[2].text()).toBe('1K / 2K / 4K；最多 16 张参考图')
      expect(row.text()).not.toContain('quality')
    }
    const qualityRow = openAISection.findAll('tbody tr').find((row) => row.find('td code').text() === 'quality')
    expect(qualityRow?.text()).toContain('gpt-image-2 支持 low、medium、high')
    expect(qualityRow?.text()).toContain('gpt-image-2.5-sunburst、gpt-image-2.5-flare、gpt-image-2.6-sunburst 和 gpt-image-2.6-flare 额外支持 xhigh、max')
    expect(openAISection.text()).toContain('1K / 2K / 4K')
    expect(openAISection.text()).toContain('最多 16 张参考图')
    expect(openAISection.text()).toContain('1280x1280')
    expect(openAISection.text()).toContain('output_compression')
    expect(openAISection.text()).toContain('background')
    expect(openAISection.text()).toContain('xhigh')
    expect(openAISection.text()).toContain('max')
    expect(openAISection.text()).toContain('background=transparent')
    expect(openAISection.text()).toContain('response_format')
    expect(openAISection.text()).toContain('b64_json')
    expect(openAISection.text()).toContain("startsWith('data:image/')")
    expect(openAISection.text()).toContain('/^https?:\\/\\//i')
    expect(openAISection.text()).toContain('或同时返回两者')
    expect(openAISection.text()).not.toContain('/v1/videos')
    expect(openAISection.text()).not.toContain('/v1/audio')
    expect(openAISection.text()).not.toContain('/pg/assets')
    expect(openAISection.findAll('.api-docs-code-card')).toHaveLength(3)

    const mappedImageSection = sections[1]
    const mappedImageModelRows = mappedImageSection.find('table').findAll('tbody tr')
    expect(mappedImageModelRows).toHaveLength(3)
    expect(mappedImageModelRows.map((row) => row.find('td code').text())).toEqual([
      'seedream-5.0-pro', 'seedream-5.0-pro-x', 'seedream-5.0-lite-x',
    ])
    expect(mappedImageModelRows[0].text()).toContain('提示词最多 8000 字符')
    expect(mappedImageModelRows[1].text()).toContain('1K / 2K；提示词最多 32000 个 UTF-16 码元；最多 10 张参考图')
    expect(mappedImageModelRows[2].text()).toContain('2K / 3K / 4K；提示词最多 32000 个 UTF-16 码元；最多 14 张参考图')
    const mappedParams = mappedImageSection.findAll('table')[1].findAll('tbody tr')
    const paramText = (name: string) => mappedParams.find((row) => row.find('td code').text() === name)?.text()
    expect(paramText('resolution')).toContain('默认 2K')
    expect(paramText('aspect_ratio')).toContain('默认 1:1')
    expect(paramText('size')).toContain('WIDTHxHEIGHT')
    expect(paramText('n')).toContain('1–10')
    expect(paramText('response_format')).toContain('默认 url')
    expect(mappedImageSection.text()).not.toContain('不支持 mask 遮罩编辑')
    expect(mappedImageSection.text()).toContain('response_format')
    expect(mappedImageSection.text()).not.toContain('(G)')
    const mappedExamples = mappedImageSection.findAll('.api-docs-code-card')
    expect(mappedExamples).toHaveLength(4)
    expect(mappedExamples[0].find('pre code').text()).toContain('"model": "seedream-5.0-pro-x"')
    expect(mappedExamples[0].find('pre code').text()).toContain('"resolution": "2K"')
    expect(mappedExamples[1].find('pre code').text()).toContain('"model": "seedream-5.0-lite-x"')
    expect(mappedExamples[1].find('pre code').text()).toContain('"resolution": "4K"')
    expect(mappedExamples[2].find('pre code').text()).toContain('-F "model=seedream-5.0-lite-x"')
    expect(mappedExamples[2].find('pre code').text()).toContain('-F "resolution=3K"')

    const sunoSection = sections[2]
    expect(sunoSection.find('table').text()).toContain('1 次请求 1 个任务')
    expect(sunoSection.find('table').text()).not.toContain('按渠道/分组按次价格计费')
    expect(sunoSection.text()).toContain('choices[0].message.content')
    expect(sunoSection.text()).toContain('Idempotency-Key')
    expect(sunoSection.text()).toContain('600 秒')
    expect(sunoSection.text()).toContain('一个响应可能包含多段音频')
    expect(sunoSection.text()).toContain('逐个下载')
    expect(sunoSection.text()).toContain('suno-output-random-a.mp3')
    expect(sunoSection.find('table').findAll('tbody tr')).toHaveLength(3)
    expect(sunoSection.text()).not.toContain('旧版 suno')
    expect(sunoSection.text()).not.toContain('/pg/assets')
    expect(sunoSection.text()).not.toContain('Suno 音频 URL 转本站地址')
    expect(sunoSection.text()).not.toContain('不支持参考音频')
    expect(sunoSection.findAll('.api-docs-code-card')).toHaveLength(2)

    const geminiSection = sections[3]
    expect(geminiSection.text()).toContain('Nano Banana 2 原生模型')
    expect(geminiSection.text()).toContain('Nano Banana 2 原生文生图')
    expect(geminiSection.text()).toContain('必须走 Gemini 原生接口')
    expect(geminiSection.text()).not.toContain('/v1/api/nano-banana')
    expect(geminiSection.text()).not.toContain('nano-banana-pro')
    expect(geminiSection.text()).not.toContain('简化 JSON')
    expect(geminiSection.findAll('.api-docs-code-card')).toHaveLength(4)
    expect(geminiSection.find('pre code').element.textContent).toContain([
      'generateContent" \\',
      '  -H "Authorization: Bearer $API_KEY" \\',
    ].join('\n'))

    const fireflySection = sections[4]
    expect(fireflySection.text()).toContain('POST /v1/videos')
    expect(fireflySection.text()).toContain('HEAD /v1/videos/{task_id}/content')
    expect(fireflySection.text()).toContain('Firefly 分镜视频')
    expect(fireflySection.text()).toContain('-F "images=@character-reference.png"')
    expect(fireflySection.text()).toContain('-F "images=@scene-reference.webp"')
    expect(fireflySection.text()).toContain('multipart/form-data 真实文件字段')
    expect(fireflySection.findAll('.api-docs-code-card')).toHaveLength(10)
    const videoModelRows = fireflySection.find('table').findAll('tbody tr')
    expect(videoModelRows).toHaveLength(7)
    const viraldance25Row = videoModelRows.find((row) => row.find('td code').text() === 'viraldance2.5-30')
    expect(viraldance25Row?.find('td code').text()).toBe('viraldance2.5-30')
    expect(fireflySection.text()).not.toContain('画布')
    expect(viraldance25Row?.findAll('td')[1].text()).toBe('sd2.5')
    expect(viraldance25Row?.findAll('td')[2].text()).toBe('4 至 30 秒；720p；支持 30 图 / 10 视频 / 10 音频')
    expect(viraldance25Row?.findAll('td')[3].text()).toBe('16:9 / 9:16 / 1:1')
    expect(fireflySection.text()).toContain('图片总数，不占普通参考图编号')
    expect(fireflySection.text()).toContain('单个 3 至 10 秒，合计不超过 30 秒')
    expect(fireflySection.text()).toContain('单个 3 至 30 秒，合计不超过 30 秒，可单独作为参考')
    expect(fireflySection.text()).toContain('viraldance2.5-30 必填，长度 1 至 5000 字符')
    const viraldance25Cards = fireflySection.findAll('.api-docs-code-card').filter((card) => card.text().includes('viraldance2.5-30'))
    expect(viraldance25Cards).toHaveLength(3)
    const createExample = viraldance25Cards[0].find('pre code').text()
    const referenceExample = viraldance25Cards[1].find('pre code').text()
    for (const example of [createExample, referenceExample]) {
      expect(example).toContain('curl "$API_URL/v1/videos"')
      expect(example).toContain('Authorization: Bearer $API_KEY')
      expect(example).toContain(['application/json" \\', "  -d '{"].join('\n'))
      const payload = JSON.parse(example.match(/-d '([\s\S]+)'$/)![1])
      expect(payload).toMatchObject({ model: 'viraldance2.5-30', resolution: '720p', async: true })
      for (const unsupported of ['ratio', 'video_urls', 'audio_urls', 'generate_audio', 'size']) {
        expect(payload).not.toHaveProperty(unsupported)
      }
    }
    expect(JSON.parse(createExample.match(/-d '([\s\S]+)'$/)![1])).toMatchObject({ duration: 30, aspect_ratio: '16:9' })
    expect(JSON.parse(referenceExample.match(/-d '([\s\S]+)'$/)![1])).toMatchObject({
      duration: 10, aspect_ratio: '9:16',
      image_urls: ['https://example.com/character.png'],
      start_image_url: 'https://example.com/start.png',
      end_image_url: 'https://example.com/end.png',
      video_reference: [{ url: 'https://example.com/motion.mp4' }],
      audio_reference: [{ url: 'https://example.com/music.mp3' }],
    })
    for (const label of ['@图片1', '@视频1', '@音频1']) expect(referenceExample).toContain(label)
    const pollExample = viraldance25Cards[2].find('pre code').text()
    expect(pollExample).toContain('每 5 秒查询一次')
    expect(pollExample).toContain('curl "$API_URL/v1/videos/$TASK_ID"')
    expect(pollExample).toContain('completed 后读取顶层 url')
    expect(pollExample).toContain('curl -L "$VIDEO_URL" -o result.mp4')
    expect(pollExample).not.toContain('$TASK_ID/content')

    const miniMaxSection = sections[6]
    expect(miniMaxSection.text()).toContain('POST /pg/assets')
    expect(miniMaxSection.text()).toContain('POST /v1/videos')
    expect(miniMaxSection.text()).toContain('4 至 15 秒')
    expect(miniMaxSection.text()).toContain('768P / 2K')
    expect(miniMaxSection.text()).toContain('"role": "first_frame"')
    expect(miniMaxSection.text()).toContain('"role": "reference_video"')
    expect(miniMaxSection.text()).toContain('单个不超过 50MB')
    expect(miniMaxSection.text()).toContain('单个不超过 15MB')
    expect(miniMaxSection.text()).toContain('task.status 为 succeeded 后再下载')
    expect(miniMaxSection.text()).toContain('参考视频按实际秒数计费')
    expect(miniMaxSection.findAll('.api-docs-code-card')).toHaveLength(5)

    const grokImageSection = sections[7]
    expect(grokImageSection.text()).toContain('POST /v1/images/edits')
    expect(grokImageSection.text()).toContain('Grok 图生图（1 至 3 张参考图）')
    expect(grokImageSection.text()).not.toContain('/v1/videos')
    expect(grokImageSection.findAll('.api-docs-code-card')).toHaveLength(4)

    const grokVideoSection = sections[8]
    expect(grokVideoSection.text()).toContain('POST /v1/videos/generations')
    expect(grokVideoSection.text()).toContain('Grok 多图生视频')
    expect(grokVideoSection.text()).toContain('"reference_images": [')
    expect(grokVideoSection.text()).not.toContain('/v1/images')
    expect(grokVideoSection.findAll('.api-docs-code-card')).toHaveLength(4)

    expect(text).not.toContain('video-v1-15s')
    expect(text).not.toContain('/v1/video/generations')
  })
})
