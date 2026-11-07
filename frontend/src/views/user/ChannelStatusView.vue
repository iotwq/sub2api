<template>
  <ChannelStatusV1View v-if="isV1" />
  <ChannelStatusV2View v-else />
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { isChannelMonitorV1Mode } from '@/utils/featureFlags'
const ChannelStatusV1View = defineAsyncComponent(() => import('./ChannelStatusV1View.vue'))
const ChannelStatusV2View = defineAsyncComponent(() => import('./ChannelStatusV2View.vue'))

const isV1 = computed(() => isChannelMonitorV1Mode())
</script>
