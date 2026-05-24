import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import JSZip from 'jszip'
import ImportDataModal from '@/components/admin/account/ImportDataModal.vue'

const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      importData: vi.fn(),
      importCodexSession: vi.fn()
    }
  }
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      if (key === 'admin.accounts.dataImportInvalidPayload') {
        return `invalid:${String(params?.file ?? '')}`
      }
      if (key === 'admin.accounts.dataImportZipNoJsonFiles') {
        return 'zip-no-json'
      }
      if (key === 'admin.accounts.dataImportZipJsonParseFailed') {
        return `zip-parse-failed:${String(params?.file ?? '')}`
      }
      if (key === 'admin.accounts.dataImportZipMixedFormats') {
        return 'zip-mixed-formats'
      }
      if (key === 'admin.accounts.dataImportCodexResultSummary') {
        return `codex-summary:${String(params?.created ?? '')}/${String(params?.updated ?? '')}/${String(params?.skipped ?? '')}/${String(params?.failed ?? '')}`
      }
      if (key === 'admin.accounts.oauth.openai.codexSessionImportSuccess') {
        return `codex-success:${String(params?.created ?? '')}/${String(params?.updated ?? '')}/${String(params?.skipped ?? '')}/${String(params?.failed ?? '')}`
      }
      if (key === 'admin.accounts.oauth.openai.codexSessionImportPartial') {
        return `codex-partial:${String(params?.created ?? '')}/${String(params?.updated ?? '')}/${String(params?.skipped ?? '')}/${String(params?.failed ?? '')}`
      }
      if (key === 'admin.accounts.oauth.openai.codexSessionImportFailed') {
        return 'codex-failed'
      }
      return key
    }
  })
}))

describe('ImportDataModal', () => {
  beforeEach(() => {
    showError.mockReset()
    showSuccess.mockReset()
    vi.resetAllMocks()
  })

  it('未选择文件时提示错误', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    await wrapper.find('form').trigger('submit')
    expect(showError).toHaveBeenCalledWith('admin.accounts.dataImportSelectFile')
  })

  it('无效 JSON 时提示解析失败', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const input = wrapper.find('input[type="file"]')
    const file = new File(['invalid json'], 'data.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: () => Promise.resolve('invalid json')
    })
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('admin.accounts.dataImportParseFailed')
  })

  it('支持导入包含多个 JSON 文件的 ZIP 压缩包', async () => {
    const { adminAPI } = await import('@/api/admin')
    vi.mocked(adminAPI.accounts.importData).mockResolvedValue({
      proxy_created: 1,
      proxy_reused: 0,
      proxy_failed: 0,
      account_created: 2,
      account_failed: 0,
      errors: []
    })

    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const zip = new JSZip()
    zip.file('a.json', JSON.stringify({
      type: 'sub2api-data',
      version: 1,
      exported_at: '2026-05-13T00:00:00.000Z',
      proxies: [{ proxy_key: 'p1', name: 'proxy-1', protocol: 'http', host: '1.1.1.1', port: 80, status: 'active' }],
      accounts: [{ name: 'acc-1', platform: 'openai', type: 'oauth', credentials: { token: 'x' }, concurrency: 1, priority: 1 }]
    }))
    zip.file('nested/b.json', JSON.stringify({
      type: 'sub2api-data',
      version: 1,
      exported_at: '2026-05-13T00:01:00.000Z',
      proxies: [],
      accounts: [{ name: 'acc-2', platform: 'openai', type: 'oauth', credentials: { token: 'y' }, concurrency: 2, priority: 2 }]
    }))

    const zipBuffer = await zip.generateAsync({ type: 'uint8array' })
    const file = new File([zipBuffer], 'bundle.zip', { type: 'application/zip' })
    Object.defineProperty(file, 'arrayBuffer', {
      value: () => Promise.resolve(zipBuffer.buffer.slice(zipBuffer.byteOffset, zipBuffer.byteOffset + zipBuffer.byteLength))
    })

    const input = wrapper.find('input[type="file"]')
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => {
      expect(adminAPI.accounts.importData).toHaveBeenCalled()
    })

    expect(showError).not.toHaveBeenCalled()
    expect(adminAPI.accounts.importData).toHaveBeenCalledWith({
      data: {
        type: 'sub2api-data',
        version: 1,
        exported_at: '2026-05-13T00:00:00.000Z',
        proxies: [{ proxy_key: 'p1', name: 'proxy-1', protocol: 'http', host: '1.1.1.1', port: 80, status: 'active' }],
        accounts: [
          { name: 'acc-1', platform: 'openai', type: 'oauth', credentials: { token: 'x' }, concurrency: 1, priority: 1 },
          { name: 'acc-2', platform: 'openai', type: 'oauth', credentials: { token: 'y' }, concurrency: 2, priority: 2 }
        ]
      },
      skip_default_group_bind: true
    })
  })

  it('支持导入包含多个账号 JSON 文件的 ZIP 压缩包', async () => {
    const { adminAPI } = await import('@/api/admin')
    vi.mocked(adminAPI.accounts.importCodexSession).mockResolvedValue({
      total: 2,
      created: 2,
      updated: 0,
      skipped: 0,
      failed: 0,
      items: []
    })

    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const zip = new JSZip()
    zip.file('account-a.json', JSON.stringify({
      accessToken: 'header.payload.signature',
      email: 'a@example.com'
    }))
    zip.file('nested/account-b.json', JSON.stringify({
      tokens: {
        access_token: 'header.payload.signature.2',
        refresh_token: 'refresh-token-2'
      },
      email: 'b@example.com'
    }))

    const zipBuffer = await zip.generateAsync({ type: 'uint8array' })
    const file = new File([zipBuffer], 'accounts.zip', { type: 'application/zip' })
    Object.defineProperty(file, 'arrayBuffer', {
      value: () => Promise.resolve(zipBuffer.buffer.slice(zipBuffer.byteOffset, zipBuffer.byteOffset + zipBuffer.byteLength))
    })

    const input = wrapper.find('input[type="file"]')
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => {
      expect(adminAPI.accounts.importCodexSession).toHaveBeenCalled()
    })

    expect(adminAPI.accounts.importData).not.toHaveBeenCalled()
    expect(adminAPI.accounts.importCodexSession).toHaveBeenCalledWith({
      contents: [
        JSON.stringify({
          accessToken: 'header.payload.signature',
          email: 'a@example.com'
        }),
        JSON.stringify({
          tokens: {
            access_token: 'header.payload.signature.2',
            refresh_token: 'refresh-token-2'
          },
          email: 'b@example.com'
        })
      ],
      skip_default_group_bind: true,
      update_existing: true
    })
    expect(showSuccess).toHaveBeenCalledWith('codex-success:2/0/0/0')
  })

  it('ZIP 内没有 JSON 文件时提示错误', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const zip = new JSZip()
    zip.file('readme.txt', 'hello')
    const zipBuffer = await zip.generateAsync({ type: 'uint8array' })
    const file = new File([zipBuffer], 'bundle.zip', { type: 'application/zip' })
    Object.defineProperty(file, 'arrayBuffer', {
      value: () => Promise.resolve(zipBuffer.buffer.slice(zipBuffer.byteOffset, zipBuffer.byteOffset + zipBuffer.byteLength))
    })

    const input = wrapper.find('input[type="file"]')
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('zip-no-json')
  })
})
