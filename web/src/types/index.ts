// 集群相关类型定义
export interface Cluster {
  id: number
  cluster_name: string
  cluster_alias?: string
  api_server: string
  auth_type: string
  token?: string
  status: string
  description?: string
  collect_interval: number
  created_at: string
  updated_at: string
  last_heartbeat?: string
  testing?: boolean  // 用于UI状态控制
}

export interface CreateClusterRequest {
  cluster_name: string
  cluster_alias?: string
  api_server: string
  auth_type: string
  auth_config: AuthConfigData
  tags?: string[]
  collect_interval?: number
  id?: number  // 用于编辑时
}

export interface AuthConfigData {
  bearer_token?: string
  cert_data?: string
  key_data?: string
  ca_data?: string
  kubeconfig?: string
}

export interface UpdateClusterRequest {
  cluster_name?: string
  cluster_alias?: string
  api_server?: string
  auth_type?: string
  auth_config?: AuthConfigData
  description?: string
  collect_interval?: number
  status?: string
}

// Pod资源信息
export interface PodResourceInfo {
  pod_name: string
  namespace: string
  node_name: string
  cluster_name: string
  memory_usage: number
  memory_request: number
  memory_limit: number
  memory_req_pct: number
  memory_limit_pct: number
  cpu_usage: number
  cpu_request: number
  cpu_limit: number
  cpu_req_pct: number
  cpu_limit_pct: number
  status: string
  issues: string[]
  creation_time: string
}

// 分析结果
export interface AnalysisResult {
  total_pods: number
  unreasonable_pods: number
  top50_problems: PodResourceInfo[]
  generated_at: string
  clusters_analyzed: number
}

// 命名空间汇总
export interface NamespaceSummary {
  namespace_name: string
  cluster_name: string
  total_pods: number
  unreasonable_pods: number
  total_memory_usage: number
  total_cpu_usage: number
  total_memory_request: number
  total_cpu_request: number
}

// Pod搜索请求
export interface PodSearchRequest {
  query?: string
  namespace?: string
  cluster?: string
  page?: number
  size?: number
}

// Pod搜索响应
export interface PodSearchResponse {
  pods: PodResourceInfo[]
  total: number
  page: number
  size: number
  total_pages: number
}

// 调度任务信息
export interface ScheduleJobInfo {
  cluster_id: number
  cluster_name: string
  interval: string
  last_run: string
  next_run: string
  status: string
  error_count: number
  last_error: string
  total_runs: number
  successful_runs: number
  restarting?: boolean  // 用于UI状态控制
}

// 调度服务状态
export interface ScheduleStatus {
  service_running: boolean
  total_jobs: number
  running_jobs: number
  error_jobs: number
  suspended_jobs: number
  stopped_jobs: number
  global_settings: GlobalScheduleSettings
}

// 全局调度设置
export interface GlobalScheduleSettings {
  enabled: boolean
  default_interval: string | number
  max_concurrent_jobs: number
  retry_max_attempts: number
  retry_interval: string | number
  enable_persistence: boolean
  health_check_interval: string | number
}

// 历史查询请求
export interface HistoryQueryRequest {
  cluster_id?: number
  namespace?: string
  pod_name?: string
  start_time?: string
  end_time?: string
  page?: number
  size?: number
  order_by?: string
  order_desc?: boolean
}

// 历史数据记录
export interface PodMetricsHistory {
  id: number
  cluster_id: number
  namespace: string
  pod_name: string
  node_name: string
  memory_usage: number
  memory_request: number
  memory_limit: number
  memory_req_pct: number
  memory_limit_pct: number
  cpu_usage: number
  cpu_request: number
  cpu_limit: number
  cpu_req_pct: number
  cpu_limit_pct: number
  status: string
  issues: string
  collected_at: string
}

// API响应包装类型
export interface ApiResponse<T = any> {
  data?: T
  message?: string
  error?: string
  count?: number
  success?: boolean
}