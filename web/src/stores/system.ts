import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { SystemStats, Notification, TrendData, ActivityItem, Alert } from '../types'
import { api } from '../utils/api'

export const useSystemStore = defineStore('system', () => {
  // 状态
  const stats = ref<SystemStats | null>(null)
  const trendData = ref<TrendData[]>([])
  const notifications = ref<Notification[]>([])
  const realtimeActivities = ref<ActivityItem[]>([])
  const systemAlerts = ref<Alert[]>([])
  const loading = ref(false)
  const trendLoading = ref(false)
  const activitiesLoading = ref(false)
  const alertsLoading = ref(false)
  const error = ref<string | null>(null)
  
  // 集群相关状态
  const currentCluster = ref<string>('') // 当前选中的集群ID

  // 获取系统统计信息
  const fetchStats = async (clusterId?: string) => {
    loading.value = true
    error.value = null
    
    try {
      // 构建URL，支持集群筛选
      let url = '/stats'
      if (clusterId && clusterId !== '') {
        url += `?cluster_id=${encodeURIComponent(clusterId)}`
      }
      
      console.log('调用统计API:', url, '集群ID:', clusterId)
      
      // 调用统计接口
      const statsRes = await api.get(url)
      
      // 直接使用统计接口返回的数据
      stats.value = {
        total_clusters: statsRes.data?.data?.total_clusters || 0,
        online_clusters: statsRes.data?.data?.online_clusters || 0,
        total_pods: statsRes.data?.data?.total_pods || 0,
        problem_pods: statsRes.data?.data?.problem_pods || 0,
        resource_efficiency: statsRes.data?.data?.resource_efficiency || 0,
        cluster_status_distribution: statsRes.data?.data?.cluster_status_distribution || [],
        last_update: statsRes.data?.data?.last_update || new Date().toISOString()
      }
      
      console.log('统计数据获取成功:', {
        集群参数: clusterId || '全部集群',
        总集群数: stats.value.total_clusters,
        在线集群: stats.value.online_clusters,
        总Pod数: stats.value.total_pods,
        问题Pod数: stats.value.problem_pods
      })
    } catch (err: any) {
      error.value = err.message || '获取系统统计信息失败'
      console.error('Failed to fetch system stats:', err)
    } finally {
      loading.value = false
    }
  }

  // 获取系统趋势数据
  const fetchTrendData = async (hours: number = 24, clusterId?: string) => {
    trendLoading.value = true
    
    try {
      // 构建URL，支持集群筛选
      let url = `/history/system-trends?hours=${hours}`
      if (clusterId && clusterId !== '') {
        url += `&cluster_id=${encodeURIComponent(clusterId)}`
      }
      
      console.log('调用趋势数据API:', url)
      
      const response = await api.get(url)
      
      // 调试日志
      console.log('趋势数据API响应:', response.data)
      
      if (response.data?.code === 0 && response.data?.data) {
        // 检查数据格式，支持多种可能的响应结构
        let trendDataArray = response.data.data
        
        // 如果data是包装对象，进一步提取
        if (trendDataArray.data && Array.isArray(trendDataArray.data)) {
          trendDataArray = trendDataArray.data
        }
        
        // 确保是数组格式
        if (Array.isArray(trendDataArray) && trendDataArray.length > 0) {
          trendData.value = trendDataArray
          console.log('成功解析趋势数据:', trendDataArray.length, '条记录')
        } else {
          console.warn('后端返回的趋势数据格式异常，使用模拟数据')
          trendData.value = generateMockTrendData(hours)
        }
      } else {
        console.warn('趋势数据API响应格式异常:', response.data)
        // 如果后端返回空数据，使用模拟数据确保图表正常显示
        trendData.value = generateMockTrendData(hours)
      }
    } catch (err: any) {
      console.error('Failed to fetch trend data:', err)
      // 发生错误时使用模拟数据确保图表正常显示
      trendData.value = generateMockTrendData(hours)
    } finally {
      trendLoading.value = false
    }
  }

  // 生成模拟趋势数据 - 作为后备方案
  const generateMockTrendData = (hours: number): TrendData[] => {
    const now = new Date()
    const data: TrendData[] = []
    
    // 根据时间范围生成不同数量的数据点
    let pointCount = 6
    if (hours <= 1) {
      pointCount = 12
    } else if (hours <= 6) {
      pointCount = 8
    }
    
    const interval = (hours * 60) / pointCount // 分钟间隔
    
    for (let i = 0; i < pointCount; i++) {
      const time = new Date(now.getTime() - (hours * 60 - interval * i) * 60000)
      data.push({
        time: time.toTimeString().slice(0, 5), // HH:MM格式
        cpu: 45 + (i % 3) * 10,    // 模拟CPU使用率波动
        memory: 60 + (i % 4) * 8,  // 模拟内存使用率波动
        pods: 120 + i * 5          // 模拟Pod数量增长
      })
    }
    
    return data
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

  // 获取实时活动数据
  const fetchRealtimeActivities = async (limit: number = 10) => {
    activitiesLoading.value = true
    
    try {
      const response = await api.get(`/activities/recent?limit=${limit}`)
      
      if (response.data?.code === 0 && response.data?.data?.data) {
        realtimeActivities.value = response.data.data.data
      } else {
        console.warn('获取实时活动数据失败，使用默认数据')
        realtimeActivities.value = []
      }
    } catch (err: any) {
      console.error('获取实时活动数据失败:', err)
      // 发生错误时清空数据
      realtimeActivities.value = []
    } finally {
      activitiesLoading.value = false
    }
  }

  // 获取系统告警数据
  const fetchSystemAlerts = async (limit: number = 10) => {
    alertsLoading.value = true
    
    try {
      const response = await api.get(`/alerts/recent?limit=${limit}`)
      
      if (response.data?.code === 0 && response.data?.data?.data) {
        systemAlerts.value = response.data.data.data
      } else {
        console.warn('获取系统告警数据失败，使用默认数据')
        systemAlerts.value = []
      }
    } catch (err: any) {
      console.error('获取系统告警数据失败:', err)
      // 发生错误时清空数据
      systemAlerts.value = []
    } finally {
      alertsLoading.value = false
    }
  }

  // 初始化示例数据（用于演示）
  const initializeSampleData = async () => {
    try {
      // 先清理旧数据
      await api.delete('/activities/cleanup')
      // 生成新的示例数据
      await api.post('/activities/sample')
      // 重新获取数据
      await Promise.all([
        fetchRealtimeActivities(),
        fetchSystemAlerts()
      ])
      return true
    } catch (err: any) {
      console.error('初始化示例数据失败:', err)
      return false
    }
  }

  // 刷新所有数据
  const refreshAllData = async (clusterId?: string) => {
    console.log('刷新所有数据，集群ID:', clusterId)
    await Promise.all([
      fetchStats(clusterId),
      fetchTrendData(24, clusterId),
      fetchRealtimeActivities(),
      fetchSystemAlerts()
    ])
  }

  // 解决告警
  const resolveAlert = async (alertId: number) => {
    try {
      await api.put(`/alerts/${alertId}/resolve`)
      // 重新获取告警数据以更新状态
      await fetchSystemAlerts()
      return true
    } catch (err: any) {
      console.error('解决告警失败:', err)
      return false
    }
  }

  // 忽略告警
  const dismissAlert = async (alertId: number) => {
    try {
      await api.put(`/alerts/${alertId}/dismiss`)
      // 重新获取告警数据以更新状态
      await fetchSystemAlerts()
      return true
    } catch (err: any) {
      console.error('忽略告警失败:', err)
      return false
    }
  }

  // 更新告警状态
  const updateAlertStatus = async (alertId: number, status: 'active' | 'resolved' | 'suppressed') => {
    try {
      await api.put(`/alerts/${alertId}/status`, { status })
      // 重新获取告警数据以更新状态
      await fetchSystemAlerts()
      return true
    } catch (err: any) {
      console.error('更新告警状态失败:', err)
      return false
    }
  }

  // 设置当前集群
  const setCurrentCluster = (clusterId: string) => {
    currentCluster.value = clusterId
    console.log('设置当前集群:', clusterId || '全部集群')
  }

  return {
    // 状态
    stats,
    trendData,
    notifications,
    realtimeActivities,
    systemAlerts,
    loading,
    trendLoading,
    activitiesLoading,
    alertsLoading,
    error,
    currentCluster,
    
    // 方法
    fetchStats,
    fetchTrendData,
    fetchRealtimeActivities,
    fetchSystemAlerts,
    addNotification,
    removeNotification,
    clearNotifications,
    initializeSampleData,
    refreshAllData,
    resolveAlert,
    dismissAlert,
    updateAlertStatus,
    setCurrentCluster
  }
})