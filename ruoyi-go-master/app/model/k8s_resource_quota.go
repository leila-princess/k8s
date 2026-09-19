package model

import "time"

type K8sResourceQuota struct {
	Id            int       `gorm:"primaryKey;autoIncrement" json:"id" swaggerignore:"true"`
	NamespaceId   int       `gorm:"not null" json:"namespaceId" swaggerignore:"true"`
	CpuLimit      string    `gorm:"size:32" json:"cpuLimit"`      // CPU限制
	CpuRequest    string    `gorm:"size:32" json:"cpuRequest"`    // CPU请求
	MemoryLimit   string    `gorm:"size:32" json:"memoryLimit"`   // 内存限制
	MemoryRequest string    `gorm:"size:32" json:"memoryRequest"` // 内存请求
	GpuLimit      string    `gorm:"size:32" json:"gpuLimit"`      // GPU限制
	GpuMemory     string    `gorm:"size:32" json:"gpuMemory"`     // GPU显存限制
	Status        string    `gorm:"size:1;default:0" json:"status" swaggerignore:"true"`
	CreateTime    time.Time `gorm:"autoCreateTime" json:"createTime" swaggerignore:"true"`
	UpdateTime    time.Time `gorm:"autoUpdateTime" json:"createTime" swaggerignore:"true"`
}
