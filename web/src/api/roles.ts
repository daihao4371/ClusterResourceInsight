/**
 * 角色管理API
 */
import axios from 'axios'
import type {
  Role,
  CreateRoleRequest,
  UpdateRoleRequest,
  User,
  PageResponse,
  ApiResponse
} from '../types/management'

const API_BASE = '/api/roles'

// 创建角色
export const createRole = (data: CreateRoleRequest) => {
  return axios.post<ApiResponse<Role>>(`${API_BASE}`, data)
}

// 获取角色列表
export const getRoles = (params: {
  page?: number
  page_size?: number
  keyword?: string
  status?: string
}) => {
  return axios.get<ApiResponse<PageResponse<Role>>>(`${API_BASE}`, { params })
}

// 获取所有激活角色（用于选择器）
export const getActiveRoles = () => {
  return axios.get<ApiResponse<Role[]>>(`${API_BASE}/active`)
}

// 获取角色详情
export const getRoleById = (id: number) => {
  return axios.get<ApiResponse<Role>>(`${API_BASE}/${id}`)
}

// 更新角色
export const updateRole = (id: number, data: UpdateRoleRequest) => {
  return axios.put<ApiResponse<Role>>(`${API_BASE}/${id}`, data)
}

// 删除角色
export const deleteRole = (id: number) => {
  return axios.delete<ApiResponse<any>>(`${API_BASE}/${id}`)
}

// 获取角色下的用户列表
export const getRoleUsers = (id: number, params: { page?: number; page_size?: number }) => {
  return axios.get<ApiResponse<PageResponse<User>>>(`${API_BASE}/${id}/users`, { params })
}

// 获取角色权限
export const getRolePermissions = (id: number) => {
  return axios.get<ApiResponse<{ permissions: string[] }>>(`${API_BASE}/${id}/permissions`)
}