package controller

import (
	"backend/app/dto"
	"backend/app/model"
	"backend/app/service"
	"backend/common/uuid"
	"backend/framework/dal" // 修改为正确的包路径
	"backend/framework/response"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sController struct{}

// CreateNamespace 创建命名空间和相关资源
func (c *K8sController) CreateNamespace(ctx *gin.Context) {
	var param struct {
		Name          string                 `json:"name"` // 租户名称，同时也是namespace名称
		ResourceQuota model.K8sResourceQuota `json:"resourceQuota"`
	}

	if err := ctx.ShouldBindJSON(&param); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	// 开启数据库事务
	err := dal.GetDB().Transaction(func(tx *gorm.DB) error {
		// 创建命名空间记录
		namespace := &model.K8sNamespace{
			Name:   param.Name,
			UserId: ctx.GetInt("userId"), // 从上下文获取当前用户ID
			Status: "0",
		}
		if err := tx.Create(namespace).Error; err != nil {
			return fmt.Errorf("创建命名空间记录失败：%v", err)
		}

		// 创建资源配额记录
		param.ResourceQuota.NamespaceId = namespace.Id
		if err := tx.Create(&param.ResourceQuota).Error; err != nil {
			return fmt.Errorf("创建资源配额记录失败：%v", err)
		}

		// 创建Kubernetes命名空间
		namespaceService := &service.K8sNamespaceService{}
		if err := namespaceService.CreateNamespace(param.Name); err != nil {
			return fmt.Errorf("创建Kubernetes命名空间失败：%v", err)
		}

		// 创建Kubernetes资源配额
		if err := namespaceService.CreateResourceQuota(param.Name, param.ResourceQuota); err != nil {
			return fmt.Errorf("创建Kubernetes资源配额失败：%v", err)
		}

		// 创建ServiceAccount
		rbacService := &service.K8sRbacService{}
		serviceAccountName := param.Name + "-sa"
		if err := rbacService.CreateServiceAccount(param.Name, serviceAccountName); err != nil {
			return fmt.Errorf("创建ServiceAccount失败：%v", err)
		}

		// 更新命名空间记录的ServiceAccount信息
		if err := tx.Model(namespace).Update("service_account", serviceAccountName).Error; err != nil {
			return fmt.Errorf("更新ServiceAccount信息失败：%v", err)
		}

		// 创建Role和RoleBinding
		if err := rbacService.CreateNamespaceRole(param.Name); err != nil {
			return fmt.Errorf("创建Role失败：%v", err)
		}

		if err := rbacService.CreateRoleBinding(param.Name, serviceAccountName); err != nil {
			return fmt.Errorf("创建RoleBinding失败：%v", err)
		}

		// 获取ServiceAccount的token
		saToken, err := rbacService.GetServiceAccountToken(param.Name, serviceAccountName)
		if err != nil {
			return fmt.Errorf("获取访问令牌失败：%v", err)
		}

		// 更新命名空间记录的token信息
		if err := tx.Model(namespace).Update("token", saToken).Error; err != nil {
			return fmt.Errorf("更新Token信息失败：%v", err)
		}
		// 返回成功响应
		response.NewSuccess().
			SetData("namespace", param.Name).
			SetData("token", saToken).
			Json(ctx)

		return nil
	})

	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}
}

// TestCreatePod 测试使用 ServiceAccount token 创建 Pod
func (c *K8sController) TestCreatePod(ctx *gin.Context) {
	fmt.Println("1. 开始处理请求")
	var req dto.TestPodRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fmt.Printf("2. 请求参数绑定失败: %v\n", err)
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}
	fmt.Println("3. 请求参数绑定成功")
	fmt.Println("4. 创建K8s客户端配置")
	// 创建 K8s 客户端配置
	config := &rest.Config{
		Host:        "https://192.168.56.101:6443",
		BearerToken: req.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // 在生产环境中应该设置为 false 并提供正确的 CA 证书
		},
	}
	fmt.Printf("使用的配置信息：\nHost: %s\nToken前20字符: %s...\n", config.Host, req.Token[:20])
	// 创建 K8s 客户端
	fmt.Println("5. 创建K8s客户端")
	// 创建 K8s 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("6. 创建K8s客户端失败: %v\n", err)
		response.NewError().SetMsg(fmt.Sprintf("创建 K8s 客户端失败：%v", err)).Json(ctx)
		return
	}
	fmt.Println("7. K8s客户端创建成功")
	// 修改 UUID 生成方式
	uuid, err := uuid.New()
	if err != nil {
		response.NewError().SetMsg(fmt.Sprintf("生成 UUID 失败：%v", err)).Json(ctx)
		return
	}

	// 创建测试 Pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("test-pod-%s", uuid),
			Namespace: req.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:latest",
				},
			},
		},
	}
	fmt.Printf("8. 尝试在命名空间 %s 创建Pod\n", req.Namespace)
	// 尝试创建 Pod
	createdPod, err := clientset.CoreV1().Pods(req.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		response.NewError().SetMsg(fmt.Sprintf("创建 Pod 失败：%v", err)).Json(ctx)
		return
	}
	fmt.Printf("10. Pod创建成功: %s\n", createdPod.Name)
	// 创建成功后立即删除 Pod
	defer func() {
		err = clientset.CoreV1().Pods(req.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("删除 Pod 失败：%v\n", err)
		}
	}()

	// 修改 response 的返回方式
	response.NewSuccess().
		SetData("podName", createdPod.Name).
		SetData("message", "Pod 创建成功并已自动删除").
		Json(ctx)
}
