/**
 * API管理API
 */
import axios from 'axios'
import type {
  ApiEndpoint,
  CreateApiRequest,
  UpdateApiRequest,
  ApiStatistics,
  PageResponse,
  ApiResponse
} from '../types/management'

// 使用绝对路径，直接指向/api/apis
const API_BASE = '/api/apis'

// 创建API
export const createApi = (data: CreateApiRequest) => {
  return axios.post<ApiResponse<ApiEndpoint>>(API_BASE, data)
}

// 获取API列表
export const getApis = (params: {
  page?: number
  page_size?: number
  keyword?: string
  group?: string
  method?: string
  status?: string
  is_public?: boolean
}) => {
  return axios.get<ApiResponse<PageResponse<ApiEndpoint>>>(`${API_BASE}`, { params })
}

// 获取API分组列表
export const getApiGroups = () => {
  return axios.get<ApiResponse<{ groups: string[] }>>(`${API_BASE}/groups`)
}

// 获取API统计信息
export const getApiStatistics = () => {
  return axios.get<ApiResponse<ApiStatistics>>(`${API_BASE}/statistics`)
}

// 获取API详情
export const getApiById = (id: number) => {
  return axios.get<ApiResponse<ApiEndpoint>>(`${API_BASE}/${id}`)
}

// 更新API
export const updateApi = (id: number, data: UpdateApiRequest) => {
  return axios.put<ApiResponse<ApiEndpoint>>(`${API_BASE}/${id}`, data)
}

// 删除API
export const deleteApi = (id: number) => {
  return axios.delete<ApiResponse<any>>(`${API_BASE}/${id}`)
}

// 批量创建API
export const batchCreateApis = (apis: CreateApiRequest[]) => {
  return axios.post<ApiResponse<any>>(`${API_BASE}/batch`, { apis })
}

// 同步API
export const syncApis = (apis: CreateApiRequest[]) => {
  return axios.post<ApiResponse<any>>(`${API_BASE}/sync`, { apis })
}