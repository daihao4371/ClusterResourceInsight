package collector

import (
	"context"
)

// DataCollector 数据收集器接口 - 定义数据收集的核心契约
type DataCollector interface {
	CollectNamespaceData(ctx context.Context, namespace, clusterName string) ([]PodResourceInfo, error)
	CollectAllClusterData(ctx context.Context, clusterName string) ([]PodResourceInfo, error)
}

// DataAnalyzer 数据分析器接口 - 定义数据分析的核心契约
type DataAnalyzer interface {
	AnalyzeResourceUsage(pods []PodResourceInfo) *AnalysisResult
	CalculateProblemScore(pod PodResourceInfo) float64
	SortByProblemSeverity(pods []PodResourceInfo)
}

// CacheManager 缓存管理器接口 - 定义缓存操作的核心契约
type CacheManager interface {
	GetCachedPods() ([]PodResourceInfo, bool)
	SetCachedPods(pods []PodResourceInfo)
	GetCachedAnalysis() (*AnalysisResult, bool)
	SetCachedAnalysis(analysis *AnalysisResult)
	InvalidateCache()
}

// ClusterCoordinator 集群协调器接口 - 定义多集群操作的核心契约
type ClusterCoordinator interface {
	CollectMultiClusterData(ctx context.Context) (*AnalysisResult, error)
	SearchPodsAcrossClusters(ctx context.Context, req PodSearchRequest) (*PodSearchResponse, error)
	GetTopResourcePods(ctx context.Context, resourceType string, limit int) ([]PodResourceInfo, error)
}

// AlertGenerator 告警生成器接口 - 定义告警生成的核心契约
type AlertGenerator interface {
	GenerateResourceAlerts(clusterID uint, clusterName string, problemPods []PodResourceInfo)
	CreateClusterSummaryAlert(clusterID uint, clusterName string, criticalCount, warningCount int)
}

// ResourceExtractor 资源提取器接口 - 定义Pod资源信息提取的核心契约
type ResourceExtractor interface {
	ExtractPodResourceInfo(pod interface{}, metrics interface{}, clusterName string) PodResourceInfo
	ExtractContainerResources(containers interface{}) (memReq, memLimit, cpuReq, cpuLimit int64)
	EstimateResourceUsage(actualUsage, request, limit int64) int64
}
