package config

import (
	"fmt"
	"log"
	"path/filepath"
	"os"

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

	log.Printf("Initializing Kubernetes config with kubeconfig: %s", kubeconfig)

	if kubeconfig == "" {
		// 尝试集群内配置
		log.Println("Trying in-cluster config...")
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("In-cluster config failed: %v", err)
			// 如果不在集群内，使用默认的 kubeconfig 文件
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
				log.Printf("Using default kubeconfig: %s", kubeconfig)
			}
		} else {
			log.Println("Successfully loaded in-cluster config")
		}
	}

	if config == nil {
		// 使用 kubeconfig 文件
		log.Printf("Loading kubeconfig from file: %s", kubeconfig)
		if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
			return nil, fmt.Errorf("kubeconfig file not found: %s", kubeconfig)
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %v", err)
		}
		log.Println("Successfully loaded kubeconfig")
	}

	// 创建 Kubernetes 客户端
	log.Println("Creating Kubernetes client...")
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// 创建 Metrics 客户端
	log.Println("Creating Metrics client...")
	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Metrics client: %v", err)
	}

	log.Println("Successfully initialized all clients")
	return &Config{
		KubeClient:    kubeClient,
		MetricsClient: metricsClient,
	}, nil
}