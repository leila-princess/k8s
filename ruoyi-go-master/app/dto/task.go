package dto

import "backend/app/model"

// ListTasksRequest 查询项目列表请求
type ListTasksRequest struct {
	PageNum  int    `form:"pageNum" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Name     string `form:"name"`
	Status   string `form:"status"`
	Type     string `form:"type"`
}

// ListTasksResponse 查询项目列表响应
type ListTasksResponse struct {
	Rows []TaskListItem `json:"rows"`
}

// TaskListItem 项目列表项
type TaskListItem struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Image        string `json:"image"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	HealthStatus string `json:"healthStatus"`
	CreateTime   string `json:"createTime"`
}

// GetTaskByIdRequest 获取项目详情请求
type GetTaskByIdRequest struct {
	TaskId int `form:"taskId" binding:"required"`
}

// TaskDetailResponse 项目详情响应
type TaskDetailResponse struct {
	Id              int                     `json:"id"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Type            string                  `json:"type"`
	CPU             float64                 `json:"cpu"`
	Memory          int                     `json:"memory"`
	MemoryUnit      string                  `json:"memoryUnit"`
	Storage         int                     `json:"storage"`
	GPU             int                     `json:"gpu"`
	GPUMemory       int                     `json:"gpuMemory"`
	GPUMemoryUnit   string                  `json:"gpuMemoryUnit"`
	EnvironmentVars []TaskEnvironmentVarDto `json:"environmentVars"`
	RunCommands     []TaskRunCommandDto     `json:"runCommands"`
}

// TaskEnvironmentVarDto 环境变量DTO
type TaskEnvironmentVarDto struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TaskRunCommandDto 运行指令DTO
type TaskRunCommandDto struct {
	Command string `json:"command"`
}

// UploadProjectRequest 上传项目请求
type UploadProjectRequest struct {
	Name           string                  `json:"name" binding:"required"`
	Description    string                  `json:"description" binding:"required"`
	Type           string                  `json:"type" binding:"required,oneof=code image"`
	BaseImage      string                  `json:"baseImage"` // 当type='code'时必填
	CPU            float64                 `json:"cpu" binding:"min=0.1"`
	Memory         int                     `json:"memory" binding:"min=128"`
	MemoryUnit     string                  `json:"memoryUnit" swaggerignore:"true"` // 'MiB' 或 'GiB'
	Storage        int                     `json:"storage" binding:"min=1"`
	GPU            int                     `json:"gpu" binding:"min=0"`
	GPUMemory      int                     `json:"gpuMemory" binding:"min=0"`
	GPUMemoryUnit  string                  `json:"gpuMemoryUnit" swaggerignore:"true"` // 'MiB' 或 'GiB'
	EnvironmentVars []TaskEnvironmentVarDto `json:"environmentVars"`
	RunCommands     []TaskRunCommandDto     `json:"runCommands"`
	CreatorId       int                     `json:"creatorId" swaggerignore:"true"`
}

// UpdateTaskRequest 更新项目请求
type UpdateTaskRequest struct {
	Id             int                     `json:"id" binding:"required"`
	Name           string                  `json:"name" binding:"required"`
	Description    string                  `json:"description" binding:"required"`
	Type           string                  `json:"type" binding:"required,oneof=code image"`
	CPU            float64                 `json:"cpu" binding:"min=0.1"`
	Memory         int                     `json:"memory" binding:"min=128"`
	MemoryUnit     string                  `json:"memoryUnit" swaggerignore:"true"` // 'MiB' 或 'GiB'
	Storage        int                     `json:"storage" binding:"min=1"`
	GPU            int                     `json:"gpu" binding:"min=0"`
	GPUMemory      int                     `json:"gpuMemory" binding:"min=0"`
	GPUMemoryUnit  string                  `json:"gpuMemoryUnit" swaggerignore:"true"` // 'MiB' 或 'GiB'
	EnvironmentVars []TaskEnvironmentVarDto `json:"environmentVars"`
	RunCommands     []TaskRunCommandDto     `json:"runCommands"`
	UpdaterId       int                     `json:"updaterId" swaggerignore:"true"`
}

// DelTaskRequest 删除项目请求
type DelTaskRequest struct {
	TaskId int `form:"taskId" binding:"required"`
}

// ExecuteTaskRequest 执行项目请求
type ExecuteTaskRequest struct {
	TaskId int `form:"taskId" binding:"required"`
}

// 转换函数
func ToTaskEnvironmentVarModel(taskId int, dto TaskEnvironmentVarDto) model.TaskEnvironmentVar {
	return model.TaskEnvironmentVar{
		TaskId: taskId,
		Key:    dto.Key,
		Value:  dto.Value,
	}
}

func ToTaskRunCommandModel(taskId int, dto TaskRunCommandDto) model.TaskRunCommand {
	return model.TaskRunCommand{
		TaskId:  taskId,
		Command: dto.Command,
	}
}

func ToTaskEnvironmentVarDto(model model.TaskEnvironmentVar) TaskEnvironmentVarDto {
	return TaskEnvironmentVarDto{
		Key:   model.Key,
		Value: model.Value,
	}
}

func ToTaskRunCommandDto(model model.TaskRunCommand) TaskRunCommandDto {
	return TaskRunCommandDto{
		Command: model.Command,
	}
}
