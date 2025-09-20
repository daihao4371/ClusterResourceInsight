import { api } from '../utils/api'
import type { ApiResponse, Cluster } from '../types'

/**
 * 集群API接口模块
 * 负责集群相关的所有后端API调用
 */

// 获取集群列表
export const getClusters = async () => {
  const response = await api.get<ApiResponse<{
    count: number
    data: any[]
  }>>('/clusters')
  
  // 将后端数据格式转换为前端期望的格式
  const clusters: Cluster[] = response.data.data.data.map((backendCluster: any) => ({
    id: backendCluster.id.toString(),
    name: backendCluster.cluster_name,
    alias: backendCluster.cluster_alias || backendCluster.cluster_name,
    endpoint: backendCluster.api_server,
    status: backendCluster.status || 'unknown',
    auth_type: backendCluster.auth_type,
    collect_interval: backendCluster.collect_interval,
    created_at: backendCluster.created_at,
    updated_at: backendCluster.updated_at,
    last_collect_at: backendCluster.last_collect_at,
    // 默认值，将通过测试接口获取实时数据
    region: 'unknown',
    nodes: 0,
    pods: 0,
    cpuUsage: 0,
    memoryUsage: 0,
    tags: backendCluster.tags ? JSON.parse(backendCluster.tags) : []
  }))
  
  return clusters
}

// 获取集群实时统计数据
export const getClustersWithStats = async () => {
  const clusters = await getClusters()
  console.log('获取基础集群数据:', clusters)
  
  // 为每个集群获取实时统计数据
  const clustersWithStats = await Promise.all(
    clusters.map(async (cluster) => {
      try {
        console.log(`正在获取集群 ${cluster.name} 的统计数据...`)
        const stats = await testCluster(parseInt(cluster.id))
        console.log(`集群 ${cluster.name} 统计数据:`, stats)
        
        const updatedCluster = {
          ...cluster,
          nodes: stats.node_count || 0,
          pods: stats.pod_count || 0,
          status: stats.status || cluster.status,
          // 添加CPU和内存使用率数据
          cpuUsage: stats.cpu_usage || 0,
          memoryUsage: stats.memory_usage || 0,
          // 添加详细的资源信息
          cpuUsedCores: stats.cpu_used_cores || 0,
          cpuTotalCores: stats.cpu_total_cores || 0,
          memoryUsedGB: stats.memory_used_gb || 0,
          memoryTotalGB: stats.memory_total_gb || 0,
          hasRealUsage: stats.has_real_usage || false,
          dataSource: stats.data_source || 'none',
          // 可以根据实际情况添加更多统计信息
          version: stats.version,
          namespace_count: stats.namespace_count,
          response_time_ms: stats.response_time_ms,
          has_metrics: stats.has_metrics,
          last_test_time: stats.test_time
        }
        
        console.log(`集群 ${cluster.name} 更新后的数据:`, {
          name: updatedCluster.name,
          nodes: updatedCluster.nodes,
          pods: updatedCluster.pods,
          status: updatedCluster.status
        })
        
        return updatedCluster
      } catch (error) {
        console.warn(`获取集群 ${cluster.name} 统计数据失败:`, error)
        // 如果获取统计数据失败，保持原有数据
        return cluster
      }
    })
  )
  
  console.log('最终集群数据:', clustersWithStats)
  return clustersWithStats
}

// 测试集群连接
export const testCluster = async (clusterId: number) => {
  const response = await api.post<ApiResponse<any>>(`/clusters/${clusterId}/test`)
  console.log(`集群${clusterId}测试API原始响应:`, response)
  console.log(`集群${clusterId}测试API数据层级:`, response.data)
  console.log(`集群${clusterId}最终数据:`, response.data.data.data)
  // 根据用户提供的实际API响应结构，数据在 response.data.data.data 中
  return response.data.data.data
}

// 添加新集群
export const addCluster = async (clusterData: {
  cluster_name: string
  cluster_alias?: string
  api_server: string
  auth_type: string
  auth_config: {
    bearer_token?: string
    client_cert?: string
    client_key?: string
    ca_cert?: string
    kubeconfig?: string
  }
  collect_interval?: number
  tags?: string[]
}) => {
  const response = await api.post<ApiResponse<any>>('/clusters', clusterData)
  return response.data.data
}

// 更新集群配置
export const updateCluster = async (clusterId: number, clusterData: any) => {
  const response = await api.put<ApiResponse<any>>(`/clusters/${clusterId}`, clusterData)
  return response.data.data
}

// 删除集群
export const deleteCluster = async (clusterId: number) => {
  const response = await api.delete<ApiResponse<any>>(`/clusters/${clusterId}`)
  return response.data.data
}

// 测试集群连接配置（创建前验证）
export const testClusterConfig = async (clusterData: {
  cluster_name: string
  api_server: string
  auth_type: string
  auth_config: {
    bearer_token?: string
    client_cert?: string
    client_key?: string
    ca_cert?: string
    kubeconfig?: string
  }
}) => {
  const response = await api.post<ApiResponse<any>>('/clusters/test', clusterData)
  return response.data.data.data
}

// 获取集群详细信息
export const getClusterDetail = async (clusterId: number) => {
  const response = await api.get<ApiResponse<any>>(`/clusters/${clusterId}`)
  return response.data.data
}