import { apiClient } from '../client'
import type { PaginatedResponse, ManagedToken, CreateManagedTokenRequest } from '@/types'

export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    search?: string
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<PaginatedResponse<ManagedToken>> {
  const { data } = await apiClient.get<PaginatedResponse<ManagedToken>>('/admin/token-management', {
    params: {
      page,
      page_size: pageSize,
      ...filters
    },
    signal: options?.signal
  })
  return data
}

export async function create(request: CreateManagedTokenRequest): Promise<ManagedToken> {
  const { data } = await apiClient.post<ManagedToken>('/admin/token-management', request)
  return data
}

export const tokenManagementAPI = {
  list,
  create
}

export default tokenManagementAPI
