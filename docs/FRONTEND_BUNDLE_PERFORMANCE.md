# Frontend bundle performance

Measured with a production build (`pnpm build` in `frontend/`, which type-checks and then runs `vite build`) and a static analysis of the route-to-chunk graph. Numbers are gzipped JavaScript, in KB.

## What changed

- `frontend/vite.config.ts` no longer merges every unmatched third-party module into one `vendor-misc` chunk. That chunk had grown to 285 KB gzipped and contained `three` (the 3D scene renderer, 176 KB gzipped on its own), `bluebird`, `highlight.js`, and the `mammoth` document parser, alongside small chart and table helpers. Because the chart and table helpers were in it, every page with a chart or a table downloaded all of it. Unmatched modules are now split by how the pages actually use them.
- `frontend/src/App.vue` loads the announcement popup and the admin compliance dialog only after sign-in (the compliance dialog only for admins). Both pull in the markdown renderer, so the home page and the sign-in page no longer download it.
- The ten authentication views import `AuthLayout` directly instead of through the `components/layout` barrel, which also re-exports the signed-in app shell and its guided-tour dependency.
- `frontend/src/components/home/HomeThreeScene.vue` stops the animation loop while the canvas is scrolled out of view and resumes it when the canvas comes back. The renderer itself is unchanged.

## Route download, before and after

| Route | Before | After | Change |
| --- | --- | --- | --- |
| `/home` | 315.5 KB, 9 files | 280.4 KB, 8 files | -35 KB |
| `/login` | 336.4 KB, 16 files | 294.7 KB, 13 files | -42 KB |
| `/dashboard` | 699.0 KB, 19 files | 413.8 KB, 19 files | -285 KB |
| `/admin/dashboard` | 698.2 KB, 17 files | 413.0 KB, 17 files | -285 KB |
| `/keys` | 680.9 KB, 31 files | 409.0 KB, 34 files | -272 KB |

The shared code every route downloads dropped from 162.1 KB to 127.1 KB gzipped, because the markdown renderer left that set.

`/keys` downloads three more files than before. The old shared chunk is now several smaller chunks, so a page pays for exactly the libraries it uses and nothing else. The extra requests are small and cacheable, and the total is smaller.

## What was checked in a running build

A preview of the production build was served with a stand-in backend, and the browser's record of downloaded files was read per route.

- No route reported a script error, and no route downloaded a `vendor-misc` file.
- Signed out, `/home` and `/login` downloaded no markdown renderer and no guided-tour library.
- Signed in, `/dashboard`, `/keys`, `/admin/dashboard`, `/admin/accounts`, `/admin/settings`, and `/chat` downloaded no `three`.
- On `/home`, the animation ran at 120 frames per second while the canvas was on screen and dropped to 0 after scrolling past it, then resumed at 120 after scrolling back.
- With an announcement queued, the popup appeared after sign-in, rendered its markdown (bold text included), and locked page scrolling while open.

## Left as is

- `/login` still downloads `three` (about 176 KB gzipped). The sign-in page shows the same animated background as the home page, so removing it would change the page. The saving is available if that background is dropped or replaced with a still image.
- The largest remaining piece on every route is the locale data at about 146 KB gzipped, which is one file per language rather than one shared file. Splitting it per page would need a change to how translations are loaded, and was not attempted.
- The dev server reads `frontend/vite.config.js`, which is git-ignored and regenerated from `vite.config.ts` by the type-check step inside `pnpm build`. A `vite build` run on its own, without that step, can keep using a stale `vite.config.js`. Use `pnpm build`.

## Second pass: signed-in pages

### What changed

- The pdf.js worker address (`pdf.worker.mjs?url`) was being grouped into the `vendor-pdf` chunk, so the chat page statically depended on the whole PDF library (132 KB gzipped) as soon as it opened. The worker address is now excluded from that chunk rule. The PDF library is downloaded only when a PDF is attached.
- Rollup's shared CommonJS helper module had been placed inside the QR code chunk after `vendor-misc` was removed, so pages using `file-saver` (keys, usage, chat) downloaded the QR code library (9 KB gzipped) without using it. The helper now has its own chunk, `vendor-cjs-helpers` (0.4 KB gzipped).
- `marked` and `DOMPurify` are split into separate chunks. The sidebar only needs `DOMPurify` for custom menu icons. `marked` is loaded on demand through the new `frontend/src/utils/markdown.ts` by the announcement bell (when its list opens), the announcement popup (when an announcement is shown), and the admin compliance dialog (when it is shown).
- `markdown.ts` passes its parse options per call instead of changing `marked`'s global defaults. Before this change, the announcement components set `breaks: true` globally on load, which the model plaza description and custom pages relied on without setting it themselves. Both now pass `{ breaks: true, gfm: true }` explicitly, so they render as before. On the public model plaza page, signed-out visitors now also get single-line breaks, which they did not before.
- The guided-tour library (`driver.js` and its CSS) is imported inside `startTour()` instead of at the top of `useOnboardingTour.ts`. Users who have already seen the tour never download it.
- `frontend/public/dragon-logo.png` went from a 256×256 image of 145 KB to a 128×128, 256-colour image of 16.6 KB. It is shown at 36 to 40 px, so 128 px still covers 3× screens.
- The user dashboard shows an error message with a refresh button when the statistics request fails. Before, the page went blank with no message.

### Extra download per page, before and after

These numbers are what each page adds on top of the shared startup code and language file, gzipped.

| Page | Before | After |
| --- | --- | --- |
| Shared signed-in shell (sidebar and header) | 58.7 KB | 41.1 KB |
| `/chat` | 244.0 KB | 97.3 KB |
| `/admin/usage` | 202.4 KB | 175.8 KB |
| `/admin/settings` | 189.6 KB | 172.0 KB |
| `/usage` | 176.3 KB | 158.8 KB |
| `/dashboard` | 138.4 KB | 120.9 KB |
| `/admin/dashboard` | 137.6 KB | 120.0 KB |
| `/keys` | 133.7 KB | 107.1 KB |

`/home` and `/login` did not change in this pass.

### What was checked in a running build

- `/dashboard`, `/admin/dashboard`, `/keys`, `/usage`, `/admin/usage`, and `/chat` loaded without script errors. None of them downloaded the PDF, QR code, or guided-tour library. Only `/chat`, which renders chat messages as Markdown, downloaded `marked`.
- `marked` was not downloaded until an announcement was shown. The popup then rendered bold text; the bell's detail view rendered bold text and line breaks.
- With the tour marked unseen, the tour started on `/admin/dashboard`, downloaded its library and stylesheet, and showed the styled dark popover. With the tour marked seen, the library was not downloaded.
- With the statistics request failing, `/dashboard` showed the error message and refresh button. After the request was restored, refresh brought the statistics back.
- Attaching a one-page PDF on `/chat` downloaded the PDF library and its worker only at that point, and the attachment showed "1 页".
- With compliance acknowledgement required, the admin compliance dialog rendered the document headings and paragraphs and showed the confirmation input.
- `/profile` and `/purchase`, which use the QR code library, loaded without errors. A QR code image itself was not generated in this check.

## Third pass: settings page and admin dashboard

### What changed

- `frontend/src/views/admin/SettingsView.vue` used to render all eleven tab panels when the page opened and only hide the inactive ones. Each panel now mounts the first time its tab is opened and then stays mounted, so switching back keeps unsaved input. The payment provider list is loaded with `defineAsyncComponent`, so it and the drag-and-drop library (`vue-draggable-plus`, 14.6 KB gzipped) are downloaded only when the payment tab is opened.
- Saving still sends every setting, including those on tabs that were never opened, because the save request is built from the page's form state rather than from the visible fields.
- `frontend/src/views/admin/DashboardView.vue` shows an error message with a refresh button when the dashboard statistics request fails. Before, it showed an error toast and an empty page. The toast is kept.

### Measured in a running build

| Check | Before | After |
| --- | --- | --- |
| Elements rendered when `/admin/settings` opens | 1,838 (154 inputs) | 145 (10 inputs) |
| Time from navigation to settings content, cached chunks | about 100 ms | about 25 ms |
| Drag-and-drop library downloaded without opening the payment tab | yes | no |
| `/admin/dashboard` when statistics fail | empty page | error message and refresh button |

Times come from a background browser tab and are only useful for comparison with each other. After the payment tab was opened, the provider list and drag-and-drop library downloaded at that point, the payment panel stayed in the page when switching to another tab, and the same panel was shown again on return. Saving with only the default tab opened sent all 278 settings fields.

### Measured and left as is

When the batch-image permission check in `frontend/src/composables/useBatchImageAccess.ts` fails, each menu switch sends one more request. With the check failing, six menu switches sent six requests; each one fails on the first page and does not page through the key list. An earlier note in this analysis described it as re-reading the whole key list on every switch, which was wrong. The retry is intentional: `useBatchImageAccess.retry.spec.ts` requires a failed lookup to be retried by the next caller, so the batch-image menu reappears as soon as the request succeeds again. It was not changed.


## Regression review: 2026-11-06

Two regressions introduced by lazy loading were reproduced and fixed:

- If a page was left while the guided-tour library was still downloading, its pending startup could create an overlay on the new page. The composable now cancels startup after its owning layout unmounts. An already active tour still persists across route changes, as before.
- Announcement and admin-compliance text was empty until the Markdown chunk finished downloading, and remained empty if that download failed. The shared renderer now displays an escaped, whitespace-preserving plain-text version while the parser is unavailable. It automatically switches to sanitized Markdown after loading succeeds. This does not change announcement read state or compliance acknowledgement rules.

Verification used production bundles and a local simulated backend. Delaying the tour chunk reproduced an unwanted overlay on `/home` before the fix and none afterward. Delaying Markdown reproduced a blank announcement before the fix and readable text afterward; the normal formatted version still appeared once loading finished. A failed Markdown download left the compliance document readable, without an unhandled exception.

The review also checked API-documentation tabs and copying at widths of 320, 768, and 1440 pixels, light/dark themes, payment-component loading and reuse, and actual PNG generation using the bundled QR library. All passed. Frontend tests passed across 407 files and 3338 cases; type checking, ESLint on the 22 reviewed/new TypeScript and Vue paths, and production builds also passed. These are local regression checks, not production backend or paid-upstream integration tests.

### Clarification: payment component versus data request

The payment provider **component code** and drag-and-drop library load when the payment tab is first opened. The provider **API request** still runs once when the settings page mounts through the existing `loadProviders()` call. Two visits to the payment tab reused the same component and did not add another provider request. The earlier description of downloading the provider list on tab entry refers to the component bundle, not deferred API data loading; that request behavior was not changed during this review.
