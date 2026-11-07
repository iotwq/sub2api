import { posix } from 'node:path'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import en from '../locales/en'
import zh from '../locales/zh'
import enUser from '../locales/en/user'
import zhUser from '../locales/zh/user'

function leaves(value: unknown, prefix = '', result: Record<string, string> = {}) {
  if (typeof value === 'string') result[prefix] = value
  else if (value && typeof value === 'object') {
    for (const [key, child] of Object.entries(value)) leaves(child, prefix ? `${prefix}.${key}` : key, result)
  }
  return result
}

// Follow user/public views and their shared components, including dynamic imports.
// Shared components may live under components/admin, so folder exclusion is unsafe.
function userAdminReferences(): string[] {
  const raw = import.meta.glob('../../**/*.{ts,vue}', { query: '?raw', import: 'default', eager: true }) as Record<string, string>
  const files = Object.fromEntries(Object.entries(raw).map(([file, source]) => [file.slice(6), source]))
  const seen = new Set<string>()
  const references = new Set<string>()
  function visit(file: string) {
    if (seen.has(file) || file.startsWith('i18n/') || file.startsWith('router/') || file.includes('/__tests__/')) return
    const source = files[file]
    if (!source) return
    seen.add(file)
    for (const match of source.matchAll(/['"`]((?:admin\.|adminAccessKeys)[^'"`\n]*)['"`]/g)) {
      references.add(match[1].split('${')[0])
    }
    for (const match of source.matchAll(/(?:from\s*|import\s*\()(['"])([^'"]+)\1/g)) {
      const ref = match[2]
      if (!ref.startsWith('@/') && !ref.startsWith('.')) continue
      const path = ref.startsWith('@/') ? ref.slice(2) : posix.join(posix.dirname(file), ref)
      const resolved = [path, `${path}.ts`, `${path}.vue`, `${path}/index.ts`].find(candidate => files[candidate])
      if (resolved) visit(resolved)
    }
  }
  for (const file of Object.keys(files)) {
    if (file.startsWith('views/') && !file.startsWith('views/admin/') && file.endsWith('.vue')) visit(file)
  }
  visit('App.vue')
  return [...references]
}

describe('split locale coverage', () => {
  it.each([{ name: 'en', full: en, user: enUser }, { name: 'zh', full: zh, user: zhUser }])('preserves every shared admin message used by $name user pages', ({ full, user }) => {
    const all = leaves(full)
    const initial = leaves(user)
    expect(initial['mcp.title']).toBe(all['mcp.title'])
    expect(initial['mcp.consentTitle']).toBe(all['mcp.consentTitle'])
    const prefixes = userAdminReferences()
    expect(prefixes.length).toBeGreaterThan(50)
    for (const prefix of prefixes) {
      const keys = Object.keys(all).filter(key => key.startsWith(prefix))
      expect(keys.length, `Unknown translation prefix ${prefix}`).toBeGreaterThan(0)
      for (const key of keys) expect(initial[key], key).toBe(all[key])
    }
    expect(initial['admin.accounts.title']).toBeUndefined()
    expect(initial['admin.settings.title']).toBeUndefined()
    expect(initial['admin.ops.title']).toBeUndefined()
  })
})

describe('locale runtime loading', () => {
  beforeEach(() => {
    vi.resetModules()
    localStorage.setItem('chinaapi_locale', 'zh')
  })

  it('loads user messages first and preserves them when admin messages arrive', async () => {
    const { i18n, initI18n, loadAdminLocaleMessages } = await import('../index')
    await initI18n()
    expect(i18n.global.te('admin.usage.group')).toBe(true)
    expect(i18n.global.te('admin.accounts.title')).toBe(false)
    const userTitle = leaves(i18n.global.getLocaleMessage('zh'))['dashboard.title']
    const merge = vi.spyOn(i18n.global, 'mergeLocaleMessage')
    await Promise.all([loadAdminLocaleMessages(), loadAdminLocaleMessages()])
    expect(merge).toHaveBeenCalledTimes(1)
    const messages = leaves(i18n.global.getLocaleMessage('zh'))
    expect(messages['admin.accounts.title']).toBe(zh.admin.accounts.title)
    expect(messages['dashboard.title']).toBe(userTitle)
    merge.mockRestore()
  })

  it('keeps administrator language switching complete', async () => {
    const { i18n, initI18n, loadAdminLocaleMessages, setLocale } = await import('../index')
    await initI18n()
    await loadAdminLocaleMessages()
    await setLocale('en')
    const messages = leaves(i18n.global.getLocaleMessage('en'))
    expect(messages['admin.accounts.title']).toBe(en.admin.accounts.title)
    expect(messages['admin.settings.title']).toBe(en.admin.settings.title)
    expect(messages['dashboard.title']).toBe(en.dashboard.title)
    expect(document.documentElement.lang).toBe('en')
    await setLocale('zh')
    expect(i18n.global.locale.value).toBe('zh')
    expect(leaves(i18n.global.getLocaleMessage('zh'))['admin.accounts.title']).toBe(zh.admin.accounts.title)
  })

  it('does not pull admin-only translations when ordinary users switch language', async () => {
    const { i18n, initI18n, setLocale } = await import('../index')
    await initI18n()
    await setLocale('en')
    expect(i18n.global.locale.value).toBe('en')
    expect(leaves(i18n.global.getLocaleMessage('en'))['dashboard.title']).toBe(en.dashboard.title)
    expect(i18n.global.te('admin.accounts.title')).toBe(false)
  })
})
