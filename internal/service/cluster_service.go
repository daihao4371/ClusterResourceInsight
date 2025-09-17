package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"cluster-resource-insight/internal/crypto"
	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/models"

	"gorm.io/gorm"
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

	log.Printf("成功创建集群配置: %s (ID: %d)", cluster.ClusterName, cluster.ID)
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

	log.Printf("成功更新集群配置: %s (ID: %d)", cluster.ClusterName, cluster.ID)
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

	log.Printf("成功删除集群配置: %s (ID: %d)", cluster.ClusterName, cluster.ID)
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
		}
	case "cert":
		config = &rest.Config{
			Host: cluster.APIServer,
			TLSClientConfig: rest.TLSClientConfig{
				CertData: []byte(authConfig.ClientCert),
				KeyData:  []byte(authConfig.ClientKey),
			},
		}
		if authConfig.CACert != "" {
			config.TLSClientConfig.CAData = []byte(authConfig.CACert)
		} else {
			config.TLSClientConfig.Insecure = true
		}
	case "kubeconfig":
		config, err = clientcmd.RESTConfigFromKubeConfig([]byte(authConfig.KubeConfig))
		if err != nil {
			return nil, nil, fmt.Errorf("解析kubeconfig失败: %v", err)
		}
	default:
		return nil, nil, fmt.Errorf("不支持的认证类型: %s", cluster.AuthType)
	}

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

// ClusterTestResult 集群连接测试结果
type ClusterTestResult struct {
	Success      bool      `json:"success"`           // 测试是否成功
	Status       string    `json:"status"`            // 集群状态：online/offline/error
	Message      string    `json:"message"`           // 测试结果消息
	Version      string    `json:"version"`           // Kubernetes版本
	NodeCount    int       `json:"node_count"`        // 节点数量
	NamespaceCount int     `json:"namespace_count"`   // 命名空间数量
	PodCount     int       `json:"pod_count"`         // Pod总数
	HasMetrics   bool      `json:"has_metrics"`       // 是否支持Metrics API
	TestTime     time.Time `json:"test_time"`         // 测试时间
	ResponseTime int64     `json:"response_time_ms"`  // 响应时间（毫秒）
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

	// 5. 测试Metrics API
	_, err = metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		result.HasMetrics = false
		log.Printf("集群 %s Metrics API 不可用: %v", cluster.ClusterName, err)
	} else {
		result.HasMetrics = true
	}

	// 测试成功
	result.Success = true
	result.Status = "online"
	result.Message = fmt.Sprintf("集群连接正常，版本: %s，节点数: %d，命名空间数: %d，Pod数: %d", 
		result.Version, result.NodeCount, result.NamespaceCount, result.PodCount)
	result.ResponseTime = time.Since(startTime).Milliseconds()

	// 更新集群状态为在线
	cs.updateClusterStatus(cluster, "online")

	log.Printf("集群连接测试成功: %s (%dms)", cluster.ClusterName, result.ResponseTime)
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

	// 测试Metrics API
	_, err = metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{Limit: 1})
	result.HasMetrics = (err == nil)

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
		log.Printf("更新集群状态失败: %v", err)
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
			semaphore <- struct{}{} // 获取信号量
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

	log.Printf("批量测试完成，共测试 %d 个集群", len(clusters))
	return results, nil
}
