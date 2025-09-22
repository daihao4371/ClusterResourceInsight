package collector

import (
	"cluster-resource-insight/internal/logger"
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// collectSingleClusterData 收集单个集群的数据
func (rc *ResourceCollector) collectSingleClusterData(ctx context.Context, clusterName string) (*AnalysisResult, error) {
	// 为整个集群数据收集设置更长的超时时间
	clusterCtx, cancel := context.WithTimeout(ctx, 300*time.Second) // 5分钟超时
	defer cancel()

	// 获取所有 namespace
	namespaces, err := rc.kubeClient.CoreV1().Namespaces().List(clusterCtx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var allPods []PodResourceInfo

	// 使用信号量限制并发命名空间处理数量，防止过度并发
	semaphore := make(chan struct{}, 3) // 最多3个命名空间并发处理
	resultChan := make(chan []PodResourceInfo, len(namespaces.Items))
	errorChan := make(chan error, len(namespaces.Items))

	for _, namespace := range namespaces.Items {
		go func(ns string) {
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 为单个命名空间设置更短的超时时间
			namespaceCtx, nsCancel := context.WithTimeout(clusterCtx, 60*time.Second) // 1分钟超时
			defer nsCancel()

			podInfos, err := rc.collectNamespacePodsData(namespaceCtx, ns, clusterName)
			if err != nil {
				errorChan <- fmt.Errorf("收集命名空间 %s 数据失败: %v", ns, err)
				return
			}
			resultChan <- podInfos
		}(namespace.Name)
	}

	// 收集所有结果，容忍部分失败
	successCount := 0
	for i := 0; i < len(namespaces.Items); i++ {
		select {
		case podInfos := <-resultChan:
			allPods = append(allPods, podInfos...)
			successCount++
		case err := <-errorChan:
			// 记录错误但继续处理其他 namespace
			logger.Error("错误: %v", err)
		case <-clusterCtx.Done():
			logger.Info("集群 %s 数据收集超时，已收集 %d/%d 个命名空间", clusterName, successCount, len(namespaces.Items))
			break
		}
	}

	logger.Info("集群 %s 数据收集完成，成功处理 %d/%d 个命名空间，共收集 %d 个Pod",
		clusterName, successCount, len(namespaces.Items), len(allPods))

	// 分析数据并找出问题
	analysisResult := rc.analyzeResourceUsage(allPods)

	return analysisResult, nil
}

func (rc *ResourceCollector) collectNamespacePodsData(ctx context.Context, namespace, clusterName string) ([]PodResourceInfo, error) {
	// 为单个命名空间设置重试机制
	maxRetries := 2
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避重试
			waitTime := time.Duration(attempt) * 2 * time.Second
			logger.Info("命名空间 %s 第%d次重试，等待 %v", namespace, attempt, waitTime)
			time.Sleep(waitTime)
		}

		// 为单次请求设置超时
		requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// 获取 Pod 列表
		pods, err := rc.kubeClient.CoreV1().Pods(namespace).List(requestCtx, metav1.ListOptions{})
		if err != nil {
			lastErr = err
			logger.Error("获取命名空间 %s Pod列表失败 (第%d次尝试): %v", namespace, attempt+1, err)
			continue
		}

		// 获取 Pod Metrics (如果 metrics server 不可用则跳过)
		var podMetrics *metricsv1beta1.PodMetricsList
		var metricsMap map[string]*metricsv1beta1.PodMetrics

		// 为Metrics请求设置更短的超时，因为Metrics服务通常更容易超时
		metricsCtx, metricsCancel := context.WithTimeout(ctx, 15*time.Second)
		defer metricsCancel()

		podMetrics, err = rc.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(metricsCtx, metav1.ListOptions{})
		if err != nil {
			// Metrics Server 不可用，只分析配置不使用实际用量
			logger.Info("警告: 无法获取命名空间 %s 的 metrics 数据 (可能 metrics-server 未安装): %v", namespace, err)
			metricsMap = make(map[string]*metricsv1beta1.PodMetrics)
		} else {
			// 创建 metrics 映射表以便快速查找
			metricsMap = make(map[string]*metricsv1beta1.PodMetrics)
			for i := range podMetrics.Items {
				metricsMap[podMetrics.Items[i].Name] = &podMetrics.Items[i]
			}
		}

		var podInfos []PodResourceInfo

		for _, pod := range pods.Items {
			// 跳过已完成的 Pod
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
				continue
			}

			// 检查 Pod 是否在运行状态
			if pod.Status.Phase != corev1.PodRunning {
				continue
			}

			podInfo := rc.extractPodResourceInfo(&pod, metricsMap[pod.Name], strings.TrimSpace(clusterName))

			// 如果 Pod 有 metrics 数据但使用量为 0，记录警告
			if metricsMap[pod.Name] != nil && podInfo.MemoryUsage == 0 && podInfo.CPUUsage == 0 {
				logger.Info("警告: Pod %s/%s 有 metrics 数据但使用量为0，可能数据收集有问题", pod.Namespace, pod.Name)
			}

			podInfos = append(podInfos, podInfo)
		}

		return podInfos, nil
	}

	return nil, fmt.Errorf("收集命名空间 %s 数据失败，尝试%d次均失败: %v", namespace, maxRetries+1, lastErr)
}

// collectAllPodsWithoutFiltering 收集所有Pod而不进行问题筛选
func (rc *ResourceCollector) collectAllPodsWithoutFiltering(ctx context.Context, clusterName string) ([]PodResourceInfo, error) {
	namespaces, err := rc.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var allPods []PodResourceInfo

	for _, namespace := range namespaces.Items {
		podInfos, err := rc.collectNamespacePodsData(ctx, namespace.Name, clusterName)
		if err != nil {
			fmt.Printf("Error collecting data from namespace %s in cluster %s: %v\n", namespace.Name, clusterName, err)
			continue
		}
		allPods = append(allPods, podInfos...)
	}

	return allPods, nil
}

// getNamespacesSummary 获取单个集群的命名空间汇总信息
func (rc *ResourceCollector) getNamespacesSummary(ctx context.Context, clusterName string) ([]NamespaceSummary, error) {
	namespaces, err := rc.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var summaries []NamespaceSummary

	for _, namespace := range namespaces.Items {
		pods, err := rc.collectNamespacePodsData(ctx, namespace.Name, clusterName)
		if err != nil {
			continue
		}

		summary := NamespaceSummary{
			NamespaceName: namespace.Name,
			ClusterName:   clusterName,
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

		summaries = append(summaries, summary)
	}

	return summaries, nil
}
