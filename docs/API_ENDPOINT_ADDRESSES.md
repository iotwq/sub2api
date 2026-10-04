# API endpoint addresses

## Settings

| Setting key | Admin label | User label | Purpose |
| --- | --- | --- | --- |
| `api_base_url` | API 端点地址 | 直连地址 | Existing primary API address |
| `optimized_api_base_url` | 国内优化地址 | 国内优化地址 | Optional user-facing optimized route |

Both values are stored in the existing key-value settings table. No database migration is required.

## Behavior boundary

- `api_base_url` keeps all existing behavior, including Use Key configuration, model chat, CC Switch import, and callback URL suggestions.
- `optimized_api_base_url` is exposed through public settings only for the API Keys page. It is displayed with copy and speed-test controls and is never assigned to the application's primary API base URL.
- The optimized address is omitted from the API Keys page when it is empty.
- Only the direct address carries the `默认` badge.
- Speed tests use `https://www.tcptest.cn/http/<encoded-address>` for each displayed address.
- Existing custom endpoints remain unchanged and continue to appear after the two fixed addresses.

## Public contract

The admin settings API and public settings API return `optimized_api_base_url` as a string. Server-injected public settings use the same field name so first-load and refreshed settings stay consistent.

## cc-switch NewAPI balance compatibility

Sub2API exposes a long-lived, balance-read-only access token for cc-switch. An authenticated user generates or resets it with the dashboard JWT:

```http
GET /api/user/token
Authorization: Bearer YOUR_DASHBOARD_JWT
```

The plaintext token is returned only by that response and is stored server-side only as a SHA-256 hash. Generating another token immediately invalidates the previous one. The token has no automatic expiry and is accepted only by the balance endpoint:

```http
GET /api/user/self
Authorization: Bearer YOUR_BALANCE_ACCESS_TOKEN
New-Api-User: YOUR_USER_ID
```

`GET /api/user/self` also continues to accept a valid dashboard JWT. `New-Api-User` is optional, but when supplied it must match the authenticated user ID. The response uses NewAPI's `success`/`data` envelope and returns `data.quota` as the current USD balance and `data.used_quota` as the accumulated spent estimate. Both fields use NewAPI's unit of 500,000 per USD, so cc-switch's NewAPI template converts them by dividing by 500,000.

In cc-switch, select the NewAPI balance query type, use the Sub2API site root (without `/v1`) as the base URL, enter the generated balance access token as `accessToken`, and enter the stable numeric user ID shown on the profile page. Normal `sk-...` model API keys are not accepted by the balance endpoint.
