# sub2api-wq 上游同步记录

记录时间：2026-05-20

## 仓库关系

- `origin`：`https://github.com/iotwq/sub2api.git`
- `upstream`：`https://github.com/Wei-Shaw/sub2api.git`
- `upstream` 是后续功能和修复的来源仓库
- `sub2api-wq` 是在 `sub2api` 最新 commit 基础上继续大改的 fork

## 当前基线

- 当前 `HEAD`：`9cf87fb1` `Merge upstream/main into main`
- 当前 `upstream/main`：`e5d6f172` `Merge pull request #2641 from Arron196/fix/channel-monitor-responses-reasoning`
- 当前 `origin/main`：`1da975eb`
- 当前分支：`main`
- 现状：本地已累计很多改动，合并时优先保留 fork 侧行为，不要为了同步上游而重写本地特性

## 已有辅助分支

- `backup/pre-upstream-sync-2026-04-25`
- `backup/pre-upstream-sync-2026-04-30`
- `backup/pre-upstream-sync-2026-05-07`
- `sync/upstream-main-2026-04-25`

## 推荐同步方式

1. 先 `git fetch upstream`
2. 从当前 `main` 新建同步分支
3. 用 `git merge --no-ff upstream/main` 合入上游
4. 只解决必要冲突，保留 `sub2api-wq` 已有的大改
5. 合并后再回到 `main`，做回归验证

## 高冲突区域

- `frontend/src/views/user/ChatView.vue`
- `frontend/src/views/user/KeysView.vue`
- `frontend/src/components/layout/*`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`
- `frontend/src/router/index.ts`
- `frontend/src/stores/app.ts`
- `frontend/src/style.css`
- `frontend/tailwind.config.js`
- `backend/internal/handler/*`
- `backend/internal/service/*`

## 同步前检查

- 先确认工作区里没有会被误删的本地改动
- 不要用 `reset --hard`、`checkout --` 之类命令清理别人/自己未提交的内容
- 如果上游改动涉及前端，优先跑前端构建或测试
- 如果上游改动涉及后端，优先跑相关 `go test`

## 备注

- 这个仓库和上游已经明显分叉，后续同步更适合“定期 merge + 人工收敛”而不是追求线性历史
- 如果只想快速定位差异，先看 `git diff --stat upstream/main...HEAD`

## 渠道监控 OpenAI 探针说明

- OpenAI 渠道监控支持 `/v1/chat/completions` 与 `/v1/responses` 两种协议。
- Grok 渠道监控支持基础文本探测，按 OpenAI Chat Completions 兼容格式请求 `/v1/chat/completions`。
- 对 `gpt-5.4`、`gpt-5.5` 等 Codex/GPT-5.x 模型，如果 chat-completions 探针收到本站网关返回的 `model_not_found`，监控会自动再用 `/v1/responses` 复测一次。
- OpenAI/Grok 探针向本站网关只发送一次请求；网关在上游失败时按账号 ID 排除已失败候选，最多覆盖 3 个不同账号（不区分 OAuth、Setup Token 或 API Key 类型），任一候选成功即返回成功。外部兼容 endpoint 会忽略该内部探针标记并按普通单请求处理。
- 不同账号探测不会绕过 `model_mapping`：如果分组内没有任何账号声明支持目标模型，网关会在选号阶段直接返回 `model_not_found`，不会向明确不支持的账号盲发请求。
- OpenAI 透传账号在调度与模型可用性诊断中按“允许任意模型”处理，避免透传上游被账号级 `model_mapping` 误判为不支持模型。
- 如果监控仍提示 `model_not_found`，优先检查监控使用的 API Key 是否绑定到正确分组，以及该分组下是否有至少一个 active 且可调度的账号或上游渠道。

## 2026-05-12 选择性同步结果

- 同步范围：`a466e80e..18790386`
- 同步策略：不直接覆盖 `sub2api-wq` 已有本地功能，优先吸收上游最近 2 天新增的功能与修复
- 已同步内容：
  - Airwallex 支付与多币种相关后后端代码、前端页面、配置项与文案
  - Codex JSON / accessToken 批量导入账号相关后端接口、前端类型与设置文案
  - OpenAI usage 在未定价模型下按 `0 cost` 记录 usage 的修复
  - `messages` cache_control 改写开关与 `antigravity_user_agent_version` 设置项
  - `CC Switch` deeplink 构建工具抽取，并保留本地 `chinaapi` 默认 provider 名
  - deploy 配置中的 `ANTIGRAVITY_USER_AGENT_VERSION`、Airwallex CSP 与数据库/Redis 端口暴露修复
- 刻意保留的本地改动：
  - `frontend` 里的本地聊天页面、路由、文案与 `chatWithKey` 入口
  - `openai_gateway_service.go` 中本地已经存在的流式断连 drain 逻辑
  - fork 侧现有品牌命名与站点默认值
- 验证结果：
  - `pnpm --dir frontend run typecheck` 通过
  - `GOPROXY=https://goproxy.cn,direct GOSUMDB=sum.golang.google.cn go test ./internal/service/...` 通过

## 2026-05-20 选择性同步结果

- 同步范围：`18790386..e5d6f172`
- 同步策略：以 `docs/UPSTREAM_SYNC.md` 记录的 `18790386` 为同步基点，构造临时合成基线后合并 `upstream/main`；只把上游增量补丁应用回当前工作区，未重置或清理现有本地修改
- 已同步内容：
  - 钉钉 OAuth 登录、邮件 locale 透传与邮件模板编辑/通知服务相关后端、前端、配置与文案
  - 用户 API Key 用量按日明细、用量按平台拆分、管理端用量列与相关统计修复
  - 渠道监控 API 模式、OpenAI 检测模板、监控协议配置界面与迁移
  - 上游模型定价同步、Gemini 3.5 Flash、Bedrock Claude Code 兼容转换与相关模型列表更新
  - OpenAI Responses/Chat Completions 兼容修复，包括 Responses 强制 Chat Completions、静默拒绝 failover、图片 `n` 参数、空 content、temperature/top_p strip、终止事件与 reasoning 输出兼容
  - Ops SLA 口径、容量/IP/本地客户端错误排除、请求错误图表口径与 ops retry replay 移除
  - 兑换码有效期、日卡额度口径、微信支付 pending reconcile、支付宝扫码二维码、支付二维码强制模式与订阅/余额通知邮件
  - 账号敏感凭证脱敏、分组账号可用数/分组停用阻断 API Key、OIDC 邮箱兼容、TOTP autofill、setup 页面 guard 等修复
- 刻意保留/收敛的本地改动：
  - 继续保留 Airwallex/多币种支付、本地聊天页面、Codex JSON / accessToken 导入、`chinaapi` 默认 provider、`messages` cache_control 改写开关与 `antigravity_user_agent_version`
  - `openai_gateway_chat_completions.go` 同时保留本地“客户端断连后短暂 drain 上游以记录 usage”的逻辑和上游 silent-refusal failover
  - Ops 中保留本地错误请求体脱敏快照字段 `request_body/request_body_truncated/request_body_bytes`，同时接受上游移除 retry/replay 控制字段
  - Docker pnpm 继续使用本地精确版本 `9.15.9`，并吸收上游对可复现构建的说明
- 验证结果：
  - 临时合并工作区：`go test ./internal/handler/... ./internal/service/... ./internal/pkg/apicompat/...` 通过
  - 主工作区验证见本次同步后的命令输出

## 2026-05-24 选择性同步结果

- 同步范围：`e5d6f172..63b0631a`
- 同步策略：先为当前本地工作区创建快照提交，再只应用上次同步点之后的上游增量补丁；修复唯一冲突后保留 fork 侧已有功能，最后用记录性 merge 标记 `upstream/main` 已同步，避免下次重复处理同一批提交
- 已同步内容：
  - 版本更新到 `0.1.130`，补充 RunAPI 赞助商资源与 README 文案
  - 安全依赖更新：`golang.org/x/net`、`golang.org/x/crypto`、`golang.org/x/text`、`golang.org/x/sys`
  - API Key ACL 支持信任反代真实 IP，拒绝日志记录真实客户端 IP
  - 注册邮箱白名单支持后缀通配符，OIDC verified email 快速登录路径加固
  - 内容审计支持按模型生效，并补充 Agent 工具循环去重与输入提取测试
  - 兑换码批量更新、订阅到期邮件提醒开关、前端支付/风控/代理资源链接等管理端改进
  - OpenAI/Bedrock 相关修复：账号测试 Chat Completions 路径、Bedrock Claude Code 兼容、Responses developer role 映射、图片生成上游错误透传、账号冷却调度优化
- 刻意保留/收敛的本地改动：
  - 继续保留本地品牌、聊天页、Airwallex/多币种、Codex 导入、`chinaapi` 默认值以及已有支付设置体验
  - `openai_images_responses.go` 同时保留本地 Responses 终态错误识别和上游新增的结构化图片上游错误透传
  - 支付设置页补回支付配置文档入口，使本地 UI 与现有测试断言保持一致
- 验证结果：
  - `go test ./internal/handler/... ./internal/service/... ./internal/repository/... ./internal/pkg/apicompat/...` 通过
  - `pnpm --dir frontend run typecheck` 通过
  - `pnpm --dir frontend exec vitest run src/views/admin/__tests__/RedeemView.batchUpdate.spec.ts src/views/admin/__tests__/RiskControlView.spec.ts src/components/charts/__tests__/TokenUsageTrend.spec.ts src/views/admin/__tests__/SettingsView.spec.ts src/components/account/__tests__/AccountTestModal.spec.ts` 通过

## 聊天页流式错误排查

聊天页会从浏览器直接向配置的 OpenAI-compatible Base URL 发起流式请求。浏览器报 `Failed to fetch` 时，前端按“连接中断”处理，不再和“模型没有返回有效内容”混在一起。

常见原因：

- 反向代理读取超时时间短于模型实际响应时间。
- 上游流在完整 SSE 终止事件发送前关闭。
- 浏览器、网络或 CDN 在流式响应期间重置连接。
- 用户浏览器无法访问当前配置的 Base URL。

用户侧提示口径：

- 网络或流中断：`连接已中断，未收到完整回复。请检查网络、反向代理超时或稍后重试。`
- 请求完成但确实没有助手内容：`没有收到有效回复。`

排查建议：

- 确认用户浏览器可以访问配置的 API Base URL。
- 为 `/v1/chat/completions` 和 `/v1/responses` 增加反向代理流式读取超时。
- 检查后端请求日志中是否有 upstream stream termination、499/client disconnect 或代理超时记录。

## 2026-06-27 upstream/main 同步结果

- 同步范围：`85a3b122..c2754222`，上游新增 72 个提交，版本推进到 `v0.1.139`。
- 同步策略：先创建 `backup/pre-upstream-sync-20260627-211603`，再 stash 当前未提交改动，合并 `upstream/main` 后恢复本地改动并处理二次冲突。
- 已同步内容：
  - Grok OAuth / subscription / quota probe 相关后端、前端、配置与测试。
  - Codex PAT 导入、Codex CLI-only 检测加固、账号级 app-server 指纹信号。
  - GPT-5.5 Codex instructions、Codex Spark 剥离 `image_generation` 工具、OpenAI chat transport failover。
  - OpenAI / Responses 兼容修复：custom tool schema、function args 去重、response failed 脱敏、无可用模型返回 404。
  - 账单/支付/统计修复：缓存命中统计、订阅汇率、余额透支防护、订单币种显示、支付 provider 卡片显示。
  - README / sponsor / deploy / admin CLI JWT fallback 等同步更新。
- 冲突处理：
  - `README_CN.md` 同时保留本地官方域名/体验说明与上游风险提示、赞助商更新。
  - `frontend/src/components/charts/GroupDistributionChart.vue`、`ModelDistributionChart.vue`、`frontend/src/views/admin/DashboardView.vue` 采用上游安全数值格式化，避免空值显示 `NaN`。
  - `frontend/src/i18n/locales/en.ts`、`zh.ts` 保留本地 Codex JSON/AT 文案并吸收上游 Codex PAT 文案。
  - `backend/internal/service/openai_account_scheduler.go` 保留上游 image native 账号不可用时回退 basic/OAuth 的逻辑。
  - `backend/internal/service/openai_codex_transform.go` 同时保留上游 Spark 去除 image tool 与本地 Agent 自定义工具不混线逻辑。
  - `backend/internal/service/openai_gateway_chat_completions.go` 将本地默认 Codex instructions 仅限制在 API Key Responses 兼容路径，OAuth 路径保持上游不注入默认 instructions 的语义。
- 验证结果：
  - `go test ./internal/service ./internal/server/routes ./internal/handler` 通过。
  - `npm test -- --run src/views/user/__tests__/ChatView.spec.ts src/stores/__tests__/app.spec.ts src/components/admin/account/__tests__/AccountTestModal.spec.ts` 通过。
  - `npm run build` 通过；仅有 Browserslist 数据过期提醒。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260627-211603`。
  - 同步前 stash：`stash@{0}`，消息 `pre-upstream-sync-20260627-211603`。
- 上游 merge 提交：`f3121a35`。

## 2026-07-02 upstream/main 同步结果

- 同步范围：`c2754222..0b8e5eec`，上游版本推进到 `v0.1.143`。
- 同步策略：先创建 `backup/pre-upstream-sync-20260702-225343`，再 stash 当前未提交改动，合并 `upstream/main` 后恢复本地改动并处理二次冲突。
- 已同步内容：
  - OpenAI / Responses 兼容修复：count_tokens 桥接、Codex image bridge tool_choice、compact 跳过生图桥接、Codex OAuth 加密 reasoning 保留、GPT-5.5 Pro Codex 模型名保留。
  - Grok 相关增强：Grok CLI 兼容路由、Grok media 路由、图片编辑上传转换、媒体生成分组开关、默认媒体分组修复。
  - Claude / Anthropic 兼容修复：Claude Code stream keepalive、API Key Bearer 认证、OAuth 请求 dateline 指纹清理、Sonnet 5 适配。
  - 账号与调度：OpenAI quota headroom 调度权重、Spark 链接型影子账号、订阅过期保存、账号级重置额度展示、可用模型统计按 requested model 聚合。
  - 用量、运维与前端：用户用量与管理端口径对齐、IP 地理位置展示、Key/Group 列设置、系统日志 API Key 过滤、分组高峰倍率、支付退款/订阅撤销恢复等修复。
- 冲突处理：
  - `README.md`、`README_CN.md`、`README_JA.md` 同时保留本地 APIKEY/SilkAPI/伊莉思等赞助商块，并吸收上游新增赞助商。
  - `backend/internal/server/routes/gateway.go` 同时保留上游 Grok 图片/视频路由和本地 `/api/nano-banana` 路由。
  - `frontend/src/components/admin/account/__tests__/AccountTestModal.spec.ts` 合并测试 helper，使其同时支持传入 `show` 和自定义账号。
  - `frontend/src/views/admin/AccountsView.vue` 保留本地弹窗按需挂载，并恢复上游 `AccountActionMenu` 常驻挂载与 Spark shadow 事件接线。
- 验证结果：
  - `go test ./internal/service ./internal/server/routes ./internal/handler` 通过。
  - `pnpm exec vue-tsc -b` 通过。
  - `pnpm exec vitest run src/components/admin/account/__tests__/AccountTestModal.spec.ts src/views/admin/__tests__/AccountsView.sparkShadow.spec.ts` 通过。
  - `pnpm build` 通过；仍有既有 Browserslist 数据过期提醒和 Vite 大 chunk 提醒。
  - `git diff --check` 与 `git diff --cached --check` 通过。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260702-225343`。
  - 同步前 stash：`stash@{0}`，消息 `pre-upstream-sync-20260702-225343`。
  - 上游 merge 提交：`290dff33`。

## 2026-07-05 upstream/main 同步结果

- 同步范围：`0b8e5eec..b650bdd6`，上游版本推进到 `v0.1.144`。
- 同步策略：先创建 `backup/pre-upstream-sync-20260705-184318`，再 stash 当前未提交改动，合并 `upstream/main` 后恢复本地改动并处理二次冲突。
- 已同步内容：
  - OpenAI / Responses：按映射后的 billing model 记录 Responses 计费，补充 Codex 图片工具 strip policy，修复 Antigravity Gemini 3.1 Pro 路由。
  - Codex 导入：避免合并 access-only 导入，按 `chatgpt_user_id` 优先匹配 Codex session 导入，补充 refresh token 缺失冲突测试。
  - 用量与错误请求：错误请求列表与用量明细的 UI、排序、筛选、列设置口径对齐。
  - 并发与容量：优化并发槽位清理、分组容量批量摘要、ops 实时账号统计性能与 usage log 队列溢出处理。
  - Anthropic / Fable：将 `7d_oi` Fable 窗口 429 识别为模型级限流。
  - 部署配置：新增 setup migration timeout 可配置项并同步 deploy 示例。
- 冲突处理：
  - `backend/internal/service/openai_codex_transform.go` 同时保留上游 Spark 图片工具 strip 与本地自定义 function tool 不自动注入原生图片工具逻辑。
  - `backend/internal/service/openai_gateway_service.go` 合并上游 Codex 显式图片工具策略与本地 `X-Sub2API-Disable-Image-Bridge` 请求级禁用判断。
  - `backend/internal/service/openai_ws_forwarder.go` 在 WebSocket ingress 中同步保留显式图片工具策略与请求级 bridge 禁用。
  - `backend/internal/service/openai_image_generation_controls_test.go` 拆分并保留账户策略 strip、悬空 `tool_choice` 归一化、自定义 function tool 不注入图片工具三类回归测试。
- 验证结果：
  - `go test ./internal/service -run 'TestOpenAIGatewayServiceForward_(AccountPolicyStripsExplicitImageTool|NormalizesDanglingImageToolChoice|CodexBridgeSkipsCustomFunctionTools|ExplicitImageToolWorksWithBridgeDisabled|CodexBridgeCanBeDisabledByRequestHeader)'` 通过。
  - `go test ./internal/service ./internal/server/routes ./internal/handler` 通过。
  - `pnpm exec vue-tsc -b` 通过。
  - `pnpm exec vitest run src/components/admin/account/__tests__/AccountTestModal.spec.ts src/views/user/__tests__/ChatView.spec.ts src/stores/__tests__/app.spec.ts` 通过；仅有既有 localstorage-file 与 Browserslist 警告。
  - `pnpm build` 通过；仍有既有 Browserslist 数据过期提醒和 Vite 大 chunk 提醒。
  - `git diff --check` 与 `git diff --cached --check` 通过。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260705-184318`。
  - 同步前 stash：`stash@{0}`，消息 `pre-upstream-sync-20260705-184318`。
  - 上游 merge 提交：`e8b64e29`。

## 2026-07-07 upstream/main 同步结果

- 同步范围：`b650bdd6..f68f3b86`，上游版本推进到 `v0.1.146`。
- 同步策略：先创建 `backup/pre-upstream-sync-20260707-215250`，再 stash 当前未提交改动，干净工作树合并 `upstream/main` 后恢复本地改动并处理二次冲突。
- 已同步内容：
  - 批量生图基础能力：新增 batch image 数据模型、队列、计费 hold/settlement、下载、worker、用户端批量生图页面与路由。
  - OpenAI / Responses / Messages：补齐 `/v1/messages` 入站 chat fallback、compact SSE usage 检测、function_call item id 剥离、web search 历史块过滤、请求体解析错误观测。
  - Grok / 图片媒体：Grok video text-only fallback、composer image bridge、Grok 图片最新价格控制。
  - 账号与调度：OpenAI 高级调度器控制、fast force priority 策略、账号请求头覆写、账号导入拖拽与批量导入增强。
  - 支付与订阅：EasyPay 自定义支付方式、订阅 CNY 汇率 opt-in、支付金额/可见支付方式限制修复。
  - 稳定性修复：Redis scan index 清理加固、usage/msg queue 缓存清理、并发缓存与请求体错误路径测试补齐。
- 冲突处理：
  - `deploy/Dockerfile` 保留上游 Node 构建内存参数与 pnpm 固定版本说明。
  - `frontend/src/components/admin/account/ImportDataModal.vue` 合并上游多 JSON/拖拽导入与本地 ZIP/Codex session 导入能力。
  - `frontend/src/__tests__/integration/data-import.spec.ts` 同时保留上游多文件导入回归测试和本地 ZIP/Codex 导入测试。
  - `backend/cmd/server/wire_gen.go` 同时注入本地 `CommunityChatHandler` 与上游 `BatchImageHandler`。
  - `backend/internal/server/routes/gateway.go` 同时保留上游 `/v1/images/batches` 路由与本地 `/v1/videos`、`/v1/api/nano-banana` 路由。
  - `frontend/src/components/layout/AppSidebar.vue` 同时保留本地接口文档、模型问答、用户群聊、文生图入口与上游批量生图入口，并保留 Logo 回首页和群聊红点逻辑。
  - `frontend/src/router/index.ts` 同时保留本地 `/api-docs`、`/community-chat`、`/image` 路由与上游 `/batch-image` 路由。
  - `frontend/src/i18n/locales/en.ts`、`frontend/src/i18n/locales/zh.ts` 合并本地导航文案与上游批量生图文案，中文继续保留“模型问答”。
- 验证结果：
  - `go test ./internal/service ./internal/server/routes ./internal/handler` 通过。
  - `go test -tags=unit ./internal/service -run 'TestRunCheckForModel_OpenAICodexModelNotFoundRetriesResponses|TestOpenAIGatewayServiceDiagnoseModelAvailability_PassthroughAllowsAnyModel|TestOpenAIAccountScheduler_PassthroughAllowsUnmappedModel'` 通过。
  - `pnpm exec vue-tsc -b` 通过。
  - `pnpm exec vitest run src/__tests__/integration/data-import.spec.ts src/components/admin/account/__tests__/AccountTestModal.spec.ts src/views/user/__tests__/ChatView.spec.ts src/stores/__tests__/app.spec.ts` 通过；仍有既有 `--localstorage-file` 与 Browserslist 过期提醒。
  - `pnpm build` 通过；仍有既有 Browserslist 数据过期提醒和 Vite 大 chunk 提醒。
  - `git diff --check` 通过。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260707-215250`，指向 `e8b64e29`。
  - 同步前 stash：`stash@{0}`，消息 `pre-upstream-sync-20260707-215250`。
  - 上游 merge 提交：`d04371cc`。

## 2026-07-10 upstream/main 同步结果

- 同步范围：`f68f3b86..12d811bd`，上游版本推进到 `v0.1.149`。
- 同步策略：先创建 `backup/pre-upstream-sync-20260710-004247`，再 stash 当前未提交改动，干净工作树合并 `upstream/main` 后恢复本地改动并处理二次冲突。
- 已同步内容：
  - 大规模拆分：上游将 `openai_gateway_service.go`、`openai_ws_forwarder.go`、`setting_service.go`、`gateway_service.go`、`usage_log_repo.go`、前端 i18n 等大文件拆分为域模块。
  - Grok / 媒体能力：新增 Grok 4.5 官方支持、Grok 媒体价格拆分、Grok video per-second 计费与 quota probe 稳定性修复。
  - OpenAI / Messages / Responses：补齐 `/v1/messages` failover 与错误回写、response.failed 错误透传、compact body-signal SSE 桥接修复、Responses 到 Anthropic instructions/developer role 映射。
  - 安全与稳定性：site_name HTML escape、site_logo/doc_url sanitize、鉴权/支付/会话缺陷修复、lenient JSON normalization 上限、scheduler 模型映射边界修复。
  - 管理与前端：用量页布局调整与延迟健康列、用户 Token 排行、用户角色编辑、Key 当前并发排序、last used IP、版本徽章历史版本回退。
- 冲突处理：
  - 保留上游 OpenAI gateway / websocket / settings / repository 拆分结构，将本地视频链路、异步视频退款、余额防透支、图片桥接禁用、Grok 视频兼容开关迁移到拆分后的文件中。
  - 前端 i18n 采用上游拆分后的 `locales/{zh,en}/` 目录结构，迁移本地“模型问答、用户群聊、接口文档、文生图、联网搜索、视频失败退款、AI 技术学习交流首页”等文案，删除旧单文件语言包。
  - `frontend/src/views/HomeView.vue` 同时保留上游 URL sanitize 安全修复与本地红色 Three.js AI 学习交流首页。
  - `frontend/src/components/layout/AppSidebar.vue` 同时保留上游侧栏滚动位置持久化与本地用户入口、群聊红点提醒。
  - 补回本地 `IMAGE_WORKSPACE_URL` 公开设置解析与 HTML 注入，保持 Docker/启动环境变量配置文生图工作台地址的能力。
- 验证结果：
  - `go test -tags unit ./internal/service ./internal/repository` 通过。
  - `go test ./internal/service ./internal/repository ./internal/handler ./internal/server/routes` 通过。
  - `pnpm build` 通过；仍有既有 Browserslist 数据过期提醒和 Vite 大 chunk 提醒。
  - `pnpm test:run src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/ChatView.spec.ts src/stores/__tests__/app.spec.ts` 通过；仍有既有 `--localstorage-file` 与 Browserslist 提醒。
  - `git diff --check` 通过。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260710-004247`。
  - 同步前 stash：`stash@{0}`，消息 `pre-upstream-sync-20260710-004247`。
  - 上游 merge 提交：`a7264de6`。

## 2026-07-10 upstream/main 同步结果

- 同步范围：`12d811bd..6dd3274a`，上游版本推进到 `v0.1.150`。
- 同步策略：先创建 `backup/pre-upstream-sync-20260710-155249`，再 stash 当前未提交改动，干净工作树合并 `upstream/main` 后恢复本地改动并处理二次冲突。
- 已同步内容：
  - OpenAI / GPT-5.6：补齐 GPT-5.6 计费、缓存写入价格、max 推理强度兼容、Codex 客户端版本升级到 `0.144.1`。
  - Compact / SSE：新增 compact SSE keepalive 与 raw output_item 保留，修复终态 output 非空但缺 compaction 时的响应合成。
  - 计费与用量：加固 GPT-5.6 用量完整性、request_type 筛选、reasoning effort 模型候选提取、支付恢复和并发审计。
  - 账号刷新与兼容：setup-token 账号纳入后台自动刷新，补充 `parallel_tool_calls` 兼容映射。
  - 前端：补齐英文 overview/resources 缺失文案，调整模型白名单、版本/功能访问相关测试。
- 冲突处理：
  - `backend/internal/pkg/apicompat/types.go` 同时保留本地 `reasoning` 嵌套字段与上游 `parallel_tool_calls` 字段。
  - `frontend/src/stores/app.ts` 保留本地 ChinaAPI 默认站点名/龙图标，同时采用上游 `publicSettingsRequest` 的公共设置请求去重逻辑。
  - `backend/internal/service/openai_gateway_forward.go` 保留上游提前解析 Codex 显式图片工具策略，同时保留本地请求级 image bridge 判断参数。
  - `backend/internal/service/openai_gateway_service.go` 同时保留上游 Codex CLI `0.144.1` 与本地 stream disconnect drain grace。
  - `backend/internal/service/openai_image_generation_controls_test.go` 同时保留上游 image_gen namespace strip 回归测试与本地悬空 `tool_choice`、自定义 function tool 不注入图片工具测试。
- 验证结果：
  - `go test ./internal/service ./internal/repository ./internal/handler ./internal/server/routes` 通过。
  - `pnpm exec vue-tsc --noEmit` 通过。
  - `pnpm build` 通过；仍有既有 Browserslist 数据过期提醒和 Vite 大 chunk 提醒。
  - `pnpm test:run src/stores/__tests__/app.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/ChatView.spec.ts` 通过；仍有既有 `--localstorage-file` 与 Browserslist 提醒。
  - `git diff --check` 通过。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260710-155249`。
  - 同步前 stash：`stash@{0}`，消息 `pre-upstream-sync-20260710-155249`。
  - 上游 merge 提交：`3c9b9045`。

## 2026-07-11 upstream/main 同步结果

- 同步范围：`6dd3274a..e316ebf5`，上游版本推进到 `v0.1.151`。
- 同步策略：先创建 `backup/pre-upstream-sync-20260711-020911`，再将当前暂存、未暂存和未跟踪改动完整保存到 stash；在干净工作树合并 `upstream/main` 后，按原暂存状态恢复本地改动。
- 已同步内容：
  - Codex 协议桥：Responses 到 Chat Completions 的 custom、tool_search、namespace 工具声明、历史调用、流式回程和 tool_choice 兼容，修复 exec/MCP 工具丢失与撞名歧义。
  - Codex 身份：根据最终 User-Agent 配对 originator，统一非透传、透传、WebSocket 和探针的出站身份，避免上游身份错配导致 404。
  - Fast/Flex 策略：支持按可信 API Key 用户配置用户级规则，并保持用户规则优先于全局规则。
  - Anthropic 用量：补齐 Responses 与 Anthropic 双向及流式路径中的 `cache_creation_input_tokens`。
  - 稳定性与 Grok：修复 ops capture writer 释放后的 nil 访问，并保留兼容的 Grok reasoning effort。
- 冲突与兼容处理：
  - 上游 merge 和 stash 恢复均未产生文本冲突；11 个与本地工作区重叠的文件由三方合并自动收敛。
  - 上游新增 Fast/Flex 鉴权测试补入本地 `OpenAIVideoTaskBindingRepository` 构造参数，适配 fork 已有的视频任务绑定依赖，不改变生产行为。
- 验证结果：
  - `go test ./internal/pkg/apicompat ./internal/pkg/openai ./internal/server/middleware ./internal/service ./internal/handler ./internal/repository ./internal/server/routes` 在兼容修复后完整通过；首次运行发现并定位了上述测试构造器参数缺口。
  - `go test ./internal/pkg/apicompat ./internal/pkg/openai ./internal/server/middleware -count=1` 通过。
  - `pnpm exec vue-tsc --noEmit` 通过。
  - `pnpm test:run src/views/admin/__tests__/SettingsView.spec.ts src/stores/__tests__/app.spec.ts` 通过，共 47 项测试；仍有既有 `--localstorage-file`、`router-link` 和 Browserslist 提醒。
  - `pnpm build` 通过；仍有既有 Browserslist 数据过期和 Vite 大 chunk 提醒。
  - `git diff --check` 与 `git diff --cached --check` 通过。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260711-020911`，指向 `3c9b9045`。
  - 同步前 stash：对象 `ae7ed01490bb2d5576e544e2a26022a3fee3b851`，当前为 `stash@{0}`，消息 `pre-upstream-sync-20260711-020911`。
  - 上游 merge 提交：`5c5adad0`。

## 2026-07-12 PR #4009 同步结果

- 同步范围：`Wei-Shaw/sub2api#4009` 当前头 `674d1a25`，合并提交为 `09791722`。
- 同步策略：先创建 `backup/pre-pr-4009-merge-20260712-142050`，再将当前 243 项暂存、未暂存和未跟踪改动保存到 stash；在干净工作树合并 PR 后按原暂存状态恢复。
- 已同步内容：
  - Grok CLI 订阅代理请求在最终 HTTP 传输边界补齐受支持的客户端身份和版本请求头，并支持受约束的 `XAI_GROK_CLI_VERSION` 覆写。
  - Grok OAuth 账号默认路由到 CLI 订阅代理，历史官方 `api.x.ai` 根路径或 `/v1` 配置会自动纠正，同时保留自定义主机和非默认端口。
  - Grok Responses 和账号连接测试支持 OAuth 与标准 xAI API Key 两种账号类型。
  - 转发前移除 xAI 不支持的 `additional_tools` 输入项和 Grok Composer reasoning 参数。
  - 补齐 Grok Composer、Build、4.20 别名及缓存输入回退价格，避免无价格时按零成本记录。
  - 前端账号页支持 Grok API Key，额度条按剩余容量显示，并在“使用密钥”中增加 Grok CLI 和 OpenCode 配置。
- 冲突与兼容处理：
  - 合并提交本身无冲突；恢复本地改动时 8 个重叠文件全部由三方合并成功，没有未解决文件。
  - 合并前后的工作区项目数均为 243，未跟踪文件内容指纹保持 `e541d9523f674fcda097257af4994334c2fa61738d8e3b2a13815466af6e5a43`。
- 验证结果：
  - `go test ./internal/repository ./internal/service` 通过。
  - `go test ./...` 通过。
  - `go mod tidy -diff` 通过，无依赖差异。
  - `pnpm test:run -- ...` 实际执行完整前端套件并通过，共 147 个测试文件、967 项测试；仍有既有 localstorage、Browserslist 和 Vue 测试警告。
  - `pnpm typecheck`、`pnpm lint:check` 和 `pnpm build` 通过；生产构建仍有既有 Browserslist 数据过期和 Vite 大 chunk 提醒。
  - `git diff --check backup/pre-pr-4009-merge-20260712-142050..HEAD` 通过。
- 回退点：
  - 合并前备份分支：`backup/pre-pr-4009-merge-20260712-142050`，指向 `5c5adad0`。
  - 合并前 stash：对象 `46a8d715bb1a9f0cb32720cd79641e1ce33bda61`，消息 `pre-pr-4009-merge-20260712-142050`。
  - PR 合并提交：`09791722cbf9a0afdeab8c93df1d29bcab17784a`；保存当前未提交改动后可执行 `git revert -m 1 09791722cbf9a0afdeab8c93df1d29bcab17784a` 回退本次合并。

## Grok APIKey 自定义上游兼容

- Grok APIKey 账号支持使用 OpenAI 兼容的自定义 Base URL，并统一覆盖模型同步、Responses、Chat Completions、图片生成/编辑和视频生成/查询/下载端点。
- APIKey 自定义地址遵循 `security.url_allowlist`：关闭白名单时沿用通用上游 URL 格式策略；开启时必须命中 `upstream_hosts`，并按 `allow_private_hosts` 控制私网地址。
- Grok OAuth 官方地址始终受信任；账号级自定义转发地址与 APIKey 账号一样遵循 `security.url_allowlist`，避免 OAuth bearer 绕过全局出站策略。
- 不需要启用 `XAI_ALLOW_UNSAFE_URL_OVERRIDES`；该环境变量会扩大所有 xAI URL 覆写范围，不应作为自定义 APIKey 上游的常规配置。

## 2026-07-13 upstream/main 同步结果

- 同步范围：`e316ebf5..7d239d62e`，上游版本推进到 `v0.1.153`，本地合并提交为 `25e6d87ab`。
- 同步策略：创建 `backup/pre-upstream-sync-20260713-214300` 后，将当前暂存、未暂存和未跟踪改动保存为 stash；干净工作树合并 `upstream/main`，随后恢复本地工作区并处理重叠改动。
- 已同步内容：
  - Codex alpha/search 独立端点与按次计费、附加工具桥接、Read 工具参数流式传输、remote compaction v2 和消息 item ID 修复。
  - Grok APIKey 第三方 Base URL、模型同步、prompt cache、Chat/Responses 适配、视频编辑与续写，以及 OAuth 媒体请求改走官方 Imagine API。
  - OpenAI WebSocket ingress 生命周期和连接上限、账号级模型冷却、GPT-5.6 OAuth 测试、账号 plan type 手动覆盖。
  - API Key 最近使用 IP 查询优化、DataTable 滚动性能、本地日期范围一致性、静态资源缓存和 Apple container 部署支持。
- 冲突与兼容处理：
  - README 和 Grok 账号基础 URL 行为采用上游版本；OAuth 文本继续走 CLI 代理，OAuth 媒体走官方 API。
  - 路由同时保留上游 Alpha Search、Grok 视频编辑/续写和本地 OpenAI 视频任务、视频文件下载、Nano Banana 接口。
  - Grok APIKey 自定义上游继续遵循本项目 `security.url_allowlist`，Responses、Messages、WebSocket bridge、媒体与模型同步共用该策略，同时保留上游缓存身份请求头。
  - 用量结构同时保留上游 `WebSearchCalls` 与本地 `RequestCount`、`MediaType`、余额预冻结防重复扣费字段。
- 验证结果：
  - `go test -tags=unit ./internal/handler ./internal/server/routes ./internal/service ./internal/repository ./internal/server/middleware -count=1` 通过。
  - 前端完整 Vitest 套件通过，共 151 个测试文件、1000 项测试；仍有既有 localStorage、Vue 测试环境和 Browserslist 提醒。
  - `pnpm build` 通过；仍有既有 Browserslist 数据过期和 Vite 大 chunk 提醒。
  - `git diff --check` 通过，无未解决冲突。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260713-214300`，指向 `09791722`。
  - 同步前工作区 stash：对象 `3ce8050f170e0d2b3634f3e8584378aa53f212d2`，消息 `pre-upstream-sync-20260713-worktree`。
  - 保存当前工作区后，可执行 `git revert -m 1 25e6d87ab5bdf3cf7470579a54f2549466191e12` 回退本次上游合并。

## 2026-07-14 upstream/main 同步结果

- 同步范围：`7d239d62e..da85cc7e4`，共 68 个上游提交、236 个文件；上游 `VERSION` 推进到 `0.1.155`，本地合并提交为 `7302e1a9e`。
- 同步策略：创建 `backup/pre-upstream-sync-20260714-221620` 后，将当前暂存、未暂存和未跟踪改动完整保存为 stash；干净工作树合并 `upstream/main`，随后恢复本地工作区并解决 14 个内容冲突。
- 已同步内容：
  - Grok Web SSO 批量导入、导入后自动探测、Free 套餐 24 小时滚动额度、额度重置识别、reasoning 清洗和渠道健康监控。
  - OpenAI 原生 Responses namespace、APIKey Codex 模型清单、图片流式完成与非流式 keepalive、Responses Lite 工具保护、HTTP/2 keepalive 和额度重置识别。
  - OpenAI 账号级长上下文计费开关及用量标记、系统日志 host 筛选、可选 Server-Timing 指标。
  - 调度器全量重建并发合并、outbox 延迟修正、账号自动暂停与代理到期事件驱动刷新。
- 冲突与兼容处理：
  - Grok 监控采用上游专用适配器和默认 `grok-4.5`；本地先行的 `171` 迁移与上游幂等 `176` 迁移可顺序执行。
  - OpenAI 图片桥同时保留上游 Responses Lite 跳过、显式图片工具防重复注入和本地请求级禁用、自定义 function tool 保护。
  - WebSocket 图片桥采用相同 Responses Lite 与本地请求级策略，保留上游连接生命周期修复。
  - 管理员使用记录同时显示本地视频失败退款负数金额和上游长上下文 `x2` 标记。
  - 部署示例同时保留本地 `IMAGE_WORKSPACE_URL` 与上游图片非流式 keepalive 配置。
  - 上游新增 Codex 模型测试补入本地 `OpenAIVideoTaskBindingRepository` 空参数，不改变生产行为。
- 验证结果：
  - `go test -tags=unit ./... -count=1` 从 `backend/` 通过，覆盖 service、handler、repository、server、pkg、Ent 和 migrations。
  - 前端完整 Vitest 套件通过，共 159 个测试文件、1065 项测试。
  - `pnpm exec eslint src/components/admin/usage/UsageTable.vue src/composables/useChannelMonitorFormat.ts src/constants/channelMonitor.ts` 通过。
  - `pnpm build` 通过；仍有既有 Browserslist 数据过期和 Vite 大 chunk 提醒。
  - `git diff --check`、`git diff --cached --check` 通过，无未解决冲突。
  - `go mod tidy -diff` 仅报告同步前与上游均已存在的 8 条冗余 `go.sum` 校验和，本次未扩大范围清理。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260714-221620`，指向 `25e6d87ab`。
  - 同步前工作区 stash：对象 `c4d2844119552dac00ebd1ee0edda69ca9d516cc`，消息 `pre-upstream-sync-20260714-221620-worktree`。
  - 保存当前工作区后，可执行 `git revert -m 1 7302e1a9e3b3074ffc4ff900a23cded9b38ec1d8` 回退本次上游合并。

## 2026-07-15 upstream/main 同步结果

- 同步范围：`da85cc7e4..eb2b8632d`，共 145 个上游提交、301 个文件；上游版本推进到 `0.1.156`，本地合并提交为 `2b35cbe02`。
- 同步策略：创建 `backup/pre-upstream-sync-20260715-221149` 后，将完整工作区保存为 stash；在干净工作树合并 `upstream/main`，随后恢复本地工作区并解决 13 个内容冲突。
- 已同步内容：
  - OpenAI Agent Identity 独立导入与恢复、Responses 首输出边界和流事件修复、APIKey 透传故障切换、图片尺寸与 Codex 工具兼容改进。
  - Grok OAuth 凭据失效切换、账号池主动恢复、账号级自定义上游和请求头、图片模型路由保护、vision 与 Messages 工具桥接。
  - 调度快照分组生命周期、退役与降级重建修复，内容审核关键词热路径优化，以及 Ops 错误记录和查询增强。
  - 管理端账号复制、静态账号重复创建保护、表格缓存修复、订阅套餐币种和管理员充值返佣设置。
- 冲突与兼容处理：
  - Grok URL 构建、安全校验、Responses/Chat 调用和 OAuth 自定义转发采用上游实现；本地补回 `/v1/models` 同步和 `/v1/videos/{id}/content` 下载，并保留 Range 转发。
  - OpenAI Responses 图片桥同时保留上游 Codex 图片 function-tool 保护和本地请求级/Agent 自定义工具保护，避免重复或混合注入。
  - Responses 流式处理采用上游首输出守卫、事件边界刷新和缺失终止事件语义；本地旧测试同步到新行为。
  - Grok OAuth 建号同时保留上游自定义上游配置和本地额外配置构建；账号页保留按需挂载弹窗，并接入上游账号复制事件。
  - Gemini 原生 OpenAI 账号、Grok APIKey 自定义上游、Grok 视频下载、OpenAI 视频/Nano Banana、社区聊天和本地首页定制均保留。
  - 补充 `@intlify/message-compiler` 直接开发依赖，使上游新增的国际化消息编译测试在 pnpm 严格依赖模式下可运行。
- 验证结果：
  - `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm test:run` 通过，共 163 个测试文件、1160 项测试。
  - `pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 通过；仍有既有 localStorage、Vue 测试环境、Browserslist 数据和 Vite 大 chunk 提醒。
  - `pnpm install --frozen-lockfile` 通过；依赖声明与锁文件一致。
  - `git diff --check`、`git diff --cached --check` 通过，无未解决冲突。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260715-221149`，指向 `7302e1a9e3b3074ffc4ff900a23cded9b38ec1d8`。
  - 同步前工作区 stash：对象 `26d13598a2f9c355cef5417fb5afa76ff2bcdaaf`，消息 `pre-upstream-sync-20260715-221149-worktree`。
  - 保存当前工作区后，可执行 `git revert -m 1 2b35cbe028c3c805625be44d4e51b9efb59f54fb` 回退本次上游合并。

## 2026-07-16 upstream/main 同步结果

- 同步范围：`eb2b8632d..bc2244c83`，共 88 个上游提交、350 个文件；上游版本推进到 `0.1.158`，本地合并提交为 `4af4d7d83`。
- 同步策略：创建 `backup/pre-upstream-sync-20260716-230929` 后，将当前暂存、未暂存和未跟踪改动完整保存为 stash；在干净工作树合并 `upstream/main`，随后恢复本地工作区并解决 15 个内容冲突。
- 已同步内容：
  - Grok 端点预设与手动切换、WebSocket v2 传输及相关协议稳定性修复。
  - 异步图片任务、对象存储、图片输入 token 计价和批量图片任务能力。
  - 上游计费倍率探测、OpenAI 成本感知调度、账号用量与调度信息展示。
  - 审计日志、会话绑定、敏感管理操作二次 2FA 验证和安全策略增强。
  - 用户批量限额、分组复制和渠道监控复制等管理能力。
- 冲突与兼容处理：
  - 路由与依赖注入采用上游审计、二次验证和异步图片任务结构，同时保留本地社区群聊、OpenAI 视频、Nano Banana 和视频内容下载路由。
  - OpenAI 账号端点能力同时保留上游 `alpha_search`、`responses` 与本地 `gemini_native`；Gemini 原生调度不参与 OpenAI 上游成本倍率优选。
  - OpenAI 调度采用上游按模型反馈和成本感知签名，本地 Nano Banana、视频处理器及测试夹具同步适配映射后的上游模型参数。
  - Grok 媒体采用上游仅对 OAuth 官方 CLI 目标注入客户端请求头，同时保留视频内容的 `Accept: */*`、`Range` 和 `If-Range` 转发。
  - Anthropic Haiku 采用上游完整 mimic/context-management 语义，并保留本地真实 Claude Code 请求头与请求体透传回归覆盖。
  - 前端账号页采用上游异步弹窗、计费倍率探测和二次验证，同时保留本地按需挂载；渠道监控文案同时保留上游复制能力与本地超时提示。
- 验证结果：
  - `go generate ./cmd/server` 通过，生成结果同时包含社区群聊、异步图片、审计、二次验证和视频任务仓储依赖。
  - `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 174 个测试文件、1226 项测试。
  - 前端测试和构建仅有既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 提醒。
  - `git diff --check` 与 `git diff --cached --check` 通过，无未解决冲突。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260716-230929`，指向 `2b35cbe028c3c805625be44d4e51b9efb59f54fb`。
  - 同步前工作区 stash：对象 `6679e88607a517f677bc81a5a3c5341069d12371`，消息 `pre-upstream-sync-20260716-230929-worktree`。
  - 保存当前工作区后，可执行 `git revert -m 1 4af4d7d83e4982eb6d9c1182ad029a6a9e2ba831` 回退本次上游合并。

## 2026-07-17 upstream/main 同步结果

- 同步范围：`bc2244c83..57914967c`，共 36 个上游提交、159 个文件；上游版本推进到 `0.1.160`，本地合并提交为 `7d3ce6bb9`。
- 同步策略：创建 `backup/pre-upstream-sync-20260717-210152` 后，将当前暂存、未暂存和未跟踪改动完整保存为 stash；在干净工作树合并 `upstream/main`，随后恢复本地工作区并解决 3 个内容冲突。
- 已同步内容：
  - OpenAI 兼容提示词安全审计，包括可配置审计节点、完整提示词事件、管理控制台、批量删除和两项数据库迁移。
  - Grok 媒体账号资格识别、不可用账号隔离、参考图载荷归一化和调度快照资格保留。
  - 显式图片生成意图修复，避免被动 `image_gen` namespace 误触发图片权限、并发槽或 Responses 能力限制。
  - 审计日志与会话绑定统一可信客户端 IP 规则，S3 配置接入二次 TOTP 验证。
  - Stripe 支付依赖按需加载并独立分包，减少非支付页面的公共依赖。
- 冲突与兼容处理：
  - Handler 依赖注入采用上游带安全审计协调器的 `ProvideGatewayHandler` 与 `ProvideOpenAIGatewayHandler`，同时保留本地社区群聊 Handler。
  - 调度快照测试同时保留上游 Grok 媒体资格字段和本地 OpenAI `gemini_native` 端点能力字段。
  - Vite 采用本地标准化路径和细分 vendor 规则，同时保留上游 Stripe 懒加载的独立 `vendor-stripe` 分包。
  - 上游提示词审计 Wire 集合补充 `PromptAdminService` 到 `PromptService` 的接口绑定，保证依赖注入可重新生成。
  - 本地 Nano Banana 和 OpenAI 视频提交入口接入统一 `checkSecurityAudit` 链，并登记到上游 POST 路由审计覆盖守卫，不使用无提示词豁免。
  - Stripe 懒加载测试兼容标准化后的 `moduleId` 变量名，仍验证独立分包规则位于公共 vendor 回退之前。
- 验证结果：
  - `GOPROXY=https://goproxy.cn,direct go generate ./cmd/server` 通过，生成结果同时包含提示词审计、本地社区群聊、异步图片和视频任务仓储。
  - `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 180 个测试文件、1259 项测试。
  - 前端测试和构建仅有既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 提醒。
  - `git diff --check` 与 `git diff --cached --check` 通过，无未解决冲突。
- 回退点：
  - 同步前备份分支：`backup/pre-upstream-sync-20260717-210152`，指向 `4af4d7d83e4982eb6d9c1182ad029a6a9e2ba831`。
  - 同步前工作区 stash：对象 `89626cecd82c4ac14036981ba6fa18fa48f54ab3`，消息 `pre-upstream-sync-20260717-210152-worktree`。
  - 保存当前工作区后，可执行 `git revert -m 1 7d3ce6bb95e39b16f21fd6ce2ebb65370463fd9e` 回退本次上游合并。
## 2026-07-18 upstream/main 同步结果

- 同步范围：`57914967c..d4b9797ff`，共 62 个上游提交、257 个文件；上游版本推进到 `0.1.161`，本地合并提交为 `3908c9365`。
- 同步策略：创建 `backup/pre-upstream-sync-20260718-232215`，将当前暂存、未暂存和未跟踪改动完整保存为 stash `c964f9ecaaca50e569237941f0dc9d37be65f1b9`，在干净工作树合并上游后恢复本地改动并解决 10 个内容冲突。
- 已同步内容：鉴权缓存失效 outbox、入口拒绝审计与清理、Grok 受保护视频内容和媒体模型映射、OAuth 恢复、模型级临时冷却、OpenAI WebSocket 生命周期、订阅续期、上游计费探测、Anthropic 监控和 Docker 跨架构构建修复。
- 冲突与兼容处理：
  - Grok 媒体保留上游 request/task ID 兼容、视频状态与 MP4 内容下载、用户/API Key 账号绑定；同时保留本地 Grok 自定义模型映射、媒体能力和 Nano Banana 路由。
  - 渠道监控合并上游 Anthropic 文本块提取与本地 OpenAI Responses、Chat、Grok 及多块文本兼容提取。
  - OpenAI CC/Responses 故障切换使用归一化包装状态码，并保留上游账号禁用判断，避免已禁用账号继续同账号重试。
  - 前端使用上游 branding favicon 和法律文档加载骨架，同时保留本地语言切换监听与品牌 Logo fallback。
  - 补齐上游构造器变更对应的本地测试 stub，并将 Ops 参数断言更新为合并后的 41 个字段；补充 Grok 视频内容两阶段状态/内容请求测试。
- 特殊补丁：上游合并代码引用但未定义 `openAIStreamDisconnectDrainGrace`，从本地已有实现补回 `5s` 常量，避免 OpenAI WebSocket 代码无法编译。
- 验证结果：
  - `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run`（182 个测试文件、1278 项测试）、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过。
  - `git diff --check`、`git diff --cached --check` 通过；前端仍有既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 警告。
- 回退点：保存当前工作区后，可执行 `git revert -m 1 3908c9365` 回退本次上游合并；同步前备份分支为 `backup/pre-upstream-sync-20260718-232215`，原工作区快照为 stash `c964f9ecaaca50e569237941f0dc9d37be65f1b9`。

## 2026-07-20 upstream/main 同步结果

- 同步范围：`d4b9797ff..e625ce3b3`，上游新增 114 个提交，版本从 `0.1.161` 推进到 `0.1.162`；本地合并提交为 `4b8bec155`。
- 同步策略：创建 `backup/pre-upstream-sync-20260720-223609`，将当前暂存、未暂存和未跟踪改动完整保存为 `stash@{0}`（对象 `5d2ba650dbc5bd0fc147c3d29b6040aab7e246f1`），在干净工作树合并上游后恢复本地改动。
- 已同步内容：上游配置与环境可达性校验、S3 密钥加密与备份、Grok 配额和 token provider、图片存储设置、异步图片与账户管理、审计/会话绑定安全加固、OpenAI Responses/Anthropic 兼容修复、管理员设置与前端批量图片指南等。
- 冲突与兼容处理：Logo 冲突保留本地 ChinaAPI/Dragon 默认 Logo，同时保留上游 `logo.svg`；`gateway.go` 保留上游 `messages/count_tokens` 以及本地 `/pg/assets`、音频接口；聊天测试同时保留上游 OAuth system message 与本地 APIKey Responses 指令测试。
- 本地恢复的功能保持不变：Gemini/Grok、OpenAI 视频/音频/素材上传/Nano Banana、社区群聊与私聊站长、渠道监控、首页/登录/接口文档和相关文档定制。
- 兼容性修正：同步上游新增的系统更新/回滚 15 分钟超时后，补齐对应前端回滚测试的 Axios 配置断言。
- 验证结果：
  - `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile` 通过。
  - `pnpm test:run --reporter=dot` 通过，185 个测试文件、1290 项测试。
  - `pnpm lint:check`、`pnpm typecheck`、`pnpm build` 通过；仅保留既有 localStorage/Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 非致命提醒。
  - `git diff --check`、`git diff --cached --check` 通过，无未解决冲突。
- 回退点：保留 `backup/pre-upstream-sync-20260720-223609` 与 `stash@{0}`；如需撤销本次上游合并，在先保存当前工作区后执行 `git revert -m 1 4b8bec155`，不要清理本地定制改动。

## 2026-07-22 upstream/main 同步结果

- 同步范围：`e625ce3b3..60013c5f1`，上游新增 69 个提交，版本从 `0.1.162` 推进到 `0.1.163`；本地合并提交为 `87c92d9d5f8e4e2da9d634eaf1ada4b46c22d3af`。
- 同步策略：创建 `backup/pre-upstream-sync-20260722-214109`，将当前暂存、未暂存和未跟踪改动完整保存为 stash `6fa7433ae927b3320474cabc4d0b70512691874a`；在干净工作树合并上游后恢复本地改动并解决 6 个内容冲突。
- 已同步内容：分组 reasoning effort 策略、Grok 错误归一化与工具协议、OpenAI 客户端工具兼容、调度快照和最近使用时间修复、用量筛选及倍率显示精度、账号和支付页面体验改进。
- 冲突与兼容处理：
  - OpenAI 调度保留上游账号排除原因，同时保留本地 Gemini 原生端点能力快照。
  - OpenAI APIKey 透传账号继续允许未配置显式模型映射的请求，保持自定义兼容上游可用。
  - Grok 保留上游内容策略 403 隔离，避免请求级安全拒绝错误消耗其他账号。
  - 对外层 HTTP 400 包裹内部 `401/403/429/501/502/503` 的响应继续进行严格格式识别，并以归一化状态驱动冷却、池内同账号重试和故障切换。
  - Grok OAuth/APIKey 模型同步采用上游实现；APIKey 自定义上游继续使用本地 URL 构建器和全局出站安全策略。
- 本地功能保持不变：Gemini 原生 OpenAI 账号、Grok APIKey 自定义上游、OpenAI 视频/音频/素材上传/Nano Banana、社区群聊与私聊站长、渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run --reporter=dot`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 190 个测试文件、1317 项测试。
  - 前端构建仅有既有 Browserslist 数据过期和 Vite 大 chunk 非致命提醒。
  - `git diff --check`、`git diff --cached --check` 通过，无未解决冲突；`upstream/main` 已确认是当前 `HEAD` 的祖先。
  - `go mod tidy -diff` 只报告历史 `go.sum` 冗余项，本次未扩大范围清理。
- 回退点：保留 `backup/pre-upstream-sync-20260722-214109` 与 stash `6fa7433ae927b3320474cabc4d0b70512691874a`；如需撤销本次上游合并，先保存当前工作区，再执行 `git revert -m 1 87c92d9d5f8e4e2da9d634eaf1ada4b46c22d3af`。

## 2026-07-23 upstream/main 同步结果

- 同步范围：`60013c5f1..cb24522dd`，上游新增 43 个提交，版本从 `0.1.163` 推进到 `0.1.164`；本地合并提交为 `8423768fc999ce5e0344ff5945cc247a8de04f9c`。
- 同步策略：创建 `backup/pre-upstream-sync-20260723-223356`，将当前暂存、未暂存和未跟踪改动完整保存为 stash `03ad1dea8f06e3bd10490e9bf6afc39f52def1cb`，在干净工作树合并上游后恢复本地改动并解决 5 个内容冲突。
- 已同步内容：复合分组模型路由与计费归属、Ollama Cloud 用量刷新、OpenAI 代理流断路器、Grok 402 冷却、模型与定价归一化、Alipay 移动端预创建支付链接，以及 GPT-5.6 测试模型修正。
- 冲突与兼容处理：
  - Gemini 原生处理同时采用上游复合分组的有效目标平台判断，并保留本地 OpenAI 分组 Gemini 原生能力调度、账号能力错误提示和签名链路隔离。
  - OpenAI 图片记录同时保留上游复合模型的客户端请求模型计费字段和本地媒体余额预扣/实际成本结算，避免图片请求重复扣费。
  - 视频路由同时保留上游复合 Grok 状态/内容查询和本地 OpenAI 视频提交、查询、MP4 下载、Nano Banana、音频及 `/pg/assets` 素材上传接口；开发 Compose 同时保留代理流断路器和图片工作区地址配置。
  - 路由测试同时覆盖复合 Grok 查找、OpenAI 视频任务路径和非 Grok 平台对 Grok 媒体接口的拒绝行为。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok 视频、音频和素材上传、Nano Banana、社区群聊与私聊站长、渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 从 `backend/` 通过。
  - `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm test:run --reporter=dot` 从 `frontend/` 通过，共 195 个测试文件、1353 项测试。
  - `pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；仍有既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 非致命提醒。
  - `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q` 通过。
  - `git diff --check`、`git diff --cached --check` 通过，无未解决冲突；`upstream/main` 已确认是当前 `HEAD` 的祖先。
  - `go mod tidy -diff` 只报告历史 `go.sum` 冗余项，本次未扩大范围清理。
- 回退点：保留 `backup/pre-upstream-sync-20260723-223356` 与 stash `03ad1dea8f06e3bd10490e9bf6afc39f52def1cb`；如需撤销本次上游合并，先保存当前工作区，再执行 `git revert -m 1 8423768fc999ce5e0344ff5945cc247a8de04f9c`。

## 2026-07-26 upstream/main 同步结果

- 同步范围：`cb24522dd..2730c1c43`，共 54 个上游提交、168 个文件；上游版本从 `0.1.164` 推进到 `0.1.165`，本地合并提交为 `235879a305bd3d168ceef903a18c9be2d87ca39e`。
- 同步策略：创建 `backup/pre-upstream-sync-20260726-144816`，将当前 263 个已跟踪及未跟踪变更路径完整保存为 stash `9205b0ba0d922e8f52ae9321b5cfaa9ab6da0173`；在干净工作树合并上游后恢复全部本地改动，stash 保留未删除。
- 已同步内容：OpenAI Live gateway 与 macOS attestation、Claude Opus 5、Responses item ID 和 namespace 清理、usage log session ID、pool 模式重试与 5xx 冷却、Gemini Chat Completions 图片输出、Ollama Cloud 用量刷新、注册邮箱别名去重、公告 Markdown 样式及前端依赖安全更新。
- 冲突与兼容处理：
  - OpenAI 图片入口采用上游 `img_quality`/`img_size` 日志字段，并保留本地尺寸档位、响应格式和输出格式观测字段。
  - 图片成功结算同时保留本地余额预扣/实际成本捕获和上游 session ID 用量记录，避免重复扣费或会话信息丢失。
  - OpenAI gateway 同时注入上游 Live attestation provider/cipher 与本地视频任务绑定仓库，重新生成 Wire 依赖图。
  - 为本地分组页测试补齐上游新增的 Live 能力探测 mock，不改变生产逻辑。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok 视频、音频和素材上传、Nano Banana、社区群聊与私聊站长、渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run --reporter=dot`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 197 个测试文件、1366 项测试。
  - 前端仅保留既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 非致命提醒。
  - `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q` 通过。
  - `git diff --check`、`git diff --cached --check`、未解决冲突检查和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
  - `go mod tidy -diff` 只报告可删除的历史 `go.sum` 冗余项，本次未扩大范围清理。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 235879a305bd3d168ceef903a18c9be2d87ca39e`；同步前备份分支为 `backup/pre-upstream-sync-20260726-144816`，完整工作区 stash 对象为 `9205b0ba0d922e8f52ae9321b5cfaa9ab6da0173`。

## 2026-07-27 upstream/main 同步结果

- 同步范围：`2730c1c43..59ce11c78`，共 62 个上游提交、142 个文件；上游版本从 `0.1.165` 推进到 `0.1.166`，本地合并提交为 `aa04f43bc292a96db4e1c9b63cb6f83fccf6108e`。
- 同步策略：创建 `backup/pre-upstream-sync-20260727-213337`，将当前 266 个已跟踪及未跟踪变更路径完整保存为 stash `d5351da0a0fd0b573ee462d20b6e3c7f23192b42`；在干净工作树合并上游后恢复全部本地改动，stash 保留未删除。
- 已同步内容：面板 API 限流与设置项、Responses/Anthropic 工具配对和 reasoning 故障切换、Gemini 号池可重试错误修复、Antigravity OpenAI 兼容转发、WebSocket 模型跟踪、用量模型与请求 ID 筛选、按币种聚合支付统计，以及渠道列表和注册页移动端体验改进。
- 冲突与兼容处理：
  - 服务器路由采用上游面板限流器及 Auth/User/Admin/Payment 新签名，同时保留本地社区群聊路由注册，避免安全能力或群聊入口任一丢失。
  - Grok WebSocket HTTP bridge 测试采用上游增强的模型映射 hook 和映射结果断言，同时保留本地既有 bridge 场景。
  - 上游 `go.sum` 未包含 Wire 命令行依赖 `github.com/google/subcommands v1.2.0` 的校验值，导致 `go generate ./cmd/server` 失败；本地仅补回两条校验值，不修改业务依赖图。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok 视频、音频和素材上传、Nano Banana、社区群聊与私聊站长、OpenAI/Grok/Anthropic 多账号渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run --reporter=dot`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 200 个测试文件、1383 项测试。
  - `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q` 与 `bash deploy/test-caddyfile-cache.sh` 通过。
  - 前端仅保留既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 非致命提醒。
  - `git diff --check`、`git diff --cached --check`、未解决冲突检查和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
  - `go mod tidy -diff` 只建议删除 Wire 生成命令实际需要的两条 `google/subcommands` 校验值，因此按可复现生成要求保留。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 aa04f43bc292a96db4e1c9b63cb6f83fccf6108e`；同步前备份分支为 `backup/pre-upstream-sync-20260727-213337`，完整工作区 stash 对象为 `d5351da0a0fd0b573ee462d20b6e3c7f23192b42`。

## 2026-07-30 upstream/main 同步结果

- 同步范围：`59ce11c78..5a6143097`，共 37 个上游提交、170 个文件；上游版本从 `0.1.166` 推进到 `0.1.168`，本地合并提交为 `878ff00d5ef26f43ca7c9aea894f4d20bc9bfc2f`。
- 同步策略：创建 `backup/pre-upstream-sync-20260730-210827`，将当前暂存、未暂存和未跟踪改动完整保存为 stash `c950cbac882d2940d5c69a09261ab7dc4c6531ee`；在干净工作树合并上游后恢复全部本地改动并解决 3 个内容冲突，stash 保留未删除。
- 已同步内容：Passkey 登录、注册与撤销保护，公开模型广场和分组定价展示，用户/API Key 指定字段更新以避免并发覆盖，安全审计配置解密恢复，Kimi K3 与 Claude Sonnet 5 状态支持，以及 OpenAI Live、GPT-5.6、模型映射透传和 Web Search 兼容修复。
- 冲突与兼容处理：
  - 服务器路由同时注册上游公开模型广场与本地社区群聊，保留上游可选 JWT 和面板限流签名。
  - 登录页保留本地视觉布局、登录协议和 OAuth 结构，同时接入上游 Passkey 登录按钮、加载状态和能力开关。
  - Wire 依赖图重新生成，同时包含上游 Passkey/模型广场和本地社区群聊/公开素材 Handler。
  - `go.sum` 接受上游 WebAuthn 所需 `go-tpm-tools` 校验值、清理旧 JWT 校验值，并继续保留 Wire 命令实际需要的两条 `google/subcommands` 校验值。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok 视频、音频和素材上传、Nano Banana、社区群聊与私聊站长、多平台渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run --reporter=dot`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 203 个测试文件、1401 项测试。
  - `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q` 与 `bash deploy/test-caddyfile-cache.sh` 通过。
  - 前端仅保留既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 非致命提醒。
  - `git diff --check`、`git diff --cached --check`、未解决冲突检查和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
  - `go mod tidy -diff` 只建议删除 Wire 生成命令实际需要的两条 `google/subcommands` 校验值，因此按可复现生成要求保留。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 878ff00d5ef26f43ca7c9aea894f4d20bc9bfc2f`；同步前备份分支为 `backup/pre-upstream-sync-20260730-210827`，完整工作区 stash 对象为 `c950cbac882d2940d5c69a09261ab7dc4c6531ee`。

## 2026-07-31 upstream/main 同步结果

- 同步范围：`5a6143097..2980ff385`，共 63 个上游提交、132 个文件；版本文件推进到 `0.1.169`，上游 HEAD 还包含该标签后的 27 个提交，本地合并提交为 `d74880bf2faf0815e83aceed1cb973418ee63d61`。
- 同步策略：创建 `backup/pre-upstream-sync-20260731-220033`，将当前暂存、未暂存和未跟踪改动完整保存为 stash `21daeede04acf323667e0583297a8c6be280112a`；在干净工作树合并上游后恢复全部本地改动并解决 4 个内容冲突，stash 保留未删除。
- 已同步内容：简洁首页预设、订阅额度周期修正、OpenAI Responses namespace/compaction/透传修复、Grok SSE 与池模式冷却修复、上游 URL path 安全校验、图片 data URL 异步卸载修复，以及支付、SMTP、模型定价与账号管理改进。
- 冲突与兼容处理：
  - 网关路由采用上游 Responses 子路径安全护栏，同时保留本地 OpenAI 视频提交/查询/下载、Grok 视频、音频、Nano Banana 和 `/pg/assets` 素材接口。
  - OpenAI APIKey 的 Gemini 原生 Bearer 转发继续支持自定义 `/v1` Base URL；模型/action URL 改走上游安全路径构造器，模型列表路径同时保留能力检查与上游分段校验。
  - OpenAI OAuth 透传采用上游新行为，在 Codex 请求缺少 `instructions` 时补默认指令，并保留对应回归覆盖。
  - 首页保留本地 AI/Agent Three.js 视觉页作为默认页，同时接入上游 `compact_home_enabled` 简洁首页；自定义 HTML/URL 首页仍具有最高优先级，并恢复上游登录态及公共配置初始化。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok 视频、音频和素材上传、Nano Banana、社区群聊与私聊站长、多平台渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run --reporter=dot`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 205 个测试文件、1430 项测试。
  - `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q` 与 `bash deploy/test-caddyfile-cache.sh` 通过。
  - 前端仅保留既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 非致命提醒。
  - `git diff --check`、`git diff --cached --check`、未解决冲突检查和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 d74880bf2faf0815e83aceed1cb973418ee63d61`；同步前备份分支为 `backup/pre-upstream-sync-20260731-220033`，完整工作区 stash 对象为 `21daeede04acf323667e0583297a8c6be280112a`。

## 2026-08-02 upstream/main 同步结果

- 同步范围：`2980ff385..7e2e9ba05`，共 36 个上游提交、190 个文件；版本从 `0.1.169` 推进到 `0.1.170`，本地合并提交为 `6971356ba35c7445b3a0ed9182bf3a86a3b9eee0`。
- 同步策略：创建 `backup/pre-upstream-sync-20260802-212128`，将当前暂存、未暂存和未跟踪改动完整保存为 stash `3f60a9a008464a9d5ee4df2cd87353f11ee2e672`；在干净工作树合并上游后恢复全部本地改动，stash 保留未删除。
- 已同步内容：OpenAI/Grok 分组利润控制、上游计费倍率探测及受控自动回写、全部 APIKey 平台的计费探测、内容审计代理与窄范围阻断、OpenAI SSE 429 重试、Anthropic 流中断用量记录、按筛选结果全选账号、批量删除并发限制，以及模型广场和账号管理体验改进。
- 数据库变化：同步新增 `192_group_profit_control.sql` 和 `193_group_profit_control_auth_cache_invalidation.sql`；本轮只同步迁移文件，没有连接或修改运行中的数据库，后续部署由服务启动迁移流程执行。
- 冲突与兼容处理：
  - README 赞助商列表采用上游最新内容，同时保留未冲突的本地项目说明。
  - OpenAI handler 测试同时保留上游 SSE 429 池内重试，以及本地渠道监控五账号预算、池内重试计一个账号、HTTP 400 换号覆盖。
  - 账号管理页接入上游“按筛选结果全选”工具，并继续使用本地异步 Modal 加载，避免重复组件注册和首屏体积回退。
  - 本地音频、视频和 Nano Banana 入口适配上游账号槽位三态结果；利润终检否决会排除当前账号并重新调度，不会产生空响应。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok 视频、音频和素材上传、Nano Banana、社区群聊与私聊站长、多平台渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - OpenAI 冲突专项的 SSE 429、五账号预算和监控 HTTP 400 换号三项测试通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run --reporter=dot`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 209 个测试文件、1465 项测试。
  - `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q` 与 `bash deploy/test-caddyfile-cache.sh` 通过。
  - 前端仅保留既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 非致命提醒。
  - `git diff --check`、`git diff --cached --check`、未解决冲突检查和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 6971356ba35c7445b3a0ed9182bf3a86a3b9eee0`；同步前备份分支为 `backup/pre-upstream-sync-20260802-212128`，完整工作区 stash 对象为 `3f60a9a008464a9d5ee4df2cd87353f11ee2e672`。

## 2026-08-06 upstream/main 同步结果

- 同步范围：`7e2e9ba05..a19c9f8d8`，共 60 个上游提交、221 个文件；版本从 `0.1.170` 推进到 `0.1.171`，本地合并提交为 `558f2a2efd131fb9eff8d4ac68053bb79f89a64a`。
- 同步策略：创建 `backup/pre-upstream-sync-20260806-234853`，将当前暂存、未暂存和未跟踪改动完整保存为 stash `4f5ec6d2cd490c96bc6f15368419f433ef3bc137`；在干净工作树合并上游后恢复全部本地改动，stash 保留未删除。恢复后逐项核对 219 个已跟踪定制路径和 49 个未跟踪文件，均无遗漏。
- 已同步内容：阿里云验证码 2.0、腾讯天御验证码及 OAuth 注册门禁；Refresh Token 轮换竞态修复；订阅续期串行化；请求取消后停止调度；OpenAI Codex 身份和版本同步；重置额度缓存及账号恢复；计费金额精度、失败用量记录、退款幂等和余额保护；Messages 临时错误故障切换与 WebSocket 终态事件保留。
- 数据库变化：本轮上游没有新增迁移文件，也没有连接或修改运行中的数据库。
- 冲突与兼容处理：
  - OpenAI 网关保留本地流断开排空时间，同时采用上游 Codex `0.146.0` 身份版本及自动版本同步机制。
  - 注册和邮箱验证页接入上游腾讯、阿里验证码配置，继续保留本站 `ChinaAPI` 默认名称及既有页面定制。
  - 公开设置契约同时覆盖上游腾讯验证码字段和本地图片工作区 URL。
  - Chat Completions 故障切换同时保留本地严格解析“外层 400、内层可重试状态”和渠道监控换号逻辑，并接入上游临时不可调度错误策略。
  - Wire 依赖图重新生成，同时包含上游验证码服务以及本地社区群聊、公开素材和媒体 Handler。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok 视频、音频和素材上传、Nano Banana、社区群聊与私聊站长、多平台渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run --reporter=dot`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 215 个测试文件、1522 项测试。
  - 生产构建后的 `go test -tags=unit ./cmd/server ./internal/web -count=1` 通过。
  - `POSTGRES_PASSWORD=test docker compose -f deploy/docker-compose.dev.yml config -q` 与 `bash deploy/test-caddyfile-cache.sh` 通过。
  - 前端仅保留既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 非致命提醒。
  - `git diff --check`、`git diff --cached --check`、未解决冲突检查和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
  - `go mod tidy -diff` 只建议删除 Wire 生成命令实际需要的两条 `google/subcommands` 校验值，因此按可复现生成要求保留。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 558f2a2efd131fb9eff8d4ac68053bb79f89a64a`；同步前备份分支为 `backup/pre-upstream-sync-20260806-234853`，完整工作区 stash 对象为 `4f5ec6d2cd490c96bc6f15368419f433ef3bc137`。

## 2026-08-08 upstream/main 同步结果

- 同步范围：`a19c9f8d8..68d8f122e`，共 43 个上游提交、192 个文件；版本从 `0.1.171` 推进到 `0.1.172`，本地合并提交为 `d78bcd679c7778abe61af04c60f07ef7da808453`。
- 同步策略：创建 `backup/pre-upstream-sync-20260808-004445`，将当前暂存、未暂存和未跟踪改动完整保存为 stash `5d1ec7a216f72a840e93f800c94b57858f6654c9`；在干净工作树合并上游后恢复全部本地改动，stash 保留未删除。恢复清单与 stash 完全一致。
- 已同步内容：OAuth pending exchange 账号接管保护、腾讯验证码区域与 CSP 修复、上游 TCP 显式连接超时、OpenAI 流式首包前故障切换恢复、订阅每日零点额度重置、Grok 视频任务绑定、Responses 工具 schema 清理，以及上游实际响应模型审计。
- 数据库变化：新增 `194_add_usage_log_upstream_response_model.sql` 与 `195_add_usage_log_upstream_model_mismatch_index_notx.sql`，用于记录上游响应模型及建立异常模型审计索引。本轮未连接或修改运行中的数据库；仅在临时 PostgreSQL 16 测试容器中验证迁移。
- 冲突与兼容处理：
  - Gemini Messages 同时保留上游响应模型观测和本地实际图片输出计数，避免模型审计或按图计费任一退化。
  - 管理端用量表同时展示上游请求/响应模型差异与本地视频失败退款标识，测试覆盖两类记录。
  - 恢复本地计费改动后补回上游模型冲突日志所需的 `zap` import，修复合并组合态编译失败。
  - Responses failover 测试夹具接入真实最小计费服务，继续验证客户端断开停止换号和在线客户端正常换号，同时保留缺价请求不放行的计费保护。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok/MiniMax 视频、音频和素材上传、Nano Banana、社区群聊与私聊站长、多平台渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run --reporter=dot`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 217 个测试文件、1542 项测试。
  - 生产构建后的 `go test -tags=unit ./internal/web ./cmd/server -count=1` 通过。
  - 临时 PostgreSQL 16/Redis 测试容器中的迁移并发、幂等、字段与索引集成测试通过；临时 Redis 测试标签已删除。
  - 根目录、部署版与 standalone 三份 Compose 配置在提供必填占位环境变量后均通过 `docker compose config --quiet`。
  - 前端仅保留既有 localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 非致命提醒。
  - `go mod tidy -diff` 仍只建议删除 Wire 生成命令实际需要的两条 `google/subcommands` 校验值，因此按可复现生成要求保留。
  - `git diff --check`、`git diff --cached --check`、严格冲突标记扫描、恢复清单核对和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 d78bcd679c7778abe61af04c60f07ef7da808453`；同步前备份分支为 `backup/pre-upstream-sync-20260808-004445`，完整工作区 stash 对象为 `5d1ec7a216f72a840e93f800c94b57858f6654c9`。

## 2026-08-12 upstream/main 同步结果

- 同步范围：`68d8f122e..1e618dbc2`，共 174 个上游提交、489 个文件；版本文件从 `0.1.172` 推进到 `0.1.173`，并包含 `v0.1.173` 标签后的 55 个提交，本地合并提交为 `32459c19a0dd98ae6408d4bc227fbf99ae59c0a9`。
- 同步策略：创建 `backup/pre-upstream-sync-20260811-231855`，将同步前全部暂存、未暂存和未跟踪改动保存为 stash `704efd34e5144ab0ee60f479738420e4cf404eb3`；在干净工作树完成记录性合并后恢复本地定制，stash 保留未删除。
- 已同步内容：Channel Monitor V2 被动聚合与用户/管理端界面、Grok Voice/音频/Web Search/视频计费及 OAuth 配额链路、OpenAI 调度阈值与容量退避、上游响应模型安全计费、API Key 输入校验、大文件备份分卷，以及 Responses、图片流、池模式鉴权重试和安全审计修复。
- 数据库变化：上游新增 17 个迁移，覆盖 Channel Monitor V2、分组视频模型价格、音频 Voice 价格、搜索价格和非 Grok 视频配置清理。本地已有 `196_openai_video_billing_reconciliation.sql` 与 `197_openai_video_submission_recovery.sql`，上游也新增了不同文件名的 `196/197`；迁移器按完整文件名判重，因此保留所有已发布文件且不重命名。本轮未连接或修改运行中的数据库，仅在临时 PostgreSQL 18.1/Redis 8.4 容器中验证。
- 冲突与兼容处理：
  - 公开设置同时保留上游注册邮箱域名额度开关和本地图片工作区 URL；Handler/Wire 同时注入 Channel Monitor V2、社区群聊与公开素材服务。
  - `/videos` 按账号平台分发到本地 OpenAI 视频或上游 Grok 视频，并继续保留 MiniMax-H3、SD2.0、Nano Banana、音频和 `/pg/assets` 路由。
  - Gemini 最终计费采用上游实际响应图片数，同时保留本地流中断增量图片计数，避免少扣或重复计算。
  - 计费链路合入上游响应模型计费，同时保留本地缺价错误语义、强制用量任务和视频价格快照；响应模型重算失败时回落基线成本。
  - MiniMax-H3 与 SD2.0 分辨率按秒计费继续使用本地计费模型，并适配上游新增的视频定价签名。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok/MiniMax 视频、音频和素材上传、Nano Banana、社区群聊与私聊站长、多平台渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test -tags=unit ./... -count=1` 从 `backend/` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm test:run --reporter=dot`、`pnpm lint:check`、`pnpm typecheck` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 230 个测试文件、1603 项测试。
  - 生产构建后的 `go test -tags=unit ./internal/web ./cmd/server -count=1` 通过。
  - 临时 PostgreSQL 18.1/Redis 8.4 中的迁移并发、幂等、字段和索引集成测试通过。
  - 根目录、deploy、开发版和 standalone 四份 Compose 配置通过 `docker compose config --quiet`；Caddy 缓存/SSE/压缩规则测试通过。
  - `git diff --check`、`git diff --cached --check`、严格冲突标记扫描、恢复清单核对和 `git merge-base --is-ancestor upstream/main HEAD` 通过。恢复清单中两个测试调整已由上游等价实现或新默认场景取代，未丢失本地业务功能。
  - `go mod tidy -diff` 仅建议删除历史冗余校验项；其中 `google/subcommands` 仍用于可复现 Wire 生成，本轮不扩大范围清理。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 32459c19a0dd98ae6408d4bc227fbf99ae59c0a9`；同步前备份分支为 `backup/pre-upstream-sync-20260811-231855`，完整工作区 stash 对象为 `704efd34e5144ab0ee60f479738420e4cf404eb3`。部署验收前不要删除该 stash。

## 2026-08-13 upstream/main 同步结果

- 同步范围：`1e618dbc2..5935e674a`，共 31 个上游提交、37 个文件；版本文件从 `0.1.173` 推进到 `0.1.175`，本地合并提交为 `2520fd7b93ff2986cc0e3a8d8582a19051ce0df6`。
- 同步策略：创建 `backup/pre-upstream-sync-20260813-002446`，将同步前全部暂存、未暂存和未跟踪改动保存为 stash `321c931f8a818091237c87239451aa3b40943eb0`；在干净工作树完成合并后使用 `--index` 恢复原有暂存状态，stash 保留未删除。
- 已同步内容：Codex OAuth 设备指纹收敛、OpenAI Responses 可见输出 TTFT 与嵌套 usage 解析、WebSocket V2 终止事件处理、安全审计范围修正、HTML 403 账号处罚修正、Gemini 工具 schema 兼容、账号成本 service tier 计价以及账号表单和运营监控展示改进。
- 数据库变化：本批次没有新增或修改数据库迁移，不需要额外执行结构兼容处理。
- 兼容处理：Git 自动合并全部重叠文件，没有文本冲突；语义检查确认上游鉴权、WebSocket、限流和审计逻辑已接入现有网关，同时保留本地 Gemini 实际图片计费、严格缺价保护、MiniMax-H3 504 恢复与延迟结算、OpenAI 视频/音频、公开素材、社区聊天和定制页面。Wire 重新生成后保持本地视频补偿器和社区/公开素材依赖完整。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test ./...` 从 `backend/` 通过。
  - MiniMax 恢复、Codex 指纹、限流、安全审计和 Gemini 相关的 `go test -race` 定向测试通过。
  - `pnpm lint:check`、`pnpm typecheck`、`pnpm test:run` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 231 个测试文件、1606 项测试。
  - 根目录、deploy、开发版和 standalone 四份 Compose 配置通过 `docker compose config --quiet`；Caddy 缓存/SSE/压缩规则测试通过。
  - `git diff --check`、`git diff --cached --check`、严格冲突标记扫描、恢复清单核对和 `git merge-base --is-ancestor upstream/main HEAD` 通过。同步前后本地改动清单均为 327 个路径，没有丢失文件。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 2520fd7b93ff2986cc0e3a8d8582a19051ce0df6`；同步前备份分支为 `backup/pre-upstream-sync-20260813-002446`，完整工作区 stash 对象为 `321c931f8a818091237c87239451aa3b40943eb0`。部署验收前不要删除该 stash。

## Grok 4.6 推理强度兼容

- Grok 渠道账号在 `grok-4.6` 上保留官方 `xhigh` 推理强度：Chat Completions 使用顶层 `reasoning_effort: "xhigh"`，Responses 使用 `reasoning.effort: "xhigh"`。
- 账号模型映射后的上游模型必须为 `grok-4.6`。该模型内置独立的官方默认 Token 价格：输入 `$2/MTok`、缓存输入 `$0.50/MTok`、输出 `$6/MTok`；其中缓存输入价不同于 Grok 4.5 的 `$0.30/MTok`。
- 渠道中为 `grok-4.6` 手动配置的输入、输出或缓存价格继续优先于默认价格；分组倍率和现有费用结算公式保持不变。
- 其他 Grok 模型保持原有兼容规则：`xhigh`、`extra-high`、`max` 和 `ultra` 继续归一化为 `high`；不支持推理强度的模型继续在转发前移除该参数。

## 2026-08-13 upstream/main 同步结果（v0.1.176）

- 同步范围：`5935e674a..fbfdcef81`，共 26 个上游提交、98 个文件；版本文件从 `0.1.175` 推进到 `0.1.176`，并包含 `v0.1.176` 标签后的 5 个提交，本地合并提交为 `744c03c7e861d65675472cc5d286f9d07e8ccbe5`。
- 同步策略：创建 `backup/pre-upstream-sync-20260813-222818`，将同步前全部暂存、未暂存和未跟踪改动保存为 stash `5440dd51f71b36df26c7826a54a9a55b4558cc8a`；在干净工作树完成合并后以三方方式恢复全部本地内容，stash 保留未删除。恢复后逐一校验未跟踪文件对象哈希，均与同步前一致。
- 已同步内容：分组逐模型定价和长上下文阶梯开关、Grok JWT 订阅档位识别、Grok 4.6 官方目录与请求支持、独立 `/x_search` 路由及计费、Chat/Responses `x_search` 参数和来源提取、定时备份多实例 Leader 锁、渠道缓存失效修复，以及 OpenAI Responses 能力探测误判修复。
- 数据库变化：新增 `221_group_model_pricing.sql`，为分组增加逐模型定价和长上下文定价开关。本轮没有连接或修改运行中的数据库；迁移文件和 Ent schema 已通过仓库测试，部署时由服务启动迁移流程执行。
- 冲突与兼容处理：
  - Grok 4.6 采用上游 `grok-4.6-latest` 别名、JWT 档位和模型级配额上下文，同时继续保留本站 Grok 4.6 `xhigh` 透传和独立官方默认定价。
  - 包装成外层 HTTP 400 的内部可重试状态仍使用标准化后的状态码参与同账号/换账号重试；上游的模型级 Grok 限流标记同时生效。
  - Grok 视频内容下载继续直接转发二进制流，避免被 JSON 响应读取路径截断；成功响应仍更新上游模型对应的用量快照。
  - MiniMax-H3 504 恢复和媒体任务继续优先使用创建阶段价格快照，避免恢复期间因分组价格变化而重新计费；上游 OpenAI/Grok 长上下文门控规则同时保留。
  - 渠道 `video` 计费继续由本站按实际分辨率和秒数处理；MiniMax-H3、SD2.0 与普通视频模板不会被上游通用图片/按次界面覆盖。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok/MiniMax 视频、音频和素材上传、Nano Banana、社区群聊与私聊站长、多平台渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test ./...` 从 `backend/` 通过；服务、Handler、路由、迁移和仓储专项均通过。
  - 备份 Leader 锁、Grok 4.6/长上下文计费、媒体价格快照和视频计费定向 `go test -race` 通过。
  - `pnpm lint:check`、`pnpm typecheck`、`pnpm test:run` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 231 个测试文件、1615 项测试。
  - 生产前端嵌入后的 `go test ./internal/web ./cmd/server -count=1` 与根目录 `docker compose config --quiet` 通过。
  - `git diff --check`、`git diff --cached --check`、严格冲突标记扫描、未跟踪文件对象哈希核对和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 744c03c7e861d65675472cc5d286f9d07e8ccbe5`；同步前备份分支为 `backup/pre-upstream-sync-20260813-222818`，完整工作区 stash 对象为 `5440dd51f71b36df26c7826a54a9a55b4558cc8a`。部署验收前不要删除该 stash。

## 2026-08-16 upstream/main 同步结果（v0.1.177）

- 同步范围：`fbfdcef81..baeac1f3d`，共 13 个上游提交、68 个文件；版本从 `0.1.176` 推进到 `0.1.177`，本地合并提交为 `d860238bcf136cd32e73d114f2a5f26cb884b83a`。
- 同步策略：创建 `backup/pre-upstream-sync-20260816-170750`，将同步前全部暂存、未暂存和未跟踪改动保存为 stash `99e4d08261806533cd36e1cbc2a62239331e5cdb`；在干净工作树完成合并后恢复原有暂存边界和全部本地文件，stash 保留未删除。恢复后的暂存、未暂存路径集合与同步前完全一致，74 个未跟踪路径及对象哈希也完全一致。
- 已同步内容：分组用量每日汇总及配置时区边界、与旧 `/responses/compact` 分离的原生远程压缩 v2、`x-codex-turn-state` 跨请求转发及跨账号回显保护、会话级 Codex beta features，以及改为显式启用并覆盖透传路径的 Codex 指纹收敛。
- 工具链变化：Go 版本推进到 `1.26.6`，前端锁文件将 `nanoid` 从 `3.3.17` 更新到 `3.3.18`；CI、发布和安全扫描工作流同步使用新的 Go 版本。
- 数据库变化：新增 `222_group_usage_daily_rollups.sql` 和 `223_group_usage_rollup_timezone.sql`，用于维护分组每日用量汇总并按配置时区重建日期边界。本轮没有连接或修改运行中的数据库；由于未配置 `TEST_DATABASE_URL`，真实 PostgreSQL 触发器集成测试未执行，部署时仍需由服务迁移流程应用并观察。
- 冲突与兼容处理：
  - OpenAI 账号测试保留上游原生 `/responses` 压缩行为，同时保留本地测试不反复修改 Gin 全局模式的约束。
  - OpenAI 网关同时保留上游 Codex turn-state 跟踪和本地视频补偿 worker 状态，媒体恢复链路没有被覆盖。
  - 语义复核修正了 Grok 4.6 定价对象误共享：Grok 4.6 继续使用官方缓存输入价 `$0.50/MTok`，不与 Grok 4.5 的 `$0.30/MTok` 共用定价对象；输入和输出价仍为 `$2/MTok` 与 `$6/MTok`。
- 本地功能保持不变：OpenAI Gemini 原生、Grok APIKey 自定义上游、SD2.0/Grok/MiniMax 视频、媒体余额补偿、音频和素材上传、Nano Banana、社区群聊与私聊站长、多平台渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 通过且没有生成额外差异；`go test ./...` 与 `go test -tags=unit ./...` 从 `backend/` 全量通过。
  - 压缩、turn-state、计费精度、分组汇总和迁移定向测试通过；服务和仓储相关定向 `go test -race` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm lint:check`、`pnpm typecheck`、`pnpm test:run` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 236 个测试文件、1636 项测试。
  - 生产前端嵌入后的 `go test ./internal/web ./cmd/server -count=1` 通过；根目录和全部 deploy Compose 文件在提供必填占位环境变量后均可成功渲染。
  - `git diff --check`、`git diff --cached --check`、严格冲突标记扫描、上游祖先关系、暂存/未暂存路径集合和未跟踪文件对象哈希核对通过。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 d860238bcf136cd32e73d114f2a5f26cb884b83a`；同步前备份分支为 `backup/pre-upstream-sync-20260816-170750`，完整工作区 stash 对象为 `99e4d08261806533cd36e1cbc2a62239331e5cdb`。部署验收前不要删除该 stash。

## 2026-08-18 upstream/main 同步结果（v0.1.178）

- 同步范围：`baeac1f3d..49504adc9`，共 107 个上游提交、301 个文件；版本从 `0.1.177` 推进到 `0.1.178`，本地合并提交为 `297b246fda9064648bee6772d789779314d14c52`。
- 同步策略：创建 `backup/pre-upstream-sync-20260818-223733`，将同步前全部暂存、未暂存和未跟踪改动保存为 stash `76a1544435ca56b496bbcec8b203352300bd715b`；在干净工作树合并上游后恢复本地内容并重建原暂存边界，stash 保留未删除。同步前的 264 个暂存路径、23 个未暂存路径、76 个未跟踪路径及全部未跟踪文件对象均已核对无遗漏；同步兼容测试修正另增加 1 个未暂存路径。
- 已同步内容：Kimi、智谱和 DeepSeek 国产供应商的一等分组、调度、协议、余额与配额支持；渠道模型分时倍率定价；渠道监控配额模式及 8 平台配额视图；OpenAI Team 联动熔断、Codex 身份收敛和指纹种子回填；OpenAI 客户端工具、Anthropic 原生转换、Gemini 工具与错误策略修复；Grok 用量聚合、账号批量设置、远程 Select 搜索及多项运维界面修复。
- 数据库变化：新增 `224_user_platform_quotas_add_cn_providers.sql`、`225_backfill_codex_fingerprint_seed.sql`、`225_channel_model_time_pricing.sql` 和 `226_channel_monitor_quota_mode.sql`。两个 `225` 文件用途和完整文件名不同，迁移器按完整文件名执行，因此均保留且不重命名。本轮没有连接或修改运行中的数据库。
- 冲突与兼容处理：
  - 渠道请求和领域模型同时保留上游 `TimePricing`、分时时段结构以及本站 `video` 计费模式和视频价格字段语义。
  - Anthropic/OpenAI 用量计费同时传递上游请求级 `pricingAt`，并保留本站缺价时返回错误、禁止静默记零费用的保护。
  - OpenAI 视频恢复继续优先使用 `BillingCostSnapshot`，没有快照时才按带 `pricingAt` 的新签名重新计价。
  - 管理端计费模式切换继续套用 MiniMax-H3、SD2.0 和普通视频模板，同时清空上游新增的分时区间，避免旧时段配置串到新计费模式。
  - 严格计费预检使用零时间检查“是否存在价格”，不把任意真实时段倍率混入预检；Grok API Key 测试同步适配上游将占位符改为 `switch` 返回值的源码结构。
- 本地功能保持不变：OpenAI Gemini 原生、Grok API Key 自定义上游、SD2.0/Grok/MiniMax 视频、视频价格快照与失败退款补偿、音频和素材上传、Nano Banana、社区群聊与私聊站长、多平台渠道监控、首页/登录/模型问答和接口文档定制。
- 验证结果：
  - `go generate ./cmd/server` 通过且没有生成额外差异；`go test ./internal/handler/admin ./internal/repository ./internal/service -count=1` 和 `go test ./... -count=1` 从 `backend/` 全部通过。
  - 分时定价、计费快照、视频失败退款、补偿 worker 和严格缺价校验的定向 `go test -race` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm lint:check`、`pnpm typecheck`、`pnpm exec vitest run --reporter=dot` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 242 个测试文件、1713 项测试。
  - 生产前端嵌入后的 `go test ./internal/web ./internal/server ./cmd/server -count=1` 通过；根目录及全部 deploy Compose 文件在提供必填占位环境变量后均可成功渲染。
  - 前端仅保留既有 TypeScript 支持范围、localStorage、Vue 测试环境、Browserslist 数据过期和 Vite 大 chunk 非致命提示。
  - `git diff --check`、`git diff --cached --check`、严格冲突标记扫描、上游祖先关系、原暂存/未暂存路径集合和未跟踪文件对象哈希核对通过。
- 回退点：先保存当前工作区，再执行 `git revert -m 1 297b246fda9064648bee6772d789779314d14c52`；同步前备份分支为 `backup/pre-upstream-sync-20260818-223733`，完整工作区 stash 对象为 `76a1544435ca56b496bbcec8b203352300bd715b`。部署验收前不要删除该 stash。

## 2026-08-20 upstream/main 同步结果（v0.1.179）

- 同步范围：`49504adc9..2bc139ab5`，共 79 个上游提交、214 个文件；版本从 `0.1.178` 推进到 `0.1.179`，本地合并提交为 `705813e63831ce8814238435670fdd27c14443bb`。
- 同步策略：创建 `backup/pre-upstream-sync-20260820-210841`，将同步前工作区完整保存为 stash `812da27c31c476a65407d8aa1ff2a3691791fac8`；在干净基线上合并上游后恢复本地内容，stash 和备份分支均保留未删除。未修改运行中的数据库、余额、任务或生产服务。
- 已同步内容：用量日志有效模型索引、Composite 国内供应商路由、渠道 service tier/context 倍率定价、渠道监控探测、OpenAI failover/容量恢复、Responses/WebSocket 行为、Grok 4.6 `xhigh` 与工具/图片处理，以及对应管理端表单和多语言标签。
- 数据库变化：仅引入上游迁移 `226_add_usage_log_effective_model_indexes_notx.sql`、`227_composite_routes_add_cn_providers.sql`、`228_channel_pricing_multipliers.sql`；本轮未连接或修改运行中的数据库，部署时由服务迁移流程执行。
- 冲突与兼容处理：保留上游端点输入 token、容量恢复和 Composite 逻辑，同时恢复本地音频端点、视频/媒体路由、渠道监控探测和媒体计费；合并渠道 Fast/Flex 与区间倍率字段；Grok 4.6 保持官方 `xhigh`，其他模型仍降级为 `high`；补齐视频缓存测试 stub 的 reasoning 内容接口和前端视频区间倍率默认值。

### 验证结果

- `go generate ./cmd/server` 通过。
- `go test ./... -count=1` 从 `backend/` 全量通过；MiniMax/Firefly 恢复、视频计费、渠道倍率和 OpenAI failover 定向 `go test -race` 通过。
- `pnpm install --frozen-lockfile`、`pnpm lint:check`、`pnpm typecheck`、`pnpm exec vitest run --reporter=dot` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 251 个测试文件、1762 项测试。
- 根目录及全部 deploy Compose 文件在提供必填占位环境变量后均成功渲染；`deploy/test-caddyfile-cache.sh` 通过。
- `git diff --check`、`git diff --cached --check`、严格冲突标记扫描和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
- 回滚点：保留备份分支 `backup/pre-upstream-sync-20260820-210841` 和 stash `812da27c31c476a65407d8aa1ff2a3691791fac8`；需要撤销同步时先保存当前工作区，再执行 `git revert -m 1 705813e63831ce8814238435670fdd27c14443bb`，不要删除用户本地媒体/社区文件。

## 2026-08-24 upstream/main 同步结果（v0.1.180）

- 同步范围：`2bc139ab5..03e8ab413`，共 176 个上游提交、479 个上游变更文件；版本从 `0.1.179` 推进到 `0.1.180`，本地合并提交为 `edf3986d837e8986f2cf35dfb1c1baa0a63bda82`。
- 同步策略：创建 `backup/pre-upstream-sync-20260824-211411`，将同步前全部暂存、未暂存和 77 个未跟踪对象保存为 stash `33437ce09706bcf69e8f71b44b82c267b60afb4e`；在干净工作区合并上游后，以三方方式恢复本地内容并保留 stash 未删除。
- 已同步内容：OAuth 出站传输插件系统、OpenAI 重置额度自动使用、Fast service tier、模型列表读取上限、Codex 账号/Guardian 粘连、Responses/Chat/WS 工具与续链兼容、Grok 容量和同账号重试、渠道分时与长上下文阶梯展示，以及对应运维、账号优先级和前端管理能力。
- 数据库变化：新增上游迁移 `229_plugins.sql` 和 `230_plugin_artifacts.sql`，用于插件元数据与制品；本轮未连接或修改运行中的数据库，部署时由服务迁移流程执行。
- 冲突与兼容处理：
  - Wire 同时保留插件管理器、额度自动重置、社区聊天、公开资源和 OpenAI 视频失败补偿器的启动/停止依赖。
  - OpenAI/Grok 错误路径同时使用包装状态码归一化和上游新增的同账号重试、容量隔离及 API-key 健康熔断语义。
  - 本地音频、Nano Banana 和视频处理器适配上游调度反馈接口的账号对象参数；WebSocket 多轮计费补充账号模型映射候选，避免渠道别名在缺价失败关闭后丢失 usage log。
  - OAuth 账号继续通过 Responses bridge 生成图片；SetupToken 与 API Key 按上游 v0.1.180 能力参与原生 Images 调度。
  - Compose 同时保留上游网关/调度变量和本站 `IMAGE_WORKSPACE_URL`；首页保留 3D 场景并接入上游模型广场入口。
- 工具链变化：Go 推进到 `1.27.0`；恢复 Wire 0.7.0 所需的 `github.com/google/subcommands v1.2.0` 校验项。前端安全更新将 `dompurify` 推进到 `3.4.14`。
- 验证结果：
  - `go generate ./cmd/server`、`go test ./... -count=1` 从 `backend/` 全量通过；冲突相关 handler/service 定向 `go test -race` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm lint:check`、`pnpm typecheck`、`pnpm exec vitest run --reporter=dot` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 258 个测试文件、1840 项测试。
  - 根目录及四份 deploy Compose 文件在提供必填占位变量后均成功渲染；`deploy/test-caddyfile-cache.sh` 通过。
  - `git diff --check`、`git diff --cached --check`、严格冲突标记扫描、77 个未跟踪对象清单指纹和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
- 回滚点：先保存当前工作区，再执行 `git revert -m 1 edf3986d837e8986f2cf35dfb1c1baa0a63bda82`；同步前备份分支为 `backup/pre-upstream-sync-20260824-211411`，完整工作区 stash 为 `33437ce09706bcf69e8f71b44b82c267b60afb4e`。部署验收前不要删除该 stash。

## 2026-08-24 upstream/main 追加同步结果（v0.1.181）

- 同步范围：`03e8ab413..e2d9b823f`，共 9 个上游提交、16 个文件；版本从 `0.1.180` 推进到 `0.1.181`，本地合并提交为 `4ba2feebcad4bffada5257befa50c6ff00a3b484`。
- 同步策略：创建 `backup/pre-upstream-sync-20260824-224151`，将当前 283 个暂存路径、1 个未暂存路径和 77 个未跟踪对象保存为 stash `bc3d458f2a9f1e69a355fe4102a944dba4568fe1`；上游合并及 `stash apply --index` 均无内容冲突，原索引边界完整恢复，stash 保留未删除。
- 已同步内容：Grok CLI 端点统一使用官方 CLI User-Agent；Gemini 兼容层递归清除不支持的工具 schema 字段并规范 enum；Responses Lite 在 `additional_tools` 可用时保留 `parallel_tool_calls`；rejected-field 重试按完整 item 类型清除输入 `status`。
- 数据库和依赖：本批没有新增数据库迁移、Go 依赖或前端依赖，不需要额外结构与依赖处理。
- 兼容检查：自动合并保留了本站 Grok 包装状态码、同账号重试、渠道监控、媒体计费和 OpenAI WebSocket 多轮 usage 逻辑；本轮无需额外生产代码补丁。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test ./... -count=1` 从 `backend/` 全量通过。
  - Grok CLI 身份、Gemini tool schema、Responses Lite `parallel_tool_calls` 和 rejected status 定向 `go test -race` 通过。
  - `pnpm lint:check`、`pnpm typecheck`、`pnpm exec vitest run --reporter=dot` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 258 个测试文件、1840 项测试。
  - 根目录及全部 deploy Compose 文件成功渲染，`deploy/test-caddyfile-cache.sh` 通过；空白、冲突标记、上游祖先关系和工作区清单复核通过。
- 回滚点：先保存当前工作区，再执行 `git revert -m 1 4ba2feebcad4bffada5257befa50c6ff00a3b484`；同步前备份分支为 `backup/pre-upstream-sync-20260824-224151`，完整工作区 stash 为 `bc3d458f2a9f1e69a355fe4102a944dba4568fe1`。

## 2026-08-25 upstream/main 同步结果（v0.1.182）

- 同步范围：`e2d9b823f..4ff136cfd`，共 26 个上游提交、50 个文件；版本从 `0.1.181` 推进到 `0.1.182`，本地合并提交为 `6911a9784e692abc38649c70a3f7d06c2c743f59`。
- 同步策略：创建 `backup/pre-upstream-sync-20260825-212816`，将同步前 283 个暂存路径、1 个未暂存路径和 77 个未跟踪对象保存为 stash `0d48d7fed3f8af6ab438f3326656396e0a67322d`；上游合并及 `stash apply --index` 均无内容冲突，原索引边界完整恢复，stash 保留未删除。
- 已同步内容：OpenAI OAuth 账号在 7 天额度耗尽 429 时暂停调度、OpenCode Go 重置时长识别、Responses Lite 工具并行约束和数值精度、OAuth 图片提示词原样透传、Composite/Kimi Code K3 路由、渠道监控 Composite 平台聚合、Anthropic 缓存明细防重复计费、Antigravity Sonnet 路由和支付完成后的余额刷新。
- 数据库和依赖：本批没有新增数据库迁移、Go 依赖或前端依赖，不需要额外结构与依赖处理。
- 兼容检查：自动合并保留了本站严格缺价保护、渠道计费、OpenAI/Grok 包装状态码、多账号重试、视频/音频/Nano Banana、WebSocket 多轮 usage、社区聊天及首页定制；本轮无需额外生产代码补丁。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test ./... -count=1` 从 `backend/` 全量通过。
  - OAuth 429、缓存计费、Responses Lite、Composite/Kimi、Antigravity 和 OAuth 图片提示词定向 `go test -race` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm lint:check`、`pnpm typecheck`、`pnpm exec vitest run --reporter=dot` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 258 个测试文件、1841 项测试。
  - 根目录及全部 deploy Compose 文件成功渲染，`deploy/test-caddyfile-cache.sh` 通过；空白、冲突标记、上游祖先关系和工作区清单复核通过。
- 回滚点：先保存当前工作区，再执行 `git revert -m 1 6911a9784e692abc38649c70a3f7d06c2c743f59`；同步前备份分支为 `backup/pre-upstream-sync-20260825-212816`，完整工作区 stash 为 `0d48d7fed3f8af6ab438f3326656396e0a67322d`。

## 2026-08-25 upstream/main 追加同步结果（v0.1.183）

- 同步范围：`4ff136cfd..7634e3c23`，共 13 个上游提交、18 个文件；版本从 `0.1.182` 推进到 `0.1.183`，本地合并提交为 `978d9e24c7fdda3669f445cb7358fe8c6b3b25d7`。
- 同步策略：创建 `backup/pre-upstream-sync-20260825-215813`，将同步前 283 个暂存路径、1 个未暂存路径和 77 个未跟踪对象保存为 stash `7f0236e36ae8c05874212f405644198b1efb71ad`；上游合并及 `stash apply --index` 均无内容冲突，原索引边界完整恢复，stash 保留未删除。
- 已同步内容：Codex `session-id` 请求头参与会话粘连、容量溢出时保留 sticky 绑定、Responses 自定义工具调用 ID 类型恢复、Kimi 并发 403 临时冷却、Antigravity 兼容 token 上限，以及邮箱换绑的别名与并发保护。
- 数据库和依赖：本批没有新增数据库迁移、Go 依赖、前端依赖或前端源码变更，不需要额外结构与依赖处理。
- 兼容检查：自动合并保留了本站 OpenAI 包装状态码、多账号/媒体调度、严格计费、WebSocket 多轮 usage、视频/音频/Nano Banana、社区聊天和定制界面；本轮无需额外生产代码补丁。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test ./... -count=1` 从 `backend/` 全量通过。
  - 工具调用 ID、邮箱别名、Kimi 403、Antigravity token、Codex session-id 和 sticky 容量溢出定向 `go test -race` 通过。
  - `pnpm install --frozen-lockfile`、`pnpm lint:check`、`pnpm typecheck`、`pnpm exec vitest run --reporter=dot` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 258 个测试文件、1841 项测试。
  - 根目录及全部 deploy Compose 文件成功渲染，`deploy/test-caddyfile-cache.sh` 通过；空白、冲突标记、上游祖先关系和工作区清单复核通过。
- 回滚点：先保存当前工作区，再执行 `git revert -m 1 978d9e24c7fdda3669f445cb7358fe8c6b3b25d7`；同步前备份分支为 `backup/pre-upstream-sync-20260825-215813`，完整工作区 stash 为 `7f0236e36ae8c05874212f405644198b1efb71ad`。

## 2026-08-31 upstream/main 同步结果（v0.1.184）

- 同步范围：`7634e3c23..200602b41`，共 173 个上游提交、343 个上游变更文件；版本从 `0.1.183` 推进到 `0.1.184`。首批合并提交为 `58b0d7449b31a5c5d0ae34adb7ad13b7552b7fbf`，验证期间新增的 2 个上游提交通过最终合并提交 `f0bcb5948` 纳入。
- 同步策略：将同步前全部暂存、未暂存和未跟踪改动保存为 stash `7c3bdfa62f53d01d9896932ee78515dc50fd43ae`（`pre-upstream-sync-20260831`）；在干净工作区合并上游后恢复本地改动，人工收敛 10 个内容冲突。最终增量同步前又保存完整工作区为 stash `b7401f38fa399ce2396f865f5fcd371024a6ff31`（`pre-upstream-sync-20260831-late`）；两个 stash 均保留未删除。
- 已同步内容：原生 compaction 与请求推理等级审计、用户可见分组访问限制、渠道监控用户权限收敛、OpenAI Responses/WebSocket 会话和 failover 修复、Grok cache identity、Anthropic 工具与流生命周期修复、国产平台用量查询、账号额度冷却、用量窗口与管理端展示修复，以及 Codex 自定义 OpenAI-compatible 上游的已知图片输入能力保留。
- 数据库变化：新增上游迁移 `231_add_usage_log_native_compaction_v2.sql`、`231_add_usage_log_requested_reasoning_effort.sql` 和 `231_user_restrict_public_groups.sql`。本站已有 `231_newapi_balance_access_token.sql`；四个迁移文件名和用途均不同，迁移器按完整文件名执行，因此保留且不重命名。本轮未连接或修改运行中的数据库。
- 冲突与兼容处理：Wire 同时保留新版渠道监控 API Key 注入和本站社区聊天；用量日志同时保留原生 compaction、请求推理等级及 MiniMax 视频输入时长/成本字段；Grok 同时保留 metadata 会话缓存身份和 Grok 4.6 `xhigh`；OpenAI 用量继续使用凭证归属账号计费并保留媒体预冻结结算；前端测试同时覆盖上游模型元数据、推理/响应模型审计和本站图片 URL 转换、视频退款与 MiniMax 视频明细。
- 同步兼容修正：MiniMax 专用模型列表适配上游新增的原始响应体返回值；4 个旧测试构造调用补齐视频任务绑定仓储参数；用量插入表的带标签单测按合并后的 64 个字段更新断言。
- 验证结果：首批合并后 `go generate ./cmd/server`、`go test ./... -count=1`、用量字段顺序的 `go test -tags=unit` 定向测试、`pnpm typecheck`、`pnpm lint:check`、265 个 Vitest 文件/1906 项测试和 `pnpm build` 全部通过；最终 2 个提交的 Codex 图片输入 service/handler 定向测试也通过。前端仅报告既有 Browserslist 数据过期、大 chunk 和测试环境提示。
- 完整性检查：无未解决 Git 冲突，严格冲突标记扫描、`git diff --check`、`git diff --cached --check` 及 `upstream/main` 祖先关系通过。
- 回滚点：先保存当前工作区，再依次执行 `git revert -m 1 f0bcb5948` 和 `git revert -m 1 58b0d7449b31a5c5d0ae34adb7ad13b7552b7fbf`；同步前基线为 `978d9e24c7fdda3669f445cb7358fe8c6b3b25d7`，原始完整工作区 stash 为 `7c3bdfa62f53d01d9896932ee78515dc50fd43ae`，最终增量前快照为 `b7401f38fa399ce2396f865f5fcd371024a6ff31`。部署验收前不要删除这两个 stash。

## 2026-09-01 upstream/main 追加同步结果（v0.1.185）

- 同步范围：`200602b41..0d27f45ea`，共 30 个上游提交、70 个文件；版本从 `0.1.184` 推进到 `0.1.185`，本地合并提交为 `529126406`。
- 同步策略：将同步前全部暂存、未暂存和未跟踪改动保存为 stash `077e946bf1f193c47de4111d9dd962f93f1471e4`（`pre-upstream-sync-20260901`）；在干净工作区合并上游后恢复本地内容，手工收敛 2 个内容冲突，stash 保留未删除。
- 已同步内容：Anthropic fallback 清理、Kimi 原生 Responses、定时自动化 bootstrap、OpenAI 容量与 WebSocket 连接恢复、价格目录驱动的长上下文计费及对应前端账号配置更新。
- 数据库和依赖：本批没有新增迁移；远程版本与前端依赖锁文件已同步，本轮未连接或修改运行中的数据库、余额、任务或服务。
- 冲突与兼容处理：计费路径采用上游价格目录驱动接口，未重新引入已被上游移除的旧入口级 `LegacyLongContext` 参数；保留本站媒体余额预冻结/结算、用量原子落库、视频/音频/Nano Banana、社区聊天和 NewAPI 余额访问等本地定制。OpenAI API Key 测试保留上游“缺少 instructions 时不合成”的行为。
- 验证结果：
  - `go test ./... -count=1` 从 `backend/` 全量通过。
  - `pnpm typecheck` 与 `pnpm lint:check` 从 `frontend/` 通过。
  - `gofmt -d`、`git diff --check`、`git diff --cached --check`、未解决索引检查、严格冲突标记扫描和 `git merge-base --is-ancestor upstream/main HEAD` 通过。
- 回滚点：先保存当前工作区，再执行 `git revert -m 1 529126406`；同步前基线为 `f0bcb5948`，完整工作区 stash 为 `077e946bf1f193c47de4111d9dd962f93f1471e4`。部署验收前不要删除该 stash。

## 2026-09-03 upstream/main 同步结果（v0.2.0）

- 同步范围：`0d27f45ea..b1748c4ea`，共 58 个上游提交、258 个变更文件；版本从 `0.1.185` 推进到 `0.2.0`，本地合并提交为 `00f0f5e9c`。
- 同步策略：将同步前全部暂存、未暂存和未跟踪改动保存为 stash `52687120d5b2f652ef7bd426f438c7eb36cfb269`（`pre-upstream-sync-20260903`），在干净基线上合并远程后恢复本地内容；stash 保留未删除，并建立同步前备份分支 `backup/pre-upstream-sync-20260903`。
- 已同步内容：分组 OpenAI Fast 免费策略、推理等级上限与模型范围、缓存写入 1 小时定价、Fable 5.1、Anthropic/Gemini/OpenAI 兼容层修复、WebSocket 终态处理、上游错误归因与运维上下文、账号调度和管理端分组/渠道界面更新。
- 数据库变化：新增上游迁移 `232_channel_cache_write_1h_pricing.sql`、`232_group_force_openai_fast.sql`、`232_group_reasoning_effort_over_limit.sql` 和 `233_group_free_openai_fast.sql`，以及对应迁移测试；本轮未连接或修改运行中的数据库、余额、任务或服务。
- 冲突与兼容处理：渠道管理请求同时保留上游 `cache_write_1h_price` 字段和本站 `video` 计费模式；默认定价接口继续返回 5 分钟/1 小时缓存写入价格；Wire 重新生成后继续注入本站社区聊天、NewAPI 余额访问、视频任务绑定与失败补偿、公开资源等服务。
- 本地功能保持不变：OpenAI 图片 URL 转换、Gemini/Nano Banana、Grok 和 MiniMax/Firefly 视频、音频、媒体余额预冻结与结算、视频失败退款补偿、社区群聊与私聊站长、多平台渠道监控、NewAPI 余额查询及定制接口文档。
- 验证结果：
  - `go generate ./cmd/server` 与 `go test ./... -count=1` 从 `backend/` 全量通过。
  - `pnpm typecheck`、`pnpm lint:check`、`pnpm exec vitest run --reporter=dot` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 267 个测试文件、1927 项测试。
  - `git diff --check`、`git diff --cached --check`、无未解决索引、严格冲突标记扫描以及 `git merge-base --is-ancestor upstream/main HEAD` 通过。
- 回滚点：先保存当前工作区，再执行 `git revert -m 1 00f0f5e9c`；同步前备份分支为 `backup/pre-upstream-sync-20260903`，完整工作区 stash 为 `52687120d5b2f652ef7bd426f438c7eb36cfb269`。部署验收前不要删除该 stash。

## 2026-09-05 upstream/main 同步结果（v0.2.1）

- 同步范围：`b1748c4ea..ab99d56e9`，共 77 个上游提交、265 个变更文件；版本从 `0.2.0` 推进到 `0.2.1`，本地合并提交为 `5b35db2c2`。
- 同步策略：将同步前完整工作区保存为 stash `pre-upstream-sync-20260905`（`stash@{0}`），并建立同步前备份分支 `backup/pre-upstream-sync-20260905`；在干净基线上合并上游后恢复本地内容，stash 和备份分支均保留未删除。
- 已同步内容：GPT-6 Astra/Codex 模型能力与固定账号清单、ultrafast service tier、上游请求标识和响应头配置、价格目录热重载、Anthropic CLI 版本覆盖、OpenAI 图片 URL 转 Base64 的账户开关与私网下载防护、WebSocket cyber policy、Gemini 自定义模型列表、OpenAI/Anthropic 兼容层及前端账号/分组/用量管理更新。
- 数据库变化：引入上游用量请求标识索引、渠道 reasoning multiplier 和分组 Codex manifest 迁移；迁移文件名与本站已有迁移不冲突。本轮未连接或修改运行中的数据库、余额、任务或生产服务。
- 冲突与兼容处理：
  - 用量日志 INSERT 同时保留上游 `upstream_request_id` 和本站视频输入时长/成本字段，统一为 65 列与 65 个参数，修正单行及无返回值 SQL 的占位符。
  - OpenAI embeddings 同时保留上游响应头传播与本站 token 提取/回退；`OpenAIForwardResult` 同时保留响应头和视频计费任务 ID。
  - 图片 URL 转换兼容上游 `images_url_to_b64_json` 与本站 `openai_image_url_to_b64_json` 两套开关；OAuth/API Key、MiniMax/Firefly、音频、Nano Banana、社区聊天和 NewAPI 余额访问等本站功能继续保留。
  - 上游模型测试和账号编辑流程同时保留 MiniMax、Grok API Key/OAuth、GPT-6 Astra 测试及本地 Grok 扩展字段清理。

### 验证结果

- `go generate ./cmd/server` 通过。
- `go test ./... -count=1` 从 `backend/` 全量通过。
- `pnpm typecheck`、`pnpm lint:check`、`pnpm exec vitest run --reporter=dot` 和 `pnpm build` 从 `frontend/` 通过；Vitest 共 270 个测试文件、1959 项测试。
- `git diff --check`、`git diff --cached --check`、无未解决索引、严格冲突标记扫描及 `git merge-base --is-ancestor upstream/main HEAD` 通过。

- 回滚点：先保存当前工作区，再执行 `git revert -m 1 5b35db2c2`；同步前备份分支为 `backup/pre-upstream-sync-20260905`，完整工作区 stash 为 `stash@{0}`（`pre-upstream-sync-20260905`）。部署验收前不要删除该 stash。

## 2026-09-07 upstream/main 同步结果（v0.2.2）

- 同步来源：`https://github.com/Wei-Shaw/sub2api.git` 的 `upstream/main`；范围为 `ab99d56e9..b7dba6267`，102 个提交、261 个变更文件，版本从 `0.2.1` 推进到 `0.2.2`。本地合并提交为 `37b98f34f75ecef744f377735f879a636578e94d`，未推送本地仓库或发布镜像。
- 同步保护：同步前基线 `5b35db2c2f16691ea9628b5d4dfa5ad81eb76c2c` 已保存为分支 `codex/pre-upstream-sync-20260907`；完整工作区保存为 stash `0aa535a2c26d7a41efddfd552d74aaf832e076ac`（`pre-upstream-sync-20260907`），保留未删除。带索引恢复因上游上下文变化失败，随后按内容恢复并解决 4 个冲突；内容已保留，但原暂存/未暂存划分未完全恢复，提交前需要重新检查暂存区。
- 已同步内容：Astra Pro/Ultra 推理和图片能力、instructions 兼容、固定账号模型目录及模型不可用切换、分组模型白名单、Claude CLI 版本要求、WebSocket 多轮额度和隔离、定价与支付/兑换限制、备份锁及前端翻译完整性检查。
- 兼容处理：保留本站 OAuth/API Key 包装错误重试和渠道监控逻辑，适配上游新增账号参数；本地视频、音频和 Nano Banana 根路径复用上游白名单路由链，状态查询与下载继续保留；重新生成依赖注入并补齐新增测试的本地依赖参数、定价前提和登录状态；补齐两项支付配置文案。首页、登录页和认证布局与同步前快照内容一致。
- 部署注意：新增 `backend/migrations/235_group_model_allowlist.sql`，将 `groups.models_list_config` 更名为 `model_allowlist` 并保留原数据。语义由仅过滤模型列表升级为同时约束请求准入；部署前应核对已启用的分组名单，避免原本可调用但不在列表中的模型被拒绝。本轮未执行生产迁移，未修改数据库、余额或运行中的服务。该范围没有 Go 依赖或前端锁文件变更。
- 验证：`go generate ./cmd/server`、`go test ./... -count=1`、`go test -tags=unit ./internal/server -run '^TestAPIContracts$' -count=1` 全部通过；`pnpm install --frozen-lockfile`、`pnpm lint:check`、`pnpm exec vitest run` 和 `pnpm build` 通过，278 个前端测试文件/1998 项测试通过，构建包含类型检查及新增翻译检查。未执行全部带 `unit`/`integration` 标签的测试或真实上游端到端请求；前端仍有大 chunk 提示。
- 完整性：无未解决索引和冲突标记，`git diff --check`、`git diff --cached --check`、上游祖先关系检查通过。同步前 92 个未跟踪文件均保留，其中 91 个内容哈希一致，视频 service 的差异仅为本轮适配新增账号参数。
- 回滚：不覆盖当前工作区的恢复方式为先执行 `git worktree add -b codex/recover-pre-sync-20260907 /tmp/sub2api-pre-sync-20260907 codex/pre-upstream-sync-20260907`，再执行 `git -C /tmp/sub2api-pre-sync-20260907 stash apply --index 0aa535a2c26d7a41efddfd552d74aaf832e076ac`，在独立目录恢复同步前基线及完整本地改动。该操作不回滚数据库；保留备份分支和 stash 至部署验收完成。

## 2026-09-08 账号测试模型列表空白兼容修复

- 上游 `f88d62ad2` 将 OpenAI 账号测试列表改为实时目录；OAuth 目录转为标准模型列表后不包含显示名称，而测试弹窗原先只展示 `display_name`，导致有模型 ID 的选项显示为空白。
- 测试弹窗现优先显示有效名称；名称缺失、空值或仅包含空白时显示模型 ID。回退后的名称同时用于下拉搜索和选中项展示，提交测试时仍使用原始模型 ID。
- 修复仅涉及测试弹窗展示，保留实时目录来源和已有名称，不修改模型白名单、上游转发、账号鉴权或实际调用逻辑。OAuth 与 API Key 均覆盖。
- 验证：先通过新增测试复现两类账号的空白名称，再确认测试弹窗与 Select 共 12 项测试、定向 ESLint 和前端生产构建通过。需要重新构建并部署包含本修复的版本后生效；本轮未发布镜像或重启服务。

## 2026-09-08 实际账号测试弹窗修正

- 更正上一节的生效范围：此前补丁修改了 `components/account/AccountTestModal.vue`，但账号管理页面实际加载的是 `components/admin/account/AccountTestModal.vue`，因此上一轮验证不能证明实际页面已修好。
- 现已在实际弹窗补齐显示名称：缺失、空值、空白名称回退到模型 ID，保留有效名称；下拉展示、搜索和选中项使用该名称，提交仍使用原始模型 ID。
- 实际弹窗的 OAuth/API Key 两项真实 Select 交互测试先复现空白，再通过显示、搜索、选择和提交验证；包含公共 Select 的 14 项测试及定向 ESLint 通过。测试不连接生产账号；需要重新构建并部署后生效。

## 2026-09-08 用量字段回归验证修正

- 合并后的用量插入包含 65 个字段；`upstream_request_id` 与本站视频输入时长、输入/输出成本字段均保留。两项测试现按实际字段数量与推理等级位置检查，未修改运行时 SQL 或记账行为。
- `cd backend && go test -tags=unit ./internal/repository -count=1` 全部通过，包含手写 INSERT 的列数/占位符及上游请求 ID、会话 ID、请求/实际推理等级字段验证。
- 后续同步若涉及用量字段，须运行上述带 `unit` 标签的仓储测试；单独执行 `go test ./...` 不会覆盖这些测试。

## 2026-09-08 新增账号的模型能力同步提示修正

- 新增账号的同步预览现按本次取得的完整能力信息区分完整、部分完整与不完整，不再依赖账号是否已保存。预览不会持久化账号；已有账号的元数据保存和失败处理保持原有行为。
- 部分成功提示改为“已获取部分模型的完整能力信息；其余模型信息仍不完整”。确实没有完整能力信息时提示“模型 ID 已同步，但尚未获取到完整的模型能力信息”，英文同步调整。这些提示描述能力信息完整程度，不代表模型无法调用。
- 元数据获取来源、补全规则及网络请求方式没有调整；上游或补全来源缺少字段时仍会提示不完整。
- 验证：新增预览的完整/部分/不完整三种回归场景，部分场景修复前失败、修复后通过；元数据同步 service/handler 测试、模型选择/创建/编辑及翻译的 104 项前端测试、定向 ESLint 与前端生产构建通过。需要发布新构建后生效，本轮未推送镜像或部署。

## 2026-09-08 upstream/main 同步结果（v0.2.3）

- 同步来源：`https://github.com/Wei-Shaw/sub2api.git` 的 `upstream/main`；范围 `b7dba6267..270eac697`，76 个提交、271 个变更文件，版本从 `0.2.2` 升至 `0.2.3`。本地合并提交为 `563aba27b28dfcaaec52c42244832556b781d8b3`。
- 同步保护：基线 `37b98f34f75ecef744f377735f879a636578e94d` 保存为 `codex/pre-upstream-sync-20260908`，完整暂存/未暂存及未跟踪内容保存为 stash `591a582408eabfb6ee36ff0d44ce91bfc7e20ee6`，保留未删除。带索引恢复因上游上下文变化失败后，按内容恢复并解决三处冲突，再通过原未暂存补丁的反向索引应用恢复原暂存划分。本轮兼容修正仍在工作区，未另行提交或推送。
- 上游新增：MiniMax 国产渠道、模型能力与余额监控；OAuth 瞬时 429 与已持久化冷却处理；HTTP/2 长流保活、客户端断线取消与用量保留；渠道缓存跨实例失效通知；Grok 媒体资格管理；代理备份/过期切换修复；注册入口控制、周用量成本估计、账号菜单定位及日志存储上限。go-redis 升级为 `v9.22.0`，前端依赖锁文件未变。
- 兼容处理：Responses 同时保留上游客户端断线后的用量提交和本站已有可计费用量判断；账号管理保留弹窗按需加载并接入新版菜单锚点定位；登录页保留本站视觉样式并遵循上游注册开关；依赖注入重新生成并验证渠道缓存通知与本站服务接线。渠道监控的本地测试由固定 8 个渠道改为跟随注册清单，适配新增 MiniMax。
- 本地保护：92 个原未跟踪文件内容哈希全部一致；上游变更范围外的已跟踪文件仅渠道监控测试及本轮记录有改动。首页/认证布局、严格图片 URL 转换、视频恢复与退款、OAuth 重试扩展、社区聊天、用量插入字段及实际账号测试弹窗修复均保留。上游新增后端模型显示名称回退，与本站前端回退兼容。
- 数据库迁移：新增 `236_group_model_allowlist_repair.sql` 修复/补齐分组模型白名单列，保留有效配置；新增 `237_add_minimax_platform.sql` 扩展平台及渠道监控约束，包含本站已有 Grok 支持。仅在隔离测试容器执行迁移，未连接生产数据库。部署新版本时会应用这两个迁移；分组白名单仍同时约束模型列表与请求准入。
- 验证：`go generate ./cmd/server`、`go test -tags=unit ./... -count=1`、`pnpm lint:check`、`pnpm build` 通过。前端完整测试共 283 个文件、2076 项，首跑 2075 项通过，唯一渠道数量旧断言修正后所属文件 2 项复测及定向 ESLint 通过；未重复运行其余已通过用例。
- 迁移与部署验证：`go test -tags=integration ./internal/repository -run '^(TestMigration236|TestMigrationsRunner_IsIdempotent_AndSchemaIsUpToDate)' -count=1 -v` 在临时 PostgreSQL/Redis 中通过，包含三种白名单结构恢复及全量迁移幂等性；`bash deploy/tests/apple-container-test.sh` 使用模拟容器命令通过。完整测试日志位于 `/tmp/sub2api-sync-20260908.8q7Qk9/`。前端保留既有大 chunk 提示；未执行生产账号端到端请求。
- 回滚：执行 `git worktree add -b codex/recover-pre-sync-20260908 /tmp/sub2api-pre-sync-20260908 codex/pre-upstream-sync-20260908`，再执行 `git -C /tmp/sub2api-pre-sync-20260908 stash apply --index 591a582408eabfb6ee36ff0d44ce91bfc7e20ee6`，在独立目录恢复同步前基线及完整本地定制，不覆盖当前工作区。保留分支和 stash 至部署验收完成。本轮未构建/发布 Docker 镜像，也未部署或推送 GitHub。

## 2026-09-09 upstream/main 同步结果（v0.2.4）

- 同步来源：`https://github.com/Wei-Shaw/sub2api.git` 的 `upstream/main`；范围 `270eac697..98d86915b`，3 个提交、19 个变更文件，版本从 `0.2.3` 升至 `0.2.4`。本地合并提交为 `cb76eeba0dd6c81e5cac671be4e05f3633db298d`。
- 同步保护：同步前基线 `563aba27b28dfcaaec52c42244832556b781d8b3` 保存为分支 `codex/pre-upstream-sync-20260909`，完整工作区保存为 stash `038cfb57e67e90550f706c5ba549bbfa32dd5d05`（`pre-upstream-sync-20260909`），均保留。带索引恢复因上游测试上下文变化失败后，按内容恢复并解决一处测试初始化冲突，再反向应用原未暂存补丁到索引，恢复原有暂存/未暂存划分。
- 上游新增：`gpt-image-2.5-flare`、`gpt-image-2.5-sunburst` 及日期快照支持、默认模型清单和价格；OAuth 账号测试补齐图片模型列表，改善上游生图错误提示；保留图片输入 token 用量；修复因生图主控模型被拒而错误冷却图片模型的问题。
- 部署注意：OAuth/Setup Token 生图的 Responses 主控模型默认由 `gpt-5.4-mini` 改为 `gpt-5.6-luna`，可通过 `SUB2API_IMAGES_MAIN_MODEL` 覆盖，例如 `SUB2API_IMAGES_MAIN_MODEL=gpt-5.6-sol`。四份 Compose 模板已接入此环境变量；实际生成图片的模型仍单独传入 `tools[].model`，显式指定的 Responses 文本主控模型保持原值。账号或分组已设置模型白名单时，需要将希望开放的新图片模型加入名单。
- 本地兼容：保留严格图片 URL 转 Base64 的账号开关、视频恢复与退款、用量插入修复、OAuth 重试扩展、实际账号测试弹窗名称回退及模型能力同步提示。冲突处理保留上游主控模型环境变量测试，两个受影响测试文件继续由本地统一测试入口设置 Gin 模式；92 个原未跟踪定制文件 SHA-256 全部一致。
- 验证：`cd backend && go test -tags=unit ./... -count=1` 全量通过；`cd frontend && pnpm exec vitest run` 的 283 个文件、2076 项测试全部通过，`pnpm lint:check`、`pnpm build` 通过，构建含翻译和类型检查。`sh deploy/tests/docker-compose-gateway-env-test.sh` 通过；四份 Compose 在默认及显式指定生图主控模型时共 8 项配置渲染检查通过。无未解决冲突，格式检查和上游祖先检查通过。日志目录为 `/tmp/sub2api-sync-20260909.46CcX6/`。
- 本批没有数据库迁移、依赖或依赖注入改动；未执行真实上游端到端验证。前端保留既有大 chunk 提示。同步已完成于本地，本轮未构建/发布镜像、推送 GitHub 或部署。
- 回滚：执行 `git worktree add -b codex/recover-pre-sync-20260909 /tmp/sub2api-pre-sync-20260909 codex/pre-upstream-sync-20260909`，再执行 `git -C /tmp/sub2api-pre-sync-20260909 stash apply --index 038cfb57e67e90550f706c5ba549bbfa32dd5d05`，在独立目录恢复同步前代码及完整本地定制，不覆盖当前工作区。保留备份分支和 stash 至部署验收完成。

## 2026-09-09 测试上游账号自定义模型兼容修复

- OpenAI 账号测试模型列表在实时上游目录成功时，也会补入账号 `model_mapping` 中的具体请求入口；因此“模型白名单＋模型映射”里添加的自定义名称可以直接选择测试。通配符入口继续不展示，避免把不可具体提交的模式当成模型。
- 测试提交仍使用所选请求入口，并通过账号现有模型映射解析到真实上游目标；没有改变上游模型发现、缓存、鉴权或转发规则。OAuth、Setup Token 和 API Key 均使用同一兼容逻辑。
- 验证：`cd backend && go test -tags=unit ./internal/service -run 'TestFetchOpenAIAccountModels|TestAccountGetMappedModel' -count=1` 通过，新增自定义映射入口、重复项和通配符回归覆盖。未执行真实上游测试。
- 回滚：恢复 `backend/internal/service/account_test_service.go` 和 `backend/internal/service/account_test_models_test.go` 至本轮修改前版本即可；同步备份分支与 stash 仍可用于完整回滚。

## 2026-09-09 镜像发布（自定义模型测试修复）

- 基于当前工作区 v0.2.4（commit `cb76eeba0dd6-dirty`）构建并推送 `iotwq/china-api:latest`，发布 `linux/amd64` 与 `linux/arm64`。
- 远端 OCI index 摘要：`sha256:5806b6cb091ff9fa9489ca2226d2647299ef5a4fa5658cdc569e664e7d73a52a`；amd64 manifest：`sha256:f41628b42711f3600b1da0fe34d1a9a71a1d842028929bca291b64896d47809a`；arm64 manifest：`sha256:b1bc94fb823acb7194ab772fa27aa4b0c0997d8aac8425eed3afd863309df4d1`。
- 构建过程包含前端翻译检查、生产构建及两个架构的 Go 编译；镜像内版本检查确认 `0.2.4`、commit `cb76eeba0dd6-dirty` 和构建时间 `2026-09-09T13:52:20Z`。Docker Hub 拉取启动验证因 CDN EOF 未完成，远端 manifest 和构建元数据已核对一致。
- 部署时拉取 `iotwq/china-api:latest` 并重新创建应用容器；如需固定版本，使用上述 index 摘要。发布前 latest 摘要为 `sha256:8c378a968668f81fabbc97d6fb8dad50048065f603e4dd967b87d17d8c993360`，可用 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:8c378a968668f81fabbc97d6fb8dad50048065f603e4dd967b87d17d8c993360` 回滚标签；本轮未部署服务或修改数据库。

## 2026-09-11 渠道监控 Anthropic 策略拒绝换号修复

- 渠道监控请求携带内部探测头且上游返回 400 `probe_request_rejected` / `request blocked by gateway policy` 时，网关现在将其转换为 `UpstreamFailoverError`，由监控账号切换循环继续探测同组下一个账号。
- 仅识别上述明确的探测策略拒绝；普通用户请求、缺少内部探测头的请求、以及其他 400 参数错误仍按原有错误语义返回，不会扩大重试或账号冷却范围。
- 变更文件：`backend/internal/service/gateway_forward.go`、`backend/internal/service/openai_gateway_upstream_errors.go`、`backend/internal/service/channel_monitor_failover_test.go`。
- 验证：`cd backend && go test -tags=unit ./internal/service -run 'TestShouldFailoverChannelMonitorProbeError|TestGatewayService' -count=1` 通过；`gofmt` 与 `git diff --check` 通过。未执行真实上游监控请求。
- 回滚：恢复上述三个文件至本轮修改前版本；部署前保留当前镜像/代码版本，出现异常时重新部署上一版固定 digest。

## 2026-09-11 镜像发布（渠道监控换号修复）

- 基于当前工作区构建并推送 `iotwq/china-api:latest`，包含渠道监控 Anthropic 策略拒绝换号修复；发布平台为 `linux/amd64` 与 `linux/arm64`。
- Docker Hub OCI index 摘要：`sha256:83089738ee533188e404a1dd5781ce2d81735f05537c77d7e4db56819ae8dc4f`；amd64 manifest：`sha256:0db9dbc6a4d0c4398b9fbdd2f89066640d50df9e8239bd29d983eb3398be684e`；arm64 manifest：`sha256:ca592d0f1d993ce21c54546a7755b8ed9f712a6cf1ec6092e0bce44cc4f0eed3`。
- 验证：远端 `docker buildx imagetools inspect iotwq/china-api:latest` 已确认 OCI index 与双架构 manifest；amd64 镜像本地加载后执行 `--version` 通过，显示版本 `0.2.4`、commit `cb76eeba0dd6-dirty`。首次构建遇到 Docker Hub registry EOF，重试后成功完成。
- 部署时拉取 `iotwq/china-api:latest` 并重新创建应用容器；如需固定版本，使用上述 index 摘要。回滚可重新部署上一版固定 digest `sha256:5806b6cb091ff9fa9489ca2226d2647299ef5a4fa5658cdc569e664e7d73a52a`。

## 2026-09-12 NAI Diffusion 图片模型路由兼容

- 将 `nai-diffusion-4-5-full` 和 `nai-diffusion-5-full` 纳入 OpenAI 兼容图片模型识别，使它们可以通过 `POST /v1/images/generations` 进入现有图片转发链路。
- 两个模型继续复用已有模型映射、渠道定价、用量统计及 URL/Base64 响应处理；未新增上游协议转换。
- 验证：`cd backend && go test -tags=unit ./internal/service -run 'TestOpenAIGatewayServiceParseOpenAIImagesRequest_(AllowsNaiImageModels|RejectsNonImageModel|AllowsGrokImageModels)$|TestImageGenerationIntent' -count=1` 通过；`gofmt` 与 `git diff --check` 通过。未执行真实 NAI 上游请求。
- 回滚：恢复 `backend/internal/service/openai_images.go` 和 `backend/internal/service/openai_images_test.go` 至本轮修改前版本，并重新部署上一版固定镜像 digest。

## 2026-09-13 镜像发布（计费修复与视频模型文档）

- 基于当前完整工作区（版本 `0.2.4`、commit `cb76eeba0dd6-dirty`、构建时间 `2026-09-13T09:30:21Z`）构建并推送 `iotwq/china-api:latest` 和版本标签 `iotwq/china-api:0.2.4-20260913-173021`，支持 `linux/amd64`、`linux/arm64`。
- 包含近期计费完整性修复、视频模型支持，以及新增的 SD2.0 / ViralDance 933 和 Wan3.0 用户接口文档；保留工作区其他既有定制。
- 两个标签均指向 OCI index `sha256:d21b296f66b0dde95e3de86f0cbb86ae2ab28e0373420417549d53bdd0d4c4b0`。amd64 manifest 为 `sha256:d689738daa76215c59302e59b7420d65278cfc1b3959ee1efed5f1fdf237c4bf`，arm64 manifest 为 `sha256:31ca6168a17519d0b11e930c8873dfa132282d8de0c200ce52cd0a4cfacb028e`。
- 验证：镜像内前端翻译与类型检查、前端生产构建及双架构 Go 编译通过；两个远端标签摘要与构建元数据一致；两个架构的镜像均从 Docker Hub 拉取并在断网、只读容器内执行 `--version` 成功。另确认镜像内包含新增用户文档和迁移 `239_pending_image_settlements.sql`、`240_video_billing_review.sql`。
- 首次构建遇到 Go 依赖下载 EOF，利用缓存重试成功；amd64 验证首次遇到 Docker Hub CDN EOF，按发布 manifest 重试拉取后验证通过。日志与元数据保存在 `/tmp/sub2api-image-20260913.154ccO/`。
- 部署时拉取上述版本标签或固定 index digest，并重新创建应用容器；正常启动会执行随镜像携带的数据库迁移。账务队列、待核对状态及检查方式见 `docs/BILLING_INTEGRITY.md`。本轮仅发布镜像，未部署服务或改写生产数据。
- 发布前 latest 为 `sha256:83089738ee533188e404a1dd5781ce2d81735f05537c77d7e4db56819ae8dc4f`。可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:83089738ee533188e404a1dd5781ce2d81735f05537c77d7e4db56819ae8dc4f` 恢复标签，或直接部署该旧摘要；已升级数据库时保留新表和账务记录，回退前核对待结算队列，旧镜像不会处理新队列及待核对任务。

## 2026-09-13 镜像发布（糖果题规则澄清）

- 基于当前工作区构建并推送 `iotwq/china-api:latest` 与 `iotwq/china-api:0.2.4-20260913-200554`，支持 `linux/amd64`、`linux/arm64`；版本 `0.2.4`，commit `cb76eeba0dd6-dirty`，构建时间 `2026-09-13T12:05:54Z`。
- 包含糖果题“允许按形状选择、不能预辨口味、仅取出计数且不放回”的明确规则，标准答案仍为 21，保留此前计费修复与全部既有定制。
- 两个标签的 OCI index 均为 `sha256:a92d2fae1eb99fa4ed0ba545002bbcc31077efab44bfc33b791f0c4bdbe02b08`；amd64 manifest 为 `sha256:6119a48f47f4b0948af517fd41b777fce3221d69beda2b4aa8fa5374ae6ed504`，arm64 manifest 为 `sha256:f47b3f3f101c11fa20b8834e858b54828b096712c7fbedc2e598812943e9ccc0`。
- 验证：双架构 Go 编译及推送通过，前端复用同内容已通过验证的构建缓存；两个远端标签摘要与构建 metadata 一致；两个架构均从 Docker Hub 拉取并在断网只读容器执行 `--version` 成功。另在发布镜像中确认新增操作规则字符串存在。日志和元数据位于 `/tmp/sub2api-image-candy.S7rd1g/`。
- 本轮未部署或重启线上服务。拉取新标签或固定摘要并重建应用容器后，后续智力检测使用新题面；历史结果不重判，以部署时间区分前后测试。题目、判分与比较边界见 `docs/channel-monitor-intelligence.md`。
- 发布前 latest 为 `sha256:d21b296f66b0dde95e3de86f0cbb86ae2ab28e0373420417549d53bdd0d4c4b0`。恢复标签可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:d21b296f66b0dde95e3de86f0cbb86ae2ab28e0373420417549d53bdd0d4c4b0`，或重新部署此前固定版本 `0.2.4-20260913-173021`；本次题面改动无新增数据库迁移，不删除历史或账务数据。

## 2026-09-13 镜像发布（智力检测 low 与延迟分离）

- 基于当前完整工作区构建并推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.4-20260913-224529`，支持 `linux/amd64`、`linux/arm64`；版本 `0.2.4`，commit `cb76eeba0dd6-dirty`，构建时间 `2026-09-13T14:45:29Z`。
- 包含普通探活与糖果题分离：智力检测耗时、答错或超时不再影响渠道状态和最新延迟；OpenAI/Codex 糖果题按 Chat Completions/Responses 协议固定发送 low 推理强度，覆盖模板并支持协议回退。保留此前糖果题规则、计费修复和视频等既有定制。
- 两个标签均指向 OCI index `sha256:451de5fe8933c21d1544a9781c1804828b2c066afb7b725fd1213ae315b59e63`。amd64 manifest：`sha256:ee9c31e0a360a31f8f35671f8f5355a8be901d25755f1777ad1dca91b7e68263`；arm64 manifest：`sha256:8baac637b92a2579c62851d9bce17cda76ff46b26e58ac5da6abb960dd11bdbf`。
- 验证：前端翻译检查、类型检查和生产构建通过；两个架构的 Go 编译及镜像推送成功；两个远端标签与构建 metadata 摘要一致；按各架构发布 manifest 从 Docker Hub 拉取后，在断网、只读容器执行 `--version` 均成功，版本、commit、构建时间一致。沿用源码修改阶段已通过的监控回归，不重复执行没有变化的用例。前端保留既有大 chunk 提示。
- 构建、元数据及版本检查日志保存在 `/tmp/sub2api-image-low.5q3AhM/`。本轮仅发布镜像，未部署或重启线上服务，也未修改数据库。拉取新镜像并重建应用容器后，后续检测使用新逻辑；旧历史延迟和智力结果不重算，推理强度及时间口径应按部署时间区分。
- 发布前 latest 为 `sha256:a92d2fae1eb99fa4ed0ba545002bbcc31077efab44bfc33b791f0c4bdbe02b08`。可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:a92d2fae1eb99fa4ed0ba545002bbcc31077efab44bfc33b791f0c4bdbe02b08` 恢复标签，或重新部署固定版本 `0.2.4-20260913-200554`；本轮两项监控修改不新增数据库迁移，保留所有历史及账务数据。

## 2026-09-15 同步原项目至 0.2.5

- 原项目为 `https://github.com/Wei-Shaw/sub2api.git`（`upstream/main`）。基线由 `98d86915becae9fe9491a91ffc6defd5235c8d2b` 更新至 `881f3202694c6bc932446931a30c27d9675178b9`，共 197 个提交；本地合并提交为 `a85098e8779876b7a5334ff94099e94564e46198`，版本由 0.2.4 升至 0.2.5。
- 同步 OAuth 原生图片接口与图片缓存计费、WebSocket 稳定性及用量窗口修正、OpenCode Go 渠道、账号/密钥/订阅管理及渠道监控路径修正等上游更新。完整实际变更文件和逐项说明见 [本次文件清单](upstream-sync-20260915-files.md)。
- 保留本地图片 URL 转 Base64 开关、视频任务恢复与失败补偿、图片待结算、cc-switch 余额查询、群聊、国内优化地址、自定义首页与登录页、模型映射测试列表，以及糖果题固定 low、独立于普通探活延迟的逻辑。同步前 117 个未跟踪文件逐一比较 Git blob，内容全部一致。
- 处理 1 处基线合并冲突和恢复本地改动时的 11 处冲突：Gemini 在合并上游带内错误观测的同时保留图片计数和读流异常时已产生用量；账号调度缓存同时保留调度阈值与能力信息；模型测试保留所有具体映射键；金额显示保留负数退款并使用 8 位小数；前端测试同时保留远端功能和本地定制用例。
- 适配新增测试的本地依赖参数与监控返回值；旧 OAuth Responses 图片流测试使用仍走该路径的 gpt-image-1，原生图片接口使用上游新增独立测试。补充 Gemini 正常结束/中途读流失败时的图片去重、用量及终止信号联合回归。旧迁移断言更新为本地 240 迁移已存在的 recovery_owned 索引，未改写已发布迁移。
- 验证：Wire 和 Ent 生成成功；前端 301 个测试文件、2,274 项用例通过，ESLint、翻译与类型检查、生产构建通过。后端执行全量 unit 检查，新增测试构造问题修正后 handler/routes 复测通过；service 完整执行得到 14,142 个通过、3 个跳过、1 个旧流测试模型不匹配，该项修正后定向复测通过，其他包通过。最终结果采用“全量检查 + 失败项定向复测”，未将有失败的那次全量运行标为全绿。额外 Gemini 联合回归通过。跳过项为原有待决身份、真实 OpenAI 对照和插件集成测试。
- 在临时 PostgreSQL/Redis 容器执行迁移并发锁、重复应用、结构核对以及三种实际用量入库验证，全部通过；Apple container 生命周期夹具和 Compose 网关环境脚本测试通过。源码与暂存区 diff 格式检查通过，无未解决冲突。测试日志、原始暂存/未暂存补丁和备份清单位于 `/tmp/sub2api-sync-20260915.I44YmZ/`；前端保留既有大 chunk/Browserslist 提示。
- 新增远端迁移 `238_opencode_go_platform.sql` 与 `238_purge_unlimited_user_platform_quotas.sql`。迁移器以完整文件名作为主键，允许与本地 `238_channel_monitor_intelligence.sql` 共存；后者远端迁移仅清理日/周/月额度全部未配置的行。后续发布并启动新版本时会执行迁移；本轮仅同步本地源码，未构建推送镜像、未部署或修改生产数据库。
- 同步前保护分支 `codex/pre-upstream-sync-20260915` 指向 `cb76eeba0dd6c81e5cac671be4e05f3633db298d`，完整本地改动快照为 stash `77d38b39662c85487868baef90ff6170f719b4ca`，仍保留。现有暂存/未暂存区分已恢复；本轮兼容调整和记录保留在工作区供后续提交。

需要恢复同步前源码时，在仓库根目录执行以下命令，创建独立恢复工作区并还原本地改动及暂存状态，保留当前目录和备份：

```sh
git worktree add -b codex/recover-pre-sync-20260915 /tmp/sub2api-pre-sync-20260915 codex/pre-upstream-sync-20260915
git -C /tmp/sub2api-pre-sync-20260915 stash apply --index 77d38b39662c85487868baef90ff6170f719b4ca
```

## 2026-09-16 镜像发布（0.2.5、群聊修复与淡金私聊入口）

- 基于当前完整工作区构建并推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.5-20260916-000512`；支持 `linux/amd64`、`linux/arm64`。版本 `0.2.5`，commit `a85098e87798-dirty`，构建时间 `2026-09-15T16:05:12Z`。
- 包含 9 月 15 日同步的上游 197 个提交、群聊/私聊九项体验修复、按已显示消息提交已读、附件恢复和上传进度，以及淡金色私聊站长入口和 CSS 轻动效；保留既有本地定制。
- 两个标签的 OCI index 均为 `sha256:480c7859fa16229ea44dccf8d29ddd9d8103e24a35f459142116bb660654262c`。amd64 manifest 为 `sha256:4a18aa1556b8b4d16ce5d0786d4657355fb092ee599dedb5308f545605f51c38`；arm64 manifest 为 `sha256:07e53417bd6ff7ca5f1033cd9e2e95704142e5caa60476fd03ce15170b951ff3`。
- 验证：镜像内前端翻译完整性 3 项、类型检查和生产打包、两个架构的 Go 编译通过；构建及推送退出 0。两个远端标签摘要与 metadata 一致；按各自 manifest 从 Docker Hub 拉取，随后在断网、只读容器执行 `--version`，两个架构均退出 0，版本、commit、构建时间一致。amd64 首次拉取遇到 auth.docker.io EOF，重试成功。保留原有前端大 chunk/Browserslist 提示。
- 证据目录 `/tmp/sub2api-image-chat-gold.guIwiq/`，包含 `build.log`、`metadata.json`、`inspect-latest.txt`、`inspect-fixed.txt`、各架构 pull/version 日志及发布前摘要。
- 本轮仅发布镜像，未部署、重启服务或修改生产数据。部署后需刷新旧前端页面以使用新的私聊已读请求参数。相较上次 0.2.4 镜像，此版含前述两个上游 238 迁移；后续启动应用时才执行，镜像拉取和版本检查不执行迁移。
- 发布前 latest 为 `sha256:451de5fe8933c21d1544a9781c1804828b2c066afb7b725fd1213ae315b59e63`（固定版本 `0.2.4-20260913-224529`）。需要恢复标签时可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:451de5fe8933c21d1544a9781c1804828b2c066afb7b725fd1213ae315b59e63`；镜像标签回滚不会回滚已执行的数据库迁移，线上回滚需结合部署后的数据库状态核对。本轮未执行回滚。

## 2026-09-16 镜像发布（私聊徽标与侧栏未读提醒）

- 基于当前完整工作区构建并推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.5-20260916-013453`，支持 `linux/amd64`、`linux/arm64`。版本 `0.2.5`，commit `a85098e87798-dirty`，构建时间 `2026-09-15T17:34:53Z`。
- 包含私聊站长入口的“新消息”徽标，以及进入群聊页面后侧栏提醒按实际未读状态保留的修复；保留上一版镜像的上游同步与全部既有定制。
- 两个标签均指向 OCI index `sha256:a6a6138b13d3cc270d4028665b8c79f1b65f5d5c14fbb0da8c51d5b67c5bfd67`。amd64 manifest：`sha256:38d9617906ff14271b3cf27a6adaa12c1315020228bc90fc118052eea1d354e0`；arm64 manifest：`sha256:6b981166925e7aefd1bab9306364ababc652d26b4ead477d97c742d48b68d60b`。
- 验证：前端翻译 3 项、类型检查、生产构建及双架构 Go 编译通过；构建推送退出 0。两个远端标签的摘要与 metadata 一致；两个架构均按发布 manifest 从 Docker Hub 拉取，在断网、只读容器执行 `--version` 退出 0，版本、commit、构建时间一致。AMD64 拉取先后遇到认证端点和 CDN EOF，第二次重试成功。保留原有前端 Browserslist/大 chunk 提示。
- 构建、摘要及拉取/版本验证证据位于 `/tmp/sub2api-image-unread.HISCNW/`。本轮未部署、重启服务或修改生产数据；相对固定版本 `0.2.5-20260916-000512`，本次新增业务改动仅涉及前端，无新增数据库迁移。部署新镜像后刷新页面使用新提醒逻辑。
- 发布前 latest 为 `sha256:480c7859fa16229ea44dccf8d29ddd9d8103e24a35f459142116bb660654262c`。需要回滚时可部署上一固定标签 `0.2.5-20260916-000512`，或执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:480c7859fa16229ea44dccf8d29ddd9d8103e24a35f459142116bb660654262c` 恢复 latest 标签；本轮未执行回滚。

## 2026-09-18 同步原项目近期提交

- 原项目 `https://github.com/Wei-Shaw/sub2api.git` 的 `upstream/main` 从 `881f3202694c6bc932446931a30c27d9675178b9` 同步至 `efe9aab1e4ec89a42ba45e8dac20e882c5409a6a`，共 52 个提交、71 个上游变更文件；本地合并提交为 `03f6b278af9e917f96e075b8e55f4bb42bb4ee79`。版本仍为 `0.2.5`，本次同步无新增数据库迁移。
- 包含暂停账号 OAuth token 刷新、客户端取消后保存 OpenAI 响应账号绑定、Gemini/Antigravity 模型映射发现、DeepSeek 工具输出图片与并行结果兼容、严格 Chat 上游 developer 角色规范化、分组用量汇总查询优化、兑换历史分页、支付/TOTP/公告/代理测试等前端修复，以及 grpc 1.83.2 依赖升级。
- 保留本地 OAuth 线路保护、指纹/TLS/并发治理、图片转换与计费补偿、视频恢复、群聊/私聊、智力监控及首页等定制。恢复本地改动时解决 Gemini 模型列表与监控测试两处冲突；OpenAI 分组继续按 Gemini-native 能力选账号，不跨到 Antigravity。新增真实处理器回归覆盖能力开启与关闭。
- 验证中发现既有智力检测取消用例存在模拟 HTTP 服务抢先返回空 200 的偶发竞争；仅调整测试等待客户端断开，不改线上逻辑和断言。修正后带 race 重复 100 次通过，后端 service 全包复测通过；首次后端全套运行中其他包均通过。
- 前端 316 个测试文件、2377 项测试全部通过，lint、类型检查及生产构建通过；后端生产构建和 `--version` 检查通过；OAuth/TLS/并发定向 race 回归通过。临时 PostgreSQL 18.1 / Redis 8.4 集成验证覆盖暂停账号刷新、兑换分页隔离与排序、分组汇总尾部以及时区/DST 边界，均通过。验证日志位于 `/tmp/sub2api-sync-20260918.tGFjml/`；未调用生产账号或修改线上数据库。
- 同步前 485 个既有脏文件逐一核对无丢失、暂存状态保持；除上游重叠文件、上述测试夹具和本轮追加文档外，内容哈希一致。原有暂存与未暂存改动保留，不将全部定制打包进合并提交。逐文件说明见 `progress.md` 的本日同步条目。
- 恢复点：分支 `codex/pre-upstream-sync-20260918` 指向同步前 `a85098e8779876b7a5334ff94099e94564e46198`；完整工作区 stash `db7153f66d990e693f0d1ee6bc42e7a770dfa936` 保留。需要恢复时，在仓库根目录依次执行以下命令，在独立目录恢复同步前代码与定制，不覆盖当前工作区：

```sh
git worktree add -b codex/recover-pre-sync-20260918 /tmp/sub2api-pre-sync-20260918 codex/pre-upstream-sync-20260918
git -C /tmp/sub2api-pre-sync-20260918 stash apply --index db7153f66d990e693f0d1ee6bc42e7a770dfa936
git -C /tmp/sub2api-pre-sync-20260918 add -f -N docs/upstream-sync-20260915-files.md
```

- 本轮仅完成本地同步和验证，未构建或推送 Docker 镜像，未推送 GitHub 或部署服务；发布需使用包含已恢复本地定制的当前工作区。

## 2026-09-18 镜像发布（最新上游同步与 OAuth 保护）

- 基于包含全部本地定制的当前工作区构建并推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.5-20260918-211628`，支持 `linux/amd64`、`linux/arm64`。版本 `0.2.5`，commit `03f6b278af9e-dirty`，构建时间 `2026-09-18T13:16:28Z`。
- 包含本日同步的上游 52 个提交、Gemini 模型列表兼容处理及此前 OAuth 线路保护/指纹/TLS/并发修复，保留媒体、计费、群聊和监控等既有定制；本轮未修改业务代码。
- 两个标签的 OCI index 均为 `sha256:d1733b09ba45e6eecdb17900a35816ad9eb2b2111984fd2dc2556b0b110586d3`。amd64 manifest 为 `sha256:b8d8207fd17a9bfe062c2f14f5c93cad34caea1cc25f3135bbac2ea240166519`，arm64 manifest 为 `sha256:f1cdbc191e3c099b57d2f56813d1b066083df273897bd762da3a64e7e862c7c0`。
- 验证：镜像内前端翻译 3 项、类型检查、生产打包和双架构 Go 编译通过，构建推送退出 0；两个远端标签摘要与构建 metadata 一致。按各架构发布 manifest 从 Docker Hub 拉取后，在断网、只读、非 root 容器中执行 `--version` 均退出 0，版本、commit、构建时间一致。AMD64 首次拉取遇到 Docker Hub token 端点 EOF，重试成功；保留原有前端 Browserslist/大 chunk 提示。
- 沿用本日源码同步阶段已通过的前端 2377 项测试、后端全包检查及 service 修正后复测、数据库集成和 OAuth/TLS 定向 race 回归；本轮代码未再变动，不重复全套测试。构建、metadata、远端清单、拉取与版本日志位于 `/tmp/sub2api-image-sync-20260918.qQjjcI/`。
- 本轮仅发布镜像，未部署、重启服务或修改生产数据；本次上游同步无新增数据库迁移。上线时拉取新镜像并重新创建应用容器；版本检查仅验证发布产物可执行，不代表已进行线上端到端验证。
- 发布前 latest 为 `sha256:933be4db3e0454c074136d89eacbce4b46292356a723f80f85a8041a68cba9bf`。需要回滚时可部署该固定摘要，或执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:933be4db3e0454c074136d89eacbce4b46292356a723f80f85a8041a68cba9bf` 恢复 latest 标签；本轮未执行回滚。

## 2026-09-18 补齐 Codex 票据分支功能，保留 0.2.5 版本号

- 按用户指定来源 `https://github.com/Tinghecui/sub2api/tree/feat/openai-codex-turn-state-ticket` 合并至 `3c2f05c957b4b93866318ec8695fc5a28fff70eb`。该分支在当前上游基线 `efe9aab1e` 上独有 6 个提交：`bc47e212b`、`d14054afc`、`1ba0b4fed`、`f79381bbd`、`c44e6893c`、`3c2f05c95`；本地合并提交为 `86f0e36653d25808ec2e6e150d43b42050d0eb88`。
- 合并票据后台采集、默认 1 小时有效期、目标模型请求注入、默认关闭的管理开关、独立采集代理与脱敏校验、账号状态展示、导出去敏以及生命周期/compact/账号编辑保护。版本继续为 `0.2.5`，未同步发布版本号、未新增迁移。
- 完整截图中的兑换分页、分组统计查询优化、暂停 OAuth 刷新、Antigravity/Gemini/DeepSeek/Chat 兼容和前端修复已由前次 52 个提交包含；逐项对应和使用方法见 [Codex 票据说明](OPENAI_CODEX_TICKETS.md)，未重复移植。
- 原有本地修改已备份并恢复；基线合并无冲突，恢复定制时解决账号保护校验、网关后台成员和编辑弹窗 3 处冲突。489 个原有脏文件无丢失，暂存状态恢复；除本轮上游交集和记录文档，内容哈希一致。
- 新增本地兼容适配：采集请求共享已开启 OAuth 保护的并发/RPM 预算，避免后台探测绕过原有限制；仍使用采集专用代理和短连接，不更改生产传输/TLS。回归先复现满载仍外发，再适配并通过；同时覆盖 HTTP/透传/WS 与四种指纹模式联合行为。
- 验证：前端 316 个测试文件、2381 项测试，以及 lint、翻译、类型检查、生产构建通过；完整后端 unit 通过，最终并发适配后再次执行票据/OAuth/HTTP/WS/TLS 定向 race 回归通过；临时数据库账号更新/批量更新集成测试通过。Wire 生成、后端 embed 编译及程序 `--version` 检查通过，版本仍为 0.2.5。证据位于 `/tmp/sub2api-codex-ticket-20260918.B6W6Uh/`。
- 默认保持关闭。本轮未访问官方账号实际采集、未做质量对照实验，也未构建镜像、推送 GitHub 或部署。开启后的缺票暂停、采集配额消耗和缓存生效边界见上述说明；不能把票据长度当作保证模型质量的依据。
- 回滚点为 `codex/pre-codex-ticket-20260918`（`03f6b278af9e917f96e075b8e55f4bb42bb4ee79`）及完整工作区 stash `20bc5d35e6d8a73d3e636d9075b5d97cc71a90ec`，均保留。独立目录恢复命令见票据说明和本轮 `progress.md`，不覆盖当前工作区。

## 2026-09-18 镜像发布（Codex 票据与并发保护兼容）

- 基于当前完整工作区构建并推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.5-20260918-224942`，支持 `linux/amd64`、`linux/arm64`。版本保持 `0.2.5`，commit 为 `86f0e36653d2-dirty`，构建时间为 `2026-09-18T14:49:42Z`；包含本轮票据分支的 6 个提交、采集请求共享 OAuth 并发/RPM 预算的本地适配，以及既有定制。
- 两个标签均指向 OCI index `sha256:f2b72eda8f9cb958194e1b54c6995d55b2c0eac5ff3bfe955fff0e8efc183099`。amd64 manifest：`sha256:49e28b09f6e8d5d7465ae5a843afe295b524007b4b0d71422731e9de09565787`；arm64 manifest：`sha256:06679afbfc6201fe3358e070590a792bb377ef3d81c943da707137d3dbb733cf`。
- 验证：镜像内前端翻译 3 项、类型检查、生产打包及双架构 Go 编译通过；构建推送退出 0，两个远端标签与 metadata 摘要一致。arm64 从 Docker Hub 拉取后执行 `--version` 通过；amd64 远端拉取连续遇到 CDN/registry EOF，改用构建缓存导出，其 manifest 与远端发布摘要完全一致，加载后执行 `--version` 通过。两者均使用断网、只读、非 root 容器，版本、commit、构建时间一致；未将 amd64 远端下载记为通过。
- 沿用源码合并阶段已通过的前端 2381 项测试、后端全量 unit、最终适配后的定向 race 和临时数据库集成验证；本轮未改业务代码。保留已有 Browserslist/大 chunk 提示。构建、远端清单、拉取尝试、同摘要导出与版本核验记录位于 `/tmp/sub2api-image-ticket-20260918.F4GrI8/`。
- 本轮仅发布镜像，未部署、重启服务或修改生产数据；票据开关仍默认关闭，无新增数据库迁移。版本执行检查不代表官方票据采集或线上端到端验证。上线时拉取新镜像并重建应用容器，票据使用前提见 [Codex 票据说明](OPENAI_CODEX_TICKETS.md)。
- 发布前 latest 为 `sha256:d1733b09ba45e6eecdb17900a35816ad9eb2b2111984fd2dc2556b0b110586d3`，固定版本 `0.2.5-20260918-211628`。需要恢复标签时可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:d1733b09ba45e6eecdb17900a35816ad9eb2b2111984fd2dc2556b0b110586d3`；本轮未执行回滚。

## 2026-09-19 同步原项目 0.2.7

- 同步原项目 `https://github.com/Wei-Shaw/sub2api.git` 的 `upstream/main` 至 `1a9d49e16f7a22c432b428fce4af8d731f1fa364`，新增 19 个提交、60 个上游变更文件。本地合并提交为 `1f0e0257a1aa1d12bc9fef28246ca80ab2f992d5`，版本随正式上游更新为 `0.2.7`；没有新增数据库迁移或依赖清单变动。
- 本次包含 Seedance Ark 原生任务接口、插件宿主 KV/账号目录及只读状态桥、Gemini thinking 变体和 SSE 注释兼容、DeepSeek 推理历史兼容、Anthropic 工具根联合类型兼容、国内 Coding Plan 额度 403 暂停识别，以及移动端模型广场入口调整。
- 保留本地 292 按账号选择与全局强制、OAuth 线路保护、Gemini/MiniMax 能力、图片转换、视频与计费补偿、群聊和监控等定制。503 个既有脏文件无丢失，原暂存内容保留；上游交集、必要兼容修改和本次记录以外的文件内容哈希一致。原有定制未整体提交进合并提交。
- 本地恢复时解决路由测试、创建账号、编辑账号和能力类型共 4 处冲突。账号界面保留 5 种能力，并修正恰好选中两个非默认能力时的省略条件；增加创建和编辑弹窗混合能力保存回归。插件账号目录改由正式 Wire provider 装配，避免再次生成时丢失上游仅写入生成文件的注入。
- Seedance 为显式启用能力：OpenAI API Key 账号配置自定义服务地址后勾选 Seedance，使用原生 `/api/v3/contents/generations/tasks`（兼容其他已登记前缀），不替换本地 `/v1/videos` 等视频路径。新接口按成功轮询返回的 completion tokens 结算，仍有需客户端轮询、任务绑定 24 小时等上游实现边界，详见 [Seedance 原生接口说明](seedance-api.md)。未对接生产视频账号实测。
- 已通过前端 316 个测试文件、2395 项测试、lint、类型检查与生产构建；后端全量 unit、Wire 生成、embed 生产构建及 `--version` 检查通过（0.2.7）。票据/OAuth/Seedance/Grok/插件和本地 TLS 定向 race 回归通过。首次 race 命令误带不存在的 `internal/pkg/http` 目录，修正命令后全部通过，未将首次非零退出记为通过。测试产物位于 `/tmp/sub2api-sync-20260919.7pjorX/`；保留原有前端 Browserslist/大 chunk 及测试环境提示。
- 本轮只同步和验证本地源码，未构建/推送 Docker 镜像、推送 GitHub、部署或修改线上数据。合并提交之外还有恢复的本地定制与本轮兼容修改，后续发布应基于完整工作区。
- 同步前恢复点为分支 `codex/pre-upstream-sync-20260919`（`86f0e36653d25808ec2e6e150d43b42050d0eb88`），完整工作区 stash `09a7c5a6082856ff0dc09794166473be84c755e1` 保留。需要恢复时在独立目录执行以下命令，不覆盖当前工作区；最后两条恢复原有文档 intent-to-add 状态：

```sh
git worktree add -b codex/recover-pre-sync-20260919 /tmp/sub2api-pre-sync-20260919 codex/pre-upstream-sync-20260919
git -C /tmp/sub2api-pre-sync-20260919 stash apply --index 09a7c5a6082856ff0dc09794166473be84c755e1
git -C /tmp/sub2api-pre-sync-20260919 restore --staged -- docs/OPENAI_CODEX_TICKETS.md docs/upstream-sync-20260915-files.md
git -C /tmp/sub2api-pre-sync-20260919 add -f -N docs/OPENAI_CODEX_TICKETS.md docs/upstream-sync-20260915-files.md
```

## 2026-09-19 镜像发布（0.2.7 与按账号 292 开关）

- 基于当前完整工作区构建并推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.7-20260919-205646`，支持 `linux/amd64`、`linux/arm64`。版本 `0.2.7`，commit `1f0e0257a1aa-dirty`，构建时间 `2026-09-19T12:56:46Z`。
- 包含本日上游同步的 19 个提交、合并兼容修正、292 按账号启用与全局强制开关，以及此前全部本地定制；本轮未修改业务代码。
- 两个标签均指向 OCI index `sha256:2561a295134e33c0a7da413838b9262d0e6e32f6f86fb2ebdfa98122ac492cdb`。amd64 manifest 为 `sha256:8275a90906d7397fa29448de77d76396f1ee22004784fb965b40d82e8132c0c7`；arm64 manifest 为 `sha256:06f78c1c39937ea4db5190193f616bc8d66b8f90598fb68b433d2f8e46044034`。
- 验证：镜像内翻译完整性 3 项、类型检查、前端生产打包和双架构 Go 编译通过；构建推送退出 0。两个远端标签摘要与 metadata 一致。两个架构均按发布 manifest 从 Docker Hub 拉取，并在断网、只读、非 root 容器执行 `--version` 通过；版本、commit、构建时间一致。AMD64 首次 registry 拉取遇到 EOF，重试成功。保留既有 Browserslist/大 chunk 提示。
- 沿用本日源码合并后已通过的前端 2395 项测试、后端全量 unit 及定向 race 回归；本輪没有业务代码变化，不重复整套测试。构建与验证证据位于 `/tmp/sub2api-image-sync-20260919.LfGeIX/`。
- 本轮仅发布镜像，未部署、重启服务或修改生产数据；相较上一版无新增数据库迁移。上线时拉取新镜像并重建应用容器；版本检查不代表已完成线上账号调用验证。292 总开关开启且未开启全局强制时，只有明确启用的账号参与打票，使用变化见 [票据说明](OPENAI_CODEX_TICKETS.md)。
- 发布前 latest 为 `sha256:f2b72eda8f9cb958194e1b54c6995d55b2c0eac5ff3bfe955fff0e8efc183099`，固定标签 `0.2.5-20260918-224942`。需要回退时先关闭 292 总开关（旧版不支持按账号选择），再部署上一固定标签；恢复 latest 标签可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:f2b72eda8f9cb958194e1b54c6995d55b2c0eac5ff3bfe955fff0e8efc183099`。本轮未执行回滚。

## 2026-09-20 镜像发布（Dola 视频与按账号 Pro/Team 打票）

- 基于包含全部本地定制的当前工作区构建并推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.7-20260920-204354`，支持 `linux/amd64`、`linux/arm64`。版本保持 `0.2.7`，commit 为 `1f0e0257a1aa-dirty`，构建时间 `2026-09-20T12:43:54Z`。
- 包含 Dola ViralDance 2.0/2.5 视频转发、按账号选择 Pro 292 / Team 332，以及移除全局强制参与。部署后须总开关和账号开关同时开启才参与打票；历史 force_all 不再生效，原来仅靠强制参与的账号恢复普通转发。操作详见 [Codex 票据说明](OPENAI_CODEX_TICKETS.md)。
- 两个标签的 OCI index 均为 `sha256:bb7fa507c94188b849710a85f087efde97a9e972c385dfa3ffaddb899b6d706a`；amd64 manifest 为 `sha256:48fc5e181208d8b90b1ced5ca68ec30b746c80839f9dfe12e9287395ed739780`，arm64 manifest 为 `sha256:5d64d46e7c705257e0520880b7735f4b2fe3f789bfefb1a5ffad41cfbb3bd136`。
- 镜像内前端翻译检查 3 项、类型检查、生产构建及双架构 Go 编译通过，构建推送退出 0。远端两标签与构建 metadata 一致；两个架构均从 Docker Hub 按上述摘要拉取成功，并在断网、只读、非 root 容器执行 `--version` 通过，版本/commit/时间一致。
- 沿用本日 Dola 定向回归、票据/账号/接口契约定向 race 及前端 216 项回归、lint/生产构建验证。本轮无业务代码修改，未部署、重启服务或修改生产数据；版本执行检查不代表线上端到端验证。日志和 metadata 位于 `/tmp/sub2api-image-20260920.9Y4zOc/`。
- 发布前 latest 为 `sha256:2561a295134e33c0a7da413838b9262d0e6e32f6f86fb2ebdfa98122ac492cdb`，固定标签为 `0.2.7-20260919-205646`。回滚前应先关闭 Codex 打票总开关，避免旧版不支持 Team 类型及恢复历史全局强制。将 Compose 镜像改为上一固定标签后执行 `docker compose up -d --no-deps --force-recreate sub2api`；若需恢复 latest，执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:2561a295134e33c0a7da413838b9262d0e6e32f6f86fb2ebdfa98122ac492cdb`。本轮未执行回滚。

## 2026-09-20 镜像发布（API 密钥页费用提示优化）

- 已推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.7-20260920-225931`，包含 `linux/amd64`、`linux/arm64`。基于当前完整工作区，版本 0.2.7，commit `1f0e0257a1aa-dirty`，构建时间 `2026-09-20T14:59:31Z`。
- 本次新增 API 密钥页青绿费用提示卡片及「图片和视频的实际价格也是计费价格除6」说明；保留前次 Dola 视频、Pro/Team 打票类型和移除全局强制的修改。无新增计费逻辑或数据库迁移。
- 两标签 OCI index：`sha256:e0c314ee1112860166cf06896f42d3e6db758789bc2395783256b2c6d62014ed`；amd64 manifest：`sha256:3f97b2b08159b1d824e367902ce66f8448d67e84cb2f1af35ceab9653563df52`；arm64 manifest：`sha256:8b17a56637dc9e189cd942cc1c1669bffc52846cb0bc72437041942d42a08764`。
- 镜像内翻译检查、类型检查、前端生产构建及双架构 Go 编译通过，构建推送退出 0；远端两标签摘要与 metadata 一致。两个平台均从 Docker Hub 按 manifest 拉取成功，并通过断网、只读、非 root 容器的 `--version` 检查，版本、commit、构建时间一致。amd64 下载耗时较长，最终成功。
- 沿用本次 UI 修改已通过的 28 项相关测试、ESLint、生产构建及卡片浅色/深色、375px 视觉预览。本轮仅发布镜像，未部署、重启或修改生产数据；版本检查不等于线上端到端验证。日志与 metadata 位于 `/tmp/sub2api-image-keys-20260920.scqMAY/`。
- 上一固定标签 `iotwq/china-api:0.2.7-20260920-204354`，摘要 `sha256:bb7fa507c94188b849710a85f087efde97a9e972c385dfa3ffaddb899b6d706a`。回退可将 Compose 镜像改为该固定标签后执行 `docker compose up -d --no-deps --force-recreate sub2api`；恢复 latest 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:bb7fa507c94188b849710a85f087efde97a9e972c385dfa3ffaddb899b6d706a`。本轮未执行回滚。

## 2026-09-20 镜像发布（Dola 2.5 视频 30 秒支持）

- 已从当前完整工作区构建并推送 `iotwq/china-api:latest` 与 `iotwq/china-api:0.2.7-20260920-234609`，支持 `linux/amd64`、`linux/arm64`。版本 0.2.7，commit `1f0e0257a1aa-dirty`，构建时间 `2026-09-20T15:46:09Z`。
- 包含 Dola 2.5 整数 4–30 秒校验和接口文档更新；Dola 2.0 仍为 4–15 秒。保留既有本地定制，无新增数据库迁移。
- 两标签 OCI index：`sha256:3a8146091986290abc32bbbbed23b34befb062644c5a59ae259f899d7e244549`；amd64 manifest：`sha256:f29f3e45cfc82bed3fbb4a5bd62fd85254d57c6e74d7a483f71fdb4a70cd0af7`；arm64 manifest：`sha256:21f04af2b43ab822f484b7156ffd41374068148620a66734b5fc928fb728fde5`。
- 镜像内翻译检查、类型检查、前端生产构建和双架构 Go 编译通过。两个远端标签与 metadata 摘要一致，两个架构均从 Docker Hub 拉取并在断网、只读、非 root 容器执行 `--version` 通过。固定标签查询和 arm64 拉取首次遇到 EOF，重试成功；保留原始失败日志。
- 沿用紧邻本轮已通过的 Dola/Viralee 服务与接口定向 race 回归、前端 4 项测试、ESLint 和生产构建；没有重复修改业务代码。证据位于 `/tmp/sub2api-image-dola-duration-20260920.9Teos8/`。本轮未部署、重启服务或请求上游实际生成视频，版本检查不代表线上端到端验证。
- 上一固定标签为 `iotwq/china-api:0.2.7-20260920-225931`，摘要 `sha256:e0c314ee1112860166cf06896f42d3e6db758789bc2395783256b2c6d62014ed`。回退可将 Compose 镜像改为上一固定标签后执行 `docker compose up -d --no-deps --force-recreate sub2api`；恢复 latest 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:e0c314ee1112860166cf06896f42d3e6db758789bc2395783256b2c6d62014ed`。回退后 Dola 2.5 将恢复旧的 15 秒上限；本轮未执行回滚。

## 2026-09-24 同步原项目 0.2.8（9 月 23 日开始）

- 同步 `https://github.com/Wei-Shaw/sub2api.git` 的 `upstream/main` 至 `a3eb7ef302961cba716dc78b39b93b60c467db0e`，新增 237 个提交、473 个上游变更文件；本地合并提交 `f867e4f9bd2f6bcf1b2c6a67b20d5124cc86748a`，版本更新为 `0.2.8`。完整路径见 [文件清单](upstream-sync-20260923-files.md)。
- 包含 GPT-6 Sol/Luna、Claude Opus 5.5、Grok 4.7 支持，Claude Code 客户端版本自动同步，OpenCode Go 用量窗口，推理等级计费倍率，简易模式可选密钥窗口限制，以及流式结束、模型映射、图片兼容、账号调度、代理恢复、备份和界面交互修复。
- 保留本地 292/332 按账号打票及关闭全局强制、OAuth/TLS/并发保护、图片转换与结算补偿、Dola 2.5 的 4–30 秒、视频恢复退款、群聊私聊和监控。505 个原有脏文件无丢失，原暂存/未暂存分类保留；上游交集及本轮兼容修正外的内容哈希不变。未将原有定制整体提交进合并提交。
- 解决基线 19 个文件及定制恢复 10 个文件的内容冲突；账号 extra 查询同时保留打票私有状态和 OpenCode 用量，恢复 API Key 自定义模型测试选择，保留本地视频时长校验和原子结算。删除合并产生的重复插件装配函数并重新生成 Wire。
- 补齐推理转换兼容：保留 `max`，GPT-6 Sol/Luna 的 `none` 同时兼容平铺和嵌套参数，采样参数遵循最终推理等级；视频结算应用配置的推理倍率，但继续拒绝不带已验证媒体时长的通用计费调用。适配上游新测试与本地用量持久化返回值、构造参数及查询列数。
- 验证：前端 351 个文件、2652 项测试、lint、类型检查和生产构建通过；后端全量 unit、最终定向 race（含打票/OAuth/TLS/Dola/模型列表/计费/参数转换）通过。临时 PostgreSQL/Redis 的迁移幂等、推理定价迁移与读写、提现幂等、账号更新与批量更新集成通过。Wire 生成、embed 编译及 `--version`（0.2.8）通过。初次编译/回归失败及修复复测日志均保留于 `/tmp/sub2api-sync-20260923.rOMJ0l/`，未把失败轮次当作通过。
- 新增迁移：`238b_content_moderation_engine_meta.sql`、`239_channel_reasoning_effort_multipliers.sql`、`240_affiliate_ledger_operation_id.sql`。迁移按完整文件名及校验和登记，与本地同数字前缀的图片结算/视频审账迁移可以并存；不重命名已发布迁移。部署前备份数据库，应用启动会自动执行未应用迁移；旧 max 倍率迁入推理等级映射。源码回滚不等于数据库回滚，不建议迁移后直接使用旧应用处理已调整的定价数据。
- 本轮未构建或推送镜像、未推送 GitHub、未部署或修改线上数据，未实测生产上游。后续发布应使用含全部本地定制的完整工作区。
- 同步前分支 `codex/pre-upstream-sync-20260923` 指向 `1f0e0257a1aa1d12bc9fef28246ca80ab2f992d5`；完整工作区 stash `2daf8ddc0d2ff7aa5262b718c4d88156d696418f` 保留。恢复可在独立目录执行下列命令，不覆盖当前工作区（目录/分支已存在时另选空名称）：

```sh
git worktree add -b codex/recover-pre-sync-20260923 /tmp/sub2api-pre-sync-20260923 codex/pre-upstream-sync-20260923
git -C /tmp/sub2api-pre-sync-20260923 stash apply --index 2daf8ddc0d2ff7aa5262b718c4d88156d696418f
git -C /tmp/sub2api-pre-sync-20260923 restore --staged -- docs/OPENAI_CODEX_TICKETS.md docs/upstream-sync-20260915-files.md
git -C /tmp/sub2api-pre-sync-20260923 add -f -N docs/OPENAI_CODEX_TICKETS.md docs/upstream-sync-20260915-files.md
```

## 2026-09-28 同步原项目 0.2.9

- 经用户明确批准，合并 Wei-Shaw/sub2api 的 `upstream/main` 至 `9a62841fd124d026cf3694fcf9b79e98addcdbdc`；相对上次基线新增 70 个提交（37 个非合并提交）、117 个文件，本地合并提交为 `274d55cd6050f9c18479dd9de2609b026e9657c8`，版本更新至 `0.2.9`。完整来源见 [文件清单](upstream-sync-20260928-files.md)。
- 吸收 Responses 多代理 beta、消息类型及工具参数转换、WebSocket 上下文窗口切换、客户端取消 499 分类、账号额度暂停与重查退避、模型白名单通配符、长上下文账号统计、图片价格回退、CC Switch 地址、模型广场与分组弹窗修复。
- 保留本站 Basispoints Responses/Chat 桥接、图片中转、缓存计量选项、312 状态检测、OAuth/TLS/并发保护、图片结算补偿、视频恢复退款、Dola 时长、群聊/私聊、余额查询及品牌配置。613 个原有未提交条目均保留，220 个未跟踪条目的已有内容未改，原暂存分类恢复；上游交集和本轮明确列出的测试兼容外，无意外内容变更。
- 基线合并处理 KeysView 的一处冲突：保留 chinaapi 默认名称，采用上游修正后的 CC Switch 余额查询脚本。恢复定制时处理三个文件冲突：同时保留客户端取消和本地限流分类；Codex 模板继续不生成 `model_catalog_json`，不恢复用户已明确移除的设置。
- 计费差异明确保留：本站缺少价格时返回结算错误，不写成零元成功记录；因此上游 Free Fast 的缺价零元兜底不照搬，新增测试按本站既有保护适配。Free Fast 有正常价格时的 Standard 客户价格和 priority 上游统计仍保留。
- 图片价格变化：渠道图片输入/输出价格留空时继承目录价格，目录也无图片价格才按既有文本价格规则回退；显式填写 0 仍表示免费。需要免费图片输出的配置应显式填 0，不能继续依赖留空自动归零。此变化不补扣历史请求。
- 部署配置变化：三个 Compose 模板的 Redis 启动命令改为 exec 参数列表，避免 shell 换行及参数解析问题；初始化配置不再写入废弃 rate_limit 默认项。无新增数据库迁移、依赖或 Dockerfile 变化；既有配置不由本次同步自动重写。
- 验证通过：前端 354 个文件、2716 项测试，lint、翻译完整性、类型检查及生产构建；后端 `go test -mod=readonly -tags=unit ./...`；BPS/OAuth/并发/取消/WS/计费/媒体重点 race，以及 apicompat 和 basispoints 两个包的完整 race；embed 后端编译与 0.2.9 版本执行；Compose 三模板在空密码和普通测试密码下共六项离线解析校验。
- 首轮发现的 Gemini 取消测试缺少本站定价前置条件、现有 BPS 图片设置快照漏七个字段，以及 Free Fast 缺价策略冲突均已处理并复测通过；没有绕过真实计费校验或删掉失败测试。
- 数据库验证缺口：`CI=true go test -mod=readonly -tags=integration ./internal/repository -run 'Test(UsageBillingRepositoryApply.*|ImageSettlement_.*|OpsClientCancellationMetricsAndDuration|AccountRepo.*Basispoints.*|SchedulerCache.*)' -count=1 -timeout=6m -v` 因 Docker daemon 未启动而失败，未执行数据库用例，不能算通过。Docker 可用后补跑；测试 harness 只创建独立临时 PostgreSQL/Redis，不使用生产数据库。
- 本轮未构建或推送镜像、未推送 GitHub、未部署、未重启 Docker 或服务、未修改生产数据，也未进行真实付费上游请求。审计和构建证据位于 `/tmp/sub2api-sync-20260928.IGQa93`。
- 恢复点：分支 `codex/pre-upstream-sync-20260928`，原 HEAD `f867e4f9bd2f6bcf1b2c6a67b20d5124cc86748a`，完整 stash `fc12914844697fdf73fcae92f8b9ceb5225afab2`，同审计目录的 `before.tgz`、`index`、补丁和哈希清单。stash 不删除；归档包含被忽略文档，临时目录清理前应另行保管。以下命令可在新的独立目录恢复同步前完整工作区，不覆盖当前文件；若目标目录已存在，先另选空目录。本轮未执行回滚：

```sh
git worktree add --detach /tmp/sub2api-pre-sync-20260928 f867e4f9bd2f6bcf1b2c6a67b20d5124cc86748a
git -C /tmp/sub2api-pre-sync-20260928 stash apply --index fc12914844697fdf73fcae92f8b9ceb5225afab2
tar -xzf /tmp/sub2api-sync-20260928.IGQa93/before.tgz -C /tmp/sub2api-pre-sync-20260928
git -C /tmp/sub2api-pre-sync-20260928 add -f -N docs/OPENAI_CODEX_SIGNAL.md docs/OPENAI_CODEX_TICKETS.md docs/upstream-sync-20260915-files.md docs/upstream-sync-20260923-files.md
```

## 2026-09-28 独立同步 ranxi2001/sub2api v2.8.20 与 v2.9.0 的 Basispoints 变更

- 对照版本：`v2.8.20` 的 `dc01c71b758e8c24bd76ea5f7ddb089fc76481d3`，以及 `v2.9.0` 的 `b3e494dbdb90db789f7c06a1453b642892badbe2`、`211202d4a8abfe7d1b8d51f5edc8cda19ecdf50f`。
- 已同步 v2.8.20 的独立 BPS 限流修复：主请求、原生附件和加密恢复请求的 429 进入现有账号切换；工具纠正 429 只冷却当前 BPS 路由；Retry-After 解析、进程内 BPS 专用冷却、三条选号路径跳过冷却账号、候选耗尽时安全 429 及脱敏响应均已适配。
- 本地仍保留现有 OAuth/Setup Token 账号、代理、并发、计费、Responses/Chat 桥接和图片附件边界；BPS 冷却不写 Codex 全局额度、账号健康状态、全局限流时间或用量账单。
- v2.9.0 的质量规则自动开启/关闭 BPS 未直接移植。它依赖当前仓库不存在的 `account_quality_bps` repository/service、Pelican 质量规则/探针计划、迁移和管理员前端；Mihomo 代理托管依赖也按既有范围另行处理。当前账号编辑中的手动 Basispoints 选择保持不变。
- 相关代码、测试和来源说明见 `backend/internal/service/openai_basispoints_ratelimit.go`、`openai_basispoints_forward.go`、`openai_basispoints_attachments.go`、`openai_basispoints_repair.go`、`backend/internal/handler/openai_gateway_handler.go`、`docs/OPENAI_OAUTH_BASISPOINTS.md` 和 `backend/internal/service/basispoints/NOTICE.md`。这些文件中的 `Excel BPS` 是本站既有兼容命名，不代表额外引入 Excel 模式。
- 本次仅修改本地源码/测试/文档，没有构建或推送镜像、部署服务、修改数据库或请求真实 OAuth 上游。完整 service/handler 单元测试和 BPS 定向回归作为发布前置验证；Docker 数据库集成仍以当前环境可用性为准。

## 2026-09-29 筛选适配 ranxi2001/sub2api v2.9.1–v2.9.4 的 BPS 修复

- 固定发布提交：v2.9.1 ca5dd3cf5299df8778cdfa752430253ad4e071e8；v2.9.2 cdab8bbb4e799bb4c85f8f87f4a2e88accd4256a；v2.9.3 faf58e440b1bddb07429f74ed63b570c11d1c0f8；v2.9.4 7dd10bfe4b635f226f0ddfa52cc65797697272d8。
- 已移植 f6666ab43（原生附件字段）、e0ba95a48（可见流式内容后禁止整体再生成）、b6617bf62（流内错误分类/取消终态）。本站额外适配现有 Chat Completions 桥接，保留缓存用量、无用量失败不记账、真实 HTTP 429/5xx 既有重试及 Codex 路由。
- a6e51622a 的 BPS 专用池和初始准入、0c74d2d7a 的 Mihomo 代理获取层在本站不存在，不复制无效实现；混合候选池在既有 TopK 预算内的容量回退以本地回归确认。自动恢复/质量体系依赖的 4930871f6、9ec3c0422、4a8faa3f8 未移植。新生图、图片策略/扩容、自动模板、WS 加速、Free 套餐禁止等新增功能或策略不混入本轮。
- 这是源码按行为适配，不是整仓 Git merge/cherry-pick；版本保持 0.2.9，HEAD 274d55cd6050f9c18479dd9de2609b026e9657c8 和原 index 不变。Wei-Shaw 0.2.10 仍未同步。
- 测试及逐项回滚说明见 progress.md 的同日三项修复和最终校验记录。施工前快照 /tmp/sub2api-bps-fixes-20260929.JAlgyz/before.tgz 包含全部受影响既有源码及文档；只允许按文件回退本次 hunk，不能整仓覆盖已有定制。未发布镜像或部署。
- 最终 service/handler/basispoints/apicompat 四包完整 unit、协议完整 race、BPS/并发保护及 handler 切换定向 race 均通过。混合候选池和取消用量回归重复 20 次通过；实际回归源码与最后测试状态一致。前端、数据库结构和部署配置没有改动，未执行前端构建或数据库迁移。

## 2026-09-30 同步原项目 0.2.10

- 经用户确认，合并 Wei-Shaw/sub2api 的 upstream/main 至 `a60a29549f488a854966aaec9541abbe006cac22`，版本更新为 `0.2.10`；新增 33 个提交、118 个上游文件，本地合并提交 `59fcda4ea34f087600114c74f31806df6160fc97`。来源及逐文件说明见 [合并清单](upstream-sync-20260930-files.md)。
- 包含 Chat/Responses 流式用量与缓存归一化、WebSocket 合成分组路由和模型所有权校验、白名单与映射冲突处理、风控用户白名单、Sonnet 5.5 协议与定价、Claude 重置额度查询、消费趋势切换和客户端配置改进。
- 两处装配冲突保留上游 Claude 重置额度查询与本站现有服务；工作区不恢复已移除的 292/332 打票。保留本站 BPS Responses/Chat、缓存计量、附件、429 冷却与账号切换，以及 OAuth/TLS/并发保护、图片转换和视频结算退款。未修改生产数据库、余额或已有用量。
- 本地原 621 项工作区状态和暂存/未暂存分类已恢复；原定制没有整体纳入合并提交。对 4486 个快照路径进行比对，包含原有 4465 个文件及 21 个缺失状态；仅上游范围、回归测试修正和本轮追加文档允许变更，其余内容与权限不变。备份保留，不删除 stash。
- 回归额外定位到既有风控测试在 worker 启动后改写 TTL 的数据竞态；只改测试为复用无 worker 的缓存装配，不变更线上并发或风控逻辑。该用例 race 连续 50 次及最终定向回归均通过。
- 验证通过：前端 355 文件、2734 项测试，lint、翻译完整性、类型检查和生产构建；后端全量 unit；service/handler/repository 的 BPS、OAuth 保护、合成路由、风控、计费与 WS 定向 race；basispoints/apicompat/tlsfingerprint/openai_ws_v2 四包完整 race；Wire diff 无差异；embed 生产编译及 0.2.10 版本检查。前端保留既有大 chunk 提示，未调整依赖。
- 隔离 PostgreSQL 18.1 / Redis 8.4 集成通过：用量趋势与聚合、用量写入与取消、计费原子性和幂等、图片超预估/释放后重试、视频审账，以及 BPS 403 移组事务。初轮把计费与全局统计套件放在同一测试库导致两条全局统计断言受其他用例写入影响；按现有 harness 分成两次独立临时数据库执行后，所有选定用例通过。保留初次失败日志，不宣称混跑或全部数据库测试通过，未修改业务 SQL 迎合断言。
- 本轮无新增数据库迁移、依赖或 Dockerfile 改动；未构建/推送 Docker 镜像、未推送 Git、未部署或调用真实付费上游。现有镜像不随本地合并自动更新；后续发布应从包含本地定制的完整工作区构建。
- 证据目录 `/tmp/sub2api-sync-20260929.6L1lvp`；目录名沿用 UTC 日期，本节采用 Asia/Shanghai 日期。原 HEAD 为 `274d55cd6050f9c18479dd9de2609b026e9657c8`，恢复分支 `codex/pre-upstream-sync-20260929-a60a29549`，stash `9929e1441d046dfbf1fa7650d04e4856b6fdb630`；归档 `before.tgz` SHA256 为 `6e3ab07f6589581ae924aaf4f9833e36e176773d8ad81f6bc6b584010034ac8c`。临时目录清理前应另行保管。
- 以下恢复步骤仅在一个不存在的新目录重建合并前工作区，不覆盖当前文件，不删除数据库；目录已存在时须另选空目录。本轮未执行回滚。原五个 intent-to-add 文件在最后恢复索引标记：

```sh
git worktree add --detach /tmp/sub2api-pre-sync-20260930 codex/pre-upstream-sync-20260929-a60a29549
git -C /tmp/sub2api-pre-sync-20260930 stash apply --index 9929e1441d046dfbf1fa7650d04e4856b6fdb630
tar -xzf /tmp/sub2api-sync-20260929.6L1lvp/before.tgz -C /tmp/sub2api-pre-sync-20260930
git -C /tmp/sub2api-pre-sync-20260930 add -f -N docs/OPENAI_CODEX_SIGNAL.md docs/OPENAI_CODEX_TICKETS.md docs/upstream-sync-20260915-files.md docs/upstream-sync-20260923-files.md docs/upstream-sync-20260928-files.md
```

## 2026-10-01 同步原项目 0.2.11

- 经用户要求，合并 Wei-Shaw/sub2api 的 upstream/main 至 `42bc7f6cf`，版本更新为 `0.2.11`；相对上次同步基线新增 23 个提交、90 个上游变更路径，本地合并提交为 `365b2e6547975a27c841e9887d0abda2fcb76cf3`。完整来源见 [合并清单](upstream-sync-20261001-files.md)。
- 吸收余额在途预冻结、预冻结续期与计费落账衔接、未定价别名估算、Claude 原生额度兑换、当前订阅计划识别、GPT-6.1 Sol、Astra Ultrafast、Codex 远程模型目录、API Key 创建限制及 Claude Code 兼容入口降级分组。
- 保留本站 public asset/community 装配、mandatory usage record fallback、Basispoints Responses/Chat、OAuth/TLS/并发保护、图片和视频结算、余额冻结及既有本地路由。没有恢复已移除的 292/332 打票逻辑，也没有引入新的数据库迁移。
- 合并冲突处理：`wire_gen.go` 同时保留上游 idempotency/Claude reset redeem 与本站装配；网关处理器保留 mandatory usage record fallback；Codex 配置默认只使用远程模型目录，不生成已移除的 `model_catalog_json`。另外让已有数据库图片冻结不再重复执行 Redis 在途预占，避免同一图片请求被双重占用余额。
- 验证：前端 355 个测试文件、2777 项断言、lint、生产构建通过；后端 embed 编译、Wire diff 及跳过环境敏感堆内存阈值用例的全量 unit 通过。该内存测试单独连续 3 次通过，但嵌入全量 suite 时两次受堆基线波动影响；不能据此宣称全量 unit 无条件通过。PostgreSQL/Redis 隔离集成中 BPS/账号/用量日志套件通过，计费/图片结算首轮同组运行出现时间窗口断言波动，单独重跑及同组重跑均通过。
- 本轮未构建或推送 Docker 镜像、未推送 GitHub、未部署、未修改生产数据或调用真实付费上游。恢复分支为 `codex/pre-upstream-sync-20260930-v0.2.11`，stash 为 `2cd1d88310fbc76fb411f30d3c3988ff6cdff087`，完整备份位于 `/tmp/sub2api-sync-20260930.S4H93B/before.tgz`。

## 2026-10-02 同步原项目 0.2.12

- 按用户要求合并 Wei-Shaw/sub2api 的 upstream/main 至 `ae501cc22`，版本更新为 `0.2.12`；相对 `42bc7f6cf` 新增 32 个提交（21 个非合并提交）、152 个上游文件，本地合并提交为 `8d433e144`。来源、逐文件用途及恢复命令见 [合并清单](upstream-sync-20261002-files.md)。
- 包含 TypeSafe Jev System One 的 API Key 接入与原生端点、充值赠金/折扣阶梯、邮箱验证码原子尝试次数和单次哈希重置 token、匿名订单查询限流、Antigravity 错误脱敏、Grok CLI 身份头修复、账号优先级快速调整和 API Key 分组排序。Axios 更新到 1.20.0。
- 账号创建基线冲突兼容两侧校验，定价测试冲突保留 Jev/Grok 全部用例。保留本站 Basispoints、OAuth/TLS/并发保护、312 信号观察、图片/视频结算与用量任务同步兜底、社区和余额查询；Codex 配置仍不生成 model_catalog_json，未恢复 292/332 采集。
- 完整工作区已备份；原 618 项状态及暂存分类恢复。对合并前 4502 个路径逐项核对，上游范围外只有本轮测试和文档的明确变更，无其他内容/权限变化；原有未提交定制没有整体纳入合并提交。
- 首轮前端三个失败是新增 TypeSafe 后旧测试按五个平台断言，已更新为完整六平台配额对象并复测。最终前端 357 个文件、2803 项测试通过，lint、翻译完整性、类型检查和生产构建通过；后端完整 unit（无跳过）、Wire diff、embed 编译与 0.2.12 版本检查通过。
- BPS/OAuth 并发、312 信号、在途预占、媒体结算、充值优惠、邮箱验证、System One、Grok 和公开订单接口的定向 race 通过。隔离 PostgreSQL 18.1/Redis 8.4 中迁移幂等/并发、赠金字段、TypeSafe 约束、平台配额和 BPS 状态验证通过；另一个独立测试库中计费原子性/幂等、图片超预估结算和视频审账通过。
- 新增 `241_add_payment_order_bonus_amount.sql` 和 `241_add_typesafe_platform.sql`，按完整名称登记，不改已发布迁移名称。应用启动会自动执行未应用迁移，发布前需备份数据库。空优惠阶梯保持原充值金额；旧未使用的明文密码重置链接在升级后需要重新申请。生产数据库迁移回退应单独评估，源码回退不等于数据库回退。
- 本轮未构建/推送 Docker 镜像、未推送 GitHub、未部署、未修改生产数据库或调用真实付费上游。证据目录 `/tmp/sub2api-sync-20261002.DEsw9a`；恢复分支 `codex/pre-upstream-sync-20261002-v0.2.12`，stash `87dc77358c35808872863e42f720311ab6999b56`，完整归档 `before.tgz` 保留。
- 最终核对期间上游有新提交，追加同步至 `3040209f205472038c1ba745a1bedd2edd9053b1`，本地最终合并提交 `eda84f38231469c4968a8e574a1c26a221ce547d`；本轮合计 38 个提交、157 个上游路径，版本仍为 0.2.12。增量修复结算期间删除 API Key 导致回滚的漏扣风险，并补齐 TypeSafe 中转账单探测。已将数据库回归适配本站原子用量/账务写入，确认重复结算只扣一次；没有引入额外迁移或前端变化。增量恢复点为分支 `codex/pre-upstream-sync-20261002-followup`、stash `8a6573d9dfb3119fefbe437b33b0526687d1a740`，整个本轮回退仍使用首次恢复点。
- 增量后 repository/service 完整 unit、计费/媒体/账单探测定向 race、隔离数据库计费/媒体最终回归、embed 编译均通过。最终 4502 路径校验和 618 项原工作区状态核对无意外变化，格式和冲突检查通过；前端沿用本轮 2803 项成功验证。

## 2026-10-02 发布 0.2.12 双架构镜像

- 从当前完整工作区（HEAD `eda84f38231469c4968a8e574a1c26a221ce547d` 及既有本地定制）构建并推送 `iotwq/china-api:latest` 和固定标签 `iotwq/china-api:0.2.12-20261002-210029`，包含上述主项目同步和删除 API Key 后继续结算的计费修复；本次未再次同步远程代码或修改业务源码。
- 两个标签的 OCI index 摘要均为 `sha256:7691894c39a2c1a684caaa70d904e0df7e40feedf19397a7071645f7b68fad90`；`linux/amd64` manifest 为 `sha256:cb5d51f950e5e5e9ed1eaa2725ddd85ba933fcbd2a4ecf70b576d1a94b88beca`，`linux/arm64` manifest 为 `sha256:574bbadd2a8aba59d0dfa056e37ef150274648a92b033eb1cc08f215cf1916e9`。`unknown/unknown` 项是构建证明，不是额外运行架构。
- 发布前再次通过隔离 PostgreSQL/Redis 的删除 Key 原子结算和重复结算幂等回归。镜像内翻译完整性 3 项、类型检查、Vite 生产构建和双架构 Go embed 编译通过；两个远端标签摘要与 build metadata 一致。分别按子 manifest 从 Docker Hub 拉取，在断网、只读、非 root 容器执行 `--version` 均通过，报告 `0.2.12`、`eda84f382314-dirty` 和构建时间 `2026-10-02T13:00:29Z`。
- 本轮只发布镜像，未部署、未执行生产数据库迁移、未改线上余额或用量。部署前备份数据库；应用启动会执行源码包含的未应用迁移（本次同步新增的两个 `241_*.sql` 见上一节），源码/镜像回退不等于数据库回退。
- 发布前 `latest` 为 `sha256:2c57dab4720feaed7b298ff711eef94f6761722a98822323db1f3f69c1dc85db`，固定标签 `iotwq/china-api:0.2.9-20260930-004733`。必要时将部署镜像固定为该标签或旧索引摘要，再执行 `docker compose pull sub2api` 和 `docker compose up -d --no-deps --force-recreate sub2api`；恢复远端 `latest` 可执行 `docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:2c57dab4720feaed7b298ff711eef94f6761722a98822323db1f3f69c1dc85db`。旧镜像不包含本次删除 Key 计费修复，已升级的生产数据库需单独评估兼容；本轮未执行回滚。
- 构建和验证证据在 `/tmp/sub2api-publish-20261002.YDsNp3`。初次远端查询和第一次构建遇到 Docker Hub EOF，重试成功；保留失败构建日志。前端保留既有 Browserslist 和大 chunk 提示，未调整依赖。

## 2026-10-02 同步 0.2.13 版本标识

- 官方 `v0.2.13` 为附注标签，对象 `7d0c0067f406c380f0a94cfc3879cdae7049b467`，实际指向提交 `3040209f205472038c1ba745a1bedd2edd9053b1`。该提交已在前轮合并中包含，且已进入当天发布的 `0.2.12-20261002-210029` 镜像；此前显示 0.2.12 是因为上游尚未提交 VERSION 更新，并非遗漏 0.2.13 的业务代码。
- 本轮主项目仅新增 `b8dece9000c68815a5b867ca5a1e6f236e173905`（`chore: sync VERSION to 0.2.13 [skip ci]`），仅将 `backend/cmd/server/VERSION` 从 0.2.12 改为 0.2.13。已合并为 `e895e0cc587f2c6be794977e240e32ee6620fdd7`，没有新增功能、协议、计费或数据库迁移变更。
- 使用独立临时 Git index 完成合并，再只同步正式 index 的 VERSION 项；未将本地定制纳入合并提交。合并后对原 4518 个路径逐项校验，只有 VERSION 内容变化；全部原工作区状态和暂存分类一致。
- 版本解析脚本返回 0.2.13；后端 `go build -mod=readonly -tags embed` 通过，生成程序的 `--version` 返回 0.2.13。业务代码未变化，沿用前轮功能回归，不重复运行全量测试。
- 本轮未制作或推送新镜像、未部署。Docker Hub latest 仍为上一节的已验证镜像，内含官方 v0.2.13 标签代码，显示版本仍为 0.2.12。
- 恢复点：`codex/pre-upstream-sync-20261002-v0.2.13` 指向 `eda84f38231469c4968a8e574a1c26a221ce547d`；临时目录 `/tmp/sub2api-sync-20261002-v0213.LiV49S` 保留原 index、暂存/未暂存补丁、VERSION 和全路径摘要。若仅需回退此次版本标识，执行 `git revert -m 1 e895e0cc587f2c6be794977e240e32ee6620fdd7 --no-commit`，只会撤销本轮 VERSION 变化；有重叠修改时先核对，勿整体覆盖工作区。本轮未执行回滚。

## 2026-10-03 发布 0.2.13 双架构镜像（图片模型路由）

- 从当前完整工作区（含未提交定制）构建并推送 `iotwq/china-api:latest`、`iotwq/china-api:0.2.13-20261003-032215`。包含刚补齐的七个 Midjourney、Seedream、Nano Banana 完整图片模型名称路由；沿用 OpenAI API Key 图片协议，未新增供应商专用协议。
- 两标签索引摘要均为 `sha256:9791acb5fa4531543ae2c3670f0e6a6732786f648caa7c00f893652661296b79`；linux/amd64 为 `sha256:940376c79375a6234440af06d9125e5366d8886c14217c66cdf1997726f2849d`，linux/arm64 为 `sha256:9cedfebd7e0bdc472a9ba56821d298dd40a496dd7d82b35217e620ee3b9337f3`。
- 版本 `0.2.13`，构建标识 `e895e0cc587f-dirty`，构建时间 `2026-10-02T19:22:15Z`；固定标签日期使用 Asia/Shanghai。镜像内翻译完整性 3 项、类型检查、前端生产打包和两个架构 Go embed 编译均通过。
- 直连 Docker Hub 的构建及拉取出现超时；使用本机已有代理和独立临时构建器完成推送，未修改 Dockerfile 或全局代理。远端两标签、平台及 metadata 摘要核对一致；按远端子摘要通过 Buildx 导入本地后，两个架构均在断网、只读、非 root 条件下执行 `--version` 成功。普通 `docker pull` 的网络超时不记为拉取成功。
- 证据目录 `/tmp/sub2api-publish-20261003.s5RZtc` 保存失败尝试、成功构建、metadata、远端核验、按摘要导入及版本检查日志。前轮图片路由模拟上游回归仍有效，本次未重跑全量业务测试或进行真实付费生图。
- 本轮未部署、未推送 GitHub、未改业务源码或生产数据；相对上一镜像没有新增迁移。服务更新后仍需配置相应模型白名单/映射与价格。
- 发布前已核实旧 latest 为 `iotwq/china-api:0.2.12-20261002-210029`，摘要 `sha256:7691894c39a2c1a684caaa70d904e0df7e40feedf19397a7071645f7b68fad90`。需要恢复远端 latest 时可执行下述命令；部署回退可将 Compose 镜像固定为旧摘要，再执行 `docker compose pull sub2api` 和 `docker compose up -d --no-deps --force-recreate sub2api`。旧镜像不含本轮新增模型路由，本次未执行回滚。

```sh
docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:7691894c39a2c1a684caaa70d904e0df7e40feedf19397a7071645f7b68fad90
```

## 2026-10-04 发布 0.2.13 双架构镜像（接口文档与 Seedream X）

- 从当前完整工作区构建并推送 `iotwq/china-api:latest` 与固定标签 `iotwq/china-api:0.2.13-20261004-190735`，包含 `seedream-5.0-pro-x` 路由支持、三个图片模型的用户接口文档及全部既有本地定制。
- 两标签索引摘要均为 `sha256:04607a3bb87893fc19756175aa07220081cc76ae291ace70fc182835da07c6e7`；linux/amd64 为 `sha256:9e4a2b65059390224a55737f74f18e66ed8b5761f7355113aeef683f33cee33a`，linux/arm64 为 `sha256:eaf34efc5818118433d0f596aa6329c0f8db64db5f7c2e33b6027d4f7c48bd91`。
- 版本 `0.2.13`，构建标识 `e895e0cc587f-dirty`，构建时间 `2026-10-04T11:07:35Z`。前端国际化检查、vue-tsc、Vite 生产构建和两个架构 Go embed 编译通过。
- 直连 Docker Hub 构建器可能受网络超时影响，本轮使用本机代理和临时多架构构建器完成推送；未修改 Dockerfile、全局代理或生产配置。两个远端标签与 build metadata、平台清单一致；按子摘要导入后，两个架构均在断网、只读、非 root 条件下执行 `--version` 成功。
- 证据目录 `/tmp/sub2api-publish-20261004.fQl3Qf` 保存构建、metadata、远端清单、导入和版本检查日志。本轮未重新执行全量业务测试或真实付费生图，沿用前轮已通过的图片路由和文档定向测试。
- 本轮未部署、未推送 GitHub、未修改生产数据库、余额或用量，也没有新增数据库迁移。
- 发布前 `latest` 摘要为 `sha256:9791acb5fa4531543ae2c3670f0e6a6732786f648caa7c00f893652661296b79`。需要回滚远端 latest 时执行：

```sh
docker buildx imagetools create --tag iotwq/china-api:latest iotwq/china-api@sha256:9791acb5fa4531543ae2c3670f0e6a6732786f648caa7c00f893652661296b79
```
