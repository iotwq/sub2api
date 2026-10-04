import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import AppSidebar from '../AppSidebar.vue'

const mocks = vi.hoisted(() => ({
  auth: { user: { id: 42 }, isAdmin: false, isAuthenticated: true, isSimpleMode: false },
  app: {
    sidebarCollapsed: false, mobileOpen: false, sidebarScrollTop: 0,
    backendModeEnabled: false, cachedPublicSettings: null, publicSettingsLoaded: true,
    siteName: 'Test', siteLogo: '', siteVersion: 'test', setMobileOpen: vi.fn(), toggleSidebar: vi.fn(),
  },
  listMessages: vi.fn(),
  getDirectUnread: vi.fn(),
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => mocks.auth,
  useAppStore: () => mocks.app,
  useOnboardingStore: () => ({ isCurrentStep: () => false }),
  useAdminSettingsStore: () => ({ customMenuItems: [], fetch: vi.fn() }),
}))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/components/common/VersionBadge.vue', () => ({ default: { template: '<span />' } }))
vi.mock('@/utils/featureFlags', () => ({ FeatureFlags: {}, makeSidebarFlag: () => () => true, resolveFeatureFlag: () => true }))
vi.mock('@/composables/useBatchImageAccess', () => ({
  useBatchImageAccess: () => ({ canUseBatchImage: { value: false }, refreshBatchImageAccess: vi.fn() }),
}))
vi.mock('@/composables/useCommunityChatRealtime', () => ({
  useCommunityChatRealtime: () => ({ start: vi.fn(), stop: vi.fn() }),
}))
vi.mock('@/api', () => ({
  communityChatAPI: {
    listMessages: mocks.listMessages,
    getDirectUnread: mocks.getDirectUnread,
    getCommunityChatLastSeenMessageId: () => 0,
  },
}))

let wrapper: VueWrapper | undefined
beforeEach(() => {
  vi.clearAllMocks()
  mocks.auth.isAdmin = false
  mocks.app.sidebarCollapsed = false
  mocks.listMessages.mockResolvedValue({ items: [] })
  mocks.getDirectUnread.mockResolvedValue({ total: 2, conversations: [] })
})
afterEach(() => { wrapper?.unmount(); wrapper = undefined })

async function openSidebar() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
  })
  await router.push('/dashboard')
  await router.isReady()
  wrapper = mount(AppSidebar, { global: { plugins: [router], stubs: { VersionBadge: true, Icon: true } } })
  await flushPromises()
  return router
}

function badgeVisible() {
  return wrapper!.find('a[href="/community-chat"] .sidebar-unread-dot').exists()
}

function seen(scope: 'group' | 'direct') {
  window.dispatchEvent(new CustomEvent('community-chat:seen', { detail: { userId: 42, scope } }))
}

describe('sidebar unread notifications while viewing community chat', () => {
  it.each([
    { admin: false, collapsed: false }, { admin: false, collapsed: true },
    { admin: true, collapsed: false }, { admin: true, collapsed: true },
  ])('keeps private reminders until all conversations are read ($admin/$collapsed)', async ({ admin, collapsed }) => {
    mocks.auth.isAdmin = admin
    mocks.app.sidebarCollapsed = collapsed
    const router = await openSidebar()
    expect(badgeVisible()).toBe(true)
    await router.push('/community-chat')
    await flushPromises()
    expect(badgeVisible()).toBe(true)
    seen('group')
    await flushPromises()
    expect(badgeVisible()).toBe(true)
    mocks.getDirectUnread.mockResolvedValue({ total: 1, conversations: [] })
    seen('direct')
    await flushPromises()
    expect(badgeVisible()).toBe(true)
    mocks.getDirectUnread.mockResolvedValue({ total: 0, conversations: [] })
    seen('direct')
    await flushPromises()
    expect(badgeVisible()).toBe(false)
  })

  it('keeps a group reminder on the chat route until the page reports it read', async () => {
    mocks.getDirectUnread.mockResolvedValue({ total: 0, conversations: [] })
    mocks.listMessages.mockResolvedValue({ items: [{ id: 8, user_id: 7 }] })
    const router = await openSidebar()
    await router.push('/community-chat')
    await flushPromises()
    expect(badgeVisible()).toBe(true)
    seen('group')
    await flushPromises()
    expect(badgeVisible()).toBe(false)
  })
})
