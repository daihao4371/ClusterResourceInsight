<template>
  <div class="space-y-4 animate-fade-in">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gradient">历史记录</h1>
        <p class="text-gray-400 text-sm">查看集群资源使用的历史趋势</p>
      </div>
      
      <div class="flex items-center space-x-3">
        <button class="btn-secondary">
          <RefreshCw class="w-4 h-4 mr-2" />
          刷新数据
        </button>
        <button class="btn-secondary">
          <Download class="w-4 h-4 mr-2" />
          导出数据
        </button>
      </div>
    </div>

    <!-- 查询条件 -->
    <div class="glass-card p-6">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">集群</label>
          <select v-model="queryParams.cluster_id" @change="onClusterChange" class="input-field">
            <option value="">所有集群</option>
            <option v-for="cluster in clusters" :key="cluster.id" :value="cluster.id">
              {{ cluster.cluster_name }}
            </option>
          </select>
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">命名空间</label>
          <select v-model="queryParams.namespace" class="input-field" @change="onNamespaceChange">
            <option value="">所有命名空间</option>
            <option v-for="namespace in namespaces" :key="namespace" :value="namespace">
              {{ namespace }}
            </option>
          </select>
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">时间范围</label>
          <select v-model="timeRange" @change="updateTimeRange" class="input-field">
            <option value="1h">最近1小时</option>
            <option value="6h">最近6小时</option>
            <option value="24h">最近24小时</option>
            <option value="7d">最近7天</option>
            <option value="30d">最近30天</option>
            <option value="custom">自定义</option>
          </select>
        </div>
        
        <!-- 自定义时间范围 -->
        <div v-if="timeRange === 'custom'" class="md:col-span-2">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">开始时间</label>
              <input
                v-model="customStartTime"
                type="datetime-local"
                class="input-field"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">结束时间</label>
              <input
                v-model="customEndTime"
                type="datetime-local"
                class="input-field"
              />
            </div>
          </div>
        </div>
        
        <div class="flex items-end">
          <button @click="searchHistory" class="btn-primary w-full">
            <Search class="w-4 h-4 mr-2" />
            查询
          </button>
        </div>
      </div>
    </div>

    <!-- 历史数据表格 -->
    <div class="glass-card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-dark-800/50">
            <tr class="border-b border-gray-700">
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                时间
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                集群/命名空间
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                资源状态
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                CPU资源
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                内存资源
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                节点
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-700">
            <tr
              v-for="record in historyData"
              :key="record.id"
              class="hover:bg-white/5 transition-colors"
            >
              <!-- 时间 -->
              <td class="px-6 py-4 text-sm text-gray-300">
                {{ formatDateTime(record.collected_at) }}
              </td>
              
              <!-- 集群/命名空间信息 -->
              <td class="px-6 py-4">
                <div>
                  <div class="font-medium text-white">{{ record.cluster?.cluster_name || record.cluster_name || 'N/A' }}</div>
                  <div class="text-sm text-gray-400">
                    {{ record.namespace }}
                  </div>
                </div>
              </td>
              
              <!-- 资源状态 -->
              <td class="px-6 py-4">
                <div class="space-y-1">
                  <span class="px-2 py-1 text-xs font-medium rounded-full bg-danger-500/20 text-danger-400 border border-danger-500/30">
                    资源不合理
                  </span>
                  <div class="text-xs text-gray-500">{{ getResourceIssues(record) }}</div>
                </div>
              </td>
              
              <!-- CPU -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div class="flex justify-between">
                    <span class="text-gray-400">请求:</span>
                    <span>{{ formatCPU(record.cpu_request) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-400">限制:</span>
                    <span>{{ formatCPU(record.cpu_limit) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-400">使用:</span>
                    <span class="text-primary-400">{{ formatCPU(record.cpu_usage) }}</span>
                  </div>
                </div>
              </td>
              
              <!-- 内存 -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div class="flex justify-between">
                    <span class="text-gray-400">请求:</span>
                    <span>{{ formatMemory(record.memory_request) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-400">限制:</span>
                    <span>{{ formatMemory(record.memory_limit) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-400">使用:</span>
                    <span class="text-primary-400">{{ formatMemory(record.memory_usage) }}</span>
                  </div>
                </div>
              </td>
              
              <!-- 节点 -->
              <td class="px-6 py-4 text-sm text-gray-300">
                {{ record.node_name }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      
      <!-- 分页 -->
      <div class="px-6 py-4 border-t border-gray-700 flex items-center justify-between">
        <div class="text-sm text-gray-400">
          显示 {{ (queryParams.page - 1) * queryParams.size + 1 }} - 
          {{ Math.min(queryParams.page * queryParams.size, total) }} 
          共 {{ total }} 条记录
        </div>
        <div class="flex items-center space-x-2">
          <button 
            @click="previousPage"
            :disabled="queryParams.page === 1"
            class="btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            上一页
          </button>
          <span class="text-sm text-gray-400">
            {{ queryParams.page }} / {{ totalPages }}
          </span>
          <button 
            @click="nextPage"
            :disabled="queryParams.page === totalPages"
            class="btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="historyData.length === 0 && !loading" class="text-center py-12">
      <Clock class="w-16 h-16 text-gray-600 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-400 mb-2">暂无历史数据</h3>
      <p class="text-gray-500">调整查询条件或等待数据收集</p>
    </div>
    
    <!-- 加载状态 -->
    <div v-if="loading" class="text-center py-12">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500 mx-auto mb-4"></div>
      <p class="text-gray-400">正在加载历史数据...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { 
  RefreshCw, 
  Download, 
  Search, 
  Clock 
} from 'lucide-vue-next'
import { useHistory, useClusters, useNamespaces } from '../composables/api'
import { formatDateTime, getTimeRange } from '../utils/date'
import type { HistoryQueryRequest } from '../types'

const { history: historyData, total, loading, fetchHistory } = useHistory()
const { clusters, fetchClusters } = useClusters()
const { namespaces, fetchNamespaces } = useNamespaces()

// 查询参数 - 只查询不合理的Pod
const queryParams = ref<HistoryQueryRequest>({
  page: 1,
  size: 20,
  cluster_id: undefined,
  namespace: '',
  start_time: '',
  end_time: '',
  only_problems: true // 只查询问题Pod
})

// 时间范围
const timeRange = ref('24h')
const customStartTime = ref('')
const customEndTime = ref('')

// 计算总页数
const totalPages = computed(() => Math.ceil(total.value / queryParams.value.size))

// 更新时间范围
const updateTimeRange = () => {
  if (timeRange.value !== 'custom') {
    const hours = timeRange.value === '1h' ? 1 :
                  timeRange.value === '6h' ? 6 :
                  timeRange.value === '24h' ? 24 :
                  timeRange.value === '7d' ? 24 * 7 :
                  24 * 30
    
    const range = getTimeRange(hours)
    queryParams.value.start_time = range.start
    queryParams.value.end_time = range.end
  } else {
    queryParams.value.start_time = customStartTime.value
    queryParams.value.end_time = customEndTime.value
  }
}

// 搜索历史数据
const searchHistory = async () => {
  updateTimeRange()
  await fetchHistory(queryParams.value)
}

// 分页操作
const previousPage = () => {
  if (queryParams.value.page > 1) {
    queryParams.value.page--
    searchHistory()
  }
}

const nextPage = () => {
  if (queryParams.value.page < totalPages.value) {
    queryParams.value.page++
    searchHistory()
  }
}

// 集群变化处理 - 当集群改变时更新命名空间列表
const onClusterChange = async () => {
  // 清空当前选择的命名空间
  queryParams.value.namespace = ''
  // 重置页码
  queryParams.value.page = 1
  
  // 根据选择的集群获取对应的命名空间
  await fetchNamespaces(queryParams.value.cluster_id || undefined)
  
  // 重新查询历史数据
  await searchHistory()
}

// 命名空间变化处理
const onNamespaceChange = () => {
  // 当命名空间变化时，重新查询数据
  queryParams.value.page = 1
  searchHistory()
}

// 获取资源问题描述
const getResourceIssues = (record: any) => {
  const issues = []
  
  // 检查CPU资源问题
  if (record.cpu_req_pct > 80) {
    issues.push('CPU请求过高')
  }
  if (record.cpu_limit_pct > 90) {
    issues.push('CPU使用达上限')
  }
  if (record.cpu_request === 0) {
    issues.push('未设置CPU请求')
  }
  
  // 检查内存资源问题
  if (record.memory_req_pct > 80) {
    issues.push('内存请求过高')
  }
  if (record.memory_limit_pct > 90) {
    issues.push('内存使用达上限')
  }
  if (record.memory_request === 0) {
    issues.push('未设置内存请求')
  }
  
  return issues.length > 0 ? issues.join(', ') : '资源配置不合理'
}

// 格式化CPU值（从millicores转换为cores）
const formatCPU = (value: number | string | null | undefined): string => {
  if (value === null || value === undefined || value === '') return 'N/A'
  const numValue = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(numValue)) return 'N/A'
  
  if (numValue >= 1000) {
    return `${(numValue / 1000).toFixed(2)} cores`
  }
  return `${numValue}m`
}

// 格式化内存值（从bytes转换为MB/GB）
const formatMemory = (value: number | string | null | undefined): string => {
  if (value === null || value === undefined || value === '') return 'N/A'
  const numValue = typeof value === 'string' ? parseInt(value) : value
  if (isNaN(numValue)) return 'N/A'
  
  if (numValue >= 1024 * 1024 * 1024) {
    return `${(numValue / (1024 * 1024 * 1024)).toFixed(2)} GB`
  } else if (numValue >= 1024 * 1024) {
    return `${(numValue / (1024 * 1024)).toFixed(0)} MB`
  } else if (numValue >= 1024) {
    return `${(numValue / 1024).toFixed(0)} KB`
  }
  return `${numValue} B`
}

onMounted(async () => {
  // 先获取集群列表
  await fetchClusters()
  
  // 初始化时获取所有命名空间（不指定集群）
  await fetchNamespaces()
  
  // 设置默认时间范围
  updateTimeRange()
  
  // 执行初始查询
  await searchHistory()
})
</script>