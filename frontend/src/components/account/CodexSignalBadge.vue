<template>
  <div v-if="eligible" class="text-xs leading-5" data-testid="codex-signal" :title="tooltip">
    <span class="inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 font-medium" :class="tone">
      <span class="h-1.5 w-1.5 rounded-full bg-current" aria-hidden="true" />
      {{ label }}
    </span>
    <div v-if="signal" class="mt-0.5 text-[10px] text-gray-500 dark:text-gray-400">
      {{ t('admin.accounts.openai.codexSignalObserved', { time: displayTime(signal.observed_at) }) }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Account } from '@/types'

const props = defineProps<{ account: Account }>()
const { t } = useI18n()
const eligible = computed(() => props.account.platform === 'openai'
  && ['oauth', 'setup-token'].includes(props.account.type) && props.account.parent_account_id == null)
const signal = computed(() => props.account.codex_signal)
const label = computed(() => t(`admin.accounts.openai.${!signal.value ? 'codexSignalUnknown' : signal.value.length === 312 ? 'codexSignalDetected' : 'codexSignalNotDetected'}`))
const tone = computed(() => signal.value?.length === 312
  ? 'bg-red-50 text-red-700 dark:bg-red-900/25 dark:text-red-300'
  : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300')
const displayTime = (value: string) => new Date(value).toLocaleString()
const tooltip = computed(() => {
  const lines = [t('admin.accounts.openai.codexSignalHelp')]
  if (signal.value) {
    lines.push(t('admin.accounts.openai.codexSignalLength', { length: signal.value.length }))
    if (signal.value.last_312_at) lines.push(t('admin.accounts.openai.codexSignalLast312', { time: displayTime(signal.value.last_312_at) }))
  }
  return lines.join('\n')
})
</script>
