const communityChatTextareaMaxHeight = 144

export function resizeCommunityChatTextarea(input: HTMLTextAreaElement): void {
  input.style.height = 'auto'
  const contentHeight = input.scrollHeight
  input.style.height = `${Math.min(contentHeight, communityChatTextareaMaxHeight)}px`
  input.style.overflowY = contentHeight > communityChatTextareaMaxHeight ? 'auto' : 'hidden'
}

export function firstCommunityChatDroppedFile(data: DataTransfer | null): File | null {
  return data?.files?.[0] ?? null
}
