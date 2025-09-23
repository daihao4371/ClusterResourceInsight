<template>
  <div class="glass-card">
    <div class="p-6 border-b border-gray-700">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h2 class="text-xl font-semibold">问题Pod详细分析</h2>
          <p class="text-sm text-gray-400 mt-1">深入分析资源配置不合理的Pod，优化集群资源使用效率</p>
        </div>
        <div class="flex items-center space-x-3">
          <select v-model="localSortBy" @change="handleSortChange" class="input-field text-sm">
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
            v-model="localSelectedCluster" 
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
            v-model="localPageSize" 
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
          @click="$emit('refresh')"
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
            v-else-if="problems && problems.length > 0"
            v-for="(pod, index) in problems"
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
                  {{ issue }}
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
            @click="$emit('page-change', pagination.page - 1)"
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
              @click="$emit('page-change', page)"
              :disabled="loading"
            >
              {{ page }}
            </button>
          </div>
          
          <button 
            class="btn-secondary text-sm px-3 py-1"
            :disabled="!pagination.has_next || loading"
            @click="$emit('page-change', pagination.page + 1)"
          >
            下一页
            <ChevronRight class="w-4 h-4 ml-1" />
          </button>
        </div>
      </div>
    </div>
    
    <!-- 表格为空时的提示 -->
    <div v-if="!loading && (!problems || problems.length === 0)" class="text-center py-12">
      <AlertTriangle class="w-16 h-16 text-gray-600 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-400 mb-2">
        {{ localSelectedCluster ? `集群"${localSelectedCluster}"暂无问题Pod` : '暂无问题Pod' }}
      </h3>
      <p class="text-gray-500">
        {{ localSelectedCluster ? '尝试切换到其他集群或查看全部集群' : '当前集群资源配置良好，未发现问题Pod' }}
      </p>
      <button 
        v-if="localSelectedCluster"
        class="btn-secondary mt-4"
        @click="clearClusterFilter"
      >
        查看全部集群
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { 
  RefreshCw, 
  Filter, 
  Lightbulb,
  AlertTriangle,
  ChevronLeft,
  ChevronRight
} from 'lucide-vue-next'
import {
  formatCpuValue,
  formatMemoryValue,
  calculateWaste,
  getRankBadgeClass,
  getIssueBadgeClass,
  getUsageBarClass,
  getWasteClass
} from '../../utils/analysis'

// 定义Pod接口类型
interface Pod {
  cluster_name: string
  namespace: string
  pod_name: string
  node_name: string
  cpu_request: number
  memory_request: number
  cpu_usage: number
  memory_usage: number
  cpu_req_pct: number
  memory_req_pct: number
  issues: string[]
}

// 定义分页接口类型
interface Pagination {
  page: number
  size: number
  total: number
  total_pages: number
  has_prev: boolean
  has_next: boolean
}

// 定义props接口
interface Props {
  problems?: Pod[]
  pagination: Pagination
  availableClusters: string[]
  loading?: boolean
  sortBy: string
  selectedCluster: string
  pageSize: number
}

// 定义事件接口
interface Emits {
  'sort-change': [sortBy: string]
  'cluster-change': [cluster: string]
  'page-size-change': [size: number]
  'page-change': [page: number]
  'refresh': []
}

// 接收props和定义emits
const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 本地状态管理，避免直接修改props
const localSortBy = ref(props.sortBy)
const localSelectedCluster = ref(props.selectedCluster)
const localPageSize = ref(props.pageSize)

// 监听props变化，同步到本地状态
watch(() => props.sortBy, (newVal) => localSortBy.value = newVal)
watch(() => props.selectedCluster, (newVal) => localSelectedCluster.value = newVal)
watch(() => props.pageSize, (newVal) => localPageSize.value = newVal)

// 可见的页码计算
const visiblePages = computed(() => {
  const current = props.pagination.page
  const total = props.pagination.total_pages
  const pages: number[] = []
  
  if (total <= 7) {
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    if (current <= 4) {
      pages.push(1, 2, 3, 4, 5)
      if (total > 5) pages.push(-1, total)
    } else if (current >= total - 3) {
      pages.push(1, -1)
      for (let i = total - 4; i <= total; i++) {
        pages.push(i)
      }
    } else {
      pages.push(1, -1)
      for (let i = current - 1; i <= current + 1; i++) {
        pages.push(i)
      }
      pages.push(-1, total)
    }
  }
  
  return pages.filter(p => p !== -1)
})

// 事件处理函数
const handleSortChange = () => {
  emit('sort-change', localSortBy.value)
}

const handleClusterChange = () => {
  emit('cluster-change', localSelectedCluster.value)
}

const handlePageSizeChange = () => {
  emit('page-size-change', localPageSize.value)
}

const clearClusterFilter = () => {
  localSelectedCluster.value = ''
  emit('cluster-change', '')
}
</script>