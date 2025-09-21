package config

import (
	"fmt"
	"os"
	"path/filepath"

	"cluster-resource-insight/internal/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Config struct {
	KubeClient    kubernetes.Interface
	MetricsClient metricsclientset.Interface
}

func NewConfig(kubeconfig string) (*Config, error) {
	var config *rest.Config
	var err error

	logger.Info("Initializing Kubernetes config with kubeconfig: %s", kubeconfig)

	if kubeconfig == "" {
		// 尝试集群内配置
		logger.Info("Trying in-cluster config...")
		config, err = rest.InClusterConfig()
		if err != nil {
			logger.Error("In-cluster config failed: %v", err)
			// 如果不在集群内，使用默认的 kubeconfig 文件
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
				logger.Info("Using default kubeconfig: %s", kubeconfig)
			}
		} else {
			logger.Info("Successfully loaded in-cluster config")
		}
	}

	if config == nil {
		// 使用 kubeconfig 文件
		logger.Info("Loading kubeconfig from file: %s", kubeconfig)
		if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
			return nil, fmt.Errorf("kubeconfig file not found: %s", kubeconfig)
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %v", err)
		}
		logger.Info("Successfully loaded kubeconfig")
	}

	// 创建 Kubernetes 客户端
	logger.Info("Creating Kubernetes client...")
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// 创建 Metrics 客户端
	logger.Info("Creating Metrics client...")
	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Metrics client: %v", err)
	}

	logger.Info("Successfully initialized all clients")
	return &Config{
		KubeClient:    kubeClient,
		MetricsClient: metricsClient,
	}, nil
}