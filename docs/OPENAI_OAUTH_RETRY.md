# OpenAI OAuth 临时错误自动重试

OpenAI OAuth 和 Setup Token 的 HTTP Responses 转发及复用账号错误分类的兼容路径支持临时错误自动重试，无需开启池模式或重新添加账号。部署包含本修复的新构建后生效。

## 处理规则

- HTTP 500、502、503、504、520 至 524，以及现有分类器识别的临时处理失败/容量过载错误，先使用现有同账号重试预算。普通 5xx 最多额外重试 3 次，即同一账号最多请求 4 次。
- 重试仍失败时，按已有账号切换上限寻找其他可用账号；没有可用账号或切换次数耗尽后，才返回最终错误。
- 瞬时 429 保留现有最长约 2 分钟的恢复窗口与等待策略，不改成固定 3 次。明确配额耗尽的 429 仍执行配额冷却，不在耗尽账号上反复重试。
- 鉴权失效、账号/工作区停用、管理员配置的暂停调度规则继续生效。请求参数、上下文长度和安全策略错误不会因本修复被转为临时重试。
- 生图接口的 `image_generation_unavailable`（包含只有文字、没有图片的情况）保留原有直接切号行为，不把它产生的本地 502 当作上游临时 5xx。
- 客户端断开或已经收到实际回复内容后，仍受现有取消及禁止重放保护，不会为重试复制已输出内容。原生 WebSocket 的重连策略未修改。

## 与池模式的区别

API Key/Bedrock 池模式用于对接上游账号池，会跳过部分默认账号状态处理。OAuth 代表真实凭据，不能为复用重试而跳过其鉴权和配额维护。本次只复用请求级重试流程，不扩展 `IsPoolMode()`，不新增 OAuth 池模式开关，Gemini 等其他渠道不变。

## 验收

在 `backend` 目录运行：

```sh
go test -tags unit ./internal/service ./internal/handler -run 'Test(IsPoolModeScope|OpenAIOAuthPassthrough|OpenAIOAuthTransientRetryClassification|OpenAIResponses_OAuthTransientRetries)' -count=1
```

测试包含上游先失败后成功、3 次重试耗尽、切换健康账号、配额与鉴权边界。部署后可通过 `openai.pool_mode_same_account_retry` 日志核对重试账号和次数；这个日志名称来自共用流程，不表示 OAuth 启用了池模式。

重试只能吸收短暂故障，不能保证上游持续不可用时仍成功；它也会增加请求等待时间。服务日志和用量记录应结合最终结果核对，不要把一次上游失败尝试当成客户端最终失败。
