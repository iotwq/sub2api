<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.dataImportTitle')"
    width="normal"
    close-on-click-outside
    @close="handleClose"
  >
    <form id="import-data-form" class="space-y-4" @submit.prevent="handleImport">
      <div class="text-sm text-gray-600 dark:text-dark-300">
        {{ t('admin.accounts.dataImportHint') }}
      </div>
      <div
        class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-xs text-amber-600 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-400"
      >
        {{ t('admin.accounts.dataImportWarning') }}
      </div>

      <div>
        <label class="input-label">{{ t('admin.accounts.dataImportFile') }}</label>
        <div
          class="flex items-center justify-between gap-3 rounded-lg border border-dashed border-gray-300 bg-gray-50 px-4 py-3 dark:border-dark-600 dark:bg-dark-800"
        >
          <div class="min-w-0">
            <div class="truncate text-sm text-gray-700 dark:text-dark-200">
              {{ fileName || t('admin.accounts.dataImportSelectFile') }}
            </div>
            <div class="text-xs text-gray-500 dark:text-dark-400">JSON (.json)</div>
          </div>
          <button type="button" class="btn btn-secondary shrink-0" @click="openFilePicker">
            {{ t('common.chooseFile') }}
          </button>
        </div>
        <input
          ref="fileInput"
          type="file"
          class="hidden"
          accept="application/json,.json,application/zip,.zip"
          @change="handleFileChange"
        />
      </div>

      <div
        v-if="result"
        class="space-y-2 rounded-xl border border-gray-200 p-4 dark:border-dark-700"
      >
        <div class="text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.accounts.dataImportResult') }}
        </div>
        <div class="text-sm text-gray-700 dark:text-dark-300">
          {{ resultSummary }}
        </div>

        <div v-if="errorLines.length" class="mt-2">
          <div class="text-sm font-medium text-red-600 dark:text-red-400">
            {{ t('admin.accounts.dataImportErrors') }}
          </div>
          <div
            class="mt-2 max-h-48 overflow-auto rounded-lg bg-gray-50 p-3 font-mono text-xs dark:bg-dark-800"
          >
            <div v-for="(line, idx) in errorLines" :key="idx" class="whitespace-pre-wrap">
              {{ line }}
            </div>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" type="button" :disabled="importing" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-primary"
          type="submit"
          form="import-data-form"
          :disabled="importing"
        >
          {{ importing ? t('admin.accounts.dataImporting') : t('admin.accounts.dataImportButton') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import JSZip from 'jszip'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { AdminDataImportResult, AdminDataPayload, CodexSessionImportResult } from '@/types'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'imported'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const importing = ref(false)
const file = ref<File | null>(null)
type ImportResultState =
  | { kind: 'data'; value: AdminDataImportResult }
  | { kind: 'codex'; value: CodexSessionImportResult }

type ImportRequestState =
  | { kind: 'data'; payload: AdminDataPayload }
  | { kind: 'codex'; contents: string[] }

const result = ref<ImportResultState | null>(null)

const fileInput = ref<HTMLInputElement | null>(null)
const fileName = computed(() => file.value?.name || '')

const errorLines = computed(() => {
  if (!result.value) return []
  if (result.value.kind === 'data') {
    return (result.value.value.errors || []).map((item) =>
      `${item.kind} ${item.name || item.proxy_key || '-'} — ${item.message}`
    )
  }

  return [...(result.value.value.errors || []), ...(result.value.value.warnings || [])].map((item) =>
    `#${item.index} ${item.name || '-'} — ${item.message}`
  )
})

const resultSummary = computed(() => {
  if (!result.value) return ''
  if (result.value.kind === 'data') {
    return t('admin.accounts.dataImportResultSummary', result.value.value)
  }
  return t('admin.accounts.dataImportCodexResultSummary', result.value.value)
})

watch(
  () => props.show,
  (open) => {
    if (open) {
      file.value = null
      result.value = null
      if (fileInput.value) {
        fileInput.value.value = ''
      }
    }
  }
)

const openFilePicker = () => {
  fileInput.value?.click()
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  file.value = target.files?.[0] || null
}

const handleClose = () => {
  if (importing.value) return
  emit('close')
}

const isZipFile = (sourceFile: File): boolean => {
  const lowerName = sourceFile.name.toLowerCase()
  return sourceFile.type === 'application/zip' || lowerName.endsWith('.zip')
}

const looksLikeJSON = (input: string): boolean => {
  const trimmed = input.trim()
  return trimmed.startsWith('{') || trimmed.startsWith('[')
}

const isLikelyRawCodexContent = (input: string): boolean => {
  const trimmed = input.trim()
  if (!trimmed || /\s/.test(trimmed)) {
    return false
  }

  return trimmed.length >= 20
}

const readFileAsText = async (sourceFile: File): Promise<string> => {
  if (typeof sourceFile.text === 'function') {
    return sourceFile.text()
  }

  if (typeof sourceFile.arrayBuffer === 'function') {
    const buffer = await sourceFile.arrayBuffer()
    return new TextDecoder().decode(buffer)
  }

  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(reader.error || new Error('Failed to read file'))
    reader.readAsText(sourceFile)
  })
}

const normalizeDataPayload = (raw: unknown, sourceName: string): AdminDataPayload => {
  if (!raw || typeof raw !== 'object') {
    throw new Error(t('admin.accounts.dataImportInvalidPayload', { file: sourceName }))
  }

  const payload = raw as Partial<AdminDataPayload>
  if (!Array.isArray(payload.proxies) || !Array.isArray(payload.accounts)) {
    throw new Error(t('admin.accounts.dataImportInvalidPayload', { file: sourceName }))
  }

  return {
    type: payload.type,
    version: payload.version,
    exported_at: typeof payload.exported_at === 'string' && payload.exported_at.trim()
      ? payload.exported_at
      : new Date().toISOString(),
    proxies: payload.proxies,
    accounts: payload.accounts,
  }
}

const isDataPayloadLike = (raw: unknown): raw is Partial<AdminDataPayload> => {
  if (!raw || typeof raw !== 'object') {
    return false
  }

  const payload = raw as Partial<AdminDataPayload>
  return Array.isArray(payload.proxies) && Array.isArray(payload.accounts)
}

const mergeDataPayloads = (payloads: AdminDataPayload[]): AdminDataPayload => {
  if (!payloads.length) {
    throw new Error(t('admin.accounts.dataImportZipNoJsonFiles'))
  }

  return {
    type: payloads.find((item) => item.type)?.type,
    version: payloads.find((item) => typeof item.version === 'number')?.version,
    exported_at: payloads[0]?.exported_at || new Date().toISOString(),
    proxies: payloads.flatMap((item) => item.proxies),
    accounts: payloads.flatMap((item) => item.accounts),
  }
}

const readZipImportRequest = async (sourceFile: File): Promise<ImportRequestState> => {
  const zip = await JSZip.loadAsync(await sourceFile.arrayBuffer())
  const zipEntries = Object.values(zip.files) as JSZip.JSZipObject[]
  const jsonEntries = zipEntries
    .filter((entry) => !entry.dir && entry.name.toLowerCase().endsWith('.json'))

  if (!jsonEntries.length) {
    throw new Error(t('admin.accounts.dataImportZipNoJsonFiles'))
  }

  const payloads: AdminDataPayload[] = []
  const codexContents: string[] = []
  let hasDataPayload = false
  let hasCodexPayload = false

  for (const entry of jsonEntries) {
    const text = await entry.async('text')
    const trimmed = text.trim()
    if (!trimmed) {
      continue
    }

    if (!looksLikeJSON(trimmed)) {
      if (!isLikelyRawCodexContent(trimmed)) {
        throw new Error(t('admin.accounts.dataImportZipJsonParseFailed', { file: entry.name }))
      }
      codexContents.push(trimmed)
      hasCodexPayload = true
      continue
    }

    if (looksLikeJSON(trimmed)) {
      let parsed: unknown
      try {
        parsed = JSON.parse(text)
      } catch {
        throw new Error(t('admin.accounts.dataImportZipJsonParseFailed', { file: entry.name }))
      }

      if (isDataPayloadLike(parsed)) {
        payloads.push(normalizeDataPayload(parsed, entry.name))
        hasDataPayload = true
        continue
      }
    }

    codexContents.push(trimmed)
    hasCodexPayload = true
  }

  if (!payloads.length && !codexContents.length) {
    throw new Error(t('admin.accounts.dataImportZipNoJsonFiles'))
  }

  if (hasDataPayload && hasCodexPayload) {
    throw new Error(t('admin.accounts.dataImportZipMixedFormats'))
  }

  if (hasDataPayload) {
    return {
      kind: 'data',
      payload: mergeDataPayloads(payloads),
    }
  }

  return {
    kind: 'codex',
    contents: codexContents,
  }
}

const readImportRequest = async (sourceFile: File): Promise<ImportRequestState> => {
  if (isZipFile(sourceFile)) {
    return readZipImportRequest(sourceFile)
  }

  const text = await readFileAsText(sourceFile)
  const trimmed = text.trim()
  if (looksLikeJSON(trimmed)) {
    const parsed = JSON.parse(text)
    if (isDataPayloadLike(parsed)) {
      return {
        kind: 'data',
        payload: normalizeDataPayload(parsed, sourceFile.name),
      }
    }

    return {
      kind: 'codex',
      contents: [trimmed],
    }
  }

  if (!isLikelyRawCodexContent(trimmed)) {
    throw new SyntaxError('Invalid import payload')
  }

  return {
    kind: 'codex',
    contents: [trimmed],
  }
}

const handleImport = async () => {
  if (!file.value) {
    appStore.showError(t('admin.accounts.dataImportSelectFile'))
    return
  }

  importing.value = true
  try {
    const importRequest = await readImportRequest(file.value)

    if (importRequest.kind === 'data') {
      const res = await adminAPI.accounts.importData({
        data: importRequest.payload,
        skip_default_group_bind: true
      })

      result.value = { kind: 'data', value: res }

      const msgParams: Record<string, unknown> = {
        account_created: res.account_created,
        account_failed: res.account_failed,
        proxy_created: res.proxy_created,
        proxy_reused: res.proxy_reused,
        proxy_failed: res.proxy_failed,
      }
      if (res.account_failed > 0 || res.proxy_failed > 0) {
        appStore.showError(t('admin.accounts.dataImportCompletedWithErrors', msgParams))
      } else {
        appStore.showSuccess(t('admin.accounts.dataImportSuccess', msgParams))
        emit('imported')
      }
      return
    }

    const res = await adminAPI.accounts.importCodexSession({
      contents: importRequest.contents,
      skip_default_group_bind: true,
      update_existing: true,
    })

    result.value = { kind: 'codex', value: res }

    const successCount = res.created + res.updated
    const msgParams = {
      created: res.created,
      updated: res.updated,
      skipped: res.skipped,
      failed: res.failed,
    }

    if (successCount > 0 && res.failed === 0) {
      appStore.showSuccess(t('admin.accounts.oauth.openai.codexSessionImportSuccess', msgParams))
      emit('imported')
      return
    }

    if (successCount > 0) {
      appStore.showWarning(t('admin.accounts.oauth.openai.codexSessionImportPartial', msgParams))
      emit('imported')
      return
    }

    appStore.showError(t('admin.accounts.oauth.openai.codexSessionImportFailed'))
  } catch (error: any) {
    if (error instanceof SyntaxError) {
      appStore.showError(t('admin.accounts.dataImportParseFailed'))
    } else {
      appStore.showError(error?.message || t('admin.accounts.dataImportFailed'))
    }
  } finally {
    importing.value = false
  }
}
</script>
