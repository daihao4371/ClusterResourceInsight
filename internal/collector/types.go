package collector

import (
	"sync"
	"time"

	"cluster-resource-insight/internal/service"

	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ResourceCollector 单集群资源收集器 - 负责从单个Kubernetes集群收集Pod资源信息
type ResourceCollector struct {
	kubeClient    kubernetes.Interface    // Kubernetes API客户端，用于访问集群基础资源
	metricsClient metricsclientset.Interface // Metrics API客户端，用于获取资源使用量数据
}

// PodResourceInfo Pod资源信息 - 包含Pod的完整资源配置和使用情况
type PodResourceInfo struct {
	// 基础信息
	PodName     string `json:"pod_name"`     // Pod名称
	Namespace   string `json:"namespace"`    // 所属命名空间
	NodeName    string `json:"node_name"`    // 运行节点名称
	ClusterName string `json:"cluster_name"` // 所属集群名称
	
	// 内存资源信息
	MemoryUsage    int64   `json:"memory_usage"`     // 实际内存使用量 (bytes)
	MemoryRequest  int64   `json:"memory_request"`   // 内存请求量 (bytes)
	MemoryLimit    int64   `json:"memory_limit"`     // 内存限制量 (bytes)
	MemoryReqPct   float64 `json:"memory_req_pct"`   // 内存使用量/请求量百分比
	MemoryLimitPct float64 `json:"memory_limit_pct"` // 内存使用量/限制量百分比
	
	// CPU资源信息
	CPUUsage    int64   `json:"cpu_usage"`     // 实际CPU使用量 (millicores)
	CPURequest  int64   `json:"cpu_request"`   // CPU请求量 (millicores)
	CPULimit    int64   `json:"cpu_limit"`     // CPU限制量 (millicores)
	CPUReqPct   float64 `json:"cpu_req_pct"`   // CPU使用量/请求量百分比
	CPULimitPct float64 `json:"cpu_limit_pct"` // CPU使用量/限制量百分比
	
	// 状态和问题信息
	Status       string    `json:"status"`        // 资源配置状态：合理/不合理
	Issues       []string  `json:"issues"`        // 发现的具体问题列表
	CreationTime time.Time `json:"creation_time"` // Pod创建时间
}

// AnalysisResult 资源分析结果 - 包含整体分析统计和问题Pod列表
type AnalysisResult struct {
	TotalPods        int               `json:"total_pods"`        // 分析的Pod总数
	UnreasonablePods int               `json:"unreasonable_pods"` // 存在问题的Pod数量
	Top50Problems    []PodResourceInfo `json:"top50_problems"`    // 最严重的50个问题Pod
	GeneratedAt      time.Time         `json:"generated_at"`      // 分析结果生成时间
	ClustersAnalyzed int               `json:"clusters_analyzed"` // 参与分析的集群数量
}

// NamespaceSummary 命名空间汇总信息 - 单个命名空间的资源使用统计
type NamespaceSummary struct {
	NamespaceName      string `json:"namespace_name"`       // 命名空间名称
	ClusterName        string `json:"cluster_name"`         // 所属集群名称
	TotalPods          int    `json:"total_pods"`           // 命名空间内Pod总数
	UnreasonablePods   int    `json:"unreasonable_pods"`    // 存在问题的Pod数量
	TotalMemoryUsage   int64  `json:"total_memory_usage"`   // 命名空间总内存使用量
	TotalCPUUsage      int64  `json:"total_cpu_usage"`      // 命名空间总CPU使用量
	TotalMemoryRequest int64  `json:"total_memory_request"` // 命名空间总内存请求量
	TotalCPURequest    int64  `json:"total_cpu_request"`    // 命名空间总CPU请求量
}

// NamespaceTreeData 命名空间树状数据结构 - 用于层次化展示命名空间内的Pod信息
type NamespaceTreeData struct {
	NamespaceName string              `json:"namespace_name"` // 命名空间名称
	ClusterName   string              `json:"cluster_name"`   // 所属集群名称
	Children      []PodResourceInfo   `json:"children"`       // 命名空间内的Pod列表
	Summary       NamespaceSummary    `json:"summary"`        // 命名空间汇总统计
}

// PodSearchRequest Pod搜索请求参数 - 定义Pod搜索和筛选的条件
type PodSearchRequest struct {
	Query     string `form:"query"`     // 搜索关键词，用于匹配Pod名称
	Namespace string `form:"namespace"` // 命名空间筛选条件
	Cluster   string `form:"cluster"`   // 集群筛选条件
	Status    string `form:"status"`    // 状态筛选条件（合理/不合理）
	Page      int    `form:"page"`      // 分页页码，从1开始
	Size      int    `form:"size"`      // 每页数据条数
}

// PodSearchResponse Pod搜索响应结果 - 包含搜索结果和分页信息
type PodSearchResponse struct {
	Pods       []PodResourceInfo `json:"pods"`        // 当前页的Pod列表
	Total      int               `json:"total"`       // 符合条件的Pod总数
	Page       int               `json:"page"`        // 当前页码
	Size       int               `json:"size"`        // 每页大小
	TotalPages int               `json:"total_pages"` // 总页数
}

// MultiClusterResourceCollector 多集群资源收集器 - 统一管理多个Kubernetes集群的资源收集
type MultiClusterResourceCollector struct {
	// 依赖的服务组件
	clusterService  *service.ClusterService  // 集群配置管理服务
	historyService  *service.HistoryService  // 历史数据持久化服务  
	activityService *service.ActivityService // 活动记录和告警服务
	
	// Pod数据缓存机制
	podsCache    []PodResourceInfo // Pod数据缓存存储
	podsCacheMux sync.RWMutex      // Pod缓存读写锁
	podsCacheExp time.Time         // Pod缓存过期时间
	
	// 分析结果缓存机制
	analysisCache    *AnalysisResult // 分析结果缓存存储
	analysisCacheMux sync.RWMutex    // 分析结果缓存读写锁
	analysisCacheExp time.Time       // 分析结果缓存过期时间
	
	// 缓存配置参数
	podCacheTTL      time.Duration // Pod数据缓存生存时间
	analysisCacheTTL time.Duration // 分析结果缓存生存时间
}

// PodDetailAnalysis Pod详细分析结果 - 包含单个Pod的完整资源分析报告
type PodDetailAnalysis struct {
	// 基础信息
	PodInfo      PodResourceInfo `json:"pod_info"`       // Pod基础资源信息
	ClusterInfo  string          `json:"cluster_info"`   // 集群环境信息
	NodeInfo     string          `json:"node_info"`      // 节点信息
	
	// 资源配置分析
	ResourceAnalysis struct {
		MemoryAnalysis struct {
			ConfigStatus    string  `json:"config_status"`     // 内存配置状态评估
			EfficiencyScore float64 `json:"efficiency_score"`  // 内存使用效率评分
			WasteAmount     int64   `json:"waste_amount"`      // 浪费的内存资源量
			Recommendations []string `json:"recommendations"`  // 内存优化建议
		} `json:"memory_analysis"`
		
		CPUAnalysis struct {
			ConfigStatus    string  `json:"config_status"`     // CPU配置状态评估
			EfficiencyScore float64 `json:"efficiency_score"`  // CPU使用效率评分
			WasteAmount     int64   `json:"waste_amount"`      // 浪费的CPU资源量
			Recommendations []string `json:"recommendations"`  // CPU优化建议
		} `json:"cpu_analysis"`
	} `json:"resource_analysis"`
	
	// 对比分析
	ComparisonAnalysis struct {
		NamespaceAverage struct {
			MemoryUsagePct float64 `json:"memory_usage_pct"` // 命名空间平均内存使用率
			CPUUsagePct    float64 `json:"cpu_usage_pct"`    // 命名空间平均CPU使用率
		} `json:"namespace_average"`
		
		ClusterAverage struct {
			MemoryUsagePct float64 `json:"memory_usage_pct"` // 集群平均内存使用率
			CPUUsagePct    float64 `json:"cpu_usage_pct"`    // 集群平均CPU使用率
		} `json:"cluster_average"`
		
		SimilarPods []PodResourceInfo `json:"similar_pods"` // 同命名空间相似Pod列表
	} `json:"comparison_analysis"`
	
	// 告警信息
	AlertsInfo struct {
		ActiveAlerts    []string `json:"active_alerts"`     // 当前活跃告警
		HistoryAlerts   []string `json:"history_alerts"`    // 历史告警记录
		SeverityLevel   string   `json:"severity_level"`    // 告警严重程度
		AlertCount      int      `json:"alert_count"`       // 告警总数
	} `json:"alerts_info"`
	
	GeneratedAt time.Time `json:"generated_at"` // 分析报告生成时间
}

// PodTrendData Pod历史趋势数据 - 包含Pod的资源使用历史趋势信息
type PodTrendData struct {
	// 基础信息
	PodInfo     PodResourceInfo `json:"pod_info"`     // Pod基础信息
	TimeRange   struct {
		StartTime time.Time `json:"start_time"` // 趋势数据开始时间
		EndTime   time.Time `json:"end_time"`   // 趋势数据结束时间
		Duration  string    `json:"duration"`   // 时间跨度描述
	} `json:"time_range"`
	
	// CPU趋势数据
	CPUTrend struct {
		DataPoints []struct {
			Timestamp time.Time `json:"timestamp"` // 数据采集时间点
			Usage     float64   `json:"usage"`     // CPU使用率百分比
			Request   int64     `json:"request"`   // CPU请求量 (millicores)
			Limit     int64     `json:"limit"`     // CPU限制量 (millicores)
		} `json:"data_points"`
		
		Statistics struct {
			Average    float64 `json:"average"`     // 平均使用率
			Peak       float64 `json:"peak"`        // 峰值使用率
			Minimum    float64 `json:"minimum"`     // 最低使用率
			Variance   float64 `json:"variance"`    // 使用率方差
		} `json:"statistics"`
	} `json:"cpu_trend"`
	
	// 内存趋势数据
	MemoryTrend struct {
		DataPoints []struct {
			Timestamp time.Time `json:"timestamp"` // 数据采集时间点
			Usage     float64   `json:"usage"`     // 内存使用率百分比
			Request   int64     `json:"request"`   // 内存请求量 (bytes)
			Limit     int64     `json:"limit"`     // 内存限制量 (bytes)
		} `json:"data_points"`
		
		Statistics struct {
			Average    float64 `json:"average"`     // 平均使用率
			Peak       float64 `json:"peak"`        // 峰值使用率
			Minimum    float64 `json:"minimum"`     // 最低使用率
			Variance   float64 `json:"variance"`    // 使用率方差
		} `json:"statistics"`
	} `json:"memory_trend"`
	
	// 异常事件标记
	EventMarkers []struct {
		Timestamp   time.Time `json:"timestamp"`   // 事件发生时间
		EventType   string    `json:"event_type"`  // 事件类型 (alert/restart/config_change)
		Description string    `json:"description"` // 事件描述
		Severity    string    `json:"severity"`    // 严重程度
	} `json:"event_markers"`
	
	GeneratedAt time.Time `json:"generated_at"` // 趋势数据生成时间
}

// PodOptimizationReport Pod优化建议报告 - 包含基于历史数据的资源优化建议
type PodOptimizationReport struct {
	// 基础信息
	PodInfo        PodResourceInfo `json:"pod_info"`         // Pod基础信息
	AnalysisPeriod struct {
		StartTime time.Time `json:"start_time"` // 分析期间开始时间
		EndTime   time.Time `json:"end_time"`   // 分析期间结束时间
		Duration  string    `json:"duration"`   // 分析时长描述
	} `json:"analysis_period"`
	
	// 资源配置建议
	ResourceRecommendations struct {
		Memory struct {
			CurrentRequest     int64   `json:"current_request"`      // 当前内存请求量
			RecommendedRequest int64   `json:"recommended_request"`  // 建议内存请求量
			CurrentLimit       int64   `json:"current_limit"`        // 当前内存限制量
			RecommendedLimit   int64   `json:"recommended_limit"`    // 建议内存限制量
			PotentialSavings   int64   `json:"potential_savings"`    // 潜在节省的内存量
			Confidence         float64 `json:"confidence"`           // 建议可信度评分
			Reasoning          string  `json:"reasoning"`            // 建议依据说明
		} `json:"memory"`
		
		CPU struct {
			CurrentRequest     int64   `json:"current_request"`      // 当前CPU请求量
			RecommendedRequest int64   `json:"recommended_request"`  // 建议CPU请求量
			CurrentLimit       int64   `json:"current_limit"`        // 当前CPU限制量
			RecommendedLimit   int64   `json:"recommended_limit"`    // 建议CPU限制量
			PotentialSavings   int64   `json:"potential_savings"`    // 潜在节省的CPU量
			Confidence         float64 `json:"confidence"`           // 建议可信度评分
			Reasoning          string  `json:"reasoning"`            // 建议依据说明
		} `json:"cpu"`
	} `json:"resource_recommendations"`
	
	// 成本优化分析
	CostOptimization struct {
		CurrentMonthlyCost    float64 `json:"current_monthly_cost"`     // 当前月度成本估算
		OptimizedMonthlyCost  float64 `json:"optimized_monthly_cost"`   // 优化后月度成本估算
		PotentialSavings      float64 `json:"potential_savings"`        // 潜在月度节省成本
		SavingsPercentage     float64 `json:"savings_percentage"`       // 节省百分比
		ROIEstimate           string  `json:"roi_estimate"`             // 投资回报率估算
	} `json:"cost_optimization"`
	
	// 性能优化建议
	PerformanceOptimization struct {
		BottleneckAnalysis []string `json:"bottleneck_analysis"`  // 性能瓶颈分析
		ScalingRecommendations []string `json:"scaling_recommendations"` // 扩缩容建议
		PerformanceRisks   []string `json:"performance_risks"`    // 性能风险提示
		OptimizationTips   []string `json:"optimization_tips"`    // 性能优化技巧
	} `json:"performance_optimization"`
	
	// 实施建议
	ImplementationGuide struct {
		Priority        string   `json:"priority"`         // 实施优先级 (high/medium/low)
		Steps          []string  `json:"steps"`            // 实施步骤
		RiskAssessment string    `json:"risk_assessment"`  // 风险评估
		TestingAdvice  []string  `json:"testing_advice"`   // 测试建议
		RollbackPlan   string    `json:"rollback_plan"`    // 回滚方案
	} `json:"implementation_guide"`
	
	GeneratedAt time.Time `json:"generated_at"` // 报告生成时间
}

// NewResourceCollector 创建单集群资源收集器实例
// 参数:
//   - kubeClient: Kubernetes API客户端，用于访问集群基础资源
//   - metricsClient: Metrics API客户端，用于获取资源使用量数据
// 返回:
//   - *ResourceCollector: 创建的资源收集器实例
//   - error: 创建过程中的错误信息
func NewResourceCollector(kubeClient kubernetes.Interface, metricsClient metricsclientset.Interface) (*ResourceCollector, error) {
	return &ResourceCollector{
		kubeClient:    kubeClient,
		metricsClient: metricsClient,
	}, nil
}