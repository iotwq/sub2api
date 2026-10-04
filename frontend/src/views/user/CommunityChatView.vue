<template>
  <AppLayout>
    <div class="community-chat-page">
      <section
        class="community-chat-panel"
        @dragenter.prevent="handleDragEnter"
        @dragover.prevent="handleDragOver"
        @dragleave.prevent="handleDragLeave"
        @drop.prevent="handleDrop"
      >
        <header class="community-chat-header">
          <div class="flex min-w-0 items-center gap-3">
            <div class="community-chat-logo">
              <Icon name="users" size="lg" />
            </div>
            <div class="min-w-0">
              <h1 class="truncate text-xl font-semibold text-gray-900 dark:text-white">
                {{ t('communityChat.title') }}
              </h1>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('communityChat.description') }}
              </p>
            </div>
          </div>

          <div class="community-contact-entry">
            <div class="community-contact-action">
              <button type="button" class="community-contact-button" aria-describedby="community-contact-hint community-direct-notification" @click="openDirectDialog">
                <span class="community-contact-icon" aria-hidden="true">
                  <Icon name="chatBubble" size="xl" />
                  <span class="community-contact-sparkle"><Icon name="sparkles" size="sm" /></span>
                </span>
                <span>{{ t('communityChat.contactOwner') }}</span>
              </button>
              <span id="community-direct-notification" class="community-direct-notification" role="status" aria-atomic="true">
                <span v-if="directUnread" class="community-direct-unread-badge">
                  <Icon name="bell" size="xs" aria-hidden="true" />
                  {{ t('communityChat.directNewMessage') }}
                </span>
              </span>
            </div>
            <p id="community-contact-hint" class="community-contact-hint">{{ t('communityChat.contactOwnerHint') }}</p>
          </div>

          <div class="community-chat-header-actions">
            <button
              type="button"
              class="community-header-icon-button"
              :class="{ 'community-header-icon-button-active': searchOpen }"
              :title="t('communityChat.search')"
              @click="toggleSearch"
            >
              <Icon name="search" size="sm" />
            </button>
            <div class="community-chat-status" :class="connectionStatusClass">
              <span class="h-2 w-2 rounded-full" :class="connectionDotClass"></span>
              {{ connectionStatusLabel }}
            </div>
          </div>
        </header>

        <div v-if="searchOpen" class="community-search-bar">
          <div class="community-search-input-shell">
            <Icon name="search" size="sm" class="shrink-0 text-gray-500 dark:text-gray-400" />
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="search"
              class="community-search-input"
              :placeholder="t('communityChat.searchPlaceholder')"
              @keydown.enter.prevent="runSearch"
            />
            <button
              v-if="searchQuery"
              type="button"
              class="community-search-clear"
              :title="t('communityChat.clearSearch')"
              @click="clearSearch"
            >
              <Icon name="x" size="xs" />
            </button>
          </div>
          <button type="button" class="btn btn-primary community-search-submit" :disabled="searching || !searchQuery.trim()" @click="runSearch">
            <Icon v-if="searching" name="refresh" size="sm" class="animate-spin" />
            <span>{{ t('communityChat.searchAction') }}</span>
          </button>
          <button type="button" class="community-search-close" :title="t('communityChat.closeSearch')" @click="closeSearch">
            <Icon name="x" size="sm" />
          </button>
        </div>

        <div v-if="searchActive" class="community-search-summary">
          <span>{{ t('communityChat.searchResults', { count: searchTotal }) }}</span>
          <div class="community-search-summary-actions">
            <span v-if="currentSearchResultIndex >= 0" class="community-search-position" aria-live="polite">
              {{ t('communityChat.searchPosition', { current: currentSearchResultIndex + 1, total: searchTotal }) }}
            </span>
            <Icon v-if="loadingSearchContext || loadingMoreSearchResults" name="refresh" size="xs" class="animate-spin" />
            <div v-if="searchTotal > 0" class="community-search-navigation">
              <button
                type="button"
                class="community-search-nav-button"
                :disabled="loadingSearchContext || loadingMoreSearchResults || !canNavigatePreviousSearchResult"
                :title="t('communityChat.previousSearchResult')"
                :aria-label="t('communityChat.previousSearchResult')"
                @click="navigateSearchResults(-1)"
              >
                <Icon name="chevronUp" size="sm" />
              </button>
              <button
                type="button"
                class="community-search-nav-button"
                :disabled="loadingSearchContext || loadingMoreSearchResults || !canNavigateNextSearchResult"
                :title="t('communityChat.nextSearchResult')"
                :aria-label="t('communityChat.nextSearchResult')"
                @click="navigateSearchResults(1)"
              >
                <Icon name="chevronDown" size="sm" />
              </button>
            </div>
            <button type="button" class="community-search-return" @click="closeSearch">{{ t('communityChat.backToLatest') }}</button>
          </div>
        </div>

        <div
          class="community-chat-message-region"
          :class="{ 'community-drop-target': isDraggingFile }"
        >
          <div v-if="isDraggingFile" class="community-drop-overlay" aria-hidden="true">
            <span class="community-drop-icon"><Icon name="upload" size="lg" /></span>
            <strong>{{ t('communityChat.dropFile') }}</strong>
            <span>{{ t('communityChat.dropFileHint') }}</span>
          </div>
          <main
            ref="messageListRef"
            class="community-chat-messages"
            @scroll="handleMessageListScroll"
          >
          <div v-if="loadingOlderHistory && !searchActive" class="flex justify-center py-2 text-gray-500 dark:text-gray-400">
            <Icon name="refresh" size="sm" class="animate-spin" />
          </div>
          <div v-if="loadingHistory || loadingSearchContext" class="community-chat-empty">
            <Icon name="refresh" size="lg" class="animate-spin" />
            <span>{{ t('communityChat.loading') }}</span>
          </div>

          <div v-else-if="displayedMessages.length === 0" class="community-chat-empty">
            <Icon :name="searchActive ? 'search' : 'chatBubble'" size="xl" />
            <span>{{ searchActive ? t('communityChat.noSearchResults') : t('communityChat.empty') }}</span>
          </div>

          <div v-else class="space-y-4">
            <article
              v-for="message in displayedMessages"
              :key="message.id"
              v-memo="[message, currentUserId, authStore.isAdmin, locale, activeSearchQuery, currentSearchResult?.id, deletingMessageId]"
              :id="`community-message-${message.id}`"
              class="community-message group"
              :class="{
                'community-message-own': message.user_id === currentUserId,
                'community-message-search-current': searchActive && currentSearchResult?.id === message.id,
              }"
            >
              <div class="community-avatar">
                <img
                  v-if="message.avatar_url"
                  :src="message.avatar_url"
                  :alt="message.username"
                  class="h-full w-full object-cover"
                />
                <span v-else>{{ avatarInitial(message.username) }}</span>
              </div>

              <div class="community-message-content">
                <div class="community-message-meta">
                  <span class="truncate font-medium text-gray-900 dark:text-white">
                    <template v-for="(segment, index) in searchTextSegments(message.username)" :key="`username-${message.id}-${index}`">
                      <mark
                        v-if="segment.matched"
                        class="community-search-keyword"
                        :class="{ 'community-search-keyword-current': isCurrentSearchResult(message) }"
                      >{{ segment.text }}</mark>
                      <template v-else>{{ segment.text }}</template>
                    </template>
                  </span>
                  <span v-if="message.sent_at" class="community-message-time">{{ formatTime(message.sent_at) }}</span>
                  <button
                    v-if="canRecall(message)"
                    type="button"
                    class="community-delete-button"
                    :disabled="deletingMessageId === message.id"
                    :title="message.user_id === currentUserId ? t('communityChat.recall') : t('communityChat.delete')"
                    @click="handleDelete(message)"
                  >
                    <Icon name="trash" size="xs" />
                  </button>
                </div>

                <div class="community-message-bubble">
                  <button
                    v-if="message.reply_to_message_id"
                    type="button"
                    class="community-reply-preview"
                    @click="scrollToMessage(message.reply_to_message_id)"
                  >
                    <span class="truncate font-medium">{{ message.reply_to_username }}</span>
                    <span class="truncate text-gray-500 dark:text-gray-400">
                      {{ replySourcePreviewText(message) }}
                    </span>
                  </button>
                  <template v-if="isImageMessage(message)">
                    <button
                      type="button"
                      class="community-image-button"
                      :title="t('communityChat.previewImage')"
                      @click="openImagePreview(message)"
                    >
                      <img
                        :src="attachmentURL(message)"
                        :alt="message.content"
                        class="community-message-image"
                        loading="lazy"
                        @error="handleAttachmentError"
                        @load="handleMessageMediaLoaded(message.id)"
                      />
                    </button>
                    <a
                      class="community-attachment-download"
                      :href="attachmentDownloadURL(message)"
                      :download="attachmentFileName(message)"
                    >
                      <Icon name="download" size="xs" />
                      {{ t('communityChat.download') }}
                    </a>
                  </template>
                  <template v-else-if="isVideoMessage(message)">
                    <video
                      class="community-message-video"
                      :src="attachmentURL(message)"
                      controls
                      preload="metadata"
                      @error="handleAttachmentError"
                      @loadedmetadata="handleMessageMediaLoaded(message.id)"
                    ></video>
                    <button
                      type="button"
                      class="community-attachment-download"
                      :title="t('communityChat.previewVideo')"
                      @click="openVideoPreview(message)"
                    >
                      <Icon name="play" size="xs" />
                      {{ t('communityChat.previewVideo') }}
                    </button>
                    <a
                      class="community-attachment-download"
                      :href="attachmentDownloadURL(message)"
                      :download="attachmentFileName(message)"
                    >
                      <Icon name="download" size="xs" />
                      {{ t('communityChat.download') }}
                    </a>
                  </template>
                  <a
                    v-else-if="isFileMessage(message)"
                    class="community-file-card"
                    :href="attachmentDownloadURL(message)"
                    :download="attachmentFileName(message)"
                  >
                    <span class="community-file-icon">
                      <Icon name="document" size="sm" />
                    </span>
                    <span class="min-w-0 flex-1">
                      <span class="block truncate font-semibold">
                        <template v-for="(segment, index) in searchTextSegments(attachmentFileName(message))" :key="`file-${message.id}-${index}`">
                          <mark
                            v-if="segment.matched"
                            class="community-search-keyword"
                            :class="{ 'community-search-keyword-current': isCurrentSearchResult(message) }"
                          >{{ segment.text }}</mark>
                          <template v-else>{{ segment.text }}</template>
                        </template>
                      </span>
                      <span class="mt-0.5 block text-xs opacity-75">{{ attachmentMeta(message) }}</span>
                    </span>
                    <Icon name="download" size="sm" class="shrink-0 opacity-70" />
                  </a>
                  <p v-if="shouldShowSearchAttachmentName(message)" class="community-search-attachment-name">
                    <template v-for="(segment, index) in searchTextSegments(attachmentFileName(message))" :key="`attachment-${message.id}-${index}`">
                      <mark
                        v-if="segment.matched"
                        class="community-search-keyword"
                        :class="{ 'community-search-keyword-current': isCurrentSearchResult(message) }"
                      >{{ segment.text }}</mark>
                      <template v-else>{{ segment.text }}</template>
                    </template>
                  </p>
                  <p v-if="shouldShowMessageText(message)" class="whitespace-pre-wrap break-words leading-6">
                    <template v-for="(segment, index) in searchTextSegments(messageDisplayText(message))" :key="`content-${message.id}-${index}`">
                      <mark
                        v-if="segment.matched"
                        class="community-search-keyword"
                        :class="{ 'community-search-keyword-current': isCurrentSearchResult(message) }"
                      >{{ segment.text }}</mark>
                      <template v-else>{{ segment.text }}</template>
                    </template>
                  </p>
                </div>

                <div class="community-message-hover-actions">
                  <button
                    type="button"
                    class="community-message-action"
                    :title="t('communityChat.reply')"
                    @click="setReplyTarget(message)"
                  >
                    {{ t('communityChat.reply') }}
                  </button>
                </div>
              </div>
            </article>
          </div>
          </main>

          <button
            v-if="newMessageCount > 0 && !searchActive"
            type="button"
            class="community-new-message-button"
            @click="scrollToBottom()"
          >
            <Icon name="arrowDown" size="sm" />
            {{ t('communityChat.newMessages', { count: newMessageCount }) }}
          </button>
        </div>

        <footer class="community-chat-composer">
          <div v-if="replyTarget" class="community-reply-target">
            <div class="min-w-0">
              <p class="truncate text-xs font-semibold text-[#8a2525] dark:text-[#f0b4a8]">
                {{ t('communityChat.replyingTo', { username: replyTarget.username }) }}
              </p>
              <p class="mt-1 truncate text-xs text-gray-600 dark:text-gray-300">
                {{ messagePreviewText(replyTarget) }}
              </p>
            </div>
            <button type="button" class="community-icon-button h-8 w-8" :title="t('communityChat.cancelReply')" @click="clearReplyTarget">
              <Icon name="x" size="xs" />
            </button>
          </div>

          <div v-if="selectedFile" class="community-image-preview">
            <div class="flex min-w-0 items-center gap-3">
              <img
                v-if="selectedFilePreview"
                :src="selectedFilePreview"
                alt=""
                class="h-12 w-12 rounded-xl object-cover"
              />
              <span v-else class="community-selected-file-icon">
                <Icon name="document" size="sm" />
              </span>
              <div class="min-w-0">
                <p class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ selectedFile.name }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ formatFileSize(selectedFile.size) }}</p>
              </div>
            </div>
            <button type="button" class="community-icon-button" :title="t('communityChat.removeFile')" @click="clearSelectedFile">
              <Icon name="x" size="sm" />
            </button>
          </div>

          <div v-if="uploadProgress" class="mb-2 space-y-1 text-xs text-gray-600 dark:text-gray-300" role="status">
            <span>{{ uploadProgressText(uploadProgress.percent) }}</span>
            <progress :value="uploadProgress.percent ?? undefined" max="100" class="h-1.5 w-full accent-emerald-600" :aria-label="t('communityChat.uploading')"></progress>
          </div>

          <div
            class="community-input-row"
            :class="{ 'community-input-row-dragging': isDraggingFile }"
          >
            <input
              ref="fileInputRef"
              type="file"
              :accept="attachmentAccept"
              class="hidden"
              @change="handleFileSelected"
            />

            <button
              type="button"
              class="community-icon-button"
              :title="t('communityChat.uploadFile')"
              :disabled="sending"
              @click="fileInputRef?.click()"
            >
              <Icon name="upload" size="sm" />
            </button>

            <div class="community-textarea-shell">
              <textarea
                ref="messageInputRef"
                v-model="draft"
                class="community-textarea community-textarea-with-emoji"
                rows="1"
                :maxlength="maxTextLength"
                :placeholder="t('communityChat.inputPlaceholder')"
                @blur="rememberDraftSelection"
                @click="rememberDraftSelection"
                @input="handleDraftInput"
                @keyup="rememberDraftSelection"
                @paste="handlePaste"
                @select="rememberDraftSelection"
                @keydown.enter.exact="handleComposerEnter($event, false)"
              ></textarea>
              <CommunityEmojiPicker
                class="community-emoji-picker-trigger"
                :locale="locale"
                :label="t('communityChat.emojiPicker')"
                @select="insertEmoji"
              />
            </div>

            <button
              type="button"
              class="btn btn-primary community-send-button"
              :disabled="!canSend || sending"
              @click="handleSend"
            >
              <Icon v-if="sending" name="refresh" size="sm" class="mr-1.5 animate-spin" />
              <Icon v-else name="arrowUp" size="sm" class="mr-1.5" />
              {{ t('communityChat.send') }}
            </button>
          </div>

          <div class="mt-2 flex items-center justify-end text-xs text-gray-500 dark:text-gray-400">
            <span>{{ draft.length }}/{{ maxTextLength }}</span>
          </div>
        </footer>
      </section>

      <BaseDialog
        :show="previewImage !== null"
        :title="previewImage ? attachmentFileName(previewImage) : ''"
        width="wide"
        :z-index="60"
        :close-on-click-outside="true"
        @close="closeImagePreview"
      >
        <div v-if="previewImage" class="community-image-lightbox">
          <img
            :src="attachmentURL(previewImage)"
            :alt="previewImage.content"
            class="community-image-lightbox-img"
            @error="handleAttachmentError"
          />
          <a
            class="btn btn-secondary"
            :href="attachmentDownloadURL(previewImage)"
            :download="attachmentFileName(previewImage)"
          >
            <Icon name="download" size="sm" class="mr-1.5" />
            {{ t('communityChat.download') }}
          </a>
        </div>
      </BaseDialog>

      <BaseDialog
        :show="previewVideo !== null"
        :title="previewVideo ? attachmentFileName(previewVideo) : ''"
        width="wide"
        :z-index="60"
        :close-on-click-outside="true"
        @close="closeVideoPreview"
      >
        <div v-if="previewVideo" class="community-image-lightbox">
          <video
            class="community-video-lightbox"
            :src="attachmentURL(previewVideo)"
            controls
            autoplay
            preload="metadata"
            @error="handleAttachmentError"
          ></video>
          <a
            class="btn btn-secondary"
            :href="attachmentDownloadURL(previewVideo)"
            :download="attachmentFileName(previewVideo)"
          >
            <Icon name="download" size="sm" class="mr-1.5" />
            {{ t('communityChat.download') }}
          </a>
        </div>
      </BaseDialog>

      <BaseDialog
        :show="directDialogOpen"
        :title="t('communityChat.contactOwner')"
        width="wide"
        :close-on-escape="previewImage === null && previewVideo === null"
        :close-on-click-outside="true"
        @close="closeDirectDialog"
      >
        <div class="community-direct-chat" :class="{ 'community-direct-chat-single': !authStore.isAdmin }">
          <aside v-if="authStore.isAdmin" class="community-direct-conversations">
            <div class="community-direct-conversations-title">
              {{ t('communityChat.directConversationList') }}
            </div>
            <div class="community-direct-user-search">
              <Icon name="search" size="sm" class="shrink-0 text-gray-500 dark:text-gray-400" />
              <input
                v-model="directUserSearch"
                type="search"
                class="min-w-0 flex-1 bg-transparent text-sm text-gray-900 outline-none placeholder:text-gray-500 dark:text-white dark:placeholder:text-gray-400"
                :placeholder="t('communityChat.directSearchPlaceholder')"
              />
              <Icon v-if="loadingDirectUsers" name="refresh" size="xs" class="shrink-0 animate-spin text-gray-500" />
            </div>
            <div v-if="loadingDirectConversations && !directUserSearch.trim()" class="community-direct-empty">
              <Icon name="refresh" size="sm" class="animate-spin" />
              <span>{{ t('communityChat.loading') }}</span>
            </div>
            <div v-else-if="directUserSearch.trim() && !loadingDirectUsers && directUserResults.length === 0" class="community-direct-empty">
              <Icon name="search" size="lg" />
              <span>{{ t('communityChat.directNoUserFound') }}</span>
            </div>
            <div v-else-if="!directUserSearch.trim() && directConversations.length === 0" class="community-direct-empty">
              <Icon name="inbox" size="lg" />
              <span>{{ t('communityChat.directNoConversation') }}</span>
            </div>
            <div v-else class="community-direct-conversation-list community-direct-scroll-list">
              <button
                v-for="conversation in visibleDirectUsers"
                :key="conversation.user_id"
                type="button"
                class="community-direct-conversation"
                :class="{ 'community-direct-conversation-active': selectedDirectUserId === conversation.user_id }"
                @click="selectDirectUser(conversation)"
              >
                <div class="community-avatar h-9 w-9">
                  <img
                    v-if="conversation.avatar_url"
                    :src="conversation.avatar_url"
                    :alt="conversation.username"
                    class="h-full w-full object-cover"
                  />
                  <span v-else>{{ avatarInitial(conversation.username) }}</span>
                </div>
                <div class="min-w-0 flex-1 text-left">
                  <div class="flex min-w-0 items-center justify-between gap-2">
                    <span class="truncate text-sm font-semibold">{{ conversation.username }}</span>
                    <span class="flex shrink-0 items-center gap-2">
                      <span v-if="'unread_count' in conversation && conversation.unread_count > 0" class="community-direct-count-badge">
                        {{ conversation.unread_count > 99 ? '99+' : conversation.unread_count }}
                      </span>
                      <span v-if="'updated_at' in conversation" class="text-[11px] opacity-70">{{ formatTime(conversation.updated_at) }}</span>
                    </span>
                  </div>
                  <p v-if="'last_message' in conversation && conversation.last_message" class="mt-1 truncate text-xs opacity-75">
                    {{ messagePreviewText(conversation.last_message) }}
                  </p>
                  <p v-else class="mt-1 truncate text-xs opacity-75">{{ t('communityChat.directStartConversation') }}</p>
                </div>
              </button>
              <button
                v-if="!directUserSearch.trim() && hasMoreDirectConversations"
                type="button"
                class="community-direct-load-more"
                :disabled="loadingMoreDirectConversations"
                @click="loadMoreDirectConversations"
              >
                <Icon v-if="loadingMoreDirectConversations" name="refresh" size="xs" class="animate-spin" />
                {{ t('communityChat.loadMore') }}
              </button>
            </div>
          </aside>

          <section class="community-direct-thread">
            <div v-if="authStore.isAdmin && !selectedDirectUserId" class="community-direct-empty community-direct-empty-thread">
              <Icon name="mail" size="xl" />
              <span>{{ t('communityChat.directSelectConversation') }}</span>
            </div>
            <template v-else>
              <div class="community-direct-thread-header">
                <div class="community-avatar h-10 w-10">
                  <img
                    v-if="directThreadAvatar"
                    :src="directThreadAvatar"
                    :alt="directThreadTitle"
                    class="h-full w-full object-cover"
                  />
                  <span v-else>{{ avatarInitial(directThreadTitle) }}</span>
                </div>
                <div class="min-w-0">
                  <p class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ directThreadTitle }}</p>
                  <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('communityChat.directDescription') }}</p>
                </div>
              </div>

              <div ref="directMessageListRef" class="community-direct-messages" @scroll="handleDirectMessageListScroll">
                <div v-if="loadingOlderDirectMessages" class="flex justify-center py-2 text-gray-500 dark:text-gray-400">
                  <Icon name="refresh" size="sm" class="animate-spin" />
                </div>
                <div v-if="loadingDirectMessages" class="community-direct-empty">
                  <Icon name="refresh" size="sm" class="animate-spin" />
                  <span>{{ t('communityChat.loading') }}</span>
                </div>
                <div v-else-if="directMessages.length === 0" class="community-direct-empty">
                  <Icon name="chatBubble" size="lg" />
                  <span>{{ t('communityChat.directEmpty') }}</span>
                </div>
                <div v-else class="space-y-4">
                  <article
                    v-for="message in directMessages"
                    :key="message.id"
                    v-memo="[message, currentUserId, authStore.isAdmin, locale, deletingDirectMessageId]"
                    class="community-message community-direct-message"
                    :class="{ 'community-message-own': message.user_id === currentUserId }"
                  >
                    <div class="community-avatar">
                      <img
                        v-if="message.avatar_url"
                        :src="message.avatar_url"
                        :alt="message.username"
                        class="h-full w-full object-cover"
                      />
                      <span v-else>{{ avatarInitial(message.username) }}</span>
                    </div>
                    <div class="community-message-content">
                      <div class="community-message-meta">
                        <span class="truncate font-medium text-gray-900 dark:text-white">{{ message.username }}</span>
                        <span v-if="message.sent_at" class="community-message-time">{{ formatTime(message.sent_at) }}</span>
                        <button
                          v-if="canRecall(message)"
                          type="button"
                          class="community-delete-button"
                          :disabled="deletingDirectMessageId === message.id"
                          :title="message.user_id === currentUserId ? t('communityChat.recall') : t('communityChat.delete')"
                          @click="handleDeleteDirect(message)"
                        >
                          <Icon name="trash" size="xs" />
                        </button>
                      </div>
                      <div class="community-message-bubble">
                        <template v-if="isImageMessage(message)">
                          <button type="button" class="community-image-button" :title="t('communityChat.previewImage')" @click="openImagePreview(message)">
                            <img :src="attachmentURL(message)" :alt="message.content" class="community-message-image" loading="lazy" @error="handleAttachmentError" @load="handleDirectMessageMediaLoaded" />
                          </button>
                          <a class="community-attachment-download" :href="attachmentDownloadURL(message)" :download="attachmentFileName(message)">
                            <Icon name="download" size="xs" />
                            {{ t('communityChat.download') }}
                          </a>
                        </template>
                        <template v-else-if="isVideoMessage(message)">
                          <video class="community-message-video" :src="attachmentURL(message)" controls preload="metadata" @error="handleAttachmentError" @loadedmetadata="handleDirectMessageMediaLoaded"></video>
                          <button type="button" class="community-attachment-download" :title="t('communityChat.previewVideo')" @click="openVideoPreview(message)">
                            <Icon name="play" size="xs" />
                            {{ t('communityChat.previewVideo') }}
                          </button>
                          <a class="community-attachment-download" :href="attachmentDownloadURL(message)" :download="attachmentFileName(message)">
                            <Icon name="download" size="xs" />
                            {{ t('communityChat.download') }}
                          </a>
                        </template>
                        <a v-else-if="isFileMessage(message)" class="community-file-card" :href="attachmentDownloadURL(message)" :download="attachmentFileName(message)">
                          <span class="community-file-icon"><Icon name="document" size="sm" /></span>
                          <span class="min-w-0 flex-1">
                            <span class="block truncate font-semibold">{{ attachmentFileName(message) }}</span>
                            <span class="mt-0.5 block text-xs opacity-75">{{ attachmentMeta(message) }}</span>
                          </span>
                          <Icon name="download" size="sm" class="shrink-0 opacity-70" />
                        </a>
                        <p v-if="shouldShowMessageText(message)" class="whitespace-pre-wrap break-words leading-6">{{ messageDisplayText(message) }}</p>
                      </div>
                    </div>
                  </article>
                </div>
              </div>

              <button
                v-if="newDirectMessageCount > 0"
                type="button"
                class="community-direct-new-message-button"
                @click="scrollDirectToBottom()"
              >
                <Icon name="arrowDown" size="sm" />
                {{ t('communityChat.newMessages', { count: newDirectMessageCount }) }}
              </button>

              <div class="community-direct-composer">
                <div v-if="selectedDirectFile" class="community-image-preview">
                  <div class="flex min-w-0 items-center gap-3">
                    <img v-if="selectedDirectFilePreview" :src="selectedDirectFilePreview" alt="" class="h-12 w-12 rounded-xl object-cover" />
                    <span v-else class="community-selected-file-icon"><Icon name="document" size="sm" /></span>
                    <div class="min-w-0">
                      <p class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ selectedDirectFile.name }}</p>
                      <p class="text-xs text-gray-500 dark:text-gray-400">{{ formatFileSize(selectedDirectFile.size) }}</p>
                    </div>
                  </div>
                  <button type="button" class="community-icon-button" :title="t('communityChat.removeFile')" @click="clearSelectedDirectFile">
                    <Icon name="x" size="sm" />
                  </button>
                </div>
                <div v-if="directUploadProgress && (!authStore.isAdmin || selectedDirectUserId === directUploadProgress.userId)" class="mb-2 space-y-1 text-xs text-gray-600 dark:text-gray-300" role="status">
                  <span>{{ uploadProgressText(directUploadProgress.percent) }}</span>
                  <progress :value="directUploadProgress.percent ?? undefined" max="100" class="h-1.5 w-full accent-emerald-600" :aria-label="t('communityChat.uploading')"></progress>
                </div>
                <div class="community-input-row">
                  <input ref="directFileInputRef" type="file" :accept="attachmentAccept" class="hidden" @change="handleDirectFileSelected" />
                  <button type="button" class="community-icon-button" :title="t('communityChat.uploadFile')" :disabled="sendingDirect" @click="directFileInputRef?.click()">
                    <Icon name="upload" size="sm" />
                  </button>
                  <textarea
                    ref="directMessageInputRef"
                    v-model="directDraft"
                    class="community-textarea"
                    rows="1"
                    :maxlength="maxTextLength"
                    :placeholder="t('communityChat.directInputPlaceholder')"
                    @input="resizeDirectMessageInput"
                    @paste="handleDirectPaste"
                    @keydown.enter.exact="handleComposerEnter($event, true)"
                  ></textarea>
                  <button type="button" class="btn btn-primary community-send-button" :disabled="!canSendDirect || sendingDirect" @click="handleDirectSend">
                    <Icon v-if="sendingDirect" name="refresh" size="sm" class="mr-1.5 animate-spin" />
                    <Icon v-else name="arrowUp" size="sm" class="mr-1.5" />
                    {{ t('communityChat.send') }}
                  </button>
                </div>
              </div>
            </template>
          </section>
        </div>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AxiosProgressEvent } from 'axios'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import CommunityEmojiPicker from '@/components/community/CommunityEmojiPicker.vue'
import Icon from '@/components/icons/Icon.vue'
import { communityChatAPI, type CommunityChatDirectConversation, type CommunityChatDirectUser, type CommunityChatEvent, type CommunityChatMessage } from '@/api'
import { useCommunityChatRealtime } from '@/composables/useCommunityChatRealtime'
import { useAppStore, useAuthStore } from '@/stores'
import { insertEmojiAtSelection } from './communityChatEmoji'
import { firstCommunityChatDroppedFile, resizeCommunityChatTextarea } from './communityChatComposer'
import { centeredCommunityChatSearchScrollTop, hasCommunityChatSearchMatch, splitCommunityChatSearchText } from './communityChatSearch'
import { isCommunityChatNearBottom, shouldFollowCommunityChatMessage } from './communityChatScroll'

const maxTextLength = 2000
const maxAttachmentBytes = 15 * 1024 * 1024
const historyPageSize = 80
const searchPageSize = 50
const searchContextRadius = 20
const directConversationPageSize = 30
const allowedAttachmentExtensions = new Set([
  'csv',
  'doc',
  'docx',
  'gif',
  'jpeg',
  'jpg',
  'markdown',
  'md',
  'mp4',
  'pdf',
  'png',
  'ppt',
  'pptx',
  'rtf',
  'txt',
  'webp',
  'xls',
  'xlsx',
])
const attachmentAccept = [
  'image/png',
  'image/jpeg',
  'image/gif',
  'image/webp',
  '.pdf',
  '.doc',
  '.docx',
  '.xls',
  '.xlsx',
  '.ppt',
  '.pptx',
  '.txt',
  '.md',
  '.markdown',
  '.mp4',
  '.csv',
  '.rtf',
].join(',')

const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const messages = ref<CommunityChatMessage[]>([])
const loadingHistory = ref(false)
const loadingOlderHistory = ref(false)
const hasOlderHistory = ref(true)
const sending = ref(false)
const uploadProgress = ref<{ percent: number | null } | null>(null)
const draft = ref('')
const selectedFile = ref<File | null>(null)
const selectedFilePreview = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)
const messageInputRef = ref<HTMLTextAreaElement | null>(null)
const draftSelectionStart = ref<number | null>(null)
const draftSelectionEnd = ref<number | null>(null)
const messageListRef = ref<HTMLElement | null>(null)
const messageListPinnedToBottom = ref(true)
const newMessageCount = ref(0)
const searchOpen = ref(false)
const searchInputRef = ref<HTMLInputElement | null>(null)
const searchQuery = ref('')
const activeSearchQuery = ref('')
const searchResults = ref<CommunityChatMessage[]>([])
const searchContextMessages = ref<CommunityChatMessage[]>([])
const searchTotal = ref(0)
const searchPage = ref(1)
const currentSearchResultIndex = ref(-1)
const searching = ref(false)
const loadingSearchContext = ref(false)
const loadingMoreSearchResults = ref(false)
const isDraggingFile = ref(false)
const deletingMessageId = ref<number | null>(null)
const previewImage = ref<CommunityChatMessage | null>(null)
const previewVideo = ref<CommunityChatMessage | null>(null)
const directDialogOpen = ref(false)
const directUnread = ref(false)
const directConversations = ref<CommunityChatDirectConversation[]>([])
const directMessages = ref<CommunityChatMessage[]>([])
const loadingDirectConversations = ref(false)
const loadingDirectMessages = ref(false)
const loadingOlderDirectMessages = ref(false)
const hasOlderDirectMessages = ref(true)
const directMessageListPinnedToBottom = ref(true)
const newDirectMessageCount = ref(0)
const sendingDirect = ref(false)
const directUploadProgress = ref<{ percent: number | null; userId?: number } | null>(null)
const directDraft = ref('')
const directDrafts = new Map<number, string>()
const selectedDirectFile = ref<File | null>(null)
const selectedDirectFilePreview = ref('')
const directFileInputRef = ref<HTMLInputElement | null>(null)
const directMessageInputRef = ref<HTMLTextAreaElement | null>(null)
const selectedDirectUserId = ref<number | null>(null)
const selectedDirectUser = ref<CommunityChatDirectUser | null>(null)
const directUserSearch = ref('')
const directUserResults = ref<CommunityChatDirectUser[]>([])
const loadingDirectUsers = ref(false)
const directConversationPage = ref(1)
const directConversationTotal = ref(0)
const loadingMoreDirectConversations = ref(false)
const directMessageListRef = ref<HTMLElement | null>(null)
const deletingDirectMessageId = ref<number | null>(null)

let attachmentSessionTimer: number | null = null
let attachmentSessionRequest: Promise<void> | null = null
const failedAttachments = new Set<HTMLImageElement | HTMLVideoElement>()
let componentUnmounted = false
let historyLoaded = false
let directMessagesRequestSequence = 0
let groupBackfillRunning = false
let directBackfillRunning = false
let lastDirectReadKey = ''
let directUserSearchTimer: number | null = null
let directUserSearchSequence = 0
let directConversationRefreshPending = false
let directConversationRefreshPromise: Promise<void> | null = null
let messageSearchSequence = 0
let searchContextSequence = 0
let searchPreviousScrollTop = 0

const realtime = useCommunityChatRealtime({
  onEvent: handleSocketEvent,
  onConnected: () => {
    void recoverAfterRealtimeConnect()
  },
})
const socketConnected = realtime.connected
const socketConnecting = realtime.connecting

const currentUserId = computed(() => authStore.user?.id ?? 0)
const canSend = computed(() => {
  return selectedFile.value !== null || draft.value.trim().length > 0
})

const searchActive = computed(() => activeSearchQuery.value.length > 0)
const displayedMessages = computed(() => searchActive.value ? searchContextMessages.value : messages.value)
const currentSearchResult = computed(() => searchResults.value[currentSearchResultIndex.value] ?? null)
const canNavigatePreviousSearchResult = computed(() => currentSearchResultIndex.value > 0)
const canNavigateNextSearchResult = computed(() => {
  return currentSearchResultIndex.value >= 0 && currentSearchResultIndex.value < searchTotal.value - 1
})

const selectedDirectConversation = computed(() => {
  if (!selectedDirectUserId.value) return null
  return directConversations.value.find((conversation) => conversation.user_id === selectedDirectUserId.value) ?? null
})

const visibleDirectUsers = computed<Array<CommunityChatDirectConversation | CommunityChatDirectUser>>(() => {
  return directUserSearch.value.trim() ? directUserResults.value : directConversations.value
})

const hasMoreDirectConversations = computed(() => directConversations.value.length < directConversationTotal.value)

const directThreadTitle = computed(() => {
  if (!authStore.isAdmin) return t('communityChat.owner')
  return selectedDirectConversation.value?.username || selectedDirectUser.value?.username || t('communityChat.directSelectConversation')
})

const directThreadAvatar = computed(() => {
  return authStore.isAdmin ? selectedDirectConversation.value?.avatar_url || selectedDirectUser.value?.avatar_url || '' : ''
})

const canSendDirect = computed(() => {
  if (sendingDirect.value || (directDraft.value.trim().length === 0 && !selectedDirectFile.value)) return false
  return !authStore.isAdmin || Boolean(selectedDirectUserId.value)
})

watch(directUserSearch, (value) => {
  if (directUserSearchTimer) window.clearTimeout(directUserSearchTimer)
  if (!value.trim()) {
    directUserResults.value = []
    loadingDirectUsers.value = false
    return
  }
  directUserSearchTimer = window.setTimeout(() => {
    void searchDirectUsers(value)
  }, 250)
})

watch(draft, () => {
  void nextTick(resizeMessageInput)
})

watch(directDraft, () => {
  void nextTick(resizeDirectMessageInput)
})

const connectionStatusLabel = computed(() => {
  if (socketConnected.value) return t('communityChat.connected')
  if (socketConnecting.value) return t('communityChat.connecting')
  return t('communityChat.disconnected')
})

const connectionStatusClass = computed(() => {
  if (socketConnected.value) return 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800/60 dark:bg-emerald-900/20 dark:text-emerald-300'
  if (socketConnecting.value) return 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-800/60 dark:bg-amber-900/20 dark:text-amber-300'
  return 'border-red-200 bg-red-50 text-red-700 dark:border-red-800/60 dark:bg-red-900/20 dark:text-red-300'
})

const connectionDotClass = computed(() => {
  if (socketConnected.value) return 'bg-emerald-500'
  if (socketConnecting.value) return 'bg-amber-500'
  return 'bg-red-500'
})

async function loadHistory(): Promise<void> {
  loadingHistory.value = true
  let loaded = false
  try {
    const result = await communityChatAPI.listMessages(1, historyPageSize)
    messages.value = mergeMessages(messages.value, result.items)
    hasOlderHistory.value = result.total > result.items.length
    historyLoaded = true
    loaded = true
  } catch (error) {
    console.error('Failed to load community chat messages:', error)
    appStore.showError(t('communityChat.loadFailed'))
  } finally {
    loadingHistory.value = false
  }
  if (loaded) {
    await scrollToBottom()
  }
}

async function toggleSearch(): Promise<void> {
  if (searchOpen.value) {
    await closeSearch()
    return
  }
  searchOpen.value = true
  await nextTick()
  searchInputRef.value?.focus()
}

async function clearSearch(): Promise<void> {
  const restorePosition = searchActive.value
  searchQuery.value = ''
  activeSearchQuery.value = ''
  searchResults.value = []
  searchContextMessages.value = []
  searchTotal.value = 0
  searchPage.value = 1
  currentSearchResultIndex.value = -1
  messageSearchSequence += 1
  searchContextSequence += 1
  searching.value = false
  loadingMoreSearchResults.value = false
  loadingSearchContext.value = false
  await nextTick()
  if (restorePosition && messageListRef.value) {
    messageListRef.value.scrollTop = searchPreviousScrollTop
  }
  searchInputRef.value?.focus()
}

async function closeSearch(): Promise<void> {
  searchOpen.value = false
  await clearSearch()
}

async function runSearch(): Promise<void> {
  const query = searchQuery.value.trim()
  if (!query || searching.value) return
  const sequence = ++messageSearchSequence
  searching.value = true
  try {
    const result = await communityChatAPI.searchMessages(query, 1, searchPageSize)
    if (sequence !== messageSearchSequence) return
    if (!searchActive.value) searchPreviousScrollTop = messageListRef.value?.scrollTop ?? 0
    activeSearchQuery.value = query
    searchResults.value = result.items
    searchContextMessages.value = []
    searchTotal.value = result.total
    searchPage.value = 1
    currentSearchResultIndex.value = -1
    if (result.items.length > 0) {
      await showSearchResult(0, 'auto')
    }
  } catch (error) {
    console.error('Failed to search community chat messages:', error)
    appStore.showError(t('communityChat.searchFailed'))
  } finally {
    if (sequence === messageSearchSequence) searching.value = false
  }
}

async function loadMoreSearchResults(): Promise<void> {
  if (!searchActive.value || loadingMoreSearchResults.value || searchResults.value.length >= searchTotal.value) return
  loadingMoreSearchResults.value = true
  const nextPage = searchPage.value + 1
  const query = activeSearchQuery.value
  try {
    const result = await communityChatAPI.searchMessages(query, nextPage, searchPageSize)
    if (query !== activeSearchQuery.value) return
    const existingIDs = new Set(searchResults.value.map((message) => message.id))
    searchResults.value.push(...result.items.filter((message) => !existingIDs.has(message.id)))
    searchTotal.value = result.total
    searchPage.value = nextPage
  } catch (error) {
    console.error('Failed to load more community chat search results:', error)
    appStore.showError(t('communityChat.searchFailed'))
  } finally {
    loadingMoreSearchResults.value = false
  }
}

async function navigateSearchResults(offset: -1 | 1): Promise<void> {
  if (!searchActive.value || currentSearchResultIndex.value < 0 || loadingSearchContext.value || loadingMoreSearchResults.value) return
  const nextIndex = currentSearchResultIndex.value + offset
  if (nextIndex < 0 || nextIndex >= searchTotal.value) return
  if (nextIndex >= searchResults.value.length) {
    await loadMoreSearchResults()
  }
  if (nextIndex >= searchResults.value.length) return
  try {
    await showSearchResult(nextIndex)
  } catch (error) {
    console.error('Failed to load community chat search result context:', error)
    appStore.showError(t('communityChat.searchFailed'))
  }
}

async function showSearchResult(index: number, behavior: 'auto' | 'smooth' = 'smooth'): Promise<void> {
  const message = searchResults.value[index]
  if (!message) return
  const sequence = ++searchContextSequence
  loadingSearchContext.value = true
  try {
    const [before, after] = await Promise.all([
      communityChatAPI.listMessagesBefore(message.id, searchContextRadius),
      communityChatAPI.listMessagesAfter(message.id, searchContextRadius),
    ])
    if (sequence !== searchContextSequence) return
    searchContextMessages.value = mergeMessages(before, [message, ...after])
    currentSearchResultIndex.value = index
    loadingSearchContext.value = false
    await scrollCurrentSearchResult(behavior)
  } finally {
    if (sequence === searchContextSequence) loadingSearchContext.value = false
  }
}

async function scrollCurrentSearchResult(behavior: 'auto' | 'smooth' = 'auto'): Promise<void> {
  const resultID = currentSearchResult.value?.id
  if (!resultID) return
  await nextTick()
  await waitForFrame()
  if (currentSearchResult.value?.id !== resultID) return

  const container = messageListRef.value
  const messageElement = container?.querySelector<HTMLElement>(`#community-message-${resultID}`)
  if (!container || !messageElement) return
  const target = messageElement.querySelector<HTMLElement>('.community-message-bubble .community-search-keyword-current')
    ?? messageElement.querySelector<HTMLElement>('.community-message-meta .community-search-keyword-current')
    ?? messageElement
  const containerRect = container.getBoundingClientRect()
  const targetRect = target.getBoundingClientRect()
  const top = centeredCommunityChatSearchScrollTop({
    containerScrollTop: container.scrollTop,
    containerTop: containerRect.top,
    containerHeight: container.clientHeight,
    targetTop: targetRect.top,
    targetHeight: targetRect.height,
  })
  container.scrollTo({ top, behavior })
}

function isCurrentSearchResult(message: CommunityChatMessage): boolean {
  return searchActive.value && currentSearchResult.value?.id === message.id
}

function searchTextSegments(text: string) {
  return splitCommunityChatSearchText(text, activeSearchQuery.value)
}

function shouldShowSearchAttachmentName(message: CommunityChatMessage): boolean {
  if (!searchActive.value || (!isImageMessage(message) && !isVideoMessage(message))) return false
  return hasCommunityChatSearchMatch(attachmentFileName(message), activeSearchQuery.value)
}

function removeSearchResult(id: number): void {
  const removedIndex = searchResults.value.findIndex((message) => message.id === id)
  searchContextMessages.value = searchContextMessages.value.filter((message) => message.id !== id)
  if (removedIndex < 0) return
  const removedCurrentResult = removedIndex === currentSearchResultIndex.value
  searchResults.value = searchResults.value.filter((message) => message.id !== id)
  searchTotal.value = Math.max(0, searchTotal.value - 1)
  if (searchResults.value.length === 0) {
    currentSearchResultIndex.value = -1
  } else if (removedIndex < currentSearchResultIndex.value) {
    currentSearchResultIndex.value -= 1
  } else {
    currentSearchResultIndex.value = Math.min(currentSearchResultIndex.value, searchResults.value.length - 1)
  }
  if (removedCurrentResult && currentSearchResultIndex.value >= 0) {
    void showSearchResult(currentSearchResultIndex.value).catch((error) => {
      console.error('Failed to restore community chat search result context after deletion:', error)
      appStore.showError(t('communityChat.searchFailed'))
    })
  }
}

async function refreshDirectUnread(): Promise<void> {
  try {
    const summary = await communityChatAPI.getDirectUnread()
    directUnread.value = summary.total > 0
    if (authStore.isAdmin && directConversations.value.length > 0) {
      const unreadByUser = new Map(summary.conversations.map((item) => [item.user_id, item.unread_count]))
      directConversations.value = directConversations.value.map((conversation) => ({
        ...conversation,
        unread_count: unreadByUser.get(conversation.user_id) ?? 0,
      }))
    }
  } catch (error) {
    console.error('Failed to restore community direct chat notification:', error)
  }
}

function handleAttachmentError(event: Event): void {
  const element = event.currentTarget
  if (element instanceof HTMLImageElement || element instanceof HTMLVideoElement) failedAttachments.add(element)
}

function renewAttachmentSession(): Promise<void> {
  if (componentUnmounted) return Promise.resolve()
  if (attachmentSessionRequest) return attachmentSessionRequest
  if (attachmentSessionTimer) window.clearTimeout(attachmentSessionTimer)
  attachmentSessionTimer = null
  attachmentSessionRequest = (async () => {
    let retryDelay = 30_000
    try {
      await communityChatAPI.establishAttachmentSession()
      if (componentUnmounted) return
      retryDelay = 4 * 60 * 1000
      for (const element of failedAttachments) {
        if (!element.isConnected) continue
        if (element instanceof HTMLVideoElement) element.load()
        else {
          const source = element.src
          element.src = source
        }
      }
      failedAttachments.clear()
    } catch (error) {
      console.error('Failed to renew community chat attachment session:', error)
    } finally {
      if (!componentUnmounted) attachmentSessionTimer = window.setTimeout(() => { void renewAttachmentSession() }, retryDelay)
    }
  })().finally(() => { attachmentSessionRequest = null })
  return attachmentSessionRequest
}

async function initializeCommunityChat(): Promise<void> {
  void renewAttachmentSession()
  if (componentUnmounted) return
  await loadHistory()
  if (componentUnmounted) return
  await refreshDirectUnread()
}

function handleSocketEvent(event: CommunityChatEvent): void {
  if (event.type === 'message_created' && event.message) {
    const isNewMessage = !messages.value.some((message) => message.id === event.message?.id)
    const isOwnMessage = event.message.user_id === currentUserId.value
    const shouldFollow = !searchActive.value && shouldFollowCommunityChatMessage(messageListPinnedToBottom.value, isOwnMessage)
    upsertMessage(event.message)
    if (shouldFollow) {
      void scrollToBottom(isOwnMessage)
    } else if (isNewMessage) {
      newMessageCount.value += 1
    }
    return
  }

  if (event.type === 'message_deleted' && event.id) {
    messages.value = messages.value.filter((message) => message.id !== event.id)
    removeSearchResult(event.id)
    return
  }

  if (event.type === 'direct_message_created' && event.message) {
    void handleDirectSocketMessage(event)
    return
  }

  if (event.type === 'direct_message_deleted' && event.id) {
    directMessages.value = directMessages.value.filter((message) => message.id !== event.id)
    void refreshDirectUnread()
    if (authStore.isAdmin && directDialogOpen.value) {
      void refreshDirectConversations()
    }
  }
}

async function recoverAfterRealtimeConnect(): Promise<void> {
  void renewAttachmentSession()
  // Capture both history cursors before any request yields to new live events.
  await Promise.all([
    historyLoaded ? backfillGroupMessages() : Promise.resolve(),
    directDialogOpen.value && (!authStore.isAdmin || selectedDirectUserId.value)
      ? backfillDirectMessages()
      : Promise.resolve(),
    refreshDirectUnread(),
  ])
}

async function backfillGroupMessages(): Promise<void> {
  if (groupBackfillRunning) return
  groupBackfillRunning = true
  const wasPinned = messageListPinnedToBottom.value
  let cursor = messages.value[messages.value.length - 1]?.id ?? 0
  let incomingCount = 0
  let hasOwnMessage = false
  try {
    if (cursor === 0) {
      await loadHistory()
      return
    }
    while (cursor > 0) {
      const batch = await communityChatAPI.listMessagesAfter(cursor, 100)
      for (const message of batch) {
        if (!messages.value.some((item) => item.id === message.id)) {
          if (message.user_id === currentUserId.value) hasOwnMessage = true
          else incomingCount += 1
        }
      }
      messages.value = mergeMessages(messages.value, batch)
      if (batch.length < 100) break
      cursor = batch[batch.length - 1]?.id ?? cursor
    }
    if (wasPinned || hasOwnMessage) await scrollToBottom(true)
    else newMessageCount.value += incomingCount
  } catch (error) {
    console.error('Failed to recover missed community chat messages:', error)
  } finally {
    groupBackfillRunning = false
  }
}

function mergeMessages(current: CommunityChatMessage[], incoming: CommunityChatMessage[]): CommunityChatMessage[] {
  const merged = new Map(current.map((message) => [message.id, message]))
  for (const message of incoming) merged.set(message.id, message)
  return Array.from(merged.values()).sort((a, b) => a.id - b.id)
}

function upsertMessage(message: CommunityChatMessage): void {
  const index = messages.value.findIndex((item) => item.id === message.id)
  if (index >= 0) {
    messages.value.splice(index, 1, message)
  } else {
    messages.value.push(message)
    messages.value.sort((a, b) => a.id - b.id)
  }
}

function upsertDirectMessage(message: CommunityChatMessage): void {
  const index = directMessages.value.findIndex((item) => item.id === message.id)
  if (index >= 0) {
    directMessages.value.splice(index, 1, message)
  } else {
    directMessages.value.push(message)
    directMessages.value.sort((a, b) => a.id - b.id)
  }
}

async function handleDirectSocketMessage(event: CommunityChatEvent): Promise<void> {
  if (!event.message) return
  const conversationUserId = event.conversation_user_id || event.message?.conversation_user_id || 0
  const isOwnMessage = event.message.user_id === currentUserId.value
  if (authStore.isAdmin) {
    if (selectedDirectUserId.value !== conversationUserId) {
      if (directDialogOpen.value) await refreshDirectConversations()
      await refreshDirectUnread()
      return
    }
  }
  if (!directDialogOpen.value) {
    await refreshDirectUnread()
    return
  }
  const shouldFollow = shouldFollowCommunityChatMessage(directMessageListPinnedToBottom.value, isOwnMessage)
  const isNewMessage = !directMessages.value.some((message) => message.id === event.message?.id)
  upsertDirectMessage(event.message)
  if (shouldFollow) {
    await scrollDirectToBottom(true)
  } else if (isNewMessage) {
    newDirectMessageCount.value += 1
    await refreshDirectUnread()
  }
  if (authStore.isAdmin) await refreshDirectConversations()
}

function uploadProgressText(percent: number | null): string {
  if (percent === null) return t('communityChat.uploading')
  if (percent === 100) return t('communityChat.uploadProcessing')
  return t('communityChat.uploadProgress', { percent })
}

function updateUploadProgress(progress: { percent: number | null } | null, event: AxiosProgressEvent): void {
  if (progress) progress.percent = event.total ? Math.min(100, Math.floor(event.loaded / event.total * 100)) : null
}

function handleComposerEnter(event: KeyboardEvent, privateChat: boolean): void {
  if (event.isComposing || event.keyCode === 229) return
  event.preventDefault()
  if (privateChat) void handleDirectSend()
  else void handleSend()
}

async function handleSend(): Promise<void> {
  if (!canSend.value || sending.value) return

  const submittedDraft = draft.value
  const submittedFile = selectedFile.value
  const submittedReplyID = replyTarget.value?.id
  sending.value = true
  try {
    let message: CommunityChatMessage
    if (submittedFile) {
      uploadProgress.value = { percent: null }
      message = await communityChatAPI.uploadFileMessage(submittedFile, submittedDraft, submittedReplyID, (event) => updateUploadProgress(uploadProgress.value, event))
      if (selectedFile.value === submittedFile) clearSelectedFile()
    } else {
      message = await communityChatAPI.sendTextMessage(submittedDraft, submittedReplyID)
    }
    upsertMessage(message)
    if (draft.value === submittedDraft) draft.value = ''
    if (replyTarget.value?.id === submittedReplyID) clearReplyTarget()
    if (searchActive.value) await closeSearch()
    await scrollToBottom(true)
  } catch (error) {
    console.error('Failed to send community chat message:', error)
    appStore.showError(t('communityChat.sendFailed'))
  } finally {
    sending.value = false
    uploadProgress.value = null
  }
}

async function openDirectDialog(): Promise<void> {
  directDialogOpen.value = true
  await nextTick()
  resizeDirectMessageInput()
  if (authStore.isAdmin) {
    const conversationsLoaded = await loadDirectConversations()
    if (!conversationsLoaded) return
    if (!selectedDirectUserId.value && directConversations.value.length > 0) {
      const messagesLoaded = await selectDirectConversation(directConversations.value[0])
      if (!messagesLoaded) return
    } else if (selectedDirectUserId.value) {
      const messagesLoaded = await loadDirectMessages(selectedDirectUserId.value)
      if (!messagesLoaded) return
    }
    return
  }
  selectedDirectUserId.value = null
  await loadDirectMessages()
}

function closeDirectDialog(): void {
  directDialogOpen.value = false
  directDraft.value = ''
  directUserSearch.value = ''
  clearSelectedDirectFile()
  directMessagesRequestSequence += 1
  newDirectMessageCount.value = 0
}

async function loadDirectConversations(reset = true): Promise<boolean> {
  if (!authStore.isAdmin) return false
  if (directConversationRefreshPromise) await directConversationRefreshPromise
  if (reset) loadingDirectConversations.value = true
  else loadingMoreDirectConversations.value = true
  try {
    const page = reset ? 1 : directConversationPage.value + 1
    const result = await communityChatAPI.listDirectConversations(page, directConversationPageSize)
    directConversations.value = reset
      ? result.items
      : mergeDirectConversations(directConversations.value, result.items)
    directConversationPage.value = page
    directConversationTotal.value = result.total
    if (directConversations.value.some((conversation) => conversation.unread_count > 0)) {
      directUnread.value = true
    }
    return true
  } catch (error) {
    console.error('Failed to load direct conversations:', error)
    appStore.showError(t('communityChat.directLoadFailed'))
    return false
  } finally {
    if (reset) loadingDirectConversations.value = false
    else loadingMoreDirectConversations.value = false
    if (directConversationRefreshPending) void refreshDirectConversations()
  }
}

function refreshDirectConversations(): Promise<void> {
  if (!authStore.isAdmin || !directDialogOpen.value || componentUnmounted) return Promise.resolve()
  directConversationRefreshPending = true
  if (directConversationRefreshPromise) return directConversationRefreshPromise
  if (loadingDirectConversations.value || loadingMoreDirectConversations.value) return Promise.resolve()
  directConversationRefreshPromise = (async () => {
    try {
      do {
        directConversationRefreshPending = false
        const pages = await Promise.all(Array.from({ length: directConversationPage.value }, (_, index) =>
          communityChatAPI.listDirectConversations(index + 1, directConversationPageSize),
        ))
        if (componentUnmounted || !directDialogOpen.value) return
        // Keep already loaded rows and their positions while updating their contents.
        directConversations.value = mergeDirectConversations(directConversations.value, pages.flatMap((page) => page.items))
        directConversationTotal.value = pages[0].total
        await refreshDirectUnread()
      } while (directConversationRefreshPending && directDialogOpen.value)
    } catch (error) {
      console.error('Failed to refresh direct conversations:', error)
    }
  })().finally(() => { directConversationRefreshPromise = null })
  return directConversationRefreshPromise
}

async function loadMoreDirectConversations(): Promise<void> {
  if (!hasMoreDirectConversations.value || loadingMoreDirectConversations.value) return
  await loadDirectConversations(false)
}

function mergeDirectConversations(current: CommunityChatDirectConversation[], incoming: CommunityChatDirectConversation[]): CommunityChatDirectConversation[] {
  const merged = new Map(current.map((conversation) => [conversation.user_id, conversation]))
  for (const conversation of incoming) merged.set(conversation.user_id, conversation)
  return Array.from(merged.values())
}

async function searchDirectUsers(search: string): Promise<void> {
  const normalized = search.trim()
  if (!normalized) return
  const sequence = ++directUserSearchSequence
  loadingDirectUsers.value = true
  try {
    const result = await communityChatAPI.searchDirectUsers(normalized, 1, 30)
    if (sequence !== directUserSearchSequence || directUserSearch.value.trim() !== normalized) return
    directUserResults.value = result.items
  } catch (error) {
    if (sequence !== directUserSearchSequence) return
    console.error('Failed to search direct message users:', error)
    appStore.showError(t('communityChat.directSearchFailed'))
  } finally {
    if (sequence === directUserSearchSequence) loadingDirectUsers.value = false
  }
}

async function loadDirectMessages(userId?: number): Promise<boolean> {
  const requestSequence = ++directMessagesRequestSequence
  const requestedUserID = userId ?? null
  loadingDirectMessages.value = true
  // Only live messages received during this load may be merged into its snapshot.
  directMessages.value = []
  try {
    const result = await communityChatAPI.listDirectMessages(userId, 1, historyPageSize)
    if (requestSequence !== directMessagesRequestSequence) return false
    if (authStore.isAdmin && selectedDirectUserId.value !== requestedUserID) return false
    directMessages.value = mergeMessages(result.items, directMessages.value)
    hasOlderDirectMessages.value = result.total > result.items.length
    directMessageListPinnedToBottom.value = true
    newDirectMessageCount.value = 0
    loadingDirectMessages.value = false
    await scrollDirectToBottom(true)
    return true
  } catch (error) {
    if (requestSequence !== directMessagesRequestSequence) return false
    console.error('Failed to load direct messages:', error)
    appStore.showError(t('communityChat.directLoadFailed'))
    return false
  } finally {
    if (requestSequence === directMessagesRequestSequence) loadingDirectMessages.value = false
  }
}

async function selectDirectConversation(conversation: CommunityChatDirectConversation): Promise<boolean> {
  switchDirectDraft(conversation.user_id)
  clearSelectedDirectFile()
  selectedDirectUser.value = {
    user_id: conversation.user_id,
    username: conversation.username,
    avatar_url: conversation.avatar_url,
  }
  selectedDirectUserId.value = conversation.user_id
  directMessages.value = []
  newDirectMessageCount.value = 0
  return loadDirectMessages(conversation.user_id)
}

async function selectDirectUser(user: CommunityChatDirectConversation | CommunityChatDirectUser): Promise<boolean> {
  const conversation = directConversations.value.find((item) => item.user_id === user.user_id)
  if (conversation) return selectDirectConversation(conversation)
  switchDirectDraft(user.user_id)
  clearSelectedDirectFile()
  selectedDirectUser.value = {
    user_id: user.user_id,
    username: user.username,
    avatar_url: user.avatar_url,
  }
  selectedDirectUserId.value = user.user_id
  directMessages.value = []
  newDirectMessageCount.value = 0
  return loadDirectMessages(user.user_id)
}

function switchDirectDraft(userId: number): void {
  if (selectedDirectUserId.value === userId) return
  if (selectedDirectUserId.value) directDrafts.set(selectedDirectUserId.value, directDraft.value)
  directDraft.value = directDrafts.get(userId) ?? ''
}

async function handleDirectSend(): Promise<void> {
  if (!canSendDirect.value) return
  const submittedDraft = directDraft.value
  const submittedFile = selectedDirectFile.value
  const submittedUserID = authStore.isAdmin ? selectedDirectUserId.value ?? undefined : undefined
  sendingDirect.value = true
  if (submittedFile) directUploadProgress.value = { percent: null, userId: submittedUserID }
  try {
    const message = submittedFile
      ? await communityChatAPI.uploadDirectFileMessage(submittedFile, submittedDraft, submittedUserID, (event) => updateUploadProgress(directUploadProgress.value, event))
      : await communityChatAPI.sendDirectTextMessage(submittedDraft, submittedUserID)
    if (selectedDirectFile.value === submittedFile) clearSelectedDirectFile()
    if (!authStore.isAdmin || selectedDirectUserId.value === submittedUserID) {
      if (directDraft.value === submittedDraft) directDraft.value = ''
      upsertDirectMessage(message)
      await scrollDirectToBottom(true)
    } else if (submittedUserID && directDrafts.get(submittedUserID) === submittedDraft) {
      directDrafts.delete(submittedUserID)
    }
    if (authStore.isAdmin) {
      await refreshDirectConversations()
    }
  } catch (error) {
    console.error('Failed to send direct message:', error)
    appStore.showError(t('communityChat.directSendFailed'))
  } finally {
    sendingDirect.value = false
    directUploadProgress.value = null
  }
}

async function backfillDirectMessages(): Promise<void> {
  if (directBackfillRunning || loadingDirectMessages.value) return
  const requestedUserID = authStore.isAdmin ? selectedDirectUserId.value ?? undefined : undefined
  if (authStore.isAdmin && !requestedUserID) return
  directBackfillRunning = true
  const wasPinned = directMessageListPinnedToBottom.value
  let cursor = directMessages.value[directMessages.value.length - 1]?.id ?? 0
  let incomingCount = 0
  let hasOwnMessage = false
  let recovered = false
  try {
    if (cursor === 0) {
      recovered = await loadDirectMessages(requestedUserID)
      return
    }
    while (cursor > 0) {
      const batch = await communityChatAPI.listDirectMessagesAfter(cursor, requestedUserID, 100)
      if (authStore.isAdmin && selectedDirectUserId.value !== requestedUserID) return
      for (const message of batch) {
        if (!directMessages.value.some((item) => item.id === message.id)) {
          if (message.user_id === currentUserId.value) hasOwnMessage = true
          else incomingCount += 1
        }
      }
      directMessages.value = mergeMessages(directMessages.value, batch)
      if (batch.length < 100) break
      cursor = batch[batch.length - 1]?.id ?? cursor
    }
    if (wasPinned || hasOwnMessage) await scrollDirectToBottom(true)
    else newDirectMessageCount.value += incomingCount
    recovered = true
  } catch (error) {
    console.error('Failed to recover missed direct messages:', error)
  } finally {
    directBackfillRunning = false
    if (recovered && (!authStore.isAdmin || selectedDirectUserId.value === requestedUserID)) {
      await markActiveDirectConversationRead()
    }
  }
}

function handleFileSelected(event: Event): void {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  target.value = ''
  if (!file) return

  setSelectedFile(file)
}

function handleDragEnter(event: DragEvent): void {
  if (!hasDraggedFiles(event.dataTransfer)) return
  isDraggingFile.value = true
}

function handleDragOver(event: DragEvent): void {
  if (!hasDraggedFiles(event.dataTransfer)) return
  isDraggingFile.value = true
  if (event.dataTransfer) event.dataTransfer.dropEffect = 'copy'
}

function handleDragLeave(event: DragEvent): void {
  if (!hasDraggedFiles(event.dataTransfer)) return
  const dropRegion = event.currentTarget as HTMLElement | null
  const nextTarget = event.relatedTarget as Node | null
  if (dropRegion && nextTarget && dropRegion.contains(nextTarget)) return
  isDraggingFile.value = false
}

function handleDrop(event: DragEvent): void {
  isDraggingFile.value = false
  const file = firstCommunityChatDroppedFile(event.dataTransfer)
  if (!file) return
  setSelectedFile(file)
  void nextTick(() => messageInputRef.value?.focus())
}

function hasDraggedFiles(data: DataTransfer | null): boolean {
  return Boolean(data && Array.from(data.types).includes('Files'))
}

function handleDirectFileSelected(event: Event): void {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  target.value = ''
  if (file) setSelectedDirectFile(file)
}

function handlePaste(event: ClipboardEvent): void {
  const file = findPastedImageFile(event.clipboardData)
  if (!file) return

  event.preventDefault()
  setSelectedFile(file)
}

function handleDirectPaste(event: ClipboardEvent): void {
  const file = findPastedImageFile(event.clipboardData)
  if (!file) return
  event.preventDefault()
  setSelectedDirectFile(file)
}

function findPastedImageFile(data: DataTransfer | null): File | null {
  if (!data) return null

  for (const item of Array.from(data.items)) {
    if (item.kind !== 'file' || !item.type.startsWith('image/')) continue
    const file = item.getAsFile()
    if (file) return file
  }

  return Array.from(data.files).find((file) => file.type.startsWith('image/')) ?? null
}

function setSelectedFile(file: File): void {
  if (!isAllowedAttachmentFile(file)) {
    appStore.showError(t('communityChat.invalidFile'))
    return
  }
  if (file.size > maxAttachmentBytes) {
    appStore.showError(t('communityChat.fileTooLarge'))
    return
  }

  clearSelectedFile()
  selectedFile.value = file
  selectedFilePreview.value = file.type.startsWith('image/') ? URL.createObjectURL(file) : ''
}

function clearSelectedFile(): void {
  if (selectedFilePreview.value) {
    URL.revokeObjectURL(selectedFilePreview.value)
  }
  selectedFile.value = null
  selectedFilePreview.value = ''
}

function setSelectedDirectFile(file: File): void {
  if (!isAllowedAttachmentFile(file)) {
    appStore.showError(t('communityChat.invalidFile'))
    return
  }
  if (file.size > maxAttachmentBytes) {
    appStore.showError(t('communityChat.fileTooLarge'))
    return
  }

  clearSelectedDirectFile()
  selectedDirectFile.value = file
  selectedDirectFilePreview.value = file.type.startsWith('image/') ? URL.createObjectURL(file) : ''
}

function clearSelectedDirectFile(): void {
  if (selectedDirectFilePreview.value) URL.revokeObjectURL(selectedDirectFilePreview.value)
  selectedDirectFile.value = null
  selectedDirectFilePreview.value = ''
}

function isAllowedAttachmentFile(file: File): boolean {
  if (file.type.startsWith('image/')) {
    return ['image/png', 'image/jpeg', 'image/gif', 'image/webp'].includes(file.type)
  }
  const ext = file.name.split('.').pop()?.toLowerCase() || ''
  return allowedAttachmentExtensions.has(ext)
}

function rememberDraftSelection(): void {
  const input = messageInputRef.value
  if (!input) return
  draftSelectionStart.value = input.selectionStart
  draftSelectionEnd.value = input.selectionEnd
}

function handleDraftInput(): void {
  rememberDraftSelection()
  resizeMessageInput()
}

function resizeMessageInput(): void {
  const input = messageInputRef.value
  if (!input) return
  resizeCommunityChatTextarea(input)
}

function resizeDirectMessageInput(): void {
  const input = directMessageInputRef.value
  if (!input) return
  resizeCommunityChatTextarea(input)
}

async function insertEmoji(emoji: string): Promise<void> {
  const result = insertEmojiAtSelection(
    draft.value,
    emoji,
    draftSelectionStart.value ?? draft.value.length,
    draftSelectionEnd.value ?? draft.value.length,
    maxTextLength,
  )
  if (!result) return

  draft.value = result.value
  draftSelectionStart.value = result.cursor
  draftSelectionEnd.value = result.cursor
  await nextTick()
  messageInputRef.value?.setSelectionRange(result.cursor, result.cursor)
  if (!window.matchMedia('(pointer: coarse)').matches) messageInputRef.value?.focus()
}

async function handleDelete(message: CommunityChatMessage): Promise<void> {
  const confirmMessage = message.user_id === currentUserId.value ? t('communityChat.recallConfirm') : t('communityChat.deleteConfirm')
  if (!confirm(confirmMessage)) return
  deletingMessageId.value = message.id
  try {
    await communityChatAPI.deleteMessage(message.id)
    messages.value = messages.value.filter((item) => item.id !== message.id)
    removeSearchResult(message.id)
  } catch (error) {
    console.error('Failed to delete community chat message:', error)
    appStore.showError(t('communityChat.deleteFailed'))
  } finally {
    deletingMessageId.value = null
  }
}

async function handleDeleteDirect(message: CommunityChatMessage): Promise<void> {
  const confirmMessage = message.user_id === currentUserId.value ? t('communityChat.recallConfirm') : t('communityChat.deleteConfirm')
  if (!confirm(confirmMessage)) return
  deletingDirectMessageId.value = message.id
  try {
    await communityChatAPI.deleteDirectMessage(message.id)
    directMessages.value = directMessages.value.filter((item) => item.id !== message.id)
    await refreshDirectUnread()
    if (authStore.isAdmin) await refreshDirectConversations()
  } catch (error) {
    console.error('Failed to delete direct message:', error)
    appStore.showError(t('communityChat.deleteFailed'))
  } finally {
    deletingDirectMessageId.value = null
  }
}

function handleMessageListScroll(): void {
  const el = messageListRef.value
  if (!el) return
  if (searchActive.value) return
  messageListPinnedToBottom.value = isCommunityChatNearBottom(el)
  if (messageListPinnedToBottom.value) {
    newMessageCount.value = 0
    markGroupNotificationsSeen()
  }
  if (el.scrollTop <= 80) void loadOlderHistory()
}

async function loadOlderHistory(): Promise<void> {
  const el = messageListRef.value
  const firstID = messages.value[0]?.id
  if (!el || !firstID || !hasOlderHistory.value || loadingOlderHistory.value || loadingHistory.value) return
  loadingOlderHistory.value = true
  const previousHeight = el.scrollHeight
  const previousTop = el.scrollTop
  try {
    const older = await communityChatAPI.listMessagesBefore(firstID, historyPageSize)
    hasOlderHistory.value = older.length === historyPageSize
    messages.value = mergeMessages(messages.value, older)
    await nextTick()
    await waitForFrame()
    el.scrollTop = previousTop + (el.scrollHeight - previousHeight)
  } catch (error) {
    console.error('Failed to load older community chat messages:', error)
    appStore.showError(t('communityChat.loadFailed'))
  } finally {
    loadingOlderHistory.value = false
  }
}

function handleMessageMediaLoaded(messageID: number): void {
  if (searchActive.value) {
    if (currentSearchResult.value?.id === messageID) void scrollCurrentSearchResult('auto')
    return
  }
  if (!messageListPinnedToBottom.value) return
  void scrollToBottom(false)
}

function handleDirectMessageMediaLoaded(): void {
  if (!directMessageListPinnedToBottom.value) return
  void scrollDirectToBottom(false)
}

async function scrollToBottom(force = true): Promise<void> {
  await nextTick()
  await waitForFrame()
  if (!force && !messageListPinnedToBottom.value) return
  const el = messageListRef.value
  if (!el) return
  el.scrollTo({ top: el.scrollHeight, behavior: 'auto' })
  messageListPinnedToBottom.value = true
  newMessageCount.value = 0
  markGroupNotificationsSeen()
}

function handleDirectMessageListScroll(): void {
  const el = directMessageListRef.value
  if (!el) return
  directMessageListPinnedToBottom.value = isCommunityChatNearBottom(el)
  if (directMessageListPinnedToBottom.value) {
    newDirectMessageCount.value = 0
    void markActiveDirectConversationRead()
  }
  if (el.scrollTop <= 80) void loadOlderDirectMessages()
}

async function loadOlderDirectMessages(): Promise<void> {
  const el = directMessageListRef.value
  const firstID = directMessages.value[0]?.id
  const requestedUserID = authStore.isAdmin ? selectedDirectUserId.value ?? undefined : undefined
  if (!el || !firstID || !hasOlderDirectMessages.value || loadingOlderDirectMessages.value || loadingDirectMessages.value) return
  loadingOlderDirectMessages.value = true
  const previousHeight = el.scrollHeight
  const previousTop = el.scrollTop
  try {
    const older = await communityChatAPI.listDirectMessagesBefore(firstID, requestedUserID, historyPageSize)
    if (authStore.isAdmin && selectedDirectUserId.value !== requestedUserID) return
    hasOlderDirectMessages.value = older.length === historyPageSize
    directMessages.value = mergeMessages(directMessages.value, older)
    await nextTick()
    await waitForFrame()
    el.scrollTop = previousTop + (el.scrollHeight - previousHeight)
  } catch (error) {
    console.error('Failed to load older direct messages:', error)
    appStore.showError(t('communityChat.directLoadFailed'))
  } finally {
    loadingOlderDirectMessages.value = false
  }
}

async function scrollDirectToBottom(force = true): Promise<void> {
  await nextTick()
  await waitForFrame()
  if (!force && !directMessageListPinnedToBottom.value) return
  const el = directMessageListRef.value
  if (!el) return
  el.scrollTo({ top: el.scrollHeight, behavior: 'auto' })
  directMessageListPinnedToBottom.value = true
  newDirectMessageCount.value = 0
  await markActiveDirectConversationRead()
}

function markGroupNotificationsSeen(): void {
  if (document.visibilityState !== 'visible' || !messageListPinnedToBottom.value) return
  const latestID = messages.value[messages.value.length - 1]?.id ?? 0
  if (currentUserId.value > 0 && latestID > 0) {
    communityChatAPI.markCommunityChatMessagesSeen(currentUserId.value, 'group', latestID)
  }
}

async function markActiveDirectConversationRead(): Promise<void> {
  if (!directDialogOpen.value || loadingDirectMessages.value || directBackfillRunning || !directMessageListPinnedToBottom.value || document.visibilityState !== 'visible') return
  const userID = authStore.isAdmin ? selectedDirectUserId.value ?? undefined : undefined
  if (authStore.isAdmin && !userID) return
  const latestMessageID = directMessages.value[directMessages.value.length - 1]?.id ?? 0
  if (latestMessageID <= 0) return
  const readKey = `${userID ?? currentUserId.value}:${latestMessageID}`
  if (lastDirectReadKey === readKey) return
  lastDirectReadKey = readKey
  try {
    const lastMessageID = await communityChatAPI.markDirectRead(latestMessageID, userID)
    if (authStore.isAdmin && userID) {
      const conversation = directConversations.value.find((item) => item.user_id === userID)
      if (conversation) conversation.unread_count = 0
      directUnread.value = directConversations.value.some((item) => item.unread_count > 0)
    } else {
      directUnread.value = false
    }
    if (currentUserId.value > 0 && lastMessageID > 0) {
      communityChatAPI.markCommunityChatMessagesSeen(currentUserId.value, 'direct', lastMessageID)
    }
    await refreshDirectUnread()
  } catch (error) {
    if (lastDirectReadKey === readKey) lastDirectReadKey = ''
    console.error('Failed to mark direct messages as read:', error)
  }
}

function handleVisibilityChange(): void {
  if (document.visibilityState !== 'visible') return
  void renewAttachmentSession()
  markGroupNotificationsSeen()
  void markActiveDirectConversationRead()
}

function waitForFrame(): Promise<void> {
  return new Promise((resolve) => {
    window.requestAnimationFrame(() => resolve())
  })
}

function avatarInitial(username: string): string {
  const trimmed = username.trim()
  if (!trimmed) return 'U'
  const normalized = trimmed.replace(/\s+/g, '')
  const chars = Array.from(normalized)
  if (chars.length <= 2) return chars.join('').toUpperCase()
  return chars.slice(0, 2).join('').toUpperCase()
}

const replyTarget = ref<CommunityChatMessage | null>(null)

function setReplyTarget(message: CommunityChatMessage): void {
  replyTarget.value = message
}

function clearReplyTarget(): void {
  replyTarget.value = null
}

function canRecall(message: CommunityChatMessage): boolean {
  return message.user_id === currentUserId.value || authStore.isAdmin
}

function messagePreviewText(message: CommunityChatMessage): string {
  const caption = messageDisplayText(message)
  if (isImageMessage(message)) {
    return caption || t('communityChat.imageMessage')
  }
  if (isVideoMessage(message)) {
    return caption || `${t('communityChat.videoMessage')} ${attachmentFileName(message)}`
  }
  if (isFileMessage(message)) {
    return caption || `${t('communityChat.fileMessage')} ${attachmentFileName(message)}`
  }
  return message.content
}

function replySourcePreviewText(message: CommunityChatMessage): string {
  if (message.reply_to_message_type === 'image') {
    return message.reply_to_content
      ? `${t('communityChat.imageMessage')} ${message.reply_to_content}`
      : t('communityChat.imageMessage')
  }
  if (message.reply_to_message_type === 'file') {
    return message.reply_to_content
      ? `${t('communityChat.fileMessage')} ${message.reply_to_content}`
      : t('communityChat.fileMessage')
  }
  return message.reply_to_content
}

function isImageMimeType(mimeType?: string): boolean {
  return Boolean(mimeType?.startsWith('image/'))
}

function isVideoMimeType(mimeType?: string): boolean {
  return mimeType === 'video/mp4'
}

function attachmentURL(message: CommunityChatMessage): string {
  return message.file_url || message.image_url || ''
}

function attachmentMIMEType(message: CommunityChatMessage): string {
  return message.file_mime_type || message.image_mime_type || ''
}

function attachmentSizeBytes(message: CommunityChatMessage): number {
  return message.file_size_bytes || message.image_size_bytes || 0
}

function isImageMessage(message: CommunityChatMessage): boolean {
  return Boolean(attachmentURL(message)) && isImageMimeType(attachmentMIMEType(message))
}

function isVideoMessage(message: CommunityChatMessage): boolean {
  return Boolean(attachmentURL(message)) && isVideoMimeType(attachmentMIMEType(message))
}

function isFileMessage(message: CommunityChatMessage): boolean {
  return Boolean(attachmentURL(message)) && !isImageMessage(message) && !isVideoMessage(message)
}

function shouldShowMessageText(message: CommunityChatMessage): boolean {
  return messageDisplayText(message).length > 0
}

function messageDisplayText(message: CommunityChatMessage): string {
  const content = message.content.trim()
  if (!content) return ''
  if (!attachmentURL(message)) return content
  return content === attachmentFileName(message).trim() ? '' : content
}

function attachmentFileName(message: CommunityChatMessage): string {
  const fileName = message.file_name || originalFileNameFromURL(attachmentURL(message)) || message.content || t('communityChat.attachment')
  return fileName.trim() || t('communityChat.attachment')
}

function originalFileNameFromURL(rawURL: string): string {
  if (!rawURL) return ''
  try {
    const parsed = new URL(rawURL, window.location.origin)
    const name = parsed.searchParams.get('name')?.trim()
    if (name) return name
    const lastPathPart = parsed.pathname.split('/').filter(Boolean).pop()
    return lastPathPart || ''
  } catch {
    return rawURL.split('?')[0]?.split('/').filter(Boolean).pop() || ''
  }
}

function attachmentDownloadURL(message: CommunityChatMessage): string {
  const url = attachmentURL(message)
  if (!url) return ''
  const separator = url.includes('?') ? '&' : '?'
  return `${url}${separator}download=1`
}

function attachmentMeta(message: CommunityChatMessage): string {
  const mimeType = attachmentMIMEType(message)
  const size = attachmentSizeBytes(message)
  const parts = []
  if (mimeType) parts.push(mimeType)
  if (size > 0) parts.push(formatFileSize(size))
  return parts.join(' · ')
}

function openImagePreview(message: CommunityChatMessage): void {
  if (!isImageMessage(message)) return
  previewImage.value = message
}

function closeImagePreview(): void {
  previewImage.value = null
}

function openVideoPreview(message: CommunityChatMessage): void {
  if (!isVideoMessage(message)) return
  previewVideo.value = message
}

function closeVideoPreview(): void {
  previewVideo.value = null
}

async function scrollToMessage(id?: number | null): Promise<void> {
  if (!id) return
  await nextTick()
  const el = document.getElementById(`community-message-${id}`)
  el?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

const messageTimeFormatter = computed(() => new Intl.DateTimeFormat(locale.value, {
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
}))

function formatTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return messageTimeFormatter.value.format(date)
}

function formatFileSize(size: number): string {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

onMounted(() => {
  componentUnmounted = false
  realtime.start()
  document.addEventListener('visibilitychange', handleVisibilityChange)
  window.addEventListener('dragend', resetFileDragState)
  window.addEventListener('drop', resetFileDragState)
  void initializeCommunityChat()
  void nextTick(resizeMessageInput)
})

onBeforeUnmount(() => {
  componentUnmounted = true
  if (attachmentSessionTimer) {
    window.clearTimeout(attachmentSessionTimer)
    attachmentSessionTimer = null
  }
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  failedAttachments.clear()
  window.removeEventListener('dragend', resetFileDragState)
  window.removeEventListener('drop', resetFileDragState)
  realtime.stop()
  clearSelectedFile()
  clearSelectedDirectFile()
  if (directUserSearchTimer) window.clearTimeout(directUserSearchTimer)
})

function resetFileDragState(): void {
  isDraggingFile.value = false
}
</script>

<style scoped>
.community-chat-page {
  @apply min-h-[calc(100vh-8rem)] rounded-[2rem] border border-[#aaa095] bg-[#b9b1a7] p-3 shadow-[0_28px_70px_rgba(15,23,42,0.14)] dark:border-[#252b35] dark:bg-[#0f1319] md:p-4;
}

.community-chat-panel {
  @apply flex h-[calc(100vh-8rem)] min-h-[640px] flex-col overflow-hidden rounded-[1.5rem] border border-[#a99f94] bg-[#c8bfb4] shadow-[0_18px_45px_rgba(15,23,42,0.14)] dark:border-[#252b35] dark:bg-[#121821]/95;
}

.community-chat-header {
  @apply flex flex-col gap-3 border-b border-[#b4aa9f] bg-[#d0c6ba] px-4 py-4 backdrop-blur dark:border-[#252b35] dark:bg-[#121821]/90 md:grid md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] md:items-center md:px-5;
}

.community-chat-logo {
  @apply flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-[#bda997] text-[#7a1f1f] shadow-inner dark:bg-[#31211f] dark:text-[#f0b4a8];
}

.community-chat-status {
  @apply inline-flex w-fit items-center gap-2 rounded-full border px-3 py-1 text-xs font-medium;
}

.community-chat-header-actions {
  @apply flex flex-wrap items-center justify-center gap-2 md:justify-self-end;
}

.community-header-icon-button {
  @apply inline-flex h-9 w-9 items-center justify-center rounded-xl border border-[#9e9184] bg-[#d8cdc0] text-gray-700 transition-colors hover:border-[#8f382f]/45 hover:text-[#8f382f] focus:outline-none focus:ring-2 focus:ring-[#8f382f]/25 dark:border-[#334052] dark:bg-[#18212c] dark:text-gray-200 dark:hover:text-[#f0b4a8];
}

.community-header-icon-button-active {
  @apply border-[#8f382f]/45 bg-[#8f382f] text-white hover:text-white dark:bg-[#7f2d24];
}

.community-search-bar {
  @apply flex items-center gap-2 border-b border-[#b4aa9f] bg-[#c5bbb0] px-4 py-2.5 dark:border-[#252b35] dark:bg-[#121821] md:px-5;
}

.community-search-input-shell {
  @apply flex min-w-0 flex-1 items-center gap-2 rounded-xl border border-[#a99b8c] bg-[#d8cdc0] px-3 focus-within:border-[#8f382f]/55 focus-within:ring-2 focus-within:ring-[#8f382f]/20 dark:border-[#2d3746] dark:bg-[#18212c] dark:focus-within:border-[#b84535]/55;
}

.community-search-input {
  @apply h-10 min-w-0 flex-1 border-0 bg-transparent text-sm text-gray-900 outline-none placeholder:text-gray-500 dark:text-white dark:placeholder:text-gray-400;
}

.community-search-clear,
.community-search-close {
  @apply inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-gray-600 transition-colors hover:bg-[#c8bbae] hover:text-[#8f382f] dark:text-gray-300 dark:hover:bg-[#202b39] dark:hover:text-[#f0b4a8];
}

.community-search-submit {
  @apply inline-flex h-10 shrink-0 items-center gap-2;
}

.community-search-summary {
  @apply flex flex-wrap items-center justify-between gap-3 border-b border-[#b4aa9f] bg-[#d2c6b8] px-4 py-2 text-xs font-medium text-gray-700 dark:border-[#252b35] dark:bg-[#18212c] dark:text-gray-200 md:px-5;
}

.community-search-summary-actions {
  @apply ml-auto flex items-center gap-2;
}

.community-search-position {
  @apply min-w-[3.5rem] text-center tabular-nums text-gray-600 dark:text-gray-300;
}

.community-search-navigation {
  @apply flex items-center overflow-hidden rounded-lg border border-[#a99b8c] bg-[#ddd2c5] dark:border-[#334052] dark:bg-[#202b39];
}

.community-search-nav-button {
  @apply inline-flex h-8 w-9 items-center justify-center text-gray-700 transition-colors hover:bg-[#c8bbae] hover:text-[#8f382f] focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[#8f382f]/30 disabled:cursor-not-allowed disabled:opacity-35 dark:text-gray-200 dark:hover:bg-[#2a3748] dark:hover:text-[#f0b4a8];
}

.community-search-nav-button + .community-search-nav-button {
  @apply border-l border-[#a99b8c] dark:border-[#334052];
}

.community-search-return {
  @apply shrink-0 font-semibold text-[#8a2525] hover:underline dark:text-[#f0b4a8];
}

.community-contact-entry {
  @apply flex flex-col items-center gap-2 self-center md:justify-self-center;
}

.community-contact-action {
  @apply relative;
}

.community-contact-button {
  @apply relative isolate inline-flex h-16 min-w-[13rem] items-center justify-center gap-3 overflow-hidden rounded-2xl border border-[#c7a35d]/70 bg-gradient-to-br from-[#fff9eb] via-[#f5e1b4] to-[#e8c580] px-5 text-lg font-bold text-[#60451e] shadow-[0_10px_24px_rgba(143,106,40,0.18),inset_0_1px_0_rgba(255,255,255,0.85)] transition-[transform,box-shadow] duration-200 hover:-translate-y-0.5 hover:shadow-[0_14px_30px_rgba(143,106,40,0.26),inset_0_1px_0_rgba(255,255,255,0.95)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#a17a35]/60 focus-visible:ring-offset-2 disabled:opacity-50 dark:border-[#eed6a3]/70 dark:from-[#f7ebd0] dark:via-[#e8ce97] dark:to-[#d4ad68] dark:text-[#503815] dark:shadow-[0_10px_30px_rgba(213,170,89,0.12),inset_0_1px_0_rgba(255,255,255,0.65)] dark:focus-visible:ring-[#e8cc93]/70 dark:focus-visible:ring-offset-[#121821];
}

.community-contact-button::before {
  content: '';
  position: absolute;
  inset: 0 auto 0 0;
  z-index: -1;
  width: 45%;
  pointer-events: none;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.65), transparent);
  opacity: 0;
  transform: translateX(-180%) skewX(-20deg);
  animation: community-contact-shimmer 7s ease-in-out infinite;
}

.community-contact-icon {
  @apply relative inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-white/60 text-[#98702e] shadow-[0_3px_10px_rgba(146,108,39,0.1),inset_0_1px_0_rgba(255,255,255,0.9)] ring-1 ring-inset ring-[#c9a66b]/35 dark:bg-white/40 dark:text-[#805919];
  animation: community-contact-float 4.5s ease-in-out infinite;
}

.community-contact-sparkle {
  @apply absolute -right-1 -top-1 inline-flex h-5 w-5 items-center justify-center rounded-full bg-[#fff9eb] text-[#b88935] shadow-sm ring-1 ring-[#dfc78f]/50;
}

.community-contact-hint {
  @apply text-xs font-medium leading-4 tracking-wide text-[#705b38] dark:text-[#cabb9a];
}

@keyframes community-contact-float {
  0%, 100% { transform: translateY(0) rotate(0); }
  50% { transform: translateY(-2px) rotate(-3deg); }
}

@keyframes community-contact-shimmer {
  0%, 65% { opacity: 0; transform: translateX(-180%) skewX(-20deg); }
  72% { opacity: 0.8; }
  88%, 100% { opacity: 0; transform: translateX(340%) skewX(-20deg); }
}

.community-direct-notification {
  @apply pointer-events-none absolute -right-2 -top-2 z-10;
}

.community-direct-unread-badge {
  @apply relative inline-flex h-6 items-center gap-1 whitespace-nowrap rounded-full border-2 border-[#fff7e6] bg-[#be3047] px-2 text-xs font-bold leading-none text-white shadow-[0_3px_10px_rgba(190,48,71,0.3)] dark:border-[#f3dfb5];
}

.community-direct-unread-badge::before {
  content: '';
  position: absolute;
  inset: -4px;
  border: 2px solid rgba(190, 48, 71, 0.45);
  border-radius: inherit;
  pointer-events: none;
  animation: community-direct-unread-pulse 2.8s ease-out infinite;
}

@keyframes community-direct-unread-pulse {
  0% { opacity: 0.65; transform: scale(1); }
  75%, 100% { opacity: 0; transform: scale(1.15, 1.35); }
}

@media (prefers-reduced-motion: reduce) {
  .community-contact-icon,
  .community-contact-button::before,
  .community-direct-unread-badge::before {
    animation: none;
  }

  .community-direct-unread-badge::before {
    opacity: 0;
  }

  .community-contact-button {
    transition: none;
    transform: none;
  }
}

.community-chat-messages {
  @apply h-full min-h-0 overflow-y-auto px-4 py-5 md:px-6;
  scrollbar-color: #8f8274 #c8bfb4;
  scrollbar-width: auto;
  background-color: #afa69b;
  background-image:
    linear-gradient(rgba(55, 43, 34, 0.075) 1px, transparent 1px),
    linear-gradient(90deg, rgba(55, 43, 34, 0.055) 1px, transparent 1px);
  background-size: 28px 28px;
  box-shadow: inset 0 18px 38px rgba(61, 48, 37, 0.09);
}

.community-chat-message-region {
  @apply relative min-h-0 flex-1;
}

.community-drop-target {
  outline: 2px solid rgba(143, 56, 47, 0.58);
  outline-offset: -4px;
}

.community-drop-overlay {
  @apply pointer-events-none absolute inset-3 z-30 flex flex-col items-center justify-center gap-2 rounded-2xl border-2 border-dashed border-[#8f382f]/70 bg-[#e4d8ca]/95 px-6 text-center text-sm text-[#6f2a25] shadow-xl backdrop-blur-sm dark:border-[#d56355]/70 dark:bg-[#111923]/95 dark:text-[#f0b4a8];
}

.community-drop-icon {
  @apply mb-1 flex h-14 w-14 items-center justify-center rounded-full bg-[#8f382f] text-white shadow-lg dark:bg-[#7f2d24];
}

.community-drop-overlay strong {
  @apply text-base;
}

.community-new-message-button {
  @apply absolute bottom-4 left-1/2 z-10 inline-flex -translate-x-1/2 items-center gap-2 rounded-full border border-[#8f382f]/35 bg-[#8f382f] px-4 py-2 text-sm font-semibold text-white shadow-lg transition-colors hover:bg-[#7f2d24] focus:outline-none focus:ring-2 focus:ring-[#8f382f]/35 focus:ring-offset-2 dark:border-[#b84535]/45 dark:bg-[#7f2d24] dark:hover:bg-[#8f382f] dark:focus:ring-[#f0b4a8]/30 dark:focus:ring-offset-[#0c1218];
}

.community-chat-messages::-webkit-scrollbar,
.community-direct-messages::-webkit-scrollbar {
  width: 14px;
}

.community-chat-messages::-webkit-scrollbar-track,
.community-direct-messages::-webkit-scrollbar-track {
  background: #c8bfb4;
  border-radius: 999px;
}

.community-chat-messages::-webkit-scrollbar-thumb,
.community-direct-messages::-webkit-scrollbar-thumb {
  background: #8f8274;
  border: 3px solid #c8bfb4;
  border-radius: 999px;
}

.community-chat-messages::-webkit-scrollbar-thumb:hover,
.community-direct-messages::-webkit-scrollbar-thumb:hover {
  background: #7a6d61;
}

:global(.dark) .community-chat-messages {
  scrollbar-color: #526174 #0f151d;
  background-color: #0c1218;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.028) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.02) 1px, transparent 1px);
}

:global(.dark) .community-chat-messages::-webkit-scrollbar-track,
:global(.dark) .community-direct-messages::-webkit-scrollbar-track {
  background: #0f151d;
}

:global(.dark) .community-chat-messages::-webkit-scrollbar-thumb,
:global(.dark) .community-direct-messages::-webkit-scrollbar-thumb {
  background: #526174;
  border-color: #0f151d;
}

:global(.dark) .community-chat-messages::-webkit-scrollbar-thumb:hover,
:global(.dark) .community-direct-messages::-webkit-scrollbar-thumb:hover {
  background: #6b7b90;
}

.community-chat-empty {
  @apply flex h-full min-h-[300px] flex-col items-center justify-center gap-3 text-sm text-gray-500 dark:text-gray-400;
}

.community-message {
  @apply flex items-end gap-3;
}

.community-message-search-current .community-message-bubble {
  @apply outline outline-2 outline-offset-4 outline-[#b84535] transition-[outline-color,box-shadow] dark:outline-[#f0b4a8];
  box-shadow: 0 0 0 7px rgba(184, 69, 53, 0.12);
}

.community-search-keyword {
  border-radius: 0.2rem;
  background: #f6d365;
  box-decoration-break: clone;
  color: #30230f;
  padding: 0 0.14rem;
  -webkit-box-decoration-break: clone;
}

.community-search-keyword-current {
  background: #ffb24a;
  box-shadow: 0 0 0 2px rgba(122, 31, 31, 0.34);
  color: #261506;
}

:global(.dark) .community-search-keyword {
  background: #f2cd5d;
  color: #21170a;
}

:global(.dark) .community-search-keyword-current {
  background: #ffad42;
  box-shadow: 0 0 0 2px rgba(255, 235, 195, 0.5);
}

.community-message-own {
  @apply flex-row-reverse;
}

.community-avatar {
  @apply flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full border border-[#9f9284] bg-[#c8b9a8] text-sm font-semibold text-[#5f4428] shadow-sm dark:border-[#334052] dark:bg-[#273244] dark:text-gray-100;
}

.community-message-content {
  @apply flex min-w-0 flex-col items-start;
  max-width: min(760px, calc(100% - 3.5rem));
}

.community-message-own .community-message-content {
  @apply items-end;
}

.community-message-meta {
  @apply mb-1.5 flex min-w-0 items-center gap-2 text-xs text-gray-500 dark:text-gray-400;
}

.community-message-own .community-message-meta {
  @apply flex-row-reverse;
}

.community-message-time {
  @apply shrink-0 rounded-full border border-[#8f8274] bg-[#e2d7c9]/95 px-2 py-0.5 text-[11px] font-semibold text-[#493d32] shadow-sm dark:border-[#3b4656] dark:bg-[#202b39] dark:text-gray-100;
}

.community-message-own .community-message-time {
  @apply border-[#b9564b]/45 bg-[#7f2d24]/90 text-white dark:border-[#b9564b]/45 dark:bg-[#7f2d24]/90 dark:text-white;
}

.community-message-bubble {
  @apply w-fit max-w-full rounded-2xl border border-[#b7aa9b] bg-[#d8cdc0] px-4 py-3 text-sm text-gray-900 shadow-[0_12px_28px_rgba(67,43,25,0.13)] dark:border-[#2d3746] dark:bg-[#18212c] dark:text-gray-100 dark:shadow-[0_16px_36px_rgba(0,0,0,0.24)];
}

.community-message-own .community-message-bubble {
  @apply border-[#8f382f]/35 bg-[#8f382f] text-white shadow-[0_12px_28px_rgba(143,56,47,0.22)] dark:border-[#b84535]/35 dark:bg-[#7f2d24] dark:text-white dark:shadow-[0_16px_36px_rgba(127,45,36,0.22)];
}

.community-message-hover-actions {
  @apply mt-1.5 flex w-full justify-start opacity-0 transition-opacity duration-150;
  pointer-events: none;
}

.community-message:hover .community-message-hover-actions,
.community-message:focus-within .community-message-hover-actions {
  opacity: 1;
  pointer-events: auto;
}

.community-message-own .community-message-hover-actions {
  @apply justify-end;
}

@media (hover: none) {
  .community-message-hover-actions {
    opacity: 1;
    pointer-events: auto;
  }
}

.community-message-image {
  @apply mb-3 max-h-80 max-w-full rounded-xl border border-[#cbbca9] object-contain dark:border-dark-600;
}

.community-message-video {
  @apply mb-3 max-h-96 max-w-full rounded-xl border border-[#cbbca9] bg-black object-contain dark:border-dark-600;
}

.community-image-button {
  @apply block max-w-full cursor-zoom-in rounded-xl text-left;
}

.community-attachment-download {
  @apply mb-2 inline-flex items-center gap-1.5 rounded-lg border border-[#b9ab9c] bg-[#cfc3b6] px-2 py-1 text-xs font-semibold text-gray-700 transition-colors hover:border-[#9f8f7f] hover:bg-[#d8cdc0] hover:text-[#8a2525] dark:border-[#2d3746] dark:bg-[#202b39] dark:text-gray-200 dark:hover:bg-dark-700 dark:hover:text-[#f0b4a8];
}

.community-search-attachment-name {
  @apply mb-2 max-w-full break-all rounded-lg border border-[#b88735]/45 bg-[#e2c98f]/50 px-2 py-1 text-xs font-semibold text-gray-800 dark:border-[#e0ad53]/35 dark:bg-[#6b4c1d]/35 dark:text-gray-100;
}

.community-file-card {
  @apply mb-1.5 flex min-w-[240px] max-w-full items-center gap-3 rounded-xl border border-[#c1b29f] bg-[#cfc3b6] p-3 text-left text-gray-800 transition-colors hover:border-[#9f8f7f] hover:bg-[#d8cdc0] dark:border-[#2d3746] dark:bg-[#202b39] dark:text-gray-100 dark:hover:bg-dark-700;
}

.community-file-icon,
.community-selected-file-icon {
  @apply flex h-12 w-12 shrink-0 items-center justify-center rounded-xl border border-[#b9ab9c] bg-[#d8cdc0] text-[#8a2525] dark:border-[#2d3746] dark:bg-[#18212c] dark:text-[#f0b4a8];
}

.community-image-lightbox {
  @apply flex flex-col items-center gap-4;
}

.community-image-lightbox-img {
  @apply max-h-[72vh] max-w-full rounded-2xl border border-[#cbbca9] object-contain shadow-lg dark:border-dark-600;
}

.community-video-lightbox {
  @apply max-h-[72vh] w-full max-w-5xl rounded-2xl border border-[#cbbca9] bg-black shadow-lg dark:border-dark-600;
}

.community-delete-button {
  @apply inline-flex h-6 w-6 items-center justify-center rounded-lg border border-transparent bg-[#c9beb1] text-gray-600 transition-all hover:border-red-200 hover:bg-red-50 hover:text-red-500 disabled:opacity-50 dark:bg-[#202b39] dark:text-gray-300 dark:hover:bg-red-900/20;
}

.community-message-action {
  @apply rounded-lg border border-[#b9ab9c] bg-[#cfc3b6] px-1.5 py-0.5 text-[11px] font-medium text-gray-700 transition-all hover:border-[#9f8f7f] hover:bg-[#d8cdc0] hover:text-[#8a2525] dark:border-[#2d3746] dark:bg-[#202b39] dark:text-gray-300 dark:hover:bg-dark-700 dark:hover:text-[#f0b4a8];
}

.community-reply-preview {
  @apply mb-2 flex max-w-full flex-col rounded-xl border-l-2 border-[#9f6833] bg-[#cbbfb1] px-3 py-2 text-left text-xs text-gray-700 dark:border-[#f0b4a8] dark:bg-black/20 dark:text-gray-200;
}

.community-reply-target {
  @apply mb-3 flex items-center justify-between gap-3 rounded-2xl border border-[#a9937c] bg-[#cbbdad] p-3 dark:border-[#5f4632] dark:bg-[#2b231a];
}

.community-chat-composer {
  @apply border-t border-[#b4aa9f] bg-[#c5bbb0] px-4 py-3 backdrop-blur dark:border-[#252b35] dark:bg-[#121821]/95 md:px-5;
}

.community-image-preview {
  @apply mb-3 flex items-center justify-between gap-3 rounded-2xl border border-[#a99b8c] bg-[#d1c5b7] p-3 dark:border-dark-600 dark:bg-dark-800/80;
}

.community-input-row {
  @apply flex items-end gap-2 rounded-2xl border border-[#a89b8d] bg-[#bdb3a8] p-2 shadow-inner dark:border-[#2d3746] dark:bg-[#0f151d];
}

.community-input-row-dragging {
  @apply border-[#8f382f] ring-2 ring-[#8f382f]/20 dark:border-[#b84535];
}

.community-icon-button {
  @apply inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-gray-800 transition-colors hover:bg-[#d1c5b7] disabled:cursor-not-allowed disabled:opacity-50 dark:text-gray-300 dark:hover:bg-[#202b39];
}

.community-textarea-shell {
  @apply relative min-w-0 flex-1;
}

.community-textarea {
  @apply min-h-[42px] flex-1 resize-none rounded-xl border-0 bg-[#d8cdc0] px-3 py-2.5 text-sm text-gray-900 outline-none placeholder:text-gray-600 focus:ring-2 focus:ring-[#8f382f]/25 dark:bg-[#18212c] dark:text-white dark:focus:ring-[#f0b4a8]/15;
  max-height: 144px;
  overflow-y: hidden;
}

.community-textarea-shell .community-textarea {
  @apply block w-full;
}

.community-textarea-with-emoji {
  padding-right: 3.25rem;
}

.community-emoji-picker-trigger {
  @apply absolute bottom-px right-1;
}

.community-send-button {
  @apply h-10 shrink-0;
}

.community-direct-chat {
  @apply grid min-h-[520px] gap-4 md:grid-cols-[280px_minmax(0,1fr)];
  height: min(70vh, 640px);
}

.community-direct-chat.community-direct-chat-single {
  @apply grid-cols-1 md:grid-cols-1;
}

.community-direct-conversations {
  @apply flex min-h-0 flex-col overflow-hidden rounded-2xl border border-[#b7aa9b] bg-[#cbbfb1] dark:border-[#2d3746] dark:bg-[#111923];
}

.community-direct-conversations-title {
  @apply border-b border-[#b7aa9b] px-4 py-3 text-sm font-semibold text-gray-900 dark:border-[#2d3746] dark:text-white;
}

.community-direct-user-search {
  @apply m-2 flex h-10 items-center gap-2 rounded-xl border border-[#a99b8c] bg-[#d8cdc0] px-3 dark:border-[#2d3746] dark:bg-[#18212c];
}

.community-direct-conversation-list {
  @apply min-h-0 flex-1 overflow-y-auto p-2;
}

.community-direct-scroll-list {
  max-height: 100%;
  overscroll-behavior: contain;
}

.community-direct-load-more {
  @apply mt-2 flex w-full items-center justify-center gap-2 rounded-xl px-3 py-2 text-xs font-semibold text-[#7a1f1f] transition-colors hover:bg-[#d8cdc0] disabled:opacity-50 dark:text-[#f0b4a8] dark:hover:bg-[#202b39];
}

.community-direct-conversation {
  @apply flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-gray-800 transition-all hover:bg-[#d8cdc0] dark:text-gray-100 dark:hover:bg-[#202b39];
}

.community-direct-conversation-active {
  @apply bg-[#8f382f] text-white shadow-sm hover:bg-[#8f382f] dark:bg-[#7f2d24] dark:hover:bg-[#7f2d24];
}

.community-direct-count-badge {
  @apply inline-flex min-w-5 items-center justify-center rounded-full bg-red-500 px-1.5 py-0.5 text-[10px] font-bold leading-none text-white shadow-sm;
}

.community-direct-thread {
  @apply relative flex min-h-0 flex-col overflow-hidden rounded-2xl border border-[#b7aa9b] bg-[#d8cdc0] dark:border-[#2d3746] dark:bg-[#0f151d];
}

.community-direct-thread-header {
  @apply flex items-center gap-3 border-b border-[#b7aa9b] bg-[#ded3c6] px-4 py-3 dark:border-[#2d3746] dark:bg-[#121821];
}

.community-direct-messages {
  @apply min-h-0 flex-1 overflow-y-auto px-4 py-5;
  scrollbar-color: #8f8274 #c8bfb4;
  scrollbar-width: auto;
  background-color: #b8afa4;
  background-image:
    linear-gradient(rgba(55, 43, 34, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(55, 43, 34, 0.045) 1px, transparent 1px);
  background-size: 26px 26px;
}

:global(.dark) .community-direct-messages {
  scrollbar-color: #526174 #0f151d;
  background-color: #0c1218;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.018) 1px, transparent 1px);
}

.community-direct-message .community-message-content {
  max-width: min(560px, calc(100% - 3.5rem));
}

.community-direct-composer {
  @apply flex flex-col gap-2 border-t border-[#b7aa9b] bg-[#c5bbb0] p-3 dark:border-[#2d3746] dark:bg-[#121821];
}

.community-direct-composer .community-input-row {
  @apply w-full;
}

.community-direct-new-message-button {
  @apply absolute bottom-20 left-1/2 z-10 inline-flex -translate-x-1/2 items-center gap-2 rounded-full border border-[#8f382f]/35 bg-[#8f382f] px-4 py-2 text-sm font-semibold text-white shadow-lg transition-colors hover:bg-[#7f2d24] dark:border-[#b84535]/45 dark:bg-[#7f2d24] dark:hover:bg-[#8f382f];
}

.community-direct-empty {
  @apply flex min-h-[180px] flex-col items-center justify-center gap-3 px-4 text-center text-sm text-gray-500 dark:text-gray-400;
}

.community-direct-empty-thread {
  @apply h-full min-h-[520px];
}

@media (max-width: 767px) {
  .community-search-bar {
    @apply flex-wrap;
  }

  .community-search-input-shell {
    flex-basis: calc(100% - 2.5rem);
  }

  .community-search-submit {
    @apply flex-1 justify-center;
  }

  .community-direct-chat {
    @apply min-h-[70vh] grid-cols-1;
    height: auto;
  }

  .community-direct-conversations {
    @apply max-h-56;
  }
}
</style>
