<template>
  <div class="h-full flex flex-col bg-gray-50">
    <!-- 页面头部 -->
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold text-gray-900">API管理</h1>
        <button @click="showCreateDialog" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
          <Plus class="w-4 h-4 inline-block mr-1" />
          新增API
        </button>
      </div>
    </div>

    <!-- 搜索栏和过滤 -->
    <div class="bg-white border-b border-gray-200 px-6 py-3">
      <div class="flex gap-3">
        <input v-model="keyword" @input="handleSearch" type="text" placeholder="搜索API路径或权限标识"
          class="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500" />
        <select v-model="methodFilter" @change="handleSearch"
          class="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500">
          <option value="">所有方法</option>
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="DELETE">DELETE</option>
          <option value="PATCH">PATCH</option>
        </select>
        <select v-model="groupFilter" @change="handleSearch"
          class="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500">
          <option value="">所有分组</option>
          <option v-for="group in apiGroups" :key="group" :value="group">
            {{ group }}
          </option>
        </select>
        <select v-model="statusFilter" @change="handleSearch"
          class="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500">
          <option value="">所有状态</option>
          <option value="true">公开接口</option>
          <option value="false">需授权</option>
        </select>
      </div>
    </div>

    <!-- API列表 -->
    <div class="flex-1 overflow-auto p-6">
      <div v-if="loading" class="text-center py-12">
        <Loader2 class="w-8 h-8 animate-spin text-blue-600 mx-auto" />
        <p class="mt-2 text-gray-500">加载中...</p>
      </div>

      <div v-else-if="apis.length === 0" class="text-center py-12">
        <Globe class="w-16 h-16 text-gray-400 mx-auto" />
        <p class="mt-4 text-gray-500">暂无API数据</p>
      </div>

      <div v-else class="bg-white rounded-lg shadow overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">请求方法</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">API路径</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">分组</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">权限标识</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">描述</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">创建时间</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="api in apis" :key="api.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getMethodClass(api.method)" class="px-2 py-1 text-xs font-semibold rounded">
                  {{ api.method }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm font-mono text-gray-900">{{ api.path }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                <span class="px-2 py-1 bg-gray-100 text-gray-800 rounded text-xs">
                  {{ api.group || '-' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <code v-if="api.permission" class="px-2 py-1 bg-blue-50 text-blue-700 rounded text-xs">
                  {{ api.permission }}
                </code>
                <span v-else class="text-gray-400">-</span>
              </td>
              <td class="px-6 py-4 text-sm text-gray-700">{{ api.description || '-' }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span v-if="api.is_public" class="px-2 py-1 text-xs bg-green-100 text-green-800 rounded-full">
                  公开
                </span>
                <span v-else class="px-2 py-1 text-xs bg-yellow-100 text-yellow-800 rounded-full">
                  需授权
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ api.created_at ? formatDate(api.created_at) : '-' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button @click="handleEdit(api)" class="text-blue-600 hover:text-blue-900 mr-3">编辑</button>
                <button @click="handleDelete(api)" class="text-red-600 hover:text-red-900">删除</button>
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

    <!-- 创建/编辑API弹窗 -->
    <div v-if="showDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-2xl mx-4">
        <div class="px-6 py-4 border-b">
          <h2 class="text-xl font-bold">{{ editingApi ? '编辑API' : '新增API' }}</h2>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">请求方法 *</label>
              <select v-model="formData.method"
                class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500">
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="DELETE">DELETE</option>
                <option value="PATCH">PATCH</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">API分组</label>
              <input v-model="formData.group" type="text" placeholder="例如: users, clusters"
                class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">API路径 *</label>
            <input v-model="formData.path" type="text" placeholder="/api/v1/users"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500 font-mono text-sm" />
            <p class="text-xs text-gray-500 mt-1">路径支持动态参数，例如: /users/:id</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">权限标识</label>
            <input v-model="formData.permission" type="text" placeholder="user:view"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500" />
            <p class="text-xs text-gray-500 mt-1">用于权限校验的标识符，格式: 资源:操作</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">描述</label>
            <textarea v-model="formData.description" rows="2" placeholder="简要描述该API的功能"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500"></textarea>
          </div>
          <div class="flex items-center">
            <input v-model="formData.is_public" type="checkbox" id="is-public"
              class="mr-2 h-4 w-4 text-blue-600 rounded" />
            <label for="is-public" class="text-sm text-gray-700">
              公开接口（无需鉴权即可访问）
            </label>
          </div>
        </div>
        <div class="px-6 py-4 border-t flex justify-end gap-2">
          <button @click="closeDialog" class="px-4 py-2 border rounded-lg hover:bg-gray-50">取消</button>
          <button @click="handleSubmit" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            {{ editingApi ? '保存' : '创建' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Plus, Loader2, Globe } from 'lucide-vue-next'
import { getApis, createApi, updateApi, deleteApi, getApiGroups } from '../api/apis'
import type { ApiEndpoint } from '../types/management'
import dayjs from 'dayjs'

// 状态
const loading = ref(false)
const apis = ref<ApiEndpoint[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const methodFilter = ref('')
const groupFilter = ref('')
const statusFilter = ref('')
const showDialog = ref(false)
const editingApi = ref<ApiEndpoint | null>(null)
const apiGroups = ref<string[]>([])

// 表单数据
const formData = ref({
  path: '',
  method: 'GET' as 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH',
  group: '',
  permission: '',
  description: '',
  is_public: false
})

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 加载API列表
const loadApis = async () => {
  loading.value = true
  try {
    const res = await getApis({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: keyword.value,
      method: methodFilter.value,
      group: groupFilter.value,
      is_public: statusFilter.value ? statusFilter.value === 'true' : undefined
    })
    apis.value = res.data.data.list
    total.value = res.data.data.total
  } catch (error) {
    console.error('加载API列表失败:', error)
    alert('加载失败，请重试')
  } finally {
    loading.value = false
  }
}

// 加载API分组
const loadApiGroups = async () => {
  try {
    const res = await getApiGroups()
    apiGroups.value = res.data.data
  } catch (error) {
    console.error('加载API分组失败:', error)
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadApis()
}

// 分页
const handlePageChange = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    loadApis()
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  editingApi.value = null
  formData.value = {
    path: '',
    method: 'GET',
    group: '',
    permission: '',
    description: '',
    is_public: false
  }
  showDialog.value = true
}

// 编辑API
const handleEdit = (api: ApiEndpoint) => {
  editingApi.value = api
  formData.value = {
    path: api.path,
    method: api.method as any,
    group: api.group || '',
    permission: api.permission || '',
    description: api.description || '',
    is_public: api.is_public
  }
  showDialog.value = true
}

// 提交表单
const handleSubmit = async () => {
  // 验证必填项
  if (!formData.value.path.trim()) {
    alert('请输入API路径')
    return
  }
  if (!formData.value.method) {
    alert('请选择请求方法')
    return
  }

  try {
    const data = {
      path: formData.value.path,
      method: formData.value.method,
      group: formData.value.group,
      permission: formData.value.permission,
      description: formData.value.description,
      is_public: formData.value.is_public
    }

    if (editingApi.value) {
      await updateApi(editingApi.value.id, data)
      alert('更新成功')
    } else {
      await createApi(data)
      alert('创建成功')
    }
    closeDialog()
    loadApis()
    loadApiGroups() // 重新加载分组（可能有新分组）
  } catch (error: any) {
    alert(error.response?.data?.msg || '操作失败')
  }
}

// 删除API
const handleDelete = async (api: ApiEndpoint) => {
  if (confirm(`确定要删除API ${api.method} ${api.path} 吗？`)) {
    try {
      await deleteApi(api.id)
      alert('删除成功')
      loadApis()
      loadApiGroups() // 重新加载分组
    } catch (error: any) {
      alert(error.response?.data?.msg || '删除失败')
    }
  }
}

// 关闭对话框
const closeDialog = () => {
  showDialog.value = false
  editingApi.value = null
}

// 格式化日期
const formatDate = (date: string) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

// HTTP方法样式
const getMethodClass = (method: string) => {
  const classes = {
    GET: 'bg-blue-100 text-blue-800',
    POST: 'bg-green-100 text-green-800',
    PUT: 'bg-yellow-100 text-yellow-800',
    DELETE: 'bg-red-100 text-red-800',
    PATCH: 'bg-purple-100 text-purple-800'
  }
  return classes[method as keyof typeof classes] || 'bg-gray-100 text-gray-800'
}

// 初始化
onMounted(() => {
  loadApis()
  loadApiGroups()
})
</script>