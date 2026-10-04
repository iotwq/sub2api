<template>
  <AppLayout>
    <div class="image-bridge-page">
      <section class="image-bridge-shell">
        <div class="image-bridge-main">
          <div class="image-bridge-heading">
            <div class="image-bridge-icon">
              <Icon name="sparkles" size="xl" />
            </div>

            <div class="min-w-0">
              <p class="image-bridge-kicker">{{ t('imageBridge.kicker') }}</p>
              <h1 class="image-bridge-title">{{ t('imageBridge.title') }}</h1>
              <p class="image-bridge-description">{{ t('imageBridge.description') }}</p>
            </div>
          </div>

          <div class="image-bridge-actions">
            <a
              :href="targetUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-primary"
            >
              <Icon name="externalLink" size="sm" class="mr-1.5" />
              {{ t('imageBridge.openAction') }}
            </a>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { detectTheme } from '@/utils/embedded-url'

const DEFAULT_IMAGE_WORKSPACE_URL = 'https://image.iotwq.top/'

const { t, locale } = useI18n()
const appStore = useAppStore()

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
})

const imageWorkspaceURL = computed(() => {
  return appStore.cachedPublicSettings?.image_workspace_url?.trim() || DEFAULT_IMAGE_WORKSPACE_URL
})

function buildTargetUrl(baseUrl: string): string {
  let url: URL
  try {
    url = new URL(baseUrl)
    if (!['http:', 'https:'].includes(url.protocol)) {
      throw new Error('Unsupported image workspace URL protocol')
    }
  } catch {
    url = new URL(DEFAULT_IMAGE_WORKSPACE_URL)
  }

  url.searchParams.set('theme', detectTheme())
  url.searchParams.set('lang', locale.value)
  url.searchParams.set('ui_mode', 'external')
  if (typeof window !== 'undefined') {
    url.searchParams.set('src_host', window.location.origin)
    url.searchParams.set('src_url', window.location.href)
  }
  return url.toString()
}

const targetUrl = computed(() => {
  return buildTargetUrl(imageWorkspaceURL.value)
})
</script>

<style scoped>
.image-bridge-page {
  @apply min-h-[calc(100vh-8rem)] rounded-[2rem] border border-[#e8e0d3] bg-[#fbf8f2] p-4 shadow-[0_28px_70px_rgba(15,23,42,0.08)] dark:border-[#2a3039] dark:bg-[#10141a] md:p-6;
}

.image-bridge-shell {
  @apply flex min-h-[240px] rounded-[1.75rem] border border-white/70 bg-white/90 p-5 shadow-[0_18px_45px_rgba(15,23,42,0.06)] dark:border-[#242933] dark:bg-[#141920]/90 md:p-7;
}

.image-bridge-main {
  @apply flex w-full flex-col gap-6 md:flex-row md:items-center md:justify-between;
}

.image-bridge-heading {
  @apply flex min-w-0 flex-col gap-5 sm:flex-row sm:items-center;
}

.image-bridge-icon {
  @apply flex h-16 w-16 shrink-0 items-center justify-center rounded-[1.35rem] bg-[#f3eadc] text-[#7c5b2d] dark:bg-[#242a33] dark:text-[#d9c292];
}

.image-bridge-kicker {
  @apply text-xs font-semibold uppercase tracking-[0.22em] text-gray-500 dark:text-gray-400;
}

.image-bridge-title {
  @apply mt-2 text-2xl font-semibold text-gray-900 dark:text-white md:text-3xl;
}

.image-bridge-description {
  @apply mt-3 max-w-2xl text-sm leading-7 text-gray-600 dark:text-gray-300;
}

.image-bridge-actions {
  @apply flex shrink-0;
}
</style>
