<template>
  <div class="space-y-6 animate-fade-in">
    <!-- 页面标题和操作 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gradient">集群管理</h1>
        <p class="text-gray-400 mt-1">管理和监控所有Kubernetes集群</p>
      </div>
      
      <div class="flex items-center space-x-3">
        <button @click="refreshAllClusters" :disabled="loading" class="btn-secondary">
          <RefreshCw class="w-4 h-4 mr-2" :class="{ 'animate-spin': loading }" />
          {{ loading ? '加载中...' : '刷新数据' }}
        </button>
        <button @click="addNewCluster" class="btn-primary">
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

    <!-- 错误提示 -->
    <div v-if="error" class="glass-card p-4 border-danger-500/50 bg-danger-500/10">
      <div class="flex items-center space-x-3">
        <div class="w-6 h-6 rounded-full bg-danger-500/20 flex items-center justify-center">
          <ExternalLink class="w-4 h-4 text-danger-400" />
        </div>
        <div>
          <p class="text-danger-400 font-medium">加载集群数据失败</p>
          <p class="text-sm text-gray-400 mt-1">{{ error }}</p>
        </div>
        <button @click="fetchClusters" class="btn-secondary ml-auto">
          重试
        </button>
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
            <p class="text-lg font-semibold">{{ cluster.nodes || 0 }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 mb-1">Pod数量</p>
            <p class="text-lg font-semibold">{{ cluster.pods || 0 }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 mb-1">K8s版本</p>
            <p class="text-sm font-medium text-primary-400">{{ cluster.version || '未知' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 mb-1">命名空间</p>
            <p class="text-lg font-semibold">{{ cluster.namespace_count || 0 }}</p>
          </div>
          <div class="col-span-2">
            <p class="text-xs text-gray-500 mb-1">
              CPU使用率
              <span v-if="cluster.dataSource === 'capacity'" class="text-yellow-400 text-[10px]">(仅容量)</span>
            </p>
            <div class="flex items-center space-x-2">
              <div class="flex-1 bg-dark-700 rounded-full h-2">
                <div 
                  class="h-2 rounded-full transition-all duration-1000"
                  :class="getResourceBarClass(cluster.cpuUsage)"
                  :style="{ width: `${cluster.cpuUsage || 0}%` }"
                ></div>
              </div>
              <span class="text-sm font-medium">{{ cluster.cpuUsage || 0 }}%</span>
            </div>
            <p v-if="cluster.hasRealUsage" class="text-[10px] text-gray-400 mt-1">
              {{ cluster.cpuUsedCores || 0 }} / {{ cluster.cpuTotalCores || 0 }} cores
            </p>
            <p v-else class="text-[10px] text-gray-400 mt-1">
              总计: {{ cluster.cpuTotalCores || 0 }} cores
            </p>
          </div>
          <div class="col-span-2">
            <p class="text-xs text-gray-500 mb-1">
              内存使用率
              <span v-if="cluster.dataSource === 'capacity'" class="text-yellow-400 text-[10px]">(仅容量)</span>
            </p>
            <div class="flex items-center space-x-2">
              <div class="flex-1 bg-dark-700 rounded-full h-2">
                <div 
                  class="h-2 rounded-full transition-all duration-1000"
                  :class="getResourceBarClass(cluster.memoryUsage)"
                  :style="{ width: `${cluster.memoryUsage || 0}%` }"
                ></div>
              </div>
              <span class="text-sm font-medium">{{ cluster.memoryUsage || 0 }}%</span>
            </div>
            <p v-if="cluster.hasRealUsage" class="text-[10px] text-gray-400 mt-1">
              {{ formatMemory(cluster.memoryUsedGB) }} / {{ formatMemory(cluster.memoryTotalGB) }}
            </p>
            <p v-else class="text-[10px] text-gray-400 mt-1">
              总计: {{ formatMemory(cluster.memoryTotalGB) }}
            </p>
          </div>
        </div>

        <!-- 集群标签 -->
        <div class="flex flex-wrap gap-2 mb-4">
          <span 
            v-for="tag in (cluster.tags || [])"
            :key="tag"
            class="px-2 py-1 text-xs bg-primary-500/20 text-primary-400 rounded-full border border-primary-500/30"
          >
            {{ tag }}
          </span>
          <!-- 如果没有标签，显示认证类型 -->
          <span 
            v-if="!cluster.tags?.length && cluster.auth_type"
            class="px-2 py-1 text-xs bg-gray-500/20 text-gray-400 rounded-full border border-gray-500/30"
          >
            {{ cluster.auth_type }}
          </span>
        </div>

        <!-- 最后更新时间 -->
        <div class="flex items-center justify-between text-sm">
          <span class="text-gray-500">
            更新: {{ formatLastUpdate(cluster) }}
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
import { ref, computed, onMounted } from 'vue'
import { 
  Server, 
  Plus, 
  Search, 
  RefreshCw, 
  MoreVertical, 
  ExternalLink 
} from 'lucide-vue-next'
import { formatDistanceToNow } from '../utils/date'
import { getClustersWithStats, testCluster } from '../api/clusters'
import type { Cluster } from '../types'

// 搜索和筛选状态
const searchQuery = ref('')
const statusFilter = ref('')
const regionFilter = ref('')

// 数据状态
const clusters = ref<Cluster[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

// 获取集群数据
const fetchClusters = async () => {
  try {
    loading.value = true
    error.value = null
    console.log('开始获取集群数据...')
    // 使用带统计数据的API获取集群信息
    clusters.value = await getClustersWithStats()
    console.log('获取到的集群数据:', clusters.value)
    
    // 输出每个集群的关键信息
    clusters.value.forEach(cluster => {
      console.log(`集群 ${cluster.name}: nodes=${cluster.nodes}, pods=${cluster.pods}, status=${cluster.status}`)
    })
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取集群数据失败'
    console.error('获取集群数据失败:', err)
  } finally {
    loading.value = false
  }
}

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
    error: 'border-danger-500/30 hover:border-danger-400/50',
    unknown: 'border-warning-500/30 hover:border-warning-400/50'
  }
  return classes[status]
}

const getStatusIndicatorClass = (status: Cluster['status']) => {
  const classes = {
    online: 'status-online',
    offline: 'status-offline',
    error: 'status-error',
    unknown: 'status-warning'
  }
  return classes[status]
}

const getResourceBarClass = (usage?: number) => {
  if (!usage) return 'bg-gray-500'
  if (usage >= 80) return 'bg-gradient-to-r from-danger-600 to-danger-400'
  if (usage >= 60) return 'bg-gradient-to-r from-warning-600 to-warning-400'
  return 'bg-gradient-to-r from-success-600 to-success-400'
}

// 操作方法
const viewClusterDetail = (cluster: Cluster) => {
  console.log('查看集群详情:', cluster)
  // TODO: 实现集群详情页面跳转
}

const refreshCluster = async (cluster: Cluster) => {
  try {
    console.log('刷新集群数据:', cluster)
    // 可以调用测试接口来刷新单个集群状态
    if (cluster.id) {
      await testCluster(parseInt(cluster.id))
      // 重新获取所有集群数据
      await fetchClusters()
    }
  } catch (err) {
    console.error('刷新集群失败:', err)
  }
}

// 刷新所有集群数据
const refreshAllClusters = async () => {
  await fetchClusters()
}

// 添加集群
const addNewCluster = () => {
  console.log('添加新集群')
  // TODO: 实现添加集群弹窗
}

// 格式化最后更新时间
const formatLastUpdate = (cluster: Cluster) => {
  const updateTime = cluster.last_collect_at || cluster.updated_at
  return updateTime ? formatDistanceToNow(updateTime) : '未知'
}

// 格式化内存显示
const formatMemory = (memoryGB?: number) => {
  if (!memoryGB) return '0 GB'
  if (memoryGB >= 1024) {
    return `${(memoryGB / 1024).toFixed(1)} TB`
  }
  return `${memoryGB.toFixed(1)} GB`
}

// 组件挂载时获取数据
onMounted(() => {
  fetchClusters()
})
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