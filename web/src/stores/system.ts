import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { SystemStats, Notification } from '../types'
import { api } from '../utils/api'

export const useSystemStore = defineStore('system', () => {
  // 状态
  const stats = ref<SystemStats | null>(null)
  const notifications = ref<Notification[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 获取系统统计信息
  const fetchStats = async () => {
    loading.value = true
    error.value = null
    
    try {
      // 并行获取多个数据源
      const [clustersRes, analysisRes] = await Promise.all([
        api.get('/clusters'),
        api.get('/analysis')
      ])
      
      // 修复数据路径 - 根据实际API响应结构
      const clusters = clustersRes.data?.data?.data || []
      const analysis = analysisRes.data?.data || {}
      
      // 确保clusters是数组
      const clustersArray = Array.isArray(clusters) ? clusters : []
      
      stats.value = {
        total_clusters: clustersArray.length,
        online_clusters: clustersArray.filter((c: any) => c.status === 'online').length,
        total_pods: analysis.total_pods || 0,
        problem_pods: analysis.unreasonable_pods || 0,
        last_update: new Date().toISOString()
      }
    } catch (err: any) {
      error.value = err.message || '获取系统统计信息失败'
      console.error('Failed to fetch system stats:', err)
    } finally {
      loading.value = false
    }
  }

  // 添加通知
  const addNotification = (notification: Omit<Notification, 'id' | 'timestamp'>) => {
    const newNotification: Notification = {
      ...notification,
      id: Date.now().toString(),
      timestamp: Date.now(),
      duration: notification.duration || 5000
    }
    
    notifications.value.push(newNotification)
    
    // 自动移除
    if (newNotification.duration > 0) {
      setTimeout(() => {
        removeNotification(newNotification.id)
      }, newNotification.duration)
    }
  }

  // 移除通知
  const removeNotification = (id: string) => {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
      notifications.value.splice(index, 1)
    }
  }

  // 清空所有通知
  const clearNotifications = () => {
    notifications.value = []
  }

  return {
    // 状态
    stats,
    notifications,
    loading,
    error,
    
    // 方法
    fetchStats,
    addNotification,
    removeNotification,
    clearNotifications
  }
})