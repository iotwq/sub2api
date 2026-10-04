import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar scroll position persistence', () => {
  it('binds a template ref to the sidebar nav element', () => {
    expect(componentSource).toContain('ref="sidebarNavRef"')
    expect(componentSource).toContain('sidebar-nav')
  })

  it('declares sidebarNavRef in script setup', () => {
    expect(componentSource).toContain("const sidebarNavRef = ref<HTMLElement | null>(null)")
  })

  it('saves scroll position on beforeUnmount', () => {
    expect(componentSource).toContain('onBeforeUnmount')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('sidebarNavRef.value.scrollTop')
  })

  it('restores scroll position on mount', () => {
    expect(componentSource).toContain('onMounted')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('nextTick')
  })
})

describe('AppSidebar collapsible groups', () => {
  it('lets the user collapse a group even while a child route is active', () => {
    // The expand state must come from the user's override first, falling back
    // to the active-route heuristic only when the user has not clicked yet.
    expect(componentSource).toContain('const groupExpandOverrides = ref<Map<string, boolean>>(new Map())')
    expect(componentSource).not.toContain('expandedGroups.value.has(item.path) || isGroupActive(item)')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar subscription feature flag', () => {
  it('gates the My Subscriptions entry behind the subscription public-settings flag', () => {
    expect(componentSource).toContain('const flagSubscription = makeSidebarFlag(FeatureFlags.subscription)')
    expect(componentSource).toMatch(/path: '\/subscriptions'[^\n]*featureFlag: flagSubscription/)
  })

  it('also hides the admin Subscription Management entry on recharge-only sites', () => {
    expect(componentSource).toMatch(/path: '\/admin\/subscriptions'[^\n]*featureFlag: flagSubscription/)
  })

  it('derives the purchase entry label from the site billing mode', () => {
    expect(componentSource).toContain("import { resolveSiteBillingMode } from '@/utils/siteBillingMode'")
    expect(componentSource).toMatch(/case 'recharge_only':\s*return t\('nav\.recharge'\)/)
    expect(componentSource).toMatch(/case 'subscription_only':\s*return t\('nav\.subscribe'\)/)
    expect(componentSource).toMatch(/path: '\/purchase'[^\n]*label: purchaseNavLabel\.value/)
  })
})

describe('AppSidebar community chat unread badge', () => {
  it('keeps the expanded badge beside the menu label in every navigation branch', () => {
    expect(componentSource.match(/class="sidebar-label sidebar-label-with-badge"/g)).toHaveLength(3)
    expect(componentSource.match(/shouldShowCommunityChatUnread\(item\.path\) && !sidebarCollapsed/g)).toHaveLength(3)
  })

  it('anchors the collapsed badge to the menu icon in every navigation branch', () => {
    expect(componentSource.match(/class="sidebar-menu-icon"/g)).toHaveLength(3)
    expect(componentSource.match(/class="sidebar-unread-dot sidebar-unread-dot-collapsed"/g)).toHaveLength(3)
  })

  it('uses a fixed-size high-contrast badge', () => {
    const unreadDotBlockMatch = componentSource.match(/\.sidebar-unread-dot\s*\{[\s\S]*?\n\}/)

    expect(unreadDotBlockMatch).not.toBeNull()
    expect(unreadDotBlockMatch?.[0]).toContain('width: 0.625rem;')
    expect(unreadDotBlockMatch?.[0]).toContain('height: 0.625rem;')
    expect(unreadDotBlockMatch?.[0]).toContain('flex: 0 0 0.625rem;')
    expect(unreadDotBlockMatch?.[0]).toContain('background: #dc2626;')
    expect(unreadDotBlockMatch?.[0]).not.toContain('right: 0.65rem;')
  })

  it('keeps group and direct unread state separate', () => {
    expect(componentSource).toContain('const communityChatGroupUnread = ref(false)')
    expect(componentSource).toContain('const communityChatDirectUnread = ref(false)')
    expect(componentSource).toContain('(communityChatGroupUnread.value || communityChatDirectUnread.value)')

    expect(componentSource).toContain('communityChatAPI.getDirectUnread()')
    expect(componentSource).toContain('communityChatDirectUnread.value = summary.total > 0')
    expect(componentSource).not.toContain('let communityChatSocket: WebSocket')
    expect(componentSource).toContain('useCommunityChatRealtime')
  })

  it('does not clip the badge and blinks unless reduced motion is requested', () => {
    expect(componentSource.match(/'sidebar-link-with-unread': shouldShowCommunityChatUnread\(item\.path\)/g)).toHaveLength(3)
    expect(componentSource).toContain('.sidebar-link-with-unread {\n  overflow: visible;')
    expect(componentSource).toContain('.sidebar-label-with-badge {')
    expect(componentSource).toContain('animation: sidebar-unread-blink 1.2s ease-in-out infinite;')
    expect(componentSource).toContain('@media (prefers-reduced-motion: reduce)')
  })
})
