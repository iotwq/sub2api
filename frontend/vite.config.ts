import { defineConfig, loadEnv, Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import checker from 'vite-plugin-checker'
import { resolve } from 'path'

function escapeHtml(value: string): string {
  return value.replace(/[&<>"']/g, (character) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  })[character] || character)
}

function isSafeImageUrl(value: string): boolean {
  const trimmed = value.trim()
  if ((trimmed.startsWith('/') && !trimmed.startsWith('//')) || /^data:image\//i.test(trimmed)) {
    return true
  }
  try {
    const parsed = new URL(trimmed)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

function injectBranding(html: string, config: { site_name?: string; site_logo?: string }): string {
  let brandedHtml = html
  const siteName = config.site_name?.trim()
  if (siteName) {
    brandedHtml = brandedHtml.replace(
      /<title>[^<]*<\/title>/i,
      `<title>${escapeHtml(siteName)} - AI API Gateway</title>`,
    )
  }

  const siteLogo = config.site_logo?.trim()
  if (siteLogo && isSafeImageUrl(siteLogo)) {
    brandedHtml = brandedHtml.replace(
      /<link\s+rel=["']icon["'][^>]*>/i,
      `<link rel="icon" href="${escapeHtml(siteLogo)}" />`,
    )
  }
  return brandedHtml
}

/**
 * Vite 插件：开发模式下注入公开配置到 index.html
 * 与生产模式的后端注入行为保持一致，消除闪烁
 */
function injectPublicSettings(backendUrl: string): Plugin {
  return {
    name: 'inject-public-settings',
    apply: 'serve',
    transformIndexHtml: {
      order: 'pre',
      async handler(html) {
        try {
          const response = await fetch(`${backendUrl}/api/v1/settings/public`, {
            signal: AbortSignal.timeout(2000)
          })
          if (response.ok) {
            const data = await response.json()
            if (data.code === 0 && data.data) {
              const script = `<script>window.__APP_CONFIG__=${JSON.stringify(data.data)};</script>`
              return injectBranding(html, data.data).replace('</head>', `${script}\n</head>`)
            }
          }
        } catch (e) {
          console.warn('[vite] 无法获取公开配置，将回退到 API 调用:', (e as Error).message)
        }
        return html
      }
    }
  }
}

function normalizeModuleId(id: string): string {
  return id.replace(/\\/g, '/')
}

export default defineConfig(({ mode, command }) => {
  // 加载环境变量
  const env = loadEnv(mode, process.cwd(), '')
  const backendUrl = env.VITE_DEV_PROXY_TARGET || 'http://localhost:8080'
  const devPort = Number(env.VITE_DEV_PORT || 3000)

  return {
    plugins: [
      vue(),
      ...(command === 'serve'
        ? [
            checker({
              vueTsc: true
            })
          ]
        : []),
      injectPublicSettings(backendUrl)
    ],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
        // 使用 vue-i18n 运行时版本，v11 JIT 编译恒开启，无需 unsafe-eval
        'vue-i18n': 'vue-i18n/dist/vue-i18n.runtime.esm-bundler.js'
      }
    },
    build: {
      outDir: '../backend/internal/web/dist',
      emptyOutDir: true,
      rollupOptions: {
        output: {
          /**
           * 手动分包配置
           * 只分离第三方库；应用源码交给 Rollup 按路由依赖图自动分包。
           */
          manualChunks(id: string) {
            const moduleId = normalizeModuleId(id)

            // Rollup 的 CommonJS 兼容辅助代码被很多库共用，单独成包，
            // 避免它被并进某个功能库（之前并进了二维码库），让无关页面跟着下载。
            if (moduleId.includes('commonjsHelpers')) return 'vendor-cjs-helpers'

            if (moduleId.includes('/node_modules/')) {
              // Vue 核心库
              if (
                moduleId.includes('/vue/') ||
                moduleId.includes('/vue-router/') ||
                moduleId.includes('/pinia/') ||
                moduleId.includes('/@vue/')
              ) {
                return 'vendor-vue'
              }

              // 国际化
              if (moduleId.includes('/vue-i18n/') || moduleId.includes('/@intlify/')) {
                return 'vendor-i18n'
              }

              // 图表库
              if (moduleId.includes('/chart.js/') || moduleId.includes('/vue-chartjs/')) {
                return 'vendor-chart'
              }

              // 大型或低频第三方库单独拆分。
              // pdf.js 的 worker 地址（?url）只是一个字符串，不并进库本体，
              // 否则聊天页引用这个地址时会在打开页面就下载整个 PDF 库。
              if (moduleId.includes('/pdfjs-dist/') && !moduleId.includes('?url')) return 'vendor-pdf'
              if (moduleId.includes('/xlsx/')) return 'vendor-xlsx'
              if (moduleId.includes('/@airwallex/components-sdk/')) return 'vendor-airwallex'
              if (moduleId.includes('/@stripe/stripe-js/')) return 'vendor-stripe'
              // marked 与 DOMPurify 分开：侧边栏只需要 DOMPurify，marked 在渲染公告等 Markdown 时才按需加载。
              if (moduleId.includes('/marked/')) return 'vendor-markdown'
              if (moduleId.includes('/dompurify/')) return 'vendor-sanitize'
              if (moduleId.includes('/qrcode/')) return 'vendor-qrcode'
              if (moduleId.includes('/driver.js/')) return 'vendor-tour'
              if (moduleId.includes('/@tanstack/vue-virtual/')) return 'vendor-virtual'
              if (moduleId.includes('/@vueuse/')) return 'vendor-vueuse'
              if (moduleId.includes('/axios/')) return 'vendor-http'
              if (moduleId.includes('/emoji-picker-element/')) return 'vendor-emoji'

              // 其余第三方库不再合并成一个公共包，交给 Rollup 按实际引用关系分包。
              // 合并时 three、mammoth、highlight.js 会和图表、表格用到的小依赖打在一起，
              // 导致所有带图表或表格的页面都要下载这些用不到的大库。
            }

            // 应用代码：按入口点自动分包，不手动干预。
          }
        }
      }
    },
    server: {
      host: '0.0.0.0',
      port: devPort,
      proxy: {
        '/api': {
          target: backendUrl,
          changeOrigin: true
        },
        '/v1': {
          target: backendUrl,
          changeOrigin: true
        },
        '/setup': {
          target: backendUrl,
          changeOrigin: true
        }
      }
    }
  }
})
