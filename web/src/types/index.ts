// API 基础类型定义
export interface ApiResponse<T = any> {
  code: number
  data: T
  msg: string
}

// 集群相关类型
export interface Cluster {
  id: string  // 前端使用string类型的ID
  name: string
  alias?: string
  endpoint: string
  status: 'online' | 'offline' | 'error' | 'unknown'
  auth_type?: string
  collect_interval?: number
  created_at?: string
  updated_at?: string
  last_collect_at?: string
  // 前端UI展示需要的额外字段
  region?: string
  nodes?: number
  pods?: number
  cpuUsage?: number
  memoryUsage?: number
  tags?: string[]
  // 实时统计数据字段
  version?: string
  namespace_count?: number
  response_time_ms?: number
  has_metrics?: boolean
  last_test_time?: string
  // 详细资源信息字段
  cpuUsedCores?: number
  cpuTotalCores?: number
  memoryUsedGB?: number
  memoryTotalGB?: number
  hasRealUsage?: boolean
  dataSource?: string
}

// 后端原始集群数据结构
export interface BackendCluster {
  id: number
  cluster_name: string
  cluster_alias?: string
  api_server: string
  auth_type: string
  status: 'online' | 'offline' | 'error' | 'unknown'
  tags?: string[]
  collect_interval: number
  last_collect_at?: string
  created_at: string
  updated_at: string
}

export interface ClusterTestResult {
  success: boolean
  status: 'online' | 'offline' | 'error'
  message: string
  version?: string
  node_count: number
  namespace_count: number
  pod_count: number
  has_metrics: boolean
  test_time: string
  response_time_ms: number
  cpu_usage?: number
  memory_usage?: number
  cpu_used_cores?: number
  cpu_total_cores?: number
  memory_used_gb?: number
  memory_total_gb?: number
  has_real_usage?: boolean
  data_source?: string
}

// Pod 相关类型
export interface Pod {
  pod_name: string  // 修正字段名，与后端API保持一致
  namespace: string
  cluster_name: string
  node_name: string
  status: string
  cpu_request: number   // 使用数字类型存储millicores
  cpu_limit: number
  memory_request: number // 使用数字类型存储bytes
  memory_limit: number
  cpu_usage: number
  memory_usage: number
  cpu_req_pct?: number   // CPU请求利用率百分比
  cpu_limit_pct?: number // CPU限制利用率百分比
  memory_req_pct?: number // 内存请求利用率百分比
  memory_limit_pct?: number // 内存限制利用率百分比
  issues: string[]
  creation_time: string
}

export interface PodSearchRequest {
  page: number
  size: number
  name?: string
  namespace?: string
  cluster_name?: string
  status?: string
  node_name?: string
  only_problems?: boolean
}

export interface PodSearchResponse {
  pods: Pod[]
  total: number
  page: number
  size: number
  has_next: boolean
}

// 命名空间相关类型
export interface NamespaceSummary {
  namespace_name: string
  cluster_name: string
  pod_count: number
  running_pods: number
  pending_pods: number
  failed_pods: number
  total_cpu_request: string
  total_memory_request: string
  total_cpu_limit: string
  total_memory_limit: string
  resource_efficiency: number
}

// 资源分析类型
export interface ResourceAnalysis {
  total_pods: number
  unreasonable_pods: number
  top50_problems: Pod[]  // 修正字段名
  generated_at: string   // 修正字段名
  clusters_analyzed: number // 新增字段
}

// 历史数据类型
export interface HistoryQueryRequest {
  page: number
  size: number
  cluster_id?: number
  namespace?: string
  pod_name?: string
  start_time?: string
  end_time?: string
}

export interface PodMetricsHistory {
  id: number
  cluster_id: number
  cluster_name: string
  namespace: string
  pod_name: string
  node_name: string
  cpu_request: string
  memory_request: string
  cpu_limit: string
  memory_limit: string
  cpu_usage: string
  memory_usage: string
  status: string
  collected_at: string
}

export interface HistoryQueryResponse {
  data: PodMetricsHistory[]
  total: number
  page: number
  size: number
  has_next: boolean
}

// 调度服务类型
export interface ScheduleJobInfo {
  cluster_id: number
  cluster_name: string
  interval: number
  last_run: string
  next_run: string
  status: 'running' | 'stopped' | 'error' | 'suspended'
  error_count: number
  last_error: string
  total_runs: number
  successful_runs: number
}

export interface GlobalScheduleSettings {
  enabled: boolean
  default_interval: number
  max_concurrent_jobs: number
  retry_max_attempts: number
  retry_interval: number
  enable_persistence: boolean
  health_check_interval: number
}

// 通知系统类型
export interface Notification {
  id: string
  type: 'success' | 'warning' | 'error' | 'info'
  title: string
  message: string
  duration?: number
  timestamp: number
}

// 统计类型
export interface SystemStats {
  total_clusters: number
  online_clusters: number
  total_pods: number
  problem_pods: number
  last_update: string
}

// 扩展的图表数据类型
export interface ChartData {
  name: string
  value: number
  color: string
  percentage?: number
}

export interface TrendData {
  time: string
  cpu: number
  memory: number
  pods: number
  storage?: number
  network?: number
}

// 实时活动类型
export interface ActivityItem {
  type: 'success' | 'warning' | 'error' | 'info'
  message: string
  time: string
  details?: string
}

// 告警类型
export interface Alert {
  level: 'high' | 'medium' | 'low'
  title: string
  description: string
  time: string
  progress?: number
}

// 节点信息类型
export interface Node {
  id: string
  name: string
  cluster_id: string
  status: 'Ready' | 'NotReady' | 'Unknown'
  roles: string[]
  version: string
  cpu_capacity: string
  memory_capacity: string
  cpu_usage: number
  memory_usage: number
  pod_count: number
  conditions: NodeCondition[]
}

export interface NodeCondition {
  type: string
  status: string
  reason: string
  message: string
  last_transition: string
}

// 容器信息类型
export interface Container {
  name: string
  image: string
  status: string
  ready: boolean
  restart_count: number
  cpu_usage: number
  memory_usage: number
}

// 扩展Pod类型
export interface ExtendedPod extends Pod {
  containers: Container[]
  restart_count: number
  age: string
  labels: Record<string, string>
  annotations: Record<string, string>
}

// 分页类型
export interface Pagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

// 查询参数类型
export interface QueryParams {
  page?: number
  page_size?: number
  search?: string
  sort?: string
  order?: 'asc' | 'desc'
  filter?: Record<string, any>
}

// WebSocket消息类型
export interface WebSocketMessage {
  type: 'stats' | 'alert' | 'activity' | 'pod_update' | 'cluster_update'
  data: any
  timestamp: string
}

// 用户相关类型
export interface User {
  id: string
  username: string
  email: string
  role: 'admin' | 'user' | 'viewer'
  permissions: string[]
  created_at: string
  last_login: string
  status: 'active' | 'inactive'
}

// 系统配置类型
export interface SystemConfig {
  refresh_interval: number
  max_history_days: number
  alert_retention_days: number
  enable_websocket: boolean
  enable_auto_refresh: boolean
  theme: 'dark' | 'light' | 'auto'
  language: 'zh-CN' | 'en-US'
}

// 路由元信息类型
export interface RouteMeta {
  title: string
  icon?: string
  requiresAuth?: boolean
  roles?: string[]
  keepAlive?: boolean
}