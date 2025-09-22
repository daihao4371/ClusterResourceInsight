package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"cluster-resource-insight/internal/crypto"
	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/models"

	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ClusterService 集群管理服务
type ClusterService struct {
	db *gorm.DB
}

// NewClusterService 创建集群服务实例
func NewClusterService() *ClusterService {
	return &ClusterService{
		db: database.GetDB(),
	}
}

// CreateClusterRequest 创建集群请求结构
type CreateClusterRequest struct {
	ClusterName     string         `json:"cluster_name" binding:"required"` // 集群名称，必填
	ClusterAlias    string         `json:"cluster_alias"`                   // 集群别名
	APIServer       string         `json:"api_server" binding:"required"`   // API Server地址，必填
	AuthType        string         `json:"auth_type" binding:"required"`    // 认证类型：token/cert/kubeconfig
	AuthConfig      AuthConfigData `json:"auth_config" binding:"required"`  // 认证配置
	Tags            []string       `json:"tags"`                            // 集群标签
	CollectInterval int            `json:"collect_interval"`                // 采集间隔（分钟）
}

// AuthConfigData 认证配置数据结构
type AuthConfigData struct {
	// Token认证
	BearerToken string `json:"bearer_token,omitempty"`

	// 证书认证
	ClientCert string `json:"client_cert,omitempty"`
	ClientKey  string `json:"client_key,omitempty"`
	CACert     string `json:"ca_cert,omitempty"`

	// Kubeconfig认证
	KubeConfig string `json:"kubeconfig,omitempty"`
}

// UpdateClusterRequest 更新集群请求结构
type UpdateClusterRequest struct {
	ClusterAlias    *string         `json:"cluster_alias"`    // 集群别名
	APIServer       *string         `json:"api_server"`       // API Server地址
	AuthType        *string         `json:"auth_type"`        // 认证类型
	AuthConfig      *AuthConfigData `json:"auth_config"`      // 认证配置
	Tags            []string        `json:"tags"`             // 集群标签
	CollectInterval *int            `json:"collect_interval"` // 采集间隔
}

// CreateCluster 创建新集群配置
func (cs *ClusterService) CreateCluster(req *CreateClusterRequest) (*models.ClusterConfig, error) {
	// 验证集群名称唯一性
	var existingCluster models.ClusterConfig
	result := cs.db.Where("cluster_name = ?", req.ClusterName).First(&existingCluster)
	if result.Error == nil {
		return nil, fmt.Errorf("集群名称 '%s' 已存在", req.ClusterName)
	} else if result.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("检查集群名称唯一性失败: %v", result.Error)
	}

	// 验证认证配置
	if err := cs.validateAuthConfig(req.AuthType, &req.AuthConfig); err != nil {
		return nil, fmt.Errorf("认证配置验证失败: %v", err)
	}

	// 加密认证配置
	authConfigJSON, err := json.Marshal(req.AuthConfig)
	if err != nil {
		return nil, fmt.Errorf("序列化认证配置失败: %v", err)
	}

	encryptedAuthConfig, err := crypto.EncryptData(string(authConfigJSON))
	if err != nil {
		return nil, fmt.Errorf("加密认证配置失败: %v", err)
	}

	// 处理标签
	tagsJSON := "[]"
	if len(req.Tags) > 0 {
		tagsBytes, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("序列化标签失败: %v", err)
		}
		tagsJSON = string(tagsBytes)
	}

	// 设置默认采集间隔
	collectInterval := req.CollectInterval
	if collectInterval <= 0 {
		collectInterval = 30 // 默认30分钟
	}

	// 创建集群配置
	cluster := &models.ClusterConfig{
		ClusterName:     req.ClusterName,
		ClusterAlias:    req.ClusterAlias,
		APIServer:       req.APIServer,
		AuthType:        req.AuthType,
		AuthConfig:      encryptedAuthConfig,
		Status:          "unknown", // 初始状态为未知，需要连接测试后更新
		Tags:            tagsJSON,
		CollectInterval: collectInterval,
	}

	// 保存到数据库
	if err := cs.db.Create(cluster).Error; err != nil {
		return nil, fmt.Errorf("保存集群配置失败: %v", err)
	}

	logger.Info("成功创建集群配置: %s (ID: %d)", cluster.ClusterName, cluster.ID)
	return cluster, nil
}

// GetClusterByID 根据ID获取集群配置
func (cs *ClusterService) GetClusterByID(clusterID uint) (*models.ClusterConfig, error) {
	var cluster models.ClusterConfig
	if err := cs.db.First(&cluster, clusterID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("集群ID %d 不存在", clusterID)
		}
		return nil, fmt.Errorf("获取集群配置失败: %v", err)
	}
	return &cluster, nil
}

// GetAllClusters 获取所有集群配置列表
func (cs *ClusterService) GetAllClusters() ([]models.ClusterConfig, error) {
	var clusters []models.ClusterConfig
	if err := cs.db.Find(&clusters).Error; err != nil {
		return nil, fmt.Errorf("获取集群列表失败: %v", err)
	}
	return clusters, nil
}

// UpdateCluster 更新集群配置
func (cs *ClusterService) UpdateCluster(clusterID uint, req *UpdateClusterRequest) (*models.ClusterConfig, error) {
	// 获取现有集群配置
	cluster, err := cs.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.ClusterAlias != nil {
		cluster.ClusterAlias = *req.ClusterAlias
	}
	if req.APIServer != nil {
		cluster.APIServer = *req.APIServer
	}
	if req.AuthType != nil && req.AuthConfig != nil {
		// 验证新的认证配置
		if err := cs.validateAuthConfig(*req.AuthType, req.AuthConfig); err != nil {
			return nil, fmt.Errorf("认证配置验证失败: %v", err)
		}

		// 加密新的认证配置
		authConfigJSON, err := json.Marshal(req.AuthConfig)
		if err != nil {
			return nil, fmt.Errorf("序列化认证配置失败: %v", err)
		}

		encryptedAuthConfig, err := crypto.EncryptData(string(authConfigJSON))
		if err != nil {
			return nil, fmt.Errorf("加密认证配置失败: %v", err)
		}

		cluster.AuthType = *req.AuthType
		cluster.AuthConfig = encryptedAuthConfig
		cluster.Status = "unknown" // 认证配置更改后重置状态
	}
	if req.Tags != nil {
		tagsBytes, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("序列化标签失败: %v", err)
		}
		cluster.Tags = string(tagsBytes)
	}
	if req.CollectInterval != nil {
		collectInterval := *req.CollectInterval
		if collectInterval <= 0 {
			collectInterval = 30
		}
		cluster.CollectInterval = collectInterval
	}

	// 保存更新
	if err := cs.db.Save(cluster).Error; err != nil {
		return nil, fmt.Errorf("更新集群配置失败: %v", err)
	}

	logger.Info("成功更新集群配置: %s (ID: %d)", cluster.ClusterName, cluster.ID)
	return cluster, nil
}

// DeleteCluster 删除集群配置（软删除）
func (cs *ClusterService) DeleteCluster(clusterID uint) error {
	cluster, err := cs.GetClusterByID(clusterID)
	if err != nil {
		return err
	}

	// 软删除集群配置
	if err := cs.db.Delete(cluster).Error; err != nil {
		return fmt.Errorf("删除集群配置失败: %v", err)
	}

	logger.Info("成功删除集群配置: %s (ID: %d)", cluster.ClusterName, cluster.ID)
	return nil
}

// validateAuthConfig 验证认证配置的完整性
func (cs *ClusterService) validateAuthConfig(authType string, authConfig *AuthConfigData) error {
	switch strings.ToLower(authType) {
	case "token":
		if authConfig.BearerToken == "" {
			return fmt.Errorf("Token认证方式需要提供bearer_token")
		}
	case "cert":
		if authConfig.ClientCert == "" || authConfig.ClientKey == "" {
			return fmt.Errorf("证书认证方式需要提供client_cert和client_key")
		}
	case "kubeconfig":
		if authConfig.KubeConfig == "" {
			return fmt.Errorf("Kubeconfig认证方式需要提供kubeconfig内容")
		}
	default:
		return fmt.Errorf("不支持的认证类型: %s", authType)
	}
	return nil
}

// GetDecryptedAuthConfig 获取解密后的认证配置
func (cs *ClusterService) GetDecryptedAuthConfig(cluster *models.ClusterConfig) (*AuthConfigData, error) {
	if cluster.AuthConfig == "" {
		return &AuthConfigData{}, nil
	}

	// 解密认证配置
	decryptedConfig, err := crypto.DecryptData(cluster.AuthConfig)
	if err != nil {
		return nil, fmt.Errorf("解密认证配置失败: %v", err)
	}

	// 反序列化JSON
	var authConfig AuthConfigData
	if err := json.Unmarshal([]byte(decryptedConfig), &authConfig); err != nil {
		return nil, fmt.Errorf("反序列化认证配置失败: %v", err)
	}

	return &authConfig, nil
}

// CreateKubernetesClient 根据集群配置创建Kubernetes客户端
func (cs *ClusterService) CreateKubernetesClient(cluster *models.ClusterConfig) (kubernetes.Interface, metricsclientset.Interface, error) {
	// 获取解密的认证配置
	authConfig, err := cs.GetDecryptedAuthConfig(cluster)
	if err != nil {
		return nil, nil, err
	}

	var config *rest.Config

	switch strings.ToLower(cluster.AuthType) {
	case "token":
		config = &rest.Config{
			Host:        cluster.APIServer,
			BearerToken: authConfig.BearerToken,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true, // 在生产环境中应该设置为false并提供CA证书
			},
			// 增加超时配置以防止长时间等待
			Timeout: 60 * time.Second,
		}
	case "cert":
		config = &rest.Config{
			Host: cluster.APIServer,
			TLSClientConfig: rest.TLSClientConfig{
				CertData: []byte(authConfig.ClientCert),
				KeyData:  []byte(authConfig.ClientKey),
				// 在内网环境中跳过证书验证，解决IP地址不匹配问题
				Insecure: true,
			},
			// 增加超时配置以防止长时间等待
			Timeout: 60 * time.Second,
		}
		// 注意：当设置Insecure为true时，不能同时设置CA证书，否则会报错
		// 在内网环境中，我们选择跳过证书验证来解决主机名不匹配问题
	case "kubeconfig":
		config, err = clientcmd.RESTConfigFromKubeConfig([]byte(authConfig.KubeConfig))
		if err != nil {
			return nil, nil, fmt.Errorf("解析kubeconfig失败: %v", err)
		}
		// 为kubeconfig方式也设置超时配置和跳过证书验证
		config.Timeout = 60 * time.Second
		// 清除可能存在的CA证书配置，然后设置跳过证书验证
		config.TLSClientConfig.CAData = nil
		config.TLSClientConfig.CAFile = ""
		config.TLSClientConfig.Insecure = true
	default:
		return nil, nil, fmt.Errorf("不支持的认证类型: %s", cluster.AuthType)
	}

	// 设置客户端限流配置，避免过度并发导致限流错误
	config.QPS = 20.0 // 每秒最多20个请求，降低请求频率
	config.Burst = 30 // 突发最多30个请求，适当降低突发量

	// 创建Kubernetes客户端
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("创建Kubernetes客户端失败: %v", err)
	}

	// 创建Metrics客户端
	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("创建Metrics客户端失败: %v", err)
	}

	return kubeClient, metricsClient, nil
}

// ClusterResourceInfo 集群资源详细信息
type ClusterResourceInfo struct {
	CPUUsage       float64 // CPU使用率百分比
	CPUUsedCores   float64 // 已使用CPU核数
	CPUTotalCores  float64 // 总CPU核数
	MemoryUsage    float64 // 内存使用率百分比
	MemoryUsedGB   float64 // 已使用内存（GB）
	MemoryTotalGB  float64 // 总内存（GB）
	HasRealUsage   bool    // 是否有真实使用率数据（来自Metrics API）
	DataSource     string  // 数据来源标识: "metrics" 或 "capacity"
}

// getClusterResourceUsage 获取集群的详细资源使用信息
func (cs *ClusterService) getClusterResourceUsage(ctx context.Context, kubeClient kubernetes.Interface, metricsClient metricsclientset.Interface) *ClusterResourceInfo {
	// 设置短超时以防止阻塞
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result := &ClusterResourceInfo{
		DataSource: "none",
	}

	// 获取节点列表 - 这是必需的基础信息
	nodes, err := kubeClient.CoreV1().Nodes().List(timeoutCtx, metav1.ListOptions{})
	if err != nil {
		logger.Error("获取节点列表失败: %v", err)
		return result
	}

	if len(nodes.Items) == 0 {
		return result
	}

	// 首先获取节点容量信息（总是可用的）
	var totalCPUCapacity, totalMemoryCapacity int64
	for _, node := range nodes.Items {
		if cpu, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
			totalCPUCapacity += cpu.MilliValue()
		}
		if memory, ok := node.Status.Capacity[corev1.ResourceMemory]; ok {
			totalMemoryCapacity += memory.Value()
		}
	}

	// 设置基础容量信息
	if totalCPUCapacity > 0 {
		result.CPUTotalCores = math.Round(float64(totalCPUCapacity)/1000.0*100) / 100
	}
	if totalMemoryCapacity > 0 {
		result.MemoryTotalGB = math.Round(float64(totalMemoryCapacity)/(1024*1024*1024)*100) / 100
	}
	result.DataSource = "capacity"

	// 尝试获取Metrics API数据
	if metricsClient != nil {
		nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().List(timeoutCtx, metav1.ListOptions{})
		if err != nil {
			logger.Info("Metrics API不可用，仅显示容量信息: %v", err)
			// 返回仅包含容量信息的结果
			return result
		}

		// 如果Metrics API可用，计算使用率
		var totalCPUUsage, totalMemoryUsage int64
		for _, nodeMetric := range nodeMetrics.Items {
			if cpu, ok := nodeMetric.Usage[corev1.ResourceCPU]; ok {
				totalCPUUsage += cpu.MilliValue()
			}
			if memory, ok := nodeMetric.Usage[corev1.ResourceMemory]; ok {
				totalMemoryUsage += memory.Value()
			}
		}

		// 计算使用量和使用率
		if totalCPUCapacity > 0 {
			result.CPUUsedCores = math.Round(float64(totalCPUUsage)/1000.0*100) / 100
			usage := (float64(totalCPUUsage) / float64(totalCPUCapacity)) * 100
			result.CPUUsage = math.Round(usage*100) / 100
		}

		if totalMemoryCapacity > 0 {
			result.MemoryUsedGB = math.Round(float64(totalMemoryUsage)/(1024*1024*1024)*100) / 100
			usage := (float64(totalMemoryUsage) / float64(totalMemoryCapacity)) * 100
			result.MemoryUsage = math.Round(usage*100) / 100
		}

		// 限制百分比在合理范围内
		if result.CPUUsage > 100 {
			result.CPUUsage = 100
		}
		if result.MemoryUsage > 100 {
			result.MemoryUsage = 100
		}

		result.HasRealUsage = true
		result.DataSource = "metrics"
		logger.Info("通过Metrics API获取资源使用率数据")
	} else {
		logger.Info("Metrics客户端不可用，仅显示容量信息")
	}

	return result
}

// ClusterTestResult 集群连接测试结果
type ClusterTestResult struct {
	Success        bool      `json:"success"`          // 测试是否成功
	Status         string    `json:"status"`           // 集群状态：online/offline/error
	Message        string    `json:"message"`          // 测试结果消息
	Version        string    `json:"version"`          // Kubernetes版本
	NodeCount      int       `json:"node_count"`       // 节点数量
	NamespaceCount int       `json:"namespace_count"`  // 命名空间数量
	PodCount       int       `json:"pod_count"`        // Pod总数
	HasMetrics     bool      `json:"has_metrics"`      // 是否支持Metrics API
	TestTime       time.Time `json:"test_time"`        // 测试时间
	ResponseTime   int64     `json:"response_time_ms"` // 响应时间（毫秒）
	// CPU资源信息
	CPUUsage       float64   `json:"cpu_usage"`        // CPU使用率百分比（保留2位小数）
	CPUUsedCores   float64   `json:"cpu_used_cores"`   // 已使用CPU核数
	CPUTotalCores  float64   `json:"cpu_total_cores"`  // 总CPU核数
	// 内存资源信息
	MemoryUsage    float64   `json:"memory_usage"`     // 内存使用率百分比（保留2位小数）
	MemoryUsedGB   float64   `json:"memory_used_gb"`   // 已使用内存（GB）
	MemoryTotalGB  float64   `json:"memory_total_gb"`  // 总内存（GB）
	// 资源数据来源信息
	HasRealUsage   bool      `json:"has_real_usage"`   // 是否有真实使用率数据
	DataSource     string    `json:"data_source"`      // 数据来源: "metrics" 或 "capacity"
}

// TestClusterConnection 测试集群连接
func (cs *ClusterService) TestClusterConnection(clusterID uint) (*ClusterTestResult, error) {
	startTime := time.Now()

	// 获取集群配置
	cluster, err := cs.GetClusterByID(clusterID)
	if err != nil {
		return &ClusterTestResult{
			Success:      false,
			Status:       "error",
			Message:      fmt.Sprintf("获取集群配置失败: %v", err),
			TestTime:     time.Now(),
			ResponseTime: time.Since(startTime).Milliseconds(),
		}, err
	}

	// 创建Kubernetes客户端
	kubeClient, metricsClient, err := cs.CreateKubernetesClient(cluster)
	if err != nil {
		result := &ClusterTestResult{
			Success:      false,
			Status:       "offline",
			Message:      fmt.Sprintf("创建客户端失败: %v", err),
			TestTime:     time.Now(),
			ResponseTime: time.Since(startTime).Milliseconds(),
		}
		// 更新集群状态为离线
		cs.updateClusterStatus(cluster, "offline")
		return result, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试API连接并获取集群信息
	result := &ClusterTestResult{
		TestTime: time.Now(),
	}

	// 1. 获取集群版本信息
	versionInfo, err := kubeClient.Discovery().ServerVersion()
	if err != nil {
		result.Success = false
		result.Status = "offline"
		result.Message = fmt.Sprintf("无法获取集群版本信息: %v", err)
		result.ResponseTime = time.Since(startTime).Milliseconds()
		cs.updateClusterStatus(cluster, "offline")
		return result, nil
	}
	result.Version = versionInfo.GitVersion

	// 2. 获取节点信息
	nodes, err := kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Success = false
		result.Status = "offline"
		result.Message = fmt.Sprintf("无法获取节点信息: %v", err)
		result.ResponseTime = time.Since(startTime).Milliseconds()
		cs.updateClusterStatus(cluster, "offline")
		return result, nil
	}
	result.NodeCount = len(nodes.Items)

	// 3. 获取命名空间信息
	namespaces, err := kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Success = false
		result.Status = "offline"
		result.Message = fmt.Sprintf("无法获取命名空间信息: %v", err)
		result.ResponseTime = time.Since(startTime).Milliseconds()
		cs.updateClusterStatus(cluster, "offline")
		return result, nil
	}
	result.NamespaceCount = len(namespaces.Items)

	// 4. 获取Pod总数
	pods, err := kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Success = false
		result.Status = "offline"
		result.Message = fmt.Sprintf("无法获取Pod信息: %v", err)
		result.ResponseTime = time.Since(startTime).Milliseconds()
		cs.updateClusterStatus(cluster, "offline")
		return result, nil
	}
	result.PodCount = len(pods.Items)

	// 5. 测试Metrics API并获取资源信息
	_, err = metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		result.HasMetrics = false
		logger.Info("集群 %s Metrics API 不可用: %v", cluster.ClusterName, err)
	} else {
		result.HasMetrics = true
	}
	
	// 6. 获取集群资源信息（无论Metrics API是否可用）
	resourceInfo := cs.getClusterResourceUsage(ctx, kubeClient, metricsClient)
	result.CPUUsage = resourceInfo.CPUUsage
	result.MemoryUsage = resourceInfo.MemoryUsage
	result.CPUUsedCores = resourceInfo.CPUUsedCores
	result.CPUTotalCores = resourceInfo.CPUTotalCores
	result.MemoryUsedGB = resourceInfo.MemoryUsedGB
	result.MemoryTotalGB = resourceInfo.MemoryTotalGB
	result.HasRealUsage = resourceInfo.HasRealUsage
	result.DataSource = resourceInfo.DataSource
	
	if resourceInfo.HasRealUsage {
		logger.Info("集群 %s 资源使用率 - CPU: %.2f%% (%.2f/%.2f cores), Memory: %.2f%% (%.2fGB/%.2fGB)", 
			cluster.ClusterName, resourceInfo.CPUUsage, resourceInfo.CPUUsedCores, resourceInfo.CPUTotalCores,
			resourceInfo.MemoryUsage, resourceInfo.MemoryUsedGB, resourceInfo.MemoryTotalGB)
	} else {
		logger.Info("集群 %s 容量信息 - CPU: %.2f cores, Memory: %.2fGB (无实时使用率数据)", 
			cluster.ClusterName, resourceInfo.CPUTotalCores, resourceInfo.MemoryTotalGB)
	}

	// 测试成功
	result.Success = true
	result.Status = "online"
	result.Message = fmt.Sprintf("集群连接正常，版本: %s，节点数: %d，命名空间数: %d，Pod数: %d",
		result.Version, result.NodeCount, result.NamespaceCount, result.PodCount)
	result.ResponseTime = time.Since(startTime).Milliseconds()

	// 更新集群状态为在线
	cs.updateClusterStatus(cluster, "online")

	logger.Info("集群连接测试成功: %s (%dms)", cluster.ClusterName, result.ResponseTime)
	return result, nil
}

// TestClusterConnectionByConfig 根据配置直接测试集群连接（用于创建前验证）
func (cs *ClusterService) TestClusterConnectionByConfig(req *CreateClusterRequest) (*ClusterTestResult, error) {
	startTime := time.Now()

	// 创建临时集群配置用于测试
	tempCluster := &models.ClusterConfig{
		ClusterName: req.ClusterName,
		APIServer:   req.APIServer,
		AuthType:    req.AuthType,
	}

	// 直接使用未加密的认证配置进行测试
	authConfigJSON, err := json.Marshal(req.AuthConfig)
	if err != nil {
		return &ClusterTestResult{
			Success:      false,
			Status:       "error",
			Message:      fmt.Sprintf("序列化认证配置失败: %v", err),
			TestTime:     time.Now(),
			ResponseTime: time.Since(startTime).Milliseconds(),
		}, err
	}

	// 临时加密配置用于测试
	encryptedConfig, err := crypto.EncryptData(string(authConfigJSON))
	if err != nil {
		return &ClusterTestResult{
			Success:      false,
			Status:       "error",
			Message:      fmt.Sprintf("加密认证配置失败: %v", err),
			TestTime:     time.Now(),
			ResponseTime: time.Since(startTime).Milliseconds(),
		}, err
	}
	tempCluster.AuthConfig = encryptedConfig

	// 创建Kubernetes客户端
	kubeClient, metricsClient, err := cs.CreateKubernetesClient(tempCluster)
	if err != nil {
		return &ClusterTestResult{
			Success:      false,
			Status:       "offline",
			Message:      fmt.Sprintf("创建客户端失败: %v", err),
			TestTime:     time.Now(),
			ResponseTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 执行基本连接测试
	result := &ClusterTestResult{
		TestTime: time.Now(),
	}

	// 获取集群版本信息（最基本的连接测试）
	versionInfo, err := kubeClient.Discovery().ServerVersion()
	if err != nil {
		result.Success = false
		result.Status = "offline"
		result.Message = fmt.Sprintf("无法连接到集群: %v", err)
		result.ResponseTime = time.Since(startTime).Milliseconds()
		return result, nil
	}

	result.Version = versionInfo.GitVersion

	// 快速获取基本信息
	nodes, _ := kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if nodes != nil {
		result.NodeCount = len(nodes.Items)
	}

	namespaces, _ := kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if namespaces != nil {
		result.NamespaceCount = len(namespaces.Items)
	}

	// 测试Metrics API并获取资源信息
	_, err = metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		result.HasMetrics = false
	} else {
		result.HasMetrics = true
	}
	
	// 获取资源信息（无论Metrics API是否可用）
	resourceInfo := cs.getClusterResourceUsage(ctx, kubeClient, metricsClient)
	result.CPUUsage = resourceInfo.CPUUsage
	result.MemoryUsage = resourceInfo.MemoryUsage
	result.CPUUsedCores = resourceInfo.CPUUsedCores
	result.CPUTotalCores = resourceInfo.CPUTotalCores
	result.MemoryUsedGB = resourceInfo.MemoryUsedGB
	result.MemoryTotalGB = resourceInfo.MemoryTotalGB
	result.HasRealUsage = resourceInfo.HasRealUsage
	result.DataSource = resourceInfo.DataSource

	result.Success = true
	result.Status = "online"
	result.Message = fmt.Sprintf("集群连接测试成功，版本: %s", result.Version)
	result.ResponseTime = time.Since(startTime).Milliseconds()

	return result, nil
}

// updateClusterStatus 更新集群状态
func (cs *ClusterService) updateClusterStatus(cluster *models.ClusterConfig, status string) {
	cluster.Status = status
	if status == "online" {
		now := time.Now()
		cluster.LastCollectAt = &now
	}

	if err := cs.db.Save(cluster).Error; err != nil {
		logger.Error("更新集群状态失败: %v", err)
	}
}

// BatchTestAllClusters 批量测试所有集群连接状态
func (cs *ClusterService) BatchTestAllClusters() (map[uint]*ClusterTestResult, error) {
	clusters, err := cs.GetAllClusters()
	if err != nil {
		return nil, err
	}

	results := make(map[uint]*ClusterTestResult)

	// 并发测试所有集群（限制并发数）
	semaphore := make(chan struct{}, 5) // 最多5个并发测试
	resultChan := make(chan struct {
		ID     uint
		Result *ClusterTestResult
	}, len(clusters))

	for _, cluster := range clusters {
		go func(c models.ClusterConfig) {
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			result, _ := cs.TestClusterConnection(c.ID)
			resultChan <- struct {
				ID     uint
				Result *ClusterTestResult
			}{ID: c.ID, Result: result}
		}(cluster)
	}

	// 收集所有结果
	for i := 0; i < len(clusters); i++ {
		res := <-resultChan
		results[res.ID] = res.Result
	}

	logger.Info("批量测试完成，共测试 %d 个集群", len(clusters))
	return results, nil
}
