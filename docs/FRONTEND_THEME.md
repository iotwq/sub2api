# Frontend Theme

The site uses a black, white, and grey palette with a light and a dark mode. The previous China Red brand color and the mint accent on `/home` and the authentication routes have been removed.

## Palette

Defined in `frontend/tailwind.config.js` and `frontend/src/style.css`.

- `gray`, `accent`, and `primary` 50-400 are neutral zinc greys. `gray` overrides Tailwind's default blue-tinted grey. `primary-50/100/200` sit one zinc step darker than `gray-50/100/200` so light tint backgrounds (`bg-primary-50`) stay visible on white.
- `dark` is the near-black dark-mode surface scale (`dark-950` page, `dark-900` / `dark-800` panels, `dark-700` / `dark-600` borders).
- `primary-500/600/700` are the solid-fill shades. They read the CSS variables `--primary-500/600/700` (space-separated RGB), which are near-black in light mode and mid-grey in dark mode. White text on `bg-primary-*` therefore stays readable in both modes, and a switch that is on (`bg-primary-600`) stays distinguishable from one that is off (`dark:bg-dark-600`).
- `primary-800/900/950` read `--primary-800/900/950`. In light mode they are the zinc 800/900/950 values. In dark mode they are mid greys, because about 150 components use them as translucent dark-mode tints and borders (`dark:bg-primary-900/20`, `dark:border-primary-800`). With fixed near-black values those tints were invisible on the near-black page.
- Semantic colors are unchanged: red for danger and errors, emerald for success, amber for warning, blue for info. Chart series colors are data colors and are unchanged.

## Shared classes

- `.btn-primary`: black with white text in light mode, white with black text in dark mode.
- `.badge-primary`, `.stat-icon-primary`, `.sidebar-link-active`: grey fills that follow the mode.
- `.brand-bg`: neutral gradient background (renamed from `.china-red-bg`), used by `KeyUsageView.vue`.
- `.btn` uses `focus-visible` for its ring, so mouse clicks no longer show it.

## Theme on first paint

An inline script at the top of `<head>` in `frontend/index.html` adds the `dark` class to `<html>` before the first paint. It uses the same rule as `initThemeClass()` in `src/main.ts`: `localStorage.theme` set to `'dark'`, or no saved theme while the system prefers dark. Before this script existed, the class was only set once the app module ran, so dark-mode users saw a light page for a moment on every full page load. Change both copies together; `main.ts` still sets the class, as a fallback.

The script carries `nonce="__CSP_NONCE_VALUE__"`. In embedded builds the backend's `FrontendServer` replaces every occurrence of that placeholder with the per-request CSP nonce, as it already does for the injected `window.__APP_CONFIG__` script, so the default policy needs no `'unsafe-inline'`. The backend leaves the placeholder unreplaced in the setup wizard, in the fallback used when public settings fail to load, and in legacy mode (`ServeEmbeddedFrontend`). There the browser blocks the script and logs one CSP error in the console, and the page flashes light as it did before; the theme is still set once `main.ts` runs. The Vite dev server sends no CSP, so the script always runs there.

## Header balance accent

The top-right available-balance badge deliberately uses Tailwind's fixed `teal` palette instead of the neutral `primary` palette. In light mode it uses a pale teal background, teal-600 icon, and teal-700 amount; in dark mode it uses a translucent teal background, teal-400 icon, and teal-300 amount. A subtle inset teal ring outlines the badge, and tabular numerals keep the amount aligned. The mobile user-menu balance uses the same amount colors. Switching themes must preserve the green accent; frozen-balance warnings retain their amber color.

Check the badge in both themes at desktop and tablet widths, plus the user-menu balance on mobile. Existing balance values, money formatting, hover details, and responsive visibility remain unchanged.

## Home and authentication stage

`HomeView.vue`, `AuthLayout.vue`, and `LoginView.vue` read the `--stage-*` variables in `style.css` (`--stage-bg-rgb`, `--stage-ink-rgb`, `--stage-surface`, `--stage-text`, `--stage-muted`, `--stage-faint`, `--stage-accent`, `--stage-accent-fg`). Light mode is white with black text; dark mode is black with white text. These pages no longer carry their own `:global(.dark)` color overrides, so change colors in the variables, not per page.

## Workspace page sizing

`/api-docs`, `/community-chat`, and `/image` match the normal (non-focus) `/chat` frame. Their responsive negative margins offset `AppLayout` padding to leave an 8px gap next to the sidebar, header, and right edge. Outer padding is 6px on mobile and 8px from the `md` breakpoint.

The chat and image panels use the same inner height as model chat: `max(420px, 100dvh - 6rem)`, or `max(420px, 100dvh - 6.25rem)` from `md`. The documentation frame includes that height plus its padding and border (`100dvh - 5.125rem`, minimum 434px/438px) and scrolls internally, keeping all model sections accessible. Short viewports retain model chat's minimum usable height.

The image page remains the existing external-workspace entry; this sizing change does not alter its destination or turn it into an embedded page. Check all four route frames at desktop, tablet, and phone widths in both themes when changing these dimensions.

The shared header title and description can shrink and truncate when space is limited, with their complete text available on hover. This prevents the longer API documentation description from pushing account controls offscreen at 1024px and adding a page-wide horizontal scrollbar.

## Other neutralized surfaces

- `src/styles/onboarding.css`: the dark tour popover uses the `dark` scale instead of slate, and the "next" button is black on white (white on black in dark mode) instead of teal.
- The tour footer wraps its progress/keyboard hints and action buttons onto separate lines when space is limited. Actions stay right-aligned and fully inside the popover; check both Chinese and English at 320px and 375px widths.
- `SettingsView.vue`: the dark settings tab bar uses the `dark` scale instead of slate, and the active-tab underline is `currentColor` instead of a teal-to-sky gradient.
- `AnnouncementBell.vue`, `AnnouncementPopup.vue`, `KeyUsageView.vue`: scrollbar and skeleton gradients moved from slate to zinc.
- `DataTable.vue`: the sticky header row and the sticky left/right columns hard-coded Tailwind's old blue-tinted grey (`rgb(17 24 39)`, `rgb(31 41 55)`), which showed as navy strips in every admin table in dark mode. They now match the table's own `dark-900` body and `dark-800` header.
- `AppSidebar.vue`: the dark section-title divider moved from old grey to `dark-700`.
- Chart.js grid, axis text, and tooltip colors moved from Tailwind's old greys to zinc in the admin dashboard, `TokenUsageTrend.vue`, both `AccountStatsModal.vue` files, `MonitorTrendChart.vue`, and the ops charts (`OpsLatencyChart`, `OpsThroughputTrendChart`, `OpsErrorTrendChart`, `OpsSwitchRateTrendChart`, `OpsErrorDistributionChart`). Dark grid lines are `#3f3f46` and dark tooltips `#27272a`.
- `RelayPulseMatrix.vue` tooltips, the `AliyunCaptchaWidget.vue` button, and the dark scrollbar thumbs in `AnnouncementBell.vue` / `AnnouncementPopup.vue` moved from old greys to zinc. The captcha change is CSS only.
- `DashboardView.vue`: the "Token / 实际消费" toggle above the user trend used a blue active state. It now uses the same grey segmented style as the "模型分布 / 用户消费榜" toggle in `ModelDistributionChart.vue` (white active pill in light mode, `dark-700` pill in dark mode).

## Not covered

- `frontend/public/dragon-logo.png` and `dragon-logo.svg` are still the red dragon artwork. They need new design assets and are not recolored in code.
- The unread dot in `AppSidebar.vue` stays red because it is a notification signal and a test asserts its color.
- Teal and cyan info callouts (for example the pricing notice on `/keys`, about 18 files) are kept as an informational color, like blue.
- The warm beige surfaces in chat, API docs, and image pages are a separate page design and were not changed.
- Light-mode helper text using `text-gray-400` (sidebar section titles, hints) measures about 2.5:1 against the page. That contrast is the same as before this change and was not adjusted.
- `CommunityEmojiPicker.vue` keeps its dark navy colors because they match the community chat page's own dark palette (`#18212c` and similar in `CommunityChatView.vue`).
- `#9ca3af` stays where it is a data or status color: the `gray` series in the ops error charts, the unknown-health cell in `RelayPulseMatrix.vue`, and the idle score in `OpsDashboardHeader.vue`.
- Light-mode text in `onboarding.css` still uses old near-black greys (`#111827`, `#1f2937`); next to zinc they look the same.
- Sky-blue info cards (for example "Codex 图片桥接策略" in the account edit modal) and platform brand colors (Anthropic orange, Alipay blue) are kept.

## Verification

From `frontend/`:

```bash
pnpm typecheck
pnpm test:run
pnpm exec vite build --outDir node_modules/.theme-check --emptyOutDir
```

Then open `/home`, `/login`, and `/key-usage` in both themes. The last build command writes to a throwaway folder because the default `outDir` is the tracked `backend/internal/web/dist`; delete `node_modules/.theme-check` afterwards.
