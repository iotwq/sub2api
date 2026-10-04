import { describe, expect, it } from 'vitest'

import { insertEmojiAtSelection } from '../communityChatEmoji'

describe('community chat emoji insertion', () => {
  it('inserts an emoji at the current cursor position', () => {
    expect(insertEmojiAtSelection('hello world', '😊', 5, 5, 2000)).toEqual({
      value: 'hello😊 world',
      cursor: 7,
    })
  })

  it('replaces the selected text and returns the next cursor position', () => {
    expect(insertEmojiAtSelection('hello world', '👍', 6, 11, 2000)).toEqual({
      value: 'hello 👍',
      cursor: 8,
    })
  })

  it('does not exceed the existing message length limit', () => {
    expect(insertEmojiAtSelection('1234', '🚀', 4, 4, 5)).toBeNull()
    expect(insertEmojiAtSelection('1234', '🚀', 2, 4, 5)).toEqual({
      value: '12🚀',
      cursor: 4,
    })
  })
})
