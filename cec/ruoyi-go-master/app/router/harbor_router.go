package router

import (
	"backend/app/controller"
	"backend/app/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterHarborApi(api *gin.RouterGroup) {
	harborController := &controller.HarborController{}

	// 基础镜像相关接口 - 公开访问
	api.GET("/harbor/base-images", harborController.ListBaseImages)

	// 需要认证的接口
	harborApi := api.Group("/harbor", middleware.AuthMiddleware())
	{
		harborApi.POST("/pull", harborController.PullBaseImage)
		harborApi.POST("/pull-private", harborController.PullPrivateImage) // 私有账号拉取
		harborApi.POST("/push", harborController.PushUserImage)
	}
}
