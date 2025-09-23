<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- Top内存请求Pod -->
    <div class="glass-card">
      <div class="p-6 border-b border-gray-700">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">Top内存请求Pod</h3>
          <button 
            class="btn-secondary text-sm"
            @click="$emit('refresh-memory')"
            :disabled="loading"
          >
            <RefreshCw :class="['w-3 h-3 mr-1', { 'animate-spin': loading }]" />
            刷新
          </button>
        </div>
      </div>
      <div class="p-4 max-h-80 overflow-y-auto">
        <div v-if="!topMemoryPods || topMemoryPods.length === 0" class="text-center py-8 text-gray-500">
          <Database class="w-8 h-8 mx-auto mb-2" />
          <p>暂无数据</p>
        </div>
        <div v-else class="space-y-3">
          <div 
            v-for="pod in topMemoryPods.slice(0, 10)"
            :key="`memory-${pod.cluster_name}-${pod.namespace}-${pod.pod_name}`"
            class="flex items-center justify-between p-3 bg-dark-800/30 rounded-lg"
          >
            <div class="flex-1 min-w-0">
              <div class="font-medium text-sm truncate">{{ pod.pod_name }}</div>
              <div class="text-xs text-gray-400">{{ pod.namespace }}</div>
            </div>
            <div class="text-right">
              <div class="text-sm font-medium">{{ formatMemoryValue(pod.memory_request) }}</div>
              <div class="text-xs text-gray-500">请求量</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Top CPU请求Pod -->
    <div class="glass-card">
      <div class="p-6 border-b border-gray-700">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">Top CPU请求Pod</h3>
          <button 
            class="btn-secondary text-sm"
            @click="$emit('refresh-cpu')"
            :disabled="loading"
          >
            <RefreshCw :class="['w-3 h-3 mr-1', { 'animate-spin': loading }]" />
            刷新
          </button>
        </div>
      </div>
      <div class="p-4 max-h-80 overflow-y-auto">
        <div v-if="!topCpuPods || topCpuPods.length === 0" class="text-center py-8 text-gray-500">
          <Cpu class="w-8 h-8 mx-auto mb-2" />
          <p>暂无数据</p>
        </div>
        <div v-else class="space-y-3">
          <div 
            v-for="pod in topCpuPods.slice(0, 10)"
            :key="`cpu-${pod.cluster_name}-${pod.namespace}-${pod.pod_name}`"
            class="flex items-center justify-between p-3 bg-dark-800/30 rounded-lg"
          >
            <div class="flex-1 min-w-0">
              <div class="font-medium text-sm truncate">{{ pod.pod_name }}</div>
              <div class="text-xs text-gray-400">{{ pod.namespace }}</div>
            </div>
            <div class="text-right">
              <div class="text-sm font-medium">{{ formatCpuValue(pod.cpu_request) }}</div>
              <div class="text-xs text-gray-500">请求量</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 命名空间汇总 -->
    <div class="glass-card">
      <div class="p-6 border-b border-gray-700">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">命名空间汇总</h3>
          <button 
            class="btn-secondary text-sm"
            @click="$emit('refresh-namespace')"
            :disabled="loading"
          >
            <RefreshCw :class="['w-3 h-3 mr-1', { 'animate-spin': loading }]" />
            刷新
          </button>
        </div>
      </div>
      <div class="p-4 max-h-80 overflow-y-auto">
        <div v-if="!namespaceSummary || namespaceSummary.length === 0" class="text-center py-8 text-gray-500">
          <Layers class="w-8 h-8 mx-auto mb-2" />
          <p>暂无数据</p>
        </div>
        <div v-else class="space-y-3">
          <div 
            v-for="ns in namespaceSummary.slice(0, 8)"
            :key="`ns-${ns.cluster_name}-${ns.namespace_name}`"
            class="p-3 bg-dark-800/30 rounded-lg"
          >
            <div class="flex items-center justify-between mb-2">
              <div class="font-medium text-sm">{{ ns.namespace_name }}</div>
              <div class="text-xs text-gray-400">{{ ns.cluster_name }}</div>
            </div>
            <div class="grid grid-cols-2 gap-2 text-xs">
              <div>
                <span class="text-gray-400">Pod数:</span>
                <span class="ml-1 text-white">{{ ns.pod_count }}</span>
              </div>
              <div>
                <span class="text-gray-400">运行中:</span>
                <span class="ml-1 text-success-400">{{ ns.running_pods }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RefreshCw, Database, Cpu, Layers } from 'lucide-vue-next'
import { formatCpuValue, formatMemoryValue } from '../../utils/analysis'

// 定义Pod接口类型
interface Pod {
  cluster_name: string
  namespace: string  
  pod_name: string
  memory_request: number
  cpu_request: number
}

// 定义命名空间接口类型
interface NamespaceSummary {
  cluster_name: string
  namespace_name: string
  pod_count: number
  running_pods: number
}

// 定义props接口
interface Props {
  topMemoryPods?: Pod[]
  topCpuPods?: Pod[]
  namespaceSummary?: NamespaceSummary[]
  loading?: boolean
}

// 定义事件接口
interface Emits {
  'refresh-memory': []
  'refresh-cpu': []
  'refresh-namespace': []
}

// 接收props和定义emits
defineProps<Props>()
defineEmits<Emits>()
</script>