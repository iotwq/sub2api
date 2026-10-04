<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="hasHomeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Compact Home Page -->
  <div
    v-else-if="compactHomeEnabled"
    data-testid="compact-home"
    class="flex min-h-screen flex-col bg-gray-50 text-gray-900 dark:bg-dark-950 dark:text-white"
  >
    <header class="border-b border-gray-200 px-4 py-4 sm:px-6 dark:border-dark-800">
      <nav class="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-3 sm:gap-4">
        <div class="flex min-w-0 flex-1 items-center gap-3">
          <img
            :src="siteLogo || '/logo.svg'"
            alt="Logo"
            class="h-9 w-9 shrink-0 rounded-lg object-contain"
          />
          <span class="min-w-0 truncate text-base font-semibold">{{ siteName }}</span>
        </div>
        <div class="flex max-w-full shrink-0 flex-wrap items-center justify-end gap-2">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <router-link
            v-if="showModelPlazaEntry"
            to="/model-plaza"
            class="flex h-10 shrink-0 items-center gap-1.5 rounded-lg px-2.5 text-sm font-medium text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('nav.modelPlaza')"
          >
            <Icon name="grid" size="md" />
            <span class="hidden sm:inline">{{ t('nav.modelPlaza') }}</span>
          </router-link>
          <button
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex min-h-10 shrink-0 items-center justify-center rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-200"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="flex min-w-0 flex-1 items-center justify-center px-4 py-16 sm:px-6">
      <div class="min-w-0 max-w-2xl text-center">
        <img
          :src="siteLogo || '/logo.svg'"
          alt="Logo"
          class="mx-auto mb-6 h-20 w-20 rounded-2xl object-contain"
        />
        <h1 class="[overflow-wrap:anywhere] text-3xl font-bold md:text-4xl">{{ siteName }}</h1>
        <p class="mt-4 whitespace-pre-wrap [overflow-wrap:anywhere] text-base text-gray-600 dark:text-dark-300">{{ siteSubtitle }}</p>
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="mt-8 inline-flex min-h-10 items-center justify-center rounded-lg bg-primary-600 px-5 py-2.5 text-sm font-medium text-white hover:bg-primary-700"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.login') }}
        </router-link>
      </div>
    </main>

    <footer class="min-w-0 border-t border-gray-200 px-4 py-5 text-center text-sm text-gray-500 [overflow-wrap:anywhere] sm:px-6 dark:border-dark-800 dark:text-dark-400">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>

  <div v-else class="frontier-home">
    <div class="home-grid" aria-hidden="true"></div>

    <header class="home-header">
      <nav class="home-nav">
        <router-link to="/home" class="home-brand" :aria-label="siteName">
          <span class="home-brand-mark">
            <img :src="siteLogo || '/dragon-logo.png'" alt="" />
          </span>
          <span class="home-brand-copy">
            <strong>{{ siteName }}</strong>
          </span>
        </router-link>

        <div class="home-nav-actions">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="home-icon-button"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <!-- Model Plaza Link -->
          <router-link
            v-if="showModelPlazaEntry"
            to="/model-plaza"
            class="inline-flex items-center gap-1.5 rounded-lg p-2 text-sm text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('nav.modelPlaza')"
          >
            <Icon name="grid" size="md" />
            <span class="hidden sm:inline">{{ t('nav.modelPlaza') }}</span>
          </router-link>

          <!-- Theme Toggle -->
          <button
            type="button"
            class="home-icon-button"
            @click="toggleTheme"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="home-account-link"
          >
            <span class="home-account-avatar">
              {{ userInitial }}
            </span>
            <span>{{ t('home.dashboard') }}</span>
            <Icon name="arrowRight" size="xs" />
          </router-link>
          <router-link
            v-else
            to="/login"
            class="home-account-link home-account-link-login"
          >
            {{ t('home.login') }}
            <Icon name="arrowRight" size="xs" />
          </router-link>
        </div>
      </nav>
    </header>

    <main class="home-main">
      <section class="home-hero" :aria-label="t('home.agent.sceneLabel')">
        <HomeThreeScene class="home-agent-scene" />
        <div class="home-scene-mask" aria-hidden="true"></div>

        <div class="home-hero-copy">
          <p class="home-eyebrow">
            <span></span>
            {{ t('home.agent.eyebrow') }}
          </p>
          <h1>{{ siteName }}</h1>
          <p class="home-hero-kicker">{{ t('home.agent.headline') }}</p>
          <p class="home-hero-description">{{ t('home.heroDescription') }}</p>

          <div class="home-hero-actions">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="home-primary-action"
            >
              <Icon name="play" size="sm" />
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
            </router-link>
            <a href="#research" class="home-secondary-action">
              {{ t('home.agent.explore') }}
              <Icon name="arrowDown" size="sm" />
            </a>
          </div>

          <div class="home-domain-list" role="list">
            <span role="listitem">LLM</span>
            <span role="listitem">MULTIMODAL</span>
            <span role="listitem">AGENT</span>
            <span role="listitem">TOOL USE</span>
          </div>
        </div>

        <div class="home-agent-map" aria-hidden="true">
          <div class="home-agent-node home-agent-node-model">
            <span class="home-node-icon"><Icon name="brain" size="sm" /></span>
            <span>
              <strong>{{ t('home.agent.nodes.model') }}</strong>
              <small>{{ t('home.agent.nodes.modelDetail') }}</small>
            </span>
          </div>
          <div class="home-agent-node home-agent-node-tools">
            <span class="home-node-icon"><Icon name="terminal" size="sm" /></span>
            <span>
              <strong>{{ t('home.agent.nodes.tools') }}</strong>
              <small>{{ t('home.agent.nodes.toolsDetail') }}</small>
            </span>
          </div>
          <div class="home-agent-node home-agent-node-memory">
            <span class="home-node-icon"><Icon name="database" size="sm" /></span>
            <span>
              <strong>{{ t('home.agent.nodes.memory') }}</strong>
              <small>{{ t('home.agent.nodes.memoryDetail') }}</small>
            </span>
          </div>
          <div class="home-agent-node home-agent-node-evaluator">
            <span class="home-node-icon"><Icon name="checkCircle" size="sm" /></span>
            <span>
              <strong>{{ t('home.agent.nodes.evaluator') }}</strong>
              <small>{{ t('home.agent.nodes.evaluatorDetail') }}</small>
            </span>
          </div>
          <div class="home-core-status">
            <small>{{ t('home.agent.core') }}</small>
            <strong>{{ t('home.agent.coreStatus') }}</strong>
          </div>
        </div>

        <div class="home-runway">
          <div class="home-trace-heading">
            <span>{{ t('home.agent.traceEyebrow') }}</span>
            <strong>{{ t('home.agent.traceTitle') }}</strong>
          </div>
          <ol class="home-trace-steps">
            <li>
              <span>01</span>
              <strong>{{ t('home.agent.steps.plan') }}</strong>
            </li>
            <li>
              <span>02</span>
              <strong>{{ t('home.agent.steps.reason') }}</strong>
            </li>
            <li>
              <span>03</span>
              <strong>{{ t('home.agent.steps.act') }}</strong>
            </li>
            <li>
              <span>04</span>
              <strong>{{ t('home.agent.steps.verify') }}</strong>
            </li>
          </ol>
          <div class="home-run-metrics">
            <span>
              <small>{{ t('home.agent.metrics.tools') }}</small>
              <strong>12</strong>
            </span>
            <span>
              <small>{{ t('home.agent.metrics.context') }}</small>
              <strong>128K</strong>
            </span>
            <span class="home-run-active">
              <i></i>
              {{ t('home.agent.metrics.active') }}
            </span>
          </div>
        </div>
      </section>

      <section id="research" class="home-research">
        <div class="home-research-inner">
          <div class="home-research-heading">
            <p>{{ t('home.agent.researchEyebrow') }}</p>
            <h2>{{ t('home.agent.researchTitle') }}</h2>
            <span>{{ t('home.agent.researchDescription') }}</span>
          </div>

          <div class="home-research-tracks">
            <article class="home-research-track">
              <span>01</span>
              <Icon name="cpu" size="md" />
              <h3>{{ t('home.agent.tracks.systems') }}</h3>
              <p>{{ t('home.agent.tracks.systemsDetail') }}</p>
            </article>
            <article class="home-research-track">
              <span>02</span>
              <Icon name="sparkles" size="md" />
              <h3>{{ t('home.agent.tracks.multimodal') }}</h3>
              <p>{{ t('home.agent.tracks.multimodalDetail') }}</p>
            </article>
            <article class="home-research-track">
              <span>03</span>
              <Icon name="terminal" size="md" />
              <h3>{{ t('home.agent.tracks.tooling') }}</h3>
              <p>{{ t('home.agent.tracks.toolingDetail') }}</p>
            </article>
            <article class="home-research-track">
              <span>04</span>
              <Icon name="beaker" size="md" />
              <h3>{{ t('home.agent.tracks.evaluation') }}</h3>
              <p>{{ t('home.agent.tracks.evaluationDetail') }}</p>
            </article>
          </div>
        </div>
      </section>
    </main>

    <footer class="home-footer">
      <div class="home-footer-inner">
        <p>
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">
          {{ t('home.docs') }}
        </a>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import HomeThreeScene from '@/components/home/HomeThreeScene.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'ChinaAPI')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const compactHomeEnabled = computed(() => appStore.cachedPublicSettings?.compact_home_enabled === true)
const modelPlazaEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.modelPlaza))

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const modelPlazaRequiresAuth = computed(
  () => appStore.cachedPublicSettings?.model_plaza_require_auth === true,
)
const showModelPlazaEntry = computed(
  () => modelPlazaEnabled.value && (isAuthenticated.value || !modelPlazaRequiresAuth.value),
)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()

  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.frontier-home {
  --home-bg: #070908;
  --home-line: rgba(216, 255, 237, 0.16);
  --home-text: #f4f7f3;
  --home-muted: #a6b0aa;
  --home-mint: #72f5d5;
  --home-lime: #d9ff66;
  --home-coral: #ff786b;
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  background: var(--home-bg);
  color: var(--home-text);
}

.frontier-home,
.frontier-home * {
  letter-spacing: 0;
}

.home-grid {
  position: absolute;
  z-index: 0;
  top: 0;
  right: 0;
  left: 0;
  height: 900px;
  pointer-events: none;
  background-image:
    linear-gradient(rgba(161, 202, 188, 0.07) 1px, transparent 1px),
    linear-gradient(90deg, rgba(161, 202, 188, 0.07) 1px, transparent 1px);
  background-size: 64px 64px;
  mask-image: linear-gradient(to bottom, black 0%, rgba(0, 0, 0, 0.6) 68%, transparent 100%);
}

.home-header {
  position: relative;
  z-index: 20;
  height: 72px;
  border-bottom: 1px solid rgba(216, 255, 237, 0.12);
  background: rgba(7, 9, 8, 0.76);
  backdrop-filter: blur(18px);
  animation: home-header-in 520ms both ease-out;
}

.home-nav {
  display: flex;
  width: min(1440px, 100%);
  height: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  margin: 0 auto;
  padding: 0 32px;
}

.home-brand {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
  color: var(--home-text);
  text-decoration: none;
}

.home-brand-mark {
  display: flex;
  width: 40px;
  height: 40px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid rgba(216, 255, 237, 0.2);
  border-radius: 6px;
  background: #111613;
}

.home-brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.home-brand-copy {
  display: grid;
  min-width: 0;
  gap: 2px;
}

.home-brand-copy strong {
  overflow: hidden;
  color: #fff;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-nav-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 8px;
}

.home-icon-button {
  display: inline-flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border: 1px solid transparent;
  border-radius: 4px;
  color: #98a49e;
  transition: border-color 160ms ease, background 160ms ease, color 160ms ease;
}

.home-icon-button:hover,
.home-icon-button:focus-visible {
  border-color: var(--home-line);
  background: rgba(216, 255, 237, 0.06);
  color: #fff;
  outline: none;
}

.home-account-link {
  display: inline-flex;
  min-height: 36px;
  align-items: center;
  gap: 7px;
  border: 1px solid rgba(114, 245, 213, 0.28);
  border-radius: 4px;
  background: rgba(114, 245, 213, 0.08);
  color: #dbfff6;
  font-size: 12px;
  font-weight: 700;
  padding: 5px 10px 5px 5px;
  text-decoration: none;
  transition: border-color 160ms ease, background 160ms ease, transform 160ms ease, box-shadow 160ms ease;
}

.home-account-link:hover,
.home-account-link:focus-visible {
  border-color: rgba(114, 245, 213, 0.54);
  background: rgba(114, 245, 213, 0.14);
  box-shadow: 0 8px 18px rgba(114, 245, 213, 0.08);
  transform: translateY(-1px);
  outline: none;
}

.home-account-link-login {
  padding-left: 12px;
}

.home-account-avatar {
  display: flex;
  width: 24px;
  height: 24px;
  align-items: center;
  justify-content: center;
  border-radius: 3px;
  background: var(--home-mint);
  color: #07100d;
  font-size: 10px;
  font-weight: 900;
}

.home-main {
  position: relative;
  z-index: 1;
}

.home-hero {
  position: relative;
  width: min(1600px, 100%);
  height: clamp(500px, calc(100svh - 104px), 760px);
  margin: 0 auto;
  overflow: hidden;
  border-right: 1px solid rgba(216, 255, 237, 0.08);
  border-bottom: 1px solid rgba(216, 255, 237, 0.12);
  border-left: 1px solid rgba(216, 255, 237, 0.08);
  animation: home-hero-in 700ms 80ms both cubic-bezier(0.22, 1, 0.36, 1);
}

.home-hero::after {
  position: absolute;
  z-index: 2;
  top: -22%;
  right: 0;
  left: 0;
  height: 18%;
  pointer-events: none;
  background: linear-gradient(180deg, transparent, rgba(114, 245, 213, 0.055), transparent);
  content: '';
  animation: home-scan 10s linear infinite;
}

.home-agent-scene {
  position: absolute;
  z-index: 0;
  inset: 0;
  width: 100%;
  height: 100%;
}

.home-scene-mask {
  position: absolute;
  z-index: 1;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(90deg, #070908 0%, rgba(7, 9, 8, 0.96) 24%, rgba(7, 9, 8, 0.44) 47%, transparent 68%),
    linear-gradient(0deg, #070908 0%, rgba(7, 9, 8, 0.86) 9%, transparent 34%);
}

.home-hero-copy {
  position: absolute;
  z-index: 4;
  top: 9%;
  left: max(32px, calc((100% - 1440px) / 2 + 44px));
  width: min(510px, 42%);
}

.home-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  margin: 0 0 18px;
  color: var(--home-mint);
  font-size: 11px;
  font-weight: 800;
  line-height: 1;
  text-transform: uppercase;
}

.home-eyebrow span,
.home-run-active i {
  width: 7px;
  height: 7px;
  flex: 0 0 auto;
  border-radius: 50%;
  background: var(--home-mint);
  box-shadow: 0 0 14px rgba(114, 245, 213, 0.92);
  animation: home-pulse 1.8s ease-in-out infinite;
}

.home-hero-copy h1 {
  margin: 0;
  color: #fff;
  font-size: 88px;
  font-weight: 900;
  line-height: 0.92;
  overflow-wrap: anywhere;
}

.home-hero-kicker {
  max-width: 480px;
  margin: 22px 0 0;
  color: #eaf0ec;
  font-size: 28px;
  font-weight: 650;
  line-height: 1.28;
}

.home-hero-description {
  max-width: 470px;
  margin: 14px 0 0;
  color: var(--home-muted);
  font-size: 15px;
  line-height: 1.72;
}

.home-hero-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-top: 24px;
}

.home-primary-action,
.home-secondary-action {
  display: inline-flex;
  min-height: 42px;
  align-items: center;
  justify-content: center;
  gap: 9px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 800;
  position: relative;
  overflow: hidden;
  padding: 0 17px;
  text-decoration: none;
  transition: transform 160ms ease, border-color 160ms ease, background 160ms ease;
}

.home-primary-action {
  border: 1px solid var(--home-mint);
  background: var(--home-mint);
  color: #06110d;
}

.home-primary-action::after {
  position: absolute;
  top: 0;
  bottom: 0;
  left: -45%;
  width: 28%;
  background: rgba(255, 255, 255, 0.42);
  content: '';
  transform: skewX(-18deg);
  transition: left 420ms ease;
}

.home-primary-action > * {
  position: relative;
  z-index: 1;
}

.home-secondary-action {
  border: 1px solid rgba(216, 255, 237, 0.2);
  background: rgba(216, 255, 237, 0.04);
  color: #e7eee9;
}

.home-primary-action:hover,
.home-primary-action:focus-visible,
.home-secondary-action:hover,
.home-secondary-action:focus-visible {
  transform: translateY(-2px);
  outline: none;
}

.home-primary-action:hover::after,
.home-primary-action:focus-visible::after {
  left: 125%;
}

.home-secondary-action:hover,
.home-secondary-action:focus-visible {
  border-color: rgba(216, 255, 237, 0.42);
  background: rgba(216, 255, 237, 0.09);
}

.home-domain-list {
  display: flex;
  flex-wrap: wrap;
  gap: 7px 18px;
  margin-top: 23px;
  color: #68736d;
  font-size: 10px;
  font-weight: 800;
}

.home-domain-list span {
  position: relative;
}

.home-domain-list span:not(:last-child)::after {
  position: absolute;
  right: -11px;
  color: #33403a;
  content: '/';
}

.home-agent-map {
  position: absolute;
  z-index: 3;
  top: 7%;
  right: 2%;
  bottom: 128px;
  left: 43%;
  pointer-events: none;
  animation: home-map-in 760ms 260ms both ease-out;
}

.home-agent-node {
  position: absolute;
  display: flex;
  min-width: 154px;
  align-items: center;
  gap: 9px;
  border-left: 2px solid var(--home-mint);
  border-radius: 0 4px 4px 0;
  background: rgba(7, 12, 10, 0.64);
  color: #dfe9e4;
  padding: 8px 10px;
  backdrop-filter: blur(10px);
  animation: home-node-float 5.2s ease-in-out infinite;
}

.home-agent-node::after {
  position: absolute;
  top: 50%;
  width: 30px;
  height: 1px;
  background: rgba(114, 245, 213, 0.32);
  content: '';
}

.home-agent-node > span:last-child {
  display: grid;
  gap: 2px;
}

.home-agent-node strong {
  font-size: 12px;
  font-weight: 800;
  line-height: 1.2;
}

.home-agent-node small {
  color: #85928b;
  font-size: 9px;
  line-height: 1.2;
  white-space: nowrap;
}

.home-node-icon {
  display: flex;
  width: 28px;
  height: 28px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(114, 245, 213, 0.3);
  border-radius: 3px;
  color: var(--home-mint);
}

.home-agent-node-model {
  top: 5%;
  right: 10%;
  border-left-color: var(--home-lime);
  animation-delay: -1.2s;
}

.home-agent-node-model::after,
.home-agent-node-tools::after {
  right: 100%;
}

.home-agent-node-model .home-node-icon {
  border-color: rgba(217, 255, 102, 0.35);
  color: var(--home-lime);
}

.home-agent-node-tools {
  top: 46%;
  right: 1%;
  border-left-color: var(--home-coral);
  animation-delay: -2.4s;
}

.home-agent-node-tools .home-node-icon {
  border-color: rgba(255, 120, 107, 0.4);
  color: var(--home-coral);
}

.home-agent-node-memory {
  bottom: 2%;
  left: 31%;
  border-left-color: #9baeff;
  animation-delay: -3.1s;
}

.home-agent-node-memory::after,
.home-agent-node-evaluator::after {
  left: 100%;
}

.home-agent-node-memory .home-node-icon {
  border-color: rgba(155, 174, 255, 0.4);
  color: #aab9ff;
}

.home-agent-node-evaluator {
  top: 28%;
  left: 2%;
  border-left-color: #fff;
  animation-delay: -4.2s;
}

.home-agent-node-evaluator .home-node-icon {
  border-color: rgba(255, 255, 255, 0.3);
  color: #fff;
}

.home-core-status {
  position: absolute;
  top: 47%;
  left: 55%;
  display: grid;
  width: 150px;
  gap: 3px;
  color: #d8fff5;
  text-align: center;
  transform: translate(-50%, -50%);
}

.home-core-status::before,
.home-core-status::after {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 190px;
  aspect-ratio: 1;
  border: 1px solid rgba(114, 245, 213, 0.18);
  border-radius: 50%;
  content: '';
  pointer-events: none;
  transform: translate(-50%, -50%);
  animation: home-core-ring 5s ease-in-out infinite;
}

.home-core-status::after {
  width: 238px;
  border-color: rgba(217, 255, 102, 0.1);
  animation-delay: -2.2s;
}

.home-core-status small {
  color: var(--home-mint);
  font-size: 9px;
  font-weight: 900;
}

.home-core-status strong {
  font-size: 11px;
  font-weight: 700;
}

.home-runway {
  position: absolute;
  z-index: 5;
  right: max(28px, calc((100% - 1440px) / 2 + 36px));
  bottom: 20px;
  left: max(28px, calc((100% - 1440px) / 2 + 36px));
  display: grid;
  min-height: 94px;
  grid-template-columns: minmax(180px, 0.78fr) minmax(420px, 2fr) minmax(210px, 0.8fr);
  align-items: stretch;
  border-top: 1px solid rgba(216, 255, 237, 0.2);
  border-bottom: 1px solid rgba(216, 255, 237, 0.2);
  background: rgba(7, 11, 9, 0.72);
  backdrop-filter: blur(16px);
  animation: home-runway-in 620ms 420ms both ease-out;
}

.home-trace-heading {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
  border-right: 1px solid rgba(216, 255, 237, 0.12);
  padding: 14px 18px;
}

.home-trace-heading span {
  color: var(--home-mint);
  font-size: 9px;
  font-weight: 900;
}

.home-trace-heading strong {
  max-width: 180px;
  color: #f3f7f4;
  font-size: 13px;
  font-weight: 750;
  line-height: 1.35;
}

.home-trace-steps {
  display: grid;
  min-width: 0;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  align-items: stretch;
  margin: 0;
  padding: 0;
  list-style: none;
}

.home-trace-steps li {
  position: relative;
  display: flex;
  min-width: 0;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
  border-right: 1px solid rgba(216, 255, 237, 0.1);
  padding: 14px 16px;
}

.home-trace-steps li::after {
  position: absolute;
  right: 14px;
  bottom: 12px;
  left: 14px;
  height: 2px;
  background: var(--home-mint);
  content: '';
  transform: scaleX(0);
  transform-origin: left;
  animation: home-trace 4.8s ease-in-out infinite;
}

.home-trace-steps li:nth-child(2)::after {
  animation-delay: 0.45s;
}

.home-trace-steps li:nth-child(3)::after {
  background: var(--home-coral);
  animation-delay: 0.9s;
}

.home-trace-steps li:nth-child(4)::after {
  background: var(--home-lime);
  animation-delay: 1.35s;
}

.home-trace-steps span {
  color: #657169;
  font-size: 9px;
  font-weight: 800;
}

.home-trace-steps strong {
  overflow: hidden;
  color: #cdd7d1;
  font-size: 11px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-run-metrics {
  display: grid;
  grid-template-columns: 1fr 1fr;
  align-items: center;
  padding: 12px 16px;
}

.home-run-metrics > span:not(.home-run-active) {
  display: grid;
  gap: 2px;
}

.home-run-metrics small {
  color: #68746d;
  font-size: 8px;
  font-weight: 800;
}

.home-run-metrics strong {
  color: #fff;
  font-size: 17px;
  font-weight: 800;
}

.home-run-active {
  display: inline-flex;
  grid-column: 1 / -1;
  align-items: center;
  gap: 7px;
  color: var(--home-mint);
  font-size: 9px;
  font-weight: 900;
}

.home-research {
  position: relative;
  background: #eef2ed;
  color: #111613;
}

.home-research-inner {
  width: min(1440px, 100%);
  margin: 0 auto;
  padding: 86px 32px 96px;
}

.home-research-heading {
  display: grid;
  grid-template-columns: minmax(160px, 0.55fr) minmax(360px, 1.3fr) minmax(260px, 0.9fr);
  align-items: end;
  gap: 36px;
  padding-bottom: 40px;
  border-bottom: 1px solid rgba(17, 22, 19, 0.2);
  animation: home-content-in 620ms 120ms both ease-out;
}

.home-research-heading p {
  margin: 0;
  color: #238c73;
  font-size: 10px;
  font-weight: 900;
}

.home-research-heading h2 {
  margin: 0;
  font-size: 44px;
  font-weight: 850;
  line-height: 1.08;
}

.home-research-heading span {
  color: #566159;
  font-size: 14px;
  line-height: 1.7;
}

.home-research-tracks {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.home-research-track {
  position: relative;
  min-width: 0;
  min-height: 270px;
  padding: 28px 26px 20px;
  border-right: 1px solid rgba(17, 22, 19, 0.16);
  transition: background 180ms ease, transform 180ms ease;
  animation: home-content-in 620ms both ease-out;
}

.home-research-track:nth-child(1) { animation-delay: 180ms; }
.home-research-track:nth-child(2) { animation-delay: 240ms; }
.home-research-track:nth-child(3) { animation-delay: 300ms; }
.home-research-track:nth-child(4) { animation-delay: 360ms; }

.home-research-track:hover {
  background: rgba(35, 140, 115, 0.08);
  transform: translateY(-3px);
}

.home-research-track::before {
  position: absolute;
  top: 0;
  right: 0;
  left: 0;
  height: 2px;
  background: #238c73;
  content: '';
  opacity: 0.7;
  transform: scaleX(0.18);
  transform-origin: left;
  transition: transform 220ms ease;
}

.home-research-track:nth-child(2)::before { background: #ca4a40; }
.home-research-track:nth-child(3)::before { background: #5168bf; }
.home-research-track:nth-child(4)::before { background: #647700; }

.home-research-track:hover::before {
  transform: scaleX(1);
}

.home-research-track > svg {
  transition: transform 220ms ease;
}

.home-research-track:hover > svg {
  transform: translateY(-2px) scale(1.08) rotate(-4deg);
}

.home-research-track:first-child {
  border-left: 1px solid rgba(17, 22, 19, 0.16);
}

.home-research-track > span {
  display: block;
  margin-bottom: 54px;
  color: #7a847e;
  font-size: 10px;
  font-weight: 800;
}

.home-research-track > svg {
  color: #238c73;
}

.home-research-track:nth-child(2) > svg {
  color: #ca4a40;
}

.home-research-track:nth-child(3) > svg {
  color: #5168bf;
}

.home-research-track:nth-child(4) > svg {
  color: #647700;
}

.home-research-track h3 {
  margin: 19px 0 9px;
  font-size: 19px;
  font-weight: 800;
}

.home-research-track p {
  max-width: 240px;
  margin: 0;
  color: #647068;
  font-size: 13px;
  line-height: 1.65;
}

:global(.dark) .home-research {
  background: #0d110f;
  color: #eef4f0;
}

:global(.dark) .home-research-heading,
:global(.dark) .home-research-track,
:global(.dark) .home-research-track:first-child {
  border-color: rgba(216, 255, 237, 0.15);
}

:global(.dark) .home-research-heading span,
:global(.dark) .home-research-track p {
  color: #929e97;
}

.home-footer {
  border-top: 1px solid rgba(216, 255, 237, 0.1);
  background: #070908;
  color: #68746d;
}

.home-footer-inner {
  display: flex;
  width: min(1440px, 100%);
  min-height: 76px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin: 0 auto;
  padding: 18px 32px;
  font-size: 12px;
}

.home-footer-inner p {
  margin: 0;
}

.home-footer-inner a {
  color: #98a49e;
  text-decoration: none;
}

.home-footer-inner a:hover {
  color: #fff;
}

@keyframes home-pulse {
  0%,
  100% {
    opacity: 0.45;
    transform: scale(0.86);
  }
  50% {
    opacity: 1;
    transform: scale(1.14);
  }
}

@keyframes home-header-in {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes home-hero-in {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes home-map-in {
  from { opacity: 0; transform: scale(0.98); }
  to { opacity: 1; transform: scale(1); }
}

@keyframes home-runway-in {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes home-content-in {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes home-core-ring {
  0%, 100% { opacity: 0.25; transform: translate(-50%, -50%) scale(0.96); }
  50% { opacity: 0.75; transform: translate(-50%, -50%) scale(1.02); }
}

@keyframes home-node-float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-5px);
  }
}

@keyframes home-trace {
  0%,
  12% {
    opacity: 0;
    transform: scaleX(0);
  }
  42%,
  70% {
    opacity: 1;
    transform: scaleX(1);
  }
  100% {
    opacity: 0;
    transform: scaleX(1);
  }
}

@media (max-width: 1200px) {
  .home-hero-copy {
    left: 36px;
    width: 42%;
  }
  .home-hero-copy h1 { font-size: 70px; }
  .home-hero-kicker { font-size: 24px; }
  .home-agent-map { left: 41%; }
  .home-runway {
    right: 28px;
    left: 28px;
    grid-template-columns: minmax(160px, 0.7fr) minmax(380px, 2fr) minmax(170px, 0.7fr);
  }
}

@media (max-width: 960px) {
  .home-hero-copy { top: 7%; width: 48%; }
  .home-hero-copy h1 { font-size: 60px; }
  .home-hero-kicker { font-size: 22px; }
  .home-hero-description { font-size: 14px; }
  .home-agent-map { right: 1%; left: 49%; }
  .home-agent-node { min-width: 132px; }
  .home-agent-node small { display: none; }
  .home-agent-node-evaluator { display: none; }
  .home-runway { grid-template-columns: 150px minmax(360px, 1fr); }
  .home-run-metrics { display: none; }
  .home-research-heading { grid-template-columns: 0.5fr 1.5fr; }
  .home-research-heading span { grid-column: 2; }
  .home-research-tracks { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .home-research-track:nth-child(2) { border-right: 0; }
  .home-research-track:nth-child(n + 3) { border-top: 1px solid rgba(17, 22, 19, 0.16); }
}

@media (max-width: 767px) {
  .home-header { height: 64px; }
  .home-nav { gap: 10px; padding: 0 16px; }
  .home-brand-mark { width: 34px; height: 34px; }
  .home-nav-actions { gap: 4px; }
  .home-icon-button { width: 34px; height: 34px; }
  .home-account-link { min-height: 34px; }
  .home-account-avatar,
  .home-account-link > span:not(.home-account-avatar) { display: none; }
  .home-account-link-login { padding: 5px 9px; }
  .home-hero {
    height: clamp(470px, calc(100svh - 96px), 640px);
    border-right: 0;
    border-left: 0;
  }
  .home-agent-scene { opacity: 0.82; }
  .home-scene-mask {
    background:
      linear-gradient(180deg, #070908 0%, rgba(7, 9, 8, 0.96) 30%, rgba(7, 9, 8, 0.32) 52%, transparent 70%),
      linear-gradient(0deg, #070908 0%, rgba(7, 9, 8, 0.88) 13%, transparent 33%);
  }
  .home-hero-copy { top: 20px; left: 18px; width: min(430px, calc(100% - 36px)); }
  .home-eyebrow { margin-bottom: 12px; font-size: 9px; }
  .home-hero-copy h1 { font-size: 50px; line-height: 0.94; }
  .home-hero-kicker { max-width: 330px; margin-top: 13px; font-size: 21px; line-height: 1.25; }
  .home-hero-description { max-width: 360px; margin-top: 9px; font-size: 13px; line-height: 1.5; }
  .home-hero-actions { gap: 7px; margin-top: 14px; }
  .home-primary-action,
  .home-secondary-action { min-height: 38px; font-size: 11px; padding: 0 12px; }
  .home-domain-list { display: none; }
  .home-agent-map { top: 275px; right: 10px; bottom: 66px; left: 10px; }
  .home-agent-node { min-width: 0; gap: 6px; padding: 5px 7px; }
  .home-agent-node::after,
  .home-agent-node small { display: none; }
  .home-agent-node strong { font-size: 9px; }
  .home-node-icon { width: 24px; height: 24px; }
  .home-agent-node-model { top: 2%; right: 2%; }
  .home-agent-node-tools { top: 48%; right: 0; }
  .home-agent-node-memory { bottom: 2%; left: 33%; }
  .home-agent-node-evaluator { display: flex; top: 44%; left: 0; }
  .home-core-status { top: 48%; left: 53%; }
  .home-core-status::before { width: 130px; }
  .home-core-status::after { width: 164px; }
  .home-core-status small { font-size: 8px; }
  .home-core-status strong { font-size: 9px; }
  .home-runway { right: 12px; bottom: 8px; left: 12px; min-height: 54px; grid-template-columns: 74px 1fr; }
  .home-trace-heading { gap: 3px; padding: 7px; }
  .home-trace-heading span { font-size: 7px; }
  .home-trace-heading strong { font-size: 8px; line-height: 1.25; }
  .home-trace-steps li { gap: 3px; padding: 7px 5px; }
  .home-trace-steps span { font-size: 7px; }
  .home-trace-steps strong { font-size: 8px; }
  .home-trace-steps li::after { right: 5px; bottom: 6px; left: 5px; }
  .home-research-inner { padding: 58px 18px 68px; }
  .home-research-heading { display: block; padding-bottom: 28px; }
  .home-research-heading h2 { margin-top: 16px; font-size: 32px; line-height: 1.12; }
  .home-research-heading span { display: block; margin-top: 16px; font-size: 13px; }
  .home-research-tracks { grid-template-columns: 1fr; }
  .home-research-track,
  .home-research-track:first-child,
  .home-research-track:nth-child(2),
  .home-research-track:nth-child(n + 3) {
    min-height: 0;
    padding: 24px 8px 26px;
    border-top: 0;
    border-right: 0;
    border-bottom: 1px solid rgba(17, 22, 19, 0.16);
    border-left: 0;
  }
  .home-research-track > span { margin-bottom: 24px; }
  .home-research-track p { max-width: none; }
  .home-footer-inner { min-height: 68px; padding: 16px 18px; }
}

@media (max-width: 430px) {
  .home-brand-copy { display: none; }
  .home-agent-node-evaluator { display: none; }
  .home-hero-copy h1 { font-size: 44px; }
  .home-hero-kicker { max-width: 290px; font-size: 19px; }
  .home-secondary-action { padding: 0 10px; }
}

@media (max-width: 767px) and (max-height: 620px) {
  .home-hero-description { display: none; }
  .home-agent-map { top: 205px; }
}

@media (prefers-reduced-motion: reduce) {
  .home-eyebrow span,
  .home-run-active i,
  .home-agent-node,
  .home-trace-steps li::after,
  .home-hero::after,
  .home-core-status::before,
  .home-core-status::after,
  .home-header,
  .home-hero,
  .home-agent-map,
  .home-runway,
  .home-research-heading,
  .home-research-track { animation: none; }
  .home-primary-action,
  .home-secondary-action,
  .home-research-track,
  .home-research-track::before,
  .home-research-track > svg { transition: none; }
}
</style>
