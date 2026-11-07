# 2026-10-28 上游 0.2.19 合并清单

## 来源与范围

- 来源：Wei-Shaw/sub2api 的 main 分支。
- 同步基线：`5c185c5c5d4f5e409e2139806e0964ae944e03e8`。
- 本次目标：`a3298b46ebe239fb0720b068a1eb3383955bdac2`，版本 `0.2.19`；2026-10-28（Asia/Shanghai）拉取并复核。
- 上游增量：76 个提交、261 个路径；最终本地合并提交 `4491998af6177e37b9b5d5041ee296e3dcb56bd6`。
- 用户已明确批准一并合入管理 API 密钥权限及两项数据库迁移代码。只执行本地合并和隔离验证，不运行生产迁移、不推送 Git、不发布镜像或部署。
- 合并前本地 266 项未提交状态已独立保管；恢复时不将它们整体提交。后续构建需使用包含这些定制的完整工作区。

## 行为与配置

- 新增管理 API 密钥，按操作和资源范围授权；说明见 `admin-api-keys.md`。原管理员登录与既有管理密钥仍保留，新增密钥权限不自动授予普通用户。
- 新增分组 `image_count_billing`：默认关闭，仅支持开启生图的 OpenAI/composite 分组。开启时专用图片入口按实际返回张数结算，token 定价不会盖过该开关；混合文本/图片请求继续保留 token 计费。独立生图倍率与本地订阅优化开关仍保留。
- Grok 增加模型目录同步、快照原子更新、响应模型审计与后台诊断；不会把旧管理页面的表单覆盖到最新目录观测结果。
- Gemini 默认模型更新，模型声明/白名单列表对齐，显式缓存同账号重试支持 `gateway.max_same_account_retries_gemini`（默认 4），并识别 fileData 图片输出。
- 同步 GPT-6.1 Terra、DeepSeek 缓存读取与角色兼容、中文 Coding Plan 秒级冷却、Claude UTF-16 指纹计算、超限上传拒绝和 Go 安全依赖更新。

- 追加合入上游 OAuth 非流式 Responses 修复：仍向 ChatGPT 请求流式响应，但按客户端要求聚合为 JSON；识别缺少 Content-Type 的 SSE，并保留未完成/取消状态与用量。最终目标包含 2026-10-28 21:41（Asia/Shanghai）合入的 PR #7979。

## 本站冲突处理

- 保留 Basispoints、Excel OAuth、CLIProxyAPI 风格的 Codex 身份配置、已删除的 312 信号状态，以及最近新增的 Grok 图片/视频和 Suno 转发；没有恢复已移除的 `svggrok-*` 别名。
- Gemini 原生接口保留本站 OpenAI API Key 接入能力，只有适用的 Gemini 分组才使用上游新增的声明模型列表逻辑。
- 账号更新同时保留本站门票/BPS 403 标记和上游 Grok 目录快照；补齐 SQL 模拟返回列。
- 按本站计费接口保留错误向上传播，不将定价失败改成静默零费用；兼容上游新增按张计费分支。
- 模型广场保留已确认的「渠道定价优先、否则默认/官方价格」以及充值比例、媒体实际倍率和映射名。上游按分组模型价展示的规则与本站要求冲突，因此不覆盖本站规则，相关新增测试按本站口径适配。
- 根据合并后 schema 重新生成 Ent 分组代码，保留 `image_count_billing` 与本站 `gpt61_sol_subscription_optimization_enabled`；Wire 装配顺序与生成器一致。
- 最后 5 个提交只影响 OAuth 响应处理及新增测试；合并后恢复本站 Basispoints 提前终止/读取错误的处理，避免上游增量覆盖本地用量与失败保护。
- 保留本站默认模型列表的 GPT-6.1 Sol 首位顺序；补齐本地 Chat 转换器对 GPT-6.1 Terra `reasoning_effort=none` / `reasoning.effort=none` 的保留，避免忽略关闭推理的显式请求。

## 数据库与回滚边界

- `247_scoped_admin_api_keys.sql` 新建管理密钥表，仅存摘要，并保留撤销/软删除记录。
- `248_group_image_count_billing.sql` 新增分组布尔字段，默认 false，不自动开启存量分组。
- 后续部署启动时会自动执行未应用迁移。部署前备份数据库；源码回退不等于数据库回退。

## 验证

- 前端冻结依赖安装、typecheck、lint、400 个测试文件 / 3299 项测试、翻译完整性和生产构建通过；产物位于 `/tmp/sub2api-sync-20261028-xemq5ef3/frontend-dist`，保留既有大 chunk 警告。
- 后端最终全量 `GOMAXPROCS=4 go test -mod=readonly -p 4 -tags=unit ./... -count=1 -timeout=15m -json` 退出码 0；60 个有测试包通过，26416 个通过事件（含子测试），没有跳过指定用例来规避失败。需要真实数据库的管理密钥用例另外显式提供独立测试 DSN 执行。
- service/handler/admin/repository/server/middleware/routes 的 Grok、Gemini、BPS/Excel、视频、图片计费、模型广场定向回归通过；Grok 目录、管理员权限、BPS/Excel、媒体的定向 race 通过。
- 最后追加的 OAuth 非流式增量通过全量后端复测；OAuth 流式/非流式、SSE 聚合、compact、BPS/Excel 定向 race 通过（同一用例集合以 `-parallel 1` 调度，`oauth-followup-race.log`），覆盖客户端非流式标志、缺失 Content-Type、incomplete/cancelled 状态和用量。
- Basispoints、apicompat、tlsfingerprint 和 openai_ws_v2 四包完整 race 通过。首轮命令将 WS 包路径误写在 pkg 下，其他三包已通过，随后按实际 service 路径完成 WS 全包；Terra 修正后 apicompat 全包 race 再次通过。
- 独立 PostgreSQL 18.1 / Redis 8.4 集成通过：完整迁移与幂等/并发、246/248 共存、Grok 目录 CAS 和普通更新、BPS 403 状态与事务；另一个隔离库通过原子计费、退款、去重、请求中删除 Key/账号后的扣费、视频审账；图片超过预估/已释放冻结后重试结算单独通过。
- 新管理密钥的独立 PostgreSQL 权限、生命周期、代理选择、并发轮换和事务回滚用例通过；加 race 复测也通过。测试容器已清理，未使用生产数据库。
- Wire diff 无差异；Ent 从合并后的 schema 重新生成，最终只有预期分组相关文件存在差异。使用完整最终后端源码和本轮前端产物在临时目录执行 embed 生产编译成功，`--version` 返回 `0.2.19` / `4491998af-dirty`；编译源码哈希与最终工作区一致。
- 保留失败证据：生成器首次 readonly 缺少工具依赖，改用临时 modfile；一次过早并行编译读到 Ent 生成中间态，生成完成后废弃该结果并重跑；上游模型目录返回签名与本地 BPS/MiniMax 分支不一致、两个图片计费测试签名、两项广场规则与默认模型排序断言均已适配。后续全量测试进一步暴露 Terra none 被本地转换器丢弃，修复运行逻辑并补覆盖后最终全量通过。未把失败命令标记为成功。
- `git diff --check`、无冲突标记/未合并索引检查通过；逐项核对合并前 4753 个路径，改动仅在上游增量与明确兼容/文档范围。53 个原有未跟踪文件哈希均一致，8 项原有删除保持删除，进度历史未改写。没有调用真实付费模型上游。
- 扩展 race 首轮暴露既有 compact 测试同时写入 Gin 全局模式的测试夹具竞争（两个相关文件与合并前哈希一致），失败日志保留为 `oauth-followup-race-parallel-failed.log`；同一用例集合保持 race 检测开启，以 `-parallel 1` 调度复测通过。未修改无关旧测试，也不据此声称全 service 包在默认并行 race 下通过。
- 后续合并前发现无人持有、已残留约 22 分钟的空 `.git/index.lock`；核对进程和 lsof 后将该锁移入证据目录，随后正常完成合并与局部 stash 恢复，未强制覆盖索引。
- 原始证据：`/tmp/sub2api-sync-20261028-xemq5ef3`，最终后端结果为 `backend-unit-followup.jsonl` / `backend-unit-followup-summary.json`，生产编译为 `production-build-followup.log`；其他失败/成功日志均保留。

## 可执行恢复

备份分支 `codex/pre-upstream-sync-20261028-212354`；stash `955af774301674df7874d00a3e929a688ce2234a`。完整合并前文件归档 `/tmp/sub2api-sync-20261028-xemq5ef3/before.tgz`，共 4753 个路径。
以下命令只在新目录重建合并前工作区，不覆盖当前工作区：

```sh
restore_dir=$(mktemp -d /tmp/sub2api-pre-sync-20261028.XXXXXX)
git worktree add --detach "$restore_dir" codex/pre-upstream-sync-20261028-212354
git -C "$restore_dir" stash apply --index 955af774301674df7874d00a3e929a688ce2234a
tar -xzf /tmp/sub2api-sync-20261028-xemq5ef3/before.tgz -C "$restore_dir"
```

临时目录如需长期保存请另行归档；不要整仓 reset。进度日志仅追加，回退时保留本次记录。

## 上游逐文件来源

| 变更 | 文件 | 本轮来源/用途 |
| --- | --- | --- |
| M | `.gitignore` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `README.md` | 3f1c8f36c chore: update Sponsors |
| M | `README_CN.md` | 3f1c8f36c chore: update Sponsors |
| M | `README_JA.md` | 3f1c8f36c chore: update Sponsors |
| A | `assets/partners/banners/wandapi-cn-dark.svg` | 3f1c8f36c chore: update Sponsors |
| A | `assets/partners/banners/wandapi-cn.svg` | 3f1c8f36c chore: update Sponsors |
| A | `assets/partners/banners/wandapi-en-dark.svg` | 3f1c8f36c chore: update Sponsors |
| A | `assets/partners/banners/wandapi-en.svg` | 3f1c8f36c chore: update Sponsors |
| A | `assets/partners/logos/NiuProxy.jpg` | 3f1c8f36c chore: update Sponsors |
| A | `assets/partners/logos/soleapi.png` | 3f1c8f36c chore: update Sponsors |
| M | `backend/cmd/server/VERSION` | a9240d501 chore: sync VERSION to 0.2.19 [skip ci] |
| M | `backend/cmd/server/wire_gen.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/ent/group.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/ent/group/group.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/ent/group/where.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/ent/group_create.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/ent/group_update.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/ent/migrate/schema.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/ent/mutation.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/ent/runtime/runtime.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/ent/schema/group.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/go.mod` | 0a2a29b32 chore(deps): bump golang.org/x/net to v0.62.0 (GO-2026-6860) |
| M | `backend/go.sum` | 0a2a29b32 chore(deps): bump golang.org/x/net to v0.62.0 (GO-2026-6860) |
| M | `backend/internal/config/config.go` | 4197fa620 feat(gemini): 同账号重试次数改为可配置 gateway.max_same_account_retries_gemini |
| M | `backend/internal/config/config_test.go` | 4197fa620 feat(gemini): 同账号重试次数改为可配置 gateway.max_same_account_retries_gemini |
| M | `backend/internal/handler/admin/account_handler.go` | d4136cea5 fix(grok): 提供账号模型目录差异诊断 |
| A | `backend/internal/handler/admin/account_handler_grok_model_diagnostics_test.go` | dfc13ce7f fix(grok): 修复模型目录审计相关 lint 错误 |
| M | `backend/internal/handler/admin/admin_helpers_test.go` | 9aa16c8c6 fix(pagination): validate numeric ranges and reuse safe offsets |
| M | `backend/internal/handler/admin/announcement_handler.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_key_accounts.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_key_accounts_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_key_codex_import.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_key_codex_import_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_key_commit_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_key_group_models.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_key_proxies.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_key_proxies_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_key_users.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_key_users_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_keys.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_admin_keys_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/custom_legacy_admin_key_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/handler/admin/group_handler.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/internal/handler/admin/idempotency_helper.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/handler/admin/ops_dashboard_handler.go` | f96528048 fix(pagination): keep page offsets from overflowing int |
| M | `backend/internal/handler/admin/promo_handler.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/handler/admin/proxy_handler.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/scoped_admin_content_response.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/scoped_admin_content_response_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/scoped_admin_subscription_response.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/scoped_admin_subscription_response_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/scoped_admin_usage_response.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/handler/admin/scoped_admin_usage_response_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/handler/admin/setting_handler_runtime.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/handler/admin/subscription_handler.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/handler/admin/usage_handler.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/handler/admin/user_handler.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/handler/dto/account_mapper_redact_test.go` | d4136cea5 fix(grok): 提供账号模型目录差异诊断 |
| M | `backend/internal/handler/dto/mappers.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/internal/handler/dto/types.go` | d4136cea5 fix(grok): 提供账号模型目录差异诊断 |
| M | `backend/internal/handler/gateway_models_test.go` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| A | `backend/internal/handler/gemini_cached_content_bound_retry_test.go` | a20741237 fix(gemini): 绑定显式缓存的请求在池模式上游网关账号上按配置同账号重试 |
| M | `backend/internal/handler/gemini_mixed_models_test.go` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `backend/internal/handler/gemini_v1beta_handler.go` | de44e3a0b fix(gemini): 分组账号都声明模型映射时，/v1beta/models 列表与详情按声明的模型应答 |
| A | `backend/internal/pkg/adminaccess/catalog.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/pkg/adminaccess/policy.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/pkg/adminaccess/policy_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/pkg/antigravity/response_transformer.go` | 14a5fe305 fix(antigravity): reject malformed non-streaming tool calls |
| A | `backend/internal/pkg/antigravity/response_transformer_malformed_test.go` | 14a5fe305 fix(antigravity): reject malformed non-streaming tool calls |
| M | `backend/internal/pkg/apicompat/anthropic_to_responses.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `backend/internal/pkg/gemini/models.go` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `backend/internal/pkg/gemini/models_test.go` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `backend/internal/pkg/geminicli/models.go` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `backend/internal/pkg/geminicli/models_test.go` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `backend/internal/pkg/openai/constants.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `backend/internal/pkg/pagination/pagination.go` | f96528048 fix(pagination): keep page offsets from overflowing int |
| M | `backend/internal/pkg/pagination/pagination_test.go` | 9aa16c8c6 fix(pagination): validate numeric ranges and reuse safe offsets |
| M | `backend/internal/pkg/response/response.go` | 9aa16c8c6 fix(pagination): validate numeric ranges and reuse safe offsets |
| M | `backend/internal/pkg/response/response_test.go` | 9aa16c8c6 fix(pagination): validate numeric ranges and reuse safe offsets |
| M | `backend/internal/repository/account_repo.go` | 02db748cd fix(grok): 保留受管目录快照并阻止非法时间与敏感失败日志 |
| A | `backend/internal/repository/account_repo_grok_managed_extra_integration_test.go` | 02db748cd fix(grok): 保留受管目录快照并阻止非法时间与敏感失败日志 |
| A | `backend/internal/repository/account_repo_grok_managed_extra_test.go` | 02db748cd fix(grok): 保留受管目录快照并阻止非法时间与敏感失败日志 |
| A | `backend/internal/repository/account_repo_grok_models.go` | 87266a7b1 fix(grok): 原子保存模型目录观测快照 |
| A | `backend/internal/repository/account_repo_grok_models_integration_test.go` | 87266a7b1 fix(grok): 原子保存模型目录观测快照 |
| A | `backend/internal/repository/account_repo_grok_models_test.go` | b1cb2eb1f fix(grok): 补齐取消与调用方事务的回归断言 |
| M | `backend/internal/repository/account_repo_opencode_go_usage_test.go` | 02db748cd fix(grok): 保留受管目录快照并阻止非法时间与敏感失败日志 |
| M | `backend/internal/repository/account_repo_upstream_billing_probe_update_test.go` | 02db748cd fix(grok): 保留受管目录快照并阻止非法时间与敏感失败日志 |
| M | `backend/internal/repository/api_key_repo.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| A | `backend/internal/repository/custom_account_proxy_selection.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_account_proxy_selection_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_admin_key_group_models.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_admin_key_group_models_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_admin_key_review_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_admin_key_scope.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_admin_key_scope_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_admin_key_transaction.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_admin_keys.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_admin_keys_postgres_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_admin_keys_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_proxy_options.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_proxy_sample_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/repository/custom_proxy_search.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/repository/group_repo.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| A | `backend/internal/repository/pagination_test.go` | 9aa16c8c6 fix(pagination): validate numeric ranges and reuse safe offsets |
| M | `backend/internal/repository/setting_repo.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/repository/wire.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/securityaudit/prompt_handler.go` | f96528048 fix(pagination): keep page offsets from overflowing int |
| M | `backend/internal/securityaudit/prompt_handler_test.go` | 9aa16c8c6 fix(pagination): validate numeric ranges and reuse safe offsets |
| M | `backend/internal/server/api_contract_test.go` | cb2959245 test(api): include image_count_billing in available groups contract |
| M | `backend/internal/server/middleware/admin_auth.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/server/middleware/custom_admin_key_rate_limit_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/server/middleware/custom_admin_keys_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/server/middleware/panel_rate_limit.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/server/routes/admin.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/server/routes/custom_admin_key_access_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/server/routes/custom_admin_key_group_models_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/server/routes/custom_admin_key_payment_rate_limit_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/server/routes/custom_admin_key_user_responses_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/server/routes/custom_admin_keys.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/server/routes/custom_admin_keys_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/server/routes/payment.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/service/account_test_service.go` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `backend/internal/service/account_test_service_gemini_test.go` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `backend/internal/service/admin_account.go` | 02db748cd fix(grok): 保留受管目录快照并阻止非法时间与敏感失败日志 |
| A | `backend/internal/service/admin_account_grok_catalog_test.go` | 02db748cd fix(grok): 保留受管目录快照并阻止非法时间与敏感失败日志 |
| M | `backend/internal/service/admin_group.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/internal/service/admin_group_duplicate.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/internal/service/admin_service.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/internal/service/admin_service_role_test.go` | 88d9c43b3 fix(admin): 管理员保护规则返回 400 而不是 500 |
| M | `backend/internal/service/admin_user.go` | 88d9c43b3 fix(admin): 管理员保护规则返回 400 而不是 500 |
| M | `backend/internal/service/antigravity_gateway_compat.go` | 14a5fe305 fix(antigravity): reject malformed non-streaming tool calls |
| M | `backend/internal/service/antigravity_gateway_gemini.go` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| A | `backend/internal/service/antigravity_gateway_nonstream_malformed_test.go` | 14a5fe305 fix(antigravity): reject malformed non-streaming tool calls |
| M | `backend/internal/service/antigravity_gateway_streaming.go` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `backend/internal/service/antigravity_image_test.go` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `backend/internal/service/api_key_auth_cache.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/internal/service/api_key_auth_cache_impl.go` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| M | `backend/internal/service/api_key_service.go` | 9aa16c8c6 fix(pagination): validate numeric ranges and reuse safe offsets |
| M | `backend/internal/service/billing_service.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `backend/internal/service/billing_service_test.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `backend/internal/service/channel_service.go` | 9eba61736 fix(channels): preserve web_search_emulation platform dimensions on update |
| A | `backend/internal/service/channel_websearch_merge_test.go` | 1352e899e test(channels): use checked type assertions in web search merge tests |
| A | `backend/internal/service/custom_admin_key_accounts.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_accounts_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_codex_import.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_codex_import_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_group_models.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_group_models_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_proxy_selection.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_proxy_selection_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_recovery_review_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_testdb_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_transaction.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_key_transaction_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_keys.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_admin_keys_test.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/internal/service/custom_proxy_options.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/service/deepseek_pricing_test.go` | 5db2d6956 fix(billing): retain DeepSeek V4 Pro pricing after September 14 |
| M | `backend/internal/service/gateway_billing_block.go` | 7bce2cfb4 fix(claude): match JavaScript UTF-16 billing fingerprint semantics |
| A | `backend/internal/service/gateway_billing_block_test.go` | 7bce2cfb4 fix(claude): match JavaScript UTF-16 billing fingerprint semantics |
| M | `backend/internal/service/gateway_billing_header_test.go` | 7bce2cfb4 fix(claude): match JavaScript UTF-16 billing fingerprint semantics |
| M | `backend/internal/service/gateway_service.go` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| M | `backend/internal/service/gateway_usage_billing.go` | 3ce09c8af fix(billing): keep tokens for channel per-request image tiers |
| M | `backend/internal/service/gemini_cached_content.go` | a20741237 fix(gemini): 绑定显式缓存的请求在池模式上游网关账号上按配置同账号重试 |
| M | `backend/internal/service/gemini_cached_content_hardening_test.go` | a20741237 fix(gemini): 绑定显式缓存的请求在池模式上游网关账号上按配置同账号重试 |
| M | `backend/internal/service/gemini_chat_completions_compat_service.go` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| M | `backend/internal/service/gemini_error_policy_single_check_test.go` | 4197fa620 feat(gemini): 同账号重试次数改为可配置 gateway.max_same_account_retries_gemini |
| M | `backend/internal/service/gemini_image_output_accounting.go` | 917dbdc70 fix(gemini): count fileData image outputs for billing |
| M | `backend/internal/service/gemini_image_output_accounting_test.go` | 917dbdc70 fix(gemini): count fileData image outputs for billing |
| M | `backend/internal/service/gemini_messages_compat_service.go` | a20741237 fix(gemini): 绑定显式缓存的请求在池模式上游网关账号上按配置同账号重试 |
| M | `backend/internal/service/gemini_models_endpoints.go` | de44e3a0b fix(gemini): 分组账号都声明模型映射时，/v1beta/models 列表与详情按声明的模型应答 |
| A | `backend/internal/service/gemini_same_account_retries_test.go` | 4197fa620 feat(gemini): 同账号重试次数改为可配置 gateway.max_same_account_retries_gemini |
| M | `backend/internal/service/gemini_vertex_api_key_forward_test.go` | 4197fa620 feat(gemini): 同账号重试次数改为可配置 gateway.max_same_account_retries_gemini |
| M | `backend/internal/service/grok_media.go` | cde6ddfd8 fix(images): reject oversized multipart uploads instead of truncating them |
| A | `backend/internal/service/grok_model_catalog.go` | dfc13ce7f fix(grok): 修复模型目录审计相关 lint 错误 |
| A | `backend/internal/service/grok_model_catalog_store.go` | 87266a7b1 fix(grok): 原子保存模型目录观测快照 |
| A | `backend/internal/service/grok_model_catalog_store_test.go` | b1cb2eb1f fix(grok): 补齐取消与调用方事务的回归断言 |
| A | `backend/internal/service/grok_model_catalog_sync_test.go` | 41a6ff8ac fix(grok): 对齐手动同步与后台目录观察 |
| A | `backend/internal/service/grok_model_catalog_test.go` | 6202da0f5 fix(grok): 统一模型目录解析与观测状态 |
| A | `backend/internal/service/grok_model_diagnostics.go` | bb11e83c4 fix(grok): 裸别名无法核验且自定义 ID 保留大小写 |
| A | `backend/internal/service/grok_model_diagnostics_test.go` | bb11e83c4 fix(grok): 裸别名无法核验且自定义 ID 保留大小写 |
| M | `backend/internal/service/grok_observed_models.go` | 02db748cd fix(grok): 保留受管目录快照并阻止非法时间与敏感失败日志 |
| A | `backend/internal/service/grok_observed_models_log_test.go` | 02db748cd fix(grok): 保留受管目录快照并阻止非法时间与敏感失败日志 |
| M | `backend/internal/service/grok_p2_test.go` | 6202da0f5 fix(grok): 统一模型目录解析与观测状态 |
| M | `backend/internal/service/grok_quota_service_test.go` | 635b83535 fix(grok): 让旧后台观察测试走 CAS 持久化 |
| A | `backend/internal/service/grok_response_model_audit_test.go` | d01f2eb42 fix(grok): 固定聚合前序模型审计并跳过非 Grok |
| M | `backend/internal/service/group.go` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/internal/service/group_service.go` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| A | `backend/internal/service/image_billing_count.go` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| A | `backend/internal/service/image_billing_count_test.go` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| A | `backend/internal/service/image_count_billing_switch_test.go` | 3ce09c8af fix(billing): keep tokens for channel per-request image tiers |
| M | `backend/internal/service/media_price_config.go` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| M | `backend/internal/service/model_plaza_service.go` | d0afece15 fix(model-plaza): prefer group model pricing over image fallback tiers |
| M | `backend/internal/service/model_plaza_service_test.go` | d0afece15 fix(model-plaza): prefer group model pricing over image fallback tiers |
| M | `backend/internal/service/openai_chat_roles.go` | ba5b21879 fix(gateway): 修复 OpenCode 转发 DeepSeek 的角色兼容 |
| M | `backend/internal/service/openai_chat_roles_test.go` | ba5b21879 fix(gateway): 修复 OpenCode 转发 DeepSeek 的角色兼容 |
| M | `backend/internal/service/openai_codex_models_service.go` | b2b1bcf26 fix(models): 模型列表刷新进入 singleflight 后复查缓存，避免并发冷启动重复请求上游 |
| M | `backend/internal/service/openai_codex_transform.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `backend/internal/service/openai_compat_model.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `backend/internal/service/openai_gateway_chat_completions.go` | d01f2eb42 fix(grok): 固定聚合前序模型审计并跳过非 Grok |
| M | `backend/internal/service/openai_gateway_chat_completions_raw.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `backend/internal/service/openai_gateway_grok.go` | 014cc016f fix(grok): 补全原生 Responses 响应模型审计 |
| M | `backend/internal/service/openai_gateway_messages.go` | d01f2eb42 fix(grok): 固定聚合前序模型审计并跳过非 Grok |
| M | `backend/internal/service/openai_gateway_messages_chat_fallback.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `backend/internal/service/openai_gateway_passthrough.go` | 4db8402b5 fix(openai): detect SSE body without Content-Type in OAuth passthrough non-stream |
| M | `backend/internal/service/openai_gateway_response_handling.go` | 3e18e78ab fix(usage): DeepSeek prompt_cache_hit_tokens 计为缓存读取 |
| M | `backend/internal/service/openai_gateway_responses_chat_fallback.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `backend/internal/service/openai_gateway_service.go` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| M | `backend/internal/service/openai_gateway_usage.go` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| A | `backend/internal/service/openai_gpt61_terra_test.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `backend/internal/service/openai_images.go` | cde6ddfd8 fix(images): reject oversized multipart uploads instead of truncating them |
| M | `backend/internal/service/openai_images_responses.go` | cde6ddfd8 fix(images): reject oversized multipart uploads instead of truncating them |
| A | `backend/internal/service/openai_images_upload_limit_test.go` | cde6ddfd8 fix(images): reject oversized multipart uploads instead of truncating them |
| M | `backend/internal/service/openai_model_alias.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| A | `backend/internal/service/openai_oauth_passthrough_non_stream_test.go` | 4db8402b5 fix(openai): detect SSE body without Content-Type in OAuth passthrough non-stream |
| A | `backend/internal/service/openai_usage_deepseek_cache_test.go` | 3e18e78ab fix(usage): DeepSeek prompt_cache_hit_tokens 计为缓存读取 |
| A | `backend/internal/service/pagination_test.go` | 9aa16c8c6 fix(pagination): validate numeric ranges and reuse safe offsets |
| M | `backend/internal/service/pricing_service.go` | 342366742 feat(openai): support GPT-6.1 Terra |
| A | `backend/internal/service/ratelimit_cn_frequency_429_test.go` | de908ed6d fix(ratelimit): 国产 Coding Plan 的频率型 429 改用秒级冷却 |
| M | `backend/internal/service/ratelimit_cn_providers.go` | de908ed6d fix(ratelimit): 国产 Coding Plan 的频率型 429 改用秒级冷却 |
| M | `backend/internal/service/response_model_billing_test.go` | 014cc016f fix(grok): 补全原生 Responses 响应模型审计 |
| M | `backend/internal/service/setting_features.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/service/setting_service.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `backend/internal/service/upstream_models.go` | d4136cea5 fix(grok): 提供账号模型目录差异诊断 |
| M | `backend/internal/service/upstream_models_test.go` | 41a6ff8ac fix(grok): 对齐手动同步与后台目录观察 |
| M | `backend/internal/service/upstream_response_model_test.go` | 014cc016f fix(grok): 补全原生 Responses 响应模型审计 |
| M | `backend/internal/service/wire.go` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/migrations/247_scoped_admin_api_keys.sql` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `backend/migrations/248_group_image_count_billing.sql` | fd998129e feat(billing): support per-image billing for multi-image requests |
| M | `backend/resources/model-pricing/model_prices_and_context_window.json` | 342366742 feat(openai): support GPT-6.1 Terra |
| A | `docs/admin-api-keys.md` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `frontend/src/api/__tests__/admin.accounts.grokModelDiagnostics.spec.ts` | 60931bce3 fix(grok): 展示模型目录诊断并保留编辑草稿 |
| M | `frontend/src/api/admin/accounts.ts` | 60931bce3 fix(grok): 展示模型目录诊断并保留编辑草稿 |
| A | `frontend/src/api/admin/adminAccessKeys.ts` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `frontend/src/components/account/EditAccountModal.vue` | 9534f1c08 fix(grok): 提高删除图标对比度并补齐同步按钮 Label in Name |
| A | `frontend/src/components/account/GrokModelDiagnosticsPanel.vue` | 9534f1c08 fix(grok): 提高删除图标对比度并补齐同步按钮 Label in Name |
| M | `frontend/src/components/account/ModelWhitelistSelector.vue` | dfc13ce7f fix(grok): 修复模型目录审计相关 lint 错误 |
| M | `frontend/src/components/account/__tests__/EditAccountModal.grokUpstream.spec.ts` | dfc13ce7f fix(grok): 修复模型目录审计相关 lint 错误 |
| A | `frontend/src/components/account/__tests__/GrokModelDiagnosticsPanel.spec.ts` | 9534f1c08 fix(grok): 提高删除图标对比度并补齐同步按钮 Label in Name |
| M | `frontend/src/components/account/__tests__/ModelWhitelistSelector.spec.ts` | 21b5b49d0 fix(grok): 统一目录同步入口并保留模式隔离 |
| A | `frontend/src/components/admin/settings/AdminAccessKeysSettings.vue` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `frontend/src/components/admin/settings/__tests__/AdminAccessKeysSettings.spec.ts` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `frontend/src/components/keys/UseKeyModal.vue` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts` | 342366742 feat(openai): support GPT-6.1 Terra |
| M | `frontend/src/composables/__tests__/useModelWhitelist.spec.ts` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `frontend/src/composables/useModelWhitelist.ts` | 31564b2e8 fix(gemini): 刷新 Gemini 默认模型列表，移除已停用的 2.x 模型 |
| M | `frontend/src/i18n/locales/en/admin/accounts.ts` | 7f18239be fix(grok): 对齐模型编辑布局与响应式展示 |
| M | `frontend/src/i18n/locales/en/admin/channels.ts` | 9eba61736 fix(channels): preserve web_search_emulation platform dimensions on update |
| M | `frontend/src/i18n/locales/en/admin/overview.ts` | fd998129e feat(billing): support per-image billing for multi-image requests |
| A | `frontend/src/i18n/locales/en/adminAccessKeys.ts` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `frontend/src/i18n/locales/en/index.ts` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `frontend/src/i18n/locales/zh/admin/accounts.ts` | 7f18239be fix(grok): 对齐模型编辑布局与响应式展示 |
| M | `frontend/src/i18n/locales/zh/admin/channels.ts` | 9eba61736 fix(channels): preserve web_search_emulation platform dimensions on update |
| M | `frontend/src/i18n/locales/zh/admin/overview.ts` | fd998129e feat(billing): support per-image billing for multi-image requests |
| A | `frontend/src/i18n/locales/zh/adminAccessKeys.ts` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `frontend/src/i18n/locales/zh/index.ts` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| M | `frontend/src/types/index.ts` | 60931bce3 fix(grok): 展示模型目录诊断并保留编辑草稿 |
| M | `frontend/src/views/admin/ChannelsView.vue` | 9eba61736 fix(channels): preserve web_search_emulation platform dimensions on update |
| M | `frontend/src/views/admin/GroupsView.vue` | aa1a1b11c fix(billing): gate per-image billing by endpoint and pricing mode |
| M | `frontend/src/views/admin/SettingsView.vue` | e445cc3e2 新增按操作与资源范围授权的管理员 API 密钥 |
| A | `frontend/src/views/admin/__tests__/ChannelsView.webSearchEmulation.spec.ts` | 9eba61736 fix(channels): preserve web_search_emulation platform dimensions on update |

## 额外兼容验证文件

- `backend/internal/pkg/apicompat/reasoning.go`：把新支持的 Terra 纳入显式 none 推理档位保留逻辑。
- `backend/internal/pkg/apicompat/reasoning_gpt61_terra_test.go`：验证平铺与嵌套参数均向 Responses 保留 none。
- `backend/internal/repository/account_repo_codex_ticket_test.go`：适配账号锁定查询新增的 Grok 观测字段，继续验证本站门票保留及伪造字段过滤。
- `backend/internal/repository/image_count_billing_migration_integration_test.go`：验证新增计费开关默认关闭、迁移重复执行不重置开关，且保留原独立倍率与订阅优化配置。
- `docs/UPSTREAM_SYNC.md`：追加本轮来源、结果和回滚点。
- `docs/upstream-sync-20261028-files.md`：本清单。
- `progress.md`：追加本轮实施、验证和恢复记录。
