import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../CommunityChatView.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('CommunityChatView emoji picker wiring', () => {
  it('uses the product emoji picker instead of a fixed quick-emoji row', () => {
    expect(componentSource).toContain('<CommunityEmojiPicker')
    expect(componentSource).toContain('@select="insertEmoji"')
    expect(componentSource).not.toContain('quickEmojis')
    expect(componentSource).not.toContain('community-emoji-row')
  })

  it('tracks and restores the public-chat textarea selection', () => {
    expect(componentSource).toContain('ref="messageInputRef"')
    expect(componentSource).toContain('@select="rememberDraftSelection"')
    expect(componentSource).toContain('insertEmojiAtSelection(')
    expect(componentSource).toContain('messageInputRef.value?.setSelectionRange(result.cursor, result.cursor)')
    expect(componentSource).toContain("if (!window.matchMedia('(pointer: coarse)').matches) messageInputRef.value?.focus()")
  })

  it('keeps the emoji trigger inside the public-chat textarea shell', () => {
    expect(componentSource).toContain('class="community-textarea-shell"')
    expect(componentSource).toContain('class="community-textarea community-textarea-with-emoji"')
    expect(componentSource).toContain('@apply absolute bottom-px right-1;')
  })

  it('supports adaptive public input, file drop, and full-history search', () => {
    expect(componentSource).toContain('@input="handleDraftInput"')
    expect(componentSource).toContain('resizeCommunityChatTextarea(input)')
    expect(componentSource).toContain('@drop.prevent="handleDrop"')
    expect(componentSource).toContain('firstCommunityChatDroppedFile(event.dataTransfer)')
    expect(componentSource).toContain('communityChatAPI.searchMessages')
    expect(componentSource).toContain('searchPreviousScrollTop')
    expect(componentSource).toContain("searchActive.value ? searchContextMessages.value : messages.value")
  })

  it('keeps the direct composer adaptive and centers the contact-owner action', () => {
    expect(componentSource).toContain('ref="directMessageInputRef"')
    expect(componentSource).toContain('@input="resizeDirectMessageInput"')
    expect(componentSource).toContain('watch(directDraft, () => {')
    expect(componentSource).toContain('function resizeDirectMessageInput(): void')
    expect(componentSource).toContain('md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)]')
  })

  it('navigates and highlights individual search results', () => {
    expect(componentSource).toContain("@click=\"navigateSearchResults(-1)\"")
    expect(componentSource).toContain("@click=\"navigateSearchResults(1)\"")
    expect(componentSource).toContain("t('communityChat.searchPosition', { current: currentSearchResultIndex + 1, total: searchTotal })")
    expect(componentSource).toContain("'community-message-search-current': searchActive && currentSearchResult?.id === message.id")
    expect(componentSource).toContain("messageElement.querySelector<HTMLElement>('.community-message-bubble .community-search-keyword-current')")
    expect(componentSource).toContain('centeredCommunityChatSearchScrollTop({')
    expect(componentSource).toContain("container.scrollTo({ top, behavior })")
    expect(componentSource).not.toContain("scrollIntoView({ behavior, block: 'center' })")
    expect(componentSource).toContain('splitCommunityChatSearchText(text, activeSearchQuery.value)')
    expect(componentSource).toContain('class="community-search-keyword"')
    expect(componentSource).toContain('await loadMoreSearchResults()')
    expect(componentSource).toContain('communityChatAPI.listMessagesBefore(message.id, searchContextRadius)')
    expect(componentSource).toContain('communityChatAPI.listMessagesAfter(message.id, searchContextRadius)')
    expect(componentSource).toContain('searchContextMessages.value = mergeMessages(before, [message, ...after])')
    expect(componentSource).not.toContain('class="community-search-more"')
  })
})
