// app/service/deployment_service.go
package service

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"backend/app/dto"
)

type DeploymentService struct {
	k8sConfig *K8sConfigService
}

func NewDeploymentService() *DeploymentService {
	return &DeploymentService{
		k8sConfig: &K8sConfigService{},
	}
}

func (s *DeploymentService) CreateUserDeployment(ctx context.Context, req *dto.CreateDeploymentRequest) (*dto.DeploymentResponse, error) {
	// 获取Kubernetes客户端
	clientset, err := s.k8sConfig.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s client: %v", err)
	}

	// 创建Deployment
	deployment := s.buildDeployment(req)
	_, err = clientset.AppsV1().Deployments(req.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %v", err)
	}

	// 创建Service（如果需要）
	if len(req.Ports) > 0 {
		if err := s.createService(ctx, req, clientset); err != nil {
			// 记录错误但不中断流程
			fmt.Printf("Warning: failed to create service: %v\n", err)
		}
	}

	return &dto.DeploymentResponse{
		Namespace: req.Namespace,
		Name:      req.Name,
		Status:    "created",
		Message:   "Deployment created successfully",
	}, nil
}

func (s *DeploymentService) CreateDeployment(ctx context.Context, req *dto.CreateDeploymentRequest) (*dto.DeploymentResponse, error) {
	// 获取Kubernetes客户端
	clientset, err := s.k8sConfig.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s client: %v", err)
	}

	// 创建命名空间（如果不存在）
	if err := s.ensureNamespace(ctx, req.Namespace, clientset); err != nil {
		return nil, fmt.Errorf("failed to ensure namespace: %v", err)
	}

	// 创建Deployment
	deployment := s.buildDeployment(req)
	_, err = clientset.AppsV1().Deployments(req.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %v", err)
	}

	// 创建Service（如果需要）
	if len(req.Ports) > 0 {
		if err := s.createService(ctx, req, clientset); err != nil {
			// 记录错误但不中断流程
			fmt.Printf("Warning: failed to create service: %v\n", err)
		}
	}

	return &dto.DeploymentResponse{
		Namespace: req.Namespace,
		Name:      req.Name,
		Status:    "created",
		Message:   "Deployment created successfully",
	}, nil
}

func (s *DeploymentService) ensureNamespace(ctx context.Context, namespace string, clientset *kubernetes.Clientset) error {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		// 命名空间不存在，创建它
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		_, err = clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		return err
	}
	return err
}

func (s *DeploymentService) buildDeployment(req *dto.CreateDeploymentRequest) *appsv1.Deployment {
	// 设置默认副本数
	replicas := req.Replicas
	if replicas == 0 {
		replicas = 1
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    req.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": req.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": req.Name,
					},
					Annotations: req.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      req.Name,
							Image:     req.Image,
							Command:   req.Command,
							Args:      req.Args,
							Ports:     s.buildContainerPorts(req.Ports),
							Env:       s.buildEnvVars(req.EnvVars),
							Resources: s.buildResourceRequirements(req),
						},
					},
				},
			},
		},
	}
}

func (s *DeploymentService) createService(ctx context.Context, req *dto.CreateDeploymentRequest, clientset *kubernetes.Clientset) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    req.Labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": req.Name,
			},
			Ports: s.buildServicePorts(req.Ports),
			Type:  corev1.ServiceTypeClusterIP,
		},
	}

	_, err := clientset.CoreV1().Services(req.Namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

// 辅助方法：构建容器端口
func (s *DeploymentService) buildContainerPorts(ports []dto.ContainerPort) []corev1.ContainerPort {
	var containerPorts []corev1.ContainerPort
	for _, port := range ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.ContainerPort,
			Protocol:      corev1.Protocol(port.Protocol),
		})
	}
	return containerPorts
}

// 辅助方法：构建环境变量
func (s *DeploymentService) buildEnvVars(envVars map[string]string) []corev1.EnvVar {
	var envs []corev1.EnvVar
	for name, value := range envVars {
		envs = append(envs, corev1.EnvVar{
			Name:  name,
			Value: value,
		})
	}
	return envs
}

// 辅助方法：构建资源需求
func (s *DeploymentService) buildResourceRequirements(req *dto.CreateDeploymentRequest) corev1.ResourceRequirements {
	resources := corev1.ResourceRequirements{}

	requests := corev1.ResourceList{}
	limits := corev1.ResourceList{}

	if req.CPURequest != "" {
		requests[corev1.ResourceCPU] = resource.MustParse(req.CPURequest)
	}
	if req.MemoryRequest != "" {
		requests[corev1.ResourceMemory] = resource.MustParse(req.MemoryRequest)
	}
	if req.CPULimit != "" {
		limits[corev1.ResourceCPU] = resource.MustParse(req.CPULimit)
	}
	if req.MemoryLimit != "" {
		limits[corev1.ResourceMemory] = resource.MustParse(req.MemoryLimit)
	}

	if len(requests) > 0 {
		resources.Requests = requests
	}
	if len(limits) > 0 {
		resources.Limits = limits
	}

	return resources
}

// 辅助方法：构建服务端口
func (s *DeploymentService) buildServicePorts(ports []dto.ContainerPort) []corev1.ServicePort {
	var servicePorts []corev1.ServicePort
	for _, port := range ports {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name:       port.Name,
			Port:       port.ContainerPort, // 使用容器端口作为服务端口
			TargetPort: intstr.FromInt(int(port.ContainerPort)),
			Protocol:   corev1.Protocol(port.Protocol),
		})
	}
	return servicePorts
}
