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

## Home and authentication stage

`HomeView.vue`, `AuthLayout.vue`, and `LoginView.vue` read the `--stage-*` variables in `style.css` (`--stage-bg-rgb`, `--stage-ink-rgb`, `--stage-surface`, `--stage-text`, `--stage-muted`, `--stage-faint`, `--stage-accent`, `--stage-accent-fg`). Light mode is white with black text; dark mode is black with white text. These pages no longer carry their own `:global(.dark)` color overrides, so change colors in the variables, not per page.

## Other neutralized surfaces

- `src/styles/onboarding.css`: the dark tour popover uses the `dark` scale instead of slate, and the "next" button is black on white (white on black in dark mode) instead of teal.
- `SettingsView.vue`: the dark settings tab bar uses the `dark` scale instead of slate, and the active-tab underline is `currentColor` instead of a teal-to-sky gradient.
- `AnnouncementBell.vue`, `AnnouncementPopup.vue`, `KeyUsageView.vue`: scrollbar and skeleton gradients moved from slate to zinc.
- `DataTable.vue`: the sticky header row and the sticky left/right columns hard-coded Tailwind's old blue-tinted grey (`rgb(17 24 39)`, `rgb(31 41 55)`), which showed as navy strips in every admin table in dark mode. They now match the table's own `dark-900` body and `dark-800` header.
- `AppSidebar.vue`: the dark section-title divider moved from old grey to `dark-700`.
- Chart.js grid, axis text, and tooltip colors moved from Tailwind's old greys to zinc in the admin dashboard, `TokenUsageTrend.vue`, both `AccountStatsModal.vue` files, `MonitorTrendChart.vue`, and the ops charts (`OpsLatencyChart`, `OpsThroughputTrendChart`, `OpsErrorTrendChart`, `OpsSwitchRateTrendChart`, `OpsErrorDistributionChart`). Dark grid lines are `#3f3f46` and dark tooltips `#27272a`.
- `RelayPulseMatrix.vue` tooltips, the `AliyunCaptchaWidget.vue` button, and the dark scrollbar thumbs in `AnnouncementBell.vue` / `AnnouncementPopup.vue` moved from old greys to zinc. The captcha change is CSS only.

## Not covered

- `frontend/public/dragon-logo.png` and `dragon-logo.svg` are still the red dragon artwork. They need new design assets and are not recolored in code.
- The unread dot in `AppSidebar.vue` stays red because it is a notification signal and a test asserts its color.
- Teal and cyan info callouts (for example the pricing notice on `/keys`, about 18 files) are kept as an informational color, like blue.
- The warm beige surfaces in chat, API docs, and image pages are a separate page design and were not changed.
- Light-mode helper text using `text-gray-400` (sidebar section titles, hints) measures about 2.5:1 against the page. That contrast is the same as before this change and was not adjusted.
- `CommunityEmojiPicker.vue` keeps its dark navy colors because they match the community chat page's own dark palette (`#18212c` and similar in `CommunityChatView.vue`).
- `#9ca3af` stays where it is a data or status color: the `gray` series in the ops error charts, the unknown-health cell in `RelayPulseMatrix.vue`, and the idle score in `OpsDashboardHeader.vue`.
- Light-mode text in `onboarding.css` still uses old near-black greys (`#111827`, `#1f2937`); next to zinc they look the same.
- The "Token / 实际消费" toggle above the admin dashboard user trend uses a blue active state, while the "模型分布 / 用户消费榜" toggle on the same page uses grey. Not changed.
- Sky-blue info cards (for example "Codex 图片桥接策略" in the account edit modal) and platform brand colors (Anthropic orange, Alipay blue) are kept.

## Verification

From `frontend/`:

```bash
pnpm typecheck
pnpm test:run
pnpm exec vite build --outDir node_modules/.theme-check --emptyOutDir
```

Then open `/home`, `/login`, and `/key-usage` in both themes. The last build command writes to a throwaway folder because the default `outDir` is the tracked `backend/internal/web/dist`; delete `node_modules/.theme-check` afterwards.
