<template>
  <div class="pods-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>Pod管理</h1>
        <p class="subtitle">查看和管理集群中的所有Pod资源</p>
      </div>
      <div class="header-actions">
        <el-button @click="refreshPods" :loading="loading" :icon="Refresh">
          刷新
        </el-button>
      </div>
    </div>

    <!-- 筛选和搜索 -->
    <el-card class="filter-card">
      <div class="filter-section">
        <div class="filter-group">
          <label>集群:</label>
          <el-select v-model="searchParams.cluster" placeholder="全部集群" clearable @change="searchPods">
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
          <el-select v-model="searchParams.namespace" placeholder="全部命名空间" clearable @change="searchPods">
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
          <el-input
            v-model="searchParams.query"
            placeholder="搜索Pod名称..."
            :prefix-icon="Search"
            clearable
            @input="debounceSearch"
          />
        </div>
        
        <div class="filter-group">
          <el-input-number
            v-model="searchParams.size"
            :min="10"
            :max="100"
            :step="10"
            controls-position="right"
            style="width: 120px"
            @change="searchPods"
          />
          <span style="margin-left: 8px; color: #909399;">条/页</span>
        </div>
      </div>
    </el-card>

    <!-- Pod列表 -->
    <el-card class="pods-table-card">
      <template #header>
        <div class="card-header">
          <span>Pod列表 ({{ podSearchResult?.total || 0 }})</span>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="podSearchResult?.pods || []"
        style="width: 100%"
        empty-text="暂无Pod数据"
      >
        <el-table-column prop="cluster_name" label="集群" width="100" />
        
        <el-table-column prop="namespace" label="命名空间" width="120" />
        
        <el-table-column prop="pod_name" label="Pod名称" width="200" show-overflow-tooltip />
        
        <el-table-column prop="node_name" label="节点" width="120" show-overflow-tooltip />
        
        <el-table-column label="内存使用" width="140" align="center">
          <template #default="{ row }">
            <div class="resource-usage">
              <div class="usage-bar">
                <el-progress
                  :percentage="Math.min(row.memory_req_pct, 100)"
                  :status="row.memory_req_pct < 20 ? 'exception' : row.memory_req_pct > 80 ? 'warning' : 'success'"
                  :stroke-width="6"
                  :show-text="false"
                />
              </div>
              <div class="usage-text">
                <span class="current">{{ formatBytes(row.memory_usage) }}</span>
                <span class="total">/ {{ formatBytes(row.memory_request) }}</span>
              </div>
              <div class="usage-percent">{{ row.memory_req_pct.toFixed(1) }}%</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="CPU使用" width="140" align="center">
          <template #default="{ row }">
            <div class="resource-usage">
              <div class="usage-bar">
                <el-progress
                  :percentage="Math.min(row.cpu_req_pct, 100)"
                  :status="row.cpu_req_pct < 15 ? 'exception' : row.cpu_req_pct > 80 ? 'warning' : 'success'"
                  :stroke-width="6"
                  :show-text="false"
                />
              </div>
              <div class="usage-text">
                <span class="current">{{ formatCPU(row.cpu_usage) }}</span>
                <span class="total">/ {{ formatCPU(row.cpu_request) }}</span>
              </div>
              <div class="usage-percent">{{ row.cpu_req_pct.toFixed(1) }}%</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '合理' ? 'success' : 'warning'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="issues" label="问题" min-width="200">
          <template #default="{ row }">
            <div v-if="row.issues && row.issues.length > 0" class="issues-container">
              <el-tag 
                v-for="issue in row.issues.slice(0, 3)" 
                :key="issue"
                type="warning" 
                size="small"
                class="issue-tag"
              >
                {{ issue }}
              </el-tag>
              <span v-if="row.issues.length > 3" class="more-issues">
                +{{ row.issues.length - 3 }}个问题
              </span>
            </div>
            <span v-else class="no-issues">无问题</span>
          </template>
        </el-table-column>
        
        <el-table-column prop="creation_time" label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.creation_time) }}
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container" v-if="podSearchResult">
        <el-pagination
          v-model:current-page="searchParams.page"
          :page-size="searchParams.size"
          :total="podSearchResult.total"
          layout="total, prev, pager, next, jumper"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { podApi } from '@/api'
import type { PodSearchRequest, PodSearchResponse } from '@/types'

// 响应式数据
const loading = ref(false)
const podSearchResult = ref<PodSearchResponse | null>(null)

// 搜索参数
const searchParams = reactive<PodSearchRequest>({
  query: '',
  namespace: '',
  cluster: '',
  page: 1,
  size: 20
})

// 用于防抖的引用
let searchTimeout: number | null = null

// 计算属性 - 提取唯一的集群和命名空间列表
const clusters = computed(() => {
  if (!podSearchResult.value?.pods) return []
  const clusterSet = new Set(podSearchResult.value.pods.map(pod => pod.cluster_name))
  return Array.from(clusterSet).sort()
})

const namespaces = computed(() => {
  if (!podSearchResult.value?.pods) return []
  const nsSet = new Set(podSearchResult.value.pods.map(pod => pod.namespace))
  return Array.from(nsSet).sort()
})

// 工具函数
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
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
const searchPods = async () => {
  try {
    loading.value = true
    const response = await podApi.searchPods(searchParams) as any
    podSearchResult.value = response
  } catch (error) {
    console.error('搜索Pod失败:', error)
    ElMessage.error('获取Pod列表失败')
  } finally {
    loading.value = false
  }
}

const debounceSearch = () => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  searchTimeout = setTimeout(() => {
    searchParams.page = 1  // 搜索时重置到第一页
    searchPods()
  }, 500)
}

const refreshPods = () => {
  searchPods()
}

const handlePageChange = (page: number) => {
  searchParams.page = page
  searchPods()
}

// 组件挂载时加载数据
onMounted(() => {
  searchPods()
})
</script>

<style scoped>
.pods-container {
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
  font-weight: bold;
}

.resource-usage {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.usage-bar {
  width: 100px;
}

.usage-text {
  font-size: 11px;
  color: #606266;
}

.usage-text .current {
  font-weight: bold;
  color: #303133;
}

.usage-text .total {
  color: #909399;
}

.usage-percent {
  font-size: 11px;
  font-weight: bold;
  color: #303133;
}

.issues-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
}

.issue-tag {
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