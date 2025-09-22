package collector

import (
	"time"

	"cluster-resource-insight/internal/logger"
)

// getCachedPods 获取缓存的Pod数据
func (mc *MultiClusterResourceCollector) getCachedPods() ([]PodResourceInfo, bool) {
	mc.podsCacheMux.RLock()
	defer mc.podsCacheMux.RUnlock()

	if time.Now().After(mc.podsCacheExp) || mc.podsCache == nil {
		return nil, false // 缓存已过期或未初始化
	}

	// 返回副本避免外部修改
	result := make([]PodResourceInfo, len(mc.podsCache))
	copy(result, mc.podsCache)
	return result, true
}

// setCachedPods 设置Pod数据缓存
func (mc *MultiClusterResourceCollector) setCachedPods(pods []PodResourceInfo) {
	mc.podsCacheMux.Lock()
	defer mc.podsCacheMux.Unlock()

	mc.podsCache = make([]PodResourceInfo, len(pods))
	copy(mc.podsCache, pods)
	mc.podsCacheExp = time.Now().Add(mc.podCacheTTL)

	logger.Info("Pod数据缓存已更新，共 %d 条记录，过期时间: %v", len(pods), mc.podsCacheExp)
}

// getCachedAnalysis 获取缓存的分析结果
func (mc *MultiClusterResourceCollector) getCachedAnalysis() (*AnalysisResult, bool) {
	mc.analysisCacheMux.RLock()
	defer mc.analysisCacheMux.RUnlock()

	if time.Now().After(mc.analysisCacheExp) || mc.analysisCache == nil {
		return nil, false // 缓存已过期或未初始化
	}

	// 返回副本避免外部修改
	result := &AnalysisResult{
		TotalPods:        mc.analysisCache.TotalPods,
		UnreasonablePods: mc.analysisCache.UnreasonablePods,
		GeneratedAt:      mc.analysisCache.GeneratedAt,
		ClustersAnalyzed: mc.analysisCache.ClustersAnalyzed,
	}

	result.Top50Problems = make([]PodResourceInfo, len(mc.analysisCache.Top50Problems))
	copy(result.Top50Problems, mc.analysisCache.Top50Problems)

	return result, true
}

// setCachedAnalysis 设置分析结果缓存
func (mc *MultiClusterResourceCollector) setCachedAnalysis(analysis *AnalysisResult) {
	mc.analysisCacheMux.Lock()
	defer mc.analysisCacheMux.Unlock()

	if analysis != nil {
		mc.analysisCache = &AnalysisResult{
			TotalPods:        analysis.TotalPods,
			UnreasonablePods: analysis.UnreasonablePods,
			GeneratedAt:      analysis.GeneratedAt,
			ClustersAnalyzed: analysis.ClustersAnalyzed,
		}

		mc.analysisCache.Top50Problems = make([]PodResourceInfo, len(analysis.Top50Problems))
		copy(mc.analysisCache.Top50Problems, analysis.Top50Problems)

		mc.analysisCacheExp = time.Now().Add(mc.analysisCacheTTL)

		logger.Info("分析结果缓存已更新，问题Pod数量: %d，过期时间: %v",
			len(analysis.Top50Problems), mc.analysisCacheExp)
	}
}

// invalidateCache 失效所有缓存
func (mc *MultiClusterResourceCollector) invalidateCache() {
	mc.podsCacheMux.Lock()
	mc.analysisCacheMux.Lock()
	defer mc.podsCacheMux.Unlock()
	defer mc.analysisCacheMux.Unlock()

	mc.podsCacheExp = time.Time{}
	mc.analysisCacheExp = time.Time{}

	logger.Info("所有缓存已失效")
}
