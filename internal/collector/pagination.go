package collector

import (
	"math"
	"sort"
)

// PodSorter Pod排序器
type PodSorter struct{}

// NewPodSorter 创建Pod排序器实例
func NewPodSorter() *PodSorter {
	return &PodSorter{}
}

// SortProblems 对问题Pod进行排序
// 支持按CPU浪费、内存浪费、总浪费程度排序
func (sorter *PodSorter) SortProblems(problems []PodResourceInfo, sortBy string) {
	switch sortBy {
	case "cpu_waste":
		sort.Slice(problems, func(i, j int) bool {
			cpuWasteI := sorter.calculateCPUWaste(problems[i])
			cpuWasteJ := sorter.calculateCPUWaste(problems[j])
			return cpuWasteI > cpuWasteJ
		})
	case "memory_waste":
		sort.Slice(problems, func(i, j int) bool {
			memoryWasteI := sorter.calculateMemoryWaste(problems[i])
			memoryWasteJ := sorter.calculateMemoryWaste(problems[j])
			return memoryWasteI > memoryWasteJ
		})
	case "total_waste":
		fallthrough
	default:
		sort.Slice(problems, func(i, j int) bool {
			totalWasteI := sorter.calculateTotalWaste(problems[i])
			totalWasteJ := sorter.calculateTotalWaste(problems[j])
			return totalWasteI > totalWasteJ
		})
	}
}

// calculateCPUWaste 计算CPU浪费程度
func (sorter *PodSorter) calculateCPUWaste(pod PodResourceInfo) float64 {
	if pod.CPUReqPct <= 0 {
		return 0
	}
	return math.Max(0, 100-pod.CPUReqPct)
}

// calculateMemoryWaste 计算内存浪费程度
func (sorter *PodSorter) calculateMemoryWaste(pod PodResourceInfo) float64 {
	if pod.MemoryReqPct <= 0 {
		return 0
	}
	return math.Max(0, 100-pod.MemoryReqPct)
}

// calculateTotalWaste 计算总浪费程度
func (sorter *PodSorter) calculateTotalWaste(pod PodResourceInfo) float64 {
	cpuWaste := sorter.calculateCPUWaste(pod)
	memoryWaste := sorter.calculateMemoryWaste(pod)
	return (cpuWaste + memoryWaste) / 2
}

// ApplyPagination 应用分页逻辑到Pod数据
func ApplyPagination(data []PodResourceInfo, page, size int) ([]PodResourceInfo, int, int, int) {
	total := len(data)
	totalPages := (total + size - 1) / size
	start := (page - 1) * size
	end := start + size

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var pagedData []PodResourceInfo
	if start < end {
		pagedData = data[start:end]
	}

	return pagedData, total, totalPages, end
}