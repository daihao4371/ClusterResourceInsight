package deduplication

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"cluster-resource-insight/internal/models"
)

// AlertDeduplicator 告警降噪管理器
type AlertDeduplicator struct {
	// 内存中的告警指纹缓存，用于快速去重
	fingerprintCache map[string]*AlertCacheEntry
	cacheMutex       sync.RWMutex
	
	// 配置参数
	suppressionDuration time.Duration // 相同告警的抑制时间
	maxCacheSize       int           // 最大缓存大小
}

// AlertCacheEntry 告警缓存条目
type AlertCacheEntry struct {
	FirstOccurred time.Time // 首次发生时间
	LastOccurred  time.Time // 最后发生时间
	Count         int       // 发生次数
	AlertID       uint      // 数据库中的告警ID
}

// NewAlertDeduplicator 创建新的告警降噪管理器
func NewAlertDeduplicator() *AlertDeduplicator {
	return &AlertDeduplicator{
		fingerprintCache:    make(map[string]*AlertCacheEntry),
		suppressionDuration: 30 * time.Minute, // 默认30分钟抑制时间
		maxCacheSize:       1000,              // 默认最大1000条缓存
	}
}

// GenerateFingerprint 生成告警指纹
// 基于集群ID、告警级别、标题生成唯一标识
func (d *AlertDeduplicator) GenerateFingerprint(clusterID uint, level, title string) string {
	content := fmt.Sprintf("%d:%s:%s", clusterID, level, title)
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// ShouldCreateAlert 判断是否应该创建新告警
// 返回值：(shouldCreate bool, existingAlertID uint, count int)
func (d *AlertDeduplicator) ShouldCreateAlert(clusterID uint, level, title string) (bool, uint, int) {
	fingerprint := d.GenerateFingerprint(clusterID, level, title)
	
	d.cacheMutex.Lock()
	defer d.cacheMutex.Unlock()
	
	now := time.Now()
	entry, exists := d.fingerprintCache[fingerprint]
	
	if !exists {
		// 新告警，创建缓存条目
		d.fingerprintCache[fingerprint] = &AlertCacheEntry{
			FirstOccurred: now,
			LastOccurred:  now,
			Count:         1,
		}
		
		// 清理过期缓存
		d.cleanExpiredCache(now)
		return true, 0, 1
	}
	
	// 检查是否在抑制期内
	if now.Sub(entry.LastOccurred) < d.suppressionDuration {
		// 在抑制期内，更新计数但不创建新告警
		entry.Count++
		entry.LastOccurred = now
		return false, entry.AlertID, entry.Count
	}
	
	// 抑制期已过，可以创建新告警
	entry.Count++
	entry.LastOccurred = now
	return true, entry.AlertID, entry.Count
}

// UpdateAlertID 更新缓存中的告警ID
func (d *AlertDeduplicator) UpdateAlertID(clusterID uint, level, title string, alertID uint) {
	fingerprint := d.GenerateFingerprint(clusterID, level, title)
	
	d.cacheMutex.Lock()
	defer d.cacheMutex.Unlock()
	
	if entry, exists := d.fingerprintCache[fingerprint]; exists {
		entry.AlertID = alertID
	}
}

// cleanExpiredCache 清理过期的缓存条目
func (d *AlertDeduplicator) cleanExpiredCache(now time.Time) {
	if len(d.fingerprintCache) <= d.maxCacheSize {
		return
	}
	
	// 删除最早的条目，保持缓存大小
	oldestTime := now
	var oldestKey string
	
	for key, entry := range d.fingerprintCache {
		// 删除超过24小时的条目
		if now.Sub(entry.LastOccurred) > 24*time.Hour {
			delete(d.fingerprintCache, key)
			continue
		}
		
		// 记录最早的条目
		if entry.FirstOccurred.Before(oldestTime) {
			oldestTime = entry.FirstOccurred
			oldestKey = key
		}
	}
	
	// 如果仍然超过最大大小，删除最早的条目
	if len(d.fingerprintCache) > d.maxCacheSize && oldestKey != "" {
		delete(d.fingerprintCache, oldestKey)
	}
}

// GetCacheStats 获取缓存统计信息
func (d *AlertDeduplicator) GetCacheStats() map[string]interface{} {
	d.cacheMutex.RLock()
	defer d.cacheMutex.RUnlock()
	
	totalCount := 0
	for _, entry := range d.fingerprintCache {
		totalCount += entry.Count
	}
	
	return map[string]interface{}{
		"cached_fingerprints": len(d.fingerprintCache),
		"total_alert_count":   totalCount,
		"suppression_duration": d.suppressionDuration.String(),
		"max_cache_size":      d.maxCacheSize,
	}
}

// SetSuppressionDuration 设置抑制时间
func (d *AlertDeduplicator) SetSuppressionDuration(duration time.Duration) {
	d.cacheMutex.Lock()
	defer d.cacheMutex.Unlock()
	d.suppressionDuration = duration
}

// PrepareDeduplicatedAlert 准备去重后的告警数据
func (d *AlertDeduplicator) PrepareDeduplicatedAlert(clusterID uint, level, title, message string) *models.AlertHistory {
	fingerprint := d.GenerateFingerprint(clusterID, level, title)
	now := time.Now()
	
	alert := &models.AlertHistory{
		ClusterID:        clusterID,
		AlertLevel:       level,
		Title:           title,
		Message:         message,
		Status:          "active",
		AlertFingerprint: fingerprint,
		Count:           1,
		FirstOccurredAt: now,
		LastOccurredAt:  now,
		TriggeredAt:     now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	
	return alert
}