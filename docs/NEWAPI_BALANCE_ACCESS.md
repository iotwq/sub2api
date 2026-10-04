# cc-switch NewAPI balance access

## User workflow

1. Sign in to Sub2API and open the profile page.
2. Copy the displayed numeric User ID.
3. In the CC Switch / NewAPI balance card, generate or reset the balance access token.
4. Copy the token immediately. It is shown only once; generating a new token invalidates the previous token.
5. In cc-switch, select the NewAPI balance query type and configure:
   - Base URL: the Sub2API site root, without `/v1`.
   - Access Token: the generated `sub_bal_...` token, without a `Bearer ` prefix.
   - User ID: the numeric ID displayed on the profile page.

cc-switch queries `GET /api/user/self` with `Authorization: Bearer <token>`. The optional `New-Api-User` header must match the authenticated user when present.

## Security boundary

- Only a SHA-256 token hash is stored in `users.newapi_access_token_hash`.
- The token has no automatic expiry, but it can be revoked immediately by generating another token.
- The token is accepted only for `GET /api/user/self`; it cannot call models, generate/reset tokens, or access other dashboard APIs.
- Dashboard JWT authentication remains supported for backward compatibility.
- Existing `sk-...` model API keys are intentionally not accepted for balance lookup because those keys may be shared with downstream applications.

## Response units

`data.quota` and `data.used_quota` use NewAPI's 500,000 units per USD convention. The current balance is `data.quota / 500000`. `used_quota` remains Sub2API's current accumulated-spend estimate, and `request_count` remains zero until a dedicated persisted request counter is introduced.
