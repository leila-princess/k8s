package controller

import (
	"backend/app/dto"
	"backend/app/model"
	"backend/app/service"
	"backend/framework/dal" // 修改为正确的数据库包导入
	"backend/framework/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TenantController struct{}

// Login 租户登录
// @Summary 租户登录
// @Description 租户用户登录接口，验证用户名和密码，返回JWT令牌
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param request body dto.TenantLoginRequest true "登录请求参数"
//   - username: string (必填) 登录用户名
//   - password: string (必填) 登录密码
//
// @Success 200 {object} response.Response{data=dto.TenantLoginResponse} "登录成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//   - data: object
//   - token: string JWT令牌，用于后续请求的认证
//
// @Failure 400 {object} response.Response "请求参数错误"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 401 {object} response.Response "用户名或密码错误"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 403 {object} response.Response "账户已停用"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Router /tenant/login [post]
func (c *TenantController) Login(ctx *gin.Context) {
	var req dto.TenantLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	tenantService := &service.TenantService{}
	token, user, err := tenantService.Login(dal.GetDB(), req.Username, req.Password)
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	// 返回token和用户信息
	response.NewSuccess().SetData("token", token).SetData("user", gin.H{
		"id":       user.Id,
		"username": user.Username,
		"status":   user.Status,
		"is_admin": user.IsAdmin,
	}).Json(ctx)
}

// CreateTenant 创建租户
// @Summary 创建新租户
// @Description 创建新的租户账号，包括用户信息和资源配额，返回用户的SA的token
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param request body dto.CreateTenantRequest true "创建租户请求参数"
//   - username: string (必填) 租户用户名
//   - password: string (必填) 租户密码
//   - resourceQuota: object (必填) 资源配额配置
//   - cpuLimit: string (必填) CPU使用上限，例如："2"表示2核
//   - cpuRequest: string (必填) CPU请求量，例如："1"表示1核
//   - memoryLimit: string (必填) 内存使用上限，例如："2Gi"表示2GB
//   - memoryRequest: string (必填) 内存请求量，例如："1Gi"表示1GB
//   - gpuLimit: string (必填) GPU使用上限，例如："1"表示1个GPU
//   - gpuMemory: string (必填) GPU显存限制，例如："8Gi"表示8GB
//
// @Success 200 {object} response.Response{data=string} "创建成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//   - data: string 返回的token
//
// @Failure 400 {object} response.Response "请求参数错误"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 409 {object} response.Response "用户名已存在"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Router /tenant [post]
func (c *TenantController) CreateTenant(ctx *gin.Context) {
	var req dto.CreateTenantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	tenantService := &service.TenantService{}
	token, err := tenantService.CreateTenant(dal.GetDB(), req.Username, req.Password, req.ResourceQuota)
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetData("token", token).Json(ctx)
}

// UpdateTenantById 通过ID更新租户信息
// @Summary 通过ID更新租户信息
// @Description 根据租户ID更新租户的基本信息，包括密码和状态
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param id path int true "租户ID" minimum(1)
// @Param request body dto.UpdateTenantRequest true "更新请求参数"
//   - password: string (可选) 新密码
//   - status: string (可选) 账户状态，"0"表示正常，"1"表示停用
//
// @Success 200 {object} response.Response "更新成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//
// @Failure 400 {object} response.Response "请求参数错误"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 401 {object} response.Response "未授权"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 403 {object} response.Response "权限不足"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 404 {object} response.Response "租户不存在"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Security ApiKeyAuth
// @Router /tenant/{id} [put]
func (c *TenantController) UpdateTenantById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		response.NewError().SetMsg("租户ID不能为空").Json(ctx)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.NewError().SetMsg("租户ID格式错误").Json(ctx)
		return
	}

	var req dto.UpdateTenantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}
	req.Id = id // 使用URL中的ID

	tenantService := &service.TenantService{}
	if err := tenantService.UpdateTenantById(dal.GetDB(), req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("更新成功").Json(ctx)
}

// UpdateTenantByUsername 通过用户名更新租户信息
// @Summary 通过用户名更新租户信息
// @Description 根据用户名更新租户的基本信息，包括密码和状态
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param username path string true "用户名" minLength(1)
// @Param request body dto.UpdateTenantRequest true "更新请求参数"
//   - password: string (可选) 新密码
//   - status: string (可选) 账户状态，"0"表示正常，"1"表示停用
//
// @Success 200 {object} response.Response "更新成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//
// @Failure 400 {object} response.Response "请求参数错误"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 401 {object} response.Response "未授权"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 403 {object} response.Response "权限不足"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 404 {object} response.Response "租户不存在"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Security ApiKeyAuth
// @Router /admin/by-username/{username} [put]
func (c *TenantController) UpdateTenantByUsername(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		response.NewError().SetMsg("用户名不能为空").Json(ctx)
		return
	}

	var req dto.UpdateTenantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}
	req.Username = username // 使用URL中的用户名

	tenantService := &service.TenantService{}
	if err := tenantService.UpdateTenantByUsername(dal.GetDB(), req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("更新成功").Json(ctx)
}

// DeleteTenant 删除租户
// @Summary 通过ID删除租户
// @Description 根据租户ID删除租户账号及其相关资源
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param id path int true "租户ID" minimum(1)
// @Success 200 {object} response.Response "删除成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//
// @Failure 400 {object} response.Response "请求参数错误"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 401 {object} response.Response "未授权"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 403 {object} response.Response "权限不足"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 404 {object} response.Response "租户不存在"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Security ApiKeyAuth
// @Router /admin/{id} [delete]
func (c *TenantController) DeleteTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		response.NewError().SetMsg("租户ID不能为空").Json(ctx)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.NewError().SetMsg("租户ID格式错误").Json(ctx)
		return
	}

	tenantService := &service.TenantService{}
	if err := tenantService.DeleteTenant(dal.GetDB(), id); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("删除成功").Json(ctx)
}

// DeleteTenantByUsername 通过用户名删除租户
// @Summary 通过用户名删除租户
// @Description 根据用户名删除租户账号及其相关资源
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param username path string true "用户名" minLength(1)
// @Success 200 {object} response.Response "删除成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//
// @Failure 400 {object} response.Response "请求参数错误"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 401 {object} response.Response "未授权"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 403 {object} response.Response "权限不足"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 404 {object} response.Response "租户不存在"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Security ApiKeyAuth
// @Router /admin/by-username/{username} [delete]
func (c *TenantController) DeleteTenantByUsername(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		response.NewError().SetMsg("用户名不能为空").Json(ctx)
		return
	}

	tenantService := &service.TenantService{}
	if err := tenantService.DeleteTenantByUsername(dal.GetDB(), username); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("删除成功").Json(ctx)
}

// UpdateTenantQuotaById 通过ID更新租户资源配额
// @Summary 通过ID更新租户资源配额
// @Description 根据租户ID更新其Kubernetes资源配额限制
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param id path int true "租户ID" minimum(1)
// @Param request body model.K8sResourceQuota true "资源配额参数"
//   - cpuLimit: string (必填) CPU使用上限，例如："2"表示2核
//   - cpuRequest: string (必填) CPU请求量，例如："1"表示1核
//   - memoryLimit: string (必填) 内存使用上限，例如："2Gi"表示2GB
//   - memoryRequest: string (必填) 内存请求量，例如："1Gi"表示1GB
//   - gpuLimit: string (必填) GPU使用上限，例如："1"表示1个GPU
//   - gpuMemory: string (必填) GPU显存限制，例如："8Gi"表示8GB
//
// @Success 200 {object} response.Response "更新成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//
// @Failure 400 {object} response.Response "请求参数错误"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 401 {object} response.Response "未授权"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 403 {object} response.Response "权限不足"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 404 {object} response.Response "租户不存在"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Security ApiKeyAuth
// @Router /admin/{id}/quota [put]
func (c *TenantController) UpdateTenantQuotaById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		response.NewError().SetMsg("租户ID不能为空").Json(ctx)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.NewError().SetMsg("租户ID格式错误").Json(ctx)
		return
	}

	var quota model.K8sResourceQuota
	if err := ctx.ShouldBindJSON(&quota); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	tenantService := &service.TenantService{}
	if err := tenantService.UpdateTenantQuotaById(dal.GetDB(), id, quota); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("更新资源配额成功").Json(ctx)
}

// UpdateTenantQuotaByUsername 通过用户名更新租户资源配额
// @Summary 通过用户名更新租户资源配额
// @Description 根据用户名更新其Kubernetes资源配额限制
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param username path string true "用户名" minLength(1)
// @Param request body model.K8sResourceQuota true "资源配额参数"
//   - cpuLimit: string (必填) CPU使用上限，例如："2"表示2核
//   - cpuRequest: string (必填) CPU请求量，例如："1"表示1核
//   - memoryLimit: string (必填) 内存使用上限，例如："2Gi"表示2GB
//   - memoryRequest: string (必填) 内存请求量，例如："1Gi"表示1GB
//   - gpuLimit: string (必填) GPU使用上限，例如："1"表示1个GPU
//   - gpuMemory: string (必填) GPU显存限制，例如："8Gi"表示8GB
//
// @Success 200 {object} response.Response "更新成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//
// @Failure 400 {object} response.Response "请求参数错误"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 401 {object} response.Response "未授权"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 403 {object} response.Response "权限不足"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 404 {object} response.Response "租户不存在"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Security ApiKeyAuth
// @Router /admin/by-username/{username}/quota [put]
func (c *TenantController) UpdateTenantQuotaByUsername(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		response.NewError().SetMsg("用户名不能为空").Json(ctx)
		return
	}

	var quota model.K8sResourceQuota
	if err := ctx.ShouldBindJSON(&quota); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	tenantService := &service.TenantService{}
	if err := tenantService.UpdateTenantQuotaByUsername(dal.GetDB(), username, quota); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("更新资源配额成功").Json(ctx)
}

// UpdatePassword 用户修改密码
// @Summary 修改当前用户密码
// @Description 已登录用户修改自己的密码
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param request body dto.UpdatePasswordRequest true "密码更新请求参数"
//   - password: string (必填) 新密码
//
// @Success 200 {object} response.Response "密码修改成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//
// @Failure 400 {object} response.Response "请求参数错误"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 401 {object} response.Response "未授权"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 404 {object} response.Response "用户不存在"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Security ApiKeyAuth
// @Router /tenant/password [put]
func (c *TenantController) UpdatePassword(ctx *gin.Context) {
	// 从上下文中获取用户ID
	userId, exists := ctx.Get("userId")
	if !exists {
		response.NewError().SetMsg("获取用户信息失败").Json(ctx)
		return
	}

	var req dto.UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	tenantService := &service.TenantService{}
	if err := tenantService.UpdatePassword(dal.GetDB(), userId.(int), req.Password); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("密码修改成功").Json(ctx)
}

// Logout 退出登录
// @Summary 退出登录
// @Description 退出当前用户的登录状态，清除token
// @Tags 租户管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "退出成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//
// @Failure 401 {object} response.Response "未授权"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Security ApiKeyAuth
// @Router /tenant/logout [post]
func (c *TenantController) Logout(ctx *gin.Context) {
	// 从上下文中获取用户ID
	userId, exists := ctx.Get("userId")
	if !exists {
		response.NewError().SetMsg("获取用户信息失败").Json(ctx)
		return
	}

	tenantService := &service.TenantService{}
	if err := tenantService.Logout(dal.GetDB(), userId.(int)); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("退出成功").Json(ctx)
}

// ListUsers 获取所有用户信息
// @Summary 获取所有用户信息
// @Description 获取系统中所有用户的信息列表（需要管理员权限）
// @Tags 租户管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]model.TenantUser} "获取成功"
//   - code: integer 状态码，200表示成功
//   - msg: string 响应消息
//   - data: array 用户信息列表
//
// @Failure 401 {object} response.Response "未授权"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Failure 403 {object} response.Response "权限不足"
//   - code: integer 状态码
//   - msg: string 错误信息
//
// @Security ApiKeyAuth
// @Router /admin/users [get]
func (c *TenantController) ListUsers(ctx *gin.Context) {
	// 获取数据库连接
	tenantService := &service.TenantService{}
	users, err := tenantService.ListUsers(dal.GetDB())
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetData("users", users).Json(ctx)
}
