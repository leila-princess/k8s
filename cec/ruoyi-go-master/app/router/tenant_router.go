package router

import (
	"backend/app/controller"
	"backend/app/middleware"
	"backend/config"
	"github.com/gin-gonic/gin"
)

// RegisterTenantApi 注册租户相关路由
func RegisterTenantApi(api *gin.RouterGroup) {
	harborController := &controller.HarborController{}
	// ArgoCD 部署相关接口
	argoCDController := controller.NewArgoCDController(config.Data)
	api.POST("/login", (&controller.TenantController{}).Login)

	api.POST("/register", (&controller.TenantController{}).CreateTenant)
	// 基础镜像相关接口 - 公开访问
	api.GET("/harbor/base-images", harborController.ListBaseImages)
	api.GET("/harbor/all-images", harborController.ListAllImages)
	// 需要管理员权限的接口
	adminApi := api.Group("/tenant")
	adminApi.Use(middleware.AuthMiddleware(), middleware.AdminRequired())
	{

		adminApi.PUT("/:id", (&controller.TenantController{}).UpdateTenantById)

		adminApi.PUT("/by-username/:username", (&controller.TenantController{}).UpdateTenantByUsername)
		adminApi.PUT("/:id/quota", (&controller.TenantController{}).UpdateTenantQuotaById)
		adminApi.PUT("/by-username/:username/quota", (&controller.TenantController{}).UpdateTenantQuotaByUsername)
		adminApi.DELETE("/:id", (&controller.TenantController{}).DeleteTenant)
		adminApi.DELETE("/by-username/:username", (&controller.TenantController{}).DeleteTenantByUsername)
		adminApi.GET("/users", (&controller.TenantController{}).ListUsers)
	}

	// 需要用户登录的接口
	userApi := api.Group("/tenant")
	userApi.Use(middleware.AuthMiddleware())
	{
		userApi.POST("/pull-private", harborController.PullPrivateImage) // 私有账号拉取
		userApi.PUT("/password", (&controller.TenantController{}).UpdatePassword)
		userApi.POST("/logout", (&controller.TenantController{}).Logout)
		userApi.POST("/push", harborController.PushUserImage)
		// 添加新的异步镜像操作接口
		userApi.POST("/async-push", harborController.AsyncPushImage)                // 异步推送镜像
		userApi.POST("/async-pull", harborController.AsyncPullImage)                // 异步拉取镜像
		userApi.GET("/task/:taskId", harborController.GetTaskStatus)                // 获取任务状态
		userApi.GET("/tasks", harborController.ListUserTasks)                       // 获取用户任务列表
		userApi.POST("/load-image", harborController.LoadImage)                     // 添加加载镜像接口
		userApi.POST("/async-build", harborController.AsyncBuildAndPushFromArchive) // 根据传入的文件构建镜像并推送
		// ArgoCD 部署相关接口
		userApi.POST("/argocd/deployment", argoCDController.CreateArgoCDDeployment)
		// 添加获取部署状态的路由
		userApi.GET("/argocd/getDeploymentStatus", argoCDController.GetDeploymentStatus)
	}

	// 项目相关接口 - 需要用户登录
	taskController := &controller.TaskController{}
	taskApi := api.Group("/tasks")
	taskApi.Use(middleware.AuthMiddleware())
	{
		// 查询项目列表
		taskApi.GET("", taskController.ListTasks)

		// 获取项目详情
		taskApi.GET("/detail", taskController.GetTaskById)

		// 上传项目
		taskApi.POST("", taskController.UploadProject)

		// 更新项目
		taskApi.PUT("", taskController.UpdateTask)

		// 删除项目
		taskApi.DELETE("", taskController.DelTask)

		// 执行项目
		taskApi.POST("/execute", taskController.ExecuteTask)
	}
}
