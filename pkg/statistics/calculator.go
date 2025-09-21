package statistics

import (
	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/models"
	"math"
	"time"

	"github.com/gin-gonic/gin"
)

// ResourceEfficiencyCalculator 资源效率计算器
type ResourceEfficiencyCalculator struct{}

// NewResourceEfficiencyCalculator 创建资源效率计算器实例
func NewResourceEfficiencyCalculator() *ResourceEfficiencyCalculator {
	return &ResourceEfficiencyCalculator{}
}

// CalculateResourceEfficiency 计算综合资源效率
// 基于Pod的实际CPU和内存使用率计算整体资源效率
func (calc *ResourceEfficiencyCalculator) CalculateResourceEfficiency(analysisResult *collector.AnalysisResult) float64 {
	if analysisResult == nil || analysisResult.TotalPods == 0 {
		return 0.0
	}

	// 获取所有Pod数据（包括问题Pod和正常Pod）
	allPods := analysisResult.Top50Problems

	// 如果没有足够的数据，返回基于问题Pod比例的简单计算
	if len(allPods) == 0 {
		if analysisResult.TotalPods == 0 {
			return 0.0
		}
		return float64(analysisResult.TotalPods-analysisResult.UnreasonablePods) / float64(analysisResult.TotalPods) * 100
	}

	var totalCPUEfficiency, totalMemoryEfficiency float64
	var validPods int

	// 计算所有Pod的平均资源使用效率
	for _, pod := range allPods {
		// CPU效率：实际使用率（如果有请求的话）
		if pod.CPURequest > 0 && pod.CPUReqPct > 0 {
			// CPUReqPct已经是使用率百分比，限制在合理范围内
			cpuEfficiency := math.Min(pod.CPUReqPct, 100.0)
			totalCPUEfficiency += cpuEfficiency
		}

		// 内存效率：实际使用率（如果有请求的话）
		if pod.MemoryRequest > 0 && pod.MemoryReqPct > 0 {
			// MemoryReqPct已经是使用率百分比，限制在合理范围内
			memoryEfficiency := math.Min(pod.MemoryReqPct, 100.0)
			totalMemoryEfficiency += memoryEfficiency
		}

		validPods++
	}

	// 如果没有有效的Pod数据，使用简单算法
	if validPods == 0 {
		if analysisResult.TotalPods == 0 {
			return 0.0
		}
		return float64(analysisResult.TotalPods-analysisResult.UnreasonablePods) / float64(analysisResult.TotalPods) * 100
	}

	// 计算平均效率
	avgCPUEfficiency := totalCPUEfficiency / float64(validPods)
	avgMemoryEfficiency := totalMemoryEfficiency / float64(validPods)

	// 综合资源效率 = (CPU效率 + 内存效率) / 2
	overallEfficiency := (avgCPUEfficiency + avgMemoryEfficiency) / 2.0

	// 确保返回值在合理范围内
	return math.Max(0.0, math.Min(100.0, overallEfficiency))
}

// ClusterStatusDistributionBuilder 集群状态分布构建器
type ClusterStatusDistributionBuilder struct{}

// NewClusterStatusDistributionBuilder 创建集群状态分布构建器实例
func NewClusterStatusDistributionBuilder() *ClusterStatusDistributionBuilder {
	return &ClusterStatusDistributionBuilder{}
}

// BuildDistribution 计算集群状态分布 - 为前端图表提供数据
// 统计不同状态的集群数量并返回图表所需的数据格式
func (builder *ClusterStatusDistributionBuilder) BuildDistribution(clusters []models.ClusterConfig) []gin.H {
	// 统计各种状态的集群数量
	statusCount := make(map[string]int)

	for _, cluster := range clusters {
		statusCount[cluster.Status]++
	}

	// 构建前端期望的数据格式
	var distribution []gin.H

	// 在线集群
	if count, exists := statusCount["online"]; exists {
		distribution = append(distribution, gin.H{
			"name":  "在线",
			"value": count,
			"color": "#22c55e", // 绿色
		})
	}

	// 离线集群
	if count, exists := statusCount["offline"]; exists {
		distribution = append(distribution, gin.H{
			"name":  "离线",
			"value": count,
			"color": "#ef4444", // 红色
		})
	}

	// 错误状态集群
	if count, exists := statusCount["error"]; exists {
		distribution = append(distribution, gin.H{
			"name":  "错误",
			"value": count,
			"color": "#f59e0b", // 橙色
		})
	}

	// 未知状态集群
	if count, exists := statusCount["unknown"]; exists {
		distribution = append(distribution, gin.H{
			"name":  "未知",
			"value": count,
			"color": "#6b7280", // 灰色
		})
	}

	// 如果没有任何状态数据，返回空分布
	if len(distribution) == 0 {
		distribution = append(distribution, gin.H{
			"name":  "在线",
			"value": 0,
			"color": "#22c55e",
		})
	}

	return distribution
}

// SystemStatsBuilder 系统统计数据构建器
type SystemStatsBuilder struct {
	efficiencyCalc      *ResourceEfficiencyCalculator
	distributionBuilder *ClusterStatusDistributionBuilder
}

// NewSystemStatsBuilder 创建系统统计数据构建器实例
func NewSystemStatsBuilder() *SystemStatsBuilder {
	return &SystemStatsBuilder{
		efficiencyCalc:      NewResourceEfficiencyCalculator(),
		distributionBuilder: NewClusterStatusDistributionBuilder(),
	}
}

// BuildSystemStats 构建系统统计响应数据
// 整合集群、Pod、资源效率等统计信息
func (builder *SystemStatsBuilder) BuildSystemStats(clusters []models.ClusterConfig, analysisResult *collector.AnalysisResult) gin.H {
	// 统计在线集群数量
	onlineClusters := 0
	for _, cluster := range clusters {
		if cluster.Status == "online" {
			onlineClusters++
		}
	}

	// 计算资源效率 - 基于所有Pod的实际使用率
	resourceEfficiency := builder.efficiencyCalc.CalculateResourceEfficiency(analysisResult)

	// 计算集群状态分布
	clusterStatusDistribution := builder.distributionBuilder.BuildDistribution(clusters)

	// 构建系统统计响应
	return gin.H{
		"total_clusters":              len(clusters),
		"online_clusters":             onlineClusters,
		"total_pods":                  analysisResult.TotalPods,
		"problem_pods":                analysisResult.UnreasonablePods,
		"resource_efficiency":         resourceEfficiency,
		"cluster_status_distribution": clusterStatusDistribution,
		"last_update":                 time.Now().Format(time.RFC3339),
	}
}