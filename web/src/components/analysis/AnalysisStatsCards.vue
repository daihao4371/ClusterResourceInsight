<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- Top内存请求Pod -->
    <div class="glass-card hover:shadow-lg transition-all duration-300" style="background: var(--card-bg); border: 1px solid var(--border-color);">
      <div class="p-6 border-b" style="border-color: var(--border-color);">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <div class="p-2 rounded-lg" style="background: var(--icon-bg-memory);">
              <Database class="w-5 h-5" style="color: var(--icon-color-memory);" />
            </div>
            <h3 class="text-lg font-semibold" style="color: var(--text-primary);">Top内存请求Pod</h3>
          </div>
          <button 
            class="btn-secondary text-sm transition-colors"
            style="background: var(--btn-secondary-bg); color: var(--btn-secondary-text); border: 1px solid var(--border-color);"
            @click="$emit('refresh-memory')"
            :disabled="loading"
          >
            <RefreshCw :class="['w-3 h-3 mr-1', { 'animate-spin': loading }]" />
            刷新
          </button>
        </div>
      </div>
      <div class="p-4 max-h-80 overflow-y-auto">
        <div v-if="!topMemoryPods || topMemoryPods.length === 0" class="text-center py-8" style="color: var(--text-secondary);">
          <Database class="w-8 h-8 mx-auto mb-2" style="color: var(--text-muted);" />
          <p>暂无数据</p>
        </div>
        <div v-else class="space-y-3">
          <div 
            v-for="pod in topMemoryPods.slice(0, 10)"
            :key="`memory-${pod.cluster_name}-${pod.namespace}-${pod.pod_name}`"
            class="flex items-center justify-between p-3 rounded-lg border transition-all duration-200 hover:shadow-sm"
            style="background: var(--item-bg); border-color: var(--border-color);"
          >
            <div class="flex-1 min-w-0">
              <div class="font-medium text-sm truncate" style="color: var(--text-primary);">{{ pod.pod_name }}</div>
              <div class="text-xs" style="color: var(--text-secondary);">{{ pod.namespace }}</div>
            </div>
            <div class="text-right">
              <div class="text-sm font-semibold px-2 py-1 rounded" style="color: var(--text-primary); background: var(--value-bg);">{{ formatMemoryValue(pod.memory_request) }}</div>
              <div class="text-xs mt-1" style="color: var(--text-muted);">请求量</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Top CPU请求Pod -->
    <div class="glass-card hover:shadow-lg transition-all duration-300" style="background: var(--card-bg); border: 1px solid var(--border-color);">
      <div class="p-6 border-b" style="border-color: var(--border-color);">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <div class="p-2 rounded-lg" style="background: var(--icon-bg-cpu);">
              <Cpu class="w-5 h-5" style="color: var(--icon-color-cpu);" />
            </div>
            <h3 class="text-lg font-semibold" style="color: var(--text-primary);">Top CPU请求Pod</h3>
          </div>
          <button 
            class="btn-secondary text-sm transition-colors"
            style="background: var(--btn-secondary-bg); color: var(--btn-secondary-text); border: 1px solid var(--border-color);"
            @click="$emit('refresh-cpu')"
            :disabled="loading"
          >
            <RefreshCw :class="['w-3 h-3 mr-1', { 'animate-spin': loading }]" />
            刷新
          </button>
        </div>
      </div>
      <div class="p-4 max-h-80 overflow-y-auto">
        <div v-if="!topCpuPods || topCpuPods.length === 0" class="text-center py-8" style="color: var(--text-secondary);">
          <Cpu class="w-8 h-8 mx-auto mb-2" style="color: var(--text-muted);" />
          <p>暂无数据</p>
        </div>
        <div v-else class="space-y-3">
          <div 
            v-for="pod in topCpuPods.slice(0, 10)"
            :key="`cpu-${pod.cluster_name}-${pod.namespace}-${pod.pod_name}`"
            class="flex items-center justify-between p-3 rounded-lg border transition-all duration-200 hover:shadow-sm"
            style="background: var(--item-bg); border-color: var(--border-color);"
          >
            <div class="flex-1 min-w-0">
              <div class="font-medium text-sm truncate" style="color: var(--text-primary);">{{ pod.pod_name }}</div>
              <div class="text-xs" style="color: var(--text-secondary);">{{ pod.namespace }}</div>
            </div>
            <div class="text-right">
              <div class="text-sm font-semibold px-2 py-1 rounded" style="color: var(--text-primary); background: var(--value-bg);">{{ formatCpuValue(pod.cpu_request) }}</div>
              <div class="text-xs mt-1" style="color: var(--text-muted);">请求量</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 命名空间汇总 -->
    <div class="glass-card hover:shadow-lg transition-all duration-300" style="background: var(--card-bg); border: 1px solid var(--border-color);">
      <div class="p-6 border-b" style="border-color: var(--border-color);">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center space-x-3">
            <div class="p-2 rounded-lg" style="background: var(--icon-bg-namespace);">
              <Layers class="w-5 h-5" style="color: var(--icon-color-namespace);" />
            </div>
            <h3 class="text-lg font-semibold" style="color: var(--text-primary);">Top 命名空间汇总</h3>
          </div>
          <button 
            class="btn-secondary text-sm transition-colors"
            style="background: var(--btn-secondary-bg); color: var(--btn-secondary-text); border: 1px solid var(--border-color);"
            @click="$emit('refresh-namespace')"
            :disabled="loading"
          >
            <RefreshCw :class="['w-3 h-3 mr-1', { 'animate-spin': loading }]" />
            刷新
          </button>
        </div>
        <!-- 排序选择器 -->
        <div class="flex items-center justify-between">
          <div class="text-xs" style="color: var(--text-muted);">
            按资源使用量排序
          </div>
          <div class="relative">
            <select
              :value="namespaceSortBy || 'combined'"
              @change="$emit('namespace-sort-change', ($event.target as HTMLSelectElement).value)"
              class="appearance-none text-xs px-2 py-1 pr-6 rounded border transition-colors"
              style="background: var(--btn-secondary-bg); color: var(--btn-secondary-text); border-color: var(--border-color);"
              :disabled="loading"
            >
              <option value="combined">综合资源</option>
              <option value="memory">内存使用量</option>
              <option value="cpu">CPU使用量</option>
            </select>
            <ChevronDown class="absolute right-1 top-1/2 transform -translate-y-1/2 w-3 h-3 pointer-events-none" style="color: var(--text-muted);" />
          </div>
        </div>
      </div>
      <div class="p-4 max-h-80 overflow-y-auto">
        <!-- 加载状态 -->
        <div v-if="loading" class="text-center py-8">
          <RefreshCw class="w-8 h-8 mx-auto mb-2 animate-spin" style="color: var(--text-primary);" />
          <p style="color: var(--text-secondary);">加载中...</p>
        </div>
        <!-- 空数据状态 -->
        <div v-else-if="!namespaceSummary || namespaceSummary.length === 0" class="text-center py-8" style="color: var(--text-secondary);">
          <Layers class="w-8 h-8 mx-auto mb-2" style="color: var(--text-muted);" />
          <p>暂无数据</p>
        </div>
        <div v-else class="space-y-3">
          <div 
            v-for="ns in namespaceSummary.slice(0, 8)"
            :key="`ns-${ns.cluster_name}-${ns.namespace_name}`"
            class="p-3 rounded-lg border transition-all duration-200 hover:shadow-sm"
            style="background: var(--item-bg); border-color: var(--border-color);"
          >
            <div class="flex items-center justify-between mb-2">
              <div class="font-medium text-sm" style="color: var(--text-primary);">{{ ns.namespace_name }}</div>
              <div class="text-xs" style="color: var(--text-secondary);">{{ ns.cluster_name }}</div>
            </div>
            <div class="grid grid-cols-2 gap-2 text-xs">
              <div>
                <span style="color: var(--text-muted);">Pod数:</span>
                <span class="ml-1 font-medium" style="color: var(--text-primary);">{{ ns.total_pods }}</span>
              </div>
              <div>
                <span style="color: var(--text-muted);">正常:</span>
                <span class="ml-1 font-medium" style="color: var(--success-color);">{{ getRunningPods(ns) }}</span>
              </div>
              <div>
                <span style="color: var(--text-muted);">内存:</span>
                <span class="ml-1 font-medium" style="color: var(--text-primary);">{{ formatMemoryValue(ns.total_memory_usage) }}</span>
              </div>
              <div>
                <span style="color: var(--text-muted);">CPU:</span>
                <span class="ml-1 font-medium" style="color: var(--text-primary);">{{ formatCpuValue(ns.total_cpu_usage) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RefreshCw, Database, Cpu, Layers, ChevronDown } from 'lucide-vue-next'
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
  total_pods: number  // 后端使用 total_pods，前端需要适配
  unreasonable_pods: number
  total_memory_usage: number
  total_cpu_usage: number
  total_memory_request: number
  total_cpu_request: number
}

// 计算运行中的Pod数量（用于向后兼容）
const getRunningPods = (summary: NamespaceSummary): number => {
  // 假设运行中的Pod数量 = 总Pod数量 - 有问题的Pod数量
  return Math.max(0, summary.total_pods - (summary.unreasonable_pods || 0))
}

// 定义props接口
interface Props {
  topMemoryPods?: Pod[]
  topCpuPods?: Pod[]
  namespaceSummary?: NamespaceSummary[]
  namespaceSortBy?: string  // 新增：命名空间排序方式
  loading?: boolean
}

// 定义事件接口
interface Emits {
  'refresh-memory': []
  'refresh-cpu': []
  'refresh-namespace': []
  'namespace-sort-change': [sortBy: string]  // 新增：命名空间排序变更事件
}

// 接收props和定义emits
defineProps<Props>()
defineEmits<Emits>()
</script>