# 2026-11-07 上游 0.2.20 合并清单

## 来源与范围

- 来源：Wei-Shaw/sub2api 的 main 分支。
- 前次同步基线：`a3298b46ebe239fb0720b068a1eb3383955bdac2`（0.2.19）。
- 本次目标：`6b4b37db43443b5e88e424fe3c69602cc3810dfb`（0.2.20）；目标提交时间为 2026-11-07 14:31:41（Asia/Shanghai），本日拉取并再次通过 `git ls-remote` 确认 main 仍为该提交。
- 上游增量 **107 个提交、458 个路径**；本地合并提交 `569c20342a7516e1703b93bfa2bd47c01227120f`。
- 用户已确认全部合入，包含管理 MCP、邮箱黑名单、鉴权/请求并发保护，以及 `249_mcp_oauth.sql` 迁移代码；本次只进行合并与隔离验证。
- 保留合并前 293 项跟踪文件状态、75 个未跟踪文件和两项 intent-to-add；原有 8 项删除保持删除。只有五个原有未跟踪语言模块/测试因兼容新 MCP 和配额文案而变化，其余原未跟踪文件内容未变。
- 本轮实际文件增量为上游 458 个路径、11 个额外兼容/测试路径，以及 4 个记录路径；不将原有定制整体提交。完整逐文件说明也追加于 `progress.md`。

## 行为与配置

- **管理 MCP**：新增 OAuth 授权、连接管理和个人/管理员工具，默认关闭。仅在显式设置 `MCP_PUBLIC_URL=https://你的域名/api/v1/mcp` 后开放接口和侧栏入口；需稳定的 HTTPS 地址，精确路径不带尾斜杠。默认授权个人只读；管理员写入还受 TOTP/二次验证和业务权限约束，详见 `mcp.md`。
- **注册邮箱黑名单**：支持精确邮箱黑名单及对命中既有用户的批量停用操作；本站没有配置名单、执行停用或修改生产用户。
- **管理 API 密钥**：补齐按操作/资源的权限、中英文展示和需要显式确认的全权兼容模式；不自动授予管理员权限。
- **图片接入**：GPT Image 2.6 两个型号补齐原生链路、参数、默认列表和价格兜底。账号支持按 1K/2K/4K 档位分流，`extra.openai_image_size_tiers` 缺省/空列表/全选表示不限制，不要求修改现有账号。
- **协议与计费**：合入 OpenAI 透传 instructions 回显、WS HTTP 桥接身份与逐轮首 token 时间、Grok Responses 输入兼容、Gemini 工具媒体与受限流读取、Antigravity 配额窗口/响应和 Claude Setup Token 周额度修复。
- **Terra**：OAuth/Setup Token 类型的 GPT-6.1 Terra 使用上游长上下文豁免；本站 `gpt61_sol_subscription_optimization_enabled` 继续只控制 Sol，不把 Sol 开关扩大成 Terra 开关。
- **请求资源保护**：新增请求体预算、读写超时和 SSE 队列边界；图片计费及其他后台工作使用受限 worker。新默认值见下表，已有明确配置优先；请求读取和下游写入超时不是模型整次生成时限。
- **性能和依赖**：合入 JSON 热路径减少解码/复制、隐私及 RPM 认证快照读取、用量 CSV 导出 429 等待重试；vue-i18n 升级至 11.4.16，依赖锁文件和 Go 安全依赖跟随上游。本地前端按需加载和近期 UI 调整继续保留。

| 配置键 | 默认值 | 作用 |
| --- | --- | --- |
| `gateway.request_body_memory_budget_bytes` | 1073741824（1 GiB） | 请求体在内存中的共享预算。 |
| `gateway.request_body_read_timeout_seconds` | 120 秒 | 限制客户端请求体读取等待。 |
| `gateway.downstream_write_timeout_seconds` | 120 秒 | 限制向慢客户端写入时的等待。 |
| `gateway.sse_queue_max_bytes` | 8388608（8 MiB） | 限制 SSE 队列积压字节数。 |
| `gateway.sse_event_max_lines` | 4096 | 限制单个 SSE 事件行数。 |
| `gateway.disable_openai_transport_error_cooldown` | false | 默认继续对 OpenAI 传输错误使用冷却，按需可显式关闭。 |

## 本站兼容处理

- 解决 15 处提交历史合并冲突和 11 处原工作树恢复冲突；保留 Basispoints/Excel OAuth、Codex 身份配置、图片/视频/音乐转发、冻结结算、模型广场、监控和近期前端改动。没有恢复已删除的 Codex 信号采集文件或 `svggrok-*` 别名。
- 图片结算接入上游异步 worker 时，先把冻结所有权交给后台任务；请求结束不提前解冻，已持久化的待结算失败继续保留冻结，待结算记录尚未持久化才使用独立 10 秒上下文释放。
- 尺寸错误诊断与本站 Native Images 可回退 Basic/OAuth 的选账号行为对齐，保留 API Key-only 限制和同账号重试。
- 认证快照使用 v28，同时保留 `RequirePrivacySet`、RPM 状态及本站 Sol 开关，拒绝此前版本的缓存快照。
- 上游请求读取错误处理同步应用到本站独有音频、视频及 Nano Banana 入口；Gemini 保留实际图片张数与部分结果用量，BPS 保留追踪及超时用量保护。
- 保留拆分后的语言模块，将 MCP 文案接到实际加载的用户入口，把 Antigravity 新文案补到共享管理员模块；不因上游单体语言入口覆盖本地性能调整。

## 数据库与发布边界

本次只新增上游迁移 `backend/migrations/249_mcp_oauth.sql`，创建授权、授权码和凭证摘要三张表：`mcp_oauth_grants`、`mcp_oauth_codes`、`mcp_oauth_credentials`。不存储原始授权码或令牌。

本次迁移仅在隔离测试数据库执行。将来部署新版本时应用启动会自动执行未应用迁移，即使 MCP 默认关闭也不能把“功能关闭”理解为“不迁移”；部署前备份数据库。代码回滚不删除已经创建的表或撤销已经发出的授权。只需要禁用 MCP 时清空 `MCP_PUBLIC_URL`，不要求删除表。

## 可执行恢复

- 合并前 HEAD：`57a3bafd515f75a7bece6797deb7c83b0f249a79`。
- 备份分支：`codex/pre-upstream-sync-20261107-145353`。
- 保留的 stash：`8099e3e253f7f3be8632ca6e6f2d5debf3dd1c81`，不删除。
- 证据目录：`/var/folders/p3/s4y4scsx14s5wr1r7693n_fw0000gn/T/sub2api-upstream-sync-20261107-nna6a1ek`；`before.tgz` 与 `before-manifest.json` 保存合并前 4893 个路径的原内容、模式和存在状态，包括被忽略的 docs 文件。
- 现有合并提交只记录合并阶段源码。恢复的本站定制和后续兼容修复仍在工作区，后续构建必须使用完整工作区，不能只拉取 `569c20342`。

以下命令在新临时目录重建合并前工作区，不覆盖当前 checkout：

```sh
restore_dir=$(mktemp -d /tmp/sub2api-pre-sync-20261107.XXXXXX)
git worktree add --detach "$restore_dir" codex/pre-upstream-sync-20261107-145353
git -C "$restore_dir" stash apply --index 8099e3e253f7f3be8632ca6e6f2d5debf3dd1c81
tar -xzf /var/folders/p3/s4y4scsx14s5wr1r7693n_fw0000gn/T/sub2api-upstream-sync-20261107-nna6a1ek/before.tgz -C "$restore_dir"
git -C "$restore_dir" add -N -f -- docs/IMAGE_RELEASE_20261029_005248.md docs/upstream-sync-20261028-files.md
```

证据目录中的 `task.patch` 仅记录本轮前后文件差异；`git apply --reverse --check /var/folders/p3/s4y4scsx14s5wr1r7693n_fw0000gn/T/sub2api-upstream-sync-20261107-nna6a1ek/task.patch` 可检验源码回退是否仍适用，但不会回退 Git 历史或数据库。优先使用上面的独立目录恢复方式，保留当前日志和后续工作。临时证据如需长期保留应另行归档。

## 验证

- 前端 `pnpm install --frozen-lockfile`、`pnpm typecheck`、`pnpm lint:check` 通过；全量 Vitest **416 个文件 / 3416 项测试通过**。补充 MCP 懒加载语言断言后对应 5 项测试复测通过。
- 使用显式 `vite.config.ts` 构建生产资源成功，输出到证据目录 `frontend-dist`，未覆盖原有后端 embed 目录；保留已有大 chunk、Browserslist 和测试环境 localStorage 提示。
- 最终后端全包 `GOMAXPROCS=4 go test -mod=readonly -p 4 -tags=unit ./... -count=1 -timeout=15m -json` 退出 0：61 个测试包、28073 个通过事件（含子测试）；52 个无测试文件的包和 41 项环境/可选测试按原条件跳过，不把跳过记为通过。管理密钥数据库用例随后显式提供隔离 DSN 验证。
- `go vet -mod=readonly ./...` 通过。handler/admin、handler、repository、service、middleware、routes 六包对 MCP、快照、图片 worker、请求资源保护、缓存和 BPS 等定向 race 通过（636 个通过事件，`-parallel 1`）；尺寸路由、冻结生命周期、Sol/Terra 与认证字段修复另有最终定向 race 复测。
- `httputil`、`apicompat`、`basispoints`、`openai_ws_v2` 四包完整 race 通过（1379 个通过事件）。
- 隔离 PostgreSQL 18.1 / Redis 8.4 的 MCP 授权码并发兑换、刷新重放、撤销、邮箱黑名单、原子计费及图片任务生命周期测试通过（115 个通过事件）；迁移幂等/并发、本站按张计费字段、Grok 目录和视频结算恢复兼容测试通过（20 个通过事件）。未连接生产数据库。
- 独立 PostgreSQL 的管理密钥生命周期、授权范围和陈旧写入保护通过（43 个通过事件）；进一步对 repository/service/admin handler 的管理密钥、凭据/模型映射写入、Codex 导入、代理选择、事务边界与旧密钥轮换执行带隔离 DSN 的 race 检查，3 包、274 个通过事件，0 跳过。测试容器已清理。
- 使用本轮生产资源、严格 nonce CSP 和本地模拟后端完成 10 项浏览器检查：MCP 用户/管理员、中英文、浅深色、375/768/1440 宽度、跳转本站接口文档、个人授权默认值，以及关闭 MCP 后隐藏入口；无未处理 JS 异常或整页横向溢出，没有提交真实授权。
- Wire 重新生成后复查 SHA-256 一致。完整后端源码与本轮前端产物在证据目录执行 embed 生产编译成功，`--version` 为 `Sub2API 0.2.20 (commit: 569c20342-dirty, built: unknown)`；构建未设置时间戳，也不是容器镜像。3794 个编译源码文件与最终工作区哈希一致。
- 保留失败证据：初轮后端暴露旧请求读取 helper 被删除、Gemini 调用参数不兼容；下一轮暴露尺寸路由夹具、OAuth 尺寸诊断和 Terra/Sol 开关合并错误。均按实际业务语义修复并完成最终全量与定向 race。浏览器首轮使用了错误的授权页地址，改为实际 `/oauth/mcp/authorize` 后复测，没有为测试改业务行为。
- 全工作区 `git diff --check` 仍报告合并前已存在的 `docs/CHANNEL_MONITOR.md:100` 文件尾空行；该文件哈希未变，不为同步任务清理无关内容。本轮增量补丁单独检查，无新增空白错误；无未合并索引或遗留冲突标记，进度日志原前缀保留。
- 没有请求真实付费上游，不能由模拟和本地测试推断线上账号授权、额度或模型生成效果。未推送 Git、未制作/推送镜像、未部署、未运行生产迁移。

主要证据：`backend-unit-final.jsonl`、`backend-vet.log`、`backend-race.jsonl`、`merge-final-race.log`、`backend-protocol-race.jsonl`、`backend-integration*.jsonl`、`backend-scoped-admin-postgres.jsonl`、`backend-admin-boundaries-race.jsonl`、`frontend-tests-initial.log`、`frontend-locale-merge.log`、`frontend-build.log`、`browser-results.json`、`production-build.log` 和 `wire-recheck.log`，均在上述证据目录。

## 上游逐文件来源

下表记录每个路径在本次上游区间内最近一次非合并提交；仅由合并提交引入的文件记录对应合并提交。同一文件可能还包含更早提交，本表不表示只合入最后一条。

| 变更 | 文件 | 本轮来源/用途 |
| --- | --- | --- |
| M | `.github/workflows/backend-ci.yml` | d851a3847 fix(frontend): upgrade vue-i18n to v11 and patch decode-uri-component |
| M | `.github/workflows/release.yml` | d851a3847 fix(frontend): upgrade vue-i18n to v11 and patch decode-uri-component |
| M | `.github/workflows/security-scan.yml` | d851a3847 fix(frontend): upgrade vue-i18n to v11 and patch decode-uri-component |
| M | `.gitignore` | 6c140b897 perf(gateway): reduce JSON copies and repeated request parsing |
| M | `README.md` | e6c19082f chore: update sponsors |
| M | `README_CN.md` | e6c19082f chore: update sponsors |
| M | `README_JA.md` | e6c19082f chore: update sponsors |
| A | `assets/partners/logos/aigocode-dark.png` | 876b96d20 chore: update sponsors |
| M | `assets/partners/logos/aigocode.png` | 876b96d20 chore: update sponsors |
| D | `assets/partners/logos/apimart.jpg` | e353c8322 chore: update sponsors |
| A | `assets/partners/logos/marucode.jpg` | 25608d216 chore: update sponsors |
| A | `assets/partners/logos/massivetokens.jpg` | e6c19082f chore: update sponsors |
| M | `backend/cmd/server/VERSION` | 6b4b37db4 chore: sync VERSION to 0.2.20 [skip ci] |
| M | `backend/cmd/server/wire.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/cmd/server/wire_gen.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/cmd/server/wire_gen_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/go.mod` | 6c140b897 perf(gateway): reduce JSON copies and repeated request parsing |
| M | `backend/go.sum` | 6c140b897 perf(gateway): reduce JSON copies and repeated request parsing |
| M | `backend/internal/config/config.go` | 489cfd5f4 feat(openai): allow disabling transport error cooldown |
| M | `backend/internal/config/config_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/config/mcp_config_test.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| A | `backend/internal/config/openai_transport_error_cooldown_test.go` | 489cfd5f4 feat(openai): allow disabling transport error cooldown |
| M | `backend/internal/handler/admin/account_handler.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/admin_service_stub_test.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `backend/internal/handler/admin/audit_log_handler.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/handler/admin/custom_admin_key_accounts.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/custom_admin_key_accounts_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/custom_admin_key_commit_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/custom_admin_key_group_models.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/handler/admin/custom_admin_key_idempotency_scope_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/handler/admin/custom_admin_key_openai.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/handler/admin/custom_admin_key_permissions_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/custom_admin_key_proxies.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/handler/admin/custom_admin_key_statistics.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/handler/admin/custom_admin_key_statistics_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/custom_admin_key_users.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/custom_admin_key_users_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/custom_admin_keys.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/handler/admin/email_blacklist_handler_test.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `backend/internal/handler/admin/idempotency_helper.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/openai_oauth_handler.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/openai_referral_handler.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/scoped_admin_content_response.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/scoped_admin_subscription_response.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/scoped_admin_subscription_response_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/handler/admin/scoped_admin_usage_response.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/handler/admin/setting_handler_email_blacklist.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `backend/internal/handler/admin/setting_handler_update.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/handler/admin/user_handler.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/handler/admin/user_handler_email_blacklist.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `backend/internal/handler/dto/settings.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/handler/gateway_handler.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/gateway_handler_chat_completions.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/gateway_handler_responses.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/gateway_systemone.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/gemini_cached_contents_handler.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/gemini_v1beta_handler.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/grok_audio.go` | 22693fb4f fix(grok): classify Grok Voice body read failures |
| A | `backend/internal/handler/grok_audio_body_test.go` | 22693fb4f fix(grok): classify Grok Voice body read failures |
| M | `backend/internal/handler/grok_media.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/handler.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| M | `backend/internal/handler/image_task_handler.go` | 0853d1603 fix(images): give async image tasks their own body memory reservation |
| M | `backend/internal/handler/image_task_handler_test.go` | 0853d1603 fix(images): give async image tasks their own body memory reservation |
| A | `backend/internal/handler/mcp_oauth_handler.go` | 575dd9bc2 fix(backend): make MCP client registration stateless |
| A | `backend/internal/handler/mcp_oauth_handler_test.go` | 575dd9bc2 fix(backend): make MCP client registration stateless |
| M | `backend/internal/handler/openai_alpha_search.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/openai_automation_bootstrap_test.go` | 4ca1a2097 perf(openai): skip full-body decodes for Codex bootstrap and compaction |
| A | `backend/internal/handler/openai_body_limit_passthrough_rule_test.go` | fa6f06510 fix(openai): let error passthrough rules override the built-in body-limit 413 |
| M | `backend/internal/handler/openai_chat_completions.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/openai_cyber_allowlist_test.go` | 7e3b6c72f fix(openai): 按已屏蔽前缀长度精确匹配 cyber transcript，修复长会话误拦 (#8519) |
| M | `backend/internal/handler/openai_embeddings.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/openai_gateway_count_tokens.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/openai_gateway_cyber_test.go` | 7e3b6c72f fix(openai): 按已屏蔽前缀长度精确匹配 cyber transcript，修复长会话误拦 (#8519) |
| M | `backend/internal/handler/openai_gateway_handler.go` | 4ca1a2097 perf(openai): skip full-body decodes for Codex bootstrap and compaction |
| M | `backend/internal/handler/openai_images.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/handler/openai_images_size_routing_test.go` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| A | `backend/internal/handler/openai_request_body_memory_test.go` | 6c140b897 perf(gateway): reduce JSON copies and repeated request parsing |
| M | `backend/internal/handler/ops_error_logger.go` | a699c2bae fix(openai): 透传放行空 instructions，响应回显按客户端原值还原 |
| M | `backend/internal/handler/ops_error_logger_test.go` | a699c2bae fix(openai): 透传放行空 instructions，响应回显按客户端原值还原 |
| M | `backend/internal/handler/request_body_limit.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/request_body_limit_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/request_body_read_log.go` | 3425ba366 fix(gateway): bound stalled request body reads and classify read failures |
| M | `backend/internal/handler/request_body_read_log_test.go` | 3425ba366 fix(gateway): bound stalled request body reads and classify read failures |
| M | `backend/internal/handler/setting_handler.go` | ee71c7f37 fix(backend): harden management MCP per review |
| A | `backend/internal/handler/usage_record_async_boundary_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/handler/usage_record_context_reuse_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/handler/wire.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| A | `backend/internal/managementmcp/auth.go` | ee71c7f37 fix(backend): harden management MCP per review |
| A | `backend/internal/managementmcp/catalog.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| A | `backend/internal/managementmcp/catalog_test.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| A | `backend/internal/managementmcp/server.go` | fb22c9212 fix(backend): bump MCP go-sdk to v1.4.1 for GO-2026-4770/4773 |
| A | `backend/internal/managementmcp/server_test.go` | fb22c9212 fix(backend): bump MCP go-sdk to v1.4.1 for GO-2026-4770/4773 |
| M | `backend/internal/middleware/rate_limiter.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/middleware/rate_limiter_test.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/pkg/adminaccess/catalog.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/pkg/adminaccess/catalog_permissions_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/pkg/adminaccess/catalog_statistics_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/pkg/adminaccess/full_access_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/pkg/adminaccess/policy.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/pkg/antigravity/attribution_test.go` | 01a1a78db fix(antigravity): neutralize stale system identities on forwarding |
| M | `backend/internal/pkg/antigravity/gemini_types.go` | aae31c743 fix(gemini): preserve tool media in native function responses |
| A | `backend/internal/pkg/antigravity/quota_summary.go` | bf4e16b90 feat(antigravity): show each quota group's real 5-hour and weekly window |
| A | `backend/internal/pkg/antigravity/quota_summary_test.go` | bf4e16b90 feat(antigravity): show each quota group's real 5-hour and weekly window |
| M | `backend/internal/pkg/antigravity/request_transformer.go` | aae31c743 fix(gemini): preserve tool media in native function responses |
| M | `backend/internal/pkg/antigravity/request_transformer_test.go` | aae31c743 fix(gemini): preserve tool media in native function responses |
| A | `backend/internal/pkg/antigravity/stream_response_shape_test.go` | c35418032 fix(antigravity): preserve direct Gemini streaming responses |
| M | `backend/internal/pkg/antigravity/stream_transformer.go` | c35418032 fix(antigravity): preserve direct Gemini streaming responses |
| A | `backend/internal/pkg/antigravity/tool_result_images_test.go` | aae31c743 fix(gemini): preserve tool media in native function responses |
| M | `backend/internal/pkg/apicompat/anthropic_to_responses.go` | e73f4c5e5 fix(openai): validate GPT-6.1 Terra effort and compat cache keys |
| M | `backend/internal/pkg/apicompat/anthropic_to_responses_response.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/pkg/apicompat/chatcompletions_anthropic_bridge.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/pkg/apicompat/chatcompletions_to_responses.go` | e73f4c5e5 fix(openai): validate GPT-6.1 Terra effort and compat cache keys |
| A | `backend/internal/pkg/apicompat/responses_message_serialization_test.go` | 2d6b64a18 fix(responses): preserve message content fields in completed responses |
| A | `backend/internal/pkg/apicompat/responses_reasoning.go` | 30ec2ecc9 fix(apicompat): preserve buffered reasoning content |
| A | `backend/internal/pkg/apicompat/responses_reasoning_content_test.go` | 30ec2ecc9 fix(apicompat): preserve buffered reasoning content |
| M | `backend/internal/pkg/apicompat/responses_stream_event_wire.go` | 2d6b64a18 fix(responses): preserve message content fields in completed responses |
| M | `backend/internal/pkg/apicompat/responses_to_anthropic.go` | 30ec2ecc9 fix(apicompat): preserve buffered reasoning content |
| M | `backend/internal/pkg/apicompat/responses_to_chatcompletions.go` | 30ec2ecc9 fix(apicompat): preserve buffered reasoning content |
| A | `backend/internal/pkg/apicompat/stream_accumulation_memory_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/pkg/apicompat/stream_string_buffer.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/pkg/apicompat/types.go` | 2d6b64a18 fix(responses): preserve message content fields in completed responses |
| A | `backend/internal/pkg/gemini/function_response.go` | aae31c743 fix(gemini): preserve tool media in native function responses |
| A | `backend/internal/pkg/gemini/function_response_test.go` | aae31c743 fix(gemini): preserve tool media in native function responses |
| M | `backend/internal/pkg/googleapi/status.go` | 3425ba366 fix(gateway): bound stalled request body reads and classify read failures |
| M | `backend/internal/pkg/googleapi/status_test.go` | 3425ba366 fix(gateway): bound stalled request body reads and classify read failures |
| M | `backend/internal/pkg/httputil/body.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/pkg/httputil/body_budget.go` | 0853d1603 fix(images): give async image tasks their own body memory reservation |
| A | `backend/internal/pkg/httputil/body_budget_test.go` | 0853d1603 fix(images): give async image tasks their own body memory reservation |
| A | `backend/internal/pkg/httputil/body_read_error.go` | 3425ba366 fix(gateway): bound stalled request body reads and classify read failures |
| A | `backend/internal/pkg/httputil/body_read_error_test.go` | 3425ba366 fix(gateway): bound stalled request body reads and classify read failures |
| A | `backend/internal/pkg/httputil/read_timeout.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/pkg/httputil/read_timeout_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/pkg/httputil/upstream_drain.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/pkg/httputil/upstream_drain_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/pkg/httputil/write_timeout.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/pkg/httputil/write_timeout_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/pkg/openai/codex_gpt61_terra.json` | bea67f412 fix(openai): use the official Codex descriptor for GPT-6.1 Terra |
| M | `backend/internal/pkg/openai/constants.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| M | `backend/internal/pkg/openai/constants_test.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| M | `backend/internal/repository/account_repo.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/repository/api_key_repo.go` | 9bb073d0d perf(openai): read the group privacy gate from the auth snapshot |
| M | `backend/internal/repository/api_key_repo_profit_projection_integration_test.go` | 1c83bb825 test(repository): assert the auth projection carries require_privacy_set |
| A | `backend/internal/repository/custom_admin_key_group_capacity.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/repository/custom_admin_key_group_models.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/repository/custom_admin_key_permissions_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/repository/custom_admin_key_scope.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/repository/custom_admin_key_scope_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/repository/custom_admin_keys.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/repository/custom_admin_keys_postgres_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/repository/custom_admin_keys_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/repository/fixtures_integration_test.go` | 1c83bb825 test(repository): assert the auth projection carries require_privacy_set |
| M | `backend/internal/repository/gateway_cache.go` | 7e3b6c72f fix(openai): 按已屏蔽前缀长度精确匹配 cyber transcript，修复长会话误拦 (#8519) |
| M | `backend/internal/repository/gateway_cache_cyber_test.go` | 7e3b6c72f fix(openai): 按已屏蔽前缀长度精确匹配 cyber transcript，修复长会话误拦 (#8519) |
| M | `backend/internal/repository/http_upstream.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/repository/http_upstream_drain_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/repository/mcp_oauth.go` | 575dd9bc2 fix(backend): make MCP client registration stateless |
| A | `backend/internal/repository/mcp_oauth_integration_test.go` | 575dd9bc2 fix(backend): make MCP client registration stateless |
| A | `backend/internal/repository/mcp_oauth_test.go` | 575dd9bc2 fix(backend): make MCP client registration stateless |
| M | `backend/internal/repository/scheduler_cache.go` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| M | `backend/internal/repository/scheduler_cache_unit_test.go` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| M | `backend/internal/repository/user_repo.go` | ee71c7f37 fix(backend): harden management MCP per review |
| A | `backend/internal/repository/user_repo_email_blacklist.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `backend/internal/repository/user_repo_email_blacklist_integration_test.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `backend/internal/repository/wire.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| M | `backend/internal/server/http.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/server/middleware/admin_auth.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/server/middleware/api_key_auth.go` | f10a45c03 fix(auth): stop reporting client disconnects and malformed keys as 500 |
| M | `backend/internal/server/middleware/api_key_auth_google.go` | f10a45c03 fix(auth): stop reporting client disconnects and malformed keys as 500 |
| M | `backend/internal/server/middleware/api_key_auth_test.go` | f10a45c03 fix(auth): stop reporting client disconnects and malformed keys as 500 |
| M | `backend/internal/server/middleware/audit_log.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/server/middleware/audit_log_test.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/server/middleware/cors.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| M | `backend/internal/server/middleware/cors_test.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| M | `backend/internal/server/middleware/custom_admin_key_rate_limit_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/server/middleware/custom_admin_keys_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/server/middleware/group_model_allowlist.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/server/middleware/panel_rate_limit.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| M | `backend/internal/server/middleware/request_body_limit.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/server/middleware/request_body_memory.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/server/middleware/request_body_memory_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/server/middleware/request_body_read_failure.go` | 3425ba366 fix(gateway): bound stalled request body reads and classify read failures |
| M | `backend/internal/server/middleware/step_up.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/server/middleware/step_up_test.go` | ee71c7f37 fix(backend): harden management MCP per review |
| A | `backend/internal/server/request_body_read_timeout_test.go` | fdfadb244 Merge upstream main; classify WithBodyReadTimeout failures |
| M | `backend/internal/server/router.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/server/routes/admin.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `backend/internal/server/routes/custom_admin_key_access_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/server/routes/custom_admin_key_full_access_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/server/routes/custom_admin_key_statistics_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/server/routes/gateway.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/server/routes/mcp.go` | fb22c9212 fix(backend): bump MCP go-sdk to v1.4.1 for GO-2026-4770/4773 |
| A | `backend/internal/server/routes/mcp_oauth_sdk_test.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| A | `backend/internal/server/routes/mcp_oauth_test.go` | 575dd9bc2 fix(backend): make MCP client registration stateless |
| A | `backend/internal/server/routes/mcp_rate_limit_test.go` | ee71c7f37 fix(backend): harden management MCP per review |
| A | `backend/internal/server/routes/mcp_test.go` | fb22c9212 fix(backend): bump MCP go-sdk to v1.4.1 for GO-2026-4770/4773 |
| M | `backend/internal/server/routes/user.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| M | `backend/internal/service/account_quota_service_account_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/account_test_service.go` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| M | `backend/internal/service/account_test_service_antigravity_test.go` | 0ee6b6c5d fix(antigravity): connection test sends the admin's prompt and lets the model answer |
| M | `backend/internal/service/account_test_service_openai_image_test.go` | 890120120 fix(openai): 修复 GPT Image 账号生图测试参数 |
| M | `backend/internal/service/account_usage_service.go` | a8d90a72e fix(anthropic): preserve setup-token weekly usage without 5h limits |
| A | `backend/internal/service/account_usage_service_setup_token_weekly_test.go` | a8d90a72e fix(anthropic): preserve setup-token weekly usage without 5h limits |
| M | `backend/internal/service/admin_service.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `backend/internal/service/admin_user_email_blacklist.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `backend/internal/service/admin_user_email_blacklist_test.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `backend/internal/service/anthropic_messages_endpoint_test.go` | 210608151 fix(anthropic): omit beta query for Cloudflare REST messages |
| A | `backend/internal/service/anthropic_messages_url.go` | 210608151 fix(anthropic): omit beta query for Cloudflare REST messages |
| A | `backend/internal/service/anthropic_messages_url_test.go` | 210608151 fix(anthropic): omit beta query for Cloudflare REST messages |
| M | `backend/internal/service/antigravity_gateway_compat.go` | aae31c743 fix(gemini): preserve tool media in native function responses |
| M | `backend/internal/service/antigravity_gateway_compat_stream.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/antigravity_gateway_compat_test.go` | 835803a70 fix(antigravity): preserve task-bearing identity preambles |
| M | `backend/internal/service/antigravity_gateway_gemini.go` | 0b535b5c0 fix(antigravity): Gemini 原生请求的 responseJsonSchema 改写为 responseSchema |
| M | `backend/internal/service/antigravity_gateway_retry.go` | 2ce1849bd fix(antigravity): scope INTERNAL 500 penalties to image models; make account tiers configurable |
| M | `backend/internal/service/antigravity_gateway_service.go` | 0ee6b6c5d fix(antigravity): connection test sends the admin's prompt and lets the model answer |
| M | `backend/internal/service/antigravity_gateway_streaming.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/antigravity_gateway_upstream.go` | d446c4a6b fix(gemini): bound upstream SSE lines and antigravity non-stream reads |
| A | `backend/internal/service/antigravity_gemini_response_schema_test.go` | 0b535b5c0 fix(antigravity): Gemini 原生请求的 responseJsonSchema 改写为 responseSchema |
| M | `backend/internal/service/antigravity_internal500_penalty.go` | 2ce1849bd fix(antigravity): scope INTERNAL 500 penalties to image models; make account tiers configurable |
| A | `backend/internal/service/antigravity_internal500_scope_test.go` | 2ce1849bd fix(antigravity): scope INTERNAL 500 penalties to image models; make account tiers configurable |
| M | `backend/internal/service/antigravity_quota_fetcher.go` | bf4e16b90 feat(antigravity): show each quota group's real 5-hour and weekly window |
| A | `backend/internal/service/antigravity_quota_groups_test.go` | bf4e16b90 feat(antigravity): show each quota group's real 5-hour and weekly window |
| A | `backend/internal/service/antigravity_tool_result_images_test.go` | aae31c743 fix(gemini): preserve tool media in native function responses |
| M | `backend/internal/service/api_key_auth_cache.go` | 9bb073d0d perf(openai): read the group privacy gate from the auth snapshot |
| M | `backend/internal/service/api_key_auth_cache_impl.go` | 9bb073d0d perf(openai): read the group privacy gate from the auth snapshot |
| M | `backend/internal/service/api_key_service.go` | f10a45c03 fix(auth): stop reporting client disconnects and malformed keys as 500 |
| M | `backend/internal/service/api_key_service_cache_test.go` | f10a45c03 fix(auth): stop reporting client disconnects and malformed keys as 500 |
| M | `backend/internal/service/audit_log.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/service/auth_email_oauth_auto.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `backend/internal/service/auth_email_oauth_blacklist_test.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `backend/internal/service/auth_registration_email_blacklist_test.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `backend/internal/service/auth_service.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `backend/internal/service/balance_notify_check_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/balance_notify_service.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/bedrock_stream.go` | a0b058a9e fix(bedrock): validate EventStream frame lengths before allocating |
| M | `backend/internal/service/bedrock_stream_test.go` | a0b058a9e fix(bedrock): validate EventStream frame lengths before allocating |
| M | `backend/internal/service/billing_cache_service.go` | 76221c9cb perf(billing): stop querying the RPM override on every request |
| M | `backend/internal/service/billing_cache_service_rpm_test.go` | 76221c9cb perf(billing): stop querying the RPM override on every request |
| A | `backend/internal/service/billing_notification_worker.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/service/billing_notification_worker_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/billing_service.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| M | `backend/internal/service/custom_admin_key_accounts.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/service/custom_admin_key_full_access_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/service/custom_admin_key_openai_options.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/service/custom_admin_key_permissions_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/service/custom_admin_key_statistics.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/service/custom_admin_key_statistics_test.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/service/custom_admin_keys.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/service/digest_session_store.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/digest_session_store_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/domain_constants.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `backend/internal/service/gateway_anthropic_passthrough.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/service/gateway_cache_control_paths_test.go` | 6c140b897 perf(gateway): reduce JSON copies and repeated request parsing |
| M | `backend/internal/service/gateway_claude_oauth_body.go` | 957a83f32 fix(gateway): settle patch budget per attempt and keep map fallback reachable |
| M | `backend/internal/service/gateway_forward_as_chat_completions.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/gateway_forward_as_responses.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/gateway_forward_compat_stream.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/gateway_request.go` | f654c23eb perf(gateway): decide thinking-block filtering without decoding the body |
| A | `backend/internal/service/gateway_request_reuse_test.go` | 957a83f32 fix(gateway): settle patch budget per attempt and keep map fallback reachable |
| A | `backend/internal/service/gateway_request_scalar_memory_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/gateway_service_benchmark_test.go` | 6c140b897 perf(gateway): reduce JSON copies and repeated request parsing |
| A | `backend/internal/service/gateway_thinking_filter_precheck_test.go` | f654c23eb perf(gateway): decide thinking-block filtering without decoding the body |
| M | `backend/internal/service/gateway_upstream_request.go` | 210608151 fix(anthropic): omit beta query for Cloudflare REST messages |
| M | `backend/internal/service/gateway_upstream_response.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/gateway_usage_billing.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/service/gateway_usage_quota_worker_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/gemini_chat_completions_compat_service.go` | 00660a227 perf(gemini): stop copying the emitted prefix on every stream delta |
| M | `backend/internal/service/gemini_compat_thinking_test.go` | aae31c743 fix(gemini): preserve tool media in native function responses |
| M | `backend/internal/service/gemini_messages_compat_service.go` | 00660a227 perf(gemini): stop copying the emitted prefix on every stream delta |
| M | `backend/internal/service/gemini_messages_compat_service_test.go` | aae31c743 fix(gemini): preserve tool media in native function responses |
| M | `backend/internal/service/gemini_native_response_signal_test.go` | d446c4a6b fix(gemini): bound upstream SSE lines and antigravity non-stream reads |
| A | `backend/internal/service/gemini_sse_line_limit_test.go` | d446c4a6b fix(gemini): bound upstream SSE lines and antigravity non-stream reads |
| A | `backend/internal/service/gemini_text_delta_test.go` | 00660a227 perf(gemini): stop copying the emitted prefix on every stream delta |
| M | `backend/internal/service/grok_quota_service.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/service/grok_sse_transport_config_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/group_capacity_service.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/service/mcp_oauth_service.go` | 575dd9bc2 fix(backend): make MCP client registration stateless |
| A | `backend/internal/service/mcp_oauth_service_test.go` | 575dd9bc2 fix(backend): make MCP client registration stateless |
| M | `backend/internal/service/openai_account_scheduler.go` | 9bb073d0d perf(openai): read the group privacy gate from the auth snapshot |
| M | `backend/internal/service/openai_apikey_responses_probe.go` | c476edbf4 fix(openai): skip responses probe when routing mode is forced |
| M | `backend/internal/service/openai_apikey_responses_probe_test.go` | c476edbf4 fix(openai): skip responses probe when routing mode is forced |
| M | `backend/internal/service/openai_codex_model_metadata.go` | bea67f412 fix(openai): use the official Codex descriptor for GPT-6.1 Terra |
| M | `backend/internal/service/openai_codex_models_service.go` | bea67f412 fix(openai): use the official Codex descriptor for GPT-6.1 Terra |
| M | `backend/internal/service/openai_compact_body_signal.go` | 4ca1a2097 perf(openai): skip full-body decodes for Codex bootstrap and compaction |
| M | `backend/internal/service/openai_compact_body_signal_test.go` | 4ca1a2097 perf(openai): skip full-body decodes for Codex bootstrap and compaction |
| M | `backend/internal/service/openai_compat_prompt_cache_key.go` | e73f4c5e5 fix(openai): validate GPT-6.1 Terra effort and compat cache keys |
| M | `backend/internal/service/openai_cyber_session_block.go` | 7e3b6c72f fix(openai): 按已屏蔽前缀长度精确匹配 cyber transcript，修复长会话误拦 (#8519) |
| M | `backend/internal/service/openai_cyber_session_block_test.go` | 7e3b6c72f fix(openai): 按已屏蔽前缀长度精确匹配 cyber transcript，修复长会话误拦 (#8519) |
| M | `backend/internal/service/openai_cyber_transcript.go` | 7e3b6c72f fix(openai): 按已屏蔽前缀长度精确匹配 cyber transcript，修复长会话误拦 (#8519) |
| M | `backend/internal/service/openai_first_output_timeout_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_anthropic_native_pump.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/service/openai_gateway_buffered_reasoning_content_test.go` | 30ec2ecc9 fix(apicompat): preserve buffered reasoning content |
| M | `backend/internal/service/openai_gateway_cc_pipeline.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_chat_completions.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_chat_completions_anthropic_native.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_chat_completions_raw.go` | e73f4c5e5 fix(openai): validate GPT-6.1 Terra effort and compat cache keys |
| M | `backend/internal/service/openai_gateway_chat_completions_test.go` | e73f4c5e5 fix(openai): validate GPT-6.1 Terra effort and compat cache keys |
| M | `backend/internal/service/openai_gateway_cn_fixes_test.go` | 17772947b fix(kimi): use short cooldown for concurrency-limit 403 |
| M | `backend/internal/service/openai_gateway_forward.go` | 957a83f32 fix(gateway): settle patch budget per attempt and keep map fallback reachable |
| M | `backend/internal/service/openai_gateway_grok.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_grok_model_input.go` | 522f1fb99 fix(grok): normalize delegated message input for Responses |
| M | `backend/internal/service/openai_gateway_grok_model_input_test.go` | 522f1fb99 fix(grok): normalize delegated message input for Responses |
| M | `backend/internal/service/openai_gateway_grok_sse_filter.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_grok_sse_filter_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_grok_test.go` | 522f1fb99 fix(grok): normalize delegated message input for Responses |
| M | `backend/internal/service/openai_gateway_grok_tool_protocol.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_grok_tool_protocol_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_messages.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_messages_anthropic_native.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_messages_chat_fallback.go` | e73f4c5e5 fix(openai): validate GPT-6.1 Terra effort and compat cache keys |
| M | `backend/internal/service/openai_gateway_passthrough.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_request_body.go` | 957a83f32 fix(gateway): settle patch budget per attempt and keep map fallback reachable |
| M | `backend/internal/service/openai_gateway_response_handling.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_responses_anthropic_native.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_responses_chat_fallback.go` | e73f4c5e5 fix(openai): validate GPT-6.1 Terra effort and compat cache keys |
| M | `backend/internal/service/openai_gateway_scheduling.go` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| M | `backend/internal/service/openai_gateway_service.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_gateway_service_hotpath_test.go` | 6c140b897 perf(gateway): reduce JSON copies and repeated request parsing |
| M | `backend/internal/service/openai_gateway_usage.go` | 64e6c260c fix(billing): exempt GPT-6.1 Terra subscriptions from long-context surcharge |
| M | `backend/internal/service/openai_gpt61_terra_test.go` | e73f4c5e5 fix(openai): validate GPT-6.1 Terra effort and compat cache keys |
| M | `backend/internal/service/openai_image_dimensions.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| A | `backend/internal/service/openai_image_size_routing.go` | 6b565e45a fix(openai): 生图尺寸溢出选号后按首轮前的绑定处理会话粘性 |
| A | `backend/internal/service/openai_image_size_routing_test.go` | 6b565e45a fix(openai): 生图尺寸溢出选号后按首轮前的绑定处理会话粘性 |
| M | `backend/internal/service/openai_images.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| A | `backend/internal/service/openai_images26_test.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| M | `backend/internal/service/openai_images_direct.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| M | `backend/internal/service/openai_images_direct_payload_test.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| M | `backend/internal/service/openai_images_direct_test.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| M | `backend/internal/service/openai_images_model_test.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| M | `backend/internal/service/openai_images_responses.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| A | `backend/internal/service/openai_images_stream_limit_error_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_messages_continuation.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_messages_digest_session.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_passthrough_normalization_test.go` | a699c2bae fix(openai): 透传放行空 instructions，响应回显按客户端原值还原 |
| A | `backend/internal/service/openai_privacy_gate_ctx_test.go` | 9bb073d0d perf(openai): read the group privacy gate from the auth snapshot |
| M | `backend/internal/service/openai_quota_service.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `backend/internal/service/openai_referral_service.go` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `backend/internal/service/openai_request_patches_benchmark_test.go` | 6c140b897 perf(gateway): reduce JSON copies and repeated request parsing |
| A | `backend/internal/service/openai_request_patches_test.go` | 957a83f32 fix(gateway): settle patch budget per attempt and keep map fallback reachable |
| A | `backend/internal/service/openai_responses_instructions_echo.go` | a699c2bae fix(openai): 透传放行空 instructions，响应回显按客户端原值还原 |
| A | `backend/internal/service/openai_responses_instructions_echo_test.go` | a699c2bae fix(openai): 透传放行空 instructions，响应回显按客户端原值还原 |
| A | `backend/internal/service/openai_responses_orphan_precheck.go` | 957a83f32 fix(gateway): settle patch budget per attempt and keep map fallback reachable |
| A | `backend/internal/service/openai_responses_orphan_precheck_benchmark_test.go` | 6c140b897 perf(gateway): reduce JSON copies and repeated request parsing |
| A | `backend/internal/service/openai_responses_orphan_precheck_test.go` | 6c140b897 perf(gateway): reduce JSON copies and repeated request parsing |
| M | `backend/internal/service/openai_sse_data.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_subscription_long_context_test.go` | 64e6c260c fix(billing): exempt GPT-6.1 Terra subscriptions from long-context surcharge |
| M | `backend/internal/service/openai_upstream_transport_error.go` | 489cfd5f4 feat(openai): allow disabling transport error cooldown |
| M | `backend/internal/service/openai_upstream_transport_error_handle_test.go` | 489cfd5f4 feat(openai): allow disabling transport error cooldown |
| M | `backend/internal/service/openai_ws_http_bridge.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/openai_ws_http_bridge_test.go` | 40d9c0964 fix(openai): apply Codex fingerprint convergence in the WebSocket HTTP bridge |
| M | `backend/internal/service/openai_ws_v2/passthrough_relay.go` | e16e69a7c fix(openai-ws): record first token time for every passthrough turn |
| M | `backend/internal/service/openai_ws_v2/passthrough_relay_test.go` | e16e69a7c fix(openai-ws): record first token time for every passthrough turn |
| M | `backend/internal/service/pricing_service.go` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| A | `backend/internal/service/ratelimit_cn_concurrency_403_test.go` | 17772947b fix(kimi): use short cooldown for concurrency-limit 403 |
| M | `backend/internal/service/ratelimit_cn_providers.go` | 17772947b fix(kimi): use short cooldown for concurrency-limit 403 |
| M | `backend/internal/service/ratelimit_service.go` | a8d90a72e fix(anthropic): preserve setup-token weekly usage without 5h limits |
| M | `backend/internal/service/ratelimit_service_cn_quota_403_test.go` | 17772947b fix(kimi): use short cooldown for concurrency-limit 403 |
| A | `backend/internal/service/ratelimit_weekly_usage_test.go` | a8d90a72e fix(anthropic): preserve setup-token weekly usage without 5h limits |
| A | `backend/internal/service/registration_email_blacklist.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `backend/internal/service/registration_email_blacklist_test.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `backend/internal/service/session_memory_cache.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/service/session_memory_cache_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/setting_public.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/service/setting_service.go` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `backend/internal/service/setting_service_public_test.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/service/settings_view.go` | ee71c7f37 fix(backend): harden management MCP per review |
| A | `backend/internal/service/sse_large_image_compat_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `backend/internal/service/sse_transport_limits.go` | d446c4a6b fix(gemini): bound upstream SSE lines and antigravity non-stream reads |
| A | `backend/internal/service/sse_transport_limits_test.go` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `backend/internal/service/user.go` | 76221c9cb perf(billing): stop querying the RPM override on every request |
| M | `backend/internal/service/wire.go` | ee71c7f37 fix(backend): harden management MCP per review |
| M | `backend/internal/web/embed_on.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| M | `backend/internal/web/embed_test.go` | f17918ea1 feat(backend): add OAuth-authenticated management MCP |
| A | `backend/migrations/249_mcp_oauth.sql` | 575dd9bc2 fix(backend): make MCP client registration stateless |
| M | `backend/resources/model-pricing/model_prices_and_context_window.json` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| M | `deploy/.env.example` | 489cfd5f4 feat(openai): allow disabling transport error cooldown |
| M | `deploy/EDGE_SECURITY.md` | 457602c8f fix(backend): bound gateway resources and background work |
| M | `deploy/config.example.yaml` | 489cfd5f4 feat(openai): allow disabling transport error cooldown |
| M | `deploy/docker-compose.dev.yml` | 489cfd5f4 feat(openai): allow disabling transport error cooldown |
| M | `deploy/docker-compose.local.yml` | 489cfd5f4 feat(openai): allow disabling transport error cooldown |
| M | `deploy/docker-compose.standalone.yml` | 489cfd5f4 feat(openai): allow disabling transport error cooldown |
| M | `deploy/docker-compose.yml` | 489cfd5f4 feat(openai): allow disabling transport error cooldown |
| M | `docs/admin-api-keys.md` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| A | `docs/agent-resource-safety.md` | 457602c8f fix(backend): bound gateway resources and background work |
| A | `docs/mcp.md` | 575dd9bc2 fix(backend): make MCP client registration stateless |
| A | `docs/moego-json-batching.md` | 957a83f32 fix(gateway): settle patch budget per attempt and keep map fallback reachable |
| M | `frontend/package.json` | ff0429651 补齐管理员密钥中英文权限展示并修复依赖安全扫描 |
| M | `frontend/pnpm-lock.yaml` | ff0429651 补齐管理员密钥中英文权限展示并修复依赖安全扫描 |
| A | `frontend/src/api/__tests__/admin.emailBlacklist.spec.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/api/__tests__/client.spec.ts` | e47abd8f9 fix(usage): 用户导出 CSV 每页取 1000 条，遇面板限流 429 按 Retry-After 等待后重试 |
| M | `frontend/src/api/admin/adminAccessKeys.ts` | 194d69e63 完善管理员密钥权限并新增需确认的全权兼容模式 |
| M | `frontend/src/api/admin/settings.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/api/admin/users.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/api/client.ts` | e47abd8f9 fix(usage): 用户导出 CSV 每页取 1000 条，遇面板限流 429 按 Retry-After 等待后重试 |
| A | `frontend/src/api/mcp.ts` | cf3300633 feat(frontend): add MCP OAuth consent and connections |
| M | `frontend/src/components/account/AccountUsageCell.vue` | 455751402 fix(antigravity): label Pro plans' separate Argon quota as its own row |
| M | `frontend/src/components/account/BulkEditAccountModal.vue` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| M | `frontend/src/components/account/CreateAccountModal.vue` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| M | `frontend/src/components/account/EditAccountModal.vue` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| A | `frontend/src/components/account/OpenAIImageSizeTiersField.vue` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| M | `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts` | 455751402 fix(antigravity): label Pro plans' separate Argon quota as its own row |
| M | `frontend/src/components/account/__tests__/BulkEditAccountModal.spec.ts` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| M | `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| M | `frontend/src/components/account/__tests__/EditAccountModal.spec.ts` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| A | `frontend/src/components/account/__tests__/OpenAIImageSizeTiersField.spec.ts` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| A | `frontend/src/components/account/openaiImageSizeTiers.ts` | e452e41e4 feat(openai): 生图账号按 1K/2K/4K 尺寸档位分流 |
| M | `frontend/src/components/admin/settings/AdminAccessKeysSettings.vue` | ff0429651 补齐管理员密钥中英文权限展示并修复依赖安全扫描 |
| A | `frontend/src/components/admin/settings/RegistrationEmailBlacklistEditor.vue` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `frontend/src/components/admin/settings/__tests__/AdminAccessKeysSettings.i18n.spec.ts` | ff0429651 补齐管理员密钥中英文权限展示并修复依赖安全扫描 |
| M | `frontend/src/components/admin/settings/__tests__/AdminAccessKeysSettings.spec.ts` | ff0429651 补齐管理员密钥中英文权限展示并修复依赖安全扫描 |
| A | `frontend/src/components/admin/settings/__tests__/RegistrationEmailBlacklistEditor.spec.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/components/icons/Icon.vue` | 8a34f1ecf fix(frontend): use the MCP logo and hide MCP when disabled |
| M | `frontend/src/components/keys/UseKeyModal.vue` | 1e0f56d76 fix(frontend): default Codex exports to GPT-6.1 Sol |
| M | `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts` | 1e0f56d76 fix(frontend): default Codex exports to GPT-6.1 Sol |
| M | `frontend/src/components/layout/AppSidebar.vue` | 8a34f1ecf fix(frontend): use the MCP logo and hide MCP when disabled |
| M | `frontend/src/components/layout/__tests__/AppSidebar.spec.ts` | 8a34f1ecf fix(frontend): use the MCP logo and hide MCP when disabled |
| M | `frontend/src/components/payment/__tests__/SubscriptionPlanCard.spec.ts` | d851a3847 fix(frontend): upgrade vue-i18n to v11 and patch decode-uri-component |
| M | `frontend/src/composables/__tests__/useModelWhitelist.spec.ts` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| A | `frontend/src/composables/useAdminAccessKeyPermissions.ts` | b0db521ce 修复权限国际化编码的 ES2020 类型兼容错误 |
| M | `frontend/src/composables/useModelWhitelist.ts` | 7f2876def fix(openai): 打通 GPT Image 2.6 原生图片链路 |
| M | `frontend/src/i18n/locales/en/admin/accounts.ts` | 455751402 fix(antigravity): label Pro plans' separate Argon quota as its own row |
| M | `frontend/src/i18n/locales/en/admin/settings.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/i18n/locales/en/adminAccessKeys.ts` | ff0429651 补齐管理员密钥中英文权限展示并修复依赖安全扫描 |
| M | `frontend/src/i18n/locales/en/common.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/i18n/locales/en/dashboard.ts` | e47abd8f9 fix(usage): 用户导出 CSV 每页取 1000 条，遇面板限流 429 按 Retry-After 等待后重试 |
| M | `frontend/src/i18n/locales/en/index.ts` | cf3300633 feat(frontend): add MCP OAuth consent and connections |
| A | `frontend/src/i18n/locales/en/mcp.ts` | 8a34f1ecf fix(frontend): use the MCP logo and hide MCP when disabled |
| M | `frontend/src/i18n/locales/zh/admin/accounts.ts` | 455751402 fix(antigravity): label Pro plans' separate Argon quota as its own row |
| M | `frontend/src/i18n/locales/zh/admin/settings.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/i18n/locales/zh/adminAccessKeys.ts` | ff0429651 补齐管理员密钥中英文权限展示并修复依赖安全扫描 |
| M | `frontend/src/i18n/locales/zh/common.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/i18n/locales/zh/dashboard.ts` | e47abd8f9 fix(usage): 用户导出 CSV 每页取 1000 条，遇面板限流 429 按 Retry-After 等待后重试 |
| M | `frontend/src/i18n/locales/zh/index.ts` | cf3300633 feat(frontend): add MCP OAuth consent and connections |
| A | `frontend/src/i18n/locales/zh/mcp.ts` | 8a34f1ecf fix(frontend): use the MCP logo and hide MCP when disabled |
| A | `frontend/src/router/__tests__/mcp-route.spec.ts` | cf3300633 feat(frontend): add MCP OAuth consent and connections |
| M | `frontend/src/router/index.ts` | cf3300633 feat(frontend): add MCP OAuth consent and connections |
| M | `frontend/src/stores/app.ts` | 8a34f1ecf fix(frontend): use the MCP logo and hide MCP when disabled |
| M | `frontend/src/types/index.ts` | bf4e16b90 feat(antigravity): show each quota group's real 5-hour and weekly window |
| M | `frontend/src/utils/__tests__/ccswitchImport.spec.ts` | 1e0f56d76 fix(frontend): default Codex exports to GPT-6.1 Sol |
| M | `frontend/src/utils/__tests__/featureFlags.spec.ts` | 8a34f1ecf fix(frontend): use the MCP logo and hide MCP when disabled |
| A | `frontend/src/utils/__tests__/registrationEmailBlacklist.spec.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/utils/ccswitchImport.ts` | e97a48d84 fix: GPT-5.5已不提供服务，默认导出模型改为GPT5.6-SOL |
| A | `frontend/src/utils/codexAuthConfig.ts` | 0d2459d98 fix(frontend): preserve Codex image configuration in key exports |
| M | `frontend/src/utils/featureFlags.ts` | 8a34f1ecf fix(frontend): use the MCP logo and hide MCP when disabled |
| A | `frontend/src/utils/registrationEmailBlacklist.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/views/admin/AuditLogView.vue` | 8a34f1ecf fix(frontend): use the MCP logo and hide MCP when disabled |
| M | `frontend/src/views/admin/SettingsView.vue` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/views/admin/__tests__/SettingsView.spec.ts` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| M | `frontend/src/views/auth/EmailVerifyView.vue` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `frontend/src/views/auth/MCPOAuthAuthorizeView.vue` | b3eb7194c fix(frontend): truncate long MCP client IDs |
| M | `frontend/src/views/auth/RegisterView.vue` | 721dfb282 feat(auth): 注册邮箱精确黑名单，保存时可批量禁用命中的已注册用户 |
| A | `frontend/src/views/auth/__tests__/MCPLoginRedirect.spec.ts` | cf3300633 feat(frontend): add MCP OAuth consent and connections |
| A | `frontend/src/views/auth/__tests__/MCPOAuthAuthorizeView.spec.ts` | 8a34f1ecf fix(frontend): use the MCP logo and hide MCP when disabled |
| A | `frontend/src/views/user/MCPView.vue` | b3eb7194c fix(frontend): truncate long MCP client IDs |
| M | `frontend/src/views/user/UsageView.vue` | e47abd8f9 fix(usage): 用户导出 CSV 每页取 1000 条，遇面板限流 429 按 Retry-After 等待后重试 |
| A | `frontend/src/views/user/__tests__/MCPView.spec.ts` | cf3300633 feat(frontend): add MCP OAuth consent and connections |
| M | `frontend/src/views/user/__tests__/UsageView.spec.ts` | e47abd8f9 fix(usage): 用户导出 CSV 每页取 1000 条，遇面板限流 429 按 Retry-After 等待后重试 |
| M | `frontend/vite.config.ts` | d851a3847 fix(frontend): upgrade vue-i18n to v11 and patch decode-uri-component |

## 本站适配与记录文件说明

| 文件 | 本站兼容处理 |
| --- | --- |
| `backend/cmd/server/wire_gen.go` | 按合并后依赖重新生成装配，接入 MCP 和受限后台任务，保留本站公开媒体清理生命周期。 |
| `backend/internal/handler/openai_audio.go` | 本站音频入口接入上游统一请求读取错误分类和资源预算失败处理。 |
| `backend/internal/handler/openai_images.go` | 合并尺寸路由和异步结算，转移余额冻结所有权给计费任务，并对齐本站 OAuth 回退的尺寸错误判断。 |
| `backend/internal/handler/openai_images_hold_worker_test.go` | 新增请求结束后异步结算持有冻结、持久待结算保留冻结及持久化失败释放冻结的回归。 |
| `backend/internal/handler/openai_images_size_routing_test.go` | 补齐计费夹具，兼容本站同账号重试并验证尺寸分流只使用允许的账号。 |
| `backend/internal/handler/openai_nano_banana.go` | 本站 Nano Banana 入口接入上游统一请求读取错误分类和资源预算失败处理。 |
| `backend/internal/handler/openai_nano_banana_billing_test.go` | 为上游测试桩增加可选响应体，支持冻结结算回归使用真实有效图片。 |
| `backend/internal/handler/openai_videos.go` | 本站视频入口接入上游统一请求读取错误分类和资源预算失败处理。 |
| `backend/internal/pkg/openai/constants.go` | 吸收 GPT Image 2.6 默认模型，保留本站媒体型号及默认顺序，消除合并重复项。 |
| `backend/internal/repository/scheduler_cache.go` | 合并上游调度缓存更新，保留本站 BPS 账号元数据及生图能力缓存。 |
| `backend/internal/service/api_key_auth_cache_impl.go` | 认证快照升至 v28，合并上游隐私和 RPM 状态字段并保留本站 Sol 订阅优化字段。 |
| `backend/internal/service/api_key_auth_cache_merge_test.go` | 新增合并后分组字段缓存往返验证，并拒绝旧 v25/v26/v27 快照。 |
| `backend/internal/service/gemini_chat_completions_compat_service.go` | 合并受限 SSE 读取及配置传递，保留本站实际图片张数和部分失败用量。 |
| `backend/internal/service/gemini_messages_compat_service.go` | 合并受限 SSE 读取，保持本站消息兼容及图片结果处理。 |
| `backend/internal/service/gemini_native_response_signal_test.go` | 适配 Gemini SSE 收集器新增配置参数并保留图片信号回归。 |
| `backend/internal/service/openai_gateway_response_handling.go` | 合入受限流读取和断连处理，同时保留本站 BPS 追踪和用量保护。 |
| `backend/internal/service/openai_gateway_usage.go` | 合入 Terra 长上下文豁免，避免误受本站 Sol 专属开关控制，保留现有媒体计费。 |
| `backend/internal/service/openai_subscription_long_context_test.go` | 覆盖 Sol 开关关闭仍不影响 Terra OAuth 长上下文豁免的边界。 |
| `docs/API_DOCS.md` | 更新 GPT Image 2.6 已补齐后端链路和定价的说明，保留真实上游未验证边界。 |
| `docs/UPSTREAM_SYNC.md` | 追加 0.2.20 同步结果、主要兼容修复、验证与迁移边界。 |
| `docs/upstream-sync-20261107-files.md` | 新增同步来源、行为配置、测试证据、逐文件清单和可执行恢复说明。 |
| `frontend/src/components/account/EditAccountModal.vue` | 同时保留本站 Suno 等账号选项和上游图片尺寸档位控件。 |
| `frontend/src/components/account/__tests__/EditAccountModal.spec.ts` | 合并上游尺寸档位用例与本站账号选项保存断言。 |
| `frontend/src/components/keys/UseKeyModal.vue` | 合入 Codex 原生工具导出更新，保留本站 GPT-6.1 Sol 默认顺序和配置。 |
| `frontend/src/i18n/__tests__/lazyLocales.spec.ts` | 新增实际用户懒加载语言包包含 MCP 页面与授权文案的断言。 |
| `frontend/src/i18n/locales/en/admin/accounts.ts` | 合并上游账号文案，继续复用本站共享语言模块及既有账号选项。 |
| `frontend/src/i18n/locales/en/admin/shared.ts` | 保留共享语言模块拆分并补入 Antigravity 配额窗口英文文案。 |
| `frontend/src/i18n/locales/en/user.ts` | 将英文 MCP 文案补入本站实际使用的用户懒加载入口。 |
| `frontend/src/i18n/locales/zh/admin/accounts.ts` | 合并上游账号文案，继续复用本站共享语言模块及既有账号选项。 |
| `frontend/src/i18n/locales/zh/admin/shared.ts` | 保留共享语言模块拆分并补入 Antigravity 配额窗口中文文案。 |
| `frontend/src/i18n/locales/zh/user.ts` | 将中文 MCP 文案补入本站实际使用的用户懒加载入口。 |
| `frontend/src/stores/app.ts` | 合并 MCP 开关与本站缓存率等公共设置，保留前端按需加载行为。 |
| `frontend/src/types/index.ts` | 合并上游 MCP/账号字段与本站缓存率、模型广场和媒体类型。 |
| `frontend/vite.config.ts` | 适配 vue-i18n v11 移除旧 JIT 定义，保留本站按需分包配置。 |
| `progress.md` | 仅在末尾追加本轮合并与验证记录，保留全部既有历史。 |
