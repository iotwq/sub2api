import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { describe, it, expect, vi } from 'vitest'
import type { Account } from '@/types'
import CodexSignalBadge from '../CodexSignalBadge.vue'
import zh from '@/i18n/locales/zh'

vi.unmock('vue-i18n')

const account = { id: 41, platform: 'openai', type: 'oauth' } as Account
function render(overrides: Partial<Account> = {}) {
  // Vitest uses the runtime-only i18n build; provide precompiled message functions.
  const messages = Object.fromEntries(Object.entries(zh.admin.accounts.openai)
    .filter(([key]) => key.startsWith('codexSignal'))
    .map(([key, value]) => [key, (ctx: { named: (key: string) => unknown }) =>
      String(value).replace(/\{(\w+)\}/g, (_, name: string) => String(ctx.named(name)))]))
  return mount(CodexSignalBadge, {
    props: { account: { ...account, ...overrides } },
    global: { plugins: [createI18n({ legacy: false, locale: 'zh', messages: { zh: { admin: { accounts: { openai: messages } } } } })] },
  })
}

describe('CodexSignalBadge', () => {
  it('distinguishes unknown, detected and not detected without claiming healthy', async () => {
    const wrapper = render()
    expect(wrapper.text()).toContain('尚未检测')
    expect(wrapper.attributes('title')).toContain('不会自动停用账号')
    await wrapper.setProps({ account: { ...account, codex_signal: { length: 312, observed_at: '2026-09-24T00:00:00Z', last_312_at: '2026-09-24T00:00:00Z' } } })
    expect(wrapper.text()).toContain('疑似降智：收到 312 信号')
    expect(wrapper.find('span').classes()).toContain('text-red-700')
    expect(wrapper.attributes('title')).toContain('最近一次 312')
    await wrapper.setProps({ account: { ...account, codex_signal: { length: 332, observed_at: '2026-09-24T01:00:00Z', last_312_at: '2026-09-24T00:00:00Z' } } })
    expect(wrapper.text()).toContain('最近响应未发现 312')
    expect(wrapper.text()).not.toContain('正常')
    expect(wrapper.attributes('title')).toContain('最近一次 312')
    expect(wrapper.find('span').classes()).not.toContain('text-red-700')
  })

  it.each([{ type: 'apikey' }, { platform: 'anthropic' }, { parent_account_id: 1 }])('does not label unrelated accounts %o', (overrides) => {
    expect(render(overrides as Partial<Account>).find('[data-testid="codex-signal"]').exists()).toBe(false)
  })
})
