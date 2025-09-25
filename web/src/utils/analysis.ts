import type { Pod } from '../types'

/**
 * 格式化CPU值
 */
export const formatCpuValue = (value: number | string | undefined): string => {
  if (!value || value === 0) return 'N/A'
  if (typeof value === 'string') return value
  if (value >= 1000) {
    return `${(value / 1000).toFixed(2)} cores`
  }
  return `${value}m`
}

/**
 * 格式化内存值
 */
export const formatMemoryValue = (value: number | string | undefined): string => {
  if (!value || value === 0) return 'N/A'
  if (typeof value === 'string') return value
  
  const bytes = typeof value === 'number' ? value : parseInt(value)
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let unitIndex = 0
  let size = bytes
  
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  
  return `${size.toFixed(1)} ${units[unitIndex]}`
}

/**
 * 计算Pod资源浪费程度
 */
export const calculateWaste = (pod: Pod): number => {
  if (!pod.cpu_req_pct && !pod.memory_req_pct) return 0
  
  const cpuWaste = pod.cpu_req_pct ? Math.max(0, 100 - pod.cpu_req_pct) : 0
  const memoryWaste = pod.memory_req_pct ? Math.max(0, 100 - pod.memory_req_pct) : 0
  
  return Math.round((cpuWaste + memoryWaste) / 2)
}

/**
 * 获取排名徽章样式
 */
export const getRankBadgeClass = (rank: number): string => {
  if (rank <= 3) return 'bg-danger-500/20 text-danger-400 border border-danger-500/30'
  if (rank <= 10) return 'bg-warning-500/20 text-warning-400 border border-warning-500/30'
  return 'bg-gray-500/20 text-gray-400 border border-gray-500/30'
}

/**
 * 获取问题徽章样式
 */
export const getIssueBadgeClass = (issue: string): string => {
  if (issue.includes('利用率过低') || issue.includes('under')) return 'bg-warning-500/20 text-warning-400'
  if (issue.includes('缺少') || issue.includes('no_')) return 'bg-danger-500/20 text-danger-400'
  if (issue.includes('差异过大')) return 'bg-orange-500/20 text-orange-400'
  return 'bg-primary-500/20 text-primary-400'
}

/**
 * 获取使用率进度条样式
 */
export const getUsageBarClass = (percentage: number): string => {
  if (percentage >= 80) return 'bg-danger-500'
  if (percentage >= 60) return 'bg-warning-500'
  if (percentage >= 40) return 'bg-success-500'
  return 'bg-primary-500'
}

/**
 * 获取浪费程度文本样式
 */
export const getWasteClass = (waste: number): string => {
  if (waste >= 50) return 'text-danger-400'
  if (waste >= 30) return 'text-warning-400'
  return 'text-success-400'
}

/**
 * 从多个数据源提取集群名称 - 提供更灵活的集群名称提取逻辑
 */
export const extractClusterNames = (
  clusters: any[],
  pods: any[],
  topMemoryPods: any[],
  topCpuPods: any[],
  namespaceSummary: any[]
): string[] => {
  const clusterNames = new Set<string>()
  
  console.log('开始提取集群名称，数据源:', {
    clustersCount: clusters?.length || 0,
    podsCount: pods?.length || 0,
    topMemoryPodsCount: topMemoryPods?.length || 0,
    topCpuPodsCount: topCpuPods?.length || 0,
    namespaceSummaryCount: namespaceSummary?.length || 0
  })
  
  // 从集群配置API获取（优先级最高）
  if (Array.isArray(clusters)) {
    clusters.forEach(cluster => {
      // 支持多种可能的字段名
      const name = cluster.cluster_name || cluster.name || cluster.clusterName || cluster.clusterId
      if (name) {
        clusterNames.add(name)
        console.log('从集群配置API获取集群名称:', name)
      }
    })
  }
  
  // 从各种Pod数据源获取
  const allPodSources = [
    { name: 'problems', data: pods },
    { name: 'topMemoryPods', data: topMemoryPods },
    { name: 'topCpuPods', data: topCpuPods }
  ]
  
  allPodSources.forEach(source => {
    if (Array.isArray(source.data)) {
      source.data.forEach(pod => {
        const clusterName = pod.cluster_name || pod.cluster || pod.clusterName
        if (clusterName) {
          clusterNames.add(clusterName)
          console.log(`从${source.name}获取集群名称:`, clusterName)
        }
      })
    }
  })
  
  // 从命名空间汇总获取
  if (Array.isArray(namespaceSummary)) {
    namespaceSummary.forEach(ns => {
      const clusterName = ns.cluster_name || ns.cluster || ns.clusterName
      if (clusterName) {
        clusterNames.add(clusterName)
        console.log('从命名空间汇总获取集群名称:', clusterName)
      }
    })
  }
  
  const result = Array.from(clusterNames).sort()
  console.log('集群名称提取完成，共找到', result.length, '个集群:', result)
  
  return result
}