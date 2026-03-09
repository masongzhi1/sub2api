<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="space-y-4">
          <TokenManagementTabs />

          <div class="flex flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-80">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
              />
              <input
                v-model="search"
                type="text"
                class="input pl-10 pr-10"
                :placeholder="t('admin.tokens.searchPlaceholder')"
                @keyup.enter="applySearch"
              />
              <button
                v-if="search"
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 transition-colors hover:text-gray-600 dark:hover:text-gray-200"
                :title="t('common.clear')"
                @click="clearSearch"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
          </div>
        </div>
      </template>

      <template #actions>
        <div class="flex justify-end gap-3">
          <button
            @click="loadTokens"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button @click="openCreateModal" class="btn btn-primary">
            <Icon name="plus" size="md" class="mr-2" />
            {{ t('admin.tokens.createToken') }}
          </button>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="tokens" :loading="loading">
          <template #cell-label="{ row }">
            <div class="min-w-[180px]">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.label }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">#{{ row.user.id }}</div>
            </div>
          </template>

          <template #cell-user="{ row }">
            <div class="min-w-[220px]">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.user.email }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ row.user.username || '-' }}</div>
            </div>
          </template>

          <template #cell-key="{ row }">
            <div class="flex min-w-[220px] items-center gap-2">
              <code class="rounded-lg bg-gray-100 px-2.5 py-1.5 text-xs text-gray-800 dark:bg-dark-700 dark:text-dark-100">
                {{ maskKey(row.api_key.key) }}
              </code>
              <button
                type="button"
                class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
                :title="t('common.copy')"
                @click="copyKey(row.api_key.key, row.api_key.id)"
              >
                <Icon :name="copiedKeyId === row.api_key.id ? 'check' : 'copy'" size="sm" />
              </button>
            </div>
          </template>

          <template #cell-group="{ row }">
            <GroupBadge
              v-if="resolveGroup(row)"
              :name="resolveGroup(row)!.name"
              :platform="resolveGroup(row)!.platform"
              :subscription-type="resolveGroup(row)!.subscription_type"
              :rate-multiplier="resolveGroup(row)!.rate_multiplier"
              :show-rate="false"
            />
            <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
          </template>

          <template #cell-expires_at="{ row }">
            <span class="text-sm text-gray-700 dark:text-dark-200">
              {{ formatDateTime(resolveExpiresAt(row)) || '-' }}
            </span>
          </template>

          <template #cell-status="{ row }">
            <span
              class="inline-flex rounded-full px-2.5 py-1 text-xs font-medium"
              :class="statusClass(resolveStatus(row))"
            >
              {{ statusLabel(resolveStatus(row)) }}
            </span>
          </template>

          <template #cell-created_at="{ row }">
            <span class="text-sm text-gray-700 dark:text-dark-200">
              {{ formatDateTime(row.api_key.created_at) || '-' }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end">
              <button
                type="button"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                :title="t('common.delete')"
                @click="handleDelete(row)"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.tokens.noTokensYet')"
              :description="t('admin.tokens.createFirstToken')"
              :action-text="t('admin.tokens.createToken')"
              @action="openCreateModal"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.tokens.deleteToken')"
      :message="t('admin.tokens.deleteConfirm', { name: deletingToken?.label || '' })"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <BaseDialog
      :show="showCreateModal"
      :title="t('admin.tokens.createToken')"
      width="normal"
      @close="closeCreateModal"
    >
      <form id="token-management-form" class="space-y-5" @submit.prevent="handleCreate">
        <div>
          <label class="input-label">{{ t('admin.tokens.form.nameLabel') }}</label>
          <input
            v-model="form.name"
            type="text"
            class="input"
            :placeholder="t('admin.tokens.form.namePlaceholder')"
            required
          />
        </div>

        <div>
          <label class="input-label">{{ t('admin.tokens.form.customKeyLabel') }}</label>
          <input
            v-model="form.custom_key"
            type="text"
            class="input font-mono"
            :placeholder="t('admin.tokens.form.customKeyPlaceholder')"
          />
          <p class="input-hint">{{ t('admin.tokens.form.customKeyHint') }}</p>
        </div>

        <div>
          <label class="input-label">{{ t('admin.tokens.form.groupLabel') }}</label>
          <Select
            v-model="form.group_id"
            :options="groupOptions"
            :placeholder="t('admin.tokens.form.groupPlaceholder')"
            :searchable="true"
          />
        </div>

        <div>
          <label class="input-label">{{ t('admin.tokens.form.validityDaysLabel') }}</label>
          <input
            v-model.number="form.validity_days"
            type="number"
            min="1"
            max="36500"
            class="input"
            :placeholder="t('admin.tokens.form.validityDaysPlaceholder')"
          />
          <p class="input-hint">{{ t('admin.tokens.form.validityDaysHint') }}</p>
        </div>

        <div>
          <label class="input-label">{{ t('admin.tokens.form.notesLabel') }}</label>
          <textarea
            v-model="form.notes"
            rows="3"
            class="input"
            :placeholder="t('admin.tokens.form.notesPlaceholder')"
          />
        </div>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeCreateModal">
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            form="token-management-form"
            class="btn btn-primary"
            :disabled="creating"
          >
            <Icon v-if="creating" name="refresh" size="sm" class="mr-2 animate-spin" />
            {{ creating ? t('common.saving') : t('admin.tokens.createToken') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { AdminGroup, ManagedToken } from '@/types'
import type { Column } from '@/components/common/types'
import { formatDateTime } from '@/utils/format'
import { useClipboard } from '@/composables/useClipboard'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import Icon from '@/components/icons/Icon.vue'
import TokenManagementTabs from '@/components/admin/token-management/TokenManagementTabs.vue'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard: clipboardCopy } = useClipboard()

const tokens = ref<ManagedToken[]>([])
const subscriptionGroups = ref<AdminGroup[]>([])
const loading = ref(false)
const creating = ref(false)
const showCreateModal = ref(false)
const showDeleteDialog = ref(false)
const search = ref('')
const copiedKeyId = ref<number | null>(null)
const deletingToken = ref<ManagedToken | null>(null)

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const form = reactive({
  name: '',
  custom_key: '',
  group_id: null as number | null,
  validity_days: 30,
  notes: ''
})

const columns = computed<Column[]>(() => [
  { key: 'label', label: t('admin.tokens.columns.label') },
  { key: 'user', label: t('admin.tokens.columns.user') },
  { key: 'key', label: t('admin.tokens.columns.key') },
  { key: 'group', label: t('admin.tokens.columns.group') },
  { key: 'expires_at', label: t('admin.tokens.columns.expires') },
  { key: 'status', label: t('admin.tokens.columns.status') },
  { key: 'created_at', label: t('admin.tokens.columns.created') },
  { key: 'actions', label: t('common.actions') }
])

const groupOptions = computed(() => {
  return subscriptionGroups.value.map((group) => ({
    value: group.id,
    label: `${group.name} · ${group.platform}`
  }))
})

function openCreateModal(): void {
  resetForm()
  showCreateModal.value = true
}

function closeCreateModal(): void {
  showCreateModal.value = false
  resetForm()
}

function resetForm(): void {
  form.name = ''
  form.custom_key = ''
  form.group_id = null
  form.validity_days = 30
  form.notes = ''
}

function maskKey(key: string): string {
  if (!key || key.length <= 12) {
    return key
  }
  return `${key.slice(0, 8)}••••${key.slice(-4)}`
}

function resolveGroup(row: ManagedToken) {
  return row.subscription?.group ?? row.api_key.group ?? null
}

function resolveExpiresAt(row: ManagedToken): string | null {
  return row.subscription?.expires_at ?? row.api_key.expires_at ?? null
}

function resolveStatus(row: ManagedToken): string {
  return row.subscription?.status ?? row.api_key.status
}

function statusLabel(status: string): string {
  if (status === 'active' || status === 'expired' || status === 'revoked') {
    return t(`admin.subscriptions.status.${status}`)
  }
  return status
}

function statusClass(status: string): string {
  if (status === 'active') {
    return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  }
  if (status === 'expired') {
    return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300'
  }
  if (status === 'revoked' || status === 'disabled' || status === 'inactive') {
    return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  }
  return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-200'
}

async function copyKey(key: string, keyId: number): Promise<void> {
  const success = await clipboardCopy(key, t('common.copiedToClipboard'))
  if (!success) {
    return
  }

  copiedKeyId.value = keyId
  window.setTimeout(() => {
    if (copiedKeyId.value === keyId) {
      copiedKeyId.value = null
    }
  }, 1800)
}

async function loadGroups(): Promise<void> {
  const groups = await adminAPI.groups.getAll()
  subscriptionGroups.value = groups.filter((group) => group.subscription_type === 'subscription')
}

async function loadTokens(): Promise<void> {
  loading.value = true
  try {
    const response = await adminAPI.tokenManagement.list(pagination.page, pagination.page_size, {
      search: search.value.trim() || undefined
    })
    tokens.value = response.items
    pagination.total = response.total
    pagination.page = response.page
    pagination.page_size = response.page_size
  } catch (error) {
    console.error(error)
    appStore.showError(t('admin.tokens.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function applySearch(): void {
  pagination.page = 1
  void loadTokens()
}

function clearSearch(): void {
  search.value = ''
  applySearch()
}

function handlePageChange(page: number): void {
  pagination.page = page
  void loadTokens()
}

function handlePageSizeChange(pageSize: number): void {
  pagination.page = 1
  pagination.page_size = pageSize
  void loadTokens()
}

function handleDelete(token: ManagedToken): void {
  deletingToken.value = token
  showDeleteDialog.value = true
}

async function confirmDelete(): Promise<void> {
  if (!deletingToken.value) {
    return
  }

  try {
    await adminAPI.tokenManagement.delete(deletingToken.value.user.id)
    appStore.showSuccess(t('admin.tokens.deletedSuccess'))
    showDeleteDialog.value = false
    deletingToken.value = null

    if (tokens.value.length === 1 && pagination.page > 1) {
      pagination.page -= 1
    }
    await loadTokens()
  } catch (error: any) {
    console.error(error)
    appStore.showError(error.response?.data?.detail || t('admin.tokens.failedToDelete'))
  }
}

async function handleCreate(): Promise<void> {
  if (!form.name.trim()) {
    appStore.showError(t('admin.tokens.nameRequired'))
    return
  }
  if (!form.group_id) {
    appStore.showError(t('admin.tokens.groupRequired'))
    return
  }
  if (!form.validity_days || form.validity_days < 1) {
    appStore.showError(t('admin.tokens.validityDaysRequired'))
    return
  }

  creating.value = true
  try {
    const created = await adminAPI.tokenManagement.create({
      name: form.name.trim(),
      group_id: form.group_id,
      validity_days: form.validity_days,
      custom_key: form.custom_key.trim() || undefined,
      notes: form.notes.trim() || undefined
    })

    closeCreateModal()
    pagination.page = 1
    await loadTokens()
    appStore.showSuccess(t('admin.tokens.createdSuccess'))

    if (created.api_key?.key) {
      await copyKey(created.api_key.key, created.api_key.id)
    }
  } catch (error) {
    console.error(error)
    appStore.showError(t('admin.tokens.failedToCreate'))
  } finally {
    creating.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadGroups(), loadTokens()])
})
</script>
