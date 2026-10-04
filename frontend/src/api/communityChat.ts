import { apiClient } from './client'
import type { AxiosProgressEvent } from 'axios'
import type { PaginatedResponse } from '@/types'

export type CommunityChatMessageType = 'text' | 'image' | 'file'
export type CommunityChatEventType = 'message_created' | 'message_deleted' | 'direct_message_created' | 'direct_message_deleted'

export interface CommunityChatMessage {
  id: number
  conversation_user_id?: number
  user_id: number
  username: string
  avatar_url: string
  message_type: CommunityChatMessageType
  content: string
  image_url: string
  image_mime_type: string
  image_size_bytes: number
  file_url?: string
  file_mime_type?: string
  file_size_bytes?: number
  file_name?: string
  reply_to_message_id?: number | null
  reply_to_username: string
  reply_to_content: string
  reply_to_message_type: CommunityChatMessageType | ''
  reply_to_image_url: string
  deleted_at?: string | null
  deleted_by?: number | null
  sent_at?: string | null
  created_at: string
  updated_at: string
}

export interface CommunityChatDirectConversation {
  user_id: number
  username: string
  avatar_url: string
  last_message?: CommunityChatMessage
  updated_at: string
  unread_count: number
}

export interface CommunityChatDirectUser {
  user_id: number
  username: string
  avatar_url: string
}

export interface CommunityChatDirectUnreadConversation {
  user_id: number
  unread_count: number
  latest_message_id: number
}

export interface CommunityChatDirectUnreadSummary {
  total: number
  conversations: CommunityChatDirectUnreadConversation[]
}

export interface CommunityChatEvent {
  type: CommunityChatEventType
  message?: CommunityChatMessage
  id?: number
  conversation_user_id?: number
}

export type CommunityChatNotificationScope = 'group' | 'direct'

const communityChatLastSeenStoragePrefix = 'community-chat:last-seen'

function communityChatLastSeenStorageKey(userId: number, scope: CommunityChatNotificationScope): string {
  return `${communityChatLastSeenStoragePrefix}:${userId}:${scope}`
}

export function getCommunityChatLastSeenMessageId(userId: number, scope: CommunityChatNotificationScope): number {
  if (userId <= 0) return 0
  const value = Number(localStorage.getItem(communityChatLastSeenStorageKey(userId, scope)))
  return Number.isSafeInteger(value) && value > 0 ? value : 0
}

export function markCommunityChatMessagesSeen(userId: number, scope: CommunityChatNotificationScope, messageId: number): void {
  if (userId <= 0 || !Number.isSafeInteger(messageId) || messageId <= 0) return
  const current = getCommunityChatLastSeenMessageId(userId, scope)
  if (messageId > current) {
    localStorage.setItem(communityChatLastSeenStorageKey(userId, scope), String(messageId))
  }
  window.dispatchEvent(new CustomEvent('community-chat:seen', {
    detail: { userId, scope, messageId },
  }))
}

export async function listMessages(page = 1, pageSize = 50): Promise<PaginatedResponse<CommunityChatMessage>> {
  const { data } = await apiClient.get<PaginatedResponse<CommunityChatMessage>>('/community-chat/messages', {
    params: {
      page,
      page_size: pageSize,
    },
  })
  return data
}

export async function searchMessages(search: string, page = 1, pageSize = 50): Promise<PaginatedResponse<CommunityChatMessage>> {
  const { data } = await apiClient.get<PaginatedResponse<CommunityChatMessage>>('/community-chat/messages', {
    params: {
      search: search.trim(),
      page,
      page_size: pageSize,
    },
  })
  return data
}

async function listMessagesByCursor(params: { before_id?: number; after_id?: number; limit?: number }): Promise<CommunityChatMessage[]> {
  const { data } = await apiClient.get<CommunityChatMessage[]>('/community-chat/messages', { params })
  return data
}

export function listMessagesBefore(beforeId: number, limit = 80): Promise<CommunityChatMessage[]> {
  return listMessagesByCursor({ before_id: beforeId, limit })
}

export function listMessagesAfter(afterId: number, limit = 100): Promise<CommunityChatMessage[]> {
  return listMessagesByCursor({ after_id: afterId, limit })
}

export async function sendTextMessage(content: string, replyToMessageId?: number): Promise<CommunityChatMessage> {
  const payload: { content: string; reply_to_message_id?: number } = { content }
  if (replyToMessageId) {
    payload.reply_to_message_id = replyToMessageId
  }
  const { data } = await apiClient.post<CommunityChatMessage>('/community-chat/messages', payload)
  return data
}

export async function uploadImageMessage(file: File, content?: string, replyToMessageId?: number): Promise<CommunityChatMessage> {
  return uploadFileMessage(file, content, replyToMessageId)
}

export async function uploadFileMessage(file: File, content?: string, replyToMessageId?: number, onUploadProgress?: (event: AxiosProgressEvent) => void): Promise<CommunityChatMessage> {
  const formData = new FormData()
  formData.append('file', file)
  if (content?.trim()) {
    formData.append('content', content.trim())
  }
  if (replyToMessageId) {
    formData.append('reply_to_message_id', String(replyToMessageId))
  }

  const { data } = await apiClient.post<CommunityChatMessage>('/community-chat/files', formData, {
    timeout: 10 * 60 * 1000,
    onUploadProgress,
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
  return data
}

export async function deleteMessage(id: number): Promise<CommunityChatMessage> {
  const { data } = await apiClient.delete<CommunityChatMessage>(`/community-chat/messages/${id}`)
  return data
}

export async function establishAttachmentSession(): Promise<void> {
  await apiClient.post('/community-chat/attachments/session')
}

export async function listDirectConversations(page = 1, pageSize = 30): Promise<PaginatedResponse<CommunityChatDirectConversation>> {
  const { data } = await apiClient.get<PaginatedResponse<CommunityChatDirectConversation>>('/community-chat/direct/conversations', {
    params: {
      page,
      page_size: pageSize,
    },
  })
  return data
}

export async function listDirectMessages(userId?: number, page = 1, pageSize = 80): Promise<PaginatedResponse<CommunityChatMessage>> {
  const params: Record<string, number> = {
    page,
    page_size: pageSize,
  }
  if (userId) {
    params.user_id = userId
  }
  const { data } = await apiClient.get<PaginatedResponse<CommunityChatMessage>>('/community-chat/direct/messages', { params })
  return data
}

async function listDirectMessagesByCursor(
  cursor: { before_id?: number; after_id?: number; limit?: number },
  userId?: number,
): Promise<CommunityChatMessage[]> {
  const params: Record<string, number> = { ...cursor } as Record<string, number>
  if (userId) params.user_id = userId
  const { data } = await apiClient.get<CommunityChatMessage[]>('/community-chat/direct/messages', { params })
  return data
}

export function listDirectMessagesBefore(beforeId: number, userId?: number, limit = 80): Promise<CommunityChatMessage[]> {
  return listDirectMessagesByCursor({ before_id: beforeId, limit }, userId)
}

export function listDirectMessagesAfter(afterId: number, userId?: number, limit = 100): Promise<CommunityChatMessage[]> {
  return listDirectMessagesByCursor({ after_id: afterId, limit }, userId)
}

export async function sendDirectTextMessage(content: string, userId?: number): Promise<CommunityChatMessage> {
  const payload: { content: string; user_id?: number } = { content }
  if (userId) {
    payload.user_id = userId
  }
  const { data } = await apiClient.post<CommunityChatMessage>('/community-chat/direct/messages', payload)
  return data
}

export async function uploadDirectFileMessage(file: File, content?: string, userId?: number, onUploadProgress?: (event: AxiosProgressEvent) => void): Promise<CommunityChatMessage> {
  const formData = new FormData()
  formData.append('file', file)
  if (content?.trim()) formData.append('content', content.trim())
  if (userId) formData.append('user_id', String(userId))

  const { data } = await apiClient.post<CommunityChatMessage>('/community-chat/direct/files', formData, {
    timeout: 10 * 60 * 1000,
    onUploadProgress,
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
  return data
}

export async function searchDirectUsers(search = '', page = 1, pageSize = 20): Promise<PaginatedResponse<CommunityChatDirectUser>> {
  const { data } = await apiClient.get<PaginatedResponse<CommunityChatDirectUser>>('/community-chat/direct/users', {
    params: {
      search,
      page,
      page_size: pageSize,
    },
  })
  return data
}

export async function getDirectUnread(): Promise<CommunityChatDirectUnreadSummary> {
  const { data } = await apiClient.get<CommunityChatDirectUnreadSummary>('/community-chat/direct/unread')
  return data
}

export async function markDirectRead(lastMessageId: number, userId?: number): Promise<number> {
  const payload: { user_id?: number; last_message_id: number } = { last_message_id: lastMessageId }
  if (userId) payload.user_id = userId
  const { data } = await apiClient.post<{ last_message_id: number }>('/community-chat/direct/read', payload)
  return data.last_message_id
}

export async function deleteDirectMessage(id: number): Promise<CommunityChatMessage> {
  const { data } = await apiClient.delete<CommunityChatMessage>(`/community-chat/direct/messages/${id}`)
  return data
}

const communityChatWebSocketProtocol = 'sub2api-chat'

export function buildCommunityChatWebSocketUrl(): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const base = import.meta.env.VITE_API_BASE_URL || '/api/v1'
  const url = new URL(`${base.replace(/\/$/, '')}/community-chat/ws`, window.location.origin)
  url.protocol = protocol
  return url.toString()
}

export function buildCommunityChatWebSocketProtocols(token: string): string[] {
  const protocols = [communityChatWebSocketProtocol]
  const normalizedToken = token.trim()
  if (normalizedToken) {
    protocols.push(`jwt.${normalizedToken}`)
  }
  return protocols
}

export const communityChatAPI = {
  listMessages,
	searchMessages,
  listMessagesBefore,
  listMessagesAfter,
  sendTextMessage,
  uploadImageMessage,
  uploadFileMessage,
  deleteMessage,
  establishAttachmentSession,
  listDirectConversations,
  listDirectMessages,
  listDirectMessagesBefore,
  listDirectMessagesAfter,
  sendDirectTextMessage,
  uploadDirectFileMessage,
  searchDirectUsers,
  getDirectUnread,
  markDirectRead,
  deleteDirectMessage,
  getCommunityChatLastSeenMessageId,
  markCommunityChatMessagesSeen,
  buildCommunityChatWebSocketUrl,
  buildCommunityChatWebSocketProtocols,
}
