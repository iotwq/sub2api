<template>
  <button
    ref="triggerRef"
    type="button"
    class="community-emoji-trigger"
    :class="{ 'community-emoji-trigger-active': open }"
    :title="label"
    :aria-label="label"
    aria-haspopup="dialog"
    :aria-expanded="open"
    aria-controls="community-emoji-picker-popover"
    @click="toggle"
  >
    <Icon name="smile" size="md" />
  </button>

  <Teleport to="body">
    <div
      v-if="open"
      id="community-emoji-picker-popover"
      ref="popoverRef"
      class="community-emoji-popover"
      role="dialog"
      :aria-label="label"
      :style="popoverStyle"
    >
      <div ref="pickerHostRef" class="community-emoji-picker-host"></div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { Picker } from 'emoji-picker-element'
import englishI18n from 'emoji-picker-element/i18n/en'
import chineseI18n from 'emoji-picker-element/i18n/zh_CN'
import englishEmojiDataURL from 'emoji-picker-element-data/en/cldr/data.json?url'
import chineseEmojiDataURL from 'emoji-picker-element-data/zh/cldr/data.json?url'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  locale: string
  label: string
}>()

const emit = defineEmits<{
  select: [emoji: string]
}>()

const viewportMargin = 8
const popoverGap = 10
const preferredWidth = 360
const preferredHeight = 420

const open = ref(false)
const triggerRef = ref<HTMLButtonElement | null>(null)
const popoverRef = ref<HTMLElement | null>(null)
const pickerHostRef = ref<HTMLElement | null>(null)
const popoverStyle = ref<Record<string, string>>({
  left: `${viewportMargin}px`,
  top: `${viewportMargin}px`,
  width: `${preferredWidth}px`,
  height: `${preferredHeight}px`,
})

let picker: Picker | null = null
let themeObserver: MutationObserver | null = null

const usesChinese = computed(() => props.locale.toLowerCase().startsWith('zh'))

function createPicker(): Picker {
  const baseI18n = usesChinese.value ? chineseI18n : englishI18n
  const instance = new Picker({
    locale: usesChinese.value ? 'zh' : 'en',
    dataSource: usesChinese.value ? chineseEmojiDataURL : englishEmojiDataURL,
    i18n: {
      ...baseI18n,
      favoritesLabel: usesChinese.value ? '常用表情' : 'Frequently used',
    },
  })
  instance.classList.add('community-emoji-picker')
  instance.addEventListener('emoji-click', handleEmojiClick)
  updatePickerTheme(instance)
  return instance
}

function mountPicker(): void {
  if (!pickerHostRef.value) return
  if (!picker) picker = createPicker()
  if (picker.parentElement !== pickerHostRef.value) pickerHostRef.value.appendChild(picker)
}

function destroyPicker(): void {
  if (!picker) return
  picker.removeEventListener('emoji-click', handleEmojiClick)
  picker.remove()
  picker = null
}

function handleEmojiClick(event: Event): void {
  const unicode = (event as CustomEvent<{ unicode?: string }>).detail.unicode
  if (unicode) emit('select', unicode)
}

async function toggle(): Promise<void> {
  open.value = !open.value
  if (!open.value) return

  await nextTick()
  updatePosition()
  mountPicker()
}

function close(options: { restoreFocus?: boolean } = {}): void {
  if (!open.value) return
  open.value = false
  if (options.restoreFocus) void nextTick(() => triggerRef.value?.focus())
}

function updatePosition(): void {
  const trigger = triggerRef.value
  if (!trigger) return

  const visualViewport = window.visualViewport
  const viewportWidth = visualViewport?.width ?? window.innerWidth
  const viewportHeight = visualViewport?.height ?? window.innerHeight
  const viewportOffsetLeft = visualViewport?.offsetLeft ?? 0
  const viewportOffsetTop = visualViewport?.offsetTop ?? 0
  const width = Math.min(preferredWidth, Math.max(1, viewportWidth - viewportMargin * 2))
  const triggerRect = trigger.getBoundingClientRect()
  const minLeft = viewportOffsetLeft + viewportMargin
  const maxLeft = viewportOffsetLeft + viewportWidth - width - viewportMargin
  const left = Math.min(Math.max(triggerRect.right - width, minLeft), maxLeft)
  const minTop = viewportOffsetTop + viewportMargin
  const viewportBottom = viewportOffsetTop + viewportHeight - viewportMargin
  const availableAbove = Math.max(0, triggerRect.top - minTop - popoverGap)
  const availableBelow = Math.max(0, viewportBottom - triggerRect.bottom - popoverGap)
  const opensAbove = availableAbove >= availableBelow
  const availableHeight = opensAbove ? availableAbove : availableBelow
  const height = Math.max(1, Math.min(preferredHeight, availableHeight))
  const maxTop = viewportOffsetTop + viewportHeight - height - viewportMargin
  const top = opensAbove
    ? Math.max(minTop, triggerRect.top - popoverGap - height)
    : Math.min(triggerRect.bottom + popoverGap, maxTop)

  popoverStyle.value = {
    left: `${Math.round(left)}px`,
    top: `${Math.round(top)}px`,
    width: `${Math.round(width)}px`,
    height: `${Math.round(height)}px`,
  }
}

function handleDocumentPointerDown(event: PointerEvent): void {
  const target = event.target as Node
  if (triggerRef.value?.contains(target) || popoverRef.value?.contains(target)) return
  close()
}

function handleDocumentKeyDown(event: KeyboardEvent): void {
  if (event.key === 'Escape') close({ restoreFocus: true })
}

function updatePickerTheme(instance: Picker | null = picker): void {
  if (!instance) return
  const dark = document.documentElement.classList.contains('dark')
  instance.classList.toggle('dark', dark)
  instance.classList.toggle('light', !dark)
}

function handleViewportChange(): void {
  if (open.value) updatePosition()
}

watch(() => props.locale, () => {
  destroyPicker()
  if (open.value) void nextTick(mountPicker)
})

watch(open, (value) => {
  if (value) updatePickerTheme()
})

document.addEventListener('pointerdown', handleDocumentPointerDown, true)
document.addEventListener('keydown', handleDocumentKeyDown)
window.addEventListener('resize', handleViewportChange)
window.addEventListener('scroll', handleViewportChange, true)
window.visualViewport?.addEventListener('resize', handleViewportChange)
window.visualViewport?.addEventListener('scroll', handleViewportChange)
themeObserver = new MutationObserver(() => updatePickerTheme())
themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', handleDocumentPointerDown, true)
  document.removeEventListener('keydown', handleDocumentKeyDown)
  window.removeEventListener('resize', handleViewportChange)
  window.removeEventListener('scroll', handleViewportChange, true)
  window.visualViewport?.removeEventListener('resize', handleViewportChange)
  window.visualViewport?.removeEventListener('scroll', handleViewportChange)
  themeObserver?.disconnect()
  destroyPicker()
})
</script>

<style scoped>
.community-emoji-trigger {
  @apply inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-gray-800 transition-colors hover:bg-[#d1c5b7] focus:outline-none focus:ring-2 focus:ring-[#8f382f]/30 disabled:cursor-not-allowed disabled:opacity-50 dark:text-gray-300 dark:hover:bg-[#202b39] dark:focus:ring-[#f0b4a8]/25;
}

.community-emoji-trigger-active {
  @apply bg-[#d1c5b7] text-[#7a1f1f] dark:bg-[#202b39] dark:text-[#f0b4a8];
}

.community-emoji-popover {
  position: fixed;
  z-index: 80;
  overflow: hidden;
  border: 1px solid #a99b8c;
  border-radius: 8px;
  background: #d8cdc0;
  box-shadow: 0 22px 55px rgba(35, 25, 17, 0.28);
}

:global(.dark) .community-emoji-popover {
  border-color: #374151;
  background: #18212c;
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.45);
}

.community-emoji-picker-host {
  width: 100%;
  height: 100%;
}

.community-emoji-picker-host :deep(.community-emoji-picker) {
  width: 100%;
  height: 100%;
  --background: #d8cdc0;
  --border-color: #a99b8c;
  --border-radius: 0;
  --button-active-background: #c1b4a6;
  --button-hover-background: #cbbfb1;
  --indicator-color: #8f382f;
  --input-border-color: #9f9284;
  --input-font-color: #2f2923;
  --input-placeholder-color: #6b625a;
  --outline-color: #8f382f;
  --emoji-size: 1.55rem;
  --emoji-padding: 0.55rem;
}

:global(.dark) .community-emoji-picker-host :deep(.community-emoji-picker) {
  --background: #18212c;
  --border-color: #374151;
  --button-active-background: #374151;
  --button-hover-background: #273244;
  --indicator-color: #f0b4a8;
  --input-border-color: #526174;
  --input-font-color: #f3f4f6;
  --input-placeholder-color: #9ca3af;
  --outline-color: #f0b4a8;
}

@media (max-width: 380px) {
  .community-emoji-picker-host :deep(.community-emoji-picker) {
    --num-columns: 7;
    --category-emoji-size: 1.15rem;
    --emoji-size: 1.4rem;
    --emoji-padding: 0.45rem;
  }
}
</style>
