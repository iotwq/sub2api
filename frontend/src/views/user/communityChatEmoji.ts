export interface EmojiInsertionResult {
  value: string
  cursor: number
}

export function insertEmojiAtSelection(
  value: string,
  emoji: string,
  selectionStart: number,
  selectionEnd: number,
  maxLength: number,
): EmojiInsertionResult | null {
  const start = Math.max(0, Math.min(selectionStart, value.length))
  const end = Math.max(start, Math.min(selectionEnd, value.length))
  const nextValue = `${value.slice(0, start)}${emoji}${value.slice(end)}`
  if (nextValue.length > maxLength) return null

  return {
    value: nextValue,
    cursor: start + emoji.length,
  }
}
