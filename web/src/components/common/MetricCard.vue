<template>
  <div 
    class="metric-card group cursor-pointer"
    :class="statusClasses"
    @click="$emit('click')"
  >
    <!-- 图标和主要数值 -->
    <div class="flex items-start justify-between mb-4">
      <div class="flex items-center space-x-3">
        <div 
          class="w-12 h-12 rounded-xl flex items-center justify-center transition-all duration-300"
          :class="iconBgClass"
        >
          <component :is="iconComponent" class="w-6 h-6" :class="iconColorClass" />
        </div>
        <div>
          <h3 class="text-sm font-medium text-gray-300">{{ title }}</h3>
          <div class="flex items-baseline space-x-2 mt-1">
            <span class="text-2xl font-bold" :class="valueColorClass">
              {{ displayValue }}
            </span>
            <span class="text-sm text-gray-400">{{ unit }}</span>
          </div>
        </div>
      </div>
      
      <!-- 趋势指标 -->
      <div v-if="trend" class="text-right">
        <span 
          class="text-sm font-medium flex items-center"
          :class="trendColorClass"
        >
          <component :is="trendIcon" class="w-4 h-4 mr-1" />
          {{ trend }}
        </span>
        <span class="text-xs text-gray-500">较上期</span>
      </div>
    </div>
    
    <!-- 进度条（如果有总数） -->
    <div v-if="total !== undefined" class="space-y-2">
      <div class="flex justify-between text-sm">
        <span class="text-gray-400">{{ value }}/{{ total }}</span>
        <span class="text-gray-400">{{ percentage }}%</span>
      </div>
      <div class="w-full bg-dark-700 rounded-full h-2">
        <div 
          class="h-2 rounded-full transition-all duration-1000 ease-out"
          :class="progressColorClass"
          :style="{ width: `${percentage}%` }"
        ></div>
      </div>
    </div>
    
    <!-- 状态指示器 -->
    <div class="flex items-center justify-between mt-4">
      <div class="flex items-center space-x-2">
        <div class="status-indicator" :class="statusIndicatorClass"></div>
        <span class="text-xs text-gray-400">{{ statusText }}</span>
      </div>
      
      <!-- 快速操作 -->
      <div class="flex space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <button 
          v-if="showRefresh"
          @click.stop="$emit('refresh')"
          class="p-1 hover:bg-white/10 rounded transition-colors"
        >
          <RotateCw class="w-3 h-3" />
        </button>
        <button 
          @click.stop="$emit('detail')"
          class="p-1 hover:bg-white/10 rounded transition-colors"
        >
          <ExternalLink class="w-3 h-3" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { 
  Server, 
  Box, 
  AlertTriangle, 
  Activity, 
  TrendingUp, 
  TrendingDown, 
  Minus,
  RotateCw,
  ExternalLink 
} from 'lucide-vue-next'

interface Props {
  title: string
  value: number
  total?: number
  unit?: string
  icon: string
  status?: 'success' | 'warning' | 'error' | 'info' | 'unknown'
  trend?: string
  showRefresh?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  unit: '',
  status: 'info',
  showRefresh: false
})

defineEmits<{
  click: []
  refresh: []
  detail: []
}>()

// 图标映射
const iconMap = {
  Server,
  Box,
  AlertTriangle,
  Activity
}

const iconComponent = computed(() => iconMap[props.icon as keyof typeof iconMap] || Activity)

// 计算百分比
const percentage = computed(() => {
  if (props.total === undefined) return 0
  return Math.round((props.value / props.total) * 100)
})

// 显示值格式化
const displayValue = computed(() => {
  if (props.value >= 1000000) {
    return `${(props.value / 1000000).toFixed(1)}M`
  }
  if (props.value >= 1000) {
    return `${(props.value / 1000).toFixed(1)}K`
  }
  return props.value.toString()
})

// 趋势图标
const trendIcon = computed(() => {
  if (!props.trend) return Minus
  const isPositive = props.trend.startsWith('+')
  const isNegative = props.trend.startsWith('-')
  
  if (isPositive) return TrendingUp
  if (isNegative) return TrendingDown
  return Minus
})

// 样式类
const statusClasses = computed(() => {
  const classes = {
    success: 'hover:glow-green',
    warning: 'hover:glow-yellow', 
    error: 'hover:glow-red',
    info: 'hover:glow-blue',
    unknown: ''
  }
  return classes[props.status]
})

const iconBgClass = computed(() => {
  const classes = {
    success: 'bg-success-500/20 group-hover:bg-success-500/30',
    warning: 'bg-warning-500/20 group-hover:bg-warning-500/30',
    error: 'bg-danger-500/20 group-hover:bg-danger-500/30',
    info: 'bg-primary-500/20 group-hover:bg-primary-500/30',
    unknown: 'bg-dark-600 group-hover:bg-dark-500'
  }
  return classes[props.status]
})

const iconColorClass = computed(() => {
  const classes = {
    success: 'text-success-400',
    warning: 'text-warning-400',
    error: 'text-danger-400',
    info: 'text-primary-400',
    unknown: 'text-gray-400'
  }
  return classes[props.status]
})

const valueColorClass = computed(() => {
  const classes = {
    success: 'text-success-400',
    warning: 'text-warning-400',
    error: 'text-danger-400',
    info: 'text-primary-400',
    unknown: 'text-white'
  }
  return classes[props.status]
})

const progressColorClass = computed(() => {
  const classes = {
    success: 'bg-gradient-to-r from-success-600 to-success-400',
    warning: 'bg-gradient-to-r from-warning-600 to-warning-400',
    error: 'bg-gradient-to-r from-danger-600 to-danger-400',
    info: 'bg-gradient-to-r from-primary-600 to-primary-400',
    unknown: 'bg-gradient-to-r from-gray-600 to-gray-400'
  }
  return classes[props.status]
})

const statusIndicatorClass = computed(() => {
  const classes = {
    success: 'status-online',
    warning: 'status-warning',
    error: 'status-error',
    info: 'status-online',
    unknown: 'status-offline'
  }
  return classes[props.status]
})

const statusText = computed(() => {
  const texts = {
    success: '运行正常',
    warning: '需要关注',
    error: '存在问题',
    info: '运行中',
    unknown: '状态未知'
  }
  return texts[props.status]
})

const trendColorClass = computed(() => {
  if (!props.trend) return 'text-gray-400'
  
  const isPositive = props.trend.startsWith('+')
  const isNegative = props.trend.startsWith('-')
  
  if (isPositive) {
    return props.status === 'error' ? 'text-danger-400' : 'text-success-400'
  }
  if (isNegative) {
    return props.status === 'error' ? 'text-success-400' : 'text-danger-400'
  }
  return 'text-gray-400'
})
</script>