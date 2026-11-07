import { createI18n } from 'vue-i18n'

type LocaleCode = 'en' | 'zh'

type LocaleMessages = Record<string, any>

const LOCALE_KEY = 'chinaapi_locale'
const DEFAULT_LOCALE: LocaleCode = 'en'
export const LOCALE_CHANGED_EVENT = 'chinaapi:locale-changed'

const localeLoaders: Record<LocaleCode, () => Promise<{ default: LocaleMessages }>> = {
  en: () => import('./locales/en/user'),
  zh: () => import('./locales/zh/user')
}

const adminLocaleLoaders: Record<LocaleCode, () => Promise<{ default: LocaleMessages }>> = {
  en: () => import('./locales/en/adminMessages'),
  zh: () => import('./locales/zh/adminMessages')
}

function isLocaleCode(value: string): value is LocaleCode {
  return value === 'en' || value === 'zh'
}

function getDefaultLocale(): LocaleCode {
  const saved = localStorage.getItem(LOCALE_KEY)
  if (saved && isLocaleCode(saved)) {
    return saved
  }

  const browserLang = navigator.language.toLowerCase()
  if (browserLang.startsWith('zh')) {
    return 'zh'
  }

  return DEFAULT_LOCALE
}

export const i18n = createI18n({
  legacy: false,
  locale: getDefaultLocale(),
  fallbackLocale: DEFAULT_LOCALE,
  messages: {},
  // 禁用 HTML 消息警告 - 引导步骤使用富文本内容（driver.js 支持 HTML）
  // 这些内容是内部定义的，不存在 XSS 风险
  warnHtmlMessage: false
})

const localeRequests = new Map<LocaleCode, Promise<void>>()
const adminLocaleRequests = new Map<LocaleCode, Promise<void>>()
let adminMessagesRequired = false

export async function loadLocaleMessages(locale: LocaleCode): Promise<void> {
  if (!localeRequests.has(locale)) {
    const request = localeLoaders[locale]().then(module => {
      i18n.global.setLocaleMessage(locale, module.default)
    }).catch(error => {
      localeRequests.delete(locale)
      throw error
    })
    localeRequests.set(locale, request)
  }
  await localeRequests.get(locale)
}

export async function loadAdminLocaleMessages(locale: LocaleCode = getLocale()): Promise<void> {
  adminMessagesRequired = true
  await loadLocaleMessages(locale)
  if (!adminLocaleRequests.has(locale)) {
    const request = adminLocaleLoaders[locale]().then(module => {
      i18n.global.mergeLocaleMessage(locale, module.default)
    }).catch(error => {
      adminLocaleRequests.delete(locale)
      throw error
    })
    adminLocaleRequests.set(locale, request)
  }
  await adminLocaleRequests.get(locale)
}

export async function initI18n(): Promise<void> {
  const current = getLocale()
  await loadLocaleMessages(current)
  document.documentElement.setAttribute('lang', current)
}

export async function setLocale(locale: string): Promise<void> {
  if (!isLocaleCode(locale)) {
    return
  }

  await loadLocaleMessages(locale)
  if (adminMessagesRequired) await loadAdminLocaleMessages(locale)
  i18n.global.locale.value = locale
  localStorage.setItem(LOCALE_KEY, locale)
  document.documentElement.setAttribute('lang', locale)
  window.dispatchEvent(new CustomEvent(LOCALE_CHANGED_EVENT, { detail: { locale } }))
}

export function getLocale(): LocaleCode {
  const current = i18n.global.locale.value
  return isLocaleCode(current) ? current : DEFAULT_LOCALE
}

export const availableLocales = [
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'zh', name: '中文', flag: '🇨🇳' }
] as const

export default i18n
