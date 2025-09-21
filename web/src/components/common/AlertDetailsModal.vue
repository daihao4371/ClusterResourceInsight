<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center">
    <!-- 背景遮罩 -->
    <div 
      class="absolute inset-0 backdrop-blur-sm"
      style="background-color: rgba(0, 0, 0, 0.5)"
      @click="$emit('close')"
    ></div>
    
    <!-- 模态框内容 -->
    <div class="relative border rounded-lg shadow-2xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-hidden"
         style="background-color: var(--bg-secondary); border-color: var(--border-color);">
      <!-- 模态框头部 -->
      <div class="flex items-center justify-between p-6 border-b"
           style="border-color: var(--border-color);">
        <div class="flex items-center space-x-3">
          <!-- 告警级别指示器 -->
          <div 
            :class="[
              'w-4 h-4 rounded-full',
              alert?.level === 'high' ? 'bg-error-500' : 
              alert?.level === 'medium' ? 'bg-warning-500' : 'bg-info-500'
            ]"
          ></div>
          <h2 class="text-xl font-semibold" style="color: var(--text-primary);">告警详情</h2>
        </div>
        
        <button 
          @click="$emit('close')"
          class="p-2 rounded-lg transition-colors hover:opacity-80"
          style="background-color: var(--bg-tertiary);"
        >
          <X class="w-5 h-5" style="color: var(--text-secondary);" />
        </button>
      </div>
      
      <!-- 模态框内容区域 -->
      <div class="p-6 overflow-y-auto max-h-[calc(90vh-120px)]">
        <div v-if="alert" class="space-y-6">
          <!-- 基本信息 -->
          <div class="space-y-4">
            <div>
              <h3 class="text-lg font-semibold mb-2" style="color: var(--text-primary);">{{ alert.title }}</h3>
              <p style="color: var(--text-secondary);">{{ alert.description }}</p>
            </div>
            
            <!-- 状态和级别信息 -->
            <div class="grid grid-cols-2 gap-4">
              <div class="glass-card p-4">
                <p class="text-sm mb-1" style="color: var(--text-muted);">告警级别</p>
                <div class="flex items-center space-x-2">
                  <div 
                    :class="[
                      'w-3 h-3 rounded-full',
                      alert.level === 'high' ? 'bg-error-500' : 
                      alert.level === 'medium' ? 'bg-warning-500' : 'bg-info-500'
                    ]"
                  ></div>
                  <span class="font-medium" style="color: var(--text-primary);">{{ levelMap[alert.level] || alert.level }}</span>
                </div>
              </div>
              
              <div class="glass-card p-4">
                <p class="text-sm mb-1" style="color: var(--text-muted);">当前状态</p>
                <span 
                  class="px-2 py-1 rounded-full text-xs font-medium"
                  :style="{
                    backgroundColor: alert.status === 'active' ? 'rgba(239, 68, 68, 0.2)' :
                                   alert.status === 'resolved' ? 'rgba(16, 185, 129, 0.2)' :
                                   'rgba(107, 114, 128, 0.2)',
                    color: alert.status === 'active' ? 'var(--error-color)' :
                           alert.status === 'resolved' ? 'var(--success-color)' :
                           'var(--text-muted)'
                  }"
                >
                  {{ statusMap[alert.status] || alert.status }}
                </span>
              </div>
            </div>
          </div>
          
          <!-- 时间信息 -->
          <div class="glass-card p-4">
            <h4 class="font-medium mb-3" style="color: var(--text-primary);">时间信息</h4>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <p style="color: var(--text-muted);">触发时间</p>
                <p style="color: var(--text-primary);">{{ alert.time }}</p>
              </div>
              <div v-if="alert.status === 'resolved'">
                <p style="color: var(--text-muted);">解决时间</p>
                <p style="color: var(--text-primary);">{{ alert.resolvedTime || '未知' }}</p>
              </div>
            </div>
          </div>
          
          <!-- 详细信息 -->
          <div class="glass-card p-4">
            <h4 class="font-medium mb-3" style="color: var(--text-primary);">详细信息</h4>
            <div class="space-y-3 text-sm">
              <div v-if="alert.clusterId">
                <p style="color: var(--text-muted);">关联集群</p>
                <p style="color: var(--text-primary);">{{ alert.clusterName || `集群 ID: ${alert.clusterId}` }}</p>
              </div>
              
              <div v-if="alert.source">
                <p style="color: var(--text-muted);">告警来源</p>
                <p style="color: var(--text-primary);">{{ getSourceLabel(alert.source) }}</p>
              </div>
              
              <div v-if="alert.details">
                <p style="color: var(--text-muted);">其他详情</p>
                <div class="rounded p-3 mt-1" style="background-color: var(--bg-tertiary);">
                  <pre class="text-xs whitespace-pre-wrap" style="color: var(--text-secondary);">{{ formatDetails(alert.details) }}</pre>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 建议操作 -->
          <div class="glass-card p-4">
            <h4 class="font-medium mb-3" style="color: var(--text-primary);">建议操作</h4>
            <div class="space-y-2 text-sm">
              <div v-if="alert.level === 'high'" class="flex items-start space-x-2">
                <AlertTriangle class="w-4 h-4 mt-0.5 flex-shrink-0" style="color: var(--error-color);" />
                <div>
                  <p class="font-medium" style="color: var(--text-primary);">紧急处理</p>
                  <p style="color: var(--text-muted);">此为高级告警，建议立即检查相关资源状态</p>
                </div>
              </div>
              
              <div v-if="alert.level === 'medium'" class="flex items-start space-x-2">
                <AlertCircle class="w-4 h-4 mt-0.5 flex-shrink-0" style="color: var(--warning-color);" />
                <div>
                  <p class="font-medium" style="color: var(--text-primary);">及时处理</p>
                  <p style="color: var(--text-muted);">建议在合适时间检查并处理此告警</p>
                </div>
              </div>
              
              <div v-if="alert.level === 'low'" class="flex items-start space-x-2">
                <Info class="w-4 h-4 mt-0.5 flex-shrink-0" style="color: var(--accent-secondary);" />
                <div>
                  <p class="font-medium" style="color: var(--text-primary);">监控观察</p>
                  <p style="color: var(--text-muted);">低级告警，可以关注但不必立即处理</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 模态框底部操作 -->
      <div class="flex items-center justify-end space-x-3 p-6 border-t"
           style="border-color: var(--border-color); background-color: var(--bg-tertiary); opacity: 0.8;">
        <button 
          @click="$emit('close')"
          class="btn-ghost"
        >
          关闭
        </button>
        
        <button 
          v-if="alert?.status === 'active'"
          @click="$emit('resolve', alert)"
          class="btn-primary"
        >
          标记为已解决
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { X, AlertTriangle, AlertCircle, Info } from 'lucide-vue-next'
import { useTheme } from '../../composables/useTheme'

// 使用主题系统
const { currentTheme } = useTheme()

// 定义组件属性
interface Props {
  visible: boolean
  alert: any
}

// 定义事件
interface Emits {
  (e: 'close'): void
  (e: 'resolve', alert: any): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 状态和级别映射
const statusMap = {
  active: '活跃',
  resolved: '已解决',
  suppressed: '已屏蔽'
}

const levelMap = {
  high: '高级',
  medium: '中级',
  low: '低级'
}

// 来源标签映射
const getSourceLabel = (source: string) => {
  const sourceMap = {
    collector: '数据收集器',
    monitor: '监控系统',
    scheduler: '调度器',
    api: 'API接口',
    system: '系统'
  }
  return sourceMap[source] || source
}

// 格式化详情信息
const formatDetails = (details: string | object) => {
  if (typeof details === 'string') {
    try {
      return JSON.stringify(JSON.parse(details), null, 2)
    } catch {
      return details
    }
  }
  return JSON.stringify(details, null, 2)
}

// 监听 ESC 键关闭模态框
watch(() => props.visible, (newVal) => {
  if (newVal) {
    document.addEventListener('keydown', handleEscape)
  } else {
    document.removeEventListener('keydown', handleEscape)
  }
})

const handleEscape = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    emit('close')
  }
}
</script>