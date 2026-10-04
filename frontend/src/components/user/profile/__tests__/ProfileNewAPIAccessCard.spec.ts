import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const { generateTokenMock, copyMock, successMock, errorMock } = vi.hoisted(() => ({
  generateTokenMock: vi.fn(),
  copyMock: vi.fn(),
  successMock: vi.fn(),
  errorMock: vi.fn(),
}))

vi.mock('@/api/user', () => ({
  generateNewAPIBalanceAccessToken: generateTokenMock,
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: copyMock }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showSuccess: successMock, showError: errorMock }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

import ProfileNewAPIAccessCard from '@/components/user/profile/ProfileNewAPIAccessCard.vue'

const ConfirmDialogStub = defineComponent({
  props: { show: Boolean },
  emits: ['confirm', 'cancel'],
  template: '<button v-if="show" data-testid="confirm-token" @click="$emit(\'confirm\')">confirm</button>',
})

describe('ProfileNewAPIAccessCard', () => {
  beforeEach(() => {
    generateTokenMock.mockReset().mockResolvedValue('sub_bal_generated-token')
    copyMock.mockReset()
    successMock.mockReset()
    errorMock.mockReset()
  })

  it('shows the user ID and displays a newly generated token once', async () => {
    const wrapper = mount(ProfileNewAPIAccessCard, {
      props: { userId: 31 },
      global: { stubs: { ConfirmDialog: ConfirmDialogStub, Icon: true } },
    })

    expect(wrapper.get('[data-testid="profile-newapi-user-id"]').text()).toBe('31')
    expect(wrapper.find('[data-testid="profile-newapi-token"]').exists()).toBe(false)

    await wrapper.get('[data-testid="profile-newapi-generate"]').trigger('click')
    await wrapper.get('[data-testid="confirm-token"]').trigger('click')
    await flushPromises()

    expect(generateTokenMock).toHaveBeenCalledTimes(1)
    expect(wrapper.get<HTMLInputElement>('[data-testid="profile-newapi-token"]').element.value).toBe(
      'sub_bal_generated-token'
    )
    expect(successMock).toHaveBeenCalledTimes(1)
  })
})
