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