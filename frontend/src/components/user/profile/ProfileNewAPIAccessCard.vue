<template>
  <section
    data-testid="profile-newapi-access-card"
    class="card border border-gray-100 bg-white/90 p-6 dark:border-dark-700 dark:bg-dark-900/50"
  >
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('profile.newapi.title') }}
        </h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {{ t('profile.newapi.description') }}
        </p>
      </div>
      <button
        type="button"
        class="btn btn-primary"
        data-testid="profile-newapi-generate"
        :disabled="generating"
        @click="confirmVisible = true"
      >
        <Icon name="refresh" size="sm" :class="generating ? 'animate-spin' : ''" />
        {{ generating ? t('profile.newapi.generating') : t('profile.newapi.generate') }}
      </button>
    </div>

    <div class="mt-5 grid gap-4 md:grid-cols-2">
      <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/60">
        <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
          {{ t('profile.newapi.userId') }}
        </p>
        <div class="mt-2 flex items-center gap-2">
          <code data-testid="profile-newapi-user-id" class="min-w-0 flex-1 truncate text-sm text-gray-900 dark:text-white">
            {{ userId }}
          </code>
          <button
            type="button"
            class="btn btn-ghost btn-icon h-8 w-8"
            :title="t('profile.newapi.copyUserId')"
            @click="copyToClipboard(String(userId))"
          >
            <Icon name="copy" size="sm" />
          </button>
        </div>
      </div>

      <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/60">
        <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
          {{ t('profile.newapi.accessToken') }}
        </p>
        <div v-if="token" class="mt-2 flex items-center gap-2">
          <input
            data-testid="profile-newapi-token"
            class="input min-w-0 flex-1 font-mono text-xs"
            type="text"
            readonly
            :value="token"
          >
          <button
            type="button"
            class="btn btn-secondary btn-icon h-9 w-9"
            :title="t('profile.newapi.copyToken')"
            @click="copyToClipboard(token)"
          >
            <Icon name="copy" size="sm" />
          </button>
        </div>
        <p v-else class="mt-2 text-sm text-gray-500 dark:text-gray-400">
          {{ t('profile.newapi.tokenUnavailable') }}
        </p>
        <p class="mt-2 text-xs text-amber-600 dark:text-amber-400">
          {{ t('profile.newapi.showOnce') }}
        </p>
      </div>
    </div>

    <ConfirmDialog
      :show="confirmVisible"
      :title="t('profile.newapi.confirmTitle')"
      :message="t('profile.newapi.confirmMessage')"
      :confirm-text="t('profile.newapi.confirm')"
      danger
      @confirm="generateToken"
      @cancel="confirmVisible = false"
    />
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { generateNewAPIBalanceAccessToken } from '@/api/user'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'

defineProps<{ userId: number }>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const appStore = useAppStore()
const confirmVisible = ref(false)
const generating = ref(false)
const token = ref('')

async function generateToken() {
  if (generating.value) return
  generating.value = true
  try {
    token.value = await generateNewAPIBalanceAccessToken()
    confirmVisible.value = false
    appStore.showSuccess(t('profile.newapi.generateSuccess'))
  } catch {
    appStore.showError(t('profile.newapi.generateFailed'))
  } finally {
    generating.value = false
  }
}
</script>
