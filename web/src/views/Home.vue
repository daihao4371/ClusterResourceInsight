<template>
  <div class="home-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1>集群资源监控概览</h1>
      <p class="subtitle">实时监控 Kubernetes 集群资源使用情况和配置合理性</p>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stat-content">
            <div class="stat-icon cluster">
              <el-icon><Setting /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ dashboardData.clustersAnalyzed }}</div>
              <div class="stat-label">分析集群</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stat-content">
            <div class="stat-icon pods">
              <el-icon><Box /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ dashboardData.totalPods }}</div>
              <div class="stat-label">总 Pod 数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stat-content">
            <div class="stat-icon problems">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ dashboardData.unreasonablePods }}</div>
              <div class="stat-label">不合理配置</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stat-content">
            <div class="stat-icon ratio">
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

    <!-- 快速操作 -->
    <el-row :gutter="20" class="actions-row">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>快速操作</span>
              <el-button type="primary" @click="refreshData" :loading="loading">
                <el-icon><Refresh /></el-icon>
                刷新数据
              </el-button>
            </div>
          </template>
          
          <div class="action-buttons">
            <el-button type="primary" @click="$router.push('/analysis')">
              <el-icon><DataAnalysis /></el-icon>
              查看资源分析
            </el-button>
            
            <el-button type="success" @click="$router.push('/clusters')">
              <el-icon><Setting /></el-icon>
              管理集群
            </el-button>
            
            <el-button type="info" @click="$router.push('/pods')">
              <el-icon><Box /></el-icon>
              Pod 管理
            </el-button>
            
            <el-button type="warning" @click="triggerCollection" :loading="collecting">
              <el-icon><Download /></el-icon>
              手动收集数据
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 命名空间概览 -->
    <el-row :gutter="20" class="namespaces-row">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>命名空间概览</span>
              <el-button type="text" @click="$router.push('/pods')">查看全部</el-button>
            </div>
          </template>
          
          <el-table 
            v-loading="namespacesLoading"
            :data="namespaces.slice(0, 10)" 
            style="width: 100%"
            empty-text="暂无数据"
          >
            <el-table-column prop="cluster_name" label="集群" width="120" />
            <el-table-column prop="namespace_name" label="命名空间" width="150" />
            <el-table-column prop="total_pods" label="Pod 总数" width="100" align="center" />
            <el-table-column prop="unreasonable_pods" label="问题 Pod" width="100" align="center">
              <template #default="scope">
                <el-tag v-if="scope.row.unreasonable_pods > 0" type="danger">
                  {{ scope.row.unreasonable_pods }}
                </el-tag>
                <span v-else>0</span>
              </template>
            </el-table-column>
            <el-table-column label="内存使用" width="120">
              <template #default="scope">
                {{ formatBytes(scope.row.total_memory_usage) }}
              </template>
            </el-table-column>
            <el-table-column label="CPU使用" width="120">
              <template #default="scope">
                {{ formatCPU(scope.row.total_cpu_usage) }}
              </template>
            </el-table-column>
            <el-table-column label="健康状态" width="100">
              <template #default="scope">
                <el-tag 
                  :type="getHealthTagType(scope.row.total_pods, scope.row.unreasonable_pods)"
                >
                  {{ getHealthStatus(scope.row.total_pods, scope.row.unreasonable_pods) }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近问题 Pod -->
    <el-row :gutter="20" class="problems-row">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最新发现的问题 Pod (Top 10)</span>
              <el-button type="text" @click="$router.push('/analysis')">查看全部</el-button>
            </div>
          </template>
          
          <el-table 
            v-loading="problemsLoading"
            :data="recentProblems" 
            style="width: 100%"
            empty-text="暂无问题 Pod"
          >
            <el-table-column prop="cluster_name" label="集群" width="100" />
            <el-table-column prop="namespace" label="命名空间" width="120" />
            <el-table-column prop="pod_name" label="Pod 名称" width="180" />
            <el-table-column label="内存使用率" width="120">
              <template #default="scope">
                <el-progress 
                  :percentage="scope.row.memory_req_pct" 
                  :status="scope.row.memory_req_pct < 20 ? 'exception' : 'success'"
                  :stroke-width="8"
                />
              </template>
            </el-table-column>
            <el-table-column label="CPU使用率" width="120">
              <template #default="scope">
                <el-progress 
                  :percentage="scope.row.cpu_req_pct" 
                  :status="scope.row.cpu_req_pct < 15 ? 'exception' : 'success'"
                  :stroke-width="8"
                />
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="80">
              <template #default="scope">
                <el-tag :type="scope.row.status === '合理' ? 'success' : 'danger'">
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="issues" label="问题描述">
              <template #default="scope">
                <el-tag 
                  v-for="issue in scope.row.issues" 
                  :key="issue" 
                  type="warning" 
                  size="small"
                  style="margin-right: 4px; margin-bottom: 2px;"
                >
                  {{ issue }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Setting, 
  Box, 
  Warning, 
  PieChart, 
  Refresh, 
  DataAnalysis, 
  Download 
} from '@element-plus/icons-vue'
import { analysisApi, statisticsApi, historyApi } from '@/api'
import type { NamespaceSummary, PodResourceInfo } from '@/types'

// 响应式数据
const loading = ref(false)
const collecting = ref(false)
const namespacesLoading = ref(false)
const problemsLoading = ref(false)

const dashboardData = ref({
  clustersAnalyzed: 0,
  totalPods: 0,
  unreasonablePods: 0,
})

const namespaces = ref<NamespaceSummary[]>([])
const recentProblems = ref<PodResourceInfo[]>([])

// 计算属性
const healthRatio = computed(() => {
  if (dashboardData.value.totalPods === 0) return 100
  const healthyPods = dashboardData.value.totalPods - dashboardData.value.unreasonablePods
  return Math.round((healthyPods / dashboardData.value.totalPods) * 100)
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

const getHealthTagType = (total: number, problems: number) => {
  if (total === 0) return 'info'
  const ratio = problems / total
  if (ratio === 0) return 'success'
  if (ratio < 0.1) return 'warning'
  return 'danger'
}

const getHealthStatus = (total: number, problems: number) => {
  if (total === 0) return '无数据'
  const ratio = problems / total
  if (ratio === 0) return '健康'
  if (ratio < 0.1) return '良好'
  if (ratio < 0.3) return '一般'
  return '较差'
}

// 数据加载函数
const loadDashboardData = async () => {
  try {
    loading.value = true
    const result = await analysisApi.getAnalysis() as any // 分析API直接返回数据，无需.data
    
    dashboardData.value = {
      clustersAnalyzed: result.clusters_analyzed || 0,
      totalPods: result.total_pods || 0,
      unreasonablePods: result.unreasonable_pods || 0,
    }
    
    // 取前10个问题Pod
    recentProblems.value = result.top50_problems?.slice(0, 10) || []
  } catch (error) {
    console.error('加载分析数据失败:', error)
    ElMessage.error('获取分析数据失败')
  } finally {
    loading.value = false
  }
}

const loadNamespaces = async () => {
  try {
    namespacesLoading.value = true
    const response = await statisticsApi.getNamespacesSummary() // 统计API返回 {data: [...]}
    namespaces.value = response.data || []
  } catch (error) {
    console.error('加载命名空间数据失败:', error)
    ElMessage.error('获取命名空间数据失败')
  } finally {
    namespacesLoading.value = false
  }
}

const refreshData = async () => {
  await Promise.all([
    loadDashboardData(),
    loadNamespaces()
  ])
  ElMessage.success('数据刷新完成')
}

const triggerCollection = async () => {
  try {
    collecting.value = true
    await historyApi.collectData(true)
    ElMessage.success('数据收集已触发，请稍后刷新查看结果')
    
    // 延迟刷新数据
    setTimeout(() => {
      refreshData()
    }, 2000)
  } catch (error) {
    console.error('触发数据收集失败:', error)
    ElMessage.error('触发数据收集失败')
  } finally {
    collecting.value = false
  }
}

// 组件挂载时加载数据
onMounted(() => {
  refreshData()
})
</script>

<style scoped>
.home-container {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  text-align: center;
  margin-bottom: 30px;
}

.page-header h1 {
  color: #303133;
  margin-bottom: 10px;
}

.subtitle {
  color: #909399;
  font-size: 16px;
}

.stats-row,
.actions-row,
.namespaces-row,
.problems-row {
  margin-bottom: 20px;
}

.stats-card .stat-content {
  display: flex;
  align-items: center;
  padding: 10px 0;
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

.stat-icon.cluster {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.pods {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.problems {
  background: linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%);
}

.stat-icon.ratio {
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
}

.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.action-buttons .el-button {
  flex: 1;
  min-width: 140px;
}
</style>