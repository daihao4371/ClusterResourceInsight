<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- CPU资源分布 -->
    <div class="glass-card hover:shadow-lg transition-all duration-300" style="background: var(--card-bg); border: 1px solid var(--border-color);">
      <div class="p-6 border-b" style="border-color: var(--border-color);">
        <div class="flex items-center space-x-3">
          <div class="p-2 bg-blue-500/10 dark:bg-blue-400/20 rounded-lg">
            <BarChart3 class="w-5 h-5 text-blue-600 dark:text-blue-400" />
          </div>
          <h3 class="text-lg font-semibold" style="color: var(--text-primary);">CPU资源分布</h3>
        </div>
      </div>
      
      <div class="p-6">
        <!-- CPU资源统计 -->
        <div class="space-y-4 mb-6">
          <div class="flex justify-between items-center">
            <span class="text-sm" style="color: var(--text-secondary);">总请求量</span>
            <span class="font-semibold" style="color: var(--text-primary);">{{ formatCpuTotal(cpuStats.totalRequest) }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-sm" style="color: var(--text-secondary);">实际使用</span>
            <span class="font-semibold" style="color: var(--text-primary);">{{ formatCpuTotal(cpuStats.totalUsage) }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-sm" style="color: var(--text-secondary);">使用率</span>
            <span class="font-semibold" style="color: var(--text-primary);">{{ cpuStats.utilizationRate }}%</span>
          </div>
        </div>
        
        <!-- CPU使用率进度条 -->
        <div class="space-y-3">
          <div class="flex justify-between text-xs" style="color: var(--text-muted);">
            <span>CPU使用率</span>
            <span>{{ cpuStats.utilizationRate }}%</span>
          </div>
          <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
            <div 
              class="bg-blue-500 dark:bg-blue-400 h-3 rounded-full transition-all duration-500 ease-out"
              :style="{ width: `${Math.min(cpuStats.utilizationRate, 100)}%` }"
            ></div>
          </div>
          <div class="text-xs" style="color: var(--text-muted);">
            <span v-if="cpuStats.utilizationRate < 50" class="text-green-600 dark:text-green-400">• 资源充足</span>
            <span v-else-if="cpuStats.utilizationRate < 80" class="text-yellow-600 dark:text-yellow-400">• 使用适中</span>
            <span v-else class="text-red-600 dark:text-red-400">• 资源紧张</span>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 内存资源分布 -->
    <div class="glass-card hover:shadow-lg transition-all duration-300" style="background: var(--card-bg); border: 1px solid var(--border-color);">
      <div class="p-6 border-b" style="border-color: var(--border-color);">
        <div class="flex items-center space-x-3">
          <div class="p-2 bg-green-500/10 dark:bg-green-400/20 rounded-lg">
            <PieChart class="w-5 h-5 text-green-600 dark:text-green-400" />
          </div>
          <h3 class="text-lg font-semibold" style="color: var(--text-primary);">内存资源分布</h3>
        </div>
      </div>
      
      <div class="p-6">
        <!-- 内存资源统计 -->
        <div class="space-y-4 mb-6">
          <div class="flex justify-between items-center">
            <span class="text-sm" style="color: var(--text-secondary);">总请求量</span>
            <span class="font-semibold" style="color: var(--text-primary);">{{ formatMemoryTotal(memoryStats.totalRequest) }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-sm" style="color: var(--text-secondary);">实际使用</span>
            <span class="font-semibold" style="color: var(--text-primary);">{{ formatMemoryTotal(memoryStats.totalUsage) }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-sm" style="color: var(--text-secondary);">使用率</span>
            <span class="font-semibold" style="color: var(--text-primary);">{{ memoryStats.utilizationRate }}%</span>
          </div>
        </div>
        
        <!-- 内存使用率进度条 -->
        <div class="space-y-3">
          <div class="flex justify-between text-xs" style="color: var(--text-muted);">
            <span>内存使用率</span>
            <span>{{ memoryStats.utilizationRate }}%</span>
          </div>
          <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
            <div 
              class="bg-green-500 dark:bg-green-400 h-3 rounded-full transition-all duration-500 ease-out"
              :style="{ width: `${Math.min(memoryStats.utilizationRate, 100)}%` }"
            ></div>
          </div>
          <div class="text-xs" style="color: var(--text-muted);">
            <span v-if="memoryStats.utilizationRate < 60" class="text-green-600 dark:text-green-400">• 内存充足</span>
            <span v-else-if="memoryStats.utilizationRate < 85" class="text-yellow-600 dark:text-yellow-400">• 使用适中</span>
            <span v-else class="text-red-600 dark:text-red-400">• 内存紧张</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { BarChart3, PieChart } from 'lucide-vue-next'

// 模拟CPU和内存统计数据
const cpuStats = ref({
  totalRequest: 0,
  totalUsage: 0,
  utilizationRate: 0
})

const memoryStats = ref({
  totalRequest: 0,
  totalUsage: 0,
  utilizationRate: 0
})

// 格式化CPU数值（毫核心）
const formatCpuTotal = (value: number): string => {
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1)}Core`
  }
  return `${value}m`
}

// 格式化内存数值（字节）
const formatMemoryTotal = (value: number): string => {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let index = 0
  let size = value
  
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index++
  }
  
  return `${size.toFixed(1)}${units[index]}`
}

// 模拟生成统计数据
const generateMockStats = () => {
  // 模拟CPU数据（毫核心）
  cpuStats.value = {
    totalRequest: Math.floor(Math.random() * 50000 + 20000), // 20-70 Core
    totalUsage: Math.floor(Math.random() * 30000 + 10000),   // 10-40 Core 使用
    utilizationRate: 0
  }
  cpuStats.value.utilizationRate = Math.floor(
    (cpuStats.value.totalUsage / cpuStats.value.totalRequest) * 100
  )
  
  // 模拟内存数据（字节）
  memoryStats.value = {
    totalRequest: Math.floor(Math.random() * 1024 * 1024 * 1024 * 100 + 1024 * 1024 * 1024 * 50), // 50-150GB
    totalUsage: Math.floor(Math.random() * 1024 * 1024 * 1024 * 80 + 1024 * 1024 * 1024 * 30),     // 30-110GB 使用
    utilizationRate: 0
  }
  memoryStats.value.utilizationRate = Math.floor(
    (memoryStats.value.totalUsage / memoryStats.value.totalRequest) * 100
  )
}

// 组件挂载时生成模拟数据
onMounted(() => {
  generateMockStats()
})
</script>