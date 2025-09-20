<template>
  <div class="w-full h-80 relative">
    <!-- 环形图中心信息 -->
    <div class="absolute inset-0 flex items-center justify-center">
      <div class="text-center">
        <div class="text-3xl font-bold text-white">{{ totalClusters }}</div>
        <div class="text-sm text-gray-400">总集群数</div>
      </div>
    </div>
    
    <!-- SVG 环形图 -->
    <svg class="w-full h-full" viewBox="0 0 200 200">
      <!-- 背景圆环 -->
      <circle
        cx="100"
        cy="100"
        :r="radius"
        fill="none"
        stroke="#374151"
        stroke-width="8"
      />
      
      <!-- 数据圆环 -->
      <circle
        v-for="(segment, index) in segments"
        :key="index"
        cx="100"
        cy="100"
        :r="radius"
        fill="none"
        :stroke="segment.color"
        stroke-width="8"
        stroke-linecap="round"
        :stroke-dasharray="segment.dashArray"
        :stroke-dashoffset="segment.dashOffset"
        :style="{ transform: `rotate(${segment.rotation}deg)`, transformOrigin: '100px 100px' }"
        class="transition-all duration-1000 ease-out"
      />
    </svg>
    
    <!-- 图例 -->
    <div class="absolute bottom-0 left-0 right-0">
      <div class="flex justify-center space-x-6">
        <div 
          v-for="item in data"
          :key="item.name"
          class="flex items-center space-x-2 cursor-pointer hover:scale-105 transition-transform"
          @click="toggleSegment(item.name)"
        >
          <div 
            class="w-3 h-3 rounded-full"
            :style="{ backgroundColor: item.color }"
          ></div>
          <span class="text-sm text-gray-300">{{ item.name }}</span>
          <span class="text-sm font-medium text-white">{{ item.value }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface ChartData {
  name: string
  value: number
  color: string
}

interface Props {
  data: ChartData[]
}

const props = defineProps<Props>()

const radius = 60
const circumference = 2 * Math.PI * radius
const hiddenSegments = ref<Set<string>>(new Set())

// 计算总数
const totalClusters = computed(() => {
  return props.data.reduce((sum, item) => sum + item.value, 0)
})

// 计算有效数据（排除隐藏的段）
const visibleData = computed(() => {
  return props.data.filter(item => !hiddenSegments.value.has(item.name))
})

// 计算可见总数
const visibleTotal = computed(() => {
  return visibleData.value.reduce((sum, item) => sum + item.value, 0)
})

// 计算每个段的参数
const segments = computed(() => {
  if (visibleTotal.value === 0) return []
  
  let currentOffset = 0
  
  return visibleData.value.map(item => {
    const percentage = item.value / visibleTotal.value
    const dashLength = percentage * circumference
    const dashArray = `${dashLength} ${circumference - dashLength}`
    const dashOffset = -currentOffset * circumference
    const rotation = -90 // 从顶部开始
    
    currentOffset += percentage
    
    return {
      color: item.color,
      dashArray,
      dashOffset,
      rotation
    }
  })
})

// 切换段的显示/隐藏
const toggleSegment = (name: string) => {
  if (hiddenSegments.value.has(name)) {
    hiddenSegments.value.delete(name)
  } else {
    hiddenSegments.value.add(name)
  }
}
</script>

<style scoped>
svg {
  filter: drop-shadow(0 0 10px rgba(59, 130, 246, 0.3));
}

circle {
  transition: stroke-dasharray 1s ease-out, stroke-dashoffset 1s ease-out;
}
</style>