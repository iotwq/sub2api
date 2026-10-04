<template>
  <div class="space-y-4 border-t border-gray-200 pt-4 dark:border-dark-600" data-testid="codex-protection">
    <div class="flex items-center justify-between gap-4">
      <label class="input-label mb-0" for="codex-protection-enabled">{{ t('admin.accounts.openai.protection.title') }}</label>
      <Toggle id="codex-protection-enabled" :model-value="modelValue.enabled" @update:model-value="set('enabled', $event)" />
    </div>
    <div v-if="modelValue.enabled" class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <label class="input-label">
        {{ t('admin.accounts.openai.protection.concurrency') }}
        <input :value="modelValue.max_concurrency" type="number" min="1" max="10000" step="1" required class="input mt-1" data-testid="protection-concurrency" @input="number('max_concurrency', $event)" />
      </label>
      <label class="input-label">
        {{ t('admin.accounts.openai.protection.wait') }}
        <input :value="modelValue.wait_seconds" type="number" min="0" max="120" step="1" required class="input mt-1" @input="number('wait_seconds', $event)" />
      </label>
      <label class="flex items-center gap-2 text-sm sm:col-span-2">
        <input type="checkbox" :checked="modelValue.strict_rpm_enabled" @change="check('strict_rpm_enabled', $event)" />
        {{ t('admin.accounts.openai.protection.strictRPM') }}
      </label>
      <template v-if="modelValue.strict_rpm_enabled">
        <label class="input-label">
          RPM
          <input :value="modelValue.rpm" type="number" min="1" max="60000" step="1" required class="input mt-1" @input="number('rpm', $event)" />
        </label>
        <label class="input-label">
          {{ t('admin.accounts.openai.protection.burst') }}
          <input :value="modelValue.burst" type="number" min="1" :max="modelValue.rpm" step="1" required class="input mt-1" @input="number('burst', $event)" />
        </label>
      </template>
      <label class="flex items-center gap-2 text-sm sm:col-span-2">
        <input type="checkbox" :checked="modelValue.adaptive_enabled" @change="check('adaptive_enabled', $event)" />
        {{ t('admin.accounts.openai.protection.adaptive') }}
      </label>
      <label v-if="modelValue.adaptive_enabled" class="input-label sm:col-span-2">
        {{ t('admin.accounts.openai.protection.adaptiveMode') }}
        <select :value="modelValue.adaptive_mode" class="input mt-1" @change="select('adaptive_mode', $event)">
          <option value="automatic">{{ t('admin.accounts.openai.protection.automatic') }}</option>
          <option value="observe">{{ t('admin.accounts.openai.protection.observe') }}</option>
        </select>
      </label>
      <label class="input-label">
        {{ t('admin.accounts.openai.protection.tls') }}
        <select :value="modelValue.tls_profile" class="input mt-1" @change="select('tls_profile', $event)">
          <option value="standard">{{ t('admin.accounts.openai.protection.standard') }}</option>
          <option value="nodejs24">Node.js 24 (HTTP/1.1)</option>
          <option value="nodejs22">Node.js 22 (HTTP/1.1)</option>
        </select>
      </label>
      <label class="input-label">
        {{ t('admin.accounts.openai.protection.integrity') }}
        <select :value="modelValue.integrity_mode" class="input mt-1" @change="select('integrity_mode', $event)">
          <option value="off">{{ t('admin.accounts.openai.protection.off') }}</option>
          <option value="observe">{{ t('admin.accounts.openai.protection.observe') }}</option>
          <option value="enforce">{{ t('admin.accounts.openai.protection.enforce') }}</option>
        </select>
      </label>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { OpenAIOAuthProtection } from './openAIOAuthProtection'

const props = defineProps<{ modelValue: OpenAIOAuthProtection }>()
const emit = defineEmits<{ 'update:modelValue': [value: OpenAIOAuthProtection] }>()
const { t } = useI18n()
function set<K extends keyof OpenAIOAuthProtection>(key: K, value: OpenAIOAuthProtection[K]) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}
function number(key: 'max_concurrency' | 'wait_seconds' | 'rpm' | 'burst', event: Event) {
  set(key, Number((event.target as HTMLInputElement).value))
}
function check(key: 'strict_rpm_enabled' | 'adaptive_enabled', event: Event) {
  set(key, (event.target as HTMLInputElement).checked)
}
function select<K extends 'adaptive_mode' | 'tls_profile' | 'integrity_mode'>(key: K, event: Event) {
  set(key, (event.target as HTMLSelectElement).value as OpenAIOAuthProtection[K])
}
</script>
