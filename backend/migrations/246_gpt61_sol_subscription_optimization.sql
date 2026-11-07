-- GPT-6.1 Sol subscription optimization is opt-in per OpenAI group.
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS gpt61_sol_subscription_optimization_enabled BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN groups.gpt61_sol_subscription_optimization_enabled IS
    'Whether GPT-6.1 Sol subscription accounts use the long-context billing exemption and 500k generated-context default';
