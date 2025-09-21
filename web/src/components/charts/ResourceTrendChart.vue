<template>
  <div class="w-full h-80 relative">
    <!-- SVG 折线图 -->
    <svg class="w-full h-full" viewBox="0 0 600 300" preserveAspectRatio="xMidYMid meet">
      <!-- 背景网格 -->
      <defs>
        <pattern id="grid" width="50" height="25" patternUnits="userSpaceOnUse">
          <path d="M 50 0 L 0 0 0 25" fill="none" stroke="#374151" stroke-width="0.5" opacity="0.3"/>
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#grid)" />
      
      <!-- Y轴标签 -->
      <g class="y-axis">
        <text v-for="(label, index) in yAxisLabels" 
              :key="index"
              :x="30" 
              :y="260 - index * 52"
              text-anchor="end"
              class="text-xs fill-gray-400"
              dominant-baseline="middle">
          {{ label }}
        </text>
      </g>
      
      <!-- X轴标签 -->
      <g class="x-axis">
        <text v-for="(point, index) in data" 
              :key="index"
              :x="50 + index * xStep" 
              y="285"
              text-anchor="middle"
              class="text-xs fill-gray-400">
          {{ point.time }}
        </text>
      </g>
      
      <!-- CPU 使用率线 -->
      <path
        :d="cpuPath"
        fill="none"
        stroke="#3b82f6"
        stroke-width="2"
        class="animate-draw"
      />
      
      <!-- 内存使用率线 -->
      <path
        :d="memoryPath"
        fill="none"
        stroke="#10b981"
        stroke-width="2"
        class="animate-draw"
      />
      
      <!-- Pod 数量线 -->
      <path
        :d="podsPath"
        fill="none"
        stroke="#f59e0b"
        stroke-width="2"
        class="animate-draw"
      />
      
      <!-- 数据点 -->
      <g class="data-points">
        <!-- CPU 数据点 -->
        <circle
          v-for="(point, index) in data"
          :key="`cpu-${index}`"
          :cx="50 + index * xStep"
          :cy="getYPosition(point.cpu, 100)"
          r="4"
          fill="#3b82f6"
          class="hover:r-6 transition-all cursor-pointer animate-fade-in"
          :style="{ animationDelay: `${index * 100}ms` }"
          @mouseenter="showTooltip($event, 'CPU', point.cpu, '%')"
          @mouseleave="hideTooltip"
        />
        
        <!-- 内存数据点 -->
        <circle
          v-for="(point, index) in data"
          :key="`memory-${index}`"
          :cx="50 + index * xStep"
          :cy="getYPosition(point.memory, 100)"
          r="4"
          fill="#10b981"
          class="hover:r-6 transition-all cursor-pointer animate-fade-in"
          :style="{ animationDelay: `${index * 100 + 50}ms` }"
          @mouseenter="showTooltip($event, '内存', point.memory, '%')"
          @mouseleave="hideTooltip"
        />
        
        <!-- Pod 数据点 -->
        <circle
          v-for="(point, index) in data"
          :key="`pods-${index}`"
          :cx="50 + index * xStep"
          :cy="getYPosition(point.pods, maxPods)"
          r="4"
          fill="#f59e0b"
          class="hover:r-6 transition-all cursor-pointer animate-fade-in"
          :style="{ animationDelay: `${index * 100 + 100}ms` }"
          @mouseenter="showTooltip($event, 'Pods', point.pods, '个')"
          @mouseleave="hideTooltip"
        />
      </g>
    </svg>
    
    <!-- 图例 -->
    <div class="absolute top-4 right-4 bg-white/10 backdrop-blur-sm rounded-lg p-3 space-y-2 border border-white/20">
      <div class="flex items-center space-x-2">
        <div class="w-3 h-0.5 bg-blue-500"></div>
        <span class="text-xs text-gray-300">CPU 使用率</span>
      </div>
      <div class="flex items-center space-x-2">
        <div class="w-3 h-0.5 bg-green-500"></div>
        <span class="text-xs text-gray-300">内存使用率</span>
      </div>
      <div class="flex items-center space-x-2">
        <div class="w-3 h-0.5 bg-yellow-500"></div>
        <span class="text-xs text-gray-300">Pod 数量</span>
      </div>
    </div>
    
    <!-- 工具提示 -->
    <div
      v-if="tooltip.visible"
      class="absolute bg-white/15 backdrop-blur-sm text-white text-xs px-3 py-2 rounded-lg border border-white/30 pointer-events-none z-10"
      :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
    >
      <div class="font-medium">{{ tooltip.label }}</div>
      <div class="text-gray-300">{{ tooltip.value }}{{ tooltip.unit }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface TrendData {
  time: string
  cpu: number
  memory: number
  pods: number
}

interface Props {
  data: TrendData[]
}

const props = defineProps<Props>()

// 工具提示状态
const tooltip = ref({
  visible: false,
  x: 0,
  y: 0,
  label: '',
  value: '',
  unit: ''
})

// 计算图表参数
const xStep = computed(() => props.data.length > 1 ? 500 / (props.data.length - 1) : 0)
const maxPods = computed(() => Math.max(...props.data.map(d => d.pods)))
const yAxisLabels = computed(() => ['0', '25', '50', '75', '100'])

// 获取 Y 轴位置
const getYPosition = (value: number, max: number) => {
  return 260 - (value / max) * 200
}

// 生成路径
const cpuPath = computed(() => {
  if (props.data.length === 0) return ''
  
  let path = `M 50 ${getYPosition(props.data[0].cpu, 100)}`
  for (let i = 1; i < props.data.length; i++) {
    path += ` L ${50 + i * xStep.value} ${getYPosition(props.data[i].cpu, 100)}`
  }
  return path
})

const memoryPath = computed(() => {
  if (props.data.length === 0) return ''
  
  let path = `M 50 ${getYPosition(props.data[0].memory, 100)}`
  for (let i = 1; i < props.data.length; i++) {
    path += ` L ${50 + i * xStep.value} ${getYPosition(props.data[i].memory, 100)}`
  }
  return path
})

const podsPath = computed(() => {
  if (props.data.length === 0) return ''
  
  let path = `M 50 ${getYPosition(props.data[0].pods, maxPods.value)}`
  for (let i = 1; i < props.data.length; i++) {
    path += ` L ${50 + i * xStep.value} ${getYPosition(props.data[i].pods, maxPods.value)}`
  }
  return path
})

// 工具提示方法
const showTooltip = (event: MouseEvent, label: string, value: number, unit: string) => {
  const rect = (event.target as SVGElement).getBoundingClientRect()
  const container = (event.target as SVGElement).closest('.relative')!.getBoundingClientRect()
  
  tooltip.value = {
    visible: true,
    x: rect.left - container.left + rect.width / 2 - 40,
    y: rect.top - container.top - 50,
    label,
    value: value.toString(),
    unit
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