<template>
  <div class="space-y-6">
    <!-- 筛选器区域 -->
    <div class="glass-card p-4" style="background: var(--card-bg); border: 1px solid var(--border-color);">
      <div class="flex flex-wrap gap-4 items-center">
        <div class="flex items-center space-x-2">
          <label class="text-sm font-medium" style="color: var(--text-secondary);">集群:</label>
          <select 
            v-model="selectedClusterId" 
            @change="onClusterChange"
            class="px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            style="background: var(--card-bg); color: var(--text-primary); border-color: var(--border-color);"
          >
            <option :value="null">全部集群</option>
            <option v-for="cluster in clusters" :key="cluster.id" :value="cluster.id">
              {{ cluster.cluster_name }}
            </option>
          </select>
        </div>
        
        <div class="flex items-center space-x-2">
          <label class="text-sm font-medium" style="color: var(--text-secondary);">命名空间:</label>
          <select 
            v-model="selectedNamespace" 
            @change="onNamespaceChange"
            :disabled="isLoadingNamespaces"
            class="px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            style="background: var(--card-bg); color: var(--text-primary); border-color: var(--border-color);"
          >
            <option value="">{{ isLoadingNamespaces ? '加载中...' : '全部命名空间' }}</option>
            <option v-for="ns in namespaces" :key="ns" :value="ns" :disabled="isLoadingNamespaces">
              {{ ns }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <!-- 资源分布卡片 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- CPU资源分布 -->
      <div class="glass-card hover:shadow-lg transition-all duration-300" style="background: var(--card-bg); border: 1px solid var(--border-color);">
        <div class="p-6 border-b" style="border-color: var(--border-color);">
          <div class="flex items-center space-x-3">
            <div class="p-2 bg-blue-500/10 dark:bg-blue-400/20 rounded-lg">
              <BarChart3 class="w-5 h-5 text-blue-600 dark:text-blue-400" />
            </div>
            <h3 class="text-lg font-semibold" style="color: var(--text-primary);">CPU资源分布</h3>
          </div>
        </div>
        
        <div class="p-6">
          <!-- CPU资源统计 -->
          <div class="space-y-4 mb-6">
            <div class="flex justify-between items-center">
              <span class="text-sm" style="color: var(--text-secondary);">总请求量</span>
              <span class="font-semibold" style="color: var(--text-primary);">{{ formatCpuTotal(cpuStats.totalRequest) }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-sm" style="color: var(--text-secondary);">实际使用</span>
              <span class="font-semibold" style="color: var(--text-primary);">{{ formatCpuTotal(cpuStats.totalUsage) }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-sm" style="color: var(--text-secondary);">使用率</span>
              <span class="font-semibold" style="color: var(--text-primary);">{{ cpuStats.utilizationRate }}%</span>
            </div>
          </div>
          
          <!-- CPU使用率进度条 -->
          <div class="space-y-3">
            <div class="flex justify-between text-xs" style="color: var(--text-muted);">
              <span>CPU使用率</span>
              <span>{{ cpuStats.utilizationRate }}%</span>
            </div>
            <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
              <div 
                class="bg-blue-500 dark:bg-blue-400 h-3 rounded-full transition-all duration-500 ease-out"
                :style="{ width: `${Math.min(cpuStats.utilizationRate, 100)}%` }"
              ></div>
            </div>
            <div class="text-xs" style="color: var(--text-muted);">
              <span v-if="cpuStats.utilizationRate < 50" class="text-green-600 dark:text-green-400">• 资源充足</span>
              <span v-else-if="cpuStats.utilizationRate < 80" class="text-yellow-600 dark:text-yellow-400">• 使用适中</span>
              <span v-else class="text-red-600 dark:text-red-400">• 资源紧张</span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 内存资源分布 -->
      <div class="glass-card hover:shadow-lg transition-all duration-300" style="background: var(--card-bg); border: 1px solid var(--border-color);">
        <div class="p-6 border-b" style="border-color: var(--border-color);">
          <div class="flex items-center space-x-3">
            <div class="p-2 bg-green-500/10 dark:bg-green-400/20 rounded-lg">
              <PieChart class="w-5 h-5 text-green-600 dark:text-green-400" />
            </div>
            <h3 class="text-lg font-semibold" style="color: var(--text-primary);">内存资源分布</h3>
          </div>
        </div>
        
        <div class="p-6">
          <!-- 内存资源统计 -->
          <div class="space-y-4 mb-6">
            <div class="flex justify-between items-center">
              <span class="text-sm" style="color: var(--text-secondary);">总请求量</span>
              <span class="font-semibold" style="color: var(--text-primary);">{{ formatMemoryTotal(memoryStats.totalRequest) }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-sm" style="color: var(--text-secondary);">实际使用</span>
              <span class="font-semibold" style="color: var(--text-primary);">{{ formatMemoryTotal(memoryStats.totalUsage) }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-sm" style="color: var(--text-secondary);">使用率</span>
              <span class="font-semibold" style="color: var(--text-primary);">{{ memoryStats.utilizationRate }}%</span>
            </div>
          </div>
          
          <!-- 内存使用率进度条 -->
          <div class="space-y-3">
            <div class="flex justify-between text-xs" style="color: var(--text-muted);">
              <span>内存使用率</span>
              <span>{{ memoryStats.utilizationRate }}%</span>
            </div>
            <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
              <div 
                class="bg-green-500 dark:bg-green-400 h-3 rounded-full transition-all duration-500 ease-out"
                :style="{ width: `${Math.min(memoryStats.utilizationRate, 100)}%` }"
              ></div>
            </div>
            <div class="text-xs" style="color: var(--text-muted);">
              <span v-if="memoryStats.utilizationRate < 60" class="text-green-600 dark:text-green-400">• 内存充足</span>
              <span v-else-if="memoryStats.utilizationRate < 85" class="text-yellow-600 dark:text-yellow-400">• 使用适中</span>
              <span v-else class="text-red-600 dark:text-red-400">• 内存紧张</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { BarChart3, PieChart } from 'lucide-vue-next'
import axios from 'axios'

// 类型定义
interface ResourceStats {
  totalRequest: number
  totalUsage: number
  utilizationRate: number
}

interface ResourceDistributionData {
  cpu: ResourceStats
  memory: ResourceStats
  clustersAnalyzed: number
  podsAnalyzed: number
  generatedAt: string
}

interface ClusterInfo {
  id: number
  cluster_name: string
  status: string
}

// 响应式数据
const cpuStats = ref<ResourceStats>({
  totalRequest: 0,
  totalUsage: 0,
  utilizationRate: 0
})

const memoryStats = ref<ResourceStats>({
  totalRequest: 0,
  totalUsage: 0,
  utilizationRate: 0
})

// 筛选相关数据
const clusters = ref<ClusterInfo[]>([])
const namespaces = ref<string[]>([])
const selectedClusterId = ref<number | null>(null)
const selectedNamespace = ref<string>('')
const isLoadingNamespaces = ref<boolean>(false) // 添加命名空间加载状态

// 格式化函数
const formatCpuTotal = (value: number): string => {
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1)}Core`
  }
  return `${value}m`
}

const formatMemoryTotal = (value: number): string => {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let index = 0
  let size = value
  
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index++
  }
  
  return `${size.toFixed(1)}${units[index]}`
}

// API调用函数
const fetchClusters = async () => {
  try {
    const response = await axios.get<{
      code: number
      data: { data: ClusterInfo[], count: number }
      msg?: string
    }>('/api/clusters')
    
    if (response.data.code === 0 && response.data.data?.data) {
      clusters.value = response.data.data.data.filter(cluster => cluster.status === 'online')
      console.log('集群列表获取成功:', clusters.value)
      
      // 默认选择第一个集群（如果还没有选择的话）
      if (clusters.value.length > 0 && selectedClusterId.value === null) {
        selectedClusterId.value = clusters.value[0].id
        console.log('默认选择第一个集群:', clusters.value[0])
      }
    } else {
      console.error('获取集群列表失败:', response.data.msg)
    }
  } catch (error) {
    console.error('集群列表API调用异常:', error)
  }
}

const fetchNamespaces = async (clusterId?: number | null) => {
  // 如果没有指定集群ID，使用当前选中的集群ID
  const targetClusterId = clusterId !== undefined ? clusterId : selectedClusterId.value
  
  console.log('开始获取命名空间列表，集群ID:', targetClusterId)
  isLoadingNamespaces.value = true
  
  try {
    // 构建API URL，根据集群ID添加查询参数
    let url = '/api/namespaces'
    if (targetClusterId) {
      url += `?cluster_id=${targetClusterId}`
    }
    
    console.log('命名空间API请求URL:', url)
    
    const response = await axios.get<{
      code: number
      data: { data: string[], count: number, cluster_id?: number }
      msg?: string
    }>(url)
    
    if (response.data.code === 0 && response.data.data?.data) {
      namespaces.value = response.data.data.data
      console.log(`命名空间列表获取成功 (集群${targetClusterId || '全部'}):`, {
        count: namespaces.value.length,
        namespaces: namespaces.value
      })
      
      // 如果当前选中的命名空间不在新的命名空间列表中，则重置选择
      if (selectedNamespace.value && !namespaces.value.includes(selectedNamespace.value)) {
        console.log(`当前选中的命名空间 "${selectedNamespace.value}" 不在新集群中，已重置`)
        selectedNamespace.value = ''
      }
    } else {
      console.error('获取命名空间列表失败:', response.data.msg)
      namespaces.value = []
      selectedNamespace.value = ''
    }
  } catch (error) {
    console.error('命名空间列表API调用异常:', error)
    namespaces.value = []
    selectedNamespace.value = ''
  } finally {
    isLoadingNamespaces.value = false
  }
}

const fetchResourceDistribution = async () => {
  console.log('开始获取资源分布统计数据...', {
    clusterId: selectedClusterId.value,
    namespace: selectedNamespace.value
  })
  
  try {
    // 构建查询参数
    const params = new URLSearchParams()
    if (selectedClusterId.value) {
      params.append('cluster_id', selectedClusterId.value.toString())
    }
    if (selectedNamespace.value) {
      params.append('namespace', selectedNamespace.value)
    }
    
    const url = `/api/statistics/resource-distribution${params.toString() ? '?' + params.toString() : ''}`
    console.log('API请求URL:', url)
    
    const response = await axios.get<{
      code: number
      data: ResourceDistributionData
      msg?: string
    }>(url)
    
    console.log('API响应状态:', response.status)
    console.log('API响应数据:', response.data)
    
    if (response.data.code === 0 && response.data.data) {
      const data = response.data.data
      
      console.log('解析后端返回的数据结构:', {
        cpu: data.cpu,
        memory: data.memory,
        clustersAnalyzed: data.clustersAnalyzed,
        podsAnalyzed: data.podsAnalyzed
      })
      
      // 更新CPU统计数据
      cpuStats.value = {
        totalRequest: data.cpu.totalRequest,
        totalUsage: data.cpu.totalUsage,
        utilizationRate: Math.round(data.cpu.utilizationRate)
      }
      
      // 更新内存统计数据
      memoryStats.value = {
        totalRequest: data.memory.totalRequest,
        totalUsage: data.memory.totalUsage,
        utilizationRate: Math.round(data.memory.utilizationRate)
      }
      
      console.log('资源分布数据更新成功:', {
        cpuStats: cpuStats.value,
        memoryStats: memoryStats.value,
        clustersAnalyzed: data.clustersAnalyzed,
        podsAnalyzed: data.podsAnalyzed
      })
    } else {
      console.error('后端返回错误信息:', {
        code: response.data.code,
        msg: response.data.msg,
        data: response.data.data
      })
    }
  } catch (error) {
    console.error('资源分布统计API调用异常:', {
      message: error?.message,
      status: error?.response?.status,
      statusText: error?.response?.statusText,
      data: error?.response?.data
    })
  }
}

// 事件处理函数
const onClusterChange = async () => {
  console.log('集群选择变更:', selectedClusterId.value)
  
  // 重置命名空间选择
  selectedNamespace.value = ''
  
  // 根据新选择的集群重新获取命名空间列表
  await fetchNamespaces(selectedClusterId.value)
  
  // 获取新的资源分布数据
  await fetchResourceDistribution()
}

const onNamespaceChange = async () => {
  console.log('命名空间选择变更:', selectedNamespace.value)
  await fetchResourceDistribution()
}

// 组件的生命周期和初始化
onMounted(async () => {
  console.log('组件初始化开始')
  
  // 首先获取集群列表
  await fetchClusters()
  
  // 然后根据默认选择的集群获取命名空间列表
  if (selectedClusterId.value) {
    await fetchNamespaces(selectedClusterId.value)
  } else {
    // 如果没有默认集群，获取全部命名空间
    await fetchNamespaces(null)
  }
  
  // 最后获取初始资源分布数据
  await fetchResourceDistribution()
  
  console.log('组件初始化完成')
})
</script>