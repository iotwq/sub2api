<template>
  <div class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-800">
    <!-- Collapsed summary header (clickable) -->
    <div
      class="flex cursor-pointer select-none items-center gap-2"
      @click="collapsed = !collapsed"
    >
      <Icon
        :name="collapsed ? 'chevronRight' : 'chevronDown'"
        size="sm"
        :stroke-width="2"
        class="flex-shrink-0 text-gray-400 transition-transform duration-200"
      />

      <!-- Summary: model tags + billing badge -->
      <div v-if="collapsed" class="flex min-w-0 flex-1 items-center gap-2 overflow-hidden">
        <!-- Compact model tags (show first 3) -->
        <div class="flex min-w-0 flex-1 flex-wrap items-center gap-1">
          <span
            v-for="(m, i) in entry.models.slice(0, 3)"
            :key="i"
            class="inline-flex shrink-0 rounded px-1.5 py-0.5 text-xs"
            :class="getPlatformTagClass(props.platform || '')"
          >
            {{ m }}
          </span>
          <span
            v-if="entry.models.length > 3"
            class="whitespace-nowrap text-xs text-gray-400"
          >
            +{{ entry.models.length - 3 }}
          </span>
          <span
            v-if="entry.models.length === 0"
            class="text-xs italic text-gray-400"
          >
            {{ t('admin.channels.form.noModels') }}
          </span>
        </div>

        <!-- Billing mode badge -->
        <span
          class="flex-shrink-0 rounded-full bg-primary-100 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
        >
          {{ billingModeLabel }}
        </span>
      </div>

      <!-- Expanded: show the label "Pricing Entry" or similar -->
      <div v-else class="flex-1 text-xs font-medium text-gray-500 dark:text-gray-400">
        {{ t('admin.channels.form.pricingEntry') }}
      </div>

      <!-- Remove button (always visible, stop propagation) -->
      <button
        type="button"
        @click.stop="emit('remove')"
        class="flex-shrink-0 rounded p-1 text-gray-400 hover:text-red-500"
      >
        <Icon name="trash" size="sm" />
      </button>
    </div>

    <!-- Expandable content with transition -->
    <div
      class="collapsible-content"
      :class="{ 'collapsible-content--collapsed': collapsed }"
    >
      <div class="collapsible-inner">
        <!-- Header: Models + Billing Mode -->
        <div class="mt-3 flex items-start gap-2">
          <div class="flex-1">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.channels.form.models') }} <span class="text-red-500">*</span>
            </label>
            <ModelTagInput
              :models="entry.models"
              :platform="props.platform"
              @update:models="onModelsUpdate($event)"
              :placeholder="t('admin.channels.form.modelsPlaceholder')"
              class="mt-1"
            />
          </div>
          <div class="w-40">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.channels.form.billingMode') }}
            </label>
            <Select
              :modelValue="entry.billing_mode"
              @update:modelValue="onBillingModeUpdate($event as BillingMode)"
              :options="billingModeOptions"
              class="mt-1"
            />
          </div>
        </div>

        <!-- Token mode -->
        <div v-if="entry.billing_mode === 'token'">
          <!-- Default prices (fallback when no interval matches) -->
          <label class="mt-3 block text-xs font-medium text-gray-500 dark:text-gray-400">
            {{ t('admin.channels.form.defaultPrices') }}
            <span class="ml-1 font-normal text-gray-400">$/MTok</span>
          </label>
          <div class="pricing-default-grid mt-1 grid gap-2">
            <div>
              <label class="text-xs text-gray-400">{{ t('admin.channels.form.inputPrice') }}</label>
              <input :value="entry.input_price" @input="emitField('input_price', ($event.target as HTMLInputElement).value)"
                type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" />
            </div>
            <div>
              <label class="text-xs text-gray-400">{{ t('admin.channels.form.outputPrice') }}</label>
              <input :value="entry.output_price" @input="emitField('output_price', ($event.target as HTMLInputElement).value)"
                type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" />
            </div>
            <div>
              <label class="text-xs text-gray-400">{{ t('admin.channels.form.cacheWrite5mPrice') }}</label>
              <input :value="entry.cache_write_price" @input="emitField('cache_write_price', ($event.target as HTMLInputElement).value)"
                type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" />
            </div>
            <div>
              <label class="text-xs text-gray-400">{{ t('admin.channels.form.cacheWrite1hPrice') }}</label>
              <input :value="entry.cache_write_1h_price" @input="emitField('cache_write_1h_price', ($event.target as HTMLInputElement).value)"
                type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" />
            </div>
            <div>
              <label class="text-xs text-gray-400">{{ t('admin.channels.form.cacheReadPrice') }}</label>
              <input :value="entry.cache_read_price" @input="emitField('cache_read_price', ($event.target as HTMLInputElement).value)"
                type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" />
            </div>
            <div>
              <label class="text-xs text-gray-400">{{ t('admin.channels.form.imageInputPrice') }}</label>
              <input :value="entry.image_input_price" @input="emitField('image_input_price', ($event.target as HTMLInputElement).value)"
                type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" />
            </div>
            <div>
              <label class="text-xs text-gray-400">{{ t('admin.channels.form.imageTokenPrice') }}</label>
              <input :value="entry.image_output_price" @input="emitField('image_output_price', ($event.target as HTMLInputElement).value)"
                type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" />
            </div>
          </div>

          <div v-if="enableTierMultipliers" class="mt-3 grid max-w-md grid-cols-1 gap-2 sm:grid-cols-2">
            <div>
              <label class="text-xs text-gray-400">{{ t('admin.channels.form.fastMultiplier') }}</label>
              <input :value="entry.fast_multiplier" @input="emitField('fast_multiplier', ($event.target as HTMLInputElement).value)"
                type="number" step="any" min="0.000001" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.multiplierPlaceholder')" />
            </div>
            <div>
              <label class="text-xs text-gray-400">{{ t('admin.channels.form.flexMultiplier') }}</label>
              <input :value="entry.flex_multiplier" @input="emitField('flex_multiplier', ($event.target as HTMLInputElement).value)"
                type="number" step="any" min="0.000001" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.multiplierPlaceholder')" />
            </div>
          </div>

          <!-- Channel token intervals; the group long-context toggle controls whether tiers apply. -->
          <div v-if="!hideTokenIntervals" class="mt-3">
            <div class="flex items-center justify-between">
              <label class="text-xs font-medium text-gray-500 dark:text-gray-400">
                {{ t('admin.channels.form.intervals') }}
                <span class="ml-1 font-normal text-gray-400">(min, max]</span>
              </label>
              <button type="button" @click="addInterval" class="text-xs text-primary-600 hover:text-primary-700">
                + {{ t('admin.channels.form.addInterval') }}
              </button>
            </div>
            <div v-if="entry.intervals && entry.intervals.length > 0" class="mt-2 space-y-2">
              <IntervalRow
                v-for="(iv, idx) in entry.intervals"
                :key="idx"
                :interval="iv"
                :mode="entry.billing_mode"
                :enable-multipliers="enableTierMultipliers"
                @update="updateInterval(idx, $event)"
                @remove="removeInterval(idx)"
              />
            </div>
          </div>

          <TimePricingSection
            v-if="enableTimePricing"
            :model-value="entry.time_pricing"
            @update:model-value="emit('update', { ...entry, time_pricing: $event })"
          />
        </div>

        <!-- Per-request mode -->
        <div v-else-if="entry.billing_mode === 'per_request'">
          <!-- Default per-request price -->
          <label class="mt-3 block text-xs font-medium text-gray-500 dark:text-gray-400">
            {{ t('admin.channels.form.defaultPerRequestPrice') }}
            <span class="ml-1 font-normal text-gray-400">$</span>
          </label>
          <div class="mt-1 w-48">
            <input :value="entry.per_request_price" @input="emitField('per_request_price', ($event.target as HTMLInputElement).value)"
              type="number" step="any" min="0" class="input text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" />
          </div>
          <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">
            {{ t('admin.channels.form.perRequestVideoHint') }}
          </p>

          <!-- Tiers -->
          <div class="mt-3 flex items-center justify-between">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.channels.form.requestTiers') }}
            </label>
            <button type="button" @click="addInterval" class="text-xs text-primary-600 hover:text-primary-700">
              + {{ t('admin.channels.form.addTier') }}
            </button>
          </div>
          <div v-if="entry.intervals && entry.intervals.length > 0" class="mt-2 space-y-2">
            <IntervalRow
              v-for="(iv, idx) in entry.intervals"
              :key="idx"
              :interval="iv"
              :mode="entry.billing_mode"
              @update="updateInterval(idx, $event)"
              @remove="removeInterval(idx)"
            />
          </div>
          <div v-else class="mt-2 rounded border border-dashed border-gray-300 p-3 text-center text-xs text-gray-400 dark:border-dark-500">
            {{ t('admin.channels.form.noTiersYet') }}
          </div>
        </div>

        <!-- Image mode -->
        <div v-else-if="entry.billing_mode === 'image'">
          <!-- Default image price (per-request, same as per_request mode) -->
          <label class="mt-3 block text-xs font-medium text-gray-500 dark:text-gray-400">
            {{ t('admin.channels.form.defaultImagePrice') }}
            <span class="ml-1 font-normal text-gray-400">$</span>
          </label>
          <div class="mt-1 w-48">
            <input :value="entry.per_request_price" @input="emitField('per_request_price', ($event.target as HTMLInputElement).value)"
              type="number" step="any" min="0" class="input text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" />
          </div>

          <!-- Image tiers -->
          <div class="mt-3 flex items-center justify-between">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.channels.form.imageTiers') }}
            </label>
            <button type="button" @click="addMediaTier" class="text-xs text-primary-600 hover:text-primary-700">
              + {{ t('admin.channels.form.addTier') }}
            </button>
          </div>
          <div v-if="entry.intervals && entry.intervals.length > 0" class="mt-2 space-y-2">
            <IntervalRow
              v-for="(iv, idx) in entry.intervals"
              :key="idx"
              :interval="iv"
              :mode="entry.billing_mode"
              @update="updateInterval(idx, $event)"
              @remove="removeInterval(idx)"
            />
          </div>
        </div>

        <!-- Video per-second mode -->
        <div v-else-if="entry.billing_mode === 'video'" class="mt-3">
          <template v-if="isMiniMaxH3Only(entry.models)">
            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.channels.form.videoSecondPrices') }}
              <span class="ml-1 font-normal text-gray-400">$/{{ t('admin.channels.form.second') }}</span>
            </label>
            <div class="mt-2 grid max-w-lg grid-cols-2 gap-3">
              <div v-for="resolution in miniMaxVideoResolutions" :key="resolution">
                <label class="text-xs text-gray-400">{{ resolution }}</label>
                <input
                  :value="videoTierPrice(resolution)"
                  @input="updateVideoTierPrice(resolution, ($event.target as HTMLInputElement).value)"
                  type="number"
                  step="any"
                  min="0"
                  class="input mt-0.5 text-sm"
                  :placeholder="t('admin.channels.form.priceRequired')"
                />
              </div>
            </div>
            <p class="mt-2 text-xs text-gray-400 dark:text-gray-500">
              {{ t('admin.channels.form.miniMaxVideoBillingHint') }}
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">
              {{ t('admin.channels.form.miniMaxMediaLimitsHint') }}
            </p>
          </template>
          <template v-else-if="isSD20VideoOnly(entry.models)">
            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.channels.form.sd20VideoSecondPrices') }}
              <span class="ml-1 font-normal text-gray-400">$/{{ t('admin.channels.form.second') }}</span>
            </label>
            <div class="mt-2 grid max-w-2xl grid-cols-2 gap-3 sm:grid-cols-3">
              <div v-for="resolution in sd20VideoResolutions(entry.models[0])" :key="resolution">
                <label class="text-xs text-gray-400">{{ resolution }}</label>
                <input
                  :data-testid="`sd20-video-price-${resolution}`"
                  :value="videoTierPrice(resolution)"
                  @input="updateVideoTierPrice(resolution, ($event.target as HTMLInputElement).value)"
                  type="number"
                  step="any"
                  min="0"
                  class="input mt-0.5 text-sm"
                  :placeholder="t('admin.channels.form.priceRequired')"
                />
              </div>
            </div>
            <p class="mt-2 text-xs text-gray-400 dark:text-gray-500">
              {{ t('admin.channels.form.sd20VideoBillingHint') }}
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">
              {{ t('admin.channels.form.sd20VideoMediaFreeHint') }}
            </p>
          </template>
          <template v-else-if="includesSD20VideoModel(entry.models)">
            <p class="text-xs text-red-500 dark:text-red-400">
              {{ t('admin.channels.form.sd20VideoModeModelRequired') }}
            </p>
          </template>
          <template v-else>
            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.channels.form.genericVideoSecondPrice') }}
              <span class="ml-1 font-normal text-gray-400">$/{{ t('admin.channels.form.second') }}</span>
            </label>
            <div class="mt-2 w-48">
              <input
                data-testid="generic-video-second-price"
                :value="entry.per_request_price"
                @input="emitField('per_request_price', ($event.target as HTMLInputElement).value)"
                type="number"
                step="any"
                min="0"
                class="input text-sm"
                :placeholder="t('admin.channels.form.priceRequired')"
              />
            </div>
            <p class="mt-2 text-xs text-gray-400 dark:text-gray-500">
              {{ t('admin.channels.form.genericVideoBillingHint') }}
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">
              {{ t('admin.channels.form.genericVideoMediaFreeHint') }}
            </p>
          </template>
        </div>

        <div class="mt-3 border-t border-gray-200 pt-3 dark:border-dark-600" data-testid="reasoning-effort-multipliers">
          <div class="flex items-center justify-between gap-2">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.channels.form.reasoningEffortMultipliers') }}
            </label>
            <button
              v-if="Object.keys(entry.reasoning_effort_multipliers || {}).length"
              type="button"
              class="text-xs text-gray-500 hover:text-red-500"
              @click="emit('update', { ...entry, reasoning_effort_multipliers: null })"
            >
              {{ t('admin.channels.form.clearReasoningEffortMultipliers') }}
            </button>
          </div>
          <p class="mt-1 text-xs text-gray-400">{{ t('admin.channels.form.reasoningEffortMultipliersHint') }}</p>
          <div class="mt-2 grid grid-cols-2 gap-2 sm:grid-cols-4 lg:grid-cols-7">
            <label v-for="effort in REASONING_EFFORT_LEVELS" :key="effort" class="text-xs text-gray-500 dark:text-gray-400">
              {{ effort }}
              <input
                :value="entry.reasoning_effort_multipliers?.[effort]"
                :aria-label="t('admin.channels.form.reasoningEffortMultiplierLabel', { effort })"
                :aria-invalid="!isValidPositiveMultiplier(entry.reasoning_effort_multipliers?.[effort])"
                :data-reasoning-effort="effort"
                @input="updateReasoningEffortMultiplier(effort, ($event.target as HTMLInputElement).value)"
                type="number"
                step="any"
                min="0"
                class="input mt-0.5 text-sm"
                :placeholder="t('admin.channels.form.reasoningEffortMultiplierDefault')"
              />
            </label>
          </div>
          <p v-if="reasoningEffortMultiplierError" role="alert" class="mt-1 text-xs text-red-500">
            {{ reasoningEffortMultiplierError }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import IntervalRow from './IntervalRow.vue'
import ModelTagInput from './ModelTagInput.vue'
import TimePricingSection from './TimePricingSection.vue'
import type { PricingFormEntry, IntervalFormEntry } from './types'
import { perTokenToMTok, getPlatformTagClass, isValidPositiveMultiplier, validateReasoningEffortMultipliers } from './types'
import { REASONING_EFFORT_LEVELS, type ReasoningEffortLevel } from '@/constants/channel'
import type { BillingMode } from '@/api/admin/channels'
import channelsAPI from '@/api/admin/channels'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  entry: PricingFormEntry
  platform?: string
  hideTokenIntervals?: boolean
  enableTimePricing?: boolean
  enableTierMultipliers?: boolean
}>(), {
  hideTokenIntervals: false,
  enableTimePricing: false,
  enableTierMultipliers: false,
})

const emit = defineEmits<{
  update: [entry: PricingFormEntry]
  remove: []
}>()

// Collapse state: entries with existing models default to collapsed
const collapsed = ref(props.entry.models.length > 0)

const billingModeOptions = computed(() => [
  { value: 'token', label: t('admin.channels.billingMode.token') },
  { value: 'per_request', label: t('admin.channels.billingMode.perRequest') },
  { value: 'image', label: t('admin.channels.billingMode.image') },
  { value: 'video', label: t('admin.channels.billingMode.video') }
])

const miniMaxVideoResolutions = ['768P', '2K'] as const
const sd20StandardVideoResolutions = ['480p', '720p', '1080p'] as const
const sd20FastVideoResolutions = ['480p', '720p'] as const

const billingModeLabel = computed(() => {
  const opt = billingModeOptions.value.find(o => o.value === props.entry.billing_mode)
  return opt ? opt.label : props.entry.billing_mode
})

const reasoningEffortMultiplierError = computed(() =>
  validateReasoningEffortMultipliers(props.entry.reasoning_effort_multipliers, t)
)

function updateReasoningEffortMultiplier(effort: ReasoningEffortLevel, value: string) {
  const multipliers = { ...props.entry.reasoning_effort_multipliers }
  if (value === '') delete multipliers[effort]
  else multipliers[effort] = value
  emit('update', {
    ...props.entry,
    reasoning_effort_multipliers: Object.keys(multipliers).length ? multipliers : null,
  })
}

function emitField(field: keyof PricingFormEntry, value: string) {
  emit('update', { ...props.entry, [field]: value === '' ? null : value })
}

function isMiniMaxH3Only(models: string[]) {
  return models.length === 1 && models[0].trim().toLowerCase() === 'minimax-h3'
}

function isFireflyVideoModel(model: string) {
  const normalized = model.trim().toLowerCase()
  return normalized === 'firefly-video-v2' || normalized === 'firefly-video-v2-fast'
}

function isSD20VideoOnly(models: string[]) {
  return models.length === 1 && isFireflyVideoModel(models[0])
}

function includesSD20VideoModel(models: string[]) {
  return models.some(isFireflyVideoModel)
}

function sd20VideoResolutions(model: string): readonly string[] {
  return model.trim().toLowerCase() === 'firefly-video-v2-fast'
    ? sd20FastVideoResolutions
    : sd20StandardVideoResolutions
}

function miniMaxVideoTemplate(models: string[]): PricingFormEntry {
  return {
    ...props.entry,
    models,
    billing_mode: 'video',
    input_price: null,
    output_price: null,
    cache_write_price: null,
    cache_read_price: null,
    image_input_price: null,
    image_output_price: null,
    per_request_price: null,
    intervals: miniMaxVideoResolutions.map((tierLabel, index) => ({
      min_tokens: 0,
      max_tokens: null,
      tier_label: tierLabel,
      input_price: null,
      output_price: null,
      cache_write_price: null,
      cache_read_price: null,
      input_multiplier: null,
      output_multiplier: null,
      cache_write_multiplier: null,
      cache_read_multiplier: null,
      per_request_price: null,
      sort_order: index,
    })),
  }
}

function sd20VideoTemplate(models: string[]): PricingFormEntry {
  const resolutions = models.length === 1 ? sd20VideoResolutions(models[0]) : []
  return {
    ...props.entry,
    models,
    billing_mode: 'video',
    input_price: null,
    output_price: null,
    cache_write_price: null,
    cache_read_price: null,
    image_input_price: null,
    image_output_price: null,
    per_request_price: null,
    intervals: resolutions.map((tierLabel, index) => ({
      min_tokens: 0,
      max_tokens: null,
      tier_label: tierLabel,
      input_price: null,
      output_price: null,
      cache_write_price: null,
      cache_read_price: null,
      input_multiplier: null,
      output_multiplier: null,
      cache_write_multiplier: null,
      cache_read_multiplier: null,
      per_request_price: null,
      sort_order: index,
    })),
  }
}

function genericVideoTemplate(models: string[], preservePrice = false): PricingFormEntry {
  return {
    ...props.entry,
    models,
    billing_mode: 'video',
    input_price: null,
    output_price: null,
    cache_write_price: null,
    cache_read_price: null,
    image_input_price: null,
    image_output_price: null,
    per_request_price: preservePrice ? props.entry.per_request_price : null,
    intervals: [],
  }
}

function onBillingModeUpdate(mode: BillingMode) {
  if (mode === 'video' && isMiniMaxH3Only(props.entry.models)) {
    emit('update', {
      ...miniMaxVideoTemplate(props.entry.models),
      time_pricing: { ...props.entry.time_pricing, periods: [] },
    })
    return
  }
  if (mode === 'video' && isSD20VideoOnly(props.entry.models)) {
    emit('update', {
      ...sd20VideoTemplate(props.entry.models),
      time_pricing: { ...props.entry.time_pricing, periods: [] },
    })
    return
  }
  if (mode === 'video') {
    emit('update', {
      ...genericVideoTemplate(props.entry.models),
      time_pricing: { ...props.entry.time_pricing, periods: [] },
    })
    return
  }
  emit('update', {
    ...props.entry,
    billing_mode: mode,
    intervals: [],
    time_pricing: { ...props.entry.time_pricing, periods: [] },
  })
}

function videoTierPrice(resolution: string) {
  return props.entry.intervals.find(iv => iv.tier_label.toUpperCase() === resolution.toUpperCase())?.per_request_price ?? null
}

function updateVideoTierPrice(resolution: string, value: string) {
  const intervals = [...props.entry.intervals]
  const index = intervals.findIndex(iv => iv.tier_label.toUpperCase() === resolution.toUpperCase())
  if (index < 0) {
    intervals.push({
      min_tokens: 0,
      max_tokens: null,
      tier_label: resolution,
      input_price: null,
      output_price: null,
      cache_write_price: null,
      cache_read_price: null,
      input_multiplier: null,
      output_multiplier: null,
      cache_write_multiplier: null,
      cache_read_multiplier: null,
      per_request_price: value === '' ? null : value,
      sort_order: intervals.length,
    })
  } else {
    intervals[index] = { ...intervals[index], per_request_price: value === '' ? null : value }
  }
  emit('update', { ...props.entry, per_request_price: null, intervals })
}

function addInterval() {
  const intervals = [...(props.entry.intervals || [])]
  intervals.push({
    min_tokens: 0, max_tokens: null, tier_label: '',
    input_price: null, output_price: null, cache_write_price: null,
    cache_write_1h_price: null,
    cache_read_price: null, per_request_price: null,
    input_multiplier: null, output_multiplier: null,
    cache_write_multiplier: null, cache_read_multiplier: null,
    sort_order: intervals.length
  })
  emit('update', { ...props.entry, intervals })
}

function addMediaTier() {
  const intervals = [...(props.entry.intervals || [])]
  const labels = props.entry.billing_mode === 'video'
    ? ['480p', '720p', '1080p']
    : ['1K', '2K', '4K', 'HD']
  intervals.push({
    min_tokens: 0, max_tokens: null, tier_label: labels[intervals.length] || '',
    input_price: null, output_price: null, cache_write_price: null,
    cache_write_1h_price: null,
    cache_read_price: null, per_request_price: null,
    input_multiplier: null, output_multiplier: null,
    cache_write_multiplier: null, cache_read_multiplier: null,
    sort_order: intervals.length
  })
  emit('update', { ...props.entry, intervals })
}

function updateInterval(idx: number, updated: IntervalFormEntry) {
  const intervals = [...(props.entry.intervals || [])]
  intervals[idx] = updated
  emit('update', { ...props.entry, intervals })
}

function removeInterval(idx: number) {
  const intervals = [...(props.entry.intervals || [])]
  intervals.splice(idx, 1)
  emit('update', { ...props.entry, intervals })
}

async function onModelsUpdate(newModels: string[]) {
  const oldModels = props.entry.models

  if (isMiniMaxH3Only(newModels)) {
    emit('update', miniMaxVideoTemplate(['MiniMax-H3']))
    return
  }
  if (isSD20VideoOnly(newModels)) {
    emit('update', sd20VideoTemplate(newModels))
    return
  }
  if (props.entry.billing_mode === 'video' || includesSD20VideoModel(newModels)) {
    emit('update', genericVideoTemplate(newModels, props.entry.billing_mode === 'video' && props.entry.intervals.length === 0))
    return
  }
  emit('update', { ...props.entry, models: newModels })

  // 只在新增模型且当前无价格时自动填充
  const addedModels = newModels.filter(m => !oldModels.includes(m))
  if (addedModels.length === 0) return

  // 检查是否所有价格字段都为空
  const e = props.entry
  const hasPrice = e.input_price != null || e.output_price != null ||
                   e.cache_write_price != null || e.cache_write_1h_price != null || e.cache_read_price != null
  if (hasPrice) return

  // 查询第一个新增模型的默认价格
  try {
    const result = await channelsAPI.getModelDefaultPricing(addedModels[0])
    if (result.found) {
      emit('update', {
        ...props.entry,
        models: newModels,
        input_price: perTokenToMTok(result.input_price ?? null),
        output_price: perTokenToMTok(result.output_price ?? null),
        cache_write_price: perTokenToMTok(result.cache_write_price ?? null),
        cache_write_1h_price: perTokenToMTok(result.cache_write_1h_price ?? null),
        cache_read_price: perTokenToMTok(result.cache_read_price ?? null),
        image_input_price: perTokenToMTok(result.image_input_price ?? null),
        image_output_price: perTokenToMTok(result.image_output_price ?? null),
        reasoning_effort_multipliers: props.entry.reasoning_effort_multipliers ?? result.reasoning_effort_multipliers ?? null,
      })
    }
  } catch {
    // 查询失败不影响用户操作
  }
}
</script>

<style scoped>
.pricing-default-grid {
  grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
}

.collapsible-content {
  display: grid;
  grid-template-rows: 1fr;
  transition: grid-template-rows 0.25s ease;
}

.collapsible-content--collapsed {
  grid-template-rows: 0fr;
}

.collapsible-inner {
  overflow: hidden;
}
</style>
