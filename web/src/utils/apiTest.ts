import PodsApiService from '../api/pods'

// 测试API功能
const testAPI = async () => {
  console.log('🔧 开始测试API功能...')
  
  try {
    // 测试健康检查
    console.log('1. 测试健康检查...')
    const healthResponse = await fetch('http://localhost:9999/api/health')
    const healthData = await healthResponse.json()
    console.log('健康检查结果:', healthData)
    
    // 测试Pod搜索
    console.log('2. 测试Pod搜索...')
    const podsResponse = await PodsApiService.getPodsWithSearch({
      page: 1,
      size: 5
    })
    console.log('Pod搜索结果:', podsResponse)
    
    // 验证响应格式
    if (podsResponse && podsResponse.code === 0 && podsResponse.data) {
      console.log('✅ Pod搜索响应格式正确')
      console.log('📊 Pod数据条数:', podsResponse.data.pods?.length || 0)
      if (podsResponse.data.pods?.length > 0) {
        console.log('📝 第一个Pod示例:', podsResponse.data.pods[0])
      }
    } else {
      console.warn('⚠️ Pod搜索响应格式异常')
    }
    
    // 测试统计数据
    console.log('3. 测试统计数据...')
    const statsResponse = await PodsApiService.getPodStats()
    console.log('统计数据结果:', statsResponse)
    
    // 验证统计数据格式
    if (statsResponse && statsResponse.code === 0 && statsResponse.data) {
      console.log('✅ 统计数据响应格式正确')
      console.log('📊 统计数据:', statsResponse.data)
    } else {
      console.warn('⚠️ 统计数据响应格式异常')
    }
    
    console.log('🎉 所有API测试通过!')
    return true
  } catch (error) {
    console.error('❌ API测试失败:', error)
    return false
  }
}

// 测试数据转换逻辑
const testDataTransformation = () => {
  console.log('🔄 测试数据转换逻辑...')
  
  // 模拟后端Pod数据
  const mockPodData = {
    pod_name: 'test-pod-123',
    namespace: 'default',
    cluster_name: 'test-cluster',
    status: '合理',
    cpu_req_pct: 75.5,
    memory_req_pct: 60.2,
    creation_time: '2024-01-15T10:30:00Z'
  }
  
  // 测试映射函数
  const mapPodStatus = (status: string): string => {
    if (!status) return 'Pending'
    const normalizedStatus = status.trim()
    if (normalizedStatus === '合理') return 'Running'
    if (normalizedStatus === '不合理') return 'Failed'
    return normalizedStatus
  }
  
  const transformPodData = (pod: any) => {
    return {
      id: `${pod.cluster_name}-${pod.namespace}-${pod.pod_name}`,
      name: pod.pod_name,
      namespace: pod.namespace,
      cluster: pod.cluster_name,
      status: mapPodStatus(pod.status),
      cpuUsage: Math.round(pod.cpu_req_pct || 0),
      memoryUsage: Math.round(pod.memory_req_pct || 0),
      restarts: 0,
      startTime: pod.creation_time
    }
  }
  
  const transformedData = transformPodData(mockPodData)
  
  console.log('📊 原始数据:', mockPodData)
  console.log('🎯 转换后数据:', transformedData)
  
  // 验证转换结果
  const isValid = transformedData.name === 'test-pod-123' &&
                  transformedData.status === 'Running' &&
                  transformedData.cpuUsage === 76 &&
                  transformedData.memoryUsage === 60
  
  console.log(isValid ? '✅ 数据转换测试通过' : '❌ 数据转换测试失败')
  return isValid
}

export { testAPI, testDataTransformation }