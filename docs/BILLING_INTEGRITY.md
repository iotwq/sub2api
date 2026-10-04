# Billing integrity

## Scope

These guarantees apply when the service runs in standard mode. Simple mode is intentionally non-billing and must not be used for a paid deployment.

## Request settlement

- Text, embedding, audio, image, video, Gemini-compatible, WebSocket turn, and Live usage all enter the shared billing pipeline.
- Requests require resolvable pricing before upstream dispatch. A missing requested, mapped, or concrete upstream price returns `400 invalid_request_error` with the model name and a prompt to contact the administrator; the request is not forwarded. Real billing infrastructure failures remain `503 billing_service_error`.
- Fixed-price media requests check that the current balance or subscription budget covers the estimated cost before dispatch.
- Exact post-response cost is authoritative. If an already completed request costs more than the remaining positive balance, the full cost is committed and the balance becomes negative. The negative balance blocks the next request.
- Billing effects, the visible `usage_logs` row, and the `billing_usage_entries` reconciliation row are committed in one PostgreSQL transaction. A usage-log insert failure rolls back the charge and its idempotency claim.
- A failed settlement is returned as an error and is not written as a misleading usage row with a positive original cost and `actual_cost = 0`.
- Gemini native `generateContent` and `streamGenerateContent` validate requested/channel/account-mapped pricing before dispatch. Model metadata and `countTokens` are not subject to this generation-price gate. Pricing calculation errors propagate from common usage billing instead of being converted into zero-cost success; explicitly configured free pricing/multipliers remain valid.

## Media behavior

- OpenAI image, Nano Banana, and video requests reserve known balance before dispatch and capture the reservation in the same transaction as the final usage row.
- Balance reservation, capture, and release amounts are quantized to the database's `NUMERIC(20,8)` scale before SQL execution. This prevents binary floating-point tails from making an exact frozen balance appear insufficient during settlement or release.
- Image billing uses the number of images actually returned. Gemini streaming image outputs are deduplicated before counting.
- Synchronous image settlement resolves the actual output size before calculating the charge. Capture, usage and ledger use that single final amount, including a difference above the reservation. The batch-image job settlement policy is unchanged.
- Completed image usage is saved to `pending_image_settlements` before the atomic charge. Failed settlements retain their hold and retry from the saved monetary snapshot; prices are not recalculated. The existing media reconciliation loop checks these records every 30 seconds, with leases and retry delays up to 10 minutes. A null `next_check_at` indicates an accounting conflict requiring review. Successful rows are removed after settlement; billing deduplication makes replay after a restart safe.
- Media capture and release serialize on the user balance. If the original hold was explicitly released, capture charges available balance once instead of consuming unrelated frozen funds. A missing/unexplained hold raises a reconciliation error; it does not trigger an invented debit or refund.
- Video billing uses the normalized model, resolution, duration, and operation. SD2.0 channel pricing resolves a separate per-second price for 480p, 720p, and 1080p as supported by each model; MiniMax H3, Grok video, and supported OpenAI-compatible video routes use the same settlement path.
- Async video preflight freezes the resolved billing model and cost. Balance capture, the visible usage row, and a later refund use that same snapshot even if channel mapping or pricing changes while the task is running.
- A create response that is already a terminal failure releases its balance reservation without creating a charge. Accepted tasks are settled once and then tracked for terminal failure.
- Live calls require explicit per-request channel pricing. A successful upstream creation is billed synchronously before the SDP is returned. If billing fails, the local call is closed and its concurrency lease is released.

## Failed video refunds

- A terminal failed or cancelled asynchronous video task creates one idempotent negative-cost refund usage row.
- The refund restores the original balance or subscription usage, API key quota and current rate-limit windows, and APIKey/Bedrock account quota windows that still contain the original usage.
- An API key set to `quota_exhausted` is restored to `active` when the refunded quota is below its limit.
- Accepted tasks are persisted and checked in the background. Explicit status requests can trigger the same reconciliation immediately; client polling is not required for the background path.
- Pending checks use leased `FOR UPDATE SKIP LOCKED` claims, so multiple service instances do not process the same task concurrently. Transient probe failures are retried with bounded backoff.
- Repeated status queries, process restarts, and concurrent workers do not refund the same task twice.
- MiniMax recoveries with an unexplained missing reservation or a conflicting billing fingerprint move to `billing_review` with no scheduled accounting retry. The unique upstream task owner remains reserved (migration `240_video_billing_review.sql`). Missing saved prices are also reviewed instead of recalculating a historical video at today's price.
- Once an upstream task has been uniquely identified, status and content requests use that task even while accounting retries or awaits review. They preserve the local task ID. Refund polling waits for the shared original billing ID to be settled, including when the actual-ID route and recovery route are separate rows. Accounting review does not label a successful upstream video as still generating.
- Video usage rows include output duration and resolution. MiniMax H3 rows also separate verified reference-video seconds and the output/input cost components; reversal rows negate the same components.

## Operational checks

- Alert on `record_usage_failed`, `billing_service_error`, atomic billing repository errors, and refund failures.
- Apply migration `239_pending_image_settlements.sql` with the normal startup migrations. Monitor `pending_image_settlements.last_error`, especially rows with `next_check_at IS NULL`. The snapshot excludes user/account/API key relation objects and their credentials. A database outage before the snapshot is saved still requires log-based reconciliation; retain the hold and request logs.
- Reconcile `usage_logs` with `billing_usage_entries` by `usage_log_id`; every standard-mode settled row must have one ledger entry.
- Monitor users with negative balances. They represent completed in-flight work whose exact cost exceeded the preflight snapshot, not free usage.
- Inspect video exceptions with `SELECT task_id, upstream_task_id, billing_task_id, recovery_last_error FROM openai_video_task_bindings WHERE recovery_status = 'billing_review';`. Check the original hold/capture/release and upstream result before resuming a reviewed task. Do not clear aggregate frozen balances or insert standalone refund rows.
- Database outages, process termination, and external infrastructure failures cannot be converted into an absolute zero-loss guarantee. Keep PostgreSQL highly available and run periodic usage-to-ledger reconciliation.
- If rolling back the application, retain both migrations and their data. The previous binary does not consume the new pending-image queue or reviewed video tasks; preserve those records for reconciliation before returning to an older build.

## Historical data

These changes do not automatically back-charge historical unsettled usage or rewrite existing balances. Historical recovery requires a separate reviewed reconciliation operation.
