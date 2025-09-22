// 通用API响应类型
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
  success?: boolean
}

// 分页参数
export interface PaginationParams {
  page?: number
  size?: number
}

// 分页响应
export interface PaginationResponse<T> {
  data: T[]
  total: number
  page: number
  size: number
  totalPages?: number
  hasNext?: boolean
  hasPrev?: boolean
}

// 通用错误类型
export interface ApiError {
  code: number
  message: string
  details?: string
}