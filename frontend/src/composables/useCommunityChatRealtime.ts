import { readonly, ref, type DeepReadonly, type Ref } from 'vue'
import { communityChatAPI, type CommunityChatEvent } from '@/api'

interface CommunityChatRealtimeSubscriber {
  onEvent: (event: CommunityChatEvent) => void
  onConnected?: () => void
}

interface CommunityChatRealtimeHandle {
  connected: DeepReadonly<Ref<boolean>>
  connecting: DeepReadonly<Ref<boolean>>
  start: () => void
  stop: () => void
}

const connected = ref(false)
const connecting = ref(false)
const subscribers = new Map<number, CommunityChatRealtimeSubscriber>()

let socket: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempt = 0
let nextSubscriberID = 1

function notifyConnected(): void {
  for (const subscriber of subscribers.values()) {
    subscriber.onConnected?.()
  }
}

function notifyEvent(event: CommunityChatEvent): void {
  for (const subscriber of subscribers.values()) {
    subscriber.onEvent(event)
  }
}

function scheduleReconnect(): void {
  if (reconnectTimer || subscribers.size === 0) return
  const baseDelay = Math.min(1000 * (2 ** reconnectAttempt), 30000)
  const delay = baseDelay + Math.floor(Math.random() * 500)
  reconnectAttempt += 1
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    connect()
  }, delay)
}

function connect(): void {
  if (subscribers.size === 0 || socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING) return
  const token = localStorage.getItem('auth_token')?.trim()
  if (!token) return

  const current = new WebSocket(
    communityChatAPI.buildCommunityChatWebSocketUrl(),
    communityChatAPI.buildCommunityChatWebSocketProtocols(token),
  )
  socket = current
  connecting.value = true

  current.onopen = () => {
    if (socket !== current) return
    connected.value = true
    connecting.value = false
    reconnectAttempt = 0
    notifyConnected()
  }
  current.onmessage = (message) => {
    if (socket !== current || typeof message.data !== 'string') return
    try {
      notifyEvent(JSON.parse(message.data) as CommunityChatEvent)
    } catch {
      // Ignore malformed frames and keep the shared connection alive.
    }
  }
  current.onerror = () => {
    if (socket === current) current.close()
  }
  current.onclose = () => {
    if (socket !== current) return
    socket = null
    connected.value = false
    connecting.value = false
    scheduleReconnect()
  }
}

function disconnectWhenIdle(): void {
  if (subscribers.size > 0) return
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  const current = socket
  socket = null
  connected.value = false
  connecting.value = false
  reconnectAttempt = 0
  current?.close()
}

export function useCommunityChatRealtime(subscriber: CommunityChatRealtimeSubscriber): CommunityChatRealtimeHandle {
  let subscriberID: number | null = null
  return {
    connected: readonly(connected),
    connecting: readonly(connecting),
    start: () => {
      if (subscriberID !== null) return
      subscriberID = nextSubscriberID
      nextSubscriberID += 1
      subscribers.set(subscriberID, subscriber)
      connect()
    },
    stop: () => {
      if (subscriberID === null) return
      subscribers.delete(subscriberID)
      subscriberID = null
      disconnectWhenIdle()
    },
  }
}
