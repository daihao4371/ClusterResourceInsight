package cache

import (
	"sync"
	"time"

	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/models"
)

// CacheManager 缓存管理器
type CacheManager struct {
	// 集群配置缓存
	clusters    []models.ClusterConfig
	clustersMux sync.RWMutex
	clustersExp time.Time

	// Pod数据缓存
	pods    []collector.PodResourceInfo
	podsMux sync.RWMutex
	podsExp time.Time

	// 分析结果缓存
	analysis    *collector.AnalysisResult
	analysisMux sync.RWMutex
	analysisExp time.Time

	// 缓存配置
	clusterCacheTTL  time.Duration // 集群配置缓存TTL
	podCacheTTL      time.Duration // Pod数据缓存TTL
	analysisCacheTTL time.Duration // 分析结果缓存TTL
}

// NewCacheManager 创建缓存管理器
func NewCacheManager() *CacheManager {
	return &CacheManager{
		clusterCacheTTL:  5 * time.Minute, // 集群配置缓存5分钟
		podCacheTTL:      2 * time.Minute, // Pod数据缓存2分钟
		analysisCacheTTL: 3 * time.Minute, // 分析结果缓存3分钟
	}
}

// SetClusters 设置集群配置缓存
func (cm *CacheManager) SetClusters(clusters []models.ClusterConfig) {
	cm.clustersMux.Lock()
	defer cm.clustersMux.Unlock()

	cm.clusters = make([]models.ClusterConfig, len(clusters))
	copy(cm.clusters, clusters)
	cm.clustersExp = time.Now().Add(cm.clusterCacheTTL)
}

// GetClusters 获取集群配置缓存
func (cm *CacheManager) GetClusters() ([]models.ClusterConfig, bool) {
	cm.clustersMux.RLock()
	defer cm.clustersMux.RUnlock()

	if time.Now().After(cm.clustersExp) {
		return nil, false // 缓存已过期
	}

	if cm.clusters == nil {
		return nil, false // 缓存未初始化
	}

	// 返回副本以避免外部修改
	result := make([]models.ClusterConfig, len(cm.clusters))
	copy(result, cm.clusters)
	return result, true
}

// SetPods 设置Pod数据缓存
func (cm *CacheManager) SetPods(pods []collector.PodResourceInfo) {
	cm.podsMux.Lock()
	defer cm.podsMux.Unlock()

	cm.pods = make([]collector.PodResourceInfo, len(pods))
	copy(cm.pods, pods)
	cm.podsExp = time.Now().Add(cm.podCacheTTL)
}

// GetPods 获取Pod数据缓存
func (cm *CacheManager) GetPods() ([]collector.PodResourceInfo, bool) {
	cm.podsMux.RLock()
	defer cm.podsMux.RUnlock()

	if time.Now().After(cm.podsExp) {
		return nil, false // 缓存已过期
	}

	if cm.pods == nil {
		return nil, false // 缓存未初始化
	}

	// 返回副本以避免外部修改
	result := make([]collector.PodResourceInfo, len(cm.pods))
	copy(result, cm.pods)
	return result, true
}

// SetAnalysis 设置分析结果缓存
func (cm *CacheManager) SetAnalysis(analysis *collector.AnalysisResult) {
	cm.analysisMux.Lock()
	defer cm.analysisMux.Unlock()

	if analysis != nil {
		// 深拷贝分析结果
		cm.analysis = &collector.AnalysisResult{
			TotalPods:        analysis.TotalPods,
			UnreasonablePods: analysis.UnreasonablePods,
			GeneratedAt:      analysis.GeneratedAt,
			ClustersAnalyzed: analysis.ClustersAnalyzed,
		}

		// 拷贝问题Pod列表
		cm.analysis.Top50Problems = make([]collector.PodResourceInfo, len(analysis.Top50Problems))
		copy(cm.analysis.Top50Problems, analysis.Top50Problems)

		cm.analysisExp = time.Now().Add(cm.analysisCacheTTL)
	}
}

// GetAnalysis 获取分析结果缓存
func (cm *CacheManager) GetAnalysis() (*collector.AnalysisResult, bool) {
	cm.analysisMux.RLock()
	defer cm.analysisMux.RUnlock()

	if time.Now().After(cm.analysisExp) {
		return nil, false // 缓存已过期
	}

	if cm.analysis == nil {
		return nil, false // 缓存未初始化
	}

	// 返回副本以避免外部修改
	result := &collector.AnalysisResult{
		TotalPods:        cm.analysis.TotalPods,
		UnreasonablePods: cm.analysis.UnreasonablePods,
		GeneratedAt:      cm.analysis.GeneratedAt,
		ClustersAnalyzed: cm.analysis.ClustersAnalyzed,
	}

	result.Top50Problems = make([]collector.PodResourceInfo, len(cm.analysis.Top50Problems))
	copy(result.Top50Problems, cm.analysis.Top50Problems)

	return result, true
}

// InvalidateAll 失效所有缓存
func (cm *CacheManager) InvalidateAll() {
	cm.clustersMux.Lock()
	cm.podsMux.Lock()
	cm.analysisMux.Lock()
	defer cm.clustersMux.Unlock()
	defer cm.podsMux.Unlock()
	defer cm.analysisMux.Unlock()

	cm.clustersExp = time.Time{}
	cm.podsExp = time.Time{}
	cm.analysisExp = time.Time{}
}

// InvalidatePods 失效Pod相关缓存
func (cm *CacheManager) InvalidatePods() {
	cm.podsMux.Lock()
	cm.analysisMux.Lock()
	defer cm.podsMux.Unlock()
	defer cm.analysisMux.Unlock()

	cm.podsExp = time.Time{}
	cm.analysisExp = time.Time{}
}

// GetCacheStatus 获取缓存状态信息
func (cm *CacheManager) GetCacheStatus() map[string]interface{} {
	cm.clustersMux.RLock()
	cm.podsMux.RLock()
	cm.analysisMux.RLock()
	defer cm.clustersMux.RUnlock()
	defer cm.podsMux.RUnlock()
	defer cm.analysisMux.RUnlock()

	now := time.Now()

	return map[string]interface{}{
		"clusters": map[string]interface{}{
			"cached":      cm.clusters != nil,
			"count":       len(cm.clusters),
			"expires":     cm.clustersExp,
			"expired":     now.After(cm.clustersExp),
			"ttl_seconds": int(cm.clustersExp.Sub(now).Seconds()),
		},
		"pods": map[string]interface{}{
			"cached":      cm.pods != nil,
			"count":       len(cm.pods),
			"expires":     cm.podsExp,
			"expired":     now.After(cm.podsExp),
			"ttl_seconds": int(cm.podsExp.Sub(now).Seconds()),
		},
		"analysis": map[string]interface{}{
			"cached":      cm.analysis != nil,
			"expires":     cm.analysisExp,
			"expired":     now.After(cm.analysisExp),
			"ttl_seconds": int(cm.analysisExp.Sub(now).Seconds()),
		},
	}
}
