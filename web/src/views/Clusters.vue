<template>
  <div class="space-y-6 animate-fade-in">
    <!-- 页面标题和操作 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gradient">集群管理</h1>
        <p class="text-gray-400 mt-1">管理和监控所有Kubernetes集群</p>
      </div>
      
      <div class="flex items-center space-x-3">
        <button class="btn-secondary">
          <RefreshCw class="w-4 h-4 mr-2" />
          刷新数据
        </button>
        <button class="btn-primary">
          <Plus class="w-4 h-4 mr-2" />
          添加集群
        </button>
      </div>
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
              placeholder="搜索集群名称、地址..."
              class="input-field pl-10"
            />
          </div>
        </div>
        
        <select v-model="statusFilter" class="input-field">
          <option value="">所有状态</option>
          <option value="online">在线</option>
          <option value="offline">离线</option>
          <option value="error">错误</option>
        </select>
        
        <select v-model="regionFilter" class="input-field">
          <option value="">所有区域</option>
          <option value="us-east">美东</option>
          <option value="us-west">美西</option>
          <option value="eu-central">欧洲</option>
          <option value="asia-pacific">亚太</option>
        </select>
      </div>
    </div>

    <!-- 集群列表 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
      <div
        v-for="cluster in filteredClusters"
        :key="cluster.id"
        class="cluster-card group"
        :class="getClusterStatusClass(cluster.status)"
        @click="viewClusterDetail(cluster)"
      >
        <!-- 集群状态指示器 -->
        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center space-x-3">
            <div 
              class="status-indicator-large"
              :class="getStatusIndicatorClass(cluster.status)"
            ></div>
            <div>
              <h3 class="text-lg font-semibold text-white">{{ cluster.name }}</h3>
              <p class="text-sm text-gray-400">{{ cluster.endpoint }}</p>
            </div>
          </div>
          
          <!-- 操作菜单 -->
          <div class="opacity-0 group-hover:opacity-100 transition-opacity">
            <button class="p-2 hover:bg-white/10 rounded-lg transition-colors">
              <MoreVertical class="w-4 h-4" />
            </button>
          </div>
        </div>

        <!-- 集群信息 -->
        <div class="grid grid-cols-2 gap-4 mb-4">
          <div>
            <p class="text-xs text-gray-500 mb-1">节点数量</p>
            <p class="text-lg font-semibold">{{ cluster.nodes }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 mb-1">Pod数量</p>
            <p class="text-lg font-semibold">{{ cluster.pods }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 mb-1">CPU使用率</p>
            <div class="flex items-center space-x-2">
              <div class="flex-1 bg-dark-700 rounded-full h-2">
                <div 
                  class="h-2 rounded-full transition-all duration-1000"
                  :class="getResourceBarClass(cluster.cpuUsage)"
                  :style="{ width: `${cluster.cpuUsage}%` }"
                ></div>
              </div>
              <span class="text-sm font-medium">{{ cluster.cpuUsage }}%</span>
            </div>
          </div>
          <div>
            <p class="text-xs text-gray-500 mb-1">内存使用率</p>
            <div class="flex items-center space-x-2">
              <div class="flex-1 bg-dark-700 rounded-full h-2">
                <div 
                  class="h-2 rounded-full transition-all duration-1000"
                  :class="getResourceBarClass(cluster.memoryUsage)"
                  :style="{ width: `${cluster.memoryUsage}%` }"
                ></div>
              </div>
              <span class="text-sm font-medium">{{ cluster.memoryUsage }}%</span>
            </div>
          </div>
        </div>

        <!-- 集群标签 -->
        <div class="flex flex-wrap gap-2 mb-4">
          <span 
            v-for="tag in cluster.tags"
            :key="tag"
            class="px-2 py-1 text-xs bg-primary-500/20 text-primary-400 rounded-full border border-primary-500/30"
          >
            {{ tag }}
          </span>
        </div>

        <!-- 最后更新时间 -->
        <div class="flex items-center justify-between text-sm">
          <span class="text-gray-500">
            更新: {{ formatDistanceToNow(cluster.lastUpdate) }}
          </span>
          <div class="flex space-x-2">
            <button 
              @click.stop="refreshCluster(cluster)"
              class="text-primary-400 hover:text-primary-300 transition-colors"
            >
              <RefreshCw class="w-4 h-4" />
            </button>
            <button 
              @click.stop="viewClusterDetail(cluster)"
              class="text-primary-400 hover:text-primary-300 transition-colors"
            >
              <ExternalLink class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="filteredClusters.length === 0" class="text-center py-12">
      <Server class="w-16 h-16 text-gray-600 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-400 mb-2">暂无集群数据</h3>
      <p class="text-gray-500 mb-6">{{ searchQuery ? '未找到匹配的集群' : '开始添加您的第一个集群' }}</p>
      <button class="btn-primary">
        <Plus class="w-4 h-4 mr-2" />
        添加集群
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { 
  Server, 
  Plus, 
  Search, 
  RefreshCw, 
  MoreVertical, 
  ExternalLink 
} from 'lucide-vue-next'
import { formatDistanceToNow } from '../utils/date'

interface Cluster {
  id: string
  name: string
  endpoint: string
  status: 'online' | 'offline' | 'error'
  region: string
  nodes: number
  pods: number
  cpuUsage: number
  memoryUsage: number
  tags: string[]
  lastUpdate: string
}

// 搜索和筛选状态
const searchQuery = ref('')
const statusFilter = ref('')
const regionFilter = ref('')

// 模拟集群数据
const clusters = ref<Cluster[]>([
  {
    id: '1',
    name: 'prod-cluster-01',
    endpoint: 'https://prod-k8s-01.example.com',
    status: 'online',
    region: 'us-east',
    nodes: 12,
    pods: 148,
    cpuUsage: 65,
    memoryUsage: 72,
    tags: ['production', 'web'],
    lastUpdate: new Date(Date.now() - 5 * 60 * 1000).toISOString()
  },
  {
    id: '2',
    name: 'dev-cluster-02',
    endpoint: 'https://dev-k8s-02.example.com',
    status: 'online',
    region: 'us-west',
    nodes: 6,
    pods: 82,
    cpuUsage: 35,
    memoryUsage: 48,
    tags: ['development', 'api'],
    lastUpdate: new Date(Date.now() - 10 * 60 * 1000).toISOString()
  },
  {
    id: '3',
    name: 'test-cluster-03',
    endpoint: 'https://test-k8s-03.example.com',
    status: 'error',
    region: 'eu-central',
    nodes: 4,
    pods: 25,
    cpuUsage: 0,
    memoryUsage: 0,
    tags: ['testing'],
    lastUpdate: new Date(Date.now() - 30 * 60 * 1000).toISOString()
  }
])

// 筛选后的集群列表
const filteredClusters = computed(() => {
  return clusters.value.filter(cluster => {
    const matchesSearch = !searchQuery.value || 
      cluster.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      cluster.endpoint.toLowerCase().includes(searchQuery.value.toLowerCase())
    
    const matchesStatus = !statusFilter.value || cluster.status === statusFilter.value
    const matchesRegion = !regionFilter.value || cluster.region === regionFilter.value
    
    return matchesSearch && matchesStatus && matchesRegion
  })
})

// 样式方法
const getClusterStatusClass = (status: Cluster['status']) => {
  const classes = {
    online: 'border-success-500/30 hover:border-success-400/50',
    offline: 'border-gray-500/30 hover:border-gray-400/50',
    error: 'border-danger-500/30 hover:border-danger-400/50'
  }
  return classes[status]
}

const getStatusIndicatorClass = (status: Cluster['status']) => {
  const classes = {
    online: 'status-online',
    offline: 'status-offline',
    error: 'status-error'
  }
  return classes[status]
}

const getResourceBarClass = (usage: number) => {
  if (usage >= 80) return 'bg-gradient-to-r from-danger-600 to-danger-400'
  if (usage >= 60) return 'bg-gradient-to-r from-warning-600 to-warning-400'
  return 'bg-gradient-to-r from-success-600 to-success-400'
}

// 操作方法
const viewClusterDetail = (cluster: Cluster) => {
  console.log('查看集群详情:', cluster)
}

const refreshCluster = (cluster: Cluster) => {
  console.log('刷新集群数据:', cluster)
}
</script>

<style scoped>
.cluster-card {
  @apply p-6 rounded-xl border backdrop-blur-sm cursor-pointer transition-all duration-300;
  background: rgba(17, 24, 39, 0.8);
}

.cluster-card:hover {
  @apply transform scale-105;
  box-shadow: 
    0 20px 25px -5px rgba(0, 0, 0, 0.3),
    0 10px 10px -5px rgba(0, 0, 0, 0.2);
}

.status-indicator-large {
  @apply w-4 h-4 rounded-full;
}
</style>