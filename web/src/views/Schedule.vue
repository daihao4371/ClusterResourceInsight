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
        <button @click="showSettingsModal = true" class="btn-primary">
          <Settings class="w-4 h-4 mr-2" />
          全局设置
        </button>
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
        <h2 class="text-xl font-semibold">调度任务列表</h2>
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
                    <span class="text-success-400">{{ job.successful_runs }}</span>
                  </div>
                  <div class="flex justify-between text-xs">
                    <span class="text-gray-400">错误:</span>
                    <span class="text-danger-400">{{ job.error_count }}</span>
                  </div>
                  <div class="w-full bg-dark-700 rounded-full h-1.5 mt-2">
                    <div 
                      class="h-1.5 rounded-full bg-gradient-to-r from-success-600 to-success-400"
                      :style="{ width: `${getSuccessPercentage(job)}%` }"
                    ></div>
                  </div>
                </div>
              </td>
              
              <!-- 操作 -->
              <td class="px-6 py-4">
                <div class="flex items-center space-x-2">
                  <button
                    v-if="job.status === 'stopped'"
                    @click="startJob(job.cluster_id)"
                    class="p-2 hover:bg-success-500/20 rounded-lg transition-colors"
                    title="启动任务"
                  >
                    <Play class="w-4 h-4 text-success-400" />
                  </button>
                  <button
                    v-if="job.status === 'running'"
                    @click="stopJob(job.cluster_id)"
                    class="p-2 hover:bg-warning-500/20 rounded-lg transition-colors"
                    title="停止任务"
                  >
                    <Pause class="w-4 h-4 text-warning-400" />
                  </button>
                  <button
                    @click="runJobNow(job.cluster_id)"
                    class="p-2 hover:bg-primary-500/20 rounded-lg transition-colors"
                    title="立即执行"
                  >
                    <Zap class="w-4 h-4 text-primary-400" />
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
  fetchJobs, 
  fetchSettings, 
  updateSettings,
  startJob: startJobApi,
  stopJob: stopJobApi
} = useSchedule()

const showSettingsModal = ref(false)
const settingsForm = ref<GlobalScheduleSettings>({
  enabled: true,
  default_interval: 5,
  max_concurrent_jobs: 3,
  retry_max_attempts: 3,
  retry_interval: 60,
  enable_persistence: true,
  health_check_interval: 30
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

// 格式化间隔时间
const formatInterval = (minutes: number) => {
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
    running: 'bg-success-500/20 text-success-400 border border-success-500/30',
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
  await fetchJobs()
}

const startJob = async (clusterId: number) => {
  try {
    await startJobApi(clusterId)
  } catch (err) {
    console.error('启动任务失败:', err)
  }
}

const stopJob = async (clusterId: number) => {
  try {
    await stopJobApi(clusterId)
  } catch (err) {
    console.error('停止任务失败:', err)
  }
}

const runJobNow = async (clusterId: number) => {
  console.log('立即执行任务:', clusterId)
  // 实现立即执行逻辑
}

const viewJobLogs = (job: ScheduleJobInfo) => {
  console.log('查看任务日志:', job)
  // 实现查看日志逻辑
}

const saveSettings = async () => {
  try {
    await updateSettings(settingsForm.value)
    showSettingsModal.value = false
  } catch (err) {
    console.error('保存设置失败:', err)
  }
}

onMounted(async () => {
  await fetchJobs()
  await fetchSettings()
  if (settings.value) {
    settingsForm.value = { ...settings.value }
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
</style>