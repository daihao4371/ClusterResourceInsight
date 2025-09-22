package collector

import (
	"cluster-resource-insight/internal/service"
)

// ConvertToServicePods 将collector.PodResourceInfo转换为service.PodResourceInfo
func ConvertToServicePods(pods []PodResourceInfo) []service.PodResourceInfo {
	servicePods := make([]service.PodResourceInfo, len(pods))

	for i, pod := range pods {
		servicePods[i] = service.PodResourceInfo{
			PodName:        pod.PodName,
			Namespace:      pod.Namespace,
			NodeName:       pod.NodeName,
			ClusterName:    pod.ClusterName,
			MemoryUsage:    pod.MemoryUsage,
			MemoryRequest:  pod.MemoryRequest,
			MemoryLimit:    pod.MemoryLimit,
			MemoryReqPct:   pod.MemoryReqPct,
			MemoryLimitPct: pod.MemoryLimitPct,
			CPUUsage:       pod.CPUUsage,
			CPURequest:     pod.CPURequest,
			CPULimit:       pod.CPULimit,
			CPUReqPct:      pod.CPUReqPct,
			CPULimitPct:    pod.CPULimitPct,
			Status:         pod.Status,
			Issues:         pod.Issues,
			CreationTime:   pod.CreationTime,
		}
	}

	return servicePods
}