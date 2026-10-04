<template>
  <div class="mt-3 w-full border-t border-gray-100 pt-3 dark:border-dark-700/60" data-testid="intelligence-timeline">
    <div class="mb-2 flex items-center justify-between gap-2 text-xs">
      <span class="text-gray-500 dark:text-gray-400" :title="t('monitorCommon.intelligence.scope')">
        {{ t('monitorCommon.intelligence.title') }}
      </span>
      <MonitorIntelligenceBadge :result="buckets[0]?.intelligence" :checked-at="buckets[0]?.checked_at" />
    </div>
    <div class="flex h-4 w-full gap-[2px]" :aria-label="t('monitorCommon.intelligence.history')">
      <span
        v-for="(point, index) in bars"
        :key="index"
        class="min-w-0 flex-1 rounded-sm transition-opacity hover:opacity-70"
        :class="color(point?.intelligence)"
        :title="tooltip(point?.intelligence, point?.checked_at)"
        :aria-label="tooltip(point?.intelligence, point?.checked_at)"
        :data-status="point?.intelligence?.status || 'untested'"
      />
    </div>
    <div class="mt-1 flex justify-between text-[9px] text-gray-400">
      <span>{{ t('monitorCommon.past') }}</span>
      <span>{{ t('monitorCommon.now') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MonitorTimelinePoint } from '@/api/channelMonitor'
import MonitorIntelligenceBadge from '@/components/common/MonitorIntelligenceBadge.vue'
import { useMonitorIntelligence } from '@/composables/useMonitorIntelligence'

const props = withDefaults(defineProps<{ buckets?: MonitorTimelinePoint[] }>(), { buckets: () => [] })
const { t } = useI18n()
const { color, tooltip } = useMonitorIntelligence()
const bars = computed<(MonitorTimelinePoint | null)[]>(() => {
  const points = props.buckets.slice(0, 60).reverse()
  return [...Array<null>(60 - points.length).fill(null), ...points]
})
</script>
