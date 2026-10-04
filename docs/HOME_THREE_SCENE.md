# Home Three.js Agent Scene

The default `/home` page uses a full-bleed Three.js scene as the primary first-screen signal for AI and Agent research. The canvas is not framed inside a decorative preview card.

## Files

- `frontend/src/components/home/HomeThreeScene.vue`: Agent orchestration network, data-flow animation, pointer response, responsive canvas sizing, and cleanup.
- `frontend/src/views/HomeView.vue`: full-bleed homepage composition, semantic Agent node labels, live execution trace, and research tracks.
- `frontend/src/components/layout/AuthLayout.vue`: shared full-bleed Agent scene for login, registration, recovery, and authentication callback routes.
- `frontend/src/i18n/locales/{zh,en}/landing.ts`: localized Agent system, execution trace, and research-track copy.
- `frontend/package.json`: adds `three` and `@types/three`.

## Behavior

- The scene is loaded by the default homepage and shared authentication layout; configured HTML or iframe home content still bypasses the homepage instance.
- The homepage header brand shows only the configured site logo and site name; it does not render the site subtitle.
- A central orchestrator connects to reasoning, tools, memory, and evaluation nodes through four curved data paths.
- Twelve visible packets move along those paths to make task flow legible rather than adding unrelated ambient motion.
- Pointer movement changes the network viewing angle without moving or resizing the page layout.
- Users with `prefers-reduced-motion: reduce` get a single rendered frame instead of a continuous animation loop.
- The canvas resizes with its container through `ResizeObserver`.
- WebGL resources are disposed when the component unmounts.
- Desktop shows the complete node descriptions and live trace metrics. Narrow screens hide secondary descriptions and compress the trace so labels do not overlap the hero copy.

## Verification

Run these commands from `frontend/`:

```bash
pnpm typecheck
pnpm exec eslint src/views/HomeView.vue src/components/home/HomeThreeScene.vue src/i18n/locales/zh/landing.ts src/i18n/locales/en/landing.ts
pnpm build
```

Then open `/home` and `/login` and confirm:

- The page is not blank.
- Exactly one `.home-three-scene` canvas is rendered on each route and has non-zero dimensions.
- Canvas pixels are nonblank and differ between frames when reduced motion is disabled.
- The scene is framed to the right of the hero copy on desktop and below it on mobile.
- Desktop and mobile widths do not create horizontal overflow or text/node overlap.
- The login/dashboard action, documentation link, locale control, theme control, and research anchor remain interactive.
