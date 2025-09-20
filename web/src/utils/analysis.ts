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
 * 从多个数据源提取集群名称
 */
export const extractClusterNames = (
  clusters: any[],
  pods: any[],
  topMemoryPods: any[],
  topCpuPods: any[],
  namespaceSummary: any[]
): string[] => {
  const clusterNames = new Set<string>()
  
  // 从集群配置API获取
  if (Array.isArray(clusters)) {
    clusters.forEach(cluster => {
      const name = cluster.name || cluster.cluster_name || cluster.clusterName
      if (name) clusterNames.add(name)
    })
  }
  
  // 从各种Pod数据源获取
  const allPodSources = [pods, topMemoryPods, topCpuPods]
  allPodSources.forEach(source => {
    if (Array.isArray(source)) {
      source.forEach(pod => {
        if (pod.cluster_name) clusterNames.add(pod.cluster_name)
      })
    }
  })
  
  // 从命名空间汇总获取
  if (Array.isArray(namespaceSummary)) {
    namespaceSummary.forEach(ns => {
      if (ns.cluster_name) clusterNames.add(ns.cluster_name)
    })
  }
  
  return Array.from(clusterNames).sort()
}