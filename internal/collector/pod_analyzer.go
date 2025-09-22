package collector

import (
	"context"
	"fmt"
	"time"
	
	"cluster-resource-insight/pkg/utils"
)

// PodAnalysisHelper Pod分析辅助器 - 提供Pod详细分析和趋势数据的辅助方法
type PodAnalysisHelper struct {
	collector *MultiClusterResourceCollector
}

// NewPodAnalysisHelper 创建Pod分析辅助器
func NewPodAnalysisHelper(collector *MultiClusterResourceCollector) *PodAnalysisHelper {
	return &PodAnalysisHelper{
		collector: collector,
	}
}

// FindPodByIdentifier 根据集群、命名空间和Pod名称查找Pod
func (helper *PodAnalysisHelper) FindPodByIdentifier(ctx context.Context, clusterName, namespace, podName string) (*PodResourceInfo, error) {
	// 通过搜索功能查找Pod
	searchReq := PodSearchRequest{
		Query:     podName,
		Namespace: namespace,
		Cluster:   clusterName,
		Page:      1,
		Size:      100,
	}

	searchResult, err := helper.collector.SearchPods(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	// 查找精确匹配的Pod
	for _, pod := range searchResult.Pods {
		if pod.PodName == podName && pod.Namespace == namespace && pod.ClusterName == clusterName {
			return &pod, nil
		}
	}

	return nil, fmt.Errorf("未找到Pod: %s/%s/%s", clusterName, namespace, podName)
}

// GetSimilarPodsInNamespace 获取同命名空间下的相似Pod
func (helper *PodAnalysisHelper) GetSimilarPodsInNamespace(ctx context.Context, clusterName, namespace string) ([]PodResourceInfo, error) {
	searchReq := PodSearchRequest{
		Namespace: namespace,
		Cluster:   clusterName,
		Page:      1,
		Size:      50, // 限制数量
	}

	searchResult, err := helper.collector.SearchPods(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	return searchResult.Pods, nil
}

// CalculateNamespaceAverage 计算命名空间平均资源使用率
func (helper *PodAnalysisHelper) CalculateNamespaceAverage(pods []PodResourceInfo) struct {
	MemoryUsagePct float64 `json:"memory_usage_pct"`
	CPUUsagePct    float64 `json:"cpu_usage_pct"`
} {
	if len(pods) == 0 {
		return struct {
			MemoryUsagePct float64 `json:"memory_usage_pct"`
			CPUUsagePct    float64 `json:"cpu_usage_pct"`
		}{0, 0}
	}

	var totalMemory, totalCPU float64
	for _, pod := range pods {
		totalMemory += pod.MemoryReqPct
		totalCPU += pod.CPUReqPct
	}

	return struct {
		MemoryUsagePct float64 `json:"memory_usage_pct"`
		CPUUsagePct    float64 `json:"cpu_usage_pct"`
	}{
		totalMemory / float64(len(pods)),
		totalCPU / float64(len(pods)),
	}
}

// CalculateClusterAverage 计算集群平均资源使用率
func (helper *PodAnalysisHelper) CalculateClusterAverage(allPods []PodResourceInfo) struct {
	MemoryUsagePct float64 `json:"memory_usage_pct"`
	CPUUsagePct    float64 `json:"cpu_usage_pct"`
} {
	if len(allPods) == 0 {
		return struct {
			MemoryUsagePct float64 `json:"memory_usage_pct"`
			CPUUsagePct    float64 `json:"cpu_usage_pct"`
		}{0, 0}
	}

	var totalMemory, totalCPU float64
	for _, pod := range allPods {
		totalMemory += pod.MemoryReqPct
		totalCPU += pod.CPUReqPct
	}

	return struct {
		MemoryUsagePct float64 `json:"memory_usage_pct"`
		CPUUsagePct    float64 `json:"cpu_usage_pct"`
	}{
		totalMemory / float64(len(allPods)),
		totalCPU / float64(len(allPods)),
	}
}

// EvaluateMemoryConfigStatus 评估内存配置状态
func (helper *PodAnalysisHelper) EvaluateMemoryConfigStatus(pod *PodResourceInfo) string {
	if pod.MemoryReqPct < 30 {
		return "资源配置过高，存在浪费"
	} else if pod.MemoryReqPct > 90 {
		return "资源配置不足，存在风险"
	}
	return "资源配置合理"
}

// CalculateMemoryEfficiency 计算内存使用效率
func (helper *PodAnalysisHelper) CalculateMemoryEfficiency(pod *PodResourceInfo) float64 {
	if pod.MemoryRequest == 0 {
		return 0
	}
	return (float64(pod.MemoryUsage) / float64(pod.MemoryRequest)) * 100
}

// CalculateMemoryWaste 计算内存浪费量
func (helper *PodAnalysisHelper) CalculateMemoryWaste(pod *PodResourceInfo) int64 {
	if pod.MemoryRequest > pod.MemoryUsage {
		return pod.MemoryRequest - pod.MemoryUsage
	}
	return 0
}

// GenerateMemoryRecommendations 生成内存优化建议
func (helper *PodAnalysisHelper) GenerateMemoryRecommendations(pod *PodResourceInfo) []string {
	var recommendations []string

	if pod.MemoryReqPct < 30 {
		recommendations = append(recommendations, "建议降低内存请求量，当前使用率偏低")
	}
	if pod.MemoryReqPct > 90 {
		recommendations = append(recommendations, "建议增加内存请求量，避免OOM风险")
	}
	if pod.MemoryLimit > 0 && pod.MemoryLimitPct > 85 {
		recommendations = append(recommendations, "建议增加内存限制量，当前接近上限")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "当前内存配置较为合理")
	}

	return recommendations
}

// EvaluateCPUConfigStatus 评估CPU配置状态
func (helper *PodAnalysisHelper) EvaluateCPUConfigStatus(pod *PodResourceInfo) string {
	if pod.CPUReqPct < 20 {
		return "CPU配置过高，存在浪费"
	} else if pod.CPUReqPct > 95 {
		return "CPU配置不足，可能影响性能"
	}
	return "CPU配置合理"
}

// CalculateCPUEfficiency 计算CPU使用效率
func (helper *PodAnalysisHelper) CalculateCPUEfficiency(pod *PodResourceInfo) float64 {
	if pod.CPURequest == 0 {
		return 0
	}
	return (float64(pod.CPUUsage) / float64(pod.CPURequest)) * 100
}

// CalculateCPUWaste 计算CPU浪费量
func (helper *PodAnalysisHelper) CalculateCPUWaste(pod *PodResourceInfo) int64 {
	if pod.CPURequest > pod.CPUUsage {
		return pod.CPURequest - pod.CPUUsage
	}
	return 0
}

// GenerateCPURecommendations 生成CPU优化建议
func (helper *PodAnalysisHelper) GenerateCPURecommendations(pod *PodResourceInfo) []string {
	var recommendations []string

	if pod.CPUReqPct < 20 {
		recommendations = append(recommendations, "建议降低CPU请求量，当前使用率偏低")
	}
	if pod.CPUReqPct > 95 {
		recommendations = append(recommendations, "建议增加CPU请求量，避免性能瓶颈")
	}
	if pod.CPULimit > 0 && pod.CPULimitPct > 90 {
		recommendations = append(recommendations, "建议增加CPU限制量，当前接近上限")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "当前CPU配置较为合理")
	}

	return recommendations
}

// GetActiveAlerts 获取活跃告警
func (helper *PodAnalysisHelper) GetActiveAlerts(clusterName, namespace, podName string) []string {
	alerts := []string{}

	// 模拟获取告警信息，实际环境中应该从告警系统获取
	if helper.collector.activityService != nil {
		// 这里可以调用activity service获取实际告警
		alerts = append(alerts, "高内存使用率告警", "CPU使用率异常告警")
	}

	return alerts
}

// GetHistoryAlerts 获取历史告警
func (helper *PodAnalysisHelper) GetHistoryAlerts(clusterName, namespace, podName string) []string {
	return []string{"内存泄漏告警(已解决)", "重启次数过多告警(已解决)"}
}

// CalculateSeverityLevel 计算告警严重程度
func (helper *PodAnalysisHelper) CalculateSeverityLevel(pod *PodResourceInfo) string {
	if pod.MemoryReqPct > 95 || pod.CPUReqPct > 95 {
		return "严重"
	} else if pod.MemoryReqPct > 80 || pod.CPUReqPct > 80 {
		return "警告"
	}
	return "正常"
}

// GenerateCPUTrendData 生成CPU趋势数据
func (helper *PodAnalysisHelper) GenerateCPUTrendData(pod *PodResourceInfo, startTime, endTime time.Time, hours int) []struct {
	Timestamp time.Time `json:"timestamp"`
	Usage     float64   `json:"usage"`
	Request   int64     `json:"request"`
	Limit     int64     `json:"limit"`
} {
	var dataPoints []struct {
		Timestamp time.Time `json:"timestamp"`
		Usage     float64   `json:"usage"`
		Request   int64     `json:"request"`
		Limit     int64     `json:"limit"`
	}

	// 生成每小时的数据点
	interval := time.Hour
	if hours <= 6 {
		interval = 30 * time.Minute // 短时间范围用更密集的数据点
	}

	for t := startTime; t.Before(endTime); t = t.Add(interval) {
		// 模拟趋势数据，实际环境中应该从监控系统获取
		baseUsage := pod.CPUReqPct
		variance := 20.0 // 20%的波动
		usage := baseUsage + (variance * (0.5 - time.Now().Sub(t).Hours()/24.0))

		if usage < 0 {
			usage = 5
		}
		if usage > 100 {
			usage = 95
		}

		dataPoints = append(dataPoints, struct {
			Timestamp time.Time `json:"timestamp"`
			Usage     float64   `json:"usage"`
			Request   int64     `json:"request"`
			Limit     int64     `json:"limit"`
		}{
			Timestamp: t,
			Usage:     usage,
			Request:   pod.CPURequest,
			Limit:     pod.CPULimit,
		})
	}

	return dataPoints
}

// GenerateMemoryTrendData 生成内存趋势数据
func (helper *PodAnalysisHelper) GenerateMemoryTrendData(pod *PodResourceInfo, startTime, endTime time.Time, hours int) []struct {
	Timestamp time.Time `json:"timestamp"`
	Usage     float64   `json:"usage"`
	Request   int64     `json:"request"`
	Limit     int64     `json:"limit"`
} {
	var dataPoints []struct {
		Timestamp time.Time `json:"timestamp"`
		Usage     float64   `json:"usage"`
		Request   int64     `json:"request"`
		Limit     int64     `json:"limit"`
	}

	// 生成每小时的数据点
	interval := time.Hour
	if hours <= 6 {
		interval = 30 * time.Minute // 短时间范围用更密集的数据点
	}

	for t := startTime; t.Before(endTime); t = t.Add(interval) {
		// 模拟趋势数据，实际环境中应该从监控系统获取
		baseUsage := pod.MemoryReqPct
		variance := 15.0 // 15%的波动
		usage := baseUsage + (variance * (0.5 - time.Now().Sub(t).Hours()/24.0))

		if usage < 0 {
			usage = 10
		}
		if usage > 100 {
			usage = 98
		}

		dataPoints = append(dataPoints, struct {
			Timestamp time.Time `json:"timestamp"`
			Usage     float64   `json:"usage"`
			Request   int64     `json:"request"`
			Limit     int64     `json:"limit"`
		}{
			Timestamp: t,
			Usage:     usage,
			Request:   pod.MemoryRequest,
			Limit:     pod.MemoryLimit,
		})
	}

	return dataPoints
}

// CalculateCPUStatistics 计算CPU统计信息
func (helper *PodAnalysisHelper) CalculateCPUStatistics(dataPoints []struct {
	Timestamp time.Time `json:"timestamp"`
	Usage     float64   `json:"usage"`
	Request   int64     `json:"request"`
	Limit     int64     `json:"limit"`
}) struct {
	Average  float64 `json:"average"`
	Peak     float64 `json:"peak"`
	Minimum  float64 `json:"minimum"`
	Variance float64 `json:"variance"`
} {
	if len(dataPoints) == 0 {
		return struct {
			Average  float64 `json:"average"`
			Peak     float64 `json:"peak"`
			Minimum  float64 `json:"minimum"`
			Variance float64 `json:"variance"`
		}{}
	}

	var sum, peak, minimum float64
	peak = dataPoints[0].Usage
	minimum = dataPoints[0].Usage

	for _, point := range dataPoints {
		sum += point.Usage
		if point.Usage > peak {
			peak = point.Usage
		}
		if point.Usage < minimum {
			minimum = point.Usage
		}
	}

	average := sum / float64(len(dataPoints))

	// 计算方差
	var varianceSum float64
	for _, point := range dataPoints {
		diff := point.Usage - average
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(dataPoints))

	return struct {
		Average  float64 `json:"average"`
		Peak     float64 `json:"peak"`
		Minimum  float64 `json:"minimum"`
		Variance float64 `json:"variance"`
	}{
		Average:  average,
		Peak:     peak,
		Minimum:  minimum,
		Variance: variance,
	}
}

// CalculateMemoryStatistics 计算内存统计信息
func (helper *PodAnalysisHelper) CalculateMemoryStatistics(dataPoints []struct {
	Timestamp time.Time `json:"timestamp"`
	Usage     float64   `json:"usage"`
	Request   int64     `json:"request"`
	Limit     int64     `json:"limit"`
}) struct {
	Average  float64 `json:"average"`
	Peak     float64 `json:"peak"`
	Minimum  float64 `json:"minimum"`
	Variance float64 `json:"variance"`
} {
	if len(dataPoints) == 0 {
		return struct {
			Average  float64 `json:"average"`
			Peak     float64 `json:"peak"`
			Minimum  float64 `json:"minimum"`
			Variance float64 `json:"variance"`
		}{}
	}

	var sum, peak, minimum float64
	peak = dataPoints[0].Usage
	minimum = dataPoints[0].Usage

	for _, point := range dataPoints {
		sum += point.Usage
		if point.Usage > peak {
			peak = point.Usage
		}
		if point.Usage < minimum {
			minimum = point.Usage
		}
	}

	average := sum / float64(len(dataPoints))

	// 计算方差
	var varianceSum float64
	for _, point := range dataPoints {
		diff := point.Usage - average
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(dataPoints))

	return struct {
		Average  float64 `json:"average"`
		Peak     float64 `json:"peak"`
		Minimum  float64 `json:"minimum"`
		Variance float64 `json:"variance"`
	}{
		Average:  average,
		Peak:     peak,
		Minimum:  minimum,
		Variance: variance,
	}
}

// GenerateEventMarkers 生成事件标记
func (helper *PodAnalysisHelper) GenerateEventMarkers(clusterName, namespace, podName string, startTime, endTime time.Time) []struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
} {
	var events []struct {
		Timestamp   time.Time `json:"timestamp"`
		EventType   string    `json:"event_type"`
		Description string    `json:"description"`
		Severity    string    `json:"severity"`
	}

	// 模拟一些事件标记
	midTime := startTime.Add(endTime.Sub(startTime) / 2)

	events = append(events, struct {
		Timestamp   time.Time `json:"timestamp"`
		EventType   string    `json:"event_type"`
		Description string    `json:"description"`
		Severity    string    `json:"severity"`
	}{
		Timestamp:   midTime,
		EventType:   "alert",
		Description: "CPU使用率超过80%",
		Severity:    "warning",
	})

	return events
}

// GetPodDetailAnalysis 获取Pod详细分析
func (helper *PodAnalysisHelper) GetPodDetailAnalysis(ctx context.Context, clusterName, namespace, podName string) (*PodDetailAnalysis, error) {
	// 查找目标Pod
	targetPod, err := helper.FindPodByIdentifier(ctx, clusterName, namespace, podName)
	if err != nil {
		return nil, fmt.Errorf("查找Pod失败: %w", err)
	}

	// 获取同命名空间的相似Pod
	similarPods, err := helper.GetSimilarPodsInNamespace(ctx, clusterName, namespace)
	if err != nil {
		return nil, fmt.Errorf("获取相似Pod失败: %w", err)
	}

	// 计算命名空间平均值
	namespaceAvg := helper.CalculateNamespaceAverage(similarPods)

	// 计算集群平均值 - 获取所有集群Pod进行计算
	allPods, err := helper.GetAllClusterPods(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群Pod失败: %w", err)
	}
	clusterAvg := helper.CalculateClusterAverage(allPods)

	// 构建详细分析
	analysis := &PodDetailAnalysis{
		PodInfo:     *targetPod,
		ClusterInfo: fmt.Sprintf("集群: %s, 节点: %s", clusterName, targetPod.NodeName),
		NodeInfo:    targetPod.NodeName,
	}

	// 设置资源分析
	analysis.ResourceAnalysis.MemoryAnalysis.ConfigStatus = helper.EvaluateMemoryConfigStatus(targetPod)
	analysis.ResourceAnalysis.MemoryAnalysis.EfficiencyScore = helper.CalculateMemoryEfficiency(targetPod)
	analysis.ResourceAnalysis.MemoryAnalysis.WasteAmount = helper.CalculateMemoryWaste(targetPod)
	analysis.ResourceAnalysis.MemoryAnalysis.Recommendations = helper.GenerateMemoryRecommendations(targetPod)

	analysis.ResourceAnalysis.CPUAnalysis.ConfigStatus = helper.EvaluateCPUConfigStatus(targetPod)
	analysis.ResourceAnalysis.CPUAnalysis.EfficiencyScore = helper.CalculateCPUEfficiency(targetPod)
	analysis.ResourceAnalysis.CPUAnalysis.WasteAmount = helper.CalculateCPUWaste(targetPod)
	analysis.ResourceAnalysis.CPUAnalysis.Recommendations = helper.GenerateCPURecommendations(targetPod)

	// 设置对比分析
	analysis.ComparisonAnalysis.NamespaceAverage.MemoryUsagePct = namespaceAvg.MemoryUsagePct
	analysis.ComparisonAnalysis.NamespaceAverage.CPUUsagePct = namespaceAvg.CPUUsagePct
	analysis.ComparisonAnalysis.ClusterAverage.MemoryUsagePct = clusterAvg.MemoryUsagePct
	analysis.ComparisonAnalysis.ClusterAverage.CPUUsagePct = clusterAvg.CPUUsagePct
	analysis.ComparisonAnalysis.SimilarPods = similarPods[:utils.MinInt(5, len(similarPods))] // 最多返回5个相似Pod

	// 设置告警信息
	analysis.AlertsInfo.ActiveAlerts = helper.GetActiveAlerts(clusterName, namespace, podName)
	analysis.AlertsInfo.HistoryAlerts = helper.GetHistoryAlerts(clusterName, namespace, podName)
	analysis.AlertsInfo.SeverityLevel = helper.CalculateSeverityLevel(targetPod)
	analysis.AlertsInfo.AlertCount = len(helper.GetActiveAlerts(clusterName, namespace, podName))

	analysis.GeneratedAt = time.Now()

	return analysis, nil
}

// GetPodTrendData 获取Pod趋势数据
func (helper *PodAnalysisHelper) GetPodTrendData(ctx context.Context, clusterName, namespace, podName, hours string) (*PodTrendData, error) {
	// 查找目标Pod
	targetPod, err := helper.FindPodByIdentifier(ctx, clusterName, namespace, podName)
	if err != nil {
		return nil, fmt.Errorf("查找Pod失败: %w", err)
	}

	// 解析时间范围
	hoursInt := 24 // 默认24小时
	if h, err := time.ParseDuration(hours + "h"); err == nil {
		hoursInt = int(h.Hours())
	}

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hoursInt) * time.Hour)

	// 生成CPU趋势数据
	cpuDataPoints := helper.GenerateCPUTrendData(targetPod, startTime, endTime, hoursInt)
	cpuStats := helper.CalculateCPUStatistics(cpuDataPoints)

	// 生成内存趋势数据
	memoryDataPoints := helper.GenerateMemoryTrendData(targetPod, startTime, endTime, hoursInt)
	memoryStats := helper.CalculateMemoryStatistics(memoryDataPoints)

	// 生成事件标记
	eventMarkers := helper.GenerateEventMarkers(clusterName, namespace, podName, startTime, endTime)

	// 构建趋势数据
	trendData := &PodTrendData{
		PodInfo: *targetPod,
	}

	// 设置时间范围
	trendData.TimeRange.StartTime = startTime
	trendData.TimeRange.EndTime = endTime
	trendData.TimeRange.Duration = fmt.Sprintf("%d小时", hoursInt)

	// 设置CPU趋势数据
	trendData.CPUTrend.DataPoints = make([]struct {
		Timestamp time.Time `json:"timestamp"`
		Usage     float64   `json:"usage"`
		Request   int64     `json:"request"`
		Limit     int64     `json:"limit"`
	}, len(cpuDataPoints))

	for i, point := range cpuDataPoints {
		trendData.CPUTrend.DataPoints[i] = struct {
			Timestamp time.Time `json:"timestamp"`
			Usage     float64   `json:"usage"`
			Request   int64     `json:"request"`
			Limit     int64     `json:"limit"`
		}{
			Timestamp: point.Timestamp,
			Usage:     point.Usage,
			Request:   point.Request,
			Limit:     point.Limit,
		}
	}
	trendData.CPUTrend.Statistics = cpuStats

	// 设置内存趋势数据
	trendData.MemoryTrend.DataPoints = make([]struct {
		Timestamp time.Time `json:"timestamp"`
		Usage     float64   `json:"usage"`
		Request   int64     `json:"request"`
		Limit     int64     `json:"limit"`
	}, len(memoryDataPoints))

	for i, point := range memoryDataPoints {
		trendData.MemoryTrend.DataPoints[i] = struct {
			Timestamp time.Time `json:"timestamp"`
			Usage     float64   `json:"usage"`
			Request   int64     `json:"request"`
			Limit     int64     `json:"limit"`
		}{
			Timestamp: point.Timestamp,
			Usage:     point.Usage,
			Request:   point.Request,
			Limit:     point.Limit,
		}
	}
	trendData.MemoryTrend.Statistics = memoryStats

	// 设置事件标记
	trendData.EventMarkers = make([]struct {
		Timestamp   time.Time `json:"timestamp"`
		EventType   string    `json:"event_type"`
		Description string    `json:"description"`
		Severity    string    `json:"severity"`
	}, len(eventMarkers))

	for i, event := range eventMarkers {
		trendData.EventMarkers[i] = struct {
			Timestamp   time.Time `json:"timestamp"`
			EventType   string    `json:"event_type"`
			Description string    `json:"description"`
			Severity    string    `json:"severity"`
		}{
			Timestamp:   event.Timestamp,
			EventType:   event.EventType,
			Description: event.Description,
			Severity:    event.Severity,
		}
	}

	trendData.GeneratedAt = time.Now()

	return trendData, nil
}

// GetAllClusterPods 获取所有集群Pod（用于计算集群平均值）
func (helper *PodAnalysisHelper) GetAllClusterPods(ctx context.Context, clusterName string) ([]PodResourceInfo, error) {
	searchReq := PodSearchRequest{
		Cluster: clusterName,
		Page:    1,
		Size:    1000, // 获取较多数据以计算平均值
	}

	searchResult, err := helper.collector.SearchPods(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	return searchResult.Pods, nil
}
