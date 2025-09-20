<template>
  <aside class="fixed left-0 top-16 bottom-0 w-64 glass-card border-r border-white/10 z-40">
    <div class="p-6">
      <!-- 导航菜单 -->
      <nav class="space-y-2">
        <router-link
          v-for="item in navigationItems"
          :key="item.name"
          :to="item.path"
          class="nav-button group"
          :class="{ 'active': $route.name === item.name }"
        >
          <component 
            :is="item.icon" 
            class="w-5 h-5 transition-transform duration-200 group-hover:scale-110" 
          />
          <span>{{ item.label }}</span>
          
          <!-- 新功能标识 -->
          <span v-if="item.isNew" 
                class="ml-auto bg-primary-500 text-white text-xs px-2 py-1 rounded-full">
            新
          </span>
          
          <!-- 状态指示器 -->
          <div v-if="item.statusCount !== undefined" 
               class="ml-auto flex items-center">
            <span class="bg-danger-500 text-white text-xs px-2 py-1 rounded-full">
              {{ item.statusCount }}
            </span>
          </div>
        </router-link>
      </nav>
      
      <!-- 快速统计卡片 -->
      <div class="mt-8 space-y-4">
        <div class="glass-card p-4">
          <h3 class="text-sm font-medium text-gray-300 mb-3">系统概览</h3>
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-sm text-gray-400">集群总数</span>
              <span class="text-lg font-semibold text-primary-400">
                {{ systemStore.stats?.total_clusters || 0 }}
              </span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-sm text-gray-400">在线集群</span>
              <span class="text-lg font-semibold text-success-400">
                {{ systemStore.stats?.online_clusters || 0 }}
              </span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-sm text-gray-400">问题Pod</span>
              <span class="text-lg font-semibold text-danger-400">
                {{ systemStore.stats?.problem_pods || 0 }}
              </span>
            </div>
          </div>
        </div>
        
        <!-- 资源健康度 -->
        <div class="glass-card p-4">
          <h3 class="text-sm font-medium text-gray-300 mb-3">资源健康度</h3>
          <div class="relative">
            <div class="w-16 h-16 mx-auto">
              <svg class="w-full h-full transform -rotate-90" viewBox="0 0 36 36">
                <path
                  d="M18 2.0845
                    a 15.9155 15.9155 0 0 1 0 31.831
                    a 15.9155 15.9155 0 0 1 0 -31.831"
                  fill="none"
                  stroke="rgba(255,255,255,0.1)"
                  stroke-width="2"
                />
                <path
                  d="M18 2.0845
                    a 15.9155 15.9155 0 0 1 0 31.831
                    a 15.9155 15.9155 0 0 1 0 -31.831"
                  fill="none"
                  :stroke="healthColor"
                  stroke-width="2"
                  :stroke-dasharray="`${healthPercentage} 100`"
                  class="transition-all duration-1000 ease-out"
                />
              </svg>
              <div class="absolute inset-0 flex items-center justify-center">
                <span class="text-lg font-bold" :class="healthTextColor">
                  {{ Math.round(healthPercentage) }}%
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { 
  LayoutDashboard, 
  Server, 
  Box, 
  BarChart3, 
  History, 
  Calendar,
  Activity
} from 'lucide-vue-next'
import { useSystemStore } from '../../stores/system'

const route = useRoute()
const systemStore = useSystemStore()

// 导航菜单项
const navigationItems = computed(() => [
  {
    name: 'Dashboard',
    path: '/',
    label: '总览',
    icon: LayoutDashboard,
  },
  {
    name: 'Clusters',
    path: '/clusters',
    label: '集群管理',
    icon: Server,
    statusCount: systemStore.stats?.total_clusters - systemStore.stats?.online_clusters || 0
  },
  {
    name: 'Pods',
    path: '/pods',
    label: 'Pod监控',
    icon: Box,
    statusCount: systemStore.stats?.problem_pods || 0
  },
  {
    name: 'Analysis',
    path: '/analysis',
    label: '资源分析',
    icon: BarChart3,
  },
  {
    name: 'History',
    path: '/history',
    label: '历史数据',
    icon: History,
  },
  {
    name: 'Schedule',
    path: '/schedule',
    label: '调度管理',
    icon: Calendar,
    isNew: true
  }
])

// 健康度计算
const healthPercentage = computed(() => {
  const stats = systemStore.stats
  if (!stats || stats.total_clusters === 0) return 0
  
  const clusterHealth = (stats.online_clusters / stats.total_clusters) * 100
  const podHealth = stats.total_pods > 0 
    ? ((stats.total_pods - stats.problem_pods) / stats.total_pods) * 100 
    : 100
  
  return (clusterHealth + podHealth) / 2
})

const healthColor = computed(() => {
  const percentage = healthPercentage.value
  if (percentage >= 80) return '#22c55e' // green
  if (percentage >= 60) return '#f59e0b' // yellow
  return '#ef4444' // red
})

const healthTextColor = computed(() => {
  const percentage = healthPercentage.value
  if (percentage >= 80) return 'text-success-400'
  if (percentage >= 60) return 'text-warning-400'
  return 'text-danger-400'
})
</script>