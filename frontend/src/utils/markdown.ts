import { shallowRef } from 'vue'
import DOMPurify from 'dompurify'

// marked 只在真正要渲染公告、合规文档等 Markdown 时才下载，不放进登录后每页都加载的外壳代码。
// 解析选项按次传入，不修改 marked 的全局默认值，避免影响其他页面的渲染结果。
type ParseFn = (src: string, options: { async: false; breaks: boolean; gfm: boolean }) => string

const parse = shallowRef<ParseFn | null>(null)
let pending: Promise<void> | null = null

export function loadMarkdown(): Promise<void> {
  if (parse.value) return Promise.resolve()
  if (!pending) {
    pending = import('marked')
      .then(({ marked }) => {
        parse.value = (src, options) => marked.parse(src, options) as string
      })
      .catch((error) => {
        pending = null
        console.error('Failed to load Markdown renderer:', error)
      })
  }
  return pending
}

// 下载期间或失败时保留可读正文；完成后，依赖它的组件会自动重新渲染。
export function renderMarkdown(content: string | null | undefined): string {
  if (!content) return ''
  if (!parse.value) {
    const fallback = document.createElement('div')
    fallback.className = 'whitespace-pre-wrap'
    fallback.textContent = content
    return fallback.outerHTML
  }
  return DOMPurify.sanitize(parse.value(content, { async: false, breaks: true, gfm: true }))
}
