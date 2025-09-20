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
      const response = await api.get<ApiResponse<Cluster[]>>('/clusters')
      clusters.value = response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取集群数据失败'
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
      const index = clusters.value.findIndex(c => c.id === id)
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
      clusters.value = clusters.value.filter(c => c.id !== id)
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
  const history = ref([])
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
  const loading = ref(false)
  const error = ref<string | null>(null)

  const fetchAnalysis = async () => {
    try {
      loading.value = true
      error.value = null
      const response = await api.get<ApiResponse<ResourceAnalysis>>('/analysis')
      analysis.value = response.data.data
      return response.data.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取分析数据失败'
      console.error('获取分析数据失败:', err)
    } finally {
      loading.value = false
    }
  }

  return {
    analysis: computed(() => analysis.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    fetchAnalysis
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
  const jobs = ref([])
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