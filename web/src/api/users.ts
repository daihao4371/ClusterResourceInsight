/**
 * 用户管理API
 */
import axios from 'axios'
import type {
  User,
  CreateUserRequest,
  UpdateUserRequest,
  ChangePasswordRequest,
  ResetPasswordRequest,
  PageResponse,
  ApiResponse
} from '../types/management'

// 使用绝对路径，直接指向/api/users
const API_BASE = '/api/users'

// 创建用户
export const createUser = (data: CreateUserRequest) => {
  return axios.post<ApiResponse<User>>(`${API_BASE}`, data)
}

// 获取用户列表
export const getUsers = (params: {
  page?: number
  page_size?: number
  keyword?: string
  status?: string
}) => {
  return axios.get<ApiResponse<PageResponse<User>>>(`${API_BASE}`, { params })
}

// 获取用户详情
export const getUserById = (id: number) => {
  return axios.get<ApiResponse<User>>(`${API_BASE}/${id}`)
}

// 更新用户
export const updateUser = (id: number, data: UpdateUserRequest) => {
  return axios.put<ApiResponse<User>>(`${API_BASE}/${id}`, data)
}

// 删除用户
export const deleteUser = (id: number) => {
  return axios.delete<ApiResponse<any>>(`${API_BASE}/${id}`)
}

// 修改密码
export const changePassword = (id: number, data: ChangePasswordRequest) => {
  return axios.put<ApiResponse<any>>(`${API_BASE}/${id}/password`, data)
}

// 重置密码
export const resetPassword = (id: number, data: ResetPasswordRequest) => {
  return axios.put<ApiResponse<any>>(`${API_BASE}/${id}/reset-password`, data)
}