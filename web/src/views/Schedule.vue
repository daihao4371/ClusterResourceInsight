<template>
  <div class="space-y-4 animate-fade-in">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gradient">调度管理</h1>
        <p class="text-gray-400 text-sm">管理集群数据收集调度任务</p>
      </div>
      
      <div class="flex items-center space-x-3">
        <button @click="refreshJobs" class="btn-secondary">
          <RefreshCw class="w-4 h-4 mr-2" />
          刷新状态
        </button>
        <button @click="toggleScheduleService" class="btn-success" :disabled="loading">
          <component :is="scheduleServiceRunning ? PauseCircle : PlayCircle" class="w-4 h-4 mr-2" />
          {{ scheduleServiceRunning ? '停止服务' : '启动服务' }}
        </button>
        <button @click="testBackendConnection" class="btn-info">
          <Zap class="w-4 h-4 mr-2" />
          测试连接
        </button>
        <button @click="showSettingsModal = true" class="btn-primary">
          <Settings class="w-4 h-4 mr-2" />
          全局设置
        </button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="glass-card bg-danger-500/10 border-danger-500/20 p-4 mb-6">
      <div class="flex items-center space-x-3">
        <AlertTriangle class="w-5 h-5 text-danger-400 flex-shrink-0" />
        <div>
          <h3 class="text-danger-400 font-medium">加载失败</h3>
          <p class="text-gray-300 text-sm mt-1">{{ error }}</p>
        </div>
        <button 
          @click="refreshJobs" 
          class="ml-auto btn-secondary text-sm py-2 px-3"
        >
          重试
        </button>
      </div>
    </div>

    <!-- 加载指示器 -->
    <div v-if="loading" class="glass-card p-6 mb-6">
      <div class="flex items-center justify-center space-x-3">
        <div class="w-6 h-6 border-2 border-primary-400 border-t-transparent rounded-full animate-spin"></div>
        <span class="text-gray-300">正在处理调度任务操作...</span>
      </div>
    </div>

    <!-- 调度概览 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <MetricCard
        title="运行中任务"
        :value="runningJobs"
        icon="PlayCircle"
        status="success"
      />
      <MetricCard
        title="停止任务"
        :value="stoppedJobs"
        icon="PauseCircle"
        status="warning"
      />
      <MetricCard
        title="错误任务"
        :value="errorJobs"
        icon="AlertTriangle"
        status="error"
      />
      <MetricCard
        title="成功率"
        :value="successRate"
        unit="%"
        icon="CheckCircle"
        :status="successRate > 90 ? 'success' : 'warning'"
      />
    </div>

    <!-- 调度任务列表 -->
    <div class="glass-card overflow-hidden">
      <div class="p-6 border-b border-gray-700">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-semibold">调度任务列表</h2>
          
          <!-- 批量操作和快捷功能 -->
          <div class="flex items-center space-x-3">
            <button 
              @click="batchStartJobs"
              class="btn-success text-sm px-3 py-2"
              title="启动所有停止的任务"
              :disabled="loading || stoppedJobs === 0"
            >
              <Play class="w-3 h-3 mr-1" />
              批量启动 ({{ stoppedJobs }})
            </button>
            
            <button 
              @click="batchStopJobs"
              class="btn-warning text-sm px-3 py-2"
              title="停止所有运行中的任务"
              :disabled="loading || runningJobs === 0"
            >
              <Pause class="w-3 h-3 mr-1" />
              批量停止 ({{ runningJobs }})
            </button>
            
            <button 
              @click="refreshTasksFromClusters"
              class="btn-primary text-sm px-3 py-2"
              title="从集群配置刷新调度任务"
              :disabled="loading"
            >
              <RefreshCw class="w-3 h-3 mr-1" />
              同步集群任务
            </button>
          </div>
        </div>
      </div>
      
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-dark-800/50">
            <tr class="border-b border-gray-700">
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                集群信息
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                状态
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                执行间隔
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                上次运行
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                下次运行
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                执行统计
              </th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">
                操作
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-700">
            <tr
              v-for="job in jobs"
              :key="job.cluster_id"
              class="hover:bg-white/5 transition-colors"
            >
              <!-- 集群信息 -->
              <td class="px-6 py-4">
                <div class="flex items-center space-x-3">
                  <div 
                    class="status-indicator"
                    :class="getClusterStatusClass(job.status)"
                  ></div>
                  <div>
                    <div class="font-medium text-white">{{ job.cluster_name }}</div>
                    <div class="text-sm text-gray-400">ID: {{ job.cluster_id }}</div>
                  </div>
                </div>
              </td>
              
              <!-- 状态 -->
              <td class="px-6 py-4">
                <span 
                  class="px-3 py-1 text-xs font-medium rounded-full flex items-center w-fit"
                  :class="getStatusBadgeClass(job.status)"
                >
                  <component :is="getStatusIcon(job.status)" class="w-3 h-3 mr-1" />
                  {{ getStatusText(job.status) }}
                </span>
              </td>
              
              <!-- 执行间隔 -->
              <td class="px-6 py-4 text-sm text-gray-300">
                {{ formatInterval(job.interval) }}
              </td>
              
              <!-- 上次运行 -->
              <td class="px-6 py-4 text-sm">
                <div class="text-gray-300">{{ formatDistanceToNow(job.last_run) }}</div>
                <div v-if="job.last_error" class="text-xs text-danger-400 mt-1">
                  错误: {{ job.last_error }}
                </div>
              </td>
              
              <!-- 下次运行 -->
              <td class="px-6 py-4 text-sm text-gray-300">
                {{ formatDistanceToNow(job.next_run) }}
              </td>
              
              <!-- 执行统计 -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div class="flex justify-between text-xs">
                    <span class="text-gray-400">总计:</span>
                    <span>{{ job.total_runs }}</span>
                  </div>
                  <div class="flex justify-between text-xs">
                    <span class="text-gray-400">成功:</span>
                    <span class="text-green-400">{{ job.successful_runs }}</span>
                  </div>
                  <div class="flex justify-between text-xs">
                    <span class="text-gray-400">错误:</span>
                    <span class="text-danger-400">{{ job.error_count }}</span>
                  </div>
                  <div class="w-full bg-dark-700 rounded-full h-1.5 mt-2">
                    <div 
                      class="h-1.5 rounded-full bg-gradient-to-r from-green-600 to-green-400"
                      :style="{ width: `${getSuccessPercentage(job)}%` }"
                    ></div>
                  </div>
                </div>
              </td>
              
              <!-- 操作 -->
              <td class="px-6 py-4">
                <div class="flex items-center space-x-1">
                  <!-- 主要控制按钮 -->
                  <button
                    v-if="job.status === 'stopped' || job.status === 'error' || job.status === 'suspended'"
                    @click="startJob(job.cluster_id)"
                    class="p-2 hover:bg-green-500/20 rounded-lg transition-colors border border-green-500/30"
                    title="启动任务"
                  >
                    <Play class="w-4 h-4 text-green-400" />
                  </button>
                  
                  <button
                    v-if="job.status === 'running'"
                    @click="stopJob(job.cluster_id)"
                    class="p-2 hover:bg-warning-500/20 rounded-lg transition-colors border border-warning-500/30"
                    title="停止任务"
                  >
                    <Pause class="w-4 h-4 text-warning-400" />
                  </button>
                  
                  <!-- 立即执行按钮 - 增强样式 -->
                  <button
                    @click="runJobNow(job.cluster_id)"
                    class="p-2 hover:bg-primary-500/20 rounded-lg transition-colors border border-primary-500/30"
                    title="立即执行"
                    :disabled="loading"
                  >
                    <Zap class="w-4 h-4 text-primary-400" />
                  </button>
                  
                  <!-- 重启按钮 - 增强样式 -->
                  <button
                    @click="restartJobAction(job.cluster_id)"
                    class="p-2 hover:bg-purple-500/20 rounded-lg transition-colors border border-purple-500/30"
                    title="重启任务"
                    :disabled="loading"
                  >
                    <RefreshCw class="w-4 h-4 text-purple-400" />
                  </button>
                  
                  <!-- 分隔线 -->
                  <div class="w-px h-6 bg-gray-600 mx-1"></div>
                  
                  <!-- 辅助操作按钮 -->
                  <button
                    @click="editJobSettings(job)"
                    class="p-2 hover:bg-blue-500/20 rounded-lg transition-colors"
                    title="编辑设置"
                  >
                    <Settings class="w-4 h-4 text-blue-400" />
                  </button>
                  
                  <button
                    @click="viewJobLogs(job)"
                    class="p-2 hover:bg-gray-500/20 rounded-lg transition-colors"
                    title="查看日志"
                  >
                    <FileText class="w-4 h-4 text-gray-400" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 全局设置模态框 -->
    <div 
      v-if="showSettingsModal"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50"
      @click="showSettingsModal = false"
    >
      <div 
        class="glass-card p-6 max-w-md w-full mx-4"
        @click.stop
      >
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold">全局调度设置</h3>
          <button 
            @click="showSettingsModal = false"
            class="p-2 hover:bg-white/10 rounded-lg transition-colors"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              启用调度系统
            </label>
            <label class="flex items-center">
              <input 
                v-model="settingsForm.enabled"
                type="checkbox" 
                class="toggle-checkbox"
              />
              <span class="ml-2 text-sm text-gray-300">
                {{ settingsForm.enabled ? '已启用' : '已禁用' }}
              </span>
            </label>
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              默认执行间隔（分钟）
            </label>
            <input
              v-model.number="settingsForm.default_interval"
              type="number"
              min="1"
              max="1440"
              class="input-field"
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              最大并发任务数
            </label>
            <input
              v-model.number="settingsForm.max_concurrent_jobs"
              type="number"
              min="1"
              max="10"
              class="input-field"
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              重试次数
            </label>
            <input
              v-model.number="settingsForm.retry_max_attempts"
              type="number"
              min="0"
              max="5"
              class="input-field"
            />
          </div>
          
          <div class="flex space-x-3 pt-4">
            <button @click="saveSettings" class="btn-primary flex-1">
              保存设置
            </button>
            <button @click="showSettingsModal = false" class="btn-secondary flex-1">
              取消
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 任务设置编辑模态框 -->
    <div 
      v-if="showJobSettingsModal && editingJob"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50"
      @click="showJobSettingsModal = false"
    >
      <div 
        class="glass-card p-6 max-w-md w-full mx-4"
        @click.stop
      >
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold">编辑任务设置</h3>
          <button 
            @click="showJobSettingsModal = false"
            class="p-2 hover:bg-white/10 rounded-lg transition-colors"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
        
        <div class="mb-4">
          <div class="text-sm text-gray-400 mb-4">
            集群: {{ editingJob.cluster_name }} (ID: {{ editingJob.cluster_id }})
          </div>
        </div>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              执行间隔（分钟）
            </label>
            <input
              v-model.number="jobSettingsForm.interval"
              type="number"
              min="1"
              max="1440"
              class="input-field"
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              启用任务
            </label>
            <label class="flex items-center">
              <input 
                v-model="jobSettingsForm.enabled"
                type="checkbox" 
                class="toggle-checkbox"
              />
              <span class="ml-2 text-sm text-gray-300">
                {{ jobSettingsForm.enabled ? '已启用' : '已禁用' }}
              </span>
            </label>
          </div>
          
          <div class="flex space-x-3 pt-4">
            <button @click="saveJobSettings" class="btn-primary flex-1">
              保存设置
            </button>
            <button @click="showJobSettingsModal = false" class="btn-secondary flex-1">
              取消
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="jobs.length === 0 && !loading" class="text-center py-12">
      <Calendar class="w-16 h-16 text-gray-600 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-400 mb-2">暂无调度任务</h3>
      <p class="text-gray-500">添加集群后将自动创建调度任务</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { 
  RefreshCw, 
  Settings, 
  PlayCircle,
  PauseCircle,
  AlertTriangle,
  CheckCircle,
  Play,
  Pause,
  Zap,
  FileText,
  X,
  Calendar
} from 'lucide-vue-next'
import MetricCard from '../components/common/MetricCard.vue'
import { useSchedule } from '../composables/api'
import { formatDistanceToNow } from '../utils/date'
import type { ScheduleJobInfo, GlobalScheduleSettings } from '../types'

const { 
  jobs, 
  settings, 
  loading, 
  error,
  fetchJobs, 
  fetchSettings, 
  updateSettings,
  updateJobSettings,
  restartJob,
  startJob: startJobApi,
  stopJob: stopJobApi,
  executeJobNow: executeJobApi,
  refreshJobsFromClusters,
  startScheduleService,
  stopScheduleService,
  getScheduleStatus
} = useSchedule()

const showSettingsModal = ref(false)
const showJobSettingsModal = ref(false)
const editingJob = ref<ScheduleJobInfo | null>(null)
const scheduleServiceStatus = ref<any>(null)

// 调度服务是否运行中
const scheduleServiceRunning = computed(() => {
  return scheduleServiceStatus.value?.service_running === true
})

const settingsForm = ref<GlobalScheduleSettings>({
  enabled: true,
  default_interval: 5,
  max_concurrent_jobs: 3,
  retry_max_attempts: 3,
  retry_interval: 60,
  enable_persistence: true,
  health_check_interval: 30
})

const jobSettingsForm = ref({
  interval: 5,
  enabled: true
})

// 统计数据
const runningJobs = computed(() => 
  jobs.value.filter((job: ScheduleJobInfo) => job.status === 'running').length
)

const stoppedJobs = computed(() => 
  jobs.value.filter((job: ScheduleJobInfo) => job.status === 'stopped').length
)

const errorJobs = computed(() => 
  jobs.value.filter((job: ScheduleJobInfo) => job.status === 'error').length
)

const successRate = computed(() => {
  const totalRuns = jobs.value.reduce((sum: number, job: ScheduleJobInfo) => sum + job.total_runs, 0)
  const successfulRuns = jobs.value.reduce((sum: number, job: ScheduleJobInfo) => sum + job.successful_runs, 0)
  return totalRuns > 0 ? Math.round((successfulRuns / totalRuns) * 100) : 0
})

// 格式化间隔时间 - 处理后端返回的nanoseconds
const formatInterval = (intervalNs: number) => {
  // Go的time.Duration以nanoseconds为单位
  const minutes = Math.round(intervalNs / 1000000000 / 60)
  if (minutes < 60) return `${minutes}分钟`
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return remainingMinutes > 0 ? `${hours}小时${remainingMinutes}分钟` : `${hours}小时`
}

// 获取成功百分比
const getSuccessPercentage = (job: ScheduleJobInfo) => {
  return job.total_runs > 0 ? Math.round((job.successful_runs / job.total_runs) * 100) : 0
}

// 样式方法
const getClusterStatusClass = (status: string) => {
  const statusMap: Record<string, string> = {
    running: 'status-online',
    stopped: 'status-warning',
    error: 'status-error',
    suspended: 'status-offline'
  }
  return statusMap[status] || 'status-offline'
}

const getStatusBadgeClass = (status: string) => {
  const statusMap: Record<string, string> = {
    running: 'bg-green-500/20 text-green-400 border border-green-500/30',
    stopped: 'bg-warning-500/20 text-warning-400 border border-warning-500/30',
    error: 'bg-danger-500/20 text-danger-400 border border-danger-500/30',
    suspended: 'bg-gray-500/20 text-gray-400 border border-gray-500/30'
  }
  return statusMap[status] || 'bg-gray-500/20 text-gray-400 border border-gray-500/30'
}

const getStatusIcon = (status: string) => {
  const iconMap: Record<string, any> = {
    running: PlayCircle,
    stopped: PauseCircle,
    error: AlertTriangle,
    suspended: PauseCircle
  }
  return iconMap[status] || PauseCircle
}

const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    running: '运行中',
    stopped: '已停止',
    error: '错误',
    suspended: '暂停'
  }
  return textMap[status] || '未知'
}

// 操作方法
const refreshJobs = async () => {
  try {
    error.value = null // 清除之前的错误状态
    await fetchJobs()
    console.log(`✓ 成功刷新调度任务列表，共 ${jobs.value.length} 个任务`)
  } catch (err) {
    console.error('刷新任务列表失败:', err)
    error.value = err instanceof Error ? err.message : '刷新失败，请稍后重试'
  }
}

const testBackendConnection = async () => {
  try {
    error.value = null
    await fetchJobs()
    await fetchSettings()
    
    const message = `✓ 后端连接测试成功！
    - 调度任务数: ${jobs.value.length}
    - 服务状态: ${settings.value ? '正常' : '未配置'}
    - 运行中任务: ${runningJobs.value}
    - 停止任务: ${stoppedJobs.value}
    - 错误任务: ${errorJobs.value}`
    
    alert(message)
    console.log('后端连接测试成功')
  } catch (err) {
    console.error('后端连接测试失败:', err)
    error.value = `后端连接测试失败: ${err instanceof Error ? err.message : '未知错误'}`
    alert(`❌ 后端连接测试失败\n错误: ${error.value}`)
  }
}

const startJob = async (clusterId: number) => {
  try {
    await startJobApi(clusterId)
    // 显示成功反馈
    // TODO: 集成通知系统显示成功消息
    console.log(`集群 ${clusterId} 的调度任务已启动`)
  } catch (err) {
    console.error('启动任务失败:', err)
    // 显示用户友好的错误信息
    error.value = `启动任务失败: ${err instanceof Error ? err.message : '未知错误'}`
  }
}

const stopJob = async (clusterId: number) => {
  try {
    await stopJobApi(clusterId)
    console.log(`集群 ${clusterId} 的调度任务已停止`)
  } catch (err) {
    console.error('停止任务失败:', err)
    error.value = `停止任务失败: ${err instanceof Error ? err.message : '未知错误'}`
  }
}

const runJobNow = async (clusterId: number) => {
  try {
    await executeJobApi(clusterId)
    // 刷新任务列表以显示最新状态
    await fetchJobs()
    console.log(`集群 ${clusterId} 的任务已触发执行`)
  } catch (err) {
    console.error('立即执行任务失败:', err)
    error.value = `执行任务失败: ${err instanceof Error ? err.message : '未知错误'}`
  }
}

const viewJobLogs = (job: ScheduleJobInfo) => {
  // 简单的日志查看实现 - 显示任务基本信息
  const logInfo = `
任务信息：
- 集群: ${job.cluster_name}
- 状态: ${getStatusText(job.status)}
- 总运行次数: ${job.total_runs}
- 成功次数: ${job.successful_runs}
- 错误次数: ${job.error_count}
- 最后错误: ${job.last_error || '无'}
- 上次运行: ${job.last_run ? formatDistanceToNow(job.last_run) : '从未运行'}
  `.trim()
  
  alert(logInfo)
}

const saveSettings = async () => {
  try {
    await updateSettings(settingsForm.value)
    showSettingsModal.value = false
    console.log('全局调度设置已保存')
    // 清除错误状态
    error.value = null
  } catch (err) {
    console.error('保存设置失败:', err)
    error.value = `保存设置失败: ${err instanceof Error ? err.message : '未知错误'}`
  }
}

const editJobSettings = (job: ScheduleJobInfo) => {
  editingJob.value = job
  // 将后端返回的nanoseconds转换为分钟
  jobSettingsForm.value = {
    interval: Math.round(job.interval / 1000000000 / 60), // 纳秒转分钟
    enabled: job.status === 'running'
  }
  showJobSettingsModal.value = true
}

const saveJobSettings = async () => {
  if (!editingJob.value) return
  
  try {
    // 参数验证
    if (jobSettingsForm.value.interval <= 0 || jobSettingsForm.value.interval > 1440) {
      error.value = '执行间隔必须在1-1440分钟范围内'
      return
    }
    
    // 将分钟转换为Go的time.Duration (nanoseconds)
    const settings = {
      interval: jobSettingsForm.value.interval * 60 * 1000000000, // 分钟转纳秒
      enabled: jobSettingsForm.value.enabled
    }
    
    await updateJobSettings(editingJob.value.cluster_id, settings)
    showJobSettingsModal.value = false
    editingJob.value = null
    error.value = null // 清除错误状态
    console.log(`集群 ${editingJob.value?.cluster_name} 的任务设置已更新`)
  } catch (err) {
    console.error('保存任务设置失败:', err)
    error.value = `保存任务设置失败: ${err instanceof Error ? err.message : '未知错误'}`
  }
}

const restartJobAction = async (clusterId: number) => {
  try {
    await restartJob(clusterId)
    error.value = null
    console.log(`集群 ${clusterId} 的任务已重启`)
  } catch (err) {
    console.error('重启任务失败:', err)
    error.value = `重启任务失败: ${err instanceof Error ? err.message : '未知错误'}`
  }
}

// 批量操作功能
const batchStartJobs = async () => {
  const stoppedJobsList = jobs.value.filter(job => 
    job.status === 'stopped' || job.status === 'error' || job.status === 'suspended'
  )
  
  if (stoppedJobsList.length === 0) {
    error.value = '没有需要启动的任务'
    return
  }

  try {
    console.log(`开始批量启动 ${stoppedJobsList.length} 个任务`)
    
    // 并发执行启动操作，但限制并发数
    const promises = stoppedJobsList.map(job => startJobApi(job.cluster_id))
    await Promise.allSettled(promises)
    
    // 刷新任务状态
    await fetchJobs()
    error.value = null
    console.log(`✓ 批量启动完成，已处理 ${stoppedJobsList.length} 个任务`)
  } catch (err) {
    console.error('批量启动失败:', err)
    error.value = `批量启动失败: ${err instanceof Error ? err.message : '未知错误'}`
  }
}

const batchStopJobs = async () => {
  const runningJobsList = jobs.value.filter(job => job.status === 'running')
  
  if (runningJobsList.length === 0) {
    error.value = '没有正在运行的任务'
    return
  }

  try {
    console.log(`开始批量停止 ${runningJobsList.length} 个任务`)
    
    // 并发执行停止操作
    const promises = runningJobsList.map(job => stopJobApi(job.cluster_id))
    await Promise.allSettled(promises)
    
    // 刷新任务状态
    await fetchJobs()
    error.value = null
    console.log(`✓ 批量停止完成，已处理 ${runningJobsList.length} 个任务`)
  } catch (err) {
    console.error('批量停止失败:', err)
    error.value = `批量停止失败: ${err instanceof Error ? err.message : '未知错误'}`
  }
}

// 从集群配置刷新调度任务 - 使用正确的后端接口
const refreshTasksFromClusters = async () => {
  try {
    console.log('正在从集群配置同步调度任务...')
    
    // 调用后端的集群任务同步接口 - 这会触发实际的任务创建
    const result = await refreshJobsFromClusters()
    
    error.value = null
    console.log(`✓ 集群任务同步完成，当前共有 ${jobs.value.length} 个调度任务`)
    
    // 显示同步结果，包含后端返回的同步信息
    const message = `集群任务同步完成！
    ${result?.sync_message || '任务已同步'}
    - 总任务数: ${jobs.value.length}
    - 运行中: ${runningJobs.value}
    - 已停止: ${stoppedJobs.value}
    - 错误任务: ${errorJobs.value}`
    
    alert(message)
  } catch (err) {
    console.error('集群任务同步失败:', err)
    error.value = `集群任务同步失败: ${err instanceof Error ? err.message : '未知错误'}`
  }
}

// 切换调度服务状态 - 启动或停止整个调度服务
const toggleScheduleService = async () => {
  try {
    if (scheduleServiceRunning.value) {
      console.log('正在停止调度服务...')
      await stopScheduleService()
      console.log('✓ 调度服务已停止')
    } else {
      console.log('正在启动调度服务...')
      await startScheduleService()
      console.log('✓ 调度服务已启动')
    }
    
    // 更新服务状态
    await refreshServiceStatus()
    error.value = null
  } catch (err) {
    console.error('切换调度服务状态失败:', err)
    error.value = `操作失败: ${err instanceof Error ? err.message : '未知错误'}`
  }
}

// 刷新调度服务状态
const refreshServiceStatus = async () => {
  try {
    scheduleServiceStatus.value = await getScheduleStatus()
    console.log('调度服务状态已更新:', scheduleServiceStatus.value)
  } catch (err) {
    console.error('获取调度服务状态失败:', err)
    // 不设置错误状态，避免影响其他功能
  }
}

onMounted(async () => {
  try {
    await fetchJobs()
    await fetchSettings()
    await refreshServiceStatus() // 获取调度服务状态
    if (settings.value) {
      settingsForm.value = { ...settings.value }
    }
  } catch (err) {
    console.error('初始化调度页面失败:', err)
    // 显示用户友好的错误提示而不是让组件崩溃
    error.value = err instanceof Error ? err.message : '页面初始化失败，请刷新重试'
  }
})
</script>

<style scoped>
.toggle-checkbox {
  @apply relative w-11 h-6 bg-gray-600 rounded-full appearance-none cursor-pointer transition-colors;
}

.toggle-checkbox:checked {
  @apply bg-primary-500;
}

.toggle-checkbox::before {
  @apply absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform content-[''];
}

.toggle-checkbox:checked::before {
  @apply transform translate-x-5;
}

/* 按钮样式增强 - 使用项目定义的颜色变量 */
.btn-success {
  background-color: rgba(16, 185, 129, 0.2);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
  transition: all 0.3s ease;
}

.btn-success:hover {
  background-color: rgba(16, 185, 129, 0.3);
}

.btn-warning {
  background-color: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
  transition: all 0.3s ease;
}

.btn-warning:hover {
  background-color: rgba(245, 158, 11, 0.3);
}

/* 加载状态按钮 */
button:disabled {
  @apply opacity-50 cursor-not-allowed;
}

/* 操作按钮悬停效果 */
.action-btn {
  @apply p-2 rounded-lg transition-all duration-200 transform hover:scale-105;
}

/* 分隔线样式 */
.separator {
  @apply w-px h-6 bg-gray-600 mx-1 opacity-50;
}
</style>