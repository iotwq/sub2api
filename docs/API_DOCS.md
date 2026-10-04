# 用户接口文档页

用户侧“接口文档”位于左侧栏，页面路径为 `/api-docs`。页面只展示调用接口所需的模型、参数、示例和必要限制，所有示例统一使用：

```http
Authorization: Bearer YOUR_API_KEY
```

## 模型分类

### OpenAI gpt-image2

- 支持 `gpt-image-2`、`gpt-image-2.5-sunburst`、`gpt-image-2.5-flare`。三个模型的调用方式一致：文生图使用 JSON `POST /v1/images/generations`，参考图和遮罩编辑使用 multipart `POST /v1/images/edits`。
- 三个模型的规格统一展示为「1K / 2K / 4K；最多 16 张参考图」；像素尺寸例如 `1280x1280`、`2048x1152`、`3840x2160`。`quality` 差异只在参数说明中列出：`gpt-image-2` 可选 `low`、`medium`、`high`；两个 2.5 模型额外支持 `xhigh`、`max`。
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

将 `model` 替换为 `gpt-image-2` 或 `gpt-image-2.5-flare` 即可使用相同接口；`gpt-image-2` 不支持 `xhigh`、`max`，请使用 `low`、`medium` 或 `high`。

### OpenAI 兼容 Seedream / Midjourney 图片模型

- OpenAI 渠道通过 API URL + Key 接入上游时，可使用 `seedream-5.0-pro`、`seedream-5.0-pro-x` 和 `Midjourney-V7`。
- 三个模型均使用本站 `POST /v1/images/generations` 文生图和 `POST /v1/images/edits` 参考图编辑接口；账号模型映射后的完整名称会转发给上游。

| 模型 | 提示词上限 | 参考图上限 | 比例 |
| --- | --- | --- | --- |
| `seedream-5.0-pro` | 8000 字符 | 10 张 | 16:9、9:16、1:1、4:3、3:4、3:2、2:3、21:9 |
| `seedream-5.0-pro-x` | 8000 字符 | 10 张 | 16:9、9:16、1:1、4:3、3:4、3:2、2:3、21:9 |
| `Midjourney-V7` | 4000 字符 | 5 张 | 16:9、9:16、1:1、4:3、3:4、4:5、5:4、3:2、2:3、21:9 |

文生图示例：

```bash
API_URL="YOUR_API_URL"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/images/generations" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "seedream-5.0-pro",
    "prompt": "一座被云雾环绕的未来城市，电影感光影",
    "size": "1024x576",
    "n": 1,
    "response_format": "b64_json"
  }'
```

参考图编辑示例：

```bash
curl "$API_URL/v1/images/edits" \
  -H "Authorization: Bearer $API_KEY" \
  -F "model=Midjourney-V7" \
  -F "prompt=保持人物外观，生成电影感街拍场景" \
  -F "image=@reference.png" \
  -F "size=1024x576" \
  -F "n=1" \
  -F "response_format=b64_json"
```

这三个模型按比例选择尺寸，实际输出尺寸以返回图片为准；不要传 GPT 专属的 `quality`、`moderation`、`background`、`output_format` 或 `output_compression` 参数，也不支持 `mask` 遮罩编辑。Midjourney V7 可能一次返回多张图片，客户端应解析完整的 `data[]`。响应可能包含 `data[].b64_json`、`data[].url` 或两者，客户端需要兼容判断。

### Gemini Nano Banana

- `gemini-3-pro-image-preview` 是 Nano Banana Pro，`gemini-3.1-flash-image-preview` 是 Nano Banana 2；两者都使用 Gemini 原生 `POST /v1beta/models/{model}:generateContent`。
- 两个原生模型使用相同的 `contents`、`generationConfig.responseModalities` 和 `generationConfig.imageConfig` 请求体，支持 2K、4K 和常用画幅。
- 原生参考图放入 `contents[].parts[].inlineData`，最多 9 张；`inlineData.data` 只传纯 base64，不带 data URL 前缀。
- `gemini-3.1-flash-image-preview` 必须使用原生接口；两个模型都不支持 mask 遮罩编辑。

### 视频模型SD2.0

- 保持现有模型表格、参数表格和示例卡片，展示 Firefly、`viraldance933`、`viraldance933-fast`、`viraldance2.5-30` 与 Dola 2.0/2.5；七个模型共用 `POST /v1/videos` 创建和 `GET /v1/videos/{task_id}` 查询。
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
- `grok-imagine-video-1.5` 和 `grok-imagine-video` 使用 `/v1/videos/generations`。
- `grok-imagine-video-1.5` 最多使用 1 张参考图；2 至 7 张参考图使用 `grok-imagine-video`。
- Grok 多图视频通过 `reference_images[].url` 传完整 data URL，只能使用 6 秒或 10 秒。
- 视频任务通过 `/v1/videos/{task_id}` 查询，通过 `/v1/videos/{task_id}/content` 下载。

### Grok Build 密钥配置

- Grok 分组下创建的密钥，在“使用密钥”的 Grok CLI 标签中会生成 `~/.grok/config.toml`；Windows 路径为 `%USERPROFILE%\.grok\config.toml`。
- 模板把当前站点 `/v1` 地址写入 `[endpoints].models_base_url`，把当前密钥写入 `[model."grok-4.6"].api_key`，默认模型为 `grok-4.6`，默认推理强度为 `xhigh`。
- 模板关闭自动更新，启用内部安装器，声明 500000 上下文窗口与 `low`、`medium`、`high`、`xhigh` 推理档位，并包含 marketplace、UI 和 CLI 设置。
- 配置文件包含可用 API 密钥，覆盖前应备份已有文件，且不能提交到代码仓库。

## 计费提示

图片和视频生成按模型规则计费。异步视频创建成功后，状态查询和内容下载不重复扣费；任务失败或取消时按系统规则退款。
