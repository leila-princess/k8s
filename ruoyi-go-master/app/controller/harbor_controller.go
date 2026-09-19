package controller

import (
	"backend/app/model"
	"backend/app/service"
	"backend/framework/dal"
	"backend/framework/response"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type HarborController struct{}

// ListBaseImages 获取基础镜像列表
func (c *HarborController) ListBaseImages(ctx *gin.Context) {
	harborService := &service.HarborService{}
	images, err := harborService.ListBaseImages()
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetData("images", images).Json(ctx)
}

// PullBaseImage 拉取基础镜像
func (c *HarborController) PullBaseImage(ctx *gin.Context) {

	var req struct {
		Repository string `json:"repository" binding:"required"`
		Tag        string `json:"tag" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	harborService := &service.HarborService{}
	if err := harborService.PullImage(req.Repository, req.Tag); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("拉取成功").Json(ctx)
}

// PullPrivateImage 使用私有账号拉取镜像
func (c *HarborController) PullPrivateImage(ctx *gin.Context) {
	// 从上下文获取用户名
	username := ctx.GetString("username")
	if username == "" {
		response.NewError().SetMsg("未获取到用户信息").Json(ctx)
		return
	}

	var req struct {
		Repository string `json:"repository" binding:"required"`
		Tag        string `json:"tag" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	harborService := &service.HarborService{}
	if err := harborService.PullImageWithAuth(username, req.Repository, req.Tag); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("拉取成功").Json(ctx)
}

func (c *HarborController) PushUserImage(ctx *gin.Context) {
	var req struct {
		ImageName string `json:"imageName" binding:"required"`
		Tag       string `json:"tag" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	// 从上下文获取用户名（认证中间件已经存入）
	username := ctx.GetString("username")
	fmt.Printf("当前用户名: %s\n", username)
	if username == "" {
		response.NewError().SetMsg("未获取到用户信息").Json(ctx)
		return
	}

	harborService := &service.HarborService{}
	// 使用用户名作为Harbor的用户名和密码
	if err := harborService.PushImage(username, req.ImageName, req.Tag); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("推送成功").Json(ctx)
}

// 在已有代码基础上添加

// AsyncPushImage 异步推送镜像
func (c *HarborController) AsyncPushImage(ctx *gin.Context) {
	var req struct {
		ImageName string `json:"imageName" binding:"required"`
		Tag       string `json:"tag" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	username := ctx.GetString("username")
	if username == "" {
		response.NewError().SetMsg("未获取到用户信息").Json(ctx)
		return
	}

	// 创建任务记录
	task := &model.ImageTask{
		TenantId:  ctx.GetInt("userId"),
		TaskType:  "push",
		Username:  username, // 添加username字段
		ImageName: req.ImageName,
		Tag:       req.Tag,
		Status:    model.TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 获取数据库连接
	db := dal.GetDB()
	if err := db.Create(task).Error; err != nil {
		response.NewError().SetMsg(fmt.Sprintf("创建任务失败: %v", err)).Json(ctx)
		return
	}

	// 启动异步任务
	harborService := &service.HarborService{}
	if err := harborService.AsyncPushImage(db, task); err != nil {
		response.NewError().SetMsg(fmt.Sprintf("启动任务失败: %v", err)).Json(ctx)
		return
	}

	response.NewSuccess().SetData("taskId", task.Id).SetMsg("任务已创建").Json(ctx)
}

// AsyncPullImage 异步拉取镜像
func (c *HarborController) AsyncPullImage(ctx *gin.Context) {
	var req struct {
		ImageName string `json:"imageName" binding:"required"`
		Tag       string `json:"tag" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	username := ctx.GetString("username")
	if username == "" {
		response.NewError().SetMsg("未获取到用户信息").Json(ctx)
		return
	}

	// 创建任务记录
	task := &model.ImageTask{
		TenantId:  ctx.GetInt("userId"),
		TaskType:  "pull",
		ImageName: req.ImageName,
		Username:  username, // 添加username字段
		Tag:       req.Tag,
		Status:    model.TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 获取数据库连接
	db := dal.GetDB()
	if err := db.Create(task).Error; err != nil {
		response.NewError().SetMsg(fmt.Sprintf("创建任务失败: %v", err)).Json(ctx)
		return
	}

	// 启动异步任务
	harborService := &service.HarborService{}
	if err := harborService.AsyncPullImage(db, task); err != nil {
		response.NewError().SetMsg(fmt.Sprintf("启动任务失败: %v", err)).Json(ctx)
		return
	}

	response.NewSuccess().SetData("taskId", task.Id).SetMsg("任务已创建").Json(ctx)
}

// GetTaskStatus 获取任务状态
func (c *HarborController) GetTaskStatus(ctx *gin.Context) {

	taskId := ctx.Param("taskId")
	if taskId == "" {
		response.NewError().SetMsg("任务ID不能为空").Json(ctx)
		return
	}

	id, err := strconv.Atoi(taskId)
	if err != nil {
		response.NewError().SetMsg("无效的任务ID").Json(ctx)
		return
	}

	// 获取数据库连接
	db := dal.GetDB()
	harborService := &service.HarborService{}
	task, err := harborService.GetTaskStatus(db, id)
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetData("task", task).Json(ctx)
}

// ListUserTasks 获取用户任务列表
func (c *HarborController) ListUserTasks(ctx *gin.Context) {
	tenantId := ctx.GetInt("userId")
	fmt.Sprintf("UserId: %v", tenantId)
	if tenantId == 0 {
		response.NewError().SetMsg("未获取到用户信息").Json(ctx)
		return
	}

	// 获取数据库连接
	db := dal.GetDB()
	harborService := &service.HarborService{}
	tasks, err := harborService.ListUserTasks(db, tenantId)
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetData("tasks", tasks).Json(ctx)
}

// LoadImage 加载镜像文件
func (c *HarborController) LoadImage(ctx *gin.Context) {
	// 从上下文获取用户信息
	username := ctx.GetString("username")
	if username == "" {
		response.NewError().SetMsg("未获取到用户信息").Json(ctx)
		return
	}

	// 获取上传的文件
	file, err := ctx.FormFile("image")
	if err != nil {
		response.NewError().SetMsg("获取上传文件失败").Json(ctx)
		return
	}

	// 检查文件扩展名
	if !strings.HasSuffix(file.Filename, ".tar") {
		response.NewError().SetMsg("只支持.tar格式的镜像文件").Json(ctx)
		return
	}

	// 创建临时目录存储文件
	uploadDir := "./uploads/images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.NewError().SetMsg(fmt.Sprintf("创建上传目录失败: %v", err)).Json(ctx)
		return
	}

	// 生成唯一的文件名
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// 保存文件
	if err := ctx.SaveUploadedFile(file, filepath); err != nil {
		response.NewError().SetMsg(fmt.Sprintf("保存文件失败: %v", err)).Json(ctx)
		return
	}

	// 创建任务记录
	task := &model.ImageLoadTask{
		ImageTask: model.ImageTask{
			TenantId:  ctx.GetInt("userId"),
			TaskType:  model.TaskTypeLoad,
			Username:  username,
			ImageName: strings.TrimSuffix(file.Filename, ".tar"),
			Tag:       "latest",
			Status:    model.TaskStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		FilePath: filepath,
	}

	// 获取数据库连接
	db := dal.GetDB()
	if err := db.Create(task).Error; err != nil {
		// 删除上传的文件
		os.Remove(filepath)
		response.NewError().SetMsg(fmt.Sprintf("创建任务失败: %v", err)).Json(ctx)
		return
	}

	// 启动异步任务
	harborService := &service.HarborService{}
	if err := harborService.LoadImage(db, task); err != nil {
		response.NewError().SetMsg(fmt.Sprintf("启动任务失败: %v", err)).Json(ctx)
		return
	}

	response.NewSuccess().SetData("taskId", task.Id).SetMsg("任务已创建").Json(ctx)
}

func (c *HarborController) AsyncBuildAndPushFromArchive(ctx *gin.Context) {
	username := ctx.GetString("username")
	if username == "" {
		response.NewError().SetMsg("未获取到用户信息").Json(ctx)
		return
	}

	var req model.BuildFromArchiveReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.NewError().SetMsg("参数错误: " + err.Error()).Json(ctx)
		return
	}

	// 读取上传文件
	file, err := ctx.FormFile("archive")
	if err != nil {
		response.NewError().SetMsg("缺少代码压缩包文件字段: archive").Json(ctx)
		return
	}

	// 保存到临时目录
	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("build-%d-%s", time.Now().UnixNano(), file.Filename))
	if err := ctx.SaveUploadedFile(file, tmpPath); err != nil {
		response.NewError().SetMsg("保存上传文件失败: " + err.Error()).Json(ctx)
		return
	}

	// 创建任务记录
	task := &model.ImageTask{
		TenantId:  ctx.GetInt("userId"),
		TaskType:  "build_push",
		Username:  username,            // 租户名（你的 Service 内会用它做 PullImageWithAuth & PushImage）
		ImageName: req.TargetImageName, // 目标仓库名
		Tag:       req.TargetTag,       // 目标标签
		Status:    model.TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db := dal.GetDB()
	if err := db.Create(task).Error; err != nil {
		_ = os.Remove(tmpPath)
		response.NewError().SetMsg("创建任务失败: " + err.Error()).Json(ctx)
		return
	}

	// 启动异步任务（内部会：ensureProject → PullImageWithAuth(租户) → build(本地标签) → PushImage(租户)）
	svc := &service.HarborService{}
	if err := svc.AsyncBuildAndPushFromArchive(
		db,
		task,
		req.BaseProject, req.BaseImageName, req.BaseTag,
		tmpPath,
	); err != nil {
		response.NewError().SetMsg("启动任务失败: " + err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().
		SetData("taskId", task.Id).
		SetMsg("构建任务已创建").
		Json(ctx)
}
