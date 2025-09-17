<template>
  <div class="schedule-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>调度任务管理</h1>
        <p class="subtitle">管理集群数据收集的定时调度任务</p>
      </div>
      <div class="header-actions">
        <el-button 
          v-if="!scheduleStatus?.service_running" 
          type="success" 
          @click="startService" 
          :loading="serviceOperating"
          :icon="VideoPlay"
        >
          启动服务
        </el-button>
        <el-button 
          v-else
          type="danger" 
          @click="stopService" 
          :loading="serviceOperating"
          :icon="VideoPause"
        >
          停止服务
        </el-button>
        <el-button @click="refreshData" :loading="loading" :icon="Refresh">
          刷新
        </el-button>
        <el-button @click="showSettings" :icon="Setting">
          全局设置
        </el-button>
      </div>
    </div>

    <!-- 服务状态概览 -->
    <el-row :gutter="20" class="status-overview">
      <el-col :span="4">
        <el-card class="status-card">
          <div class="status-content">
            <div class="status-icon" :class="scheduleStatus?.service_running ? 'running' : 'stopped'">
              <el-icon v-if="scheduleStatus?.service_running"><Check /></el-icon>
              <el-icon v-else><Close /></el-icon>
            </div>
            <div class="status-info">
              <div class="status-text">{{ scheduleStatus?.service_running ? '运行中' : '已停止' }}</div>
              <div class="status-label">服务状态</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="5">
        <el-card class="status-card">
          <div class="status-content">
            <div class="status-value">{{ scheduleStatus?.total_jobs || 0 }}</div>
            <div class="status-label">总任务数</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="5">
        <el-card class="status-card">
          <div class="status-content">
            <div class="status-value running">{{ scheduleStatus?.running_jobs || 0 }}</div>
            <div class="status-label">运行中</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="5">
        <el-card class="status-card">
          <div class="status-content">
            <div class="status-value error">{{ scheduleStatus?.error_jobs || 0 }}</div>
            <div class="status-label">错误任务</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="5">
        <el-card class="status-card">
          <div class="status-content">
            <div class="status-value suspended">{{ scheduleStatus?.suspended_jobs || 0 }}</div>
            <div class="status-label">暂停任务</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 任务列表 -->
    <el-card class="jobs-table-card">
      <template #header>
        <div class="card-header">
          <span>调度任务列表</span>
          <div class="table-actions">
            <el-input
              v-model="searchText"
              placeholder="搜索集群..."
              :prefix-icon="Search"
              clearable
              style="width: 200px;"
              @input="filterJobs"
            />
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="filteredJobs"
        style="width: 100%"
        empty-text="暂无调度任务"
      >
        <el-table-column prop="cluster_name" label="集群名称" width="150">
          <template #default="{ row }">
            <div class="cluster-name">
              <el-icon class="cluster-icon"><Setting /></el-icon>
              {{ row.cluster_name }}
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="status" label="任务状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getJobStatusType(row.status)" size="small">
              {{ getJobStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="interval" label="执行间隔" width="120" align="center">
          <template #default="{ row }">
            {{ formatInterval(row.interval) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="last_run" label="上次执行" width="160">
          <template #default="{ row }">
            {{ row.last_run ? formatDate(row.last_run) : '从未执行' }}
          </template>
        </el-table-column>
        
        <el-table-column prop="next_run" label="下次执行" width="160">
          <template #default="{ row }">
            {{ row.next_run ? formatDate(row.next_run) : '未计划' }}
          </template>
        </el-table-column>
        
        <el-table-column label="执行统计" width="120" align="center">
          <template #default="{ row }">
            <div class="execution-stats">
              <div class="success-count">成功: {{ row.successful_runs }}</div>
              <div class="total-count">总计: {{ row.total_runs }}</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="成功率" width="100" align="center">
          <template #default="{ row }">
            <div class="success-rate">
              {{ calculateSuccessRate(row) }}%
            </div>
            <el-progress
              :percentage="calculateSuccessRate(row)"
              :stroke-width="4"
              :show-text="false"
              :status="calculateSuccessRate(row) < 80 ? 'exception' : 'success'"
            />
          </template>
        </el-table-column>
        
        <el-table-column prop="error_count" label="错误次数" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.error_count > 0" type="danger" size="small">
              {{ row.error_count }}
            </el-tag>
            <span v-else class="no-errors">0</span>
          </template>
        </el-table-column>
        
        <el-table-column prop="last_error" label="最近错误" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.last_error" class="error-message">{{ row.last_error }}</span>
            <span v-else class="no-error">无错误</span>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template #default="{ row }">
            <el-button 
              type="primary" 
              size="small" 
              @click="restartJob(row)" 
              :loading="row.restarting"
              :disabled="!scheduleStatus?.service_running"
            >
              <el-icon><RefreshRight /></el-icon>
              重启
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 全局设置对话框 -->
    <el-dialog
      v-model="settingsVisible"
      title="全局调度设置"
      width="600px"
      @close="resetSettingsForm"
    >
      <el-form
        ref="settingsFormRef"
        :model="settingsForm"
        :rules="settingsRules"
        label-width="140px"
      >
        <el-form-item label="启用调度" prop="enabled">
          <el-switch v-model="settingsForm.enabled" />
          <span style="margin-left: 10px; color: #909399;">是否启用定时数据收集</span>
        </el-form-item>
        
        <el-form-item label="默认采集间隔" prop="default_interval">
          <el-input-number
            v-model="settingsForm.default_interval"
            :min="5"
            :max="1440"
            controls-position="right"
            style="width: 200px"
          />
          <span style="margin-left: 10px; color: #909399;">分钟</span>
        </el-form-item>
        
        <el-form-item label="最大并发任务" prop="max_concurrent_jobs">
          <el-input-number
            v-model="settingsForm.max_concurrent_jobs"
            :min="1"
            :max="20"
            controls-position="right"
            style="width: 200px"
          />
        </el-form-item>
        
        <el-form-item label="最大重试次数" prop="retry_max_attempts">
          <el-input-number
            v-model="settingsForm.retry_max_attempts"
            :min="1"
            :max="10"
            controls-position="right"
            style="width: 200px"
          />
        </el-form-item>
        
        <el-form-item label="重试间隔" prop="retry_interval">
          <el-input-number
            v-model="settingsForm.retry_interval"
            :min="1"
            :max="60"
            controls-position="right"
            style="width: 200px"
          />
          <span style="margin-left: 10px; color: #909399;">分钟</span>
        </el-form-item>
        
        <el-form-item label="启用数据持久化" prop="enable_persistence">
          <el-switch v-model="settingsForm.enable_persistence" />
          <span style="margin-left: 10px; color: #909399;">将收集的数据保存到数据库</span>
        </el-form-item>
        
        <el-form-item label="健康检查间隔" prop="health_check_interval">
          <el-input-number
            v-model="settingsForm.health_check_interval"
            :min="5"
            :max="60"
            controls-position="right"
            style="width: 200px"
          />
          <span style="margin-left: 10px; color: #909399;">分钟</span>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="settingsVisible = false">取消</el-button>
          <el-button type="primary" @click="saveSettings" :loading="settingsSaving">
            保存设置
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import {
  VideoPlay,
  VideoPause,
  Refresh,
  Setting,
  Search,
  Check,
  Close,
  RefreshRight
} from '@element-plus/icons-vue'
import { scheduleApi } from '@/api'
import type { ScheduleStatus, ScheduleJobInfo, GlobalScheduleSettings } from '@/types'

// 响应式数据
const loading = ref(false)
const serviceOperating = ref(false)
const settingsVisible = ref(false)
const settingsSaving = ref(false)
const searchText = ref('')

const scheduleStatus = ref<ScheduleStatus | null>(null)
const allJobs = ref<ScheduleJobInfo[]>([])
const filteredJobs = ref<ScheduleJobInfo[]>([])

const settingsFormRef = ref<FormInstance>()
const settingsForm = reactive<GlobalScheduleSettings>({
  enabled: true,
  default_interval: '30m',
  max_concurrent_jobs: 5,
  retry_max_attempts: 3,
  retry_interval: '5m',
  enable_persistence: true,
  health_check_interval: '10m'
})

// 表单验证规则
const settingsRules = {
  default_interval: [
    { required: true, message: '请设置默认采集间隔', trigger: 'blur' }
  ],
  max_concurrent_jobs: [
    { required: true, message: '请设置最大并发任务数', trigger: 'blur' }
  ],
  retry_max_attempts: [
    { required: true, message: '请设置最大重试次数', trigger: 'blur' }
  ]
}

// 工具函数
const formatDate = (dateStr: string): string => {
  return new Date(dateStr).toLocaleString('zh-CN')
}

const formatInterval = (intervalStr: string): string => {
  // 处理类似 "30m0s" 的格式
  const match = intervalStr.match(/(\d+)m/)
  if (match) {
    return `${match[1]}分钟`
  }
  return intervalStr
}

const getJobStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    running: 'success',
    stopped: 'info',
    error: 'danger',
    suspended: 'warning'
  }
  return statusMap[status] || 'info'
}

const getJobStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    running: '运行中',
    stopped: '已停止',
    error: '错误',
    suspended: '暂停'
  }
  return statusMap[status] || status
}

const calculateSuccessRate = (job: ScheduleJobInfo): number => {
  if (job.total_runs === 0) return 100
  return Math.round((job.successful_runs / job.total_runs) * 100)
}

// 业务逻辑函数
const loadScheduleStatus = async () => {
  try {
    const response = await scheduleApi.getStatus()
    scheduleStatus.value = response.data
  } catch (error) {
    console.error('获取调度状态失败:', error)
    ElMessage.error('获取调度状态失败')
  }
}

const loadScheduleJobs = async () => {
  try {
    loading.value = true
    const response = await scheduleApi.getJobs()
    allJobs.value = response.data || []
    filteredJobs.value = allJobs.value
  } catch (error) {
    console.error('获取调度任务失败:', error)
    ElMessage.error('获取调度任务失败')
  } finally {
    loading.value = false
  }
}

const refreshData = async () => {
  await Promise.all([
    loadScheduleStatus(),
    loadScheduleJobs()
  ])
}

const filterJobs = () => {
  if (!searchText.value.trim()) {
    filteredJobs.value = allJobs.value
    return
  }
  
  const search = searchText.value.toLowerCase()
  filteredJobs.value = allJobs.value.filter(job =>
    job.cluster_name.toLowerCase().includes(search)
  )
}

const startService = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要启动调度服务吗？启动后将开始定时收集各集群的数据。',
      '确认启动',
      {
        confirmButtonText: '启动',
        cancelButtonText: '取消',
        type: 'info'
      }
    )
    
    serviceOperating.value = true
    await scheduleApi.start()
    
    ElMessage.success('调度服务启动成功')
    await refreshData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('启动调度服务失败:', error)
      ElMessage.error('启动调度服务失败')
    }
  } finally {
    serviceOperating.value = false
  }
}

const stopService = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要停止调度服务吗？停止后将暂停所有数据收集任务。',
      '确认停止',
      {
        confirmButtonText: '停止',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    serviceOperating.value = true
    await scheduleApi.stop()
    
    ElMessage.success('调度服务停止成功')
    await refreshData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('停止调度服务失败:', error)
      ElMessage.error('停止调度服务失败')
    }
  } finally {
    serviceOperating.value = false
  }
}

const restartJob = async (job: ScheduleJobInfo) => {
  try {
    await ElMessageBox.confirm(
      `确定要重启集群 "${job.cluster_name}" 的调度任务吗？`,
      '确认重启',
      {
        confirmButtonText: '重启',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    job.restarting = true
    await scheduleApi.restartJob(job.cluster_id)
    
    ElMessage.success(`集群 ${job.cluster_name} 的任务重启成功`)
    await loadScheduleJobs()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重启任务失败:', error)
      ElMessage.error('重启任务失败')
    }
  } finally {
    job.restarting = false
  }
}

const showSettings = () => {
  if (scheduleStatus.value?.global_settings) {
    const settings = scheduleStatus.value.global_settings
    
    // 转换时间格式
    settingsForm.enabled = settings.enabled
    settingsForm.default_interval = typeof settings.default_interval === 'string' 
      ? parseInt(settings.default_interval.replace('m0s', '')) 
      : settings.default_interval
    settingsForm.max_concurrent_jobs = settings.max_concurrent_jobs
    settingsForm.retry_max_attempts = settings.retry_max_attempts  
    settingsForm.retry_interval = typeof settings.retry_interval === 'string'
      ? parseInt(settings.retry_interval.replace('m0s', ''))
      : settings.retry_interval
    settingsForm.enable_persistence = settings.enable_persistence
    settingsForm.health_check_interval = typeof settings.health_check_interval === 'string'
      ? parseInt(settings.health_check_interval.replace('m0s', ''))
      : settings.health_check_interval
  }
  
  settingsVisible.value = true
}

const resetSettingsForm = () => {
  settingsFormRef.value?.resetFields()
}

const saveSettings = async () => {
  if (!settingsFormRef.value) return
  
  try {
    await settingsFormRef.value.validate()
    
    settingsSaving.value = true
    
    // 转换为后端需要的格式
    const settingsData = {
      enabled: settingsForm.enabled,
      default_interval: `${settingsForm.default_interval}m0s`,
      max_concurrent_jobs: settingsForm.max_concurrent_jobs,
      retry_max_attempts: settingsForm.retry_max_attempts,
      retry_interval: `${settingsForm.retry_interval}m0s`,
      enable_persistence: settingsForm.enable_persistence,
      health_check_interval: `${settingsForm.health_check_interval}m0s`
    }
    
    await scheduleApi.updateSettings(settingsData)
    
    ElMessage.success('设置保存成功')
    settingsVisible.value = false
    await loadScheduleStatus()
  } catch (error) {
    console.error('保存设置失败:', error)
    ElMessage.error('保存设置失败')
  } finally {
    settingsSaving.value = false
  }
}

// 组件挂载
onMounted(() => {
  refreshData()
})
</script>

<style scoped>
.schedule-container {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 20px;
}

.header-left h1 {
  margin: 0 0 4px 0;
  color: #303133;
}

.subtitle {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.status-overview {
  margin-bottom: 20px;
}

.status-card .status-content {
  text-align: center;
  padding: 20px 10px;
}

.status-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 10px;
  font-size: 20px;
  color: white;
}

.status-icon.running {
  background-color: #67c23a;
}

.status-icon.stopped {
  background-color: #909399;
}

.status-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
  margin-bottom: 8px;
}

.status-value.running {
  color: #67c23a;
}

.status-value.error {
  color: #f56c6c;
}

.status-value.suspended {
  color: #e6a23c;
}

.status-text {
  font-size: 16px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 4px;
}

.status-label {
  color: #909399;
  font-size: 12px;
}

.jobs-table-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
}

.table-actions {
  display: flex;
  gap: 8px;
}

.cluster-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cluster-icon {
  color: #409eff;
}

.execution-stats {
  text-align: center;
}

.success-count {
  color: #67c23a;
  font-size: 12px;
}

.total-count {
  color: #909399;
  font-size: 12px;
  margin-top: 2px;
}

.success-rate {
  font-size: 12px;
  color: #303133;
  margin-bottom: 4px;
}

.no-errors {
  color: #67c23a;
}

.error-message {
  color: #f56c6c;
  font-size: 12px;
}

.no-error {
  color: #67c23a;
  font-size: 12px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>