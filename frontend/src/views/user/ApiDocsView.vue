<template>
  <AppLayout>
    <div class="api-docs-page">
      <header class="api-docs-header">
        <div class="api-docs-heading">
          <div class="api-docs-icon">
            <Icon name="document" size="xl" />
          </div>
          <div class="min-w-0">
            <p class="api-docs-kicker">API Reference</p>
            <h1 class="api-docs-title">{{ t('apiDocs.title') }}</h1>
            <p class="api-docs-description">{{ t('apiDocs.description') }}</p>
          </div>
        </div>
      </header>

      <section class="api-docs-note">
        <Icon name="infoCircle" size="md" class="shrink-0 text-[#a33a2b]" />
        <p>
          所有示例使用 <code>Authorization: Bearer YOUR_API_KEY</code>。请使用已开通对应模型的 API Key。
        </p>
      </section>

      <section
        v-for="model in modelDocs"
        :key="model.id"
        class="api-docs-section"
      >
        <div class="api-docs-model-header">
          <div>
            <p class="api-docs-model-eyebrow">{{ model.provider }}</p>
            <h2 class="api-docs-model-title">{{ model.name }}</h2>
            <p class="api-docs-model-description">{{ model.description }}</p>
          </div>
          <div class="api-docs-endpoints">
            <span
              v-for="endpoint in model.endpoints"
              :key="endpoint"
              class="api-docs-endpoint"
            >
              {{ endpoint }}
            </span>
          </div>
        </div>

        <div v-if="model.availableModels && model.availableModels.length" class="api-docs-subsection">
          <h3 class="api-docs-subtitle">可用模型</h3>
          <div class="api-docs-table-wrap">
            <table class="api-docs-table">
              <thead>
                <tr>
                  <th>模型名称</th>
                  <th>说明</th>
                  <th>规格</th>
                  <th>画幅</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="availableModel in model.availableModels || []" :key="availableModel.name">
                  <td><code>{{ availableModel.name }}</code></td>
                  <td>{{ availableModel.description }}</td>
                  <td>{{ availableModel.specification }}</td>
                  <td>{{ availableModel.aspectRatio }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="api-docs-subsection">
          <h3 class="api-docs-subtitle">支持参数</h3>
          <div class="api-docs-table-wrap">
            <table class="api-docs-table">
              <thead>
                <tr>
                  <th>参数</th>
                  <th>类型</th>
                  <th>说明</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="param in model.params" :key="param.name">
                  <td><code>{{ param.name }}</code></td>
                  <td>{{ param.type }}</td>
                  <td>{{ param.description }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="api-docs-subsection">
          <h3 class="api-docs-subtitle">调用示例</h3>
          <div class="api-docs-code-grid">
            <article
              v-for="example in model.examples"
              :key="example.title"
              class="api-docs-code-card"
            >
              <div class="api-docs-code-header">
                <div>
                  <p class="api-docs-code-language">{{ example.language }}</p>
                  <h4 class="api-docs-code-title">{{ example.title }}</h4>
                </div>
                <button
                  type="button"
                  class="api-docs-copy-button"
                  :title="t('common.copy')"
                  @click="copyCode(example.code)"
                >
                  <Icon name="copy" size="sm" />
                </button>
              </div>
              <pre><code>{{ example.code }}</code></pre>
            </article>
          </div>
        </div>

        <div v-if="model.notes.length" class="api-docs-model-notes">
          <p v-for="note in model.notes" :key="note">{{ note }}</p>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'

interface ApiParam {
  name: string
  type: string
  description: string
}

interface CodeExample {
  title: string
  language: string
  code: string
}

interface AvailableModel {
  name: string
  description: string
  specification: string
  aspectRatio: string
}

interface ModelDoc {
  id: string
  name: string
  provider: string
  description: string
  endpoints: string[]
  availableModels?: AvailableModel[]
  params: ApiParam[]
  examples: CodeExample[]
  notes: string[]
}

const apiBaseUrl = 'YOUR_API_URL'

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

function copyCode(code: string): void {
  void copyToClipboard(code)
}

const gptImage2GenerateCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/images/generations" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-image-2",
    "prompt": "一只白色暹罗猫，电影级光影，精细毛发",
    "size": "1280x1280",
    "quality": "medium",
    "background": "transparent",
    "output_format": "png",
    "response_format": "b64_json",
    "moderation": "auto",
    "n": 1
  }'`

const gptImage2EditCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/images/edits" \\
  -H "Authorization: Bearer $API_KEY" \\
  -F "model=gpt-image-2" \\
  -F "prompt=把参考图里的猫改成戴贝雷帽的可爱海獭" \\
  -F "image=@reference.png" \\
  -F "mask=@mask.png" \\
  -F "size=1280x1280" \\
  -F "quality=medium" \\
  -F "background=transparent" \\
  -F "output_format=png" \\
  -F "response_format=b64_json" \\
  -F "moderation=auto"`

const gptImage2ResponseHandlingJavascript = `const MIME_BY_FORMAT = {
  png: 'image/png',
  jpeg: 'image/jpeg',
  webp: 'image/webp',
}

async function imageItemToBlob(item, outputFormat = 'png') {
  // 1. data[].b64_json may be raw Base64 or a complete data URL.
  if (typeof item.b64_json === 'string' && item.b64_json) {
    const dataUrl = item.b64_json.startsWith('data:image/')
      ? item.b64_json
      : 'data:' + (MIME_BY_FORMAT[outputFormat] || 'image/png') + ';base64,' + item.b64_json
    return await (await fetch(dataUrl)).blob()
  }

  // 2. data[].url may be a data URL.
  if (typeof item.url === 'string' && item.url.startsWith('data:image/')) {
    return await (await fetch(item.url)).blob()
  }

  // 3. data[].url may be an HTTP(S) address.
  if (typeof item.url === 'string' && /^https?:\\/\\//i.test(item.url)) {
    const response = await fetch(item.url)
    if (!response.ok) throw new Error('Image download failed: HTTP ' + response.status)
    return await response.blob()
  }

  throw new Error('No usable image data returned')
}

async function imageResponseToBlobs(apiResponse, requestedOutputFormat = 'png') {
  const payload = await apiResponse.json()
  const outputFormat = payload.output_format || requestedOutputFormat
  return await Promise.all(
    (payload.data || []).map((item) => imageItemToBlob(item, outputFormat)),
  )
}`

const mappedImageGenerateCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/images/generations" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "seedream-5.0-pro",
    "prompt": "一座被云雾环绕的未来城市，电影感光影",
    "size": "1024x576",
    "n": 1,
    "response_format": "b64_json"
  }'`

const mappedImageEditCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/images/edits" \\
  -H "Authorization: Bearer $API_KEY" \\
  -F "model=Midjourney-V7" \\
  -F "prompt=保持人物外观，生成电影感街拍场景" \\
  -F "image=@reference.png" \\
  -F "size=1024x576" \\
  -F "n=1" \\
  -F "response_format=b64_json"`

const fireflyVideoCreateCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "firefly-video-v2-fast",
    "prompt": "雨后的城市街道，镜头平稳向前移动，霓虹灯倒映在路面",
    "negative_prompt": "文字，字幕，水印，画面抖动",
    "duration": 5,
    "resolution": "720p",
    "aspect_ratio": "16:9",
    "generate_audio": true
  }'`

const fireflyVideoShotsCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "firefly-video-v2",
    "duration": 10,
    "resolution": "1080p",
    "aspect_ratio": "16:9",
    "shots": [
      {"prompt": "广角镜头，两个人走进阳光照射的教室", "duration": 5},
      {"prompt": "近景镜头，两人在窗边相视微笑", "duration": 5}
    ]
  }'`

const fireflyVideoReferenceCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -F "model=firefly-video-v2" \\
  -F "prompt=保持参考人物和场景一致，生成自然的电影感运动" \\
  -F "duration=15" \\
  -F "resolution=720p" \\
  -F "aspect_ratio=16:9" \\
  -F "generate_audio=true" \\
  -F "first_frame=@first-frame.jpg" \\
  -F "last_frame=@last-frame.jpg" \\
  -F "images=@character-reference.png" \\
  -F "images=@scene-reference.webp" \\
  -F "videos=@motion-reference.mp4" \\
  -F "audios=@dialogue-reference.wav"`

const fireflyVideoPollCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"
TASK_ID="YOUR_TASK_ID"

curl "$API_URL/v1/videos/$TASK_ID" \\
  -H "Authorization: Bearer $API_KEY"

curl -I "$API_URL/v1/videos/$TASK_ID/content" \\
  -H "Authorization: Bearer $API_KEY"

curl -L "$API_URL/v1/videos/$TASK_ID/content" \\
  -H "Authorization: Bearer $API_KEY" \\
  -o result.mp4`

const viraldanceVideoCreateCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "viraldance933",
    "prompt": "雨后的城市街道，镜头平稳向前移动，霓虹灯倒映在路面",
    "duration": 8,
    "resolution": "720p",
    "ratio": "16:9"
  }'`

const viraldanceVideoReferenceCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "viraldance933-fast",
    "prompt": "保持 @Image1 中的人物外观，参考 @Video1 的动作和 @Audio1 的节奏，生成自然的电影感视频",
    "duration": 10,
    "resolution": "720p",
    "ratio": "9:16",
    "image_urls": ["https://example.com/character.png"],
    "video_urls": ["https://example.com/motion.mp4"],
    "audio_urls": ["https://example.com/music.mp3"]
  }'`

const viraldanceVideoPollCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"
TASK_ID="YOUR_TASK_ID"

# 使用创建响应中的 task_id 或 id，建议每 10 秒查询一次
curl "$API_URL/v1/videos/$TASK_ID" \\
  -H "Authorization: Bearer $API_KEY"

# status 为 succeeded 或 completed 后，依次读取 video.url、result_url、url
# 将成功响应中的实际视频地址填入下方，再播放或下载
VIDEO_URL="YOUR_VIDEO_URL"
curl -L "$VIDEO_URL" -o result.mp4`

const viraldance25VideoCreateCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "viraldance2.5-30",
    "prompt": "雨后的城市街道，镜头平稳向前移动，霓虹灯倒映在路面",
    "duration": 30,
    "aspect_ratio": "16:9",
    "resolution": "720p",
    "async": true
  }'`

const viraldance25VideoReferenceCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "viraldance2.5-30",
    "prompt": "从首帧自然过渡到尾帧，保持 @图片1 中的人物外观，参考 @视频1 的动作和 @音频1 的节奏",
    "duration": 10,
    "aspect_ratio": "9:16",
    "resolution": "720p",
    "async": true,
    "image_urls": ["https://example.com/character.png"],
    "start_image_url": "https://example.com/start.png",
    "end_image_url": "https://example.com/end.png",
    "video_reference": [{"url": "https://example.com/motion.mp4"}],
    "audio_reference": [{"url": "https://example.com/music.mp3"}]
  }'`

const viraldance25VideoPollCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"
TASK_ID="YOUR_TASK_ID"

# 使用创建响应中的 task_id 或 id，建议每 5 秒查询一次
curl "$API_URL/v1/videos/$TASK_ID" \\
  -H "Authorization: Bearer $API_KEY"

# submitted、queued、processing、in_progress 表示仍在处理中，不要重复创建任务
# status 为 completed 后读取顶层 url；failed 时查看 error
# 将本站任务查询响应中的实际视频地址填入下方，不使用 /content
VIDEO_URL="YOUR_VIDEO_URL"
curl -L "$VIDEO_URL" -o result.mp4`

const wanVideoCreateCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "wan3.0x-480p",
    "prompt": "一只橘猫在月光下的屋顶上奔跑，远处城市霓虹闪烁，镜头平稳跟随",
    "duration": 4,
    "ratio": "16:9"
  }'`

const wanVideoImageCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "wan3.0x",
    "prompt": "保持 @Image1 中的产品外观与颜色，镜头缓慢环绕，生成自然光下的产品展示视频",
    "duration": 8,
    "ratio": "9:16",
    "image_url": "YOUR_PUBLIC_IMAGE_URL"
  }'`

const wanVideoReferenceCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "wan3.0x-1080p",
    "prompt": "保持 @Image1 中的人物外观，参考 @Video1 的动作和 @Audio1 的节奏，生成流畅的电影感视频",
    "duration": 12,
    "ratio": "16:9",
    "image_urls": ["YOUR_PUBLIC_IMAGE_URL"],
    "video_urls": ["YOUR_PUBLIC_VIDEO_URL"],
    "audio_urls": ["YOUR_PUBLIC_AUDIO_URL"]
  }'`

const wanVideoPollCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"
TASK_ID="YOUR_TASK_ID"

# 使用创建响应中的 task_id 或 id，建议每 10 秒查询一次
curl "$API_URL/v1/videos/$TASK_ID" \\
  -H "Authorization: Bearer $API_KEY"

# status 为 completed 或 succeeded 后，优先读取 metadata.url，其次 url
# 将本站任务查询响应中的实际视频地址填入下方，再播放或下载
VIDEO_URL="YOUR_VIDEO_URL"
curl -L "$VIDEO_URL" -o result.mp4`

const miniMaxVideoCreateCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "MiniMax-H3",
    "content": [
      {
        "type": "text",
        "text": "史诗级太空歌剧预告，女舰长站在观景窗前，镜头缓慢推进"
      }
    ],
    "resolution": "2K",
    "duration": 5,
    "ratio": "16:9"
  }'`

const miniMaxAssetUploadCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

# 分别上传本地参考图片、视频和音频，保存每次响应中的 data.url
curl "$API_URL/pg/assets" \\
  -H "Authorization: Bearer $API_KEY" \\
  -F "kind=image" \\
  -F "file=@first-frame.jpg"

curl "$API_URL/pg/assets" \\
  -H "Authorization: Bearer $API_KEY" \\
  -F "kind=video" \\
  -F "file=@motion-reference.mp4"

curl "$API_URL/pg/assets" \\
  -H "Authorization: Bearer $API_KEY" \\
  -F "kind=audio" \\
  -F "file=@music-reference.mp3"`

const miniMaxFrameVideoCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "MiniMax-H3",
    "content": [
      {"type": "text", "text": "从首帧自然过渡到尾帧，镜头平稳推进"},
      {
        "type": "image_url",
        "image_url": {"url": "https://YOUR_API_HOST/pg/assets/ASSET_ID/first-frame.jpg"},
        "role": "first_frame"
      },
      {
        "type": "image_url",
        "image_url": {"url": "https://YOUR_API_HOST/pg/assets/ASSET_ID/last-frame.jpg"},
        "role": "last_frame"
      }
    ],
    "resolution": "2K",
    "duration": 5,
    "ratio": "adaptive"
  }'`

const miniMaxReferenceVideoCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "MiniMax-H3",
    "content": [
      {"type": "text", "text": "保持参考人物和场景风格，使用参考动作节奏生成视频"},
      {
        "type": "image_url",
        "image_url": {"url": "https://YOUR_API_HOST/pg/assets/ASSET_ID/reference.jpg"},
        "role": "reference_image"
      },
      {
        "type": "video_url",
        "video_url": {"url": "https://YOUR_API_HOST/pg/assets/ASSET_ID/reference.mp4"},
        "role": "reference_video"
      },
      {
        "type": "audio_url",
        "audio_url": {"url": "https://YOUR_API_HOST/pg/assets/ASSET_ID/reference.mp3"},
        "role": "reference_audio"
      }
    ],
    "resolution": "768P",
    "duration": 10,
    "ratio": "adaptive"
  }'`

const miniMaxVideoPollCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"
TASK_ID="YOUR_TASK_ID"

# task.status 为 succeeded 后再下载
curl "$API_URL/v1/videos/$TASK_ID" \\
  -H "Authorization: Bearer $API_KEY"

curl -L "$API_URL/v1/videos/$TASK_ID/content" \\
  -H "Authorization: Bearer $API_KEY" \\
  -o minimax-h3.mp4`

const geminiNativeCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1beta/models/gemini-3-pro-image-preview:generateContent" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "contents": [{
      "role": "user",
      "parts": [{"text": "赛博城市里的中国神龙，红金配色，电影海报质感"}]
    }],
    "generationConfig": {
      "maxOutputTokens": 8192,
      "responseModalities": ["TEXT", "IMAGE"],
      "imageConfig": {"aspectRatio": "16:9", "imageSize": "2K"}
    }
  }'`

const geminiNativeReferenceCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1beta/models/gemini-3-pro-image-preview:generateContent" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "contents": [{
      "role": "user",
      "parts": [
        {"text": "保留参考图主体，生成一张 9:16 电影海报"},
        {"inlineData": {"mimeType": "image/png", "data": "BASE64_IMAGE_DATA"}}
      ]
    }],
    "generationConfig": {
      "responseModalities": ["TEXT", "IMAGE"],
      "imageConfig": {"aspectRatio": "9:16", "imageSize": "4K"}
    }
  }'`

const geminiNativeFlashCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1beta/models/gemini-3.1-flash-image-preview:generateContent" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "contents": [{
      "role": "user",
      "parts": [{"text": "生成一张 16:9 的未来城市产品海报"}]
    }],
    "generationConfig": {
      "maxOutputTokens": 8192,
      "responseModalities": ["TEXT", "IMAGE"],
      "imageConfig": {"aspectRatio": "16:9", "imageSize": "2K"}
    }
  }'`

const grokImageCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/images/generations" \\\
  -H "Authorization: Bearer $API_KEY" \\\
  -H "Content-Type: application/json" \\\
  -d '{
    "model": "grok-imagine-image-quality",
    "prompt": "未来城市中的银色跑车，电影级光影",
    "resolution": "2k",
    "aspect_ratio": "16:9",
    "n": 1
  }'`

const grokImageEditCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/images/edits" \\\
  -H "Authorization: Bearer $API_KEY" \\\
  -H "Content-Type: application/json" \\\
  -d '{
    "model": "grok-imagine-image-quality",
    "prompt": "保留人物与服装特征，把背景改成夜晚的未来城市",
    "resolution": "2k",
    "aspect_ratio": "9:16",
    "images": [
      {"type": "image_url", "url": "data:image/jpeg;base64,BASE64_IMAGE_1"},
      {"type": "image_url", "url": "data:image/jpeg;base64,BASE64_IMAGE_2"}
    ],
    "n": 1
  }'`

const grokVideoCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos/generations" \\\
  -H "Authorization: Bearer $API_KEY" \\\
  -H "Content-Type: application/json" \\\
  -d '{
    "model": "grok-imagine-video",
    "prompt": "海浪拍打黑色礁石，镜头缓慢推进",
    "duration": 10,
    "aspect_ratio": "16:9",
    "resolution": "720p"
  }'`

const grokVideo15Curl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos/generations" \\\
  -H "Authorization: Bearer $API_KEY" \\\
  -H "Content-Type: application/json" \\\
  -d '{
    "model": "grok-imagine-video-1.5",
    "prompt": "让参考图中的人物自然转身，镜头轻微环绕",
    "duration": 10,
    "aspect_ratio": "9:16",
    "resolution": "720p",
    "image": {"image_url": "data:image/png;base64,BASE64_IMAGE_DATA"}
  }'`

const grokVideoMultiReferenceCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"

curl "$API_URL/v1/videos/generations" \\\
  -H "Authorization: Bearer $API_KEY" \\\
  -H "Content-Type: application/json" \\\
  -d '{
    "model": "grok-imagine-video",
    "prompt": "按参考图顺序展示产品不同角度，镜头平滑转场",
    "duration": 10,
    "aspect_ratio": "16:9",
    "resolution": "720p",
    "reference_images": [
      {"url": "data:image/jpeg;base64,BASE64_IMAGE_1"},
      {"url": "data:image/jpeg;base64,BASE64_IMAGE_2"},
      {"url": "data:image/jpeg;base64,BASE64_IMAGE_3"}
    ]
  }'`

const grokVideoPollCurl = `API_URL="${apiBaseUrl}"
API_KEY="YOUR_API_KEY"
TASK_ID="YOUR_TASK_ID"

curl "$API_URL/v1/videos/$TASK_ID" \\\
  -H "Authorization: Bearer $API_KEY"

curl -L "$API_URL/v1/videos/$TASK_ID/content" \\\
  -H "Authorization: Bearer $API_KEY" \\\
  -o result.mp4`

const modelDocs: ModelDoc[] = [
  {
    id: 'openai-gpt-image2',
    name: 'OpenAI gpt-image2',
    provider: 'OpenAI',
    description: 'gpt-image-2、gpt-image-2.5-sunburst 与 gpt-image-2.5-flare 使用同一套同步文生图、参考图生成和 PNG 遮罩编辑接口。',
    endpoints: [
      'POST /v1/images/generations',
      'POST /v1/images/edits',
    ],
    availableModels: [
      { name: 'gpt-image-2', description: '文生图、参考图生成和遮罩编辑', specification: '1K / 2K / 4K；最多 16 张参考图', aspectRatio: '由 size 的像素宽高指定' },
      { name: 'gpt-image-2.5-sunburst', description: '文生图、参考图生成和遮罩编辑', specification: '1K / 2K / 4K；最多 16 张参考图', aspectRatio: '由 size 的像素宽高指定' },
      { name: 'gpt-image-2.5-flare', description: '文生图、参考图生成和遮罩编辑', specification: '1K / 2K / 4K；最多 16 张参考图', aspectRatio: '由 size 的像素宽高指定' },
    ],
    params: [
      { name: 'model / prompt', type: 'string', description: '图片接口必填，model 可传 gpt-image-2、gpt-image-2.5-sunburst 或 gpt-image-2.5-flare，prompt 填写生成或编辑要求。' },
      { name: 'size', type: 'string', description: '图片输出像素尺寸，例如 1280x1280、2048x1152、3840x2160；也可传 auto。' },
      { name: 'quality', type: 'string', description: 'gpt-image-2 支持 low、medium、high；gpt-image-2.5-sunburst 和 gpt-image-2.5-flare 额外支持 xhigh、max。' },
      { name: 'background', type: 'string', description: '可选 transparent，用于生成透明背景；使用时请同时设置 output_format=png。' },
      { name: 'output_format', type: 'string', description: '可选 png、jpeg、webp；background=transparent 时必须使用 png。' },
      { name: 'output_compression', type: 'number', description: '仅 jpeg、webp 使用；png 不要传该字段。' },
      { name: 'response_format', type: 'string', description: '可选 b64_json、url，用于声明期望返回格式；不同上游可能返回 data[].b64_json、data[].url，或同时返回两者，客户端需要兼容判断。' },
      { name: 'moderation', type: 'string', description: '可选 auto、low。' },
      { name: 'n', type: 'number', description: '生成数量，可选，默认 1。' },
      { name: 'image', type: 'file', description: '仅 /v1/images/edits，multipart 上传参考图；最多 16 张，多图时重复使用 image 字段。' },
      { name: 'mask', type: 'file', description: '仅 /v1/images/edits，可选 PNG 遮罩，对应第一张参考图。' },
    ],
    examples: [
      { title: 'gpt-image-2 文生图', language: 'curl', code: gptImage2GenerateCurl },
      { title: 'gpt-image-2 图片编辑', language: 'curl', code: gptImage2EditCurl },
      { title: '兼容解析 Base64 和 URL 图片', language: 'javascript', code: gptImage2ResponseHandlingJavascript },
    ],
    notes: [
      '文生图使用 JSON；参考图和遮罩编辑使用 multipart/form-data，不要手写 multipart 的 Content-Type boundary。',
      '三个 OpenAI 图片模型的接口和调用方式一致；多张参考图请重复提交 image 字段，遮罩图使用 PNG，并与第一张参考图尺寸一致。',
      '需要透明底时传 background=transparent，并将 output_format 设置为 png；这是上游图片接口参数，不需要把“透明背景”拼进模型名称。',
      'png 不需要传 output_compression；jpeg、webp 可按需设置压缩质量。同步响应可能返回 data[].b64_json、data[].url，或同时返回两者，请按示例兼容解析。',
    ],
  },
  {
    id: 'openai-mapped-image-models',
    name: 'OpenAI 兼容 Seedream / Midjourney',
    provider: 'OpenAI API URL + Key',
    description: 'Seedream 5.0 Pro、Seedream 5.0 Pro X 与 Midjourney V7 使用本站统一的 OpenAI 图片接口，支持文生图和参考图编辑。',
    endpoints: [
      'POST /v1/images/generations',
      'POST /v1/images/edits',
    ],
    availableModels: [
      { name: 'seedream-5.0-pro', description: 'Seedream 5.0 Pro', specification: '提示词最多 8000 字符；最多 10 张参考图', aspectRatio: '16:9 / 9:16 / 1:1 / 4:3 / 3:4 / 3:2 / 2:3 / 21:9' },
      { name: 'seedream-5.0-pro-x', description: 'Seedream 5.0 Pro X', specification: '提示词最多 8000 字符；最多 10 张参考图', aspectRatio: '16:9 / 9:16 / 1:1 / 4:3 / 3:4 / 3:2 / 2:3 / 21:9' },
      { name: 'Midjourney-V7', description: 'Midjourney V7', specification: '提示词最多 4000 字符；最多 5 张参考图；可能返回多张变体', aspectRatio: '16:9 / 9:16 / 1:1 / 4:3 / 3:4 / 4:5 / 5:4 / 3:2 / 2:3 / 21:9' },
    ],
    params: [
      { name: 'model / prompt', type: 'string', description: '必填。model 使用上表中的完整模型名称；Seedream 提示词最多 8000 字符，Midjourney V7 最多 4000 字符。' },
      { name: 'size', type: 'string', description: '可选，用宽高表达比例，例如 1024x576（16:9）、1024x1024（1:1）；实际输出尺寸以返回图片为准。' },
      { name: 'n', type: 'number', description: '可选，默认 1。Midjourney V7 可能一次返回多张变体，客户端应解析完整 data 数组。' },
      { name: 'response_format', type: 'string', description: '可选 b64_json、url。建议使用 b64_json；客户端仍应兼容 data[].b64_json 和 data[].url 两种返回格式。' },
      { name: 'image', type: 'file', description: '仅 /v1/images/edits，multipart 上传参考图；Seedream 最多 10 张，Midjourney V7 最多 5 张，多图时重复使用 image 字段。' },
    ],
    examples: [
      { title: 'Seedream 5.0 Pro 文生图', language: 'curl', code: mappedImageGenerateCurl },
      { title: 'Midjourney V7 参考图编辑', language: 'curl', code: mappedImageEditCurl },
      { title: '兼容解析 Base64 和 URL 图片', language: 'javascript', code: gptImage2ResponseHandlingJavascript },
    ],
    notes: [
      '文生图使用 JSON；参考图编辑使用 multipart/form-data，不要手写 multipart 的 Content-Type boundary。',
      '这三个模型只使用 OpenAI API URL + Key 账号；模型映射后的完整名称会原样转发到上游。',
      '这三个模型按比例选择尺寸，不使用 GPT 专属的 quality、moderation、background、output_format 或 output_compression 参数，也不支持 mask 遮罩编辑。',
      '参考图请使用 image 字段重复上传；响应可能只返回 data[].b64_json、只返回 data[].url，或同时返回两者，请按示例兼容解析。',
    ],
  },
  {
    id: 'gemini-nano-banana',
    name: 'Gemini Nano Banana',
    provider: 'Gemini 图片模型',
    description: '包含 Nano Banana Pro 和 Nano Banana 2。两个模型使用相同的 Gemini 原生 generateContent 请求体、参考图格式和响应解析。',
    endpoints: [
      'POST /v1beta/models/gemini-3-pro-image-preview:generateContent',
      'POST /v1beta/models/gemini-3.1-flash-image-preview:generateContent',
    ],
    availableModels: [
      { name: 'gemini-3-pro-image-preview', description: 'Nano Banana Pro 原生模型', specification: '2K / 4K；最多 9 张参考图', aspectRatio: '1:1 / 16:9 / 9:16 / 21:9 等' },
      { name: 'gemini-3.1-flash-image-preview', description: 'Nano Banana 2 原生模型', specification: '2K / 4K；最多 9 张参考图', aspectRatio: '1:1 / 16:9 / 9:16 / 21:9 等' },
    ],
    params: [
      { name: 'contents', type: 'array', description: 'Gemini 原生接口必填。提示词放 parts.text，参考图放 parts.inlineData，最多 9 张。' },
      { name: 'generationConfig.responseModalities', type: 'string[]', description: 'Gemini 原生图片生成传 ["TEXT", "IMAGE"]。' },
      { name: 'generationConfig.imageConfig', type: 'object', description: '两个原生模型都使用 aspectRatio 和 imageSize，支持 2K、4K。' },
    ],
    examples: [
      { title: 'Nano Banana Pro 原生文生图', language: 'curl', code: geminiNativeCurl },
      { title: 'Nano Banana Pro 原生参考图生成', language: 'curl', code: geminiNativeReferenceCurl },
      { title: 'Nano Banana 2 原生文生图', language: 'curl', code: geminiNativeFlashCurl },
    ],
    notes: [
      '两个原生模型的图片都位于 candidates[].content.parts[].inlineData；将 data 按 base64 解码即可保存。',
      '原生参考图只传纯 base64 到 inlineData.data，不要包含 data:image/...;base64, 前缀。',
      'gemini-3.1-flash-image-preview 必须走 Gemini 原生接口；两个模型都不支持 mask 遮罩编辑。',
    ],
  },
  {
    id: 'strongest-video-v2',
    name: '视频模型SD2.0',
    provider: 'Firefly Video v2 / ViralDance 933 / viraldance2.5-30',
    description: '支持文生视频和多模态参考生成。Firefly 支持分镜与文件上传，ViralDance 933 与 viraldance2.5-30 使用 JSON 与素材 URL；后者额外支持首尾帧和更多参考素材。任务创建后异步查询状态，完成后播放或下载视频。',
    endpoints: [
      'POST /v1/videos',
      'GET /v1/videos/{task_id}',
      'HEAD /v1/videos/{task_id}/content',
      'GET /v1/videos/{task_id}/content',
    ],
    availableModels: [
      { name: 'firefly-video-v2-fast', description: '快速版', specification: '5 / 10 / 15 秒；480p / 720p', aspectRatio: '21:9 / 16:9 / 4:3 / 1:1 / 3:4 / 9:16' },
      { name: 'firefly-video-v2', description: '标准版', specification: '5 / 10 / 15 秒；480p / 720p / 1080p', aspectRatio: '21:9 / 16:9 / 4:3 / 1:1 / 3:4 / 9:16' },
      { name: 'viraldance933', description: 'ViralDance 933 标准版', specification: '4 至 15 秒（整数）；720p', aspectRatio: '16:9 / 9:16 / 1:1 / 4:3' },
      { name: 'viraldance933-fast', description: 'ViralDance 933 快速版', specification: '4 至 15 秒（整数）；720p', aspectRatio: '16:9 / 9:16 / 1:1 / 4:3' },
      { name: 'viraldance2.5-30', description: '文生视频、首尾帧和多模态参考生成', specification: '4 至 30 秒（整数）；720p；最多 30 图 / 10 视频 / 10 音频', aspectRatio: '16:9 / 9:16 / 1:1' },
      { name: 'dola-viraldance2.0', description: 'Dola ViralDance 2.0', specification: '4 至 15 秒（整数）；720p', aspectRatio: '16:9 / 9:16 / 1:1 / 4:3' },
      { name: 'dola-viraldance2.5', description: 'Dola ViralDance 2.5', specification: '4 至 30 秒（整数）；720p', aspectRatio: '16:9 / 9:16 / 1:1 / 4:3' },
    ],
    params: [
      { name: 'model', type: 'string', description: '必填，填写上表中的完整模型名称。viraldance2.5-30 与 dola-viraldance2.5 的素材参数不同，请勿混用。Dola 沿用 ViralDance 933 的 JSON 格式和素材参数，时长按各自模型范围填写。' },
      { name: 'prompt', type: 'string', description: 'viraldance2.5-30 必填，长度 1 至 5000 字符；ViralDance 933 必填，长度 1 至 2500 字符；Firefly 与 shots 二选一，长度 1 至 2500 字符。' },
      { name: 'negative_prompt', type: 'string', description: 'Firefly 可选，最多 2500 字符。' },
      { name: 'duration', type: 'integer', description: 'ViralDance 933 与 Dola 2.0 必填 4 至 15 的整数；Dola 2.5 必填 4 至 30 的整数；viraldance2.5-30 支持 4 至 30 的整数，默认 5；Firefly 可选 5、10、15，默认 5，multipart 中使用字符串。' },
      { name: 'resolution', type: 'string', description: '默认 720p。ViralDance 933、Dola 和 viraldance2.5-30 仅支持 720p；Firefly 快速版支持 480p、720p，标准版额外支持 1080p。' },
      { name: 'aspect_ratio', type: 'string', description: 'viraldance2.5-30 使用，支持 16:9、9:16、1:1，默认 16:9，不使用 ratio；Firefly 可选 21:9、16:9、4:3、1:1、3:4、9:16，默认 16:9。' },
      { name: 'ratio', type: 'string', description: 'ViralDance 933 必填，顶层字段，可选 16:9、9:16、1:1、4:3。' },
      { name: 'generate_audio', type: 'boolean', description: 'ViralDance 933 当前固定为 true，可省略；Firefly 可选 true 或 false，multipart 中使用字符串；viraldance2.5-30 不传此参数。' },
      { name: 'async', type: 'boolean', description: 'viraldance2.5-30 传 true，创建异步任务后通过任务 ID 查询结果。' },
      { name: 'shots', type: 'array', description: 'Firefly 可选 2 至 15 个分镜，每项包含 prompt 和 duration，分镜时长之和必须等于总时长。' },
      { name: 'first_frame / last_frame', type: 'file', description: 'Firefly multipart 首帧、尾帧图片，各最多 1 张。' },
      { name: 'images / videos / audios', type: 'file[] / string[]', description: 'Firefly 使用 multipart 重复文件字段；ViralDance 933 使用 JSON URL 数组。图片合计最多 9 张，视频最多 3 个，音频最多 3 个。' },
      { name: 'image_url / image_urls', type: 'string / string[]', description: 'ViralDance 933 单张或多张参考图片的公网 URL，最多 9 张；多图推荐 image_urls。viraldance2.5-30 普通参考图使用 image_urls，与首尾帧合计最多 30 张。' },
      { name: 'start_image_url / end_image_url', type: 'string', description: 'viraldance2.5-30 可选首帧、尾帧图片的公网 URL，各最多 1 张，计入 30 张图片总数，不占普通参考图编号。' },
      { name: 'video_reference', type: 'object[]', description: 'viraldance2.5-30 参考视频，格式 [{"url":"公网视频地址"}]，最多 10 个；单个 3 至 10 秒，合计不超过 30 秒。' },
      { name: 'audio_reference', type: 'object[]', description: 'viraldance2.5-30 参考音频，格式 [{"url":"公网音频地址"}]，最多 10 个；单个 3 至 30 秒，合计不超过 30 秒，可单独作为参考。' },
      { name: 'video_url / video_urls', type: 'string / string[]', description: 'ViralDance 933 单条或多条参考视频的公网 URL，最多 3 条；多视频推荐 video_urls。' },
      { name: 'audio_url / audio_urls', type: 'string / string[]', description: 'ViralDance 933 单条或多条参考音频的公网 URL，最多 3 条；必须同时提供图片或视频。' },
      { name: 'elements', type: 'object[]', description: 'ViralDance 933 高级主体绑定，最多 20 项；每项至少包含 frontal_image_url 或非空的 reference_image_urls 数组。' },
    ],
    examples: [
      { title: 'Firefly 文生视频', language: 'curl', code: fireflyVideoCreateCurl },
      { title: 'Firefly 分镜视频', language: 'curl', code: fireflyVideoShotsCurl },
      { title: 'Firefly 参考素材上传', language: 'curl', code: fireflyVideoReferenceCurl },
      { title: 'Firefly 查询、检查和下载', language: 'curl', code: fireflyVideoPollCurl },
      { title: 'ViralDance 933 文生视频', language: 'curl', code: viraldanceVideoCreateCurl },
      { title: 'ViralDance 933 Fast 多模态参考', language: 'curl', code: viraldanceVideoReferenceCurl },
      { title: 'ViralDance 933 查询和下载', language: 'curl', code: viraldanceVideoPollCurl },
      { title: 'viraldance2.5-30 文生视频', language: 'curl', code: viraldance25VideoCreateCurl },
      { title: 'viraldance2.5-30 首尾帧与多模态参考', language: 'curl', code: viraldance25VideoReferenceCurl },
      { title: 'viraldance2.5-30 查询和下载', language: 'curl', code: viraldance25VideoPollCurl },
    ],
    notes: [
      'Firefly 没有参考素材时使用 JSON；包含参考素材时必须使用 multipart/form-data 真实文件字段，并让客户端自动生成 boundary。',
      'Firefly 参考素材不接受 URL、Base64、asset_id 或本地路径字符串；多张同类素材需重复提交同名字段。',
      'Firefly 图片支持 JPEG、PNG、WebP；视频支持 MP4、MOV；音频支持 MP3、WAV。整个请求最大 384 MiB，服务端和反向代理需要同步配置上传限制。',
      'Firefly 建议每 3 至 5 秒查询一次，状态 completed 后通过 /content 下载；Range 下载支持拖动播放和断点续传。',
      'ViralDance 933 两个模型均使用 application/json，参考素材填写公网可直接下载的 URL。请将示例中的 example.com 地址替换为自己的素材地址；单图可只传 image_url，音频不能单独作为参考。',
      'ViralDance 933 素材引用从 1 开始，例如 @Image1、@Video1、@Audio1，分别对应素材数组中的第一项；提示词中的引用需要同时提交对应 URL。',
      '创建成功只代表任务已受理。ViralDance 933 保存 task_id 或 id，每 10 秒查询一次；queued、processing、in_progress 时继续等待，不要重复创建任务。',
      'ViralDance 933 的 status 为 succeeded 或 completed 后，优先读取 video.url，其次 result_url 或 url，直接播放或下载；/content 文件接口仅适用于本节的 Firefly。failed 表示失败，可查看 error。',
      'viraldance2.5-30 使用 application/json 和公网可直接下载的素材 URL，不直接传本地路径、Base64 或 multipart 文件。示例中的 example.com 地址须替换为自己的素材地址；可按需删除不使用的参考字段，单图也使用 image_urls 数组。',
      'viraldance2.5-30 普通图片、视频、音频分别使用 @图片1、@视频1、@音频1，按同类数组顺序从 1 开始；首尾帧不占普通图片编号，用“首帧”“尾帧”描述。不要套用 933 的 ratio、video_urls、audio_urls、generate_audio，也不要额外传 size。',
      'viraldance2.5-30 图片支持 JPG、PNG、WebP；视频支持 MP4、MOV；音频支持 MP3、WAV、M4A、AAC、OGG、WebM。图片（含首尾帧）、视频、音频分别最多 30 张、10 个、10 个。',
      'viraldance2.5-30 创建后保存 task_id 或 id，使用本站和同一分组密钥每 5 秒查询一次。submitted、queued、processing、in_progress 时继续等待，不要重复创建；completed 后读取顶层 url 播放或下载，不使用 /content；failed 时查看 error。',
    ],
  },
  {
    id: 'wan3-video',
    name: '视频模型Wan3.0',
    provider: 'Wan 3.0',
    description: '支持文生视频、图生视频，以及图片、视频、音频混合参考。三个模型使用相同的 JSON 参数，通过模型名称选择分辨率，创建后异步查询生成结果。',
    endpoints: [
      'POST /v1/videos',
      'GET /v1/videos/{task_id}',
    ],
    availableModels: [
      { name: 'wan3.0x', description: '720p 视频生成', specification: '2 至 30 秒（整数）；720p', aspectRatio: '16:9 / 4:3 / 1:1 / 3:4 / 9:16' },
      { name: 'wan3.0x-480p', description: '480p 视频生成', specification: '2 至 30 秒（整数）；480p', aspectRatio: '16:9 / 4:3 / 1:1 / 3:4 / 9:16' },
      { name: 'wan3.0x-1080p', description: '1080p 视频生成', specification: '2 至 30 秒（整数）；1080p', aspectRatio: '16:9 / 4:3 / 1:1 / 3:4 / 9:16' },
    ],
    params: [
      { name: 'model', type: 'string', description: '必填，填写上表中的完整模型名称；分辨率由模型名称决定，无需传 resolution 或 size。' },
      { name: 'prompt', type: 'string', description: '必填，描述主体、动作、镜头和风格；图生视频和参考生成也必须填写。' },
      { name: 'duration', type: 'integer', description: '必填，2 至 30 秒的整数，使用 JSON 数值。' },
      { name: 'ratio', type: 'string', description: '必填，顶层字段，可选 16:9、4:3、1:1、3:4、9:16。' },
      { name: 'image_url / image_urls', type: 'string / string[]', description: '可选，单张或多张参考图片的公网 URL，最多 10 张；多图推荐 image_urls。' },
      { name: 'video_url / video_urls', type: 'string / string[]', description: '可选，单条或多条参考视频的公网 URL，最多 5 条；多视频推荐 video_urls。' },
      { name: 'audio_url / audio_urls', type: 'string / string[]', description: '可选，单条或多条参考音频的公网 URL，最多 5 条；多音频推荐 audio_urls。' },
      { name: 'images / videos / audios', type: 'string[]', description: '分别为 image_urls、video_urls、audio_urls 的兼容写法，均使用 JSON URL 数组。' },
    ],
    examples: [
      { title: 'Wan 3.0 文生视频（480p）', language: 'curl', code: wanVideoCreateCurl },
      { title: 'Wan 3.0 单图生视频（720p）', language: 'curl', code: wanVideoImageCurl },
      { title: 'Wan 3.0 多模态参考（1080p）', language: 'curl', code: wanVideoReferenceCurl },
      { title: 'Wan 3.0 查询和下载', language: 'curl', code: wanVideoPollCurl },
    ],
    notes: [
      '将 YOUR_API_URL 替换为本站 API 地址，YOUR_API_KEY 替换为已开通对应模型的本站密钥。三个模型的参数相同，可按所需分辨率替换 model。',
      '所有请求使用 application/json。将 YOUR_PUBLIC_IMAGE_URL、YOUR_PUBLIC_VIDEO_URL、YOUR_PUBLIC_AUDIO_URL 替换为自己的素材地址，素材必须能在公网直接下载，无需登录或 Cookie。',
      '素材编号按同类数组顺序从 1 开始，例如 @Image1、@Video1、@Audio1；提示词中的引用需要同时提交对应素材 URL。',
      '创建成功只代表任务已受理。保存 task_id 或 id，继续使用本站地址和同一分组密钥，每 10 秒查询一次；queued、processing、in_progress 时继续等待，不要重复创建任务。',
      'status 为 completed 或 succeeded 后，优先读取 metadata.url，其次 url，使用返回地址播放或下载；本组模型不使用 /content 文件接口。failed 表示失败，可查看 error。',
    ],
  },
  {
    id: 'minimax-h3-video',
    name: '视频模型MiniMax-H3',
    provider: 'MiniMax H3',
    description: '支持文生视频、首尾帧生视频和多模态参考生成。请求使用 JSON content[]，创建成功后通过任务 ID 查询并下载视频。',
    endpoints: [
      'POST /pg/assets',
      'POST /v1/videos',
      'GET /v1/videos/{task_id}',
      'GET /v1/videos/{task_id}/content',
    ],
    availableModels: [
      { name: 'MiniMax-H3', description: '文生视频、首尾帧和多模态参考视频', specification: '4 至 15 秒；768P / 2K', aspectRatio: '21:9 / 16:9 / 4:3 / 1:1 / 3:4 / 9:16 / adaptive' },
    ],
    params: [
      { name: 'model', type: 'string', description: '必填，固定填写 MiniMax-H3，注意大小写。' },
      { name: 'content', type: 'array', description: '必填。提示词和参考素材都放在数组中，提示词使用 text 类型。' },
      { name: 'content[].type', type: 'string', description: '可选 text、image_url、video_url、audio_url；URL 放入对应的 image_url.url、video_url.url 或 audio_url.url。' },
      { name: 'content[].role', type: 'string', description: '素材必须指定角色：first_frame、last_frame、reference_image、reference_video 或 reference_audio。' },
      { name: 'resolution', type: 'string', description: '必填，可选 768P 或 2K。' },
      { name: 'duration', type: 'integer', description: '必填，4 至 15 秒的整数。' },
      { name: 'ratio', type: 'string', description: '可选 21:9、16:9、4:3、1:1、3:4、9:16；首尾帧或多参考模式可使用 adaptive。' },
      { name: 'task_id', type: 'string', description: '创建成功后返回，用于查询 task.status 和下载视频。' },
    ],
    examples: [
      { title: 'MiniMax-H3 文生视频', language: 'curl', code: miniMaxVideoCreateCurl },
      { title: '上传 MiniMax-H3 参考素材', language: 'curl', code: miniMaxAssetUploadCurl },
      { title: 'MiniMax-H3 首尾帧生视频', language: 'curl', code: miniMaxFrameVideoCurl },
      { title: 'MiniMax-H3 多模态参考生视频', language: 'curl', code: miniMaxReferenceVideoCurl },
      { title: '查询和下载 MiniMax-H3 视频', language: 'curl', code: miniMaxVideoPollCurl },
    ],
    notes: [
      '只有提示词时直接提交 JSON。使用本地素材时先调用 /pg/assets，随后把响应中的 data.url 放入对应 content[] URL 字段。',
      '文/图生视频模式最多使用 1 张首帧和 1 张尾帧；多参考模式最多使用 9 张参考图片、3 个参考视频和 3 个参考音频，两种模式不能混用。',
      '参考图片支持 JPEG、PNG、WebP；参考视频支持 MP4、MOV，单个不超过 50MB、2 至 15 秒且合计不超过 15 秒；参考音频支持 MP3、WAV，单个不超过 15MB、2 至 15 秒且合计不超过 15 秒。',
      '参考视频必须使用本站 /pg/assets 地址以核验计费时长；参考图片和音频不额外计费，参考视频按实际秒数计费。',
      '创建成功只表示任务已受理。建议每 3 至 5 秒查询一次；task.status 为 succeeded 后再下载，failed 或 cancelled 表示任务失败。',
    ],
  },
  {
    id: 'grok-image-video',
    name: 'Grok 图片视频模型',
    provider: 'xAI Grok',
    description: 'Grok 图片使用同步生成接口；Grok 视频使用异步任务接口，提交后查询状态，完成后下载 mp4。',
    endpoints: [
      'POST /v1/images/generations',
      'POST /v1/images/edits',
      'POST /v1/videos/generations',
      'GET /v1/videos/{task_id}',
      'GET /v1/videos/{task_id}/content',
    ],
    availableModels: [
      { name: 'grok-imagine-image-quality', description: '高质量文生图与参考图生成', specification: '1K / 2K；最多 3 张参考图', aspectRatio: '1:1 / 16:9 / 9:16 / 2:3 等' },
      { name: 'grok-imagine-video-1.5', description: '单图生视频；纯文本自动使用标准版', specification: '6 / 10 / 15 秒，480p / 720p', aspectRatio: '2:3 / 3:2 / 1:1 / 9:16 / 16:9' },
      { name: 'grok-imagine-video', description: '文生视频或多图参考视频', specification: '6 / 10 / 15 秒；多图限 6 / 10 秒', aspectRatio: '2:3 / 3:2 / 1:1 / 9:16 / 16:9' },
    ],
    params: [
      { name: 'model / prompt', type: 'string', description: '必填，填写上表模型名和生成提示词。' },
      { name: 'resolution', type: 'string', description: '图片传 1k 或 2k；视频传 480p 或 720p。' },
      { name: 'aspect_ratio', type: 'string', description: '必填或建议填写，使用上表支持的画幅比例。' },
      { name: 'n', type: 'number', description: '图片生成数量，可选，默认 1。' },
      { name: 'duration', type: 'number', description: '视频时长，可选 6、10、15；多图参考只能使用 6 或 10 秒。' },
      { name: 'image / images', type: 'object / object[]', description: 'Grok 图生图使用；单图传 image，多图传 images，格式为 {"type":"image_url","url":"data:image/...;base64,..."}，最多 3 张。' },
      { name: 'image', type: 'object', description: 'Grok 1.5 单图生视频使用，格式为 {"image_url":"data:image/...;base64,..."}。' },
      { name: 'reference_images', type: 'object[]', description: '2 至 7 张参考图使用 grok-imagine-video，格式为 [{"url":"data:image/...;base64,..."}]。' },
      { name: 'task_id', type: 'string', description: '视频提交后返回，用于查询和下载。' },
    ],
    examples: [
      { title: 'Grok 高质量图片', language: 'curl', code: grokImageCurl },
      { title: 'Grok 图生图（1 至 3 张参考图）', language: 'curl', code: grokImageEditCurl },
      { title: 'Grok 文生视频', language: 'curl', code: grokVideoCurl },
      { title: 'Grok 1.5 单图生视频', language: 'curl', code: grokVideo15Curl },
      { title: 'Grok 多图生视频', language: 'curl', code: grokVideoMultiReferenceCurl },
      { title: '查询和下载视频', language: 'curl', code: grokVideoPollCurl },
    ],
    notes: [
      'grok-imagine-video-1.5 最多使用 1 张参考图；2 至 7 张参考图请使用 grok-imagine-video。',
      'grok-imagine-image-quality 图生图最多使用 3 张参考图，不支持 mask 遮罩编辑。',
      '纯文生视频可直接使用 grok-imagine-video；视频创建成功后不会同步返回 mp4。',
    ],
  },
]
</script>

<style scoped>
.api-docs-page {
  @apply min-h-[calc(100vh-8rem)] rounded-[2rem] border border-[#e8e0d3] bg-[#fbf8f2] p-4 shadow-[0_28px_70px_rgba(15,23,42,0.08)] dark:border-[#2a3039] dark:bg-[#10141a] md:p-6;
}

.api-docs-header {
  @apply flex flex-col gap-5 rounded-[1.75rem] border border-white/70 bg-white/90 p-5 shadow-[0_18px_45px_rgba(15,23,42,0.06)] dark:border-[#242933] dark:bg-[#141920]/90 md:flex-row md:items-center md:justify-between md:p-7;
}

.api-docs-heading {
  @apply flex min-w-0 flex-col gap-4 sm:flex-row sm:items-center;
}

.api-docs-icon {
  @apply flex h-16 w-16 shrink-0 items-center justify-center rounded-[1.35rem] bg-[#f3eadc] text-[#7c5b2d] dark:bg-[#242a33] dark:text-[#d9c292];
}

.api-docs-kicker {
  @apply text-xs font-semibold uppercase tracking-[0.22em] text-gray-500 dark:text-gray-400;
}

.api-docs-title {
  @apply mt-2 text-2xl font-semibold text-gray-900 dark:text-white md:text-3xl;
}

.api-docs-description {
  @apply mt-3 max-w-3xl text-sm leading-7 text-gray-600 dark:text-gray-300;
}

.api-docs-note code {
  @apply rounded-lg bg-white px-2 py-1 text-xs text-[#8a2525] dark:bg-dark-900 dark:text-[#f0b4a8];
}

.api-docs-note {
  @apply mt-4 flex gap-3 rounded-2xl border border-[#eadfcd] bg-[#fff7ed] p-4 text-sm leading-7 text-gray-700 dark:border-[#5f4632] dark:bg-[#2b231a] dark:text-gray-200;
}

.api-docs-section {
  @apply mt-6 rounded-[1.75rem] border border-white/70 bg-white/90 p-5 shadow-[0_18px_45px_rgba(15,23,42,0.06)] dark:border-[#242933] dark:bg-[#141920]/90 md:p-7;
}

.api-docs-model-header {
  @apply flex flex-col gap-4 border-b border-[#efe8dd] pb-5 dark:border-dark-700 lg:flex-row lg:items-start lg:justify-between;
}

.api-docs-model-eyebrow {
  @apply text-xs font-semibold uppercase tracking-[0.18em] text-[#a33a2b] dark:text-[#f0b4a8];
}

.api-docs-model-title {
  @apply mt-2 text-2xl font-semibold text-gray-900 dark:text-white;
}

.api-docs-model-description {
  @apply mt-3 max-w-3xl text-sm leading-7 text-gray-600 dark:text-gray-300;
}

.api-docs-endpoints {
  @apply flex flex-wrap gap-2 lg:justify-end;
}

.api-docs-endpoint {
  @apply rounded-full border border-[#eadfcd] bg-[#fbf8f3] px-3 py-1.5 font-mono text-xs text-gray-700 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200;
}

.api-docs-subsection {
  @apply mt-6;
}

.api-docs-subtitle {
  @apply text-base font-semibold text-gray-900 dark:text-white;
}

.api-docs-table-wrap {
  @apply mt-3 overflow-x-auto rounded-2xl border border-[#eadfcd] dark:border-dark-600;
}

.api-docs-table {
  @apply min-w-full divide-y divide-[#eadfcd] text-left text-sm dark:divide-dark-600;
}

.api-docs-table th {
  @apply bg-[#fbf8f3] px-4 py-3 text-xs font-semibold uppercase tracking-[0.12em] text-gray-500 dark:bg-dark-800 dark:text-gray-400;
}

.api-docs-table td {
  @apply px-4 py-3 align-top leading-6 text-gray-700 dark:text-gray-200;
}

.api-docs-table tr {
  @apply border-b border-[#f1e8dc] last:border-0 dark:border-dark-700;
}

.api-docs-table code {
  @apply rounded-lg bg-[#fbf8f3] px-2 py-1 font-mono text-xs text-[#8a2525] dark:bg-dark-800 dark:text-[#f0b4a8];
}

.api-docs-code-grid {
  @apply mt-3 grid gap-4 xl:grid-cols-2;
}

.api-docs-code-card {
  @apply overflow-hidden rounded-2xl border border-[#eadfcd] bg-[#10141a] text-gray-100 shadow-sm dark:border-dark-600;
}

.api-docs-code-header {
  @apply flex items-start justify-between gap-3 border-b border-white/10 bg-white/5 px-4 py-3;
}

.api-docs-code-language {
  @apply text-xs font-semibold uppercase tracking-[0.16em] text-[#f0b4a8];
}

.api-docs-code-title {
  @apply mt-1 text-sm font-semibold text-white;
}

.api-docs-copy-button {
  @apply inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-white/10 bg-white/10 text-gray-200 transition hover:bg-white/15;
}

.api-docs-code-card pre {
  @apply max-h-[32rem] overflow-auto p-4 text-xs leading-6;
}

.api-docs-code-card code {
  @apply whitespace-pre font-mono;
}

.api-docs-model-notes {
  @apply mt-5 space-y-2 rounded-2xl border border-[#eadfcd] bg-[#fbf8f3] p-4 text-sm leading-7 text-gray-700 dark:border-dark-600 dark:bg-dark-800/80 dark:text-gray-200;
}
</style>
