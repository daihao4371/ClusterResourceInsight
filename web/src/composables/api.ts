import { ref, computed } from 'vue'
import { api } from '../utils/api'
import { useSystemStore } from '../stores/system'
import type { 
  ApiResponse, 
  Cluster, 
  Pod, 
  PodSearchRequest, 
  PodSearchResponse,
  HistoryQueryRequest,
  HistoryQueryResponse,
  ResourceAnalysis,
  NamespaceSummary
} from '../types'

/**
 * 集群相关数据获取
 */
export function useClusters() {
  const clusters = ref<Cluster[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const fetchClusters = async () => {
    try {
      loading.value = true
      error.value = null
      const response = await api.get<ApiResponse<any>>('/clusters')
      // 确保返回的数据是数组格式
      if (response.data && response.data.data && Array.isArray(response.data.data)) {
        clusters.value = response.data.data
      } else {
        clusters.value = []
        console.warn('集群数据格式异常，使用空数组')
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取集群数据失败'
      clusters.value = [] // 确保发生错误时也是数组
      console.error('获取集群数据失败:', err)
    } finally {
      loading.value = false
    }
  }

  const addCluster = async (clusterData: Partial<Cluster>) => {
    try {
      loading.value = true
      const response = await api.post<ApiResponse<Cluster>>('/clusters', clusterData)
      clusters.value.push(response.data.data)
      return response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '添加集群失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  const updateCluster = async (id: number, clusterData: Partial<Cluster>) => {
    try {
      loading.value = true
      const response = await api.put<ApiResponse<Cluster>>(`/clusters/${id}`, clusterData)
      const index = clusters.value.findIndex(c => Number(c.id) === id)
      if (index !== -1) {
        clusters.value[index] = response.data.data
      }
      return response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新集群失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteCluster = async (id: number) => {
    try {
      loading.value = true
      await api.delete(`/clusters/${id}`)
      clusters.value = clusters.value.filter(c => Number(c.id) !== id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '删除集群失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  const testCluster = async (id: number) => {
    try {
      const response = await api.post<ApiResponse<any>>(`/clusters/${id}/test`)
      return response.data.data
    } catch (err) {
      throw err
    }
  }

  return {
    clusters: computed(() => clusters.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    fetchClusters,
    addCluster,
    updateCluster,
    deleteCluster,
    testCluster
  }
}

/**
 * Pod相关数据获取
 */
export function usePods() {
  const pods = ref<Pod[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)

  const searchPods = async (params: PodSearchRequest) => {
    try {
      loading.value = true
      error.value = null
      const response = await api.post<ApiResponse<PodSearchResponse>>('/pods/search', params)
      pods.value = response.data.data.pods
      total.value = response.data.data.total
      return response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '搜索Pod失败'
      console.error('搜索Pod失败:', err)
    } finally {
      loading.value = false
    }
  }

  const getPodDetail = async (clusterName: string, namespace: string, podName: string) => {
    try {
      const response = await api.get<ApiResponse<Pod>>(`/pods/detail`, {
        params: { cluster_name: clusterName, namespace, pod_name: podName }
      })
      return response.data.data
    } catch (err) {
      throw err
    }
  }

  return {
    pods: computed(() => pods.value),
    total: computed(() => total.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    searchPods,
    getPodDetail
  }
}

/**
 * 历史数据获取
 */
export function useHistory() {
  const history = ref<any[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)

  const fetchHistory = async (params: HistoryQueryRequest) => {
    try {
      loading.value = true
      error.value = null
      const response = await api.post<ApiResponse<HistoryQueryResponse>>('/history/query', params)
      history.value = response.data.data.data
      total.value = response.data.data.total
      return response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取历史数据失败'
      console.error('获取历史数据失败:', err)
    } finally {
      loading.value = false
    }
  }

  return {
    history: computed(() => history.value),
    total: computed(() => total.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    fetchHistory
  }
}

/**
 * 资源分析数据获取
 */
export function useAnalysis() {
  const analysis = ref<ResourceAnalysis | null>(null)
  const topMemoryPods = ref<Pod[]>([])
  const topCpuPods = ref<Pod[]>([])
  const namespaceSummary = ref<NamespaceSummary[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 分页相关状态
  const pagination = ref({
    page: 1,
    size: 10,
    total: 0,
    total_pages: 0,
    has_next: false,
    has_prev: false
  })

  // 筛选相关状态
  const filters = ref({
    cluster_name: ''
  })

  const fetchAnalysis = async (page = 1, size = 50, clusterName = '') => {
    try {
      loading.value = true
      error.value = null
      const params: any = { page, size }
      if (clusterName) {
        params.cluster_name = clusterName
      }
      
      const response = await api.get<ApiResponse<any>>('/analysis', { params })
      
      // 处理新的分页响应格式
      if (response.data.data && (response.data as any).pagination) {
        analysis.value = response.data.data
        pagination.value = (response.data as any).pagination
        filters.value.cluster_name = (response.data as any).filter?.cluster_name || ''
      } else {
        // 兼容旧格式
        analysis.value = response.data.data || response.data
      }
      
      return response.data.data || response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取分析数据失败'
      console.error('获取分析数据失败:', err)
    } finally {
      loading.value = false
    }
  }

  // 专门获取问题Pod分页数据的接口，使用新的后端接口结构
  const fetchProblemsWithPagination = async (page = 1, size = 10, clusterName = '', sortBy = 'total_waste') => {
    try {
      loading.value = true
      error.value = null
      const params: any = { page, size }
      if (clusterName) {
        params.cluster_name = clusterName
      }
      if (sortBy) {
        params.sort_by = sortBy
      }
      
      const response = await api.get<ApiResponse<any>>('/pods/problems', { params })
      
      // 后端新接口返回结构：{code: 0, data: {cluster_name, data: [...], pagination: {...}, sort_by}, msg}
      const responseData = response.data.data // 获取实际数据部分
      
      // 如果analysis.value不存在，先初始化基础结构
      if (!analysis.value) {
        analysis.value = {
          total_pods: 0,
          unreasonable_pods: 0,
          top50_problems: [],
          generated_at: new Date().toISOString(),
          clusters_analyzed: 0
        }
      }
      
      // 更新问题Pod数据 - Pod数组在responseData.data中
      if (responseData && responseData.data && Array.isArray(responseData.data)) {
        analysis.value.top50_problems = responseData.data
        // 同时更新统计数据
        analysis.value.unreasonable_pods = responseData.pagination?.total || responseData.data.length
      } else {
        analysis.value.top50_problems = []
        console.warn('API返回的pods数据格式异常:', responseData)
      }
      
      // 更新分页信息 - 分页信息在responseData.pagination中
      if (responseData.pagination) {
        pagination.value = {
          page: responseData.pagination.page || 1,
          size: responseData.pagination.size || 10,
          total: responseData.pagination.total || 0,
          total_pages: responseData.pagination.total_pages || 0,
          has_next: responseData.pagination.has_next || false,
          has_prev: responseData.pagination.has_prev || false
        }
      } else {
        // 如果没有分页信息，使用默认值
        pagination.value = {
          page: 1,
          size: 10,
          total: 0,
          total_pages: 0,
          has_next: false,
          has_prev: false
        }
      }
      
      // 更新筛选信息
      filters.value.cluster_name = responseData.cluster_name || clusterName || ''
      
      return responseData
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取问题Pod数据失败'
      console.error('获取问题Pod数据失败:', err)
      
      // 错误时确保数据结构正确
      if (analysis.value) {
        analysis.value.top50_problems = []
      }
      pagination.value = {
        page: 1,
        size: 10,
        total: 0,
        total_pages: 0,
        has_next: false,
        has_prev: false
      }
      
      throw err
    } finally {
      loading.value = false
    }
  }

  const fetchTopMemoryPods = async (limit: number = 20) => {
    try {
      const response = await api.get<ApiResponse<any>>('/statistics/top-memory-request', {
        params: { limit }
      })
      // 确保返回的数据是数组格式
      if (response.data && Array.isArray(response.data.data)) {
        topMemoryPods.value = response.data.data
      } else {
        topMemoryPods.value = []
        console.warn('Top内存Pod数据格式异常，使用空数组')
      }
      return response.data
    } catch (err) {
      topMemoryPods.value = [] // 确保发生错误时也是数组
      console.error('获取Top内存Pod失败:', err)
      throw err
    }
  }

  const fetchTopCpuPods = async (limit: number = 20) => {
    try {
      const response = await api.get<ApiResponse<any>>('/statistics/top-cpu-request', {
        params: { limit }
      })
      // 确保返回的数据是数组格式
      if (response.data && Array.isArray(response.data.data)) {
        topCpuPods.value = response.data.data
      } else {
        topCpuPods.value = []
        console.warn('Top CPU Pod数据格式异常，使用空数组')
      }
      return response.data
    } catch (err) {
      topCpuPods.value = [] // 确保发生错误时也是数组
      console.error('获取Top CPU Pod失败:', err)
      throw err
    }
  }

  const fetchNamespaceSummary = async () => {
    try {
      const response = await api.get<ApiResponse<any>>('/statistics/namespace-summary')
      // 确保返回的数据是数组格式
      if (response.data && Array.isArray(response.data.data)) {
        namespaceSummary.value = response.data.data
      } else {
        namespaceSummary.value = []
        console.warn('命名空间汇总数据格式异常，使用空数组')
      }
      return response.data
    } catch (err) {
      namespaceSummary.value = [] // 确保发生错误时也是数组
      console.error('获取命名空间汇总失败:', err)
      throw err
    }
  }

  const triggerDataCollection = async (enablePersistence: boolean = true) => {
    try {
      loading.value = true
      const response = await api.post<ApiResponse<ResourceAnalysis>>('/history/collect', null, {
        params: { persistence: enablePersistence }
      })
      analysis.value = response.data.data || response.data
      return response.data.data || response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '触发数据收集失败'
      console.error('触发数据收集失败:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  const refreshAllData = async () => {
    try {
      loading.value = true
      error.value = null
      
      // 并行获取所有分析数据
      const [analysisResult, topMemoryResult, topCpuResult, namespaceResult] = await Promise.allSettled([
        fetchAnalysis(),
        fetchTopMemoryPods(20),
        fetchTopCpuPods(20),
        fetchNamespaceSummary()
      ])

      // 处理结果，即使部分失败也要展示成功的数据
      if (analysisResult.status === 'rejected') {
        console.warn('分析数据获取失败:', analysisResult.reason)
      }
      if (topMemoryResult.status === 'rejected') {
        console.warn('Top内存数据获取失败:', topMemoryResult.reason)
      }
      if (topCpuResult.status === 'rejected') {
        console.warn('Top CPU数据获取失败:', topCpuResult.reason)
      }
      if (namespaceResult.status === 'rejected') {
        console.warn('命名空间数据获取失败:', namespaceResult.reason)
      }

      return {
        analysis: analysisResult.status === 'fulfilled' ? analysisResult.value : null,
        topMemory: topMemoryResult.status === 'fulfilled' ? topMemoryResult.value : null,
        topCpu: topCpuResult.status === 'fulfilled' ? topCpuResult.value : null,
        namespace: namespaceResult.status === 'fulfilled' ? namespaceResult.value : null
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '刷新数据失败'
      console.error('刷新数据失败:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    analysis: computed(() => analysis.value),
    topMemoryPods: computed(() => topMemoryPods.value),
    topCpuPods: computed(() => topCpuPods.value),
    namespaceSummary: computed(() => namespaceSummary.value),
    pagination: computed(() => pagination.value),
    filters: computed(() => filters.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    fetchAnalysis,
    fetchProblemsWithPagination,
    fetchTopMemoryPods,
    fetchTopCpuPods,
    fetchNamespaceSummary,
    triggerDataCollection,
    refreshAllData
  }
}

/**
 * 命名空间数据获取
 */
export function useNamespaces() {
  const namespaces = ref<NamespaceSummary[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const fetchNamespaces = async (clusterName?: string) => {
    try {
      loading.value = true
      error.value = null
      const params = clusterName ? { cluster_name: clusterName } : {}
      const response = await api.get<ApiResponse<NamespaceSummary[]>>('/namespaces/summary', { params })
      namespaces.value = response.data.data
      return response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取命名空间数据失败'
      console.error('获取命名空间数据失败:', err)
    } finally {
      loading.value = false
    }
  }

  return {
    namespaces: computed(() => namespaces.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    fetchNamespaces
  }
}

/**
 * 系统统计数据获取
 */
export function useStats() {
  const systemStore = useSystemStore()

  const refreshStats = async () => {
    await systemStore.fetchStats()
  }

  return {
    stats: computed(() => systemStore.stats),
    loading: computed(() => systemStore.loading),
    refreshStats
  }
}

/**
 * 调度任务数据获取
 */
export function useSchedule() {
  const jobs = ref<any[]>([])
  const settings = ref(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const fetchJobs = async () => {
    try {
      loading.value = true
      error.value = null
      const response = await api.get<ApiResponse<any[]>>('/schedule/jobs')
      jobs.value = response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取调度任务失败'
      console.error('获取调度任务失败:', err)
    } finally {
      loading.value = false
    }
  }

  const fetchSettings = async () => {
    try {
      const response = await api.get<ApiResponse<any>>('/schedule/settings')
      settings.value = response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取调度设置失败'
      console.error('获取调度设置失败:', err)
    }
  }

  const updateSettings = async (newSettings: any) => {
    try {
      loading.value = true
      const response = await api.put<ApiResponse<any>>('/schedule/settings', newSettings)
      settings.value = response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新调度设置失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  const startJob = async (clusterId: number) => {
    try {
      await api.post(`/schedule/jobs/${clusterId}/start`)
      await fetchJobs()
    } catch (err) {
      throw err
    }
  }

  const stopJob = async (clusterId: number) => {
    try {
      await api.post(`/schedule/jobs/${clusterId}/stop`)
      await fetchJobs()
    } catch (err) {
      throw err
    }
  }

  return {
    jobs: computed(() => jobs.value),
    settings: computed(() => settings.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    fetchJobs,
    fetchSettings,
    updateSettings,
    startJob,
    stopJob
  }
}

/**
 * 通用数据刷新
 */
export function useRefresh() {
  const refreshing = ref(false)

  const refresh = async (refreshFn: () => Promise<any>) => {
    if (refreshing.value) return
    
    try {
      refreshing.value = true
      await refreshFn()
    } catch (err) {
      console.error('刷新数据失败:', err)
      throw err
    } finally {
      refreshing.value = false
    }
  }

  return {
    refreshing: computed(() => refreshing.value),
    refresh
  }
}

/**
 * 活动优化配置管理
 */
export function useActivityOptimization() {
  const config = ref(null)
  const optimizationResult = ref(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 获取优化配置
  const fetchOptimizationConfig = async () => {
    try {
      loading.value = true
      error.value = null
      const response = await api.get<ApiResponse<any>>('/activities/optimization/config')
      config.value = response.data.data
      return response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取优化配置失败'
      console.error('获取优化配置失败:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  // 更新优化配置
  const updateOptimizationConfig = async (newConfig: any) => {
    try {
      loading.value = true
      error.value = null
      const response = await api.put<ApiResponse<any>>('/activities/optimization/config', newConfig)
      config.value = response.data.data
      return response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新优化配置失败'
      console.error('更新优化配置失败:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  // 执行活动优化
  const executeOptimization = async () => {
    try {
      loading.value = true
      error.value = null
      const response = await api.post<ApiResponse<any>>('/activities/optimization/execute')
      optimizationResult.value = response.data.data
      return response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '执行优化失败'
      console.error('执行优化失败:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  // 获取活动统计
  const getActivityStats = async (hours: number = 24) => {
    try {
      const response = await api.get<ApiResponse<any>>('/activities/stats', {
        params: { hours }
      })
      return response.data.data
    } catch (err) {
      console.error('获取活动统计失败:', err)
      throw err
    }
  }

  return {
    config: computed(() => config.value),
    optimizationResult: computed(() => optimizationResult.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    fetchOptimizationConfig,
    updateOptimizationConfig,
    executeOptimization,
    getActivityStats
  }
}