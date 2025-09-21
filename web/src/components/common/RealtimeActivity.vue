<template>
  <div class="space-y-3 max-h-64 overflow-y-auto custom-scrollbar">
    <div
      v-for="(activity, index) in activities"
      :key="index"
      class="flex items-start space-x-3 p-3 rounded-lg hover:bg-white/5 transition-all duration-300 animate-slide-up"
      :style="{ animationDelay: `${index * 100}ms` }"
    >
      <!-- 状态图标 -->
      <div class="flex-shrink-0 mt-0.5">
        <div 
          class="w-8 h-8 rounded-full flex items-center justify-center"
          :class="getStatusBgClass(activity.type)"
        >
          <component 
            :is="getStatusIcon(activity.type)" 
            class="w-4 h-4"
            :class="getStatusIconClass(activity.type)"
          />
        </div>
      </div>
      
      <!-- 活动内容 -->
      <div class="flex-1 min-w-0">
        <div class="flex items-start justify-between">
          <p class="text-sm text-gray-200 leading-relaxed">
            {{ activity.message }}
          </p>
          <span class="text-xs text-gray-500 flex-shrink-0 ml-2">
            {{ activity.time }}
          </span>
        </div>
        
        <!-- 活动详情（如果有） -->
        <div v-if="activity.details" class="mt-2 text-xs text-gray-400">
          {{ activity.details }}
        </div>
      </div>
      
      <!-- 时间轴线条 -->
      <div 
        v-if="index < activities.length - 1"
        class="absolute left-6 mt-8 w-0.5 h-6 bg-gradient-to-b from-current to-transparent opacity-20"
        :class="getStatusIconClass(activity.type)"
      ></div>
    </div>
    
    <!-- 空状态 -->
    <div v-if="activities.length === 0" class="text-center py-8">
      <Activity class="w-12 h-12 text-gray-600 mx-auto mb-3" />
      <p class="text-gray-500">暂无实时活动</p>
    </div>
    
    <!-- 加载更多 -->
    <div v-if="hasMore" class="text-center py-3">
      <button 
        @click="$emit('load-more')"
        class="text-sm text-primary-400 hover:text-primary-300 transition-colors"
      >
        加载更多活动
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { 
  CheckCircle, 
  AlertTriangle, 
  XCircle, 
  Info, 
  Activity,
  Zap,
  Refresh
} from 'lucide-vue-next'

interface ActivityItem {
  type: 'success' | 'warning' | 'error' | 'info'
  message: string
  time: string
  details?: string
}

interface Props {
  activities: ActivityItem[]
  hasMore?: boolean
}

defineProps<Props>()

defineEmits<{
  'load-more': []
}>()

// 获取状态图标
const getStatusIcon = (type: ActivityItem['type']) => {
  const icons = {
    success: CheckCircle,
    warning: AlertTriangle,
    error: XCircle,
    info: Info
  }
  return icons[type]
}

// 获取状态背景样式
const getStatusBgClass = (type: ActivityItem['type']) => {
  const classes = {
    success: 'bg-success-500/20 border border-success-500/30',
    warning: 'bg-warning-500/20 border border-warning-500/30',
    error: 'bg-danger-500/20 border border-danger-500/30',
    info: 'bg-primary-500/20 border border-primary-500/30'
  }
  return classes[type]
}

// 获取状态图标颜色
const getStatusIconClass = (type: ActivityItem['type']) => {
  const classes = {
    success: 'text-success-400',
    warning: 'text-warning-400',
    error: 'text-danger-400',
    info: 'text-primary-400'
  }
  return classes[type]
}
</script>

<style scoped>
.custom-scrollbar {
  scrollbar-width: thin;
  scrollbar-color: #4b5563 transparent;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: #4b5563;
  border-radius: 2px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background-color: #6b7280;
}

@keyframes slide-up {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-slide-up {
  opacity: 0;
  animation: slide-up 0.5s ease-out forwards;
}
</style>