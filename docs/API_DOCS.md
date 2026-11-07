# 用户接口文档页

用户侧“接口文档”位于左侧栏，页面路径为 `/api-docs`。页面只展示调用接口所需的模型、参数、示例和必要限制，所有示例统一使用：

```http
Authorization: Bearer YOUR_API_KEY
```

### 使用密钥生成的 OpenAI Codex 配置

在 OpenAI 分组的“使用密钥”弹窗中，普通 Codex CLI 和 Codex WebSocket 配置默认使用 `gpt-6.1-sol`。如果已同步的远程模型目录不包含该模型，配置会回退到目录中的第一个可用模型；其他分组的默认模型不受影响。

## 页面导航

页面不再把所有模型从上到下排开，而是用标签栏切换，一次只显示一个模型组：

- 第一行是分类：图片生成、视频生成、音乐生成。分类旁的数字是该分类下可调用的模型名称数量（各模型组“可用模型”表的行数之和），手机宽度下不显示数字。
- 第二行是当前分类下的模型组：图片生成为 gpt-image2、Seedream、Nano Banana、Grok；视频生成为 SD2.0、Wan3.0、MiniMax-H3、Grok；音乐生成为 Suno。切换分类时默认显示该分类的第一个模型组。
- 页面上 Grok 分成“Grok 图片模型”和“Grok 视频模型”两组，分别放在图片和视频分类下，参数、示例和说明按图片、视频拆开，内容本身不变。本文下方的“Grok 图片视频模型”一节仍合并描述。
- 宽度 640px 及以上时标签栏固定在文档区顶部；滚到下方再切换模型，内容会回到标签栏下方的起始位置。手机宽度下标签栏随页面滚动，避免占用约三分之一的屏幕。
- 默认显示 OpenAI gpt-image2。当前选中的模型组不写入网址，刷新页面后回到默认组。未选中的模型组被隐藏，浏览器的页内查找只能找到当前显示的模型组内容。
- 新增模型组时，需要在 `frontend/src/views/user/ApiDocsView.vue` 的 `modelDocs` 中为它填写 `category`（`image`、`video` 或 `music`）和标签栏显示的简称 `tabLabel`。

## 模型分类

### OpenAI gpt-image2

- 支持 `gpt-image-2`、`gpt-image-2.5-sunburst`、`gpt-image-2.5-flare`、`gpt-image-2.6-sunburst`、`gpt-image-2.6-flare`。五个模型的调用方式一致：文生图使用 JSON `POST /v1/images/generations`，参考图和遮罩编辑使用 multipart `POST /v1/images/edits`。
- 五个模型的规格统一展示为「1K / 2K / 4K；最多 16 张参考图」；像素尺寸例如 `1280x1280`、`2048x1152`、`3840x2160`。`quality` 差异只在参数说明中列出：`gpt-image-2` 可选 `low`、`medium`、`high`；两个 2.5 模型和两个 2.6 模型额外支持 `xhigh`、`max`。
- 2026-11-07 合并上游 0.2.20 后，2.6 两个模型已接入原生图片链路：OAuth 使用直调图片端点，不回退到无法保留 2.6 型号的 Responses 图片工具链路；API Key 账号沿用原生 Images 转发。后端默认模型列表、前端模型白名单和静态价格目录已补齐 2.6，远程价格缺失时采用 2.5/2.6 的兜底价格，不再落到旧 `gpt-image-2` 价格。本次通过路由、参数、定价回归测试及生产编译，尚未验证真实付费上游生成。
- 需要透明背景时，在 JSON 或 multipart 表单中传 `background: "transparent"`，并同时传 `output_format: "png"`（multipart 写作 `background=transparent`、`output_format=png`）。
- 输出格式可选 PNG、JPEG、WebP。`output_compression` 仅用于 JPEG、WebP，PNG 不要传该字段；审核参数可选 `auto`、`low`。
- `response_format` 可选 `b64_json`、`url`，用于声明期望返回格式。不同上游可能返回 `data[].b64_json`、`data[].url`，或同时返回两者，客户端需要兼容判断。
- 图片编辑最多传 16 张参考图，多图时重复使用 multipart 的 `image` 字段。`mask` 为可选 PNG 遮罩，对应第一张参考图。
- 用户页面提供 JavaScript 解析示例，分别处理纯 Base64 `data[].b64_json`、`data[].url` 内的 data URL，以及 `data[].url` 内需要下载的 HTTP(S) 地址。

透明背景文生图示例：

```json
{
  "model": "gpt-image-2.5-sunburst",
  "prompt": "一只白色暹罗猫，透明背景",
  "quality": "xhigh",
  "background": "transparent",
  "output_format": "png"
}
```

将 `model` 替换为 `gpt-image-2`、`gpt-image-2.5-flare`、`gpt-image-2.6-sunburst` 或 `gpt-image-2.6-flare` 即可使用相同接口；`gpt-image-2` 不支持 `xhigh`、`max`，请使用 `low`、`medium` 或 `high`。

### OpenAI 兼容 Seedream 图片模型

- 用户页面顶部标签使用 `Seedream`，主标题保留 `OpenAI 兼容 Seedream`；不再展示账号类型/模型转发说明及 GPT 专属参数限制两条提示。
- 用户侧展示 `seedream-5.0-pro`、`seedream-5.0-pro-x`、`seedream-5.0-lite-x`，使用本站 `POST /v1/images/generations` 文生图和 `POST /v1/images/edits` 参考图编辑接口。
- Pro X / Lite X 的参数依据本次提供的 G 渠道接口说明更新；原有 `seedream-5.0-pro` 的 8000 字符说明保持不变。
- 用户页面及调用示例仅使用公开模型 ID，不展示上游型号。管理员需在现有账号或渠道模型映射中配置：`seedream-5.0-pro-x` → `seedream-5.0-pro(G)`，`seedream-5.0-lite-x` → `seedream-5.0-lite(G)`。代码允许两个公开别名进入图片链路，但不自动覆盖账号映射或价格。

| 用户模型 ID | 分辨率 | 提示词上限 | 参考图上限 | 比例 |
| --- | --- | --- | --- | --- |
| `seedream-5.0-pro` | 1K / 2K | 8000 字符 | 10 张 | 16:9、9:16、1:1、4:3、3:4、3:2、2:3、21:9 |
| `seedream-5.0-pro-x` | 1K / 2K，默认 2K | 32000 个 UTF-16 码元 | 10 张 | 1:1、2:3、3:2、3:4、4:3、9:16、16:9、21:9 |
| `seedream-5.0-lite-x` | 2K / 3K / 4K，默认 2K | 32000 个 UTF-16 码元 | 14 张 | 1:1、2:3、3:2、3:4、4:3、9:16、16:9、21:9 |

Pro X / Lite X 的 `prompt` 不能为空；`n` 为整数 1–10，默认 1。`resolution` 选择各自支持的档位，`aspect_ratio` 默认 1:1。`size` 默认 auto，也支持对应分辨率档位、比例或 WIDTHxHEIGHT；像素尺寸映射到最接近的可用比例与档位，不保证精确像素。`response_format` 可选 url / b64_json，默认 url。

Pro X 文生图示例：

```bash
API_URL="YOUR_API_URL"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/images/generations" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "seedream-5.0-pro-x",
    "prompt": "一座被云雾环绕的未来城市，电影感光影",
    "resolution": "2K",
    "aspect_ratio": "16:9",
    "n": 1,
    "response_format": "url"
  }'
```

Lite X 文生图使用相同接口，将 model 改为 `seedream-5.0-lite-x`，可选择 `resolution: "4K"`。页面提供独立可复制示例。

Lite X 参考图编辑示例：

```bash
API_URL="YOUR_API_URL"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/images/edits" \
  -H "Authorization: Bearer $API_KEY" \
  -F "model=seedream-5.0-lite-x" \
  -F "prompt=保持人物外观，生成电影感街拍场景" \
  -F "image=@reference.png" \
  -F "resolution=3K" \
  -F "aspect_ratio=16:9" \
  -F "n=1" \
  -F "response_format=b64_json"
```

Pro X / Lite X 无独立质量档位，使用模型默认质量。多图参考使用重复的 multipart `image` 字段，不要手写 Content-Type boundary。客户端应解析完整 `data[]`，兼容 `data[].b64_json` 和 `data[].url`；账号开启 URL 转 Base64 时可能统一返回 b64_json。

本次更新公开文档和 Lite X 别名准入，不调整本站已有计费规则；不会将供应商的结算说明自动作为本站价格或无图退款承诺。

### OpenAI 兼容 Grok Imagine 图片模型

- OpenAI 渠道通过 API URL + Key 接入上游时，可使用 `grok-imagine-image`、`grok-imagine-image-2.0`、`grok-imagine-image-quality`。
- 这些模型使用本站 `POST /v1/images/generations` 文生图和 `POST /v1/images/edits` 参考图接口；请求体中的模型名会按直接请求或账号模型映射结果原样转发给上游。
- 这组模型只选择 OpenAI API Key 账号，不进入 Codex OAuth/Setup Token 生图路径。上游必须实际支持对应的 OpenAI 兼容图片接口，管理员还需配置模型白名单/映射、图片生成权限和渠道图片定价。

### Gemini Nano Banana

- `gemini-3-pro-image-preview` 是 Nano Banana Pro，`gemini-3.1-flash-image-preview` 是 Nano Banana 2，`gemini-nano-banana-2.1` 是 Nano Banana 2.1；三个模型都使用 Gemini 原生 `POST /v1beta/models/{model}:generateContent`。
- 三个原生模型使用相同的 `contents`、`generationConfig.responseModalities` 和 `generationConfig.imageConfig` 请求体，支持 2K、4K 和常用画幅；文生图和参考图生成只需替换请求 URL 中的模型名称。
- 原生参考图放入 `contents[].parts[].inlineData`，最多 9 张；`inlineData.data` 只传纯 base64，不带 data URL 前缀。
- `gemini-3.1-flash-image-preview` 和 `gemini-nano-banana-2.1` 必须使用原生接口；三个模型都不支持 mask 遮罩编辑。
- 用户页面为 Nano Banana 2.1 新增原生文生图 curl 示例，图片继续从 `candidates[].content.parts[].inlineData` 读取，并将 `data` 按 base64 解码；所有示例仅使用本站 API 地址与用户密钥占位符。

### 音乐模型

- 模型包括 `suno-v6-mini`、`suno-v6` 和 `suno-v6-wild`，均使用本站 OpenAI 兼容 `POST /v1/chat/completions` 接口，令牌分组请选择“音乐”。
- `messages` 支持字符串或纯文本数组；v6 系列合并提示词最多 3000 个 UTF-16 单元，可选 `instrumental` 布尔字段，默认 false。
- 每次请求创建一个独立音乐任务，`n` 只能为 `1`；请求正文会按原样转发到上游。
- `stream=false` 返回完整 JSON；`stream=true` 会在音乐生成完成后以 SSE 返回最终结果，不提供中间进度事件。一个响应可能包含多段音频，下载链接都位于 `choices[0].message.content` 的文本中。
- 客户端应从 `choices[0].message.content` 提取全部 HTTP(S) 音频链接、去重后逐个下载；不要只读取第一条链接。

```bash
API_URL="YOUR_API_URL"
API_KEY="YOUR_API_KEY"

curl --max-time 600 "$API_URL/v1/chat/completions" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: suno-demo-$(date +%s)" \
  -d '{
    "model": "suno-v6",
    "messages": [
      {"role": "user", "content": "Create upbeat piano background music with a warm mood and gentle percussion."}
    ],
    "instrumental": true,
    "n": 1,
    "stream": false
  }'
```

完整响应示例：

```json
{
  "id": "chatcmpl-suno-example",
  "object": "chat.completion",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "第一段音频：https://example.com/suno-output-random-a.mp3\n第二段音频：https://example.com/suno-output-random-b.mp3"
      },
      "finish_reason": "stop"
    }
  ]
}
```

支持参数：

| 参数 | 类型 | 默认值 / 范围 | 说明 |
| --- | --- | --- | --- |
| `model` | string | `suno-v6-mini` / `suno-v6` / `suno-v6-wild` | 必填，使用完整模型名。 |
| `messages` | array | 1–32 条 | 支持字符串或纯文本数组；合并提示词最多 3000 个 UTF-16 单元。 |
| `instrumental` | boolean | `false` | 可选，是否生成纯音乐。 |
| `n` | integer | `1` | 每次只创建一个独立任务。 |
| `stream` | boolean | `false` | 流式响应只返回最终结果，不返回进度事件。 |
| `stream_options.include_usage` | boolean | `false` | 可选，流式末尾附加协议兼容的 usage。 |
| `Idempotency-Key` | header | — | 可选恢复标识；结果不明确时不要自动重复提交。 |

音乐生成可能需要数分钟，实际生成可能超过 200 秒；客户端和反向代理的读取/响应超时建议至少设置为 600 秒，并关闭自动重试。结果不明确时保留同一个 `Idempotency-Key`，先核查原请求，不要换新标识重复提交，以免创建重复音乐任务。

### 视频模型SD2.0

- 保持现有模型表格、参数表格和示例卡片，展示 Firefly、`viraldance933`、`viraldance933-fast`、`viraldance2.5-30` 与 Dola 2.0/2.5；七个模型共用 `POST /v1/videos` 创建和 `GET /v1/videos/{task_id}` 查询。

| 模型名称 | 说明 | 规格 | 画幅 |
| --- | --- | --- | --- |
| `firefly-video-v2-fast` | sd2.0-fast | 5 / 10 / 15 秒；480p / 720p；支持 9 图 / 3 视频 / 3 音频 | 21:9 / 16:9 / 4:3 / 1:1 / 3:4 / 9:16 |
| `firefly-video-v2` | sd2.0 | 5 / 10 / 15 秒；480p / 720p / 1080p；支持 9 图 / 3 视频 / 3 音频 | 21:9 / 16:9 / 4:3 / 1:1 / 3:4 / 9:16 |
| `viraldance933` | sd2.0 | 4 至 15 秒；720p；支持 9 图 / 3 视频 / 3 音频 | 16:9 / 9:16 / 1:1 / 4:3 |
| `viraldance933-fast` | sd2.0-fast | 4 至 15 秒；720p；支持 9 图 / 3 视频 / 3 音频 | 16:9 / 9:16 / 1:1 / 4:3 |
| `viraldance2.5-30` | sd2.5 | 4 至 30 秒；720p；支持 30 图 / 10 视频 / 10 音频 | 16:9 / 9:16 / 1:1 |
| `dola-viraldance2.0` | sd2.0 | 4 至 15 秒；720p；支持 9 图 | 16:9 / 9:16 / 1:1 / 4:3 |
| `dola-viraldance2.5` | sd2.5 | 4 至 30 秒；720p；支持 9 图 | 16:9 / 9:16 / 1:1 / 4:3 |

- Firefly：`firefly-video-v2-fast` 支持 480p、720p；`firefly-video-v2` 额外支持 1080p。两个模型都支持 5、10、15 秒，画幅使用 `aspect_ratio`。
- Firefly 纯文本和分镜请求使用 JSON；包含参考素材时必须使用真实 `multipart/form-data` 文件字段，包括单个 `first_frame`、单个 `last_frame`，以及可重复的 `images`、`videos`、`audios`。图片合计最多 9 张，视频最多 3 个，音频最多 3 个。
- Firefly 参考素材不接受 URL、Base64、`asset_id` 或本地路径字符串。multipart 普通参数均使用字符串，客户端必须自动生成 Content-Type boundary。`prompt` 与 `shots` 二选一；`shots` 支持 2 至 15 个分镜，每段包含 `prompt` 和 `duration`，分镜总时长必须等于任务 `duration`。
- Firefly 建议每 3 至 5 秒查询，`completed` 后通过 `HEAD /v1/videos/{task_id}/content` 获取文件信息，通过 `GET /v1/videos/{task_id}/content` 播放或下载，并支持 Range。上游允许整个请求最大 384 MiB；本站默认 `gateway.max_body_size` 为 256 MiB，确需更大请求时必须同时调整该配置和反向代理上传限制。
- ViralDance 933：`viraldance933` 为标准版，`viraldance933-fast` 为快速版，两者参数相同。必须使用 JSON，填写 `prompt`、整数 `duration`（4–15 秒）和顶层 `ratio`（16:9、9:16、1:1、4:3）。分辨率固定为 720p；`generate_audio` 当前固定为 true，可省略。Dola `dola-viraldance2.0` 使用 4–15 秒，`dola-viraldance2.5` 使用 4–30 秒，其余参数沿用同一视频接口。
- ViralDance 933 参考素材使用公网可直接下载的 URL。单素材可用 `image_url`、`video_url`、`audio_url`，多素材推荐 `image_urls`、`video_urls`、`audio_urls` 数组（兼容 `images`、`videos`、`audios`），分别最多 9 张、3 条、3 条。音频必须与图片或视频一起提交。`@Image1`、`@Video1`、`@Audio1` 按数组顺序从 1 开始引用；`elements` 支持最多 20 个主体，每项至少包含 `frontal_image_url` 或非空的 `reference_image_urls`。
- ViralDance 933 创建后保存 `task_id` 或 `id`，建议每 10 秒查询；`queued`、`processing`、`in_progress` 时继续等待，不重复提交。`succeeded` 或 `completed` 后依次读取 `video.url`、`result_url`、`url` 播放或下载；不使用 Firefly 的 `/content` 文件接口。`failed` 时读取 `error`。
- `viraldance2.5-30`：模型列表、参数说明和示例标题统一使用完整模型名，不展示画布简称；不要与 `dola-viraldance2.5` 混用。使用 JSON，`prompt` 必填、最多 5000 字符；`duration` 支持 4–30 秒整数，默认 5 秒；`aspect_ratio` 支持 16:9、9:16、1:1，默认 16:9；`resolution` 固定 `720p`，传 `async: true`。
- `viraldance2.5-30` 普通参考图使用 `image_urls`，首尾帧分别使用 `start_image_url` / `end_image_url`，各最多 1 张，图片总数（含首尾帧）最多 30 张。视频与音频分别使用 `video_reference: [{"url":"..."}]` / `audio_reference: [{"url":"..."}]`，各最多 10 个；参考视频单个 3–10 秒、合计不超过 30 秒，参考音频单个 3–30 秒、合计不超过 30 秒。音频可以单独作为参考，不沿用 933 的搭配限制。
- `viraldance2.5-30` 素材必须为公网可直接下载的 URL，不直接提交本地路径、Base64 或 multipart 文件。图片支持 JPG/PNG/WebP，视频支持 MP4/MOV，音频支持 MP3/WAV/M4A/AAC/OGG/WebM。提示词用 `@图片1`、`@视频1`、`@音频1` 按各自数组顺序引用；首尾帧不占普通图片编号。不要套用 933 的 `ratio`、`video_urls`、`audio_urls`、`generate_audio`，也不要额外传 `size`。
- `viraldance2.5-30` 创建后保存 `task_id` 或 `id`，使用本站和同一分组密钥每 5 秒查询。`submitted`、`queued`、`processing`、`in_progress` 时继续等待，不重复创建；`completed` 后读取顶层 `url` 播放或下载，不使用 `/content`；`failed` 时查看 `error`。
- 用户页面保留原 Firefly 与 ViralDance 933 示例，新增 `viraldance2.5-30` 文生视频、首尾帧与多模态参考、查询和下载三个示例；地址与密钥继续使用本站占位符，不包含供应商地址或密钥。该模型的参数与调用方式对照 `../gpt_image_playground` 无限画布的 `src/lib/jmVideo.ts`、`src/lib/openaiCompatibleImageApi.ts` 及对应测试；本次只补文档，不改后端路由与计费。

### 视频模型Wan3.0

- 独立分类位于 SD2.0 之后，保持现有模型表格、参数表格和示例卡片风格。仅展示本站调用方式，API 地址与密钥使用 `YOUR_API_URL`、`YOUR_API_KEY` 占位符，参考素材使用用户自己的公网 URL 占位符。
- 模型：`wan3.0x` 为 720p，`wan3.0x-480p` 为 480p，`wan3.0x-1080p` 为 1080p。分辨率由模型名称决定，无需传 `resolution` 或 `size`。
- 统一使用 JSON `POST /v1/videos` 创建任务。必填 `prompt`、整数 `duration`（2–30 秒）和顶层 `ratio`（16:9、4:3、1:1、3:4、9:16）；图生视频和参考生成也必须填写提示词。
- 参考素材使用公网可直接下载的 URL。单素材可用 `image_url`、`video_url`、`audio_url`，多素材推荐 `image_urls`、`video_urls`、`audio_urls` 数组（兼容 `images`、`videos`、`audios`），分别最多 10 张、5 条、5 条；`@Image1`、`@Video1`、`@Audio1` 按同类数组顺序从 1 开始引用。
- 创建后保存 `task_id` 或 `id`，使用本站 `GET /v1/videos/{task_id}` 和同一分组密钥每 10 秒查询一次。`queued`、`processing`、`in_progress` 时继续等待，不重复提交；`completed` 或 `succeeded` 后优先读取 `metadata.url`，其次 `url`，使用结果地址播放或下载，不使用 `/content` 文件接口。`failed` 时读取 `error`。
- 用户页面提供 480p 文生视频、720p 单图生视频、1080p 多模态参考、查询和下载四个示例。

### 视频模型MiniMax-H3

- 模型名固定为 `MiniMax-H3`，通过 `POST /v1/videos` 创建异步任务。请求使用 JSON `content[]`，提示词使用 `{"type":"text","text":"..."}`。
- `resolution` 必须为 `768P` 或 `2K`；`duration` 为 4 至 15 秒的整数；`ratio` 支持常用横竖屏比例，首尾帧和多参考模式可使用 `adaptive`。
- 素材分别使用 `image_url`、`video_url`、`audio_url` 类型，并明确填写 `first_frame`、`last_frame`、`reference_image`、`reference_video` 或 `reference_audio` 角色。
- 本地参考素材先通过 `POST /pg/assets` 上传，再把响应中的 `data.url` 放入对应的 URL 字段。文/图生视频最多使用一张首帧和一张尾帧；多参考模式最多使用 9 张图片、3 个视频和 3 个音频，两种模式不能混用。
- 参考图片支持 JPEG、PNG、WebP；参考视频支持 MP4、MOV，单个不超过 50MB、2 至 15 秒且合计不超过 15 秒；参考音频支持 MP3、WAV，单个不超过 15MB、2 至 15 秒且合计不超过 15 秒。参考视频必须使用本站素材 URL 以核验时长并按实际秒数计费，参考图片和音频不额外计费。
- 创建响应中的 `task_id` 用于 `GET /v1/videos/{task_id}`。当 `task.status` 为 `succeeded` 后，通过 `GET /v1/videos/{task_id}/content` 下载 MP4；`failed` 或 `cancelled` 表示任务失败。
- 如果上游已经创建任务、但创建响应因网关超时丢失，本站仍返回 HTTP 200 和一个 `task_recovery_...` 任务 ID。客户端必须继续使用该 ID 查询状态，不要重新提交创建请求；恢复期间状态保持 `processing`，恢复成功后按普通任务查询和下载。
- 用户页面提供文生视频、素材上传、首尾帧、多模态参考、查询和下载的完整 `curl` 示例。

### Grok 图片视频模型

- `grok-imagine-image-quality` 文生图使用 `/v1/images/generations`，图生图使用 `/v1/images/edits`；图生图最多 3 张参考图，不支持 mask。
- `grok-imagine-image-2.0` 使用相同接口、请求参数和响应解析，只需替换请求体中的 `model`；规格展示为 `1K / 2K；最多 3 张参考图`，并提供独立的文生图与参考图 curl 示例。
- 2.0 的参考图上限按 `../gpt_image_playground/src/lib/grokImagine.ts` 的 `GROK_IMAGINE_MAX_INPUT_IMAGES = 3` 和 `src/lib/openaiCompatibleImageApi.ts` 的提交前校验核对；该项目当前也会拒绝超过 3 张的参考图，尚未实现更高上限。参考图通过 JSON `image` / `images` 传 `{ "type": "image_url", "url": "data:image/...;base64,..." }`，不支持 mask。
- Grok 平台下的 `grok-imagine-video-1.5` 和 `grok-imagine-video` 使用 `/v1/videos/generations`。
- 上述两个视频模型也支持 OpenAI 平台的 API URL + Key 账号：本站 `/v1/videos` 转发上游 `/v1/videos`，本站 `/v1/videos/generations` 转发上游同路径。`api.kunrouai.com` 上游按所提供文档应使用 `/v1/videos`；不支持 Codex OAuth/Setup Token 账号。JSON `image.url` 或 multipart `input_reference` 文件图生视频可省略 prompt，文生视频仍需 prompt；支持 `seconds`、`size` 计费别名。
- Grok 平台的 `grok-imagine-video-1.5` 最多使用 1 张参考图；2 至 7 张参考图使用 `grok-imagine-video`。
- Grok 平台多图视频通过 `reference_images[].url` 传完整 data URL，只能使用 6 秒或 10 秒；不适用于上述单张首帧的 OpenAI 兼容上游。
- 视频任务通过 `/v1/videos/{task_id}` 查询，创建响应的 `request_id`、`task_id` 或 `id` 均可作为任务 ID，`completed` / `done` 表示完成。创建先返回任务，不同步返回视频。`api.kunrouai.com` 上游查询结果不含 URL，完成后使用本站 `/v1/videos/{id}/content` 下载 MP4；Grok 平台沿用原内容下载逻辑。其他 OpenAI API Key 上游仅提供 `video.url` 时使用该 URL。本站未增加 `/v1/tasks/{id}/artifacts` 文件列表接口。接入配置及计费规则见 `docs/OPENAI_MEDIA_COMPAT.md` 的 Grok Imagine 视频章节。

### Grok Build 密钥配置

- Grok 分组下创建的密钥，在“使用密钥”的 Grok CLI 标签中会生成 `~/.grok/config.toml`；Windows 路径为 `%USERPROFILE%\.grok\config.toml`。
- 模板把当前站点 `/v1` 地址写入 `[endpoints].models_base_url`，把当前密钥写入 `[model."grok-4.6"].api_key`，默认模型为 `grok-4.6`，默认推理强度为 `xhigh`。
- 模板关闭自动更新，启用内部安装器，声明 500000 上下文窗口与 `low`、`medium`、`high`、`xhigh` 推理档位，并包含 marketplace、UI 和 CLI 设置。
- 配置文件包含可用 API 密钥，覆盖前应备份已有文件，且不能提交到代码仓库。

## 计费提示

图片和视频生成按模型规则计费。异步视频创建成功后，状态查询和内容下载不重复扣费；任务失败或取消时按系统规则退款。
