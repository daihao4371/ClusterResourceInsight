<template>
  <div class="analysis-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>Pod资源分析</h1>
        <p class="subtitle">分析集群中Pod的资源使用情况和配置合理性</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="refreshAnalysis" :loading="loading" :icon="Refresh">
          刷新分析
        </el-button>
        <el-button @click="exportReport" :loading="exporting" :icon="Download">
          导出报告
        </el-button>
      </div>
    </div>

    <!-- 分析概览统计 -->
    <el-row :gutter="20" class="analysis-stats">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon clusters-icon">
              <el-icon><Setting /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ analysisData?.clusters_analyzed || 0 }}</div>
              <div class="stat-label">分析集群</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon pods-icon">
              <el-icon><Box /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ analysisData?.total_pods || 0 }}</div>
              <div class="stat-label">总Pod数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon problems-icon">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ analysisData?.unreasonable_pods || 0 }}</div>
              <div class="stat-label">问题Pod</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon health-icon">
              <el-icon><PieChart /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ healthRatio }}%</div>
              <div class="stat-label">健康比例</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选和搜索 -->
    <el-card class="filter-card">
      <div class="filter-section">
        <div class="filter-group">
          <label>集群筛选:</label>
          <el-select v-model="filters.cluster" placeholder="全部集群" clearable @change="applyFilters">
            <el-option label="全部集群" value="" />
            <el-option 
              v-for="cluster in clusters" 
              :key="cluster"
              :label="cluster" 
              :value="cluster" 
            />
          </el-select>
        </div>
        
        <div class="filter-group">
          <label>命名空间:</label>
          <el-select v-model="filters.namespace" placeholder="全部命名空间" clearable @change="applyFilters">
            <el-option label="全部命名空间" value="" />
            <el-option 
              v-for="ns in namespaces" 
              :key="ns"
              :label="ns" 
              :value="ns" 
            />
          </el-select>
        </div>
        
        <div class="filter-group">
          <label>问题类型:</label>
          <el-select v-model="filters.issueType" placeholder="全部问题" clearable @change="applyFilters">
            <el-option label="全部问题" value="" />
            <el-option label="内存利用率过低" value="内存利用率过低" />
            <el-option label="CPU利用率过低" value="CPU利用率过低" />
            <el-option label="资源配置缺失" value="资源配置缺失" />
            <el-option label="配置差异过大" value="配置差异过大" />
          </el-select>
        </div>
        
        <div class="filter-group">
          <el-input
            v-model="filters.search"
            placeholder="搜索Pod名称..."
            :prefix-icon="Search"
            clearable
            @input="applyFilters"
          />
        </div>
      </div>
    </el-card>

    <!-- 问题Pod列表 -->
    <el-card class="pods-table-card">
      <template #header>
        <div class="card-header">
          <span>问题Pod列表 ({{ filteredPods.length }})</span>
          <div class="table-actions">
            <el-button size="small" @click="toggleAllDetails">
              {{ showAllDetails ? '收起详情' : '展开详情' }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="paginatedPods"
        style="width: 100%"
        empty-text="暂无问题Pod数据"
        row-key="pod_name"
        :expand-row-keys="showAllDetails ? filteredPods.map(p => p.pod_name) : []"
      >
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="pod-details">
              <el-row :gutter="20">
                <el-col :span="12">
                  <div class="detail-section">
                    <h4>内存使用情况</h4>
                    <div class="resource-progress">
                      <div class="progress-item">
                        <span>请求利用率:</span>
                        <el-progress 
                          :percentage="Math.min(row.memory_req_pct, 100)"
                          :status="row.memory_req_pct < 20 ? 'exception' : 'success'"
                          :stroke-width="8"
                        >
                          <template #default>
                            <span class="progress-text">{{ row.memory_req_pct.toFixed(1) }}%</span>
                          </template>
                        </el-progress>
                      </div>
                      <div class="progress-item">
                        <span>限制利用率:</span>
                        <el-progress 
                          :percentage="Math.min(row.memory_limit_pct, 100)"
                          :status="row.memory_limit_pct < 15 ? 'exception' : 'success'"
                          :stroke-width="8"
                        >
                          <template #default>
                            <span class="progress-text">{{ row.memory_limit_pct.toFixed(1) }}%</span>
                          </template>
                        </el-progress>
                      </div>
                    </div>
                  </div>
                </el-col>
                <el-col :span="12">
                  <div class="detail-section">
                    <h4>CPU使用情况</h4>
                    <div class="resource-progress">
                      <div class="progress-item">
                        <span>请求利用率:</span>
                        <el-progress 
                          :percentage="Math.min(row.cpu_req_pct, 100)"
                          :status="row.cpu_req_pct < 15 ? 'exception' : 'success'"
                          :stroke-width="8"
                        >
                          <template #default>
                            <span class="progress-text">{{ row.cpu_req_pct.toFixed(1) }}%</span>
                          </template>
                        </el-progress>
                      </div>
                      <div class="progress-item">
                        <span>限制利用率:</span>
                        <el-progress 
                          :percentage="Math.min(row.cpu_limit_pct, 100)"
                          :status="row.cpu_limit_pct < 10 ? 'exception' : 'success'"
                          :stroke-width="8"
                        >
                          <template #default>
                            <span class="progress-text">{{ row.cpu_limit_pct.toFixed(1) }}%</span>
                          </template>
                        </el-progress>
                      </div>
                    </div>
                  </div>
                </el-col>
              </el-row>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="cluster_name" label="集群" width="100" />
        
        <el-table-column prop="namespace" label="命名空间" width="120" />
        
        <el-table-column prop="pod_name" label="Pod名称" width="200" show-overflow-tooltip />
        
        <el-table-column prop="node_name" label="节点" width="120" show-overflow-tooltip />
        
        <el-table-column label="内存使用" width="120" align="center">
          <template #default="{ row }">
            <div class="resource-info">
              <div class="usage-text">{{ formatBytes(row.memory_usage) }}</div>
              <div class="request-text">请求: {{ formatBytes(row.memory_request) }}</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="CPU使用" width="120" align="center">
          <template #default="{ row }">
            <div class="resource-info">
              <div class="usage-text">{{ formatCPU(row.cpu_usage) }}</div>
              <div class="request-text">请求: {{ formatCPU(row.cpu_request) }}</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '合理' ? 'success' : 'danger'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="issues" label="问题描述" min-width="200">
          <template #default="{ row }">
            <div class="issues-container">
              <el-tag 
                v-for="issue in row.issues" 
                :key="issue"
                type="warning" 
                size="small"
                class="issue-tag"
              >
                {{ issue }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="creation_time" label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.creation_time) }}
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.currentPage"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="filteredPods.length"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Refresh,
  Download,
  Search,
  Setting,
  Box,
  Warning,
  PieChart
} from '@element-plus/icons-vue'
import { analysisApi } from '@/api'
import type { AnalysisResult, PodResourceInfo } from '@/types'

// 响应式数据
const loading = ref(false)
const exporting = ref(false)
const showAllDetails = ref(false)

const analysisData = ref<AnalysisResult | null>(null)
const allPods = ref<PodResourceInfo[]>([])
const filteredPods = ref<PodResourceInfo[]>([])

// 筛选条件
const filters = reactive({
  cluster: '',
  namespace: '',
  issueType: '',
  search: ''
})

// 分页配置
const pagination = reactive({
  currentPage: 1,
  pageSize: 20
})

// 计算属性
const healthRatio = computed(() => {
  if (!analysisData.value || analysisData.value.total_pods === 0) return 100
  const healthyPods = analysisData.value.total_pods - analysisData.value.unreasonable_pods
  return Math.round((healthyPods / analysisData.value.total_pods) * 100)
})

const clusters = computed(() => {
  const clusterSet = new Set(allPods.value.map(pod => pod.cluster_name))
  return Array.from(clusterSet).sort()
})

const namespaces = computed(() => {
  const nsSet = new Set(allPods.value.map(pod => pod.namespace))
  return Array.from(nsSet).sort()
})

const paginatedPods = computed(() => {
  const start = (pagination.currentPage - 1) * pagination.pageSize
  const end = start + pagination.pageSize
  return filteredPods.value.slice(start, end)
})

// 工具函数
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatCPU = (millicores: number): string => {
  if (millicores >= 1000) {
    return parseFloat((millicores / 1000).toFixed(2)) + ' 核'
  }
  return millicores + 'm'
}

const formatDate = (dateStr: string): string => {
  return new Date(dateStr).toLocaleString('zh-CN')
}

// 业务逻辑函数
const loadAnalysisData = async () => {
  try {
    loading.value = true
    const result = await analysisApi.getAnalysis() as any // 分析API直接返回数据
    
    analysisData.value = result
    allPods.value = result.top50_problems || []
    applyFilters()
    
    ElMessage.success('分析数据加载完成')
  } catch (error) {
    console.error('加载分析数据失败:', error)
    ElMessage.error('获取分析数据失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  let filtered = [...allPods.value]
  
  // 集群筛选
  if (filters.cluster) {
    filtered = filtered.filter(pod => pod.cluster_name === filters.cluster)
  }
  
  // 命名空间筛选
  if (filters.namespace) {
    filtered = filtered.filter(pod => pod.namespace === filters.namespace)
  }
  
  // 问题类型筛选
  if (filters.issueType) {
    filtered = filtered.filter(pod => 
      pod.issues.some(issue => issue.includes(filters.issueType))
    )
  }
  
  // 搜索筛选
  if (filters.search) {
    const searchLower = filters.search.toLowerCase()
    filtered = filtered.filter(pod =>
      pod.pod_name.toLowerCase().includes(searchLower) ||
      pod.node_name.toLowerCase().includes(searchLower)
    )
  }
  
  filteredPods.value = filtered
  pagination.currentPage = 1  // 重置到第一页
}

const refreshAnalysis = () => {
  loadAnalysisData()
}

const exportReport = async () => {
  try {
    exporting.value = true
    
    // 创建CSV内容
    const headers = [
      '集群', '命名空间', 'Pod名称', '节点名称',
      '内存使用(MB)', '内存请求(MB)', '内存利用率(%)',
      'CPU使用(m)', 'CPU请求(m)', 'CPU利用率(%)',
      '状态', '问题描述', '创建时间'
    ]
    
    const rows = filteredPods.value.map(pod => [
      pod.cluster_name,
      pod.namespace,
      pod.pod_name,
      pod.node_name,
      Math.round(pod.memory_usage / 1024 / 1024),
      Math.round(pod.memory_request / 1024 / 1024),
      pod.memory_req_pct.toFixed(1),
      pod.cpu_usage,
      pod.cpu_request,
      pod.cpu_req_pct.toFixed(1),
      pod.status,
      pod.issues.join(';'),
      formatDate(pod.creation_time)
    ])
    
    const csvContent = [headers, ...rows]
      .map(row => row.map(cell => `"${cell}"`).join(','))
      .join('\n')
    
    // 创建下载链接
    const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `pod-analysis-report-${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.csv`
    link.click()
    
    window.URL.revokeObjectURL(url)
    ElMessage.success('报告导出成功')
  } catch (error) {
    console.error('导出报告失败:', error)
    ElMessage.error('导出报告失败')
  } finally {
    exporting.value = false
  }
}

const toggleAllDetails = () => {
  showAllDetails.value = !showAllDetails.value
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.currentPage = 1
}

const handleCurrentChange = (page: number) => {
  pagination.currentPage = page
}

// 组件挂载时加载数据
onMounted(() => {
  loadAnalysisData()
})
</script>

<style scoped>
.analysis-container {
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

.analysis-stats {
  margin-bottom: 20px;
}

.stat-card .stat-content {
  display: flex;
  align-items: center;
  padding: 20px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 24px;
  color: white;
}

.clusters-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.pods-icon {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.problems-icon {
  background: linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%);
}

.health-icon {
  background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
}

.stat-label {
  color: #909399;
  font-size: 14px;
  margin-top: 4px;
}

.filter-card {
  margin-bottom: 20px;
}

.filter-section {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  align-items: center;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-group label {
  font-size: 14px;
  color: #606266;
  white-space: nowrap;
}

.filter-group .el-select,
.filter-group .el-input {
  width: 160px;
}

.pods-table-card {
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

.resource-info {
  text-align: center;
}

.usage-text {
  font-weight: bold;
  color: #303133;
}

.request-text {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.issues-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.issue-tag {
  margin-bottom: 2px;
}

.pod-details {
  padding: 20px;
  background-color: #f8f9fa;
  border-radius: 4px;
  margin: 10px 0;
}

.detail-section h4 {
  color: #303133;
  margin: 0 0 15px 0;
  font-size: 16px;
}

.resource-progress {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.progress-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.progress-item span:first-child {
  width: 80px;
  font-size: 14px;
  color: #606266;
}

.progress-item .el-progress {
  flex: 1;
}

.progress-text {
  font-size: 12px;
  color: #606266;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>