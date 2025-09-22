<template>
  <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" v-if="visible" @click="handleBackdropClick">
    <div class="bg-dark-800 rounded-lg border border-gray-700 w-full max-w-6xl max-h-[90vh] overflow-hidden flex flex-col" @click.stop>
      <!-- 模态框头部 -->
      <div class="flex items-center justify-between p-6 border-b border-gray-700">
        <div>
          <h2 class="text-xl font-semibold text-white">Pod资源分析详情</h2>
          <p class="text-sm text-gray-400 mt-1" v-if="analysis">
            {{ analysis.pod_info.cluster_name }} / {{ analysis.pod_info.namespace }} / {{ analysis.pod_info.pod_name }}
          </p>
        </div>
        <button @click="close" class="p-2 hover:bg-gray-700 rounded-lg transition-colors">
          <X class="w-5 h-5 text-gray-400" />
        </button>
      </div>

      <!-- 模态框内容 -->
      <div class="flex-1 overflow-y-auto p-6">
        <!-- 加载状态 -->
        <div v-if="loading" class="flex items-center justify-center py-12">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
          <span class="ml-3 text-gray-400">正在加载分析数据...</span>
        </div>

        <!-- 错误状态 -->
        <div v-else-if="error" class="text-center py-12">
          <AlertTriangle class="w-12 h-12 text-danger-400 mx-auto mb-4" />
          <h3 class="text-lg font-semibold text-danger-400 mb-2">加载失败</h3>
          <p class="text-gray-400 mb-4">{{ error }}</p>
          <button @click="loadAnalysis" class="btn-secondary">重新加载</button>
        </div>

        <!-- 分析内容 -->
        <div v-else-if="analysis" class="space-y-6">
          <!-- 基础信息 -->
          <div class="glass-card p-6">
            <h3 class="text-lg font-semibold text-white mb-4">基础信息</h3>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label class="text-sm text-gray-400">集群环境</label>
                <p class="text-white font-medium">{{ analysis.cluster_info }}</p>
              </div>
              <div>
                <label class="text-sm text-gray-400">运行节点</label>
                <p class="text-white font-medium">{{ analysis.node_info }}</p>
              </div>
              <div>
                <label class="text-sm text-gray-400">创建时间</label>
                <p class="text-white font-medium">{{ formatDate(analysis.pod_info.creation_time) }}</p>
              </div>
            </div>
          </div>

          <!-- 资源分析 -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- 内存分析 -->
            <div class="glass-card p-6">
              <h3 class="text-lg font-semibold text-white mb-4 flex items-center">
                <Database class="w-5 h-5 mr-2 text-primary-400" />
                内存分析
              </h3>
              <div class="space-y-4">
                <div>
                  <label class="text-sm text-gray-400">配置状态</label>
                  <p class="text-white font-medium" :class="getStatusColor(analysis.resource_analysis.memory_analysis.config_status)">
                    {{ analysis.resource_analysis.memory_analysis.config_status }}
                  </p>
                </div>
                <div>
                  <label class="text-sm text-gray-400">使用效率</label>
                  <p class="text-white font-medium">{{ analysis.resource_analysis.memory_analysis.efficiency_score.toFixed(1) }}%</p>
                  <div class="w-full bg-gray-700 rounded-full h-2 mt-1">
                    <div class="bg-primary-500 h-2 rounded-full" :style="{ width: `${Math.min(analysis.resource_analysis.memory_analysis.efficiency_score, 100)}%` }"></div>
                  </div>
                </div>
                <div>
                  <label class="text-sm text-gray-400">浪费资源</label>
                  <p class="text-white font-medium">{{ formatBytes(analysis.resource_analysis.memory_analysis.waste_amount) }}</p>
                </div>
                <div>
                  <label class="text-sm text-gray-400">优化建议</label>
                  <ul class="text-sm text-gray-300 space-y-1 mt-1">
                    <li v-for="rec in analysis.resource_analysis.memory_analysis.recommendations" :key="rec" class="flex items-start">
                      <span class="w-1.5 h-1.5 bg-primary-400 rounded-full mt-2 mr-2 flex-shrink-0"></span>
                      {{ rec }}
                    </li>
                  </ul>
                </div>
              </div>
            </div>

            <!-- CPU分析 -->
            <div class="glass-card p-6">
              <h3 class="text-lg font-semibold text-white mb-4 flex items-center">
                <Cpu class="w-5 h-5 mr-2 text-warning-400" />
                CPU分析
              </h3>
              <div class="space-y-4">
                <div>
                  <label class="text-sm text-gray-400">配置状态</label>
                  <p class="text-white font-medium" :class="getStatusColor(analysis.resource_analysis.cpu_analysis.config_status)">
                    {{ analysis.resource_analysis.cpu_analysis.config_status }}
                  </p>
                </div>
                <div>
                  <label class="text-sm text-gray-400">使用效率</label>
                  <p class="text-white font-medium">{{ analysis.resource_analysis.cpu_analysis.efficiency_score.toFixed(1) }}%</p>
                  <div class="w-full bg-gray-700 rounded-full h-2 mt-1">
                    <div class="bg-warning-500 h-2 rounded-full" :style="{ width: `${Math.min(analysis.resource_analysis.cpu_analysis.efficiency_score, 100)}%` }"></div>
                  </div>
                </div>
                <div>
                  <label class="text-sm text-gray-400">浪费资源</label>
                  <p class="text-white font-medium">{{ formatMillicores(analysis.resource_analysis.cpu_analysis.waste_amount) }}</p>
                </div>
                <div>
                  <label class="text-sm text-gray-400">优化建议</label>
                  <ul class="text-sm text-gray-300 space-y-1 mt-1">
                    <li v-for="rec in analysis.resource_analysis.cpu_analysis.recommendations" :key="rec" class="flex items-start">
                      <span class="w-1.5 h-1.5 bg-warning-400 rounded-full mt-2 mr-2 flex-shrink-0"></span>
                      {{ rec }}
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

          <!-- 对比分析 -->
          <div class="glass-card p-6">
            <h3 class="text-lg font-semibold text-white mb-4 flex items-center">
              <BarChart class="w-5 h-5 mr-2 text-success-400" />
              对比分析
            </h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <h4 class="text-md font-medium text-white mb-3">与命名空间平均值对比</h4>
                <div class="space-y-3">
                  <div>
                    <div class="flex justify-between text-sm mb-1">
                      <span class="text-gray-400">内存使用率</span>
                      <span class="text-white">{{ analysis.comparison_analysis.namespace_average.memory_usage_pct.toFixed(1) }}%</span>
                    </div>
                    <div class="w-full bg-gray-700 rounded-full h-2">
                      <div class="bg-primary-500 h-2 rounded-full" :style="{ width: `${Math.min(analysis.comparison_analysis.namespace_average.memory_usage_pct, 100)}%` }"></div>
                    </div>
                  </div>
                  <div>
                    <div class="flex justify-between text-sm mb-1">
                      <span class="text-gray-400">CPU使用率</span>
                      <span class="text-white">{{ analysis.comparison_analysis.namespace_average.cpu_usage_pct.toFixed(1) }}%</span>
                    </div>
                    <div class="w-full bg-gray-700 rounded-full h-2">
                      <div class="bg-warning-500 h-2 rounded-full" :style="{ width: `${Math.min(analysis.comparison_analysis.namespace_average.cpu_usage_pct, 100)}%` }"></div>
                    </div>
                  </div>
                </div>
              </div>
              <div>
                <h4 class="text-md font-medium text-white mb-3">与集群平均值对比</h4>
                <div class="space-y-3">
                  <div>
                    <div class="flex justify-between text-sm mb-1">
                      <span class="text-gray-400">内存使用率</span>
                      <span class="text-white">{{ analysis.comparison_analysis.cluster_average.memory_usage_pct.toFixed(1) }}%</span>
                    </div>
                    <div class="w-full bg-gray-700 rounded-full h-2">
                      <div class="bg-primary-500 h-2 rounded-full" :style="{ width: `${Math.min(analysis.comparison_analysis.cluster_average.memory_usage_pct, 100)}%` }"></div>
                    </div>
                  </div>
                  <div>
                    <div class="flex justify-between text-sm mb-1">
                      <span class="text-gray-400">CPU使用率</span>
                      <span class="text-white">{{ analysis.comparison_analysis.cluster_average.cpu_usage_pct.toFixed(1) }}%</span>
                    </div>
                    <div class="w-full bg-gray-700 rounded-full h-2">
                      <div class="bg-warning-500 h-2 rounded-full" :style="{ width: `${Math.min(analysis.comparison_analysis.cluster_average.cpu_usage_pct, 100)}%` }"></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 告警信息 -->
          <div class="glass-card p-6" v-if="analysis.alerts_info.alert_count > 0">
            <h3 class="text-lg font-semibold text-white mb-4 flex items-center">
              <AlertTriangle class="w-5 h-5 mr-2 text-danger-400" />
              告警信息
              <span class="ml-2 px-2 py-1 text-xs rounded-full bg-danger-500/20 text-danger-400">
                {{ analysis.alerts_info.severity_level }}
              </span>
            </h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div v-if="analysis.alerts_info.active_alerts.length > 0">
                <h4 class="text-md font-medium text-white mb-3">活跃告警</h4>
                <ul class="space-y-2">
                  <li v-for="alert in analysis.alerts_info.active_alerts" :key="alert" 
                      class="flex items-center text-sm text-danger-300 bg-danger-500/10 p-2 rounded">
                    <AlertCircle class="w-4 h-4 mr-2 flex-shrink-0" />
                    {{ alert }}
                  </li>
                </ul>
              </div>
              <div v-if="analysis.alerts_info.history_alerts.length > 0">
                <h4 class="text-md font-medium text-white mb-3">历史告警</h4>
                <ul class="space-y-2">
                  <li v-for="alert in analysis.alerts_info.history_alerts" :key="alert" 
                      class="flex items-center text-sm text-gray-400 bg-gray-700/30 p-2 rounded">
                    <CheckCircle class="w-4 h-4 mr-2 flex-shrink-0" />
                    {{ alert }}
                  </li>
                </ul>
              </div>
            </div>
          </div>

          <!-- 资源配置变化历史 -->
          <div class="glass-card p-6" v-if="analysis.resource_analysis.config_history?.length > 0">
            <h3 class="text-lg font-semibold text-white mb-4 flex items-center">
              <GitCommit class="w-5 h-5 mr-2 text-info-400" />
              资源配置变化历史
            </h3>
            <div class="space-y-3">
              <div v-for="(config, index) in analysis.resource_analysis.config_history.slice(0, 5)" 
                   :key="index"
                   class="flex items-center justify-between p-3 bg-gray-800/50 rounded-lg border border-gray-700">
                <div>
                  <div class="text-sm font-medium text-white">
                    CPU: {{ config.cpu_limit }} / 内存: {{ formatBytes(config.memory_limit) }}
                  </div>
                  <div class="text-xs text-gray-400 mt-1">
                    {{ formatDate(config.changed_at) }} · {{ config.reason || '配置更新' }}
                  </div>
                </div>
                <span :class="getConfigChangeClass(config.change_type)" 
                      class="px-2 py-1 text-xs rounded-full">
                  {{ getConfigChangeText(config.change_type) }}
                </span>
              </div>
            </div>
          </div>

          <!-- 性能分析和建议 -->
          <div class="glass-card p-6">
            <h3 class="text-lg font-semibold text-white mb-4 flex items-center">
              <Target class="w-5 h-5 mr-2 text-purple-400" />
              性能分析与优化建议
            </h3>
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <!-- 性能评分 -->
              <div class="space-y-4">
                <h4 class="text-md font-medium text-white">性能评分</h4>
                <div class="space-y-3">
                  <div>
                    <div class="flex justify-between text-sm mb-2">
                      <span class="text-gray-400">资源利用效率</span>
                      <span class="text-white font-medium">{{ getResourceEfficiencyScore() }}/100</span>
                    </div>
                    <div class="w-full bg-gray-700 rounded-full h-2">
                      <div :class="getScoreBarClass(getResourceEfficiencyScore())" 
                           class="h-2 rounded-full transition-all duration-1000"
                           :style="{ width: `${getResourceEfficiencyScore()}%` }"></div>
                    </div>
                  </div>
                  <div>
                    <div class="flex justify-between text-sm mb-2">
                      <span class="text-gray-400">稳定性评分</span>
                      <span class="text-white font-medium">{{ getStabilityScore() }}/100</span>
                    </div>
                    <div class="w-full bg-gray-700 rounded-full h-2">
                      <div :class="getScoreBarClass(getStabilityScore())" 
                           class="h-2 rounded-full transition-all duration-1000"
                           :style="{ width: `${getStabilityScore()}%` }"></div>
                    </div>
                  </div>
                </div>
              </div>
              
              <!-- 优化建议 -->
              <div>
                <h4 class="text-md font-medium text-white mb-3">智能优化建议</h4>
                <div class="space-y-2">
                  <div v-for="suggestion in getOptimizationSuggestions()" 
                       :key="suggestion.id"
                       class="flex items-start p-3 rounded-lg border border-gray-700 bg-gray-800/30">
                    <component :is="suggestion.icon" 
                               :class="suggestion.iconColor" 
                               class="w-4 h-4 mt-0.5 mr-3 flex-shrink-0" />
                    <div class="flex-1">
                      <div class="text-sm font-medium text-white">{{ suggestion.title }}</div>
                      <div class="text-xs text-gray-400 mt-1">{{ suggestion.description }}</div>
                      <div v-if="suggestion.expectedImprovement" 
                           class="text-xs text-success-400 mt-1">
                        预期改善: {{ suggestion.expectedImprovement }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="glass-card p-6" v-if="analysis.comparison_analysis.similar_pods.length > 0">
            <h3 class="text-lg font-semibold text-white mb-4 flex items-center">
              <Users class="w-5 h-5 mr-2 text-info-400" />
              同命名空间相似Pod
            </h3>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead class="text-gray-400 border-b border-gray-700">
                  <tr>
                    <th class="text-left py-2">Pod名称</th>
                    <th class="text-left py-2">内存使用率</th>
                    <th class="text-left py-2">CPU使用率</th>
                    <th class="text-left py-2">状态</th>
                  </tr>
                </thead>
                <tbody class="text-gray-300">
                  <tr v-for="pod in analysis.comparison_analysis.similar_pods.slice(0, 5)" :key="pod.pod_name" 
                      class="border-b border-gray-700/50">
                    <td class="py-2 font-medium text-white">{{ pod.pod_name }}</td>
                    <td class="py-2">{{ pod.memory_req_pct.toFixed(1) }}%</td>
                    <td class="py-2">{{ pod.cpu_req_pct.toFixed(1) }}%</td>
                    <td class="py-2">
                      <span :class="getPodStatusClass(pod.status)" class="px-2 py-1 text-xs rounded-full">
                        {{ pod.status }}
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <!-- 模态框底部 -->
      <div class="flex items-center justify-end gap-3 p-6 border-t border-gray-700">
        <button @click="close" class="btn-secondary">关闭</button>
        <button @click="openTrendModal" class="btn-primary" v-if="analysis">
          <TrendingUp class="w-4 h-4 mr-2" />
          查看趋势图表
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { 
  X, 
  AlertTriangle, 
  Database, 
  Cpu, 
  BarChart, 
  AlertCircle, 
  CheckCircle, 
  Users,
  TrendingUp,
  GitCommit,
  Target,
  Zap,
  Settings,
  TrendingDown,
  Shield
} from 'lucide-vue-next'
import PodsApiService, { type PodDetailAnalysis, type Pod } from '../../api/pods'

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
  openTrend: [cluster: string, namespace: string, podName: string]
}>()

// 响应式数据
const loading = ref(false)
const error = ref<string | null>(null)
const analysis = ref<PodDetailAnalysis | null>(null)

// 监听props变化
watch([() => props.visible, () => props.cluster, () => props.namespace, () => props.podName], 
  ([visible, cluster, namespace, podName]) => {
    if (visible && cluster && namespace && podName) {
      loadAnalysis()
    }
  }, 
  { immediate: true }
)

// 加载分析数据
const loadAnalysis = async () => {
  if (!props.cluster || !props.namespace || !props.podName) return

  try {
    loading.value = true
    error.value = null
    
    const response = await PodsApiService.getPodDetailAnalysis(
      props.cluster, 
      props.namespace, 
      props.podName
    )
    
    if (response.code === 0 && response.data) {
      analysis.value = response.data
    } else {
      throw new Error(response.message || '获取分析数据失败')
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
    console.error('加载Pod详细分析失败:', err)
  } finally {
    loading.value = false
  }
}

// 关闭模态框
const close = () => {
  emit('close')
  analysis.value = null
  error.value = null
}

// 打开趋势模态框
const openTrendModal = () => {
  if (props.cluster && props.namespace && props.podName) {
    emit('openTrend', props.cluster, props.namespace, props.podName)
  }
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

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatMillicores = (millicores: number) => {
  if (millicores >= 1000) {
    return (millicores / 1000).toFixed(2) + ' CPU'
  }
  return millicores + 'm'
}

const getStatusColor = (status: string) => {
  if (status.includes('合理') || status.includes('正常')) {
    return 'text-success-400'
  }
  if (status.includes('不足') || status.includes('风险') || status.includes('过高')) {
    return 'text-danger-400'
  }
  if (status.includes('浪费')) {
    return 'text-warning-400'
  }
  return 'text-gray-300'
}

const getPodStatusClass = (status: string) => {
  if (status === '合理') {
    return 'bg-success-500/20 text-success-400'
  }
  if (status === '不合理') {
    return 'bg-danger-500/20 text-danger-400'
  }
  return 'bg-gray-500/20 text-gray-400'
}

// 新增的辅助方法
const getConfigChangeClass = (changeType: string) => {
  switch (changeType) {
    case 'increase':
    case 'scale_up':
      return 'bg-success-500/20 text-success-400'
    case 'decrease':
    case 'scale_down':
      return 'bg-warning-500/20 text-warning-400'
    case 'optimize':
      return 'bg-info-500/20 text-info-400'
    default:
      return 'bg-gray-500/20 text-gray-400'
  }
}

const getConfigChangeText = (changeType: string) => {
  switch (changeType) {
    case 'increase':
    case 'scale_up':
      return '扩容'
    case 'decrease':
    case 'scale_down':
      return '缩容'
    case 'optimize':
      return '优化'
    default:
      return '变更'
  }
}

// 计算资源效率评分
const getResourceEfficiencyScore = () => {
  if (!analysis.value) return 0
  
  const cpuEfficiency = analysis.value.resource_analysis.cpu_analysis.efficiency_score
  const memoryEfficiency = analysis.value.resource_analysis.memory_analysis.efficiency_score
  
  return Math.round((cpuEfficiency + memoryEfficiency) / 2)
}

// 计算稳定性评分
const getStabilityScore = () => {
  if (!analysis.value) return 0
  
  // 基于告警数量和重启次数计算稳定性
  const alertCount = analysis.value.alerts_info.alert_count || 0
  const restartCount = 0 // 可以从其他地方获取重启次数
  
  let score = 100
  score -= alertCount * 10  // 每个告警扣10分
  score -= restartCount * 5  // 每次重启扣5分
  
  return Math.max(0, Math.min(100, score))
}

// 评分条颜色
const getScoreBarClass = (score: number) => {
  if (score >= 80) return 'bg-gradient-to-r from-success-600 to-success-400'
  if (score >= 60) return 'bg-gradient-to-r from-warning-600 to-warning-400'
  return 'bg-gradient-to-r from-danger-600 to-danger-400'
}

// 生成优化建议
const getOptimizationSuggestions = () => {
  if (!analysis.value) return []
  
  const suggestions = []
  const cpuAnalysis = analysis.value.resource_analysis.cpu_analysis
  const memoryAnalysis = analysis.value.resource_analysis.memory_analysis
  
  // CPU相关建议
  if (cpuAnalysis.efficiency_score < 60) {
    suggestions.push({
      id: 'cpu-optimize',
      icon: Cpu,
      iconColor: 'text-warning-400',
      title: 'CPU资源优化',
      description: 'CPU利用率较低，建议调整资源配置以提高效率',
      expectedImprovement: '预计可节省20-30%的CPU资源'
    })
  }
  
  // 内存相关建议
  if (memoryAnalysis.efficiency_score < 60) {
    suggestions.push({
      id: 'memory-optimize',
      icon: Database,
      iconColor: 'text-primary-400',
      title: '内存资源优化',
      description: '内存使用效率偏低，可以适当减少内存限制',
      expectedImprovement: '预计可节省15-25%的内存资源'
    })
  }
  
  // 告警相关建议
  if (analysis.value.alerts_info.alert_count > 0) {
    suggestions.push({
      id: 'alert-reduce',
      icon: Shield,
      iconColor: 'text-danger-400',
      title: '告警处理',
      description: '存在活跃告警，建议优先处理以提高稳定性',
      expectedImprovement: '减少系统不稳定风险'
    })
  }
  
  // 性能优化建议
  if (getResourceEfficiencyScore() > 90) {
    suggestions.push({
      id: 'performance-tune',
      icon: Zap,
      iconColor: 'text-success-400',
      title: '性能调优',
      description: '资源配置良好，可进一步优化应用性能',
      expectedImprovement: '提升5-10%的响应速度'
    })
  }
  
  // 配置标准化建议
  suggestions.push({
    id: 'config-standardize',
    icon: Settings,
    iconColor: 'text-info-400',
    title: '配置标准化',
    description: '建议采用团队标准的资源配置模板',
    expectedImprovement: '提高配置一致性和可维护性'
  })
  
  return suggestions.slice(0, 4) // 最多显示4个建议
}
</script>