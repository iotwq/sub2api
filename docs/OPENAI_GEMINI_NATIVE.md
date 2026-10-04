# OpenAI 账号转发 Gemini 原生接口

## 适用场景

当同一个兼容上游同时提供 OpenAI API 和 Gemini 原生 `v1beta` API 时，可以继续把账号保存为 OpenAI APIKey 账号，并为该账号单独开启 Gemini 原生接口能力。

该能力支持：

- `GET /v1beta/models`
- `GET /v1beta/models/{model}`
- `POST /v1beta/models/{model}:generateContent`
- `POST /v1beta/models/{model}:streamGenerateContent?alt=sse`
- `POST /v1beta/models/{model}:countTokens`

## 账号配置

在管理后台创建或编辑 OpenAI APIKey 账号：

1. 平台选择 `OpenAI`，类型选择 `APIKey`。
2. Base URL 填写兼容上游根地址，例如 `https://sub.g-aisc.com`。已有的 `/v1` 结尾地址也兼容。
3. API Key 填写上游密钥。
4. 在“端点能力”中开启“Gemini 原生接口”。该能力默认关闭。
5. 如果账号配置了模型白名单或模型映射，加入 `gemini-3-pro-image-preview` 等需要调用的模型。

开启后，Gemini 原生请求仍在原 OpenAI 分组内调度，并使用以下认证方式访问上游：

```http
Authorization: Bearer <ACCOUNT_API_KEY>
Content-Type: application/json
```

请求 body 和 SSE 响应按 Gemini 原生协议透传。账号级模型映射只修改 URL path 中的模型名。

## 行为边界

- 未开启该能力的 OpenAI 账号不会被 Gemini 原生入口调度。
- OpenAI OAuth 账号不支持该能力。
- 原有 `/v1/responses`、Chat Completions、Embeddings、图片及视频调用不受影响。
- Gemini 平台账号继续使用原有 Gemini 鉴权与转发逻辑。
- 使用记录归属原 OpenAI 分组和账号，上游端点记录为 `/v1beta/models`。
- 该能力不负责把 OpenAI 请求格式转换成 Gemini 请求格式，调用方必须直接使用 Gemini 原生接口。

## 调度说明

- `openai_capabilities` 会写入不含访问令牌的调度元数据，开启负载批量调度时也能正确识别 `gemini_native`。
- 保持 `gateway.scheduling.load_batch_enabled` 为默认开启状态即可，不需要为了使用 Gemini 原生接口关闭负载调度。
- 账号能力编辑后会刷新单账号调度快照；服务启动时也会执行全量快照重建。

## 验证示例

```bash
curl "https://<SUB2API_HOST>/v1beta/models/gemini-3-pro-image-preview:generateContent" \
  -H "Authorization: Bearer <SUB2API_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{
      "role": "user",
      "parts": [{"text": "生成一张未来城市图片"}]
    }],
    "generationConfig": {
      "responseModalities": ["TEXT", "IMAGE"]
    }
  }'
```
