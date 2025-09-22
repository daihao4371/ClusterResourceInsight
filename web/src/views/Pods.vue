<template>
  <div class="space-y-4 animate-fade-in">
    <!-- 错误提示 -->
    <div v-if="error" class="bg-danger-500/20 border border-danger-500/30 rounded-lg p-4 mb-4">
      <div class="flex items-center space-x-2">
        <AlertTriangle class="w-5 h-5 text-danger-400" />
        <span class="text-danger-400 font-medium">加载失败</span>
      </div>
      <p class="text-sm text-danger-300 mt-2">{{ error }}</p>
      <button @click="refreshData" class="btn-secondary mt-3">
        重新加载
      </button>
    </div>
    <!-- 页面标题和操作 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gradient">Pod监控</h1>
        <p class="text-gray-400 text-sm">实时监控所有Pod状态和资源使用情况</p>
      </div>
      
      <div class="flex items-center space-x-3">
        <button @click="refreshData" class="btn-secondary" :disabled="refreshing">
          <RefreshCw class="w-4 h-4 mr-2" :class="{ 'animate-spin': refreshing }" />
          {{ refreshing ? '刷新中...' : '刷新数据' }}
        </button>
        <button @click="runApiTest" class="btn-secondary">
          <Eye class="w-4 h-4 mr-2" />
          调试数据
        </button>
        <button class="btn-secondary">
          <Download class="w-4 h-4 mr-2" />
          导出报告
        </button>
      </div>
    </div>

    <!-- 统计概览 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <MetricCard
        title="运行中的Pod"
        :value="podStats.running"
        icon="Box"
        status="success"
        trend="+2.5%"
      />
      <MetricCard
        title="待调度Pod"
        :value="podStats.pending"
        icon="Clock"
        status="warning"
        trend="-12%"
      />
      <MetricCard
        title="失败Pod"
        :value="podStats.failed"
        icon="AlertTriangle"
        status="error"
        trend="-8.3%"
      />
      <MetricCard
        title="CPU使用率"
        :value="podStats.avgCpuUsage"
        unit="%"
        icon="Activity"
        :status="podStats.avgCpuUsage > 80 ? 'error' : 'success'"
        trend="+1.2%"
      />
    </div>

    <!-- 筛选和搜索 -->
    <div class="glass-card p-4">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex-1 min-w-64">
          <div class="relative">
            <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索Pod名称、命名空间..."
              class="input-field pl-10"
            />
          </div>
        </div>
        
        <select v-model="statusFilter" class="input-field">
          <option value="">所有状态</option>
          <option v-for="status in availableStatuses" :key="status" :value="status">
            {{ formatStatusDisplay(status) }}
          </option>
        </select>
        
        <select v-model="namespaceFilter" class="input-field">
          <option value="">所有命名空间</option>
          <option v-for="namespace in availableNamespaces" :key="namespace" :value="namespace">
            {{ namespace }}
          </option>
        </select>
        
        <select v-model="clusterFilter" class="input-field">
          <option value="">所有集群</option>
          <option v-for="cluster in availableClusters" :key="cluster" :value="cluster">
            {{ cluster }}
          </option>
        </select>
      </div>
    </div>

    <!-- Pod列表表格 -->
    <div class="glass-card overflow-hidden">
      <!-- 加载状态 -->
      <div v-if="loading && pods.length === 0" class="text-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500 mx-auto mb-4"></div>
        <p class="text-gray-400">正在加载Pod数据...</p>
      </div>
      
      <!-- 空状态 -->
      <div v-else-if="!loading && filteredPods.length === 0" class="text-center py-12">
        <Box class="w-16 h-16 text-gray-600 mx-auto mb-4" />
        <h3 class="text-lg font-semibold text-gray-400 mb-2">暂无Pod数据</h3>
        <p class="text-gray-500">{{ error ? '加载失败' : (searchQuery ? '未找到匹配的Pod' : '当前没有运行的Pod') }}</p>
        <button v-if="error" @click="refreshData" class="btn-secondary mt-4">
          重新加载
        </button>
      </div>
      
      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-dark-800/50">
            <tr class="border-b border-gray-700">
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Pod信息
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                状态
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                资源使用
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                重启次数
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                运行时间
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                操作
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-700">
            <tr
              v-for="pod in paginatedPods"
              :key="pod.id"
              class="hover:bg-white/5 transition-colors"
            >
              <!-- Pod信息 -->
              <td class="px-6 py-4">
                <div class="flex items-center space-x-3">
                  <div 
                    class="status-indicator"
                    :class="getStatusIndicatorClass(pod.status)"
                  ></div>
                  <div>
                    <div class="font-medium text-white">{{ pod.name }}</div>
                    <div class="text-sm text-gray-400">{{ pod.namespace }}/{{ pod.cluster }}</div>
                  </div>
                </div>
              </td>
              
              <!-- 状态 -->
              <td class="px-6 py-4">
                <span 
                  class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="getStatusBadgeClass(pod.status)"
                >
                  {{ pod.status }}
                </span>
              </td>
              
              <!-- 资源使用 -->
              <td class="px-6 py-4">
                <div class="space-y-2">
                  <div class="flex items-center space-x-2">
                    <span class="text-xs text-gray-400 w-12">CPU:</span>
                    <div class="flex-1 bg-dark-700 rounded-full h-1.5 max-w-20">
                      <div 
                        class="h-1.5 rounded-full transition-all duration-1000"
                        :class="getResourceBarClass(pod.cpuUsage)"
                        :style="{ width: `${Math.min(pod.cpuUsage, 100)}%` }"
                      ></div>
                    </div>
                    <span class="text-xs text-gray-300 w-10">{{ pod.cpuUsage }}%</span>
                  </div>
                  <div class="flex items-center space-x-2">
                    <span class="text-xs text-gray-400 w-12">MEM:</span>
                    <div class="flex-1 bg-dark-700 rounded-full h-1.5 max-w-20">
                      <div 
                        class="h-1.5 rounded-full transition-all duration-1000"
                        :class="getResourceBarClass(pod.memoryUsage)"
                        :style="{ width: `${Math.min(pod.memoryUsage, 100)}%` }"
                      ></div>
                    </div>
                    <span class="text-xs text-gray-300 w-10">{{ pod.memoryUsage }}%</span>
                  </div>
                </div>
              </td>
              
              <!-- 重启次数 -->
              <td class="px-6 py-4">
                <span 
                  class="text-sm"
                  :class="pod.restarts > 0 ? 'text-warning-400' : 'text-gray-300'"
                >
                  {{ pod.restarts }}
                </span>
              </td>
              
              <!-- 运行时间 -->
              <td class="px-6 py-4 text-sm text-gray-300">
                {{ formatDistanceToNow(pod.startTime) }}
              </td>
              
              <!-- 操作 - 保留核心功能按钮 -->
              <td class="px-6 py-4">
                <div class="flex items-center space-x-2">
                  <button 
                    @click="viewPodDetail(pod)"
                    class="p-1 hover:bg-white/10 rounded transition-colors"
                    title="查看详情"
                  >
                    <Eye class="w-4 h-4 text-primary-400" />
                  </button>
                  <button 
                    @click="viewPodTrend(pod)"
                    class="p-1 hover:bg-white/10 rounded transition-colors"
                    title="资源趋势"
                  >
                    <TrendingUp class="w-4 h-4 text-success-400" />
                  </button>
                  <button 
                    v-if="pod.status === 'Failed'"
                    @click="restartPod(pod)"
                    class="p-1 hover:bg-white/10 rounded transition-colors"
                    title="重启Pod"
                  >
                    <RotateCw class="w-4 h-4 text-warning-400" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      
      <!-- 列表末尾加载状态 -->
      <div v-if="loading && pods.length > 0" class="px-6 py-4 border-t border-gray-700 text-center">
        <div class="flex items-center justify-center space-x-2">
          <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary-500"></div>
          <span class="text-sm text-gray-400">加载中...</span>
        </div>
      </div>
      
      <!-- 分页 -->
      <div v-if="filteredPods.length > 0" class="px-6 py-4 border-t border-gray-700 flex items-center justify-between">
        <div class="text-sm text-gray-400">
          显示 {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, filteredPods.length) }} 
          共 {{ filteredPods.length }} 条
        </div>
        <div class="flex items-center space-x-2">
          <button 
            @click="currentPage--"
            :disabled="currentPage === 1"
            class="btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            上一页
          </button>
          <span class="text-sm text-gray-400">
            {{ currentPage }} / {{ totalPages }}
          </span>
          <button 
            @click="currentPage++"
            :disabled="currentPage === totalPages"
            class="btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <!-- Pod详细分析模态框 -->
    <PodDetailModal
      :visible="showDetailModal"
      :cluster="selectedPod?.cluster"
      :namespace="selectedPod?.namespace"
      :pod-name="selectedPod?.name"
      @close="showDetailModal = false"
      @open-trend="openTrendModal"
    />

    <!-- Pod趋势图表模态框 -->
    <PodTrendModal
      :visible="showTrendModal"
      :cluster="selectedPod?.cluster"
      :namespace="selectedPod?.namespace"
      :pod-name="selectedPod?.name"
      @close="showTrendModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { 
  Box,
  AlertTriangle,
  Search,
  RefreshCw,
  Download,
  Eye,
  RotateCw,
  TrendingUp
} from 'lucide-vue-next'
import MetricCard from '../components/common/MetricCard.vue'
import PodDetailModal from '../components/modals/PodDetailModal.vue'
import PodTrendModal from '../components/modals/PodTrendModal.vue'
import { formatDistanceToNow } from '../utils/date'
import PodsApiService, { Pod, PodStats } from '../api/pods'

// Pod接口定义已移到api/pods.ts中，这里创建数据转换函数
interface DisplayPod {
  id: string
  name: string
  namespace: string
  cluster: string
  status: 'Running' | 'Pending' | 'Failed' | 'Succeeded' | string
  cpuUsage: number
  memoryUsage: number
  restarts: number
  startTime: string
}

// 将后端Pod数据转换为前端显示格式
const transformPodData = (pod: Pod): DisplayPod => {
  return {
    id: `${pod.cluster_name}-${pod.namespace}-${pod.pod_name}`,
    name: pod.pod_name,
    namespace: pod.namespace,
    cluster: pod.cluster_name,
    status: mapPodStatus(pod.status),
    cpuUsage: Math.round(pod.cpu_req_pct || 0),
    memoryUsage: Math.round(pod.memory_req_pct || 0),
    restarts: 0, // 后端暂未提供此字段
    startTime: pod.creation_time
  }
}

// 映射Pod状态 - 后端状态到前端显示状态的映射
const mapPodStatus = (status: string): 'Running' | 'Pending' | 'Failed' | 'Succeeded' | string => {
  // 确保状态字符串不为空并且进行正确的映射
  if (!status) return 'Pending'
  
  const normalizedStatus = status.trim()
  
  // 根据后端实际返回的状态值进行映射
  if (normalizedStatus === '合理') return 'Running'
  if (normalizedStatus === '不合理') return 'Failed'
  
  // 如果是其他状态，直接返回（可能是英文状态）
  return normalizedStatus
}

// 映射前端状态筛选到后端状态
const mapStatusFilter = (frontendStatus: string): string => {
  const statusMap: Record<string, string> = {
    'Running': '合理',
    'Failed': '不合理',
    'Pending': '',  // 后端可能没有对应的状态
    'Succeeded': '',  // 后端可能没有对应的状态
    '正常': '合理',
    '异常': '不合理'
  }
  return statusMap[frontendStatus] || frontendStatus
}

// 筛选状态
const searchQuery = ref('')
const statusFilter = ref('')
const namespaceFilter = ref('')
const clusterFilter = ref('')

// 分页
const currentPage = ref(1)
const pageSize = ref(20)

// 统计数据
const podStats = ref<PodStats>({
  running: 0,
  pending: 0,
  failed: 0,
  avgCpuUsage: 0
})

// 加载状态和错误处理
const loading = ref(false)
const error = ref<string | null>(null)
const refreshing = ref(false)

// Pod数据
const pods = ref<DisplayPod[]>([])        // 原始Pod数据列表
const rawPods = ref<Pod[]>([])            // 从后端获取的原始数据

// 筛选选项数据
const availableNamespaces = ref<string[]>([])
const availableClusters = ref<string[]>([])
const availableStatuses = ref<string[]>([])

// 模态框状态
const showDetailModal = ref(false)
const showTrendModal = ref(false)
const selectedPod = ref<DisplayPod | null>(null)

// 格式化状态显示文本
const formatStatusDisplay = (status: string): string => {
  const statusMap: Record<string, string> = {
    '合理': '正常',
    '不合理': '异常',
    'Running': '运行中',
    'Pending': '待调度',
    'Failed': '失败',
    'Succeeded': '成功'
  }
  return statusMap[status] || status
}

// 筛选后的Pod列表
const filteredPods = computed(() => {
  return pods.value.filter(pod => {
    const matchesSearch = !searchQuery.value || 
      pod.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      pod.namespace.toLowerCase().includes(searchQuery.value.toLowerCase())
    
    const matchesStatus = !statusFilter.value || pod.status === statusFilter.value
    const matchesNamespace = !namespaceFilter.value || pod.namespace === namespaceFilter.value
    const matchesCluster = !clusterFilter.value || pod.cluster === clusterFilter.value
    
    return matchesSearch && matchesStatus && matchesNamespace && matchesCluster
  })
})

// 分页后的Pod列表
const paginatedPods = computed(() => {
  const filtered = filteredPods.value
  const startIndex = (currentPage.value - 1) * pageSize.value
  const endIndex = startIndex + pageSize.value
  return filtered.slice(startIndex, endIndex)
})

// 总页数
const totalPages = computed(() => Math.ceil(filteredPods.value.length / pageSize.value))

// 样式方法
const getStatusIndicatorClass = (status: string) => {
  const classes: Record<string, string> = {
    Running: 'status-online',
    Pending: 'status-warning',
    Failed: 'status-error',
    Succeeded: 'status-online'
  }
  return classes[status] || 'status-warning'
}

const getStatusBadgeClass = (status: string) => {
  const classes: Record<string, string> = {
    Running: 'bg-success-500/20 text-success-400 border border-success-500/30',
    Pending: 'bg-warning-500/20 text-warning-400 border border-warning-500/30',
    Failed: 'bg-danger-500/20 text-danger-400 border border-danger-500/30',
    Succeeded: 'bg-success-500/20 text-success-400 border border-success-500/30'
  }
  return classes[status] || 'bg-warning-500/20 text-warning-400 border border-warning-500/30'
}

const getResourceBarClass = (usage: number) => {
  if (usage >= 80) return 'bg-gradient-to-r from-danger-600 to-danger-400'
  if (usage >= 60) return 'bg-gradient-to-r from-warning-600 to-warning-400'
  return 'bg-gradient-to-r from-success-600 to-success-400'
}

// API数据加载方法
const loadPodsData = async () => {
  try {
    loading.value = true
    error.value = null
    
    // 构建搜索参数
    const searchParams = {
      page: currentPage.value,
      size: pageSize.value,
      query: searchQuery.value || undefined,
      namespace: namespaceFilter.value || undefined,
      cluster: clusterFilter.value || undefined,
      status: statusFilter.value ? mapStatusFilter(statusFilter.value) : undefined
    }
    
    // 调用API获取Pod数据
    const response = await PodsApiService.getPodsWithSearch(searchParams)
    
    // 处理后端统一响应格式 {code: 0, data: {...}, message: "操作成功"}
    if (response && response.code === 0 && response.data) {
      // 从响应中提取实际的Pod数据 - 后端返回PodSearchResponse结构
      const podSearchResponse = response.data
      rawPods.value = podSearchResponse.pods || []
      pods.value = rawPods.value.map(transformPodData)
      
      // 更新统计数据
      updateStatsFromData()
      
      console.log('Pod数据加载成功:', {
        total: podSearchResponse.total,
        count: rawPods.value.length,
        page: podSearchResponse.page,
        筛选条件: searchParams
      })
    } else {
      throw new Error(response?.message || '响应格式错误')
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载Pod数据失败'
    console.error('加载Pod数据失败:', err)
  } finally {
    loading.value = false
  }
}

// 从数据中计算统计信息
const updateStatsFromData = () => {
  const running = pods.value.filter(p => p.status === 'Running').length
  const pending = pods.value.filter(p => p.status === 'Pending').length
  const failed = pods.value.filter(p => p.status === 'Failed').length
  const avgCpu = pods.value.length > 0 
    ? Math.round(pods.value.reduce((sum, p) => sum + p.cpuUsage, 0) / pods.value.length)
    : 0
    
  podStats.value = {
    running,
    pending,
    failed,
    avgCpuUsage: avgCpu
  }
}

// 加载统计数据
const loadStatsData = async () => {
  try {
    const statsResponse = await PodsApiService.getPodStats()
    // 处理后端统一响应格式 {code: 0, data: {...}, msg: "操作成功"}
    if (statsResponse && statsResponse.code === 0 && statsResponse.data) {
      const stats = statsResponse.data
      podStats.value = {
        running: (stats.total_pods || 0) - (stats.unreasonable_pods || 0),
        pending: 0, // 系统统计API可能不提供此数据
        failed: stats.unreasonable_pods || 0,
        avgCpuUsage: Math.round(stats.avg_cpu_usage || 0)
      }
      
      console.log('统计数据加载成功:', podStats.value)
    } else {
      throw new Error(statsResponse?.message || '统计数据响应格式错误')
    }
  } catch (err) {
    console.warn('加载统计数据失败，使用默认值:', err)
    // 从当前Pod数据计算统计信息作为后备方案
    updateStatsFromData()
  }
}

// 加载筛选选项 - 支持按集群筛选命名空间
const loadFilterOptions = async (cluster?: string) => {
  try {
    const response = await PodsApiService.getFilterOptions(cluster)
    
    // 处理后端统一响应格式 {code: 0, data: {...}, msg: "操作成功"}
    if (response && response.code === 0 && response.data) {
      const filterOptions = response.data
      availableNamespaces.value = filterOptions.namespaces || []
      availableClusters.value = filterOptions.clusters || []
      availableStatuses.value = filterOptions.statuses || []
      
      console.log('筛选选项加载成功:', {
        namespaces: availableNamespaces.value.length,
        clusters: availableClusters.value.length,
        statuses: availableStatuses.value.length,
        集群筛选: cluster || '全部'
      })
    } else {
      throw new Error(response?.message || '筛选选项响应格式错误')
    }
  } catch (err) {
    console.warn('加载筛选选项失败，使用默认值:', err)
    // 提供默认的筛选选项作为后备方案
    availableNamespaces.value = ['default', 'kube-system', 'monitoring']
    availableClusters.value = ['prod-cluster-01', 'dev-cluster-02', 'test-cluster-03']
    availableStatuses.value = ['合理', '不合理']
  }
}

// 刷新数据
const refreshData = async () => {
  refreshing.value = true
  await Promise.all([
    loadPodsData(),
    loadStatsData(),
    loadFilterOptions() // 添加筛选选项的加载
  ])
  refreshing.value = false
}

// 操作方法
const viewPodDetail = (pod: DisplayPod) => {
  selectedPod.value = pod
  showDetailModal.value = true
}

const viewPodTrend = (pod: DisplayPod) => {
  selectedPod.value = pod
  showTrendModal.value = true
}

const openTrendModal = (cluster: string, namespace: string, podName: string) => {
  // 从详情模态框打开趋势模态框
  showDetailModal.value = false
  // 更新selectedPod以匹配传入的参数
  if (selectedPod.value) {
    selectedPod.value.cluster = cluster
    selectedPod.value.namespace = namespace
    selectedPod.value.name = podName
  }
  showTrendModal.value = true
}

const restartPod = (pod: DisplayPod) => {
  console.log('重启Pod:', pod)
}

// API测试方法 - 输出当前数据状态用于调试
const runApiTest = async () => {
  console.log('🔧 开始运行数据状态检查...')
  
  console.log('📋 当前Pod数据状态:')
  console.log('- 原始Pod数量:', rawPods.value.length)
  console.log('- 转换后Pod数量:', pods.value.length)
  console.log('- 筛选后Pod数量:', filteredPods.value.length)
  console.log('- 当前页Pod数量:', paginatedPods.value.length)
  console.log('- 统计数据:', podStats.value)
  
  if (pods.value.length > 0) {
    console.log('- 第一个Pod示例:', pods.value[0])
  }
  
  // 重新加载数据以确保最新状态
  await refreshData()
}

// 页面初始化
onMounted(async () => {
  await Promise.all([
    loadPodsData(),
    loadStatsData(),
    loadFilterOptions() // 添加筛选选项的加载
  ])
})

// 监听筛选条件变化 - 使用防抖避免频繁请求
watch([searchQuery, statusFilter, namespaceFilter, clusterFilter], () => {
  currentPage.value = 1
  loadPodsData()
})

// 监听集群筛选变化 - 当集群切换时重新加载筛选选项并清空命名空间筛选
watch(clusterFilter, async (newCluster, oldCluster) => {
  // 只有当集群真正发生变化时才执行
  if (newCluster !== oldCluster) {
    // 清空当前的命名空间筛选，避免显示错误的命名空间
    namespaceFilter.value = ''
    
    // 重新加载筛选选项，传入选中的集群参数
    await loadFilterOptions(newCluster || undefined)
    
    console.log('集群筛选变化:', {
      从: oldCluster || '全部',
      到: newCluster || '全部',
      命名空间已清空: true
    })
  }
})

// 监听分页变化
watch([currentPage, pageSize], () => {
  loadPodsData()
})
</script>

<style scoped>
table {
  border-collapse: separate;
  border-spacing: 0;
}

th:first-child {
  border-top-left-radius: 0.5rem;
}

th:last-child {
  border-top-right-radius: 0.5rem;
}
</style>