package model

import (
	"time"
)

type Task struct {
	Id           int       `gorm:"primaryKey;autoIncrement" json:"id" swaggerignore:"true"`
	Name         string    `gorm:"size:128;not null" json:"name"`
	Description  string    `gorm:"type:text" json:"description"`
	Type         string    `gorm:"size:16;not null" json:"type"` // 'code' 或 'image'
	BaseImage    string    `gorm:"size:256" json:"baseImage"`    // 当type='code'时使用
	Image        string    `gorm:"size:256" json:"image"`        // 镜像
	Status       string    `gorm:"size:16;default:'stopped'" json:"status"` // 状态
	HealthStatus string    `gorm:"size:16;default:'unknown'" json:"healthStatus"` // 健康状态
	CPU          float64   `gorm:"type:decimal(5,2);default:1.0" json:"cpu"` // CPU配额（核心数）
	Memory       int       `gorm:"default:1024" json:"memory"` // 内存配额（MB）
	MemoryUnit   string    `gorm:"size:8;default:'MiB'" json:"memoryUnit"` // 内存单位：'MiB' 或 'GiB'
	Storage      int       `gorm:"default:10" json:"storage"` // 存储配额（GB）
	GPU          int       `gorm:"default:0" json:"gpu"` // GPU配额（个）
	GPUMemory    int       `gorm:"default:0" json:"gpuMemory"` // 显存配额（MB）
	GPUMemoryUnit string   `gorm:"size:8;default:'MiB'" json:"gpuMemoryUnit"` // 显存单位：'MiB' 或 'GiB'
	CreatorId    int       `gorm:"not null" json:"creatorId" swaggerignore:"true"`
	UpdaterId    int       `gorm:"not null" json:"updaterId" swaggerignore:"true"`
	CreateTime   time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime   time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

type TaskEnvironmentVar struct {
	Id     int    `gorm:"primaryKey;autoIncrement" json:"id" swaggerignore:"true"`
	TaskId int    `gorm:"not null;index" json:"taskId" swaggerignore:"true"`
	Key    string `gorm:"size:128;not null" json:"key"`
	Value  string `gorm:"size:512;not null" json:"value"`
}

type TaskRunCommand struct {
	Id      int    `gorm:"primaryKey;autoIncrement" json:"id" swaggerignore:"true"`
	TaskId  int    `gorm:"not null;index" json:"taskId" swaggerignore:"true"`
	Command string `gorm:"type:text;not null" json:"command"`
}
