CREATE TABLE IF NOT EXISTS pending_image_settlements (
    request_id TEXT NOT NULL,
    api_key_id BIGINT NOT NULL,
    request_fingerprint TEXT NOT NULL,
    snapshot JSONB NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    next_check_at TIMESTAMPTZ DEFAULT NOW(),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (request_id, api_key_id)
);

CREATE INDEX IF NOT EXISTS idx_pending_image_settlements_due
    ON pending_image_settlements (next_check_at) WHERE next_check_at IS NOT NULL;

COMMENT ON TABLE pending_image_settlements IS 'Completed image usage awaiting atomic settlement; NULL next_check_at requires accounting review';
