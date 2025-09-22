<template>
  <!-- 修复的响应式布局模态框，支持主题切换并防止内容溢出 -->
  <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" v-if="visible" @click="handleBackdropClick">
    <div 
      :class="[
        themeClasses.modal,
        'rounded-lg border w-full max-w-7xl max-h-[95vh] overflow-hidden flex flex-col shadow-2xl'
      ]" 
      @click.stop
    >
      <!-- 模态框头部 - 支持主题切换 -->
      <div :class="['flex items-center justify-between p-6 border-b', themeClasses.border]">
        <div>
          <h2 :class="['text-xl font-semibold', themeClasses.text.primary]">Pod资源使用趋势</h2>
          <p :class="['text-sm mt-1', themeClasses.text.muted]" v-if="trendData">
            {{ trendData.pod_info.cluster_name }} / {{ trendData.pod_info.namespace }} / {{ trendData.pod_info.pod_name }}
          </p>
        </div>
        <div class="flex items-center gap-3">
          <!-- 主题切换按钮 -->
          <button
            @click="toggleTheme"
            :class="['p-2 rounded-lg transition-colors', themeClasses.hoverBg]"
            :title="`切换到${getNextThemeDisplayName()}`"
          >
            <component :is="themeIcon" class="w-5 h-5" :class="themeClasses.text.muted" />
          </button>
          <!-- 时间范围选择 -->
          <select v-model="selectedHours" @change="loadTrendData" :class="['input-field text-sm', themeClasses.select]">
            <option value="6">最近6小时</option>
            <option value="24">最近24小时</option>
            <option value="72">最近3天</option>
            <option value="168">最近7天</option>
          </select>
          <button @click="close" :class="['p-2 rounded-lg transition-colors', themeClasses.hoverBg]">
            <X class="w-5 h-5" :class="themeClasses.text.muted" />
          </button>
        </div>
      </div>

      <!-- 模态框内容 - 优化滚动和溢出控制 -->
      <div class="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-600 scrollbar-track-transparent">
        <div class="p-6 space-y-6">
          <!-- 加载状态 -->
          <div v-if="loading" class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
            <span :class="['ml-3', themeClasses.text.muted]">正在加载趋势数据...</span>
          </div>

          <!-- 错误状态 -->
          <div v-else-if="error" class="text-center py-12">
            <AlertTriangle class="w-12 h-12 text-danger-400 mx-auto mb-4" />
            <h3 class="text-lg font-semibold text-danger-400 mb-2">加载失败</h3>
            <p :class="['mb-4', themeClasses.text.muted]">{{ error }}</p>
            <button @click="loadTrendData" class="btn-secondary">重新加载</button>
          </div>

          <!-- 趋势内容 -->
          <div v-else-if="trendData" class="space-y-6">
            <!-- 时间范围信息 -->
            <div :class="['p-4 rounded-lg', themeClasses.card]">
              <div class="flex items-center justify-between">
                <h3 :class="['text-lg font-semibold', themeClasses.text.primary]">分析周期</h3>
                <span :class="['text-sm', themeClasses.text.muted]">{{ trendData.time_range.duration }}</span>
              </div>
              <p :class="['text-sm mt-1', themeClasses.text.muted]">
                {{ formatDate(trendData.time_range.start_time) }} 至 {{ formatDate(trendData.time_range.end_time) }}
              </p>
            </div>

            <!-- 统计概览 - 响应式网格布局 -->
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              <!-- CPU统计 -->
              <div :class="['p-4 rounded-lg', themeClasses.card]">
                <h4 :class="['text-sm font-medium mb-2', themeClasses.text.secondary]">CPU平均使用率</h4>
                <p class="text-xl font-bold text-warning-400">{{ trendData.cpu_trend.statistics.average.toFixed(1) }}%</p>
                <p :class="['text-xs mt-1', themeClasses.text.muted]">峰值: {{ trendData.cpu_trend.statistics.peak.toFixed(1) }}%</p>
              </div>
              <div :class="['p-4 rounded-lg', themeClasses.card]">
                <h4 :class="['text-sm font-medium mb-2', themeClasses.text.secondary]">CPU最低使用率</h4>
                <p class="text-xl font-bold text-info-400">{{ trendData.cpu_trend.statistics.minimum.toFixed(1) }}%</p>
                <p :class="['text-xs mt-1', themeClasses.text.muted]">方差: {{ trendData.cpu_trend.statistics.variance.toFixed(2) }}</p>
              </div>
              <!-- 内存统计 -->
              <div :class="['p-4 rounded-lg', themeClasses.card]">
                <h4 :class="['text-sm font-medium mb-2', themeClasses.text.secondary]">内存平均使用率</h4>
                <p class="text-xl font-bold text-primary-400">{{ trendData.memory_trend.statistics.average.toFixed(1) }}%</p>
                <p :class="['text-xs mt-1', themeClasses.text.muted]">峰值: {{ trendData.memory_trend.statistics.peak.toFixed(1) }}%</p>
              </div>
              <div :class="['p-4 rounded-lg', themeClasses.card]">
                <h4 :class="['text-sm font-medium mb-2', themeClasses.text.secondary]">内存最低使用率</h4>
                <p class="text-xl font-bold text-success-400">{{ trendData.memory_trend.statistics.minimum.toFixed(1) }}%</p>
                <p :class="['text-xs mt-1', themeClasses.text.muted]">方差: {{ trendData.memory_trend.statistics.variance.toFixed(2) }}</p>
              </div>
            </div>

            <!-- CPU趋势图表 - 修复布局约束 -->
            <div :class="['p-6 rounded-lg', themeClasses.card]">
              <h3 :class="['text-lg font-semibold mb-4 flex items-center', themeClasses.text.primary]">
                <Cpu class="w-5 h-5 mr-2 text-warning-400" />
                CPU使用率趋势
              </h3>
              <!-- 增强的图表容器 - 固定高度并防止溢出 -->
              <div class="h-80 w-full relative overflow-hidden rounded-lg border" :class="[themeClasses.chartContainer, themeClasses.border]">
                <div class="absolute inset-4">
                  <ResourceTrendChart
                    :data="cpuChartData"
                    :options="cpuChartOptions"
                    type="cpu"
                    :theme="currentTheme"
                  />
                </div>
              </div>
            </div>

            <!-- 内存趋势图表 - 修复布局约束 -->
            <div :class="['p-6 rounded-lg', themeClasses.card]">
              <h3 :class="['text-lg font-semibold mb-4 flex items-center', themeClasses.text.primary]">
                <Database class="w-5 h-5 mr-2 text-primary-400" />
                内存使用率趋势
              </h3>
              <!-- 增强的图表容器 - 固定高度并防止溢出 -->
              <div class="h-80 w-full relative overflow-hidden rounded-lg border" :class="[themeClasses.chartContainer, themeClasses.border]">
                <div class="absolute inset-4">
                  <ResourceTrendChart
                    :data="memoryChartData"
                    :options="memoryChartOptions"
                    type="memory"
                    :theme="currentTheme"
                  />
                </div>
              </div>
            </div>

            <!-- 事件标记 - 修复布局约束 -->
            <div :class="['p-6 rounded-lg', themeClasses.card]" v-if="trendData.event_markers.length > 0">
              <h3 :class="['text-lg font-semibold mb-4 flex items-center', themeClasses.text.primary]">
                <AlertTriangle class="w-5 h-5 mr-2 text-danger-400" />
                重要事件
              </h3>
              <!-- 事件列表容器 - 防止水平溢出 -->
              <div class="space-y-3 max-w-full overflow-hidden">
                <div v-for="event in trendData.event_markers" :key="event.timestamp" 
                     :class="['flex items-center p-3 rounded-lg border w-full', themeClasses.eventItem, themeClasses.border]">
                  <div :class="getEventIconClass(event.severity)" class="w-3 h-3 rounded-full mr-3 flex-shrink-0"></div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center justify-between gap-2">
                      <h4 :class="['text-sm font-medium truncate', themeClasses.text.primary]">{{ event.description }}</h4>
                      <span :class="['px-2 py-1 text-xs rounded-full whitespace-nowrap flex-shrink-0', getEventSeverityClass(event.severity)]">
                        {{ event.severity }}
                      </span>
                    </div>
                    <p :class="['text-xs mt-1 truncate', themeClasses.text.muted]">
                      {{ formatDate(event.timestamp) }} · {{ event.event_type }}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 模态框底部 -->
      <div :class="['flex items-center justify-end gap-3 p-6 border-t', themeClasses.border]">
        <button @click="exportData" class="btn-secondary" v-if="trendData">
          <Download class="w-4 h-4 mr-2" />
          导出数据
        </button>
        <button @click="close" class="btn-primary">关闭</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { 
  X, 
  AlertTriangle, 
  Database, 
  Cpu, 
  Download,
  Sun,
  Moon,
  Monitor
} from 'lucide-vue-next'
import PodsApiService, { type PodTrendData } from '../../api/pods'
import ResourceTrendChart from '../charts/ResourceTrendChart.vue'
import { useTheme } from '../../composables/useTheme'

// Props
interface Props {
  visible: boolean
  cluster?: string
  namespace?: string
  podName?: string
}

const props = defineProps<Props>()

// Emits
const emit = defineEmits<{
  close: []
}>()

// 主题系统集成
const { currentTheme, toggleTheme, getThemeDisplayName, getThemeIcon } = useTheme()

// 主题样式类计算
const themeClasses = computed(() => {
  const isDark = currentTheme.value === 'dark'
  
  return {
    modal: isDark ? 'bg-dark-800' : 'bg-white',
    border: isDark ? 'border-gray-700' : 'border-gray-200',
    card: isDark ? 'bg-dark-800/50 border-gray-700' : 'bg-white border-gray-200',
    chartContainer: isDark ? 'bg-dark-900/50' : 'bg-gray-50',
    eventItem: isDark ? 'bg-gray-800/50' : 'bg-gray-50',
    hoverBg: isDark ? 'hover:bg-gray-700' : 'hover:bg-gray-100',
    select: isDark ? 'bg-dark-700 border-gray-600 text-gray-200' : 'bg-white border-gray-300 text-gray-900',
    text: {
      primary: isDark ? 'text-white' : 'text-gray-900',
      secondary: isDark ? 'text-gray-300' : 'text-gray-700',
      muted: isDark ? 'text-gray-400' : 'text-gray-500',
      accent: isDark ? 'text-primary-400' : 'text-blue-600'
    }
  }
})

// 主题图标映射
const themeIcon = computed(() => {
  const iconMap = {
    dark: Sun,
    light: Moon,
    auto: Monitor
  }
  return iconMap[currentTheme.value] || Monitor
})

// 获取下一个主题的显示名称
const getNextThemeDisplayName = () => {
  const nextTheme = currentTheme.value === 'dark' ? 'light' : 'dark'
  return getThemeDisplayName(nextTheme)
}

// 响应式数据
const loading = ref(false)
const error = ref<string | null>(null)
const trendData = ref<PodTrendData | null>(null)
const selectedHours = ref(24)

// 监听props变化
watch([() => props.visible, () => props.cluster, () => props.namespace, () => props.podName], 
  ([visible, cluster, namespace, podName]) => {
    if (visible && cluster && namespace && podName) {
      loadTrendData()
    }
  }, 
  { immediate: true }
)

// 加载趋势数据
const loadTrendData = async () => {
  if (!props.cluster || !props.namespace || !props.podName) return

  try {
    loading.value = true
    error.value = null
    
    const response = await PodsApiService.getPodTrendData(
      props.cluster, 
      props.namespace, 
      props.podName,
      selectedHours.value
    )
    
    if (response.code === 0 && response.data) {
      trendData.value = response.data
    } else {
      throw new Error(response.message || '获取趋势数据失败')
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
    console.error('加载Pod趋势数据失败:', err)
  } finally {
    loading.value = false
  }
}

// CPU图表数据
const cpuChartData = computed(() => {
  if (!trendData.value) return null
  
  return {
    labels: trendData.value.cpu_trend.data_points.map(point => {
      return new Date(point.timestamp).toLocaleTimeString('zh-CN', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    }),
    datasets: [
      {
        label: 'CPU使用率',
        data: trendData.value.cpu_trend.data_points.map(point => point.usage),
        borderColor: currentTheme.value === 'dark' ? '#fb7185' : '#f59e0b',
        backgroundColor: currentTheme.value === 'dark' ? 'rgba(251, 113, 133, 0.1)' : 'rgba(245, 158, 11, 0.1)',
        tension: 0.4,
        fill: true
      }
    ]
  }
})

// 内存图表数据
const memoryChartData = computed(() => {
  if (!trendData.value) return null
  
  return {
    labels: trendData.value.memory_trend.data_points.map(point => {
      return new Date(point.timestamp).toLocaleTimeString('zh-CN', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    }),
    datasets: [
      {
        label: '内存使用率',
        data: trendData.value.memory_trend.data_points.map(point => point.usage),
        borderColor: '#3b82f6',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        tension: 0.4,
        fill: true
      }
    ]
  }
})

// 图表配置 - 动态主题支持
const chartOptions = computed(() => {
  const isDark = currentTheme.value === 'dark'
  const textColor = isDark ? '#9ca3af' : '#6b7280'
  const gridColor = isDark ? 'rgba(156, 163, 175, 0.1)' : 'rgba(229, 231, 235, 0.5)'
  
  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: true,
        labels: {
          color: textColor
        }
      }
    },
    scales: {
      x: {
        ticks: {
          color: textColor
        },
        grid: {
          color: gridColor
        }
      },
      y: {
        beginAtZero: true,
        max: 100,
        ticks: {
          color: textColor,
          callback: function(value: any) {
            return value + '%'
          }
        },
        grid: {
          color: gridColor
        }
      }
    }
  }
})

const cpuChartOptions = computed(() => chartOptions.value)
const memoryChartOptions = computed(() => chartOptions.value)

// 关闭模态框
const close = () => {
  emit('close')
  trendData.value = null
  error.value = null
}

// 导出数据
const exportData = () => {
  if (!trendData.value) return
  
  const data = {
    pod_info: trendData.value.pod_info,
    time_range: trendData.value.time_range,
    cpu_data: trendData.value.cpu_trend.data_points,
    memory_data: trendData.value.memory_trend.data_points,
    statistics: {
      cpu: trendData.value.cpu_trend.statistics,
      memory: trendData.value.memory_trend.statistics
    },
    events: trendData.value.event_markers
  }
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `pod-trend-${props.cluster}-${props.namespace}-${props.podName}-${new Date().toISOString().split('T')[0]}.json`
  a.click()
  URL.revokeObjectURL(url)
}

// 处理背景点击
const handleBackdropClick = (event: MouseEvent) => {
  if (event.target === event.currentTarget) {
    close()
  }
}

// 工具方法
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

// 事件图标样式 - 支持主题适配
const getEventIconClass = (severity: string) => {
  switch (severity) {
    case 'critical':
      return 'bg-danger-500'
    case 'warning':
      return 'bg-warning-500'
    case 'info':
      return 'bg-info-500'
    default:
      return 'bg-gray-500'
  }
}

// 事件严重程度样式 - 支持主题适配
const getEventSeverityClass = (severity: string) => {
  const isDark = currentTheme.value === 'dark'
  
  switch (severity) {
    case 'critical':
      return isDark ? 'bg-danger-500/20 text-danger-400' : 'bg-danger-100 text-danger-700'
    case 'warning':
      return isDark ? 'bg-warning-500/20 text-warning-400' : 'bg-warning-100 text-warning-700'
    case 'info':
      return isDark ? 'bg-info-500/20 text-info-400' : 'bg-info-100 text-info-700'
    default:
      return isDark ? 'bg-gray-500/20 text-gray-400' : 'bg-gray-100 text-gray-700'
  }
}
</script>

<style scoped>
/* 自定义滚动条样式 - 支持主题适配 */
.scrollbar-thin {
  scrollbar-width: thin;
}

.scrollbar-thumb-gray-600 {
  scrollbar-color: #4b5563 transparent;
}

.scrollbar-track-transparent {
  scrollbar-track-color: transparent;
}

/* Webkit滚动条样式 */
.scrollbar-thin::-webkit-scrollbar {
  width: 6px;
}

.scrollbar-thin::-webkit-scrollbar-track {
  background: transparent;
}

.scrollbar-thin::-webkit-scrollbar-thumb {
  background-color: #4b5563;
  border-radius: 3px;
  border: 2px solid transparent;
}

.scrollbar-thin::-webkit-scrollbar-thumb:hover {
  background-color: #6b7280;
}

/* 响应式文本截断 */
.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 动画优化 */
@keyframes fade-in {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.animate-fade-in {
  animation: fade-in 0.2s ease-out;
}

/* 确保内容不会水平溢出 */
.max-w-full {
  max-width: 100%;
}

.min-w-0 {
  min-width: 0;
}

/* 弹性布局改进 */
.flex-shrink-0 {
  flex-shrink: 0;
}

/* 响应式间距 */
@media (max-width: 768px) {
  .grid-cols-1.sm\\:grid-cols-2.lg\\:grid-cols-4 {
    grid-template-columns: repeat(1, minmax(0, 1fr));
  }
}

@media (min-width: 768px) {
  .grid-cols-1.sm\\:grid-cols-2.lg\\:grid-cols-4 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .grid-cols-1.sm\\:grid-cols-2.lg\\:grid-cols-4 {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>