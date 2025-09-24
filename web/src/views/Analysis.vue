<template>
  <div class="space-y-4 animate-fade-in">
    <!-- 页面标题 -->
    <AnalysisHeader 
      :loading="loading"
      @refresh="refreshAllAnalysisData"
      @collect="triggerFullCollection"
    />

    <!-- 深度分析功能区 -->
    <div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
      <!-- 统计卡片区域 -->
      <div class="lg:col-span-3">
        <AnalysisStatsCards
          :top-memory-pods="topMemoryPods"
          :top-cpu-pods="topCpuPods"
          :namespace-summary="namespaceSummary"
          :namespace-sort-by="namespaceSortBy"
          :loading="loading"
          @refresh-memory="fetchTopMemoryPods(20)"
          @refresh-cpu="fetchTopCpuPods(20)"
          @refresh-namespace="refreshNamespaceSummary"
          @namespace-sort-change="handleNamespaceSortChange"
        />
      </div>

      <!-- 活动数据优化卡片 -->
      <ActivityOptimizationCard
        :optimization-result="optimizationResult"
        :loading="loadingOptimization"
        @open-config="toggleOptimizationConfig"
      />
    </div>

    <!-- 活动优化配置弹窗 -->
    <ActivityOptimizationModal
      :show="showOptimizationConfig"
      :config="optimizationConfig"
      :loading="loadingOptimization"
      @close="closeOptimizationConfig"
      @save="saveOptimizationConfig"
    />

    <!-- 资源分布图表 -->
    <ResourceChartsSection />

    <!-- 问题Pod详细分析表格 -->
    <ProblemPodsTable
      :problems="sortedProblems"
      :pagination="pagination"
      :available-clusters="availableClusters"
      :loading="loading"
      :sort-by="sortBy"
      :selected-cluster="selectedCluster"
      :page-size="pageSize"
      @sort-change="handleSortChange"
      @cluster-change="handleClusterChange"
      @page-size-change="handlePageSizeChange"
      @page-change="goToPage"
      @refresh="refreshProblems"
    />

    <!-- 空状态 -->
    <div v-if="!analysisData && !loading" class="text-center py-12">
      <BarChart3 class="w-16 h-16 text-gray-600 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-400 mb-2">暂无分析数据</h3>
      <p class="text-gray-500 mb-6">点击刷新按钮开始资源分析</p>
      <button 
        class="btn-primary"
        @click="refreshAllAnalysisData"
      >
        <RefreshCw class="w-4 h-4 mr-2" />
        开始分析
      </button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && !analysisData" class="text-center py-12">
      <RefreshCw class="w-16 h-16 text-primary-500 mx-auto mb-4 animate-spin" />
      <h3 class="text-lg font-semibold text-gray-400 mb-2">正在分析资源配置</h3>
      <p class="text-gray-500">请稍等，正在收集和分析集群数据...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RefreshCw, BarChart3 } from 'lucide-vue-next'
import { useAnalysis, useClusters, useActivityOptimization } from '../composables/api'
import { extractClusterNames } from '../utils/analysis'

// 导入组件
import AnalysisHeader from '../components/analysis/AnalysisHeader.vue'
import AnalysisStatsCards from '../components/analysis/AnalysisStatsCards.vue'
import ActivityOptimizationCard from '../components/analysis/ActivityOptimizationCard.vue'
import ActivityOptimizationModal from '../components/analysis/ActivityOptimizationModal.vue'
import ResourceChartsSection from '../components/analysis/ResourceChartsSection.vue'
import ProblemPodsTable from '../components/analysis/ProblemPodsTable.vue'

// 使用API组合函数
const { 
  analysis: analysisData, 
  topMemoryPods,
  topCpuPods,
  namespaceSummary,
  pagination,
  loading, 
  fetchAnalysis,
  fetchProblemsWithPagination,
  fetchTopMemoryPods,
  fetchTopCpuPods,
  fetchNamespaceSummary,
  triggerDataCollection,
  refreshAllData
} = useAnalysis()

const { clusters, fetchClusters } = useClusters()

// 活动优化相关
const {
  config: optimizationConfig,
  optimizationResult,
  loading: loadingOptimization,
  fetchOptimizationConfig,
  updateOptimizationConfig,
  executeOptimization: performOptimization
} = useActivityOptimization()

// 本地状态
const showOptimizationConfig = ref(false)
const sortBy = ref('total_waste')
const selectedCluster = ref('')
const pageSize = ref(10)
const namespaceSortBy = ref('combined')  // 新增：命名空间排序状态

// 计算属性
const availableClusters = computed(() => {
  return extractClusterNames(
    clusters.value || [],
    analysisData.value?.top50_problems || [],
    topMemoryPods.value || [],
    topCpuPods.value || [],
    namespaceSummary.value || []
  )
})

const sortedProblems = computed(() => {
  return analysisData.value?.top50_problems || []
})

// 活动优化相关方法
const toggleOptimizationConfig = async () => {
  if (!showOptimizationConfig.value) {
    try {
      await fetchOptimizationConfig()
      showOptimizationConfig.value = true
    } catch (error) {
      console.error('获取优化配置失败:', error)
    }
  } else {
    showOptimizationConfig.value = false
  }
}

const closeOptimizationConfig = () => {
  showOptimizationConfig.value = false
}

const saveOptimizationConfig = async (config: any) => {
  try {
    await updateOptimizationConfig(config)
    showOptimizationConfig.value = false
  } catch (error) {
    console.error('保存配置失败:', error)
  }
}

// 静默执行优化操作，不影响用户界面
const executeOptimizationSilently = async () => {
  try {
    // 后台静默执行告警去重，不显示loading状态
    await performOptimization()
  } catch (error) {
    // 优化失败不影响主要功能，仅记录错误
    console.warn('后台优化执行失败:', error)
  }
}

const executeOptimization = async () => {
  try {
    await performOptimization()
  } catch (error) {
    console.error('执行优化失败:', error)
  }
}

// 命名空间排序相关方法
const refreshNamespaceSummary = async () => {
  await fetchNamespaceSummary(10, namespaceSortBy.value)
}

const handleNamespaceSortChange = async (newSortBy: string) => {
  namespaceSortBy.value = newSortBy
  await fetchNamespaceSummary(10, newSortBy)
}

// 数据刷新方法 - 集成自动去重优化
const refreshAllAnalysisData = async () => {
  await Promise.allSettled([
    fetchTopMemoryPods(20),
    fetchTopCpuPods(20),
    fetchNamespaceSummary(10, namespaceSortBy.value)
  ])
  
  // 自动执行告警去重优化，提升数据质量
  await executeOptimizationSilently()
}

const refreshProblems = async () => {
  await fetchProblemsWithPagination(
    pagination.value.page, 
    pageSize.value, 
    selectedCluster.value,
    sortBy.value
  )
  
  if (!analysisData.value) {
    await fetchAnalysis().catch(() => {})
  }
  
  // 在问题数据刷新后自动优化
  await executeOptimizationSilently()
}

// 分页和筛选处理方法
const goToPage = async (page: number) => {
  if (page < 1 || page > pagination.value.total_pages || loading.value) return
  await fetchProblemsWithPagination(page, pageSize.value, selectedCluster.value, sortBy.value)
}

const handlePageSizeChange = async (size: number) => {
  pageSize.value = size
  await fetchProblemsWithPagination(1, size, selectedCluster.value, sortBy.value)
}

const handleClusterChange = async (cluster: string) => {
  selectedCluster.value = cluster
  await fetchProblemsWithPagination(1, pageSize.value, cluster, sortBy.value)
}

const handleSortChange = async (sort: string) => {
  sortBy.value = sort
  await fetchProblemsWithPagination(1, pageSize.value, selectedCluster.value, sort)
}

const triggerFullCollection = async () => {
  await triggerDataCollection(true)
  await refreshAllData()
  await fetchProblemsWithPagination(1, pageSize.value, selectedCluster.value, sortBy.value)
  
  // 数据收集完成后自动执行优化
  await executeOptimizationSilently()
}

// 组件挂载时初始化数据
onMounted(async () => {
  await fetchClusters().catch(() => {})
  await fetchProblemsWithPagination(1, pageSize.value, selectedCluster.value, sortBy.value)
  
  await Promise.allSettled([
    fetchTopMemoryPods(20),
    fetchTopCpuPods(20),
    fetchNamespaceSummary(10, namespaceSortBy.value),
    fetchAnalysis(),
    fetchOptimizationConfig()
  ])
  
  // 初始化完成后自动执行一次优化
  await executeOptimizationSilently()
})
</script>