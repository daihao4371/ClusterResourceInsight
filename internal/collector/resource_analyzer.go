package collector

import (
	"context"
	"fmt"
	"time"
)

// 分析Pod资源使用情况
func (rc *ResourceCollector) analyzeResourceUsage(pods []PodResourceInfo) *AnalysisResult {
	var unreasonablePods []PodResourceInfo

	for i := range pods {
		pod := &pods[i]
		issues := []string{}

		// 检查内存利用率 (现在总是有使用数据，包括估算值)
		if pod.MemoryRequest > 0 && pod.MemoryReqPct > 0 && pod.MemoryReqPct < 20 {
			issues = append(issues, "内存请求利用率过低")
		}
		if pod.MemoryLimit > 0 && pod.MemoryLimitPct > 0 && pod.MemoryLimitPct < 15 {
			issues = append(issues, "内存限制利用率过低")
		}

		// 检查 CPU 利用率 (现在总是有使用数据，包括估算值)
		if pod.CPURequest > 0 && pod.CPUReqPct > 0 && pod.CPUReqPct < 15 {
			issues = append(issues, "CPU请求利用率过低")
		}
		if pod.CPULimit > 0 && pod.CPULimitPct > 0 && pod.CPULimitPct < 10 {
			issues = append(issues, "CPU限制利用率过低")
		}

		// 检查配置缺失
		if pod.MemoryRequest == 0 {
			issues = append(issues, "缺少内存请求配置")
		}
		if pod.CPURequest == 0 {
			issues = append(issues, "缺少CPU请求配置")
		}

		// 检查请求和限制差异过大
		if pod.MemoryLimit > 0 && pod.MemoryRequest > 0 {
			ratio := float64(pod.MemoryLimit) / float64(pod.MemoryRequest)
			if ratio > 3.0 {
				issues = append(issues, "内存请求和限制差异过大")
			}
		}
		if pod.CPULimit > 0 && pod.CPURequest > 0 {
			ratio := float64(pod.CPULimit) / float64(pod.CPURequest)
			if ratio > 3.0 {
				issues = append(issues, "CPU请求和限制差异过大")
			}
		}

		// 移除之前的 Metrics 依赖检查，因为现在总是有数据

		if len(issues) > 0 {
			pod.Status = "不合理"
			pod.Issues = issues
			unreasonablePods = append(unreasonablePods, *pod)
		}
	}

	// 按问题严重程度排序（利用率最低的排在前面）
	rc.sortPodsByProblemSeverity(unreasonablePods)

	// 取前50个
	top50 := unreasonablePods
	if len(unreasonablePods) > 50 {
		top50 = unreasonablePods[:50]
	}

	return &AnalysisResult{
		TotalPods:        len(pods),
		UnreasonablePods: len(unreasonablePods),
		Top50Problems:    top50,
		GeneratedAt:      time.Now(),
	}
}

// analyzeMultiClusterData 分析多集群数据
func (mc *MultiClusterResourceCollector) analyzeMultiClusterData(pods []PodResourceInfo) *AnalysisResult {
	var unreasonablePods []PodResourceInfo

	for i := range pods {
		pod := &pods[i]
		issues := []string{}

		// 检查内存利用率 (现在总是有使用数据，包括估算值)
		if pod.MemoryRequest > 0 && pod.MemoryReqPct > 0 && pod.MemoryReqPct < 20 {
			issues = append(issues, "内存请求利用率过低")
		}
		if pod.MemoryLimit > 0 && pod.MemoryLimitPct > 0 && pod.MemoryLimitPct < 15 {
			issues = append(issues, "内存限制利用率过低")
		}

		// 检查 CPU 利用率 (现在总是有使用数据，包括估算值)
		if pod.CPURequest > 0 && pod.CPUReqPct > 0 && pod.CPUReqPct < 15 {
			issues = append(issues, "CPU请求利用率过低")
		}
		if pod.CPULimit > 0 && pod.CPULimitPct > 0 && pod.CPULimitPct < 10 {
			issues = append(issues, "CPU限制利用率过低")
		}

		// 检查配置缺失
		if pod.MemoryRequest == 0 {
			issues = append(issues, "缺少内存请求配置")
		}
		if pod.CPURequest == 0 {
			issues = append(issues, "缺少CPU请求配置")
		}

		// 检查请求和限制差异过大
		if pod.MemoryLimit > 0 && pod.MemoryRequest > 0 {
			ratio := float64(pod.MemoryLimit) / float64(pod.MemoryRequest)
			if ratio > 3.0 {
				issues = append(issues, "内存请求和限制差异过大")
			}
		}
		if pod.CPULimit > 0 && pod.CPURequest > 0 {
			ratio := float64(pod.CPULimit) / float64(pod.CPURequest)
			if ratio > 3.0 {
				issues = append(issues, "CPU请求和限制差异过大")
			}
		}

		// 移除之前的 Metrics 依赖检查，因为现在总是有数据

		if len(issues) > 0 {
			pod.Status = "不合理"
			pod.Issues = issues
			unreasonablePods = append(unreasonablePods, *pod)
		}
	}

	// 按问题严重程度排序（利用率最低的排在前面）
	mc.sortPodsByProblemSeverity(unreasonablePods)

	// 取前50个
	top50 := unreasonablePods
	if len(unreasonablePods) > 50 {
		top50 = unreasonablePods[:50]
	}

	return &AnalysisResult{
		TotalPods:        len(pods),
		UnreasonablePods: len(unreasonablePods),
		Top50Problems:    top50,
		GeneratedAt:      time.Now(),
	}
}

// 按问题严重程度排序Pod
func (rc *ResourceCollector) sortPodsByProblemSeverity(pods []PodResourceInfo) {
	// 简单的排序：按内存和CPU的最低利用率排序
	for i := 0; i < len(pods)-1; i++ {
		for j := i + 1; j < len(pods); j++ {
			// 计算问题严重程度分数（利用率越低分数越高）
			scoreI := rc.calculateProblemScore(pods[i])
			scoreJ := rc.calculateProblemScore(pods[j])

			if scoreI < scoreJ {
				pods[i], pods[j] = pods[j], pods[i]
			}
		}
	}
}

// 多集群版本的排序方法
func (mc *MultiClusterResourceCollector) sortPodsByProblemSeverity(pods []PodResourceInfo) {
	// 简单的排序：按内存和CPU的最低利用率排序
	for i := 0; i < len(pods)-1; i++ {
		for j := i + 1; j < len(pods); j++ {
			// 计算问题严重程度分数（利用率越低分数越高）
			scoreI := mc.calculateProblemScore(pods[i])
			scoreJ := mc.calculateProblemScore(pods[j])

			if scoreI < scoreJ {
				pods[i], pods[j] = pods[j], pods[i]
			}
		}
	}
}

// 计算Pod问题严重程度分数
func (rc *ResourceCollector) calculateProblemScore(pod PodResourceInfo) float64 {
	score := 0.0

	// 内存利用率问题得分
	if pod.MemoryRequest > 0 {
		score += (20.0 - pod.MemoryReqPct) / 20.0 * 100
	}
	if pod.MemoryLimit > 0 {
		score += (15.0 - pod.MemoryLimitPct) / 15.0 * 100
	}

	// CPU 利用率问题得分
	if pod.CPURequest > 0 {
		score += (15.0 - pod.CPUReqPct) / 15.0 * 100
	}
	if pod.CPULimit > 0 {
		score += (10.0 - pod.CPULimitPct) / 10.0 * 100
	}

	// 配置缺失问题得分
	if pod.MemoryRequest == 0 {
		score += 200
	}
	if pod.CPURequest == 0 {
		score += 200
	}

	return score
}

// GetTopMemoryRequestPods 获取内存请求量最大的前N个Pod
func (mc *MultiClusterResourceCollector) GetTopMemoryRequestPods(ctx context.Context, limit int) ([]PodResourceInfo, error) {
	// 收集所有Pod并按内存请求量排序
	allPods := []PodResourceInfo{}

	clusters, err := mc.clusterService.GetAllClusters()
	if err != nil {
		return nil, fmt.Errorf("获取集群列表失败: %v", err)
	}

	for _, cluster := range clusters {
		if cluster.Status != "online" {
			continue
		}

		kubeClient, metricsClient, err := mc.clusterService.CreateKubernetesClient(&cluster)
		if err != nil {
			continue
		}

		singleCollector := &ResourceCollector{
			kubeClient:    kubeClient,
			metricsClient: metricsClient,
		}

		clusterPods, err := singleCollector.collectAllPodsWithoutFiltering(ctx, cluster.ClusterName)
		if err != nil {
			continue
		}

		allPods = append(allPods, clusterPods...)
	}

	// 按内存请求量排序（从大到小）
	for i := 0; i < len(allPods)-1; i++ {
		for j := i + 1; j < len(allPods); j++ {
			if allPods[i].MemoryRequest < allPods[j].MemoryRequest {
				allPods[i], allPods[j] = allPods[j], allPods[i]
			}
		}
	}

	// 返回前N个
	if len(allPods) > limit {
		return allPods[:limit], nil
	}
	return allPods, nil
}

// GetTopCPURequestPods 获取CPU请求量最大的前N个Pod
func (mc *MultiClusterResourceCollector) GetTopCPURequestPods(ctx context.Context, limit int) ([]PodResourceInfo, error) {
	// 收集所有Pod并按CPU请求量排序
	allPods := []PodResourceInfo{}

	clusters, err := mc.clusterService.GetAllClusters()
	if err != nil {
		return nil, fmt.Errorf("获取集群列表失败: %v", err)
	}

	for _, cluster := range clusters {
		if cluster.Status != "online" {
			continue
		}

		kubeClient, metricsClient, err := mc.clusterService.CreateKubernetesClient(&cluster)
		if err != nil {
			continue
		}

		singleCollector := &ResourceCollector{
			kubeClient:    kubeClient,
			metricsClient: metricsClient,
		}

		clusterPods, err := singleCollector.collectAllPodsWithoutFiltering(ctx, cluster.ClusterName)
		if err != nil {
			continue
		}

		allPods = append(allPods, clusterPods...)
	}

	// 按CPU请求量排序（从大到小）
	for i := 0; i < len(allPods)-1; i++ {
		for j := i + 1; j < len(allPods); j++ {
			if allPods[i].CPURequest < allPods[j].CPURequest {
				allPods[i], allPods[j] = allPods[j], allPods[i]
			}
		}
	}

	// 返回前N个
	if len(allPods) > limit {
		return allPods[:limit], nil
	}
	return allPods, nil
}

// GetNamespaceTreeData 获取命名空间的树状数据
func (mc *MultiClusterResourceCollector) GetNamespaceTreeData(ctx context.Context, namespace string) (*NamespaceTreeData, error) {
	pods, err := mc.GetNamespacePods(ctx, namespace)
	if err != nil {
		return nil, err
	}

	if len(pods) == 0 {
		return &NamespaceTreeData{
			NamespaceName: namespace,
			Children:      []PodResourceInfo{},
			Summary:       NamespaceSummary{NamespaceName: namespace},
		}, nil
	}

	// 计算汇总信息
	summary := NamespaceSummary{
		NamespaceName: namespace,
		ClusterName:   pods[0].ClusterName,
		TotalPods:     len(pods),
	}

	for _, pod := range pods {
		summary.TotalMemoryUsage += pod.MemoryUsage
		summary.TotalCPUUsage += pod.CPUUsage
		summary.TotalMemoryRequest += pod.MemoryRequest
		summary.TotalCPURequest += pod.CPURequest

		if pod.Status == "不合理" {
			summary.UnreasonablePods++
		}
	}

	return &NamespaceTreeData{
		NamespaceName: namespace,
		ClusterName:   summary.ClusterName,
		Children:      pods,
		Summary:       summary,
	}, nil
}

func (mc *MultiClusterResourceCollector) calculateProblemScore(pod PodResourceInfo) float64 {
	score := 0.0

	// 内存利用率问题得分
	if pod.MemoryRequest > 0 {
		score += (20.0 - pod.MemoryReqPct) / 20.0 * 100
	}
	if pod.MemoryLimit > 0 {
		score += (15.0 - pod.MemoryLimitPct) / 15.0 * 100
	}

	// CPU 利用率问题得分
	if pod.CPURequest > 0 {
		score += (15.0 - pod.CPUReqPct) / 15.0 * 100
	}
	if pod.CPULimit > 0 {
		score += (10.0 - pod.CPULimitPct) / 10.0 * 100
	}

	// 配置缺失问题得分
	if pod.MemoryRequest == 0 {
		score += 200
	}
	if pod.CPURequest == 0 {
		score += 200
	}

	return score
}
