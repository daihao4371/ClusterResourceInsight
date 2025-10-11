/**
 * 审计日志API
 */
import axios from 'axios'
import type {
  AuditLog,
  AuditLogQueryParams,
  AuditStatistics,
  PageResponse,
  ApiResponse
} from '../types/management'

const API_BASE = '/api'

// 查询审计日志
export const getAuditLogs = (params: AuditLogQueryParams) => {
  return axios.get<ApiResponse<PageResponse<AuditLog>>>(`${API_BASE}/audit/logs`, { params })
}

// 获取审计日志详情
export const getAuditLogById = (id: number) => {
  return axios.get<ApiResponse<AuditLog>>(`${API_BASE}/audit/logs/${id}`)
}

// 获取指定用户的审计日志
export const getUserAuditLogs = (userId: number, params: { page?: number; page_size?: number }) => {
  return axios.get<ApiResponse<PageResponse<AuditLog>>>(
    `${API_BASE}/audit/users/${userId}/logs`,
    { params }
  )
}

// 获取审计统计信息
export const getAuditStatistics = (params: { start_time?: string; end_time?: string }) => {
  return axios.get<ApiResponse<AuditStatistics>>(`${API_BASE}/audit/statistics`, { params })
}

// 获取操作趋势
export const getOperationTrend = (days = 7) => {
  return axios.get<ApiResponse<{ trend: any[] }>>(`${API_BASE}/audit/trend`, {
    params: { days }
  })
}

// 清理旧审计日志
export const cleanupOldLogs = (days = 90) => {
  return axios.delete<ApiResponse<{ message: string; deleted_count: number }>>(
    `${API_BASE}/audit/cleanup`,
    { params: { days } }
  )
}

// 导出审计日志
export const exportAuditLogs = (params: Omit<AuditLogQueryParams, 'page' | 'page_size'>) => {
  return axios.get(`${API_BASE}/audit/export`, {
    params,
    responseType: 'blob'
  })
}