package router

import (
	"github.com/gin-gonic/gin"
)

// Register 注册所有路由
// @title Ruoyi-Go API
// @version 1.0
// @description Ruoyi-Go 后台管理系统 API 文档
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host 39.96.159.232:3000
// @BasePath /api/v1
func Register(server *gin.Engine) {
	// 创建 v1 版本的路由组
	v1 := server.Group("/api/v1")

	// 创建不带认证中间件的路由组，用于公开接口
	publicApi := v1.Group("/")
	RegisterTenantApi(publicApi) // 注册租户相关路由（包含任务相关路由）

	// K8s 相关接口
	k8sApi := v1.Group("/k8s")
	RegisterK8sApi(k8sApi)

	// // Pipeline相关接口
	// RegisterPipelineApi(v1) // 直接使用v1路由组，不要再创建pipeline子组
	// Harbor相关接口
	// RegisterHarborApi(v1)
}
