ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS video_input_duration_seconds DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS video_output_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS video_input_cost DECIMAL(20, 10) NOT NULL DEFAULT 0;

COMMENT ON COLUMN usage_logs.video_input_duration_seconds IS 'MiniMax-H3 服务端核验的参考视频计费总时长（秒）';
COMMENT ON COLUMN usage_logs.video_output_cost IS '生成视频部分的倍率前费用';
COMMENT ON COLUMN usage_logs.video_input_cost IS '参考视频部分的倍率前费用；不含图片和音频';

ALTER TABLE openai_video_task_bindings
    ADD COLUMN IF NOT EXISTS api_key_id BIGINT REFERENCES api_keys(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS api_key_quota_limited BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS api_key_rate_limited BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS compensation_status VARCHAR(16) NOT NULL DEFAULT 'inactive',
    ADD COLUMN IF NOT EXISTS next_check_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS check_attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_check_error TEXT,
    ADD COLUMN IF NOT EXISTS checked_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_openai_video_task_bindings_compensation_pending
    ON openai_video_task_bindings (next_check_at, id)
    WHERE compensation_status = 'pending';

COMMENT ON COLUMN openai_video_task_bindings.compensation_status IS '异步视频失败补偿状态：inactive/pending/completed/refunded';
COMMENT ON COLUMN openai_video_task_bindings.next_check_at IS '后台失败补偿下一次状态检查时间';
