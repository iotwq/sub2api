<template>
  <div class="auth-shell">
    <section class="auth-visual" :aria-label="t('auth.access.sceneLabel')">
      <HomeThreeScene class="auth-scene" />
      <div class="auth-visual-mask" aria-hidden="true"></div>
      <div class="auth-grid" aria-hidden="true"></div>

      <router-link to="/home" class="auth-brand" :aria-label="siteName">
        <span class="auth-brand-mark">
          <img :src="siteLogo || '/dragon-logo.png'" alt="" />
        </span>
        <span class="auth-brand-copy">
          <strong>{{ siteName }}</strong>
          <small>{{ siteSubtitle }}</small>
        </span>
      </router-link>

      <div class="auth-visual-copy">
        <p class="auth-eyebrow">
          <span></span>
          {{ t('auth.access.eyebrow') }}
        </p>
        <h1>{{ t('auth.access.title') }}</h1>
        <p>{{ t('auth.access.description') }}</p>

        <ol class="auth-pipeline">
          <li><span>01</span>{{ t('auth.access.plan') }}</li>
          <li><span>02</span>{{ t('auth.access.reason') }}</li>
          <li><span>03</span>{{ t('auth.access.act') }}</li>
          <li><span>04</span>{{ t('auth.access.verify') }}</li>
        </ol>
      </div>

      <div class="auth-visual-status">
        <span><i></i>{{ t('auth.access.live') }}</span>
        <strong>LLM / TOOLS / MEMORY / EVAL</strong>
      </div>
    </section>

    <section class="auth-workspace">
      <div class="auth-toolbar">
        <router-link to="/home" class="auth-back-link">
          <Icon name="arrowLeft" size="sm" />
          {{ t('auth.access.backHome') }}
        </router-link>
        <LocaleSwitcher />
      </div>

      <div class="auth-form-wrap">
        <div class="auth-panel">
          <slot />
        </div>

        <div class="auth-footer-slot">
          <slot name="footer" />
        </div>

        <p class="auth-copyright">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import HomeThreeScene from '@/components/home/HomeThreeScene.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'ChinaAPI')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || t('home.heroSubtitle'))
const currentYear = computed(() => new Date().getFullYear())
</script>

<style scoped>
.auth-shell {
  display: grid;
  min-height: 100svh;
  grid-template-columns: minmax(420px, 0.92fr) minmax(520px, 1.08fr);
  background: #070908;
  isolation: isolate;
}

.auth-shell,
.auth-shell * {
  letter-spacing: 0;
}

.auth-visual {
  position: sticky;
  top: 0;
  height: 100svh;
  min-height: 620px;
  overflow: hidden;
  border-right: 1px solid rgba(216, 255, 237, 0.14);
  background: #070908;
  color: #f4f7f3;
}

.auth-visual::after {
  position: absolute;
  z-index: 3;
  top: -20%;
  right: 0;
  left: 0;
  height: 18%;
  pointer-events: none;
  background: linear-gradient(180deg, transparent, rgba(114, 245, 213, 0.06), transparent);
  content: '';
  animation: auth-scan 8s linear infinite;
}

.auth-scene,
.auth-visual-mask,
.auth-grid {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.auth-scene {
  z-index: 0;
  opacity: 0.92;
}

.auth-visual-mask {
  z-index: 1;
  pointer-events: none;
  background:
    linear-gradient(180deg, #070908 0%, rgba(7, 9, 8, 0.18) 28%, rgba(7, 9, 8, 0.16) 62%, #070908 100%),
    linear-gradient(90deg, rgba(7, 9, 8, 0.22), transparent 56%, rgba(7, 9, 8, 0.5));
}

.auth-grid {
  z-index: 2;
  pointer-events: none;
  opacity: 0.5;
  background-image:
    linear-gradient(rgba(161, 202, 188, 0.09) 1px, transparent 1px),
    linear-gradient(90deg, rgba(161, 202, 188, 0.09) 1px, transparent 1px);
  background-size: 64px 64px;
  mask-image: linear-gradient(to bottom, black, transparent 74%);
}

.auth-brand {
  position: absolute;
  z-index: 5;
  top: 28px;
  left: 32px;
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 11px;
  color: #fff;
  text-decoration: none;
}

.auth-brand-mark {
  display: flex;
  width: 40px;
  height: 40px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid rgba(216, 255, 237, 0.22);
  border-radius: 6px;
  background: #111613;
}

.auth-brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.auth-brand-copy {
  display: grid;
  min-width: 0;
  gap: 2px;
}

.auth-brand-copy strong {
  overflow: hidden;
  max-width: 220px;
  font-size: 15px;
  font-weight: 850;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.auth-brand-copy small {
  overflow: hidden;
  max-width: 250px;
  color: #839089;
  font-size: 9px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.auth-visual-copy {
  position: absolute;
  z-index: 4;
  right: 46px;
  bottom: 142px;
  left: 46px;
  max-width: 560px;
  animation: auth-copy-in 720ms 220ms both cubic-bezier(0.22, 1, 0.36, 1);
}

.auth-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  margin: 0 0 14px;
  color: #72f5d5;
  font-size: 10px;
  font-weight: 900;
}

.auth-eyebrow span,
.auth-visual-status i {
  width: 7px;
  height: 7px;
  flex: 0 0 auto;
  border-radius: 50%;
  background: #72f5d5;
  box-shadow: 0 0 14px rgba(114, 245, 213, 0.9);
  animation: auth-pulse 1.8s ease-in-out infinite;
}

.auth-visual-copy h1 {
  max-width: 520px;
  margin: 0;
  font-size: 46px;
  font-weight: 850;
  line-height: 1.06;
  overflow-wrap: anywhere;
}

.auth-visual-copy > p:last-of-type {
  max-width: 500px;
  margin: 18px 0 0;
  color: #9da9a2;
  font-size: 14px;
  line-height: 1.75;
}

.auth-pipeline {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  margin: 28px 0 0;
  padding: 0;
  border-top: 1px solid rgba(216, 255, 237, 0.18);
  border-bottom: 1px solid rgba(216, 255, 237, 0.18);
  list-style: none;
}

.auth-pipeline li {
  display: grid;
  min-width: 0;
  gap: 7px;
  border-right: 1px solid rgba(216, 255, 237, 0.12);
  color: #cbd5cf;
  font-size: 10px;
  font-weight: 750;
  padding: 12px 10px;
}

.auth-pipeline li:last-child {
  border-right: 0;
}

.auth-pipeline span {
  color: #536059;
  font-size: 8px;
  font-weight: 800;
}

.auth-visual-status {
  position: absolute;
  z-index: 4;
  right: 32px;
  bottom: 30px;
  left: 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  color: #72f5d5;
  font-size: 9px;
  font-weight: 850;
}

.auth-visual-status span {
  display: inline-flex;
  align-items: center;
  gap: 7px;
}

.auth-visual-status strong {
  color: #58645e;
  font-size: 8px;
  font-weight: 800;
}

.auth-workspace {
  position: relative;
  min-width: 0;
  min-height: 100svh;
  background: #eef2ed;
  color: #111613;
  isolation: isolate;
  overflow: hidden;
}

.auth-workspace::before {
  position: absolute;
  pointer-events: none;
  content: '';
}

.auth-workspace::before {
  inset: 0;
  z-index: 0;
  background-image:
    linear-gradient(rgba(35, 140, 115, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(35, 140, 115, 0.045) 1px, transparent 1px);
  background-size: 44px 44px;
  mask-image: linear-gradient(135deg, black 0%, transparent 72%);
}

.auth-toolbar {
  position: absolute;
  z-index: 6;
  top: 22px;
  right: 28px;
  left: 28px;
  display: flex;
  min-height: 38px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  animation: auth-toolbar-in 520ms 80ms both ease-out;
}

.auth-back-link {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: #67726b;
  font-size: 11px;
  font-weight: 750;
  text-decoration: none;
  transition: color 160ms ease;
}

.auth-back-link:hover,
.auth-back-link:focus-visible {
  color: #111613;
  outline: none;
}

.auth-form-wrap {
  position: relative;
  z-index: 1;
  display: flex;
  width: min(540px, calc(100% - 64px));
  min-height: 100svh;
  flex-direction: column;
  justify-content: center;
  margin: 0 auto;
  padding: 96px 0 42px;
}

.auth-panel {
  width: 100%;
  border-top: 1px solid rgba(17, 22, 19, 0.2);
  border-bottom: 1px solid rgba(17, 22, 19, 0.14);
  padding: 34px 0 32px;
  animation: auth-panel-in 620ms 120ms both cubic-bezier(0.22, 1, 0.36, 1);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.32);
}

.auth-footer-slot {
  min-height: 24px;
  margin-top: 20px;
  color: #69746d;
  font-size: 13px;
  text-align: center;
}

.auth-copyright {
  margin: 30px 0 0;
  color: #8b958f;
  font-size: 10px;
  text-align: center;
}

:global(.dark) .auth-workspace {
  background: #0d110f;
  color: #f0f5f1;
}

:global(.dark) .auth-workspace::before {
  background-image:
    linear-gradient(rgba(114, 245, 213, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(114, 245, 213, 0.04) 1px, transparent 1px);
}

:global(.dark) .auth-back-link,
:global(.dark) .auth-footer-slot {
  color: #8f9a93;
}

:global(.dark) .auth-back-link:hover,
:global(.dark) .auth-back-link:focus-visible {
  color: #fff;
}

:global(.dark) .auth-panel {
  border-color: rgba(216, 255, 237, 0.14);
  box-shadow: inset 0 1px 0 rgba(216, 255, 237, 0.04);
}

@keyframes auth-pulse {
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

@keyframes auth-toolbar-in {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes auth-panel-in {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes auth-copy-in {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes auth-scan {
  from { transform: translateY(0); }
  to { transform: translateY(680%); }
}

@media (max-width: 1100px) {
  .auth-shell {
    grid-template-columns: minmax(360px, 0.82fr) minmax(500px, 1.18fr);
  }

  .auth-visual-copy {
    right: 34px;
    left: 34px;
  }

  .auth-visual-copy h1 {
    font-size: 38px;
  }

  .auth-pipeline li {
    padding: 11px 6px;
  }
}

@media (max-width: 900px) {
  .auth-shell {
    display: block;
  }

  .auth-visual {
    position: relative;
    height: 220px;
    min-height: 0;
    border-right: 0;
    border-bottom: 1px solid rgba(216, 255, 237, 0.14);
  }

  .auth-brand {
    top: 18px;
    left: 20px;
  }

  .auth-visual-copy {
    right: 20px;
    bottom: 24px;
    left: 20px;
  }

  .auth-eyebrow {
    margin-bottom: 9px;
    font-size: 8px;
  }

  .auth-visual-copy h1 {
    max-width: 440px;
    font-size: 28px;
  }

  .auth-visual-copy > p:last-of-type,
  .auth-pipeline,
  .auth-visual-status {
    display: none;
  }

  .auth-workspace {
    min-height: calc(100svh - 220px);
  }

  .auth-toolbar {
    top: 16px;
    right: 20px;
    left: 20px;
  }

  .auth-form-wrap {
    width: min(540px, calc(100% - 40px));
    min-height: calc(100svh - 220px);
    padding: 78px 0 34px;
  }
}

@media (max-width: 560px) {
  .auth-visual {
    height: 166px;
  }

  .auth-brand-copy small {
    display: none;
  }

  .auth-brand-mark {
    width: 34px;
    height: 34px;
  }

  .auth-visual-copy {
    bottom: 18px;
  }

  .auth-visual-copy h1 {
    max-width: 320px;
    font-size: 23px;
  }

  .auth-workspace,
  .auth-form-wrap {
    min-height: calc(100svh - 166px);
  }

  .auth-toolbar {
    top: 12px;
  }

  .auth-back-link {
    font-size: 10px;
  }

  .auth-form-wrap {
    width: calc(100% - 36px);
    padding-top: 66px;
  }

  .auth-panel {
    padding: 28px 0 26px;
  }
}

@media (max-height: 760px) and (min-width: 901px) {
  .auth-visual-copy {
    bottom: 104px;
  }

  .auth-visual-copy > p:last-of-type {
    display: none;
  }

  .auth-pipeline {
    margin-top: 20px;
  }

  .auth-form-wrap {
    justify-content: flex-start;
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-eyebrow span,
  .auth-visual-status i,
  .auth-visual-copy,
  .auth-visual::after,
  .auth-toolbar,
  .auth-panel {
    animation: none;
  }
}
</style>
