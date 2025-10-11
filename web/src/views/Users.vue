<template>
  <div class="h-full flex flex-col bg-gray-50">
    <!-- 页面头部 -->
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold text-gray-900">用户管理</h1>
        <button @click="showCreateDialog" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
          <Plus class="w-4 h-4 inline-block mr-1" />
          新增用户
        </button>
      </div>
    </div>

    <!-- 搜索栏 -->
    <div class="bg-white border-b border-gray-200 px-6 py-3">
      <div class="flex gap-3">
        <input v-model="keyword" @input="handleSearch" type="text" placeholder="搜索用户名、真实姓名或邮箱"
          class="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500" />
        <select v-model="statusFilter" @change="handleSearch" class="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500">
          <option value="">所有状态</option>
          <option value="active">正常</option>
          <option value="inactive">已禁用</option>
          <option value="locked">已锁定</option>
        </select>
      </div>
    </div>

    <!-- 用户列表 -->
    <div class="flex-1 overflow-auto p-6">
      <div v-if="loading" class="text-center py-12">
        <Loader2 class="w-8 h-8 animate-spin text-blue-600 mx-auto" />
        <p class="mt-2 text-gray-500">加载中...</p>
      </div>

      <div v-else-if="users.length === 0" class="text-center py-12">
        <UserX class="w-16 h-16 text-gray-400 mx-auto" />
        <p class="mt-4 text-gray-500">暂无用户数据</p>
      </div>

      <div v-else class="bg-white rounded-lg shadow overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户名</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">真实姓名</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">邮箱</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">角色</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">最后登录</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="user in users" :key="user.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ user.username }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700">{{ user.real_name || '-' }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700">{{ user.email }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <div class="flex flex-wrap gap-1">
                  <span v-for="role in user.roles" :key="role.id" class="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded">
                    {{ role.role_name }}
                  </span>
                  <span v-if="!user.roles || user.roles.length === 0" class="text-gray-400">-</span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getStatusClass(user.status)" class="px-2 py-1 text-xs rounded-full">
                  {{ getStatusText(user.status) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ user.last_login_at ? formatDate(user.last_login_at) : '从未登录' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button @click="handleEdit(user)" class="text-blue-600 hover:text-blue-900 mr-3">编辑</button>
                <button @click="handleResetPassword(user)" class="text-orange-600 hover:text-orange-900 mr-3">重置密码</button>
                <button @click="handleDelete(user)" class="text-red-600 hover:text-red-900">删除</button>
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

    <!-- 创建/编辑用户弹窗 -->
    <div v-if="showDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-md mx-4">
        <div class="px-6 py-4 border-b">
          <h2 class="text-xl font-bold">{{ editingUser ? '编辑用户' : '新增用户' }}</h2>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">用户名 *</label>
            <input v-model="formData.username" type="text" :disabled="!!editingUser"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500" />
          </div>
          <div v-if="!editingUser">
            <label class="block text-sm font-medium text-gray-700 mb-1">密码 *</label>
            <input v-model="formData.password" type="password"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">邮箱 *</label>
            <input v-model="formData.email" type="email"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">真实姓名</label>
            <input v-model="formData.real_name" type="text"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">手机号</label>
            <input v-model="formData.phone" type="tel"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">角色</label>
            <select v-model="formData.role_ids" multiple
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500">
              <option v-for="role in availableRoles" :key="role.id" :value="role.id">
                {{ role.role_name }}
              </option>
            </select>
          </div>
          <div v-if="editingUser">
            <label class="block text-sm font-medium text-gray-700 mb-1">状态</label>
            <select v-model="formData.status"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500">
              <option value="active">正常</option>
              <option value="inactive">禁用</option>
              <option value="locked">锁定</option>
            </select>
          </div>
        </div>
        <div class="px-6 py-4 border-t flex justify-end gap-2">
          <button @click="closeDialog" class="px-4 py-2 border rounded-lg hover:bg-gray-50">取消</button>
          <button @click="handleSubmit" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            {{ editingUser ? '保存' : '创建' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Plus, Loader2, UserX } from 'lucide-vue-next'
import { getUsers, createUser, updateUser, deleteUser, resetPassword } from '../api/users'
import { getActiveRoles } from '../api/roles'
import type { User, Role } from '../types/management'
import dayjs from 'dayjs'

// 状态
const loading = ref(false)
const users = ref<User[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const statusFilter = ref('')
const showDialog = ref(false)
const editingUser = ref<User | null>(null)
const availableRoles = ref<Role[]>([])

// 表单数据
const formData = ref({
  username: '',
  password: '',
  email: '',
  real_name: '',
  phone: '',
  status: 'active' as 'active' | 'inactive' | 'locked',
  role_ids: [] as number[]
})

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const res = await getUsers({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: keyword.value,
      status: statusFilter.value
    })
    users.value = res.data.data.list
    total.value = res.data.data.total
  } catch (error) {
    console.error('加载用户列表失败:', error)
    alert('加载失败，请重试')
  } finally {
    loading.value = false
  }
}

// 加载可用角色
const loadRoles = async () => {
  try {
    const res = await getActiveRoles()
    availableRoles.value = res.data.data
  } catch (error) {
    console.error('加载角色列表失败:', error)
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadUsers()
}

// 分页
const handlePageChange = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    loadUsers()
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  editingUser.value = null
  formData.value = {
    username: '',
    password: '',
    email: '',
    real_name: '',
    phone: '',
    status: 'active',
    role_ids: []
  }
  showDialog.value = true
}

// 编辑用户
const handleEdit = (user: User) => {
  editingUser.value = user
  formData.value = {
    username: user.username,
    password: '',
    email: user.email,
    real_name: user.real_name || '',
    phone: user.phone || '',
    status: user.status,
    role_ids: user.roles?.map(r => r.id) || []
  }
  showDialog.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    if (editingUser.value) {
      await updateUser(editingUser.value.id, {
        email: formData.value.email,
        real_name: formData.value.real_name,
        phone: formData.value.phone,
        status: formData.value.status,
        role_ids: formData.value.role_ids
      })
      alert('更新成功')
    } else {
      await createUser(formData.value)
      alert('创建成功')
    }
    closeDialog()
    loadUsers()
  } catch (error: any) {
    alert(error.response?.data?.msg || '操作失败')
  }
}

// 重置密码
const handleResetPassword = async (user: User) => {
  const newPassword = prompt(`重置用户 ${user.username} 的密码:`)
  if (newPassword) {
    try {
      await resetPassword(user.id, { new_password: newPassword })
      alert('密码重置成功')
    } catch (error: any) {
      alert(error.response?.data?.msg || '重置失败')
    }
  }
}

// 删除用户
const handleDelete = async (user: User) => {
  if (confirm(`确定要删除用户 ${user.username} 吗？`)) {
    try {
      await deleteUser(user.id)
      alert('删除成功')
      loadUsers()
    } catch (error: any) {
      alert(error.response?.data?.msg || '删除失败')
    }
  }
}

// 关闭对话框
const closeDialog = () => {
  showDialog.value = false
  editingUser.value = null
}

// 格式化日期
const formatDate = (date: string) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

// 状态样式
const getStatusClass = (status: string) => {
  const classes = {
    active: 'bg-green-100 text-green-800',
    inactive: 'bg-gray-100 text-gray-800',
    locked: 'bg-red-100 text-red-800'
  }
  return classes[status as keyof typeof classes] || classes.inactive
}

// 状态文本
const getStatusText = (status: string) => {
  const texts = {
    active: '正常',
    inactive: '已禁用',
    locked: '已锁定'
  }
  return texts[status as keyof typeof texts] || status
}

// 初始化
onMounted(() => {
  loadUsers()
  loadRoles()
})
</script>