# 2026-10-09 上游 0.2.15 合并清单

- 来源：Wei-Shaw/sub2api 的 upstream/main，固定提交 `3a6fd1c9db07203ca308aaba69e502bc1f35b307`。
- 范围：上一上游基线 `3f1a2ea0a` 至 `3a6fd1c9d`，146 个提交（86 个非合并提交），278 个上游变更文件。
- 同步前本地 HEAD：`7775571f4bf513081ff651843b4553112f545554`；恢复分支：`codex/pre-upstream-sync-20261009-v0.2.15`。
- 采用非快进合并；保留本站 Basispoints、OAuth/TLS/并发保护、312 观测、图片/视频路由与结算、群聊/私聊和余额查询。没有恢复 292/332 打票或 model_catalog_json。
- 无关桌面应用 bundle 维持未跟踪，不纳入提交。
- 新增迁移 242 移除两项数据库平台 CHECK，平台验证转由应用层、repository 和 Ent 共用清单执行；监控 provider 约束继续保留。部署前备份数据库；应用启动时执行迁移。代码回滚不会恢复数据库约束。
- 合并验证和部署边界见 progress.md 与 UPSTREAM_SYNC.md 的同日记录。
- 新增供应商 Claude 计费测试按本站已有缺价保护适配：未显式配置价格的 CN/OpenCode Claude 别名返回 ErrModelPricingUnavailable，不写零元成功账单；实际计费代码不因此变更。
- embed 静态资源回归使用本站实际保留的 dragon-logo.png，替代旧测试引用的已不存在的 logo.png；不改品牌素材或静态服务行为。

## 上游提交

```text
a4ea8059b fix(openai): recover rejected encrypted reasoning signatures
5ace0699b fix(pricing): preserve exponents when trimming scaled prices
e180a939a fix(users): show normalized errors when bulk updating limits
636faa695 fix(profile): ignore obsolete notification email code responses
7d0bf9c87 fix(usage): preserve zero user multipliers in cost details
3a1a37b7b fix(ops): retain byte units for sub-byte throughput
cee76030b fix(announcements): report failed read confirmations accurately
8bc375b59 fix(batch-image): allow access lookup retries after request failures
8a495b0a5 fix(usage): ignore obsolete cleanup task list responses
8167d85b3 fix(ui): close selectors when they become disabled
2e2b8f6d5 fix(auth): avoid starting OAuth code cooldowns after unmount
00a922c42 fix(backup): prevent polling from restarting after navigation
e0019da46 fix(ui): limit Escape to the most recently opened dialog
c64145312 fix(auth): share pending step-up verification prompts
75420462a fix(plugins): ignore obsolete configuration UI sessions
21d90fb5e fix(auth): keep cleared compliance state isolated from pending requests
3613174cf fix(grok): fail over silent-refusal empty response.completed on Grok Responses streams
27f4a1780 test(grok): cover empty response.completed failover for PlatformGrok
73381fc9c fix(openai-ws): bill each later turn with the current group pricing
71e97d135 fix(models): enforce account mappings for remote Codex discovery
f3d752693 fix(apicompat): flatten named chat tool_choice for Responses
9a8ab4f88 fix(apicompat): keep chat developer messages as developer role
7a767d5ce fix(apicompat): pair legacy chat function calls with their results
15f469c86 fix(gateway): ignore negative content block index in buffered Anthropic paths
775c1a3c1 fix(apicompat): preserve refusals when converting Responses to Chat
b6a157c19 fix(openai): account image input usage on WS passthrough
1888de135 fix(models): list composite exact route aliases
440784914 fix(channels): 追加模型时不再改写已留空的价格条目
c0e5a4670 fix(codex): never advertise null service_tiers in models manifest
dd7ee2170 style(codex): use tagged switch for capability conflict fallback
18380398d chore(codex): drop added comments
4b6ed15a8 fix(custom-pages): ignore obsolete Markdown responses
1c996fe30 fix(payment): stop late Stripe callbacks after leaving the page
abdfbb704 fix(web): bypass bare API aliases from embedded frontend fallback
5800abf1c fix(models): preserve composite catalog isolation after base update
852f0f824 fix(anthropic): preserve requested tool change beta headers
abbe42ac4 fix(profile): preserve email binding draft during user refresh
856d7ea90 fix: expose mapped antigravity test models
f385bba92 docs: keep connectivity test fix out of README
8f636513f fix(admin): keep scheduled test results scoped to the expanded plan
4551fb0d8 fix(monitor): ignore obsolete template manager list responses
b47204251 fix(ops): discard stale alert event pages after filter changes
81bd85ec5 fix(ops): require loaded settings before saving
b07b1fb5d fix(ops): validate alert rule durations as whole minutes
da4419e62 fix(anthropic): keep cache TTL order valid on the mimic path
679a7186f feat(accounts): compact account search and filters
9f84204c2 fix(streaming): preserve chat stream heartbeat comments
412718edf fix(apicompat): always include `signature` on Anthropic thinking blocks
0173861a5 fix(channel-monitor): build zhipu probe URL from the configured endpoint
5ef517d0f fix(accounts): attribute connectivity test failure logs
54810f8c8 fix(grok): include consent token and origin in SSO device approval
0dccd7b6f fix(opencode): correct Zen qwen3.8-max endpoint and reject unsupported models
aa3b04900 fix(opencode): cap Retry-After hint at 24h for usage refresh backoff
480f668ba fix(opencode): cap Retry-After hint for billing probe schedule
df89228c2 fix(opencode): report 403 as neutral forbidden in usage snapshot
d5f490e10 fix(usage): replace misleading long-context multiplier badge
7b1aa9314 fix(groups): preserve numeric values when editing RPM overrides
795a7e278 fix(payment): stop Airwallex checkout after leaving the page
f022a0457 fix(admin): use the saved passthrough rule enabled state
3f2f66f32 fix(profile): preserve username edits made during save
8811c125e fix(groups): ignore stale rate multiplier loads
61e1afdd4 perf(openai): lazily rebuild input when stripping namespaces
6f6a27f9c feat(usage): show per-request output TPS by default
38dd23f83 chore: remove PR screenshots
1817a8428 feat(ops): show per-request output TPS percentiles
18182f112 style(ops): align output TPS tooltip indentation
7cd3dc85c fix(channel-monitor): require request samples before health scoring
db29b3a77 Merge pull request #7811 from kingsleydon/fix/codex-service-tiers-null-fallback
6db4171cf Merge pull request #7779 from SPHINX998/fix/ws-turn-group-pricing
ef16f8a3d Merge pull request #7792 from SPHINX998/fix/ws-passthrough-image-usage
f087fdb20 Merge pull request #7785 from SPHINX998/fix/apicompat-chat-named-tool-choice
f9bd97cc1 Merge pull request #7786 from SPHINX998/fix/apicompat-chat-developer-role
8d5046680 Merge pull request #7787 from SPHINX998/fix/apicompat-chat-legacy-function-call
9d49d275f Merge pull request #7788 from SPHINX998/fix/apicompat-responses-refusal-to-chat
cc4a1e3f6 Merge pull request #7789 from SPHINX998/fix/anthropic-buffered-negative-block-index
e473fba99 Merge pull request #7869 from 3219378872/fix/thinking-signature-key
f19cd4fde Merge pull request #7844 from cyhhao/fix/anthropic-cache-ttl-order
129369204 Merge pull request #7827 from Achordchan/fix/anthropic-tool-change-betas
2afe7d457 Merge pull request #7910 from liuyh06/codex/fix-responses-namespace-noop
6cacccc76 Merge pull request #7775 from YahooYuan666/fix/grok-empty-completed-silent-refusal
90afaf120 Merge pull request #7854 from wucm667/fix/issue-7852-chat-stream-heartbeats
bbf1d3778 Merge pull request #7891 from aishangwuji/fix/opencode-go-usage-retry-after-and-403
3c274768c Merge pull request #7889 from aishangwuji/fix/opencode-zen-routing-and-unsupported-models
bb0d4a53c Merge pull request #7887 from whhphd/fix/grok-sso-consent
4b8c55770 Merge pull request #7877 from hezhaoqian1/fix/upstream-channel-monitor-zhipu-path
3a7fd4ae7 Merge pull request #7837 from yJader/fix/antigravity-account-test-models
458865ed1 Merge pull request #7878 from Achordchan/fix/account-test-log-context
bc63bd287 Merge pull request #7782 from lyy0709/codex/fix-remote-model-catalog
7ff5cff1b Merge pull request #7822 from neotan/fix/issue-7670-bare-api-alias-bypass
4bcad0657 Merge pull request #7800 from duckducktw/fix/pricing-entry-autofill-append
7fc5332b4 Merge pull request #7904 from ouchihao/codex/bug-group-rate-load-race
b126b9452 Merge pull request #7903 from ouchihao/codex/bug-profile-username-draft
cdbf1d6f4 Merge pull request #7902 from ouchihao/codex/bug-passthrough-toggle-state
c2f0edc13 Merge pull request #7901 from ouchihao/codex/bug-airwallex-navigation
1865f2846 Merge pull request #7900 from ouchihao/codex/bug-rpm-edit-number
d1202bea5 Merge pull request #7843 from ouchihao/codex/bug-alert-rule-integer-duration
486b226b6 Merge pull request #7842 from ouchihao/codex/bug-ops-settings-load-guard
02aad478f Merge pull request #7841 from ouchihao/codex/bug-alert-events-pagination-race
2c6bd63de Merge pull request #7840 from ouchihao/codex/bug-monitor-template-list-race
f3bcad3c3 Merge pull request #7839 from ouchihao/codex/bug-scheduled-results-race
4747827e3 Merge pull request #7819 from ouchihao/codex/bug-stripe-payment-unmount
b94974467 Merge pull request #7818 from ouchihao/codex/bug-custom-page-request-race
48c647b5d Merge pull request #7833 from ShanYi-Hui/fix/preserve-email-binding-draft
ba3af6125 Merge pull request #7897 from wucm667/fix/issue-7894-long-context-marker
0ac774e37 Merge pull request #7853 from wucm667/feat/issue-7847-compact-account-filters
8e2c4fad3 Merge pull request #7793 from wucm667/fix/issue-7784-composite-route-model-list
745ff1d3c Merge pull request #7911 from lyy0709/codex/usage-output-tps
2eb7ccb6d Merge pull request #7703 from wucm667/fix/issue-7700-thinking-signature
9e46f2d77 Merge pull request #7732 from ouchihao/codex/bug-step-up-concurrent-prompts
068239305 Merge pull request #7765 from ouchihao/codex/bug-compliance-reset-race
2c13fb595 Merge pull request #7764 from ouchihao/codex/bug-plugin-config-session-race
3dc2a9c37 Merge pull request #7731 from ouchihao/codex/bug-dialog-escape-stack
0b51cbe15 Merge pull request #7729 from ouchihao/codex/bug-backup-polling-unmount
3149d6285 Merge pull request #7724 from ouchihao/codex/bug-cleanup-task-race
6cfa0f4ef Merge pull request #7723 from ouchihao/codex/bug-batch-image-access-retry
9663b5ccf Merge pull request #7722 from ouchihao/codex/bug-announcement-read-failure
b0c2c4d62 Merge pull request #7720 from ouchihao/codex/bug-usage-zero-rate
a61833898 Merge pull request #7719 from ouchihao/codex/bug-notify-email-timers
d068b866f Merge pull request #7718 from ouchihao/codex/bug-bulk-user-error-message
5a6e826f8 Merge pull request #7716 from ouchihao/codex/bug-scaled-price-exponent
80dd7288a Merge pull request #7728 from ouchihao/codex/bug-oauth-code-unmount
3a012ac0c Merge pull request #7725 from ouchihao/codex/bug-select-disabled-open
3e6964339 Merge pull request #7721 from ouchihao/codex/bug-fractional-byte-rate
5fc0e486c Merge pull request #7916 from wucm667/fix/issue-7914-monitor-request-sample
c063a4016 refactor(provider): 以 provider profile 收敛多协议供应商默认端点表
3f8eea080 refactor(provider): profile 端点改为按协议登记，Responses 路径约定并入 profile
8ac2936a3 refactor(gateway): 三个入站入口的上游协议分流收敛为 resolveUpstreamProtocol
105c955c8 fix(gateway): /v1/responses 转 Anthropic 时统一以去空白的请求模型兜底映射
3b89cc08d refactor(platform): 后端平台列表改由平台清单派生
eefa7dd9d refactor(platform): 平台白名单由数据库约束改为应用层校验
c556aee8f refactor(platform): 多协议供应商的转发与探测判断改查 profile
7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动
e2939cd8f feat(platform): 接入 Command Code 平台
874871a25 feat(platform): 接入 Cline 平台
926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存
7ca08d521 fix(frontend): 按平台能力展示模型同步入口并同步 Zen 规则
387ee20ed test(platform): 补录制路由基准与平台约束迁移防护
c61f6ebaa fix(security): 升级 Go 1.27.2 与 x/net 修复安全扫描
5f93bb3b3 fix(ci): 升级 lint 支持 Go 1.27.2 并保留 HTTP2 兼容检查
d85cebecc fix(ci): 升级 schema 导入器支持 Go 1.27.2 导出格式
d0392edf0 Merge pull request #7618 from SavitarC/feat/provider-profile
dab3b87ea fix(openai): declare web_search when OAuth history replays web_search_call
64caa9af8 fix(openai): declare web_search via additional_tools for Responses Lite
85e311095 fix(lint): check type assertions in web_search history helper and test
f2669c8cf Merge pull request #7939 from IrisRainbowNeko/fix/issue-7927-web-search-history-tool
3a6fd1c9d chore: sync VERSION to 0.2.15 [skip ci]
```

## 文件与用途

| 文件 | 状态 | 本轮改动 |
| --- | --- | --- |
| `.github/workflows/backend-ci.yml` | M | c61f6ebaa fix(security): 升级 Go 1.27.2 与 x/net 修复安全扫描；5f93bb3b3 fix(ci): 升级 lint 支持 Go 1.27.2 并保留 HTTP2 兼容检查 |
| `.github/workflows/release.yml` | M | c61f6ebaa fix(security): 升级 Go 1.27.2 与 x/net 修复安全扫描 |
| `.github/workflows/security-scan.yml` | M | c61f6ebaa fix(security): 升级 Go 1.27.2 与 x/net 修复安全扫描 |
| `Dockerfile` | M | c61f6ebaa fix(security): 升级 Go 1.27.2 与 x/net 修复安全扫描 |
| `Makefile` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/.golangci.yml` | M | 5f93bb3b3 fix(ci): 升级 lint 支持 Go 1.27.2 并保留 HTTP2 兼容检查 |
| `backend/Dockerfile` | M | c61f6ebaa fix(security): 升级 Go 1.27.2 与 x/net 修复安全扫描 |
| `backend/cmd/server/VERSION` | M | 3a6fd1c9d chore: sync VERSION to 0.2.15 [skip ci] |
| `backend/ent/runtime/runtime.go` | M | eefa7dd9d refactor(platform): 平台白名单由数据库约束改为应用层校验 |
| `backend/ent/schema/composite_model_route.go` | M | eefa7dd9d refactor(platform): 平台白名单由数据库约束改为应用层校验 |
| `backend/ent/schema/user_platform_quota.go` | M | eefa7dd9d refactor(platform): 平台白名单由数据库约束改为应用层校验 |
| `backend/go.mod` | M | c61f6ebaa fix(security): 升级 Go 1.27.2 与 x/net 修复安全扫描；d85cebecc fix(ci): 升级 schema 导入器支持 Go 1.27.2 导出格式 |
| `backend/go.sum` | M | c61f6ebaa fix(security): 升级 Go 1.27.2 与 x/net 修复安全扫描；d85cebecc fix(ci): 升级 schema 导入器支持 Go 1.27.2 导出格式 |
| `backend/internal/domain/constants.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/domain/platforms.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/domain/platforms_test.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/handler/admin/account_handler.go` | M | 856d7ea90 fix: expose mapped antigravity test models；c556aee8f refactor(platform): 多协议供应商的转发与探测判断改查 profile |
| `backend/internal/handler/admin/account_handler_available_models_test.go` | M | 856d7ea90 fix: expose mapped antigravity test models |
| `backend/internal/handler/admin/channel_handler.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/handler/admin/group_handler.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/handler/admin/group_handler_platform_test.go` | M | 926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/handler/admin/platform_validation.go` | A | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生；926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/handler/gateway_handler.go` | M | 1888de135 fix(models): list composite exact route aliases；3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/handler/gateway_models_test.go` | M | 1888de135 fix(models): list composite exact route aliases |
| `backend/internal/handler/openai_codex_models_handler_test.go` | M | 71e97d135 fix(models): enforce account mappings for remote Codex discovery |
| `backend/internal/handler/openai_gateway_handler.go` | M | 73381fc9c fix(openai-ws): bill each later turn with the current group pricing；3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/handler/openai_gateway_handler_test.go` | M | 73381fc9c fix(openai-ws): bill each later turn with the current group pricing |
| `backend/internal/handler/openai_gateway_ws_group_pricing_test.go` | A | 73381fc9c fix(openai-ws): bill each later turn with the current group pricing |
| `backend/internal/model/error_passthrough_rule.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/model/error_passthrough_rule_test.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/pkg/antigravity/claude_types.go` | M | 412718edf fix(apicompat): always include `signature` on Anthropic thinking blocks |
| `backend/internal/pkg/antigravity/claude_types_test.go` | M | 412718edf fix(apicompat): always include `signature` on Anthropic thinking blocks |
| `backend/internal/pkg/antigravity/stream_transformer.go` | M | 412718edf fix(apicompat): always include `signature` on Anthropic thinking blocks |
| `backend/internal/pkg/apicompat/anthropic_responses_test.go` | M | 412718edf fix(apicompat): always include `signature` on Anthropic thinking blocks |
| `backend/internal/pkg/apicompat/chatcompletions_to_responses.go` | M | 9a8ab4f88 fix(apicompat): keep chat developer messages as developer role；7a767d5ce fix(apicompat): pair legacy chat function calls with their results |
| `backend/internal/pkg/apicompat/chatcompletions_to_responses_developer_role_test.go` | A | 9a8ab4f88 fix(apicompat): keep chat developer messages as developer role |
| `backend/internal/pkg/apicompat/chatcompletions_to_responses_legacy_function_call_test.go` | A | 7a767d5ce fix(apicompat): pair legacy chat function calls with their results |
| `backend/internal/pkg/apicompat/chatcompletions_to_responses_tool_choice_test.go` | A | f3d752693 fix(apicompat): flatten named chat tool_choice for Responses |
| `backend/internal/pkg/apicompat/responses_to_chatcompletions.go` | M | 775c1a3c1 fix(apicompat): preserve refusals when converting Responses to Chat |
| `backend/internal/pkg/apicompat/responses_to_chatcompletions_refusal_test.go` | A | 775c1a3c1 fix(apicompat): preserve refusals when converting Responses to Chat |
| `backend/internal/pkg/apicompat/thinking_signature_wire_test.go` | A | 412718edf fix(apicompat): always include `signature` on Anthropic thinking blocks |
| `backend/internal/pkg/apicompat/types.go` | M | 775c1a3c1 fix(apicompat): preserve refusals when converting Responses to Chat；412718edf fix(apicompat): always include `signature` on Anthropic thinking blocks |
| `backend/internal/pkg/claude/constants.go` | M | 852f0f824 fix(anthropic): preserve requested tool change beta headers |
| `backend/internal/pkg/xai/sso_device.go` | M | 54810f8c8 fix(grok): include consent token and origin in SSO device approval |
| `backend/internal/pkg/xai/sso_device_test.go` | M | 54810f8c8 fix(grok): include consent token and origin in SSO device approval |
| `backend/internal/repository/ops_repo_dashboard.go` | M | 1817a8428 feat(ops): show per-request output TPS percentiles |
| `backend/internal/repository/ops_repo_dashboard_integration_test.go` | A | 1817a8428 feat(ops): show per-request output TPS percentiles |
| `backend/internal/repository/ops_repo_dashboard_timeout_test.go` | M | 1817a8428 feat(ops): show per-request output TPS percentiles |
| `backend/internal/repository/platform_list_db_integration_test.go` | A | eefa7dd9d refactor(platform): 平台白名单由数据库约束改为应用层校验 |
| `backend/internal/repository/user_platform_quota_repo.go` | M | eefa7dd9d refactor(platform): 平台白名单由数据库约束改为应用层校验 |
| `backend/internal/server/api_contract_test.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/server/routes/gateway.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/service/account.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/account_header_override.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/service/account_scheduling_threshold_eval.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/service/account_service.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/service/account_test_logging.go` | A | 5ef517d0f fix(accounts): attribute connectivity test failure logs |
| `backend/internal/service/account_test_logging_test.go` | A | 5ef517d0f fix(accounts): attribute connectivity test failure logs |
| `backend/internal/service/account_test_service.go` | M | 合入供应商路由与后台测试日志分类；保留本地 HTTP Header 初始化。 |
| `backend/internal/service/account_test_service_cn_adaptive.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/admin_account.go` | M | 8ac2936a3 refactor(gateway): 三个入站入口的上游协议分流收敛为 resolveUpstreamProtocol |
| `backend/internal/service/admin_group.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/service/anthropic_apikey_auth.go` | M | c556aee8f refactor(platform): 多协议供应商的转发与探测判断改查 profile |
| `backend/internal/service/anthropic_buffered_negative_index_test.go` | A | 15f469c86 fix(gateway): ignore negative content block index in buffered Anthropic paths |
| `backend/internal/service/auth_service.go` | M | eefa7dd9d refactor(platform): 平台白名单由数据库约束改为应用层校验 |
| `backend/internal/service/channel_monitor_checker.go` | M | 合入智谱监控端点修复；保留探测重试、智力测试请求/答案及延迟隔离。 |
| `backend/internal/service/channel_monitor_endpoint_test.go` | M | 0173861a5 fix(channel-monitor): build zhipu probe URL from the configured endpoint |
| `backend/internal/service/channel_monitor_v2.go` | M | 7cd3dc85c fix(channel-monitor): require request samples before health scoring |
| `backend/internal/service/channel_monitor_v2_test.go` | M | 7cd3dc85c fix(channel-monitor): require request samples before health scoring |
| `backend/internal/service/channel_service.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/service/channel_service_test.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/cline.go` | A | 874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/cline_test.go` | A | 874871a25 feat(platform): 接入 Cline 平台；926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/service/cline_usage.go` | A | 874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/cn_provider_balance_check_service.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/cn_provider_balance_service.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/cn_provider_quota_service.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/cn_providers_test.go` | M | 3f8eea080 refactor(provider): profile 端点改为按协议登记，Responses 路径约定并入 profile |
| `backend/internal/service/command_code.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/service/command_code_test.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台；926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/service/command_code_usage.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/service/command_code_usage_test.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/composite_platform.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/service/composite_platform_test.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/composite_route_resolver.go` | M | 1888de135 fix(models): list composite exact route aliases |
| `backend/internal/service/composite_route_resolver_test.go` | M | 同步上游本轮变更 |
| `backend/internal/service/domain_constants.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/gateway_cache_ttl_order_test.go` | A | da4419e62 fix(anthropic): keep cache TTL order valid on the mimic path |
| `backend/internal/service/gateway_claude_oauth_body.go` | M | da4419e62 fix(anthropic): keep cache TTL order valid on the mimic path |
| `backend/internal/service/gateway_forward_as_chat_completions.go` | M | 15f469c86 fix(gateway): ignore negative content block index in buffered Anthropic paths；9f84204c2 fix(streaming): preserve chat stream heartbeat comments |
| `backend/internal/service/gateway_forward_as_chat_completions_ping_test.go` | A | 9f84204c2 fix(streaming): preserve chat stream heartbeat comments |
| `backend/internal/service/gateway_forward_as_responses.go` | M | 15f469c86 fix(gateway): ignore negative content block index in buffered Anthropic paths |
| `backend/internal/service/gateway_service.go` | M | 1888de135 fix(models): list composite exact route aliases |
| `backend/internal/service/gateway_tool_changes_beta_test.go` | A | 852f0f824 fix(anthropic): preserve requested tool change beta headers |
| `backend/internal/service/gateway_upstream_request.go` | M | 852f0f824 fix(anthropic): preserve requested tool change beta headers |
| `backend/internal/service/model_protocol_catalog.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台；926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/service/model_protocol_catalog_test.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台；926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/service/model_rate_limit.go` | M | 874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/openai_apikey_responses_probe.go` | M | c556aee8f refactor(platform): 多协议供应商的转发与探测判断改查 profile |
| `backend/internal/service/openai_codex_model_metadata.go` | M | dd7ee2170 style(codex): use tagged switch for capability conflict fallback；18380398d chore(codex): drop added comments |
| `backend/internal/service/openai_codex_model_metadata_test.go` | M | c0e5a4670 fix(codex): never advertise null service_tiers in models manifest；18380398d chore(codex): drop added comments |
| `backend/internal/service/openai_codex_models_pinned.go` | M | 71e97d135 fix(models): enforce account mappings for remote Codex discovery |
| `backend/internal/service/openai_codex_models_pinned_test.go` | M | 71e97d135 fix(models): enforce account mappings for remote Codex discovery |
| `backend/internal/service/openai_codex_transform.go` | M | dab3b87ea fix(openai): declare web_search when OAuth history replays web_search_call；64caa9af8 fix(openai): declare web_search via additional_tools for Responses Lite |
| `backend/internal/service/openai_gateway_chat_completions.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/service/openai_gateway_chat_completions_anthropic_native.go` | M | 15f469c86 fix(gateway): ignore negative content block index in buffered Anthropic paths |
| `backend/internal/service/openai_gateway_count_tokens.go` | M | c556aee8f refactor(platform): 多协议供应商的转发与探测判断改查 profile |
| `backend/internal/service/openai_gateway_forward.go` | M | 合入统一协议路由、web_search 历史声明和加密错误恢复；保留本站 BPS 分流及 enforce 禁止裁剪的规则。 |
| `backend/internal/service/openai_gateway_messages.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/service/openai_gateway_messages_anthropic_native.go` | M | 8ac2936a3 refactor(gateway): 三个入站入口的上游协议分流收敛为 resolveUpstreamProtocol；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/service/openai_gateway_passthrough.go` | M | 64caa9af8 fix(openai): declare web_search via additional_tools for Responses Lite |
| `backend/internal/service/openai_gateway_provider_claude_billing_test.go` | A | 同步聚合平台 Claude 定价覆盖；CN/OpenCode 缺价断言保留本站结算错误、不写零元成功账单的规则。 |
| `backend/internal/service/openai_gateway_request_body.go` | M | dab3b87ea fix(openai): declare web_search when OAuth history replays web_search_call；64caa9af8 fix(openai): declare web_search via additional_tools for Responses Lite |
| `backend/internal/service/openai_gateway_response_handling.go` | M | 3613174cf fix(grok): fail over silent-refusal empty response.completed on Grok Responses streams |
| `backend/internal/service/openai_gateway_responses_anthropic_native.go` | M | 15f469c86 fix(gateway): ignore negative content block index in buffered Anthropic paths |
| `backend/internal/service/openai_gateway_responses_empty_completed_test.go` | M | 27f4a1780 test(grok): cover empty response.completed failover for PlatformGrok |
| `backend/internal/service/openai_gateway_scheduling.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/service/openai_gateway_service_hotpath_test.go` | M | a4ea8059b fix(openai): recover rejected encrypted reasoning signatures |
| `backend/internal/service/openai_models_list.go` | M | 71e97d135 fix(models): enforce account mappings for remote Codex discovery |
| `backend/internal/service/openai_models_list_test.go` | M | 71e97d135 fix(models): enforce account mappings for remote Codex discovery |
| `backend/internal/service/openai_oauth_web_search_history.go` | A | 64caa9af8 fix(openai): declare web_search via additional_tools for Responses Lite；85e311095 fix(lint): check type assertions in web_search history helper and test |
| `backend/internal/service/openai_oauth_web_search_history_forward_test.go` | A | dab3b87ea fix(openai): declare web_search when OAuth history replays web_search_call；64caa9af8 fix(openai): declare web_search via additional_tools for Responses Lite |
| `backend/internal/service/openai_oauth_web_search_history_test.go` | A | 64caa9af8 fix(openai): declare web_search via additional_tools for Responses Lite；85e311095 fix(lint): check type assertions in web_search history helper and test |
| `backend/internal/service/openai_responses_namespace.go` | M | 61e1afdd4 perf(openai): lazily rebuild input when stripping namespaces |
| `backend/internal/service/openai_responses_namespace_benchmark_test.go` | A | 61e1afdd4 perf(openai): lazily rebuild input when stripping namespaces |
| `backend/internal/service/openai_responses_namespace_test.go` | M | 61e1afdd4 perf(openai): lazily rebuild input when stripping namespaces |
| `backend/internal/service/openai_ws_v2/passthrough_relay.go` | M | b6a157c19 fix(openai): account image input usage on WS passthrough |
| `backend/internal/service/openai_ws_v2/passthrough_relay_image_usage_test.go` | A | b6a157c19 fix(openai): account image input usage on WS passthrough |
| `backend/internal/service/openai_ws_v2_passthrough_adapter.go` | M | b6a157c19 fix(openai): account image input usage on WS passthrough |
| `backend/internal/service/openai_ws_v2_passthrough_image_usage_test.go` | A | b6a157c19 fix(openai): account image input usage on WS passthrough |
| `backend/internal/service/opencode_go.go` | M | 8ac2936a3 refactor(gateway): 三个入站入口的上游协议分流收敛为 resolveUpstreamProtocol；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/service/opencode_go_test.go` | M | 8ac2936a3 refactor(gateway): 三个入站入口的上游协议分流收敛为 resolveUpstreamProtocol；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/service/opencode_go_usage.go` | M | aa3b04900 fix(opencode): cap Retry-After hint at 24h for usage refresh backoff；df89228c2 fix(opencode): report 403 as neutral forbidden in usage snapshot |
| `backend/internal/service/opencode_go_usage_test.go` | M | aa3b04900 fix(opencode): cap Retry-After hint at 24h for usage refresh backoff；df89228c2 fix(opencode): report 403 as neutral forbidden in usage snapshot |
| `backend/internal/service/opencode_unsupported_models_routing_test.go` | A | 387ee20ed test(platform): 补录制路由基准与平台约束迁移防护 |
| `backend/internal/service/opencode_unsupported_models_test.go` | A | 0dccd7b6f fix(opencode): correct Zen qwen3.8-max endpoint and reject unsupported models；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/service/ops_dashboard_models.go` | M | 1817a8428 feat(ops): show per-request output TPS percentiles |
| `backend/internal/service/platform_catalog.go` | A | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/service/platform_catalog_test.go` | A | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `backend/internal/service/platform_list_test.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/provider_profile.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/provider_profile_test.go` | A | 3f8eea080 refactor(provider): profile 端点改为按协议登记，Responses 路径约定并入 profile；8ac2936a3 refactor(gateway): 三个入站入口的上游协议分流收敛为 resolveUpstreamProtocol |
| `backend/internal/service/ratelimit_cline.go` | A | 874871a25 feat(platform): 接入 Cline 平台；926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/service/ratelimit_cline_test.go` | A | 926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/service/ratelimit_cn_providers.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/ratelimit_command_code.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/service/ratelimit_service.go` | M | 874871a25 feat(platform): 接入 Cline 平台；926441461 fix(provider): 修复 Claude 计费与钱包冷却并隔离模型目录缓存 |
| `backend/internal/service/scheduler_snapshot_service.go` | M | 3b89cc08d refactor(platform): 后端平台列表改由平台清单派生 |
| `backend/internal/service/testdata/record_upstream_protocol_routing_main.py` | A | 387ee20ed test(platform): 补录制路由基准与平台约束迁移防护 |
| `backend/internal/service/testdata/upstream_protocol_routing_main_golden.json` | A | 387ee20ed test(platform): 补录制路由基准与平台约束迁移防护 |
| `backend/internal/service/upstream_billing_probe.go` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `backend/internal/service/upstream_billing_probe_test.go` | M | 480f668ba fix(opencode): cap Retry-After hint for billing probe schedule |
| `backend/internal/service/upstream_models.go` | M | c556aee8f refactor(platform): 多协议供应商的转发与探测判断改查 profile |
| `backend/internal/service/upstream_protocol_routing.go` | A | 8ac2936a3 refactor(gateway): 三个入站入口的上游协议分流收敛为 resolveUpstreamProtocol；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/service/upstream_protocol_routing_matrix_test.go` | A | e2939cd8f feat(platform): 接入 Command Code 平台；387ee20ed test(platform): 补录制路由基准与平台约束迁移防护 |
| `backend/internal/service/upstream_protocol_routing_test.go` | A | 105c955c8 fix(gateway): /v1/responses 转 Anthropic 时统一以去空白的请求模型兜底映射；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `backend/internal/web/embed_on.go` | M | abdfbb704 fix(web): bypass bare API aliases from embedded frontend fallback |
| `backend/internal/web/embed_test.go` | M | 同步裸 API 路由不进入前端回退的回归；静态资源断言改用本站现有 dragon-logo.png。 |
| `backend/migrations/242_drop_platform_check_constraints.sql` | A | eefa7dd9d refactor(platform): 平台白名单由数据库约束改为应用层校验 |
| `backend/migrations/platform_check_constraints_drop_migration_test.go` | A | eefa7dd9d refactor(platform): 平台白名单由数据库约束改为应用层校验；387ee20ed test(platform): 补录制路由基准与平台约束迁移防护 |
| `deploy/Dockerfile` | M | c61f6ebaa fix(security): 升级 Go 1.27.2 与 x/net 修复安全扫描 |
| `frontend/src/api/__tests__/settings.authSourceDefaults.spec.ts` | M | 更新 13 平台默认值断言，保留本站空额度校验。 |
| `frontend/src/api/admin/cnProviders.ts` | M | e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/api/admin/ops.ts` | M | 1817a8428 feat(ops): show per-request output TPS percentiles |
| `frontend/src/api/admin/settings.ts` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/api/admin/users.ts` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/components/account/AccountUsageCell.vue` | M | e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/components/account/CNProviderBalanceCell.vue` | M | e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/components/account/CNProviderQuotaCell.vue` | M | e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/components/account/CreateAccountModal.vue` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/components/account/EditAccountModal.vue` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/components/account/ModelWhitelistSelector.vue` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动；7ca08d521 fix(frontend): 按平台能力展示模型同步入口并同步 Zen 规则 |
| `frontend/src/components/account/OpenCodeGoProtocolRulesEditor.vue` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts` | M | 874871a25 feat(platform): 接入 Cline 平台；7ca08d521 fix(frontend): 按平台能力展示模型同步入口并同步 Zen 规则 |
| `frontend/src/components/account/__tests__/EditAccountModal.spec.ts` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/components/account/__tests__/ModelWhitelistSelector.spec.ts` | M | 7ca08d521 fix(frontend): 按平台能力展示模型同步入口并同步 Zen 规则 |
| `frontend/src/components/account/__tests__/OpenCodeGoProtocolRulesEditor.spec.ts` | A | e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/components/account/__tests__/credentialsBuilder.platformCatalog.spec.ts` | A | 874871a25 feat(platform): 接入 Cline 平台；7ca08d521 fix(frontend): 按平台能力展示模型同步入口并同步 Zen 规则 |
| `frontend/src/components/account/__tests__/credentialsBuilder.spec.ts` | M | e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/components/account/credentialsBuilder.ts` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/components/admin/ErrorPassthroughRulesModal.vue` | M | f022a0457 fix(admin): use the saved passthrough rule enabled state |
| `frontend/src/components/admin/__tests__/ErrorPassthroughRulesModal.toggle.spec.ts` | A | f022a0457 fix(admin): use the saved passthrough rule enabled state |
| `frontend/src/components/admin/account/AccountTableFilters.vue` | M | 679a7186f feat(accounts): compact account search and filters |
| `frontend/src/components/admin/account/ScheduledTestsPanel.vue` | M | 8f636513f fix(admin): keep scheduled test results scoped to the expanded plan |
| `frontend/src/components/admin/account/__tests__/AccountTableFilters.spec.ts` | A | 679a7186f feat(accounts): compact account search and filters |
| `frontend/src/components/admin/account/__tests__/ScheduledTestsPanel.results.spec.ts` | A | 8f636513f fix(admin): keep scheduled test results scoped to the expanded plan |
| `frontend/src/components/admin/channel/PricingEntryCard.vue` | M | 440784914 fix(channels): 追加模型时不再改写已留空的价格条目 |
| `frontend/src/components/admin/channel/__tests__/PricingEntryCard.modelDefaultPrice.spec.ts` | A | 440784914 fix(channels): 追加模型时不再改写已留空的价格条目 |
| `frontend/src/components/admin/group/GroupRPMOverridesModal.vue` | M | 7b1aa9314 fix(groups): preserve numeric values when editing RPM overrides；7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/components/admin/group/GroupRateMultipliersModal.vue` | M | 8811c125e fix(groups): ignore stale rate multiplier loads；7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/components/admin/group/__tests__/GroupRPMOverridesModal.edit.spec.ts` | A | 7b1aa9314 fix(groups): preserve numeric values when editing RPM overrides |
| `frontend/src/components/admin/group/__tests__/GroupRateMultipliersModal.requests.spec.ts` | A | 8811c125e fix(groups): ignore stale rate multiplier loads |
| `frontend/src/components/admin/monitor/MonitorTemplateManagerDialog.vue` | M | 4551fb0d8 fix(monitor): ignore obsolete template manager list responses |
| `frontend/src/components/admin/monitor/__tests__/MonitorTemplateManagerDialog.requests.spec.ts` | A | 4551fb0d8 fix(monitor): ignore obsolete template manager list responses |
| `frontend/src/components/admin/usage/UsageCleanupDialog.vue` | M | 8a495b0a5 fix(usage): ignore obsolete cleanup task list responses |
| `frontend/src/components/admin/usage/UsageTable.vue` | M | d5f490e10 fix(usage): replace misleading long-context multiplier badge；6f6a27f9c feat(usage): show per-request output TPS by default |
| `frontend/src/components/admin/usage/__tests__/UsageCleanupDialog.spec.ts` | A | 8a495b0a5 fix(usage): ignore obsolete cleanup task list responses |
| `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts` | M | d5f490e10 fix(usage): replace misleading long-context multiplier badge；6f6a27f9c feat(usage): show per-request output TPS by default |
| `frontend/src/components/admin/user/BulkEditUserModal.vue` | M | e180a939a fix(users): show normalized errors when bulk updating limits |
| `frontend/src/components/admin/user/UserPlatformQuotaModal.vue` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/components/admin/user/__tests__/BulkEditUserModal.spec.ts` | M | e180a939a fix(users): show normalized errors when bulk updating limits |
| `frontend/src/components/admin/user/__tests__/UserPlatformQuotaModal.spec.ts` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/components/auth/PendingOAuthCreateAccountForm.vue` | M | 2e2b8f6d5 fix(auth): avoid starting OAuth code cooldowns after unmount |
| `frontend/src/components/auth/__tests__/PendingOAuthCreateAccountForm.spec.ts` | M | 2e2b8f6d5 fix(auth): avoid starting OAuth code cooldowns after unmount |
| `frontend/src/components/common/AnnouncementBell.vue` | M | cee76030b fix(announcements): report failed read confirmations accurately |
| `frontend/src/components/common/BaseDialog.vue` | M | e0019da46 fix(ui): limit Escape to the most recently opened dialog |
| `frontend/src/components/common/PlatformIcon.vue` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/components/common/ProxySelector.vue` | M | 8167d85b3 fix(ui): close selectors when they become disabled |
| `frontend/src/components/common/Select.vue` | M | 8167d85b3 fix(ui): close selectors when they become disabled |
| `frontend/src/components/common/__tests__/AnnouncementBell.spec.ts` | A | cee76030b fix(announcements): report failed read confirmations accurately |
| `frontend/src/components/common/__tests__/BaseDialog.escape.spec.ts` | A | e0019da46 fix(ui): limit Escape to the most recently opened dialog |
| `frontend/src/components/common/__tests__/disabledSelectors.spec.ts` | A | 8167d85b3 fix(ui): close selectors when they become disabled |
| `frontend/src/components/keys/UseKeyModal.vue` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts` | M | 采用独立 features 断言，保留 API-key 远程发现及不生成 model_catalog_json 的校验。 |
| `frontend/src/components/user/UserPlatformQuotaCell.vue` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/components/user/profile/ProfileBalanceNotifyCard.vue` | M | 636faa695 fix(profile): ignore obsolete notification email code responses |
| `frontend/src/components/user/profile/ProfileEditForm.vue` | M | 3f2f66f32 fix(profile): preserve username edits made during save |
| `frontend/src/components/user/profile/ProfileIdentityBindingsSection.vue` | M | abbe42ac4 fix(profile): preserve email binding draft during user refresh |
| `frontend/src/components/user/profile/__tests__/ProfileBalanceNotifyCard.spec.ts` | M | 636faa695 fix(profile): ignore obsolete notification email code responses |
| `frontend/src/components/user/profile/__tests__/ProfileEditForm.draft.spec.ts` | A | 3f2f66f32 fix(profile): preserve username edits made during save |
| `frontend/src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts` | M | abbe42ac4 fix(profile): preserve email binding draft during user refresh |
| `frontend/src/composables/__tests__/useBatchImageAccess.retry.spec.ts` | A | 8bc375b59 fix(batch-image): allow access lookup retries after request failures |
| `frontend/src/composables/__tests__/useModelWhitelist.spec.ts` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/composables/__tests__/useStepUp.spec.ts` | M | c64145312 fix(auth): share pending step-up verification prompts |
| `frontend/src/composables/useBatchImageAccess.ts` | M | 8bc375b59 fix(batch-image): allow access lookup retries after request failures |
| `frontend/src/composables/useModelWhitelist.ts` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/composables/useStepUp.ts` | M | c64145312 fix(auth): share pending step-up verification prompts |
| `frontend/src/constants/__tests__/platforms.spec.ts` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/constants/platformCatalog.builtin.json` | A | 874871a25 feat(platform): 接入 Cline 平台；7ca08d521 fix(frontend): 按平台能力展示模型同步入口并同步 Zen 规则 |
| `frontend/src/constants/platformCatalog.ts` | A | e2939cd8f feat(platform): 接入 Command Code 平台；7ca08d521 fix(frontend): 按平台能力展示模型同步入口并同步 Zen 规则 |
| `frontend/src/constants/platforms.ts` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/i18n/locales/en/admin/accounts.ts` | M | df89228c2 fix(opencode): report 403 as neutral forbidden in usage snapshot；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/i18n/locales/en/admin/ops.ts` | M | 1817a8428 feat(ops): show per-request output TPS percentiles；18182f112 style(ops): align output TPS tooltip indentation |
| `frontend/src/i18n/locales/en/admin/resources.ts` | M | d5f490e10 fix(usage): replace misleading long-context multiplier badge |
| `frontend/src/i18n/locales/en/dashboard.ts` | M | 6f6a27f9c feat(usage): show per-request output TPS by default |
| `frontend/src/i18n/locales/zh/admin/accounts.ts` | M | df89228c2 fix(opencode): report 403 as neutral forbidden in usage snapshot；e2939cd8f feat(platform): 接入 Command Code 平台 |
| `frontend/src/i18n/locales/zh/admin/ops.ts` | M | 1817a8428 feat(ops): show per-request output TPS percentiles；18182f112 style(ops): align output TPS tooltip indentation |
| `frontend/src/i18n/locales/zh/admin/resources.ts` | M | d5f490e10 fix(usage): replace misleading long-context multiplier badge |
| `frontend/src/i18n/locales/zh/dashboard.ts` | M | 6f6a27f9c feat(usage): show per-request output TPS by default |
| `frontend/src/stores/__tests__/adminCompliance.reset.spec.ts` | A | 21d90fb5e fix(auth): keep cleared compliance state isolated from pending requests |
| `frontend/src/stores/adminCompliance.ts` | M | 21d90fb5e fix(auth): keep cleared compliance state isolated from pending requests |
| `frontend/src/stores/announcements.ts` | M | cee76030b fix(announcements): report failed read confirmations accurately |
| `frontend/src/types/index.ts` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/utils/__tests__/formatBytes.spec.ts` | A | 3a1a37b7b fix(ops): retain byte units for sub-byte throughput |
| `frontend/src/utils/__tests__/latencyHealth.spec.ts` | M | 6f6a27f9c feat(usage): show per-request output TPS by default |
| `frontend/src/utils/__tests__/pricing.formatScaled.spec.ts` | A | 5ace0699b fix(pricing): preserve exponents when trimming scaled prices |
| `frontend/src/utils/format.ts` | M | 3a1a37b7b fix(ops): retain byte units for sub-byte throughput |
| `frontend/src/utils/keyGroupProviders.ts` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/utils/latencyHealth.ts` | M | 6f6a27f9c feat(usage): show per-request output TPS by default |
| `frontend/src/utils/platformColors.ts` | M | e2939cd8f feat(platform): 接入 Command Code 平台；874871a25 feat(platform): 接入 Cline 平台 |
| `frontend/src/utils/pricing.ts` | M | 5ace0699b fix(pricing): preserve exponents when trimming scaled prices |
| `frontend/src/views/admin/AccountsView.vue` | M | 679a7186f feat(accounts): compact account search and filters |
| `frontend/src/views/admin/BackupView.vue` | M | 00a922c42 fix(backup): prevent polling from restarting after navigation |
| `frontend/src/views/admin/ChannelsView.vue` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动；7ca08d521 fix(frontend): 按平台能力展示模型同步入口并同步 Zen 规则 |
| `frontend/src/views/admin/GroupsView.vue` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/views/admin/PluginsView.vue` | M | 75420462a fix(plugins): ignore obsolete configuration UI sessions |
| `frontend/src/views/admin/SettingsView.vue` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/views/admin/__tests__/BackupView.spec.ts` | M | 00a922c42 fix(backup): prevent polling from restarting after navigation |
| `frontend/src/views/admin/__tests__/ChannelsView.modelSync.spec.ts` | A | 7ca08d521 fix(frontend): 按平台能力展示模型同步入口并同步 Zen 规则 |
| `frontend/src/views/admin/__tests__/PluginsView.spec.ts` | M | 75420462a fix(plugins): ignore obsolete configuration UI sessions |
| `frontend/src/views/admin/__tests__/channelPlatformOptions.spec.ts` | M | 7ffa38a8a feat(frontend): 平台列表与多协议账号表单改由平台清单驱动 |
| `frontend/src/views/admin/ops/components/OpsAlertEventsCard.vue` | M | b47204251 fix(ops): discard stale alert event pages after filter changes |
| `frontend/src/views/admin/ops/components/OpsAlertRulesCard.vue` | M | b07b1fb5d fix(ops): validate alert rule durations as whole minutes |
| `frontend/src/views/admin/ops/components/OpsDashboardHeader.vue` | M | 1817a8428 feat(ops): show per-request output TPS percentiles |
| `frontend/src/views/admin/ops/components/OpsDashboardSkeleton.vue` | M | 1817a8428 feat(ops): show per-request output TPS percentiles |
| `frontend/src/views/admin/ops/components/OpsSettingsDialog.vue` | M | 81bd85ec5 fix(ops): require loaded settings before saving |
| `frontend/src/views/admin/ops/components/__tests__/OpsAlertEventsCard.pagination.spec.ts` | A | b47204251 fix(ops): discard stale alert event pages after filter changes |
| `frontend/src/views/admin/ops/components/__tests__/OpsAlertRulesCard.duration.spec.ts` | A | b07b1fb5d fix(ops): validate alert rule durations as whole minutes |
| `frontend/src/views/admin/ops/components/__tests__/OpsDashboardHeader.spec.ts` | A | 1817a8428 feat(ops): show per-request output TPS percentiles |
| `frontend/src/views/admin/ops/components/__tests__/OpsSettingsDialog.loading.spec.ts` | A | 81bd85ec5 fix(ops): require loaded settings before saving |
| `frontend/src/views/user/AirwallexPaymentView.vue` | M | 795a7e278 fix(payment): stop Airwallex checkout after leaving the page |
| `frontend/src/views/user/CustomPageView.vue` | M | 4b6ed15a8 fix(custom-pages): ignore obsolete Markdown responses |
| `frontend/src/views/user/StripePaymentView.vue` | M | 1c996fe30 fix(payment): stop late Stripe callbacks after leaving the page |
| `frontend/src/views/user/__tests__/AirwallexPaymentView.lifecycle.spec.ts` | A | 795a7e278 fix(payment): stop Airwallex checkout after leaving the page |
| `frontend/src/views/user/__tests__/CustomPageView.race.spec.ts` | A | 4b6ed15a8 fix(custom-pages): ignore obsolete Markdown responses |
| `frontend/src/views/user/__tests__/StripePaymentView.spec.ts` | M | 1c996fe30 fix(payment): stop late Stripe callbacks after leaving the page |
| `frontend/src/views/user/__tests__/UsageView.spec.ts` | M | 6f6a27f9c feat(usage): show per-request output TPS by default |

## 恢复方法

在未部署情况下，可从同步前恢复分支创建独立工作区核对或重新构建：

```sh
git worktree add --detach /tmp/sub2api-pre-sync-20261009 codex/pre-upstream-sync-20261009-v0.2.15
```

如需在当前分支撤销此次合并，先确认无未提交工作，再以本轮合并提交为目标执行 `git revert -m 1 <merge-commit>`，生成反向提交；勿使用 hard reset。数据库迁移不由 Git 回滚，新增平台数据存在时不可直接恢复旧 CHECK；本轮不操作生产数据库或执行回滚。
