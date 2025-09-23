import { ApiResponse, PaginationParams, PaginationResponse } from './types'

// Pod相关的数据类型定义 - 与后端PodResourceInfo对齐
export interface Pod {
  // 基础信息
  pod_name: string
  namespace: string
  node_name: string
  cluster_name: string
  
  // 内存信息 
  memory_usage: number      // 实际使用量 (bytes)
  memory_request: number    // 请求量 (bytes)
  memory_limit: number      // 限制量 (bytes)
  memory_req_pct: number    // 使用量/请求量 百分比
  memory_limit_pct: number  // 使用量/限制量 百分比
  
  // CPU信息
  cpu_usage: number         // 实际使用量 (millicores)
  cpu_request: number       // 请求量 (millicores)
  cpu_limit: number         // 限制量 (millicores)
  cpu_req_pct: number       // 使用量/请求量 百分比
  cpu_limit_pct: number     // 使用量/限制量 百分比
  
  // 状态信息
  status: string            // 合理/不合理
  issues: string[]          // 问题描述
  creation_time: string     // 创建时间
  
  // 前端显示用的计算属性（可选）
  id?: string               // 前端生成的唯一ID
  name?: string             // pod_name的别名
  cluster?: string          // cluster_name的别名
  cpuUsage?: number         // cpu_req_pct的别名（百分比）
  memoryUsage?: number      // memory_req_pct的别名（百分比）
  restarts?: number         // 重启次数
  startTime?: string        // creation_time的别名
}

export interface Container {
  name: string
  image: string
  ready: boolean
  restartCount: number
  status: string
}

// Pod搜索请求参数 - 与后端PodSearchRequest对齐
export interface PodSearchRequest extends PaginationParams {
  query?: string        // 搜索关键词（Pod名称）
  namespace?: string    // 命名空间筛选
  cluster?: string      // 集群筛选
  status?: string       // 状态筛选（合理/不合理）
  sortBy?: string       // 排序字段（前端使用，后端可能不支持）
}

// Pod统计数据
export interface PodStats {
  running: number
  avgCpuUsage: number
  // 后端兼容字段
  total_pods?: number          // 后端返回的字段名
  unreasonable_pods?: number   // 后端返回的字段名
  avg_cpu_usage?: number       // 后端返回的字段名
  avgMemoryUsage?: number
}

// 筛选选项接口
export interface FilterOptions {
  namespaces: string[]
  clusters: string[]
  statuses: string[]
}

// Pod详细分析数据类型
export interface PodDetailAnalysis {
  pod_info: Pod
  cluster_info: string
  node_info: string
  resource_analysis: {
    memory_analysis: {
      config_status: string
      efficiency_score: number
      waste_amount: number
      recommendations: string[]
    }
    cpu_analysis: {
      config_status: string
      efficiency_score: number
      waste_amount: number
      recommendations: string[]
    }
  }
  comparison_analysis: {
    namespace_average: {
      memory_usage_pct: number
      cpu_usage_pct: number
    }
    cluster_average: {
      memory_usage_pct: number
      cpu_usage_pct: number
    }
    similar_pods: Pod[]
  }
  alerts_info: {
    active_alerts: string[]
    history_alerts: string[]
    severity_level: string
    alert_count: number
  }
  generated_at: string
}

// Pod趋势数据类型
export interface PodTrendData {
  pod_info: Pod
  time_range: {
    start_time: string
    end_time: string
    duration: string
  }
  cpu_trend: {
    data_points: Array<{
      timestamp: string
      usage: number
      request: number
      limit: number
    }>
    statistics: {
      average: number
      peak: number
      minimum: number
      variance: number
    }
  }
  memory_trend: {
    data_points: Array<{
      timestamp: string
      usage: number
      request: number
      limit: number
    }>
    statistics: {
      average: number
      peak: number
      minimum: number
      variance: number
    }
  }
  event_markers: Array<{
    timestamp: string
    event_type: string
    description: string
    severity: string
  }>
  generated_at: string
}

// Pod列表响应
export interface PodsListResponse extends PaginationResponse<Pod> {
  stats?: PodStats
}

// API基础URL
let process;
const API_BASE_URL = typeof process !== 'undefined' && process.env.NODE_ENV === 'production'
  ? '/api' 
  : 'http://localhost:9999/api'

// HTTP客户端封装
class HttpClient {
  private baseURL: string

  constructor(baseURL: string) {
    this.baseURL = baseURL
  }

  async get<T>(url: string, params?: Record<string, any>): Promise<T> {
    const searchParams = new URLSearchParams()
    if (params) {
      Object.keys(params).forEach(key => {
        if (params[key] !== undefined && params[key] !== null && params[key] !== '') {
          searchParams.append(key, String(params[key]))
        }
      })
    }

    const fullUrl = `${this.baseURL}${url}${searchParams.toString() ? '?' + searchParams.toString() : ''}`
    
    try {
      const response = await fetch(fullUrl, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error('API请求失败:', error)
      throw error
    }
  }

  async post<T>(url: string, data?: any): Promise<T> {
    try {
      const response = await fetch(`${this.baseURL}${url}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: data ? JSON.stringify(data) : undefined,
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const result = await response.json()
      return result
    } catch (error) {
      console.error('API请求失败:', error)
      throw error
    }
  }
}

// 创建HTTP客户端实例
const httpClient = new HttpClient(API_BASE_URL)

// Pod相关的API服务
export class PodsApiService {
  // 获取Pod列表（支持分页和筛选）
  static async getPodsWithSearch(params: PodSearchRequest): Promise<ApiResponse<PodsListResponse>> {
    return httpClient.get<ApiResponse<PodsListResponse>>('/pods/search', params)
  }

  // 获取Pod列表（简单分页）
  static async getPodsList(page: number = 1, size: number = 20): Promise<ApiResponse<PodsListResponse>> {
    return httpClient.get<ApiResponse<PodsListResponse>>('/pods/list', { page, size })
  }

  // 获取问题Pod列表
  static async getProblemsWithPagination(params: {
    page?: number
    size?: number
    cluster_name?: string
    sort_by?: string
  }): Promise<ApiResponse<PodsListResponse>> {
    return httpClient.get<ApiResponse<PodsListResponse>>('/pods/problems', params)
  }

  // 获取Pod统计数据
  static async getPodStats(): Promise<ApiResponse<PodStats>> {
    return httpClient.get<ApiResponse<PodStats>>('/stats')
  }

  // 获取基础Pod数据（兼容原有接口）
  static async getPodsData(limit: number = 50, onlyProblems: boolean = true): Promise<ApiResponse<any>> {
    return httpClient.get<ApiResponse<any>>('/pods', {
      limit,
      only_problems: onlyProblems.toString()
    })
  }

  // 触发数据收集
  static async triggerDataCollection(): Promise<ApiResponse<any>> {
    return httpClient.post<ApiResponse<any>>('/history/collect')
  }

  // 获取筛选选项 - 支持按集群筛选命名空间
  static async getFilterOptions(cluster?: string): Promise<ApiResponse<FilterOptions>> {
    const params = cluster ? { cluster } : {}
    return httpClient.get<ApiResponse<FilterOptions>>('/pods/filter-options', params)
  }

  // 获取Pod详细分析
  static async getPodDetailAnalysis(
    cluster: string, 
    namespace: string, 
    podName: string
  ): Promise<ApiResponse<PodDetailAnalysis>> {
    return httpClient.get<ApiResponse<PodDetailAnalysis>>(
      `/pods/${cluster}/${namespace}/${podName}/detail`
    )
  }

  // 获取Pod趋势数据
  static async getPodTrendData(
    cluster: string, 
    namespace: string, 
    podName: string, 
    hours: number = 24
  ): Promise<ApiResponse<PodTrendData>> {
    return httpClient.get<ApiResponse<PodTrendData>>(
      `/pods/${cluster}/${namespace}/${podName}/trend`,
      { hours: hours.toString() }
    )
  }
}

// 导出默认实例
export default PodsApiService