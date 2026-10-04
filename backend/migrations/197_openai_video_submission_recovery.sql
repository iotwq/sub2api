ALTER TABLE openai_video_task_bindings
    ADD COLUMN IF NOT EXISTS upstream_task_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS billing_task_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS recovery_status VARCHAR(16) NOT NULL DEFAULT 'inactive',
    ADD COLUMN IF NOT EXISTS recovery_baseline JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS recovery_signature JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS recovery_billing JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS recovery_next_check_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS recovery_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS recovery_attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS recovery_last_error TEXT,
    ADD COLUMN IF NOT EXISTS recovery_checked_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_openai_video_task_bindings_recovery_pending
    ON openai_video_task_bindings (recovery_next_check_at, id)
    WHERE recovery_status IN ('submitting', 'pending', 'identified');

CREATE UNIQUE INDEX IF NOT EXISTS idx_openai_video_task_bindings_recovery_upstream
    ON openai_video_task_bindings (account_id, upstream_task_id)
    WHERE recovery_status IN ('identified', 'matched') AND upstream_task_id <> '';

COMMENT ON COLUMN openai_video_task_bindings.upstream_task_id IS '实际发送到原上游账号查询的任务 ID；恢复任务与客户端任务 ID 不同时使用';
COMMENT ON COLUMN openai_video_task_bindings.billing_task_id IS '创建请求级计费幂等任务 ID，正常响应和 504 恢复共用';
COMMENT ON COLUMN openai_video_task_bindings.recovery_status IS 'MiniMax-H3 创建恢复状态：inactive/submitting/pending/identified/matched/ambiguous/failed/cancelled';
COMMENT ON COLUMN openai_video_task_bindings.recovery_baseline IS '提交前原账号可见的任务 ID 基线';
COMMENT ON COLUMN openai_video_task_bindings.recovery_signature IS '用于唯一候选匹配的模型、分辨率、时长、比例和素材数量';
COMMENT ON COLUMN openai_video_task_bindings.recovery_billing IS '提交前冻结的计费、余额预占和使用记录快照';
