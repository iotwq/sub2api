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
