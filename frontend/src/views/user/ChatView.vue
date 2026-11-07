<template>
  <AppLayout>
    <div class="chat-page relative -m-2 overflow-hidden rounded-2xl border border-[#e8e0d3] bg-[#fcfaf6] p-1.5 dark:border-[#2a2f38] dark:bg-[#11161d] md:-m-4 md:p-2 lg:-m-6" :class="{ 'chat-page-focus': focusMode }">

      <div class="chat-workspace relative grid min-h-0 gap-2" :class="{ 'has-history': historyOpen }">
        <button v-if="historyOpen" class="absolute inset-0 z-20 bg-black/20 lg:hidden" :aria-label="t('chat.hideHistory')" @click="historyOpen = false"></button>
        <aside v-show="historyOpen" data-testid="chat-history" class="chat-session-sidebar z-30 flex min-h-0 flex-col overflow-hidden rounded-2xl border border-[#efe8dd] bg-white dark:border-[#242933] dark:bg-[#10141a]">
          <div class="border-b border-[#efe8dd] p-3 dark:border-dark-700">
            <div class="flex flex-col gap-2">
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('chat.title') }}
                </h2>
                <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">
                  {{ t('chat.sidebarDescription') }}
                </p>
              </div>
              <button class="btn btn-primary self-start !px-3 !py-1.5 !text-sm" :disabled="isStreaming" @click="handleCreateSession">
                <Icon name="plus" size="sm" class="mr-1.5" />
                {{ t('chat.newChat') }}
              </button>
            </div>
          </div>

          <div class="min-h-0 flex-1 overflow-y-auto p-2">
            <div v-if="sessions.length === 0" class="rounded-2xl border border-dashed border-[#e5dccf] bg-[#fbf8f3] p-5 text-center text-sm text-gray-500 dark:border-dark-600 dark:bg-dark-900/40 dark:text-gray-400">
              {{ t('chat.noSessions') }}
            </div>

            <div v-else class="space-y-2.5">
              <div
                v-for="session in sessions"
                :key="session.id"
                class="group flex w-full items-start gap-2 rounded-xl border p-2.5 transition-all duration-200"
                :class="[
                  session.id === activeSession?.id
                    ? 'border-[#d8c3a1] bg-[#fbf6ee] shadow-[0_12px_28px_rgba(124,91,45,0.10)] dark:border-[#8a6a35] dark:bg-[#2a2117]'
                    : 'border-transparent bg-white/75 hover:border-[#eadfcd] hover:bg-white dark:bg-dark-800/80 dark:hover:border-dark-500 dark:hover:bg-dark-800',
                  isStreaming ? 'cursor-not-allowed opacity-70' : ''
                ]"
              >
                <button
                  type="button"
                  class="min-w-0 flex-1 text-left"
                  role="button"
                  :aria-pressed="session.id === activeSession?.id"
                  :disabled="isStreaming"
                  @click="selectSession(session.id)"
                >
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
                </button>
                <button
                  type="button"
                  class="shrink-0 rounded-xl p-1 text-gray-400 opacity-0 transition-all hover:bg-red-50 hover:text-red-500 group-hover:opacity-100 dark:hover:bg-red-900/20"
                  :title="t('common.delete')"
                  :disabled="isStreaming"
                  @click="removeSession(session.id)"
                >
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </div>
          </div>
        </aside>

        <section class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-2xl bg-[#fcfaf6] dark:bg-[#11161d]">
          <div class="chat-toolbar border-b border-[#efe8dd] px-3 py-2.5 dark:border-dark-700">
            <div class="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <h1 class="w-full truncate text-base font-semibold text-gray-900 dark:text-white lg:w-auto lg:max-w-full" :title="activeSession?.title || t('chat.emptyTitle')">
                    {{ activeSession?.title || t('chat.emptyTitle') }}
                  </h1>
                  <span
                    v-if="activeSession?.model"
                    class="hidden max-w-48 truncate rounded-full border border-[#eadfcd] bg-white px-2 py-0.5 text-xs font-medium text-gray-600 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 lg:inline-block"
                  >
                    {{ activeSession.model }}
                  </span>
                </div>
              </div>

              <div class="flex shrink-0 flex-wrap items-center gap-1 sm:gap-1.5">
                <button type="button" class="chat-action-button" :aria-expanded="historyOpen" @click="historyOpen = !historyOpen">
                  <Icon name="menu" size="sm" />{{ t(historyOpen ? 'chat.hideHistory' : 'chat.showHistory') }}
                </button>
                <button type="button" class="chat-action-button" :aria-expanded="connectionSettingsOpen" @click="connectionSettingsOpen = !connectionSettingsOpen">
                  <Icon name="cog" size="sm" />{{ t('chat.connectionSettings') }}
                </button>
                <button type="button" class="chat-action-button" :aria-pressed="focusMode" @click="focusMode = !focusMode">
                  <Icon name="grid" size="sm" />{{ t(focusMode ? 'chat.exitFocus' : 'chat.focusMode') }}
                </button>
                <button class="btn btn-secondary" :disabled="!activeSession || isStreaming || !activeSession.messages.length" @click="clearCurrentConversation">
                  {{ t('chat.clearMessages') }}
                </button>
                <button class="btn btn-secondary" :disabled="!sessions.length || isStreaming" @click="clearAllSessions">
                  {{ t('chat.clearAll') }}
                </button>
              </div>
            </div>
          </div>

          <div v-show="connectionSettingsOpen" class="grid gap-3 border-b border-[#efe8dd] px-5 py-3 dark:border-dark-700 sm:grid-cols-2" data-testid="chat-connection-settings">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-300">
              {{ t('chat.baseUrl') }}
              <input v-model="baseUrl" type="text" class="input mt-1.5" :placeholder="t('chat.baseUrlPlaceholder')" :disabled="isStreaming" />
            </label>
            <label class="text-xs font-medium text-gray-600 dark:text-gray-300">
              {{ t('chat.apiKey') }}
              <input v-model="apiKey" type="password" autocomplete="off" class="input mt-1.5" :placeholder="t('chat.apiKeyPlaceholder')" :disabled="isStreaming" />
            </label>
          </div>

          <div
            ref="messagesContainerRef"
            data-testid="chat-messages"
            class="min-h-0 flex-1 overflow-y-auto px-2 py-3 md:px-3"
            @scroll.passive="handleMessagesScroll"
          >
            <div v-if="!activeSession || !activeSession.messages.length" class="flex min-h-full items-center justify-center">
              <div class="max-w-xl rounded-[2rem] border border-dashed border-[#dfd3c0] bg-white/75 px-5 py-6 text-center shadow-[0_18px_48px_rgba(15,23,42,0.06)] dark:border-dark-600 dark:bg-dark-900/55 sm:px-8 sm:py-10">
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
                <div class="mt-6 grid grid-cols-2 gap-2 text-left">
                  <button v-for="prompt in starterPrompts" :key="prompt.key" type="button" :data-testid="`chat-starter-${prompt.key}`" class="rounded-xl border border-[#eadfcd] bg-white px-3 py-3 text-sm text-gray-700 transition-colors hover:border-primary-300 hover:bg-primary-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-primary-700" @click="useStarterPrompt(prompt.key)">
                    <Icon :name="prompt.icon" size="sm" class="mb-2 text-primary-500" />{{ t(`chat.starters.${prompt.key}.title`) }}
                  </button>
                </div>
              </div>
            </div>

            <div v-else class="mx-auto flex w-full max-w-[1440px] flex-col gap-4">
              <div
                v-for="message in activeSession.messages"
                :key="message.id"
                class="flex"
                :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
              >
                <article
                  class="group w-full rounded-2xl px-3 py-3 md:px-4"
                  :class="[messageBubbleWidthClass(message), messageCardClass(message)]"
                >
                  <header class="mb-2 flex flex-wrap items-start justify-between gap-2">
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
                      class="flex items-center gap-2 opacity-100 transition-opacity md:opacity-0 md:group-hover:opacity-100"
                    >
                      <button v-if="message.role === 'user'" type="button" class="chat-action-button" :disabled="isStreaming || processingAttachments" :title="t('chat.newRequestHint')" @click="beginEditMessage(message)">
                        <Icon name="edit" size="sm" />{{ t('chat.editMessage') }}
                      </button>
                      <button v-if="message.role === 'assistant'" type="button" class="chat-action-button" :disabled="!canSubmitRequest" :title="t('chat.newRequestHint')" @click="regenerateMessage(message.id)">
                        <Icon name="refresh" size="sm" />{{ t(message.response?.status === 'failed' ? 'chat.retry' : 'chat.regenerate') }}
                      </button>
                      <button
                        v-if="message.role === 'assistant' && message.content.trim()"
                        type="button"
                        class="chat-action-button"
                        @click="copyMessageContent(message)"
                      >
                        <Icon :name="copiedMessageId === message.id ? 'check' : 'copy'" size="sm" />
                        <span>{{ copiedMessageId === message.id ? t('chat.copied') : t('chat.copyMessage') }}</span>
                      </button>
                    </div>
                  </header>

                  <div v-if="message.response" class="mb-3 flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400" role="status">
                    <Icon :name="message.response.status === 'completed' ? 'checkCircle' : message.response.status === 'failed' ? 'exclamationCircle' : 'clock'" size="sm" />
                    <span>{{ t(`chat.status.${message.response.status}`) }}</span>
                    <span>{{ t('chat.elapsed', { seconds: ((isStreaming && streamingMessageId === message.id ? elapsedMs : message.response.elapsedMs || 0) / 1000).toFixed(1) }) }}</span>
                  </div>
                  <details v-if="message.response?.reasoningSummary" class="mb-3 rounded-xl border border-[#eadfcd] px-3 py-2 text-sm dark:border-dark-600">
                    <summary class="cursor-pointer text-gray-500 dark:text-gray-400">{{ t('chat.reasoningSummary') }}</summary>
                    <p class="mt-2 whitespace-pre-wrap text-gray-700 dark:text-gray-200">{{ message.response.reasoningSummary }}</p>
                  </details>

                  <div
                    v-if="message.content.trim()"
                    class="chat-markdown"
                    :class="message.role === 'user' ? 'chat-markdown-user' : 'chat-markdown-assistant'"
                    @click="handleMarkdownActionClick"
                    v-html="renderMessage(message)"
                  ></div>

                  <div v-if="message.response?.sources?.length" class="mt-4 flex flex-wrap items-center gap-2" data-testid="chat-sources">
                    <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('chat.sources') }}</span>
                    <a v-for="source in message.response.sources" :key="source.url" :href="source.url" :title="source.url" target="_blank" rel="noopener noreferrer" class="max-w-64 truncate rounded-full border border-[#eadfcd] px-3 py-1 text-xs text-primary-700 hover:bg-primary-50 dark:border-dark-600 dark:text-primary-300 dark:hover:bg-dark-800">{{ source.title }}</a>
                  </div>

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

            </div>

            <button
              v-if="showScrollToLatest"
              type="button"
              data-testid="chat-scroll-to-latest"
              class="sticky bottom-2 ml-auto mt-4 flex h-10 w-fit items-center gap-2 rounded-full border border-[#ded2c1] bg-white/95 px-3.5 text-sm font-medium text-gray-700 shadow-[0_10px_28px_rgba(15,23,42,0.14)] backdrop-blur transition-colors hover:border-primary-300 hover:text-primary-700 dark:border-dark-600 dark:bg-dark-800/95 dark:text-gray-200 dark:hover:border-primary-700 dark:hover:text-primary-300"
              @click="scrollMessagesToBottom(true)"
            >
              <Icon name="chevronDown" size="sm" />
              {{ t('chat.scrollToLatest') }}
            </button>
          </div>

          <div class="chat-composer shrink-0 border-t border-[#efe8dd] px-2 py-2 dark:border-dark-700 md:px-3">
            <div class="mx-auto w-full max-w-[1440px]">
              <div class="mb-2 grid grid-cols-1 items-center gap-x-3 gap-y-1.5 sm:grid-cols-[minmax(0,1fr)_auto]">
                <div class="flex min-w-0 flex-1 items-center gap-2">
                  <span class="shrink-0 text-xs font-medium text-gray-600 dark:text-gray-300">{{ t('chat.model') }}</span>
                  <Select class="min-w-0 flex-1" data-testid="chat-model-select" :model-value="selectedModel || null" :options="modelOptions" :placeholder="t('chat.modelPlaceholder')" :aria-label="t('chat.model')" :disabled="isStreaming || loadingModels" searchable @update:model-value="handleModelChange" />
                  <button type="button" class="chat-action-button shrink-0" :disabled="loadingModels || isStreaming" @click="loadModels">{{ loadingModels ? t('chat.loadingModels') : t('chat.refreshModels') }}</button>
                </div>
                <div class="flex items-center gap-2">
                  <label for="chat-reasoning-effort" class="shrink-0 text-xs font-medium text-gray-600 dark:text-gray-300" :title="t('chat.reasoningEffortHint')">{{ t('chat.reasoningEffort') }}</label>
                  <Select id="chat-reasoning-effort" class="w-32" data-testid="chat-reasoning-effort-select" :model-value="selectedReasoningEffort" :options="reasoningEffortOptions" :aria-label="t('chat.reasoningEffort')" aria-describedby="chat-reasoning-effort-hint" :disabled="isStreaming" @update:model-value="handleReasoningEffortChange" />
                  <span id="chat-reasoning-effort-hint" class="sr-only">{{ t('chat.reasoningEffortHint') }}</span>
                </div>
              </div>
              <div v-if="errorMessage" class="mb-2.5 rounded-[1.15rem] border border-red-200 bg-red-50 px-3.5 py-2.5 text-[13px] text-red-700 dark:border-red-900/60 dark:bg-red-900/20 dark:text-red-300">
                {{ errorMessage }}
              </div>

              <form
                class="rounded-2xl border border-[#eadfcd] bg-white/94 p-2 shadow-[0_8px_24px_rgba(15,23,42,0.05)] transition-colors dark:border-[#2b3039] dark:bg-[#10141a]/94"
                :class="composerClass"
                @submit.prevent="handlePrimaryAction"
                @dragenter.prevent="handleComposerDragEnter"
                @dragover.prevent="handleComposerDragOver"
                @dragleave.prevent="handleComposerDragLeave"
                @drop.prevent="handleComposerDrop"
              >
                <div v-if="editingMessageId" class="mb-3 flex items-start justify-between gap-3 rounded-xl bg-primary-50 px-3 py-2 text-xs text-primary-700 dark:bg-primary-950/30 dark:text-primary-200">
                  <span>{{ t(composerMode === 'retry' ? 'chat.retryHint' : 'chat.editHint') }}</span>
                  <button type="button" class="shrink-0 underline" @click="cancelEditMessage">{{ t('common.cancel') }}</button>
                </div>
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
                  rows="1"
                  class="block min-h-[40px] w-full resize-none overflow-y-auto border-0 bg-transparent px-2 py-2 text-[15px] leading-6 text-gray-900 outline-none placeholder:text-gray-400 dark:text-white dark:placeholder:text-gray-500"
                  :placeholder="t('chat.inputPlaceholder')"
                  :disabled="processingAttachments"
                  @input="resizeComposerTextarea"
                  @keydown.enter.exact="handleComposerEnter"
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

                <div class="mt-1.5 flex flex-wrap items-center justify-between gap-2 border-t border-[#efe8dd] px-1 pt-1.5 dark:border-dark-700">
                  <div class="flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                    <button
                      type="button"
                      class="inline-flex items-center gap-1.5 rounded-full border border-[#eadfcd] bg-[#faf6ef] px-3.5 py-1.5 transition-colors hover:border-[#d9c4a1] hover:text-[#7b5a2d] dark:border-[#343944] dark:bg-[#181d24] dark:hover:border-[#8d6b39] dark:hover:text-[#dcc391]"
                      :disabled="processingAttachments"
                      @click="openAttachmentPicker"
                    >
                      <Icon name="upload" size="sm" />
                      {{ t('chat.uploadAttachment') }}
                    </button>
                    <span class="hidden leading-5 sm:inline">{{ processingAttachments ? attachmentProcessingLabel : t('chat.supportedFileHint') }}</span>
                  </div>

                  <div class="flex items-center gap-2">
                    <button
                      type="submit"
                      class="btn btn-primary !px-3 !py-1.5 !text-sm"
                      :disabled="isStreaming ? !activeAbortController : !canSend"
                    >
                      <Icon :name="isStreaming ? 'x' : 'arrowUp'" size="sm" class="mr-1.5" />
                      {{ isStreaming ? t('chat.stop') : t('chat.send') }}
                    </button>
                  </div>
                </div>
                <p class="mt-2 px-2 text-xs leading-5 text-gray-500 dark:text-gray-400 sm:hidden">{{ processingAttachments ? attachmentProcessingLabel : t('chat.supportedFileHint') }}</p>
              </form>
              <p v-if="setupHint" class="mt-2 text-xs text-gray-500 dark:text-gray-400">{{ setupHint }}</p>
            </div>
          </div>
        </section>
      </div>
    </div>
    <BaseDialog :show="Boolean(codePreview)" :title="t('chat.codePreview')" width="extra-wide" :z-index="80" @close="codePreview = null">
      <div v-if="codePreview" data-testid="chat-code-preview">
        <div class="mb-3 flex items-center justify-between gap-3">
          <span class="text-xs font-semibold text-gray-500">{{ formatCodeLanguageLabel(codePreview.language) }}</span>
          <button type="button" class="chat-action-button" @click="downloadCode(codePreview.text, codePreview.language)"><Icon name="download" size="sm" />{{ t('chat.downloadCode') }}</button>
        </div>
        <pre class="max-h-[65dvh] overflow-auto rounded-xl bg-[#181b22] p-4 text-sm leading-6 text-slate-100"><code v-html="highlightCode(codePreview.text, codePreview.language)"></code></pre>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import hljs from 'highlight.js/lib/core'
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import python from 'highlight.js/lib/languages/python'
import bash from 'highlight.js/lib/languages/bash'
import json from 'highlight.js/lib/languages/json'
import go from 'highlight.js/lib/languages/go'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import yaml from 'highlight.js/lib/languages/yaml'
import sql from 'highlight.js/lib/languages/sql'
import 'highlight.js/styles/github-dark.css'
import { saveAs } from 'file-saver'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { chatAPI } from '@/api/chat'
import { useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import { useLocalChat } from '@/composables/useLocalChat'
import type { OpenAIModel } from '@/types'
import type {
  ChatCompletionContentPart,
  ChatCompletionRequest,
  ChatCompletionMessage,
  ResponsesContentPart,
  ResponsesInputItem,
  ResponsesStreamEvent,
  ResponsesAnnotation,
} from '@/api/chat'
import type { ChatReasoningEffort, ChatResponseStatus, LocalChatAttachment, LocalChatMessage, LocalChatResponse } from '@/composables/useLocalChat'
import type { TextItem } from 'pdfjs-dist/types/src/display/api'
import pdfWorkerUrl from 'pdfjs-dist/build/pdf.worker.mjs?url'

const REASONING_EFFORTS: ChatReasoningEffort[] = ['low', 'medium', 'high', 'xhigh', 'max']
for (const [name, grammar] of Object.entries({ javascript, typescript, python, bash, json, go, xml, css, yaml, sql })) {
  hljs.registerLanguage(name, grammar)
}
const reasoningEffortOptions = REASONING_EFFORTS.map((effort) => ({ value: effort, label: effort }))
const starterPrompts = [
  { key: 'explain', icon: 'terminal' }, { key: 'debug', icon: 'search' },
  { key: 'script', icon: 'document' }, { key: 'document', icon: 'book' },
] as const
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
const MIN_COMPOSER_HEIGHT_PX = 40
const MAX_COMPOSER_HEIGHT_PX = 200
const SCROLL_FOLLOW_THRESHOLD_PX = 120

interface AttachmentFeedback {
  id: string
  kind: 'error' | 'warning'
  fileName: string
  message: string
}

const { t, locale } = useI18n()
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
  deleteMessage,
  clearAllSessions: clearLocalSessions,
  saveDraftConfig,
  consumeDraftConfig,
  flushPendingChatUpdates,
} = useLocalChat()

const baseUrl = ref('')
const historyOpen = ref(window.innerWidth >= 1024)
const connectionSettingsOpen = ref(false)
const focusMode = ref(false)
const apiKey = ref('')
const selectedModel = ref('')
const selectedReasoningEffort = ref<ChatReasoningEffort>('medium')
const models = ref<OpenAIModel[]>([])
const loadingModels = ref(false)
const draftMessage = ref('')
const editingMessageId = ref<string | null>(null)
const composerMode = ref<'edit' | 'retry'>('edit')
const draftAttachments = ref<LocalChatAttachment[]>([])
const isStreaming = ref(false)
const streamStatus = ref<ChatResponseStatus>('waiting')
const streamingMessageId = ref<string | null>(null)
const elapsedMs = ref(0)
let elapsedTimer: ReturnType<typeof setInterval> | undefined
let flushActiveReply: (() => void) | null = null
let renderingStreamingMessage = false
const renderedMessages = new Map<string, { content: string; locale: string; streaming: boolean; html: string }>()
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
const codePreview = ref<{ text: string; language: string } | null>(null)
const isNearMessagesBottom = ref(true)

const messageRenderer = new marked.Renderer()
messageRenderer.code = ({ text, lang }) => {
  const language = normalizeCodeLanguage(lang)
  const codeContent = renderingStreamingMessage ? escapeHtml(text) : highlightCode(text, language)
  const languageClass = language ? ` class="language-${escapeHtmlAttribute(language)}"` : ''
  const label = formatCodeLanguageLabel(language)
  const copyCodeLabel = t('chat.copyCode')

  return [
    `<div class="chat-code-block" data-code-language="${escapeHtmlAttribute(language)}">`,
    '  <div class="chat-code-header">',
    `    <span class="chat-code-label">${escapeHtml(label)}</span>`,
    '    <div class="chat-code-actions">',
    `    <button type="button" class="chat-code-copy" data-code-action="preview">${escapeHtml(t('chat.codePreview'))}</button>`,
    `    <button type="button" class="chat-code-copy" data-code-action="download">${escapeHtml(t('chat.downloadCode'))}</button>`,
    `    <button type="button" class="chat-code-copy" data-copy-type="code" data-code-action="copy" data-copy-default-label="${escapeHtmlAttribute(copyCodeLabel)}" aria-label="${escapeHtmlAttribute(copyCodeLabel)}">${escapeHtml(copyCodeLabel)}</button>`,
    '    </div>',
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

/**
 * Responses is reserved for models known to implement the OpenAI Responses
 * contract. Unknown or third-party models use Chat Completions so they do not
 * receive Responses-only fields such as web_search and reasoning.
 */
function modelUsesResponses(modelID: string): boolean {
  const normalized = modelID.trim().toLowerCase()
  if (!normalized) return false
  if (/^(claude|grok|gemini|suno|gpt-image|deepseek|kimi|qwen|moonshot|glm|minimax|firefly-video|sora-|wan3|viral|dola-)/.test(normalized)) {
    return false
  }
  if (/^(gpt-|o[1-9](?:-|$)|codex-)/.test(normalized)) return true
  const metadata = models.value.find((model) => model.id === modelID)
  const owner = metadata?.owned_by?.trim().toLowerCase()
  return owner === 'openai' || owner === 'codex'
}

const showScrollToLatest = computed(() =>
  Boolean(activeSession.value?.messages.length && !isNearMessagesBottom.value)
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

const canSubmitRequest = computed(() =>
  Boolean(apiKey.value.trim() && baseUrl.value.trim() &&
    models.value.some((model) => model.id === selectedModel.value) &&
    !loadingModels.value && !isStreaming.value && !processingAttachments.value)
)

const canSend = computed(() =>
  Boolean(
    (draftMessage.value.trim() || draftAttachments.value.length > 0) &&
    canSubmitRequest.value
  )
)

const setupHint = computed(() => {
  if (!apiKey.value.trim() || !baseUrl.value.trim()) return t('chat.setupConnection')
  if (!models.value.length && !loadingModels.value) return t('chat.setupModels')
  return ''
})

watch(
  () => activeSession.value?.id,
  () => {
    renderedMessages.clear()
    editingMessageId.value = null
    draftMessage.value = ''
    draftAttachments.value = []
    syncActiveSessionConfig()
    void scrollMessagesToBottom(true)
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

watch([historyOpen, focusMode], () => {
  void nextTick().then(() => resizeComposerTextarea())
})

watch([baseUrl, apiKey], () => {
  models.value = []
  errorMessage.value = ''
})

watch([baseUrl, apiKey, selectedModel, selectedReasoningEffort], () => {
  if (!activeSession.value) return
  updateSession(activeSession.value.id, {
    baseUrl: baseUrl.value.trim(),
    apiKey: apiKey.value.trim(),
    model: selectedModel.value.trim(),
    reasoningEffort: selectedReasoningEffort.value,
    apiKeyHint: maskKey(apiKey.value),
  })
})

onMounted(async () => {
  await appStore.fetchPublicSettings()

  baseUrl.value = appStore.cachedPublicSettings?.api_base_url || `${window.location.origin}/v1`

  const draft = consumeDraftConfig()
  if (draft) {
    if (activeSession.value && (draft.apiKey !== activeSession.value.apiKey || draft.baseUrl !== activeSession.value.baseUrl)) {
      setActiveSession(null)
    }
    baseUrl.value = draft.baseUrl || baseUrl.value
    apiKey.value = draft.apiKey || ''
    selectedModel.value = draft.model?.trim() || ''
    selectedReasoningEffort.value = normalizeReasoningEffort(draft.reasoningEffort)
  } else if (!activeSession.value && sessions.value[0]) {
    setActiveSession(sessions.value[0].id)
  }

  if (!draft) {
    syncActiveSessionConfig()
  }
  connectionSettingsOpen.value = !apiKey.value.trim()

  if (apiKey.value.trim() && baseUrl.value.trim()) {
    await loadModels()
  }

  await nextTick()
  resizeComposerTextarea()
})

onUnmounted(() => {
  flushChatOnLeave()
  clearInterval(elapsedTimer)
  document.removeEventListener('keydown', handleWorkspaceEscape)
  window.removeEventListener('pagehide', flushChatOnLeave)
  window.removeEventListener('resize', resizeComposerTextarea)
})

onMounted(() => {
  document.addEventListener('keydown', handleWorkspaceEscape)
  window.addEventListener('pagehide', flushChatOnLeave)
  window.addEventListener('resize', resizeComposerTextarea)
})

function flushChatOnLeave() {
  flushActiveReply?.()
  activeAbortController.value?.abort()
  if (isStreaming.value && activeSession.value && streamingMessageId.value) {
    updateMessage(activeSession.value.id, streamingMessageId.value, getMessageContent(activeSession.value.id, streamingMessageId.value), {
      status: 'stopped', elapsedMs: elapsedMs.value,
    })
  }
  flushPendingChatUpdates()
}

function handleWorkspaceEscape(event: KeyboardEvent) {
  if (codePreview.value) return
  if (event.key === 'Escape') focusMode.value = false
}

function handleCreateSession() {
  if (isStreaming.value) return
  createSession({
    title: t('chat.defaultSessionTitle'),
    model: selectedModel.value.trim(),
    reasoningEffort: selectedReasoningEffort.value,
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

function useStarterPrompt(key: typeof starterPrompts[number]['key']) {
  draftMessage.value = t(`chat.starters.${key}.prompt`)
  if (!apiKey.value.trim()) connectionSettingsOpen.value = true
  void nextTick().then(() => composerTextareaRef.value?.focus())
}

function selectSession(id: string) {
  if (isStreaming.value) return
  setActiveSession(id)
  draftMessage.value = ''
  draftAttachments.value = []
  errorMessage.value = ''
  attachmentFeedbackMessages.value = []
  copiedMessageId.value = null
}

function removeSession(id: string) {
  if (isStreaming.value) return
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
  cancelEditMessage()
}

function clearAllSessions() {
  clearLocalSessions()
  copiedMessageId.value = null
  cancelEditMessage()
}

function beginEditMessage(message: LocalChatMessage) {
  if (isStreaming.value || processingAttachments.value || message.role !== 'user') return
  editingMessageId.value = message.id
  composerMode.value = 'edit'
  draftMessage.value = message.content
  draftAttachments.value = message.attachments?.map((attachment) => ({ ...attachment })) || []
  void nextTick().then(() => composerTextareaRef.value?.focus())
}

function cancelEditMessage() {
  editingMessageId.value = null
  draftMessage.value = ''
  draftAttachments.value = []
}

async function regenerateMessage(assistantMessageId: string) {
  if (!canSubmitRequest.value || !activeSession.value) return
  const index = activeSession.value.messages.findIndex((message) => message.id === assistantMessageId)
  if (index < 0) return
  const userMessage = activeSession.value.messages.slice(0, index).reverse().find((message) => message.role === 'user')
  if (!userMessage) return
  beginEditMessage(userMessage)
  composerMode.value = 'retry'
  await sendMessage()
}

function handlePrimaryAction() {
  if (isStreaming.value) {
    stopStreaming()
    return
  }

  void sendMessage()
}

function handleComposerEnter(event: KeyboardEvent) {
  if (event.isComposing || isStreaming.value) {
    return
  }

  event.preventDefault()
  void sendMessage()
}

function handleModelChange(value: string | number | boolean | null) {
  selectedModel.value = String(value ?? '').trim()
}

function handleReasoningEffortChange(value: string | number | boolean | null) {
  selectedReasoningEffort.value = normalizeReasoningEffort(String(value ?? ''))
}

function normalizeReasoningEffort(value?: string): ChatReasoningEffort {
  return REASONING_EFFORTS.find((effort) => effort === value) || 'medium'
}

async function loadModels() {
  if (loadingModels.value || isStreaming.value) return
  const requestBaseUrl = baseUrl.value.trim()
  const requestApiKey = apiKey.value.trim()
  if (!requestBaseUrl || !requestApiKey) {
    models.value = []
    errorMessage.value = t(requestBaseUrl ? 'chat.errors.apiKeyRequired' : 'chat.errors.baseUrlRequired')
    return
  }

  loadingModels.value = true
  errorMessage.value = ''

  try {
    const result = await chatAPI.fetchOpenAIModels(requestBaseUrl, requestApiKey)
    if (requestBaseUrl !== baseUrl.value.trim() || requestApiKey !== apiKey.value.trim()) return
    models.value = [...new Map(result.filter((model) => model.id?.trim()).map((model) => [model.id, model])).values()]
    if (!models.value.some((model) => model.id === selectedModel.value)) {
      selectedModel.value = models.value[0]?.id || ''
    }
    if (!models.value.length) errorMessage.value = t('chat.errors.noModels')
  } catch (error) {
    if (requestBaseUrl !== baseUrl.value.trim() || requestApiKey !== apiKey.value.trim()) return
    models.value = []
    errorMessage.value = (error as Error).message || t('chat.errors.loadModelsFailed')
  } finally {
    loadingModels.value = false
  }
}

async function sendMessage() {
  if (!canSend.value) {
    return
  }

  const prompt = draftMessage.value.trim()
  const model = selectedModel.value.trim()
  const attachments = draftAttachments.value.map((attachment) => ({ ...attachment }))
  if (!activeSession.value) {
    createSession({
      title: t('chat.defaultSessionTitle'), model, reasoningEffort: selectedReasoningEffort.value,
      apiKey: apiKey.value, baseUrl: baseUrl.value, apiKeyHint: maskKey(apiKey.value),
    })
  }
  if (!activeSession.value) return
  let sessionId = activeSession.value.id
  if (editingMessageId.value) {
    const source = activeSession.value
    const index = source.messages.findIndex((message) => message.id === editingMessageId.value)
    if (index < 0) return
    const branch = createSession({
      title: `${source.title} · ${t('chat.branch')}`,
      model, reasoningEffort: selectedReasoningEffort.value,
      apiKey: apiKey.value, baseUrl: baseUrl.value, apiKeyHint: maskKey(apiKey.value),
    })
    sessionId = branch.id
    updateSession(sessionId, { messages: source.messages.slice(0, index).map((message) => ({
      ...message, attachments: message.attachments?.map((attachment) => ({ ...attachment })),
    })) })
    editingMessageId.value = null
  }
  errorMessage.value = ''
  attachmentFeedbackMessages.value = []
  copiedMessageId.value = null

  const userMessage = appendMessage(sessionId, 'user', prompt, attachments)
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
    reasoningEffort: selectedReasoningEffort.value,
    apiKeyHint: maskKey(apiKey.value),
  })

  const assistantMessage = appendMessage(sessionId, 'assistant', '')
  if (!assistantMessage) {
    return
  }
  void scrollMessagesToBottom(true)

  const history: ChatCompletionMessage[] = activeSession.value.messages
    .filter((message) => message.id !== assistantMessage.id)
    .map((message) => buildChatCompletionMessage(message))
  const useResponses = modelUsesResponses(model)

  const controller = new AbortController()
  activeAbortController.value = controller
  isStreaming.value = true
  streamingMessageId.value = assistantMessage.id
  streamStatus.value = 'connecting'
  elapsedMs.value = 0
  const startedAt = Date.now()
  let upstreamCompleted = false
  let pendingText = ''
  let renderTimer: ReturnType<typeof setTimeout> | undefined
  const flushReply = () => {
    clearTimeout(renderTimer)
    renderTimer = undefined
    if (!pendingText) return
    const text = pendingText
    pendingText = ''
    updateMessage(sessionId, assistantMessage.id, getMessageContent(sessionId, assistantMessage.id) + text, { status: 'outputting' })
    void scrollMessagesToBottom()
  }
  flushActiveReply = flushReply
  elapsedTimer = setInterval(() => { elapsedMs.value = Date.now() - startedAt }, 250)
  const updateResponse = (status: ChatResponseStatus, reasoningSummary?: string, sources?: LocalChatResponse['sources']) => {
    streamStatus.value = status
    updateMessage(sessionId, assistantMessage.id, getMessageContent(sessionId, assistantMessage.id), {
      status, elapsedMs: Date.now() - startedAt,
      ...(reasoningSummary !== undefined ? { reasoningSummary } : {}),
      ...(sources ? { sources } : {}),
    })
  }
  updateResponse('connecting')

  try {
    const streamOptions = {
      signal: controller.signal,
      onOpen: () => updateResponse('waiting'),
      onDelta: (delta: string) => {
        if (streamStatus.value !== 'outputting') updateResponse('outputting')
        pendingText += delta
        if (renderTimer === undefined) renderTimer = setTimeout(flushReply, 50)
      },
      onEvent: (event: ResponsesStreamEvent) => {
        const annotations = [
          ...(event.annotation ? [event.annotation] : []),
          ...(event.part?.annotations || []),
          ...(event.item?.content?.flatMap((part) => part.annotations || []) || []),
          ...(event.item?.action?.sources || []),
          ...(event.response?.output?.flatMap((item) => item.content?.flatMap((part) => part.annotations || []) || []) || []),
          ...(event.response?.output?.flatMap((item) => item.action?.sources || []) || []),
        ]
        const sources = collectResponseSources(annotations)
        if (sources.length) {
          const existing = sessions.value.find((session) => session.id === sessionId)?.messages.find((message) => message.id === assistantMessage.id)?.response?.sources || []
          updateResponse(streamStatus.value, undefined, [...new Map([...existing, ...sources].map((source) => [source.url, source])).values()])
        }
        if (event.type === 'response.created' || event.type === 'response.in_progress') updateResponse('waiting')
        if (event.type?.startsWith('response.web_search_call.') && !event.type.endsWith('.completed')) updateResponse('searching')
        if (event.type === 'response.output_item.added' && event.item?.type === 'reasoning') updateResponse('thinking')
        if (event.type === 'response.reasoning_summary_text.delta' && event.delta) {
          const summary = sessions.value.find((session) => session.id === sessionId)?.messages.find((message) => message.id === assistantMessage.id)?.response?.reasoningSummary || ''
          updateResponse('thinking', summary + event.delta)
        }
        if (event.type === 'response.output_item.done' && event.item?.type === 'reasoning') {
          const summary = event.item.summary?.map((part) => part.text || '').join('\n')
          if (summary) updateResponse(streamStatus.value, summary)
        }
        if (event.type === 'response.web_search_call.completed' && streamStatus.value === 'searching') updateResponse('waiting')
        if (event.type === 'response.completed' || event.type === 'response.done') {
          upstreamCompleted = true
          const summary = event.response?.output?.filter((item) => item.type === 'reasoning').flatMap((item) => item.summary || []).map((part) => part.text || '').join('\n')
          if (summary) updateResponse(streamStatus.value, summary)
        }
      },
    }

    if (useResponses) {
      await chatAPI.streamResponses(
        baseUrl.value,
        apiKey.value,
        {
          model,
          input: history.map((message) => buildResponsesInputItem(message)),
          reasoning: {
            effort: selectedReasoningEffort.value,
          },
          tools: [{ type: 'web_search' }],
          store: false,
        },
        streamOptions
      )
    } else {
      const chatPayload: ChatCompletionRequest = { model, messages: history, stream: true }
      updateResponse('waiting')
      await chatAPI.streamChatCompletions(
        baseUrl.value,
        apiKey.value,
        chatPayload,
        { signal: controller.signal, onDelta: streamOptions.onDelta }
      )
      upstreamCompleted = true
    }
    flushReply()
    if (!upstreamCompleted) throw new Error(t('chat.errors.incompleteReply'))
    if (!getMessageContent(sessionId, assistantMessage.id).trim()) throw new Error(t('chat.errors.emptyAssistantReply'))
    updateResponse('completed')
  } catch (error) {
    flushReply()
    if ((error as Error).name === 'AbortError') {
      updateResponse('stopped')
      if (!getMessageContent(sessionId, assistantMessage.id).trim()) {
        deleteMessage(sessionId, assistantMessage.id)
      }
      return
    }

    const normalizedError = normalizeChatSendError(error)
    updateResponse('failed')
    errorMessage.value = normalizedError.message
    const currentContent = getMessageContent(sessionId, assistantMessage.id)
    if (!currentContent.trim()) {
      updateMessage(
        sessionId,
        assistantMessage.id,
        normalizedError.useAssistantErrorMessage ? normalizedError.message : t('chat.errors.emptyAssistantReply')
      )
    }
    if (!draftMessage.value.trim() && !draftAttachments.value.length && userMessage) {
      draftMessage.value = prompt
      draftAttachments.value = attachments
      editingMessageId.value = userMessage.id
      composerMode.value = 'retry'
    }
  } finally {
    flushReply()
    flushActiveReply = null
    flushPendingChatUpdates()
    elapsedMs.value = Date.now() - startedAt
    clearInterval(elapsedTimer)
    isStreaming.value = false
    activeAbortController.value = null
    streamingMessageId.value = null
  }
}

function normalizeChatSendError(error: unknown): { message: string; useAssistantErrorMessage: boolean } {
  const rawMessage = error instanceof Error ? error.message : ''
  const message = rawMessage || t('chat.errors.sendFailed')
  const lowerMessage = message.toLowerCase()
  const isNetworkError =
    lowerMessage.includes('failed to fetch') ||
    lowerMessage.includes('fetch failed') ||
    lowerMessage.includes('networkerror') ||
    lowerMessage.includes('network error') ||
    lowerMessage.includes('network request failed') ||
    lowerMessage.includes('load failed') ||
    lowerMessage.includes('connection reset') ||
    lowerMessage.includes('connection closed') ||
    lowerMessage.includes('terminated') ||
    lowerMessage.includes('unexpected end of stream') ||
    lowerMessage.includes('body stream')

  if (isNetworkError) {
    return {
      message: t('chat.errors.networkFailed'),
      useAssistantErrorMessage: true,
    }
  }

  return {
    message,
    useAssistantErrorMessage: false,
  }
}

function stopStreaming() {
  activeAbortController.value?.abort()
}

function renderMessage(message: LocalChatMessage) {
  const streaming = isStreaming.value && streamingMessageId.value === message.id
  const cached = renderedMessages.get(message.id)
  if (cached && cached.content === message.content && cached.locale === locale.value && cached.streaming === streaming) return cached.html
  renderingStreamingMessage = streaming
  const html = marked.parse(message.content || '', {
    breaks: true,
    gfm: true,
    renderer: messageRenderer,
  }) as string
  renderingStreamingMessage = false

  const sanitized = DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    ADD_TAGS: ['button'],
    ADD_ATTR: ['type', 'data-copy-type', 'data-copy-default-label', 'data-code-action', 'data-code-language', 'aria-label'],
  })
  renderedMessages.set(message.id, { content: message.content, locale: locale.value, streaming, html: sanitized })
  return sanitized
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
  const actionButton = target?.closest('[data-code-action]') as HTMLButtonElement | null
  if (!actionButton) {
    return
  }

  event.preventDefault()

  const codeElement = actionButton.closest('.chat-code-block')?.querySelector('pre code')
  const codeText = codeElement?.textContent ?? ''
  if (!codeText.trim()) {
    return
  }

  const language = (actionButton.closest('.chat-code-block') as HTMLElement)?.dataset.codeLanguage || ''
  if (actionButton.dataset.codeAction === 'preview') {
    codePreview.value = { text: codeText, language }
    return
  }
  if (actionButton.dataset.codeAction === 'download') {
    downloadCode(codeText, language)
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

function highlightCode(text: string, language: string): string {
  return DOMPurify.sanitize(hljs.getLanguage(language)
    ? hljs.highlight(text, { language, ignoreIllegals: true }).value
    : escapeHtml(text))
}

function downloadCode(text: string, language: string) {
  const extensions: Record<string, string> = { javascript: 'js', typescript: 'ts', python: 'py', bash: 'sh', shell: 'sh', xml: 'html', yaml: 'yaml', json: 'json', go: 'go', css: 'css', sql: 'sql', js: 'js', ts: 'ts', py: 'py' }
  saveAs(new Blob([text], { type: 'text/plain;charset=utf-8' }), `snippet.${extensions[language] || 'txt'}`)
}

function collectResponseSources(annotations: ResponsesAnnotation[]) {
  return annotations.flatMap((annotation) => {
    try {
      const url = new URL(annotation.url || '')
      if (url.protocol !== 'https:' && url.protocol !== 'http:') return []
      return [{ url: url.href, title: annotation.title?.trim() || url.hostname }]
    } catch {
      return []
    }
  })
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

  return 'w-full max-w-full'
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

function handleMessagesScroll() {
  const container = messagesContainerRef.value
  if (!container) return
  const distanceFromBottom = container.scrollHeight - container.scrollTop - container.clientHeight
  isNearMessagesBottom.value = distanceFromBottom <= SCROLL_FOLLOW_THRESHOLD_PX
}

async function scrollMessagesToBottom(force = false) {
  if (!force && !isNearMessagesBottom.value) return
  await nextTick()
  if (!force && !isNearMessagesBottom.value) return
  if (!messagesContainerRef.value) return
  messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight
  isNearMessagesBottom.value = true
}

function resizeComposerTextarea() {
  const textarea = composerTextareaRef.value
  if (!textarea) return

  textarea.style.height = 'auto'
  const nextHeight = Math.min(
    Math.max(textarea.scrollHeight, MIN_COMPOSER_HEIGHT_PX),
    Math.max(MIN_COMPOSER_HEIGHT_PX, Math.min(MAX_COMPOSER_HEIGHT_PX, window.innerHeight * 0.25))
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
  selectedModel.value = activeSession.value.model?.trim() || ''
  selectedReasoningEffort.value = normalizeReasoningEffort(activeSession.value.reasoningEffort)
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

function buildResponsesInputItem(message: ChatCompletionMessage): ResponsesInputItem {
  return {
    role: message.role,
    content: buildResponsesMessageContent(message),
  }
}

function buildResponsesMessageContent(
  message: ChatCompletionMessage
): ResponsesInputItem['content'] {
  if (typeof message.content === 'string') {
    return message.content
  }

  const parts: ResponsesContentPart[] = []
  for (const part of message.content) {
    if (part.type === 'text' && part.text) {
      parts.push({
        type: message.role === 'assistant' ? 'output_text' : 'input_text',
        text: part.text,
      })
      continue
    }

    if (part.type === 'image_url' && part.image_url.url) {
      parts.push({
        type: 'input_image',
        image_url: part.image_url.url,
      })
    }
  }

  return parts.length ? parts : ''
}

function formatTextAttachmentForModel(attachment: LocalChatAttachment): string {
  return [
    `[Attached file: ${attachment.name}]`,
    attachment.textContent?.trim() || '',
    `[End of file: ${attachment.name}]`,
  ].join('\n')
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
  if (processingAttachments.value) {
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
    const pdfjsLib = await import('pdfjs-dist/build/pdf.mjs')
    if (!pdfjsLib.GlobalWorkerOptions.workerSrc) {
      pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorkerUrl
    }
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
    const mammoth = (await import('mammoth')).default
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
  if (processingAttachments.value) {
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
    reasoningEffort: selectedReasoningEffort.value,
  })
}

watch([baseUrl, apiKey, selectedModel, selectedReasoningEffort], persistCurrentDraft)
watch(() => router.currentRoute.value.fullPath, persistCurrentDraft)
</script>

<style scoped>
.chat-workspace {
  height: max(420px, calc(100dvh - 6rem));
  grid-template-columns: minmax(0, 1fr);
}

.chat-session-sidebar {
  position: absolute;
  inset: 0 auto 0 0;
  width: min(240px, 85%);
}

.chat-page-focus {
  position: fixed;
  inset: 0;
  z-index: 60;
  margin: 0;
  border-radius: 0;
}

.chat-page-focus .chat-workspace {
  height: calc(100dvh - 0.875rem);
}

@media (min-width: 768px) {
  .chat-workspace {
    height: max(420px, calc(100dvh - 6.25rem));
  }
  .chat-page-focus .chat-workspace {
    height: calc(100dvh - 1.125rem);
  }
}

@media (min-width: 1024px) {
  .chat-workspace.has-history {
    grid-template-columns: 200px minmax(0, 1fr);
  }
  .chat-session-sidebar {
    position: relative;
    width: auto;
  }
}

.chat-toolbar .btn {
  @apply px-3 py-1.5 text-xs;
}

.chat-composer :deep(.select-trigger) {
  min-height: 36px;
  padding-top: 6px;
  padding-bottom: 6px;
}

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

.chat-markdown :deep(.chat-code-actions) {
  @apply flex flex-wrap items-center justify-end gap-1.5;
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
