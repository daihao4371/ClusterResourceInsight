package collector

import (
	"cluster-resource-insight/pkg/utils"
	"fmt"
)

// generateResourceAlerts 为问题Pod生成资源使用率告警
func (mc *MultiClusterResourceCollector) generateResourceAlerts(clusterID uint, clusterName string, problemPods []PodResourceInfo) {
	if mc.activityService == nil {
		return
	}

	criticalCount := 0
	warningCount := 0

	for _, pod := range problemPods {
		// 分析问题严重程度并生成相应告警
		isCritical := false
		alertMessage := ""

		// 检查是否为严重问题（利用率极低或配置缺失）
		if utils.Contains(pod.Issues, "缺少内存请求配置") || utils.Contains(pod.Issues, "缺少CPU请求配置") {
			isCritical = true
			alertMessage = fmt.Sprintf("Pod %s/%s 缺少资源配置", pod.Namespace, pod.PodName)
			criticalCount++
		} else if (pod.MemoryReqPct > 0 && pod.MemoryReqPct < 10) || (pod.CPUReqPct > 0 && pod.CPUReqPct < 5) {
			isCritical = true
			alertMessage = fmt.Sprintf("Pod %s/%s 资源利用率极低：内存 %.1f%%, CPU %.1f%%", pod.Namespace, pod.PodName, pod.MemoryReqPct, pod.CPUReqPct)
			criticalCount++
		} else {
			alertMessage = fmt.Sprintf("Pod %s/%s 资源配置不合理", pod.Namespace, pod.PodName)
			warningCount++
		}

		// 记录资源告警活动
		alertType := "warning"
		if isCritical {
			alertType = "critical"
		}
		mc.activityService.RecordResourceAlert(clusterID, clusterName, pod.PodName, alertType, alertMessage)

		// 为严重问题创建系统告警记录
		if isCritical {
			level := "error"
			if criticalCount <= 5 { // 只为前5个最严重的问题创建告警记录，避免告警过多
				title := "严重资源配置问题"
				mc.activityService.CreateAlert(clusterID, level, title, alertMessage, "active")
			}
		}
	}

	// 创建集群级别的汇总告警
	if criticalCount > 0 || warningCount > 0 {
		title := "集群资源配置问题汇总"
		message := fmt.Sprintf("发现 %d 个严重问题，%d 个一般问题", criticalCount, warningCount)
		level := "warning"
		if criticalCount > 5 {
			level = "error"
		}
		mc.activityService.CreateAlert(clusterID, level, title, message, "active")
	}
}
