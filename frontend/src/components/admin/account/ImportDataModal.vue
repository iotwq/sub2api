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
          class="flex items-center justify-between gap-3 rounded-lg border border-dashed px-4 py-3 transition-colors"
          :class="dragActive
            ? 'border-primary-400 bg-primary-50/70 dark:border-primary-500 dark:bg-primary-900/20'
            : 'border-gray-300 bg-gray-50 dark:border-dark-600 dark:bg-dark-800'"
          @dragenter.prevent="handleDragEnter"
          @dragover.prevent
          @dragleave.prevent="handleDragLeave"
          @drop.prevent="handleDrop"
        >
          <div class="min-w-0">
            <div class="truncate text-sm text-gray-700 dark:text-dark-200" :title="fileListTitle">
              {{ selectedFilesLabel || t('admin.accounts.dataImportSelectFile') }}
            </div>
            <div class="text-xs text-gray-500 dark:text-dark-400">
              JSON (.json), ZIP (.zip)
              <span v-if="files.length > 1"> · {{ fileListTitle }}</span>
            </div>
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
          multiple
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

type ImportResultState =
  | { kind: 'data'; value: AdminDataImportResult }
  | { kind: 'codex'; value: CodexSessionImportResult }

type ImportRequestState =
  | { kind: 'data'; payload: AdminDataPayload }
  | { kind: 'codex'; contents: string[] }

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const importing = ref(false)
const files = ref<File[]>([])
const dragDepth = ref(0)
const dragActive = computed(() => dragDepth.value > 0)
const hasCreatedData = ref(false)
const result = ref<ImportResultState | null>(null)

const fileInput = ref<HTMLInputElement | null>(null)
const selectedFilesLabel = computed(() => {
  if (files.value.length === 0) return ''
  if (files.value.length === 1) return files.value[0]?.name || ''
  return t('admin.accounts.selectedCount', { count: files.value.length })
})
const fileListTitle = computed(() => files.value.map((item) => item.name).join(', '))

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
      files.value = []
      dragDepth.value = 0
      hasCreatedData.value = false
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
  setSelectedFiles(target.files)
  target.value = ''
}

const handleClose = () => {
  if (importing.value) return
  if (hasCreatedData.value) {
    hasCreatedData.value = false
    emit('imported')
  }
  emit('close')
}

const isZipFile = (sourceFile: File): boolean => {
  const lowerName = sourceFile.name.toLowerCase()
  return sourceFile.type === 'application/zip' || lowerName.endsWith('.zip')
}

const isImportFile = (sourceFile: File): boolean => {
  const lowerName = sourceFile.name.toLowerCase()
  return lowerName.endsWith('.json') || lowerName.endsWith('.zip') ||
    sourceFile.type === 'application/json' || sourceFile.type === 'application/zip'
}

const setSelectedFiles = (sourceFiles: FileList | File[] | null | undefined) => {
  if (importing.value) return
  const incoming = Array.from(sourceFiles || [])
  const picked = incoming.filter(isImportFile)
  if (!picked.length) {
    appStore.showError(t('admin.accounts.dataImportSelectFile'))
    return
  }
  if (picked.length < incoming.length) {
    appStore.showWarning(
      t('admin.accounts.dataImportIgnoredFiles', { count: incoming.length - picked.length })
    )
  }
  files.value = picked
  result.value = null
}

const handleDragEnter = () => {
  if (importing.value) return
  dragDepth.value += 1
}

const handleDragLeave = () => {
  dragDepth.value = Math.max(0, dragDepth.value - 1)
}

const handleDrop = (event: DragEvent) => {
  dragDepth.value = 0
  if (importing.value) return
  setSelectedFiles(event.dataTransfer?.files)
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

const SUPPORTED_DATA_TYPES = ['sub2api-data', 'sub2api-bundle']
const SUPPORTED_DATA_VERSION = 1

const isValidDataPayload = (payload: unknown): payload is AdminDataPayload => {
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) return false
  const candidate = payload as Record<string, unknown>
  if (
    candidate.type !== undefined &&
    candidate.type !== '' &&
    !SUPPORTED_DATA_TYPES.includes(candidate.type as string)
  ) {
    return false
  }
  if (
    candidate.version !== undefined &&
    candidate.version !== 0 &&
    candidate.version !== SUPPORTED_DATA_VERSION
  ) {
    return false
  }
  return Array.isArray(candidate.proxies) && Array.isArray(candidate.accounts)
}

const normalizeDataPayload = (raw: unknown, sourceName: string): AdminDataPayload => {
  if (!isValidDataPayload(raw)) {
    throw new Error(t('admin.accounts.dataImportInvalidPayload', { file: sourceName }))
  }

  const payload = raw as AdminDataPayload
  return {
    type: payload.type,
    version: payload.version,
    exported_at: typeof payload.exported_at === 'string' && payload.exported_at.trim()
      ? payload.exported_at
      : new Date().toISOString(),
    proxies: payload.proxies,
    accounts: payload.accounts,
    skipped_shadows: payload.skipped_shadows,
  }
}

const isCodexSessionPayloadLike = (raw: unknown): boolean => {
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) {
    return false
  }

  const candidate = raw as Record<string, unknown>
  if (
    typeof candidate.accessToken === 'string' ||
    typeof candidate.access_token === 'string' ||
    typeof candidate.sessionToken === 'string'
  ) {
    return true
  }

  const tokens = candidate.tokens
  return Boolean(
    tokens &&
    typeof tokens === 'object' &&
    !Array.isArray(tokens) &&
    typeof (tokens as Record<string, unknown>).access_token === 'string'
  )
}

const mergeDataPayloads = (payloads: AdminDataPayload[]): AdminDataPayload => {
  const [firstPayload] = payloads
  if (!firstPayload) {
    throw new Error(t('admin.accounts.dataImportZipNoJsonFiles'))
  }
  if (payloads.length === 1) return firstPayload

  return {
    type: payloads.find((item) => typeof item.type === 'string')?.type,
    version: payloads.find((item) => typeof item.version === 'number')?.version,
    exported_at: firstPayload.exported_at || new Date().toISOString(),
    proxies: payloads.flatMap((item) => item.proxies),
    accounts: payloads.flatMap((item) => item.accounts),
    skipped_shadows: payloads.reduce((sum, item) => {
      const count = Number(item.skipped_shadows || 0)
      return Number.isFinite(count) ? sum + count : sum
    }, 0)
  }
}

const classifyTextImport = (text: string, sourceName: string): ImportRequestState => {
  const trimmed = text.trim()
  if (!trimmed) {
    throw new Error(t('admin.accounts.dataImportInvalidPayload', { file: sourceName }))
  }

  if (!looksLikeJSON(trimmed)) {
    if (!isLikelyRawCodexContent(trimmed)) {
      throw new Error(t('admin.accounts.dataImportZipJsonParseFailed', { file: sourceName }))
    }
    return { kind: 'codex', contents: [trimmed] }
  }

  let parsed: unknown
  try {
    parsed = JSON.parse(text)
  } catch {
    throw new Error(t('admin.accounts.dataImportZipJsonParseFailed', { file: sourceName }))
  }

  if (isValidDataPayload(parsed)) {
    return { kind: 'data', payload: normalizeDataPayload(parsed, sourceName) }
  }
  if (isCodexSessionPayloadLike(parsed)) {
    return { kind: 'codex', contents: [trimmed] }
  }

  throw new Error(t('admin.accounts.dataImportInvalidPayload', { file: sourceName }))
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

  for (const entry of jsonEntries) {
    const item = classifyTextImport(await entry.async('text'), entry.name)
    if (item.kind === 'data') {
      payloads.push(item.payload)
    } else {
      codexContents.push(...item.contents)
    }
  }

  if (payloads.length && codexContents.length) {
    throw new Error(t('admin.accounts.dataImportZipMixedFormats'))
  }
  if (payloads.length) {
    return { kind: 'data', payload: mergeDataPayloads(payloads) }
  }
  if (codexContents.length) {
    return { kind: 'codex', contents: codexContents }
  }

  throw new Error(t('admin.accounts.dataImportZipNoJsonFiles'))
}

const readImportRequest = async (sourceFile: File): Promise<ImportRequestState> => {
  if (isZipFile(sourceFile)) {
    return readZipImportRequest(sourceFile)
  }

  try {
    return classifyTextImport(await readFileAsText(sourceFile), sourceFile.name)
  } catch (error) {
    if (error instanceof SyntaxError) {
      throw new Error(t('admin.accounts.dataImportParseFailedFile', { name: sourceFile.name }))
    }
    throw error
  }
}

const readSelectedImportRequest = async (): Promise<ImportRequestState> => {
  const payloads: AdminDataPayload[] = []
  const codexContents: string[] = []

  for (const sourceFile of files.value) {
    const item = await readImportRequest(sourceFile)
    if (item.kind === 'data') {
      payloads.push(item.payload)
    } else {
      codexContents.push(...item.contents)
    }
  }

  if (payloads.length && codexContents.length) {
    throw new Error(t('admin.accounts.dataImportZipMixedFormats'))
  }
  if (payloads.length) {
    return { kind: 'data', payload: mergeDataPayloads(payloads) }
  }
  if (codexContents.length) {
    return { kind: 'codex', contents: codexContents }
  }

  throw new Error(t('admin.accounts.dataImportSelectFile'))
}

const handleImport = async () => {
  if (files.value.length === 0) {
    appStore.showError(t('admin.accounts.dataImportSelectFile'))
    return
  }

  importing.value = true
  try {
    const importRequest = await readSelectedImportRequest()

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
        if (res.account_created > 0 || res.proxy_created > 0) {
          hasCreatedData.value = true
        }
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
    appStore.showError(error?.message || t('admin.accounts.dataImportFailed'))
  } finally {
    importing.value = false
  }
}
</script>
