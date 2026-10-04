ALTER TABLE users
    ADD COLUMN IF NOT EXISTS newapi_access_token_hash CHAR(64);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_newapi_access_token_hash
    ON users (newapi_access_token_hash)
    WHERE newapi_access_token_hash IS NOT NULL AND deleted_at IS NULL;
