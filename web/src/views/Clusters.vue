<template>
  <div class="clusters-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>集群管理</h1>
        <p class="subtitle">管理和监控 Kubernetes 集群配置</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="showAddDialog" :icon="Plus">
          添加集群
        </el-button>
        <el-button @click="batchTestClusters" :loading="batchTesting" :icon="Connection">
          批量测试
        </el-button>
        <el-button @click="refreshClusters" :loading="loading" :icon="Refresh">
          刷新
        </el-button>
      </div>
    </div>

    <!-- 集群统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ clusters.length }}</div>
            <div class="stat-label">总集群数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value online">{{ onlineClusters }}</div>
            <div class="stat-label">在线集群</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value offline">{{ offlineClusters }}</div>
            <div class="stat-label">离线集群</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value unknown">{{ unknownClusters }}</div>
            <div class="stat-label">未知状态</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 集群列表 -->
    <el-card class="clusters-table-card">
      <template #header>
        <div class="card-header">
          <span>集群列表</span>
          <div class="search-box">
            <el-input
              v-model="searchText"
              placeholder="搜索集群..."
              :prefix-icon="Search"
              clearable
              @input="filterClusters"
            />
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="filteredClusters"
        style="width: 100%"
        empty-text="暂无集群数据"
      >
        <el-table-column prop="cluster_name" label="集群名称" width="150">
          <template #default="{ row }">
            <div class="cluster-name">
              <el-icon class="cluster-icon"><Setting /></el-icon>
              {{ row.cluster_name }}
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="cluster_alias" label="别名" width="120" />
        
        <el-table-column prop="api_server" label="API Server" width="200" show-overflow-tooltip />
        
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="collect_interval" label="采集间隔" width="100" align="center">
          <template #default="{ row }">
            {{ row.collect_interval }}分钟
          </template>
        </el-table-column>
        
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="240" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="testCluster(row)" :loading="row.testing">
              <el-icon><Connection /></el-icon>
              测试
            </el-button>
            <el-button size="small" @click="editCluster(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button type="danger" size="small" @click="deleteCluster(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加/编辑集群对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑集群' : '添加集群'"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="clusterFormRef"
        :model="clusterForm"
        :rules="formRules"
        label-width="120px"
      >
        <el-form-item label="集群名称" prop="cluster_name">
          <el-input v-model="clusterForm.cluster_name" placeholder="请输入集群名称" />
        </el-form-item>
        
        <el-form-item label="集群别名" prop="cluster_alias">
          <el-input v-model="clusterForm.cluster_alias" placeholder="请输入集群别名（可选）" />
        </el-form-item>
        
        <el-form-item label="API Server" prop="api_server">
          <el-input v-model="clusterForm.api_server" placeholder="https://kubernetes-api-server:6443" />
        </el-form-item>
        
        <el-form-item label="认证类型" prop="auth_type">
          <el-select v-model="clusterForm.auth_type" style="width: 100%">
            <el-option label="Token认证" value="token" />
            <el-option label="证书认证" value="cert" />
            <el-option label="Kubeconfig" value="kubeconfig" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="认证配置" prop="auth_config">
          <el-input
            v-model="clusterForm.auth_config"
            type="textarea"
            :rows="4"
            :placeholder="getAuthPlaceholder()"
          />
        </el-form-item>
        
        <el-form-item label="采集间隔" prop="collect_interval">
          <el-input-number
            v-model="clusterForm.collect_interval"
            :min="5"
            :max="1440"
            controls-position="right"
            style="width: 200px"
          />
          <span style="margin-left: 10px; color: #909399;">分钟</span>
        </el-form-item>
        
        <el-form-item label="集群标签">
          <el-input v-model="tagsInput" placeholder="用逗号分隔多个标签" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveCluster" :loading="saving">
            {{ isEditing ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import {
  Plus,
  Refresh,
  Search,
  Connection,
  Setting,
  Edit,
  Delete
} from '@element-plus/icons-vue'
import { clusterApi } from '@/api'
import type { Cluster, CreateClusterRequest, UpdateClusterRequest } from '@/types'

// 响应式数据定义
const loading = ref(false)
const saving = ref(false)
const batchTesting = ref(false)
const dialogVisible = ref(false)
const isEditing = ref(false)
const searchText = ref('')
const tagsInput = ref('')

const clusters = ref<Cluster[]>([])
const filteredClusters = ref<Cluster[]>([])

const clusterFormRef = ref<FormInstance>()
const clusterForm = reactive<CreateClusterRequest>({
  cluster_name: '',
  cluster_alias: '',
  api_server: '',
  auth_type: 'token',
  auth_config: {
    bearer_token: ''
  },
  tags: [],
  collect_interval: 30
})

// 计算属性
const onlineClusters = computed(() => 
  clusters.value.filter(c => c.status === 'online').length
)

const offlineClusters = computed(() => 
  clusters.value.filter(c => c.status === 'offline').length
)

const unknownClusters = computed(() => 
  clusters.value.filter(c => c.status === 'unknown').length
)

// 表单验证规则
const formRules = {
  cluster_name: [
    { required: true, message: '请输入集群名称', trigger: 'blur' }
  ],
  api_server: [
    { required: true, message: '请输入API Server地址', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL', trigger: 'blur' }
  ],
  auth_type: [
    { required: true, message: '请选择认证类型', trigger: 'change' }
  ],
  auth_config: [
    { required: true, message: '请输入认证配置', trigger: 'blur' }
  ]
}

// 工具函数
const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString('zh-CN')
}

const getStatusTagType = (status: string) => {
  const statusMap: Record<string, string> = {
    online: 'success',
    offline: 'danger',
    unknown: 'warning'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    online: '在线',
    offline: '离线',
    unknown: '未知'
  }
  return statusMap[status] || status
}

const getAuthPlaceholder = () => {
  const placeholders: Record<string, string> = {
    token: 'Bearer Token',
    cert: '客户端证书内容',
    kubeconfig: 'Kubeconfig文件内容'
  }
  return placeholders[clusterForm.auth_type] || ''
}

// 业务逻辑函数
const loadClusters = async () => {
  try {
    loading.value = true
    const response = await clusterApi.getClusters()
    clusters.value = response.data || []
    filteredClusters.value = clusters.value
  } catch (error) {
    console.error('加载集群列表失败:', error)
    ElMessage.error('获取集群列表失败')
  } finally {
    loading.value = false
  }
}

const filterClusters = () => {
  if (!searchText.value.trim()) {
    filteredClusters.value = clusters.value
    return
  }
  
  const search = searchText.value.toLowerCase()
  filteredClusters.value = clusters.value.filter(cluster =>
    cluster.cluster_name.toLowerCase().includes(search) ||
    cluster.cluster_alias?.toLowerCase().includes(search) ||
    cluster.api_server.toLowerCase().includes(search)
  )
}

const refreshClusters = () => {
  loadClusters()
}

const showAddDialog = () => {
  isEditing.value = false
  dialogVisible.value = true
}

const editCluster = (cluster: Cluster) => {
  isEditing.value = true
  
  // 填充表单数据
  clusterForm.cluster_name = cluster.cluster_name
  clusterForm.cluster_alias = cluster.cluster_alias || ''
  clusterForm.api_server = cluster.api_server
  clusterForm.auth_type = cluster.auth_type
  clusterForm.collect_interval = cluster.collect_interval
  
  // 设置当前编辑的集群ID
  clusterForm.id = cluster.id
  
  dialogVisible.value = true
}

const resetForm = () => {
  Object.assign(clusterForm, {
    cluster_name: '',
    cluster_alias: '',
    api_server: '',
    auth_type: 'token',
    auth_config: { bearer_token: '' },
    tags: [],
    collect_interval: 30
  })
  tagsInput.value = ''
  clusterFormRef.value?.resetFields()
}

const saveCluster = async () => {
  if (!clusterFormRef.value) return
  
  try {
    await clusterFormRef.value.validate()
    
    saving.value = true
    
    // 处理标签
    clusterForm.tags = tagsInput.value
      .split(',')
      .map(tag => tag.trim())
      .filter(tag => tag.length > 0)
    
    if (isEditing.value) {
      await clusterApi.updateCluster(clusterForm.id!, clusterForm as UpdateClusterRequest)
      ElMessage.success('集群更新成功')
    } else {
      await clusterApi.createCluster(clusterForm)
      ElMessage.success('集群创建成功')
    }
    
    dialogVisible.value = false
    await loadClusters()
  } catch (error) {
    console.error('保存集群失败:', error)
    ElMessage.error('保存集群失败')
  } finally {
    saving.value = false
  }
}

const testCluster = async (cluster: Cluster) => {
  try {
    cluster.testing = true
    const result = await clusterApi.testCluster(cluster.id) as any
    
    if (result.success) {
      ElMessage.success(`集群 ${cluster.cluster_name} 连接正常`)
    } else {
      ElMessage.warning(`集群 ${cluster.cluster_name} 连接失败: ${result.message}`)
    }
  } catch (error) {
    console.error('测试集群连接失败:', error)
    ElMessage.error('测试集群连接失败')
  } finally {
    cluster.testing = false
  }
}

const batchTestClusters = async () => {
  try {
    batchTesting.value = true
    const results = await clusterApi.batchTestClusters() as any
    
    let successCount = 0
    let failCount = 0
    
    results.forEach((result: any) => {
      if (result.success) {
        successCount++
      } else {
        failCount++
      }
    })
    
    ElMessage.info(`批量测试完成: 成功 ${successCount} 个，失败 ${failCount} 个`)
    await loadClusters()
  } catch (error) {
    console.error('批量测试失败:', error)
    ElMessage.error('批量测试失败')
  } finally {
    batchTesting.value = false
  }
}

const deleteCluster = async (cluster: Cluster) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除集群 "${cluster.cluster_name}" 吗？此操作不可撤销。`,
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await clusterApi.deleteCluster(cluster.id)
    ElMessage.success('集群删除成功')
    await loadClusters()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除集群失败:', error)
      ElMessage.error('删除集群失败')
    }
  }
}

// 组件挂载
onMounted(() => {
  loadClusters()
})
</script>

<style scoped>
.clusters-container {
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
  font-size: 28px;
  font-weight: bold;
  line-height: 1;
  margin-bottom: 8px;
}

.stat-value.online {
  color: #67c23a;
}

.stat-value.offline {
  color: #f56c6c;
}

.stat-value.unknown {
  color: #e6a23c;
}

.stat-label {
  color: #909399;
  font-size: 14px;
}

.clusters-table-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
}

.search-box {
  width: 250px;
}

.cluster-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cluster-icon {
  color: #409eff;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>