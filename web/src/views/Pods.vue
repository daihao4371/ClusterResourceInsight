<template>
  <div class="space-y-6 animate-fade-in">
    <!-- 页面标题和操作 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gradient">Pod监控</h1>
        <p class="text-gray-400 mt-1">实时监控所有Pod状态和资源使用情况</p>
      </div>
      
      <div class="flex items-center space-x-3">
        <button class="btn-secondary">
          <RefreshCw class="w-4 h-4 mr-2" />
          刷新数据
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
          <option value="Running">运行中</option>
          <option value="Pending">待调度</option>
          <option value="Failed">失败</option>
          <option value="Succeeded">成功</option>
        </select>
        
        <select v-model="namespaceFilter" class="input-field">
          <option value="">所有命名空间</option>
          <option value="default">default</option>
          <option value="kube-system">kube-system</option>
          <option value="monitoring">monitoring</option>
          <option value="production">production</option>
        </select>
        
        <select v-model="clusterFilter" class="input-field">
          <option value="">所有集群</option>
          <option value="prod-cluster-01">prod-cluster-01</option>
          <option value="dev-cluster-02">dev-cluster-02</option>
          <option value="test-cluster-03">test-cluster-03</option>
        </select>
      </div>
    </div>

    <!-- Pod列表表格 -->
    <div class="glass-card overflow-hidden">
      <div class="overflow-x-auto">
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
              v-for="pod in filteredPods"
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
                        :style="{ width: `${pod.cpuUsage}%` }"
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
                        :style="{ width: `${pod.memoryUsage}%` }"
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
              
              <!-- 操作 -->
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
                    @click="viewPodLogs(pod)"
                    class="p-1 hover:bg-white/10 rounded transition-colors"
                    title="查看日志"
                  >
                    <FileText class="w-4 h-4 text-success-400" />
                  </button>
                  <button 
                    v-if="pod.status === 'Failed'"
                    @click="restartPod(pod)"
                    class="p-1 hover:bg-white/10 rounded transition-colors"
                    title="重启Pod"
                  >
                    <RotateCw class="w-4 h-4 text-warning-400" />
                  </button>
                  <button 
                    @click="deletePod(pod)"
                    class="p-1 hover:bg-white/10 rounded transition-colors"
                    title="删除Pod"
                  >
                    <Trash2 class="w-4 h-4 text-danger-400" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      
      <!-- 分页 -->
      <div class="px-6 py-4 border-t border-gray-700 flex items-center justify-between">
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

    <!-- 空状态 -->
    <div v-if="filteredPods.length === 0" class="text-center py-12">
      <Box class="w-16 h-16 text-gray-600 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-400 mb-2">暂无Pod数据</h3>
      <p class="text-gray-500">{{ searchQuery ? '未找到匹配的Pod' : '当前没有运行的Pod' }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { 
  Box,
  Clock,
  AlertTriangle,
  Activity,
  Search,
  RefreshCw,
  Download,
  Eye,
  FileText,
  RotateCw,
  Trash2
} from 'lucide-vue-next'
import MetricCard from '../components/common/MetricCard.vue'
import { formatDistanceToNow } from '../utils/date'

interface Pod {
  id: string
  name: string
  namespace: string
  cluster: string
  status: 'Running' | 'Pending' | 'Failed' | 'Succeeded'
  cpuUsage: number
  memoryUsage: number
  restarts: number
  startTime: string
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
const podStats = ref({
  running: 245,
  pending: 12,
  failed: 8,
  avgCpuUsage: 68
})

// 模拟Pod数据
const pods = ref<Pod[]>([
  {
    id: '1',
    name: 'nginx-deployment-7d6c4f8d9b-abc12',
    namespace: 'production',
    cluster: 'prod-cluster-01',
    status: 'Running',
    cpuUsage: 45,
    memoryUsage: 62,
    restarts: 0,
    startTime: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString()
  },
  {
    id: '2',
    name: 'api-service-6b8f5c9d8e-def34',
    namespace: 'production',
    cluster: 'prod-cluster-01',
    status: 'Running',
    cpuUsage: 78,
    memoryUsage: 84,
    restarts: 2,
    startTime: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString()
  },
  {
    id: '3',
    name: 'worker-job-5c7d4e9f2a-ghi56',
    namespace: 'default',
    cluster: 'dev-cluster-02',
    status: 'Failed',
    cpuUsage: 0,
    memoryUsage: 0,
    restarts: 5,
    startTime: new Date(Date.now() - 30 * 60 * 1000).toISOString()
  }
])

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

// 总页数
const totalPages = computed(() => Math.ceil(filteredPods.value.length / pageSize.value))

// 样式方法
const getStatusIndicatorClass = (status: Pod['status']) => {
  const classes = {
    Running: 'status-online',
    Pending: 'status-warning',
    Failed: 'status-error',
    Succeeded: 'status-online'
  }
  return classes[status]
}

const getStatusBadgeClass = (status: Pod['status']) => {
  const classes = {
    Running: 'bg-success-500/20 text-success-400 border border-success-500/30',
    Pending: 'bg-warning-500/20 text-warning-400 border border-warning-500/30',
    Failed: 'bg-danger-500/20 text-danger-400 border border-danger-500/30',
    Succeeded: 'bg-success-500/20 text-success-400 border border-success-500/30'
  }
  return classes[status]
}

const getResourceBarClass = (usage: number) => {
  if (usage >= 80) return 'bg-gradient-to-r from-danger-600 to-danger-400'
  if (usage >= 60) return 'bg-gradient-to-r from-warning-600 to-warning-400'
  return 'bg-gradient-to-r from-success-600 to-success-400'
}

// 操作方法
const viewPodDetail = (pod: Pod) => {
  console.log('查看Pod详情:', pod)
}

const viewPodLogs = (pod: Pod) => {
  console.log('查看Pod日志:', pod)
}

const restartPod = (pod: Pod) => {
  console.log('重启Pod:', pod)
}

const deletePod = (pod: Pod) => {
  console.log('删除Pod:', pod)
}
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