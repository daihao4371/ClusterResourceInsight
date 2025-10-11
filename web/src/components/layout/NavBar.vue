<template>
  <nav class="fixed top-0 left-0 right-0 z-50 h-12 glass-card border-b border-white/10">
    <div class="flex items-center justify-between px-4 h-full">
      <!-- Logo 和标题 -->
      <div class="flex items-center space-x-3">
        <div class="flex items-center space-x-2">
          <div class="w-6 h-6 bg-gradient-to-r from-primary-500 to-primary-600 rounded-lg flex items-center justify-center">
            <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 2L2 7v10c0 5.55 3.84 9.739 9 11 5.16-1.261 9-5.45 9-11V7l-10-5z"/>
            </svg>
          </div>
          <div>
            <h1 class="text-lg font-semibold text-gradient">ClusterResourceInsight</h1>
            <p class="text-xs text-gray-400 leading-none">K8s多集群资源监控平台</p>
          </div>
        </div>
      </div>
      
      <!-- 中间状态信息 -->
      <div class="flex items-center space-x-6">
        <!-- 系统状态 -->
        <div class="flex items-center space-x-2">
          <div class="status-indicator" :class="systemStatusClass"></div>
          <span class="text-sm text-gray-300">{{ systemStatusText }}</span>
        </div>
        
        <!-- 最后更新时间 -->
        <div class="text-sm text-gray-400">
          最后更新: {{ lastUpdateTime }}
        </div>
      </div>
      
      <!-- 右侧操作区 -->
      <div class="flex items-center space-x-3">
        <!-- 主题切换器 -->
        <ThemeSwitcher />

        <!-- 刷新按钮 -->
        <button
          @click="refreshData"
          :disabled="isRefreshing"
          class="btn-secondary p-1.5 w-8 h-8 flex items-center justify-center"
          :class="{ 'animate-spin': isRefreshing }"
        >
          <RotateCw class="w-3.5 h-3.5" />
        </button>

        <!-- 设置按钮 -->
        <button class="btn-secondary p-1.5 w-8 h-8 flex items-center justify-center">
          <Settings class="w-3.5 h-3.5" />
        </button>

        <!-- 通知中心 -->
        <div class="relative">
          <button class="btn-secondary p-1.5 w-8 h-8 flex items-center justify-center">
            <Bell class="w-3.5 h-3.5" />
            <span v-if="notificationCount > 0"
                  class="absolute -top-1 -right-1 bg-danger-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center text-xs">
              {{ notificationCount }}
            </span>
          </button>
        </div>

        <!-- 用户菜单 -->
        <div class="relative">
          <button
            @click="showUserMenu = !showUserMenu"
            class="flex items-center space-x-2 btn-secondary px-3 py-1.5"
          >
            <UserCircle class="w-4 h-4" />
            <span class="text-sm">{{ authStore.username }}</span>
            <ChevronDown class="w-3 h-3" />
          </button>

          <!-- 下拉菜单 -->
          <div
            v-if="showUserMenu"
            class="absolute right-0 mt-2 w-48 glass-card border border-white/10 rounded-lg shadow-lg py-2 z-50"
          >
            <div class="px-4 py-2 border-b border-white/10">
              <p class="text-sm font-medium text-white">{{ authStore.userInfo?.realName || authStore.username }}</p>
              <p class="text-xs text-gray-400">{{ authStore.userInfo?.email }}</p>
            </div>
            <button
              @click="handleLogout"
              class="w-full px-4 py-2 text-sm text-left hover:bg-white/5 flex items-center space-x-2 text-red-400 hover:text-red-300"
            >
              <LogOut class="w-4 h-4" />
              <span>退出登录</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { RotateCw, Settings, Bell, UserCircle, ChevronDown, LogOut } from 'lucide-vue-next'
import { useSystemStore } from '../../stores/system'
import { useAuthStore } from '../../stores/auth'
import { formatDistanceToNow } from '../../utils/date'
import ThemeSwitcher from '../common/ThemeSwitcher.vue'

const router = useRouter()
const systemStore = useSystemStore()
const authStore = useAuthStore()
const isRefreshing = ref(false)
const notificationCount = ref(0)
const showUserMenu = ref(false)

// 系统状态计算属性
const systemStatusClass = computed(() => {
  const onlineRatio = systemStore.stats?.online_clusters / systemStore.stats?.total_clusters || 0
  if (onlineRatio >= 0.9) return 'status-online'
  if (onlineRatio >= 0.7) return 'status-warning'
  return 'status-error'
})

const systemStatusText = computed(() => {
  const stats = systemStore.stats
  if (!stats) return '状态未知'
  return `${stats.online_clusters}/${stats.total_clusters} 集群在线`
})

const lastUpdateTime = computed(() => {
  if (!systemStore.stats?.last_update) return '暂无数据'
  return formatDistanceToNow(systemStore.stats.last_update)
})

// 刷新数据
const refreshData = async () => {
  if (isRefreshing.value) return

  isRefreshing.value = true
  try {
    await systemStore.fetchStats()
  } finally {
    isRefreshing.value = false
  }
}

// 退出登录
const handleLogout = async () => {
  showUserMenu.value = false
  await authStore.logout()
  router.push('/login')
}

// 点击外部关闭菜单
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.relative')) {
    showUserMenu.value = false
  }
}

// 定时刷新
let refreshInterval: number

onMounted(() => {
  // 初始加载数据
  systemStore.fetchStats()

  // 每30秒自动刷新
  refreshInterval = setInterval(() => {
    systemStore.fetchStats()
  }, 30000)

  // 监听点击外部关闭菜单
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
  document.removeEventListener('click', handleClickOutside)
})
</script>