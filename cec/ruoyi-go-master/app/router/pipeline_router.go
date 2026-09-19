package router

import (
	"backend/app/controller"
	"backend/app/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterPipelineApi(api *gin.RouterGroup) {
	// 直接在传入的路由组上注册路由
	pipelineController := &controller.PipelineController{}

	// 添加认证中间件的路由组
	pipelineApi := api.Group("/pipeline", middleware.AuthMiddleware())
	{
		pipelineApi.POST("", pipelineController.CreatePipeline)          // 创建流水线
		pipelineApi.POST("/trigger", pipelineController.TriggerPipeline) // 手动触发流水线
	}
}
