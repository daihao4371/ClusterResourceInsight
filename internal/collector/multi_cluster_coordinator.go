package collector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/models"
	"cluster-resource-insight/internal/service"
)

// MultiClusterCoordinator 多集群协调器实现 - 负责协调多个集群的资源收集和数据聚合
type MultiClusterCoordinator struct {
	// 内嵌多集群资源收集器，提供完整的多集群功能
	*MultiClusterResourceCollector
}

// NewMultiClusterResourceCollector 创建多集群资源收集器
// 初始化多集群资源收集器，配置依赖服务和缓存参数
// 返回:
//   - *MultiClusterResourceCollector: 配置完成的多集群资源收集器实例
func NewMultiClusterResourceCollector() *MultiClusterResourceCollector {
	return &MultiClusterResourceCollector{
		clusterService:   service.NewClusterService(),
		historyService:   service.NewHistoryService(),
		activityService:  service.NewActivityService(),
		podCacheTTL:      2 * time.Minute, // Pod数据缓存2分钟
		analysisCacheTTL: 3 * time.Minute, // 分析结果缓存3分钟
	}
}

// CollectAllPodsData 兼容原有单集群接口的实现，改为多集群模式
// 为了保持向后兼容性，当单集群收集器的kubeClient为空时，自动切换到多集群模式
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和超时
//
// 返回:
//   - *AnalysisResult: 聚合后的分析结果
//   - error: 收集过程中的错误信息
//
// CollectAllPodsData 兼容原有单集群接口的实现，改为多集群模式
func (rc *ResourceCollector) CollectAllPodsData(ctx context.Context) (*AnalysisResult, error) {
	// 如果是空的收集器，使用多集群模式
	if rc.kubeClient == nil {
		multiCollector := NewMultiClusterResourceCollector()
		return multiCollector.CollectAllClustersData(ctx)
	}

	// 原有的单集群逻辑
	return rc.collectSingleClusterData(ctx, "default-cluster")
}

// CollectAllClustersData 收集所有集群的数据
// 不启用持久化的简化版本，适用于实时查询场景
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和超时
//
// 返回:
//   - *AnalysisResult: 所有集群的聚合分析结果
//   - error: 收集过程中的错误信息
func (mc *MultiClusterResourceCollector) CollectAllClustersData(ctx context.Context) (*AnalysisResult, error) {
	return mc.CollectAllClustersDataWithPersistence(ctx, false)
}

// CollectAllClustersDataWithPersistence 收集所有集群的数据并可选择持久化
// 核心的多集群数据收集方法，支持缓存机制、并发处理、持久化存储和告警生成
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和超时
//   - enablePersistence: 是否启用数据持久化到数据库
//
// 返回:
//   - *AnalysisResult: 包含所有集群聚合数据的分析结果
//   - error: 收集过程中的错误信息
func (mc *MultiClusterResourceCollector) CollectAllClustersDataWithPersistence(ctx context.Context, enablePersistence bool) (*AnalysisResult, error) {
	// 首先尝试从缓存获取分析结果（仅在非持久化模式下）
	if !enablePersistence {
		if cachedAnalysis, cached := mc.getCachedAnalysis(); cached {
			logger.Info("使用缓存的分析结果: %d 个问题Pod", len(cachedAnalysis.Top50Problems))
			return cachedAnalysis, nil
		}
		logger.Info("分析结果缓存未命中，开始数据收集和分析...")
	}

	// 获取所有集群配置
	clusters, err := mc.clusterService.GetAllClusters()
	if err != nil {
		return nil, fmt.Errorf("获取集群列表失败: %v", err)
	}

	if len(clusters) == 0 {
		// 如果没有配置集群，返回空结果
		return &AnalysisResult{
			TotalPods:        0,
			UnreasonablePods: 0,
			Top50Problems:    []PodResourceInfo{},
			GeneratedAt:      time.Now(),
			ClustersAnalyzed: 0,
		}, nil
	}

	var allPods []PodResourceInfo
	clustersAnalyzed := 0

	// 使用信号量限制并发集群处理数量，防止过度并发
	semaphore := make(chan struct{}, 2) // 最多2个集群并发处理
	resultChan := make(chan struct {
		ClusterName string
		Pods        []PodResourceInfo
		Success     bool
	}, len(clusters))

	// 为整个多集群数据收集设置更长的超时时间
	allClustersCtx, cancel := context.WithTimeout(ctx, 600*time.Second) // 10分钟超时
	defer cancel()

	// 遍历所有集群收集数据
	activeGoroutines := 0
	for _, cluster := range clusters {
		if cluster.Status != "online" {
			logger.Info("跳过离线集群: %s (状态: %s)", cluster.ClusterName, cluster.Status)
			continue
		}

		activeGoroutines++
		go func(c models.ClusterConfig) {
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 为单个集群设置超时时间
			clusterCtx, clusterCancel := context.WithTimeout(allClustersCtx, 300*time.Second) // 5分钟超时
			defer clusterCancel()

			logger.Info("开始收集集群 %s 的数据...", c.ClusterName)

			// 为每个集群创建客户端
			kubeClient, metricsClient, err := mc.clusterService.CreateKubernetesClient(&c)
			if err != nil {
				logger.Error("创建集群 %s 的客户端失败: %v", c.ClusterName, err)
				// 记录集群连接失败活动
				if mc.activityService != nil {
					mc.activityService.RecordClusterConnection(c.ID, c.ClusterName, false, fmt.Sprintf("客户端创建失败: %v", err))
				}
				resultChan <- struct {
					ClusterName string
					Pods        []PodResourceInfo
					Success     bool
				}{ClusterName: c.ClusterName, Success: false}
				return
			}

			// 记录集群连接成功活动
			if mc.activityService != nil {
				mc.activityService.RecordClusterConnection(c.ID, c.ClusterName, true, "集群客户端创建成功")
			}

			// 创建单集群收集器
			singleCollector := &ResourceCollector{
				kubeClient:    kubeClient,
				metricsClient: metricsClient,
			}

			// 收集该集群的数据
			clusterResult, err := singleCollector.collectSingleClusterData(clusterCtx, c.ClusterName)
			if err != nil {
				logger.Error("收集集群 %s 数据失败: %v", c.ClusterName, err)
				// 记录数据收集失败活动
				if mc.activityService != nil {
					mc.activityService.RecordDataCollection(c.ID, c.ClusterName, 0, false)
				}
				resultChan <- struct {
					ClusterName string
					Pods        []PodResourceInfo
					Success     bool
				}{ClusterName: c.ClusterName, Success: false}
				return
			}

			// 记录数据收集成功活动
			if mc.activityService != nil {
				mc.activityService.RecordDataCollection(c.ID, c.ClusterName, len(clusterResult.Top50Problems), true)
			}

			// 为每个 Pod 添加集群名称标识
			for i := range clusterResult.Top50Problems {
				clusterResult.Top50Problems[i].ClusterName = c.ClusterName
			}

			// 如果启用持久化，保存到数据库
			if enablePersistence && mc.historyService != nil {
				// 收集该集群所有Pod数据（不仅仅是问题Pod）
				allClusterPods, err := singleCollector.collectAllPodsWithoutFiltering(clusterCtx, c.ClusterName)
				if err == nil {
					// 转换为service.PodResourceInfo格式
					servicePods := ConvertToServicePods(allClusterPods)
					// 保存历史数据
					if saveErr := mc.historyService.SavePodMetrics(c.ID, servicePods); saveErr != nil {
						logger.Error("保存集群 %s 历史数据失败: %v", c.ClusterName, saveErr)
					} else {
						logger.Info("成功保存集群 %s 的 %d 条Pod监控数据", c.ClusterName, len(allClusterPods))
					}
				}
			}

			logger.Info("集群 %s 数据收集完成，共收集 %d 个问题Pod", c.ClusterName, len(clusterResult.Top50Problems))

			// 生成资源使用率告警
			if mc.activityService != nil {
				mc.generateResourceAlerts(c.ID, c.ClusterName, clusterResult.Top50Problems)
			}

			resultChan <- struct {
				ClusterName string
				Pods        []PodResourceInfo
				Success     bool
			}{ClusterName: c.ClusterName, Pods: clusterResult.Top50Problems, Success: true}
		}(cluster)
	}

	// 收集所有结果，容忍部分失败
	for i := 0; i < activeGoroutines; i++ {
		select {
		case result := <-resultChan:
			if result.Success {
				allPods = append(allPods, result.Pods...)
				clustersAnalyzed++
				logger.Info("集群 %s 数据收集成功", result.ClusterName)
			} else {
				logger.Error("集群 %s 数据收集失败", result.ClusterName)
			}
		case <-allClustersCtx.Done():
			logger.Info("多集群数据收集超时，已处理 %d/%d 个集群", clustersAnalyzed, len(clusters))
			goto analysis
		}
	}

analysis:

	// 重新分析合并后的数据
	analysisResult := mc.analyzeMultiClusterData(allPods)
	analysisResult.ClustersAnalyzed = clustersAnalyzed

	logger.Info("多集群数据收集完成，成功处理 %d/%d 个集群，共收集 %d 个问题Pod",
		clustersAnalyzed, len(clusters), len(allPods))

	// 缓存分析结果（除非是强制持久化模式）
	if !enablePersistence {
		mc.setCachedAnalysis(analysisResult)
	}

	return analysisResult, nil
}

// GetNamespacesSummary 获取所有命名空间汇总信息
func (mc *MultiClusterResourceCollector) GetNamespacesSummary(ctx context.Context) ([]NamespaceSummary, error) {
	clusters, err := mc.clusterService.GetAllClusters()
	if err != nil {
		return nil, fmt.Errorf("获取集群列表失败: %v", err)
	}

	var summaries []NamespaceSummary

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

		namespaceSummaries, err := singleCollector.getNamespacesSummary(ctx, cluster.ClusterName)
		if err != nil {
			continue
		}

		summaries = append(summaries, namespaceSummaries...)
	}

	return summaries, nil
}

// GetNamespacePods 获取指定命名空间下的所有Pod
func (mc *MultiClusterResourceCollector) GetNamespacePods(ctx context.Context, namespace string) ([]PodResourceInfo, error) {
	clusters, err := mc.clusterService.GetAllClusters()
	if err != nil {
		return nil, fmt.Errorf("获取集群列表失败: %v", err)
	}

	var allPods []PodResourceInfo

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

		pods, err := singleCollector.collectNamespacePodsData(ctx, namespace, cluster.ClusterName)
		if err != nil {
			continue
		}

		allPods = append(allPods, pods...)
	}

	return allPods, nil
}

// SearchPods 搜索Pod
func (mc *MultiClusterResourceCollector) SearchPods(ctx context.Context, req PodSearchRequest) (*PodSearchResponse, error) {
	logger.Info("开始搜索Pod，筛选条件: 集群=%s, 命名空间=%s, 状态=%s, 查询=%s, 页码=%d, 每页=%d",
		req.Cluster, req.Namespace, req.Status, req.Query, req.Page, req.Size)

	// 首先尝试从缓存获取数据
	allPods, cached := mc.getCachedPods()

	if !cached {
		logger.Info("Pod缓存未命中，开始收集数据...")

		// 缓存未命中，收集所有Pod数据
		clusters, err := mc.clusterService.GetAllClusters()
		if err != nil {
			return nil, fmt.Errorf("获取集群列表失败: %v", err)
		}

		for _, cluster := range clusters {
			if cluster.Status != "online" {
				logger.Info("跳过离线集群: %s (状态: %s)", cluster.ClusterName, cluster.Status)
				continue
			}

			// 如果指定了集群筛选，跳过不匹配的集群以提高性能
			if req.Cluster != "" && cluster.ClusterName != req.Cluster {
				logger.Info("跳过不匹配的集群: %s (筛选条件: %s)", cluster.ClusterName, req.Cluster)
				continue
			}

			kubeClient, metricsClient, err := mc.clusterService.CreateKubernetesClient(&cluster)
			if err != nil {
				logger.Error("创建集群 %s 客户端失败，跳过: %v", cluster.ClusterName, err)
				continue
			}

			singleCollector := &ResourceCollector{
				kubeClient:    kubeClient,
				metricsClient: metricsClient,
			}

			clusterPods, err := singleCollector.collectAllPodsWithoutFiltering(ctx, cluster.ClusterName)
			if err != nil {
				logger.Error("收集集群 %s Pod数据失败，跳过: %v", cluster.ClusterName, err)
				continue
			}

			logger.Info("成功收集集群 %s 的 %d 个Pod", cluster.ClusterName, len(clusterPods))
			allPods = append(allPods, clusterPods...)
		}

		// 更新缓存
		mc.setCachedPods(allPods)
		logger.Info("Pod数据收集完成，共 %d 条记录已缓存", len(allPods))
	} else {
		logger.Info("使用缓存的Pod数据，共 %d 条记录", len(allPods))
	}

	// 应用筛选条件
	logger.Info("应用筛选条件前Pod数量: %d", len(allPods))
	filteredPods := mc.filterPods(allPods, req)
	logger.Info("应用筛选条件后Pod数量: %d", len(filteredPods))

	// 计算分页
	total := len(filteredPods)
	totalPages := (total + req.Size - 1) / req.Size

	start := (req.Page - 1) * req.Size
	end := start + req.Size

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var resultPods []PodResourceInfo
	if start < end {
		resultPods = filteredPods[start:end]
	}

	logger.Info("返回分页结果: 总数=%d, 当前页=%d, 每页=%d, 总页数=%d, 返回数量=%d",
		total, req.Page, req.Size, totalPages, len(resultPods))

	return &PodSearchResponse{
		Pods:       resultPods,
		Total:      total,
		Page:       req.Page,
		Size:       req.Size,
		TotalPages: totalPages,
	}, nil
}

// filterPods 根据搜索条件筛选Pod
func (mc *MultiClusterResourceCollector) filterPods(pods []PodResourceInfo, req PodSearchRequest) []PodResourceInfo {
	var filtered []PodResourceInfo

	for _, pod := range pods {
		// 应用集群筛选 - 这是核心修复
		if req.Cluster != "" && pod.ClusterName != req.Cluster {
			continue
		}

		// 应用命名空间筛选
		if req.Namespace != "" && pod.Namespace != req.Namespace {
			continue
		}

		// 应用状态筛选
		if req.Status != "" && pod.Status != req.Status {
			continue
		}

		// 应用搜索关键词筛选
		if req.Query != "" {
			if !strings.Contains(strings.ToLower(pod.PodName), strings.ToLower(req.Query)) {
				continue
			}
		}

		filtered = append(filtered, pod)
	}

	return filtered
}
