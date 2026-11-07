import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'
import DashboardView from '../DashboardView.vue'

const mocks = vi.hoisted(() => ({ refreshUser: vi.fn(), getDashboardStats: vi.fn() }))
const auth = reactive({ user: { balance: 10 }, isSimpleMode: false, refreshUser: mocks.refreshUser })
vi.mock('@/stores/auth', () => ({ useAuthStore: () => auth }))
vi.mock('@/api/usage', () => ({ usageAPI: {
  getDashboardStats: mocks.getDashboardStats,
  getDashboardTrend: async () => ({ trend: [] }),
  getDashboardModels: async () => ({ models: [] }),
  getByDateRange: async () => ({ items: [] }),
} }))
vi.mock('@/api/user', () => ({ getMyPlatformQuotas: async () => ({ platform_quotas: [] }) }))
vi.mock('vue-i18n', async (importOriginal) => ({ ...(await importOriginal<typeof import('vue-i18n')>()), useI18n: () => ({ t: (key: string) => key }) }))

function openDashboard() {
  return mount(DashboardView, { global: { stubs: {
    AppLayout: { template: '<div><slot /></div>' },
    LoadingSpinner: true,
    UserDashboardStats: { props: ['stats', 'balance'], template: '<div data-test="stats">{{ stats.total_requests }} / {{ balance }}</div>' },
    UserDashboardCharts: true,
    UserDashboardRecentUsage: true,
    UserDashboardQuickActions: true,
  } } })
}

beforeEach(() => {
  vi.clearAllMocks()
  auth.user.balance = 10
  mocks.getDashboardStats.mockResolvedValue({ total_requests: 12 })
})

describe('dashboard independent loading', () => {
  it('renders statistics while profile refresh is pending and updates balance afterward', async () => {
    let resolveProfile!: () => void
    mocks.refreshUser.mockReturnValue(new Promise<void>((resolve) => { resolveProfile = resolve }))
    const wrapper = openDashboard()
    await flushPromises()
    expect(wrapper.get('[data-test="stats"]').text()).toBe('12 / 10')
    auth.user.balance = 20
    resolveProfile()
    await flushPromises()
    expect(wrapper.get('[data-test="stats"]').text()).toBe('12 / 20')
    wrapper.unmount()
  })

  it('keeps statistics visible when the profile refresh fails', async () => {
    const log = vi.spyOn(console, 'error').mockImplementation(() => {})
    mocks.refreshUser.mockRejectedValue(new Error('profile unavailable'))
    const wrapper = openDashboard()
    await flushPromises()
    expect(wrapper.get('[data-test="stats"]').text()).toBe('12 / 10')
    wrapper.unmount()
    log.mockRestore()
  })

  it('shows a retry action when statistics fail to load and recovers on retry', async () => {
    const log = vi.spyOn(console, 'error').mockImplementation(() => {})
    mocks.refreshUser.mockResolvedValue(undefined)
    mocks.getDashboardStats.mockRejectedValueOnce(new Error('stats unavailable'))
    const wrapper = openDashboard()
    await flushPromises()
    expect(wrapper.find('[data-test="stats"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('dashboard.loadFailed')
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.get('[data-test="stats"]').text()).toBe('12 / 10')
    wrapper.unmount()
    log.mockRestore()
  })
})
