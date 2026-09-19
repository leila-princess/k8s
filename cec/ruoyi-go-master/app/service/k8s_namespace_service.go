package service

import (
	"backend/app/model"
	"backend/config"
	"context"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sNamespaceService struct{}

// CreateNamespace 创建新的命名空间
func (s *K8sNamespaceService) CreateNamespace(name string) error {
	client, err := (&K8sConfigService{}).GetClient()
	if err != nil {
		return err
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	_, err = client.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	return err
}
func (s *K8sNamespaceService) CreateResourceQuota(namespace string, quota model.K8sResourceQuota) error {
	client, err := (&K8sConfigService{}).GetClient()
	if err != nil {
		return err
	}

	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace + "-quota",
			Namespace: namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceCPU:            resource.MustParse(quota.CpuLimit),
				corev1.ResourceMemory:         resource.MustParse(quota.MemoryLimit),
				corev1.ResourceRequestsCPU:    resource.MustParse(quota.CpuRequest),
				corev1.ResourceRequestsMemory: resource.MustParse(quota.MemoryRequest),
				"nvidia.com/gpu":              resource.MustParse(quota.GpuLimit),
				"nvidia.com/gpu-memory":       resource.MustParse(quota.GpuMemory),
			},
		},
	}

	_, err = client.CoreV1().ResourceQuotas(namespace).Create(context.TODO(), resourceQuota, metav1.CreateOptions{})
	return err
}

func (s *K8sNamespaceService) CreateDefaultLimitRange(namespace string) error {
	client, err := (&K8sConfigService{}).GetClient()
	if err != nil {
		return err
	}
	limitRange := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace + "-limitrange",
			Namespace: namespace,
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{
				{
					Type: corev1.LimitTypeContainer,
					Default: corev1.ResourceList{
						corev1.ResourceCPU:    config.DefaultContainerCPULimit,
						corev1.ResourceMemory: config.DefaultContainerMemoryLimit,
					},
					DefaultRequest: corev1.ResourceList{
						corev1.ResourceCPU:    config.DefaultContainerCPURequest,
						corev1.ResourceMemory: config.DefaultContainerMemoryRequest,
					},
				},
			},
		},
	}

	_, err = client.CoreV1().LimitRanges(namespace).Create(context.TODO(), limitRange, metav1.CreateOptions{})
	return err
}

// UpdateResourceQuota 更新命名空间的资源配额
func (s *K8sNamespaceService) UpdateResourceQuota(namespace string, quota model.K8sResourceQuota) error {
	client, err := (&K8sConfigService{}).GetClient()
	if err != nil {
		return err
	}

	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace + "-quota",
			Namespace: namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceCPU:            resource.MustParse(quota.CpuLimit),
				corev1.ResourceMemory:         resource.MustParse(quota.MemoryLimit),
				corev1.ResourceRequestsCPU:    resource.MustParse(quota.CpuRequest),
				corev1.ResourceRequestsMemory: resource.MustParse(quota.MemoryRequest),
				"nvidia.com/gpu":              resource.MustParse(quota.GpuLimit),
				"nvidia.com/gpu-memory":       resource.MustParse(quota.GpuMemory),
			},
		},
	}

	// 获取现有的ResourceQuota
	_, err = client.CoreV1().ResourceQuotas(namespace).Get(context.TODO(), namespace+"-quota", metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// 如果不存在，则创建
			_, err = client.CoreV1().ResourceQuotas(namespace).Create(context.TODO(), resourceQuota, metav1.CreateOptions{})
			return err
		}
		return err
	}

	// 如果存在，则更新
	_, err = client.CoreV1().ResourceQuotas(namespace).Update(context.TODO(), resourceQuota, metav1.UpdateOptions{})
	return err
}
