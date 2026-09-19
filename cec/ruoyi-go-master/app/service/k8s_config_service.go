package service

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sConfigService struct{}

// GetClient 获取Kubernetes客户端
func (s *K8sConfigService) GetClient() (*kubernetes.Clientset, error) {
	// 从配置文件加载配置
	config, err := clientcmd.BuildConfigFromFlags("", "admin.conf")
	if err != nil {
		return nil, err
	}

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
