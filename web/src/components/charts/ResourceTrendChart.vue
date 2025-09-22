<template>
  <div class="w-full h-full relative min-h-0">
    <!-- 加载状态 -->
    <div v-if="!data || !data.labels || data.labels.length === 0" class="flex items-center justify-center h-full">
      <div class="text-center text-gray-400">
        <div class="animate-pulse">暂无图表数据</div>
      </div>
    </div>
    
    <!-- SVG 折线图 - 使用完整容器空间并确保不溢出 -->
    <svg v-else class="w-full h-full block" viewBox="0 0 840 420" preserveAspectRatio="xMidYMid meet">
      <!-- 背景网格 -->
      <defs>
        <pattern id="grid" width="60" height="30" patternUnits="userSpaceOnUse">
          <path d="M 60 0 L 0 0 0 30" fill="none" :stroke="themeConfig.gridColor" stroke-width="0.5" opacity="0.3"/>
        </pattern>
        <!-- 渐变定义 -->
        <linearGradient :id="`gradient-${type}`" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" :stop-color="gradientColor" stop-opacity="0.3"/>
          <stop offset="100%" :stop-color="gradientColor" stop-opacity="0.05"/>
        </linearGradient>
      </defs>
      <rect width="100%" height="100%" fill="url(#grid)" />
      
      <!-- Y轴 -->
      <line x1="80" y1="60" x2="80" y2="360" :stroke="themeConfig.gridColor" stroke-width="1"/>
      
      <!-- X轴 -->
      <line x1="80" y1="360" x2="760" y2="360" :stroke="themeConfig.gridColor" stroke-width="1"/>
      
      <!-- Y轴标签 -->
      <g class="y-axis">
        <text v-for="(label, index) in yAxisLabels" 
              :key="index"
              :x="70" 
              :y="360 - index * 60"
              text-anchor="end"
              class="text-xs"
              :fill="themeConfig.textColor"
              dominant-baseline="middle">
          {{ label }}%
        </text>
        <!-- Y轴网格线 -->
        <line v-for="(label, index) in yAxisLabels" 
              :key="`grid-y-${index}`"
              :x1="80" 
              :x2="760"
              :y1="360 - index * 60"
              :y2="360 - index * 60"
              :stroke="themeConfig.gridColor" 
              stroke-width="0.5" 
              opacity="0.3"/>
      </g>
      
      <!-- X轴标签 -->
      <g class="x-axis">
        <text v-for="(label, index) in displayLabels" 
              :key="index"
              :x="80 + index * xStep" 
              y="380"
              text-anchor="middle"
              class="text-xs"
              :fill="themeConfig.textColor">
          {{ label }}
        </text>
        <!-- X轴网格线 -->
        <line v-for="(label, index) in displayLabels" 
              :key="`grid-x-${index}`"
              :x1="80 + index * xStep" 
              :x2="80 + index * xStep"
              y1="60"
              y2="360"
              :stroke="themeConfig.gridColor" 
              stroke-width="0.5" 
              opacity="0.2"/>
      </g>
      
      <!-- 面积填充 -->
      <path
        v-if="areaPath"
        :d="areaPath"
        :fill="`url(#gradient-${type})`"
        class="animate-draw"
      />
      
      <!-- 趋势线 -->
      <path
        v-if="linePath"
        :d="linePath"
        fill="none"
        :stroke="lineColor"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="animate-draw"
      />
      
      <!-- 数据点 -->
      <g class="data-points">
        <circle
          v-for="(value, index) in chartData"
          :key="`point-${index}`"
          :cx="80 + index * xStep"
          :cy="getYPosition(value)"
          r="4"
          :fill="pointColor"
          stroke="#1f2937"
          stroke-width="2"
          class="hover:r-6 transition-all cursor-pointer animate-fade-in"
          :style="{ animationDelay: `${index * 50}ms` }"
          @mouseenter="showTooltip($event, index, value)"
          @mouseleave="hideTooltip"
        />
      </g>
      
      <!-- 峰值标记 -->
      <g v-if="showPeakMarker" class="peak-markers">
        <circle
          :cx="80 + peakIndex * xStep"
          :cy="getYPosition(peakValue)"
          r="6"
          fill="none"
          :stroke="lineColor"
          stroke-width="2"
          class="animate-pulse"
        />
        <text
          :x="80 + peakIndex * xStep"
          :y="getYPosition(peakValue) - 15"
          text-anchor="middle"
          class="text-xs fill-danger-400 font-medium">
          峰值 {{ peakValue.toFixed(1) }}%
        </text>
      </g>
    </svg>
    
    <!-- 图例 -->
    <div v-if="showLegend" class="absolute top-4 right-4 bg-white/10 backdrop-blur-sm rounded-lg p-3 border border-white/20">
      <div class="flex items-center space-x-2">
        <div class="w-3 h-0.5" :style="{ backgroundColor: lineColor }"></div>
        <span class="text-xs text-gray-300">{{ legendLabel }}</span>
      </div>
    </div>
    
    <!-- 工具提示 - 支持主题适配 -->
    <div
      v-if="tooltip.visible"
      :class="[
        'absolute text-xs px-3 py-2 rounded-lg border pointer-events-none z-20 shadow-lg backdrop-blur-sm',
        props.theme === 'dark' 
          ? 'bg-dark-900/95 text-white border-gray-600' 
          : 'bg-white/95 text-gray-900 border-gray-300'
      ]"
      :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
    >
      <div class="font-medium">{{ tooltip.time }}</div>
      <div :class="props.theme === 'dark' ? 'text-primary-300' : 'text-primary-600'">
        {{ tooltip.label }}: {{ tooltip.value.toFixed(1) }}%
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

// 支持Chart.js数据格式的接口定义
interface ChartData {
  labels: string[]
  datasets: Array<{
    label: string
    data: number[]
    borderColor?: string
    backgroundColor?: string
    tension?: number
    fill?: boolean
  }>
}

interface Props {
  data: ChartData | null
  options?: any
  type: 'cpu' | 'memory'
  theme?: string  // 新增主题属性
}

const props = withDefaults(defineProps<Props>(), {
  data: null,
  options: () => ({}),
  type: 'cpu',
  theme: 'dark'  // 默认暗色主题
})

// 工具提示状态
const tooltip = ref({
  visible: false,
  x: 0,
  y: 0,
  time: '',
  label: '',
  value: 0
})

// 从props.data中提取图表数据
const chartData = computed(() => {
  if (!props.data || !props.data.datasets || props.data.datasets.length === 0) {
    return []
  }
  return props.data.datasets[0].data
})

const chartLabels = computed(() => {
  if (!props.data || !props.data.labels) {
    return []
  }
  return props.data.labels
})

// 根据数据点数量决定显示的标签密度
const displayLabels = computed(() => {
  const labels = chartLabels.value
  if (labels.length <= 8) {
    return labels
  }
  
  // 当数据点过多时，只显示关键时间点
  const step = Math.ceil(labels.length / 6)
  return labels.filter((_, index) => index % step === 0 || index === labels.length - 1)
})

// 计算图表参数 - 适配新的更大画布尺寸
const xStep = computed(() => {
  const totalPoints = displayLabels.value.length
  return totalPoints > 1 ? 680 / (totalPoints - 1) : 0 // 680px是有效绘图宽度(760-80)
})

const yAxisLabels = computed(() => ['0', '20', '40', '60', '80', '100'])

// 颜色主题配置 - 支持动态主题切换
const themeConfig = computed(() => {
  const isDark = props.theme === 'dark'
  
  switch (props.type) {
    case 'cpu':
      return {
        lineColor: isDark ? '#fb7185' : '#f59e0b', // 暗色主题用玫瑰色，亮色主题用琥珀色
        pointColor: isDark ? '#f43f5e' : '#f59e0b',
        gradientColor: isDark ? '#fb7185' : '#f59e0b',
        legendLabel: 'CPU 使用率',
        gridColor: isDark ? '#374151' : '#e5e7eb',
        textColor: isDark ? '#9ca3af' : '#6b7280',
        backgroundColor: isDark ? '#111827' : '#f9fafb'
      }
    case 'memory':
      return {
        lineColor: '#3b82f6', // 内存统一使用蓝色
        pointColor: isDark ? '#2563eb' : '#1d4ed8',
        gradientColor: '#3b82f6',
        legendLabel: '内存使用率',
        gridColor: isDark ? '#374151' : '#e5e7eb',
        textColor: isDark ? '#9ca3af' : '#6b7280',
        backgroundColor: isDark ? '#111827' : '#f9fafb'
      }
    default:
      return {
        lineColor: '#10b981',
        pointColor: '#059669',
        gradientColor: '#10b981',
        legendLabel: '资源使用率',
        gridColor: isDark ? '#374151' : '#e5e7eb',
        textColor: isDark ? '#9ca3af' : '#6b7280',
        backgroundColor: isDark ? '#111827' : '#f9fafb'
      }
  }
})

const lineColor = computed(() => themeConfig.value.lineColor)
const pointColor = computed(() => themeConfig.value.pointColor)
const gradientColor = computed(() => themeConfig.value.gradientColor)
const legendLabel = computed(() => themeConfig.value.legendLabel)

// 获取 Y 轴位置（基于百分比）- 适配新的坐标系统
const getYPosition = (value: number) => {
  return 360 - (Math.min(Math.max(value, 0), 100) / 100) * 300 // Y轴范围从60到360，共300px高度
}

// 生成折线路径 - 使用新的坐标系统
const linePath = computed(() => {
  const data = chartData.value
  if (data.length === 0) return ''
  
  // 使用平滑曲线算法
  let path = `M 80 ${getYPosition(data[0])}`
  
  for (let i = 1; i < data.length; i++) {
    const x = 80 + i * (680 / (data.length - 1)) // 使用680px宽度
    const y = getYPosition(data[i])
    
    if (i === 1) {
      // 第一段直接连线
      path += ` L ${x} ${y}`
    } else {
      // 使用二次贝塞尔曲线创建平滑效果
      const prevX = 80 + (i - 1) * (680 / (data.length - 1))
      const prevY = getYPosition(data[i - 1])
      const cpX = (prevX + x) / 2
      
      path += ` Q ${cpX} ${prevY} ${x} ${y}`
    }
  }
  
  return path
})

// 生成面积路径 - 使用新的坐标系统
const areaPath = computed(() => {
  const data = chartData.value
  if (data.length === 0) return ''
  
  let path = linePath.value
  
  // 闭合路径形成面积
  const lastX = 80 + (data.length - 1) * (680 / (data.length - 1))
  path += ` L ${lastX} 360 L 80 360 Z` // 底部Y坐标改为360
  
  return path
})

// 峰值检测
const peakData = computed(() => {
  const data = chartData.value
  if (data.length === 0) return { index: 0, value: 0 }
  
  const maxValue = Math.max(...data)
  const maxIndex = data.indexOf(maxValue)
  
  return { index: maxIndex, value: maxValue }
})

const peakIndex = computed(() => peakData.value.index)
const peakValue = computed(() => peakData.value.value)
const showPeakMarker = computed(() => peakValue.value > 80) // 只在高峰值时显示标记
const showLegend = computed(() => true)

// 工具提示方法 - 优化定位以防止溢出
const showTooltip = (event: MouseEvent, dataIndex: number, value: number) => {
  const rect = (event.target as SVGElement).getBoundingClientRect()
  const container = (event.target as SVGElement).closest('.relative')!.getBoundingClientRect()
  
  // 计算在实际数据中的索引
  const actualIndex = Math.round(dataIndex * (chartLabels.value.length - 1) / (chartData.value.length - 1))
  const timeLabel = chartLabels.value[actualIndex] || chartLabels.value[dataIndex] || '未知时间'
  
  // 工具提示尺寸和边距约束
  const tooltipWidth = 140
  const tooltipHeight = 50
  const margin = 10
  
  // 计算X位置，确保不会超出容器边界
  let tooltipX = rect.left - container.left + rect.width / 2 - tooltipWidth / 2
  const maxX = container.width - tooltipWidth - margin
  tooltipX = Math.min(Math.max(tooltipX, margin), maxX)
  
  // 计算Y位置，确保不会超出容器边界
  let tooltipY = rect.top - container.top - tooltipHeight - margin
  if (tooltipY < margin) {
    tooltipY = rect.bottom - container.top + margin // 如果顶部空间不够，显示在下方
  }
  
  tooltip.value = {
    visible: true,
    x: tooltipX,
    y: tooltipY,
    time: timeLabel,
    label: legendLabel.value,
    value: value
  }
}

const hideTooltip = () => {
  tooltip.value.visible = false
}
</script>

<style scoped>
@keyframes draw {
  to {
    stroke-dashoffset: 0;
  }
}

.animate-draw {
  stroke-dasharray: 1000;
  stroke-dashoffset: 1000;
  animation: draw 2s ease-out forwards;
}

@keyframes fade-in {
  from {
    opacity: 0;
    transform: scale(0);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.animate-fade-in {
  opacity: 0;
  animation: fade-in 0.5s ease-out forwards;
}
</style>