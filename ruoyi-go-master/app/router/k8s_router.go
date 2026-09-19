package router

import (
	"backend/app/controller"

	"github.com/gin-gonic/gin"
)

// Kubernetes路由组
func RegisterK8sApi(api *gin.RouterGroup) {
	// Kubernetes命名空间相关接口
	k8sController := &controller.K8sController{}
	deploymentController := controller.NewDeploymentController()

	// 添加测试接口路由
	api.POST("/test/pod", k8sController.TestCreatePod)
	api.POST("/namespace", k8sController.CreateNamespace)

	// Deployment相关接口（目前只有创建功能可用）
	api.POST("/deployment", deploymentController.CreateDeployment)
}
