import axios, { type AxiosResponse } from 'axios'
import type { ApiResponse } from '../types'

// 创建 axios 实例
export const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 可以在这里添加认证 token 等
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    // 统一处理响应数据
    if (response.data.code !== 0) {
      throw new Error(response.data.msg || '请求失败')
    }
    return response
  },
  (error) => {
    // 统一处理错误
    let message = '网络错误'
    
    if (error.response) {
      const { status, data } = error.response
      switch (status) {
        case 400:
          message = data?.msg || '请求参数错误'
          break
        case 401:
          message = '未授权访问'
          break
        case 403:
          message = '访问被禁止'
          break
        case 404:
          message = '请求的资源不存在'
          break
        case 500:
          message = data?.msg || '服务器内部错误'
          break
        default:
          message = data?.msg || `请求失败 (${status})`
      }
    } else if (error.request) {
      message = '网络连接失败'
    } else {
      message = error.message || '请求失败'
    }
    
    return Promise.reject(new Error(message))
  }
)

export default api