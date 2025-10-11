import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api'

export interface UserInfo {
  id: number
  username: string
  email: string
  realName?: string
  phone?: string
  status: string
  roles?: Array<{
    id: number
    role_name: string
    role_code: string
  }>
}

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const token = ref<string | null>(localStorage.getItem('token'))
  const userInfo = ref<UserInfo | null>(null)

  // 从localStorage恢复用户信息
  const storedUser = localStorage.getItem('userInfo')
  if (storedUser) {
    try {
      userInfo.value = JSON.parse(storedUser)
    } catch (e) {
      localStorage.removeItem('userInfo')
    }
  }

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => userInfo.value?.username || '')
  const roles = computed(() => userInfo.value?.roles || [])
  const roleCodes = computed(() => roles.value.map(r => r.role_code))
  const isAdmin = computed(() => roleCodes.value.includes('admin') || roleCodes.value.includes('superadmin'))

  // 登录
  const login = async (username: string, password: string) => {
    try {
      const res = await axios.post<{
        code: number
        data: {
          token: string
          user_info: UserInfo
        }
        msg: string
      }>(`${API_BASE}/auth/login`, { username, password })

      if (res.data.code === 0) {
        token.value = res.data.data.token
        userInfo.value = res.data.data.user_info

        // 持久化到localStorage
        localStorage.setItem('token', res.data.data.token)
        localStorage.setItem('userInfo', JSON.stringify(res.data.data.user_info))

        return { success: true }
      } else {
        return { success: false, message: res.data.msg }
      }
    } catch (error: any) {
      const message = error.response?.data?.msg || '登录失败,请重试'
      return { success: false, message }
    }
  }

  // 登出
  const logout = async () => {
    try {
      await axios.post(`${API_BASE}/auth/logout`)
    } catch (e) {
      console.error('Logout error:', e)
    } finally {
      // 无论后端接口成功与否,都清除本地数据
      token.value = null
      userInfo.value = null
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
    }
  }

  // 获取当前用户信息(刷新)
  const fetchUserInfo = async () => {
    if (!token.value) return

    try {
      const res = await axios.get<{
        code: number
        data: UserInfo
        msg: string
      }>(`${API_BASE}/auth/me`, {
        headers: {
          Authorization: `Bearer ${token.value}`
        }
      })

      if (res.data.code === 0) {
        userInfo.value = res.data.data
        localStorage.setItem('userInfo', JSON.stringify(res.data.data))
      }
    } catch (error) {
      console.error('Fetch user info error:', error)
      // Token可能已失效,清除登录状态
      logout()
    }
  }

  // 检查是否有指定权限
  const hasPermission = (permission: string): boolean => {
    // 超级管理员拥有所有权限
    if (isAdmin.value) return true

    // TODO: 实现详细的权限检查逻辑
    // 需要从roles中提取permissions字段并检查
    return false
  }

  // 检查是否有指定角色
  const hasRole = (roleCode: string): boolean => {
    return roleCodes.value.includes(roleCode)
  }

  return {
    // 状态
    token,
    userInfo,

    // 计算属性
    isLoggedIn,
    username,
    roles,
    roleCodes,
    isAdmin,

    // 方法
    login,
    logout,
    fetchUserInfo,
    hasPermission,
    hasRole
  }
})