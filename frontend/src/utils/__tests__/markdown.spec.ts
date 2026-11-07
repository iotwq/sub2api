import { describe, expect, it, vi } from 'vitest'
import { loadMarkdown, renderMarkdown } from '../markdown'

const renderer = vi.hoisted(() => {
  let resolve!: () => void
  const ready = new Promise<void>((done) => { resolve = done })
  return { ready, resolve }
})

vi.mock('marked', async (importOriginal) => {
  await renderer.ready
  return importOriginal<typeof import('marked')>()
})

describe('lazy Markdown content', () => {
  it('keeps content readable and inert while the renderer loads, then renders sanitized Markdown', async () => {
    const content = '## Important notice\n\n**Read this first**\n<img src=x onerror="alert(1)">'
    const loading = loadMarkdown()
    try {
      const container = document.createElement('div')
      container.innerHTML = renderMarkdown(content)
      expect(container.textContent).toBe(content)
      expect(container.querySelector('img')).toBeNull()
    } finally {
      renderer.resolve()
      await loading
    }

    const rendered = document.createElement('div')
    rendered.innerHTML = renderMarkdown(content)
    expect(rendered.querySelector('h2')?.textContent).toBe('Important notice')
    expect(rendered.querySelector('strong')?.textContent).toBe('Read this first')
    expect(rendered.querySelector('[onerror]')).toBeNull()
  })

  it('does not add placeholder content to empty announcements', () => {
    expect(renderMarkdown('')).toBe('')
    expect(renderMarkdown(null)).toBe('')
    expect(renderMarkdown(undefined)).toBe('')
  })
})
