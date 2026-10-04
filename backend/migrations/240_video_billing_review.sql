-- A task requiring accounting review must retain its unique upstream owner.
CREATE UNIQUE INDEX IF NOT EXISTS idx_openai_video_task_bindings_recovery_owned
    ON openai_video_task_bindings (account_id, upstream_task_id)
    WHERE recovery_status IN ('identified', 'matched', 'billing_review') AND upstream_task_id <> '';

DROP INDEX IF EXISTS idx_openai_video_task_bindings_recovery_upstream;

COMMENT ON COLUMN openai_video_task_bindings.recovery_status IS 'MiniMax-H3 recovery: inactive/submitting/pending/identified/matched/ambiguous/failed/cancelled/billing_review';
