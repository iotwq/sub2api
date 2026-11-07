import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import { useOnboardingTour } from '../useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'

const mocks = vi.hoisted(() => {
  let resolveDriver!: () => void
  let reportRequested!: () => void
  const ready = new Promise<void>((resolve) => { resolveDriver = resolve })
  const requested = new Promise<void>((resolve) => { reportRequested = resolve })
  return { ready, requested, resolveDriver, reportRequested, driver: vi.fn() }
})

vi.mock('driver.js', async () => {
  mocks.reportRequested()
  await mocks.ready
  return { driver: mocks.driver }
})
vi.mock('driver.js/dist/driver.css', () => ({}))
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ user: { id: 1, role: 'admin' }, isSimpleMode: false }),
}))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

function mountTour() {
  let tour!: ReturnType<typeof useOnboardingTour>
  const wrapper = mount(defineComponent({
    setup() {
      tour = useOnboardingTour({ autoStart: false })
      return () => null
    },
  }))
  return { wrapper, tour }
}

describe('onboarding lazy loading lifecycle', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mocks.driver.mockImplementation((config: { onDestroyed: () => void }) => ({
      drive: vi.fn(),
      isActive: () => true,
      destroy: vi.fn(() => config.onDestroyed()),
    }))
  })

  afterEach(() => {
    for (const result of mocks.driver.mock.results) {
      if (result.type === 'return') result.value.destroy()
    }
    vi.clearAllMocks()
  })

  it('does not start the old page tour after navigating away during the first library download', async () => {
    const oldPage = mountTour()
    const oldStart = oldPage.tour.startTour()
    await mocks.requested
    oldPage.wrapper.unmount()

    mocks.resolveDriver()
    await oldStart
    expect(mocks.driver.mock.calls.length).toBe(0)

    const currentPage = mountTour()
    await currentPage.tour.startTour()
    expect(mocks.driver).toHaveBeenCalledTimes(1)
    expect(useOnboardingStore().getDriverInstance()?.drive).toHaveBeenCalledTimes(1)
    currentPage.wrapper.unmount()
  })

  it('keeps an already active tour available across page navigation', async () => {
    const oldPage = mountTour()
    await oldPage.tour.startTour()
    const driver = useOnboardingStore().getDriverInstance()
    oldPage.wrapper.unmount()

    const currentPage = mountTour()
    await flushPromises()
    expect(useOnboardingStore().getDriverInstance()).toBe(driver)
    expect(driver?.destroy).not.toHaveBeenCalled()
    expect(mocks.driver).toHaveBeenCalledTimes(1)
    currentPage.wrapper.unmount()
  })
})
