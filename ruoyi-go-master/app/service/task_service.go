package service

import (
	"backend/app/dto"
	"backend/app/model"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TaskService struct{}

// ListTasks 查询项目列表
func (s *TaskService) ListTasks(db *gorm.DB, req dto.ListTasksRequest) (*dto.ListTasksResponse, error) {
	var tasks []model.Task
	query := db.Model(&model.Task{})

	// 构建查询条件
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	// 只有当PageSize大于0时才应用分页
	if req.PageSize > 0 {
		offset := (req.PageNum - 1) * req.PageSize
		query = query.Offset(offset).Limit(req.PageSize)
	}

	err := query.Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	// 构建响应
	var rows []dto.TaskListItem
	for _, task := range tasks {
		rows = append(rows, dto.TaskListItem{
			Id:           task.Id,
			Name:         task.Name,
			Image:        task.Image,
			Type:         task.Type,
			Status:       task.Status,
			HealthStatus: task.HealthStatus,
			CreateTime:   task.CreateTime.Format(time.RFC3339),
		})
	}

	return &dto.ListTasksResponse{Rows: rows}, nil
}

// GetTaskById 获取项目详情
func (s *TaskService) GetTaskById(db *gorm.DB, taskId int) (*dto.TaskDetailResponse, error) {
	var task model.Task
	err := db.First(&task, taskId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}

	// 获取环境变量
	var envVars []model.TaskEnvironmentVar
	err = db.Where("task_id = ?", taskId).Find(&envVars).Error
	if err != nil {
		return nil, err
	}

	// 获取运行指令
	var runCommands []model.TaskRunCommand
	err = db.Where("task_id = ?", taskId).Find(&runCommands).Error
	if err != nil {
		return nil, err
	}

	// 转换为DTO
	var envVarDtos []dto.TaskEnvironmentVarDto
	for _, envVar := range envVars {
		envVarDtos = append(envVarDtos, dto.ToTaskEnvironmentVarDto(envVar))
	}

	var runCommandDtos []dto.TaskRunCommandDto
	for _, cmd := range runCommands {
		runCommandDtos = append(runCommandDtos, dto.ToTaskRunCommandDto(cmd))
	}

	return &dto.TaskDetailResponse{
		Id:              task.Id,
		Name:            task.Name,
		Description:     task.Description,
		Type:            task.Type,
		CPU:             task.CPU,
		Memory:          task.Memory,
		MemoryUnit:      task.MemoryUnit,
		Storage:         task.Storage,
		GPU:             task.GPU,
		GPUMemory:       task.GPUMemory,
		GPUMemoryUnit:   task.GPUMemoryUnit,
		EnvironmentVars: envVarDtos,
		RunCommands:     runCommandDtos,
	}, nil
}

// UploadProject 上传项目
func (s *TaskService) UploadProject(db *gorm.DB, req dto.UploadProjectRequest) error {
	// 验证基础镜像（当type='code'时必填）
	if req.Type == "code" && req.BaseImage == "" {
		return errors.New("当项目类型为code时，基础镜像为必填项")
	}

	// 转换内存单位
	memory := req.Memory
	if req.MemoryUnit == "GiB" {
		memory = req.Memory * 1024 // GiB 转换为 MB
	}

	// 转换显存单位
	gpuMemory := req.GPUMemory
	if req.GPUMemoryUnit == "GiB" {
		gpuMemory = req.GPUMemory * 1024 // GiB 转换为 MB
	}

	// 开始事务
	return db.Transaction(func(tx *gorm.DB) error {
		// 创建项目
		task := model.Task{
			Name:          req.Name,
			Description:   req.Description,
			Type:          req.Type,
			BaseImage:     req.BaseImage,
			CPU:           req.CPU,
			Memory:        memory,
			MemoryUnit:    req.MemoryUnit,
			Storage:       req.Storage,
			GPU:           req.GPU,
			GPUMemory:     gpuMemory,
			GPUMemoryUnit: req.GPUMemoryUnit,
			CreatorId:     req.CreatorId,
			UpdaterId:     req.CreatorId,
			Status:        "stopped",
			HealthStatus:  "unknown",
		}

		if err := tx.Create(&task).Error; err != nil {
			return err
		}

		fmt.Printf("创建任务成功，任务ID: %d\n", task.Id)
		fmt.Printf("环境变量数量: %d\n", len(req.EnvironmentVars))
		fmt.Printf("运行指令数量: %d\n", len(req.RunCommands))

		// 创建环境变量
		for i, envVar := range req.EnvironmentVars {
			if envVar.Key != "" && envVar.Value != "" {
				fmt.Printf("创建环境变量 %d: Key=%s, Value=%s\n", i, envVar.Key, envVar.Value)
				envVarModel := dto.ToTaskEnvironmentVarModel(task.Id, envVar)
				if err := tx.Create(&envVarModel).Error; err != nil {
					fmt.Printf("创建环境变量失败: %v\n", err)
					return err
				}
				fmt.Printf("环境变量创建成功\n")
			} else {
				fmt.Printf("跳过环境变量 %d: Key或Value为空\n", i)
			}
		}

		// 创建运行指令
		for i, cmd := range req.RunCommands {
			if cmd.Command != "" {
				fmt.Printf("创建运行指令 %d: Command=%s\n", i, cmd.Command)
				cmdModel := dto.ToTaskRunCommandModel(task.Id, cmd)
				if err := tx.Create(&cmdModel).Error; err != nil {
					fmt.Printf("创建运行指令失败: %v\n", err)
					return err
				}
				fmt.Printf("运行指令创建成功\n")
			} else {
				fmt.Printf("跳运行指令 %d: Command为空\n", i)
			}
		}

		fmt.Printf("所有关联数据创建完成\n")
		return nil
	})
}

// UpdateTask 更新项目
func (s *TaskService) UpdateTask(db *gorm.DB, req dto.UpdateTaskRequest) error {
	// 检查项目是否存在
	var task model.Task
	err := db.First(&task, req.Id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在")
		}
		return err
	}

	// 转换内存单位
	memory := req.Memory
	if req.MemoryUnit == "GiB" {
		memory = req.Memory * 1024 // GiB 转换为 MB
	}

	// 转换显存单位
	gpuMemory := req.GPUMemory
	if req.GPUMemoryUnit == "GiB" {
		gpuMemory = req.GPUMemory * 1024 // GiB 转换为 MB
	}

	// 开始事务
	return db.Transaction(func(tx *gorm.DB) error {
		// 更新项目基本信息
		task.Name = req.Name
		task.Description = req.Description
		task.Type = req.Type
		task.CPU = req.CPU
		task.Memory = memory
		task.MemoryUnit = req.MemoryUnit
		task.Storage = req.Storage
		task.GPU = req.GPU
		task.GPUMemory = gpuMemory
		task.GPUMemoryUnit = req.GPUMemoryUnit
		task.UpdaterId = req.UpdaterId

		if err := tx.Save(&task).Error; err != nil {
			return err
		}

		// 删除原有的环境变量
		if err := tx.Where("task_id = ?", req.Id).Delete(&model.TaskEnvironmentVar{}).Error; err != nil {
			return err
		}

		// 创建新的环境变量
		for _, envVar := range req.EnvironmentVars {
			if envVar.Key != "" && envVar.Value != "" {
				envVarModel := dto.ToTaskEnvironmentVarModel(req.Id, envVar)
				if err := tx.Create(&envVarModel).Error; err != nil {
					return err
				}
			}
		}

		// 删除原有的运行指令
		if err := tx.Where("task_id = ?", req.Id).Delete(&model.TaskRunCommand{}).Error; err != nil {
			return err
		}

		// 创建新的运行指令
		for _, cmd := range req.RunCommands {
			if cmd.Command != "" {
				cmdModel := dto.ToTaskRunCommandModel(req.Id, cmd)
				if err := tx.Create(&cmdModel).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// DelTask 删除项目
func (s *TaskService) DelTask(db *gorm.DB, taskId int) error {
	// 检查项目是否存在
	var task model.Task
	err := db.First(&task, taskId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在")
		}
		return err
	}

	// 开始事务
	return db.Transaction(func(tx *gorm.DB) error {
		// 删除环境变量
		if err := tx.Where("task_id = ?", taskId).Delete(&model.TaskEnvironmentVar{}).Error; err != nil {
			return err
		}

		// 删除运行指令
		if err := tx.Where("task_id = ?", taskId).Delete(&model.TaskRunCommand{}).Error; err != nil {
			return err
		}

		// 删除项目
		if err := tx.Delete(&model.Task{}, taskId).Error; err != nil {
			return err
		}

		return nil
	})
}

// ExecuteTask 执行项目
func (s *TaskService) ExecuteTask(db *gorm.DB, taskId int) error {
	// 检查项目是否存在
	var task model.Task
	err := db.First(&task, taskId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在")
		}
		return err
	}

	// 更新项目状态为运行中
	task.Status = "running"
	task.HealthStatus = "healthy" // 假设执行后健康状态为健康

	if err := db.Save(&task).Error; err != nil {
		return err
	}

	// TODO: 这里应该添加实际的执行逻辑，比如调用K8s API创建部署等
	fmt.Printf("执行项目: %s (ID: %d)\n", task.Name, task.Id)

	return nil
}
