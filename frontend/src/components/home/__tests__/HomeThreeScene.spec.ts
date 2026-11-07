import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import HomeThreeScene from '../HomeThreeScene.vue'

const sceneMocks = vi.hoisted(() => {
  let finishLoading!: () => void
  const loading = new Promise<void>((resolve) => { finishLoading = resolve })
  const renderer = {
    setClearColor: vi.fn(),
    setPixelRatio: vi.fn(),
    setSize: vi.fn(),
    render: vi.fn(),
    dispose: vi.fn(),
  }
  return {
    loading,
    finishLoading,
    renderer,
    createRenderer: vi.fn(function () { return renderer }),
  }
})

vi.mock('three', async (importOriginal) => {
  await sceneMocks.loading
  const actual = await importOriginal<typeof import('three')>()
  return { ...actual, WebGLRenderer: sceneMocks.createRenderer }
})

describe('HomeThreeScene lifecycle', () => {
  let onVisibilityChange: IntersectionObserverCallback
  let onResize: ResizeObserverCallback
  let frames: Map<number, FrameRequestCallback>
  let nextFrameId: number
  let wrapper: ReturnType<typeof mount> | undefined
  const disconnectVisibility = vi.fn()
  const disconnectResize = vi.fn()

  function runFrame() {
    const entry = frames.entries().next().value
    expect(entry).toBeDefined()
    const [id, callback] = entry!
    frames.delete(id)
    callback(16)
  }

  function setVisible(isIntersecting: boolean) {
    onVisibilityChange([{ isIntersecting } as IntersectionObserverEntry], {} as IntersectionObserver)
  }

  beforeEach(() => {
    vi.clearAllMocks()
    frames = new Map()
    nextFrameId = 0
    vi.spyOn(window, 'matchMedia').mockReturnValue({ matches: false } as MediaQueryList)
    vi.spyOn(window, 'requestAnimationFrame').mockImplementation((callback) => {
      const id = ++nextFrameId
      frames.set(id, callback)
      return id
    })
    vi.spyOn(window, 'cancelAnimationFrame').mockImplementation((id) => { frames.delete(id) })
    vi.stubGlobal('IntersectionObserver', class {
      constructor(callback: IntersectionObserverCallback) { onVisibilityChange = callback }
      observe = vi.fn()
      disconnect = disconnectVisibility
    })
    vi.stubGlobal('ResizeObserver', class {
      constructor(callback: ResizeObserverCallback) { onResize = callback }
      observe = vi.fn()
      disconnect = disconnectResize
    })
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = undefined
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('does not initialize the scene when loading finishes after navigation away', async () => {
    wrapper = mount(HomeThreeScene)
    wrapper.unmount()
    wrapper = undefined

    sceneMocks.finishLoading()
    await import('three')
    await flushPromises()

    expect(sceneMocks.createRenderer).not.toHaveBeenCalled()
    expect(frames.size).toBe(0)
  })

  it('pauses offscreen and resumes only one animation loop when visible', async () => {
    wrapper = mount(HomeThreeScene)
    await flushPromises()
    expect(sceneMocks.renderer.render).toHaveBeenCalledOnce()
    expect(frames.size).toBe(1)

    setVisible(false)
    runFrame()
    expect(frames.size).toBe(0)

    setVisible(true)
    setVisible(true)
    expect(frames.size).toBe(1)
    runFrame()
    expect(frames.size).toBe(1)
  })

  it('releases the renderer, observers and scheduled frame on unmount', async () => {
    wrapper = mount(HomeThreeScene)
    await flushPromises()
    wrapper.unmount()
    wrapper = undefined

    expect(sceneMocks.renderer.dispose).toHaveBeenCalledOnce()
    expect(disconnectVisibility).toHaveBeenCalledOnce()
    expect(disconnectResize).toHaveBeenCalledOnce()
    expect(frames.size).toBe(0)
  })

  it('redraws the static scene after resizing without enabling animation', async () => {
    vi.mocked(window.matchMedia).mockReturnValue({ matches: true } as MediaQueryList)
    wrapper = mount(HomeThreeScene)
    await flushPromises()
    const rendersBeforeResize = sceneMocks.renderer.render.mock.calls.length

    onResize([], {} as ResizeObserver)
    expect(sceneMocks.renderer.render).toHaveBeenCalledTimes(rendersBeforeResize + 1)
    setVisible(false)
    setVisible(true)
    expect(frames.size).toBe(0)
  })
})
