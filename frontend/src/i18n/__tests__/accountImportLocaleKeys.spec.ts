import { describe, expect, it } from 'vitest'

import enAdminAccounts from '../locales/en/admin/accounts'
import zhAdminAccounts from '../locales/zh/admin/accounts'

const importedKeys = [
  'dataImportInvalidPayload',
  'dataImportZipNoJsonFiles',
  'dataImportZipJsonParseFailed',
  'dataImportZipMixedFormats',
  'dataImportCodexResultSummary'
] as const

describe.each([
  ['en', enAdminAccounts],
  ['zh', zhAdminAccounts]
])('account import locale keys (%s)', (_locale, messages) => {
  it.each(importedKeys)('defines %s', (key) => {
    expect(messages.accounts[key]).toBeTruthy()
  })
})
