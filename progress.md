## 2026-06-24 - Task: Clarify chat stream failure handling

### What was done
- Adjusted chat streaming parsing so final SSE events are still processed even when the connection closes without a trailing blank line.
- Changed upstream SSE error and failed events to surface as real send errors instead of being silently ignored and later shown as an empty assistant reply.
- Added a separate chat-page network interruption message for `Failed to fetch` and stream-disconnect style failures.
- Documented the operational difference between a transport interruption and a genuinely empty assistant response.

### Testing
- `pnpm exec vitest run src/views/user/__tests__/ChatView.spec.ts` passed with 15 tests.
- `pnpm vue-tsc -b` passed.
- `pnpm build` passed.

### Notes
- `frontend/src/api/chat.ts`: tightened SSE parsing, EOF flushing, SSE error extraction, and Responses JSON error propagation.
- `frontend/src/views/user/ChatView.vue`: normalized chat send errors so network interruptions no longer show as empty model replies.
- `frontend/src/i18n/locales/zh.ts`: added the Chinese network interruption copy.
- `frontend/src/i18n/locales/en.ts`: added the English network interruption copy.
- `frontend/src/views/user/__tests__/ChatView.spec.ts`: covered unterminated final SSE events and browser `Failed to fetch` behavior.
- `docs/UPSTREAM_SYNC.md`: added the deployment/runtime checklist for chat stream interruptions.
- Rollback: revert the files listed above, or restore from the git diff for this task before staging.

## 2026-06-27 - Task: Sync latest upstream/main

### What was done
- Fetched `Wei-Shaw/sub2api` and merged upstream range `85a3b122..c2754222` into local `main`, bringing upstream to `v0.1.139`.
- Preserved the existing local uncommitted changes by creating backup branch `backup/pre-upstream-sync-20260627-211603` and stash `pre-upstream-sync-20260627-211603`, then restored the stash after the upstream merge.
- Resolved upstream merge conflicts in README, dashboard charts, i18n, OpenAI image scheduling, Codex transform, and chat-completions tests.
- Kept local API Key Responses default-instructions behavior while preserving upstream OAuth behavior that does not inject default Codex instructions.

### Testing
- `go test ./internal/service ./internal/server/routes ./internal/handler` passed.
- `npm test -- --run src/views/user/__tests__/ChatView.spec.ts src/stores/__tests__/app.spec.ts src/components/admin/account/__tests__/AccountTestModal.spec.ts` passed with 40 tests.
- `npm run build` passed; only the existing Browserslist data warning appeared.
- `git diff --check` passed.

### Notes
- `docs/UPSTREAM_SYNC.md`: recorded the 2026-06-27 upstream sync range, conflict decisions, validation, and rollback points.
- `README_CN.md`: preserved local official-domain/demo notes while absorbing upstream warnings and sponsor updates.
- `frontend/src/components/charts/GroupDistributionChart.vue`: adopted upstream finite-number formatting for safer chart values.
- `frontend/src/components/charts/ModelDistributionChart.vue`: adopted upstream finite-number formatting for safer chart values.
- `frontend/src/views/admin/DashboardView.vue`: adopted upstream finite-number formatting for dashboard metrics.
- `frontend/src/i18n/locales/en.ts`: merged upstream Codex PAT text with existing local Codex JSON/AT text.
- `frontend/src/i18n/locales/zh.ts`: merged upstream Codex PAT text with existing local Codex JSON/AT text.
- `backend/internal/service/openai_account_scheduler.go`: preserved upstream image native-to-basic fallback behavior.
- `backend/internal/service/openai_codex_transform.go`: combined upstream Spark image-tool stripping with local custom function-tool bridge protection.
- `backend/internal/service/openai_gateway_chat_completions.go`: limited default Codex instruction injection to API Key Responses-compatible forwarding.
- `backend/internal/service/openai_gateway_chat_completions_test.go`: retained separate upstream OAuth and local API Key instruction behavior tests.
- Rollback: use backup branch `backup/pre-upstream-sync-20260627-211603`, stash `pre-upstream-sync-20260627-211603`, or revert merge commit `f3121a35` plus the restored local working-tree changes.

## 2026-06-27 - Task: Implement native community chat

### What was done
- Added a lightweight in-project user group chat page at `/community-chat` with a left-sidebar entry named “用户群聊”.
- Added PostgreSQL-backed message persistence, WebSocket realtime push, text messages, emoji shortcuts, image upload, and admin soft-delete.
- Stored uploaded chat images under the existing `DATA_DIR` tree so deployments can persist them with the existing data volume.
- Documented the feature scope, APIs, storage path, and deployment notes.

### Testing
- `go test ./internal/service ./internal/repository ./internal/handler ./internal/server/routes` passed.
- `pnpm build` passed; only the existing Browserslist data warning appeared.

### Notes
- `.gitignore`: allowed the new `docs/COMMUNITY_CHAT.md` file to be version-controlled despite the existing `docs/*` ignore rule.
- `backend/migrations/154_community_chat.sql`: added the `community_chat_messages` table and indexes.
- `backend/internal/service/community_chat.go`: added message validation, admin delete authorization, and in-process WebSocket event broadcasting.
- `backend/internal/repository/community_chat_repo.go`: added SQL persistence and pagination for chat messages.
- `backend/internal/handler/community_chat_handler.go`: added HTTP handlers, JWT WebSocket auth, local image upload, and upload serving.
- `backend/internal/server/routes/community_chat.go`: registered community chat HTTP and WebSocket routes.
- `backend/internal/server/router.go`: mounted community chat routes under `/api/v1`.
- `backend/internal/handler/handler.go`: added the community chat handler to the handler aggregate.
- `backend/internal/handler/wire.go`: wired the community chat handler provider.
- `backend/internal/repository/wire.go`: wired the community chat repository provider.
- `backend/internal/service/wire.go`: wired the community chat service provider.
- `backend/cmd/server/wire_gen.go`: updated generated dependency wiring for the new chat components.
- `frontend/src/api/communityChat.ts`: added typed API helpers and WebSocket URL construction.
- `frontend/src/api/index.ts`: exported community chat API types and helpers.
- `frontend/src/router/index.ts`: added the authenticated `/community-chat` route.
- `frontend/src/components/layout/AppSidebar.vue`: added the “用户群聊” menu entry.
- `frontend/src/views/user/CommunityChatView.vue`: added the chat UI, realtime handling, emoji shortcuts, image upload, and admin delete action.
- `frontend/src/i18n/locales/zh.ts`: added Chinese navigation and page copy.
- `frontend/src/i18n/locales/en.ts`: added English navigation and page copy.
- `docs/COMMUNITY_CHAT.md`: documented usage, persistence, APIs, and deployment considerations.
- Rollback: revert the files listed above and remove migration `154_community_chat.sql`; if the migration has already run, drop `community_chat_messages` and delete `<DATA_DIR>/chat/uploads` only if historical chat images are no longer needed.

## 2026-06-27 - Task: Improve community chat replies, recall, unread badge, and visual contrast

### What was done
- Added quote-style replies so users can reply to other users' chat messages without introducing threaded rooms.
- Changed chat deletion behavior so users can recall their own messages while admins can still delete any message.
- Added a red unread dot on the left-sidebar “用户群聊” entry when another user sends a new group chat message while the current user is outside the group chat page.
- Adjusted the group chat message area and own-message bubble colors so the chat background and message content are easier to distinguish.
- Improved fallback avatars to use one or two characters from the user's display name when no profile avatar exists.

### Testing
- `go test ./internal/service ./internal/repository ./internal/handler ./internal/server/routes` passed.
- `pnpm build` passed; only the existing Browserslist data warning appeared.
- `git diff --check` passed.

### Notes
- `backend/migrations/155_community_chat_reply.sql`: added reply snapshot columns and an index without modifying the already-created `154_community_chat.sql` migration.
- `backend/internal/service/community_chat.go`: added reply snapshot creation, display-name fallback, reply preview truncation, and owner-or-admin recall authorization.
- `backend/internal/repository/community_chat_repo.go`: persisted and read reply fields for list, create, get, and delete flows.
- `backend/internal/handler/community_chat_handler.go`: accepted `reply_to_message_id` for text and image messages.
- `frontend/src/api/communityChat.ts`: added reply fields and `reply_to_message_id` support to text/image send helpers.
- `frontend/src/views/user/CommunityChatView.vue`: added reply controls, reply preview, own-message recall, fallback avatar initials, and improved chat colors.
- `frontend/src/components/layout/AppSidebar.vue`: added the “用户群聊” unread red dot and lightweight WebSocket listener for other-user messages.
- `frontend/src/i18n/locales/zh.ts`: added Chinese copy for reply, recall, image-message labels, and unread-related UI.
- `frontend/src/i18n/locales/en.ts`: added English copy for reply, recall, image-message labels, and unread-related UI.
- `docs/COMMUNITY_CHAT.md`: documented replies, recall behavior, unread badge behavior, and reply request payloads.
- Rollback: revert the files listed above and remove migration `155_community_chat_reply.sql`; if migration 155 has already run, create a forward rollback migration to drop the reply columns/index rather than editing applied migrations.

## 2026-06-27 - Task: Relax community chat image upload limit

### What was done
- Increased the community chat image upload limit from 5 MB to 15 MB.
- Kept frontend pre-upload validation, backend validation, error copy, and documentation aligned on the same 15 MB limit.
- Added a hard multipart request-body limit around chat image upload so oversized requests are rejected before consuming unnecessary temporary storage.

### Testing
- `go test ./internal/service ./internal/handler` passed.
- `pnpm build` passed; only the existing Browserslist data warning appeared.
- `git diff --check` passed for the changed community chat files.

### Notes
- `backend/internal/service/community_chat.go`: changed `CommunityChatMaxImageBytes` to 15 MB and updated the backend error text.
- `backend/internal/handler/community_chat_handler.go`: added `http.MaxBytesReader` for chat image upload and aligned multipart parsing with the new image limit.
- `frontend/src/views/user/CommunityChatView.vue`: updated client-side image size validation to 15 MB.
- `frontend/src/i18n/locales/zh.ts`: updated Chinese upload limit copy to 15 MB.
- `frontend/src/i18n/locales/en.ts`: updated English upload limit copy to 15 MB.
- `docs/COMMUNITY_CHAT.md`: documented the new 15 MB image upload limit.
- Rollback: revert the files listed above, or change the same limit constants and copy back to 5 MB if product policy needs the old cap.

## 2026-06-27 - Task: Rename chat sidebar entry

### What was done
- Renamed the original Chinese left-sidebar “聊天” entry to “模型问答”.
- Kept the existing `/chat` route, chat page behavior, and English label unchanged.

### Testing
- `pnpm exec vue-tsc -b` passed.
- `git diff --check` passed for the changed locale and progress files.

### Notes
- `frontend/src/i18n/locales/zh.ts`: updated `nav.chat` from “聊天” to “模型问答”.
- Rollback: revert this file change, or change `nav.chat` back to “聊天”.

## 2026-06-28 - Task: Add user API documentation page

### What was done
- Added a left-sidebar “接口文档” entry for users and admins' personal menu.
- Added a new `/api-docs` page documenting `gpt-image-2` and `nano-banana-pro` image API usage.
- Included supported parameters, default `api_url` of `https://api.iotwq.top`, and copy-ready curl/Python examples based on `../gpt_image_playground/`.
- Documented the maintenance source and scope of the new static API docs page.

### Testing
- `pnpm exec vue-tsc -b` passed.
- `pnpm build` passed; only the existing Browserslist data warning appeared.
- `git diff --check` passed for the changed API docs files.

### Notes
- `.gitignore`: allowed `docs/API_DOCS.md` to be version-controlled despite the existing `docs/*` ignore rule.
- `docs/API_DOCS.md`: documented the user-facing API docs page, default API URL, reference source, and maintenance notes.
- `frontend/src/views/user/ApiDocsView.vue`: added the static API docs UI, parameter tables, and copyable curl/Python examples.
- `frontend/src/router/index.ts`: registered the authenticated `/api-docs` route.
- `frontend/src/components/layout/AppSidebar.vue`: added the “接口文档” sidebar item with a document icon.
- `frontend/src/i18n/locales/zh.ts`: added Chinese navigation and page title/description copy.
- `frontend/src/i18n/locales/en.ts`: added English navigation and page title/description copy.
- Rollback: revert the files listed above; no database or backend migration rollback is needed because the page is frontend-only.

## 2026-06-28 - Task: Add video generation API docs

### What was done
- Extended the user-side “接口文档” page with video generation docs for `video-ds-2.0-fast` and `video-ds-2.0`.
- Added the async video task flow: create task, poll task status, and stream or download mp4 content.
- Added copy-ready curl, JavaScript, and Python examples using the default API URL `https://api.iotwq.top`.
- Updated the static docs and page description so the interface docs clearly cover both image and video generation APIs.

### Testing
- `pnpm exec vue-tsc -b` passed.
- `pnpm build` passed; only the existing Browserslist data warning appeared.
- `git diff --check` passed for the changed API docs files.

### Notes
- `frontend/src/views/user/ApiDocsView.vue`: added the video generation model section, parameters, endpoints, and curl/JavaScript/Python examples.
- `frontend/src/i18n/locales/zh.ts`: updated the Chinese API docs description to include image and video generation examples.
- `frontend/src/i18n/locales/en.ts`: updated the English API docs description to include image and video generation examples.
- `docs/API_DOCS.md`: documented `video-ds-2.0-fast` and `video-ds-2.0` endpoints and maintenance scope.
- `progress.md`: recorded this video API documentation update and validation evidence.
- Rollback: revert the files listed above; this is a frontend documentation-only change with no backend or database migration rollback.

## 2026-06-28 - Task: Refine chat copy and add Sora 2 API docs

### What was done
- Removed the community chat composer hint text about supported text, emoji, and 15 MB images while keeping the upload behavior unchanged.
- Removed the unwanted source-reference wording from the `nano-banana-pro` API documentation description.
- Added a Sora 2 section to the user-side “接口文档” page with available models, async task flow, parameters, and copy-ready curl/Python examples using `https://api.iotwq.top`.
- Rebuilt `iotwq/china-api:latest` and restarted the local compose stack so the running `http://127.0.0.1:8080` service serves the updated page.

### Testing
- `pnpm exec vue-tsc -b` passed.
- `pnpm build` passed; only the existing Browserslist data warning appeared.
- `git diff --check` passed for the changed API docs and chat files.
- `docker build -t iotwq/china-api:latest .` passed.
- `docker compose up -d --force-recreate` passed; `sub2api`, `postgres`, and `redis` are healthy.
- `curl http://127.0.0.1:8080/health` returned 200, and `curl http://127.0.0.1:8080/` returned 200.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: removed the composer hint text area and kept only the character counter.
- `frontend/src/i18n/locales/zh.ts`: removed the unused Chinese community chat composer hint copy.
- `frontend/src/i18n/locales/en.ts`: removed the unused English community chat composer hint copy.
- `frontend/src/views/user/ApiDocsView.vue`: removed the nano-banana source wording and added Sora 2 model docs, available-model table, and curl/Python examples.
- `docs/API_DOCS.md`: updated the API docs scope to include Sora 2 video models.
- `progress.md`: recorded this UI copy and Sora 2 documentation update.
- Rollback: revert the files listed above, then rebuild `iotwq/china-api:latest` and rerun `docker compose up -d --force-recreate`; no backend or database migration rollback is needed.

## 2026-06-28 - Task: Clean user-facing API docs wording

### What was done
- Removed the Sora 2 note that explained the sample `API_URL` had been replaced.
- Simplified the top API docs note to only state the required `Authorization` header.
- Reworded model notes and parameter descriptions so the page reads as user-facing API documentation instead of implementation notes.
- Cleaned the static API docs maintenance file to describe page coverage without source-reference wording.
- Rebuilt `iotwq/china-api:latest` and restarted the local compose stack so `http://127.0.0.1:8080` serves the updated page.

### Testing
- `pnpm exec vue-tsc -b` passed.
- `pnpm build` passed; only the existing Browserslist data warning appeared.
- `git diff --check` passed for the changed API docs files.
- User-facing docs scan found no remaining matches for the removed internal wording in `frontend/src/views/user/ApiDocsView.vue` and `docs/API_DOCS.md`.
- `docker build -t iotwq/china-api:latest .` passed.
- `docker compose up -d --force-recreate` passed; `sub2api`, `postgres`, and `redis` are healthy.
- `curl http://127.0.0.1:8080/health` returned 200, and `curl http://127.0.0.1:8080/api-docs` returned 200.

### Notes
- `frontend/src/views/user/ApiDocsView.vue`: removed owner-facing wording and tightened API docs notes for end users.
- `docs/API_DOCS.md`: removed source/maintenance wording and kept only page coverage details.
- `progress.md`: recorded this user-facing API docs cleanup and verification evidence.
- Rollback: revert the files listed above, then rebuild `iotwq/china-api:latest` and rerun `docker compose up -d --force-recreate`; no backend or database migration rollback is needed.

## 2026-06-28 - Task: Improve community chat visual design

### What was done
- Redesigned the community chat message area so bubbles no longer stretch across the full row.
- Added left/right message alignment, softer dark-mode contrast, subtle message-area texture, lighter avatars, and cleaner bubble shadows.
- Reduced visual noise by making reply/delete actions visible on mobile and hover-revealed on desktop.
- Refined the composer into a rounded chat input bar with cleaner emoji buttons and upload/send controls.
- Rebuilt `iotwq/china-api:latest` and restarted the local compose stack so `http://127.0.0.1:8080` serves the updated chat page.

### Testing
- `pnpm exec vue-tsc -b` passed.
- `pnpm build` passed; only the existing Browserslist data warning appeared.
- `git diff --check` passed for the changed community chat file.
- `docker build -t iotwq/china-api:latest .` passed after retrying a Docker Hub metadata EOF.
- `docker compose up -d --force-recreate` passed; `sub2api`, `postgres`, and `redis` are healthy.
- `curl http://127.0.0.1:8080/health` returned 200, and `curl http://127.0.0.1:8080/community-chat` returned 200.
- Browser visual check reached the app but redirected to login because no authenticated browser session was available in the automation tab.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: updated only presentation classes and scoped styles for the chat layout, bubbles, message area, actions, and composer.
- `progress.md`: recorded this visual redesign and verification evidence.
- Rollback: revert the files listed above, then rebuild `iotwq/china-api:latest` and rerun `docker compose up -d --force-recreate`; no backend or database migration rollback is needed.

## 2026-06-28 - Task: Reduce community chat background brightness

### What was done
- Replaced the large white surfaces in the community chat light theme with lower-brightness warm gray tones.
- Darkened the message area, panel, header, composer, emoji buttons, image preview, input row, textarea, and message bubbles to reduce glare.
- Changed own-message bubbles to a deep red tone with white text so messages remain distinct without a bright white background.
- Rebuilt `iotwq/china-api:latest` and restarted the local compose stack so `http://127.0.0.1:8080` serves the updated chat page.

### Testing
- `pnpm exec vue-tsc -b` passed.
- `pnpm build` passed; only the existing Browserslist data warning appeared.
- `git diff --check` passed for the changed community chat file.
- `docker build -t iotwq/china-api:latest .` passed.
- `docker compose up -d --force-recreate` passed; `sub2api`, `postgres`, and `redis` are healthy.
- `curl http://127.0.0.1:8080/health` returned 200, and `curl http://127.0.0.1:8080/community-chat` returned 200.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: lowered the light-theme brightness across the group chat page while keeping the existing layout and behavior.
- `progress.md`: recorded this brightness reduction and verification evidence.
- Rollback: revert the files listed above, then rebuild `iotwq/china-api:latest` and rerun `docker compose up -d --force-recreate`; no backend or database migration rollback is needed.

## 2026-06-28 - Task: Further darken community chat background

### What was done
- Further reduced the light-theme brightness of the community chat page after the background still felt too white and glaring.
- Darkened the outer page, chat panel, header, message area, normal message bubbles, composer, emoji buttons, image preview, input row, and textarea.
- Kept the existing chat behavior unchanged, including messages, image upload, replies, recall, unread state, and WebSocket handling.
- Rebuilt `iotwq/china-api:latest` and restarted the local compose stack so `http://127.0.0.1:8080` serves the updated chat page.

### Testing
- `git diff --check -- frontend/src/views/user/CommunityChatView.vue progress.md` passed.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; only the existing Browserslist data warning appeared.
- `docker build -t iotwq/china-api:latest .` passed.
- `docker compose up -d --force-recreate` passed; `sub2api`, `postgres`, and `redis` are healthy.
- `curl http://127.0.0.1:8080/health` returned 200, and `curl http://127.0.0.1:8080/community-chat` returned 200.
- `docker image inspect iotwq/china-api:latest` reports `sha256:f18bb12bf057c1634fead699b792338b9455a09df9b11ac7b34d0502e9595f83 arm64`.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: further darkened the community chat light-theme surfaces while preserving the existing layout and behavior.
- `progress.md`: recorded this glare-reduction pass and verification evidence.
- Rollback: revert the files listed above, then rebuild `iotwq/china-api:latest` and rerun `docker compose up -d --force-recreate`; no backend or database migration rollback is needed.

## 2026-06-28 - Task: Keep community chat message actions visible

### What was done
- Changed community chat reply and recall/delete controls from hover-revealed controls to always-visible controls.
- Kept the existing permissions unchanged: users can reply to visible messages, users can recall their own messages, and admins can delete any message.
- Rebuilt `iotwq/china-api:latest` and restarted the local compose stack so `http://127.0.0.1:8080` serves the updated chat page.

### Testing
- `git diff --check -- frontend/src/views/user/CommunityChatView.vue progress.md` passed.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; only the existing Browserslist data warning appeared.
- `docker build -t iotwq/china-api:latest .` passed.
- `docker compose up -d --force-recreate` passed; `sub2api`, `postgres`, and `redis` are healthy.
- `curl http://127.0.0.1:8080/health` returned 200, and `curl http://127.0.0.1:8080/community-chat` returned 200.
- `docker image inspect iotwq/china-api:latest` reports `sha256:1ebb479ed945cbd54650e891705061c72cb37ed283b8eb3c199aa40f82d6c143 arm64`.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: made reply and recall/delete action buttons always visible for visible chat messages.
- `progress.md`: recorded this interaction fix and verification evidence.
- Rollback: revert the files listed above, then rebuild `iotwq/china-api:latest` and rerun `docker compose up -d --force-recreate`; no backend or database migration rollback is needed.

## 2026-06-28 - Task: Move community chat reply action to hover

### What was done
- Confirmed community chat delete visibility and permission rules remain limited to the message owner and admins.
- Moved the reply action out of the message metadata row so it no longer appears directly beside the username and time.
- Changed the reply action to appear when hovering over a message on pointer devices, while keeping it visible on touch devices where hover is unavailable.
- Rebuilt `iotwq/china-api:latest` and restarted the local compose stack so `http://127.0.0.1:8080` serves the updated chat page.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; only the existing Browserslist data warning appeared.
- `docker build -t iotwq/china-api:latest .` passed.
- `docker compose up -d --force-recreate` passed; `sub2api`, `postgres`, and `redis` are healthy.
- `curl http://127.0.0.1:8080/health` returned 200, and `curl http://127.0.0.1:8080/community-chat` returned 200.
- `docker image inspect iotwq/china-api:latest` reports `sha256:b609f38fe8a168db67186106e9f16911fef2f7491ccf06fc6109fd9674560792 arm64`.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: moved reply controls below the message bubble and made them hover-revealed on pointer devices.
- `progress.md`: recorded this reply-action interaction adjustment and verification evidence.
- Rollback: revert the files listed above, then rebuild `iotwq/china-api:latest` and rerun `docker compose up -d --force-recreate`; no backend or database migration rollback is needed.

## 2026-06-28 - Task: Refresh homepage with Three.js AI visual

### What was done
- Added a scoped Three.js homepage visual for the default `/home` page, focused on the right-side AI learning scene.
- Reworked the homepage hero and lower content area into a darker red, future-facing AI learning style while keeping the existing site name, logo, login, dashboard, language, theme, and docs actions.
- Added a compatibility fallback so old API conversion default subtitle text is shown as the AI learning subtitle on the homepage.
- Added a short maintenance note for the homepage Three.js scene.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; only the existing Browserslist data warning and existing large chunk warning appeared.
- `git diff --check -- frontend/src/views/HomeView.vue frontend/src/components/home/HomeThreeScene.vue frontend/package.json frontend/pnpm-lock.yaml` passed.
- Opened `http://127.0.0.1:3000/home` in the in-app browser and confirmed the page renders, one `.home-three-scene` canvas exists, and the subtitle displays `AI技术 学习交流`.
- Checked desktop `1440x900` and mobile `390x844` browser viewports; both had no horizontal overflow.

### Notes
- `frontend/src/components/home/HomeThreeScene.vue`: added the responsive Three.js AI learning visual with animation cleanup.
- `frontend/src/views/HomeView.vue`: integrated the Three.js visual and refreshed the hero, lower signal cards, and topic console styling.
- `frontend/package.json`: added `three` and `@types/three`.
- `frontend/pnpm-lock.yaml`: recorded the new Three.js dependencies.
- `docs/HOME_THREE_SCENE.md`: documented where the scene lives and how to verify it.
- `.gitignore`: allowed the homepage Three.js maintenance note to be tracked under `docs/`.
- `progress.md`: recorded this homepage visual refresh and verification evidence.
- Rollback: revert the files listed above, run `pnpm install` in `frontend/` to restore dependencies, then rebuild the frontend; no backend or database rollback is needed.

## 2026-07-02 - Task: Fix Claude channel monitor response parsing

### What was done
- Fixed channel monitor probe parsing so Anthropic/Claude checks can read valid answer text from compatible OpenAI chat responses, Responses API responses, and Anthropic responses where the first content block is not text.
- Kept the existing probe request body, retry behavior, latency thresholds, and real request routing unchanged.
- Added regression coverage for the Claude-compatible response shapes that previously produced `challenge mismatch ... got ""`.

### Testing
- `go test -tags unit ./internal/service -run 'TestRunCheckForModel_(AnthropicProvider|OpenAIResponses|OpenAI_Default|Retries|AllProbe|Replace|Merge|OffMode)'` passed from `backend/`.
- `go test -tags unit ./internal/service` passed from `backend/`.
- `git diff --check -- backend/internal/service/channel_monitor_checker.go backend/internal/service/channel_monitor_checker_body_test.go progress.md` passed.

### Notes
- `backend/internal/service/channel_monitor_checker.go`: added fallback response-text extraction for channel monitor probes without changing provider request construction.
- `backend/internal/service/channel_monitor_checker_body_test.go`: added fake upstream response formats and regression tests for Claude-compatible monitor responses.
- `progress.md`: recorded this channel monitor parsing fix and verification evidence.
- Rollback: revert the files listed above; no frontend, database, configuration, or deployment rollback is needed.

## 2026-07-02 - Task: Sync latest upstream sub2api changes

### What was done
- Fetched `Wei-Shaw/sub2api` and merged `upstream/main` from `c2754222` to `0b8e5eec` (`v0.1.143`).
- Created a pre-sync backup branch and preserved the current local worktree in a stash before merging.
- Restored local modifications after the upstream merge and resolved conflicts while keeping local nano-banana, homepage, community chat, API docs, image bridge, and channel-monitor changes.
- Documented the sync scope, conflict handling, verification, and rollback points.

### Testing
- `go test ./internal/service ./internal/server/routes ./internal/handler` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm exec vitest run src/components/admin/account/__tests__/AccountTestModal.spec.ts src/views/admin/__tests__/AccountsView.sparkShadow.spec.ts` passed from `frontend/`; only the existing localstorage-file and Browserslist warnings appeared.
- `pnpm build` passed from `frontend/`; only the existing Browserslist and Vite large chunk warnings appeared.
- `git diff --check` passed.
- `git diff --cached --check` passed.

### Notes
- `README.md`: resolved sponsor-list merge conflict by keeping local sponsor blocks and upstream additions.
- `README_CN.md`: resolved Chinese sponsor-list merge conflict by keeping local sponsor blocks and upstream additions.
- `README_JA.md`: resolved Japanese sponsor-list merge conflict by keeping local sponsor blocks and upstream additions.
- `backend/internal/server/routes/gateway.go`: resolved route conflict by keeping upstream Grok image/video handlers and local nano-banana route.
- `frontend/src/components/admin/account/__tests__/AccountTestModal.spec.ts`: resolved test helper conflict for both visible-state and custom-account test cases.
- `frontend/src/views/admin/AccountsView.vue`: resolved modal/action-menu conflict so Spark shadow handling and local lazy modal mounting both remain available.
- `docs/UPSTREAM_SYNC.md`: recorded the upstream range, merged features, conflict handling, test evidence, and rollback points.
- `progress.md`: recorded this upstream sync and verification evidence.
- Rollback: use backup branch `backup/pre-upstream-sync-20260702-225343`, stash `pre-upstream-sync-20260702-225343`, or revert merge commit `290dff33`; if local restored worktree changes need to be undone separately, compare against the same stash before making any destructive changes.

## 2026-07-02 - Task: Record post-sync checkpoint

### What was done
- Recorded the current checkpoint after syncing `Wei-Shaw/sub2api` upstream changes through `0b8e5eec` (`v0.1.143`).
- Confirmed the upstream merge commit is `290dff33` and the local feature worktree was restored after the merge.
- Noted that the remaining modified and untracked files are the preserved local project changes, not unresolved merge conflicts.

### Testing
- `tail -n 50 progress.md` confirmed the upstream sync record was already present before this checkpoint.
- `git diff --check -- progress.md` passed.

### Notes
- `progress.md`: appended this post-sync checkpoint for continuity before any later image build, deployment, or follow-up fix.
- Rollback: remove this checkpoint entry from `progress.md`; no code, database, frontend asset, or deployment rollback is needed.

## 2026-07-05 - Task: Sync latest upstream sub2api changes

### What was done
- Fetched `Wei-Shaw/sub2api` and merged `upstream/main` from `0b8e5eec` to `b650bdd6` (`v0.1.144`).
- Created backup branch `backup/pre-upstream-sync-20260705-184318` and preserved the current local worktree in stash `pre-upstream-sync-20260705-184318` before merging.
- Restored local modifications after the upstream merge and resolved Codex/OpenAI image bridge conflicts while preserving local nano-banana, community chat, API docs, homepage, image bridge, and channel monitor changes.
- Documented the sync scope, conflict handling, verification, and rollback points.

### Testing
- `go test ./internal/service -run 'TestOpenAIGatewayServiceForward_(AccountPolicyStripsExplicitImageTool|NormalizesDanglingImageToolChoice|CodexBridgeSkipsCustomFunctionTools|ExplicitImageToolWorksWithBridgeDisabled|CodexBridgeCanBeDisabledByRequestHeader)'` passed from `backend/`.
- `go test ./internal/service ./internal/server/routes ./internal/handler` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm exec vitest run src/components/admin/account/__tests__/AccountTestModal.spec.ts src/views/user/__tests__/ChatView.spec.ts src/stores/__tests__/app.spec.ts` passed from `frontend/`; only the existing localstorage-file and Browserslist warnings appeared.
- `pnpm build` passed from `frontend/`; only the existing Browserslist and Vite large chunk warnings appeared.
- `git diff --check` and `git diff --cached --check` passed.

### Notes
- `backend/internal/service/openai_codex_transform.go`: kept upstream Spark image tool stripping and restored the local custom-function-tool bridge guard.
- `backend/internal/service/openai_gateway_service.go`: combined upstream Codex explicit image tool policy with the local request-level image bridge disable header.
- `backend/internal/service/openai_ws_forwarder.go`: applied the same Codex image bridge policy combination to WebSocket ingress payloads.
- `backend/internal/service/openai_image_generation_controls_test.go`: kept regression coverage for policy stripping, dangling `tool_choice` normalization, and custom function tool behavior.
- `docs/UPSTREAM_SYNC.md`: recorded the upstream range, merged features, conflict handling, verification, and rollback points.
- `progress.md`: recorded this upstream sync and verification evidence.
- Rollback: use backup branch `backup/pre-upstream-sync-20260705-184318`, stash `pre-upstream-sync-20260705-184318`, or revert merge commit `e8b64e29`; compare local restored worktree changes against the same stash before any destructive cleanup.

## 2026-07-05 - Task: Add async OpenAI-compatible video gateway

### What was done
- Added OpenAI-compatible async video task routing for `POST /v1/videos`, `GET /v1/videos/{task_id}`, and `GET /v1/videos/{task_id}/content`.
- Forwarded `video-ds-2.0-fast` and `video-ds-2.0` requests through OpenAI API key accounts, including model mapping, upstream failover on task creation, status polling, range-aware mp4 streaming, and task-to-account sticky binding.
- Kept existing Grok `/v1/videos/generations` behavior separate from the new OpenAI video task API.
- Left separate video usage billing out of this change to avoid incorrectly charging video tasks as image or token usage.

### Testing
- `go test ./internal/service ./internal/server/routes ./internal/handler` passed from `backend/`.
- `git diff --check` passed.
- `git diff --cached --check` passed.

### Notes
- `backend/internal/service/openai_videos.go`: added request parsing, model validation, upstream forwarding, sticky task binding helpers, JSON response forwarding, and video content streaming.
- `backend/internal/handler/openai_videos.go`: added authenticated create/status/content handlers with scheduling, moderation for create requests, failover on create, and task sticky binding.
- `backend/internal/server/routes/gateway.go`: registered OpenAI `/videos` task routes while preserving Grok generation routes.
- `backend/internal/handler/endpoint.go`: normalized bare `/videos` requests to the canonical video endpoint.
- `backend/internal/handler/grok_media.go`: allowed Grok status handlers to read either `request_id` or `task_id` route parameters.
- `backend/internal/service/openai_videos_test.go`: added service tests for create, status, content streaming, headers, and task ID extraction.
- `backend/internal/server/routes/gateway_test.go`: added route coverage for OpenAI video task paths and legacy Grok video path rejection on OpenAI groups.
- `progress.md`: recorded this video gateway implementation and verification evidence.
- Rollback: remove the new video service and handler files, revert the route/endpoint/Grok parameter changes and tests listed above, then rerun `go test ./internal/service ./internal/server/routes ./internal/handler`.

## 2026-07-05 - Task: Persist OpenAI video task account bindings

### What was done
- Added per-user persistent storage for OpenAI-compatible video `task_id` to upstream account bindings, so status polling and mp4 downloads can recover the original upstream account after cache expiry or service restart without sharing bindings across users.
- Kept Redis sticky-session caching as the fast path and extended video task cache bindings to a dedicated 7-day TTL.
- Restored the persistent binding before video status/content routing, then reused the existing account scheduler and health checks instead of bypassing routing safeguards.
- Documented the video task routing behavior for API documentation maintenance.

### Testing
- `go test ./internal/handler` passed from `backend/` after refreshing stale test constructor arguments.
- `go test ./internal/service ./internal/repository ./internal/server/routes ./internal/handler ./cmd/server` passed from `backend/`.
- `go test ./internal/service -run 'Test(BindOpenAIVideoTaskAccount|RestoreOpenAIVideoTaskStickySession|ForwardOpenAIVideo|ParseOpenAIVideo)' -count=1` passed from `backend/`.
- `git diff --check` passed.
- `git diff --cached --check` passed.

### Notes
- `backend/migrations/159_openai_video_task_bindings.sql`: added the user-scoped video task binding table and indexes.
- `backend/internal/service/openai_videos.go`: added the binding repository port, persistent bind/restore helpers, and video-specific sticky TTL.
- `backend/internal/repository/openai_video_task_binding_repo.go`: added SQL persistence for video task bindings.
- `backend/internal/handler/openai_videos.go`: restores task binding before status and content routing.
- `backend/internal/service/openai_videos_test.go`: added tests for persistent bind and cache restore behavior.
- `backend/internal/service/openai_gateway_service.go`: injected the video binding repository into the OpenAI gateway service.
- `backend/internal/repository/wire.go` and `backend/cmd/server/wire_gen.go`: wired the new repository into application startup.
- `backend/internal/handler/openai_gateway_handler_test.go`, `backend/internal/handler/openai_images_failover_test.go`, `backend/internal/service/openai_gateway_record_usage_test.go`, and `backend/internal/service/openai_ws_protocol_forward_test.go`: updated test constructors for the new dependency.
- `docs/API_DOCS.md`: documented that video status/download requests are routed back to the current user's original upstream account by `task_id`.
- `progress.md`: recorded this persistent video binding implementation and verification evidence.
- Rollback: revert the files listed above; if migration `159_openai_video_task_bindings.sql` has already been applied, drop `openai_video_task_bindings` in a new forward migration before redeploying the reverted code.

## 2026-07-05 - Task: Bill async video tasks per successful creation

### What was done
- Added video task usage billing for `POST /v1/videos`: after upstream creation succeeds and returns `task_id`, the service records one billable per-request usage item.
- Used a stable video usage request ID derived from `task_id`, so repeated usage-record attempts for the same task do not double charge.
- Kept `GET /v1/videos/{task_id}` status polling and `GET /v1/videos/{task_id}/content` downloads non-billable.
- Reused the existing channel pricing page `按次` mode and added a short hint that video tasks are charged once after creation succeeds.
- Documented the video per-request billing behavior for API docs maintenance.

### Testing
- `go test ./internal/service -run 'TestOpenAIGatewayServiceRecordUsage_Channel(PerRequestVideo|ImageBilling)|TestOpenAIVideo|TestParseOpenAIVideo|TestBindOpenAIVideo|TestRestoreOpenAIVideo|TestForwardOpenAIVideo' -count=1` passed from `backend/`.
- `go test ./internal/service ./internal/repository ./internal/server/routes ./internal/handler ./cmd/server` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; only the existing Browserslist stale-data notice and Vite large chunk warning appeared.

### Notes
- `backend/internal/handler/openai_videos.go`: records mandatory video usage after successful task creation, with `MediaType=video`, `RequestCount=1`, and `task_id`-based billing request ID.
- `backend/internal/service/openai_videos.go`: added the stable video usage request ID helper with a length-safe hash fallback.
- `backend/internal/service/openai_gateway_service.go`: added per-request OpenAI cost calculation for video-style usage and allows selected results to use their own billing request ID.
- `backend/internal/service/gateway_service.go`: carries media type into the usage billing command fingerprint.
- `backend/internal/service/openai_gateway_record_usage_test.go`: added regression coverage for video per-request pricing and task ID dedupe behavior.
- `frontend/src/components/admin/channel/PricingEntryCard.vue`: shows a short video billing hint under the per-request price field.
- `frontend/src/i18n/locales/en.ts` and `frontend/src/i18n/locales/zh.ts`: added the per-request video billing hint text.
- `docs/API_DOCS.md`: documented video model per-request billing and non-billable polling/download behavior.
- `progress.md`: recorded this billing implementation and verification evidence.
- Rollback: revert the files listed above; video task routing and persistent task bindings can remain, but `POST /v1/videos` will no longer produce a per-request billing record after the rollback.

## 2026-07-05 - Task: Enable Sora 2 models on OpenAI-compatible video gateway

### What was done
- Added Sora 2 video models to the OpenAI-compatible `/v1/videos` model allowlist: `sora-2-landscape-8s`, `sora-2-landscape-12s`, `sora-2-portrait-8s`, and `sora-2-portrait-12s`.
- Kept the existing OpenAI API key channel path, async task binding, polling, and per-request billing behavior unchanged.
- Updated video API documentation to list the complete supported video model set.

### Testing
- `go test ./internal/service -run 'Test(ParseOpenAIVideoCreateRequest|ForwardOpenAIVideoCreate|OpenAIVideo)' -count=1` passed from `backend/`.
- `go test ./internal/service` passed from `backend/`.

### Notes
- `backend/internal/service/openai_videos.go`: expanded video model validation and unsupported-model error text to include Sora 2 models.
- `backend/internal/service/openai_videos_test.go`: added Sora 2 parsing and forwarding coverage.
- `docs/API_DOCS.md`: documented the full supported video model list for the video gateway.
- `progress.md`: recorded this Sora 2 video gateway enablement and verification evidence.
- Rollback: revert the files listed above; Sora 2 models will again be rejected by `/v1/videos`, while existing `video-ds-2.0-fast` and `video-ds-2.0` behavior remains available.

## 2026-07-05 - Task: Show timestamps only for new community chat messages

### What was done
- Added a nullable `sent_at` field for community chat messages and write it only when new text or image messages are created.
- Returned `sent_at` through the existing message list and WebSocket payloads.
- Updated the user chat page to display a message time only when `sent_at` is present, so existing historical messages are not backfilled or shown with old `created_at` time.
- Documented that chat send time is recorded for new messages only.

### Testing
- `go test ./internal/service ./internal/repository ./internal/handler` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.

### Notes
- `backend/migrations/160_community_chat_sent_at.sql`: added nullable `sent_at` without backfilling existing rows.
- `backend/internal/service/community_chat.go`: exposed optional `sent_at` on community chat messages.
- `backend/internal/repository/community_chat_repo.go`: writes `sent_at=NOW()` for new messages and scans nullable times for list/get/delete responses.
- `frontend/src/api/communityChat.ts`: added optional `sent_at` to the message type.
- `frontend/src/views/user/CommunityChatView.vue`: displays message time from `sent_at` only when present.
- `docs/COMMUNITY_CHAT.md`: documented new-message-only send time behavior.
- `progress.md`: recorded this timestamp display change and verification evidence.
- Rollback: revert the files listed above; if migration `160_community_chat_sent_at.sql` has already been applied, leave the nullable column unused or remove it with a forward migration.

## 2026-07-05 - Task: Make community chat timestamps visible

### What was done
- Changed the community chat message time from a faint inline gray label to a higher-contrast timestamp badge.
- Kept the existing `sent_at` behavior unchanged: only messages that already have `sent_at` render a send time.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check` passed from the repository root.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: updated the message timestamp class and styling so send times are visible on the chat background and on self messages.
- `progress.md`: recorded this timestamp visibility fix and verification evidence.
- Rollback: revert the `frontend/src/views/user/CommunityChatView.vue` and `progress.md` changes from this task.

## 2026-07-05 - Task: Align video-ds API docs seconds parameter

### What was done
- Updated the user-facing `video-ds-2.0` / `video-ds-2.0-fast` API examples so `seconds` is shown as the string `"15"`.
- Updated the `seconds` parameter type in the interface documentation from `integer` to `string`.

### Testing
- `rg -n "seconds" frontend/src/views/user/ApiDocsView.vue docs/API_DOCS.md` confirmed the remaining video examples use string seconds.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check` passed from the repository root.

### Notes
- `frontend/src/views/user/ApiDocsView.vue`: changed curl, JavaScript, Python examples and the parameter table for the video-ds seconds field.
- `progress.md`: recorded this API documentation correction and verification evidence.
- Rollback: revert the `frontend/src/views/user/ApiDocsView.vue` and `progress.md` changes from this task.

## 2026-07-06 - Task: Replace fixed API docs URL with placeholder

### What was done
- Replaced the fixed API docs example base URL with `YOUR_API_URL`.
- Removed the header API URL badge from the user-facing interface documentation page.
- Updated the interface documentation maintenance note so examples are described as user-replaced service URLs rather than a fixed default address.

### Testing
- `rg -n "api.iotwq.top|默认 API 地址|api-docs-base-url|apiBaseUrl =|YOUR_API_URL" frontend/src/views/user/ApiDocsView.vue docs/API_DOCS.md` confirmed no fixed API URL or removed header badge code remains.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check` passed from the repository root.

### Notes
- `frontend/src/views/user/ApiDocsView.vue`: changed API examples to use `YOUR_API_URL`, removed the top-right API URL badge, and removed its unused styles.
- `docs/API_DOCS.md`: removed the fixed default API address and documented `YOUR_API_URL` as the placeholder.
- `progress.md`: recorded this API docs URL placeholder change and verification evidence.
- Rollback: revert the `frontend/src/views/user/ApiDocsView.vue`, `docs/API_DOCS.md`, and `progress.md` changes from this task.

## 2026-07-07 - Task: Fix OpenAI channel monitor model support false negatives

### What was done
- Aligned OpenAI passthrough account scheduling and model-availability diagnosis so passthrough accounts are treated as accepting any requested model.
- Added a channel monitor retry path for Codex/GPT-5.x OpenAI probes: if `/v1/chat/completions` returns the local gateway `model_not_found`, the probe retries once through `/v1/responses`.
- Added regression coverage for the monitor fallback and passthrough model support behavior.
- Documented the OpenAI channel monitor probe behavior for future operations.

### Testing
- `go test -tags=unit ./internal/service -run 'TestRunCheckForModel_OpenAICodexModelNotFoundRetriesResponses|TestOpenAIGatewayServiceDiagnoseModelAvailability_PassthroughAllowsAnyModel|TestOpenAIAccountScheduler_PassthroughAllowsUnmappedModel'` passed from `backend/`.
- `go test -tags=unit ./internal/service` passed from `backend/`.
- `git diff --check` passed from the repository root.

### Notes
- `backend/internal/service/channel_monitor_checker.go`: added OpenAI Codex/GPT-5.x monitor fallback from chat completions to Responses after local `model_not_found`.
- `backend/internal/service/channel_monitor_checker_body_test.go`: added coverage for the fallback probe path.
- `backend/internal/service/openai_gateway_model_availability.go`: made OpenAI model-availability diagnosis passthrough-aware.
- `backend/internal/service/openai_gateway_model_availability_test.go`: added passthrough availability and scheduler compatibility coverage.
- `backend/internal/service/openai_account_scheduler.go`: aligned scheduler model compatibility with OpenAI passthrough semantics.
- `backend/internal/service/openai_gateway_service.go`: aligned legacy OpenAI account eligibility with OpenAI passthrough semantics.
- `docs/UPSTREAM_SYNC.md`: documented OpenAI monitor protocol retry and passthrough model support behavior.
- `progress.md`: recorded this channel monitor stability fix and verification evidence.
- Rollback: revert the files listed above; monitors will again rely strictly on their configured OpenAI protocol and passthrough accounts will again be checked only by explicit `model_mapping`.

## 2026-07-07 - Task: Sync upstream sub2api to v0.1.146

### What was done
- Fetched `Wei-Shaw/sub2api` and merged `upstream/main` from `b650bdd6` to `f68f3b86`, creating merge commit `d04371cc`.
- Preserved the local modified features by stashing them before the merge, restoring them afterwards, and resolving the second-pass conflicts.
- Merged upstream batch image routes/menu/docs with local community chat, API docs, image bridge, video gateway, nano-banana, and channel monitor changes.
- Kept a pre-sync backup branch and stash for rollback.

### Testing
- `go test ./internal/service ./internal/server/routes ./internal/handler` passed from `backend/`.
- `go test -tags=unit ./internal/service -run 'TestRunCheckForModel_OpenAICodexModelNotFoundRetriesResponses|TestOpenAIGatewayServiceDiagnoseModelAvailability_PassthroughAllowsAnyModel|TestOpenAIAccountScheduler_PassthroughAllowsUnmappedModel'` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm exec vitest run src/__tests__/integration/data-import.spec.ts src/components/admin/account/__tests__/AccountTestModal.spec.ts src/views/user/__tests__/ChatView.spec.ts src/stores/__tests__/app.spec.ts` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; existing Browserslist and Vite large chunk warnings remain.
- `git diff --check` passed from the repository root.

### Notes
- `docs/UPSTREAM_SYNC.md`: documented the 2026-07-07 upstream sync scope, conflict handling, validation, and rollback points.
- `deploy/Dockerfile`: resolved upstream Docker build comment/memory conflict.
- `frontend/src/components/admin/account/ImportDataModal.vue`: merged upstream multi-file/drag-drop import with local ZIP and Codex session import support.
- `frontend/src/__tests__/integration/data-import.spec.ts`: merged upstream data import tests with local ZIP/Codex import tests.
- `backend/cmd/server/wire_gen.go`: resolved handler wiring so both community chat and batch image handlers are provided.
- `backend/internal/server/routes/gateway.go`: resolved gateway route conflict so batch image, video, Grok video, and nano-banana routes coexist.
- `frontend/src/components/layout/AppSidebar.vue`: resolved navigation conflict so local user entries and upstream batch image entry coexist.
- `frontend/src/router/index.ts`: resolved user route conflict so batch image, API docs, community chat, and image bridge routes coexist.
- `frontend/src/i18n/locales/en.ts`: merged navigation translation keys for local entries and upstream batch image.
- `frontend/src/i18n/locales/zh.ts`: merged navigation translation keys and kept the local “模型问答” label.
- `backend/internal/service/openai_gateway_model_availability_test.go`: added the unit build tag required by its existing test helper dependency after restoring local changes.
- `progress.md`: recorded this upstream sync and verification evidence.
- Rollback: reset or revert merge commit `d04371cc`, then re-apply the saved local stash if needed; the pre-sync branch is `backup/pre-upstream-sync-20260707-215250` and the saved stash is `stash@{0}` with message `pre-upstream-sync-20260707-215250`.

## 2026-07-07 - Task: Add site-owner private chat from community chat

### What was done
- Added a “联系站长” entry inside the community chat page so regular users can privately message the site owner/admin instead of posting publicly.
- Added an admin-side private conversation list in the same dialog so admins can select a user conversation and reply directly.
- Kept public group-chat unread reminders scoped to public group messages; private chat WebSocket events are filtered to the relevant user and admins only.
- Updated community chat documentation for private chat storage, endpoints, and notification boundaries.

### Testing
- `go test ./internal/service ./internal/repository ./internal/handler ./internal/server/routes` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; existing Browserslist data and Vite large chunk warnings remain.
- `git diff --check` passed from the repository root.

### Notes
- `backend/migrations/170_community_chat_direct_messages.sql`: added the private site-owner chat message table and indexes.
- `backend/internal/service/community_chat.go`: added direct chat service models, permissions, message creation, listing, and private WebSocket event metadata.
- `backend/internal/repository/community_chat_repo.go`: added PostgreSQL persistence for direct messages and admin conversation listing.
- `backend/internal/handler/community_chat_handler.go`: added direct chat HTTP handlers and filtered private WebSocket events by user/admin visibility.
- `backend/internal/server/routes/community_chat.go`: registered direct chat routes under `/api/v1/community-chat/direct`.
- `frontend/src/api/communityChat.ts`: added direct chat API methods, event type, and conversation typings.
- `frontend/src/api/index.ts`: exported the direct conversation type.
- `frontend/src/views/user/CommunityChatView.vue`: added the “联系站长” button, private chat dialog, admin conversation list, and matching styles.
- `frontend/src/i18n/locales/zh.ts`: added Chinese private chat labels and errors.
- `frontend/src/i18n/locales/en.ts`: added English private chat labels and errors.
- `docs/COMMUNITY_CHAT.md`: documented the site-owner private chat behavior, endpoints, and WebSocket visibility.
- `progress.md`: recorded this private chat implementation and verification evidence.
- Rollback: revert the files listed above; if migration `170_community_chat_direct_messages.sql` has already been applied, create a forward migration to drop `community_chat_direct_messages` and its indexes or leave the unused table in place.

## 2026-07-07 - Task: Add as-sd2.0-fast video model support

### What was done
- Added `as-sd2.0-fast` to the OpenAI-compatible async video model allowlist for `/v1/videos`.
- Kept model mapping supported so an existing request model such as `video-ds-2.0-fast` can be mapped upstream to `as-sd2.0-fast`.
- Listed `as-sd2.0-fast` as an available model in the user-facing API docs.

### Testing
- `go test ./internal/service -run 'TestParseOpenAIVideoCreateRequest|TestForwardOpenAIVideo'` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- backend/internal/service/openai_videos.go backend/internal/service/openai_videos_test.go frontend/src/views/user/ApiDocsView.vue docs/API_DOCS.md` passed from the repository root.

### Notes
- `backend/internal/service/openai_videos.go`: added `as-sd2.0-fast` to the accepted OpenAI video models.
- `backend/internal/service/openai_videos_test.go`: added parsing and upstream model-mapping coverage for `as-sd2.0-fast`.
- `frontend/src/views/user/ApiDocsView.vue`: added `as-sd2.0-fast` to the video API docs and available-model table.
- `docs/API_DOCS.md`: documented `as-sd2.0-fast` in the video route scope.
- `progress.md`: recorded this video model allowlist update and verification evidence.
- Rollback: revert the files listed above; `as-sd2.0-fast` will again be rejected by `/v1/videos`, while existing `video-ds-2.0-fast`, `video-ds-2.0`, and Sora 2 behavior remains available.

## 2026-07-08 - Task: Add YCYAPI video-v1-15s compatibility

### What was done
- Added `video-v1-15s` to the OpenAI-compatible async video model allowlist for `/v1/videos`.
- Added a YCYAPI-compatible upstream path branch: task creation and status polling use `/v1/video/generations`, while video content download keeps `/v1/videos/{task_id}/content`.
- Kept YCYAPI routing scoped to explicitly configured YCYAPI-compatible OpenAI API Key accounts, avoiding accidental routing to normal OpenAI video accounts.
- Added user-facing API docs for the 15 second YCYAPI model and its `prompt`, `ratio`, `image`, and `images` parameters.

### Testing
- `go test ./internal/service -run 'TestParseOpenAIVideoCreateRequest|TestForwardOpenAIVideo'` passed from `backend/`.
- `go test ./internal/service ./internal/handler -run 'TestParseOpenAIVideoCreateRequest|TestForwardOpenAIVideo|TestGateway|TestOpenAI'` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- backend/internal/service/openai_videos.go backend/internal/service/openai_videos_test.go backend/internal/handler/openai_videos.go frontend/src/views/user/ApiDocsView.vue docs/API_DOCS.md` passed from the repository root.

### Notes
- `backend/internal/service/openai_videos.go`: added `video-v1-15s`, YCYAPI account gating, and YCYAPI-specific create/status upstream URL selection.
- `backend/internal/service/openai_videos_test.go`: added parsing, create forwarding, and status polling coverage for YCYAPI `video-v1-15s`.
- `backend/internal/handler/openai_videos.go`: records the YCYAPI upstream endpoint path for task creation usage logs.
- `frontend/src/views/user/ApiDocsView.vue`: added a user-facing YCYAPI `video-v1-15s` section with curl and Python examples.
- `docs/API_DOCS.md`: documented YCYAPI routing and account setup expectations.
- `progress.md`: recorded this compatibility implementation and verification evidence.
- Rollback: revert the files listed above; `video-v1-15s` will again be rejected by `/v1/videos`, while existing `video-ds`, `as-sd2.0-fast`, and Sora 2 video behavior remains available.

## 2026-07-08 - Task: Build and deploy iotwq/china-api latest multi-arch image

### What was done
- Built and pushed `iotwq/china-api:latest` as a multi-architecture Docker image for `linux/amd64` and `linux/arm64`.
- Stopped the old current-directory Docker Compose services and restarted them with the newly pulled `iotwq/china-api:latest` image.
- Verified the local running application container is using the new arm64 image from the pushed multi-arch manifest.

### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 -t iotwq/china-api:latest --push .` passed.
- `docker buildx imagetools inspect iotwq/china-api:latest` showed `linux/amd64` and `linux/arm64` manifests under digest `sha256:7ee45c3136cd5f60405dc25d2926cbad075905b23a0cabf1cfa3617e5e16f519`.
- `docker compose down` followed by `docker compose pull sub2api && docker compose up -d` completed.
- `docker compose ps` showed `sub2api`, `sub2api-postgres`, and `sub2api-redis` healthy.
- `curl -fsS http://127.0.0.1:8080/health` returned `{"status":"ok"}`.

### Notes
- `iotwq/china-api:latest`: pushed multi-arch manifest digest `sha256:7ee45c3136cd5f60405dc25d2926cbad075905b23a0cabf1cfa3617e5e16f519`.
- `docker-compose.yaml`: used as the deployment compose file; no file content was changed in this task.
- `progress.md`: recorded the image build, push, compose restart, and verification evidence.
- Rollback: deploy a previous known-good image tag or digest by changing the compose image reference and running `docker compose pull sub2api && docker compose up -d`; local data directories `data`, `postgres_data`, and `redis_data` were preserved.

## 2026-07-08 - Task: Polish community chat and video API docs wording

### What was done
- Renamed the user community chat entry from “联系站长” to “私聊站长”.
- Updated `video-ds-2.0` and `as-sd2.0-fast` duration text in the user API docs to “通常 15 秒”.
- Simplified the `video-v1-15s` API docs into the same task flow used by the other video models: submit task, poll status, and download content.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- frontend/src/i18n/locales/zh.ts frontend/src/views/user/ApiDocsView.vue docs/API_DOCS.md docs/COMMUNITY_CHAT.md` passed from the repository root.

### Notes
- `frontend/src/i18n/locales/zh.ts`: changed the community chat site-owner action label to “私聊站长”.
- `frontend/src/views/user/ApiDocsView.vue`: updated video model duration wording and simplified the `video-v1-15s` examples.
- `docs/COMMUNITY_CHAT.md`: aligned the private site-owner chat wording.
- `progress.md`: recorded this wording and documentation update.
- Rollback: revert the files listed above to restore the previous “联系站长” label and earlier API docs wording.

## 2026-07-08 - Task: Correct YCYAPI video-v1-15s API docs path

### What was done
- Corrected the user-facing `video-v1-15s` API docs to use YCYAPI's `/v1/video/generations` path for task creation and status polling.
- Kept `video-v1-15s` content download on `/v1/videos/{task_id}/content`, matching the YCYAPI task result flow.
- Added matching gateway aliases so users can call `/v1/video/generations` and `/v1/video/generations/{task_id}` directly through this service.
- Kept `video-ds-2.0`, `video-ds-2.0-fast`, `as-sd2.0-fast`, and Sora 2 docs on their existing `/v1/videos` flow.

### Testing
- `go test ./internal/handler ./internal/server/routes ./internal/service -run 'TestNormalizeInboundEndpoint|TestDeriveUpstreamEndpoint|TestGatewayRoutesOpenAIVideoTaskPathsAreRegistered|TestForwardOpenAIVideo|TestParseOpenAIVideoCreateRequest'` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- frontend/src/views/user/ApiDocsView.vue docs/API_DOCS.md backend/internal/handler/endpoint.go backend/internal/handler/endpoint_test.go backend/internal/server/routes/gateway.go backend/internal/server/routes/gateway_test.go` passed from the repository root.

### Notes
- `frontend/src/views/user/ApiDocsView.vue`: changed only the `video-v1-15s` section to show `/v1/video/generations` for create/status examples.
- `docs/API_DOCS.md`: updated the YCYAPI route description and made video billing wording independent of one create path.
- `backend/internal/server/routes/gateway.go`: registered OpenAI-platform aliases for YCYAPI-style video create/status paths.
- `backend/internal/server/routes/gateway_test.go`: added route coverage for the new YCYAPI-style aliases.
- `backend/internal/handler/endpoint.go`: added endpoint normalization for `/v1/video/generations` so logs and upstream endpoint tracking stay consistent.
- `backend/internal/handler/endpoint_test.go`: added normalization and upstream endpoint derivation coverage for the new endpoint.
- `progress.md`: recorded this route and documentation correction.
- Rollback: revert the files listed above to remove the `/v1/video/generations` aliases and restore the previous `video-v1-15s` docs that used `/v1/videos`.

## 2026-07-08 - Task: Build and push iotwq/china-api latest multi-arch image

### What was done
- Built `iotwq/china-api:latest` from the current workspace code as a multi-architecture Docker image.
- Pushed the new `latest` manifest to Docker Hub for both `linux/amd64` and `linux/arm64`.

### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 -t iotwq/china-api:latest --push .` passed.
- `docker buildx imagetools inspect iotwq/china-api:latest` showed `linux/amd64` and `linux/arm64` manifests under digest `sha256:838b4d745beefc0a322aeff8748aebe0deab44a92b1ade7d0c585600425c4f16`.

### Notes
- `iotwq/china-api:latest`: pushed multi-arch manifest digest `sha256:838b4d745beefc0a322aeff8748aebe0deab44a92b1ade7d0c585600425c4f16`.
- Build warnings: frontend build still reports stale Browserslist data and Vite chunk-size warnings; these did not fail the image build.
- `progress.md`: recorded the image build and Docker Hub push verification evidence.
- Rollback: redeploy a previous known-good image digest or tag by changing the compose image reference, then run `docker compose pull sub2api && docker compose up -d`.

## 2026-07-08 - Task: Fix direct chat owner view layout for normal users

### What was done
- Adjusted the “私聊站长” dialog so normal users use a single-column chat layout.
- Kept the station-owner/admin view on the existing two-column layout with the conversation list and active thread.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- frontend/src/views/user/CommunityChatView.vue` passed from the repository root.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: added a non-admin direct-chat layout class so the thread fills the modal instead of staying in the left grid column.
- `progress.md`: recorded this direct chat layout correction.
- Rollback: revert the `CommunityChatView.vue` class and style changes to restore the previous shared grid behavior for admin and normal users.

## 2026-07-08 - Task: Support pasted images in community chat

### What was done
- Added paste handling to the community chat message input so users can paste images directly from the clipboard.
- Reused the existing image validation and preview flow, including image type checks and the 15 MB size limit.
- Confirmed the sidebar community chat unread indicator is already driven by WebSocket `message_created` events from other users when the current route is outside `/community-chat`.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- frontend/src/views/user/CommunityChatView.vue` passed from the repository root.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: added clipboard image extraction and wired the group chat textarea paste event to the existing image preview/upload flow.
- `frontend/src/components/layout/AppSidebar.vue`: read-only confirmation; existing logic sets the community chat unread dot for other users' group messages while outside the chat page.
- `progress.md`: recorded this pasted-image support update and notification behavior check.
- Rollback: revert the `CommunityChatView.vue` paste handler changes to restore upload-button-only image selection.

## 2026-07-08 - Task: Notify site owner for direct chat messages

### What was done
- Updated the sidebar community chat unread badge logic to include direct messages sent to the site owner.
- Kept the existing group chat unread behavior unchanged.
- Scoped the direct-message badge to admin/site-owner accounts and ignored messages sent by the current user.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- frontend/src/components/layout/AppSidebar.vue` passed from the repository root.

### Notes
- `frontend/src/components/layout/AppSidebar.vue`: extended the WebSocket badge handler to set the community chat unread dot for `direct_message_created` events visible to the site owner.
- `progress.md`: recorded this site-owner direct-message notification update.
- Rollback: revert the `AppSidebar.vue` badge handler changes to restore group-message-only unread behavior.

## 2026-07-08 - Task: Fix PDF upload processing worker configuration

### What was done
- Configured the PDF.js worker URL used by the model chat PDF attachment parser.
- Kept the existing PDF text extraction flow unchanged; only the missing worker configuration was added.
- Updated the ChatView test mock so PDF worker configuration remains covered in tests.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm exec vitest run src/views/user/__tests__/ChatView.spec.ts` passed from `frontend/`.
- `git diff --check -- frontend/src/views/user/ChatView.vue frontend/src/views/user/__tests__/ChatView.spec.ts` passed from the repository root.

### Notes
- `frontend/src/views/user/ChatView.vue`: imports the PDF.js worker asset URL and assigns it to `GlobalWorkerOptions.workerSrc` before parsing PDFs.
- `frontend/src/views/user/__tests__/ChatView.spec.ts`: adds `GlobalWorkerOptions` to the mocked PDF.js module.
- `progress.md`: recorded this PDF worker configuration fix.
- Rollback: revert the `ChatView.vue` and `ChatView.spec.ts` changes to restore the previous PDF.js setup.

## 2026-07-08 - Task: Add Grok to channel monitor

### What was done
- Added Grok as a supported channel monitor provider in backend validation, request binding, ent provider enums, and database CHECK constraints.
- Reused the OpenAI Chat Completions compatible probe for Grok basic text model checks, including Bearer auth and `/v1/chat/completions`.
- Added Grok to the admin channel monitor UI provider filter, create form, request template manager, shared labels, badges, and monitor provider icon.

### Testing
- `go test -tags unit ./internal/service -run TestRunCheckForModel_Grok_DefaultChatRequest` passed from `backend/`.
- `go test -tags unit ./internal/service -run 'TestRunCheckForModel|TestChannelMonitor'` passed from `backend/`.
- `go test ./internal/handler/admin ./internal/service` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- <changed Grok channel monitor files>` passed from the repository root.

### Notes
- `backend/internal/service/channel_monitor_const.go`: added the Grok provider constant and updated the invalid-provider error text.
- `backend/internal/service/channel_monitor_checker.go`: registered Grok with the existing OpenAI-compatible chat probe and protected Grok merge-mode body keys.
- `backend/internal/service/channel_monitor_checker_body_test.go`: added coverage for the Grok text probe path, request body, and authorization header.
- `backend/internal/service/channel_monitor_template_types.go`: updated request template provider error text.
- `backend/internal/handler/admin/channel_monitor_handler.go`: allowed `grok` in monitor create/update request binding.
- `backend/internal/handler/admin/channel_monitor_template_handler.go`: allowed `grok` in request template create binding.
- `backend/ent/schema/channel_monitor.go`: added `grok` to the monitor provider enum schema.
- `backend/ent/schema/channel_monitor_request_template.go`: added `grok` to the request template provider enum schema.
- `backend/ent/migrate/schema.go`: aligned generated migration metadata with the Grok provider enum.
- `backend/ent/channelmonitor/channelmonitor.go`: added the generated Grok provider enum value and validator branch.
- `backend/ent/channelmonitorrequesttemplate/channelmonitorrequesttemplate.go`: added the generated Grok provider enum value and validator branch.
- `backend/migrations/171_channel_monitor_grok_provider.sql`: widens channel monitor provider CHECK constraints to include `grok`.
- `frontend/src/api/admin/channelMonitor.ts`: added `grok` to the admin monitor provider type.
- `frontend/src/constants/channelMonitor.ts`: added the Grok provider constant and included it in the shared provider list.
- `frontend/src/composables/useChannelMonitorFormat.ts`: added Grok labels, badge classes, picker classes, and monitor card gradient.
- `frontend/src/components/user/monitor/ProviderIcon.vue`: added the Grok/xAI icon path for monitor cards and provider buttons.
- `frontend/src/components/admin/monitor/MonitorFiltersBar.vue`: added Grok to the provider filter.
- `frontend/src/components/admin/monitor/MonitorFormDialog.vue`: added Grok to the create/edit provider picker and adjusted the provider grid for four providers.
- `frontend/src/components/admin/monitor/MonitorTemplateManagerDialog.vue`: added Grok to template provider tabs, creation picker, and provider counts.
- `frontend/src/i18n/locales/zh.ts`: added the Grok monitor provider label.
- `frontend/src/i18n/locales/en.ts`: added the Grok monitor provider label.
- `docs/UPSTREAM_SYNC.md`: documented that Grok channel monitor uses the basic Chat Completions compatible text probe.
- `progress.md`: recorded this Grok channel monitor update.
- Rollback: revert the files listed above and remove `backend/migrations/171_channel_monitor_grok_provider.sql`; if the migration has already run in PostgreSQL, first delete or change any `provider='grok'` monitor/template rows, then recreate both provider CHECK constraints with only `openai`, `anthropic`, and `gemini`.

## 2026-07-08 - Task: Add community chat image preview and file attachments

### What was done
- Added click-to-preview and download actions for images sent in user community chat.
- Expanded community chat upload from image-only to common attachments: images, PDF, Office documents, Markdown, text, CSV, and RTF, with the existing 15 MB limit.
- Added file-card rendering and download support for non-image attachments while keeping the old image upload endpoint compatible.

### Testing
- `go test ./internal/service ./internal/handler ./internal/server/routes` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- <changed community chat attachment files>` passed from the repository root.

### Notes
- `backend/internal/service/community_chat.go`: added the `file` message type, attachment size/error constants, and response aliases for file metadata.
- `backend/internal/handler/community_chat_handler.go`: generalized upload handling to accept `file` or legacy `image` multipart fields, validate allowed attachment types, and support forced download with `download=1`.
- `backend/internal/server/routes/community_chat.go`: added `POST /community-chat/files` while keeping `POST /community-chat/images`.
- `backend/migrations/172_community_chat_file_messages.sql`: widens public community chat message type CHECK constraints to include `file`.
- `frontend/src/api/communityChat.ts`: added the `file` message type and `uploadFileMessage` API.
- `frontend/src/views/user/CommunityChatView.vue`: adds attachment selection, file validation, image lightbox preview, image download, and non-image file cards.
- `frontend/src/i18n/locales/zh.ts`: adds community chat attachment labels and validation messages.
- `frontend/src/i18n/locales/en.ts`: adds community chat attachment labels and validation messages.
- `docs/COMMUNITY_CHAT.md`: documents attachment support, the new upload endpoint, and download behavior.
- `progress.md`: recorded this community chat attachment update.
- Rollback: revert the files listed above and remove `backend/migrations/172_community_chat_file_messages.sql`; if the migration has already run, first delete or convert `message_type='file'` rows, then recreate `community_chat_messages_type_check` with only `text` and `image`.

## 2026-07-08 - Task: Improve community chat scrolling and MP4 attachments

### What was done
- Made the community chat history scroll to the latest messages by default after loading, with an extra frame wait so layout/media sizing does not leave the view at the top.
- Widened and recolored the community chat and direct chat scrollbars to make them easier to grab with a mouse.
- Added MP4 to the community chat attachment whitelist and rendered MP4 messages with an inline video player plus download action.

### Testing
- `go test ./internal/service ./internal/handler ./internal/server/routes` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- <changed community chat scroll/video files>` passed from the repository root.

### Notes
- `backend/internal/handler/community_chat_handler.go`: added `.mp4` / `video/mp4` to the attachment whitelist.
- `frontend/src/views/user/CommunityChatView.vue`: adds video rendering/download, robust bottom scrolling after history/media load, MP4 client-side validation, and wider visible scrollbars.
- `frontend/src/i18n/locales/zh.ts`: adds the community chat video label and MP4 validation copy.
- `frontend/src/i18n/locales/en.ts`: adds the community chat video label and MP4 validation copy.
- `docs/COMMUNITY_CHAT.md`: documents MP4 attachment support and video rendering behavior.
- `progress.md`: recorded this community chat scroll and MP4 update.
- Rollback: revert the files listed above to restore the previous image/document-only attachment behavior and default scrollbar styling.

## 2026-07-08 - Task: Add enlarged MP4 preview in community chat

### What was done
- Added a “放大播放” action for MP4 messages in user community chat.
- Reused the existing attachment preview dialog pattern so MP4 videos can play in a larger modal without downloading.
- Kept the inline MP4 player and download action in the chat bubble.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- <changed community chat video preview files>` passed from the repository root.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: added video preview state, open/close handlers, a large video dialog, and the MP4 preview button.
- `frontend/src/i18n/locales/zh.ts`: added the “放大播放” community chat label.
- `frontend/src/i18n/locales/en.ts`: added the large-player community chat label.
- `docs/COMMUNITY_CHAT.md`: documented MP4 enlarged playback behavior.
- `progress.md`: recorded this MP4 preview update.
- Rollback: revert the files listed above to restore inline-only MP4 playback.

## 2026-07-09 - Task: Audit and fix new community chat additions

### What was done
- Checked the recent community chat, attachment, private chat, PDF upload, and Grok monitor additions.
- Fixed attachment uploads so user-entered message text remains the message caption while downloads keep the original uploaded filename.
- Fixed the site-owner private chat conversation list so it can show the user's stored profile avatar.
- Added small backend tests for attachment filename aliasing and download filename sanitization.

### Testing
- `go test ./internal/service ./internal/handler ./internal/server/routes` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm exec vitest run src/views/user/__tests__/ChatView.spec.ts` passed from `frontend/`.
- `go test -tags unit ./internal/service -run 'TestRunCheckForModel_Grok_DefaultChatRequest|TestRunCheckForModel|TestChannelMonitor'` passed from `backend/`.
- `git diff --check -- backend/internal/handler/community_chat_handler.go backend/internal/handler/community_chat_handler_test.go backend/internal/service/community_chat.go backend/internal/service/community_chat_test.go backend/internal/repository/community_chat_repo.go frontend/src/views/user/CommunityChatView.vue docs/COMMUNITY_CHAT.md progress.md` passed from the repository root.

### Notes
- `backend/internal/handler/community_chat_handler.go`: preserves original attachment filenames in the upload URL and uses them for forced downloads.
- `backend/internal/handler/community_chat_handler_test.go`: covers attachment download filename sanitization.
- `backend/internal/service/community_chat.go`: returns attachment file aliases using the original filename query parameter when available.
- `backend/internal/service/community_chat_test.go`: covers attachment filename alias normalization.
- `backend/internal/repository/community_chat_repo.go`: reads user avatar URLs for site-owner private chat conversation rows.
- `frontend/src/views/user/CommunityChatView.vue`: displays attachment captions separately from file/video cards and resolves original filenames from API fields or URLs.
- `docs/COMMUNITY_CHAT.md`: documents original-name attachment downloads and site-owner private chat avatar behavior.
- `progress.md`: recorded this audit and fix.
- Rollback: revert the files listed above to restore the previous behavior where attachment message content doubled as the file name and private chat conversation avatars were blank.

## 2026-07-09 - Task: Block underfunded per-request video generation

### What was done
- Confirmed video generation could be started when the user only had a positive balance, even if the configured per-request video price was higher than the balance.
- Added a pre-dispatch estimated-cost check for OpenAI-compatible video creation so balance-mode users must have enough balance for the single video price before the upstream task is created.
- Added the same estimated-cost coverage check for subscription groups so a video task cannot start when the remaining daily, weekly, or monthly subscription budget cannot cover the video price.
- Reused the existing video per-request pricing calculation so the preflight amount matches the final usage billing amount.

### Testing
- `go test -tags unit ./internal/service -run 'TestCheckEstimatedCostCoverage|TestEstimateOpenAIVideoCreateCostUsesPerRequestPricing|TestParseOpenAIVideoCreateRequest'` passed from `backend/`.
- `go test ./internal/service ./internal/handler ./internal/server/routes` passed from `backend/`.
- `git diff --check -- backend/internal/service/billing_cache_service.go backend/internal/service/billing_cache_service_balance_test.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_videos_test.go backend/internal/handler/openai_videos.go docs/API_DOCS.md progress.md` passed from the repository root.

### Notes
- `backend/internal/service/billing_cache_service.go`: added estimated-cost coverage checks for balance and subscription billing modes.
- `backend/internal/service/billing_cache_service_balance_test.go`: covers rejection when balance is lower than the estimated per-request video cost.
- `backend/internal/service/openai_gateway_service.go`: added OpenAI-compatible video create cost estimation using the same per-request pricing path as final billing.
- `backend/internal/service/openai_videos_test.go`: verifies video cost estimation resolves configured per-request pricing.
- `backend/internal/handler/openai_videos.go`: runs the estimated-cost check before dispatching video creation upstream.
- `docs/API_DOCS.md`: documents that video creation requires enough balance or subscription budget to cover the single-call price.
- `progress.md`: recorded this billing guard fix.
- Rollback: revert the files listed above to restore the previous behavior where video creation only required a positive balance before final billing.

## 2026-07-09 - Task: Add balance holds for concurrent image and video generation

### What was done
- Added a pre-dispatch balance hold flow for OpenAI-compatible image and video generation so concurrent high-cost media requests cannot all pass on the same available balance.
- Reused the existing atomic balance/frozen-balance billing repository operations: successful holds reduce available balance before upstream dispatch, successful upstream creation captures the held amount, and failed dispatch releases the hold.
- Updated usage recording so requests whose balance was already captured from a hold still write usage logs and quota/account usage, but do not deduct user balance a second time.
- Added image preflight cost estimation based on model, requested image count, and size tier; video continues to use the per-request estimate added earlier.

### Testing
- `go test -tags unit ./internal/service -run 'TestCheckEstimatedCostCoverage|TestEstimateOpenAIVideoCreateCostUsesPerRequestPricing|TestOpenAIGatewayServiceRecordUsage_BalanceAlreadyCapturedSkipsBalanceDeduction|TestOpenAIGatewayServiceRecordUsage_ChannelPerRequestVideoBillingUsesTaskID'` passed from `backend/`.
- `go test ./internal/service ./internal/handler ./internal/server/routes` passed from `backend/`.

### Notes
- `backend/internal/service/openai_media_balance_hold.go`: added reusable OpenAI media balance hold, capture, and release helpers.
- `backend/internal/service/openai_gateway_service.go`: added image cost estimation and the usage-recording flag for already-captured balance.
- `backend/internal/service/gateway_service.go`: prevents balance deduction and balance-cache deduction when a media hold has already captured the balance.
- `backend/internal/handler/openai_images.go`: reserves balance before image upstream dispatch, captures on success, and releases on failure.
- `backend/internal/handler/openai_videos.go`: reserves balance before video upstream dispatch, captures on task creation success, and releases on failure.
- `backend/internal/service/openai_gateway_record_usage_test.go`: covers that already-captured balance does not get deducted again during usage recording.
- `docs/API_DOCS.md`: documents image/video balance holds and concurrent overdraft protection.
- `progress.md`: recorded this media balance hold fix.
- Rollback: revert the files listed above to return to pre-dispatch balance checks without atomic balance holds.

## 2026-07-09 - Task: Audit media balance hold billing safeguards

### What was done
- Reviewed the abnormal image/video overdraft fix across preflight cost checks, balance holds, capture, release, and final usage billing.
- Tightened OpenAI-compatible image pre-dispatch holds so token-priced image models keep using token billing and are not held with image-tier pricing.
- Extended the already-captured balance guard to the legacy billing fallback path so a captured hold cannot be deducted again there.
- Added regression coverage for token-priced image hold skipping and legacy-path balance deduction skipping.

### Testing
- `go test -tags unit ./internal/service -run 'TestCheckEstimatedCostCoverage|TestEstimateOpenAIVideoCreateCostUsesPerRequestPricing|TestEstimateOpenAIImagesCost_SkipsTokenPricedImageModel|TestOpenAIGatewayServiceRecordUsage_BalanceAlreadyCapturedSkipsBalanceDeduction|TestOpenAIGatewayServiceRecordUsage_BalanceAlreadyCapturedSkipsLegacyBalanceDeduction|TestOpenAIGatewayServiceRecordUsage_ChannelPerRequestVideoBillingUsesTaskID'` passed from `backend/`.
- `go test ./internal/service ./internal/handler ./internal/server/routes` passed from `backend/`.
- `git diff --check -- backend/internal/service/openai_media_balance_hold.go backend/internal/service/openai_gateway_service.go backend/internal/service/gateway_service.go backend/internal/service/openai_gateway_record_usage_test.go backend/internal/service/billing_cache_service.go backend/internal/service/billing_cache_service_balance_test.go backend/internal/service/openai_videos_test.go backend/internal/handler/openai_images.go backend/internal/handler/openai_videos.go docs/API_DOCS.md progress.md` passed from the repository root.

### Notes
- `backend/internal/service/openai_gateway_service.go`: skips image balance holds for channel pricing configured as token billing.
- `backend/internal/service/gateway_service.go`: makes the legacy balance deduction path honor the already-captured balance flag.
- `backend/internal/service/openai_gateway_record_usage_test.go`: covers token-priced image hold skipping and legacy-path double-deduction prevention.
- `progress.md`: recorded this billing safeguard audit.
- Rollback: revert the files listed above to return to the previous media balance hold behavior.

## 2026-07-09 - Task: Keep media balance holds out of simple mode

### What was done
- Found that the new media balance hold helper could still reserve balance in simple run mode even though final usage recording does not bill in that mode.
- Updated the OpenAI media hold entry point to skip holds when `run_mode` is `simple`.
- Added regression coverage to confirm simple mode does not call the reserve path.

### Testing
- `go test -tags unit ./internal/service -run 'TestCheckEstimatedCostCoverage|TestEstimateOpenAIVideoCreateCostUsesPerRequestPricing|TestEstimateOpenAIImagesCost_SkipsTokenPricedImageModel|TestReserveOpenAIMediaBalance_SkipsSimpleMode|TestOpenAIGatewayServiceRecordUsage_BalanceAlreadyCapturedSkipsBalanceDeduction|TestOpenAIGatewayServiceRecordUsage_BalanceAlreadyCapturedSkipsLegacyBalanceDeduction|TestOpenAIGatewayServiceRecordUsage_ChannelPerRequestVideoBillingUsesTaskID'` passed from `backend/`.
- `go test ./internal/service ./internal/handler ./internal/server/routes` passed from `backend/`.

### Notes
- `backend/internal/service/openai_media_balance_hold.go`: skips media balance holds in simple run mode.
- `backend/internal/service/openai_gateway_record_usage_test.go`: adds simple-mode hold skip coverage and an in-file reserve stub.
- `progress.md`: recorded this simple-mode billing safeguard.
- Rollback: revert the files listed above to restore the previous media hold behavior in simple run mode.

## 2026-07-09 - Task: Align video-v1-15s API docs with canvas workflow

### What was done
- Checked `../gpt_image_playground` canvas video node logic for the working `video-v1-15s` flow.
- Updated the user-facing API docs so `video-v1-15s` uses `/v1/video/generations` for create/status, `ratio` for aspect ratio, and `image`/`images` for reference images.
- Documented the model-specific defaults and limits: default `9:16`, fixed 15 seconds, fixed 720p, and up to 8 reference images.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- frontend/src/views/user/ApiDocsView.vue docs/API_DOCS.md progress.md` passed from the repository root.

### Notes
- `frontend/src/views/user/ApiDocsView.vue`: updated `video-v1-15s` examples, parameters, status polling, and notes to match the canvas implementation.
- `docs/API_DOCS.md`: recorded the YCYAPI request body shape and field restrictions.
- `progress.md`: recorded this API docs alignment.
- Rollback: revert the files listed above to restore the previous `video-v1-15s` documentation.

## 2026-07-09 - Task: Clean API docs wording for end users

### What was done
- Reviewed the API docs page wording from an end-user perspective.
- Removed upstream/vendor-facing wording from the `video-v1-15s` section and kept the content focused on model usage, parameters, and examples.
- Replaced vague/internal phrases such as video channel and service limits with user-facing interface wording.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check -- frontend/src/views/user/ApiDocsView.vue docs/API_DOCS.md progress.md` passed from the repository root.

### Notes
- `frontend/src/views/user/ApiDocsView.vue`: cleaned user-facing API docs wording for video model sections.
- `docs/API_DOCS.md`: removed owner/operator-facing video routing and billing wording from the API docs summary.
- `progress.md`: recorded this user-facing wording cleanup.
- Rollback: revert the files listed above to restore the previous API docs wording.

## 2026-07-09 - Task: Add Grok video 1.5 fallback account switch

### What was done
- Changed Grok media forwarding so `grok-imagine-video-1.5` is forwarded unchanged by default for both text-only and image-to-video requests.
- Added a Grok account-level compatibility switch in account creation and editing. When enabled, only text-only `grok-imagine-video-1.5` requests fall back to `grok-imagine-video`; image-to-video requests still use 1.5.
- Documented the default Grok Build behavior and the compatibility switch for old upstreams.

### Testing
- `go test -tags unit ./internal/service -run 'TestNormalizeGrokMediaModelForEndpoint|TestAccountIsGrokVideo15TextFallbackEnabled|TestForwardGrokMediaVideoGenerationPreservesTextOnly15ByDefault|TestForwardGrokMediaVideoGenerationFallbacksTextOnly15WhenAccountEnabled|TestForwardGrokMediaVideoGenerationPreservesImageToVideoModel'` passed from `backend/`.
- `go test ./internal/service` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts` passed from `frontend/`.
- `git diff --check -- backend/internal/service/grok_media.go backend/internal/service/openai_gateway_grok_test.go frontend/src/components/account/CreateAccountModal.vue frontend/src/components/account/EditAccountModal.vue frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts docs/API_DOCS.md` passed from the repository root.

### Notes
- `backend/internal/service/grok_media.go`: reads `extra.grok_video_15_text_fallback_enabled` from Grok accounts before applying the old text-only 1.5 fallback.
- `backend/internal/service/openai_gateway_grok_test.go`: covers default passthrough, compatibility fallback, and account extra parsing.
- `frontend/src/components/account/CreateAccountModal.vue`: adds the Grok account creation switch and writes the extra flag.
- `frontend/src/components/account/EditAccountModal.vue`: loads and saves the Grok account switch.
- `frontend/src/i18n/locales/zh.ts`: adds Chinese labels for the switch.
- `frontend/src/i18n/locales/en.ts`: adds English labels for the switch.
- `docs/API_DOCS.md`: documents the Grok 1.5 default passthrough behavior and compatibility switch.
- `progress.md`: recorded this Grok media compatibility change.
- Rollback: revert the files listed above to restore unconditional text-only `grok-imagine-video-1.5` fallback behavior.

## 2026-07-09 - Task: Support Grok video content downloads

### What was done
- Added Grok group support for `GET /v1/videos/{task_id}/content`, so OpenAI-style video download fallback no longer stops at the platform gate.
- Routed Grok video content downloads to the same task-bound Grok account used by status polling.
- Streamed Grok video content responses directly and forwarded range headers for browser playback and resumable downloads.
- Documented the Grok video query and content download endpoints.

### Testing
- `go test -tags unit ./internal/pkg/xai ./internal/service -run 'TestBuildGrokMediaURLs|Test(GrokMediaGenerationGateCoversImagesAndVideo|ForwardGrokMediaVideoContentUsesGETWithoutBody|ForwardGrokMediaVideoStatusUsesGETWithoutBody|NormalizeGrokMediaModelForEndpoint|AccountIsGrokVideo15TextFallbackEnabled)'` passed from `backend/`.
- `go test ./internal/handler ./internal/server/routes` passed from `backend/`.
- `go test -tags unit ./internal/pkg/xai` passed from `backend/`.
- `go test -tags unit ./internal/service` passed from `backend/`.
- `git diff --check -- backend/internal/pkg/xai/oauth.go backend/internal/pkg/xai/oauth_test.go backend/internal/service/grok_media.go backend/internal/handler/grok_media.go backend/internal/server/routes/gateway.go backend/internal/server/routes/gateway_test.go backend/internal/service/openai_gateway_grok_test.go docs/API_DOCS.md` passed from the repository root.

### Notes
- `backend/internal/pkg/xai/oauth.go`: added the Grok video content URL builder.
- `backend/internal/pkg/xai/oauth_test.go`: covers the Grok video content URL shape.
- `backend/internal/service/grok_media.go`: adds the Grok video content endpoint, routes it to `/videos/{id}/content`, and streams successful content responses.
- `backend/internal/handler/grok_media.go`: adds the Grok video content handler and validates task IDs for content requests.
- `backend/internal/server/routes/gateway.go`: allows Grok groups to handle `/v1/videos/{task_id}/content`.
- `backend/internal/server/routes/gateway_test.go`: covers Grok route registration for video content downloads.
- `backend/internal/service/openai_gateway_grok_test.go`: covers Grok video content forwarding without a request body.
- `docs/API_DOCS.md`: records the Grok video status and content endpoints.
- `progress.md`: recorded this Grok video content download change.
- Rollback: revert the files listed above to restore the previous behavior where Grok groups do not support `/v1/videos/{task_id}/content`.

## 2026-07-09 - Task: Refund failed async video balance charges

### What was done
- Added one-time balance refund support for async OpenAI-compatible video tasks when a later status poll reports a failed or cancelled terminal state.
- Reused the existing billing idempotency table with a video refund request ID derived from `task_id`, so repeated failed-status polling does not refund more than once.
- Added a negative usage log reversal row after a successful refund, so the refund is visible in usage records and cost totals can be reconciled.
- Documented the async video failure refund behavior in API docs and the user-facing API docs page.

### Testing
- `go test ./internal/service -run 'TestOpenAIVideoStatusFailed|TestForwardOpenAIVideoStatus|TestRefundFailedOpenAIVideoTask|TestOpenAIGatewayServiceRecordUsage_ChannelPerRequestVideoBillingUsesTaskID|TestOpenAIGatewayServiceRecordUsage_BalanceAlreadyCaptured'` passed from `backend/`.
- `go test ./internal/repository -run 'TestUsageBillingRepository|TestUsageLogRepository'` passed from `backend/`.
- `go test ./internal/service ./internal/handler ./internal/repository ./internal/server/routes` passed from `backend/`.
- `go test ./internal/server ./internal/server/routes` passed from `backend/`.
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `git diff --check` passed from the repository root.

### Notes
- `backend/internal/handler/openai_videos.go`: schedules a refund task after status polling returns a failed or cancelled video status.
- `backend/internal/service/openai_videos.go`: adds video failure-state detection, refund request ID generation, and balance refund orchestration.
- `backend/internal/service/openai_gateway_service.go`: carries JSON response bodies in OpenAI forward results so video status responses can be inspected.
- `backend/internal/service/account_usage_service.go`: adds a usage-log lookup contract for original billed requests.
- `backend/internal/repository/usage_log_repo.go`: adds lookup by `request_id` and `api_key_id`.
- `backend/internal/repository/usage_billing_repo.go`: supports negative balance billing commands as refunds.
- `backend/internal/service/openai_gateway_record_usage_test.go`: covers one-time video refund behavior and subscription-skip behavior.
- `backend/internal/service/openai_videos_test.go`: covers failed/cancelled status detection and forwarded status response bodies.
- `docs/API_DOCS.md`: documents one-time automatic balance refunds for failed or cancelled async video tasks.
- `frontend/src/views/user/ApiDocsView.vue`: adds the user-facing video refund note.
- `progress.md`: recorded this async video refund change.
- Rollback: revert the files listed above to remove failed-video automatic refunds and restore the previous no-refund behavior.

## 2026-07-10 - Task: Label failed video refund usage records

### What was done
- Added a dedicated usage-record label for async video failure refund rows created with the `openai-video-refund:` request ID prefix.
- Kept the original video model visible below the refund label so users can still see which model the refund belongs to.
- Displayed refund amounts with a clear negative currency format in the table and cost tooltip.

### Testing
- `pnpm exec vitest run src/components/admin/usage/__tests__/UsageTable.spec.ts` passed from `frontend/`.
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `git diff --check -- frontend/src/components/admin/usage/UsageTable.vue frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts progress.md` passed from the repository root.

### Notes
- `frontend/src/components/admin/usage/UsageTable.vue`: detects video refund rows and renders the “视频失败退款” label plus signed refund cost formatting.
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`: covers refund label rendering and negative refund amount display.
- `frontend/src/i18n/locales/zh.ts`: adds the Chinese refund label.
- `frontend/src/i18n/locales/en.ts`: adds the English refund label.
- `progress.md`: recorded this usage-record display change.
- Rollback: revert the files listed above to restore the previous generic model and cost display for refund usage rows.

## 2026-07-10 - Task: Sync upstream Wei-Shaw/sub2api to v0.1.149

### What was done
- Synced upstream `Wei-Shaw/sub2api` from `f68f3b86` to `12d811bd` while preserving local ChinaAPI features.
- Kept upstream's large backend and i18n split refactor, then migrated local video routing, video billing/refund, media balance guard, image bridge controls, Grok video compatibility, community chat, image workspace, API docs, and AI learning homepage customizations onto the new structure.
- Resolved frontend conflicts so the sidebar keeps local user entries and unread chat badges while also preserving upstream sidebar scroll restoration.
- Resolved i18n conflicts by moving local `zh.ts/en.ts` keys into upstream split locale modules and deleting the old single locale files.
- Fixed a merge gap in public settings so `IMAGE_WORKSPACE_URL` still controls the image workspace URL after upstream split `setting_service.go`.

### Testing
- `go test -tags unit ./internal/service ./internal/repository` passed from `backend/`.
- `go test ./internal/service ./internal/repository ./internal/handler ./internal/server/routes` passed from `backend/`.
- `pnpm build` passed from `frontend/`; it still reports existing Browserslist age and Vite large chunk warnings.
- `pnpm test:run src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/ChatView.spec.ts src/stores/__tests__/app.spec.ts` passed from `frontend/`; it still reports existing `--localstorage-file` and Browserslist warnings.
- `git diff --check` passed from the repository root.

### Notes
- `.dockerignore`: synchronized upstream ignore updates.
- `.gitignore`: synchronized upstream ignore updates.
- `backend/cmd/server/wire_gen.go`: kept dependency injection wiring compatible with upstream and local handlers.
- `backend/internal/repository/usage_billing_repo.go`: preserved local insufficient-balance guard behavior across the upstream repository state.
- `backend/internal/repository/usage_billing_repo_unit_test.go`: updated repository tests for guarded balance behavior.
- `backend/internal/repository/usage_log_repo.go`: accepted upstream repository split state.
- `backend/internal/repository/usage_log_repo_query.go`: kept local usage-log lookup helpers for video refund reconciliation.
- `backend/internal/service/gateway_service.go`: accepted upstream service split state.
- `backend/internal/service/gateway_usage_billing.go`: preserved local balance-captured media billing behavior.
- `backend/internal/service/openai_gateway_forward.go`: kept local image bridge request controls in the upstream split gateway path.
- `backend/internal/service/openai_gateway_grok_test.go`: merged upstream Grok fallback tests with local Grok video resolution/duration expectations.
- `backend/internal/service/openai_gateway_passthrough.go`: kept dangling image `tool_choice` normalization in passthrough forwarding.
- `backend/internal/service/openai_gateway_request_body.go`: kept local Responses image tool-choice normalization helper.
- `backend/internal/service/openai_gateway_response_handling.go`: preserved graceful client-disconnect stream handling.
- `backend/internal/service/openai_gateway_service.go`: merged upstream gateway split with local media result fields, billing repository injection, and image bridge disable header.
- `backend/internal/service/openai_gateway_usage.go`: preserved local image/video per-request cost estimation and media billing metadata.
- `backend/internal/service/openai_ws_forwarder.go`: accepted upstream WebSocket split state.
- `backend/internal/service/openai_ws_forwarder_ingress.go`: kept local image bridge disable header behavior in WS ingress.
- `backend/internal/service/setting_public.go`: restored `IMAGE_WORKSPACE_URL` public setting resolution after upstream settings split.
- `backend/internal/service/setting_service.go`: restored image workspace URL constants used by service tests.
- `backend/internal/service/setting_service.go` split companions under `backend/internal/service/`: synchronized upstream settings/service decomposition while preserving local public settings behavior.
- `docs/UPSTREAM_SYNC.md`: appended the upstream `v0.1.149` sync record, conflict decisions, tests, and rollback points.
- `frontend/src/components/layout/AppSidebar.vue`: merged upstream sidebar scroll persistence with local user navigation entries and community chat unread badges.
- `frontend/src/views/HomeView.vue`: merged upstream URL sanitization with the local red Three.js AI learning homepage.
- `frontend/src/views/KeyUsageView.vue`: removed an unused `githubUrl` variable that blocked `vue-tsc`.
- `frontend/src/i18n/locales/en.ts`: removed the old single-file English locale after migrating local keys into split modules.
- `frontend/src/i18n/locales/zh.ts`: removed the old single-file Chinese locale after migrating local keys into split modules.
- `frontend/src/i18n/locales/en/common.ts`: added local navigation labels to the split English common locale.
- `frontend/src/i18n/locales/zh/common.ts`: added local navigation labels to the split Chinese common locale.
- `frontend/src/i18n/locales/en/dashboard.ts`: added the video failure refund usage label to the split English dashboard locale.
- `frontend/src/i18n/locales/zh/dashboard.ts`: added the video failure refund usage label to the split Chinese dashboard locale.
- `frontend/src/i18n/locales/en/landing.ts`: migrated local AI learning homepage copy to the split English landing locale.
- `frontend/src/i18n/locales/zh/landing.ts`: migrated local AI learning homepage copy to the split Chinese landing locale.
- `frontend/src/i18n/locales/en/misc.ts`: migrated local model chat, community chat, API docs, and image workspace text to the split English misc locale.
- `frontend/src/i18n/locales/zh/misc.ts`: migrated local model chat, community chat, API docs, and image workspace text to the split Chinese misc locale.
- `progress.md`: recorded this upstream sync and verification.
- Rollback: reset to `backup/pre-upstream-sync-20260710-004247` to return to the pre-sync tree, or revert merge commit `a7264de6` plus the post-merge conflict-resolution edits listed above. The pre-sync stash is `stash@{0}` with message `pre-upstream-sync-20260710-004247`.

## 2026-07-10 - Task: Restore Grok 1.5 text-only video fallback

### What was done
- Restored Grok video routing so text-only `grok-imagine-video-1.5` requests are automatically rewritten to `grok-imagine-video`.
- Kept image-to-video requests on `grok-imagine-video-1.5` when a reference image is present.
- Removed the Grok account page compatibility switch because the fallback is now fixed default behavior again.
- Cleans the legacy `grok_video_15_text_fallback_enabled` extra field when Grok accounts are saved.
- Updated the API docs to describe the restored default routing behavior.

### Testing
- `go test -tags unit ./internal/service -run 'TestNormalizeGrokMediaModelForEndpoint|TestForwardGrokMediaVideoGenerationFallbacksTextOnly15|TestForwardGrokMediaVideoGenerationPreservesImageToVideoModel|TestForwardGrokMediaVideoContentUsesGETWithoutBody|TestForwardGrokMediaVideoStatusUsesGETWithoutBody'` passed from `backend/`.
- `go test ./internal/service` passed from `backend/`.
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `git diff --check` passed from the repository root.

### Notes
- `backend/internal/service/grok_media.go`: removed the account-level switch and made text-only `grok-imagine-video-1.5` fallback unconditional.
- `backend/internal/service/openai_gateway_grok_test.go`: updated Grok media tests for text-only fallback and image-to-video passthrough.
- `frontend/src/components/account/CreateAccountModal.vue`: removed the Grok fallback switch and strips the legacy extra field on save.
- `frontend/src/components/account/EditAccountModal.vue`: removed the Grok fallback switch and strips the legacy extra field on save.
- `docs/API_DOCS.md`: documents that text-only `grok-imagine-video-1.5` downgrades to `grok-imagine-video`, while image-to-video keeps 1.5.
- `progress.md`: recorded this Grok fallback restoration.
- Rollback: revert the files listed above to restore the previous account-level optional fallback behavior.

## 2026-07-10 - Task: Fix Grok provider label in channel monitor

### What was done
- Added the missing Grok provider label to the shared channel monitor locale keys.
- Channel monitor provider selectors and status formatting now display `Grok` instead of the raw `monitorCommon.providers.grok` i18n key.

### Testing
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `git diff --check -- frontend/src/i18n/locales/zh/dashboard.ts frontend/src/i18n/locales/en/dashboard.ts` passed from the repository root.

### Notes
- `frontend/src/i18n/locales/zh/dashboard.ts`: added `monitorCommon.providers.grok`.
- `frontend/src/i18n/locales/en/dashboard.ts`: added `monitorCommon.providers.grok`.
- `progress.md`: recorded this channel monitor label fix.
- Rollback: revert the three files listed above to restore the previous locale state.

## 2026-07-10 - Task: Sync upstream Wei-Shaw/sub2api to v0.1.150

### What was done
- Synced upstream `Wei-Shaw/sub2api` from `12d811bd` to `6dd3274a` while preserving local ChinaAPI changes.
- Created backup branch `backup/pre-upstream-sync-20260710-155249` and stash `pre-upstream-sync-20260710-155249` before merging.
- Merged upstream GPT-5.6 billing/cache fixes, compact SSE/raw output fixes, setup-token refresh, payment/concurrency hardening, `parallel_tool_calls` compatibility, and frontend i18n/test updates.
- Restored the local staged and untracked feature work after the upstream merge, then resolved the remaining OpenAI gateway/image-bridge conflict points.

### Testing
- `go test ./internal/service ./internal/repository ./internal/handler ./internal/server/routes` passed from `backend/`.
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; existing Browserslist age and Vite large chunk warnings remain.
- `pnpm test:run src/stores/__tests__/app.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/ChatView.spec.ts` passed from `frontend/`; existing `--localstorage-file` and Browserslist warnings remain.
- `git diff --check` passed from the repository root.

### Notes
- `backend/internal/pkg/apicompat/types.go`: kept local nested chat reasoning support and added upstream `parallel_tool_calls`.
- `frontend/src/stores/app.ts`: kept local ChinaAPI public-setting defaults while adopting upstream in-flight request de-duplication.
- `backend/internal/service/openai_gateway_forward.go`: merged upstream Codex image-tool policy flow with local request-aware image bridge gating.
- `backend/internal/service/openai_gateway_service.go`: kept upstream Codex CLI version `0.144.1` and local stream-disconnect drain grace.
- `backend/internal/service/openai_image_generation_controls_test.go`: retained upstream namespace-strip coverage and local dangling image tool-choice/custom function tool tests.
- `docs/UPSTREAM_SYNC.md`: appended the `v0.1.150` sync range, conflict decisions, tests, and rollback points.
- `progress.md`: recorded this upstream sync and verification.
- Upstream-touched files: all changes from `12d811bd..6dd3274a` were merged by commit `3c9b9045`; see `docs/UPSTREAM_SYNC.md` for the business-level file groups and decisions.
- Rollback: reset to `backup/pre-upstream-sync-20260710-155249` to return to the pre-sync HEAD, or revert merge commit `3c9b9045` plus the post-merge conflict-resolution edits listed above. The pre-sync stash is `stash@{0}` with message `pre-upstream-sync-20260710-155249`.

## 2026-07-10 - Task: Restore API key chat action labels

### What was done
- Restored the missing Chinese and English labels for the API key chat action.
- The key list now displays a readable chat action instead of the raw `keys.chatWithKey` translation key.

### Testing
- `pnpm exec vitest run src/i18n/__tests__/localesNoKeyCollision.spec.ts` passed from `frontend/` (6 tests).
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `git diff --check -- frontend/src/i18n/locales/zh/dashboard.ts frontend/src/i18n/locales/en/dashboard.ts` passed from the repository root.
- Confirmed `keys.chatWithKey` is defined in both locale files and matches the key used by `KeysView.vue`.

### Notes
- `frontend/src/i18n/locales/zh/dashboard.ts`: restored the Chinese `keys.chatWithKey` label.
- `frontend/src/i18n/locales/en/dashboard.ts`: restored the English `keys.chatWithKey` label.
- `progress.md`: recorded this locale fix and its verification.
- Rollback: remove the two `chatWithKey` locale entries and this appended task record to restore the pre-task state.

## 2026-07-10 - Task: Restore table page-size regression coverage

### What was done
- Restored the page-size test to the established rule that an injected system default overrides stale browser storage.
- Removed the contradictory test expectation introduced while restoring local work after the upstream sync.

### Testing
- `pnpm exec vitest run src/composables/__tests__/usePersistedPageSize.spec.ts src/utils/__tests__/tablePreferences.spec.ts src/stores/__tests__/app.spec.ts` passed from `frontend/` (34 tests).
- `git diff --check -- frontend/src/composables/__tests__/usePersistedPageSize.spec.ts` passed from the repository root.

### Notes
- `frontend/src/composables/__tests__/usePersistedPageSize.spec.ts`: restored coverage for the configured-default precedence behavior implemented by `getPersistedPageSize`.
- `progress.md`: recorded this post-sync test regression fix.
- Rollback: restore the pre-task test title and expected value `50`, then remove this appended task record.

## 2026-07-10 - Task: Restore account import locale keys after i18n split

### What was done
- Restored five Chinese and English account-import messages omitted when the single locale files were split into domain modules.
- Added direct regression coverage for invalid payloads, ZIP import errors, mixed ZIP formats, and Codex import summaries.
- Compared the pre-sync locale leaf keys with the current assembled locales and confirmed no referenced migration omissions remain.

### Testing
- `pnpm exec vitest run src/i18n/__tests__/accountImportLocaleKeys.spec.ts src/i18n/__tests__/localesNoKeyCollision.spec.ts src/__tests__/integration/data-import.spec.ts` passed from `frontend/` (25 tests).
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- The structured pre-sync/current locale audit reported empty `migratedEnMissing` and `migratedZhMissing` lists.
- `git diff --check -- frontend/src/i18n/locales/zh/admin/accounts.ts frontend/src/i18n/locales/en/admin/accounts.ts frontend/src/i18n/__tests__/accountImportLocaleKeys.spec.ts` passed from the repository root.

### Notes
- `frontend/src/i18n/locales/zh/admin/accounts.ts`: restored five Chinese account-import messages from the pre-split locale.
- `frontend/src/i18n/locales/en/admin/accounts.ts`: restored the matching five English account-import messages.
- `frontend/src/i18n/__tests__/accountImportLocaleKeys.spec.ts`: added regression coverage that loads the real locale modules.
- `progress.md`: recorded this i18n migration fix and verification.
- Rollback: remove the five restored keys from each account locale, delete the new regression test, and remove this appended task record.

## 2026-07-10 - Task: Restore unit API contract repository compatibility

### What was done
- Updated the unit API contract usage-log repository stub for the new video-refund lookup contract.
- The stub now resolves usage rows by request ID and API key ID and returns the production not-found error when no row matches.

### Testing
- `go test -tags unit ./internal/server` passed from `backend/`.
- `go test ./internal/server` passed from `backend/`.
- `git diff --check -- backend/internal/server/api_contract_test.go` passed from the repository root.

### Notes
- `backend/internal/server/api_contract_test.go`: implemented `GetByRequestIDAndAPIKey` on `stubUsageLogRepo` so the unit-tag API contract suite compiles against `UsageLogRepository`.
- `progress.md`: recorded this post-sync unit test compatibility fix.
- Rollback: remove the added stub method and this appended task record to restore the pre-task state.

## 2026-07-10 - Task: Restore channel locale keys after upstream sync

### What was done
- Restored the Chinese and English timeout message used when a channel monitor check exceeds its request limit.
- Restored the Chinese and English billing hint explaining that video models in per-request mode are charged once per generation request.
- Added regression coverage that loads the real channel locale modules and verifies both keys.

### Testing
- `pnpm exec vitest run src/i18n/__tests__/channelLocaleKeys.spec.ts src/i18n/__tests__/localesNoKeyCollision.spec.ts` passed from `frontend/` (10 tests).
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `git diff --check -- frontend/src/i18n/locales/zh/admin/channels.ts frontend/src/i18n/locales/en/admin/channels.ts frontend/src/i18n/__tests__/channelLocaleKeys.spec.ts` passed from the repository root.

### Notes
- `frontend/src/i18n/locales/zh/admin/channels.ts`: added the Chinese monitor-timeout message and per-request video billing hint.
- `frontend/src/i18n/locales/en/admin/channels.ts`: added the matching English messages.
- `frontend/src/i18n/__tests__/channelLocaleKeys.spec.ts`: added direct regression coverage against both real locale modules.
- `progress.md`: recorded this post-sync locale repair and verification.
- Rollback: remove `runTimeout` and `perRequestVideoHint` from both channel locale modules, delete `frontend/src/i18n/__tests__/channelLocaleKeys.spec.ts`, and remove this appended task record.

## 2026-07-10 - Task: Restore image failover unit test fixture compatibility

### What was done
- Updated the image failover handler test fixture to provide the billing service now required by pre-request image cost estimation.
- Preserved the existing behavior assertions that two eligible accounts are attempted before a clear upstream error is returned.

### Testing
- `go test -tags unit ./internal/handler -run '^TestOpenAIGatewayHandlerImages_ServerErrorFailsOverAndReturnsClearErrorWhenExhausted$' -count=1` passed from `backend/`.
- `go test -tags unit ./internal/handler` passed from `backend/`.
- `git diff --check -- backend/internal/handler/openai_images_failover_test.go` passed from the repository root.

### Notes
- `backend/internal/handler/openai_images_failover_test.go`: injected `BillingService` into the gateway service test fixture so image estimation no longer stops the test before upstream failover.
- `progress.md`: recorded this post-sync test fixture repair and verification.
- Rollback: replace the test fixture's `service.NewBillingService(cfg, nil)` argument with `nil` and remove this appended task record.

## 2026-07-10 - Task: Isolate OpenAI video task cache entries by user

### What was done
- Added the user ID to OpenAI video task sticky-session hashes so identical upstream task IDs cannot share cached account bindings across users.
- Kept persistent task bindings as the fallback for existing tasks and updated the non-create video routing seed to use the same user-scoped hash.
- Strengthened the cross-user regression test by warming user A's cache before restoring the same task ID for user B.

### Testing
- `go test ./internal/service ./internal/handler -run 'OpenAIVideo|VideoTask' -count=1` passed from `backend/`.
- `go test -race ./internal/service -run 'Test(BindOpenAIVideoTaskAccount|RestoreOpenAIVideoTaskStickySession)' -count=1` passed from `backend/`.
- `git diff --check -- backend/internal/service/openai_videos.go backend/internal/handler/openai_videos.go backend/internal/service/openai_videos_test.go` passed from the repository root.

### Notes
- `backend/internal/service/openai_videos.go`: made cached video task session hashes user-scoped for bind and restore operations.
- `backend/internal/handler/openai_videos.go`: aligned status/content routing with the user-scoped task hash.
- `backend/internal/service/openai_videos_test.go`: added a populated-cache cross-user regression scenario and updated expected cache keys.
- `progress.md`: recorded the video task isolation fix and verification.
- Rollback: remove the user ID parameter from `OpenAIVideoTaskSessionHash`, restore task-only call sites and test keys, revert the populated-cache regression setup, and remove this appended task record.

## 2026-07-10 - Task: Close Nano Banana image billing gaps

### What was done
- Added request-count parsing for Nano Banana with a default of one and positive-integer validation.
- Added pre-request image cost estimation, balance/subscription coverage checks, and balance hold reserve/capture/release handling matching the standard OpenAI images path.
- Reconciled successful requests against the actual response image count and marked captured balances so usage recording cannot deduct the same amount twice.
- Expanded response accounting to recognize image arrays, single image objects, direct URLs, and base64 payloads, falling back to the requested count when a successful response has no recognized shape.

### Testing
- `go test ./internal/handler -run 'TestNanoBanana' -count=1` passed from `backend/`.
- `go test -race ./internal/service -run 'Test(ParseOpenAINanoBanana|ExtractNanoBanana)' -count=1` passed from `backend/`.
- `go test ./internal/handler ./internal/service` passed from `backend/`.
- `go vet ./internal/handler ./internal/service` passed from `backend/`.
- `git diff --check -- backend/internal/handler/openai_nano_banana.go backend/internal/handler/openai_nano_banana_billing_test.go backend/internal/service/openai_nano_banana.go backend/internal/service/openai_nano_banana_test.go` passed from the repository root.

### Notes
- `backend/internal/handler/openai_nano_banana.go`: added estimated-cost gating and transactional media balance hold settlement around upstream generation.
- `backend/internal/handler/openai_nano_banana_billing_test.go`: added handler-level insufficient-balance and reserve/capture/no-double-deduction regressions.
- `backend/internal/service/openai_nano_banana.go`: retained request count and expanded successful response image accounting with request-count fallback.
- `backend/internal/service/openai_nano_banana_test.go`: added request-count validation and response-shape coverage.
- `progress.md`: recorded the Nano Banana billing repair and verification.
- Rollback: remove the Nano Banana estimate/hold/capture flow, remove request-count and expanded response accounting, delete `backend/internal/handler/openai_nano_banana_billing_test.go`, restore the previous service tests, and remove this appended task record.

## 2026-07-10 - Task: Harden community chat WebSocket authentication

### What was done
- Removed community chat JWTs from WebSocket query strings and moved browser authentication to the established `Sec-WebSocket-Protocol` pattern using `sub2api-chat, jwt.<token>`.
- Configured the server to negotiate only the fixed `sub2api-chat` protocol so the JWT is not echoed in the handshake response.
- Replaced the unconditional Origin allowance with same-host validation plus explicit `cors.allowed_origins` support for cross-origin deployments.
- Updated both the chat page and sidebar unread socket, tests, generated wiring, and operational documentation.

### Testing
- `go test -race ./internal/handler -run 'Test(ExtractCommunityChat|IsAllowedCommunityChat|NormalizeCommunityChat)' -count=1` passed from `backend/`.
- `go test ./cmd/server ./internal/handler ./internal/server/routes` passed from `backend/`.
- `pnpm exec vitest run src/api/__tests__/communityChat.spec.ts` passed from `frontend/` (2 tests).
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `pnpm exec eslint src/api/communityChat.ts src/api/__tests__/communityChat.spec.ts src/views/user/CommunityChatView.vue src/components/layout/AppSidebar.vue` passed from `frontend/`.
- `git diff --check -- backend/internal/handler/community_chat_handler.go backend/internal/handler/community_chat_handler_test.go backend/cmd/server/wire_gen.go frontend/src/api/communityChat.ts frontend/src/api/__tests__/communityChat.spec.ts frontend/src/views/user/CommunityChatView.vue frontend/src/components/layout/AppSidebar.vue docs/COMMUNITY_CHAT.md` passed from the repository root.

### Notes
- `backend/internal/handler/community_chat_handler.go`: added subprotocol JWT extraction, fixed-protocol negotiation, and configured Origin validation.
- `backend/internal/handler/community_chat_handler_test.go`: covered subprotocol-only token extraction and allowed/disallowed Origin cases.
- `backend/cmd/server/wire_gen.go`: passed the full configuration to the community chat handler.
- `frontend/src/api/communityChat.ts`: split the token-free WebSocket URL from the authenticated protocol list.
- `frontend/src/api/__tests__/communityChat.spec.ts`: verified that URLs contain no JWT and protocols carry the prefixed token.
- `frontend/src/views/user/CommunityChatView.vue`: switched the chat socket to subprotocol authentication.
- `frontend/src/components/layout/AppSidebar.vue`: switched the unread-badge socket to subprotocol authentication.
- `docs/COMMUNITY_CHAT.md`: documented the new handshake and Origin configuration requirements.
- `progress.md`: recorded the WebSocket security fix and verification.
- Rollback: restore query-string JWT construction and extraction, restore the previous handler constructor and generated wire call, remove Origin/subprotocol tests and documentation changes, and remove this appended task record.

## 2026-07-10 - Task: Protect community chat attachments and clean up deleted files

### What was done
- Required a valid, current login for every community chat attachment read while preserving native browser image, video, and download streaming through a short-lived path-scoped HttpOnly Cookie.
- Added an authenticated attachment-session endpoint and frontend renewal flow that establishes the Cookie before loading chat history.
- Added a database active-message check so attachments from soft-deleted messages return 404 even when their local file could not be removed.
- Removed local attachment files after successful message deletion, while keeping deletion successful and logging unexpected filesystem cleanup failures.

### Testing
- `go test ./internal/handler ./internal/service ./internal/repository ./internal/server/routes` passed from `backend/`.
- `go test -race ./internal/handler -run 'Test(EstablishCommunityChat|ServeCommunityChat|DeleteCommunityChat|CommunityChatLocal|NormalizeCommunityChat|ExtractCommunityChat|IsAllowedCommunityChat)' -count=1` passed from `backend/`.
- `pnpm exec vitest run src/api/__tests__/communityChat.spec.ts` passed from `frontend/` (3 tests).
- `pnpm typecheck` passed from `frontend/`.
- `pnpm exec eslint src/api/communityChat.ts src/api/__tests__/communityChat.spec.ts src/views/user/CommunityChatView.vue` passed from `frontend/`.

### Notes
- `backend/internal/handler/community_chat_handler.go`: added attachment Cookie issuance, JWT/user-state validation, active-message enforcement, and safe local-file cleanup after deletion.
- `backend/internal/handler/community_chat_handler_test.go`: covered unauthenticated denial, restricted Cookie creation, active and deleted attachment reads, physical cleanup, and external URL rejection.
- `backend/internal/repository/community_chat_repo.go`: added the active attachment lookup against non-deleted messages.
- `backend/internal/service/community_chat.go`: exposed the attachment activity check through the chat service contract.
- `backend/internal/server/routes/community_chat.go`: registered the authenticated attachment-session endpoint.
- `frontend/src/api/communityChat.ts`: added the attachment-session API call.
- `frontend/src/api/__tests__/communityChat.spec.ts`: verified the attachment-session request contract.
- `frontend/src/views/user/CommunityChatView.vue`: established and renewed attachment authentication before loading history.
- `docs/COMMUNITY_CHAT.md`: documented attachment authentication, deletion behavior, and same-origin proxy requirements.
- `progress.md`: recorded the attachment security and lifecycle repair.
- Rollback: remove the attachment-session route/API/timer, restore unauthenticated `ServeUpload`, remove `IsAttachmentActive` from the service and repository contracts, remove post-delete file cleanup and the added tests/docs, then remove this appended task record.

## 2026-07-10 - Task: Align OAuth registration tests and clear auth store lint regressions

### What was done
- Updated the pending OAuth account-creation regression test to match the current request contract, including the affiliate code and excluding omitted optional fields.
- Simplified the current-user refresh request deduplication flow by removing an ineffective rethrow-only wrapper and using an immutable request Promise without changing stale-response or 401 handling.

### Testing
- `pnpm exec vitest run src/views/auth/__tests__/EmailVerifyView.spec.ts src/stores/__tests__/auth.spec.ts` passed from `frontend/` (29 tests).
- `pnpm exec eslint src/stores/auth.ts src/views/auth/__tests__/EmailVerifyView.spec.ts` passed from `frontend/`.
- `pnpm typecheck` passed from `frontend/`.

### Notes
- `frontend/src/views/auth/__tests__/EmailVerifyView.spec.ts`: aligned the expected pending-account payload with the affiliate-code registration data.
- `frontend/src/stores/auth.ts`: removed the no-op catch and made the refresh request Promise immutable while preserving cleanup and deduplication.
- `progress.md`: recorded the OAuth contract and auth lint repair.
- Rollback: restore the previous pending-account payload expectation and the previous `refreshUser` try/catch plus mutable Promise declaration, then remove this appended task record.

## 2026-07-10 - Task: Remove service test global-state races

### What was done
- Set Gin test mode once in the service package `TestMain` before any tests start.
- Removed 429 repeated `gin.SetMode(gin.TestMode)` calls from 58 service test files so parallel tests no longer write Gin package globals concurrently.
- Made the asynchronous dashboard recompute test stub use an atomic call counter, removing the next race exposed by the full package race run.

### Testing
- `go test ./internal/service` passed from `backend/`.
- `go test -race ./internal/service -run 'TestUsageCleanupServiceExecuteTaskDashboardRecompute(Error|Success)' -count=1` passed from `backend/`.
- `go test -race ./internal/service` passed from `backend/` in 58.126 seconds.
- `go vet ./internal/service` passed from `backend/`.
- `git diff --check` passed from the repository root.

### Notes
- `backend/internal/service/main_test.go`: added the package-level one-time Gin test-mode setup.
- `backend/internal/service/usage_cleanup_service_test.go`: made asynchronous recompute call tracking atomic.
- `backend/internal/service/account_credential_shadow_skip_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/account_test_service_gemini_test.go`: removed the redundant Gin mode write and now-unused import.
- `backend/internal/service/account_test_service_grok_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/account_test_service_openai_compact_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/account_test_service_openai_image_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/account_test_service_openai_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/antigravity_gateway_service_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/error_passthrough_runtime_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gateway_anthropic_apikey_passthrough_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gateway_anthropic_vertex_beta_filter_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gateway_anthropic_vertex_service_account_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gateway_context_management_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gateway_forward_as_chat_completions_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gateway_forward_as_responses_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gateway_non_streaming_response_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gateway_service_streaming_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gateway_streaming_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gemini_error_policy_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/gemini_messages_compat_service_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_client_restriction_detector_hardening_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_client_restriction_detector_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_client_transport_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_compact_model_mapping_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_compact_stream_bridge_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_compat_model_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_cyber_policy_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_cyber_session_block_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_embeddings_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_failover_cached_body_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_fast_policy_ws_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_chat_completions_raw_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_chat_completions_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_compat_cyber_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_count_tokens_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_grok_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_messages_chat_fallback_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_messages_failed_response_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_messages_transport_failover_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_passthrough_function_args_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_response_failed_passthrough_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_responses_chat_fallback_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_service_codex_cli_only_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_service_hotpath_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gateway_service_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_gpt56_max_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_image_generation_controls_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_images_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_oauth_passthrough_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_reasoning_effort_candidates_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_upstream_transport_error_handle_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_videos_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_ws_forwarder_success_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_ws_http_bridge_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_ws_protocol_forward_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/openai_ws_ratelimit_signal_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/ops_metrics_collector_test.go`: removed redundant per-test Gin mode writes.
- `backend/internal/service/ratelimit_service_openai_image_test.go`: removed redundant per-test Gin mode writes.
- `progress.md`: recorded the service race-test infrastructure repair.
- Rollback: delete `backend/internal/service/main_test.go`, restore each removed per-test `gin.SetMode(gin.TestMode)` call and the Gemini test import, restore the non-atomic dashboard stub counter and assertions, then remove this appended task record.

## 2026-07-10 - Task: Close community attachment cache and unmount races

### What was done
- Marked authenticated attachment responses as private and non-cacheable so revoked sessions and deleted messages cannot be bypassed through browser or intermediary caches.
- Disabled MIME sniffing for uploaded attachment responses.
- Prevented an in-flight attachment-session initialization from starting renewal timers, history loading, or WebSockets after the chat component has unmounted.

### Testing
- `go test -race ./internal/handler -run 'Test(EstablishCommunityChat|ServeCommunityChat|DeleteCommunityChat|CommunityChatLocal)' -count=1` passed from `backend/`.
- `pnpm typecheck` passed from `frontend/`.
- `pnpm exec eslint src/views/user/CommunityChatView.vue` passed from `frontend/`.
- `git diff --check` passed from the repository root.

### Notes
- `backend/internal/handler/community_chat_handler.go`: added private no-store caching and MIME nosniff headers to authenticated attachment responses.
- `backend/internal/handler/community_chat_handler_test.go`: verified both attachment response security headers.
- `frontend/src/views/user/CommunityChatView.vue`: stopped asynchronous chat initialization after component unmount.
- `docs/COMMUNITY_CHAT.md`: documented attachment cache and MIME response protections.
- `progress.md`: recorded the final attachment lifecycle hardening.
- Rollback: remove the two attachment response headers and their assertions, remove the component unmount guard checks, restore the previous documentation, and remove this appended task record.

## 2026-07-11 - Task: Probe distinct upstream accounts in channel monitoring

### What was done
- Replaced three independent OpenAI/Grok monitor HTTP retries with one gateway request that can fail over across at most three different eligible accounts in the bound group.
- Reused the gateway's failed-account exclusion set for OAuth, Setup Token, and API Key accounts without distinguishing account types.
- Disabled same-account pool retries for monitor probes, while preserving the configured failover behavior for ordinary API traffic.
- Kept account model mappings authoritative: accounts that do not declare support for the requested model are not probed.
- Extended the monitor request timeout budget to cover up to three sequential upstream account attempts.

### Testing
- `go test -tags=unit ./internal/service -count=1` passed from `backend/`.
- `go test ./internal/handler` passed from `backend/`.
- `go test ./internal/server/routes` passed from `backend/`.
- `git diff --check -- backend/internal/service/channel_monitor_const.go backend/internal/service/channel_monitor_checker.go backend/internal/service/channel_monitor_checker_body_test.go backend/internal/handler/channel_monitor_probe.go backend/internal/handler/channel_monitor_probe_test.go backend/internal/handler/openai_chat_completions.go backend/internal/handler/openai_gateway_handler.go docs/UPSTREAM_SYNC.md` passed from the repository root.

### Notes
- `backend/internal/service/channel_monitor_const.go`: defines the three-account probe contract and the combined monitor request timeout budget.
- `backend/internal/service/channel_monitor_checker.go`: sends one protected probe-attempt header instead of repeating independent requests.
- `backend/internal/service/channel_monitor_checker_body_test.go`: verifies one-request behavior, the probe header, and protection from template header overrides.
- `backend/internal/handler/channel_monitor_probe.go`: converts the monitor marker into a two-switch, distinct-account failover policy.
- `backend/internal/handler/channel_monitor_probe_test.go`: covers ordinary requests, the three-account cap, lower configured caps, and invalid marker values.
- `backend/internal/handler/openai_chat_completions.go`: applies distinct-account monitor failover to Chat Completions probes.
- `backend/internal/handler/openai_gateway_handler.go`: applies distinct-account monitor failover to Responses and Messages probes.
- `docs/UPSTREAM_SYNC.md`: documents distinct-account probing, external endpoint behavior, and the model-mapping prerequisite.
- `progress.md`: records this channel-monitor routing change and verification evidence.
- Rollback: reverse the unstaged hunks in the tracked files listed above with `git diff -- <tracked-files> | git apply -R`, delete `backend/internal/handler/channel_monitor_probe.go` and `backend/internal/handler/channel_monitor_probe_test.go`, then remove this final task block from `progress.md`; this leaves the pre-existing staged changes intact.

## 2026-07-11 - Task: Sync upstream Wei-Shaw/sub2api to v0.1.151

### What was done
- Synced upstream `Wei-Shaw/sub2api` from `6dd3274a` to `e316ebf5` while preserving the existing staged, unstaged, and untracked ChinaAPI work.
- Created backup branch `backup/pre-upstream-sync-20260711-020911` and stash `pre-upstream-sync-20260711-020911` before creating merge commit `5c5adad0`.
- Merged Codex custom/tool_search/namespace bridge support, Codex outbound identity pairing, user-scoped Fast/Flex policy, Anthropic cache-creation usage accounting, Grok reasoning-effort preservation, and ops capture-writer nil guards.
- Adapted the new upstream Fast/Flex middleware test to the fork's additional OpenAI video task binding repository constructor dependency.

### Testing
- `go test ./internal/pkg/apicompat ./internal/pkg/openai ./internal/server/middleware ./internal/service ./internal/handler ./internal/repository ./internal/server/routes` passed in full after the compatibility fix; the initial run exposed the fork constructor compatibility gap.
- `go test ./internal/server/middleware` passed after the compatibility fix.
- `go test ./internal/pkg/apicompat ./internal/pkg/openai ./internal/server/middleware -count=1` passed.
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `pnpm test:run src/views/admin/__tests__/SettingsView.spec.ts src/stores/__tests__/app.spec.ts` passed from `frontend/` (47 tests); existing local-storage, unresolved test `router-link`, and Browserslist warnings remain.
- `pnpm build` passed from `frontend/`; existing Browserslist age and Vite large chunk warnings remain.
- `git diff --check` and `git diff --cached --check` passed before the final log append.

### Notes
- `backend/cmd/server/VERSION`: advanced the embedded upstream version to `v0.1.151`.
- `backend/internal/handler/admin/admin_helpers_test.go`: adopted the upstream settings helper expectation.
- `backend/internal/handler/dto/settings.go`: exposed user-scoped Fast/Flex rule fields.
- `backend/internal/handler/ops_capture_writer_nil_test.go`: added coverage for post-release capture-writer access.
- `backend/internal/handler/ops_error_logger.go`: added complete nil guards to capture-writer delegation.
- `backend/internal/pkg/apicompat/anthropic_to_responses_response.go`: propagated Anthropic cache-creation usage into Responses output.
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`: added custom, tool_search, and namespace request/response bridge handling.
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_custom_tools_test.go`: covered Codex custom and namespaced tool bridge behavior.
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`: aligned bridge expectations with the expanded tool conversion.
- `backend/internal/pkg/apicompat/responses_anthropic_cache_creation_test.go`: covered cache-creation usage conversion.
- `backend/internal/pkg/apicompat/responses_stream_event_wire.go`: retained custom and namespace tool fields in stream events.
- `backend/internal/pkg/apicompat/responses_stream_event_wire_test.go`: covered the expanded stream-event wire fields.
- `backend/internal/pkg/apicompat/responses_to_anthropic.go`: propagated cache-creation usage to Anthropic responses.
- `backend/internal/pkg/apicompat/types.go`: added the upstream custom, namespace, tool-search, and usage fields while preserving fork fields.
- `backend/internal/pkg/ctxkey/ctxkey.go`: added trusted API-key user context support.
- `backend/internal/pkg/openai/request.go`: added Codex User-Agent and originator identity pairing.
- `backend/internal/pkg/openai/request_identity_test.go`: covered official and fallback Codex identity pairs.
- `backend/internal/server/middleware/api_key_auth.go`: forwarded the authenticated user ID into gateway context.
- `backend/internal/server/middleware/api_key_auth_test.go`: covered authenticated user context propagation.
- `backend/internal/server/middleware/openai_fast_policy_forwarding_test.go`: added upstream user-scoped policy coverage and supplied the fork video-binding repository argument.
- `backend/internal/service/account_test_service.go`: aligned account tests with final Codex identities.
- `backend/internal/service/account_usage_service.go`: aligned usage probes with final Codex identities.
- `backend/internal/service/openai_codex_identity.go`: centralized final outbound Codex identity enforcement.
- `backend/internal/service/openai_codex_identity_test.go`: covered final outbound identity enforcement.
- `backend/internal/service/openai_fast_policy_test.go`: covered user-scoped rule matching and precedence.
- `backend/internal/service/openai_fast_policy_ws_test.go`: covered user identity propagation through WebSocket policy evaluation.
- `backend/internal/service/openai_gateway_forward.go`: enforced the final Codex identity before forwarding.
- `backend/internal/service/openai_gateway_grok.go`: preserved compatible Grok reasoning effort.
- `backend/internal/service/openai_gateway_grok_test.go`: covered Grok reasoning-effort compatibility.
- `backend/internal/service/openai_gateway_messages.go`: marked messages bridge requests before identity enforcement.
- `backend/internal/service/openai_gateway_messages_chat_fallback.go`: passed expanded tool metadata through fallback conversion.
- `backend/internal/service/openai_gateway_passthrough.go`: enforced paired Codex identity on passthrough requests.
- `backend/internal/service/openai_gateway_request_body.go`: evaluated Fast/Flex rules against the trusted user ID.
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`: restored custom, namespace, and tool-search calls on fallback responses.
- `backend/internal/service/openai_gateway_service_test.go`: covered the expanded gateway identity and bridge behavior.
- `backend/internal/service/openai_oauth_passthrough_test.go`: aligned passthrough identity expectations.
- `backend/internal/service/openai_ws_forwarder_payload.go`: enforced paired Codex identity for WebSocket payloads.
- `backend/internal/service/openai_ws_forwarder_success_test.go`: covered WebSocket identity behavior.
- `backend/internal/service/setting_features.go`: added user IDs to Fast/Flex policy rules.
- `backend/internal/service/settings_view.go`: exposed the expanded Fast/Flex settings view.
- `frontend/src/api/admin/settings.ts`: added user IDs to the Fast/Flex settings contract.
- `frontend/src/i18n/locales/en/admin/settings.ts`: added English user-scope settings text.
- `frontend/src/i18n/locales/zh/admin/settings.ts`: added Chinese user-scope settings text.
- `frontend/src/views/admin/SettingsView.vue`: added user selection controls for Fast/Flex policy rules.
- `docs/UPSTREAM_SYNC.md`: recorded the sync range, behavior, verification, and rollback points.
- `progress.md`: recorded this upstream synchronization and its verification evidence.
- Rollback: use `backup/pre-upstream-sync-20260711-020911` to locate pre-sync commit `3c9b9045`; preserve or reapply stash object `ae7ed01490bb2d5576e544e2a26022a3fee3b851` for the pre-existing worktree, revert merge commit `5c5adad0`, remove the middleware test constructor compatibility argument, and remove these appended documentation records.

## 2026-07-11 - Task: Redesign the homepage as an AI Agent research experience

### What was done
- Replaced the split marketing hero and framed decorative orb with a full-bleed Agent orchestration scene that gives the first viewport a clear AI research subject.
- Connected a central orchestrator to reasoning, tools, memory, and evaluation nodes, with visible data packets, pointer-responsive perspective, and a reduced-motion static state.
- Added a live four-stage execution trace and compact runtime metrics so the animation communicates task progress instead of ambient movement.
- Replaced the repeated feature cards and topic console with an unframed research-track band covering Agent systems, multimodal intelligence, tools/protocols, and evaluation.
- Added desktop, tablet, short-screen, and mobile layout rules that keep the next section visible and progressively remove secondary labels before they can overlap primary content.

### Testing
- `pnpm typecheck` passed from `frontend/`.
- `pnpm exec eslint src/views/HomeView.vue src/components/home/HomeThreeScene.vue src/i18n/locales/zh/landing.ts src/i18n/locales/en/landing.ts` passed from `frontend/`.
- `pnpm test:run src/stores/__tests__/app.spec.ts src/router/__tests__/title.spec.ts` passed from `frontend/` (32 tests); the existing local-storage warning remains.
- `pnpm build` passed from `frontend/`; the existing Browserslist age and Vite large chunk warnings remain.
- `curl -I http://127.0.0.1:3000/home` returned HTTP 200 from the running Vite server.
- `git diff --check -- frontend/src/views/HomeView.vue frontend/src/components/home/HomeThreeScene.vue frontend/src/i18n/locales/zh/landing.ts frontend/src/i18n/locales/en/landing.ts` passed.
- Browser screenshots, visual overlap inspection, interaction clicks, and Canvas pixel-difference checks were not run because the Browser runtime reported no available browser instances. This remains the visual QA gap.

### Notes
- `frontend/src/views/HomeView.vue`: rebuilt the first viewport, live Agent trace, research tracks, color system, and responsive states while preserving custom home content and existing navigation actions.
- `frontend/src/components/home/HomeThreeScene.vue`: replaced the generic orbiting core with the interactive orchestrator graph, four capability nodes, curved connections, moving packets, deterministic background points, and resource cleanup.
- `frontend/src/i18n/locales/zh/landing.ts`: added Chinese Agent network, execution trace, node, metric, and research-track copy.
- `frontend/src/i18n/locales/en/landing.ts`: added matching English Agent network, execution trace, node, metric, and research-track copy.
- `docs/HOME_THREE_SCENE.md`: documented the new scene composition, responsive behavior, and required visual verification.
- `progress.md`: recorded the homepage redesign, verification evidence, and remaining browser QA gap.
- Rollback: reverse the unstaged hunks in the six tracked or existing task files listed above, restore the previous `HomeThreeScene.vue` content, and remove this appended task record; do not alter unrelated staged homepage or locale changes.

## 2026-07-11 - Task: Align the login experience with the AI Agent homepage

### What was done
- Replaced the centered red-gradient glass-card authentication shell with a desktop split layout that continues the homepage Agent network into a full-height visual band and uses a quiet authentication workspace for forms.
- Added tablet and mobile states that collapse the Agent scene into a compact top band while preserving form width, scrolling, locale access, and a direct route back to the homepage.
- Restyled the login title, email/password fields, password visibility control, forgot-password action, loading state, primary action, OAuth divider, provider buttons, and registration footer without changing authentication logic.
- Kept Turnstile, login-agreement gating, 2FA, OAuth providers, registration, recovery, verification, callback routes, and post-login redirects unchanged.

### Testing
- `pnpm typecheck` passed from `frontend/`.
- `pnpm exec eslint src/components/layout/AuthLayout.vue src/views/auth/LoginView.vue src/i18n/locales/zh/common.ts src/i18n/locales/en/common.ts` passed from `frontend/`.
- `pnpm test:run src/views/auth/__tests__ src/components/auth/__tests__` passed from `frontend/` (10 files, 90 tests); the existing local-storage and Browserslist warnings remain.
- `pnpm build` passed from `frontend/`; the existing Browserslist age and Vite large chunk warnings remain.
- `curl` checks returned HTTP 200 for both `http://127.0.0.1:3000/home` and `http://127.0.0.1:3000/login`.
- `git diff --check -- frontend/src/components/layout/AuthLayout.vue frontend/src/views/auth/LoginView.vue frontend/src/i18n/locales/zh/common.ts frontend/src/i18n/locales/en/common.ts` passed.
- Browser screenshots, visual overlap inspection, form interaction clicks, and Canvas pixel checks were not run because the Browser runtime again reported no available browser instances. This remains the visual QA gap.

### Notes
- `frontend/src/components/layout/AuthLayout.vue`: added the responsive Agent visual band, shared authentication workspace, home navigation, locale control, and light/dark layout states.
- `frontend/src/views/auth/LoginView.vue`: added the login-specific hierarchy, field, action, loading, OAuth, link, responsive, and accessibility styles while preserving handlers and state.
- `frontend/src/i18n/locales/zh/common.ts`: added Chinese secure-access, workspace, pipeline, and password-visibility text.
- `frontend/src/i18n/locales/en/common.ts`: added matching English authentication experience text.
- `docs/HOME_THREE_SCENE.md`: documented that the Agent canvas is shared by homepage and authentication routes.
- `docs/AUTH_EXPERIENCE.md`: documented layout behavior, login control behavior, breakpoints, and verification commands.
- `.gitignore`: allowed the authentication experience document through the repository's explicit `docs/` whitelist.
- `frontend/src/views/auth/VISUAL_GUIDE.md`: replaced obsolete glass-card diagrams with a pointer to the maintained authentication guide.
- `progress.md`: recorded the authentication visual alignment, verification evidence, and browser QA gap.
- Rollback: restore the prior `AuthLayout.vue`, reverse the login-only template/style changes and the four authentication locale additions, restore the previous auth visual guide, remove `docs/AUTH_EXPERIENCE.md` and its `.gitignore` whitelist entry, restore the homepage-only wording in `docs/HOME_THREE_SCENE.md`, and remove this appended task record.

## 2026-07-11 - Task: Restore community chat notifications after refresh

### What was done
- Persisted separate group-chat and direct-message read cursors for each signed-in user in the current browser.
- Restored the sidebar unread indicator after refresh by checking the latest public message and the latest direct message visible to the current user.
- Extended real-time direct-message reminders to both conversation participants, including administrators receiving new owner-contact messages.
- Kept the existing WebSocket delivery path and database schema unchanged.

### Testing
- `pnpm exec vitest run src/api/__tests__/communityChat.spec.ts` passed from `frontend/` (4 tests).
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `pnpm exec eslint src/api/communityChat.ts src/api/__tests__/communityChat.spec.ts src/components/layout/AppSidebar.vue` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; the existing Browserslist age and Vite large chunk warnings remain.
- `git diff --check -- frontend/src/api/communityChat.ts frontend/src/api/__tests__/communityChat.spec.ts frontend/src/components/layout/AppSidebar.vue docs/COMMUNITY_CHAT.md progress.md` passed from the repository root.

### Notes
- `frontend/src/api/communityChat.ts`: added per-user, per-scope monotonic read-cursor helpers.
- `frontend/src/api/__tests__/communityChat.spec.ts`: added regression coverage for cursor isolation and monotonic updates.
- `frontend/src/components/layout/AppSidebar.vue`: restores unread state from persisted messages and handles group/direct WebSocket events consistently.
- `docs/COMMUNITY_CHAT.md`: documented refresh recovery and the current-browser scope of read positions.
- `progress.md`: recorded this notification repair and its verification evidence.
- Rollback: reverse the notification-cursor additions in the four implementation/documentation files above and remove this appended task record; no database rollback is required.

## 2026-07-11 - Task: Forward Gemini native APIs through opted-in OpenAI accounts

### What was done
- Added a default-off Gemini native endpoint capability for OpenAI APIKey accounts without changing their platform, group, credentials, or existing OpenAI routes.
- Routed Gemini native model metadata, synchronous generation, streaming generation, and token-count requests through capability-filtered OpenAI account scheduling.
- Forwarded compatible requests to the account base URL with Bearer authentication, preserved Gemini request bodies and SSE responses, retained account model mapping, and normalized legacy `/v1` base URL suffixes.
- Kept existing Gemini account forwarding unchanged and recorded successful OpenAI-account passthrough usage against the original OpenAI group/account with `/v1beta/models` as the upstream endpoint.
- Added the management UI capability control and operator documentation for configuring compatible upstreams such as `https://sub.g-aisc.com`.

### Testing
- `go test -tags=unit ./internal/service ./internal/handler ./cmd/server` passed from `backend/`.
- `go test -tags=unit ./internal/service -run 'Test(OpenAIGeminiNativeCapabilityDefaultsOff|ForwardNativeOpenAI|ForwardAIStudioGETOpenAI)' -count=1` passed from `backend/` after the final service-level capability guard was added.
- `go test ./cmd/server` passed from `backend/` after the final service-level capability guard was added.
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `pnpm exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts` passed from `frontend/` (26 tests); the existing local-storage argument and Browserslist age warnings remain.
- `git diff --check` passed for all files changed by this task.

### Notes
- `.gitignore`: allowed the Gemini native OpenAI-account guide through the repository's explicit `docs/` whitelist.
- `backend/cmd/server/wire_gen.go`: injected the existing OpenAI gateway service into the shared gateway handler.
- `backend/internal/handler/gateway_handler.go`: stored the OpenAI gateway service required for capability-aware account selection.
- `backend/internal/handler/gemini_v1beta_handler.go`: accepted OpenAI groups only through the new capability-aware scheduling path and recorded the real Gemini upstream endpoint.
- `backend/internal/service/account.go`: added the default-off `gemini_native` OpenAI endpoint capability.
- `backend/internal/service/openai_gateway_scheduling.go`: added OpenAI account selection methods that require the Gemini native capability.
- `backend/internal/service/gemini_messages_compat_service.go`: added Bearer-authenticated Gemini native POST/GET passthrough while preserving existing Gemini account behavior.
- `backend/internal/service/gemini_openai_native_test.go`: covered capability defaults, Bearer auth, URL construction, body preservation, SSE, GET forwarding, and rejection paths.
- `frontend/src/types/index.ts`: added the Gemini native OpenAI endpoint capability type.
- `frontend/src/components/account/CreateAccountModal.vue`: added the default-off capability to OpenAI APIKey account creation.
- `frontend/src/components/account/EditAccountModal.vue`: loaded and saved the new capability for existing OpenAI APIKey accounts.
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`: verified that enabling the capability persists it in account credentials.
- `frontend/src/i18n/locales/en/admin/accounts.ts`: added English capability text and behavior guidance.
- `frontend/src/i18n/locales/zh/admin/accounts.ts`: added Chinese capability text and behavior guidance.
- `docs/OPENAI_GEMINI_NATIVE.md`: documented supported endpoints, account configuration, authentication, boundaries, and a verification request.
- `progress.md`: recorded this compatibility implementation and verification evidence.
- Rollback: reverse only the Gemini-native-related hunks in the files listed above, remove `backend/internal/service/gemini_openai_native_test.go` and `docs/OPENAI_GEMINI_NATIVE.md`, remove the matching `.gitignore` whitelist entry, and remove this appended task record; no database rollback is required.

## 2026-07-11 - Task: Reorganize the public image and video API documentation

### What was done
- Reorganized the user-facing API reference into four model families: International Jimeng SD2.0, Gemini Nano Banana Pro, Grok image/video, and OpenAI image/video.
- Grouped `video-ds-2.0`, `video-ds-2.0-fast`, `as-sd2.0-fast`, and `video-v1-15s` together while clearly separating their two request paths and parameter formats.
- Added Gemini native `gemini-3-pro-image-preview` generateContent and reference-image examples beside the existing `nano-banana-pro` JSON interface.
- Added public request examples and limits for `grok-imagine-image-quality`, `grok-imagine-video-1.5`, and `grok-imagine-video`.
- Grouped `gpt-image-2` and the Sora 2 variants under the OpenAI image/video family and removed duplicate language examples that did not add user-facing information.

### Testing
- `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts` passed from `frontend/` (1 test).
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `pnpm exec eslint src/views/user/ApiDocsView.vue src/views/user/__tests__/ApiDocsView.spec.ts` passed from `frontend/` with no warnings.
- `pnpm build` passed from `frontend/`; the existing Browserslist age and large chunk warnings remain.
- `git diff --check -- frontend/src/views/user/ApiDocsView.vue frontend/src/views/user/__tests__/ApiDocsView.spec.ts docs/API_DOCS.md progress.md` passed.
- Browser visual inspection was not run because the local browser policy blocked access to the development URL.

### Notes
- `frontend/src/views/user/ApiDocsView.vue`: replaced model-by-model sections with four user-facing model families, concise parameter guidance, and directly usable curl examples.
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: verifies the four family headings, required model names, and key public endpoints.
- `docs/API_DOCS.md`: documents the maintained public-page scope, model grouping, protocol differences, and concise billing behavior.
- `progress.md`: recorded the API documentation reorganization and verification evidence.
- Rollback: remove `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`, restore the previous `frontend/src/views/user/ApiDocsView.vue` and `docs/API_DOCS.md` contents from the pre-task worktree state, and remove this appended task record; no backend or database rollback is required.

## 2026-07-11 - Task: Update and improve the model chat experience

### What was done
- Replaced the model chat choices with `gpt-5.6-sol` and `gpt-5.6-terra`, made Terra the default, and migrated legacy local GPT-5.5/GPT-5.4/GPT-5.4 Mini selections.
- Updated deep thinking for GPT-5.6 to use `max` while retaining `medium` as the default reasoning effort.
- Locked session navigation and request-affecting configuration during streaming so the visible settings stay aligned with the active request.
- Prevented IME composition Enter events from sending early and removed empty assistant placeholders when a response is stopped before content arrives.
- Preserved the reader's scroll position during long streaming replies and added a compact jump-to-latest control.

### Testing
- `pnpm exec eslint src/api/chat.ts src/components/common/Toggle.vue src/composables/useLocalChat.ts src/views/user/ChatView.vue src/views/user/__tests__/ChatView.spec.ts src/i18n/locales/zh/misc.ts src/i18n/locales/en/misc.ts` passed from `frontend/`.
- `pnpm exec vitest run src/views/user/__tests__/ChatView.spec.ts` passed from `frontend/` (18 tests).
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; the existing Browserslist age and Vite large chunk warnings remain.
- `git diff --check` passed for all files changed by this task.
- The development server returned HTTP 200 for `/chat`, but browser interaction was redirected to the real login page. Chat-page visual verification remains pending because the local administrator password is unknown and no login or password reset was authorized.

### Notes
- `.gitignore`: allowed the maintained model chat guide through the repository's explicit `docs/` whitelist.
- `frontend/src/api/chat.ts`: allowed the GPT-5.6 `max` reasoning effort in chat request types.
- `frontend/src/components/common/Toggle.vue`: added a backward-compatible disabled state used while a chat request is active.
- `frontend/src/composables/useLocalChat.ts`: added targeted message removal for stopped empty responses.
- `frontend/src/views/user/ChatView.vue`: updated models and reasoning, migrated legacy selections, locked active request configuration, fixed IME and stop behavior, and added reader-controlled streaming scroll.
- `frontend/src/views/user/__tests__/ChatView.spec.ts`: covered model replacement and migration, IME safety, streaming locks, empty-stop cleanup, and scroll-follow behavior.
- `frontend/src/i18n/locales/zh/misc.ts`: updated Chinese reasoning guidance and added the jump-to-latest label.
- `frontend/src/i18n/locales/en/misc.ts`: updated matching English chat copy.
- `docs/MODEL_CHAT.md`: documented supported models, migration, streaming interaction rules, reasoning effort, and local data scope.
- `progress.md`: recorded this model chat improvement and its verification evidence.
- Rollback: reverse the model-chat hunks in the files listed above, remove `docs/MODEL_CHAT.md` and its `.gitignore` whitelist entry, and remove this appended task record; no database rollback is required.

## 2026-07-11 - Task: Complete public image and video API examples

### What was done
- Renamed `video-v1-15s` as an SD2.0 fast alias and added separate single-image and ordered multi-image transition examples, including the 8-image limit and image/prompt guidance.
- Added a simplified Nano Banana Pro JSON reference-image example alongside the existing Gemini native text and reference-image examples.
- Added Grok quality image editing and multi-reference video examples with their real JSON fields and reference limits.
- Completed the Sora 2 workflow with an authenticated MP4 content download example after the task reaches `completed`.

### Testing
- `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts` passed from `frontend/` (1 test).
- `pnpm exec vue-tsc --noEmit` passed from `frontend/`.
- `pnpm exec eslint src/views/user/ApiDocsView.vue src/views/user/__tests__/ApiDocsView.spec.ts` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; the existing Browserslist age and Vite large chunk warnings remain.
- The fresh development server returned HTTP 200 for `http://127.0.0.1:3001/api-docs`.
- Browser visual verification remained inconclusive because the existing port-3000 Vite process returned an empty app root; the fresh port-3001 server was started after that browser check.
- Whitespace checks passed for the task files.

### Notes
- `frontend/src/views/user/ApiDocsView.vue`: added the four missing user-facing request/download examples and corrected related model, endpoint, parameter, and limit text.
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: locked the new example titles, endpoint paths, JSON fields, download filename, and per-family example counts.
- `docs/API_DOCS.md`: documented single/multi-reference field differences and the complete Sora download workflow.
- `progress.md`: recorded this API documentation completion and its verification evidence.
- Rollback: reverse the latest API documentation hunks in the three files above and remove this appended task record; no backend or database rollback is required.

## 2026-07-12 - Task: Merge upstream PR #4009 Grok integration fixes

### What was done
- Merged `Wei-Shaw/sub2api#4009` at head `674d1a25` into local `main` as merge commit `09791722`.
- Added Grok CLI transport identity, OAuth subscription-proxy routing, xAI API-key account support, Responses payload sanitization, Composer/Build/4.20 fallback pricing, remaining-capacity quota display, and Grok CLI/OpenCode setup output.
- Preserved and restored all pre-existing staged, unstaged, and untracked work; all eight overlapping files merged without unresolved conflicts.
- Recorded the synchronization behavior, validation, backup branch, stash object, and rollback command in the upstream synchronization guide.

### Testing
- `go test ./internal/repository ./internal/service` passed from `backend/`.
- `go test ./...` passed from `backend/`.
- `go mod tidy -diff` passed from `backend/` with no dependency diff.
- The frontend Vitest run passed the complete suite: 147 test files and 967 tests; existing local-storage, Browserslist, Vue component-resolution, and i18n test warnings remain.
- `pnpm typecheck` passed from `frontend/`.
- `pnpm lint:check` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; the existing Browserslist age and Vite large-chunk warnings remain.
- `git diff --check backup/pre-pr-4009-merge-20260712-142050..HEAD` passed.
- The worktree item count remained 243 after restoration, and the untracked-file fingerprint remained `e541d9523f674fcda097257af4994334c2fa61738d8e3b2a13815466af6e5a43`.

### Notes
- `README.md`: documented the supported Grok API-key, Grok CLI, and OpenCode setup paths.
- `backend/go.mod`: updated the `golang.org/x/mod` dependency required for CLI version validation.
- `backend/internal/handler/admin/grok_oauth_handler_test.go`: aligned the Grok OAuth account expectation with the subscription proxy.
- `backend/internal/repository/http_upstream.go`: applies Grok CLI identity and a validated client version at the final transport boundary.
- `backend/internal/repository/http_upstream_test.go`: covers transport identity, version fallback, injection rejection, and host scoping.
- `backend/internal/service/account.go`: routes OAuth accounts away from official credit API URLs while preserving custom upstreams.
- `backend/internal/service/account_base_url_test.go`: covers OAuth URL normalization and API-key/custom-host behavior.
- `backend/internal/service/account_test_service.go`: supports Grok OAuth and API-key connection tests.
- `backend/internal/service/account_test_service_grok_test.go`: updates the Grok connection-test base URL expectation.
- `backend/internal/service/billing_service.go`: adds non-zero Grok cached-input prices and current model aliases.
- `backend/internal/service/billing_service_test.go`: verifies Grok fallback pricing and aliases.
- `backend/internal/service/grok_oauth_service.go`: defaults newly built OAuth credentials to the CLI subscription proxy.
- `backend/internal/service/grok_oauth_service_test.go`: verifies the OAuth credential proxy default.
- `backend/internal/service/grok_quota_service_test.go`: aligns quota probing with the CLI proxy.
- `backend/internal/service/openai_gateway_grok.go`: supports API-key Responses forwarding and sanitizes unsupported Grok inputs.
- `backend/internal/service/openai_gateway_grok_test.go`: covers API-key forwarding, CLI proxy paths, additional tools, and Composer reasoning cleanup.
- `frontend/src/components/account/AccountUsageCell.vue`: marks Grok quota values as remaining capacity.
- `frontend/src/components/account/CreateAccountModal.vue`: enables Grok API-key account creation with the official endpoint default.
- `frontend/src/components/account/EditAccountModal.vue`: supports editing Grok API-key credentials and endpoint defaults.
- `frontend/src/components/account/UsageProgressBar.vue`: renders remaining-capacity width and severity correctly.
- `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts`: verifies Grok remaining-capacity presentation.
- `frontend/src/components/account/__tests__/CreateAccountModal.grok.spec.ts`: covers Grok API-key account creation.
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`: covers Grok API-key editing while retaining local account capability tests.
- `frontend/src/components/account/__tests__/UsageProgressBar.spec.ts`: covers remaining-capacity colors, width, and labels.
- `frontend/src/components/keys/UseKeyModal.vue`: generates Grok CLI and Responses-capable OpenCode configuration.
- `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts`: verifies Grok client tabs and generated configuration.
- `frontend/src/composables/__tests__/useGrokOAuth.spec.ts`: verifies the OAuth subscription-proxy credential.
- `frontend/src/composables/useGrokOAuth.ts`: stores the CLI proxy for new Grok OAuth accounts.
- `frontend/src/i18n/locales/en/dashboard.ts`: adds English Grok CLI setup guidance.
- `frontend/src/i18n/locales/zh/dashboard.ts`: adds Chinese Grok CLI setup guidance.
- `docs/UPSTREAM_SYNC.md`: records the PR scope, restoration evidence, validation, and rollback points.
- `progress.md`: records this merge and its verification evidence.
- Rollback: first save or commit the current worktree, then run `git revert -m 1 09791722cbf9a0afdeab8c93df1d29bcab17784a`; the pre-merge commit is also available at `backup/pre-pr-4009-merge-20260712-142050`, and the pre-merge worktree snapshot is stash object `46a8d715bb1a9f0cb32720cd79641e1ce33bda61`.

## 2026-07-12 - Task: Fix Gemini native scheduling for OpenAI accounts

### What was done
- Preserved the non-sensitive `openai_capabilities` field in slim scheduler account metadata so load-aware scheduling can recognize OpenAI API-key accounts opted into Gemini native passthrough.
- Added regression coverage for metadata filtering and the load-batch-enabled Gemini native account selection path while retaining the default-off capability boundary.
- Documented that Gemini native passthrough works with load batch scheduling enabled and that account edits and service startup refresh scheduler snapshots.

### Testing
- `go test -tags=unit ./internal/repository -run 'TestBuildSchedulerMetadataAccount_(KeepsOpenAIEndpointCapabilities|KeepsSparkShadowRoutingIdentity)$' -count=1` passed from `backend/`.
- `go test -tags=unit ./internal/service -run 'TestOpenAISelectAccountForGeminiNativeWithLoadAwareness_UsesCapabilityFromSchedulerSnapshot|TestOpenAIGeminiNativeCapabilityDefaultsOff' -count=1` passed from `backend/`.
- `go test -tags=unit ./internal/repository ./internal/service -count=1` passed from `backend/`.
- A temporary image built from the current worktree completed successfully, including the frontend production build and embedded backend build.
- With load batch scheduling left enabled, an isolated OpenAI group/account/API key using the supplied temporary upstream returned HTTP 200 from `/v1beta/models`, exposed `gemini-3-pro-image-preview`, and returned one image candidate with `finishReason: STOP` from `generateContent`.
- The Redis scheduler metadata for the test account contained `openai_capabilities` before the successful request; all temporary database objects, containers, and the temporary image were removed afterward, and ports 8080, 3001, and 18080 were closed.
- `git diff --check` passed for all files changed by this task.

### Notes
- `backend/internal/repository/scheduler_cache.go`: added `openai_capabilities` to the scheduler credential metadata whitelist.
- `backend/internal/repository/scheduler_cache_unit_test.go`: verifies capabilities survive metadata filtering while access tokens remain excluded.
- `backend/internal/service/scheduler_snapshot_hydration_test.go`: verifies Gemini native account selection succeeds through the load-aware scheduler and returns the hydrated account.
- `docs/OPENAI_GEMINI_NATIVE.md`: documented scheduler compatibility, defaults, and snapshot refresh behavior.
- `progress.md`: recorded the fix, validation evidence, cleanup, and rollback instructions.
- Rollback: run `git restore -- backend/internal/repository/scheduler_cache.go backend/internal/repository/scheduler_cache_unit_test.go backend/internal/service/scheduler_snapshot_hydration_test.go`, then remove the `调度说明` section added to `docs/OPENAI_GEMINI_NATIVE.md` and this task record; restart the service to rebuild scheduler snapshots with the reverted behavior.

## 2026-07-12 - Task: Support custom upstreams for Grok APIKey accounts

### What was done
- Added a Grok APIKey URL builder that applies the existing global upstream URL policy while keeping Grok OAuth on the xAI official-host allowlist.
- Routed Grok Responses, Chat Completions, account connection tests, image generation/editing, and video generation/status/content through the account-type-aware URL policy.
- Added live `/v1/models` synchronization for Grok APIKey accounts using Bearer authentication.
- Documented the APIKey/OAuth security boundary and the interaction with `security.url_allowlist`.

### Testing
- Targeted custom-upstream, allowlist, private-host, OAuth restriction, model-sync, account-test, Responses, Chat, and media URL tests passed in `backend/internal/service`.
- `go test ./internal/service -count=1` passed from `backend/` (53.141s).
- `go test -tags=unit ./internal/service -count=1` passed from `backend/` (93.817s).
- A temporary image built successfully from the current worktree.
- Using an isolated local Grok APIKey account with the supplied temporary upstream, model synchronization returned HTTP 200 with 13 models including `grok-4.3` and `grok-4.5`; the account test for `grok-4.5` returned HTTP 200 with `test_complete` and no Base URL error.
- All temporary accounts, groups, containers, and the temporary image were removed; ports 8080 and 3001 were closed.
- `git diff --check` passed for the task files, and the supplied temporary key was absent from the repository diff.

### Notes
- `backend/internal/service/grok_upstream_url.go`: centralizes account-type-aware Grok endpoint URL validation and construction.
- `backend/internal/service/grok_custom_upstream_test.go`: covers custom public hosts, the global allowlist, private-host rejection, OAuth restrictions, and all Grok endpoint shapes.
- `backend/internal/service/account_test_service.go`: uses the APIKey-aware Grok Responses URL for connection tests.
- `backend/internal/service/openai_gateway_grok.go`: routes all Grok Responses request builders through the shared policy.
- `backend/internal/service/openai_gateway_chat_completions_raw.go`: supports custom Grok APIKey Chat Completions URLs.
- `backend/internal/service/openai_gateway_messages.go`: preserves the shared Grok policy for Anthropic-compatible requests.
- `backend/internal/service/openai_ws_http_bridge.go`: preserves the shared Grok policy for WebSocket-to-HTTP bridging.
- `backend/internal/service/grok_media.go`: supports custom Grok APIKey image and video endpoints without weakening OAuth validation.
- `backend/internal/service/upstream_models.go`: adds authenticated Grok APIKey model synchronization.
- `backend/internal/service/upstream_models_test.go`: verifies Grok model request construction and response parsing.
- `backend/internal/service/openai_gateway_grok_test.go`: updates Grok Responses request-builder coverage for the service-aware URL policy.
- `docs/UPSTREAM_SYNC.md`: documents custom Grok APIKey upstream behavior and security settings.
- `progress.md`: records implementation, validation, cleanup, and rollback instructions.
- Rollback: remove `backend/internal/service/grok_upstream_url.go` and `backend/internal/service/grok_custom_upstream_test.go`; reverse this task's hunks in the nine existing backend files listed above; remove the `Grok APIKey 自定义上游兼容` documentation section and this task record; then rebuild the service image. No database rollback is required.

## 2026-07-13 - Task: Sync upstream/main through v0.1.153

### What was done
- Merged 85 upstream commits from `Wei-Shaw/sub2api` (`e316ebf5..7d239d62e`) into local `main` as merge commit `25e6d87ab`.
- Accepted upstream Alpha Search billing, Grok OAuth media routing, Grok video edit/extension, WebSocket lifecycle, scheduler cooldown, account plan type, frontend performance, and deployment changes.
- Restored the pre-existing local worktree and resolved 14 overlapping files so local OpenAI video tasks/content download, Nano Banana, Gemini native scheduling, Grok APIKey custom upstream security policy, media billing, and UI customizations remain available.

### Testing
- `go test -tags=unit ./internal/handler ./internal/server/routes ./internal/service ./internal/repository ./internal/server/middleware -count=1` passed from `backend/`.
- The complete frontend Vitest suite passed: 151 files and 1000 tests.
- `pnpm build` passed and generated the embedded frontend bundle; only existing Browserslist age and large-chunk warnings remained.
- `git diff --check` passed, and `git diff --name-only --diff-filter=U` returned no files.

### Notes
- Upstream merge file set: all files in `git diff --name-only 09791722..25e6d87ab`; this is the authoritative list of the 182-file merge result.
- `backend/internal/handler/endpoint.go`: retained Alpha Search and Grok mutation endpoints alongside local OpenAI video endpoints.
- `backend/internal/server/routes/gateway.go`: registered Grok edit/extension plus local video content and Nano Banana routes without Gin parameter conflicts.
- `backend/internal/service/account.go`: accepted upstream separation of Grok OAuth text and media base URLs.
- `backend/internal/service/grok_upstream_url.go`: extended the local APIKey-aware URL policy to upstream video edit/extension and OAuth media routing.
- `backend/internal/service/grok_media.go`: combined upstream OAuth media transport and CLI identity with local custom APIKey upstreams and binary content download.
- `backend/internal/service/openai_gateway_grok.go`: combined upstream prompt-cache identity with the local service-aware Grok URL builder.
- `backend/internal/service/openai_gateway_usage.go`: retained both Alpha Search per-call billing and local generalized media billing.
- `backend/internal/service/upstream_models.go`: kept Grok model synchronization on the shared APIKey URL security policy.
- `docs/UPSTREAM_SYNC.md`: recorded synchronized features, conflict decisions, validation, and rollback points.
- `progress.md`: appended this task record.
- Rollback: first preserve the current worktree, then run `git revert -m 1 25e6d87ab5bdf3cf7470579a54f2549466191e12`; the pre-sync commit is also at `backup/pre-upstream-sync-20260713-214300`, and the original worktree snapshot remains in stash object `3ce8050f170e0d2b3634f3e8584378aa53f212d2`.

## 2026-07-13 - Task: Remove the homepage brand subtitle

### What was done
- Removed the fixed/default site subtitle from the homepage header brand so the top-left area now shows only the configured logo and site name.
- Confirmed the community chat jump is caused by unconditional client-side scroll-to-bottom calls on new messages, message replacement, and media load events; no chat behavior was changed in this task.

### Testing
- `pnpm exec vue-tsc -b` passed from `frontend/`.
- `pnpm exec eslint src/views/HomeView.vue` passed from `frontend/`.
- `pnpm build` passed from `frontend/`; only the existing Browserslist age and large-chunk warnings remained.

### Notes
- `frontend/src/views/HomeView.vue`: removed the header subtitle markup, computed fallback, and orphaned responsive styles.
- `docs/HOME_THREE_SCENE.md`: documented that the homepage brand renders only the site logo and name.
- `progress.md`: appended this task record.
- Rollback: before staging this task, run `git restore --worktree frontend/src/views/HomeView.vue`; then remove the added behavior bullet from `docs/HOME_THREE_SCENE.md` and this task record.

## 2026-07-13 - Task: Preserve reading position in community chat

### What was done
- Replaced unconditional group-chat scrolling with a bottom-pinned state that follows incoming messages only when the user is already near the latest message.
- Kept automatic following for the current user's own messages and for the initial history load.
- Prevented lazy image and video metadata loads from moving users who are reading older messages.
- Added a floating, counted new-message button that returns to the latest message and clears when the user reaches the bottom.

### Testing
- `pnpm test:run src/views/user/__tests__/communityChatScroll.spec.ts src/api/__tests__/communityChat.spec.ts` passed: 2 files and 8 tests.
- `pnpm exec eslint src/views/user/CommunityChatView.vue src/views/user/communityChatScroll.ts src/views/user/__tests__/communityChatScroll.spec.ts src/i18n/locales/zh/misc.ts src/i18n/locales/en/misc.ts` passed.
- `pnpm exec vue-tsc -b` passed.
- `pnpm build` passed; only the existing Browserslist age and large-chunk warnings remained.

### Notes
- `frontend/src/views/user/CommunityChatView.vue`: tracks bottom proximity, preserves historical reading position, and renders the new-message control.
- `frontend/src/views/user/communityChatScroll.ts`: contains the bottom-threshold and follow-decision helpers.
- `frontend/src/views/user/__tests__/communityChatScroll.spec.ts`: covers threshold, incoming-message, and own-message follow behavior.
- `frontend/src/i18n/locales/zh/misc.ts`: adds the Chinese new-message count label.
- `frontend/src/i18n/locales/en/misc.ts`: adds the English new-message count label.
- `docs/COMMUNITY_CHAT.md`: documents the new scroll-follow behavior.
- `progress.md`: appended this task record.
- Rollback: remove `frontend/src/views/user/communityChatScroll.ts` and `frontend/src/views/user/__tests__/communityChatScroll.spec.ts`, then reverse this task's hunks in `CommunityChatView.vue`, both `misc.ts` locale files, `docs/COMMUNITY_CHAT.md`, and this task record. No backend or database rollback is required.

## 2026-07-13 - Task: Align and emphasize the community chat unread badge

### What was done
- Moved the expanded sidebar unread badge into the “用户群聊” label so it follows the menu text instead of being positioned against the sidebar container.
- Anchored the badge to the community-chat icon while the sidebar is collapsed.
- Increased the badge to 12 px with a stronger red fill, white ring, and subtle glow for clearer visibility.

### Testing
- `pnpm test:run src/components/layout/__tests__/AppSidebar.spec.ts` passed: 1 file and 9 tests.
- `pnpm exec eslint src/components/layout/AppSidebar.vue src/components/layout/__tests__/AppSidebar.spec.ts` passed.
- `pnpm exec vue-tsc -b` passed.
- `pnpm build` passed; only the existing Browserslist age and large-chunk warnings remained.
- `git diff --check -- frontend/src/components/layout/AppSidebar.vue frontend/src/components/layout/__tests__/AppSidebar.spec.ts docs/COMMUNITY_CHAT.md` passed.

### Notes
- `frontend/src/components/layout/AppSidebar.vue`: groups the unread badge with the community-chat label or collapsed icon and increases its visual prominence.
- `frontend/src/components/layout/__tests__/AppSidebar.spec.ts`: verifies badge placement in all sidebar navigation branches and locks the larger badge style.
- `docs/COMMUNITY_CHAT.md`: documents expanded and collapsed unread-badge placement.
- `progress.md`: appended this task record.
- Rollback: reverse this task's hunks in `frontend/src/components/layout/AppSidebar.vue`, `frontend/src/components/layout/__tests__/AppSidebar.spec.ts`, and `docs/COMMUNITY_CHAT.md`, then remove this task record from `progress.md`. No backend or database rollback is required.

## 2026-07-14 - Task: Sync upstream/main through 0.1.155

### What was done
- Merged 68 upstream commits from `Wei-Shaw/sub2api` (`7d239d62e..da85cc7e4`) into local `main` as merge commit `7302e1a9e`.
- Accepted upstream Grok SSO/import probing, quota and monitoring improvements; OpenAI namespace, image, connection, and long-context billing changes; Server-Timing; and scheduler rebuild fixes.
- Restored the complete pre-sync worktree and resolved 14 content conflicts while retaining local OpenAI video, Nano Banana, Gemini native, custom Grok APIKey upstream, media billing, image bridge, and UI behavior.

### Testing
- `go test -tags=unit ./... -count=1` passed from `backend/`, including service, handler, repository, server, packages, Ent, and migrations.
- The complete frontend Vitest suite passed: 159 files and 1065 tests.
- `pnpm exec eslint src/components/admin/usage/UsageTable.vue src/composables/useChannelMonitorFormat.ts src/constants/channelMonitor.ts` passed.
- `pnpm build` passed; only the existing Browserslist age and large-chunk warnings remained.
- `git diff --check` and `git diff --cached --check` passed, and no unresolved files remained.
- `go mod tidy -diff` reported only eight stale `go.sum` checksum lines already shared by the pre-sync tree and upstream; they were intentionally left unchanged as out of scope.

### Notes
- Upstream merge file set: all files in `git diff --name-only 25e6d87ab..7302e1a9e`; this is the authoritative list of the 236-file merge result.
- `backend/internal/service/channel_monitor_checker.go` and related monitor files: use the upstream Grok adapter/default while retaining local retry-status test coverage.
- `backend/internal/service/openai_codex_transform.go`, `openai_gateway_forward.go`, and `openai_ws_forwarder_ingress.go`: combine upstream Responses Lite and duplicate-tool protections with local request-level image bridge controls.
- `backend/internal/handler/openai_codex_models_handler_test.go`: supplies the fork-specific video task binding dependency to the upstream failover test.
- `frontend/src/components/admin/usage/UsageTable.vue`: retains signed video refunds and adds the upstream long-context billing marker.
- `deploy/.env.example`: retains the local image workspace URL beside upstream image keepalive settings.
- `docs/UPSTREAM_SYNC.md`: records synchronized features, conflict decisions, validation, and rollback points.
- `progress.md`: appended this task record.
- Rollback: first preserve the current worktree, then run `git revert -m 1 7302e1a9e3b3074ffc4ff900a23cded9b38ec1d8`; the pre-sync commit is also at `backup/pre-upstream-sync-20260714-221620`, and the original worktree snapshot remains in stash object `c4d2844119552dac00ebd1ee0edda69ca9d516cc`.

## 2026-07-14 - Task: Separate and emphasize community direct-message reminders

### What was done
- Split community group-chat and direct-message unread state so entering the group chat no longer clears private-message reminders.
- Added an independent blinking unread dot beside the private-message button for both administrators and regular users; private reminders clear only after the dialog loads successfully or a visible thread receives the message.
- Kept the sidebar as a combined group/direct reminder, prevented both badge containers from clipping the dot, and added a reduced-motion fallback.

### Testing
- `pnpm test:run src/api/__tests__/communityChat.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts src/views/user/__tests__/communityChatUnread.spec.ts src/views/user/__tests__/communityChatScroll.spec.ts` passed: 4 files and 24 tests.
- `pnpm exec eslint src/api/communityChat.ts src/api/__tests__/communityChat.spec.ts src/components/layout/AppSidebar.vue src/components/layout/__tests__/AppSidebar.spec.ts src/views/user/CommunityChatView.vue src/views/user/__tests__/communityChatUnread.spec.ts` passed.
- `pnpm exec vue-tsc -b` passed.
- `pnpm build` passed; only the existing Browserslist age and large-chunk warnings remained.
- `git diff --check -- frontend/src/api/communityChat.ts frontend/src/api/__tests__/communityChat.spec.ts frontend/src/components/layout/AppSidebar.vue frontend/src/components/layout/__tests__/AppSidebar.spec.ts frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/communityChatUnread.spec.ts docs/COMMUNITY_CHAT.md` passed.

### Notes
- `frontend/src/api/communityChat.ts`: dispatches a same-page event when a group or direct-message cursor is marked seen.
- `frontend/src/api/__tests__/communityChat.spec.ts`: verifies seen-cursor events and their scope details.
- `frontend/src/components/layout/AppSidebar.vue`: separates group/direct unread state, preserves direct unread on group-chat entry, and renders an unclipped blinking summary badge.
- `frontend/src/components/layout/__tests__/AppSidebar.spec.ts`: covers separated state, group-only clearing, unclipped layout, badge size, animation, and reduced-motion behavior.
- `frontend/src/views/user/CommunityChatView.vue`: restores direct unread state, renders the private-message button badge, and clears it only when the private thread is actually viewed.
- `frontend/src/views/user/__tests__/communityChatUnread.spec.ts`: covers private badge placement, unread restoration, socket updates, seen timing, and animation styles.
- `docs/COMMUNITY_CHAT.md`: documents the separate group/direct reminder semantics and blinking badge placement.
- `progress.md`: records implementation, validation, and rollback instructions.
- Rollback: run `git restore --worktree frontend/src/components/layout/AppSidebar.vue frontend/src/components/layout/__tests__/AppSidebar.spec.ts` to restore their pre-task staged versions; remove `frontend/src/views/user/__tests__/communityChatUnread.spec.ts`; then reverse only the `community-chat:seen`, `directUnread`/`latestDirectMessageId`, `community-direct-unread-dot`, and updated reminder-description hunks in the five untracked community-chat source, test, and documentation files listed above. No backend or database rollback is required.

## 2026-07-14 - Task: Build and push the latest multi-architecture image

### What was done
- Built `iotwq/china-api:latest` from the complete current workspace code for `linux/amd64` and `linux/arm64`.
- Pushed the new multi-architecture image index to Docker Hub without starting or changing any local Compose services.

### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 -t iotwq/china-api:latest --push .` passed.
- `docker buildx imagetools inspect iotwq/china-api:latest` returned index digest `sha256:e8dd69ea5c1f722c4df9c3481383228c3e1fb108c45b45e017f2650fd273db0f` with both `linux/amd64` and `linux/arm64` manifests.
- Raw manifest inspection returned AMD64 digest `sha256:12e53ad3d17ee1776e86c1e43692c6f3bb748ca21bde4fb2597230905885f43a` and ARM64 digest `sha256:a771180a00188b1865f270af13c90bda82c8caae0543346c2c1399c68041e384`.
- Port checks confirmed no local listeners on `8080` or `3001` after the push.

### Notes
- `iotwq/china-api:latest`: now points to multi-architecture digest `sha256:e8dd69ea5c1f722c4df9c3481383228c3e1fb108c45b45e017f2650fd273db0f`.
- Build warnings: the existing Browserslist age and Vite chunk-size warnings remained non-fatal.
- `progress.md`: records the image build, push, remote manifest verification, and rollback digest.
- Rollback: run `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:afd1919f6d56aceeafa3ec8ce31a965bf4c19a0d06bdba742c650d1ba50549e8` to repoint `latest` to the previously published multi-architecture image.

## 2026-07-15 - Task: Sync upstream/main through 0.1.156

### What was done
- Merged 145 upstream commits from `Wei-Shaw/sub2api` (`da85cc7e4..eb2b8632d`) into local `main` as merge commit `2b35cbe02`.
- Accepted upstream OpenAI Agent Identity, first-output/failover, Grok OAuth recovery and custom upstream, scheduler lifecycle, content moderation, account duplication, subscription currency and frontend reliability changes.
- Restored the complete pre-sync worktree and resolved 13 content conflicts while retaining local Gemini native OpenAI accounts, Grok APIKey model sync and video download, OpenAI video/Nano Banana, media billing, community chat and frontend customizations.
- Added the direct internationalization compiler development dependency required by the new upstream locale compilation test under pnpm strict dependency resolution.

### Testing
- `go test -tags=unit ./... -count=1` passed from `backend/` after updating fork-specific constructor fixtures and Ops argument-count coverage.
- `pnpm test:run` passed: 163 files and 1160 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` passed; only the existing localStorage, Vue test-environment, Browserslist age, and Vite large-chunk warnings remained.
- `pnpm install --frozen-lockfile` passed and linked `@intlify/message-compiler@9.14.5` as a direct dev dependency.
- Targeted Grok/Gemini/image bridge/streaming backend tests and account modal/page frontend tests passed during conflict resolution.
- `git diff --check` and `git diff --cached --check` passed, with no unresolved files or conflict markers.

### Notes
- Upstream merge file set: every file returned by `git diff --name-only 7302e1a9e3b3074ffc4ff900a23cded9b38ec1d8..2b35cbe028c3c805625be44d4e51b9efb59f54fb`; this is the authoritative 301-file merge list.
- `backend/internal/service/account_test_service.go`: uses the upstream Grok Responses URL builder while retaining local quota recovery handling.
- `backend/internal/service/grok_upstream_url.go`: uses the upstream account-aware URL security policy and adds local model-list and video-content URLs.
- `backend/internal/service/grok_media.go`: uses upstream media URL/usage handling while retaining MP4 streaming, Range forwarding, and OAuth-only CLI headers.
- `backend/internal/service/upstream_models.go`: routes Grok APIKey model synchronization through the new upstream-compatible URL builder.
- `backend/internal/service/openai_codex_transform.go`: combines exact Codex image function protection with local Agent/request-level image bridge controls.
- `backend/internal/service/openai_gateway_chat_completions_raw.go`: adopts the upstream Grok Chat Completions URL path.
- `backend/internal/service/openai_gateway_grok.go`: adopts the upstream config-aware Grok Responses request builder.
- `backend/internal/service/openai_gateway_grok_chat_bridge.go`: adopts the upstream Grok Responses bridge call contract.
- `backend/internal/service/openai_gateway_messages.go`: adopts the upstream Grok Messages-to-Responses request call contract.
- `backend/internal/service/openai_gateway_response_handling.go`: keeps upstream first-output, event flush, cancellation, and missing-terminal semantics.
- `backend/internal/service/openai_ws_http_bridge.go`: adopts the upstream Grok request builder in the WebSocket-to-HTTP bridge.
- `backend/internal/service/grok_custom_upstream_test.go`: verifies custom Grok models, Responses, Chat, media, OAuth policy, and video-content URLs through the new builders.
- `backend/internal/service/openai_gateway_grok_test.go`: updates Grok request-builder tests to the upstream global helper.
- `backend/internal/service/openai_codex_transform_test.go`: aligns merged image-function cases with the local custom-function injection guard.
- `backend/internal/service/openai_gateway_service_test.go`: aligns the old client-disconnect case with upstream failover-before-output behavior.
- `backend/internal/handler/openai_gateway_credential_failover_loop_test.go`: supplies the fork-specific video task binding dependency to an upstream constructor fixture.
- `backend/internal/handler/openai_gateway_handler_test.go`: supplies the fork-specific video task binding dependency to an upstream failover fixture.
- `backend/internal/handler/openai_responses_failover_cancel_test.go`: supplies the fork-specific video task binding dependency to an upstream cancellation fixture.
- `backend/internal/repository/ops_repo_args_test.go`: updates the expected Ops insert argument count for three new upstream fields.
- `frontend/src/components/account/CreateAccountModal.vue`: combines upstream Grok OAuth upstream settings with local extra configuration.
- `frontend/src/views/admin/AccountsView.vue`: retains conditional modal mounting and wires the upstream duplicate-account event.
- `frontend/package.json`: declares the internationalization message compiler used directly by the upstream test.
- `frontend/pnpm-lock.yaml`: locks the direct compiler development dependency without changing its resolved version.
- `docs/UPSTREAM_SYNC.md`: records the synchronized behavior, conflict decisions, validation, and rollback points, and corrects the Grok OAuth URL-policy note.
- `progress.md`: records this task, testing evidence, changed-file inventory, and rollback instructions.
- Rollback: first preserve the current worktree, then run `git revert -m 1 2b35cbe028c3c805625be44d4e51b9efb59f54fb`; the pre-sync commit is also at `backup/pre-upstream-sync-20260715-221149`, and the original worktree snapshot remains in stash object `26d13598a2f9c355cef5417fb5afa76ff2bcdaaf`.

## 2026-07-16 - Task: Build and push the latest multi-architecture image

### What was done
- Built `iotwq/china-api:latest` from the complete current workspace for `linux/amd64` and `linux/arm64`.
- Pushed the new OCI multi-architecture image index to Docker Hub without starting or changing local Compose services.

### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 -t iotwq/china-api:latest --push .` passed, including frozen frontend dependency installation, Vue type checking/build, Go release compilation, runtime image assembly, and registry upload.
- `docker buildx imagetools inspect iotwq/china-api:latest` returned index digest `sha256:cd2e952c227654e46dfef5181b08f5c9d66113be9c7c8f7727ad188327aa6362` with both target platforms.
- Raw manifest inspection returned AMD64 digest `sha256:56a25fc2c7893a263a26f09adcae9fabe8d9c2fa83f90eef83dc0b35031e3159` and ARM64 digest `sha256:527edafc37dce0bd0752c86a6a07863678cb6c4994f746d9500d9e9518ae6b3d`.
- Port checks confirmed no local listeners on `8080` or `3001` after the push.

### Notes
- `iotwq/china-api:latest`: now points to multi-architecture digest `sha256:cd2e952c227654e46dfef5181b08f5c9d66113be9c7c8f7727ad188327aa6362`.
- `progress.md`: records the build, push, remote manifest verification, and rollback point for this release.
- Build warnings: the existing Browserslist age and Vite chunk-size warnings remained non-fatal.
- Rollback: run `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:e8dd69ea5c1f722c4df9c3481383228c3e1fb108c45b45e017f2650fd273db0f` to repoint `latest` to the previously published multi-architecture image.

## 2026-07-16 - Task: Sync upstream/main through 0.1.158

### What was done
- Merged 88 upstream commits from `Wei-Shaw/sub2api` (`eb2b8632d..bc2244c83`) into local `main` as merge commit `4af4d7d83`.
- Accepted upstream Grok endpoint/WebSocket v2, async image tasks and object storage, image input-token pricing, upstream billing-rate probing, cost-aware OpenAI scheduling, audit logging, session binding, step-up 2FA, batch limits and duplication workflows.
- Restored the complete pre-sync worktree and resolved 15 content conflicts while retaining local Gemini-native OpenAI accounts, Grok video downloads, OpenAI video/Nano Banana, community chat and frontend customizations.
- Adapted the local media handlers and test fixtures to the upstream model-aware scheduling feedback and authentication constructor signatures.

### Testing
- `GOPROXY=https://goproxy.cn,direct go generate ./cmd/server` passed after the default Go proxy returned one transient `EOF`.
- Targeted backend handler, route, service and server tests passed after the merge adaptations.
- `go test -tags=unit ./... -count=1` passed from `backend/`, including the service suite and OpenAI WebSocket v2 package.
- `pnpm install --frozen-lockfile` passed from `frontend/` with no lockfile changes.
- `pnpm test:run` passed: 174 files and 1226 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` passed; only the existing localStorage, Vue test-environment, Browserslist age, and Vite large-chunk warnings remained.
- `git diff --check`, `git diff --cached --check`, and the upstream ancestry check passed with no unresolved conflict markers.

### Notes
- Upstream merge file set: every file returned by `git diff --name-status 2b35cbe028c3c805625be44d4e51b9efb59f54fb..4af4d7d83e4982eb6d9c1182ad029a6a9e2ba831`; this is the authoritative 350-file upstream inventory.
- `.gitignore`: retains the upstream async-image documentation exception and all local documentation exceptions.
- `backend/cmd/server/wire_gen.go`: regenerates dependency injection with upstream audit/step-up and async-image services plus local community chat and video task bindings.
- `backend/internal/handler/endpoint.go`: keeps both upstream image-task endpoint normalization and local video-generation normalization.
- `backend/internal/handler/endpoint_test.go`: covers both merged endpoint families.
- `backend/internal/handler/gateway_key_billing_test.go`: supplies the fork-specific video task binding dependency to the merged OpenAI gateway constructor.
- `backend/internal/handler/community_chat_handler_test.go`: supplies the new request context to token generation.
- `backend/internal/handler/openai_nano_banana.go`: reports scheduling results with the normalized mapped upstream model.
- `backend/internal/handler/openai_videos.go`: reports scheduling results with the account-mapped video model.
- `backend/internal/server/router.go`: keeps upstream audit and step-up middleware while registering local community chat routes.
- `backend/internal/server/routes/gateway.go`: keeps upstream async image routes together with local OpenAI/Grok video and Nano Banana routes.
- `backend/internal/server/routes/gateway_key_billing_test.go`: supplies the fork-specific video task binding dependency to the route test fixture.
- `backend/internal/service/account.go`: combines upstream Alpha Search and Responses capabilities with local Gemini-native capability.
- `backend/internal/service/gateway_anthropic_apikey_passthrough_test.go`: keeps the local real-Claude-Code passthrough regression beside upstream Haiku behavior.
- `backend/internal/service/gateway_context_management_test.go`: adopts upstream Haiku context-management preservation expectations.
- `backend/internal/service/grok_media.go`: keeps Range-capable video streaming and limits CLI headers to OAuth CLI proxy targets.
- `backend/internal/service/openai_gateway_scheduling.go`: combines upstream cost-aware selection with local Gemini-native selectors.
- `backend/internal/service/openai_gateway_service.go`: keeps upstream WebSocket terminal events and model transient state plus local response bodies and video task bindings.
- `frontend/src/i18n/locales/en/admin/channels.ts`: combines upstream monitor duplication copy with the local timeout message.
- `frontend/src/i18n/locales/zh/admin/channels.ts`: combines upstream monitor duplication copy with the local timeout message.
- `frontend/src/views/admin/AccountsView.vue`: combines upstream async account modals and step-up verification with local conditional modal mounting.
- `docs/UPSTREAM_SYNC.md`: records the 0.1.158 synchronized behavior, conflict decisions, validation, and rollback points.
- `progress.md`: appends this task record without rewriting prior history.
- Rollback: first preserve the current worktree, then run `git revert -m 1 4af4d7d83e4982eb6d9c1182ad029a6a9e2ba831`; the pre-sync commit is also at `backup/pre-upstream-sync-20260716-230929`, and the original worktree snapshot remains in stash object `6679e88607a517f677bc81a5a3c5341069d12371`.

## 2026-07-16 - Task: Build and push the 0.1.158 multi-architecture image

### What was done
- Built `iotwq/china-api:latest` from the complete current workspace after the upstream `0.1.158` synchronization.
- Published a new OCI multi-architecture index for `linux/amd64` and `linux/arm64` to Docker Hub without starting or changing local Compose services.
- Reused the completed build cache and restarted only the dedicated BuildKit builder after transient Docker Hub TLS timeouts interrupted the first two push attempts.

### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 -t iotwq/china-api:latest --push .` completed successfully, including frozen frontend dependency installation, Vue production builds, Go release compilation, runtime image assembly, and registry upload.
- `docker buildx imagetools inspect iotwq/china-api:latest` returned index digest `sha256:3a448d4c709a48b4d12c535ae6c905ff113965d6b50426482365196ef3e48e87`.
- Raw remote manifest inspection returned AMD64 digest `sha256:9fd534acca5b64fdef5949492515a1dfdd1c522b43020c8973edf5168b371cdc` and ARM64 digest `sha256:a57a56378b6f84148188d0b0cf0dd33bd96bbdcf40815a1df39429dc3d93bcc9`.
- Port checks confirmed no local listeners on `8080` or `3001` after the push.
- `git diff --check` and `git diff --cached --check` passed after recording the release.

### Notes
- `iotwq/china-api:latest`: now points to multi-architecture digest `sha256:3a448d4c709a48b4d12c535ae6c905ff113965d6b50426482365196ef3e48e87`.
- `progress.md`: appends this build, push, remote verification, network retry, and rollback record.
- Build warnings: the existing Node URL deprecation, Browserslist age, and Vite chunk-size warnings remained non-fatal.
- Rollback: run `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:cd2e952c227654e46dfef5181b08f5c9d66113be9c7c8f7727ad188327aa6362` to repoint `latest` to the previously published multi-architecture image.

## 2026-07-17 - Task: Sync upstream/main through 0.1.160

### What was done
- Merged 36 upstream commits from `Wei-Shaw/sub2api` (`bc2244c83..57914967c`) into local `main` as merge commit `7d3ce6bb9`.
- Accepted upstream prompt security auditing, Grok media eligibility quarantine, explicit image-generation intent, trusted-client-IP hardening, S3 step-up verification and Stripe lazy-loading changes.
- Restored the complete pre-sync worktree and resolved 3 content conflicts while retaining local Gemini-native OpenAI accounts, Grok custom upstream/video downloads, OpenAI video/Nano Banana, community chat and frontend customizations.
- Integrated local Nano Banana and OpenAI video routes with the new prompt-audit coordinator and route coverage guard.
- Added the missing Wire binding required to regenerate the upstream prompt-audit dependency graph.

### Testing
- `GOPROXY=https://goproxy.cn,direct go generate ./cmd/server` passed after adding the missing `PromptAdminService` binding.
- Targeted handler, repository, security-audit, route, server and service tests passed after local media audit integration.
- `go test -tags=unit ./... -count=1` passed from `backend/`, including security audit, service, WebSocket v2 and migration packages.
- `pnpm install --frozen-lockfile` passed from `frontend/` with no dependency drift.
- `pnpm test:run` passed: 180 files and 1259 tests after adapting the Stripe source-contract test to the normalized module ID.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` passed; only the existing localStorage, Vue test-environment, Browserslist age, and Vite large-chunk warnings remained.
- `git diff --check`, `git diff --cached --check`, and the upstream ancestry check passed with no unresolved conflict markers.

### Notes
- Upstream merge file set: every file returned by `git diff --name-status 4af4d7d83e4982eb6d9c1182ad029a6a9e2ba831..7d3ce6bb95e39b16f21fd6ce2ebb65370463fd9e`; this is the authoritative 159-file upstream inventory.
- `backend/cmd/server/wire_gen.go`: regenerated the combined prompt-audit, community-chat, async-image and OpenAI-video dependency graph.
- `backend/internal/handler/wire.go`: keeps the local community-chat provider and adopts upstream security-audit-aware gateway providers.
- `backend/internal/handler/openai_nano_banana.go`: routes Nano Banana prompts through the new security-audit coordinator.
- `backend/internal/handler/openai_videos.go`: routes OpenAI video creation prompts through the new security-audit coordinator.
- `backend/internal/repository/scheduler_cache_unit_test.go`: covers both upstream Grok media eligibility and local OpenAI endpoint capabilities in scheduler metadata.
- `backend/internal/securityaudit/prompt_module.go`: binds the prompt admin interface to the concrete prompt service so Wire regeneration succeeds.
- `backend/internal/server/routes/prompt_audit_route_coverage_test.go`: classifies local video and Nano Banana POST routes as audited endpoints.
- `frontend/vite.config.ts`: retains normalized detailed vendor splitting while placing Stripe in an independent lazy chunk.
- `frontend/src/views/user/__tests__/stripeLazyLoading.spec.ts`: accepts normalized module IDs while preserving the Stripe-before-misc contract.
- `docs/UPSTREAM_SYNC.md`: records the 0.1.160 synchronized behavior, compatibility decisions, validation and rollback points.
- `progress.md`: appends this task record without rewriting prior history.
- Rollback: first preserve the current worktree, then run `git revert -m 1 7d3ce6bb95e39b16f21fd6ce2ebb65370463fd9e`; the pre-sync commit is also at `backup/pre-upstream-sync-20260717-210152`, and the original worktree snapshot remains in stash object `89626cecd82c4ac14036981ba6fa18fa48f54ab3`.

## 2026-07-17 - Task: Build and push the 0.1.160 multi-architecture image

### What was done
- Built `iotwq/china-api:latest` from the complete current workspace at application version `0.1.160` and Git HEAD `7d3ce6bb95e39b16f21fd6ce2ebb65370463fd9e`.
- Published a new OCI multi-architecture image index for `linux/amd64` and `linux/arm64` to Docker Hub without starting or changing local Compose services.

### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 -t iotwq/china-api:latest --push .` passed, including frozen frontend dependency installation, Vue production builds, Go release compilation, runtime image assembly, and registry upload.
- `docker buildx imagetools inspect iotwq/china-api:latest` returned index digest `sha256:141ecb9c9ad73c37d4b8446ce2f33366a1de014838a256793676bdb9f29df6af`.
- Raw remote manifest inspection returned AMD64 digest `sha256:6d46147695d8aa48722772d6df5b86f2b632b2cb94146c926b46314a6d2ec1d2` and ARM64 digest `sha256:1eddec7d732b6811df56993bf22dfdab2471ffcff1f163d3735c8c779e0e2084`.
- Port checks confirmed no local listeners on `8080` or `3001` after the push.
- The frontend build completed with only the existing Browserslist age and Vite large-chunk warnings.

### Notes
- `iotwq/china-api:latest`: now points to multi-architecture digest `sha256:141ecb9c9ad73c37d4b8446ce2f33366a1de014838a256793676bdb9f29df6af`.
- `progress.md`: appends this build, push, remote verification, warning, and rollback record.
- Rollback: run `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:3a448d4c709a48b4d12c535ae6c905ff113965d6b50426482365196ef3e48e87` to repoint `latest` to the previously published multi-architecture image.

## 2026-07-17 - Task: Align channel monitor retries with pool-mode upstream behavior

### What was done
- Changed channel-monitor probes from a limit of three distinct local accounts to a shared budget of three real upstream attempts.
- Allowed pool-mode same-account retries during monitor probes while making both same-account retries and account switches consume the same three-attempt budget.
- Preserved ordinary request behavior when the internal monitor header is absent.
- Added strict normalization for outer HTTP 400 responses whose structured error message starts with `API returned <status>:` for explicitly configured retry statuses `401`, `403`, `429`, `501`, `502`, and `503`.
- Kept generic HTTP 400 responses, malformed wrappers, embedded markers, wrong-case prefixes, and unlisted statuses non-retryable.

### Testing
- Targeted service tests passed for the six wrapped statuses, strict negative cases, single external monitor requests, and protected internal probe headers.
- Targeted handler tests passed for monitor policy resolution, unchanged ordinary pool failover, and a pool retry count of 9 being capped at exactly three monitor attempts.
- `go test -tags=unit ./... -count=1` passed from `backend/`, including handler, service, OpenAI WebSocket v2, server, repository, migration, and route packages.
- `git diff --check` passed before the final record was appended.

### Notes
- `.gitignore`: allows the channel-monitor behavior document to be tracked.
- `backend/internal/handler/channel_monitor_probe.go`: interprets the internal probe header as a total attempt budget and exposes the shared remaining-budget check.
- `backend/internal/handler/channel_monitor_probe_test.go`: verifies ordinary requests stay uncapped and monitor probes stop after three attempts.
- `backend/internal/handler/openai_chat_completions.go`: applies the shared monitor budget to Chat Completions same-account retries and account switches.
- `backend/internal/handler/openai_gateway_handler.go`: applies the same budget to Responses and compatible Messages failover loops.
- `backend/internal/handler/openai_gateway_handler_test.go`: keeps the ordinary pool failover regression and adds an end-to-end three-attempt monitor cap test.
- `backend/internal/service/channel_monitor_const.go`: renames and documents the monitor limit as total upstream attempts.
- `backend/internal/service/channel_monitor_checker.go`: sends the updated three-attempt monitor header.
- `backend/internal/service/channel_monitor_checker_body_test.go`: verifies the updated internal header contract.
- `backend/internal/service/openai_gateway_upstream_errors.go`: strictly extracts the six allowed nested retry statuses from structured outer-400 errors.
- `backend/internal/service/openai_gateway_cc_pipeline.go`: uses normalized retry statuses for OpenAI-compatible Chat Completions failover and pool retries.
- `backend/internal/service/openai_gateway_forward.go`: uses normalized retry statuses for OpenAI Responses failover and pool retries.
- `backend/internal/service/openai_gateway_chat_completions_raw.go`: applies normalized retry statuses to raw Grok Chat Completions forwarding.
- `backend/internal/service/openai_gateway_grok_chat_bridge.go`: applies normalized retry statuses to the Grok Chat-to-Responses bridge.
- `backend/internal/service/openai_gateway_grok.go`: applies normalized retry statuses to Grok Responses forwarding.
- `backend/internal/service/openai_gateway_service_codex_cli_only_test.go`: covers all six allowed wrappers and strict non-retryable 400 variants.
- `docs/CHANNEL_MONITOR.md`: documents the attempt budget, explicit retry-code configuration, strict wrapper format, and external-endpoint limitation.
- `progress.md`: appends this implementation, verification, file inventory, and rollback record.
- Rollback: revert the commit that records the scoped files above; before commit, the pre-task staged index is the rollback point for tracked files, and the two pre-existing untracked `channel_monitor_probe*.go` files must be restored to their prior distinct-account policy rather than deleted.

## 2026-07-17 - Task: Build and push the channel-monitor retry multi-architecture image

### What was done
- Built `iotwq/china-api:latest` from the complete current workspace, including the channel-monitor pool retry changes, at application version `0.1.160` and Git HEAD `7d3ce6bb95e39b16f21fd6ce2ebb65370463fd9e`.
- Published a new OCI multi-architecture image index for `linux/amd64` and `linux/arm64` to Docker Hub without starting or changing local Compose services.

### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 -t iotwq/china-api:latest --push .` passed, including frozen frontend dependency installation, Vue production build, Go release compilation, runtime image assembly, and registry upload.
- `docker buildx imagetools inspect iotwq/china-api:latest` returned index digest `sha256:191081ae0d9c124a4af27af09e139c35bcd26f6f9f69d1eec5c01af32e26cb64`.
- Raw remote manifest inspection returned AMD64 digest `sha256:87d6f14148978f1744644428a52ccfd9482eb08b23d70ae4ddc3c3371172a361` and ARM64 digest `sha256:8420a183e044f963fead7b2742461ba5990e5589b719b8f2451a71b8e9f00526`.
- Port checks confirmed no local listeners on `8080` or `3001` after the push.

### Notes
- `iotwq/china-api:latest`: now points to multi-architecture digest `sha256:191081ae0d9c124a4af27af09e139c35bcd26f6f9f69d1eec5c01af32e26cb64`.
- `progress.md`: appends this build, push, remote verification, warning, and rollback record.
- Build warnings: one npm registry `ECONNRESET` recovered automatically; the existing Node URL deprecation, Browserslist age, and Vite large-chunk warnings remained non-fatal.
- Rollback: run `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:141ecb9c9ad73c37d4b8446ce2f33366a1de014838a256793676bdb9f29df6af` to repoint `latest` to the previously published multi-architecture image.
## 2026-07-18 - Task: Synchronize upstream Wei-Shaw/sub2api 0.1.161

### What was done
- Merged upstream `57914967c..d4b9797ff` (62 commits) into the local branch as merge commit `3908c9365`.
- Restored the local Gemini/Grok media, OpenAI video/Nano Banana, community chat, channel-monitor, and frontend customizations while incorporating upstream authentication, audit, scheduling, media, WebSocket, billing-probe, and deployment changes.
- Resolved the 10 content conflicts and adapted local tests to the upstream video-task binding, gateway constructor, Ops argument, and model-availability repository contracts.

### Testing
- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm install --frozen-lockfile` from `frontend/`: passed.
- `pnpm test:run`: passed, 182 test files and 1278 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build`: passed.
- `git diff --check` and `git diff --cached --check`: passed with no unresolved conflict markers. Existing frontend localStorage/Vue environment, Browserslist age, and Vite chunk-size warnings remain non-fatal.

### Notes
- `backend/internal/handler/grok_media.go`: keeps upstream video lookup IDs, protected content access, and owner binding while retaining local media behavior.
- `backend/internal/server/routes/gateway.go`: keeps one upstream video content route and the local Nano Banana route.
- `backend/internal/service/channel_monitor_checker.go`: combines Anthropic block extraction with local OpenAI/Responses/Grok text extraction.
- `backend/internal/service/openai_gateway_cc_pipeline.go`: preserves normalized retry status and upstream account-disable semantics.
- `backend/internal/service/openai_gateway_forward.go`: applies the same failover behavior to Responses/CC forwarding.
- `backend/internal/service/openai_gateway_grok_test.go`: retains local Grok model/media regressions and adapts video content coverage to the status-then-content flow.
- `backend/internal/repository/ops_repo_args_test.go`: updates the merged Ops argument count to 41.
- `backend/internal/service/openai_gateway_model_availability_test.go`: implements the upstream model-candidate repository method in the local test stub.
- `frontend/src/App.vue`: uses the shared branding favicon helper and retains locale-change title updates.
- `frontend/src/views/public/LegalDocumentView.vue`: keeps upstream settings loading skeleton with local logo fallback.
- `docs/UPSTREAM_SYNC.md`: records the merge range, conflict decisions, validation, and rollback points.
- Rollback: preserve the current worktree, then run `git revert -m 1 3908c9365`; the pre-sync branch is `backup/pre-upstream-sync-20260718-232215`, and the original worktree snapshot is stash `c964f9ecaaca50e569237941f0dc9d37be65f1b9`.

## 2026-07-19 - Task: Add OpenAI-compatible asset upload and audio forwarding

### What was done
- Added authenticated `POST /pg/assets` temporary media upload and public `GET /pg/assets/{asset_id}/{filename}` serving for image, video, and audio references.
- Added OpenAI APIKey forwarding for `/v1/audio/speech`, `/v1/audio/transcriptions`, and `/v1/audio/translations`, including JSON and multipart request preservation, model mapping, upstream validation, proxies, header overrides, failover, and usage extraction.
- Added audio endpoint routing/normalization, APIKey-only audio scheduling capability, user-facing API examples, and compatibility documentation.

### Testing
- `go test ./cmd/server ./internal/...` from `backend/`: passed.
- Focused asset/audio/route tests: passed, including binary speech passthrough, multipart transcription preservation and usage extraction, asset URL serving, and prompt-audit route classification.
- `pnpm typecheck` from `frontend/`: passed.
- `pnpm build` from `frontend/`: passed. Existing Browserslist age and Vite large-chunk warnings remain non-fatal.
- `git diff --check` and `git diff --cached --check`: passed.

### Notes
- `backend/internal/handler/public_asset_handler.go`: stores bounded temporary assets under `pricing.data_dir/pg/assets`, sanitizes paths, and serves configured/request-derived public URLs.
- `backend/internal/handler/public_asset_handler_test.go`: covers authenticated upload, response metadata, and public file serving.
- `backend/internal/handler/openai_audio.go`: validates audio requests, selects APIKey accounts, forwards JSON/multipart bodies, and records usage.
- `backend/internal/service/openai_audio.go`: implements OpenAI audio endpoint URL construction, authentication, upstream response streaming, and failover handling.
- `backend/internal/service/openai_audio_test.go`: covers speech binary passthrough and multipart transcription forwarding.
- `backend/internal/service/account.go`: adds the APIKey-only audio scheduler capability without requiring a new optional account flag.
- `backend/internal/handler/handler.go`, `backend/internal/handler/wire.go`, `backend/cmd/server/wire_gen.go`: register and inject the public asset handler.
- `backend/internal/server/routes/gateway.go`: registers asset and audio routes for `/v1` and the existing bare OpenAI aliases.
- `backend/internal/handler/endpoint.go`: classifies audio endpoints for operations and usage attribution.
- `backend/internal/server/routes/prompt_audit_route_coverage_test.go`: classifies new audio and asset routes for prompt-audit coverage.
- `frontend/src/views/user/ApiDocsView.vue`, `docs/API_DOCS.md`, `docs/OPENAI_MEDIA_COMPAT.md`: document asset upload, audio calls, limits, and temporary URL behavior.
- Rollback: this task is not committed; preserve unrelated dirty changes, then reverse only the listed route/capability/DI additions and remove the three new backend implementation/test files plus `docs/OPENAI_MEDIA_COMPAT.md`. Do not use `git reset --hard` or restore the whole worktree.

## 2026-07-19 - Task: Update API docs regression expectations for media examples

### What was done
- Updated the user-facing API docs test to include the new temporary asset upload and three audio examples in the OpenAI media section.

### Testing
- `pnpm vitest run src/views/user/__tests__/ApiDocsView.spec.ts` from `frontend/`: passed.

### Notes
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: expects nine OpenAI media examples and checks the new asset/audio documentation.
- Rollback: remove only the added assertions from this test; keep the implementation and documentation changes from the preceding task record.

## 2026-07-19 - Task: Harden multipart audio validation

### What was done
- Required both `model` and `file` fields for transcription and translation requests while preserving the original multipart payload for forwarding.

### Testing
- Focused backend asset/audio/route tests passed after the validation change.

### Notes
- `backend/internal/handler/openai_audio.go`: validates multipart field presence and keeps model rewrite behavior.
- Rollback: remove the `hasFile`/multipart field validation branch while retaining the endpoint implementation if an upstream intentionally accepts model-only requests.

## 2026-07-19 - Task: Track the media compatibility document

### What was done
- Added an exact `.gitignore` exception so `docs/OPENAI_MEDIA_COMPAT.md` is included in the repository deliverable.

### Testing
- `git check-ignore -v docs/OPENAI_MEDIA_COMPAT.md` no longer reports the file as ignored.

### Notes
- `.gitignore`: whitelists only the new media compatibility document under the existing `docs/*` rule.
- Rollback: remove the single `!docs/OPENAI_MEDIA_COMPAT.md` exception to restore the previous ignore behavior.

## 2026-07-19 - Task: Build and push the latest multi-architecture image

### What was done
- Built `iotwq/china-api:latest` from the complete current workspace at application version `0.1.161` and commit `3908c936599b`.
- Published the OCI image index for `linux/amd64` and `linux/arm64` to Docker Hub.

### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --build-arg VERSION=0.1.161 --build-arg COMMIT=3908c936599b -t iotwq/china-api:latest --push .`: passed, including frontend production build, Go cross-compilation, runtime image assembly, and registry upload.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with index digest `sha256:3a67ce2ebc7a54d7eb36f339d416a48f6783536eb89e23ffc9261ef7552e50e3`, amd64 digest `sha256:34ed4c11f9f230522f42b65181d36667e8a9a51eac74aef1796b0efbfcb4f6ab`, and arm64 digest `sha256:bd3522332f186450a20d08ab9b32a24181fa2f1cc4cc80d0988a00895647fae4`.
- `git diff --check` and `git diff --cached --check`: passed before the build.
- Existing frontend Browserslist age and Vite large-chunk warnings remained non-fatal. A secondary `docker manifest inspect` read hit a Docker Hub CloudFront EOF after the successful index verification.

### Notes
- `progress.md`: records the build inputs, remote platform digests, warnings, and rollback point.
- Rollback: run `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:191081ae0d9c124a4af27af09e139c35bcd26f6f9f69d1eec5c01af32e26cb64` to repoint `latest` to the previous verified multi-architecture image.

## 2026-07-19 - Task: Fix public asset routing and republish the multi-architecture image

### What was done
- Fixed the embedded frontend middleware so `POST /pg/assets` and public asset reads reach the backend instead of being served the SPA HTML fallback.
- Kept public asset route registration available when a handler was not supplied by generated dependency-injection code, and added route and embedded-frontend regressions.
- Rebuilt and published `iotwq/china-api:latest` as a Docker manifest list for both `linux/amd64` and `linux/arm64`.

### Testing
- `go test ./internal/handler ./internal/server/routes` from `backend/`: passed.
- `go test -tags embed ./internal/web` from `backend/`: passed.
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with index digest `sha256:7f5c6721214183ba101f850940302b6925e81b4ffd0a5c002a66a40adc5888a4`, amd64 digest `sha256:79e4baa688dd65f25fad5bf2316f934976a663fcda866390fcee0974645f28a1`, and arm64 digest `sha256:46ef5746bf3eb9333460d3f98db6544c61137b27ec05a7b3c64ce2c38323415b`.
- Started the published ARM64 manifest in an isolated PostgreSQL/Redis environment: `/health` returned `200`, `POST /pg/assets` with an invalid key returned JSON `401`, and a missing public asset returned an empty `404` instead of frontend HTML.
- `git diff --check` passed for the affected route, embedded-frontend, documentation, and progress files.

### Notes
- `backend/internal/server/routes/gateway.go`: registers the public asset endpoints with a handler fallback.
- `backend/internal/server/routes/gateway_test.go`: verifies both public asset routes are registered.
- `backend/internal/web/embed_on.go`: bypasses embedded SPA handling for `/pg/assets` and its child paths.
- `backend/internal/web/embed_test.go`: verifies public asset paths continue to the backend router.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents the root-path reverse-proxy requirement for public assets.
- `progress.md`: records the routing fix, multi-architecture publication, and isolated runtime evidence.
- Rollback: reverse only the public asset route fallback and embedded-frontend bypass changes listed above; to repoint Docker Hub immediately, run `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:0c64752154aff4085f33e7d7de4ed0c9f2aedade8c42ad8d11356ce9a3598038`.

## 2026-07-19 - Task: Fix temporary asset public URLs for video references

### What was done
- Fixed temporary asset uploads returning URLs such as `https://api.iotwq.top/password/pg/assets/...`, where the configured management frontend path caused video upstreams to receive HTML instead of the uploaded image.
- Public asset URLs now use the upload request's public protocol and host, including trusted reverse-proxy headers, and always place the asset under the root `/pg/assets/...` path.
- Kept asset authentication, storage, TTL, file validation, and the OpenAI video request forwarding contract unchanged.

### Testing
- `go test ./internal/handler -run TestPublicAssetUploadAndServe -count=1` from `backend/`: passed; a configured frontend URL ending in `/password` no longer affects the returned asset URL, and the served file reports `Content-Type: image/png`.
- `go test ./internal/handler ./internal/server/routes ./internal/web` from `backend/`: passed.
- `go test -tags embed ./internal/web -count=1` from `backend/`: passed.
- `gofmt -d internal/handler/public_asset_handler.go internal/handler/public_asset_handler_test.go`: no output.
- Live non-billable reproduction against the currently deployed image: `POST https://api.iotwq.top/pg/assets` returned a URL under `/password/pg/assets/...`; fetching that URL returned HTTP 200 with `Content-Type: text/html` and the management SPA, matching the video task's `invalid image_url` failure.

### Notes
- `backend/internal/handler/public_asset_handler.go`: derives temporary asset public URLs from the API request origin instead of `server.frontend_url`.
- `backend/internal/handler/public_asset_handler_test.go`: covers a management frontend subpath, reverse-proxy origin, and image response content type.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents the API-origin rule for temporary asset URLs.
- `progress.md`: records the root cause, fix, and verification.
- The fix is not pushed to Docker Hub and is not deployed to `api.iotwq.top` in this task; the live service will continue returning the bad `/password/pg/assets/...` URL until its image is rebuilt and deployed.
- Rollback: restore the `PublicAssetHandler` configured frontend fields and the previous `publicBaseURL` precedence, revert the matching test and documentation paragraph, and remove this progress entry. Do not reset the dirty worktree.

## 2026-07-19 - Task: Count channel monitor retries by distinct account

### What was done
- Increased an OpenAI-compatible or Grok model check from three real upstream attempts to at most five distinct local accounts.
- Kept all configured pool-mode retries within one account probe slot, so exhausting one upstream pool still leaves account budget for other eligible accounts.
- Kept first-success behavior: a successful account response makes the channel check successful; all-failure checks stop after five distinct accounts or when no eligible account remains.
- Kept ordinary gateway requests unchanged when the internal channel-monitor header is absent.

### Testing
- `go test ./internal/handler -run 'TestResolveChannelMonitorProbePolicy|TestOpenAIResponses_ChannelMonitorPoolRetriesCountAsOneOfFiveAccounts' -count=1` from `backend/`: passed; verified three pool retries count as one account, the fifth distinct account can make the request successful, and a sixth account is not called after five failures.
- `go test ./internal/handler ./internal/service` from `backend/`: passed.
- `git diff --check` for the affected handler, service, test, documentation, and progress files: passed.

### Notes
- `backend/internal/handler/channel_monitor_probe.go`: defines the five-distinct-account monitor policy while preserving the configured global switch safety limit.
- `backend/internal/handler/channel_monitor_probe_test.go`: verifies the five-account budget and ordinary-request behavior.
- `backend/internal/handler/openai_chat_completions.go`: counts distinct monitored accounts without charging same-account pool retries.
- `backend/internal/handler/openai_gateway_handler.go`: applies the same account-based budget to Responses and Messages failover loops.
- `backend/internal/handler/openai_gateway_handler_test.go`: covers pool retry grouping, success on the fifth account, and the five-account failure cap.
- `backend/internal/service/channel_monitor_const.go`: raises the internal monitor account budget to five and extends the monitor request time budget accordingly.
- `docs/CHANNEL_MONITOR.md`: documents account-based probe budgeting, first-success status, and behavior when fewer accounts are available.
- `progress.md`: records implementation scope, verification evidence, and rollback instructions.
- Rollback: restore `ChannelMonitorProbeAttempts` to `3`, restore attempt-based accounting in the three OpenAI handler loops and probe policy, then revert the matching tests and `docs/CHANNEL_MONITOR.md` section. Do not reset the dirty worktree.

## 2026-07-19 - Task: Align SD2.0 API documentation with the deployed playground

### What was done
- Updated the user-facing International Jimeng SD2.0 section from the verified `gpt_image_playground` request behavior.
- Clarified the full, fast, fallback, and fixed-15-second model roles, including real-person capability guidance and the two distinct endpoint protocols.
- Documented 5/10/15-second string values, image/video/audio count and format limits, media duration and size limits, and the `/pg/assets` upload flow.
- Added image, video, and audio upload commands directly to the SD2.0 examples and made clear that `as-sd2.0-fast` fallback is a client strategy rather than automatic gateway behavior.

### Testing
- `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts` from `frontend/`: passed.
- `pnpm exec eslint src/views/user/ApiDocsView.vue src/views/user/__tests__/ApiDocsView.spec.ts` from `frontend/`: passed.
- `pnpm build` from `frontend/`: passed; existing Browserslist age and Vite large-chunk warnings remained non-fatal.

### Notes
- `frontend/src/views/user/ApiDocsView.vue`: expands the SD2.0 model, parameter, upload, and fallback guidance shown to users.
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: verifies the new capability labels, media limits, upload commands, and example count.
- `docs/API_DOCS.md`: records the same SD2.0 public contract and client-fallback boundary for future maintenance.
- `progress.md`: records the completed documentation task and verification evidence.
- Rollback: remove the added SD2.0 capability and upload details, restore the previous six SD2.0 examples and assertions, and revert the matching `docs/API_DOCS.md` bullets. Do not reset the dirty worktree.

## 2026-07-19 - Task: Publish the latest multi-architecture Docker image

### What was done
- Built `iotwq/china-api:latest` from the complete current workspace at application version `0.1.161` with commit marker `3908c936599b`.
- Published a Docker Hub manifest list for `linux/amd64` and `linux/arm64`.

### Testing
- `git diff --check` and `git diff --cached --check`: passed before building.
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.161 --build-arg COMMIT=3908c936599b --tag iotwq/china-api:latest --push .`: passed, including the frontend production build and both Go target builds.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with index digest `sha256:64fca3f29c1ea18a6c656d0ca6f799ea9e77f97745133db53832935e5a4b1ec8`, amd64 digest `sha256:42307f3c032c0b80d3f3ff245fc495ad1c84c4e2caae2368ad3efc01e0e7dd26`, and arm64 digest `sha256:4a40b12b090d017eba5d99ad4636130a66a4a1acb70baea43f7aa80a2edd78bf`.
- `docker run --rm --pull always --platform linux/arm64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.161 (commit: 3908c936599b)` from the remote digest.
- Existing frontend Browserslist age and Vite large-chunk warnings remained non-fatal.

### Notes
- `progress.md`: records the Docker Hub publication, remote platform digests, runtime verification, and rollback point.
- Rollback: run `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:6851c69d1976adfa01231054385c8dc0ebbf61e6c5e585da3c56f3d8fd0ca19b` to restore the previously published multi-architecture image.

## 2026-07-20 - Task: Synchronize upstream Wei-Shaw/sub2api to v0.1.162

### What was done

- Merged upstream range `d4b9797ff..e625ce3b3` (114 commits) into local `main` as `4b8bec155`, advancing the upstream baseline from `0.1.161` to `0.1.162`.
- Preserved local Gemini/Grok, OpenAI media and asset routes, community chat, channel monitoring, frontend, and API documentation customizations while resolving the five logo conflicts and two stash-restore conflicts.
- Updated the rollback API test to assert the upstream 15-minute update/rollback request timeout introduced by the synchronized implementation.

### Testing

- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm install --frozen-lockfile` from `frontend/`: passed.
- `pnpm test:run --reporter=dot`: passed, 185 test files and 1290 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build`: passed. Existing localStorage/Vue test-environment, Browserslist age, and Vite large-chunk warnings remain non-fatal.
- `git diff --check` and `git diff --cached --check`: passed; no unresolved conflicts.

### Notes

- `frontend/src/api/__tests__/admin.system.rollback.spec.ts`: aligns rollback assertions with the upstream 15-minute Axios timeout.
- `docs/UPSTREAM_SYNC.md`: records the `v0.1.162` synchronization range, conflict decisions, verification, and rollback points.
- `progress.md`: records this synchronization task and its verification evidence.
- Rollback: preserve the current worktree first, then run `git revert -m 1 4b8bec155`; the complete pre-sync worktree remains available as `backup/pre-upstream-sync-20260720-223609` and `stash@{0}` (`5d2ba650dbc5bd0fc147c3d29b6040aab7e246f1`).

## 2026-07-20 - Task: Add the recharge-ratio notice to the API keys page

### What was done

- Added a visible information notice above the API key page controls explaining that the site recharge ratio is 1:6 and the effective multiplier is the current group multiplier divided by 6.
- Added matching Chinese and English locale text and kept the notice informational only, without changing billing calculations.

### Testing

- `pnpm exec vitest run src/views/user/__tests__/KeysView.spec.ts src/i18n/__tests__/localesMessageCompile.spec.ts`: passed, 2 test files and 11 tests.
- `pnpm exec eslint src/views/user/KeysView.vue src/i18n/locales/zh/dashboard.ts src/i18n/locales/en/dashboard.ts src/views/user/__tests__/KeysView.spec.ts`: passed.
- `pnpm typecheck`: passed.
- `pnpm build`: passed; existing Browserslist age and Vite large-chunk warnings remain non-fatal.
- `git diff --check` and `git diff --cached --check`: passed.

### Notes

- `frontend/src/views/user/KeysView.vue`: renders the information notice above the page actions.
- `frontend/src/i18n/locales/zh/dashboard.ts`: adds the requested Chinese notice.
- `frontend/src/i18n/locales/en/dashboard.ts`: adds the corresponding English notice.
- `frontend/src/views/user/__tests__/KeysView.spec.ts`: verifies the notice text is rendered.
- `progress.md`: records the implementation, verification, and rollback instructions.
- Rollback: remove the `pricing-notice` block from `KeysView.vue`, remove both `keys.pricingNotice` locale entries and the matching test assertion, then remove this progress entry. Do not reset the dirty worktree.

## 2026-07-20 - Task: Enable complete public legal policies and explicit login consent

### What was done

- Added complete ChinaAPI user agreement, privacy policy, usage policy, and refund policy content with public `/legal/...` routes compatible with the referenced login experience.
- Enabled checkbox consent by default for new installations and upgraded only the original four-document empty template on existing installations; administrator-customized policies remain unchanged.
- Updated the administrator defaults and login consent copy so login and registration show `我已阅读并同意 用户协议、隐私政策、使用政策、退款政策` and block submission until accepted.
- Documented deployment checks and clarified that publishing legal policies does not by itself remove a Google Safe Browsing or Search Console warning.

### Testing

- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm test:run --reporter=dot` from `frontend/`: passed, 186 test files and 1291 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; existing Browserslist age and Vite large-chunk warnings remained non-fatal.
- `git diff --check` and `git diff --cached --check`: passed; no unresolved conflicts.

### Notes

- `.gitignore`: allows the legal policy deployment document to be tracked under `docs/`.
- `backend/internal/service/legal_policy_defaults.go`: contains the four default policies and the idempotent legacy-empty-template upgrade.
- `backend/internal/service/setting_parse.go`: runs the narrow legacy upgrade during existing-installation initialization and enables complete defaults for new installations.
- `backend/internal/service/setting_public.go`: preserves both checkbox and modal agreement modes while using the new default documents.
- `backend/internal/service/setting_service.go`: changes new agreement defaults to checkbox mode and the current policy date.
- `backend/internal/service/setting_service_update_test.go`: verifies new-installation defaults, legacy upgrades, and preservation of custom policies.
- `backend/internal/server/api_contract_test.go`: keeps API contract fixtures explicit so legacy modal responses remain covered.
- `frontend/src/views/admin/SettingsView.vue`: aligns administrator document labels and defaults with the four public policies.
- `frontend/src/i18n/locales/zh/common.ts`: renders the requested Chinese consent sentence spacing.
- `frontend/src/components/auth/__tests__/LoginAgreementPrompt.spec.ts`: verifies all four links and checkbox acceptance behavior.
- `docs/LEGAL_POLICIES.md`: records public routes, upgrade behavior, deployment verification, and Google review limitations.
- `progress.md`: records this task, verification evidence, and rollback instructions.
- Rollback: delete `backend/internal/service/legal_policy_defaults.go`, `frontend/src/components/auth/__tests__/LoginAgreementPrompt.spec.ts`, and `docs/LEGAL_POLICIES.md`; remove the `docs/LEGAL_POLICIES.md` exception from `.gitignore`; restore the old four empty documents, `modal`, `false`, and `2026-03-31` defaults in the listed setting and administrator files; remove the matching tests and this progress entry. Do not reset the dirty worktree.

## 2026-07-20 - Task: Roll back the default legal policy implementation

### What was done

- Removed the newly added built-in legal policy text, automatic legacy-template upgrade, checkbox-consent default, related frontend test, and deployment document at the user's request.
- Restored the original configuration-driven behavior: the existing four empty document templates remain disabled by default and administrators can configure them from system settings.
- Reverted the local test database settings changed during startup verification to `enabled=false`, `mode=modal`, `updated_at=2026-03-31`, and the original four empty documents.
- Stopped and removed the temporary local service dependency containers; no application image was built or pushed.

### Testing

- `go test -tags=unit ./internal/service -run 'TestSettingService_' -count=1` from `backend/`: passed.
- Queried the local PostgreSQL settings after restoration: confirmed the original disabled/modal/empty-template values.
- Confirmed the temporary Go service, PostgreSQL container, and Redis container were stopped; port `8080` was released.

### Notes

- `backend/internal/service/legal_policy_defaults.go`: removed the untracked default-policy implementation.
- `frontend/src/components/auth/__tests__/LoginAgreementPrompt.spec.ts`: removed the untracked policy-link test added in the previous task.
- `docs/LEGAL_POLICIES.md`: removed the untracked deployment document added in the previous task.
- `.gitignore`, backend settings/service files, `SettingsView.vue`, `common.ts`, and API contract fixtures: restored to their pre-policy-change worktree baseline without touching unrelated staged changes.
- `progress.md`: records this rollback and the local database cleanup.
- Rollback: the current state is intentionally the original settings-driven implementation. Re-enabling built-in policies requires explicitly restoring the removed patch; no commit was created for that optional feature, and the dirty worktree must not be reset wholesale.

## 2026-07-21 - Task: Review current sync/customizations and publish v0.1.162 multi-architecture image

### What was done

- Checked `upstream/main` after fetch; it remains at `e625ce3b3` / v0.1.162, so the local merge `4b8bec155` is current and no additional upstream synchronization was needed.
- Reviewed the synchronized/customized worktree for unresolved conflicts, build-context leakage, and the recently added API-key pricing notice; no release-blocking issue was found. The recent legal-policy implementation remains removed as requested.
- Re-ran the complete backend and frontend verification suite, then built and pushed `iotwq/china-api:latest` for both `linux/amd64` and `linux/arm64`.

### Testing

- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm test:run --reporter=dot` from `frontend/`: passed, 185 test files and 1290 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; existing Browserslist age and Vite large-chunk warnings remained non-fatal.
- `git diff --check`, `git diff --cached --check`, and unresolved-conflict check: passed.
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.162 --build-arg COMMIT=4b8bec155 --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with manifest digest `sha256:66f78fbd3b9b3f40da73fd1b24b0aed085d89fa81aec433c5150c89a03754ac3`, amd64 digest `sha256:8d81e62fcb989db49ac3c405f7d1a17535988a7f4f3f58e731beb529286cb158`, and arm64 digest `sha256:b3c8c2cdd1f39ef15ec9e15fb9a276f74af9762051c93bcff9de29bfb321b5b7`.
- `docker run --rm --pull always --platform linux/arm64 iotwq/china-api:latest --version` and the equivalent amd64 command: passed, both reported `Sub2API 0.1.162 (commit: 4b8bec155)`.

### Notes

- `progress.md`: records the release review, verification evidence, remote image digests, and rollback point; no feature source files were changed in this review/publish task.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:64fca3f29c1ea18a6c656d0ca6f799ea9e77f97745133db53832935e5a4b1ec8`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-07-21 - Task: Refund terminal SD2.0 video generation errors

### What was done

- Extended video status failure recognition to cover the observed `server_error` payload whose message explicitly contains `generation failed`, while keeping generic HTTP and server errors non-refundable.
- Preserved the task ID and response body when an upstream video status query returns an HTTP error, allowing the existing idempotent balance refund path to process confirmed terminal generation failures.
- Applied the same refund decision before normal and error response handling, so both HTTP 200 error payloads and HTTP 4xx/5xx terminal failure payloads produce the existing negative “video failure refund” usage record.

### Testing

- Added regression coverage for the observed `generation failed: generate error: An error occurred.` payload, generic `server_error`, invalid-request errors, and HTTP 500 status responses retaining refund metadata.
- `go test ./internal/service -run 'TestOpenAIVideoStatusFailed|TestForwardOpenAIVideoStatusPreservesTerminalErrorResponse|TestRefundFailedOpenAIVideoTask' -count=1` from `backend/`: passed.
- `go test ./internal/service ./internal/handler -count=1` from `backend/`: passed.
- `go test -tags=unit ./... -count=1` from `backend/`: passed.

### Notes

- `backend/internal/service/openai_videos.go`: recognizes the strict terminal error shape and retains status error response metadata.
- `backend/internal/service/openai_videos_test.go`: covers terminal error classification and HTTP 500 response preservation.
- `backend/internal/handler/openai_videos.go`: triggers the existing refund path for classified status responses before returning an upstream HTTP error.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents video failure refund triggers, exclusions, idempotency, and polling requirements.
- `progress.md`: records this task and verification evidence.
- Rollback: remove the strict `error.type == server_error` and `generation failed` classification, restore HTTP error forwarding to return no `OpenAIForwardResult`, move the refund check back below successful forwarding, remove the matching regression test and documentation section, then remove this progress entry. Preserve the dirty worktree and do not use a wholesale reset.

## 2026-07-21 - Task: Handle SD2.0 video 500/502/503/504 failures safely

### What was done

- Extended terminal video failure recognition to support `server_error`, `upstream_error`, missing error type, and plain-text responses when the message explicitly contains `generation failed`.
- Covered HTTP 500, 502, 503, and 504 status responses while keeping generic gateway and transient error pages non-refundable; status codes alone never trigger a refund.
- Updated the media compatibility documentation to describe retry behavior, terminal-failure matching, and the distinction between releasing an uncharged submission hold and refunding a completed charge.

### Testing

- Added regression coverage for terminal and transient JSON/plain-text error bodies.
- Added HTTP 500/502/503/504 status forwarding coverage with task ID and response-body preservation.
- `go test ./internal/service -run 'TestOpenAIVideoStatusFailed|TestForwardOpenAIVideoStatusPreservesTerminalErrorResponse|TestRefundFailedOpenAIVideoTask' -count=1` from `backend/`: passed.
- `go test -tags=unit ./internal/service ./internal/handler -count=1` from `backend/`: passed.

### Notes

- `backend/internal/service/openai_videos.go`: classifies explicit terminal generation errors across supported response shapes.
- `backend/internal/service/openai_videos_test.go`: covers 500/502/503/504 terminal responses and transient gateway errors.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents the safe handling and retry rules for video upstream errors.
- `progress.md`: records this task and verification evidence.
- Rollback: restore the previous `OpenAIVideoStatusFailed` implementation that only accepts status fields and `server_error` JSON, remove the added status/error tests and documentation wording, then remove this progress entry. Preserve the dirty worktree and do not use a wholesale reset.

## 2026-07-22 - Task: Publish latest SD2.0 refund fixes as a multi-architecture image

### What was done

- Built the current customized worktree as `iotwq/china-api:latest` with application version `0.1.162` and source baseline commit `4b8bec155`.
- Published a Docker manifest containing native `linux/amd64` and `linux/arm64` images, including the latest SD2.0 terminal failure and refund handling.

### Testing

- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.162 --build-arg COMMIT=4b8bec155 --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with manifest digest `sha256:41a62045d9d28ae2df6b3b6d821140a38d57f5cac06ecfd7d86b6b5f622a4c69`, amd64 digest `sha256:535946c535372e14880b4c25299d2c4d8fdeb5609d448053774665b6ecce221b`, and arm64 digest `sha256:bc94b43da550fe52afc6855e00835a73690c4a871c2077014b3923ba5cdd27dc`.
- `docker run --rm --pull always --platform linux/amd64 iotwq/china-api:latest --version` and the equivalent arm64 command: passed; both reported `Sub2API 0.1.162 (commit: 4b8bec155)`.

### Notes

- `progress.md`: records this image publication, remote manifest digests, architecture verification, and rollback point; no business source files were changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:66f78fbd3b9b3f40da73fd1b24b0aed085d89fa81aec433c5150c89a03754ac3`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-07-22 - Task: Synchronize upstream/main v0.1.163

### What was done

- Fetched and merged 69 upstream commits in `e625ce3b3..60013c5f1`, advancing the upstream baseline from v0.1.162 to v0.1.163 in merge commit `87c92d9d5f8e4e2da9d634eaf1ada4b46c22d3af`.
- Preserved the existing dirty worktree through a backup branch and stash, then restored all 211 staged paths, 7 unstaged paths, and 45 untracked paths after the merge.
- Resolved six overlapping OpenAI/Grok scheduler and gateway files by keeping upstream scheduling diagnostics, Grok content-policy isolation and model synchronization while preserving the fork's Gemini-native scheduling, passthrough model behavior, wrapped-status retry rules, and custom Grok APIKey upstream security policy.

### Testing

- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm install --frozen-lockfile` from `frontend/`: passed.
- `pnpm test:run --reporter=dot` from `frontend/`: passed, 190 test files and 1317 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; existing Browserslist age and Vite large-chunk warnings remained non-fatal.
- `git diff --check`, `git diff --cached --check`, unresolved-conflict check, and `git merge-base --is-ancestor upstream/main HEAD`: passed.
- `go mod tidy -diff` only reported removable historical `go.sum` entries; no unrelated dependency cleanup was applied.

### Notes

- `backend/internal/service/openai_account_scheduler.go`: resolved the scheduler overlap while retaining upstream exclusion diagnostics and local Gemini-native account capability behavior.
- `backend/internal/service/openai_gateway_cc_pipeline.go`: retained local strict wrapped-status recognition in the upstream retry flow.
- `backend/internal/service/openai_gateway_chat_completions_raw.go`: retained local wrapped-status retry and passthrough compatibility alongside upstream behavior.
- `backend/internal/service/openai_gateway_grok.go`: combined upstream content-policy isolation with normalized retry, cooldown, and failover decisions.
- `backend/internal/service/openai_gateway_grok_chat_bridge.go`: kept the same normalized Grok retry behavior for the chat bridge.
- `backend/internal/service/upstream_models.go`: accepted upstream Grok OAuth/APIKey model synchronization while preserving local custom APIKey URL security handling.
- `docs/UPSTREAM_SYNC.md`: documents the v0.1.163 range, compatibility decisions, verification evidence, and recovery points.
- `progress.md`: records this synchronization and its verification evidence.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 87c92d9d5f8e4e2da9d634eaf1ada4b46c22d3af`. The pre-sync branch is `backup/pre-upstream-sync-20260722-214109`, and the full pre-sync worktree stash object is `6fa7433ae927b3320474cabc4d0b70512691874a`.

## 2026-07-22 - Task: Improve community chat reliability and add direct-message deletion

### What was done

- Added cursor-based group/direct history loading, reconnect backfill, immediate HTTP send acknowledgement, draft-preserving sends, stable scroll restoration, and stale admin-conversation request protection.
- Added server-side per-reader/per-conversation direct-message read state, admin per-user unread counts, private-thread bottom-only read acknowledgement, and shared sidebar/chat-page WebSocket usage.
- Added soft deletion for direct messages: senders can delete their own private messages, administrators can delete any private message, and deletion events are sent only to the conversation user and administrators.
- Added Redis Pub/Sub fan-out for multi-instance realtime events and fail-open per-IP limits for chat messages, uploads, and delete operations.

### Testing

- `go generate ./cmd/server` from `backend/`: passed and regenerated the Wire dependency graph with Redis-backed community chat service startup.
- `go test -tags=unit ./internal/service ./internal/repository ./internal/handler ./internal/server/routes ./internal/server -run 'CommunityChat|DirectUnread|DeleteDirect' -count=1` from `backend/`: passed.
- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm test:run --reporter=dot` from `frontend/`: passed, 191 test files and 1321 tests.
- Final focused chat suite: 5 test files and 28 tests passed; `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` also passed.
- `git diff --check`, `git diff --cached --check`, and unresolved-conflict check: passed.
- Manual two-account browser testing was not run because the local PostgreSQL/Redis application stack was not active; this remains a deployment smoke-test item.

### Notes

- `backend/migrations/186_community_chat_reliability.sql`: adds private-message soft-delete columns and per-reader/per-conversation read state.
- `backend/internal/service/community_chat.go`: adds cursor history, direct unread/read/delete operations, deletion events, and Redis event fan-out.
- `backend/internal/service/community_chat_test.go`: verifies direct-delete permissions, unread aggregation, and cross-instance Redis delivery without self-duplicates.
- `backend/internal/repository/community_chat_repo.go`: implements visible-message cursor queries, private soft deletion, unread aggregation, and monotonic read cursors.
- `backend/internal/repository/community_chat_repo_test.go`: verifies deleted-message filtering, per-conversation unread grouping, and soft deletion.
- `backend/internal/handler/community_chat_handler.go`: exposes cursor, unread, read and private-delete handlers and protects private deletion event visibility.
- `backend/internal/handler/community_chat_handler_test.go`: verifies cursor validation and private deletion event audience rules.
- `backend/internal/server/routes/community_chat.go`: registers the new endpoints and fail-open chat/upload/delete rate limits.
- `backend/internal/server/router.go`: passes the existing Redis client into community chat route registration.
- `backend/internal/service/wire.go`: starts the community chat Redis subscriber through dependency injection.
- `backend/cmd/server/wire_gen.go`: regenerates the Redis-aware community chat service wiring.
- `frontend/src/api/communityChat.ts`: adds cursor, direct unread/read/delete contracts and the private-delete event type.
- `frontend/src/api/index.ts`: exports the direct unread API type.
- `frontend/src/api/__tests__/communityChat.spec.ts`: verifies cursor and direct unread/read/delete requests.
- `frontend/src/composables/useCommunityChatRealtime.ts`: provides one reconnecting WebSocket shared by all mounted chat consumers.
- `frontend/src/composables/__tests__/useCommunityChatRealtime.spec.ts`: proves multiple subscribers share one socket and close it only when idle.
- `frontend/src/views/user/CommunityChatView.vue`: adds reliable merging/backfill, history pagination, stable scrolling, draft protection, direct unread counts and private deletion UI.
- `frontend/src/views/user/__tests__/communityChatUnread.spec.ts`: verifies server unread handling, bottom-only read acknowledgement and private deletion UI.
- `frontend/src/components/layout/AppSidebar.vue`: uses the shared socket and restores private unread state from the server.
- `frontend/src/components/layout/__tests__/AppSidebar.spec.ts`: verifies separate group/private unread state and removal of the sidebar-owned socket.
- `docs/COMMUNITY_CHAT.md`: documents reliability behavior, endpoints, permissions, rate limits, Redis and shared attachment storage requirements.
- `progress.md`: records this task, verification evidence, remaining browser smoke test and rollback point.
- Rollback: redeploy the previous published image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:41a62045d9d28ae2df6b3b6d821140a38d57f5cac06ecfd7d86b6b5f622a4c69`. If migration 186 has already run, first stop application writes, then execute `DROP TABLE IF EXISTS community_chat_direct_read_states; DROP INDEX IF EXISTS idx_community_chat_direct_messages_visible_conversation; ALTER TABLE community_chat_direct_messages DROP COLUMN IF EXISTS deleted_by, DROP COLUMN IF EXISTS deleted_at;`, and deploy the previous image. Preserve the dirty worktree; do not use a wholesale reset.

## 2026-07-23 - Task: Publish latest community chat improvements as a multi-architecture image

### What was done

- Built the current customized worktree as `iotwq/china-api:latest` with application version `0.1.163` and source baseline commit `87c92d9d5`.
- Published a Docker manifest containing native `linux/amd64` and `linux/arm64` images, including the latest community chat reliability and direct-message deletion changes.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, and unresolved-conflict checks: passed.
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.163 --build-arg COMMIT=87c92d9d5 --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with manifest digest `sha256:2c78e06ab51a216ac44610c69c46c156f01402543b23625fff72dcbbea6b45d5`, amd64 digest `sha256:0f256647426338036e948c2475918b509f06c0eb8cd53aaf36233a2a489b1ea4`, and arm64 digest `sha256:fa000272db610c59a67309ce72e20101a158facba7044a20d64e934eea26d234`.
- `docker run --rm --pull always --platform linux/amd64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.163 (commit: 87c92d9d5)`.
- `docker run --rm --pull always --platform linux/arm64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.163 (commit: 87c92d9d5)`.

### Notes

- `progress.md`: records this image publication, remote manifest digests, architecture verification, and rollback point; no business source files were changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:41a62045d9d28ae2df6b3b6d821140a38d57f5cac06ecfd7d86b6b5f622a4c69`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-07-23 - Task: Synchronize upstream/main v0.1.164

### What was done

- Fetched and merged 43 upstream commits in `60013c5f1..cb24522dd`, advancing the upstream baseline from v0.1.163 to v0.1.164 in merge commit `8423768fc999ce5e0344ff5945cc247a8de04f9c`.
- Preserved the existing dirty worktree through backup branch `backup/pre-upstream-sync-20260723-223356` and stash `03ad1dea8f06e3bd10490e9bf6afc39f52def1cb`, then restored the local staged, unstaged, and untracked changes.
- Resolved five overlapping files by retaining upstream composite routing, billing, and proxy-circuit behavior while preserving local Gemini-native OpenAI routing, media balance holds, OpenAI video task APIs, audio/assets endpoints, and Grok media compatibility.

### Testing

- `go generate ./cmd/server` from `backend/`: passed.
- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm test:run --reporter=dot` from `frontend/`: passed, 195 test files and 1353 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; existing Browserslist and Vite large-chunk warnings remained non-fatal.
- `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q`: passed.
- `git diff --check`, `git diff --cached --check`, unresolved-conflict check, and `git merge-base --is-ancestor upstream/main HEAD`: passed.
- `go mod tidy -diff` only reported removable historical `go.sum` entries; no unrelated dependency cleanup was applied.

### Notes

- `backend/internal/handler/gemini_v1beta_handler.go`: combines upstream effective composite target-platform checks with OpenAI Gemini-native account selection and failover behavior.
- `backend/internal/handler/openai_images.go`: combines upstream composite-model usage fields with local media balance hold capture.
- `backend/internal/server/routes/gateway.go`: combines upstream composite media routing with local OpenAI video, audio, Nano Banana, and `/pg/assets` routes.
- `backend/internal/server/routes/gateway_test.go`: retains composite Grok lookup coverage and OpenAI video route/rejection coverage.
- `deploy/docker-compose.dev.yml`: keeps upstream proxy stream circuit settings and local image workspace URL.
- `backend/cmd/server/wire_gen.go`: regenerated dependency wiring for the upstream composite/Ollama additions and local media/chat services.
- `docs/UPSTREAM_SYNC.md`: records this synchronization range, compatibility decisions, verification evidence, and recovery points.
- `progress.md`: records this synchronization and its verification evidence.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 8423768fc999ce5e0344ff5945cc247a8de04f9c`. The pre-sync branch is `backup/pre-upstream-sync-20260723-223356`, and the full pre-sync worktree stash object is `03ad1dea8f06e3bd10490e9bf6afc39f52def1cb`.

## 2026-07-23 - Task: Publish synchronized v0.1.164 as a multi-architecture image

### What was done

- Built the current customized worktree as `iotwq/china-api:latest` with application version `0.1.164` and source baseline commit `8423768fc`.
- Published a Docker manifest containing native `linux/amd64` and `linux/arm64` images, including the synchronized upstream v0.1.164 features and all restored local customizations.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, and unresolved-conflict checks: passed.
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.164 --build-arg COMMIT=8423768fc --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with manifest digest `sha256:954ced38755d29defbc07ec822b85ad2ae2487f18dc1dc5c91d942c438d81b89`, amd64 digest `sha256:3467bde1e7cc2beb48dc2212fd21e532348f6eecdf20275b23d21f0a0247b138`, and arm64 digest `sha256:d8f38165b4f5d6afad1a256a9546a68190eca75f32ed9a6e7e9208d35db33d43`.
- `docker run --rm --pull always --platform linux/amd64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.164 (commit: 8423768fc)`.
- `docker run --rm --pull always --platform linux/arm64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.164 (commit: 8423768fc)`.

### Notes

- `progress.md`: records this image publication, remote manifest digests, architecture verification, and rollback point; no business source files were changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:2c78e06ab51a216ac44610c69c46c156f01402543b23625fff72dcbbea6b45d5`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-07-26 - Task: Synchronize upstream/main v0.1.165

### What was done

- Fetched and merged 54 upstream commits in `cb24522dd..2730c1c43`, advancing the upstream baseline from v0.1.164 to v0.1.165 in merge commit `235879a305bd3d168ceef903a18c9be2d87ca39e`.
- Preserved the dirty worktree through backup branch `backup/pre-upstream-sync-20260726-144816` and full stash `9205b0ba0d922e8f52ae9321b5cfaa9ab6da0173`, then restored all 263 changed paths without omissions.
- Combined upstream OpenAI Live, session ID, image logging, retry, Gemini image, Ollama usage and registration safety changes with the fork's Gemini-native, media billing/video, Grok custom upstream, channel monitoring and community chat customizations.
- Updated local GroupsView tests for the upstream Live capability probe without changing production behavior.

### Testing

- `go generate ./cmd/server` from `backend/`: passed.
- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `go mod tidy -diff` only reported removable historical `go.sum` entries; no unrelated dependency cleanup was applied.
- `pnpm install --frozen-lockfile` from `frontend/`: passed.
- Focused GroupsView compatibility suite: passed, 2 test files and 10 tests.
- `pnpm test:run --reporter=dot` from `frontend/`: passed, 197 test files and 1366 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; existing Browserslist and Vite large-chunk warnings remained non-fatal.
- `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q`: passed.
- `git diff --check`, `git diff --cached --check`, unresolved-conflict check, and `git merge-base --is-ancestor upstream/main HEAD`: passed.

### Notes

- `backend/internal/handler/openai_images.go`: retains local media balance reservation/capture and extended image metadata while adding upstream session ID recording and image log field names.
- `backend/internal/service/openai_gateway_service.go`: combines the local OpenAI video task binding repository with upstream Live attestation dependencies.
- `backend/cmd/server/wire_gen.go`: regenerated the merged dependency graph.
- `frontend/src/views/admin/__tests__/GroupsView.columnSettings.spec.ts`: adds an isolated Live capability API mock for local column tests.
- `frontend/src/views/admin/__tests__/GroupsView.duplicate.spec.ts`: adds an isolated Live capability API mock for local duplication tests.
- `docs/UPSTREAM_SYNC.md`: records the v0.1.165 synchronization scope, compatibility decisions, verification evidence and recovery points.
- `progress.md`: records this synchronization and its verification evidence.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 235879a305bd3d168ceef903a18c9be2d87ca39e`. The pre-sync branch is `backup/pre-upstream-sync-20260726-144816`, and the full pre-sync worktree stash object is `9205b0ba0d922e8f52ae9321b5cfaa9ab6da0173`.

## 2026-07-26 - Task: Publish synchronized v0.1.165 as a multi-architecture image

### What was done

- Built the current customized worktree as `iotwq/china-api:latest` with application version `0.1.165` and source baseline commit `235879a30`.
- Published a Docker manifest containing native `linux/amd64` and `linux/arm64` images, including the synchronized upstream v0.1.165 features and all restored local customizations.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, and unresolved-conflict checks: passed.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.165 --build-arg COMMIT=235879a30 --tag iotwq/china-api:latest --push .`: passed.
- Two earlier build attempts stopped before project compilation because npm Registry and Alpine package downloads returned transient network errors; neither attempt updated the remote tag.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with manifest digest `sha256:ebd77b306eb830269ca88692a3c36913674cb256b207fae9beb7b880c5be85b9`, amd64 digest `sha256:5560b3f8c0103e0cfd727afe595a05366f49cab276aaa79cfea827a6ee47fc50`, and arm64 digest `sha256:2d5a85100611c0dd0d48b3783bd46b10d531b9de8532e79300c45852d1e5a238`.
- `docker run --rm --pull always --platform linux/amd64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.165 (commit: 235879a30)`.
- `docker run --rm --pull always --platform linux/arm64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.165 (commit: 235879a30)`.

### Notes

- `progress.md`: records this image publication, remote manifest digests, architecture verification, transient registry retries, and rollback point; no business source files were changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:954ced38755d29defbc07ec822b85ad2ae2487f18dc1dc5c91d942c438d81b89`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-07-27 - Task: Add five-account probing for Anthropic channel monitoring

### What was done

- Added the existing five-distinct-account monitor probe budget to Anthropic checks.
- Applied that budget to the Anthropic Messages gateway while preserving pool-mode same-account retries as one account attempt.
- Stopped Anthropic monitor probes from recycling failed accounts after all eligible accounts have been exhausted; the first successful account still determines the monitor response and channel status.

### Testing

- `go test -tags=unit ./internal/handler ./internal/service -run 'Test(ChannelMonitorProbePolicyCapsFailoverAtFiveAccounts|ResolveChannelMonitorProbePolicy|RunCheckForModel_OffMode_PreservesDefaultBody|RunCheckForModel_OpenAI_DefaultChatRequest)$' -count=1` from `backend/`: passed.
- `go test -tags=unit ./internal/handler ./internal/service -count=1` from `backend/`: passed.
- `git diff --check`: passed.

### Notes

- `backend/internal/service/channel_monitor_checker.go`: sends the internal five-account probe budget on Anthropic checks.
- `backend/internal/service/channel_monitor_checker_body_test.go`: verifies the Anthropic checker sends the internal probe header.
- `backend/internal/handler/channel_monitor_probe.go`: prevents monitor probes from reusing failed accounts after candidate exhaustion.
- `backend/internal/handler/channel_monitor_probe_test.go`: verifies ordinary requests remain unchanged and monitor failover stops after five different accounts.
- `backend/internal/handler/gateway_handler.go`: applies the monitor-aware switch budget to the Anthropic Messages failover loop.
- `docs/CHANNEL_MONITOR.md`: documents Anthropic five-account probing and account-exhaustion behavior.
- `progress.md`: records the implementation, verification evidence, and rollback command.
- Rollback: from the repository root, run `git diff -- backend/internal/service/channel_monitor_checker.go backend/internal/service/channel_monitor_checker_body_test.go backend/internal/handler/channel_monitor_probe.go backend/internal/handler/channel_monitor_probe_test.go backend/internal/handler/gateway_handler.go docs/CHANNEL_MONITOR.md | git apply --reverse`; this reverses only the unstaged task patch and preserves earlier staged customizations.

## 2026-07-27 - Task: Publish Anthropic monitor update as a multi-architecture image

### What was done

- Built the current customized worktree as `iotwq/china-api:latest` with application version `0.1.165` and source baseline commit `235879a30`.
- Published native `linux/amd64` and `linux/arm64` images containing the Anthropic five-account channel-monitor probing update and all existing local customizations.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, and unresolved-conflict checks: passed.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.165 --build-arg COMMIT=235879a30 --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with manifest digest `sha256:26b5e3ef8f920cd43eb7269164697c7fecd82c0c489d9afaa37d31f6ee92220d`, amd64 digest `sha256:7a77eaae22b0eb22712b2d0aec0c4a933382c33e1ee3efbef90656e98ff984d7`, and arm64 digest `sha256:843eca8c8c42f7c5fd5c3b6050c6e2b854d4fca807500f56f16eedcf59d6ece1`.
- `docker run --rm --pull always --platform linux/amd64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.165 (commit: 235879a30)`.
- `docker run --rm --pull always --platform linux/arm64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.165 (commit: 235879a30)`.

### Notes

- `progress.md`: records the multi-architecture publication, remote digests, runtime verification, and rollback point; no business source files were changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:ebd77b306eb830269ca88692a3c36913674cb256b207fae9beb7b880c5be85b9`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-07-27 - Task: Synchronize upstream/main v0.1.166

### What was done

- Fetched and merged 62 upstream commits in `2730c1c43..59ce11c78`, advancing the upstream baseline from v0.1.165 to v0.1.166 in merge commit `aa04f43bc292a96db4e1c9b63cb6f83fccf6108e`.
- Preserved the dirty worktree through backup branch `backup/pre-upstream-sync-20260727-213337` and full stash `d5351da0a0fd0b573ee462d20b6e3c7f23192b42`, then restored all 266 changed paths.
- Combined upstream panel API rate limiting and Grok WebSocket model-mapping coverage with the fork's community chat route and existing media, channel-monitor, Gemini-native, Grok custom-upstream, frontend, and API documentation customizations.
- Restored the two `github.com/google/subcommands v1.2.0` checksums required by the Wire command so the merged dependency graph remains reproducibly generatable.

### Testing

- Initial `go generate ./cmd/server` identified the missing Wire command checksum; after adding only those two `go.sum` entries, the command passed.
- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm install --frozen-lockfile` from `frontend/`: passed.
- `pnpm test:run --reporter=dot` from `frontend/`: passed, 200 test files and 1383 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; existing localStorage, Vue test-environment, Browserslist, and large-chunk warnings remained non-fatal.
- `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q`: passed.
- `bash deploy/test-caddyfile-cache.sh`: passed.
- `git diff --check`, `git diff --cached --check`, unresolved-conflict check, and `git merge-base --is-ancestor upstream/main HEAD`: passed.
- `go mod tidy -diff` only proposed removing the two checksums required by `go generate`; they were intentionally retained without changing `go.mod`.

### Notes

- `backend/internal/server/router.go`: keeps upstream panel rate limiting on standard panel routes and restores the local community chat route.
- `backend/internal/service/openai_ws_http_bridge_test.go`: keeps upstream Grok mapped-model WebSocket bridge coverage together with local bridge tests.
- `backend/go.sum`: restores only the Wire CLI's missing `google/subcommands` checksums.
- `backend/cmd/server/wire_gen.go`: was regenerated successfully against the merged dependency graph.
- `docs/UPSTREAM_SYNC.md`: records the v0.1.166 synchronization scope, compatibility decisions, verification evidence, and recovery points.
- `progress.md`: records this synchronization and its verification evidence.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 aa04f43bc292a96db4e1c9b63cb6f83fccf6108e`. The pre-sync branch is `backup/pre-upstream-sync-20260727-213337`, and the full pre-sync worktree stash object is `d5351da0a0fd0b573ee462d20b6e3c7f23192b42`.

## 2026-07-27 - Task: Publish synchronized v0.1.166 as a multi-architecture image

### What was done

- Built the complete current customized worktree as `iotwq/china-api:latest` with application version `0.1.166` and source baseline commit `aa04f43bc`.
- Published native `linux/amd64` and `linux/arm64` images to Docker Hub under one multi-architecture OCI index.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, and unresolved-conflict checks: passed.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.166 --build-arg COMMIT=aa04f43bc --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with manifest digest `sha256:9a55057c6d8a43d3144f9abb4c0e67b07e11bc504555d35312ad5ff0680a0f7a`, amd64 digest `sha256:20840f5a73ae6bfcd8a7ecdbadf1a26a01afc3f4b1487e0a01986f8a582a286c`, and arm64 digest `sha256:a92bd3d2a44b66921ec2751ddb52552ec97a50c50b8fc102ea2fd0b7ad6616b5`.
- `docker run --rm --pull always --platform linux/amd64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.166 (commit: aa04f43bc)`.
- `docker run --rm --pull always --platform linux/arm64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.166 (commit: aa04f43bc)`.

### Notes

- `progress.md`: records this Docker Hub publication, remote architecture digests, runtime verification, and rollback point; no business source files were changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:26b5e3ef8f920cd43eb7269164697c7fecd82c0c489d9afaa37d31f6ee92220d`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-07-30 - Task: Synchronize upstream/main v0.1.168

### What was done

- Fetched and merged 37 upstream commits in `59ce11c78..5a6143097`, advancing the upstream baseline from v0.1.166 to v0.1.168 in merge commit `878ff00d5ef26f43ca7c9aea894f4d20bc9bfc2f`.
- Preserved the dirty worktree through backup branch `backup/pre-upstream-sync-20260730-210827` and full stash `c950cbac882d2940d5c69a09261ab7dc4c6531ee`, then restored the staged, unstaged, and untracked customizations without deleting the stash.
- Combined upstream Passkey authentication and public model plaza routing with the fork's customized login experience, community chat, public media assets, Gemini-native, Grok custom-upstream, channel-monitor, and API documentation features.
- Reconciled the WebAuthn dependency checksums and regenerated the complete Wire dependency graph.

### Testing

- `go generate ./cmd/server` from `backend/`: passed after retaining the Wire CLI's required `github.com/google/subcommands v1.2.0` checksums.
- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm install --frozen-lockfile` from `frontend/`: passed.
- `pnpm test:run --reporter=dot` from `frontend/`: passed, 203 test files and 1401 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; existing localStorage, Vue test-environment, Browserslist, and large-chunk warnings remained non-fatal.
- `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q`: passed.
- `bash deploy/test-caddyfile-cache.sh`: passed.
- `git diff --check`, `git diff --cached --check`, unresolved-conflict check, and `git merge-base --is-ancestor upstream/main HEAD`: passed.
- `go mod tidy -diff` only proposed removing the two checksums required by `go generate`; they were intentionally retained without changing `go.mod`.

### Notes

- `backend/internal/server/router.go`: registers both the upstream model plaza and the local community chat using their respective authentication and rate-limit middleware.
- `frontend/src/views/auth/LoginView.vue`: retains the customized login layout and agreement/OAuth flow while exposing upstream Passkey sign-in.
- `backend/internal/handler/handler.go` and `backend/internal/handler/wire.go`: retain local community chat/public asset handlers and include upstream Passkey/model plaza handlers.
- `backend/cmd/server/wire_gen.go`: was regenerated successfully against the combined dependency graph.
- `backend/go.sum`: adds the WebAuthn transitive checksums, removes obsolete JWT checksums, and retains the Wire CLI checksums required for reproducible generation.
- `docs/UPSTREAM_SYNC.md`: records the v0.1.168 synchronization scope, compatibility decisions, verification evidence, and recovery points.
- `progress.md`: records this synchronization and its verification evidence.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 878ff00d5ef26f43ca7c9aea894f4d20bc9bfc2f`. The pre-sync branch is `backup/pre-upstream-sync-20260730-210827`, and the full pre-sync worktree stash object is `c950cbac882d2940d5c69a09261ab7dc4c6531ee`.

## 2026-07-30 - Task: Base channel status on the final successful account probe

### What was done

- Changed current-sub2api channel monitor checks to classify green or yellow from the final successful account attempt latency instead of the total wall-clock time accumulated across earlier failed attempts.
- Kept external compatible endpoints backward compatible: when no internal successful-attempt latency is returned, monitoring continues to use the complete request latency.
- Removed the previous summary anti-jitter override so a run in which every applicable account fails remains red instead of being changed to yellow by an older successful history entry.
- Applied the monitor-only timing path to OpenAI Chat Completions, OpenAI Responses, OpenAI-to-Anthropic Messages, Anthropic Messages, and Grok's shared OpenAI-compatible chat route without changing ordinary request retry behavior.

### Testing

- Focused handler tests for probe-policy limits, current-attempt latency reset, and ordinary-request header isolation: passed.
- Focused service tests for fast/slow successful-attempt classification, all-failure status aggregation, external endpoint behavior, Responses fallback, Grok, and Anthropic response parsing: passed.
- `go test -tags=unit ./internal/handler -count=1` from `backend/`: passed in 30.919 seconds.
- `go test -tags=unit ./internal/service -count=1` from `backend/`: passed in 143.609 seconds.
- Final `git diff --check`, `git diff --cached --check`, and unresolved-conflict check: passed.

### Notes

- `backend/internal/handler/channel_monitor_probe.go`: wraps monitor-only responses and resets the reported latency at the start of each account attempt.
- `backend/internal/handler/channel_monitor_probe_test.go`: verifies previous failed-attempt time is excluded and ordinary requests do not expose the monitor metric.
- `backend/internal/handler/gateway_handler.go`: marks each Anthropic Messages account attempt for monitor timing.
- `backend/internal/handler/openai_chat_completions.go`: marks each OpenAI/Grok-compatible chat account attempt for monitor timing.
- `backend/internal/handler/openai_gateway_handler.go`: marks each OpenAI Responses and OpenAI Messages bridge account attempt for monitor timing.
- `backend/internal/service/channel_monitor_checker.go`: accepts the successful-attempt latency and uses it for stored latency and green/yellow classification.
- `backend/internal/service/channel_monitor_checker_body_test.go`: covers fast final success as green and slow final success as yellow.
- `backend/internal/service/channel_monitor_const.go`: defines the internal successful-attempt latency response header and removes the obsolete anti-jitter window.
- `backend/internal/service/channel_monitor_aggregator.go`: makes the latest completed run authoritative for the channel summary status.
- `backend/internal/service/channel_monitor_aggregator_test.go`: verifies a latest all-failure result remains an error/red status.
- `docs/CHANNEL_MONITOR.md`: documents successful-attempt latency classification, all-failure red status, and external endpoint fallback behavior.
- `progress.md`: records this implementation, verification evidence, changed files, and rollback instructions.
- Rollback: revert only the files listed in this task record to their state before this entry; this restores cumulative request latency and the previous recent-history yellow anti-jitter behavior without changing the five-account probe budget.

## 2026-07-30 - Task: Publish v0.1.168 as a multi-architecture image

### What was done

- Built the complete current customized worktree as `iotwq/china-api:latest` with application version `0.1.168` and source baseline commit `878ff00d5`.
- Published native `linux/amd64` and `linux/arm64` images to Docker Hub under one multi-architecture OCI index.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, and unresolved-conflict checks: passed.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.168 --build-arg COMMIT=878ff00d5 --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with manifest digest `sha256:a592a6e87d60f64f926f0104291a9e5b5a3a3c8407f217c0bd535c104a5cda3b`, amd64 digest `sha256:2ab3449f99f13bb4aa466f410d75493bd68975cc70249d341acb0e35a604f472`, and arm64 digest `sha256:44b7c70e4eb1bf21a42c4a350644edf1eeab4ba3ec02aa75480054abe0071e86`.
- `docker run --rm --pull always --platform linux/amd64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.168 (commit: 878ff00d5)` after retrying one transient Docker Hub CDN `EOF`.
- `docker run --rm --pull always --platform linux/arm64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.168 (commit: 878ff00d5)`.

### Notes

- `progress.md`: records this Docker Hub publication, remote architecture digests, runtime verification, and rollback point; no business source files were changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:9a55057c6d8a43d3144f9abb4c0e67b07e11bc504555d35312ad5ff0680a0f7a`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-07-31 - Task: Continue OpenAI channel probes after upstream HTTP 400

### What was done

- Changed OpenAI channel-monitor probes so any upstream HTTP 400 excludes the current account and continues with the next distinct eligible account, within the existing five-account budget.
- Applied the monitor-only rule to Chat Completions, Responses, and Responses passthrough forwarding without changing ordinary user-request handling.
- Preserved existing pool behavior: HTTP 400 only retries on the same pool account when that account explicitly includes 400 in `pool_mode_retry_status_codes`.
- Kept successful status classification unchanged: the first later success ends probing and its single-attempt latency determines green or yellow; five applicable account failures remain red.

### Testing

- `go test -tags=unit ./internal/service -run '^TestChannelMonitorBadRequestFailoverIsMonitorOnly$' -count=1`: passed; a generic HTTP 400 becomes next-account failover only with the exact internal monitor header.
- `go test -tags=unit ./internal/handler -run '^TestOpenAIResponses_ChannelMonitorBadRequestContinuesToHealthyAccount$' -count=1`: passed; verified `400, 400, success`, ordinary-request non-retry, and a six-account all-400 pool capped at the first five accounts.
- `go test -tags=unit ./internal/handler -count=1`: passed in 30.017 seconds.
- `go test -tags=unit ./internal/service -count=1`: passed in 146.762 seconds.
- Final `git diff --check`, `git diff --cached --check`, and unresolved-conflict check: passed.

### Notes

- `backend/internal/service/openai_gateway_upstream_errors.go`: recognizes the exact internal monitor probe header for monitor-only HTTP 400 failover.
- `backend/internal/service/openai_gateway_cc_pipeline.go`: applies the monitor-only rule to OpenAI-compatible Chat Completions forwarding.
- `backend/internal/service/openai_gateway_forward.go`: applies the monitor-only rule to transformed Responses forwarding.
- `backend/internal/service/openai_gateway_passthrough.go`: applies the monitor-only rule to Responses passthrough forwarding.
- `backend/internal/service/openai_gateway_service_codex_cli_only_test.go`: verifies monitor and ordinary HTTP 400 classification remain isolated.
- `backend/internal/handler/openai_gateway_handler_test.go`: verifies distinct-account switching, later success, ordinary-request isolation, and the five-account cap.
- `docs/CHANNEL_MONITOR.md`: documents the OpenAI monitor-only HTTP 400 exception and unchanged ordinary-request behavior.
- `progress.md`: records this implementation, verification evidence, changed files, and rollback point.
- Rollback point: restore the seven implementation, test, and documentation files listed above to their state immediately before this entry; remove only the monitor-only HTTP 400 helper, its three forwarding call sites, the two new tests, and the corresponding documentation paragraphs, preserving all earlier channel-monitor latency and five-account changes.

## 2026-07-31 - Task: Synchronize upstream/main v0.1.169

### What was done

- Fetched and merged 63 upstream commits in `5a6143097..2980ff385`, advancing the version file from v0.1.168 to v0.1.169 in merge commit `d74880bf2faf0815e83aceed1cb973418ee63d61`.
- Preserved the dirty worktree through backup branch `backup/pre-upstream-sync-20260731-220033` and full stash object `21daeede04acf323667e0583297a8c6be280112a`, then restored all tracked and untracked customizations without deleting the stash.
- Resolved four content conflicts by retaining upstream request-path security, compact-home, OAuth instruction, and initialization behavior together with the fork's Gemini-native, media gateway, and AI/Agent home features.
- Regenerated the complete Wire dependency graph and rebuilt the embedded frontend assets.

### Testing

- `go generate ./cmd/server` from `backend/`: passed.
- `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm install --frozen-lockfile` from `frontend/`: passed.
- `pnpm test:run --reporter=dot` from `frontend/`: passed, 205 test files and 1430 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; existing localStorage, Vue test-environment, Browserslist, and large-chunk warnings remained non-fatal.
- `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q`: passed.
- `bash deploy/test-caddyfile-cache.sh`: passed.
- `git diff --check`, `git diff --cached --check`, unresolved-conflict check, and `git merge-base --is-ancestor upstream/main HEAD`: passed.

### Notes

- Upstream merge inventory: `git show --name-only --format= d74880bf2faf0815e83aceed1cb973418ee63d61` lists all 132 upstream-changed files; the merge brings in v0.1.169 backend, frontend, deployment, security, billing, and test changes without a new upstream database migration.
- `backend/internal/server/routes/gateway.go`: combines the upstream Responses subpath guard with local video content, audio, Nano Banana, and public asset routing.
- `backend/internal/service/gemini_messages_compat_service.go`: retains OpenAI APIKey Gemini-native Bearer forwarding while applying upstream model/action and GET-path safety validation.
- `backend/internal/service/openai_oauth_passthrough_test.go`: accepts the upstream default-instructions behavior for Codex OAuth passthrough.
- `frontend/src/views/HomeView.vue`: keeps the customized AI/Agent default home, adds upstream compact-home selection, and restores auth/settings initialization.
- `frontend/src/views/__tests__/HomeView.compact.spec.ts`: validates compact/custom/default precedence against the customized home and isolates the Three.js component in jsdom.
- `backend/cmd/server/wire_gen.go`: was regenerated against the combined upstream and local dependency graph.
- `backend/internal/web/dist/`: was regenerated by the successful production frontend build.
- `docs/UPSTREAM_SYNC.md`: records synchronization scope, compatibility decisions, verification evidence, and recovery points.
- `progress.md`: records this task, tests, changed-file inventory, and rollback instructions.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 d74880bf2faf0815e83aceed1cb973418ee63d61`. The pre-sync branch is `backup/pre-upstream-sync-20260731-220033`, and the full pre-sync worktree stash object is `21daeede04acf323667e0583297a8c6be280112a`.

## 2026-07-31 - Task: Publish v0.1.169 as a multi-architecture image

### What was done

- Built the complete current customized worktree as `iotwq/china-api:latest` with application version `0.1.169` and source baseline commit `d74880bf2`.
- Published native `linux/amd64` and `linux/arm64` images to Docker Hub under one multi-architecture OCI index.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, and unresolved-conflict checks: passed.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.169 --build-arg COMMIT=d74880bf2 --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with manifest digest `sha256:15eef76991be7f64add85e58f6ca110c472a3c2a3bd091e26d367dff02adffd4`, amd64 digest `sha256:f274e2421eb05b4726a228223ac2795061bcf1e73bc97ba7cf6177e2e4c333e8`, and arm64 digest `sha256:5866598b95dd0146d918d985d087e02e564676f374e3b69c1837f1cf69caa715`.
- `docker run --rm --pull always --platform linux/amd64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.169 (commit: d74880bf2)`.
- `docker run --rm --pull always --platform linux/arm64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.169 (commit: d74880bf2)`.

### Notes

- `progress.md`: records this Docker Hub publication, remote architecture digests, runtime verification, and rollback point; no business source files were changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:a592a6e87d60f64f926f0104291a9e5b5a3a3c8407f217c0bd535c104a5cda3b`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-08-02 - Task: Synchronize upstream/main v0.1.170

### What was done

- Fetched and merged 36 upstream commits in `2980ff385..7e2e9ba05`, advancing the project from v0.1.169 to v0.1.170 in merge commit `6971356ba35c7445b3a0ed9182bf3a86a3b9eee0`.
- Preserved the complete customized worktree through backup branch `backup/pre-upstream-sync-20260802-212128` and stash object `3f60a9a008464a9d5ee4df2cd87353f11ee2e672`, then restored all tracked and untracked customizations without deleting the stash.
- Resolved README, OpenAI handler test, and account-management page conflicts while retaining the fork's media gateway, channel-monitor, Gemini-native, Grok custom-upstream, community-chat, and customized frontend behavior.
- Adapted the fork's audio, video, and Nano Banana handlers to the upstream three-state account-slot result so profit-control vetoes reselect an eligible account instead of returning an empty response.
- Regenerated the Wire dependency graph and embedded frontend assets.

### Testing

- `go test -tags=unit ./internal/handler -run 'TestOpenAIResponses_(APIKeyPassthroughSSERateLimitUsesConfiguredPoolRetry|ChannelMonitorPoolRetriesCountAsOneOfFiveAccounts|ChannelMonitorBadRequestContinuesToHealthyAccount)$' -count=1`: passed.
- `go generate ./cmd/server` and `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm install --frozen-lockfile` and `pnpm test:run --reporter=dot` from `frontend/`: passed, 209 test files and 1465 tests.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; existing localStorage, Vue test-environment, Browserslist, and large-chunk warnings remained non-fatal.
- `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q`: passed.
- `bash deploy/test-caddyfile-cache.sh`: passed.
- `git diff --check`, `git diff --cached --check`, unresolved-conflict check, and `git merge-base --is-ancestor upstream/main HEAD`: passed.

### Notes

- Upstream merge inventory: `git show --name-only --format= 6971356ba35c7445b3a0ed9182bf3a86a3b9eee0` lists the 190 upstream-changed files, including group profit control, expanded billing probes, moderation, gateway, admin UI, and test changes.
- `backend/migrations/192_group_profit_control.sql` and `backend/migrations/193_group_profit_control_auth_cache_invalidation.sql`: add group profit-control storage and auth-cache invalidation; no live database was accessed during this task.
- `backend/internal/handler/openai_audio.go`: handles profit-control slot vetoes by excluding and reselecting accounts.
- `backend/internal/handler/openai_nano_banana.go` and `backend/internal/handler/openai_videos.go`: adapt custom media handlers to the upstream slot-result contract with defensive account reselection.
- `backend/internal/handler/openai_gateway_handler_test.go`: combines upstream SSE 429 pool-retry coverage with local five-account and monitor HTTP 400 failover coverage.
- `frontend/src/views/admin/AccountsView.vue`: adds upstream filtered-result account selection while retaining async modal loading.
- `README_CN.md` and `README_JA.md`: accept the upstream sponsor-list update while preserving non-conflicting local documentation.
- `backend/cmd/server/wire_gen.go` and `backend/internal/web/dist/`: were regenerated against the combined upstream and local code.
- `docs/UPSTREAM_SYNC.md`: records the synchronization scope, compatibility decisions, verification evidence, migration boundary, and recovery points.
- `progress.md`: records this task, testing evidence, changed-file inventory, and rollback instructions.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 6971356ba35c7445b3a0ed9182bf3a86a3b9eee0`. The pre-sync branch is `backup/pre-upstream-sync-20260802-212128`, and the full pre-sync worktree stash object is `3f60a9a008464a9d5ee4df2cd87353f11ee2e672`.

## 2026-08-02 - Task: Reorganize the public image and video API documentation

### What was done

- Reordered the public API documentation so OpenAI appears first, Gemini Nano Banana second, international Jimeng SD2.0 third, and Grok fourth.
- Removed all public Sora 2 documentation while retaining `gpt-image-2`, and aligned its generation and multipart edit examples with the current image playground behavior, including 1K/2K/4K sizes, up to 16 reference images, masks, output formats, compression, moderation, and quality.
- Renamed the Gemini family to Nano Banana, added `gemini-3.1-flash-image-preview` as Nano Banana 2, and documented its shared Gemini-native request contract with `gemini-3-pro-image-preview`.
- Aligned the SD2.0 model table, request protocols, reference-media limits, fixed 720p behavior, string-valued duration, polling, downloads, and client-managed fast-model fallback with the current image playground implementation.
- Corrected the Gemini curl examples so their copied shell continuation lines remain executable.

### Testing

- `pnpm test:run src/views/user/__tests__/ApiDocsView.spec.ts --reporter=dot` from `frontend/`: passed, including section order, model coverage, Sora exclusion, media limits, and copied Gemini curl continuation syntax.
- `pnpm lint:check` from `frontend/`: passed.
- `pnpm typecheck` from `frontend/`: passed.
- `pnpm build` from `frontend/`: passed; the existing Browserslist age and large-chunk warnings remained non-fatal.
- `git diff --check`, `git diff --cached --check`, untracked-file whitespace checks, and scoped Sora-content checks: passed.

### Notes

- `frontend/src/views/user/ApiDocsView.vue`: reorganizes the four public model families and updates their user-facing endpoints, parameters, examples, and operational notes.
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: verifies family ordering, documented model and endpoint coverage, Sora removal, playground-aligned constraints, and copyable Gemini shell syntax.
- `docs/API_DOCS.md`: records the public documentation contract and the image-playground compatibility details for future maintenance.
- `progress.md`: records this documentation task, verification evidence, changed files, and rollback point.
- Rollback point: restore the three API documentation files from the preserved pre-sync stash object `3f60a9a008464a9d5ee4df2cd87353f11ee2e672^3` with `git show <object>:<path>` and overwrite only the matching path; preserve all other dirty-worktree changes.

## 2026-08-02 - Task: Publish v0.1.170 as a multi-architecture image

### What was done

- Built the complete current customized worktree as `iotwq/china-api:latest` with application version `0.1.170` and source baseline commit `6971356ba`.
- Published native `linux/amd64` and `linux/arm64` images to Docker Hub under one multi-architecture OCI index.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, and unresolved-conflict checks: passed.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.170 --build-arg COMMIT=6971356ba --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with OCI index digest `sha256:65e83d492d388fe95860e0952387b90e42537d9948718567875eb5612f12eb67`, amd64 digest `sha256:6fee71605a2295216039b99e9f890530b7979bc3fd1dedb145a58887cc0c51ac`, and arm64 digest `sha256:b2b497b51c3bf831195435d51f56181c04d51143ee943eaaf1dbfc9950b35aa3`.
- `docker run --rm --pull always --platform linux/amd64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.170 (commit: 6971356ba)`.
- `docker run --rm --pull always --platform linux/arm64 iotwq/china-api:latest --version`: passed and reported `Sub2API 0.1.170 (commit: 6971356ba)`.

### Notes

- `progress.md`: records this Docker Hub publication, remote architecture digests, runtime verification, and rollback point; no business source files were changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:15eef76991be7f64add85e58f6ca110c472a3c2a3bd091e26d367dff02adffd4`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-08-06 - Task: Synchronize upstream/main v0.1.171

### What was done

- Fetched and merged 60 upstream commits in `7e2e9ba05..a19c9f8d8`, advancing the project from v0.1.170 to v0.1.171 in merge commit `558f2a2efd131fb9eff8d4ac68053bb79f89a64a`.
- Preserved the complete customized worktree through backup branch `backup/pre-upstream-sync-20260806-234853` and stash object `4f5ec6d2cd490c96bc6f15368419f433ef3bc137`, then restored and verified all 219 tracked customization paths and 49 untracked files without deleting the stash.
- Integrated upstream captcha gates, refresh-token race protection, subscription renewal serialization, scheduler cancellation, Codex identity/version synchronization, billing precision, refund safeguards, and OpenAI failover fixes.
- Resolved gateway and authentication-page conflicts while retaining the fork's channel-monitor retry behavior, media gateway, Gemini-native support, Grok custom upstreams, community chat, image workspace, and ChinaAPI branding fallback.
- Regenerated the Wire dependency graph and embedded frontend assets. No database migration was added or applied.

### Testing

- `go generate ./cmd/server` and `go test -tags=unit ./... -count=1` from `backend/`: passed.
- `pnpm install --frozen-lockfile`, `pnpm test:run --reporter=dot`, `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; Vitest reported 215 test files and 1522 tests.
- `go test -tags=unit ./cmd/server ./internal/web -count=1` after the frontend production build: passed.
- `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q`: passed.
- `bash deploy/test-caddyfile-cache.sh`: passed.
- Stash restoration inventory comparison, `git diff --check`, `git diff --cached --check`, unresolved-conflict checks, and `git merge-base --is-ancestor upstream/main HEAD`: passed.
- `go mod tidy -diff` only proposed removing the two `google/subcommands` checksums required by the reproducible Wire generation command, so they were retained.

### Notes

- Upstream merge inventory: `git show --name-only --format= 558f2a2efd131fb9eff8d4ac68053bb79f89a64a` lists the 221 upstream-changed files covering authentication, concurrency, gateway, billing, payments, settings, and frontend captcha flows.
- `backend/internal/service/openai_gateway_service.go`: retains the local stream-disconnect drain grace while adopting the upstream Codex identity version and synchronization behavior.
- `backend/internal/service/openai_gateway_cc_pipeline.go`: combines strict wrapped-status/channel-monitor failover with upstream temporary-unschedulable account policy handling.
- `backend/internal/handler/setting_handler_public_test.go`: verifies both upstream Tencent captcha settings and the local image workspace URL contract.
- `frontend/src/views/auth/EmailVerifyView.vue` and `frontend/src/views/auth/RegisterView.vue`: integrate Tencent/Aliyun captcha configuration while retaining the ChinaAPI fallback name.
- `backend/cmd/server/wire_gen.go`: was regenerated with upstream captcha providers and the fork's media, asset, and community-chat dependencies.
- `backend/go.mod`, `backend/go.sum`, `frontend/package.json`, and `frontend/pnpm-lock.yaml`: contain the merged upstream captcha and runtime dependencies while retaining existing fork dependencies.
- `docs/UPSTREAM_SYNC.md`: records synchronization scope, risk-approved authentication/concurrency changes, compatibility decisions, verification, and recovery points.
- `progress.md`: records this synchronization task, testing evidence, changed-file inventory, and rollback instructions.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 558f2a2efd131fb9eff8d4ac68053bb79f89a64a`. The pre-sync branch is `backup/pre-upstream-sync-20260806-234853`, and the full pre-sync worktree stash object is `4f5ec6d2cd490c96bc6f15368419f433ef3bc137`.

## 2026-08-07 - Task: Add MiniMax H3 V2 video forwarding to OpenAI APIKey accounts

### What was done

- Added an explicit MiniMax H3 video endpoint capability for OpenAI APIKey accounts, including create/edit form persistence, capability-aware scheduling, and fixed `MiniMax-H3` model synchronization.
- Added the complete MiniMax H3 V2 asynchronous video flow through the existing OpenAI-compatible video gateway: create, query, list, delete, Context-IR, regeneration, and generated-content download.
- Preserved native MiniMax `content[]` requests and added compatibility conversion for text-only `prompt`, `seconds`, and `aspect_ratio`; media roles remain explicit and are never inferred.
- Added account binding for asynchronous tasks, safe unsigned CDN downloads, model-priced billing for task-creating operations, and idempotent balance refunds for failed or cancelled tasks.
- Changed MiniMax-capable account connection tests to use the read-only task-list endpoint, avoiding a billable video generation during routine administrator tests.

### Testing

- `go test ./internal/service ./internal/handler ./internal/server/routes`: passed.
- Targeted MiniMax capability, model-sync, forwarding, account-test, and six-route tests passed with both normal and `unit` build tags.
- `go test -tags=unit ./... -count=1` from `backend/`: passed, including the 146.592-second `internal/service` suite.
- The Create/Edit account modal Vitest files passed with 60 tests; `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/` passed. Existing Browserslist-age and large-chunk warnings remained non-fatal.
- With the user-authorized temporary credential, one `MiniMax-H3` 768P/4-second task returned HTTP 200, progressed through `running` to `succeeded`, appeared in the task list, and exposed an unsigned CDN URL that returned HTTP 206 with `video/mp4` for a Range request. The created task was then deleted successfully with HTTP 200. The temporary credential was not written to repository files or logs.
- Context-IR and regeneration request paths were verified with mocked protocol tests rather than additional live, billable task creation.
- `git diff --check`, `git diff --cached --check`, and scoped untracked-file whitespace checks passed before this log entry.

### Notes

- `backend/internal/service/account.go`: defines the opt-in MiniMax H3 endpoint capability for OpenAI APIKey accounts.
- `backend/internal/service/account_test_service.go`: uses a read-only MiniMax task-list request for account connectivity tests.
- `backend/internal/service/account_test_service_openai_test.go`: verifies the MiniMax account test URL, method, authentication, and SSE completion.
- `backend/internal/service/upstream_models.go`: returns the fixed official `MiniMax-H3` model for capable accounts without requiring `/v1/models` support.
- `backend/internal/service/upstream_models_test.go`: verifies fixed MiniMax model synchronization performs no upstream models request.
- `backend/internal/service/openai_gateway_service.go`: carries terminal asynchronous task failure state to the video handler.
- `backend/internal/service/openai_videos.go`: integrates MiniMax tasks with the existing video model, scheduling, billing, binding, and refund flow.
- `backend/internal/service/openai_minimax_video.go`: implements MiniMax V2 endpoint mapping, request compatibility, response URL rewriting, status handling, and unsigned content download.
- `backend/internal/service/openai_videos_test.go`: covers native and compatibility requests, all six upstream operations, capability gating, URL rewriting, deletion failure state, and content download authentication isolation.
- `backend/internal/handler/openai_videos.go`: exposes MiniMax operations through the authenticated OpenAI video gateway and applies concurrency, moderation, billing, binding, and refund handling.
- `backend/internal/server/routes/gateway.go`: registers versioned and unversioned list, delete, Context-IR, and regeneration routes.
- `backend/internal/server/routes/gateway_test.go`: verifies all MiniMax-compatible public route aliases are registered for OpenAI groups.
- `backend/internal/server/routes/prompt_audit_route_coverage_test.go`: records the two new task-creating routes in prompt-audit coverage.
- `frontend/src/types/index.ts`: adds `minimax_video` to the OpenAI endpoint capability type.
- `frontend/src/components/account/CreateAccountModal.vue`: allows administrators to enable MiniMax H3 video when creating an OpenAI APIKey account.
- `frontend/src/components/account/EditAccountModal.vue`: persists and restores MiniMax H3 video capability for existing accounts.
- `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts`: verifies new-account capability submission.
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`: verifies existing-account capability submission.
- `frontend/src/i18n/locales/en/admin/accounts.ts`: adds the English MiniMax H3 capability label and explanation.
- `frontend/src/i18n/locales/zh/admin/accounts.ts`: adds the Chinese MiniMax H3 capability label and explanation.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents account setup, six public-to-upstream route mappings, request formats, billing, task-list scope, downloads, and dedicated-account scheduling guidance.
- `progress.md`: records this implementation, verification evidence, changed files, and rollback guidance.
- Rollback: preserve the dirty worktree first; restore the tracked files listed above from the current index with `git restore --worktree -- <paths>`, remove `backend/internal/service/openai_minimax_video.go`, and reverse only the MiniMax-specific symbols/routes/tests/documentation in the pre-existing untracked OpenAI video files. Do not remove those shared video files because they also contain the earlier SD2.0 implementation.

## 2026-08-07 - Task: Publish v0.1.171 as a multi-architecture image

### What was done

- Built the complete current customized worktree as `iotwq/china-api:latest` with application version `0.1.171` and source baseline commit `558f2a2ef`.
- Published native `linux/amd64` and `linux/arm64` images to Docker Hub under one multi-architecture OCI index.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, and unresolved-conflict checks passed.
- The first build attempt stopped before compilation on a transient Docker Hub Dockerfile-frontend metadata `EOF`; the unchanged retry completed successfully.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.171 --build-arg COMMIT=558f2a2ef --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: passed with OCI index digest `sha256:8b3c122869ab9b6edfff51163d810183c424934feb47bf80dcfb022ed5ee3c03`, amd64 digest `sha256:86834c1446089f79f5f46e2d9ee1aba8684b279be3e964ebe33c48f37343de50`, and arm64 digest `sha256:51f8ba2fb034579610e625087a7b86be944bb47d628d58ae374922ba10b1ff5c`.
- Fresh remote pulls and `--version` runs passed for both platforms; each reported `Sub2API 0.1.171 (commit: 558f2a2ef)`.

### Notes

- `progress.md`: records this Docker Hub publication, remote architecture digests, runtime verification, and rollback point; no business source files were changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:65e83d492d388fe95860e0952387b90e42537d9948718567875eb5612f12eb67`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-08-07 - Task: Close residual-balance billing loop

### What was done

- Restored full post-response balance settlement when the exact request cost exceeds the user's remaining positive balance. The incurred request is now committed atomically, the balance becomes negative, and the existing exhausted-balance cache invalidation blocks subsequent requests.
- Preserved full `actual_cost` accounting for successfully settled requests; only genuine database settlement failures remain recorded as unsettled usage with `actual_cost = 0`.
- Added regression coverage for the observed residual balance and cost values, plus concurrent finalization of multiple already in-flight requests. No historical usage was back-charged and no existing user balance was changed.

### Testing

- The new residual-balance regression test failed before the implementation because the repository returned `ErrInsufficientBalance` and attempted to roll back instead of performing the overdraft update.
- `go test -tags=unit ./internal/repository ./internal/service -run 'Test(DeductUsageBillingBalance|ApplyUsageBillingEffects|SyncBalanceCacheAfterDeduction|CheckBillingEligibility)' -count=1`: passed.
- `TESTCONTAINERS_RYUK_DISABLED=true SUB2API_TEST_POSTGRES_IMAGE=postgres:16-alpine go test -tags=integration ./internal/repository -run 'TestUsageBillingRepositoryApply_(ConcurrentResidualBalanceChargesEveryRequest|DeduplicatesBalanceBilling)$' -count=1`: passed against real PostgreSQL; two concurrent `$0.00118608` settlements from a `$0.00032025` balance produced `-$0.00205191` exactly. A temporary local `redis:8.4-alpine` alias used by the test harness was removed afterward.
- `go test -tags=unit ./internal/repository ./internal/service -count=1`: passed; repository completed in 2.593 seconds and service completed in 145.975 seconds.

### Notes

- `backend/internal/repository/usage_billing_repo.go`: commits the complete incurred balance cost after the sufficient-balance guard misses instead of returning an error that rolls back the transaction.
- `backend/internal/repository/usage_billing_repo_unit_test.go`: verifies residual-balance full charging, overdraft signaling, and missing-user behavior.
- `backend/internal/repository/usage_billing_repo_integration_test.go`: verifies concurrent completed requests serialize and all settle against the same residual balance.
- `docs/PAYMENT_CN.md`: documents exact post-response settlement, negative-balance blocking, `actual_cost` semantics, concurrency, idempotency, and the no-backfill boundary.
- `progress.md`: records this implementation, verification evidence, changed-file inventory, and rollback point.
- Rollback: preserve the dirty worktree, then run `git restore --worktree -- backend/internal/repository/usage_billing_repo.go backend/internal/repository/usage_billing_repo_unit_test.go backend/internal/repository/usage_billing_repo_integration_test.go docs/PAYMENT_CN.md`; retain this progress entry as the audit trail. No database rollback is required because this task added no migration and changed no stored data.

## 2026-08-07 - Task: Harden all active model billing and video refunds

### What was done

- Closed unpriced-request paths for OpenAI and generic token gateways, aligned requested/channel/account model mapping before the price gate, and preserved billable partial-stream usage.
- Routed text, embedding, audio, image, Nano Banana, Gemini image, video, WebSocket turn, and Live usage through mandatory settlement behavior; queue saturation now executes the billing task synchronously.
- Made balance/subscription/API key/account quota effects, the visible usage row, and the billing ledger entry one PostgreSQL transaction. Settlement failure no longer writes a misleading zero-cost usage row.
- Made image and video balance-reservation capture atomic with final usage, charged Gemini images by the actual deduplicated output count, and corrected SD2.0, MiniMax, Sora-compatible, and Grok video billing parameters.
- Extended failed/cancelled video refunds to restore balance or subscription usage, API key quota and active rate-limit windows, account quota windows, and exhausted API key status. Refund rows are negative, visible, and idempotent.
- Added explicit per-request Live pricing and synchronous settlement before returning SDP. A post-create billing failure now closes the local Live call and releases its concurrency lease.
- No historical usage was back-charged, no stored balance was edited, and no database migration was added.

### Testing

- `go test ./internal/service ./internal/handler ./internal/repository -run '^$'`: passed compilation.
- Targeted service and handler billing tests covering residual balances, missing pricing, mapped pricing, media estimates, Gemini image counts, audio, embeddings, queue fallback, Live settlement, and video refunds: passed.
- `go test -tags=unit ./internal/repository ./internal/service -run 'TestUsageBillingRepositoryApplyWithUsageLog|TestApplyUsageBillingEffects|TestDeductUsageBillingBalance|TestCompositeBillableModel|TestBillableModelWithFallback|TestHasResolvableTokenPricing' -count=1`: passed.
- `go test ./internal/service -count=1`, `go test ./internal/handler -count=1`, `go test ./internal/repository -count=1`, and `go test ./internal/server/... -count=1`: passed after updating obsolete zero-cost Live and unsettled-usage assertions to the atomic-settlement contract.
- `TESTCONTAINERS_RYUK_DISABLED=true SUB2API_TEST_POSTGRES_IMAGE=postgres:16-alpine go test -tags=integration ./internal/repository -run 'TestUsageBillingRepositoryApplyWithUsageLog' -count=1`: passed against real PostgreSQL. It verified charge/log/ledger atomic commit, rollback on usage-log failure, and refund restoration of all current quota windows.
- Docker Hub metadata returned `EOF` for the harness Redis image, so the already-local `redis:7-alpine` image was temporarily tagged as `redis:8.4-alpine`; the temporary tag was removed immediately after the passing integration test.
- `git diff --check`: passed before documentation and progress-log append; a final check is required after this entry.

### Notes

- `backend/internal/handler/gateway_handler.go`: rejects generic gateway requests when no billable model price resolves and preserves partial billable usage.
- `backend/internal/handler/gateway_handler_chat_completions.go`: applies the generic token price gate before chat-completions forwarding.
- `backend/internal/handler/gateway_handler_responses.go`: applies the generic token price gate before Responses forwarding.
- `backend/internal/handler/openai_chat_completions.go`: applies the OpenAI token price gate before chat-completions forwarding.
- `backend/internal/handler/openai_embeddings.go`: validates mapped embedding pricing before forwarding and submits mandatory usage settlement.
- `backend/internal/handler/openai_gateway_handler.go`: adds OpenAI Responses, Messages, and per-turn WebSocket pricing gates and mandatory usage settlement.
- `backend/internal/handler/openai_images.go`: settles image balance reservations atomically with final usage.
- `backend/internal/handler/openai_live.go`: checks per-request pricing and budget, bills synchronously, and aborts a created call on settlement failure.
- `backend/internal/handler/openai_nano_banana.go`: settles Nano Banana balance reservations atomically using actual returned image count.
- `backend/internal/handler/openai_nano_banana_billing_test.go`: verifies preflight rejection and atomic estimated-versus-actual image settlement.
- `backend/internal/handler/openai_videos.go`: preflights video cost, preserves mandatory usage settlement, and triggers terminal-task refunds.
- `backend/internal/handler/usage_record_submit_task_test.go`: verifies mandatory usage tasks execute without being dropped.
- `backend/internal/handler/usage_record_task_fallback_test.go`: verifies queue-full and stopped-pool synchronous billing fallback.
- `backend/internal/repository/usage_billing_repo.go`: implements atomic charge/log/ledger commit, full residual-balance settlement, and complete quota refunds.
- `backend/internal/repository/usage_billing_repo_unit_test.go`: covers atomic commit/rollback and residual-balance overdraft behavior.
- `backend/internal/repository/usage_billing_repo_integration_test.go`: covers real-PostgreSQL atomic settlement and refund restoration.
- `backend/internal/service/billing_service.go`: prevents unsupported non-Grok videos from silently using an invented zero/default price.
- `backend/internal/service/gateway_record_usage_test.go`: locks generic gateway billing-failure behavior to no misleading zero-cost row.
- `backend/internal/service/gateway_usage_billing.go`: builds refundable billing commands, performs atomic settlement, and fails closed on missing pricing.
- `backend/internal/service/gemini_chat_completions_compat_service.go`: preserves actual Gemini image counts in compatibility responses.
- `backend/internal/service/gemini_messages_compat_service.go`: deduplicates streaming Gemini images and carries actual counts into billing.
- `backend/internal/service/openai_embeddings.go`: estimates usage when an embeddings upstream omits the usage object.
- `backend/internal/service/openai_embeddings_test.go`: verifies embedding usage estimation.
- `backend/internal/service/openai_gateway_record_usage_test.go`: covers missing prices, mapped/composite price fallback, atomic media settlement, Live per-request billing, and full video refunds.
- `backend/internal/service/openai_gateway_service.go`: exposes the shared usage-billing repository needed by the hardened gateway flows.
- `backend/internal/service/openai_gateway_usage.go`: removes zero-cost missing-price fallback, handles per-request media/Live costs, and atomically captures media reservations.
- `backend/internal/service/openai_live.go`: moves Live usage out of finalization and adds idempotent abort cleanup.
- `backend/internal/service/openai_live_types.go`: retains the created Live record for immediate billing-failure cleanup.
- `backend/internal/service/openai_live_lifecycle_test.go`: verifies no duplicate Live usage row and one-time lease release.
- `backend/internal/service/openai_model_mapping.go`: shares the exact Messages upstream-model resolution between forwarding and pricing.
- `backend/internal/service/openai_model_mapping_test.go`: verifies channel mapping is resolved before account mapping for Messages pricing.
- `backend/internal/service/openai_videos.go`: calculates model-specific video billing, persists task bindings, and applies complete idempotent refunds.
- `backend/internal/service/openai_videos_test.go`: verifies video billing parameters and failure-refund behavior.
- `backend/internal/service/usage_billing.go`: extends atomic commands with refund-window time and media reservation capture.
- `backend/internal/service/usage_log_helpers.go`: recognizes request-count and video usage as billable partial results.
- `docs/BILLING_INTEGRITY.md`: documents the settlement guarantees, media and refund behavior, operational checks, and infrastructure boundary.
- `progress.md`: records this hardening task, verification evidence, changed-file inventory, and rollback guidance.
- Rollback: preserve the current dirty worktree first. For tracked files, restore the pre-task staged baseline with `git restore --worktree -- backend/internal/handler/gateway_handler.go backend/internal/handler/gateway_handler_chat_completions.go backend/internal/handler/gateway_handler_responses.go backend/internal/handler/openai_chat_completions.go backend/internal/handler/openai_embeddings.go backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_images.go backend/internal/handler/openai_live.go backend/internal/handler/openai_nano_banana.go backend/internal/handler/usage_record_submit_task_test.go backend/internal/handler/usage_record_task_fallback_test.go backend/internal/repository/usage_billing_repo.go backend/internal/repository/usage_billing_repo_unit_test.go backend/internal/repository/usage_billing_repo_integration_test.go backend/internal/service/billing_service.go backend/internal/service/gateway_record_usage_test.go backend/internal/service/gateway_usage_billing.go backend/internal/service/gemini_chat_completions_compat_service.go backend/internal/service/gemini_messages_compat_service.go backend/internal/service/openai_embeddings.go backend/internal/service/openai_embeddings_test.go backend/internal/service/openai_gateway_record_usage_test.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_usage.go backend/internal/service/openai_live.go backend/internal/service/openai_live_lifecycle_test.go backend/internal/service/openai_model_mapping.go backend/internal/service/openai_model_mapping_test.go backend/internal/service/usage_billing.go backend/internal/service/usage_log_helpers.go`; remove only `docs/BILLING_INTEGRITY.md`. The shared untracked media files `backend/internal/handler/openai_videos.go`, `backend/internal/service/openai_live_types.go`, `backend/internal/service/openai_videos.go`, and `backend/internal/service/openai_videos_test.go` predate this task and must not be deleted; reverse only the billing/refund hunks documented above if a partial rollback is required. No database rollback is required.

## 2026-08-07 - Task: Finalize billing integrity delivery records

### What was done

- Added the billing integrity guide to the repository documentation allowlist so the settlement, refund, monitoring, and infrastructure boundaries are retained in future commits.
- Completed the final repository consistency checks after the billing implementation and documentation changes. No additional runtime behavior was changed in this finalization step.

### Testing

- `git diff --check`: passed.
- `git diff --cached --check`: passed.
- `git diff --name-only --diff-filter=U`: passed with no unresolved conflicts.
- `git status --short --untracked-files=all -- docs/BILLING_INTEGRITY.md .gitignore progress.md`: confirmed the billing guide is visible as an untracked deliverable rather than ignored.

### Notes

- `.gitignore`: allows `docs/BILLING_INTEGRITY.md` to be tracked with the other maintained project documentation.
- `docs/BILLING_INTEGRITY.md`: documents the finalized billing guarantees, refund behavior, operational checks, and unavoidable infrastructure boundary.
- `progress.md`: appends this final documentation and verification audit record.
- Rollback: remove only the `!docs/BILLING_INTEGRITY.md` allowlist line and delete `docs/BILLING_INTEGRITY.md`; retain both progress entries as the append-only audit trail. No database rollback is required.

## 2026-08-07 - Task: Publish billing-hardened v0.1.171 multi-architecture image

### What was done

- Built the complete current customized worktree, including the finalized billing and video-refund fixes, as `iotwq/china-api:latest` with application version `0.1.171` and source baseline commit `558f2a2ef`.
- Published native `linux/amd64` and `linux/arm64` images to Docker Hub under one OCI multi-architecture index.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, and unresolved-conflict checks passed.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.171 --build-arg COMMIT=558f2a2ef --tag iotwq/china-api:latest --push .`: passed.
- Docker Registry API verification over IPv4 confirmed remote OCI index digest `sha256:54e5ad087bd1f184fce229947232b17068dec92a9a1178be4e53c7dde04290fe`, amd64 digest `sha256:498584beff55bb2e223377d9811cc167452ae7225f9526589331b50cc4b14e4a`, and arm64 digest `sha256:aa937944b71a3c7fdeba24bcc0a6302247134e172e1ebe7597679587ccc0e3cb`.
- The arm64 child image was loaded from the exact cached remote digest and executed successfully; `--version` reported `Sub2API 0.1.171 (commit: 558f2a2ef)`. The temporary verification tag was removed afterward.
- Docker Desktop's direct post-push reads repeatedly failed at Docker Hub's IPv6 endpoint with connection reset/EOF, so a fresh daemon pull and amd64 runtime execution were not completed. This did not affect the successful push or independent IPv4 registry verification of both remote child manifests.

### Notes

- `progress.md`: records this Docker Hub publication, architecture digests, runtime evidence, network limitation, and rollback point; no business source file was changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:8b3c122869ab9b6edfff51163d810183c424934feb47bf80dcfb022ed5ee3c03`; preserve the current dirty worktree and do not use a wholesale reset.

## 2026-08-08 - Task: Synchronize upstream/main v0.1.172

### What was done

- Fetched and merged 43 upstream commits in `a19c9f8d8..68d8f122e`, advancing the project from v0.1.171 to v0.1.172 in merge commit `d78bcd679c7778abe61af04c60f07ef7da808453`.
- Preserved the complete customized worktree through backup branch `backup/pre-upstream-sync-20260808-004445` and stash object `5d1ec7a216f72a840e93f800c94b57858f6654c9`, restored every tracked and untracked path, and retained the stash as the exact recovery point.
- Integrated the user-approved authentication, concurrency, networking, and database changes, including OAuth adoption protection, TCP dial timeout, pre-output streaming failover, daily subscription reset, response-model auditing, and migrations 194/195.
- Resolved three content conflicts while preserving Gemini actual-image billing, video-refund display, and all existing fork features. Repaired two merge-combination build/test fixture gaps without weakening pricing enforcement.
- Regenerated Wire dependencies and the embedded production frontend. No running database or deployed service was changed.

### Testing

- `go generate ./cmd/server` and `go test -tags=unit ./... -count=1` from `backend/`: passed.
- Targeted Gemini, usage billing, OAuth, captcha, TCP timeout, streaming/failover, response-model audit, and Responses schema tests: passed.
- `pnpm install --frozen-lockfile`, `pnpm test:run --reporter=dot`, `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed; Vitest reported 217 test files and 1542 tests.
- `go test -tags=unit ./internal/web ./cmd/server -count=1` after the frontend production build: passed.
- `TESTCONTAINERS_RYUK_DISABLED=true SUB2API_TEST_POSTGRES_IMAGE=postgres:16-alpine go test -tags=integration ./internal/repository -run 'TestMigrationsRunner_(ConcurrentInstancesSerializeOnSessionLock|IsIdempotent_AndSchemaIsUpToDate|AuthIdentityAndPaymentSchemaStayAligned)$' -count=1`: passed against temporary PostgreSQL and Redis containers; the temporary local Redis compatibility tag was removed.
- Root, deploy, and standalone Compose files passed `docker compose config --quiet` with non-persistent validation-only values for required environment variables.
- Stash inventory comparison confirmed no missing or extra restored paths. `git diff --check`, `git diff --cached --check`, strict conflict-marker scanning, generated-file checks, and `git merge-base --is-ancestor upstream/main HEAD` passed.
- `go mod tidy -diff` only proposed removing the two `google/subcommands` checksums required by reproducible Wire generation, so they were retained.

### Notes

- Upstream merge inventory: `git show --name-only --format= d78bcd679c7778abe61af04c60f07ef7da808453` lists the 192 upstream-changed files across authentication, concurrency, networking, database schema, gateway behavior, usage auditing, and frontend settings.
- `backend/internal/service/gemini_messages_compat_service.go`: combines upstream response-model observation with the fork's deduplicated actual Gemini image counting and billing metadata.
- `frontend/src/components/admin/usage/UsageTable.vue`: combines upstream requested/sent/response model auditing with the fork's visible failed-video refund rows.
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`: covers both model mismatch/variant markers and failed-video refund display.
- `backend/internal/service/openai_gateway_usage.go`: restores the `zap` dependency required by the merged upstream response-model conflict warning after local billing hardening was reapplied.
- `backend/internal/handler/openai_responses_failover_cancel_test.go`: supplies the real minimal billing service now required before Responses account failover, preserving both cancellation and connected-client assertions.
- `backend/migrations/194_add_usage_log_upstream_response_model.sql`: adds the upstream response-model and mismatch audit fields.
- `backend/migrations/195_add_usage_log_upstream_model_mismatch_index_notx.sql`: adds the idempotent concurrent partial index used by mismatch queries.
- `backend/cmd/server/wire_gen.go` and `backend/internal/web/embed_on.go`: are regenerated from the merged dependency graph and production frontend assets.
- `docs/UPSTREAM_SYNC.md`: records synchronization scope, conflict decisions, database impact, verification evidence, and recovery points.
- `progress.md`: appends this task audit record; existing history is unchanged.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 d78bcd679c7778abe61af04c60f07ef7da808453`. The pre-sync branch is `backup/pre-upstream-sync-20260808-004445`, and the full pre-sync worktree stash object is `5d1ec7a216f72a840e93f800c94b57858f6654c9`. Do not drop that stash until the deployment has been accepted.

## 2026-08-08 - Task: Return an actionable error for unpriced models

### What was done

- Changed pricing preflight failures caused by missing model prices from `503 billing_service_error` to `400 invalid_request_error`, including the requested model name and an instruction to contact the administrator.
- Applied the response consistently to text, Responses, Messages, Chat Completions, embeddings, images, Nano Banana, video, audio, Live, and WebSocket turn pricing checks. Real billing infrastructure and post-response settlement failures remain `503 billing_service_error`.
- Added a MiniMax-H3 video regression proving that an unpriced request is rejected before any upstream request, balance reservation, or settlement occurs.

### Testing

- `go test -tags=unit ./internal/handler -run 'TestPricingPreflightErrorDetails|TestVideosRejectsUnpricedMiniMaxH3BeforeUpstream' -count=1 -v`: passed.
- `go test -tags=unit ./internal/handler ./internal/service -count=1`: passed.
- `go test -tags=unit ./... -count=1`: passed for all backend unit packages.
- `rg -n -C 3 'billing_service_error|Model pricing is not configured|pricing is not configured' backend/internal/handler`: confirmed remaining 503 responses are limited to billing infrastructure and post-create Live settlement failures.
- `git diff --check`: passed before the progress entry was appended.

### Notes

- `backend/internal/handler/gateway_handler.go`: adds the shared unpriced-model message and pricing preflight error classification, and uses it for generic Messages requests.
- `backend/internal/handler/gateway_handler_chat_completions.go`: returns the actionable unpriced-model error for generic Chat Completions.
- `backend/internal/handler/gateway_handler_responses.go`: returns the actionable unpriced-model error for generic Responses.
- `backend/internal/handler/openai_chat_completions.go`: returns the actionable unpriced-model error for OpenAI Chat Completions.
- `backend/internal/handler/openai_embeddings.go`: returns the actionable unpriced-model error for embeddings.
- `backend/internal/handler/openai_gateway_handler.go`: covers OpenAI Responses, Messages, and WebSocket turn pricing; unpriced WebSocket turns close as a policy violation rather than a retryable service outage.
- `backend/internal/handler/openai_images.go`: returns the actionable unpriced-model error for image cost estimation.
- `backend/internal/handler/openai_nano_banana.go`: returns the actionable unpriced-model error for Nano Banana cost estimation.
- `backend/internal/handler/openai_videos.go`: returns the actionable unpriced-model error for video creation, including MiniMax-H3.
- `backend/internal/handler/openai_audio.go`: returns the actionable unpriced-model error for audio requests.
- `backend/internal/handler/openai_live.go`: returns the actionable unpriced-model error for Live preflight while retaining 503 for a real post-create settlement failure.
- `backend/internal/handler/gateway_handler_billing_error_test.go`: verifies missing-price and infrastructure-failure classifications remain distinct.
- `backend/internal/handler/openai_videos_pricing_error_test.go`: verifies the MiniMax-H3 HTTP response and zero upstream/billing side effects.
- `docs/BILLING_INTEGRITY.md`: documents the new HTTP 400 contract and the retained HTTP 503 infrastructure contract.
- `progress.md`: records implementation, verification, file inventory, and rollback guidance.
- Rollback: preserve the dirty worktree first. Restore the tracked handler files to their staged pre-task state with `git restore --worktree -- backend/internal/handler/gateway_handler.go backend/internal/handler/gateway_handler_chat_completions.go backend/internal/handler/gateway_handler_responses.go backend/internal/handler/openai_chat_completions.go backend/internal/handler/openai_embeddings.go backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_images.go backend/internal/handler/openai_live.go backend/internal/handler/openai_nano_banana.go backend/internal/handler/gateway_handler_billing_error_test.go`; delete only `backend/internal/handler/openai_videos_pricing_error_test.go`; reverse only the `pricingPreflightErrorDetails` call sites in the pre-existing untracked `backend/internal/handler/openai_audio.go` and `backend/internal/handler/openai_videos.go`; restore the previous missing-pricing sentence in `docs/BILLING_INTEGRITY.md`. Retain this append-only progress entry. No database rollback is required.

## 2026-08-08 - Task: Publish current v0.1.172 multi-architecture image

### What was done

- Built the complete current customized worktree as `iotwq/china-api:latest`, including the actionable unpriced-model error response, with application version `0.1.172` and source baseline commit `d78bcd679`.
- Published native `linux/amd64` and `linux/arm64` images to Docker Hub under one OCI multi-architecture index.

### Testing

- Preflight `git diff --check`, `git diff --cached --check`, unresolved-conflict check, and strict conflict-marker scan: passed.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.172 --build-arg COMMIT=d78bcd679 --build-arg DATE=2026-08-07T18:01:17Z --tag iotwq/china-api:latest --push .`: passed.
- Remote Docker Hub inspection confirmed OCI index digest `sha256:f251c47dab0bbf2ac0f10f6e4a261072b4405520b959085ba6cc0672bf5f67f9`, amd64 digest `sha256:f3138f3969d85540d9cac5fb0715d0f58b42e31fe3bdb3d05b4d81c4cd4cf2bb`, and arm64 digest `sha256:d71921d808dc1cd89ebd52c1827c5582937efddc0a7695dbb6fc3759f881a0ad`.
- Pulled and executed both exact remote child digests. Each `--version` invocation reported `Sub2API 0.1.172 (commit: d78bcd679, built: 2026-08-07T18:01:17Z)`.

### Notes

- `progress.md`: records the Docker Hub publication, architecture digests, executable verification, and rollback point; no business source or configuration file was changed during publication.
- Rollback: restore the previous multi-architecture image with `docker buildx imagetools create -t iotwq/china-api:latest iotwq/china-api@sha256:54e5ad087bd1f184fce229947232b17068dec92a9a1178be4e53c7dde04290fe`. Preserve the current dirty worktree; no database rollback is required.

## 2026-08-08 - Task: Add MiniMax-H3 per-second channel pricing and verified media billing

### What was done

- Added a fixed `MiniMax-H3` channel pricing template with required `768P` and `2K` USD-per-second prices. Selecting the exact model in the channel editor creates the template automatically; the backend rejects incomplete, duplicate, mixed-model, or unsupported video pricing rules.
- Implemented the billing formula `（输出视频时长 + 服务端核验的参考视频实际总时长）× 当前分辨率每秒价格 × 有效视频倍率` for standard MiniMax-H3 video creation. The same verified input duration is used by preflight balance coverage and final usage settlement.
- Kept up to 9 reference images and 3 reference audios free, charged up to 3 reference videos by parsed MP4/MOV duration, and preserved existing token, image, generic video, and legacy per-request pricing behavior.
- Required paid reference videos to use this service's `/pg/assets` URL matching `server.frontend_url`, preventing client-supplied durations or external URLs from becoming untrusted billing inputs. Added actionable HTTP 400 responses for media validation failures.
- Updated the MiniMax-H3 operator documentation with template setup, exact formula, media constraints, trusted-asset requirement, and the unchanged legacy `per_request` behavior. No database migration or deployed-service change was required.

### Testing

- `go test ./internal/service ./internal/handler ./internal/handler/admin -run 'Test(EstimateMiniMaxH3VideoBilling|RecordUsageMiniMaxH3VideoBilling|ValidatePricingBillingMode|BillingModeIsValid|PricingPreflightErrorDetails|OpenAIVideo)' -count=1`: passed.
- `go test -tags=unit ./... -count=1`: passed for all backend unit packages.
- `go test -tags=unit ./internal/web ./cmd/server -count=1`: passed after the frontend production build.
- `pnpm vitest run src/components/admin/channel/__tests__/PricingEntryCard.spec.ts src/components/admin/channel/__tests__/types.spec.ts --reporter=dot`: passed 2 files and 8 tests.
- `pnpm test:run --reporter=dot`: passed 218 files and 1543 tests.
- `pnpm typecheck`, `pnpm lint:check`, and `pnpm build`: passed. The production build completed with only the existing Browserslist age and large-chunk warnings.
- `go mod tidy -diff`: only proposed removing legacy `go-test/deep v1.0.3` and `google/subcommands` checksums retained for the existing reproducible toolchain; the new `mp4ff` dependency is otherwise tidy.
- `git diff --check`, `git diff --cached --check`, unresolved-conflict inspection, and strict conflict-marker scanning of task files: passed.
- No real MiniMax-H3 paid generation request was made; duration parsing, cost calculation, final usage settlement, free image/audio behavior, count limits, and invalid external video rejection were verified with local fixtures and automated tests.

### Notes

- `backend/go.mod`: adds the MP4/MOV duration parser as a direct dependency.
- `backend/go.sum`: records the new parser and transitive module checksums while retaining pre-existing tool-generation checksums.
- `backend/internal/handler/admin/channel_handler.go`: accepts the new `video` channel billing mode in admin requests.
- `backend/internal/handler/gateway_handler.go`: maps verified video-input validation failures to actionable client errors without changing real billing infrastructure failures.
- `backend/internal/handler/gateway_handler_billing_error_test.go`: covers missing pricing, infrastructure failures, and MiniMax-H3 media validation error classification.
- `backend/internal/handler/openai_videos.go`: carries the verified reference-video duration from preflight into the final video usage settlement.
- `backend/internal/service/billing_service.go`: prevents generic unified billing from processing video mode without verified media duration.
- `backend/internal/service/channel.go`: makes `video` a valid tier-label billing mode.
- `backend/internal/service/channel_test.go`: verifies `video` is accepted as a valid billing mode.
- `backend/internal/service/channel_service.go`: validates the exact standalone MiniMax-H3 768P/2K template and its prices.
- `backend/internal/service/channel_service_test.go`: covers valid, wrong-model, and incomplete video template configurations.
- `backend/internal/service/model_pricing_resolver.go`: resolves video pricing tiers through the existing channel pricing cache.
- `backend/internal/service/openai_gateway_service.go`: carries server-verified reference-video seconds in the video forwarding result.
- `backend/internal/service/openai_gateway_usage.go`: estimates and settles MiniMax-H3 output plus reference-video seconds at the selected resolution price.
- `backend/internal/service/minimax_video_billing.go`: validates 9/3/3 media limits, resolves trusted local assets, and parses MP4/MOV durations.
- `backend/internal/service/minimax_video_billing_test.go`: verifies 768P/2K prices, multipliers, free image/audio inputs, paid reference-video duration, final settlement, and invalid media rejection.
- `frontend/src/constants/channel.ts`: adds the shared `video` billing mode type.
- `frontend/src/components/admin/channel/PricingEntryCard.vue`: creates and renders the fixed MiniMax-H3 per-second pricing template.
- `frontend/src/components/admin/channel/__tests__/PricingEntryCard.spec.ts`: verifies selecting MiniMax-H3 produces the expected fixed template.
- `frontend/src/components/admin/channel/__tests__/types.spec.ts`: verifies video tier validation retains label-based interval semantics.
- `frontend/src/views/admin/ChannelsView.vue`: blocks saving invalid MiniMax-H3 video templates before submission.
- `frontend/src/i18n/locales/zh/admin/channels.ts`: adds Chinese labels, formula, and media-limit guidance.
- `frontend/src/i18n/locales/en/admin/channels.ts`: adds the matching English pricing and validation text.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents configuration, billing formula, material rules, and operational requirements.
- `progress.md`: appends this implementation, verification, file inventory, and rollback record.
- Rollback: preserve the dirty worktree first. Restore the tracked task files to their pre-task index with `git restore --worktree -- backend/go.mod backend/go.sum backend/internal/handler/admin/channel_handler.go backend/internal/handler/gateway_handler.go backend/internal/handler/gateway_handler_billing_error_test.go backend/internal/service/billing_service.go backend/internal/service/channel.go backend/internal/service/channel_test.go backend/internal/service/channel_service.go backend/internal/service/channel_service_test.go backend/internal/service/model_pricing_resolver.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_usage.go frontend/src/constants/channel.ts frontend/src/components/admin/channel/PricingEntryCard.vue frontend/src/components/admin/channel/__tests__/types.spec.ts frontend/src/views/admin/ChannelsView.vue frontend/src/i18n/locales/zh/admin/channels.ts frontend/src/i18n/locales/en/admin/channels.ts`; delete only the task-created `backend/internal/service/minimax_video_billing.go`, `backend/internal/service/minimax_video_billing_test.go`, and `frontend/src/components/admin/channel/__tests__/PricingEntryCard.spec.ts`. The pre-existing untracked `backend/internal/handler/openai_videos.go` and `docs/OPENAI_MEDIA_COMPAT.md` must not be deleted; reverse only the verified-input-duration and MiniMax-H3 pricing-template sections described above. Retain this append-only progress entry. No database rollback is required.

## 2026-08-08 - Task: Add Firefly Video v2 forwarding and remove video-v1-15s

### What was done

- Added `firefly-video-v2` and `firefly-video-v2-fast` to the OpenAI APIKey video path. JSON text and storyboard requests now create tasks through `POST /v1/videos`; status, file metadata, Range playback, and downloads use the matching `/v1/videos/{task_id}` interfaces.
- Added Content-Type-aware Firefly multipart parsing and structured forwarding for real `first_frame`, `last_frame`, repeated `images`, `videos`, and `audios` file parts. Account model mapping rewrites only the multipart `model` field while retaining file names, part headers, repeated fields, and bytes.
- Normalized multipart duration and resolution before preflight and final video billing. Firefly requests omitted by the client use the documented 5-second, 720p defaults without changing MiniMax-H3, Jimeng SD2.0, Grok, image, or token billing rules.
- Added `HEAD /v1/videos/{task_id}/content` passthrough and preserved content metadata without writing a response body.
- Removed `video-v1-15s`, both public `/v1/video/generations` aliases, and all YCY domain/path special-casing. Updated the user API page and operator documentation with Firefly JSON, storyboard, multipart, polling, HEAD, download, and upload-limit guidance.

### Testing

- `go test -tags=unit ./... -count=1` from `backend/`: passed for all backend unit packages.
- Firefly/model/removal-focused service, handler, and route tests: passed, including multipart repeated-file byte preservation, multipart model mapping, billing parameters, HEAD behavior, and legacy route HTTP 404.
- `pnpm test:run -- frontend/src/views/user/__tests__/ApiDocsView.spec.ts --reporter=dot`: passed the complete frontend suite, 218 files and 1543 tests.
- `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts --reporter=dot`: passed the focused API documentation test.
- `pnpm typecheck`, `pnpm lint:check`, and `pnpm build`: passed. The production build reported only the existing Browserslist age and large-chunk warnings.
- `git diff --check`: passed before this append-only progress entry.
- No paid Firefly task was submitted to the real upstream; protocol behavior was verified with byte-level multipart fixtures and recorded HTTP responses.

### Notes

- `backend/internal/service/openai_videos.go`: adds Firefly model recognition, JSON/storyboard and multipart parsing, billing normalization, structured model rewriting, unified `/v1/videos` forwarding, and HEAD content handling while deleting the legacy YCY route branch.
- `backend/internal/service/openai_videos_test.go`: covers Firefly acceptance, storyboard prompts, old-model rejection, JSON forwarding, multipart preservation/model mapping, billing defaults, and HEAD responses.
- `backend/internal/service/openai_minimax_video.go`: makes the shared content endpoint honor inbound HEAD requests without downloading the MiniMax content body.
- `backend/internal/service/openai_gateway_record_usage_test.go`: replaces legacy `video-v1-15s` refund fixtures with the supported Firefly fast model.
- `backend/internal/handler/openai_videos.go`: parses requests using their Content-Type and supplies normalized multipart parameters to preflight billing.
- `backend/internal/handler/endpoint.go` and `backend/internal/handler/endpoint_test.go`: remove canonical normalization and upstream derivation for `/v1/video/generations`.
- `backend/internal/server/routes/gateway.go`, `backend/internal/server/routes/gateway_test.go`, and `backend/internal/server/routes/prompt_audit_route_coverage_test.go`: remove legacy routes, add authenticated HEAD routes, and verify route and audit coverage.
- `frontend/src/views/user/ApiDocsView.vue` and `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: remove the old model/API examples and add the Firefly user-facing interface section and regression coverage.
- `docs/API_DOCS.md`: records the public Firefly capabilities and removes `video-v1-15s` from the Jimeng family.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents OpenAI APIKey upstream setup, true multipart forwarding, and the configurable 256/384 MiB deployment boundary.
- `progress.md`: appends this implementation, verification, file inventory, and rollback point.
- Rollback: preserve the dirty worktree first. Restore the tracked task files to their pre-task staged state with `git restore --worktree -- backend/internal/handler/endpoint.go backend/internal/handler/endpoint_test.go backend/internal/server/routes/gateway.go backend/internal/server/routes/gateway_test.go backend/internal/server/routes/prompt_audit_route_coverage_test.go backend/internal/service/openai_gateway_record_usage_test.go`. The video service, video handler, API docs, and their tests were pre-existing untracked files and must not be deleted; reverse only the Firefly, multipart, HEAD, and legacy-route-removal changes listed above. Retain this append-only progress entry. No database rollback is required.

## 2026-08-08 - Task: Add generic video per-second channel pricing

### What was done

- Added a generic “video (per second)” channel pricing template with one administrator-defined USD-per-output-second price. `firefly-video-v2` and `firefly-video-v2-fast` select this template automatically and may share one pricing rule.
- Implemented `output duration × video count × per-second price × effective video multiplier` for generic video estimation and final usage settlement. Reference image, video, and audio inputs are deliberately excluded from generic video charges.
- Kept the existing MiniMax-H3 768P/2K template and its verified reference-video duration charge unchanged. MiniMax-H3 remains a standalone pricing rule and cannot be mixed into a generic video rule.
- Added backend validation for missing generic prices, unsupported generic tiers, and mixed MiniMax-H3 rules; updated the channel editor, bilingual guidance, operator documentation, and regression coverage. No database migration was required.

### Testing

- Focused backend pricing, settlement, validation, and MiniMax-H3 regression tests: passed. A 15-second Firefly request at `$0.17/second` and multiplier `0.5` produced `$2.55` original cost and `$1.275` actual cost; a nonzero reference-video duration did not change the generic video cost.
- `go test -tags=unit ./... -count=1`: passed for all backend unit packages.
- `pnpm test:run --reporter=dot`: passed 218 frontend test files and 1545 tests.
- `pnpm typecheck`, `pnpm lint:check`, and `pnpm build`: passed. The production build reported only the existing Browserslist age and large-chunk warnings.
- `go test -tags=unit ./internal/web ./cmd/server -count=1`: passed against the generated frontend production assets.
- `git diff --check`, `git diff --cached --check`, unresolved-conflict inspection, and strict conflict-marker scanning of task files: passed.

### Notes

- `backend/internal/service/channel.go`: documents the existing price field as the generic video output-per-second storage location.
- `backend/internal/service/channel_service.go`: separates generic video pricing validation from the MiniMax-H3 specialized template.
- `backend/internal/service/channel_service_test.go`: covers shared Firefly pricing, missing prices, unsupported tiers, and MiniMax-H3 mixing rejection.
- `backend/internal/service/model_pricing_resolver.go`: resolves the generic video default price through the existing channel-pricing structure.
- `backend/internal/service/openai_gateway_usage.go`: estimates and settles generic video output seconds while leaving MiniMax-H3 reference-video charging intact.
- `backend/internal/service/openai_video_per_second_billing_test.go`: verifies estimation, multipliers, output counts, free reference media, and final usage settlement.
- `frontend/src/components/admin/channel/PricingEntryCard.vue`: renders the generic one-price template and preserves the specialized MiniMax-H3 controls.
- `frontend/src/components/admin/channel/__tests__/PricingEntryCard.spec.ts`: verifies Firefly auto-selection and editing the output-per-second price.
- `frontend/src/views/admin/ChannelsView.vue`: validates generic and MiniMax-H3 video templates before saving.
- `frontend/src/i18n/locales/zh/admin/channels.ts`: adds Chinese generic video pricing labels, formula, and free-reference-media guidance.
- `frontend/src/i18n/locales/en/admin/channels.ts`: adds the matching English guidance.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents configuration, formula, shared-price rules, free reference media, and the MiniMax-H3 exception.
- `progress.md`: records implementation, verification, file inventory, and rollback guidance.
- Rollback: preserve the dirty worktree and reverse only the generic-video branches and copy described in this entry; do not run broad `git restore` commands because these files contain earlier uncommitted MiniMax-H3 and Firefly work. Delete only `backend/internal/service/openai_video_per_second_billing_test.go`; retain the pre-existing `docs/OPENAI_MEDIA_COMPAT.md` and `frontend/src/components/admin/channel/__tests__/PricingEntryCard.spec.ts` while removing only this entry's generic-template additions. No database rollback is required.

## 2026-08-08 - Task: Make asynchronous video billing and failure refunds consistent

### What was done

- Frozen the resolved billing model and price result at video submission. Balance reservation, atomic capture, the visible usage row, and any later refund now use that same snapshot even if channel mapping, pricing, or multipliers change while the task is running.
- Changed immediate terminal creation failures to release the reservation without charging. Accepted asynchronous tasks are persisted and reconciled by a lifecycle-managed background worker; terminal failures create one idempotent negative refund record while pending and transient probe failures are retried with leased multi-instance claims and bounded backoff.
- Added output video count, duration, resolution, output cost, and MiniMax-H3 verified reference-video duration/cost fields to usage persistence and both user/admin usage displays. MiniMax-H3 images and audio remain free, generic video reference media remains free, and refund rows negate the same output/input cost components.
- Added migration `196_openai_video_billing_reconciliation.sql` and updated billing/media operations documentation. The database and concurrency changes were implemented under the user's explicit approval.

### Testing

- Price-snapshot regression passed: a task estimated at `$1` remained a `$1` capture, `$1` usage row, and `-$1` refund after its configured price was changed to `$2` before settlement.
- Compensation regressions passed for pending reschedule, successful completion, transient HTTP 503 retry, terminal HTTP 500 generation-failure refund, duplicate-refund idempotency, and PostgreSQL `FOR UPDATE SKIP LOCKED` claiming.
- `go test -tags=unit ./... -count=1` from `backend/`: passed for all backend unit packages, including the migration schema and usage API contract suites.
- `go test -tags=unit ./internal/web ./cmd/server -count=1`: passed against the completed frontend production build.
- Focused usage-table tests passed 19 cases; the full frontend suite passed 218 files and 1546 tests. `pnpm typecheck`, `pnpm lint:check`, and `pnpm build` passed with only the existing Browserslist age and large-chunk warnings.
- `git diff --check`, `git diff --cached --check`, and strict conflict-marker scanning of task files passed.
- No paid production video task was submitted; settlement, refund, retry, persistence, and display behavior were verified with deterministic automated fixtures.

### Notes

- `backend/cmd/server/main.go`: starts asynchronous video failure reconciliation with the application.
- `backend/cmd/server/wire.go`: exposes the gateway/API-key services and stops reconciliation during cleanup.
- `backend/cmd/server/wire_gen.go`: regenerates dependency wiring for the worker lifecycle changes.
- `backend/internal/handler/dto/mappers.go`: maps video billing details into usage API responses.
- `backend/internal/handler/dto/types.go`: adds the public usage DTO fields for output and reference-video billing.
- `backend/internal/handler/openai_videos.go`: uses one preflight cost snapshot, handles immediate failures without charge, arms accepted tasks, and settles the reservation atomically.
- `backend/internal/repository/migrations_schema_integration_test.go`: verifies the new usage and compensation schema.
- `backend/internal/repository/openai_video_task_binding_repo.go`: persists, leases, and updates asynchronous compensation work.
- `backend/internal/repository/openai_video_task_binding_repo_test.go`: verifies leased `SKIP LOCKED` claiming.
- `backend/internal/repository/usage_log_repo_insert.go`: persists the three new video cost/detail columns in all insert paths.
- `backend/internal/repository/usage_log_repo_query.go`: reads the new video cost/detail columns in all usage queries.
- `backend/internal/server/api_contract_test.go`: updates the paginated usage response contract for video fields.
- `backend/internal/service/openai_gateway_record_usage_test.go`: covers immutable price settlement and matching refunds.
- `backend/internal/service/openai_gateway_service.go`: carries verified reference-video seconds and worker lifecycle state.
- `backend/internal/service/openai_gateway_usage.go`: applies the immutable snapshot and splits output/reference-video original costs.
- `backend/internal/service/openai_video_compensator.go`: implements persistent status reconciliation and terminal-failure refunds.
- `backend/internal/service/openai_video_compensator_test.go`: covers reconciliation states, retry behavior, and refund idempotency.
- `backend/internal/service/openai_videos.go`: defines compensation persistence, strict task-state recognition, and reversal usage records.
- `backend/internal/service/openai_videos_test.go`: covers immediate terminal failures and video task behavior.
- `backend/internal/service/usage_log.go`: adds internal output/reference-video billing fields.
- `backend/migrations/196_openai_video_billing_reconciliation.sql`: adds usage detail columns and persistent compensation state/indexes.
- `frontend/src/components/admin/usage/UsageTable.vue`: displays output seconds/resolution and MiniMax-H3 reference-video charge details for user and admin views.
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`: verifies normal and negative-refund video detail rendering.
- `frontend/src/i18n/locales/en/dashboard.ts`: adds English video usage labels.
- `frontend/src/i18n/locales/zh/dashboard.ts`: adds Chinese video usage labels.
- `frontend/src/types/index.ts`: adds the frontend usage response fields.
- `docs/BILLING_INTEGRITY.md`: documents immutable video settlement and reliable failed-task refunds.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents terminal-state handling, frozen prices, and visible video usage details.
- `progress.md`: records this implementation, validation evidence, file inventory, and rollback point.
- Rollback: preserve the dirty worktree, stop the service, and reverse only the hunks/files listed in this entry; do not delete pre-existing untracked video files wholesale. Deploying the preceding application build while retaining migration 196 is the preferred rollback because the added nullable/defaulted columns are backward compatible. If schema removal is mandatory, export the affected tables first, then drop `idx_openai_video_task_bindings_compensation_pending`, the compensation columns added to `openai_video_task_bindings`, and the three video detail columns added to `usage_logs`.

## 2026-08-08 - Task: Simplify image and Firefly video API documentation

### What was done

- Renamed the first user-facing family to “OpenAI gpt-image2” and limited it to `gpt-image-2` image generation and multipart image editing. Removed its material-upload, speech, transcription, translation, and video-facing content.
- Replaced the former SD2.0 family and separate Firefly family with one “最强视频模型2.0” section. Removed the previous models and examples, and documented only `firefly-video-v2` and `firefly-video-v2-fast`.
- Kept four directly usable Firefly examples for text-to-video, storyboard generation, multipart reference files, and task status/file/download requests. Updated the formal user API document to match the page.

### Testing

- `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts --reporter=dot`: passed 1 file and 1 test.
- `pnpm typecheck` and `pnpm lint:check`: passed.
- `pnpm build`: passed with only the existing Browserslist age and large-chunk warnings.
- `go test -tags=unit ./internal/web ./cmd/server -count=1`: passed against the rebuilt frontend assets.
- Runtime-source scan confirmed the removed title, former model names, `/pg/assets`, and OpenAI audio endpoints are absent from the user page and `docs/API_DOCS.md`.
- `git diff --check` and `git diff --cached --check`: passed.

### Notes

- `frontend/src/views/user/ApiDocsView.vue`: reduces the OpenAI section to images and replaces the old video sections with the Firefly-only “最强视频模型2.0” documentation and examples.
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: verifies the four-family order, required Firefly calls, image-only OpenAI content, and absence of removed documentation.
- `docs/API_DOCS.md`: aligns the formal user documentation with the simplified image and Firefly video sections.
- `progress.md`: records the documentation change, verification evidence, file inventory, and rollback point.
- Rollback: preserve the dirty worktree and reverse only the hunks described in this entry in the three documentation/page files above; do not restore the files wholesale because they contain earlier uncommitted interface-document work. No database or backend routing rollback is required.

## 2026-08-08 - Task: Rename the Firefly documentation family

### What was done

- Renamed the user-facing Firefly family from “最强视频模型2.0” to “视频模型SD2.0” without changing its models, parameters, examples, or backend behavior.

### Testing

- `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts --reporter=dot`: passed 1 file and 1 test.
- Runtime-source scan confirmed all three current documentation/test locations use “视频模型SD2.0” and no longer use the previous title.
- `git diff --check`: passed.

### Notes

- `frontend/src/views/user/ApiDocsView.vue`: updates the displayed video-family title.
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: updates the expected family title.
- `docs/API_DOCS.md`: updates the formal user documentation heading.
- `progress.md`: appends this rename record without rewriting the preceding history.
- Rollback: replace “视频模型SD2.0” with “最强视频模型2.0” only in the three current files listed above. No database, API, or backend rollback is required.

## 2026-08-08 - Task: Add MiniMax-H3 user API documentation

### What was done

- Added a standalone “视频模型MiniMax-H3” family after “视频模型SD2.0”, covering text-to-video, first/last-frame video, and multimodal reference video generation.
- Documented the public `/pg/assets` upload flow and `/v1/videos` create, query, and download flow with directly usable curl examples.
- Explained the `content[]` request structure, supported resolutions, duration and aspect-ratio options, material roles and limits, task states, and reference-video billing behavior in user-facing language.

### Testing

- `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts --reporter=dot`: passed 1 file and 1 test, including MiniMax-H3 order, examples, material limits, and billing guidance.
- `pnpm build`: passed with only the existing Browserslist age and large-chunk warnings.
- `go test -tags=unit ./internal/web ./cmd/server -count=1`: passed against the rebuilt frontend assets.
- `git diff --check` and `git diff --cached --check`: passed.

### Notes

- `frontend/src/views/user/ApiDocsView.vue`: adds the MiniMax-H3 model family, parameters, notes, and five curl examples.
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: verifies the five-family order and MiniMax-H3 calls, constraints, task state, and billing text.
- `docs/API_DOCS.md`: records the MiniMax-H3 public API contract and user-visible material rules.
- `progress.md`: records this documentation change, validation evidence, file inventory, and rollback point.
- Rollback: preserve the dirty worktree and reverse only the MiniMax-H3 section, its curl constants, and its test assertions in the three documentation/page files above; then remove this progress entry. No backend route, billing, database, or deployment rollback is required.

## 2026-08-08 - Task: Add resolution-based SD2.0 video pricing

### What was done

- Replaced the Firefly unified per-second channel template with an SD2.0 resolution template: `firefly-video-v2-fast` requires 480p/720p prices, while `firefly-video-v2` requires 480p/720p/1080p prices.
- Required each SD2.0 model to use a standalone pricing rule, rejected missing, zero, duplicate, unsupported, or legacy unified prices, and allowed existing unified rules to be migrated by entering their first resolution price.
- Applied the selected resolution price to preflight balance checks, final user charges, immutable task cost snapshots, and failed-task refunds. Reference image, video, and audio materials remain free for these models.
- Extended user and admin usage details to show the actual video resolution and derived output-video price per second alongside duration and cost.

### Testing

- `pnpm exec vitest run src/components/admin/channel/__tests__/PricingEntryCard.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts --reporter=dot`: passed 2 files and 23 tests.
- `pnpm typecheck` and `pnpm lint:check`: passed.
- `pnpm test:run`: full frontend suite passed; existing Vue test-stub, localStorage, Browserslist, and expected stderr warnings remain unchanged.
- `pnpm build`: passed with only the existing Browserslist age and large-chunk warnings.
- `go test -tags=unit ./internal/service ./internal/handler ./internal/repository -count=1`: passed.
- `go test -tags=unit ./internal/service -count=1`: passed after the final pricing validation changes.
- `go test -tags=unit ./internal/web ./cmd/server -count=1`: passed against the rebuilt frontend assets.
- `git diff --check` and `git diff --cached --check`: passed.

### Notes

- `backend/internal/service/channel.go`: clarifies the persisted video price field semantics.
- `backend/internal/service/channel_service.go`: validates standalone SD2.0 model rules and their required positive resolution prices.
- `backend/internal/service/channel_service_test.go`: covers valid standard/fast templates and invalid mixed, incomplete, and zero-price configurations.
- `backend/internal/service/video_billing_resolution.go`: defines canonical SD2.0 resolutions supported by each model.
- `backend/internal/service/openai_gateway_usage.go`: selects the SD2.0 channel price by the request's actual resolution.
- `backend/internal/service/openai_video_per_second_billing_test.go`: verifies all five model/resolution combinations, unsupported fast 1080p, free reference media, multiplier application, and recorded resolution.
- `backend/internal/service/openai_gateway_record_usage_test.go`: verifies the resolution and frozen cost survive settlement and failure refund unchanged.
- `frontend/src/components/admin/channel/PricingEntryCard.vue`: renders the SD2.0 resolution inputs and migrates an edited legacy unified rule.
- `frontend/src/components/admin/channel/__tests__/PricingEntryCard.spec.ts`: verifies standard, fast, MiniMax-H3, and legacy-rule migration templates.
- `frontend/src/views/admin/ChannelsView.vue`: enforces standalone models and complete positive resolution pricing before save.
- `frontend/src/components/admin/usage/UsageTable.vue`: displays video resolution and the derived output-video unit price.
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`: verifies video unit-price and reference-video billing details.
- `frontend/src/i18n/locales/zh/admin/channels.ts`: adds Chinese SD2.0 pricing guidance and validation messages.
- `frontend/src/i18n/locales/en/admin/channels.ts`: adds matching English SD2.0 pricing guidance and validation messages.
- `frontend/src/i18n/locales/zh/dashboard.ts`: adds the Chinese output-video unit-price label.
- `frontend/src/i18n/locales/en/dashboard.ts`: adds the matching English output-video unit-price label.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents SD2.0 resolution pricing, settlement/refund consistency, and legacy-rule migration.
- `docs/BILLING_INTEGRITY.md`: records resolution-aware SD2.0 billing in the integrity contract.
- `progress.md`: records this implementation, validation evidence, file inventory, migration note, and rollback point.
- Migration: existing Firefly `video` rules with one unified `per_request_price` must be split into one rule per model and populated with all supported resolution prices before those models can be called.
- Rollback: preserve the dirty worktree and reverse only the SD2.0 pricing hunks listed in this entry, restoring Firefly channel-video pricing to `PerRequestPrice` and removing the SD2.0 resolution validators/UI. No database rollback is required; existing interval rows use the current pricing schema and may remain unused after application rollback.

## 2026-08-11 - Task: Recover MiniMax-H3 submissions after upstream 504

### What was done

- Added persistent MiniMax-H3 submission recovery around `POST /v1/videos`: the original account task baseline, request signature, billing snapshot, balance hold and task ownership are saved before the paid upstream request.
- Converted an explicit MiniMax-H3 creation HTTP 504 into a normal HTTP 200 `task_recovery_...` response. The client can keep polling that task ID while a background worker uses the original account to find and bind the unique matching upstream task; the creation request is never failed over or resubmitted.
- Required status and content requests for persisted async tasks to stay on their original account even when advanced sticky weighting or sticky health escape is enabled.
- Made recovery restart-safe and multi-instance safe with leased `SKIP LOCKED` claims. Baseline, exact request attributes and a unique upstream-task ownership index prevent old tasks, multiple candidates and concurrent recoveries from being guessed or cross-bound.
- Reused one billing task ID and immutable price/hold snapshot across the immediate response and recovery worker. Existing usage is not billed twice; unresolved or ambiguous recovery releases the hold or writes the existing idempotent failed-video refund.
- Documented the client contract: continue polling the returned local task ID and do not retry task creation after a recovered 504.

### Testing

- `go test -count=1 ./...`: passed the full backend suite before the final hard-bound task scheduling guard.
- `go test -count=1 ./internal/service ./internal/handler ./internal/repository ./cmd/server`: passed after the final scheduling guard; service, handler, repository and startup wiring all passed without cache.
- `go test -race ./internal/service -run 'Test(OpenAIGatewayService_SelectAccountWithScheduler_RequiredTaskAccountDoesNotEscape|MiniMaxVideoRecoveryWorker|ForwardOpenAIVideoMiniMaxCreate504)' -count=1`: passed.
- `go test -tags integration ./internal/repository -run TestMigrationsRunner_IsIdempotent_AndSchemaIsUpToDate -count=1 -v`: passed against temporary PostgreSQL and Redis containers after one initial Docker Hub Redis pull failed with an external OAuth EOF.
- Worker tests passed for restart claims, a unique exact candidate, no-candidate rescheduling, ambiguous-candidate hold release and existing-usage billing idempotency. Service tests passed for 504 conversion and recovered status/content ID rewriting.
- `git diff --check` and `git diff --cached --check`: passed.
- No real MiniMax paid creation request was sent. The production-only remaining validation is that the configured metaso task-list response exposes the documented model, resolution, duration, ratio and reference-media count fields needed for a unique match.

### Notes

- `backend/migrations/197_openai_video_submission_recovery.sql`: adds persistent recovery state, billing snapshots, worker scheduling fields and unique upstream-task ownership.
- `backend/internal/handler/openai_videos.go`: prepares recovery before submission, returns a normal local task on 504, routes recovery status/content and keeps bound tasks on the original account.
- `backend/internal/repository/openai_video_recovery_repo.go`: persists, reads, leases and updates recovery records and client-to-upstream task routes.
- `backend/internal/repository/openai_video_task_binding_repo.go`: carries upstream and billing task IDs through asynchronous failure compensation.
- `backend/internal/repository/openai_video_task_binding_repo_test.go`: verifies leased recovery records can be scanned after restart.
- `backend/internal/repository/migrations_schema_integration_test.go`: verifies all recovery columns and indexes in a real migrated schema.
- `backend/internal/service/openai_gateway_service.go`: adds recovery and billing identifiers to video forward results and worker lifecycle state.
- `backend/internal/service/openai_videos.go`: intercepts only explicit MiniMax-H3 creation 504 responses and preserves local task IDs for recovered status and downloads.
- `backend/internal/service/openai_minimax_video.go`: maps recovered MiniMax status and content responses without exposing the upstream task ID.
- `backend/internal/service/openai_minimax_video_recovery.go`: builds request signatures, captures task-list baselines, parses task summaries and finds exact new candidates on the original account.
- `backend/internal/service/openai_video_compensator.go`: runs persistent recovery, idempotent billing, task compensation and unresolved-submission hold release/refund.
- `backend/internal/service/openai_minimax_video_recovery_test.go`: verifies media counts, baseline exclusion and exact candidate matching.
- `backend/internal/service/openai_video_compensator_test.go`: verifies restart recovery, unique/no/ambiguous candidates, hold release and no duplicate billing.
- `backend/internal/service/openai_videos_test.go`: verifies 504 normalization and recovered status/download task-ID rewriting.
- `backend/internal/service/openai_account_scheduler.go`: adds the opt-in hard account binding used only by persisted async task operations.
- `backend/internal/service/openai_account_scheduler_test.go`: verifies a bound video task cannot escape to another account because of health scoring.
- `backend/internal/service/request_metadata.go`: carries the required original account through the existing scheduler without changing ordinary requests.
- `docs/API_DOCS.md`: tells clients how to handle a `task_recovery_...` response.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents recovery matching, timeout, billing, restart and ambiguity behavior.
- `progress.md`: records this implementation, verification evidence, file inventory and rollback point.
- Migration: deploy migration 197 before serving the new image. Existing rows receive inactive/default recovery values; the new unique index only covers identified or matched non-empty recovery task IDs.
- Rollback: first stop all application instances, then roll back the application files listed above. If database rollback is required, drop indexes `idx_openai_video_task_bindings_recovery_upstream` and `idx_openai_video_task_bindings_recovery_pending`, then drop the 11 columns added by migration 197 from `openai_video_task_bindings`; preserve migration 196 compensation columns and all existing task rows. Do not drop recovery columns while any application instance still runs this code.

## 2026-08-12 - Task: Synchronize upstream/main v0.1.173

### What was done

- Fetched and merged 174 upstream commits in `68d8f122e..1e618dbc2`, advancing the version file from v0.1.172 to v0.1.173 and including 55 post-tag commits in merge commit `32459c19a0dd98ae6408d4bc227fbf99ae59c0a9`.
- Preserved the customized worktree through backup branch `backup/pre-upstream-sync-20260811-231855` and stash object `704efd34e5144ab0ee60f479738420e4cf404eb3`, restored tracked and untracked paths, and retained the stash as the exact recovery point.
- Integrated the user-approved authentication, concurrency, networking, billing, and database changes, including Channel Monitor V2, complete Grok media/search support, scheduling thresholds, response-model billing, API Key validation, backup parts, and 17 migrations.
- Resolved 23 restore conflicts while retaining the fork's Gemini native image accounting, custom OpenAI/Grok media routes, MiniMax-H3 recovery, resolution-aware video billing, community chat, public assets, and mandatory usage settlement.
- Regenerated Wire dependencies and repaired only the two merge-combination test stubs required by the expanded upstream interfaces. No running service or business database was changed.

### Testing

- `go generate ./cmd/server`: passed from `backend/`.
- `go test -tags=unit ./... -count=1`: passed the complete backend suite, including service, handler, routing, repository, migration, web embed, and server packages.
- `go test -tags=integration ./internal/repository -run 'TestMigrationsRunner_(ConcurrentInstancesSerializeOnSessionLock|IsIdempotent_AndSchemaIsUpToDate)' -count=1 -v`: passed against temporary PostgreSQL 18.1 and Redis 8.4 containers.
- `pnpm install --frozen-lockfile`, `pnpm test:run --reporter=dot`, `pnpm lint:check`, `pnpm typecheck`, and `pnpm build`: passed; Vitest reported 230 files and 1603 tests. Existing JSDOM/Vue, Browserslist, and large-chunk warnings remain non-fatal.
- `go test -tags=unit ./internal/web ./cmd/server -count=1`: passed after rebuilding the production frontend.
- Root, deploy, development, and standalone Compose files passed `docker compose config --quiet`; `bash deploy/test-caddyfile-cache.sh` passed.
- `git diff --check`, `git diff --cached --check`, strict conflict-marker scanning, unresolved-index checks, and `git merge-base --is-ancestor upstream/main HEAD`: passed.
- Stash inventory comparison accounted for all 327 original paths: 325 remain restored changes; the rollback timeout test is now represented by the upstream equivalent expression, and the group duplicate test intentionally uses the upstream Live-capability-disabled scenario. One added usage repository test stub method is the required upstream interface adaptation.
- `go mod tidy -diff` proposed only historical checksum removals, including the `google/subcommands` entries required by reproducible Wire generation; no dependency files were rewritten by that check.

### Notes

- Upstream merge inventory: `git diff --name-status backup/pre-upstream-sync-20260811-231855..upstream/main` lists all 489 upstream-changed files; `git show --name-only --format= 32459c19a0dd98ae6408d4bc227fbf99ae59c0a9` identifies the merge-resolution surface.
- `backend/cmd/server/wire_gen.go`: regenerates the combined dependency graph for upstream Channel Monitor V2/Grok services and local community/public-media handlers.
- `backend/internal/handler/dto/settings.go`, `backend/internal/handler/setting_handler.go`, and `backend/internal/service/setting_public.go`: combine the upstream public registration-domain setting with the fork's image workspace setting.
- `backend/internal/handler/handler.go`, `backend/internal/handler/wire.go`, and `backend/internal/server/routes/gateway.go`: retain local OpenAI media/community integrations while registering upstream Grok media, Voice, Web Search, and monitoring routes.
- `backend/internal/handler/openai_gateway_handler.go`, `backend/internal/service/openai_gateway_service.go`, and `backend/internal/service/openai_gateway_grok.go`: combine upstream gateway dependencies and Grok behavior with local video task binding and recovery services.
- `backend/internal/service/gateway_usage_billing.go` and `backend/internal/service/openai_gateway_usage.go`: combine response-model billing with the fork's strict pricing errors, mandatory usage tasks, and immutable video settlement data.
- `backend/internal/service/gemini_messages_compat_service.go`: retains actual Gemini image output accounting while adopting upstream response-model observations.
- `backend/internal/service/video_billing_resolution.go`: adapts local MiniMax-H3 and SD2.0 resolution pricing to the upstream video-price capability signature.
- `backend/internal/service/account_usage_service_batch_test.go`: adds the request-ID/API-Key lookup required by the expanded usage repository interface.
- `backend/internal/service/openai_videos_test.go`: adds no-op Grok pending-billing cache methods to the existing local video gateway test stub.
- `backend/migrations/194_channel_monitor_v2.sql`, `backend/migrations/195_channel_monitor_mode.sql`, `backend/migrations/196_channel_monitor_v2_ignored_error_categories.sql`, `backend/migrations/197_channel_monitor_v2_seed_popular_models.sql`, `backend/migrations/198_channel_monitor_v2_health_thresholds.sql`, `backend/migrations/199_channel_monitor_v2_fixed_rollups.sql`, `backend/migrations/200_channel_monitor_v2_rollup_permissions.sql`, `backend/migrations/201_channel_monitor_v2_refresh_5m.sql`, `backend/migrations/202_channel_monitor_v2_full_table_permissions.sql`, `backend/migrations/203_channel_monitor_v2_default_ignore_and_cache.sql`, `backend/migrations/204_channel_monitor_hide_throughput.sql`, `backend/migrations/205_channel_monitor_v2_reset_factory_cache_thresholds.sql`, and `backend/migrations/206_channel_monitor_v2_privacy_defaults.sql`: add and evolve the upstream Channel Monitor V2 schema without renaming the fork's same-prefix migrations.
- `backend/migrations/217_group_video_model_prices.sql`, `backend/migrations/218_group_audio_voice_pricing.sql`, `backend/migrations/219_group_search_price_per_1k.sql`, and `backend/migrations/220_clear_non_grok_video_generation_config.sql`: add upstream group-level Grok media/search pricing and cleanup semantics.
- `frontend/src/api/__tests__/admin.system.rollback.spec.ts`, `frontend/src/views/admin/__tests__/GroupsView.columnSettings.spec.ts`, and `frontend/src/views/admin/__tests__/GroupsView.duplicate.spec.ts`: adopt upstream timeout and capability defaults while retaining the fork's local group-column coverage.
- `docs/UPSTREAM_SYNC.md`: records scope, conflict choices, migration impact, validation evidence, and recovery points.
- `progress.md`: appends this audit record without rewriting existing history.
- Rollback: preserve the dirty worktree, then run `git revert -m 1 32459c19a0dd98ae6408d4bc227fbf99ae59c0a9`. The pre-sync branch is `backup/pre-upstream-sync-20260811-231855`, and the full pre-sync worktree stash object is `704efd34e5144ab0ee60f479738420e4cf404eb3`. Do not drop that stash before deployment acceptance.

## 2026-08-12 - Task: Publish iotwq/china-api:latest multi-architecture image

### What was done

- Built the complete current worktree as Sub2API `0.1.173` with commit metadata `32459c19a` for `linux/amd64` and `linux/arm64`.
- Pushed both architectures as `iotwq/china-api:latest` to Docker Hub and replaced the previous remote OCI index only after both child manifests were ready.
- Preserved the previous remote index digest as an explicit rollback point and removed the two local verification tags after validation.

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.173 --build-arg COMMIT=32459c19a --tag iotwq/china-api:latest --push .`: passed.
- Independent post-push `docker buildx imagetools inspect iotwq/china-api:latest` confirmed remote OCI index `sha256:cc7b6daf0efe831382ec2f3111492b7548bf7341862dffaee7e3705d42949d65` with amd64 child `sha256:8db9b9b4d698caabc741ee7dfc129a6307b9a93847be4a6e73f9627da18524ca` and arm64 child `sha256:4359bfe16329f0736384aa60ff47ddb54c258710dca4f9d4787f1f53b8163477`.
- Loaded both exact child manifests from the completed Buildx cache and ran each platform with `--version`; amd64 and arm64 both reported `Sub2API 0.1.173 (commit: 32459c19a)`.
- Two direct post-push arm64 blob pulls were interrupted before startup by Docker Hub CloudFront/OAuth EOF responses. This did not affect the successful push, independent remote manifest read, or exact-child runtime checks; no failed container was left running.

### Notes

- `progress.md`: records the Docker Hub publication, architecture digests, runtime evidence, transient registry-download limitation, and rollback point; no business source or deployment file was changed by this publication task.
- Previous remote index: `sha256:7ec3d8a29abf752683348a1ab89000d744f118c0f41eed19e7f281449d4f613e` (amd64 `sha256:6d5db4287353f0cd49b34c4d260aece0dd3131611bdffa615f3e0a5ad40583c9`, arm64 `sha256:851edc54655dfe749b1aac6ae1a2fc28764969bbe7f1331f601478e495122a90`).
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:7ec3d8a29abf752683348a1ab89000d744f118c0f41eed19e7f281449d4f613e`; this restores the previous multi-architecture index without changing the dirty worktree.

## 2026-08-12 - Task: Fix MiniMax-H3 504 recovery billing and permanent processing state

### What was done

- Changed explicit MiniMax-H3 creation HTTP 504 handling to retain the existing balance hold without immediately creating a successful usage charge. The immutable recovery billing snapshot is settled only after the original account exposes one unique matching upstream task.
- Made an identified recovery remain claimable while billing, compensation arming, or the final `matched` update is incomplete. Retry scheduling now uses a fresh bounded context, so an expired worker context cannot silently strand the record.
- Allowed already-identified tasks to survive the three-minute discovery deadline and restored legacy records that were billed and armed but left as `identified` with no next recovery time.
- Routed persisted MiniMax recovery status and content requests through the MiniMax capability on the original account. Active finalization still returns `processing`; only a completed `matched` record or the precise legacy stranded shape may query the real upstream task, preventing unbilled delivery.
- Kept all changes inside the existing recovery schema and MiniMax path. Other video models and ordinary OpenAI media billing behavior are unchanged.

### Testing

- `go test ./...`: passed the complete backend suite after the final billing-delivery guard.
- `go test -race ./internal/service ./internal/handler ./internal/repository -run 'Test(MiniMaxVideoRecoveryWorker|VideosMiniMax504KeepsHoldWithoutImmediateCharge|VideoStatusUsesIdentifiedUpstreamTaskBeforeMatchedUpdate|VideoStatusKeepsActivelyFinalizingIdentifiedTaskProcessing|ClaimOpenAIVideoRecoveries)' -count=1`: passed.
- Handler regression tests verified a 504 response returns HTTP 200 with `task_recovery_...`, retains one balance hold, and creates no usage row, capture, or charge before unique matching.
- Recovery tests verified expired identified records still complete, a failed `matched` write is rescheduled and succeeds on the next pass, legacy stranded records query `succeeded`, and actively finalizing records remain `processing` until settlement is ready.
- `git diff --check` and `git diff --cached --check`: passed before the progress entry; final whitespace validation was repeated after it.
- No real MiniMax paid request was sent. No database migration or production data update was required.

### Notes

- `backend/internal/handler/openai_videos.go`: defers 504 billing, guards delivery until settlement, and forces recovered status/content requests through the MiniMax capability and original account.
- `backend/internal/handler/openai_videos_recovery_billing_test.go`: covers hold-only 504 handling, legacy identified-task delivery, and active-finalization blocking.
- `backend/internal/repository/openai_video_recovery_repo.go`: makes legacy `identified` rows with a real upstream task and missing next-check time claimable again.
- `backend/internal/repository/openai_video_task_binding_repo_test.go`: verifies the legacy stranded-row claim condition remains in the leased `SKIP LOCKED` query.
- `backend/internal/service/openai_video_compensator.go`: preserves identified recovery scheduling, retries finalization with a fresh context, and does not expire an already-identified task.
- `backend/internal/service/openai_video_compensator_test.go`: verifies restart/idempotency, expired identified recovery, and retry after a failed final `matched` update.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents hold-only 504 behavior, settlement timing, retry semantics, and legacy stranded-task recovery.
- `progress.md`: records this implementation, verification evidence, file inventory, and rollback point.
- Rollback: redeploy the pre-fix image `iotwq/china-api@sha256:cc7b6daf0efe831382ec2f3111492b7548bf7341862dffaee7e3705d42949d65`; no schema rollback is needed. Source rollback must reverse only the MiniMax recovery hunks listed in this entry and leave unrelated dirty-worktree changes intact.

## 2026-08-13 - Task: Merge upstream v0.1.175

### What was done

- Fetched and merged all 31 upstream commits in `1e618dbc2..5935e674a`, advancing the version file from `0.1.173` to `0.1.175` in merge commit `2520fd7b93ff2986cc0e3a8d8582a19051ce0df6`.
- Preserved the complete customized worktree through backup branch `backup/pre-upstream-sync-20260813-002446` and stash object `321c931f8a818091237c87239451aa3b40943eb0`, then restored the original staged, unstaged, and untracked state with no missing paths.
- Integrated the approved authentication, networking, rate-limit, and security-audit changes, including Codex OAuth fingerprint convergence, WebSocket V2 terminal-event handling, Responses TTFT/usage parsing, HTML 403 account-penalty protection, Gemini schema normalization, and service-tier account cost reporting.
- Retained the fork's Gemini image accounting, strict billing safeguards, MiniMax-H3 504 recovery and deferred settlement, OpenAI media routes, public assets, community chat, channel monitoring, and customized frontend. No database migration was added or changed by this upstream batch.

### Testing

- `go generate ./cmd/server`: passed and retained the local video compensator, community chat, and public asset dependency graph.
- `go test ./...`: passed the complete backend suite.
- `go test -race ./internal/service ./internal/handler ./internal/repository -run 'Test(MiniMaxVideoRecoveryWorker|VideosMiniMax504KeepsHoldWithoutImmediateCharge|VideoStatusUsesIdentifiedUpstreamTaskBeforeMatchedUpdate|VideoStatusKeepsActivelyFinalizingIdentifiedTaskProcessing|CodexFingerprint|Ratelimit|SecurityAudit|Gemini)' -count=1`: passed.
- `pnpm lint:check`, `pnpm typecheck`, `pnpm test:run`, and `pnpm build`: passed; Vitest reported 231 files and 1606 tests. Existing Browserslist and large-chunk warnings remain non-fatal.
- Root, deploy, development, and standalone Compose files passed `docker compose config --quiet`; `bash deploy/test-caddyfile-cache.sh` passed.
- `git diff --check`, `git diff --cached --check`, unresolved-index checks, strict conflict-marker scanning, `git merge-base --is-ancestor upstream/main HEAD`, and the 327-path stash restoration comparison passed.

### Notes

- `docs/UPSTREAM_SYNC.md`: records the 31-commit scope, behavior changes, no-migration result, compatibility decisions, verification evidence, and recovery points.
- `progress.md`: appends this implementation and verification record without rewriting prior history.
- Upstream merge commit `2520fd7b93ff2986cc0e3a8d8582a19051ce0df6`: contains the 37 upstream-changed files; all pre-existing local modifications remain outside that merge commit in their original index/worktree state.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 2520fd7b93ff2986cc0e3a8d8582a19051ce0df6`. The exact pre-sync source point is branch `backup/pre-upstream-sync-20260813-002446`, and the full pre-sync worktree is stash object `321c931f8a818091237c87239451aa3b40943eb0`; do not drop it before deployment acceptance.

## 2026-08-13 - Task: Publish iotwq/china-api:latest multi-architecture image

### What was done

- Built the complete current worktree as Sub2API `0.1.175` with commit metadata `2520fd7b9` for `linux/amd64` and `linux/arm64`.
- Pushed both architectures to Docker Hub as `iotwq/china-api:latest`, replacing the previous remote index only after both child manifests and all layers were uploaded successfully.
- Preserved the previous remote index digest as the deployment rollback point and independently pulled and ran both new child manifests.

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.175 --build-arg COMMIT=2520fd7b9 --tag iotwq/china-api:latest --push .`: passed.
- Independent `docker buildx imagetools inspect iotwq/china-api:latest` confirmed remote OCI index `sha256:f806a5bbde1d1824f46e870f956b4dc4d53e8c6a18e8f6e4f74046e563530911` with amd64 child `sha256:e7abae5aa8d265a74a7ea0a91b1b3be8df15fea0afeaf7513ab57196aae80e48` and arm64 child `sha256:763f8165b95b8f5761fcd242194cab250274d95c8df09d30d5a442c224ae8c15`.
- Running each exact remote child manifest with `/app/sub2api --version` passed; both reported `Sub2API 0.1.175 (commit: 2520fd7b9)`.

### Notes

- `progress.md`: records the Docker Hub publication, remote architecture digests, runtime verification, and rollback point; no business source or deployment configuration was changed by this publication task.
- Previous remote index: `sha256:cc7b6daf0efe831382ec2f3111492b7548bf7341862dffaee7e3705d42949d65` (amd64 `sha256:8db9b9b4d698caabc741ee7dfc129a6307b9a93847be4a6e73f9627da18524ca`, arm64 `sha256:4359bfe16329f0736384aa60ff47ddb54c258710dca4f9d4787f1f53b8163477`).
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:cc7b6daf0efe831382ec2f3111492b7548bf7341862dffaee7e3705d42949d65`; this restores the previous multi-architecture index without changing the source worktree.

## 2026-08-13 - Task: Preserve Grok 4.6 xhigh reasoning effort

### What was done

- Added `grok-4.6` to the Grok reasoning-effort capability check and preserved its official `xhigh` value on Chat Completions and Responses requests.
- Kept the existing compatibility behavior for all other Grok models: unsupported models still remove the field, while legacy `xhigh`, `extra-high`, `max`, and `ultra` aliases still normalize to `high`.
- Left channel pricing, account scheduling, model mapping, authentication, and billing unchanged.

### Testing

- Before the implementation, the new end-to-end request-body regressions failed because both Grok 4.6 outbound payloads had removed `xhigh`, reproducing the reported behavior.
- `env GOCACHE=/tmp/sub2api-go-cache go test -tags=unit ./internal/service -run 'TestForwardAsRawChatCompletions_Grok46PreservesXHighEffort|TestForwardGrokResponsesAPIKeyPreservesGrok46XHighEffort|TestPatchGrokResponsesBodyNormalizesReasoningEffortAliases|TestNormalizeGrokChatReasoningEffort' -count=1`: passed after the fix.
- `env GOCACHE=/tmp/sub2api-go-cache go test -tags=unit ./internal/service -count=1`: passed the complete service unit suite outside the restricted sandbox because existing `httptest` cases require loopback listeners.

### Notes

- `backend/internal/service/openai_gateway_grok.go`: recognizes Grok 4.6 reasoning capability and preserves only its official `xhigh` value.
- `backend/internal/service/openai_gateway_grok_test.go`: verifies the real Responses forwarding request retains nested Grok 4.6 `xhigh` and existing Grok 4.5 aliases still normalize to `high`.
- `backend/internal/service/openai_gateway_chat_completions_raw_test.go`: verifies the real APIKey Chat Completions forwarding request retains Grok 4.6 `xhigh`.
- `docs/UPSTREAM_SYNC.md`: documents supported request formats, pricing independence, and legacy compatibility behavior.
- `progress.md`: appends this implementation, validation evidence, file inventory, and rollback instructions.
- Rollback: reverse only the Grok 4.6 reasoning hunks in the five files listed above; no database, pricing, or deployment rollback is required.

## 2026-08-13 - Task: Add default Grok 4.6 pricing and model catalog support

### What was done

- Added an exact `grok-4.6` fallback pricing entry that reuses the existing Grok 4.5 token rates: $2/MTok input, $0.50/MTok cached input, and $6/MTok output.
- Added `grok-4.6` to deterministic token-pricing recognition and the built-in Grok model catalog/identity mapping, without adding wildcard family pricing or unrequested aliases.
- Preserved channel-pricing precedence: an administrator's existing Grok 4.6 input, output, or cache prices continue to override the new defaults. Existing group multipliers and settlement calculations are unchanged.

### Testing

- Before implementation, the new regressions reproduced `pricing not found for model: grok-4.6` and confirmed the built-in Grok catalog omitted the model.
- `env GOCACHE=/tmp/sub2api-go-cache go test -tags=unit ./internal/service ./internal/pkg/xai -run 'TestGetModelPricing_Grok46MatchesGrok45Default|TestGetModelPricingWithChannel_Grok46ChannelPriceOverridesDefault|TestDefaultModelMappingExcludesCrossClientWildcards|TestResolveGrokTextResponsesModelID' -count=1`: passed after the fix.
- `env GOCACHE=/tmp/sub2api-go-cache go test -tags=unit ./internal/service ./internal/pkg/xai -count=1`: passed the complete service and xAI unit suites outside the restricted sandbox because existing `httptest` cases require loopback listeners.
- `env GOCACHE=/tmp/sub2api-go-cache go test -tags=unit ./internal/handler/admin -run 'TestAccountHandlerGetAvailableModels_GrokDefaultsToXAIModelsWithoutMapping' -count=1`: passed and confirmed the administrator model endpoint exposes Grok 4.6 by default.
- Cost regression verified identical token usage and multiplier produce identical `TotalCost` and `ActualCost` for Grok 4.5 and Grok 4.6; channel override regression verified configured prices take precedence.

### Notes

- `backend/internal/service/billing_service.go`: adds exact Grok 4.6 fallback and deterministic pricing recognition by sharing the Grok 4.5 rate object.
- `backend/internal/service/billing_service_test.go`: verifies default rate equality, real cost equality, deterministic recognition, unknown-model rejection, and channel override precedence.
- `backend/internal/pkg/xai/models.go`: adds Grok 4.6 to the built-in text model catalog and identity mapping.
- `backend/internal/pkg/xai/models_test.go`: verifies default catalog, mapping, Responses model recognition, and canonical resolution.
- `backend/internal/handler/admin/account_handler_available_models_test.go`: verifies the administrator account-model endpoint exposes Grok 4.6 by default.
- `docs/UPSTREAM_SYNC.md`: documents the default rates and channel-pricing precedence.
- `progress.md`: appends this implementation, validation evidence, file inventory, and rollback instructions.
- Rollback: reverse only the Grok 4.6 pricing/catalog hunks in the seven files listed above; no database or data migration rollback is required.

## 2026-08-13 - Task: Publish latest Grok 4.6 updates to iotwq/china-api:latest

### What was done

- Built the complete current worktree as Sub2API `0.1.175` with commit metadata `2520fd7b9` for `linux/amd64` and `linux/arm64`.
- Published both architectures to Docker Hub as `iotwq/china-api:latest`, including the latest Grok 4.6 reasoning-effort and default-pricing changes.
- Preserved the previous remote multi-architecture index as the deployment rollback point.

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.1.175 --build-arg COMMIT=2520fd7b9 --tag iotwq/china-api:latest --push .`: passed.
- Independent `docker buildx imagetools inspect iotwq/china-api:latest` confirmed remote index `sha256:87c71c33cb9c0eeda51e1eb806cf295d3c7f9669855c02ad2371b950dc6f8fac`, amd64 child `sha256:0cf466f0122956d7328dd0222a5c365a2fe0558eb3aeb8c25cdbc797a551a3fc`, and arm64 child `sha256:5e00f315b4f808e1c5bc0944033d823aabcbddd9db5e78553a6b2b8b43f943b7`.
- The exact remote arm64 child ran successfully and reported `Sub2API 0.1.175 (commit: 2520fd7b9)`.
- The amd64 child download encountered transient Docker Hub OAuth/CloudFront EOF responses. The exact same manifest `sha256:0cf466f0122956d7328dd0222a5c365a2fe0558eb3aeb8c25cdbc797a551a3fc` was loaded from the completed BuildKit cache and ran successfully with the same version and commit output.

### Notes

- `progress.md`: records this Docker Hub publication, remote architecture digests, runtime verification, transient registry-download issue, and rollback point; no business source or deployment configuration was changed by this publication task.
- Previous remote index: `sha256:f806a5bbde1d1824f46e870f956b4dc4d53e8c6a18e8f6e4f74046e563530911` (amd64 `sha256:e7abae5aa8d265a74a7ea0a91b1b3be8df15fea0afeaf7513ab57196aae80e48`, arm64 `sha256:763f8165b95b8f5761fcd242194cab250274d95c8df09d30d5a442c224ae8c15`).
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:f806a5bbde1d1824f46e870f956b4dc4d53e8c6a18e8f6e4f74046e563530911`; this restores the previous multi-architecture index without changing the source worktree.

## 2026-08-13 - Task: Sync upstream/main through v0.1.176

### What was done

- Merged 26 upstream commits from `5935e674a` through `fbfdcef81`, including the five commits after tag `v0.1.176`, into local merge commit `744c03c7e861d65675472cc5d286f9d07e8ccbe5`.
- Integrated upstream group model pricing, Grok subscription-tier and 4.6 support, `/x_search`, backup leader locking, cache invalidation, and Responses probe fixes while retaining all existing local media, billing-safety, monitoring, Gemini-native, and community-chat behavior.
- Resolved eight overlapping Grok and billing files semantically, then fixed three return/signature adaptations and the video-pricing form branch exposed by the combined code.
- Added upstream migration `221_group_model_pricing.sql`; no running database, service, or Docker image was modified by this task.

### Testing

- `env GOCACHE=/tmp/sub2api-go-build-cache go generate ./cmd/server`: passed.
- `env GOCACHE=/tmp/sub2api-go-build-cache go test ./...`: passed for the complete backend.
- Targeted `go test -race` runs for backup leader locking, Grok 4.6/long-context billing, billing snapshots, MiniMax billing, and video billing passed.
- `pnpm lint:check`, `pnpm typecheck`, `pnpm test:run`, and `pnpm build`: passed; Vitest reported 231 files and 1615 tests passed.
- `go test ./internal/web ./cmd/server -count=1`, migration/repository tests, and `docker compose config --quiet`: passed.
- `git diff --check`, `git diff --cached --check`, strict conflict-marker scanning, upstream ancestry, and the exact hash comparison for all pre-sync untracked files passed.

### Notes

- `backend/internal/pkg/xai/models.go`: combines the upstream Grok 4.6 latest alias with the local built-in Grok model catalog.
- `backend/internal/service/billing_service.go`: keeps Grok 4.6 priced exactly like Grok 4.5 and prevents generic unified billing from bypassing verified video duration handling.
- `backend/internal/service/channel.go`: preserves the JSON pricing contract and local `video` billing mode comments and validation.
- `backend/internal/service/grok_media.go`: combines upstream model-scoped usage snapshots with local binary video-content streaming.
- `backend/internal/service/openai_gateway_chat_completions_raw.go`: combines model-scoped Grok failure handling with normalized wrapped-status retries.
- `backend/internal/service/openai_gateway_grok.go`: keeps upstream Grok 4.6 aliases and local Grok 4.6 `xhigh` capability handling.
- `backend/internal/service/openai_gateway_grok_chat_bridge.go`: combines model-scoped Grok failure handling with normalized wrapped-status retries on the bridge path.
- `backend/internal/service/openai_gateway_usage.go`: combines upstream long-context gates with local immutable media billing snapshots and verified per-second video pricing.
- `backend/internal/service/gateway_usage_billing.go`: adapts local image and audio billing branches to the upstream error-returning cost contract.
- `frontend/src/components/admin/channel/PricingEntryCard.vue`: keeps image pricing on the upstream media form while preserving dedicated MiniMax-H3, SD2.0, and generic per-second video templates.
- `backend/cmd/server/wire_gen.go`: regenerated the dependency graph after synchronization.
- `backend/migrations/221_group_model_pricing.sql`: adds the upstream group model-pricing and long-context-pricing fields.
- `docs/UPSTREAM_SYNC.md`: records scope, migration impact, compatibility decisions, verification, and recovery points.
- `progress.md`: appends this synchronization record without rewriting previous entries.
- Upstream merge commit `744c03c7e861d65675472cc5d286f9d07e8ccbe5` contains the remaining upstream changes across 98 files.
- Rollback: first preserve the current dirty worktree, then run `git revert -m 1 744c03c7e861d65675472cc5d286f9d07e8ccbe5`. The exact pre-sync source point is branch `backup/pre-upstream-sync-20260813-222818`, and the full pre-sync worktree is stash object `5440dd51f71b36df26c7826a54a9a55b4558cc8a`; do not drop it before deployment acceptance.

## 2026-08-13 - Task: Fix MiniMax-H3 recovered task queries returning invalid task_id

### What was done

- Fixed recovered MiniMax-H3 task status and content requests so the original account is a hard routing requirement under both the default load scheduler and the optional advanced scheduler. An unavailable original account now stops routing instead of querying the task through another API Key.
- Added strict handling for MiniMax-H3 creation responses that use HTTP 200 while reporting `error.http_code = 504`; these responses now enter the existing `task_recovery_...` flow instead of being treated as successful task creation.
- Added recovery route context to existing request logs, including the client task ID, recovered upstream task ID, bound account ID, and the existing selected-account field. No database schema, pricing formula, settlement, or refund behavior was changed.

### Testing

- Before the fix, focused regressions reproduced all three defects: default scheduling selected account `21912` instead of required account `21911`, an unavailable required account switched to another account, and an HTTP 200 wrapped 504 was not marked `SubmissionUncertain`.
- `go test ./internal/service -run 'Test(OpenAIGatewayService_SelectAccountWithScheduler_RequiredTaskAccount.*LegacyScheduler|ForwardOpenAIVideoMiniMaxCreateWrapped504)' -count=1`: passed after the fix.
- `go test -race ./internal/service ./internal/handler ./internal/repository -run 'Test(OpenAIGatewayService_SelectAccountWithScheduler_RequiredTaskAccount|ForwardOpenAIVideoMiniMax|MiniMaxVideoRecoveryWorker|VideosMiniMax504KeepsHoldWithoutImmediateCharge|VideoStatusUsesIdentifiedUpstreamTaskBeforeMatchedUpdate|VideoStatusKeepsActivelyFinalizingIdentifiedTaskProcessing|ClaimOpenAIVideoRecoveries|OpenAIVideoTaskBinding)' -count=1`: passed.
- `go test ./internal/service ./internal/handler ./internal/repository -count=1`: passed the complete service, handler, and repository package suites.
- `git diff --check` for all task files: passed. No real MiniMax-H3 generation request was sent and no paid upstream task was created during verification.

### Notes

- `backend/internal/service/openai_account_scheduler.go`: enforces required-account selection in the default scheduler without falling back to other accounts.
- `backend/internal/service/openai_account_scheduler_test.go`: covers required-account selection and unavailable-account fail-closed behavior with the advanced scheduler disabled.
- `backend/internal/service/openai_videos.go`: recognizes strict HTTP 200 wrapped MiniMax 504 responses and preserves their upstream request ID in recovery diagnostics.
- `backend/internal/service/openai_videos_test.go`: verifies wrapped 504 responses return the local recovery task without writing the wrapped error to the client.
- `backend/internal/handler/openai_videos.go`: adds bound account and upstream task identity to the existing video request log context.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents wrapped 504 recovery and hard original-account routing in both scheduler modes.
- `progress.md`: appends this implementation, validation evidence, file inventory, and rollback instructions.
- Rollback: preserve the dirty worktree and reverse only the required-account helper/call in `backend/internal/service/openai_account_scheduler.go`, the two legacy-scheduler tests, the wrapped-504 branch/test, the added video request-log fields, and the two documentation paragraphs described above; do not delete the pre-existing untracked video files or use a broad `git restore`. No database or data rollback is required.

## 2026-08-14 - Task: Improve site-owner private messaging

### What was done

- Added image, MP4, and supported document attachments to site-owner private messages, including previews, downloads, optional captions, paste-to-upload for images, and local-file cleanup after message deletion.
- Added administrator-only active-user search so the site owner can select a user and send the first private message before a conversation exists.
- Constrained the administrator conversation area to a stable height with independent vertical scrolling and paginated loading, while preserving per-conversation unread state and real-time updates.
- Restricted private attachment reads to the matching user and administrators; public community-chat attachments retain their existing logged-in-user visibility.
- Added migration `222_community_chat_direct_attachments.sql` to allow `text`, `image`, and `file` direct-message types. No new table or column was introduced.
- Raised private-message media previews above the private-message dialog and kept direct media loading from changing the public-chat scroll position.

### Testing

- `go test -race ./internal/service ./internal/handler ./internal/repository -run 'CommunityChat|Direct' -count=1`: passed, including attachment validation, user targeting, multipart upload, ownership checks, deletion cleanup, unread behavior, and WebSocket visibility regressions.
- `go test ./internal/service ./internal/handler ./internal/repository ./internal/server/routes ./migrations -count=1`: passed the complete related backend package suites.
- `pnpm exec eslint src/api/communityChat.ts src/views/user/CommunityChatView.vue src/api/__tests__/communityChat.spec.ts src/views/user/__tests__/communityChatUnread.spec.ts src/i18n/locales/zh/misc.ts src/i18n/locales/en/misc.ts`: passed.
- `pnpm exec vitest run src/api/__tests__/communityChat.spec.ts src/views/user/__tests__/communityChatUnread.spec.ts src/views/user/__tests__/communityChatScroll.spec.ts`: passed 19 tests.
- `pnpm typecheck`: passed.
- `pnpm build`: passed; Vite emitted only the existing stale Browserslist-data and large-chunk warnings.
- `git diff --check` for the task's tracked files: passed. The repository does not provide a `prettier` executable, so no dependency was added and no repository-wide formatting was attempted.

### Notes

- `backend/migrations/222_community_chat_direct_attachments.sql`: expands the direct-message type constraint to `text`, `image`, and `file`.
- `backend/migrations/community_chat_direct_attachments_test.go`: verifies the direct-attachment migration contract.
- `backend/internal/service/community_chat.go`: validates private attachments, normalizes attachment responses, searches active users for administrators, and enforces private attachment access.
- `backend/internal/service/community_chat_test.go`: covers private attachments, invalid inputs, administrator targeting, user-target isolation, and user search permissions.
- `backend/internal/repository/community_chat_repo.go`: persists private attachment metadata and checks public/private attachment visibility in one query.
- `backend/internal/repository/community_chat_repo_test.go`: covers attachment persistence and matching/unrelated-user access outcomes.
- `backend/internal/handler/community_chat_handler.go`: adds private multipart upload and user-search handlers, authenticated attachment ownership checks, and private-file deletion cleanup.
- `backend/internal/handler/community_chat_handler_test.go`: covers real multipart private upload, private-file cleanup, attachment sessions, and event visibility.
- `backend/internal/server/routes/community_chat.go`: registers administrator user search and private-file upload routes with the existing authentication and rate limits.
- `frontend/src/api/communityChat.ts`: adds private attachment upload and administrator user-search clients and types.
- `frontend/src/api/index.ts`: exports the private-user result type.
- `frontend/src/api/__tests__/communityChat.spec.ts`: verifies private multipart fields and user-search query parameters.
- `frontend/src/views/user/CommunityChatView.vue`: adds private attachment rendering/composition, administrator search and first-contact flow, bounded scrolling, pagination, and nested media-preview behavior.
- `frontend/src/views/user/__tests__/communityChatUnread.spec.ts`: covers private attachment wiring, user search, bounded scrolling, pagination, and preview layering.
- `frontend/src/i18n/locales/zh/misc.ts`: adds Chinese labels for private user search, first contact, empty results, failures, and pagination.
- `frontend/src/i18n/locales/en/misc.ts`: adds matching English private-message labels.
- `docs/COMMUNITY_CHAT.md`: documents private attachments, administrator-initiated conversations, access rules, and the new endpoints.
- `progress.md`: records this implementation, validation evidence, file inventory, and rollback point.
- Rollback point: deploy the prior published image `iotwq/china-api@sha256:87c71c33cb9c0eeda51e1eb806cf295d3c7f9669855c02ad2371b950dc6f8fac`. Migration 222 may safely remain because it only broadens an existing check constraint; before reverting that constraint to text-only, remove all direct `image` and `file` rows or PostgreSQL will reject the rollback.

## 2026-08-14 - Task: Add a product-ready community chat emoji picker

### What was done

- Replaced the fixed eight-emoji shortcut row in public community chat with a smile button inside the message input and a full Unicode Emoji picker.
- Added localized search, category navigation, browser-local frequently used Emoji, skin-tone selection, keyboard accessibility, outside-click and Escape closing, and automatic light/dark theme matching.
- Preserved the textarea selection so an Emoji inserts at the current cursor or replaces selected text while continuing to enforce the existing 2,000-character message limit.
- Added viewport-aware popover positioning for desktop, narrow mobile screens, scrolling, resizing, and visual-viewport changes. Emoji data is packaged locally and does not depend on a third-party CDN.
- Kept the Emoji code in its own lazy route chunk so unrelated pages do not download the picker; coarse-pointer devices also avoid reopening the soft keyboard after every Emoji selection.
- Kept the scope to public community chat and standard Unicode Emoji; private messaging, stickers, GIFs, backend message formats, and database structures were not changed.

### Testing

- `pnpm exec vitest run src/components/community/__tests__/CommunityEmojiPicker.spec.ts src/views/user/__tests__/communityChatEmoji.spec.ts src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/views/user/__tests__/communityChatUnread.spec.ts src/views/user/__tests__/communityChatScroll.spec.ts`: passed 21 tests, covering open/close, localized local data, repeated selection, outside click, Escape focus restoration, 320px mobile positioning, cursor insertion, length limits, and existing unread/scroll behavior.
- `pnpm test:run`: passed the complete frontend suite with 234 files and 1,629 tests.
- `pnpm lint:check`: passed the complete frontend static analysis.
- `pnpm exec eslint src/components/community/CommunityEmojiPicker.vue src/components/community/__tests__/CommunityEmojiPicker.spec.ts src/views/user/CommunityChatView.vue src/views/user/communityChatEmoji.ts src/views/user/__tests__/communityChatEmoji.spec.ts src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/components/icons/Icon.vue src/i18n/locales/zh/misc.ts src/i18n/locales/en/misc.ts`: passed.
- `pnpm typecheck`: passed.
- `pnpm build`: passed and emitted the Chinese and English Emoji data as same-origin static assets (424.49 KB and 436.56 KB; 79.25 KB and 72.30 KB gzip). Vite reported only the existing stale Browserslist-data and large shared-chunk warnings.
- Real-page browser inspection could not run because the user's saved browser permission blocks automated access to `http://127.0.0.1`; no alternate browser surface or permission bypass was used.

### Notes

- `frontend/package.json`: adds pinned `emoji-picker-element` and `emoji-picker-element-data` runtime dependencies.
- `frontend/pnpm-lock.yaml`: locks the two Emoji picker packages and integrity hashes.
- `frontend/vite.config.ts`: isolates the picker runtime in a dedicated lazy vendor chunk.
- `frontend/src/components/community/CommunityEmojiPicker.vue`: implements the accessible, localized, theme-aware, responsive Emoji popover with same-origin data.
- `frontend/src/components/community/__tests__/CommunityEmojiPicker.spec.ts`: verifies the picker lifecycle, selection, dismissal, localization, and narrow-screen placement.
- `frontend/src/components/icons/Icon.vue`: adds the smile outline used by the composer button.
- `frontend/src/views/user/CommunityChatView.vue`: replaces fixed shortcuts with the picker and preserves textarea cursor/selection state.
- `frontend/src/views/user/communityChatEmoji.ts`: provides the length-safe cursor insertion helper.
- `frontend/src/views/user/__tests__/communityChatEmoji.spec.ts`: verifies insertion, selection replacement, and length-limit behavior.
- `frontend/src/views/user/__tests__/communityChatEmojiWiring.spec.ts`: verifies the public-chat-only picker wiring and embedded trigger layout.
- `frontend/src/i18n/locales/zh/misc.ts`: adds the Chinese picker button label.
- `frontend/src/i18n/locales/en/misc.ts`: adds the English picker button label.
- `docs/COMMUNITY_CHAT.md`: documents Emoji capabilities, local data behavior, and the stickers/GIF boundary.
- `progress.md`: records this implementation, validation evidence, affected files, and rollback method.
- Rollback: remove the Emoji component/helper/tests, revert the smile icon, locale key, public composer integration, documentation paragraph, and the two dependencies with `pnpm remove emoji-picker-element emoji-picker-element-data`; then run `pnpm build`. No database or backend rollback is required.

## 2026-08-14 - Task: Improve community chat composing, file drop, and message search

### What was done

- Made the public community-chat textarea grow upward with entered lines, cap at approximately six lines, scroll internally beyond the cap, and shrink after sending or clearing.
- Added drag-and-drop selection for one supported image, MP4, or document across the public chat panel, with a visible drop target and the existing attachment preview and validation. Dropping selects the first file and never sends it automatically.
- Added authenticated, paginated full-history search across undeleted public message content and sender names. Matching is case-insensitive and treats `%` and `_` as literal characters.
- Added a search toolbar, result count, empty/loading/load-more states, and return-to-latest action. Search preserves the previous chat scroll position and is not displaced by real-time incoming messages.
- Kept private messages, message formats, attachment limits, authorization rules, and database structures unchanged.

### Testing

- `go test -race ./internal/service ./internal/repository ./internal/handler -run 'CommunityChat' -count=1`: passed, including normalized search input, literal special-character matching, deleted-message filtering contracts, and existing chat reliability coverage.
- `go test ./internal/service ./internal/repository ./internal/handler ./internal/server/routes -count=1`: passed the complete related backend package suites.
- `pnpm exec vitest run src/api/__tests__/communityChat.spec.ts src/views/user/__tests__/communityChatComposer.spec.ts src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/views/user/__tests__/communityChatEmoji.spec.ts src/views/user/__tests__/communityChatUnread.spec.ts src/views/user/__tests__/communityChatScroll.spec.ts src/components/community/__tests__/CommunityEmojiPicker.spec.ts`: passed 34 tests.
- Targeted ESLint for the changed community-chat API, view, composer helper, tests, and locales: passed.
- `pnpm typecheck`: passed.
- `pnpm build`: passed; Vite emitted only the existing stale Browserslist-data and large shared-chunk warnings.
- Live API checks against the restarted source backend: `/health` returned 200; authenticated keyword search returned the expected one result and pagination; literal `%/_` search returned 200 with zero matches; unauthenticated search returned 401.
- Automated browser visual inspection could not run because the user's saved browser permission blocks access to `http://127.0.0.1:3000`; the restriction was respected and no alternate browser or bypass was attempted.
- `git diff --check`: passed.

### Notes

- `backend/internal/service/community_chat.go`: exposes normalized public-message search and attachment alias normalization through the existing chat service.
- `backend/internal/service/community_chat_test.go`: verifies search trimming, pagination limits, and result normalization.
- `backend/internal/repository/community_chat_repo.go`: queries undeleted public messages by literal, case-insensitive content or sender-name containment.
- `backend/internal/repository/community_chat_repo_test.go`: verifies literal special-character search, ordering, pagination, and result scanning.
- `backend/internal/handler/community_chat_handler.go`: accepts the `search` query on the existing authenticated messages endpoint and rejects combining search with cursors.
- `frontend/src/api/communityChat.ts`: adds the paginated community-message search client.
- `frontend/src/api/__tests__/communityChat.spec.ts`: verifies trimmed search query and pagination parameters.
- `frontend/src/views/user/CommunityChatView.vue`: adds the search experience, drag target, scroll preservation, real-time interaction rules, and adaptive composer wiring.
- `frontend/src/views/user/communityChatComposer.ts`: contains the bounded textarea resize and first-dropped-file helpers.
- `frontend/src/views/user/__tests__/communityChatComposer.spec.ts`: verifies grow, cap, shrink, and dropped-file selection behavior.
- `frontend/src/views/user/__tests__/communityChatEmojiWiring.spec.ts`: verifies search, drag-and-drop, and adaptive input integration without regressing the Emoji picker.
- `frontend/src/i18n/locales/zh/misc.ts`: adds Chinese search and drag-drop labels.
- `frontend/src/i18n/locales/en/misc.ts`: adds matching English labels.
- `docs/COMMUNITY_CHAT.md`: documents adaptive composing, drag-and-drop behavior, history search, and the API query contract.
- `progress.md`: records this task, validation evidence, file inventory, and rollback point.
- Rollback: revert the files listed above to the state before this task and remove `frontend/src/views/user/communityChatComposer.ts` plus its test. No database rollback is required because this task adds no migration or persistent schema change.
- Additional full regression: `pnpm lint:check && pnpm test:run` passed all 235 frontend test files and 1,633 tests; existing test-environment warnings remained non-fatal.

## 2026-08-14 - Task: Add community chat search-result navigation

### What was done

- Added a current-result counter and previous/next arrow controls after a community-chat search so users can inspect matching messages one by one.
- Search starts at the first result, scrolls the selected message into the center of the chat area, and highlights its message bubble. The previous/next controls are disabled at the respective result boundaries.
- Navigating beyond the first loaded result page automatically loads the next search page and continues positioning without requiring a separate “load more” action.
- Kept navigation bound to the last submitted search query when the input text is edited but has not been submitted again, and kept the selected result index valid when a matching message is deleted.

### Testing

- `pnpm exec vitest run src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/api/__tests__/communityChat.spec.ts src/views/user/__tests__/communityChatScroll.spec.ts`: passed 19 tests, including search-navigation wiring and existing chat API/scroll behavior.
- `pnpm lint:check`: passed the complete frontend static analysis.
- `pnpm test:run`: passed all 235 frontend test files and 1,634 tests.
- `pnpm exec eslint src/views/user/CommunityChatView.vue src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/i18n/locales/zh/misc.ts src/i18n/locales/en/misc.ts`: passed after the final changes.
- `pnpm typecheck`: passed.
- `pnpm build`: passed; Vite emitted only the existing stale Browserslist-data and large shared-chunk warnings.
- Automated browser interaction could not access `http://127.0.0.1:3000` because the user's saved browser permission blocks that local address; the restriction was respected and no alternate browser or bypass was used.

### Notes

- `frontend/src/views/user/CommunityChatView.vue`: adds result position state, arrow controls, automatic cross-page navigation, centered scrolling, selected-bubble highlighting, and deletion-safe index maintenance.
- `frontend/src/views/user/__tests__/communityChatEmojiWiring.spec.ts`: verifies previous/next controls, position copy, highlighting, centered scrolling, and automatic result loading.
- `frontend/src/i18n/locales/zh/misc.ts`: adds Chinese result-position and navigation labels.
- `frontend/src/i18n/locales/en/misc.ts`: adds matching English result-position and navigation labels.
- `docs/COMMUNITY_CHAT.md`: documents sequential result navigation, highlighting, and automatic cross-page loading.
- `progress.md`: records this implementation, validation evidence, file inventory, and rollback method.
- Rollback: reverse this task's hunks in the six files listed above. No backend, database, or API rollback is required.

## 2026-08-14 - Task: Show conversation context around community chat search results

### What was done

- Changed community-chat search from rendering only matching messages to rendering a continuous conversation window around the selected match.
- Each selected result now shows up to 20 messages before it, the highlighted matching message, and up to 20 messages after it in their original chronological order.
- Kept search matches as the previous/next navigation index only. Arrow navigation loads the target result's context first and then scrolls to its exact highlighted position; failed context loading keeps the previous result selected.
- Removed the separate result-list pagination button because navigating beyond the loaded match index now fetches the next search page automatically.
- Added request sequencing and close-search state reset so stale context responses cannot replace a newer result or leave the search controls loading.

### Testing

- `pnpm exec vitest run src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/api/__tests__/communityChat.spec.ts src/views/user/__tests__/communityChatScroll.spec.ts`: passed 19 tests, covering context API wiring, search navigation, and existing cursor/scroll behavior.
- `pnpm exec eslint src/views/user/CommunityChatView.vue src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/i18n/locales/zh/misc.ts src/i18n/locales/en/misc.ts`: passed.
- `pnpm typecheck`: passed.
- `pnpm lint:check`: passed the complete frontend static analysis.
- `pnpm test:run`: passed all 235 frontend test files and 1,634 tests.
- `pnpm build`: passed; Vite emitted only the existing stale Browserslist-data and large shared-chunk warnings.

### Notes

- `frontend/src/views/user/CommunityChatView.vue`: separates match indexes from rendered context messages and loads/positions each selected conversation window.
- `frontend/src/views/user/__tests__/communityChatEmojiWiring.spec.ts`: verifies contextual before/after loading, chronological merge, navigation positioning, and removal of the standalone result-list loader.
- `frontend/src/i18n/locales/zh/misc.ts`: removes the no-longer-used Chinese result-list pagination label.
- `frontend/src/i18n/locales/en/misc.ts`: removes the matching English pagination label.
- `docs/COMMUNITY_CHAT.md`: documents contextual search results and the 20-message window on each side.
- `progress.md`: records this implementation, verification evidence, file inventory, and rollback method.
- Rollback: reverse this task's hunks in the six files listed above to restore isolated matching-message rendering. No backend, database, or API rollback is required.

## 2026-08-15 - Task: Prevent floating-point tails from stranding media balance holds

### What was done

- Reproduced a Firefly video balance release failure where the runtime amount `1.32 * 5` became `6.6000000000000005` while PostgreSQL stored the frozen balance as `6.60000000`.
- Quantized media and batch-image hold and actual amounts to the existing eight-decimal billing scale before reserve, capture, or release SQL runs.
- Preserved the existing raw-amount billing fingerprint so requests created before deployment remain idempotency-compatible when retried afterward.
- Kept the change source-only: no database schema, running service, configuration, or user balance was modified.

### Testing

- Before the fix, `go test ./internal/service -run '^TestBatchImageBalanceHoldCommandQuantizesRuntimeCalculatedAmounts$' -count=1` failed with expected `6.6`, actual `6.6000000000000005`.
- `go test ./internal/service -run '^(TestBatchImageBalanceHoldCommandQuantizesRuntimeCalculatedAmounts|TestUsageBillingCommandQuantizesBalanceAndQuotaIdentically|TestQuantizeUsageBillingAmountBoundaries|TestNormalizeQuantizesEveryMonetaryField)$' -count=1`: passed.
- `go test -tags=unit ./internal/repository -run 'Test(UsageBillingRepositoryReleaseBatchImageBalance_QuantizesHoldAmount|ReserveUsageBillingBatchImageBalance|CaptureUsageBillingBatchImageBalance|ReleaseUsageBillingBatchImageBalance)' -count=1`: passed.
- `go test ./internal/service ./internal/repository -count=1`: passed.
- Targeted `go test -race` runs for the service media-billing path and repository hold-release regression passed.
- `gofmt -d` and `git diff --check` for the changed Go files passed.

### Notes

- `backend/internal/service/usage_billing.go`: normalizes hold and actual amounts with the shared `NUMERIC(20,8)` billing quantizer after generating the backward-compatible fingerprint.
- `backend/internal/service/usage_billing_quantize_test.go`: reproduces the runtime multiplication tail and verifies both amount quantization and fingerprint compatibility.
- `backend/internal/repository/usage_billing_repo_unit_test.go`: verifies the release SQL receives `6.6` for a runtime `1.32 * 5` hold.
- `docs/BILLING_INTEGRITY.md`: documents the precision guarantee for reserve, capture, and release operations.
- `progress.md`: records this implementation, verification evidence, file inventory, and rollback point.
- Rollback: reverse only this task's hunks in the five files listed above. No database or data rollback is required because this task adds no migration and changes no persisted balance directly.

## 2026-08-16 - Task: Sync upstream/main through v0.1.177

### What was done

- Fetched and merged all 13 upstream commits in `fbfdcef81..baeac1f3d`, advancing the source version from `0.1.176` to `0.1.177` in merge commit `d860238bcf136cd32e73d114f2a5f26cb884b83a`.
- Preserved the customized worktree through backup branch `backup/pre-upstream-sync-20260816-170750` and stash object `99e4d08261806533cd36e1cbc2a62239331e5cdb`, then restored the original staged, unstaged, and 74 untracked paths with matching object hashes.
- Integrated group-usage daily rollups and configured timezones, native remote compaction v2, Codex turn-state relay and echo protection, session beta features, opt-in fingerprint convergence, Go 1.26.6, and the frontend lockfile security update.
- Resolved the two restore conflicts while retaining both upstream behavior and local media recovery, and corrected the combined-tree Grok 4.6 fallback to keep its independent official `$0.50/MTok` cache rate.
- Added no deployment action, service restart, production database access, or balance mutation.

### Testing

- `go generate ./cmd/server`: passed with no generated diff.
- Targeted compaction, turn-state, billing precision, group-rollup, and migration tests passed; targeted service and repository `go test -race` suites also passed.
- `go test -tags=unit ./...` and the final `go test ./...` from `backend/`: passed completely.
- `pnpm install --frozen-lockfile`, `pnpm lint:check`, `pnpm typecheck`, `pnpm test:run`, and `pnpm build` from `frontend/`: passed; Vitest reported 236 files and 1,636 tests.
- `go test ./internal/web ./cmd/server -count=1`: passed after the frontend production build.
- Root and all deploy Compose files rendered successfully with the required placeholder environment values.
- `git diff --check`, `git diff --cached --check`, strict conflict-marker scanning, `git merge-base --is-ancestor upstream/main HEAD`, staged/unstaged path restoration, and untracked object-hash restoration passed.
- `TEST_DATABASE_URL` is not configured, so the real PostgreSQL trigger integration tests for the new rollup migrations were not run; no claim of live-database validation is made.

### Notes

- `.github/workflows/backend-ci.yml`: updates backend CI to the upstream Go toolchain version.
- `.github/workflows/release.yml`: updates release builds to the upstream Go toolchain version.
- `.github/workflows/security-scan.yml`: updates security scanning to the upstream Go toolchain version.
- `backend/cmd/server/VERSION`: advances the application version to `0.1.177`.
- `backend/go.mod`: advances the declared Go version to `1.26.6`.
- `backend/internal/config/config.go`: adds the configured group-usage rollup timezone setting.
- `backend/internal/config/config_test.go`: verifies rollup timezone configuration behavior.
- `backend/internal/handler/admin/group_handler.go`: returns group usage summaries through the rollup-aware path.
- `backend/internal/handler/openai_gateway_compact_body_signal_test.go`: expands native compaction request-signal coverage.
- `backend/internal/handler/openai_gateway_compact_log_test.go`: updates compact logging expectations for the native path.
- `backend/internal/handler/openai_gateway_handler.go`: separates native remote compaction v2 routing from the legacy compact endpoint.
- `backend/internal/handler/openai_profit_slot_recheck_test.go`: updates profit-slot test setup for current scheduling behavior.
- `backend/internal/pkg/usagestats/usage_log_types.go`: carries group rollup fields in usage statistics types.
- `backend/internal/repository/custom_group_usage_rollup_repo.go`: implements daily group-usage rollup persistence and queries.
- `backend/internal/repository/custom_group_usage_timezone_test.go`: verifies configured-timezone date boundaries.
- `backend/internal/repository/dashboard_aggregation_group_usage_test.go`: covers dashboard aggregation through group rollups.
- `backend/internal/repository/dashboard_aggregation_repo.go`: reads group usage from daily rollups.
- `backend/internal/repository/group_usage_rollup_trigger_integration_test.go`: adds PostgreSQL trigger integration coverage for rollup maintenance.
- `backend/internal/repository/usage_cleanup_repo.go`: keeps rollups consistent when old usage data is cleaned up.
- `backend/internal/repository/usage_cleanup_repo_test.go`: verifies cleanup and rollup consistency.
- `backend/internal/repository/usage_log_repo_group_summary_test.go`: verifies group summary queries against rollup data.
- `backend/internal/repository/usage_log_repo_trend.go`: applies configured timezone boundaries to usage trends.
- `backend/internal/service/account_test_service.go`: uses the native Responses compaction test path.
- `backend/internal/service/account_test_service_openai_compact_test.go`: verifies native compaction while preserving local Gin test isolation.
- `backend/internal/service/custom_group_usage_rollup.go`: exposes group rollup operations to services.
- `backend/internal/service/custom_group_usage_rollup_test.go`: verifies rollup service behavior.
- `backend/internal/service/dashboard_aggregation_service.go`: integrates rollup-backed group totals into dashboard aggregation.
- `backend/internal/service/dashboard_aggregation_service_test.go`: verifies the dashboard rollup integration.
- `backend/internal/service/dashboard_service.go`: wires rollup-aware aggregation into dashboard responses.
- `backend/internal/service/openai_account_scheduler.go`: carries Codex compaction and turn-state scheduling context.
- `backend/internal/service/openai_account_scheduler_compact_test.go`: covers account selection for native compaction requests.
- `backend/internal/service/openai_agent_identity_compat_test.go`: updates identity compatibility expectations for opt-in convergence.
- `backend/internal/service/openai_channel_restriction_test.go`: verifies channel restrictions with the new Codex request paths.
- `backend/internal/service/openai_codex_fingerprint.go`: makes fingerprint convergence opt-in and applies it to passthrough.
- `backend/internal/service/openai_codex_fingerprint_test.go`: verifies opt-in fingerprint behavior across forwarding paths.
- `backend/internal/service/openai_codex_turn_state.go`: implements `x-codex-turn-state` tracking and cross-account echo protection.
- `backend/internal/service/openai_codex_turn_state_test.go`: verifies turn-state relay, isolation, and echo protection.
- `backend/internal/service/openai_compact_body_signal.go`: recognizes native remote compaction v2 body signals.
- `backend/internal/service/openai_compact_probe.go`: probes native compaction capability independently of the legacy endpoint.
- `backend/internal/service/openai_compact_probe_test.go`: verifies native and legacy compaction probe separation.
- `backend/internal/service/openai_compaction_context.go`: carries compaction state through a request lifecycle.
- `backend/internal/service/openai_gateway_forward.go`: forwards Codex session state on the standard gateway path.
- `backend/internal/service/openai_gateway_passthrough.go`: applies turn-state and opt-in fingerprint behavior to passthrough.
- `backend/internal/service/openai_gateway_request_body.go`: includes the new compaction context in request preparation.
- `backend/internal/service/openai_gateway_response_handling.go`: captures upstream Codex turn state from responses.
- `backend/internal/service/openai_gateway_responses_chat_fallback_test.go`: verifies fallback behavior with the new state handling.
- `backend/internal/service/openai_gateway_scheduling.go`: propagates compaction and turn-state metadata through scheduling.
- `backend/internal/service/openai_gateway_service.go`: retains upstream Codex turn-state state alongside the local video compensation worker.
- `backend/internal/service/openai_gateway_service_test.go`: updates gateway construction for the combined service state.
- `backend/internal/service/openai_oauth_passthrough_test.go`: verifies OAuth passthrough fingerprint and session behavior.
- `backend/internal/service/openai_profit_control_pricing_test.go`: updates pricing tests for current scheduling context.
- `backend/internal/service/openai_ws_forwarder_payload.go`: forwards session-level Codex beta features on WebSocket payloads.
- `backend/internal/service/upstream_path_guard_test.go`: verifies the new upstream compaction path remains guarded.
- `backend/internal/service/billing_service.go`: keeps Grok 4.6 on an independent official `$0.50/MTok` cache-rate object in the combined tree.
- `backend/internal/service/billing_service_test.go`: verifies Grok 4.6 official fallback pricing and channel override precedence.
- `backend/migrations/222_group_usage_daily_rollups.sql`: creates and maintains daily group-usage rollups.
- `backend/migrations/223_group_usage_rollup_timezone.sql`: rebuilds rollup boundaries using the configured timezone.
- `backend/migrations/group_usage_rollup_migration_test.go`: verifies migration definitions and ordering.
- `frontend/pnpm-lock.yaml`: updates locked `nanoid` from `3.3.17` to `3.3.18`.
- `frontend/src/api/__tests__/admin.groups.usage-summary.spec.ts`: verifies the group usage-summary API contract.
- `frontend/src/api/admin/groups.ts`: adds rollup-backed group usage summary fields to the client API.
- `frontend/src/components/account/BulkEditAccountModal.vue`: exposes session-level Codex beta feature options in bulk editing.
- `frontend/src/components/account/CreateAccountModal.vue`: exposes session-level Codex beta features during account creation.
- `frontend/src/components/account/EditAccountModal.vue`: exposes session-level Codex beta features during account editing.
- `frontend/src/i18n/locales/en/admin/accounts.ts`: adds English labels for the account beta-feature controls.
- `frontend/src/i18n/locales/en/admin/overview.ts`: adds the English group-usage summary label.
- `frontend/src/i18n/locales/zh/admin/accounts.ts`: adds Chinese labels for the account beta-feature controls.
- `frontend/src/i18n/locales/zh/admin/overview.ts`: adds the Chinese group-usage summary label.
- `frontend/src/views/admin/GroupsView.vue`: displays rollup-backed group usage summaries.
- `frontend/src/views/admin/__tests__/GroupsView.columnSettings.spec.ts`: verifies the new summary column behavior.
- `docs/UPSTREAM_SYNC.md`: records sync scope, migration impact, compatibility decisions, validation evidence, and recovery points.
- `progress.md`: appends this task record without rewriting existing history.
- Rollback: preserve the dirty worktree, then run `git revert -m 1 d860238bcf136cd32e73d114f2a5f26cb884b83a`. The exact pre-sync source point is branch `backup/pre-upstream-sync-20260816-170750`, and the full pre-sync worktree is stash object `99e4d08261806533cd36e1cbc2a62239331e5cdb`; do not drop that stash before deployment acceptance.

## 2026-08-16 - Task: Stabilize community chat search positioning and highlight matches

### What was done

- Replaced page-dependent `scrollIntoView` positioning with an explicit message-list calculation that centers the actual matched keyword inside the visible chat region.
- Added safe, case-insensitive literal highlighting for every displayed match in message text, sender names, document names, and image/video attachment names. The current result uses a stronger visual marker while its message bubble remains outlined.
- Recalibrated only the current search result after its image or video dimensions finish loading, preventing media layout shifts from leaving the selected keyword off target without changing ordinary chat follow behavior.
- Kept search API behavior, result ordering, context radius, message storage, private chat, and database structures unchanged.

### Testing

- Before implementation, the new targeted suite failed because the search helper did not exist and the component still used whole-message `scrollIntoView` without keyword markup.
- `pnpm exec vitest run src/views/user/__tests__/communityChatSearch.spec.ts src/views/user/__tests__/communityChatEmojiWiring.spec.ts`: passed 8 tests after the fix.
- `pnpm exec eslint src/views/user/CommunityChatView.vue src/views/user/communityChatSearch.ts src/views/user/__tests__/communityChatSearch.spec.ts src/views/user/__tests__/communityChatEmojiWiring.spec.ts`: passed.
- `pnpm typecheck` and `pnpm lint:check`: passed.
- `pnpm test:run`: passed all 237 frontend test files and 1,639 tests.
- `pnpm build`: passed; only the existing stale Browserslist-data and large shared-chunk warnings remained.
- The local Vite page loaded successfully, but `/community-chat` redirected to login because no local backend or authenticated session was available. No production account or data was used, so authenticated browser visual validation remains unclaimed.
- `git diff --check`: passed after the final documentation and progress updates.

### Notes

- `frontend/src/views/user/CommunityChatView.vue`: renders literal match markers, centers the preferred keyword within the message list, and recalibrates the current result after media loading.
- `frontend/src/views/user/communityChatSearch.ts`: provides safe literal text segmentation and container-relative center-position calculation.
- `frontend/src/views/user/__tests__/communityChatSearch.spec.ts`: verifies case-insensitive and special-character matches plus container-relative centering.
- `frontend/src/views/user/__tests__/communityChatEmojiWiring.spec.ts`: verifies the view uses keyword-first container scrolling and no longer uses whole-message `scrollIntoView` for search navigation.
- `docs/COMMUNITY_CHAT.md`: documents precise keyword positioning, match highlighting, attachment-name coverage, and media-load recalibration.
- `progress.md`: appends this implementation, validation evidence, file inventory, browser limitation, and rollback point.
- Rollback: reverse only this task's hunks in the six files listed above and remove `frontend/src/views/user/communityChatSearch.ts` plus its test. No backend, API, database, or data rollback is required.

## 2026-08-16 - Task: Improve the contact-owner entry and direct-message composer

### What was done

- Moved the contact-owner action out of the compact right-side control group and into the true center column of the community-chat header on desktop.
- Increased the action to a 48-pixel-high, minimum 176-pixel-wide primary button with a larger mail icon, stronger weight and shadow, while preserving its unread indicator and click behavior.
- Made the direct-message textarea reuse the public composer height calculation, growing with content up to the existing six-line cap and shrinking after send, clear, close, or reopen.
- Kept direct-message APIs, message sending, attachments, unread state, authorization, and database structures unchanged.

### Testing

- Before implementation, the new regression failed because the direct textarea had no element reference or resize event and the contact-owner button remained inside the right-side action group.
- `pnpm exec vitest run src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/views/user/__tests__/communityChatComposer.spec.ts`: passed 8 tests after the fix.
- `pnpm exec eslint src/views/user/CommunityChatView.vue src/views/user/__tests__/communityChatEmojiWiring.spec.ts`: passed.
- `pnpm typecheck` and `pnpm lint:check`: passed.
- `pnpm test:run`: passed all 237 frontend test files and 1,640 tests.
- `pnpm build`: passed; only the existing stale Browserslist-data and large shared-chunk warnings remained.
- Authenticated browser visual validation was not run because the local Vite instance has no backend or login session; no production account or data was used.
- `git diff --check` and `git diff --cached --check`: passed after the final documentation and progress updates.

### Notes

- `frontend/src/views/user/CommunityChatView.vue`: centers and enlarges the contact-owner action and applies the shared adaptive-height behavior to the direct-message textarea.
- `frontend/src/views/user/__tests__/communityChatEmojiWiring.spec.ts`: verifies the private composer resize wiring and the responsive centered-header contract.
- `docs/COMMUNITY_CHAT.md`: documents the prominent contact-owner entry and adaptive private composer behavior.
- `progress.md`: appends this implementation, validation evidence, affected-file inventory, visual-validation boundary, and rollback point.
- Rollback: reverse only this task's hunks in the four files listed above. No backend, API, database, or data rollback is required.

## 2026-08-16 - Task: Extend MiniMax-H3 delayed submission recovery

### What was done

- Extended the finite MiniMax-H3 task discovery window from three minutes to ten minutes, measured from before the upstream create request, so delayed upstream tasks can still be associated and billed.
- Preserved the existing unique-candidate ownership check, persisted recovery behavior, idempotent billing, and final failure release path.
- Added regression coverage for the observed four-minute-thirty-three-second submission delay, retention of the `$18` hold before the final deadline, and release after the final deadline.
- Kept the change source-only: no deployment, service restart, database change, usage-record mutation, or balance mutation was performed.

### Testing

- Before the fix, `go test ./internal/service -run '^(TestPrepareMiniMaxVideoRecoveryAllowsTenMinutesForDelayedSubmission|TestMiniMaxVideoRecoveryWorkerResumesPersistedUniqueCandidateWithoutRebilling|TestMiniMaxVideoRecoveryWorkerReschedulesWhenNoCandidateExists|TestMiniMaxVideoRecoveryWorkerReleasesHoldAfterFinalDeadline|TestMiniMaxVideoRecoveryWorkerCompletesExpiredIdentifiedTaskWithoutTaskListLookup)$' -count=1` failed because prepared recoveries still expired after three minutes.
- The targeted MiniMax-H3 recovery suite passed after the fix, including unique matching, delayed no-candidate rescheduling, final-deadline release, expired identified-task settlement, matched-update retry, and ambiguous-candidate release.
- `go test -race ./internal/service -run '^(TestPrepareMiniMaxVideoRecoveryAllowsTenMinutesForDelayedSubmission|TestMiniMaxVideoRecoveryWorker.*)$' -count=1`: passed.
- `go test ./internal/service ./internal/handler ./internal/repository -count=1`: passed.
- `go test ./... -count=1` from `backend/`: passed completely.
- `gofmt -d`, `git diff --check`, `git diff --cached --check`, and strict conflict-marker scanning for the affected files passed.

### Notes

- `backend/internal/service/openai_minimax_video_recovery.go`: extends the persisted MiniMax-H3 recovery deadline to ten minutes.
- `backend/internal/service/openai_minimax_video_recovery_test.go`: verifies newly prepared recovery records receive the ten-minute deadline.
- `backend/internal/service/openai_video_compensator_test.go`: verifies delayed matching, hold retention before expiry, final release after expiry, and identified-task behavior.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents the ten-minute window, when it starts, and the longer hold-retention tradeoff.
- `progress.md`: records this implementation, verification evidence, file inventory, and rollback point.
- Rollback: reverse only this task's hunks in the five files listed above: restore `openAIVideoRecoveryWindow` to `3 * time.Minute`, remove the new and expanded delayed-recovery assertions, and restore the three-minute documentation paragraph. Do not delete the untracked recovery source files because they contain pre-existing local work. No database or data rollback is required.

## 2026-08-16 - Task: Allow MiniMax-H3 recovery through upstream queue delays

### What was done

- Extended the finite MiniMax-H3 task discovery window from ten minutes to fifteen minutes to cover the observed submission delay plus approximately five minutes of upstream queueing and the next background reconciliation cycle.
- Advanced the delayed-candidate regression to ten minutes and thirty seconds after request start while preserving unique matching, idempotent billing, and final-deadline release behavior.
- Kept the change source-only: no deployment, service restart, database change, usage-record mutation, or balance mutation was performed.

### Testing

- Before the fix, `go test ./internal/service -run '^(TestPrepareMiniMaxVideoRecoveryAllowsFifteenMinutesForDelayedSubmission|TestMiniMaxVideoRecoveryWorkerResumesPersistedUniqueCandidateWithoutRebilling)$' -count=1` failed because prepared recoveries still expired after ten minutes.
- The targeted MiniMax-H3 recovery suite passed after the fix, including delayed unique matching, hold retention, final-deadline release, identified-task settlement, matched-update retry, and ambiguous-candidate release.
- `go test -race ./internal/service -run '^(TestPrepareMiniMaxVideoRecoveryAllowsFifteenMinutesForDelayedSubmission|TestMiniMaxVideoRecoveryWorker.*)$' -count=1`: passed.
- The targeted MiniMax-H3 `504` creation and identified-task status handler tests passed.
- `go test ./... -count=1` from `backend/`: passed completely.
- `gofmt -d`, `git diff --check`, `git diff --cached --check`, and strict conflict-marker scanning for the affected files passed.

### Notes

- `backend/internal/service/openai_minimax_video_recovery.go`: extends the persisted MiniMax-H3 recovery deadline from ten to fifteen minutes.
- `backend/internal/service/openai_minimax_video_recovery_test.go`: verifies newly prepared recoveries receive the fifteen-minute deadline.
- `backend/internal/service/openai_video_compensator_test.go`: verifies a unique upstream task can still be associated ten minutes and thirty seconds after request start.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents upstream queue coverage and the maximum additional hold-retention time.
- `progress.md`: records this follow-up implementation, verification evidence, file inventory, and rollback point.
- Rollback: reverse only this task's hunks in the five files listed above: restore `openAIVideoRecoveryWindow` and its deadline test to ten minutes, restore the delayed-candidate fixture to four minutes and thirty-three seconds, and restore the ten-minute documentation paragraph. Do not delete the untracked recovery source files because they contain pre-existing local work. No database or data rollback is required.

## 2026-08-16 - Task: Replace the Grok Build key configuration template

### What was done

- Replaced the Grok CLI `~/.grok/config.toml` content shown under “Use Key” for Grok-group API keys with the requested Grok 4.6 template.
- Kept the public `/v1` service URL and the selected user's API key dynamically injected instead of hardcoding the provided example values.
- Preserved the existing macOS/Linux and Windows paths, temporary shell environment block, and the separate Claude Code, Codex, and OpenCode configurations.
- Updated the bilingual UI guidance and manual/user documentation to match the new default model, reasoning settings, installer, marketplace, UI, and CLI fields.

### Testing

- Before implementation, the targeted Grok template test failed because the generated configuration did not contain the requested `grok-4.6` model block.
- `pnpm exec vitest run src/components/keys/__tests__/UseKeyModal.spec.ts`: passed all 11 tests.
- Targeted ESLint for the component, test, and locale files passed; `pnpm typecheck` passed.
- `pnpm test:run`: passed all 237 frontend test files and 1,640 tests.
- `pnpm lint:check` and `pnpm build`: passed; the build reported only the existing stale Browserslist-data and large shared-chunk warnings.
- `git diff --check`, `git diff --cached --check`, secret-value scanning, and strict conflict-marker scanning for the affected files passed.

### Notes

- `frontend/src/components/keys/UseKeyModal.vue`: generates the requested Grok 4.6 TOML with the current service URL and API key.
- `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts`: verifies the generated TOML exactly while retaining Windows-path and OpenCode coverage.
- `frontend/src/i18n/locales/zh/dashboard.ts`: updates Chinese Grok CLI save and secret-handling guidance.
- `frontend/src/i18n/locales/en/dashboard.ts`: updates English Grok CLI save and secret-handling guidance.
- `README.md`: replaces the manual Grok Build configuration and smoke-test model with the new template.
- `docs/API_DOCS.md`: documents how Grok-group keys populate the Grok Build template and warns that it contains the usable key.
- `progress.md`: records this implementation, verification evidence, file inventory, and rollback point.
- Rollback: reverse only this task's hunks in the seven files listed above to restore the prior multi-model/env-key Grok template and guidance. Do not restore the entire locale, API-docs, or progress files because they contain pre-existing local work. No backend, database, API-key, or deployment rollback is required.

## 2026-08-16 - Task: Publish the latest multi-architecture china-api image

### What was done

- Built the complete current worktree as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed the multi-architecture image to Docker Hub.
- Published application version `0.1.177`, commit metadata `d860238bc`, and build time `2026-08-16T14:35:33Z`.
- Used `golang:1.26.6-alpine` for the backend build because the current `backend/go.mod` requires Go 1.26.6; no Dockerfile or source file was changed for this override.
- Replaced the previous remote `latest` index `sha256:52e152574f82d97ad5b5b0272975fe90c5604651ccb26e8242369747c9bb9984` with `sha256:92c3a20c07c12107fc97896c8fb21e1ba0d0c937a34e85bfdacd3455c822a1c8`.

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.26.6-alpine --build-arg VERSION=0.1.177 --build-arg COMMIT=d860238bc --build-arg DATE=2026-08-16T14:35:33Z --tag iotwq/china-api:latest --push .`: passed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed index `sha256:92c3a20c07c12107fc97896c8fb21e1ba0d0c937a34e85bfdacd3455c822a1c8`, amd64 manifest `sha256:bba7806b5bf8c0574946727b71b87e34d9a1c646c9d0dfa2f57041c65477811c`, and arm64 manifest `sha256:42dec5cfaccfba4f00a00c76b5c6cd4bfd44b0ea434ae7b4b7e29576c391e7d7`.
- Fresh Docker Hub pulls and `--version` execution passed for both architectures; each reported `Sub2API 0.1.177 (commit: d860238bc, built: 2026-08-16T14:35:33Z)`. The first amd64 tag pull hit a transient Docker Hub `EOF`; the digest-based retry completed successfully.
- The existing full backend and frontend validation for this worktree passed before publication: backend `go test ./...`, 237 frontend test files with 1,640 tests, frontend typecheck, lint, and production build.
- `git diff --check` and `git diff --cached --check`: passed before the publication record was appended.

### Notes

- `progress.md`: records the image publication, immutable digests, validation evidence, Go toolchain override, and rollback point.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:52e152574f82d97ad5b5b0272975fe90c5604651ccb26e8242369747c9bb9984` to restore the previous multi-architecture image index. This changes only the remote tag and does not alter repository files, databases, or running containers.

## 2026-08-17 - Task: Keep MiniMax-H3 recovery alive through long upstream queues

### What was done

- Extended the finite MiniMax-H3 task-discovery window from fifteen minutes to thirty minutes to cover the observed three-minute prompt-processing delay, three-minute prompt optimization, thirteen-minute queue, and three-minute generation timeline.
- Kept local status responses in `processing` while the recovery remains active, so callers do not receive a premature failure at the former fifteen-minute deadline.
- Preserved unique-candidate ownership checks, idempotent billing, recovery after service restart, unlimited follow-up once a real upstream task ID is identified, and final hold release when no task can be confirmed by the new deadline.
- Kept this change source-only: no image publication, deployment, service restart, database mutation, balance mutation, or historical failed-task recovery was performed.

### Testing

- Before the fix, `go test ./internal/service -run '^(TestPrepareMiniMaxVideoRecoveryAllowsThirtyMinutesForDelayedSubmission|TestMiniMaxVideoRecoveryWorkerResumesCandidateAfterTwentyTwoMinutesWithoutRebilling)$' -count=1` failed: the prepared deadline was still fifteen minutes and a unique candidate appearing after twenty-two minutes and thirty seconds was marked `failed` instead of `matched`.
- `go test ./internal/service -run '^(TestPrepareMiniMaxVideoRecoveryAllowsThirtyMinutesForDelayedSubmission|TestMiniMaxVideoRecoveryWorker.*)$' -count=1`: passed after the fix.
- The same targeted recovery suite passed with `-race`, covering delayed unique matching, idempotent billing, identified-task settlement after deadline, transient matched-update retry, no-candidate rescheduling, final release, and ambiguous-candidate release.
- Targeted MiniMax-H3 `504` creation, hold ownership, pricing rejection, recovery repository, and video-forwarding tests passed.
- `go test ./... -count=1` from `backend/`: passed completely.
- `gofmt -d`, `git diff --check`, and `git diff --cached --check`: passed before the final progress entry.

### Notes

- `backend/internal/service/openai_minimax_video_recovery.go`: extends the task-discovery deadline from fifteen to thirty minutes.
- `backend/internal/service/openai_minimax_video_recovery_test.go`: verifies newly prepared recovery records receive the thirty-minute deadline.
- `backend/internal/service/openai_video_compensator_test.go`: verifies a unique upstream task appearing after twenty-two minutes and thirty seconds is still matched without duplicate billing.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents the observed long-queue timeline, processing behavior, and longer hold-retention tradeoff.
- `progress.md`: records this implementation, validation evidence, affected-file inventory, deployment boundary, and rollback point.
- Rollback: reverse only this task's hunks in the five files listed above: restore `openAIVideoRecoveryWindow` and its deadline test to fifteen minutes, restore the delayed-candidate fixture and test name to the prior ten-minute-thirty-second scenario, and restore the fifteen-minute documentation paragraph. Do not delete the untracked recovery files because they contain pre-existing local work. No database or data rollback is required.

## 2026-08-17 - Task: Publish the MiniMax-H3 recovery update to china-api latest

### What was done

- Built the complete current worktree, including the thirty-minute MiniMax-H3 recovery update, as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed it to Docker Hub.
- Published application version `0.1.177`, commit metadata `d860238bc`, and build time `2026-08-16T16:14:21Z`.
- Used `golang:1.26.6-alpine` because the current `backend/go.mod` requires Go 1.26.6; no Dockerfile or source file was changed for the build override.
- Replaced the previous remote `latest` index `sha256:92c3a20c07c12107fc97896c8fb21e1ba0d0c937a34e85bfdacd3455c822a1c8` with `sha256:12d214a9614758f222da13b620c8e0f33a7f7459f7ec50a4893b62bee34eb6e5`.

### Testing

- The source worktree passed backend `go test ./... -count=1`, targeted MiniMax-H3 recovery tests, and the recovery race suite before publication.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.26.6-alpine --build-arg VERSION=0.1.177 --build-arg COMMIT=d860238bc --build-arg DATE=2026-08-16T16:14:21Z --tag iotwq/china-api:latest --push .`: passed; the embedded frontend and both backend binaries built successfully.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed index `sha256:12d214a9614758f222da13b620c8e0f33a7f7459f7ec50a4893b62bee34eb6e5`, amd64 manifest `sha256:62b72e72348f2574dc24cafb4ff3baa182a9b7d13e14bcd18bf62a993e8e282b`, and arm64 manifest `sha256:4d0007a0d111b214dcab6b245f83cb290f4239a579051d5874cd1e4c52e0e452`.
- Fresh Docker Hub pulls and `--version` execution passed for both architectures; each reported `Sub2API 0.1.177 (commit: d860238bc, built: 2026-08-16T16:14:21Z)`.
- `git diff --check` and `git diff --cached --check`: passed before publication.

### Notes

- `progress.md`: records the image publication, immutable digests, validation evidence, Go toolchain override, and rollback point.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:92c3a20c07c12107fc97896c8fb21e1ba0d0c937a34e85bfdacd3455c822a1c8` to restore the previous multi-architecture image index. This changes only the remote tag and does not alter repository files, databases, or running containers.

## 2026-08-17 - Task: Recover MiniMax-H3 tasks when upstream media counts are unavailable

### What was done

- Changed delayed MiniMax-H3 task matching so missing media-count fields and the observed unreliable zero values are treated as unknown instead of as exact zero counts. A trustworthy positive upstream count still rejects a request with a different count.
- Kept model, resolution, duration, and ratio matching strict, and preserved the submission baseline, unique-candidate requirement, original-account routing, idempotent billing, and final failure release behavior.
- Added worker coverage for a delayed `2`-image and `1`-audio request moving from `pending` through `identified` to `matched` exactly once when the upstream list reports `input_image_count=0` and omits audio count.
- Added delivery coverage proving that a recovered local task ID queries the bound upstream task ID, rewrites status output back to the local ID, and downloads the completed MP4 through the local content endpoint.
- Kept this change source-only: no database schema or historical record was changed, no balance was mutated, and no deployment or image publication was performed.

### Testing

- Before the production fix, the new unavailable-media-count regression failed with `got 0 candidates, want 1`; the known-positive mismatch test and the existing baseline and multiple-candidate protections passed.
- `go test ./internal/service -run 'MiniMaxVideoRecovery|PrepareMiniMaxVideoRecovery' -count=1`: passed.
- `go test -race ./internal/service -run 'MiniMaxVideoRecovery' -count=1`: passed.
- `go test ./internal/handler -run 'Video.*Recovery|VideosMiniMax|RecoveredVideo' -count=1`: passed, including status-ID rewriting and MP4 content download.
- `go test ./internal/repository -run 'OpenAIVideo.*Recovery|OpenAIVideoTaskBinding' -count=1`: passed.
- `go test ./... -count=1` from `backend/`: passed completely.
- `gofmt -d`, `git diff --check`, `git diff --cached --check`, targeted untracked-file whitespace checks, and strict conflict-marker scanning for the affected files produced no diagnostics.

### Notes

- `backend/internal/service/openai_minimax_video_recovery.go`: tracks whether each upstream media count is trustworthy and only applies positive known counts as candidate filters.
- `backend/internal/service/openai_minimax_video_recovery_test.go`: covers unknown zero or missing counts and known positive count mismatches while retaining baseline and multiple-candidate coverage.
- `backend/internal/service/openai_video_compensator_test.go`: verifies delayed unknown-count recovery, single compensation arming, and idempotent repeated worker execution.
- `backend/internal/handler/openai_videos_recovery_billing_test.go`: verifies recovered status routing, local task-ID preservation, and MP4 delivery through the local content endpoint.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents strict core matching and the known-versus-unknown media-count rule.
- `progress.md`: records the implementation, regression evidence, verification, file inventory, and rollback point.
- Rollback: reverse only this task's hunks in the six files listed above: restore exact image, video, and audio count comparison in the recovery matcher, remove the ephemeral known-count flags and positive-count parser, remove the three new regression tests and their test-stub support, and restore the prior recovery-matching documentation paragraph. Do not delete the untracked recovery files because they contain pre-existing local work. No database, balance, deployment, or image rollback is required.

## 2026-08-17 - Task: Publish the MiniMax-H3 media-count recovery to china-api latest

### What was done

- Built the complete current worktree, including the MiniMax-H3 unknown-media-count recovery and recovered-video delivery coverage, as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed the multi-architecture image to Docker Hub.
- Published application version `0.1.177`, source baseline commit metadata `d860238bc`, and build time `2026-08-16T17:47:12Z`.
- Used `golang:1.26.6-alpine` because the current `backend/go.mod` requires Go 1.26.6; no Dockerfile or source file was changed for this build override.
- Replaced the previous remote `latest` index `sha256:12d214a9614758f222da13b620c8e0f33a7f7459f7ec50a4893b62bee34eb6e5` with `sha256:cdff222246e39361d147e989994fafb93f75f01743b35b1f575d871388dea011`.

### Testing

- The current source had already passed `go test ./... -count=1` from `backend/` in the immediately preceding MiniMax-H3 implementation task; before publication, `go test ./internal/service -run 'MiniMaxVideoRecovery|PrepareMiniMaxVideoRecovery' -count=1` passed again.
- `git diff --check` and `git diff --cached --check` passed before the image build.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.26.6-alpine --build-arg VERSION=0.1.177 --build-arg COMMIT=d860238bc --build-arg DATE=2026-08-16T17:47:12Z --tag iotwq/china-api:latest --push .`: passed; the embedded frontend and both backend binaries built successfully.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed index `sha256:cdff222246e39361d147e989994fafb93f75f01743b35b1f575d871388dea011`, amd64 manifest `sha256:22fb859ab583c790d09b0bf8661d52c5f307762b4b0c8616d76db0283af6d1e2`, and arm64 manifest `sha256:cadf5ffdd08c8eb79c22974ea4083e8990960d045aa9aca380a295328cb9ee05`.
- Fresh Docker Hub pulls and `--version` execution passed for both architectures; each reported `Sub2API 0.1.177 (commit: d860238bc, built: 2026-08-16T17:47:12Z)`.

### Notes

- `progress.md`: records the image publication, immutable digests, validation evidence, build metadata, and rollback point.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:12d214a9614758f222da13b620c8e0f33a7f7459f7ec50a4893b62bee34eb6e5` to restore the previous multi-architecture image index. This changes only the remote tag and does not alter repository files, databases, balances, or running containers.

## 2026-08-17 - Task: Restore failed OpenAI video refunds for historical request IDs

### What was done

- Stopped ordinary OpenAI-compatible video charges from being rewritten as `grok-video:openai-video:<task_id>`; Grok stabilization now runs only when the result already carries the explicit `grok-video:` marker.
- Made failed-video refunds query both the current `openai-video:<billing_task_id>` and historical `grok-video:openai-video:<billing_task_id>` forms, always scoped to the current API Key.
- Refused automatic refund when both forms exist, while retaining `openai-video-refund:<billing_task_id>` as the atomic idempotency key so repeated compensation cannot refund twice.
- Kept the change source-only. Usage row `2897732`, the reported USD 19.80 charge, and the 42 production compensation records were not modified or replayed; no image was built or deployed.

### Testing

- Before the production fix, `go test ./internal/service -run '^TestOpenAIVideoCompensatorRefundsTerminalFailureOnce$' -count=1` failed after the historical double-prefixed charge could not be found: expected compensation status `refunded`, got `pending`.
- Before the production fix, `go test ./internal/service -run '^TestOpenAIGatewayServiceRecordUsage_ChannelPerRequestVideoBillingUsesTaskID$' -count=1` failed: expected `openai-video:task-video-123`, got `grok-video:openai-video:task-video-123`.
- `go test ./internal/service -run 'RefundFailedOpenAIVideoTask|OpenAIVideoCompensatorRefundsTerminalFailureOnce|ChannelPerRequestVideoBillingUsesTaskID|GrokVideo' -count=1`: passed after the fix, covering current and historical charge IDs, USD 19.80 reversal, ambiguity rejection, refund idempotency, ordinary video IDs, and native Grok IDs.
- `go test -race ./internal/service -run 'RefundFailedOpenAIVideoTask|OpenAIVideoCompensator' -count=1`: passed.
- `go test ./internal/repository -run 'UsageBilling|UsageLog' -count=1`: passed.
- `go test ./... -count=1` from `backend/`: passed the complete backend suite.
- `gofmt -d` on the four affected Go files, `git diff --check`, `git diff --cached --check`, and strict conflict-marker scanning produced no diagnostics.

### Notes

- `backend/internal/service/openai_gateway_usage.go`: limits stable Grok video request-ID handling to explicitly marked Grok results.
- `backend/internal/service/openai_videos.go`: looks up both exact current and historical charge IDs, rejects ambiguous matches, and preserves refund retry and idempotency behavior.
- `backend/internal/service/openai_gateway_record_usage_test.go`: covers ordinary video ID preservation, native Grok ID preservation, current-format refunds, and ambiguous dual records.
- `backend/internal/service/openai_video_compensator_test.go`: reproduces and verifies the historical double-prefixed USD 19.80 compensation path.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents request-ID ownership, historical refund compatibility, API-Key scoping, and ambiguity handling.
- `progress.md`: records this implementation, red-to-green evidence, production-data boundary, file inventory, and rollback instructions.
- Rollback: reverse only this task's hunks in the six files listed above to restore unconditional video request-ID stabilization and single-form refund lookup. Do not delete the untracked video service, compensator test, or media compatibility document because they contain pre-existing local work. No database, balance, production-compensation, deployment, or image rollback is required.

## 2026-08-17 - Task: Publish the failed-video refund fix to china-api latest

### What was done

- Built the complete current worktree, including the current and historical video charge-ID refund compatibility fix, as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed the multi-architecture image to Docker Hub.
- Published application version `0.1.177`, source baseline commit metadata `d860238bc`, and build time `2026-08-16T18:52:02Z`.
- Used `golang:1.26.6-alpine` because the current `backend/go.mod` requires Go 1.26.6; no Dockerfile or source file was changed for the build override.
- Replaced the previous remote `latest` index `sha256:cdff222246e39361d147e989994fafb93f75f01743b35b1f575d871388dea011` with `sha256:1f6a425d2fd8c9a4bf4ceb2b4e48989e6eb83662a25570956f69b846680effbf`.

### Testing

- The immediately preceding source task passed `go test ./... -count=1` from `backend/`, targeted current/historical refund tests, the refund race suite, repository billing tests, and formatting/whitespace checks.
- `git diff --check` and `git diff --cached --check` passed before the image build.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.26.6-alpine --build-arg VERSION=0.1.177 --build-arg COMMIT=d860238bc --build-arg DATE=2026-08-16T18:52:02Z --tag iotwq/china-api:latest --push .`: passed; both binaries compiled and the remote index was committed only after all layers and child manifests were uploaded.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed remote index `sha256:1f6a425d2fd8c9a4bf4ceb2b4e48989e6eb83662a25570956f69b846680effbf`, amd64 manifest `sha256:0cc3cc49ba322716661f6e0c8593c521f61b2966a793c63b6f2473b7fc266469`, and arm64 manifest `sha256:7cbacd9ecea2f2c27e6f2deb7733c7a7ad0ae72f62f5c2f65c21cb5ee737a169`.
- Fresh Docker Hub pulls and `--version` execution passed for both exact child manifests; each reported `Sub2API 0.1.177 (commit: d860238bc, built: 2026-08-16T18:52:02Z)`. The first arm64 pull hit a Docker Hub OAuth `EOF`, and the second hit a CloudFront blob `EOF`; a digest-pinned retry completed successfully before runtime validation.

### Notes

- `progress.md`: records the image publication, immutable digests, runtime evidence, transient registry-download retries, and rollback point; no business source or deployment configuration was changed by this publication task.
- No database, balance, compensation record, running container, or deployed service was modified. The publication only replaced the Docker Hub `latest` tag.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:cdff222246e39361d147e989994fafb93f75f01743b35b1f575d871388dea011` to restore the previous multi-architecture image index without changing repository files or production data.

## 2026-08-18 - Task: Sync upstream/main through v0.1.178

### What was done

- Merged 107 upstream commits from `baeac1f3d..49504adc9` into `main` as merge commit `297b246fda9064648bee6772d789779314d14c52`, advancing the application version from `0.1.177` to `0.1.178`.
- Preserved the complete pre-sync worktree in `backup/pre-upstream-sync-20260818-223733` and stash `76a1544435ca56b496bbcec8b203352300bd715b`; the stash was not dropped.
- Restored the original 264 staged paths, 23 unstaged paths, and 76 untracked paths with exact path-set checks and exact untracked object-hash checks. The one additional unstaged path is the sync-specific Grok source-shape test correction.
- Resolved five overlapping files by retaining both upstream channel time pricing and local video billing, strict pricing failures, OpenAI video cost snapshots, and video-specific pricing form templates.
- Integrated migrations `224`, both distinct `225` files, and `226` without renaming or touching a running database.

### Testing

- `go generate ./cmd/server`: passed twice through the package directives and produced no additional diff.
- `go test ./internal/handler/admin ./internal/repository ./internal/service -count=1`: passed.
- `go test ./... -count=1` from `backend/`: passed the complete backend suite.
- `go test -race ./internal/service -run 'TimePricing|OpenAIVideoBillingSnapshot|RefundFailedOpenAIVideoTask|OpenAIVideoCompensator|RecordUsage_MissingPricing|ValidateOpenAITokenPricing' -count=1`: passed.
- `pnpm install --frozen-lockfile`, `pnpm lint:check`, `pnpm typecheck`, and `pnpm build` from `frontend/`: passed.
- Pricing-entry targeted Vitest run: 4 files and 39 tests passed. Full `pnpm exec vitest run --reporter=dot`: 242 files and 1713 tests passed after correcting the stale Grok source-shape assertion.
- `go test ./internal/web ./internal/server ./cmd/server -count=1` after the production frontend build: passed.
- Root, deploy, local, dev, and standalone Compose configurations rendered successfully with validation-only required environment values.
- `git diff --check`, `git diff --cached --check`, strict conflict-marker scanning, `git merge-base --is-ancestor upstream/main HEAD`, staged/unstaged path-set comparison, and untracked object-hash comparison: passed.

### Notes

- The upstream merge changes 301 files; the exact committed inventory is available with `git diff --name-status baeac1f3d..49504adc9`.
- `backend/internal/handler/admin/channel_handler.go`: combines the upstream time-pricing request shape with the local `video` billing-mode validation.
- `backend/internal/service/channel.go`: combines upstream time-pricing domain fields with local video pricing semantics and comments.
- `backend/internal/service/gateway_usage_billing.go`: carries request-level pricing time through cost calculation while preserving strict pricing error returns.
- `backend/internal/service/openai_gateway_usage.go`: preserves video cost snapshots and strict validation while adapting both calculation paths to the new pricing-time signatures.
- `frontend/src/components/admin/channel/PricingEntryCard.vue`: retains local video templates and clears time-pricing periods whenever billing mode changes.
- `frontend/src/components/account/__tests__/CreateAccountModal.grok.spec.ts`: checks the current switch-based xAI API key placeholder without changing account behavior.
- `docs/UPSTREAM_SYNC.md`: records the v0.1.178 scope, migrations, compatibility choices, validation, backup, and rollback point.
- `progress.md`: records this sync implementation, verification evidence, direct file inventory, and rollback instructions.
- Rollback: first preserve the current worktree, then run `git revert -m 1 297b246fda9064648bee6772d789779314d14c52`. The pre-sync branch and stash above remain available; no database or running service rollback was performed by this task.

## 2026-08-19 - Task: Add an independent domestic optimized API address

### What was done

- Added `optimized_api_base_url` as an independently persisted and publicly exposed setting without changing the existing `api_base_url` behavior or database schema.
- Added a separate admin input for the domestic optimized route and kept the existing API endpoint field as the sole source for Use Key, model chat, CC Switch, and callback URL logic.
- Updated the API Keys page to label the primary route as `直连地址`, label the optional second route as `国内优化地址`, keep the default badge only on the primary route, and provide route-specific copy and speed-test controls.
- Kept the optimized route hidden when unset and adjusted the endpoint chips so labels remain intact and URLs wrap without overflow on narrow screens.

### Testing

- Red phase: `go test -tags unit ./internal/service ./internal/handler -run 'OptimizedAPIBaseURL' -count=1` failed because the setting key and service/public fields did not yet exist.
- Red phase: the three targeted frontend files produced three expected failures for the missing optimized address label, missing KeysView prop, and missing admin input.
- `go test -tags unit ./internal/service ./internal/handler -run 'OptimizedAPIBaseURL' -count=1`: passed.
- `go test -tags unit ./internal/server -run '^TestAPIContracts$' -count=1`: passed after adding the new field to both strict admin settings contracts.
- `go test -tags unit ./internal/service ./internal/handler ./internal/handler/admin ./internal/server -count=1`: passed.
- Targeted Vitest run for `EndpointPopover`, `KeysView`, and `SettingsView`: 3 files and 50 tests passed.
- Full `pnpm exec vitest run --reporter=dot`: 242 files and 1716 tests passed.
- `pnpm lint:check`, `pnpm typecheck`, and `pnpm build`: passed.
- Browser QA with the real endpoint component passed at 1280 px and 390 px widths: both routes were visible, only the direct route showed `默认`, each speed-test link targeted its own encoded address, labels stayed on one line, and neither chip overflowed the viewport.
- `git diff --check` and `git diff --cached --check`: passed.

### Notes

- `.gitignore`: allows the endpoint-address behavior document to be tracked under the repository's ignored-by-default `docs/` directory.
- `backend/internal/service/domain_constants.go`: defines the independent optimized-address setting key.
- `backend/internal/service/settings_view.go`: carries the optimized address in admin and public service views.
- `backend/internal/service/setting_parse.go`: reads the optimized address from persisted settings.
- `backend/internal/service/setting_update.go`: persists the optimized address independently from the primary address.
- `backend/internal/service/setting_public.go`: loads and injects the optimized address through the public settings contract.
- `backend/internal/service/setting_service_public_test.go`: verifies public service exposure of the optimized address.
- `backend/internal/service/setting_service_update_test.go`: verifies independent persistence of the optimized address.
- `backend/internal/handler/dto/settings.go`: adds the optimized address to admin and public response DTOs.
- `backend/internal/handler/setting_handler.go`: maps the optimized address into the public response.
- `backend/internal/handler/setting_handler_public_test.go`: verifies the public HTTP response field.
- `backend/internal/handler/admin/setting_handler.go`: maps the optimized address into the admin settings response.
- `backend/internal/handler/admin/setting_handler_update.go`: accepts, saves, and returns the optimized address on admin updates.
- `backend/internal/handler/admin/setting_handler_audit.go`: records optimized-address changes in the settings audit.
- `backend/internal/server/api_contract_test.go`: covers populated and empty optimized-address admin contracts.
- `frontend/src/types/index.ts`: adds the optimized address to public settings typing.
- `frontend/src/api/admin/settings.ts`: adds the optimized address to admin read and update typing.
- `frontend/src/stores/app.ts`: provides an empty optimized address in the cached-settings fallback.
- `frontend/src/views/admin/SettingsView.vue`: adds the manual optimized-address input and save payload.
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`: verifies loading and submitting both addresses independently.
- `frontend/src/views/user/KeysView.vue`: passes the optimized address only to the endpoint display component.
- `frontend/src/views/user/__tests__/KeysView.spec.ts`: verifies the dedicated display prop without replacing the primary address.
- `frontend/src/components/keys/EndpointPopover.vue`: renders the direct and optimized routes with route-specific speed tests and responsive wrapping.
- `frontend/src/components/keys/__tests__/EndpointPopover.spec.ts`: verifies labels, the single default badge, and the optimized speed-test URL.
- `frontend/src/i18n/locales/zh/admin/settings.ts`: adds Chinese admin copy for the optimized address and its scope.
- `frontend/src/i18n/locales/en/admin/settings.ts`: adds the English admin copy.
- `frontend/src/i18n/locales/zh/dashboard.ts`: renames the primary user label and adds the optimized label in Chinese.
- `frontend/src/i18n/locales/en/dashboard.ts`: renames the primary user label and adds the optimized label in English.
- `docs/API_ENDPOINT_ADDRESSES.md`: documents the two-address contract and strict behavior boundary.
- `progress.md`: records implementation, verification, file inventory, and rollback instructions.
- Rollback: while this task remains the only unstaged change in the listed product files, run `git restore --worktree -- .gitignore backend/internal/service/domain_constants.go backend/internal/service/settings_view.go backend/internal/service/setting_parse.go backend/internal/service/setting_update.go backend/internal/service/setting_public.go backend/internal/service/setting_service_public_test.go backend/internal/service/setting_service_update_test.go backend/internal/handler/dto/settings.go backend/internal/handler/setting_handler.go backend/internal/handler/setting_handler_public_test.go backend/internal/handler/admin/setting_handler.go backend/internal/handler/admin/setting_handler_update.go backend/internal/handler/admin/setting_handler_audit.go backend/internal/server/api_contract_test.go frontend/src/types/index.ts frontend/src/api/admin/settings.ts frontend/src/stores/app.ts frontend/src/views/admin/SettingsView.vue frontend/src/views/admin/__tests__/SettingsView.spec.ts frontend/src/views/user/KeysView.vue frontend/src/views/user/__tests__/KeysView.spec.ts frontend/src/components/keys/EndpointPopover.vue frontend/src/components/keys/__tests__/EndpointPopover.spec.ts frontend/src/i18n/locales/zh/admin/settings.ts frontend/src/i18n/locales/en/admin/settings.ts frontend/src/i18n/locales/zh/dashboard.ts frontend/src/i18n/locales/en/dashboard.ts`, remove `docs/API_ENDPOINT_ADDRESSES.md`, and delete only this final `progress.md` section. This returns tracked product files to their pre-task index state without touching existing staged work, databases, balances, or running services.

## 2026-08-19 - Task: Document image2 response format compatibility

### What was done

- Added `response_format` to the user-facing `gpt-image-2` parameter table with `b64_json` and `url` as the documented values, while requiring clients to tolerate either response field.
- Added a JavaScript response parser example covering raw Base64, a data URL in `data[].url`, and an HTTP(S) address in `data[].url`, with all three paths normalized to image `Blob` values.
- Kept the scope limited to these two requested documentation additions and synchronized the repository API documentation summary.

### Testing

- Red phase: `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts` failed on the missing `response_format` text before implementation.
- `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts`: passed, 1 file and 1 test.
- `pnpm typecheck`: passed.
- `pnpm lint:check`: passed.
- `pnpm build`: passed; Vite transformed 1476 modules and produced the production frontend bundle.
- Local browser rendering was attempted at `http://127.0.0.1:4176/api-docs`, but direct navigation was unavailable: both browser surfaces returned `ERR_BLOCKED_BY_CLIENT`, while the Vite development proxy treated `/api-docs` as a backend path and the backend was not running. Visual screenshot verification was therefore unavailable. The mounted component test verified the rendered parameter text, all three parser branches, and the third code card.
- `git diff --check`: passed.

### Notes

- `frontend/src/views/user/ApiDocsView.vue`: adds the `response_format` row, three-branch JavaScript parser example, and matching response note.
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: verifies the new parameter, parser branches, and third image2 code card.
- `docs/API_DOCS.md`: synchronizes the public image2 response-format contract and parser coverage.
- `progress.md`: records this implementation, verification evidence, visual-test limitation, and rollback point.
- Rollback: apply the inverse of this task's isolated hunks only: remove `gptImage2ResponseHandlingJavascript`, the `response_format` parameter row and third image2 example, restore the previous single-line response note and API-doc bullet, restore the image2 code-card assertion from 3 to 2, remove the four added response assertions, and delete only this final `progress.md` section. Do not remove the four files because they predated this task as user-owned untracked or staged work.

## 2026-08-20 - Task: Restore Firefly video status routing after MiniMax recovery changes

### What was done

- Corrected the MiniMax recovery-route classification so the database default `recovery_status=inactive` remains an ordinary Firefly task binding instead of being treated as a MiniMax-H3 recovery task.
- Restored Firefly status and content forwarding through the original bound account and the standard `/v1/videos/{task_id}` paths, while leaving MiniMax pending, identified, matched, failure, billing, timeout, and unique-task recovery behavior unchanged.
- Kept the fix source-only: no database schema, task row, account state, usage record, balance, running service, or Docker image was modified.

### Testing

- Red phase: `go test ./internal/handler -run '^TestFireflyVideoStatusAndContentDoNotUseInactiveMiniMaxRecoveryRoute$' -count=1` reproduced HTTP 503 `No available compatible accounts` before any upstream call.
- Green phase: the same test passed after the fix and verified `completed` status, MP4 content delivery, and exact upstream paths `/v1/videos/firefly-task-1` and `/v1/videos/firefly-task-1/content`.
- `go test ./internal/handler -run 'Video(Status|.*Recovery|sMiniMax|Firefly)' -count=1`: passed.
- `go test ./internal/service -run 'MiniMaxVideoRecovery|PrepareMiniMaxVideoRecovery|RequiredTaskAccount' -count=1`: passed.
- `go test ./internal/repository -run 'OpenAIVideo.*Recovery|OpenAIVideoTaskBinding' -count=1`: passed.
- `go test ./... -count=1` from `backend/`: passed the complete backend suite.
- `gofmt -d`, tracked and untracked whitespace checks, staged whitespace checks, and strict conflict-marker scanning produced no diagnostics.

### Notes

- `backend/internal/service/openai_minimax_video_recovery.go`: defines `inactive` as the explicit non-recovery state shared with the handler.
- `backend/internal/handler/openai_videos.go`: excludes `inactive` bindings from MiniMax-H3 capability routing.
- `backend/internal/handler/openai_videos_recovery_billing_test.go`: reproduces the Firefly 503 regression and verifies status plus content routing.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents the boundary between ordinary Firefly bindings and MiniMax recovery tasks.
- `progress.md`: appends this implementation, validation evidence, file inventory, and rollback instructions.
- Rollback: reverse only this task's isolated hunks in the five files above: restore the non-empty-only recovery check, remove the `OpenAIVideoRecoveryInactive` constant and Firefly regression test/stub, remove the added Firefly routing paragraph, and delete only this final `progress.md` section. The currently published pre-fix image remains available at `iotwq/china-api@sha256:08a46db87463f38e7282c4b2e7ca2e86f549149b053562f0ea1d7189692689c2`; no database or data rollback is required.
- Configuration boundary: a normal successful MiniMax task and an ordinary Firefly task both retain an `inactive` real-task binding, so protocol selection still uses the account's MiniMax endpoint capability. Firefly accounts must not enable that capability; MiniMax remains on a dedicated account as documented. Persisting a per-task protocol would require a separately approved schema and migration change.

## 2026-08-20 - Task: Publish the Firefly video routing fix to china-api latest

### What was done

- Built the complete current worktree, including the Firefly `inactive` task-routing correction, as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed the multi-architecture image to Docker Hub.
- Published application version `0.1.178`, source baseline commit metadata `297b246fd`, and build time `2026-08-19T16:47:03Z`.
- Replaced the previous remote `latest` index `sha256:08a46db87463f38e7282c4b2e7ca2e86f549149b053562f0ea1d7189692689c2` with `sha256:478e69768b611910445f358c688e27138ec2a2a9d3d42b1d07ecaea9b43e36ec`.

### Testing

- The immediately preceding Firefly routing task passed `go test ./... -count=1` from `backend/`; before publication, `go test ./internal/handler -run '^TestFireflyVideoStatusAndContentDoNotUseInactiveMiniMaxRecoveryRoute$' -count=1` passed again.
- `git diff --check` and `git diff --cached --check` passed before the image build.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.26.6-alpine --build-arg VERSION=0.1.178 --build-arg COMMIT=297b246fd --build-arg DATE=2026-08-19T16:47:03Z --tag iotwq/china-api:latest --push .`: passed; the frontend and both backend binaries built successfully.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed remote index `sha256:478e69768b611910445f358c688e27138ec2a2a9d3d42b1d07ecaea9b43e36ec`, amd64 manifest `sha256:8f127a78e0f9b6042b5830ee479913fbe7ada6f86906c9626183ef0b9bf7c273`, and arm64 manifest `sha256:5fc627cf7fe407d68d26881eb235fc2ed1b713c5e07a04f8e338ba86ffe04955`.
- Fresh Docker Hub pulls and `--version` execution passed for both architectures; each reported `Sub2API 0.1.178 (commit: 297b246fd, built: 2026-08-19T16:47:03Z)`.

### Notes

- `progress.md`: records the multi-architecture publication, immutable digests, runtime verification, and rollback point; no business source, database, balance, production task, or running service was modified by this publication step.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:08a46db87463f38e7282c4b2e7ca2e86f549149b053562f0ea1d7189692689c2` to restore the previous multi-architecture image index without changing repository files or production data.

## 2026-08-20 - Task: Synchronize recent upstream commits

### What was done

- Fetched `upstream/main` through `2bc139ab5` and merged 79 recent commits into `main` as `705813e63831ce8814238435670fdd27c14443bb`; application version advanced to `0.1.179`.
- Restored all local video, audio, media, community, endpoint, and documentation changes after the merge. Resolved 10 overlapping files by combining upstream Composite/capacity/failover/Grok/channel-multiplier changes with local media billing and Firefly/MiniMax routing behavior.
- Kept upstream migrations `226_add_usage_log_effective_model_indexes_notx.sql`, `227_composite_routes_add_cn_providers.sql`, and `228_channel_pricing_multipliers.sql` without adding schema changes. Preserved recovery branch `backup/pre-upstream-sync-20260820-210841` and stash `812da27c31c476a65407d8aa1ff2a3691791fac8`.

### Testing

- `go generate ./cmd/server`: passed.
- `go test ./... -count=1` from `backend/`: passed after a transient single-test 401 was reproduced as a passing test on five reruns.
- Targeted `go test -race` covering Firefly/MiniMax recovery, video billing, channel multipliers, and OpenAI failover: passed.
- `pnpm install --frozen-lockfile`, `pnpm lint:check`, `pnpm typecheck`, `pnpm exec vitest run --reporter=dot`, and `pnpm build` from `frontend/`: passed; 251 files and 1762 tests passed.
- Compose rendering for root and all deploy files passed with temporary required-variable values; `deploy/test-caddyfile-cache.sh` passed. `git diff --check`, `git diff --cached --check`, conflict-marker scan, and upstream ancestor check passed.

### Notes

- `backend/internal/handler/endpoint.go`, `openai_gateway_handler.go`, `openai_live.go`: merged endpoint/audio support, Composite routing, capacity recovery, monitor probes, and existing media behavior.
- `backend/internal/service/channel.go`, `openai_gateway_grok.go`, `openai_gateway_upstream_errors.go`: merged multiplier pricing, Grok 4.6 reasoning support, and upstream retry/error classification.
- `backend/internal/service/openai_videos_test.go`: added reasoning-cache methods required by the upstream `GatewayCache` interface.
- `frontend/src/components/admin/channel/PricingEntryCard.vue`: initialized interval multiplier fields in video tier constructors.
- `docs/UPSTREAM_SYNC.md`: appended this synchronization record.
- Rollback: preserve the backup branch and stash; after saving any newer work, use `git revert -m 1 705813e63831ce8814238435670fdd27c14443bb` to reverse the merge. Do not drop the stash or delete user-owned untracked files.

## 2026-08-20 - Task: Publish synchronized latest code to china-api latest

### What was done

- Built the complete current worktree after the upstream synchronization as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed the multi-architecture image to Docker Hub.
- Published application version `0.1.179`, source baseline commit metadata `705813e63`, and build time `2026-08-20T14:29:46Z`.
- Replaced the previous remote `latest` index `sha256:478e69768b611910445f358c688e27138ec2a2a9d3d42b1d07ecaea9b43e36ec` with `sha256:f11f74b31181e2e4ceb3e1b90ca6e8cb0bfbea43105f8af3b67ada8653cb35ed`.

### Testing

- Before publication, `git diff --check`, `git diff --cached --check`, and `go test ./internal/handler -run '^TestFireflyVideoStatusAndContentDoNotUseInactiveMiniMaxRecoveryRoute$' -count=1` passed; the immediately preceding synchronization task also passed the complete backend and frontend suites.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.26.6-alpine --build-arg VERSION=0.1.179 --build-arg COMMIT=705813e63 --build-arg DATE=2026-08-20T14:29:46Z --tag iotwq/china-api:latest --push .`: passed; the frontend and both backend binaries built successfully.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed remote index `sha256:f11f74b31181e2e4ceb3e1b90ca6e8cb0bfbea43105f8af3b67ada8653cb35ed`, amd64 manifest `sha256:d2e7a7901f8add558c18483ca9f6cdd044393370621e752466d2d597877b8be4`, and arm64 manifest `sha256:c2c2708bdf368f7b0ccf545019c561887ad09d4053c26d3ad66c02969d090896`.
- arm64 was pulled and executed by its immutable Docker Hub digest; it reported `Sub2API 0.1.179 (commit: 705813e63, built: 2026-08-20T14:29:46Z)`.
- Docker Hub's CDN repeatedly reset the amd64 config-blob download with EOF during post-push verification. A cache-only `linux/amd64` load using the identical build arguments exported the exact published amd64 manifest `sha256:d2e7a7901f8add558c18483ca9f6cdd044393370621e752466d2d597877b8be4`, and its runtime reported the same `Sub2API 0.1.179` version metadata.
- Final `git diff --check` and `git diff --cached --check` passed before this record was appended.

### Notes

- `progress.md`: records the multi-architecture publication, immutable digests, runtime verification, transient Docker Hub CDN verification issue, and rollback point; no business source, database, balance, production task, or running service was modified by this publication step.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:478e69768b611910445f358c688e27138ec2a2a9d3d42b1d07ecaea9b43e36ec` to restore the previous multi-architecture image index without changing repository files or production data.

## 2026-08-24 - Task: Synchronize recent upstream commits

### What was done

- Fetched `upstream/main` through `03e8ab413` and merged 176 recent commits into `main` as `edf3986d837e8986f2cf35dfb1c1baa0a63bda82`; application version advanced from `0.1.179` to `0.1.180`.
- Restored the complete local worktree after the merge, including staged customizations, the existing unstaged publication record, and all 77 untracked media/community files. Combined upstream plugin, auto-reset, Fast tier, continuation, retry, pricing, and model-plaza behavior with local video, audio, Nano Banana, billing, monitoring, image-entry, and community features.
- Adapted local media handlers to the new account-object scheduling feedback API, added the account-mapped upstream model as a fallback usage-billing candidate for multi-turn WebSocket aliases, and retained the v0.1.180 SetupToken native Images capability while keeping OAuth on the Responses bridge.
- Preserved recovery branch `backup/pre-upstream-sync-20260824-211411` and stash `33437ce09706bcf69e8f71b44b82c267b60afb4e`; no Git push, image build, database connection, balance change, production task mutation, or running-service restart was performed.

### Testing

- `go generate ./cmd/server`: passed after restoring Wire 0.7.0's locked `github.com/google/subcommands v1.2.0` checksum entries for Go 1.27.
- `go test ./... -count=1` from `backend/`: passed the complete backend suite, including handler, service, repository, routes, migrations, plugin API, and generated server packages.
- Targeted `go test -race` for the merged WebSocket multi-turn usage, SetupToken/native Images, compact normalization, and Grok media geometry/video paths: passed.
- `pnpm install --frozen-lockfile`, `pnpm lint:check`, `pnpm typecheck`, `pnpm exec vitest run --reporter=dot`, and `pnpm build` from `frontend/`: passed; 258 files and 1840 tests passed.
- Root Compose and `deploy/docker-compose.yml`, `deploy/docker-compose.local.yml`, `deploy/docker-compose.dev.yml`, and `deploy/docker-compose.standalone.yml` rendered successfully with temporary required variables; `deploy/test-caddyfile-cache.sh` passed.
- `git diff --check`, `git diff --cached --check`, strict conflict-marker scanning, upstream ancestor verification, and the 77-object untracked-path fingerprint check passed.

### Notes

- Upstream merge inventory: 479 upstream files changed between `2bc139ab5` and `03e8ab413`; exact upstream inventory is available with `git diff --name-status 2bc139ab5..03e8ab413`.
- `backend/cmd/server/main.go`, `wire.go`, `wire_gen.go`: combine plugin lifecycle and quota auto-reset with the local video compensator and local handler dependencies.
- `backend/internal/handler/openai_gateway_handler.go`, `openai_audio.go`, `openai_nano_banana.go`, `openai_videos.go`: merge continuation/probe routing and adapt local media scheduling feedback to the new account argument.
- `backend/internal/service/account.go`, `openai_account_scheduler.go`, `openai_gateway_*.go`: combine Guardian/continuation/Fast/retry behavior with local wrapped-status, monitoring, Grok, media, and WebSocket billing compatibility.
- `backend/go.sum`: restores the Wire subcommands checksum required by the Go 1.27 toolchain without changing the locked dependency version.
- `deploy/docker-compose*.yml`, `frontend/src/views/HomeView.vue`: combine new upstream gateway controls and model-plaza entry with the existing image workspace and 3D home page.
- `docs/UPSTREAM_SYNC.md`, `progress.md`: record the synchronization range, compatibility decisions, verification evidence, file groups, and rollback points.
- Rollback: preserve the backup branch and stash; after saving newer work, use `git revert -m 1 edf3986d837e8986f2cf35dfb1c1baa0a63bda82` to reverse the upstream merge. Do not drop stash `33437ce09706bcf69e8f71b44b82c267b60afb4e` or delete user-owned untracked files.

## 2026-08-24 - Task: Synchronize nine additional upstream commits

### What was done

- Fetched `upstream/main` through `e2d9b823f` and merged 9 additional commits into `main` as `4ba2feebcad4bffada5257befa50c6ff00a3b484`; application version advanced from `0.1.180` to `0.1.181`.
- Integrated the official Grok CLI User-Agent, recursive Gemini tool-schema sanitization, Responses Lite `parallel_tool_calls` preservation, and whole-item rejected-status cleanup. All six overlapping service files merged automatically with the existing local retry, monitoring, media, and billing behavior.
- Restored the exact pre-sync worktree boundary with `stash apply --index`: 283 staged paths, one unstaged `progress.md` path, and 77 untracked media/community objects. Preserved backup branch `backup/pre-upstream-sync-20260824-224151` and stash `bc3d458f2a9f1e69a355fe4102a944dba4568fe1`.
- No Git push, Docker image build, database connection, balance change, production task mutation, or running-service restart was performed.

### Testing

- `go generate ./cmd/server`: passed.
- `go test ./... -count=1` from `backend/`: passed the complete backend suite.
- Targeted `go test -race` covering Grok CLI identity, Gemini tool-schema cleanup, Responses Lite parallel tools, and rejected-field retry normalization: passed.
- `pnpm lint:check`, `pnpm typecheck`, `pnpm exec vitest run --reporter=dot`, and `pnpm build` from `frontend/`: passed; 258 files and 1840 tests passed.
- Root Compose and all four deploy Compose files rendered with temporary required variables; `deploy/test-caddyfile-cache.sh` passed.
- `git diff --check`, `git diff --cached --check`, strict conflict-marker scanning, upstream ancestor verification, and the 77-object untracked-path fingerprint check passed.

### Notes

- `backend/internal/pkg/xai/billing.go`, `cli_identity_test.go`, `backend/internal/repository/http_upstream_test.go`: align Grok CLI traffic with the official CLI User-Agent while preserving direct API-host identity.
- `backend/internal/service/gemini_messages_compat_service*.go`: recursively sanitizes unsupported nested schema fields and invalid enum values.
- `backend/internal/service/openai_gateway_request_body*.go`, `openai_responses_rejected_field_retry*.go`: retains valid Responses Lite parallel-tool intent and clears rejected status fields consistently.
- `backend/internal/service/openai_gateway_grok*.go`, `grok_observed_models.go`, `grok_upstream_headers.go`: merges the official identity change with existing Grok request, quota, retry, and usage behavior.
- `docs/UPSTREAM_SYNC.md`, `progress.md`: record this additional synchronization, validation evidence, file groups, and rollback points.
- Rollback: preserve the backup branch and stash; after saving newer work, use `git revert -m 1 4ba2feebcad4bffada5257befa50c6ff00a3b484` to reverse this additional upstream merge. Do not drop stash `bc3d458f2a9f1e69a355fe4102a944dba4568fe1` or delete user-owned untracked files.

## 2026-08-24 - Task: Publish china-api 0.1.181 latest

### What was done

- Built the complete current worktree as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed the multi-architecture image to Docker Hub.
- Published application version `0.1.181`, source baseline commit metadata `4ba2feebc`, and build time `2026-08-24T15:13:54Z` using `golang:1.27.0-alpine`.
- Replaced the previous remote `latest` index `sha256:f11f74b31181e2e4ceb3e1b90ca6e8cb0bfbea43105f8af3b67ada8653cb35ed` with `sha256:4cb92a9d24bc23d206511f812630039f9dc14d878eaef3d3e8b06635776d1c9f`.

### Testing

- The immediately preceding v0.1.181 synchronization passed `go test ./... -count=1`, targeted race tests, frontend lint/typecheck, all 1840 Vitest tests, production build, Compose rendering, and Caddy verification.
- `git diff --check` and `git diff --cached --check` passed before the image build.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.1.181 --build-arg COMMIT=4ba2feebc --build-arg DATE=2026-08-24T15:13:54Z --tag iotwq/china-api:latest --push .`: passed; the frontend and both Go binaries built successfully.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed remote index `sha256:4cb92a9d24bc23d206511f812630039f9dc14d878eaef3d3e8b06635776d1c9f`, amd64 manifest `sha256:9880b5b458208e26a6652cd1202374cbb57dee2a05121ebd2182da86c8732150`, and arm64 manifest `sha256:a8dafa5f1b6857669b905c7959a014524ac9560b00d6f22710c0f025a1d04342`.
- Fresh Docker Hub pulls and immutable-digest execution passed for both architectures; each reported `Sub2API 0.1.181 (commit: 4ba2feebc, built: 2026-08-24T15:13:54Z)`.

### Notes

- `progress.md`: records the multi-architecture publication, immutable digests, runtime verification, and rollback point. No business source, database, balance, production task, or running service was modified by this publication step.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:f11f74b31181e2e4ceb3e1b90ca6e8cb0bfbea43105f8af3b67ada8653cb35ed` to restore the previous multi-architecture image index without changing repository files or production data.

## 2026-08-25 - Task: Synchronize recent upstream commits to v0.1.182

### What was done

- Fetched `upstream/main` through `4ff136cfd` and merged 26 recent commits into `main` as `6911a9784e692abc38649c70a3f7d06c2c743f59`; application version advanced from `0.1.181` to `0.1.182`.
- Integrated OAuth 429 quota-aware scheduling, Responses Lite normalization, Composite/Kimi routing, channel-monitor aggregation, Anthropic cache billing safety, Antigravity model routing, OAuth image prompt preservation, and payment-result balance refresh.
- Restored the exact pre-sync worktree boundary with `stash apply --index`: 283 staged paths, one unstaged `progress.md` path, and 77 untracked media/community objects. All overlapping files merged automatically while retaining local media, billing, retry, monitoring, WebSocket, community, and UI behavior.
- Preserved backup branch `backup/pre-upstream-sync-20260825-212816` and stash `0d48d7fed3f8af6ab438f3326656396e0a67322d`; no Git push, image build, database connection, balance change, production task mutation, or running-service restart was performed.

### Testing

- `go generate ./cmd/server`: passed.
- `go test ./... -count=1` from `backend/`: passed the complete backend suite.
- Targeted `go test -race` covering OAuth 429 reset handling, cache-creation billing, Responses Lite, Composite/Kimi routing, Antigravity model mapping, channel-monitor aggregation, and OAuth image prompt preservation: passed.
- `pnpm install --frozen-lockfile`, `pnpm lint:check`, `pnpm typecheck`, `pnpm exec vitest run --reporter=dot`, and `pnpm build` from `frontend/`: passed; 258 files and 1841 tests passed.
- Root Compose and all four deploy Compose files rendered with temporary required variables; `deploy/test-caddyfile-cache.sh` passed.
- `git diff --check`, `git diff --cached --check`, strict conflict-marker scanning, upstream ancestor verification, and the 77-object untracked-path fingerprint check passed.

### Notes

- `backend/internal/service/openai_account_runtime_block_fastpath*.go`, `ratelimit_service*.go`, `openai_gateway_upstream_errors.go`: add quota-aware OAuth 429 scheduling and OpenCode Go reset-duration handling.
- `backend/internal/service/openai_responses_lite_tools*.go`, `openai_gateway_request_body*.go`, `openai_ws_*.go`: pin valid Responses Lite tool behavior across HTTP and WebSocket without losing numeric precision.
- `backend/internal/service/billing_service*.go`, `gateway_anthropic_*.go`: prevent contradictory Anthropic cache breakdowns from being billed twice.
- `backend/internal/service/composite_*.go`, `backend/internal/repository/channel_monitor_v2_*.go`: resolve Kimi Code K3 and Composite monitoring to concrete platforms.
- `backend/internal/service/openai_images*.go`, `account_test_service*.go`, `antigravity_*.go`, `frontend/src/views/user/PaymentResultView.vue`: preserve OAuth image prompts, correct Antigravity model routing, and refresh balances after payment completion.
- `docs/UPSTREAM_SYNC.md`, `progress.md`: record the synchronization, validation evidence, file groups, and rollback points.
- Rollback: preserve the backup branch and stash; after saving newer work, use `git revert -m 1 6911a9784e692abc38649c70a3f7d06c2c743f59` to reverse this upstream merge. Do not drop stash `0d48d7fed3f8af6ab438f3326656396e0a67322d` or delete user-owned untracked files.

## 2026-08-25 - Task: Synchronize additional upstream commits to v0.1.183

### What was done

- Fetched `upstream/main` through `7634e3c23` and merged 13 additional commits into `main` as `978d9e24c7fdda3669f445cb7358fe8c6b3b25d7`; application version advanced from `0.1.182` to `0.1.183`.
- Integrated Codex session-id affinity, sticky capacity spillover preservation, typed Responses client-tool item IDs, Kimi concurrency-403 recovery, Antigravity token clamping, and email-binding alias/concurrency guards.
- Restored the exact pre-sync worktree boundary with `stash apply --index`: 283 staged paths, one unstaged `progress.md` path, and 77 untracked media/community objects. The overlapping scheduling and gateway tests merged automatically while retaining local retry, monitoring, billing, media, WebSocket, community, and UI behavior.
- Preserved backup branch `backup/pre-upstream-sync-20260825-215813` and stash `7f0236e36ae8c05874212f405644198b1efb71ad`; no Git push, image build, database connection, balance change, production task mutation, or running-service restart was performed.

### Testing

- `go generate ./cmd/server`: passed.
- `go test ./... -count=1` from `backend/`: passed the complete backend suite.
- Targeted `go test -race` covering Responses tool-call IDs, email-binding aliases, Kimi concurrency 403, Antigravity token limits, Codex session-id affinity, and sticky capacity spillover: passed.
- `pnpm install --frozen-lockfile`, `pnpm lint:check`, `pnpm typecheck`, `pnpm exec vitest run --reporter=dot`, and `pnpm build` from `frontend/`: passed; 258 files and 1841 tests passed.
- Root Compose and all four deploy Compose files rendered with temporary required variables; `deploy/test-caddyfile-cache.sh` passed.
- `git diff --check`, `git diff --cached --check`, strict conflict-marker scanning, upstream ancestor verification, and the 77-object untracked-path fingerprint check passed.

### Notes

- `backend/internal/pkg/apicompat/responses_client_tools*.go`: preserves and restores typed custom/tool-search call item IDs.
- `backend/internal/repository/user_repo.go`, `backend/internal/service/auth_email_binding.go`, `auth_service_email_bind_test.go`: prevent email alias duplication and concurrent inbox binding races.
- `backend/internal/service/openai_gateway_scheduling.go`, `openai_gateway_service_test.go`, `openai_ws_forwarder_logutil*.go`: retain sticky bindings through capacity spillover and honor Codex session-id headers.
- `backend/internal/service/ratelimit_cn_providers.go`, `ratelimit_service.go`, `openai_gateway_cn_fixes_test.go`: treat exact Kimi concurrency 403 responses as temporary capacity signals without weakening unrelated 403 policy.
- `backend/internal/service/antigravity_gateway_compat*.go`: clamps compatible token limits while preserving existing model mapping behavior.
- `docs/UPSTREAM_SYNC.md`, `progress.md`: record the synchronization, validation evidence, file groups, and rollback points.
- Rollback: preserve the backup branch and stash; after saving newer work, use `git revert -m 1 978d9e24c7fdda3669f445cb7358fe8c6b3b25d7` to reverse this upstream merge. Do not drop stash `7f0236e36ae8c05874212f405644198b1efb71ad` or delete user-owned untracked files.

## 2026-08-25 - Task: Publish china-api 0.1.183 latest

### What was done

- Built the complete current worktree as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed the multi-architecture image to Docker Hub.
- Published application version `0.1.183`, source baseline commit metadata `978d9e24c`, and build time `2026-08-25T14:14:21Z` using `golang:1.27.0-alpine`.
- Replaced the previous remote `latest` index `sha256:4cb92a9d24bc23d206511f812630039f9dc14d878eaef3d3e8b06635776d1c9f` with `sha256:2c9341a79e773938fa2c8b072f0022a4b7eec6f8b48a449399f32b267cbe6ba3`.

### Testing

- The immediately preceding v0.1.183 synchronization passed the complete backend suite, targeted race tests, frontend lint/typecheck, all 1841 Vitest tests, production build, Compose rendering, and Caddy verification.
- `git diff --check` and `git diff --cached --check` passed before the image build.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.1.183 --build-arg COMMIT=978d9e24c --build-arg DATE=2026-08-25T14:14:21Z --tag iotwq/china-api:latest --push .`: passed; the frontend and both Go binaries built successfully.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed remote index `sha256:2c9341a79e773938fa2c8b072f0022a4b7eec6f8b48a449399f32b267cbe6ba3`, amd64 manifest `sha256:46d22879d9749b44a86aaad299a25acbf91141946baee96c5cf1854ce7e6c42b`, and arm64 manifest `sha256:887a5b1ba03322278db4a644ab81bf32936fb8cc2494d59ae492d906b795eca0`.
- Fresh Docker Hub pulls and immutable-digest execution passed for both architectures; each reported `Sub2API 0.1.183 (commit: 978d9e24c, built: 2026-08-25T14:14:21Z)`. The first amd64 authorization attempt hit a transient Docker Hub token EOF; the retry downloaded and executed the image successfully.

### Notes

- `progress.md`: records the multi-architecture publication, immutable digests, runtime verification, transient registry retry, and rollback point. No business source, database, balance, production task, or running service was modified by this publication step.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:4cb92a9d24bc23d206511f812630039f9dc14d878eaef3d3e8b06635776d1c9f` to restore the previous multi-architecture image index without changing repository files or production data.

## 2026-08-28 - Task: Add cc-switch NewAPI balance endpoint

### What was done

- Added the read-only `GET /api/user/self` compatibility endpoint expected by cc-switch's NewAPI usage template.
- Reused the existing dashboard JWT authentication, backend-mode guard, session binding checks, and panel rate limiting; validated the optional `New-Api-User` header against the authenticated JWT user ID.
- Returned the NewAPI `success/data` envelope with `quota` and `used_quota` in the required 500,000-units-per-USD format. The endpoint does not accept API keys or issue JWTs.
- Documented the endpoint, headers, unit conversion, and access-token requirement in `docs/API_ENDPOINT_ADDRESSES.md`.

### Testing

- `go test ./... -count=1` from `backend/`: passed the complete backend suite.
- `go test -tags=unit ./internal/handler -run 'TestAuthHandlerGetNewAPISelf|TestAuthHandlerGetCurrentUser' -count=1`: passed the NewAPI success and user-hint mismatch cases alongside the existing current-user compatibility test.
- `go test ./internal/handler ./internal/server ./internal/server/routes ./internal/server/middleware -count=1`: passed handler, route, and authentication middleware regression coverage.
- `git diff --check` and `git diff --cached --check`: passed; strict conflict-marker and `gofmt -d` checks produced no diagnostics.

### Notes

- `backend/internal/handler/auth_handler.go`: adds NewAPI-compatible self response and safe USD-to-NewAPI quota conversion.
- `backend/internal/handler/auth_current_user_test.go`: adds success and mismatched `New-Api-User` coverage while preserving the existing test's user-owned formatting.
- `backend/internal/server/router.go`: registers `/api/user/self` with the existing JWT, backend-mode, and panel-rate-limit middleware chain.
- `docs/API_ENDPOINT_ADDRESSES.md`: documents cc-switch NewAPI balance integration and token/header requirements.
- `progress.md`: records this implementation, validation evidence, and rollback instructions.
- Rollback: reverse only this task's changes in the four files above (remove `GetNewAPISelf`, its helpers/tests, the `/api/user/self` route, and the NewAPI documentation section), then delete only this final `progress.md` section. No database or production-data rollback is required.

## 2026-08-29 - Task: Publish current china-api latest image

### What was done

- Built the current worktree as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, including the cc-switch NewAPI balance endpoint changes, and pushed the multi-architecture image to Docker Hub.
- Published application version `0.1.183`, source baseline metadata `978d9e24c`, and build time `2026-08-28T17:11:25Z` using `golang:1.27.0-alpine`.
- Replaced the previous `latest` index with remote manifest list `sha256:5e9509256c4021fffc561da2900f0bd4b5a3360a2bfc78a117fcd4bc0f353d48`.

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.1.183 --build-arg COMMIT=978d9e24c --build-arg DATE=2026-08-28T17:11:25Z --tag iotwq/china-api:latest --push .`: passed after one transient Docker Hub authorization EOF retry; both frontend and Go binaries built successfully.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed remote index `sha256:5e9509256c4021fffc561da2900f0bd4b5a3360a2bfc78a117fcd4bc0f353d48`, amd64 manifest `sha256:cc7175bbce471f96b2bd1b514183c028a847b8ec859e3897d23ea1a75f3a787a`, and arm64 manifest `sha256:8b29e09de9b784ab09238d1dcf2deeb99d942b3332cd7096a807f2178a170923`.
- Immutable amd64 pull and `docker run ... /app/sub2api --version` passed, reporting `Sub2API 0.1.183 (commit: 978d9e24c, built: 2026-08-28T17:11:25Z)`.
- Immutable arm64 pull/runtime verification was attempted twice but was blocked by transient Docker Hub authorization/CDN EOF responses; a local arm64 `--load` retry was also blocked while resolving the Dockerfile syntax image metadata. Arm64 remote manifest presence was confirmed, but arm64 runtime execution is not claimed as verified.
- `git diff --check` and `git diff --cached --check` passed before this record was appended.

### Notes

- `progress.md`: records the multi-architecture publication, immutable manifests, amd64 runtime verification, Docker Hub transient failures, and rollback point. No database, balance, production task, or running service was modified by this publication step.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:2c9341a79e773938fa2c8b072f0022a4b7eec6f8b48a449399f32b267cbe6ba3` to restore the preceding `latest` index without changing repository files or production data.

## 2026-08-29 - Task: Hide gpt-image-2 upstream image URLs

### What was done

- Changed synchronous `gpt-image-2` responses routed through OpenAI API URL/API Key accounts so upstream `data[].url` images are downloaded by Sub2API and returned only as `data[].b64_json`; existing Base64 is preserved and any accompanying URL is removed.
- Kept URL fetching on the selected account's proxy path without forwarding the upstream API key, and bounded it to HTTP(S), 60 seconds, 20 MiB per downloaded image, and five validated redirects. Download or validation failures return a generic upstream failure instead of exposing or falling back to the original URL.
- Kept other image models unchanged and updated the user API documentation and media compatibility documentation to describe the Base64-only client contract and URL allowlist requirements.

### Testing

- `go test ./... -count=1` from `backend/`: passed the complete backend suite.
- `go test -race ./internal/service -run 'TestOpenAIImagesAPIKeyGPTImage2|TestGPTImage2ResultDownload|TestOpenAIImagesAPIKeyOtherModels' -count=1`: passed URL conversion, data URL conversion, existing Base64 preservation, URL removal, failure privacy, redirect validation, and non-target model coverage.
- `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts --reporter=dot`: passed the updated public image2 documentation contract test.
- `pnpm typecheck`, `pnpm lint:check`, and `pnpm build` from `frontend/`: passed; the production build completed with only the existing Browserslist age and large-chunk warnings.
- `git diff --check`, `git diff --cached --check`, and `gofmt -d` for the changed Go files: passed with no diagnostics.

### Notes

- `backend/internal/service/openai_images.go`: invokes response privacy normalization before writing API Key image JSON to the client.
- `backend/internal/service/openai_images_response_privacy.go`: validates, downloads, bounds, and converts gpt-image-2 URL results without exposing upstream locations.
- `backend/internal/service/openai_images_response_privacy_test.go`: covers Base64 and URL response variants, failure privacy, redirect policy, and model scoping.
- `frontend/src/views/user/ApiDocsView.vue`: changes image2 examples and response handling to the Base64-only client contract.
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: verifies the updated image2 documentation and removes the old client-side URL-fetch expectation.
- `docs/API_DOCS.md`: records the public `data[].b64_json` response contract.
- `docs/OPENAI_MEDIA_COMPAT.md`: records server-side URL conversion limits, proxy behavior, and allowlist requirements.
- `progress.md`: records this implementation, validation evidence, changed files, and rollback instructions.
- Rollback: remove `openai_images_response_privacy.go` and its test, restore the original non-streaming handler call/signature in `openai_images.go`, revert the image2 documentation/test wording, and delete only this final `progress.md` section. No database or production-data rollback is required.

## 2026-08-29 - Task: Make OpenAI image URL conversion account-configurable

### What was done

- Replaced automatic gpt-image-2 URL conversion with an OpenAI API Key account switch named “Convert image URLs to Base64”, stored as `accounts.extra.openai_image_url_to_b64_json`; the switch is available on both account creation and editing and defaults to off without a database migration.
- When the switch is off, synchronous image JSON is preserved exactly, including responses that contain both `data[].url` and `data[].b64_json`. When enabled, every image item URL for that account is removed: existing Base64 is retained, while URL-only items are downloaded and converted to `b64_json`. The rule applies to all image models routed through that account.
- Restored the public image2 documentation and JavaScript example so clients detect and parse Base64, data URLs, and HTTP(S) URLs, including responses that contain both fields.

### Testing

- `go test ./... -count=1` from `backend/`: passed the complete backend suite.
- `go test -race ./internal/service -run 'TestOpenAIImagesAPIKeyConversion|TestOpenAIImageResultDownload|TestAccountShouldConvert' -count=1`: passed switch-default, enabled conversion, mixed URL/Base64, other-model, failure privacy, and redirect validation coverage.
- `pnpm exec vitest run src/components/account/__tests__/CreateAccountModal.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts src/views/user/__tests__/ApiDocsView.spec.ts --reporter=dot`: passed all 79 account-form and public-documentation tests.
- `pnpm typecheck`, `pnpm lint:check`, and `pnpm build` from `frontend/`: passed; the production build completed with only the existing Browserslist age and large-chunk warnings.
- `git diff --check`, `git diff --cached --check`, and `gofmt -d` for all changed Go files: passed with no diagnostics.

### Notes

- `backend/internal/service/account.go`: defines the account extra key and strict OpenAI API Key switch reader.
- `backend/internal/service/openai_images.go`: gates synchronous response conversion on the selected account switch instead of the model name.
- `backend/internal/service/openai_images_response_conversion.go`: implements bounded URL/data-URL downloading, validation, and Base64 normalization when enabled.
- `backend/internal/service/openai_images_response_conversion_test.go`: verifies enabled and disabled behavior, mixed fields, model scope, failure privacy, and redirects.
- `frontend/src/components/account/CreateAccountModal.vue`, `EditAccountModal.vue`: add, initialize, load, save, and clear the account switch.
- `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts`, `EditAccountModal.spec.ts`: verify visibility, default-off creation, persistence, and clearing.
- `frontend/src/i18n/locales/zh/admin/accounts.ts`, `frontend/src/i18n/locales/en/admin/accounts.ts`: add bilingual switch labels and behavior descriptions.
- `frontend/src/views/user/ApiDocsView.vue`, `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`, `docs/API_DOCS.md`: restore client-side Base64 and URL compatibility guidance and tests.
- `docs/OPENAI_MEDIA_COMPAT.md`: documents the account switch, storage key, default behavior, conversion limits, proxy use, and allowlist policy.
- `progress.md`: records this correction, final behavior, validation evidence, changed files, and rollback instructions.
- Rollback: remove the account switch UI/i18n/tests and `OpenAIImageURLToB64JSONExtraKey` reader, remove `openai_images_response_conversion.go` and its tests, restore the original non-streaming response handler signature/call, restore the prior documentation wording as required, and delete only this final `progress.md` section. No schema or production-data rollback is required; existing extra values become inert if the reader is removed.

## 2026-08-29 - Task: Remove the Nano Banana Pro simplified API from user documentation

### What was done

- Removed the `/v1/api/nano-banana` simplified endpoint, `nano-banana-pro` simplified model name, simplified-only parameters, two simplified request examples, and related notes from the user API documentation.
- Kept Nano Banana Pro as the Gemini native `gemini-3-pro-image-preview` model alongside Nano Banana 2, with only the two native `generateContent` endpoints, native parameters, and three native examples displayed.
- Left the backend simplified endpoint implementation unchanged; this task changes only what is presented in user-facing API documentation.

### Testing

- `pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts --reporter=dot`: passed and verifies the Gemini section contains only three native examples and none of the simplified endpoint/model wording.
- `pnpm typecheck`: passed.
- `pnpm lint:check`: passed when rerun independently; the first parallel run raced with Vitest deleting its temporary config module and failed with a transient `ENOENT`, not a source diagnostic.
- `pnpm build`: passed; the production build completed with only the existing Browserslist age and large-chunk warnings.
- `git diff --check` and `git diff --cached --check`: passed.

### Notes

- `frontend/src/views/user/ApiDocsView.vue`: removes simplified Nano Banana constants, endpoint, model, parameters, examples, and notes while retaining native Gemini documentation.
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`: updates the Gemini documentation contract to require only native content.
- `docs/API_DOCS.md`: removes the simplified Nano Banana entry from the formal user documentation record.
- `progress.md`: records this documentation-only change, validation evidence, changed files, and rollback instructions.
- Rollback: restore the two simplified request constants and their model/endpoint/parameter/example/note entries in `ApiDocsView.vue`, restore the corresponding test expectations and `docs/API_DOCS.md` bullets, then delete only this final `progress.md` section. No backend, database, or production-data rollback is required.

## 2026-08-29 - Task: Publish the latest china-api image

### What was done

- Built the complete current worktree as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, including the account-level OpenAI image URL conversion switch and the latest user API documentation changes, then pushed the multi-architecture image to Docker Hub.
- Published application version `0.1.183`, source baseline metadata `978d9e24c`, and build time `2026-08-29T14:10:27Z` using `golang:1.27.0-alpine`.
- Replaced the preceding `latest` index `sha256:5e9509256c4021fffc561da2900f0bd4b5a3360a2bfc78a117fcd4bc0f353d48` with `sha256:8c87bc073ee97e20b2d28c7ee5592532785fe453ee41ef1e7d1c19c9c0c49dcc`.

### Testing

- The immediately preceding source work passed the complete backend suite, targeted race tests, account and API-documentation tests, frontend typecheck/lint, and frontend production build.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.1.183 --build-arg COMMIT=978d9e24c --build-arg DATE=2026-08-29T14:10:27Z --tag iotwq/china-api:latest --push .`: passed; the frontend and both Go binaries built successfully and the multi-architecture index was pushed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed remote index `sha256:8c87bc073ee97e20b2d28c7ee5592532785fe453ee41ef1e7d1c19c9c0c49dcc`, amd64 manifest `sha256:d6e0647bdadaa487f136cb2b313f45dbd2223a5910c2927cbc83647044d896b5`, and arm64 manifest `sha256:9819453173afb90b4e66826eb6f39fbc18ae260dcbe958db06880c86145f095c`.
- Immutable-digest pulls and runtime version checks passed for both architectures; each reported `Sub2API 0.1.183 (commit: 978d9e24c, built: 2026-08-29T14:10:27Z)`. The first amd64 CDN blob download hit a transient EOF; the retry completed successfully.
- `git diff --check` and `git diff --cached --check`: passed before this record was appended.

### Notes

- `progress.md`: records the multi-architecture publication, immutable digests, both architecture runtime checks, transient Docker Hub retry, and rollback point. No business source, database, balance, production task, or running service was modified by this publication step.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:5e9509256c4021fffc561da2900f0bd4b5a3360a2bfc78a117fcd4bc0f353d48` to restore the preceding multi-architecture image index without changing repository files or production data.

## 2026-08-30 - Task: Add long-lived cc-switch NewAPI balance access

### What was done

- Added NewAPI-compatible `GET /api/user/self` balance querying for cc-switch. The endpoint accepts the existing dashboard JWT or a dedicated long-lived `sub_bal_...` credential and returns the current balance in NewAPI quota units.
- Added authenticated `GET /api/user/token` generation/reset. The plaintext credential is shown once, only its SHA-256 hash is stored, and rotating it immediately invalidates the prior credential.
- Restricted the dedicated credential to the balance endpoint. It cannot call models, rotate itself, or authenticate other dashboard APIs; an optional `New-Api-User` header must match the authenticated user.
- Added the profile-page User ID and token controls, user guidance, migration, dependency injection, focused backend/frontend coverage, and formal cc-switch configuration documentation.
- Verified the behavior against official NewAPI source commit `918427d8ab41f6adaa4113d0496f1f8621855b70`: current NewAPI exposes User ID, token rotation, and `/api/user/self` balance querying; this implementation deliberately narrows the generated token to balance lookup only.

### Testing

- `cd backend && go generate ./cmd/server`: passed and regenerated dependency injection with the new repository and service.
- `cd backend && go test ./... -count=1`: passed for the complete backend, including handlers, repositories, middleware, services, server routes, and migrations.
- `cd backend && go test -race ./internal/service ./internal/repository ./internal/server/middleware -run 'TestNewAPI' -count=1`: passed with the race detector.
- `cd frontend && pnpm typecheck`: passed.
- `cd frontend && pnpm lint:check`: passed.
- `cd frontend && pnpm exec vitest run --reporter=dot`: passed, 260 test files and 1847 tests.
- `cd frontend && pnpm build`: passed; only the existing stale Browserslist data and large-chunk warnings were reported.
- `gofmt -d` on all changed Go files, `git diff --check`, and `git diff --cached --check`: passed before this record was appended.

### Notes

- `.gitignore`: allows the dedicated NewAPI balance access guide to be delivered from `docs/`.
- `backend/migrations/231_newapi_balance_access_token.sql`: adds hash-only token storage and the active-user uniqueness index.
- `backend/migrations/newapi_balance_access_token_migration_test.go`: verifies the migration stores only a fixed-length hash and creates the partial unique index.
- `backend/internal/service/newapi_access_token.go`: generates, hashes, rotates, and authenticates dedicated balance tokens.
- `backend/internal/service/newapi_access_token_test.go`: covers generation, hashing, authentication, and invalid-token rejection.
- `backend/internal/repository/newapi_access_token_repo.go`: persists rotated hashes and resolves active user subjects by hash.
- `backend/internal/repository/newapi_access_token_repo_test.go`: covers token rotation and active-user lookup SQL.
- `backend/internal/server/middleware/newapi_balance_auth.go`: dispatches dedicated balance tokens while preserving dashboard JWT fallback.
- `backend/internal/server/middleware/newapi_balance_auth_test.go`: verifies dedicated-token authentication and JWT fallback.
- `backend/internal/handler/auth_handler.go`: adds token rotation and NewAPI-compatible self/balance responses.
- `backend/internal/handler/auth_current_user_test.go`: verifies quota conversion, User ID matching, and one-time token responses.
- `backend/internal/server/router.go`: registers the root `/api/user/token` and `/api/user/self` compatibility routes with their separate authentication boundaries.
- `backend/internal/server/api_contract_test.go`: updates handler construction for the new dependency.
- `backend/internal/repository/wire.go`: registers the balance-token repository provider.
- `backend/internal/service/wire.go`: registers the balance-token service provider.
- `backend/cmd/server/wire_gen.go`: wires the new repository and service into the production auth handler.
- `frontend/src/api/user.ts`: calls the root token-generation endpoint without appending `/api/v1`.
- `frontend/src/api/__tests__/userNewAPIBalance.spec.ts`: verifies the root endpoint URL and response handling.
- `frontend/src/components/user/profile/ProfileNewAPIAccessCard.vue`: displays User ID and one-time token generation/reset and copy controls.
- `frontend/src/components/user/profile/__tests__/ProfileNewAPIAccessCard.spec.ts`: verifies User ID display and one-time token presentation.
- `frontend/src/views/user/ProfileView.vue`: adds the cc-switch/NewAPI balance access card to the profile page.
- `frontend/src/views/user/__tests__/ProfileView.spec.ts`: verifies the profile page includes the new card.
- `frontend/src/i18n/locales/zh/dashboard.ts`: adds Chinese labels, confirmations, and token lifecycle guidance.
- `frontend/src/i18n/locales/en/dashboard.ts`: adds matching English labels, confirmations, and token lifecycle guidance.
- `docs/API_ENDPOINT_ADDRESSES.md`: documents the NewAPI-compatible endpoints, units, and cc-switch fields alongside endpoint configuration.
- `docs/NEWAPI_BALANCE_ACCESS.md`: provides the dedicated user workflow and security boundary reference.
- `progress.md`: records this implementation, verification evidence, changed files, and rollback instructions.
- Rollback: revert the eventual commit containing this task with `git revert <commit-containing-this-task>`. If migration 231 has already run, first revoke use of all `sub_bal_...` tokens, then remove storage with `psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c 'BEGIN; DROP INDEX IF EXISTS idx_users_newapi_access_token_hash; ALTER TABLE users DROP COLUMN IF EXISTS newapi_access_token_hash; COMMIT;'`. No user balance or usage data is changed by this feature.

## 2026-08-31 - Task: Publish the latest china-api image with NewAPI balance access

### What was done

- Built the complete current worktree as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, including the cc-switch/NewAPI balance access implementation, then pushed the multi-architecture image to Docker Hub.
- Published application version `0.1.183`, source baseline metadata `978d9e24c`, and build time `2026-08-30T16:04:14Z` using `golang:1.27.0-alpine`.
- Replaced the preceding `latest` index `sha256:8c87bc073ee97e20b2d28c7ee5592532785fe453ee41ef1e7d1c19c9c0c49dcc` with `sha256:d60bb469121a9c212e2cbdce93794509d69634538901933f9ca422a93c699e2e`.

### Testing

- The immediately preceding cc-switch/NewAPI balance source task passed the complete backend suite, targeted race detector checks, frontend typecheck/lint, 1847 frontend tests, and the frontend production build.
- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.1.183 --build-arg COMMIT=978d9e24c --build-arg DATE=2026-08-30T16:04:14Z --tag iotwq/china-api:latest --push .`: passed; the frontend and both Go binaries built successfully and the multi-architecture index was pushed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed remote index `sha256:d60bb469121a9c212e2cbdce93794509d69634538901933f9ca422a93c699e2e`, amd64 manifest `sha256:69134450c920d5e4d3cad3b5c559148845971a0136bbf000eb0c5b2b9a6030b4`, and arm64 manifest `sha256:54bbc1b7c8788ae969f92e8764b3344692c80815fcbabdd4bb932efb5ed2c066`.
- Immutable-index pulls and runtime version checks passed for both architectures; each reported `Sub2API 0.1.183 (commit: 978d9e24c, built: 2026-08-30T16:04:14Z)`.
- `git diff --check` and `git diff --cached --check`: passed before this record was appended.

### Notes

- `progress.md`: records the multi-architecture publication, immutable digests, both architecture runtime checks, and rollback point. No business source, database, balance, production task, or running service was modified by this publication step.
- Rollback: run `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:8c87bc073ee97e20b2d28c7ee5592532785fe453ee41ef1e7d1c19c9c0c49dcc` to restore the preceding multi-architecture image index without changing repository files or production data.

## 2026-08-31 - Task: Sync upstream/main through v0.1.184

### What was done

- Fetched both remotes and merged `upstream/main` from `7634e3c23` through the final remote tip `200602b41` into local `main`; merge commits are `58b0d7449b31a5c5d0ae34adb7ad13b7552b7fbf` and `f0bcb5948`.
- Preserved the complete pre-sync worktree in stash `7c3bdfa62f53d01d9896932ee78515dc50fd43ae`, restored it after the merge, and resolved all 10 content conflicts while retaining both upstream features and local media/community customizations.
- Reconciled the merged usage-log schema to 64 ordered insert/scan fields, regenerated Wire, and fixed upstream signature changes exposed by compilation without changing production data or a running service.
- Re-fetched before completion and included the final 2 commits that preserve known image-input capabilities for Codex custom providers; the second stash `b7401f38fa399ce2396f865f5fcd371024a6ff31` protects the fully validated intermediate worktree.

### Testing

- `cd backend && go generate ./cmd/server`: passed.
- `cd backend && go test ./... -count=1`: passed for the complete backend.
- `cd backend && go test -tags=unit ./internal/repository -run 'TestPrepareUsageLogInsert_SessionIDArgWiring|TestPrepareUsageLogInsert_SessionIDNullWhenAbsent' -count=1`: passed and verifies the merged 64-field usage-log argument ordering.
- `cd frontend && pnpm typecheck`: passed.
- `cd frontend && pnpm lint:check`: passed.
- `cd frontend && pnpm exec vitest run --reporter=dot`: passed, 265 test files and 1906 tests.
- `cd frontend && pnpm build`: passed; only the existing Browserslist age and large-chunk warnings were reported.
- Final upstream increment: targeted `internal/service` Codex image-capability tests and `internal/handler` Codex model tests passed after merging `200602b41`.
- `git diff --check`, `git diff --cached --check`, strict conflict-marker scanning, unresolved-index checks, and the `upstream/main` ancestor check passed.

### Notes

- Merge commits `58b0d7449b31a5c5d0ae34adb7ad13b7552b7fbf` and `f0bcb5948`: contain the authoritative 343-file upstream change list (`git diff --name-only 7634e3c23..200602b41`).
- `backend/cmd/server/wire_gen.go`: retains upstream channel-monitor API Key injection and local community-chat dependency injection after regeneration.
- `backend/internal/repository/usage_log_repo_insert.go`, `backend/internal/repository/usage_log_repo_query.go`, `backend/internal/repository/usage_log_session_id_unit_test.go`: merge upstream reasoning/compaction columns with local video accounting columns and pin their 64-field order.
- `backend/internal/service/openai_gateway_grok.go`, `backend/internal/service/openai_gateway_grok_test.go`: use the exported Grok 4.6 reasoning capability check and retain both cache-identity and `xhigh` coverage.
- `backend/internal/service/openai_gateway_usage.go`, `backend/internal/service/openai_gateway_service_test.go`: retain credential-account billing, native compaction, media balance settlement, and the updated stream failure contract.
- `backend/internal/service/upstream_models.go`: adapts the local MiniMax model-list shortcut to the upstream three-value return contract.
- `backend/internal/handler/openai_codex_models_handler_test.go`, `backend/internal/handler/openai_ws_v2_passthrough_cyber_test.go`: supply the new video task-binding repository constructor argument.
- `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts`, `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`: retain upstream metadata/reasoning audit coverage and local image-conversion/video-billing coverage.
- `docs/UPSTREAM_SYNC.md`: records the v0.1.184 scope, migrations, compatibility decisions, verification, and rollback point.
- `progress.md`: records this sync task and its validation evidence. Rollback by first saving the current worktree, then running `git revert -m 1 f0bcb5948` followed by `git revert -m 1 58b0d7449b31a5c5d0ae34adb7ad13b7552b7fbf`; the pre-sync base is `978d9e24c7fdda3669f445cb7358fe8c6b3b25d7`, the original recoverable worktree remains in stash `7c3bdfa62f53d01d9896932ee78515dc50fd43ae`, and the validated intermediate snapshot remains in `b7401f38fa399ce2396f865f5fcd371024a6ff31`.

## 2026-09-01 - Task: Sync latest upstream/main through v0.1.185

### What was done

- Fetched `upstream/main` and merged the latest remote tip `0d27f45ea` into local `main` with merge commit `529126406`; the branch now contains all 30 upstream commits since `200602b41`.
- Saved the complete pre-sync worktree, including staged and untracked local customizations, in stash `077e946bf1f193c47de4111d9dd962f93f1471e4` (`pre-upstream-sync-20260901`) and restored it after the merge.
- Resolved the two overlapping files by keeping the upstream price-directory billing API and API-key request behavior while retaining the local media balance settlement, atomic usage persistence, and other existing customizations; updated four local billing test calls to the current helper signature.

### Testing

- `cd backend && go test ./... -count=1`: passed for the complete backend suite.
- `cd frontend && pnpm typecheck`: passed.
- `cd frontend && pnpm lint:check`: passed.
- `cd backend && gofmt -d internal/service/gateway_usage_billing.go internal/service/openai_gateway_service_hotpath_test.go internal/service/openai_gateway_record_usage_test.go`: passed with no formatting diff.
- `git diff --check`, `git diff --cached --check`, unresolved-index checks, strict conflict-marker scan, and `git merge-base --is-ancestor upstream/main HEAD`: passed.

### Notes

- `529126406`: merge commit containing the 70-file `upstream/main` update from `200602b41` through `0d27f45ea`, including v0.1.185 changes.
- `backend/internal/service/gateway_usage_billing.go`: reconciles the upstream data-driven token/image billing signatures with local atomic media settlement and usage persistence.
- `backend/internal/service/openai_gateway_service_hotpath_test.go`: keeps the upstream API-key missing-instructions regression coverage after conflict resolution.
- `backend/internal/service/openai_gateway_record_usage_test.go`: updates local image billing tests to the current single-result cost helper.
- `docs/UPSTREAM_SYNC.md`: documents the v0.1.185 sync scope, compatibility decisions, verification, and rollback point.
- `progress.md`: records this synchronization task and its test evidence.
- Rollback: first preserve the current worktree, then run `git revert -m 1 529126406`; restore the complete pre-sync worktree with `git stash apply --index 077e946bf1f193c47de4111d9dd962f93f1471e4` if needed. Keep the stash until deployment acceptance.

## 2026-09-02 - Task: Build and publish the latest china-api image

### What was done

- Built the current worktree as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, using application version `0.1.185`, commit `529126406888572072e73619ac998a0838803940`, and pushed the multi-architecture image to Docker Hub.
- The published OCI index is `sha256:7ee5def8f99b725cfe2f6de4b01dfac2d980065187931a64905c15d029d5b07a`.

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.1.185 --build-arg COMMIT=529126406888572072e73619ac998a0838803940 --build-arg DATE=2026-09-01T17:29:47Z --tag iotwq/china-api:latest --push .`: passed; frontend and both Go binaries built successfully and the image was pushed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed amd64 manifest `sha256:a76f0ae03801faa03f543c829501b126fafe395ea4e7b3ae112c000bb0f17735` and arm64 manifest `sha256:4f36858b777b3e0cf0efb42dd7e6bde9240f3623c53cb2afa5c9e19bbb2feba6` under the published index.
- `git diff --check` and `git diff --cached --check`: passed after publication.

### Notes

- `progress.md`: records the image build, remote digest verification, and architecture manifests.
- No source, database, running service, or existing local customization was changed by the publication step. Rollback: restore the previous multi-architecture index `sha256:d60bb469121a9c212e2cbdce93794509d69634538901933f9ca422a93c699e2e` with `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:d60bb469121a9c212e2cbdce93794509d69634538901933f9ca422a93c699e2e`.

## 2026-09-02 - Task: Fix usage log INSERT placeholder mismatch

### What was done

- Fixed both hand-written `usage_logs` INSERT paths so their 64 declared columns receive all 64 prepared arguments, including `session_id`, `native_compaction_v2`, and `created_at`.
- Tightened the single-row and no-result repository tests to require the `$61` through `$64` placeholder suffix, preventing this column/value mismatch from regressing.

### Testing

- `cd backend && go test ./internal/repository -run 'TestUsageLogRepositoryCreateSyncRequestTypeAndLegacyFields|TestUsageLogRepositoryCreate_PersistsServiceTier|TestExecUsageLogInsertNoResult_PersistsRequestedModel|TestPrepareUsageLogInsert_ArgCountMatchesTypes|TestPrepareUsageLogInsert_PersistsNativeCompactionV2WithoutChangingRequestType' -count=1`: passed.
- `cd backend && go test ./... -count=1`: passed for the complete backend suite.
- `gofmt -w backend/internal/repository/usage_log_repo_insert.go backend/internal/repository/usage_log_repo_request_type_test.go`: passed.
- `git diff --check` and `git diff --cached --check`: passed.

### Notes

- `backend/internal/repository/usage_log_repo_insert.go`: adds `$63` and `$64` to the single-row and no-result INSERT statements.
- `backend/internal/repository/usage_log_repo_request_type_test.go`: asserts both INSERT forms contain all four trailing placeholders.
- `progress.md`: records this repair and validation evidence.
- Rollback: remove only the `$63, $64` additions and the three strengthened SQL expectation patterns from this task, preserving all other working-tree changes and data.

## 2026-09-02 - Task: Build and publish usage-log INSERT fix

### What was done

- Built the current worktree, including the 64-parameter `usage_logs` INSERT fix, as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed the multi-architecture image to Docker Hub.
- The published OCI index is `sha256:93d38bae30e78db4018d5564b1b34b87857b5f7274e4fcc309367f1155910433`.

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.1.185 --build-arg COMMIT=529126406888572072e73619ac998a0838803940 --build-arg DATE=2026-09-01T18:19:07Z --tag iotwq/china-api:latest --push .`: passed after one transient Alpine package-index TLS retry; both backend targets and the embedded frontend completed successfully.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed amd64 manifest `sha256:3254624cb160bc95af71673c66a3ecd3edf26e24004b78b33b66751ef95c58d8` and arm64 manifest `sha256:c074357f29ed5436b93d01a7fe903f648e6596cf3cb24eb01142ff8b6b00af50` under the published index.
- `git diff --check` and `git diff --cached --check`: passed before publication.

### Notes

- `progress.md`: records the image publication, remote digest, and architecture manifests.
- No database, balance, running service, or unrelated worktree customization was changed by the build/push operation.
- Rollback: restore the preceding multi-architecture index `sha256:478e69768b611910445f358c688e27138ec2a2a9d3d42b1d07ecaea9b43e36ec` with `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:478e69768b611910445f358c688e27138ec2a2a9d3d42b1d07ecaea9b43e36ec`.

## 2026-09-03 - Task: Sync upstream/main through v0.2.0

### What was done

- Fetched `upstream/main` and merged the latest remote tip `b1748c4ea` into local `main` with merge commit `00f0f5e9c`; the branch now contains 58 upstream commits since `0d27f45ea` and version `0.2.0`.
- Saved the complete pre-sync worktree, including staged, unstaged, and untracked local customizations, in stash `52687120d5b2f652ef7bd426f438c7eb36cfb269` (`pre-upstream-sync-20260903`) and restored it after the merge. Resolved the single overlapping channel handler by retaining both upstream one-hour cache pricing and the local video billing mode.
- Regenerated Wire for the merged dependency graph and preserved the local media, video, community-chat, NewAPI balance, and public-resource injections.

### Testing

- `cd backend && go generate ./cmd/server`: passed.
- `cd backend && go test ./... -count=1`: passed for the complete backend suite.
- `cd frontend && pnpm typecheck`: passed.
- `cd frontend && pnpm lint:check`: passed.
- `cd frontend && pnpm exec vitest run --reporter=dot`: passed; 267 test files and 1927 tests.
- `cd frontend && pnpm build`: passed; only the existing Browserslist age and large-chunk warnings were reported.
- `git diff --check` and `git diff --cached --check`: passed; no unmerged index entries or conflict markers remain, and `upstream/main` is an ancestor of `HEAD`.

### Notes

- `00f0f5e9c`: merge commit containing the 58-commit `upstream/main` update from `0d27f45ea` through `b1748c4ea`, including v0.2.0.
- `backend/internal/handler/admin/channel_handler.go`: keeps upstream `cache_write_1h_price` request/response/default-pricing support together with the local `video` billing mode.
- `backend/cmd/server/wire_gen.go`: regenerated dependency injection while retaining local community chat, NewAPI balance access, video task binding/compensation, and public asset services.
- `docs/UPSTREAM_SYNC.md`: documents the v0.2.0 sync scope, migrations, compatibility decision, and verification.
- `progress.md`: records this synchronization task and its validation evidence.
- No database, balance, running service, or production task was modified by the synchronization.
- Rollback: first preserve the current worktree, then run `git revert -m 1 00f0f5e9c`; restore the pre-sync worktree with `git stash apply --index 52687120d5b2f652ef7bd426f438c7eb36cfb269` if needed. Keep the backup branch and stash until deployment acceptance.

## 2026-09-03 - Task: Build and publish latest image

### What was done

- Built the current synchronized worktree as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed the multi-architecture image to Docker Hub.
- Published OCI index: `sha256:1bcea716634dd32bb769c8c95785b8401cb9a50d50d5740ec203830b31e0bf52`.

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.0 --build-arg COMMIT=00f0f5e9cbcd826de98cdfa8ff9284ec0f821616 --build-arg DATE=2026-09-03T13:54:16Z --tag iotwq/china-api:latest --push .`: passed; frontend and both Go binaries built successfully and the image was pushed after transient registry retries.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed amd64 manifest `sha256:a1725c5ecb7f18a0dfc4baffcd5541d3f5c1f6f902a32f8731ee770a43a474b1` and arm64 manifest `sha256:4d46c036f5308b3479a4b796e8b6f350dc423c05d71625a8b0818d54c70f96fb` under the published index.
- `git diff --check` and `git diff --cached --check`: passed.

### Notes

- `progress.md`: records the image publication, remote digest, and architecture manifests.
- No database, balance, running service, or source customization was changed by the publication step. Rollback: restore the previous multi-architecture index `sha256:93d38bae30e78db4018d5564b1b34b87857b5f7274e4fcc309367f1155910433` with `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:93d38bae30e78db4018d5564b1b34b87857b5f7274e4fcc309367f1155910433`.

## 2026-09-05 - Task: Sync latest upstream/main

### What was done

- Fetched `upstream/main` and merged remote tip `ab99d56e9` into local `main` with merge commit `5b35db2c2`; the branch now contains 77 upstream commits since `b1748c4ea` and version `0.2.1`.
- Saved the complete pre-sync worktree in stash `pre-upstream-sync-20260905` (`stash@{0}`) and created backup branch `backup/pre-upstream-sync-20260905`; restored the local staged, unstaged, and untracked customizations after the merge.
- Resolved overlapping changes while retaining local community chat, direct messages, NewAPI balance access, MiniMax/Firefly video recovery and billing, audio/Nano Banana, image URL/Base64 conversion, custom API docs, and monitoring features. The merged usage-log INSERT paths now keep 65 columns and 65 placeholders, including upstream request ID and local video billing fields.

### Testing

- `cd backend && go generate ./cmd/server`: passed.
- `cd backend && go test ./... -count=1`: passed for the complete backend suite, including repository and service tests.
- `cd frontend && pnpm typecheck`: passed.
- `cd frontend && pnpm lint:check`: passed.
- `cd frontend && pnpm exec vitest run --reporter=dot`: passed; 270 test files and 1959 tests.
- `cd frontend && pnpm build`: passed.
- `git diff --check`, `git diff --cached --check`, unresolved-index checks, strict conflict-marker scan, and `git merge-base --is-ancestor upstream/main HEAD`: passed.

### Notes

- `5b35db2c2`: merge commit containing the 77-commit `upstream/main` update from `b1748c4ea` through `ab99d56e9`, including v0.2.1.
- `docs/UPSTREAM_SYNC.md`: records the v0.2.1 synchronization range, compatibility decisions, verification, and rollback point.
- `backend/internal/repository/usage_log_repo_insert.go`: retains upstream request identifiers and local video billing fields with matching 65-column INSERT statements.
- `backend/internal/service/openai_images_b64_backfill.go`: retains the upstream URL-to-Base64 backfill and private-destination protections alongside the local image conversion switch.
- No database, balance, running service, or production task was modified by this synchronization.
- Rollback: first preserve the current worktree, then run `git revert -m 1 5b35db2c2`; restore the pre-sync worktree with `git stash apply --index stash@{0}` (or the `pre-upstream-sync-20260905` entry) if needed. Keep the backup branch and stash until deployment acceptance.

## 2026-09-05 - Task: Build and publish latest china-api image

### What was done

- Built the current `0.2.1` worktree, including the local OpenAI image URL-to-Base64 conversion source, as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`, then pushed the multi-architecture image to Docker Hub.
- Published OCI index: `sha256:ed06937aced350c68b95c4d5c88d8ef0e4f66fcf8d6f024d2393f9c0607ac20c`.

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.1 --build-arg COMMIT=5b35db2c2f16691ea9628b5d4dfa5ad81eb76c2c --build-arg DATE=2026-09-05T13:41:48Z --tag iotwq/china-api:latest --push .`: passed; frontend and both Go binaries built successfully and the image was pushed.
- `docker buildx imagetools inspect iotwq/china-api:latest`: confirmed amd64 manifest `sha256:8cd1f4161488c0c9ed8c0856c24bb90d482c6411074b881874e69d7bb90da0d7` and arm64 manifest `sha256:2024778bdbb62d288823730000c25402014a58a7b446404cd544111984727d8d` under the published index.
- `git diff --check` and `git diff --cached --check`: passed.

### Notes

- `progress.md`: records the image publication, remote digest, and architecture manifests.
- No database, balance, running service, or source customization was changed by the publication step. Rollback: restore the preceding multi-architecture index `sha256:1bcea716634dd32bb769c8c95785b8401cb9a50d50d5740ec203830b31e0bf52` with `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:1bcea716634dd32bb769c8c95785b8401cb9a50d50d5740ec203830b31e0bf52`.

## 2026-09-06 - Task: OpenAI OAuth 临时错误自动重试

### What was done

- 修复 OpenAI OAuth/Setup Token 的 passthrough 临时 5xx 直接返回客户端问题，并让这些请求复用现有同账号有界重试和账号切换流程。
- 普通临时 5xx 最多额外重试 3 次；保留瞬时 429 的已有恢复窗口、配额耗尽冷却、鉴权和安全策略处理，不改变 API Key/Bedrock 池模式或其他渠道。
- 核对前端后撤回本轮早期扩大 `IsPoolMode()` 的尝试：OAuth 页面没有池模式入口，而且共用池模式会绕过真实凭据的状态维护。最终无需新开关，`account.go` 没有本轮改动。

### Testing

- `cd backend && go test -tags unit ./internal/service ./internal/handler -run 'Test(IsPoolModeScope|OpenAIOAuthPassthrough|OpenAIOAuthTransientRetryClassification|OpenAIResponses_OAuthTransientRetries|OpenAIGatewayService_OpenAIPassthrough_RetryableStatusesTriggerFailover)' -count=1`：服务层通过；处理器初次因模拟成功响应不符合 OAuth SSE 协议失败，修正测试数据后 `go test ./internal/handler -run '^TestOpenAIResponses_OAuthTransientRetries$' -count=1` 通过。
- 首轮 service/handler 整包发现生图 `image_generation_unavailable` 的本地 502 不应启用同账号重试，已保持其直接切号语义；两个真实上游故障用例更新为先重试再切号。`go test -tags unit ./internal/service ./internal/handler -run 'Test(OpenAIGatewayHandlerImages_ServerErrorFailsOverAndReturnsClearErrorWhenExhausted|OpenAIGatewayHandlerResponses_Failover|OpenAIGatewayServiceForwardImages_.*|OpenAIResponses_OAuthTransientRetries)' -count=1` 通过。
- `cd backend && go test ./... -run '^$' -count=1` 全项目编译检查通过，49 个有测试包编译通过，无失败；机器可读日志 `/tmp/sub2api-oauth-retry-compile-20260906.jsonl`。
- 最终 `cd backend && go test -tags unit ./... -count=1`：54 个测试包通过，包括 `internal/service`（176.254s）和 `internal/handler`（50.446s）；仅 `internal/repository` 的 `TestPrepareUsageLogInsert_SessionIDArgWiring`、`TestPrepareUsageLogInsert_RequestedReasoningEffortArgWiring` 失败，分别硬编码 64 列但当前 65 列、以及旧字段下标。两项用 `go test -tags unit ./internal/repository -run '^TestPrepareUsageLogInsert_(SessionIDArgWiring|RequestedReasoningEffortArgWiring)$' -count=1` 可独立复现；`git diff -- backend/internal/repository` 为空，本轮未修改账务源码或这些测试。完整机器可读日志 `/tmp/sub2api-oauth-retry-final-20260906.jsonl`。全量测试不能报告为全部通过。
- `git diff --check` 及本轮 Go 文件 `gofmt -l` 检查通过。未连接生产上游、修改数据库、构建推送镜像或重启服务。

### Notes

- `backend/internal/service/openai_gateway_passthrough.go`：允许 OpenAI OAuth/Setup Token 的临时错误进入故障转移，保留请求级错误拦截。
- `backend/internal/service/openai_gateway_upstream_errors.go`：为 OAuth 临时错误启用现有有界同账号重试，不改变 429 专用处理。
- `backend/internal/service/openai_images_responses.go`：保留生图能力不足/只有文字的直接切号行为，不将本地 502 纳入新增重试。
- `backend/internal/service/account_pool_mode_test.go`：确保 OAuth 不会进入跳过账号状态维护的池模式。
- `backend/internal/service/openai_access_state_failover_test.go`：验证临时 5xx、过载、配额、鉴权和平台边界。
- `backend/internal/service/openai_oauth_passthrough_test.go`：更新旧的 OAuth 502/503/504 直接返回预期。
- `backend/internal/handler/openai_gateway_handler_test.go`：模拟 OAuth 上游失败后成功、重试耗尽和跨账号恢复。
- `backend/internal/handler/openai_images_failover_test.go`：更新真实上游生图错误的重试次数与日志预期。
- `backend/internal/handler/openai_responses_failover_cancel_test.go`：更新在线客户端重试后切号预期，保留断连立即停止的断言。
- `docs/OPENAI_OAUTH_RETRY.md`：说明自动生效方式、重试边界及验收命令。
- `.gitignore`：仅将本轮新增说明加入文档白名单。
- `progress.md`：追加本轮实现及验证记录。
- 回滚点是本轮开始前的 Git index，原有定制均保持暂存且未改动。未重新暂存且已保留后续变更时，可用 `git restore --worktree -- .gitignore backend/internal/service/openai_gateway_passthrough.go backend/internal/service/openai_gateway_upstream_errors.go backend/internal/service/openai_images_responses.go backend/internal/service/account_pool_mode_test.go backend/internal/service/openai_access_state_failover_test.go backend/internal/service/openai_oauth_passthrough_test.go backend/internal/handler/openai_gateway_handler_test.go backend/internal/handler/openai_images_failover_test.go backend/internal/handler/openai_responses_failover_cancel_test.go progress.md` 回滚本轮已跟踪文件；新增文档可移到仓库外保留。不涉及数据库迁移。

## 2026-09-06 - Task: 优化入口首页与登录页视觉体验

### What was done

- 为认证布局增加轻量网格背景、环形层次和工具栏/表单面板渐入效果，保持现有深浅色主题与响应式结构。
- 为登录页增加标题、字段、验证码/提交区和第三方登录区的错峰入场，输入聚焦图标反馈，以及登录和 OAuth 按钮的悬停/聚焦阴影反馈；未改动认证、验证码、OAuth、2FA 或路由逻辑。
- 为首页导航、Hero、Agent 节点、执行轨迹和研究方向增加分层渐入、错峰动画与研究卡片 hover 反馈，并统一支持 `prefers-reduced-motion` 降级。

### Testing

- `cd frontend && pnpm test:run`：通过，270 个测试文件、1959 条断言。
- `cd frontend && pnpm typecheck`：通过。
- `cd frontend && pnpm build`：通过，Vite 生产构建完成。
- 使用本地 Vite 预览检查桌面首页与登录页截图：布局、表单、导航和首屏动效渲染正常；公共配置请求因未启动后端产生预期网络错误，不影响静态页面验证。
- `git diff --check`：通过。

### Notes

- `frontend/src/components/layout/AuthLayout.vue`：增加认证工作区背景层与进入动效，并保持移动端断点。
- `frontend/src/views/auth/LoginView.vue`：增加登录内容错峰入场、焦点反馈和按钮交互状态。
- `frontend/src/views/HomeView.vue`：增加首页首屏/研究区域分层动效与 hover 反馈。
- `progress.md`：记录本轮 UI 改动与验证结果。
- 回滚：恢复上述三个前端文件到本轮开始前版本，并保留或删除本轮日志段；不要回滚其他并行修改。

## 2026-09-06 - Task: Build and publish latest china-api image

### What was done

- Built the current `0.2.1` worktree, including the latest homepage and login visual enhancements, as `iotwq/china-api:latest` for `linux/amd64` and `linux/arm64`.
- Pushed the multi-architecture image to Docker Hub with build metadata `5b35db2c2f16-dirty` and UTC build time `2026-09-06T12:25:57Z`.

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.1 --build-arg COMMIT=5b35db2c2f16-dirty --build-arg DATE=2026-09-06T12:25:57Z --tag iotwq/china-api:latest --push .`：通过；前端和两个 Go 目标均构建成功并完成推送。首次 Docker Hub 授权 `EOF` 后重试成功。
- `docker buildx imagetools inspect iotwq/china-api:latest`：确认远端 OCI index `sha256:dbf743a1a42d8ece5e134ff6bf8c0a211170bb2a0ef358e1d6e52d382c10c11b`，amd64 manifest `sha256:71e23a0f6ca0fba6fe60181503b167f1e6f3528987bfc3f369093b74dea7dd7b`，arm64 manifest `sha256:058f5c01f8d933de52805891a2e19da446995c1bc568b34d041cfda2a02ed31f`。

### Notes

- `progress.md`：记录本次镜像发布、远端 digest 和架构 manifest。
- 本次仅构建和推送镜像，未修改数据库、余额、运行中服务或其他源代码。
- 回滚：使用 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:ed06937aced350c68b95c4d5c88d8ef0e4f66fcf8d6f024d2393f9c0607ac20c` 恢复上一版多架构 index。

## 2026-09-06 - Task: 增强入口与登录页动态视觉效果

### What was done

- 在首页 Hero 场景增加低对比度扫光、中心节点环形脉冲、CTA 光泽掠过和账户入口 hover 层次。
- 在研究方向区域增加彩色顶部强调线、图标轻微反馈和更明确的 hover 背景。
- 在登录表单增加字段焦点强调条，并为认证场景增加低频扫描光带；所有新增动画均支持 `prefers-reduced-motion`，移动端对中心环形效果进行了缩小。

### Testing

- `cd frontend && pnpm exec eslint src/components/layout/AuthLayout.vue src/views/auth/LoginView.vue src/views/HomeView.vue`：通过。
- `cd frontend && pnpm typecheck`：通过。
- `cd frontend && pnpm test:run`：通过，270 个测试文件、1959 条断言。
- `cd frontend && pnpm build`：通过，Vite 生产构建完成。
- 使用本地 Vite 预览检查首页与登录页截图：扫光、环形脉冲、CTA 光泽、焦点强调和桌面布局均正常；公共配置接口因未启动后端产生连接错误，不影响静态视觉验证。
- `git diff --check`：通过。

### Notes

- `frontend/src/views/HomeView.vue`：增加 Hero 扫光、核心节点脉冲、CTA 光泽和研究轨道交互反馈。
- `frontend/src/components/layout/AuthLayout.vue`：增加认证场景扫光和表单容器层次。
- `frontend/src/views/auth/LoginView.vue`：增加登录字段焦点强调条。
- `progress.md`：记录本轮动态视觉增强及验证结果。
- 回滚：恢复上述三个前端文件到本轮开始前版本，并保留或删除本轮日志段；不要回滚其他并行修改。

## 2026-09-07 - Task: 同步原始上游最新提交至 v0.2.2

### What was done

- 合并原始上游 `Wei-Shaw/sub2api` 的 `ab99d56e9..b7dba6267`，共 102 个提交、261 个文件；本地合并提交为 `37b98f34f75ecef744f377735f879a636578e94d`。
- 同步前保存基线分支和完整工作区 stash，合并后恢复本地定制并处理 4 个冲突；保留图片转换、视频恢复/退款、音频、社区聊天、余额查询、OAuth 重试及最新首页/登录视觉效果。
- 适配新版账号切换参数与依赖注入，使本地扩展路由接入上游分组白名单，并补齐合并相关测试和翻译文案。
- 记录新增迁移和分组白名单由展示限制升级为调用限制的部署影响；未构建或推送镜像，未推送 Git，未修改生产数据库或重启服务。

### Testing

- `cd backend && go generate ./cmd/server`：通过。
- `cd backend && go test ./... -count=1`：最终全量通过；日志 `/tmp/sub2api-sync-20260907-backend-verified.log`。首轮构造参数缺失和第二轮白名单测试的未定价模型冲突均修正并重测。
- `cd backend && go test -tags=unit ./internal/server -run '^TestAPIContracts$' -count=1`：通过，验证合并后的 API 契约依赖。
- 新增 14 项本地视频、音频、Nano Banana 根路径及 `/v1` 白名单路由用例，在全量后端测试中通过；两项 WebSocket 白名单测试使用已有定价且不在白名单内的模型，拒绝与关闭白名单后的放行断言均通过。
- `cd frontend && pnpm install --frozen-lockfile`、`pnpm lint:check`、`pnpm exec vitest run`、`pnpm build`：全部通过，278 个测试文件、1998 项测试；构建包含类型和翻译完整性检查，仍有大 chunk 提示。
- `git diff --check`、`git diff --cached --check`、未解决索引、严格冲突标记、`git merge-base --is-ancestor upstream/main HEAD` 及已编辑 Go 测试格式检查通过。
- 原有 92 个未跟踪文件无丢失，91 个内容哈希未变，视频 service 仅适配新参数；首页、登录页和认证布局与同步前 stash 比较无差异。未执行所有带标签测试或真实上游端到端请求。

### Notes

- 上游变更文件清单：`git diff --name-only ab99d56e9 b7dba6267`；通过合并提交纳入上游原始改动，未逐项重写。
- `backend/internal/service/openai_gateway_cc_pipeline.go`：合并账号感知切换与本地包装错误、监控重试判断。
- `backend/internal/service/openai_gateway_forward.go`：保留本地重试状态处理并适配上游账号参数。
- `backend/internal/service/openai_nano_banana.go`：补齐上游新增的账号参数。
- `backend/internal/service/openai_videos.go`：补齐上游新增的账号参数。
- `backend/internal/server/routes/gateway.go`：保留本站媒体和资源接口，并复用上游白名单根路由链。
- `backend/internal/server/api_contract_test.go`：同时保留上游管理员服务参数和本站余额 Token 依赖。
- `backend/cmd/server/wire_gen.go`：重新生成上游与本站服务的依赖注入。
- `backend/internal/handler/openai_codex_models_handler_test.go`：补齐本站视频仓储构造参数。
- `backend/internal/server/routes/gateway_models_pinned_test.go`：补齐本站视频仓储构造参数。
- `backend/internal/server/routes/gateway_model_allowlist_test.go`：增加本站媒体路由白名单回归用例。
- `backend/internal/handler/openai_gateway_ws_model_allowlist_test.go`：两项测试改用已有定价模型，独立验证白名单行为。
- `frontend/src/i18n/locales/en/admin/settings.ts`：补齐两项本站支付配置英文文案。
- `frontend/src/i18n/locales/zh/admin/settings.ts`：补齐两项本站支付配置中文文案。
- `frontend/src/views/admin/__tests__/GroupsView.codexManifest.spec.ts`：补齐本站登录状态 mock。
- `docs/UPSTREAM_SYNC.md`：追加同步范围、验证、部署注意及回滚说明。
- `progress.md`：追加本轮记录；不改写历史。
- 原暂存/未暂存划分未完全恢复，提交前须重新检查暂存区；本地定制和本轮兼容修正未另行提交。
- 回滚点：分支 `codex/pre-upstream-sync-20260907` 和 stash `0aa535a2c26d7a41efddfd552d74aaf832e076ac`。可执行恢复：`git worktree add -b codex/recover-pre-sync-20260907 /tmp/sub2api-pre-sync-20260907 codex/pre-upstream-sync-20260907`，随后 `git -C /tmp/sub2api-pre-sync-20260907 stash apply --index 0aa535a2c26d7a41efddfd552d74aaf832e076ac`；在独立目录恢复完整同步前状态，不覆盖当前工作区。不要删除该 stash。

## 2026-09-07 - Task: 构建 v0.2.2 镜像并尝试推送 Docker Hub

### What was done

- 基于当前完整工作区完成 `linux/amd64` 和 `linux/arm64` 的前端及 Go 后端构建，版本 `0.2.2`，构建标识 `37b98f34f75e-dirty`，时间 `2026-09-07T15:06:55Z`。
- 尝试推送 `iotwq/china-api:latest`，Docker Hub 返回 `insufficient_scope: authorization failed`，发布未完成。当前使用 desktop 凭据存储，其凭据列表为空，需用户重新登录有推送权限的 Docker Hub 账号后继续。
- 本轮未修改业务代码；模型同步能力元数据提示仅在上一轮分析，未修复，因此本次构建不包含该修复。未部署或重启服务。

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.2 --build-arg COMMIT=37b98f34f75e-dirty --build-arg DATE=2026-09-07T15:06:55Z --tag iotwq/china-api:latest --metadata-file /tmp/sub2api-image-20260907-metadata.json --push .`：前端类型检查、翻译检查、生产构建及双架构 Go 编译均通过，最终 push 阶段退出 1；日志 `/tmp/sub2api-image-20260907-build.log`。
- 构建阶段生成但未成功发布的 index 为 `sha256:b7b463d4eb9aa6430c0b1c8b9c0352e79b5b86063b7e3c9f784e43818dfbb6bd`，不能作为可拉取交付物。
- `docker buildx imagetools inspect iotwq/china-api:latest`：推送失败后重新检查，远端仍为旧摘要 `sha256:dbf743a1a42d8ece5e134ff6bf8c0a211170bb2a0ef358e1d6e52d382c10c11b`；一次匿名令牌请求 EOF 后重试查询成功。
- `git diff --check` 和 `git diff --cached --check`：通过。

### Notes

- `progress.md`：追加构建成功、推送受阻及远端版本未变化的记录。
- 继续方式：用户执行 `docker login` 登录有 `iotwq/china-api` 推送权限的账号后，重新执行上述 buildx 命令，复用已有构建缓存。
- 回滚：远端 latest 未变化，无需回滚；旧版可继续通过 `iotwq/china-api@sha256:dbf743a1a42d8ece5e134ff6bf8c0a211170bb2a0ef358e1d6e52d382c10c11b` 拉取。

## 2026-09-07 - Task: 登录后完成 v0.2.2 镜像推送

### What was done

- 用户重新登录 Docker Hub 后，复用上轮构建缓存，成功推送 `iotwq/china-api:latest`，包含 `linux/amd64` 和 `linux/arm64`。
- 发布版本为 `0.2.2`，构建标识 `37b98f34f75e-dirty`，构建时间 `2026-09-07T15:06:55Z`；未修改业务代码、数据库或运行中的服务。

### Testing

- 重跑上轮完整 buildx 命令成功退出 0，镜像层与 manifest 均推送成功；日志 `/tmp/sub2api-image-20260907-push-retry.log`，构建元数据 `/tmp/sub2api-image-20260907-metadata.json`。
- `docker buildx imagetools inspect iotwq/china-api:latest`：确认远端 index 为 `sha256:b7b463d4eb9aa6430c0b1c8b9c0352e79b5b86063b7e3c9f784e43818dfbb6bd`，与本次构建元数据一致。
- 确认 amd64 manifest 为 `sha256:9b112de450cd87caff1faff0c27068d89c7b79ece4ad5b9127448b23b144d566`，arm64 manifest 为 `sha256:f6538dc44d6c90225c0e75b4ea5d3a6512dc6009ce4b385ddc46e33cdfa2bc14`。
- `git diff --check`：通过。前端检查和双架构编译复用上一轮成功构建缓存；本轮未执行生产运行验收。

### Notes

- `progress.md`：追加本轮发布成功、远端摘要及回滚记录，不改写前一轮推送失败历史。
- 账号模型能力元数据提示问题仍未修复；分组模型白名单的部署注意事项见 `docs/UPSTREAM_SYNC.md`。
- 回滚 latest：`docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:dbf743a1a42d8ece5e134ff6bf8c0a211170bb2a0ef358e1d6e52d382c10c11b`；运行环境也可直接固定该旧摘要。本轮未执行回滚或部署。

## 2026-09-08 - Task: 修复账号测试模型选项名称空白

### What was done

- 测试弹窗加载模型后为缺失、空值或空白显示名称回退使用模型 ID，保留有效显示名称，使下拉选项、选中项及搜索正常显示和工作。
- 使用真实 Select 组件为 OAuth、API Key 两类账号补充显示、搜索、选择和提交测试的回归验证；测试请求仍携带原始模型 ID。
- 保留模型白名单、实时目录及账号调用逻辑；本轮未修改后端、构建 Docker 镜像或部署运行服务。

### Testing

- 修复前：`cd frontend && pnpm exec vitest run src/components/account/__tests__/AccountTestModal.spec.ts` 新增两项测试均以显示名称为空而失败，稳定复现故障；日志 `/tmp/sub2api-account-picker-before-20260908.log`。
- 修复后：`cd frontend && pnpm exec vitest run src/components/account/__tests__/AccountTestModal.spec.ts src/components/common/__tests__/Select.spec.ts` 通过，2 个文件/12 项测试；日志 `/tmp/sub2api-account-picker-after-20260908.log`。
- `cd frontend && pnpm exec eslint src/components/account/AccountTestModal.vue src/components/account/__tests__/AccountTestModal.spec.ts` 通过。
- `cd frontend && pnpm build` 通过，包含翻译完整性检查、TypeScript 构建检查和 Vite 生产构建；保留已有大 chunk 提示，日志 `/tmp/sub2api-account-picker-build-20260908.log`。
- `git diff --check` 通过。验证使用模拟模型响应和测试请求，未对生产账号执行付费请求或现场浏览器验收。

### Notes

- `frontend/src/components/account/AccountTestModal.vue`：加载模型时补齐用于展示和搜索的名称。
- `frontend/src/components/account/__tests__/AccountTestModal.spec.ts`：新增 OAuth/API Key 的真实下拉框交互回归测试。
- `docs/UPSTREAM_SYNC.md`：追加上游目录变更引发的空白选项问题、修复行为与部署说明。
- `progress.md`：追加本轮修复与验证记录。
- 回滚：上述两个前端文件本轮修改前内容仍保存在暂存区；在未暂存本轮修改且没有后续编辑的前提下，执行 `git restore --worktree -- frontend/src/components/account/AccountTestModal.vue frontend/src/components/account/__tests__/AccountTestModal.spec.ts` 可仅撤销本轮代码和测试改动，保留之前已暂存的本地定制；文档日志保留作为历史记录。

## 2026-09-08 - Task: 修正账号管理实际使用的测试弹窗

### What was done

- 更正上一轮修复落在旧组件的遗漏，在账号管理实际加载的弹窗中为缺失或空白名称回退显示模型 ID。
- 为 OAuth/API Key 两类账号验证真实下拉框的显示、搜索、选择和提交；模型 ID、测试协议及上游调用方式保持原有行为。

### Testing

- 修复前，实际弹窗新增两项测试均以空白名称失败；日志 `/tmp/sub2api-merge-fixes-20260908.oAldZW/picker-before.log`。
- `cd frontend && pnpm exec vitest run src/components/admin/account/__tests__/AccountTestModal.spec.ts src/components/common/__tests__/Select.spec.ts`：14 项通过；日志 `/tmp/sub2api-merge-fixes-20260908.oAldZW/picker-after.log`。
- `cd frontend && pnpm exec eslint src/components/admin/account/AccountTestModal.vue src/components/admin/account/__tests__/AccountTestModal.spec.ts`：通过。使用模拟模型接口和测试请求，未连接生产账号。

### Notes

- `frontend/src/components/admin/account/AccountTestModal.vue`：在实际使用的弹窗中补齐模型显示名称。
- `frontend/src/components/admin/account/__tests__/AccountTestModal.spec.ts`：新增两类账号的真实 Select 交互回归测试。
- `docs/UPSTREAM_SYNC.md`：追加旧修复未生效的更正与实际修复行为。
- `progress.md`：追加本轮记录，不改写历史。
- 回滚：在仓库根目录执行 `tar -xf /tmp/sub2api-merge-fixes-20260908.oAldZW/before.tar frontend/src/components/admin/account/AccountTestModal.vue frontend/src/components/admin/account/__tests__/AccountTestModal.spec.ts`，恢复本轮施工前内容并保留此前定制；如已有后续编辑，应先核对差异。文档与日志保留历史记录。本轮未发布镜像或部署。

## 2026-09-08 - Task: 修正合并后的用量字段测试断言

### What was done

- 两项旧断言更新为当前 65 个字段及请求/实际推理等级的位置，保留合并后的上游请求 ID 与本站视频成本字段。
- 仅修复测试断言，未修改 SQL、数据库结构、余额或记账流程。

### Testing

- 上轮完整带标签测试已稳定复现两项失败；证据 `/tmp/sub2api-merge-audit-unit-20260908.log`。
- `cd backend && go test -tags=unit ./internal/repository -count=1`：全部通过，包含 SQL 列数/占位符匹配及用量字段位置测试；日志 `/tmp/sub2api-merge-fixes-20260908.oAldZW/repository-after.log`。
- `gofmt -d backend/internal/repository/usage_log_session_id_unit_test.go` 无差异，`git diff --check` 通过；未进行生产数据库验账。

### Notes

- `backend/internal/repository/usage_log_session_id_unit_test.go`：修正字段总数及推理等级参数索引断言。
- `docs/UPSTREAM_SYNC.md`：追加用量合并后的验证要求，明确需要 `unit` 标签。
- `progress.md`：追加本轮记录。
- 回滚：在仓库根目录执行 `tar -xf /tmp/sub2api-merge-fixes-20260908.oAldZW/before.tar backend/internal/repository/usage_log_session_id_unit_test.go` 恢复施工前测试内容；如已有后续编辑，应先核对差异。文档与日志保留历史记录。

## 2026-09-08 - Task: 修正新增账号模型能力同步预览提示

### What was done

- 根据实际取得的完整能力信息区分部分成功与不完整，修复未保存账号因未落库而误报全部能力信息未更新的问题。
- 中英文提示统一描述获取结果，避免预览宣称已经保存；保留已有账号持久化行为、真实不完整提示及获取来源和网络请求方式。

### Testing

- 新增完整/部分/不完整三种预览场景；修复前部分场景稳定失败，日志 `/tmp/sub2api-merge-fixes-20260908.oAldZW/metadata-before.log`。
- `cd backend && go test -tags=unit ./internal/service ./internal/handler/admin -run 'Test(SyncUpstreamModelCatalog|AccountHandlerSyncUpstreamModels)' -count=1`：通过，覆盖预览不持久化、部分信息保存、保存失败及已有快照保留；日志 `/tmp/sub2api-merge-fixes-20260908.oAldZW/metadata-after.log`。
- `cd frontend && pnpm exec vitest run src/components/account/__tests__/ModelWhitelistSelector.spec.ts src/components/account/__tests__/CreateAccountModal.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts src/i18n/__tests__/localeKeyCompleteness.spec.ts`：104 项通过；日志 `/tmp/sub2api-merge-fixes-20260908.oAldZW/metadata-ui.log`。
- `cd frontend && pnpm exec eslint src/i18n/locales/zh/admin/accounts.ts src/i18n/locales/en/admin/accounts.ts`：通过。修改的 Go 文件 `gofmt -d` 无差异。
- `cd frontend && pnpm build`：通过，包含翻译完整性、TypeScript 检查及生产打包；日志 `/tmp/sub2api-merge-fixes-20260908.oAldZW/frontend-build.log`。保留既有大 chunk 提示；本轮未执行生产账号请求、数据库验账或部署。

### Notes

- `backend/internal/service/upstream_models.go`：同步提示按完整能力信息判断，不依赖持久化状态。
- `backend/internal/service/upstream_models_test.go`：新增三种未保存账号的同步预览回归场景。
- `frontend/src/i18n/locales/zh/admin/accounts.ts`：中文提示区分取得部分完整信息与信息仍不完整。
- `frontend/src/i18n/locales/en/admin/accounts.ts`：同步英文提示语义。
- `docs/UPSTREAM_SYNC.md`：追加预览行为、提示含义及验证和部署说明。
- `progress.md`：追加本轮记录。
- 回滚：在仓库根目录执行 `tar -xf /tmp/sub2api-merge-fixes-20260908.oAldZW/before.tar backend/internal/service/upstream_models.go backend/internal/service/upstream_models_test.go frontend/src/i18n/locales/zh/admin/accounts.ts frontend/src/i18n/locales/en/admin/accounts.ts` 恢复施工前源码；如已有后续编辑，应先核对差异。文档与日志保留历史记录，回滚后重新构建前端以更新构建产物。

## 2026-09-08 - Task: 构建并推送包含合并遗漏修复的新镜像

### What was done

- 基于当前完整工作区构建并成功推送 `iotwq/china-api:latest`，包含最新实际账号测试弹窗及模型能力同步提示修复，保留本站定制功能。
- 版本为 `0.2.2`，构建标识 `37b98f34f75e-dirty`，构建时间 `2026-09-07T18:01:06Z`（北京时间 2026-09-08 02:01:06），支持 `linux/amd64` 和 `linux/arm64`。
- Docker Hub 的 latest 已更新为 `sha256:0c68a6472f7082d321af781d4cbce88aa8a3d2e361123f967c2d480568c4ca2e`。本轮未修改业务源码、数据库或运行中的服务。

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.2 --build-arg COMMIT=37b98f34f75e-dirty --build-arg DATE=2026-09-07T18:01:06Z --tag iotwq/china-api:latest --metadata-file /tmp/sub2api-image-20260908.8SjZuY/metadata.json --progress=plain --push .`：成功退出 0，包含镜像内翻译/类型检查、前端生产构建和双架构 Go 编译；日志 `/tmp/sub2api-image-20260908.8SjZuY/build.log`。
- `docker buildx imagetools inspect iotwq/china-api:latest`：远端 index 与构建元数据一致；amd64 manifest 为 `sha256:ec7ef174004f0a6e3e58469d2eb4e2fe32b4348a837f3f7e7438d900fc7e8f6c`，arm64 manifest 为 `sha256:a54f053f8e6dd941bbfb378bb305299980f452edd26708eb203d70992f6b1b41`。
- 分别使用 `docker run --rm --platform linux/arm64 --network none --read-only --entrypoint /app/sub2api iotwq/china-api@sha256:0c68a6472f7082d321af781d4cbce88aa8a3d2e361123f967c2d480568c4ca2e --version` 及 `--platform linux/amd64` 执行镜像版本检查：均成功退出 0，版本、构建标识和时间一致；日志 `/tmp/sub2api-image-20260908.8SjZuY/version-arm64.log`、`/tmp/sub2api-image-20260908.8SjZuY/version-amd64.log`。
- 源码功能验证沿用紧邻上一轮通过的前端 118 项相关测试、后端完整带标签仓储测试及模型同步 service/handler 测试；本轮没有新增业务修改，未重复全量测试。未执行生产数据库验账或服务健康验收。

### Notes

- `progress.md`：追加本轮构建、推送、双架构版本检查及回滚记录。
- 部署时拉取 `iotwq/china-api:latest` 并重新创建服务容器；也可固定本轮 index 摘要，避免后续 latest 变化。本轮未执行部署。
- 回滚点为发布前镜像 `iotwq/china-api@sha256:b7b463d4eb9aa6430c0b1c8b9c0352e79b5b86063b7e3c9f784e43818dfbb6bd`。恢复 latest 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:b7b463d4eb9aa6430c0b1c8b9c0352e79b5b86063b7e3c9f784e43818dfbb6bd`；该命令仅作为回滚说明，本轮未执行。

## 2026-09-08 - Task: 同步上游 v0.2.3 并保留本站定制

### What was done

- 将 `upstream/main` 的 `b7dba6267..270eac697` 共 76 个提交、271 个变更文件合并到 main，合并提交 `563aba27b28dfcaaec52c42244832556b781d8b3`，版本 `0.2.3`。
- 保存完整恢复点后解决三处内容冲突，同时保留本站可计费用量处理、弹窗按需加载及登录视觉样式，并纳入上游断线处理、菜单定位和注册入口控制。
- 保留已修复的实际模型测试弹窗、用量字段断言和能力同步提示；92 个未跟踪定制文件内容一致。上游新增后端显示名称回退与前端回退共存。
- 新增 MiniMax 平台和白名单修复迁移已通过隔离验证；未修改生产数据库、余额或运行服务，未推送 GitHub 或发布镜像。

### Testing

- `cd backend && go generate ./cmd/server`：通过；日志 `/tmp/sub2api-sync-20260908.8q7Qk9/wire.log`。
- `cd backend && go test -tags=unit ./... -count=1`：全量通过，包含用量 SQL 结构、媒体恢复/退款、OAuth 重试及模型同步测试；日志 `/tmp/sub2api-sync-20260908.8q7Qk9/backend-unit.log`。
- `cd frontend && pnpm lint:check` 和 `pnpm build`：通过，构建包含翻译和类型检查；日志 `/tmp/sub2api-sync-20260908.8q7Qk9/frontend-lint.log`、`frontend-build.log`。
- `cd frontend && pnpm exec vitest run`：283 个文件/2076 项，首跑仅本地渠道数量硬编码断言失败，2075 项通过；修正后 `pnpm exec vitest run src/views/admin/__tests__/ChannelMonitorView.grok.spec.ts` 两项通过，定向 ESLint 通过。日志 `/tmp/sub2api-sync-20260908.8q7Qk9/frontend-tests.log`、`frontend-tests-recheck.log`；未重跑其余已通过用例。
- `cd backend && go test -tags=integration ./internal/repository -run '^(TestMigration236|TestMigrationsRunner_IsIdempotent_AndSchemaIsUpToDate)' -count=1 -v`：通过，测试自行创建隔离 PostgreSQL/Redis 并应用全部迁移，三个 236 修复场景及迁移幂等性验证通过；日志 `/tmp/sub2api-sync-20260908.8q7Qk9/migrations-integration.log`。
- `bash deploy/tests/apple-container-test.sh`：模拟 Apple 容器生命周期测试通过，未操作真实服务；日志 `/tmp/sub2api-sync-20260908.8q7Qk9/apple-container-test.log`。
- 未解决索引为空；严格冲突标记检查、`git diff --check`、`git diff --cached --check`、上游提交祖先检查通过。原 92 个未跟踪文件 SHA-256 全部一致，清单位于 `/tmp/sub2api-sync-20260908.8q7Qk9/untracked-before.sha256`；上游范围外已跟踪内容仅本轮渠道测试和记录有变化。未执行生产端到端或性能验证。

### Notes

- 上游文件清单可执行 `git diff --name-only 37b98f34f75ecef744f377735f879a636578e94d 563aba27b28dfcaaec52c42244832556b781d8b3` 查看；上游原始改动由合并提交纳入。
- `backend/internal/handler/openai_gateway_handler.go`：合并客户端断线提交用量与本站完整可计费用量判断。
- `frontend/src/views/admin/AccountsView.vue`：保留弹窗按需加载，使用上游菜单锚点属性。
- `frontend/src/views/auth/LoginView.vue`：保留本站登录样式，合并公共设置加载及注册开关条件。
- `backend/cmd/server/wire_gen.go`：重新生成并核对上游渠道缓存通知与本站服务的依赖注入。
- `frontend/src/views/admin/__tests__/ChannelMonitorView.grok.spec.ts`：渠道数量断言适配 MiniMax 新增平台。
- `docs/UPSTREAM_SYNC.md`：追加本轮同步范围、兼容处理、验证和部署注意。
- `progress.md`：追加本轮记录，不改写历史。
- 恢复点：分支 `codex/pre-upstream-sync-20260908` 与完整 stash `591a582408eabfb6ee36ff0d44ce91bfc7e20ee6`，均保留。初次 `stash apply --index` 因上下文变化失败；按内容恢复并解决冲突后，`git apply --cached --reverse /tmp/sub2api-sync-20260908.8q7Qk9/worktree-before.patch` 成功恢复原未暂存改动的索引划分，不改变工作区内容。本轮兼容修正未单独提交。
- 可执行回滚：`git worktree add -b codex/recover-pre-sync-20260908 /tmp/sub2api-pre-sync-20260908 codex/pre-upstream-sync-20260908`，随后 `git -C /tmp/sub2api-pre-sync-20260908 stash apply --index 591a582408eabfb6ee36ff0d44ce91bfc7e20ee6`，在独立目录恢复完整同步前内容，不覆盖当前工作区；本轮未执行回滚或生产迁移。

## 2026-09-08 - Task: 构建并推送 v0.2.3 双架构镜像

### What was done

- 基于同步后的完整工作区成功构建并推送 `iotwq/china-api:latest`，包含 v0.2.3 上游更新、本站定制及合并兼容修正。
- 版本 `0.2.3`，构建标识 `563aba27b28d-dirty`，构建时间 `2026-09-08T15:12:23Z`（北京时间 23:12:23）；发布 `linux/amd64` 和 `linux/arm64`。
- 远端 index 摘要为 `sha256:8c378a968668f81fabbc97d6fb8dad50048065f603e4dd967b87d17d8c993360`。本轮未修改业务源码、生产数据库或运行中的服务。

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.3 --build-arg COMMIT=563aba27b28d-dirty --build-arg DATE=2026-09-08T15:12:23Z --tag iotwq/china-api:latest --metadata-file /tmp/sub2api-image-v023-20260908.3RA38f/metadata.json --progress=plain --push .`：退出 0，镜像内前端翻译/类型检查和生产构建、双架构 Go 编译与推送均通过；日志 `/tmp/sub2api-image-v023-20260908.3RA38f/build.log`。
- `docker buildx imagetools inspect iotwq/china-api:latest`：远端摘要与构建元数据一致，amd64 manifest 为 `sha256:7c6bfee8750d62ddb23f00a932812169e8f6da00d14f7d5bfa4e5ec791477ca8`，arm64 manifest 为 `sha256:0ced909acff7be98dbfca880cb0b6f97f85975f05a498b90611cad0096038937`。
- arm64 按发布 index 摘要拉取，使用 `docker run --rm --platform linux/arm64 --network none --read-only --entrypoint /app/sub2api iotwq/china-api@sha256:8c378a968668f81fabbc97d6fb8dad50048065f603e4dd967b87d17d8c993360 --version` 成功验证版本、构建标识与时间；日志 `/tmp/sub2api-image-v023-20260908.3RA38f/version-arm64.log`。
- amd64 两次远端拉取因 Docker Hub CDN EOF 中断；随后以相同构建参数、`--platform linux/amd64 --tag iotwq/china-api:verify-v023-amd64 --load` 从构建缓存加载到本地。`docker image inspect` 确认本地镜像 ID 与已发布 amd64 manifest 摘要完全一致；`docker run --rm --pull never --platform linux/amd64 --network none --read-only --entrypoint /app/sub2api iotwq/china-api:verify-v023-amd64 --version` 成功退出 0，版本、构建标识及时间一致。验证日志 `amd64-local-load.log`、`version-amd64-local.log` 位于同一临时目录。
- 源码功能测试及隔离数据库迁移测试沿用紧邻上一轮同步验证，本轮未修改业务源码，未重复全量测试。未执行生产服务健康验收或数据库迁移。

### Notes

- `progress.md`：追加发布摘要、构建及双架构版本验证证据与回滚点。
- `iotwq/china-api:verify-v023-amd64` 仅为本地验证标签，未推送。构建保留既有前端大 chunk 提示。
- 部署新版本时会应用上游 236/237 两个迁移，含白名单结构修复及 MiniMax 平台约束扩展；详情见 `docs/UPSTREAM_SYNC.md`。本轮未部署。
- 发布前镜像为 `iotwq/china-api@sha256:0c68a6472f7082d321af781d4cbce88aa8a3d2e361123f967c2d480568c4ca2e`。恢复 latest 的可执行命令为 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:0c68a6472f7082d321af781d4cbce88aa8a3d2e361123f967c2d480568c4ca2e`；该操作仅恢复镜像标签，不回滚数据库，本轮未执行。

## 2026-09-09 - Task: 同步上游 v0.2.4 并保留本站定制

### What was done

- 将 `upstream/main` 的 `270eac697..98d86915b` 共 3 个提交、19 个变更文件合并到 main，合并提交 `cb76eeba0dd6c81e5cac671be4e05f3633db298d`，版本 `0.2.4`。
- 同步 Image 2.5 模型、定价、OAuth 生图主控模型配置、测试列表和错误反馈修复，保留本站图片转换、视频恢复/退款、用量修复、OAuth 重试及前端定制。
- 先备份完整工作区，恢复时解决唯一的测试初始化冲突，保留上游环境变量覆盖测试并沿用本地统一 Gin 初始化。92 个原未跟踪文件内容哈希全部一致。
- 本批没有数据库迁移、依赖或依赖注入改动，未重新生成 Wire；未发布镜像、推送 GitHub 或部署服务。

### Testing

- `cd backend && go test -tags=unit ./... -count=1`：全量通过，包含新生图主控模型/图片模型分离、定价、权限、用量，以及本站账务和恢复等单元回归；日志 `/tmp/sub2api-sync-20260909.46CcX6/backend-unit.log`。
- `cd frontend && pnpm exec vitest run`：283 个文件、2076 项测试全部通过；日志 `/tmp/sub2api-sync-20260909.46CcX6/frontend-tests.log`。
- `cd frontend && pnpm lint:check`、`pnpm build`：通过，构建包含翻译和类型检查；日志 `/tmp/sub2api-sync-20260909.46CcX6/frontend-lint.log`、`frontend-build.log`。保留既有大 chunk 提示。
- `sh deploy/tests/docker-compose-gateway-env-test.sh`：通过。四份 Compose 均使用 `docker compose --env-file /dev/null -f <文件> config --format json` 验证 `services.sub2api.environment.SUB2API_IMAGES_MAIN_MODEL`，空值回退 `gpt-5.6-luna`、显式值保留 `gpt-5.6-sol`，共 8 项通过；仅用验证用数据库/Redis占位配置渲染，未创建服务。
- 无未解决索引和严格冲突标记，`git diff --check`、`git diff --cached --check`、两处测试的 `gofmt -d` 和上游祖先关系检查通过。原 92 个未跟踪文件 SHA-256 全部一致，清单 `/tmp/sub2api-sync-20260909.46CcX6/untracked-before.sha256`。未执行真实上游端到端或生产验证。

### Notes

- `README_CN.md`：同步上游 Image 2.5 主控模型配置说明。
- `backend/cmd/server/VERSION`：同步版本 `0.2.4`。
- `backend/internal/pkg/openai/constants.go`：同步新图片模型及默认清单。
- `backend/internal/pkg/openai/constants_test.go`：同步模型清单回归验证。
- `backend/internal/service/account_test_models_test.go`：同步 OAuth 图片模型测试列表验证。
- `backend/internal/service/account_test_service.go`：同步图片模型列表补齐和上游错误反馈。
- `backend/internal/service/account_test_service_openai_image_test.go`：同步主控模型/图片模型及错误反馈测试，解决本地 Gin 初始化冲突。
- `backend/internal/service/openai_codex_transform.go`：同步 Responses 图片模型规范化和主控模型选择。
- `backend/internal/service/openai_images.go`：同步图片输入用量和模型错误处理，保留本站转换逻辑。
- `backend/internal/service/openai_images_model_test.go`：同步新增主控模型、计价、权限、用量回归，沿用本地统一 Gin 初始化。
- `backend/internal/service/openai_images_responses.go`：同步可配置主控模型，独立保留请求图片模型。
- `backend/internal/service/pricing_service.go`：同步 Image 2.5 默认计价。
- `backend/resources/model-pricing/model_prices_and_context_window.json`：同步内置模型价格数据。
- `deploy/.env.example`：同步生图主控模型环境变量说明。
- `deploy/docker-compose.dev.yml`：同步开发部署的主控模型环境变量。
- `deploy/docker-compose.local.yml`：同步本地部署的主控模型环境变量。
- `deploy/docker-compose.standalone.yml`：同步独立部署的主控模型环境变量。
- `deploy/docker-compose.yml`：同步标准部署的主控模型环境变量。
- `frontend/src/composables/useModelWhitelist.ts`：同步新图片模型的可选清单。
- `docs/UPSTREAM_SYNC.md`：追加同步范围、兼容处理、配置使用及回滚说明。
- `progress.md`：追加本轮改动与验证，不改写历史记录。
- 完整上游清单可执行 `git diff --name-only 563aba27b28dfcaaec52c42244832556b781d8b3 cb76eeba0dd6c81e5cac671be4e05f3633db298d` 查看。原暂存/未暂存划分通过 `git apply --cached --reverse /tmp/sub2api-sync-20260909.46CcX6/worktree-before.patch` 恢复，本轮兼容修正和记录未另行提交。
- 回滚点为分支 `codex/pre-upstream-sync-20260909` 和完整 stash `038cfb57e67e90550f706c5ba549bbfa32dd5d05`，均保留。可执行 `git worktree add -b codex/recover-pre-sync-20260909 /tmp/sub2api-pre-sync-20260909 codex/pre-upstream-sync-20260909`，再执行 `git -C /tmp/sub2api-pre-sync-20260909 stash apply --index 038cfb57e67e90550f706c5ba549bbfa32dd5d05`，独立恢复同步前完整代码，不覆盖当前工作区。

## 2026-09-09 - Task: 修复测试上游账号自定义模型列表

### What was done

- 修复 OpenAI 账号测试列表与模型白名单/映射不一致的问题：实时目录返回后，补入账号已配置的具体映射入口，包含自定义模型名称；通配符入口不加入下拉列表。
- 保留原有测试提交和账号映射解析，因此选择自定义名称后会按配置调用真实目标模型。OAuth、Setup Token、API Key 共用该行为。

### Testing

- `cd backend && go test -tags=unit ./internal/service -run 'TestFetchOpenAIAccountModels|TestAccountGetMappedModel' -count=1`：通过，日志由本轮命令直接输出；覆盖实时目录、自定义映射入口、重复项、通配符排除及目标模型解析。
- `gofmt -w backend/internal/service/account_test_service.go backend/internal/service/account_test_models_test.go`：通过。
- `git diff --check`：通过。未执行真实上游端到端请求。

### Notes

- `backend/internal/service/account_test_service.go`：在 OpenAI 测试模型目录中补入已配置的具体映射键，并过滤通配符。
- `backend/internal/service/account_test_models_test.go`：新增自定义映射键列表和目标解析回归测试。
- `docs/UPSTREAM_SYNC.md`：追加本次兼容修复、验证及回滚说明。
- `progress.md`：追加本轮记录。
- 回滚方式：恢复上述两个源码文件到本轮修改前版本；若需回到本次上游同步前的完整状态，可在独立 worktree 执行 `git worktree add -b codex/recover-pre-sync-20260909 /tmp/sub2api-pre-sync-20260909 codex/pre-upstream-sync-20260909`，再应用 stash `038cfb57e67e90550f706c5ba549bbfa32dd5d05`。

## 2026-09-09 - Task: 构建并推送自定义模型测试修复镜像

### What was done

- 基于当前 v0.2.4 工作区构建并推送 `iotwq/china-api:latest` 双架构镜像，包含自定义模型测试列表修复。
- 远端 OCI index 为 `sha256:5806b6cb091ff9fa9489ca2226d2647299ef5a4fa5658cdc569e664e7d73a52a`，amd64/arm64 manifest 已确认。

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.4 --build-arg COMMIT=cb76eeba0dd6-dirty --build-arg DATE=2026-09-09T13:52:20Z --tag iotwq/china-api:latest --push .`：退出 0，前端生产构建、双架构 Go 编译和 Docker Hub 推送通过；日志 `/tmp/sub2api-image-v024-79lMI9/build.log`。
- `docker buildx imagetools inspect iotwq/china-api:latest`：通过，远端 index 与两个 manifest 摘要为本轮构建结果。
- `docker buildx build --builder codex-multiarch --platform linux/amd64 --load ...`：使用同一构建参数从缓存加载本地验证镜像；`docker run --rm --pull never --network none --read-only ... --version` 通过，版本 `0.2.4`、commit 和构建时间一致；日志 `/tmp/sub2api-image-v024-79lMI9/version-amd64.log`。
- 直接从 Docker Hub 按 index 摘要拉取启动时遇到 CDN EOF，已停止该次下载；不影响已完成的推送和 manifest 校验。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加镜像摘要、架构 manifest、部署和回滚说明。
- `progress.md`：追加本轮构建、推送、验证和回滚记录。
- 本轮未修改业务源码、数据库、余额或运行中的服务；发布前 latest 摘要为 `sha256:8c378a968668f81fabbc97d6fb8dad50048065f603e4dd967b87d17d8c993360`，可用文档中的 `docker buildx imagetools create` 命令恢复。

## 2026-09-11 - Task: 修复渠道监控 Anthropic 策略拒绝不换号

### What was done

- 接入渠道监控内部探测头下的 Anthropic 400 策略拒绝换号逻辑，识别 `probe_request_rejected` 和 `request blocked by gateway policy` 后返回 `UpstreamFailoverError`，使监控继续选择同组下一个上游账号。
- 保留普通 400、普通用户请求和其他监控参数错误的原有处理，避免无意义重试和错误冷却。
- 增加策略拒绝、包装响应、普通参数错误及内部头校验回归测试。

### Testing

- `gofmt -w backend/internal/service/gateway_forward.go backend/internal/service/openai_gateway_upstream_errors.go backend/internal/service/channel_monitor_failover_test.go`：通过。
- `cd backend && go test -tags=unit ./internal/service -run 'TestShouldFailoverChannelMonitorProbeError|TestGatewayService' -count=1`：通过。
- `git diff --check`：通过。未执行真实上游渠道监控请求。

### Notes

- `backend/internal/service/gateway_forward.go`：将监控策略拒绝纳入 400 账号切换分支，并兼容空配置日志路径。
- `backend/internal/service/openai_gateway_upstream_errors.go`：新增严格的监控策略拒绝识别与组合判定。
- `backend/internal/service/channel_monitor_failover_test.go`：新增监控策略拒绝换号回归测试。
- `docs/UPSTREAM_SYNC.md`：记录行为变更、验证和回滚方式。
- `progress.md`：追加本轮施工记录。
- 回滚方式：恢复上述三个源码文件至本轮修改前版本，并重新部署上一版固定镜像 digest；本轮未修改数据库或运行中服务。

## 2026-09-11 - Task: 构建并推送渠道监控修复镜像

### What was done

- 基于当前工作区构建并推送 `iotwq/china-api:latest` 双架构镜像，包含 Anthropic 渠道监控策略拒绝换号修复。
- Docker Hub OCI index 为 `sha256:83089738ee533188e404a1dd5781ce2d81735f05537c77d7e4db56819ae8dc4f`，amd64/arm64 manifest 已确认。

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.4 --build-arg COMMIT=cb76eeba0dd6-dirty --tag iotwq/china-api:latest --push .`：重试后退出成功，前端生产构建、双架构 Go 编译和 Docker Hub 推送通过；日志 `/tmp/sub2api-image-TLfy1v/build-retry.log`。
- `docker buildx imagetools inspect iotwq/china-api:latest`：通过，远端 index 与双架构摘要与发布结果一致。
- amd64 本地加载镜像执行 `docker run --rm --pull never --network none --read-only iotwq/china-api:verify --version`：通过，显示版本 `0.2.4`、commit `cb76eeba0dd6-dirty`。
- `git diff --check`：通过。首次构建因 Docker Hub registry EOF 中断，未产生远端更新；重试成功完成。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加本次镜像摘要、验证、部署和回滚说明。
- `progress.md`：追加本轮构建、推送、验证和回滚记录。
- 本轮未修改数据库、余额或运行中的服务；回滚可重新部署上一版固定 digest `sha256:5806b6cb091ff9fa9489ca2226d2647299ef5a4fa5658cdc569e664e7d73a52a`。

## 2026-09-12 - Task: 支持 NAI Diffusion 图片模型

### What was done

- 将 `nai-diffusion-4-5-full` 和 `nai-diffusion-5-full` 纳入 OpenAI 兼容图片模型识别，允许通过 `/v1/images/generations` 使用。
- 保留现有模型映射、渠道定价、用量统计和图片响应处理逻辑；未修改数据库或上游协议。
- 增加两个模型的图片请求解析回归测试。

### Testing

- `gofmt -w backend/internal/service/openai_images.go backend/internal/service/openai_images_test.go`：通过。
- `cd backend && go test -tags=unit ./internal/service -run 'TestOpenAIGatewayServiceParseOpenAIImagesRequest_(AllowsNaiImageModels|RejectsNonImageModel|AllowsGrokImageModels)$|TestImageGenerationIntent' -count=1`：通过。
- `git diff --check`：通过。未执行真实 NAI 上游生图请求。

### Notes

- `backend/internal/service/openai_images.go`：新增 NAI Diffusion 图片模型识别。
- `backend/internal/service/openai_images_test.go`：新增两个 NAI 模型的图片请求校验测试。
- `docs/OPENAI_MEDIA_COMPAT.md`：补充 NAI 图片模型兼容说明。
- `docs/UPSTREAM_SYNC.md`：记录本轮兼容改动、验证和回滚方式。
- `progress.md`：追加本轮施工记录。
- 回滚方式：恢复上述两个源码文件至本轮修改前版本，并重新部署上一版固定镜像 digest。

## 2026-09-13 - Task: V1 主动渠道监控增加糖果题智力检测

### What was done

- 在每个主动监控项增加默认关闭的智力检测开关，按该项配置的密钥、端点和主/附加模型运行；同名展示分组不联动，V2 不新增主动请求。
- 查证社区常用糖果题与答案 21，使用最坏情况枚举独立验证；仅对最终答案判分，不因中间推导提到 21 就判通过。
- 用户卡片增加最近 60 次智力历史条：合格绿色、不合格红色、未完成或未检测灰色，悬停显示原因与时间；管理端和详情中也逐模型显示。
- 保留渠道连通状态与可用率独立统计。启用时替换普通加减题、不因答错追加重试；兼容各协议的输出预算、正常 error:null 和多片段最终文本，超时结果仍能落历史。
- 增加数据库迁移及使用文档，保留原有脏工作区，并生成仅含本轮变更的可逆补丁。

### Testing

- `cd backend && go generate ./ent`：通过；生成差异仅涉及监控字段和对应 Ent 元数据，go.mod/go.sum 无变化。
- 监控相关 Go 单元/映射测试（`-tags=unit`，筛选 `Test.*(ChannelMonitor|Monitor|Challenge|RunCheck)`）：service、handler、admin 三包分批确认通过。新增测试覆盖正确/错误/截断/无最终答案、error:null、多协议请求、默认关闭、无答错重试、创建/编辑/复制、V2 拦截及取消上下文后历史保存。
- `go test -tags=integration ./internal/repository -run 'TestChannelMonitor(IntelligenceRoundTrip|QuotaModeRoundTrip|HistoryQuotaRoundTrip)' -count=1 -v`：3 项通过；实际启动隔离 PostgreSQL/Redis，执行迁移并验证各查询回读。
- 前端 `MonitorIntelligence.spec.ts` 与 `MonitorFormDialog.accountSelector.spec.ts`：15 项通过；原有 `MonitorPrimaryModelCell.spec.ts`、`MonitorCard.quota.spec.ts` 及语言键完整性检查通过。
- `pnpm run typecheck`、本轮前端文件定向 ESLint：通过。
- `pnpm exec vite build --outDir /tmp/sub2api-monitor-intelligence.qUtyPO/frontend-dist`：通过；`go build -o /tmp/sub2api-monitor-intelligence.qUtyPO/server ./cmd/server`：通过。
- 使用本地真实 Vue 组件和中文语言包做浏览器视觉检查，确认红/绿/灰徽标、双排时间线、附加模型状态和时间提示可见。验证用的两个临时预览入口已删除，开发服务器已关闭。
- `git diff --check` 和 `git apply --reverse --check /tmp/sub2api-monitor-intelligence.qUtyPO/intelligence.patch`：通过。
- 未请求真实计费上游；未发布镜像、重启线上服务或修改线上数据库。

### Notes

- `backend/ent/channelmonitor/channelmonitor.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/channelmonitor/where.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/channelmonitor.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/channelmonitor_create.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/channelmonitor_update.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/channelmonitorhistory/channelmonitorhistory.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/channelmonitorhistory/where.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/channelmonitorhistory.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/channelmonitorhistory_create.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/channelmonitorhistory_update.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/migrate/schema.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/mutation.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/runtime/runtime.go`：根据监控 schema 重新生成对应的 ORM 字段、读写或迁移元数据。
- `backend/ent/schema/channel_monitor.go`：为监控配置新增默认关闭的开关定义。
- `backend/ent/schema/channel_monitor_history.go`：为历史记录新增可空智力结果定义。
- `backend/internal/domain/channel_monitor_intelligence.go`：定义独立的智力检测结果和原因数据。
- `backend/internal/handler/admin/channel_monitor_handler.go`：接收开关参数并返回智力结果、历史与附加模型状态。
- `backend/internal/handler/admin/channel_monitor_intelligence_test.go`：验证管理端开关三态和结果 DTO。
- `backend/internal/handler/channel_monitor_intelligence_test.go`：验证用户列表、时间线和详情结果透出。
- `backend/internal/handler/channel_monitor_user_handler.go`：向用户返回独立检测开关、结果及时间线。
- `backend/internal/handler/dto/channel_monitor.go`：为共享附加模型 DTO 增加智力结果。
- `backend/internal/repository/channel_monitor_intelligence_integration_test.go`：验证配置隔离、历史读写、批量查询及可用率独立性。
- `backend/internal/repository/channel_monitor_repo.go`：保存开关和结果，补齐最新记录及历史查询。
- `backend/internal/service/channel_monitor_aggregator.go`：将检测结果聚合到管理端、用户列表和详情。
- `backend/internal/service/channel_monitor_checker.go`：接入糖果题探测、多片段文本解析和完整请求耗时。
- `backend/internal/service/channel_monitor_intelligence.go`：实现题面、最终答案判定、截断识别及输出预算。
- `backend/internal/service/channel_monitor_intelligence_test.go`：验证多协议、判分、默认行为、配置生命周期、超时保存和答案枚举。
- `backend/internal/service/channel_monitor_service.go`：贯通创建、编辑、复制及定时检测，并保留超时样本。
- `backend/internal/service/channel_monitor_types.go`：扩展服务内部监控配置、检测结果及历史视图类型。
- `backend/migrations/238_channel_monitor_intelligence.sql`：新增默认关闭的开关列及可空历史结果列。
- `frontend/src/api/admin/channelMonitor.ts`：增加管理端开关与结果类型。
- `frontend/src/api/channelMonitor.ts`：增加用户端列表、历史及详情结果类型。
- `frontend/src/components/admin/monitor/MonitorFormDialog.vue`：增加默认关闭的开关与不兼容模式提示和验证。
- `frontend/src/components/admin/monitor/MonitorPrimaryModelCell.vue`：显示主模型和附加模型最近检测结果。
- `frontend/src/components/admin/monitor/MonitorRunResultDialog.vue`：逐模型展示手动检测结果。
- `frontend/src/components/common/MonitorIntelligenceBadge.vue`：提供红、绿、灰状态徽标及提示。
- `frontend/src/components/user/MonitorDetailDialog.vue`：增加各模型的智力结果列。
- `frontend/src/components/user/__tests__/MonitorIntelligence.spec.ts`：验证历史顺序、颜色、悬停提示、可见性及手动结果展示。
- `frontend/src/components/user/monitor/MonitorCard.vue`：保留渠道状态并添加主模型历史条和附加模型结果。
- `frontend/src/components/user/monitor/MonitorIntelligenceTimeline.vue`：渲染最近 60 次独立智力状态条。
- `frontend/src/composables/useMonitorIntelligence.ts`：统一状态文本、颜色和原因提示。
- `frontend/src/i18n/locales/en/dashboard.ts`：补充英文开关说明和检测状态。
- `frontend/src/i18n/locales/zh/dashboard.ts`：补充中文开关说明和检测状态。
- `frontend/src/views/admin/__tests__/MonitorFormDialog.accountSelector.spec.ts`：增加真实开关保存、切换及不兼容模式交互测试。
- `docs/channel-monitor-intelligence.md`：记录使用方式、题目来源、判分范围、验证和回滚。
- `progress.md`：追加本轮施工与验证记录。
- 回滚：先在监控编辑中关闭智力检测即可恢复普通探活；源码在仓库根目录执行 `git apply --reverse --check /tmp/sub2api-monitor-intelligence.qUtyPO/intelligence.patch`，检查通过后执行 `git apply --reverse /tmp/sub2api-monitor-intelligence.qUtyPO/intelligence.patch`。该补丁仅涵盖本轮 45 个源码/文档文件，不回退历史日志或更早定制；修改前快照位于 `/tmp/sub2api-monitor-intelligence.qUtyPO/before.tar`。数据库新增列可保留，不需要删除数据。

## 2026-09-13 - Task: 接入 Wan 3.0x 与 ViralDance 四类视频上游

### What was done

- 对照用户提供的 `wan3.0x（新）.md`、`viraldance933.md`、`viraldance2.5-30.md`、`viraldance2.5-15.md` 检查当前 OpenAI APIKey 视频链路，确认原有模型识别会拒绝文档中的 8 个精确模型名。
- 将 8 个模型接入现有异步 `/v1/videos` 创建、任务绑定和状态查询链路；保留参考素材字段、布尔值、状态、错误和不同位置的视频 URL，不套用 MiniMax V2 协议或 Firefly multipart。
- 补齐各模型的分辨率和默认时长，修复 Wan / ViralDance 30 秒请求在预估、结算和用量记录中被旧 15 秒上限截断的问题。沿用管理员配置的按次/按秒价格及现有幂等退款，不自动引用供应商示例金额。
- 沿用现有整数秒账务结构，对这些模型的必填时长、范围、类型及小数秒在上游调用前校验，避免静默少计费；这是本站整秒接入限制，已在文档明确说明。
- 更新接入配置、模型差异、价格优先级、创建和查询示例。四份文档均未提供 `/content` 接口，本次按其协议返回可播放/下载的 URL，不新增视频文件代理。

### Testing

- 修复前执行 `cd backend && go test ./internal/service -run '^TestViraleeVideo' -count=1`：稳定复现 8 个模型在解析阶段被拒，以及 30 秒仅按 15 秒计费、2.5-30 默认时长错误。
- 新增定向 service 测试通过：覆盖 8 个模型、根地址和 `/v1` 地址、`id` / `task_id`、参考图片/视频/音频/首尾帧/元素数组、`async` 与 `generate_audio: false`、完成 URL 的全部文档位置、排队与失败状态、默认秒数、按次价格、完整 30 秒的渠道与分组秒价、非法时长拒绝，以及用量/扣费/退款一致和重复退款一次生效。
- `cd backend && go test ./internal/handler -run '^TestViraleeVideos' -count=1`：通过；8 个模型均经过实际 handler 完成提交、预占/扣费、任务持久化接口、绑定原账号查询、成功 URL 返回和不重复记账。测试仓储模拟真实数据库的 `inactive` 默认值。
- `cd backend && go test -tags=unit ./internal/service ./internal/handler ./internal/server/routes -run 'Test.*(Video|Videos|SD20|MiniMax|Firefly)' -count=1`：三个包均通过；同时回归既有 Firefly、MiniMax 恢复、视频计费、退款和路由测试。
- `cd backend && go build -o /tmp/sub2api-viralee-video.ksKJXr/server ./cmd/server`：通过。
- 本轮 Go 文件 `gofmt -l` 无输出；`git diff --check`、`git apply --reverse --check /tmp/sub2api-viralee-video.ksKJXr/viralee-video.patch` 通过。
- 本次没有可用的该上游测试密钥，验证使用模拟 HTTP 响应及账务仓储，未执行真实付费生成；未发布镜像、修改线上账号/价格、重启服务或变更数据库结构。

### Notes

- `backend/internal/service/openai_viralee_video.go`：定义 8 个精确模型的分辨率、整秒范围和默认值，并在提交前验证时长。
- `backend/internal/service/openai_videos.go`：接入新模型识别、请求校验和对应计费参数，保持现有异步 JSON 转发。
- `backend/internal/service/video_billing_resolution.go`：新增按模型处理时长的逻辑，保留旧模型的 15 秒规则。
- `backend/internal/service/openai_gateway_usage.go`：在预估、用量记录和渠道秒价计算中使用对应模型的完整时长。
- `backend/internal/service/billing_service.go`：让分组视频秒价计算也支持新模型的 30 秒时长。
- `backend/internal/service/openai_viralee_video_test.go`：验证文档请求/响应透传、默认值、计费、时长校验及幂等退款。
- `backend/internal/handler/openai_viralee_videos_test.go`：验证完整创建、账号绑定、结算和查询返回链路。
- `docs/OPENAI_MEDIA_COMPAT.md`：补充四类视频服务的配置、参数、计费方式、整秒限制和 URL 读取方法。
- `progress.md`：仅追加本轮实现与验证记录。
- 回滚：在仓库根目录执行 `git apply --reverse --check /tmp/sub2api-viralee-video.ksKJXr/viralee-video.patch`，检查通过后执行 `git apply --reverse /tmp/sub2api-viralee-video.ksKJXr/viralee-video.patch`。补丁仅包含本轮 8 个代码/文档文件，不回退此前定制、渠道智力检测或历史日志。修改前快照为 `/tmp/sub2api-viralee-video.ksKJXr/before.tar`，无需回退数据库。

## 2026-09-13 - Task: 修复同步图片结算差额及失败后释放预扣款

### What was done

- 按最终图片尺寸与数量统一计算实际费用，预冻结差额与用量、扣款流水在同一事务结算。
- 已完成图片的结算失败保留冻结款与持久化待结算快照，后台按原金额重试；重复任务不重复扣款。
- 媒体冻结的扣款/释放互斥；明确已释放的原冻结从余额结算，不占用其他请求的冻结款；无法解释的冻结状态返回待核对错误。未更改批量图片任务的超额策略。

### Testing

- 修复前审计覆盖请求 1K / 输出 2K 的明细 $2、扣款 $1 不一致；本轮 `TestImageSettlementUsesResolvedOutputCost` 验证两者均为 $2。
- `cd backend && go test -tags=unit ./internal/service ./internal/handler ./internal/repository -run 'Test(ImageSettlementUses|NanoBanana|UsageBillingRepositoryApplyWithUsageLog|CaptureUsageBillingBatch|UsageLogStaticInsertShape|PrepareUsageLogInsert_ArgCount)' -count=1`：通过，包括 SQL 列/参数数目及原子回滚回归。
- `cd backend && go test -tags=integration ./internal/repository -run '^TestImageSettlement_OverEstimateReleasedHoldAndRetry$' -count=1 -v`：在一次性 PostgreSQL 18.1 / Redis 8.4 容器通过；覆盖超预估、原冻结已释放、失败后重试、6 个并发重放、其他冻结保留、余额/配额/用量/流水一致及快照无凭据。
- 未连接生产数据库，未补扣历史费用，未发布镜像。

### Notes

- `backend/internal/service/usage_billing.go`：增加已完成媒体的结算策略及可识别的冻结状态错误。
- `backend/internal/service/openai_media_balance_hold.go`：媒体冻结使用完成后结算策略。
- `backend/internal/service/openai_gateway_usage.go`：捕获金额直接采用最终计费结果。
- `backend/internal/service/gateway_usage_billing.go`：图片结算快照、失败重试及余额缓存同步。
- `backend/internal/service/pending_image_settlement.go`：按持久化原金额执行后台结算与缓存同步。
- `backend/internal/service/openai_video_compensator.go`：在现有媒体补偿循环处理待结算图片。
- `backend/internal/repository/usage_billing_repo.go`：支持差额结算、扣款释放互斥与原冻结已释放的核对。
- `backend/internal/repository/pending_image_settlement_repo.go`：待结算快照持久化、租约领取与重试状态。
- `backend/migrations/239_pending_image_settlements.sql`：增加待结算图片表，无历史账务改写。
- `backend/internal/handler/openai_images.go`、`backend/internal/handler/openai_nano_banana.go`：去除重复金额计算，成功生成后不自动释放结算失败的预扣。
- `backend/internal/service/image_settlement_test.go`：最终尺寸计费一致性回归。
- `backend/internal/handler/openai_nano_banana_billing_test.go`：成功生成后记账失败保留预扣回归。
- `backend/internal/repository/image_settlement_integration_test.go`：真实数据库账务一致性、回滚及并发重试验证。
- `docs/BILLING_INTEGRITY.md`：记录结算行为、迁移和运维边界。
- `progress.md`：追加本轮记录。
- 回滚点：修改前快照 `/tmp/sub2api-billing-fix.kI6FJx/before.tar`；本次整体修复结束后提供仅包含本次改动的反向补丁。可将原文件提取至临时目录逐文件恢复；不要直接覆盖整个工作区。上线后回滚代码时保留待结算表与记录，先停止相应补偿进程，不删除业务数据。

## 2026-09-13 - Task: 修复 Gemini 缺价归零及原生生成缺少价格预检

### What was done

- 原生 Gemini 生成在选定账号后按请求、渠道映射和实际转发模型检查价格，缺价时在转发前返回可操作错误。
- 通用计费返回价格计算错误，禁止有用量但因缺价而成功写入零费用；显式免费倍率与已定价模型映射保持可用。
- 仅对生成操作增加预检，不对模型列表和 countTokens 增加此价格门槛。

### Testing

- `cd backend && go test -tags=unit ./internal/service ./internal/handler -run 'TestGeminiBillingIntegrity|TestGeminiNativeMissing|TestGatewayService.*(RecordUsage|Pricing|Fallback)|Test.*Gemini.*' -count=1`：通过。
- 新回归确认缺价的 1500 token 请求返回错误、0 条用量/扣款调用；实际模型有价的别名请求及显式免费倍率均成功。
- 两个原生生成 action 经真实 handler 选号后返回 400，未进入上游转发；回归现有 Gemini 兼容、计费与模型映射用例。

### Notes

- `backend/internal/service/gateway_usage_billing.go`：费用计算错误逐层传播，不再默认归零。
- `backend/internal/handler/gemini_v1beta_handler.go`：原生生成按实际选定账号进行价格预检。
- `backend/internal/service/openai_gateway_record_usage_test.go`：已有图片计费验证适配显式错误返回。
- `backend/internal/service/gemini_billing_integrity_test.go`：新增缺价、别名回退和免费倍率验证。
- `backend/internal/handler/gemini_pricing_preflight_test.go`：新增同步/流式 Gemini 缺价请求转发前拦截验证。
- `docs/BILLING_INTEGRITY.md`：说明 Gemini 预检范围和免费策略。
- `progress.md`：追加本轮记录。
- 回滚：原文件均保存在 `/tmp/sub2api-billing-fix.kI6FJx/before.tar`，可先用 `tar -xf /tmp/sub2api-billing-fix.kI6FJx/before.tar -C <空的临时目录>` 提取比较，整体修复结束提供本轮专用反向补丁。没有线上价格、用量或余额修改；历史两笔缺价记录仍须人工核对。

## 2026-09-13 - Task: 修复视频冻结异常循环并完成计费完整性验证

### What was done

- 视频恢复遇到无法核实的冻结金额或幂等冲突时转入 billing_review，保存原任务归属和错误，停止无效扣款/退款轮询。
- 已匹配唯一上游任务后正常查询状态和下载视频，账务待重试/待核对不再掩盖已生成结果；实际 ID 路由与本地恢复路由共用原计费 ID 控制退款。
- 冻结明确已释放时可在原事务中从可用余额结算一次；没有明确释放证据的冻结缺失不自动补扣或退款。
- 删除独立的媒体实扣金额输入，图片和视频均以最终费用/保存的价格快照为唯一来源；增加同一冻结被不同用量 ID 重用的冲突检查。
- 补充图片后台实际重放的回归，确认价格变更后仍按原快照结算。
- 复核此前的 SQL 列/参数数量、明细和扣款事务原子性修复，不重复改写历史记录。

### Testing

- `cd backend && go test -tags=unit ./internal/service ./internal/handler ./internal/repository ./migrations -run 'Test.*(Billing|RecordUsage|Pricing|ImageSettlement|NanoBanana|Video|Videos|MiniMax|Firefly|GeminiNativeMissing|UsageLogStaticInsert|PrepareUsageLogInsert_ArgCount)' -count=1`：四个包全部通过。
- `cd backend && go test -tags=integration ./internal/repository -run 'Test(ImageSettlement|MediaBillingReview|VideoBillingReview|UsageBillingRepositoryApplyWithUsageLog)' -count=1 -v`：在一次性 PostgreSQL 18.1 / Redis 8.4 容器全部通过；验证超额结算、已释放/未解释冻结、并发幂等、重复冻结使用被拒绝、凭据不入快照、用量/余额/配额/流水一致、待核对排除轮询、唯一归属及核对后恢复退款领取。
- `cd backend && go build -o /tmp/sub2api-billing-fix.kI6FJx/server ./cmd/server`：通过。
- 本轮 Go 文件 `gofmt -l` 无输出；`git diff --check` 通过；`git apply --reverse --check /tmp/sub2api-billing-fix.kI6FJx/billing-fix.patch` 通过。
- 没有连接生产数据库或修改线上余额。历史 813 条漏记、9,205 条旧零费用和具体四个视频结果仍依赖线上日志/任务证据核对；本轮不自动补账。数据库在快照持久化前整体不可用仍需日志核账。

### Notes

以下为整个本次计费修复的改动清单（包含前两项任务及收尾适配）：

- `backend/internal/service/gateway_usage_billing.go`：图片持久化结算接入、缓存处理和计价错误传播。
- `backend/internal/service/gemini_billing_integrity_test.go`：缺价、模型映射和显式免费倍率验证。
- `backend/internal/service/image_settlement_test.go`：图片输出尺寸与实扣金额一致性验证。
- `backend/internal/service/media_billing_recovery_test.go`：图片后台重放原价格及视频冻结异常停止重试验证。
- `backend/internal/service/openai_gateway_record_usage_test.go`：适配计价错误返回和唯一金额输入。
- `backend/internal/service/openai_gateway_usage.go`：删除独立实扣金额，以最终计价结果捕获冻结。
- `backend/internal/service/openai_media_balance_hold.go`：为媒体冻结启用完成后差额结算策略。
- `backend/internal/service/openai_minimax_video_recovery.go`：增加账务待核对状态常量。
- `backend/internal/service/openai_video_compensator.go`：处理图片补偿与视频待核对，缺少视频历史价格不重新估价。
- `backend/internal/service/openai_viralee_video_test.go`：适配删除冗余实扣输入，保留原价格快照验证。
- `backend/internal/service/pending_image_settlement.go`：持久化图片结算接口和原金额重试工作流。
- `backend/internal/service/usage_billing.go`：完成媒体策略与冻结不一致错误。
- `backend/internal/handler/gemini_pricing_preflight_test.go`：原生生成接口缺价转发前拦截回归。
- `backend/internal/handler/gemini_v1beta_handler.go`：标准模式生成请求价格预检。
- `backend/internal/handler/openai_images.go`：去除重复估价，生成后移交冻结给结算。
- `backend/internal/handler/openai_nano_banana.go`：采用相同冻结移交与统一实扣策略。
- `backend/internal/handler/openai_nano_banana_billing_test.go`：记账错误不释放成功生成预扣回归。
- `backend/internal/handler/openai_videos.go`：已知任务状态与账务解耦，未结算不提前退款。
- `backend/internal/handler/openai_videos_recovery_billing_test.go`：已识别、待核对、未识别任务的查询/下载验证。
- `backend/internal/repository/image_settlement_integration_test.go`：真实 PostgreSQL 原子结算、超额与幂等验证。
- `backend/internal/repository/media_billing_review_integration_test.go`：未解释冻结、唯一任务归属与补偿暂停/恢复验证。
- `backend/internal/repository/openai_video_task_binding_repo.go`：原计费 ID 尚未结算时暂停普通退款补偿领取。
- `backend/internal/repository/pending_image_settlement_repo.go`：快照持久化、租约领取、重试/待核对状态与完成清理。
- `backend/internal/repository/usage_billing_repo.go`：扣款/释放互斥、按实际费用补差及防止重用冻结。
- `backend/migrations/239_pending_image_settlements.sql`：新增待结算图片表。
- `backend/migrations/240_video_billing_review.sql`：让待核对视频继续保有上游唯一绑定。
- `docs/BILLING_INTEGRITY.md`：记录正式行为、迁移、运维检查与回滚限制。
- `progress.md`：仅追加三个任务闭环记录。
- 回滚：在仓库根目录执行 `git apply --reverse --check /tmp/sub2api-billing-fix.kI6FJx/billing-fix.patch`，通过后执行 `git apply --reverse /tmp/sub2api-billing-fix.kI6FJx/billing-fix.patch`；补丁仅包含本次 27 个源码/测试/迁移/文档文件，保留既有定制和 progress 历史。修改前快照为 `/tmp/sub2api-billing-fix.kI6FJx/before.tar`。线上若已执行迁移，回退代码不删除新表、队列或账务记录；旧二进制不会处理新队列/待核对状态，须先保存并核对待办。

## 2026-09-13 - Task: 在用户接口文档的视频模型 SD2.0 中补充 ViralDance 933 两个模型

### What was done

- 在现有 SD2.0 分类新增 `viraldance933`、`viraldance933-fast`，沿用模型表格、参数表格和示例卡片风格。
- 根据用户提供的 933 上游文档补充 4–15 整数秒、720p、ratio、URL 参考素材和主体绑定参数；将 Firefly 专属文件、分镜与下载说明标明适用范围。
- 补充标准版文生视频、Fast 多模态参考、查询和下载三个示例，说明成功状态与 video.url/result_url/url 读取顺序。
- 仅更新文档展示及既有页面测试的示例数量，不修改路由、模型映射或计费行为。

### Testing

- `cd frontend && pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts`：1 个测试通过，既有五个分类与其他模型示例保持通过。
- `cd frontend && pnpm run typecheck`：通过。
- `cd frontend && pnpm exec eslint src/views/user/ApiDocsView.vue src/views/user/__tests__/ApiDocsView.spec.ts`：通过。
- 提取三个新增示例，仅执行 `bash -n` 检查 Shell 语法、`JSON.parse` 检查创建请求 JSON：通过；未执行 curl 或调用上游生成视频。
- 三个文档/测试文件对本轮修改前快照执行 `git diff --no-index --check`，无空白错误。

### Notes

- `frontend/src/views/user/ApiDocsView.vue`：新增两个模型的规格、参数、调用示例与结果处理说明，保持原布局样式。
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`：既有 SD2.0 示例卡片数量从 4 调整为 7。
- `docs/API_DOCS.md`：同步四个 SD2.0 模型的调用方式与适用范围。
- `progress.md`：追加本轮实现、验证和回滚记录。
- 回滚点：本轮修改前的三个文件保存在 `/tmp/sub2api-viraldance-docs.hHhJkd/`。在仓库根目录分别执行以下命令恢复本轮文档改动；恢复前核对没有后续编辑，保留 progress 历史：

```sh
cp /tmp/sub2api-viraldance-docs.hHhJkd/ApiDocsView.vue frontend/src/views/user/ApiDocsView.vue
cp /tmp/sub2api-viraldance-docs.hHhJkd/ApiDocsView.spec.ts frontend/src/views/user/__tests__/ApiDocsView.spec.ts
cp /tmp/sub2api-viraldance-docs.hHhJkd/API_DOCS.md docs/API_DOCS.md
```

- 本轮未构建/推送镜像、未部署服务，线上页面需随新前端构建发布后生效。

## 2026-09-13 - Task: 在用户接口文档新增视频模型 Wan3.0 分类

### What was done

- 在 SD2.0 之后新增“视频模型Wan3.0”，展示 `wan3.0x`、`wan3.0x-480p`、`wan3.0x-1080p`，沿用现有模型表格、参数表格和示例卡片。
- 补充模型名决定分辨率、2–30 整数秒、ratio、参考素材数量与引用规则，以及异步状态和 metadata.url/url 结果读取说明。
- 提供 480p 文生视频、720p 单图生视频、1080p 多模态参考、查询和下载四个示例，API 地址、密钥、素材与结果地址均使用占位符，不披露供应商域名、凭据或接入配置。
- 仅修改用户文档内容、对应维护文档和既有分类测试，不修改转发或计费行为。

### Testing

- `cd frontend && pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts`：1 个测试通过，分类由五个调整为六个，其余分类验证通过。
- `cd frontend && pnpm run typecheck`：通过。
- `cd frontend && pnpm exec eslint src/views/user/ApiDocsView.vue src/views/user/__tests__/ApiDocsView.spec.ts`：通过。
- 四个新增示例经过 `bash -n` 语法检查，三个创建请求经过 `JSON.parse` 检查：通过；未执行网络请求或生成付费视频。
- 三个文档/测试文件对本轮修改前快照执行 `git diff --no-index --check`，无空白错误；人工核对新增页面内容未包含供应商地址或密钥。

### Notes

- `frontend/src/views/user/ApiDocsView.vue`：新增 Wan3.0 分类、三个模型的参数和四个用户调用示例，保留原布局样式。
- `frontend/src/views/user/__tests__/ApiDocsView.spec.ts`：适配六个分类、新增标题和后续分类索引。
- `docs/API_DOCS.md`：同步 Wan3.0 用户文档范围与参数、结果处理说明。
- `progress.md`：追加本轮修改、验证及回滚记录。
- 回滚点：本轮修改前文件保存在 `/tmp/sub2api-wan-docs.EUoUQI/`，包含上一轮 ViralDance 文档成果。在仓库根目录执行以下命令仅恢复本轮文档改动；恢复前核对没有后续编辑，保留进度历史：

```sh
cp /tmp/sub2api-wan-docs.EUoUQI/ApiDocsView.vue frontend/src/views/user/ApiDocsView.vue
cp /tmp/sub2api-wan-docs.EUoUQI/ApiDocsView.spec.ts frontend/src/views/user/__tests__/ApiDocsView.spec.ts
cp /tmp/sub2api-wan-docs.EUoUQI/API_DOCS.md docs/API_DOCS.md
```

- 本轮未构建镜像或部署，线上文档需随新前端构建发布后生效。

## 2026-09-13 - Task: 构建并推送包含计费修复和视频文档的新镜像

### What was done

- 基于完整当前工作区构建并推送 `iotwq/china-api:latest` 与版本标签 `iotwq/china-api:0.2.4-20260913-173021`，包含既有定制、近期计费修复和新增视频模型文档。
- 发布 linux/amd64 与 linux/arm64，两个标签均指向 OCI index `sha256:d21b296f66b0dde95e3de86f0cbb86ae2ab28e0373420417549d53bdd0d4c4b0`。
- 版本信息为 `0.2.4`、commit `cb76eeba0dd6-dirty`、构建时间 `2026-09-13T09:30:21Z`，未提交或改写现有源码。

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.4 --build-arg COMMIT=cb76eeba0dd6-dirty --build-arg DATE=2026-09-13T09:30:21Z --tag iotwq/china-api:latest --tag iotwq/china-api:0.2.4-20260913-173021 --metadata-file /tmp/sub2api-image-20260913.154ccO/metadata.json --progress=plain --push .`：重试后退出 0，完成翻译检查（3 个测试）、类型检查、前端生产构建及两个架构 Go 编译和推送。
- 分别执行 `docker buildx imagetools inspect` 检查两个标签，摘要与构建 metadata 一致。amd64 manifest：`sha256:d689738daa76215c59302e59b7420d65278cfc1b3959ee1efed5f1fdf237c4bf`；arm64 manifest：`sha256:31ca6168a17519d0b11e930c8873dfa132282d8de0c200ce52cd0a4cfacb028e`。
- 两个架构从 Docker Hub 拉取后，通过 `docker run --rm --platform <架构> --network none --read-only --entrypoint /app/sub2api <发布摘要> --version` 验证：均退出 0，版本、commit 与构建时间一致。
- 在 ARM64 发布镜像的 `/app/sub2api` 中核对“视频模型Wan3.0”、ViralDance 示例、239 待结算图片迁移与 240 视频账务待核对迁移：全部存在。
- 首次构建在 Go 依赖下载时发生 EOF；重试成功，记录在 `build.log`、`build-retry.log`。首次 amd64 拉取遇到 CDN EOF，重试成功，记录在 `version-amd64.log`、`version-amd64-retry.log`。全部日志目录 `/tmp/sub2api-image-20260913.154ccO/`。
- 本轮不重复执行刚已通过的计费与文档回归，不连接生产数据库或执行实际生成请求；镜像版本检查不等于生产请求链路验收。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加发布标签、摘要、验证、迁移与回滚说明。
- `progress.md`：追加本轮构建发布记录。
- 本轮未部署或重启线上服务、未修改数据库或余额；生产环境需拉取镜像并重新创建应用容器后生效。
- 回滚点：发布前 `iotwq/china-api@sha256:83089738ee533188e404a1dd5781ce2d81735f05537c77d7e4db56819ae8dc4f`。恢复 latest 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:83089738ee533188e404a1dd5781ce2d81735f05537c77d7e4db56819ae8dc4f`；本轮未执行回滚。若已部署新版本，保留迁移后的队列与账务数据，先核对待结算/待审核任务再回退旧二进制。

## 2026-09-13 - Task: 明确渠道智力检测糖果题的操作规则

### What was done

- 在实际发送的糖果题中明确允许取出前凭手感选择形状、不能预先辨别口味、仅实际取出计数且不能放回，可事先安排各形状数量，减少被理解为完全随机抓取的歧义。
- 保留标准答案 21 和唯一最终答案规则，题面不包含解法或答案；将既有错误回答回归调整为用户实际遇到的完整答案 29。
- 同步文档中的题面、固定选取策略证明及 29 对应的不同操作规则，说明后续检测生效及新旧结果比较边界。
- 复核现有用户页面已提供“不代表模型身份或综合能力”的悬停提示，无需新增重复文案。

### Testing

- `cd backend && go test -tags=unit ./internal/service -run '^(TestMonitorCandyFinalAnswer|TestMonitorCandyReferenceAnswerByEnumeration|TestMonitorIntelligenceProviderRequests|TestMonitorIntelligenceWrongAnswerDoesNotRetryOrFailHealth|TestMonitorIntelligenceDefaultOffAndConfigValidation)$' -count=1`：通过。
- 覆盖 OpenAI Chat Completions/Responses、Anthropic、Gemini 均发送新题面，21 判绿、完整答案 29 判红但不重试或改变可用状态，截断/无唯一答案不判绿，以及最坏情况最小答案仍为 21。
- 两个 Go 文件 `gofmt -l` 无输出；三个改动文件与本轮修改前快照的 `git diff --no-index --check` 无空白错误。
- 未调用真实计费模型进行前后对照；测试验证题面传递和程序判分，不代表真实模型答题率已经改善。

### Notes

- `backend/internal/service/channel_monitor_intelligence.go`：补充糖果题操作规则，保留原参数、预算和答案。
- `backend/internal/service/channel_monitor_intelligence_test.go`：现有错误回答用例改为完整 29 响应，验证答错与渠道可用性分离。
- `docs/channel-monitor-intelligence.md`：说明新规则、答案依据和部署前后历史结果边界。
- `progress.md`：追加本轮修改、验证和回滚记录。
- 本轮未构建/推送镜像、未修改线上配置或数据库；部署新构建后用于后续检测，历史记录不重判。历史记录没有题目版本字段，应按部署时间区分新旧题面。
- 回滚点：修改前文件保存在 `/tmp/sub2api-candy-prompt.W1fujo/`。确认没有后续重叠编辑后，在仓库根目录执行以下命令恢复本轮三个文件，保留 progress 历史：

```sh
cp /tmp/sub2api-candy-prompt.W1fujo/channel_monitor_intelligence.go backend/internal/service/channel_monitor_intelligence.go
cp /tmp/sub2api-candy-prompt.W1fujo/channel_monitor_intelligence_test.go backend/internal/service/channel_monitor_intelligence_test.go
cp /tmp/sub2api-candy-prompt.W1fujo/channel-monitor-intelligence.md docs/channel-monitor-intelligence.md
```

## 2026-09-13 - Task: 构建并推送糖果题规则完善镜像

### What was done

- 基于当前工作区构建并推送 `iotwq/china-api:latest` 及固定版本标签 `iotwq/china-api:0.2.4-20260913-200554`，支持 linux/amd64 与 linux/arm64。
- 镜像包含明确形状选择与计数规则的新糖果题，保留标准答案 21 及此前计费修复、视频文档等已有内容。
- 发布 OCI index：`sha256:a92d2fae1eb99fa4ed0ba545002bbcc31077efab44bfc33b791f0c4bdbe02b08`；版本信息为 `0.2.4`、commit `cb76eeba0dd6-dirty`、构建时间 `2026-09-13T12:05:54Z`。

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.4 --build-arg COMMIT=cb76eeba0dd6-dirty --build-arg DATE=2026-09-13T12:05:54Z --tag iotwq/china-api:latest --tag iotwq/china-api:0.2.4-20260913-200554 --metadata-file /tmp/sub2api-image-candy.S7rd1g/metadata.json --progress=plain --push .`：退出 0，双架构后端编译和推送成功；前端为上一轮同内容的已验证构建缓存。
- 两个标签分别执行 `docker buildx imagetools inspect`，远端摘要与 metadata 一致。amd64 manifest：`sha256:6119a48f47f4b0948af517fd41b777fce3221d69beda2b4aa8fa5374ae6ed504`；arm64 manifest：`sha256:f47b3f3f101c11fa20b8834e858b54828b096712c7fbedc2e598812943e9ccc0`。
- 两个架构按发布 index 从 Docker Hub 拉取，通过 `docker run --rm --platform <架构> --network none --read-only --entrypoint /app/sub2api <发布摘要> --version` 验证，均退出 0，版本、commit、时间一致。
- 在 ARM64 发布镜像 `/app/sub2api` 中确认完整形状选择、未取出不计数及不能放回的规则字符串存在，检查退出 0。
- 沿用上一轮已通过的题面传递、判分与最坏情况枚举回归，不重复执行无变化的测试。构建/版本检查不包含真实模型答题率验证。
- 构建及检查日志：`/tmp/sub2api-image-candy.S7rd1g/build.log`、`metadata.json`、`version-arm64.log`、`version-amd64.log`。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加本次发布标签、摘要、验证、部署和回滚说明。
- `progress.md`：追加本轮发布闭环记录。
- 本轮未修改业务源码、线上数据库或服务；部署后仅新的检测使用完善后的题面，原历史结果保留。
- 回滚：在需要时执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:d21b296f66b0dde95e3de86f0cbb86ae2ab28e0373420417549d53bdd0d4c4b0` 恢复发布前 latest，或重新部署 `iotwq/china-api:0.2.4-20260913-173021`。本轮未执行回滚；题面变更不新增数据库迁移，保留所有历史及账务数据。

## 2026-09-13 - Task: 分离渠道智力检测与普通探测的延迟和健康状态

### What was done

- 开启智力检测时，每个模型先执行普通探活，成功且本轮未取消后再单独请求糖果题；糖果题仅补充智力结果，渠道状态、延迟、消息与检测时间保留普通探活的值。
- 糖果题的长耗时、错误答案、请求失败、截断或超时不再引起渠道降级或改变可用率；普通探活本身响应缓慢仍遵循原有阈值，失败则跳过糖果题并记录智力未完成。
- 保留关闭开关时的行为、请求参数、账号重试和 OpenAI 协议回退，移除糖果题单独进入慢响应判断的逻辑，不新增数据库字段或并发流程。
- 更新中英文开关说明和使用文档，说明每轮增加轻量请求、共享整轮时限、统一刷新，以及旧历史延迟不回写。

### Testing

- 修复前运行 `cd backend && go test -tags=unit ./internal/service -run '^TestMonitorIntelligence(SlowAnswerDoesNotDegradeHealth|ProviderRequests)$' -count=1`，成功复现回归失败：糖果题耗时超过 6 秒且答对仍返回 degraded，多协议检测延迟也取自糖果题而非普通探活。
- 修复后 `cd backend && go test -tags=unit ./internal/service ./internal/handler ./internal/handler/admin -run 'Test.*(ChannelMonitor|Monitor|Challenge|RunCheck)' -count=1`：三个包均通过。
- 回归覆盖 OpenAI Chat/Responses、Anthropic、Gemini 的两次独立请求及输出预算；糖果题超过 6 秒不降级；普通探活 7000 ms 仍降级；错误答案不重试；智力 HTTP 错误、取消、截断、空响应不污染健康结果；普通探活失败跳过糖果题；OpenAI 回退保留普通探活的延迟和消息；整轮真实截止时间耗尽后，已成功的普通探活状态及 123 ms 延迟仍写入历史，智力结果为未完成。
- `cd frontend && pnpm exec vitest run src/components/user/__tests__/MonitorIntelligence.spec.ts src/views/admin/__tests__/MonitorFormDialog.accountSelector.spec.ts src/components/admin/monitor/__tests__/MonitorPrimaryModelCell.spec.ts src/i18n/__tests__/localeKeyCompleteness.spec.ts`：4 个文件、20 项测试全部通过。
- `cd frontend && pnpm exec eslint src/i18n/locales/zh/dashboard.ts src/i18n/locales/en/dashboard.ts`：通过。两个 Go 文件 `gofmt -l` 无输出，五个改动文件相对本轮快照的 `git diff --no-index --check` 无空白错误；旧说明关键词检索无残留。
- 使用本地模拟接口验证，未请求真实计费模型；本轮未构建或推送镜像、未部署服务或修改数据库。

### Notes

- `backend/internal/service/channel_monitor_checker.go`：普通探活与糖果题分开执行，仅合并智力结果，取消糖果题慢响应降级。
- `backend/internal/service/channel_monitor_intelligence_test.go`：调整既有多协议和错误答案测试，新增耗时、错误隔离、协议回退及超时历史保存回归。
- `frontend/src/i18n/locales/zh/dashboard.ts`：开关说明改为独立请求与独立状态、延迟。
- `frontend/src/i18n/locales/en/dashboard.ts`：同步英文开关说明。
- `docs/channel-monitor-intelligence.md`：更新检测流程、时间和状态口径、请求成本、刷新及部署后的历史边界。
- `progress.md`：追加本轮实施、验证与回滚记录。
- 开启智力检测的每个模型每轮增加一次轻量请求；本轮结果仍在全部检测完成或超时后统一刷新。部署新构建后仅影响后续检测，旧历史延迟不重新计算。
- 回滚点：修改前五个文件保存在 `/tmp/sub2api-intelligence-latency.hhhFNB/`。确认没有后续重叠修改后，在仓库根目录执行以下命令，仅恢复本轮文件并保留 progress 历史：

```sh
for task_file in backend/internal/service/channel_monitor_checker.go backend/internal/service/channel_monitor_intelligence_test.go frontend/src/i18n/locales/zh/dashboard.ts frontend/src/i18n/locales/en/dashboard.ts docs/channel-monitor-intelligence.md; do
  cp "/tmp/sub2api-intelligence-latency.hhhFNB/$task_file" "$task_file"
done
```

## 2026-09-13 - Task: 将 OpenAI 智力检测推理强度固定为 low

### What was done

- 在 OpenAI 智力检测请求完成模板合并后显式设置 low：Chat Completions 使用 reasoning_effort，Responses 使用 reasoning.effort；覆盖模板中的 high/xhigh，并清除混入的另一协议推理字段。
- 保留 Responses 的其他 reasoning 设置、原有 8192-token 输出预算和糖果题判分，Chat 自动回退 Responses 时同样使用 low。
- 普通探活保留原模板或默认推理参数，不修改保存的监控配置，也不向其他 provider 注入 OpenAI 参数。
- 按 OpenAI 官方推理指南及 Chat Completions 参数文档核对字段格式，同步说明适用范围、第三方执行边界及部署前后结果不可直接比较。

### Testing

- 修改前执行 `cd backend && go test -tags=unit ./internal/service -run '^(TestMonitorIntelligenceProviderRequests|TestMonitorIntelligenceWrongAnswerDoesNotRetryOrFailHealth|TestMonitorIntelligenceResponsesFallbackPreservesHealth)$' -count=1`：预期回归失败，确认模板仍发送 high/xhigh、无模板及协议回退没有显式 low。
- 修改后执行 `cd backend && go test -tags=unit ./internal/service -run '^(TestMonitorIntelligence.*|TestMonitorCandyFinalAnswer|TestChannelMonitorIntelligenceLifecycleAndTimedOutHistory|TestRunCheckForModel.*)$' -count=1`：通过。
- 本地模拟 HTTP 接口验证默认配置和混合模板均按协议发送 low；普通探活参数及保存的模板未被改写；Responses summary 保留；协议回退、答错不重试、超时历史保存及智力与健康延迟分离回归均通过。
- 两个 Go 文件 `gofmt -l` 无输出；本轮三个改动文件相对修改前快照的 `git diff --no-index --check` 无空白错误。
- 未调用真实计费模型验证第三方是否遵循参数；未构建或推送镜像、未部署服务或修改数据库。

### Notes

- `backend/internal/service/channel_monitor_intelligence.go`：在 OpenAI 智力检测请求中固定 low，按协议规范化推理字段。
- `backend/internal/service/channel_monitor_intelligence_test.go`：扩展实际请求捕获断言，验证模板覆盖、默认参数、普通探活隔离及协议回退。
- `docs/channel-monitor-intelligence.md`：更新推理强度、协议字段、适用范围、官方依据和历史比较边界。
- `progress.md`：追加本轮实施、验证和回滚记录。
- 需要部署包含本次修改的新构建后用于后续检测；其他 provider 仍使用原参数，历史结果不重算。
- 回滚点：修改前文件保存在 `/tmp/sub2api-intelligence-low.1Yab3Q/`。确认没有后续重叠编辑后，在仓库根目录执行以下命令，仅恢复本轮三个文件并保留进度历史：

```sh
for task_file in backend/internal/service/channel_monitor_intelligence.go backend/internal/service/channel_monitor_intelligence_test.go docs/channel-monitor-intelligence.md; do
  cp "/tmp/sub2api-intelligence-low.1Yab3Q/$task_file" "$task_file"
done
```

## 2026-09-13 - Task: 构建并推送智力检测 low 与延迟分离镜像

### What was done

- 基于当前完整工作区构建并推送 `iotwq/china-api:latest` 与固定版本 `iotwq/china-api:0.2.4-20260913-224529`，包含智力检测与普通探活分离、OpenAI/Codex 智力检测固定 low，以及此前全部既有定制。
- 发布 `linux/amd64` 和 `linux/arm64`，版本为 `0.2.4`，commit `cb76eeba0dd6-dirty`，构建时间 `2026-09-13T14:45:29Z`。
- 两个标签共同的 OCI index 为 `sha256:451de5fe8933c21d1544a9781c1804828b2c066afb7b725fd1213ae315b59e63`，已记录发布前摘要及回滚方法。

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg GOLANG_IMAGE=golang:1.27.0-alpine --build-arg VERSION=0.2.4 --build-arg COMMIT=cb76eeba0dd6-dirty --build-arg DATE=2026-09-13T14:45:29Z --tag iotwq/china-api:latest --tag iotwq/china-api:0.2.4-20260913-224529 --metadata-file /tmp/sub2api-image-low.5q3AhM/metadata.json --progress=plain --push .`：退出 0。
- 构建内前端翻译 3 项检查、类型检查、生产打包和双架构 Go 编译通过；保留既有前端大 chunk 提示。源码阶段已通过的监控、智力及参数覆盖回归本轮不重复运行。
- 两个标签分别执行 `docker buildx imagetools inspect`，远端 index 摘要与 metadata 一致；amd64 manifest 为 `sha256:ee9c31e0a360a31f8f35671f8f5355a8be901d25755f1777ad1dca91b7e68263`，arm64 manifest 为 `sha256:8baac637b92a2579c62851d9bce17cda76ff46b26e58ac5da6abb960dd11bdbf`。
- 分别执行 `docker pull --platform linux/<架构> iotwq/china-api@<对应 manifest>`，随后通过 `docker run --rm --platform linux/<架构> --network none --read-only --entrypoint /app/sub2api iotwq/china-api@<对应 manifest> --version` 验证，两个架构均退出 0，版本、commit、构建时间一致。
- 构建日志及验证证据：`/tmp/sub2api-image-low.5q3AhM/build.log`、`metadata.json`、`version-amd64.log`、`version-arm64.log`。版本检查不等同于生产数据库及真实上游请求验收。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加发布标签、摘要、双架构验证、部署及回滚说明。
- `progress.md`：追加本轮发布闭环记录。
- 本轮未修改业务源码、未部署或重启线上服务、未修改数据库。拉取并重建应用容器后，后续检测使用 low 及独立延迟逻辑，旧历史不重算。
- 回滚点为发布前 `iotwq/china-api@sha256:a92d2fae1eb99fa4ed0ba545002bbcc31077efab44bfc33b791f0c4bdbe02b08`。需要时可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:a92d2fae1eb99fa4ed0ba545002bbcc31077efab44bfc33b791f0c4bdbe02b08` 恢复 latest，或重新部署 `iotwq/china-api:0.2.4-20260913-200554`。本轮未执行回滚，不删除历史或账务数据。

## 2026-09-15 - Task: 同步原项目最新提交至本地 0.2.5

### What was done

- 将原项目 upstream/main 从 98d86915b 同步到 881f32026，共纳入 197 个提交，生成本地合并提交 a85098e8779876b7a5334ff94099e94564e46198。
- 合并上游原生 OAuth 图片、图片缓存计费、WebSocket、OpenCode Go 和管理页面更新；处理基线及本地改动恢复冲突，保留本地账务修复、视频、群聊、余额查询、端点展示、模型列表及智力检测等定制。
- 合并 Gemini 错误观测与图片用量计数并增加联合回归；适配新增测试调用及图片接口分类，修正既有迁移测试的旧索引名称；重新生成依赖注入和 Ent 代码。
- 为所有原有改动保留分支、stash 和文件备份，恢复原有暂存意图；117 个原有未跟踪文件内容与备份一致。

### Testing

- `cd backend && go generate ./cmd/server && go generate ./ent`：通过。
- `cd frontend && pnpm exec vitest run`：301 个文件、2,274 项测试通过；`pnpm lint:check` 与 `pnpm build`（翻译/类型/生产打包）通过。
- `cd backend && go test -tags=unit ./... -count=1`：执行全量检查，发现新增测试调用接口与本地定制不一致；修正后 handler、routes 包复测通过，其余非 service 包已通过。首次 service 复测被错误的测试返回值断言阻塞，定位后修正并重新完整执行。
- `go test -tags=unit ./internal/service -count=1 -json`：14,142 项通过、3 项原有测试跳过、1 项旧 Responses 流测试使用了已切换原生接口的模型；修正模型后 `go test -tags=unit ./internal/service -run '^TestOpenAIGatewayServiceForwardImages_OAuthStreamingReturnsResponseFailedReason$' -count=1` 通过。最终采用全量检查加失败项定向复测，无遗留已知失败；未重复执行已通过且代码未变的其他用例。
- `go test -tags=unit ./internal/service -run '^TestCollectGeminiSSEObserved_' -count=1`：通过，覆盖图片累计去重、正常结束/读取失败时保留计费信息及上游终止信号。监控端点和后台 OAuth 图片测试定向复测通过。
- `CI=1 go test -tags=integration ./internal/repository -run 'TestMigrationsRunner_|TestUsageLogRepositoryCreateSyncRequestTypeAndLegacyFields|TestUsageLogRepositoryCreate_PersistsServiceTier|TestExecUsageLogInsertNoResult_PersistsRequestedModel' -count=1 -v`：临时 PostgreSQL/Redis 中通过，验证迁移共存、幂等及实际用量入库。
- `bash deploy/tests/apple-container-test.sh`、`bash deploy/tests/docker-compose-gateway-env-test.sh`：通过，使用测试夹具；`git diff --check`、`git diff --cached --check`、上游提交祖先关系及未合并项检查通过。
- 日志与备份目录：`/tmp/sub2api-sync-20260915.I44YmZ/`。未调用真实付费上游，未对生产数据库或运行中服务做验收。

### Notes

- 实际同步变更的 448 个文件及逐项说明列在 `docs/upstream-sync-20260915-files.md`，涵盖源码、测试、生成代码、配置和远端文档；兼容处理文件单独说明，避免把原有本地改动记为本次新增。
- `docs/UPSTREAM_SYNC.md`：追加同步版本范围、兼容结果、验证证据、迁移影响及恢复命令。
- `docs/upstream-sync-20260915-files.md`：新增本轮完整文件清单。
- `progress.md`：追加本次同步闭环记录。
- 新增远端 238 OpenCode 平台和未配置额度行清理迁移，本轮仅同步文件；后续服务启动时执行。前端原有大 chunk 和 Browserslist 提示保留。本轮未构建或推送镜像、未部署、未修改生产账务。
- 回滚点：分支 `codex/pre-upstream-sync-20260915`（cb76eeba0dd6）与 stash `77d38b39662c85487868baef90ff6170f719b4ca`。在仓库根目录执行以下命令可生成同步前的独立恢复工作区，不覆盖当前文件：

```sh
git worktree add -b codex/recover-pre-sync-20260915 /tmp/sub2api-pre-sync-20260915 codex/pre-upstream-sync-20260915
git -C /tmp/sub2api-pre-sync-20260915 stash apply --index 77d38b39662c85487868baef90ff6170f719b4ca
```

## 2026-09-15 - Task: 群聊体验修复 1/9：断线补拉

### What was done

- 群聊与私聊立即并行捕获重连游标并补拉，避免新实时消息跨过离线缺口；空会话重新取历史。

### Testing

- `cd frontend && pnpm exec vitest run src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：先复现 4 项失败，修复后 4 项通过，包括空群聊/私聊与互不阻塞的补拉。
- 证据：`/tmp/sub2api-chat-fixes.8W9IVo/01-before.log`、`01-after.log`。

### Notes

- `frontend/src/views/user/CommunityChatView.vue`：修正重连补拉时序和空会话处理。
- `frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：新增真实组件交互测试及 mock 接口夹具。
- `docs/COMMUNITY_CHAT.md`：追加重连语义。
- `progress.md`：追加本轮记录。
- 回滚：仓库根目录执行 `tar -xzf /tmp/sub2api-chat-fixes.8W9IVo/before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue docs/COMMUNITY_CHAT.md`；将本轮新增测试移动到备份目录：`mv frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts /tmp/sub2api-chat-fixes.8W9IVo/behavior-rolled-back.spec.ts`。仅在没有后续重叠改动时执行；保留进度历史。

## 2026-09-15 - Task: 群聊体验修复 2/9：私聊历史与实时消息合并

### What was done

- 历史响应按 ID 合并加载期间的新消息，同时保留跨会话请求隔离，重新打开替换过期快照。

### Testing

- 真实组件回归先复现 1 项失败，修复后 7 项通过；新增覆盖历史/实时交错、重新打开、快速切换用户。日志：`/tmp/sub2api-chat-fixes.8W9IVo/02-before.log`、`02-after.log`。

### Notes

- `frontend/src/views/user/CommunityChatView.vue`：更新加载开始及历史合并逻辑。
- `frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：新增三项会话隔离回归。
- `docs/COMMUNITY_CHAT.md`：记录加载与实时合并语义。
- `progress.md`：追加记录。
- 回滚（无后续重叠编辑时）：`tar -xzf /tmp/sub2api-chat-fixes.8W9IVo/02-before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts docs/COMMUNITY_CHAT.md`。保留进度历史。

## 2026-09-15 - Task: 群聊体验修复 3/9：私聊已读上限

### What was done

- 已读请求携带页面实际展示的消息上限，服务端限定到对应会话内的该上限，避免提前消除未展示消息的未读提醒；加载历史时不提交已读。

### Testing

- 新增组件场景先复现失败；修复后组件/前端 API/未读相关 24 项通过。后端 service、handler、repository 定向测试通过，包含非法参数和 SQL 上限回归。证据：`/tmp/sub2api-chat-fixes.8W9IVo/03-before.log`、`03-after.log`、`03-backend.log`。

### Notes

- `frontend/src/views/user/CommunityChatView.vue`：加载结束后才标记已读，传入已展示消息上限。
- `frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：新增历史加载期间禁止已读及消息上限回归。
- `frontend/src/api/communityChat.ts`：已读请求携带 last_message_id。
- `frontend/src/api/__tests__/communityChat.spec.ts`：验证新的请求体。
- `frontend/src/views/user/__tests__/communityChatUnread.spec.ts`：更新已读接口断言。
- `backend/internal/handler/community_chat_handler.go`：校验并传递已展示消息 ID。
- `backend/internal/service/community_chat.go`：在既有会话权限校验后传递已读上限。
- `backend/internal/repository/community_chat_repo.go`：SQL 限定到客户端已展示 ID，保留单调推进。
- `backend/internal/handler/community_chat_handler_test.go`：覆盖缺省及非法上限返回 400。
- `backend/internal/repository/community_chat_repo_test.go`：覆盖 SQL 按会话及已展示 ID 限定。
- `docs/COMMUNITY_CHAT.md`：更新接口参数和旧页刷新说明。
- `progress.md`：追加记录。
- 请求参数变化在本次用户授权的已读修复范围内，无数据库结构变更；旧页面需刷新。回滚（无后续重叠编辑时）：`tar -xzf /tmp/sub2api-chat-fixes.8W9IVo/03-before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts frontend/src/api/communityChat.ts frontend/src/api/__tests__/communityChat.spec.ts frontend/src/views/user/__tests__/communityChatUnread.spec.ts backend/internal/handler/community_chat_handler.go backend/internal/service/community_chat.go backend/internal/repository/community_chat_repo.go backend/internal/handler/community_chat_handler_test.go backend/internal/repository/community_chat_repo_test.go docs/COMMUNITY_CHAT.md`。保留进度历史。

## 2026-09-15 - Task: 群聊体验修复 4/9：私聊草稿隔离

### What was done
- 管理员切换私聊对象时分别保存文字草稿，避免把上一人的回复误发给下一人；异步发送完成只清理原收件人的已提交草稿。

### Testing
- 真实页面交互回归先复现 2 个失败，修复后 10 项通过，包含发送期间切换对象且输入相同文字。证据：`/tmp/sub2api-chat-fixes.8W9IVo/04-before.log`、`04-after.log`。

### Notes
- `frontend/src/views/user/CommunityChatView.vue`：按收件人隔离草稿及发送成功清理。
- `frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：新增切换及异步发送交互回归。
- `docs/COMMUNITY_CHAT.md`：说明草稿生命周期。
- `progress.md`：追加本轮记录。
- 回滚：确认无后续重叠修改后，在仓库根目录执行 `tar -xzf /tmp/sub2api-chat-fixes.8W9IVo/04-before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts docs/COMMUNITY_CHAT.md`；保留进度历史。

## 2026-09-15 - Task: 群聊体验修复 5/9：中文输入法防误发

### What was done
- 群聊和私聊在中文输入法确认阶段不发送消息，保留正常 Enter 发送及 Shift+Enter 换行。

### Testing
- 两种聊天输入框各先复现误发，修复后真实页面交互 12 项通过；日志 `/tmp/sub2api-chat-fixes.8W9IVo/05-before.log`、`05-after.log`。

### Notes
- `frontend/src/views/user/CommunityChatView.vue`：发送前识别组合输入及 keyCode 229。
- `frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：输入法、换行和正常发送回归。
- `docs/COMMUNITY_CHAT.md`：补充输入行为。
- `progress.md`：追加本轮记录。
- 回滚：确认无后续重叠修改后执行 `tar -xzf /tmp/sub2api-chat-fixes.8W9IVo/05-before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts docs/COMMUNITY_CHAT.md`，保留进度历史。

## 2026-09-15 - Task: 群聊体验修复 6/9：长记录输入渲染

### What was done
- 消息列表复用未变化的渲染结果，时间格式器按语言缓存，减少输入时历史消息的重复计算。

### Testing
- 群聊和私聊各加载 800 条记录，修复前一次输入创建 800 个格式器，修复后时间格式及日期解析调用均为 0；覆盖新消息内容、搜索高亮、删除按钮状态和语言切换。15 项真实页面交互回归通过。证据：`/tmp/sub2api-chat-fixes.8W9IVo/06-before.log`、`06-after.log`。
- jsdom 验证的是重复计算消除，不作为真实设备帧率或无限历史性能承诺。

### Notes
- `frontend/src/views/user/CommunityChatView.vue`：消息渲染依赖缓存及时间格式器复用。
- `frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：大列表输入和可见更新回归。
- `docs/COMMUNITY_CHAT.md`：说明优化范围。
- `progress.md`：追加本轮记录。
- 回滚：确认无后续重叠修改后执行 `tar -xzf /tmp/sub2api-chat-fixes.8W9IVo/06-before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts docs/COMMUNITY_CHAT.md`，保留进度历史。

## 2026-09-15 - Task: 群聊体验修复 7/9：会话列表静默刷新

### What was done
- 私信到达、发送和删除后的会话更新改为静默合并，保留分页和现有行位置；合并刷新期间的连续通知，避免并发响应倒序覆盖。

### Testing
- 修复前复现加载动画闪烁及 5 次通知触发 5 个并发刷新；修复后保留已加载 60 个会话和第 2 页状态，连续通知合并为当前及后续刷新，最终内容为最新。17 项页面交互测试通过，证据 `/tmp/sub2api-chat-fixes.8W9IVo/07-before.log`、`07-after.log`。删除入口同样使用此静默刷新函数，最终集成测试再次覆盖。

### Notes
- `frontend/src/views/user/CommunityChatView.vue`：后台会话刷新合并及前台分页协调。
- `frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：分页保留与突发事件回归。
- `docs/COMMUNITY_CHAT.md`：说明静默刷新行为。
- `progress.md`：追加本轮记录。
- 回滚：确认无后续重叠修改后执行 `tar -xzf /tmp/sub2api-chat-fixes.8W9IVo/07-before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts docs/COMMUNITY_CHAT.md`，保留进度历史。

## 2026-09-15 - Task: 群聊体验修复 8/9：附件会话自动恢复

### What was done
- 文字历史独立加载；附件会话失败后定时恢复，重连和回到页面时也尝试恢复，合并并发续期；恢复后只重载失败媒体。

### Testing
- 先复现 4 个失败，修复后 21 项页面回归通过，覆盖初始请求卡住不影响文字、30 秒失败重试、媒体选择性重载、恢复请求合并和卸载清理。证据 `/tmp/sub2api-chat-fixes.8W9IVo/08-before.log`、`08-after.log`。

### Notes
- `frontend/src/views/user/CommunityChatView.vue`：附件续期与失败媒体恢复生命周期。
- `frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：断网恢复、定时器和媒体交互验证。
- `docs/COMMUNITY_CHAT.md`：记录恢复时机和间隔。
- `progress.md`：追加本轮记录。
- 回滚：确认无后续重叠修改后执行 `tar -xzf /tmp/sub2api-chat-fixes.8W9IVo/08-before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts docs/COMMUNITY_CHAT.md`，保留进度历史。

## 2026-09-15 - Task: 群聊体验修复 9/9：上传进度与慢速网络

### What was done
- 群聊和私聊上传独立延长为 10 分钟超时，显示上传及服务端确认阶段；失败保留草稿/附件，私聊进度关联原收件人。

### Testing
- 先复现 4 项失败，修复后页面与 API 共 35 项通过；覆盖上传专用超时、其他请求仍为 30 秒、50% 进度、100% 未确认状态及失败保留内容。证据 `/tmp/sub2api-chat-fixes.8W9IVo/09-before.log`、`09-after.log`。

### Notes
- `frontend/src/api/communityChat.ts`：两个附件上传接口增加专用超时和进度回调。
- `frontend/src/api/__tests__/communityChat.spec.ts`：校验超时隔离及上传进度回调。
- `frontend/src/views/user/CommunityChatView.vue`：显示上传进度和确认阶段。
- `frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：两个输入区的进度与失败保留测试。
- `frontend/src/i18n/locales/zh/misc.ts`：中文上传提示。
- `frontend/src/i18n/locales/en/misc.ts`：英文上传提示。
- `docs/COMMUNITY_CHAT.md`：上传行为及代理限制说明。
- `progress.md`：追加本轮记录。
- 回滚：确认无后续重叠修改后执行 `tar -xzf /tmp/sub2api-chat-fixes.8W9IVo/09-before.tar.gz -C . frontend/src/api/communityChat.ts frontend/src/api/__tests__/communityChat.spec.ts frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts frontend/src/i18n/locales/zh/misc.ts frontend/src/i18n/locales/en/misc.ts docs/COMMUNITY_CHAT.md`，保留进度历史。

## 2026-09-15 - Task: 群聊 9 项修复联合验证与收尾

### What was done

- 联合验证发现重连补拉期间实时私信会提前确认缺口内消息，补齐第 1/3 项关联场景：补拉期间暂停已读，成功合并并渲染后再推进。
- 修正附件重载写法的 lint 错误，增加删除通知同样保持会话分页的交互验证；前 3 项本轮新增日志归位到末尾，原进度历史逐字节保留。

### Testing

- `cd frontend && pnpm exec vitest run src/views/user/__tests__/CommunityChatView.behavior.spec.ts src/views/user/__tests__/communityChatComposer.spec.ts src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/views/user/__tests__/communityChatEmoji.spec.ts src/views/user/__tests__/communityChatSearch.spec.ts src/views/user/__tests__/communityChatScroll.spec.ts src/views/user/__tests__/communityChatUnread.spec.ts src/api/__tests__/communityChat.spec.ts src/composables/__tests__/useCommunityChatRealtime.spec.ts src/components/community/__tests__/CommunityEmojiPicker.spec.ts`：10 个文件、67 项通过，其中真实组件交互 25 项。重连已读场景先复现 1 项失败再修复，证据 `final-review-before.log`、`chat-final.log`。
- `cd frontend && pnpm lint:check`：通过。首次发现本轮图片重载自赋值，修正；并行构建一度造成临时配置文件被扫描后消失，改为串行执行 lint 后通过。证据 `lint-final.log`。
- `cd frontend && pnpm build`：翻译完整性 3 项、TypeScript 检查、生产打包通过；保留既有大 chunk/Browserslist 提示。证据 `build-final.log`。
- `cd backend && go test -tags=unit ./internal/service ./internal/handler ./internal/repository -run 'Test.*CommunityChat|Test.*DirectMessage' -count=1`：三个包通过。证据 `backend-final.log`；无数据库迁移，未运行生产账务或真实聊天发送。
- `git diff --check`、`git diff --cached --check` 无空白错误，未合并项为空；与修改前备份比对确认 `progress.md` 原历史字节未变化，所有本轮记录仅追加在后。
- 证据目录 `/tmp/sub2api-chat-fixes.8W9IVo/`。本轮验证覆盖组件模拟交互和后端单元测试，不等同于生产环境断网、代理超时和慢速大附件验收。

### Notes

- `frontend/src/views/user/CommunityChatView.vue`：补拉与已读串联、附件重载 lint 收尾及本轮时间格式器缩进。
- `frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：补拉已读和删除通知交互回归。
- `docs/COMMUNITY_CHAT.md`：联合验证与前后端同步发布说明。
- `progress.md`：追加验证结果并将本轮误插入的 1–3 项记录归位；未改变任何原历史内容。
- 其余实际修改文件已分别记录在 1–9 项 Notes；现有其他定制保持原状。本轮未制作或推送 Docker 镜像、未部署服务、未修改生产数据。
- 本次收尾回滚：确认无后续重叠修改后执行 `tar -xzf /tmp/sub2api-chat-fixes.8W9IVo/final-review-before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/CommunityChatView.behavior.spec.ts docs/COMMUNITY_CHAT.md`，保留进度历史。完整 9 项修复的分步回滚见前述记录；旧页面在发布后需刷新。

## 2026-09-15 - Task: 优化私聊站长入口图标与提示

### What was done

- 私聊入口图标从 20px 放大为 32px 对话图标，增加 44px 半透明图标底座，按钮加高到 64px 并使用暖红渐变、圆角和悬停阴影。
- 在入口下方增加“小问题请直接私聊”小字及英文翻译；保留桌面/手机居中布局、未读提醒及原点击动作。

### Testing

- `cd frontend && pnpm exec vitest run src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/views/user/__tests__/CommunityChatView.behavior.spec.ts src/i18n/__tests__/localeKeyCompleteness.spec.ts`：34 项通过；移除旧外观测试对 20px 信封及 48px 按钮的硬编码限制，没有新增纯样式回归测试。
- `cd frontend && pnpm exec eslint src/views/user/CommunityChatView.vue src/views/user/__tests__/communityChatEmojiWiring.spec.ts src/i18n/locales/zh/misc.ts src/i18n/locales/en/misc.ts`：通过。
- `cd frontend && pnpm build`：翻译、类型检查和生产打包通过；保留原有 Browserslist/大 chunk 提示。`git diff --check` 通过。
- 验证日志：`/tmp/sub2api-contact-owner.b26mRb/tests.log`、`lint.log`、`build.log`。未进行真实浏览器视觉验收，未发送私信或部署服务。

### Notes

- `frontend/src/views/user/CommunityChatView.vue`：放大并美化入口，加入说明小字和可访问描述。
- `frontend/src/i18n/locales/zh/misc.ts`：中文入口提示。
- `frontend/src/i18n/locales/en/misc.ts`：对应英文提示。
- `frontend/src/views/user/__tests__/communityChatEmojiWiring.spec.ts`：移除过期外观断言，保留原功能布局检查。
- `docs/COMMUNITY_CHAT.md`：记录入口外观及文案。
- `progress.md`：追加本轮结果和回滚方法。
- 回滚：确认无后续重叠编辑后，在仓库根目录执行 `tar -xzf /tmp/sub2api-contact-owner.b26mRb/before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/communityChatEmojiWiring.spec.ts frontend/src/i18n/locales/zh/misc.ts frontend/src/i18n/locales/en/misc.ts docs/COMMUNITY_CHAT.md`，保留进度历史。发布新的前端构建后生效。

## 2026-09-15 - Task: 私聊站长入口淡金配色与轻动效

### What was done

- 私聊入口改为香槟淡金渐变、深金文字与珠光图标底座，增加星光装饰，替换原暖红按钮配色。
- 添加图标轻微浮动、按钮间歇柔光掠过及悬停抬升；只使用 CSS transform/opacity 动画，系统选择减少动态效果时关闭。
- 保留“小问题请直接私聊”提示、原有入口尺寸、居中布局、未读红点和私聊打开逻辑。

### Testing

- `cd frontend && pnpm exec eslint src/views/user/CommunityChatView.vue`：通过。
- `cd frontend && pnpm build`：翻译完整性 3 项、类型检查和生产打包通过；保留原有 Browserslist/大 chunk 提示。证据 `/tmp/sub2api-contact-gold.ynBGrQ/build.log`、`lint.log`。
- 使用当前源码中的标题栏模板、图标路径及生产构建 CSS 制作临时预览，在本地浏览器检查 1000px 桌面和 390px 手机宽度的明暗主题；入口、文案及未读提醒显示正常，无裁切。仅验证入口外观，不等同于生产服务完整交互验收。
- `git diff --check` 通过；本轮纯外观调整未新增样式断言测试，未修改后端或发送私信。临时预览位于 `/tmp/sub2api-contact-gold.ynBGrQ/preview.cjs`，核验后已关闭预览页及监听进程。

### Notes

- `frontend/src/views/user/CommunityChatView.vue`：淡金按钮、珠光图标及动效与减少动态效果适配。
- `docs/COMMUNITY_CHAT.md`：更新当前入口外观与动效说明。
- `progress.md`：追加本轮实现、验证与回滚记录。
- 回滚：确认无后续重叠修改后，在仓库根目录执行 `tar -xzf /tmp/sub2api-contact-gold.ynBGrQ/before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue docs/COMMUNITY_CHAT.md`，保留进度历史。本轮未制作或推送镜像，未部署服务，发布新前端构建后生效。

## 2026-09-16 - Task: 构建并推送 0.2.5 群聊修复与淡金入口镜像

### What was done

- 使用当前完整源码工作区构建并推送 `iotwq/china-api:latest` 和 `iotwq/china-api:0.2.5-20260916-000512`，覆盖 amd64/arm64，包含近期上游同步、群聊九项修复及淡金私聊入口。
- 版本为 `0.2.5`，commit `a85098e87798-dirty`，构建时间 `2026-09-15T16:05:12Z`；保留原有暂存、未暂存及未跟踪业务源码，不提交或覆盖本地定制。

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.2.5 --build-arg COMMIT=a85098e87798-dirty --build-arg DATE=2026-09-15T16:05:12Z --tag iotwq/china-api:latest --tag iotwq/china-api:0.2.5-20260916-000512 --metadata-file /tmp/sub2api-image-chat-gold.guIwiq/metadata.json --progress=plain --push .`：退出 0；Dockerfile 使用 Go 1.27.0、Node 24。
- 镜像内翻译 3 项、前端类型检查/生产构建、双架构 Go 编译通过；源码阶段已验证的群聊回归沿用原证据，不重复无变化的测试。
- 两个标签分别执行 `docker buildx imagetools inspect`，远端 OCI index 与 metadata 一致：`sha256:480c7859fa16229ea44dccf8d29ddd9d8103e24a35f459142116bb660654262c`。
- amd64 manifest `sha256:4a18aa1556b8b4d16ce5d0786d4657355fb092ee599dedb5308f545605f51c38`、arm64 manifest `sha256:07e53417bd6ff7ca5f1033cd9e2e95704142e5caa60476fd03ce15170b951ff3` 分别使用 `docker pull --platform linux/<架构> iotwq/china-api@<manifest>` 拉取，随后 `docker run --rm --platform linux/<架构> --network none --read-only --entrypoint /app/sub2api iotwq/china-api@<manifest> --version` 均退出 0，版本信息一致。amd64 首次拉取 auth.docker.io 返回 EOF，重试后成功。
- `git diff --check`、`git diff --cached --check` 通过；无未解决合并项。构建保留既有前端大 chunk/Browserslist 提示。
- 证据 `/tmp/sub2api-image-chat-gold.guIwiq/`；镜像版本检查不等同于生产数据库和实际用户请求验收。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加发布标签、摘要、验证、部署影响及回滚说明。
- `progress.md`：追加本轮发布闭环记录。
- 本轮未改业务代码、未部署或重启服务、未修改生产数据。后续部署后需刷新旧页面，新版本启动将执行此前上游同步引入的迁移。
- 镜像回滚点为发布前 `iotwq/china-api@sha256:451de5fe8933c21d1544a9781c1804828b2c066afb7b725fd1213ae315b59e63`。恢复 latest 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:451de5fe8933c21d1544a9781c1804828b2c066afb7b725fd1213ae315b59e63`，或部署旧固定标签 `0.2.4-20260913-224529`；不回滚数据库，需核对上线后的数据库状态。本轮未执行回滚。

## 2026-09-16 - Task: 强化私聊站长入口未读提醒

### What was done

- 用户和站长的共用私聊入口由 10px 小红点改为右上角 24px 高的红色“新消息”徽标，加入铃铛和浅色描边。
- 徽标绝对定位，不挤动原金色入口；文字常亮，仅外圈轻微呼吸。增加屏幕阅读器状态提示，保留减少动态效果设置和原已读/未读逻辑。

### Testing

- `cd frontend && pnpm exec vitest run src/views/user/__tests__/communityChatUnread.spec.ts src/views/user/__tests__/CommunityChatView.behavior.spec.ts src/i18n/__tests__/localeKeyCompleteness.spec.ts`：34 项通过，含既有双向私聊、历史/已读交互及翻译完整性；更新原测试中过时红点外观断言，未新增镜像实现的样式测试。
- `cd frontend && pnpm exec eslint src/views/user/CommunityChatView.vue src/views/user/__tests__/communityChatUnread.spec.ts src/i18n/locales/zh/misc.ts src/i18n/locales/en/misc.ts`：通过。
- `cd frontend && pnpm build`：翻译、类型和生产打包通过，保留原有 Browserslist/大 chunk 提示。`git diff --check` 通过。
- 使用当前标题栏模板、图标路径与生产 CSS 的本地浏览器预览检查 1000px 桌面和 390px 手机宽度：亮色未读徽标清晰且不裁切，暗色已读状态保持居中。预览仅覆盖外观，不发送私信；已关闭临时页面和进程。
- 证据 `/tmp/sub2api-direct-badge.bgo39s/tests.log`、`lint.log`、`build.log`，临时预览代码 `preview.cjs`。

### Notes

- `frontend/src/views/user/CommunityChatView.vue`：入口角标、提示语及辅助访问状态，替换旧红点动画。
- `frontend/src/views/user/__tests__/communityChatUnread.spec.ts`：更新既有提示及辅助访问断言。
- `frontend/src/i18n/locales/zh/misc.ts`：增加“新消息”文案。
- `frontend/src/i18n/locales/en/misc.ts`：对应英文文案。
- `docs/COMMUNITY_CHAT.md`：记录徽标外观和状态语义。
- `progress.md`：追加本轮结果和回滚说明。
- 回滚：确认无后续重叠编辑后，在仓库根目录执行 `tar -xzf /tmp/sub2api-direct-badge.bgo39s/before.tar.gz -C . frontend/src/views/user/CommunityChatView.vue frontend/src/views/user/__tests__/communityChatUnread.spec.ts frontend/src/i18n/locales/zh/misc.ts frontend/src/i18n/locales/en/misc.ts docs/COMMUNITY_CHAT.md`，保留进度历史。本轮未改后端、未构建/推送 Docker 镜像、未部署服务；发布新前端后生效。

## 2026-09-16 - Task: 修复进入群聊页面后侧栏未读提醒提前隐藏

### What was done

- 移除侧栏未读提示对当前群聊路由的排除条件，进入群聊页本身不再隐藏提醒。
- 沿用现有独立群聊/私聊已读事件及服务端私聊未读查询；只阅读群聊或部分私聊后，其他未读仍保留提醒。本轮只修复此状态问题，不调整侧栏外观。

### Testing

- 新增真实 AppSidebar 组件和内存路由交互回归，先复现 5 项失败（均为导航后提示过早消失），修复后通过；覆盖普通用户/管理员与侧栏展开/收起四种组合，以及群聊独立阅读、私聊仍有未读和全部已读状态。
- `cd frontend && pnpm exec vitest run src/components/layout/__tests__/AppSidebar.spec.ts src/components/layout/__tests__/AppSidebar.unread.spec.ts src/views/user/__tests__/communityChatUnread.spec.ts src/views/user/__tests__/CommunityChatView.behavior.spec.ts`：4 个文件、51 项通过。证据 `/tmp/sub2api-sidebar-unread.TWG6hZ/before.log`、`after.log`。
- `cd frontend && pnpm exec eslint src/components/layout/AppSidebar.vue src/components/layout/__tests__/AppSidebar.unread.spec.ts` 与 `pnpm exec vue-tsc -b`：通过。证据 `lint.log`、`types.log`。
- `git diff --check` 通过。测试使用 mock 未读接口，未发送真实消息；本轮未改后端、未重新打包/推送镜像或部署。

### Notes

- `frontend/src/components/layout/AppSidebar.vue`：移除当前路由导致的未读展示抑制。
- `frontend/src/components/layout/__tests__/AppSidebar.unread.spec.ts`：新增导航和已读事件真实组件回归。
- `docs/COMMUNITY_CHAT.md`：说明侧栏提醒按阅读状态清除。
- `progress.md`：追加本轮验证和回滚记录。
- 回滚：确认无后续重叠编辑后，在仓库根目录执行 `tar -xzf /tmp/sub2api-sidebar-unread.TWG6hZ/before.tar.gz -C . frontend/src/components/layout/AppSidebar.vue docs/COMMUNITY_CHAT.md`，再将本轮新增测试移至备份目录：`mv frontend/src/components/layout/__tests__/AppSidebar.unread.spec.ts /tmp/sub2api-sidebar-unread.TWG6hZ/AppSidebar.unread.rolled-back.spec.ts`；保留进度历史。发布新的前端版本后生效。

## 2026-09-16 - Task: 构建并推送私聊徽标与侧栏未读修复镜像

### What was done

- 使用当前完整源码构建并推送 `iotwq/china-api:latest` 和固定标签 `iotwq/china-api:0.2.5-20260916-013453`，发布 amd64/arm64 镜像，包含私聊“新消息”徽标和侧栏未读提醒修复。
- 版本 `0.2.5`，commit `a85098e87798-dirty`，构建时间 `2026-09-15T17:34:53Z`；保留工作区所有既有定制与暂存状态，本轮未修改业务代码。

### Testing

- `docker buildx build --builder codex-multiarch --network=host --platform linux/amd64,linux/arm64 --provenance=false --sbom=false --build-arg VERSION=0.2.5 --build-arg COMMIT=a85098e87798-dirty --build-arg DATE=2026-09-15T17:34:53Z --tag iotwq/china-api:latest --tag iotwq/china-api:0.2.5-20260916-013453 --metadata-file /tmp/sub2api-image-unread.HISCNW/metadata.json --progress=plain --push .`：退出 0。
- 构建内翻译完整性 3 项、前端类型检查/生产构建、双架构 Go 编译通过；沿用源码修改阶段的徽标 34 项与侧栏/群聊 51 项回归证据，不重复没有变化的测试。前端保留既有 Browserslist/大 chunk 提示。
- 两个标签分别执行 `docker buildx imagetools inspect`，远端 index 均与 metadata 一致：`sha256:a6a6138b13d3cc270d4028665b8c79f1b65f5d5c14fbb0da8c51d5b67c5bfd67`。
- amd64 manifest `sha256:38d9617906ff14271b3cf27a6adaa12c1315020228bc90fc118052eea1d354e0`、arm64 manifest `sha256:6b981166925e7aefd1bab9306364ababc652d26b4ead477d97c742d48b68d60b` 使用 `docker pull --platform linux/<架构> iotwq/china-api@<manifest>` 拉取；随后执行 `docker run --rm --platform linux/<架构> --network none --read-only --entrypoint /app/sub2api iotwq/china-api@<manifest> --version`，两个架构均退出 0，版本信息一致。AMD64 拉取遇到认证端点和 CDN EOF 后，第二次重试成功。
- `git diff --check`、`git diff --cached --check` 通过，无未解决合并项。发布日志与元数据位于 `/tmp/sub2api-image-unread.HISCNW/`；版本检查未连接生产数据库或验证真实用户请求。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加发布标签、双架构摘要、验证结果、部署影响及回滚方式。
- `progress.md`：追加本轮发布结果和验证记录，保留原有历史。
- 相比上一固定版本，本次新增业务改动仅涉及前端，无新增数据库迁移；本轮未部署、重启服务、修改生产数据或推送 Git 提交。部署新镜像后需刷新页面。
- 回滚点：`iotwq/china-api:0.2.5-20260916-000512`（index `sha256:480c7859fa16229ea44dccf8d29ddd9d8103e24a35f459142116bb660654262c`）。恢复 latest 标签可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:480c7859fa16229ea44dccf8d29ddd9d8103e24a35f459142116bb660654262c`；本轮未执行回滚。

## 2026-09-17 - Task: OpenAI OAuth Codex 线路独立并发与协议保护

### What was done

- 经用户明确授权，对照 `/Users/wangqiang/Project/grok_build/sub2新站` 选择性接入账号级保护，不替换当前较新的 WS 读取循环、执行隔离、账号调度或计费实现。
- 为 OpenAI OAuth / Setup Token 非影子账号提供默认关闭的保护开关；Redis 原子限制在途数、滑动 RPM 与突发额度，支持有界排队、429/5xx 自适应降速和逐步恢复。排队耗尽走既有切换账号路径，本地满载不记为账号健康故障。
- HTTP 按响应体生命周期占用额度，WS 按生成轮次占用；保留客户端断开后收集上游计费用量的并发名额，真实终止或退出后释放。原账号并发上限仍然生效。
- 保护开启时默认 device 身份，HTTP/WSS 使用持久种子和一致的头/体标识；保留不同用户与会话的隔离。可选共享 uTLS 传输，支持 HTTP/HTTPS CONNECT 和 SOCKS，验证证书、按账号/代理/模板隔离池，不改变默认传输。
- 新增请求语义完整性观察/阻止模式；覆盖上下文、推理、工具、模型及连续性字段，接受已知等价协议转换，不记录用户提示词或凭据。
- 新入口位于账号管理的编辑弹窗；补中英文标签、使用边界与回退说明。没有同步参考站的账号健康隔离状态机、强制固定并发或大范围策略界面。

### Testing

- `cd backend && go generate ./cmd/server`：Wire 生成通过，运行构造器接入 Redis 控制器；未新增依赖或数据库迁移。
- `go test -mod=readonly -tags=unit ./internal/service ./internal/handler ./internal/repository ./internal/pkg/tlsfingerprint ./internal/service/openai_ws_v2 -run 'Test.*(OpenAI|Codex|AccountTraffic|Mode1|HTTPUpstream|LocalTLS|TLSFingerprint|AccountTest|UpstreamFailover|FingerprintSeed|BulkUpdateAccounts|UpdateAccountExtra)' -count=1 -timeout=5m`：5 个包全部通过。覆盖既有 OAuth 重试、账号测试、指纹、WS 连接池/读取循环、透传及相关处理器。
- `go test -mod=readonly -race -tags=unit ./internal/service ./internal/repository ./internal/pkg/tlsfingerprint -run 'Test.*(AccountTraffic|CodexProtection|Mode1|LocalTLS|OpenAIWSConnPool)' -count=1 -timeout=4m`：3 个包全部通过。新增测试覆盖 Redis 多客户端并发、RPM/突发、降速/恢复、取消/排队边界、完整流释放、WS 失败不计成功、断开后 drain 名额保留、不同会话隔离、实际 HTTP/WSS ClientHello 与证书拒绝、完整 WS 透传入口。
- `go build -mod=readonly -tags embed -o /tmp/sub2api-codex-protection.rc4XgL/sub2api ./cmd/server`：通过，未连接生产数据库。
- `cd frontend && pnpm exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts src/i18n/__tests__/localeKeyCompleteness.spec.ts`：2 个文件、72 项通过；实际编辑弹窗覆盖 OAuth/Setup Token 保存与关闭、显式指纹 off、保留已有参数、其他账号无新增入口。
- `pnpm exec vue-tsc --noEmit`、本轮前端文件 ESLint、`pnpm run build`：通过。保留既有 Browserslist / 大 chunk 警告；构建产物为被忽略的 web/dist，没有手改产物或依赖锁文件。
- Playwright 检查新设置组件 1200×900 与 390×844：无横向溢出/JS 错误，开关隐藏和恢复、数值输入通过；已目视检查截图。证据 `/tmp/sub2api-codex-protection.rc4XgL/desktop.png`、`mobile.png`；隔离预览 http://127.0.0.1:5175/__protection_preview.html 不调用生产账号接口。
- `git diff --check`、`git diff --cached --check` 通过；`git apply -R --check /tmp/sub2api-codex-protection.rc4XgL/task.patch` 通过，回退补丁只包含本轮 45 个源码/文档文件，不含已有定制和进度历史。
- 未做官方线路的真实负载/回答质量 A/B 验证，因此不声称已规避官方限制或证明防止“降智”。未构建或推送 Docker 镜像，未部署、修改余额或生产账号。

### Notes

- `backend/cmd/server/wire_gen.go`：生成账号流控的运行依赖注入。
- `backend/internal/repository/wire.go`：注册受控 HTTP 上游及 Redis 缓存。
- `backend/internal/repository/http_upstream.go`：保留原发送器并添加独立 Codex TLS 池。
- `backend/internal/repository/scheduler_cache.go`：在账号快照中保留保护配置。
- `backend/internal/repository/account_traffic_cache.go`：新增 Redis 原子准入、租约、降速与恢复状态。
- `backend/internal/repository/account_traffic_cache_test.go`：验证跨实例预算、令牌桶及自适应恢复。
- `backend/internal/repository/account_traffic_regression_test.go`：验证降 RPM 等待时间及旧策略不能绕过限制。
- `backend/internal/repository/codex_protection_tls_test.go`：验证 HTTP TLS 握手与池隔离。
- `backend/internal/service/admin_account.go`：创建、编辑、增量和批量保存时校验新配置。
- `backend/internal/service/gateway_service.go`：本地保护拒绝不污染账号健康指标。
- `backend/internal/service/account_traffic_policy.go`：定义 OAuth 独立配置及请求级限额错误。
- `backend/internal/service/account_traffic_service.go`：实现有界等待、流生命周期和原生 WS drain 租约。
- `backend/internal/service/account_traffic_ws.go`：实现透传 WS 每轮计数与释放。
- `backend/internal/service/account_traffic_events.go`：分类流内终止结果及限流/服务器错误。
- `backend/internal/service/account_traffic_service_test.go`：验证关闭、拒绝、流读取与 WS 轮次行为。
- `backend/internal/service/account_traffic_regression_test.go`：验证错误分类与其他平台隔离。
- `backend/internal/service/account_request_integrity.go`：按账号开关选择观察或阻止语义损失。
- `backend/internal/service/openai_mode1_integrity.go`：保存请求快照并比较关键语义字段。
- `backend/internal/service/openai_mode1_semantics.go`：规范化已知 Codex 等价转换后比较。
- `backend/internal/service/openai_mode1_integrity_test.go`：验证可接受兼容变换及真实信息损失。
- `backend/internal/service/openai_oauth_protection.go`：限定 Responses 路径及可选 TLS 模板。
- `backend/internal/service/openai_oauth_protection_test.go`：验证路由边界、排队、重试计数、身份及 drain。
- `backend/internal/service/openai_codex_fingerprint.go`：保护下默认 device、对齐字段别名与 WS 元数据。
- `backend/internal/service/openai_plugin_transport.go`：真实请求与账号测试共用保护发送入口。
- `backend/internal/service/openai_upstream_transport_error.go`：本地限额保留请求级 failover 语义。
- `backend/internal/service/openai_gateway_forward.go`：HTTP 请求完整性验证及 enforce 下禁止有损密文重试。
- `backend/internal/service/openai_gateway_passthrough.go`：透传构造时验证请求语义。
- `backend/internal/service/openai_ws_client.go`：可选 WSS TLS 模板，保留现有 ping/reader 行为。
- `backend/internal/service/openai_ws_tls.go`：按账号/目标/代理/模板缓存 WSS TLS 传输。
- `backend/internal/service/openai_ws_pool.go`：保护下按会话与模板匹配，克隆模板快照。
- `backend/internal/service/openai_ws_forwarder_v2.go`：HTTP→WS 轮次准入、结果观察与 drain 占用。
- `backend/internal/service/openai_ws_forwarder_ingress.go`：原生 WS 池化路径的身份、完整性与逐轮准入。
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`：WS 透传路径接入保护、完整性与可选 TLS。
- `backend/internal/service/codex_protection_ws_tls_test.go`：验证真实 WSS 握手、透传入口和设备一致性。
- `backend/internal/pkg/tlsfingerprint/profile.go`：新增不可变克隆及模板内容摘要。
- `backend/internal/pkg/tlsfingerprint/builtin_profiles.go`：新增可选 Node.js 兼容模板。
- `backend/internal/pkg/tlsfingerprint/transport.go`：提供证书验证的 uTLS 与代理隧道传输。
- `backend/internal/pkg/tlsfingerprint/transport_local_test.go`：验证直连、代理、取消与错误拒绝。
- `frontend/src/components/account/EditAccountModal.vue`：添加独立保护配置的读取和保存入口。
- `frontend/src/components/account/OpenAIOAuthProtectionFields.vue`：新增响应式设置控件。
- `frontend/src/components/account/openAIOAuthProtection.ts`：定义默认值并保留已有高级配置。
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`：新增真实编辑弹窗保护交互回归。
- `frontend/src/i18n/locales/zh/admin/accounts.ts`：新增中文设置标签。
- `frontend/src/i18n/locales/en/admin/accounts.ts`：新增英文设置标签。
- `docs/OPENAI_OAUTH_PROTECTION.md`：记录启用方式、默认值、边界和验证命令。
- `progress.md`：仅追加本轮实现、测试与回退证据。
- 回退：在仓库根目录先执行 `git apply -R --check /tmp/sub2api-codex-protection.rc4XgL/task.patch`，检查通过后执行 `git apply -R /tmp/sub2api-codex-protection.rc4XgL/task.patch`。该补丁恢复施工前原文件并移除本轮新增文件，保留本日志与此前未提交定制；遇到重叠编辑时检查会失败，应人工合并，不强行覆盖。备份 `before.tar.gz` 与 `additional-before.tar.gz` 同目录可恢复，临时目录需在长期保留前另行归档。未执行回退。
- 发布后默认仍关闭；最小启用动作是仅选一个 OAuth 账号灰度开启。影子账号与重复导入同一官方凭据的不同本地账号记录不共享本轮预算；所有可用账号均饱和时仍可能返回 429/503。关闭保护后需恢复原指纹选择并重连 WS 才能完整回到原配置。

## 2026-09-17 - Task: 修复 OAuth HTTP 保护等待取消后仍发送请求

### What was done

- 按用户指定顺序完成第一项：普通 Responses HTTP 和透传路径把客户端取消信号带入保护准入，等待取消及取得名额同时取消时均停止发送并释放已取得名额。
- 准入完成后停止关联客户端取消，保留原有独立上游生命周期和用量收集；未开启保护、API Key 及独立媒体路径不改变行为。未增加 FIFO 或调整并发默认值。

### Testing

- 新增真实 `Forward` 回归先运行失败：普通/透传、等待中/刚取得名额四种场景均错误发送 1 次；修复后全部通过。
- `cd backend && go test -mod=readonly -race -tags=unit ./internal/service ./internal/handler -run 'Test.*(AccountTraffic|CodexProtectionHTTP|CodexProtectionBounded|CodexProtectionNativeDrain|OpenAIGatewayHandlerResponses_Failover|OpenAIGatewayService_.*Passthrough|HandleOpenAIUpstreamTransportError)' -count=1 -timeout=3m`：两个包通过，覆盖发出后断线仍收集用量、名额仅释放一次、提前取消、关闭保护/API Key/图片路径隔离、既有错误切换及取消回归。
- `git diff --check` 通过。测试仅使用本地模拟，不调用官方账号或生产数据。

### Notes

- `backend/internal/service/account_traffic_service.go`：准入阶段关联客户端取消，准入后恢复独立上游生命周期。
- `backend/internal/service/openai_gateway_forward.go`：普通 HTTP 转发传递原客户端上下文。
- `backend/internal/service/openai_gateway_passthrough.go`：透传 HTTP 转发传递原客户端上下文。
- `backend/internal/service/account_traffic_http_cancel_test.go`：新增真实转发取消、并发释放、用量收集和范围隔离回归。
- `docs/OPENAI_OAUTH_PROTECTION.md`：说明等待取消与已发送请求的生命周期边界。
- `progress.md`：仅追加第一项结果与验证。
- 回滚：确认没有后续重叠改动后，在仓库根目录执行 `tar -xzf /tmp/sub2api-protection-fixes.CLrxjp/http-before.tar.gz -C .`，再执行 `mv backend/internal/service/account_traffic_http_cancel_test.go /tmp/sub2api-protection-fixes.CLrxjp/account_traffic_http_cancel_test.go.rolled-back`；保留进度历史。备份只包含以上四个既有源码/文档文件，保留本轮之前的定制；若后续两项已完成，应先按日志逆序回滚。本轮未执行回滚、构建镜像或部署。

## 2026-09-17 - Task: 修复 OAuth 保护下 WS 请求追踪 ID 导致续接失败

### What was done

- 按顺序完成第二项：开启保护时，WS 连接匹配不再把每请求变化的追踪 ID 当作稳定会话身份，指定连接的后续请求可以继续复用。
- 保留设备、session、thread、conversation、window、beta features 及 TLS 传输隔离；不放宽其他会话的复用，也不改变未开启保护的原有匹配规则。

### Testing

- 新增指定连接的真实连接池准入回归先失败：保护开启的 off/device/session/full 四种指纹模式均因仅更换追踪 ID 报 preferred connection unavailable；修复后通过。
- `cd backend && go test -mod=readonly -race -tags=unit ./internal/service -run 'Test.*(CodexProtectionWS|CodexProtectionDevice|OpenAIWSConnPool|NormalizeOpenAIWS|CodexFingerprint)' -count=1 -timeout=3m`：通过，覆盖七类稳定身份变化仍隔离、关闭保护时的四种旧模式、原连接池并发与可选 TLS 握手。
- `git diff --check` 通过。本轮使用本地连接池和回环测试，未调用官方线路。

### Notes

- `backend/internal/service/openai_ws_pool.go`：仅保护账号忽略请求级追踪 ID 的连接匹配。
- `backend/internal/service/openai_ws_protection_compatibility_test.go`：新增指定续接、稳定身份隔离和旧模式保持测试。
- `docs/OPENAI_OAUTH_PROTECTION.md`：补充追踪信息与会话身份的区别及复用边界。
- `progress.md`：追加第二项验证与回退记录。
- 回滚：在没有后续重叠编辑时，在仓库根目录执行 `tar -xzf /tmp/sub2api-protection-fixes.CLrxjp/ws-compat-before.tar.gz -C .`，再执行 `mv backend/internal/service/openai_ws_protection_compatibility_test.go /tmp/sub2api-protection-fixes.CLrxjp/openai_ws_protection_compatibility_test.go.rolled-back`；保留进度历史。该备份包含第一项完成后的文档，不覆盖第一项代码修复；若第三项已完成需先逆序回滚第三项。本轮未执行回滚或部署。

## 2026-09-17 - Task: 修复 OAuth 保护下原生 WS 握手失败未纳入自适应统计

### What was done

- 按用户指定顺序完成第三项：原生 WS 池化入口和 WS 透传入口在生成名额建立前，针对真实上游 429/500–599 握手失败写入同一自适应失败窗口；HTTP→WS 已取得名额的握手失败继续由名额结束逻辑记录，避免重复统计。
- 观察模式不占用或释放生成租约，不增加 accepted、completed 或 duration；网络超时、鉴权失败、客户端取消和本地保护 503 不会降低建议并发。Redis 侧保留修订号与签名校验，旧策略不能覆盖新策略。
- 未增加 FIFO、未修改默认并发、未改变 API Key/非 OpenAI OAuth/媒体路径。

### Testing

- `cd backend && go test -mod=readonly -race -tags=unit ./internal/service ./internal/repository -run 'Test.*(CodexProtectionWS|AccountTrafficStandaloneHandshake|AccountTrafficAdaptive|AccountTrafficStalePolicy)' -count=1 -timeout=5m`：通过。
- `cd backend && go test -mod=readonly -tags=unit ./internal/service ./internal/handler ./internal/repository ./internal/pkg/tlsfingerprint ./internal/service/openai_ws_v2 -run 'Test.*(OpenAI|Codex|AccountTraffic|Mode1|HTTPUpstream|LocalTLS|TLSFingerprint|AccountTest|UpstreamFailover|FingerprintSeed|BulkUpdateAccounts|UpdateAccountExtra)' -count=1 -timeout=5m`：通过。
- `git diff --check`：通过。测试使用本地 Redis 仿真、回环连接和错误桩，未调用官方账号或生产数据。

### Notes

- `backend/internal/service/account_traffic_service.go`：补充 WS 握手错误分类和独立失败观察入口。
- `backend/internal/repository/account_traffic_cache.go`：复用现有自适应 Lua 逻辑，增加无租约握手失败观察模式。
- `backend/internal/service/openai_ws_forwarder_ingress.go`：原生池化 WS 握手失败接入观察。
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`：原生透传 WS 握手失败接入观察。
- `backend/internal/service/account_traffic_ws_failure_test.go`：新增真实错误分类、取消和账号范围回归。
- `backend/internal/repository/account_traffic_cache_test.go`：新增无租约失败统计、租约保留和旧策略隔离回归。
- `docs/OPENAI_OAUTH_PROTECTION.md`：说明三条 WS 路径的握手失败统计边界。
- `progress.md`：追加本轮实现、验证与回滚记录。
- 回滚：确认没有后续重叠编辑后，在仓库根目录执行 `tar -xzf /tmp/sub2api-protection-fixes.CLrxjp/ws-failure-before.tar.gz -C .`，再移除 `backend/internal/service/account_traffic_ws_failure_test.go`，并用 `git apply -R` 或人工反向移除 `backend/internal/repository/account_traffic_cache_test.go` 本轮新增测试；本轮完成态备份为 `/tmp/sub2api-protection-fixes.CLrxjp/ws-failure-after.tar.gz`。不要使用重置工作区的方式覆盖其他本地定制。本轮未部署、未构建或推送镜像。

## 2026-09-18 - Task: 同步原项目最新 52 个提交并验证本地定制兼容

### What was done

- 将原项目 `Wei-Shaw/sub2api` 的 `upstream/main` 从 `881f3202694c6bc932446931a30c27d9675178b9` 合并至 `efe9aab1e4ec89a42ba45e8dac20e882c5409a6a`，共 52 个提交；本地合并提交为 `03f6b278af9e917f96e075b8e55f4bb42bb4ee79`。版本维持 0.2.5，无新增数据库迁移。
- 同步 OAuth 刷新、响应绑定、Gemini/Antigravity 模型发现、DeepSeek 媒体工具结果、Chat 角色兼容、分组统计查询、兑换分页、前端交互与依赖修复；保留本地媒体、计费、群聊、监控及 OAuth 指纹/TLS/并发等定制。
- 基线合并无冲突；恢复本地定制时解决 Gemini 列表与监控测试两处冲突。补充 OpenAI 分组 Gemini-native 能力隔离的真实处理器回归；避免新模型发现逻辑调用不相关平台。
- 修复验证中发现的既有取消测试偶发竞争：模拟 HTTP 服务取消客户端后等待其断开，防止先返回空 200；不改变监控生产行为或测试断言。
- 485 个既有脏文件逐一核对未丢失、暂存状态保持；内容变动限于上游交集、上述测试及本轮文档追加。保留同步前分支、工作区 stash 与快照，不覆盖或整体提交用户既有定制。

### Testing

- 前端 `pnpm test:run`：316 个测试文件、2377 项测试通过；`pnpm lint:check`、`pnpm build` 通过，覆盖翻译检查、类型检查及生产构建；保留已有 Browserslist/大 chunk 提示。
- 后端 `go test -mod=readonly -tags=unit ./... -count=1 -timeout=6m`：首次仅 service 包的 `TestMonitorIntelligenceHealthIsolation/candy_canceled` 偶发失败，其他包全部通过；修正模拟服务器后，`go test -mod=readonly -tags=unit ./internal/service -count=1 -timeout=6m` 全包通过。
- `go test -mod=readonly -race -tags=unit ./internal/service -run '^TestMonitorIntelligenceHealthIsolation$' -count=100 -timeout=2m` 通过。
- `go test -mod=readonly -race -tags=unit ./internal/service ./internal/repository ./internal/pkg/tlsfingerprint -run 'Test.*(CodexProtection|AccountTraffic|Mode1|LocalTLS|HTTPUpstream|BindHTTPResponseAccount)' -count=1 -timeout=5m` 通过。
- `go test -mod=readonly -tags=integration ./internal/repository -run 'TestAccountRepoSuite/TestListOAuthRefreshCandidatePage_GrokCursorAndExclusions|TestRedeemCodeRepoSuite/TestHistoryPaginationIsolationAndOrdering|TestGroupUsageSummary|TestGroupUsageRollupSyncRebuildsAfterTimezoneChange' -count=1 -timeout=5m -v` 通过，使用临时 PostgreSQL 18.1 和 Redis 8.4，未访问生产数据库。
- `go build -mod=readonly -tags=embed -o /tmp/sub2api-sync-20260918.tGFjml/server ./cmd/server` 通过；生成程序 `--version` 退出 0，版本 0.2.5。上述 Go 命令在 backend 目录执行。
- `git diff --check`、`git diff --cached --check`、上游祖先关系及无未解决冲突检查通过；快照内容与暂存状态核对通过。证据目录：`/tmp/sub2api-sync-20260918.tGFjml/`。

### Notes

- `.gitignore`：同步新增 Antigravity 排查文档的跟踪白名单。
- `backend/go.mod`：同步 grpc 1.83.2 与关联依赖升级，保留本地媒体依赖。
- `backend/go.sum`：同步 grpc 1.83.2 与关联依赖升级，保留本地媒体依赖。
- `backend/internal/handler/gateway_model_allowlist_listing_test.go`：同步 Gemini 原生模型列表中的 Antigravity 实际映射发现及验证。
- `backend/internal/handler/gemini_mixed_models_test.go`：同步 Gemini 原生模型列表中的 Antigravity 实际映射发现及验证。
- `backend/internal/handler/gemini_v1beta_handler.go`：合并上游模型发现，同时保留 OpenAI Gemini-native 能力边界及专用账号选择。
- `backend/internal/handler/gemini_v1beta_handler_test.go`：同步 Gemini 原生模型列表中的 Antigravity 实际映射发现及验证。
- `backend/internal/handler/redeem_handler.go`：同步用户兑换历史分页、稳定排序及对应验证。
- `backend/internal/handler/redeem_handler_test.go`：同步用户兑换历史分页、稳定排序及对应验证。
- `backend/internal/pkg/antigravity/attribution_test.go`：同步 Antigravity Claude 归属元数据清理及对应验证。
- `backend/internal/pkg/antigravity/request_transformer.go`：同步 Antigravity Claude 归属元数据清理及对应验证。
- `backend/internal/pkg/apicompat/responses_tool_output_media.go`：同步 DeepSeek 工具输出中的媒体兼容，并保持并行工具结果连续。
- `backend/internal/pkg/apicompat/responses_tool_output_media_test.go`：同步 DeepSeek 工具输出中的媒体兼容，并保持并行工具结果连续。
- `backend/internal/repository/account_repo.go`：同步暂停调度账号仍参与 OAuth token 刷新及对应验证。
- `backend/internal/repository/account_repo_integration_test.go`：同步暂停调度账号仍参与 OAuth token 刷新及对应验证。
- `backend/internal/repository/account_repo_temp_unsched_test.go`：同步暂停调度账号仍参与 OAuth token 刷新及对应验证。
- `backend/internal/repository/custom_group_usage_rollup_repo.go`：同步分组用量汇总尾部查询边界优化及跨日、时区回归。
- `backend/internal/repository/redeem_code_repo.go`：同步用户兑换历史分页、稳定排序及对应验证。
- `backend/internal/repository/redeem_code_repo_sort_integration_test.go`：同步用户兑换历史分页、稳定排序及对应验证。
- `backend/internal/repository/usage_log_repo_group_summary_test.go`：同步分组用量汇总尾部查询边界优化及跨日、时区回归。
- `backend/internal/service/cn_providers_test.go`：同步 DeepSeek Responses 工具结果图片转换及对应验证。
- `backend/internal/service/gemini_messages_compat_service.go`：同步 Gemini 原生模型列表中的 Antigravity 实际映射发现及验证。
- `backend/internal/service/openai_chat_roles.go`：同步严格 Chat 上游的 developer 角色规范化及对应验证。
- `backend/internal/service/openai_chat_roles_test.go`：同步严格 Chat 上游的 developer 角色规范化及对应验证。
- `backend/internal/service/openai_codex_models_service.go`：同步 Codex 模型清单键名校验兼容修正。
- `backend/internal/service/openai_codex_models_service_test.go`：同步 Codex 模型清单键名校验兼容修正。
- `backend/internal/service/openai_gateway_chat_completions_raw.go`：同步严格 Chat 上游的 developer 角色规范化及对应验证。
- `backend/internal/service/openai_gateway_request_body.go`：同步 DeepSeek Responses 工具结果图片转换及对应验证。
- `backend/internal/service/openai_gateway_response_handling.go`：同步客户端取消后仍在有界独立上下文保存响应账号绑定。
- `backend/internal/service/openai_gateway_service_test.go`：同步客户端取消后仍在有界独立上下文保存响应账号绑定。
- `backend/internal/service/redeem_service.go`：同步用户兑换历史分页、稳定排序及对应验证。
- `backend/internal/service/token_refresh_service_candidates_test.go`：同步暂停调度账号仍参与 OAuth token 刷新及对应验证。
- `docs/ANTIGRAVITY_ATTRIBUTION_429.md`：同步上游归属元数据兼容修复的排查说明。
- `frontend/src/api/__tests__/redeem.spec.ts`：同步用户兑换历史分页、稳定排序及对应验证。
- `frontend/src/api/redeem.ts`：同步用户兑换历史分页、稳定排序及对应验证。
- `frontend/src/components/admin/channel/ModelTagInput.vue`：同步空模型输入框允许 Tab 移出及键盘回归。
- `frontend/src/components/admin/channel/__tests__/ModelTagInput.keyboard.spec.ts`：同步空模型输入框允许 Tab 移出及键盘回归。
- `frontend/src/components/admin/payment/AdminRefundDialog.vue`：同步退款余额提醒使用本次申请金额及对应验证。
- `frontend/src/components/admin/payment/__tests__/AdminRefundDialog.balance.spec.ts`：同步退款余额提醒使用本次申请金额及对应验证。
- `frontend/src/components/admin/user/UserPlatformQuotaModal.vue`：同步平台额度禁止保存负数及对应验证。
- `frontend/src/components/admin/user/__tests__/UserPlatformQuotaModal.spec.ts`：同步平台额度禁止保存负数及对应验证。
- `frontend/src/components/common/BaseDialog.vue`：同步多弹窗实例标题 ID 唯一性及对应验证。
- `frontend/src/components/common/Pagination.vue`：同步分页数字跳转输入兼容及对应验证。
- `frontend/src/components/common/ProxySelector.vue`：同步批量代理测试复用在途去重及对应验证。
- `frontend/src/components/common/__tests__/BaseDialog.ids.spec.ts`：同步多弹窗实例标题 ID 唯一性及对应验证。
- `frontend/src/components/common/__tests__/Pagination.jump.spec.ts`：同步分页数字跳转输入兼容及对应验证。
- `frontend/src/components/common/__tests__/ProxySelector.testing.spec.ts`：同步批量代理测试复用在途去重及对应验证。
- `frontend/src/components/payment/AmountInput.vue`：同步无效充值金额输入恢复已接受文本及对应验证。
- `frontend/src/components/payment/__tests__/AmountInput.spec.ts`：同步无效充值金额输入恢复已接受文本及对应验证。
- `frontend/src/components/user/profile/TotpDisableDialog.vue`：同步 TOTP 禁用时显示规范化接口错误及对应验证。
- `frontend/src/components/user/profile/TotpSetupModal.vue`：同步 TOTP 数字输入与状态一致及对应验证。
- `frontend/src/components/user/profile/__tests__/Totp.errors.spec.ts`：同步 TOTP 禁用时显示规范化接口错误及对应验证。
- `frontend/src/components/user/profile/__tests__/TotpSetupModal.inputs.spec.ts`：同步 TOTP 数字输入与状态一致及对应验证。
- `frontend/src/composables/__tests__/useClipboard.spec.ts`：同步剪贴板降级复制抛错反馈及对应验证。
- `frontend/src/composables/useClipboard.ts`：同步剪贴板降级复制抛错反馈及对应验证。
- `frontend/src/i18n/locales/en/dashboard.ts`：同步用户兑换历史分页、稳定排序及对应验证。
- `frontend/src/i18n/locales/zh/dashboard.ts`：同步用户兑换历史分页、稳定排序及对应验证。
- `frontend/src/stores/__tests__/announcements.markAll.spec.ts`：同步公告批量已读部分失败时保留成功项及对应验证。
- `frontend/src/stores/__tests__/payment.config.spec.ts`：同步支付配置等待已有请求完成及对应验证。
- `frontend/src/stores/__tests__/subscriptions.clear.spec.ts`：同步订阅状态清理时复位加载标记及对应验证。
- `frontend/src/stores/announcements.ts`：同步公告批量已读部分失败时保留成功项及对应验证。
- `frontend/src/stores/payment.ts`：同步支付配置等待已有请求完成及对应验证。
- `frontend/src/stores/subscriptions.ts`：同步订阅状态清理时复位加载标记及对应验证。
- `frontend/src/views/admin/__tests__/ChannelMonitorView.grok.spec.ts`：解决恢复冲突，保留本地 PROVIDERS.length 动态断言，内容与同步前一致。
- `frontend/src/views/admin/__tests__/GroupsView.codexManifest.spec.ts`：合并上游 Pinia 初始化与本地既有测试桩。
- `frontend/src/views/auth/RegisterView.vue`：同步注册页推广码加载时避免闪烁及对应验证。
- `frontend/src/views/auth/__tests__/RegisterView.spec.ts`：同步注册页推广码加载时避免闪烁及对应验证。
- `frontend/src/views/user/RedeemView.vue`：同步用户兑换历史分页、稳定排序及对应验证。
- `frontend/src/views/user/UserOrdersView.vue`：同步切换订单状态时重置分页及对应验证。
- `frontend/src/views/user/__tests__/RedeemView.spec.ts`：同步用户兑换历史分页、稳定排序及对应验证。
- `frontend/src/views/user/__tests__/UserOrdersView.filters.spec.ts`：同步切换订单状态时重置分页及对应验证。
- `backend/internal/handler/gemini_openai_models_test.go`：新增 OpenAI Gemini-native 能力开启/关闭及不调用 Antigravity 仓库的回归。
- `backend/internal/service/channel_monitor_intelligence_test.go`：仅稳定取消测试的模拟服务生命周期，消除空 200 与取消的竞争。
- `docs/UPSTREAM_SYNC.md`：追加本次同步范围、兼容处理、验证、发布边界与恢复命令。
- `progress.md`：仅追加本轮记录及逐文件说明。
- 回滚点：`codex/pre-upstream-sync-20260918`（`a85098e8779876b7a5334ff94099e94564e46198`）与 stash `db7153f66d990e693f0d1ee6bc42e7a770dfa936`。在仓库根目录依次执行 `git worktree add -b codex/recover-pre-sync-20260918 /tmp/sub2api-pre-sync-20260918 codex/pre-upstream-sync-20260918`、`git -C /tmp/sub2api-pre-sync-20260918 stash apply --index db7153f66d990e693f0d1ee6bc42e7a770dfa936`、`git -C /tmp/sub2api-pre-sync-20260918 add -f -N docs/upstream-sync-20260915-files.md`，可在独立目录恢复同步前代码和定制；原有被忽略文档仍留在当前目录，勿删除备份或用强制重置覆盖当前工作区。
- 本轮未执行回滚、未推送 GitHub、未制作或推送 Docker 镜像、未部署服务。

## 2026-09-18 - Task: 构建并推送最新双架构 china-api 镜像

### What was done

- 基于完整当前工作区发布 `iotwq/china-api:latest` 与 `iotwq/china-api:0.2.5-20260918-211628`，支持 `linux/amd64` 和 `linux/arm64`，包含本日上游同步及全部既有定制。
- 使用版本 `0.2.5`、commit `03f6b278af9e-dirty`、构建时间 `2026-09-18T13:16:28Z`；两个标签均指向 OCI index `sha256:d1733b09ba45e6eecdb17900a35816ad9eb2b2111984fd2dc2556b0b110586d3`。
- 记录两个架构的发布 manifest：amd64 `sha256:b8d8207fd17a9bfe062c2f14f5c93cad34caea1cc25f3135bbac2ea240166519`，arm64 `sha256:f1cdbc191e3c099b57d2f56813d1b066083df273897bd762da3a64e7e862c7c0`。本轮未改动业务代码、部署服务或访问生产数据库。

### Testing

- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --file Dockerfile --build-arg VERSION=0.2.5 --build-arg COMMIT=03f6b278af9e-dirty --build-arg DATE=2026-09-18T13:16:28Z --tag iotwq/china-api:latest --tag iotwq/china-api:0.2.5-20260918-211628 --metadata-file /tmp/sub2api-image-sync-20260918.qQjjcI/metadata.json --progress plain --push .`：退出 0。
- 构建过程的翻译完整性 3 项、前端类型检查/生产构建、两个架构 Go 编译均通过；保留已有 Browserslist/大 chunk 提示。本轮沿用前一任务完整源码回归结果，不重复无变化的全套测试。
- `docker buildx imagetools inspect` 核验 latest 与固定标签：远端摘要均与 metadata 一致，含两个目标架构及 provenance 附件。
- 两个架构均按上述 manifest 从 Docker Hub 拉取，并执行 `docker run --rm --pull never --platform linux/<arch> --network none --read-only --user 1000:1000 iotwq/china-api@<manifest> --version`，均退出 0，版本、commit、构建时间一致。AMD64 首次拉取遇到认证端点 EOF，重试通过。
- 构建、推送、远端清单、拉取及版本日志位于 `/tmp/sub2api-image-sync-20260918.qQjjcI/`；`git diff --check`、`git diff --cached --check` 通过。未运行线上端到端请求。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加固定标签、双架构 digest、验证结果及回滚方式。
- `progress.md`：仅追加本轮发布、验证与回滚记录。
- 回滚：发布前 latest 为 `sha256:933be4db3e0454c074136d89eacbce4b46292356a723f80f85a8041a68cba9bf`；可直接部署 `iotwq/china-api@sha256:933be4db3e0454c074136d89eacbce4b46292356a723f80f85a8041a68cba9bf`，或执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:933be4db3e0454c074136d89eacbce4b46292356a723f80f85a8041a68cba9bf` 恢复标签。本轮未执行回滚，标签恢复不会自动替换已运行容器。

## 2026-09-18 - Task: 合并指定分支的 Codex 票据功能并核对完整 0.2.6 功能截图

### What was done

- 从指定 `Tinghecui/sub2api` 票据分支合并 6 个缺失提交至 `3c2f05c957b4b93866318ec8695fc5a28fff70eb`，产生本地合并提交 `86f0e36653d25808ec2e6e150d43b42050d0eb88`；本轮导入 61 个上游变更文件，版本按要求保持 `0.2.5`，无新增迁移。
- 补齐票据采集与 1 小时缓存、目标模型注入、默认关闭的后台开关、独立代理配置/校验/去敏、账号摘要以及账号保存、compact 调度和生命周期修复。核对完整截图的其他条目已包含在前次上游同步中。
- 解决恢复本地定制时的 3 处冲突，同时保留 OAuth 策略校验、视频补偿状态和账号保护界面。原 489 个修改文件无丢失，原暂存区区分恢复。
- 发现采集直接请求会绕过本地并发保护，以回归证实后复用现有并发/RPM 准入，保留专用短连接/代理和业务 TLS 选择。补充普通 HTTP、透传、WS 与四种指纹模式联合回归。
- 更新功能文档、完整截图逐项对应、开启前提及独立工作区回滚方式。本轮未开启线上开关、未构建镜像、未推送或部署。

### Testing

- `cd frontend && pnpm test:run`：316 个测试文件、2381 项测试通过；`pnpm lint:check`、`pnpm build`（含翻译完整性和类型检查）通过，保留已有 Browserslist/大 chunk 提示。
- `cd backend && go test -mod=readonly -tags=unit ./... -count=1 -timeout=6m`：全量通过；本轮后续只增加采集并发适配及联合回归，按以下定向 race 验证最终改动。
- 新增 `TestCodexTicketHarvestRespectsExistingProtection` 先失败，证明满载采集仍发出且未计入保护预算；4 行准入适配后通过。关闭保护不触及预算、真实 429 统计、仅读头释放以及三条传输路径身份一致均通过。
- `go test -mod=readonly -race -tags=unit ./internal/service ./internal/repository ./internal/handler ./internal/pkg/tlsfingerprint -run 'Test.*(CodexTicket|OpenAICodexTicket|CodexProtection|AccountTraffic|TurnState|Mode1|HTTPUpstream|LocalTLS)' -count=1 -timeout=5m`：最终适配后通过。另已通过 admin 票据设置 race 回归。
- `go test -mod=readonly -tags=integration ./internal/repository -run 'TestAccountRepoSuite/(TestUpdate|TestBulkUpdate)|TestLockAndMergeAccountExtra|TestCodexTicketExtra' -count=1 -timeout=5m -v`：临时数据库账号更新、Extra 合并、快照同步、批量更新及票据保留相关验证通过，未使用生产数据库。
- `go generate ./cmd/server`、`go build -mod=readonly -tags=embed -o /tmp/sub2api-codex-ticket-20260918.B6W6Uh/server ./cmd/server` 通过；生成程序 `--version` 退出 0，显示 0.2.5。上述 Go 命令均在 backend 执行。
- 无未解决冲突；`git diff --check`、`git diff --cached --check`、上游祖先关系及原有文件完整性核对通过。日志与快照：`/tmp/sub2api-codex-ticket-20260918.B6W6Uh/`。未对官方上游验证真实票据或模型质量。

### Notes

- `backend/cmd/server/wire.go`：接入票据采集器的停机清理，保留本地补偿任务清理。
- `backend/cmd/server/wire_gen.go`：重新生成依赖注入，接入票据设置/清理并保留受控 HTTP 传输。
- `backend/internal/config/config.go`：加入默认关闭、TTL、目标模型及采集代理等配置。
- `backend/internal/handler/admin/account_codex_ticket_test.go`：验证管理端票据摘要遵从实时开关。
- `backend/internal/handler/admin/account_data.go`：账号导出移除服务端票据和旧代理私有字段。
- `backend/internal/handler/admin/account_data_handler_test.go`：验证导出去敏不修改原账号数据。
- `backend/internal/handler/admin/account_handler.go`：在账号详情和列表中返回脱敏票据摘要。
- `backend/internal/handler/admin/setting_handler.go`：设置读取返回开关和脱敏代理状态。
- `backend/internal/handler/admin/setting_handler_audit.go`：增加票据开关/代理变更审计字段。
- `backend/internal/handler/admin/setting_handler_codex_ticket_test.go`：验证设置更新、脱敏占位值和旧配置保留。
- `backend/internal/handler/admin/setting_handler_update.go`：保存票据开关与代理并保留未提交或脱敏值。
- `backend/internal/handler/dto/account_mapper_redact_test.go`：验证账号响应不泄露票据正文。
- `backend/internal/handler/dto/mappers.go`：剔除票据私有字段并保留只读摘要映射。
- `backend/internal/handler/dto/settings.go`：补充票据设置响应字段。
- `backend/internal/handler/dto/types.go`：补充账号票据摘要类型。
- `backend/internal/handler/wire.go`：向账号管理处理器注入实时设置服务。
- `backend/internal/repository/account_repo.go`：账号保存行锁内保留最新票据，票据更新不触发调度桶重建。
- `backend/internal/repository/account_repo_codex_ticket_test.go`：验证保存时拒绝伪造/陈旧票据和调度中性更新。
- `backend/internal/repository/account_repo_upstream_billing_probe_update_test.go`：适配账号行锁查询新增的完整 Extra 列。
- `backend/internal/repository/http_upstream.go`：增加独立 HTTP/1.1 不复用连接的采集传输。
- `backend/internal/repository/http_upstream_test.go`：验证采集传输不会复用连接或启用 HTTP/2。
- `backend/internal/server/api_contract_test.go`：更新新增票据设置字段的 API 契约。
- `backend/internal/service/admin_account.go`：新建和编辑拒绝客户端票据注入，同时保留本地并发策略校验。
- `backend/internal/service/admin_account_codex_ticket_test.go`：验证新建/编辑/导出票据私有字段边界。
- `backend/internal/service/domain_constants.go`：增加票据开关和采集代理设置键。
- `backend/internal/service/http_upstream_profile.go`：增加专用采集传输类型。
- `backend/internal/service/http_upstream_profile_test.go`：验证专用传输类型解析。
- `backend/internal/service/openai_account_runtime_block_fastpath.go`：无有效票据时按真实出站模型暂停目标账号调度。
- `backend/internal/service/openai_account_runtime_block_fastpath_test.go`：适配 compact 维度的运行态调度回归。
- `backend/internal/service/openai_account_runtime_transient_test.go`：适配新增 compact 判断参数。
- `backend/internal/service/openai_account_scheduler.go`：候选账号过滤接入票据有效性及 compact 模型口径。
- `backend/internal/service/openai_codex_ticket.go`：实现后台采集、有效期、注入和去敏，并适配本地并发/RPM 预算。
- `backend/internal/service/openai_codex_ticket_lifecycle_test.go`：覆盖采集停止取消、插件隔离、仅读头及影子账号豁免。
- `backend/internal/service/openai_codex_ticket_test.go`：覆盖票据长度/有效期/模型隔离、刷新和 compact 调度。
- `backend/internal/service/openai_codex_turn_state.go`：更新回合状态守卫与新增服务端注入的关系说明。
- `backend/internal/service/openai_gateway_forward.go`：普通 HTTP 出站接入票据注入并保留本地保护流程。
- `backend/internal/service/openai_gateway_messages.go`：Anthropic 兼容桥接入同账号同模型票据。
- `backend/internal/service/openai_gateway_passthrough.go`：HTTP 透传出站接入票据注入并保留本地保护流程。
- `backend/internal/service/openai_gateway_scheduling.go`：各账号选择路径按实际出站模型判断票据。
- `backend/internal/service/openai_gateway_service.go`：接入票据缓存与后台生命周期，同时保留视频补偿成员。
- `backend/internal/service/openai_guardian_affinity_test.go`：适配运行态检查的 compact 参数。
- `backend/internal/service/openai_ws_forwarder_payload.go`：WS 握手接入票据，保留会话和设备身份投影。
- `backend/internal/service/openai_ws_forwarder_support.go`：上一响应绑定账号检查接入 compact 票据判断。
- `backend/internal/service/setting_codex_ticket_test.go`：覆盖实时开关、代理缓存、脱敏与 URL 校验。
- `backend/internal/service/setting_gateway_runtime.go`：加入票据设置短期缓存和失效逻辑。
- `backend/internal/service/setting_parse.go`：解析后台开关与采集代理，支持启动配置回退。
- `backend/internal/service/setting_service.go`：加入票据设置缓存状态与并发去重。
- `backend/internal/service/setting_update.go`：校验保存代理并使运行态缓存失效。
- `backend/internal/service/settings_view.go`：补充系统设置票据字段。
- `deploy/config.example.yaml`：列出票据默认关闭及各配置项的使用说明。
- `frontend/src/api/admin/settings.ts`：增加票据设置请求/响应类型。
- `frontend/src/components/account/AccountUsageCell.vue`：展示按模型的票据有效期或无票状态。
- `frontend/src/components/account/EditAccountModal.vue`：同时保留票据状态、现有 OAuth 保护及指纹配置。
- `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts`：验证票据摘要展示及 setup-token 用量查询边界。
- `frontend/src/i18n/locales/en/admin/accounts.ts`：补充英文账号票据状态文案。
- `frontend/src/i18n/locales/en/admin/settings.ts`：补充英文票据开关与代理配置文案。
- `frontend/src/i18n/locales/zh/admin/accounts.ts`：补充中文账号票据状态文案。
- `frontend/src/i18n/locales/zh/admin/settings.ts`：补充中文票据开关与代理配置文案。
- `frontend/src/types/index.ts`：增加账号票据摘要前端类型。
- `frontend/src/views/admin/SettingsView.vue`：加入默认关闭的票据开关与独立代理输入。
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`：验证票据开关和代理保存交互。
- `backend/internal/service/openai_codex_ticket_protection_test.go`：新增采集预算及三条出站路径/四种指纹模式共存回归。
- `docs/OPENAI_CODEX_TICKETS.md`：记录用法、默认行为、风险边界、截图逐项对应和回滚命令。
- `docs/OPENAI_OAUTH_PROTECTION.md`：追加后台采集与既有并发/RPM 保护的关系。
- `docs/UPSTREAM_SYNC.md`：追加本次定向合并、版本保持、验证和回滚点。
- `progress.md`：仅追加本轮施工记录。
- 回滚：保留分支 `codex/pre-codex-ticket-20260918`（`03f6b278af9e917f96e075b8e55f4bb42bb4ee79`）和 stash `20bc5d35e6d8a73d3e636d9075b5d97cc71a90ec`。在仓库根目录依次执行 `git worktree add -b codex/recover-pre-ticket-20260918 /tmp/sub2api-pre-ticket-20260918 codex/pre-codex-ticket-20260918`、`git -C /tmp/sub2api-pre-ticket-20260918 stash apply --index 20bc5d35e6d8a73d3e636d9075b5d97cc71a90ec`、`git -C /tmp/sub2api-pre-ticket-20260918 add -f -N docs/upstream-sync-20260915-files.md`，可在独立目录恢复原代码与定制；已有被忽略文档仍保存在当前工作区，不删除原目录和备份。本轮未执行回滚。

## 2026-09-18 - Task: 构建并推送 Codex 票据功能双架构镜像

### What was done

- 基于当前完整工作区构建并发布 `iotwq/china-api:latest` 与 `iotwq/china-api:0.2.5-20260918-224942`，覆盖 `linux/amd64`、`linux/arm64`；保留全部本地定制，包含刚合并的票据功能与采集并发兼容适配。
- 应用版本保持 `0.2.5`，构建标识 `86f0e36653d2-dirty`，时间 `2026-09-18T14:49:42Z`。两个标签均为 `sha256:f2b72eda8f9cb958194e1b54c6995d55b2c0eac5ff3bfe955fff0e8efc183099`。
- 记录发布产物、验证边界与上一版回滚摘要。本轮未改业务代码、未推送 GitHub、未部署或操作生产数据库。

### Testing

- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --file Dockerfile --build-arg VERSION=0.2.5 --build-arg COMMIT=86f0e36653d2-dirty --build-arg DATE=2026-09-18T14:49:42Z --tag iotwq/china-api:latest --tag iotwq/china-api:0.2.5-20260918-224942 --metadata-file /tmp/sub2api-image-ticket-20260918.F4GrI8/metadata.json --progress plain --push .`：退出 0；镜像内前端翻译 3 项、类型检查、生产构建及两个架构 Go 编译均通过，保留已有 Browserslist/大 chunk 提示。
- 远端 `latest` 与固定标签均通过 `docker buildx imagetools inspect`，摘要与 metadata 一致，包含 amd64、arm64 两个运行清单；amd64 为 `sha256:49e28b09f6e8d5d7465ae5a843afe295b524007b4b0d71422731e9de09565787`，arm64 为 `sha256:06679afbfc6201fe3358e070590a792bb377ef3d81c943da707137d3dbb733cf`。
- arm64 按发布 manifest 从 Docker Hub 拉取成功；amd64 拉取及有限重试均遇到 CDN/registry EOF，改用相同参数与构建缓存导出 Docker 镜像，导出 manifest 和加载后的 RepoDigests 均与远端发布清单完全一致。首次按 config digest 运行被当前 Docker 镜像存储报告不存在，改用已验证的固定标签后成功；保留失败日志，不将远端拉取记为成功。
- 两个架构在断网、只读、用户 `1000:1000` 的临时容器执行 `--version` 均退出 0，版本、commit 和时间完全一致。程序化核对两个远端标签、构建 metadata、amd64 本地清单及两份版本输出，结果 PASS；此验证不是线上端到端验证。
- 本轮沿用源码阶段已通过的前端 2381 项、后端全量 unit、最终兼容适配后的定向 race 和临时数据库集成测试，不重复无源码变化的全套测试。`git diff --check` 与 `git diff --cached --check` 通过。证据目录 `/tmp/sub2api-image-ticket-20260918.F4GrI8/`。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加镜像标签、双架构摘要、验证方式、部署边界及上一版回滚命令。
- `progress.md`：仅追加本次构建推送和验证记录。
- 回滚：上一固定标签为 `iotwq/china-api:0.2.5-20260918-211628`；恢复 latest 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:d1733b09ba45e6eecdb17900a35816ad9eb2b2111984fd2dc2556b0b110586d3`。本轮未执行回滚。

## 2026-09-19 - Task: 292 打票按账号选择并支持全局强制

### What was done

- 保留「292 打票」为总开关，新增默认关闭的「强制所有账号使用 292 打票」。总开关开启但未强制时，原有和新建账号默认正常转发，仅管理员在添加/编辑账号时明确启用的账号参与采集、注入及目标模型缺票拦截。全局强制不会改写账号选择，关闭强制后恢复各账号选择。
- 后台采集、实际请求准入/注入和管理端票据摘要共用账号参与条件；未参与账号即使保留历史票据也不注入，不显示缺票暂停。所有参与账号继续使用全局打票代理；API Key、其他渠道、影子账号仍不参与。
- 按本次需求，参与账号无有效票据时始终拦截目标模型；旧 fail_closed=false 配置保留解析但不再绕过拦截，文档明确升级变化。总开关关闭时统一停用票据流程，其他模型保持原有行为。
- 检查发现调度缓存投影会丢弃新增账号选择，以失败回归证实后加入投影白名单；选择变更仍触发原有调度缓存同步，票据私有正文不进入候选元数据。
- 更新中英文界面、配置说明及升级/回退文档。版本保持 0.2.5，无数据库迁移；未操作生产设置、调用官方采集、构建 Docker 镜像或部署。

### Testing

- 后端定向 unit：service/admin 的票据用例通过，覆盖默认关闭、显式启用、全局强制与关闭总开关，以及采集/准入/注入/状态一致性。新增保存测试验证账号开关可反复开关且保留已有票据。
- `cd backend && go test -mod=readonly -race -tags=unit ./internal/service ./internal/handler/admin ./internal/repository -run 'Test.*(CodexTicket|OpenAICodexTicket|AccountCodexTicket|LockAndMergeAccountExtra|AccountTraffic|CodexProtection)' -count=1 -timeout=5m`：通过，含 HTTP/透传/WS 身份兼容、采集并发保护、停止取消及设置热更新回归。
- `TestSchedulerMetadataCodexTicketChoiceSurvivesProjection`：先失败（期望 true，实际 nil）；加入投影后，`go test -mod=readonly -race -tags=unit ./internal/repository -run 'Test.*(SchedulerMetadata|FilterScheduler|CodexTicket|LockAndMergeAccountExtra)' -count=1 -timeout=3m` 全部通过。
- `go test -mod=readonly -tags=unit ./internal/server ./internal/config -run 'TestAPIContracts|TestConfigKeysAreEnvReachable' -count=1 -timeout=3m`：通过，验证新增全局字段的接口契约和环境变量可达性。
- 前端添加/编辑账号、用量状态及设置页 4 个测试文件共 203 项通过；包含真实组件勾选、保存、回填和导入账号流程。`pnpm lint:check`、`pnpm build`（含翻译完整性、vue-tsc、Vite）通过。保留既有测试 router-link stub、Node localStorage、Browserslist 和大 chunk 提示。
- 后端 embed 生产构建通过，最终缓存投影补齐后重新构建；`git diff --check` 与 `git diff --cached --check` 通过。本轮验证使用模拟上游与仓库夹具，未连接生产数据库或验证官方 292 票据实际效果。日志保存在 `/tmp/sub2api-ticket-scope-20260919.EaxAXA/`。

### Notes

- `backend/internal/config/config.go`：新增默认关闭的 force_all，标注旧 fail_closed 弃用。
- `backend/internal/service/openai_codex_ticket.go`：统一按总开关、强制开关和账号选择控制采集、准入、注入及状态。
- `backend/internal/service/domain_constants.go`：新增全局强制设置键。
- `backend/internal/service/setting_service.go`：增加强制开关的短期缓存和去重状态。
- `backend/internal/service/setting_gateway_runtime.go`：复用布尔设置读取路径，支持强制开关热更新和缓存失效。
- `backend/internal/service/settings_view.go`：增加管理设置中的全局强制字段。
- `backend/internal/service/setting_parse.go`：读取全局强制值并支持启动配置回退。
- `backend/internal/service/setting_update.go`：持久化全局强制值。
- `backend/internal/handler/dto/settings.go`：公开全局强制设置响应字段。
- `backend/internal/handler/admin/setting_handler.go`：设置读取返回全局强制值。
- `backend/internal/handler/admin/setting_handler_update.go`：保存及回显全局强制值，旧请求未提交时保留原值。
- `backend/internal/handler/admin/setting_handler_audit.go`：记录全局强制开关的变更审计。
- `backend/internal/handler/admin/account_handler.go`：账号列表和详情状态采用实时强制策略。
- `backend/internal/repository/scheduler_cache.go`：候选账号缓存保留账号打票选择。
- `backend/internal/service/openai_codex_ticket_selection_test.go`：新增策略组合、历史票据不自动启用、热更新和其他账号隔离测试。
- `backend/internal/service/openai_codex_ticket_test.go`：既有票据夹具显式启用，断言参与账号缺票拦截。
- `backend/internal/service/openai_codex_ticket_protection_test.go`：保护线路联合回归显式启用账号打票。
- `backend/internal/service/admin_account_codex_ticket_test.go`：验证账号选择保存、关闭和票据保留。
- `backend/internal/handler/admin/account_codex_ticket_test.go`：验证账号状态与按账号/强制策略一致。
- `backend/internal/handler/admin/setting_handler_codex_ticket_test.go`：验证强制开关保存、省略保留和热更新。
- `backend/internal/repository/account_repo_codex_ticket_test.go`：验证账号选择变更需要同步调度状态。
- `backend/internal/repository/scheduler_cache_test.go`：验证账号选择经调度投影保留，票据正文继续剔除。
- `backend/internal/server/api_contract_test.go`：更新设置返回字段契约。
- `frontend/src/api/admin/settings.ts`：增加全局强制设置类型。
- `frontend/src/views/admin/SettingsView.vue`：增加全局强制开关及保存逻辑。
- `frontend/src/components/account/CreateAccountModal.vue`：增加默认关闭的账号打票选择及创建/导入提交。
- `frontend/src/components/account/EditAccountModal.vue`：增加账号选择的展示、回填、关闭和保存。
- `frontend/src/i18n/locales/zh/admin/settings.ts`：明确总开关及强制全部的中文语义。
- `frontend/src/i18n/locales/en/admin/settings.ts`：同步英文设置说明。
- `frontend/src/i18n/locales/zh/admin/accounts.ts`：新增账号选择中文文案并更新票据状态说明。
- `frontend/src/i18n/locales/en/admin/accounts.ts`：同步英文账号文案。
- `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts`：验证创建默认关闭及导入提交账号选择。
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`：验证账号开关回填、保存及适用范围。
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`：验证强制开关默认关闭及独立提交。
- `deploy/config.example.yaml`：说明 force_all 配置及旧 fail_closed 的行为变化。
- `docs/OPENAI_CODEX_TICKETS.md`：更新使用流程、策略表、兼容边界及回退步骤。
- `progress.md`：仅追加本轮实现与验证记录。
- 回退：仅恢复原链路可在后台关闭「292 打票」并保存。若需回滚程序，先关闭总开关，再在部署文件中将应用镜像改为 `iotwq/china-api:0.2.5-20260918-224942`，执行 `docker compose up -d --no-deps --force-recreate sub2api`；旧镜像忽略账号选择，总开关开启会恢复旧的全账号打票行为。本轮未执行回滚。

## 2026-09-19 - Task: 同步原项目最新提交至 0.2.7 并保留本地定制

### What was done

- 从原项目 `Wei-Shaw/sub2api` 获取并合并 `upstream/main` 至 `1a9d49e16f7a22c432b428fce4af8d731f1fa364`，同步 19 个提交、60 个上游变更文件；本地合并提交 `1f0e0257a1aa1d12bc9fef28246ca80ab2f992d5`。版本随正式上游更新为 0.2.7，无新增迁移或依赖清单变动。
- 包含 Seedance 原生异步视频任务、插件宿主服务/只读状态桥、Gemini thinking 与 SSE、DeepSeek 推理历史、Anthropic 工具联合类型、国内 Coding Plan 额度 403 处理及移动端模型广场入口修复。
- 合并前备份并恢复 503 个本地已有修改文件，保留原暂存内容及 292 按账号选择、全局强制、OAuth 线路保护、媒体、账务和群聊等定制。恢复时解决 4 处冲突：路由测试、创建账号、编辑账号、能力类型；本地 Gemini/MiniMax 和上游 Seedance 能力共存，恰好两个非默认能力不会误当默认而被省略。
- 将插件账号目录注入纳入正式 Wire provider，防止重新生成装配代码时丢失上游的手工注入；增加创建/编辑账号混合能力提交回归。变动限于本次同步及兼容，未构建/推送 Docker 镜像、推送 GitHub、部署或修改线上数据。

### Testing

- `cd frontend && pnpm test:run`：316 个测试文件、2395 项测试通过，包含 6 项新增混合能力创建/回填/保存交互回归，以及已有 292 设置和账号选择验证。
- `cd frontend && pnpm lint:check`、`pnpm build`：通过，包含翻译完整性、vue-tsc 类型检查及生产打包；保留既有 Node localStorage、组件 stub、Browserslist 和大 chunk 提示。
- `cd backend && go generate ./cmd/server`：通过；生成文件保留本地控制传输、票据生命周期及视频补偿装配，并通过正式 provider 注入插件账号目录。
- `cd backend && go test -mod=readonly -tags=unit ./... -count=1 -timeout=6m`：全量通过，覆盖新增 Seedance 转发/任务归属/结算认领、Gemini/DeepSeek/Anthropic 兼容及插件 Redis 仿真存储等。总耗时约 221 秒。
- `cd backend && go test -mod=readonly -race -tags=unit ./internal/service ./internal/handler ./internal/handler/admin ./internal/repository ./internal/pkg/tlsfingerprint -run 'Test.*(Seedance|Plugin|CodexTicket|AccountCodexTicket|AccountTraffic|CodexProtection|GrokMedia|GrokVideo|LocalTLS)' -count=1 -timeout=5m`：全部通过。首次命令误带不存在的 `internal/pkg/http`，因此退出 1；删除错误路径并补入 LocalTLS 用例后重跑通过，未将首次运行记作全通过。
- 后端 `go build -mod=readonly -tags embed -trimpath -ldflags '-s -w -X main.Version=0.2.7' -o /tmp/sub2api-sync-20260919.7pjorX/sub2api ./cmd/server` 通过；该产物执行 `--version` 返回 0.2.7，退出 0。
- `git diff --check`、`git diff --cached --check` 通过，无未解决冲突。基于合并前 SHA256 清单核对：503 个原文件无缺失、无意外内容修改、无原暂存状态丢失。本次上游重叠、兼容适配、测试与记录之外的文件哈希一致。
- 完整日志、原始工作区补丁、逐文件快照和验证脚本位于 `/tmp/sub2api-sync-20260919.7pjorX/`。本轮使用模拟上游与测试夹具，未连接生产账号验证官方票据或视频结果，未执行生产数据库操作。

### Notes

- `backend/cmd/server/VERSION`：跟随上游更新为 0.2.7。
- `backend/cmd/server/wire_gen.go`：重新生成插件 KV 及账号目录装配，保留本地依赖注入。
- `backend/internal/handler/admin/plugin_handler.go`：同步插件只读状态桥处理器。
- `backend/internal/handler/endpoint.go`：登记 Seedance 任务端点。
- `backend/internal/handler/grok_media.go`：在保留本地逻辑基础上增加 Seedance 平台选择、转发与结算分支。
- `backend/internal/handler/grok_media_slots_test.go`：调整共享媒体处理器夹具以验证两类平台。
- `backend/internal/handler/seedance.go`：同步 Seedance 原生任务处理与完成结算认领。
- `backend/internal/handler/seedance_test.go`：同步任务生命周期、归属隔离和重复轮询认领测试。
- `backend/internal/pkg/apicompat/responses_to_anthropic_request.go`：接入工具根联合类型转换。
- `backend/internal/pkg/apicompat/responses_to_anthropic_tool_schema.go`：同步 Anthropic 工具根 schema 摊平逻辑。
- `backend/internal/pkg/apicompat/responses_to_anthropic_tools_test.go`：同步根联合类型兼容验证。
- `backend/internal/repository/plugin_kv_store.go`：同步插件按命名空间隔离的 Redis 存储。
- `backend/internal/repository/plugin_kv_store_test.go`：同步存储隔离、TTL 及键值验证测试。
- `backend/internal/repository/wire.go`：登记插件 KV 存储 provider。
- `backend/internal/server/routes/admin.go`：同步插件只读桥管理路由。
- `backend/internal/server/routes/gateway.go`：同步 Seedance 四种任务路由前缀。
- `backend/internal/server/routes/gateway_model_allowlist_test.go`：保留本地视频/音频/Nano Banana 断言并加入 Seedance。
- `backend/internal/server/routes/seedance_test.go`：同步 Seedance 路由与内容类型验证。
- `backend/internal/service/account.go`：增加显式 Seedance 能力判定，保留本地能力。
- `backend/internal/service/antigravity_gateway_gemini.go`：接入裸模型 thinking 变体解析。
- `backend/internal/service/antigravity_gateway_streaming.go`：同步 Gemini SSE 注释兼容处理。
- `backend/internal/service/antigravity_gemini_thinking_variant.go`：同步 thinking 配置与模型变体解析逻辑。
- `backend/internal/service/antigravity_gemini_thinking_variant_test.go`：同步 thinking 变体选择测试。
- `backend/internal/service/gateway_forward_as_responses_test.go`：合并 DeepSeek reasoning_content 转发验证。
- `backend/internal/service/gemini_sse_comment_compat.go`：同步 SSE 注释保活兼容逻辑。
- `backend/internal/service/gemini_sse_comment_compat_test.go`：同步 SSE 注释与内容混合回归。
- `backend/internal/service/grok_media.go`：扩展共享任务查找平台选择，保留原 Grok 入口。
- `backend/internal/service/openai_gateway_cc_pipeline.go`：同步 DeepSeek thinking 历史占位补齐。
- `backend/internal/service/openai_gateway_deepseek_chat_reasoning_test.go`：同步推理历史缺失与明文保留验证。
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`：同步 DeepSeek 推理占位判断逻辑。
- `backend/internal/service/openai_plugin_account_directory.go`：同步插件 OAuth 账号目录及出站身份解析。
- `backend/internal/service/plugin_host_services.go`：同步插件宿主存储、能力检查及账号目录 RPC。
- `backend/internal/service/plugin_host_services_broker_test.go`：同步宿主服务 broker 验证。
- `backend/internal/service/plugin_host_services_test.go`：同步宿主存储与账号目录能力隔离验证。
- `backend/internal/service/plugin_manager.go`：同步宿主依赖、账号目录注入和只读桥转发。
- `backend/internal/service/plugin_manager_routing_test.go`：适配新增宿主依赖与桥测试。
- `backend/internal/service/plugin_runtime.go`：同步插件宿主服务启动及生命周期。
- `backend/internal/service/plugin_runtime_integration_test.go`：适配新增 runtime 依赖参数。
- `backend/internal/service/ratelimit_cn_providers.go`：同步国内 Coding Plan 额度耗尽 403 识别。
- `backend/internal/service/ratelimit_service.go`：接入新增配额暂停条件。
- `backend/internal/service/ratelimit_service_401_test.go`：同步既有限流测试夹具调整。
- `backend/internal/service/ratelimit_service_cn_quota_403_test.go`：同步配额 403 暂停与误判隔离测试。
- `backend/internal/service/seedance.go`：同步原生任务协议、模型映射、转发及 token 用量读取。
- `backend/internal/service/seedance_test.go`：同步创建/查询/删除、能力限制和错误透传验证。
- `backend/internal/service/wire.go`：用正式插件 provider 保留重新生成后的账号目录注入。
- `backend/pkg/pluginapi/README.md`：同步插件宿主服务使用说明。
- `backend/pkg/pluginapi/docs/ui-bridge.md`：同步只读桥接说明。
- `backend/pkg/pluginapi/v1/plugin.pb.go`：同步宿主服务消息生成代码。
- `backend/pkg/pluginapi/v1/plugin.proto`：同步宿主服务及账号目录协议定义。
- `backend/pkg/pluginapi/v1/plugin_grpc.pb.go`：同步宿主服务 gRPC 生成代码。
- `backend/pkg/pluginapi/v1/runtime.go`：同步 SDK 宿主客户端连接支持。
- `docs/seedance-api.md`：同步原生接口、计费、轮询和任务绑定边界说明。
- `frontend/src/api/admin/plugins.ts`：同步插件只读状态 API。
- `frontend/src/components/account/BulkEditAccountModal.vue`：同步 Seedance 批量能力选择。
- `frontend/src/components/account/CreateAccountModal.vue`：合并 Seedance 与本地 Gemini/MiniMax 能力，保留账号打票选择。
- `frontend/src/components/account/EditAccountModal.vue`：合并五类能力回填与保存，保留本地账号保护设置。
- `frontend/src/components/account/__tests__/BulkEditAccountModal.spec.ts`：同步 Seedance 批量编辑验证。
- `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts`：新增混合能力的实际创建提交验证。
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`：合并 Seedance 验证并增加混合能力回填保存测试。
- `frontend/src/components/layout/AppHeader.vue`：同步移动端模型广场入口调整。
- `frontend/src/types/index.ts`：合并 Seedance 与本地 Gemini/MiniMax 能力类型。
- `frontend/src/views/admin/PluginsView.vue`：同步插件只读状态展示。
- `docs/UPSTREAM_SYNC.md`：追加同步范围、兼容处理、验证、部署边界与回滚说明。
- `progress.md`：仅追加本轮执行与验证记录。
- 回滚点：分支 `codex/pre-upstream-sync-20260919` 指向合并前 `86f0e36653d25808ec2e6e150d43b42050d0eb88`，完整 stash `09a7c5a6082856ff0dc09794166473be84c755e1` 保留。需要恢复时在仓库根目录依次执行以下命令，在独立目录恢复，不覆盖当前工作区；本轮未执行回滚。

```sh
git worktree add -b codex/recover-pre-sync-20260919 /tmp/sub2api-pre-sync-20260919 codex/pre-upstream-sync-20260919
git -C /tmp/sub2api-pre-sync-20260919 stash apply --index 09a7c5a6082856ff0dc09794166473be84c755e1
git -C /tmp/sub2api-pre-sync-20260919 restore --staged -- docs/OPENAI_CODEX_TICKETS.md docs/upstream-sync-20260915-files.md
git -C /tmp/sub2api-pre-sync-20260919 add -f -N docs/OPENAI_CODEX_TICKETS.md docs/upstream-sync-20260915-files.md
```

## 2026-09-19 - Task: 构建并推送 0.2.7 双架构镜像

### What was done

- 从包含本地未提交定制的完整工作区构建并推送 `iotwq/china-api:latest` 和 `iotwq/china-api:0.2.7-20260919-205646`，支持 linux/amd64、linux/arm64。版本为 0.2.7，commit 为 1f0e0257a1aa-dirty，构建时间为 2026-09-19T12:56:46Z。
- 包含最新上游 19 个提交与兼容修复、292 按账号选择和全局强制，以及原有 OAuth、媒体计费、群聊等本地定制。本轮未修改业务代码，未部署或修改生产数据。
- 两个标签的 OCI index 均为 `sha256:2561a295134e33c0a7da413838b9262d0e6e32f6f86fb2ebdfa98122ac492cdb`。amd64 为 `sha256:8275a90906d7397fa29448de77d76396f1ee22004784fb965b40d82e8132c0c7`，arm64 为 `sha256:06f78c1c39937ea4db5190193f616bc8d66b8f90598fb68b433d2f8e46044034`。

### Testing

- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64` 携带版本、commit、时间参数及双标签执行 `--push`，退出 0；镜像内翻译完整性、类型检查、前端生产构建和两个架构的后端编译通过。
- `docker buildx imagetools inspect` 核对两个远端标签的 index 与两个平台 manifest，均与构建 metadata 一致。
- 两个架构均从 Docker Hub 按上述 manifest 拉取成功，并通过 `docker run --rm --platform <架构> --network none --read-only --user 1000:1000 --entrypoint /app/sub2api <镜像摘要> --version` 检查，退出 0，版本、commit、时间一致。AMD64 首次拉取 EOF，重试通过。
- 验证脚本核对 metadata、远端清单、本地架构/RepoDigests 和两个版本日志，返回 PASS。证据保存于 `/tmp/sub2api-image-sync-20260919.LfGeIX/`。
- 沿用前轮源码同步的前端 2395 项测试、后端全量 unit 和定向 race 结果；本轮无业务代码变化。保留既有前端 Browserslist/大 chunk 提示。发布前 `git diff --check` 和 `git diff --cached --check` 通过，本地修改保留核对无异常。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加镜像标签、摘要、平台验证、292 升级边界及回滚说明。
- `progress.md`：仅追加本轮构建推送与验证记录。
- 回滚：上一固定标签为 `iotwq/china-api:0.2.5-20260918-224942`。先关闭 292 总开关，避免旧版恢复全账号打票行为，再部署上一固定标签；需要恢复 latest 时执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:f2b72eda8f9cb958194e1b54c6995d55b2c0eac5ff3bfe955fff0e8efc183099`。本轮未执行回滚。

## 2026-09-20 - Task: 支持 Dola ViralDance 2.0 和 2.5 视频模型

### What was done

- 根据用户确认及完整读取的 `viraldance933.md`，两个新模型除名称外沿用 933 协议：通过普通 OpenAI APIKey 账号调用 `/v1/videos`，JSON 参数透传；必填 4–15 整数秒 duration，计费分辨率为 720p。
- 将 `dola-viraldance2.0`、`dola-viraldance2.5` 加入现有视频规格表及支持模型提示；自动复用现有解析、账号能力判断、模型映射、预扣结算、绑定账号查询与失败退款路径，无需新增路由或专用能力开关。
- 原样使用新模型名，只有管理员配置账号映射时才替换为映射目标；不自动改名为 933、不自动继承其他模型价格、不开放未知名称后缀。实际素材、比例、分辨率参数校验沿用已有透传策略由上游负责；本地仍在调用前拒绝缺失、非整数及超范围 duration。
- 同步接入说明及回归测试。未修改其他模型的业务规则、前端用户接口页、生产账号、价格或数据库；未构建或推送镜像、未部署。

### Testing

- 先添加两个模型的创建转发/处理器用例，执行 `cd backend && go test -mod=readonly -tags=unit ./internal/service ./internal/handler -run 'TestViraleeVideo(sCreateBindChargeAndPoll|CreateAndResultPassthrough)/dola' -count=1 -timeout=3m`：修改前两个模型均复现不支持错误，处理器返回 HTTP 400，证明拦截发生在本地。
- 实现后执行 `cd backend && go test -mod=readonly -tags=unit ./internal/service ./internal/handler ./internal/server/routes -run 'Test.*(Viralee|Dola|OpenAIVideo|Videos|VideoStatus|MiniMax|Firefly|Seedance)' -count=1 -timeout=5m`：三个包全部通过。
- 新模型回归覆盖 4/15 秒边界、720p 计费、无 `/v1` 和带 `/v1` 的基础地址、id/task_id 返回、多模态和 elements 透传、原模型名及账号映射、排队/处理/完成/失败状态、原账号绑定查询、查询不重复计费、按秒/按次定价、用量记录及重复退款只退一次；缺时长、小数、字符串、null、3/16 秒和只传 seconds 均拒绝。
- 保留 Wan 30 秒计费/退款回归，并运行既有 Firefly、MiniMax、Seedance 和视频路由用例。测试基于模拟上游，不代表已对真实上游完成付费生成验证。
- `gofmt -l` 检查四个本轮 Go 文件无输出；`git diff --check` 通过。按施工前备份核对，生产逻辑只改模型规格表一行与支持模型提示一行。

### Notes

- `backend/internal/service/openai_viralee_video.go`：将两个 Dola 模型接入 933 的 720p、4–15 秒必填规格。
- `backend/internal/service/openai_videos.go`：在支持模型错误提示中列出两个新增名称。
- `backend/internal/service/openai_viralee_video_test.go`：增加新模型转发、映射、时长校验和计费测试，扩展既有记录及幂等退款测试。
- `backend/internal/handler/openai_viralee_videos_test.go`：扩展创建、冻结结算、账号绑定、查询及不重复收费测试。
- `docs/OPENAI_MEDIA_COMPAT.md`：记录两模型接入条件、参数、映射、价格与查询方式。
- `progress.md`：仅追加本轮实现与验证记录。
- 回滚点：施工前六个文件完整副本保留于 `/tmp/sub2api-dola-video-20260920.VzQVHD/`。如需只关闭新增支持，在 `lookupViraleeVideoSpec` 的 933 分支移除两个 `dola-viraldance` 名称，并从支持模型提示移除同名项即可；原有模型继续工作。若需恢复本轮全部实现文件且其后没有新增修改，可在仓库根目录执行下列命令；保留 progress 历史，不覆盖其他定制。本轮未执行回滚。

```sh
cp /tmp/sub2api-dola-video-20260920.VzQVHD/openai_videos.go backend/internal/service/openai_videos.go
cp /tmp/sub2api-dola-video-20260920.VzQVHD/openai_viralee_video.go backend/internal/service/openai_viralee_video.go
cp /tmp/sub2api-dola-video-20260920.VzQVHD/openai_viralee_video_test.go backend/internal/service/openai_viralee_video_test.go
cp /tmp/sub2api-dola-video-20260920.VzQVHD/openai_viralee_videos_test.go backend/internal/handler/openai_viralee_videos_test.go
cp /tmp/sub2api-dola-video-20260920.VzQVHD/OPENAI_MEDIA_COMPAT.md docs/OPENAI_MEDIA_COMPAT.md
```

## 2026-09-20 - Task: Codex 打票支持按账号选择 Pro 292 / Team 332

### What was done

- 用户确认 Team 长度为 332，原消息中的 232 为笔误。在编辑 OpenAI OAuth / setup-token 非影子账号时新增打票类型：Pro（292）、Team（332）、跟随服务配置（默认 292）；不自动推断套餐，不改变账号参与开关。
- 统一后台采集验票、临期刷新、缓存/持久化取票、调度缺票拦截、HTTP/透传 HTTP/WS 注入及状态摘要的目标长度。账号选择优先于全局默认；全局强制只覆盖是否参与，不覆盖账号类型。
- 账号类型存入 Extra 的 codex_ticket_type，取值 pro/team，省略时保留旧服务配置回退；创建、编辑、Extra 更新和批量更新校验合法值。调度元数据投影保留新字段，类型变更触发现有缓存同步。
- 管理状态增加 target_length，缺票文案按实际 292/332 展示；系统设置和账号参与开关改为通用「Codex 打票」名称，保持原设置键、值和代理配置。类型切换后长度不匹配的旧票据不会被注入或视为就绪。
- 保持一小时有效期、提前十分钟刷新、采集代理、协议、并发/RPM 预算、账号/模型隔离、默认不参与及总开关行为。本轮不改数据库结构，不自动启用任何账号；版本保持 0.2.7，未调用官方账号、构建 Docker 镜像或部署。

### Testing

- 后端首轮票据/账号/缓存定向 race 测试通过；扩展 Pro/Team 的 HTTP、透传 HTTP、WS 与四种指纹模式联合覆盖后，运行 `cd backend && go test -mod=readonly -race -tags=unit ./internal/service ./internal/handler/admin ./internal/repository ./internal/server -run 'Test.*(CodexTicket|OpenAICodexTicket|AccountCodexTicket|LockAndMergeAccountExtra|SchedulerMetadataCodexTicket|APIContracts|AccountTraffic|CodexProtection)' -count=1 -timeout=5m`：全部通过。
- 新回归覆盖 astra/sol、Pro/Team 错长度拒绝、全局强制保留账号类型、有效票据注入与状态一致性、持久化后重启取票、有效期一小时、有效票跳过采集、临期自动刷新、切换类型后旧缓存拒绝、仅选择类型不自动参与、默认继承已有服务配置，以及创建/编辑保存和四类写入入口拒绝无效类型。
- 调度缓存测试验证类型不丢失，私有票据仍不进入元数据；类型修改会产生调度同步，服务端票据字段继续保持原有保护。
- 前端 5 个文件共 216 项测试通过，包含编辑账号类型选择/保存/回填/切回默认、参与开关独立、API Key/影子账号隐藏、Team 缺票文案使用 332，以及创建/设置页既有回归和翻译完整性。
- `cd frontend && pnpm lint:check`、`pnpm build` 通过（含 vue-tsc 及生产打包）；后端 `go build -mod=readonly -tags embed -trimpath -o /tmp/sub2api-ticket-types-20260920.ME3r4Y/sub2api ./cmd/server` 通过，执行 `--version` 退出 0，版本为 0.2.7。
- `git diff --check`、`git diff --cached --check` 通过；本轮 Go 文件格式检查通过。保留既有测试 router-link stub、Node localStorage、Browserslist 和大 chunk 提示。使用模拟上游和票据，未实测官方 Team 采集成功率或模型质量。
- 测试日志、构建产物及逐文件施工前备份位于 `/tmp/sub2api-ticket-types-20260920.ME3r4Y/`。

### Notes

- `backend/internal/service/openai_codex_ticket.go`：增加账号类型与统一目标长度解析/校验，接入采集、刷新、注入、调度和状态。
- `backend/internal/service/admin_account.go`：四类账号写入入口校验新类型。
- `backend/internal/service/openai_codex_ticket_type_test.go`：新增 Pro/Team 采集、类型切换、临期刷新、缓存及持久化、默认兼容和保存校验回归。
- `backend/internal/service/openai_codex_ticket_protection_test.go`：联合传输/指纹用例扩展到 292 和 332。
- `backend/internal/repository/scheduler_cache.go`：在调度元数据中保留账号打票类型。
- `backend/internal/repository/scheduler_cache_test.go`：验证新类型投影保留及票据正文剔除。
- `backend/internal/repository/account_repo_codex_ticket_test.go`：验证类型更新触发调度同步。
- `frontend/src/components/account/EditAccountModal.vue`：新增类型选择、回填、保存及动态缺票提示。
- `frontend/src/components/account/AccountUsageCell.vue`：账号列表缺票提示使用服务端目标长度。
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`：验证类型选择、保存回填、默认恢复和适用账号范围。
- `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts`：验证 Team 提示使用 332。
- `frontend/src/types/index.ts`：新增兼容旧服务的可选 target_length 状态字段。
- `frontend/src/i18n/locales/zh/admin/accounts.ts`：增加类型选择中文文案及动态长度提示。
- `frontend/src/i18n/locales/en/admin/accounts.ts`：同步英文账号类型与提示文案。
- `frontend/src/i18n/locales/zh/admin/settings.ts`：设置改为通用 Codex 打票命名，明确强制参与不覆盖类型。
- `frontend/src/i18n/locales/en/admin/settings.ts`：同步英文设置命名和强制行为说明。
- `deploy/config.example.yaml`：明确 target_length 仅为账号未选择类型时的回退值。
- `docs/OPENAI_CODEX_TICKETS.md`：更新操作流程、升级、类型切换边界和回退说明。
- `progress.md`：仅追加本轮施工与验证记录。
- 回滚点：本轮之前文件保留于 `/tmp/sub2api-ticket-types-20260920.ME3r4Y/before/`（保留此前 Dola 改动和所有其他定制）。仅恢复正常转发可在后台关闭「Codex 打票」总开关；若回滚已部署服务，应先关闭总开关，再改为上一已发布镜像 `iotwq/china-api:0.2.7-20260919-205646` 并执行 `docker compose up -d --no-deps --force-recreate sub2api`，避免旧版忽略 Team 类型后按 292 错误暂停。本轮未执行回滚。

## 2026-09-20 - Task: 移除 Codex 打票全局强制参与

### What was done

- 按用户要求移除「强制所有账号使用 Codex 打票」页面开关、设置 API 字段、存取及审计逻辑、运行时缓存和 YAML/env 配置入口。
- 参与规则统一为总开关开启且合格账号明确启用；历史数据库、YAML 和环境变量中的 force_all 不再影响账号。无需迁移或清理数据库，旧客户端提交该设置字段会被忽略。
- 保留按账号启用、Pro 292 / Team 332、采集代理、票据生命周期及现有并发保护。原来仅因全局强制参与的账号升级后恢复正常转发；已手动启用的账号继续验票和缺票拦截。
- 本轮未构建或推送 Docker 镜像、未部署、未修改线上数据。保留此前 Dola 模型和 Pro/Team 修改，版本保持 0.2.7。

### Testing

- 后端通过：`cd backend && go test -mod=readonly -race -tags=unit ./internal/service ./internal/handler/admin ./internal/repository ./internal/server -run 'Test.*(CodexTicket|OpenAICodexTicket|AccountCodexTicket|LockAndMergeAccountExtra|SchedulerMetadataCodexTicket|APIContracts|AccountTraffic|CodexProtection)' -count=1 -timeout=5m`。
- 回归覆盖历史强制 true 下的采集、调度、注入、状态展示、总开关热更新、设置 API 旧字段忽略和不回显，以及 Pro/Team 目标长度和其他账号隔离。测试使用模拟上游，未调用官方账号。
- 前端 5 个文件共 216 项测试通过：CreateAccountModal、EditAccountModal、AccountUsageCell、SettingsView、localeKeyCompleteness；新增验证旧字段不会重新显示开关或提交。
- `cd frontend && pnpm run lint:check`、`pnpm run build` 通过，包含类型检查及生产打包。保留既有 router-link stub 和打包体积提示。
- `git diff --check`、`git diff --cached --check` 通过；逐文件与施工前快照核对，仅修改当前移除范围。运行代码和部署示例的强制全部引用已清除，兼容回归及升级文档保留旧键说明。
- 日志及施工前快照：`/tmp/sub2api-remove-force-all-20260920.d4lflA/`；前端生成文件位于忽略的 `backend/internal/web/dist/`。

### Notes

- `backend/internal/config/config.go`：移除 force_all 配置字段与默认值。
- `backend/internal/service/domain_constants.go`：移除强制全部设置键常量。
- `backend/internal/service/settings_view.go`：移除强制全部设置视图字段。
- `backend/internal/service/setting_service.go`：移除强制全部缓存与 singleflight 状态。
- `backend/internal/service/setting_gateway_runtime.go`：移除强制全部运行时读取和缓存失效处理。
- `backend/internal/service/setting_parse.go`：停止解析数据库强制值及启动配置回退。
- `backend/internal/service/setting_update.go`：停止持久化强制全部设置。
- `backend/internal/service/openai_codex_ticket.go`：参与、调度状态和注入判断仅使用总开关及账号选择。
- `backend/internal/handler/dto/settings.go`：设置响应移除强制全部字段。
- `backend/internal/handler/admin/setting_handler.go`：读取设置响应不再包含强制全部。
- `backend/internal/handler/admin/setting_handler_update.go`：设置更新不再接受、保存或回显强制全部。
- `backend/internal/handler/admin/setting_handler_audit.go`：移除强制全部审计差异项。
- `backend/internal/handler/admin/account_handler.go`：账号状态查询停止读取全局强制。
- `backend/internal/server/api_contract_test.go`：更新设置接口契约，移除已废弃字段。
- `backend/internal/handler/admin/account_codex_ticket_test.go`：验证旧强制值不展示未参与账号的缺票暂停。
- `backend/internal/handler/admin/setting_handler_codex_ticket_test.go`：验证旧客户端请求字段被忽略，旧数据库键不再回显。
- `backend/internal/service/openai_codex_ticket_selection_test.go`：验证旧强制值不能绕过账号选择，保留总开关热更新与账号类型隔离。
- `backend/internal/service/openai_codex_ticket_type_test.go`：Pro/Team 回归改用明确账号参与，保留原验票覆盖。
- `frontend/src/api/admin/settings.ts`：移除前端设置请求及响应类型中的强制字段。
- `frontend/src/views/admin/SettingsView.vue`：移除页面强制开关、初始值及保存字段。
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`：验证旧响应字段不产生开关，也不会被再次提交。
- `frontend/src/i18n/locales/zh/admin/settings.ts`：移除强制全部中文设置文案。
- `frontend/src/i18n/locales/en/admin/settings.ts`：移除强制全部英文设置文案。
- `frontend/src/i18n/locales/zh/admin/accounts.ts`：账号中文说明去除强制覆盖。
- `frontend/src/i18n/locales/en/admin/accounts.ts`：账号英文说明去除强制覆盖。
- `deploy/config.example.yaml`：移除 force_all 及对应环境变量示例，明确必须按账号启用。
- `docs/OPENAI_CODEX_TICKETS.md`：更新使用规则、兼容行为及回退方法。
- `progress.md`：仅追加本轮改动、验证和回滚记录。
- 可执行回滚：先备份后续修改，在仓库根执行 `tar -xf /tmp/sub2api-remove-force-all-20260920.d4lflA/before.tar` 恢复本轮前文件（包括原有定制，不变更 Git 暂存区）。若回滚已部署版本，先关闭 Codex 打票总开关，防止旧强制值重新生效；本轮未执行回滚。

## 2026-09-20 - Task: 构建并推送最新 0.2.7 双架构镜像

### What was done

- 从当前完整工作区构建并推送 `iotwq/china-api:latest` 和 `iotwq/china-api:0.2.7-20260920-204354`，包含 linux/amd64、linux/arm64；版本 0.2.7，commit 1f0e0257a1aa-dirty，构建时间 2026-09-20T12:43:54Z。
- 镜像包含本日 Dola 视频转发、账号 Pro 292 / Team 332 类型选择及移除全局强制参与，同时保留此前定制。本轮未修改业务代码、未部署服务或修改生产数据。
- 两标签 OCI index 为 `sha256:bb7fa507c94188b849710a85f087efde97a9e972c385dfa3ffaddb899b6d706a`，amd64 为 `sha256:48fc5e181208d8b90b1ced5ca68ec30b746c80839f9dfe12e9287395ed739780`，arm64 为 `sha256:5d64d46e7c705257e0520880b7735f4b2fe3f789bfefb1a5ffad41cfbb3bd136`。

### Testing

- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --file Dockerfile --build-arg VERSION=0.2.7 --build-arg COMMIT=1f0e0257a1aa-dirty --build-arg DATE=2026-09-20T12:43:54Z --tag iotwq/china-api:latest --tag iotwq/china-api:0.2.7-20260920-204354 --metadata-file /tmp/sub2api-image-20260920.9Y4zOc/metadata.json --progress plain --push .`：退出 0。
- 镜像内翻译 3 项、类型检查、前端生产构建和两个架构的 Go 编译通过；保留既有 Browserslist/大 chunk 提示。
- `docker buildx imagetools inspect` 核对两个标签，index 与 metadata 一致，明确包含 linux/amd64 和 linux/arm64；unknown/unknown 项为构建证明，不是额外运行架构。
- 两架构按各自 manifest 从 Docker Hub 拉取成功，随后 `docker run --rm --pull never --platform <架构> --network none --read-only --user 1000:1000 --entrypoint /app/sub2api <摘要镜像> --version` 均退出 0，版本、commit 和时间一致。
- 沿用本日 Dola 回归、票据/账号/接口契约定向 race、前端 216 项回归和 lint 结果，本轮不重复整套测试。发布前后 `git diff --check`、`git diff --cached --check` 通过。
- 证据目录：`/tmp/sub2api-image-20260920.9Y4zOc/`，包括 build、inspect-latest/fixed、pull-amd64/arm64、version-amd64/arm64 日志及 metadata。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加本轮发布标签、摘要、双架构验证、升级边界及回滚命令。
- `progress.md`：仅追加本轮构建推送记录。
- 回滚点：`iotwq/china-api:0.2.7-20260919-205646` / `sha256:2561a295134e33c0a7da413838b9262d0e6e32f6f86fb2ebdfa98122ac492cdb`。回滚前关闭 Codex 打票总开关，避免旧版按 292 处理 Team 或重新启用历史强制配置；Compose 改为上一固定标签后执行 `docker compose up -d --no-deps --force-recreate sub2api`。恢复 latest 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:2561a295134e33c0a7da413838b9262d0e6e32f6f86fb2ebdfa98122ac492cdb`；本轮未执行回滚。

## 2026-09-20 - Task: 优化 API 密钥页费用提示文案与视觉

### What was done

- 按用户原文补充「图片和视频的实际价格也是计费价格除6」，同步英文翻译；充值及计费逻辑不变。
- 将原 primary 色边框提示替换为青绿渐变圆角卡片，增加计算器图标、实际费用换算标题、1:6 充值比例标识、顶部细渐变及轻阴影。
- 标题、标识和正文支持窄屏换行与深色模式，使用 note 语义及装饰图标隐藏标记。未新增动画、弹窗或依赖，未改密钥操作和布局组件。

### Testing

- `cd frontend && pnpm exec vitest run src/views/user/__tests__/KeysView.spec.ts src/i18n/__tests__/localeKeyCompleteness.spec.ts`：2 个文件、28 项测试通过，保留既有测试 router 注入提示。
- `pnpm exec eslint src/views/user/KeysView.vue src/i18n/locales/zh/dashboard.ts src/i18n/locales/en/dashboard.ts` 通过。
- `pnpm run build` 通过，包含翻译检查、vue-tsc 和生产打包；仅有既有 Browserslist/大 chunk 提示。
- 浏览器使用源码卡片模板片段、实际中文翻译和生产 CSS 验证桌面、375px 窄屏的浅色/深色表现，正文完整显示且无截断；这是局部视觉预览，未声称完成登录后完整页面端到端验证。临时预览标签与服务已关闭，viewport 已恢复。
- `git diff --check`、`git diff --cached --check` 通过；测试及构建日志保留于 `/tmp/sub2api-key-notice-20260920.qFbOte/`。本轮未构建推送 Docker 镜像或部署。

### Notes

- `frontend/src/views/user/KeysView.vue`：替换顶部提示卡片的结构、样式及可访问性标记。
- `frontend/src/i18n/locales/zh/dashboard.ts`：更新用户要求的完整说明并补标题和标识文案。
- `frontend/src/i18n/locales/en/dashboard.ts`：同步英文说明、标题和标识。
- `docs/API_KEYS_UI.md`：新增提示展示、验证边界与部署状态说明。
- `progress.md`：仅追加本轮记录。
- 回滚：备份后续修改后，在仓库根执行 `tar -xf /tmp/sub2api-key-notice-20260920.qFbOte/before.tar` 恢复三个源码文件及本轮前日志；本轮新建文档可执行 `mv docs/API_KEYS_UI.md /tmp/sub2api-key-notice-20260920.qFbOte/API_KEYS_UI.rolled-back.md` 移出（无数据永久删除）。本轮未执行回滚。

## 2026-09-20 - Task: 发布 API 密钥页提示优化双架构镜像

### What was done

- 从包含本地未提交定制的当前工作区构建并推送 `iotwq/china-api:latest`、`iotwq/china-api:0.2.7-20260920-225931`；平台 linux/amd64、linux/arm64，版本 0.2.7，commit 1f0e0257a1aa-dirty，构建时间 2026-09-20T14:59:31Z。
- 包含新的 API 密钥页费用说明与提示卡片，同时保留此前 Dola、Pro/Team 和移除全局强制等修改。本轮不改业务代码、未部署或修改生产数据。
- 两标签 index 为 `sha256:e0c314ee1112860166cf06896f42d3e6db758789bc2395783256b2c6d62014ed`；amd64 为 `sha256:3f97b2b08159b1d824e367902ce66f8448d67e84cb2f1af35ceab9653563df52`，arm64 为 `sha256:8b17a56637dc9e189cd942cc1c1669bffc52846cb0bc72437041942d42a08764`。

### Testing

- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 --file Dockerfile --build-arg VERSION=0.2.7 --build-arg COMMIT=1f0e0257a1aa-dirty --build-arg DATE=2026-09-20T14:59:31Z --tag iotwq/china-api:latest --tag iotwq/china-api:0.2.7-20260920-225931 --metadata-file /tmp/sub2api-image-keys-20260920.scqMAY/metadata.json --progress plain --push .` 退出 0；前端翻译、类型检查、生产构建及两平台 Go 编译通过。
- 两个远端标签使用 `docker buildx imagetools inspect` 验证；脚本断言摘要与 metadata 一致，并包含两个目标架构，PASS。
- 两架构按 manifest 从 Docker Hub 拉取成功；amd64 下载耗时 158 秒，最终退出 0。两者通过 `docker run --rm --pull never --platform <架构> --network none --read-only --user 1000:1000 --entrypoint /app/sub2api <摘要镜像> --version` 验证，均为 0.2.7，commit/时间正确。
- 沿用 UI 改动的 28 项测试、定向 lint、生产构建和浅色/深色窄屏预览；没有新业务代码变更，不重复测试。保留既有 Browserslist/大 chunk 提示。
- 发布前后 `git diff --check`、`git diff --cached --check` 通过。构建、远端清单、拉取及版本日志、metadata 位于 `/tmp/sub2api-image-keys-20260920.scqMAY/`。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加当前发布标签、摘要、架构验证与回滚说明。
- `progress.md`：仅追加本轮发布与验证记录。
- 回滚点为 `iotwq/china-api:0.2.7-20260920-204354` / `sha256:bb7fa507c94188b849710a85f087efde97a9e972c385dfa3ffaddb899b6d706a`。Compose 镜像改为上一固定标签后执行 `docker compose up -d --no-deps --force-recreate sub2api`；恢复 latest 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:bb7fa507c94188b849710a85f087efde97a9e972c385dfa3ffaddb899b6d706a`。本轮未执行回滚。

## 2026-09-20 - Task: 修复 Dola ViralDance 2.5 视频时长范围

### What was done

- 将 `dola-viraldance2.5` 的服务端时长校验从 4–15 秒改为 4–30 秒；`dola-viraldance2.0` 继续保持 4–15 秒。
- 保留整数、必填、超范围和小数拒绝逻辑，校验仍在上游请求和计费前执行，避免无效请求产生任务或扣费。
- 更新 API 文档、媒体兼容说明和用户接口文档中的 Dola 模型时长范围，并在用户模型表中明确 2.0/2.5 的区别。未改变上游路径、模型映射、轮询、计费或退款逻辑。

### Testing

- `cd backend && go test -mod=readonly -tags=unit ./internal/service -run 'Test(DolaVideo|ViraleeVideo)' -count=1 -timeout=5m`：通过。
- 新增边界竞态验证 `go test -mod=readonly -race -tags=unit ./internal/service -run 'Test(DolaVideoCreateAndResultPassthrough|DolaVideoMappedUpstreamModel|DolaVideoRejectsInvalidDurationBeforeBilling)$' -count=1 -timeout=5m`：通过；覆盖 2.5 的 16/30 秒放行及 31 秒、缺失、小数、字符串、null 拒绝。
- 全量 Dola/Viralee 竞态命令触发仓库已有的测试服务竞态：Codex 打票后台 goroutine 读取服务字段，同时账务退款测试写入同一测试服务状态；该失败不涉及本轮时长逻辑，未修改无关并发代码。
- `cd frontend && pnpm exec vitest run src/i18n/__tests__/localeKeyCompleteness.spec.ts src/views/user/__tests__/ApiDocsView.spec.ts`：4 项通过；`pnpm exec eslint src/views/user/ApiDocsView.vue` 通过；`pnpm run build` 通过，保留既有 Browserslist/大 chunk 提示。
- `git diff --check`、`git diff --cached --check` 通过。施工前快照和日志位于 `/tmp/sub2api-dola-duration-20260920.9iZPTI/`。本轮未构建镜像或部署。

### Notes

- `backend/internal/service/openai_viralee_video.go`：按 Dola 2.0/2.5 分离时长规格。
- `backend/internal/service/openai_viralee_video_test.go`：增加 Dola 2.5 的 4–30 秒边界与转发回归。
- `docs/API_DOCS.md`：补充 Dola 两模型时长差异。
- `docs/OPENAI_MEDIA_COMPAT.md`：更新 Dola 表格、请求参数和范围说明。
- `frontend/src/views/user/ApiDocsView.vue`：更新用户可见模型和 duration 参数说明。
- `progress.md`：追加本轮修复、验证和回滚点。
- 回滚：保存后续修改后，在仓库根执行 `tar -xf /tmp/sub2api-dola-duration-20260920.9iZPTI/before.tar` 恢复本轮前文件；该操作会覆盖这些路径之后的修改，执行前应先备份。本轮未执行回滚。

## 2026-09-20 - Task: 复核并补齐 Dola 2.5 时长修复验证

### What was done

- 复核本任务施工前快照：唯一生产后端改动为将 Dola 2.5 分离为 4–30 秒规格；Dola 2.0 与 933 继续 4–15 秒。请求解析与预估计费共用规格，视频时长读取不会截为 15 秒。
- 修正上轮重复的边界测试结构，补齐 2.0 的 4/15、2.5 的 4/15/16/30 秒合法值，以及两者超上限、低于下限、缺失、小数、字符串、null 等拒绝。
- 补齐 2.5 的 16/30 秒按秒计费、30 秒按次计费、30 秒预冻结与结算、用量保存、失败幂等退款和创建后账号绑定查询验证。
- 上轮 race 命令中的 TestDolaVideoCreateAndResultPassthrough 不存在，正确名称为 TestViraleeVideoCreateAndResultPassthrough；本轮采用 DolaVideo/ViraleeVideo 前缀覆盖全部相关服务与接口用例，补上此前漏跑的转发验证。
- 解决本任务账务测试已复现的竞态：测试构造器自动启动打票后台任务，测试又修改 accountRepo；在替换依赖前同步停止该后台任务，并为接口测试追加清理。只改测试，不修改生产线程、请求、计费或退款逻辑。
- 修正文档“仅模型名不同”和模型数量的旧描述，明确 Dola 素材参数与 933 相同、时长分别校验。

### Testing

- `cd backend && go test -mod=readonly -race -tags=unit ./internal/service ./internal/handler -run 'Test(DolaVideo|ViraleeVideo)' -count=1 -timeout=5m`：两个包全部通过；本次包含此前发生竞态的账务退款用例，未跳过。
- 模拟上游测试验证 30 秒请求字段原样转发及完整结果回传；接口测试验证实际创建 200、按 30 秒预冻结和结算、用量保存 30 秒及轮询不重复扣款；退款测试验证 30 秒对应费用仅退款一次。
- 前端 API 文档与翻译完整性 4 项测试、ApiDocsView ESLint、类型检查及生产构建通过。保留已有 Node localStorage、Browserslist、chunk 体积提示。
- `git diff --check`、`git diff --cached --check` 通过。由于相关文件是之前已存在的未跟踪定制，本轮同时与任务前 tar 快照做逐文件比较，未将 Git 无 diff 输出当作无改动依据。
- 本轮测试日志实际位于 `/tmp/sub2api-dola-duration-review.npqWhf/`；上轮 `/tmp/sub2api-dola-duration-20260920.9iZPTI/` 只有施工前 tar，先前日志中的“快照和日志”表述过宽，以本条为准。未向生产上游发起付费请求，未构建或推送 Docker 镜像、未部署。

### Notes

- `backend/internal/service/openai_viralee_video_test.go`：整理合法/非法边界，增加长视频计费退款验证，在替换依赖前关闭后台采集以消除测试竞态。
- `backend/internal/handler/openai_viralee_videos_test.go`：Dola 2.5 接口用例使用 30 秒，验证预冻结、扣费及用量时长，清理后台任务。
- `frontend/src/views/user/ApiDocsView.vue`：明确 Dola 沿用 933 的 JSON/素材参数，时长使用独立范围。
- `docs/API_DOCS.md`：同步当前六个模型的文档概述。
- `docs/OPENAI_MEDIA_COMPAT.md`：去除“仅模型名不同”的过时表述，明确两模型时长差别。
- `progress.md`：追加本次复核结果、更正验证范围与证据位置。
- 回滚：先备份后续修改，在仓库根执行 `tar -xf /tmp/sub2api-dola-duration-review.npqWhf/before.tar` 仅恢复本次复核前文件。若需撤销整个时长修复，再执行 `tar -xf /tmp/sub2api-dola-duration-20260920.9iZPTI/before.tar`；上述恢复会覆盖相关文件之后的修改，执行前先备份。本轮未执行回滚。

## 2026-09-20 - Task: 构建并发布 Dola 2.5 时长修复双架构镜像

### What was done

- 基于当前完整工作区构建并推送 `iotwq/china-api:latest` 和 `iotwq/china-api:0.2.7-20260920-234609`，包含 amd64、arm64 双架构及 Dola 2.5 的 4–30 秒修复；Dola 2.0 保持 4–15 秒。
- 版本保持 0.2.7，commit 为 `1f0e0257a1aa-dirty`，构建时间 `2026-09-20T15:46:09Z`。本轮未修改业务代码、部署服务或修改生产数据。

### Testing

- Docker Buildx 构建推送退出 0；镜像内翻译检查 3 项、类型检查、前端生产构建与双架构 Go 编译通过。沿用前轮已完成的 Dola/Viralee race 回归和前端验证。
- 两个远端标签与构建 metadata 均指向 `sha256:3a8146091986290abc32bbbbed23b34befb062644c5a59ae259f899d7e244549`。
- amd64 manifest 为 `sha256:f29f3e45cfc82bed3fbb4a5bd62fd85254d57c6e74d7a483f71fdb4a70cd0af7`；arm64 manifest 为 `sha256:21f04af2b43ab822f484b7156ffd41374068148620a66734b5fc928fb728fde5`。均从远端拉取成功，并在断网、只读、非 root 容器运行 `--version` 通过，版本、commit、时间一致。
- 固定标签查询及 arm64 拉取首次发生 Docker Hub EOF，重试通过。构建、查询、拉取及版本检查日志和 metadata 保存在 `/tmp/sub2api-image-dola-duration-20260920.9Teos8/`。未做线上端到端或付费视频生成验证。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加本次镜像标签、摘要、验证证据及回滚说明。
- `progress.md`：追加本轮发布记录；无其他仓库文件改动。
- 回滚：将 Compose 镜像设为 `iotwq/china-api:0.2.7-20260920-225931` 后执行 `docker compose up -d --no-deps --force-recreate sub2api`；恢复 latest 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:e0c314ee1112860166cf06896f42d3e6db758789bc2395783256b2c6d62014ed`。旧镜像的 Dola 2.5 上限仍为 15 秒。本轮未执行回滚。

## 2026-09-24 - Task: 同步 GitHub 原项目最新提交至 0.2.8

### What was done

- 9 月 23 日开始同步，合并 Wei-Shaw/sub2api upstream/main 的 237 个提交、473 个文件至 `a3eb7ef302961cba716dc78b39b93b60c467db0e`；本地合并 `f867e4f9bd2f6bcf1b2c6a67b20d5124cc86748a`，版本 0.2.8。
- 保留 505 个原有未提交文件及原暂存分类。基线合并处理 19 个冲突文件，恢复定制处理 11 个冲突文件；原始 index、补丁、哈希清单及 stash 保留。原有业务定制没有整体提交。
- 整合 OpenCode 用量与 Codex 私有打票字段、账号测试列表、前端视频价格与推理倍率表单、简易模式窗口与原子用量结算；保留 Dola 30 秒、图片补偿、视频恢复退款及 OAuth 保护。
- 修正自动合并未发现的重复插件 provider、上游新增测试与本地构造参数/SQL 列数/函数返回值不兼容；Wire 重新生成。
- 修复本地推理参数归一化丢失新版 max/GPT-6 none 的问题，补覆盖平铺/嵌套及采样参数的回归；本地视频结算补入上游推理倍率，保留必须使用已验证时长的安全限制。

### Testing

- `cd frontend && pnpm test:run`：351 个文件、2652 项通过；`pnpm run lint:check`、`pnpm run build`（含翻译、类型检查）通过。保留已有 localStorage、Browserslist、大 chunk 提示。
- `cd backend && go test -mod=readonly -tags=unit ./...` 最终全部通过（211 秒）；首次编译及 6 项回归失败已修正并复测，详见各轮日志。
- `go test -mod=readonly -race -tags=unit ./internal/service ./internal/handler ./internal/repository ./internal/pkg/tlsfingerprint ./internal/pkg/apicompat -run 'Test(DolaVideo|ViraleeVideo|OpenAICodexTicket|CodexTicket|OpenAIOAuthProtection|AccountTraffic|LockAndMergeAccount|FetchOpenAIAccountModels|LocalTLS|ChatCompletionsToResponsesSyncedReasoning|.*SimpleMode.*|.*Reasoning.*Pricing.*|.*PendingImageSettlement.*|GPT6)' -count=1 -timeout=6m`：最终全部通过，无跳过 TLS 用例。
- 临时 Docker PostgreSQL/Redis：迁移幂等、239 推理倍率迁移、渠道/账号统计/分组推理价格读写与计费、提现幂等集成通过；另用 `-run '^TestAccountRepoSuite$/(TestUpdate$|TestBulkUpdate|TestUpdateExtra|TestUpdate_SyncSchedulerSnapshotOnCredentialsChange)'` 验证账号更新及批量更新通过。仅操作隔离测试数据库。
- `go generate ./cmd/server`、embed 生产编译及程序 `--version` 通过（0.2.8）。`git diff --check` 和 `git diff --cached --check` 通过。保留检查显示 missing、unexpectedChanges、stagingChanges 均为空。
- 证据目录 `/tmp/sub2api-sync-20260923.rOMJ0l/`：各轮测试、构建、版本日志、before.json、原 index 和补丁、上游补丁及恢复校验日志。本轮未做付费请求、生产端到端、镜像发布或部署。

### Notes

- 上游变更完整文件清单及每项操作见 `docs/upstream-sync-20260923-files.md`（473 项，含以下交集文件）；本轮只同步及必要兼容，不扩大到独立功能开发。
- `backend/cmd/server/wire_gen.go`、`backend/internal/service/wire.go`、`backend/internal/handler/wire.go`：合并账号服务装配，去除重复 provider 并重新生成依赖注入。
- `backend/internal/handler/admin/account_handler.go`、`backend/internal/handler/admin/setting_handler.go`、`backend/internal/handler/admin/setting_handler_update.go`、`backend/internal/handler/dto/mappers.go`、`backend/internal/handler/dto/settings.go`、`backend/internal/handler/dto/types.go`：并存打票、Claude 版本同步及 OpenCode 用量字段。
- `backend/internal/repository/account_repo.go`：合并受管用量字段与私有票据的锁定读取及保存。
- `backend/internal/repository/account_repo_upstream_billing_probe_update_test.go`、`backend/internal/repository/account_repo_opencode_go_usage_test.go`、`backend/internal/repository/account_repo_codex_ticket_test.go`：同步新增查询列的测试数据，保留票据脱敏及防伪验证。
- `backend/internal/service/domain_constants.go`、`backend/internal/service/setting_parse.go`、`backend/internal/service/setting_service.go`、`backend/internal/service/setting_update.go`、`backend/internal/service/settings_view.go`：合并独立配置项、缓存和更新流程。
- `backend/internal/handler/admin/channel_handler.go`、`frontend/src/components/admin/channel/PricingEntryCard.vue`：同时保留视频模式与上游推理等级倍率。
- `backend/internal/service/account.go`：保留本地原生图片账号类型约束，同时支持新增 API Key 图片能力。
- `backend/internal/service/account_test_service.go`、`backend/internal/service/account_test_models_test.go`：保留 API Key 配置模型的测试选择，与上游 OAuth 图片及透传列表规则兼容。
- `backend/internal/service/billing_cache_service.go`、`backend/internal/service/billing_service.go`、`backend/internal/service/gateway_usage_billing.go`、`backend/internal/service/openai_gateway_usage.go`：合并简易模式窗口与本地预冻结/补偿路径，保留视频时长要求，补入推理倍率。
- `backend/internal/pkg/apicompat/types.go`、`backend/internal/pkg/apicompat/chatcompletions_to_responses.go`、`backend/internal/pkg/apicompat/reasoning.go`、`backend/internal/pkg/apicompat/reasoning_sync_test.go`：保留嵌套推理字段、缓存参数及 max/GPT-6 none 转换，补对应测试。
- `backend/internal/service/gateway_image_reasoning_pricing_test.go`、`backend/internal/service/gateway_usage_billing_simple_mode_test.go`、`backend/internal/service/reasoning_effort_billing_test.go`、`backend/internal/service/billing_service_test.go`：适配本地计费返回值/视频前置约束及新增 Grok 4.7 支持。
- `backend/internal/repository/http_upstream_billing_lifecycle_test.go`、`backend/internal/server/routes/composite_images_compatible_test.go`：补本地视频依赖参数及打票后台清理。
- `.gitignore`、`frontend/src/api/admin/settings.ts`、`frontend/src/types/index.ts`、`frontend/src/views/admin/SettingsView.vue`：合并文档例外、后台配置和类型字段。
- `docs/UPSTREAM_SYNC.md`：追加使用变化、验证、新迁移及恢复步骤；`docs/upstream-sync-20260923-files.md`：新增文件清单；`progress.md`：追加本轮记录。
- 新增三项迁移且依赖清单未变；部署前备份数据库，旧 max 倍率会迁入推理等级映射，同数字前缀的本地迁移按完整文件名独立登记。源码恢复不自动撤销数据库变更。
- 回滚点：`codex/pre-upstream-sync-20260923` 与 stash `2daf8ddc0d2ff7aa5262b718c4d88156d696418f`。执行 `git worktree add -b codex/recover-pre-sync-20260923 /tmp/sub2api-pre-sync-20260923 codex/pre-upstream-sync-20260923`，然后 `git -C /tmp/sub2api-pre-sync-20260923 stash apply --index 2daf8ddc0d2ff7aa5262b718c4d88156d696418f` 可独立恢复同步前代码和全部定制；恢复两份文档 intent-to-add 的命令见同步记录。本轮未执行回滚。
- 本轮记录计数更正：定制恢复实际为 10 个冲突文件，前文 11 为笔误；基线 19 个冲突文件及 473 个上游变更文件的计数不变。


## 2026-09-24 - Task: 移除 Codex 292/332 打票并增加 OAuth 312 信号检测
### What was done
- 移除主动采集、续票、票据注入、缺票调度暂停、全局打票设置与账号 Pro/Team 打票选择；历史配置不再影响请求放行。
- 被动观察实际 OAuth HTTP 响应、账号测试、连接池新建 WS 和原生 WS 握手。只保存长度与观测时间，312 显示疑似降智；无有效头不清除旧状态，不把请求回带头或复用握手误当新检测。
- 账号列表用量区和编辑弹窗显示只读状态、最近检测时间及最后一次 312 时间；不根据检测结果停用账号，不主动请求上游。现有会话续接、TLS、线路保护、计费和其他渠道路径保留。
- 使用既有延迟服务按账号合并保存，数据库行锁下保留最新观测及最后一次 312；管理员编辑/导入不能伪造或覆盖受控字段。旧票据和代理密码继续脱敏。
### Testing
- 后端全量：`cd backend && go test -mod=readonly -tags=unit ./...` 通过（最终 222.466 秒）；旧设置契约断言已同步移除。
- 真实数据库：`CI=true SUB2API_TEST_POSTGRES_IMAGE=postgres:18-alpine go test -mod=readonly -tags=integration ./internal/repository -run '^TestAccountCodexSignalPersistenceConcurrentEdits$' -count=1` 通过；使用测试工具新建的隔离 PostgreSQL/Redis，未连接业务数据库。
- 定向竞态：`go test -mod=readonly -race -tags=unit ./internal/service ./internal/repository ./internal/handler/dto -run 'TestCodexSignal|TestAdminAccountPreservesCodexObservation|TestAccountCodexSignal|TestCodexProtection|TestOpenAICodexTurnState|TestLockAndMergeAccountExtra' -count=1` 通过，涵盖 HTTP/WS 正常响应、旧打票选择不阻断、重用握手不刷新、并发保存/失败重试与原有保护。
- 前端全量：`pnpm exec vitest run` 352 个文件、2645 项测试通过；`pnpm run lint:check`、`pnpm run build`（含类型与 i18n 检查）通过。生产构建仍有既有大 chunk 提示。
- `go generate ./cmd/server` 和前端最终产物生成后的 `go build -mod=readonly -tags=embed -o /tmp/sub2api-codex-signal-20260924.NhGvTh/sub2api ./cmd/server` 通过；`git diff --check` 通过。
- 验证记录位于 `/tmp/sub2api-codex-signal-20260924.NhGvTh/` 的 `backend-final.log`、`integration.log`、`signal-race.log`、`frontend-final.log`、`frontend-lint-final.log`、`frontend-build-final.log` 与 `backend-build-final.log`。未用真实官方账号验证“312 等同能力下降”，实现只承诺按长度显示观测信号。
### Notes
- 当前 scope 为退役打票及新增被动信号，不同步新的远程提交、不改版本号、不构建推送镜像、不部署、不修改生产数据库。原有未提交定制完整保留。
- 回滚快照：`/tmp/sub2api-codex-signal-20260924.NhGvTh/before.tgz`，原索引副本为同目录 `index`。执行 `git worktree add --detach /tmp/sub2api-before-codex-signal f867e4f9bd2f6bcf1b2c6a67b20d5124cc86748a`，再执行 `tar -xzf /tmp/sub2api-codex-signal-20260924.NhGvTh/before.tgz -C /tmp/sub2api-before-codex-signal` 可在独立目录恢复本轮前全部源码定制。回退旧服务前需关闭历史打票开关，避免旧版本重新拦截。
- 本轮改动文件（含删除的旧测试，均可从上述快照恢复）：
- `backend/cmd/server/wire.go`：移除打票后台生命周期和专用设置依赖，同步生成依赖注入代码。
- `backend/cmd/server/wire_gen.go`：移除打票后台生命周期和专用设置依赖，同步生成依赖注入代码。
- `backend/internal/config/config.go`：移除已退役打票的配置结构、默认值或部署示例。
- `backend/internal/handler/admin/account_codex_ticket_test.go`：删除仅服务于主动打票、续票、类型选择或旧状态界面的测试；替代观测逻辑由新增回归覆盖。
- `backend/internal/handler/admin/account_handler.go`：移除后台打票策略富化，账号状态改由 DTO 的观测摘要返回。
- `backend/internal/handler/admin/setting_handler.go`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `backend/internal/handler/admin/setting_handler_audit.go`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `backend/internal/handler/admin/setting_handler_codex_ticket_test.go`：删除仅服务于主动打票、续票、类型选择或旧状态界面的测试；替代观测逻辑由新增回归覆盖。
- `backend/internal/handler/admin/setting_handler_update.go`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `backend/internal/handler/dto/account_codex_signal_test.go`：验证账号详情和列表输出只读摘要且不泄露历史 state 或采集代理密码。
- `backend/internal/handler/dto/mappers.go`：将账号票据状态替换为只读 312 观测摘要，并保持列表/详情映射一致。
- `backend/internal/handler/dto/settings.go`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `backend/internal/handler/dto/types.go`：将账号票据状态替换为只读 312 观测摘要，并保持列表/详情映射一致。
- `backend/internal/handler/openai_viralee_videos_test.go`：删除构造服务时已不存在的打票采集器清理调用，保留原功能回归。
- `backend/internal/handler/wire.go`：移除打票后台生命周期和专用设置依赖，同步生成依赖注入代码。
- `backend/internal/repository/account_repo.go`：更新服务端观测与历史票据字段保留说明，沿用账号编辑行锁保护。
- `backend/internal/repository/account_repo_codex_signal.go`：新增账号行锁下的观测合并保存，保留较新结果及最后一次 312 时间。
- `backend/internal/repository/account_repo_codex_signal_integration_test.go`：新增真实 PostgreSQL 并发观测、乱序写入和管理员旧快照保存的回归验证。
- `backend/internal/repository/account_repo_codex_ticket_test.go`：保留历史私有字段合并测试，将已退役配置常量改为历史键字面量。
- `backend/internal/repository/http_upstream_billing_lifecycle_test.go`：删除构造服务时已不存在的打票采集器清理调用，保留原功能回归。
- `backend/internal/repository/scheduler_cache.go`：移除打票选择字段的调度元数据投影及其专用测试。
- `backend/internal/repository/scheduler_cache_test.go`：移除打票选择字段的调度元数据投影及其专用测试。
- `backend/internal/server/api_contract_test.go`：同步系统设置接口契约，删除已退役打票响应字段断言。
- `backend/internal/server/routes/composite_images_compatible_test.go`：删除构造服务时已不存在的打票采集器清理调用，保留原功能回归。
- `backend/internal/service/admin_account.go`：删除已退役打票类型的校验，保留服务端受控字段的创建与更新保护。
- `backend/internal/service/admin_account_codex_signal_test.go`：验证账号创建和编辑不能伪造观测结果或覆盖服务器保留的历史私有字段。
- `backend/internal/service/admin_account_codex_ticket_test.go`：删除仅服务于主动打票、续票、类型选择或旧状态界面的测试；替代观测逻辑由新增回归覆盖。
- `backend/internal/service/deferred_service.go`：接入已有延迟更新周期，合并保存观测而不在请求路径写数据库。
- `backend/internal/service/domain_constants.go`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `backend/internal/service/openai_account_runtime_block_fastpath.go`：移除缺票导致的调度拦截，保留原有限流和冷却判定。
- `backend/internal/service/openai_codex_signal.go`：新增上游响应头被动识别、只读摘要、按账号合并及延迟保存重试。
- `backend/internal/service/openai_codex_signal_test.go`：验证 HTTP、账号测试、连接池/原生 WS、缺失信号、并发合并和失败重试。
- `backend/internal/service/openai_codex_ticket.go`：移除主动打票、存取票据、续票和缺票门控，仅保留历史字段脱敏及只读状态合并保护。
- `backend/internal/service/openai_codex_ticket_lifecycle_test.go`：删除仅服务于主动打票、续票、类型选择或旧状态界面的测试；替代观测逻辑由新增回归覆盖。
- `backend/internal/service/openai_codex_ticket_protection_test.go`：删除仅服务于主动打票、续票、类型选择或旧状态界面的测试；替代观测逻辑由新增回归覆盖。
- `backend/internal/service/openai_codex_ticket_selection_test.go`：删除仅服务于主动打票、续票、类型选择或旧状态界面的测试；替代观测逻辑由新增回归覆盖。
- `backend/internal/service/openai_codex_ticket_test.go`：删除仅服务于主动打票、续票、类型选择或旧状态界面的测试；替代观测逻辑由新增回归覆盖。
- `backend/internal/service/openai_codex_ticket_type_test.go`：删除仅服务于主动打票、续票、类型选择或旧状态界面的测试；替代观测逻辑由新增回归覆盖。
- `backend/internal/service/openai_codex_turn_state.go`：移除过时的打票注入注释，保留原有回合透传和跨账号回带保护。
- `backend/internal/service/openai_gateway_forward.go`：移除请求发送前的票据注入与缺票失败，保留原认证、会话和转发路径。
- `backend/internal/service/openai_gateway_messages.go`：移除请求发送前的票据注入与缺票失败，保留原认证、会话和转发路径。
- `backend/internal/service/openai_gateway_passthrough.go`：移除请求发送前的票据注入与缺票失败，保留原认证、会话和转发路径。
- `backend/internal/service/openai_gateway_service.go`：删除打票内存状态与后台采集器启动。
- `backend/internal/service/openai_plugin_transport.go`：在实际 HTTP 与账号测试成功响应后采集信号，保持响应体和转发行为。
- `backend/internal/service/openai_viralee_video_test.go`：删除构造服务时已不存在的打票采集器清理调用，保留原功能回归。
- `backend/internal/service/openai_ws_forwarder.go`：为运行期连接池绑定上游握手观测回调。
- `backend/internal/service/openai_ws_forwarder_payload.go`：移除请求发送前的票据注入与缺票失败，保留原认证、会话和转发路径。
- `backend/internal/service/openai_ws_pool.go`：在新建成功握手后通知观测回调，不把连接复用视为新检测。
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`：接入原生 WebSocket 成功握手的响应头检测。
- `backend/internal/service/setting_codex_ticket_test.go`：删除仅服务于主动打票、续票、类型选择或旧状态界面的测试；替代观测逻辑由新增回归覆盖。
- `backend/internal/service/setting_gateway_runtime.go`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `backend/internal/service/setting_parse.go`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `backend/internal/service/setting_service.go`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `backend/internal/service/setting_update.go`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `backend/internal/service/settings_view.go`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `deploy/config.example.yaml`：移除已退役打票的配置结构、默认值或部署示例。
- `docs/OPENAI_CODEX_SIGNAL.md`：新增 312 检测用法、信号边界、持久化说明和独立恢复步骤。
- `docs/OPENAI_CODEX_TICKETS.md`：新增退役说明，保留历史合并和旧版本操作记录供追溯。
- `frontend/src/api/admin/settings.ts`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `frontend/src/components/account/AccountUsageCell.vue`：将账号用量区域的票据倒计时改为被动信号状态并更新交互验证。
- `frontend/src/components/account/CodexSignalBadge.vue`：新增只读信号提示、最近观测时间和历史 312 悬停说明，仅用于 OpenAI OAuth 类账号。
- `frontend/src/components/account/CreateAccountModal.vue`：移除新建账号打票参与开关、提交字段及其旧测试。
- `frontend/src/components/account/EditAccountModal.vue`：移除账号打票开关和 Pro/Team 选择，编辑页显示只读信号摘要并清理旧交互测试。
- `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts`：将账号用量区域的票据倒计时改为被动信号状态并更新交互验证。
- `frontend/src/components/account/__tests__/CodexSignalBadge.spec.ts`：验证未知、312、非 312 的颜色文案和不适用账号隐藏。
- `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts`：移除新建账号打票参与开关、提交字段及其旧测试。
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`：移除账号打票开关和 Pro/Team 选择，编辑页显示只读信号摘要并清理旧交互测试。
- `frontend/src/i18n/locales/en/admin/accounts.ts`：移除打票文案，新增中英检测状态、时间和能力判定限制提示。
- `frontend/src/i18n/locales/en/admin/settings.ts`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `frontend/src/i18n/locales/zh/admin/accounts.ts`：移除打票文案，新增中英检测状态、时间和能力判定限制提示。
- `frontend/src/i18n/locales/zh/admin/settings.ts`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `frontend/src/types/index.ts`：将账号票据状态替换为只读 312 观测摘要，并保持列表/详情映射一致。
- `frontend/src/views/admin/SettingsView.vue`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`：移除打票总开关、采集代理及对应缓存、设置接口、页面和旧测试。
- `progress.md`：仅在末尾追加本轮结果、验证、文件清单和回滚点。

## 2026-09-25 - Task: 增加 OpenAI OAuth Basispoints 可选上游
### What was done
- 在账号编辑页增加 OpenAI OAuth/Setup Token 的 Responses 上游选择，默认保持 ChatGPT Codex，选择 Basispoints 后仅 HTTP Responses 使用 Basispoints 地址。
- Basispoints 请求补充账号身份头，并优先使用账号保存的 `chatgpt_account_id`；缺失时尝试从 access token 的 OpenAI auth claim 提取，冲突则拒绝发送。
- Basispoints 上游请求中的 `max` 推理强度自动降级为 `xhigh`；默认 ChatGPT Codex、API Key、WebSocket、图片和其他 OpenAI 路径保持原逻辑。
### Testing
- `go test ./internal/service -run 'Test(BuildUpstreamRequestBasispointsUsesEndpointAndDowngradesMax|OpenAIOAuthResponsesEndpointModeDefaultsToChatGPT|NormalizeOpenAIBasispointsReasoningBodyDowngradesMax|ApplyOpenAIBasispointsHeadersUsesStoredAccountID|OpenAIChatGPTAccountIDFromAccessToken)' -count=1` 通过。
- `go test ./internal/service -count=1` 通过（135.550s）。
- `pnpm exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts` 通过（76 tests）。
- `pnpm exec vitest run src/i18n/__tests__/localeKeyCompleteness.spec.ts` 与 `pnpm run typecheck` 通过。
- `git diff --check` 通过；未使用真实 OAuth Token 请求上游。
### Notes
- 改动文件：`backend/internal/service/openai_basispoints.go`、`backend/internal/service/openai_basispoints_test.go`、`backend/internal/service/openai_gateway_service.go`、`backend/internal/service/openai_gateway_forward.go`、`backend/internal/service/openai_gateway_passthrough.go`、`frontend/src/components/account/EditAccountModal.vue`、`frontend/src/components/account/__tests__/EditAccountModal.spec.ts`、`frontend/src/i18n/locales/zh/admin/accounts.ts`、`frontend/src/i18n/locales/en/admin/accounts.ts`、`docs/OPENAI_OAUTH_BASISPOINTS.md`。其中部分文件在本轮前已有未提交改动，未覆盖或清理这些既有改动。
- 回滚方式：在账号编辑页将所有账号恢复为“ChatGPT Codex（默认）”，或删除 `accounts.extra.openai_oauth_responses_endpoint`；代码回滚时删除新增的 `openai_basispoints.go`、`openai_basispoints_test.go`、`docs/OPENAI_OAUTH_BASISPOINTS.md`，并移除上述请求构建、账号编辑和文案中的 Basispoints 增量。未构建、部署或推送镜像。

## 2026-09-25 - Task: 收敛 Basispoints 影响范围
### What was done
- 修正图片 OAuth 自建请求复用通用请求构建器时的边界：Basispoints 地址、身份头和 `max → xhigh` 只对 HTTP Responses 生效，不再混入图片请求。
- 保留 OAuth/Setup Token 的默认 ChatGPT Codex 行为及其他 OpenAI 路径不变。
### Testing
- 新增并通过 Basispoints 图片隔离回归测试；定向 Basispoints 后端测试全部通过。
- `git diff --check` 通过。
### Notes
- 改动文件：`backend/internal/service/openai_basispoints.go`、`backend/internal/service/openai_basispoints_test.go`、`backend/internal/service/openai_gateway_forward.go`、`backend/internal/service/openai_gateway_passthrough.go`；仅收紧本轮新增路径的图片边界。
- 回滚方式：恢复上述四个文件到本轮前版本；账号配置仍可通过选择“ChatGPT Codex（默认）”回滚。未构建、部署或推送镜像。

## 2026-09-25 - Task: Basispoints 回归复核
### What was done
- 对账号级 Basispoints 路由、请求头、推理强度降级、默认模式和图片隔离进行最终复核。
### Testing
- `go test ./internal/service -count=1` 通过（136.038s）。
### Notes
- 本轮未新增源码文件；复核范围为前述 Basispoints 实现及其图片边界修正。未构建、部署或推送镜像。

## 2026-09-25 - Task: 构建并推送双架构 china-api 镜像
### What was done
- 基于当前工作区代码构建并推送 `iotwq/china-api:latest`，目标平台为 `linux/amd64` 与 `linux/arm64`。
- 远端镜像 manifest digest 为 `sha256:ec0d3ea3312d6cace6144af26b5f8724373d0af274e6c006c44b80212d8ed97f`。
### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 -t iotwq/china-api:latest --push .` 通过。
- `docker buildx imagetools inspect iotwq/china-api:latest` 确认 manifest 包含 `linux/amd64` 与 `linux/arm64`。
- `docker manifest inspect iotwq/china-api:latest` 再次确认两个架构的子 manifest 均已发布。
### Notes
- 改动文件：`progress.md`，追加本次构建、推送及双架构验证记录；工作区其他既有修改未覆盖或清理。
- 回滚方式：部署时继续使用上一版本的镜像 digest 或固定 tag；本次未修改源码、配置或数据库。

## 2026-09-25 - Task: 修复 Basispoints 请求体 422 和工具协议兼容
### What was done
- 对照 ranxi2001/sub2api 的固定提交 `6b0c0ddbd1649caad5d980e92368059b1a5d1158`，确认旧实现只改 URL/身份头，仍发送 Codex 请求体，缺少 Basispoints 的顶层推理参数、模型选择和工具协议。
- 为显式选择 Basispoints 的 OAuth/Setup Token HTTP Responses 增加独立适配，保留 `max → xhigh`，支持文本、客户端工具、完整多轮历史、compact 及流式/非流式回传，普通账号测试同步接入。
- 保留实际账号代理、并发控制、取消、限流重试和用量记账入口；将请求强度和实际强度分别记录。选定工作区优先于 token 默认工作区，由上游认证成员关系。
- 默认 ChatGPT Codex、API Key、WebSocket、独立图片/视频路径保持原行为；只移植协议模块，未引入参考项目的公开图片中继、全局配置、账户池或部署改动。
### Testing
- 修复前新增 `TestBasispointsForwardWireProtocol`，两个流式/非流式用例均在顶层 `reasoning_effort` 缺失处稳定失败；修复后通过。
- `go test ./internal/service/basispoints -count=1` 通过（84 个顶层协议测试及其子用例）。
- `go test -tags unit ./internal/service ./internal/service/basispoints -count=1` 通过（服务层 191.176s）。
- `go test -tags unit ./internal/service -run 'Basispoints|OpenAIOAuthResponsesEndpoint|OpenAIChatGPTAccountIDFromAccessToken' -count=1` 在最终无用量失败结果调整后通过（1.134s）。
- 定向 `go test -race` 通过，覆盖 Basispoints、工具多轮回传、历史缓存、并发保护及取消；未发现数据竞态。
- 前端账号编辑与语言完整性验证共 79 项通过，`pnpm run typecheck` 通过。
- 后端 `go build -o <临时目录>/sub2api ./cmd/server` 通过，`git diff --check` 通过。
- 未使用真实生产 OAuth Token 调用 Basispoints；用户补充的 `422: Invalid request body.` 与源码中确认的协议缺口一致，但线上成功仍需部署后实测。
### Notes
- 回滚方式：在“账号管理 → 编辑账号”将 `OAuth Responses 上游地址` 改回 `ChatGPT Codex（默认）`，然后新建会话；不需数据库迁移。完整源码回滚不得 reset 整个脏工作区，应仅撤销下列本轮 Basispoints 增量。
- 本轮未构建或推送 Docker 镜像、未部署、未改变生产账户/计费数据；先前未提交修改全部保留。
- 改动文件：
- `backend/go.mod`：增加协议结构化输出验证依赖 jsonschema/v6 v6.0.3。
- `backend/go.sum`：记录新依赖的校验和。
- `backend/internal/service/openai_basispoints.go`：替换 URL 直改逻辑为独立路由条件和 Basispoints 请求头构造，按选定工作区优先解析身份。
- `backend/internal/service/openai_basispoints_forward.go`：新增请求及 SSE 协议桥接、工具历史隔离、compact 转换、用量接入和账号测试复用。
- `backend/internal/service/openai_basispoints_test.go`：新增请求格式、路由隔离、工具往返、错误状态、计费结果、账号测试和并发取消回归。
- `backend/internal/service/openai_gateway_forward.go`：在 Codex 通用转换前分流选定 Basispoints 的 HTTP 请求，移除通用构建器中的旧 URL/请求头改写。
- `backend/internal/service/openai_gateway_passthrough.go`：移除旧 Basispoints 透传改写，由专用适配统一处理。
- `backend/internal/service/account_test_service.go`：普通账号文本测试按 Basispoints 配置使用同一实际协议。
- `frontend/src/i18n/locales/zh/admin/accounts.ts`：更新账号配置说明及图片输入/切换会话限制。
- `frontend/src/i18n/locales/en/admin/accounts.ts`：同步英文账号配置说明。
- `docs/OPENAI_OAUTH_BASISPOINTS.md`：说明 422 原因、完整协议适配、边界、固定参考提交和配置回滚步骤。
- `backend/internal/service/basispoints/NOTICE.md`：保留原作者归属并记录本次固定移植来源与范围。
- `backend/internal/service/basispoints/agent_message_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/basispoints_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/catalog.go`：移植工具参数说明生成。
- `backend/internal/service/basispoints/catalog_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/content.go`：移植输入内容类型验证和不包含私密内容的诊断。
- `backend/internal/service/basispoints/content_diagnostics_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/custom_transport.go`：移植 raw custom 工具传输解析。
- `backend/internal/service/basispoints/custom_transport_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/deep_audit_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/deep_envelope_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/direct_call_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/envelope.go`：移植工具 JSON 封装解析和有限格式兼容。
- `backend/internal/service/basispoints/envelope_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/history_collision_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/history_recovery_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/images.go`：移植 HTTPS 图片引用校验；不包含公开图片中继。
- `backend/internal/service/basispoints/images_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/invocation_recovery_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/plan.go`：移植客户端声明的 update_plan 工具兼容。
- `backend/internal/service/basispoints/plan_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/raw_custom_call_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/request.go`：移植 Basispoints 请求白名单、顶层推理强度、指令转换与会话元数据。
- `backend/internal/service/basispoints/request_compatibility_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/route.go`：移植生图及指定原生搜索的原线路选择，调整图片中继注释。
- `backend/internal/service/basispoints/route_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/stream.go`：移植增量 SSE 转换、终态验证与响应体取消/关闭。
- `backend/internal/service/basispoints/structured_output.go`：移植结构化输出的最终验证，禁止外部 schema 加载。
- `backend/internal/service/basispoints/structured_output_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/terminal_tools_test.go`：移植对应协议回归测试，验证工具、历史、数据格式或流处理行为。
- `backend/internal/service/basispoints/tools.go`：移植客户端工具目录、受限历史缓存、返回调用还原与多轮回传。
- `progress.md`：仅追加本轮修复和验证记录。

## 2026-09-25 - Task: 构建并推送 Basispoints 修复双架构镜像
### What was done
- 基于当前工作区最新代码（包括 Basispoints 422 协议适配修复）构建并推送 `iotwq/china-api:latest`，包含 `linux/amd64` 与 `linux/arm64`，版本保持源码中的 `0.2.8`。
- 远端镜像索引摘要：`sha256:a3c75ba76f1bdb3559b7c0ea9b1d52a6f2c589503af2f30781cc806daee350bf`。
### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 -t iotwq/china-api:latest --push .` 通过，包含前端语言完整性检查、类型检查、生产构建及两架构后端编译。
- `docker buildx imagetools inspect iotwq/china-api:latest` 确认远端摘要与构建输出一致，包含 amd64 子摘要 `sha256:9cc0991e45506024215ae4411f8d0ec825a36fbc4f568ff7946fd8885d769128` 和 arm64 子摘要 `sha256:a6120f027854662912b7c113c8d3edf28f25f16ca77289ee5d5de210a3be302c`。
- amd64 按索引摘要拉取并执行 `/app/sub2api -version` 通过；arm64 拉取遇到 Docker Hub EOF 后重试，按对应子摘要拉取并执行版本命令通过。两者均输出 `Sub2API 0.2.8`，退出码为 0。
- 镜像运行验证仅执行版本命令，未连接数据库或启动正式服务；本轮未重复前述已通过的功能回归，也未使用真实 OAuth Token 验证上游。
### Notes
- 改动文件：`progress.md`，仅在末尾追加本轮构建、推送、双架构验证和回滚记录；保留全部既有工作区修改。
- 回滚方式：将部署 Compose 中服务的 `image` 固定为 `iotwq/china-api@sha256:ec0d3ea3312d6cace6144af26b5f8724373d0af274e6c006c44b80212d8ed97f`，然后执行 `docker compose pull sub2api` 和 `docker compose up -d --force-recreate sub2api`。
- 本轮仅发布镜像，未部署、重启线上服务或修改生产数据。

## 2026-09-26 - Task: 逐项核对 v2.8.13 Basispoints 修复并补齐推理等级默认展示
### What was done
- 核验 `v2.8.13` 标签确实指向上次协议移植的 `6b0c0ddbd1649caad5d980e92368059b1a5d1158`，澄清发布版累计修复与该快照之后增量提交的区别。
- 确认 #73 工具解析/大整数/安全诊断及 #78 子代理明文修复已在本地；依据 #79 的 `beb86d6caf53709ff386485422ba2a32d1768eed` 补齐管理员用量页推理等级默认可见，保留已有列设置。
- 增加原始推理等级经过分组策略、Basispoints 适配和用量入账后仍被正确保留的回归；明确“缓存创建转普通输入”可选计费开关未移植，不改变当前计费和下游用量。
### Testing
- 先移植上游真实页面回归：修复前 4 项中 3 项失败，原因是推理等级列默认隐藏；应用一行生产修复后页面集成测试 4 项全部通过。
- `pnpm exec vitest run src/__tests__/integration/usage-reasoning-effort.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts`：33 项通过，包含默认映射展示、保存的隐藏设置与重新开启、未映射单值及普通用户边界。
- `go test -tags unit ./internal/service -run 'Basispoints|OpenAIOAuthResponsesEndpoint|OpenAIChatGPTAccountIDFromAccessToken' -count=1` 通过（0.901s）；新增 8 个流式/非流式用例覆盖 max 降为 xhigh、分组降为 xhigh/high 及原值不变，并核对写入的原始/实际强度。
- `go test ./internal/service/basispoints -count=1` 全量通过（0.406s），涵盖 #73 和 #78 的协议回归。
- `pnpm run typecheck` 与本轮两个前端文件的定向 ESLint 检查通过。
### Notes
- 改动文件：`frontend/src/views/admin/UsageView.vue`：仅从默认隐藏列移除推理等级，不覆盖用户已保存偏好。
- 改动文件：`frontend/src/__tests__/integration/usage-reasoning-effort.spec.ts`：移植 #79 的真实页面展示与隐藏偏好回归。
- 改动文件：`backend/internal/service/openai_basispoints_test.go`：增加分组映射前后推理等级贯穿转发与入账的验证，未改后端生产逻辑。
- 改动文件：`docs/OPENAI_OAUTH_BASISPOINTS.md`：记录四项发布修复的实际集成状态及使用边界。
- 改动文件：`progress.md`：仅在末尾追加本轮记录。
- 回滚方式：将管理员页面 `DEFAULT_HIDDEN_COLUMNS` 恢复包含 `reasoning_effort` 即可撤回运行行为；完整本轮前文件快照位于 `/tmp/sub2api-bps-release-audit.DooygE/before.tar`，可用 `tar -xf /tmp/sub2api-bps-release-audit.DooygE/before.tar -C /Users/wangqiang/Project/codex/sub2api-wq` 恢复这五个文件（仅适用于这些文件没有后续新改动时，不得覆盖后续工作）。
- 未引入批量设置、请求正文采集、凭证守护或可选缓存计费策略；未修改数据库、版本和生产数据，未构建、推送或部署。

## 2026-09-26 - Task: 补齐仅 Basispoints 生效的缓存创建转普通输入
### What was done
- 按用户授权补齐 PR #73 中 49bb7b049e1eea21f023b319eaacfaeed467331a 的下游用量同步逻辑及本地缺失的可选计费开关；账号编辑选择 Basispoints 后显示该开关，默认关闭，保存后仅影响后续请求。
- 同时校验账号选择的模式、开关布尔值及当前请求实际端点；默认 Codex、原生回退、WebSocket、API Key 和其他渠道保持原计费及返回用量。切回默认 Codex 并保存时移除该配置，避免隐藏开关残留生效。
- 开启后缓存创建按普通输入计费，后台记录与下游流式/非流式 usage 同步清除创建/写入量及 TTL 分项；缓存读取、总输入和输出保持不变，内部转发结果保留上游原始计数，不将创建伪装成读取。
### Testing
- 先增加记账回归，确认实现接入前“Basispoints 开启”用例稳定失败：期望普通输入 900，实际仍为 700。接入计费转换后通过，其余默认模式和原生回退对照用例保持通过。
- 新增回归覆盖 OAuth/Setup Token、开关开启/关闭、Responses/compact、流式/非流式、直接 JSON、默认 Codex 与原生回退隔离、模型别名保留、所有缓存创建别名与 TTL 分项、大整数及非用量正文保留；验证返回 usage、后台 token 分桶和费用一致，上游原始计数不被改写。
- `go test -tags unit ./internal/service ./internal/service/basispoints -count=1` 全量通过：service 191.054s，basispoints 0.519s。
- `go test -race -tags unit ./internal/service -run 'Test(BasispointsCacheCreationAsInput|NormalizeOpenAIBasispointsUsage|BasispointsForwardWireProtocol|OpenAIGatewayServiceRecordUsage_GPT56SeparatesCacheWriteForBillingAndStats)' -count=1` 通过（2.183s）。
- `pnpm exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts src/i18n/__tests__/localeKeyCompleteness.spec.ts`：84 项通过，包含默认关闭、保存与回填、关闭、切回 Codex 清理、API Key 不展示和切换账号不串状态。
- `pnpm run typecheck`、四个前端修改文件的定向 ESLint、Go 文件格式检查及 `git diff --check` 通过。
- 验证使用本地模拟上游和账务替身；未使用真实 OAuth Token 调用上游或连接生产数据库，不能据此断言上游实际缓存命中率提高。
### Notes
- `backend/internal/service/openai_basispoints_usage.go`：新增实际端点与账号开关双重限定，移植仅清除下游缓存创建计数的转换函数。
- `backend/internal/service/openai_basispoints_usage_test.go`：增加请求转发至下游/记账闭环、费用、格式及默认线路隔离回归。
- `backend/internal/service/openai_gateway_response_handling.go`：在已采集原始 usage 的响应路径中，仅对符合条件的 Basispoints 下游响应同步用量。
- `backend/internal/service/openai_gateway_usage.go`：仅在符合条件时将缓存创建纳入普通输入，并同步计费与用量记录；保留所有既有视频及其他计费改动。
- `frontend/src/components/account/EditAccountModal.vue`：增加 Basispoints 专属开关、默认关闭、状态回填/重置和切回默认时的配置清理。
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`：验证账号编辑真实交互及默认 Codex 隔离。
- `frontend/src/i18n/locales/zh/admin/accounts.ts`：增加开关中文名称与计费/使用边界提示。
- `frontend/src/i18n/locales/en/admin/accounts.ts`：同步英文名称与提示。
- `docs/OPENAI_OAUTH_BASISPOINTS.md`：补充开启方式、用量计算示例、实际端点限定、PR #73 覆盖状态及回滚说明。
- `progress.md`：仅在末尾追加本轮记录，不改写既有历史。
- 回滚方式：在账号管理编辑该账号，关闭“缓存创建转普通输入”并保存；或将“OAuth Responses 上游地址”切回“ChatGPT Codex（默认）”并保存。无需数据库迁移或覆盖工作区文件，历史账单不变。
- 本轮未修改历史余额或账务数据，未更改版本，未构建、推送或部署镜像；保留全部既有未提交改动。

## 2026-09-26 - Task: 渠道监控智力结果展示实际糖果题答案
### What was done
- 在现有判分完成后保存实际解析出的最终数字，用户端、管理员端、历史检测条与手动检测弹窗统一显示“智力合格/不合格 · 糖果题答案：实际数字 · 检测时间”；不再用“最终答案不正确”替代答案。
- 复用既有 intelligence JSONB 中的可选 answer 字段，无需数据库结构迁移；旧记录未保存答案时明确显示“未记录”，不根据合格状态推测答案。
- 仅保存解析出的最终数字，不存储完整回答或推理；未完成/未检测原因、红绿状态、参考答案、low 推理强度、渠道状态与延迟计算保持原样。
### Testing
- 先补回归并确认红灯：后端答案保存测试的 5 个有效数字用例失败；前端 6 项展示测试中 4 项失败。实现后对应回归通过，覆盖 21、29、0、前导零、重复答案、无最终答案、截断响应和不保存正文。
- `go test -tags unit ./internal/service ./internal/handler ./internal/handler/admin -run 'Test.*(ChannelMonitor|Monitor|Challenge|RunCheck)' -count=1` 通过：service 8.534s，handler 3.660s，admin 0.629s；验证多协议答案采集、公开 DTO、历史聚合和普通探活延迟隔离。
- `go test -tags integration ./internal/repository -run '^TestChannelMonitorIntelligenceRoundTrip$' -count=1 -v` 通过（5.108s）；使用本地临时 PostgreSQL/Redis 容器，验证正确/错误答案随 JSONB 入库并在历史、最新结果与时间线回读，未连接生产数据库。
- 前端 MonitorIntelligence、MonitorFormDialog.accountSelector、MonitorPrimaryModelCell、MonitorCard.quota、localeKeyCompleteness 共 5 个测试文件、28 项测试通过，覆盖实际答案、旧记录、红绿颜色、无障碍提示与检测时间。
- `pnpm run typecheck`、5 个本轮前端修改文件的定向 ESLint、定向 Go 格式化及 `git diff --check` 通过。
- 本轮独立补丁 `git apply --reverse --check /tmp/sub2api-candy-answer.mwHvT7/answer-display.patch` 校验通过；没有实际执行回滚。
### Notes
- `backend/internal/domain/channel_monitor_intelligence.go`：为结果增加可选的最终数字 answer 字段。
- `backend/internal/service/channel_monitor_intelligence.go`：仅在得到唯一有效最终答案后保存该数字，不改变判分流程。
- `backend/internal/service/channel_monitor_intelligence_test.go`：增加最终答案序列化回归，并验证多协议及用户视图聚合保留答案。
- `backend/internal/handler/channel_monitor_intelligence_test.go`：验证公开列表、详情及历史 DTO 保留答案。
- `backend/internal/repository/channel_monitor_intelligence_integration_test.go`：验证正确和错误答案的真实数据库往返。
- `frontend/src/api/admin/channelMonitor.ts`：同步可选 answer 类型，兼容旧记录和旧服务响应。
- `frontend/src/composables/useMonitorIntelligence.ts`：统一合格/不合格的答案与时间提示，保留未完成原因。
- `frontend/src/components/user/__tests__/MonitorIntelligence.spec.ts`：增加用户历史与管理员弹窗答案展示、旧记录及灰色状态回归。
- `frontend/src/i18n/locales/zh/dashboard.ts`：增加中文实际答案与未记录提示，移除不再使用的泛化错误文案。
- `frontend/src/i18n/locales/en/dashboard.ts`：同步英文答案与未记录提示。
- `docs/channel-monitor-intelligence.md`：说明展示示例、历史边界、JSONB 兼容性及本轮回滚方式。
- `progress.md`：仅追加本轮实现与验证记录。
- 回滚方式：在仓库根目录先执行 `git apply --reverse --check /tmp/sub2api-candy-answer.mwHvT7/answer-display.patch`，确认无后续重叠改动后执行 `git apply --reverse /tmp/sub2api-candy-answer.mwHvT7/answer-display.patch`。只回退本轮源码、测试和文档；保留既有本地定制、历史日志和已保存的 JSONB 字段，无需数据库回滚。
- 本轮未更改判题网络请求、并发、计费、版本或线上数据；未使用真实上游进行检测，未构建、推送或部署镜像。

## 2026-09-26 - Task: 构建并推送最新 Basispoints 与糖果答案展示双架构镜像
### What was done
- 按北京时间记录本次发布；基于当前工作区（HEAD `f867e4f9b` 及既有未提交改动）构建并推送 `iotwq/china-api:latest`，包含近期 Basispoints 修改、仅 Basispoints 生效的缓存创建转普通输入及糖果题实际答案展示。
- 使用现有 `codex-multiarch` builder 和 Dockerfile 构建 `linux/amd64`、`linux/arm64`，版本保持 `0.2.8`，未修改源码、构建配置或版本号。
- 新远端镜像索引摘要：`sha256:6a8856ba91fbdea4056926ba344e1a41b266ed7df8efebb37509d6bf47e51594`。
### Testing
- `docker buildx build --builder codex-multiarch --platform linux/amd64,linux/arm64 -t iotwq/china-api:latest --push .` 成功，退出码 0；前端语言检查 3 项、类型检查、Vite 生产打包及两种架构的 Go embed 编译均完成。
- `docker buildx imagetools inspect iotwq/china-api:latest` 确认远端索引与构建输出一致，包含 `linux/amd64` 和 `linux/arm64`；其余 `unknown/unknown` 项为构建证明，不是额外运行架构。
- amd64 子摘要：`sha256:2cf994105cce50e88fd82084337b1c6269c963e6bfe8ceb93570185b71c09cec`；arm64 子摘要：`sha256:52a58f63aa0a5c9115249bc3d22b6add703485a0da2c9e0e91934325286db35f`。
- 从 Docker Hub 按新固定摘要拉取，并分别运行 `docker run --rm --network none --platform linux/amd64 --entrypoint /app/sub2api iotwq/china-api@sha256:6a8856ba91fbdea4056926ba344e1a41b266ed7df8efebb37509d6bf47e51594 -version` 与 arm64 对应命令；两者均正常退出并报告 `Sub2API 0.2.8`，构建时间 `2026-09-25T18:03:38Z`。
- arm64 第一次拉取遇到 Docker Hub registry EOF；原命令重试后成功，未修改认证、源码或镜像。前端 Browserslist 数据和较大分包提示属于非阻断警告。
- 冒烟容器禁用网络、未挂载数据目录，仅执行版本检查；未连接生产数据库，不代表完成真实上游业务请求验证。
### Notes
- `progress.md`：仅在末尾追加本轮发布摘要、双架构验证和回滚信息，保留所有既有工作区改动。
- 回滚点为本轮推送前的 `iotwq/china-api@sha256:a3c75ba76f1bdb3559b7c0ea9b1d52a6f2c589503af2f30781cc806daee350bf`。需要回退时，将部署 Compose 的 `sub2api.image` 固定为该完整引用，再执行 `docker compose pull sub2api` 和 `docker compose up -d --force-recreate sub2api`；本轮未执行部署或回滚。
- 本轮只完成镜像构建、推送和本地隔离验证，未重启线上服务、修改配置或生产数据。


## 2026-09-26 - Task: 补齐 Basispoints 内嵌图片中转及后台配置
### What was done
- 用户确认后，按参考项目 v2.8.12 的图片中转实现补齐本地缺失模块；将已鉴权 Basispoints HTTP 请求中的 Base64/data URL 图片暂存为本服务短期 HTTPS 链接，再进入原协议转换与转发路径。
- 在“系统设置 → 功能开关”增加默认关闭的 Basispoints 图片中转开关与公网 HTTPS 根地址；使用现有 settings 键值配置，无数据库迁移，未传字段保留现有值。
- 增加私有文件权限、格式校验、单图/单请求大小上限、每实例磁盘配额、30 分钟有效期、过期/停机清理、读取并发和额外转换内存预算；只有 GET/HEAD 能力链接，不提供匿名上传，常规访问日志不记录 token。
- 未移植参考项目会影响普通 OpenAI 请求的全局准入中间件；保护仅挂在实际 Basispoints 分支并要求中转开启。默认 Codex、API Key、WS、独立图片/视频、原生能力回退及既有计费逻辑保持原样。
### Testing
- `go test ./internal/service/basispoints -run ImageRelay -count=1` 通过；`go test -race ./internal/service/basispoints -count=1` 通过（6.209s），覆盖磁盘、格式、配额、TTL、并发与协议转换。
- `go test -tags unit ./internal/service ./internal/handler/admin ./internal/server/routes ./internal/server/middleware -run 'Basispoints|BPSImage|ImageRelay' -count=1` 通过；`go test -race -tags unit ./internal/service -run 'BasispointsImage|ExcelBPSImage' -count=1` 在最终转发代码上通过（2.326s），验证模拟上游收到 HTTPS 链接、按链接读回原始图片、默认线路未改写、流式/非流式用量保持及取消/错误后释放预算。
- 扩大至 `go test -tags unit ./internal/service/... ./internal/handler/... ./internal/server/... ./cmd/server`：service 全量（197.107s）、handler 全量（52.286s）、admin、dto、quotaview、routes、middleware、cmd/server 通过。首次仅 server 的两处管理员设置契约预期缺少新增默认字段；补齐断言后 `go test -tags unit ./internal/server/... ./cmd/server -count=1` 全部通过。
- 前端 SettingsView、EditAccountModal、localeKeyCompleteness 最终共 129 项测试通过；`pnpm run typecheck`、5 个本轮前端文件定向 ESLint、`pnpm exec vite build` 通过（10.78s）。保留既有 router-link/浏览器数据及较大分包非阻断提示，未为此扩展清理范围。
- 定向 Go 格式检查及 `git diff --check` 通过。本轮补丁反向检查通过，未实际执行回滚。
- 验证仅使用本地临时目录、设置替身与模拟上游；未使用真实 OAuth Token、未修改生产配置或账务，不能代替部署后的公网 DNS/证书/反向代理和真实 Basispoints 请求验证。
### Notes
- `backend/internal/service/openai_gateway_service.go`：增加受锁保护的实例级中转引用和关闭标记。
- `backend/internal/service/openai_basispoints_forward.go`：在协议转换前重写内嵌图片，并在额外正文复制前准入，错误和结束均释放预算。
- `backend/internal/service/settings_view.go`：为管理员系统设置增加中转开关和地址字段。
- `backend/internal/service/setting_parse.go`：增加默认关闭值及设置读取，不改公开设置。
- `backend/internal/service/setting_update.go`：保存前校验并规范化 HTTPS 根地址，同批写入两个配置键。
- `backend/internal/handler/dto/settings.go`：仅管理员系统设置响应增加中转字段。
- `backend/internal/handler/admin/setting_handler.go`：在管理员读取设置时返回中转状态。
- `backend/internal/handler/admin/setting_handler_update.go`：补更新与回包字段，未传入时保留原值。
- `backend/internal/server/routes/gateway.go`：注册不要求用户 API Key 的 GET/HEAD 能力链接，无公开上传路由。
- `backend/internal/server/middleware/logger.go`：普通访问日志对图片 token 路径脱敏。
- `backend/internal/server/middleware/request_logger.go`：请求级日志统一脱敏图片 token 路径。
- `backend/cmd/server/wire.go`：将图片中转关闭纳入既有服务清理链。
- `backend/cmd/server/wire_gen.go`：同步生成入口内的同一清理行为。
- `frontend/src/views/admin/SettingsView.vue`：新增 Basispoints 图片中转卡片、默认关闭、地址校验与保存。
- `frontend/src/api/admin/settings.ts`：补系统设置和更新请求的类型字段。
- `frontend/src/i18n/locales/zh/admin/settings.ts`：增加中文开启方式、隐私、到期和多副本说明。
- `frontend/src/i18n/locales/en/admin/settings.ts`：同步英文说明与错误提示。
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`：验证真实设置交互、回填、关闭保留地址及非法输入；隔离无关邮件编辑组件。
- `docs/OPENAI_OAUTH_BASISPOINTS.md`：更正旧的不支持内嵌图片说明，补开启步骤、限制、隐私、部署与回滚边界。
- `progress.md`：仅在末尾追加本轮实现、验证和回滚记录。
- `backend/internal/service/basispoints/NOTICE.md`：补充 v2.8.12 图片中转来源及本地准入范围差异。
- `backend/internal/server/api_contract_test.go`：仅更新两个管理员设置接口预期中的新增默认字段。
- `backend/internal/service/basispoints/image_relay.go`：移植私有磁盘暂存、签名短链、格式与大小校验、TTL 清理、磁盘配额和下载限流。
- `backend/internal/service/basispoints/image_relay_test.go`：移植图片往返、作用域、格式、大小、批次原子性、到期和并发复用验证。
- `backend/internal/service/basispoints/image_relay_disk_test.go`：移植私有文件权限、预占释放、异常退出清理及并发下载验证。
- `backend/internal/service/setting_excel_bps_image.go`：以现有设置表保存开关和 HTTPS 根地址，默认关闭并验证安全地址。
- `backend/internal/service/setting_excel_bps_image_test.go`：验证保存即时生效、地址切换、关闭读取和无效更新原子性。
- `backend/internal/handler/openai_excel_bps_image.go`：增加只读图片处理入口。
- `backend/internal/service/openai_basispoints_image.go`：接入懒初始化、动态配置、只读短链访问和停机清理，匿名读取不创建存储。
- `backend/internal/service/basispoints/image_relay_admission.go`：新增仅在已选中 Basispoints 且中转启用时调用的有界内存/在途请求准入。
- `backend/internal/handler/admin/setting_handler_excel_bps_image_test.go`：验证配置读写、局部更新保留、非法值拒绝及公开设置不泄漏地址。
- `backend/internal/server/routes/excel_bps_image_test.go`：验证读取入口可达、无公开上传且不存在图片安全返回 404。
- `backend/internal/server/middleware/bps_image_logger_test.go`：验证图片路径脱敏且其他路径不变。
- `backend/internal/service/openai_basispoints_image_test.go`：验证 OAuth/Setup Token 流式和非流式完整转发、默认 Codex/API Key 隔离、关闭和资源释放。
- `backend/internal/service/basispoints/image_relay_admission_test.go`：验证请求大小、预算、并发、取消、重复释放与关闭后拒绝。
- 独立备份与验证日志位于 `/tmp/sub2api-bps-image-backup.XCFwmB`；可执行回滚：在仓库根目录先运行 `git apply --reverse --check /tmp/sub2api-bps-image-backup.XCFwmB/image-relay.patch`，确认无后续重叠修改后运行 `git apply --reverse /tmp/sub2api-bps-image-backup.XCFwmB/image-relay.patch`。补丁仅覆盖本轮源码、测试及文档，不回滚追加日志，不覆盖之前的本地定制。部署后的行为回滚可直接关闭后台图片中转开关。
- 开启时必须填写上游可访问的本服务 HTTPS 根地址，代理需放行短链读取并避免缓存/记录 token；链接持有者可读取图片。多实例需要让读取命中原实例，重启会使旧链接失效；这些使用边界已写入页面提示和文档。
- 本轮不改变版本号、不构建或推送 Docker 镜像、不部署或重启服务；保留所有既有工作区改动。


## 2026-09-26 - Task: 移植 ranxi2001/sub2api v2.8.14 的 Basispoints 功能与修复

### What was done

- 以远端 v2.8.14 固定提交 `e39898c680ecd69381e549ae54c97011107f1757` 为依据，按功能移植 #83/#91、#84、#86、#87、#89，不整仓覆盖本地定制，版本保持 0.2.8。
- 修复代码型 function 工具的嵌套转义，保留命名空间、大整数、权限参数和历史往返；只转译工具报文，不执行工具代码。
- 图片中转新增请求体/处理预算/在途请求配置，默认 64 MiB / 1,024 MiB / 128 路；按真实 body 计量，容量按并发下限扩展，仅作用于已启用中转且实际走 Basispoints 的分支，保留既有网关限制。
- 增加账号级、默认关闭的“遇到 BPS 403 时自动切回 Codex”，仅真实通用 HTTP 403 触发；使用凭据和配置条件更新、同事务调度事件及提交后缓存刷新，原请求不自动重放。
- 增加账号测试中的 BPS 工具往返模式：最多三阶段上游请求，只验证随机 echo 标记；同账号不重叠、每实例最多三个，不能静默回退原生 Codex 后误报成功。
- BPS 正常响应更新额度快照、429 按耗尽窗口/reset 或既有兜底设置冷却；普通账号测试与真实请求共用状态处理。保留本地原有有界重试/换号和计费逻辑，不把整个账号改成错误态。
- 回归发现并修复空响应头兼容问题；先复现工具探测忽略取消后仍完成三阶段的问题，再补传请求 context 和每阶段取消检查，确保取消后不继续提交、名额被释放。
- 同步后台中英文配置说明、使用/容量/隐私边界、来源声明和独立回滚补丁。未导入无关采集、TPS、长上下文徽标和质量默认模型功能，未修改默认 Codex、API Key、WS 或独立图片/视频业务路径。

### Testing

- `go test -tags unit ./internal/service ./internal/repository -count=1`：最终全量通过，日志 `service-final.log`。
- `go test -tags unit ./internal/handler/... ./internal/server/... ./cmd/server -count=1`：通过，包含设置接口契约及默认路由；日志 `handler-server.log`。
- `go test -race ./internal/service/basispoints -count=1`：完整协议包通过，6.226 秒，日志 `protocol-final.log`。
- `go test -race ./internal/service/basispoints ./internal/service ./internal/repository -run 'Basispoints|BPSProbe|ImageRelay|FUNCTION_CODE|FunctionCode' -count=1`：最终定向竞态回归通过，包含新增取消修复、容量、403、429和调度投影；日志 `final-race.log`。
- 新增取消用例修复前稳定失败（两个场景），修复后通过；前后日志分别为 `probe-cancel-before.log` 和 `final-race.log`。
- 真实 PostgreSQL 18.1 + Redis 8.4 集成：`go test -overlay /tmp/sub2api-bps-v2814.ZTdyO7/integration-local-overlay.json -race -tags integration ./internal/repository -run DisableBasispointsOn403 -count=1 -v` 通过，覆盖16路并发仅一次切换、调度事件/缓存一致、旧配置/凭据不覆盖、Setup Token，以及事务失败回滚。原 Testcontainers 的临时端口映射失败；临时 overlay 仅替换测试连接初始化，业务SQL和测试正文不变、未跳过测试、未修改仓库 harness。使用独立临时数据库，未连接生产库；测试容器和其临时卷已清理。日志 `db-integration-local.log`。
- 前端 SettingsView、EditAccountModal、AccountTestModal、语言完整性共 142 项测试通过；`pnpm run typecheck`、本轮前端文件定向 ESLint、`pnpm exec vite build` 均通过，保留既有构建警告而不扩大修改范围。
- `git diff --check` 和 `git apply --reverse --check /tmp/sub2api-bps-v2814.ZTdyO7/v2814.patch` 通过。未使用生产 OAuth Token 请求真实上游，未验证真实公网图片地址的 DNS/证书配置。

### Notes

- `backend/internal/service/basispoints/catalog.go`：为符合条件的 function 工具登记 FUNCTION_CODE 传输标记。
- `backend/internal/service/basispoints/request.go`：在请求协议说明中加入原始代码参数的传输契约。
- `backend/internal/service/basispoints/tools.go`：接入代码工具的参数打包和返回恢复，保留原 custom 行为。
- `backend/internal/service/basispoints/history_collision_test.go`：更新代码工具历史恢复的回归断言。
- `backend/internal/service/basispoints/history_recovery_test.go`：对齐新增传输标记的历史重放测试。
- `backend/internal/service/basispoints/NOTICE.md`：补充 v2.8.14 来源和本地适配边界。
- `docs/OPENAI_OAUTH_BASISPOINTS.md`：增加版本对照、配置/测试入口、容量风险和回滚说明。
- `backend/internal/service/setting_excel_bps_image.go`：增加三个图片容量设置、默认值和范围校验。
- `backend/internal/service/setting_parse.go`：补充图片容量默认值及后台读取解析。
- `backend/internal/service/setting_update.go`：验证并保存图片容量，复用已有设置表。
- `backend/internal/service/settings_view.go`：增加管理侧图片容量字段。
- `backend/internal/handler/admin/setting_handler_update.go`：支持图片容量局部更新，拒绝非法值并保留未提交字段。
- `backend/internal/handler/admin/setting_handler.go`：返回图片容量的实际后台配置。
- `backend/internal/handler/dto/settings.go`：为管理员设置 DTO 补齐三个容量字段。
- `backend/internal/service/basispoints/image_relay_admission.go`：按实际 body 动态执行容量预算和在途上限。
- `backend/internal/service/basispoints/image_relay_admission_test.go`：对齐默认 128 路/1,024 MiB 预算断言。
- `backend/internal/service/openai_basispoints_image.go`：将已验证的动态容量应用到当前图片中转实例。
- `backend/internal/service/setting_excel_bps_image_test.go`：验证容量默认值、动态更新及非法容量拒绝。
- `backend/internal/handler/admin/setting_handler_excel_bps_image_test.go`：验证容量 API 字段、局部更新保留和显式零值拒绝。
- `frontend/src/views/admin/SettingsView.vue`：新增图片容量输入、前端校验及保存参数。
- `frontend/src/api/admin/settings.ts`：同步设置读写接口的图片容量类型。
- `frontend/src/i18n/locales/zh/admin/settings.ts`：增加图片容量中文标签、预算及内存风险说明。
- `frontend/src/i18n/locales/en/admin/settings.ts`：增加图片容量英文标签、预算及内存风险说明。
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`：覆盖容量展示、保存和非法输入交互。
- `backend/internal/service/basispoints/image_relay.go`：在中转实例内保存加锁保护的动态准入配置。
- `backend/internal/service/openai_basispoints_image_test.go`：明确旧满载场景使用的容量，以验证资源释放。
- `backend/internal/service/openai_basispoints_forward.go`：在实际转发和普通账号测试中统一处理用量、429和可选403，并修复空响应头兼容。
- `backend/internal/service/ratelimit_service.go`：抽取既有限流落库部分供 BPS 立即调用，保留默认 Codex 原重试分支。
- `backend/internal/service/account_test_service.go`：增加 BPS 工具测试入口和探测名额状态。
- `backend/internal/service/openai_compact_probe.go`：登记并规范化 bps_tools 测试模式。
- `backend/internal/service/openai_gateway_forward.go`：确保 BPS 专项探测不能静默回退原生 Codex 后误报成功。
- `backend/internal/repository/scheduler_cache.go`：在调度投影中保留 BPS 端点、403和缓存创建开关。
- `frontend/src/components/account/EditAccountModal.vue`：增加默认关闭的账号级403回退开关及加载保存。
- `frontend/src/components/account/AccountTestModal.vue`：为 BPS 账号显示工具往返测试，保留默认 Codex 测试模式。
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`：覆盖403开关默认关闭、保存、账号切换及撤销行为。
- `frontend/src/components/account/__tests__/AccountTestModal.spec.ts`：覆盖 BPS 工具测试实际提交参数和原 compact 模式。
- `frontend/src/i18n/locales/zh/admin/accounts.ts`：增加403回退和工具测试的中文说明。
- `frontend/src/i18n/locales/en/admin/accounts.ts`：增加403回退和工具测试的英文说明。
- `backend/internal/server/api_contract_test.go`：更新两处设置接口契约中的容量默认值。
- `backend/internal/service/basispoints/function_code_transport.go`：移植精确匹配代码型工具及原始代码封装/解包。
- `backend/internal/service/basispoints/function_code_transport_test.go`：覆盖嵌套转义、大整数、命名空间、权限参数和历史往返。
- `backend/internal/service/basispoints/image_relay_capacity_test.go`：验证512路小请求、真实字节计量和动态降额释放。
- `backend/internal/repository/account_repo_basispoints.go`：事务化切换端点并同事务写调度事件，提交后刷新缓存。
- `backend/internal/repository/account_repo_basispoints_auto_disable_test.go`：验证数据库失败、回滚与事件写入原子性。
- `backend/internal/repository/account_repo_basispoints_integration_test.go`：验证真实数据库条件更新、16路并发幂等及Setup Token兼容。
- `backend/internal/service/openai_basispoints_fallback.go`：增加账号资格/默认关闭检查及统一响应状态处理。
- `backend/internal/service/account_test_service_bps_probe.go`：实现三阶段非执行性echo探测、名额控制及取消后不继续提交。
- `backend/internal/service/account_test_service_bps_probe_test.go`：覆盖三阶段成功/错误、模式隔离、准入和客户端取消。
- `backend/internal/service/openai_basispoints_auto_disable_test.go`：覆盖真实请求和账号测试的403 opt-in、权限错误豁免及不重放。
- `backend/internal/service/openai_basispoints_ratelimit_test.go`：覆盖快照、耗尽窗口、冷却、取消、错误写入及原有failover。
- `backend/internal/service/basispoints_account_test_fixture_test.go`：提供脱敏、确定性的BPS账号测试样例。
- `backend/internal/repository/scheduler_cache_basispoints_test.go`：验证BPS调度字段保留且不泄漏凭据、不修改旧快照。
- `progress.md`：仅追加本轮实施、验证证据和回滚说明。
- 独立备份、回归日志和回滚补丁位于 `/tmp/sub2api-bps-v2814.ZTdyO7`。在仓库根执行 `git apply --reverse --check /tmp/sub2api-bps-v2814.ZTdyO7/v2814.patch`，无后续重叠修改时再执行 `git apply --reverse /tmp/sub2api-bps-v2814.ZTdyO7/v2814.patch`。该补丁只撤销本轮52个源码/测试/文档文件的增量，不回滚历史进度日志，也不撤销此前图片中转或其他本地定制。
- 运行行为可关闭账号403回退开关、将账号切回默认 Codex，或关闭图片中转；不需数据库迁移。容量调大可能增加内存压力，512路配置对应至少4,096 MiB估算预算；设置不会突破反向代理或原网关更低的限制。
- 本轮未构建或推送 Docker 镜像、未部署/重启服务、未修改生产配置或账务；后续发布需构建新镜像。临时集成验证的端口仅为本次测试使用，不是部署设置。

## 2026-09-26 - Task: 构建并推送包含 Basispoints v2.8.14 适配的双架构镜像

### What was done

- 使用当前完整工作区及既有 Dockerfile 构建并推送 `iotwq/china-api:latest` 和固定标签 `iotwq/china-api:0.2.8-20260926-125430`，包含此前完成的 Basispoints v2.8.14 适配与图片中转改动，未拉取新源码、未变更版本号或覆盖既有定制。
- 发布 `linux/amd64` 与 `linux/arm64` 双架构；版本保持 `0.2.8`，构建标识 `f867e4f9bd2f-dirty`，构建时间 `2026-09-26T04:54:30Z`。
- 两个远端标签均指向索引摘要 `sha256:37c28bbe7306e91e2de4602fbcd60132c3640250ab41626b91c8756d7ef48b3b`。amd64 摘要为 `sha256:e6ab39ca12f6d24bec82a43c22511bb3732c996db3c00cb7e5ef5de0c550affa`；arm64 摘要为 `sha256:b41313ed5a1c2c5a4eccd25049b2b5547e9eedb12ee3ef777d578ba201ecdc30`。

### Testing

- Docker Buildx 双架构生产构建及 `--push` 成功退出；镜像内前端语言完整性测试 3 项、TypeScript 检查、Vite 生产构建及两个架构的 Go 嵌入式编译通过。本轮未重复上轮源码全量回归，也未执行真实上游业务请求。
- 分别读取 Docker Hub 的 latest 与固定标签清单，两者包含 amd64、arm64，且摘要与构建 metadata 一致；附加 unknown/unknown 清单为构建证明，不是额外运行架构。
- 按两个已发布的精确摘要加载镜像，分别在 `--network none --read-only --user 1000:1000` 的临时容器中执行 `/app/sub2api -version`，均成功返回 `0.2.8 / f867e4f9bd2f-dirty / 2026-09-26T04:54:30Z`。未挂载数据或连接生产数据库；容器自动删除。
- 初始构建及一次重试被 Docker Hub IPv6 连接重置阻断；通过当前已有本机代理的命令级 `HTTP_PROXY/HTTPS_PROXY=http://127.0.0.1:10808` 完成构建、推送和清单核验，未修改系统或 Docker 全局代理。普通 Docker daemon pull 返回网络 EOF，改由可工作的 Buildx 按同一远端摘要加载，加载后的镜像摘要及架构一致，版本冒烟通过。
- `git diff --check` 通过。构建日志、metadata、远端清单及双架构版本日志保存在 `/tmp/sub2api-image-bps2814-20260926.kVIy2X`。

### Notes

- `progress.md`：仅追加本次发布、验证、网络处理和回滚记录；本轮未修改源码、版本、部署配置、数据库或账务，未部署或重启服务。
- 发布前已读取并确认旧 latest 摘要为 `sha256:6a8856ba91fbdea4056926ba344e1a41b266ed7df8efebb37509d6bf47e51594`。如需回滚仓库 latest 标签，可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:6a8856ba91fbdea4056926ba344e1a41b266ed7df8efebb37509d6bf47e51594`；该命令只回退远端标签，不改本次固定标签，也不会自动回退已运行的容器。本轮未执行回滚。

## 2026-09-26 - Task: 删除使用密钥中 Codex CLI 模板的本地模型目录配置行

### What was done

- 从普通 OpenAI、OpenAI WebSocket、Grok 及其他分组共用的 Codex 配置模板删除 `model_catalog_json`，macOS / Linux 与 Windows 展示和复制结果均不再包含该字段。
- 仅删除四处模板配置行，保留 API 地址、模型、推理等级、鉴权、WebSocket 和模型目录获取/下载逻辑；未修改用户本机的 Codex 配置文件。

### Testing

- 新增四个模板分支的双系统展示/复制回归用例，修复前均因存在 `model_catalog_json` 稳定失败；修复后 `pnpm exec vitest run src/components/keys/__tests__/UseKeyModal.spec.ts` 全部 27 项通过，包含原有目录下载和模型选择用例。
- `pnpm run typecheck` 及 `pnpm exec eslint src/components/keys/UseKeyModal.vue src/components/keys/__tests__/UseKeyModal.spec.ts` 通过。
- 模板源码搜索确认没有残留 `model_catalog_json`；与施工前快照比较确认业务组件仅删除四行配置。`git diff --check` 和本轮增量补丁的反向应用检查通过。

### Notes

- `frontend/src/components/keys/UseKeyModal.vue`：删除四个 Codex 模板生成分支中的本地目录覆盖配置。
- `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts`：调整旧断言，补双系统展示和复制验证，保留模型目录下载覆盖。
- `docs/CODEX_CLI_CONFIG_TEMPLATE.md`：说明模板调整范围、现有客户端配置需手动更新，以及验证命令。
- `progress.md`：仅追加本轮变更、验证和回滚记录。
- 回滚：在仓库根先执行 `git apply --reverse --check /tmp/sub2api-codex-catalog-line.z19HgX/change.patch`，确认无后续重叠改动后执行 `git apply --reverse /tmp/sub2api-codex-catalog-line.z19HgX/change.patch`。补丁仅撤销本轮两个源码/测试文件和新增文档，不覆盖此前定制或改写历史日志。
- 本轮未构建或推送镜像、未部署或重启服务；页面需发布更新后生效，已复制到用户客户端的配置不会自动修改。

## 2026-09-26 - Task: 移植 v2.8.15–v2.8.17 独立 Basispoints 修复（不含 Mihomo）

### What was done

- 对照 ranxi2001/sub2api 的 v2.8.14（e39898c680ecd69381e549ae54c97011107f1757）至 v2.8.17（26b324b）源码移植独立 BPS 增量；按用户确认排除 Mihomo 托管代理、订阅、内核管理及会话代理池，保留项目版本 0.2.8。不是整仓升级，也未引入 Excel 或上游按模型批量设置界面。
- 完善 custom、FUNCTION_CODE、FUNCTION_CMD 和历史工具恢复，增加有界且隔离的工具目录缓存、参数校验与纠错。纠错使用原账号、代理和并发控制，释放前次响应后才再次准入，累计真实用量；账号三阶段探测仍最多三次上游请求。
- 增加原生图片附件模式和可配置图片限制，保留 HTTPS 中转与请求体/预算/在途限制；上传前校验图片及请求，失败不继续生成，不自动切换传输方式。
- 隔离 BPS 429 与 Codex 冷却/换号；真实 BPS 401 纳入现有凭据策略；增加默认关闭的账号级 403 分组动作，事务化保护管理员最新修改、重复调用和调度缓存，与切回 Codex 独立。
- 更新后台设置、账号编辑、双语提示和说明；保留默认 Codex/API Key 路线，不修改数据库结构或生产账务。

### Testing

- 最终受影响后端包全量回归通过：go test ./internal/service ./internal/repository ./internal/handler/... ./internal/server/... ./cmd/server/...；service 139.371s、repository 2.506s、handler 46.012s、admin 3.023s、middleware 3.838s、routes 10.302s，其他包通过或命中缓存。日志 backend-final.log。
- 协议包完整竞态测试通过：go test -race ./internal/service/basispoints -count=1（14.380s）；go test -tags unit -race ./internal/service -run Basispoints|ExcelBPS -count=1 通过（4.510s），包含显式 null 目标验证。日志 protocol-race.log、bps-final-race.log。
- 真实 PostgreSQL/Redis 集成竞态测试通过：go test -tags integration -race ./internal/repository -run Basispoints -count=1 -v（5.337s）。实际执行 OAuth/Setup Token、16 路并发幂等、已有绑定优先级、管理员修改保护、事务失败回滚及缓存更新，并非跳过测试。日志 repository-final-race.log。
- 前端 EditAccountModal 90 项、SettingsView 48 项及语言完整性 3 项，共 141 项通过；定向 ESLint 通过；pnpm run build 中语言检查、vue-tsc 与 Vite 生产构建通过（9.73s）。测试存在组件桩提示，构建保留原有大包警告，无失败。日志 frontend-final.log、frontend-lint.log、frontend-build.log。
- git diff --check 通过；73 个源码/测试/文档文件的本轮增量反向应用检查通过。集成测试容器已自行清理，未操作原有 Buildx 容器。证据及基线保存在 /tmp/sub2api-bps-v2817.tS88Uv。
- 未使用生产 OAuth 凭据请求真实 BPS；本地验证不能证明真实上游权限、上传可用性或模型质量。本轮未构建或推送镜像、未部署或重启服务。

### Notes

- 回滚仅撤销本轮增量：在仓库根先执行 git apply --reverse --check /tmp/sub2api-bps-v2817.tS88Uv/v2817.patch，确认无后续重叠修改后执行 git apply --reverse /tmp/sub2api-bps-v2817.tS88Uv/v2817.patch。反向检查已通过；补丁不含 progress.md，也不撤销此前 BPS v2.8.14、Codex 模板及其他定制。临时目录不保证长期保留，长期回滚需保存补丁和基线。
- 可关闭 403 分组动作、将账号切回默认 Codex，或关闭内嵌图片处理。关闭 403 动作不会恢复已经变更的分组；“离开所有分组”须明确选择，恢复需管理员手动绑定。
- 改动文件清单：
- backend/internal/handler/admin/setting_handler.go：返回图片模式和限制字段。
- backend/internal/handler/admin/setting_handler_excel_bps_image_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/handler/admin/setting_handler_update.go：接收新图片设置并拒绝显式无效值。
- backend/internal/handler/dto/settings.go：补充设置响应的图片处理字段。
- backend/internal/repository/account_repo_basispoints_groups.go：事务化处理 403 分组条件更新及调度 outbox。
- backend/internal/repository/account_repo_basispoints_groups_integration_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/repository/account_repo_basispoints_groups_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/repository/scheduler_cache.go：保留 403 分组动作的调度元数据。
- backend/internal/service/account_basispoints_groups.go：定义可选 403 分组动作与明确的目标校验。
- backend/internal/service/admin_account.go：在创建、更新、extra 和批量入口校验分组动作。
- backend/internal/service/admin_service_basispoints_groups_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/NOTICE.md：补充 v2.8.15–v2.8.17 来源说明。
- backend/internal/service/basispoints/agent_message_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/attachments.go：增加原生附件预处理与受限附件 ID 缓存。
- backend/internal/service/basispoints/attachments_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/basispoints_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/catalog.go：兼容 additional_tools 和工具目录合并。
- backend/internal/service/basispoints/catalog_cache.go：增加身份隔离、有界且版本化的工具目录缓存。
- backend/internal/service/basispoints/content.go：完善附件及工具结果的内容转换。
- backend/internal/service/basispoints/content_diagnostics_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/custom_history_transport_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/deep_audit_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/function_cmd_transport.go：支持命令工具的 FUNCTION_CMD 传输。
- backend/internal/service/basispoints/function_cmd_transport_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/image_relay.go：接入共享准入与可配置中转存储限制。
- backend/internal/service/basispoints/image_relay_admission.go：提取附件与中转共用的并发和预算准入。
- backend/internal/service/basispoints/image_relay_disk_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/image_relay_limits.go：集中解析图片限制并校验相互约束。
- backend/internal/service/basispoints/image_relay_limits_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/image_relay_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/images.go：支持可配图片数量、合法 file_id 和 original 精度。
- backend/internal/service/basispoints/images_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/plan.go：完善工具计划和纠错目标判定。
- backend/internal/service/basispoints/protocol_completion_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/replay_cache_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/request.go：集成工具目录恢复、图片内容和 BPS 压缩阈值。
- backend/internal/service/basispoints/stream.go：增加有界纠错、用量累计和流式收尾。
- backend/internal/service/basispoints/tool_repair.go：构造并解析保留原始代码的纠错请求。
- backend/internal/service/basispoints/tool_repair_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/basispoints/tool_schema.go：在派发工具前验证参数结构。
- backend/internal/service/basispoints/tools.go：完善 custom、FUNCTION_CODE 和命令工具转换。
- backend/internal/service/basispoints/unknown_tool_repair.go：仅在工具尚未派发时允许一次未知目标恢复。
- backend/internal/service/openai_basispoints_attachments.go：通过原账号认证、代理和准入上传图片附件。
- backend/internal/service/openai_basispoints_attachments_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/openai_basispoints_auth_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/openai_basispoints_auto_disable_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/openai_basispoints_completion_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/openai_basispoints_fallback.go：隔离 429、处理 401 并串接独立 403 动作。
- backend/internal/service/openai_basispoints_forward.go：集成目录、图片和纠错，并保留探测最多三次请求。
- backend/internal/service/openai_basispoints_groups_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/openai_basispoints_image.go：按请求读取图片模式和限制，共享准入。
- backend/internal/service/openai_basispoints_image_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/openai_basispoints_ratelimit_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/openai_basispoints_repair.go：接入同账号纠错、名额释放、取消和失败用量。
- backend/internal/service/openai_basispoints_repair_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/openai_basispoints_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/openai_gateway_response_handling.go：仅对实际 BPS 路径隔离重放并保留终态用量。
- backend/internal/service/openai_gateway_service.go：持有图片准入、附件及工具目录缓存。
- backend/internal/service/setting_excel_bps_image.go：定义图片模式、默认限制及验证。
- backend/internal/service/setting_excel_bps_image_test.go：补充或更新本轮 BPS 行为的回归验证。
- backend/internal/service/setting_parse.go：解析保存的图片处理配置。
- backend/internal/service/setting_update.go：更新新字段且保留旧客户端省略的配置。
- backend/internal/service/settings_view.go：向设置视图暴露图片模式和限制。
- docs/OPENAI_OAUTH_BASISPOINTS.md：说明移植范围、行为、使用方式和风险。
- frontend/src/api/admin/settings.ts：补充图片设置的前端 API 类型。
- frontend/src/components/account/EditAccountModal.vue：增加默认关闭的 403 分组动作和明确目标选择。
- frontend/src/components/account/__tests__/EditAccountModal.spec.ts：补充或更新本轮 BPS 行为的回归验证。
- frontend/src/i18n/locales/en/admin/accounts.ts：补充英文 403 分组动作提示。
- frontend/src/i18n/locales/en/admin/settings.ts：补充英文图片模式、限制及风险提示。
- frontend/src/i18n/locales/zh/admin/accounts.ts：补充中文 403 分组动作提示。
- frontend/src/i18n/locales/zh/admin/settings.ts：补充中文图片模式、限制及风险提示。
- frontend/src/views/admin/SettingsView.vue：增加原生图片模式及可见的图片限制。
- frontend/src/views/admin/__tests__/SettingsView.spec.ts：补充或更新本轮 BPS 行为的回归验证。
- progress.md：仅追加本轮实施、验证与回滚记录。
- 复跑服务竞态验证时须将正则用引号包裹，避免 shell 把竖线当管道：在 backend 目录执行 `go test -tags unit -race ./internal/service -run 'Basispoints|ExcelBPS' -count=1`。

## 2026-09-26 - Task: 构建并推送包含独立 BPS v2.8.15–v2.8.17 修复的双架构镜像

### What was done

- 基于当前完整工作区及既有 Dockerfile 构建并推送 `iotwq/china-api:latest`，同时发布固定标签 `iotwq/china-api:0.2.8-20260926-235652`；保留全部本地定制，没有拉取新源码、变更版本或修改部署配置。
- 包含此前完成的独立 BPS v2.8.15–v2.8.17 修复与 Codex CLI 模板调整；不包含尚未实施的 EPUSDT 接入或 BPS 429 自动重试恢复，也未引入 Mihomo。
- 版本保持 `0.2.8`，构建标识 `f867e4f9bd2f-dirty`，构建时间 `2026-09-26T15:56:52Z`，支持 `linux/amd64`、`linux/arm64`。
- 两个远端标签均指向索引摘要 `sha256:895349847c65764b82a45e871ff2a39a6666b9ab3a16335ba7252987d64d6dc1`；amd64 摘要 `sha256:69bc4ffe59b6a72c41d573227576204fd6a530251cadcc6afbdd735595a14bde`，arm64 摘要 `sha256:d6041b45dff2e32096e7ffc8484de980335870cdba8444d911392c4527fd9c3f`。

### Testing

- Docker Buildx 双架构生产构建及两个标签的 `--push` 成功退出；容器内前端语言完整性 3 项测试、vue-tsc 类型检查、Vite 生产构建及两个架构的 Go 嵌入式编译通过。本轮未重复上轮源码全量回归。
- 重新读取 Docker Hub 的 latest 与固定标签清单，自动核对两者摘要均与构建 metadata 一致，且运行架构恰为 amd64、arm64；附加 unknown/unknown 项为构建证明。
- 分别拉取本次发布的 amd64 索引目标与 arm64 平台摘要，在 `--network none --read-only --user 1000:1000` 的临时容器内执行 `/app/sub2api -version`，均返回 `0.2.8 / f867e4f9bd2f-dirty / 2026-09-26T15:56:52Z`。镜像元数据中的平台及对应摘要也一致；容器自动删除，没有挂载数据或连接生产数据库。
- 首次远端标签核验遇到 Docker Hub token 请求 TLS handshake timeout / EOF，重试后通过；构建与清单核验使用已有本机代理的命令级 `HTTP_PROXY/HTTPS_PROXY=http://127.0.0.1:10808`，未修改系统或 Docker 全局代理、登录凭据。两个架构的普通 Docker pull 均成功。
- `git diff --check` 通过；构建日志、metadata、发布前后清单与双架构版本日志位于 `/tmp/sub2api-image-bps2817.OSenw4`。

### Notes

- `progress.md`：仅追加本轮构建发布、验证和回滚记录；本轮未修改源码、版本、数据库或账务，未部署或重启服务。
- 发布前已实时确认旧 latest 摘要为 `sha256:37c28bbe7306e91e2de4602fbcd60132c3640250ab41626b91c8756d7ef48b3b`。需要回滚 Docker Hub latest 时，可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:37c28bbe7306e91e2de4602fbcd60132c3640250ab41626b91c8756d7ef48b3b`；只回退远端 latest 标签，不改本次固定标签，也不会自动回退已部署的容器。本轮未执行回滚。


## 2026-09-27 - Task: 移植 ranxi2001/sub2api v2.8.18–v2.8.19 独立 Basispoints 修复

### What was done

- 以 v2.8.18 c1008182bd1bb8f50ff95133fa486fb9d4676811、v2.8.19 08356987417f205cd704bf79180def20fcbbf445 为固定来源，适配前次 v2.8.17 后的独立 BPS 增量；保留本地版本 0.2.8 及任务开始前的未提交改动，不整仓升级。
- 合并工具批次修复、截图与历史图片、子代理归属、同账号一次密文恢复、模型目录能力限制、安全错误位置、独立 H2 反馈、403 真实动作标记、后台测试隔离和客户端取消/耗时逻辑；提供两个默认关闭的账号兼容选项。
- 按后续 #140 覆盖 #134 的最终策略，BPS 账号普通 auto/none 工具目录不再静默切回 Codex；省略不可执行的声明并明确能力不可用，强制工具选择仍拒绝；没有引入普通转发已无效的省略工具开关。
- 不引入 Mihomo、IP/会话代理池、预热/订阅/内核管理或无关管理员功能；不恢复 BPS 429 重试，不接入 EPUSDT。不修改数据库结构、用户余额或生产配置，不构建、推送或部署镜像。
- 新的真实数据库回归发现任务开始前已存在的共享错误日志 SQL 与迁移 136 不一致：仍引用 request_body、request_body_truncated、request_body_bytes。已单独请求扩大范围修复，尚未收到批准，因此保留原逻辑及失败证据，不宣称该持久化功能已可用。

### Testing

- 后端最终 go test ./... 通过（含未受影响包的缓存）；internal/service 本轮重测 139.181s。完整结果：/tmp/sub2api-bps-v2819.8oZYKu/backend-final-verified.log。
- 最终定向工具路由/探测/选项回归通过；覆盖 30 个工具类型 × 选择方式 × 流式组合及探测不发送额外请求。结果：/tmp/sub2api-bps-v2819.8oZYKu/hosted-policy-final.log。
- go test -race ./internal/service ./internal/repository -run "Basispoints|ExcelBPS|BPS" -count=1 通过；协议包和 transportdiag 的完整 race 另行通过，未将此前 no tests to run 当作通过证据。结果：/tmp/sub2api-bps-v2819.8oZYKu/race-final.log、protocol-race.log。
- 真实临时 PostgreSQL/Redis 的 BPS 403 分组、关闭、并发幂等、标记保护及调度投影集成回归通过；CI=1 go test -tags=integration ./internal/repository -run Basispoints -count=1 -v，结果：/tmp/sub2api-bps-v2819.8oZYKu/integration-bps.log。未连接生产数据库。
- 未通过：TestOpsClientCancellationMetricsAndDuration 在真实最新迁移表结构中报 pq: column "request_body" of relation "ops_error_logs" does not exist。任务前备份亦含该三列，证据：/tmp/sub2api-bps-v2819.8oZYKu/integration.log。共享 SQL 修复待确认，不将标准 go test 的通过表述为全量数据库集成通过。
- 前端账号编辑、403 徽标、语言完整性共 114 项测试通过；pnpm run build 的 i18n、vue-tsc、Vite 生产构建通过，只有现有 Browserslist/chunk-size 提示。结果：/tmp/sub2api-bps-v2819.8oZYKu/frontend-final-verified.log、frontend-build-final.log。
- git diff --check 通过；最终与本轮 before 快照逐文件对比，88 个代码/测试/NOTICE 文件变化，无非任务的纯格式增量；另更新本说明文档和本日志。增量回滚补丁 reverse --check 通过。
- 未使用真实 OAuth token 调用官方 BPS；模拟与本地数据库测试不保证上游权限、代理可达性或模型质量。

### Notes

- backend/internal/handler/ops_error_logger.go：记录请求总耗时，并把可信客户端取消与上游错误、业务限额区分。
- backend/internal/handler/ops_error_logger_duration_cancel_test.go：验证取消分类、字段安全性和请求耗时。
- backend/internal/pkg/ctxkey/ctxkey.go：增加请求开始时间的内部上下文键。
- backend/internal/repository/account_repo.go：在既有账号编辑锁内保留真实 BPS 403 标记，拒绝伪造并在重开 BPS 时清理关闭标记。
- backend/internal/repository/account_repo_basispoints.go：在真实的 BPS 403 自动关闭动作中事务化保存时间标记。
- backend/internal/repository/account_repo_basispoints_auto_disable_test.go：验证关闭动作的原有条件及新增时间标记。
- backend/internal/repository/account_repo_basispoints_groups.go：在真实的 BPS 403 分组动作中保存目标与时间标记。
- backend/internal/repository/account_repo_basispoints_groups_integration_test.go：验证真实 PostgreSQL 的分组动作及原 extra 保留。
- backend/internal/repository/account_repo_basispoints_marker_integration_test.go：验证真实数据库标记不可伪造、编辑保留及重新启用后的清理。
- backend/internal/repository/http_upstream.go：接入独立 BPS 连接池和传输反馈，保留本地原有请求发送、Codex TLS 与代理逻辑。
- backend/internal/repository/http_upstream_bps.go：对确认的 HTTP 代理 H2 故障启用同代理后续请求的短期 H1 尝试，不重放当前请求。
- backend/internal/repository/http_upstream_bps_eof_test.go：验证响应体异常结束、取消及普通 EOF 不被错误混淆。
- backend/internal/repository/http_upstream_bps_policy_test.go：验证 BPS 协议降级状态、时限、代理隔离和默认线路不变。
- backend/internal/repository/ops_repo.go：为错误日志增加已有 duration_ms 字段的写入，未擅自修复既有已删除字段引用。
- backend/internal/repository/ops_repo_args_test.go：更新错误日志参数位置断言以覆盖新增耗时。
- backend/internal/repository/ops_repo_cancellation_integration_test.go：增加真实数据库取消统计与耗时回归，明确暴露既有共享 SQL 的已删列问题。
- backend/internal/repository/ops_repo_dashboard.go：将客户端取消从 SLA 和业务限额统计中排除。
- backend/internal/repository/ops_repo_duration_test.go：验证错误日志目标列、占位符、参数数量与耗时位置一致。
- backend/internal/repository/ops_repo_preagg.go：使聚合指标按取消分类排除非上游故障。
- backend/internal/repository/ops_repo_trends.go：使趋势和分布统计保持取消分类一致。
- backend/internal/repository/scheduler_cache.go：保留两个 BPS 兼容选项及真实 403 动作标记的调度投影。
- backend/internal/server/middleware/logger.go：从请求入口记录内部开始时间用于总耗时。
- backend/internal/server/routes/gateway_models_pinned_test.go：补充既有 pinned 模型测试桩的候选查询接口，验证新增模型能力过滤。
- backend/internal/service/account_basispoints_403_marker.go：集中定义和保护实际 403 动作的账号元数据。
- backend/internal/service/account_basispoints_403_marker_test.go：验证对应行为：集中定义和保护实际 403 动作的账号元数据。
- backend/internal/service/account_basispoints_groups.go：校验本轮两个 BPS 兼容选项为布尔值。
- backend/internal/service/account_basispoints_options.go：定义仅在 BPS OAuth-like 模式生效的忽略图片与历史密文选项。
- backend/internal/service/account_basispoints_options_test.go：验证对应行为：定义仅在 BPS OAuth-like 模式生效的忽略图片与历史密文选项。
- backend/internal/service/account_test_service.go：初始化后台测试请求头，避免后台测试写请求头时 panic。
- backend/internal/service/basispoints/NOTICE.md：记录 v2.8.18/v2.8.19 来源、移植范围、最终工具策略与已知数据库限制。
- backend/internal/service/basispoints/agent_message_test.go：验证子代理历史消息、内容校验和图片处理。
- backend/internal/service/basispoints/attachments.go：验证并保留工具原生截图，仍按请求计数且保留本地字节限额。
- backend/internal/service/basispoints/attachments_test.go：验证并保留工具原生截图，仍按请求计数且保留本地字节限额。
- backend/internal/service/basispoints/catalog_cache.go：支持非执行性描述变化的工具目录复用并保留作用域隔离。
- backend/internal/service/basispoints/content.go：改进安全内容校验、历史消息和图片归一化。
- backend/internal/service/basispoints/content_test.go：验证对应协议行为：改进安全内容校验、历史消息和图片归一化。
- backend/internal/service/basispoints/encrypted_content.go：提供显式历史密文省略，不改动工具参数与可用明文。
- backend/internal/service/basispoints/encrypted_content_test.go：验证对应协议行为：提供显式历史密文省略，不改动工具参数与可用明文。
- backend/internal/service/basispoints/history_message_images_test.go：验证子代理与历史消息图片的安全处理。
- backend/internal/service/basispoints/history_messages.go：将历史 author/recipient 降为描述文本，不允许提升角色权限。
- backend/internal/service/basispoints/history_messages_test.go：验证对应协议行为：将历史 author/recipient 降为描述文本，不允许提升角色权限。
- backend/internal/service/basispoints/history_recovery_test.go：覆盖历史工具标识及目录恢复的兼容行为。
- backend/internal/service/basispoints/image_relay_test.go：扩展图片中转对新增内容结构的回归覆盖。
- backend/internal/service/basispoints/images.go：扩展图片识别与可选省略，使用本地开关名提供操作提示。
- backend/internal/service/basispoints/request.go：接入本请求校验的截图允许列表、工具历史转换和安全错误定位。
- backend/internal/service/basispoints/route.go：按最终工具选择语义识别严格探测的原生能力需求。
- backend/internal/service/basispoints/route_test.go：验证对应协议行为：按最终工具选择语义识别严格探测的原生能力需求。
- backend/internal/service/basispoints/tool_duplicates_test.go：验证非执行性目录差异可兼容、执行约束冲突仍拒绝。
- backend/internal/service/basispoints/tool_images.go：将 HTTPS/file_id 工具图片移入相邻上下文并保留关联，验证内嵌截图。
- backend/internal/service/basispoints/tool_images_test.go：验证对应协议行为：将 HTTPS/file_id 工具图片移入相邻上下文并保留关联，验证内嵌截图。
- backend/internal/service/basispoints/tool_repair.go：纠错时保留已校验工具操作与原始代码，重组后整批重新验证。
- backend/internal/service/basispoints/tool_repair_test.go：验证对应协议行为：纠错时保留已校验工具操作与原始代码，重组后整批重新验证。
- backend/internal/service/basispoints/tools.go：改进工具目录兼容及上下文转换，同时保留权限和协议校验。
- backend/internal/service/http_upstream_profile.go：增加与 Codex 隔离的 BPS HTTP profile。
- backend/internal/service/openai_basispoints.go：采用独立 BPS 传输 profile，并保持所选 BPS 路由不按工具目录切回 Codex。
- backend/internal/service/openai_basispoints_attachments.go：原生图片上传保留取消原因，并使用独立 BPS 传输 profile。
- backend/internal/service/openai_basispoints_cancellation.go：仅同时确认入站请求取消和返回取消原因时识别客户端取消。
- backend/internal/service/openai_basispoints_cancellation_test.go：验证对应行为：仅同时确认入站请求取消和返回取消原因时识别客户端取消。
- backend/internal/service/openai_basispoints_encrypted.go：准备一次同路由的明确密文拒绝恢复，仅删除可安全剥离的 opaque reasoning。
- backend/internal/service/openai_basispoints_encrypted_test.go：验证对应行为：准备一次同路由的明确密文拒绝恢复，仅删除可安全剥离的 opaque reasoning。
- backend/internal/service/openai_basispoints_forward.go：接入图片/密文选项、原生截图、同路由恢复、安全错误字段与客户端取消处理。
- backend/internal/service/openai_basispoints_history_messages_test.go：验证历史 author/recipient 与子代理消息通过实际 BPS 转发路径转换。
- backend/internal/service/openai_basispoints_hosted_tools_test.go：验证普通工具目录固定走 BPS并提示不可用能力，强制选择拒绝，探测不增加请求。
- backend/internal/service/openai_basispoints_ignore_encrypted_content_e2e_test.go：验证显式忽略历史密文及默认不启用的完整转发行为。
- backend/internal/service/openai_basispoints_ignore_images_e2e_test.go：验证图片处理关闭时显式忽略图片及默认行为隔离。
- backend/internal/service/openai_basispoints_models_manifest.go：根据实际候选路由限制 BPS 模型的加密多代理能力，保留缓存与路由隔离。
- backend/internal/service/openai_basispoints_models_manifest_test.go：验证对应行为：根据实际候选路由限制 BPS 模型的加密多代理能力，保留缓存与路由隔离。
- backend/internal/service/openai_basispoints_repair_test.go：验证工具纠错保留已通过操作、使用相同线路并累计真实用量。
- backend/internal/service/openai_basispoints_test.go：验证对应行为：采用独立 BPS 传输 profile，并保持所选 BPS 路由不按工具目录切回 Codex。
- backend/internal/service/openai_basispoints_tool_images_test.go：验证工具截图、图片转移、限额与原生附件的完整转发链路。
- backend/internal/service/openai_basispoints_v2819_test_helpers_test.go：验证对应行为：为本轮协议和取消回归提供可控模拟响应。
- backend/internal/service/openai_codex_models_service.go：在返回模型目录副本时按实际路由限制能力，并重算最终 ETag。
- backend/internal/service/openai_compact_stream_bridge.go：为已提交的 compact SSE 错误保留安全 param 字段。
- backend/internal/service/openai_compact_stream_failure_param_test.go：验证 compact 错误 param 的 SSE 回传且不破坏既有结构。
- backend/internal/service/ops_port.go：为错误日志写入模型增加可选总耗时字段。
- backend/internal/service/ops_upstream_context.go：增加可信客户端取消标记、逻辑 499 与请求总耗时上下文。
- backend/internal/service/scheduled_test_runner_panic_test.go：验证单个后台测试 panic 被隔离且后续计划继续。
- backend/internal/service/scheduled_test_runner_service.go：在单个定时测试计划边界恢复 panic，不引入新的调度系统。
- backend/internal/util/transportdiag/error.go：分类传输错误，不记录敏感请求内容。
- backend/internal/util/transportdiag/trace.go：记录脱敏网络阶段及协商协议，供 BPS 连接策略使用。
- backend/internal/util/transportdiag/trace_test.go：验证网络跟踪并发安全及脱敏行为。
- frontend/src/components/account/Basispoints403Badge.vue：显示真实 BPS 403 处理状态、时间及目标，避免误称账号封禁。
- frontend/src/components/account/EditAccountModal.vue：增加两个默认关闭的 BPS 兼容选项并在切回 Codex 时清理选择。
- frontend/src/components/account/__tests__/Basispoints403Badge.spec.ts：覆盖 403 标记时间、动作目标、当前账号状态及提示文案。
- frontend/src/components/account/__tests__/EditAccountModal.spec.ts：覆盖 OAuth/Setup Token 选项保存、默认关闭和 Codex 隔离。
- frontend/src/i18n/locales/en/admin/accounts.ts：补充 BPS 兼容选项与 403 处理状态的英文说明。
- frontend/src/i18n/locales/zh/admin/accounts.ts：补充 BPS 兼容选项与 403 处理状态的中文说明。
- frontend/src/views/admin/AccountsView.vue：在 OpenAI OAuth/Setup Token 账号状态栏显示 BPS 403 处理徽标。
- docs/OPENAI_OAUTH_BASISPOINTS.md：补充本轮行为、两个兼容选项、403 标记、排除范围与数据库发布前置限制。
- progress.md：仅在末尾追加本轮移植、验证、未通过项目及增量回滚说明。

- 回滚点：本轮代码修改前的逐文件快照位于 /tmp/sub2api-bps-v2819.8oZYKu/before；未使用 Git HEAD 回滚，避免覆盖用户此前未提交内容。
- 可执行代码/UI回滚：在 /Users/wangqiang/Project/codex/sub2api-wq 执行 git apply --reverse --check /tmp/sub2api-bps-v2819.8oZYKu/code-increment.patch，确认无冲突后执行 git apply --reverse /tmp/sub2api-bps-v2819.8oZYKu/code-increment.patch。本轮仅验证检查，未执行回滚。补丁包含上述 88 个代码/测试/NOTICE 文件，SHA-256 bd8d4c1a3c0658c7666b358fcf7781878b878b23c3824b471ed0ed48a93798a1。
- 回滚补丁不删除 docs/OPENAI_OAUTH_BASISPOINTS.md 和 progress.md：文档被仓库既有 docs/* 忽略规则排除在初始源码快照之外，保留文档与日志作为审计；实际回滚后应追加回滚记录并注明该节暂未启用，不改写历史。审计目录位于 /tmp，清理临时目录前需自行保留补丁和验证日志。


## 2026-09-27 - Task: 经确认修复共享错误日志 INSERT 与已迁移表结构不兼容

### What was done

- 根据用户对前轮阻塞的明确确认，先在全新临时 PostgreSQL 的完整迁移表结构中复现 request_body 列不存在，再仅移除共用错误日志 INSERT 的 request_body、request_body_truncated、request_body_bytes 及对应参数。单条和批量写入共用此修复，统一为 39 列、39 个连续占位符和 39 个参数。
- 保留已有请求时间、密钥前缀、耗时、上游错误及取消分类；新增回归断言，即使调用方仍携带旧 replay 字段也不再写入这些已删除列，尾部字段不能错位。生产代码只改 ops_repo.go，不改数据库迁移、表结构、余额、用量计费、请求重试或其他线路。
- 更新 BPS 文档和来源说明，解除前轮共享 SQL 数据库验证阻塞；保持此前移植与其他未提交改动。版本保持 0.2.8，未构建、推送、部署镜像，未连接生产数据库。

### Testing

- 修复前：CI=1 go test -tags=integration ./internal/repository -run ^TestOpsClientCancellationMetricsAndDuration$ -count=1 -v 明确失败，错误为 pq: column "request_body" of relation "ops_error_logs" does not exist；证据：/tmp/sub2api-ops-sql-fix.69YsBn/integration-before.log。
- 修复后：带 unit 标签的定向错误日志测试通过，覆盖已删字段禁止写入、列/占位符/参数一致、显式 0 状态、nil/0/正数/负数耗时及单条/批量调用；证据：/tmp/sub2api-ops-sql-fix.69YsBn/unit-targeted.log。首次新增测试将旧字段按值而非指针构造的编译问题已修正，最终测试通过。
- 后端 go test ./... 通过；go test -tags=unit ./internal/repository -count=1 全部通过；go test -race ./internal/repository -run "OpsError|OpsClientCancellation" -count=1 通过。证据：/tmp/sub2api-ops-sql-fix.69YsBn/backend-full-final.log、repository-unit.log、race.log。
- CI=1 go test -tags=integration ./internal/repository -run "OpsClientCancellationMetricsAndDuration|Basispoints" -count=1 -v 通过。真实 PostgreSQL/Redis、完整现有迁移下验证单条/批量实际写入、时间与密钥前缀、耗时读回、取消分类、SLA/趋势/分布/日小时聚合，以及 BPS 403 分组、标记、并发与缓存回归。证据：/tmp/sub2api-ops-sql-fix.69YsBn/integration-after-retry.log。
- 数据库修复后首次复验因临时 testcontainers/ryuk 端口未映射而在启动阶段失败；未修改 Docker 配置、未跳过测试，原命令重试后通过。失败日志保留：/tmp/sub2api-ops-sql-fix.69YsBn/integration-after.log。所有数据库写入均限于测试容器。
- git diff --check 及本轮增量 git apply --reverse --check 通过。无前端源码修改，本轮未重复前端构建；不将本地验证表述为生产已生效，也不自动补回历史丢失日志。

### Notes

- backend/internal/repository/ops_repo.go：删除三个已废弃的 INSERT 目标及参数，将占位符尾部从 $42 对齐至 $39。
- backend/internal/repository/ops_repo_args_test.go：将共用参数数目断言调整为 39，保留显式 0 上游状态验证。
- backend/internal/repository/ops_repo_duration_test.go：增加已删列禁止回归及尾部参数位置验证，更新单条/批量耗时参数断言。
- backend/internal/repository/ops_repo_cancellation_integration_test.go：用仍携带旧 replay 字段的输入验证真实写入成功，并核对单条/批量记录的时间与密钥前缀。
- docs/OPENAI_OAUTH_BASISPOINTS.md：将前轮共享 SQL 阻塞更新为已确认修复且完成真实数据库验证，说明部署后生效且不补历史。
- backend/internal/service/basispoints/NOTICE.md：注明经批准的共享 SQL 兼容修复及真实数据库通过，不涉及 schema 或计费变更。
- progress.md：仅追加本轮复现、修复、验证及回滚记录，保留前轮失败证据不改写。
- 本轮修改前逐文件快照：/tmp/sub2api-ops-sql-fix.69YsBn/before。可执行回滚：在 /Users/wangqiang/Project/codex/sub2api-wq 先执行 git apply --reverse --check /tmp/sub2api-ops-sql-fix.69YsBn/sql-fix.patch，检查通过后执行 git apply --reverse /tmp/sub2api-ops-sql-fix.69YsBn/sql-fix.patch。本轮只验证检查，未执行回滚。补丁只包含上述 6 个源码/测试/文档文件，不回退历史或本轮 progress.md；SHA-256 02de51e6332b92cf7362d913892f2f54da317c6c692e7c241550001992f8c9db。
- 此回滚会恢复已确认存在的错误日志写入缺陷，非必要不应部署回滚结果；若需连同前轮 BPS 移植一并回退，应先回退本轮 SQL 修复再检查前轮补丁。审计文件位于 /tmp，清理前需保存补丁与日志。

## 2026-09-27 - Task: 为选用 Basispoints 的 OAuth 账号补齐 Chat Completions 桥接

### What was done

- 按用户对 Chat Completions → BPS 协议桥接的明确批准，先复现该入口仍使用默认 Codex 地址的问题，再让 OpenAI OAuth/Setup Token 账号选用 Basispoints 时复用既有 BPS 转发、认证、图片处理、工具恢复、并发准入和错误策略。默认 Codex、API Key、其他渠道与 WebSocket 分流保持不变；不修改数据库、余额、账号设置或重试次数。
- Chat 请求先转换为 Responses，再进入同一 BPS 适配链路；模型映射只执行一次，保留请求别名、计费模型、原始/实际推理强度和真实响应 ID，max 在 BPS 上仍降为 xhigh。客户端分别收到 Chat Completions JSON 或 chunk SSE，不泄露 BPS 原生工具名称；终态-only 文本补齐流式正文，已有增量不重复输出。
- 复用缓存创建转普通输入开关：内部保留原始用量，只有 BPS 的下游 Chat usage 做对应归零。终态/纠错失败保留已取得用量且不触发 Codex 终态重放；发送前取消不请求上游，流式写入断开保留原有受限收尾行为。原 BPS 429 直接返回、初始 5xx 有界 failover 等策略未调整。
- 同步使用文档并保留此前未提交代码与 SQL 修复，版本保持 0.2.8。本轮未构建/推送/部署镜像，未使用真实 OAuth Token 请求上游，未连接生产数据库。

### Testing

- 修复前新增测试失败，明确显示预期 Basispoints URL、实际 Codex URL，覆盖 OAuth/Setup Token 与流式/非流式四种组合；证据：/tmp/sub2api-bps-chat.vX8gVE/before.log。测试 fixture 初始类型不匹配已修正，最终红灯来自真实路线不符，不是编译失败。
- go test ./internal/service -run '^TestBasispointsChatCompletions' -count=1 -v 通过，10 个测试组、33 个子场景：请求与返回协议、一次模型映射、max→xhigh、原始响应 ID、工具往返/大整数、缓存开关与真实计费入口、失败用量、HTTP 错误策略、默认线路隔离、取消/断连、增量正文/Responses 形状兼容、图片关闭/忽略/HTTPS 中转/原生附件，以及纠错 429 不重放。证据：chat-focused-final.log。
- 现有 BPS、ForwardAsChatCompletions、Chat 消费器/缓冲读错定向回归通过，证据：gateway-regression.log。初版桥接测试识别到仅终态正文未输出的流式缺口，已修复并保留回归；中间证据：bridge-basic.log。
- 后端标准全量 go test ./... 通过，证据：backend-full.log。
- go test -race ./internal/service -run 'Test(Basispoints|ForwardAsChatCompletions|HandleChat|ChatCompletionsBuffered|OpenAI.*Basispoints|NormalizeOpenAIBasispoints)' -count=1 通过；go test -race ./internal/service/basispoints ./internal/pkg/apicompat -count=1 全部通过。证据：gateway-race.log、protocol-race.log。
- 上述日志均位于 /tmp/sub2api-bps-chat.vX8gVE。gofmt -l 检查本轮三个 Go 文件无输出，git diff --check 通过。未修改前端和 SQL，本轮未重复前端构建或数据库集成；模拟上游验证不代表生产已部署或账号权限已验证。
- 本轮增量 git apply --reverse --check 通过，仅检查未执行。生成回滚补丁时一次末尾空白裁剪导致校验失败，已恢复完整 diff 上下文后再次校验通过，不影响源码测试。

### Notes

- backend/internal/service/openai_gateway_chat_completions.go：新增选择 BPS 的 Chat 入口分流；回程保持 Chat JSON/SSE，补齐终态正文、响应 ID、失败用量和下游缓存计数处理，仅对实际 BPS 路径生效。
- backend/internal/service/openai_basispoints_forward.go：保留原 Responses 入口封装，共用完整 BPS 转发流程并允许选择 Chat 响应消费者，避免二次模型映射或复制网络处理。
- backend/internal/service/openai_basispoints_chat_completions_test.go：新增 33 个子场景验证两种账号、协议输出、工具、用量、图片、错误策略及默认路径隔离。
- docs/OPENAI_OAUTH_BASISPOINTS.md：说明两个 HTTP 入口的使用方式、格式、兼容边界及部署后验证要求。
- progress.md：仅末尾追加本轮实现、验证及增量回滚记录。
- 本轮开始时四个已有目标文件的独立快照位于 /tmp/sub2api-bps-chat.vX8gVE/before，包含被既有忽略规则排除的文档；未以 Git HEAD 替换用户先前未提交内容。新增测试没有旧版本。
- 可执行回滚：在 /Users/wangqiang/Project/codex/sub2api-wq 执行 git apply --reverse --check /tmp/sub2api-bps-chat.vX8gVE/chat-bridge.patch，确认无冲突后执行 git apply --reverse /tmp/sub2api-bps-chat.vX8gVE/chat-bridge.patch。补丁只含本轮两个源码文件、新增测试和文档，不回退 progress.md 历史；SHA-256 e1ae93b9177f6b00cc7e4bbbb8f56b93e9f27edf7d78ca81d61f98f3b4e13be6。本轮未执行回滚；若后续修改重叠，须先处理冲突，不强制覆盖。临时审计目录清理前应另行保存补丁与日志。

## 2026-09-27 - Task: 构建并推送包含 Basispoints Chat 桥接的双架构镜像

### What was done

- 基于当前完整本地工作区完成生产镜像构建并推送 Docker Hub，保留全部既有未提交改动，未拉取或合并远程代码。
- 发布 iotwq/china-api:latest 与固定标签 iotwq/china-api:0.2.8-20260927-224808，均包含 linux/amd64、linux/arm64；版本保持 0.2.8，构建标识 f867e4f9bd2f-dirty，构建时间 2026-09-27T14:48:08Z。
- 本次镜像包含此前完成的独立 BPS v2.8.18–v2.8.19 修复、共享错误日志 INSERT 兼容修复与 Chat Completions → Basispoints 桥接；本轮未再改业务逻辑。
- 两个远端标签共同指向索引摘要 sha256:de124ae99296846d3647464bc73de9b751a1a71c28d68faa28e54b26c411952e。
- amd64 镜像摘要：sha256:477e71864d1852148e0e3ea6ab51f27f7809f36a029253166a80c261f3b88655；arm64 镜像摘要：sha256:8155024db90ffd298fbf368b2a6e4ec491b555727a9fcc20ef3d7eba7906b081。

### Testing

- Docker Buildx 双架构生产构建与推送退出码为 0；前端语言完整性检查、vue-tsc 类型检查、Vite 生产打包及两个架构的 Go 编译均通过。
- 独立查询 Docker Hub 的 latest 和固定标签，确认两者索引摘要一致，运行平台包含 linux/amd64 与 linux/arm64；其余 unknown/unknown 条目为构建证明，不是额外运行架构。
- 分别按精确架构摘要拉取镜像，并使用 --network none、--read-only、--user 1000:1000、--entrypoint /app/sub2api 运行 -version；两个临时容器均退出 0 并自动删除，返回一致的版本、构建标识与构建时间。未启动业务服务或连接数据库。
- 检查本地镜像元数据，两份产物的 Architecture、Os 与 RepoDigests 均匹配对应平台和远端摘要。
- 本轮未重复前一任务已通过的后端全量与竞态源码测试，生产构建和隔离版本冒烟不替代部署后的真实业务请求验证。
- 构建、发布、远端核验与冒烟证据位于 /tmp/sub2api-image-bps-chat-20260927.AM0Twk，包含 build.log、metadata.json、release.txt、remote-latest.txt、remote-release.txt、image-amd64.json、image-arm64.json 及两个 version 日志。

### Notes

- 改动文件清单：progress.md，仅在末尾追加本轮发布、验证和回滚记录；未修改源码、版本、配置、数据库、Dockerfile 或部署文件。
- 使用既有 codex-multiarch builder；构建与远端查询使用已有本机代理，仅通过命令级 HTTP_PROXY/HTTPS_PROXY 设置，未修改系统或 Docker 全局代理、凭据及登录状态。
- 本轮未部署或重启线上服务。上线时需在部署环境拉取新镜像并重建 sub2api 容器，再验证实际请求与用量记录。
- 发布前 latest 摘要为 sha256:895349847c65764b82a45e871ff2a39a6666b9ab3a16335ba7252987d64d6dc1。需要回滚远端标签时可执行 docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:895349847c65764b82a45e871ff2a39a6666b9ab3a16335ba7252987d64d6dc1；随后由部署环境重新拉取和重建容器。本轮未执行回滚。
- 本轮前日志快照为 /tmp/sub2api-image-bps-chat-20260927.AM0Twk/progress-before.md；若仅回退发布记录，应只删除本轮追加区段，保留其余历史及后续记录。临时审计目录清理前应另行保存发布证据。

## 2026-09-28 - Task: 经批准同步原项目 0.2.9 并保留本站协议与计费保护

### What was done

- 用户明确批准本次协议、调度及计费变更后，将 Wei-Shaw/sub2api 的 70 个新增提交、117 个文件合并至 9a62841fd124d026cf3694fcf9b79e98addcdbdc；本地合并 274d55cd6050f9c18479dd9de2609b026e9657c8，版本 0.2.9。
- 纳入 Responses/WS/工具协议、客户端断连 499、额度暂停及重查退避、模型白名单、图片定价和长上下文账号统计、CC Switch、模型广场与前端弹窗等修复。无新数据库迁移、依赖或 Dockerfile 改动。
- 基线解决一个 KeysView 冲突，恢复定制解决三个文件冲突；同时保留本地 OAuth 并发保护、Basispoints Responses/Chat 及图片/用量逻辑、视频恢复退款、监控和聊天等功能，Codex 模板不恢复 model_catalog_json。
- 保留本站缺价报错规则，不照搬上游 Free Fast 缺价零元兜底；补入相应缺价回归。图片价格留空继承目录价、显式 0 免费的上游修复正常纳入，不处理历史账务。
- 为新 Gemini 断连测试配置真实可解析的测试模型价格；补齐此前 BPS 图片设置的七个接口快照字段，未修改相关生产接口或默认设置。
- 保存全部原有定制、原索引、暂存/未暂存补丁、文件哈希和忽略文档。完整快照包含 4444 个现存文件及 6 个原有删除状态；613 个原有未提交条目无丢失、暂存分类保留，220 个原有未跟踪条目内容无变化。

### Testing

- 前端 pnpm test:run：354 个测试文件、2716 项通过；pnpm run lint:check 与 pnpm run build（含翻译完整性、vue-tsc 和 Vite）通过。保留已有 Node localStorage、Browserslist 和大 bundle 提示。
- 后端 go test -mod=readonly -tags=unit ./... 最终通过；service 包最终执行约 191 秒，全部其他包成功或无测试。首轮 Gemini 取消两场景、BPS 设置两快照和 Free Fast 缺价场景失败，已按本轮兼容约定修复后全量复测，首轮日志保留。
- go test -mod=readonly -race -tags=unit ./internal/service ./internal/handler ./internal/repository ./internal/pkg/apicompat ./internal/pkg/tlsfingerprint -run 'Test(.*Basispoints.*|.*AccountTraffic.*|.*CodexSignal.*|.*OpenAIOAuthProtection.*|.*LocalTLS.*|.*ContextWindow.*|.*ClientCancel.*|.*ClientDisconnect.*|.*FreeOpenAIFast.*|.*MissingPricing.*|.*ChannelImage.*|.*LongContext.*|.*AlphaSearch.*|.*ImageSettlement.*|.*Minimax.*|.*VideoCompensator.*|.*DolaVideo.*|.*ViraleeVideo.*)' -count=1 -timeout=8m 通过；其中 apicompat 未命中该定向筛选，另以 go test -mod=readonly -race -tags=unit ./internal/pkg/apicompat ./internal/service/basispoints -count=1 验证两个完整包通过。
- CGO_ENABLED=0、embed 后端生产编译通过；产物 /tmp/sub2api-sync-20260928.IGQa93/sub2api 的 -version 返回 0.2.9、274d55cd6-local、2026-09-28。只执行版本输出，未启动业务服务。
- 三个 Compose 模板使用 --env-file /dev/null 及合成测试密码离线解析，空密码/普通密码共六种组合通过，Redis exec 参数和健康检查认证一致；未启动 Redis，未将该语法验证描述为容器运行验证。
- 数据库集成未通过且未运行用例：CI=true 的隔离 repository 集成测试在 Docker 前置检查失败，当前 desktop-linux socket 无法连接 daemon。未启动或重启 Docker，未改用生产数据库。Docker 可用后需补跑文档中列出的计费原子性、图片结算、BPS 账号和取消统计集成命令。
- git diff --check、git diff --cached --check 通过；本轮手改 Go 文件 gofmt -l 无输出。文件核对 missing、unexpected、changedUntracked、stagingChanges 均为空。progress.md 原历史保持字节前缀不变。
- 证据目录 /tmp/sub2api-sync-20260928.IGQa93：backend-unit.log、backend-unit-final.log、fixture-regression.log、backend-race.log、protocol-race.log、frontend-tests.log、frontend-lint.log、frontend-build.log、backend-build-final.log、version-final.log、compose-validation.json、billing-integration.log、preservation-final.json。

### Notes

- 上游 117 个变更文件及逐项提交来源见 docs/upstream-sync-20260928-files.md，清单覆盖本轮基线合并全部文件，不代表覆盖本站定制。
- frontend/src/views/user/KeysView.vue：解决基线冲突，保留 chinaapi 名称并采用上游规范化的 CC Switch 余额查询脚本。
- backend/internal/service/openai_upstream_transport_error.go：同时保留上游客户端取消快速退出和本地并发限制分类。
- frontend/src/components/keys/UseKeyModal.vue：保留删除 model_catalog_json 的既定行为，移除合并引入的孤立常量。
- frontend/src/components/keys/__tests__/UseKeyModal.spec.ts：统一 Windows 配置不含 model_catalog_json 的断言，保留目录文件路径展示校验。
- backend/internal/handler/gemini_client_cancel_test.go：为上游新增测试补有效定价依赖，验证真实转发后取消为 499，同时保留缺价前置拦截。
- backend/internal/server/api_contract_test.go：补现有 BPS 图片配置字段的两份响应预期，不改生产逻辑。
- backend/internal/service/openai_gateway_usage.go：保留本站缺价拒绝结算，Standard 复算的价格错误不得静默忽略。
- backend/internal/service/openai_gateway_record_usage_test.go：保留 Free Fast 缺价报错且不落零元成功记录的回归。
- docs/UPSTREAM_SYNC.md：追加本轮使用、图片价格变化、验证缺口和完整恢复步骤；docs/upstream-sync-20260928-files.md：新增来源及兼容处理清单；progress.md：仅末尾追加本轮记录。
- 初次 stash apply --index 因上游上下文变化失败，改为恢复内容、解决冲突，再反向应用原未暂存补丁到索引；两个重叠文件定向重建索引，四份文档恢复 intent-to-add。原业务定制未整体提交，新测试兼容改动留在工作区。
- 恢复点为 codex/pre-upstream-sync-20260928（f867e4f9bd2f6bcf1b2c6a67b20d5124cc86748a）、stash fc12914844697fdf73fcae92f8b9ceb5225afab2 及审计目录的 before.tgz/index。可执行独立恢复命令见 docs/UPSTREAM_SYNC.md 本日章节，包含原始定制、忽略文档和四份文档的暂存意图；不覆盖当前工作区，不删除 stash。临时审计资料应在清理前另行备份。
- 本轮未构建或推送镜像、未推送 GitHub、未部署或修改线上数据，未进行真实上游付费请求，也未执行回滚。

## 2026-09-28 - Task: 同步 ranxi2001/sub2api v2.8.20 与 v2.9.0 的 Basispoints 修复

### What was done

- 对照 `v2.8.20`（`dc01c71b758e8c24bd76ea5f7ddb089fc76481d3`）完成独立 Basispoints 429 修复适配：主请求、原生附件和请求尚未输出内容时的加密恢复请求可以复用现有账号切换；已输出内容后的工具纠正 429 只冷却当前账号的 BPS 路由，不重放原请求。
- 增加 BPS 专用进程内冷却，解析合法 `Retry-After` 秒数/HTTP 日期，缺失时使用现有 429 默认回避设置；普通、负载感知和刷新后的选号路径均跳过冷却账号，候选耗尽时通过现有处理器返回脱敏 429，不改 Codex 全局额度、账号健康、全局限流状态或计费。
- 复核 `v2.9.0`（`b3e494dbd`、`211202d4a`）的质量规则自动开启/关闭 BPS。该部分依赖当前仓库不存在的 `account_quality_bps` repository/service、Pelican 质量计划/探针、迁移及管理员前端，未直接复制；Mihomo 代理托管按此前范围另行处理。账号编辑中的手动 Basispoints 选择保持不变。
- 更新 BPS 测试断言，使服务层 429 按可切换的 `UpstreamFailoverError` 验证；同步更新协议、来源和未移植边界文档。版本保持 `0.2.9`，没有构建、推送或部署镜像，也未请求真实 OAuth 上游。

### Testing

- `cd backend && go test -mod=readonly -tags=unit ./internal/service ./internal/handler` 通过；最终日志：`/tmp/bps-full-test-final.log`。
- `cd backend && go test -mod=readonly -tags=unit ./internal/service -run 'TestBasispoints(NativeUploadFailureStopsWithoutQuotaWrite|NativeAttachmentFailureDoesNotGenerate|AutoDisableOn403|ChatCompletionsHTTPErrorPolicy|InvalidEncryptedContentRetryIsBounded|UpstreamErrorsPreserveStatusAndFailover)' -count=1` 通过。
- `cd backend && go test -mod=readonly -race -tags=unit ./internal/service -run 'TestBasispoints|TestExcelBPS|TestOpenAI.*BPS' -count=1 -timeout=8m` 通过，日志：`/tmp/bps-service-race-final.log`。
- `cd backend && go test -mod=readonly -race -tags=unit ./internal/handler -run 'TestBasispoints|TestInferenceFailoverExhaustion' -count=1 -timeout=8m` 通过，日志：`/tmp/bps-handler-race-final.log`。
- `gofmt` 已检查本轮修改的 Go 测试文件；`git diff --check` 与 `git diff --cached --check` 均通过。

### Notes

- `backend/internal/service/openai_basispoints_ratelimit.go`：新增 BPS 429 专用错误、Retry-After 解析、账号级路由冷却和冷却判断。
- `backend/internal/service/openai_basispoints_forward.go`：主 BPS 请求 429 failover、安全响应及恢复请求处理。
- `backend/internal/service/openai_basispoints_attachments.go`：附件上传 429 进入同一账号切换路径。
- `backend/internal/service/openai_basispoints_repair.go`：工具纠正 429 仅冷却、不重放用户请求。
- `backend/internal/service/openai_gateway_scheduling.go`、`backend/internal/service/openai_account_scheduler.go`、`backend/internal/service/openai_account_runtime_block_fastpath.go`：在三条 OpenAI 选号/运行时路径接入 BPS 冷却隔离。
- `backend/internal/handler/openai_gateway_handler.go`：候选耗尽后返回脱敏 BPS 429；`openai_gateway_credential_failover_test.go` 增加安全 429 回归。
- `backend/internal/service/openai_basispoints_attachments_test.go`、`openai_basispoints_completion_test.go`、`openai_basispoints_auto_disable_test.go`、`openai_basispoints_chat_completions_test.go`、`openai_basispoints_encrypted_test.go`、`openai_basispoints_test.go`：按 v2.8.20 的 failover 语义更新断言。
- `docs/OPENAI_OAUTH_BASISPOINTS.md`：记录 v2.8.20 行为和 v2.9.0 质量自动启停未移植的边界；`backend/internal/service/basispoints/NOTICE.md`：记录来源和适配范围；`docs/UPSTREAM_SYNC.md`：追加同步记录。
- 回滚方式：本轮文件与当前大量未提交工作区交叠，不能使用 `git reset` 或整仓 checkout。若需回退，先保存当前工作区，再按本记录列出的文件从本轮前快照或对应 patch 逐文件恢复；恢复前用 `git diff --check` 和定向测试确认，不触碰其他本地定制。此前本轮源码改动前快照和补丁位于 `/tmp/sub2api-bps-sync-20260928-224200/`，不执行自动回滚。

## 2026-09-29 - Task: 更新 OpenAI 图片接口文档

### What was done

- 在 OpenAI 图片接口文档中加入 `gpt-image-2.5-sunburst` 和 `gpt-image-2.5-flare`，说明它们与 `gpt-image-2` 使用相同的生成、编辑和参考图调用方式。
- 补充透明背景参数 `background: "transparent"` 与 `output_format: "png"`，并分别给出 JSON 和 multipart 调用示例及使用约束。
- 明确质量参数范围：`gpt-image-2` 支持 `low`、`medium`、`high`；两个 2.5 模型支持 `low`、`medium`、`high`、`xhigh`、`max`。
- 本轮只更新用户接口文档和对应前端文档测试，未修改后端路由、协议、计费或模型能力判断。

### Testing

- `cd frontend && pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts` 通过：1 个测试文件、1 项测试通过。
- `git diff --check` 通过；仅保留既有测试环境的 localStorage 与 Browserslist 提示。

### Notes

- 改动文件清单：`frontend/src/views/user/ApiDocsView.vue`（用户可见 OpenAI 图片模型、参数和示例）；`docs/API_DOCS.md`（接口文档 Markdown）；`frontend/src/views/user/__tests__/ApiDocsView.spec.ts`（新模型、质量和透明背景文档断言）；`progress.md`（追加本轮记录）。
- 回滚方式：当前工作区包含大量其他未提交改动，不能使用 `git reset` 或整仓 checkout；如需回滚，仅保存当前工作区后逐文件恢复上述三份文档文件的本轮内容，并保留其他改动。

## 2026-09-29 - Task: 移植 BPS 原生图片附件兼容修复

### What was done

- 经用户批准仅筛选适用的 BPS 修复，移植 ranxi2001/sub2api v2.9.4 的 f6666ab43：在完整校验、历史翻译后将消息内附件引用收敛为 type/file_id，覆盖新上传、历史及工具转入的图片。HTTPS 图片、内联工具截图及客户端原始历史保持原状。
- 未引入 BPS 生图、Mihomo 或质量运维依赖，未合并仍待确认的 Wei-Shaw 0.2.10。

### Testing

- 新增回归在修复前稳定失败：历史/新上传附件仍带 detail 和 client_metadata；原生重复附件缓存场景也失败。
- 修复后 `cd backend && go test -mod=readonly ./internal/service/basispoints -count=1` 通过。
- `cd backend && go test -mod=readonly -tags=unit ./internal/service -run 'TestBasispoints.*(Attachment|ToolImages|AgentImage|NativeImage)' -count=1` 通过。
- 前后验证日志保存在 /tmp/sub2api-bps-fixes-20260929.JAlgyz，分别为 image-reference-before.log、image-protocol-after.log、image-service-after.log。

### Notes

- `backend/internal/service/basispoints/images.go`：新增已验证附件字段归一化；`basispoints/request.go`：在历史翻译完成后调用归一化。
- `backend/internal/service/basispoints/image_file_reference_test.go`：新增附件字段、非法输入及缓存复用回归；`basispoints/history_message_images_test.go`、`basispoints/tool_images_test.go`：更新附件精确字段断言并保留 HTTPS 原值断言。
- `backend/internal/service/openai_basispoints_attachments_test.go`、`openai_basispoints_tool_images_test.go`：验证实际发送的附件仅含两个允许字段。
- `docs/OPENAI_OAUTH_BASISPOINTS.md`：追加本轮附件兼容行为；`progress.md`：仅末尾追加记录。
- 回滚点：/tmp/sub2api-bps-fixes-20260929.JAlgyz/before.tgz（源码及两份 BPS 文档/日志的施工前快照），同目录保存原 index 和暂存/未暂存补丁。可用 `restore_dir=$(mktemp -d /tmp/sub2api-bps-restore.XXXXXX); tar -xzf /tmp/sub2api-bps-fixes-20260929.JAlgyz/before.tgz -C "$restore_dir"` 提取后仅恢复本条列出的既有文件，并删除本轮新增测试；不要整仓 reset 或覆盖其他定制。

## 2026-09-29 - Task: 移植 BPS 流式工具重新生成边界修复

### What was done

- 适配 e0ba95a48：流式客户端收到可见内容后禁止未知工具触发整段重新生成；非流式缓冲和已知工具的局部参数纠正不变。
- 增加协议和实际转发路径验证，不影响默认 Codex 路由、计费和账号切换预算。

### Testing

- 修复前 `go test -mod=readonly ./internal/service/basispoints -run 'TestUnknownToolRegenerationStopsAfterVisibleContent|TestReasoningSummaryArrivesBeforeTerminal' -count=1` 稳定复现可见内容后仍重新生成。
- 修复后 `cd backend && go test -mod=readonly ./internal/service/basispoints -count=1` 通过。
- `cd backend && go test -mod=readonly -tags=unit ./internal/service -run 'TestBasispoints.*(UnknownTarget|ToolCorrection|ChatCompletions)' -count=1` 通过。
- 日志：/tmp/sub2api-bps-fixes-20260929.JAlgyz/visible-stream-{before,protocol-after,service-after}.log。

### Notes

- `backend/internal/service/basispoints/request.go`：记录客户端原始 stream；`basispoints/stream.go`：可见输出后禁止整体再生成。
- `basispoints/stream_repair_boundary_test.go`：新增流式修复边界测试；`basispoints/tool_repair_test.go`：保留流式已知工具纠正回归；`backend/internal/service/openai_basispoints_repair_test.go`：新增实际转发只请求一次的回归。
- `docs/OPENAI_OAUTH_BASISPOINTS.md`：记录可见输出的行为边界；`progress.md`：追加本轮闭环记录。
- 回滚：使用 /tmp/sub2api-bps-fixes-20260929.JAlgyz/before.tgz 按上一条命令提取，仅恢复本条列出的既有源码文件，移除新增 stream_repair_boundary_test.go；request.go 与前一项附件修复重叠，单项回滚应仅撤回 clientStream 字段及赋值。不要整仓 reset。

## 2026-09-29 - Task: 适配 BPS 流内错误分类与取消终态

### What was done

- 适配 b6617bf62 的安全错误分类与取消终态，覆盖 Responses、本站已有 Chat Completions 桥接及工具纠正读取器。取消不再被当作成功或等待 EOF。
- 区分真实 HTTP 错误和已受理请求的流内错误：后者不重放；流内 429 仅冷却 BPS，401/403 不触发 Codex 鉴权/403 账号动作，Ops 保留实际 HTTP 200 和语义状态。
- 保留原始失败用量、无用量不入账、缓存计量开关、已有 HTTP 拒绝重试、客户端取消及 compact SSE 保活行为；没有修改账务口径或数据库。

### Testing

- 移植前新增服务回归复现取消被视为成功、错误终态丢失及状态分类不足，日志 failure-service-before.log。
- 协议包全量通过：go test -mod=readonly ./internal/service/basispoints -count=1。
- BPS 全部定向通过：go test -mod=readonly -tags=unit ./internal/service -run 'TestBasispoints|TestExcelBPS|TestOpenAI.*BPS' -count=1。
- 补充原始缓存用量、零用量失败、工具纠正、流内失败与 Chat 双模式回归通过：go test -mod=readonly -tags=unit ./internal/service -run 'TestBasispoints.*Failure|TestBasispointsToolCorrection|TestBasispointsTruncated' -count=1。
- 初次扩大验证发现直接照搬上游的非 nil 失败结果会破坏本站“无用量不入账”约束；已恢复既有返回策略并复测通过，未删除或弱化原有断言。证据位于 /tmp/sub2api-bps-fixes-20260929.JAlgyz/failure-{protocol-after,bps-after,bps-final,usage-final}.log。

### Notes

- backend/internal/service/basispoints/upstream_failure.go、upstream_failure_test.go：新增安全分类与停止读取的回归；basispoints/stream.go、tool_repair.go、unknown_tool_repair.go：识别取消并保留类型化纠正错误。
- backend/internal/service/openai_basispoints_failure.go、openai_basispoints_failure_test.go：新增本地安全终态诊断与 Responses/Chat/用量回归。
- backend/internal/service/openai_basispoints_repair.go：非流式返回正确分类且保留用量；openai_basispoints_forward.go：保留错误终态与既有用量/取消返回规则。
- backend/internal/service/openai_gateway_response_handling.go、openai_gateway_chat_completions.go、openai_gateway_messages.go：仅在实际 BPS 端点下处理失败语义及保存共享读取器会丢失的状态；默认 Codex 分支不变。
- backend/internal/service/openai_basispoints_chat_completions_test.go：更新为安全分类错误而非原始上游文案；docs/OPENAI_OAUTH_BASISPOINTS.md、progress.md：补充行为说明及本次验证。
- 回滚点为 /tmp/sub2api-bps-fixes-20260929.JAlgyz/before.tgz；用前文 mktemp/tar 命令提取后逐文件恢复上述既有文件，新增的两组 failure.go/failure_test.go 可单独移除。stream.go 与上一项重叠，单项回滚仅撤回取消/错误分类 hunk，保留可见内容边界。不得整仓 reset 或覆盖其他本地定制。

## 2026-09-29 - Task: 确认 BPS 调度修复适用性并记录移植边界

### What was done

- 复核 a6e51622a、0c74d2d7a：本站没有导致问题的 BPS 独立池裁剪、OpenAITurnAdmission 或 Mihomo 托管代理层，因此没有复制新调度政策或空壳错误类型。
- 为现有混合候选池补充满载回退、抢位失败回退、全忙等待及排除账号不被重新选中的回归；未改变默认 TopK、优先级、账号绑定或并发上限。
- 文档记录三项已适配修复及未适配的依赖/新功能，保留来源和回滚依据。补充 Chat 取消缺少嵌套 status 时仍以事件类型为准的验证。

### Testing

- cd backend && go test -mod=readonly -tags=unit ./internal/service -run 'TestBasispointsMixedPoolKeepsAvailableCapacity|TestBasispointsFailurePreservesRawCacheUsage' -count=20 通过。
- 初始直接使用上游 TopK=1 和 BPS 固定等待优先断言与本站共享预算/加权等待策略不符；测试改为明确 TopK=2 覆盖两个候选、全忙时只断言合法等待账号。未修改生产调度代码以迎合上游断言。
- 验证日志：/tmp/sub2api-bps-fixes-20260929.JAlgyz/mixed-pool-regression-verified.log；此前失败轮次日志保留。

### Notes

- backend/internal/service/openai_basispoints_scheduler_compat_test.go：新增混合候选池四场景回归。
- backend/internal/service/openai_gateway_chat_completions.go：BPS 缓冲失败按已识别终态而非仅嵌套 status 判断；openai_basispoints_failure_test.go：覆盖取消无嵌套 status 和原始缓存用量。
- docs/OPENAI_OAUTH_BASISPOINTS.md：说明适用性及账号/重试边界；docs/UPSTREAM_SYNC.md：固定版本来源与未引入依赖；backend/internal/service/basispoints/NOTICE.md：补充上游修复和行为参考归属；progress.md：本次闭环记录。
- 回滚：提取 /tmp/sub2api-bps-fixes-20260929.JAlgyz/before.tgz 到新的临时目录，逐文件核对本轮 hunk，删除新建 openai_basispoints_scheduler_compat_test.go；文档和共享源码不得整文件覆盖后续独立改动。版本和 index 未改变，未部署或变更数据库。

## 2026-09-29 - Task: BPS 适用修复最终回归与交付检查

### What was done

- 完成三项适用修复的最终验证，确认 Responses/Chat 共用路径、图片附件、流式工具修复、HTTP 429 切换、取消与原始用量行为兼容。
- 对照施工前快照检查实际增量，未修改 handler/repository 源码、数据库、前端、部署文件、版本或 Git index。progress.md 施工前字节前缀完全一致，仅追加历史。
- 没有进行整仓 Git merge、commit/push、Docker 构建/推送或部署；没有使用真实账号请求上游。现有其他定制保持原状。

### Testing

- 最终全量定向包：cd backend && go test -mod=readonly -tags=unit ./internal/service ./internal/handler ./internal/service/basispoints ./internal/pkg/apicompat -count=1 -timeout=10m。全部通过，分别耗时 193.223s / 50.132s / 1.156s / 1.308s；日志 full-unit-verified.log。
- 协议完整竞态：go test -mod=readonly -race ./internal/service/basispoints -count=1 -timeout=5m，通过；protocol-race-final.log。
- BPS/账号流量保护竞态：go test -mod=readonly -race -tags=unit ./internal/service -run 'TestBasispoints|TestExcelBPS|TestOpenAI.*BPS|TestAccountTraffic|TestCodexTraffic' -count=1 -timeout=8m，通过；service-race-verified.log。
- handler 切换竞态：go test -mod=readonly -race -tags=unit ./internal/handler -run 'TestBasispoints|TestInferenceFailoverExhaustion|TestOpenAIGateway.*Failover' -count=1 -timeout=8m，通过；handler-race-verified.log。
- 混合池/取消用量回归 -count=20 通过；mixed-pool-regression-verified.log。日志均在 /tmp/sub2api-bps-fixes-20260929.JAlgyz/。
- gofmt -l 对本轮新增/修改 Go 文件无输出；git diff --check、git diff --cached --check 通过。与解包快照进行 git diff --no-index --check 无空白错误（退出 1 表示内容差异）；handler/repository 递归对比无差异；cmp index-before .git/index 一致；progress.md 原历史前缀 cmp 一致。
- 没有前端或数据库改动，未重复前端构建或迁移测试；模拟上游通过不能证明真实 BPS 服务、账号权限或输出质量。

### Notes

- docs/OPENAI_OAUTH_BASISPOINTS.md：补最终验证和使用边界；docs/UPSTREAM_SYNC.md：补版本筛选验证结果；backend/internal/service/basispoints/NOTICE.md：来源段落排版；progress.md：最终验证记录。源码和测试文件清单见同日三项修复及调度兼容检查记录，无额外源码改动。
- 回滚点：/tmp/sub2api-bps-fixes-20260929.JAlgyz/before.tgz，SHA256 eb62e76f2f24c9477939587240c9c0f4945174da85f0e977c4c7da67943d1008；原 HEAD 274d55cd6050f9c18479dd9de2609b026e9657c8，index-before 与 staged-before.patch/unstaged-before.patch 同目录保留。
- 可执行提取命令：restore_dir=$(mktemp -d /tmp/sub2api-bps-restore.XXXXXX); tar -xzf /tmp/sub2api-bps-fixes-20260929.JAlgyz/before.tgz -C "$restore_dir"。只按前述文件清单核对和恢复本轮 hunk，新增测试和 failure 文件单独移除；不要整仓覆盖、reset 或回写整个 progress.md 历史。备份在临时目录，系统清理前应另行保管。

## 2026-09-29 - Task: 统一 OpenAI 图片模型规格文案

### What was done

- 按用户要求将 gpt-image-2.5-sunburst 和 gpt-image-2.5-flare 的模型列表规格统一为「1K / 2K / 4K；最多 16 张参考图」，与 gpt-image-2 一致。
- quality 参数中关于两个 2.5 模型额外支持 xhigh、max 的说明保持不变；不再在模型规格栏重复该说明。仅调整文档和文档测试，未修改后端模型能力、接口、计费或其他功能。

### Testing

- cd frontend && pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts 通过：1 个文件、1 项测试；新增断言逐行确认三个模型规格完全一致，且 quality 参数说明仍然存在。日志：/tmp/sub2api-image-doc-20260929.CMGSLs/test.log。
- git diff --check、git diff --cached --check 通过；保留现有 localStorage/Browserslist 环境提示，未因此更新依赖。

### Notes

- frontend/src/views/user/ApiDocsView.vue：替换两个 2.5 模型规格文案；docs/API_DOCS.md：同步三个模型统一规格的说明；frontend/src/views/user/__tests__/ApiDocsView.spec.ts：补充规格和 quality 参数的精确回归；progress.md：仅追加本轮记录。
- 回滚点：/tmp/sub2api-image-doc-20260929.CMGSLs/before.tgz 包含本轮前上述四个文件。可用 restore_dir=$(mktemp -d /tmp/sub2api-image-doc-restore.XXXXXX); tar -xzf /tmp/sub2api-image-doc-20260929.CMGSLs/before.tgz -C "$restore_dir" 提取后，仅恢复本轮文案及测试增量；不要整仓 reset 或覆盖其他本地改动，不回写 progress.md 历史。
- 未构建、推送镜像或部署。

## 2026-09-30 - Task: SD2.0 接口文档补充 ViralDance 2.5

### What was done

- 根据截图和 ../gpt_image_playground 无限画布的实际构造请求、轮询代码及测试，确认画布显示名 viraldance2.5 对应请求模型 viraldance2.5-30，区别于现有 dola-viraldance2.5。按原有风格加入 SD2.0 模型列表。
- 补充 4–30 秒、720p、三种画幅、5000 字符提示词、30 图/10 视频/10 音频以及首尾帧限制；明确 aspect_ratio、async、start_image_url/end_image_url、video_reference/audio_reference 和中文素材标签的用法。
- 新增文生视频、首尾帧与多模态参考、查询下载三个示例；每 5 秒查询任务，完成后读取顶层 url，不使用 Firefly 的 /content。示例仅使用本站 API/Key 占位符与 example.com 素材地址，不包含实际上游地址或凭据。
- 只读确认现有后端已识别该模型及参数透传路径。本轮未改后端路由、协议、计费、版本或其他模型功能，保留上一轮图片模型文案。

### Testing

- cd frontend && pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts 通过：1 个文件、1 项测试。新增断言核对实际模型名、规格、素材字段、三份示例、5 秒轮询及顶层 url；解析示例 JSON 并确认未混入 933 参数，保留原有六分类和其他模型回归。日志：/tmp/sub2api-sd25-doc-20260929.X9C0tK/test.log。
- cd frontend && pnpm exec eslint src/views/user/ApiDocsView.vue src/views/user/__tests__/ApiDocsView.spec.ts 通过，无输出；未使用 --fix。日志：/tmp/sub2api-sd25-doc-20260929.X9C0tK/eslint.log。
- 针对修改文件的 git diff --check 以及与本轮备份比较的 git diff --no-index --check 无空白错误；已检查实际增量只落在 SD2.0 文档、对应示例及测试中。
- 保留已有 localStorage/Browserslist 环境提示，未更新依赖；未请求真实付费接口，未进行生产部署验收。

### Notes

- frontend/src/views/user/ApiDocsView.vue：在 SD2.0 分类新增实际模型、参数限制、三份 curl 示例及调用注意事项。
- frontend/src/views/user/__tests__/ApiDocsView.spec.ts：补模型规格、JSON 示例可解析性、参数隔离及轮询下载说明的定向回归。
- docs/API_DOCS.md：同步 ViralDance 2.5 调用、素材限制和参考依据。
- progress.md：仅追加本轮实现和验证记录；本轮跨越午夜，按完成时本地日期记录。
- 回滚点：/tmp/sub2api-sd25-doc-20260929.X9C0tK/before.tgz 包含本轮前上述四个文件，已经包含上一轮完成的图片文案。可用 restore_dir=$(mktemp -d /tmp/sub2api-sd25-doc-restore.XXXXXX); tar -xzf /tmp/sub2api-sd25-doc-20260929.X9C0tK/before.tgz -C "$restore_dir" 提取核对，仅恢复本轮新增内容；不要整仓 reset、覆盖后续改动或回写 progress.md 历史。临时备份应在系统清理前另行保管。
- 未构建、推送镜像或部署，未提交或推送 Git。

## 2026-09-29 - Task: SD2.0 文档统一展示完整模型名 viraldance2.5-30

### What was done

- 按用户要求，在模型列表、参数说明、示例标题和注意事项中统一展示 viraldance2.5-30，移除画布简称说明；模型规格、请求示例和其他模型保持不变。

### Testing

- cd frontend && pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts 通过：1 个文件、1 项测试，确认完整模型名、无画布别名以及原有调用示例。日志：/tmp/sub2api-sd25-name.iMr9jY/test.log。
- cd frontend && pnpm exec eslint src/views/user/ApiDocsView.vue src/views/user/__tests__/ApiDocsView.spec.ts 通过，无输出；未使用 --fix。

### Notes

- frontend/src/views/user/ApiDocsView.vue：统一新增模型的用户可见名称，不改实际请求示例。
- frontend/src/views/user/__tests__/ApiDocsView.spec.ts：同步名称断言，并断言不展示画布说明。
- docs/API_DOCS.md：同步完整模型名称的展示规范。
- progress.md：仅追加本轮验证记录。
- 回滚点：/tmp/sub2api-sd25-name.iMr9jY/before.tgz 包含本轮前上述四个文件；提取后仅恢复本轮文案与测试差异，保留已有模型文档和 progress.md 历史。未改后端、计费或版本，未构建镜像、部署或推送。

## 2026-09-29 - Task: 构建并推送包含最新接口文档的双架构镜像

### What was done

- 从当前完整本地工作区构建并推送 Docker Hub，包含最新接口文档、统一的 viraldance2.5-30 名称及既有全部定制；未拉取或合并远程代码，未改版本和业务逻辑。
- 发布 iotwq/china-api:latest 与固定标签 iotwq/china-api:0.2.9-20260930-004733，支持 linux/amd64 和 linux/arm64。版本 0.2.9，构建标识 274d55cd6050-dirty，构建时间 2026-09-29T16:47:33Z；固定标签时间使用本机 Asia/Shanghai 时区。
- 两个标签共同指向索引摘要 sha256:2c57dab4720feaed7b298ff711eef94f6761722a98822323db1f3f69c1dc85db。amd64 摘要为 sha256:a912d56bc9405cfdeb40a7a94d43fcbf8ca3fb63daa492d3bfb62ec8a4d782d4，arm64 摘要为 sha256:aa7296efae3beb1245d5d36da95218d71ee2b6266ed07977385e435e9ef7e9cd。

### Testing

- Docker Buildx 使用既有 codex-multiarch builder 完成双架构生产构建与推送，退出 0。镜像内翻译完整性 3 项、vue-tsc 类型检查、Vite 生产构建及两个架构 Go 编译均通过。
- cd frontend && pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts 通过：1 文件、1 项测试。
- 独立查询 Docker Hub 的 latest 和固定标签，索引摘要均与 metadata.json 一致，运行架构包含 linux/amd64、linux/arm64；unknown/unknown 条目为构建证明。
- 两个架构均按精确 manifest 从 Docker Hub 拉取；分别在 --network none、--read-only、--user 1000:1000 的临时容器执行 -version，均退出 0，版本、commit、构建时间一致，镜像 Architecture、Os 与 RepoDigests 均匹配。
- 两个架构均在隔离容器中检查嵌入的前端资源，确认存在 viraldance2.5-30 文生视频和首尾帧与多模态参考标题；检查退出 0，临时容器自动删除。
- 首次远端标签查询及首次 arm64 拉取遇到网络 EOF，重试成功；arm64 首次失败日志已保留。前端既有 localStorage、Browserslist 和 bundle 提示不影响本次通过，未修改依赖。
- 证据目录：/tmp/sub2api-image-sd25.EAwISr，包含 build.log、metadata.json、remote-latest.txt、remote-release.txt、api-docs-test.log、pull-*.log、image-*.json、version-*.log。未重复前轮后端全量/竞态测试，未调用真实付费模型或验证线上业务。

### Notes

- 改动文件清单：progress.md，仅末尾追加本轮发布、验证和回滚记录；源码、Dockerfile、配置、版本、数据库及 Git index 未修改。
- 未部署、重启线上服务或推送 Git；部署端需自行拉取并重建应用容器，发布成功不代表运行中的服务已经更新。
- 发布前 latest 摘要为 sha256:c043d8437d29682266e77865691d3caaec67433f10856cf89702c12cae8e2507。恢复标签可执行 docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:c043d8437d29682266e77865691d3caaec67433f10856cf89702c12cae8e2507；或在部署配置固定该旧摘要。本轮未执行回滚，标签回滚不会撤销数据库变更。
- 日志回滚点：/tmp/sub2api-image-sd25.EAwISr/progress-before.md；若撤销本轮记录，只移除本轮追加段，保留所有历史及后续改动。临时审计目录清理前请另行保存。

## 2026-09-30 - Task: 合并 Wei-Shaw/sub2api 0.2.10 并保留本站定制

### What was done

- 按用户对协议、调度、风控与计费更新的明确确认，合并指定的 33 个提交、118 个文件至上游 a60a29549；本地合并提交为 59fcda4ea34f087600114c74f31806df6160fc97，版本更新为 0.2.10。未扩大到其他上游或发布任务。
- 保留 BPS 的 Responses/Chat 桥接、缓存计量、附件及专用限流冷却，以及 OAuth/TLS/并发保护、图片转换、视频恢复退款和结算。两处装配冲突纳入 Claude 重置额度查询，不恢复已移除的 292/332 打票。
- 先验证完整归档，合并后恢复原 621 项脏状态及暂存/未暂存边界；原有定制未被整体提交。额外只修正风控缓存回退测试的初始化竞态，不改变线上业务逻辑。

### Testing

- `cd frontend && pnpm test:run`：355 文件、2734 项测试通过；`pnpm run lint:check` 与 `pnpm run build` 通过，后者包含翻译检查、类型检查和 Vite 生产构建。保留已有大 chunk 提示。
- `cd backend && go test -mod=readonly -tags=unit ./... -count=1 -timeout=10m` 全量通过；后续仅修改一个测试装配，不改业务源码。
- 首轮定向 race 检出既有 TestRecordCyberPolicyEvent_RuntimeSnapshotRefreshFailureKeepsStaleScope 在后台 worker 启动后写 runtimeCacheTTL 的竞态；修正后该用例 `-race -count=50` 通过。
- 最终 `go test -mod=readonly -race -tags=unit ./internal/service ./internal/handler ./internal/repository -run 'Basispoints|BPS|CodexProtection|AccountTraffic|Composite|CyberPolicy|CyberAllowlist|UsageBilling|Settlement|OpenAI.*(Cancel|WS)' -count=1 -timeout=10m` 通过；`go test -mod=readonly -race -tags=unit ./internal/service/basispoints ./internal/pkg/apicompat ./internal/pkg/tlsfingerprint ./internal/service/openai_ws_v2 -count=1 -timeout=5m` 四包完整 race 通过。
- 数据库使用 Testcontainers 临时 PostgreSQL 18.1 / Redis 8.4，并设置 CI=true 防止环境不可用时静默跳过；未连接生产数据库。初次合并执行时两条全局统计断言受到其他测试已提交数据影响：预期请求数 2 实际 7，预期活跃用户 1 实际 5；日志保留，不计为通过。
- 将 `^TestUsageLogRepoSuite$` 与 `TestUsageLogRepository|TestUsageBillingRepositoryApply|TestImageSettlement|TestMoveBasispointsOn403|TestBasispoints403Marker|TestMediaBillingReview|TestVideoBillingReview` 分别放入全新的测试容器运行，两组均通过，覆盖新增消费排名、原子结算和退款幂等；未修改 SQL 或弱化断言。此结论不表示全部集成测试混跑已通过。
- `go run -mod=readonly github.com/google/wire/cmd/wire diff ./cmd/server` 无差异；CGO_ENABLED=0 的 embed 生产编译及 `-version` 通过，输出 0.2.10 / 59fcda4ea-local。
- 备份校验覆盖 4465 个原文件；合并保留检查覆盖 4486 个路径状态，上游范围外的原内容与权限除本轮测试修正/文档追加外不变。最终校验包含原暂存分类、无未解冲突、差异空白检查和 progress.md 原历史字节前缀完整性。
- 证据位于 `/tmp/sub2api-sync-20260929.6L1lvp`：frontend-tests.log、frontend-lint.log、frontend-build.log、backend-unit.log、race-core.log（首轮失败）、race-fixture-repeat.log、race-core-final.log、race-protocol.log、integration.log（混跑失败）、integration-usage-isolated.log、integration-billing-isolated.log、wire-diff.log、backend-build.log、backend-version.log、preservation-final.json。

### Notes

- 本轮无数据库迁移、依赖或 Dockerfile 新改动；未构建/推送镜像、未推送 Git、未部署、未重启业务服务、未请求真实付费模型。回归不能替代线上真实账号验证。
- 回滚点：分支 `codex/pre-upstream-sync-20260929-a60a29549`（原 HEAD 274d55cd6050f9c18479dd9de2609b026e9657c8），stash `9929e1441d046dfbf1fa7650d04e4856b6fdb630` 及审计目录 `before.tgz`；完整不覆盖式恢复命令见 docs/UPSTREAM_SYNC.md 的 2026-09-30 段落。禁止整仓 reset 或清理 stash 覆盖定制，临时备份需另行保管。
- 测试修正单独撤销时，仅将 content_moderation_cyber_test.go 本轮两行恢复为归档版本；文档与日志仅撤销本轮追加段，不改历史或后续内容。
- 本轮附加文件清单：`backend/internal/service/content_moderation_cyber_test.go` 修正测试初始化竞态；`docs/UPSTREAM_SYNC.md` 追加同步、验证和恢复步骤；`docs/upstream-sync-20260930-files.md` 新增提交来源和逐文件说明；`progress.md` 仅追加本轮记录。
- 上游 118 个文件逐项清单如下，两个 wiring 文件含本站兼容处理；完整 33 个提交列表见 docs/upstream-sync-20260930-files.md。

| 文件 | 本轮变更 |
| --- | --- |
| `backend/cmd/server/VERSION` | 同步上游 chore: sync VERSION to 0.2.10 [skip ci]。 |
| `backend/cmd/server/wire_gen.go` | 同步上游 feat(claude): expose sanitized native reset credit status；fix(gateway): resolve composite routes for websocket aliases。同时保留本站服务装配及旧打票移除状态。 |
| `backend/internal/domain/constants.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/domain/constants_test.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/handler/admin/account_handler.go` | 同步上游 feat(claude): expose sanitized native reset credit status。 |
| `backend/internal/handler/admin/claude_reset_handler.go` | 同步上游 feat(claude): expose sanitized native reset credit status。 |
| `backend/internal/handler/admin/dashboard_handler.go` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `backend/internal/handler/admin/dashboard_handler_cache_test.go` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `backend/internal/handler/admin/dashboard_query_cache.go` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `backend/internal/handler/admin/dashboard_snapshot_v2_handler.go` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `backend/internal/handler/admin/setting_handler.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/handler/admin/setting_handler_audit.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/handler/admin/setting_handler_update.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/handler/dto/settings.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/handler/openai_cyber_allowlist.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/handler/openai_cyber_allowlist_test.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/handler/openai_gateway_handler.go` | 同步上游 feat: add risk control user allowlist；fix(gateway): resolve composite routes for websocket aliases。 |
| `backend/internal/handler/openai_gateway_handler_test.go` | 同步上游 fix(gateway): resolve composite routes for websocket aliases。 |
| `backend/internal/handler/openai_gateway_ws_composite_test.go` | 同步上游 fix(gateway): resolve composite routes for websocket aliases。 |
| `backend/internal/handler/openai_ws_v2_passthrough_cyber_test.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/handler/wire.go` | 同步上游 feat(claude): expose sanitized native reset credit status；fix(gateway): resolve composite routes for websocket aliases。同时保留本站服务装配及旧打票移除状态。 |
| `backend/internal/pkg/apicompat/anthropic_responses_test.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/pkg/apicompat/anthropic_to_responses_response.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/pkg/apicompat/responses_to_anthropic_request.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/pkg/apicompat/types.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/pkg/claude/constants.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/pkg/claude/constants_model_test.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/pkg/claude/effort_catalog.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/pkg/claude/effort_catalog_test.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/repository/content_moderation_repo.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/repository/content_moderation_repo_test.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/repository/usage_log_repo_integration_test.go` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `backend/internal/repository/usage_log_repo_trend.go` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `backend/internal/server/api_contract_test.go` | 同步上游 feat: add risk control user allowlist；feat(dashboard): toggle recent usage between tokens and spending。 |
| `backend/internal/server/routes/admin.go` | 同步上游 feat(claude): expose sanitized native reset credit status。 |
| `backend/internal/service/account_usage_service.go` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `backend/internal/service/account_usage_service_batch_test.go` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `backend/internal/service/anthropic_chat_stream_usage_test.go` | 同步上游 fix(gateway): avoid ambiguous late-cache input subtraction；fix(gateway): normalize streamed Anthropic usage；fix(gateway): forward received chat stream usage。 |
| `backend/internal/service/antigravity_gateway_compat_stream.go` | 同步上游 fix(antigravity): keep compat streams alive before first content。 |
| `backend/internal/service/antigravity_gateway_compat_test.go` | 同步上游 fix(antigravity): keep compat streams alive before first content。 |
| `backend/internal/service/bedrock_request.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/bedrock_request_test.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/billing_service.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/billing_service_test.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/claude_reset_credits.go` | 同步上游 fix(claude): harden reset credit projection and cell state (review R2)；fix(claude): explicitly close status HTTP response；feat(claude): expose sanitized native reset credit status。 |
| `backend/internal/service/claude_reset_credits_test.go` | 同步上游 fix(claude): harden reset credit projection and cell state (review R2)；feat(claude): expose sanitized native reset credit status。 |
| `backend/internal/service/content_moderation.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/cyber_policy_allowlist.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/cyber_policy_allowlist_test.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/dashboard_service.go` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `backend/internal/service/domain_constants.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/gateway_anthropic_apikey_passthrough_test.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/gateway_anthropic_passthrough.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/gateway_bedrock.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/gateway_claude_oauth_body.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/gateway_count_tokens.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/gateway_forward.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/gateway_forward_as_chat_completions.go` | 同步上游 feat: support Claude Sonnet 5.5；fix(gateway): normalize streamed Anthropic usage；fix(gateway): forward received chat stream usage。 |
| `backend/internal/service/gateway_forward_as_chat_completions_test.go` | 同步上游 fix(gateway): forward received chat stream usage。 |
| `backend/internal/service/gateway_forward_as_responses.go` | 同步上游 feat: support Claude Sonnet 5.5；fix(gateway): avoid ambiguous late-cache input subtraction；fix(gateway): normalize streamed Anthropic usage。 |
| `backend/internal/service/gateway_forward_as_responses_test.go` | 同步上游 feat: support Claude Sonnet 5.5；fix(gateway): avoid ambiguous late-cache input subtraction；fix(gateway): normalize streamed Anthropic usage。 |
| `backend/internal/service/gateway_request.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/gateway_request_test.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/gateway_sonnet55_toolset_beta_test.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/gateway_tool_rewrite.go` | 同步上游 fix(anthropic): rewrite tool names in one body pass。 |
| `backend/internal/service/gateway_tool_rewrite_test.go` | 同步上游 fix(anthropic): rewrite tool names in one body pass。 |
| `backend/internal/service/gateway_upstream_request.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/openai_account_scheduler.go` | 同步上游 fix(gateway): constrain account model route ownership。 |
| `backend/internal/service/openai_account_scheduler_test.go` | 同步上游 fix(gateway): enforce composite model ownership in legacy scheduling；fix(gateway): constrain account model route ownership。 |
| `backend/internal/service/openai_codex_models_service.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/openai_codex_models_service_test.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/openai_gateway_anthropic_native_pump_test.go` | 同步上游 fix(gateway): forward received chat stream usage。 |
| `backend/internal/service/openai_gateway_chat_completions_anthropic_native.go` | 同步上游 feat: support Claude Sonnet 5.5；fix(gateway): normalize streamed Anthropic usage；fix(gateway): forward received chat stream usage。 |
| `backend/internal/service/openai_gateway_cn_fixes_test.go` | 同步上游 fix(gateway): avoid ambiguous late-cache input subtraction；fix(gateway): normalize streamed Anthropic usage。 |
| `backend/internal/service/openai_gateway_messages_anthropic_native.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/openai_gateway_responses_anthropic_native.go` | 同步上游 feat: support Claude Sonnet 5.5；fix(gateway): normalize streamed Anthropic usage。 |
| `backend/internal/service/openai_gateway_scheduling.go` | 同步上游 fix(gateway): enforce composite model ownership in legacy scheduling。 |
| `backend/internal/service/openai_images_json_keepalive_test.go` | 同步上游 fix(anthropic): rewrite tool names in one body pass。 |
| `backend/internal/service/openai_ws_forwarder_ingress.go` | 同步上游 fix(gateway): resolve composite routes for websocket aliases。 |
| `backend/internal/service/pricing_service.go` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `backend/internal/service/risk_control_allowlist_test.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/setting_cyber_allowlist_test.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/setting_gateway_runtime.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/setting_parse.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/setting_service.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/setting_update.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/settings_view.go` | 同步上游 feat: add risk control user allowlist。 |
| `backend/internal/service/wire.go` | 同步上游 feat(claude): expose sanitized native reset credit status。 |
| `backend/resources/model-pricing/model_prices_and_context_window.json` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `docs/COMPOSITE_GROUPS.md` | 同步上游 fix(gateway): resolve composite routes for websocket aliases。 |
| `frontend/src/api/admin/claudeResetCredits.ts` | 同步上游 fix(claude): harden reset credit projection and cell state (review R2)；feat(accounts): show Claude reset credit status on demand。 |
| `frontend/src/api/admin/dashboard.ts` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `frontend/src/api/admin/settings.ts` | 同步上游 feat: add risk control user allowlist。 |
| `frontend/src/components/account/AccountStatusIndicator.vue` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `frontend/src/components/account/AccountUsageCell.vue` | 同步上游 fix(accounts): address review of Claude reset credit count；feat(accounts): align Claude reset credit query with Codex reset count button；feat(accounts): show Claude reset credit status on demand。 |
| `frontend/src/components/account/BulkEditAccountModal.vue` | 同步上游 fix(account): prevent whitelist model mapping conflicts。 |
| `frontend/src/components/account/ClaudeResetCreditsCell.vue` | 同步上游 fix(claude): harden reset credit projection and cell state (review R2)；fix(accounts): address review of Claude reset credit count；feat(accounts): align Claude reset credit query with Codex reset count button；feat(accounts): show Claude reset credit status on demand。 |
| `frontend/src/components/account/CreateAccountModal.vue` | 同步上游 fix(account): prevent whitelist model mapping conflicts。 |
| `frontend/src/components/account/EditAccountModal.vue` | 同步上游 fix(account): prevent whitelist model mapping conflicts。 |
| `frontend/src/components/account/ModelWhitelistSelector.vue` | 同步上游 fix(account): prevent whitelist model mapping conflicts。 |
| `frontend/src/components/account/__tests__/AccountStatusIndicator.spec.ts` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `frontend/src/components/account/__tests__/ClaudeResetCreditsCell.spec.ts` | 同步上游 fix(claude): harden reset credit projection and cell state (review R2)；fix(accounts): address review of Claude reset credit count；feat(accounts): align Claude reset credit query with Codex reset count button；feat(accounts): show Claude reset credit status on demand。 |
| `frontend/src/components/account/__tests__/EditAccountModal.spec.ts` | 同步上游 fix(account): prevent whitelist model mapping conflicts。 |
| `frontend/src/components/account/__tests__/ModelWhitelistSelector.spec.ts` | 同步上游 fix(account): prevent whitelist model mapping conflicts。 |
| `frontend/src/components/keys/UseKeyModal.vue` | 同步上游 feat: support Claude Sonnet 5.5；fix: omit Codex model catalog for OpenAI groups；feat(frontend): hide unsupported clients for Claude Code-only groups。 |
| `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts` | 同步上游 feat: support Claude Sonnet 5.5；fix: omit Codex model catalog for OpenAI groups；feat(frontend): hide unsupported clients for Claude Code-only groups。 |
| `frontend/src/composables/__tests__/useModelWhitelist.spec.ts` | 同步上游 feat: support Claude Sonnet 5.5；fix(account): prevent whitelist model mapping conflicts。 |
| `frontend/src/composables/useModelWhitelist.ts` | 同步上游 feat: support Claude Sonnet 5.5。 |
| `frontend/src/i18n/locales/en/admin/accounts.ts` | 同步上游 fix(accounts): address review of Claude reset credit count；feat(accounts): align Claude reset credit query with Codex reset count button；feat(accounts): show Claude reset credit status on demand；fix(account): prevent whitelist model mapping conflicts。 |
| `frontend/src/i18n/locales/en/admin/overview.ts` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `frontend/src/i18n/locales/en/admin/settings.ts` | 同步上游 feat: add risk control user allowlist。 |
| `frontend/src/i18n/locales/zh/admin/accounts.ts` | 同步上游 fix(accounts): address review of Claude reset credit count；feat(accounts): align Claude reset credit query with Codex reset count button；feat(accounts): show Claude reset credit status on demand；fix(account): prevent whitelist model mapping conflicts。 |
| `frontend/src/i18n/locales/zh/admin/overview.ts` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `frontend/src/i18n/locales/zh/admin/settings.ts` | 同步上游 feat: add risk control user allowlist。 |
| `frontend/src/views/admin/DashboardView.vue` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `frontend/src/views/admin/SettingsView.vue` | 同步上游 feat: add risk control user allowlist。 |
| `frontend/src/views/admin/__tests__/DashboardView.spec.ts` | 同步上游 feat(dashboard): toggle recent usage between tokens and spending。 |
| `frontend/src/views/user/KeysView.vue` | 同步上游 feat(frontend): hide unsupported clients for Claude Code-only groups。 |

## 2026-10-01 - Task: 合并主项目近期提交至 0.2.11

### What was done

- 从 Wei-Shaw/sub2api 的 upstream/main 合并至提交 42bc7f6cf，版本更新为 0.2.11，本地合并提交为 365b2e6547975a27c841e9887d0abda2fcb76cf3。
- 保留本站 Basispoints、OAuth/TLS/并发保护、mandatory usage record fallback、图片/视频结算与余额冻结等定制；服务装配同时接入上游 idempotency 和 Claude reset redeem。
- 修复合并后的 Codex 配置再次引用已删除 model_catalog_json 常量的问题，恢复本站默认只生成远程目录配置；图片请求沿用既有数据库冻结，跳过重复 Redis 在途预占。完整上游路径清单见 docs/upstream-sync-20261001-files.md。

### Testing

- 前端 355 个测试文件、2777 项断言通过；lint、生产构建通过。
- 后端 Wire diff、CGO_DISABLED embed 编译通过；全量 unit 在跳过 TestInflightEstimate_AccountMappingNoDBAndBoundedMemory 后通过。该内存阈值测试单独连续 3 次通过，但嵌入全量 suite 时两次受堆基线波动影响，未宣称无条件全量通过。
- 图片/在途预占定向 handler 回归通过。
- PostgreSQL/Redis 隔离集成中 BPS、账号、用量日志套件通过；计费/图片结算套件首轮同组运行出现窗口断言波动，单测重跑和同组重跑均通过。
- git diff --check、git diff --cached --check 通过，git ls-files -u 无未合并项；上游相对基线未新增 backend/migrations 迁移。

### Notes

| 文件 | 本轮改动 |
| --- | --- |
| `backend/cmd/server/wire_gen.go` | 解决合并冲突，保留本站装配并接入上游 idempotency/Claude reset redeem。 |
| `backend/internal/handler/gateway_handler.go` | 保留 mandatory usage record fallback。 |
| `backend/internal/handler/openai_gateway_handler.go` | 保留 OpenAI 计费任务同步兜底及在途预留交接。 |
| `backend/internal/handler/openai_images.go` | 已有数据库图片冻结时跳过通用 Redis 在途预占，避免重复占用。 |
| `frontend/src/components/keys/UseKeyModal.vue` | 移除失效的本地目录模式常量引用，默认仅生成远程目录配置。 |
| `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts` | 更新远程目录配置断言，确认不再生成 model_catalog_json。 |
| `docs/UPSTREAM_SYNC.md` | 追加 0.2.11 合并范围、兼容处理、验证边界和恢复点。 |
| `docs/upstream-sync-20261001-files.md` | 保存本轮上游提交与完整变更路径。 |
| `progress.md` | 追加本轮施工、验证和回滚记录。 |

回滚点：合并前恢复分支 `codex/pre-upstream-sync-20260930-v0.2.11`，stash `2cd1d88310fbc76fb411f30d3c3988ff6cdff087`，完整备份 `/tmp/sub2api-sync-20260930.S4H93B/before.tgz`。需要回退时在新目录应用恢复分支与 stash，不在当前工作区执行 `git reset --hard`，也不删除未跟踪文件。

## 2026-10-02 - Task: 合并主项目近期提交至 0.2.12

### What was done

- 将 Wei-Shaw/sub2api 的 upstream/main 从 42bc7f6cf 同步至 ae501cc22，吸收 32 个提交、152 个上游变更路径；版本 0.2.12，本地合并提交 8d433e144c596eda109f197a66b062d7bf7bdea0。
- 合并 TypeSafe 原生接口、充值优惠阶梯、邮箱/订单安全修复、Grok CLI 身份头和界面改进；保留本站 Basispoints、OAuth/TLS/并发保护、312 信号、图片/视频结算、计费任务兜底、社区和余额查询。
- 处理账号创建基线及 Jev/Grok 定价测试冲突；更新平台配额测试并补赠金字段/TypeSafe 约束的 PostgreSQL 校验。原 618 项工作区状态和暂存分类保留。

### Testing

- 前端首轮 3 项失败均来自新增平台后旧五平台断言；修正后完整测试 357 文件、2803 项通过。lint:check、最后改动文件 ESLint、check:i18n、vue-tsc 和生产构建通过。
- 后端 go test -mod=readonly -tags=unit ./... 全量通过（本轮没有排除任何内存阈值用例）；Wire diff 无差异，CGO_ENABLED=0 embed 编译和 --version（0.2.12）通过。
- service/handler/admin/repository/routes/securityaudit/typesafe/xai 的 BPS、OAuth 并发、312、在途预占、媒体、充值阶梯、邮箱验证、System One、Grok 和匿名订单限流定向 race 通过。
- CI=true 隔离 PostgreSQL 18.1/Redis 8.4 验证迁移幂等/并发、赠金字段默认值、TypeSafe 约束、平台配额和 BPS 状态；另一独立数据库的计费原子性/幂等、图片结算和视频审账通过。
- 原 4502 个快照路径内容/模式比对无范围外意外变化；618 项原工作区状态及暂存分类无变化。git diff --check、git diff --cached --check 通过，Git 索引无未合并项。
- 完整验证与备份证据：/tmp/sub2api-sync-20261002.DEsw9a，首轮失败日志保留。未使用生产数据库或真实 OAuth/付费上游请求。

### Notes

| 文件 | 本轮改动 |
| --- | --- |
| `.github/SECURITY.md` | 同步上游 docs(security): add dev@sub2api.org as fallback reporting email；docs: add security policy with private vulnerability reporting。 |
| `README.md` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；fix(typesafe): align System One validation with upstream schema；feat: add native TypeSafe Jev System One support。 |
| `README_CN.md` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；fix(typesafe): align System One validation with upstream schema；feat: add native TypeSafe Jev System One support。 |
| `backend/cmd/server/VERSION` | 同步上游 chore: sync VERSION to 0.2.12 [skip ci]。 |
| `backend/ent/migrate/schema.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/ent/mutation.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/ent/paymentorder.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/ent/paymentorder/paymentorder.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/ent/paymentorder/where.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/ent/paymentorder_create.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/ent/paymentorder_update.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/ent/runtime/runtime.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/ent/schema/payment_order.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/ent/schema/user_platform_quota.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/domain/constants.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/admin/account_handler.go` | 同步上游 fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `backend/internal/handler/admin/account_handler_available_models_test.go` | 同步上游 fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `backend/internal/handler/admin/channel_handler.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/admin/channel_handler_test.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/admin/group_handler.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/admin/group_handler_platform_test.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/admin/payment_handler.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/handler/admin/setting_handler.go` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/handler/admin/setting_handler_recharge_bonus.go` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/handler/admin/setting_handler_update.go` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/handler/auth_oauth_pending_flow_test.go` | 同步上游 fix(email): atomic verify-code attempts and hashed single-use reset tokens。 |
| `backend/internal/handler/dto/recharge_bonus_tiers.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/handler/dto/settings.go` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/handler/endpoint.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/endpoint_test.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/gateway_handler.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/gateway_handler_chat_completions.go` | 同步上游 fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `backend/internal/handler/gateway_handler_responses.go` | 同步上游 fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `backend/internal/handler/gateway_models_retrieve_test.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/gateway_models_test.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/gateway_systemone.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/gateway_systemone_test.go` | 同步上游 fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/handler/payment_handler.go` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/handler/user_handler_test.go` | 同步上游 fix(email): atomic verify-code attempts and hashed single-use reset tokens。 |
| `backend/internal/model/error_passthrough_rule.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径。 |
| `backend/internal/model/error_passthrough_rule_test.go` | 同步上游 test(model): 错误透传平台列表守卫纳入 typesafe。 |
| `backend/internal/pkg/typesafe/client.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/pkg/typesafe/systemone.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；fix(typesafe): align System One validation with upstream schema；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/pkg/typesafe/systemone_test.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；fix(typesafe): align System One validation with upstream schema；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/pkg/xai/billing.go` | 同步上游 修复 Grok CLI 版本门槛并对齐官方无界面请求头。 |
| `backend/internal/pkg/xai/billing_test.go` | 同步上游 依据正常交互式 CLI 抓包修正 Grok 身份头；修复 Grok CLI 版本门槛并对齐官方无界面请求头。 |
| `backend/internal/pkg/xai/cli_identity.go` | 同步上游 依据正常交互式 CLI 抓包修正 Grok 身份头；修复 Grok CLI 版本门槛并对齐官方无界面请求头。 |
| `backend/internal/pkg/xai/cli_identity_test.go` | 同步上游 依据正常交互式 CLI 抓包修正 Grok 身份头；修复 Grok CLI 版本门槛并对齐官方无界面请求头。 |
| `backend/internal/repository/api_key_repo.go` | 同步上游 feat(keys): support sorting API keys by group name。 |
| `backend/internal/repository/api_key_repo_sort_test.go` | 同步上游 feat(keys): support sorting API keys by group name。 |
| `backend/internal/repository/email_cache.go` | 同步上游 fix(email): atomic verify-code attempts and hashed single-use reset tokens。 |
| `backend/internal/repository/email_cache_atomic_test.go` | 同步上游 fix(email): atomic verify-code attempts and hashed single-use reset tokens。 |
| `backend/internal/repository/http_upstream.go` | 同步上游 修复 Grok CLI 版本门槛并对齐官方无界面请求头。 |
| `backend/internal/repository/http_upstream_test.go` | 同步上游 依据正常交互式 CLI 抓包修正 Grok 身份头；修复 Grok CLI 版本门槛并对齐官方无界面请求头。 |
| `backend/internal/securityaudit/prompt_snapshot.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `backend/internal/securityaudit/prompt_snapshot_test.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `backend/internal/server/api_contract_test.go` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/server/router.go` | 同步上游 fix(payment): rate-limit anonymous public order verify endpoint。 |
| `backend/internal/server/routes/gateway.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/server/routes/gateway_model_allowlist_test.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/server/routes/payment.go` | 同步上游 fix(payment): rate-limit anonymous public order verify endpoint。 |
| `backend/internal/server/routes/payment_public_rate_limit_test.go` | 同步上游 fix(payment): rate-limit anonymous public order verify endpoint。 |
| `backend/internal/server/routes/prompt_audit_route_coverage_test.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/account.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/account_service.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/account_test_service.go` | 同步上游 fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `backend/internal/service/account_test_service_typesafe.go` | 同步上游 fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `backend/internal/service/account_test_service_typesafe_test.go` | 同步上游 fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `backend/internal/service/admin_account.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/admin_group.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/antigravity_gateway_gemini.go` | 同步上游 fix(antigravity): sanitize upstream error body before returning to client。 |
| `backend/internal/service/antigravity_upstream_error_sanitize.go` | 同步上游 fix(antigravity): sanitize upstream error body before returning to client。 |
| `backend/internal/service/antigravity_upstream_error_sanitize_test.go` | 同步上游 fix(antigravity): sanitize upstream error body before returning to client。 |
| `backend/internal/service/auth_service_email_bind_test.go` | 同步上游 fix(email): atomic verify-code attempts and hashed single-use reset tokens。 |
| `backend/internal/service/auth_service_register_test.go` | 同步上游 fix(email): atomic verify-code attempts and hashed single-use reset tokens。 |
| `backend/internal/service/billing_service.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/billing_service_test.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/channel_service.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/channel_service_test.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/composite_platform.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/composite_platform_test.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/content_moderation.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/content_moderation_input.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/content_moderation_systemone_test.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/domain_constants.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/email_service.go` | 同步上游 fix(email): atomic verify-code attempts and hashed single-use reset tokens。 |
| `backend/internal/service/email_service_reset_token_test.go` | 同步上游 fix(email): atomic verify-code attempts and hashed single-use reset tokens。 |
| `backend/internal/service/gateway_systemone.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/gateway_systemone_test.go` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略；fix(typesafe): align System One validation with upstream schema；feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/grok_upstream_headers.go` | 同步上游 依据正常交互式 CLI 抓包修正 Grok 身份头；修复 Grok CLI 版本门槛并对齐官方无界面请求头。 |
| `backend/internal/service/grok_upstream_headers_test.go` | 同步上游 依据正常交互式 CLI 抓包修正 Grok 身份头；修复 Grok CLI 版本门槛并对齐官方无界面请求头。 |
| `backend/internal/service/openai_gateway_grok.go` | 同步上游 依据正常交互式 CLI 抓包修正 Grok 身份头；修复 Grok CLI 版本门槛并对齐官方无界面请求头。 |
| `backend/internal/service/payment_config_service.go` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/service/payment_fulfillment.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/service/payment_order.go` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/service/payment_order_provider_snapshot_test.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/service/payment_recharge_bonus.go` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/service/payment_recharge_bonus_test.go` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/service/payment_service.go` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/internal/service/scheduler_snapshot_service.go` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `backend/internal/service/user_service.go` | 同步上游 fix(email): atomic verify-code attempts and hashed single-use reset tokens。 |
| `backend/migrations/241_add_payment_order_bonus_amount.sql` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `backend/migrations/241_add_typesafe_platform.sql` | 同步上游 fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `backend/migrations/typesafe_platform_migration_test.go` | 同步上游 fix(typesafe): 补齐平台迁移、审计覆盖、端点隔离与错误策略。 |
| `frontend/package.json` | 同步上游 升级 Axios 至 1.20.0 修复前端安全审计。 |
| `frontend/pnpm-lock.yaml` | 同步上游 升级 Axios 至 1.20.0 修复前端安全审计。 |
| `frontend/src/api/admin/settings.ts` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）；feat: add native TypeSafe Jev System One support。 |
| `frontend/src/api/admin/users.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/components/account/AccountPriorityCell.vue` | 同步上游 fix(admin): keep disabled priority stepper hidden until hover；feat(admin): inline quick-adjust stepper for account priority。 |
| `frontend/src/components/account/CreateAccountModal.vue` | 同步上游 fix(typesafe): 堵住校验解析差异与审计盲区，收紧模型列表口径；feat: add native TypeSafe Jev System One support。 |
| `frontend/src/components/account/EditAccountModal.vue` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/components/account/__tests__/AccountPriorityCell.spec.ts` | 同步上游 fix(admin): keep disabled priority stepper hidden until hover；feat(admin): inline quick-adjust stepper for account priority。 |
| `frontend/src/components/admin/payment/AdminOrderDetail.vue` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/components/admin/payment/AdminOrderTable.vue` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/components/admin/settings/RechargeBonusTierEditor.vue` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/components/admin/user/__tests__/UserPlatformQuotaModal.spec.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/components/keys/UseKeyModal.vue` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/components/payment/AmountInput.vue` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/components/payment/OrderTable.vue` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/components/payment/PaymentStatusPanel.vue` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/components/payment/__tests__/AmountInput.spec.ts` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/components/user/UserPlatformQuotaCell.vue` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/components/user/__tests__/UserPlatformQuotaCell.spec.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/components/user/dashboard/UserDashboardStats.vue` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/composables/useModelWhitelist.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/constants/__tests__/platforms.spec.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/constants/platforms.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/i18n/locales/en/admin/accounts.ts` | 同步上游 fix(i18n): correct the custom error codes warning text；feat(admin): inline quick-adjust stepper for account priority；feat: add native TypeSafe Jev System One support。 |
| `frontend/src/i18n/locales/en/admin/overview.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/i18n/locales/en/admin/settings.ts` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/i18n/locales/en/dashboard.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/i18n/locales/en/misc.ts` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/i18n/locales/zh/admin/accounts.ts` | 同步上游 fix(i18n): correct the custom error codes warning text；feat(admin): inline quick-adjust stepper for account priority；feat: add native TypeSafe Jev System One support。 |
| `frontend/src/i18n/locales/zh/admin/overview.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/i18n/locales/zh/admin/settings.ts` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/i18n/locales/zh/dashboard.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/i18n/locales/zh/misc.ts` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/types/index.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/types/payment.ts` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/utils/__tests__/rechargeBonus.spec.ts` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/utils/keyGroupProviders.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/utils/platformColors.ts` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/utils/rechargeBonus.ts` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/views/admin/AccountsView.vue` | 同步上游 feat(admin): inline quick-adjust stepper for account priority。 |
| `frontend/src/views/admin/ChannelsView.vue` | 同步上游 feat: add native TypeSafe Jev System One support。 |
| `frontend/src/views/admin/SettingsView.vue` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）；feat: add native TypeSafe Jev System One support。 |
| `frontend/src/views/admin/__tests__/SettingsView.spec.ts` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）；feat: add native TypeSafe Jev System One support。 |
| `frontend/src/views/user/KeysView.vue` | 同步上游 feat(keys): support sorting API keys by group name。 |
| `frontend/src/views/user/PaymentResultView.vue` | 同步上游 feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/views/user/PaymentView.vue` | 同步上游 feat(payment): 充值优惠阶梯支持折扣模式，快捷金额角标改为促销价签；feat(payment): 充值赠送阶梯（按金额档位额外赠送到账额度）。 |
| `frontend/src/views/user/__tests__/KeysView.spec.ts` | 同步上游 feat(keys): support sorting API keys by group name；feat: add native TypeSafe Jev System One support。 |
| `frontend/src/api/__tests__/settings.authSourceDefaults.spec.ts` | 配额默认值断言增加 TypeSafe，确认归一化与提交返回完整平台集合。 |
| `backend/internal/repository/migrations_schema_integration_test.go` | 补充新赠金字段默认值及 TypeSafe 两表平台约束的真实数据库验证。 |
| `docs/UPSTREAM_SYNC.md` | 追加本轮范围、兼容处理、验证和部署迁移影响。 |
| `docs/upstream-sync-20261002-files.md` | 保存全部来源提交、文件用途及独立目录恢复命令。 |
| `progress.md` | 追加本轮闭环记录和完整文件清单。 |

- 新增迁移 241_add_payment_order_bonus_amount.sql、241_add_typesafe_platform.sql；部署前备份数据库，空优惠阶梯不改变充值金额；升级前未使用的明文密码重置链接需重新申请。没有构建/推送镜像、推送 GitHub、部署或改动线上账务。
- 回滚点：分支 codex/pre-upstream-sync-20261002-v0.2.12（原 HEAD 365b2e6547975a27c841e9887d0abda2fcb76cf3）；stash 87dc77358c35808872863e42f720311ab6999b56；归档 /tmp/sub2api-sync-20261002.DEsw9a/before.tgz（SHA256 f54f0e7810a32f0a469048f677c65da6cbd511f3b480782bdbb09a90f0c49bd8）。
- 可执行恢复方式见 docs/upstream-sync-20261002-files.md：在不存在的新目录建立恢复分支工作树、stash apply --index，再解压完整归档；不对当前目录执行破坏性回退。已经迁移的生产数据库需单独评估回退，本轮未执行回滚。

## 2026-10-02 - Task: 同轮追加合并主项目计费修复至 3040209f2

### What was done

- 最终核对发现主项目继续更新，追加合并 ae501cc22 至 3040209f205472038c1ba745a1bedd2edd9053b1 的 6 个提交、5 个后端文件；最终本地合并提交 eda84f38231469c4968a8e574a1c26a221ce547d，版本仍为 0.2.12。本轮合计 38 个上游提交、157 个上游路径。
- 同步结算期间 API Key 删除后的继续结算修复及 TypeSafe 中转账单探测支持；保留本站原子用量/账务落账、媒体冻结和退款策略。
- 将上游删除 Key 的数据库回归改为覆盖本站实际原子结算接口，确认用量、扣款流水和幂等记录只产生一次；没有新增业务开关或迁移。

### Testing

- repository 与 service 两包完整 unit 通过；计费、图片/媒体和上游计费探测定向 race 通过。
- 隔离 PostgreSQL/Redis 中删除 Key 后的原子落账和重复请求幂等验证通过，计费/媒体同组最终回归全部通过。
- 增量后重新 embed 编译及 --version（0.2.12）通过；前端无新增修改，沿用本轮已通过的 357 文件/2803 项测试、lint 和生产构建。
- 原 4502 个快照路径最终比对无范围外意外变化，618 项原工作区状态及暂存分类保持一致；git diff --check、git diff --cached --check 通过，Git 无未合并索引。
- 增量证据同在 /tmp/sub2api-sync-20261002.DEsw9a 的 backend-followup-unit.log、followup-race.log、db-followup-billing-final.log、backend-build-final.log 和 preservation-final-followup.json。

### Notes

| 文件 | 本轮改动 |
| --- | --- |
| `backend/internal/repository/usage_billing_repo.go` | 正向结算时允许已删除 API Key 缺失其自身计数，用户和账号结算继续；真实 SQL 错误仍回滚。 |
| `backend/internal/repository/usage_billing_repo_integration_test.go` | 覆盖删除 Key 后的本站原子用量/流水落账及重试仅扣一次。 |
| `backend/internal/repository/usage_billing_repo_unit_test.go` | 导入删除 Key 不阻止扣费和真实 SQL 错误仍失败的用例。 |
| `backend/internal/service/upstream_billing_probe.go` | 纳入 TypeSafe API Key 探测身份，官方域名排除逻辑增加 typesafe.ai。 |
| `backend/internal/service/upstream_billing_probe_multiplatform_test.go` | 导入 TypeSafe 身份、域名和创建账号探测开关回归。 |
| `docs/UPSTREAM_SYNC.md` | 补充最终上游提交、累计范围和增量恢复点。 |
| `docs/upstream-sync-20261002-files.md` | 追加六个来源提交、五个路径和恢复说明。 |
| `progress.md` | 记录增量同步闭环与验证证据。 |

- 回滚整个本轮仍使用首次恢复分支 codex/pre-upstream-sync-20261002-v0.2.12、stash 87dc77358c35808872863e42f720311ab6999b56 和已校验归档 before.tgz。只回退增量可在新的独立目录建立 codex/pre-upstream-sync-20261002-followup 工作树，再应用 stash 8a6573d9dfb3119fefbe437b33b0526687d1a740；已有目录不要覆盖。具体恢复命令见 docs/upstream-sync-20261002-files.md。
- 本轮未构建/推送镜像、推送 GitHub、部署、修改线上数据或请求真实付费上游；不执行生产数据库回退。

## 2026-10-02 - Task: 构建并推送包含删除密钥计费修复的双架构镜像

### What was done

- 按用户要求从当前完整本地工作区构建并推送 `iotwq/china-api:latest` 和固定标签 `iotwq/china-api:0.2.12-20261002-210029`，支持 `linux/amd64`、`linux/arm64`；包含近期主项目同步及删除 API Key 后继续结算的修复，没有额外合并、改变版本或业务源码。
- 两标签共用索引摘要 `sha256:7691894c39a2c1a684caaa70d904e0df7e40feedf19397a7071645f7b68fad90`。amd64 子摘要为 `sha256:cb5d51f950e5e5e9ed1eaa2725ddd85ba933fcbd2a4ecf70b576d1a94b88beca`，arm64 子摘要为 `sha256:574bbadd2a8aba59d0dfa056e37ef150274648a92b033eb1cc08f215cf1916e9`。
- 版本保持 `0.2.12`，构建标识 `eda84f382314-dirty`，构建时间 `2026-10-02T13:00:29Z`；固定标签及本轮日志日期使用 Asia/Shanghai。本次未部署或执行生产迁移。

### Testing

- 发布前执行 `CI=true go test -mod=readonly -tags=integration ./internal/repository -run '^TestUsageBillingRepositoryApply_DeletedAPIKeyStillBillsBalance$' -count=1 -timeout=4m -v`，隔离 PostgreSQL/Redis 回归通过：删除 Key 后余额仍扣除 $1.25、用量/流水/幂等各一条，重复结算不重复扣费。
- Docker Buildx 使用既有 `codex-multiarch` builder，从当前工作区完成双架构构建和两标签推送，退出 0；镜像内翻译完整性 3 项、vue-tsc 类型检查、Vite 生产打包和两个架构 Go embed 编译通过。构建输出和 metadata 位于 `/tmp/sub2api-publish-20261002.YDsNp3/build-retry1.log`、`build-metadata.json`。
- `docker buildx imagetools inspect` 分别读取两个远端标签，索引摘要与 metadata 一致，确认 amd64、arm64 两种运行架构；其余 manifest 是构建证明。
- 按各自子摘要从 Docker Hub 拉取两个架构成功，分别用 `--network none --read-only --user 1000:1000 --cap-drop ALL --security-opt no-new-privileges` 运行 `/app/sub2api --version`，均退出 0 并报告一致版本、构建标识和时间。证据为同目录 `inspect-*.log`、`pull-*.log` 和 `version-*.log`。
- 初次 Docker Hub 查询和第一次构建的 Dockerfile 语法镜像 metadata 查询遇到 EOF，重试成功；原始失败构建日志保留为 `build.log`。保留前端既有 Browserslist/大 chunk 提示，未因提示改变依赖或配置。
- 发布记录追加后 `git diff --check`、`git diff --cached --check` 通过，Git 无未合并索引；未改动原有暂存分类。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加双架构发布标签、摘要、验证证据及部署/数据库兼容和回滚注意事项。
- `progress.md`：仅在末尾追加本轮构建推送和验证记录，保留历史日志及全部既有工作区修改。
- 回滚点是发布前的 `iotwq/china-api:0.2.9-20260930-004733`、索引摘要 `sha256:2c57dab4720feaed7b298ff711eef94f6761722a98822323db1f3f69c1dc85db`。恢复远端 latest 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:2c57dab4720feaed7b298ff711eef94f6761722a98822323db1f3f69c1dc85db`；部署回退需在 Compose 固定旧标签/摘要后执行 `docker compose pull sub2api` 和 `docker compose up -d --no-deps --force-recreate sub2api`。旧镜像不含此次漏扣修复，迁移后的数据库兼容需单独评估；本轮未执行回滚。
- 本轮只改变上述两份记录，不修改业务源码、Dockerfile、依赖、部署配置、生产数据库、余额或用量，不推送 GitHub、不调用真实付费上游。后续部署前备份数据库，并留意启动时自动应用尚未应用的迁移。

## 2026-10-02 - Task: 核对官方 0.2.13 并同步最新版本标识提交

### What was done

- 核实官方 v0.2.13 标签实际指向 3040209f205472038c1ba745a1bedd2edd9053b1，已由上一轮合并并包含在刚发布的镜像；显示 0.2.12 仅因上游版本文件后续才更新。
- 从 Wei-Shaw/sub2api 拉取 main，唯一新增提交 b8dece9000c68815a5b867ca5a1e6f236e173905 只更新 VERSION。合并提交为 e895e0cc587f2c6be794977e240e32ee6620fdd7，本地版本同步为 0.2.13。
- 独立临时 index 完成 merge，正式 index 仅同步 VERSION；全部原有定制、暂存和未暂存状态保留，没有新增业务功能或数据库迁移。

### Testing

- git ls-remote 确认官方标签及解引用提交；git merge-base --is-ancestor 确认 v0.2.13 标签代码早已包含，最新 b8dece900 也已包含。
- 对 4518 个原有工作区路径校验内容和权限，合并后只有 backend/cmd/server/VERSION 变化；git status --porcelain=v1 -z 与合并前逐字节一致，原暂存分类保持不变。
- backend/scripts/resolve-version.sh 输出 0.2.13；后端 go build -mod=readonly -tags embed 成功，临时二进制 --version 输出 Sub2API 0.2.13、commit e895e0cc587f-dirty。
- 本轮仅改变版本标识，沿用上一轮业务回归；不将未重新运行的全量测试记为本轮执行。

### Notes

- backend/cmd/server/VERSION：合并上游版本标识 0.2.12 → 0.2.13。
- docs/UPSTREAM_SYNC.md：追加官方标签与源码版本差异、合并结果、验证和恢复说明。
- progress.md：仅末尾追加本轮记录。
- 回滚点为 codex/pre-upstream-sync-20261002-v0.2.13（eda84f38231469c4968a8e574a1c26a221ce547d）。仅回退本轮版本标识可执行 git revert -m 1 e895e0cc587f2c6be794977e240e32ee6620fdd7 --no-commit；勿整体清理原有定制。备份证据目录 /tmp/sub2api-sync-20261002-v0213.LiV49S 保存原 index、工作区状态、暂存/未暂存补丁、版本文件和路径摘要。
- 未构建/推送 Docker 镜像、未推送 GitHub、未部署、未改生产数据。现有 latest 仍显示 0.2.12，但已经包含官方 v0.2.13 标签的业务代码。

## 2026-10-03 - Task: 支持 Midjourney V7 与 Seedream 5.0 Pro 图片模型转发

### What was done

- 按用户截图将 Midjourney-V7、Midjourney V7 和 seedream-5.0-pro 纳入 OpenAI API Key 兼容图片路由，支持中文/英文括号标签，如 Midjourney V7（满血）、seedream-5.0-pro（X）。标签仅在模型分类时忽略，转发保留完整映射名称。
- 请求解析与渠道映射后的能力筛选均限制为 OpenAI API Key 账号，复用既有 generations/edits、上传文件透传和图片响应转换；未扩展 Codex 原生模型识别、计费或限流策略。
- 核对截图另外三个 GPT 图片模型及带标签的映射，原逻辑已经支持；本次一并加入映射与编辑上传回归。

### Testing

- 修改前先运行新增回归，Midjourney/Seedream 请求稳定失败于 images endpoint requires an image model，GPT 模型映射可通过；日志 /tmp/sub2api-image-models-20261003.linQ3R/before.log。
- 修改后 service 定向测试通过，覆盖新模型及括号标签准入、无关模型拒绝、API Key 能力筛选与 OAuth/Setup Token 隔离、五组截图模型映射、JSON generations、multipart edits 文件与自定义字段保留、图片计数及 URL 转 Base64。
- 同轮 Gemini 兼容路由、原图片解析/API Key 转发、GPT 2.5、图片响应转换回归通过；handler 生图定向回归通过。命令分别为 go test -mod=readonly -tags=unit ./internal/service -run '^(TestCompatibleImages|TestOpenAIGatewayServiceParseOpenAIImagesRequest|TestOpenAIGatewayServiceForwardImages_APIKey|TestOpenAIImagesAPIKeyConversion|TestOpenAIImagesResponsesDriverAndImageModels|TestGPTImage25)' -count=1 -timeout=4m，以及 go test -mod=readonly -tags=unit ./internal/handler -run '^(TestOpenAIGatewayHandlerImages|TestOpenAIImages)' -count=1 -timeout=4m；日志同目录 service.log、handler.log。
- git diff --check 和 git diff --cached --check 通过。测试使用模拟上游，未验证该供应商真实接口或图片编辑能力，未发起付费生图。

### Notes

- backend/internal/service/openai_images.go：扩展仅供 API Key 的兼容模型分类和能力选择，精确支持本次两款模型及标签。
- backend/internal/service/openai_images_midjourney_seedream_test.go：新增准入、账号能力隔离及截图五组模型的转发回归。
- docs/OPENAI_MEDIA_COMPAT.md：补充模型映射、使用条件、接口和验证边界。
- progress.md：仅末尾追加本轮施工与验证记录。
- 回滚点为 /tmp/sub2api-image-models-20261003.linQ3R 中施工前的 openai_images.go、OPENAI_MEDIA_COMPAT.md、progress.md。若之后没有同文件新修改，可将前两份备份复制回对应文件并移除本轮新增测试文件，保留本条审计记录；如已有新修改，只按备份差异反向撤销本轮 hunk，不覆盖后续修改。本轮未执行回滚。
- 未构建或推送镜像、未部署，未修改线上账号配置、定价、余额或数据库；需发布后线上实例才获得新增路由。

## 2026-10-03 - Task: 补齐七个完整图片模型名与 Nano Banana 2 转发

### What was done

- 核对用户列出的七个名称，Midjourney V7（满血）、Midjourney-V7、seedream-5.0-pro（X）、seedream-5.0-pro 和 seedream-5.0-pro（满血）已有支持；本轮补充 nano-banana2-pro（满血）和 nano-banana2（满血），沿用括号标签识别。
- 七个完整名称均可直接请求，也可按账号映射转发；保留完整模型名称。新增 Nano Banana 名称使用既有 OpenAI API Key images 路径，不转换为 Gemini 原生模型或专用 nano-banana 接口。

### Testing

- 修改前新增用例确认仅 Nano Banana 2/Pro 准入与转发失败；其他五个完整名称的直接请求通过。证据 /tmp/sub2api-image-models-nano-20261003.JV1kJy/before.log。
- 修改后执行 go test -mod=readonly -tags=unit ./internal/service -run '^(TestCompatibleImages|TestOpenAIGatewayServiceParseOpenAIImagesRequest|TestOpenAIGatewayServiceForwardImages_APIKey|TestOpenAIImagesAPIKeyConversion|TestOpenAIImagesResponsesDriverAndImageModels)' -count=1 -timeout=4m，通过，日志同目录 after.log。
- 验证七个完整名称在 generations JSON 与 edits multipart 下的直接转发，以及基础名称映射为带标签名称；覆盖上传文件、自定义字段、响应图片计数、URL 转 Base64、API Key 能力限制及 GPT/Gemini 既有图片路由回归。使用模拟上游，未发起真实付费生图。
- git diff --check、git diff --cached --check 通过；仅扩展兼容模型准入，不改变计费、并发或鉴权实现。

### Notes

- backend/internal/service/openai_images.go：兼容模型分类增加 nano-banana2 和 nano-banana2-pro，复用原括号标签处理。
- backend/internal/service/openai_images_midjourney_seedream_test.go：加入全部七个完整名称的直接转发、更多账号映射及无关模型拒绝验证。
- docs/OPENAI_MEDIA_COMPAT.md：列出全部名称，说明直接调用、模型映射和标准 OpenAI 接口边界。
- progress.md：仅末尾追加本轮记录。
- 回滚点：/tmp/sub2api-image-models-nano-20261003.JV1kJy 保存本轮修改前上述源码、测试及媒体兼容文档。确认没有后续同文件修改时，可将目录内同名备份复制回原路径；已有后续修改则只反向撤销本轮差异，保留上一轮 Midjourney/Seedream 支持及历史日志。本轮未执行回滚。
- 尚未构建、推送镜像或部署；线上生效仍需发布。上游需提供对应 OpenAI 图片接口，并在本站配置模型白名单/映射及价格。

## 2026-10-03 - Task: 构建推送包含七个图片模型名称支持的双架构镜像

### What was done

- 从当前完整工作区构建并发布 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.13-20261003-032215`，包含最新版本标识、图片路由和全部现有本地定制，支持 linux/amd64、linux/arm64。
- 两标签共用索引摘要 `sha256:9791acb5fa4531543ae2c3670f0e6a6732786f648caa7c00f893652661296b79`。amd64 子摘要 `sha256:940376c79375a6234440af06d9125e5366d8886c14217c66cdf1997726f2849d`，arm64 子摘要 `sha256:9cedfebd7e0bdc472a9ba56821d298dd40a496dd7d82b35217e620ee3b9337f3`。
- 版本 0.2.13，构建标识 e895e0cc587f-dirty，构建时间 2026-10-02T19:22:15Z；本轮没有同步额外上游代码或部署线上服务。

### Testing

- Docker Buildx 双架构构建和两个标签推送退出 0；镜像内 locale 完整性 3 项、vue-tsc、Vite 生产构建及两个架构 Go embed 编译通过。保留既有 Browserslist/大 chunk 提示，未调整依赖。
- `docker buildx imagetools inspect` 分别核对远端 latest 与固定标签，索引摘要等于 build-metadata.json，包含 amd64/arm64 及各自构建证明。
- 普通 Docker Hub 连接及 `docker pull` 超时；本机已有代理可用，因此使用独立临时代理构建器发布，并通过只含 ARG/FROM 的临时 Dockerfile 按远端子摘要导入镜像。两个导入镜像摘要、架构与已发布子 manifest 一致，分别在 `--network none --read-only --user 1000:1000 --cap-drop ALL --security-opt no-new-privileges` 下执行 `/app/sub2api --version`，均退出 0 并返回预期版本、提交和时间。未将失败的普通拉取记为成功。
- 沿用紧邻前轮已通过的七模型 generations/edits、映射、图片转换及账号类型隔离回归；本次没有重新执行全量业务测试或真实上游生图。
- 证据目录 `/tmp/sub2api-publish-20261003.s5RZtc`：build.log、build-retry1.log 为直连失败；build-proxy.log、build-metadata.json 为成功发布；inspect-*.log、import-*.log、version-*.log 为远端和运行验证，pull-*.log 保留失败记录。
- 发布记录完成后执行 `git diff --check` 与 `git diff --cached --check`；未改变原暂存分类。

### Notes

- docs/UPSTREAM_SYNC.md：追加发布标签、摘要、验证边界和回滚说明。
- progress.md：仅末尾追加本轮发布记录。
- 本轮临时构建器 codex-publish-proxy-20261003 在验证后清理，原 codex-multiarch 构建器保留；未改全局代理、Dockerfile、依赖、部署配置、生产数据库、余额或用量。
- 回滚点为发布前已核实的 `iotwq/china-api:0.2.12-20261002-210029` / `sha256:7691894c39a2c1a684caaa70d904e0df7e40feedf19397a7071645f7b68fad90`。恢复远端 latest：`docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:7691894c39a2c1a684caaa70d904e0df7e40feedf19397a7071645f7b68fad90`；部署回退固定该摘要后执行 `docker compose pull sub2api` 和 `docker compose up -d --no-deps --force-recreate sub2api`。本轮未执行回滚；旧镜像不包含新模型路由。

## 2026-10-04 - Task: 支持 seedream-5.0-pro-x 图片模型转发

### What was done

- 将 `seedream-5.0-pro-x` 纳入 OpenAI API Key 兼容图片路由，支持直接请求以及通过账号模型映射转发到上游。
- 复用现有 `/v1/images/generations`、`/v1/images/edits`、图片响应转换和 API Key 能力限制；不改变 OAuth/Setup Token 图片路径、计费或其他模型路由。

### Testing

- 扩展图片模型准入与映射转发回归，覆盖 JSON generations、multipart edits、原始模型保留、映射后的上游模型名和 URL 转 Base64 行为。
- 定向 service 测试 `go test -mod=readonly -tags=unit ./internal/service -run '^(TestCompatibleImagesMidjourneySeedreamAdmission|TestCompatibleImagesScreenshotModelMapping)$' -count=1 -timeout=4m` 通过；`gofmt`、`git diff --check` 和 `git diff --cached --check` 均通过。未构建镜像或部署。

### Notes

- `backend/internal/service/openai_images.go`：增加 `seedream-5.0-pro-x` 兼容模型识别。
- `backend/internal/service/openai_images_midjourney_seedream_test.go`：增加直接准入和映射转发覆盖。
- `docs/OPENAI_MEDIA_COMPAT.md`：补充模型名称和映射示例。
- `docs/API_DOCS.md`、`frontend/src/views/user/ApiDocsView.vue`：新增三个模型的用户接口说明、参数限制和调用示例；`frontend/src/views/user/__tests__/ApiDocsView.spec.ts`：更新文档分组和模型断言。
- `progress.md`：追加本轮施工记录。
- 测试：`pnpm exec vitest run src/views/user/__tests__/ApiDocsView.spec.ts` 与 `pnpm exec vue-tsc --noEmit` 均通过；`git diff --check` 与 `git diff --cached --check` 通过。
- 回滚点：本轮仅涉及上述源码、测试和文档文件；可按本轮差异反向应用，保留此前未相关的本地修改。本轮未执行回滚。

## 2026-10-04 - Task: 构建推送当前双架构镜像

### What was done

- 从当前完整工作区构建并推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.13-20261004-190735`，支持 linux/amd64、linux/arm64；包含最新 Seedream X 路由和用户接口文档。
- 两标签共用索引摘要 `sha256:04607a3bb87893fc19756175aa07220081cc76ae291ace70fc182835da07c6e7`；amd64 子摘要 `sha256:9e4a2b65059390224a55737f74f18e66ed8b5761f7355113aeef683f33cee33a`，arm64 子摘要 `sha256:eaf34efc5818118433d0f596aa6329c0f8db64db5f7c2e33b6027d4f7c48bd91`。
- 版本 0.2.13，构建标识 e895e0cc587f-dirty，构建时间 2026-10-04T11:07:35Z；本轮未部署。

### Testing

- Docker Buildx 双架构构建和两个标签推送退出 0；前端国际化检查、vue-tsc、Vite 生产构建和两个架构 Go embed 编译通过。
- `docker buildx imagetools inspect` 核对 latest 与固定标签，索引摘要和 build metadata 一致，包含 amd64、arm64 及构建证明。
- 按远端子摘要导入本地镜像后，两个架构分别在 `--network none --read-only --user 1000:1000 --cap-drop ALL --security-opt no-new-privileges` 下执行 `/app/sub2api --version` 成功，均返回 0.2.13、正确提交和构建时间。
- 沿用前轮已通过的图片路由及接口文档定向测试；本轮没有重新执行全量业务测试或真实付费请求。`git diff --check` 与 `git diff --cached --check` 通过。
- 证据目录 `/tmp/sub2api-publish-20261004.fQl3Qf`：build.log、build-metadata.json、inspect-*.log、import-*.log、version-*.log。

### Notes

- `docs/UPSTREAM_SYNC.md`：追加本轮镜像标签、摘要、验证边界和回滚说明。
- `progress.md`：追加本轮发布记录。
- 临时构建器已清理；原 `codex-multiarch` 构建器保留。未修改 Dockerfile、依赖、部署配置、生产数据库、余额或用量。
- 回滚点为发布前 `iotwq/china-api:latest@sha256:9791acb5fa4531543ae2c3670f0e6a6732786f648caa7c00f893652661296b79`。恢复远端 latest：`docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:9791acb5fa4531543ae2c3670f0e6a6732786f648caa7c00f893652661296b79`；部署回退需固定该摘要后执行 `docker compose pull sub2api` 和 `docker compose up -d --no-deps --force-recreate sub2api`。本轮未执行回滚。

## 2026-10-04 - Task: 提交并推送当前工作区代码

### What was done

- 将当前项目工作区中的全部源码、测试、配置和文档变更整理为一次提交，目标远程为 `https://github.com/iotwq/sub2api.git` 的 `main` 分支。
- 保留根目录与项目无关的桌面应用生成 bundle 为未跟踪文件，避免把外部构建产物发布到项目仓库。

### Testing

- 提交前执行 `git diff --check` 和 `git diff --cached --check`，均通过。
- 推送完成后使用 `git ls-remote origin refs/heads/main` 核对远端分支提交与本地提交一致。

### Notes

- 本轮纳入当前工作区已有的全部项目变更，未同步 `upstream`、未部署、未重建镜像。
- 回滚点为提交前本地 `HEAD`；可使用 `git revert <本轮提交哈希>` 创建可审计的反向提交，或在部署前固定使用既有镜像摘要回退。

## 2026-10-07 - Task: 同步 Wei-Shaw/sub2api 最新提交

### What was done

- 将上游 `Wei-Shaw/sub2api` 的最新 8 个提交合并到当前 `main`，包含版本 0.2.14、EasyPay 回调防伪、初始化管理员凭据加固、前端依赖审计修复和远程 Codex 模型目录 API-key 发现。
- 合并 `frontend/package.json` 时保留本地 `three` 依赖，并采用上游 Vue 3.5.43 与 `source-map-js` 安全覆盖规则。
- 修复本地 Codex 配置生成器缺失的远程模型目录模式声明，并同步更新配置断言，保持现有不写入 `model_catalog_json` 的行为。

### Testing

- `go test -mod=readonly ./internal/setup ./internal/payment/provider ./internal/service -run 'Test(EasyPay|PaymentResume|Setup|Validate)' -count=1 -timeout=10m` 通过。
- `pnpm exec vitest run src/components/keys/__tests__/UseKeyModal.spec.ts src/views/admin/__tests__/SettingsView.spec.ts --reporter=dot`：84 项通过。
- `pnpm exec vue-tsc --noEmit`、`git diff --check` 和 `git diff --cached --check` 通过；保留测试中的既有 jsdom/router 警告。

### Notes

- 合并提交：`2d6043c25`；随后追加兼容性修复提交并推送到 `origin/main`。
- 本轮未部署、未重建镜像、未修改数据库或生产配置；根目录桌面应用 bundle 继续保持未跟踪。
- 回滚点为合并前提交 `54ae6320ccaed6480750cdb6db18941ad10cd9c6`；可使用 `git revert 2d6043c25` 及后续修复提交逐项回滚。

## 2026-10-07 - Task: 构建并推送 0.2.14 双架构镜像

### What was done

- 基于当前 `main`（`d9fe9fdbf300`）构建并推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.14-20261007-170451`。
- 两个标签共用索引摘要 `sha256:8a90b7a4707f24a782eff7ac9b743059d077a9853dd50370a2a4163131ffd4d4`，包含 `linux/amd64`（`sha256:a8c126fbc3c5319e6e0bec35378ea2af8d19cf968f3467844503c0e9e6af2f27`）和 `linux/arm64`（`sha256:174125d8128b1f1be6535b39ad79751fee5f4913d64d9e392afe5ba77a96e7df`）。

### Testing

- Docker Buildx 双架构构建和两个标签推送退出 0；镜像内国际化检查、`vue-tsc`、Vite 生产构建及两个架构 Go embed 编译通过。
- `docker buildx imagetools inspect` 核对两个远端标签，索引摘要和架构 manifest 一致。
- 按固定索引摘要拉取 amd64/arm64 镜像，并在无网络、只读文件系统、非 root、去除 capabilities 的条件下执行 `/app/sub2api --version`，两个架构均返回版本 0.2.14、提交 `d9fe9fdbf300` 和构建时间 `2026-10-07T09:04:51Z`。
- Docker Hub 首次拉取遇到认证/CDN EOF，重试成功；证据目录为 `/tmp/sub2api-publish-20261007`。

### Notes

- 本轮未修改业务源码、Dockerfile、部署配置、数据库或线上服务；仅追加 `progress.md` 发布记录。
- 回滚点为发布前 `iotwq/china-api:latest@sha256:04607a3bb87893fc19756175aa07220081cc76ae291ace70fc182835da07c6e7`。恢复 latest 可执行：`docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:04607a3bb87893fc19756175aa07220081cc76ae291ace70fc182835da07c6e7`；本轮未执行回滚或部署。
