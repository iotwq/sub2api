CREATE TABLE IF NOT EXISTS openai_video_task_bindings (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id TEXT NOT NULL,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT openai_video_task_bindings_task_id_check CHECK (length(task_id) > 0 AND length(task_id) <= 512),
    CONSTRAINT openai_video_task_bindings_unique_task UNIQUE (user_id, group_id, task_id)
);

CREATE INDEX IF NOT EXISTS idx_openai_video_task_bindings_expires_at
    ON openai_video_task_bindings (expires_at);

CREATE INDEX IF NOT EXISTS idx_openai_video_task_bindings_account
    ON openai_video_task_bindings (account_id);

CREATE INDEX IF NOT EXISTS idx_openai_video_task_bindings_user
    ON openai_video_task_bindings (user_id, created_at DESC);
