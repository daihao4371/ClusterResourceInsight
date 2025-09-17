<template>
  <div class="history-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>历史数据</h1>
        <p class="subtitle">查看Pod资源使用的历史记录和趋势</p>
      </div>
      <div class="header-actions">
        <el-button @click="refreshData" :loading="loading" :icon="Refresh">
          刷新
        </el-button>
        <el-button @click="triggerCollection" :loading="collecting" :icon="Download">
          手动收集
        </el-button>
      </div>
    </div>

    <!-- 历史统计 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ statistics?.total_records || 0 }}</div>
            <div class="stat-label">历史记录总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ statistics?.cluster_count || 0 }}</div>
            <div class="stat-label">监控集群数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ statistics?.namespace_count || 0 }}</div>
            <div class="stat-label">涉及命名空间</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ formatDate(statistics?.latest_record) }}</div>
            <div class="stat-label">最新记录时间</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 查询条件 -->
    <el-card class="query-card">
      <div class="query-section">
        <el-form :model="queryParams" inline>
          <el-form-item label="集群ID:">
            <el-input-number
              v-model="queryParams.cluster_id"
              :min="0"
              placeholder="全部集群"
              controls-position="right"
              style="width: 120px"
            />
          </el-form-item>
          
          <el-form-item label="命名空间:">
            <el-input
              v-model="queryParams.namespace"
              placeholder="全部命名空间"
              style="width: 140px"
            />
          </el-form-item>
          
          <el-form-item label="Pod名称:">
            <el-input
              v-model="queryParams.pod_name"
              placeholder="Pod名称"
              style="width: 140px"
            />
          </el-form-item>
          
          <el-form-item label="时间范围:">
            <el-date-picker
              v-model="timeRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width: 350px"
            />
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" @click="queryHistory" :loading="loading">
              查询
            </el-button>
            <el-button @click="resetQuery">
              重置
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-card>

    <!-- 历史记录表格 -->
    <el-card class="history-table-card">
      <template #header>
        <div class="card-header">
          <span>历史记录 ({{ historyData?.total || 0 }})</span>
          <div class="table-actions">
            <el-button size="small" @click="cleanupOldData" type="warning">
              清理过期数据
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="historyData?.data || []"
        style="width: 100%"
        empty-text="暂无历史数据"
      >
        <el-table-column prop="cluster_id" label="集群ID" width="80" align="center" />
        
        <el-table-column prop="namespace" label="命名空间" width="120" />
        
        <el-table-column prop="pod_name" label="Pod名称" width="180" show-overflow-tooltip />
        
        <el-table-column prop="node_name" label="节点" width="120" show-overflow-tooltip />
        
        <el-table-column label="内存指标" width="180" align="center">
          <template #default="{ row }">
            <div class="metrics-info">
              <div class="metric-row">
                <span class="metric-label">使用:</span>
                <span class="metric-value">{{ formatBytes(row.memory_usage) }}</span>
              </div>
              <div class="metric-row">
                <span class="metric-label">请求:</span>
                <span class="metric-value">{{ formatBytes(row.memory_request) }}</span>
              </div>
              <div class="metric-row">
                <span class="metric-label">利用率:</span>
                <span class="metric-value" :class="getUtilizationClass(row.memory_req_pct)">
                  {{ row.memory_req_pct.toFixed(1) }}%
                </span>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="CPU指标" width="180" align="center">
          <template #default="{ row }">
            <div class="metrics-info">
              <div class="metric-row">
                <span class="metric-label">使用:</span>
                <span class="metric-value">{{ formatCPU(row.cpu_usage) }}</span>
              </div>
              <div class="metric-row">
                <span class="metric-label">请求:</span>
                <span class="metric-value">{{ formatCPU(row.cpu_request) }}</span>
              </div>
              <div class="metric-row">
                <span class="metric-label">利用率:</span>
                <span class="metric-value" :class="getUtilizationClass(row.cpu_req_pct)">
                  {{ row.cpu_req_pct.toFixed(1) }}%
                </span>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'reasonable' ? 'success' : 'warning'" size="small">
              {{ row.status === 'reasonable' ? '合理' : '不合理' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="collected_at" label="收集时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.collected_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="问题" min-width="200">
          <template #default="{ row }">
            <div v-if="row.issues && JSON.parse(row.issues).length > 0">
              <el-tag 
                v-for="issue in JSON.parse(row.issues).slice(0, 2)" 
                :key="issue"
                type="warning" 
                size="small"
                class="issue-tag"
              >
                {{ issue }}
              </el-tag>
              <span v-if="JSON.parse(row.issues).length > 2" class="more-issues">
                +{{ JSON.parse(row.issues).length - 2 }}个
              </span>
            </div>
            <span v-else class="no-issues">无问题</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container" v-if="historyData">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="historyData.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Download } from '@element-plus/icons-vue'
import { historyApi } from '@/api'
import type { HistoryQueryRequest, PodMetricsHistory } from '@/types'

// 响应式数据
const loading = ref(false)
const collecting = ref(false)
const timeRange = ref<[string, string]>(['', ''])

const statistics = ref<any>(null)
const historyData = ref<{
  data: PodMetricsHistory[]
  total: number
  page: number
  size: number
  total_pages: number
} | null>(null)

// 查询参数
const queryParams = reactive<HistoryQueryRequest>({
  cluster_id: 0,
  namespace: '',
  pod_name: '',
  start_time: '',
  end_time: '',
  page: 1,
  size: 20,
  order_by: 'collected_at',
  order_desc: true
})

// 工具函数
const formatBytes = (bytes: number): string => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const formatCPU = (millicores: number): string => {
  if (!millicores) return '0m'
  if (millicores >= 1000) {
    return parseFloat((millicores / 1000).toFixed(2)) + ' 核'
  }
  return millicores + 'm'
}

const formatDate = (dateStr?: string): string => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const getUtilizationClass = (percent: number): string => {
  if (percent < 20) return 'low-utilization'
  if (percent > 80) return 'high-utilization'
  return 'normal-utilization'
}

// 业务逻辑函数
const loadStatistics = async () => {
  try {
    const response = await historyApi.getStatistics()
    statistics.value = response.data
  } catch (error) {
    console.error('加载历史统计失败:', error)
    ElMessage.error('获取历史统计失败')
  }
}

const queryHistory = async () => {
  try {
    loading.value = true
    
    // 处理时间范围
    if (timeRange.value && timeRange.value[0] && timeRange.value[1]) {
      queryParams.start_time = timeRange.value[0]
      queryParams.end_time = timeRange.value[1]
    } else {
      queryParams.start_time = ''
      queryParams.end_time = ''
    }
    
    // 清理空值参数
    const cleanParams = Object.fromEntries(
      Object.entries(queryParams).filter(([_, value]) => 
        value !== '' && value !== 0 && value !== undefined && value !== null
      )
    )
    
    const response = await historyApi.queryHistory(cleanParams) as any
    historyData.value = response
  } catch (error) {
    console.error('查询历史数据失败:', error)
    ElMessage.error('查询历史数据失败')
  } finally {
    loading.value = false
  }
}

const resetQuery = () => {
  Object.assign(queryParams, {
    cluster_id: 0,
    namespace: '',
    pod_name: '',
    start_time: '',
    end_time: '',
    page: 1,
    size: 20,
    order_by: 'collected_at',
    order_desc: true
  })
  timeRange.value = ['', '']
  queryHistory()
}

const refreshData = async () => {
  await Promise.all([
    loadStatistics(),
    queryHistory()
  ])
}

const triggerCollection = async () => {
  try {
    collecting.value = true
    await historyApi.collectData(true)
    ElMessage.success('数据收集已触发')
    
    // 延迟3秒后刷新数据
    setTimeout(() => {
      refreshData()
    }, 3000)
  } catch (error) {
    console.error('触发数据收集失败:', error)
    ElMessage.error('触发数据收集失败')
  } finally {
    collecting.value = false
  }
}

const cleanupOldData = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清理30天前的历史数据吗？此操作不可撤销。',
      '确认清理',
      {
        confirmButtonText: '清理',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await historyApi.cleanupData(30)
    ElMessage.success('过期数据清理成功')
    await refreshData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('清理过期数据失败:', error)
      ElMessage.error('清理过期数据失败')
    }
  }
}

const handleSizeChange = (size: number) => {
  queryParams.size = size
  queryParams.page = 1
  queryHistory()
}

const handleCurrentChange = (page: number) => {
  queryParams.page = page
  queryHistory()
}

// 组件挂载
onMounted(() => {
  refreshData()
})
</script>

<style scoped>
.history-container {
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

.stats-row {
  margin-bottom: 20px;
}

.stat-card .stat-content {
  text-align: center;
  padding: 20px;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
  margin-bottom: 8px;
}

.stat-label {
  color: #909399;
  font-size: 14px;
}

.query-card {
  margin-bottom: 20px;
}

.query-section {
  padding: 10px 0;
}

.history-table-card {
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

.metrics-info {
  text-align: left;
  font-size: 12px;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 2px;
}

.metric-label {
  color: #909399;
  width: 40px;
}

.metric-value {
  font-weight: bold;
  color: #303133;
}

.low-utilization {
  color: #f56c6c !important;
}

.normal-utilization {
  color: #67c23a !important;
}

.high-utilization {
  color: #e6a23c !important;
}

.issue-tag {
  margin-right: 4px;
  margin-bottom: 2px;
}

.more-issues {
  font-size: 12px;
  color: #909399;
}

.no-issues {
  color: #67c23a;
  font-size: 12px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>