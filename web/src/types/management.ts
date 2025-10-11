/**
 * 管理模块类型定义
 */

// ==================== 用户管理 ====================
export interface User {
  id: number
  username: string
  email: string
  real_name: string
  phone: string
  status: 'active' | 'inactive' | 'locked'
  last_login_at?: string
  last_login_ip?: string
  created_at: string
  updated_at: string
  roles?: Role[]
}

export interface CreateUserRequest {
  username: string
  password: string
  email: string
  real_name?: string
  phone?: string
  role_ids?: number[]
}

export interface UpdateUserRequest {
  email?: string
  real_name?: string
  phone?: string
  status?: 'active' | 'inactive' | 'locked'
  role_ids?: number[]
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface ResetPasswordRequest {
  new_password: string
}

// ==================== 角色管理 ====================
export interface Role {
  id: number
  role_name: string
  role_code: string
  description: string
  permissions: string // JSON字符串
  is_system: boolean
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
  users?: User[]
}

export interface CreateRoleRequest {
  role_name: string
  role_code: string
  description?: string
  permissions: string[]
}

export interface UpdateRoleRequest {
  role_name?: string
  description?: string
  permissions?: string[]
  status?: 'active' | 'inactive'
}

// ==================== API管理 ====================
export interface ApiEndpoint {
  id: number
  path: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  group: string
  description: string
  permission: string
  is_public: boolean
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
}

export interface CreateApiRequest {
  path: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  group?: string
  description?: string
  permission?: string
  is_public?: boolean
}

export interface UpdateApiRequest {
  path?: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  group?: string
  description?: string
  permission?: string
  is_public?: boolean
  status?: 'active' | 'inactive'
}

export interface ApiStatistics {
  total: number
  by_status: Array<{status: string, count: number}>
  by_group: Array<{group: string, count: number}>
  by_method: Array<{method: string, count: number}>
  public_count: number
  private_count: number
}

// ==================== 审计日志 ====================
export interface AuditLog {
  id: number
  user_id: number
  username: string
  operate_type: 'create' | 'update' | 'delete' | 'query' | 'login' | 'logout'
  module: string
  resource: string
  resource_id: string
  method: string
  path: string
  request_body: string
  response_code: number
  ip: string
  user_agent: string
  error_msg: string
  duration: number
  created_at: string
  user?: User
}

export interface AuditLogQueryParams {
  page?: number
  page_size?: number
  user_id?: number
  username?: string
  operate_type?: string
  module?: string
  resource?: string
  start_time?: string
  end_time?: string
  min_duration?: number
  max_duration?: number
  response_code?: number
}

export interface AuditStatistics {
  total: number
  by_operate_type: Array<{operate_type: string, count: number}>
  by_module: Array<{module: string, count: number}>
  by_user: Array<{username: string, count: number}>
  by_status: Array<{response_code: number, count: number}>
  avg_duration: number
  slow_request_count: number
  error_count: number
}

// ==================== 分页响应 ====================
export interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// ==================== API响应 ====================
export interface ApiResponse<T = any> {
  code: number
  data: T
  msg: string
}