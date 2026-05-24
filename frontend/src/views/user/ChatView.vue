<template>
  <AppLayout>
    <div class="chat-page relative overflow-hidden rounded-[2rem] border border-[#e8e0d3] bg-[radial-gradient(circle_at_top,_rgba(255,255,255,0.96),_rgba(249,246,240,0.93)_42%,_rgba(240,234,226,0.88)_100%)] p-3 shadow-[0_32px_80px_rgba(15,23,42,0.08)] dark:border-[#2a2f38] dark:bg-[radial-gradient(circle_at_top,_rgba(33,28,26,0.96),_rgba(15,18,23,0.95)_46%,_rgba(8,10,14,1)_100%)] md:p-4">
      <div class="pointer-events-none absolute inset-x-0 top-0 h-40 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.78),transparent_72%)] dark:bg-[radial-gradient(circle_at_top,rgba(255,244,214,0.08),transparent_72%)]"></div>

      <div class="relative flex h-[calc(100vh-8rem)] min-h-[680px] flex-col gap-4 xl:grid xl:grid-cols-[minmax(248px,272px)_minmax(0,1fr)]">
        <aside class="flex w-full flex-col overflow-hidden rounded-[1.75rem] border border-white/70 bg-white/82 shadow-[0_18px_45px_rgba(15,23,42,0.06)] backdrop-blur dark:border-[#242933] dark:bg-[#10141a]/84 xl:w-auto">
          <div class="border-b border-[#efe8dd] p-4 dark:border-dark-700">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t('chat.title') }}
                </h2>
                <p class="mt-1 text-sm leading-6 text-gray-500 dark:text-gray-400">
                  {{ t('chat.sidebarDescription') }}
                </p>
              </div>
              <button class="btn btn-primary shrink-0" @click="handleCreateSession">
                <Icon name="plus" size="sm" class="mr-1.5" />
                {{ t('chat.newChat') }}
              </button>
            </div>
          </div>

          <div class="border-b border-[#efe8dd] p-4 dark:border-dark-700">
            <div class="space-y-3.5">
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-[0.2em] text-gray-500 dark:text-gray-400">
                  {{ t('chat.baseUrl') }}
                </label>
                <input
                  v-model="baseUrl"
                  type="text"
                  class="input"
                  :placeholder="t('chat.baseUrlPlaceholder')"
                />
              </div>

              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-[0.2em] text-gray-500 dark:text-gray-400">
                  {{ t('chat.apiKey') }}
                </label>
                <input
                  v-model="apiKey"
                  type="password"
                  autocomplete="off"
                  class="input"
                  :placeholder="t('chat.apiKeyPlaceholder')"
                />
              </div>

              <div>
                <div class="mb-1.5 flex items-center justify-between gap-2">
                  <label class="block text-xs font-semibold uppercase tracking-[0.2em] text-gray-500 dark:text-gray-400">
                    {{ t('chat.model') }}
                  </label>
                  <button
                    type="button"
                    class="text-xs font-medium text-primary-600 transition-colors hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
                    :disabled="loadingModels"
                    @click="loadModels"
                  >
                    {{ loadingModels ? t('chat.loadingModels') : t('chat.refreshModels') }}
                  </button>
                </div>
                <Select
                  :model-value="selectedModel || null"
                  :options="modelOptions"
                  :placeholder="t('chat.modelPlaceholder')"
                  searchable
                  @update:model-value="handleModelChange"
                />
              </div>

              <div class="flex items-center justify-between gap-3 rounded-[1.35rem] border border-[#eadfcd] bg-[#faf7f1] px-3.5 py-3 dark:border-dark-600 dark:bg-dark-800/90">
                <div class="min-w-0">
                  <div class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-500 dark:text-gray-400">
                    {{ t('chat.deepThinking') }}
                  </div>
                  <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">
                    {{ t('chat.deepThinkingHint') }}
                  </p>
                </div>

                <div class="flex shrink-0 items-center gap-2">
                  <span
                    class="rounded-full px-2.5 py-1 text-[11px] font-semibold tracking-[0.2em]"
                    :class="deepThinkingEnabled
                      ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-200'
                      : 'bg-gray-200 text-gray-600 dark:bg-dark-700 dark:text-gray-300'"
                  >
                    {{ deepThinkingEnabled ? t('chat.deepThinkingStateOn') : t('chat.deepThinkingStateOff') }}
                  </span>
                  <Toggle
                    v-model="deepThinkingEnabled"
                    :aria-label="t('chat.deepThinking')"
                    data-testid="chat-deep-thinking-toggle"
                  />
                </div>
              </div>

            </div>
          </div>

          <div class="min-h-0 flex-1 overflow-y-auto p-3">
            <div v-if="sessions.length === 0" class="rounded-2xl border border-dashed border-[#e5dccf] bg-[#fbf8f3] p-5 text-center text-sm text-gray-500 dark:border-dark-600 dark:bg-dark-900/40 dark:text-gray-400">
              {{ t('chat.noSessions') }}
            </div>

            <div v-else class="space-y-2.5">
              <button
                v-for="session in sessions"
                :key="session.id"
                class="group w-full rounded-[1.35rem] border p-3.5 text-left transition-all duration-200"
                :class="session.id === activeSession?.id
                  ? 'border-[#d8c3a1] bg-[#fbf6ee] shadow-[0_12px_28px_rgba(124,91,45,0.10)] dark:border-[#8a6a35] dark:bg-[#2a2117]'
                  : 'border-transparent bg-white/75 hover:border-[#eadfcd] hover:bg-white dark:bg-dark-800/80 dark:hover:border-dark-500 dark:hover:bg-dark-800'"
                @click="selectSession(session.id)"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <div class="truncate text-sm font-semibold text-gray-900 dark:text-white">
                      {{ session.title }}
                    </div>
                    <div class="mt-1.5 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                      <span v-if="session.apiKeyHint" class="truncate">{{ session.apiKeyHint }}</span>
                    </div>
                    <div v-if="session.messages.length" class="mt-2 line-clamp-2 text-xs leading-5 text-gray-500 dark:text-gray-400">
                      {{ sessionMessagePreview(session.messages[session.messages.length - 1]) }}
                    </div>
                  </div>
                  <button
                    type="button"
                    class="rounded-xl p-1 text-gray-400 opacity-0 transition-all hover:bg-red-50 hover:text-red-500 group-hover:opacity-100 dark:hover:bg-red-900/20"
                    :title="t('common.delete')"
                    @click.stop="removeSession(session.id)"
                  >
                    <Icon name="trash" size="sm" />
                  </button>
                </div>
              </button>
            </div>
          </div>
        </aside>

        <section class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-[1.75rem] border border-white/70 bg-[#fcfaf6]/88 shadow-[inset_0_1px_0_rgba(255,255,255,0.7),0_20px_56px_rgba(15,23,42,0.07)] backdrop-blur dark:border-[#242933] dark:bg-[#11161d]/84">
          <div class="border-b border-[#efe8dd] px-5 py-4 dark:border-dark-700">
            <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
              <div>
                <div class="flex flex-wrap items-center gap-2">
                  <h1 class="text-xl font-semibold text-gray-900 dark:text-white">
                    {{ activeSession?.title || t('chat.emptyTitle') }}
                  </h1>
                  <span
                    v-if="activeSession?.model"
                    class="rounded-full border border-[#eadfcd] bg-white px-3 py-1 text-xs font-medium text-gray-600 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300"
                  >
                    {{ activeSession.model }}
                  </span>
                </div>
              </div>

              <div class="flex flex-wrap items-center gap-2">
                <button class="btn btn-secondary" :disabled="!activeSession || isStreaming || !activeSession.messages.length" @click="clearCurrentConversation">
                  {{ t('chat.clearMessages') }}
                </button>
                <button class="btn btn-secondary" :disabled="!sessions.length || isStreaming" @click="clearAllSessions">
                  {{ t('chat.clearAll') }}
                </button>
              </div>
            </div>
          </div>

          <div
            ref="messagesContainerRef"
            class="min-h-0 flex-1 overflow-y-auto bg-[linear-gradient(180deg,rgba(255,255,255,0.26)_0%,rgba(247,242,234,0.72)_100%)] px-4 py-5 dark:bg-[linear-gradient(180deg,rgba(30,27,25,0.12)_0%,rgba(11,13,17,0.34)_100%)] md:px-5 md:py-6"
          >
            <div v-if="!activeSession" class="flex h-full items-center justify-center">
              <div class="max-w-xl rounded-[2rem] border border-dashed border-[#dfd3c0] bg-white/75 px-8 py-10 text-center shadow-[0_18px_48px_rgba(15,23,42,0.06)] dark:border-dark-600 dark:bg-dark-900/55">
                <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-[1.4rem] bg-[#f5ecdf] text-[#886633] dark:bg-dark-800 dark:text-[#d8be8d]">
                  <Icon name="chatBubble" size="xl" />
                </div>
                <h2 class="mt-5 text-2xl font-semibold text-gray-900 dark:text-white">
                  {{ t('chat.startTitle') }}
                </h2>
                <p class="mt-3 text-sm leading-7 text-gray-500 dark:text-gray-400">
                  {{ t('chat.startDescription') }}
                </p>
                <button class="btn btn-primary mt-6" @click="handleCreateSession">
                  {{ t('chat.newChat') }}
                </button>
              </div>
            </div>

            <div v-else class="mx-auto flex w-full max-w-[1120px] flex-col gap-5">
              <div
                v-for="message in activeSession.messages"
                :key="message.id"
                class="flex"
                :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
              >
                <article
                  class="group w-full rounded-[1.75rem] px-4 py-4 shadow-[0_18px_44px_rgba(15,23,42,0.06)] md:px-5"
                  :class="[messageBubbleWidthClass(message), messageCardClass(message)]"
                >
                  <header class="mb-3 flex items-start justify-between gap-3">
                    <div class="min-w-0">
                      <div
                        class="inline-flex items-center gap-2 rounded-full px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.24em]"
                        :class="messageMetaClass(message)"
                      >
                        <span
                          class="flex h-5 w-5 items-center justify-center rounded-full"
                          :class="message.role === 'user'
                            ? 'bg-white/12 text-white'
                            : 'bg-[#f5ecdf] text-[#8a6734] dark:bg-dark-700 dark:text-[#d8be8d]'"
                        >
                          <Icon :name="message.role === 'user' ? 'user' : 'sparkles'" size="xs" />
                        </span>
                        <span>{{ message.role === 'user' ? t('chat.you') : t('chat.assistant') }}</span>
                        <span class="opacity-70">{{ formatRelativeTime(message.createdAt) }}</span>
                      </div>
                    </div>

                    <div
                      v-if="message.role === 'assistant' && message.content.trim()"
                      class="flex items-center gap-2 opacity-100 transition-opacity md:opacity-0 md:group-hover:opacity-100"
                    >
                      <button
                        type="button"
                        class="chat-action-button"
                        @click="copyMessageContent(message)"
                      >
                        <Icon :name="copiedMessageId === message.id ? 'check' : 'copy'" size="sm" />
                        <span>{{ copiedMessageId === message.id ? t('chat.copied') : t('chat.copyMessage') }}</span>
                      </button>
                    </div>
                  </header>

                  <div
                    v-if="message.content.trim()"
                    class="chat-markdown"
                    :class="message.role === 'user' ? 'chat-markdown-user' : 'chat-markdown-assistant'"
                    @click="handleMarkdownActionClick"
                    v-html="renderMessage(message.content)"
                  ></div>

                  <div v-if="message.attachments?.length" class="mt-4 space-y-2.5">
                    <div
                      v-for="attachment in message.attachments"
                      :key="attachment.id"
                      class="overflow-hidden rounded-[1.25rem]"
                    >
                      <div
                        v-if="attachment.kind === 'image' && attachment.dataUrl"
                        class="overflow-hidden rounded-[1.25rem] border"
                        :class="messageAttachmentClass(message)"
                      >
                        <img
                          :src="attachment.dataUrl"
                          :alt="attachment.name"
                          class="max-h-72 w-full object-contain"
                        />
                        <div class="flex items-center justify-between gap-2 px-3 py-2 text-xs opacity-80">
                          <span class="truncate">{{ attachment.name }}</span>
                          <span>{{ formatAttachmentMeta(attachment) }}</span>
                        </div>
                      </div>

                      <div
                        v-else
                        class="flex items-center gap-3 rounded-[1.25rem] border px-3 py-3"
                        :class="messageAttachmentClass(message)"
                      >
                        <div
                          class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl"
                          :class="message.role === 'user'
                            ? 'bg-white/10 text-white'
                            : 'bg-white text-[#7c5b2d] dark:bg-dark-700 dark:text-[#d8be8d]'"
                        >
                          <Icon :name="attachmentIconName(attachment)" size="md" />
                        </div>
                        <div class="min-w-0 flex-1">
                          <div class="truncate text-sm font-medium">{{ attachment.name }}</div>
                          <div class="mt-1 text-xs opacity-70">{{ formatAttachmentMeta(attachment) }}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </article>
              </div>

              <div v-if="isStreaming" class="flex justify-start">
                <div class="inline-flex items-center gap-3 rounded-full border border-[#e8ddcd] bg-white/90 px-4 py-2.5 text-sm text-gray-600 shadow-sm dark:border-dark-600 dark:bg-dark-900/85 dark:text-gray-300">
                  <span class="h-2.5 w-2.5 animate-pulse rounded-full bg-primary-500"></span>
                  <span>{{ t('chat.streaming') }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="border-t border-[#efe8dd] px-4 py-4 dark:border-dark-700 md:px-5">
            <div class="mx-auto w-full max-w-5xl">
              <div v-if="errorMessage" class="mb-2.5 rounded-[1.15rem] border border-red-200 bg-red-50 px-3.5 py-2.5 text-[13px] text-red-700 dark:border-red-900/60 dark:bg-red-900/20 dark:text-red-300">
                {{ errorMessage }}
              </div>

              <form
                class="rounded-[1.6rem] border border-[#eadfcd] bg-white/94 p-3 shadow-[0_16px_46px_rgba(15,23,42,0.06)] transition-colors dark:border-[#2b3039] dark:bg-[#10141a]/94 md:p-3.5"
                :class="composerClass"
                @submit.prevent="handlePrimaryAction"
                @dragenter.prevent="handleComposerDragEnter"
                @dragover.prevent="handleComposerDragOver"
                @dragleave.prevent="handleComposerDragLeave"
                @drop.prevent="handleComposerDrop"
              >
                <input
                  ref="attachmentInputRef"
                  type="file"
                  class="hidden"
                  :accept="attachmentFileAccept"
                  multiple
                  @change="handleAttachmentSelection"
                />

                <div
                  v-if="isDragActive"
                  class="mb-3 rounded-2xl border border-dashed border-primary-300 bg-primary-50/80 px-4 py-3 text-sm text-primary-700 dark:border-primary-700 dark:bg-primary-900/20 dark:text-primary-200"
                >
                  <div class="flex items-center gap-2 font-medium">
                    <Icon name="upload" size="sm" />
                    {{ t('chat.dropzoneTitle') }}
                  </div>
                  <div class="mt-1 text-xs text-primary-600/90 dark:text-primary-200/80">
                    {{ t('chat.dropzoneDescription') }}
                  </div>
                </div>

                <textarea
                  ref="composerTextareaRef"
                  v-model="draftMessage"
                  class="min-h-[92px] w-full resize-none overflow-y-auto border-0 bg-transparent px-2 py-1.5 text-[15px] leading-6 text-gray-900 outline-none placeholder:text-gray-400 dark:text-white dark:placeholder:text-gray-500"
                  :placeholder="t('chat.inputPlaceholder')"
                  :disabled="!activeSession || processingAttachments"
                  @input="resizeComposerTextarea"
                  @keydown.enter.exact.prevent="handleComposerEnter"
                  @keydown.enter.shift.exact.stop
                ></textarea>

                <div v-if="draftAttachments.length" class="mt-3 flex flex-wrap gap-2 px-2">
                  <div
                    v-for="attachment in draftAttachments"
                    :key="attachment.id"
                    class="flex items-center gap-3 rounded-[1.25rem] border border-[#e9dfd1] bg-[#faf7f1] px-3 py-2.5 text-sm dark:border-dark-600 dark:bg-dark-800"
                  >
                    <img
                      v-if="attachment.kind === 'image' && attachment.dataUrl"
                      :src="attachment.dataUrl"
                      :alt="attachment.name"
                      class="h-12 w-12 rounded-xl object-cover"
                    />
                    <div
                      v-else
                      class="flex h-12 w-12 items-center justify-center rounded-xl bg-white text-[#8d7449] dark:bg-dark-700 dark:text-[#d7c097]"
                    >
                      <Icon :name="attachmentIconName(attachment)" size="md" />
                    </div>
                    <div class="min-w-0">
                      <div class="max-w-[180px] truncate font-medium text-gray-900 dark:text-white">
                        {{ attachment.name }}
                      </div>
                      <div class="text-xs text-gray-500 dark:text-gray-400">
                        {{ formatAttachmentMeta(attachment) }}
                      </div>
                    </div>
                    <button
                      type="button"
                      class="rounded-xl p-1 text-gray-400 transition-colors hover:bg-gray-200 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-200"
                      :title="t('chat.removeAttachment')"
                      @click="removeDraftAttachment(attachment.id)"
                    >
                      <Icon name="x" size="sm" />
                    </button>
                  </div>
                </div>

                <div
                  v-if="attachmentFeedbackMessages.length"
                  class="mt-3 space-y-2 px-2"
                >
                  <div
                    v-for="feedback in attachmentFeedbackMessages"
                    :key="feedback.id"
                    class="rounded-2xl border px-3 py-2 text-sm"
                    :class="feedback.kind === 'error'
                      ? 'border-red-200 bg-red-50 text-red-700 dark:border-red-900/60 dark:bg-red-900/20 dark:text-red-300'
                      : 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900/60 dark:bg-amber-900/20 dark:text-amber-300'"
                  >
                    <div class="flex items-start gap-2">
                      <Icon :name="feedback.kind === 'error' ? 'exclamationTriangle' : 'infoCircle'" size="sm" class="mt-0.5 shrink-0" />
                      <div class="min-w-0">
                        <div class="font-medium">
                          {{ feedback.fileName }}
                        </div>
                        <div class="mt-0.5 break-words text-xs opacity-90">
                          {{ feedback.message }}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="mt-2.5 flex flex-wrap items-center justify-between gap-3 border-t border-[#efe8dd] px-2 pt-2.5 dark:border-dark-700">
                  <div class="flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                    <button
                      type="button"
                      class="inline-flex items-center gap-1.5 rounded-full border border-[#eadfcd] bg-[#faf6ef] px-3.5 py-1.5 transition-colors hover:border-[#d9c4a1] hover:text-[#7b5a2d] dark:border-[#343944] dark:bg-[#181d24] dark:hover:border-[#8d6b39] dark:hover:text-[#dcc391]"
                      :disabled="!activeSession || processingAttachments"
                      @click="openAttachmentPicker"
                    >
                      <Icon name="upload" size="sm" />
                      {{ t('chat.uploadAttachment') }}
                    </button>
                    <span class="leading-5">{{ processingAttachments ? attachmentProcessingLabel : t('chat.supportedFileHint') }}</span>
                  </div>

                  <div class="flex items-center gap-2">
                    <button
                      type="submit"
                      class="btn btn-primary"
                      :disabled="isStreaming ? !activeAbortController : !canSend"
                    >
                      <Icon :name="isStreaming ? 'x' : 'arrowUp'" size="sm" class="mr-1.5" />
                      {{ isStreaming ? t('chat.stop') : t('chat.send') }}
                    </button>
                  </div>
                </div>
              </form>
            </div>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import mammoth from 'mammoth'
import * as pdfjsLib from 'pdfjs-dist/build/pdf.mjs'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import { chatAPI } from '@/api/chat'
import { useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import { useLocalChat } from '@/composables/useLocalChat'
import type { OpenAIModel } from '@/types'
import type { ChatCompletionContentPart, ChatCompletionMessage } from '@/api/chat'
import type { LocalChatAttachment, LocalChatMessage } from '@/composables/useLocalChat'
import type { TextItem } from 'pdfjs-dist/types/src/display/api'

const DEFAULT_CHAT_MODEL = 'gpt-5.4'
const MAX_ATTACHMENTS = 5
const MAX_ATTACHMENT_FILE_SIZE_BYTES = 10 * 1024 * 1024
const MAX_IMAGE_SIZE_BYTES = MAX_ATTACHMENT_FILE_SIZE_BYTES
const MAX_TEXT_FILE_SIZE_BYTES = MAX_ATTACHMENT_FILE_SIZE_BYTES
const MAX_PDF_FILE_SIZE_BYTES = MAX_ATTACHMENT_FILE_SIZE_BYTES
const MAX_DOCX_FILE_SIZE_BYTES = MAX_ATTACHMENT_FILE_SIZE_BYTES
const SUPPORTED_TEXT_FILE_EXTENSIONS = new Set([
  'txt', 'md', 'markdown', 'json', 'csv', 'tsv', 'js', 'jsx', 'ts', 'tsx', 'css',
  'html', 'htm', 'xml', 'yaml', 'yml', 'py', 'go', 'java', 'c', 'cpp', 'h', 'hpp',
  'sql', 'log', 'ini', 'conf', 'sh', 'bash'
])
const FIXED_CHAT_MODELS: OpenAIModel[] = [
  { id: 'gpt-5.5', display_name: 'GPT-5.5' },
  { id: 'gpt-5.4', display_name: 'GPT-5.4' },
  { id: 'gpt-5.4-mini', display_name: 'GPT-5.4 Mini' },
]
const attachmentFileAccept = [
  'image/*',
  '.txt,.md,.markdown,.json,.csv,.tsv,.js,.jsx,.ts,.tsx,.css,.html,.htm,.xml,.yaml,.yml,.py,.go,.java,.c,.cpp,.h,.hpp,.sql,.log,.ini,.conf,.sh,.bash,.pdf,.docx',
  'text/*',
  'application/json',
  'application/xml',
  'application/pdf',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
].join(',')
const PDF_PAGE_PREVIEW_LIMIT = 80
const COPY_FEEDBACK_DURATION_MS = 1800
const MIN_COMPOSER_HEIGHT_PX = 92
const MAX_COMPOSER_HEIGHT_PX = 220

interface AttachmentFeedback {
  id: string
  kind: 'error' | 'warning'
  fileName: string
  message: string
}

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const {
  sessions,
  activeSession,
  createSession,
  setActiveSession,
  updateSession,
  deleteSession,
  appendMessage,
  updateMessage,
  clearAllSessions: clearLocalSessions,
  saveDraftConfig,
  consumeDraftConfig,
} = useLocalChat()

const baseUrl = ref('')
const apiKey = ref('')
const selectedModel = ref('')
const deepThinkingEnabled = ref(false)
const models = ref<OpenAIModel[]>(createFixedModelOptions())
const loadingModels = ref(false)
const draftMessage = ref('')
const draftAttachments = ref<LocalChatAttachment[]>([])
const isStreaming = ref(false)
const processingAttachments = ref(false)
const processingAttachmentName = ref('')
const errorMessage = ref('')
const attachmentFeedbackMessages = ref<AttachmentFeedback[]>([])
const messagesContainerRef = ref<HTMLElement | null>(null)
const activeAbortController = ref<AbortController | null>(null)
const attachmentInputRef = ref<HTMLInputElement | null>(null)
const composerTextareaRef = ref<HTMLTextAreaElement | null>(null)
const isDragActive = ref(false)
const dragDepth = ref(0)
const copiedMessageId = ref<string | null>(null)
const selectedReasoningEffort = computed(() => (deepThinkingEnabled.value ? 'xhigh' : 'medium'))

const messageRenderer = new marked.Renderer()
messageRenderer.code = ({ text, lang, escaped }) => {
  const language = normalizeCodeLanguage(lang)
  const codeContent = escaped ? text : escapeHtml(text)
  const languageClass = language ? ` class="language-${escapeHtmlAttribute(language)}"` : ''
  const label = formatCodeLanguageLabel(language)
  const copyCodeLabel = t('chat.copyCode')

  return [
    '<div class="chat-code-block">',
    '  <div class="chat-code-header">',
    `    <span class="chat-code-label">${escapeHtml(label)}</span>`,
    `    <button type="button" class="chat-code-copy" data-copy-type="code" data-copy-default-label="${escapeHtmlAttribute(copyCodeLabel)}" aria-label="${escapeHtmlAttribute(copyCodeLabel)}">${escapeHtml(copyCodeLabel)}</button>`,
    '  </div>',
    `  <pre><code${languageClass}>${codeContent.replace(/\n$/, '')}\n</code></pre>`,
    '</div>',
  ].join('')
}

const modelOptions = computed(() =>
  models.value.map((model) => ({
    value: model.id,
    label: model.display_name || model.id,
  }))
)

const composerClass = computed(() => {
  if (isDragActive.value) {
    return 'border-primary-300 bg-[#fff6ed] dark:border-primary-700 dark:bg-primary-950/10'
  }

  return 'border-[#eadfcd] dark:border-dark-600'
})

const attachmentProcessingLabel = computed(() =>
  processingAttachmentName.value
    ? t('chat.processingAttachmentWithName', { name: processingAttachmentName.value })
    : t('chat.processingAttachments')
)

const canSend = computed(() =>
  Boolean(
    activeSession.value &&
    (draftMessage.value.trim() || draftAttachments.value.length > 0) &&
    apiKey.value.trim() &&
    baseUrl.value.trim() &&
    selectedModel.value.trim() &&
    !isStreaming.value &&
    !processingAttachments.value
  )
)

watch(
  () => activeSession.value?.id,
  () => {
    draftMessage.value = ''
    draftAttachments.value = []
    syncActiveSessionConfig()
    void scrollMessagesToBottom()
    void nextTick().then(() => resizeComposerTextarea())
  }
)

watch(
  () => activeSession.value?.messages.length,
  () => {
    void scrollMessagesToBottom()
  }
)

watch(draftMessage, () => {
  void nextTick().then(() => resizeComposerTextarea())
})

watch([baseUrl, apiKey, selectedModel, deepThinkingEnabled], () => {
  if (!activeSession.value) return
  updateSession(activeSession.value.id, {
    baseUrl: baseUrl.value.trim(),
    apiKey: apiKey.value.trim(),
    model: selectedModel.value.trim(),
    deepThinkingEnabled: deepThinkingEnabled.value,
    apiKeyHint: maskKey(apiKey.value),
  })
})

onMounted(async () => {
  await appStore.fetchPublicSettings()

  baseUrl.value = appStore.cachedPublicSettings?.api_base_url || `${window.location.origin}/v1`
  selectedModel.value = selectedModel.value || DEFAULT_CHAT_MODEL

  const draft = consumeDraftConfig()
  if (draft) {
    baseUrl.value = draft.baseUrl || baseUrl.value
    apiKey.value = draft.apiKey || ''
    selectedModel.value = draft.model || selectedModel.value || DEFAULT_CHAT_MODEL
    deepThinkingEnabled.value = draft.deepThinkingEnabled ?? false
  } else if (!activeSession.value && sessions.value[0]) {
    setActiveSession(sessions.value[0].id)
  }

  syncActiveSessionConfig()

  if (apiKey.value.trim() && baseUrl.value.trim()) {
    await loadModels()
  }

  await nextTick()
  resizeComposerTextarea()
})

onUnmounted(() => {
  activeAbortController.value?.abort()
})

function handleCreateSession() {
  createSession({
    title: t('chat.defaultSessionTitle'),
    model: selectedModel.value.trim() || DEFAULT_CHAT_MODEL,
    deepThinkingEnabled: deepThinkingEnabled.value,
    apiKey: apiKey.value,
    baseUrl: baseUrl.value,
    apiKeyHint: maskKey(apiKey.value),
  })
  draftMessage.value = ''
  draftAttachments.value = []
  errorMessage.value = ''
  attachmentFeedbackMessages.value = []
  copiedMessageId.value = null
}

function selectSession(id: string) {
  setActiveSession(id)
  draftMessage.value = ''
  draftAttachments.value = []
  errorMessage.value = ''
  attachmentFeedbackMessages.value = []
  copiedMessageId.value = null
}

function removeSession(id: string) {
  deleteSession(id)
  errorMessage.value = ''
}

function clearCurrentConversation() {
  if (!activeSession.value) return
  updateSession(activeSession.value.id, {
    messages: [],
    title: t('chat.defaultSessionTitle'),
  })
  copiedMessageId.value = null
}

function clearAllSessions() {
  clearLocalSessions()
  copiedMessageId.value = null
}

function handlePrimaryAction() {
  if (isStreaming.value) {
    stopStreaming()
    return
  }

  void sendMessage()
}

function handleComposerEnter() {
  if (isStreaming.value) {
    return
  }

  void sendMessage()
}

function handleModelChange(value: string | number | boolean | null) {
  selectedModel.value = String(value ?? '').trim()
}

async function loadModels() {
  const fallbackModels = createFixedModelOptions()
  if (!apiKey.value.trim() || !baseUrl.value.trim()) {
    models.value = fallbackModels
    selectedModel.value = selectedModel.value || DEFAULT_CHAT_MODEL
    return
  }

  loadingModels.value = true
  errorMessage.value = ''

  try {
    const result = await chatAPI.fetchOpenAIModels(baseUrl.value, apiKey.value)
    const byId = new Map(result.map((item) => [item.id, item]))
    models.value = FIXED_CHAT_MODELS.map((fallbackModel) => ({
      ...fallbackModel,
      ...(byId.get(fallbackModel.id) || {}),
    }))
    if (!selectedModel.value.trim()) {
      selectedModel.value = DEFAULT_CHAT_MODEL
    }
  } catch (error) {
    models.value = fallbackModels
    errorMessage.value = (error as Error).message || t('chat.errors.loadModelsFailed')
  } finally {
    loadingModels.value = false
  }
}

async function sendMessage() {
  if (!canSend.value || !activeSession.value) {
    return
  }

  const prompt = draftMessage.value.trim()
  const sessionId = activeSession.value.id
  const model = selectedModel.value.trim() || DEFAULT_CHAT_MODEL
  const attachments = draftAttachments.value.map((attachment) => ({ ...attachment }))
  errorMessage.value = ''
  attachmentFeedbackMessages.value = []
  copiedMessageId.value = null

  appendMessage(sessionId, 'user', prompt, attachments)
  if (isUntitledSession(activeSession.value.title)) {
    updateSession(sessionId, {
      title: summarizeSessionTitle(buildSessionSeedText(prompt, attachments)),
    })
  }
  draftMessage.value = ''
  draftAttachments.value = []
  updateSession(sessionId, {
    apiKey: apiKey.value.trim(),
    baseUrl: baseUrl.value.trim(),
    model,
    deepThinkingEnabled: deepThinkingEnabled.value,
    apiKeyHint: maskKey(apiKey.value),
  })

  const assistantMessage = appendMessage(sessionId, 'assistant', '')
  if (!assistantMessage) {
    return
  }

  const history: ChatCompletionMessage[] = activeSession.value.messages
    .filter((message) => message.id !== assistantMessage.id)
    .map((message) => buildChatCompletionMessage(message))

  const controller = new AbortController()
  activeAbortController.value = controller
  isStreaming.value = true

  try {
    await chatAPI.streamChatCompletions(
      baseUrl.value,
      apiKey.value,
      {
        model,
        messages: history,
        reasoning: {
          effort: selectedReasoningEffort.value,
        },
      },
      {
        signal: controller.signal,
        onDelta: (delta) => {
          const currentContent = getMessageContent(sessionId, assistantMessage.id)
          updateMessage(sessionId, assistantMessage.id, currentContent + delta)
          void scrollMessagesToBottom()
        },
        onErrorEvent: (message) => {
          errorMessage.value = message
        },
      }
    )
  } catch (error) {
    if ((error as Error).name === 'AbortError') {
      return
    }

    errorMessage.value = (error as Error).message || t('chat.errors.sendFailed')
    const currentContent = getMessageContent(sessionId, assistantMessage.id)
    if (!currentContent.trim()) {
      updateMessage(sessionId, assistantMessage.id, t('chat.errors.emptyAssistantReply'))
    }
  } finally {
    isStreaming.value = false
    activeAbortController.value = null
  }
}

function stopStreaming() {
  activeAbortController.value?.abort()
}

function renderMessage(content: string) {
  const html = marked.parse(content || '', {
    breaks: true,
    gfm: true,
    renderer: messageRenderer,
  }) as string

  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    ADD_TAGS: ['button'],
    ADD_ATTR: ['type', 'data-copy-type', 'data-copy-default-label', 'aria-label'],
  })
}

async function copyMessageContent(message: LocalChatMessage) {
  if (!message.content.trim()) {
    return
  }

  const success = await copyToClipboard(message.content, t('chat.messageCopied'))
  if (!success) {
    return
  }

  copiedMessageId.value = message.id
  window.setTimeout(() => {
    if (copiedMessageId.value === message.id) {
      copiedMessageId.value = null
    }
  }, COPY_FEEDBACK_DURATION_MS)
}

async function handleMarkdownActionClick(event: MouseEvent) {
  const target = event.target as HTMLElement | null
  const actionButton = target?.closest('[data-copy-type="code"]') as HTMLButtonElement | null
  if (!actionButton) {
    return
  }

  event.preventDefault()

  const codeElement = actionButton.closest('.chat-code-block')?.querySelector('pre code')
  const codeText = codeElement?.textContent ?? ''
  if (!codeText.trim()) {
    return
  }

  const success = await copyToClipboard(codeText, t('chat.codeCopied'))
  if (!success) {
    return
  }

  const defaultLabel = actionButton.dataset.copyDefaultLabel || t('chat.copyCode')
  actionButton.textContent = t('chat.copied')
  actionButton.classList.add('is-copied')

  window.setTimeout(() => {
    actionButton.textContent = defaultLabel
    actionButton.classList.remove('is-copied')
  }, COPY_FEEDBACK_DURATION_MS)
}

function messageCardClass(message: LocalChatMessage) {
  if (message.role === 'user') {
    return 'bg-[#2a2d34] text-white dark:bg-[#1b1d23] dark:text-slate-100'
  }

  return 'border border-[#ece4d8] bg-white/97 text-slate-800 dark:border-[#2a3039] dark:bg-[#141920]/96 dark:text-slate-100'
}

function messageBubbleWidthClass(message: LocalChatMessage) {
  if (message.role === 'user') {
    return 'max-w-[88%] md:max-w-[78%] xl:max-w-[72%]'
  }

  return 'max-w-full xl:max-w-[92%]'
}

function messageMetaClass(message: LocalChatMessage) {
  if (message.role === 'user') {
    return 'bg-white/8 text-white/92'
  }

  return 'bg-[#faf4ea] text-[#7a5a2c] dark:bg-[#1d222b] dark:text-[#d9c292]'
}

function messageAttachmentClass(message: LocalChatMessage) {
  if (message.role === 'user') {
    return 'border-white/10 bg-white/6 text-white'
  }

  return 'border-[#eadfcd] bg-[#f8f3eb] text-gray-700 dark:border-[#2a3039] dark:bg-[#191d25]/90 dark:text-gray-200'
}

function formatRelativeTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  return new Intl.DateTimeFormat(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

async function scrollMessagesToBottom() {
  await nextTick()
  if (!messagesContainerRef.value) return
  messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight
}

function resizeComposerTextarea() {
  const textarea = composerTextareaRef.value
  if (!textarea) return

  textarea.style.height = 'auto'
  const nextHeight = Math.min(
    Math.max(textarea.scrollHeight, MIN_COMPOSER_HEIGHT_PX),
    MAX_COMPOSER_HEIGHT_PX
  )
  textarea.style.height = `${nextHeight}px`
  textarea.style.overflowY = textarea.scrollHeight > nextHeight ? 'auto' : 'hidden'
}

function maskKey(value: string): string {
  const trimmed = value.trim()
  if (!trimmed) return ''
  if (trimmed.length <= 8) return trimmed
  return `${trimmed.slice(0, 4)}...${trimmed.slice(-4)}`
}

function syncActiveSessionConfig() {
  if (!activeSession.value) return

  baseUrl.value =
    activeSession.value.baseUrl?.trim() ||
    baseUrl.value ||
    appStore.cachedPublicSettings?.api_base_url ||
    `${window.location.origin}/v1`
  apiKey.value = activeSession.value.apiKey || apiKey.value
  selectedModel.value = activeSession.value.model || selectedModel.value || DEFAULT_CHAT_MODEL
  deepThinkingEnabled.value = activeSession.value.deepThinkingEnabled ?? false
}

function isUntitledSession(title: string): boolean {
  const normalized = title.trim()
  return !normalized || normalized === 'New Chat' || normalized === '新对话' || normalized === t('chat.defaultSessionTitle')
}

function summarizeSessionTitle(input: string): string {
  const normalized = input.replace(/\s+/g, ' ').trim()
  if (!normalized) return t('chat.defaultSessionTitle')
  return normalized.length > 36 ? `${normalized.slice(0, 36)}...` : normalized
}

function sessionMessagePreview(message?: LocalChatMessage): string {
  if (!message) return ''
  if (message.content.trim()) {
    return formatPreviewText(message.content)
  }
  if (message.attachments?.length) {
    return t('chat.attachmentsSummary', {
      count: message.attachments.length,
      name: message.attachments[0]?.name || '',
    })
  }
  return ''
}

function buildSessionSeedText(content: string, attachments: LocalChatAttachment[]): string {
  const trimmed = content.trim()
  if (trimmed) return trimmed
  if (!attachments.length) return ''
  return attachments.map((attachment) => attachment.name).join(', ')
}

function getMessageContent(sessionId: string, messageId: string): string {
  const session = sessions.value.find((item) => item.id === sessionId)
  return session?.messages.find((message) => message.id === messageId)?.content || ''
}

function buildChatCompletionMessage(message: LocalChatMessage): ChatCompletionMessage {
  if (message.role !== 'user') {
    return {
      role: message.role === 'assistant' ? 'assistant' : 'system',
      content: message.content,
    }
  }

  return {
    role: 'user',
    content: buildUserMessageContent(message),
  }
}

function buildUserMessageContent(message: LocalChatMessage): ChatCompletionMessage['content'] {
  if (!message.attachments?.length) {
    return message.content
  }

  const parts: ChatCompletionContentPart[] = []
  if (message.content.trim()) {
    parts.push({
      type: 'text',
      text: message.content.trim(),
    })
  }

  for (const attachment of message.attachments) {
    if (attachment.kind === 'image' && attachment.dataUrl) {
      parts.push({
        type: 'image_url',
        image_url: {
          url: attachment.dataUrl,
          detail: 'auto',
        },
      })
      continue
    }

    if (attachment.kind === 'text' && attachment.textContent) {
      parts.push({
        type: 'text',
        text: formatTextAttachmentForModel(attachment),
      })
    }
  }

  if (!parts.length) {
    return message.content
  }

  return parts
}

function formatTextAttachmentForModel(attachment: LocalChatAttachment): string {
  return [
    `[Attached file: ${attachment.name}]`,
    attachment.textContent?.trim() || '',
    `[End of file: ${attachment.name}]`,
  ].join('\n')
}

function createFixedModelOptions(): OpenAIModel[] {
  return FIXED_CHAT_MODELS.map((model) => ({ ...model }))
}

function openAttachmentPicker() {
  attachmentInputRef.value?.click()
}

function handleComposerDragEnter(event: DragEvent) {
  if (!canAcceptDrag(event)) return
  dragDepth.value += 1
  isDragActive.value = true
}

function handleComposerDragOver(event: DragEvent) {
  if (!canAcceptDrag(event)) return
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'copy'
  }
  isDragActive.value = true
}

function handleComposerDragLeave(event: DragEvent) {
  if (!canAcceptDrag(event)) return
  dragDepth.value = Math.max(0, dragDepth.value - 1)
  if (dragDepth.value === 0) {
    isDragActive.value = false
  }
}

async function handleComposerDrop(event: DragEvent) {
  if (!activeSession.value || processingAttachments.value) {
    resetDragState()
    return
  }

  const files = Array.from(event.dataTransfer?.files || [])
  resetDragState()

  if (!files.length) {
    return
  }

  await processIncomingFiles(files)
}

async function handleAttachmentSelection(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  await processIncomingFiles(files)
}

async function processIncomingFiles(files: File[]) {
  if (!files.length) {
    return
  }

  processingAttachments.value = true
  processingAttachmentName.value = ''
  errorMessage.value = ''
  attachmentFeedbackMessages.value = []
  const nextAttachments = [...draftAttachments.value]
  const feedbackMessages: AttachmentFeedback[] = []

  try {
    for (const file of files) {
      if (nextAttachments.length >= MAX_ATTACHMENTS) {
        feedbackMessages.push(createAttachmentFeedback(
          'error',
          file.name,
          t('chat.errors.tooManyAttachments', { count: MAX_ATTACHMENTS })
        ))
        break
      }

      processingAttachmentName.value = file.name

      try {
        const attachment = await createAttachmentFromFile(file)
        nextAttachments.push(attachment)
      } catch (error) {
        feedbackMessages.push(createAttachmentFeedback(
          'error',
          file.name,
          (error as Error).message || t('chat.errors.attachmentReadFailed')
        ))
      }
    }
  } catch (error) {
    errorMessage.value = (error as Error).message || t('chat.errors.attachmentReadFailed')
  } finally {
    draftAttachments.value = nextAttachments
    attachmentFeedbackMessages.value = feedbackMessages
    processingAttachments.value = false
    processingAttachmentName.value = ''
  }
}

async function createAttachmentFromFile(
  file: File,
  preferredKind?: LocalChatAttachment['kind']
): Promise<LocalChatAttachment> {
  const kind = preferredKind || inferAttachmentKind(file)
  if (kind === 'image') {
    return createImageAttachment(file)
  }

  return createTextAttachment(file)
}

async function createImageAttachment(file: File): Promise<LocalChatAttachment> {
  if (!file.type.startsWith('image/')) {
    throw new Error(t('chat.errors.unsupportedFileType'))
  }
  if (file.size > MAX_IMAGE_SIZE_BYTES) {
    throw new Error(t('chat.errors.imageTooLarge'))
  }

  return {
    id: createAttachmentId(file),
    kind: 'image',
    sourceType: 'image',
    name: file.name,
    mimeType: file.type || 'image/*',
    size: file.size,
    dataUrl: await readFileAsDataUrl(file),
  }
}

async function createTextAttachment(file: File): Promise<LocalChatAttachment> {
  if (isPdfFile(file)) {
    return createPdfAttachment(file)
  }

  if (isDocxFile(file)) {
    return createDocxAttachment(file)
  }

  if (!isPlainTextLikeFile(file)) {
    throw new Error(t('chat.errors.unsupportedFileType'))
  }
  if (file.size > MAX_TEXT_FILE_SIZE_BYTES) {
    throw new Error(t('chat.errors.fileTooLarge'))
  }

  return {
    id: createAttachmentId(file),
    kind: 'text',
    sourceType: 'text',
    name: file.name,
    mimeType: file.type || 'text/plain',
    size: file.size,
    textContent: await file.text(),
  }
}

async function createPdfAttachment(file: File): Promise<LocalChatAttachment> {
  if (file.size > MAX_PDF_FILE_SIZE_BYTES) {
    throw new Error(t('chat.errors.pdfTooLarge'))
  }

  try {
    const arrayBuffer = await file.arrayBuffer()
    const pdf = await pdfjsLib.getDocument({
      data: arrayBuffer,
      useWorkerFetch: false,
    }).promise

    const pages: string[] = []
    for (let pageNumber = 1; pageNumber <= pdf.numPages; pageNumber += 1) {
      const page = await pdf.getPage(pageNumber)
      const textContent = await page.getTextContent()
      const pageText = textContent.items
        .map((item) => ('str' in item ? (item as TextItem).str || '' : ''))
        .join(' ')
        .replace(/\s+/g, ' ')
        .trim()

      if (pageText) {
        pages.push(`[Page ${pageNumber}] ${pageText}`)
      }
    }

    const textContent = pages.join('\n\n').trim()
    if (!textContent) {
      throw new Error(t('chat.errors.pdfNoExtractableText'))
    }

    return {
      id: createAttachmentId(file),
      kind: 'text',
      sourceType: 'pdf',
      name: file.name,
      mimeType: file.type || 'application/pdf',
      size: file.size,
      pageCount: pdf.numPages,
      textContent,
    }
  } catch (error) {
    throw new Error(formatAttachmentProcessingError('pdf', error))
  }
}

async function createDocxAttachment(file: File): Promise<LocalChatAttachment> {
  if (file.size > MAX_DOCX_FILE_SIZE_BYTES) {
    throw new Error(t('chat.errors.docxTooLarge'))
  }

  try {
    const arrayBuffer = await file.arrayBuffer()
    const result = await mammoth.extractRawText({ arrayBuffer })
    const textContent = (result.value || '').replace(/\n{3,}/g, '\n\n').trim()

    if (!textContent) {
      throw new Error(t('chat.errors.docxNoExtractableText'))
    }

    return {
      id: createAttachmentId(file),
      kind: 'text',
      sourceType: 'docx',
      name: file.name,
      mimeType: file.type || 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      size: file.size,
      textContent,
    }
  } catch (error) {
    throw new Error(formatAttachmentProcessingError('docx', error))
  }
}

function removeDraftAttachment(id: string) {
  draftAttachments.value = draftAttachments.value.filter((attachment) => attachment.id !== id)
}

function isPlainTextLikeFile(file: File): boolean {
  if (file.type.startsWith('text/')) {
    return true
  }

  if (file.type === 'application/json' || file.type === 'application/xml') {
    return true
  }

  const extension = file.name.split('.').pop()?.toLowerCase() || ''
  return SUPPORTED_TEXT_FILE_EXTENSIONS.has(extension)
}

function isPdfFile(file: File): boolean {
  return file.type === 'application/pdf' || file.name.toLowerCase().endsWith('.pdf')
}

function isDocxFile(file: File): boolean {
  return file.type === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' || file.name.toLowerCase().endsWith('.docx')
}

function inferAttachmentKind(file: File): LocalChatAttachment['kind'] {
  if (file.type.startsWith('image/')) {
    return 'image'
  }

  return 'text'
}

function createAttachmentId(file: File): string {
  return `attachment_${file.name}_${file.size}_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
}

function readFileAsDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('Failed to read file'))
    reader.readAsDataURL(file)
  })
}

function formatFileSize(value: number): string {
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  return `${(value / (1024 * 1024)).toFixed(1)} MB`
}

function attachmentIconName(attachment: LocalChatAttachment): 'document' | 'book' | 'clipboard' {
  if (attachment.sourceType === 'pdf') {
    return 'book'
  }

  if (attachment.sourceType === 'docx') {
    return 'clipboard'
  }

  return 'document'
}

function formatAttachmentMeta(attachment: LocalChatAttachment): string {
  const parts = [formatFileSize(attachment.size)]

  if (attachment.sourceType === 'pdf' && attachment.pageCount) {
    parts.push(t('chat.pdfPageCount', { count: attachment.pageCount }))
  } else if (attachment.sourceType === 'docx') {
    parts.push(t('chat.wordAttachmentLabel'))
  } else if (attachment.kind === 'image') {
    parts.push(t('chat.imageAttachmentLabel'))
  } else {
    parts.push(t('chat.textAttachmentLabel'))
  }

  return parts.join(' · ')
}

function canAcceptDrag(event: DragEvent): boolean {
  if (!activeSession.value || processingAttachments.value) {
    return false
  }

  const types = Array.from(event.dataTransfer?.types || [])
  return types.includes('Files')
}

function resetDragState() {
  dragDepth.value = 0
  isDragActive.value = false
}

function createAttachmentFeedback(
  kind: AttachmentFeedback['kind'],
  fileName: string,
  message: string
): AttachmentFeedback {
  return {
    id: `feedback_${fileName}_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
    kind,
    fileName,
    message,
  }
}

function formatAttachmentProcessingError(
  type: 'pdf' | 'docx',
  error: unknown
): string {
  const rawMessage = (error as Error)?.message?.trim()

  if (!rawMessage) {
    return type === 'pdf'
      ? t('chat.errors.pdfReadFailed')
      : t('chat.errors.docxReadFailed')
  }

  if (rawMessage === t('chat.errors.pdfNoExtractableText') || rawMessage === t('chat.errors.docxNoExtractableText')) {
    return rawMessage
  }

  if (type === 'pdf') {
    return t('chat.errors.pdfReadFailedWithReason', {
      reason: shortenErrorMessage(rawMessage),
    })
  }

  return t('chat.errors.docxReadFailedWithReason', {
    reason: shortenErrorMessage(rawMessage),
  })
}

function shortenErrorMessage(input: string): string {
  const normalized = input.replace(/\s+/g, ' ').trim()
  if (!normalized) {
    return t('chat.errors.unknownReason')
  }

  return normalized.length > PDF_PAGE_PREVIEW_LIMIT
    ? `${normalized.slice(0, PDF_PAGE_PREVIEW_LIMIT)}...`
    : normalized
}

function normalizeCodeLanguage(value?: string | null): string {
  return (value || '').trim().split(/\s+/)[0]?.toLowerCase() || ''
}

function formatCodeLanguageLabel(value: string): string {
  if (!value) return 'TEXT'

  if (value === 'bash' || value === 'shell' || value === 'sh' || value === 'zsh') {
    return 'SHELL'
  }

  if (value === 'ts') return 'TYPESCRIPT'
  if (value === 'js') return 'JAVASCRIPT'
  if (value === 'py') return 'PYTHON'
  if (value === 'md') return 'MARKDOWN'
  if (value === 'json') return 'JSON'

  return value.toUpperCase()
}

function formatPreviewText(input: string): string {
  const normalized = input
    .replace(/```[\s\S]*?```/g, '[code]')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/^#{1,6}\s+/gm, '')
    .replace(/^\s*>\s?/gm, '')
    .replace(/[*_~]/g, '')
    .replace(/\n+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()

  if (!normalized) {
    return ''
  }

  return normalized.length > 96 ? `${normalized.slice(0, 96)}...` : normalized
}

function escapeHtml(input: string): string {
  return input
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function escapeHtmlAttribute(input: string): string {
  return escapeHtml(input).replace(/`/g, '&#96;')
}

// Keep an escape hatch for future navigation-based prefill without exposing the key in the URL.
function persistCurrentDraft() {
  saveDraftConfig({
    apiKey: apiKey.value,
    baseUrl: baseUrl.value,
    model: selectedModel.value,
    sessionTitle: activeSession.value?.title || t('chat.defaultSessionTitle'),
    deepThinkingEnabled: deepThinkingEnabled.value,
  })
}

watch([baseUrl, apiKey, selectedModel, deepThinkingEnabled], persistCurrentDraft)
watch(() => router.currentRoute.value.fullPath, persistCurrentDraft)
</script>

<style scoped>
.input {
  @apply w-full rounded-2xl border border-[#e5dac7] px-3 py-2.5 text-sm text-gray-900 outline-none transition-all placeholder:text-gray-400 focus:border-[#d3b991] focus:ring-2 focus:ring-[#efe3d1] dark:border-[#2d323c] dark:bg-[#141920] dark:text-white dark:placeholder:text-gray-500 dark:focus:border-[#9a7640] dark:focus:ring-[#3a2b12];
  background-color: rgba(255, 255, 255, 0.92);
}

.chat-action-button {
  @apply inline-flex items-center gap-1.5 rounded-full border border-[#e7dccb] px-3 py-1.5 text-xs font-medium text-gray-600 transition-colors hover:border-[#d9c4a1] hover:text-[#7b5a2d] dark:border-dark-600 dark:text-gray-300 dark:hover:border-[#8d6b39] dark:hover:text-[#dcc391];
  background-color: rgba(255, 255, 255, 0.9);
}

.dark .chat-action-button {
  background-color: rgba(26, 30, 37, 0.88);
}

.chat-markdown {
  @apply break-words text-[15px] leading-7;
}

.chat-markdown :deep(h1),
.chat-markdown :deep(h2),
.chat-markdown :deep(h3),
.chat-markdown :deep(h4) {
  @apply mb-3 mt-6 font-semibold tracking-tight;
}

.chat-markdown :deep(h1) {
  @apply text-2xl;
}

.chat-markdown :deep(h2) {
  @apply text-xl;
}

.chat-markdown :deep(h3) {
  @apply text-lg;
}

.chat-markdown :deep(h4) {
  @apply text-base;
}

.chat-markdown :deep(p),
.chat-markdown :deep(ul),
.chat-markdown :deep(ol),
.chat-markdown :deep(blockquote),
.chat-markdown :deep(table),
.chat-markdown :deep(.chat-code-block) {
  @apply mb-4;
}

.chat-markdown :deep(ul) {
  @apply list-disc pl-6;
}

.chat-markdown :deep(ol) {
  @apply list-decimal pl-6;
}

.chat-markdown :deep(li) {
  @apply mb-1.5;
}

.chat-markdown :deep(li > p) {
  margin-bottom: 0.25rem;
}

.chat-markdown :deep(a) {
  @apply font-medium underline decoration-1 underline-offset-4;
}

.chat-markdown :deep(blockquote) {
  @apply rounded-2xl border-l-4 px-4 py-3;
}

.chat-markdown :deep(table) {
  @apply w-full overflow-hidden rounded-2xl border border-collapse;
}

.chat-markdown :deep(th),
.chat-markdown :deep(td) {
  @apply border px-3 py-2 text-left align-top;
}

.chat-markdown :deep(th) {
  @apply text-sm font-semibold;
}

.chat-markdown :deep(hr) {
  @apply my-6 border-0 border-t;
}

.chat-markdown :deep(code) {
  @apply break-words rounded-lg px-1.5 py-0.5 font-mono text-[0.92em];
}

.chat-markdown :deep(pre) {
  @apply m-0 overflow-x-auto bg-transparent p-4 font-mono text-[13px] leading-6;
}

.chat-markdown :deep(pre code) {
  @apply rounded-none bg-transparent p-0 text-inherit;
}

.chat-markdown :deep(.chat-code-block) {
  @apply overflow-hidden rounded-[1.15rem] border;
}

.chat-markdown :deep(.chat-code-header) {
  @apply flex items-center justify-between gap-3 px-4 py-2.5 text-[11px] font-semibold uppercase tracking-[0.24em];
}

.chat-markdown :deep(.chat-code-label) {
  @apply truncate;
}

.chat-markdown :deep(.chat-code-copy) {
  @apply rounded-full border px-3 py-1 text-[11px] font-medium tracking-normal transition-colors;
}

.chat-markdown :deep(.chat-code-copy.is-copied) {
  @apply border-primary-300 bg-primary-100 text-primary-700 dark:border-primary-700 dark:text-primary-200;
}

.dark .chat-markdown :deep(.chat-code-copy.is-copied) {
  background: rgba(88, 28, 135, 0.28);
}

.chat-markdown :deep(p:last-child) {
  margin-bottom: 0;
}

.chat-markdown-assistant {
  @apply text-slate-700 dark:text-slate-100;
}

.chat-markdown-assistant :deep(a) {
  @apply text-[#7b5a2d] hover:text-[#5f431d] dark:text-[#dcc391] dark:hover:text-[#f0d7a5];
}

.chat-markdown-assistant :deep(blockquote) {
  @apply border-[#d9c4a1] bg-[#faf4ea] text-slate-600 dark:border-[#7a5a29] dark:bg-[#1b2029] dark:text-slate-200;
}

.chat-markdown-assistant :deep(table),
.chat-markdown-assistant :deep(th),
.chat-markdown-assistant :deep(td) {
  @apply border-[#eadfcd] dark:border-dark-600;
}

.chat-markdown-assistant :deep(th) {
  @apply bg-[#f8f2e9] text-slate-700 dark:bg-[#1b2029] dark:text-slate-100;
}

.chat-markdown-assistant :deep(hr) {
  @apply border-[#eadfcd] dark:border-dark-600;
}

.chat-markdown-assistant :deep(:not(pre) > code) {
  @apply bg-[#f7efe4] text-[#8b6230] dark:bg-[#1b2029] dark:text-[#e1c998];
}

.chat-markdown-assistant :deep(.chat-code-block) {
  @apply border-[#e6d9c6] bg-[#181b22] text-slate-100 dark:border-[#2b3038] dark:bg-[#0f1217];
}

.chat-markdown-assistant :deep(.chat-code-header) {
  @apply border-b text-slate-300;
  border-color: rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
}

.chat-markdown-assistant :deep(.chat-code-copy) {
  @apply text-slate-200;
  border-color: rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.06);
}

.chat-markdown-assistant :deep(.chat-code-copy:hover) {
  border-color: rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.12);
}

.chat-markdown-user {
  @apply text-white;
  opacity: 0.92;
}

.chat-markdown-user :deep(a) {
  @apply text-white underline hover:text-white;
  text-decoration-color: rgba(255, 255, 255, 0.4);
}

.chat-markdown-user :deep(blockquote) {
  border-color: rgba(255, 255, 255, 0.18);
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.86);
}

.chat-markdown-user :deep(table),
.chat-markdown-user :deep(th),
.chat-markdown-user :deep(td) {
  border-color: rgba(255, 255, 255, 0.12);
}

.chat-markdown-user :deep(th) {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.92);
}

.chat-markdown-user :deep(hr) {
  border-color: rgba(255, 255, 255, 0.12);
}

.chat-markdown-user :deep(:not(pre) > code) {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.96);
}

.chat-markdown-user :deep(.chat-code-block) {
  border-color: rgba(255, 255, 255, 0.1);
  background: rgba(3, 7, 18, 0.35);
  color: rgba(255, 255, 255, 0.95);
}

.chat-markdown-user :deep(.chat-code-header) {
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.72);
}

.chat-markdown-user :deep(.chat-code-copy) {
  border-color: rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.92);
}

.chat-markdown-user :deep(.chat-code-copy:hover) {
  background: rgba(255, 255, 255, 0.14);
}
</style>
