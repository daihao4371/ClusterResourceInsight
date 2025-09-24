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

    <!-- 系统测试面板 -->
    <div v-if="showTestPanel" class="glass-card p-6 border-l-4 border-blue-500">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold">告警降噪与数据清理测试</h2>
        <button @click="toggleTestPanel" class="text-gray-400 hover:text-gray-600">
          ✕
        </button>
      </div>
      
      <!-- 数据库统计显示 -->
      <div v-if="databaseStats" class="mb-6 p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
        <h3 class="text-lg font-semibold mb-3">当前数据状态</h3>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div class="text-center">
            <div class="text-2xl font-bold text-blue-600">{{ databaseStats.total_activities }}</div>
            <div class="text-sm text-gray-600 dark:text-gray-400">总活动数</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-red-600">{{ databaseStats.total_alerts }}</div>
            <div class="text-sm text-gray-600 dark:text-gray-400">总告警数</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-yellow-600">{{ databaseStats.duplicate_alerts }}</div>
            <div class="text-sm text-gray-600 dark:text-gray-400">重复告警</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-green-600">{{ databaseStats.alert_status?.active || 0 }}</div>
            <div class="text-sm text-gray-600 dark:text-gray-400">活跃告警</div>
          </div>
        </div>
      </div>

      <!-- 测试操作按钮组 -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
        <!-- 获取统计信息 -->
        <div class="p-4 border rounded-lg border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20">
          <div class="flex items-center mb-2">
            <div class="w-3 h-3 bg-blue-500 rounded-full mr-2"></div>
            <h4 class="font-medium text-blue-700 dark:text-blue-300">数据库状态</h4>
          </div>
          <button @click="fetchDatabaseStats" 
                  :disabled="testLoading.stats"
                  class="w-full bg-blue-500 hover:bg-blue-600 disabled:bg-blue-400 disabled:cursor-not-allowed text-white font-medium py-2 px-4 rounded transition-colors text-sm shadow-md hover:shadow-lg">
            <div class="flex items-center justify-center">
              <svg v-if="testLoading.stats" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>{{ testLoading.stats ? '获取中...' : '📊 刷新统计' }}</span>
            </div>
          </button>
        </div>

        <!-- 告警去重 -->
        <div class="p-4 border rounded-lg border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-900/20">
          <div class="flex items-center mb-2">
            <div class="w-3 h-3 bg-green-500 rounded-full mr-2"></div>
            <h4 class="font-medium text-green-700 dark:text-green-300">告警降噪</h4>
          </div>
          <div class="text-center p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
            <div class="flex items-center justify-center text-green-600 dark:text-green-400 mb-1">
              <svg class="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path>
              </svg>
              <span class="text-sm font-medium">自动降噪</span>
            </div>
            <div class="text-xs text-green-600 dark:text-green-400">
              系统已启用自动去重功能，页面刷新时自动执行
            </div>
          </div>
          <div v-if="lastDeduplicationResult" class="mt-2 text-xs text-green-600 dark:text-green-400 bg-green-100 dark:bg-green-800/30 px-2 py-1 rounded">
            ✅ 上次去重: 删除 {{ lastDeduplicationResult.removed_count }} 条重复记录
          </div>
        </div>

        <!-- 数据清理 -->
        <div class="p-4 border rounded-lg border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20">
          <div class="flex items-center mb-2">
            <div class="w-3 h-3 bg-red-500 rounded-full mr-2"></div>
            <h4 class="font-medium text-red-700 dark:text-red-300">数据清理</h4>
          </div>
          <div class="flex items-center space-x-2 mb-2">
            <input v-model.number="retentionDays" 
                   type="number" 
                   min="0" 
                   max="365"
                   placeholder="天数"
                   class="w-16 px-2 py-1 border border-red-300 dark:border-red-600 rounded text-xs focus:ring-2 focus:ring-red-500 focus:border-red-500">
            <span class="text-xs text-red-600 dark:text-red-400">天前数据将被删除</span>
          </div>
          <button @click="showCleanupConfirm" 
                  :disabled="testLoading.cleanup"
                  class="w-full bg-red-500 hover:bg-red-600 disabled:bg-red-400 disabled:cursor-not-allowed text-white font-medium py-2 px-4 rounded transition-colors text-sm shadow-md hover:shadow-lg">
            <div class="flex items-center justify-center">
              <svg v-if="testLoading.cleanup" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>{{ testLoading.cleanup ? '清理中...' : '⚠️ 执行清理' }}</span>
            </div>
          </button>
          <div v-if="lastCleanupResult" class="mt-2 text-xs text-red-600 dark:text-red-400 bg-red-100 dark:bg-red-800/30 px-2 py-1 rounded">
            ✅ 上次清理: 活动-{{ lastCleanupResult.removed_activities }}条, 告警-{{ lastCleanupResult.removed_alerts }}条
          </div>
        </div>
      </div>

      <!-- 测试日志 -->
      <div v-if="testLogs.length > 0" class="p-3 bg-gray-100 dark:bg-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-2">
          <h4 class="font-medium text-sm">操作日志</h4>
          <button @click="clearTestLogs" class="text-xs text-blue-600 hover:text-blue-800">
            清空
          </button>
        </div>
        <div class="max-h-32 overflow-y-auto space-y-1">
          <div v-for="log in testLogs.slice(0, 10)" :key="log.timestamp" 
               class="text-xs"
               :class="getLogClass(log.level)">
            [{{ formatLogTime(log.timestamp) }}] {{ log.message }}
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

    <!-- 数据清理确认弹窗 -->
    <div v-if="showCleanupConfirmModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click="closeCleanupConfirm">
      <div class="bg-white dark:bg-gray-800 rounded-lg p-6 w-full max-w-md m-4" @click.stop>
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">确认数据清理</h3>
          <button @click="closeCleanupConfirm" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
            ✕
          </button>
        </div>
        
        <div class="mb-6">
          <div class="flex items-center p-4 mb-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
            <svg class="w-5 h-5 text-red-500 mr-3" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path>
            </svg>
            <div>
              <h4 class="text-sm font-semibold text-red-800 dark:text-red-300">⚠️ 危险操作警告</h4>
              <p class="text-sm text-red-700 dark:text-red-400 mt-1">此操作将永久删除历史数据，无法恢复！</p>
            </div>
          </div>
          
          <div class="space-y-2 text-sm text-gray-600 dark:text-gray-300">
            <p><strong>清理范围：</strong>{{ retentionDays }} 天前的所有活动和告警数据</p>
            <p><strong>预估影响：</strong>可能删除大量历史记录</p>
            <p><strong>建议：</strong>请确保已备份重要数据</p>
          </div>
        </div>
        
        <div class="flex space-x-3">
          <button 
            @click="confirmCleanup"
            :disabled="testLoading.cleanup"
            class="flex-1 bg-red-500 hover:bg-red-600 disabled:bg-red-400 disabled:cursor-not-allowed text-white font-medium py-2 px-4 rounded transition-colors">
            <div class="flex items-center justify-center">
              <svg v-if="testLoading.cleanup" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>{{ testLoading.cleanup ? '清理中...' : '确认清理' }}</span>
            </div>
          </button>
          <button 
            @click="closeCleanupConfirm"
            class="flex-1 bg-gray-300 hover:bg-gray-400 dark:bg-gray-600 dark:hover:bg-gray-500 text-gray-800 dark:text-white font-medium py-2 px-4 rounded transition-colors">
            取消
          </button>
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
const showCleanupConfirmModal = ref(false)

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
    
    // 数据刷新后自动执行去重
    await executeDeduplicationSilently()
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

// 测试面板相关状态
const showTestPanel = ref(true) // 默认显示测试面板，方便前后端联调
const databaseStats = ref(null)
const testLoading = ref({
  stats: false,
  cleanup: false
})
const testLogs = ref<Array<{level: string, message: string, timestamp: Date}>>([])
const retentionDays = ref(7) // 数据清理保留天数
const lastDeduplicationResult = ref(null)
const lastCleanupResult = ref(null)

// 测试面板方法
const toggleTestPanel = () => {
  showTestPanel.value = !showTestPanel.value
}

const fetchDatabaseStats = async () => {
  testLoading.value.stats = true
  try {
    const response = await fetch('/api/activities/database-stats')
    const result = await response.json()
    
    if (result.code === 0) {
      databaseStats.value = result.data
      addTestLog('success', `获取数据库统计成功 - 总告警: ${result.data.total_alerts}, 重复: ${result.data.duplicate_alerts}`)
    } else {
      throw new Error(result.msg)
    }
  } catch (error) {
    addTestLog('error', `获取数据库统计失败: ${error.message}`)
    console.error('获取数据库统计失败:', error)
  } finally {
    testLoading.value.stats = false
  }
}

// 显示数据清理确认弹窗
const showCleanupConfirm = () => {
  showCleanupConfirmModal.value = true
}

// 关闭数据清理确认弹窗
const closeCleanupConfirm = () => {
  showCleanupConfirmModal.value = false
}

// 确认执行数据清理
const confirmCleanup = async () => {
  await executeCleanup()
  closeCleanupConfirm()
}

// 实际执行数据清理的方法
const executeCleanup = async () => {
  testLoading.value.cleanup = true
  try {
    const response = await fetch(`/api/activities/cleanup?retention_days=${retentionDays.value}&with_stats=true`, {
      method: 'DELETE'
    })
    const result = await response.json()
    
    if (result.code === 0) {
      lastCleanupResult.value = result.data
      addTestLog('success', `数据清理完成 - 活动删除: ${result.data.removed_activities}条, 告警删除: ${result.data.removed_alerts}条`)
      // 更新数据库统计
      await fetchDatabaseStats()
      // 刷新告警列表
      await refreshAlerts()
    } else {
      throw new Error(result.msg)
    }
  } catch (error) {
    addTestLog('error', `数据清理失败: ${error.message}`)
    console.error('数据清理失败:', error)
  } finally {
    testLoading.value.cleanup = false
  }
}

const addTestLog = (level: string, message: string) => {
  testLogs.value.unshift({
    level,
    message,
    timestamp: new Date()
  })
  
  // 限制日志数量
  if (testLogs.value.length > 50) {
    testLogs.value = testLogs.value.slice(0, 50)
  }
}

const clearTestLogs = () => {
  testLogs.value = []
}

const getLogClass = (level: string) => {
  switch (level) {
    case 'success':
      return 'text-green-600'
    case 'error':
      return 'text-red-600'
    case 'warning':
      return 'text-yellow-600'
    default:
      return 'text-gray-600'
  }
}

const formatLogTime = (timestamp: Date) => {
  return timestamp.toLocaleTimeString('zh-CN', { 
    hour12: false, 
    hour: '2-digit', 
    minute: '2-digit', 
    second: '2-digit' 
  })
}

// 静默执行告警去重，不影响用户界面
const executeDeduplicationSilently = async () => {
  try {
    const response = await fetch('/api/alerts/deduplicate', {
      method: 'POST'
    })
    const result = await response.json()
    
    if (result.code === 0) {
      lastDeduplicationResult.value = result.data
      addTestLog('success', `自动去重完成 - 删除: ${result.data.removed_count}条, 保留: ${result.data.success_count}条`)
      // 更新数据库统计
      await fetchDatabaseStats()
    } else {
      throw new Error(result.msg)
    }
  } catch (error: any) {
    addTestLog('warning', `自动去重失败: ${error.message}`)
    console.warn('自动去重失败:', error)
  }
}

// 生命周期
onMounted(() => {
  refreshAlerts()
  // 初始化时获取数据库统计信息
  fetchDatabaseStats()
  // 初始化完成后自动执行一次去重
  executeDeduplicationSilently()
})
</script>