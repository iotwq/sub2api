import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../CommunityChatView.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('CommunityChatView direct-message unread badge', () => {
  it('renders an independent badge beside the direct-message button', () => {
    expect(componentSource).toContain('<span v-if="directUnread" class="community-direct-unread-badge">')
    expect(componentSource).toContain("t('communityChat.directNewMessage')")
    expect(componentSource).toContain('const directUnread = ref(false)')
  })

  it('restores and updates unread state for incoming direct messages', () => {
    expect(componentSource).toContain('await refreshDirectUnread()')
    expect(componentSource).toContain('communityChatAPI.getDirectUnread()')
    expect(componentSource).toContain("event.type === 'direct_message_created'")
  })

  it('marks only the visible bottom-most direct thread as read', () => {
    const markReadBlock = componentSource.match(/async function markActiveDirectConversationRead\(\): Promise<void> \{[\s\S]*?\n\}/)

    expect(markReadBlock?.[0]).toContain('!directDialogOpen.value')
    expect(markReadBlock?.[0]).toContain('!directMessageListPinnedToBottom.value')
    expect(markReadBlock?.[0]).toContain("document.visibilityState !== 'visible'")
    expect(markReadBlock?.[0]).toContain('communityChatAPI.markDirectRead(latestMessageID, userID)')
  })

  it('supports deleting direct messages and per-conversation unread badges', () => {
    expect(componentSource).toContain('@click="handleDeleteDirect(message)"')
    expect(componentSource).toContain('communityChatAPI.deleteDirectMessage(message.id)')
    expect(componentSource).toContain('conversation.unread_count > 0')
  })

  it('exposes the notification to assistive technology with a reduced-motion fallback', () => {
    expect(componentSource).toContain('aria-describedby="community-contact-hint community-direct-notification"')
    expect(componentSource).toContain('id="community-direct-notification" class="community-direct-notification" role="status" aria-atomic="true"')
    expect(componentSource).toContain('@media (prefers-reduced-motion: reduce)')
  })

  it('supports direct attachments and admin-initiated conversations', () => {
    expect(componentSource).toContain('communityChatAPI.uploadDirectFileMessage')
    expect(componentSource).toContain('communityChatAPI.searchDirectUsers')
    expect(componentSource).toContain('selectedDirectFile')
    expect(componentSource).toContain('directUserSearch')
    expect(componentSource).toContain('community-direct-conversation-list community-direct-scroll-list')
    expect(componentSource).toContain('height: min(70vh, 640px);')
    expect(componentSource).toContain('@apply min-h-0 flex-1 overflow-y-auto p-2;')
    expect(componentSource).toContain('loadMoreDirectConversations')
    expect(componentSource).toContain(':z-index="60"')
    expect(componentSource).toContain(':close-on-escape="previewImage === null && previewVideo === null"')
  })
})
