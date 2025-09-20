<template>
  <div class="space-y-6 animate-fade-in">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gradient">历史记录</h1>
        <p class="text-gray-400 mt-1">查看集群资源使用的历史趋势</p>
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
          <select v-model="queryParams.cluster_id" class="input-field">
            <option value="">所有集群</option>
            <option v-for="cluster in clusters" :key="cluster.id" :value="cluster.id">
              {{ cluster.cluster_name }}
            </option>
          </select>
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">命名空间</label>
          <input
            v-model="queryParams.namespace"
            type="text"
            placeholder="输入命名空间"
            class="input-field"
          />
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Pod名称</label>
          <input
            v-model="queryParams.pod_name"
            type="text"
            placeholder="输入Pod名称"
            class="input-field"
          />
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
                Pod信息
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                状态
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                CPU
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                内存
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
              
              <!-- Pod信息 -->
              <td class="px-6 py-4">
                <div>
                  <div class="font-medium text-white">{{ record.pod_name }}</div>
                  <div class="text-sm text-gray-400">
                    {{ record.namespace }} / {{ record.cluster_name }}
                  </div>
                </div>
              </td>
              
              <!-- 状态 -->
              <td class="px-6 py-4">
                <span 
                  class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="getStatusBadgeClass(record.status)"
                >
                  {{ record.status }}
                </span>
              </td>
              
              <!-- CPU -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div class="flex justify-between">
                    <span class="text-gray-400">请求:</span>
                    <span>{{ record.cpu_request || 'N/A' }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-400">限制:</span>
                    <span>{{ record.cpu_limit || 'N/A' }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-400">使用:</span>
                    <span class="text-primary-400">{{ record.cpu_usage || 'N/A' }}</span>
                  </div>
                </div>
              </td>
              
              <!-- 内存 -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div class="flex justify-between">
                    <span class="text-gray-400">请求:</span>
                    <span>{{ record.memory_request || 'N/A' }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-400">限制:</span>
                    <span>{{ record.memory_limit || 'N/A' }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-400">使用:</span>
                    <span class="text-primary-400">{{ record.memory_usage || 'N/A' }}</span>
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
import { useHistory, useClusters } from '../composables/api'
import { formatDateTime, getTimeRange } from '../utils/date'
import type { HistoryQueryRequest } from '../types'

const { history: historyData, total, loading, fetchHistory } = useHistory()
const { clusters, fetchClusters } = useClusters()

// 查询参数
const queryParams = ref<HistoryQueryRequest>({
  page: 1,
  size: 20,
  cluster_id: undefined,
  namespace: '',
  pod_name: '',
  start_time: '',
  end_time: ''
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

// 状态徽章样式
const getStatusBadgeClass = (status: string) => {
  const statusMap: Record<string, string> = {
    'Running': 'bg-success-500/20 text-success-400 border border-success-500/30',
    'Pending': 'bg-warning-500/20 text-warning-400 border border-warning-500/30',
    'Failed': 'bg-danger-500/20 text-danger-400 border border-danger-500/30',
    'Succeeded': 'bg-success-500/20 text-success-400 border border-success-500/30'
  }
  return statusMap[status] || 'bg-gray-500/20 text-gray-400 border border-gray-500/30'
}

onMounted(async () => {
  await fetchClusters()
  updateTimeRange()
  await searchHistory()
})
</script>