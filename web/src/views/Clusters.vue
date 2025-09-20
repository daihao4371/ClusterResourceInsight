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
              class="input-field search-input"
            />
          </div>
        </div>
        
        <select v-model="statusFilter" class="input-field">
          <option value="">所有状态</option>
          <option value="online">在线</option>
          <option value="offline">离线</option>
          <option value="error">错误</option>
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
          <div class="relative">
            <button 
              @click.stop="toggleDropdown(cluster.id)"
              class="p-2 hover:bg-white/10 rounded-lg transition-colors opacity-0 group-hover:opacity-100"
            >
              <MoreVertical class="w-4 h-4" />
            </button>
            
            <!-- 下拉菜单 -->
            <div 
              v-if="activeDropdown === cluster.id"
              class="absolute right-0 mt-2 w-48 bg-dark-800 border border-gray-700 rounded-lg shadow-lg z-10"
            >
              <button 
                @click.stop="editCluster(cluster)"
                class="w-full text-left px-4 py-2 text-sm hover:bg-dark-700 transition-colors flex items-center space-x-2"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
                <span>编辑集群</span>
              </button>
              <button 
                @click.stop="confirmDeleteCluster(cluster)"
                class="w-full text-left px-4 py-2 text-sm text-danger-400 hover:bg-dark-700 transition-colors flex items-center space-x-2"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                <span>删除集群</span>
              </button>
            </div>
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
      <button @click="addNewCluster" class="btn-primary">
        <Plus class="w-4 h-4 mr-2" />
        添加集群
      </button>
    </div>

    <!-- 添加/编辑集群弹窗 -->
    <div v-if="showClusterModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div class="glass-card w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold">{{ isEditing ? '编辑集群' : '添加集群' }}</h2>
            <button @click="closeClusterModal" class="text-gray-400 hover:text-white">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <form @submit.prevent="submitCluster" class="space-y-4">
            <!-- 基本信息 -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium mb-2">集群名称 *</label>
                <input 
                  v-model="clusterForm.cluster_name" 
                  type="text" 
                  required 
                  :disabled="isEditing"
                  class="input-field" 
                  placeholder="输入集群名称"
                />
              </div>
              <div>
                <label class="block text-sm font-medium mb-2">集群别名</label>
                <input 
                  v-model="clusterForm.cluster_alias" 
                  type="text" 
                  class="input-field" 
                  placeholder="输入集群别名"
                />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium mb-2">API Server地址 *</label>
              <input 
                v-model="clusterForm.api_server" 
                type="url" 
                required 
                class="input-field" 
                placeholder="https://k8s-api.example.com:6443"
              />
            </div>

            <!-- 认证配置 -->
            <div>
              <label class="block text-sm font-medium mb-2">认证类型 *</label>
              <select v-model="clusterForm.auth_type" required class="input-field">
                <option value="">选择认证类型</option>
                <option value="token">Bearer Token</option>
                <option value="cert">证书认证</option>
                <option value="kubeconfig">Kubeconfig</option>
              </select>
            </div>

            <!-- Token认证 -->
            <div v-if="clusterForm.auth_type === 'token'">
              <label class="block text-sm font-medium mb-2">Bearer Token *</label>
              <textarea 
                v-model="clusterForm.auth_config.bearer_token" 
                required 
                class="input-field" 
                rows="3"
                placeholder="输入Bearer Token"
              ></textarea>
            </div>

            <!-- 证书认证 -->
            <div v-if="clusterForm.auth_type === 'cert'" class="space-y-4">
              <div>
                <label class="block text-sm font-medium mb-2">客户端证书 *</label>
                <textarea 
                  v-model="clusterForm.auth_config.client_cert" 
                  required 
                  class="input-field" 
                  rows="4"
                  placeholder="-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
                ></textarea>
              </div>
              <div>
                <label class="block text-sm font-medium mb-2">客户端私钥 *</label>
                <textarea 
                  v-model="clusterForm.auth_config.client_key" 
                  required 
                  class="input-field" 
                  rows="4"
                  placeholder="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
                ></textarea>
              </div>
              <div>
                <label class="block text-sm font-medium mb-2">CA证书</label>
                <textarea 
                  v-model="clusterForm.auth_config.ca_cert" 
                  class="input-field" 
                  rows="4"
                  placeholder="-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
                ></textarea>
              </div>
            </div>

            <!-- Kubeconfig认证 -->
            <div v-if="clusterForm.auth_type === 'kubeconfig'">
              <label class="block text-sm font-medium mb-2">Kubeconfig内容 *</label>
              <textarea 
                v-model="clusterForm.auth_config.kubeconfig" 
                required 
                class="input-field" 
                rows="6"
                placeholder="apiVersion: v1\nclusters:\n..."
              ></textarea>
            </div>

            <!-- 高级配置 -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium mb-2">采集间隔（分钟）</label>
                <input 
                  v-model.number="clusterForm.collect_interval" 
                  type="number" 
                  min="1" 
                  max="1440" 
                  class="input-field" 
                  placeholder="30"
                />
              </div>
              <div>
                <label class="block text-sm font-medium mb-2">集群标签</label>
                <input 
                  v-model="tagsInput" 
                  type="text" 
                  class="input-field" 
                  placeholder="生产,东部,kubernetes（用逗号分隔）"
                />
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="flex justify-end space-x-3 pt-4">
              <button type="button" @click="closeClusterModal" class="btn-secondary">
                取消
              </button>
              <button 
                type="button" 
                @click="testClusterConnection" 
                :disabled="submitting || !canTestConnection"
                class="btn-secondary"
              >
                <RefreshCw class="w-4 h-4 mr-2" :class="{ 'animate-spin': testing }" />
                {{ testing ? '测试中...' : '测试连接' }}
              </button>
              <button 
                type="submit" 
                :disabled="submitting"
                class="btn-primary"
              >
                <RefreshCw v-if="submitting" class="w-4 h-4 mr-2 animate-spin" />
                {{ submitting ? '保存中...' : (isEditing ? '更新集群' : '创建集群') }}
              </button>
            </div>

            <!-- 测试结果 -->
            <div v-if="testResult" class="mt-4 p-4 rounded-lg" :class="testResult.success ? 'bg-success-500/10 border border-success-500/30' : 'bg-danger-500/10 border border-danger-500/30'">
              <div class="flex items-center space-x-2">
                <div class="w-4 h-4 rounded-full" :class="testResult.success ? 'bg-success-500' : 'bg-danger-500'"></div>
                <span class="font-medium" :class="testResult.success ? 'text-success-400' : 'text-danger-400'">
                  {{ testResult.success ? '连接成功' : '连接失败' }}
                </span>
              </div>
              <p class="text-sm mt-2" :class="testResult.success ? 'text-success-300' : 'text-danger-300'">
                {{ testResult.message }}
              </p>
              <div v-if="testResult.success && testResult.version" class="text-xs text-gray-400 mt-2">
                版本: {{ testResult.version }} | 节点: {{ testResult.node_count }} | 响应时间: {{ testResult.response_time_ms }}ms
              </div>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div class="glass-card w-full max-w-md mx-4">
        <div class="p-6">
          <div class="flex items-center space-x-3 mb-4">
            <div class="w-12 h-12 rounded-full bg-danger-500/20 flex items-center justify-center">
              <svg class="w-6 h-6 text-danger-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-danger-400">确认删除</h3>
              <p class="text-sm text-gray-400">此操作不可撤销</p>
            </div>
          </div>
          
          <p class="text-gray-300 mb-6">
            确定要删除集群 <span class="font-semibold text-white">{{ clusterToDelete?.name }}</span> 吗？
            这将删除所有相关的配置和历史数据。
          </p>
          
          <div class="flex justify-end space-x-3">
            <button @click="cancelDelete" class="btn-secondary">
              取消
            </button>
            <button @click="executeDelete" :disabled="deleting" class="btn-danger">
              <RefreshCw v-if="deleting" class="w-4 h-4 mr-2 animate-spin" />
              {{ deleting ? '删除中...' : '确认删除' }}
            </button>
          </div>
        </div>
      </div>
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
import { getClustersWithStats, testCluster, addCluster, updateCluster, deleteCluster, testClusterConfig } from '../api/clusters'
import type { Cluster } from '../types'

// 搜索和筛选状态
const searchQuery = ref('')
const statusFilter = ref('')

// 数据状态
const clusters = ref<Cluster[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

// UI状态
const activeDropdown = ref<string | null>(null)
const showClusterModal = ref(false)
const showDeleteModal = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const testing = ref(false)
const deleting = ref(false)

// 表单数据
const clusterForm = ref({
  cluster_name: '',
  cluster_alias: '',
  api_server: '',
  auth_type: '',
  auth_config: {
    bearer_token: '',
    client_cert: '',
    client_key: '',
    ca_cert: '',
    kubeconfig: ''
  },
  collect_interval: 30,
  tags: [] as string[]
})

const tagsInput = ref('')
const testResult = ref<any>(null)
const clusterToDelete = ref<Cluster | null>(null)
const editingClusterId = ref<number | null>(null)

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

// 计算属性
const canTestConnection = computed(() => {
  return clusterForm.value.cluster_name && 
         clusterForm.value.api_server && 
         clusterForm.value.auth_type &&
         (
           (clusterForm.value.auth_type === 'token' && clusterForm.value.auth_config.bearer_token) ||
           (clusterForm.value.auth_type === 'cert' && clusterForm.value.auth_config.client_cert && clusterForm.value.auth_config.client_key) ||
           (clusterForm.value.auth_type === 'kubeconfig' && clusterForm.value.auth_config.kubeconfig)
         )
})
const filteredClusters = computed(() => {
  return clusters.value.filter(cluster => {
    // 搜索条件匹配：集群名称或API服务器地址
    const matchesSearch = !searchQuery.value || 
      cluster.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      cluster.endpoint.toLowerCase().includes(searchQuery.value.toLowerCase())
    
    // 状态筛选匹配
    const matchesStatus = !statusFilter.value || cluster.status === statusFilter.value
    
    return matchesSearch && matchesStatus
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

// 点击外部关闭下拉菜单
const toggleDropdown = (clusterId: string) => {
  activeDropdown.value = activeDropdown.value === clusterId ? null : clusterId
}

// 重置表单
const resetForm = () => {
  clusterForm.value = {
    cluster_name: '',
    cluster_alias: '',
    api_server: '',
    auth_type: '',
    auth_config: {
      bearer_token: '',
      client_cert: '',
      client_key: '',
      ca_cert: '',
      kubeconfig: ''
    },
    collect_interval: 30,
    tags: []
  }
  tagsInput.value = ''
  testResult.value = null
}

// 关闭弹窗
const closeClusterModal = () => {
  showClusterModal.value = false
  isEditing.value = false
  editingClusterId.value = null
  resetForm()
}

// 测试集群连接
const testClusterConnection = async () => {
  if (!canTestConnection.value) return
  
  try {
    testing.value = true
    testResult.value = null
    
    // 构建测试请求数据
    const testData = {
      cluster_name: clusterForm.value.cluster_name,
      api_server: clusterForm.value.api_server,
      auth_type: clusterForm.value.auth_type,
      auth_config: {
        ...clusterForm.value.auth_config
      }
    }
    
    // 调用测试API
    const result = await testClusterConfig(testData)
    
    testResult.value = {
      success: result.success || true,
      message: result.message || '连接测试成功',
      ...result
    }
  } catch (err) {
    testResult.value = {
      success: false,
      message: err instanceof Error ? err.message : '连接测试失败'
    }
  } finally {
    testing.value = false
  }
}

// 提交表单
const submitCluster = async () => {
  try {
    submitting.value = true
    
    // 处理标签数据
    const tags = tagsInput.value 
      ? tagsInput.value.split(',').map(tag => tag.trim()).filter(tag => tag)
      : []
    
    const formData = {
      cluster_name: clusterForm.value.cluster_name,
      cluster_alias: clusterForm.value.cluster_alias || clusterForm.value.cluster_name,
      api_server: clusterForm.value.api_server,
      auth_type: clusterForm.value.auth_type,
      auth_config: clusterForm.value.auth_config,
      collect_interval: clusterForm.value.collect_interval || 30,
      tags
    }
    
    if (isEditing.value && editingClusterId.value) {
      await updateCluster(editingClusterId.value, formData)
    } else {
      await addCluster(formData)
    }
    
    // 刷新集群列表
    await fetchClusters()
    closeClusterModal()
    
  } catch (err) {
    console.error('保存集群失败:', err)
    alert(err instanceof Error ? err.message : '保存集群失败')
  } finally {
    submitting.value = false
  }
}
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
  activeDropdown.value = null
  isEditing.value = false
  resetForm()
  showClusterModal.value = true
}

// 编辑集群
const editCluster = async (cluster: Cluster) => {
  console.log('编辑集群:', cluster)
  activeDropdown.value = null
  
  try {
    // 获取集群详细信息
    const response = await fetch(`/api/clusters/${cluster.id}`)
    const result = await response.json()
    
    if (result.code === 0) {
      const clusterData = result.data.data
      
      // 填充表单数据
      clusterForm.value = {
        cluster_name: clusterData.cluster_name,
        cluster_alias: clusterData.cluster_alias || '',
        api_server: clusterData.api_server,
        auth_type: clusterData.auth_type,
        auth_config: {
          bearer_token: '',
          client_cert: '',
          client_key: '',
          ca_cert: '',
          kubeconfig: ''
        },
        collect_interval: clusterData.collect_interval || 30,
        tags: clusterData.tags ? JSON.parse(clusterData.tags) : []
      }
      
      // 设置标签输入
      tagsInput.value = clusterForm.value.tags.join(', ')
      
      isEditing.value = true
      editingClusterId.value = parseInt(cluster.id)
      showClusterModal.value = true
    }
  } catch (err) {
    console.error('获取集群详情失败:', err)
    alert('获取集群详情失败')
  }
}

// 确认删除集群
const confirmDeleteCluster = (cluster: Cluster) => {
  console.log('删除集群:', cluster)
  activeDropdown.value = null
  clusterToDelete.value = cluster
  showDeleteModal.value = true
}

// 取消删除
const cancelDelete = () => {
  showDeleteModal.value = false
  clusterToDelete.value = null
}

// 执行删除
const executeDelete = async () => {
  if (!clusterToDelete.value) return
  
  try {
    deleting.value = true
    await deleteCluster(parseInt(clusterToDelete.value.id))
    
    // 刷新集群列表
    await fetchClusters()
    showDeleteModal.value = false
    clusterToDelete.value = null
    
  } catch (err) {
    console.error('删除集群失败:', err)
    alert(err instanceof Error ? err.message : '删除集群失败')
  } finally {
    deleting.value = false
  }
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
  
  // 点击外部关闭下拉菜单
  document.addEventListener('click', () => {
    activeDropdown.value = null
  })
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

.btn-secondary {
  @apply px-4 py-2 bg-gray-600 hover:bg-gray-500 text-white rounded-lg transition-colors;
}

.btn-primary {
  @apply px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-lg transition-colors;
}

.btn-danger {
  @apply px-4 py-2 bg-danger-600 hover:bg-danger-500 text-white rounded-lg transition-colors;
}

.input-field {
  @apply w-full px-3 py-2 bg-dark-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:border-primary-500 focus:outline-none;
}

/* 搜索框专用样式，确保文本与放大镜图标不重叠 */
.search-input {
  padding-left: 2.75rem; /* 44px，为放大镜图标预留足够空间 */
}

.glass-card {
  @apply rounded-xl border border-gray-700/50 backdrop-blur-sm;
  background: rgba(17, 24, 39, 0.8);
}

/* 状态指示器样式 */
.status-online {
  @apply bg-success-500;
  box-shadow: 0 0 10px rgba(34, 197, 94, 0.5);
}

.status-offline {
  @apply bg-gray-500;
}

.status-error {
  @apply bg-danger-500;
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.5);
}

.status-warning {
  @apply bg-warning-500;
  box-shadow: 0 0 10px rgba(245, 158, 11, 0.5);
}

/* 渐变样式 */
.text-gradient {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 动画样式 */
.animate-fade-in {
  animation: fadeIn 0.6s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 响应式调整 */
@media (max-width: 768px) {
  .cluster-card:hover {
    transform: none;
  }
}
</style>