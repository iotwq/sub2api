# OpenAI 媒体兼容接口

本项目的 OpenAI APIKey 分组支持以下媒体转发能力：

## 图片

当前 OpenAI 兼容图片接口允许以下 NovelAI Diffusion 模型直接使用 `POST /v1/images/generations`：`nai-diffusion-4-5-full`、`nai-diffusion-5-full`。它们与其他图片模型共用模型映射、上游转发、价格和响应格式处理；在渠道定价中配置价格后即可参与计费。

### Midjourney V7 / Seedream 5.0 Pro / Nano Banana 2

OpenAI 渠道通过 API URL + Key 接入、且提供 OpenAI 兼容图片接口的服务，可使用 `Midjourney-V7`、`seedream-5.0-pro`、`seedream-5.0-pro-x`、`nano-banana2` 和 `nano-banana2-pro`。以下八个完整模型名均可直接作为请求的 `model`，无需先改成无后缀名称：

- `Midjourney V7（满血）`
- `Midjourney-V7`
- `seedream-5.0-pro（X）`
- `seedream-5.0-pro`
- `seedream-5.0-pro-x`
- `seedream-5.0-pro（满血）`
- `nano-banana2-pro（满血）`
- `nano-banana2（满血）`

也可配置账号模型映射，例如：

| 用户请求模型 | 映射后的模型名示例 |
| --- | --- |
| `Midjourney-V7` | `Midjourney V7（满血）` |
| `seedream-5.0-pro` | `seedream-5.0-pro（X）` |
| `seedream-5.0-pro` | `seedream-5.0-pro（满血）` |
| `seedream-5.0-pro` | `seedream-5.0-pro-x` |
| `nano-banana2-pro` | `nano-banana2-pro（满血）` |
| `nano-banana2` | `nano-banana2（满血）` |

模型分类忽略大小写，兼容 `Midjourney V7` 的空格写法，以及中文/英文括号标签；真正转发时保留映射右侧名称的原始大小写、空格和标签，不改写为 GPT 模型。

调用本站 `POST /v1/images/generations`，编辑请求沿用 `POST /v1/images/edits` 的 JSON 或 multipart 格式；转发仍使用账号配置的地址、密钥和代理。上游必须实际支持对应 OpenAI 接口、参数和图片响应，本站不会将请求转换为 Midjourney 专用任务协议，也不保证供应商提供编辑能力。请求参数和上传文件沿用现有透传，仅替换映射后的 `model`。

以上兼容模型仅选择 OpenAI API Key 账号；渠道映射后也采用相同限制，不进入 Codex OAuth/Setup Token 生图路径。管理员需为目标分组开启图片生成，配置账号模型白名单/映射及渠道图片定价；本次路由支持不新增或修改模型价格。同步图片结果继续遵循账号的“图片 URL 转 Base64”开关。

这里的 `nano-banana2` 名称使用标准 `/v1/images/*` 转发，不会自动改名为 Gemini 模型，也不会切换到 `/v1/api/nano-banana` 专用接口。

`gpt-image-2`、`gpt-image-2.5-flare`、`gpt-image-2.5-sunburst` 及其括号标签名称已有路由支持。

### 图片响应 URL 转 Base64 开关

OpenAI APIKey 账号新增和编辑页面提供“图片 URL 转 Base64”开关，保存到 `accounts.extra.openai_image_url_to_b64_json`。开关默认关闭，已有账号与新账号都不会自动改变上游响应：`data[].url`、`data[].b64_json` 或两者同时存在时均原样返回。

管理员开启开关后，该账号的所有同步图片 JSON 响应统一处理：

- 已有 `b64_json` 时原样保留，并删除同一项中的 `url`；
- data URL 或 HTTP(S) URL 会由服务端读取图片并转换为标准 Base64；
- 客户端最终只收到 `data[].b64_json`，不会收到原始上游 URL；
- 下载失败、非图片响应、单图超过 20 MiB、超过 5 次重定向或总下载时间超过 60 秒时返回上游失败，不回退透传 URL；
- 下载请求不携带上游 API Key，并沿用生成账号的代理配置。

启用 `security.url_allowlist` 时，图片初始 URL 和每一跳重定向的域名都必须包含在 `security.url_allowlist.upstream_hosts` 中；同时沿用 `allow_private_hosts` 策略进行 DNS 解析后地址校验。关闭白名单时仍只接受配置允许的 HTTP(S) URL。

## 视频

- `POST /v1/videos`
- `GET /v1/videos/{task_id}`
- `HEAD /v1/videos/{task_id}/content`
- `GET /v1/videos/{task_id}/content`

即梦 SD2.0 请求体会保留 `seconds`、`aspect_ratio`、`images`、`videos`、`audios` 等字段并原样转发。该系列的参考素材使用公网 URL；本地文件先调用素材上传接口。

### Wan 3.0x / ViralDance 视频

支持下表中的 10 个精确模型名，使用普通 OpenAI APIKey 账号。`dola-viraldance2.0` 和 `dola-viraldance2.5` 沿用 `viraldance933` 的接口和素材参数，时长按各自模型校验：2.0 为 4–15 秒，2.5 为 4–30 秒。

1. 在 OpenAI 渠道添加 API URL + Key 账号，Base URL 填 `https://api.viralee.top`（末尾带 `/v1` 也可以），Key 填上游密钥，并绑定目标 OpenAI 分组。
2. 将所需模型加入账号模型白名单/映射，分组开启“允许图片生成”（现有视频入口共用此开关）。不要为此账号开启“MiniMax H3 视频”端点能力，否则任务查询会使用不兼容的 MiniMax `/v2` 协议。
3. 在渠道定价中为模型配置实际对客单价，再用该分组下的**本站 API Key**调用本站 `POST /v1/videos`；不要把上游 Key 发给用户。模型可路由并不代表已经配置价格，缺少有效价格时会在创建上游任务前拒绝请求。

| 模型 | 分辨率 | `duration`（整数秒） | 定价说明 |
| --- | --- | --- | --- |
| `wan3.0x` | 720p | 必填，2–30 | 可选“视频（按秒）”，填写美元/秒 |
| `wan3.0x-480p` | 480p | 必填，2–30 | 同上，按独立模型配置单价 |
| `wan3.0x-1080p` | 1080p | 必填，2–30 | 同上，按独立模型配置单价 |
| `viraldance933`、`viraldance933-fast` | 720p | 必填，4–15 | 上游文档未确定计费口径，按实际合同配置按次或按秒 |
| `dola-viraldance2.0` | 720p | 必填，4–15 | 独立配置对客价格，按次或按秒；不自动继承 933 的价格 |
| `dola-viraldance2.5` | 720p | 必填，4–30 | 独立配置对客价格，按次或按秒；不自动继承 933 的价格 |
| `viraldance2.5-30` | 720p | 4–30，省略默认 5 | 可选“视频（按秒）”，填写美元/秒 |
| `viraldance2.5-15` | 720p | 4–15，省略默认 15 | 上游按条计费，可选“按次”，填写每条美元价格 |
| `viraldance2.5-480p-15` | 480p | 4–15，省略默认 15 | 同上，时长不改变按次费用 |

不会自动采用供应商文档里的价格或币种。通用“视频（按秒）”使用该模型的单一每秒价格，30 秒请求完整按 30 秒计费；参考图片/视频/音频不额外计费。按次规则不乘视频时长。已有分组级视频秒价仍按原优先级生效，因此若希望这两款 15 秒模型按渠道的按次价格计费，应避免给该分组配置会覆盖它们的统一视频秒价。

`viraldance2.5-30` 上游文档把 `duration` 标为 number，但当前本站视频账务采用整数秒；本站仅接受 4–30 的整秒数值，小数、字符串、null 和超范围时长在调用上游前返回 400，不会静默截断计费。其他模型的时长同样使用 JSON 数值形式的整数。省略时长仅适用于表中有默认值的模型；不要用 `seconds` 代替这四类接口的 `duration`。

**Dola 模型使用**

`dola-viraldance2.0` 和 `dola-viraldance2.5` 共用本站 `POST /v1/videos` 创建、`GET /v1/videos/{task_id}` 查询，使用 JSON；不需要新增端点或开启 Seedance/MiniMax 专用能力。账号服务地址应填写实际接入地址，不因模型名而固定到某个供应商。账号模型白名单/映射和渠道定价都需配置对应模型。

请求需显式填写 `duration` 和顶层 `ratio`（16:9、9:16、1:1、4:3）：`dola-viraldance2.0` 为 4–15 秒，`dola-viraldance2.5` 为 4–30 秒；均为整数，分辨率固定 720p，`resolution` 可省略；`generate_audio` 可省略或传 true。缺失时长、用 `seconds` 替代、小数或超范围时长会在创建前拒绝。素材和比例等其他参数沿用 933 透传行为，由上游按文档校验。

模型名直接传给上游；配置账号模型映射时仅替换 `model`，其余字段保留。可将已支持的 `viraldance933` 映射为上述任一模型。此功能不自动开放任意自定义模型名，也不把 Dola 模型改名为 933。

```json
{
  "model": "dola-viraldance2.0",
  "prompt": "海边日落，镜头缓慢推进",
  "duration": 8,
  "ratio": "16:9",
  "resolution": "720p"
}
```

使用 2.5 时将 `model` 改为 `dola-viraldance2.5`。创建后保存 `id` 或 `task_id`，每约 10 秒查询；`succeeded`/`completed` 时优先读取 `video.url`，其次 `result_url`、`url`。查询固定到创建账号且不重复计费，明确失败沿用幂等退款补偿。文档未提供 `/content`，请直接使用成功响应的视频 URL。

请求使用 `application/json`，除已有模型映射外保持原字段结构：

- Wan 与 ViralDance 933 / 2.5-15 使用 `ratio`；2.5-30 使用 `aspect_ratio`，或按文档传 `size`。Wan 的分辨率由模型名决定，不需要额外填写 `resolution` / `size`。
- `image_url`、`image_urls`、`images`、`video_url`、`video_urls`、`videos`、`audio_url`、`audio_urls`、`audios` 按对应模型文档透传。2.5-30 的 `video_reference` / `audio_reference` 对象数组、`start_image_url` / `end_image_url`，以及 933 的 `elements` 均保留。
- 素材 URL 必须公开可访问；各模型的素材数量、参考时长和组合限制由上游校验。2.5-15 不支持视频参考；不要把其他系列支持的字段套用到它。
- `async: true`、`generate_audio: false` 等布尔值不会被删除或重写。请求 JSON 不会被转成 Firefly multipart 或 MiniMax `content[]`。

示例（地址和密钥替换为本站地址及用户分组 Key）：

```bash
curl 'https://你的本站地址/v1/videos' \
  -H 'Authorization: Bearer 你的本站APIKey' \
  -H 'Content-Type: application/json' \
  -d '{"model":"wan3.0x","prompt":"海边日落，镜头缓慢推进","duration":30,"ratio":"16:9"}'
```

创建返回异步任务，不代表生成已完成。保存响应中的 `task_id` 或 `id`，继续使用同一分组的 Key 查询本站 `GET /v1/videos/{task_id}`。任务始终使用创建时绑定的账号，不在每次查询时重新提交；查询不重复计费。Wan / 933 / 2.5-15 建议间隔 10 秒，2.5-30 建议 3–5 秒：

```python
task_id = created.get("task_id") or created.get("id")
# result 是 GET /v1/videos/{task_id} 的 JSON 响应
status = result.get("status")
if status in ("completed", "succeeded"):
    video_url = (
        (result.get("metadata") or {}).get("url")  # Wan
        or (result.get("video") or {}).get("url")  # 933 / 2.5-15
        or result.get("result_url")
        or result.get("url")                      # 2.5-30 或兼容回退
    )
elif status == "failed":
    error = result.get("error")
# submitted / queued / processing / in_progress：等待下次查询。
```

响应中的 URL、任务状态、进度和错误原样返回。以上四类上游没有在文档中提供 `/content` 下载接口，客户端应在成功后读取上述视频 URL 播放或下载；不要假设本站 `/v1/videos/{task_id}/content` 能代理这类文件。创建成功后沿用现有用量记账，任务明确失败后沿用原扣费记录的幂等退款及后台补偿流程。

### Firefly Video v2

YCYAPI Firefly Video v2 上游按普通 OpenAI APIKey 账号添加，Base URL 填协议根地址（例如 `https://ycyapi.cn`），模型使用 `firefly-video-v2` 或 `firefly-video-v2-fast`。创建、查询和内容接口统一转发到上游 `/v1/videos` 路径，不使用旧 `/v1/video/generations`。

普通 Firefly 任务的账号绑定状态为 `inactive`，表示未进入 MiniMax-H3 创建超时恢复。Firefly 账号不要开启“MiniMax H3 视频”端点能力；按此配置，后续状态查询和内容下载仍固定到创建任务的原账号，并继续使用标准 `/v1/videos/{task_id}` 路径，不会进入 MiniMax V2 能力筛选或 `/v2/query/video_generation` 路径。

无参考素材时使用 JSON；包含参考图片、视频或音频时，请求必须使用真实 `multipart/form-data` 文件字段。系统会保留重复的 `images`、`videos`、`audios` 字段、文件名、Content-Type 和文件字节；配置模型映射时只重写 multipart 的 `model` 普通字段。

文件字段为 `first_frame`、`last_frame`、`images`、`videos`、`audios`。不支持素材 URL、Base64、`asset_id` 或路径字符串。上游最大请求为 384 MiB，而本站 `gateway.max_body_size` 默认是 256 MiB；确需接收 256 至 384 MiB 的请求时，需要显式提高该配置，并同步提高 Nginx Proxy Manager 等反向代理的上传限制。

#### SD2.0 按分辨率、按秒计费模板

管理员在渠道定价中添加 `firefly-video-v2` 或 `firefly-video-v2-fast` 后，页面会自动切换到 SD2.0 专用的“视频（按秒）”模板。每个输入框的单位都是美元/秒，价格由管理员自行填写且必须大于 0：

- `firefly-video-v2-fast`：必须配置 480p、720p；
- `firefly-video-v2`：必须配置 480p、720p、1080p。

两个模型必须分别建立定价规则，不能放在同一条规则中共享价格，因为同一分辨率在标准版和快速版中的单价也可能不同。

```text
原始费用 = 输出视频时长 × 视频数量 × 请求分辨率对应的每秒价格
用户扣费 = 原始费用 × 当前分组的有效视频倍率
```

模板只计算生成的视频，不对参考图片、参考视频或参考音频单独计费。创建任务时解析出的模型、分辨率、时长、渠道价格和倍率会形成费用快照；余额预占、最终扣款和失败退款使用同一份快照。用户和管理员的使用记录会显示实际模型、分辨率、输出秒数、对应的每秒价格和费用。

旧的统一每秒价不会自动推算为各分辨率价格。升级后需要在渠道定价中把标准版和快速版拆成两条规则，并填写各自支持的全部分辨率价格。MiniMax-H3 不使用 SD2.0 模板，仍按下文的 768P/2K 专用规则计费。

### MiniMax H3 V2

MiniMax H3 中转上游仍按 OpenAI APIKey 账号添加：

- Base URL 填协议根地址，例如 `https://metaso.cn/api/minimax`，不要填具体的 `/v2/video_generation` 接口；
- API Key 填上游的 `mk-...`；
- 在“端点能力”中开启“MiniMax H3 视频”；
- 专用视频账号应取消 Chat Completions 和 Embeddings，只保留“MiniMax H3 视频”，避免该账号参与文本模型调度；
- 同步模型会直接返回 `MiniMax-H3`，账号测试使用只读任务列表，不会创建视频或产生视频费用。

对外接口与 MiniMax V2 上游接口的映射如下：

| 本站接口 | MiniMax V2 上游接口 | 是否计费 |
| --- | --- | --- |
| `POST /v1/videos` | `POST /v2/video_generation` | 按 `MiniMax-H3` 渠道配置计费 |
| `GET /v1/videos/{task_id}` | `GET /v2/query/video_generation/{task_id}` | 否 |
| `GET /v1/videos` | `GET /v2/query/video_generation` | 否 |
| `DELETE /v1/videos/{task_id}` | `DELETE /v2/video_generation/{task_id}` | 否 |
| `POST /v1/videos/context-ir` | `POST /v2/h3_context_ir` | 按 `MiniMax-H3` 渠道单次价格 |
| `POST /v1/videos/regenerations` | `POST /v2/video_regeneration` | 按 `MiniMax-H3` 渠道单次价格 |

#### MiniMax-H3 按秒计费模板

管理员在渠道定价中单独添加完整模型名 `MiniMax-H3` 后，页面会自动切换到“视频（按秒）”模板，并生成两个必填价格：`768P` 和 `2K`，单位均为美元/秒。保存时必须同时填写两个非负价格，且该计费模式不能与其他模型共用同一条定价规则。

标准视频生成的费用为：

```text
原始费用 =（输出视频时长 + 服务端核验的参考视频实际总时长）× 请求分辨率对应的每秒价格
用户扣费 = 原始费用 × 当前分组的有效视频倍率
```

- 参考图片最多 9 张，不按数量额外收费；
- 参考视频最多 3 个，单个 2–15 秒、合计不超过 15 秒，按实际总时长使用输出分辨率的每秒价格计费；
- 参考音频最多 3 个，不额外收费；
- 参考视频必须先上传到本站 `/pg/assets`，服务端会读取本地 MP4/MOV 文件核验时长。外部视频 URL 无法形成可信计费依据，会在请求上游前返回 `400 invalid_request_error`；
- `server.frontend_url` 必须配置为本站对外 API 地址，并与素材 URL 的主机一致，例如 `https://api.iotwq.top`。

已存在的 `per_request` MiniMax-H3 渠道定价不会自动转换，仍按原来的单次价格执行。按秒模板用于包含明确 `resolution` 和 `duration` 的标准视频生成请求；Context-IR 和视频再生成没有足够的分辨率、时长定价依据时会在上游调用前被拒绝，避免产生无法正确结算的费用。

文生视频、首帧/尾帧和多模态参考请求可直接使用 MiniMax 官方 `content[]` 结构，字段和 `role` 原样转发：

```json
{
  "model": "MiniMax-H3",
  "content": [
    { "type": "text", "text": "镜头缓慢推进，蒸汽自然上升" },
    {
      "type": "image_url",
      "image_url": { "url": "https://example.com/first-frame.png" },
      "role": "first_frame"
    }
  ],
  "resolution": "2K",
  "duration": 5,
  "ratio": "adaptive"
}
```

兼容旧式纯文本字段：`prompt` 会转换为文本 `content[]`，`seconds` 和 `aspect_ratio` 会分别转换为 `duration` 和 `ratio`。参考图、参考视频等媒体输入必须使用官方 `content[]` 并明确填写 `role`；系统不会猜测素材是首帧、尾帧还是普通参考。

任务查询中的上游 CDN 地址会改写为本站 `GET /v1/videos/{task_id}/content`。下载时本站先查询任务，再拉取 `task.content.url`；CDN 请求不会携带 MiniMax API Key。

#### MiniMax-H3 创建超时恢复

`POST /v1/videos` 创建 MiniMax-H3 任务时，如果上游明确返回 HTTP `504`，或者返回 HTTP `200` 但响应体严格包含 `error.http_code = 504`，系统不会切换账号或重新提交，以免同一请求重复创建和重复收费。系统改为向客户端返回 HTTP `200`：

```json
{
  "task_id": "task_recovery_...",
  "status": "processing"
}
```

客户端把该任务 ID 当作普通视频任务继续调用 `GET /v1/videos/{task_id}`。尚未识别真实上游任务或仍在完成结算时，查询返回 `processing`，不会把上游创建阶段的 `504` 继续传给客户端；识别真实任务并完成结算后，状态查询和内容下载固定到原账号并使用真实任务 ID 转发。原账号绑定在普通负载调度和高级调度中都是硬约束：原账号暂时不可用时不会切换到其他账号查询，避免真实任务 ID 在错误 API Key 下返回 `invalid task_id (2013)`。旧版本中已经结算、但最终 `matched` 写入失败且失去调度时间的记录也能直接恢复查询，不会永久停留在 `processing`。对外仍使用同一个 `task_recovery_...` ID。

提交前，系统会记录原账号已有任务基线以及本次模型、分辨率、时长、比例和参考素材数量。后台只在原账号中检查基线后新增的任务：模型、分辨率、时长和比例必须完全一致；上游任务列表返回的参考素材数量只有在可信正数时才参与校验，字段缺失或返回 `0` 时按“未知”处理，不会错误排除实际包含图片或音频的任务。只有最终候选唯一时才自动绑定，多个候选仍停止关联以保护任务归属。恢复记录和计费快照持久化到数据库，服务重启后会继续处理；同一上游任务不能被两条恢复记录同时认领。收到 `504` 时只保留创建前的余额预占，不写正式用量或扣费记录；唯一候选绑定成功后才按创建时冻结的价格快照结算。

恢复窗口为 30 分钟，从向上游发起创建请求前开始计算，覆盖实测约三分钟提示词处理启动、三分钟提示词优化、十三分钟排队、三分钟生成，以及任务列表延迟出现后的后台调度时间。暂时没有候选时继续轮询；出现多个候选或到期仍不能唯一确认时停止自动关联，并释放余额预占，已经结算的请求会按失败任务规则退款。客户端在恢复期间通过正常状态查询持续得到 `processing`，不会在十五分钟时提前失败；到期仍无法确认时才得到 `failed`。此保护优先避免串号，无法确认时不会猜测“最新任务”。已经安全识别并保存真实任务 ID 的记录不受三十分钟确认窗口限制；后续结算或状态推进发生瞬时错误时会持续重试，避免已生成视频丢失交付。与原来的三分钟窗口相比，真正未创建成功的请求可能额外保留余额预占最多二十七分钟。

使用旧的 `per_request` 定价时，Context-IR 和视频再生成仍与普通创建共用 `MiniMax-H3` 的按次渠道价格。任务列表只查询本次调度选中的一个上游账号，不聚合分组内多个 MiniMax 账号的任务。

### 视频失败退款

视频创建响应已经明确为终态失败时，系统会释放预占且不产生扣费。异步任务创建成功并结算后，客户端查询或后台补偿探测发现以下终态失败时，会按创建时冻结的费用快照回补余额或订阅用量，并写入一条负数的“视频失败退款”使用记录：

- 状态字段为 `failed`、`failure`、`error`、`cancelled` 或 `canceled`；
- 状态接口返回 HTTP `500`、`502`、`503` 或 `504`，且错误正文明确包含 `generation failed`（支持 `server_error`、`upstream_error`、缺省错误类型和纯文本正文）。该规则同时覆盖 HTTP 成功响应和 HTTP 错误响应。

状态码本身不足以触发退款：普通鉴权失败、任务不存在、限流、临时上游错误、网关 HTML/纯文本提示或不含明确生成失败标记的通用 `server_error` 会按退避策略继续探测。退款按任务 ID 幂等处理，客户端重复查询、多个服务实例并发探测或进程重启都不会重复回补。

正式视频扣费的 `request_id` 按媒体链路区分：普通 OpenAI 兼容视频使用 `openai-video:<billing_task_id>`；只有原本已经携带 `grok-video:` 标识的 Grok 视频才保留 Grok 稳定标识，不再给普通视频追加 Grok 前缀。失败退款按 API Key 精确查询当前 `openai-video:<billing_task_id>`，并兼容历史误写的 `grok-video:openai-video:<billing_task_id>`；两种记录同时存在时不自动选择，避免错退或重复退款。退款仍使用 `openai-video-refund:<billing_task_id>` 作为原子幂等标识。

任务创建时的模型映射、渠道单价、视频倍率和计费结果会作为同一份费用快照用于余额预占、最终结算、使用记录和失败退款。管理员在任务创建后修改定价，不会改变该任务已经确认的扣费或退款金额。

用户和管理员的使用记录会显示输出视频数量、输出秒数和分辨率。MiniMax-H3 还会单独显示服务端核验的参考视频计费秒数、参考视频原始费用和倍率后的实际扣费；图片和音频仍不产生输入素材费用。

## 临时素材

```bash
curl "$API_URL/pg/assets" \
  -H "Authorization: Bearer $API_KEY" \
  -F "kind=image" \
  -F "file=@reference.png"
```

支持的 `kind` 和格式：

| kind | 格式 | 单文件上限 |
| --- | --- | ---: |
| `image` | jpg、jpeg、png、webp | 32 MiB |
| `video` | mp4、mov、webm | 100 MiB |
| `audio` | mp3、m4a、wav、aac、ogg | 15 MiB |

返回的 `data.url` 可用于视频 `images`、`videos`、`audios` 数组。资源保存 24 小时，读取 URL 不需要 API Key。

临时 URL 使用素材上传请求的公网协议和主机生成，不使用 `server.frontend_url`。即使管理前端部署在 `/password` 等子路径，返回 URL 仍为 API 域名下的 `/pg/assets/...`，避免视频上游读取到管理站 HTML。

`/pg/assets` 是根路径接口，不带 `/v1` 前缀。反向代理需要将该路径转发到 Sub2API 后端，不能交给前端 SPA 回退；否则上传请求会收到管理站 HTML 而不是 JSON。

## 音频

### 文字转语音

`POST /v1/audio/speech` 使用 JSON 请求体，字段由上游模型决定，至少包含 `model` 和 `input`。成功时透传上游音频二进制响应。

### 音频转文字与翻译

- `POST /v1/audio/transcriptions`
- `POST /v1/audio/translations`

这两个接口必须使用 `multipart/form-data`，文件字段名为 `file`，模型字段名为 `model`。请求体和上游响应会保留 OpenAI 兼容格式。

所有音频接口仅从 OpenAI APIKey 账号池选择账号；自定义 Base URL 会按账号配置转发，支持账号级请求头覆写和代理。
