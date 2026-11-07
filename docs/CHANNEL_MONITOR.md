# Channel Monitor Retry Behavior

## Request budget

For OpenAI, Anthropic, Grok, Gemini, DeepSeek, Zhipu GLM, and Kimi checks, the monitor
sends one request with the internal
`X-Sub2API-Channel-Monitor-Probe-Attempts: 5` header. A current sub2api gateway treats
that value as a budget of five distinct local accounts for the checked model.

- A pool-mode account may use its configured same-account retries first; that entire
  retry sequence consumes one account slot.
- After that account fails, the gateway may switch to another eligible account until
  five distinct accounts have been checked.
- Anthropic and Gemini monitoring do not recycle an already-failed account when no other eligible
  account remains; fewer than five eligible accounts therefore produce fewer probes.
- The first successful account makes the gateway request successful. For a current
  sub2api gateway, the monitor calculates green/yellow status from that successful
  account attempt only; time spent waiting for earlier failed accounts is excluded.
- A successful attempt below 10 seconds is green (normal). At or above 10 seconds,
  it is yellow and labelled "延迟较高" ("High latency") in cards, details, history
  tooltips, and admin check results. The result is red only when every applicable
  account attempt fails. Pure-quota warnings retain their existing label because
  they do not measure probe latency.
- Requests without the monitor header keep the normal account retry and switch limits.
- The effective budget also respects a lower configured account-switch limit. Gemini
  explicit-cache requests stay on their bound account. These limits and existing
  non-retryable errors or request cancellation can end a probe before five accounts.

Gemini native `generateContent` monitoring uses the same budget and successful-attempt
latency rule. DeepSeek, Zhipu GLM, and Kimi reuse the OpenAI-compatible Chat Completions
probe path. Failed attempts never overwrite a later successful attempt's status; if no
attempt succeeds within the applicable budget, the monitor records a failure.

An external compatible endpoint that does not implement this internal header receives
one monitor request and controls its own upstream retry behavior. Because it cannot
report the final internal attempt separately, the monitor keeps using the full request
latency for that external endpoint.

The 10-second threshold applies to new checks after deployment. Stored historical
statuses are not recalculated; their yellow probe labels use the new wording.

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

## Channel status platform groups (V1)

The channel status card view groups visible monitors by platform in this order:
OpenAI, Anthropic, Grok, Gemini, DeepSeek, Zhipu GLM, and Kimi. Other supported
platforms with visible monitors follow in their original order of appearance.

Each platform heading shows its icon, name, and the number of visible monitor cards
under it. The count represents monitor cards, not the number of models checked by
those monitors. Platforms with no visible cards have no heading. Card order within
each platform follows the monitor list returned by the API.

Refreshing the monitor list updates platform groups and counts automatically. The
existing card details, availability windows, refresh behavior, and responsive grid
remain available. These grouping rules apply to the V1 card view; the V2 trend view
is unchanged.
# 用户渠道列表的多窗口汇总（2026-11-04）

`GET /api/v1/channel-monitors` 在现有 `availability_7d` 基础上增加主模型的 `availability_15d`、`availability_30d`。新增字段为数字或 `null`：对应窗口批量查询失败时返回 `null`（前端显示未知），没有该模型历史时沿用详情接口的 0；可用率公式不变。

V1 用户页直接使用列表的三个窗口值。切换 7/15/30 天不请求各分组详情，手动/自动刷新一次列表即更新全部窗口；打开卡片详情仍使用原详情接口。后端一次批量最新状态、三个窗口各一次批量统计、一次批量时间线，加上启用监控列表查询；调用次数不随监控分组数量增长，没有新增数据库结构或缓存失效延迟。前后端应同版本发布；新前端连接旧后端时 15/30 天会显示未知。

