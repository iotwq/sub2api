import { describe, expect, it } from 'vitest'

import { firstCommunityChatDroppedFile, resizeCommunityChatTextarea } from '../communityChatComposer'

describe('community chat composer', () => {
  it('grows with content, caps at six lines, and shrinks after clearing', () => {
    const textarea = document.createElement('textarea')
    let scrollHeight = 84
    Object.defineProperty(textarea, 'scrollHeight', { get: () => scrollHeight })

    resizeCommunityChatTextarea(textarea)
    expect(textarea.style.height).toBe('84px')
    expect(textarea.style.overflowY).toBe('hidden')

    scrollHeight = 220
    resizeCommunityChatTextarea(textarea)
    expect(textarea.style.height).toBe('144px')
    expect(textarea.style.overflowY).toBe('auto')

    scrollHeight = 42
    resizeCommunityChatTextarea(textarea)
    expect(textarea.style.height).toBe('42px')
    expect(textarea.style.overflowY).toBe('hidden')
  })

  it('uses the first dropped file without sending it automatically', () => {
    const first = new File(['first'], 'first.png', { type: 'image/png' })
    const second = new File(['second'], 'second.pdf', { type: 'application/pdf' })
    const transfer = { files: [first, second] } as unknown as DataTransfer

    expect(firstCommunityChatDroppedFile(transfer)).toBe(first)
    expect(firstCommunityChatDroppedFile(null)).toBeNull()
  })
})
