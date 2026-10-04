<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-medium"
    :class="result?.status === 'passed'
      ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400'
      : result?.status === 'failed'
        ? 'bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-400'
        : 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'"
    :title="tooltip(result, checkedAt)"
  >
    <span class="h-1.5 w-1.5 rounded-full" :class="color(result)" aria-hidden="true" />
    {{ label(result) }}
  </span>
</template>

<script setup lang="ts">
import type { MonitorIntelligenceResult } from '@/api/admin/channelMonitor'
import { useMonitorIntelligence } from '@/composables/useMonitorIntelligence'

defineProps<{ result?: MonitorIntelligenceResult | null; checkedAt?: string }>()
const { label, color, tooltip } = useMonitorIntelligence()
</script>
