package models

import (
	"time"
	"gorm.io/gorm"
)

// ClusterConfig 集群配置表模型
type ClusterConfig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ClusterName     string    `gorm:"uniqueIndex;size:100;not null" json:"cluster_name"`         // 集群名称，唯一索引
	ClusterAlias    string    `gorm:"size:100" json:"cluster_alias"`                             // 集群别名
	APIServer       string    `gorm:"size:255;not null" json:"api_server"`                       // API Server 地址
	AuthType        string    `gorm:"size:20;not null;default:'kubeconfig'" json:"auth_type"`    // 认证类型：token/cert/kubeconfig
	AuthConfig      string    `gorm:"type:text" json:"auth_config"`                              // 认证配置（加密存储的JSON）
	Status          string    `gorm:"size:20;default:'unknown'" json:"status"`                   // 集群状态：online/offline/unknown
	Tags            string    `gorm:"type:json" json:"tags"`                                     // 集群标签（JSON格式）
	CollectInterval int       `gorm:"default:30" json:"collect_interval"`                        // 采集间隔（分钟）
	LastCollectAt   *time.Time `json:"last_collect_at"`                                          // 最后采集时间
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`                                       // 软删除
}

// PodMetricsHistory Pod 监控历史表模型
type PodMetricsHistory struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ClusterID     uint      `gorm:"index;not null" json:"cluster_id"`                    // 集群ID，建立索引
	Namespace     string    `gorm:"size:100;not null;index" json:"namespace"`           // 命名空间，建立索引
	PodName       string    `gorm:"size:255;not null;index" json:"pod_name"`            // Pod名称，建立索引
	NodeName      string    `gorm:"size:100" json:"node_name"`                          // 节点名称
	
	// 内存相关字段（单位：字节）
	MemoryUsage   int64     `json:"memory_usage"`                                       // 内存实际使用量
	MemoryRequest int64     `json:"memory_request"`                                     // 内存请求量
	MemoryLimit   int64     `json:"memory_limit"`                                       // 内存限制量
	MemoryReqPct  float64   `json:"memory_req_pct"`                                     // 内存请求利用率
	MemoryLimitPct float64  `json:"memory_limit_pct"`                                   // 内存限制利用率
	
	// CPU相关字段（单位：millicores）
	CPUUsage      int64     `json:"cpu_usage"`                                          // CPU实际使用量
	CPURequest    int64     `json:"cpu_request"`                                        // CPU请求量
	CPULimit      int64     `json:"cpu_limit"`                                          // CPU限制量
	CPUReqPct     float64   `json:"cpu_req_pct"`                                        // CPU请求利用率
	CPULimitPct   float64   `json:"cpu_limit_pct"`                                      // CPU限制利用率
	
	// 状态和问题描述
	Status        string    `gorm:"size:20;default:'reasonable'" json:"status"`         // 状态：reasonable/unreasonable
	Issues        string    `gorm:"type:json" json:"issues"`                            // 问题描述（JSON数组）
	
	CollectedAt   time.Time `gorm:"index" json:"collected_at"`                          // 采集时间，建立索引
	CreatedAt     time.Time `json:"created_at"`
	
	// 外键关联
	Cluster       ClusterConfig `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// SystemSettings 系统配置表模型
type SystemSettings struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"uniqueIndex;size:100;not null" json:"key"`       // 配置键，唯一索引
	Value     string    `gorm:"type:text" json:"value"`                         // 配置值
	ValueType string    `gorm:"size:20;default:'string'" json:"value_type"`     // 值类型：string/int/bool/json
	Description string  `gorm:"size:255" json:"description"`                    // 配置描述
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AlertRule 告警规则配置表模型
type AlertRule struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:100;not null" json:"name"`                     // 规则名称
	Type            string    `gorm:"size:50;not null" json:"type"`                      // 规则类型：resource_usage/config_missing/cluster_error
	Conditions      string    `gorm:"type:json" json:"conditions"`                       // 告警条件（JSON格式）
	Severity        string    `gorm:"size:20;default:'warning'" json:"severity"`         // 严重程度：info/warning/error/critical
	Enabled         bool      `gorm:"default:true" json:"enabled"`                       // 是否启用
	NotifyChannels  string    `gorm:"type:json" json:"notify_channels"`                  // 通知渠道（JSON数组）
	Description     string    `gorm:"size:500" json:"description"`                       // 规则描述
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// AlertHistory 告警历史记录表模型
type AlertHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	RuleID      *uint     `gorm:"index" json:"rule_id"`                          // 告警规则ID，可为空
	ClusterID   uint      `gorm:"index" json:"cluster_id"`                       // 集群ID
	AlertLevel  string    `gorm:"size:20" json:"alert_level"`                    // 告警级别
	Title       string    `gorm:"size:255" json:"title"`                         // 告警标题
	Message     string    `gorm:"type:text" json:"message"`                      // 告警消息
	Status      string    `gorm:"size:20;default:'active'" json:"status"`        // 状态：active/resolved/suppressed
	TriggeredAt time.Time `gorm:"index" json:"triggered_at"`                     // 触发时间
	ResolvedAt  *time.Time `json:"resolved_at"`                                  // 解决时间
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// 外键关联
	Rule        *AlertRule    `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	Cluster     ClusterConfig `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// SystemActivity 系统活动记录表模型
type SystemActivity struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"size:20;not null;index" json:"type"`            // 活动类型：success/warning/error/info
	ClusterID   uint      `gorm:"index" json:"cluster_id"`                       // 集群ID，可选
	Title       string    `gorm:"size:255;not null" json:"title"`                // 活动标题
	Message     string    `gorm:"type:text" json:"message"`                      // 活动详情
	Source      string    `gorm:"size:50" json:"source"`                         // 来源：collector/scheduler/api/system
	Details     string    `gorm:"type:json" json:"details"`                      // 详细信息（JSON格式）
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
	
	// 外键关联
	Cluster     ClusterConfig `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// TableName 指定表名
func (ClusterConfig) TableName() string {
	return "cluster_configs"
}

func (PodMetricsHistory) TableName() string {
	return "pod_metrics_history"
}

func (SystemSettings) TableName() string {
	return "system_settings"
}

func (AlertRule) TableName() string {
	return "alert_rules"
}

func (AlertHistory) TableName() string {
	return "alert_history"
}

func (SystemActivity) TableName() string {
	return "system_activities"
}