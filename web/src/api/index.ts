import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建axios实例
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 这里可以添加token等认证信息
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    console.error('API请求错误:', error)
    
    let message = '请求失败'
    if (error.response) {
      message = error.response.data?.error || `请求失败 (${error.response.status})`
    } else if (error.request) {
      message = '网络连接失败'
    }
    
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

// API接口定义
export const clusterApi = {
  // 获取所有集群
  getClusters: () => api.get('/clusters'),
  
  // 根据ID获取集群
  getCluster: (id: number) => api.get(`/clusters/${id}`),
  
  // 创建集群
  createCluster: (data: any) => api.post('/clusters', data),
  
  // 更新集群
  updateCluster: (id: number, data: any) => api.put(`/clusters/${id}`, data),
  
  // 删除集群
  deleteCluster: (id: number) => api.delete(`/clusters/${id}`),
  
  // 测试集群连接
  testCluster: (id: number) => api.post(`/clusters/${id}/test`),
  
  // 批量测试所有集群
  batchTestClusters: () => api.post('/clusters/batch-test'),
}

export const analysisApi = {
  // 获取资源分析结果
  getAnalysis: () => api.get('/analysis'),
  
  // 获取Pod数据
  getPods: (params?: any) => api.get('/pods', { params }),
}

export const statisticsApi = {
  // 获取内存使用Top Pod
  getTopMemoryPods: (limit = 50) => api.get(`/statistics/top-memory-request?limit=${limit}`),
  
  // 获取CPU使用Top Pod
  getTopCpuPods: (limit = 50) => api.get(`/statistics/top-cpu-request?limit=${limit}`),
  
  // 获取命名空间汇总
  getNamespacesSummary: () => api.get('/statistics/namespace-summary'),
}

export const namespaceApi = {
  // 获取所有命名空间
  getNamespaces: () => api.get('/namespaces'),
  
  // 获取命名空间下的Pod
  getNamespacePods: (namespace: string) => api.get(`/namespaces/${namespace}/pods`),
  
  // 获取命名空间树状数据
  getNamespaceTreeData: (namespace: string) => api.get(`/namespaces/${namespace}/tree-data`),
}

export const podApi = {
  // 搜索Pod
  searchPods: (params: any) => api.get('/pods/search', { params }),
  
  // 获取Pod列表
  listPods: (params: any) => api.get('/pods/list', { params }),
}

export const scheduleApi = {
  // 获取调度服务状态
  getStatus: () => api.get('/schedule/status'),
  
  // 启动调度服务
  start: () => api.post('/schedule/start'),
  
  // 停止调度服务
  stop: () => api.post('/schedule/stop'),
  
  // 获取所有调度任务
  getJobs: () => api.get('/schedule/jobs'),
  
  // 重启集群任务
  restartJob: (clusterId: number) => api.post(`/schedule/jobs/${clusterId}/restart`),
  
  // 更新调度设置
  updateSettings: (data: any) => api.put('/schedule/settings', data),
}

export const historyApi = {
  // 查询历史数据
  queryHistory: (params: any) => api.get('/history/query', { params }),
  
  // 获取趋势数据
  getTrends: (params: any) => api.get('/history/trends', { params }),
  
  // 获取历史统计
  getStatistics: () => api.get('/history/statistics'),
  
  // 触发数据收集
  collectData: (persistence = true) => api.post(`/history/collect?persistence=${persistence}`),
  
  // 清理过期数据
  cleanupData: (retentionDays = 30) => api.delete(`/history/cleanup?retention_days=${retentionDays}`),
}

export default api