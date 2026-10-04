-- Optional V1 candy check; historical probes remain untested.
ALTER TABLE channel_monitors
    ADD COLUMN IF NOT EXISTS intelligence_enabled BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE channel_monitor_histories
    ADD COLUMN IF NOT EXISTS intelligence JSONB;
