<template>
  <div class="space-y-4 animate-fade-in">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gradient">系统总览</h1>
        <p class="text-gray-400 text-sm">实时监控多集群资源状态与性能指标</p>
      </div>
      
      <div class="flex items-center space-x-4">
        <div class="text-sm text-gray-400">
          <span class="inline-block w-2 h-2 bg-success-500 rounded-full animate-pulse mr-2"></span>
          实时监控中
        </div>
      </div>
    </div>

    <!-- 关键指标卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <!-- 集群状态 -->
      <MetricCard
        title="集群状态"
        :value="systemStore.stats?.online_clusters || 0"
        :total="systemStore.stats?.total_clusters || 0"
        unit="个在线"
        icon="Server"
        :status="clusterStatus"
        :trend="clusterTrend"
      />
      
      <!-- Pod 总数 -->
      <MetricCard
        title="Pod 总数"
        :value="systemStore.stats?.total_pods || 0"
        unit="个"
        icon="Box"
        :status="podStatus"
        :trend="podTrend"
      />
      
      <!-- 问题 Pod -->
      <MetricCard
        title="问题 Pod"
        :value="systemStore.stats?.problem_pods || 0"
        unit="个"
        icon="AlertTriangle"
        status="error"
        :trend="problemTrend"
      />
      
      <!-- 资源效率 -->
      <MetricCard
        title="资源效率"
        :value="resourceEfficiency"
        unit="%"
        icon="Activity"
        :status="efficiencyStatus"
        :trend="efficiencyTrend"
      />
    </div>

    <!-- 图表区域 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 集群状态分布 -->
      <div class="glass-card p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-semibold">集群状态分布</h2>
          <router-link 
            to="/clusters" 
            class="btn-secondary text-sm hover:scale-105 transition-transform"
          >
            查看详情
          </router-link>
        </div>
        <ClusterStatusChart :data="clusterData" />
      </div>
      
      <!-- 资源使用趋势 -->
      <div class="glass-card p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-semibold">资源使用趋势</h2>
          <select 
            v-model="selectedTimeRange" 
            @change="onTimeRangeChange"
            class="input-field text-sm"
          >
            <option value="1">最近1小时</option>
            <option value="6">最近6小时</option>
            <option value="24">最近24小时</option>
            <option value="168">最近7天</option>
          </select>
        </div>
        <div v-if="systemStore.trendLoading" class="flex items-center justify-center h-80">
          <div class="text-gray-400">加载中...</div>
        </div>
        <ResourceTrendChart v-else :data="systemStore.trendData" />
      </div>
    </div>

    <!-- 实时活动和告警 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 实时活动 -->
      <div class="glass-card p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-semibold">实时活动</h2>
          <div class="flex items-center space-x-2">
            <div class="w-2 h-2 bg-success-500 rounded-full animate-pulse"></div>
            <span class="text-sm text-gray-400">实时更新</span>
          </div>
        </div>
        <RealtimeActivity 
          :activities="systemStore.realtimeActivities" 
          :loading="systemStore.activitiesLoading"
        />
      </div>
      
      <!-- 系统告警 -->
      <div class="glass-card p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-semibold">系统告警</h2>
          <router-link to="/alerts" class="btn-secondary text-sm">
            查看全部
          </router-link>
        </div>
        <SystemAlerts 
          :alerts="systemStore.systemAlerts" 
          :loading="systemStore.alertsLoading"
          @resolve-alert="handleResolveAlert"
          @dismiss-alert="handleDismissAlert"
          @view-detail="handleViewAlertDetail"
        />
      </div>
    </div>

    <!-- 快速操作 -->
    <div class="glass-card p-6">
      <h2 class="text-xl font-semibold mb-4">快速操作</h2>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <router-link 
          to="/clusters" 
          class="flex flex-col items-center p-4 rounded-lg hover:bg-white/5 transition-colors group"
        >
          <div class="w-12 h-12 bg-primary-500/20 rounded-full flex items-center justify-center mb-2 group-hover:scale-110 transition-transform">
            <Server class="w-6 h-6 text-primary-400" />
          </div>
          <span class="text-sm font-medium">管理集群</span>
        </router-link>
        
        <router-link 
          to="/analysis" 
          class="flex flex-col items-center p-4 rounded-lg hover:bg-white/5 transition-colors group"
        >
          <div class="w-12 h-12 bg-warning-500/20 rounded-full flex items-center justify-center mb-2 group-hover:scale-110 transition-transform">
            <BarChart3 class="w-6 h-6 text-warning-400" />
          </div>
          <span class="text-sm font-medium">资源分析</span>
        </router-link>
        
        <router-link 
          to="/pods" 
          class="flex flex-col items-center p-4 rounded-lg hover:bg-white/5 transition-colors group"
        >
          <div class="w-12 h-12 bg-success-500/20 rounded-full flex items-center justify-center mb-2 group-hover:scale-110 transition-transform">
            <Box class="w-6 h-6 text-success-400" />
          </div>
          <span class="text-sm font-medium">Pod监控</span>
        </router-link>
        
        <router-link 
          to="/schedule" 
          class="flex flex-col items-center p-4 rounded-lg hover:bg-white/5 transition-colors group"
        >
          <div class="w-12 h-12 bg-purple-500/20 rounded-full flex items-center justify-center mb-2 group-hover:scale-110 transition-transform">
            <Calendar class="w-6 h-6 text-purple-400" />
          </div>
          <span class="text-sm font-medium">调度管理</span>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Server, Box, AlertTriangle, Activity, BarChart3, Calendar } from 'lucide-vue-next'
import { useSystemStore } from '../stores/system'
import MetricCard from '../components/common/MetricCard.vue'
import ClusterStatusChart from '../components/charts/ClusterStatusChart.vue'
import ResourceTrendChart from '../components/charts/ResourceTrendChart.vue'
import RealtimeActivity from '../components/common/RealtimeActivity.vue'
import SystemAlerts from '../components/common/SystemAlerts.vue'

const systemStore = useSystemStore()
const router = useRouter()

// 计算属性
const clusterStatus = computed(() => {
  const stats = systemStore.stats
  if (!stats) return 'unknown'
  const ratio = stats.online_clusters / stats.total_clusters
  return ratio >= 0.9 ? 'success' : ratio >= 0.7 ? 'warning' : 'error'
})

const podStatus = computed(() => {
  const stats = systemStore.stats
  if (!stats) return 'unknown'
  return stats.problem_pods === 0 ? 'success' : stats.problem_pods < 10 ? 'warning' : 'error'
})

const resourceEfficiency = computed(() => {
  const stats = systemStore.stats
  if (!stats) return 0
  // 直接使用后端计算的资源效率，并保留一位小数
  return Math.round((stats.resource_efficiency || 0) * 10) / 10
})

const efficiencyStatus = computed(() => {
  const efficiency = resourceEfficiency.value
  return efficiency >= 90 ? 'success' : efficiency >= 70 ? 'warning' : 'error'
})

// 模拟数据
const clusterTrend = ref('+2.3%')
const podTrend = ref('+12.5%')
const problemTrend = ref('-8.1%')
const efficiencyTrend = ref('+1.2%')

// 时间范围选择器状态
const selectedTimeRange = ref('24') // 默认24小时

const clusterData = computed(() => {
  const stats = systemStore.stats
  if (!stats || !stats.cluster_status_distribution || stats.cluster_status_distribution.length === 0) {
    // 返回默认数据结构
    return [
      { name: '在线', value: stats?.online_clusters || 0, color: '#22c55e' },
      { name: '离线', value: (stats?.total_clusters || 0) - (stats?.online_clusters || 0), color: '#ef4444' }
    ]
  }
  // 使用后端提供的实际数据
  return stats.cluster_status_distribution
})

// 时间范围改变处理
const onTimeRangeChange = () => {
  const hours = parseInt(selectedTimeRange.value)
  systemStore.fetchTrendData(hours)
}

// 处理解决告警
const handleResolveAlert = async (alert: any) => {
  try {
    const success = await systemStore.resolveAlert(alert.id)
    if (success) {
      console.log('告警已标记为已解决')
      // 可以添加成功提示
    } else {
      console.error('解决告警失败')
    }
  } catch (error) {
    console.error('解决告警失败:', error)
  }
}

// 处理忽略告警
const handleDismissAlert = async (alert: any) => {
  try {
    const success = await systemStore.dismissAlert(alert.id)
    if (success) {
      console.log('告警已忽略')
      // 可以添加成功提示
    } else {
      console.error('忽略告警失败')
    }
  } catch (error) {
    console.error('忽略告警失败:', error)
  }
}

// 处理查看告警详情
const handleViewAlertDetail = (alert: any) => {
  // 跳转到告警页面
  router.push('/alerts')
}

onMounted(() => {
  // 初始化数据
  systemStore.fetchStats()
  // 初始化趋势数据 - 默认24小时
  systemStore.fetchTrendData(24)
  // 初始化实时活动和告警数据
  systemStore.fetchRealtimeActivities()
  systemStore.fetchSystemAlerts()
})
</script>