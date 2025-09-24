<template>
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-bold text-gradient">资源分析</h1>
      <p class="text-gray-400 text-sm">深度分析集群资源配置与使用效率</p>
    </div>
    
    <div class="flex items-center space-x-3">
      <button 
        class="btn-secondary"
        :disabled="loading"
        @click="$emit('refresh')"
      >
        <RefreshCw :class="['w-4 h-4 mr-2', { 'animate-spin': loading }]" />
        {{ loading ? '分析中...' : '刷新分析' }}
      </button>
      <button 
        class="btn-secondary"
        @click="$emit('collect')"
      >
        <Database class="w-4 h-4 mr-2" />
        深度收集
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RefreshCw, Database } from 'lucide-vue-next'

// 定义props接口
interface Props {
  loading?: boolean
}

// 定义事件接口  
interface Emits {
  refresh: []
  collect: []
}

// 接收props和定义emits
defineProps<Props>()
defineEmits<Emits>()
</script>