<template>
  <div class="glass-card hover:shadow-lg transition-all duration-300" style="background: var(--card-bg); border: 1px solid var(--border-color);">
    <div class="p-6 border-b" style="border-color: var(--border-color);">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <div class="p-2 rounded-lg" style="background: var(--icon-bg-optimization);">
            <Settings class="w-5 h-5" style="color: var(--icon-color-optimization);" />
          </div>
          <h3 class="text-lg font-semibold" style="color: var(--text-primary);">活动数据优化</h3>
        </div>
        <button 
          class="btn-secondary text-sm transition-colors"
          style="background: var(--btn-secondary-bg); color: var(--btn-secondary-text); border: 1px solid var(--border-color);"
          @click="$emit('open-config')"
          :disabled="loading"
        >
          <Settings :class="['w-3 h-3 mr-1', { 'animate-spin': loading }]" />
          配置
        </button>
      </div>
    </div>
    <div class="p-4">
      <!-- 优化统计显示 -->
      <div v-if="optimizationResult" class="space-y-3 mb-4">
        <div class="text-sm" style="color: var(--text-secondary);">最近优化结果</div>
        <div class="grid grid-cols-2 gap-3 text-xs">
          <div class="p-2 rounded border" style="background: var(--item-bg); border-color: var(--border-color);">
            <div style="color: var(--text-muted);">去重</div>
            <div class="font-medium" style="color: var(--text-primary);">{{ optimizationResult.duplicates_removed || 0 }}条</div>
          </div>
          <div class="p-2 rounded border" style="background: var(--item-bg); border-color: var(--border-color);">
            <div style="color: var(--text-muted);">降噪</div>
            <div class="font-medium" style="color: var(--text-primary);">{{ optimizationResult.noise_filtered || 0 }}条</div>
          </div>
          <div class="p-2 rounded border" style="background: var(--item-bg); border-color: var(--border-color);">
            <div style="color: var(--text-muted);">聚合</div>
            <div class="font-medium" style="color: var(--text-primary);">{{ optimizationResult.aggregations?.length || 0 }}组</div>
          </div>
          <div class="p-2 rounded border" style="background: var(--item-bg); border-color: var(--border-color);">
            <div style="color: var(--text-muted);">处理时间</div>
            <div class="font-medium text-xs" style="color: var(--text-primary);">{{ formatOptimizationTime(optimizationResult.processed_at) }}</div>
          </div>
        </div>
      </div>
      
      <!-- 操作按钮 -->
      <div class="space-y-2">
        <div class="text-center p-3 rounded-lg border" style="background: var(--success-bg); border-color: var(--success-border);">
          <div class="flex items-center justify-center mb-1" style="color: var(--success-color);">
            <svg class="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path>
            </svg>
            <span class="text-sm font-medium">智能优化</span>
          </div>
          <div class="text-xs" style="color: var(--success-color);">
            系统已启用自动去重和降噪功能
          </div>
        </div>
        <div class="text-xs text-center" style="color: var(--text-muted);">
          数据刷新时自动执行去重、降噪和聚合操作
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Settings } from 'lucide-vue-next'

// 定义优化结果接口类型
interface OptimizationResult {
  duplicates_removed: number
  noise_filtered: number
  aggregations?: Array<any>
  processed_at: string
}

// 定义props接口
interface Props {
  optimizationResult?: OptimizationResult | null
  loading?: boolean
}

// 定义事件接口
interface Emits {
  'open-config': []
}

// 接收props和定义emits
defineProps<Props>()
defineEmits<Emits>()

// 格式化优化时间的工具函数
const formatOptimizationTime = (timestamp: string) => {
  if (!timestamp) return '未知'
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / (1000 * 60))
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (minutes < 1440) return `${Math.floor(minutes / 60)}小时前`
  return `${Math.floor(minutes / 1440)}天前`
}
</script>