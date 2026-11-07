<template>
  <div>
    <div
      v-if="loading && items.length === 0"
      class="grid gap-5 grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4"
    >
      <div
        v-for="i in 6"
        :key="i"
        class="p-5 rounded-2xl min-h-[280px] bg-white/70 dark:bg-dark-800/60 border border-gray-200/80 dark:border-dark-700/70 animate-pulse"
      >
        <div class="flex items-start gap-3">
          <div class="w-9 h-9 rounded-xl bg-gray-200 dark:bg-dark-700"></div>
          <div class="flex-1 space-y-2">
            <div class="h-4 w-2/3 rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="h-3 w-1/2 rounded bg-gray-200 dark:bg-dark-700"></div>
          </div>
          <div class="h-6 w-16 rounded-full bg-gray-200 dark:bg-dark-700"></div>
        </div>
        <div class="mt-5 grid grid-cols-2 gap-2">
          <div class="h-16 rounded-xl bg-gray-100 dark:bg-dark-900/40"></div>
          <div class="h-16 rounded-xl bg-gray-100 dark:bg-dark-900/40"></div>
        </div>
        <div class="mt-6 h-5 w-full rounded bg-gray-100 dark:bg-dark-900/40"></div>
      </div>
    </div>

    <EmptyState
      v-else-if="items.length === 0"
      :title="t('channelStatus.empty.title')"
      :description="t('channelStatus.empty.description')"
    />

    <div v-else class="space-y-8">
      <section
        v-for="group in providerGroups"
        :key="group.provider"
        :aria-labelledby="`monitor-provider-${group.provider}`"
      >
        <h2
          :id="`monitor-provider-${group.provider}`"
          class="mb-4 flex items-center gap-3 text-xl font-bold text-gray-900 dark:text-gray-100"
        >
          <span
            class="grid h-10 w-10 place-items-center rounded-xl ring-1 ring-black/5 dark:ring-white/10"
            :class="providerBadgeClass(group.provider)"
          >
            <ProviderIcon :provider="group.provider" :size="24" />
          </span>
          <span>{{ t(`admin.groups.platforms.${group.provider}`) }}</span>
          <span class="inline-flex h-7 min-w-7 items-center justify-center rounded-full bg-gray-100 px-2 text-sm font-semibold text-gray-600 ring-1 ring-gray-200 dark:bg-dark-700 dark:text-gray-300 dark:ring-dark-600">
            {{ group.items.length }}
          </span>
        </h2>
        <div class="grid gap-5 grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
          <MonitorCard
            v-for="item in group.items"
            :key="item.id"
            :item="item"
            :window="window"
            :availability-value="resolveAvailability(item)"
            :countdown-seconds="countdownSeconds"
            @click="emit('cardClick', item)"
          />
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Provider, UserMonitorView } from '@/api/channelMonitor'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'
import EmptyState from '@/components/common/EmptyState.vue'
import MonitorCard from './MonitorCard.vue'
import ProviderIcon from './ProviderIcon.vue'

const props = defineProps<{
  items: UserMonitorView[]
  window: '7d' | '15d' | '30d'
  countdownSeconds: number
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'cardClick', item: UserMonitorView): void
}>()

const { t } = useI18n()
const { providerBadgeClass } = useChannelMonitorFormat()

const providerOrder: Provider[] = ['openai', 'anthropic', 'grok', 'gemini', 'deepseek', 'zhipu', 'kimi']
const providerGroups = computed(() => {
  const groups = new Map<Provider, UserMonitorView[]>(providerOrder.map(provider => [provider, []]))
  for (const item of props.items) {
    const group = groups.get(item.provider)
    if (group) group.push(item)
    else groups.set(item.provider, [item])
  }
  return Array.from(groups, ([provider, items]) => ({ provider, items }))
    .filter(group => group.items.length > 0)
})

function resolveAvailability(item: UserMonitorView): number | null {
  return item[`availability_${props.window}`] ?? null
}
</script>
