<template>
  <div class="cluster-selector">
    <div class="flex items-center space-x-3">
      <div class="flex items-center space-x-2">
        <Server class="w-4 h-4 text-gray-400" />
        <span class="text-sm text-gray-400 font-medium">集群:</span>
      </div>
      
      <div class="relative">
        <select 
          v-model="selectedCluster" 
          :disabled="loading"
          class="cluster-select"
          :class="{ 'loading': loading }"
        >
          <option value="">全部集群</option>
          <option 
            v-for="cluster in clusters" 
            :key="cluster.id || cluster.name" 
            :value="String(cluster.id || cluster.name)"
          >
            {{ cluster.name }}
            <span v-if="cluster.status" class="status-indicator">
              ({{ formatClusterStatus(cluster.status) }})
            </span>
          </option>
        </select>
        
        <!-- 加载指示器 -->
        <div v-if="loading" class="absolute right-3 top-1/2 transform -translate-y-1/2">
          <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary-500"></div>
        </div>
      </div>
      
      <!-- 集群状态指示器和统计信息 -->
      <div v-if="selectedClusterInfo" class="flex items-center space-x-3">
        <div class="flex items-center space-x-2">
          <div 
            class="w-2 h-2 rounded-full"
            :class="getStatusIndicatorClass(selectedClusterInfo.status)"
          ></div>
          <span class="text-xs text-gray-400">
            {{ formatClusterStatus(selectedClusterInfo.status) }}
          </span>
        </div>
        
        <!-- 集群统计信息（如果有的话） -->
        <div v-if="selectedClusterInfo.pods_count !== undefined" class="text-xs text-gray-500">
          {{ selectedClusterInfo.pods_count }} Pods
        </div>
      </div>
      
      <!-- 全部集群模式显示总数 -->
      <div v-else-if="!selectedCluster && clusters.length > 0" class="text-xs text-gray-500">
        共 {{ clusters.length }} 个集群
      </div>
      
      <!-- 刷新按钮 -->
      <button 
        @click="handleRefresh"
        :disabled="loading"
        class="refresh-btn"
        title="刷新集群列表"
      >
        <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
      </button>
    </div>
    
    <!-- 错误提示 -->
    <div v-if="error" class="mt-2 text-xs text-danger-400 flex items-center space-x-1">
      <AlertCircle class="w-3 h-3" />
      <span>{{ error }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Server, RefreshCw, AlertCircle } from 'lucide-vue-next'

// 集群接口定义
interface Cluster {
  id?: string | number
  name: string
  status?: 'online' | 'offline' | 'unknown'
  description?: string
  nodes_count?: number
  pods_count?: number
}

// Props
interface Props {
  modelValue?: string  // 当前选中的集群ID
  autoLoad?: boolean   // 是否自动加载集群列表
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  autoLoad: true
})

// Emits
const emit = defineEmits<{
  'update:modelValue': [value: string]
  'cluster-change': [cluster: string, clusterInfo: Cluster | null]
  'refresh': []
}>()

// 响应式状态
const clusters = ref<Cluster[]>([])
const selectedCluster = ref(props.modelValue)
const loading = ref(false)
const error = ref<string | null>(null)

// 计算当前选中集群的详细信息
const selectedClusterInfo = computed(() => {
  if (!selectedCluster.value) return null
  return clusters.value.find(cluster => 
    String(cluster.id || cluster.name) === selectedCluster.value
  ) || null
})

// 监听外部值变化
watch(() => props.modelValue, (newValue) => {
  selectedCluster.value = newValue || ''
})

// 监听内部值变化并向外发送
watch(selectedCluster, (newValue, oldValue) => {
  // 确保双向绑定正常工作
  emit('update:modelValue', newValue)
  
  // 只有当值真正发生变化时才触发集群变更事件
  if (newValue !== oldValue) {
    handleClusterChange()
  }
})

// 加载集群列表
const loadClusters = async () => {
  try {
    loading.value = true
    error.value = null
    
    const response = await fetch('/api/clusters')
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }
    
    const data = await response.json()
    
    // 处理后端统一响应格式
    if (data.code === 0 && data.data) {
      // 适配不同的数据结构 - 可能是 data.data 或 data.data.data
      let clusterList = data.data
      if (data.data.data && Array.isArray(data.data.data)) {
        clusterList = data.data.data
      } else if (Array.isArray(data.data)) {
        clusterList = data.data
      }
      
      // 转换为统一格式，确保有正确的id字段
      clusters.value = clusterList.map((cluster: any) => ({
        id: cluster.id || cluster.cluster_id, // 支持不同的ID字段名
        name: cluster.name || cluster.cluster_name, // 支持不同的名称字段名
        status: cluster.status || 'unknown',
        description: cluster.description || cluster.cluster_alias,
        nodes_count: cluster.nodes_count || cluster.node_count || 0,
        pods_count: cluster.pods_count || cluster.pod_count || 0
      }))
      
      console.log('集群列表加载成功:', clusters.value.length, '个集群', clusters.value)
    } else {
      throw new Error(data.message || '获取集群列表失败')
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载集群列表失败'
    console.error('加载集群列表失败:', err)
    
    // 提供默认的备用数据
    clusters.value = []
  } finally {
    loading.value = false
  }
}

// 处理集群切换 - 增强版本，防重复触发
const handleClusterChange = () => {
  const clusterInfo = selectedClusterInfo.value
  
  // 发送集群变更事件给父组件
  emit('cluster-change', selectedCluster.value, clusterInfo)
  
  console.log('集群切换事件:', {
    选中集群ID: selectedCluster.value,
    显示名称: selectedCluster.value || '全部集群',
    集群详情: clusterInfo
  })
}

// 处理刷新
const handleRefresh = () => {
  emit('refresh')
  loadClusters()
}

// 格式化集群状态显示
const formatClusterStatus = (status?: string): string => {
  const statusMap: Record<string, string> = {
    'online': '在线',
    'offline': '离线',
    'unknown': '未知'
  }
  return statusMap[status || 'unknown'] || '未知'
}

// 获取状态指示器样式
const getStatusIndicatorClass = (status?: string): string => {
  const classMap: Record<string, string> = {
    'online': 'bg-success-500',
    'offline': 'bg-danger-500',
    'unknown': 'bg-gray-500'
  }
  return classMap[status || 'unknown'] || 'bg-gray-500'
}

// 初始化
onMounted(async () => {
  // 首先加载集群列表
  if (props.autoLoad) {
    await loadClusters()
  }
  
  // 确保初始状态正确设置
  if (props.modelValue !== undefined) {
    selectedCluster.value = props.modelValue
  }
  
  // 如果没有设置初始值，确保触发一次全部集群的change事件
  if (!selectedCluster.value) {
    console.log('初始化：设置为全部集群模式')
    handleClusterChange()
  }
})

// 暴露方法供父组件调用
defineExpose({
  loadClusters,
  selectedClusterInfo
})
</script>

<style scoped>
.cluster-selector {
  @apply flex flex-col;
}

.cluster-select {
  @apply input-field min-w-48 text-sm pr-10;
  @apply transition-all duration-200;
}

.cluster-select:disabled {
  @apply opacity-60 cursor-not-allowed;
}

.cluster-select.loading {
  @apply pr-12; /* 为加载指示器留出空间 */
}

.refresh-btn {
  @apply p-2 rounded-lg transition-colors;
  @apply hover:bg-white/10 disabled:opacity-50 disabled:cursor-not-allowed;
  @apply text-gray-400 hover:text-gray-300;
}

.status-indicator {
  @apply text-xs opacity-75;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .cluster-select {
    @apply min-w-40;
  }
}

/* 选择器选项样式增强 */
.cluster-select option {
  @apply bg-dark-800 text-white;
}

/* 焦点状态 */
.cluster-select:focus {
  @apply ring-2 ring-primary-500/30 border-primary-500;
}
</style>