<template>
  <div class="space-y-4 animate-fade-in">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gradient">资源分析</h1>
        <p class="text-gray-400 text-sm">深度分析集群资源配置与使用效率</p>
      </div>
      
      <div class="flex items-center space-x-3">
        <button 
          class="btn-secondary"
          :disabled="loading"
          @click="refreshAllAnalysisData"
        >
          <RefreshCw :class="['w-4 h-4 mr-2', { 'animate-spin': loading }]" />
          {{ loading ? '分析中...' : '刷新分析' }}
        </button>
        <button 
          class="btn-secondary"
          @click="triggerFullCollection"
        >
          <Database class="w-4 h-4 mr-2" />
          深度收集
        </button>
      </div>
    </div>



    <!-- 深度分析功能区 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Top内存请求Pod -->
      <div class="glass-card">
        <div class="p-6 border-b border-gray-700">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">Top内存请求Pod</h3>
            <button 
              class="btn-secondary text-sm"
              @click="fetchTopMemoryPods(20)"
              :disabled="loading"
            >
              <RefreshCw :class="['w-3 h-3 mr-1', { 'animate-spin': loading }]" />
              刷新
            </button>
          </div>
        </div>
        <div class="p-4 max-h-80 overflow-y-auto">
          <div v-if="!topMemoryPods || topMemoryPods.length === 0" class="text-center py-8 text-gray-500">
            <Database class="w-8 h-8 mx-auto mb-2" />
            <p>暂无数据</p>
          </div>
          <div v-else class="space-y-3">
            <div 
              v-for="pod in (Array.isArray(topMemoryPods) ? topMemoryPods.slice(0, 10) : [])"
              :key="`memory-${pod.cluster_name}-${pod.namespace}-${pod.pod_name}`"
              class="flex items-center justify-between p-3 bg-dark-800/30 rounded-lg"
            >
              <div class="flex-1 min-w-0">
                <div class="font-medium text-sm truncate">{{ pod.pod_name }}</div>
                <div class="text-xs text-gray-400">{{ pod.namespace }}</div>
              </div>
              <div class="text-right">
                <div class="text-sm font-medium">{{ formatMemoryValue(pod.memory_request) }}</div>
                <div class="text-xs text-gray-500">请求量</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Top CPU请求Pod -->
      <div class="glass-card">
        <div class="p-6 border-b border-gray-700">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">Top CPU请求Pod</h3>
            <button 
              class="btn-secondary text-sm"
              @click="fetchTopCpuPods(20)"
              :disabled="loading"
            >
              <RefreshCw :class="['w-3 h-3 mr-1', { 'animate-spin': loading }]" />
              刷新
            </button>
          </div>
        </div>
        <div class="p-4 max-h-80 overflow-y-auto">
          <div v-if="!topCpuPods || topCpuPods.length === 0" class="text-center py-8 text-gray-500">
            <Cpu class="w-8 h-8 mx-auto mb-2" />
            <p>暂无数据</p>
          </div>
          <div v-else class="space-y-3">
            <div 
              v-for="pod in (Array.isArray(topCpuPods) ? topCpuPods.slice(0, 10) : [])"
              :key="`cpu-${pod.cluster_name}-${pod.namespace}-${pod.pod_name}`"
              class="flex items-center justify-between p-3 bg-dark-800/30 rounded-lg"
            >
              <div class="flex-1 min-w-0">
                <div class="font-medium text-sm truncate">{{ pod.pod_name }}</div>
                <div class="text-xs text-gray-400">{{ pod.namespace }}</div>
              </div>
              <div class="text-right">
                <div class="text-sm font-medium">{{ formatCpuValue(pod.cpu_request) }}</div>
                <div class="text-xs text-gray-500">请求量</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 命名空间汇总 -->
      <div class="glass-card">
        <div class="p-6 border-b border-gray-700">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">命名空间汇总</h3>
            <button 
              class="btn-secondary text-sm"
              @click="fetchNamespaceSummary()"
              :disabled="loading"
            >
              <RefreshCw :class="['w-3 h-3 mr-1', { 'animate-spin': loading }]" />
              刷新
            </button>
          </div>
        </div>
        <div class="p-4 max-h-80 overflow-y-auto">
          <div v-if="!namespaceSummary || namespaceSummary.length === 0" class="text-center py-8 text-gray-500">
            <Layers class="w-8 h-8 mx-auto mb-2" />
            <p>暂无数据</p>
          </div>
          <div v-else class="space-y-3">
            <div 
              v-for="ns in (Array.isArray(namespaceSummary) ? namespaceSummary.slice(0, 8) : [])"
              :key="`ns-${ns.cluster_name}-${ns.namespace_name}`"
              class="p-3 bg-dark-800/30 rounded-lg"
            >
              <div class="flex items-center justify-between mb-2">
                <div class="font-medium text-sm">{{ ns.namespace_name }}</div>
                <div class="text-xs text-gray-400">{{ ns.cluster_name }}</div>
              </div>
              <div class="grid grid-cols-2 gap-2 text-xs">
                <div>
                  <span class="text-gray-400">Pod数:</span>
                  <span class="ml-1 text-white">{{ ns.pod_count }}</span>
                </div>
                <div>
                  <span class="text-gray-400">运行中:</span>
                  <span class="ml-1 text-success-400">{{ ns.running_pods }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 资源分布图表 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- CPU资源分布 -->
      <div class="glass-card p-6">
        <h3 class="text-lg font-semibold mb-4">CPU资源分布</h3>
        <div class="h-64 flex items-center justify-center text-gray-500">
          <BarChart3 class="w-12 h-12 mr-3" />
          图表组件开发中...
        </div>
      </div>
      
      <!-- 内存资源分布 -->
      <div class="glass-card p-6">
        <h3 class="text-lg font-semibold mb-4">内存资源分布</h3>
        <div class="h-64 flex items-center justify-center text-gray-500">
          <PieChart class="w-12 h-12 mr-3" />
          图表组件开发中...
        </div>
      </div>
    </div>

    <!-- 问题Pod详细分析表格 -->
    <div class="glass-card">
      <div class="p-6 border-b border-gray-700">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h2 class="text-xl font-semibold">问题Pod详细分析</h2>
            <p class="text-sm text-gray-400 mt-1">深入分析资源配置不合理的Pod，优化集群资源使用效率</p>
          </div>
          <div class="flex items-center space-x-3">
            <select v-model="sortBy" @change="handleSortChange" class="input-field text-sm">
              <option value="cpu_waste">CPU浪费</option>
              <option value="memory_waste">内存浪费</option>
              <option value="total_waste">总浪费</option>
            </select>
            <button class="btn-secondary text-sm">
              <Filter class="w-4 h-4 mr-2" />
              筛选
            </button>
          </div>
        </div>
        
        <!-- 分页和筛选控制区域 -->
        <div class="flex items-center justify-between flex-wrap gap-4">
          <!-- 集群筛选器 -->
          <div class="flex items-center space-x-3">
            <label class="text-sm text-gray-400">集群筛选:</label>
            <select 
              v-model="selectedCluster" 
              @change="handleClusterChange"
              class="input-field text-sm w-48"
              :disabled="loading"
            >
              <option value="">全部集群 ({{ pagination.total }} 条)</option>
              <option 
                v-for="cluster in availableClusters" 
                :key="cluster" 
                :value="cluster"
              >
                {{ cluster }}
              </option>
            </select>
            <!-- 集群数量提示 -->
            <span class="text-xs text-gray-500">
              共 {{ availableClusters.length }} 个集群
            </span>
          </div>
          
          <!-- 每页数量选择器 -->
          <div class="flex items-center space-x-3">
            <label class="text-sm text-gray-400">每页显示:</label>
            <select 
              v-model="pageSize" 
              @change="handlePageSizeChange"
              class="input-field text-sm w-20"
            >
              <option value="10">10</option>
              <option value="30">30</option>
              <option value="50">50</option>
            </select>
            <span class="text-sm text-gray-400">条</span>
          </div>
          
          <!-- 刷新按钮 -->
          <button 
            class="btn-secondary text-sm"
            @click="refreshProblems"
            :disabled="loading"
          >
            <RefreshCw :class="['w-4 h-4 mr-2', { 'animate-spin': loading }]" />
            刷新
          </button>
        </div>
      </div>
      
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-dark-800/50">
            <tr class="border-b border-gray-700">
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">排名</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">Pod信息</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">配置问题</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">资源请求</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">实际使用</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">浪费程度</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">建议</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-700">
            <!-- 数据加载状态 -->
            <tr v-if="loading" class="animate-pulse">
              <td colspan="7" class="px-6 py-8 text-center text-gray-400">
                <RefreshCw class="w-6 h-6 mx-auto mb-2 animate-spin" />
                正在加载问题Pod数据...
              </td>
            </tr>
            
            <!-- 问题Pod数据行 -->
            <tr
              v-else-if="sortedProblems && sortedProblems.length > 0"
              v-for="(pod, index) in sortedProblems"
              :key="`${pod.cluster_name}-${pod.namespace}-${pod.pod_name}`"
              class="hover:bg-white/5 transition-colors"
            >
              <!-- 排名 -->
              <td class="px-6 py-4">
                <div class="flex items-center">
                  <span 
                    class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium"
                    :class="getRankBadgeClass((pagination.page - 1) * pagination.size + index + 1)"
                  >
                    {{ (pagination.page - 1) * pagination.size + index + 1 }}
                  </span>
                </div>
              </td>
              
              <!-- Pod信息 -->
              <td class="px-6 py-4">
                <div>
                  <div class="font-medium text-white">{{ pod.pod_name }}</div>
                  <div class="text-sm text-gray-400">{{ pod.namespace }}/{{ pod.cluster_name }}</div>
                  <div class="text-xs text-gray-500 mt-1">节点: {{ pod.node_name }}</div>
                </div>
              </td>
              
              <!-- 配置问题 -->
              <td class="px-6 py-4">
                <div class="space-y-1">
                  <span 
                    v-for="issue in pod.issues"
                    :key="issue"
                    class="inline-block px-2 py-1 text-xs rounded-full"
                    :class="getIssueBadgeClass(issue)"
                  >
                    {{ getIssueText(issue) }}
                  </span>
                </div>
              </td>
              
              <!-- 资源请求 -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div>CPU: {{ formatCpuValue(pod.cpu_request) }}</div>
                  <div>内存: {{ formatMemoryValue(pod.memory_request) }}</div>
                </div>
              </td>
              
              <!-- 实际使用 -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div class="flex items-center space-x-2">
                    <span>{{ formatCpuValue(pod.cpu_usage) }}</span>
                    <div 
                      v-if="pod.cpu_req_pct > 0"
                      class="w-16 h-1.5 bg-dark-700 rounded-full overflow-hidden"
                    >
                      <div 
                        class="h-full rounded-full transition-all"
                        :class="getUsageBarClass(pod.cpu_req_pct)"
                        :style="{ width: `${Math.min(pod.cpu_req_pct, 100)}%` }"
                      ></div>
                    </div>
                    <span class="text-xs text-gray-400">{{ pod.cpu_req_pct?.toFixed(1) }}%</span>
                  </div>
                  <div class="flex items-center space-x-2">
                    <span>{{ formatMemoryValue(pod.memory_usage) }}</span>
                    <div 
                      v-if="pod.memory_req_pct > 0"
                      class="w-16 h-1.5 bg-dark-700 rounded-full overflow-hidden"
                    >
                      <div 
                        class="h-full rounded-full transition-all"
                        :class="getUsageBarClass(pod.memory_req_pct)"
                        :style="{ width: `${Math.min(pod.memory_req_pct, 100)}%` }"
                      ></div>
                    </div>
                    <span class="text-xs text-gray-400">{{ pod.memory_req_pct?.toFixed(1) }}%</span>
                  </div>
                </div>
              </td>
              
              <!-- 浪费程度 -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div 
                    class="font-medium"
                    :class="getWasteClass(calculateWaste(pod))"
                  >
                    {{ calculateWaste(pod) }}%
                  </div>
                  <div class="text-xs text-gray-500">资源浪费</div>
                </div>
              </td>
              
              <!-- 建议 -->
              <td class="px-6 py-4">
                <button class="btn-secondary text-sm">
                  <Lightbulb class="w-4 h-4 mr-1" />
                  优化建议
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      
      <!-- 分页器 -->
      <div class="p-4 border-t border-gray-700">
        <div class="flex items-center justify-between">
          <!-- 分页信息 -->
          <div class="text-sm text-gray-400">
            显示第 {{ (pagination.page - 1) * pagination.size + 1 }} - 
            {{ Math.min(pagination.page * pagination.size, pagination.total) }} 条，
            共 {{ pagination.total }} 条记录
          </div>
          
          <!-- 分页控制 -->
          <div class="flex items-center space-x-2">
            <button 
              class="btn-secondary text-sm px-3 py-1"
              :disabled="!pagination.has_prev || loading"
              @click="goToPage(pagination.page - 1)"
            >
              <ChevronLeft class="w-4 h-4" />
              上一页
            </button>
            
            <!-- 页码显示 -->
            <div class="flex items-center space-x-1">
              <button
                v-for="page in visiblePages"
                :key="page"
                class="px-3 py-1 text-sm rounded-md transition-colors"
                :class="page === pagination.page 
                  ? 'bg-primary-500 text-white' 
                  : 'bg-dark-700 text-gray-300 hover:bg-dark-600'"
                @click="goToPage(page)"
                :disabled="loading"
              >
                {{ page }}
              </button>
            </div>
            
            <button 
              class="btn-secondary text-sm px-3 py-1"
              :disabled="!pagination.has_next || loading"
              @click="goToPage(pagination.page + 1)"
            >
              下一页
              <ChevronRight class="w-4 h-4 ml-1" />
            </button>
          </div>
        </div>
      </div>
      
      <!-- 表格为空时的提示 -->
      <div v-if="!loading && (!sortedProblems || sortedProblems.length === 0)" class="text-center py-12">
        <AlertTriangle class="w-16 h-16 text-gray-600 mx-auto mb-4" />
        <h3 class="text-lg font-semibold text-gray-400 mb-2">
          {{ selectedCluster ? `集群"${selectedCluster}"暂无问题Pod` : '暂无问题Pod' }}
        </h3>
        <p class="text-gray-500">
          {{ selectedCluster ? '尝试切换到其他集群或查看全部集群' : '当前集群资源配置良好，未发现问题Pod' }}
        </p>
        <button 
          v-if="selectedCluster"
          class="btn-secondary mt-4"
          @click="selectedCluster = ''; handleClusterChange()"
        >
          查看全部集群
        </button>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="!analysisData && !loading" class="text-center py-12">
      <BarChart3 class="w-16 h-16 text-gray-600 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-400 mb-2">暂无分析数据</h3>
      <p class="text-gray-500 mb-6">点击刷新按钮开始资源分析</p>
      <button 
        class="btn-primary"
        @click="refreshAllAnalysisData"
      >
        <RefreshCw class="w-4 h-4 mr-2" />
        开始分析
      </button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && !analysisData" class="text-center py-12">
      <RefreshCw class="w-16 h-16 text-primary-500 mx-auto mb-4 animate-spin" />
      <h3 class="text-lg font-semibold text-gray-400 mb-2">正在分析资源配置</h3>
      <p class="text-gray-500">请稍等，正在收集和分析集群数据...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { 
  RefreshCw, 
  Filter, 
  Lightbulb,
  BarChart3, 
  PieChart,
  AlertTriangle,
  Database,
  Cpu,
  Layers,
  ChevronLeft,
  ChevronRight
} from 'lucide-vue-next'
import { useAnalysis, useClusters } from '../composables/api'
import {
  formatCpuValue,
  formatMemoryValue,
  calculateWaste,
  getRankBadgeClass,
  getIssueBadgeClass,
  getUsageBarClass,
  getWasteClass,
  extractClusterNames
} from '../utils/analysis'

const { 
  analysis: analysisData, 
  topMemoryPods,
  topCpuPods,
  namespaceSummary,
  pagination,
  loading, 
  fetchAnalysis,
  fetchProblemsWithPagination,
  fetchTopMemoryPods,
  fetchTopCpuPods,
  fetchNamespaceSummary,
  triggerDataCollection,
  refreshAllData
} = useAnalysis()

const { clusters, fetchClusters } = useClusters()

const sortBy = ref('total_waste')
const selectedCluster = ref('')
const pageSize = ref(10)

const availableClusters = computed(() => {
  return extractClusterNames(
    clusters.value || [],
    analysisData.value?.top50_problems || [],
    topMemoryPods.value || [],
    topCpuPods.value || [],
    namespaceSummary.value || []
  )
})

// 可见的页码
const visiblePages = computed(() => {
  const current = pagination.value.page
  const total = pagination.value.total_pages
  const pages: number[] = []
  
  if (total <= 7) {
    // 总页数少于7页，显示所有页码
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    // 总页数多于7页，智能显示页码
    if (current <= 4) {
      // 当前页在前4页
      pages.push(1, 2, 3, 4, 5)
      if (total > 5) pages.push(-1, total) // -1表示省略号
    } else if (current >= total - 3) {
      // 当前页在后4页
      pages.push(1, -1)
      for (let i = total - 4; i <= total; i++) {
        pages.push(i)
      }
    } else {
      // 当前页在中间
      pages.push(1, -1)
      for (let i = current - 1; i <= current + 1; i++) {
        pages.push(i)
      }
      pages.push(-1, total)
    }
  }
  
  return pages.filter(p => p !== -1) // 暂时过滤掉省略号
})


const sortedProblems = computed(() => {
  return analysisData.value?.top50_problems || []
})


const refreshAllAnalysisData = async () => {
  await Promise.allSettled([
    fetchTopMemoryPods(20),
    fetchTopCpuPods(20),
    fetchNamespaceSummary()
  ])
}

const refreshProblems = async () => {
  await fetchProblemsWithPagination(
    pagination.value.page, 
    pageSize.value, 
    selectedCluster.value,
    sortBy.value
  )
  
  if (!analysisData.value) {
    await fetchAnalysis().catch(() => {})
  }
}

const goToPage = async (page: number) => {
  if (page < 1 || page > pagination.value.total_pages || loading.value) return
  await fetchProblemsWithPagination(page, pageSize.value, selectedCluster.value, sortBy.value)
}

const handlePageSizeChange = async () => {
  await fetchProblemsWithPagination(1, pageSize.value, selectedCluster.value, sortBy.value)
}

const handleClusterChange = async () => {
  await fetchProblemsWithPagination(1, pageSize.value, selectedCluster.value, sortBy.value)
}

const handleSortChange = async () => {
  await fetchProblemsWithPagination(1, pageSize.value, selectedCluster.value, sortBy.value)
}

const triggerFullCollection = async () => {
  await triggerDataCollection(true)
  await refreshAllData()
  await fetchProblemsWithPagination(1, pageSize.value, selectedCluster.value, sortBy.value)
}

const getIssueText = (issue: string) => issue

onMounted(async () => {
  await fetchClusters().catch(() => {})
  await fetchProblemsWithPagination(1, pageSize.value, selectedCluster.value, sortBy.value)
  
  await Promise.allSettled([
    fetchTopMemoryPods(20),
    fetchTopCpuPods(20),
    fetchNamespaceSummary(),
    fetchAnalysis()
  ])
})
</script>