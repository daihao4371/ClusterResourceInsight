<template>
  <div class="space-y-3 max-h-64 overflow-y-auto custom-scrollbar">
    <div
      v-for="(alert, index) in alerts"
      :key="index"
      class="alert-card group"
      :class="getAlertClass(alert.level)"
      @click="$emit('alert-click', alert)"
    >
      <!-- 告警级别指示器 -->
      <div class="flex items-start space-x-3">
        <div class="flex-shrink-0">
          <div 
            class="w-3 h-3 rounded-full animate-pulse"
            :class="getIndicatorClass(alert.level)"
          ></div>
        </div>
        
        <!-- 告警内容 -->
        <div class="flex-1 min-w-0">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <h4 class="text-sm font-medium text-white group-hover:text-gray-100 transition-colors">
                {{ alert.title }}
              </h4>
              <p class="text-sm text-gray-300 mt-1 leading-relaxed">
                {{ alert.description }}
              </p>
            </div>
            
            <!-- 告警级别标签 -->
            <span 
              class="ml-3 flex-shrink-0 px-2 py-1 text-xs font-medium rounded-full"
              :class="getLevelBadgeClass(alert.level)"
            >
              {{ getLevelText(alert.level) }}
            </span>
          </div>
          
          <!-- 时间和操作 -->
          <div class="flex items-center justify-between mt-3">
            <span class="text-xs text-gray-500">
              {{ alert.time }}
            </span>
            
            <!-- 快速操作 -->
            <div class="flex space-x-1 opacity-60 group-hover:opacity-100 transition-opacity">
              <button 
                @click.stop="$emit('resolve-alert', alert)"
                class="p-1 hover:bg-white/10 rounded transition-colors hover:scale-110 active:scale-95"
                title="标记为已解决"
              >
                <Check class="w-3 h-3 text-success-400" />
              </button>
              <button 
                @click.stop="$emit('dismiss-alert', alert)"
                class="p-1 hover:bg-white/10 rounded transition-colors hover:scale-110 active:scale-95"
                title="忽略告警"
              >
                <X class="w-3 h-3 text-gray-400" />
              </button>
              <button 
                @click.stop="$emit('view-detail', alert)"
                class="p-1 hover:bg-white/10 rounded transition-colors hover:scale-110 active:scale-95"
                title="查看详情"
              >
                <ExternalLink class="w-3 h-3 text-primary-400" />
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 进度条（如果有处理进度） -->
      <div v-if="alert.progress !== undefined" class="mt-3">
        <div class="flex justify-between text-xs text-gray-400 mb-1">
          <span>处理进度</span>
          <span>{{ alert.progress }}%</span>
        </div>
        <div class="w-full bg-dark-700 rounded-full h-1.5">
          <div 
            class="h-1.5 rounded-full transition-all duration-1000"
            :class="getProgressClass(alert.level)"
            :style="{ width: `${alert.progress}%` }"
          ></div>
        </div>
      </div>
    </div>
    
    <!-- 空状态 -->
    <div v-if="alerts.length === 0" class="text-center py-8">
      <ShieldCheck class="w-12 h-12 text-gray-600 mx-auto mb-3" />
      <p class="text-gray-500">暂无系统告警</p>
      <p class="text-xs text-gray-600 mt-1">系统运行正常</p>
    </div>
    
    <!-- 统计信息 -->
    <div v-if="alerts.length > 0" class="border-t border-gray-700 pt-3 mt-4">
      <div class="flex justify-between text-xs text-gray-400">
        <span>共 {{ alerts.length }} 条告警</span>
        <div class="flex space-x-4">
          <span v-if="highAlerts > 0" class="text-danger-400">
            高危: {{ highAlerts }}
          </span>
          <span v-if="mediumAlerts > 0" class="text-warning-400">
            中危: {{ mediumAlerts }}
          </span>
          <span v-if="lowAlerts > 0" class="text-success-400">
            低危: {{ lowAlerts }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, Check, X, ExternalLink, ShieldCheck } from 'lucide-vue-next'

interface Alert {
  id: number
  level: 'high' | 'medium' | 'low'
  title: string
  description: string
  time: string
  status?: string
  progress?: number
}

interface Props {
  alerts: Alert[]
}

const props = defineProps<Props>()

defineEmits<{
  'alert-click': [alert: Alert]
  'resolve-alert': [alert: Alert]
  'dismiss-alert': [alert: Alert]
  'view-detail': [alert: Alert]
}>()

// 计算告警统计
const highAlerts = computed(() => props.alerts.filter(a => a.level === 'high').length)
const mediumAlerts = computed(() => props.alerts.filter(a => a.level === 'medium').length)
const lowAlerts = computed(() => props.alerts.filter(a => a.level === 'low').length)

// 获取告警卡片样式
const getAlertClass = (level: Alert['level']) => {
  const classes = {
    high: 'border-danger-500/50 hover:border-danger-400/70 hover:bg-danger-500/5',
    medium: 'border-warning-500/50 hover:border-warning-400/70 hover:bg-warning-500/5',
    low: 'border-success-500/50 hover:border-success-400/70 hover:bg-success-500/5'
  }
  return classes[level]
}

// 获取指示器样式
const getIndicatorClass = (level: Alert['level']) => {
  const classes = {
    high: 'bg-danger-500',
    medium: 'bg-warning-500',
    low: 'bg-success-500'
  }
  return classes[level]
}

// 获取级别标签样式
const getLevelBadgeClass = (level: Alert['level']) => {
  const classes = {
    high: 'bg-danger-500/20 text-danger-400 border border-danger-500/30',
    medium: 'bg-warning-500/20 text-warning-400 border border-warning-500/30',
    low: 'bg-success-500/20 text-success-400 border border-success-500/30'
  }
  return classes[level]
}

// 获取进度条样式
const getProgressClass = (level: Alert['level']) => {
  const classes = {
    high: 'bg-gradient-to-r from-danger-600 to-danger-400',
    medium: 'bg-gradient-to-r from-warning-600 to-warning-400',
    low: 'bg-gradient-to-r from-success-600 to-success-400'
  }
  return classes[level]
}

// 获取级别文本
const getLevelText = (level: Alert['level']) => {
  const texts = {
    high: '高危',
    medium: '中危',
    low: '低危'
  }
  return texts[level]
}
</script>

<style scoped>
.alert-card {
  @apply p-4 rounded-lg border backdrop-blur-sm cursor-pointer transition-all duration-300;
  background: rgba(255, 255, 255, 0.1);
}

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
</style>