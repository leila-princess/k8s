package controller

import (
	"backend/app/dto"
	"backend/app/service"
	"backend/framework/response"
	"context"

	"github.com/gin-gonic/gin"
)

type DeploymentController struct {
	deploymentService *service.DeploymentService
}

func NewDeploymentController() *DeploymentController {
	return &DeploymentController{
		deploymentService: service.NewDeploymentService(),
	}
}
// CreateUserDeployment 在用户命名空间中创建部署
func (c *DeploymentController) CreateUserDeployment(ctx *gin.Context) {
	var req dto.CreateUserDeploymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg("参数错误: " + err.Error()).Json(ctx)
		return
	}

	// 从上下文中获取用户名作为命名空间
	username, exists := ctx.Get("username")
	if !exists {
		response.NewError().SetMsg("无法获取用户信息").Json(ctx)
		return
	}

	// 转换请求为标准部署请求
	deployReq := &dto.CreateDeploymentRequest{
		Namespace:     username.(string),
		Name:          req.Name,
		Image:         req.Image,
		Command:       req.Command,
		Args:          req.Args,
		EnvVars:       req.EnvVars,
		CPURequest:    req.CPURequest,
		MemoryRequest: req.MemoryRequest,
		CPULimit:      req.CPULimit,
		MemoryLimit:   req.MemoryLimit,
		Labels:        req.Labels,
		Annotations:   req.Annotations,
		Replicas:      req.Replicas,
		Ports:         req.Ports,
	}

	// 创建部署
	deployment, err := c.deploymentService.CreateUserDeployment(context.Background(), deployReq)
	if err != nil {
		response.NewError().SetMsg("创建部署失败: " + err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetData("deployment", deployment).Json(ctx)
}
// CreateDeployment 创建部署
func (c *DeploymentController) CreateDeployment(ctx *gin.Context) {
	var req dto.CreateDeploymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg("参数错误: " + err.Error()).Json(ctx)
		return
	}

	// 创建部署
	deployment, err := c.deploymentService.CreateDeployment(context.Background(), &req)
	if err != nil {
		response.NewError().SetMsg("创建部署失败: " + err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetData("deployment", deployment).Json(ctx)
}

// GetDeployment 获取部署详情
func (c *DeploymentController) GetDeployment(ctx *gin.Context) {
	response.NewError().SetMsg("获取部署详情功能暂未实现").Json(ctx)
}

// ListDeployments 获取部署列表
func (c *DeploymentController) ListDeployments(ctx *gin.Context) {
	response.NewError().SetMsg("获取部署列表功能暂未实现").Json(ctx)
}

// UpdateDeployment 更新部署
func (c *DeploymentController) UpdateDeployment(ctx *gin.Context) {
	var req dto.CreateDeploymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg("参数错误: " + err.Error()).Json(ctx)
		return
	}

	response.NewError().SetMsg("更新部署功能暂未实现").Json(ctx)
}

// DeleteDeployment 删除部署
func (c *DeploymentController) DeleteDeployment(ctx *gin.Context) {
	response.NewError().SetMsg("删除部署功能暂未实现").Json(ctx)
}

// ScaleDeployment 扩缩容部署
func (c *DeploymentController) ScaleDeployment(ctx *gin.Context) {
	var req struct {
		Replicas int32 `json:"replicas"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg("参数错误: " + err.Error()).Json(ctx)
		return
	}

	response.NewError().SetMsg("扩缩容部署功能暂未实现").Json(ctx)
}
