package utils

import (
	"fmt"
	"strings"
)

// 辅助函数：格式化字节为人类可读格式
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatMillicores 格式化millicores为人类可读格式
func FormatMillicores(millicores int64) string {
	if millicores >= 1000 {
		return fmt.Sprintf("%.2f", float64(millicores)/1000.0)
	}
	return fmt.Sprintf("%dm", millicores)
}

// Contains 检查字符串切片是否包含特定字符串
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// CalculatePaginationInfo 计算分页信息
func CalculatePaginationInfo(total, page, size int) (start, end, totalPages int) {
	totalPages = (total + size - 1) / size
	start = (page - 1) * size
	end = start + size

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return start, end, totalPages
}

// ValidateResourceThresholds 验证资源阈值配置的合理性
func ValidateResourceThresholds(memThreshold, cpuThreshold, ratioThreshold float64) error {
	if memThreshold < 0 || memThreshold > 100 {
		return fmt.Errorf("内存阈值必须在0-100之间，当前值: %.2f", memThreshold)
	}
	if cpuThreshold < 0 || cpuThreshold > 100 {
		return fmt.Errorf("CPU阈值必须在0-100之间，当前值: %.2f", cpuThreshold)
	}
	if ratioThreshold < 1 {
		return fmt.Errorf("比例阈值必须大于等于1，当前值: %.2f", ratioThreshold)
	}
	return nil
}

// SanitizeClusterName 清理集群名称，移除多余空格
func SanitizeClusterName(clusterName string) string {
	return strings.TrimSpace(clusterName)
}

// BuildProblemSummary 构建问题摘要信息
func BuildProblemSummary(issues [][]string) map[string]int {
	summary := make(map[string]int)

	for _, podIssues := range issues {
		for _, issue := range podIssues {
			summary[issue]++
		}
	}

	return summary
}
