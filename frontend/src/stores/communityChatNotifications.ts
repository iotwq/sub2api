import { defineStore } from 'pinia'
import { ref } from 'vue'
import { communityChatAPI, type CommunityChatEvent } from '@/api'
import { useCommunityChatRealtime } from '@/composables/useCommunityChatRealtime'

export const useCommunityChatNotificationsStore = defineStore('communityChatNotifications', () => {
  const groupUnread = ref(false)
  const directUnread = ref(false)
  let userId: number | null = null
  let generation = 0
  let directSequence = 0
  let refreshPromise: Promise<void> | null = null

  function applyGroupMessage(message?: { id: number; user_id: number }) {
    if (!message || userId === null) return
    const lastSeenId = communityChatAPI.getCommunityChatLastSeenMessageId(userId, 'group')
    if (message.user_id !== userId && message.id > lastSeenId) groupUnread.value = true
  }

  async function refreshGroup(expectedGeneration: number) {
    try {
      const result = await communityChatAPI.listMessages(1, 1)
      if (generation !== expectedGeneration) return
      applyGroupMessage(result.items[result.items.length - 1])
    } catch (error) {
      console.error('Failed to restore community chat notification:', error)
    }
  }

  async function refreshDirect() {
    if (userId === null) return
    const expectedGeneration = generation
    const sequence = ++directSequence
    try {
      const summary = await communityChatAPI.getDirectUnread()
      if (generation === expectedGeneration && sequence === directSequence) {
        directUnread.value = summary.total > 0
      }
    } catch (error) {
      console.error('Failed to refresh community direct chat notification:', error)
    }
  }

  function refreshUnread(): Promise<void> {
    if (userId === null) return Promise.resolve()
    if (refreshPromise) return refreshPromise
    const request = Promise.all([refreshGroup(generation), refreshDirect()])
      .then(() => undefined)
      .finally(() => {
        if (refreshPromise === request) refreshPromise = null
      })
    refreshPromise = request
    return request
  }

  function handleEvent(event: CommunityChatEvent) {
    if (event.type === 'message_created') applyGroupMessage(event.message)
    else if (event.type === 'direct_message_created' || event.type === 'direct_message_deleted') {
      void refreshDirect()
    }
  }

  const realtime = useCommunityChatRealtime({
    onEvent: handleEvent,
    onConnected: () => { void refreshUnread() },
  })

  function handleSeen(event: Event) {
    const detail = (event as CustomEvent<{ userId: number; scope: 'group' | 'direct' }>).detail
    if (!detail || detail.userId !== userId) return
    if (detail.scope === 'group') groupUnread.value = false
    else void refreshDirect()
  }

  function stopSession() {
    if (userId === null) return
    userId = null
    generation++
    refreshPromise = null
    groupUnread.value = false
    directUnread.value = false
    window.removeEventListener('community-chat:seen', handleSeen)
    realtime.stop()
  }

  function startSession(id: number) {
    if (userId === id) return
    stopSession()
    userId = id
    window.addEventListener('community-chat:seen', handleSeen)
    void refreshUnread()
    realtime.start()
  }

  return { groupUnread, directUnread, startSession, stopSession }
})
