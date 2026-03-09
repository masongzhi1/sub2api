import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import KeysView from '../KeysView.vue'

const {
  list,
  create,
  update,
  deleteKey,
  toggleStatus,
  getAvailable,
  getUserGroupRates,
  getPublicSettings,
  getDashboardApiKeysUsage,
  showError,
  showWarning,
  showSuccess,
  isCurrentStep,
  nextStep,
  copyToClipboard,
} = vi.hoisted(() => ({
  list: vi.fn(),
  create: vi.fn(),
  update: vi.fn(),
  deleteKey: vi.fn(),
  toggleStatus: vi.fn(),
  getAvailable: vi.fn(),
  getUserGroupRates: vi.fn(),
  getPublicSettings: vi.fn(),
  getDashboardApiKeysUsage: vi.fn(),
  showError: vi.fn(),
  showWarning: vi.fn(),
  showSuccess: vi.fn(),
  isCurrentStep: vi.fn(),
  nextStep: vi.fn(),
  copyToClipboard: vi.fn(),
}))

let isAPIKeyLogin = true
let isManagedTokenUser = false

const messages: Record<string, string> = {
  'keys.createDisabledForApiKeyLogin': '当前使用 API Key 登录，暂不支持创建新密钥。请改用账号密码登录后再创建。',
  'keys.createDisabledForManagedTokenUser': '令牌管理生成的账号不支持在前端创建新密钥。',
  'keys.createUnavailableEmptyState': '当前登录方式仅支持查看现有密钥，暂不支持创建新密钥。',
  'keys.createUnavailableForManagedTokenUser': '令牌管理生成的账号不提供此入口。',
  'keys.createKey': '创建密钥',
}

vi.mock('@/api', () => ({
  keysAPI: {
    list,
    create,
    update,
    delete: deleteKey,
    toggleStatus,
  },
  authAPI: {
    getPublicSettings,
  },
  usageAPI: {
    getDashboardApiKeysUsage,
  },
  userGroupsAPI: {
    getAvailable,
    getUserGroupRates,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showWarning,
    showSuccess,
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    isAPIKeyLogin,
    isManagedTokenUser,
  }),
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({
    isCurrentStep,
    nextStep,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const TablePageLayoutStub = {
  template: '<div><slot name="filters" /><slot name="actions" /><slot name="table" /><slot name="pagination" /></div>',
}
const DataTableStub = {
  props: ['data', 'loading'],
  template: '<div><slot v-if="!loading && Array.isArray(data) && data.length === 0" name="empty" /></div>',
}
const EmptyStateStub = {
  props: ['title', 'description', 'actionText'],
  emits: ['action'],
  template: `
    <div class="empty-state">
      <div class="empty-title">{{ title }}</div>
      <div class="empty-description">{{ description }}</div>
      <button v-if="actionText" class="empty-action" @click="$emit('action')">{{ actionText }}</button>
    </div>
  `,
}

function mountView() {
  return mount(KeysView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        TablePageLayout: TablePageLayoutStub,
        DataTable: DataTableStub,
        EmptyState: EmptyStateStub,
        BaseDialog: true,
        ConfirmDialog: true,
        Pagination: true,
        Select: true,
        SearchInput: true,
        Icon: true,
        UseKeyModal: true,
        GroupBadge: true,
        GroupOptionItem: true,
        Teleport: true,
      },
    },
  })
}

describe('user KeysView 创建限制', () => {
  beforeEach(() => {
    isAPIKeyLogin = true
    isManagedTokenUser = false
    list.mockReset()
    create.mockReset()
    update.mockReset()
    deleteKey.mockReset()
    toggleStatus.mockReset()
    getAvailable.mockReset()
    getUserGroupRates.mockReset()
    getPublicSettings.mockReset()
    getDashboardApiKeysUsage.mockReset()
    showError.mockReset()
    showWarning.mockReset()
    showSuccess.mockReset()
    isCurrentStep.mockReset()
    nextStep.mockReset()
    copyToClipboard.mockReset()

    list.mockResolvedValue({ items: [], total: 0, pages: 0 })
    getAvailable.mockResolvedValue([])
    getUserGroupRates.mockResolvedValue({})
    getPublicSettings.mockResolvedValue({})
    getDashboardApiKeysUsage.mockResolvedValue({ stats: {} })
    copyToClipboard.mockResolvedValue(true)
    isCurrentStep.mockReturnValue(false)
  })

  it('API Key 登录时禁用创建按钮并隐藏空状态创建动作', async () => {
    const wrapper = mountView()
    await flushPromises()

    const createButton = wrapper.get('[data-tour="keys-create-btn"]')
    expect(createButton.attributes('disabled')).toBeDefined()
    expect(createButton.attributes('title')).toBe(messages['keys.createDisabledForApiKeyLogin'])
    expect(wrapper.find('.empty-action').exists()).toBe(false)
    expect(wrapper.text()).toContain(messages['keys.createUnavailableEmptyState'])
  })

  it('API Key 登录时即使直接触发提交也不会调用创建接口', async () => {
    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.formData.name = 'Recovered Key'
    setupState.formData.group_id = 1

    await setupState.handleSubmit()

    expect(showWarning).toHaveBeenCalledWith(messages['keys.createDisabledForApiKeyLogin'])
    expect(create).not.toHaveBeenCalled()
  })

  it('令牌管理账号时禁用创建按钮并隐藏空状态创建动作', async () => {
    isAPIKeyLogin = false
    isManagedTokenUser = true

    const wrapper = mountView()
    await flushPromises()

    const createButton = wrapper.get('[data-tour="keys-create-btn"]')
    expect(createButton.attributes('disabled')).toBeDefined()
    expect(createButton.attributes('title')).toBe(messages['keys.createDisabledForManagedTokenUser'])
    expect(wrapper.find('.empty-action').exists()).toBe(false)
    expect(wrapper.text()).toContain(messages['keys.createUnavailableForManagedTokenUser'])
  })
})
