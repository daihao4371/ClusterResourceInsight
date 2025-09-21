<template>
  <div class="space-y-6 animate-fade-in">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gradient">系统告警</h1>
        <p class="text-sm" style="color: var(--text-muted);">监控系统告警状态与历史记录</p>
      </div>
      
      <div class="flex items-center space-x-4">
        <div class="text-sm" style="color: var(--text-muted);">
          <span class="inline-block w-2 h-2 bg-warning-500 rounded-full animate-pulse mr-2"></span>
          实时监控
        </div>
        <button 
          @click="refreshAlerts"
          :disabled="loading"
          class="btn-primary text-sm"
        >
          {{ loading ? '刷新中...' : '刷新数据' }}
        </button>
      </div>
    </div>

    <!-- 告警统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <div class="glass-card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm" style="color: var(--text-muted);">活跃告警</p>
            <p class="text-2xl font-bold text-error-400">{{ alertStats.active }}</p>
          </div>
          <div class="w-12 h-12 bg-error-500/20 rounded-full flex items-center justify-center">
            <AlertTriangle class="w-6 h-6 text-error-400" />
          </div>
        </div>
      </div>

      <div class="glass-card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm" style="color: var(--text-muted);">已解决</p>
            <p class="text-2xl font-bold text-success-400">{{ alertStats.resolved }}</p>
          </div>
          <div class="w-12 h-12 bg-success-500/20 rounded-full flex items-center justify-center">
            <CheckCircle class="w-6 h-6 text-success-400" />
          </div>
        </div>
      </div>

      <div class="glass-card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm" style="color: var(--text-muted);">高级告警</p>
            <p class="text-2xl font-bold text-warning-400">{{ alertStats.high }}</p>
          </div>
          <div class="w-12 h-12 bg-warning-500/20 rounded-full flex items-center justify-center">
            <AlertCircle class="w-6 h-6 text-warning-400" />
          </div>
        </div>
      </div>

      <div class="glass-card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm" style="color: var(--text-muted);">总计</p>
            <p class="text-2xl font-bold text-primary-400">{{ alerts.length }}</p>
          </div>
          <div class="w-12 h-12 bg-primary-500/20 rounded-full flex items-center justify-center">
            <Bell class="w-6 h-6 text-primary-400" />
          </div>
        </div>
      </div>
    </div>

    <!-- 告警列表 -->
    <div class="glass-card">
      <div class="p-6 border-b" style="border-color: var(--border-color);">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-semibold">告警列表</h2>
          <div class="flex items-center space-x-4">
            <!-- 级别筛选 -->
            <select 
              v-model="levelFilter" 
              @change="filterAlerts"
              class="input-field text-sm"
            >
              <option value="">全部级别</option>
              <option value="high">高级</option>
              <option value="medium">中级</option>
              <option value="low">低级</option>
            </select>
            
            <!-- 状态筛选 -->
            <select 
              v-model="statusFilter" 
              @change="filterAlerts"
              class="input-field text-sm"
            >
              <option value="">全部状态</option>
              <option value="active">活跃</option>
              <option value="resolved">已解决</option>
              <option value="suppressed">已屏蔽</option>
            </select>
          </div>
        </div>
      </div>

      <div class="p-6">
        <div v-if="loading" class="flex items-center justify-center py-12">
          <div style="color: var(--text-muted);">加载中...</div>
        </div>
        
        <div v-else-if="filteredAlerts.length === 0" class="flex flex-col items-center justify-center py-12" style="color: var(--text-muted);">
          <Bell class="w-16 h-16 mb-4 opacity-50" />
          <p class="text-lg">暂无告警数据</p>
          <p class="text-sm">系统运行正常</p>
        </div>
        
        <div v-else class="space-y-4">
          <div 
            v-for="alert in paginatedAlerts" 
            :key="alert.id || Math.random()"
            class="bg-white/5 rounded-lg p-4 hover:bg-white/10 transition-colors"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1">
                <div class="flex items-center space-x-3 mb-2">
                  <!-- 告警级别指示器 -->
                  <div 
                    :class="[
                      'w-3 h-3 rounded-full',
                      alert.level === 'high' ? 'bg-error-500' : 
                      alert.level === 'medium' ? 'bg-warning-500' : 'bg-info-500'
                    ]"
                  ></div>
                  
                  <h3 class="font-semibold" style="color: var(--text-primary);">{{ alert.title }}</h3>
                  
                  <!-- 状态标签 -->
                  <span 
                    class="px-2 py-1 rounded-full text-xs font-medium"
                    :style="{
                      backgroundColor: alert.status === 'active' ? 'rgba(239, 68, 68, 0.2)' :
                                     alert.status === 'resolved' ? 'rgba(16, 185, 129, 0.2)' :
                                     'rgba(107, 114, 128, 0.2)',
                      color: alert.status === 'active' ? 'var(--error-color)' :
                             alert.status === 'resolved' ? 'var(--success-color)' :
                             'var(--text-muted)'
                    }"
                  >
                    {{ statusMap[alert.status] || alert.status }}
                  </span>
                </div>
                
                <p style="color: var(--text-secondary);" class="mb-2">{{ alert.description }}</p>
                
                <div class="flex items-center space-x-4 text-sm" style="color: var(--text-muted);">
                  <span>{{ alert.time }}</span>
                  <span>级别: {{ levelMap[alert.level] || alert.level }}</span>
                </div>
              </div>
              
              <!-- 操作按钮 -->
              <div class="flex items-center space-x-2 ml-4">
                <button 
                  v-if="alert.status === 'active'"
                  @click="resolveAlert(alert)"
                  class="btn-secondary text-xs"
                >
                  标记已解决
                </button>
                <button 
                  @click="viewAlertDetails(alert)"
                  class="btn-ghost text-xs"
                >
                  查看详情
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 分页控件 -->
        <div v-if="filteredAlerts.length > pageSize" class="flex items-center justify-between mt-6 pt-6 border-t" style="border-color: var(--border-color);">
          <div class="text-sm" style="color: var(--text-muted);">
            显示 {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, filteredAlerts.length) }} 
            条，共 {{ filteredAlerts.length }} 条
          </div>
          
          <div class="flex items-center space-x-2">
            <button 
              @click="currentPage--"
              :disabled="currentPage === 1"
              class="btn-ghost text-sm"
            >
              上一页
            </button>
            
            <span class="text-sm" style="color: var(--text-muted);">
              {{ currentPage }} / {{ totalPages }}
            </span>
            
            <button 
              @click="currentPage++"
              :disabled="currentPage === totalPages"
              class="btn-ghost text-sm"
            >
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 告警详情模态框 -->
    <AlertDetailsModal
      :visible="showDetailsModal"
      :alert="selectedAlert"
      @close="closeDetailsModal"
      @resolve="handleResolveFromModal"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { AlertTriangle, CheckCircle, AlertCircle, Bell } from 'lucide-vue-next'
import { useSystemStore } from '../stores/system'
import AlertDetailsModal from '../components/common/AlertDetailsModal.vue'

const systemStore = useSystemStore()

// 响应式数据
const loading = ref(false)
const alerts = ref([])
const levelFilter = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(10)

// 模态框相关状态
const showDetailsModal = ref(false)
const selectedAlert = ref(null)

// 状态和级别映射
const statusMap = {
  active: '活跃',
  resolved: '已解决',
  suppressed: '已屏蔽'
}

const levelMap = {
  high: '高级',
  medium: '中级',
  low: '低级'
}

// 计算属性
const alertStats = computed(() => {
  const stats = {
    active: 0,
    resolved: 0,
    high: 0
  }
  
  alerts.value.forEach(alert => {
    if (alert.status === 'active') stats.active++
    if (alert.status === 'resolved') stats.resolved++
    if (alert.level === 'high') stats.high++
  })
  
  return stats
})

const filteredAlerts = computed(() => {
  let filtered = alerts.value
  
  if (levelFilter.value) {
    filtered = filtered.filter(alert => alert.level === levelFilter.value)
  }
  
  if (statusFilter.value) {
    filtered = filtered.filter(alert => alert.status === statusFilter.value)
  }
  
  return filtered
})

const totalPages = computed(() => {
  return Math.ceil(filteredAlerts.value.length / pageSize.value)
})

const paginatedAlerts = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredAlerts.value.slice(start, end)
})

// 方法
const refreshAlerts = async () => {
  loading.value = true
  try {
    // 获取告警数据
    await systemStore.fetchSystemAlerts()
    alerts.value = systemStore.systemAlerts
  } catch (error) {
    console.error('获取告警数据失败:', error)
  } finally {
    loading.value = false
  }
}

const filterAlerts = () => {
  currentPage.value = 1
}

const resolveAlert = async (alert) => {
  try {
    // 调用API来解决告警
    const success = await systemStore.resolveAlert(alert.id)
    if (success) {
      // 更新本地状态
      alert.status = 'resolved'
      
      // 如果模态框打开且是同一个告警，也要更新
      if (selectedAlert.value && selectedAlert.value.id === alert.id) {
        selectedAlert.value.status = 'resolved'
      }
      
      // 可以添加成功提示
      console.log('告警已标记为已解决')
    } else {
      console.error('解决告警失败')
    }
  } catch (error) {
    console.error('解决告警失败:', error)
  }
}

const viewAlertDetails = (alert) => {
  // 扩展告警数据以包含更多详细信息
  const enhancedAlert = {
    ...alert,
    clusterId: extractClusterIdFromDescription(alert.description),
    clusterName: extractClusterNameFromDescription(alert.description),
    source: 'system', // 默认来源，实际应该从API获取
    details: alert.details || generateDefaultDetails(alert),
    resolvedTime: alert.status === 'resolved' ? '刚刚' : null
  }
  
  selectedAlert.value = enhancedAlert
  showDetailsModal.value = true
}

// 关闭详情模态框
const closeDetailsModal = () => {
  showDetailsModal.value = false
  selectedAlert.value = null
}

// 从模态框解决告警
const handleResolveFromModal = (alert) => {
  resolveAlert(alert)
  closeDetailsModal()
}

// 辅助函数：从描述中提取集群名称
const extractClusterNameFromDescription = (description: string) => {
  const match = description.match(/\[(.*?)\]/)
  return match ? match[1] : null
}

// 辅助函数：从描述中提取集群ID（模拟）
const extractClusterIdFromDescription = (description: string) => {
  // 这里应该根据实际的数据结构来实现
  const clusterName = extractClusterNameFromDescription(description)
  if (clusterName === 'dao-cloud') return 3
  if (clusterName === 'orbstack') return 1
  return null
}

// 辅助函数：生成默认详情信息
const generateDefaultDetails = (alert) => {
  return {
    alertId: Math.random().toString(36).substr(2, 9),
    triggeredBy: '系统监控',
    affectedResources: alert.description.includes('Pod') ? ['Pod'] : ['集群'],
    severity: alert.level,
    category: alert.description.includes('连接') ? '连接问题' : '资源问题'
  }
}

// 生命周期
onMounted(() => {
  refreshAlerts()
})
</script>