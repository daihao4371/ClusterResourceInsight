<template>
  <div class="h-full flex flex-col bg-gray-50">
    <!-- 页面头部 -->
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold text-gray-900">角色管理</h1>
        <button @click="showCreateDialog" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
          <Plus class="w-4 h-4 inline-block mr-1" />
          新增角色
        </button>
      </div>
    </div>

    <!-- 搜索栏 -->
    <div class="bg-white border-b border-gray-200 px-6 py-3">
      <div class="flex gap-3">
        <input v-model="keyword" @input="handleSearch" type="text" placeholder="搜索角色名称或角色代码"
          class="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500" />
      </div>
    </div>

    <!-- 角色列表 -->
    <div class="flex-1 overflow-auto p-6">
      <div v-if="loading" class="text-center py-12">
        <Loader2 class="w-8 h-8 animate-spin text-blue-600 mx-auto" />
        <p class="mt-2 text-gray-500">加载中...</p>
      </div>

      <div v-else-if="roles.length === 0" class="text-center py-12">
        <Shield class="w-16 h-16 text-gray-400 mx-auto" />
        <p class="mt-4 text-gray-500">暂无角色数据</p>
      </div>

      <div v-else class="bg-white rounded-lg shadow overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">角色名称</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">角色代码</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">描述</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">权限数</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户数</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">系统角色</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">创建时间</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="role in roles" :key="role.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ role.role_name }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                <code class="px-2 py-1 bg-gray-100 rounded">{{ role.role_code }}</code>
              </td>
              <td class="px-6 py-4 text-sm text-gray-700">{{ role.description || '-' }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                <span class="px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-xs">
                  {{ getPermissionCount(role.permissions) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                <button @click="showRoleUsers(role)" class="text-blue-600 hover:underline">
                  {{ role.user_count || 0 }} 人
                </button>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span v-if="role.is_system" class="px-2 py-1 text-xs bg-purple-100 text-purple-800 rounded-full">
                  系统角色
                </span>
                <span v-else class="text-gray-400">-</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ role.created_at ? formatDate(role.created_at) : '-' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button @click="handleEdit(role)" class="text-blue-600 hover:text-blue-900 mr-3">编辑</button>
                <button @click="handleDelete(role)" :disabled="role.is_system"
                  :class="role.is_system ? 'text-gray-400 cursor-not-allowed' : 'text-red-600 hover:text-red-900'">
                  删除
                </button>
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

    <!-- 创建/编辑角色弹窗 -->
    <div v-if="showDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto">
        <div class="px-6 py-4 border-b sticky top-0 bg-white">
          <h2 class="text-xl font-bold">{{ editingRole ? '编辑角色' : '新增角色' }}</h2>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">角色名称 *</label>
            <input v-model="formData.role_name" type="text"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">角色代码 *</label>
            <input v-model="formData.role_code" type="text" :disabled="editingRole?.is_system"
              placeholder="例如: admin, viewer, operator"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500 disabled:bg-gray-100" />
            <p class="text-xs text-gray-500 mt-1">角色代码用于程序内部识别，只能包含字母、数字和下划线</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">角色描述</label>
            <textarea v-model="formData.description" rows="3"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:border-blue-500"></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">权限配置 *</label>
            <div class="border rounded-lg p-4 space-y-2 max-h-60 overflow-y-auto">
              <div class="flex items-center">
                <input v-model="allPermissionsSelected" @change="toggleAllPermissions" type="checkbox"
                  id="perm-all" class="mr-2 h-4 w-4 text-blue-600 rounded" />
                <label for="perm-all" class="text-sm font-medium text-gray-900">全部权限</label>
              </div>
              <hr class="my-2" />
              <div v-for="(group, groupName) in permissionGroups" :key="groupName" class="space-y-2">
                <div class="font-medium text-sm text-gray-700 mt-3">{{ groupName }}</div>
                <div v-for="perm in group" :key="perm.code" class="flex items-center ml-4">
                  <input v-model="formData.permissions" :value="perm.code" type="checkbox"
                    :id="'perm-' + perm.code" class="mr-2 h-4 w-4 text-blue-600 rounded" />
                  <label :for="'perm-' + perm.code" class="text-sm text-gray-700">
                    {{ perm.name }}
                    <span class="text-xs text-gray-500">({{ perm.code }})</span>
                  </label>
                </div>
              </div>
            </div>
            <p class="text-xs text-gray-500 mt-1">已选择 {{ formData.permissions.length }} 项权限</p>
          </div>
        </div>
        <div class="px-6 py-4 border-t flex justify-end gap-2 sticky bottom-0 bg-white">
          <button @click="closeDialog" class="px-4 py-2 border rounded-lg hover:bg-gray-50">取消</button>
          <button @click="handleSubmit" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            {{ editingRole ? '保存' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 角色用户列表弹窗 -->
    <div v-if="showUsersDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-3xl mx-4 max-h-[80vh] overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b">
          <h2 class="text-xl font-bold">角色用户列表 - {{ selectedRole?.role_name }}</h2>
        </div>
        <div class="px-6 py-4 flex-1 overflow-y-auto">
          <div v-if="roleUsers.length === 0" class="text-center py-8 text-gray-500">
            该角色暂无关联用户
          </div>
          <table v-else class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">用户名</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">真实姓名</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">邮箱</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="user in roleUsers" :key="user.id">
                <td class="px-4 py-3 text-sm font-medium text-gray-900">{{ user.username }}</td>
                <td class="px-4 py-3 text-sm text-gray-700">{{ user.real_name || '-' }}</td>
                <td class="px-4 py-3 text-sm text-gray-700">{{ user.email }}</td>
                <td class="px-4 py-3">
                  <span :class="getStatusClass(user.status)" class="px-2 py-1 text-xs rounded-full">
                    {{ getStatusText(user.status) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="px-6 py-4 border-t flex justify-end">
          <button @click="closeUsersDialog" class="px-4 py-2 border rounded-lg hover:bg-gray-50">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Plus, Loader2, Shield } from 'lucide-vue-next'
import { getRoles, createRole, updateRole, deleteRole, getRoleUsers } from '../api/roles'
import type { Role, User } from '../types/management'
import dayjs from 'dayjs'

// 状态
const loading = ref(false)
const roles = ref<Role[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const showDialog = ref(false)
const showUsersDialog = ref(false)
const editingRole = ref<Role | null>(null)
const selectedRole = ref<Role | null>(null)
const roleUsers = ref<User[]>([])

// 表单数据
const formData = ref({
  role_name: '',
  role_code: '',
  description: '',
  permissions: [] as string[]
})

// 权限分组配置（实际项目中应从后端获取）
const permissionGroups = ref({
  '集群管理': [
    { code: 'cluster:view', name: '查看集群' },
    { code: 'cluster:create', name: '创建集群' },
    { code: 'cluster:update', name: '更新集群' },
    { code: 'cluster:delete', name: '删除集群' }
  ],
  'Pod管理': [
    { code: 'pod:view', name: '查看Pod' },
    { code: 'pod:delete', name: '删除Pod' },
    { code: 'pod:logs', name: '查看日志' }
  ],
  '资源分析': [
    { code: 'analysis:view', name: '查看分析' },
    { code: 'analysis:export', name: '导出报告' }
  ],
  '调度管理': [
    { code: 'schedule:view', name: '查看调度' },
    { code: 'schedule:create', name: '创建调度' },
    { code: 'schedule:update', name: '更新调度' },
    { code: 'schedule:delete', name: '删除调度' },
    { code: 'schedule:execute', name: '执行调度' }
  ],
  '告警管理': [
    { code: 'alert:view', name: '查看告警' },
    { code: 'alert:handle', name: '处理告警' }
  ],
  '系统管理': [
    { code: 'user:view', name: '查看用户' },
    { code: 'user:create', name: '创建用户' },
    { code: 'user:update', name: '更新用户' },
    { code: 'user:delete', name: '删除用户' },
    { code: 'role:view', name: '查看角色' },
    { code: 'role:create', name: '创建角色' },
    { code: 'role:update', name: '更新角色' },
    { code: 'role:delete', name: '删除角色' },
    { code: 'api:view', name: '查看API' },
    { code: 'api:manage', name: '管理API' },
    { code: 'audit:view', name: '查看审计日志' },
    { code: 'audit:export', name: '导出审计日志' }
  ]
})

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 全选权限checkbox状态
const allPermissionsSelected = computed({
  get: () => {
    const allPerms = getAllPermissionCodes()
    return allPerms.length > 0 && formData.value.permissions.length === allPerms.length
  },
  set: () => { /* 由toggleAllPermissions处理 */ }
})

// 获取所有权限代码
const getAllPermissionCodes = () => {
  const codes: string[] = []
  Object.values(permissionGroups.value).forEach(group => {
    group.forEach(perm => codes.push(perm.code))
  })
  return codes
}

// 切换全选
const toggleAllPermissions = () => {
  if (allPermissionsSelected.value) {
    formData.value.permissions = []
  } else {
    formData.value.permissions = getAllPermissionCodes()
  }
}

// 获取权限数量
const getPermissionCount = (permissions: string) => {
  try {
    const perms = JSON.parse(permissions || '[]')
    // 如果包含通配符*，表示全部权限
    if (perms.includes('*')) return '全部'
    return perms.length
  } catch {
    return 0
  }
}

// 加载角色列表
const loadRoles = async () => {
  loading.value = true
  try {
    const res = await getRoles({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: keyword.value
    })
    roles.value = res.data.data.list
    total.value = res.data.data.total
  } catch (error) {
    console.error('加载角色列表失败:', error)
    alert('加载失败，请重试')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadRoles()
}

// 分页
const handlePageChange = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    loadRoles()
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  editingRole.value = null
  formData.value = {
    role_name: '',
    role_code: '',
    description: '',
    permissions: []
  }
  showDialog.value = true
}

// 编辑角色
const handleEdit = (role: Role) => {
  editingRole.value = role

  // 解析权限
  let permissions: string[] = []
  try {
    const parsed = JSON.parse(role.permissions || '[]')
    // 如果是通配符*，则选中所有权限
    if (parsed.includes('*')) {
      permissions = getAllPermissionCodes()
    } else {
      permissions = parsed
    }
  } catch {
    permissions = []
  }

  formData.value = {
    role_name: role.role_name,
    role_code: role.role_code,
    description: role.description || '',
    permissions
  }
  showDialog.value = true
}

// 提交表单
const handleSubmit = async () => {
  // 验证必填项
  if (!formData.value.role_name.trim()) {
    alert('请输入角色名称')
    return
  }
  if (!formData.value.role_code.trim()) {
    alert('请输入角色代码')
    return
  }
  if (formData.value.permissions.length === 0) {
    alert('请至少选择一项权限')
    return
  }

  try {
    const data = {
      role_name: formData.value.role_name,
      role_code: formData.value.role_code,
      description: formData.value.description,
      permissions: JSON.stringify(formData.value.permissions)
    }

    if (editingRole.value) {
      await updateRole(editingRole.value.id, data)
      alert('更新成功')
    } else {
      await createRole(data)
      alert('创建成功')
    }
    closeDialog()
    loadRoles()
  } catch (error: any) {
    alert(error.response?.data?.msg || '操作失败')
  }
}

// 删除角色
const handleDelete = async (role: Role) => {
  if (role.is_system) {
    alert('系统角色不能删除')
    return
  }

  if (confirm(`确定要删除角色 ${role.role_name} 吗？`)) {
    try {
      await deleteRole(role.id)
      alert('删除成功')
      loadRoles()
    } catch (error: any) {
      alert(error.response?.data?.msg || '删除失败')
    }
  }
}

// 显示角色用户
const showRoleUsers = async (role: Role) => {
  selectedRole.value = role
  try {
    const res = await getRoleUsers(role.id)
    roleUsers.value = res.data.data
    showUsersDialog.value = true
  } catch (error: any) {
    alert(error.response?.data?.msg || '获取用户列表失败')
  }
}

// 关闭对话框
const closeDialog = () => {
  showDialog.value = false
  editingRole.value = null
}

const closeUsersDialog = () => {
  showUsersDialog.value = false
  selectedRole.value = null
  roleUsers.value = []
}

// 格式化日期
const formatDate = (date: string) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

// 状态样式（复用）
const getStatusClass = (status: string) => {
  const classes = {
    active: 'bg-green-100 text-green-800',
    inactive: 'bg-gray-100 text-gray-800',
    locked: 'bg-red-100 text-red-800'
  }
  return classes[status as keyof typeof classes] || classes.inactive
}

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
  loadRoles()
})
</script>