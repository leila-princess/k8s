package controller

import (
	"backend/app/dto"
	"backend/app/service"
	"backend/framework/dal"
	"backend/framework/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskController struct {
	taskService *service.TaskService
}

func NewTaskController() *TaskController {
	return &TaskController{
		taskService: &service.TaskService{},
	}
}

// ListTasks 查询项目列表
// @Summary 查询项目列表
// @Description 查询项目列表
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param name query string false "项目名称"
// @Param status query string false "状态"
// @Param type query string false "类型"
// @Success 200 {object} response.Response{data=dto.ListTasksResponse} "成功"
// @Router /api/v1/tasks [get]
func (c *TaskController) ListTasks(ctx *gin.Context) {
	var req dto.ListTasksRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.NewError().SetMsg("参数错误: " + err.Error()).Json(ctx)
		return
	}

	// 设置默认分页参数
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 1000 // 设置一个较大的值来返回所有数据
	}

	result, err := c.taskService.ListTasks(dal.GetDB(), req)
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetData("rows", result.Rows).Json(ctx)
}

// GetTaskById 获取项目详情
// @Summary 获取项目详情
// @Description 获取项目详情
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param taskId query int true "项目ID"
// @Success 200 {object} response.Response{data=dto.TaskDetailResponse} "成功"
// @Router /api/v1/tasks/detail [get]
func (c *TaskController) GetTaskById(ctx *gin.Context) {
	var req dto.GetTaskByIdRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.NewError().SetMsg("参数错误: " + err.Error()).Json(ctx)
		return
	}

	result, err := c.taskService.GetTaskById(dal.GetDB(), req.TaskId)
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetData("task", result).Json(ctx)
}

// UploadProject 上传项目
// @Summary 上传项目
// @Description 上传项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param request body dto.UploadProjectRequest true "上传项目请求"
// @Success 200 {object} response.Response "成功"
// @Router /api/v1/tasks [post]
func (c *TaskController) UploadProject(ctx *gin.Context) {
	var req dto.UploadProjectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg("参数错误: " + err.Error()).Json(ctx)
		return
	}

	// 从上下文中获取用户ID
	userId, exists := ctx.Get("userId")
	if !exists {
		response.NewError().SetMsg("用户未登录").Json(ctx)
		return
	}

	req.CreatorId = userId.(int)

	// 打印调试信息
	fmt.Printf("接收到上传项目请求:\n")
	fmt.Printf("项目名称: %s\n", req.Name)
	fmt.Printf("项目类型: %s\n", req.Type)
	fmt.Printf("环境变量数量: %d\n", len(req.EnvironmentVars))
	fmt.Printf("运行指令数量: %d\n", len(req.RunCommands))
	
	for i, env := range req.EnvironmentVars {
		fmt.Printf("环境变量 %d: Key=%s, Value=%s\n", i, env.Key, env.Value)
	}
	
	for i, cmd := range req.RunCommands {
		fmt.Printf("运行指令 %d: Command=%s\n", i, cmd.Command)
	}

	err := c.taskService.UploadProject(dal.GetDB(), req)
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("项目上传成功").Json(ctx)
}

// UpdateTask 更新项目
// @Summary 更新项目
// @Description 更新项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param request body dto.UpdateTaskRequest true "更新项目请求"
// @Success 200 {object} response.Response "成功"
// @Router /api/v1/tasks [put]
func (c *TaskController) UpdateTask(ctx *gin.Context) {
	var req dto.UpdateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg("参数错误: " + err.Error()).Json(ctx)
		return
	}

	// 从上下文中获取用户ID
	userId, exists := ctx.Get("userId")
	if !exists {
		response.NewError().SetMsg("用户未登录").Json(ctx)
		return
	}

	req.UpdaterId = userId.(int)

	err := c.taskService.UpdateTask(dal.GetDB(), req)
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("项目更新成功").Json(ctx)
}

// DelTask 删除项目
// @Summary 删除项目
// @Description 删除项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param taskId query int true "项目ID"
// @Success 200 {object} response.Response "成功"
// @Router /api/v1/tasks [delete]
func (c *TaskController) DelTask(ctx *gin.Context) {
	taskIdStr := ctx.Query("taskId")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		response.NewError().SetMsg("参数错误: taskId必须是数字").Json(ctx)
		return
	}

	err = c.taskService.DelTask(dal.GetDB(), taskId)
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("项目删除成功").Json(ctx)
}

// ExecuteTask 执行项目
// @Summary 执行项目
// @Description 执行项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param taskId query int true "项目ID"
// @Success 200 {object} response.Response "成功"
// @Router /api/v1/tasks/execute [post]
func (c *TaskController) ExecuteTask(ctx *gin.Context) {
	taskIdStr := ctx.Query("taskId")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		response.NewError().SetMsg("参数错误: taskId必须是数字").Json(ctx)
		return
	}

	err = c.taskService.ExecuteTask(dal.GetDB(), taskId)
	if err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("项目执行成功").Json(ctx)
}
