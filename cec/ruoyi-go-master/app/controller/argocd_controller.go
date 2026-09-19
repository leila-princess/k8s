package controller

import (
	"backend/app/dto"
	"backend/app/service"
	"backend/config"
	"backend/framework/dal"
	"backend/framework/response"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

type ArgoCDController struct {
	argoCDService *service.ArgoCDService
}

func NewArgoCDController(config *config.Config) *ArgoCDController {
	// 确保数据库已初始化
	db := dal.GetDB()
	if db == nil {
		panic("Database not initialized")
	}
	return &ArgoCDController{
		argoCDService: service.NewArgoCDService(
			db,
			config.Gitlab.BaseURL,
			config.Gitlab.Token,
		),
	}
}

// CreateArgoCDDeployment 创建ArgoCD部署
func (c *ArgoCDController) CreateArgoCDDeployment(ctx *gin.Context) {
	var req dto.ArgoCDDeployRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg("参数错误: " + err.Error()).Json(ctx)
		return
	}
	// 如果是定时触发，将时间字符串转换为时间戳
	if req.TriggerType == "schedule" && req.Schedule != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", req.Schedule, time.Local)
		if err != nil {
			response.NewError().SetMsg("时间格式错误，请使用 YYYY-MM-DD HH:mm:ss 格式").Json(ctx)
			return
		}
		// 检查是否是未来时间
		if t.Before(time.Now()) {
			response.NewError().SetMsg("定时触发时间必须是未来时间").Json(ctx)
			return
		}
	}
	// 从上下文中获取用户名
	username, exists := ctx.Get("username")
	if !exists {
		response.NewError().SetMsg("无法获取用户信息").Json(ctx)
		return
	}
	// 打印用户名
	fmt.Println("当前用户名:", username)
	// 创建部署
	deployment, err := c.argoCDService.CreateArgoCDDeployment(context.Background(), username.(string), &req)
	if err != nil {
		response.NewError().SetMsg("创建部署失败: " + err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetData("deployment", deployment).Json(ctx)
}

// GetDeploymentStatus 获取部署状态
func (c *ArgoCDController) GetDeploymentStatus(ctx *gin.Context) {
	
	username, exists := ctx.Get("username")
	if !exists {
		response.NewError().SetMsg("无法获取用户信息").Json(ctx)
		return
	}

	// 获取ArgoCD应用状态
	status, err := c.argoCDService.GetApplicationStatus(username.(string))
	if err != nil {
		response.NewError().SetMsg("获取ArgoCD状态失败: " + err.Error()).Json(ctx)
		return
	}

	// 直接返回完整的状态信息
	response.NewSuccess().SetData("deployment", status).Json(ctx)
}
