<template>
  <div class="h-full flex flex-col bg-gray-50">
    <!-- 页面头部 -->
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold text-gray-900">审计日志</h1>
        <button @click="handleExport" :disabled="exporting"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50">
          <Download class="w-4 h-4 inline-block mr-1" />
          {{ exporting ? '导出中...' : '导出日志' }}
        </button>
      </div>
    </div>

    <!-- 高级搜索栏 -->
    <div class="bg-white border-b border-gray-200 px-6 py-3">
      <div class="grid grid-cols-6 gap-3">
        <input v-model="filters.username" @input="handleSearch" type="text" placeholder="用户名"
          class="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500" />

        <select v-model="filters.operate_type" @change="handleSearch"
          class="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500">
          <option value="">所有操作</option>
          <option value="create">创建</option>
          <option value="update">更新</option>
          <option value="delete">删除</option>
          <option value="query">查询</option>
          <option value="login">登录</option>
          <option value="logout">登出</option>
        </select>

        <select v-model="filters.module" @change="handleSearch"
          class="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500">
          <option value="">所有模块</option>
          <option value="user">用户</option>
          <option value="role">角色</option>
          <option value="api">API</option>
          <option value="cluster">集群</option>
          <option value="pod">Pod</option>
          <option value="schedule">调度</option>
          <option value="alert">告警</option>
        </select>

        <select v-model="filters.status_code" @change="handleSearch"
          class="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500">
          <option value="">所有状态</option>
          <option value="200">成功(200)</option>
          <option value="400">请求错误(400)</option>
          <option value="401">未授权(401)</option>
          <option value="403">禁止访问(403)</option>
          <option value="404">未找到(404)</option>
          <option value="500">服务器错误(500)</option>
        </select>

        <input v-model="filters.start_time" @change="handleSearch" type="datetime-local"
          class="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500" />

        <input v-model="filters.end_time" @change="handleSearch" type="datetime-local"
          class="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500" />
      </div>
      <div class="mt-2 flex items-center gap-2">
        <label class="text-sm text-gray-600">耗时过滤:</label>
        <input v-model.number="filters.min_duration" @input="handleSearch" type="number"
          placeholder="最小(ms)" min="0"
          class="w-28 px-3 py-1 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500 text-sm" />
        <span class="text-gray-400">-</span>
        <input v-model.number="filters.max_duration" @input="handleSearch" type="number"
          placeholder="最大(ms)" min="0"
          class="w-28 px-3 py-1 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500 text-sm" />
        <button @click="resetFilters" class="ml-auto px-3 py-1 text-sm text-gray-600 hover:text-gray-900 border rounded-lg">
          重置筛选
        </button>
      </div>
    </div>

    <!-- 日志列表 -->
    <div class="flex-1 overflow-auto p-6">
      <div v-if="loading" class="text-center py-12">
        <Loader2 class="w-8 h-8 animate-spin text-blue-600 mx-auto" />
        <p class="mt-2 text-gray-500">加载中...</p>
      </div>

      <div v-else-if="logs.length === 0" class="text-center py-12">
        <FileText class="w-16 h-16 text-gray-400 mx-auto" />
        <p class="mt-4 text-gray-500">暂无审计日志</p>
      </div>

      <div v-else class="bg-white rounded-lg shadow overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">时间</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作类型</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">模块</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">IP地址</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">请求路径</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">耗时</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="log in logs" :key="log.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ formatDate(log.created_at) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                {{ log.username }}
                <div class="text-xs text-gray-500">ID: {{ log.user_id }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getOperateTypeClass(log.operate_type)" class="px-2 py-1 text-xs rounded-full">
                  {{ getOperateTypeName(log.operate_type) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                <span class="px-2 py-1 bg-gray-100 rounded text-xs">
                  {{ log.module || '-' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700 font-mono">
                {{ log.ip_address || '-' }}
              </td>
              <td class="px-6 py-4 text-sm font-mono text-gray-700">
                <div class="max-w-xs truncate" :title="log.request_path">
                  {{ log.request_method }} {{ log.request_path }}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getStatusClass(log.status_code)" class="px-2 py-1 text-xs rounded">
                  {{ log.status_code }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <span :class="getDurationClass(log.duration)">
                  {{ log.duration }}ms
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button @click="showDetail(log)" class="text-blue-600 hover:text-blue-900">详情</button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- 分页 -->
        <div class="bg-gray-50 px-6 py-3 flex items-center justify-between border-t border-gray-200">
          <div class="text-sm text-gray-500">
            共 {{ total }} 条记录，第 {{ currentPage }} / {{ totalPages }} 页
          </div>
          <div class="flex gap-2">
            <button @click="handlePageChange(currentPage - 1)" :disabled="currentPage === 1"
              class="px-3 py-1 border rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed">
              上一页
            </button>
            <button @click="handlePageChange(currentPage + 1)" :disabled="currentPage === totalPages"
              class="px-3 py-1 border rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed">
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 日志详情弹窗 -->
    <div v-if="showDetailDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-4xl mx-4 max-h-[90vh] overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b">
          <h2 class="text-xl font-bold">审计日志详情</h2>
        </div>
        <div v-if="selectedLog" class="px-6 py-4 flex-1 overflow-y-auto space-y-4">
          <!-- 基本信息 -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-sm font-medium text-gray-500">日志ID</label>
              <div class="mt-1 text-sm text-gray-900">{{ selectedLog.id }}</div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500">操作时间</label>
              <div class="mt-1 text-sm text-gray-900">{{ formatDate(selectedLog.created_at) }}</div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500">用户信息</label>
              <div class="mt-1 text-sm text-gray-900">{{ selectedLog.username }} (ID: {{ selectedLog.user_id }})</div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500">IP地址</label>
              <div class="mt-1 text-sm text-gray-900 font-mono">{{ selectedLog.ip_address || '-' }}</div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500">操作类型</label>
              <div class="mt-1">
                <span :class="getOperateTypeClass(selectedLog.operate_type)" class="px-2 py-1 text-xs rounded-full">
                  {{ getOperateTypeName(selectedLog.operate_type) }}
                </span>
              </div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500">模块</label>
              <div class="mt-1 text-sm text-gray-900">{{ selectedLog.module || '-' }}</div>
            </div>
          </div>

          <!-- 请求信息 -->
          <div>
            <label class="text-sm font-medium text-gray-500">请求路径</label>
            <div class="mt-1 text-sm text-gray-900 font-mono">
              {{ selectedLog.request_method }} {{ selectedLog.request_path }}
            </div>
          </div>

          <div>
            <label class="text-sm font-medium text-gray-500">请求体（已脱敏）</label>
            <pre class="mt-1 text-xs bg-gray-50 p-3 rounded border overflow-x-auto">{{ formatJson(selectedLog.request_body) }}</pre>
          </div>

          <!-- 响应信息 -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-sm font-medium text-gray-500">响应状态</label>
              <div class="mt-1">
                <span :class="getStatusClass(selectedLog.status_code)" class="px-2 py-1 text-xs rounded">
                  {{ selectedLog.status_code }}
                </span>
              </div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500">执行耗时</label>
              <div class="mt-1">
                <span :class="getDurationClass(selectedLog.duration)" class="text-sm">
                  {{ selectedLog.duration }}ms
                </span>
              </div>
            </div>
          </div>

          <div v-if="selectedLog.error_msg">
            <label class="text-sm font-medium text-red-600">错误信息</label>
            <div class="mt-1 text-sm text-red-700 bg-red-50 p-3 rounded border border-red-200">
              {{ selectedLog.error_msg }}
            </div>
          </div>

          <div v-if="selectedLog.user_agent">
            <label class="text-sm font-medium text-gray-500">User-Agent</label>
            <div class="mt-1 text-xs text-gray-700 font-mono bg-gray-50 p-2 rounded">
              {{ selectedLog.user_agent }}
            </div>
          </div>
        </div>
        <div class="px-6 py-4 border-t flex justify-end">
          <button @click="closeDetailDialog" class="px-4 py-2 border rounded-lg hover:bg-gray-50">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Download, Loader2, FileText } from 'lucide-vue-next'
import { queryAuditLogs, exportAuditLogs } from '../api/audit'
import type { AuditLog } from '../types/management'
import dayjs from 'dayjs'

// 状态
const loading = ref(false)
const exporting = ref(false)
const logs = ref<AuditLog[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const showDetailDialog = ref(false)
const selectedLog = ref<AuditLog | null>(null)

// 筛选条件
const filters = ref({
  username: '',
  operate_type: '',
  module: '',
  status_code: '',
  start_time: '',
  end_time: '',
  min_duration: undefined as number | undefined,
  max_duration: undefined as number | undefined
})

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 加载审计日志
const loadLogs = async () => {
  loading.value = true
  try {
    const res = await queryAuditLogs({
      page: currentPage.value,
      page_size: pageSize.value,
      username: filters.value.username || undefined,
      operate_type: filters.value.operate_type || undefined,
      module: filters.value.module || undefined,
      status_code: filters.value.status_code ? parseInt(filters.value.status_code) : undefined,
      start_time: filters.value.start_time || undefined,
      end_time: filters.value.end_time || undefined,
      min_duration: filters.value.min_duration,
      max_duration: filters.value.max_duration
    })
    logs.value = res.data.data.list
    total.value = res.data.data.total
  } catch (error) {
    console.error('加载审计日志失败:', error)
    alert('加载失败，请重试')
  } finally {
    loading.value = false
  }
}

// 搜索（防抖处理可选）
const handleSearch = () => {
  currentPage.value = 1
  loadLogs()
}

// 重置筛选
const resetFilters = () => {
  filters.value = {
    username: '',
    operate_type: '',
    module: '',
    status_code: '',
    start_time: '',
    end_time: '',
    min_duration: undefined,
    max_duration: undefined
  }
  handleSearch()
}

// 分页
const handlePageChange = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    loadLogs()
  }
}

// 显示详情
const showDetail = (log: AuditLog) => {
  selectedLog.value = log
  showDetailDialog.value = true
}

const closeDetailDialog = () => {
  showDetailDialog.value = false
  selectedLog.value = null
}

// 导出日志
const handleExport = async () => {
  if (exporting.value) return

  const confirmed = confirm('确定要导出当前筛选条件下的审计日志吗？')
  if (!confirmed) return

  exporting.value = true
  try {
    const res = await exportAuditLogs({
      username: filters.value.username || undefined,
      operate_type: filters.value.operate_type || undefined,
      module: filters.value.module || undefined,
      status_code: filters.value.status_code ? parseInt(filters.value.status_code) : undefined,
      start_time: filters.value.start_time || undefined,
      end_time: filters.value.end_time || undefined,
      min_duration: filters.value.min_duration,
      max_duration: filters.value.max_duration
    })

    // 创建下载链接
    const blob = new Blob([res.data], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `audit_logs_${dayjs().format('YYYYMMDD_HHmmss')}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    alert('导出成功')
  } catch (error: any) {
    alert(error.response?.data?.msg || '导出失败')
  } finally {
    exporting.value = false
  }
}

// 格式化日期
const formatDate = (date: string) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

// 格式化JSON
const formatJson = (jsonStr: string) => {
  try {
    const obj = JSON.parse(jsonStr || '{}')
    return JSON.stringify(obj, null, 2)
  } catch {
    return jsonStr || '-'
  }
}

// 操作类型样式
const getOperateTypeClass = (type: string) => {
  const classes = {
    create: 'bg-green-100 text-green-800',
    update: 'bg-blue-100 text-blue-800',
    delete: 'bg-red-100 text-red-800',
    query: 'bg-gray-100 text-gray-800',
    login: 'bg-purple-100 text-purple-800',
    logout: 'bg-yellow-100 text-yellow-800'
  }
  return classes[type as keyof typeof classes] || 'bg-gray-100 text-gray-800'
}

const getOperateTypeName = (type: string) => {
  const names: Record<string, string> = {
    create: '创建',
    update: '更新',
    delete: '删除',
    query: '查询',
    login: '登录',
    logout: '登出'
  }
  return names[type] || type
}

// 状态码样式
const getStatusClass = (statusCode: number) => {
  if (statusCode >= 200 && statusCode < 300) {
    return 'bg-green-100 text-green-800'
  } else if (statusCode >= 400 && statusCode < 500) {
    return 'bg-yellow-100 text-yellow-800'
  } else if (statusCode >= 500) {
    return 'bg-red-100 text-red-800'
  }
  return 'bg-gray-100 text-gray-800'
}

// 耗时样式
const getDurationClass = (duration: number) => {
  if (duration < 100) return 'text-green-600'
  if (duration < 500) return 'text-yellow-600'
  if (duration < 1000) return 'text-orange-600'
  return 'text-red-600 font-semibold'
}

// 初始化
onMounted(() => {
  loadLogs()
})
</script>