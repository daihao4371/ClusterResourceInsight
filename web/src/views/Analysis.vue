<template>
  <div class="space-y-6 animate-fade-in">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gradient">资源分析</h1>
        <p class="text-gray-400 mt-1">深度分析集群资源配置与使用效率</p>
      </div>
      
      <div class="flex items-center space-x-3">
        <button class="btn-secondary">
          <RefreshCw class="w-4 h-4 mr-2" />
          刷新分析
        </button>
        <button class="btn-secondary">
          <Download class="w-4 h-4 mr-2" />
          导出报告
        </button>
      </div>
    </div>

    <!-- 分析概览 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <MetricCard
        title="总Pod数量"
        :value="analysisData?.total_pods || 0"
        icon="Box"
        status="info"
      />
      <MetricCard
        title="异常配置Pod"
        :value="analysisData?.unreasonable_pods || 0"
        icon="AlertTriangle"
        status="warning"
      />
      <MetricCard
        title="资源利用率"
        :value="resourceUtilization"
        unit="%"
        icon="Activity"
        :status="resourceUtilization > 70 ? 'success' : 'warning'"
      />
      <MetricCard
        title="优化潜力"
        :value="optimizationPotential"
        unit="%"
        icon="TrendingUp"
        status="success"
      />
    </div>

    <!-- 问题Pod Top 50 -->
    <div class="glass-card">
      <div class="p-6 border-b border-gray-700">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-semibold">问题Pod Top 50</h2>
          <div class="flex items-center space-x-3">
            <select v-model="sortBy" class="input-field text-sm">
              <option value="cpu_waste">CPU浪费</option>
              <option value="memory_waste">内存浪费</option>
              <option value="total_waste">总浪费</option>
            </select>
            <button class="btn-secondary text-sm">
              <Filter class="w-4 h-4 mr-2" />
              筛选
            </button>
          </div>
        </div>
      </div>
      
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-dark-800/50">
            <tr class="border-b border-gray-700">
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">排名</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">Pod信息</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">配置问题</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">资源请求</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">实际使用</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">浪费程度</th>
              <th class="px-6 py-4 text-left text-xs font-medium text-gray-400 uppercase">建议</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-700">
            <tr
              v-for="(pod, index) in sortedProblems"
              :key="`${pod.cluster_name}-${pod.namespace}-${pod.name}`"
              class="hover:bg-white/5 transition-colors"
            >
              <!-- 排名 -->
              <td class="px-6 py-4">
                <div class="flex items-center">
                  <span 
                    class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium"
                    :class="getRankBadgeClass(index + 1)"
                  >
                    {{ index + 1 }}
                  </span>
                </div>
              </td>
              
              <!-- Pod信息 -->
              <td class="px-6 py-4">
                <div>
                  <div class="font-medium text-white">{{ pod.name }}</div>
                  <div class="text-sm text-gray-400">{{ pod.namespace }}/{{ pod.cluster_name }}</div>
                </div>
              </td>
              
              <!-- 配置问题 -->
              <td class="px-6 py-4">
                <div class="space-y-1">
                  <span 
                    v-for="issue in pod.issues"
                    :key="issue"
                    class="inline-block px-2 py-1 text-xs rounded-full"
                    :class="getIssueBadgeClass(issue)"
                  >
                    {{ getIssueText(issue) }}
                  </span>
                </div>
              </td>
              
              <!-- 资源请求 -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div>CPU: {{ pod.cpu_request || 'N/A' }}</div>
                  <div>MEM: {{ pod.memory_request || 'N/A' }}</div>
                </div>
              </td>
              
              <!-- 实际使用 -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div class="flex items-center space-x-2">
                    <span>{{ pod.cpu_usage || 'N/A' }}</span>
                    <div 
                      v-if="pod.cpu_usage_percent"
                      class="w-16 h-1.5 bg-dark-700 rounded-full overflow-hidden"
                    >
                      <div 
                        class="h-full rounded-full transition-all"
                        :class="getUsageBarClass(pod.cpu_usage_percent)"
                        :style="{ width: `${Math.min(pod.cpu_usage_percent, 100)}%` }"
                      ></div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-2">
                    <span>{{ pod.memory_usage || 'N/A' }}</span>
                    <div 
                      v-if="pod.memory_usage_percent"
                      class="w-16 h-1.5 bg-dark-700 rounded-full overflow-hidden"
                    >
                      <div 
                        class="h-full rounded-full transition-all"
                        :class="getUsageBarClass(pod.memory_usage_percent)"
                        :style="{ width: `${Math.min(pod.memory_usage_percent, 100)}%` }"
                      ></div>
                    </div>
                  </div>
                </div>
              </td>
              
              <!-- 浪费程度 -->
              <td class="px-6 py-4">
                <div class="text-sm space-y-1">
                  <div 
                    class="font-medium"
                    :class="getWasteClass(calculateWaste(pod))"
                  >
                    {{ calculateWaste(pod) }}%
                  </div>
                  <div class="text-xs text-gray-500">资源浪费</div>
                </div>
              </td>
              
              <!-- 建议 -->
              <td class="px-6 py-4">
                <button class="btn-secondary text-sm">
                  <Lightbulb class="w-4 h-4 mr-1" />
                  优化建议
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 资源分布图表 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- CPU资源分布 -->
      <div class="glass-card p-6">
        <h3 class="text-lg font-semibold mb-4">CPU资源分布</h3>
        <div class="h-64 flex items-center justify-center text-gray-500">
          <BarChart3 class="w-12 h-12 mr-3" />
          图表组件开发中...
        </div>
      </div>
      
      <!-- 内存资源分布 -->
      <div class="glass-card p-6">
        <h3 class="text-lg font-semibold mb-4">内存资源分布</h3>
        <div class="h-64 flex items-center justify-center text-gray-500">
          <PieChart class="w-12 h-12 mr-3" />
          图表组件开发中...
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="!analysisData" class="text-center py-12">
      <BarChart3 class="w-16 h-16 text-gray-600 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-400 mb-2">暂无分析数据</h3>
      <p class="text-gray-500 mb-6">点击刷新按钮开始资源分析</p>
      <button class="btn-primary">
        <RefreshCw class="w-4 h-4 mr-2" />
        开始分析
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { 
  RefreshCw, 
  Download, 
  Filter, 
  Lightbulb,
  BarChart3, 
  PieChart,
  Box,
  AlertTriangle,
  Activity,
  TrendingUp
} from 'lucide-vue-next'
import MetricCard from '../components/common/MetricCard.vue'
import { useAnalysis } from '../composables/api'
import type { Pod } from '../types'

const { analysis: analysisData, loading, fetchAnalysis } = useAnalysis()

const sortBy = ref('cpu_waste')

// 计算资源利用率
const resourceUtilization = computed(() => {
  if (!analysisData.value) return 0
  const total = analysisData.value.total_pods
  const problems = analysisData.value.unreasonable_pods
  return total > 0 ? Math.round(((total - problems) / total) * 100) : 0
})

// 计算优化潜力
const optimizationPotential = computed(() => {
  if (!analysisData.value) return 0
  return Math.min(100 - resourceUtilization.value, 30)
})

// 排序后的问题Pod列表
const sortedProblems = computed(() => {
  if (!analysisData.value?.top_50_problems) return []
  
  const pods = [...analysisData.value.top_50_problems]
  
  return pods.sort((a, b) => {
    const wasteA = calculateWaste(a)
    const wasteB = calculateWaste(b)
    return wasteB - wasteA
  })
})

// 计算浪费程度
const calculateWaste = (pod: Pod) => {
  if (!pod.cpu_usage_percent && !pod.memory_usage_percent) return 0
  
  const cpuWaste = pod.cpu_usage_percent ? 100 - pod.cpu_usage_percent : 0
  const memoryWaste = pod.memory_usage_percent ? 100 - pod.memory_usage_percent : 0
  
  return Math.round((cpuWaste + memoryWaste) / 2)
}

// 样式方法
const getRankBadgeClass = (rank: number) => {
  if (rank <= 3) return 'bg-danger-500/20 text-danger-400 border border-danger-500/30'
  if (rank <= 10) return 'bg-warning-500/20 text-warning-400 border border-warning-500/30'
  return 'bg-gray-500/20 text-gray-400 border border-gray-500/30'
}

const getIssueBadgeClass = (issue: string) => {
  if (issue.includes('over')) return 'bg-danger-500/20 text-danger-400'
  if (issue.includes('under')) return 'bg-warning-500/20 text-warning-400'
  return 'bg-primary-500/20 text-primary-400'
}

const getIssueText = (issue: string) => {
  const issueMap: Record<string, string> = {
    'cpu_over_request': 'CPU超配',
    'memory_over_request': '内存超配',
    'cpu_under_utilization': 'CPU利用率低',
    'memory_under_utilization': '内存利用率低',
    'no_limits': '无资源限制',
    'no_requests': '无资源请求'
  }
  return issueMap[issue] || issue
}

const getUsageBarClass = (percentage: number) => {
  if (percentage >= 80) return 'bg-danger-500'
  if (percentage >= 60) return 'bg-warning-500'
  return 'bg-success-500'
}

const getWasteClass = (waste: number) => {
  if (waste >= 50) return 'text-danger-400'
  if (waste >= 30) return 'text-warning-400'
  return 'text-success-400'
}

onMounted(() => {
  fetchAnalysis()
})
</script>