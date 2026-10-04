# Channel Monitor Retry Behavior

## Request budget

For OpenAI-compatible, Anthropic, and Grok checks, the monitor sends one request with the internal
`X-Sub2API-Channel-Monitor-Probe-Attempts: 5` header. A current sub2api gateway treats
that value as a budget of five distinct local accounts for the checked model.

- A pool-mode account may use its configured same-account retries first; that entire
  retry sequence consumes one account slot.
- After that account fails, the gateway may switch to another eligible account until
  five distinct accounts have been checked.
- Anthropic monitoring does not recycle an already-failed account when no other eligible
  account remains; fewer than five eligible accounts therefore produce fewer probes.
- The first successful account makes the gateway request successful. For a current
  sub2api gateway, the monitor calculates green/yellow status from that successful
  account attempt only; time spent waiting for earlier failed accounts is excluded.
- A successful attempt below the degraded threshold is green, and a successful attempt
  at or above the threshold is yellow. The result is red only when every applicable
  account attempt fails.
- Requests without the monitor header keep the normal account retry and switch limits.

An external compatible endpoint that does not implement this internal header receives
one monitor request and controls its own upstream retry behavior. Because it cannot
report the final internal attempt separately, the monitor keeps using the full request
latency for that external endpoint.

## Retry status configuration

The account-level `pool_mode_retry_status_codes` setting remains authoritative. The
default status list is unchanged. To retry the statuses covered by the channel monitor
pool scenario, explicitly configure:

```json
[401, 403, 429, 501, 502, 503]
```

Direct responses with a configured status enter same-account pool retry. If an upstream
pool wraps one of those statuses in an outer HTTP 400 response, sub2api recognizes only
this exact structured message prefix in `error.message` or top-level `message`:

```text
API returned <status>:
```

`<status>` must be one of `401`, `403`, `429`, `501`, `502`, or `503`. For ordinary user
requests, other HTTP 400 responses, different prefixes, embedded occurrences, missing
colons, and unlisted statuses remain non-retryable request errors.

Channel-monitor probes are the narrow exception: an OpenAI account that returns any HTTP
400 is excluded from that probe and the gateway continues with the next distinct eligible
account, within the same five-account budget. This accommodates upstream accounts that
block synthetic health checks behind a generic 400 response. It does not change ordinary
request behavior or make 400 a same-account pool retry status unless the account explicitly
configures 400 in `pool_mode_retry_status_codes`.

## Operational notes

The five-account cap limits health-check latency, token usage, and accidental retry
amplification. A failed monitor result is recorded only after the applicable account
budget is exhausted, no other eligible account is available, or the gateway receives a
non-retryable response. OpenAI monitor-only HTTP 400 responses are retryable across
distinct accounts. If fewer than five eligible accounts exist, only the available accounts
can be checked.
